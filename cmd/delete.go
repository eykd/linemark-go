package cmd

import (
	"context"
	"fmt"

	"github.com/eykd/linemark-go/internal/domain"
	"github.com/spf13/cobra"
)

// DeleteResult holds the outcome of a delete operation.
type DeleteResult struct{}

// DeleteRunner defines the interface for running the delete operation.
type DeleteRunner interface {
	Delete(ctx context.Context, selector string, apply bool) (*DeleteResult, error)
}

// NewDeleteCmd creates the delete command with the given runner.
func NewDeleteCmd(runner DeleteRunner) *cobra.Command {
	cmd := &cobra.Command{
		Use:          "delete <selector>",
		Short:        "Delete a node from the outline",
		Args:         cobra.ExactArgs(1),
		SilenceUsage: true,
		RunE: func(cmd *cobra.Command, args []string) error {
			selector := args[0]
			if _, err := domain.ParseSelector(selector); err != nil {
				return fmt.Errorf("invalid selector %q: %w", selector, err)
			}

			isDryRun := GetDryRun()
			_, err := runner.Delete(cmd.Context(), selector, !isDryRun)
			return err
		},
	}

	return cmd
}
