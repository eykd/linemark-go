package slug

import "testing"

func TestSlug(t *testing.T) {
	tests := []struct {
		name  string
		input string
		want  string
	}{
		{"empty input", "", ""},
		{"simple lowercase", "hello", "hello"},
		{"uppercase to lowercase", "Hello World", "hello-world"},
		{"spaces to dashes", "one two three", "one-two-three"},
		{"multiple spaces collapse", "one   two", "one-two"},
		{"tabs to dashes", "one\ttwo", "one-two"},
		{"newlines to dashes", "one\ntwo", "one-two"},
		{"mixed whitespace", "one \t\n two", "one-two"},
		{"leading whitespace trimmed", "  hello", "hello"},
		{"trailing whitespace trimmed", "hello  ", "hello"},
		{"leading and trailing whitespace", "  hello  ", "hello"},
		{"diacritics removed", "Café", "cafe"},
		{"umlaut removed", "Über", "uber"},
		{"multiple diacritics", "résumé", "resume"},
		{"special chars stripped", "hello!world", "helloworld"},
		{"mixed special and spaces", "hello - world!", "hello-world"},
		{"multiple dashes collapse", "hello---world", "hello-world"},
		{"all special chars returns empty", "!!!", ""},
		{"em dash returns empty", "\u2014", ""},
		{"numbers preserved", "chapter-1", "chapter-1"},
		{"mixed alphanumeric", "Part 2: The Return", "part-2-the-return"},
		{"unicode normalization", "na\u0308ive", "naive"},
		{"leading dashes trimmed", "---hello", "hello"},
		{"trailing dashes trimmed", "hello---", "hello"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := Slug(tt.input)
			if got != tt.want {
				t.Errorf("Slug(%q) = %q, want %q", tt.input, got, tt.want)
			}
		})
	}
}
