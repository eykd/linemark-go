package outline

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"testing"

	"github.com/eykd/linemark-go/internal/domain"
	"github.com/eykd/linemark-go/internal/lock"
)

// fakeFileDeleter is a test double for the FileDeleter interface.
type fakeFileDeleter struct {
	deleted    []string
	err        error
	failOnFile string // if set, only this file causes an error
}

func (f *fakeFileDeleter) DeleteFile(_ context.Context, filename string) error {
	if f.failOnFile != "" && filename == f.failOnFile {
		return f.err
	}
	if f.err != nil && f.failOnFile == "" {
		return f.err
	}
	f.deleted = append(f.deleted, filename)
	return nil
}

// fakeFileRenamer is a test double for the FileRenamer interface.
type fakeFileRenamer struct {
	renames    [][2]string // ordered (old, new) pairs
	err        error
	failOnFile string
}

func (f *fakeFileRenamer) RenameFile(_ context.Context, oldName, newName string) error {
	if f.failOnFile != "" && oldName == f.failOnFile {
		return f.err
	}
	if f.err != nil && f.failOnFile == "" {
		return f.err
	}
	f.renames = append(f.renames, [2]string{oldName, newName})
	return nil
}

// newDeleteTestService creates an OutlineService configured for delete tests.
func newDeleteTestService(files []string, deleter *fakeFileDeleter, renamer *fakeFileRenamer) *OutlineService {
	reader := &fakeDirectoryReader{files: files}
	locker := &mockLocker{}
	svc := NewOutlineService(reader, nil, locker, nil)
	svc.deleter = deleter
	svc.renamer = renamer
	return svc
}

func TestOutlineService_Delete_LeafNode(t *testing.T) {
	files := []string{
		"100_SID001AABB_draft_part-one.md",
		"100_SID001AABB_notes.md",
		"200_SID004GGHH_draft_part-two.md",
	}
	deleter := &fakeFileDeleter{}
	svc := newDeleteTestService(files, deleter, nil)

	sel, _ := domain.ParseSelector("100")
	result, err := svc.Delete(context.Background(), sel, domain.DeleteModeDefault, true)

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(result.FilesDeleted) != 2 {
		t.Errorf("files_deleted count = %d, want 2", len(result.FilesDeleted))
	}
	if len(deleter.deleted) != 2 {
		t.Errorf("actual deletions = %d, want 2", len(deleter.deleted))
	}
	if len(result.SIDsPreserved) != 1 || result.SIDsPreserved[0] != "SID001AABB" {
		t.Errorf("sids_preserved = %v, want [SID001AABB]", result.SIDsPreserved)
	}
}

func TestOutlineService_Delete_DefaultMode_NodeWithChildren_Errors(t *testing.T) {
	files := []string{
		"100_SID001AABB_draft_part-one.md",
		"100-100_SID002CCDD_draft_chapter-one.md",
		"200_SID004GGHH_draft_part-two.md",
	}
	deleter := &fakeFileDeleter{}
	svc := newDeleteTestService(files, deleter, nil)

	sel, _ := domain.ParseSelector("100")
	_, err := svc.Delete(context.Background(), sel, domain.DeleteModeDefault, true)

	if err == nil {
		t.Fatal("expected error when deleting node with children in default mode")
	}
	if !errors.Is(err, ErrNodeHasChildren) {
		t.Errorf("error = %v, want ErrNodeHasChildren", err)
	}
	if len(deleter.deleted) != 0 {
		t.Error("no files should be deleted when node has children in default mode")
	}
}

