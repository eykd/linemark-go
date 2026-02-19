// Package outline provides the application service for managing outline operations.
package outline

import (
	"context"
	"errors"
	"fmt"
	"sort"
	"strconv"
	"strings"

	"github.com/eykd/linemark-go/internal/domain"
	"github.com/eykd/linemark-go/internal/frontmatter"
	"github.com/eykd/linemark-go/internal/slug"
)

// ErrNodeNotFound is returned when a selector matches no node.
var ErrNodeNotFound = errors.New("node not found")

// ErrAmbiguousSelector is returned when a selector matches more than one node.
var ErrAmbiguousSelector = errors.New("ambiguous selector")

// ErrNodeHasChildren is returned when deleting a node with children in default mode.
var ErrNodeHasChildren = errors.New("node has children; use --recursive or --promote")

// ErrInsufficientGaps is returned when promoting children but not enough sibling gaps exist.
var ErrInsufficientGaps = errors.New("insufficient gaps for promoted children")

// ErrCycleDetected is returned when a move would create a cycle.
var ErrCycleDetected = errors.New("cycle detected")

// DirectoryReader abstracts reading filenames from the project directory.
type DirectoryReader interface {
	ReadDir(ctx context.Context) ([]string, error)
}

// FileWriter abstracts writing files to the project directory.
type FileWriter interface {
	WriteFile(ctx context.Context, filename, content string) error
}

// Locker abstracts advisory lock acquisition for mutating commands.
type Locker interface {
	TryLock(ctx context.Context) error
	Unlock() error
}

// SIDReserver abstracts reserving a new stable ID.
type SIDReserver interface {
	Reserve(ctx context.Context) (string, error)
}

// FileDeleter abstracts deleting files from the project directory.
type FileDeleter interface {
	DeleteFile(ctx context.Context, filename string) error
}

// FileRenamer abstracts renaming files in the project directory.
type FileRenamer interface {
	RenameFile(ctx context.Context, oldName, newName string) error
}

// ContentReader abstracts reading file contents from the project directory.
type ContentReader interface {
	ReadFile(ctx context.Context, filename string) (string, error)
}

// OutlineBuilder abstracts building an Outline from parsed files.
type OutlineBuilder interface {
	BuildOutline(files []domain.ParsedFile) (domain.Outline, []domain.Finding, error)
}

// defaultOutlineBuilder delegates to domain.BuildOutline.
type defaultOutlineBuilder struct{}

func (d *defaultOutlineBuilder) BuildOutline(files []domain.ParsedFile) (domain.Outline, []domain.Finding, error) {
	return domain.BuildOutline(files)
}

// ModifyResult holds the result of a mutating outline operation.
type ModifyResult struct {
	Filename string
	NodeMP   string
	NodeSID  string
}

// ListResult holds the result of listing document types.
type ListResult struct {
	Types   []string
	NodeMP  string
	NodeSID string
}

// CheckResult holds the result of checking the outline.
type CheckResult struct {
	Findings []domain.Finding
}

// RepairAction describes a single repair that was applied.
type RepairAction struct {
	Type domain.FindingType
	Old  string
	New  string
}

// RepairResult holds the result of repairing the outline.
type RepairResult struct {
	Repairs    []RepairAction
	Unrepaired []domain.Finding
}

// CompactResult holds the result of a compact operation.
type CompactResult struct {
	Renames map[string]string
}

// RenameResult holds the result of a rename operation.
type RenameResult struct {
	OldTitle string
	NewTitle string
	Renames  map[string]string
}

// LoadResult holds the result of loading the outline from disk.
type LoadResult struct {
	Outline  domain.Outline
	Findings []domain.Finding
}

// AddResult holds the result of adding a new node to the outline.
type AddResult struct {
	SID      string
	MP       string
	Filename string
}

// OutlineService coordinates outline mutations with advisory locking.
type OutlineService struct {
	reader        DirectoryReader
	writer        FileWriter
	locker        Locker
	reserver      SIDReserver
	builder       OutlineBuilder
	deleter       FileDeleter
	renamer       FileRenamer
	contentReader ContentReader
}

