package parser_test

import (
	"testing"

	"github.com/derispewss/gonami-projects/internal/parser"
)

func TestFindAmount(t *testing.T) {
	tests := []struct {
		input  string
		expect int64
		ok     bool
	}{

		{"1k", 1000, true},
		{"5k", 5000, true},
		{"10k", 10000, true},
		{"10 ribu", 10000, true},
		{"10ribuan", 10000, true},
		{"10rb", 10000, true},
		{"50 ribu", 50000, true},
		{"Rp25.000 untuk makan", 25000, true},
		{"barusan beli rokok 20k", 20000, true},
		{"tadi makan sama temen 50 ribu", 50000, true},

		{"1 juta", 1000000, true},
		{"1jt", 1000000, true},
		{"1,5 juta", 1500000, true},
		{"1.5 juta", 1500000, true},
		{"2.5jt", 2500000, true},
		{"1.2jt", 1200000, true},
		{"gaji masuk 6jt", 6000000, true},

		{"Rp10.000", 10000, true},
		{"Rp 10.000", 10000, true},
		{"Rp1.500.000", 1500000, true},

		{"126500", 126500, true},
		{"bayar 2000", 2000, true},

		{"makan siang dua puluh lima ribu", 25000, true},
		{"tadi keluar lima belas ribu", 15000, true},
		{"beli sepatu tiga ratus lima puluh ribu", 350000, true},
		{"gaji masuk tiga koma lima juta", 3500000, true},

		{"halo", 0, false},
		{"bayar sesuatu", 0, false},
		{"", 0, false},
	}

	for _, tc := range tests {
		t.Run(tc.input, func(t *testing.T) {
			m, ok := parser.FindAmount(tc.input)
			if ok != tc.ok {
				t.Fatalf("expected ok=%v, got %v for input %q", tc.ok, ok, tc.input)
			}
			if ok && m.Value != tc.expect {
				t.Fatalf("expected %d, got %d for input %q", tc.expect, m.Value, tc.input)
			}
		})
	}
}