func TestOutlineService_Delete_Recursive_RemovesSubtree(t *testing.T) {
	files := []string{
		"100_SID001AABB_draft_part-one.md",
		"100_SID001AABB_notes.md",
		"100-100_SID002CCDD_draft_chapter-one.md",
		"100-100_SID002CCDD_notes.md",
		"100-200_SID003EEFF_draft_chapter-two.md",
		"100-200_SID003EEFF_notes.md",
		"200_SID004GGHH_draft_part-two.md",
	}
	deleter := &fakeFileDeleter{}
	svc := newDeleteTestService(files, deleter, nil)

	sel, _ := domain.ParseSelector("100")
	result, err := svc.Delete(context.Background(), sel, domain.DeleteModeRecursive, true)

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(result.FilesDeleted) != 6 {
		t.Errorf("files_deleted count = %d, want 6 (parent + 2 children x 2 files)", len(result.FilesDeleted))
	}
	if len(deleter.deleted) != 6 {
		t.Errorf("actual deletions = %d, want 6", len(deleter.deleted))
	}
	// Sibling should be untouched
	for _, f := range deleter.deleted {
		if f == "200_SID004GGHH_draft_part-two.md" {
			t.Error("sibling node should not be deleted in recursive mode")
		}
	}
	if len(result.SIDsPreserved) != 3 {
		t.Errorf("sids_preserved count = %d, want 3", len(result.SIDsPreserved))
	}
}

func TestOutlineService_Delete_Recursive_DeepNesting(t *testing.T) {
	files := []string{
		"100_SID001AABB_draft_root.md",
		"100-100_SID002CCDD_draft_child.md",
		"100-100-100_SID003EEFF_draft_grandchild.md",
	}
	deleter := &fakeFileDeleter{}
	svc := newDeleteTestService(files, deleter, nil)

	sel, _ := domain.ParseSelector("100")
	result, err := svc.Delete(context.Background(), sel, domain.DeleteModeRecursive, true)

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(result.FilesDeleted) != 3 {
		t.Errorf("files_deleted count = %d, want 3 (root + child + grandchild)", len(result.FilesDeleted))
	}
	if len(result.SIDsPreserved) != 3 {
		t.Errorf("sids_preserved count = %d, want 3", len(result.SIDsPreserved))
	}
}

func TestOutlineService_Delete_Promote_RenumbersChildren(t *testing.T) {
	// Node 100 has 2 children; sibling 300 leaves ample gaps at root level
	files := []string{
		"100_SID001AABB_draft_part-one.md",
		"100_SID001AABB_notes.md",
		"100-100_SID002CCDD_draft_chapter-one.md",
		"100-100_SID002CCDD_notes.md",
		"100-200_SID003EEFF_draft_chapter-two.md",
		"100-200_SID003EEFF_notes.md",
		"300_SID004GGHH_draft_part-two.md",
	}
	deleter := &fakeFileDeleter{}
	renamer := &fakeFileRenamer{}
	svc := newDeleteTestService(files, deleter, renamer)

	sel, _ := domain.ParseSelector("100")
	result, err := svc.Delete(context.Background(), sel, domain.DeleteModePromote, true)

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	// Parent files should be deleted
	if len(result.FilesDeleted) != 2 {
		t.Errorf("files_deleted count = %d, want 2 (parent files only)", len(result.FilesDeleted))
	}
	if len(deleter.deleted) != 2 {
		t.Errorf("actual deletions = %d, want 2", len(deleter.deleted))
	}
	// Children should be renamed (promoted to root level): 4 files total
	if len(result.FilesRenamed) != 4 {
		t.Errorf("files_renamed count = %d, want 4 (2 children x 2 files)", len(result.FilesRenamed))
	}
	if len(renamer.renames) != 4 {
		t.Errorf("actual renames = %d, want 4", len(renamer.renames))
	}
	// Verify promoted children are at root level (no dashes in MP segment)
	for _, pair := range renamer.renames {
		newName := pair[1]
		pf, parseErr := domain.ParseFilename(newName)
		if parseErr != nil {
			t.Errorf("renamed file %q is not a valid filename: %v", newName, parseErr)
			continue
		}
		if pf.Depth != 1 {
			t.Errorf("promoted file %q has depth %d, want 1 (root level)", newName, pf.Depth)
		}
	}
	// Only the deleted parent's SID should be preserved
	if len(result.SIDsPreserved) != 1 || result.SIDsPreserved[0] != "SID001AABB" {
		t.Errorf("sids_preserved = %v, want [SID001AABB]", result.SIDsPreserved)
	}
}