// NewOutlineService creates an OutlineService with the given dependencies.
func NewOutlineService(reader DirectoryReader, writer FileWriter, locker Locker, reserver SIDReserver) *OutlineService {
	return &OutlineService{
		reader:   reader,
		writer:   writer,
		locker:   locker,
		reserver: reserver,
		builder:  &defaultOutlineBuilder{},
	}
}

// AddType adds a document type to a node, acquiring an advisory lock first.
func (s *OutlineService) AddType(ctx context.Context, docType, selector string) (*ModifyResult, error) {
	if err := s.locker.TryLock(ctx); err != nil {
		return nil, err
	}
	defer s.locker.Unlock()

	if s.reader == nil {
		return &ModifyResult{}, nil
	}

	return s.addTypeImpl(ctx, docType, selector)
}

// addTypeImpl performs the I/O operations for AddType.
func (s *OutlineService) addTypeImpl(ctx context.Context, docType, selector string) (*ModifyResult, error) {
	files, err := s.reader.ReadDir(ctx)
	if err != nil {
		return nil, err
	}

	parsed := parseValidFiles(files)

	nodeMP, nodeSID, err := findNodeMP(parsed, selector)
	if err != nil {
		return nil, err
	}

	filename := domain.GenerateFilename(nodeMP, nodeSID, docType, "")
	if err := s.writer.WriteFile(ctx, filename, ""); err != nil {
		return nil, err
	}

	return &ModifyResult{
		Filename: filename,
		NodeMP:   nodeMP,
		NodeSID:  nodeSID,
	}, nil
}

// RemoveType removes a document type from a node, acquiring an advisory lock first.
func (s *OutlineService) RemoveType(ctx context.Context, docType, selector string) (*ModifyResult, error) {
	if err := s.locker.TryLock(ctx); err != nil {
		return nil, err
	}
	defer s.locker.Unlock()

	if s.reader == nil {
		return &ModifyResult{}, nil
	}

	return s.removeTypeImpl(ctx, docType, selector)
}

// removeTypeImpl performs the I/O operations for RemoveType.
func (s *OutlineService) removeTypeImpl(ctx context.Context, docType, selector string) (*ModifyResult, error) {
	files, err := s.reader.ReadDir(ctx)
	if err != nil {
		return nil, err
	}

	parsed := parseValidFiles(files)

	nodeMP, nodeSID, err := findNodeMP(parsed, selector)
	if err != nil {
		return nil, err
	}

	filename := domain.GenerateFilename(nodeMP, nodeSID, docType, "")
	if err := s.deleter.DeleteFile(ctx, filename); err != nil {
		return nil, err
	}

	return &ModifyResult{
		Filename: filename,
		NodeMP:   nodeMP,
		NodeSID:  nodeSID,
	}, nil
}

// ListTypes lists document types for a node without acquiring an advisory lock.
func (s *OutlineService) ListTypes(ctx context.Context, selector string) (*ListResult, error) {
	if s.reader == nil {
		return &ListResult{}, nil
	}

	return s.listTypesImpl(ctx, selector)
}

// listTypesImpl performs the I/O operations for ListTypes.
func (s *OutlineService) listTypesImpl(ctx context.Context, selector string) (*ListResult, error) {
	files, err := s.reader.ReadDir(ctx)
	if err != nil {
		return nil, err
	}

	parsed := parseValidFiles(files)

	nodeMP, nodeSID, err := findNodeMP(parsed, selector)
	if err != nil {
		return nil, err
	}

	var types []string
	for _, pf := range parsed {
		if pf.MP == nodeMP {
			types = append(types, pf.DocType)
		}
	}
	sort.Strings(types)

	return &ListResult{
		Types:   types,
		NodeMP:  nodeMP,
		NodeSID: nodeSID,
	}, nil
}

