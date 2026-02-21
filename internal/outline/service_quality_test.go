package outline

import (
	"context"
	"fmt"
	"strconv"
	"strings"
	"testing"

	"github.com/eykd/linemark-go/internal/domain"
)

// --- Issue: countAvailableGaps undercounts available positions ---
//
// countAvailableGaps returns maxNum - len(unique), which only counts gaps
// below the highest occupied number. When no siblings exist (occupied=[]),
// it returns 0 — even though all 999 positions are available.
// This causes promoteChildren to spuriously return ErrInsufficientGaps
// when the target node has no siblings at its parent level.

func TestOutlineService_Delete_Promote_OnlySibling_Succeeds(t *testing.T) {
	// Node 100 is the ONLY root-level node. It has one child (100-100).
	// Promoting should succeed since the entire root level is empty
	// after the target is removed, with 999 available positions.
	files := []string{
		"100_SID001AABB_draft_parent.md",
		"100-100_SID002CCDD_draft_child.md",
	}
	deleter := &fakeFileDeleter{}
	renamer := &fakeFileRenamer{}
	svc := newDeleteTestService(files, deleter, renamer)

	sel, _ := domain.ParseSelector("100")
	result, err := svc.Delete(context.Background(), sel, domain.DeleteModePromote, true)

	if err != nil {
		t.Fatalf("promote should succeed when target is the only sibling, got: %v", err)
	}
	if len(result.FilesDeleted) != 1 {
		t.Errorf("files_deleted = %d, want 1 (parent only)", len(result.FilesDeleted))
	}
	if len(result.FilesRenamed) != 1 {
		t.Errorf("files_renamed = %d, want 1 (promoted child)", len(result.FilesRenamed))
	}
}

func TestCountAvailableGaps_EmptyOccupied(t *testing.T) {
	// When no positions are occupied, all 999 positions are available.
	got := countAvailableGaps(nil)
	if got < 1 {
		t.Errorf("countAvailableGaps(nil) = %d, want >= 1 (all positions available)", got)
	}
}

func TestCountAvailableGaps_SingleOccupied(t *testing.T) {
	// When only position 500 is occupied, positions 1-499 and 501-999 are available.
	got := countAvailableGaps([]int{500})
	want := 998 // 999 total - 1 occupied
	if got != want {
		t.Errorf("countAvailableGaps([500]) = %d, want %d", got, want)
	}
}

// --- Issue: rollbackRenames swallows errors with _ = ---
//
// rollbackRenames (service.go) uses `_ = renamer.RenameFile(...)` which
// violates the project standard "All errors must be handled explicitly
// (no `_` for errors)". When a rollback rename fails, the caller has
// no way to know the filesystem may be in an inconsistent state.
// The error returned by applyRenames should include rollback failure info.

func TestApplyRenames_RollbackFailure_IncludesRollbackError(t *testing.T) {
	// First rename succeeds, second rename fails (triggering rollback),
	// then the rollback rename also fails. The returned error should
	// mention both the original failure and the rollback failure.
	renamer := &rollbackFailingRenamer{
		failOnCall:         1, // second forward rename fails
		forwardErr:         fmt.Errorf("disk full"),
		failRollback:       true,
		rollbackErr:        fmt.Errorf("device removed"),
		calls:              0,
		rollbackInProgress: false,
	}

	renames := map[string]string{
		"old-a.md": "new-a.md",
		"old-b.md": "new-b.md",
	}

	err := applyRenames(context.Background(), renamer, renames)
	if err == nil {
		t.Fatal("expected error from applyRenames")
	}

	// The error must mention the original failure cause
	if !strings.Contains(err.Error(), "disk full") {
		t.Errorf("error should contain original cause 'disk full', got: %v", err)
	}

	// The error should also mention the rollback failure so the user
	// knows the filesystem may be in an inconsistent state
	if !strings.Contains(err.Error(), "rollback") || !strings.Contains(err.Error(), "device removed") {
		t.Errorf("error should mention rollback failure 'device removed', got: %v", err)
	}
}

// rollbackFailingRenamer is a test double that fails on a specific forward
// call and also fails during rollback renames.
type rollbackFailingRenamer struct {
	failOnCall         int
	forwardErr         error
	failRollback       bool
	rollbackErr        error
	calls              int
	rollbackInProgress bool
	renames            [][2]string
}

func (r *rollbackFailingRenamer) RenameFile(_ context.Context, oldName, newName string) error {
	r.renames = append(r.renames, [2]string{oldName, newName})

	// During rollback, fail if configured to do so
	if r.rollbackInProgress && r.failRollback {
		return r.rollbackErr
	}

	// During forward renames, fail on the specified call
	if r.calls == r.failOnCall {
		r.calls++
		r.rollbackInProgress = true
		return r.forwardErr
	}
	r.calls++
	return nil
}

// --- Issue: Move ignores existing children at the target level ---
//
// OutlineService.Move always uses NextSiblingNumber(nil) to pick the
// new node position, completely ignoring children that already exist
// under the target parent. This causes MP collisions and also means
// the --before and --after flags are silently ignored.

