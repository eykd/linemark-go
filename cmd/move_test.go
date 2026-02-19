package cmd

import (
	"bytes"
	"context"
	"errors"
	"testing"

	"github.com/eykd/linemark-go/internal/domain"
	"github.com/spf13/cobra"
)

// mockMoveRunner is a test double for MoveRunner.
type mockMoveRunner struct {
	result   *MoveResult
	err      error
	called   bool
	selector string
	to       string
	apply    bool
}

func (m *mockMoveRunner) Move(ctx context.Context, selector string, to string, apply bool) (*MoveResult, error) {
	m.called = true
	m.selector = selector
	m.to = to
	m.apply = apply
	return m.result, m.err
}

func newTestMoveCmd(runner *mockMoveRunner, args ...string) (*cobra.Command, *bytes.Buffer) {
	cmd := NewMoveCmd(runner)
	if len(args) > 0 {
		cmd.SetArgs(args)
	}
	buf := new(bytes.Buffer)
	cmd.SetOut(buf)
	cmd.SetErr(new(bytes.Buffer))
	return cmd, buf
}

func TestMoveCmd_ValidSelectors(t *testing.T) {
	tests := []struct {
		name string
		args []string
	}{
		{"implicit MP source and target", []string{"001-200", "--to", "300"}},
		{"implicit SID source", []string{"A3F7c9Qx7Lm2", "--to", "300"}},
		{"explicit MP prefix source", []string{"mp:001-200", "--to", "300"}},
		{"explicit SID prefix source", []string{"sid:A3F7c9Qx7Lm2", "--to", "300"}},
		{"explicit MP prefix target", []string{"100", "--to", "mp:300"}},
		{"explicit SID prefix target", []string{"100", "--to", "sid:B8kQ2mNp4Rs1"}},
		{"both explicit prefixes", []string{"mp:100", "--to", "sid:A3F7c9Qx7Lm2"}},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			runner := &mockMoveRunner{result: &MoveResult{}}
			cmd, _ := newTestMoveCmd(runner, tt.args...)

			err := cmd.Execute()

			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if !runner.called {
				t.Error("runner should be called with valid selectors")
			}
		})
	}
}

func TestMoveCmd_RejectsInvalidSourceSelector(t *testing.T) {
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
			runner := &mockMoveRunner{result: &MoveResult{}}
			cmd, _ := newTestMoveCmd(runner, tt.selector, "--to", "100")

			err := cmd.Execute()

			if err == nil {
				t.Fatal("expected error for invalid source selector")
			}
			if !errors.Is(err, domain.ErrInvalidSelector) {
				t.Errorf("error should wrap ErrInvalidSelector, got: %v", err)
			}
			if runner.called {
				t.Error("runner should not be called with invalid source selector")
			}
		})
	}
}

func TestMoveCmd_RejectsInvalidTargetSelector(t *testing.T) {
	tests := []struct {
		name string
		to   string
	}{
		{"special chars", "abc!@#"},
		{"too short", "ab"},
		{"unknown prefix", "foo:123"},
		{"mp prefix bad value", "mp:invalid"},
		{"zero segment", "000"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			runner := &mockMoveRunner{result: &MoveResult{}}
			cmd, _ := newTestMoveCmd(runner, "100", "--to", tt.to)

			err := cmd.Execute()

			if err == nil {
				t.Fatal("expected error for invalid target selector")
			}
			if !errors.Is(err, domain.ErrInvalidSelector) {
				t.Errorf("error should wrap ErrInvalidSelector, got: %v", err)
			}
			if runner.called {
				t.Error("runner should not be called with invalid target selector")
			}
		})
	}
}

func TestMoveCmd_RequiresExactlyOneArg(t *testing.T) {
	tests := []struct {
		name string
		args []string
	}{
		{"no args", nil},
		{"too many args", []string{"001", "002", "--to", "300"}},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			runner := &mockMoveRunner{result: &MoveResult{}}
			cmd, _ := newTestMoveCmd(runner, tt.args...)

			err := cmd.Execute()

			if err == nil {
				t.Fatal("expected error")
			}
		})
	}
}

func TestMoveCmd_RequiresTo(t *testing.T) {
	runner := &mockMoveRunner{result: &MoveResult{}}
	cmd, _ := newTestMoveCmd(runner, "100")

	err := cmd.Execute()

	if err == nil {
		t.Fatal("expected error when --to is missing")
	}
	if runner.called {
		t.Error("runner should not be called without --to")
	}
}
