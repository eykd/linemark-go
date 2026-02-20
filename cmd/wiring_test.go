package cmd

import (
	"testing"

	"github.com/eykd/linemark-go/internal/outline"
)

func TestBuildCommandTree_WithNilService(t *testing.T) {
	root := BuildCommandTree(nil)

	if root == nil {
		t.Fatal("expected root command, got nil")
	}

	// All subcommands should be registered
	wantCommands := []string{"add", "check", "compact", "delete", "doctor", "list", "move", "rename", "types"}
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

func TestBuildCommandTree_WithNilService_CommandsReturnNotInProject(t *testing.T) {
	root := BuildCommandTree(nil)
	root.SetArgs([]string{"list"})

	err := root.Execute()

	if err == nil {
		t.Fatal("expected error for nil service")
	}
	if err.Error() != ErrNotInProject.Error() {
		t.Errorf("error = %q, want %q", err.Error(), ErrNotInProject.Error())
	}
}

func TestBuildCommandTree_SubcommandCount(t *testing.T) {
	root := BuildCommandTree(nil)

	want := 9
	got := len(root.Commands())
	if got != want {
		t.Errorf("subcommands = %d, want %d", got, want)
	}
}

func TestBuildCommandTree_WithService(t *testing.T) {
	svc := outline.NewOutlineService(nil, nil, nil, nil)
	root := BuildCommandTree(svc)

	if root == nil {
		t.Fatal("expected root command, got nil")
	}

	want := 9
	got := len(root.Commands())
	if got != want {
		t.Errorf("subcommands = %d, want %d", got, want)
	}
}
