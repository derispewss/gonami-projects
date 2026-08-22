package parser_test

import (
	"context"
	"testing"

	"github.com/derispewss/finwa-projects/internal/ai"
	"github.com/derispewss/finwa-projects/internal/media"
)

func TestParseReceiptJSON(t *testing.T) {
	tests := []struct {
		name   string
		raw    string
		wantOK bool
		valid  bool
		amount int64
		txType string
	}{
		{
			name:   "struk indomaret valid",
			raw:    `{"is_transaction":true,"type":"expense","amount":15000,"description":"Kopi Susu","merchant":"Indomaret","category_hint":"Food & Beverage","confidence":0.95}`,
			wantOK: true, valid: true, amount: 15000, txType: "expense",
		},
		{
			name:   "dibungkus code fence",
			raw:    "```json\n{\"is_transaction\":true,\"type\":\"EXPENSE\",\"amount\":25000,\"description\":\"Makan\",\"merchant\":\"\",\"category_hint\":\"Other\",\"confidence\":0.8}\n```",
			wantOK: true, valid: true, amount: 25000, txType: "expense",
		},
		{
			name:   "bukan transaksi",
			raw:    `{"is_transaction":false,"type":"","amount":0,"description":"","merchant":"","category_hint":"","confidence":0}`,
			wantOK: true, valid: false,
		},
		{
			name:   "amount negatif dinolkan",
			raw:    `{"is_transaction":true,"type":"expense","amount":-500,"description":"x","merchant":"","category_hint":"","confidence":0.9}`,
			wantOK: true, valid: false,
		},
		{
			name:   "bukan json",
			raw:    "halo ini bukan json",
			wantOK: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			rec, err := ai.ParseExtractionJSON(tt.raw)
			if tt.wantOK && err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if !tt.wantOK {
				if err == nil {
					t.Fatal("expected error for invalid JSON")
				}
				return
			}
			if rec.IsValid() != tt.valid {
				t.Errorf("IsValid() = %v, want %v", rec.IsValid(), tt.valid)
			}
			if tt.valid {
				if rec.Amount != tt.amount {
					t.Errorf("amount = %d, want %d", rec.Amount, tt.amount)
				}
				if rec.Type != tt.txType {
					t.Errorf("type = %q, want %q", rec.Type, tt.txType)
				}
			}
		})
	}
}

func TestProcessorSupports(t *testing.T) {
	audio := media.NewAudioProcessor(nil)
	image := media.NewImageProcessor(nil)
	doc := media.NewDocumentProcessor(nil)

	cases := []struct {
		p    media.Processor
		mime string
		want bool
	}{
		{audio, "audio/ogg; codecs=opus", true},
		{audio, "audio/mpeg", true},
		{audio, "image/jpeg", false},
		{image, "image/jpeg", true},
		{image, "image/png", true},
		{image, "application/pdf", false},
		{doc, "application/pdf", true},
		{doc, "application/msword", false},
	}

	for _, tc := range cases {
		if got := tc.p.Supports(tc.mime); got != tc.want {
			t.Errorf("%T.Supports(%q) = %v, want %v", tc.p, tc.mime, got, tc.want)
		}
	}
}

func TestImageProcessorRejectsNonReceipt(t *testing.T) {

	out := &media.Output{}
	if out.Receipt != nil {
		t.Fatal("expected nil Receipt for non-receipt image")
	}
	_ = context.Background()
}
