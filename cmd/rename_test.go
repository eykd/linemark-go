package cmd

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"strings"
	"testing"

	"github.com/eykd/linemark-go/internal/domain"
	"github.com/spf13/cobra"
)

// mockRenameRunner is a test double for RenameRunner.
type mockRenameRunner struct {
	result   *RenameResult
	err      error
	called   bool
	selector string
	newTitle string
	apply    bool
}

func (m *mockRenameRunner) Rename(ctx context.Context, selector string, newTitle string, apply bool) (*RenameResult, error) {
	m.called = true
	m.selector = selector
	m.newTitle = newTitle
	m.apply = apply
	return m.result, m.err
}

func newTestRenameCmd(runner *mockRenameRunner, args ...string) (*cobra.Command, *bytes.Buffer) {
	cmd := NewRenameCmd(runner)
	if len(args) > 0 {
		cmd.SetArgs(args)
	}
	buf := new(bytes.Buffer)
	cmd.SetOut(buf)
	cmd.SetErr(new(bytes.Buffer))
	return cmd, buf
}

// newTestRootRenameCmd creates a rename command wired through root (for global flags like --dry-run, --json),
// capturing stdout into the returned buffer.
func newTestRootRenameCmd(runner *mockRenameRunner, args ...string) (*cobra.Command, *bytes.Buffer) {
	root := NewRootCmd()
	cmd := NewRenameCmd(runner)
	root.AddCommand(cmd)
	buf := new(bytes.Buffer)
	root.SetOut(buf)
	root.SetErr(new(bytes.Buffer))
	if len(args) > 0 {
		root.SetArgs(args)
	}
	return root, buf
}

// renameFixture returns a standard test fixture for rename results.
func renameFixture() *RenameResult {
	return &RenameResult{
		Node: RenameNodeInfo{
			MP:       "001-200",
			SID:      "A3F7c9Qx7Lm2",
			OldTitle: "Chapter One",
			NewTitle: "The Beginning",
		},
		Renames: []RenameEntry{
			{Old: "001-200_A3F7c9Qx7Lm2_draft_chapter-one.md", New: "001-200_A3F7c9Qx7Lm2_draft_the-beginning.md"},
		},
	}
}

func TestRenameCmd_ValidSelector(t *testing.T) {
	tests := []struct {
		name     string
		selector string
	}{
		{"implicit MP", "001-200"},
		{"implicit SID", "A3F7c9Qx7Lm2"},
		{"explicit MP prefix", "mp:001-200"},
		{"explicit SID prefix", "sid:A3F7c9Qx7Lm2"},
		{"single segment MP", "100"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			runner := &mockRenameRunner{result: renameFixture()}
			cmd, _ := newTestRenameCmd(runner, tt.selector, "New Title")

			err := cmd.Execute()

			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if !runner.called {
				t.Error("runner should be called with valid selector")
			}
			if runner.newTitle != "New Title" {
				t.Errorf("newTitle = %q, want %q", runner.newTitle, "New Title")
			}
		})
	}
}

func TestRenameCmd_RejectsInvalidSelector(t *testing.T) {
	tests := []struct {
		name     string
		selector string
	}{
		{"special chars", "abc!@#"},
		{"too short", "ab"},
		{"unknown prefix", "foo:123"},
		{"mp prefix bad value", "mp:invalid"},
		{"sid prefix bad value", "sid:ab"},
		{"zero segment", "000"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			runner := &mockRenameRunner{result: renameFixture()}
			cmd, _ := newTestRenameCmd(runner, tt.selector, "New Title")

			err := cmd.Execute()

			if err == nil {
				t.Fatal("expected error for invalid selector")
			}
			if !errors.Is(err, domain.ErrInvalidSelector) {
				t.Errorf("error should wrap ErrInvalidSelector, got: %v", err)
			}
			if runner.called {
				t.Error("runner should not be called with invalid selector")
			}
		})
	}
}

