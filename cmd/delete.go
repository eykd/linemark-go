package cmd

import (
	"context"
	"fmt"

	"github.com/eykd/linemark-go/internal/domain"
	"github.com/spf13/cobra"
)

// DeleteResult holds the outcome of a delete operation.
type DeleteResult struct {
	FilesDeleted  []string          `json:"files_deleted"`
	FilesRenamed  map[string]string `json:"files_renamed,omitempty"`
	SIDsPreserved []string          `json:"sids_preserved"`
	Planned       bool              `json:"planned"`
}

// DeleteRunner defines the interface for running the delete operation.
type DeleteRunner interface {
	Delete(ctx context.Context, selector string, mode domain.DeleteMode, apply bool) (*DeleteResult, error)
}

// Confirmer defines the interface for confirming destructive operations.
type Confirmer interface {
	Confirm(prompt string) (bool, error)
}

// NewDeleteCmd creates the delete command with the given runner.
// An optional Confirmer may be passed to prompt for confirmation before deleting.
func NewDeleteCmd(runner DeleteRunner, confirmers ...Confirmer) *cobra.Command {
	var confirmer Confirmer
	if len(confirmers) > 0 {
		confirmer = confirmers[0]
	}

	var jsonOutput bool
	var recursive bool
	var promote bool
	var force bool

	cmd := &cobra.Command{
		Use:          "delete <selector>",
		Short:        "Delete a node from the outline",
		Args:         cobra.ExactArgs(1),
		SilenceUsage: true,
		RunE: func(cmd *cobra.Command, args []string) error {
			if recursive && promote {
				return fmt.Errorf("--recursive and --promote are mutually exclusive")
			}

			selector := args[0]
			if _, err := domain.ParseSelector(selector); err != nil {
				return fmt.Errorf("invalid selector %q: %w", selector, err)
			}

			if confirmer != nil && !force {
				confirmed, err := confirmer.Confirm(fmt.Sprintf("Delete %s?", selector))
				if err != nil {
					return err
				}
				if !confirmed {
					return fmt.Errorf("aborted")
				}
			}

			mode := domain.DeleteModeDefault
			if recursive {
				mode = domain.DeleteModeRecursive
			} else if promote {
				mode = domain.DeleteModePromote
			}

			isDryRun := GetDryRun()
			result, err := runner.Delete(cmd.Context(), selector, mode, !isDryRun)
			if err != nil {
				return err
			}

			if isDryRun {
				result.Planned = true
			}

			if jsonOutput || GetJSON() {
				writeJSON(cmd.OutOrStdout(), result)
			} else {
				for _, f := range result.FilesDeleted {
					fmt.Fprintln(cmd.OutOrStdout(), f)
				}
				for _, newName := range result.FilesRenamed {
					fmt.Fprintln(cmd.OutOrStdout(), newName)
				}
			}
			return nil
		},
	}

	cmd.Flags().BoolVar(&jsonOutput, "json", false, "Output results as JSON")
	cmd.Flags().BoolVarP(&recursive, "recursive", "r", false, "Delete node and entire subtree")
	cmd.Flags().BoolVarP(&promote, "promote", "p", false, "Delete node and promote children")
	cmd.Flags().BoolVar(&force, "force", false, "Skip confirmation prompt")

	return cmd
}
