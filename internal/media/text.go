package media

import (
	"context"
	"strings"
)

type TextProcessor struct{}

func NewTextProcessor() *TextProcessor { return &TextProcessor{} }

func (p *TextProcessor) Supports(mimeType string) bool {
	m := baseMIME(mimeType)
	return strings.HasPrefix(m, "text/") || m == "application/json"
}

func (p *TextProcessor) Process(_ context.Context, in Input) (*Output, error) {
	return &Output{Transcript: string(in.Data)}, nil
}
