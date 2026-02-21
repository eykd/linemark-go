package outline

import (
	"context"
	"errors"
	"testing"

	"github.com/eykd/linemark-go/internal/lock"
)

func TestOutlineService_Compact(t *testing.T) {
	tests := []struct {
		name        string
		files       []string
		selector    string
		apply       bool
		wantRenames map[string]string
		wantErr     bool
	}{
		{
			name: "renumbers root children at consistent spacing",
			files: []string{
				"101_SIDA12345AB_draft_first.md",
				"101_SIDA12345AB_notes.md",
				"102_SIDB12345AB_draft_second.md",
				"102_SIDB12345AB_notes.md",
				"103_SIDC12345AB_draft_third.md",
				"103_SIDC12345AB_notes.md",
			},
			selector: "",
			apply:    true,
			wantRenames: map[string]string{
				"101_SIDA12345AB_draft_first.md":  "100_SIDA12345AB_draft_first.md",
				"101_SIDA12345AB_notes.md":        "100_SIDA12345AB_notes.md",
				"102_SIDB12345AB_draft_second.md": "200_SIDB12345AB_draft_second.md",
				"102_SIDB12345AB_notes.md":        "200_SIDB12345AB_notes.md",
				"103_SIDC12345AB_draft_third.md":  "300_SIDC12345AB_draft_third.md",
				"103_SIDC12345AB_notes.md":        "300_SIDC12345AB_notes.md",
			},
		},
		{
			name: "scoped to subtree renumbers only children of selector",
			files: []string{
				"100_SIDA12345AB_draft_parent.md",
				"100_SIDA12345AB_notes.md",
				"100-011_SIDB12345AB_draft_child-one.md",
				"100-011_SIDB12345AB_notes.md",
				"100-012_SIDC12345AB_draft_child-two.md",
				"100-012_SIDC12345AB_notes.md",
				"200_SIDD12345AB_draft_other.md",
				"200_SIDD12345AB_notes.md",
			},
			selector: "100",
			apply:    true,
			wantRenames: map[string]string{
				"100-011_SIDB12345AB_draft_child-one.md": "100-100_SIDB12345AB_draft_child-one.md",
				"100-011_SIDB12345AB_notes.md":           "100-100_SIDB12345AB_notes.md",
				"100-012_SIDC12345AB_draft_child-two.md": "100-200_SIDC12345AB_draft_child-two.md",
				"100-012_SIDC12345AB_notes.md":           "100-200_SIDC12345AB_notes.md",
			},
		},
		{
			name: "dry run returns planned renames without executing",
			files: []string{
				"101_SIDA12345AB_draft_first.md",
				"101_SIDA12345AB_notes.md",
			},
			selector: "",
			apply:    false,
			wantRenames: map[string]string{
				"101_SIDA12345AB_draft_first.md": "100_SIDA12345AB_draft_first.md",
				"101_SIDA12345AB_notes.md":       "100_SIDA12345AB_notes.md",
			},
		},
		{
			name: "no renames when already at consistent spacing",
			files: []string{
				"100_SIDA12345AB_draft_first.md",
				"100_SIDA12345AB_notes.md",
				"200_SIDB12345AB_draft_second.md",
				"200_SIDB12345AB_notes.md",
			},
			selector:    "",
			apply:       true,
			wantRenames: map[string]string{},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			reader := &fakeDirectoryReader{files: tt.files}
			renamer := &fakeFileRenamer{}
			svc := NewOutlineService(reader, nil, &mockLocker{}, nil)
			svc.renamer = renamer

			result, err := svc.Compact(context.Background(), tt.selector, tt.apply)

			if tt.wantErr {
				if err == nil {
					t.Fatal("expected error, got nil")
				}
				return
			}
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}

			if len(result.Renames) != len(tt.wantRenames) {
				t.Errorf("renames count = %d, want %d; got %v",
					len(result.Renames), len(tt.wantRenames), result.Renames)
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

			if tt.apply && len(tt.wantRenames) > 0 && len(renamer.renames) == 0 {
				t.Error("expected file renames to be executed when apply=true")
			}
			if !tt.apply && len(renamer.renames) != 0 {
				t.Error("expected no file renames when apply=false")
			}
		})
	}
}

