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

// --- Apply mode tests ---

// mockRepairRunner is a test double for RepairRunner.
type mockRepairRunner struct {
	result *RepairResult
	err    error
	called bool
}

func (m *mockRepairRunner) Repair(ctx context.Context) (*RepairResult, error) {
	m.called = true
	return m.result, m.err
}

// repairJSONOutput is a test-only type for parsing JSON output from lmk doctor --apply --json.
type repairJSONOutput struct {
	Repairs    []repairJSONAction `json:"repairs"`
	Unrepaired []checkJSONFinding `json:"unrepaired"`
	Summary    struct {
		Repaired   int `json:"repaired"`
		Unrepaired int `json:"unrepaired"`
	} `json:"summary"`
}

type repairJSONAction struct {
	Type   string `json:"type"`
	Action string `json:"action"`
	Old    string `json:"old"`
	New    string `json:"new"`
}

func TestDoctorCmd_Apply_RepairScenarios_JSON(t *testing.T) {
	tests := []struct {
		name           string
		repairs        []RepairAction
		unrepaired     []CheckFinding
		wantRepaired   int
		wantUnrepaired int
		wantErr        bool
	}{
		{
			name: "fixes slug drift",
			repairs: []RepairAction{
				{
					Type:   FindingSlugDrift,
					Action: "rename",
					Old:    "001-200_A3F7c9Qx7Lm2_draft_chpter-one.md",
					New:    "001-200_A3F7c9Qx7Lm2_draft_chapter-one.md",
				},
			},
			wantRepaired:   1,
			wantUnrepaired: 0,
		},
		{
			name: "creates missing notes",
			repairs: []RepairAction{
				{
					Type:   FindingMissingNotes,
					Action: "create",
					Old:    "",
					New:    "001-200_A3F7c9Qx7Lm2_notes_chapter-one.md",
				},
			},
			wantRepaired:   1,
			wantUnrepaired: 0,
		},
		{
			name: "reserves unreserved SIDs",
			repairs: []RepairAction{
				{
					Type:   FindingUnreservedSID,
					Action: "reserve",
					Old:    "",
					New:    ".linemark/ids/A3F7c9Qx7Lm2",
				},
			},
			wantRepaired:   1,
			wantUnrepaired: 0,
		},
		{
			name: "reports duplicate SIDs as unrepaired",
			unrepaired: []CheckFinding{
				{
					Type:     FindingDuplicateSID,
					Severity: SeverityError,
					Message:  "SID 'A3F7c9Qx7Lm2' used by nodes at 001 and 002",
					Path:     "002_A3F7c9Qx7Lm2_draft_chapter-two.md",
				},
			},
			wantRepaired:   0,
			wantUnrepaired: 1,
			wantErr:        true,
		},
		{
			name: "mixed repairs and unrepaired",
			repairs: []RepairAction{
				{
					Type:   FindingSlugDrift,
					Action: "rename",
					Old:    "001-200_A3F7c9Qx7Lm2_draft_chpter-one.md",
					New:    "001-200_A3F7c9Qx7Lm2_draft_chapter-one.md",
				},
				{
					Type:   FindingMissingNotes,
					Action: "create",
					Old:    "",
					New:    "002-200_B4G8d0Ry8Mn3_notes_chapter-two.md",
				},
			},
			unrepaired: []CheckFinding{
				{
					Type:     FindingDuplicateSID,
					Severity: SeverityError,
					Message:  "duplicate sid",
					Path:     "003_C5H9e1Sz9No4_draft_chapter-three.md",
				},
			},
			wantRepaired:   2,
			wantUnrepaired: 1,
			wantErr:        true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			checker := &mockCheckRunner{result: &CheckResult{}}
			repairer := &mockRepairRunner{
				result: &RepairResult{
					Repairs:    tt.repairs,
					Unrepaired: tt.unrepaired,
				},
			}
			cmd := NewDoctorCmd(checker, repairer)
			cmd.SetArgs([]string{"--apply", "--json"})
			buf := new(bytes.Buffer)
			cmd.SetOut(buf)
			cmd.SetErr(new(bytes.Buffer))

			err := cmd.Execute()

			if tt.wantErr {
				var unrepairedErr *UnrepairedError
				if !errors.As(err, &unrepairedErr) {
					t.Fatalf("expected UnrepairedError, got %T: %v", err, err)
				}
				if unrepairedErr.Count != tt.wantUnrepaired {
					t.Errorf("UnrepairedError.Count = %d, want %d", unrepairedErr.Count, tt.wantUnrepaired)
				}
			} else if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}

			var output repairJSONOutput
			if jsonErr := json.Unmarshal(buf.Bytes(), &output); jsonErr != nil {
				t.Fatalf("invalid JSON: %v\nraw: %s", jsonErr, buf.String())
			}
			if len(output.Repairs) != len(tt.repairs) {
				t.Errorf("repairs count = %d, want %d", len(output.Repairs), len(tt.repairs))
			}
			if len(output.Unrepaired) != len(tt.unrepaired) {
				t.Errorf("unrepaired count = %d, want %d", len(output.Unrepaired), len(tt.unrepaired))
			}
			if output.Summary.Repaired != tt.wantRepaired {
				t.Errorf("summary repaired = %d, want %d", output.Summary.Repaired, tt.wantRepaired)
			}
			if output.Summary.Unrepaired != tt.wantUnrepaired {
				t.Errorf("summary unrepaired = %d, want %d", output.Summary.Unrepaired, tt.wantUnrepaired)
			}
		})
	}
}

