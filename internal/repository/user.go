package repository

import (
	"context"

	"github.com/derispewss/gonami-projects/internal/domain"
	"github.com/jackc/pgx/v5/pgxpool"
)

type UserRepo struct {
	db *pgxpool.Pool
}

func NewUserRepo(db *pgxpool.Pool) *UserRepo {
	return &UserRepo{db: db}
}

func (r *UserRepo) GetOrCreateByJID(ctx context.Context, jid, name string) (*domain.User, error) {
	query := `
		INSERT INTO users (whatsapp_jid, name, currency)
		VALUES ($1, $2, 'IDR')
		ON CONFLICT (whatsapp_jid) DO UPDATE SET
			name = EXCLUDED.name,
			updated_at = NOW()
		RETURNING id, whatsapp_jid, name, currency, created_at, updated_at
	`

	user := &domain.User{}
	err := r.db.QueryRow(ctx, query, jid, name).Scan(
		&user.ID, &user.WhatsAppJID, &user.Name, &user.Currency,
		&user.CreatedAt, &user.UpdatedAt,
	)
	if err != nil {
		return nil, err
	}

	return user, nil
}
