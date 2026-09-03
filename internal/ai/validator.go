package ai

import (
	"encoding/json"
	"fmt"
	"strings"
)

var validTypes = map[string]bool{
	"expense": true, "income": true, "transfer": true,
}

func cleanRaw(raw string) string {
	raw = strings.TrimSpace(raw)
	raw = strings.TrimPrefix(raw, "```json")
	raw = strings.TrimPrefix(raw, "```")
	raw = strings.TrimSuffix(raw, "```")
	return strings.TrimSpace(raw)
}

func extractJSONObject(raw string) string {
	start := strings.IndexByte(raw, '{')
	end := strings.LastIndexByte(raw, '}')
	if start >= 0 && end > start {
		return raw[start : end+1]
	}
	return raw
}

func normalizeExtraction(rec *Extraction) {
	rec.Type = strings.ToLower(strings.TrimSpace(rec.Type))
	if rec.Type != "" && !validTypes[rec.Type] {
		rec.Type = ""
	}
	if rec.Confidence < 0 {
		rec.Confidence = 0
	}
	if rec.Confidence > 1 {
		rec.Confidence = 1
	}
	rec.Description = strings.TrimSpace(rec.Description)
	if len(rec.Description) > 60 {
		rec.Description = rec.Description[:60]
	}
	rec.Merchant = strings.TrimSpace(rec.Merchant)
	rec.CategoryHint = strings.TrimSpace(rec.CategoryHint)
	rec.DateHint = strings.TrimSpace(rec.DateHint)
	if rec.Amount < 0 {
		rec.Amount = 0
	}
}

func ParseExtractionJSON(raw string) (*Extraction, error) {
	raw = cleanRaw(raw)
	raw = extractJSONObject(raw)

	var rec Extraction
	if err := json.Unmarshal([]byte(raw), &rec); err != nil {
		return nil, fmt.Errorf("bukan JSON valid: %w", err)
	}
	normalizeExtraction(&rec)
	return &rec, nil
}

func ParseExtractionJSONArray(raw string) ([]*Extraction, error) {
	raw = cleanRaw(raw)

	start := strings.IndexByte(raw, '[')
	end := strings.LastIndexByte(raw, ']')
	if start >= 0 && end > start {
		raw = raw[start : end+1]
	} else {
		single, err := ParseExtractionJSON(raw)
		if err != nil {
			return nil, fmt.Errorf("bukan JSON array valid: %w", err)
		}
		return []*Extraction{single}, nil
	}

	var list []Extraction
	if err := json.Unmarshal([]byte(raw), &list); err != nil {
		return nil, fmt.Errorf("bukan JSON array valid: %w", err)
	}
	out := make([]*Extraction, 0, len(list))
	for i := range list {
		normalizeExtraction(&list[i])
		out = append(out, &list[i])
	}
	return out, nil
}
