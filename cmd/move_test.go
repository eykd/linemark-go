package cmd

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"strings"
	"testing"

	"github.com/eykd/linemark-go/internal/domain"
	"github.com/spf13/cobra"
)

// mockMoveRunner is a test double for MoveRunner.
type mockMoveRunner struct {
	result   *MoveResult
	err      error
	called   bool
	selector string
	to       string
	before   string
	after    string
	apply    bool
}

func (m *mockMoveRunner) Move(ctx context.Context, selector string, to string, before string, after string, apply bool) (*MoveResult, error) {
	m.called = true
	m.selector = selector
	m.to = to
	m.before = before
	m.after = after
	m.apply = apply
	return m.result, m.err
}

func newTestMoveCmd(runner *mockMoveRunner, args ...string) (*cobra.Command, *bytes.Buffer) {
	cmd := NewMoveCmd(runner)
	if len(args) > 0 {
		cmd.SetArgs(args)
	}
	buf := new(bytes.Buffer)
	cmd.SetOut(buf)
	cmd.SetErr(new(bytes.Buffer))
	return cmd, buf
}

func TestMoveCmd_ValidSelectors(t *testing.T) {
	tests := []struct {
		name string
		args []string
	}{
		{"implicit MP source and target", []string{"001-200", "--to", "300"}},
		{"implicit SID source", []string{"A3F7c9Qx7Lm2", "--to", "300"}},
		{"explicit MP prefix source", []string{"mp:001-200", "--to", "300"}},
		{"explicit SID prefix source", []string{"sid:A3F7c9Qx7Lm2", "--to", "300"}},
		{"explicit MP prefix target", []string{"100", "--to", "mp:300"}},
		{"explicit SID prefix target", []string{"100", "--to", "sid:B8kQ2mNp4Rs1"}},
		{"both explicit prefixes", []string{"mp:100", "--to", "sid:A3F7c9Qx7Lm2"}},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			runner := &mockMoveRunner{result: &MoveResult{}}
			cmd, _ := newTestMoveCmd(runner, tt.args...)

			err := cmd.Execute()

			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if !runner.called {
				t.Error("runner should be called with valid selectors")
			}
		})
	}
}

func TestMoveCmd_RejectsInvalidSourceSelector(t *testing.T) {
	tests := []struct {
		name     string
		selector string
	}{
		{"special chars", "abc!@#"},
		{"too short", "ab"},
		{"unknown prefix", "foo:123"},
		{"mp prefix bad value", "mp:invalid"},
		{"sid prefix bad value", "sid:ab"},
		{"zero segment", "000"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			runner := &mockMoveRunner{result: &MoveResult{}}
			cmd, _ := newTestMoveCmd(runner, tt.selector, "--to", "100")

			err := cmd.Execute()

			if err == nil {
				t.Fatal("expected error for invalid source selector")
			}
			if !errors.Is(err, domain.ErrInvalidSelector) {
				t.Errorf("error should wrap ErrInvalidSelector, got: %v", err)
			}
			if runner.called {
				t.Error("runner should not be called with invalid source selector")
			}
		})
	}
}

func TestMoveCmd_RejectsInvalidTargetSelector(t *testing.T) {
	tests := []struct {
		name string
		to   string
	}{
		{"special chars", "abc!@#"},
		{"too short", "ab"},
		{"unknown prefix", "foo:123"},
		{"mp prefix bad value", "mp:invalid"},
		{"zero segment", "000"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			runner := &mockMoveRunner{result: &MoveResult{}}
			cmd, _ := newTestMoveCmd(runner, "100", "--to", tt.to)

			err := cmd.Execute()

			if err == nil {
				t.Fatal("expected error for invalid target selector")
			}
			if !errors.Is(err, domain.ErrInvalidSelector) {
				t.Errorf("error should wrap ErrInvalidSelector, got: %v", err)
			}
			if runner.called {
				t.Error("runner should not be called with invalid target selector")
			}
		})
	}
}

func TestMoveCmd_RequiresExactlyOneArg(t *testing.T) {
	tests := []struct {
		name string
		args []string
	}{
		{"no args", nil},
		{"too many args", []string{"001", "002", "--to", "300"}},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			runner := &mockMoveRunner{result: &MoveResult{}}
			cmd, _ := newTestMoveCmd(runner, tt.args...)

			err := cmd.Execute()

			if err == nil {
				t.Fatal("expected error")
			}
		})
	}
}

