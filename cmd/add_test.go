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

// mockAddRunner is a test double for AddRunner.
type mockAddRunner struct {
	result          *AddResult
	err             error
	calledTitle     string
	applyPassed     bool
	called          bool
	calledChildOf   string
	calledSiblingOf string
}

func (m *mockAddRunner) Add(ctx context.Context, title string, apply bool, childOf string, siblingOf string) (*AddResult, error) {
	m.called = true
	m.calledTitle = title
	m.applyPassed = apply
	m.calledChildOf = childOf
	m.calledSiblingOf = siblingOf
	return m.result, m.err
}

// chapterOneResult returns a standard test fixture for add results.
func chapterOneResult() *AddResult {
	return &AddResult{
		Node: AddNodeInfo{
			MP:    "100",
			SID:   "A3F7c9Qx7Lm2",
			Title: "Chapter One",
		},
		FilesCreated: []string{
			"100_A3F7c9Qx7Lm2_draft_chapter-one.md",
			"100_A3F7c9Qx7Lm2_notes.md",
		},
	}
}

// newTestAddCmd creates an add command wired to the given runner,
// capturing stdout into the returned buffer.
func newTestAddCmd(runner *mockAddRunner, args ...string) (*cobra.Command, *bytes.Buffer) {
	cmd := NewAddCmd(runner)
	if len(args) > 0 {
		cmd.SetArgs(args)
	}
	buf := new(bytes.Buffer)
	cmd.SetOut(buf)
	cmd.SetErr(new(bytes.Buffer))
	return cmd, buf
}

func TestAddCmd_RegisteredWithRoot(t *testing.T) {
	found := false
	for _, sub := range rootCmd.Commands() {
		if sub.Name() == "add" {
			found = true
			break
		}
	}
	if !found {
		t.Error("add command not registered with root")
	}
}

func TestAddCmd_ArgumentValidation(t *testing.T) {
	tests := []struct {
		name string
		args []string
	}{
		{"requires title", nil},
		{"rejects too many args", []string{"Title One", "Title Two"}},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			runner := &mockAddRunner{result: &AddResult{}}
			cmd, _ := newTestAddCmd(runner, tt.args...)

			err := cmd.Execute()

			if err == nil {
				t.Fatal("expected error")
			}
		})
	}
}

func TestAddCmd_RunnerInteraction(t *testing.T) {
	runner := &mockAddRunner{result: chapterOneResult()}
	cmd, _ := newTestAddCmd(runner, "Chapter One")

	err := cmd.Execute()

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !runner.called {
		t.Error("runner should be called on execute")
	}
	if runner.calledTitle != "Chapter One" {
		t.Errorf("title = %q, want %q", runner.calledTitle, "Chapter One")
	}
	if !runner.applyPassed {
		t.Error("apply should be true by default (no --dry-run)")
	}
}

func TestAddCmd_HasJSONFlag(t *testing.T) {
	runner := &mockAddRunner{result: &AddResult{}}
	cmd := NewAddCmd(runner)

	flag := cmd.Flags().Lookup("json")
	if flag == nil {
		t.Fatal("add command should have --json flag")
	}
	if flag.DefValue != "false" {
		t.Errorf("--json default = %q, want %q", flag.DefValue, "false")
	}
}

func TestAddCmd_JSONOutput(t *testing.T) {
	runner := &mockAddRunner{result: chapterOneResult()}
	cmd, buf := newTestAddCmd(runner, "--json", "Chapter One")

	err := cmd.Execute()

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	var output struct {
		Node struct {
			MP    string `json:"mp"`
			SID   string `json:"sid"`
			Title string `json:"title"`
		} `json:"node"`
		FilesCreated []string `json:"files_created"`
		Planned      bool     `json:"planned"`
	}
	if jsonErr := json.Unmarshal(buf.Bytes(), &output); jsonErr != nil {
		t.Fatalf("invalid JSON output: %v\nraw: %s", jsonErr, buf.String())
	}
	if output.Node.MP != "100" {
		t.Errorf("node.mp = %q, want %q", output.Node.MP, "100")
	}
	if output.Node.SID != "A3F7c9Qx7Lm2" {
		t.Errorf("node.sid = %q, want %q", output.Node.SID, "A3F7c9Qx7Lm2")
	}
	if output.Node.Title != "Chapter One" {
		t.Errorf("node.title = %q, want %q", output.Node.Title, "Chapter One")
	}
	if len(output.FilesCreated) != 2 {
		t.Fatalf("files_created count = %d, want 2", len(output.FilesCreated))
	}
	if output.FilesCreated[0] != "100_A3F7c9Qx7Lm2_draft_chapter-one.md" {
		t.Errorf("files_created[0] = %q, want %q", output.FilesCreated[0], "100_A3F7c9Qx7Lm2_draft_chapter-one.md")
	}
	if output.FilesCreated[1] != "100_A3F7c9Qx7Lm2_notes.md" {
		t.Errorf("files_created[1] = %q, want %q", output.FilesCreated[1], "100_A3F7c9Qx7Lm2_notes.md")
	}
	if output.Planned {
		t.Error("planned should be false when not in dry-run mode")
	}
}

