package ai

import (
	"context"
	"errors"
	"testing"
	"time"
)

func TestShouldFallback(t *testing.T) {
	tests := []struct {
		name string
		err  error
		want bool
	}{
		{"nil tidak fallback", nil, false},
		{"budget habis harus fallback", ErrBudgetExceeded, true},
		{"ratelimit 429", errors.New("gemini: http 429: rate limit"), true},
		{"quota exhausted", errors.New("resource exhausted: quota"), true},
		{"503 unavailable", errors.New("http 503 Service Unavailable"), true},
		{"network connection refused", errors.New("Post: dial tcp: connection refused"), true},
		{"timeout deadline", errors.New("context deadline exceeded"), true},
		{"bad input tidak fallback", errors.New("output tidak valid: bukan JSON"), false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := shouldFallback(tt.err); got != tt.want {
				t.Errorf("shouldFallback(%v) = %v, want %v", tt.err, got, tt.want)
			}
		})
	}
}

type stubPrimary struct {
	err error
}

func (s *stubPrimary) TranscribeAudio(ctx context.Context, data []byte, mimeType string) (string, error) {
	return "", s.err
}
func (s *stubPrimary) ExtractReceipts(ctx context.Context, data []byte, mimeType string) ([]*Extraction, error) {
	return nil, s.err
}
func (s *stubPrimary) ExtractFromChatText(ctx context.Context, text string, now time.Time) (*Extraction, error) {
	return nil, s.err
}
func (s *stubPrimary) ExtractFromStatementTexts(ctx context.Context, text string, now time.Time) ([]*Extraction, error) {
	return nil, s.err
}

type stubFallback struct {
	text string
}

func (s *stubFallback) TranscribeAudio(ctx context.Context, data []byte, mimeType string) (string, error) {
	return s.text, nil
}
func (s *stubFallback) ExtractReceipts(ctx context.Context, data []byte, mimeType string) ([]*Extraction, error) {
	return nil, ErrVisionUnsupported
}
func (s *stubFallback) ExtractFromChatText(ctx context.Context, text string, now time.Time) (*Extraction, error) {
	return &Extraction{IsTransaction: true, Amount: 1000, Type: "expense", Description: "kopi", Confidence: 0.9}, nil
}
func (s *stubFallback) ExtractFromStatementTexts(ctx context.Context, text string, now time.Time) ([]*Extraction, error) {
	return nil, nil
}

func TestFallbackClientText(t *testing.T) {
	fc := NewFallbackClient(&stubPrimary{err: ErrBudgetExceeded}, &stubFallback{})
	rec, err := fc.ExtractFromChatText(context.Background(), "beli kopi 15k", time.Now())
	if err != nil {
		t.Fatalf("unexpected err: %v", err)
	}
	if rec == nil || rec.Amount != 1000 {
		t.Fatalf("expected fallback extraction, got %+v", rec)
	}
}

func TestFallbackClientPrimaryOK(t *testing.T) {
	fc := NewFallbackClient(&stubPrimary{err: nil}, &stubFallback{})
	out, err := fc.TranscribeAudio(context.Background(), nil, "audio/mpeg")
	if err != nil {
		t.Fatalf("unexpected err: %v", err)
	}
	if out != "" {
		t.Fatalf("expected empty primary output, got %q", out)
	}
}

func TestFallbackClientVisionUnsupported(t *testing.T) {
	fc := NewFallbackClient(&stubPrimary{err: ErrBudgetExceeded}, &stubFallback{})
	_, err := fc.ExtractReceipts(context.Background(), nil, "image/jpeg")
	if !errors.Is(err, ErrVisionExhausted) {
		t.Fatalf("expected ErrVisionExhausted, got %v", err)
	}
}
