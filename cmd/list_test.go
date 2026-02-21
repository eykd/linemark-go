package cmd

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"strings"
	"testing"

	"github.com/eykd/linemark-go/internal/domain"
	"github.com/spf13/cobra"
)

// mockListRunner is a test double for ListRunner.
type mockListRunner struct {
	result *ListResult
	err    error
}

func (m *mockListRunner) List(ctx context.Context) (*ListResult, error) {
	return m.result, m.err
}

// newTestListCmd creates a list command wired to the given runner,
// capturing stdout into the returned buffer.
func newTestListCmd(runner *mockListRunner, args ...string) (*cobra.Command, *bytes.Buffer) {
	cmd := NewListCmd(runner)
	if len(args) > 0 {
		cmd.SetArgs(args)
	}
	buf := new(bytes.Buffer)
	cmd.SetOut(buf)
	cmd.SetErr(new(bytes.Buffer))
	return cmd, buf
}

// newTestRootListCmd creates a list command wired through root (for global flags like --json),
// capturing stdout into the returned buffer.
func newTestRootListCmd(runner *mockListRunner, args ...string) (*cobra.Command, *bytes.Buffer) {
	root := NewRootCmd()
	cmd := NewListCmd(runner)
	root.AddCommand(cmd)
	buf := new(bytes.Buffer)
	root.SetOut(buf)
	root.SetErr(new(bytes.Buffer))
	if len(args) > 0 {
		root.SetArgs(args)
	}
	return root, buf
}

// helper to build domain nodes for test fixtures.
func mustMP(s string) domain.MaterializedPath {
	mp, err := domain.NewMaterializedPath(s)
	if err != nil {
		panic(fmt.Sprintf("invalid MP %q: %v", s, err))
	}
	return mp
}

// threeNodeOutline returns a flat outline with root, child, and grandchild.
func threeNodeOutline() *ListResult {
	return &ListResult{
		Outline: domain.Outline{
			Nodes: []domain.Node{
				{
					MP:    mustMP("001"),
					SID:   "A3F7c9Qx7Lm2",
					Title: "Overview",
					Documents: []domain.Document{
						{Type: "draft", Filename: "001_A3F7c9Qx7Lm2_draft_overview.md"},
						{Type: "notes", Filename: "001_A3F7c9Qx7Lm2_notes.md"},
					},
				},
				{
					MP:    mustMP("001-100"),
					SID:   "B8kQ2mNp4Rs1",
					Title: "Part One",
					Documents: []domain.Document{
						{Type: "draft", Filename: "001-100_B8kQ2mNp4Rs1_draft_part-one.md"},
						{Type: "notes", Filename: "001-100_B8kQ2mNp4Rs1_notes.md"},
					},
				},
				{
					MP:    mustMP("001-100-200"),
					SID:   "C2xL9pQr5Tm3",
					Title: "Chapter 1",
					Documents: []domain.Document{
						{Type: "draft", Filename: "001-100-200_C2xL9pQr5Tm3_draft_chapter-1.md"},
					},
				},
			},
		},
	}
}

// twoRootOutline returns a flat outline with two root nodes and children.
func twoRootOutline() *ListResult {
	return &ListResult{
		Outline: domain.Outline{
			Nodes: []domain.Node{
				{
					MP:    mustMP("001"),
					SID:   "A3F7c9Qx7Lm2",
					Title: "Overview",
					Documents: []domain.Document{
						{Type: "draft", Filename: "001_A3F7c9Qx7Lm2_draft_overview.md"},
						{Type: "notes", Filename: "001_A3F7c9Qx7Lm2_notes.md"},
					},
				},
				{
					MP:    mustMP("001-100"),
					SID:   "B8kQ2mNp4Rs1",
					Title: "Part One",
					Documents: []domain.Document{
						{Type: "draft", Filename: "001-100_B8kQ2mNp4Rs1_draft_part-one.md"},
						{Type: "notes", Filename: "001-100_B8kQ2mNp4Rs1_notes.md"},
					},
				},
				{
					MP:    mustMP("001-100-200"),
					SID:   "C2xL9pQr5Tm3",
					Title: "Chapter 1",
					Documents: []domain.Document{
						{Type: "draft", Filename: "001-100-200_C2xL9pQr5Tm3_draft_chapter-1.md"},
					},
				},
				{
					MP:    mustMP("001-100-300"),
					SID:   "D4yM0rSt6Un4",
					Title: "Chapter 2",
					Documents: []domain.Document{
						{Type: "draft", Filename: "001-100-300_D4yM0rSt6Un4_draft_chapter-2.md"},
					},
				},
				{
					MP:    mustMP("002"),
					SID:   "E6zN1sUv7Wo5",
					Title: "Part Two",
					Documents: []domain.Document{
						{Type: "draft", Filename: "002_E6zN1sUv7Wo5_draft_part-two.md"},
						{Type: "notes", Filename: "002_E6zN1sUv7Wo5_notes.md"},
					},
				},
			},
		},
	}
}