func TestMoveCmd_RequiresTo(t *testing.T) {
	runner := &mockMoveRunner{result: &MoveResult{}}
	cmd, _ := newTestMoveCmd(runner, "100")

	err := cmd.Execute()

	if err == nil {
		t.Fatal("expected error when --to is missing")
	}
	if runner.called {
		t.Error("runner should not be called without --to")
	}
}

// moveFixture returns a standard test fixture for move results.
func moveFixture() *MoveResult {
	return &MoveResult{
		Renames: []RenameEntry{
			{Old: "001-200_A3F7c9Qx7Lm2_draft_chapter-one.md", New: "002-100_A3F7c9Qx7Lm2_draft_chapter-one.md"},
		},
	}
}

// newTestRootMoveCmd creates a move command wired through root (for global flags like --dry-run, --json),
// capturing stdout into the returned buffer.
func newTestRootMoveCmd(runner *mockMoveRunner, args ...string) (*cobra.Command, *bytes.Buffer) {
	root := NewRootCmd()
	cmd := NewMoveCmd(runner)
	root.AddCommand(cmd)
	buf := new(bytes.Buffer)
	root.SetOut(buf)
	root.SetErr(new(bytes.Buffer))
	if len(args) > 0 {
		root.SetArgs(args)
	}
	return root, buf
}

func TestMoveCmd_HasFlags(t *testing.T) {
	tests := []struct {
		name     string
		flagName string
		wantDef  string
	}{
		{"has --json flag", "json", "false"},
		{"has --before flag", "before", ""},
		{"has --after flag", "after", ""},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			runner := &mockMoveRunner{result: &MoveResult{}}
			cmd := NewMoveCmd(runner)

			flag := cmd.Flags().Lookup(tt.flagName)
			if flag == nil {
				t.Fatalf("move command should have --%s flag", tt.flagName)
			}
			if flag.DefValue != tt.wantDef {
				t.Errorf("--%s default = %q, want %q", tt.flagName, flag.DefValue, tt.wantDef)
			}
		})
	}
}

func TestMoveCmd_BeforeAndAfterMutuallyExclusive(t *testing.T) {
	runner := &mockMoveRunner{result: moveFixture()}
	cmd, _ := newTestMoveCmd(runner, "100", "--to", "200", "--before", "300", "--after", "400")

	err := cmd.Execute()

	if err == nil {
		t.Fatal("expected error when both --before and --after are set")
	}
	if !strings.Contains(err.Error(), "before") || !strings.Contains(err.Error(), "after") {
		t.Errorf("error should mention conflicting flags, got: %v", err)
	}
	if runner.called {
		t.Error("runner should not be called when flags conflict")
	}
}

func TestMoveCmd_RunnerInteraction(t *testing.T) {
	runner := &mockMoveRunner{result: moveFixture()}
	cmd, _ := newTestMoveCmd(runner, "001-200", "--to", "300")

	err := cmd.Execute()

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !runner.called {
		t.Error("runner should be called on execute")
	}
	if runner.selector != "001-200" {
		t.Errorf("selector = %q, want %q", runner.selector, "001-200")
	}
	if runner.to != "300" {
		t.Errorf("to = %q, want %q", runner.to, "300")
	}
	if !runner.apply {
		t.Error("apply should be true by default (no --dry-run)")
	}
}

func TestMoveCmd_JSONOutput(t *testing.T) {
	runner := &mockMoveRunner{result: moveFixture()}
	root, buf := newTestRootMoveCmd(runner, "--json", "move", "001-200", "--to", "300")

	err := root.Execute()

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	var output MoveResult
	if jsonErr := json.Unmarshal(buf.Bytes(), &output); jsonErr != nil {
		t.Fatalf("invalid JSON output: %v\nraw: %s", jsonErr, buf.String())
	}
	if len(output.Renames) != 1 {
		t.Fatalf("renames count = %d, want 1", len(output.Renames))
	}
	if output.Renames[0].Old != "001-200_A3F7c9Qx7Lm2_draft_chapter-one.md" {
		t.Errorf("renames[0].old = %q, want %q", output.Renames[0].Old, "001-200_A3F7c9Qx7Lm2_draft_chapter-one.md")
	}
	if output.Renames[0].New != "002-100_A3F7c9Qx7Lm2_draft_chapter-one.md" {
		t.Errorf("renames[0].new = %q, want %q", output.Renames[0].New, "002-100_A3F7c9Qx7Lm2_draft_chapter-one.md")
	}
	if output.Planned {
		t.Error("planned should be false when not in dry-run mode")
	}
}

