package domain

import (
	"sort"
	"testing"
)

func TestBuildOutline_SingleNode(t *testing.T) {
	files := []ParsedFile{
		{MP: "100", SID: "A3F7c9Qx7Lm2", DocType: "draft", Slug: "chapter-one", PathParts: []string{"100"}, Depth: 1},
		{MP: "100", SID: "A3F7c9Qx7Lm2", DocType: "notes", Slug: "", PathParts: []string{"100"}, Depth: 1},
	}

	outline, findings, err := BuildOutline(files)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(findings) != 0 {
		t.Errorf("expected no findings, got %d", len(findings))
	}
	if len(outline.Nodes) != 1 {
		t.Fatalf("expected 1 node, got %d", len(outline.Nodes))
	}

	node := outline.Nodes[0]
	if node.SID != "A3F7c9Qx7Lm2" {
		t.Errorf("SID = %q, want %q", node.SID, "A3F7c9Qx7Lm2")
	}
	if node.MP.String() != "100" {
		t.Errorf("MP = %q, want %q", node.MP.String(), "100")
	}
	if len(node.Documents) != 2 {
		t.Errorf("Documents length = %d, want 2", len(node.Documents))
	}
}

func TestBuildOutline_GroupsBySID(t *testing.T) {
	files := []ParsedFile{
		{MP: "100", SID: "AAAAAAAAAAAA", DocType: "draft", Slug: "first", PathParts: []string{"100"}, Depth: 1},
		{MP: "100", SID: "AAAAAAAAAAAA", DocType: "notes", Slug: "", PathParts: []string{"100"}, Depth: 1},
		{MP: "200", SID: "BBBBBBBBBBBB", DocType: "draft", Slug: "second", PathParts: []string{"200"}, Depth: 1},
		{MP: "200", SID: "BBBBBBBBBBBB", DocType: "notes", Slug: "", PathParts: []string{"200"}, Depth: 1},
	}

	outline, findings, err := BuildOutline(files)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(findings) != 0 {
		t.Errorf("expected no findings, got %d", len(findings))
	}
	if len(outline.Nodes) != 2 {
		t.Fatalf("expected 2 nodes, got %d", len(outline.Nodes))
	}

	if outline.Nodes[0].SID != "AAAAAAAAAAAA" {
		t.Errorf("Nodes[0].SID = %q, want %q", outline.Nodes[0].SID, "AAAAAAAAAAAA")
	}
	if outline.Nodes[1].SID != "BBBBBBBBBBBB" {
		t.Errorf("Nodes[1].SID = %q, want %q", outline.Nodes[1].SID, "BBBBBBBBBBBB")
	}
}

func TestBuildOutline_SortsByMP(t *testing.T) {
	files := []ParsedFile{
		{MP: "300", SID: "CCCCCCCCCCCC", DocType: "draft", Slug: "third", PathParts: []string{"300"}, Depth: 1},
		{MP: "300", SID: "CCCCCCCCCCCC", DocType: "notes", Slug: "", PathParts: []string{"300"}, Depth: 1},
		{MP: "100", SID: "AAAAAAAAAAAA", DocType: "draft", Slug: "first", PathParts: []string{"100"}, Depth: 1},
		{MP: "100", SID: "AAAAAAAAAAAA", DocType: "notes", Slug: "", PathParts: []string{"100"}, Depth: 1},
		{MP: "200", SID: "BBBBBBBBBBBB", DocType: "draft", Slug: "second", PathParts: []string{"200"}, Depth: 1},
		{MP: "200", SID: "BBBBBBBBBBBB", DocType: "notes", Slug: "", PathParts: []string{"200"}, Depth: 1},
	}

	outline, _, err := BuildOutline(files)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	want := []string{"100", "200", "300"}
	for i, node := range outline.Nodes {
		if node.MP.String() != want[i] {
			t.Errorf("Nodes[%d].MP = %q, want %q", i, node.MP.String(), want[i])
		}
	}
}

