package cmd

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"strings"
	"testing"
)

// mockCheckRunner is a test double for CheckRunner.
type mockCheckRunner struct {
	result *CheckResult
	err    error
}

func (m *mockCheckRunner) Check(ctx context.Context) (*CheckResult, error) {
	return m.result, m.err
}

// checkJSONOutput is a test-only type for parsing JSON output from lmk check --json.
type checkJSONOutput struct {
	Findings []checkJSONFinding `json:"findings"`
	Summary  struct {
		Errors   int `json:"errors"`
		Warnings int `json:"warnings"`
	} `json:"summary"`
}

type checkJSONFinding struct {
	Type     string `json:"type"`
	Severity string `json:"severity"`
	Message  string `json:"message"`
	Path     string `json:"path"`
}

func TestCheckCmd_RegisteredWithRoot(t *testing.T) {
	found := false
	for _, sub := range rootCmd.Commands() {
		if sub.Use == "check" {
			found = true
			break
		}
	}
	if !found {
		t.Error("check command not registered with root")
	}
}

func TestCheckCmd_NoFindings(t *testing.T) {
	runner := &mockCheckRunner{
		result: &CheckResult{},
	}
	cmd := NewCheckCmd(runner)
	buf := new(bytes.Buffer)
	cmd.SetOut(buf)
	cmd.SetErr(new(bytes.Buffer))

	err := cmd.Execute()

	if err != nil {
		t.Fatalf("expected no error for clean check, got %v", err)
	}
}

func TestCheckCmd_NoFindings_JSON(t *testing.T) {
	runner := &mockCheckRunner{
		result: &CheckResult{},
	}
	cmd := NewCheckCmd(runner)
	cmd.SetArgs([]string{"--json"})
	buf := new(bytes.Buffer)
	cmd.SetOut(buf)
	cmd.SetErr(new(bytes.Buffer))

	err := cmd.Execute()

	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	var output checkJSONOutput
	if jsonErr := json.Unmarshal(buf.Bytes(), &output); jsonErr != nil {
		t.Fatalf("invalid JSON output: %v\nraw: %s", jsonErr, buf.String())
	}
	if len(output.Findings) != 0 {
		t.Errorf("expected 0 findings, got %d", len(output.Findings))
	}
	if output.Summary.Errors != 0 {
		t.Errorf("summary errors = %d, want 0", output.Summary.Errors)
	}
	if output.Summary.Warnings != 0 {
		t.Errorf("summary warnings = %d, want 0", output.Summary.Warnings)
	}
}

func TestCheckCmd_FindingTypes(t *testing.T) {
	tests := []struct {
		name    string
		finding CheckFinding
		wantSev string
	}{
		{
			name: "invalid_filename",
			finding: CheckFinding{
				Type:     FindingInvalidFilename,
				Severity: SeverityError,
				Message:  "File does not match canonical filename pattern",
				Path:     "bad-file.md",
			},
			wantSev: "error",
		},
		{
			name: "duplicate_sid",
			finding: CheckFinding{
				Type:     FindingDuplicateSID,
				Severity: SeverityError,
				Message:  "SID 'A3F7c9Qx7Lm2' used by nodes at 001 and 002",
				Path:     "002_A3F7c9Qx7Lm2_draft_chapter-two.md",
			},
			wantSev: "error",
		},
		{
			name: "slug_drift",
			finding: CheckFinding{
				Type:     FindingSlugDrift,
				Severity: SeverityWarning,
				Message:  "Filename slug 'chpter-one' does not match title slug 'chapter-one'",
				Path:     "001_A3F7c9Qx7Lm2_draft_chpter-one.md",
			},
			wantSev: "warning",
		},
		{
			name: "missing_draft",
			finding: CheckFinding{
				Type:     FindingMissingDraft,
				Severity: SeverityError,
				Message:  "Node 001 (A3F7c9Qx7Lm2) has no draft document",
				Path:     "001_A3F7c9Qx7Lm2_notes_chapter-one.md",
			},
			wantSev: "error",
		},
		{
			name: "missing_notes",
			finding: CheckFinding{
				Type:     FindingMissingNotes,
				Severity: SeverityWarning,
				Message:  "Node 001 (A3F7c9Qx7Lm2) has no notes document",
				Path:     "001_A3F7c9Qx7Lm2_draft_chapter-one.md",
			},
			wantSev: "warning",
		},
		{
			name: "malformed_frontmatter",
			finding: CheckFinding{
				Type:     FindingMalformedFrontmatter,
				Severity: SeverityError,
				Message:  "Cannot parse YAML frontmatter",
				Path:     "001_A3F7c9Qx7Lm2_draft_chapter-one.md",
			},
			wantSev: "error",
		},
		{
			name: "orphaned_reservation",
			finding: CheckFinding{
				Type:     FindingOrphanedReservation,
				Severity: SeverityWarning,
				Message:  "SID reservation 'A3F7c9Qx7Lm2' has no content files",
				Path:     ".linemark/ids/A3F7c9Qx7Lm2",
			},
			wantSev: "warning",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			runner := &mockCheckRunner{
				result: &CheckResult{
					Findings: []CheckFinding{tt.finding},
				},
			}
			cmd := NewCheckCmd(runner)
			cmd.SetArgs([]string{"--json"})
			buf := new(bytes.Buffer)
			cmd.SetOut(buf)
			cmd.SetErr(new(bytes.Buffer))

			err := cmd.Execute()

			var findingsErr *FindingsDetectedError
			if !errors.As(err, &findingsErr) {
				t.Fatalf("expected FindingsDetectedError, got %T: %v", err, err)
			}

			var output checkJSONOutput
			if jsonErr := json.Unmarshal(buf.Bytes(), &output); jsonErr != nil {
				t.Fatalf("invalid JSON: %v\nraw: %s", jsonErr, buf.String())
			}
			if len(output.Findings) != 1 {
				t.Fatalf("expected 1 finding, got %d", len(output.Findings))
			}
			got := output.Findings[0]
			if got.Type != string(tt.finding.Type) {
				t.Errorf("type = %q, want %q", got.Type, tt.finding.Type)
			}
			if got.Severity != tt.wantSev {
				t.Errorf("severity = %q, want %q", got.Severity, tt.wantSev)
			}
			if got.Message != tt.finding.Message {
				t.Errorf("message = %q, want %q", got.Message, tt.finding.Message)
			}
			if got.Path != tt.finding.Path {
				t.Errorf("path = %q, want %q", got.Path, tt.finding.Path)
			}
		})
	}
}

