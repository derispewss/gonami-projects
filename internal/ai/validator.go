package ai

import (
	"encoding/json"
	"fmt"
	"strings"
)

var validTypes = map[string]bool{
	"expense": true, "income": true, "transfer": true,
}

func ParseExtractionJSON(raw string) (*Extraction, error) {
	raw = strings.TrimSpace(raw)
	raw = strings.TrimPrefix(raw, "```json")
	raw = strings.TrimPrefix(raw, "```")
	raw = strings.TrimSuffix(raw, "```")
	raw = strings.TrimSpace(raw)

	var rec Extraction
	if err := json.Unmarshal([]byte(raw), &rec); err != nil {
		return nil, fmt.Errorf("bukan JSON valid: %w", err)
	}

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
	return &rec, nil
}
