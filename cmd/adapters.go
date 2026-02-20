package cmd

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/eykd/linemark-go/internal/domain"
	"github.com/eykd/linemark-go/internal/outline"
)

// outlineServicer abstracts the outline.OutlineService methods used by adapters.
type outlineServicer interface {
	Add(ctx context.Context, title, parentMP string, opts ...outline.AddOption) (*outline.AddResult, error)
	Load(ctx context.Context) (*outline.LoadResult, error)
	Check(ctx context.Context) (*outline.CheckResult, error)
	Repair(ctx context.Context) (*outline.RepairResult, error)
	Delete(ctx context.Context, sel domain.Selector, mode domain.DeleteMode, apply bool) (*outline.DeleteResult, error)
	Move(ctx context.Context, source, target domain.Selector, before, after string, apply bool) (*outline.MoveResult, error)
	Rename(ctx context.Context, selector, newTitle string, apply bool) (*outline.RenameResult, error)
	Compact(ctx context.Context, selector string, apply bool) (*outline.CompactResult, error)
	ResolveSelector(ctx context.Context, sel domain.Selector) (domain.Node, error)
	ListTypes(ctx context.Context, selector string) (*outline.ListResult, error)
	AddType(ctx context.Context, docType, selector string) (*outline.ModifyResult, error)
	RemoveType(ctx context.Context, docType, selector string) (*outline.ModifyResult, error)
}

// parentMP returns the parent MP of the given MP, or "" for root-level.
func parentMP(mp string) string {
	if i := strings.LastIndex(mp, "-"); i >= 0 {
		return mp[:i]
	}
	return ""
}

// --- addAdapter ---

type addAdapter struct {
	svc outlineServicer
}

func (a *addAdapter) Add(ctx context.Context, title string, apply bool, p Placement) (*AddResult, error) {
	var parentMPStr string
	var opts []outline.AddOption

	switch {
	case p.ChildOf != "":
		sel, _ := domain.ParseSelector(p.ChildOf)
		node, err := a.svc.ResolveSelector(ctx, sel)
		if err != nil {
			return nil, err
		}
		parentMPStr = node.MP.String()

	case p.SiblingOf != "":
		sel, _ := domain.ParseSelector(p.SiblingOf)
		node, err := a.svc.ResolveSelector(ctx, sel)
		if err != nil {
			return nil, err
		}
		parentMPStr = parentMP(node.MP.String())

	case p.Before != "":
		sel, _ := domain.ParseSelector(p.Before)
		node, err := a.svc.ResolveSelector(ctx, sel)
		if err != nil {
			return nil, err
		}
		parentMPStr = parentMP(node.MP.String())
		opts = append(opts, outline.AddBefore(node.MP.String()))

	case p.After != "":
		sel, _ := domain.ParseSelector(p.After)
		node, err := a.svc.ResolveSelector(ctx, sel)
		if err != nil {
			return nil, err
		}
		parentMPStr = parentMP(node.MP.String())
		opts = append(opts, outline.AddAfter(node.MP.String()))
	}

	if !apply {
		opts = append(opts, outline.AddApply(false))
	}
	svcResult, err := a.svc.Add(ctx, title, parentMPStr, opts...)
	if err != nil {
		return nil, err
	}

	result := &AddResult{
		Node: AddNodeInfo{
			MP:    svcResult.MP,
			SID:   svcResult.SID,
			Title: title,
		},
	}
	if apply {
		result.FilesCreated = []string{svcResult.Filename}
	} else {
		result.FilesPlanned = []string{svcResult.Filename}
	}
	return result, nil
}

// --- bootstrapAddAdapter ---

// bootstrapAddAdapter creates .linemark/ in CWD and wires a service before
// delegating to a real addAdapter. Used when no project exists yet.
type bootstrapAddAdapter struct {
	getwd       func() (string, error)
	wireService func(root string) (outlineServicer, error)
}

