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

// mockDeleteRunner is a test double for DeleteRunner.
type mockDeleteRunner struct {
	result   *DeleteResult
	err      error
	called   bool
	selector string
	apply    bool
}

func (m *mockDeleteRunner) Delete(ctx context.Context, selector string, mode domain.DeleteMode, apply bool) (*DeleteResult, error) {
	m.called = true
	m.selector = selector
	m.apply = apply
	return m.result, m.err
}

func newTestDeleteCmd(runner *mockDeleteRunner, args ...string) (*cobra.Command, *bytes.Buffer) {
	cmd := NewDeleteCmd(runner)
	if len(args) > 0 {
		cmd.SetArgs(args)
	}
	buf := new(bytes.Buffer)
	cmd.SetOut(buf)
	cmd.SetErr(new(bytes.Buffer))
	return cmd, buf
}

func TestDeleteCmd_ValidSelector(t *testing.T) {
	tests := []struct {
		name     string
		selector string
	}{
		{"implicit MP", "001-200"},
		{"implicit SID", "A3F7c9Qx7Lm2"},
		{"explicit MP prefix", "mp:001-200"},
		{"explicit SID prefix", "sid:A3F7c9Qx7Lm2"},
		{"single segment MP", "100"},
		{"explicit single segment MP", "mp:100"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			runner := &mockDeleteRunner{result: &DeleteResult{}}
			cmd, _ := newTestDeleteCmd(runner, tt.selector)

			err := cmd.Execute()

			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if !runner.called {
				t.Error("runner should be called with valid selector")
			}
		})
	}
}

func TestDeleteCmd_RejectsInvalidSelector(t *testing.T) {
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
			runner := &mockDeleteRunner{result: &DeleteResult{}}
			cmd, _ := newTestDeleteCmd(runner, tt.selector)

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

func TestDeleteCmd_RequiresExactlyOneArg(t *testing.T) {
	tests := []struct {
		name string
		args []string
	}{
		{"no args", nil},
		{"too many args", []string{"001", "002"}},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			runner := &mockDeleteRunner{result: &DeleteResult{}}
			cmd, _ := newTestDeleteCmd(runner, tt.args...)

			err := cmd.Execute()

			if err == nil {
				t.Fatal("expected error")
			}
		})
	}
}

// leafDeleteResult returns a standard test fixture for delete results.
func leafDeleteResult() *DeleteResult {
	return &DeleteResult{
		FilesDeleted: []string{
			"001-200_A3F7c9Qx7Lm2_draft_chapter-one.md",
			"001-200_A3F7c9Qx7Lm2_notes.md",
		},
		SIDsPreserved: []string{"A3F7c9Qx7Lm2"},
	}
}

// newTestRootDeleteCmd creates a delete command wired through root (for global flags like --dry-run, --json),
// capturing stdout into the returned buffer.
func newTestRootDeleteCmd(runner *mockDeleteRunner, args ...string) (*cobra.Command, *bytes.Buffer) {
	root := NewRootCmd()
	cmd := NewDeleteCmd(runner)
	root.AddCommand(cmd)
	buf := new(bytes.Buffer)
	root.SetOut(buf)
	root.SetErr(new(bytes.Buffer))
	if len(args) > 0 {
		root.SetArgs(args)
	}
	return root, buf
}

func TestDeleteCmd_HasJSONFlag(t *testing.T) {
	runner := &mockDeleteRunner{result: &DeleteResult{}}
	cmd := NewDeleteCmd(runner)

	flag := cmd.Flags().Lookup("json")
	if flag == nil {
		t.Fatal("delete command should have --json flag")
	}
	if flag.DefValue != "false" {
		t.Errorf("--json default = %q, want %q", flag.DefValue, "false")
	}
}

