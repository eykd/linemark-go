package cmd

import (
	"bytes"
	"errors"
	"strings"
	"testing"

	"github.com/spf13/cobra"
)

func TestContextError_Error(t *testing.T) {
	tests := []struct {
		name string
		err  *ContextError
		want string
	}{
		{
			name: "op and path",
			err:  &ContextError{Op: "read", Path: "/foo/bar.md", Err: errors.New("permission denied")},
			want: "read: /foo/bar.md: permission denied",
		},
		{
			name: "op only",
			err:  &ContextError{Op: "compact", Err: errors.New("invalid selector")},
			want: "compact: invalid selector",
		},
		{
			name: "path only",
			err:  &ContextError{Path: "/foo/bar.md", Err: errors.New("not found")},
			want: "/foo/bar.md: not found",
		},
		{
			name: "error only",
			err:  &ContextError{Err: errors.New("unknown error")},
			want: "unknown error",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := tt.err.Error()
			if got != tt.want {
				t.Errorf("Error() = %q, want %q", got, tt.want)
			}
		})
	}
}

func TestContextError_Unwrap(t *testing.T) {
	inner := errors.New("inner error")
	err := &ContextError{Op: "read", Err: inner}

	if !errors.Is(err, inner) {
		t.Error("ContextError should unwrap to inner error")
	}
}

func TestContextError_Unwrap_ExitCoder(t *testing.T) {
	inner := &FindingsDetectedError{Errors: 2, Warnings: 1}
	err := &ContextError{Op: "check", Err: inner}

	// ExitCodeFromError should find the wrapped ExitCoder
	code := ExitCodeFromError(err)
	if code != 2 {
		t.Errorf("ExitCodeFromError(ContextError wrapping FindingsDetectedError) = %d, want 2", code)
	}
}

func TestFormatError(t *testing.T) {
	tests := []struct {
		name string
		err  error
		want string
	}{
		{
			name: "simple error",
			err:  errors.New("something failed"),
			want: "lmk: something failed\n",
		},
		{
			name: "context error with op and path",
			err:  &ContextError{Op: "read", Path: "/foo/bar.md", Err: errors.New("permission denied")},
			want: "lmk: read: /foo/bar.md: permission denied\n",
		},
		{
			name: "context error with op only",
			err:  &ContextError{Op: "compact", Err: errors.New("invalid selector")},
			want: "lmk: compact: invalid selector\n",
		},
		{
			name: "context error with path only",
			err:  &ContextError{Path: "/foo/bar.md", Err: errors.New("not found")},
			want: "lmk: /foo/bar.md: not found\n",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := FormatError(tt.err)
			if got != tt.want {
				t.Errorf("FormatError() = %q, want %q", got, tt.want)
			}
		})
	}
}

func TestRunCLI_ExitCodes(t *testing.T) {
	tests := []struct {
		name     string
		runErr   error
		wantCode int
	}{
		{
			name:     "nil error returns 0",
			runErr:   nil,
			wantCode: 0,
		},
		{
			name:     "generic error returns 1",
			runErr:   errors.New("something went wrong"),
			wantCode: 1,
		},
		{
			name:     "findings detected returns 2",
			runErr:   &FindingsDetectedError{Errors: 1, Warnings: 0},
			wantCode: 2,
		},
		{
			name:     "unrepaired error returns 2",
			runErr:   &UnrepairedError{Count: 3},
			wantCode: 2,
		},
		{
			name:     "context error wrapping generic returns 1",
			runErr:   &ContextError{Op: "read", Err: errors.New("fail")},
			wantCode: 1,
		},
		{
			name:     "context error wrapping findings returns 2",
			runErr:   &ContextError{Op: "check", Err: &FindingsDetectedError{Errors: 1}},
			wantCode: 2,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cmd := &cobra.Command{
				Use:          "test",
				SilenceUsage: true,
				RunE: func(cmd *cobra.Command, args []string) error {
					return tt.runErr
				},
			}

			stdout := new(bytes.Buffer)
			stderr := new(bytes.Buffer)

			got := RunCLI(cmd, []string{}, stdout, stderr)
			if got != tt.wantCode {
				t.Errorf("RunCLI() exit code = %d, want %d", got, tt.wantCode)
			}
		})
	}
}

func TestRunCLI_ErrorsWrittenToStderr(t *testing.T) {
	cmd := &cobra.Command{
		Use:          "test",
		SilenceUsage: true,
		RunE: func(cmd *cobra.Command, args []string) error {
			return &ContextError{Op: "read", Path: "/foo/bar.md", Err: errors.New("permission denied")}
		},
	}

	stdout := new(bytes.Buffer)
	stderr := new(bytes.Buffer)

	RunCLI(cmd, []string{}, stdout, stderr)

	// Error message should appear in stderr with lmk prefix
	if !strings.Contains(stderr.String(), "lmk:") {
		t.Errorf("expected 'lmk:' prefix in stderr, got: %q", stderr.String())
	}
	if !strings.Contains(stderr.String(), "permission denied") {
		t.Errorf("expected error message in stderr, got: %q", stderr.String())
	}
	if !strings.Contains(stderr.String(), "/foo/bar.md") {
		t.Errorf("expected file path in stderr, got: %q", stderr.String())
	}

	// Error should NOT appear in stdout
	if strings.Contains(stdout.String(), "permission denied") {
		t.Errorf("error should not appear in stdout, got: %q", stdout.String())
	}
}

func TestRunCLI_NoStderrOnSuccess(t *testing.T) {
	cmd := &cobra.Command{
		Use: "test",
		RunE: func(cmd *cobra.Command, args []string) error {
			return nil
		},
	}

	stdout := new(bytes.Buffer)
	stderr := new(bytes.Buffer)

	code := RunCLI(cmd, []string{}, stdout, stderr)

	if code != 0 {
		t.Errorf("exit code = %d, want 0", code)
	}
	if stderr.Len() > 0 {
		t.Errorf("stderr should be empty on success, got: %q", stderr.String())
	}
}

func TestRootCmd_SilenceErrors(t *testing.T) {
	cmd := NewRootCmd()
	if !cmd.SilenceErrors {
		t.Error("root command should have SilenceErrors = true for consistent error handling")
	}
}