func TestListCmd_RegisteredWithRoot(t *testing.T) {
	found := false
	for _, sub := range rootCmd.Commands() {
		if sub.Name() == "list" {
			found = true
			break
		}
	}
	if !found {
		t.Error("list command not registered with root")
	}
}

func TestListCmd_EmptyOutline(t *testing.T) {
	runner := &mockListRunner{
		result: &ListResult{
			Outline: domain.Outline{Nodes: []domain.Node{}},
		},
	}
	cmd, buf := newTestListCmd(runner)

	err := cmd.Execute()

	if err != nil {
		t.Fatalf("expected no error for empty outline, got %v", err)
	}
	if buf.String() != "" {
		t.Errorf("expected no output for empty outline, got %q", buf.String())
	}
}

func TestListCmd_EmptyOutline_JSON(t *testing.T) {
	runner := &mockListRunner{
		result: &ListResult{
			Outline: domain.Outline{Nodes: []domain.Node{}},
		},
	}
	cmd, buf := newTestListCmd(runner, "--json")

	err := cmd.Execute()

	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	var output struct {
		Nodes []interface{} `json:"nodes"`
	}
	if jsonErr := json.Unmarshal(buf.Bytes(), &output); jsonErr != nil {
		t.Fatalf("invalid JSON output: %v\nraw: %s", jsonErr, buf.String())
	}
	if len(output.Nodes) != 0 {
		t.Errorf("expected 0 nodes, got %d", len(output.Nodes))
	}
}

func TestListCmd_TreeDisplay(t *testing.T) {
	runner := &mockListRunner{result: twoRootOutline()}
	cmd, buf := newTestListCmd(runner)

	err := cmd.Execute()

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	output := buf.String()

	// Root node has no prefix
	if !strings.Contains(output, "Overview (A3F7c9Qx7Lm2)") {
		t.Errorf("output should contain root node with title and SID, got:\n%s", output)
	}

	// Box-drawing characters for tree structure
	if !strings.Contains(output, "├──") {
		t.Errorf("output should contain box-drawing tee character, got:\n%s", output)
	}
	if !strings.Contains(output, "└──") {
		t.Errorf("output should contain box-drawing corner character, got:\n%s", output)
	}
	if !strings.Contains(output, "│") {
		t.Errorf("output should contain box-drawing vertical bar, got:\n%s", output)
	}

	// Children are indented under parent
	if !strings.Contains(output, "Part One (B8kQ2mNp4Rs1)") {
		t.Errorf("output should contain child node, got:\n%s", output)
	}
	if !strings.Contains(output, "Chapter 1 (C2xL9pQr5Tm3)") {
		t.Errorf("output should contain grandchild node, got:\n%s", output)
	}
	if !strings.Contains(output, "Chapter 2 (D4yM0rSt6Un4)") {
		t.Errorf("output should contain grandchild node, got:\n%s", output)
	}
	if !strings.Contains(output, "Part Two (E6zN1sUv7Wo5)") {
		t.Errorf("output should contain second root node, got:\n%s", output)
	}

	// Output should not be valid JSON
	var parsed map[string]interface{}
	if json.Unmarshal(buf.Bytes(), &parsed) == nil {
		t.Errorf("output should not be valid JSON without --json flag, got: %s", output)
	}
}

