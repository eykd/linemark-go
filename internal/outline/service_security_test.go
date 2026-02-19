package outline

import (
	"testing"

	"github.com/eykd/linemark-go/internal/frontmatter"
)

func TestFormatFrontmatter_NewlineInTitle_RoundTrips(t *testing.T) {
	// formatFrontmatter must produce valid YAML even when the title
	// contains newline characters. The result should round-trip through
	// frontmatter.GetTitle without losing or corrupting data.
	tests := []struct {
		name  string
		title string
	}{
		{
			"title with newline",
			"Line one\nLine two",
		},
		{
			"title with newline and colon",
			"Part 1: Introduction\nPart 2: Body",
		},
		{
			"title with multiple newlines",
			"A\nB\nC",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			content := formatFrontmatter(tt.title)

			got, err := frontmatter.GetTitle(content)
			if err != nil {
				t.Fatalf("GetTitle() error = %v, content = %q", err, content)
			}
			if got != tt.title {
				t.Errorf("round-trip failed: got %q, want %q", got, tt.title)
			}
		})
	}
}

func TestFormatFrontmatter_DoubleQuotesWithColon_RoundTrips(t *testing.T) {
	// formatFrontmatter must handle double-quote characters inside the title
	// when combined with colons that trigger quoting. Inner quotes need escaping.
	title := `He said "hello": a greeting`
	content := formatFrontmatter(title)

	got, err := frontmatter.GetTitle(content)
	if err != nil {
		t.Fatalf("GetTitle() error = %v, content = %q", err, content)
	}
	if got != title {
		t.Errorf("round-trip failed: got %q, want %q", got, title)
	}
}

func TestFormatFrontmatter_YAMLInjection_Prevented(t *testing.T) {
	// A crafted title should not inject additional YAML keys into the
	// frontmatter. Even if the title contains YAML-like key: value syntax,
	// it must be encoded as a single scalar value.
	title := "foo\nevil_key: injected_value"
	content := formatFrontmatter(title)

	fm, _, err := frontmatter.Split(content)
	if err != nil {
		t.Fatalf("Split() error = %v", err)
	}

	// The frontmatter should not contain "evil_key" as a separate YAML key
	got, err := frontmatter.GetTitle(content)
	if err != nil {
		t.Fatalf("GetTitle() error = %v, fm = %q", err, fm)
	}
	if got != title {
		t.Errorf("title round-trip failed: got %q, want %q", got, title)
	}
}
