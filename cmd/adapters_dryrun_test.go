package cmd

// Integration tests for --dry-run behavior at the adapter+service boundary.
//
// These tests use the REAL outline.OutlineService (with fake I/O) wired through
// the REAL adapter implementations to verify that apply=false prevents file
// mutations all the way down to the disk layer.
//
// The bugs under test:
//   - typesAdapter.AddType ignores the apply parameter → writes file even in dry-run
//   - typesAdapter.RemoveType ignores the apply parameter → deletes file even in dry-run
//   - addAdapter.Add passes apply to the outline service which does not respect it →
//     writes files even in dry-run

import (
	"context"
	"strings"
	"testing"

	"github.com/eykd/linemark-go/internal/outline"
)

// --- Test doubles for outline interfaces (cmd-side, integration tests only) ---

// dryRunFakeReader returns a fixed set of filenames.
type dryRunFakeReader struct {
	files []string
}

func (r *dryRunFakeReader) ReadDir(_ context.Context) ([]string, error) {
	return r.files, nil
}

// dryRunFakeWriter records every WriteFile call.
type dryRunFakeWriter struct {
	written []string
}

func (w *dryRunFakeWriter) WriteFile(_ context.Context, filename, _ string) error {
	w.written = append(w.written, filename)
	return nil
}

// dryRunFakeDeleter records every DeleteFile call.
type dryRunFakeDeleter struct {
	deleted []string
}

func (d *dryRunFakeDeleter) DeleteFile(_ context.Context, filename string) error {
	d.deleted = append(d.deleted, filename)
	return nil
}

// dryRunFakeLocker always succeeds.
type dryRunFakeLocker struct{}

func (l *dryRunFakeLocker) TryLock(_ context.Context) error { return nil }
func (l *dryRunFakeLocker) Unlock() error                   { return nil }

// dryRunFakeSIDReserver returns a fixed SID.
type dryRunFakeSIDReserver struct {
	sid string
}

func (r *dryRunFakeSIDReserver) Reserve(_ context.Context) (string, error) {
	return r.sid, nil
}

// dryRunStubSlugifier returns a simple lowercased slug.
type dryRunStubSlugifier struct{}

func (s *dryRunStubSlugifier) Slug(title string) string {
	return strings.ToLower(strings.ReplaceAll(title, " ", "-"))
}

// dryRunStubFMHandler is a minimal FrontmatterHandler stub.
type dryRunStubFMHandler struct{}

func (h *dryRunStubFMHandler) GetTitle(input string) (string, error)           { return "", nil }
func (h *dryRunStubFMHandler) SetTitle(input, newTitle string) (string, error) { return input, nil }
func (h *dryRunStubFMHandler) EncodeYAMLValue(s string) string                 { return s }
func (h *dryRunStubFMHandler) Serialize(fm, body string) string                { return fm }

// --- typesAdapter.AddType dry-run integration test ---

// TestTypesAdapter_AddType_DryRunDoesNotWriteFile verifies that when apply=false
// is passed to typesAdapter.AddType, no file is written to disk.
//
// Currently FAILS because typesAdapter.AddType ignores the apply parameter and
// calls svc.AddType without it, causing the service to always write the file.
func TestTypesAdapter_AddType_DryRunDoesNotWriteFile(t *testing.T) {
	files := []string{
		"100_SID001AABB_draft_hello.md",
		"100_SID001AABB_notes.md",
	}
	reader := &dryRunFakeReader{files: files}
	writer := &dryRunFakeWriter{}
	locker := &dryRunFakeLocker{}
	svc := outline.NewOutlineService(reader, writer, locker, nil)
	adapter := &typesAdapter{svc: svc}

	result, err := adapter.AddType(context.Background(), "characters", "100", false)

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(writer.written) != 0 {
		t.Errorf("dry-run (apply=false) should not write any files, but wrote: %v", writer.written)
	}
	if result.Filename == "" {
		t.Error("result.Filename should be populated even in dry-run mode (plan output)")
	}
}

// TestTypesAdapter_AddType_ApplyWritesFile verifies that apply=true causes the file to be written.
// This is a positive control to ensure the test infrastructure works correctly.
func TestTypesAdapter_AddType_ApplyWritesFile(t *testing.T) {
	files := []string{
		"100_SID001AABB_draft_hello.md",
		"100_SID001AABB_notes.md",
	}
	reader := &dryRunFakeReader{files: files}
	writer := &dryRunFakeWriter{}
	locker := &dryRunFakeLocker{}
	svc := outline.NewOutlineService(reader, writer, locker, nil)
	adapter := &typesAdapter{svc: svc}

	_, err := adapter.AddType(context.Background(), "characters", "100", true)

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(writer.written) != 1 {
		t.Errorf("apply=true should write exactly one file, got: %v", writer.written)
	}
}

