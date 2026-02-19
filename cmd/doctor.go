package cmd

import (
	"github.com/spf13/cobra"
)

// NewDoctorCmd creates the doctor command with the given runner.
func NewDoctorCmd(runner CheckRunner) *cobra.Command {
	var applyFlag bool
	var jsonOutput bool

	cmd := &cobra.Command{
		Use:          "doctor",
		Short:        "Diagnose and fix project issues",
		SilenceUsage: true,
		RunE: func(cmd *cobra.Command, args []string) error {
			result, err := runner.Check(cmd.Context())
			if err != nil {
				return err
			}

			errCount, warnCount := countBySeverity(result.Findings)

			if jsonOutput {
				formatCheckJSON(cmd.OutOrStdout(), result.Findings, errCount, warnCount)
			} else {
				formatCheckHuman(cmd.OutOrStdout(), result.Findings, errCount, warnCount)
			}

			if len(result.Findings) > 0 {
				return &FindingsDetectedError{Errors: errCount, Warnings: warnCount}
			}
			return nil
		},
	}

	cmd.Flags().BoolVar(&applyFlag, "apply", false, "Apply automatic fixes")
	cmd.Flags().BoolVar(&jsonOutput, "json", false, "Output results as JSON")

	return cmd
}

