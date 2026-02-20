package cmd

import (
	"context"
	"errors"
	"testing"

	"github.com/eykd/linemark-go/internal/domain"
	"github.com/eykd/linemark-go/internal/outline"
)

// --- stubOutlineService provides controllable returns for adapter tests ---

type stubOutlineService struct {
	addResult        *outline.AddResult
	addErr           error
	loadResult       *outline.LoadResult
	loadErr          error
	checkResult      *outline.CheckResult
	checkErr         error
	repairResult     *outline.RepairResult
	repairErr        error
	deleteResult     *outline.DeleteResult
	deleteErr        error
	moveResult       *outline.MoveResult
	moveErr          error
	renameResult     *outline.RenameResult
	renameErr        error
	compactResult    *outline.CompactResult
	compactErr       error
	listTypesResult  *outline.ListResult
	listTypesErr     error
	addTypeResult    *outline.ModifyResult
	addTypeErr       error
	removeTypeResult *outline.ModifyResult
	removeTypeErr    error

	// Captured calls
	addTitle     string
	addParentMP  string
	addOpts      []outline.AddOption
	deleteMode   domain.DeleteMode
	deleteSel    domain.Selector
	deleteApply  bool
	moveSrc      domain.Selector
	moveTgt      domain.Selector
	moveBefore   string
	moveAfter    string
	moveApply    bool
	renameSel    string
	renameTitle  string
	renameApply  bool
	compactSel   string
	compactApply bool
	resolvedNode domain.Node
	resolveErr   error
}

func (s *stubOutlineService) Add(ctx context.Context, title, parentMP string, opts ...outline.AddOption) (*outline.AddResult, error) {
	s.addTitle = title
	s.addParentMP = parentMP
	s.addOpts = opts
	return s.addResult, s.addErr
}

func (s *stubOutlineService) Load(ctx context.Context) (*outline.LoadResult, error) {
	return s.loadResult, s.loadErr
}

func (s *stubOutlineService) Check(ctx context.Context) (*outline.CheckResult, error) {
	return s.checkResult, s.checkErr
}

func (s *stubOutlineService) Repair(ctx context.Context) (*outline.RepairResult, error) {
	return s.repairResult, s.repairErr
}

func (s *stubOutlineService) Delete(ctx context.Context, sel domain.Selector, mode domain.DeleteMode, apply bool) (*outline.DeleteResult, error) {
	s.deleteSel = sel
	s.deleteMode = mode
	s.deleteApply = apply
	return s.deleteResult, s.deleteErr
}

func (s *stubOutlineService) Move(ctx context.Context, source, target domain.Selector, before, after string, apply bool) (*outline.MoveResult, error) {
	s.moveSrc = source
	s.moveTgt = target
	s.moveBefore = before
	s.moveAfter = after
	s.moveApply = apply
	return s.moveResult, s.moveErr
}

func (s *stubOutlineService) Rename(ctx context.Context, selector, newTitle string, apply bool) (*outline.RenameResult, error) {
	s.renameSel = selector
	s.renameTitle = newTitle
	s.renameApply = apply
	return s.renameResult, s.renameErr
}

func (s *stubOutlineService) Compact(ctx context.Context, selector string, apply bool) (*outline.CompactResult, error) {
	s.compactSel = selector
	s.compactApply = apply
	return s.compactResult, s.compactErr
}

func (s *stubOutlineService) ResolveSelector(ctx context.Context, sel domain.Selector) (domain.Node, error) {
	return s.resolvedNode, s.resolveErr
}

func (s *stubOutlineService) ListTypes(ctx context.Context, selector string) (*outline.ListResult, error) {
	return s.listTypesResult, s.listTypesErr
}

func (s *stubOutlineService) AddType(ctx context.Context, docType, selector string) (*outline.ModifyResult, error) {
	return s.addTypeResult, s.addTypeErr
}

func (s *stubOutlineService) RemoveType(ctx context.Context, docType, selector string) (*outline.ModifyResult, error) {
	return s.removeTypeResult, s.removeTypeErr
}

// --- addAdapter tests ---

func TestAddAdapter_DefaultPlacement(t *testing.T) {
	stub := &stubOutlineService{
		addResult: &outline.AddResult{SID: "ABCD12345678", MP: "100", Filename: "100_ABCD12345678_draft_hello.md"},
	}
	adapter := &addAdapter{svc: stub}

	result, err := adapter.Add(context.Background(), "Hello", true, Placement{})

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if stub.addParentMP != "" {
		t.Errorf("parentMP = %q, want empty (root)", stub.addParentMP)
	}
	if result.Node.SID != "ABCD12345678" {
		t.Errorf("SID = %q, want %q", result.Node.SID, "ABCD12345678")
	}
	if result.Node.MP != "100" {
		t.Errorf("MP = %q, want %q", result.Node.MP, "100")
	}
	if len(result.FilesCreated) != 1 {
		t.Errorf("files created = %d, want 1", len(result.FilesCreated))
	}
}

