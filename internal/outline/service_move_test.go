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

// newMoveTestService creates an OutlineService configured for move tests.
func newMoveTestService(files []string, renamer *fakeFileRenamer) *OutlineService {
	reader := &fakeDirectoryReader{files: files}
	locker := &mockLocker{}
	svc := NewOutlineService(reader, nil, locker, nil)
	svc.renamer = renamer
	return svc
}

// moveTestFiles returns a standard fixture:
//
//	100 (part-one, SID001AABB) with notes
//	  100-200 (chapter, SID002CCDD)
//	    100-200-100 (scene, SID003EEFF)
//	200 (part-two, SID004GGHH)
func moveTestFiles() []string {
	return []string{
		"100_SID001AABB_draft_part-one.md",
		"100_SID001AABB_notes.md",
		"100-200_SID002CCDD_draft_chapter.md",
		"100-200-100_SID003EEFF_draft_scene.md",
		"200_SID004GGHH_draft_part-two.md",
	}
}

func TestOutlineService_Move_CycleDetection(t *testing.T) {
	tests := []struct {
		name   string
		source string
		target string
	}{
		{"move to own child", "100", "100-200"},
		{"move to own grandchild", "100", "100-200-100"},
		{"move to self", "100", "100"},
		{"move child to own descendant", "100-200", "100-200-100"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			renamer := &fakeFileRenamer{}
			svc := newMoveTestService(moveTestFiles(), renamer)

			sourceSel, _ := domain.ParseSelector(tt.source)
			targetSel, _ := domain.ParseSelector(tt.target)
			_, err := svc.Move(context.Background(), sourceSel, targetSel, "", "", true)

			if err == nil {
				t.Fatal("expected cycle detection error")
			}
			if !errors.Is(err, ErrCycleDetected) {
				t.Errorf("error = %v, want ErrCycleDetected", err)
			}
			if len(renamer.renames) != 0 {
				t.Errorf("no renames should occur when cycle detected, got %d", len(renamer.renames))
			}
		})
	}
}

func TestOutlineService_Move_CycleDetection_ErrorMessage(t *testing.T) {
	renamer := &fakeFileRenamer{}
	svc := newMoveTestService(moveTestFiles(), renamer)

	sourceSel, _ := domain.ParseSelector("100")
	targetSel, _ := domain.ParseSelector("100-200")
	_, err := svc.Move(context.Background(), sourceSel, targetSel, "", "", true)

	if err == nil {
		t.Fatal("expected error")
	}
	if !strings.Contains(err.Error(), "100") || !strings.Contains(err.Error(), "100-200") {
		t.Errorf("error should mention source and target MPs, got: %v", err)
	}
	if !strings.Contains(err.Error(), "descendant") {
		t.Errorf("error should mention 'descendant', got: %v", err)
	}
}

func TestOutlineService_Move_CycleDetection_BeforeAnyRenames(t *testing.T) {
	// Cycle check must happen before file renames begin (plan.md §Cycle Detection).
	renamer := &fakeFileRenamer{}
	svc := newMoveTestService(moveTestFiles(), renamer)

	sourceSel, _ := domain.ParseSelector("100")
	targetSel, _ := domain.ParseSelector("100-200")
	_, err := svc.Move(context.Background(), sourceSel, targetSel, "", "", true)

	if err == nil {
		t.Fatal("expected cycle detection error")
	}
	if len(renamer.renames) != 0 {
		t.Errorf("cycle should be detected before any renames, got %d renames", len(renamer.renames))
	}
}

func TestOutlineService_Move_NoCycle_Succeeds(t *testing.T) {
	tests := []struct {
		name   string
		source string
		target string
	}{
		{"move to sibling", "100-200", "200"},
		{"move to different subtree", "100-200-100", "200"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			renamer := &fakeFileRenamer{}
			svc := newMoveTestService(moveTestFiles(), renamer)

			sourceSel, _ := domain.ParseSelector(tt.source)
			targetSel, _ := domain.ParseSelector(tt.target)
			result, err := svc.Move(context.Background(), sourceSel, targetSel, "", "", true)

			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if result == nil {
				t.Fatal("expected non-nil result")
			}
			if len(result.Renames) == 0 {
				t.Error("expected at least one rename entry")
			}
		})
	}
}

