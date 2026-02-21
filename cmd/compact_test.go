package cmd

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"strings"
	"testing"

	"github.com/spf13/cobra"
)

// mockCompactRunner is a test double for CompactRunner.
type mockCompactRunner struct {
	result      *CompactResult
	err         error
	calledWith  string
	applyPassed bool
	called      bool
}

func (m *mockCompactRunner) Compact(ctx context.Context, selector string, apply bool) (*CompactResult, error) {
	m.called = true
	m.calledWith = selector
	m.applyPassed = apply
	return m.result, m.err
}

// newTestCompactCmd creates a compact command wired to the given runner,
// capturing stdout into the returned buffer.
func newTestCompactCmd(runner *mockCompactRunner, args ...string) (*cobra.Command, *bytes.Buffer) {
	cmd := NewCompactCmd(runner)
	if len(args) > 0 {
		cmd.SetArgs(args)
	}
	buf := new(bytes.Buffer)
	cmd.SetOut(buf)
	cmd.SetErr(new(bytes.Buffer))
	return cmd, buf
}

func TestCompactCmd_RegisteredWithRoot(t *testing.T) {
	found := false
	for _, sub := range rootCmd.Commands() {
		if sub.Name() == "compact" {
			found = true
			break
		}
	}
	if !found {
		t.Error("compact command not registered with root")
	}
}

func TestCompactCmd_HasApplyFlag(t *testing.T) {
	runner := &mockCompactRunner{result: &CompactResult{}}
	cmd := NewCompactCmd(runner)

	flag := cmd.Flags().Lookup("apply")
	if flag == nil {
		t.Fatal("compact command should have --apply flag")
	}
	if flag.DefValue != "false" {
		t.Errorf("--apply default = %q, want %q", flag.DefValue, "false")
	}
}

func TestCompactCmd_HasJSONFlag(t *testing.T) {
	runner := &mockCompactRunner{result: &CompactResult{}}
	cmd := NewCompactCmd(runner)

	flag := cmd.Flags().Lookup("json")
	if flag == nil {
		t.Fatal("compact command should have --json flag")
	}
	if flag.DefValue != "false" {
		t.Errorf("--json default = %q, want %q", flag.DefValue, "false")
	}
}

func TestCompactCmd_ReportMode_NoRenames(t *testing.T) {
	runner := &mockCompactRunner{
		result: &CompactResult{},
	}
	cmd, _ := newTestCompactCmd(runner)

	err := cmd.Execute()

	if err != nil {
		t.Fatalf("expected no error for clean compact, got %v", err)
	}
	if runner.applyPassed {
		t.Error("apply should be false in report mode")
	}
}

func TestCompactCmd_ReportMode_NoRenames_JSON(t *testing.T) {
	runner := &mockCompactRunner{
		result: &CompactResult{},
	}
	cmd, buf := newTestCompactCmd(runner, "--json")

	err := cmd.Execute()

	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	var output CompactResult
	if jsonErr := json.Unmarshal(buf.Bytes(), &output); jsonErr != nil {
		t.Fatalf("invalid JSON output: %v\nraw: %s", jsonErr, buf.String())
	}
	if len(output.Renames) != 0 {
		t.Errorf("expected 0 renames, got %d", len(output.Renames))
	}
	if output.FilesAffected != 0 {
		t.Errorf("files_affected = %d, want 0", output.FilesAffected)
	}
	if output.Warning != nil {
		t.Errorf("warning = %v, want nil", output.Warning)
	}
}

func TestCompactCmd_ReportMode_WithRenames_JSON(t *testing.T) {
	runner := &mockCompactRunner{
		result: &CompactResult{
			Renames: []RenameEntry{
				{Old: "001-101_A3F7c_draft_x.md", New: "001-100_A3F7c_draft_x.md"},
				{Old: "001-102_B4G8d_draft_y.md", New: "001-200_B4G8d_draft_y.md"},
			},
			FilesAffected: 2,
		},
	}
	cmd, buf := newTestCompactCmd(runner, "--json")

	err := cmd.Execute()

	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	var output CompactResult
	if jsonErr := json.Unmarshal(buf.Bytes(), &output); jsonErr != nil {
		t.Fatalf("invalid JSON: %v\nraw: %s", jsonErr, buf.String())
	}
	if len(output.Renames) != 2 {
		t.Fatalf("renames count = %d, want 2", len(output.Renames))
	}
	if output.Renames[0].Old != "001-101_A3F7c_draft_x.md" {
		t.Errorf("renames[0].old = %q, want %q", output.Renames[0].Old, "001-101_A3F7c_draft_x.md")
	}
	if output.Renames[0].New != "001-100_A3F7c_draft_x.md" {
		t.Errorf("renames[0].new = %q, want %q", output.Renames[0].New, "001-100_A3F7c_draft_x.md")
	}
	if output.FilesAffected != 2 {
		t.Errorf("files_affected = %d, want 2", output.FilesAffected)
	}
	if output.Warning != nil {
		t.Errorf("warning = %v, want nil", output.Warning)
	}
}

