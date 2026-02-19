package outline

import (
	"context"
	"errors"
	"testing"
)

func TestOutlineService_ListTypes_ReturnsDocumentTypes(t *testing.T) {
	// Given: A node with draft, notes, and a custom document type
	files := []string{
		"100_SID001AABB_draft_hello.md",
		"100_SID001AABB_notes.md",
		"100_SID001AABB_characters.md",
	}
	reader := &fakeDirectoryReader{files: files}
	svc := NewOutlineService(reader, nil, &mockLocker{}, nil)

	// When: ListTypes is called for that node
	result, err := svc.ListTypes(context.Background(), "100")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	// Then: All document types should be returned
	wantTypes := []string{"characters", "draft", "notes"}
	if len(result.Types) != len(wantTypes) {
		t.Fatalf("types count = %d, want %d; got %v", len(result.Types), len(wantTypes), result.Types)
	}
	for i, want := range wantTypes {
		if result.Types[i] != want {
			t.Errorf("types[%d] = %q, want %q", i, result.Types[i], want)
		}
	}

	// And: Node identification should be populated
	if result.NodeMP != "100" {
		t.Errorf("NodeMP = %q, want %q", result.NodeMP, "100")
	}
	if result.NodeSID != "SID001AABB" {
		t.Errorf("NodeSID = %q, want %q", result.NodeSID, "SID001AABB")
	}
}

func TestOutlineService_ListTypes_NodeNotFound(t *testing.T) {
	reader := &fakeDirectoryReader{files: []string{}}
	svc := NewOutlineService(reader, nil, &mockLocker{}, nil)

	_, err := svc.ListTypes(context.Background(), "999")

	if !errors.Is(err, ErrNodeNotFound) {
		t.Errorf("expected ErrNodeNotFound, got %v", err)
	}
}

func TestOutlineService_AddType_CreatesFile(t *testing.T) {
	// Given: An existing node with draft and notes
	files := []string{
		"100_SID001AABB_draft_hello.md",
		"100_SID001AABB_notes.md",
	}
	reader := &fakeDirectoryReader{files: files}
	writer := &fakeFileWriter{}
	svc := NewOutlineService(reader, writer, &mockLocker{}, nil)

	// When: AddType is called for a new document type
	result, err := svc.AddType(context.Background(), "characters", "100")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	// Then: A new file should be created for the type
	wantFilename := "100_SID001AABB_characters.md"
	if result.Filename != wantFilename {
		t.Errorf("Filename = %q, want %q", result.Filename, wantFilename)
	}

	if _, ok := writer.written[wantFilename]; !ok {
		t.Errorf("expected file %s to be written, got writes: %v", wantFilename, writer.written)
	}

	// And: Node identification should be populated
	if result.NodeMP != "100" {
		t.Errorf("NodeMP = %q, want %q", result.NodeMP, "100")
	}
	if result.NodeSID != "SID001AABB" {
		t.Errorf("NodeSID = %q, want %q", result.NodeSID, "SID001AABB")
	}
}

func TestOutlineService_AddType_NodeNotFound(t *testing.T) {
	reader := &fakeDirectoryReader{files: []string{}}
	svc := NewOutlineService(reader, &fakeFileWriter{}, &mockLocker{}, nil)

	_, err := svc.AddType(context.Background(), "characters", "999")

	if !errors.Is(err, ErrNodeNotFound) {
		t.Errorf("expected ErrNodeNotFound, got %v", err)
	}
}

func TestOutlineService_RemoveType_DeletesFile(t *testing.T) {
	// Given: A node with draft, notes, and a custom document type
	files := []string{
		"100_SID001AABB_draft_hello.md",
		"100_SID001AABB_notes.md",
		"100_SID001AABB_characters.md",
	}
	reader := &fakeDirectoryReader{files: files}
	deleter := &fakeFileDeleter{}
	svc := NewOutlineService(reader, nil, &mockLocker{}, nil)
	svc.deleter = deleter

	// When: RemoveType is called for the custom type
	result, err := svc.RemoveType(context.Background(), "characters", "100")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	// Then: The file should be deleted
	wantFilename := "100_SID001AABB_characters.md"
	if result.Filename != wantFilename {
		t.Errorf("Filename = %q, want %q", result.Filename, wantFilename)
	}

	foundDelete := false
	for _, d := range deleter.deleted {
		if d == wantFilename {
			foundDelete = true
			break
		}
	}
	if !foundDelete {
		t.Errorf("expected file %s to be deleted, got deletes: %v", wantFilename, deleter.deleted)
	}

	// And: Node identification should be populated
	if result.NodeMP != "100" {
		t.Errorf("NodeMP = %q, want %q", result.NodeMP, "100")
	}
}

func TestOutlineService_RemoveType_NodeNotFound(t *testing.T) {
	reader := &fakeDirectoryReader{files: []string{}}
	deleter := &fakeFileDeleter{}
	svc := NewOutlineService(reader, nil, &mockLocker{}, nil)
	svc.deleter = deleter

	_, err := svc.RemoveType(context.Background(), "characters", "999")

	if !errors.Is(err, ErrNodeNotFound) {
		t.Errorf("expected ErrNodeNotFound, got %v", err)
	}
}