func TestDeleteCmd_RunnerInteraction(t *testing.T) {
	runner := &mockDeleteRunner{result: leafDeleteResult()}
	cmd, _ := newTestDeleteCmd(runner, "001-200")

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
	if !runner.apply {
		t.Error("apply should be true by default (no --dry-run)")
	}
}

func TestDeleteCmd_JSONOutput(t *testing.T) {
	runner := &mockDeleteRunner{result: leafDeleteResult()}
	cmd, buf := newTestDeleteCmd(runner, "--json", "001-200")

	err := cmd.Execute()

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	var output DeleteResult
	if jsonErr := json.Unmarshal(buf.Bytes(), &output); jsonErr != nil {
		t.Fatalf("invalid JSON output: %v\nraw: %s", jsonErr, buf.String())
	}
	if len(output.FilesDeleted) != 2 {
		t.Fatalf("files_deleted count = %d, want 2", len(output.FilesDeleted))
	}
	if output.FilesDeleted[0] != "001-200_A3F7c9Qx7Lm2_draft_chapter-one.md" {
		t.Errorf("files_deleted[0] = %q, want %q", output.FilesDeleted[0], "001-200_A3F7c9Qx7Lm2_draft_chapter-one.md")
	}
	if output.FilesDeleted[1] != "001-200_A3F7c9Qx7Lm2_notes.md" {
		t.Errorf("files_deleted[1] = %q, want %q", output.FilesDeleted[1], "001-200_A3F7c9Qx7Lm2_notes.md")
	}
	if len(output.SIDsPreserved) != 1 {
		t.Fatalf("sids_preserved count = %d, want 1", len(output.SIDsPreserved))
	}
	if output.SIDsPreserved[0] != "A3F7c9Qx7Lm2" {
		t.Errorf("sids_preserved[0] = %q, want %q", output.SIDsPreserved[0], "A3F7c9Qx7Lm2")
	}
	if output.Planned {
		t.Error("planned should be false when not in dry-run mode")
	}
}

func TestDeleteCmd_HumanReadableOutput(t *testing.T) {
	runner := &mockDeleteRunner{result: leafDeleteResult()}
	cmd, buf := newTestDeleteCmd(runner, "001-200")

	err := cmd.Execute()

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	output := buf.String()
	if !strings.Contains(output, "001-200_A3F7c9Qx7Lm2_draft_chapter-one.md") {
		t.Errorf("output should contain draft filename, got: %q", output)
	}
	if !strings.Contains(output, "001-200_A3F7c9Qx7Lm2_notes.md") {
		t.Errorf("output should contain notes filename, got: %q", output)
	}

	var parsed map[string]interface{}
	if json.Unmarshal(buf.Bytes(), &parsed) == nil {
		t.Errorf("output should not be valid JSON without --json flag, got: %s", output)
	}
}

func TestDeleteCmd_ServiceError(t *testing.T) {
	runner := &mockDeleteRunner{
		err: fmt.Errorf("filesystem error"),
	}
	cmd, _ := newTestDeleteCmd(runner, "001-200")

	err := cmd.Execute()

	if err == nil {
		t.Fatal("expected error for service failure")
	}
	if !strings.Contains(err.Error(), "filesystem error") {
		t.Errorf("error should contain cause, got: %v", err)
	}
}

func TestDeleteCmd_ContextCancellation(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	cancel()

	runner := &mockDeleteRunner{
		err: ctx.Err(),
	}
	cmd, _ := newTestDeleteCmd(runner, "001-200")

	err := cmd.Execute()

	if err == nil {
		t.Fatal("expected error for cancelled context")
	}
}