func TestAddCmd_JSONOutput_TwoFilesCreated(t *testing.T) {
	runner := &mockAddRunner{
		result: &AddResult{
			Node: AddNodeInfo{
				MP:    "100",
				SID:   "B8kQ2mNp4Rs1",
				Title: "My Novel",
			},
			FilesCreated: []string{
				"100_B8kQ2mNp4Rs1_draft_my-novel.md",
				"100_B8kQ2mNp4Rs1_notes.md",
			},
		},
	}
	cmd, buf := newTestAddCmd(runner, "--json", "My Novel")

	err := cmd.Execute()

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	var output struct {
		FilesCreated []string `json:"files_created"`
	}
	if jsonErr := json.Unmarshal(buf.Bytes(), &output); jsonErr != nil {
		t.Fatalf("invalid JSON: %v\nraw: %s", jsonErr, buf.String())
	}
	if len(output.FilesCreated) != 2 {
		t.Fatalf("expected 2 files_created (draft and notes), got %d", len(output.FilesCreated))
	}
	hasDraft := false
	hasNotes := false
	for _, f := range output.FilesCreated {
		if strings.Contains(f, "_draft_") {
			hasDraft = true
		}
		if strings.Contains(f, "_notes") {
			hasNotes = true
		}
	}
	if !hasDraft {
		t.Error("files_created should contain a draft file")
	}
	if !hasNotes {
		t.Error("files_created should contain a notes file")
	}
}

func TestAddCmd_HumanReadableOutput(t *testing.T) {
	runner := &mockAddRunner{result: chapterOneResult()}
	cmd, buf := newTestAddCmd(runner, "Chapter One")

	err := cmd.Execute()

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	output := buf.String()
	if !strings.Contains(output, "100_A3F7c9Qx7Lm2_draft_chapter-one.md") {
		t.Errorf("output should contain draft filename, got: %q", output)
	}
	if !strings.Contains(output, "100_A3F7c9Qx7Lm2_notes.md") {
		t.Errorf("output should contain notes filename, got: %q", output)
	}

	var parsed map[string]interface{}
	if json.Unmarshal(buf.Bytes(), &parsed) == nil {
		t.Errorf("output should not be valid JSON without --json flag, got: %s", output)
	}
}

func TestAddCmd_ServiceError(t *testing.T) {
	runner := &mockAddRunner{
		err: fmt.Errorf("filesystem error"),
	}
	cmd, _ := newTestAddCmd(runner, "Chapter One")

	err := cmd.Execute()

	if err == nil {
		t.Fatal("expected error for service failure")
	}
	if !strings.Contains(err.Error(), "filesystem error") {
		t.Errorf("error should contain cause, got: %v", err)
	}
}

func TestAddCmd_ContextCancellation(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	cancel()

	runner := &mockAddRunner{
		err: ctx.Err(),
	}
	cmd, _ := newTestAddCmd(runner, "Chapter One")

	err := cmd.Execute()

	if err == nil {
		t.Fatal("expected error for cancelled context")
	}
}

func TestAddCmd_DryRunBehavior(t *testing.T) {
	tests := []struct {
		name      string
		args      []string
		wantApply bool
	}{
		{
			name:      "dry-run prevents mutation",
			args:      []string{"add", "--json", "--dry-run", "Chapter One"},
			wantApply: false,
		},
		{
			name:      "without dry-run applies",
			args:      []string{"add", "--json", "Chapter One"},
			wantApply: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			runner := &mockAddRunner{result: chapterOneResult()}
			root := NewRootCmd()
			cmd := NewAddCmd(runner)
			root.AddCommand(cmd)
			buf := new(bytes.Buffer)
			root.SetOut(buf)
			root.SetErr(new(bytes.Buffer))
			root.SetArgs(tt.args)

			err := root.Execute()

			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if runner.applyPassed != tt.wantApply {
				t.Errorf("apply = %v, want %v", runner.applyPassed, tt.wantApply)
			}
		})
	}
}