func TestAddAdapter_ChildOfPlacement(t *testing.T) {
	stub := &stubOutlineService{
		resolvedNode: domain.Node{
			MP:  mustParseMP("100"),
			SID: "PARENT123456",
		},
		addResult: &outline.AddResult{SID: "CHILD1234567", MP: "100-100", Filename: "100-100_CHILD1234567_draft_child.md"},
	}
	adapter := &addAdapter{svc: stub}

	_, err := adapter.Add(context.Background(), "Child", true, Placement{ChildOf: "100"})

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if stub.addParentMP != "100" {
		t.Errorf("parentMP = %q, want %q", stub.addParentMP, "100")
	}
}

func TestAddAdapter_BeforePlacement(t *testing.T) {
	stub := &stubOutlineService{
		resolvedNode: domain.Node{
			MP:  mustParseMP("200"),
			SID: "TARGET123456",
		},
		addResult: &outline.AddResult{SID: "NEWNODE12345", MP: "150", Filename: "150_NEWNODE12345_draft_new.md"},
	}
	adapter := &addAdapter{svc: stub}

	_, err := adapter.Add(context.Background(), "New", true, Placement{Before: "200"})

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	// Parent should be root (empty) since 200 is at root level
	if stub.addParentMP != "" {
		t.Errorf("parentMP = %q, want empty (root)", stub.addParentMP)
	}
	// Should have AddBefore option set
	if len(stub.addOpts) != 1 {
		t.Errorf("addOpts = %d, want 1", len(stub.addOpts))
	}
}

func TestAddAdapter_ServiceError(t *testing.T) {
	stub := &stubOutlineService{
		addErr: errors.New("disk full"),
	}
	adapter := &addAdapter{svc: stub}

	_, err := adapter.Add(context.Background(), "Hello", true, Placement{})

	if err == nil {
		t.Fatal("expected error")
	}
}

// --- checkAdapter tests ---

func TestCheckAdapter_ConvertsFindings(t *testing.T) {
	stub := &stubOutlineService{
		checkResult: &outline.CheckResult{
			Findings: []domain.Finding{
				{Type: domain.FindingInvalidFilename, Severity: domain.SeverityWarning, Message: "bad file", Path: "foo.txt"},
				{Type: domain.FindingDuplicateSID, Severity: domain.SeverityError, Message: "dup sid"},
			},
		},
	}
	adapter := &checkAdapter{svc: stub}

	result, err := adapter.Check(context.Background())

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(result.Findings) != 2 {
		t.Fatalf("findings = %d, want 2", len(result.Findings))
	}
	if result.Findings[0].Type != FindingInvalidFilename {
		t.Errorf("type = %q, want %q", result.Findings[0].Type, FindingInvalidFilename)
	}
	if result.Findings[0].Severity != SeverityWarning {
		t.Errorf("severity = %q, want %q", result.Findings[0].Severity, SeverityWarning)
	}
}

// --- repairAdapter tests ---

func TestRepairAdapter_ConvertsRepairs(t *testing.T) {
	stub := &stubOutlineService{
		repairResult: &outline.RepairResult{
			Repairs: []outline.RepairAction{
				{Type: domain.FindingSlugDrift, Old: "old.md", New: "new.md"},
			},
			Unrepaired: []domain.Finding{
				{Type: domain.FindingDuplicateSID, Severity: domain.SeverityError, Message: "dup"},
			},
		},
	}
	adapter := &repairAdapter{svc: stub}

	result, err := adapter.Repair(context.Background())

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(result.Repairs) != 1 {
		t.Fatalf("repairs = %d, want 1", len(result.Repairs))
	}
	if result.Repairs[0].Type != FindingSlugDrift {
		t.Errorf("type = %q, want %q", result.Repairs[0].Type, FindingSlugDrift)
	}
	if len(result.Unrepaired) != 1 {
		t.Fatalf("unrepaired = %d, want 1", len(result.Unrepaired))
	}
}

// --- listAdapter tests ---

