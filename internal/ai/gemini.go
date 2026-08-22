package ai

import (
	"context"
	_ "embed"
	"fmt"
	"log/slog"
	"strings"
	"time"

	"google.golang.org/genai"
)

var (
	receiptPrompt string

	receiptDocumentPrompt string

	fallbackChatPrompt string

	statementTextPrompt string
)

type Gemini struct {
	client  *genai.Client
	model   string
	txModel string
	budget  *TokenSaver
	maxOut  int32
}

func NewGemini(apiKey, model, txModel string, budget *TokenSaver, maxOutputTokens int32) (*Gemini, error) {
	cli, err := genai.NewClient(context.Background(), &genai.ClientConfig{
		APIKey:  apiKey,
		Backend: genai.BackendGeminiAPI,
	})
	if err != nil {
		return nil, fmt.Errorf("gagal inisialisasi gemini client: %w", err)
	}
	if model == "" {
		model = "gemini-2.0-flash"
	}
	return &Gemini{
		client:  cli,
		model:   model,
		txModel: txModel,
		budget:  budget,
		maxOut:  maxOutputTokens,
	}, nil
}

func cleanMIME(mimeType string) string {
	if i := strings.IndexByte(mimeType, ';'); i >= 0 {
		mimeType = mimeType[:i]
	}
	return strings.TrimSpace(mimeType)
}

func withToday(prompt string, now time.Time) string {
	today := now.In(wibLoc).Format("2006-01-02")
	return strings.ReplaceAll(prompt, "{{TODAY}}", today)
}

func (g *Gemini) generate(ctx context.Context, parts []*genai.Part, cfg *genai.GenerateContentConfig) (string, error) {
	if g.budget != nil && !g.budget.Allow() {
		slog.Warn("token saver: kuota harian habis — panggilan LLM ditolak")
		return "", ErrBudgetExceeded
	}
	resp, err := g.client.Models.GenerateContent(ctx, g.model,
		[]*genai.Content{{Parts: parts}}, cfg)
	if err != nil {
		return "", err
	}
	return resp.Text(), nil
}

func (g *Gemini) generateText(ctx context.Context, prompt string) (string, error) {
	model := g.txModel
	if model == "" {
		model = g.model
	}
	if g.budget != nil && !g.budget.Allow() {
		slog.Warn("token saver: kuota harian habis — panggilan LLM ditolak")
		return "", ErrBudgetExceeded
	}
	cfg := &genai.GenerateContentConfig{
		Temperature:      ptr(float32(0.1)),
		ResponseMIMEType: "application/json",
	}
	if g.maxOut > 0 {
		cfg.MaxOutputTokens = g.maxOut
	}
	resp, err := g.client.Models.GenerateContent(ctx, model,
		[]*genai.Content{{Parts: []*genai.Part{{Text: prompt}}}}, cfg)
	if err != nil {
		return "", err
	}
	return resp.Text(), nil
}

func (g *Gemini) TranscribeAudio(ctx context.Context, data []byte, mimeType string) (string, error) {
	parts := []*genai.Part{
		{InlineData: &genai.Blob{Data: data, MIMEType: cleanMIME(mimeType)}},
		{Text: sttPrompt},
	}
	out, err := g.generate(ctx, parts, &genai.GenerateContentConfig{Temperature: ptr(float32(0.1))})
	if err != nil {
		return "", fmt.Errorf("gemini stt gagal: %w", err)
	}
	return strings.TrimSpace(out), nil
}

func (g *Gemini) ExtractReceipt(ctx context.Context, data []byte, mimeType string) (*Extraction, error) {
	mime := cleanMIME(mimeType)
	prompt := receiptPrompt
	if strings.HasPrefix(mime, "application/") {
		prompt = withToday(receiptDocumentPrompt, time.Now())
	}
	parts := []*genai.Part{
		{InlineData: &genai.Blob{Data: data, MIMEType: mime}},
		{Text: prompt},
	}
	out, err := g.generate(ctx, parts, &genai.GenerateContentConfig{
		Temperature:      ptr(float32(0.1)),
		ResponseMIMEType: "application/json",
	})
	if err != nil {
		return nil, fmt.Errorf("gemini vision gagal: %w", err)
	}

	rec, err := ParseExtractionJSON(out)
	if err != nil {
		return nil, fmt.Errorf("output gemini tidak valid: %w", err)
	}
	return rec, nil
}

func (g *Gemini) ExtractFromChatText(ctx context.Context, text string, now time.Time) (*Extraction, error) {
	out, err := g.generateText(ctx, withToday(fallbackChatPrompt, now)+"\n\nMESSAGE:\n"+text)
	if err != nil {
		return nil, fmt.Errorf("llm fallback gagal: %w", err)
	}
	rec, err := ParseExtractionJSON(out)
	if err != nil {
		return nil, fmt.Errorf("output llm tidak valid: %w", err)
	}
	return rec, nil
}

func (g *Gemini) ExtractFromStatementText(ctx context.Context, text string, now time.Time) (*Extraction, error) {
	out, err := g.generateText(ctx, withToday(statementTextPrompt, now)+"\n\nSTATEMENT TEXT:\n"+text)
	if err != nil {
		return nil, fmt.Errorf("ekstraksi statement gagal: %w", err)
	}
	rec, err := ParseExtractionJSON(out)
	if err != nil {
		return nil, fmt.Errorf("output tidak valid: %w", err)
	}
	return rec, nil
}

const sttPrompt = `Transcribe this voice note. It is an Indonesian personal finance message.
Output ONLY the transcript text verbatim in Indonesian, with no commentary, no quotes, and no markdown.`

func ptr[T any](v T) *T { return &v }
