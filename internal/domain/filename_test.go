package domain

import (
	"testing"
)

func TestParseFilename_ValidFilenames(t *testing.T) {
	tests := []struct {
		name     string
		filename string
		want     ParsedFile
	}{
		{
			name:     "root node with slug",
			filename: "001_A3F7c9Qx7Lm2_draft_my-novel.md",
			want: ParsedFile{
				MP:        "001",
				SID:       "A3F7c9Qx7Lm2",
				DocType:   "draft",
				Slug:      "my-novel",
				PathParts: []string{"001"},
				Depth:     1,
			},
		},
		{
			name:     "root node without slug",
			filename: "001_A3F7c9Qx7Lm2_notes.md",
			want: ParsedFile{
				MP:        "001",
				SID:       "A3F7c9Qx7Lm2",
				DocType:   "notes",
				Slug:      "",
				PathParts: []string{"001"},
				Depth:     1,
			},
		},
		{
			name:     "nested node with slug",
			filename: "001-200-010_B8kQ2mNp4Rs1_draft_chapter-one.md",
			want: ParsedFile{
				MP:        "001-200-010",
				SID:       "B8kQ2mNp4Rs1",
				DocType:   "draft",
				Slug:      "chapter-one",
				PathParts: []string{"001", "200", "010"},
				Depth:     3,
			},
		},
		{
			name:     "nested node without slug",
			filename: "001-200-010_B8kQ2mNp4Rs1_characters.md",
			want: ParsedFile{
				MP:        "001-200-010",
				SID:       "B8kQ2mNp4Rs1",
				DocType:   "characters",
				Slug:      "",
				PathParts: []string{"001", "200", "010"},
				Depth:     3,
			},
		},
		{
			name:     "two-segment path",
			filename: "100-200_abcdefghijkl_draft_hello-world.md",
			want: ParsedFile{
				MP:        "100-200",
				SID:       "abcdefghijkl",
				DocType:   "draft",
				Slug:      "hello-world",
				PathParts: []string{"100", "200"},
				Depth:     2,
			},
		},
		{
			name:     "8 character SID minimum",
			filename: "001_AbCd1234_draft_short-sid.md",
			want: ParsedFile{
				MP:        "001",
				SID:       "AbCd1234",
				DocType:   "draft",
				Slug:      "short-sid",
				PathParts: []string{"001"},
				Depth:     1,
			},
		},
		{
			name:     "slug with multiple hyphens",
			filename: "001_A3F7c9Qx7Lm2_draft_my-long-chapter-title.md",
			want: ParsedFile{
				MP:        "001",
				SID:       "A3F7c9Qx7Lm2",
				DocType:   "draft",
				Slug:      "my-long-chapter-title",
				PathParts: []string{"001"},
				Depth:     1,
			},
		},
		{
			name:     "slug with numbers",
			filename: "001_A3F7c9Qx7Lm2_draft_chapter-42.md",
			want: ParsedFile{
				MP:        "001",
				SID:       "A3F7c9Qx7Lm2",
				DocType:   "draft",
				Slug:      "chapter-42",
				PathParts: []string{"001"},
				Depth:     1,
			},
		},
		{
			name:     "custom doc type",
			filename: "001_A3F7c9Qx7Lm2_locations_the-shire.md",
			want: ParsedFile{
				MP:        "001",
				SID:       "A3F7c9Qx7Lm2",
				DocType:   "locations",
				Slug:      "the-shire",
				PathParts: []string{"001"},
				Depth:     1,
			},
		},
		{
			name:     "slug containing underscores after first three parts",
			filename: "001_A3F7c9Qx7Lm2_draft_slug_with_underscores.md",
			want: ParsedFile{
				MP:        "001",
				SID:       "A3F7c9Qx7Lm2",
				DocType:   "draft",
				Slug:      "slug_with_underscores",
				PathParts: []string{"001"},
				Depth:     1,
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := ParseFilename(tt.filename)
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if got.MP != tt.want.MP {
				t.Errorf("MP = %q, want %q", got.MP, tt.want.MP)
			}
			if got.SID != tt.want.SID {
				t.Errorf("SID = %q, want %q", got.SID, tt.want.SID)
			}
			if got.DocType != tt.want.DocType {
				t.Errorf("DocType = %q, want %q", got.DocType, tt.want.DocType)
			}
			if got.Slug != tt.want.Slug {
				t.Errorf("Slug = %q, want %q", got.Slug, tt.want.Slug)
			}
			if len(got.PathParts) != len(tt.want.PathParts) {
				t.Fatalf("PathParts length = %d, want %d", len(got.PathParts), len(tt.want.PathParts))
			}
			for i, part := range got.PathParts {
				if part != tt.want.PathParts[i] {
					t.Errorf("PathParts[%d] = %q, want %q", i, part, tt.want.PathParts[i])
				}
			}
			if got.Depth != tt.want.Depth {
				t.Errorf("Depth = %d, want %d", got.Depth, tt.want.Depth)
			}
		})
	}
}

