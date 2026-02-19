package cmd

import (
	"context"
	"testing"

	"github.com/spf13/cobra"
)

func TestExecute(t *testing.T) {
	// Reset args to avoid test pollution
	rootCmd.SetArgs([]string{})

	err := Execute()
	if err != nil {
		t.Errorf("Execute() returned unexpected error: %v", err)
	}
}

func TestRootCmd_Metadata(t *testing.T) {
	tests := []struct {
		name string
		got  string
		want string
	}{
		{"Use", rootCmd.Use, "lmk"},
		{"Short", rootCmd.Short, "Manage long-form prose projects with organized Markdown files"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.got != tt.want {
				t.Errorf("got %q, want %q", tt.got, tt.want)
			}
		})
	}
}

func TestRootCmd_PersistentFlags(t *testing.T) {
	tests := []struct {
		name      string
		flag      string
		shorthand string
		defValue  string
	}{
		{"verbose", "verbose", "v", "false"},
		{"json", "json", "", "false"},
		{"dry-run", "dry-run", "", "false"},
	}

	cmd := NewRootCmd()
	for _, tt := range tests {
		t.Run(tt.flag, func(t *testing.T) {
			f := cmd.PersistentFlags().Lookup(tt.flag)
			if f == nil {
				t.Fatalf("expected --%s persistent flag to exist", tt.flag)
			}
			if f.DefValue != tt.defValue {
				t.Errorf("--%s default = %q, want %q", tt.flag, f.DefValue, tt.defValue)
			}
			if tt.shorthand != "" {
				sf := cmd.PersistentFlags().ShorthandLookup(tt.shorthand)
				if sf == nil {
					t.Fatalf("expected -%s shorthand for --%s", tt.shorthand, tt.flag)
				}
			}
		})
	}
}

func TestExecuteContext(t *testing.T) {
	// Reset args to avoid test pollution
	rootCmd.SetArgs([]string{})

	ctx := context.Background()
	err := ExecuteContext(ctx)
	if err != nil {
		t.Errorf("ExecuteContext() returned unexpected error: %v", err)
	}
}

func TestExecuteContext_WithCancelledContext(t *testing.T) {
	// Create a cancelled context
	ctx, cancel := context.WithCancel(context.Background())
	cancel() // Cancel immediately

	// ExecuteContext should not panic with a cancelled context.
	// An error may or may not occur depending on command timing.
	rootCmd.SetArgs([]string{})
	if err := ExecuteContext(ctx); err != nil {
		t.Logf("ExecuteContext with cancelled context returned error (acceptable): %v", err)
	}
}

func TestRootCmd_GetterDefaults(t *testing.T) {
	tests := []struct {
		name   string
		getter func() bool
	}{
		{"GetVerbose", GetVerbose},
		{"GetJSON", GetJSON},
		{"GetDryRun", GetDryRun},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.getter() {
				t.Errorf("%s() should default to false", tt.name)
			}
		})
	}
}

func TestRootCmd_GlobalFlagParsing(t *testing.T) {
	tests := []struct {
		name  string
		args  []string
		check func() bool
		want  bool
	}{
		{"--verbose sets GetVerbose true", []string{"--verbose"}, GetVerbose, true},
		{"--json sets GetJSON true", []string{"--json"}, GetJSON, true},
		{"--dry-run sets GetDryRun true", []string{"--dry-run"}, GetDryRun, true},
		{"-v shorthand sets GetVerbose true", []string{"-v"}, GetVerbose, true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cmd := NewRootCmd()
			cmd.SetArgs(tt.args)
			if err := cmd.Execute(); err != nil {
				t.Fatalf("Execute() error: %v", err)
			}
			if got := tt.check(); got != tt.want {
				t.Errorf("getter() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestRootCmd_SilenceUsage(t *testing.T) {
	cmd := NewRootCmd()
	if !cmd.SilenceUsage {
		t.Error("expected SilenceUsage to be true to prevent printing usage on errors")
	}
}

func TestExecuteContext_ContextPropagation(t *testing.T) {
	type ctxKey struct{}
	ctx := context.WithValue(context.Background(), ctxKey{}, "test-value")

	var capturedCtx context.Context
	cmd := NewRootCmd()
	sub := &cobra.Command{
		Use: "testctx",
		RunE: func(cmd *cobra.Command, args []string) error {
			capturedCtx = cmd.Context()
			return nil
		},
	}
	cmd.AddCommand(sub)
	cmd.SetArgs([]string{"testctx"})

	if err := cmd.ExecuteContext(ctx); err != nil {
		t.Fatalf("ExecuteContext() error: %v", err)
	}
	if capturedCtx == nil {
		t.Fatal("subcommand did not receive context")
	}
	if got := capturedCtx.Value(ctxKey{}); got != "test-value" {
		t.Errorf("context value = %v, want %q", got, "test-value")
	}
}
