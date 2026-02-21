package cmd

import (
	"bytes"
	"context"
	"encoding/json"
	"testing"
)

// --- Planned field on result types ---

func TestCompactResult_PlannedField(t *testing.T) {
	tests := []struct {
		name    string
		planned bool
	}{
		{"defaults to false", false},
		{"can be set to true", true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := CompactResult{Planned: tt.planned}
			if result.Planned != tt.planned {
				t.Errorf("Planned = %v, want %v", result.Planned, tt.planned)
			}
		})
	}
}

func TestCompactResult_PlannedField_JSON(t *testing.T) {
	result := CompactResult{
		Renames:       []RenameEntry{{Old: "a.md", New: "b.md"}},
		FilesAffected: 1,
		Planned:       true,
	}
	data, err := json.Marshal(result)
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

func TestTypesModifyResult_PlannedField(t *testing.T) {
	tests := []struct {
		name    string
		planned bool
	}{
		{"defaults to false", false},
		{"can be set to true", true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := TypesModifyResult{Planned: tt.planned}
			if result.Planned != tt.planned {
				t.Errorf("Planned = %v, want %v", result.Planned, tt.planned)
			}
		})
	}
}

func TestTypesModifyResult_PlannedField_JSON(t *testing.T) {
	result := TypesModifyResult{
		Node:     NodeInfo{MP: "001", SID: "ABC123"},
		Filename: "test.md",
		Planned:  true,
	}
	data, err := json.Marshal(result)
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

// --- Compact command --dry-run behavior ---

func TestCompactCmd_DryRunBehavior(t *testing.T) {
	tests := []struct {
		name      string
		args      []string
		wantApply bool
	}{
		{
			name:      "dry-run overrides apply",
			args:      []string{"compact", "--apply", "--dry-run", "--json"},
			wantApply: false,
		},
		{
			name:      "dry-run without apply",
			args:      []string{"compact", "--dry-run", "--json"},
			wantApply: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			runner := &mockCompactRunner{
				result: &CompactResult{
					Renames:       []RenameEntry{{Old: "a.md", New: "b.md"}},
					FilesAffected: 1,
				},
			}
			root := NewRootCmd()
			cmd := NewCompactCmd(runner)
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

			var output CompactResult
			if jsonErr := json.Unmarshal(buf.Bytes(), &output); jsonErr != nil {
				t.Fatalf("invalid JSON: %v\nraw: %s", jsonErr, buf.String())
			}
			if !output.Planned {
				t.Error("result.Planned should be true when --dry-run is active")
			}
		})
	}
}

// --- Types add/remove --dry-run behavior ---

func TestTypesAddCmd_DryRunSetsPlanned(t *testing.T) {
	svc := &mockTypesService{
		addResult: &TypesModifyResult{
			Node:     NodeInfo{MP: "001", SID: "A3F7c9Qx7Lm2"},
			Filename: "001_A3F7c9Qx7Lm2_characters.md",
		},
	}
	root := NewRootCmd()
	typesCmd := NewTypesCmd(svc)
	root.AddCommand(typesCmd)
	buf := new(bytes.Buffer)
	root.SetOut(buf)
	root.SetErr(new(bytes.Buffer))
	root.SetArgs([]string{"types", "add", "--json", "--dry-run", "characters", "001"})

	err := root.Execute()

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	var output struct {
		Planned bool `json:"planned"`
	}
	if jsonErr := json.Unmarshal(buf.Bytes(), &output); jsonErr != nil {
		t.Fatalf("invalid JSON: %v\nraw: %s", jsonErr, buf.String())
	}
	if !output.Planned {
		t.Error("result.planned should be true when --dry-run is active")
	}
}

// --- Types add/remove --dry-run must prevent service mutation ---
//
// spyTypesService records the apply parameter passed to AddType/RemoveType.
// It expects the interface to include an apply bool parameter (like CompactRunner.Compact)
// so that --dry-run can prevent file modifications at the service boundary.
type spyTypesService struct {
	addResult    *TypesModifyResult
	addErr       error
	removeResult *TypesModifyResult
	removeErr    error
	applyPassed  bool
}

func (s *spyTypesService) ListTypes(ctx context.Context, selector string) (*TypesListResult, error) {
	return nil, nil
}

