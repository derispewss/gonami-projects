package application

import (
	"context"
	"time"

	"github.com/derispewss/gonami-projects/internal/domain"
	"github.com/derispewss/gonami-projects/internal/repository"
	"github.com/google/uuid"
)

type WalletUC struct {
	users   *repository.UserRepo
	txs     *repository.TransactionRepo
	wallets *repository.WalletRepo
}

func NewWalletUC(u *repository.UserRepo, t *repository.TransactionRepo,
	w *repository.WalletRepo) *WalletUC {
	return &WalletUC{users: u, txs: t, wallets: w}
}

func (uc *WalletUC) Add(ctx context.Context, jid, name string) (*domain.Wallet, error) {
	if name == "" || len(name) > 100 {
		return nil, domain.ErrInvalidInput
	}
	user, err := uc.users.GetOrCreateByJID(ctx, jid, "")
	if err != nil {
		return nil, err
	}
	w := &domain.Wallet{UserID: user.ID, Name: name}
	if err := uc.wallets.Create(ctx, w); err != nil {
		return nil, err
	}
	return w, nil
}

func (uc *WalletUC) Switch(ctx context.Context, jid, name string) (*domain.Wallet, error) {
	user, err := uc.users.GetOrCreateByJID(ctx, jid, "")
	if err != nil {
		return nil, err
	}
	w, err := uc.wallets.GetByName(ctx, user.ID, name)
	if err != nil {
		return nil, err
	}
	if err := uc.wallets.SetActive(ctx, user.ID, w.ID); err != nil {
		return nil, err
	}
	return w, nil
}

func (uc *WalletUC) Deactivate(ctx context.Context, jid string) error {
	user, err := uc.users.GetOrCreateByJID(ctx, jid, "")
	if err != nil {
		return err
	}
	return uc.wallets.ClearActive(ctx, user.ID)
}

func (uc *WalletUC) List(ctx context.Context, jid string) (*WalletSummaryResult, error) {
	user, err := uc.users.GetOrCreateByJID(ctx, jid, "")
	if err != nil {
		return nil, err
	}

	items, err := uc.wallets.ListByUser(ctx, user.ID)
	if err != nil {
		return nil, err
	}

	loc, _ := time.LoadLocation("Asia/Jakarta")
	now := time.Now().In(loc)
	from := time.Date(now.Year(), now.Month(), 1, 0, 0, 0, 0, loc)
	to := from.AddDate(0, 1, 0)

	res := &WalletSummaryResult{
		Wallets:  items,
		ActiveID: user.ActiveWalletID,
		Spend:    make(map[string]int64),
	}
	for _, w := range items {
		_, expense, err := uc.txs.BalanceByWallet(ctx, user.ID, w.ID, from, to)
		if err != nil {
			return nil, err
		}
		res.Spend[w.Name] = expense
	}
	return res, nil
}

type WalletSummaryResult struct {
	Wallets  []*domain.Wallet
	ActiveID *uuid.UUID
	Spend    map[string]int64
}
