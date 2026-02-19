package outline

import (
	"context"
	"errors"
	"fmt"
	"testing"

	"github.com/eykd/linemark-go/internal/domain"
	"github.com/eykd/linemark-go/internal/lock"
)

// --- Test doubles for new interfaces ---

// fakeDirectoryReader is a test double for the DirectoryReader interface.
type fakeDirectoryReader struct {
	files []string
	err   error
}

func (f *fakeDirectoryReader) ReadDir(_ context.Context) ([]string, error) {
	return f.files, f.err
}

// fakeFileWriter is a test double for the FileWriter interface.
type fakeFileWriter struct {
	written  map[string]string
	writeErr error
}

func (f *fakeFileWriter) WriteFile(_ context.Context, filename, content string) error {
	if f.writeErr != nil {
		return f.writeErr
	}
	if f.written == nil {
		f.written = make(map[string]string)
	}
	f.written[filename] = content
	return nil
}

// fakeSIDReserver is a test double for the SIDReserver interface.
type fakeSIDReserver struct {
	sid string
	err error
}

func (f *fakeSIDReserver) Reserve(_ context.Context) (string, error) {
	return f.sid, f.err
}

// newTestOutlineService creates an OutlineService with all dependencies for testing.
func newTestOutlineService(reader DirectoryReader, writer FileWriter, locker Locker, reserver SIDReserver) *OutlineService {
	return &OutlineService{
		reader:   reader,
		writer:   writer,
		locker:   locker,
		reserver: reserver,
	}
}

// --- Load tests ---

func TestOutlineService_Load(t *testing.T) {
	tests := []struct {
		name         string
		files        []string
		readDirErr   error
		wantNodes    int
		wantFindings int
		wantErr      bool
	}{
		{
			name:      "empty directory returns empty outline",
			files:     []string{},
			wantNodes: 0,
		},
		{
			name:      "single valid file returns one node",
			files:     []string{"001_ABCD1234EF_draft_hello-world.md"},
			wantNodes: 1,
		},
		{
			name: "two files same SID returns one node",
			files: []string{
				"001_ABCD1234EF_draft_hello-world.md",
				"001_ABCD1234EF_notes.md",
			},
			wantNodes: 1,
		},
		{
			name: "two files different SIDs returns two nodes",
			files: []string{
				"001_ABCD1234EF_draft_hello-world.md",
				"002_WXYZ5678GH01_draft_goodbye.md",
			},
			wantNodes: 2,
		},
		{
			name:         "invalid filename produces finding",
			files:        []string{"not-a-valid-file.txt"},
			wantNodes:    0,
			wantFindings: 1,
		},
		{
			name: "mix of valid and invalid files",
			files: []string{
				"001_ABCD1234EF_draft_hello-world.md",
				"bad-file.txt",
			},
			wantNodes:    1,
			wantFindings: 1,
		},
		{
			name:       "ReadDir error is propagated",
			readDirErr: fmt.Errorf("permission denied"),
			wantErr:    true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			reader := &fakeDirectoryReader{files: tt.files, err: tt.readDirErr}
			svc := newTestOutlineService(reader, nil, &mockLocker{}, nil)

			result, err := svc.Load(context.Background())

			if tt.wantErr {
				if err == nil {
					t.Fatal("expected error, got nil")
				}
				return
			}
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if got := len(result.Outline.Nodes); got != tt.wantNodes {
				t.Errorf("nodes = %d, want %d", got, tt.wantNodes)
			}
			if got := len(result.Findings); got != tt.wantFindings {
				t.Errorf("findings = %d, want %d", got, tt.wantFindings)
			}
		})
	}
}

func TestOutlineService_Load_BypassesLocking(t *testing.T) {
	locker := &mockLocker{tryLockErr: lock.ErrAlreadyLocked}
	reader := &fakeDirectoryReader{files: []string{}}
	svc := newTestOutlineService(reader, nil, locker, nil)

	_, err := svc.Load(context.Background())

	if err != nil {
		t.Fatalf("Load should succeed without lock: %v", err)
	}
	if locker.tryLockCalled {
		t.Error("Load should not call TryLock")
	}
}

