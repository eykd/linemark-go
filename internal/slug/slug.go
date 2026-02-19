package slug

import (
	"strings"
	"unicode"

	"golang.org/x/text/unicode/norm"
)

// Slug converts a string into a URL-friendly slug.
// It NFD-normalizes, strips combining marks, lowercases,
// converts whitespace to dashes, strips non-alphanumeric
// characters, and collapses consecutive dashes.
func Slug(s string) string {
	normalized := norm.NFD.String(s)

	var b strings.Builder
	prevDash := false

	for _, r := range normalized {
		switch {
		case unicode.Is(unicode.Mn, r):
			continue
		case unicode.IsLetter(r), unicode.IsDigit(r):
			b.WriteRune(unicode.ToLower(r))
			prevDash = false
		case unicode.IsSpace(r), r == '-':
			if !prevDash && b.Len() > 0 {
				b.WriteRune('-')
				prevDash = true
			}
		}
	}

	return strings.TrimRight(b.String(), "-")
}
