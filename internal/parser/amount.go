package parser

import (
	"math"
	"regexp"
	"strconv"
	"strings"
)

var (
	reJuta = regexp.MustCompile(`(?i)(?:rp\s*\.?\s*)?(\d+(?:[.,]\d+)?)\s*(?:juta|jt)\b`)

	reKilo = regexp.MustCompile(`(?i)(?:rp\s*\.?\s*)?(\d+)\s*(?:k|rb|ribu(?:an)?)\b`)

	reDot = regexp.MustCompile(`(?:rp\s*\.?\s*)?(\d{1,3}(?:\.\d{3})+)`)

	rePlain = regexp.MustCompile(`\brp\s*\.?\s*(\d+)\b|\b(\d{4,12})\b`)
)

type amountMatch struct {
	Value      int64
	Start, End int
	Kind       amountKind
}

func FindAmount(text string) (amountMatch, bool) {
	text = convertSpelledNumbers(text)
	return findAmountRaw(text)
}

func findAmountRaw(text string) (amountMatch, bool) {
	for _, fn := range []func(string) (amountMatch, bool){
		findJuta, findKilo, findDot, findPlain,
	} {
		if m, ok := fn(text); ok {
			return m, true
		}
	}
	return amountMatch{}, false
}

func findJuta(text string) (amountMatch, bool) {
	loc := reJuta.FindStringSubmatchIndex(text)
	if loc == nil {
		return amountMatch{}, false
	}
	numStr := text[loc[2]:loc[3]]
	val, err := parseDecimalNumber(numStr)
	if err != nil || val <= 0 || val > maxAmount {
		return amountMatch{}, false
	}
	return amountMatch{Value: int64(val * 1_000_000), Start: loc[0], End: loc[1]}, true
}

func findKilo(text string) (amountMatch, bool) {
	loc := reKilo.FindStringSubmatchIndex(text)
	if loc == nil {
		return amountMatch{}, false
	}
	numStr := strings.ReplaceAll(text[loc[2]:loc[3]], ".", "")
	n, err := strconv.ParseInt(numStr, 10, 64)
	if err != nil || n <= 0 || n > maxAmount {
		return amountMatch{}, false
	}
	return amountMatch{Value: n * 1_000, Start: loc[0], End: loc[1]}, true
}

func findDot(text string) (amountMatch, bool) {
	loc := reDot.FindStringSubmatchIndex(text)
	if loc == nil {
		return amountMatch{}, false
	}
	numStr := strings.ReplaceAll(text[loc[2]:loc[3]], ".", "")
	n, err := strconv.ParseInt(numStr, 10, 64)
	if err != nil || n <= 0 || n > maxAmount {
		return amountMatch{}, false
	}
	return amountMatch{Value: n, Start: loc[0], End: loc[1]}, true
}

func findPlain(text string) (amountMatch, bool) {
	loc := rePlain.FindStringSubmatchIndex(text)
	if loc == nil {
		return amountMatch{}, false
	}

	var numStr string

	if loc[2] >= 0 {
		numStr = text[loc[2]:loc[3]]
	} else if loc[4] >= 0 {

		numStr = text[loc[4]:loc[5]]
	}

	if numStr == "" {
		return amountMatch{}, false
	}

	n, err := strconv.ParseInt(strings.ReplaceAll(numStr, ".", ""), 10, 64)
	if err != nil || n <= 0 || n > maxAmount {
		return amountMatch{}, false
	}
	return amountMatch{Value: n, Start: loc[0], End: loc[1]}, true
}

func parseDecimalNumber(s string) (float64, error) {
	s = strings.ReplaceAll(s, ",", ".")
	f, err := strconv.ParseFloat(s, 64)
	if err != nil || math.IsNaN(f) || math.IsInf(f, 0) {
		return 0, err
	}
	return f, nil
}

const maxAmount = 99_000_000_000