func (b *bootstrapAddAdapter) Add(ctx context.Context, title string, apply bool, p Placement) (*AddResult, error) {
	cwd, err := b.getwd()
	if err != nil {
		return nil, fmt.Errorf("getting working directory: %w", err)
	}

	dir := filepath.Join(cwd, ".linemark")
	if err := os.MkdirAll(dir, 0o755); err != nil {
		return nil, fmt.Errorf("creating .linemark directory: %w", err)
	}

	svc, err := b.wireService(cwd)
	if err != nil {
		return nil, err
	}

	real := &addAdapter{svc: svc}
	return real.Add(ctx, title, apply, p)
}

// --- checkAdapter ---

type checkAdapter struct {
	svc outlineServicer
}

func (a *checkAdapter) Check(ctx context.Context) (*CheckResult, error) {
	svcResult, err := a.svc.Check(ctx)
	if err != nil {
		return nil, err
	}

	findings := make([]CheckFinding, len(svcResult.Findings))
	for i, f := range svcResult.Findings {
		findings[i] = convertFinding(f)
	}
	return &CheckResult{Findings: findings}, nil
}

// --- repairAdapter ---

type repairAdapter struct {
	svc outlineServicer
}

func (a *repairAdapter) Repair(ctx context.Context) (*RepairResult, error) {
	svcResult, err := a.svc.Repair(ctx)
	if err != nil {
		return nil, err
	}

	repairs := make([]RepairAction, len(svcResult.Repairs))
	for i, r := range svcResult.Repairs {
		repairs[i] = RepairAction{
			Type:   FindingType(r.Type),
			Action: "repaired",
			Old:    r.Old,
			New:    r.New,
		}
	}

	unrepaired := make([]CheckFinding, len(svcResult.Unrepaired))
	for i, f := range svcResult.Unrepaired {
		unrepaired[i] = convertFinding(f)
	}

	return &RepairResult{
		Repairs:    repairs,
		Unrepaired: unrepaired,
	}, nil
}

// --- listAdapter ---

type listAdapter struct {
	svc outlineServicer
}

func (a *listAdapter) List(ctx context.Context) (*ListResult, error) {
	svcResult, err := a.svc.Load(ctx)
	if err != nil {
		return nil, err
	}
	return &ListResult{Outline: svcResult.Outline}, nil
}

// --- deleteAdapter ---

type deleteAdapter struct {
	svc outlineServicer
}

func (a *deleteAdapter) Delete(ctx context.Context, selector string, mode domain.DeleteMode, apply bool) (*DeleteResult, error) {
	sel, err := domain.ParseSelector(selector)
	if err != nil {
		return nil, err
	}

	svcResult, err := a.svc.Delete(ctx, sel, mode, apply)
	if err != nil {
		return nil, err
	}

	return &DeleteResult{
		FilesDeleted:  svcResult.FilesDeleted,
		FilesRenamed:  svcResult.FilesRenamed,
		SIDsPreserved: svcResult.SIDsPreserved,
	}, nil
}

// --- moveAdapter ---

type moveAdapter struct {
	svc outlineServicer
}

func (a *moveAdapter) Move(ctx context.Context, selector, to, before, after string, apply bool) (*MoveResult, error) {
	srcSel, err := domain.ParseSelector(selector)
	if err != nil {
		return nil, err
	}
	tgtSel, err := domain.ParseSelector(to)
	if err != nil {
		return nil, err
	}

	svcResult, err := a.svc.Move(ctx, srcSel, tgtSel, before, after, apply)
	if err != nil {
		return nil, err
	}

	return &MoveResult{Renames: convertRenames(svcResult.Renames)}, nil
}

// --- renameAdapter ---

type renameAdapter struct {
	svc outlineServicer
}

func (a *renameAdapter) Rename(ctx context.Context, selector, newTitle string, apply bool) (*RenameResult, error) {
	svcResult, err := a.svc.Rename(ctx, selector, newTitle, apply)
	if err != nil {
		return nil, err
	}

	return &RenameResult{
		Node: RenameNodeInfo{
			OldTitle: svcResult.OldTitle,
			NewTitle: svcResult.NewTitle,
		},
		Renames: convertRenames(svcResult.Renames),
	}, nil
}

