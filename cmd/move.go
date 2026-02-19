package cmd

import (
	"context"
	"fmt"

	"github.com/eykd/linemark-go/internal/domain"
	"github.com/spf13/cobra"
)

// MoveResult holds the outcome of a move operation.
type MoveResult struct{}

// MoveRunner defines the interface for running the move operation.
type MoveRunner interface {
	Move(ctx context.Context, selector string, to string, apply bool) (*MoveResult, error)
}

// NewMoveCmd creates the move command with the given runner.
func NewMoveCmd(runner MoveRunner) *cobra.Command {
	var to string

	cmd := &cobra.Command{
		Use:          "move <selector>",
		Short:        "Move a node to a new position",
		Args:         cobra.ExactArgs(1),
		SilenceUsage: true,
		RunE: func(cmd *cobra.Command, args []string) error {
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
			_, err := runner.Move(cmd.Context(), selector, to, !isDryRun)
			return err
		},
	}

	cmd.Flags().StringVar(&to, "to", "", "Target position to move to")

	return cmd
}