// --- typesAdapter.RemoveType dry-run integration test ---

// TestTypesAdapter_RemoveType_DryRunDoesNotDeleteFile verifies that when apply=false
// is passed to typesAdapter.RemoveType, no file is deleted from disk.
//
// Currently FAILS because typesAdapter.RemoveType ignores the apply parameter and
// calls svc.RemoveType without it, causing the service to always delete the file.
func TestTypesAdapter_RemoveType_DryRunDoesNotDeleteFile(t *testing.T) {
	files := []string{
		"100_SID001AABB_draft_hello.md",
		"100_SID001AABB_notes.md",
		"100_SID001AABB_characters.md",
	}
	reader := &dryRunFakeReader{files: files}
	deleter := &dryRunFakeDeleter{}
	locker := &dryRunFakeLocker{}
	svc := outline.NewOutlineService(reader, nil, locker, nil, outline.WithDeleter(deleter))
	adapter := &typesAdapter{svc: svc}

	result, err := adapter.RemoveType(context.Background(), "characters", "100", false)

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(deleter.deleted) != 0 {
		t.Errorf("dry-run (apply=false) should not delete any files, but deleted: %v", deleter.deleted)
	}
	if result.Filename == "" {
		t.Error("result.Filename should be populated even in dry-run mode (plan output)")
	}
}

// TestTypesAdapter_RemoveType_ApplyDeletesFile verifies that apply=true causes the file to be deleted.
// This is a positive control to ensure the test infrastructure works correctly.
func TestTypesAdapter_RemoveType_ApplyDeletesFile(t *testing.T) {
	files := []string{
		"100_SID001AABB_draft_hello.md",
		"100_SID001AABB_notes.md",
		"100_SID001AABB_characters.md",
	}
	reader := &dryRunFakeReader{files: files}
	deleter := &dryRunFakeDeleter{}
	locker := &dryRunFakeLocker{}
	svc := outline.NewOutlineService(reader, nil, locker, nil, outline.WithDeleter(deleter))
	adapter := &typesAdapter{svc: svc}

	_, err := adapter.RemoveType(context.Background(), "characters", "100", true)

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(deleter.deleted) != 1 {
		t.Errorf("apply=true should delete exactly one file, got: %v", deleter.deleted)
	}
}

// --- addAdapter.Add dry-run integration test ---

// TestAddAdapter_DryRunDoesNotWriteFiles verifies that when apply=false is passed
// to addAdapter.Add, no files are written to disk.
//
// Currently FAILS because addAdapter.Add calls a.svc.Add without passing apply,
// and outline.OutlineService.Add always writes files regardless.
func TestAddAdapter_DryRunDoesNotWriteFiles(t *testing.T) {
	reader := &dryRunFakeReader{files: []string{}}
	writer := &dryRunFakeWriter{}
	locker := &dryRunFakeLocker{}
	reserver := &dryRunFakeSIDReserver{sid: "ABCD12345678"}
	svc := outline.NewOutlineService(
		reader, writer, locker, reserver,
		outline.WithSlugifier(&dryRunStubSlugifier{}),
		outline.WithFrontmatterHandler(&dryRunStubFMHandler{}),
	)
	adapter := &addAdapter{svc: svc}

	result, err := adapter.Add(context.Background(), "Chapter One", false, Placement{})

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(writer.written) != 0 {
		t.Errorf("dry-run (apply=false) should not write any files, but wrote: %v", writer.written)
	}
	if len(result.FilesPlanned) == 0 {
		t.Error("result.FilesPlanned should contain planned filenames in dry-run mode")
	}
	if len(result.FilesCreated) != 0 {
		t.Errorf("result.FilesCreated should be empty in dry-run mode, got: %v", result.FilesCreated)
	}
}

// TestAddAdapter_ApplyWritesFiles verifies that apply=true causes files to be written.
// This is a positive control to ensure the test infrastructure works correctly.
func TestAddAdapter_ApplyWritesFiles(t *testing.T) {
	reader := &dryRunFakeReader{files: []string{}}
	writer := &dryRunFakeWriter{}
	locker := &dryRunFakeLocker{}
	reserver := &dryRunFakeSIDReserver{sid: "ABCD12345678"}
	svc := outline.NewOutlineService(
		reader, writer, locker, reserver,
		outline.WithSlugifier(&dryRunStubSlugifier{}),
		outline.WithFrontmatterHandler(&dryRunStubFMHandler{}),
	)
	adapter := &addAdapter{svc: svc}

	result, err := adapter.Add(context.Background(), "Chapter One", true, Placement{})

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(writer.written) == 0 {
		t.Error("apply=true should write files to disk")
	}
	if len(result.FilesCreated) == 0 {
		t.Error("result.FilesCreated should be populated when apply=true")
	}
}