func TestParseFilename_InvalidFilenames(t *testing.T) {
	tests := []struct {
		name     string
		filename string
	}{
		{"empty string", ""},
		{"no extension", "001_A3F7c9Qx7Lm2_draft_slug"},
		{"wrong extension", "001_A3F7c9Qx7Lm2_draft_slug.txt"},
		{"missing mp", "_A3F7c9Qx7Lm2_draft_slug.md"},
		{"mp with leading zero segment 000", "000_A3F7c9Qx7Lm2_draft.md"},
		{"mp segment too short", "01_A3F7c9Qx7Lm2_draft.md"},
		{"mp segment too long", "0001_A3F7c9Qx7Lm2_draft.md"},
		{"sid too short 7 chars", "001_AbCd123_draft.md"},
		{"sid too long 13 chars", "001_AbCd1234567890_draft.md"},
		{"sid with special chars", "001_AbCd!@#$_draft.md"},
		{"missing doc type", "001_A3F7c9Qx7Lm2.md"},
		{"doc type with uppercase", "001_A3F7c9Qx7Lm2_Draft.md"},
		{"doc type with numbers", "001_A3F7c9Qx7Lm2_draft1.md"},
		{"plain text file", "readme.md"},
		{"just underscores", "_____.md"},
		{"path separator in name", "foo/001_A3F7c9Qx7Lm2_draft.md"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := ParseFilename(tt.filename)
			if err == nil {
				t.Error("expected error, got nil")
			}
		})
	}
}

func TestGenerateFilename_WithSlug(t *testing.T) {
	tests := []struct {
		name    string
		mp      string
		sid     string
		docType string
		slug    string
		want    string
	}{
		{
			name:    "root draft with slug",
			mp:      "001",
			sid:     "A3F7c9Qx7Lm2",
			docType: "draft",
			slug:    "my-novel",
			want:    "001_A3F7c9Qx7Lm2_draft_my-novel.md",
		},
		{
			name:    "nested with slug",
			mp:      "001-200-010",
			sid:     "B8kQ2mNp4Rs1",
			docType: "draft",
			slug:    "chapter-one",
			want:    "001-200-010_B8kQ2mNp4Rs1_draft_chapter-one.md",
		},
		{
			name:    "custom doc type with slug",
			mp:      "100",
			sid:     "xyzXYZ123456",
			docType: "characters",
			slug:    "main-cast",
			want:    "100_xyzXYZ123456_characters_main-cast.md",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := GenerateFilename(tt.mp, tt.sid, tt.docType, tt.slug)
			if got != tt.want {
				t.Errorf("got %q, want %q", got, tt.want)
			}
		})
	}
}

func TestGenerateFilename_EmptySlug(t *testing.T) {
	tests := []struct {
		name    string
		mp      string
		sid     string
		docType string
		want    string
	}{
		{
			name:    "notes without slug",
			mp:      "001",
			sid:     "A3F7c9Qx7Lm2",
			docType: "notes",
			want:    "001_A3F7c9Qx7Lm2_notes.md",
		},
		{
			name:    "draft without slug",
			mp:      "200-300",
			sid:     "abcdefghijkl",
			docType: "draft",
			want:    "200-300_abcdefghijkl_draft.md",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := GenerateFilename(tt.mp, tt.sid, tt.docType, "")
			if got != tt.want {
				t.Errorf("got %q, want %q", got, tt.want)
			}
		})
	}
}

func TestParseFilename_RoundTrip(t *testing.T) {
	tests := []struct {
		name     string
		filename string
	}{
		{"with slug", "001_A3F7c9Qx7Lm2_draft_my-novel.md"},
		{"without slug", "001_A3F7c9Qx7Lm2_notes.md"},
		{"nested with slug", "001-200-010_B8kQ2mNp4Rs1_draft_chapter-one.md"},
		{"nested without slug", "100-200_abcdefghijkl_characters.md"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			parsed, err := ParseFilename(tt.filename)
			if err != nil {
				t.Fatalf("parse failed: %v", err)
			}
			regenerated := GenerateFilename(parsed.MP, parsed.SID, parsed.DocType, parsed.Slug)
			if regenerated != tt.filename {
				t.Errorf("round-trip failed: got %q, want %q", regenerated, tt.filename)
			}
		})
	}
}