func TestDoctorCmd_Apply_NoFindings(t *testing.T) {
	checker := &mockCheckRunner{result: &CheckResult{}}
	repairer := &mockRepairRunner{
		result: &RepairResult{},
	}
	cmd := NewDoctorCmd(checker, repairer)
	cmd.SetArgs([]string{"--apply", "--json"})
	buf := new(bytes.Buffer)
	cmd.SetOut(buf)
	cmd.SetErr(new(bytes.Buffer))

	err := cmd.Execute()

	if err != nil {
		t.Fatalf("expected no error for clean apply, got %v", err)
	}

	var output repairJSONOutput
	if jsonErr := json.Unmarshal(buf.Bytes(), &output); jsonErr != nil {
		t.Fatalf("invalid JSON: %v\nraw: %s", jsonErr, buf.String())
	}
	if len(output.Repairs) != 0 {
		t.Errorf("repairs count = %d, want 0", len(output.Repairs))
	}
	if len(output.Unrepaired) != 0 {
		t.Errorf("unrepaired count = %d, want 0", len(output.Unrepaired))
	}
	if output.Summary.Repaired != 0 {
		t.Errorf("summary repaired = %d, want 0", output.Summary.Repaired)
	}
	if output.Summary.Unrepaired != 0 {
		t.Errorf("summary unrepaired = %d, want 0", output.Summary.Unrepaired)
	}
}

func TestDoctorCmd_Apply_ServiceError(t *testing.T) {
	checker := &mockCheckRunner{result: &CheckResult{}}
	repairer := &mockRepairRunner{
		err: fmt.Errorf("filesystem error"),
	}
	cmd := NewDoctorCmd(checker, repairer)
	cmd.SetArgs([]string{"--apply"})
	cmd.SetOut(new(bytes.Buffer))
	cmd.SetErr(new(bytes.Buffer))

	err := cmd.Execute()

	if err == nil {
		t.Fatal("expected error for repair service failure")
	}
	var unrepairedErr *UnrepairedError
	if errors.As(err, &unrepairedErr) {
		t.Error("service error should not be UnrepairedError")
	}
	if !strings.Contains(err.Error(), "filesystem error") {
		t.Errorf("error should contain cause, got: %v", err)
	}
}