func TestAddCmd_DryRunSetsPlanned(t *testing.T) {
	runner := &mockAddRunner{
		result: &AddResult{
			Node: AddNodeInfo{
				MP:    "100",
				SID:   "(pending)",
				Title: "Chapter One",
			},
			FilesPlanned: []string{
				"100_(sid)_draft_chapter-one.md",
				"100_(sid)_notes.md",
			},
		},
	}
	root := NewRootCmd()
	cmd := NewAddCmd(runner)
	root.AddCommand(cmd)
	buf := new(bytes.Buffer)
	root.SetOut(buf)
	root.SetErr(new(bytes.Buffer))
	root.SetArgs([]string{"add", "--json", "--dry-run", "Chapter One"})

	err := root.Execute()

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	var output struct {
		Node struct {
			SID string `json:"sid"`
		} `json:"node"`
		FilesPlanned []string `json:"files_planned"`
		Planned      bool     `json:"planned"`
	}
	if jsonErr := json.Unmarshal(buf.Bytes(), &output); jsonErr != nil {
		t.Fatalf("invalid JSON: %v\nraw: %s", jsonErr, buf.String())
	}
	if !output.Planned {
		t.Error("result.planned should be true when --dry-run is active")
	}
	if len(output.FilesPlanned) != 2 {
		t.Fatalf("files_planned count = %d, want 2", len(output.FilesPlanned))
	}
}

func TestAddCmd_GlobalJSONFlag(t *testing.T) {
	runner := &mockAddRunner{result: chapterOneResult()}
	root := NewRootCmd()
	cmd := NewAddCmd(runner)
	root.AddCommand(cmd)
	buf := new(bytes.Buffer)
	root.SetOut(buf)
	root.SetErr(new(bytes.Buffer))
	root.SetArgs([]string{"--json", "add", "Chapter One"})

	err := root.Execute()

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	var parsed map[string]interface{}
	if jsonErr := json.Unmarshal(buf.Bytes(), &parsed); jsonErr != nil {
		t.Errorf("expected valid JSON output with global --json flag, got: %s", buf.String())
	}
}

func TestAddResult_PlannedField(t *testing.T) {
	tests := []struct {
		name    string
		planned bool
	}{
		{"defaults to false", false},
		{"can be set to true", true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := AddResult{Planned: tt.planned}
			if result.Planned != tt.planned {
				t.Errorf("Planned = %v, want %v", result.Planned, tt.planned)
			}
		})
	}
}

