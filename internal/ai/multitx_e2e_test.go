package ai

import (
	"context"
	"os"
	"strings"
	"testing"
	"time"
)

// E2E tests hit the real Gemini API. Mereka skip tanpa GEMINI_API_KEY/GEMINI_MODEL
// dan/atau tanpa testdata. Jalankan manual dengan: go test -run 'TestExtract.*MultiTxE2E' ./internal/ai/

func TestExtractReceiptsMultiTxE2E(t *testing.T) {
	key := os.Getenv("GEMINI_API_KEY")
	model := os.Getenv("GEMINI_MODEL")
	if key == "" || model == "" {
		t.Skip("tanpa GEMINI_API_KEY/GEMINI_MODEL")
	}
	raw, err := os.ReadFile("testdata/multi_receipt.png")
	if err != nil {
		t.Skip("testdata/multi_receipt.png tidak ada")
	}
	g, err := NewGemini(key, model, "", nil, 0)
	if err != nil {
		t.Fatal(err)
	}
	list, err := g.ExtractReceipts(context.Background(), raw, "image/png")
	if err != nil {
		if strings.Contains(err.Error(), "429") || strings.Contains(err.Error(), "quota") {
			t.Skipf("rate limit: %v", err)
		}
		t.Fatal(err)
	}
	t.Logf("ditemukan %d trx", len(list))
	for _, e := range list {
		t.Logf("tx=%v type=%s amt=%d desc=%q conf=%.2f", e.IsTransaction, e.Type, e.Amount, e.Description, e.Confidence)
	}
}

func TestExtractFromStatementTextsMultiTxE2E(t *testing.T) {
	key := os.Getenv("GEMINI_API_KEY")
	model := os.Getenv("GEMINI_MODEL")
	if key == "" || model == "" {
		t.Skip("tanpa GEMINI_API_KEY/GEMINI_MODEL")
	}
	stmt := `Tanggal        Keterangan            Masuk       Keluar     Saldo
01/09/2026   Transfer Masuk Gaji   5000000                5000123
01/09/2026   Belanja Indomaret                 45000     4955123
02/09/2026   Gojek Transport                  25000     4930123
02/09/2026   Token Listrik 200rb              200000     4730123
03/09/2026   Shopee Belanja Online            120000     4610123`
	g, err := NewGemini(key, model, "", nil, 0)
	if err != nil {
		t.Fatal(err)
	}
	list, err := g.ExtractFromStatementTexts(context.Background(), stmt, time.Now())
	if err != nil {
		if strings.Contains(err.Error(), "429") || strings.Contains(err.Error(), "quota") {
			t.Skipf("rate limit: %v", err)
		}
		t.Fatal(err)
	}
	if len(list) == 0 {
		t.Skip("tidak ada transaksi diekstrak (model mengembalikan kosong)")
	}
	for _, e := range list {
		t.Logf("tx=%v type=%s amt=%d desc=%q", e.IsTransaction, e.Type, e.Amount, e.Description)
	}
}