func TestDeleteCmd_DryRunBehavior(t *testing.T) {
	tests := []struct {
		name      string
		args      []string
		wantApply bool
	}{
		{
			name:      "dry-run prevents mutation",
			args:      []string{"delete", "--json", "--dry-run", "001-200"},
			wantApply: false,
		},
		{
			name:      "without dry-run applies",
			args:      []string{"delete", "--json", "001-200"},
			wantApply: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			runner := &mockDeleteRunner{result: leafDeleteResult()}
			root, _ := newTestRootDeleteCmd(runner, tt.args...)

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

func TestDeleteCmd_DryRunSetsPlanned(t *testing.T) {
	runner := &mockDeleteRunner{
		result: leafDeleteResult(),
	}
	root, buf := newTestRootDeleteCmd(runner, "delete", "--json", "--dry-run", "001-200")

	err := root.Execute()

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	var output DeleteResult
	if jsonErr := json.Unmarshal(buf.Bytes(), &output); jsonErr != nil {
		t.Fatalf("invalid JSON: %v\nraw: %s", jsonErr, buf.String())
	}
	if !output.Planned {
		t.Error("result.planned should be true when --dry-run is active")
	}
}

func TestDeleteCmd_GlobalJSONFlag(t *testing.T) {
	runner := &mockDeleteRunner{result: leafDeleteResult()}
	root, buf := newTestRootDeleteCmd(runner, "--json", "delete", "001-200")

	err := root.Execute()

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	var parsed map[string]interface{}
	if jsonErr := json.Unmarshal(buf.Bytes(), &parsed); jsonErr != nil {
		t.Errorf("expected valid JSON output with global --json flag, got: %s", buf.String())
	}
}

func TestDeleteResult_FilesDeleted_JSONTag(t *testing.T) {
	result := DeleteResult{
		FilesDeleted: []string{
			"001-200_A3F7c9Qx7Lm2_draft_chapter-one.md",
			"001-200_A3F7c9Qx7Lm2_notes.md",
		},
		SIDsPreserved: []string{"A3F7c9Qx7Lm2"},
	}
	data, err := json.Marshal(result)
	if err != nil {
		t.Fatalf("failed to marshal: %v", err)
	}

	var parsed map[string]interface{}
	if err := json.Unmarshal(data, &parsed); err != nil {
		t.Fatalf("failed to unmarshal: %v", err)
	}
	if _, ok := parsed["files_deleted"]; !ok {
		t.Fatal("JSON should include 'files_deleted' key")
	}
	files := parsed["files_deleted"].([]interface{})
	if len(files) != 2 {
		t.Errorf("files_deleted count = %d, want 2", len(files))
	}
}

func TestDeleteResult_SIDsPreserved_JSONTag(t *testing.T) {
	result := DeleteResult{
		FilesDeleted:  []string{"001-200_A3F7c9Qx7Lm2_draft_chapter-one.md"},
		SIDsPreserved: []string{"A3F7c9Qx7Lm2"},
	}
	data, err := json.Marshal(result)
	if err != nil {
		t.Fatalf("failed to marshal: %v", err)
	}

	var parsed map[string]interface{}
	if err := json.Unmarshal(data, &parsed); err != nil {
		t.Fatalf("failed to unmarshal: %v", err)
	}
	if _, ok := parsed["sids_preserved"]; !ok {
		t.Fatal("JSON should include 'sids_preserved' key")
	}
	sids := parsed["sids_preserved"].([]interface{})
	if len(sids) != 1 {
		t.Errorf("sids_preserved count = %d, want 1", len(sids))
	}
	if sids[0] != "A3F7c9Qx7Lm2" {
		t.Errorf("sids_preserved[0] = %v, want %q", sids[0], "A3F7c9Qx7Lm2")
	}
}

func TestDeleteResult_PlannedField(t *testing.T) {
	tests := []struct {
		name    string
		planned bool
	}{
		{"defaults to false", false},
		{"can be set to true", true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := DeleteResult{Planned: tt.planned}
			if result.Planned != tt.planned {
				t.Errorf("Planned = %v, want %v", result.Planned, tt.planned)
			}
		})
	}
}

func TestDeleteResult_PlannedField_JSON(t *testing.T) {
	fixture := leafDeleteResult()
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
