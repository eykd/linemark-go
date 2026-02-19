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

func TestDoctorCmd_RegisteredWithRoot(t *testing.T) {
	found := false
	for _, sub := range rootCmd.Commands() {
		if sub.Use == "doctor" {
			found = true
			break
		}
	}
	if !found {
		t.Error("doctor command not registered with root")
	}
}

func TestDoctorCmd_ReportMode_NoFindings(t *testing.T) {
	runner := &mockCheckRunner{
		result: &CheckResult{},
	}
	cmd := NewDoctorCmd(runner)
	buf := new(bytes.Buffer)
	cmd.SetOut(buf)
	cmd.SetErr(new(bytes.Buffer))

	err := cmd.Execute()

	if err != nil {
		t.Fatalf("expected no error for clean doctor, got %v", err)
	}
}

func TestDoctorCmd_ReportMode_NoFindings_JSON(t *testing.T) {
	runner := &mockCheckRunner{
		result: &CheckResult{},
	}
	cmd := NewDoctorCmd(runner)
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

func TestDoctorCmd_ReportMode_WithFindings(t *testing.T) {
	tests := []struct {
		name     string
		findings []CheckFinding
		wantErrs int
		wantWarn int
	}{
		{
			name: "single error finding",
			findings: []CheckFinding{
				{Type: FindingInvalidFilename, Severity: SeverityError, Message: "bad file", Path: "bad.md"},
			},
			wantErrs: 1,
			wantWarn: 0,
		},
		{
			name: "single warning finding",
			findings: []CheckFinding{
				{Type: FindingSlugDrift, Severity: SeverityWarning, Message: "slug drift", Path: "drift.md"},
			},
			wantErrs: 0,
			wantWarn: 1,
		},
		{
			name: "mixed findings",
			findings: []CheckFinding{
				{Type: FindingInvalidFilename, Severity: SeverityError, Message: "bad file", Path: "bad.md"},
				{Type: FindingDuplicateSID, Severity: SeverityError, Message: "dup sid", Path: "dup.md"},
				{Type: FindingSlugDrift, Severity: SeverityWarning, Message: "slug drift", Path: "drift.md"},
				{Type: FindingMissingNotes, Severity: SeverityWarning, Message: "missing notes", Path: "notes.md"},
				{Type: FindingOrphanedReservation, Severity: SeverityWarning, Message: "orphan", Path: ".linemark/ids/abc"},
			},
			wantErrs: 2,
			wantWarn: 3,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			runner := &mockCheckRunner{
				result: &CheckResult{Findings: tt.findings},
			}
			cmd := NewDoctorCmd(runner)
			cmd.SetArgs([]string{"--json"})
			buf := new(bytes.Buffer)
			cmd.SetOut(buf)
			cmd.SetErr(new(bytes.Buffer))

			err := cmd.Execute()

			var findingsErr *FindingsDetectedError
			if !errors.As(err, &findingsErr) {
				t.Fatalf("expected FindingsDetectedError, got %T: %v", err, err)
			}
			if findingsErr.Errors != tt.wantErrs {
				t.Errorf("FindingsDetectedError.Errors = %d, want %d", findingsErr.Errors, tt.wantErrs)
			}
			if findingsErr.Warnings != tt.wantWarn {
				t.Errorf("FindingsDetectedError.Warnings = %d, want %d", findingsErr.Warnings, tt.wantWarn)
			}

			var output checkJSONOutput
			if jsonErr := json.Unmarshal(buf.Bytes(), &output); jsonErr != nil {
				t.Fatalf("invalid JSON: %v\nraw: %s", jsonErr, buf.String())
			}
			if len(output.Findings) != len(tt.findings) {
				t.Errorf("findings count = %d, want %d", len(output.Findings), len(tt.findings))
			}
			if output.Summary.Errors != tt.wantErrs {
				t.Errorf("summary errors = %d, want %d", output.Summary.Errors, tt.wantErrs)
			}
			if output.Summary.Warnings != tt.wantWarn {
				t.Errorf("summary warnings = %d, want %d", output.Summary.Warnings, tt.wantWarn)
			}
		})
	}
}

func TestDoctorCmd_ReportMode_HumanReadableOutput(t *testing.T) {
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
	cmd := NewDoctorCmd(runner)
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

func TestDoctorCmd_ReportMode_HumanReadable_Summary(t *testing.T) {
	runner := &mockCheckRunner{
		result: &CheckResult{
			Findings: []CheckFinding{
				{Type: FindingInvalidFilename, Severity: SeverityError, Message: "bad", Path: "a.md"},
				{Type: FindingSlugDrift, Severity: SeverityWarning, Message: "drift", Path: "b.md"},
				{Type: FindingMissingNotes, Severity: SeverityWarning, Message: "missing", Path: "c.md"},
			},
		},
	}
	cmd := NewDoctorCmd(runner)
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

func TestDoctorCmd_ServiceError(t *testing.T) {
	runner := &mockCheckRunner{
		err: fmt.Errorf("filesystem error"),
	}
	cmd := NewDoctorCmd(runner)
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

func TestDoctorCmd_ContextCancellation(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	cancel()

	runner := &mockCheckRunner{
		err: ctx.Err(),
	}
	cmd := NewDoctorCmd(runner)
	cmd.SetOut(new(bytes.Buffer))
	cmd.SetErr(new(bytes.Buffer))

	err := cmd.Execute()

	if err == nil {
		t.Fatal("expected error for cancelled context")
	}
}

func TestDoctorCmd_HasApplyFlag(t *testing.T) {
	runner := &mockCheckRunner{
		result: &CheckResult{},
	}
	cmd := NewDoctorCmd(runner)

	flag := cmd.Flags().Lookup("apply")
	if flag == nil {
		t.Fatal("doctor command should have --apply flag")
	}
	if flag.DefValue != "false" {
		t.Errorf("--apply default = %q, want %q", flag.DefValue, "false")
	}
}

func TestDoctorCmd_HasJSONFlag(t *testing.T) {
	runner := &mockCheckRunner{
		result: &CheckResult{},
	}
	cmd := NewDoctorCmd(runner)

	flag := cmd.Flags().Lookup("json")
	if flag == nil {
		t.Fatal("doctor command should have --json flag")
	}
	if flag.DefValue != "false" {
		t.Errorf("--json default = %q, want %q", flag.DefValue, "false")
	}
}