func TestMoveCmd_HumanReadableOutput(t *testing.T) {
	runner := &mockMoveRunner{result: moveFixture()}
	cmd, buf := newTestMoveCmd(runner, "001-200", "--to", "300")

	err := cmd.Execute()

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	output := buf.String()
	if !strings.Contains(output, "001-200_A3F7c9Qx7Lm2_draft_chapter-one.md") {
		t.Errorf("output should contain old filename, got: %q", output)
	}
	if !strings.Contains(output, "002-100_A3F7c9Qx7Lm2_draft_chapter-one.md") {
		t.Errorf("output should contain new filename, got: %q", output)
	}

	var parsed map[string]interface{}
	if json.Unmarshal(buf.Bytes(), &parsed) == nil {
		t.Errorf("output should not be valid JSON without --json flag, got: %s", output)
	}
}

func TestMoveCmd_ServiceError(t *testing.T) {
	runner := &mockMoveRunner{
		err: fmt.Errorf("filesystem error"),
	}
	cmd, _ := newTestMoveCmd(runner, "001-200", "--to", "300")

	err := cmd.Execute()

	if err == nil {
		t.Fatal("expected error for service failure")
	}
	if !strings.Contains(err.Error(), "filesystem error") {
		t.Errorf("error should contain cause, got: %v", err)
	}
}

func TestMoveCmd_ContextCancellation(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	cancel()

	runner := &mockMoveRunner{
		err: ctx.Err(),
	}
	cmd, _ := newTestMoveCmd(runner, "001-200", "--to", "300")

	err := cmd.Execute()

	if err == nil {
		t.Fatal("expected error for cancelled context")
	}
}

func TestMoveCmd_DryRunBehavior(t *testing.T) {
	tests := []struct {
		name      string
		args      []string
		wantApply bool
	}{
		{
			name:      "dry-run prevents mutation",
			args:      []string{"move", "--json", "--dry-run", "001-200", "--to", "300"},
			wantApply: false,
		},
		{
			name:      "without dry-run applies",
			args:      []string{"move", "--json", "001-200", "--to", "300"},
			wantApply: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			runner := &mockMoveRunner{result: moveFixture()}
			root, _ := newTestRootMoveCmd(runner, tt.args...)

			err := root.Execute()

			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if runner.apply != tt.wantApply {
				t.Errorf("apply = %v, want %v", runner.apply, tt.wantApply)
			}
		})
	}
}

func TestMoveCmd_DryRunSetsPlanned(t *testing.T) {
	runner := &mockMoveRunner{
		result: moveFixture(),
	}
	root, buf := newTestRootMoveCmd(runner, "move", "--json", "--dry-run", "001-200", "--to", "300")

	err := root.Execute()

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	var output MoveResult
	if jsonErr := json.Unmarshal(buf.Bytes(), &output); jsonErr != nil {
		t.Fatalf("invalid JSON: %v\nraw: %s", jsonErr, buf.String())
	}
	if !output.Planned {
		t.Error("result.planned should be true when --dry-run is active")
	}
}

func TestMoveCmd_GlobalJSONFlag(t *testing.T) {
	runner := &mockMoveRunner{result: moveFixture()}
	root, buf := newTestRootMoveCmd(runner, "--json", "move", "001-200", "--to", "300")

	err := root.Execute()

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	var parsed map[string]interface{}
	if jsonErr := json.Unmarshal(buf.Bytes(), &parsed); jsonErr != nil {
		t.Errorf("expected valid JSON output with global --json flag, got: %s", buf.String())
	}
}

func TestMoveResult_Renames_JSONTag(t *testing.T) {
	result := moveFixture()
	data, err := json.Marshal(result)
	if err != nil {
		t.Fatalf("failed to marshal: %v", err)
	}

	var parsed map[string]interface{}
	if err := json.Unmarshal(data, &parsed); err != nil {
		t.Fatalf("failed to unmarshal: %v", err)
	}
	renames, ok := parsed["renames"]
	if !ok {
		t.Fatal("JSON should include 'renames' key")
	}
	renameArr := renames.([]interface{})
	if len(renameArr) != 1 {
		t.Fatalf("renames count = %d, want 1", len(renameArr))
	}
	entry := renameArr[0].(map[string]interface{})
	if entry["old"] != "001-200_A3F7c9Qx7Lm2_draft_chapter-one.md" {
		t.Errorf("renames[0].old = %v, want %q", entry["old"], "001-200_A3F7c9Qx7Lm2_draft_chapter-one.md")
	}
	if entry["new"] != "002-100_A3F7c9Qx7Lm2_draft_chapter-one.md" {
		t.Errorf("renames[0].new = %v, want %q", entry["new"], "002-100_A3F7c9Qx7Lm2_draft_chapter-one.md")
	}
}

