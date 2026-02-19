package cmd

import (
	"fmt"
	"io"

	"github.com/spf13/cobra"
)

// NewDoctorCmd creates the doctor command with the given runners.
func NewDoctorCmd(runner CheckRunner, repairers ...RepairRunner) *cobra.Command {
	var applyFlag bool
	var jsonOutput bool

	var repairer RepairRunner
	if len(repairers) > 0 {
		repairer = repairers[0]
	}

	cmd := &cobra.Command{
		Use:          "doctor",
		Short:        "Diagnose and fix project issues",
		SilenceUsage: true,
		RunE: func(cmd *cobra.Command, args []string) error {
			if applyFlag {
				return runRepairAndReport(cmd, repairer, jsonOutput || GetJSON())
			}
			return runCheckAndReport(cmd, runner, jsonOutput || GetJSON())
		},
	}

	cmd.Flags().BoolVar(&applyFlag, "apply", false, "Apply automatic fixes")
	cmd.Flags().BoolVar(&jsonOutput, "json", false, "Output results as JSON")

	return cmd
}

// repairJSONResponse is the JSON output structure for the repair command.
type repairJSONResponse struct {
	Repairs    []RepairAction `json:"repairs"`
	Unrepaired []CheckFinding `json:"unrepaired"`
	Summary    struct {
		Repaired   int `json:"repaired"`
		Unrepaired int `json:"unrepaired"`
	} `json:"summary"`
}

// formatRepairJSON writes repair results as JSON to w.
func formatRepairJSON(w io.Writer, repairs []RepairAction, unrepaired []CheckFinding) {
	if repairs == nil {
		repairs = []RepairAction{}
	}
	if unrepaired == nil {
		unrepaired = []CheckFinding{}
	}
	out := repairJSONResponse{
		Repairs:    repairs,
		Unrepaired: unrepaired,
	}
	out.Summary.Repaired = len(repairs)
	out.Summary.Unrepaired = len(unrepaired)
	writeJSONImpl(w, out)
}

// formatRepairHuman writes repair results as human-readable text to w.
func formatRepairHuman(w io.Writer, repairs []RepairAction, unrepaired []CheckFinding) {
	for _, r := range repairs {
		fmt.Fprintf(w, "%s [%s] %s -> %s\n", r.Type, r.Action, r.Old, r.New)
	}
	for _, f := range unrepaired {
		fmt.Fprintf(w, "%s [%s] %s: %s\n", f.Path, f.Severity, f.Type, f.Message)
	}
	repaired := len(repairs)
	unrepairedCount := len(unrepaired)
	if repaired > 0 || unrepairedCount > 0 {
		fmt.Fprintf(w, "\n%d repaired, %d unrepaired\n", repaired, unrepairedCount)
	}
}

// runRepairAndReport runs the repairer and formats results as JSON or human-readable text.
func runRepairAndReport(cmd *cobra.Command, repairer RepairRunner, jsonOutput bool) error {
	result, err := repairer.Repair(cmd.Context())
	if err != nil {
		return err
	}

	if jsonOutput {
		formatRepairJSON(cmd.OutOrStdout(), result.Repairs, result.Unrepaired)
	} else {
		formatRepairHuman(cmd.OutOrStdout(), result.Repairs, result.Unrepaired)
	}

	if len(result.Unrepaired) > 0 {
		return &UnrepairedError{Count: len(result.Unrepaired)}
	}
	return nil
}
