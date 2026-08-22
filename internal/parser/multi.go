package parser

import (
	"sort"
	"strconv"
	"strings"
)

type amountKind int

const (
	kindBarePlain amountKind = iota
	kindRpPlain
	kindDot
	kindKilo
	kindJuta
)

func (m amountMatch) sumEligible() bool { return m.Kind != kindBarePlain }

var conjunctionTokens = map[string]bool{
	"dan": true, "adn": true, "dn": true,
	"plus": true, "and": true, "&": true, "+": true,
}

func findAllAmounts(text string) []amountMatch {
	var accepted []amountMatch

	addNonOverlapping := func(ms []amountMatch) {
		for _, m := range ms {
			overlap := false
			for _, a := range accepted {
				if m.Start < a.End && a.Start < m.End {
					overlap = true
					break
				}
			}
			if !overlap {
				accepted = append(accepted, m)
			}
		}
	}

	for _, loc := range reJuta.FindAllStringSubmatchIndex(text, -1) {
		val, err := parseDecimalNumber(text[loc[2]:loc[3]])
		if err != nil || val <= 0 || val > maxAmount {
			continue
		}
		addNonOverlapping([]amountMatch{{
			Value: int64(val * 1_000_000), Start: loc[0], End: loc[1], Kind: kindJuta,
		}})
	}

	for _, loc := range reKilo.FindAllStringSubmatchIndex(text, -1) {
		n, err := strconv.ParseInt(strings.ReplaceAll(text[loc[2]:loc[3]], ".", ""), 10, 64)
		if err != nil || n <= 0 || n > maxAmount {
			continue
		}
		addNonOverlapping([]amountMatch{{
			Value: n * 1_000, Start: loc[0], End: loc[1], Kind: kindKilo,
		}})
	}

	for _, loc := range reDot.FindAllStringSubmatchIndex(text, -1) {
		n, err := strconv.ParseInt(strings.ReplaceAll(text[loc[2]:loc[3]], ".", ""), 10, 64)
		if err != nil || n <= 0 || n > maxAmount {
			continue
		}
		addNonOverlapping([]amountMatch{{
			Value: n, Start: loc[0], End: loc[1], Kind: kindDot,
		}})
	}

	for _, loc := range rePlain.FindAllStringSubmatchIndex(text, -1) {
		kind := kindBarePlain
		var numStr string
		if loc[2] >= 0 {
			kind = kindRpPlain
			numStr = text[loc[2]:loc[3]]
		} else if loc[4] >= 0 {
			numStr = text[loc[4]:loc[5]]
		}
		if numStr == "" {
			continue
		}
		n, err := strconv.ParseInt(strings.ReplaceAll(numStr, ".", ""), 10, 64)
		if err != nil || n <= 0 || n > maxAmount {
			continue
		}
		addNonOverlapping([]amountMatch{{
			Value: n, Start: loc[0], End: loc[1], Kind: kind,
		}})
	}

	sort.Slice(accepted, func(i, j int) bool { return accepted[i].Start < accepted[j].Start })
	return accepted
}

func hasConjunction(gap string) bool {
	for _, f := range strings.Fields(gap) {
		if conjunctionTokens[strings.Trim(f, ",.;:!?()\"'")] {
			return true
		}
	}
	return false
}

func sumItemChain(text string, matches []amountMatch) (int64, []amountMatch) {
	start := -1
	for i, m := range matches {
		if m.sumEligible() {
			start = i
			break
		}
	}
	if start < 0 {
		return 0, nil
	}

	chain := []amountMatch{matches[start]}
	total := matches[start].Value
	for i := start + 1; i < len(matches); i++ {
		next := matches[i]
		if !next.sumEligible() {
			break
		}
		gap := text[chain[len(chain)-1].End:next.Start]
		if !hasConjunction(gap) {
			break
		}
		if total+next.Value > maxAmount {
			break
		}
		total += next.Value
		chain = append(chain, next)
	}
	if len(chain) < 2 {
		return 0, nil
	}
	return total, chain
}
