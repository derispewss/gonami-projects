package application

import (
	"context"

	"github.com/derispewss/gonami-projects/internal/repository"
)

type ResetData struct {
	users   *repository.UserRepo
	txs     *repository.TransactionRepo
	budgets *repository.BudgetRepo
	wallets *repository.WalletRepo
	cats    *repository.CategoryRepo
	drafts  *repository.DraftRepo
}

func NewResetData(u *repository.UserRepo, t *repository.TransactionRepo,
	b *repository.BudgetRepo, w *repository.WalletRepo,
	c *repository.CategoryRepo, d *repository.DraftRepo) *ResetData {
	return &ResetData{users: u, txs: t, budgets: b, wallets: w, cats: c, drafts: d}
}

func (uc *ResetData) DeleteAll(ctx context.Context, jid string) (int64, error) {
	user, err := uc.users.GetOrCreateByJID(ctx, jid, "")
	if err != nil {
		return 0, err
	}

	if err := uc.drafts.CancelAllPending(ctx, user.ID); err != nil {
		return 0, err
	}
	if err := uc.wallets.DeleteByUser(ctx, user.ID); err != nil {
		return 0, err
	}
	if err := uc.budgets.DeleteByUser(ctx, user.ID); err != nil {
		return 0, err
	}
	if err := uc.cats.DeleteCustomByUser(ctx, user.ID); err != nil {
		return 0, err
	}
	if err := uc.txs.DeleteAllByUser(ctx, user.ID); err != nil {
		return 0, err
	}

	count, err := uc.txs.CountByUser(ctx, user.ID)
	if err != nil {
		return 0, err
	}
	return count, nil
}
