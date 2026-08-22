package parser_test

import (
	"testing"

	"github.com/derispewss/gonami-projects/internal/parser"
)

func TestDetectIntent(t *testing.T) {
	cases := []struct {
		text string
		want parser.CommandIntent
		pd   parser.Period
	}{

		{"rekap minggu ini dong", parser.IntentRekap, parser.PeriodWeekly},
		{"rekap hari ini coy", parser.IntentRekap, parser.PeriodDaily},
		{"Rekap Bulan Ini!", parser.IntentRekap, parser.PeriodMonthly},
		{"laporan bulan ini ya", parser.IntentRekap, parser.PeriodMonthly},
		{"ringkasan kemarin donk", parser.IntentRekap, parser.PeriodDaily},
		{"berapa pengeluaran gue minggu ini", parser.IntentRekap, parser.PeriodWeekly},
		{"pengeluaran bulan ini", parser.IntentRekap, parser.PeriodMonthly},
		{"pemasukan hari ini gimana", parser.IntentRekap, parser.PeriodDaily},
		{"summary mingguan plis", parser.IntentRekap, parser.PeriodWeekly},

		{"saldo gue berapa sih", parser.IntentSaldo, parser.PeriodMonthly},
		{"sisa uang berapa", parser.IntentSaldo, parser.PeriodMonthly},
		{"cashflow", parser.IntentSaldo, parser.PeriodMonthly},
		{"cek saldo dulu deh", parser.IntentSaldo, parser.PeriodMonthly},
		{"uangku sisa apa", parser.IntentSaldo, parser.PeriodMonthly},

		{"riwayat", parser.IntentRiwayat, parser.PeriodMonthly},
		{"transaksi terakhir dong", parser.IntentRiwayat, parser.PeriodMonthly},
		{"riwayat transaksi gue", parser.IntentRiwayat, parser.PeriodMonthly},

		{"hapus transaksi terakhir", parser.IntentHapus, parser.PeriodMonthly},
		{"apus yang terakhir ya", parser.IntentHapus, parser.PeriodMonthly},
		{"delete transaksi tadi", parser.IntentHapus, parser.PeriodDaily},

		{"help", parser.IntentHelp, parser.PeriodMonthly},
		{"bantuan dong min", parser.IntentHelp, parser.PeriodMonthly},

		{"kopi 25k", parser.IntentNone, parser.PeriodMonthly},
		{"makan siang 20rb tadi", parser.IntentNone, parser.PeriodMonthly},
		{"beli bakso enak", parser.IntentNone, parser.PeriodMonthly},
		{"halo bro", parser.IntentNone, parser.PeriodMonthly},
	}

	for _, c := range cases {
		gotKind, gotPeriod := parser.DetectIntent(c.text)
		if gotKind != c.want || gotPeriod != c.pd {
			t.Errorf("DetectIntent(%q) = (%v, %v), want (%v, %v)",
				c.text, gotKind, gotPeriod, c.want, c.pd)
		}
	}
}

func TestMatchConfirmation(t *testing.T) {
	cases := []struct {
		text    string
		yes     bool
		decided bool
	}{

		{"iya", true, true},
		{"ya", true, true},
		{"iya dong", true, true},
		{"gaskeun", true, true},
		{"ok simpan", true, true},
		{"sip banget", true, true},
		{"betul coy", true, true},
		{"Oke Simpan Aja", true, true},

		{"tidak", false, true},
		{"ga usah", false, true},
		{"jangan dulu", false, true},
		{"batalin deh", false, true},
		{"nggak", false, true},
		{"ulang", false, true},

		{"kopi 15k", false, false},
		{"iya tapi jangan sekarang", false, false},
		{"", false, false},
		{"halo", false, false},
		{"rekap minggu ini", false, false},
	}

	for _, c := range cases {
		yes, decided := parser.MatchConfirmation(c.text)
		if yes != c.yes || decided != c.decided {
			t.Errorf("MatchConfirmation(%q) = (%v, %v), want (%v, %v)",
				c.text, yes, decided, c.yes, c.decided)
		}
	}
}
