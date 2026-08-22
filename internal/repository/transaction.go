package repository

import (
	"context"
	"time"

	"github.com/derispewss/finwa-projects/internal/domain"
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
			merchant, transaction_date, source_type, raw_message
		)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9)
		RETURNING id, created_at, updated_at
	`
	err := r.db.QueryRow(ctx, query,
		tx.UserID, tx.Type, tx.Amount, tx.Description, tx.CategoryID,
		tx.Merchant, tx.TransactionDate, tx.SourceType, tx.RawMessage,
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
