package cmd

import (
	"context"
	"encoding/json"
	"fmt"
	"io"

	"github.com/spf13/cobra"
)

// RenameEntry represents a single file rename operation.
type RenameEntry struct {
	Old string `json:"old"`
	New string `json:"new"`
}

// CompactResult holds the outcome of a compact operation.
type CompactResult struct {
	Renames       []RenameEntry `json:"renames"`
	FilesAffected int           `json:"files_affected"`
	Warning       *string       `json:"warning"`
	Planned       bool          `json:"planned"`
}

// CompactRunner executes the compact operation.
type CompactRunner interface {
	Compact(ctx context.Context, selector string, apply bool) (*CompactResult, error)
}

// NewCompactCmd creates the compact command with the given runner.
func NewCompactCmd(runner CompactRunner) *cobra.Command {
	var applyFlag bool
	var jsonOutput bool

	cmd := &cobra.Command{
		Use:          "compact [selector]",
		Short:        "Renumber outline nodes to consistent spacing",
		SilenceUsage: true,
		Args:         cobra.MaximumNArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			var selector string
			if len(args) > 0 {
				selector = args[0]
			}

			apply := applyFlag
			if GetDryRun() {
				apply = false
			}

			result, err := runner.Compact(cmd.Context(), selector, apply)
			if err != nil {
				return err
			}

			if GetDryRun() {
				result.Planned = true
			}

			if jsonOutput || GetJSON() {
				return writeCompactJSON(cmd.OutOrStdout(), result)
			}
			return writeCompactHuman(cmd.OutOrStdout(), result)
		},
	}

	cmd.Flags().BoolVar(&applyFlag, "apply", false, "Execute the renumbering (default is report only)")
	cmd.Flags().BoolVar(&jsonOutput, "json", false, "Output results as JSON")

	return cmd
}

func writeCompactJSON(w io.Writer, result *CompactResult) error {
	return json.NewEncoder(w).Encode(result)
}

func writeCompactHuman(w io.Writer, result *CompactResult) error {
	for _, r := range result.Renames {
		fmt.Fprintf(w, "  %s -> %s\n", r.Old, r.New)
	}
	if result.Warning != nil {
		fmt.Fprintf(w, "Warning: %s\n", *result.Warning)
	}
	if result.FilesAffected > 0 {
		fmt.Fprintf(w, "%d file(s) affected\n", result.FilesAffected)
	}
	return nil
}
