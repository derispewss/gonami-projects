package ai

import (
	"context"
	"time"
)

type Extraction struct {
	IsTransaction bool    `json:"is_transaction"`
	Type          string  `json:"type"`
	Amount        int64   `json:"amount"`
	Description   string  `json:"description"`
	Merchant      string  `json:"merchant"`
	CategoryHint  string  `json:"category_hint"`
	DateHint      string  `json:"date_hint"`
	Confidence    float64 `json:"confidence"`
}

func (r *Extraction) IsValid() bool {
	return r.IsTransaction && r.Amount > 0 && r.Type != "" && validTypes[r.Type]
}

type AIClient interface {
	TranscribeAudio(ctx context.Context, data []byte, mimeType string) (string, error)

	ExtractReceipts(ctx context.Context, data []byte, mimeType string) ([]*Extraction, error)

	ExtractFromChatText(ctx context.Context, text string, now time.Time) (*Extraction, error)

	ExtractFromStatementTexts(ctx context.Context, text string, now time.Time) ([]*Extraction, error)
}