func TestAddResult_PlannedField_JSON(t *testing.T) {
	fixture := chapterOneResult()
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

func TestAddNodeInfo_JSONFields(t *testing.T) {
	info := AddNodeInfo{
		MP:    "001-200",
		SID:   "A3F7c9Qx7Lm2",
		Title: "Chapter One",
	}
	data, err := json.Marshal(info)
	if err != nil {
		t.Fatalf("failed to marshal: %v", err)
	}

	var parsed map[string]interface{}
	if err := json.Unmarshal(data, &parsed); err != nil {
		t.Fatalf("failed to unmarshal: %v", err)
	}
	if parsed["mp"] != "001-200" {
		t.Errorf("mp = %v, want %q", parsed["mp"], "001-200")
	}
	if parsed["sid"] != "A3F7c9Qx7Lm2" {
		t.Errorf("sid = %v, want %q", parsed["sid"], "A3F7c9Qx7Lm2")
	}
	if parsed["title"] != "Chapter One" {
		t.Errorf("title = %v, want %q", parsed["title"], "Chapter One")
	}
}

func TestAddResult_FilesCreated_JSONTag(t *testing.T) {
	result := AddResult{
		Node: AddNodeInfo{MP: "100", SID: "abc", Title: "Test"},
		FilesCreated: []string{
			"100_abc_draft_test.md",
			"100_abc_notes.md",
		},
	}
	data, err := json.Marshal(result)
	if err != nil {
		t.Fatalf("failed to marshal: %v", err)
	}

	var parsed map[string]interface{}
	if err := json.Unmarshal(data, &parsed); err != nil {
		t.Fatalf("failed to unmarshal: %v", err)
	}
	if _, ok := parsed["files_created"]; !ok {
		t.Fatal("JSON should include 'files_created' key")
	}
	files := parsed["files_created"].([]interface{})
	if len(files) != 2 {
		t.Errorf("files_created count = %d, want 2", len(files))
	}
}

func TestAddResult_FilesPlanned_JSONTag(t *testing.T) {
	result := AddResult{
		Node: AddNodeInfo{MP: "100", SID: "(pending)", Title: "Test"},
		FilesPlanned: []string{
			"100_(sid)_draft_test.md",
			"100_(sid)_notes.md",
		},
		Planned: true,
	}
	data, err := json.Marshal(result)
	if err != nil {
		t.Fatalf("failed to marshal: %v", err)
	}

	var parsed map[string]interface{}
	if err := json.Unmarshal(data, &parsed); err != nil {
		t.Fatalf("failed to unmarshal: %v", err)
	}
	if _, ok := parsed["files_planned"]; !ok {
		t.Fatal("JSON should include 'files_planned' key")
	}
	files := parsed["files_planned"].([]interface{})
	if len(files) != 2 {
		t.Errorf("files_planned count = %d, want 2", len(files))
	}
}

func TestAddCmd_HasPlacementFlags(t *testing.T) {
	tests := []struct {
		name     string
		flagName string
	}{
		{"has --child-of flag", "child-of"},
		{"has --sibling-of flag", "sibling-of"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			runner := &mockAddRunner{result: &AddResult{}}
			cmd := NewAddCmd(runner)

			flag := cmd.Flags().Lookup(tt.flagName)
			if flag == nil {
				t.Fatalf("add command should have --%s flag", tt.flagName)
			}
		})
	}
}

func TestAddCmd_PlacementFlagPassthrough(t *testing.T) {
	tests := []struct {
		name          string
		args          []string
		wantChildOf   string
		wantSiblingOf string
	}{
		{
			name:        "child-of with MP selector",
			args:        []string{"--child-of", "100", "Chapter One"},
			wantChildOf: "100",
		},
		{
			name:          "sibling-of with MP selector",
			args:          []string{"--sibling-of", "200", "Chapter Two"},
			wantSiblingOf: "200",
		},
		{
			name:        "child-of with explicit SID selector",
			args:        []string{"--child-of", "sid:A3F7c9Qx7Lm2", "Scene One"},
			wantChildOf: "sid:A3F7c9Qx7Lm2",
		},
		{
			name:          "sibling-of with explicit SID selector",
			args:          []string{"--sibling-of", "sid:B8kQ2mNp4Rs1", "Scene Two"},
			wantSiblingOf: "sid:B8kQ2mNp4Rs1",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			runner := &mockAddRunner{result: chapterOneResult()}
			cmd, _ := newTestAddCmd(runner, tt.args...)

			err := cmd.Execute()
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if runner.calledChildOf != tt.wantChildOf {
				t.Errorf("childOf = %q, want %q", runner.calledChildOf, tt.wantChildOf)
			}
			if runner.calledSiblingOf != tt.wantSiblingOf {
				t.Errorf("siblingOf = %q, want %q", runner.calledSiblingOf, tt.wantSiblingOf)
			}
		})
	}
}

func TestAddCmd_PlacementFlagsMutuallyExclusive(t *testing.T) {
	runner := &mockAddRunner{result: chapterOneResult()}
	cmd, _ := newTestAddCmd(runner, "--child-of", "100", "--sibling-of", "200", "Chapter One")

	err := cmd.Execute()

	if err == nil {
		t.Fatal("expected error when both --child-of and --sibling-of are specified")
	}
	if !strings.Contains(err.Error(), "mutually exclusive") {
		t.Errorf("error should mention mutual exclusivity, got: %v", err)
	}
	if runner.called {
		t.Error("runner should not be called when flags are mutually exclusive")
	}
}

func TestAddCmd_PlacementDefaultsEmpty(t *testing.T) {
	runner := &mockAddRunner{result: chapterOneResult()}
	cmd, _ := newTestAddCmd(runner, "Chapter One")

	err := cmd.Execute()

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if runner.calledChildOf != "" {
		t.Errorf("childOf should be empty by default, got %q", runner.calledChildOf)
	}
	if runner.calledSiblingOf != "" {
		t.Errorf("siblingOf should be empty by default, got %q", runner.calledSiblingOf)
	}
}