func TestCheckCmd_MixedFindings_Summary(t *testing.T) {
	runner := &mockCheckRunner{
		result: &CheckResult{
			Findings: []CheckFinding{
				{Type: FindingInvalidFilename, Severity: SeverityError, Message: "bad file", Path: "bad.md"},
				{Type: FindingDuplicateSID, Severity: SeverityError, Message: "dup sid", Path: "dup.md"},
				{Type: FindingSlugDrift, Severity: SeverityWarning, Message: "slug drift", Path: "drift.md"},
				{Type: FindingMissingNotes, Severity: SeverityWarning, Message: "missing notes", Path: "notes.md"},
				{Type: FindingOrphanedReservation, Severity: SeverityWarning, Message: "orphan", Path: ".linemark/ids/abc"},
			},
		},
	}
	cmd := NewCheckCmd(runner)
	cmd.SetArgs([]string{"--json"})
	buf := new(bytes.Buffer)
	cmd.SetOut(buf)
	cmd.SetErr(new(bytes.Buffer))

	err := cmd.Execute()

	var findingsErr *FindingsDetectedError
	if !errors.As(err, &findingsErr) {
		t.Fatalf("expected FindingsDetectedError, got %T: %v", err, err)
	}

	var output checkJSONOutput
	if jsonErr := json.Unmarshal(buf.Bytes(), &output); jsonErr != nil {
		t.Fatalf("invalid JSON: %v\nraw: %s", jsonErr, buf.String())
	}
	if output.Summary.Errors != 2 {
		t.Errorf("summary errors = %d, want 2", output.Summary.Errors)
	}
	if output.Summary.Warnings != 3 {
		t.Errorf("summary warnings = %d, want 3", output.Summary.Warnings)
	}
	if len(output.Findings) != 5 {
		t.Errorf("findings count = %d, want 5", len(output.Findings))
	}
}

func TestCheckCmd_FindingsDetectedError_Counts(t *testing.T) {
	runner := &mockCheckRunner{
		result: &CheckResult{
			Findings: []CheckFinding{
				{Type: FindingMalformedFrontmatter, Severity: SeverityError, Message: "bad yaml", Path: "a.md"},
				{Type: FindingSlugDrift, Severity: SeverityWarning, Message: "drift", Path: "b.md"},
			},
		},
	}
	cmd := NewCheckCmd(runner)
	cmd.SetArgs([]string{"--json"})
	cmd.SetOut(new(bytes.Buffer))
	cmd.SetErr(new(bytes.Buffer))

	err := cmd.Execute()

	var findingsErr *FindingsDetectedError
	if !errors.As(err, &findingsErr) {
		t.Fatalf("expected FindingsDetectedError, got %T: %v", err, err)
	}
	if findingsErr.Errors != 1 {
		t.Errorf("FindingsDetectedError.Errors = %d, want 1", findingsErr.Errors)
	}
	if findingsErr.Warnings != 1 {
		t.Errorf("FindingsDetectedError.Warnings = %d, want 1", findingsErr.Warnings)
	}
}

