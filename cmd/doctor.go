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
			return runCheckAndReport(cmd, runner, jsonOutput)
		},
	}

	cmd.Flags().BoolVar(&applyFlag, "apply", false, "Apply automatic fixes")
	cmd.Flags().BoolVar(&jsonOutput, "json", false, "Output results as JSON")

	return cmd
}
