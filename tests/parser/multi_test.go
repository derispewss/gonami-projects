package parser_test

import (
	"testing"
	"time"

	"github.com/derispewss/finwa-projects/internal/parser"
)

func TestMultiItemSumming(t *testing.T) {
	now := time.Date(2026, 8, 22, 10, 0, 0, 0, time.UTC)

	tests := []struct {
		name      string
		text      string
		wantTotal int64
		wantDesc  string
	}{
		{
			name:      "dua item dengan typo sambung",
			text:      "aku beli ketoprak 25k adn esteh 3k",
			wantTotal: 28000,
			wantDesc:  "ketoprak",
		},
		{
			name:      "tiga item rantai",
			text:      "beli ciki 5k + aice 10rb & esteh 3k",
			wantTotal: 18000,
		},
		{
			name:      "nominal dengan titik ribuan",
			text:      "beli ayam 10.000 dan minum es teh 5.000",
			wantTotal: 15000,
		},
		{
			name:      "pemasukan dua item",
			text:      "gajian 2jt dan bonus 500rb",
			wantTotal: 2500000,
		},
		{
			name:      "single item tidak berubah",
			text:      "beli kopi 15rb",
			wantTotal: 15000,
		},
		{
			name:      "tanpa kata sambung tidak dijumlah",
			text:      "beli 15rb tadi pagi",
			wantTotal: 15000,
		},
		{
			name:      "no rekening tidak ikut dijumlah",
			text:      "transfer 25k ke bca 1234567890",
			wantTotal: 25000,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			res, err := parser.ParseDeterministic(tt.text, now)
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if res.Amount != tt.wantTotal {
				t.Errorf("amount = %d, want %d", res.Amount, tt.wantTotal)
			}
			if tt.wantDesc != "" && !containsFold(res.Description, tt.wantDesc) {
				t.Errorf("description %q tidak memuat %q", res.Description, tt.wantDesc)
			}
		})
	}
}

func containsFold(s, sub string) bool {
	return len(s) >= len(sub) && (s == sub || indexOf(lower(s), lower(sub)) >= 0)
}

func lower(s string) string {
	b := []byte(s)
	for i := range b {
		if b[i] >= 'A' && b[i] <= 'Z' {
			b[i] += 32
		}
	}
	return string(b)
}

func indexOf(s, sub string) int {
	for i := 0; i+len(sub) <= len(s); i++ {
		if s[i:i+len(sub)] == sub {
			return i
		}
	}
	return -1
}
