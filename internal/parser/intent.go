package parser

import (
	"strings"
)

type CommandIntent uint8

const (
	IntentNone CommandIntent = iota
	IntentRekap
	IntentSaldo
	IntentRiwayat
	IntentHapus
	IntentHelp
	IntentBudget
	IntentInsight
	IntentExport
	IntentWallet
	IntentKategori
	IntentReset
)

type Period uint8

const (
	PeriodMonthly Period = iota
	PeriodWeekly
	PeriodDaily
)

var fillers = map[string]bool{
	"dong": true, "dongs": true, "donk": true, "dunk": true, "dung": true,
	"coy": true, "cuy": true, "bro": true, "sis": true, "bre": true,
	"bang": true, "gan": true, "min": true, "om": true, "kak": true, "dek": true,
	"deh": true, "aja": true, "saja": true, "nih": true, "tuh": true, "sih": true,
	"kah": true, "yah": true, "pls": true, "plis": true, "please": true,
	"banget": true, "bgt": true, "kok": true, "woy": true, "woi": true,
	"dulu": true, "keun": true,
}

var confirmWords = map[string]bool{
	"iya": true, "iyah": true, "ya": true, "y": true, "yes": true, "yup": true,
	"yups": true, "ok": true, "oke": true, "oksy": true, "sip": true,
	"sipp": true, "gas": true, "gaskeun": true, "gaskan": true, "gaspol": true,
	"gass": true, "simpan": true, "save": true, "betul": true, "bener": true,
	"benar": true, "setuju": true, "lanjut": true, "mantap": true, "aman": true,
	"boleh": true, "go": true,
}

var rejectWords = map[string]bool{
	"tidak": true, "nggak": true, "ngga": true, "enggak": true, "engga": true,
	"gak": true, "ga": true, "gk": true, "gamau": true, "no": true, "n": true,
	"nope": true, "cancel": true, "batal": true, "batalin": true, "jangan": true,
	"stop": true, "skip": true, "ulang": true,
}

func NormalizeCommand(s string) []string {
	fields := strings.FieldsFunc(strings.ToLower(s), func(r rune) bool {
		return !('a' <= r && r <= 'z') && !('0' <= r && r <= '9')
	})
	tokens := make([]string, 0, len(fields))
	for _, w := range fields {
		if w == "" || fillers[w] {
			continue
		}
		tokens = append(tokens, w)
	}
	return tokens
}

func hasAny(tokens []string, words ...string) bool {
	return HasAnyToken(tokens, words...)
}

func HasAnyToken(tokens []string, words ...string) bool {
	for _, t := range tokens {
		for _, w := range words {
			if t == w {
				return true
			}
		}
	}
	return false
}

func hasBigram(tokens []string, bigram string) bool {
	parts := strings.Fields(bigram)
	if len(parts) != 2 {
		return false
	}
	for i := 0; i+1 < len(tokens); i++ {
		if tokens[i] == parts[0] && tokens[i+1] == parts[1] {
			return true
		}
	}
	return false
}

func hasPrefix(tokens []string, prefix string) bool {
	for _, t := range tokens {
		if strings.HasPrefix(t, prefix) {
			return true
		}
	}
	return false
}

func detectPeriod(tokens []string) Period {
	switch {
	case hasBigram(tokens, "hari ini"), hasAny(tokens, "kemarin", "tadi"):
		return PeriodDaily
	case hasPrefix(tokens, "minggu"):
		return PeriodWeekly
	default:
		return PeriodMonthly
	}
}

