package cmd

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"strings"
	"testing"
)

// mockTypesService is a test double for TypesService.
type mockTypesService struct {
	listResult   *TypesListResult
	listErr      error
	addResult    *TypesModifyResult
	addErr       error
	removeResult *TypesModifyResult
	removeErr    error
}

func (m *mockTypesService) ListTypes(ctx context.Context, selector string) (*TypesListResult, error) {
	return m.listResult, m.listErr
}

func (m *mockTypesService) AddType(ctx context.Context, docType, selector string) (*TypesModifyResult, error) {
	return m.addResult, m.addErr
}

func (m *mockTypesService) RemoveType(ctx context.Context, docType, selector string) (*TypesModifyResult, error) {
	return m.removeResult, m.removeErr
}

// typesListJSONOutput is a test-only type for parsing JSON output from lmk types list --json.
type typesListJSONOutput struct {
	Node struct {
		MP  string `json:"mp"`
		SID string `json:"sid"`
	} `json:"node"`
	Types []string `json:"types"`
}

// typesModifyJSONOutput is a test-only type for parsing JSON output from lmk types add/remove --json.
type typesModifyJSONOutput struct {
	Node struct {
		MP  string `json:"mp"`
		SID string `json:"sid"`
	} `json:"node"`
	Filename string `json:"filename"`
}

func TestTypesCmd_RegisteredWithRoot(t *testing.T) {
	found := false
	for _, sub := range rootCmd.Commands() {
		if sub.Use == "types" {
			found = true
			break
		}
	}
	if !found {
		t.Error("types command not registered with root")
	}
}

func TestTypesCmd_HasSubcommands(t *testing.T) {
	svc := &mockTypesService{}
	cmd := NewTypesCmd(svc)

	subcommands := cmd.Commands()
	names := make(map[string]bool)
	for _, sub := range subcommands {
		names[sub.Name()] = true
	}

	for _, want := range []string{"list", "add", "remove"} {
		if !names[want] {
			t.Errorf("types command missing subcommand %q", want)
		}
	}
}

// --- types list tests ---

func TestTypesListCmd_HumanOutput(t *testing.T) {
	svc := &mockTypesService{
		listResult: &TypesListResult{
			Node:  NodeInfo{MP: "001", SID: "A3F7c9Qx7Lm2"},
			Types: []string{"draft", "notes"},
		},
	}
	cmd := NewTypesCmd(svc)
	cmd.SetArgs([]string{"list", "001"})
	buf := new(bytes.Buffer)
	cmd.SetOut(buf)
	cmd.SetErr(new(bytes.Buffer))

	err := cmd.Execute()

	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	output := buf.String()
	if !strings.Contains(output, "draft") {
		t.Errorf("output should contain 'draft', got: %q", output)
	}
	if !strings.Contains(output, "notes") {
		t.Errorf("output should contain 'notes', got: %q", output)
	}
}

func TestTypesListCmd_JSONOutput(t *testing.T) {
	svc := &mockTypesService{
		listResult: &TypesListResult{
			Node:  NodeInfo{MP: "001", SID: "A3F7c9Qx7Lm2"},
			Types: []string{"draft", "notes", "characters"},
		},
	}
	cmd := NewTypesCmd(svc)
	cmd.SetArgs([]string{"list", "--json", "001"})
	buf := new(bytes.Buffer)
	cmd.SetOut(buf)
	cmd.SetErr(new(bytes.Buffer))

	err := cmd.Execute()

	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	var output typesListJSONOutput
	if jsonErr := json.Unmarshal(buf.Bytes(), &output); jsonErr != nil {
		t.Fatalf("invalid JSON output: %v\nraw: %s", jsonErr, buf.String())
	}
	if output.Node.MP != "001" {
		t.Errorf("node.mp = %q, want %q", output.Node.MP, "001")
	}
	if output.Node.SID != "A3F7c9Qx7Lm2" {
		t.Errorf("node.sid = %q, want %q", output.Node.SID, "A3F7c9Qx7Lm2")
	}
	if len(output.Types) != 3 {
		t.Fatalf("types count = %d, want 3", len(output.Types))
	}
	wantTypes := []string{"draft", "notes", "characters"}
	for i, want := range wantTypes {
		if output.Types[i] != want {
			t.Errorf("types[%d] = %q, want %q", i, output.Types[i], want)
		}
	}
}

