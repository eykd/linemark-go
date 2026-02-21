package cmd

import (
	"bytes"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestInitCmd_CreatesLinemarkDir(t *testing.T) {
	tmp := t.TempDir()

	cmd := NewInitCmd(func() (string, error) { return tmp, nil })
	buf := new(bytes.Buffer)
	cmd.SetOut(buf)
	cmd.SetErr(new(bytes.Buffer))

	err := cmd.Execute()

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	info, statErr := os.Stat(filepath.Join(tmp, ".linemark"))
	if statErr != nil {
		t.Fatalf(".linemark dir not created: %v", statErr)
	}
	if !info.IsDir() {
		t.Fatal(".linemark should be a directory")
	}
}

func TestInitCmd_PrintsConfirmation(t *testing.T) {
	tmp := t.TempDir()

	cmd := NewInitCmd(func() (string, error) { return tmp, nil })
	buf := new(bytes.Buffer)
	cmd.SetOut(buf)
	cmd.SetErr(new(bytes.Buffer))

	err := cmd.Execute()

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !strings.Contains(buf.String(), "Initialized linemark project") {
		t.Errorf("expected confirmation message, got: %q", buf.String())
	}
}

func TestInitCmd_AlreadyExists(t *testing.T) {
	tmp := t.TempDir()
	if err := os.Mkdir(filepath.Join(tmp, ".linemark"), 0o755); err != nil {
		t.Fatal(err)
	}

	cmd := NewInitCmd(func() (string, error) { return tmp, nil })
	buf := new(bytes.Buffer)
	cmd.SetOut(buf)
	cmd.SetErr(new(bytes.Buffer))

	err := cmd.Execute()

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !strings.Contains(buf.String(), "already initialized") {
		t.Errorf("expected 'already initialized' message, got: %q", buf.String())
	}
}

func TestInitCmd_GetwdError(t *testing.T) {
	cmd := NewInitCmd(func() (string, error) { return "", os.ErrPermission })
	cmd.SetOut(new(bytes.Buffer))
	cmd.SetErr(new(bytes.Buffer))

	err := cmd.Execute()

	if err == nil {
		t.Fatal("expected error when getwd fails")
	}
}

func TestInitCmd_MkdirAllError(t *testing.T) {
	tmp := t.TempDir()
	// Create a file at the path where .linemark should go to force MkdirAll error
	blockingFile := filepath.Join(tmp, ".linemark")
	if err := os.WriteFile(blockingFile, []byte("block"), 0o444); err != nil {
		t.Fatal(err)
	}

	cmd := NewInitCmd(func() (string, error) { return tmp, nil })
	cmd.SetOut(new(bytes.Buffer))
	cmd.SetErr(new(bytes.Buffer))

	err := cmd.Execute()

	if err == nil {
		t.Fatal("expected error when MkdirAll fails")
	}
}

func TestInitCmd_Metadata(t *testing.T) {
	cmd := NewInitCmd(func() (string, error) { return t.TempDir(), nil })
	if cmd.Use != "init" {
		t.Errorf("Use = %q, want %q", cmd.Use, "init")
	}
	if cmd.Short == "" {
		t.Error("Short description should not be empty")
	}
}