// Check validates the outline without acquiring an advisory lock.
func (s *OutlineService) Check(ctx context.Context) (*CheckResult, error) {
	var files []string
	if s.reader != nil {
		var err error
		files, err = s.reader.ReadDir(ctx)
		if err != nil {
			return nil, err
		}
	}

	parsed, findings := parseFilesWithFindings(files)

	outline, buildFindings, err := s.builder.BuildOutline(parsed)
	if err != nil {
		return nil, err
	}

	findings = append(findings, buildFindings...)
	findings = append(findings, findMissingDocTypeFindings(outline.Nodes)...)

	return &CheckResult{Findings: findings}, nil
}

// Repair repairs the outline, acquiring an advisory lock first.
func (s *OutlineService) Repair(ctx context.Context) (*RepairResult, error) {
	if err := s.locker.TryLock(ctx); err != nil {
		return nil, err
	}
	defer s.locker.Unlock()

	if s.reader == nil {
		return &RepairResult{}, nil
	}

	return s.repairImpl(ctx)
}

// repairImpl performs the I/O operations for Repair.
func (s *OutlineService) repairImpl(ctx context.Context) (*RepairResult, error) {
	files, err := s.reader.ReadDir(ctx)
	if err != nil {
		return nil, err
	}

	parsed := parseValidFiles(files)

	outline, buildFindings, err := s.builder.BuildOutline(parsed)
	if err != nil {
		return nil, err
	}

	result := &RepairResult{}

	// Collect unrepaired findings (e.g., duplicate SIDs)
	result.Unrepaired = append(result.Unrepaired, buildFindings...)

	// Repair missing doc types
	for _, node := range outline.Nodes {
		hasNotes := false
		for _, doc := range node.Documents {
			if doc.Type == domain.DocTypeNotes {
				hasNotes = true
			}
		}
		if !hasNotes {
			filename := domain.GenerateFilename(node.MP.String(), node.SID, domain.DocTypeNotes, "")
			if err := s.writer.WriteFile(ctx, filename, ""); err != nil {
				return nil, err
			}
			result.Repairs = append(result.Repairs, RepairAction{
				Type: domain.FindingMissingDocType,
				New:  filename,
			})
		}
	}

	// Repair slug drift
	if s.contentReader != nil {
		for _, node := range outline.Nodes {
			for _, doc := range node.Documents {
				if doc.Type != domain.DocTypeDraft {
					continue
				}
				content, err := s.contentReader.ReadFile(ctx, doc.Filename)
				if err != nil {
					continue
				}
				title, err := frontmatter.GetTitle(content)
				if err != nil || title == "" {
					continue
				}
				expectedSlug := slug.Slug(title)
				pf, err := domain.ParseFilename(doc.Filename)
				if err != nil {
					continue
				}
				if pf.Slug != expectedSlug {
					newFilename := domain.GenerateFilename(pf.MP, pf.SID, pf.DocType, expectedSlug)
					if err := s.renamer.RenameFile(ctx, doc.Filename, newFilename); err != nil {
						return nil, err
					}
					result.Repairs = append(result.Repairs, RepairAction{
						Type: domain.FindingSlugDrift,
						Old:  doc.Filename,
						New:  newFilename,
					})
				}
			}
		}
	}

	return result, nil
}

// Load reads the project directory and builds an Outline without acquiring a lock.
func (s *OutlineService) Load(ctx context.Context) (*LoadResult, error) {
	files, err := s.reader.ReadDir(ctx)
	if err != nil {
		return nil, err
	}

	parsed, findings := parseFilesWithFindings(files)

	outline, buildFindings, err := s.builder.BuildOutline(parsed)
	if err != nil {
		return nil, err
	}

	findings = append(findings, buildFindings...)

	return &LoadResult{
		Outline:  outline,
		Findings: findings,
	}, nil
}

// ResolveSelector loads the outline and returns the node matching the given selector.
func (s *OutlineService) ResolveSelector(ctx context.Context, sel domain.Selector) (domain.Node, error) {
	result, err := s.Load(ctx)
	if err != nil {
		return domain.Node{}, err
	}

	if sel.Kind() == domain.SelectorMP {
		return findNodeByMP(result.Outline.Nodes, sel.Value())
	}
	return findNodeBySID(result.Outline.Nodes, sel.Value())
}

