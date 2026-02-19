package cmd

import (
	"context"
	"fmt"

	"github.com/eykd/linemark-go/internal/domain"
	"github.com/spf13/cobra"
)

// RenameNodeInfo holds node identification details for rename results.
type RenameNodeInfo struct {
	MP       string `json:"mp"`
	SID      string `json:"sid"`
	OldTitle string `json:"old_title"`
	NewTitle string `json:"new_title"`
}

// RenameResult holds the outcome of a rename operation.
type RenameResult struct {
	Node    RenameNodeInfo `json:"node"`
	Renames []RenameEntry  `json:"renames"`
	Planned bool           `json:"planned"`
}

// RenameRunner defines the interface for running the rename operation.
type RenameRunner interface {
	Rename(ctx context.Context, selector string, newTitle string, apply bool) (*RenameResult, error)
}

// NewRenameCmd creates the rename command with the given runner.
func NewRenameCmd(runner RenameRunner) *cobra.Command {
	var jsonOutput bool

	cmd := &cobra.Command{
		Use:          "rename <selector> <new-title>",
		Short:        "Rename a node",
		Args:         cobra.ExactArgs(2),
		SilenceUsage: true,
		RunE: func(cmd *cobra.Command, args []string) error {
			selector := args[0]
			if _, err := domain.ParseSelector(selector); err != nil {
				return fmt.Errorf("invalid selector %q: %w", selector, err)
			}

			isDryRun := GetDryRun()
			result, err := runner.Rename(cmd.Context(), selector, args[1], !isDryRun)
			if err != nil {
				return err
			}

			if isDryRun {
				result.Planned = true
			}

			if jsonOutput || GetJSON() {
				writeJSON(cmd.OutOrStdout(), result)
			} else {
				fmt.Fprintf(cmd.OutOrStdout(), "Renamed %q to %q\n", result.Node.OldTitle, result.Node.NewTitle)
				for _, r := range result.Renames {
					fmt.Fprintf(cmd.OutOrStdout(), "  %s -> %s\n", r.Old, r.New)
				}
			}
			return nil
		},
	}

	cmd.Flags().BoolVar(&jsonOutput, "json", false, "Output results as JSON")

	return cmd
}
