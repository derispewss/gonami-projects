package parser

import (
	"regexp"
	"strings"
)

var passiveIncomeWords = []string{
	"dikirimin", "dikirim", "ditransferin", "ditransfer ke saya",
	"di_transfer", "diterima", "diberi", "dapat dari", "dapet dari",
	"masuk dari", "terima dari",
}

var transferVerbs = []string{"transfer", "kirim", "pindahin", "pindah", "topup ke"}

var incomeWords = []string{
	"gaji", "dapat", "dapet", "terima", "masuk", "pemasukan", "income",
	"dibayar", "bonus", "thr", "untung", "profit", "penjualan", "jual",
	"fee", "honor", "komisi", "cashback", "refund", " reimbourse",
}

var expenseWords = []string{
	"beli", "belanja", "bayar", "bayarin", "keluar", "pengeluaran",
	"makan", "minum", "jajan", "top up", "topup", "isi", "bensin",
	"parkir", "cicil", "cicilan", "sewa", "kos", "kontrakan",
	"beliin", "traktir", "nujun", "perbaiki", "servis", "cuci",
}

type typeMatch struct {
	Type     string
	Explicit bool
}

func DetectType(lower string) typeMatch {
	// Stem finance verbs so affixed forms ("dibelikan", "membelikan") match.
	stemmed := stemSentence(lower)

	if stemsAny(stemmed, passiveIncomeWords) {
		return typeMatch{Type: "income", Explicit: true}
	}

	if hasTransferWithDestination(lower) {
		return typeMatch{Type: "transfer", Explicit: true}
	}

	for _, w := range incomeWords {
		if containsWord(stemmed, w) || containsWord(lower, w) {
			return typeMatch{Type: "income", Explicit: true}
		}
	}

	for _, w := range expenseWords {
		if containsWord(stemmed, w) || containsWord(lower, w) {
			return typeMatch{Type: "expense", Explicit: true}
		}
	}

	return typeMatch{Type: "expense", Explicit: false}
}

// stemsAny reports whether any keyword appears in the stemmed text.
func stemsAny(stemmed string, keywords []string) bool {
	for _, w := range keywords {
		if containsWord(stemmed, w) {
			return true
		}
	}
	return false
}

func hasTransferWithDestination(lower string) bool {
	hasVerb := false
	for _, v := range transferVerbs {
		if containsWord(lower, v) {
			hasVerb = true
			break
		}
	}
	if !hasVerb {
		return false
	}
	_, ok := FindTransferDestination(lower)
	return ok
}

var reDestination = regexp.MustCompile(`\bke\s+([a-z][a-z0-9'.]*(?:\s+[a-z][a-z0-9'.]*){0,2})`)

var selfReferences = []string{"saya", "aku", "gw", "gue", "aku sendiri", "rekening saya"}

func FindTransferDestination(lower string) (string, bool) {
	loc := reDestination.FindStringSubmatchIndex(lower)
	if loc == nil {
		return "", false
	}
	dest := lower[loc[2]:loc[3]]
	dest = trimStopSuffix(dest)

	if dest == "" {
		return "", false
	}

	for _, s := range selfReferences {
		if dest == s || strings.HasPrefix(dest, s+" ") {
			return "", false
		}
	}
	return dest, true
}

func trimStopSuffix(s string) string {
	fields := strings.Fields(s)
	stop := map[string]bool{
		"dan": true, "atau": true, "untuk": true, "buat": true,
		"dari": true, "pada": true, "yang": true, "tadi": true,
		"kemarin": true, "barusan": true, "hari": true, "ini": true,
	}
	for len(fields) > 1 && stop[fields[len(fields)-1]] {
		fields = fields[:len(fields)-1]
	}
	return strings.Join(fields, " ")
}

func containsWord(s, word string) bool {
	return indexOfPhrase(s, word) >= 0
}

var hedges = []string{
	"kayaknya", "sepertinya", "mungkin", "kira-kira", "kurang lebih",
	"sekitar", "barangkali", "agaknya", "gak tau", "nggak tau", "pokoknya",
}

func CountHedges(lower string) int {
	n := 0
	for _, h := range hedges {
		if containsWord(lower, h) {
			n++
		}
	}
	return n
}