// findNodeByMP returns the first node whose MP matches the given value.
func findNodeByMP(nodes []domain.Node, mp string) (domain.Node, error) {
	for _, n := range nodes {
		if n.MP.String() == mp {
			return n, nil
		}
	}
	return domain.Node{}, ErrNodeNotFound
}

// findNodeBySID returns the unique node with the given SID, or an error if
// zero or multiple nodes match.
func findNodeBySID(nodes []domain.Node, sid string) (domain.Node, error) {
	var matches []domain.Node
	for _, n := range nodes {
		if n.SID == sid {
			matches = append(matches, n)
		}
	}
	switch len(matches) {
	case 0:
		return domain.Node{}, ErrNodeNotFound
	case 1:
		return matches[0], nil
	default:
		return domain.Node{}, ErrAmbiguousSelector
	}
}

// MoveResult holds the result of a move operation.
type MoveResult struct {
	Renames map[string]string
}

// DeleteResult holds the result of a delete operation at the service level.
type DeleteResult struct {
	FilesDeleted  []string
	FilesRenamed  map[string]string
	SIDsPreserved []string
}

// Delete removes a node from the outline, acquiring an advisory lock first.
func (s *OutlineService) Delete(ctx context.Context, sel domain.Selector, mode domain.DeleteMode, apply bool) (*DeleteResult, error) {
	if err := s.locker.TryLock(ctx); err != nil {
		return nil, err
	}
	defer s.locker.Unlock()

	files, err := s.reader.ReadDir(ctx)
	if err != nil {
		return nil, err
	}

	parsed := parseValidFiles(files)

	// Resolve the target node
	targetMP, targetSID, err := resolveTarget(parsed, sel)
	if err != nil {
		return nil, err
	}

	// Collect target and descendant files in a single pass.
	var targetFiles []string
	var descendantFiles []domain.ParsedFile
	for _, pf := range parsed {
		if pf.MP == targetMP {
			targetFiles = append(targetFiles, generateName(pf))
		} else if strings.HasPrefix(pf.MP, targetMP+"-") {
			descendantFiles = append(descendantFiles, pf)
		}
	}

	hasChildren := len(descendantFiles) > 0

	switch mode {
	case domain.DeleteModeDefault:
		if hasChildren {
			return nil, ErrNodeHasChildren
		}
		return s.deleteFiles(ctx, targetFiles, nil, []string{targetSID}, apply)
	case domain.DeleteModeRecursive:
		allFiles := append([]string{}, targetFiles...)
		sidSet := map[string]bool{targetSID: true}
		for _, pf := range descendantFiles {
			allFiles = append(allFiles, generateName(pf))
			sidSet[pf.SID] = true
		}
		var sids []string
		for sid := range sidSet {
			sids = append(sids, sid)
		}
		return s.deleteFiles(ctx, allFiles, nil, sids, apply)
	default: // DeleteModePromote
		return s.promoteChildren(ctx, parsed, targetMP, targetSID, targetFiles, descendantFiles, apply)
	}
}

// Move relocates a node and its descendants under a new parent, acquiring an advisory lock first.
func (s *OutlineService) Move(ctx context.Context, source, target domain.Selector, before, after string, apply bool) (*MoveResult, error) {
	if err := s.locker.TryLock(ctx); err != nil {
		return nil, err
	}
	defer s.locker.Unlock()

	files, err := s.reader.ReadDir(ctx)
	if err != nil {
		return nil, err
	}

	parsed := parseValidFiles(files)

	sourceMP, _, err := resolveTarget(parsed, source)
	if err != nil {
		return nil, err
	}
	targetMP, _, err := resolveTarget(parsed, target)
	if err != nil {
		return nil, err
	}

	if targetMP == sourceMP || strings.HasPrefix(targetMP, sourceMP+"-") {
		return nil, fmt.Errorf("cannot move %s to descendant %s: %w", sourceMP, targetMP, ErrCycleDetected)
	}

	nextNum, _ := domain.NextSiblingNumber(nil)
	newSourceMP := buildChildMP(targetMP, nextNum)

	renames := map[string]string{}
	for _, pf := range parsed {
		if pf.MP == sourceMP || strings.HasPrefix(pf.MP, sourceMP+"-") {
			oldName := generateName(pf)
			newMP := newSourceMP + pf.MP[len(sourceMP):]
			newName := domain.GenerateFilename(newMP, pf.SID, pf.DocType, pf.Slug)
			renames[oldName] = newName
		}
	}

	result := &MoveResult{Renames: renames}

	if !apply {
		return result, nil
	}

	if err := applyRenames(ctx, s.renamer, renames); err != nil {
		return nil, err
	}

	return result, nil
}

