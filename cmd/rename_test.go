package cmd

import (
	"bytes"
	"context"
	"errors"
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
			runner := &mockRenameRunner{result: &RenameResult{}}
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
			runner := &mockRenameRunner{result: &RenameResult{}}
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
			runner := &mockRenameRunner{result: &RenameResult{}}
			cmd, _ := newTestRenameCmd(runner, tt.args...)

			err := cmd.Execute()

			if err == nil {
				t.Fatal("expected error")
			}
		})
	}
}
