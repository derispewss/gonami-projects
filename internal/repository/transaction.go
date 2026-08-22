package repository

import (
	"context"
	"time"

	"github.com/derispewss/gonami-projects/internal/domain"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

type TransactionRepo struct {
	db *pgxpool.Pool
}

func NewTransactionRepo(db *pgxpool.Pool) *TransactionRepo {
	return &TransactionRepo{db: db}
}

func (r *TransactionRepo) Create(ctx context.Context, tx *domain.Transaction) error {
	query := `
		INSERT INTO transactions (
			user_id, type, amount, description, category_id,
			merchant, transaction_date, source_type, raw_message, wallet_id
		)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10)
		RETURNING id, created_at, updated_at
	`
	err := r.db.QueryRow(ctx, query,
		tx.UserID, tx.Type, tx.Amount, tx.Description, tx.CategoryID,
		tx.Merchant, tx.TransactionDate, tx.SourceType, tx.RawMessage, tx.WalletID,
	).Scan(&tx.ID, &tx.CreatedAt, &tx.UpdatedAt)
	return err
}

func (r *TransactionRepo) GetLatest(ctx context.Context, userID uuid.UUID, limit int) ([]*domain.Transaction, error) {
	query := `
		SELECT 
			t.id, t.user_id, t.type, t.amount, t.description, 
			t.category_id, c.name, t.merchant, t.transaction_date, 
			t.source_type, t.created_at, t.updated_at
		FROM transactions t
		LEFT JOIN categories c ON c.id = t.category_id
		WHERE t.user_id = $1
		ORDER BY t.created_at DESC
		LIMIT $2
	`
	rows, err := r.db.Query(ctx, query, userID, limit)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var list []*domain.Transaction
	for rows.Next() {
		tx := &domain.Transaction{}
		var catName *string
		err := rows.Scan(
			&tx.ID, &tx.UserID, &tx.Type, &tx.Amount, &tx.Description,
			&tx.CategoryID, &catName, &tx.Merchant, &tx.TransactionDate,
			&tx.SourceType, &tx.CreatedAt, &tx.UpdatedAt,
		)
		if err != nil {
			return nil, err
		}
		if catName != nil {
			tx.CategoryName = *catName
		}
		list = append(list, tx)
	}
	return list, rows.Err()
}

func (r *TransactionRepo) DeleteLast(ctx context.Context, userID uuid.UUID) (*domain.Transaction, error) {
	query := `
		DELETE FROM transactions 
		WHERE id = (
			SELECT id FROM transactions 
			WHERE user_id = $1 
			ORDER BY created_at DESC LIMIT 1
		)
		RETURNING id, type, amount, description
	`
	tx := &domain.Transaction{}
	err := r.db.QueryRow(ctx, query, userID).Scan(&tx.ID, &tx.Type, &tx.Amount, &tx.Description)
	if err != nil {
		if err == pgx.ErrNoRows {
			return nil, domain.ErrNotFound
		}
		return nil, err
	}
	return tx, nil
}

func (r *TransactionRepo) SumByType(ctx context.Context, userID uuid.UUID, from, to time.Time) (map[domain.TransactionType]int64, error) {
	query := `
		SELECT type, COALESCE(SUM(amount), 0)
		FROM transactions
		WHERE user_id = $1 AND transaction_date >= $2 AND transaction_date < $3
		GROUP BY type
	`
	rows, err := r.db.Query(ctx, query, userID, from, to)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	sums := make(map[domain.TransactionType]int64)
	for rows.Next() {
		var t string
		var total int64
		if err := rows.Scan(&t, &total); err != nil {
			return nil, err
		}
		sums[domain.TransactionType(t)] = total
	}
	return sums, rows.Err()
}

func (r *TransactionRepo) SumByCategory(ctx context.Context, userID uuid.UUID, txType domain.TransactionType, from, to time.Time) ([]domain.CategorySummary, error) {
	query := `
		SELECT COALESCE(c.name, 'Lainnya') as cat_name, COALESCE(SUM(t.amount), 0) as total
		FROM transactions t
		LEFT JOIN categories c ON c.id = t.category_id
		WHERE t.user_id = $1 AND t.type = $2 AND t.transaction_date >= $3 AND t.transaction_date < $4
		GROUP BY c.name
		ORDER BY total DESC
	`
	rows, err := r.db.Query(ctx, query, userID, txType, from, to)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var list []domain.CategorySummary
	for rows.Next() {
		var s domain.CategorySummary
		if err := rows.Scan(&s.CategoryName, &s.Total); err != nil {
			return nil, err
		}
		list = append(list, s)
	}
	return list, rows.Err()
}

func (r *TransactionRepo) Balance(ctx context.Context, userID uuid.UUID) (income, expense, net int64, err error) {
	query := `
		SELECT type, COALESCE(SUM(amount), 0)
		FROM transactions
		WHERE user_id = $1 AND type IN ('income', 'expense')
		GROUP BY type
	`
	rows, err := r.db.Query(ctx, query, userID)
	if err != nil {
		return 0, 0, 0, err
	}
	defer rows.Close()

	for rows.Next() {
		var t string
		var total int64
		if err := rows.Scan(&t, &total); err != nil {
			return 0, 0, 0, err
		}
		if t == "income" {
			income = total
		} else if t == "expense" {
			expense = total
		}
	}
	net = income - expense
	return income, expense, net, rows.Err()
}

func (r *TransactionRepo) ListByRange(ctx context.Context, userID uuid.UUID, from, to time.Time) ([]*domain.Transaction, error) {
	query := `
		SELECT id, user_id, type, amount, description, category_id,
		       merchant, transaction_date, source_type
		FROM transactions
		WHERE user_id = $1 AND transaction_date >= $2 AND transaction_date < $3
		ORDER BY transaction_date ASC, created_at ASC
	`
	rows, err := r.db.Query(ctx, query, userID, from, to)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var list []*domain.Transaction
	for rows.Next() {
		tx := &domain.Transaction{}
		if err := rows.Scan(&tx.ID, &tx.UserID, &tx.Type, &tx.Amount, &tx.Description,
			&tx.CategoryID, &tx.Merchant, &tx.TransactionDate, &tx.SourceType); err != nil {
			return nil, err
		}
		list = append(list, tx)
	}
	return list, rows.Err()
}

type MonthlyCategoryTotal struct {
	Month      time.Time
	CategoryID *uuid.UUID
	Category   string
	Total      int64
}

func (r *TransactionRepo) SumByCategoryMonthly(ctx context.Context, userID uuid.UUID, months int) ([]MonthlyCategoryTotal, error) {
	query := `
		SELECT date_trunc('month', t.transaction_date)::date AS month,
		       t.category_id,
		       COALESCE(c.name, 'Lainnya') AS cat_name,
		       SUM(t.amount) AS total
		FROM transactions t
		LEFT JOIN categories c ON c.id = t.category_id
		WHERE t.user_id = $1 AND t.type = 'expense'
		  AND t.transaction_date >= date_trunc('month', NOW()) - make_interval(months => $2::int - 1)
		GROUP BY month, t.category_id, c.name
		ORDER BY month ASC, total DESC
	`
	rows, err := r.db.Query(ctx, query, userID, months)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var list []MonthlyCategoryTotal
	for rows.Next() {
		var m MonthlyCategoryTotal
		if err := rows.Scan(&m.Month, &m.CategoryID, &m.Category, &m.Total); err != nil {
			return nil, err
		}
		list = append(list, m)
	}
	return list, rows.Err()
}

type RecurringSubscription struct {
	Description string
	Count       int
	AvgAmount   int64
	LastSeen    time.Time
}

func (r *TransactionRepo) FindRecurring(ctx context.Context, userID uuid.UUID) ([]RecurringSubscription, error) {
	query := `
		SELECT LOWER(description) AS d,
		       COUNT(*) AS cnt,
		       (AVG(amount))::bigint AS avg_amt,
		       MAX(transaction_date) AS last_seen
		FROM transactions
		WHERE user_id = $1 AND type = 'expense'
		  AND transaction_date >= date_trunc('month', NOW()) - make_interval(months => 5)
		  AND COALESCE(NULLIF(TRIM(description), ''), '') <> ''
		GROUP BY d
		HAVING COUNT(DISTINCT date_trunc('month', transaction_date)) >= 2
		   AND COUNT(*) >= 2
		   AND MIN(amount)::float / NULLIF(MAX(amount), 0) >= 0.8
		ORDER BY avg_amt DESC
		LIMIT 10
	`
	rows, err := r.db.Query(ctx, query, userID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var list []RecurringSubscription
	for rows.Next() {
		var s RecurringSubscription
		if err := rows.Scan(&s.Description, &s.Count, &s.AvgAmount, &s.LastSeen); err != nil {
			return nil, err
		}
		list = append(list, s)
	}
	return list, rows.Err()
}

func (r *TransactionRepo) BalanceByWallet(ctx context.Context, userID, walletID uuid.UUID, from, to time.Time) (income, expense int64, err error) {
	query := `
		SELECT type, COALESCE(SUM(amount), 0)
		FROM transactions
		WHERE user_id = $1 AND wallet_id = $2
		  AND transaction_date >= $3 AND transaction_date < $4
		  AND type IN ('income', 'expense')
		GROUP BY type
	`
	rows, err := r.db.Query(ctx, query, userID, walletID, from, to)
	if err != nil {
		return 0, 0, err
	}
	defer rows.Close()

	for rows.Next() {
		var t string
		var total int64
		if err := rows.Scan(&t, &total); err != nil {
			return 0, 0, err
		}
		if t == "income" {
			income = total
		} else if t == "expense" {
			expense = total
		}
	}
	return income, expense, rows.Err()
}
