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
		ext, err := p.client.ExtractFromStatementText(ctx, text, time.Now())
		if err == nil {
			return &Output{
				StatementText: truncateRunes(text, 500),
				Receipt:       toReceiptResult(ext),
			}, nil
		}
		slog.Warn("ekstraksi statement text gagal — fallback ke vision", "error", err)
	}

	rec, err := p.client.ExtractReceipt(ctx, in.Data, in.MimeType)
	if err != nil {
		return nil, fmt.Errorf("ekstraksi struk pdf gagal: %w", err)
	}
	return &Output{Receipt: toReceiptResult(rec)}, nil
}

func toReceiptResult(ext *ai.Extraction) *ReceiptResult {
	if ext == nil || !ext.IsValid() {
		return nil
	}
	desc := ext.Description
	if desc == "" && ext.Merchant != "" {
		desc = ext.Merchant
	}
	return &ReceiptResult{
		Type:        ext.Type,
		Amount:      ext.Amount,
		Description: desc,
		Merchant:    ext.Merchant,
		Category:    ext.CategoryHint,
		DateHint:    ext.DateHint,
		Confidence:  ext.Confidence,
	}
}

func truncateRunes(s string, max int) string {
	r := []rune(s)
	if len(r) <= max {
		return s
	}
	return string(r[:max]) + "…"
}
