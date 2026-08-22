package format

import (
	"fmt"
	"strings"
	"time"
	"unicode"
)

var monthsID = []string{
	"Januari", "Februari", "Maret", "April", "Mei", "Juni",
	"Juli", "Agustus", "September", "Oktober", "November", "Desember",
}

var daysID = map[time.Weekday]string{
	time.Sunday:    "Minggu",
	time.Monday:    "Senin",
	time.Tuesday:   "Selasa",
	time.Wednesday: "Rabu",
	time.Thursday:  "Kamis",
	time.Friday:    "Jumat",
	time.Saturday:  "Sabtu",
}

func Rupiah(amount int64) string {
	neg := amount < 0
	if neg {
		amount = -amount
	}
	digits := fmt.Sprintf("%d", amount)
	var b strings.Builder
	pre := len(digits) % 3
	if pre > 0 {
		b.WriteString(digits[:pre])
	}
	for i := pre; i < len(digits); i += 3 {
		if b.Len() > 0 {
			b.WriteByte('.')
		}
		b.WriteString(digits[i : i+3])
	}
	out := b.String()
	if out == "" {
		out = "0"
	}
	if neg {
		return "-Rp" + out
	}
	return "Rp" + out
}

func DateLongID(t time.Time) string {
	return fmt.Sprintf("%s, %d %s %d",
		daysID[t.Weekday()], t.Day(), monthsID[t.Month()-1], t.Year())
}

func DateShortID(t time.Time) string {
	short := []string{"Jan", "Feb", "Mar", "Apr", "Mei", "Jun",
		"Jul", "Agu", "Sep", "Okt", "Nov", "Des"}
	return fmt.Sprintf("%02d %s", t.Day(), short[int(t.Month())-1])
}

func MonthYearID(t time.Time) string {
	return fmt.Sprintf("%s %d", monthsID[int(t.Month())-1], t.Year())
}

func CategoryEmoji(name string) string {
	switch name {
	case "Food & Beverage":
		return "🍜"
	case "Transportation":
		return "🚗"
	case "Shopping":
		return "🛒"
	case "Bills":
		return "💡"
	case "Entertainment":
		return "🎬"
	case "Health":
		return "💊"
	case "Education":
		return "📚"
	case "Travel":
		return "✈️"
	case "Subscription":
		return "🔁"
	case "Salary":
		return "💼"
	case "Freelance":
		return "💻"
	case "Business":
		return "🏪"
	case "Gift":
		return "🎁"
	case "Investment":
		return "📈"
	default:
		return "📦"
	}
}

func TitleCase(s string) string {
	words := strings.Fields(s)
	for i, w := range words {
		r := []rune(w)
		r[0] = unicode.ToUpper(r[0])
		words[i] = string(r)
	}
	return strings.Join(words, " ")
}

func Truncate(s string, n int) string {
	runes := []rune(s)
	if len(runes) <= n {
		return s
	}
	if n <= 3 {
		return string(runes[:n])
	}
	return string(runes[:n-3]) + "..."
}
