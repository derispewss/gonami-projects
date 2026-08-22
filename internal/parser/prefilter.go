package parser

import (
	"regexp"
	"strings"
)

var reAmountSignal = regexp.MustCompile(
	`(\d|rp\s*\.?\s*\d|\b\d+\s*k\b|juta|ribu|ratus|puluh belas|koma)`)

var financeVerbs = map[string]bool{

	"beli": true, "bayar": true, "belanja": true, "makan": true, "minum": true,
	"jajan": true, "kopi": true, "bensin": true, "isi": true, "topup": true, "top": true,
	"parkir": true, "tol": true, "ongkir": true, "sewa": true, "kos": true,
	"kontrakan": true, "listrik": true, "pulsa": true, "token": true,
	"langganan": true, "obat": true, "tiket": true,

	"gaji": true, "gajian": true, "transfer": true, "kirim": true, "jual": true,
	"terima": true, "dapat": true, "bonus": true, "thr": true, "cashback": true,
	"refund": true, "setor": true, "narik": true, "upah": true, "untung": true,
}

func MayContainTransaction(s string) bool {
	s = strings.ToLower(strings.TrimSpace(s))
	s = normalizeGluedWords(s)
	if len([]rune(s)) < 4 {
		return false
	}
	if reAmountSignal.MatchString(s) {
		return true
	}
	for _, w := range strings.Fields(s) {
		w = strings.Trim(w, ",.!?:;()\"'")
		if w == "" {
			continue
		}
		if financeVerbs[w] {
			return true
		}

		if len(w) > 5 {
			for v := range financeVerbs {
				if len(v) >= 4 && strings.HasPrefix(w, v) {
					return true
				}
			}
		}
	}
	return false
}