func TestMoveResult_PlannedField(t *testing.T) {
	tests := []struct {
		name    string
		planned bool
	}{
		{"defaults to false", false},
		{"can be set to true", true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := MoveResult{Planned: tt.planned}
			if result.Planned != tt.planned {
				t.Errorf("Planned = %v, want %v", result.Planned, tt.planned)
			}
		})
	}
}

func TestMoveResult_PlannedField_JSON(t *testing.T) {
	fixture := moveFixture()
	fixture.Planned = true
	data, err := json.Marshal(fixture)
	if err != nil {
		t.Fatalf("failed to marshal: %v", err)
	}

	var parsed map[string]interface{}
	if err := json.Unmarshal(data, &parsed); err != nil {
		t.Fatalf("failed to unmarshal: %v", err)
	}
	if _, ok := parsed["planned"]; !ok {
		t.Fatal("JSON should include 'planned' key")
	}
	if parsed["planned"] != true {
		t.Errorf("planned = %v, want true", parsed["planned"])
	}
}

func TestMoveResult_MultipleRenames(t *testing.T) {
	result := &MoveResult{
		Renames: []RenameEntry{
			{Old: "001-200_A3F7c9Qx7Lm2_draft_chapter-one.md", New: "002-100_A3F7c9Qx7Lm2_draft_chapter-one.md"},
			{Old: "001-200-100_B8kQ2mNp4Rs1_draft_scene-one.md", New: "002-100-100_B8kQ2mNp4Rs1_draft_scene-one.md"},
		},
	}

	data, err := json.Marshal(result)
	if err != nil {
		t.Fatalf("failed to marshal: %v", err)
	}

	var output MoveResult
	if err := json.Unmarshal(data, &output); err != nil {
		t.Fatalf("failed to unmarshal: %v", err)
	}
	if len(output.Renames) != 2 {
		t.Fatalf("renames count = %d, want 2", len(output.Renames))
	}
}

func TestMoveCmd_ForwardsPlacementFlags(t *testing.T) {
	tests := []struct {
		name       string
		args       []string
		wantBefore string
		wantAfter  string
	}{
		{
			name:       "before flag forwarded",
			args:       []string{"001-200", "--to", "300", "--before", "400"},
			wantBefore: "400",
			wantAfter:  "",
		},
		{
			name:       "after flag forwarded",
			args:       []string{"001-200", "--to", "300", "--after", "500"},
			wantBefore: "",
			wantAfter:  "500",
		},
		{
			name:       "no placement flags",
			args:       []string{"001-200", "--to", "300"},
			wantBefore: "",
			wantAfter:  "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			runner := &mockMoveRunner{result: moveFixture()}
			cmd, _ := newTestMoveCmd(runner, tt.args...)

			err := cmd.Execute()

			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if runner.before != tt.wantBefore {
				t.Errorf("before = %q, want %q", runner.before, tt.wantBefore)
			}
			if runner.after != tt.wantAfter {
				t.Errorf("after = %q, want %q", runner.after, tt.wantAfter)
			}
		})
	}
}

func TestMoveCmd_RegisteredWithRoot(t *testing.T) {
	found := false
	for _, sub := range rootCmd.Commands() {
		if sub.Name() == "move" {
			found = true
			break
		}
	}
	if !found {
		t.Error("move command not registered with root")
	}
}

func TestMoveCmd_HumanReadableMultipleRenames(t *testing.T) {
	result := &MoveResult{
		Renames: []RenameEntry{
			{Old: "001-200_A3F7c9Qx7Lm2_draft_chapter-one.md", New: "002-100_A3F7c9Qx7Lm2_draft_chapter-one.md"},
			{Old: "001-200-100_B8kQ2mNp4Rs1_draft_scene-one.md", New: "002-100-100_B8kQ2mNp4Rs1_draft_scene-one.md"},
		},
	}

	runner := &mockMoveRunner{result: result}
	cmd, buf := newTestMoveCmd(runner, "001-200", "--to", "002")

	err := cmd.Execute()

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	output := buf.String()
	if !strings.Contains(output, "002-100_A3F7c9Qx7Lm2_draft_chapter-one.md") {
		t.Errorf("output should contain renamed node filename, got: %q", output)
	}
	if !strings.Contains(output, "002-100-100_B8kQ2mNp4Rs1_draft_scene-one.md") {
		t.Errorf("output should contain renamed descendant filename, got: %q", output)
	}
}