// Compact renumbers nodes at consistent spacing, acquiring an advisory lock first.
func (s *OutlineService) Compact(ctx context.Context, selector string, apply bool) (*CompactResult, error) {
	if err := s.locker.TryLock(ctx); err != nil {
		return nil, err
	}
	defer s.locker.Unlock()

	return s.compactImpl(ctx, selector, apply)
}

// compactImpl performs the I/O operations for Compact.
func (s *OutlineService) compactImpl(ctx context.Context, selector string, apply bool) (*CompactResult, error) {
	files, err := s.reader.ReadDir(ctx)
	if err != nil {
		return nil, err
	}

	parsed := parseValidFiles(files)
	renames := s.compactChildrenImpl(parsed, selector)

	result := &CompactResult{Renames: renames}

	if !apply {
		return result, nil
	}

	if err := applyRenames(ctx, s.renamer, renames); err != nil {
		return nil, err
	}

	return result, nil
}

// compactChildrenImpl computes renames to compact children of parentMP at consistent spacing.
func (s *OutlineService) compactChildrenImpl(parsed []domain.ParsedFile, parentMP string) map[string]string {
	renames := map[string]string{}

	// Find unique direct child MPs
	seen := map[string]bool{}
	var childMPs []string
	for _, pf := range parsed {
		if isDirectChild(pf, parentMP) && !seen[pf.MP] {
			seen[pf.MP] = true
			childMPs = append(childMPs, pf.MP)
		}
	}
	sort.Strings(childMPs)

	if len(childMPs) == 0 {
		return renames
	}

	// Compute new numbers
	newNums, err := domain.CompactNumbers(len(childMPs))
	if err != nil {
		return renames
	}

	// Build rename map for each child and its descendants
	for i, oldChildMP := range childMPs {
		newChildMP := buildChildMP(parentMP, newNums[i])

		for _, pf := range parsed {
			if pf.MP == oldChildMP || strings.HasPrefix(pf.MP, oldChildMP+"-") {
				oldName := generateName(pf)
				newMP := newChildMP + pf.MP[len(oldChildMP):]
				newName := domain.GenerateFilename(newMP, pf.SID, pf.DocType, pf.Slug)
				if oldName != newName {
					renames[oldName] = newName
				}
			}
		}

		// Recursively compact the children of this child
		childRenames := s.compactChildrenImpl(parsed, oldChildMP)
		for old, newName := range childRenames {
			// Apply the parent rename if the file was also renamed
			if parentNew, ok := renames[old]; ok {
				delete(renames, old)
				renames[parentNew] = newName
			} else {
				renames[old] = newName
			}
		}
	}

	return renames
}

// Rename changes the title and slug of a node, acquiring an advisory lock first.
func (s *OutlineService) Rename(ctx context.Context, selector, newTitle string, apply bool) (*RenameResult, error) {
	if err := s.locker.TryLock(ctx); err != nil {
		return nil, err
	}
	defer s.locker.Unlock()

	return s.renameImpl(ctx, selector, newTitle, apply)
}

