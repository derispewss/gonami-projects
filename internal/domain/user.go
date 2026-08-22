package domain

import (
	"time"

	"github.com/google/uuid"
)

type User struct {
	ID             uuid.UUID
	WhatsAppJID    string
	Name           string
	Currency       string
	ActiveWalletID *uuid.UUID
	CreatedAt      time.Time
	UpdatedAt      time.Time
}