func TestBuildOutline_HierarchicalSortOrder(t *testing.T) {
	files := []ParsedFile{
		{MP: "200", SID: "SID200aaaaaa", DocType: "draft", Slug: "b", PathParts: []string{"200"}, Depth: 1},
		{MP: "200", SID: "SID200aaaaaa", DocType: "notes", Slug: "", PathParts: []string{"200"}, Depth: 1},
		{MP: "001-200", SID: "SID001200aaa", DocType: "draft", Slug: "a-child", PathParts: []string{"001", "200"}, Depth: 2},
		{MP: "001-200", SID: "SID001200aaa", DocType: "notes", Slug: "", PathParts: []string{"001", "200"}, Depth: 2},
		{MP: "001", SID: "SID001aaaaaa", DocType: "draft", Slug: "a", PathParts: []string{"001"}, Depth: 1},
		{MP: "001", SID: "SID001aaaaaa", DocType: "notes", Slug: "", PathParts: []string{"001"}, Depth: 1},
		{MP: "100", SID: "SID100aaaaaa", DocType: "draft", Slug: "c", PathParts: []string{"100"}, Depth: 1},
		{MP: "100", SID: "SID100aaaaaa", DocType: "notes", Slug: "", PathParts: []string{"100"}, Depth: 1},
		{MP: "001-100", SID: "SID001100aaa", DocType: "draft", Slug: "a-child2", PathParts: []string{"001", "100"}, Depth: 2},
		{MP: "001-100", SID: "SID001100aaa", DocType: "notes", Slug: "", PathParts: []string{"001", "100"}, Depth: 2},
	}

	outline, _, err := BuildOutline(files)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	want := []string{"001", "001-100", "001-200", "100", "200"}
	if len(outline.Nodes) != len(want) {
		t.Fatalf("expected %d nodes, got %d", len(want), len(outline.Nodes))
	}
	for i, node := range outline.Nodes {
		if node.MP.String() != want[i] {
			t.Errorf("Nodes[%d].MP = %q, want %q", i, node.MP.String(), want[i])
		}
	}
}

func TestBuildOutline_DocumentsGroupedByNode(t *testing.T) {
	files := []ParsedFile{
		{MP: "100", SID: "AAAAAAAAAAAA", DocType: "draft", Slug: "my-doc", PathParts: []string{"100"}, Depth: 1},
		{MP: "100", SID: "AAAAAAAAAAAA", DocType: "notes", Slug: "", PathParts: []string{"100"}, Depth: 1},
		{MP: "100", SID: "AAAAAAAAAAAA", DocType: "research", Slug: "", PathParts: []string{"100"}, Depth: 1},
	}

	outline, _, err := BuildOutline(files)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if len(outline.Nodes) != 1 {
		t.Fatalf("expected 1 node, got %d", len(outline.Nodes))
	}
	node := outline.Nodes[0]
	if len(node.Documents) != 3 {
		t.Fatalf("expected 3 documents, got %d", len(node.Documents))
	}

	docTypes := make([]string, len(node.Documents))
	for i, d := range node.Documents {
		docTypes[i] = d.Type
	}
	sort.Strings(docTypes)
	wantTypes := []string{"draft", "notes", "research"}
	for i, dt := range docTypes {
		if dt != wantTypes[i] {
			t.Errorf("docTypes[%d] = %q, want %q", i, dt, wantTypes[i])
		}
	}
}

func TestBuildOutline_DocumentFilenameGenerated(t *testing.T) {
	files := []ParsedFile{
		{MP: "100", SID: "AAAAAAAAAAAA", DocType: "draft", Slug: "my-doc", PathParts: []string{"100"}, Depth: 1},
	}

	outline, _, err := BuildOutline(files)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if len(outline.Nodes) != 1 {
		t.Fatalf("expected 1 node, got %d", len(outline.Nodes))
	}

	doc := outline.Nodes[0].Documents[0]
	wantFilename := "100_AAAAAAAAAAAA_draft_my-doc.md"
	if doc.Filename != wantFilename {
		t.Errorf("Document.Filename = %q, want %q", doc.Filename, wantFilename)
	}
	if doc.Type != "draft" {
		t.Errorf("Document.Type = %q, want %q", doc.Type, "draft")
	}
}

func TestBuildOutline_EmptyInput(t *testing.T) {
	outline, findings, err := BuildOutline(nil)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(findings) != 0 {
		t.Errorf("expected no findings, got %d", len(findings))
	}
	if len(outline.Nodes) != 0 {
		t.Errorf("expected 0 nodes, got %d", len(outline.Nodes))
	}
}