func (s *spyTypesService) AddType(ctx context.Context, docType, selector string, apply bool) (*TypesModifyResult, error) {
	s.applyPassed = apply
	return s.addResult, s.addErr
}

func (s *spyTypesService) RemoveType(ctx context.Context, docType, selector string, apply bool) (*TypesModifyResult, error) {
	s.applyPassed = apply
	return s.removeResult, s.removeErr
}

func TestTypesAddCmd_DryRunPreventsServiceMutation(t *testing.T) {
	tests := []struct {
		name      string
		args      []string
		wantApply bool
	}{
		{
			name:      "dry-run passes apply=false to service",
			args:      []string{"types", "add", "--json", "--dry-run", "characters", "001"},
			wantApply: false,
		},
		{
			name:      "without dry-run passes apply=true to service",
			args:      []string{"types", "add", "--json", "characters", "001"},
			wantApply: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			svc := &spyTypesService{
				addResult: &TypesModifyResult{
					Node:     NodeInfo{MP: "001", SID: "A3F7c9Qx7Lm2"},
					Filename: "001_A3F7c9Qx7Lm2_characters.md",
				},
			}
			root := NewRootCmd()
			typesCmd := NewTypesCmd(svc)
			root.AddCommand(typesCmd)
			buf := new(bytes.Buffer)
			root.SetOut(buf)
			root.SetErr(new(bytes.Buffer))
			root.SetArgs(tt.args)

			err := root.Execute()

			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if svc.applyPassed != tt.wantApply {
				t.Errorf("apply = %v, want %v", svc.applyPassed, tt.wantApply)
			}

			var output CompactResult
			if jsonErr := json.Unmarshal(buf.Bytes(), &output); jsonErr != nil {
				t.Fatalf("invalid JSON: %v\nraw: %s", jsonErr, buf.String())
			}
		})
	}
}

func TestTypesRemoveCmd_DryRunPreventsServiceMutation(t *testing.T) {
	tests := []struct {
		name      string
		args      []string
		wantApply bool
	}{
		{
			name:      "dry-run passes apply=false to service",
			args:      []string{"types", "remove", "--json", "--dry-run", "characters", "001"},
			wantApply: false,
		},
		{
			name:      "without dry-run passes apply=true to service",
			args:      []string{"types", "remove", "--json", "characters", "001"},
			wantApply: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			svc := &spyTypesService{
				removeResult: &TypesModifyResult{
					Node:     NodeInfo{MP: "001", SID: "A3F7c9Qx7Lm2"},
					Filename: "001_A3F7c9Qx7Lm2_characters.md",
				},
			}
			root := NewRootCmd()
			typesCmd := NewTypesCmd(svc)
			root.AddCommand(typesCmd)
			buf := new(bytes.Buffer)
			root.SetOut(buf)
			root.SetErr(new(bytes.Buffer))
			root.SetArgs(tt.args)

			err := root.Execute()

			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if svc.applyPassed != tt.wantApply {
				t.Errorf("apply = %v, want %v", svc.applyPassed, tt.wantApply)
			}

			var output CompactResult
			if jsonErr := json.Unmarshal(buf.Bytes(), &output); jsonErr != nil {
				t.Fatalf("invalid JSON: %v\nraw: %s", jsonErr, buf.String())
			}
		})
	}
}

func TestTypesRemoveCmd_DryRunSetsPlanned(t *testing.T) {
	svc := &mockTypesService{
		removeResult: &TypesModifyResult{
			Node:     NodeInfo{MP: "001", SID: "A3F7c9Qx7Lm2"},
			Filename: "001_A3F7c9Qx7Lm2_characters.md",
		},
	}
	root := NewRootCmd()
	typesCmd := NewTypesCmd(svc)
	root.AddCommand(typesCmd)
	buf := new(bytes.Buffer)
	root.SetOut(buf)
	root.SetErr(new(bytes.Buffer))
	root.SetArgs([]string{"types", "remove", "--json", "--dry-run", "characters", "001"})

	err := root.Execute()

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	var output struct {
		Planned bool `json:"planned"`
	}
	if jsonErr := json.Unmarshal(buf.Bytes(), &output); jsonErr != nil {
		t.Fatalf("invalid JSON: %v\nraw: %s", jsonErr, buf.String())
	}
	if !output.Planned {
		t.Error("result.planned should be true when --dry-run is active")
	}
}
