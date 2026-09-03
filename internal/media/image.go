package media

import (
	"context"
	"fmt"

	"github.com/derispewss/gonami-projects/internal/ai"
)

type ImageProcessor struct {
	client ai.AIClient
}

func NewImageProcessor(client ai.AIClient) *ImageProcessor {
	return &ImageProcessor{client: client}
}

func (p *ImageProcessor) Supports(mimeType string) bool {
	return isImageMIME(mimeType)
}

func (p *ImageProcessor) Process(ctx context.Context, in Input) (*Output, error) {
	if len(in.Data) == 0 {
		return nil, fmt.Errorf("data gambar kosong (0 bytes)")
	}
	extrs, err := p.client.ExtractReceipts(ctx, in.Data, in.MimeType)
	if err != nil {
		return nil, fmt.Errorf("ekstraksi struk gagal: %w", err)
	}
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
	return &Output{Receipts: receipts}, nil
}
