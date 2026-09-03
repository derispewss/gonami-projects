package media

import (
	"context"
	"strings"
)

type Input struct {
	Data     []byte
	MimeType string
}

type Output struct {
	Transcript    string
	Receipts      []ReceiptResult
	StatementText string
}

type ReceiptResult struct {
	Type        string
	Amount      int64
	Description string
	Merchant    string
	Category    string
	DateHint    string
	Confidence  float64
}

type Processor interface {
	Supports(mimeType string) bool
	Process(ctx context.Context, in Input) (*Output, error)
}

func baseMIME(mimeType string) string {
	if i := strings.IndexByte(mimeType, ';'); i >= 0 {
		mimeType = mimeType[:i]
	}
	return strings.TrimSpace(strings.ToLower(mimeType))
}

func isImageMIME(mimeType string) bool {
	switch baseMIME(mimeType) {
	case "image/jpeg", "image/png", "image/webp", "image/heic", "image/heif":
		return true
	}
	return false
}
