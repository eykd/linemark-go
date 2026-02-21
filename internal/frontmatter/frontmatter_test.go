package frontmatter

import (
	"strings"
	"testing"
)

func TestSplit(t *testing.T) {
	tests := []struct {
		name    string
		input   string
		wantFM  string
		wantBod string
		wantErr bool
	}{
		{
			"valid frontmatter with body",
			"---\ntitle: Hello\n---\nBody text",
			"title: Hello\n",
			"Body text",
			false,
		},
		{
			"empty frontmatter",
			"---\n---\nBody text",
			"",
			"Body text",
			false,
		},
		{
			"no frontmatter",
			"Just body text",
			"",
			"Just body text",
			false,
		},
		{
			"frontmatter only no body",
			"---\ntitle: Hello\n---\n",
			"title: Hello\n",
			"",
			false,
		},
		{
			"frontmatter at EOF without trailing newline",
			"---\ntitle: Hello\n---",
			"title: Hello\n",
			"",
			false,
		},
		{
			"unclosed frontmatter",
			"---\ntitle: Hello\n",
			"",
			"",
			true,
		},
		{
			"multiple fields preserved",
			"---\ntitle: Hello\nauthor: Alice\ntags: [a, b]\n---\nBody",
			"title: Hello\nauthor: Alice\ntags: [a, b]\n",
			"Body",
			false,
		},
		{
			"body with dashes not confused for delimiter",
			"---\ntitle: Hello\n---\nSome text\n---\nMore text",
			"title: Hello\n",
			"Some text\n---\nMore text",
			false,
		},
		{
			"empty document",
			"",
			"",
			"",
			false,
		},
		{
			"only opening delimiter",
			"---\n",
			"",
			"",
			true,
		},
		{
			"body contains triple dashes mid-line",
			"---\ntitle: X\n---\ntext --- more",
			"title: X\n",
			"text --- more",
			false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			fm, body, err := Split(tt.input)

			if (err != nil) != tt.wantErr {
				t.Errorf("Split() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if err != nil {
				return
			}
			if fm != tt.wantFM {
				t.Errorf("Split() frontmatter = %q, want %q", fm, tt.wantFM)
			}
			if body != tt.wantBod {
				t.Errorf("Split() body = %q, want %q", body, tt.wantBod)
			}
		})
	}
}

func TestGetTitle(t *testing.T) {
	tests := []struct {
		name    string
		input   string
		want    string
		wantErr bool
	}{
		{
			"simple title",
			"---\ntitle: My Title\n---\nBody",
			"My Title",
			false,
		},
		{
			"quoted title",
			"---\ntitle: \"Quoted Title\"\n---\n",
			"Quoted Title",
			false,
		},
		{
			"title with unicode",
			"---\ntitle: Café Épée\n---\n",
			"Café Épée",
			false,
		},
		{
			"missing title field",
			"---\nauthor: Alice\n---\nBody",
			"",
			false,
		},
		{
			"empty frontmatter",
			"---\n---\nBody",
			"",
			false,
		},
		{
			"no frontmatter",
			"Just body text",
			"",
			false,
		},
		{
			"bool title returns error",
			"---\ntitle: true\n---\n",
			"",
			true,
		},
		{
			"integer title returns error",
			"---\ntitle: 42\n---\n",
			"",
			true,
		},
		{
			"malformed yaml returns error",
			"---\ntitle: [unclosed\n---\n",
			"",
			true,
		},
		{
			"title among other fields",
			"---\nauthor: Bob\ntitle: Found It\ntags: [x]\n---\n",
			"Found It",
			false,
		},
		{
			"empty title value",
			"---\ntitle: \"\"\n---\n",
			"",
			false,
		},
		{
			"title with colon",
			"---\ntitle: \"Part 1: The Beginning\"\n---\n",
			"Part 1: The Beginning",
			false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := GetTitle(tt.input)

			if (err != nil) != tt.wantErr {
				t.Errorf("GetTitle() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if err != nil {
				return
			}
			if got != tt.want {
				t.Errorf("GetTitle() = %q, want %q", got, tt.want)
			}
		})
	}
}