func TestDoctorCmd_Apply_HumanReadableOutput(t *testing.T) {
	checker := &mockCheckRunner{result: &CheckResult{}}
	repairer := &mockRepairRunner{
		result: &RepairResult{
			Repairs: []RepairAction{
				{
					Type:   FindingSlugDrift,
					Action: "rename",
					Old:    "001-200_A3F7c9Qx7Lm2_draft_chpter-one.md",
					New:    "001-200_A3F7c9Qx7Lm2_draft_chapter-one.md",
				},
			},
			Unrepaired: []CheckFinding{
				{
					Type:     FindingDuplicateSID,
					Severity: SeverityError,
					Message:  "SID 'A3F7c9Qx7Lm2' used by nodes at 001 and 002",
					Path:     "002_A3F7c9Qx7Lm2_draft_chapter-two.md",
				},
			},
		},
	}
	cmd := NewDoctorCmd(checker, repairer)
	cmd.SetArgs([]string{"--apply"})
	buf := new(bytes.Buffer)
	cmd.SetOut(buf)
	cmd.SetErr(new(bytes.Buffer))

	_ = cmd.Execute()

	output := buf.String()
	if !strings.Contains(output, "slug_drift") {
		t.Errorf("human output should contain repair type, got: %q", output)
	}
	if !strings.Contains(output, "rename") {
		t.Errorf("human output should contain action, got: %q", output)
	}
	if !strings.Contains(output, "chpter-one") {
		t.Errorf("human output should contain old path, got: %q", output)
	}
	if !strings.Contains(output, "chapter-one") {
		t.Errorf("human output should contain new path, got: %q", output)
	}
	if !strings.Contains(output, "duplicate_sid") {
		t.Errorf("human output should contain unrepaired type, got: %q", output)
	}
}

func TestDoctorCmd_Apply_HumanReadable_Summary(t *testing.T) {
	checker := &mockCheckRunner{result: &CheckResult{}}
	repairer := &mockRepairRunner{
		result: &RepairResult{
			Repairs: []RepairAction{
				{Type: FindingSlugDrift, Action: "rename", Old: "a.md", New: "b.md"},
				{Type: FindingMissingNotes, Action: "create", Old: "", New: "c.md"},
			},
			Unrepaired: []CheckFinding{
				{Type: FindingDuplicateSID, Severity: SeverityError, Message: "dup", Path: "d.md"},
			},
		},
	}
	cmd := NewDoctorCmd(checker, repairer)
	cmd.SetArgs([]string{"--apply"})
	buf := new(bytes.Buffer)
	cmd.SetOut(buf)
	cmd.SetErr(new(bytes.Buffer))

	_ = cmd.Execute()

	output := buf.String()
	if !strings.Contains(output, "2 repaired") {
		t.Errorf("human output should contain repaired count, got: %q", output)
	}
	if !strings.Contains(output, "1 unrepaired") {
		t.Errorf("human output should contain unrepaired count, got: %q", output)
	}
}

func TestDoctorCmd_WithoutApply_DoesNotCallRepairer(t *testing.T) {
	checker := &mockCheckRunner{result: &CheckResult{}}
	repairer := &mockRepairRunner{
		result: &RepairResult{},
	}
	cmd := NewDoctorCmd(checker, repairer)
	// No --apply flag
	cmd.SetOut(new(bytes.Buffer))
	cmd.SetErr(new(bytes.Buffer))

	_ = cmd.Execute()

	if repairer.called {
		t.Error("repairer should not be called without --apply flag")
	}
}

func TestDoctorCmd_Apply_Idempotent(t *testing.T) {
	checker := &mockCheckRunner{result: &CheckResult{}}
	// Second run returns no repairs (everything already fixed)
	repairer := &mockRepairRunner{
		result: &RepairResult{},
	}
	cmd := NewDoctorCmd(checker, repairer)
	cmd.SetArgs([]string{"--apply", "--json"})
	buf := new(bytes.Buffer)
	cmd.SetOut(buf)
	cmd.SetErr(new(bytes.Buffer))

	err := cmd.Execute()

	if err != nil {
		t.Fatalf("expected no error for idempotent run, got %v", err)
	}

	var output repairJSONOutput
	if jsonErr := json.Unmarshal(buf.Bytes(), &output); jsonErr != nil {
		t.Fatalf("invalid JSON: %v\nraw: %s", jsonErr, buf.String())
	}
	if output.Summary.Repaired != 0 {
		t.Errorf("idempotent run should have 0 repaired, got %d", output.Summary.Repaired)
	}
	if output.Summary.Unrepaired != 0 {
		t.Errorf("idempotent run should have 0 unrepaired, got %d", output.Summary.Unrepaired)
	}
}