func TestListAdapter_ConvertsOutline(t *testing.T) {
	stub := &stubOutlineService{
		loadResult: &outline.LoadResult{
			Outline: domain.Outline{
				Nodes: []domain.Node{
					{MP: mustParseMP("100"), SID: "ABC", Title: "First"},
				},
			},
		},
	}
	adapter := &listAdapter{svc: stub}

	result, err := adapter.List(context.Background())

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(result.Outline.Nodes) != 1 {
		t.Fatalf("nodes = %d, want 1", len(result.Outline.Nodes))
	}
}

// --- deleteAdapter tests ---

func TestDeleteAdapter_ConvertsSelectorAndMode(t *testing.T) {
	stub := &stubOutlineService{
		deleteResult: &outline.DeleteResult{
			FilesDeleted:  []string{"100_ABC_draft_hello.md"},
			SIDsPreserved: []string{"ABC"},
		},
	}
	adapter := &deleteAdapter{svc: stub}

	result, err := adapter.Delete(context.Background(), "100", domain.DeleteModeRecursive, true)

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if stub.deleteMode != domain.DeleteModeRecursive {
		t.Errorf("mode = %v, want recursive", stub.deleteMode)
	}
	if stub.deleteApply != true {
		t.Error("apply should be true")
	}
	if len(result.FilesDeleted) != 1 {
		t.Errorf("files deleted = %d, want 1", len(result.FilesDeleted))
	}
}

// --- moveAdapter tests ---

func TestMoveAdapter_PassesThroughArgs(t *testing.T) {
	stub := &stubOutlineService{
		moveResult: &outline.MoveResult{
			Renames: map[string]string{"old.md": "new.md"},
		},
	}
	adapter := &moveAdapter{svc: stub}

	result, err := adapter.Move(context.Background(), "100", "200", "200-100", "", true)

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if stub.moveBefore != "200-100" {
		t.Errorf("before = %q, want %q", stub.moveBefore, "200-100")
	}
	if len(result.Renames) != 1 {
		t.Errorf("renames = %d, want 1", len(result.Renames))
	}
}

// --- renameAdapter tests ---

func TestRenameAdapter_ConvertsResult(t *testing.T) {
	stub := &stubOutlineService{
		renameResult: &outline.RenameResult{
			OldTitle: "Old",
			NewTitle: "New",
			Renames:  map[string]string{"old.md": "new.md"},
		},
	}
	adapter := &renameAdapter{svc: stub}

	result, err := adapter.Rename(context.Background(), "100", "New", true)

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result.Node.OldTitle != "Old" {
		t.Errorf("old title = %q, want %q", result.Node.OldTitle, "Old")
	}
	if len(result.Renames) != 1 {
		t.Errorf("renames = %d, want 1", len(result.Renames))
	}
}

// --- compactAdapter tests ---

func TestCompactAdapter_ConvertsResult(t *testing.T) {
	stub := &stubOutlineService{
		compactResult: &outline.CompactResult{
			Renames: map[string]string{"old.md": "new.md"},
		},
	}
	adapter := &compactAdapter{svc: stub}

	result, err := adapter.Compact(context.Background(), "", true)

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(result.Renames) != 1 {
		t.Errorf("renames = %d, want 1", len(result.Renames))
	}
	if result.FilesAffected != 1 {
		t.Errorf("files affected = %d, want 1", result.FilesAffected)
	}
}

// --- typesAdapter tests ---

func TestTypesAdapter_ListTypes(t *testing.T) {
	stub := &stubOutlineService{
		listTypesResult: &outline.ListResult{
			Types:   []string{"draft", "notes"},
			NodeMP:  "100",
			NodeSID: "ABC",
		},
	}
	adapter := &typesAdapter{svc: stub}

	result, err := adapter.ListTypes(context.Background(), "100")

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(result.Types) != 2 {
		t.Errorf("types = %d, want 2", len(result.Types))
	}
	if result.Node.MP != "100" {
		t.Errorf("MP = %q, want %q", result.Node.MP, "100")
	}
}

func TestTypesAdapter_AddType(t *testing.T) {
	stub := &stubOutlineService{
		addTypeResult: &outline.ModifyResult{
			Filename: "100_ABC_notes.md",
			NodeMP:   "100",
			NodeSID:  "ABC",
		},
	}
	adapter := &typesAdapter{svc: stub}

	result, err := adapter.AddType(context.Background(), "notes", "100", true)

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result.Filename != "100_ABC_notes.md" {
		t.Errorf("filename = %q, want %q", result.Filename, "100_ABC_notes.md")
	}
}

