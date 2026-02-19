package cmd

import (
	"context"
	"fmt"

	"github.com/eykd/linemark-go/internal/domain"
	"github.com/spf13/cobra"
)

// MoveResult holds the outcome of a move operation.
type MoveResult struct {
	Renames []RenameEntry `json:"renames"`
	Planned bool          `json:"planned"`
}

// MoveRunner defines the interface for running the move operation.
type MoveRunner interface {
	Move(ctx context.Context, selector string, to string, apply bool) (*MoveResult, error)
}

// NewMoveCmd creates the move command with the given runner.
func NewMoveCmd(runner MoveRunner) *cobra.Command {
	var to string
	var jsonOutput bool
	var before string
	var after string

	cmd := &cobra.Command{
		Use:          "move <selector>",
		Short:        "Move a node to a new position",
		Args:         cobra.ExactArgs(1),
		SilenceUsage: true,
		RunE: func(cmd *cobra.Command, args []string) error {
			if before != "" && after != "" {
				return fmt.Errorf("--before and --after are mutually exclusive")
			}

			selector := args[0]
			if _, err := domain.ParseSelector(selector); err != nil {
				return fmt.Errorf("invalid selector %q: %w", selector, err)
			}

			if to == "" {
				return fmt.Errorf("required flag \"to\" not set")
			}
			if _, err := domain.ParseSelector(to); err != nil {
				return fmt.Errorf("invalid target selector %q: %w", to, err)
			}

			isDryRun := GetDryRun()
			result, err := runner.Move(cmd.Context(), selector, to, !isDryRun)
			if err != nil {
				return err
			}

			if isDryRun {
				result.Planned = true
			}

			if jsonOutput || GetJSON() {
				writeJSON(cmd.OutOrStdout(), result)
			} else {
				for _, r := range result.Renames {
					fmt.Fprintf(cmd.OutOrStdout(), "  %s -> %s\n", r.Old, r.New)
				}
			}
			return nil
		},
	}

	cmd.Flags().StringVar(&to, "to", "", "Target position to move to")
	cmd.Flags().BoolVar(&jsonOutput, "json", false, "Output results as JSON")
	cmd.Flags().StringVar(&before, "before", "", "Place before this sibling")
	cmd.Flags().StringVar(&after, "after", "", "Place after this sibling")

	return cmd
}
