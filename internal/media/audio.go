package media

import (
	"context"
	"fmt"
	"strings"

	"github.com/derispewss/gonami-projects/internal/ai"
)

type AudioProcessor struct {
	client ai.AIClient
}

func NewAudioProcessor(client ai.AIClient) *AudioProcessor {
	return &AudioProcessor{client: client}
}

func (p *AudioProcessor) Supports(mimeType string) bool {
	return strings.HasPrefix(baseMIME(mimeType), "audio/")
}

func (p *AudioProcessor) Process(ctx context.Context, in Input) (*Output, error) {
	transcript, err := p.client.TranscribeAudio(ctx, in.Data, in.MimeType)
	if err != nil {
		return nil, fmt.Errorf("transkripsi audio gagal: %w", err)
	}
	return &Output{Transcript: transcript}, nil
}