func TestOutlineService_Delete_Promote_InsufficientGaps(t *testing.T) {
	// Node 001 has 2 children, but siblings 002 and 003 leave no room
	// After removing 001, occupied = {002, 003}. Only 1 gap available (position 001).
	files := []string{
		"001_SID001AABB_draft_node.md",
		"001-100_SID002CCDD_draft_child-one.md",
		"001-200_SID003EEFF_draft_child-two.md",
		"002_SID004GGHH_draft_sibling-one.md",
		"003_SID005IIJJ_draft_sibling-two.md",
	}
	deleter := &fakeFileDeleter{}
	renamer := &fakeFileRenamer{}
	svc := newDeleteTestService(files, deleter, renamer)

	sel, _ := domain.ParseSelector("001")
	_, err := svc.Delete(context.Background(), sel, domain.DeleteModePromote, true)

	if err == nil {
		t.Fatal("expected error for insufficient sibling gaps")
	}
	if !errors.Is(err, ErrInsufficientGaps) {
		t.Errorf("error = %v, want ErrInsufficientGaps", err)
	}
	// No files should be modified when validation fails early
	if len(deleter.deleted) != 0 {
		t.Errorf("no files should be deleted when promote fails, got %d deletions", len(deleter.deleted))
	}
	if len(renamer.renames) != 0 {
		t.Errorf("no files should be renamed when promote fails, got %d renames", len(renamer.renames))
	}
}

func TestOutlineService_Delete_PartialFailure_ReturnsError(t *testing.T) {
	files := []string{
		"100_SID001AABB_draft_part-one.md",
		"100_SID001AABB_notes.md",
		"100-100_SID002CCDD_draft_chapter-one.md",
		"100-100_SID002CCDD_notes.md",
	}
	deleter := &fakeFileDeleter{
		failOnFile: "100-100_SID002CCDD_notes.md",
		err:        fmt.Errorf("disk I/O error"),
	}
	svc := newDeleteTestService(files, deleter, nil)

	sel, _ := domain.ParseSelector("100")
	_, err := svc.Delete(context.Background(), sel, domain.DeleteModeRecursive, true)

	if err == nil {
		t.Fatal("expected error for partial delete failure")
	}
	if !strings.Contains(err.Error(), "disk I/O error") {
		t.Errorf("error should contain cause, got: %v", err)
	}
}

func TestOutlineService_Delete_Promote_PartialRenameFailure(t *testing.T) {
	files := []string{
		"100_SID001AABB_draft_part-one.md",
		"100-100_SID002CCDD_draft_child-one.md",
		"100-200_SID003EEFF_draft_child-two.md",
		"300_SID004GGHH_draft_sibling.md",
	}
	deleter := &fakeFileDeleter{}
	renamer := &fakeFileRenamer{
		failOnFile: "100-200_SID003EEFF_draft_child-two.md",
		err:        fmt.Errorf("permission denied"),
	}
	svc := newDeleteTestService(files, deleter, renamer)

	sel, _ := domain.ParseSelector("100")
	_, err := svc.Delete(context.Background(), sel, domain.DeleteModePromote, true)

	if err == nil {
		t.Fatal("expected error for partial rename failure")
	}
	if !strings.Contains(err.Error(), "permission denied") {
		t.Errorf("error should contain cause, got: %v", err)
	}
}