func TestListCmd_TreeDisplay_ExactFormat(t *testing.T) {
	runner := &mockListRunner{result: twoRootOutline()}
	cmd, buf := newTestListCmd(runner)

	err := cmd.Execute()

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	// Verify the exact tree format matches the contract specification
	want := strings.Join([]string{
		"Overview (A3F7c9Qx7Lm2)",
		"├── Part One (B8kQ2mNp4Rs1)",
		"│   ├── Chapter 1 (C2xL9pQr5Tm3)",
		"│   └── Chapter 2 (D4yM0rSt6Un4)",
		"└── Part Two (E6zN1sUv7Wo5)",
		"",
	}, "\n")

	got := buf.String()
	if got != want {
		t.Errorf("tree output mismatch.\nwant:\n%s\ngot:\n%s", want, got)
	}
}

func TestListCmd_SingleRootNoChildren(t *testing.T) {
	runner := &mockListRunner{
		result: &ListResult{
			Outline: domain.Outline{
				Nodes: []domain.Node{
					{
						MP:    mustMP("001"),
						SID:   "A3F7c9Qx7Lm2",
						Title: "Overview",
						Documents: []domain.Document{
							{Type: "draft", Filename: "001_A3F7c9Qx7Lm2_draft_overview.md"},
						},
					},
				},
			},
		},
	}
	cmd, buf := newTestListCmd(runner)

	err := cmd.Execute()

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	want := "Overview (A3F7c9Qx7Lm2)\n"
	got := buf.String()
	if got != want {
		t.Errorf("output = %q, want %q", got, want)
	}
}

func TestListCmd_DepthFlag(t *testing.T) {
	tests := []struct {
		name        string
		depth       string
		wantPresent []string
		wantAbsent  []string
	}{
		{
			name:        "depth 1 shows roots only",
			depth:       "1",
			wantPresent: []string{"Overview (A3F7c9Qx7Lm2)"},
			wantAbsent:  []string{"Part One", "Chapter 1"},
		},
		{
			name:        "depth 2 shows roots and children",
			depth:       "2",
			wantPresent: []string{"Overview (A3F7c9Qx7Lm2)", "Part One (B8kQ2mNp4Rs1)"},
			wantAbsent:  []string{"Chapter 1"},
		},
		{
			name:        "depth 0 means unlimited",
			depth:       "0",
			wantPresent: []string{"Overview (A3F7c9Qx7Lm2)", "Part One (B8kQ2mNp4Rs1)", "Chapter 1 (C2xL9pQr5Tm3)"},
			wantAbsent:  []string{},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			runner := &mockListRunner{result: threeNodeOutline()}
			cmd, buf := newTestListCmd(runner, "--depth", tt.depth)

			err := cmd.Execute()

			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}

			output := buf.String()
			for _, s := range tt.wantPresent {
				if !strings.Contains(output, s) {
					t.Errorf("output should contain %q, got:\n%s", s, output)
				}
			}
			for _, s := range tt.wantAbsent {
				if strings.Contains(output, s) {
					t.Errorf("output should NOT contain %q, got:\n%s", s, output)
				}
			}
		})
	}
}

func TestListCmd_TypeFilter(t *testing.T) {
	// Outline where only some nodes have "characters" doc type
	runner := &mockListRunner{
		result: &ListResult{
			Outline: domain.Outline{
				Nodes: []domain.Node{
					{
						MP:    mustMP("001"),
						SID:   "A3F7c9Qx7Lm2",
						Title: "Overview",
						Documents: []domain.Document{
							{Type: "draft", Filename: "001_A3F7c9Qx7Lm2_draft_overview.md"},
							{Type: "characters", Filename: "001_A3F7c9Qx7Lm2_characters_overview.md"},
						},
					},
					{
						MP:    mustMP("002"),
						SID:   "B8kQ2mNp4Rs1",
						Title: "Part One",
						Documents: []domain.Document{
							{Type: "draft", Filename: "002_B8kQ2mNp4Rs1_draft_part-one.md"},
						},
					},
					{
						MP:    mustMP("003"),
						SID:   "C2xL9pQr5Tm3",
						Title: "Part Two",
						Documents: []domain.Document{
							{Type: "draft", Filename: "003_C2xL9pQr5Tm3_draft_part-two.md"},
							{Type: "characters", Filename: "003_C2xL9pQr5Tm3_characters_part-two.md"},
						},
					},
				},
			},
		},
	}
	cmd, buf := newTestListCmd(runner, "--type", "characters")

	err := cmd.Execute()

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	output := buf.String()
	if !strings.Contains(output, "Overview (A3F7c9Qx7Lm2)") {
		t.Errorf("output should contain node with characters type, got:\n%s", output)
	}
	if strings.Contains(output, "Part One") {
		t.Errorf("output should NOT contain node without characters type, got:\n%s", output)
	}
	if !strings.Contains(output, "Part Two (C2xL9pQr5Tm3)") {
		t.Errorf("output should contain node with characters type, got:\n%s", output)
	}
}

