package domain

import (
	"encoding/json"
	"time"

	"github.com/google/uuid"
)

type DraftStatus string

const (
	DraftPending   DraftStatus = "pending"
	DraftConfirmed DraftStatus = "confirmed"
	DraftRejected  DraftStatus = "rejected"
	DraftExpired   DraftStatus = "expired"
)

type TransactionDraft struct {
	ID            uuid.UUID
	UserID        uuid.UUID
	SourceType    SourceType
	RawContent    string
	ExtractedData json.RawMessage
	Confidence    float64
	Status        DraftStatus
	ExpiresAt     time.Time
	CreatedAt     time.Time
	UpdatedAt     time.Time
}
