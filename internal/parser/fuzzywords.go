package parser

// Lightweight Indonesian stemming + fuzzy word matching for the rule-based
// Layer-1 parser. These run fully offline (no LLM) so the bot can keep
// understanding common transactions even when the AI APIs are exhausted.

import (
	"strings"
	"unicode/utf8"
)

// stems registers a small set of Indonesian affix rules tuned for the finance
// verbs used in gonami (beli, bayar, jajan, isi, kirim, jual, dll).
func stemWord(w string) string {
	w = strings.ToLower(strings.Trim(strings.TrimSpace(w), ",.;:!?()\"'"))
	if len([]rune(w)) < 4 {
		return w
	}

	// words that must not be stemmed (already root / proper)
	root := map[string]bool{
		"beli": true, "bayar": true, "jajan": true, "makan": true,
		"minum": true, "isi": true, "jual": true, "kirim": true,
		"masuk": true, "keluar": true, "dapat": true, "terima": true,
		"bonus": true, "gaji": true, "transfer": true, "langganan": true,
		"token": true, "listrik": true, "pulsa": true, "sewa": true,
		"parkir": true, "cicil": true, "topup": true,
	}
	if root[w] {
		return w
	}

	// strip suffix first (kan/i/an/an)
	for _, suf := range []string{"kan", "kah", "lah", "an", "i", "nya"} {
		if strings.HasSuffix(w, suf) && len([]rune(w))-len([]rune(suf)) >= 3 {
			candidate := w[:len(w)-len(suf)]
			if isLikelyVerb(candidate) {
				w = candidate
				break
			}
		}
	}

	// strip prefixes
	for _, pre := range []string{"memper", "member", "meny", "meng", "mem",
		"men", "me", "diper", "dik", "di", "per", "ber", "ter", "pe"} {
		if strings.HasPrefix(w, pre) && len([]rune(w))-len([]rune(pre)) >= 3 {
			candidate := w[len(pre):]
			if isLikelyVerb(candidate) {
				w = candidate
				break
			}
		}
	}

	// common vowel harmony cleanups after stripping
	w = strings.ReplaceAll(w, "ny", "")
	return w
}

func isLikelyVerb(w string) bool {
	// requires at least 3 letters and end in a vowel-ish consonant cluster
	n := utf8.RuneCountInString(w)
	if n < 3 {
		return false
	}
	_, ok := digitWords[w]
	if ok {
		return false
	}
	return true
}

func stemSentence(s string) string {
	words := strings.Fields(s)
	for i, w := range words {
		words[i] = stemWord(w)
	}
	return strings.Join(words, " ")
}

// levenshtein returns the edit distance between two strings.
func levenshtein(a, b string) int {
	ra := []rune(strings.ToLower(a))
	rb := []rune(strings.ToLower(b))
	la, lb := len(ra), len(rb)
	if la == 0 {
		return lb
	}
	if lb == 0 {
		return la
	}
	prev := make([]int, lb+1)
	curr := make([]int, lb+1)
	for j := 0; j <= lb; j++ {
		prev[j] = j
	}
	for i := 1; i <= la; i++ {
		curr[0] = i
		for j := 1; j <= lb; j++ {
			cost := 1
			if ra[i-1] == rb[j-1] {
				cost = 0
			}
			curr[j] = min3(curr[j-1]+1, prev[j]+1, prev[j-1]+cost)
		}
		prev, curr = curr, prev
	}
	return prev[lb]
}

func min3(a, b, c int) int {
	if b < a {
		a = b
	}
	if c < a {
		a = c
	}
	return a
}

// fuzzyMatchWord reports whether any token in text matches keyword within
// maxDist edit distance (typo tolerance). Longer keywords allow more distance.
func fuzzyMatchWord(text, keyword string, maxDist int) bool {
	kw := strings.ToLower(keyword)
	for _, tok := range strings.Fields(text) {
		tok = strings.ToLower(strings.Trim(tok, ",.;:!?()\"'-/"))
		if tok == kw {
			return true
		}
		if utf8.RuneCountInString(tok) >= 4 && levenshtein(tok, kw) <= maxDist {
			return true
		}
	}
	return false
}
