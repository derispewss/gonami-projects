package application

import (
	"context"

	"github.com/derispewss/gonami-projects/internal/domain"
	"github.com/derispewss/gonami-projects/internal/repository"
)

type ManageTransaction struct {
	users *repository.UserRepo
	txs   *repository.TransactionRepo
}

func NewManageTransaction(u *repository.UserRepo, t *repository.TransactionRepo) *ManageTransaction {
	return &ManageTransaction{users: u, txs: t}
}

func (uc *ManageTransaction) GetLastTransactions(ctx context.Context, jid string, limit int) ([]*domain.Transaction, error) {
	user, err := uc.users.GetOrCreateByJID(ctx, jid, "")
	if err != nil {
		return nil, err
	}

	return uc.txs.GetLatest(ctx, user.ID, limit)
}

func (uc *ManageTransaction) DeleteLast(ctx context.Context, jid string) (*domain.Transaction, error) {
	user, err := uc.users.GetOrCreateByJID(ctx, jid, "")
	if err != nil {
		return nil, err
	}

	return uc.txs.DeleteLast(ctx, user.ID)
}