func TestListCmd_HasFlags(t *testing.T) {
	tests := []struct {
		name     string
		flagName string
	}{
		{"has --depth flag", "depth"},
		{"has --type flag", "type"},
		{"has --json flag", "json"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			runner := &mockListRunner{result: &ListResult{}}
			cmd := NewListCmd(runner)

			flag := cmd.Flags().Lookup(tt.flagName)
			if flag == nil {
				t.Fatalf("list command should have --%s flag", tt.flagName)
			}
		})
	}
}

func TestListCmd_JSONOutput(t *testing.T) {
	runner := &mockListRunner{result: twoRootOutline()}
	cmd, buf := newTestListCmd(runner, "--json")

	err := cmd.Execute()

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	var output struct {
		Nodes []struct {
			MP       string   `json:"mp"`
			SID      string   `json:"sid"`
			Title    string   `json:"title"`
			Depth    int      `json:"depth"`
			Types    []string `json:"types"`
			Children []struct {
				MP       string   `json:"mp"`
				SID      string   `json:"sid"`
				Title    string   `json:"title"`
				Depth    int      `json:"depth"`
				Types    []string `json:"types"`
				Children []struct {
					MP       string        `json:"mp"`
					SID      string        `json:"sid"`
					Title    string        `json:"title"`
					Depth    int           `json:"depth"`
					Types    []string      `json:"types"`
					Children []interface{} `json:"children"`
				} `json:"children"`
			} `json:"children"`
		} `json:"nodes"`
	}
	if jsonErr := json.Unmarshal(buf.Bytes(), &output); jsonErr != nil {
		t.Fatalf("invalid JSON output: %v\nraw: %s", jsonErr, buf.String())
	}

	// Two root nodes
	if len(output.Nodes) != 2 {
		t.Fatalf("expected 2 root nodes, got %d", len(output.Nodes))
	}

	// First root node
	root := output.Nodes[0]
	if root.MP != "001" {
		t.Errorf("root.mp = %q, want %q", root.MP, "001")
	}
	if root.SID != "A3F7c9Qx7Lm2" {
		t.Errorf("root.sid = %q, want %q", root.SID, "A3F7c9Qx7Lm2")
	}
	if root.Title != "Overview" {
		t.Errorf("root.title = %q, want %q", root.Title, "Overview")
	}
	if root.Depth != 1 {
		t.Errorf("root.depth = %d, want 1", root.Depth)
	}
	if len(root.Types) != 2 {
		t.Fatalf("root.types count = %d, want 2", len(root.Types))
	}

	// Root has one child (Part One)
	if len(root.Children) != 1 {
		t.Fatalf("root.children count = %d, want 1", len(root.Children))
	}
	child := root.Children[0]
	if child.MP != "001-100" {
		t.Errorf("child.mp = %q, want %q", child.MP, "001-100")
	}
	if child.SID != "B8kQ2mNp4Rs1" {
		t.Errorf("child.sid = %q, want %q", child.SID, "B8kQ2mNp4Rs1")
	}
	if child.Depth != 2 {
		t.Errorf("child.depth = %d, want 2", child.Depth)
	}

	// Part One has two children (Chapter 1 and Chapter 2)
	if len(child.Children) != 2 {
		t.Fatalf("child.children count = %d, want 2", len(child.Children))
	}
	if child.Children[0].SID != "C2xL9pQr5Tm3" {
		t.Errorf("grandchild[0].sid = %q, want %q", child.Children[0].SID, "C2xL9pQr5Tm3")
	}
	if child.Children[1].SID != "D4yM0rSt6Un4" {
		t.Errorf("grandchild[1].sid = %q, want %q", child.Children[1].SID, "D4yM0rSt6Un4")
	}

	// Grandchildren have no children
	if len(child.Children[0].Children) != 0 {
		t.Errorf("grandchild should have no children, got %d", len(child.Children[0].Children))
	}

	// Second root node (Part Two, no children)
	root2 := output.Nodes[1]
	if root2.SID != "E6zN1sUv7Wo5" {
		t.Errorf("root2.sid = %q, want %q", root2.SID, "E6zN1sUv7Wo5")
	}
	if len(root2.Children) != 0 {
		t.Errorf("root2 should have no children, got %d", len(root2.Children))
	}
}

