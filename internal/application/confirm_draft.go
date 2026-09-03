package application

import (
	"context"
	"encoding/json"
	"errors"
	"log/slog"
	"strings"

	"github.com/derispewss/gonami-projects/internal/domain"
	"github.com/derispewss/gonami-projects/internal/parser"
	"github.com/derispewss/gonami-projects/internal/repository"
	"github.com/google/uuid"
)

type ConfirmDraft struct {
	users  *repository.UserRepo
	txs    *repository.TransactionRepo
	cats   *repository.CategoryRepo
	drafts *repository.DraftRepo
}

func NewConfirmDraft(
	u *repository.UserRepo, t *repository.TransactionRepo,
	c *repository.CategoryRepo, d *repository.DraftRepo,
) *ConfirmDraft {
	return &ConfirmDraft{users: u, txs: t, cats: c, drafts: d}
}

func (uc *ConfirmDraft) Confirm(ctx context.Context, jid string) ([]*domain.Transaction, error) {
	user, err := uc.users.GetOrCreateByJID(ctx, jid, "")
	if err != nil {
		return nil, err
	}

	draft, err := uc.drafts.LatestPending(ctx, user.ID)
	if err != nil {
		return nil, err
	}

	results, err := ParseBatchResults(draft.ExtractedData)
	if err != nil {
		return nil, err
	}

	if err := uc.drafts.UpdateStatus(ctx, draft.ID, domain.DraftPending, domain.DraftConfirmed); err != nil {
		return nil, err
	}

	var txs []*domain.Transaction
	for _, res := range results {
		var catID *uuid.UUID
		if res.Category != "" {
			cat, errCat := uc.cats.FindByNameAndType(ctx, user.ID, res.Category, res.Type)
			if errCat == nil {
				catID = &cat.ID
			}
		}

		tx := &domain.Transaction{
			UserID:          user.ID,
			Type:            res.Type,
			Amount:          res.Amount,
			Description:     res.Description,
			CategoryID:      catID,
			CategoryName:    res.Category,
			Merchant:        res.Merchant,
			TransactionDate: res.Date,
			SourceType:      draft.SourceType,
			SourceMessageID: "",
			RawMessage:      draft.RawContent,
			WalletID:        user.ActiveWalletID,
		}
		if err := uc.txs.Create(ctx, tx); err != nil {
			return nil, err
		}
		txs = append(txs, tx)
	}

	slog.Info("draft confirmed and txs created", "draft_id", draft.ID, "count", len(txs))
	return txs, nil
}

// ParseBatchResults decodes a draft's extracted_data which may be either a
// single parser.Result object (legacy) or a JSON array of results (new).
func ParseBatchResults(data json.RawMessage) ([]*parser.Result, error) {
	trimmed := []byte(strings.TrimSpace(string(data)))
	if len(trimmed) > 0 && trimmed[0] == '[' {
		var list []*parser.Result
		if err := json.Unmarshal(trimmed, &list); err != nil {
			return nil, err
		}
		return list, nil
	}
	var single parser.Result
	if err := json.Unmarshal(trimmed, &single); err != nil {
		return nil, err
	}
	return []*parser.Result{&single}, nil
}

func (uc *ConfirmDraft) Reject(ctx context.Context, jid string) error {
	user, err := uc.users.GetOrCreateByJID(ctx, jid, "")
	if err != nil {
		return err
	}

	draft, err := uc.drafts.LatestPending(ctx, user.ID)
	if err != nil {
		if errors.Is(err, domain.ErrNotFound) {
			return nil
		}
		return err
	}

	if err := uc.drafts.UpdateStatus(ctx, draft.ID, domain.DraftPending, domain.DraftRejected); err != nil {
		return err
	}

	slog.Info("draft rejected", "draft_id", draft.ID)
	return nil
}
