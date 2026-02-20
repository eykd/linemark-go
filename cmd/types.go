package cmd

import (
	"context"
	"fmt"

	"github.com/spf13/cobra"
)

// NodeInfo holds node identification details.
type NodeInfo struct {
	MP  string `json:"mp"`
	SID string `json:"sid"`
}

// TypesListResult holds the result of listing types for a node.
type TypesListResult struct {
	Node  NodeInfo `json:"node"`
	Types []string `json:"types"`
}

// TypesModifyResult holds the result of adding or removing a type.
type TypesModifyResult struct {
	Node     NodeInfo `json:"node"`
	Filename string   `json:"filename"`
	Planned  bool     `json:"planned"`
}

// TypesService defines the interface for managing document types.
type TypesService interface {
	ListTypes(ctx context.Context, selector string) (*TypesListResult, error)
	AddType(ctx context.Context, docType, selector string, apply bool) (*TypesModifyResult, error)
	RemoveType(ctx context.Context, docType, selector string, apply bool) (*TypesModifyResult, error)
}

// NewTypesCmd creates the types command with the given service.
func NewTypesCmd(svc TypesService) *cobra.Command {
	cmd := &cobra.Command{
		Use:          "types",
		Short:        "Manage document types for a node",
		SilenceUsage: true,
	}

	cmd.AddCommand(newTypesListCmd(svc))
	cmd.AddCommand(newTypesAddCmd(svc))
	cmd.AddCommand(newTypesRemoveCmd(svc))

	return cmd
}

func newTypesListCmd(svc TypesService) *cobra.Command {
	var jsonOutput bool

	cmd := &cobra.Command{
		Use:          "list <selector>",
		Short:        "List document types for a node",
		Args:         cobra.ExactArgs(1),
		SilenceUsage: true,
		RunE: func(cmd *cobra.Command, args []string) error {
			if svc == nil {
				return ErrNotInProject
			}
			result, err := svc.ListTypes(cmd.Context(), args[0])
			if err != nil {
				return err
			}

			if jsonOutput || GetJSON() {
				writeJSON(cmd.OutOrStdout(), result)
			} else {
				for _, t := range result.Types {
					fmt.Fprintln(cmd.OutOrStdout(), t)
				}
			}
			return nil
		},
	}

	cmd.Flags().BoolVar(&jsonOutput, "json", false, "Output results as JSON")

	return cmd
}

func newTypesAddCmd(svc TypesService) *cobra.Command {
	var jsonOutput bool

	cmd := &cobra.Command{
		Use:          "add <type> <selector>",
		Short:        "Add a document type to a node",
		Args:         cobra.ExactArgs(2),
		SilenceUsage: true,
		RunE: func(cmd *cobra.Command, args []string) error {
			if svc == nil {
				return ErrNotInProject
			}
			isDryRun := GetDryRun()
			result, err := svc.AddType(cmd.Context(), args[0], args[1], !isDryRun)
			if err != nil {
				return err
			}

			if isDryRun {
				result.Planned = true
			}

			if jsonOutput || GetJSON() {
				writeJSON(cmd.OutOrStdout(), result)
			} else if isDryRun {
				fmt.Fprintf(cmd.OutOrStdout(), "Would add %s\n", result.Filename)
			} else {
				fmt.Fprintf(cmd.OutOrStdout(), "Added %s\n", result.Filename)
			}
			return nil
		},
	}

	cmd.Flags().BoolVar(&jsonOutput, "json", false, "Output results as JSON")

	return cmd
}

func newTypesRemoveCmd(svc TypesService) *cobra.Command {
	var jsonOutput bool

	cmd := &cobra.Command{
		Use:          "remove <type> <selector>",
		Short:        "Remove a document type from a node",
		Args:         cobra.ExactArgs(2),
		SilenceUsage: true,
		RunE: func(cmd *cobra.Command, args []string) error {
			if svc == nil {
				return ErrNotInProject
			}
			isDryRun := GetDryRun()
			result, err := svc.RemoveType(cmd.Context(), args[0], args[1], !isDryRun)
			if err != nil {
				return err
			}

			if isDryRun {
				result.Planned = true
			}

			if jsonOutput || GetJSON() {
				writeJSON(cmd.OutOrStdout(), result)
			} else if isDryRun {
				fmt.Fprintf(cmd.OutOrStdout(), "Would remove %s\n", result.Filename)
			} else {
				fmt.Fprintf(cmd.OutOrStdout(), "Removed %s\n", result.Filename)
			}
			return nil
		},
	}

	cmd.Flags().BoolVar(&jsonOutput, "json", false, "Output results as JSON")

	return cmd
}
