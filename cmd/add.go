package cmd

import (
	"context"
	"fmt"

	"github.com/eykd/linemark-go/internal/domain"
	"github.com/spf13/cobra"
)

// AddNodeInfo holds node identification details for add results.
type AddNodeInfo struct {
	MP    string `json:"mp"`
	SID   string `json:"sid"`
	Title string `json:"title"`
}

// AddResult holds the outcome of an add operation.
type AddResult struct {
	Node         AddNodeInfo `json:"node"`
	FilesCreated []string    `json:"files_created"`
	FilesPlanned []string    `json:"files_planned"`
	Planned      bool        `json:"planned"`
}

// Placement holds options for positioning a new node relative to existing nodes.
type Placement struct {
	ChildOf   string
	SiblingOf string
	Before    string
	After     string
}

// AddRunner defines the interface for running the add operation.
type AddRunner interface {
	Add(ctx context.Context, title string, apply bool, placement Placement) (*AddResult, error)
}

// NewAddCmd creates the add command with the given runner.
func NewAddCmd(runner AddRunner) *cobra.Command {
	var jsonOutput bool
	var childOf string
	var siblingOf string
	var before string
	var after string

	cmd := &cobra.Command{
		Use:          "add <title>",
		Short:        "Add a new node to the outline",
		Args:         cobra.ExactArgs(1),
		SilenceUsage: true,
		RunE: func(cmd *cobra.Command, args []string) error {
			if runner == nil {
				return ErrNotInProject
			}
			if childOf != "" && siblingOf != "" {
				return fmt.Errorf("--child-of and --sibling-of are mutually exclusive")
			}
			if before != "" && after != "" {
				return fmt.Errorf("--before and --after are mutually exclusive")
			}

			for _, pair := range []struct {
				flag  string
				value string
			}{
				{"--child-of", childOf},
				{"--sibling-of", siblingOf},
				{"--before", before},
				{"--after", after},
			} {
				if pair.value != "" {
					if _, err := domain.ParseSelector(pair.value); err != nil {
						return fmt.Errorf("invalid selector for %s %q: %w", pair.flag, pair.value, err)
					}
				}
			}

			isDryRun := GetDryRun()
			placement := Placement{
				ChildOf:   childOf,
				SiblingOf: siblingOf,
				Before:    before,
				After:     after,
			}
			result, err := runner.Add(cmd.Context(), args[0], !isDryRun, placement)
			if err != nil {
				return err
			}

			if isDryRun {
				result.Planned = true
			}

			if jsonOutput || GetJSON() {
				writeJSON(cmd.OutOrStdout(), result)
			} else if isDryRun {
				for _, f := range result.FilesPlanned {
					fmt.Fprintf(cmd.OutOrStdout(), "Would create %s\n", f)
				}
			} else {
				for _, f := range result.FilesCreated {
					fmt.Fprintf(cmd.OutOrStdout(), "Created %s\n", f)
				}
			}
			return nil
		},
	}

	cmd.Flags().BoolVar(&jsonOutput, "json", false, "Output results as JSON")
	cmd.Flags().StringVar(&childOf, "child-of", "", "Add as last child of the specified node")
	cmd.Flags().StringVar(&siblingOf, "sibling-of", "", "Add immediately after the specified node")
	cmd.Flags().StringVar(&before, "before", "", "Insert before the specified sibling node")
	cmd.Flags().StringVar(&after, "after", "", "Insert after the specified sibling node")

	return cmd
}
