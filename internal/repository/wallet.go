package repository

import (
	"context"

	"github.com/derispewss/gonami-projects/internal/domain"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

type WalletRepo struct {
	db *pgxpool.Pool
}

func NewWalletRepo(db *pgxpool.Pool) *WalletRepo {
	return &WalletRepo{db: db}
}

func (r *WalletRepo) Create(ctx context.Context, w *domain.Wallet) error {
	query := `
		INSERT INTO wallets (user_id, name)
		VALUES ($1, $2)
		ON CONFLICT (user_id, name) DO UPDATE SET name = EXCLUDED.name
		RETURNING id, created_at
	`
	return r.db.QueryRow(ctx, query, w.UserID, w.Name).Scan(&w.ID, &w.CreatedAt)
}

func (r *WalletRepo) GetByName(ctx context.Context, userID uuid.UUID, name string) (*domain.Wallet, error) {
	w := &domain.Wallet{}
	err := r.db.QueryRow(ctx,
		`SELECT id, user_id, name, created_at FROM wallets WHERE user_id = $1 AND LOWER(name) = LOWER($2)`,
		userID, name).
		Scan(&w.ID, &w.UserID, &w.Name, &w.CreatedAt)
	if err != nil {
		if err == pgx.ErrNoRows {
			return nil, domain.ErrNotFound
		}
		return nil, err
	}
	return w, nil
}

func (r *WalletRepo) ListByUser(ctx context.Context, userID uuid.UUID) ([]*domain.Wallet, error) {
	rows, err := r.db.Query(ctx,
		`SELECT id, user_id, name, created_at FROM wallets WHERE user_id = $1 ORDER BY created_at`,
		userID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var list []*domain.Wallet
	for rows.Next() {
		w := &domain.Wallet{}
		if err := rows.Scan(&w.ID, &w.UserID, &w.Name, &w.CreatedAt); err != nil {
			return nil, err
		}
		list = append(list, w)
	}
	return list, rows.Err()
}

func (r *WalletRepo) SetActive(ctx context.Context, userID, walletID uuid.UUID) error {
	_, err := r.db.Exec(ctx,
		`UPDATE users SET active_wallet_id = $2, updated_at = NOW() WHERE id = $1`,
		userID, walletID)
	return err
}

func (r *WalletRepo) ClearActive(ctx context.Context, userID uuid.UUID) error {
	_, err := r.db.Exec(ctx,
		`UPDATE users SET active_wallet_id = NULL, updated_at = NOW() WHERE id = $1`,
		userID)
	return err
}