func TestSetTitle(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		newTitle string
		wantErr  bool
		check    func(t *testing.T, result string)
	}{
		{
			"update existing title",
			"---\ntitle: Old Title\n---\nBody text",
			"New Title",
			false,
			func(t *testing.T, result string) {
				t.Helper()
				title, err := GetTitle(result)
				if err != nil {
					t.Fatalf("GetTitle on result: %v", err)
				}
				if title != "New Title" {
					t.Errorf("title = %q, want %q", title, "New Title")
				}
				_, body, err := Split(result)
				if err != nil {
					t.Fatalf("Split on result: %v", err)
				}
				if body != "Body text" {
					t.Errorf("body = %q, want %q", body, "Body text")
				}
			},
		},
		{
			"preserves unknown fields",
			"---\nauthor: Alice\ntitle: Old\ntags: [a, b]\n---\nBody",
			"New",
			false,
			func(t *testing.T, result string) {
				t.Helper()
				fm, _, err := Split(result)
				if err != nil {
					t.Fatalf("Split: %v", err)
				}
				if !strings.Contains(fm, "author") {
					t.Error("lost author field")
				}
				if !strings.Contains(fm, "tags") {
					t.Error("lost tags field")
				}
			},
		},
		{
			"preserves field order",
			"---\nauthor: Alice\ntitle: Old\ntags: [a]\n---\n",
			"New",
			false,
			func(t *testing.T, result string) {
				t.Helper()
				fm, _, err := Split(result)
				if err != nil {
					t.Fatalf("Split: %v", err)
				}
				authorIdx := strings.Index(fm, "author")
				titleIdx := strings.Index(fm, "title")
				tagsIdx := strings.Index(fm, "tags")
				if authorIdx >= titleIdx {
					t.Error("author should appear before title")
				}
				if titleIdx >= tagsIdx {
					t.Error("title should appear before tags")
				}
			},
		},
		{
			"adds title when missing",
			"---\nauthor: Alice\n---\nBody",
			"Added Title",
			false,
			func(t *testing.T, result string) {
				t.Helper()
				title, err := GetTitle(result)
				if err != nil {
					t.Fatalf("GetTitle: %v", err)
				}
				if title != "Added Title" {
					t.Errorf("title = %q, want %q", title, "Added Title")
				}
				fm, _, err := Split(result)
				if err != nil {
					t.Fatalf("Split: %v", err)
				}
				if !strings.Contains(fm, "author") {
					t.Error("lost author field")
				}
			},
		},
		{
			"title with yaml special chars is safe",
			"---\ntitle: Old\n---\n",
			"foo\nnew_key: injected",
			false,
			func(t *testing.T, result string) {
				t.Helper()
				title, err := GetTitle(result)
				if err != nil {
					t.Fatalf("GetTitle: %v", err)
				}
				if title != "foo\nnew_key: injected" {
					t.Errorf("title = %q, want exact string with newline", title)
				}
				fm, _, err := Split(result)
				if err != nil {
					t.Fatalf("Split: %v", err)
				}
				// The injected string should NOT appear as a separate YAML key
				if strings.Contains(fm, "new_key: injected") {
					t.Error("YAML injection: new_key appeared as separate field")
				}
			},
		},
		{
			"unclosed frontmatter returns error",
			"---\ntitle: Hello\n",
			"New",
			true,
			nil,
		},
		{
			"malformed yaml returns error",
			"---\ntitle: [unclosed\n---\n",
			"New",
			true,
			nil,
		},
		{
			"empty frontmatter adds title",
			"---\n---\nBody",
			"New Title",
			false,
			func(t *testing.T, result string) {
				t.Helper()
				title, err := GetTitle(result)
				if err != nil {
					t.Fatalf("GetTitle: %v", err)
				}
				if title != "New Title" {
					t.Errorf("title = %q, want %q", title, "New Title")
				}
				_, body, err := Split(result)
				if err != nil {
					t.Fatalf("Split: %v", err)
				}
				if body != "Body" {
					t.Errorf("body = %q, want %q", body, "Body")
				}
			},
		},
		{
			"no frontmatter creates frontmatter",
			"Just body text",
			"New Title",
			false,
			func(t *testing.T, result string) {
				t.Helper()
				title, err := GetTitle(result)
				if err != nil {
					t.Fatalf("GetTitle: %v", err)
				}
				if title != "New Title" {
					t.Errorf("title = %q, want %q", title, "New Title")
				}
				_, body, err := Split(result)
				if err != nil {
					t.Fatalf("Split: %v", err)
				}
				if body != "Just body text" {
					t.Errorf("body = %q, want %q", body, "Just body text")
				}
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := SetTitle(tt.input, tt.newTitle)

			if (err != nil) != tt.wantErr {
				t.Errorf("SetTitle() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if err != nil {
				return
			}
			if tt.check != nil {
				tt.check(t, result)
			}
		})
	}
}

func TestSetTitle_PreservesComments(t *testing.T) {
	input := "---\n# A comment\ntitle: Old\n---\n"
	result, err := SetTitle(input, "New")
	if err != nil {
		t.Fatalf("SetTitle() error = %v", err)
	}

	fm, _, err := Split(result)
	if err != nil {
		t.Fatalf("Split() error = %v", err)
	}
	if !strings.Contains(fm, "comment") {
		t.Error("comment was lost during round-trip")
	}
}

func TestGetTitle_RoundTrip(t *testing.T) {
	// Setting a title then getting it should return the same value.
	input := "---\ntitle: Original\n---\nBody"
	updated, err := SetTitle(input, "Round Trip Test")
	if err != nil {
		t.Fatalf("SetTitle() error = %v", err)
	}

	got, err := GetTitle(updated)
	if err != nil {
		t.Fatalf("GetTitle() error = %v", err)
	}
	if got != "Round Trip Test" {
		t.Errorf("round-trip title = %q, want %q", got, "Round Trip Test")
	}
}

func TestSplit_Reassembly(t *testing.T) {
	// Splitting then reassembling with Serialize should preserve the document.
	input := "---\ntitle: Hello\n---\nBody text"
	fm, body, err := Split(input)
	if err != nil {
		t.Fatalf("Split() error = %v", err)
	}

	reassembled := Serialize(fm, body)

	fm2, body2, err := Split(reassembled)
	if err != nil {
		t.Fatalf("Split reassembled: %v", err)
	}
	if fm2 != fm {
		t.Errorf("frontmatter changed: %q vs %q", fm2, fm)
	}
	if body2 != body {
		t.Errorf("body changed: %q vs %q", body2, body)
	}
}

func TestSerialize(t *testing.T) {
	tests := []struct {
		name string
		fm   string
		body string
		want string
	}{
		{
			"basic reassembly",
			"title: Hello\n",
			"Body text",
			"---\ntitle: Hello\n---\nBody text",
		},
		{
			"empty frontmatter",
			"",
			"Body text",
			"---\n---\nBody text",
		},
		{
			"empty body",
			"title: Hello\n",
			"",
			"---\ntitle: Hello\n---\n",
		},
		{
			"both empty",
			"",
			"",
			"---\n---\n",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := Serialize(tt.fm, tt.body)
			if got != tt.want {
				t.Errorf("Serialize() = %q, want %q", got, tt.want)
			}
		})
	}
}
