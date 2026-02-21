package domain

import "testing"

func TestValidateDocType_RejectsPathTraversal(t *testing.T) {
	tests := []struct {
		name    string
		docType string
	}{
		{"path traversal with slashes", "../../../etc"},
		{"forward slash", "foo/bar"},
		{"backslash", `foo\bar`},
		{"null byte", "notes\x00"},
		{"dot-dot", ".."},
		{"single dot", "."},
		{"dash in doc type", "my-type"},
		{"uppercase letters", "Draft"},
		{"digits in doc type", "type1"},
		{"space in doc type", "my type"},
		{"empty string", ""},
		{"underscore in doc type", "my_type"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := ValidateDocType(tt.docType)
			if err == nil {
				t.Errorf("ValidateDocType(%q) = nil, want error", tt.docType)
			}
		})
	}
}

func TestValidateDocType_AcceptsValid(t *testing.T) {
	tests := []struct {
		name    string
		docType string
	}{
		{"draft", "draft"},
		{"notes", "notes"},
		{"characters", "characters"},
		{"locations", "locations"},
		{"outline", "outline"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := ValidateDocType(tt.docType)
			if err != nil {
				t.Errorf("ValidateDocType(%q) = %v, want nil", tt.docType, err)
			}
		})
	}
}

func TestParseFilename_RejectsSlugWithPathSeparators(t *testing.T) {
	tests := []struct {
		name     string
		filename string
	}{
		{
			"slug with forward slash",
			"001_A3F7c9Qx7Lm2_draft_foo/bar.md",
		},
		{
			"slug with path traversal",
			"001_A3F7c9Qx7Lm2_draft_../../../etc/passwd.md",
		},
		{
			"slug with backslash",
			"001_A3F7c9Qx7Lm2_draft_foo\\bar.md",
		},
		{
			"slug with null byte",
			"001_A3F7c9Qx7Lm2_draft_foo\x00bar.md",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := ParseFilename(tt.filename)
			if err == nil {
				t.Errorf("ParseFilename(%q) should reject filenames with path separators in slug", tt.filename)
			}
		})
	}
}