func TestListCmd_JSONOutput_DepthFilter(t *testing.T) {
	runner := &mockListRunner{result: threeNodeOutline()}
	cmd, buf := newTestListCmd(runner, "--json", "--depth", "2")

	err := cmd.Execute()

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	var output struct {
		Nodes []struct {
			Children []struct {
				Children []interface{} `json:"children"`
			} `json:"children"`
		} `json:"nodes"`
	}
	if jsonErr := json.Unmarshal(buf.Bytes(), &output); jsonErr != nil {
		t.Fatalf("invalid JSON output: %v\nraw: %s", jsonErr, buf.String())
	}

	if len(output.Nodes) != 1 {
		t.Fatalf("expected 1 root node, got %d", len(output.Nodes))
	}
	if len(output.Nodes[0].Children) != 1 {
		t.Fatalf("expected 1 child, got %d", len(output.Nodes[0].Children))
	}
	// Depth 2 means grandchild (depth 3) should be excluded
	if len(output.Nodes[0].Children[0].Children) != 0 {
		t.Errorf("depth=2 should exclude grandchildren, got %d", len(output.Nodes[0].Children[0].Children))
	}
}

func TestListCmd_ServiceError(t *testing.T) {
	runner := &mockListRunner{
		err: fmt.Errorf("filesystem error"),
	}
	cmd, _ := newTestListCmd(runner)

	err := cmd.Execute()

	if err == nil {
		t.Fatal("expected error for service failure")
	}
	if !strings.Contains(err.Error(), "filesystem error") {
		t.Errorf("error should contain cause, got: %v", err)
	}
}

func TestListCmd_ContextCancellation(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	cancel()

	runner := &mockListRunner{
		err: ctx.Err(),
	}
	cmd, _ := newTestListCmd(runner)

	err := cmd.Execute()

	if err == nil {
		t.Fatal("expected error for cancelled context")
	}
}

func TestListCmd_GlobalJSONFlag(t *testing.T) {
	runner := &mockListRunner{result: twoRootOutline()}
	root, buf := newTestRootListCmd(runner, "--json", "list")

	err := root.Execute()

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	var parsed map[string]interface{}
	if jsonErr := json.Unmarshal(buf.Bytes(), &parsed); jsonErr != nil {
		t.Errorf("expected valid JSON output with global --json flag, got: %s", buf.String())
	}
}

// threeRootOutline returns a flat outline matching the bug description:
// three root nodes where the first two have children.
// Expected text rendering:
//
//	act-one (A1B2C3D4E5F6)
//	├── scene-1 (A1B2C3D4E5G7)
//	└── scene-2 (A1B2C3D4E5H8)
//	act-two (B2C3D4E5F6G7)
//	└── scene-3 (B2C3D4E5F6H8)
//	act-three (C3D4E5F6G7H8)
func threeRootOutline() *ListResult {
	return &ListResult{
		Outline: domain.Outline{
			Nodes: []domain.Node{
				{
					MP:        mustMP("001"),
					SID:       "A1B2C3D4E5F6",
					Title:     "act-one",
					Documents: []domain.Document{{Type: "draft", Filename: "001_A1B2C3D4E5F6_draft_act-one.md"}},
				},
				{
					MP:        mustMP("001-100"),
					SID:       "A1B2C3D4E5G7",
					Title:     "scene-1",
					Documents: []domain.Document{{Type: "draft", Filename: "001-100_A1B2C3D4E5G7_draft_scene-1.md"}},
				},
				{
					MP:        mustMP("001-200"),
					SID:       "A1B2C3D4E5H8",
					Title:     "scene-2",
					Documents: []domain.Document{{Type: "draft", Filename: "001-200_A1B2C3D4E5H8_draft_scene-2.md"}},
				},
				{
					MP:        mustMP("002"),
					SID:       "B2C3D4E5F6G7",
					Title:     "act-two",
					Documents: []domain.Document{{Type: "draft", Filename: "002_B2C3D4E5F6G7_draft_act-two.md"}},
				},
				{
					MP:        mustMP("002-100"),
					SID:       "B2C3D4E5F6H8",
					Title:     "scene-3",
					Documents: []domain.Document{{Type: "draft", Filename: "002-100_B2C3D4E5F6H8_draft_scene-3.md"}},
				},
				{
					MP:        mustMP("003"),
					SID:       "C3D4E5F6G7H8",
					Title:     "act-three",
					Documents: []domain.Document{{Type: "draft", Filename: "003_C3D4E5F6G7H8_draft_act-three.md"}},
				},
			},
		},
	}
}

