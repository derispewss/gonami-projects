package application

import (
	"context"
	"time"

	"github.com/derispewss/gonami-projects/internal/domain"
	"github.com/derispewss/gonami-projects/internal/parser"
	"github.com/derispewss/gonami-projects/internal/repository"
)

type BudgetUC struct {
	users   *repository.UserRepo
	txs     *repository.TransactionRepo
	budgets *repository.BudgetRepo
}

func NewBudgetUC(u *repository.UserRepo, t *repository.TransactionRepo,
	b *repository.BudgetRepo) *BudgetUC {
	return &BudgetUC{users: u, txs: t, budgets: b}
}

func (uc *BudgetUC) Set(ctx context.Context, jid string, cmd *parser.BudgetCommand) (*domain.Budget, bool, error) {
	if cmd.Delete {
		return nil, false, uc.Delete(ctx, jid, cmd.Category)
	}
	if cmd.Category == "" || cmd.Amount <= 0 {
		return nil, false, domain.ErrInvalidInput
	}

	user, err := uc.users.GetOrCreateByJID(ctx, jid, "")
	if err != nil {
		return nil, false, err
	}

	updated := false
	if _, err := uc.budgets.Get(ctx, user.ID, cmd.Category); err == nil {
		updated = true
	}

	b := &domain.Budget{UserID: user.ID, CategoryName: cmd.Category, MonthlyLimit: cmd.Amount}
	if err := uc.budgets.Upsert(ctx, b); err != nil {
		return nil, false, err
	}
	return b, updated, nil
}

func (uc *BudgetUC) Delete(ctx context.Context, jid, category string) error {
	if category == "" {
		return domain.ErrInvalidInput
	}
	user, err := uc.users.GetOrCreateByJID(ctx, jid, "")
	if err != nil {
		return err
	}
	return uc.budgets.Delete(ctx, user.ID, category)
}

type BudgetStatus struct {
	Name   string
	Limit  int64
	Spent  int64
	Ratio  float64
	Breach bool
	Warned bool
}

func (uc *BudgetUC) Status(ctx context.Context, jid string) ([]BudgetStatus, error) {
	user, err := uc.users.GetOrCreateByJID(ctx, jid, "")
	if err != nil {
		return nil, err
	}

	items, err := uc.budgets.ListByUser(ctx, user.ID)
	if err != nil {
		return nil, err
	}
	if len(items) == 0 {
		return []BudgetStatus{}, nil
	}

	loc, _ := time.LoadLocation("Asia/Jakarta")
	now := time.Now().In(loc)
	from := time.Date(now.Year(), now.Month(), 1, 0, 0, 0, 0, loc)
	to := from.AddDate(0, 1, 0)

	sums, err := uc.txs.SumByCategory(ctx, user.ID, domain.TypeExpense, from, to)
	if err != nil {
		return nil, err
	}
	spentByCat := make(map[string]int64)
	for _, s := range sums {
		spentByCat[lower(s.CategoryName)] += s.Total
	}

	out := make([]BudgetStatus, 0, len(items))
	for _, b := range items {
		st := BudgetStatus{Name: b.CategoryName, Limit: b.MonthlyLimit}
		st.Spent = spentByCat[lower(b.CategoryName)]
		if st.Limit > 0 {
			st.Ratio = float64(st.Spent) / float64(st.Limit)
		}
		st.Breach = st.Ratio >= 1
		st.Warned = !st.Breach && st.Ratio >= 0.8
		out = append(out, st)
	}
	return out, nil
}

func lower(s string) string {
	b := []byte(s)
	for i := range b {
		if 'A' <= b[i] && b[i] <= 'Z' {
			b[i] += 'a' - 'A'
		}
	}
	return string(b)
}