func TestCompactCmd_ReportMode_HumanReadableOutput(t *testing.T) {
	runner := &mockCompactRunner{
		result: &CompactResult{
			Renames: []RenameEntry{
				{Old: "001-101_A3F7c_draft_x.md", New: "001-100_A3F7c_draft_x.md"},
				{Old: "001-102_B4G8d_draft_y.md", New: "001-200_B4G8d_draft_y.md"},
			},
			FilesAffected: 2,
		},
	}
	cmd, buf := newTestCompactCmd(runner)

	err := cmd.Execute()

	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	output := buf.String()
	if !strings.Contains(output, "001-101_A3F7c_draft_x.md") {
		t.Errorf("human output should contain old filename, got: %q", output)
	}
	if !strings.Contains(output, "001-100_A3F7c_draft_x.md") {
		t.Errorf("human output should contain new filename, got: %q", output)
	}
	if !strings.Contains(output, "2") {
		t.Errorf("human output should contain files affected count, got: %q", output)
	}
}

func TestCompactCmd_ReportMode_HumanReadable_Summary(t *testing.T) {
	runner := &mockCompactRunner{
		result: &CompactResult{
			Renames: []RenameEntry{
				{Old: "a.md", New: "b.md"},
				{Old: "c.md", New: "d.md"},
				{Old: "e.md", New: "f.md"},
			},
			FilesAffected: 3,
		},
	}
	cmd, buf := newTestCompactCmd(runner)

	_ = cmd.Execute()

	output := buf.String()
	if !strings.Contains(output, "3 file") {
		t.Errorf("human output should contain file count, got: %q", output)
	}
}

func TestCompactCmd_ApplyMode_NoRenames(t *testing.T) {
	runner := &mockCompactRunner{
		result: &CompactResult{},
	}
	cmd, _ := newTestCompactCmd(runner, "--apply")

	err := cmd.Execute()

	if err != nil {
		t.Fatalf("expected no error for clean apply, got %v", err)
	}
	if !runner.applyPassed {
		t.Error("apply should be true when --apply flag is set")
	}
}

func TestCompactCmd_ApplyMode_NoRenames_JSON(t *testing.T) {
	runner := &mockCompactRunner{
		result: &CompactResult{},
	}
	cmd, buf := newTestCompactCmd(runner, "--apply", "--json")

	err := cmd.Execute()

	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	var output CompactResult
	if jsonErr := json.Unmarshal(buf.Bytes(), &output); jsonErr != nil {
		t.Fatalf("invalid JSON: %v\nraw: %s", jsonErr, buf.String())
	}
	if len(output.Renames) != 0 {
		t.Errorf("renames count = %d, want 0", len(output.Renames))
	}
	if output.FilesAffected != 0 {
		t.Errorf("files_affected = %d, want 0", output.FilesAffected)
	}
}

func TestCompactCmd_ApplyMode_WithRenames_JSON(t *testing.T) {
	runner := &mockCompactRunner{
		result: &CompactResult{
			Renames: []RenameEntry{
				{Old: "001-101_A3F7c_draft_x.md", New: "001-100_A3F7c_draft_x.md"},
				{Old: "001-103_B4G8d_draft_y.md", New: "001-200_B4G8d_draft_y.md"},
				{Old: "001-150_C5H9e_draft_z.md", New: "001-300_C5H9e_draft_z.md"},
			},
			FilesAffected: 3,
		},
	}
	cmd, buf := newTestCompactCmd(runner, "--apply", "--json")

	err := cmd.Execute()

	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	var output CompactResult
	if jsonErr := json.Unmarshal(buf.Bytes(), &output); jsonErr != nil {
		t.Fatalf("invalid JSON: %v\nraw: %s", jsonErr, buf.String())
	}
	if len(output.Renames) != 3 {
		t.Fatalf("renames count = %d, want 3", len(output.Renames))
	}
	if output.FilesAffected != 3 {
		t.Errorf("files_affected = %d, want 3", output.FilesAffected)
	}
	if !runner.applyPassed {
		t.Error("apply should be true when --apply flag is set")
	}
}

func TestCompactCmd_ApplyMode_HumanReadableOutput(t *testing.T) {
	runner := &mockCompactRunner{
		result: &CompactResult{
			Renames: []RenameEntry{
				{Old: "001-101_A3F7c_draft_x.md", New: "001-100_A3F7c_draft_x.md"},
			},
			FilesAffected: 1,
		},
	}
	cmd, buf := newTestCompactCmd(runner, "--apply")

	err := cmd.Execute()

	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	output := buf.String()
	if !strings.Contains(output, "001-101_A3F7c_draft_x.md") {
		t.Errorf("human output should contain old filename, got: %q", output)
	}
	if !strings.Contains(output, "001-100_A3F7c_draft_x.md") {
		t.Errorf("human output should contain new filename, got: %q", output)
	}
}