func TestOutlineService_Move_DryRun_DoesNotMutate(t *testing.T) {
	renamer := &fakeFileRenamer{}
	svc := newMoveTestService(moveTestFiles(), renamer)

	sourceSel, _ := domain.ParseSelector("100-200")
	targetSel, _ := domain.ParseSelector("200")
	result, err := svc.Move(context.Background(), sourceSel, targetSel, "", "", false)

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(result.Renames) == 0 {
		t.Error("expected planned renames in dry-run result")
	}
	if len(renamer.renames) != 0 {
		t.Errorf("actual renames = %d, want 0 (dry run)", len(renamer.renames))
	}
}

func TestOutlineService_Move_DryRun_ShowsPlannedRenames(t *testing.T) {
	renamer := &fakeFileRenamer{}
	svc := newMoveTestService(moveTestFiles(), renamer)

	sourceSel, _ := domain.ParseSelector("100-200-100")
	targetSel, _ := domain.ParseSelector("200")
	result, err := svc.Move(context.Background(), sourceSel, targetSel, "", "", false)

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	oldName := "100-200-100_SID003EEFF_draft_scene.md"
	newName, ok := result.Renames[oldName]
	if !ok {
		t.Fatalf("expected rename for %q, got renames: %v", oldName, result.Renames)
	}
	pf, parseErr := domain.ParseFilename(newName)
	if parseErr != nil {
		t.Fatalf("new filename %q is invalid: %v", newName, parseErr)
	}
	if !strings.HasPrefix(pf.MP, "200-") {
		t.Errorf("new MP = %q, want prefix '200-'", pf.MP)
	}
	if pf.SID != "SID003EEFF" {
		t.Errorf("SID = %q, want SID003EEFF", pf.SID)
	}
}

func TestOutlineService_Move_RenamesDescendants(t *testing.T) {
	renamer := &fakeFileRenamer{}
	svc := newMoveTestService(moveTestFiles(), renamer)

	sourceSel, _ := domain.ParseSelector("100")
	targetSel, _ := domain.ParseSelector("200")
	result, err := svc.Move(context.Background(), sourceSel, targetSel, "", "", true)

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	// Should rename: 100 node files (draft + notes) + 100-200 descendant + 100-200-100 descendant
	if len(result.Renames) != 4 {
		t.Errorf("renames count = %d, want 4 (node + descendants)", len(result.Renames))
	}
	for _, newName := range result.Renames {
		pf, parseErr := domain.ParseFilename(newName)
		if parseErr != nil {
			t.Errorf("new filename %q is invalid: %v", newName, parseErr)
			continue
		}
		if !strings.HasPrefix(pf.MP, "200-") {
			t.Errorf("new MP = %q, want prefix '200-'", pf.MP)
		}
	}
}

func TestOutlineService_Move_PartialFailure_RollsBack(t *testing.T) {
	// Move 100-200 to 200; 100-200 has one child (100-200-100), so 2 renames needed.
	// countingFileRenamer fails on 2nd call → 1st should be rolled back.
	files := moveTestFiles()
	reader := &fakeDirectoryReader{files: files}
	locker := &mockLocker{}
	renamer := &countingFileRenamer{
		failOnCall: 1,
		err:        fmt.Errorf("disk full"),
	}
	svc := NewOutlineService(reader, nil, locker, nil)
	svc.renamer = renamer

	sourceSel, _ := domain.ParseSelector("100-200")
	targetSel, _ := domain.ParseSelector("200")
	_, err := svc.Move(context.Background(), sourceSel, targetSel, "", "", true)

	if err == nil {
		t.Fatal("expected error for partial rename failure")
	}
	if !strings.Contains(err.Error(), "disk full") {
		t.Errorf("error should contain original cause, got: %v", err)
	}
	if len(renamer.renames) < 2 {
		t.Fatalf("expected at least 2 rename calls (1 forward + 1 rollback), got %d", len(renamer.renames))
	}
	forward := renamer.renames[0]
	rollback := renamer.renames[len(renamer.renames)-1]
	if rollback[0] != forward[1] || rollback[1] != forward[0] {
		t.Errorf("rollback should reverse forward rename:\n"+
			"  forward:  %s -> %s\n"+
			"  rollback: %s -> %s",
			forward[0], forward[1], rollback[0], rollback[1])
	}
}

