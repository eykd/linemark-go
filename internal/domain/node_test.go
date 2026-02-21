package domain

import (
	"sort"
	"testing"
)

func TestNewMaterializedPath_ValidPaths(t *testing.T) {
	tests := []struct {
		name  string
		input string
	}{
		{"single segment root", "001"},
		{"single segment high", "999"},
		{"two segments", "001-200"},
		{"three segments", "001-200-010"},
		{"four segments", "100-200-300-400"},
		{"minimum non-zero", "001-001-001"},
		{"mixed values", "100-050-999"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mp, err := NewMaterializedPath(tt.input)
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if mp.String() != tt.input {
				t.Errorf("String() = %q, want %q", mp.String(), tt.input)
			}
		})
	}
}

func TestNewMaterializedPath_InvalidPaths(t *testing.T) {
	tests := []struct {
		name  string
		input string
	}{
		{"empty string", ""},
		{"zero segment", "000"},
		{"zero in second segment", "001-000"},
		{"zero in third segment", "001-200-000"},
		{"segment too short", "01"},
		{"segment too long", "0001"},
		{"letters in segment", "abc"},
		{"mixed invalid segments", "001-ab-200"},
		{"trailing dash", "001-"},
		{"leading dash", "-001"},
		{"double dash", "001--200"},
		{"whitespace", " 001"},
		{"segment with spaces", "001 200"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := NewMaterializedPath(tt.input)
			if err == nil {
				t.Error("expected error, got nil")
			}
		})
	}
}

func TestMaterializedPath_Depth(t *testing.T) {
	tests := []struct {
		name  string
		input string
		want  int
	}{
		{"single segment", "001", 1},
		{"two segments", "001-200", 2},
		{"three segments", "001-200-010", 3},
		{"four segments", "100-200-300-400", 4},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mp, err := NewMaterializedPath(tt.input)
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if got := mp.Depth(); got != tt.want {
				t.Errorf("Depth() = %d, want %d", got, tt.want)
			}
		})
	}
}

func TestMaterializedPath_Parent(t *testing.T) {
	tests := []struct {
		name       string
		input      string
		wantParent string
		wantEmpty  bool
	}{
		{"root has no parent", "001", "", true},
		{"two segments returns root", "001-200", "001", false},
		{"three segments returns two", "001-200-010", "001-200", false},
		{"four segments returns three", "100-200-300-400", "100-200-300", false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mp, err := NewMaterializedPath(tt.input)
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			parent, hasParent := mp.Parent()
			if hasParent == tt.wantEmpty {
				t.Errorf("hasParent = %v, want %v", hasParent, !tt.wantEmpty)
			}
			if !tt.wantEmpty {
				if parent.String() != tt.wantParent {
					t.Errorf("Parent() = %q, want %q", parent.String(), tt.wantParent)
				}
			}
		})
	}
}

func TestMaterializedPath_Segments(t *testing.T) {
	tests := []struct {
		name  string
		input string
		want  []string
	}{
		{"single segment", "001", []string{"001"}},
		{"two segments", "001-200", []string{"001", "200"}},
		{"three segments", "001-200-010", []string{"001", "200", "010"}},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mp, err := NewMaterializedPath(tt.input)
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			got := mp.Segments()
			if len(got) != len(tt.want) {
				t.Fatalf("Segments() length = %d, want %d", len(got), len(tt.want))
			}
			for i, seg := range got {
				if seg != tt.want[i] {
					t.Errorf("Segments()[%d] = %q, want %q", i, seg, tt.want[i])
				}
			}
		})
	}
}

func TestMaterializedPath_SortOrder(t *testing.T) {
	input := []string{"200", "001-200", "001", "100-200-300", "001-100", "100"}
	want := []string{"001", "001-100", "001-200", "100", "100-200-300", "200"}

	paths := make([]MaterializedPath, len(input))
	for i, s := range input {
		mp, err := NewMaterializedPath(s)
		if err != nil {
			t.Fatalf("unexpected error for %q: %v", s, err)
		}
		paths[i] = mp
	}

	sort.Slice(paths, func(i, j int) bool {
		return paths[i].String() < paths[j].String()
	})

	for i, mp := range paths {
		if mp.String() != want[i] {
			t.Errorf("sorted[%d] = %q, want %q", i, mp.String(), want[i])
		}
	}
}

func TestDocument_Fields(t *testing.T) {
	doc := Document{
		Type:     "draft",
		Filename: "001_A3F7c9Qx7Lm2_draft_my-novel.md",
		Content:  "# Chapter One\n\nIt was a dark and stormy night.",
	}

	if doc.Type != "draft" {
		t.Errorf("Type = %q, want %q", doc.Type, "draft")
	}
	if doc.Filename != "001_A3F7c9Qx7Lm2_draft_my-novel.md" {
		t.Errorf("Filename = %q, want %q", doc.Filename, "001_A3F7c9Qx7Lm2_draft_my-novel.md")
	}
	if doc.Content != "# Chapter One\n\nIt was a dark and stormy night." {
		t.Errorf("Content = %q, want %q", doc.Content, "# Chapter One\n\nIt was a dark and stormy night.")
	}
}

func TestNode_Fields(t *testing.T) {
	mp, err := NewMaterializedPath("001-200")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	node := Node{
		MP:    mp,
		SID:   "A3F7c9Qx7Lm2",
		Title: "Chapter One",
		Documents: []Document{
			{Type: "draft", Filename: "001-200_A3F7c9Qx7Lm2_draft_chapter-one.md", Content: ""},
			{Type: "notes", Filename: "001-200_A3F7c9Qx7Lm2_notes.md", Content: ""},
		},
	}

	if node.MP.String() != "001-200" {
		t.Errorf("MP = %q, want %q", node.MP.String(), "001-200")
	}
	if node.SID != "A3F7c9Qx7Lm2" {
		t.Errorf("SID = %q, want %q", node.SID, "A3F7c9Qx7Lm2")
	}
	if node.Title != "Chapter One" {
		t.Errorf("Title = %q, want %q", node.Title, "Chapter One")
	}
	if len(node.Documents) != 2 {
		t.Fatalf("Documents length = %d, want 2", len(node.Documents))
	}
	if node.Documents[0].Type != "draft" {
		t.Errorf("Documents[0].Type = %q, want %q", node.Documents[0].Type, "draft")
	}
	if node.Documents[1].Type != "notes" {
		t.Errorf("Documents[1].Type = %q, want %q", node.Documents[1].Type, "notes")
	}
}

func TestNode_SortByMaterializedPath(t *testing.T) {
	paths := []string{"200", "001-200", "001", "100"}
	want := []string{"001", "001-200", "100", "200"}

	nodes := make([]Node, len(paths))
	for i, p := range paths {
		mp, err := NewMaterializedPath(p)
		if err != nil {
			t.Fatalf("unexpected error for %q: %v", p, err)
		}
		nodes[i] = Node{MP: mp, SID: "sid" + p, Title: "Title " + p}
	}

	sort.Slice(nodes, func(i, j int) bool {
		return nodes[i].MP.String() < nodes[j].MP.String()
	})

	for i, node := range nodes {
		if node.MP.String() != want[i] {
			t.Errorf("sorted[%d].MP = %q, want %q", i, node.MP.String(), want[i])
		}
	}
}
