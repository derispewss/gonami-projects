package ai

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log/slog"
	"mime/multipart"
	"net/http"
	"strings"
	"time"
)

// Groq implements AIClient against the OpenAI-compatible Groq API. It is used
// as a cloud fallback when the primary provider (Gemini) is exhausted or down.
// Groq has no vision model, so ExtractReceipts is unsupported here.
type Groq struct {
	apiKey  string
	baseURL string
	model   string
	whisper string
	http    *http.Client
}

type GroqConfig struct {
	APIKey  string
	BaseURL string
	Model   string
	Whisper string
}

func NewGroq(cfg GroqConfig) *Groq {
	base := strings.TrimRight(cfg.BaseURL, "/")
	if base == "" {
		base = "https://api.groq.com/openai/v1"
	}
	model := cfg.Model
	if model == "" {
		model = "llama-3.3-70b-versatile"
	}
	whisper := cfg.Whisper
	if whisper == "" {
		whisper = "whisper-large-v3"
	}
	return &Groq{
		apiKey:  cfg.APIKey,
		baseURL: base,
		model:   model,
		whisper: whisper,
		http:    &http.Client{Timeout: 60 * time.Second},
	}
}

type groqMessage struct {
	Role    string `json:"role"`
	Content string `json:"content"`
}

type groqChatRequest struct {
	Model          string        `json:"model"`
	Messages       []groqMessage `json:"messages"`
	Temperature    float64       `json:"temperature,omitempty"`
	ResponseFormat *struct {
		Type string `json:"type"`
	} `json:"response_format,omitempty"`
}

type groqError struct {
	Message string `json:"message"`
	Type    string `json:"type"`
	Code    string `json:"code"`
}

type groqChatResponse struct {
	Choices []struct {
		Message struct {
			Content string `json:"content"`
		} `json:"message"`
	} `json:"choices"`
	Error *groqError `json:"error"`
}

func (g *Groq) chat(ctx context.Context, prompt string, jsonMode bool) (string, error) {
	req := groqChatRequest{
		Model:       g.model,
		Messages:    []groqMessage{{Role: "user", Content: prompt}},
		Temperature: 0.1,
	}
	if jsonMode {
		req.ResponseFormat = &struct {
			Type string `json:"type"`
		}{Type: "json_object"}
	}

	body, err := json.Marshal(req)
	if err != nil {
		return "", err
	}

	httpReq, err := http.NewRequestWithContext(ctx, http.MethodPost,
		g.baseURL+"/chat/completions", bytes.NewReader(body))
	if err != nil {
		return "", err
	}
	httpReq.Header.Set("Content-Type", "application/json")
	httpReq.Header.Set("Authorization", "Bearer "+g.apiKey)

	resp, err := g.http.Do(httpReq)
	if err != nil {
		return "", fmt.Errorf("groq request gagal: %w", err)
	}
	defer resp.Body.Close()

	raw, _ := io.ReadAll(resp.Body)
	if resp.StatusCode != http.StatusOK {
		var errResp groqChatResponse
		_ = json.Unmarshal(raw, &errResp)
		msg := "unknown"
		if errResp.Error != nil && errResp.Error.Message != "" {
			msg = errResp.Error.Message
		}
		if resp.StatusCode == http.StatusTooManyRequests {
			slog.Warn("groq quota/ratelimit", "detail", truncate(msg, 160))
		}
		return "", fmt.Errorf("groq http %d: %s", resp.StatusCode, msg)
	}

	var out groqChatResponse
	if err := json.Unmarshal(raw, &out); err != nil {
		return "", err
	}
	if len(out.Choices) == 0 {
		return "", fmt.Errorf("groq: tidak ada hasil")
	}
	return out.Choices[0].Message.Content, nil
}

func (g *Groq) TranscribeAudio(ctx context.Context, data []byte, mimeType string) (string, error) {
	var buf bytes.Buffer
	mw := multipart.NewWriter(&buf)
	if err := mw.WriteField("model", g.whisper); err != nil {
		return "", err
	}
	fw, err := mw.CreateFormFile("file", "audio."+extForMIME(mimeType))
	if err != nil {
		return "", err
	}
	if _, err := fw.Write(data); err != nil {
		return "", err
	}
	if err := mw.Close(); err != nil {
		return "", err
	}

	httpReq, err := http.NewRequestWithContext(ctx, http.MethodPost,
		g.baseURL+"/audio/transcriptions", &buf)
	if err != nil {
		return "", err
	}
	httpReq.Header.Set("Content-Type", mw.FormDataContentType())
	httpReq.Header.Set("Authorization", "Bearer "+g.apiKey)

	resp, err := g.http.Do(httpReq)
	if err != nil {
		return "", fmt.Errorf("groq stt gagal: %w", err)
	}
	defer resp.Body.Close()

	raw, _ := io.ReadAll(resp.Body)
	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("groq stt http %d: %s", resp.StatusCode, truncate(string(raw), 200))
	}

	var res struct {
		Text string `json:"text"`
	}
	if err := json.Unmarshal(raw, &res); err != nil {
		return "", err
	}
	return strings.TrimSpace(res.Text), nil
}

func (g *Groq) chatToJSON(ctx context.Context, prompt string) (string, error) {
	return g.chat(ctx, prompt, true)
}

func (g *Groq) ExtractReceipts(ctx context.Context, data []byte, mimeType string) ([]*Extraction, error) {
	return nil, ErrVisionUnsupported
}

func (g *Groq) ExtractFromChatText(ctx context.Context, text string, now time.Time) (*Extraction, error) {
	out, err := g.chatToJSON(ctx, withToday(fallbackChatPrompt, now)+"\n\nMESSAGE:\n"+text)
	if err != nil {
		return nil, fmt.Errorf("groq llm fallback gagal: %w", err)
	}
	rec, err := ParseExtractionJSON(out)
	if err != nil {
		return nil, fmt.Errorf("groq output llm tidak valid: %w", err)
	}
	return rec, nil
}

func (g *Groq) ExtractFromStatementTexts(ctx context.Context, text string, now time.Time) ([]*Extraction, error) {
	out, err := g.chatToJSON(ctx, withToday(statementTextPrompt, now)+"\n\nSTATEMENT TEXT:\n"+text)
	if err != nil {
		return nil, fmt.Errorf("groq ekstraksi statement gagal: %w", err)
	}
	list, err := ParseExtractionJSONArray(out)
	if err != nil {
		return nil, fmt.Errorf("groq output statement tidak valid: %w", err)
	}
	return list, nil
}

func extForMIME(mimeType string) string {
	if i := strings.IndexByte(mimeType, ';'); i >= 0 {
		mimeType = mimeType[:i]
	}
	switch strings.TrimSpace(strings.ToLower(mimeType)) {
	case "audio/ogg", "audio/opus":
		return "ogg"
	case "audio/mpeg", "audio/mp3":
		return "mp3"
	case "audio/mp4", "audio/x-m4a":
		return "m4a"
	case "audio/wav", "audio/x-wav":
		return "wav"
	case "audio/webm":
		return "webm"
	default:
		return "mp3"
	}
}

func truncate(s string, n int) string {
	if len(s) <= n {
		return s
	}
	return s[:n]
}

var _ AIClient = (*Groq)(nil)
