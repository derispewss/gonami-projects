package domain

import (
	"time"

	"github.com/google/uuid"
)

type TransactionType string

const (
	TypeExpense  TransactionType = "expense"
	TypeIncome   TransactionType = "income"
	TypeTransfer TransactionType = "transfer"
)

func (t TransactionType) IsValid() bool {
	switch t {
	case TypeExpense, TypeIncome, TypeTransfer:
		return true
	}
	return false
}

type SourceType string

const (
	SourceText  SourceType = "text"
	SourceAudio SourceType = "audio"
	SourceImage SourceType = "image"
	SourcePDF   SourceType = "pdf"
)

func (s SourceType) IsValid() bool {
	switch s {
	case SourceText, SourceAudio, SourceImage, SourcePDF:
		return true
	}
	return false
}

type Transaction struct {
	ID              uuid.UUID
	UserID          uuid.UUID
	Type            TransactionType
	Amount          int64
	Description     string
	CategoryID      *uuid.UUID
	CategoryName    string
	Merchant        string
	WalletID        *uuid.UUID
	TransactionDate time.Time
	SourceType      SourceType
	SourceMessageID string
	RawMessage      string
	CreatedAt       time.Time
	UpdatedAt       time.Time
}

type TransactionItem struct {
	ID            uuid.UUID
	TransactionID uuid.UUID
	Name          string
	Quantity      int
	Amount        int64
	CreatedAt     time.Time
}

type CategorySummary struct {
	CategoryName string
	Total        int64
}
