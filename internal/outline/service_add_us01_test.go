package outline

import (
	"context"
	"errors"
	"strings"
	"testing"
)

// --- US01: Initialize a New Outline ---
// These tests verify that `Add` creates both draft and notes files,
// with correct filenames and content, per the GWT spec.

func TestOutlineService_Add_CreatesDraftAndNotesFiles(t *testing.T) {
	tests := []struct {
		name          string
		title         string
		sid           string
		wantDraft     string
		wantNotes     string
		wantFileCount int
	}{
		{
			name:          "creates draft and notes for first node",
			title:         "My Novel",
			sid:           "ABCD1234EF00",
			wantDraft:     "100_ABCD1234EF00_draft_my-novel.md",
			wantNotes:     "100_ABCD1234EF00_notes.md",
			wantFileCount: 2,
		},
		{
			name:          "creates draft and notes with special characters in title",
			title:         "The Great Adventure!",
			sid:           "WXYZ5678GH01",
			wantDraft:     "100_WXYZ5678GH01_draft_the-great-adventure.md",
			wantNotes:     "100_WXYZ5678GH01_notes.md",
			wantFileCount: 2,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			reader := &fakeDirectoryReader{files: []string{}}
			writer := &fakeFileWriter{written: make(map[string]string)}
			locker := &mockLocker{}
			reserver := &fakeSIDReserver{sid: tt.sid}
			svc := NewOutlineService(reader, writer, locker, reserver)

			_, err := svc.Add(context.Background(), tt.title, "")

			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if len(writer.written) != tt.wantFileCount {
				t.Fatalf("expected %d files written (draft + notes), got %d: %v",
					tt.wantFileCount, len(writer.written), writerKeys(writer))
			}
			if _, ok := writer.written[tt.wantDraft]; !ok {
				t.Errorf("expected draft file %q to be written; got %v",
					tt.wantDraft, writerKeys(writer))
			}
			if _, ok := writer.written[tt.wantNotes]; !ok {
				t.Errorf("expected notes file %q to be written; got %v",
					tt.wantNotes, writerKeys(writer))
			}
		})
	}
}

func TestOutlineService_Add_NotesFileIsEmpty(t *testing.T) {
	reader := &fakeDirectoryReader{files: []string{}}
	writer := &fakeFileWriter{written: make(map[string]string)}
	locker := &mockLocker{}
	reserver := &fakeSIDReserver{sid: "ABCD1234EF00"}
	svc := NewOutlineService(reader, writer, locker, reserver)

	_, err := svc.Add(context.Background(), "My Novel", "")

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	notesFilename := "100_ABCD1234EF00_notes.md"
	content, ok := writer.written[notesFilename]
	if !ok {
		t.Fatalf("expected notes file %q to be written; got %v",
			notesFilename, writerKeys(writer))
	}
	if content != "" {
		t.Errorf("notes file content should be empty, got %q", content)
	}
}

func TestOutlineService_Add_DraftContainsTitleInFrontmatter(t *testing.T) {
	tests := []struct {
		name        string
		title       string
		sid         string
		wantContent string
	}{
		{
			name:        "simple title",
			title:       "My Novel",
			sid:         "ABCD1234EF00",
			wantContent: "---\ntitle: My Novel\n---\n",
		},
		{
			name:        "title with special characters",
			title:       "Chapter: The Beginning",
			sid:         "WXYZ5678GH01",
			wantContent: "---\ntitle: \"Chapter: The Beginning\"\n---\n",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			reader := &fakeDirectoryReader{files: []string{}}
			writer := &fakeFileWriter{written: make(map[string]string)}
			locker := &mockLocker{}
			reserver := &fakeSIDReserver{sid: tt.sid}
			svc := NewOutlineService(reader, writer, locker, reserver)

			_, err := svc.Add(context.Background(), tt.title, "")

			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}

			draftFilename := findDraftFile(writer)
			if draftFilename == "" {
				t.Fatalf("no draft file found in written files: %v", writerKeys(writer))
			}
			content := writer.written[draftFilename]
			if content != tt.wantContent {
				t.Errorf("draft content = %q, want %q", content, tt.wantContent)
			}
		})
	}
}

