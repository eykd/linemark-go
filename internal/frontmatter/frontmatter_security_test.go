package frontmatter

import "testing"

func TestSetTitle_RoundTrip_DoubleQuotesWithColon(t *testing.T) {
	// A title containing both double-quote characters and a colon triggers
	// the double-quoted YAML encoding path. Inner quotes must be escaped
	// so that the YAML round-trips correctly through GetTitle.
	input := "---\ntitle: Old\n---\nBody"
	title := `He said "hello": a greeting`

	result, err := SetTitle(input, title)
	if err != nil {
		t.Fatalf("SetTitle() error = %v", err)
	}

	got, err := GetTitle(result)
	if err != nil {
		t.Fatalf("GetTitle() error = %v", err)
	}
	if got != title {
		t.Errorf("round-trip failed: got %q, want %q", got, title)
	}
}

func TestSetTitle_RoundTrip_DoubleQuotesWithNewline(t *testing.T) {
	// A title with double-quote characters and a newline triggers
	// the double-quoted encoding path. Inner quotes must be escaped.
	input := "---\ntitle: Old\n---\nBody"
	title := "Line one \"quoted\"\nLine two"

	result, err := SetTitle(input, title)
	if err != nil {
		t.Fatalf("SetTitle() error = %v", err)
	}

	got, err := GetTitle(result)
	if err != nil {
		t.Fatalf("GetTitle() error = %v", err)
	}
	if got != title {
		t.Errorf("round-trip failed: got %q, want %q", got, title)
	}
}

func TestSetTitle_RoundTrip_HashAtStart(t *testing.T) {
	// A title starting with # could be interpreted as a YAML comment
	// if not properly quoted, causing the title to be lost on re-read.
	input := "---\ntitle: Old\n---\nBody"
	title := "#hashtag title"

	result, err := SetTitle(input, title)
	if err != nil {
		t.Fatalf("SetTitle() error = %v", err)
	}

	got, err := GetTitle(result)
	if err != nil {
		t.Fatalf("GetTitle() error = %v", err)
	}
	if got != title {
		t.Errorf("round-trip failed: got %q, want %q", got, title)
	}
}

func TestSetTitle_RoundTrip_BackslashEscapes(t *testing.T) {
	// A title containing backslash sequences that look like YAML escapes
	// (e.g., \n literal) should round-trip without being interpreted.
	input := "---\ntitle: Old\n---\nBody"
	title := `path: C:\new\folder`

	result, err := SetTitle(input, title)
	if err != nil {
		t.Fatalf("SetTitle() error = %v", err)
	}

	got, err := GetTitle(result)
	if err != nil {
		t.Fatalf("GetTitle() error = %v", err)
	}
	if got != title {
		t.Errorf("round-trip failed: got %q, want %q", got, title)
	}
}

func TestEncodeYAMLValue_EscapesDoubleQuotes(t *testing.T) {
	// When a value contains a colon (triggering double-quoted output),
	// any double-quote characters inside must be escaped.
	tests := []struct {
		name  string
		input string
	}{
		{"quote with colon", `say "hello": greeting`},
		{"quote with newline", "say \"hello\"\nworld"},
		{"multiple quotes with colon", `"a": "b"`},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			encoded := EncodeYAMLValue(tt.input)
			// The encoded value should be valid when placed in a YAML document
			doc := "---\ntitle: " + encoded + "\n---\n"
			got, err := GetTitle(doc)
			if err != nil {
				t.Fatalf("GetTitle() error = %v (encoded = %q)", err, encoded)
			}
			if got != tt.input {
				t.Errorf("round-trip failed: got %q, want %q (encoded = %q)", got, tt.input, encoded)
			}
		})
	}
}