// --- compactAdapter ---

type compactAdapter struct {
	svc outlineServicer
}

func (a *compactAdapter) Compact(ctx context.Context, selector string, apply bool) (*CompactResult, error) {
	svcResult, err := a.svc.Compact(ctx, selector, apply)
	if err != nil {
		return nil, err
	}

	result := &CompactResult{
		Renames:       convertRenames(svcResult.Renames),
		FilesAffected: len(svcResult.Renames),
	}
	if result.FilesAffected > 50 {
		w := fmt.Sprintf("compact affects %d files — review with --dry-run before applying", result.FilesAffected)
		result.Warning = &w
	}
	return result, nil
}

// --- typesAdapter ---

type typesAdapter struct {
	svc outlineServicer
}

func (a *typesAdapter) ListTypes(ctx context.Context, selector string) (*TypesListResult, error) {
	svcResult, err := a.svc.ListTypes(ctx, selector)
	if err != nil {
		return nil, err
	}
	return &TypesListResult{
		Node:  NodeInfo{MP: svcResult.NodeMP, SID: svcResult.NodeSID},
		Types: svcResult.Types,
	}, nil
}

func (a *typesAdapter) AddType(ctx context.Context, docType, selector string, apply bool) (*TypesModifyResult, error) {
	if !apply {
		if err := domain.ValidateDocType(docType); err != nil {
			return nil, err
		}
		sel, err := domain.ParseSelector(selector)
		if err != nil {
			return nil, err
		}
		node, err := a.svc.ResolveSelector(ctx, sel)
		if err != nil {
			return nil, err
		}
		filename := domain.GenerateFilename(node.MP.String(), node.SID, docType, "")
		return &TypesModifyResult{
			Node:     NodeInfo{MP: node.MP.String(), SID: node.SID},
			Filename: filename,
		}, nil
	}
	svcResult, err := a.svc.AddType(ctx, docType, selector)
	if err != nil {
		return nil, err
	}
	return &TypesModifyResult{
		Node:     NodeInfo{MP: svcResult.NodeMP, SID: svcResult.NodeSID},
		Filename: svcResult.Filename,
	}, nil
}

func (a *typesAdapter) RemoveType(ctx context.Context, docType, selector string, apply bool) (*TypesModifyResult, error) {
	if !apply {
		if err := domain.ValidateDocType(docType); err != nil {
			return nil, err
		}
		sel, err := domain.ParseSelector(selector)
		if err != nil {
			return nil, err
		}
		node, err := a.svc.ResolveSelector(ctx, sel)
		if err != nil {
			return nil, err
		}
		filename := domain.GenerateFilename(node.MP.String(), node.SID, docType, "")
		return &TypesModifyResult{
			Node:     NodeInfo{MP: node.MP.String(), SID: node.SID},
			Filename: filename,
		}, nil
	}
	svcResult, err := a.svc.RemoveType(ctx, docType, selector)
	if err != nil {
		return nil, err
	}
	return &TypesModifyResult{
		Node:     NodeInfo{MP: svcResult.NodeMP, SID: svcResult.NodeSID},
		Filename: svcResult.Filename,
	}, nil
}

// convertRenames converts a rename map to a slice of RenameEntry.
func convertRenames(m map[string]string) []RenameEntry {
	entries := make([]RenameEntry, 0, len(m))
	for old, newName := range m {
		entries = append(entries, RenameEntry{Old: old, New: newName})
	}
	return entries
}

// convertFinding converts a domain.Finding to a cmd.CheckFinding.
func convertFinding(f domain.Finding) CheckFinding {
	return CheckFinding{
		Type:     FindingType(f.Type),
		Severity: Severity(f.Severity),
		Message:  f.Message,
		Path:     f.Path,
	}
}