func TestDoctorCmd_Apply_UnrepairedError_ExitCode(t *testing.T) {
	checker := &mockCheckRunner{result: &CheckResult{}}
	repairer := &mockRepairRunner{
		result: &RepairResult{
			Unrepaired: []CheckFinding{
				{Type: FindingDuplicateSID, Severity: SeverityError, Message: "dup", Path: "a.md"},
			},
		},
	}
	cmd := NewDoctorCmd(checker, repairer)
	cmd.SetArgs([]string{"--apply"})
	cmd.SetOut(new(bytes.Buffer))
	cmd.SetErr(new(bytes.Buffer))

	err := cmd.Execute()

	exitCode := ExitCodeFromError(err)
	if exitCode != 2 {
		t.Errorf("exit code = %d, want 2 for unrepaired findings", exitCode)
	}
}

// trackingCheckRunner wraps mockCheckRunner and tracks whether Check was called.
type trackingCheckRunner struct {
	mockCheckRunner
	called bool
}

func (t *trackingCheckRunner) Check(ctx context.Context) (*CheckResult, error) {
	t.called = true
	return t.mockCheckRunner.Check(ctx)
}

// TestDoctorCmd_Apply_CheckerInvokedAfterRepair verifies that the doctor
// command invokes the checker in --apply mode to detect post-repair findings
// that the repairer itself did not attempt to fix.
func TestDoctorCmd_Apply_CheckerInvokedAfterRepair(t *testing.T) {
	checker := &trackingCheckRunner{mockCheckRunner: mockCheckRunner{result: &CheckResult{}}}
	repairer := &mockRepairRunner{result: &RepairResult{}}
	cmd := NewDoctorCmd(checker, repairer)
	cmd.SetArgs([]string{"--apply"})
	cmd.SetOut(new(bytes.Buffer))
	cmd.SetErr(new(bytes.Buffer))

	_ = cmd.Execute()

	if !checker.called {
		t.Error("checker should be invoked in --apply mode to validate post-repair state")
	}
}

// TestDoctorCmd_Apply_ExitCode2_WhenPostRepairCheckFindsIssues verifies that
// --apply returns exit code 2 when the post-repair check still detects findings,
// even if the repairer itself reported zero unrepaired findings (e.g. invalid_filename
// which the repairer silently skips).
func TestDoctorCmd_Apply_ExitCode2_WhenPostRepairCheckFindsIssues(t *testing.T) {
	tests := []struct {
		name              string
		repairResult      *RepairResult
		postCheckFindings []CheckFinding
		wantExitCode      int
	}{
		{
			name:         "invalid filename skipped by repairer causes exit 2",
			repairResult: &RepairResult{},
			postCheckFindings: []CheckFinding{
				{
					Type:     FindingInvalidFilename,
					Severity: SeverityError,
					Message:  "filename does not match canonical pattern",
					Path:     "999_INVALID_draft_bad.md",
				},
			},
			wantExitCode: 2,
		},
		{
			name: "clean post-repair check exits 0",
			repairResult: &RepairResult{
				Repairs: []RepairAction{
					{Type: FindingSlugDrift, Action: "rename", Old: "old.md", New: "new.md"},
				},
			},
			postCheckFindings: nil,
			wantExitCode:      0,
		},
		{
			name: "partial repair with remaining invalid filename causes exit 2",
			repairResult: &RepairResult{
				Repairs: []RepairAction{
					{Type: FindingSlugDrift, Action: "rename", Old: "old.md", New: "new.md"},
				},
			},
			postCheckFindings: []CheckFinding{
				{
					Type:     FindingInvalidFilename,
					Severity: SeverityError,
					Message:  "filename does not match canonical pattern",
					Path:     "999_INVALID_draft_bad.md",
				},
			},
			wantExitCode: 2,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			checker := &mockCheckRunner{
				result: &CheckResult{Findings: tt.postCheckFindings},
			}
			repairer := &mockRepairRunner{result: tt.repairResult}
			cmd := NewDoctorCmd(checker, repairer)
			cmd.SetArgs([]string{"--apply"})
			cmd.SetOut(new(bytes.Buffer))
			cmd.SetErr(new(bytes.Buffer))

			err := cmd.Execute()

			gotExitCode := ExitCodeFromError(err)
			if gotExitCode != tt.wantExitCode {
				t.Errorf("exit code = %d, want %d (err = %v)", gotExitCode, tt.wantExitCode, err)
			}
		})
	}
}
