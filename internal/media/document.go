package media

import (
	"context"
	"fmt"
	"log/slog"
	"strings"
	"time"

	"github.com/derispewss/gonami-projects/internal/ai"
)

const minStatementTextLen = 300

type DocumentProcessor struct {
	client ai.AIClient
}

func NewDocumentProcessor(client ai.AIClient) *DocumentProcessor {
	return &DocumentProcessor{client: client}
}

func (p *DocumentProcessor) Supports(mimeType string) bool {
	return IsPDFMIME(mimeType)
}

func (p *DocumentProcessor) Process(ctx context.Context, in Input) (*Output, error) {
	text, textErr := ExtractPDFText(in.Data)
	if textErr == nil && len([]rune(strings.TrimSpace(text))) >= minStatementTextLen {
		extrs, err := p.client.ExtractFromStatementTexts(ctx, text, time.Now())
		if err == nil && len(extrs) > 0 {
			receipts := toReceiptResults(extrs)
			return &Output{
				StatementText: truncateRunes(text, 500),
				Receipts:      receipts,
			}, nil
		}
		if err != nil {
			slog.Warn("ekstraksi statement text gagal — fallback ke vision", "error", err)
		}
	}

	extrs, err := p.client.ExtractReceipts(ctx, in.Data, in.MimeType)
	if err != nil {
		return nil, fmt.Errorf("ekstraksi struk pdf gagal: %w", err)
	}
	return &Output{Receipts: toReceiptResults(extrs)}, nil
}

func toReceiptResults(extrs []*ai.Extraction) []ReceiptResult {
	var receipts []ReceiptResult
	for _, ext := range extrs {
		if ext == nil || !ext.IsValid() {
			continue
		}
		desc := ext.Description
		if desc == "" && ext.Merchant != "" {
			desc = ext.Merchant
		}
		receipts = append(receipts, ReceiptResult{
			Type:        ext.Type,
			Amount:      ext.Amount,
			Description: desc,
			Merchant:    ext.Merchant,
			Category:    ext.CategoryHint,
			DateHint:    ext.DateHint,
			Confidence:  ext.Confidence,
		})
	}
	return receipts
}

func truncateRunes(s string, max int) string {
	r := []rune(s)
	if len(r) <= max {
		return s
	}
	return string(r[:max]) + "…"
}