func TestOutlineService_Move_AvoidsCollisionWithExistingChildren(t *testing.T) {
	// Target (200) already has a child at 200-100.
	// Moving source (300) under 200 should NOT produce MP 200-100
	// because that would collide with the existing child.
	files := []string{
		"200_SID002CCDD_draft_target.md",
		"200-100_SID003EEFF_draft_existing-child.md",
		"300_SID004GGHH_draft_source.md",
	}
	renamer := &fakeFileRenamer{}
	reader := &fakeDirectoryReader{files: files}
	locker := &mockLocker{}
	svc := NewOutlineService(reader, nil, locker, nil)
	svc.renamer = renamer

	sourceSel, _ := domain.ParseSelector("300")
	targetSel, _ := domain.ParseSelector("200")
	result, err := svc.Move(context.Background(), sourceSel, targetSel, "", "", true)

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	// The new position must not collide with existing child 200-100
	for _, newName := range result.Renames {
		pf, parseErr := domain.ParseFilename(newName)
		if parseErr != nil {
			t.Errorf("invalid new filename %q: %v", newName, parseErr)
			continue
		}
		if pf.MP == "200-100" {
			t.Errorf("moved node got MP 200-100, which collides with existing child SID003EEFF")
		}
	}
}

func TestOutlineService_Move_Before_PlacesNodeBeforeSibling(t *testing.T) {
	// Target (200) has children at 200-100 and 200-500.
	// Move 300 under 200 with before="200-500".
	// Expected: new child MP should be between 100 and 500 (e.g., 200-200 or 200-300).
	files := []string{
		"200_SID002CCDD_draft_target.md",
		"200-100_SID003EEFF_draft_child-one.md",
		"200-500_SID005IIJJ_draft_child-two.md",
		"300_SID004GGHH_draft_source.md",
	}
	renamer := &fakeFileRenamer{}
	reader := &fakeDirectoryReader{files: files}
	locker := &mockLocker{}
	svc := NewOutlineService(reader, nil, locker, nil)
	svc.renamer = renamer

	sourceSel, _ := domain.ParseSelector("300")
	targetSel, _ := domain.ParseSelector("200")
	result, err := svc.Move(context.Background(), sourceSel, targetSel, "200-500", "", true)

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(result.Renames) == 0 {
		t.Fatal("expected at least one rename")
	}

	// Check the new position is before 500 (and after existing child at 100)
	for _, newName := range result.Renames {
		pf, parseErr := domain.ParseFilename(newName)
		if parseErr != nil {
			t.Errorf("invalid new filename %q: %v", newName, parseErr)
			continue
		}
		if strings.HasPrefix(pf.MP, "200-") && pf.Depth == 2 {
			num, _ := strconv.Atoi(pf.PathParts[1])
			if num >= 500 {
				t.Errorf("moved node at %s (num=%d), want num < 500 (--before=200-500)", pf.MP, num)
			}
			if num <= 100 {
				t.Errorf("moved node at %s (num=%d), want num > 100 (after existing child)", pf.MP, num)
			}
		}
	}
}

func TestOutlineService_Move_After_PlacesNodeAfterSibling(t *testing.T) {
	// Target (200) has children at 200-100, 200-300, and 200-500.
	// Move 300 under 200 with after="200-100".
	// Expected: new child should be placed after 200-100, between 100 and 300.
	files := []string{
		"200_SID002CCDD_draft_target.md",
		"200-100_SID003EEFF_draft_child-one.md",
		"200-300_SID005IIJJ_draft_child-two.md",
		"200-500_SID006KKLL_draft_child-three.md",
		"300_SID004GGHH_draft_source.md",
	}
	renamer := &fakeFileRenamer{}
	reader := &fakeDirectoryReader{files: files}
	locker := &mockLocker{}
	svc := NewOutlineService(reader, nil, locker, nil)
	svc.renamer = renamer

	sourceSel, _ := domain.ParseSelector("300")
	targetSel, _ := domain.ParseSelector("200")
	result, err := svc.Move(context.Background(), sourceSel, targetSel, "", "200-100", true)

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(result.Renames) == 0 {
		t.Fatal("expected at least one rename")
	}

	// Check the new position is after 100 and before 300
	for _, newName := range result.Renames {
		pf, parseErr := domain.ParseFilename(newName)
		if parseErr != nil {
			t.Errorf("invalid new filename %q: %v", newName, parseErr)
			continue
		}
		if strings.HasPrefix(pf.MP, "200-") && pf.Depth == 2 {
			num, _ := strconv.Atoi(pf.PathParts[1])
			if num <= 100 {
				t.Errorf("moved node at %s (num=%d), want num > 100 (--after=200-100)", pf.MP, num)
			}
			if num >= 300 {
				t.Errorf("moved node at %s (num=%d), want num < 300 (before next sibling)", pf.MP, num)
			}
		}
	}
}
