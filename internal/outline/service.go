// Package outline provides the application service for managing outline operations.
package outline

import (
	"context"
)

// Locker abstracts advisory lock acquisition for mutating commands.
type Locker interface {
	TryLock(ctx context.Context) error
	Unlock() error
}

// ModifyResult holds the result of a mutating outline operation.
type ModifyResult struct{}

// ListResult holds the result of listing document types.
type ListResult struct{}

// CheckResult holds the result of checking the outline.
type CheckResult struct{}

// RepairResult holds the result of repairing the outline.
type RepairResult struct{}

// OutlineService coordinates outline mutations with advisory locking.
type OutlineService struct {
	locker Locker
}

// NewOutlineService creates an OutlineService with the given Locker.
func NewOutlineService(locker Locker) *OutlineService {
	return &OutlineService{locker: locker}
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
