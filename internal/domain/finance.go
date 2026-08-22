package domain

import (
	"time"

	"github.com/google/uuid"
)

type Budget struct {
	ID           uuid.UUID
	UserID       uuid.UUID
	CategoryName string
	MonthlyLimit int64
	CreatedAt    time.Time
	UpdatedAt    time.Time
}

type Wallet struct {
	ID        uuid.UUID
	UserID    uuid.UUID
	Name      string
	CreatedAt time.Time
}