func TestOutlineService_Load_NodeDetails(t *testing.T) {
	reader := &fakeDirectoryReader{
		files: []string{"100_ABCD1234EF_draft_hello-world.md"},
	}
	svc := newTestOutlineService(reader, nil, &mockLocker{}, nil)

	result, err := svc.Load(context.Background())

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(result.Outline.Nodes) != 1 {
		t.Fatalf("expected 1 node, got %d", len(result.Outline.Nodes))
	}
	node := result.Outline.Nodes[0]
	if node.SID != "ABCD1234EF" {
		t.Errorf("SID = %q, want %q", node.SID, "ABCD1234EF")
	}
	if node.MP.String() != "100" {
		t.Errorf("MP = %q, want %q", node.MP.String(), "100")
	}
	if node.Title != "hello-world" {
		t.Errorf("Title = %q, want %q", node.Title, "hello-world")
	}
	if len(node.Documents) != 1 {
		t.Fatalf("expected 1 document, got %d", len(node.Documents))
	}
	if node.Documents[0].Type != "draft" {
		t.Errorf("DocType = %q, want %q", node.Documents[0].Type, "draft")
	}
}

func TestOutlineService_Load_FindingDetails(t *testing.T) {
	reader := &fakeDirectoryReader{
		files: []string{"not-valid.txt"},
	}
	svc := newTestOutlineService(reader, nil, &mockLocker{}, nil)

	result, err := svc.Load(context.Background())

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(result.Findings) != 1 {
		t.Fatalf("expected 1 finding, got %d", len(result.Findings))
	}
	f := result.Findings[0]
	if f.Type != domain.FindingInvalidFilename {
		t.Errorf("finding type = %q, want %q", f.Type, domain.FindingInvalidFilename)
	}
	if f.Severity != domain.SeverityWarning {
		t.Errorf("finding severity = %q, want %q", f.Severity, domain.SeverityWarning)
	}
	if f.Path != "not-valid.txt" {
		t.Errorf("finding path = %q, want %q", f.Path, "not-valid.txt")
	}
}

// --- Add tests ---

func TestOutlineService_Add(t *testing.T) {
	tests := []struct {
		name         string
		files        []string
		title        string
		parentMP     string
		sid          string
		wantFilename string
	}{
		{
			name:         "adds first node at root level",
			files:        []string{},
			title:        "Hello World",
			parentMP:     "",
			sid:          "ABCD1234EF00",
			wantFilename: "100_ABCD1234EF00_draft_hello-world.md",
		},
		{
			name:         "adds second node after existing",
			files:        []string{"100_WXYZ5678GH00_draft_first.md"},
			title:        "Second Node",
			parentMP:     "",
			sid:          "NEWID12345AB",
			wantFilename: "200_NEWID12345AB_draft_second-node.md",
		},
		{
			name:         "adds child node under parent",
			files:        []string{"100_WXYZ5678GH00_draft_parent.md"},
			title:        "Child Node",
			parentMP:     "100",
			sid:          "CHILD1234567",
			wantFilename: "100-100_CHILD1234567_draft_child-node.md",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			reader := &fakeDirectoryReader{files: tt.files}
			writer := &fakeFileWriter{written: make(map[string]string)}
			locker := &mockLocker{}
			reserver := &fakeSIDReserver{sid: tt.sid}
			svc := newTestOutlineService(reader, writer, locker, reserver)

			result, err := svc.Add(context.Background(), tt.title, tt.parentMP)

			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if result.Filename != tt.wantFilename {
				t.Errorf("filename = %q, want %q", result.Filename, tt.wantFilename)
			}
			if _, ok := writer.written[tt.wantFilename]; !ok {
				t.Errorf("expected file %q to be written", tt.wantFilename)
			}
		})
	}
}

func TestOutlineService_Add_ResultDetails(t *testing.T) {
	reader := &fakeDirectoryReader{files: []string{}}
	writer := &fakeFileWriter{written: make(map[string]string)}
	locker := &mockLocker{}
	reserver := &fakeSIDReserver{sid: "ABCD1234EF00"}
	svc := newTestOutlineService(reader, writer, locker, reserver)

	result, err := svc.Add(context.Background(), "Hello World", "")

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result.SID != "ABCD1234EF00" {
		t.Errorf("SID = %q, want %q", result.SID, "ABCD1234EF00")
	}
	if result.MP != "100" {
		t.Errorf("MP = %q, want %q", result.MP, "100")
	}
	if result.Filename != "100_ABCD1234EF00_draft_hello-world.md" {
		t.Errorf("filename = %q, want %q", result.Filename, "100_ABCD1234EF00_draft_hello-world.md")
	}
}

