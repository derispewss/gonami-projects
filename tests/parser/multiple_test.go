package parser_test

import (
	"testing"
	"time"

	"github.com/derispewss/gonami-projects/internal/parser"
)

func TestParseMultiple(t *testing.T) {
	now := time.Date(2026, 9, 4, 10, 0, 0, 0, time.UTC)

	tests := []struct {
		name      string
		text      string
		wantCount int
		wantTotal int64
	}{
		{
			name:      "single transaction unchanged",
			text:      "beli kopi 15rb",
			wantCount: 1,
			wantTotal: 15000,
		},
		{
			name:      "dua item dipecah jadi dua transaksi",
			text:      "beli mie goreng 7k dan esteh 3k",
			wantCount: 2,
			wantTotal: 10000,
		},
		{
			name:      "tiga item dipisah",
			text:      "beli ciki 5k + aice 10rb & esteh 3k",
			wantCount: 3,
			wantTotal: 18000,
		},
		{
			name:      "kuantitas dikali",
			text:      "beli kopi 5k 2 gelas",
			wantCount: 1,
			wantTotal: 10000,
		},
		{
			name:      "kuantitas pada salah satu dari beberapa",
			text:      "beli kopi 5k 2 gelas dan teh 3k",
			wantCount: 2,
			wantTotal: 13000,
		},
		{
			name:      "bukan transaksi",
			text:      "tes",
			wantCount: 0,
			wantTotal: 0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			res, err := parser.ParseMultiple(tt.text, now)
			if tt.wantCount == 0 {
				if err == nil {
					t.Fatalf("expected error for %q, got %v", tt.text, res)
				}
				return
			}
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if len(res) != tt.wantCount {
				t.Fatalf("count = %d, want %d (%+v)", len(res), tt.wantCount, res)
			}
			var sum int64
			for _, r := range res {
				sum += r.Amount
			}
			if sum != tt.wantTotal {
				t.Errorf("total = %d, want %d", sum, tt.wantTotal)
			}
		})
	}
}

func TestParseMultipleQuantityDescription(t *testing.T) {
	now := time.Date(2026, 9, 4, 10, 0, 0, 0, time.UTC)
	res, err := parser.ParseMultiple("beli kopi 5k 2 gelas", now)
	if err != nil || len(res) != 1 {
		t.Fatalf("err=%v res=%+v", err, res)
	}
	if res[0].Description != "kopi" {
		t.Errorf("description = %q, want %q", res[0].Description, "kopi")
	}
}

func TestStemmingAndFuzzy(t *testing.T) {
	now := time.Date(2026, 9, 4, 10, 0, 0, 0, time.UTC)

	tests := []struct {
		name       string
		text       string
		wantType   string
		wantCateg  string
		wantAmount int64
	}{
		{
			name:       "kata kerja berimbuhan terstem dengan benar",
			text:       "aku membelikan ibu nasi 20k",
			wantType:   "expense",
			wantCateg:  "Food & Beverage",
			wantAmount: 20000,
		},
		{
			name:       "dibayarkan terdeteksi",
			text:       "sudah dibayarkan listrik 300k",
			wantType:   "expense",
			wantCateg:  "Bills",
			wantAmount: 300000,
		},
		{
			name:       "fuzzy typo brand indomarter terdeteksi",
			text:       "beli di indomarret 15k",
			wantType:   "expense",
			wantCateg:  "Food & Beverage",
			wantAmount: 15000,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			res, err := parser.ParseMultiple(tt.text, now)
			if err != nil || len(res) != 1 {
				t.Fatalf("err=%v res=%+v", err, res)
			}
			r := res[0]
			if string(r.Type) != tt.wantType {
				t.Errorf("type = %q, want %q", r.Type, tt.wantType)
			}
			if r.Category != tt.wantCateg {
				t.Errorf("category = %q, want %q", r.Category, tt.wantCateg)
			}
			if r.Amount != tt.wantAmount {
				t.Errorf("amount = %d, want %d", r.Amount, tt.wantAmount)
			}
		})
	}
}