func TestTypesAdapter_RemoveType(t *testing.T) {
	stub := &stubOutlineService{
		removeTypeResult: &outline.ModifyResult{
			Filename: "100_ABC_notes.md",
			NodeMP:   "100",
			NodeSID:  "ABC",
		},
	}
	adapter := &typesAdapter{svc: stub}

	result, err := adapter.RemoveType(context.Background(), "notes", "100", true)

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result.Filename != "100_ABC_notes.md" {
		t.Errorf("filename = %q, want %q", result.Filename, "100_ABC_notes.md")
	}
}

// --- parentMP tests ---

func TestParentMP_RootLevel(t *testing.T) {
	if got := parentMP("100"); got != "" {
		t.Errorf("parentMP(%q) = %q, want empty", "100", got)
	}
}

func TestParentMP_Nested(t *testing.T) {
	if got := parentMP("100-200"); got != "100" {
		t.Errorf("parentMP(%q) = %q, want %q", "100-200", got, "100")
	}
}

func TestParentMP_DeeplyNested(t *testing.T) {
	if got := parentMP("100-200-300"); got != "100-200" {
		t.Errorf("parentMP(%q) = %q, want %q", "100-200-300", got, "100-200")
	}
}

// --- addAdapter error path tests ---

func TestAddAdapter_SiblingOfPlacement(t *testing.T) {
	stub := &stubOutlineService{
		resolvedNode: domain.Node{
			MP:  mustParseMP("100-200"),
			SID: "SIBLING12345",
		},
		addResult: &outline.AddResult{SID: "NEWNODE12345", MP: "100-300", Filename: "100-300_NEWNODE12345_draft_new.md"},
	}
	adapter := &addAdapter{svc: stub}

	_, err := adapter.Add(context.Background(), "New", true, Placement{SiblingOf: "100-200"})

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if stub.addParentMP != "100" {
		t.Errorf("parentMP = %q, want %q", stub.addParentMP, "100")
	}
}

func TestAddAdapter_AfterPlacement(t *testing.T) {
	stub := &stubOutlineService{
		resolvedNode: domain.Node{
			MP:  mustParseMP("100"),
			SID: "TARGET123456",
		},
		addResult: &outline.AddResult{SID: "NEWNODE12345", MP: "200", Filename: "200_NEWNODE12345_draft_new.md"},
	}
	adapter := &addAdapter{svc: stub}

	_, err := adapter.Add(context.Background(), "New", true, Placement{After: "100"})

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if stub.addParentMP != "" {
		t.Errorf("parentMP = %q, want empty (root)", stub.addParentMP)
	}
	if len(stub.addOpts) != 1 {
		t.Errorf("addOpts = %d, want 1", len(stub.addOpts))
	}
}

func TestAddAdapter_DryRun(t *testing.T) {
	stub := &stubOutlineService{
		addResult: &outline.AddResult{SID: "ABCD12345678", MP: "100", Filename: "100_ABCD12345678_draft_hello.md"},
	}
	adapter := &addAdapter{svc: stub}

	result, err := adapter.Add(context.Background(), "Hello", false, Placement{})

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(result.FilesPlanned) != 1 {
		t.Errorf("files planned = %d, want 1", len(result.FilesPlanned))
	}
	if len(result.FilesCreated) != 0 {
		t.Errorf("files created = %d, want 0", len(result.FilesCreated))
	}
}

func TestAddAdapter_ChildOfResolveError(t *testing.T) {
	stub := &stubOutlineService{
		resolveErr: errors.New("node not found"),
	}
	adapter := &addAdapter{svc: stub}

	_, err := adapter.Add(context.Background(), "Child", true, Placement{ChildOf: "999"})

	if err == nil {
		t.Fatal("expected error")
	}
}

func TestAddAdapter_SiblingOfResolveError(t *testing.T) {
	stub := &stubOutlineService{
		resolveErr: errors.New("node not found"),
	}
	adapter := &addAdapter{svc: stub}

	_, err := adapter.Add(context.Background(), "Sibling", true, Placement{SiblingOf: "999"})

	if err == nil {
		t.Fatal("expected error")
	}
}

func TestAddAdapter_BeforeResolveError(t *testing.T) {
	stub := &stubOutlineService{
		resolveErr: errors.New("node not found"),
	}
	adapter := &addAdapter{svc: stub}

	_, err := adapter.Add(context.Background(), "Before", true, Placement{Before: "999"})

	if err == nil {
		t.Fatal("expected error")
	}
}

func TestAddAdapter_AfterResolveError(t *testing.T) {
	stub := &stubOutlineService{
		resolveErr: errors.New("node not found"),
	}
	adapter := &addAdapter{svc: stub}

	_, err := adapter.Add(context.Background(), "After", true, Placement{After: "999"})

	if err == nil {
		t.Fatal("expected error")
	}
}

