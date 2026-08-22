package application

import (
	"context"
	"encoding/json"
	"errors"
	"log/slog"

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

func (uc *ConfirmDraft) Confirm(ctx context.Context, jid string) (*domain.Transaction, error) {
	user, err := uc.users.GetOrCreateByJID(ctx, jid, "")
	if err != nil {
		return nil, err
	}

	draft, err := uc.drafts.LatestPending(ctx, user.ID)
	if err != nil {
		return nil, err
	}

	var res parser.Result
	if err := json.Unmarshal(draft.ExtractedData, &res); err != nil {
		return nil, err
	}

	if err := uc.drafts.UpdateStatus(ctx, draft.ID, domain.DraftPending, domain.DraftConfirmed); err != nil {
		return nil, err
	}

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
	}

	if err := uc.txs.Create(ctx, tx); err != nil {
		return nil, err
	}

	slog.Info("draft confirmed and tx created", "draft_id", draft.ID, "tx_id", tx.ID)
	return tx, nil
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
