package cmd

import (
	"context"
	"encoding/json"
	"fmt"

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

			result, err := runner.Compact(cmd.Context(), selector, applyFlag)
			if err != nil {
				return err
			}

			if jsonOutput || GetJSON() {
				return writeCompactJSON(cmd, result)
			}
			return writeCompactHuman(cmd, result)
		},
	}

	cmd.Flags().BoolVar(&applyFlag, "apply", false, "Execute the renumbering (default is report only)")
	cmd.Flags().BoolVar(&jsonOutput, "json", false, "Output results as JSON")

	return cmd
}

func writeCompactJSON(cmd *cobra.Command, result *CompactResult) error {
	enc := json.NewEncoder(cmd.OutOrStdout())
	return enc.Encode(result)
}

func writeCompactHuman(cmd *cobra.Command, result *CompactResult) error {
	w := cmd.OutOrStdout()
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
