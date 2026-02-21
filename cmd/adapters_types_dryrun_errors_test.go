package cmd

// Coverage tests for typesAdapter.AddType and typesAdapter.RemoveType dry-run error paths.
// The happy-path dry-run tests live in adapters_dryrun_test.go; these tests cover
// the three early-return error branches in each method's apply=false block.

import (
	"context"
	"errors"
	"testing"
)

// --- typesAdapter.AddType dry-run error paths ---

func TestTypesAdapter_AddType_DryRun_InvalidDocType(t *testing.T) {
	adapter := &typesAdapter{svc: &stubOutlineService{}}

	_, err := adapter.AddType(context.Background(), "INVALID", "100", false)

	if err == nil {
		t.Fatal("expected error for invalid doc type, got nil")
	}
}

func TestTypesAdapter_AddType_DryRun_InvalidSelector(t *testing.T) {
	adapter := &typesAdapter{svc: &stubOutlineService{}}

	_, err := adapter.AddType(context.Background(), "characters", "", false)

	if err == nil {
		t.Fatal("expected error for empty selector, got nil")
	}
}

func TestTypesAdapter_AddType_DryRun_ResolveError(t *testing.T) {
	stub := &stubOutlineService{resolveErr: errors.New("node not found")}
	adapter := &typesAdapter{svc: stub}

	_, err := adapter.AddType(context.Background(), "characters", "100", false)

	if err == nil {
		t.Fatal("expected error from ResolveSelector, got nil")
	}
}

// --- typesAdapter.RemoveType dry-run error paths ---

func TestTypesAdapter_RemoveType_DryRun_InvalidDocType(t *testing.T) {
	adapter := &typesAdapter{svc: &stubOutlineService{}}

	_, err := adapter.RemoveType(context.Background(), "INVALID", "100", false)

	if err == nil {
		t.Fatal("expected error for invalid doc type, got nil")
	}
}

func TestTypesAdapter_RemoveType_DryRun_InvalidSelector(t *testing.T) {
	adapter := &typesAdapter{svc: &stubOutlineService{}}

	_, err := adapter.RemoveType(context.Background(), "characters", "", false)

	if err == nil {
		t.Fatal("expected error for empty selector, got nil")
	}
}

func TestTypesAdapter_RemoveType_DryRun_ResolveError(t *testing.T) {
	stub := &stubOutlineService{resolveErr: errors.New("node not found")}
	adapter := &typesAdapter{svc: stub}

	_, err := adapter.RemoveType(context.Background(), "characters", "100", false)

	if err == nil {
		t.Fatal("expected error from ResolveSelector, got nil")
	}
}
