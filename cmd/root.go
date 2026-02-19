// Package cmd contains the CLI commands for the lmk application.
package cmd

import (
	"context"

	"github.com/spf13/cobra"
)

var rootCmd *cobra.Command

// verbose holds the global --verbose flag state.
var verbose bool

// jsonFlag holds the global --json flag state.
var jsonFlag bool

// dryRun holds the global --dry-run flag state.
var dryRun bool

func init() {
	rootCmd = NewRootCmd()
	rootCmd.AddCommand(NewCheckCmd(nil))
	rootCmd.AddCommand(NewDoctorCmd(nil))
	rootCmd.AddCommand(NewTypesCmd(nil))
	rootCmd.AddCommand(NewCompactCmd(nil))
}

// GetVerbose returns the current verbose flag state.
// This is used by other packages to check if debug logging is enabled.
func GetVerbose() bool {
	return verbose
}

// GetJSON returns the current global --json flag state.
func GetJSON() bool {
	return jsonFlag
}

// GetDryRun returns the current global --dry-run flag state.
func GetDryRun() bool {
	return dryRun
}

// NewRootCmd creates a new root command instance.
// This is useful for testing to get a fresh command tree.
func NewRootCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:           "lmk",
		Short:         "Manage long-form prose projects with organized Markdown files",
		Long:          "lmk is a CLI tool for managing long-form prose projects using organized Markdown files.",
		SilenceErrors: true,
		SilenceUsage:  true,
	}

	// Add persistent flags (available to all subcommands)
	cmd.PersistentFlags().BoolVarP(&verbose, "verbose", "v", false, "Enable debug logging to stderr")
	cmd.PersistentFlags().BoolVar(&jsonFlag, "json", false, "Output results as JSON")
	cmd.PersistentFlags().BoolVar(&dryRun, "dry-run", false, "Preview changes without modifying files")

	return cmd
}

// Execute runs the root command and returns any error.
// Deprecated: Use ExecuteContext instead for proper signal handling.
func Execute() error {
	return rootCmd.Execute()
}

// ExecuteContext runs the root command with the given context.
// This enables graceful shutdown via context cancellation (e.g., on SIGINT).
func ExecuteContext(ctx context.Context) error {
	return rootCmd.ExecuteContext(ctx)
}
