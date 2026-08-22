package parser

import (
	"testing"
	"time"

	"github.com/derispewss/finwa-projects/internal/parser"
)

func TestDetectDate(t *testing.T) {
	loc, _ := time.LoadLocation("Asia/Jakarta")
	now := time.Date(2026, time.August, 21, 15, 30, 0, 0, loc)
	today := time.Date(2026, time.August, 21, 0, 0, 0, 0, loc)
	yesterday := time.Date(2026, time.August, 20, 0, 0, 0, 0, loc)

	tests := []struct {
		input  string
		expect time.Time
	}{
		{"hari ini", today},
		{"kemarin", yesterday},
		{"barusan", today},
		{"tadi", today},
		{"semalam", yesterday},
		{"21 agustus", time.Date(2026, time.August, 21, 0, 0, 0, 0, loc)},
		{"beli kopi 15k", today},
	}

	for _, tc := range tests {
		t.Run(tc.input, func(t *testing.T) {
			match := parser.DetectDate(tc.input, now)
			if !match.Date.Equal(tc.expect) {
				t.Fatalf("expected date %v, got %v for input %q", tc.expect, match.Date, tc.input)
			}
		})
	}
}
