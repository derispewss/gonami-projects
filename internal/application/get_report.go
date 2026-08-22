package application

import (
	"context"
	"time"

	"github.com/derispewss/gonami-projects/internal/domain"
	"github.com/derispewss/gonami-projects/internal/repository"
)

type ReportType string

const (
	ReportDaily   ReportType = "daily"
	ReportWeekly  ReportType = "weekly"
	ReportMonthly ReportType = "monthly"
)

type ReportOutcome struct {
	Type     ReportType
	From     time.Time
	To       time.Time
	Income   int64
	Expense  int64
	Expenses []domain.CategorySummary
}

type GetReport struct {
	users *repository.UserRepo
	txs   *repository.TransactionRepo
}

func NewGetReport(u *repository.UserRepo, t *repository.TransactionRepo) *GetReport {
	return &GetReport{users: u, txs: t}
}

func (uc *GetReport) Rekap(ctx context.Context, jid string, rType ReportType) (*ReportOutcome, error) {
	user, err := uc.users.GetOrCreateByJID(ctx, jid, "")
	if err != nil {
		return nil, err
	}

	loc, _ := time.LoadLocation("Asia/Jakarta")
	now := time.Now().In(loc)

	var from, to time.Time
	today := time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, loc)

	switch rType {
	case ReportDaily:
		from = today
		to = today.AddDate(0, 0, 1)
	case ReportWeekly:

		wd := int(today.Weekday())
		if wd == 0 {
			wd = 7
		}
		from = today.AddDate(0, 0, -(wd - 1))
		to = from.AddDate(0, 0, 7)
	case ReportMonthly:
		from = time.Date(now.Year(), now.Month(), 1, 0, 0, 0, 0, loc)
		to = from.AddDate(0, 1, 0)
	}

	sums, err := uc.txs.SumByType(ctx, user.ID, from, to)
	if err != nil {
		return nil, err
	}

	expenses, err := uc.txs.SumByCategory(ctx, user.ID, domain.TypeExpense, from, to)
	if err != nil {
		return nil, err
	}

	return &ReportOutcome{
		Type:     rType,
		From:     from,
		To:       to,
		Income:   sums[domain.TypeIncome],
		Expense:  sums[domain.TypeExpense],
		Expenses: expenses,
	}, nil
}
