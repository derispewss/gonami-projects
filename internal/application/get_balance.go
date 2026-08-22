package application

import (
	"context"

	"github.com/derispewss/finwa-projects/internal/repository"
)

type BalanceOutcome struct {
	TotalIncome  int64
	TotalExpense int64
	NetBalance   int64
}

type GetBalance struct {
	users *repository.UserRepo
	txs   *repository.TransactionRepo
}

func NewGetBalance(u *repository.UserRepo, t *repository.TransactionRepo) *GetBalance {
	return &GetBalance{users: u, txs: t}
}

func (uc *GetBalance) Balance(ctx context.Context, jid string) (*BalanceOutcome, error) {
	user, err := uc.users.GetOrCreateByJID(ctx, jid, "")
	if err != nil {
		return nil, err
	}

	income, expense, net, err := uc.txs.Balance(ctx, user.ID)
	if err != nil {
		return nil, err
	}

	return &BalanceOutcome{
		TotalIncome:  income,
		TotalExpense: expense,
		NetBalance:   net,
	}, nil
}
