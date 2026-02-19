package cmd

import (
	"context"
	"testing"
)

func TestExecute(t *testing.T) {
	// Reset args to avoid test pollution
	rootCmd.SetArgs([]string{})

	err := Execute()
	if err != nil {
		t.Errorf("Execute() returned unexpected error: %v", err)
	}
}

func TestRootCommandUse(t *testing.T) {
	if rootCmd.Use != "lmk" {
		t.Errorf("rootCmd.Use = %q, want %q", rootCmd.Use, "lmk")
	}
}

func TestRootCommandShort(t *testing.T) {
	want := "Manage long-form prose projects with organized Markdown files"
	if rootCmd.Short != want {
		t.Errorf("rootCmd.Short = %q, want %q", rootCmd.Short, want)
	}
}

func TestRootCommandVerboseFlag(t *testing.T) {
	cmd := NewRootCmd()

	// Check that --verbose flag exists as a persistent flag
	verboseFlag := cmd.PersistentFlags().Lookup("verbose")
	if verboseFlag == nil {
		t.Fatal("expected --verbose persistent flag to exist")
	}

	// Check short flag -v exists
	vFlag := cmd.PersistentFlags().ShorthandLookup("v")
	if vFlag == nil {
		t.Fatal("expected -v shorthand for --verbose")
	}

	// Default should be false
	if verboseFlag.DefValue != "false" {
		t.Errorf("--verbose default = %q, want %q", verboseFlag.DefValue, "false")
	}
}

func TestGetVerbose(t *testing.T) {
	// Default should be false
	if GetVerbose() {
		t.Error("GetVerbose() should default to false")
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

	// ExecuteContext should still work (Cobra handles context gracefully)
	rootCmd.SetArgs([]string{})
	err := ExecuteContext(ctx)
	// A cancelled context may or may not produce an error depending on command
	// The important thing is it doesn't panic
	_ = err
}

func TestRootCommandJSONFlag(t *testing.T) {
	cmd := NewRootCmd()

	// Check that --json flag exists as a persistent flag
	jsonFlag := cmd.PersistentFlags().Lookup("json")
	if jsonFlag == nil {
		t.Fatal("expected --json persistent flag to exist")
	}

	// Default should be false
	if jsonFlag.DefValue != "false" {
		t.Errorf("--json default = %q, want %q", jsonFlag.DefValue, "false")
	}
}

func TestGetJSON(t *testing.T) {
	// Default should be false
	if GetJSON() {
		t.Error("GetJSON() should default to false")
	}
}