func TestOutlineService_Delete_Locking(t *testing.T) {
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
		{
			name:       "propagates TryLock error",
			tryLockErr: fmt.Errorf("permission denied"),
			wantErr:    true,
			wantUnlock: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			locker := &mockLocker{tryLockErr: tt.tryLockErr}
			reader := &fakeDirectoryReader{
				files: []string{"100_SID001AABB_draft_node.md"},
			}
			deleter := &fakeFileDeleter{}
			svc := NewOutlineService(reader, nil, locker, nil)
			svc.deleter = deleter

			sel, _ := domain.ParseSelector("100")
			_, err := svc.Delete(context.Background(), sel, domain.DeleteModeDefault, true)

			if !locker.tryLockCalled {
				t.Error("Delete should call TryLock")
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

func TestOutlineService_Delete_NodeNotFound(t *testing.T) {
	files := []string{"200_SID004GGHH_draft_other.md"}
	deleter := &fakeFileDeleter{}
	svc := newDeleteTestService(files, deleter, nil)

	sel, _ := domain.ParseSelector("100")
	_, err := svc.Delete(context.Background(), sel, domain.DeleteModeDefault, true)

	if err == nil {
		t.Fatal("expected error for node not found")
	}
	if !errors.Is(err, ErrNodeNotFound) {
		t.Errorf("error = %v, want ErrNodeNotFound", err)
	}
	if len(deleter.deleted) != 0 {
		t.Error("no files should be deleted when node not found")
	}
}

func TestOutlineService_Delete_DryRun_DoesNotMutate(t *testing.T) {
	files := []string{
		"100_SID001AABB_draft_part-one.md",
		"100_SID001AABB_notes.md",
	}
	deleter := &fakeFileDeleter{}
	svc := newDeleteTestService(files, deleter, nil)

	sel, _ := domain.ParseSelector("100")
	result, err := svc.Delete(context.Background(), sel, domain.DeleteModeDefault, false)

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(result.FilesDeleted) != 2 {
		t.Errorf("planned files_deleted = %d, want 2", len(result.FilesDeleted))
	}
	if len(deleter.deleted) != 0 {
		t.Errorf("actual deletions = %d, want 0 (dry run should not mutate)", len(deleter.deleted))
	}
}

func TestOutlineService_Delete_Recursive_DryRun(t *testing.T) {
	files := []string{
		"100_SID001AABB_draft_part-one.md",
		"100-100_SID002CCDD_draft_chapter-one.md",
		"100-200_SID003EEFF_draft_chapter-two.md",
	}
	deleter := &fakeFileDeleter{}
	svc := newDeleteTestService(files, deleter, nil)

	sel, _ := domain.ParseSelector("100")
	result, err := svc.Delete(context.Background(), sel, domain.DeleteModeRecursive, false)

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(result.FilesDeleted) != 3 {
		t.Errorf("planned files_deleted = %d, want 3", len(result.FilesDeleted))
	}
	if len(deleter.deleted) != 0 {
		t.Errorf("actual deletions = %d, want 0 (dry run)", len(deleter.deleted))
	}
}

func TestOutlineService_Delete_Promote_DryRun(t *testing.T) {
	files := []string{
		"100_SID001AABB_draft_part-one.md",
		"100-100_SID002CCDD_draft_child.md",
		"300_SID004GGHH_draft_sibling.md",
	}
	deleter := &fakeFileDeleter{}
	renamer := &fakeFileRenamer{}
	svc := newDeleteTestService(files, deleter, renamer)

	sel, _ := domain.ParseSelector("100")
	result, err := svc.Delete(context.Background(), sel, domain.DeleteModePromote, false)

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(result.FilesDeleted) != 1 {
		t.Errorf("planned files_deleted = %d, want 1", len(result.FilesDeleted))
	}
	if len(result.FilesRenamed) != 1 {
		t.Errorf("planned files_renamed = %d, want 1", len(result.FilesRenamed))
	}
	if len(deleter.deleted) != 0 {
		t.Errorf("actual deletions = %d, want 0 (dry run)", len(deleter.deleted))
	}
	if len(renamer.renames) != 0 {
		t.Errorf("actual renames = %d, want 0 (dry run)", len(renamer.renames))
	}
}

func TestOutlineService_Delete_ReadDirError(t *testing.T) {
	reader := &fakeDirectoryReader{err: fmt.Errorf("I/O error")}
	locker := &mockLocker{}
	deleter := &fakeFileDeleter{}
	svc := NewOutlineService(reader, nil, locker, nil)
	svc.deleter = deleter

	sel, _ := domain.ParseSelector("100")
	_, err := svc.Delete(context.Background(), sel, domain.DeleteModeDefault, true)

	if err == nil {
		t.Fatal("expected error from ReadDir")
	}
	if locker.unlockCalled != true {
		t.Error("lock should be released after ReadDir error")
	}
}

func TestOutlineService_Delete_BySID(t *testing.T) {
	files := []string{
		"100_SID001AABB_draft_part-one.md",
		"100_SID001AABB_notes.md",
	}
	deleter := &fakeFileDeleter{}
	svc := newDeleteTestService(files, deleter, nil)

	sel, _ := domain.ParseSelector("SID001AABB")
	result, err := svc.Delete(context.Background(), sel, domain.DeleteModeDefault, true)

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(result.FilesDeleted) != 2 {
		t.Errorf("files_deleted count = %d, want 2", len(result.FilesDeleted))
	}
}
