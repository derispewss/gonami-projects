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
	rec, err := p.client.ExtractReceipt(ctx, in.Data, in.MimeType)
	if err != nil {
		return nil, fmt.Errorf("ekstraksi struk gagal: %w", err)
	}
	if rec == nil || !rec.IsValid() {

		return &Output{}, nil
	}
	return &Output{Receipt: &ReceiptResult{
		Type:        rec.Type,
		Amount:      rec.Amount,
		Description: rec.Description,
		Merchant:    rec.Merchant,
		Category:    rec.CategoryHint,
		Confidence:  rec.Confidence,
	}}, nil
}
