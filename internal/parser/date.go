package parser

import (
	"fmt"
	"strings"
	"time"
)

var monthNamesID = map[string]time.Month{
	"januari": time.January, "jan": time.January,
	"februari": time.February, "feb": time.February,
	"maret": time.March, "mar": time.March,
	"april": time.April, "apr": time.April,
	"mei":  time.May,
	"juni": time.June, "jun": time.June,
	"juli": time.July, "jul": time.July,
	"agustus": time.August, "agu": time.August, "agt": time.August,
	"september": time.September, "sep": time.September,
	"oktober": time.October, "okt": time.October,
	"november": time.November, "nov": time.November,
	"desember": time.December, "des": time.December,
}

type dateMatch struct {
	Date       time.Time
	Explicit   bool
	Start, End int
}

func DetectDate(lower string, now time.Time) dateMatch {
	today := startOfDay(now)
	yesterday := today.AddDate(0, 0, -1)

	markers := []struct {
		words []string
		day   time.Time
	}{
		{[]string{"kemarin lusa"}, today.AddDate(0, 0, -2)},
		{[]string{"kemarin"}, yesterday},
		{[]string{"hari ini"}, today},
		{[]string{"barusan"}, today},
		{[]string{"baru saja"}, today},
		{[]string{"tadi"}, today},
		{[]string{"semalam"}, yesterday},
	}

	for _, m := range markers {
		for _, w := range m.words {
			idx := indexOfPhrase(lower, w)
			if idx >= 0 {
				return dateMatch{Date: m.day, Explicit: true, Start: idx, End: idx + len(w)}
			}
		}
	}

	if dm, ok := detectDayMonth(lower, now); ok {
		return dm
	}

	return dateMatch{Date: today, Explicit: false}
}

func detectDayMonth(lower string, now time.Time) (dateMatch, bool) {
	var day int

	fields := strings.Fields(lower)
	for i := 0; i < len(fields)-1; i++ {
		d := strings.TrimSuffix(fields[i], ".")
		if !isDigits(d) || len(d) > 2 {
			continue
		}
		m := strings.TrimSuffix(fields[i+1], ".")
		if mon, ok := monthNamesID[m]; ok {
			day = atoi(d)
			date := time.Date(now.Year(), mon, day, 12, 0, 0, 0, now.Location())

			if date.After(now.AddDate(0, 0, 30)) {
				date = date.AddDate(-1, 0, 0)
			}
			start := strings.Index(lower, fields[i])
			end := start + len(fields[i]) + 1 + len(fields[i+1])
			if end > len(lower) {
				end = len(lower)
			}
			return dateMatch{
				Date:     startOfDay(date),
				Explicit: true,
				Start:    start,
				End:      end,
			}, true
		}
	}
	return dateMatch{}, false
}

func startOfDay(t time.Time) time.Time {
	return time.Date(t.Year(), t.Month(), t.Day(), 0, 0, 0, 0, t.Location())
}

func indexOfPhrase(s, phrase string) int {
	search := 0
	for {
		idx := strings.Index(s[search:], phrase)
		if idx < 0 {
			return -1
		}
		idx += search
		beforeOK := idx == 0 || !isWordChar(rune(s[idx-1]))
		end := idx + len(phrase)
		afterOK := end >= len(s) || !isWordChar(rune(s[end]))
		if beforeOK && afterOK {
			return idx
		}
		search = idx + 1
		if search >= len(s) {
			return -1
		}
	}
}

func isWordChar(r rune) bool {
	return r == '_' || (r >= 'a' && r <= 'z') || (r >= 'A' && r <= 'Z') ||
		(r >= '0' && r <= '9')
}

func isDigits(s string) bool {
	if s == "" {
		return false
	}
	for _, r := range s {
		if r < '0' || r > '9' {
			return false
		}
	}
	return true
}

func atoi(s string) int {
	n := 0
	fmt.Sscanf(s, "%d", &n)
	return n
}
