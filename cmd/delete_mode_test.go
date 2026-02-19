package cmd

import (
	"bytes"
	"context"
	"encoding/json"
	"strings"
	"testing"

	"github.com/eykd/linemark-go/internal/domain"
	"github.com/spf13/cobra"
)

// mockDeleteModeRunner captures the delete mode passed by the command.
type mockDeleteModeRunner struct {
	result   *DeleteResult
	err      error
	called   bool
	selector string
	mode     domain.DeleteMode
	apply    bool
}

func (m *mockDeleteModeRunner) Delete(ctx context.Context, selector string, mode domain.DeleteMode, apply bool) (*DeleteResult, error) {
	m.called = true
	m.selector = selector
	m.mode = mode
	m.apply = apply
	return m.result, m.err
}

func newTestDeleteModeCmd(runner *mockDeleteModeRunner, args ...string) (*cobra.Command, *bytes.Buffer) {
	cmd := NewDeleteCmd(runner)
	if len(args) > 0 {
		cmd.SetArgs(args)
	}
	buf := new(bytes.Buffer)
	cmd.SetOut(buf)
	cmd.SetErr(new(bytes.Buffer))
	return cmd, buf
}

func TestDeleteCmd_HasRecursiveFlag(t *testing.T) {
	runner := &mockDeleteModeRunner{result: &DeleteResult{}}
	cmd := NewDeleteCmd(runner)

	flag := cmd.Flags().Lookup("recursive")
	if flag == nil {
		t.Fatal("delete command should have --recursive flag")
	}
	if flag.Shorthand != "r" {
		t.Errorf("--recursive shorthand = %q, want %q", flag.Shorthand, "r")
	}
	if flag.DefValue != "false" {
		t.Errorf("--recursive default = %q, want %q", flag.DefValue, "false")
	}
}

func TestDeleteCmd_HasPromoteFlag(t *testing.T) {
	runner := &mockDeleteModeRunner{result: &DeleteResult{}}
	cmd := NewDeleteCmd(runner)

	flag := cmd.Flags().Lookup("promote")
	if flag == nil {
		t.Fatal("delete command should have --promote flag")
	}
	if flag.Shorthand != "p" {
		t.Errorf("--promote shorthand = %q, want %q", flag.Shorthand, "p")
	}
	if flag.DefValue != "false" {
		t.Errorf("--promote default = %q, want %q", flag.DefValue, "false")
	}
}

func TestDeleteCmd_RecursiveAndPromoteMutuallyExclusive(t *testing.T) {
	runner := &mockDeleteModeRunner{result: &DeleteResult{}}
	cmd, _ := newTestDeleteModeCmd(runner, "--recursive", "--promote", "001-200")

	err := cmd.Execute()

	if err == nil {
		t.Fatal("expected error when both --recursive and --promote are set")
	}
	if !strings.Contains(err.Error(), "recursive") || !strings.Contains(err.Error(), "promote") {
		t.Errorf("error should mention conflicting flags, got: %v", err)
	}
	if runner.called {
		t.Error("runner should not be called when flags conflict")
	}
}

func TestDeleteCmd_PassesDeleteMode(t *testing.T) {
	tests := []struct {
		name     string
		args     []string
		wantMode domain.DeleteMode
	}{
		{"default mode", []string{"001-200"}, domain.DeleteModeDefault},
		{"recursive mode", []string{"--recursive", "001-200"}, domain.DeleteModeRecursive},
		{"promote mode", []string{"--promote", "001-200"}, domain.DeleteModePromote},
		{"recursive shorthand", []string{"-r", "001-200"}, domain.DeleteModeRecursive},
		{"promote shorthand", []string{"-p", "001-200"}, domain.DeleteModePromote},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			runner := &mockDeleteModeRunner{result: &DeleteResult{}}
			cmd, _ := newTestDeleteModeCmd(runner, tt.args...)

			err := cmd.Execute()

			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if !runner.called {
				t.Fatal("runner should be called")
			}
			if runner.mode != tt.wantMode {
				t.Errorf("mode = %v, want %v", runner.mode, tt.wantMode)
			}
		})
	}
}