// renameImpl performs the I/O operations for Rename.
func (s *OutlineService) renameImpl(ctx context.Context, selector, newTitle string, apply bool) (*RenameResult, error) {
	files, err := s.reader.ReadDir(ctx)
	if err != nil {
		return nil, err
	}

	parsed := parseValidFiles(files)

	nodeMP, nodeSID, err := findNodeMP(parsed, selector)
	if err != nil {
		return nil, err
	}

	// Find the draft file to get the old title
	var draftFile domain.ParsedFile
	var foundDraft bool
	for _, pf := range parsed {
		if pf.MP == nodeMP && pf.DocType == domain.DocTypeDraft {
			draftFile = pf
			foundDraft = true
			break
		}
	}

	var oldTitle string
	if foundDraft {
		oldFilename := generateName(draftFile)
		content, err := s.contentReader.ReadFile(ctx, oldFilename)
		if err == nil {
			oldTitle, _ = frontmatter.GetTitle(content)
		}
	}

	newSlug := slug.Slug(newTitle)
	renames := map[string]string{}

	for _, pf := range parsed {
		if pf.MP == nodeMP && pf.DocType == domain.DocTypeDraft {
			oldName := generateName(pf)
			newName := domain.GenerateFilename(pf.MP, pf.SID, pf.DocType, newSlug)
			if oldName != newName {
				renames[oldName] = newName
			}
		}
	}

	result := &RenameResult{
		OldTitle: oldTitle,
		NewTitle: newTitle,
		Renames:  renames,
	}

	if !apply {
		return result, nil
	}

	if err := applyRenames(ctx, s.renamer, renames); err != nil {
		return nil, err
	}

	// Update frontmatter title in the draft file and write to new filename
	if foundDraft {
		oldFilename := generateName(draftFile)
		content, err := s.contentReader.ReadFile(ctx, oldFilename)
		if err == nil {
			updatedContent, err := frontmatter.SetTitle(content, newTitle)
			if err == nil {
				newFilename := domain.GenerateFilename(nodeMP, nodeSID, domain.DocTypeDraft, newSlug)
				if writeErr := s.writer.WriteFile(ctx, newFilename, updatedContent); writeErr != nil {
					return nil, writeErr
				}
			}
		}
	}

	return result, nil
}

// findNodeMP finds the MP and SID of the first node matching the given selector string.
func findNodeMP(parsed []domain.ParsedFile, selector string) (string, string, error) {
	for _, pf := range parsed {
		if pf.MP == selector {
			return pf.MP, pf.SID, nil
		}
	}
	return "", "", ErrNodeNotFound
}

// findMissingDocTypeFindings checks each node for missing required document types.
func findMissingDocTypeFindings(nodes []domain.Node) []domain.Finding {
	var findings []domain.Finding
	for _, node := range nodes {
		hasDraft := false
		hasNotes := false
		for _, doc := range node.Documents {
			if doc.Type == domain.DocTypeDraft {
				hasDraft = true
			}
			if doc.Type == domain.DocTypeNotes {
				hasNotes = true
			}
		}
		if !hasDraft {
			findings = append(findings, domain.Finding{
				Type:     domain.FindingMissingDocType,
				Severity: domain.SeverityError,
				Message:  fmt.Sprintf("node %s missing draft", node.SID),
			})
		}
		if !hasNotes {
			findings = append(findings, domain.Finding{
				Type:     domain.FindingMissingDocType,
				Severity: domain.SeverityError,
				Message:  fmt.Sprintf("node %s missing notes", node.SID),
			})
		}
	}
	return findings
}

// parseFilesWithFindings parses filenames, returning parsed files and findings for invalid ones.
func parseFilesWithFindings(files []string) ([]domain.ParsedFile, []domain.Finding) {
	var parsed []domain.ParsedFile
	var findings []domain.Finding
	for _, f := range files {
		pf, err := domain.ParseFilename(f)
		if err != nil {
			findings = append(findings, domain.Finding{
				Type:     domain.FindingInvalidFilename,
				Severity: domain.SeverityWarning,
				Message:  fmt.Sprintf("invalid filename: %s", f),
				Path:     f,
			})
			continue
		}
		parsed = append(parsed, pf)
	}
	return parsed, findings
}

// parseValidFiles parses filenames, silently skipping invalid ones.
func parseValidFiles(files []string) []domain.ParsedFile {
	parsed, _ := parseFilesWithFindings(files)
	return parsed
}

