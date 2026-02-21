package cmd

import (
	"bytes"
	"testing"

	"github.com/eykd/linemark-go/internal/outline"
)

// TestAllProjectCommands_OutsideProject_ConsistentlyError verifies that ALL
// commands which require a project consistently return ErrNotInProject when
// run in a directory without a .linemark/ project — including add, which
// currently auto-initializes instead of erroring.
//
// This test uses the same bootstrap wiring that ExecuteContextImpl currently
// applies when no project is found, so it captures the real inconsistency:
// add bootstraps while all other commands error.
func TestAllProjectCommands_OutsideProject_ConsistentlyError(t *testing.T) {
	// Simulate what ExecuteContextImpl does when no project is found:
	// it passes a bootstrapAddAdapter as the second argument.
	stub := &stubOutlineService{
		addResult: &outline.AddResult{
			SID:      "ABCD12345678",
			MP:       "100",
			Filename: "100_ABCD12345678_draft_hello.md",
		},
	}
	bootstrap := &bootstrapAddAdapter{
		getwd: func() (string, error) { return t.TempDir(), nil },
		wireService: func(root string) (outlineServicer, error) {
			return stub, nil
		},
	}

	tests := []struct {
		name string
		args []string
	}{
		{"add", []string{"add", "My Novel"}},
		{"check", []string{"check"}},
		{"doctor", []string{"doctor"}},
		{"list", []string{"list"}},
		{"compact", []string{"compact"}},
		{"delete", []string{"delete", "100"}},
		{"move", []string{"move", "100", "--to", "200"}},
		{"rename", []string{"rename", "100", "New Title"}},
		{"types list", []string{"types", "list", "100"}},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			root := BuildCommandTree(nil, bootstrap)
			root.SetOut(new(bytes.Buffer))
			root.SetErr(new(bytes.Buffer))
			root.SetArgs(tt.args)

			err := root.Execute()
			if err == nil {
				t.Fatalf("%s: expected ErrNotInProject but got nil (command may have silently succeeded outside a project)", tt.name)
			}
			if err.Error() != ErrNotInProject.Error() {
				t.Errorf("%s: got error %q, want %q", tt.name, err.Error(), ErrNotInProject.Error())
			}
		})
	}
}

// TestAddCmd_DoesNotAutoInitializeProject verifies that the add command never
// auto-initializes a .linemark/ directory when one doesn't exist. It should
// return ErrNotInProject like every other project-requiring command.
//
// This is a regression guard: the bootstrapAddAdapter used to be wired in
// ExecuteContextImpl so that add would silently create .linemark/ and
// succeed — inconsistent with list, check, doctor, and all other commands.
func TestAddCmd_DoesNotAutoInitializeProject(t *testing.T) {
	stub := &stubOutlineService{
		addResult: &outline.AddResult{
			SID:      "ABCD12345678",
			MP:       "100",
			Filename: "100_ABCD12345678_draft_hello.md",
		},
	}
	bootstrap := &bootstrapAddAdapter{
		getwd: func() (string, error) { return t.TempDir(), nil },
		wireService: func(root string) (outlineServicer, error) {
			return stub, nil
		},
	}

	root := BuildCommandTree(nil, bootstrap)
	root.SetOut(new(bytes.Buffer))
	root.SetErr(new(bytes.Buffer))
	root.SetArgs([]string{"add", "My Novel"})

	err := root.Execute()
	if err == nil {
		t.Fatal("add outside a project should return ErrNotInProject, but succeeded (auto-initialized the project)")
	}
	if err.Error() != ErrNotInProject.Error() {
		t.Errorf("add error = %q, want %q", err.Error(), ErrNotInProject.Error())
	}
}