func TestRenameCmd_RequiresExactlyTwoArgs(t *testing.T) {
	tests := []struct {
		name string
		args []string
	}{
		{"no args", nil},
		{"only selector", []string{"001"}},
		{"too many args", []string{"001", "New Title", "extra"}},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			runner := &mockRenameRunner{result: renameFixture()}
			cmd, _ := newTestRenameCmd(runner, tt.args...)

			err := cmd.Execute()

			if err == nil {
				t.Fatal("expected error")
			}
		})
	}
}

func TestRenameCmd_RunnerInteraction(t *testing.T) {
	runner := &mockRenameRunner{result: renameFixture()}
	cmd, _ := newTestRenameCmd(runner, "001-200", "The Beginning")

	err := cmd.Execute()

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !runner.called {
		t.Error("runner should be called on execute")
	}
	if runner.selector != "001-200" {
		t.Errorf("selector = %q, want %q", runner.selector, "001-200")
	}
	if runner.newTitle != "The Beginning" {
		t.Errorf("newTitle = %q, want %q", runner.newTitle, "The Beginning")
	}
	if !runner.apply {
		t.Error("apply should be true by default (no --dry-run)")
	}
}

func TestRenameCmd_JSONOutput(t *testing.T) {
	runner := &mockRenameRunner{result: renameFixture()}
	root, buf := newTestRootRenameCmd(runner, "--json", "rename", "001-200", "The Beginning")

	err := root.Execute()

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	var output RenameResult
	if jsonErr := json.Unmarshal(buf.Bytes(), &output); jsonErr != nil {
		t.Fatalf("invalid JSON output: %v\nraw: %s", jsonErr, buf.String())
	}
	if output.Node.MP != "001-200" {
		t.Errorf("node.mp = %q, want %q", output.Node.MP, "001-200")
	}
	if output.Node.SID != "A3F7c9Qx7Lm2" {
		t.Errorf("node.sid = %q, want %q", output.Node.SID, "A3F7c9Qx7Lm2")
	}
	if output.Node.OldTitle != "Chapter One" {
		t.Errorf("node.old_title = %q, want %q", output.Node.OldTitle, "Chapter One")
	}
	if output.Node.NewTitle != "The Beginning" {
		t.Errorf("node.new_title = %q, want %q", output.Node.NewTitle, "The Beginning")
	}
	if len(output.Renames) != 1 {
		t.Fatalf("renames count = %d, want 1", len(output.Renames))
	}
	if output.Renames[0].Old != "001-200_A3F7c9Qx7Lm2_draft_chapter-one.md" {
		t.Errorf("renames[0].old = %q, want %q", output.Renames[0].Old, "001-200_A3F7c9Qx7Lm2_draft_chapter-one.md")
	}
	if output.Renames[0].New != "001-200_A3F7c9Qx7Lm2_draft_the-beginning.md" {
		t.Errorf("renames[0].new = %q, want %q", output.Renames[0].New, "001-200_A3F7c9Qx7Lm2_draft_the-beginning.md")
	}
	if output.Planned {
		t.Error("planned should be false when not in dry-run mode")
	}
}

func TestRenameCmd_HumanReadableOutput(t *testing.T) {
	runner := &mockRenameRunner{result: renameFixture()}
	cmd, buf := newTestRenameCmd(runner, "001-200", "The Beginning")

	err := cmd.Execute()

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	output := buf.String()
	if !strings.Contains(output, "Chapter One") {
		t.Errorf("output should contain old title, got: %q", output)
	}
	if !strings.Contains(output, "The Beginning") {
		t.Errorf("output should contain new title, got: %q", output)
	}
	if !strings.Contains(output, "001-200_A3F7c9Qx7Lm2_draft_the-beginning.md") {
		t.Errorf("output should contain renamed filename, got: %q", output)
	}

	var parsed map[string]interface{}
	if json.Unmarshal(buf.Bytes(), &parsed) == nil {
		t.Errorf("output should not be valid JSON without --json flag, got: %s", output)
	}
}

func TestRenameCmd_ServiceError(t *testing.T) {
	runner := &mockRenameRunner{
		err: fmt.Errorf("filesystem error"),
	}
	cmd, _ := newTestRenameCmd(runner, "001-200", "New Title")

	err := cmd.Execute()

	if err == nil {
		t.Fatal("expected error for service failure")
	}
	if !strings.Contains(err.Error(), "filesystem error") {
		t.Errorf("error should contain cause, got: %v", err)
	}
}

