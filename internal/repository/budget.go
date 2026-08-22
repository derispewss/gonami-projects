package repository

import (
	"context"

	"github.com/derispewss/gonami-projects/internal/domain"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"
)

type BudgetRepo struct {
	db *pgxpool.Pool
}

func NewBudgetRepo(db *pgxpool.Pool) *BudgetRepo {
	return &BudgetRepo{db: db}
}

func (r *BudgetRepo) Upsert(ctx context.Context, b *domain.Budget) error {
	query := `
		INSERT INTO budgets (user_id, category_name, monthly_limit)
		VALUES ($1, $2, $3)
		ON CONFLICT (user_id, category_name) DO UPDATE SET
			monthly_limit = EXCLUDED.monthly_limit,
			updated_at = NOW()
		RETURNING id, created_at, updated_at
	`
	return r.db.QueryRow(ctx, query, b.UserID, b.CategoryName, b.MonthlyLimit).
		Scan(&b.ID, &b.CreatedAt, &b.UpdatedAt)
}

func (r *BudgetRepo) Delete(ctx context.Context, userID uuid.UUID, categoryName string) error {
	ct, err := r.db.Exec(ctx,
		`DELETE FROM budgets WHERE user_id = $1 AND LOWER(category_name) = LOWER($2)`,
		userID, categoryName)
	if err != nil {
		return err
	}
	if ct.RowsAffected() == 0 {
		return domain.ErrNotFound
	}
	return nil
}

func (r *BudgetRepo) ListByUser(ctx context.Context, userID uuid.UUID) ([]*domain.Budget, error) {
	query := `
		SELECT id, user_id, category_name, monthly_limit, created_at, updated_at
		FROM budgets
		WHERE user_id = $1
		ORDER BY monthly_limit DESC
	`
	rows, err := r.db.Query(ctx, query, userID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var list []*domain.Budget
	for rows.Next() {
		b := &domain.Budget{}
		if err := rows.Scan(&b.ID, &b.UserID, &b.CategoryName, &b.MonthlyLimit, &b.CreatedAt, &b.UpdatedAt); err != nil {
			return nil, err
		}
		list = append(list, b)
	}
	return list, rows.Err()
}