func TestCompactCmd_Warning_MoreThan50Files(t *testing.T) {
	warning := "51 files affected, this is a large operation"
	runner := &mockCompactRunner{
		result: &CompactResult{
			Renames:       make([]RenameEntry, 51),
			FilesAffected: 51,
			Warning:       &warning,
		},
	}
	cmd, buf := newTestCompactCmd(runner, "--json")

	err := cmd.Execute()

	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	var output CompactResult
	if jsonErr := json.Unmarshal(buf.Bytes(), &output); jsonErr != nil {
		t.Fatalf("invalid JSON: %v\nraw: %s", jsonErr, buf.String())
	}
	if output.Warning == nil {
		t.Fatal("expected warning for >50 files, got nil")
	}
	if !strings.Contains(*output.Warning, "51") {
		t.Errorf("warning should mention file count, got: %q", *output.Warning)
	}
}

func TestCompactCmd_Warning_HumanReadable(t *testing.T) {
	warning := "51 files affected, this is a large operation"
	runner := &mockCompactRunner{
		result: &CompactResult{
			Renames:       make([]RenameEntry, 51),
			FilesAffected: 51,
			Warning:       &warning,
		},
	}
	cmd, buf := newTestCompactCmd(runner)

	_ = cmd.Execute()

	output := buf.String()
	if !strings.Contains(output, "51") {
		t.Errorf("human output should contain warning about file count, got: %q", output)
	}
}

func TestCompactCmd_NoWarning_50OrFewerFiles(t *testing.T) {
	runner := &mockCompactRunner{
		result: &CompactResult{
			Renames:       make([]RenameEntry, 50),
			FilesAffected: 50,
		},
	}
	cmd, buf := newTestCompactCmd(runner, "--json")

	err := cmd.Execute()

	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	var output CompactResult
	if jsonErr := json.Unmarshal(buf.Bytes(), &output); jsonErr != nil {
		t.Fatalf("invalid JSON: %v\nraw: %s", jsonErr, buf.String())
	}
	if output.Warning != nil {
		t.Errorf("expected no warning for <=50 files, got: %q", *output.Warning)
	}
}

func TestCompactCmd_SelectorArgument(t *testing.T) {
	tests := []struct {
		name         string
		args         []string
		wantSelector string
	}{
		{
			name:         "no selector compacts entire outline",
			args:         []string{"--json"},
			wantSelector: "",
		},
		{
			name:         "selector compacts subtree",
			args:         []string{"001", "--json"},
			wantSelector: "001",
		},
		{
			name:         "nested selector",
			args:         []string{"001-200", "--json"},
			wantSelector: "001-200",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			runner := &mockCompactRunner{
				result: &CompactResult{},
			}
			cmd, _ := newTestCompactCmd(runner, tt.args...)

			err := cmd.Execute()

			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if runner.calledWith != tt.wantSelector {
				t.Errorf("selector = %q, want %q", runner.calledWith, tt.wantSelector)
			}
		})
	}
}

func TestCompactCmd_ServiceError(t *testing.T) {
	runner := &mockCompactRunner{
		err: fmt.Errorf("filesystem error"),
	}
	cmd, _ := newTestCompactCmd(runner)

	err := cmd.Execute()

	if err == nil {
		t.Fatal("expected error for service failure")
	}
	if !strings.Contains(err.Error(), "filesystem error") {
		t.Errorf("error should contain cause, got: %v", err)
	}
}

func TestCompactCmd_ServiceError_NotExitCoder(t *testing.T) {
	runner := &mockCompactRunner{
		err: fmt.Errorf("filesystem error"),
	}
	cmd, _ := newTestCompactCmd(runner)

	err := cmd.Execute()

	exitCode := ExitCodeFromError(err)
	if exitCode != 1 {
		t.Errorf("exit code = %d, want 1 for generic service error", exitCode)
	}
}

func TestCompactCmd_ContextCancellation(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	cancel()

	runner := &mockCompactRunner{
		err: ctx.Err(),
	}
	cmd, _ := newTestCompactCmd(runner)

	err := cmd.Execute()

	if err == nil {
		t.Fatal("expected error for cancelled context")
	}
}

func TestCompactCmd_ApplyPassedToService(t *testing.T) {
	tests := []struct {
		name      string
		args      []string
		wantApply bool
	}{
		{
			name:      "without --apply flag",
			args:      nil,
			wantApply: false,
		},
		{
			name:      "with --apply flag",
			args:      []string{"--apply"},
			wantApply: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			runner := &mockCompactRunner{
				result: &CompactResult{},
			}
			cmd, _ := newTestCompactCmd(runner, tt.args...)

			err := cmd.Execute()

			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if runner.applyPassed != tt.wantApply {
				t.Errorf("apply = %v, want %v", runner.applyPassed, tt.wantApply)
			}
		})
	}
}

func TestCompactCmd_ServiceCalled(t *testing.T) {
	runner := &mockCompactRunner{
		result: &CompactResult{},
	}
	cmd, _ := newTestCompactCmd(runner)

	err := cmd.Execute()

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !runner.called {
		t.Error("compact runner should be called on execute")
	}
}
