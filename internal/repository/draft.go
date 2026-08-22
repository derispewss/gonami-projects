package repository

import (
	"context"

	"github.com/derispewss/finwa-projects/internal/domain"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

type DraftRepo struct {
	db *pgxpool.Pool
}

func NewDraftRepo(db *pgxpool.Pool) *DraftRepo {
	return &DraftRepo{db: db}
}

func (r *DraftRepo) Create(ctx context.Context, d *domain.TransactionDraft) error {
	query := `
		INSERT INTO transaction_drafts (
			user_id, source_type, raw_content, extracted_data,
			confidence, status, expires_at
		)
		VALUES ($1, $2, $3, $4, $5, $6, $7)
		RETURNING id, created_at, updated_at
	`
	err := r.db.QueryRow(ctx, query,
		d.UserID, d.SourceType, d.RawContent, d.ExtractedData,
		d.Confidence, d.Status, d.ExpiresAt,
	).Scan(&d.ID, &d.CreatedAt, &d.UpdatedAt)
	return err
}

func (r *DraftRepo) LatestPending(ctx context.Context, userID uuid.UUID) (*domain.TransactionDraft, error) {
	query := `
		SELECT 
			id, user_id, source_type, raw_content, extracted_data, 
			confidence, status, expires_at, created_at, updated_at
		FROM transaction_drafts
		WHERE user_id = $1 AND status = 'pending' AND expires_at > NOW()
		ORDER BY created_at DESC
		LIMIT 1
	`
	d := &domain.TransactionDraft{}
	var extractedData []byte
	err := r.db.QueryRow(ctx, query, userID).Scan(
		&d.ID, &d.UserID, &d.SourceType, &d.RawContent, &extractedData,
		&d.Confidence, &d.Status, &d.ExpiresAt, &d.CreatedAt, &d.UpdatedAt,
	)
	if err != nil {
		if err == pgx.ErrNoRows {
			return nil, domain.ErrNotFound
		}
		return nil, err
	}
	d.ExtractedData = extractedData
	return d, nil
}

func (r *DraftRepo) UpdateStatus(ctx context.Context, id uuid.UUID, from, to domain.DraftStatus) error {
	query := `
		UPDATE transaction_drafts 
		SET status = $1, updated_at = NOW() 
		WHERE id = $2 AND status = $3
	`
	cmdTag, err := r.db.Exec(ctx, query, to, id, from)
	if err != nil {
		return err
	}
	if cmdTag.RowsAffected() == 0 {
		return domain.ErrAlreadyHandled
	}
	return nil
}

func (r *DraftRepo) ExpireOldDrafts(ctx context.Context) (int64, error) {
	query := `
		UPDATE transaction_drafts 
		SET status = 'expired', updated_at = NOW() 
		WHERE status = 'pending' AND expires_at <= NOW()
	`
	cmdTag, err := r.db.Exec(ctx, query)
	if err != nil {
		return 0, err
	}
	return cmdTag.RowsAffected(), nil
}

func (r *DraftRepo) CancelAllPending(ctx context.Context, userID uuid.UUID) error {
	query := `
		UPDATE transaction_drafts 
		SET status = 'expired', updated_at = NOW() 
		WHERE user_id = $1 AND status = 'pending'
	`
	_, err := r.db.Exec(ctx, query, userID)
	return err
}
