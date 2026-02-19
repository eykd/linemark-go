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
	// NFD normalize to decompose characters.
	s = norm.NFD.String(s)

	// Strip combining (Mn) marks and lowercase.
	var b strings.Builder
	for _, r := range s {
		if unicode.Is(unicode.Mn, r) {
			continue
		}
		b.WriteRune(unicode.ToLower(r))
	}
	s = b.String()

	// Replace whitespace with dashes.
	b.Reset()
	for _, r := range s {
		if unicode.IsSpace(r) {
			b.WriteRune('-')
		} else {
			b.WriteRune(r)
		}
	}
	s = b.String()

	// Strip non-alphanumeric, non-dash characters.
	b.Reset()
	for _, r := range s {
		if r == '-' || unicode.IsLetter(r) || unicode.IsDigit(r) {
			b.WriteRune(r)
		}
	}
	s = b.String()

	// Collapse consecutive dashes.
	for strings.Contains(s, "--") {
		s = strings.ReplaceAll(s, "--", "-")
	}

	// Trim leading and trailing dashes.
	s = strings.Trim(s, "-")

	return s
}