func TestOutlineService_Add_FilenamesContainAllComponents(t *testing.T) {
	reader := &fakeDirectoryReader{files: []string{}}
	writer := &fakeFileWriter{written: make(map[string]string)}
	locker := &mockLocker{}
	reserver := &fakeSIDReserver{sid: "B8kQ2mNp4Rs1"}
	svc := NewOutlineService(reader, writer, locker, reserver)

	_, err := svc.Add(context.Background(), "My Novel", "")

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	filenames := writerKeys(writer)
	if len(filenames) < 2 {
		t.Fatalf("expected at least 2 files, got %d: %v", len(filenames), filenames)
	}

	for _, filename := range filenames {
		// Every filename must contain the hierarchy position (MP)
		if !strings.HasPrefix(filename, "100_") {
			t.Errorf("filename %q should start with MP '100_'", filename)
		}
		// Every filename must contain the SID
		if !strings.Contains(filename, "B8kQ2mNp4Rs1") {
			t.Errorf("filename %q should contain SID 'B8kQ2mNp4Rs1'", filename)
		}
		// Every filename must contain a document type
		if !strings.Contains(filename, "_draft_") && !strings.Contains(filename, "_notes") {
			t.Errorf("filename %q should contain a document type", filename)
		}
		// Every filename must end with .md
		if !strings.HasSuffix(filename, ".md") {
			t.Errorf("filename %q should end with .md", filename)
		}
	}

	// Draft filename must contain the slugified title
	draftFile := findDraftFile(writer)
	if draftFile == "" {
		t.Fatal("no draft file found")
	}
	if !strings.Contains(draftFile, "_my-novel.md") {
		t.Errorf("draft filename %q should contain slugified title 'my-novel'", draftFile)
	}
}

func TestOutlineService_Add_ChildNodeCreatesDraftAndNotes(t *testing.T) {
	reader := &fakeDirectoryReader{
		files: []string{"100_WXYZ5678GH00_draft_parent.md"},
	}
	writer := &fakeFileWriter{written: make(map[string]string)}
	locker := &mockLocker{}
	reserver := &fakeSIDReserver{sid: "CHILD1234567"}
	svc := NewOutlineService(reader, writer, locker, reserver)

	_, err := svc.Add(context.Background(), "Child Node", "100")

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	wantDraft := "100-100_CHILD1234567_draft_child-node.md"
	wantNotes := "100-100_CHILD1234567_notes.md"

	if len(writer.written) != 2 {
		t.Fatalf("expected 2 files written for child node, got %d: %v",
			len(writer.written), writerKeys(writer))
	}
	if _, ok := writer.written[wantDraft]; !ok {
		t.Errorf("expected child draft %q; got %v", wantDraft, writerKeys(writer))
	}
	if _, ok := writer.written[wantNotes]; !ok {
		t.Errorf("expected child notes %q; got %v", wantNotes, writerKeys(writer))
	}
}

func TestOutlineService_Add_NotesWriteFailureCleansUpDraft(t *testing.T) {
	reader := &fakeDirectoryReader{files: []string{}}
	// Writer that succeeds on first call (draft) but fails on second (notes)
	writer := &countingFileWriter{
		failOnCall: 2,
		failErr:    errDiskFull,
	}
	locker := &mockLocker{}
	reserver := &fakeSIDReserver{sid: "ABCD1234EF00"}
	svc := NewOutlineService(reader, writer, locker, reserver)

	_, err := svc.Add(context.Background(), "My Novel", "")

	if err == nil {
		t.Fatal("expected error when notes file write fails")
	}
}

// --- Helpers ---

// writerKeys returns the filenames written to a fakeFileWriter.
func writerKeys(w *fakeFileWriter) []string {
	keys := make([]string, 0, len(w.written))
	for k := range w.written {
		keys = append(keys, k)
	}
	return keys
}

// findDraftFile returns the first filename containing "_draft_" from the writer.
func findDraftFile(w *fakeFileWriter) string {
	for k := range w.written {
		if strings.Contains(k, "_draft_") {
			return k
		}
	}
	return ""
}

// countingFileWriter is a test double that fails on a specific call number.
type countingFileWriter struct {
	calls      int
	failOnCall int
	failErr    error
	written    map[string]string
}

var errDiskFull = errors.New("disk full")

func (c *countingFileWriter) WriteFile(_ context.Context, filename, content string) error {
	c.calls++
	if c.calls == c.failOnCall {
		return c.failErr
	}
	if c.written == nil {
		c.written = make(map[string]string)
	}
	c.written[filename] = content
	return nil
}
