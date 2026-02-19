// Package outline provides the application service for managing outline operations.
package outline

import (
	"context"
	"errors"
	"fmt"
	"strconv"
	"strings"

	"github.com/eykd/linemark-go/internal/domain"
	"github.com/eykd/linemark-go/internal/slug"
)

// ErrNodeNotFound is returned when a selector matches no node.
var ErrNodeNotFound = errors.New("node not found")

// ErrAmbiguousSelector is returned when a selector matches more than one node.
var ErrAmbiguousSelector = errors.New("ambiguous selector")

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
type ModifyResult struct{}

// ListResult holds the result of listing document types.
type ListResult struct{}

// CheckResult holds the result of checking the outline.
type CheckResult struct{}

// RepairResult holds the result of repairing the outline.
type RepairResult struct{}

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
	reader   DirectoryReader
	writer   FileWriter
	locker   Locker
	reserver SIDReserver
	builder  OutlineBuilder
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

	return &ModifyResult{}, nil
}

// RemoveType removes a document type from a node, acquiring an advisory lock first.
func (s *OutlineService) RemoveType(ctx context.Context, docType, selector string) (*ModifyResult, error) {
	if err := s.locker.TryLock(ctx); err != nil {
		return nil, err
	}
	defer s.locker.Unlock()

	return &ModifyResult{}, nil
}

// ListTypes lists document types for a node without acquiring an advisory lock.
func (s *OutlineService) ListTypes(ctx context.Context, selector string) (*ListResult, error) {
	return &ListResult{}, nil
}

// Check validates the outline without acquiring an advisory lock.
func (s *OutlineService) Check(ctx context.Context) (*CheckResult, error) {
	return &CheckResult{}, nil
}

// Repair repairs the outline, acquiring an advisory lock first.
func (s *OutlineService) Repair(ctx context.Context) (*RepairResult, error) {
	if err := s.locker.TryLock(ctx); err != nil {
		return nil, err
	}
	defer s.locker.Unlock()

	return &RepairResult{}, nil
}

// Load reads the project directory and builds an Outline without acquiring a lock.
func (s *OutlineService) Load(ctx context.Context) (*LoadResult, error) {
	files, err := s.reader.ReadDir(ctx)
	if err != nil {
		return nil, err
	}

	var parsed []domain.ParsedFile
	var findings []domain.Finding

	for _, f := range files {
		pf, parseErr := domain.ParseFilename(f)
		if parseErr != nil {
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
	filename := domain.GenerateFilename(mp, sid, "draft", slugStr)
	content := formatFrontmatter(title)

	if err := s.writer.WriteFile(ctx, filename, content); err != nil {
		return nil, err
	}

	notesFilename := domain.GenerateFilename(mp, sid, "notes", "")
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
