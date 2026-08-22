package parser

import (
	"math"
	"regexp"
	"strconv"
	"strings"
)

var digitWords = map[string]float64{
	"nol": 0, "satu": 1, "dua": 2, "tiga": 3, "empat": 4, "lima": 5,
	"enam": 6, "tujuh": 7, "delapan": 8, "sembilan": 9,
}

var reGlued = regexp.MustCompile(
	`(satu|dua|tiga|empat|lima|enam|tujuh|delapan|sembilan)(puluh|belas|ratus)\b`)

var reSePrefix = regexp.MustCompile(`\bse(puluh|belas|ratus|ribu)\b`)

func normalizeGluedWords(s string) string {
	s = reGlued.ReplaceAllString(s, "$1 $2")
	s = reSePrefix.ReplaceAllString(s, "satu $1")
	return s
}

func scanSequence(words []string, i int) (value float64, consumed int, endedWithScale bool) {
	var (
		h, t, u    float64
		pending    float64
		pendingSet bool
		frac       float64
		fracMode   bool
		fracDigits int
		total      float64
		scaleSeen  bool
	)

	groupValue := func() float64 { return h*100 + t*10 + u + frac }

	fail := func(j int) (float64, int, bool) {
		return finish(total+groupValue(), j-i, scaleSeen)
	}

	j := i
	for ; j < len(words); j++ {
		w := words[j]
		switch w {
		case "koma":
			if fracMode || scaleSeen || (!pendingSet && h == 0 && t == 0 && u == 0) {
				return fail(j)
			}
			fracMode, fracDigits = true, 0

		case "puluh":
			if fracMode || scaleSeen || !pendingSet {
				return fail(j)
			}
			t = pending
			pending, pendingSet = 0, false

		case "belas":
			if fracMode || scaleSeen || !pendingSet {
				return fail(j)
			}
			u = 10 + pending
			pending, pendingSet = 0, false

		case "ratus":
			if fracMode || scaleSeen || !pendingSet {
				return fail(j)
			}
			h = pending
			pending, pendingSet = 0, false

		case "ribu", "juta":
			if fracMode && fracDigits == 0 {
				return fail(j)
			}
			if pendingSet {
				if t != 0 || u != 0 {
					return fail(j)
				}
				u = pending
				pending, pendingSet = 0, false
			}
			mult := groupValue()
			if mult == 0 {
				mult = 1
			}
			if w == "ribu" {
				total += mult * 1000
			} else {
				total += mult * 1000000
			}
			h, t, u = 0, 0, 0
			frac, fracMode, fracDigits = 0, false, 0
			scaleSeen = true

		default:
			d, isDigit := digitWords[w]
			if !isDigit {
				return fail(j)
			}
			if fracMode {
				fracDigits++
				frac += d / math.Pow(10, float64(fracDigits))
				continue
			}
			switch {
			case pendingSet:
				return fail(j)
			case u != 0:
				return fail(j)
			case t != 0:
				u = d
			default:
				pending, pendingSet = d, true
			}
		}
	}
	return finish(total+groupValue(), j-i, scaleSeen)
}

func finish(value float64, consumed int, scale bool) (float64, int, bool) {
	if consumed < 0 {
		consumed = 0
	}
	return value, consumed, scale
}

func convertSpelledNumbers(text string) string {
	text = normalizeGluedWords(text)
	words := strings.Fields(text)

	out := make([]string, 0, len(words))
	replaced := false

	i := 0
	for i < len(words) {
		val, consumed, withScale := scanSequence(words, i)
		if withScale && consumed >= 2 && val > 0 {
			out = append(out, formatNumberWordValue(val))
			i += consumed
			replaced = true
			continue
		}
		out = append(out, words[i])
		i++
	}

	if !replaced {
		return strings.Join(words, " ")
	}
	return strings.Join(out, " ")
}

func formatNumberWordValue(v float64) string {
	rounded := math.Round(v*100) / 100
	if rounded == math.Trunc(rounded) {
		return strconv.FormatInt(int64(rounded), 10)
	}

	return strconv.FormatInt(int64(math.Round(rounded)), 10)
}
