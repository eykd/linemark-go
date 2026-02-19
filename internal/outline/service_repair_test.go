package outline

import (
	"context"
	"testing"

	"github.com/eykd/linemark-go/internal/domain"
)

func TestOutlineService_Repair_CreatesMissingNotes(t *testing.T) {
	// Given: A node with only a draft file (missing notes)
	files := []string{
		"100_SID001AABB_draft_my-title.md",
	}
	reader := &fakeDirectoryReader{files: files}
	writer := &fakeFileWriter{}
	svc := NewOutlineService(reader, writer, &mockLocker{}, nil)

	// When: Repair is called
	result, err := svc.Repair(context.Background())
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	// Then: A notes file should have been created
	wantFilename := "100_SID001AABB_notes.md"
	if _, ok := writer.written[wantFilename]; !ok {
		t.Errorf("expected notes file %s to be created, got writes: %v", wantFilename, writer.written)
	}

	// And: The repair should be reported in results
	foundRepair := false
	for _, repair := range result.Repairs {
		if repair.Type == domain.FindingMissingDocType && repair.New == wantFilename {
			foundRepair = true
			break
		}
	}
	if !foundRepair {
		t.Errorf("expected repair action for missing notes, got: %v", result.Repairs)
	}
}

func TestOutlineService_Repair_FixesSlugDrift(t *testing.T) {
	// Given: A node whose filename slug doesn't match the frontmatter title
	files := []string{
		"100_SID001AABB_draft_wrong-slug.md",
		"100_SID001AABB_notes.md",
	}
	reader := &fakeDirectoryReader{files: files}
	contentReader := &fakeContentReader{
		contents: map[string]string{
			"100_SID001AABB_draft_wrong-slug.md": "---\ntitle: Correct Title\n---\n",
		},
	}
	renamer := &fakeFileRenamer{}
	svc := NewOutlineService(reader, nil, &mockLocker{}, nil)
	svc.renamer = renamer
	svc.contentReader = contentReader

	// When: Repair is called
	result, err := svc.Repair(context.Background())
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	// Then: The draft file should be renamed to match the title slug
	foundRename := false
	for _, repair := range result.Repairs {
		if repair.Type == domain.FindingSlugDrift {
			if repair.Old != "100_SID001AABB_draft_wrong-slug.md" {
				t.Errorf("repair.Old = %q, want %q", repair.Old, "100_SID001AABB_draft_wrong-slug.md")
			}
			if repair.New != "100_SID001AABB_draft_correct-title.md" {
				t.Errorf("repair.New = %q, want %q", repair.New, "100_SID001AABB_draft_correct-title.md")
			}
			foundRename = true
			break
		}
	}
	if !foundRename {
		t.Errorf("expected slug drift repair action, got: %v", result.Repairs)
	}
}

func TestOutlineService_Repair_ReturnsUnrepairedFindings(t *testing.T) {
	// Given: An outline with findings that cannot be auto-repaired (e.g., duplicate SIDs)
	files := []string{
		"100_SID001AABB_draft_first.md",
		"100_SID001AABB_notes.md",
		"200_SID001AABB_draft_duplicate.md",
		"200_SID001AABB_notes.md",
	}
	reader := &fakeDirectoryReader{files: files}
	svc := NewOutlineService(reader, &fakeFileWriter{}, &mockLocker{}, nil)

	// When: Repair is called
	result, err := svc.Repair(context.Background())
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	// Then: Duplicate SID findings should appear as unrepaired
	foundUnrepaired := false
	for _, finding := range result.Unrepaired {
		if finding.Type == domain.FindingDuplicateSID {
			foundUnrepaired = true
			break
		}
	}
	if !foundUnrepaired {
		t.Errorf("expected unrepaired finding for duplicate SID, got: %v", result.Unrepaired)
	}
}

func TestOutlineService_Repair_CleanOutlineReturnsNoActions(t *testing.T) {
	// Given: A clean outline with no issues
	files := []string{
		"100_SID001AABB_draft_hello.md",
		"100_SID001AABB_notes.md",
	}
	reader := &fakeDirectoryReader{files: files}
	contentReader := &fakeContentReader{
		contents: map[string]string{
			"100_SID001AABB_draft_hello.md": "---\ntitle: hello\n---\n",
		},
	}
	svc := NewOutlineService(reader, &fakeFileWriter{}, &mockLocker{}, nil)
	svc.contentReader = contentReader

	// When: Repair is called
	result, err := svc.Repair(context.Background())
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	// Then: No repairs or unrepaired findings
	if len(result.Repairs) != 0 {
		t.Errorf("expected 0 repairs, got %d: %v", len(result.Repairs), result.Repairs)
	}
	if len(result.Unrepaired) != 0 {
		t.Errorf("expected 0 unrepaired, got %d: %v", len(result.Unrepaired), result.Unrepaired)
	}
}
