package cmd

import (
	"bytes"
	"encoding/json"
	"testing"

	"github.com/spf13/cobra"
)

// TestGlobalJSONFlag_ProducesJSONOutput verifies that --json as a root-level
// persistent flag produces valid JSON output for each command.
func TestGlobalJSONFlag_ProducesJSONOutput(t *testing.T) {
	tests := []struct {
		name  string
		args  []string
		setup func() *cobra.Command
	}{
		{
			name: "check",
			args: []string{"--json", "check"},
			setup: func() *cobra.Command {
				root := NewRootCmd()
				root.AddCommand(NewCheckCmd(&mockCheckRunner{
					result: &CheckResult{},
				}))
				return root
			},
		},
		{
			name: "doctor report",
			args: []string{"--json", "doctor"},
			setup: func() *cobra.Command {
				root := NewRootCmd()
				root.AddCommand(NewDoctorCmd(&mockCheckRunner{
					result: &CheckResult{},
				}))
				return root
			},
		},
		{
			name: "doctor apply",
			args: []string{"--json", "doctor", "--apply"},
			setup: func() *cobra.Command {
				root := NewRootCmd()
				root.AddCommand(NewDoctorCmd(
					&mockCheckRunner{result: &CheckResult{}},
					&mockRepairRunner{result: &RepairResult{}},
				))
				return root
			},
		},
		{
			name: "compact",
			args: []string{"--json", "compact"},
			setup: func() *cobra.Command {
				root := NewRootCmd()
				root.AddCommand(NewCompactCmd(&mockCompactRunner{
					result: &CompactResult{},
				}))
				return root
			},
		},
		{
			name: "types list",
			args: []string{"--json", "types", "list", "001"},
			setup: func() *cobra.Command {
				root := NewRootCmd()
				root.AddCommand(NewTypesCmd(&mockTypesService{
					listResult: &TypesListResult{
						Node:  NodeInfo{MP: "001", SID: "A3F7c9Qx7Lm2"},
						Types: []string{"draft"},
					},
				}))
				return root
			},
		},
		{
			name: "types add",
			args: []string{"--json", "types", "add", "notes", "001"},
			setup: func() *cobra.Command {
				root := NewRootCmd()
				root.AddCommand(NewTypesCmd(&mockTypesService{
					addResult: &TypesModifyResult{
						Node:     NodeInfo{MP: "001", SID: "A3F7c9Qx7Lm2"},
						Filename: "001_A3F7c9Qx7Lm2_notes.md",
					},
				}))
				return root
			},
		},
		{
			name: "types remove",
			args: []string{"--json", "types", "remove", "notes", "001"},
			setup: func() *cobra.Command {
				root := NewRootCmd()
				root.AddCommand(NewTypesCmd(&mockTypesService{
					removeResult: &TypesModifyResult{
						Node:     NodeInfo{MP: "001", SID: "A3F7c9Qx7Lm2"},
						Filename: "001_A3F7c9Qx7Lm2_notes.md",
					},
				}))
				return root
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			root := tt.setup()
			buf := new(bytes.Buffer)
			root.SetOut(buf)
			root.SetErr(new(bytes.Buffer))
			root.SetArgs(tt.args)

			err := root.Execute()
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}

			var parsed map[string]interface{}
			if jsonErr := json.Unmarshal(buf.Bytes(), &parsed); jsonErr != nil {
				t.Errorf("expected valid JSON output with global --json flag, got: %s", buf.String())
			}
		})
	}
}

// TestDefaultOutput_IsHumanReadable verifies that without --json flag, commands
// produce human-readable (non-JSON) output by default when run through root.
func TestDefaultOutput_IsHumanReadable(t *testing.T) {
	tests := []struct {
		name  string
		args  []string
		setup func() *cobra.Command
	}{
		{
			name: "check defaults to human",
			args: []string{"check"},
			setup: func() *cobra.Command {
				root := NewRootCmd()
				root.AddCommand(NewCheckCmd(&mockCheckRunner{
					result: &CheckResult{
						Findings: []CheckFinding{
							{Type: FindingSlugDrift, Severity: SeverityWarning, Message: "drift", Path: "a.md"},
						},
					},
				}))
				return root
			},
		},
		{
			name: "compact defaults to human",
			args: []string{"compact"},
			setup: func() *cobra.Command {
				root := NewRootCmd()
				root.AddCommand(NewCompactCmd(&mockCompactRunner{
					result: &CompactResult{
						Renames:       []RenameEntry{{Old: "a.md", New: "b.md"}},
						FilesAffected: 1,
					},
				}))
				return root
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			root := tt.setup()
			buf := new(bytes.Buffer)
			root.SetOut(buf)
			root.SetErr(new(bytes.Buffer))
			root.SetArgs(tt.args)

			_ = root.Execute()

			output := buf.String()
			if output == "" {
				t.Fatal("expected non-empty output")
			}
			var parsed map[string]interface{}
			if json.Unmarshal(buf.Bytes(), &parsed) == nil {
				t.Errorf("expected human-readable output without --json flag, got valid JSON: %s", output)
			}
		})
	}
}