func TestTypesListCmd_ServiceError(t *testing.T) {
	svc := &mockTypesService{
		listErr: fmt.Errorf("node not found"),
	}
	cmd := NewTypesCmd(svc)
	cmd.SetArgs([]string{"list", "999"})
	cmd.SetOut(new(bytes.Buffer))
	cmd.SetErr(new(bytes.Buffer))

	err := cmd.Execute()

	if err == nil {
		t.Fatal("expected error for service failure")
	}
	if !strings.Contains(err.Error(), "node not found") {
		t.Errorf("error should contain cause, got: %v", err)
	}
}

func TestTypesListCmd_MissingSelectorArg(t *testing.T) {
	svc := &mockTypesService{}
	cmd := NewTypesCmd(svc)
	cmd.SetArgs([]string{"list"})
	cmd.SetOut(new(bytes.Buffer))
	cmd.SetErr(new(bytes.Buffer))

	err := cmd.Execute()

	if err == nil {
		t.Fatal("expected error for missing selector argument")
	}
}

func TestTypesListCmd_ContextCancellation(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	cancel()

	svc := &mockTypesService{
		listErr: ctx.Err(),
	}
	cmd := NewTypesCmd(svc)
	cmd.SetArgs([]string{"list", "001"})
	cmd.SetOut(new(bytes.Buffer))
	cmd.SetErr(new(bytes.Buffer))

	err := cmd.Execute()

	if err == nil {
		t.Fatal("expected error for cancelled context")
	}
}

// --- types add tests ---

func TestTypesAddCmd_HumanOutput(t *testing.T) {
	svc := &mockTypesService{
		addResult: &TypesModifyResult{
			Node:     NodeInfo{MP: "001", SID: "A3F7c9Qx7Lm2"},
			Filename: "001_A3F7c9Qx7Lm2_characters.md",
		},
	}
	cmd := NewTypesCmd(svc)
	cmd.SetArgs([]string{"add", "characters", "001"})
	buf := new(bytes.Buffer)
	cmd.SetOut(buf)
	cmd.SetErr(new(bytes.Buffer))

	err := cmd.Execute()

	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	output := buf.String()
	if !strings.Contains(output, "001_A3F7c9Qx7Lm2_characters.md") {
		t.Errorf("output should contain created filename, got: %q", output)
	}
}

func TestTypesAddCmd_JSONOutput(t *testing.T) {
	svc := &mockTypesService{
		addResult: &TypesModifyResult{
			Node:     NodeInfo{MP: "001", SID: "A3F7c9Qx7Lm2"},
			Filename: "001_A3F7c9Qx7Lm2_characters.md",
		},
	}
	cmd := NewTypesCmd(svc)
	cmd.SetArgs([]string{"add", "--json", "characters", "001"})
	buf := new(bytes.Buffer)
	cmd.SetOut(buf)
	cmd.SetErr(new(bytes.Buffer))

	err := cmd.Execute()

	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	var output typesModifyJSONOutput
	if jsonErr := json.Unmarshal(buf.Bytes(), &output); jsonErr != nil {
		t.Fatalf("invalid JSON output: %v\nraw: %s", jsonErr, buf.String())
	}
	if output.Node.MP != "001" {
		t.Errorf("node.mp = %q, want %q", output.Node.MP, "001")
	}
	if output.Node.SID != "A3F7c9Qx7Lm2" {
		t.Errorf("node.sid = %q, want %q", output.Node.SID, "A3F7c9Qx7Lm2")
	}
	if output.Filename != "001_A3F7c9Qx7Lm2_characters.md" {
		t.Errorf("filename = %q, want %q", output.Filename, "001_A3F7c9Qx7Lm2_characters.md")
	}
}

func TestTypesAddCmd_ServiceError(t *testing.T) {
	svc := &mockTypesService{
		addErr: fmt.Errorf("lock acquisition failed"),
	}
	cmd := NewTypesCmd(svc)
	cmd.SetArgs([]string{"add", "characters", "001"})
	cmd.SetOut(new(bytes.Buffer))
	cmd.SetErr(new(bytes.Buffer))

	err := cmd.Execute()

	if err == nil {
		t.Fatal("expected error for service failure")
	}
	if !strings.Contains(err.Error(), "lock acquisition failed") {
		t.Errorf("error should contain cause, got: %v", err)
	}
}