func TestRenameCmd_ContextCancellation(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	cancel()

	runner := &mockRenameRunner{
		err: ctx.Err(),
	}
	cmd, _ := newTestRenameCmd(runner, "001-200", "New Title")

	err := cmd.Execute()

	if err == nil {
		t.Fatal("expected error for cancelled context")
	}
}

func TestRenameCmd_DryRunBehavior(t *testing.T) {
	tests := []struct {
		name      string
		args      []string
		wantApply bool
	}{
		{
			name:      "dry-run prevents mutation",
			args:      []string{"rename", "--json", "--dry-run", "001-200", "New Title"},
			wantApply: false,
		},
		{
			name:      "without dry-run applies",
			args:      []string{"rename", "--json", "001-200", "New Title"},
			wantApply: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			runner := &mockRenameRunner{result: renameFixture()}
			root, _ := newTestRootRenameCmd(runner, tt.args...)

			err := root.Execute()

			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if runner.apply != tt.wantApply {
				t.Errorf("apply = %v, want %v", runner.apply, tt.wantApply)
			}
		})
	}
}

func TestRenameCmd_DryRunSetsPlanned(t *testing.T) {
	runner := &mockRenameRunner{
		result: renameFixture(),
	}
	root, buf := newTestRootRenameCmd(runner, "rename", "--json", "--dry-run", "001-200", "New Title")

	err := root.Execute()

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	var output RenameResult
	if jsonErr := json.Unmarshal(buf.Bytes(), &output); jsonErr != nil {
		t.Fatalf("invalid JSON: %v\nraw: %s", jsonErr, buf.String())
	}
	if !output.Planned {
		t.Error("result.planned should be true when --dry-run is active")
	}
}

func TestRenameCmd_GlobalJSONFlag(t *testing.T) {
	runner := &mockRenameRunner{result: renameFixture()}
	root, buf := newTestRootRenameCmd(runner, "--json", "rename", "001-200", "The Beginning")

	err := root.Execute()

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	var parsed map[string]interface{}
	if jsonErr := json.Unmarshal(buf.Bytes(), &parsed); jsonErr != nil {
		t.Errorf("expected valid JSON output with global --json flag, got: %s", buf.String())
	}
}

func TestRenameResult_NodeInfo_JSONTags(t *testing.T) {
	result := renameFixture()
	data, err := json.Marshal(result)
	if err != nil {
		t.Fatalf("failed to marshal: %v", err)
	}

	var parsed map[string]interface{}
	if err := json.Unmarshal(data, &parsed); err != nil {
		t.Fatalf("failed to unmarshal: %v", err)
	}

	node, ok := parsed["node"]
	if !ok {
		t.Fatal("JSON should include 'node' key")
	}
	nodeMap := node.(map[string]interface{})

	checks := map[string]string{
		"mp":        "001-200",
		"sid":       "A3F7c9Qx7Lm2",
		"old_title": "Chapter One",
		"new_title": "The Beginning",
	}
	for key, want := range checks {
		got, exists := nodeMap[key]
		if !exists {
			t.Errorf("node JSON should include %q key", key)
			continue
		}
		if got != want {
			t.Errorf("node.%s = %v, want %q", key, got, want)
		}
	}
}

func TestRenameResult_Renames_JSONTag(t *testing.T) {
	result := renameFixture()
	data, err := json.Marshal(result)
	if err != nil {
		t.Fatalf("failed to marshal: %v", err)
	}

	var parsed map[string]interface{}
	if err := json.Unmarshal(data, &parsed); err != nil {
		t.Fatalf("failed to unmarshal: %v", err)
	}
	renames, ok := parsed["renames"]
	if !ok {
		t.Fatal("JSON should include 'renames' key")
	}
	renameArr := renames.([]interface{})
	if len(renameArr) != 1 {
		t.Fatalf("renames count = %d, want 1", len(renameArr))
	}
	entry := renameArr[0].(map[string]interface{})
	if entry["old"] != "001-200_A3F7c9Qx7Lm2_draft_chapter-one.md" {
		t.Errorf("renames[0].old = %v, want %q", entry["old"], "001-200_A3F7c9Qx7Lm2_draft_chapter-one.md")
	}
	if entry["new"] != "001-200_A3F7c9Qx7Lm2_draft_the-beginning.md" {
		t.Errorf("renames[0].new = %v, want %q", entry["new"], "001-200_A3F7c9Qx7Lm2_draft_the-beginning.md")
	}
}