func TestBuildOutline_EmptySlice(t *testing.T) {
	outline, findings, err := BuildOutline([]ParsedFile{})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(findings) != 0 {
		t.Errorf("expected no findings, got %d", len(findings))
	}
	if len(outline.Nodes) != 0 {
		t.Errorf("expected 0 nodes, got %d", len(outline.Nodes))
	}
}

func TestBuildOutline_DuplicateSIDDifferentMPs(t *testing.T) {
	files := []ParsedFile{
		{MP: "100", SID: "AAAAAAAAAAAA", DocType: "draft", Slug: "first", PathParts: []string{"100"}, Depth: 1},
		{MP: "200", SID: "AAAAAAAAAAAA", DocType: "draft", Slug: "second", PathParts: []string{"200"}, Depth: 1},
	}

	_, findings, err := BuildOutline(files)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	found := false
	for _, f := range findings {
		if f.Type == FindingDuplicateSID {
			found = true
			break
		}
	}
	if !found {
		t.Error("expected FindingDuplicateSID finding, got none")
	}
}

func TestBuildOutline_Deterministic(t *testing.T) {
	files := []ParsedFile{
		{MP: "300", SID: "CCCCCCCCCCCC", DocType: "draft", Slug: "third", PathParts: []string{"300"}, Depth: 1},
		{MP: "300", SID: "CCCCCCCCCCCC", DocType: "notes", Slug: "", PathParts: []string{"300"}, Depth: 1},
		{MP: "100", SID: "AAAAAAAAAAAA", DocType: "notes", Slug: "", PathParts: []string{"100"}, Depth: 1},
		{MP: "100", SID: "AAAAAAAAAAAA", DocType: "draft", Slug: "first", PathParts: []string{"100"}, Depth: 1},
		{MP: "200", SID: "BBBBBBBBBBBB", DocType: "notes", Slug: "", PathParts: []string{"200"}, Depth: 1},
		{MP: "200", SID: "BBBBBBBBBBBB", DocType: "draft", Slug: "second", PathParts: []string{"200"}, Depth: 1},
	}

	outline1, _, err := BuildOutline(files)
	if err != nil {
		t.Fatalf("first call unexpected error: %v", err)
	}
	outline2, _, err := BuildOutline(files)
	if err != nil {
		t.Fatalf("second call unexpected error: %v", err)
	}

	if len(outline1.Nodes) != len(outline2.Nodes) {
		t.Fatalf("node count mismatch: %d vs %d", len(outline1.Nodes), len(outline2.Nodes))
	}
	for i := range outline1.Nodes {
		if outline1.Nodes[i].SID != outline2.Nodes[i].SID {
			t.Errorf("Nodes[%d].SID mismatch: %q vs %q", i, outline1.Nodes[i].SID, outline2.Nodes[i].SID)
		}
		if outline1.Nodes[i].MP.String() != outline2.Nodes[i].MP.String() {
			t.Errorf("Nodes[%d].MP mismatch: %q vs %q", i, outline1.Nodes[i].MP.String(), outline2.Nodes[i].MP.String())
		}
	}
}

func TestBuildOutline_ParentChildRelationship(t *testing.T) {
	files := []ParsedFile{
		{MP: "100", SID: "ParentSID001a", DocType: "draft", Slug: "parent", PathParts: []string{"100"}, Depth: 1},
		{MP: "100", SID: "ParentSID001a", DocType: "notes", Slug: "", PathParts: []string{"100"}, Depth: 1},
		{MP: "100-100", SID: "ChildSID001aa", DocType: "draft", Slug: "child", PathParts: []string{"100", "100"}, Depth: 2},
		{MP: "100-100", SID: "ChildSID001aa", DocType: "notes", Slug: "", PathParts: []string{"100", "100"}, Depth: 2},
		{MP: "100-100-100", SID: "GrandChild01a", DocType: "draft", Slug: "grandchild", PathParts: []string{"100", "100", "100"}, Depth: 3},
		{MP: "100-100-100", SID: "GrandChild01a", DocType: "notes", Slug: "", PathParts: []string{"100", "100", "100"}, Depth: 3},
	}

	outline, _, err := BuildOutline(files)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if len(outline.Nodes) != 3 {
		t.Fatalf("expected 3 nodes, got %d", len(outline.Nodes))
	}

	// Verify parent-child via MP prefix relationship
	parent := outline.Nodes[0]
	child := outline.Nodes[1]
	grandchild := outline.Nodes[2]

	if !parent.MP.IsAncestorOf(child.MP) {
		t.Errorf("expected %q to be ancestor of %q", parent.MP.String(), child.MP.String())
	}
	if !parent.MP.IsAncestorOf(grandchild.MP) {
		t.Errorf("expected %q to be ancestor of %q", parent.MP.String(), grandchild.MP.String())
	}
	if !child.MP.IsAncestorOf(grandchild.MP) {
		t.Errorf("expected %q to be ancestor of %q", child.MP.String(), grandchild.MP.String())
	}
}

