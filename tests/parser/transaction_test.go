package parser

import (
	"testing"
	"time"

	"github.com/derispewss/gonami-projects/internal/domain"
	"github.com/derispewss/gonami-projects/internal/parser"
)

func TestParseTransaction(t *testing.T) {
	loc, _ := time.LoadLocation("Asia/Jakarta")
	now := time.Date(2026, time.August, 21, 15, 30, 0, 0, loc)

	tests := []struct {
		name       string
		input      string
		expectType domain.TransactionType
		expectAmt  int64
		expectCat  string
		expectConf float64
	}{

		{
			name:       "beli coca cola 10k",
			input:      "beli coca cola 10k",
			expectType: domain.TypeExpense,
			expectAmt:  10000,
			expectCat:  "Food & Beverage",
			expectConf: 0.80,
		},
		{
			name:       "makan 25 ribu",
			input:      "makan 25 ribu",
			expectType: domain.TypeExpense,
			expectAmt:  25000,
			expectCat:  "Food & Beverage",
			expectConf: 0.80,
		},
		{
			name:       "bayar kos 1.2jt",
			input:      "bayar kos 1.2jt",
			expectType: domain.TypeExpense,
			expectAmt:  1200000,
			expectCat:  "Bills",
			expectConf: 0.80,
		},
		{
			name:       "kemarin beli kopi 15k",
			input:      "kemarin beli kopi 15k",
			expectType: domain.TypeExpense,
			expectAmt:  15000,
			expectCat:  "Food & Beverage",
			expectConf: 0.80,
		},

		{
			name:       "gaji 5 juta",
			input:      "gaji 5 juta",
			expectType: domain.TypeIncome,
			expectAmt:  5000000,
			expectCat:  "Salary",
			expectConf: 0.80,
		},
		{
			name:       "gaji masuk 6jt",
			input:      "gaji masuk 6jt",
			expectType: domain.TypeIncome,
			expectAmt:  6000000,
			expectCat:  "Salary",
			expectConf: 0.80,
		},

		{
			name:       "transfer 500k ke budi",
			input:      "transfer 500k ke budi",
			expectType: domain.TypeTransfer,
			expectAmt:  500000,
			expectCat:  "",
			expectConf: 0.80,
		},
		{
			name:       "tadi transfer ke budi 100k",
			input:      "tadi transfer ke budi 100k",
			expectType: domain.TypeTransfer,
			expectAmt:  100000,
			expectCat:  "",
			expectConf: 0.80,
		},

		{
			name:       "kayaknya bayar sesuatu 20k",
			input:      "kayaknya bayar sesuatu 20k",
			expectType: domain.TypeExpense,
			expectAmt:  20000,
			expectCat:  "",
			expectConf: 0.50,
		},
	}

	p := parser.New()

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			res, err := p.Parse(nil, tc.input, now)
			if err != nil {
				t.Fatalf("expected success, got error: %v", err)
			}
			if res.Type != tc.expectType {
				t.Errorf("expected type %s, got %s", tc.expectType, res.Type)
			}
			if res.Amount != tc.expectAmt {
				t.Errorf("expected amount %d, got %d", tc.expectAmt, res.Amount)
			}
			if res.Category != tc.expectCat {
				t.Errorf("expected category %s, got %s", tc.expectCat, res.Category)
			}
			if res.Confidence < tc.expectConf {
				t.Errorf("expected confidence >= %f, got %f", tc.expectConf, res.Confidence)
			}
		})
	}
}

func TestMalformedTransaction(t *testing.T) {
	loc, _ := time.LoadLocation("Asia/Jakarta")
	now := time.Now().In(loc)

	p := parser.New()

	tests := []string{
		"",
		"halo",
		"bayar sesuatu",
		"sesuatu tadi",
	}

	for _, input := range tests {
		t.Run(input, func(t *testing.T) {
			_, err := p.Parse(nil, input, now)
			if err == nil {
				t.Fatalf("expected error for malformed input: %q", input)
			}
		})
	}
}
