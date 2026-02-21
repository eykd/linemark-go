package cmd

// Tests for human-readable (non-JSON) output in --dry-run mode.
//
// The bugs under test:
//   - `lmk types add --dry-run` prints "Added <file>" instead of something like "Would add <file>"
//   - `lmk types remove --dry-run` prints "Removed <file>" instead of "Would remove <file>"
//   - `lmk add --dry-run` produces no output at all (FilesPlanned is never printed)

import (
	"bytes"
	"strings"
	"testing"
)

// TestTypesAddCmd_DryRunHumanOutput_DoesNotSayAdded verifies that
// `types add --dry-run` does not print the normal "Added" message.
//
// Currently FAILS because newTypesAddCmd always prints "Added <filename>"
// regardless of dry-run mode.
func TestTypesAddCmd_DryRunHumanOutput_DoesNotSayAdded(t *testing.T) {
	svc := &mockTypesService{
		addResult: &TypesModifyResult{
			Node:     NodeInfo{MP: "001", SID: "A3F7c9Qx7Lm2"},
			Filename: "001_A3F7c9Qx7Lm2_characters.md",
		},
	}
	root := NewRootCmd()
	typesCmd := NewTypesCmd(svc)
	root.AddCommand(typesCmd)
	buf := new(bytes.Buffer)
	root.SetOut(buf)
	root.SetErr(new(bytes.Buffer))
	root.SetArgs([]string{"types", "add", "--dry-run", "characters", "001"})

	err := root.Execute()

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	output := buf.String()
	if strings.Contains(output, "Added") {
		t.Errorf("dry-run output should not say 'Added', got: %q", output)
	}
	if output == "" {
		t.Error("dry-run output should not be empty; expected a 'would add' message")
	}
}

// TestTypesRemoveCmd_DryRunHumanOutput_DoesNotSayRemoved verifies that
// `types remove --dry-run` does not print the normal "Removed" message.
//
// Currently FAILS because newTypesRemoveCmd always prints "Removed <filename>"
// regardless of dry-run mode.
func TestTypesRemoveCmd_DryRunHumanOutput_DoesNotSayRemoved(t *testing.T) {
	svc := &mockTypesService{
		removeResult: &TypesModifyResult{
			Node:     NodeInfo{MP: "001", SID: "A3F7c9Qx7Lm2"},
			Filename: "001_A3F7c9Qx7Lm2_characters.md",
		},
	}
	root := NewRootCmd()
	typesCmd := NewTypesCmd(svc)
	root.AddCommand(typesCmd)
	buf := new(bytes.Buffer)
	root.SetOut(buf)
	root.SetErr(new(bytes.Buffer))
	root.SetArgs([]string{"types", "remove", "--dry-run", "characters", "001"})

	err := root.Execute()

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	output := buf.String()
	if strings.Contains(output, "Removed") {
		t.Errorf("dry-run output should not say 'Removed', got: %q", output)
	}
	if output == "" {
		t.Error("dry-run output should not be empty; expected a 'would remove' message")
	}
}

// TestAddCmd_DryRunHumanOutput_PrintsPlannedFiles verifies that
// `add --dry-run` prints the planned filenames rather than silently producing no output.
//
// Currently FAILS because add.go only iterates result.FilesCreated (which is empty
// in dry-run mode) and never prints result.FilesPlanned.
func TestAddCmd_DryRunHumanOutput_PrintsPlannedFiles(t *testing.T) {
	runner := &mockAddRunner{
		result: &AddResult{
			Node: AddNodeInfo{
				MP:    "100",
				SID:   "(pending)",
				Title: "Chapter One",
			},
			FilesPlanned: []string{
				"100_(sid)_draft_chapter-one.md",
				"100_(sid)_notes.md",
			},
		},
	}
	root, buf := newTestRootAddCmd(runner, "add", "--dry-run", "Chapter One")

	err := root.Execute()

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	output := buf.String()
	if output == "" {
		t.Error("dry-run output should not be empty; expected planned file names to be printed")
	}
	for _, f := range runner.result.FilesPlanned {
		if !strings.Contains(output, f) {
			t.Errorf("dry-run output should contain planned file %q, got: %q", f, output)
		}
	}
}