// generateName reconstructs the filename from a ParsedFile's components.
func generateName(pf domain.ParsedFile) string {
	return domain.GenerateFilename(pf.MP, pf.SID, pf.DocType, pf.Slug)
}

// resolveTarget finds the MP and SID of the node matching the selector.
func resolveTarget(parsed []domain.ParsedFile, sel domain.Selector) (string, string, error) {
	isSID := sel.Kind() == domain.SelectorSID
	for _, pf := range parsed {
		if isSID && pf.SID == sel.Value() {
			return pf.MP, pf.SID, nil
		}
		if !isSID && pf.MP == sel.Value() {
			return pf.MP, pf.SID, nil
		}
	}
	return "", "", ErrNodeNotFound
}

// promoteChildren handles the promote delete mode.
func (s *OutlineService) promoteChildren(ctx context.Context, parsed []domain.ParsedFile, targetMP, targetSID string, targetFiles []string, descendantFiles []domain.ParsedFile, apply bool) (*DeleteResult, error) {
	// Find the parent MP of the target (empty string for root-level nodes).
	parentMP := ""
	if i := strings.LastIndex(targetMP, "-"); i >= 0 {
		parentMP = targetMP[:i]
	}

	// Find unique direct child nodes of the target
	seen := map[string]bool{}
	var childMPs []string
	targetDepth := strings.Count(targetMP, "-") + 1
	for _, pf := range descendantFiles {
		pfDepth := strings.Count(pf.MP, "-") + 1
		if pfDepth == targetDepth+1 && !seen[pf.MP] {
			seen[pf.MP] = true
			childMPs = append(childMPs, pf.MP)
		}
	}

	// Find occupied sibling numbers at the parent level (excluding the target)
	var siblingNums []int
	for _, pf := range parsed {
		if isDirectChild(pf, parentMP) && pf.MP != targetMP {
			n, _ := strconv.Atoi(pf.PathParts[len(pf.PathParts)-1])
			siblingNums = append(siblingNums, n)
		}
	}

	// Check for sufficient gaps to place all children
	need := len(childMPs)
	available := countAvailableGaps(siblingNums)
	if available < need {
		return nil, fmt.Errorf("%w: need %d slots, have %d", ErrInsufficientGaps, need, available)
	}

	// Assign new numbers for promoted children
	occupied := append([]int{}, siblingNums...)
	newNumbers := make([]int, need)
	for i := range childMPs {
		num, _ := domain.NextSiblingNumber(occupied)
		newNumbers[i] = num
		occupied = append(occupied, num)
	}

	// Build rename map
	renames := map[string]string{}
	for i, childMP := range childMPs {
		newMP := buildChildMP(parentMP, newNumbers[i])
		for _, pf := range descendantFiles {
			if pf.MP == childMP {
				oldName := generateName(pf)
				newName := domain.GenerateFilename(newMP, pf.SID, pf.DocType, pf.Slug)
				renames[oldName] = newName
			}
		}
	}

	return s.deleteFiles(ctx, targetFiles, renames, []string{targetSID}, apply)
}

// countAvailableGaps returns the number of available sibling positions
// in the range [1, max(occupied)], excluding the occupied positions themselves.
func countAvailableGaps(occupied []int) int {
	unique := map[int]bool{}
	maxNum := 0
	for _, n := range occupied {
		unique[n] = true
		if n > maxNum {
			maxNum = n
		}
	}
	return maxNum - len(unique)
}

// deleteFiles performs the actual file deletions and renames, or just plans them.
func (s *OutlineService) deleteFiles(ctx context.Context, toDelete []string, toRename map[string]string, sids []string, apply bool) (*DeleteResult, error) {
	result := &DeleteResult{
		FilesDeleted:  toDelete,
		FilesRenamed:  toRename,
		SIDsPreserved: sids,
	}

	if !apply {
		return result, nil
	}

	// Perform renames first (before deletes)
	if s.renamer != nil {
		if err := applyRenames(ctx, s.renamer, toRename); err != nil {
			return nil, err
		}
	}

	// Perform deletes, tracking completed deletions for diagnostics.
	var deleted []string
	for _, f := range toDelete {
		if err := s.deleter.DeleteFile(ctx, f); err != nil {
			return nil, fmt.Errorf("delete %s: %w (already deleted: %s)", f, err, strings.Join(deleted, ", "))
		}
		deleted = append(deleted, f)
	}

	return result, nil
}