func DetectIntent(text string) (CommandIntent, Period) {
	lower := strings.ToLower(text)
	if reAmountSignal.MatchString(lower) {
		return IntentNone, PeriodMonthly
	}

	tokens := NormalizeCommand(lower)
	if len(tokens) == 0 {
		return IntentNone, PeriodMonthly
	}
	period := detectPeriod(tokens)

	switch {

	case (hasAny(tokens, "hapus", "apus", "delete") && hasAny(tokens, "semua", "data")) ||
		hasAny(tokens, "reset") || hasBigram(tokens, "bersihkan semua") ||
		hasBigram(tokens, "hapus semua"):
		return IntentReset, period

	case hasAny(tokens, "hapus", "apus", "delete") &&
		(hasAny(tokens, "terakhir") || hasAny(tokens, "transaksi")):
		return IntentHapus, period

	case hasAny(tokens, "budget", "anggaran"):
		return IntentBudget, period

	case hasAny(tokens, "insight", "insights", "analisa", "anomali", "deteksi"),
		hasBigram(tokens, "pengeluaran berulang"), hasBigram(tokens, "langganan bulanan"),
		hasBigram(tokens, "tips hemat"):
		return IntentInsight, period

	case hasAny(tokens, "export", "ekspor", "unduh"), hasBigram(tokens, "kirim laporan"),
		hasBigram(tokens, "download laporan"):
		return IntentExport, period

	case hasAny(tokens, "dompet", "wallet", "rekening"):
		return IntentWallet, period

	case hasAny(tokens, "kategori"):
		return IntentKategori, period

	case hasAny(tokens, "riwayat"), hasBigram(tokens, "transaksi terakhir"),
		hasBigram(tokens, "catatan transaksi"):
		return IntentRiwayat, period

	case hasAny(tokens, "saldo", "cashflow", "cashflowku", "uangku", "uanggue"),
		hasBigram(tokens, "sisa uang"), hasBigram(tokens, "sisa dana"),
		hasBigram(tokens, "uang saya"), hasBigram(tokens, "uang gue"),
		hasBigram(tokens, "uang ku"):
		return IntentSaldo, period

	case hasAny(tokens, "rekap", "ringkasan", "laporan", "summary", "recap"),
		hasBigram(tokens, "berapa pengeluaran"), hasBigram(tokens, "berapa pemasukan"),
		hasAny(tokens, "pengeluaran", "pemasukan"):
		return IntentRekap, period

	case hasAny(tokens, "help", "bantuan"):
		return IntentHelp, period
	}

	return IntentNone, period
}

func MatchConfirmation(s string) (yes bool, decided bool) {
	lower := strings.ToLower(strings.TrimSpace(s))
	if lower == "" {
		return false, false
	}

	if reAmountSignal.MatchString(lower) {
		return false, false
	}

	tokens := NormalizeCommand(lower)
	var hitConfirm, hitReject bool
	for _, t := range tokens {
		switch {
		case confirmWords[t]:
			hitConfirm = true
		case rejectWords[t]:
			hitReject = true
		}
	}

	if hitConfirm == hitReject {
		return false, false
	}
	return hitConfirm, true
}

type BudgetCommand struct {
	Category string
	Amount   int64
	Delete   bool
}

func DetectBudgetCommand(text string) (*BudgetCommand, bool) {
	lower := strings.ToLower(text)

	tokens := NormalizeCommand(lower)
	hasKeyword := hasAny(tokens, "budget", "anggaran")
	if !hasKeyword {
		return nil, false
	}

	if hasAny(tokens, "hapus", "apus", "delete", "buang") {
		words := removeBudgetKeywords(tokens)
		words = removeCommandVerbs(words)
		if len(words) == 0 {
			return nil, true
		}
		return &BudgetCommand{Category: strings.Join(words, " "), Delete: true}, true
	}

	m, found := FindAmount(lower)
	if !found || m.Value <= 0 {
		return nil, false
	}

	before := strings.TrimSpace(lower[:m.Start])
	words := NormalizeCommand(before)
	words = removeBudgetKeywords(words)
	words = removeCommandVerbs(words)
	if len(words) == 0 || len(words) > 3 {
		return nil, false
	}
	return &BudgetCommand{Category: strings.Join(words, " "), Amount: m.Value}, true
}

func removeBudgetKeywords(tokens []string) []string {
	out := make([]string, 0, len(tokens))
	for _, t := range tokens {
		switch t {
		case "budget", "anggaran":
		default:
			out = append(out, t)
		}
	}
	return out
}

func removeCommandVerbs(tokens []string) []string {
	out := make([]string, 0, len(tokens))
	for _, t := range tokens {
		switch t {
		case "set", "atur", "pasang", "ubah", "buat", "tambah", "bikin",
			"untuk", "per", "bulan", "bulanan", "kategori", "maksimal",
			"max", "limit", "hapus", "apus", "delete", "buang",
			"revisi", "jadi", "menjadi":
		default:
			out = append(out, t)
		}
	}
	return out
}
