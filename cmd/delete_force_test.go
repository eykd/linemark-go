package cmd

import (
	"bytes"
	"fmt"
	"strings"
	"testing"

	"github.com/spf13/cobra"
)

// mockConfirmer is a test double for confirming destructive operations.
type mockConfirmer struct {
	result bool
	err    error
	called bool
	prompt string
}

func (m *mockConfirmer) Confirm(prompt string) (bool, error) {
	m.called = true
	m.prompt = prompt
	return m.result, m.err
}

// newTestDeleteCmdWithConfirmer creates a delete command with a confirmer for testing.
func newTestDeleteCmdWithConfirmer(runner *mockDeleteRunner, confirmer *mockConfirmer, args ...string) (*cobra.Command, *bytes.Buffer) {
	cmd := NewDeleteCmd(runner, confirmer)
	if len(args) > 0 {
		cmd.SetArgs(args)
	}
	buf := new(bytes.Buffer)
	cmd.SetOut(buf)
	cmd.SetErr(new(bytes.Buffer))
	return cmd, buf
}

func TestDeleteCmd_HasForceFlag(t *testing.T) {
	runner := &mockDeleteRunner{result: &DeleteResult{}}
	confirmer := &mockConfirmer{}
	cmd, _ := newTestDeleteCmdWithConfirmer(runner, confirmer, "001-200")

	flag := cmd.Flags().Lookup("force")
	if flag == nil {
		t.Fatal("delete command should have --force flag")
	}
	if flag.DefValue != "false" {
		t.Errorf("--force default = %q, want %q", flag.DefValue, "false")
	}
}

func TestDeleteCmd_ConfirmationBehavior(t *testing.T) {
	tests := []struct {
		name             string
		force            bool
		confirmResult    bool
		confirmErr       error
		wantConfirmCall  bool
		wantRunnerCalled bool
		wantErr          bool
		wantErrContains  string
	}{
		{
			name:             "force skips confirmation",
			force:            true,
			wantConfirmCall:  false,
			wantRunnerCalled: true,
		},
		{
			name:             "no force user confirms proceeds",
			force:            false,
			confirmResult:    true,
			wantConfirmCall:  true,
			wantRunnerCalled: true,
		},
		{
			name:             "no force user denies aborts",
			force:            false,
			confirmResult:    false,
			wantConfirmCall:  true,
			wantRunnerCalled: false,
			wantErr:          true,
			wantErrContains:  "aborted",
		},
		{
			name:             "no force confirmer error propagates",
			force:            false,
			confirmErr:       fmt.Errorf("not a terminal"),
			wantConfirmCall:  true,
			wantRunnerCalled: false,
			wantErr:          true,
			wantErrContains:  "not a terminal",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			runner := &mockDeleteRunner{result: leafDeleteResult()}
			confirmer := &mockConfirmer{result: tt.confirmResult, err: tt.confirmErr}

			args := []string{"001-200"}
			if tt.force {
				args = append([]string{"--force"}, args...)
			}
			cmd, _ := newTestDeleteCmdWithConfirmer(runner, confirmer, args...)

			err := cmd.Execute()

			if (err != nil) != tt.wantErr {
				t.Errorf("error = %v, wantErr %v", err, tt.wantErr)
			}
			if tt.wantErrContains != "" && err != nil && !strings.Contains(err.Error(), tt.wantErrContains) {
				t.Errorf("error = %q, want to contain %q", err.Error(), tt.wantErrContains)
			}
			if runner.called != tt.wantRunnerCalled {
				t.Errorf("runner.called = %v, want %v", runner.called, tt.wantRunnerCalled)
			}
			if confirmer.called != tt.wantConfirmCall {
				t.Errorf("confirmer.called = %v, want %v", confirmer.called, tt.wantConfirmCall)
			}
		})
	}
}

func TestDeleteCmd_ConfirmationPromptIncludesSelector(t *testing.T) {
	runner := &mockDeleteRunner{result: leafDeleteResult()}
	confirmer := &mockConfirmer{result: true}
	cmd, _ := newTestDeleteCmdWithConfirmer(runner, confirmer, "001-200")

	err := cmd.Execute()

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !confirmer.called {
		t.Fatal("confirmer should be called without --force")
	}
	if !strings.Contains(confirmer.prompt, "001-200") {
		t.Errorf("prompt = %q, want to contain selector %q", confirmer.prompt, "001-200")
	}
}