// --- error path tests for other adapters ---

func TestCheckAdapter_ServiceError(t *testing.T) {
	stub := &stubOutlineService{checkErr: errors.New("check failed")}
	adapter := &checkAdapter{svc: stub}

	_, err := adapter.Check(context.Background())

	if err == nil {
		t.Fatal("expected error")
	}
}

func TestRepairAdapter_ServiceError(t *testing.T) {
	stub := &stubOutlineService{repairErr: errors.New("repair failed")}
	adapter := &repairAdapter{svc: stub}

	_, err := adapter.Repair(context.Background())

	if err == nil {
		t.Fatal("expected error")
	}
}

func TestListAdapter_ServiceError(t *testing.T) {
	stub := &stubOutlineService{loadErr: errors.New("load failed")}
	adapter := &listAdapter{svc: stub}

	_, err := adapter.List(context.Background())

	if err == nil {
		t.Fatal("expected error")
	}
}

func TestDeleteAdapter_InvalidSelector(t *testing.T) {
	adapter := &deleteAdapter{svc: &stubOutlineService{}}

	_, err := adapter.Delete(context.Background(), "bad!selector", domain.DeleteModeDefault, true)

	if err == nil {
		t.Fatal("expected error for invalid selector")
	}
}

func TestDeleteAdapter_ServiceError(t *testing.T) {
	stub := &stubOutlineService{deleteErr: errors.New("delete failed")}
	adapter := &deleteAdapter{svc: stub}

	_, err := adapter.Delete(context.Background(), "100", domain.DeleteModeDefault, true)

	if err == nil {
		t.Fatal("expected error")
	}
}

func TestMoveAdapter_InvalidSourceSelector(t *testing.T) {
	adapter := &moveAdapter{svc: &stubOutlineService{}}

	_, err := adapter.Move(context.Background(), "bad!selector", "200", "", "", true)

	if err == nil {
		t.Fatal("expected error for invalid source selector")
	}
}

func TestMoveAdapter_InvalidTargetSelector(t *testing.T) {
	adapter := &moveAdapter{svc: &stubOutlineService{}}

	_, err := adapter.Move(context.Background(), "100", "bad!selector", "", "", true)

	if err == nil {
		t.Fatal("expected error for invalid target selector")
	}
}

func TestMoveAdapter_ServiceError(t *testing.T) {
	stub := &stubOutlineService{moveErr: errors.New("move failed")}
	adapter := &moveAdapter{svc: stub}

	_, err := adapter.Move(context.Background(), "100", "200", "", "", true)

	if err == nil {
		t.Fatal("expected error")
	}
}

func TestRenameAdapter_ServiceError(t *testing.T) {
	stub := &stubOutlineService{renameErr: errors.New("rename failed")}
	adapter := &renameAdapter{svc: stub}

	_, err := adapter.Rename(context.Background(), "100", "New", true)

	if err == nil {
		t.Fatal("expected error")
	}
}

func TestCompactAdapter_ServiceError(t *testing.T) {
	stub := &stubOutlineService{compactErr: errors.New("compact failed")}
	adapter := &compactAdapter{svc: stub}

	_, err := adapter.Compact(context.Background(), "", true)

	if err == nil {
		t.Fatal("expected error")
	}
}

func TestTypesAdapter_ListTypesError(t *testing.T) {
	stub := &stubOutlineService{listTypesErr: errors.New("list failed")}
	adapter := &typesAdapter{svc: stub}

	_, err := adapter.ListTypes(context.Background(), "100")

	if err == nil {
		t.Fatal("expected error")
	}
}

func TestTypesAdapter_AddTypeError(t *testing.T) {
	stub := &stubOutlineService{addTypeErr: errors.New("add failed")}
	adapter := &typesAdapter{svc: stub}

	_, err := adapter.AddType(context.Background(), "notes", "100", true)

	if err == nil {
		t.Fatal("expected error")
	}
}

func TestTypesAdapter_RemoveTypeError(t *testing.T) {
	stub := &stubOutlineService{removeTypeErr: errors.New("remove failed")}
	adapter := &typesAdapter{svc: stub}

	_, err := adapter.RemoveType(context.Background(), "notes", "100", true)

	if err == nil {
		t.Fatal("expected error")
	}
}

// mustParseMP is a test helper that panics on invalid MP.
func mustParseMP(s string) domain.MaterializedPath {
	mp, err := domain.NewMaterializedPath(s)
	if err != nil {
		panic(err)
	}
	return mp
}