func TestOutlineService_Add_FileContent(t *testing.T) {
	reader := &fakeDirectoryReader{files: []string{}}
	writer := &fakeFileWriter{written: make(map[string]string)}
	locker := &mockLocker{}
	reserver := &fakeSIDReserver{sid: "ABCD1234EF00"}
	svc := newTestOutlineService(reader, writer, locker, reserver)

	_, err := svc.Add(context.Background(), "Hello World", "")

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	filename := "100_ABCD1234EF00_draft_hello-world.md"
	content, ok := writer.written[filename]
	if !ok {
		t.Fatalf("expected file %q to be written", filename)
	}
	wantContent := "---\ntitle: Hello World\n---\n"
	if content != wantContent {
		t.Errorf("content = %q, want %q", content, wantContent)
	}
}

func TestOutlineService_Add_Locking(t *testing.T) {
	tests := []struct {
		name       string
		tryLockErr error
		wantErr    bool
		wantErrIs  error
		wantUnlock bool
	}{
		{
			name:       "acquires and releases lock",
			wantUnlock: true,
		},
		{
			name:       "fails fast when already locked",
			tryLockErr: lock.ErrAlreadyLocked,
			wantErr:    true,
			wantErrIs:  lock.ErrAlreadyLocked,
			wantUnlock: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			locker := &mockLocker{tryLockErr: tt.tryLockErr}
			reader := &fakeDirectoryReader{files: []string{}}
			writer := &fakeFileWriter{written: make(map[string]string)}
			reserver := &fakeSIDReserver{sid: "ABCD1234EF00"}
			svc := newTestOutlineService(reader, writer, locker, reserver)

			_, err := svc.Add(context.Background(), "Test", "")

			if !locker.tryLockCalled {
				t.Error("Add should call TryLock")
			}
			if tt.wantErr {
				if err == nil {
					t.Fatal("expected error, got nil")
				}
				if tt.wantErrIs != nil && !errors.Is(err, tt.wantErrIs) {
					t.Errorf("error = %v, want %v", err, tt.wantErrIs)
				}
			} else if err != nil {
				t.Errorf("unexpected error: %v", err)
			}
			if locker.unlockCalled != tt.wantUnlock {
				t.Errorf("unlock called = %v, want %v", locker.unlockCalled, tt.wantUnlock)
			}
		})
	}
}

func TestOutlineService_Add_SIDReserverError(t *testing.T) {
	reserveErr := fmt.Errorf("SID collision")
	locker := &mockLocker{}
	reader := &fakeDirectoryReader{files: []string{}}
	writer := &fakeFileWriter{written: make(map[string]string)}
	reserver := &fakeSIDReserver{err: reserveErr}
	svc := newTestOutlineService(reader, writer, locker, reserver)

	_, err := svc.Add(context.Background(), "Test", "")

	if err == nil {
		t.Fatal("expected error from SID reservation")
	}
	if !locker.unlockCalled {
		t.Error("lock should be released after SIDReserver error")
	}
	if len(writer.written) != 0 {
		t.Error("no file should be written when SID reservation fails")
	}
}

func TestOutlineService_Add_FileWriterError(t *testing.T) {
	writeErr := fmt.Errorf("disk full")
	locker := &mockLocker{}
	reader := &fakeDirectoryReader{files: []string{}}
	writer := &fakeFileWriter{writeErr: writeErr}
	reserver := &fakeSIDReserver{sid: "ABCD1234EF00"}
	svc := newTestOutlineService(reader, writer, locker, reserver)

	_, err := svc.Add(context.Background(), "Test", "")

	if err == nil {
		t.Fatal("expected error from FileWriter")
	}
	if !locker.unlockCalled {
		t.Error("lock should be released after FileWriter error")
	}
}

func TestOutlineService_Add_ReadDirError(t *testing.T) {
	readErr := fmt.Errorf("I/O error")
	reader := &fakeDirectoryReader{err: readErr}
	writer := &fakeFileWriter{written: make(map[string]string)}
	locker := &mockLocker{}
	reserver := &fakeSIDReserver{sid: "ABCD1234EF00"}
	svc := newTestOutlineService(reader, writer, locker, reserver)

	_, err := svc.Add(context.Background(), "Test", "")

	if err == nil {
		t.Fatal("expected error from ReadDir during Add")
	}
	if !locker.unlockCalled {
		t.Error("lock should be released after ReadDir error")
	}
}