// TestRenderTreeText_MultipleRootsRenderedAtTopLevel verifies that each root node
// is printed at the top level (no branch characters), and sibling roots are NOT
// displayed as children of the first root.
func TestRenderTreeText_MultipleRootsRenderedAtTopLevel(t *testing.T) {
	roots := []*treeNode{
		{
			Title: "act-one", SID: "A1B2C3D4E5F6",
			Children: []*treeNode{
				{Title: "scene-1", SID: "A1B2C3D4E5G7", Children: []*treeNode{}},
				{Title: "scene-2", SID: "A1B2C3D4E5H8", Children: []*treeNode{}},
			},
		},
		{
			Title: "act-two", SID: "B2C3D4E5F6G7",
			Children: []*treeNode{
				{Title: "scene-3", SID: "B2C3D4E5F6H8", Children: []*treeNode{}},
			},
		},
		{
			Title:    "act-three",
			SID:      "C3D4E5F6G7H8",
			Children: []*treeNode{},
		},
	}

	var buf bytes.Buffer
	renderTreeText(&buf, roots)

	want := strings.Join([]string{
		"act-one (A1B2C3D4E5F6)",
		"├── scene-1 (A1B2C3D4E5G7)",
		"└── scene-2 (A1B2C3D4E5H8)",
		"act-two (B2C3D4E5F6G7)",
		"└── scene-3 (B2C3D4E5F6H8)",
		"act-three (C3D4E5F6G7H8)",
		"",
	}, "\n")

	got := buf.String()
	if got != want {
		t.Errorf("renderTreeText with multiple roots:\nwant:\n%s\ngot:\n%s", want, got)
	}
}

// TestListCmd_ThreeRootsExactFormat verifies the exact text output when three
// root-level nodes each have children. Each root must appear at the top level
// without tree branch connectors.
func TestListCmd_ThreeRootsExactFormat(t *testing.T) {
	runner := &mockListRunner{result: threeRootOutline()}
	cmd, buf := newTestListCmd(runner)

	err := cmd.Execute()

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	want := strings.Join([]string{
		"act-one (A1B2C3D4E5F6)",
		"├── scene-1 (A1B2C3D4E5G7)",
		"└── scene-2 (A1B2C3D4E5H8)",
		"act-two (B2C3D4E5F6G7)",
		"└── scene-3 (B2C3D4E5F6H8)",
		"act-three (C3D4E5F6G7H8)",
		"",
	}, "\n")

	got := buf.String()
	if got != want {
		t.Errorf("tree output mismatch.\nwant:\n%s\ngot:\n%s", want, got)
	}
}

// TestListCmd_MultipleRootsDepthOneExactFormat verifies that --depth=1 with
// multiple root nodes renders each root at the top level without any branch
// characters — roots must not be shown as children of the first root.
func TestListCmd_MultipleRootsDepthOneExactFormat(t *testing.T) {
	runner := &mockListRunner{result: threeRootOutline()}
	cmd, buf := newTestListCmd(runner, "--depth", "1")

	err := cmd.Execute()

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	want := strings.Join([]string{
		"act-one (A1B2C3D4E5F6)",
		"act-two (B2C3D4E5F6G7)",
		"act-three (C3D4E5F6G7H8)",
		"",
	}, "\n")

	got := buf.String()
	if got != want {
		t.Errorf("tree output with --depth=1 mismatch.\nwant:\n%s\ngot:\n%s", want, got)
	}
}