func TestOutlineService_Move_Locking(t *testing.T) {
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
			reader := &fakeDirectoryReader{files: moveTestFiles()}
			renamer := &fakeFileRenamer{}
			svc := NewOutlineService(reader, nil, locker, nil)
			svc.renamer = renamer

			sourceSel, _ := domain.ParseSelector("100-200")
			targetSel, _ := domain.ParseSelector("200")
			_, err := svc.Move(context.Background(), sourceSel, targetSel, "", "", true)

			if !locker.tryLockCalled {
				t.Error("Move should call TryLock")
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

func TestOutlineService_Move_SourceNotFound(t *testing.T) {
	renamer := &fakeFileRenamer{}
	svc := newMoveTestService(moveTestFiles(), renamer)

	sourceSel, _ := domain.ParseSelector("999")
	targetSel, _ := domain.ParseSelector("200")
	_, err := svc.Move(context.Background(), sourceSel, targetSel, "", "", true)

	if err == nil {
		t.Fatal("expected error for source not found")
	}
	if !errors.Is(err, ErrNodeNotFound) {
		t.Errorf("error = %v, want ErrNodeNotFound", err)
	}
}

func TestOutlineService_Move_TargetNotFound(t *testing.T) {
	renamer := &fakeFileRenamer{}
	svc := newMoveTestService(moveTestFiles(), renamer)

	sourceSel, _ := domain.ParseSelector("100")
	targetSel, _ := domain.ParseSelector("999")
	_, err := svc.Move(context.Background(), sourceSel, targetSel, "", "", true)

	if err == nil {
		t.Fatal("expected error for target not found")
	}
	if !errors.Is(err, ErrNodeNotFound) {
		t.Errorf("error = %v, want ErrNodeNotFound", err)
	}
}

func TestOutlineService_Move_ReadDirError(t *testing.T) {
	reader := &fakeDirectoryReader{err: fmt.Errorf("I/O error")}
	locker := &mockLocker{}
	svc := NewOutlineService(reader, nil, locker, nil)

	sourceSel, _ := domain.ParseSelector("100")
	targetSel, _ := domain.ParseSelector("200")
	_, err := svc.Move(context.Background(), sourceSel, targetSel, "", "", true)

	if err == nil {
		t.Fatal("expected error from ReadDir")
	}
	if locker.unlockCalled != true {
		t.Error("lock should be released after ReadDir error")
	}
}

func TestOutlineService_Move_NoSlotAvailable(t *testing.T) {
	// Two adjacent siblings at 200-001 and 200-002 leave no gap between them.
	// Moving with before="200-002" should fail with no slot available.
	files := []string{
		"200_SID002CCDD_draft_target.md",
		"200-001_SID003EEFF_draft_child-one.md",
		"200-002_SID005IIJJ_draft_child-two.md",
		"300_SID004GGHH_draft_source.md",
	}
	renamer := &fakeFileRenamer{}
	reader := &fakeDirectoryReader{files: files}
	locker := &mockLocker{}
	svc := NewOutlineService(reader, nil, locker, nil)
	svc.renamer = renamer

	sourceSel, _ := domain.ParseSelector("300")
	targetSel, _ := domain.ParseSelector("200")
	_, err := svc.Move(context.Background(), sourceSel, targetSel, "200-002", "", true)

	if err == nil {
		t.Fatal("expected error when no slot available between adjacent siblings")
	}
	if !errors.Is(err, domain.ErrNoSlotAvailable) {
		t.Errorf("error = %v, want ErrNoSlotAvailable", err)
	}
}

func TestOutlineService_Move_PreservesSIDs(t *testing.T) {
	renamer := &fakeFileRenamer{}
	svc := newMoveTestService(moveTestFiles(), renamer)

	sourceSel, _ := domain.ParseSelector("100-200")
	targetSel, _ := domain.ParseSelector("200")
	result, err := svc.Move(context.Background(), sourceSel, targetSel, "", "", true)

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	for oldName, newName := range result.Renames {
		oldPF, _ := domain.ParseFilename(oldName)
		newPF, parseErr := domain.ParseFilename(newName)
		if parseErr != nil {
			t.Errorf("new filename %q is invalid: %v", newName, parseErr)
			continue
		}
		if oldPF.SID != newPF.SID {
			t.Errorf("SID changed from %q to %q for %q", oldPF.SID, newPF.SID, oldName)
		}
	}
}