// applyRenames executes a set of renames with rollback on failure.
func applyRenames(ctx context.Context, renamer FileRenamer, renames map[string]string) error {
	var completed [][2]string
	for oldName, newName := range renames {
		if err := renamer.RenameFile(ctx, oldName, newName); err != nil {
			rollbackRenames(ctx, renamer, completed)
			return fmt.Errorf("rename %s -> %s: %w", oldName, newName, err)
		}
		completed = append(completed, [2]string{oldName, newName})
	}
	return nil
}

// rollbackRenames reverses already-completed renames on a best-effort basis.
func rollbackRenames(ctx context.Context, renamer FileRenamer, completed [][2]string) {
	for i := len(completed) - 1; i >= 0; i-- {
		_ = renamer.RenameFile(ctx, completed[i][1], completed[i][0])
	}
}

// Add creates a new node in the outline, acquiring an advisory lock first.
func (s *OutlineService) Add(ctx context.Context, title, parentMP string) (*AddResult, error) {
	if err := s.locker.TryLock(ctx); err != nil {
		return nil, err
	}
	defer s.locker.Unlock()

	files, err := s.reader.ReadDir(ctx)
	if err != nil {
		return nil, err
	}

	sid, err := s.reserver.Reserve(ctx)
	if err != nil {
		return nil, err
	}

	occupied := collectChildNumbers(files, parentMP)

	nextNum, err := domain.NextSiblingNumber(occupied)
	if err != nil {
		return nil, err
	}

	mp := buildChildMP(parentMP, nextNum)

	slugStr := slug.Slug(title)
	filename := domain.GenerateFilename(mp, sid, domain.DocTypeDraft, slugStr)
	content := formatFrontmatter(title)

	if err := s.writer.WriteFile(ctx, filename, content); err != nil {
		return nil, err
	}

	notesFilename := domain.GenerateFilename(mp, sid, domain.DocTypeNotes, "")
	if err := s.writer.WriteFile(ctx, notesFilename, ""); err != nil {
		return nil, err
	}

	return &AddResult{
		SID:      sid,
		MP:       mp,
		Filename: filename,
	}, nil
}

// formatFrontmatter creates YAML frontmatter with a title field.
func formatFrontmatter(title string) string {
	value := title
	if strings.Contains(title, ":") {
		value = `"` + title + `"`
	}
	return fmt.Sprintf("---\ntitle: %s\n---\n", value)
}

// buildChildMP constructs an MP path by appending a numbered segment under parentMP.
func buildChildMP(parentMP string, num int) string {
	segment := fmt.Sprintf("%03d", num)
	if parentMP == "" {
		return segment
	}
	return parentMP + "-" + segment
}

// isDirectChild reports whether pf is a direct child of parentMP.
func isDirectChild(pf domain.ParsedFile, parentMP string) bool {
	if parentMP == "" {
		return pf.Depth == 1
	}
	parentDepth := strings.Count(parentMP, "-") + 1
	return strings.HasPrefix(pf.MP, parentMP+"-") && pf.Depth == parentDepth+1
}

// collectChildNumbers returns the occupied sibling numbers under parentMP.
func collectChildNumbers(files []string, parentMP string) []int {
	var occupied []int
	for _, f := range files {
		pf, parseErr := domain.ParseFilename(f)
		if parseErr != nil {
			continue
		}
		if isDirectChild(pf, parentMP) {
			// PathParts are guaranteed numeric by ParseFilename regex (\d{3}).
			n, _ := strconv.Atoi(pf.PathParts[len(pf.PathParts)-1])
			occupied = append(occupied, n)
		}
	}
	return occupied
}
