package repository

import (
	"context"

	"github.com/derispewss/gonami-projects/internal/domain"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

type CategoryRepo struct {
	db *pgxpool.Pool
}

func NewCategoryRepo(db *pgxpool.Pool) *CategoryRepo {
	return &CategoryRepo{db: db}
}

func (r *CategoryRepo) FindByNameAndType(ctx context.Context, userID uuid.UUID, name string, txType domain.TransactionType) (*domain.Category, error) {

	query := `
		SELECT id, user_id, name, type, created_at, updated_at
		FROM categories
		WHERE (user_id = $1 OR user_id IS NULL)
		  AND type = $2 AND name = $3
		ORDER BY user_id NULLS LAST
		LIMIT 1
	`

	cat := &domain.Category{}
	err := r.db.QueryRow(ctx, query, userID, txType, name).Scan(
		&cat.ID, &cat.UserID, &cat.Name, &cat.Type, &cat.CreatedAt, &cat.UpdatedAt,
	)
	if err != nil {
		if err == pgx.ErrNoRows {
			return nil, domain.ErrNotFound
		}
		return nil, err
	}

	return cat, nil
}

func (r *CategoryRepo) CreateCustom(ctx context.Context, userID uuid.UUID, name string, txType domain.TransactionType) (*domain.Category, error) {
	query := `
		INSERT INTO categories (user_id, name, type)
		VALUES ($1, $2, $3)
		ON CONFLICT (user_id, name, type) DO UPDATE SET updated_at = NOW()
		RETURNING id, user_id, name, type, created_at, updated_at
	`
	cat := &domain.Category{}
	err := r.db.QueryRow(ctx, query, userID, name, txType).Scan(
		&cat.ID, &cat.UserID, &cat.Name, &cat.Type, &cat.CreatedAt, &cat.UpdatedAt,
	)
	if err != nil {
		return nil, err
	}
	return cat, nil
}

func (r *CategoryRepo) ListAvailable(ctx context.Context, userID uuid.UUID) ([]domain.Category, error) {
	query := `
		SELECT id, user_id, name, type, created_at, updated_at
		FROM categories
		WHERE user_id IS NULL OR user_id = $1
		ORDER BY user_id NULLS LAST, name ASC
	`
	rows, err := r.db.Query(ctx, query, userID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var list []domain.Category
	for rows.Next() {
		var c domain.Category
		if err := rows.Scan(&c.ID, &c.UserID, &c.Name, &c.Type, &c.CreatedAt, &c.UpdatedAt); err != nil {
			return nil, err
		}
		list = append(list, c)
	}
	return list, rows.Err()
}
