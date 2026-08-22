package parser

import (
	"errors"
	"strings"
	"time"

	"github.com/derispewss/gonami-projects/internal/domain"
)

var ErrNotTransaction = errors.New("teks bukan transaksi")

func ParseDeterministic(text string, now time.Time) (*Result, error) {
	raw := strings.TrimSpace(text)
	if len([]rune(raw)) < 3 {
		return nil, ErrNotTransaction
	}

	lower := strings.ToLower(raw)

	lower = convertSpelledNumbers(lower)

	matches := findAllAmounts(lower)
	if len(matches) == 0 {
		return nil, ErrNotTransaction
	}

	var total int64
	var summedSpans []amountMatch
	if t, chain := sumItemChain(lower, matches); len(chain) >= 2 {
		total = t
		summedSpans = chain
	} else {
		total = matches[0].Value
		summedSpans = []amountMatch{matches[0]}
	}

	dm := DetectDate(lower, now)

	tm := DetectType(lower)

	cr := DetectCategory(lower, tm.Type)

	dest := ""
	if tm.Type == "transfer" {
		dest, _ = FindTransferDestination(lower)
	}

	desc := buildDescription(lower, dm.Start, dm.End, summedSpans)

	description := desc
	if tm.Type == "transfer" && dest != "" {
		description = "Transfer ke " + TitleCaseID(dest)
	}
	if description == "" {
		switch {
		case cr.Merchant != "":
			description = cr.Merchant
		case cr.Name != "":
			description = cr.Name
		default:
			description = "Transaksi"
		}
	}

	conf := 0.40
	if tm.Explicit {
		conf += 0.35
	} else {
		conf += 0.10
	}
	if cr.Matched {
		conf += 0.15
	}
	if dm.Explicit {
		conf += 0.05
	}
	if cr.Merchant != "" {
		conf += 0.03
	}
	if tm.Type == "transfer" {
		if dest != "" {
			conf += 0.07
		} else {
			conf -= 0.15
		}
	}
	conf -= float64(CountHedges(lower)) * 0.15
	if len(summedSpans) >= 2 {
		conf += 0.05
	}

	conf = clampConfidence(conf)

	res := &Result{
		Type:        domain.TransactionType(tm.Type),
		Amount:      total,
		Description: description,
		Category:    cr.Name,
		Merchant:    cr.Merchant,
		Destination: dest,
		Date:        dm.Date,
		Confidence:  conf,
	}

	return res, nil
}

func buildDescription(lower string, dateStart, dateEnd int, amounts []amountMatch) string {
	var b strings.Builder
	for i := 0; i < len(lower); i++ {
		inDate := dateStart >= 0 && i >= dateStart && i < dateEnd
		if inDate {
			continue
		}
		inAmt := false
		for _, am := range amounts {
			if am.Start >= 0 && i >= am.Start && i < am.End {
				inAmt = true
				break
			}
		}
		if inAmt {
			continue
		}
		b.WriteByte(lower[i])
	}

	fields := strings.Fields(b.String())

	stopwords := map[string]bool{
		"untuk": true, "buat": true, "dengan": true, "sama": true,
		"rp": true, "rp.": true, "rupiah": true,

		"dan": true, "adn": true, "dn": true, "plus": true, "and": true,
		"&": true, "+": true,
		"kayaknya": true, "sepertinya": true, "mungkin": true,
		"sekitar": true, "kira-kira": true, "kurang": true, "lebih": true,
	}
	var kept []string
	for _, f := range fields {
		f = strings.Trim(f, ",.;:!?")
		if f == "" || stopwords[f] {
			continue
		}
		kept = append(kept, f)
	}

	leadingVerbs := map[string]bool{
		"beli": true, "belanja": true, "bayar": true,
		"bayarin": true, "jajan": true, "keluar": true,
	}

	pronouns := map[string]bool{
		"aku": true, "saya": true, "gue": true, "gw": true, "ak": true, "w": true,
	}
	for len(kept) > 1 && (leadingVerbs[kept[0]] || pronouns[kept[0]]) {
		kept = kept[1:]
	}

	if len(kept) > 6 {
		kept = kept[:6]
	}

	return strings.Join(kept, " ")
}

func TitleCaseID(s string) string {
	words := strings.Fields(s)
	var out []string
	for _, w := range words {
		if w == "" {
			continue
		}
		out = append(out, strings.ToUpper(w[:1])+w[1:])
	}
	return strings.Join(out, " ")
}

func clampConfidence(f float64) float64 {
	if f > 0.95 {
		f = 0.95
	}
	if f < 0 {
		f = 0
	}
	return float64(int(f*100+0.5)) / 100
}
