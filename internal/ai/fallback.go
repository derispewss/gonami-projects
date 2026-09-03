package ai

import (
	"context"
	"errors"
	"log/slog"
	"net/http"
	"strconv"
	"strings"
	"time"
)

// ErrVisionUnsupported indicates the fallback provider cannot process vision
// media (e.g. Groq has no image model).
var ErrVisionUnsupported = errors.New("penyedia cadangan tidak mendukung vision")

// ErrVisionExhausted indicates the vision pipeline could not run because the
// primary provider is exhausted and no vision-capable fallback is available.
// Callers use this to surface a friendly message to the user.
var ErrVisionExhausted = errors.New("fitur baca media tidak tersedia saat ini")

// FallbackClient wraps a primary AIClient and a fallback AIClient. When the
// primary fails with a transient/quota/network error (or budget exhaustion),
// the same call is retried on the fallback provider so the bot keeps working
// regardless of which single API is unavailable.
type FallbackClient struct {
	primary  AIClient
	fallback AIClient
}

func NewFallbackClient(primary, fallback AIClient) *FallbackClient {
	return &FallbackClient{primary: primary, fallback: fallback}
}

// shouldFallback reports whether a primary error should trigger the fallback
// provider (quota/ratelimit/network/5xx/budget) vs a hard error (bad input).
func shouldFallback(err error) bool {
	if err == nil {
		return false
	}
	if errors.Is(err, ErrBudgetExceeded) {
		return true
	}

	msg := strings.ToLower(err.Error())
	httpErr := httpErrCode(msg)
	if httpErr == http.StatusTooManyRequests || httpErr == http.StatusServiceUnavailable ||
		httpErr == http.StatusGatewayTimeout || httpErr == http.StatusBadGateway {
		return true
	}

	for _, needle := range []string{
		"429", "rate limit", "quota", "resource exhausted", "limit",
		"too many requests", "overloaded", "unavailable", "temporarily",
		"timeout", "timed out", "connection refused", "connect:", "tls",
		"503", "502", "504", "500 internal", "deadline exceeded",
	} {
		if strings.Contains(msg, needle) {
			return true
		}
	}
	return false
}

// httpErrCode sniffs a status code embedded in an error message.
func httpErrCode(msg string) int {
	idx := strings.Index(msg, "http ")
	if idx < 0 {
		return 0
	}
	rest := msg[idx+5:]
	spaceIdx := strings.IndexByte(rest, ' ')
	if spaceIdx < 0 {
		spaceIdx = len(rest)
	}
	var code int
	var err error
	code, err = strconv.Atoi(rest[:spaceIdx])
	if err != nil {
		return 0
	}
	return code
}

func (f *FallbackClient) TranscribeAudio(ctx context.Context, data []byte, mimeType string) (string, error) {
	out, err := f.primary.TranscribeAudio(ctx, data, mimeType)
	if err != nil && f.fallback != nil && shouldFallback(err) {
		slog.Warn("primary transkripsi gagal — memakai fallback", "err", err)
		return f.fallback.TranscribeAudio(ctx, data, mimeType)
	}
	return out, err
}

func (f *FallbackClient) ExtractReceipts(ctx context.Context, data []byte, mimeType string) ([]*Extraction, error) {
	out, err := f.primary.ExtractReceipts(ctx, data, mimeType)
	if err != nil && f.fallback != nil && shouldFallback(err) {
		slog.Warn("primary vision gagal — mencoba fallback", "err", err)
		out, ferr := f.fallback.ExtractReceipts(ctx, data, mimeType)
		if ferr == nil {
			return out, nil
		}
		if errors.Is(ferr, ErrVisionUnsupported) {
			// fallback can't see images; report vision unavailable distinctly
			return nil, ErrVisionExhausted
		}
		return nil, errors.Join(err, ferr)
	}
	return out, err
}

func (f *FallbackClient) ExtractFromChatText(ctx context.Context, text string, now time.Time) (*Extraction, error) {
	out, err := f.primary.ExtractFromChatText(ctx, text, now)
	if err != nil && f.fallback != nil && shouldFallback(err) {
		slog.Warn("primary teks gagal — memakai fallback", "err", err)
		return f.fallback.ExtractFromChatText(ctx, text, now)
	}
	return out, err
}

func (f *FallbackClient) ExtractFromStatementTexts(ctx context.Context, text string, now time.Time) ([]*Extraction, error) {
	out, err := f.primary.ExtractFromStatementTexts(ctx, text, now)
	if err != nil && f.fallback != nil && shouldFallback(err) {
		slog.Warn("primary statement gagal — memakai fallback", "err", err)
		return f.fallback.ExtractFromStatementTexts(ctx, text, now)
	}
	return out, err
}

var _ AIClient = (*FallbackClient)(nil)