func TestOutlineService_Compact_AcquiresAndReleasesLock(t *testing.T) {
	files := []string{
		"101_SIDA12345AB_draft_first.md",
		"101_SIDA12345AB_notes.md",
	}
	locker := &mockLocker{}
	svc := NewOutlineService(&fakeDirectoryReader{files: files}, nil, locker, nil)
	svc.renamer = &fakeFileRenamer{}

	_, err := svc.Compact(context.Background(), "", true)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if !locker.tryLockCalled {
		t.Error("Compact should acquire lock")
	}
	if !locker.unlockCalled {
		t.Error("Compact should release lock")
	}
}

func TestOutlineService_Compact_FailsWhenLocked(t *testing.T) {
	locker := &mockLocker{tryLockErr: lock.ErrAlreadyLocked}
	svc := NewOutlineService(nil, nil, locker, nil)

	_, err := svc.Compact(context.Background(), "", true)

	if !errors.Is(err, lock.ErrAlreadyLocked) {
		t.Errorf("expected ErrAlreadyLocked, got %v", err)
	}
}

func TestOutlineService_Compact_CompactsDescendantsRecursively(t *testing.T) {
	// A parent node with grandchildren that also need compacting
	files := []string{
		"100_SIDA12345AB_draft_parent.md",
		"100_SIDA12345AB_notes.md",
		"100-011_SIDB12345AB_draft_child.md",
		"100-011_SIDB12345AB_notes.md",
		"100-011-021_SIDC12345AB_draft_grandchild.md",
		"100-011-021_SIDC12345AB_notes.md",
	}
	renamer := &fakeFileRenamer{}
	svc := NewOutlineService(&fakeDirectoryReader{files: files}, nil, &mockLocker{}, nil)
	svc.renamer = renamer

	result, err := svc.Compact(context.Background(), "100", true)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	// Child 011 → 100, grandchild 021 → 100
	wantChildDraft := "100-100_SIDB12345AB_draft_child.md"
	gotChildDraft, ok := result.Renames["100-011_SIDB12345AB_draft_child.md"]
	if !ok {
		t.Fatalf("expected rename for child draft, got renames: %v", result.Renames)
	}
	if gotChildDraft != wantChildDraft {
		t.Errorf("child draft rename = %q, want %q", gotChildDraft, wantChildDraft)
	}
}

func TestOutlineService_Compact_RenumbersParentAndChildSimultaneously(t *testing.T) {
	// When a parent node is renumbered AND its children also need compacting,
	// the rename map must use actual filenames (on disk) as keys and apply
	// both the parent prefix change and child compacting to produce correct targets.
	files := []string{
		"010_SIDA12345AB_draft_prologue.md",
		"010_SIDA12345AB_notes.md",
		"100_SIDB12345AB_draft_chapter-one.md",
		"100_SIDB12345AB_notes.md",
		"100-300_SIDC12345AB_draft_scene-c.md",
		"100-300_SIDC12345AB_notes.md",
	}
	renamer := &fakeFileRenamer{}
	svc := NewOutlineService(&fakeDirectoryReader{files: files}, nil, &mockLocker{}, nil)
	svc.renamer = renamer

	result, err := svc.Compact(context.Background(), "", true)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	wantRenames := map[string]string{
		"010_SIDA12345AB_draft_prologue.md":    "100_SIDA12345AB_draft_prologue.md",
		"010_SIDA12345AB_notes.md":             "100_SIDA12345AB_notes.md",
		"100_SIDB12345AB_draft_chapter-one.md": "200_SIDB12345AB_draft_chapter-one.md",
		"100_SIDB12345AB_notes.md":             "200_SIDB12345AB_notes.md",
		"100-300_SIDC12345AB_draft_scene-c.md": "200-100_SIDC12345AB_draft_scene-c.md",
		"100-300_SIDC12345AB_notes.md":         "200-100_SIDC12345AB_notes.md",
	}

	if len(result.Renames) != len(wantRenames) {
		t.Errorf("renames count = %d, want %d; got %v",
			len(result.Renames), len(wantRenames), result.Renames)
	}

	for oldName, wantNewName := range wantRenames {
		gotNewName, ok := result.Renames[oldName]
		if !ok {
			t.Errorf("expected rename for %s, got renames: %v", oldName, result.Renames)
			continue
		}
		if gotNewName != wantNewName {
			t.Errorf("rename %s = %q, want %q", oldName, gotNewName, wantNewName)
		}
	}
}
