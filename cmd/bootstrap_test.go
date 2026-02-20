package cmd

import (
	"context"
	"os"
	"path/filepath"
	"testing"

	"github.com/eykd/linemark-go/internal/outline"
)

func TestBootstrapAddAdapter_CreatesLinemarkDirAndDelegates(t *testing.T) {
	tmp := t.TempDir()
	stub := &stubOutlineService{
		addResult: &outline.AddResult{SID: "ABCD12345678", MP: "100", Filename: "100_ABCD12345678_draft_hello.md"},
	}

	adapter := &bootstrapAddAdapter{
		getwd:       func() (string, error) { return tmp, nil },
		wireService: func(root string) (outlineServicer, error) { return stub, nil },
	}

	result, err := adapter.Add(context.Background(), "Hello", true, Placement{})

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	// Verify .linemark/ was created
	info, statErr := os.Stat(filepath.Join(tmp, ".linemark"))
	if statErr != nil {
		t.Fatalf(".linemark dir not created: %v", statErr)
	}
	if !info.IsDir() {
		t.Fatal(".linemark should be a directory")
	}

	// Verify delegation
	if result.Node.SID != "ABCD12345678" {
		t.Errorf("SID = %q, want %q", result.Node.SID, "ABCD12345678")
	}
	if result.Node.MP != "100" {
		t.Errorf("MP = %q, want %q", result.Node.MP, "100")
	}
	if stub.addTitle != "Hello" {
		t.Errorf("title = %q, want %q", stub.addTitle, "Hello")
	}
}

func TestBootstrapAddAdapter_GetwdError(t *testing.T) {
	adapter := &bootstrapAddAdapter{
		getwd: func() (string, error) { return "", os.ErrPermission },
	}

	_, err := adapter.Add(context.Background(), "Hello", true, Placement{})

	if err == nil {
		t.Fatal("expected error when getwd fails")
	}
}

func TestBootstrapAddAdapter_WireServiceError(t *testing.T) {
	tmp := t.TempDir()
	adapter := &bootstrapAddAdapter{
		getwd:       func() (string, error) { return tmp, nil },
		wireService: func(root string) (outlineServicer, error) { return nil, os.ErrPermission },
	}

	_, err := adapter.Add(context.Background(), "Hello", true, Placement{})

	if err == nil {
		t.Fatal("expected error when wireService fails")
	}
}

func TestBootstrapAddAdapter_MkdirAllError(t *testing.T) {
	// Use a path where MkdirAll will fail (file instead of directory)
	tmp := t.TempDir()
	// Create a file at the path where .linemark should go
	blockingFile := filepath.Join(tmp, ".linemark")
	if err := os.WriteFile(blockingFile, []byte("block"), 0o444); err != nil {
		t.Fatal(err)
	}

	adapter := &bootstrapAddAdapter{
		getwd: func() (string, error) { return tmp, nil },
		wireService: func(root string) (outlineServicer, error) {
			return &stubOutlineService{}, nil
		},
	}

	_, err := adapter.Add(context.Background(), "Hello", true, Placement{})

	if err == nil {
		t.Fatal("expected error when MkdirAll fails")
	}
}

func TestBootstrapAddAdapter_PassesProjectRoot(t *testing.T) {
	tmp := t.TempDir()
	var capturedRoot string
	stub := &stubOutlineService{
		addResult: &outline.AddResult{SID: "ABCD12345678", MP: "100", Filename: "100_ABCD12345678_draft_hello.md"},
	}

	adapter := &bootstrapAddAdapter{
		getwd: func() (string, error) { return tmp, nil },
		wireService: func(root string) (outlineServicer, error) {
			capturedRoot = root
			return stub, nil
		},
	}

	_, err := adapter.Add(context.Background(), "Hello", true, Placement{})

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if capturedRoot != tmp {
		t.Errorf("wireService root = %q, want %q", capturedRoot, tmp)
	}
}