func TestTypesAddCmd_MissingArgs(t *testing.T) {
	tests := []struct {
		name string
		args []string
	}{
		{"no args", []string{"add"}},
		{"type only", []string{"add", "characters"}},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			svc := &mockTypesService{}
			cmd := NewTypesCmd(svc)
			cmd.SetArgs(tt.args)
			cmd.SetOut(new(bytes.Buffer))
			cmd.SetErr(new(bytes.Buffer))

			err := cmd.Execute()

			if err == nil {
				t.Fatal("expected error for missing arguments")
			}
		})
	}
}

// --- types remove tests ---

func TestTypesRemoveCmd_HumanOutput(t *testing.T) {
	svc := &mockTypesService{
		removeResult: &TypesModifyResult{
			Node:     NodeInfo{MP: "001", SID: "A3F7c9Qx7Lm2"},
			Filename: "001_A3F7c9Qx7Lm2_characters.md",
		},
	}
	cmd := NewTypesCmd(svc)
	cmd.SetArgs([]string{"remove", "characters", "001"})
	buf := new(bytes.Buffer)
	cmd.SetOut(buf)
	cmd.SetErr(new(bytes.Buffer))

	err := cmd.Execute()

	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	output := buf.String()
	if !strings.Contains(output, "001_A3F7c9Qx7Lm2_characters.md") {
		t.Errorf("output should contain removed filename, got: %q", output)
	}
}

func TestTypesRemoveCmd_JSONOutput(t *testing.T) {
	svc := &mockTypesService{
		removeResult: &TypesModifyResult{
			Node:     NodeInfo{MP: "001", SID: "A3F7c9Qx7Lm2"},
			Filename: "001_A3F7c9Qx7Lm2_characters.md",
		},
	}
	cmd := NewTypesCmd(svc)
	cmd.SetArgs([]string{"remove", "--json", "characters", "001"})
	buf := new(bytes.Buffer)
	cmd.SetOut(buf)
	cmd.SetErr(new(bytes.Buffer))

	err := cmd.Execute()

	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	var output typesModifyJSONOutput
	if jsonErr := json.Unmarshal(buf.Bytes(), &output); jsonErr != nil {
		t.Fatalf("invalid JSON output: %v\nraw: %s", jsonErr, buf.String())
	}
	if output.Node.MP != "001" {
		t.Errorf("node.mp = %q, want %q", output.Node.MP, "001")
	}
	if output.Node.SID != "A3F7c9Qx7Lm2" {
		t.Errorf("node.sid = %q, want %q", output.Node.SID, "A3F7c9Qx7Lm2")
	}
	if output.Filename != "001_A3F7c9Qx7Lm2_characters.md" {
		t.Errorf("filename = %q, want %q", output.Filename, "001_A3F7c9Qx7Lm2_characters.md")
	}
}

func TestTypesRemoveCmd_CannotRemoveDraft(t *testing.T) {
	svc := &mockTypesService{
		removeErr: fmt.Errorf("cannot remove draft document"),
	}
	cmd := NewTypesCmd(svc)
	cmd.SetArgs([]string{"remove", "draft", "001"})
	cmd.SetOut(new(bytes.Buffer))
	cmd.SetErr(new(bytes.Buffer))

	err := cmd.Execute()

	if err == nil {
		t.Fatal("expected error when removing draft")
	}
	if !strings.Contains(err.Error(), "cannot remove draft") {
		t.Errorf("error should mention draft, got: %v", err)
	}
}

func TestTypesRemoveCmd_ServiceError(t *testing.T) {
	svc := &mockTypesService{
		removeErr: fmt.Errorf("file not found"),
	}
	cmd := NewTypesCmd(svc)
	cmd.SetArgs([]string{"remove", "characters", "001"})
	cmd.SetOut(new(bytes.Buffer))
	cmd.SetErr(new(bytes.Buffer))

	err := cmd.Execute()

	if err == nil {
		t.Fatal("expected error for service failure")
	}
	if !strings.Contains(err.Error(), "file not found") {
		t.Errorf("error should contain cause, got: %v", err)
	}
}

func TestTypesRemoveCmd_MissingArgs(t *testing.T) {
	tests := []struct {
		name string
		args []string
	}{
		{"no args", []string{"remove"}},
		{"type only", []string{"remove", "characters"}},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			svc := &mockTypesService{}
			cmd := NewTypesCmd(svc)
			cmd.SetArgs(tt.args)
			cmd.SetOut(new(bytes.Buffer))
			cmd.SetErr(new(bytes.Buffer))

			err := cmd.Execute()

			if err == nil {
				t.Fatal("expected error for missing arguments")
			}
		})
	}
}
