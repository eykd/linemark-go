package cmd

import (
	"bytes"
	"strings"
	"testing"

	"github.com/spf13/cobra"
)

func TestNilGuard_AddCmd(t *testing.T) {
	cmd := NewAddCmd(nil)
	root := NewRootCmd()
	root.AddCommand(cmd)
	buf := &bytes.Buffer{}
	root.SetOut(buf)
	root.SetArgs([]string{"add", "My Title"})

	err := root.Execute()

	if err == nil {
		t.Fatal("expected error for nil runner")
	}
	if !strings.Contains(err.Error(), "not in a linemark project") {
		t.Errorf("error = %q, want message about not being in a linemark project", err.Error())
	}
}

func TestNilGuard_CheckCmd(t *testing.T) {
	cmd := NewCheckCmd(nil)
	runNilGuardTest(t, cmd, []string{})
}

func TestNilGuard_DoctorCmd_CheckMode(t *testing.T) {
	cmd := NewDoctorCmd(nil)
	runNilGuardTest(t, cmd, []string{})
}

func TestNilGuard_DoctorCmd_RepairMode(t *testing.T) {
	cmd := NewDoctorCmd(nil, nil)
	root := NewRootCmd()
	root.AddCommand(cmd)
	buf := &bytes.Buffer{}
	root.SetOut(buf)
	root.SetArgs([]string{"doctor", "--apply"})

	err := root.Execute()

	if err == nil {
		t.Fatal("expected error for nil runner")
	}
	if !strings.Contains(err.Error(), "not in a linemark project") {
		t.Errorf("error = %q, want message about not being in a linemark project", err.Error())
	}
}

func TestNilGuard_ListCmd(t *testing.T) {
	cmd := NewListCmd(nil)
	runNilGuardTest(t, cmd, []string{})
}

func TestNilGuard_DeleteCmd(t *testing.T) {
	cmd := NewDeleteCmd(nil)
	root := NewRootCmd()
	root.AddCommand(cmd)
	buf := &bytes.Buffer{}
	root.SetOut(buf)
	root.SetArgs([]string{"delete", "100"})

	err := root.Execute()

	if err == nil {
		t.Fatal("expected error for nil runner")
	}
	if !strings.Contains(err.Error(), "not in a linemark project") {
		t.Errorf("error = %q, want message about not being in a linemark project", err.Error())
	}
}

func TestNilGuard_MoveCmd(t *testing.T) {
	cmd := NewMoveCmd(nil)
	root := NewRootCmd()
	root.AddCommand(cmd)
	buf := &bytes.Buffer{}
	root.SetOut(buf)
	root.SetArgs([]string{"move", "100", "--to", "200"})

	err := root.Execute()

	if err == nil {
		t.Fatal("expected error for nil runner")
	}
	if !strings.Contains(err.Error(), "not in a linemark project") {
		t.Errorf("error = %q, want message about not being in a linemark project", err.Error())
	}
}

func TestNilGuard_RenameCmd(t *testing.T) {
	cmd := NewRenameCmd(nil)
	root := NewRootCmd()
	root.AddCommand(cmd)
	buf := &bytes.Buffer{}
	root.SetOut(buf)
	root.SetArgs([]string{"rename", "100", "New Title"})

	err := root.Execute()

	if err == nil {
		t.Fatal("expected error for nil runner")
	}
	if !strings.Contains(err.Error(), "not in a linemark project") {
		t.Errorf("error = %q, want message about not being in a linemark project", err.Error())
	}
}

func TestNilGuard_CompactCmd(t *testing.T) {
	cmd := NewCompactCmd(nil)
	runNilGuardTest(t, cmd, []string{})
}

func TestNilGuard_TypesListCmd(t *testing.T) {
	cmd := NewTypesCmd(nil)
	root := NewRootCmd()
	root.AddCommand(cmd)
	buf := &bytes.Buffer{}
	root.SetOut(buf)
	root.SetArgs([]string{"types", "list", "100"})

	err := root.Execute()

	if err == nil {
		t.Fatal("expected error for nil runner")
	}
	if !strings.Contains(err.Error(), "not in a linemark project") {
		t.Errorf("error = %q, want message about not being in a linemark project", err.Error())
	}
}

func TestNilGuard_TypesAddCmd(t *testing.T) {
	cmd := NewTypesCmd(nil)
	root := NewRootCmd()
	root.AddCommand(cmd)
	buf := &bytes.Buffer{}
	root.SetOut(buf)
	root.SetArgs([]string{"types", "add", "notes", "100"})

	err := root.Execute()

	if err == nil {
		t.Fatal("expected error for nil runner")
	}
	if !strings.Contains(err.Error(), "not in a linemark project") {
		t.Errorf("error = %q, want message about not being in a linemark project", err.Error())
	}
}

func TestNilGuard_TypesRemoveCmd(t *testing.T) {
	cmd := NewTypesCmd(nil)
	root := NewRootCmd()
	root.AddCommand(cmd)
	buf := &bytes.Buffer{}
	root.SetOut(buf)
	root.SetArgs([]string{"types", "remove", "notes", "100"})

	err := root.Execute()

	if err == nil {
		t.Fatal("expected error for nil runner")
	}
	if !strings.Contains(err.Error(), "not in a linemark project") {
		t.Errorf("error = %q, want message about not being in a linemark project", err.Error())
	}
}

func runNilGuardTest(t *testing.T, cmd *cobra.Command, args []string) {
	t.Helper()
	root := NewRootCmd()
	root.AddCommand(cmd)
	buf := &bytes.Buffer{}
	root.SetOut(buf)
	root.SetArgs(append([]string{cmd.Name()}, args...))

	err := root.Execute()

	if err == nil {
		t.Fatal("expected error for nil runner")
	}
	if !strings.Contains(err.Error(), "not in a linemark project") {
		t.Errorf("error = %q, want message about not being in a linemark project", err.Error())
	}
}
