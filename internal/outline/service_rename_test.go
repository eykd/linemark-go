package outline

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"testing"

	"github.com/eykd/linemark-go/internal/lock"
)

// fakeContentReader is a test double for the ContentReader interface.
type fakeContentReader struct {
	contents map[string]string
	err      error
}

func (f *fakeContentReader) ReadFile(_ context.Context, filename string) (string, error) {
	if f.err != nil {
		return "", f.err
	}
	return f.contents[filename], nil
}

func TestOutlineService_Rename(t *testing.T) {
	tests := []struct {
		name         string
		files        []string
		contents     map[string]string
		selector     string
		newTitle     string
		apply        bool
		wantOldTitle string
		wantNewTitle string
		wantRenames  map[string]string
		wantErr      bool
		wantErrIs    error
	}{
		{
			name: "renames draft file with new slug",
			files: []string{
				"100_SID001AABB_draft_old-title.md",
				"100_SID001AABB_notes.md",
			},
			contents: map[string]string{
				"100_SID001AABB_draft_old-title.md": "---\ntitle: Old Title\n---\n",
			},
			selector:     "100",
			newTitle:     "New Title",
			apply:        true,
			wantOldTitle: "Old Title",
			wantNewTitle: "New Title",
			wantRenames: map[string]string{
				"100_SID001AABB_draft_old-title.md": "100_SID001AABB_draft_new-title.md",
			},
		},
		{
			name: "dry run returns planned renames without executing",
			files: []string{
				"100_SID001AABB_draft_old-title.md",
				"100_SID001AABB_notes.md",
			},
			contents: map[string]string{
				"100_SID001AABB_draft_old-title.md": "---\ntitle: Old Title\n---\n",
			},
			selector:     "100",
			newTitle:     "New Title",
			apply:        false,
			wantOldTitle: "Old Title",
			wantNewTitle: "New Title",
			wantRenames: map[string]string{
				"100_SID001AABB_draft_old-title.md": "100_SID001AABB_draft_new-title.md",
			},
		},
		{
			name: "renames draft file by SID selector",
			files: []string{
				"100_SID001AABB_draft_old-title.md",
				"100_SID001AABB_notes.md",
			},
			contents: map[string]string{
				"100_SID001AABB_draft_old-title.md": "---\ntitle: Old Title\n---\n",
			},
			selector:     "SID001AABB",
			newTitle:     "New Title",
			apply:        true,
			wantOldTitle: "Old Title",
			wantNewTitle: "New Title",
			wantRenames: map[string]string{
				"100_SID001AABB_draft_old-title.md": "100_SID001AABB_draft_new-title.md",
			},
		},
		{
			name:      "returns error when node not found",
			files:     []string{},
			selector:  "999",
			newTitle:  "New Title",
			apply:     true,
			wantErr:   true,
			wantErrIs: ErrNodeNotFound,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			reader := &fakeDirectoryReader{files: tt.files}
			renamer := &fakeFileRenamer{}
			writer := &fakeFileWriter{}
			contentReader := &fakeContentReader{contents: tt.contents}
			svc := NewOutlineService(reader, writer, &mockLocker{}, nil)
			svc.renamer = renamer
			svc.contentReader = contentReader

			result, err := svc.Rename(context.Background(), tt.selector, tt.newTitle, tt.apply)

			if tt.wantErr {
				if err == nil {
					t.Fatal("expected error, got nil")
				}
				if tt.wantErrIs != nil && !errors.Is(err, tt.wantErrIs) {
					t.Errorf("error = %v, want %v", err, tt.wantErrIs)
				}
				return
			}
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}

			if result.OldTitle != tt.wantOldTitle {
				t.Errorf("OldTitle = %q, want %q", result.OldTitle, tt.wantOldTitle)
			}
			if result.NewTitle != tt.wantNewTitle {
				t.Errorf("NewTitle = %q, want %q", result.NewTitle, tt.wantNewTitle)
			}

			for oldName, wantNewName := range tt.wantRenames {
				gotNewName, ok := result.Renames[oldName]
				if !ok {
					t.Errorf("expected rename for %s, got renames: %v", oldName, result.Renames)
					continue
				}
				if gotNewName != wantNewName {
					t.Errorf("rename %s = %q, want %q", oldName, gotNewName, wantNewName)
				}
			}

			if tt.apply && len(renamer.renames) == 0 && len(tt.wantRenames) > 0 {
				t.Error("expected file renames to be executed when apply=true")
			}
			if !tt.apply && len(renamer.renames) != 0 {
				t.Error("expected no file renames when apply=false")
			}
		})
	}
}

func TestOutlineService_Rename_UpdatesFrontmatterTitle(t *testing.T) {
	files := []string{
		"100_SID001AABB_draft_old-title.md",
		"100_SID001AABB_notes.md",
	}
	contentReader := &fakeContentReader{
		contents: map[string]string{
			"100_SID001AABB_draft_old-title.md": "---\ntitle: Old Title\n---\nSome content\n",
		},
	}
	writer := &fakeFileWriter{}
	renamer := &fakeFileRenamer{}
	reader := &fakeDirectoryReader{files: files}
	svc := NewOutlineService(reader, writer, &mockLocker{}, nil)
	svc.renamer = renamer
	svc.contentReader = contentReader

	_, err := svc.Rename(context.Background(), "100", "New Title", true)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	// The draft file should be written with the updated frontmatter title
	newFilename := "100_SID001AABB_draft_new-title.md"
	content, ok := writer.written[newFilename]
	if !ok {
		t.Fatalf("expected write to %s, got writes: %v", newFilename, writer.written)
	}
	if !strings.Contains(content, "title: New Title") {
		t.Errorf("content should contain updated title, got: %q", content)
	}
}

func TestOutlineService_Rename_PropagatesReaderError(t *testing.T) {
	reader := &fakeDirectoryReader{err: fmt.Errorf("disk error")}
	svc := NewOutlineService(reader, nil, &mockLocker{}, nil)
	svc.contentReader = &fakeContentReader{}

	_, err := svc.Rename(context.Background(), "100", "New Title", true)

	if err == nil {
		t.Fatal("expected error from ReadDir")
	}
}

func TestOutlineService_Rename_AcquiresAndReleasesLock(t *testing.T) {
	files := []string{
		"100_SID001AABB_draft_hello.md",
		"100_SID001AABB_notes.md",
	}
	contentReader := &fakeContentReader{
		contents: map[string]string{
			"100_SID001AABB_draft_hello.md": "---\ntitle: Hello\n---\n",
		},
	}
	locker := &mockLocker{}
	svc := NewOutlineService(&fakeDirectoryReader{files: files}, &fakeFileWriter{}, locker, nil)
	svc.renamer = &fakeFileRenamer{}
	svc.contentReader = contentReader

	_, err := svc.Rename(context.Background(), "100", "World", true)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if !locker.tryLockCalled {
		t.Error("Rename should acquire lock")
	}
	if !locker.unlockCalled {
		t.Error("Rename should release lock")
	}
}

func TestOutlineService_Rename_FailsWhenLocked(t *testing.T) {
	locker := &mockLocker{tryLockErr: lock.ErrAlreadyLocked}
	svc := NewOutlineService(nil, nil, locker, nil)

	_, err := svc.Rename(context.Background(), "100", "New Title", true)

	if !errors.Is(err, lock.ErrAlreadyLocked) {
		t.Errorf("expected ErrAlreadyLocked, got %v", err)
	}
}