func TestCheckCmd_ServiceError(t *testing.T) {
	runner := &mockCheckRunner{
		err: fmt.Errorf("filesystem error"),
	}
	cmd := NewCheckCmd(runner)
	cmd.SetOut(new(bytes.Buffer))
	cmd.SetErr(new(bytes.Buffer))

	err := cmd.Execute()

	if err == nil {
		t.Fatal("expected error for service failure")
	}
	var findingsErr *FindingsDetectedError
	if errors.As(err, &findingsErr) {
		t.Error("service error should not be FindingsDetectedError")
	}
	if !strings.Contains(err.Error(), "filesystem error") {
		t.Errorf("error should contain cause, got: %v", err)
	}
}

func TestCheckCmd_HumanReadableOutput(t *testing.T) {
	runner := &mockCheckRunner{
		result: &CheckResult{
			Findings: []CheckFinding{
				{
					Type:     FindingSlugDrift,
					Severity: SeverityWarning,
					Message:  "Filename slug 'chpter-one' does not match title slug 'chapter-one'",
					Path:     "001_A3F7c9Qx7Lm2_draft_chpter-one.md",
				},
			},
		},
	}
	cmd := NewCheckCmd(runner)
	// No --json flag: human-readable output
	buf := new(bytes.Buffer)
	cmd.SetOut(buf)
	cmd.SetErr(new(bytes.Buffer))

	err := cmd.Execute()

	var findingsErr *FindingsDetectedError
	if !errors.As(err, &findingsErr) {
		t.Fatalf("expected FindingsDetectedError, got %T: %v", err, err)
	}

	output := buf.String()
	if !strings.Contains(output, "slug_drift") {
		t.Errorf("human output should contain finding type, got: %q", output)
	}
	if !strings.Contains(output, "warning") {
		t.Errorf("human output should contain severity, got: %q", output)
	}
	if !strings.Contains(output, "chpter-one") {
		t.Errorf("human output should contain message content, got: %q", output)
	}
	if !strings.Contains(output, "001_A3F7c9Qx7Lm2_draft_chpter-one.md") {
		t.Errorf("human output should contain file path, got: %q", output)
	}
}

func TestCheckCmd_HumanReadableOutput_Summary(t *testing.T) {
	runner := &mockCheckRunner{
		result: &CheckResult{
			Findings: []CheckFinding{
				{Type: FindingInvalidFilename, Severity: SeverityError, Message: "bad", Path: "a.md"},
				{Type: FindingSlugDrift, Severity: SeverityWarning, Message: "drift", Path: "b.md"},
				{Type: FindingMissingNotes, Severity: SeverityWarning, Message: "missing", Path: "c.md"},
			},
		},
	}
	cmd := NewCheckCmd(runner)
	buf := new(bytes.Buffer)
	cmd.SetOut(buf)
	cmd.SetErr(new(bytes.Buffer))

	_ = cmd.Execute()

	output := buf.String()
	if !strings.Contains(output, "1 error") {
		t.Errorf("human output should contain error count, got: %q", output)
	}
	if !strings.Contains(output, "2 warning") {
		t.Errorf("human output should contain warning count, got: %q", output)
	}
}

func TestFindingsDetectedError_ExitCode(t *testing.T) {
	tests := []struct {
		name     string
		err      *FindingsDetectedError
		wantCode int
	}{
		{
			name:     "errors only",
			err:      &FindingsDetectedError{Errors: 3, Warnings: 0},
			wantCode: 2,
		},
		{
			name:     "warnings only",
			err:      &FindingsDetectedError{Errors: 0, Warnings: 2},
			wantCode: 2,
		},
		{
			name:     "mixed errors and warnings",
			err:      &FindingsDetectedError{Errors: 1, Warnings: 1},
			wantCode: 2,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := tt.err.ExitCode()
			if got != tt.wantCode {
				t.Errorf("ExitCode() = %d, want %d", got, tt.wantCode)
			}
		})
	}
}

func TestExitCodeFromError(t *testing.T) {
	tests := []struct {
		name     string
		err      error
		wantCode int
	}{
		{
			name:     "nil error returns 0",
			err:      nil,
			wantCode: 0,
		},
		{
			name:     "generic error returns 1",
			err:      fmt.Errorf("something went wrong"),
			wantCode: 1,
		},
		{
			name:     "findings detected returns 2",
			err:      &FindingsDetectedError{Errors: 1, Warnings: 0},
			wantCode: 2,
		},
		{
			name:     "wrapped findings detected returns 2",
			err:      fmt.Errorf("check failed: %w", &FindingsDetectedError{Errors: 2, Warnings: 1}),
			wantCode: 2,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := ExitCodeFromError(tt.err)
			if got != tt.wantCode {
				t.Errorf("ExitCodeFromError() = %d, want %d", got, tt.wantCode)
			}
		})
	}
}

func TestCheckCmd_ContextCancellation(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	cancel()

	runner := &mockCheckRunner{
		err: ctx.Err(),
	}
	cmd := NewCheckCmd(runner)
	cmd.SetOut(new(bytes.Buffer))
	cmd.SetErr(new(bytes.Buffer))

	err := cmd.Execute()

	if err == nil {
		t.Fatal("expected error for cancelled context")
	}
}
