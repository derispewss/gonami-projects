package application

import (
	"context"
	"sort"
	"time"

	"github.com/derispewss/gonami-projects/internal/repository"
)

type InsightUC struct {
	users *repository.UserRepo
	txs   *repository.TransactionRepo
}

func NewInsightUC(u *repository.UserRepo, t *repository.TransactionRepo) *InsightUC {
	return &InsightUC{users: u, txs: t}
}

type Anomaly struct {
	Category   string
	Current    int64
	Average    int64
	Multiplier float64
}

type InsightsResult struct {
	Month       time.Time
	Anomalies   []Anomaly
	Recurring   []repository.RecurringSubscription
	HasInsights bool
}

func (uc *InsightUC) Get(ctx context.Context, jid string) (*InsightsResult, error) {
	user, err := uc.users.GetOrCreateByJID(ctx, jid, "")
	if err != nil {
		return nil, err
	}

	rows, err := uc.txs.SumByCategoryMonthly(ctx, user.ID, 4)
	if err != nil {
		return nil, err
	}

	loc, _ := time.LoadLocation("Asia/Jakarta")
	now := time.Now().In(loc)
	thisMonth := time.Date(now.Year(), now.Month(), 1, 0, 0, 0, 0, loc)

	type acc struct {
		sum     int64
		count   int
		current int64
	}
	stats := make(map[string]*acc)
	for _, r := range rows {
		a, ok := stats[r.Category]
		if !ok {
			a = &acc{}
			stats[r.Category] = a
		}
		month := r.Month.In(loc)
		if month.Equal(thisMonth) {
			a.current += r.Total
		} else {
			a.sum += r.Total
			a.count++
		}
	}

	res := &InsightsResult{Month: thisMonth}
	for cat, a := range stats {
		if a.current == 0 || a.count == 0 {
			continue
		}
		avg := a.sum / int64(a.count)
		if avg <= 0 {
			continue
		}
		mult := float64(a.current) / float64(avg)
		if mult >= 1.5 {
			res.Anomalies = append(res.Anomalies, Anomaly{
				Category:   cat,
				Current:    a.current,
				Average:    avg,
				Multiplier: mult,
			})
		}
	}
	sort.Slice(res.Anomalies, func(i, j int) bool {
		return res.Anomalies[i].Multiplier > res.Anomalies[j].Multiplier
	})
	if len(res.Anomalies) > 5 {
		res.Anomalies = res.Anomalies[:5]
	}

	recurring, err := uc.txs.FindRecurring(ctx, user.ID)
	if err != nil {
		return nil, err
	}
	res.Recurring = recurring

	res.HasInsights = len(res.Anomalies) > 0 || len(res.Recurring) > 0
	return res, nil
}
