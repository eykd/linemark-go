package cmd

import (
	"bytes"
	"context"
	"errors"
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

func (m *mockDeleteRunner) Delete(ctx context.Context, selector string, apply bool) (*DeleteResult, error) {
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