func TestDeleteCmd_RecursiveDeleteJSONOutput(t *testing.T) {
	result := &DeleteResult{
		FilesDeleted: []string{
			"100_SID001AABB_draft_part-one.md",
			"100_SID001AABB_notes.md",
			"100-100_SID002CCDD_draft_chapter-one.md",
			"100-100_SID002CCDD_notes.md",
			"100-200_SID003EEFF_draft_chapter-two.md",
			"100-200_SID003EEFF_notes.md",
		},
		SIDsPreserved: []string{"SID001AABB", "SID002CCDD", "SID003EEFF"},
	}
	runner := &mockDeleteModeRunner{result: result}
	cmd, buf := newTestDeleteModeCmd(runner, "--json", "--recursive", "100")

	err := cmd.Execute()

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	var output DeleteResult
	if jsonErr := json.Unmarshal(buf.Bytes(), &output); jsonErr != nil {
		t.Fatalf("invalid JSON output: %v\nraw: %s", jsonErr, buf.String())
	}
	if len(output.FilesDeleted) != 6 {
		t.Errorf("files_deleted count = %d, want 6", len(output.FilesDeleted))
	}
	if len(output.SIDsPreserved) != 3 {
		t.Errorf("sids_preserved count = %d, want 3", len(output.SIDsPreserved))
	}
}

func TestDeleteCmd_PromoteDeleteJSONOutput(t *testing.T) {
	result := &DeleteResult{
		FilesDeleted: []string{
			"100_SID001AABB_draft_part-one.md",
			"100_SID001AABB_notes.md",
		},
		FilesRenamed: map[string]string{
			"100-100_SID002CCDD_draft_chapter-one.md": "100_SID002CCDD_draft_chapter-one.md",
			"100-100_SID002CCDD_notes.md":             "100_SID002CCDD_notes.md",
			"100-200_SID003EEFF_draft_chapter-two.md": "200_SID003EEFF_draft_chapter-two.md",
			"100-200_SID003EEFF_notes.md":             "200_SID003EEFF_notes.md",
		},
		SIDsPreserved: []string{"SID001AABB"},
	}
	runner := &mockDeleteModeRunner{result: result}
	cmd, buf := newTestDeleteModeCmd(runner, "--json", "--promote", "100")

	err := cmd.Execute()

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	var output map[string]interface{}
	if jsonErr := json.Unmarshal(buf.Bytes(), &output); jsonErr != nil {
		t.Fatalf("invalid JSON output: %v\nraw: %s", jsonErr, buf.String())
	}
	if _, ok := output["files_renamed"]; !ok {
		t.Error("JSON output should include 'files_renamed' for promote mode")
	}
	renamed := output["files_renamed"].(map[string]interface{})
	if len(renamed) != 4 {
		t.Errorf("files_renamed count = %d, want 4", len(renamed))
	}
}

func TestDeleteCmd_PromoteHumanOutput_ShowsRenames(t *testing.T) {
	result := &DeleteResult{
		FilesDeleted: []string{
			"100_SID001AABB_draft_part-one.md",
		},
		FilesRenamed: map[string]string{
			"100-100_SID002CCDD_draft_chapter-one.md": "100_SID002CCDD_draft_chapter-one.md",
		},
		SIDsPreserved: []string{"SID001AABB"},
	}
	runner := &mockDeleteModeRunner{result: result}
	cmd, buf := newTestDeleteModeCmd(runner, "--promote", "100")

	err := cmd.Execute()

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	output := buf.String()
	if !strings.Contains(output, "100_SID001AABB_draft_part-one.md") {
		t.Error("output should list deleted files")
	}
	if !strings.Contains(output, "100_SID002CCDD_draft_chapter-one.md") {
		t.Error("output should show renamed file destinations")
	}
}
