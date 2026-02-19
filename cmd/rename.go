package cmd

import (
	"context"
	"fmt"

	"github.com/eykd/linemark-go/internal/domain"
	"github.com/spf13/cobra"
)

// RenameResult holds the outcome of a rename operation.
type RenameResult struct{}

// RenameRunner defines the interface for running the rename operation.
type RenameRunner interface {
	Rename(ctx context.Context, selector string, newTitle string, apply bool) (*RenameResult, error)
}

// NewRenameCmd creates the rename command with the given runner.
func NewRenameCmd(runner RenameRunner) *cobra.Command {
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
			_, err := runner.Rename(cmd.Context(), selector, args[1], !isDryRun)
			return err
		},
	}

	return cmd
}
