package cmd

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"

	"github.com/spf13/cobra"
)

// FindingType represents the kind of check finding.
type FindingType string

const (
	// FindingInvalidFilename indicates a file does not match the canonical filename pattern.
	FindingInvalidFilename FindingType = "invalid_filename"
	// FindingDuplicateSID indicates a SID is used by multiple nodes.
	FindingDuplicateSID FindingType = "duplicate_sid"
	// FindingSlugDrift indicates a filename slug does not match the title slug.
	FindingSlugDrift FindingType = "slug_drift"
	// FindingMissingDraft indicates a node has no draft document.
	FindingMissingDraft FindingType = "missing_draft"
	// FindingMissingNotes indicates a node has no notes document.
	FindingMissingNotes FindingType = "missing_notes"
	// FindingMalformedFrontmatter indicates YAML frontmatter cannot be parsed.
	FindingMalformedFrontmatter FindingType = "malformed_frontmatter"
	// FindingOrphanedReservation indicates a SID reservation has no content files.
	FindingOrphanedReservation FindingType = "orphaned_reservation"
	// FindingUnreservedSID indicates a SID is in use but not reserved.
	FindingUnreservedSID FindingType = "unreserved_sid"
)

// Severity represents the severity level of a check finding.
type Severity string

const (
	// SeverityError represents an error-level finding.
	SeverityError Severity = "error"
	// SeverityWarning represents a warning-level finding.
	SeverityWarning Severity = "warning"
)

// CheckFinding represents a single finding from the check command.
type CheckFinding struct {
	Type     FindingType `json:"type"`
	Severity Severity    `json:"severity"`
	Message  string      `json:"message"`
	Path     string      `json:"path"`
}

// CheckResult holds all findings from a check run.
type CheckResult struct {
	Findings []CheckFinding `json:"findings"`
}

// CheckRunner defines the interface for running project checks.
type CheckRunner interface {
	Check(ctx context.Context) (*CheckResult, error)
}

// FindingsDetectedError is returned when check detects findings.
type FindingsDetectedError struct {
	Errors   int
	Warnings int
}

// Error implements the error interface.
func (e *FindingsDetectedError) Error() string {
	return fmt.Sprintf("check found %d errors, %d warnings", e.Errors, e.Warnings)
}

// ExitCode returns the exit code for findings (always 2).
func (e *FindingsDetectedError) ExitCode() int {
	return 2
}

// RepairAction represents a single repair action performed.
type RepairAction struct {
	Type   FindingType `json:"type"`
	Action string      `json:"action"`
	Old    string      `json:"old"`
	New    string      `json:"new"`
}

// RepairResult holds all repairs and unrepaired findings from a repair run.
type RepairResult struct {
	Repairs    []RepairAction `json:"repairs"`
	Unrepaired []CheckFinding `json:"unrepaired"`
}

// RepairRunner defines the interface for running project repairs.
type RepairRunner interface {
	Repair(ctx context.Context) (*RepairResult, error)
}

// UnrepairedError is returned when repair leaves unresolved findings.
type UnrepairedError struct {
	Count int
}

// Error implements the error interface.
func (e *UnrepairedError) Error() string {
	return fmt.Sprintf("repair left %d unrepaired findings", e.Count)
}

// ExitCode returns the exit code for unrepaired findings (always 2).
func (e *UnrepairedError) ExitCode() int {
	return 2
}

// ExitCoder is implemented by errors that carry a specific process exit code.
type ExitCoder interface {
	ExitCode() int
}

// ExitCodeFromError returns the appropriate exit code for an error.
// nil returns 0, ExitCoder errors return their code, all others return 1.
func ExitCodeFromError(err error) int {
	if err == nil {
		return 0
	}
	var coder ExitCoder
	if errors.As(err, &coder) {
		return coder.ExitCode()
	}
	return 1
}

// checkJSONResponse is the JSON output structure for the check command.
type checkJSONResponse struct {
	Findings []CheckFinding `json:"findings"`
	Summary  struct {
		Errors   int `json:"errors"`
		Warnings int `json:"warnings"`
	} `json:"summary"`
}

// writeJSONImpl encodes v as JSON to w, handling I/O errors at the boundary.
func writeJSONImpl(w io.Writer, v interface{}) {
	if err := json.NewEncoder(w).Encode(v); err != nil {
		fmt.Fprintf(w, "{\"error\":%q}\n", err.Error())
	}
}

// countBySeverity counts errors and warnings in a slice of findings.
func countBySeverity(findings []CheckFinding) (errCount, warnCount int) {
	for _, f := range findings {
		if f.Severity == SeverityError {
			errCount++
		} else {
			warnCount++
		}
	}
	return
}

// formatCheckJSON writes findings as JSON to w.
func formatCheckJSON(w io.Writer, findings []CheckFinding, errCount, warnCount int) {
	if findings == nil {
		findings = []CheckFinding{}
	}
	out := checkJSONResponse{Findings: findings}
	out.Summary.Errors = errCount
	out.Summary.Warnings = warnCount
	writeJSONImpl(w, out)
}

// formatCheckHuman writes findings as human-readable text to w.
func formatCheckHuman(w io.Writer, findings []CheckFinding, errCount, warnCount int) {
	for _, f := range findings {
		fmt.Fprintf(w, "%s [%s] %s: %s\n", f.Path, f.Severity, f.Type, f.Message)
	}
	if errCount > 0 || warnCount > 0 {
		fmt.Fprintf(w, "\n%d error(s), %d warning(s)\n", errCount, warnCount)
	}
}

// runCheckAndReport runs the checker and formats findings as JSON or human-readable text.
// It returns a FindingsDetectedError if any findings are present.
func runCheckAndReport(cmd *cobra.Command, runner CheckRunner, jsonOutput bool) error {
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
}

// NewCheckCmd creates the check command with the given runner.
func NewCheckCmd(runner CheckRunner) *cobra.Command {
	var jsonOutput bool

	cmd := &cobra.Command{
		Use:          "check",
		Short:        "Validate project structure and content",
		SilenceUsage: true,
		RunE: func(cmd *cobra.Command, args []string) error {
			return runCheckAndReport(cmd, runner, jsonOutput || GetJSON())
		},
	}

	cmd.Flags().BoolVar(&jsonOutput, "json", false, "Output results as JSON")

	return cmd
}