func TestBuildOutline_SlugUsedAsTitle(t *testing.T) {
	files := []ParsedFile{
		{MP: "100", SID: "AAAAAAAAAAAA", DocType: "draft", Slug: "my-great-novel", PathParts: []string{"100"}, Depth: 1},
		{MP: "100", SID: "AAAAAAAAAAAA", DocType: "notes", Slug: "", PathParts: []string{"100"}, Depth: 1},
	}

	outline, _, err := BuildOutline(files)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if len(outline.Nodes) != 1 {
		t.Fatalf("expected 1 node, got %d", len(outline.Nodes))
	}

	// The draft slug should be used to derive the node's title
	node := outline.Nodes[0]
	if node.Title != "my-great-novel" {
		t.Errorf("Title = %q, want %q", node.Title, "my-great-novel")
	}
}

func TestBuildOutline_NoSlugEmptyTitle(t *testing.T) {
	files := []ParsedFile{
		{MP: "100", SID: "AAAAAAAAAAAA", DocType: "draft", Slug: "", PathParts: []string{"100"}, Depth: 1},
		{MP: "100", SID: "AAAAAAAAAAAA", DocType: "notes", Slug: "", PathParts: []string{"100"}, Depth: 1},
	}

	outline, _, err := BuildOutline(files)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if len(outline.Nodes) != 1 {
		t.Fatalf("expected 1 node, got %d", len(outline.Nodes))
	}

	node := outline.Nodes[0]
	if node.Title != "" {
		t.Errorf("Title = %q, want empty string", node.Title)
	}
}

func TestBuildOutline_DocumentsSortedByType(t *testing.T) {
	files := []ParsedFile{
		{MP: "100", SID: "AAAAAAAAAAAA", DocType: "notes", Slug: "", PathParts: []string{"100"}, Depth: 1},
		{MP: "100", SID: "AAAAAAAAAAAA", DocType: "research", Slug: "", PathParts: []string{"100"}, Depth: 1},
		{MP: "100", SID: "AAAAAAAAAAAA", DocType: "draft", Slug: "test", PathParts: []string{"100"}, Depth: 1},
	}

	outline, _, err := BuildOutline(files)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if len(outline.Nodes) != 1 {
		t.Fatalf("expected 1 node, got %d", len(outline.Nodes))
	}

	docs := outline.Nodes[0].Documents
	if len(docs) != 3 {
		t.Fatalf("expected 3 documents, got %d", len(docs))
	}

	// Documents should be sorted by type for deterministic output
	for i := 1; i < len(docs); i++ {
		if docs[i-1].Type > docs[i].Type {
			t.Errorf("documents not sorted: %q comes before %q", docs[i-1].Type, docs[i].Type)
		}
	}
}

func TestBuildOutline_InvalidMP(t *testing.T) {
	files := []ParsedFile{
		{MP: "000", SID: "AAAAAAAAAAAA", DocType: "draft", Slug: "bad", PathParts: []string{"000"}, Depth: 1},
	}

	_, _, err := BuildOutline(files)
	if err == nil {
		t.Error("expected error for invalid MP, got nil")
	}
}

func TestOutline_NodesFieldInitialized(t *testing.T) {
	outline, _, err := BuildOutline(nil)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	// Nodes should be an empty slice, not nil
	if outline.Nodes == nil {
		t.Error("expected Nodes to be non-nil empty slice")
	}
}
