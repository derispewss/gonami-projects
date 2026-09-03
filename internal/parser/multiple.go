package parser

import (
	"regexp"
	"strconv"
	"strings"
	"time"
)

// reQuantity matches a small count followed by a unit word, e.g. "2 gelas",
// "3 porsi", "5 bungkus". Used to multiply a transaction amount when the
// quantity phrase appears AFTER the price ("kopi 5k 2 gelas" -> 10k).
var reQuantity = regexp.MustCompile(`(?i)\b(\d{1,3})\s+(gelas|porsi|buah|pasang|kotak|batang|bks|bungkus|lbr|lembar|pack|pcs|botol|cup|biji|butir|ekor|strip|sachet|liter|kg|plastik|cup|karton)\b`)

// ParseMultiple splits a text message into one or more transactions. Amounts
// joined by conjunctions ("dan", "plus", "&", ...) are treated as separate
// transactions instead of being summed, and a trailing quantity phrase
// multiplies the price.
func ParseMultiple(text string, now time.Time) ([]*Result, error) {
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

	// Find clause split points at the midpoint of conjunction gaps between
	// consecutive amounts.
	splitPts := []int{0}
	for i := 0; i+1 < len(matches); i++ {
		gap := lower[matches[i].End:matches[i+1].Start]
		if !hasConjunction(gap) {
			continue
		}
		mid := (matches[i].End + matches[i+1].Start) / 2
		if mid > splitPts[len(splitPts)-1] {
			splitPts = append(splitPts, mid)
		}
	}
	splitPts = append(splitPts, len(lower))

	var results []*Result
	for i := 0; i+1 < len(splitPts); i++ {
		clause := strings.TrimSpace(lower[splitPts[i]:splitPts[i+1]])
		if clause == "" {
			continue
		}
		if !clauseHasAmount(clause) {
			continue
		}
		parseClause := stripTrailingQuantity(clause)
		r, err := ParseDeterministic(parseClause, now)
		if err != nil {
			continue
		}
		applyQuantityMultiply(clause, r)
		results = append(results, r)
	}

	if len(results) == 0 {
		return nil, ErrNotTransaction
	}
	return results, nil
}

func clauseHasAmount(clause string) bool {
	_, ok := FindAmount(clause)
	return ok
}

// stripTrailingQuantity removes a trailing "N <unit>" phrase (e.g. "2 gelas")
// that follows the price so it does not leak into the description.
func stripTrailingQuantity(clause string) string {
	matches := findAllAmounts(clause)
	if len(matches) == 0 {
		return clause
	}
	last := matches[len(matches)-1]
	tail := clause[last.End:]
	loc := reQuantity.FindStringIndex(tail)
	if loc == nil {
		return clause
	}
	return strings.TrimSpace(clause[:last.End+loc[0]])
}

// applyQuantityMultiply multiplies a transaction amount when a quantity
// phrase ("N <unit>") appears after the price. Only applied for counts >= 2.
func applyQuantityMultiply(clause string, r *Result) {
	matches := findAllAmounts(clause)
	if len(matches) == 0 {
		return
	}
	last := matches[len(matches)-1]
	tail := clause[last.End:]
	loc := reQuantity.FindStringSubmatchIndex(tail)
	if loc == nil {
		return
	}
	n, err := strconv.Atoi(tail[loc[2]:loc[3]])
	if err != nil || n < 2 || n > 999 || r.Amount <= 0 {
		return
	}
	if r.Amount*int64(n) <= maxAmount {
		r.Amount *= int64(n)
	}
}
