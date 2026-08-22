package parser_test

import (
	"testing"

	"github.com/derispewss/finwa-projects/internal/parser"
)

func TestMayContainTransaction(t *testing.T) {
	cases := []struct {
		text string
		want bool
	}{
		{"beli kopi 15rb", true},
		{"jajan ciki dua ribu", true},
		{"transfer ke bca 2jt", true},
		{"topup 50k ke dana", true},
		{"gajian bulan ini 5 juta", true},
		{"bayar listrik pake qris", true},
		{"halo bro gimana kabarnya", false},
		{"ok sip", false},
		{"mantap sekali bro", false},
		{"bantu donk cariin makanan", true},
	}

	for _, tc := range cases {
		if got := parser.MayContainTransaction(tc.text); got != tc.want {
			t.Errorf("MayContainTransaction(%q) = %v, want %v", tc.text, got, tc.want)
		}
	}
}
