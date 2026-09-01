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

func TestDetectIntentAdvanced(t *testing.T) {
	cases := []struct {
		text string
		want parser.CommandIntent
		pd   parser.Period
	}{
		{"budget", parser.IntentBudget, parser.PeriodMonthly},
		{"cek budget dong", parser.IntentBudget, parser.PeriodMonthly},
		{"anggaran bulan ini", parser.IntentBudget, parser.PeriodMonthly},

		{"insight", parser.IntentInsight, parser.PeriodMonthly},
		{"analisa pengeluaran gue", parser.IntentInsight, parser.PeriodMonthly},
		{"ada anomali nggak", parser.IntentInsight, parser.PeriodMonthly},
		{"pengeluaran berulang", parser.IntentInsight, parser.PeriodMonthly},

		{"export", parser.IntentExport, parser.PeriodMonthly},
		{"ekspor minggu ini", parser.IntentExport, parser.PeriodWeekly},
		{"unduh laporan hari ini", parser.IntentExport, parser.PeriodDaily},

		{"dompet", parser.IntentWallet, parser.PeriodMonthly},
		{"pakai dompet bca", parser.IntentWallet, parser.PeriodMonthly},
		{"buat dompet ovo", parser.IntentWallet, parser.PeriodMonthly},

		{"kategori", parser.IntentKategori, parser.PeriodMonthly},
		{"tambah kategori skincare", parser.IntentKategori, parser.PeriodMonthly},

		{"beli kopi 15k", parser.IntentNone, parser.PeriodMonthly},
	}

	for _, c := range cases {
		got, pd := parser.DetectIntent(c.text)
		if got != c.want || pd != c.pd {
			t.Errorf("DetectIntent(%q) = (%v, %v), want (%v, %v)",
				c.text, got, pd, c.want, c.pd)
		}
	}
}

func TestDetectBudgetCommand(t *testing.T) {
	cases := []struct {
		text     string
		ok       bool
		category string
		amount   int64
		del      bool
	}{
		{"budget makan 500rb", true, "makan", 500000, false},
		{"budget transportasi 250k", true, "transportasi", 250000, false},
		{"atur budget jajan bulanan 1000000", true, "jajan", 1000000, false},
		{"budget kopi Rp50.000", true, "kopi", 50000, false},
		{"hapus budget makan", true, "makan", 0, true},
		{"budget", false, "", 0, false},
		{"rekap bulan ini", false, "", 0, false},
		{"beli kopi 15k", false, "", 0, false},
	}

	for _, c := range cases {
		cmd, ok := parser.DetectBudgetCommand(c.text)
		if ok != c.ok {
			t.Errorf("DetectBudgetCommand(%q) ok = %v, want %v", c.text, ok, c.ok)
			continue
		}
		if !ok {
			continue
		}
		if cmd.Delete != c.del {
			t.Errorf("DetectBudgetCommand(%q).Delete = %v, want %v", c.text, cmd.Delete, c.del)
		}
		if !c.del && (cmd.Category != c.category || cmd.Amount != c.amount) {
			t.Errorf("DetectBudgetCommand(%q) = (%q, %d), want (%q, %d)",
				c.text, cmd.Category, cmd.Amount, c.category, c.amount)
		}
	}
}

func TestDetectIntentReset(t *testing.T) {
	got, _ := parser.DetectIntent("hapus semua data")
	if got != parser.IntentReset {
		t.Errorf("DetectIntent(hapus semua data) = %v, want IntentReset", got)
	}
	got, _ = parser.DetectIntent("hapus transaksi terakhir")
	if got != parser.IntentHapus {
		t.Errorf("DetectIntent(hapus transaksi terakhir) = %v, want IntentHapus", got)
	}
	got, _ = parser.DetectIntent("reset data gue")
	if got != parser.IntentReset {
		t.Errorf("DetectIntent(reset data gue) = %v, want IntentReset", got)
	}
}

func TestDetectBudgetCommandAdjust(t *testing.T) {
	cmd, ok := parser.DetectBudgetCommand("ubah budget makan 700rb")
	if !ok || cmd.Category != "makan" || cmd.Amount != 700000 || cmd.Delete {
		t.Errorf("DetectBudgetCommand(ubah budget makan 700rb) = %+v ok=%v, want cat=makan amt=700000", cmd, ok)
	}
	cmd, ok = parser.DetectBudgetCommand("revisi budget transportasi jadi 300k")
	if !ok || cmd.Category != "transportasi" || cmd.Amount != 300000 {
		t.Errorf("DetectBudgetCommand(revisi budget...) = %+v ok=%v", cmd, ok)
	}
	cmd, ok = parser.DetectBudgetCommand("budget makan 500rb")
	if !ok || cmd.Category != "makan" || cmd.Amount != 500000 {
		t.Errorf("DetectBudgetCommand(budget makan 500rb) = %+v ok=%v", cmd, ok)
	}
}
