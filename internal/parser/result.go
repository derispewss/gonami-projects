package parser

import (
	"time"

	"github.com/derispewss/gonami-projects/internal/domain"
)

type Result struct {
	Type        domain.TransactionType `json:"type"`
	Amount      int64                  `json:"amount"`
	Description string                 `json:"description"`
	Category    string                 `json:"category,omitempty"`
	Merchant    string                 `json:"merchant,omitempty"`
	Destination string                 `json:"destination,omitempty"`
	Date        time.Time              `json:"date"`
	Confidence  float64                `json:"confidence"`
}

const (
	confidenceAutoSave   = 0.80
	confidenceAskConfirm = 0.50
)

func (r *Result) ShouldAutoSave() bool {
	return r.Confidence >= confidenceAutoSave
}

func (r *Result) NeedsConfirmation() bool {
	return r.Confidence >= confidenceAskConfirm && r.Confidence < confidenceAutoSave
}
