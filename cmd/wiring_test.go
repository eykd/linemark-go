package cmd

import (
	"bytes"
	"testing"

	"github.com/eykd/linemark-go/internal/outline"
)

func TestBuildCommandTree_WithNilService(t *testing.T) {
	root := BuildCommandTree(nil, nil)

	if root == nil {
		t.Fatal("expected root command, got nil")
	}

	// All subcommands should be registered
	wantCommands := []string{"add", "check", "compact", "delete", "doctor", "init", "list", "move", "rename", "types"}
	for _, name := range wantCommands {
		found := false
		for _, sub := range root.Commands() {
			if sub.Name() == name {
				found = true
				break
			}
		}
		if !found {
			t.Errorf("expected subcommand %q to be registered", name)
		}
	}
}

// TestBuildCommandTree_AllCommandsHandleNilService verifies that every command
// registered in BuildCommandTree(nil, nil) returns ErrNotInProject when run
// without a service, except for init which works without one. If a new command
// is added without a nil guard, this test catches it.
func TestBuildCommandTree_AllCommandsHandleNilService(t *testing.T) {
	commands := []struct {
		args    []string
		wantErr string // empty means no error expected
	}{
		{[]string{"add", "Title"}, ErrNotInProject.Error()},
		{[]string{"check"}, ErrNotInProject.Error()},
		{[]string{"compact"}, ErrNotInProject.Error()},
		{[]string{"delete", "100"}, ErrNotInProject.Error()},
		{[]string{"doctor"}, ErrNotInProject.Error()},
		{[]string{"list"}, ErrNotInProject.Error()},
		{[]string{"move", "100", "--to", "200"}, ErrNotInProject.Error()},
		{[]string{"rename", "100", "New"}, ErrNotInProject.Error()},
		{[]string{"types", "list", "100"}, ErrNotInProject.Error()},
		{[]string{"init", "--help"}, ""}, // init works without service
	}
	for _, tt := range commands {
		t.Run(tt.args[0], func(t *testing.T) {
			cmd := BuildCommandTree(nil, nil)
			cmd.SetArgs(tt.args)
			cmd.SetOut(new(bytes.Buffer))
			cmd.SetErr(new(bytes.Buffer))
			err := cmd.Execute()
			if tt.wantErr == "" {
				if err != nil {
					t.Errorf("unexpected error: %v", err)
				}
			} else {
				if err == nil || err.Error() != tt.wantErr {
					t.Errorf("got %v, want %q", err, tt.wantErr)
				}
			}
		})
	}
}

func TestBuildCommandTree_SubcommandCount(t *testing.T) {
	root := BuildCommandTree(nil, nil)

	want := 10
	got := len(root.Commands())
	if got != want {
		t.Errorf("subcommands = %d, want %d", got, want)
	}
}

func TestBuildCommandTree_WithService(t *testing.T) {
	svc := outline.NewOutlineService(nil, nil, nil, nil)
	root := BuildCommandTree(svc, nil)

	if root == nil {
		t.Fatal("expected root command, got nil")
	}

	want := 10
	got := len(root.Commands())
	if got != want {
		t.Errorf("subcommands = %d, want %d", got, want)
	}
}
