package application

import (
	"context"
	"encoding/json"
	"errors"
	"log/slog"
	"time"

	"github.com/derispewss/finwa-projects/internal/ai"
	"github.com/derispewss/finwa-projects/internal/config"
	"github.com/derispewss/finwa-projects/internal/domain"
	"github.com/derispewss/finwa-projects/internal/parser"
	"github.com/derispewss/finwa-projects/internal/repository"
	"github.com/google/uuid"
)

type RecordStatus string

const (
	RecordSaved   RecordStatus = "saved"
	RecordDraft   RecordStatus = "draft"
	RecordUnclear RecordStatus = "unclear"
)

type RecordOutcome struct {
	Status  RecordStatus
	Tx      *domain.Transaction
	Draft   *domain.TransactionDraft
	Parsed  *parser.Result
	Message string
}

type RecordTransaction struct {
	users  *repository.UserRepo
	cats   *repository.CategoryRepo
	txs    *repository.TransactionRepo
	drafts *repository.DraftRepo
	prs    *parser.Parser
	cfg    *config.Config
	ai     ai.AIClient
}

func NewRecordTransaction(
	u *repository.UserRepo, c *repository.CategoryRepo,
	t *repository.TransactionRepo, d *repository.DraftRepo,
	p *parser.Parser, cfg *config.Config, aiClient ai.AIClient,
) *RecordTransaction {
	return &RecordTransaction{users: u, cats: c, txs: t, drafts: d, prs: p, cfg: cfg, ai: aiClient}
}

func (uc *RecordTransaction) FromText(ctx context.Context, jid, pushName, text, msgID string) (*RecordOutcome, error) {
	return uc.FromParsed(ctx, jid, pushName, nil, domain.SourceText, text, msgID)
}

func (uc *RecordTransaction) FromParsed(ctx context.Context, jid, pushName string,
	res *parser.Result, source domain.SourceType, rawContent, msgID string) (*RecordOutcome, error) {

	user, err := uc.users.GetOrCreateByJID(ctx, jid, pushName)
	if err != nil {
		return nil, err
	}

	_ = uc.drafts.CancelAllPending(ctx, user.ID)

	fromLLM := false
	loc, _ := time.LoadLocation("Asia/Jakarta")
	now := time.Now().In(loc)

	if res == nil {
		parsed, perr := uc.prs.Parse(ctx, rawContent, now)
		if perr != nil && perr != parser.ErrNotTransaction {
			return nil, perr
		}
		res = parsed
	}

	if (res == nil || res.Confidence < uc.cfg.ConfidenceAskConfirm) && source == domain.SourceText {
		if cand := uc.llmFallback(ctx, rawContent, now); cand != nil {
			slog.Info("layer-2 llm berhasil mengekstrak transaksi",
				"amount", cand.Amount, "confidence", cand.Confidence)
			res = cand
			fromLLM = true
		}
	}

	if res == nil || res.Confidence < uc.cfg.ConfidenceAskConfirm {
		return &RecordOutcome{Status: RecordUnclear}, nil
	}

	var catID *uuid.UUID
	if res.Category != "" {
		cat, errCat := uc.cats.FindByNameAndType(ctx, user.ID, res.Category, res.Type)
		if errCat == nil {
			catID = &cat.ID
		}
	}

	if !fromLLM && res.Confidence >= uc.cfg.ConfidenceAutoSave {
		tx := &domain.Transaction{
			UserID:          user.ID,
			Type:            res.Type,
			Amount:          res.Amount,
			Description:     res.Description,
			CategoryID:      catID,
			CategoryName:    res.Category,
			Merchant:        res.Merchant,
			TransactionDate: res.Date,
			SourceType:      source,
			SourceMessageID: msgID,
			RawMessage:      rawContent,
		}
		if err := uc.txs.Create(ctx, tx); err != nil {
			return nil, err
		}
		slog.Info("transaction auto-saved", "tx_id", tx.ID, "user_id", user.ID, "source", source)
		return &RecordOutcome{Status: RecordSaved, Tx: tx, Parsed: res}, nil
	}

	extracted, _ := json.Marshal(res)
	draft := &domain.TransactionDraft{
		UserID:        user.ID,
		SourceType:    source,
		RawContent:    rawContent,
		ExtractedData: extracted,
		Confidence:    res.Confidence,
		Status:        domain.DraftPending,
		ExpiresAt:     time.Now().Add(time.Duration(uc.cfg.DraftExpiryMinutes) * time.Minute),
	}

	if err := uc.drafts.Create(ctx, draft); err != nil {
		return nil, err
	}

	slog.Info("draft created", "draft_id", draft.ID, "user_id", user.ID, "source", source)
	return &RecordOutcome{Status: RecordDraft, Draft: draft, Parsed: res}, nil
}

var wibLoc = func() *time.Location {
	loc, err := time.LoadLocation("Asia/Jakarta")
	if err != nil {
		return time.FixedZone("WIB", 7*3600)
	}
	return loc
}()

func (uc *RecordTransaction) llmFallback(ctx context.Context, text string, now time.Time) *parser.Result {
	if uc.ai == nil || !parser.MayContainTransaction(text) {
		return nil
	}
	ext, err := uc.ai.ExtractFromChatText(ctx, text, now)
	if err != nil {
		if errors.Is(err, ai.ErrBudgetExceeded) {
			slog.Warn("llm fallback dilewati — budget harian habis")
		} else {
			slog.Debug("llm fallback gagal", "error", err)
		}
		return nil
	}
	if ext == nil || !ext.IsValid() || ext.Confidence < uc.cfg.ConfidenceAskConfirm {
		return nil
	}

	txDate := now
	if ext.DateHint != "" {
		for _, layout := range []string{"2006-01-02", "02/01/2006", "2 January 2006"} {
			if d, perr := time.ParseInLocation(layout, ext.DateHint, wibLoc); perr == nil {
				txDate = d
				break
			}
		}
	}

	desc := ext.Description
	if desc == "" && ext.Merchant != "" {
		desc = ext.Merchant
	}
	return &parser.Result{
		Type:        domain.TransactionType(ext.Type),
		Amount:      ext.Amount,
		Description: desc,
		Category:    ext.CategoryHint,
		Merchant:    ext.Merchant,
		Date:        txDate,
		Confidence:  ext.Confidence,
	}
}