func TestRenameResult_PlannedField(t *testing.T) {
	tests := []struct {
		name    string
		planned bool
	}{
		{"defaults to false", false},
		{"can be set to true", true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := RenameResult{Planned: tt.planned}
			if result.Planned != tt.planned {
				t.Errorf("Planned = %v, want %v", result.Planned, tt.planned)
			}
		})
	}
}

func TestRenameCmd_RegisteredWithRoot(t *testing.T) {
	found := false
	for _, sub := range rootCmd.Commands() {
		if sub.Name() == "rename" {
			found = true
			break
		}
	}
	if !found {
		t.Error("rename command not registered with root")
	}
}

func TestRenameResult_PlannedField_JSON(t *testing.T) {
	fixture := renameFixture()
	fixture.Planned = true
	data, err := json.Marshal(fixture)
	if err != nil {
		t.Fatalf("failed to marshal: %v", err)
	}

	var parsed map[string]interface{}
	if err := json.Unmarshal(data, &parsed); err != nil {
		t.Fatalf("failed to unmarshal: %v", err)
	}
	if _, ok := parsed["planned"]; !ok {
		t.Fatal("JSON should include 'planned' key")
	}
	if parsed["planned"] != true {
		t.Errorf("planned = %v, want true", parsed["planned"])
	}
}

func TestRenameResult_MultipleRenames(t *testing.T) {
	result := &RenameResult{
		Node: RenameNodeInfo{
			MP:       "001-200",
			SID:      "A3F7c9Qx7Lm2",
			OldTitle: "Chapter One",
			NewTitle: "The Beginning",
		},
		Renames: []RenameEntry{
			{Old: "001-200_A3F7c9Qx7Lm2_draft_chapter-one.md", New: "001-200_A3F7c9Qx7Lm2_draft_the-beginning.md"},
			{Old: "001-200_A3F7c9Qx7Lm2_notes_chapter-one.md", New: "001-200_A3F7c9Qx7Lm2_notes_the-beginning.md"},
		},
	}

	data, err := json.Marshal(result)
	if err != nil {
		t.Fatalf("failed to marshal: %v", err)
	}

	var output RenameResult
	if err := json.Unmarshal(data, &output); err != nil {
		t.Fatalf("failed to unmarshal: %v", err)
	}
	if len(output.Renames) != 2 {
		t.Fatalf("renames count = %d, want 2", len(output.Renames))
	}
}

func TestRenameCmd_HumanReadableMultipleRenames(t *testing.T) {
	result := &RenameResult{
		Node: RenameNodeInfo{
			MP:       "001-200",
			SID:      "A3F7c9Qx7Lm2",
			OldTitle: "Chapter One",
			NewTitle: "The Beginning",
		},
		Renames: []RenameEntry{
			{Old: "001-200_A3F7c9Qx7Lm2_draft_chapter-one.md", New: "001-200_A3F7c9Qx7Lm2_draft_the-beginning.md"},
			{Old: "001-200_A3F7c9Qx7Lm2_notes_chapter-one.md", New: "001-200_A3F7c9Qx7Lm2_notes_the-beginning.md"},
		},
	}

	runner := &mockRenameRunner{result: result}
	cmd, buf := newTestRenameCmd(runner, "001-200", "The Beginning")

	err := cmd.Execute()

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	output := buf.String()
	if !strings.Contains(output, "001-200_A3F7c9Qx7Lm2_draft_the-beginning.md") {
		t.Errorf("output should contain renamed draft filename, got: %q", output)
	}
	if !strings.Contains(output, "001-200_A3F7c9Qx7Lm2_notes_the-beginning.md") {
		t.Errorf("output should contain renamed notes filename, got: %q", output)
	}
}
