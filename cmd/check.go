package cmd

import (
	"context"
	"encoding/json"
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

// writeJSONImpl encodes v as JSON to w, handling I/O errors at the boundary.
func writeJSONImpl(w io.Writer, v interface{}) {
	if err := json.NewEncoder(w).Encode(v); err != nil {
		fmt.Fprintf(w, "{\"error\":%q}\n", err.Error())
	}
}

// NewCheckCmd creates the check command with the given runner.
func NewCheckCmd(runner CheckRunner) *cobra.Command {
	var jsonOutput bool

	cmd := &cobra.Command{
		Use:          "check",
		Short:        "Validate project structure and content",
		SilenceUsage: true,
		RunE: func(cmd *cobra.Command, args []string) error {
			ctx := cmd.Context()
			result, err := runner.Check(ctx)
			if err != nil {
				return err
			}

			var errCount, warnCount int
			for _, f := range result.Findings {
				if f.Severity == SeverityError {
					errCount++
				} else {
					warnCount++
				}
			}

			if jsonOutput {
				findings := result.Findings
				if findings == nil {
					findings = []CheckFinding{}
				}
				out := struct {
					Findings []CheckFinding `json:"findings"`
					Summary  struct {
						Errors   int `json:"errors"`
						Warnings int `json:"warnings"`
					} `json:"summary"`
				}{
					Findings: findings,
				}
				out.Summary.Errors = errCount
				out.Summary.Warnings = warnCount

				writeJSONImpl(cmd.OutOrStdout(), out)
			} else {
				for _, f := range result.Findings {
					fmt.Fprintf(cmd.OutOrStdout(), "%s [%s] %s: %s\n",
						f.Path, f.Severity, f.Type, f.Message)
				}
				if errCount > 0 || warnCount > 0 {
					fmt.Fprintf(cmd.OutOrStdout(), "\n%d error(s), %d warning(s)\n",
						errCount, warnCount)
				}
			}

			if len(result.Findings) > 0 {
				return &FindingsDetectedError{
					Errors:   errCount,
					Warnings: warnCount,
				}
			}

			return nil
		},
	}

	cmd.Flags().BoolVar(&jsonOutput, "json", false, "Output results as JSON")

	return cmd
}
