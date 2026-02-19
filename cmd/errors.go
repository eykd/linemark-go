package cmd

import (
	"errors"
	"fmt"
	"io"

	"github.com/spf13/cobra"
)

// ContextError adds operation and path context to an underlying error.
type ContextError struct {
	Op   string
	Path string
	Err  error
}

// Error returns the formatted error string with context.
func (e *ContextError) Error() string {
	if e.Op != "" && e.Path != "" {
		return e.Op + ": " + e.Path + ": " + e.Err.Error()
	}
	if e.Op != "" {
		return e.Op + ": " + e.Err.Error()
	}
	if e.Path != "" {
		return e.Path + ": " + e.Err.Error()
	}
	return e.Err.Error()
}

// Unwrap returns the underlying error.
func (e *ContextError) Unwrap() error {
	return e.Err
}

// FormatError formats an error with the "lmk: " prefix and trailing newline.
func FormatError(err error) string {
	return fmt.Sprintf("lmk: %s\n", err.Error())
}

// ExitCoder is implemented by errors that carry a specific process exit code.
type ExitCoder interface {
	ExitCode() int
}

// ExitCodeFromError returns the appropriate exit code for an error.
// nil returns 0, ExitCoder errors return their code, all others return 1.
func ExitCodeFromError(err error) int {
	if err == nil {
		return 0
	}
	var coder ExitCoder
	if errors.As(err, &coder) {
		return coder.ExitCode()
	}
	return 1
}

// RunCLI executes the command with the given args, writing output to stdout
// and errors to stderr. It returns the appropriate exit code.
func RunCLI(cmd *cobra.Command, args []string, stdout io.Writer, stderr io.Writer) int {
	cmd.SetOut(stdout)
	cmd.SetErr(stderr)
	cmd.SetArgs(args)

	err := cmd.Execute()
	if err != nil {
		fmt.Fprint(stderr, FormatError(err))
		return ExitCodeFromError(err)
	}
	return 0
}
