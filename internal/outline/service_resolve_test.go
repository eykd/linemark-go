package outline

import (
	"context"
	"errors"
	"fmt"
	"testing"

	"github.com/eykd/linemark-go/internal/domain"
)

func TestOutlineService_ResolveSelector(t *testing.T) {
	mp1, _ := domain.NewMaterializedPath("001")
	mp2, _ := domain.NewMaterializedPath("002")
	testOutline := domain.Outline{
		Nodes: []domain.Node{
			{MP: mp1, SID: "ABCD1234EF", Title: "first"},
			{MP: mp2, SID: "WXYZ5678GH01", Title: "second"},
		},
	}

	tests := []struct {
		name     string
		selector string
		wantSID  string
		wantMP   string
		wantErr  error
	}{
		{
			name:     "resolves by bare MP",
			selector: "001",
			wantSID:  "ABCD1234EF",
			wantMP:   "001",
		},
		{
			name:     "resolves by explicit MP prefix",
			selector: "mp:002",
			wantSID:  "WXYZ5678GH01",
			wantMP:   "002",
		},
		{
			name:     "resolves by bare SID",
			selector: "ABCD1234EF",
			wantSID:  "ABCD1234EF",
			wantMP:   "001",
		},
		{
			name:     "resolves by explicit SID prefix",
			selector: "sid:WXYZ5678GH01",
			wantSID:  "WXYZ5678GH01",
			wantMP:   "002",
		},
		{
			name:     "not found by MP",
			selector: "999",
			wantErr:  ErrNodeNotFound,
		},
		{
			name:     "not found by explicit SID",
			selector: "sid:ZZZZZZZZZZ",
			wantErr:  ErrNodeNotFound,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			builder := &fakeOutlineBuilder{outline: testOutline}
			reader := &fakeDirectoryReader{files: []string{"001_ABCD1234EF_draft_first.md"}}
			svc := NewOutlineService(reader, nil, &mockLocker{}, nil)
			svc.builder = builder

			sel, err := domain.ParseSelector(tt.selector)
			if err != nil {
				t.Fatalf("failed to parse selector %q: %v", tt.selector, err)
			}

			node, err := svc.ResolveSelector(context.Background(), sel)

			if tt.wantErr != nil {
				if err == nil {
					t.Fatal("expected error, got nil")
				}
				if !errors.Is(err, tt.wantErr) {
					t.Errorf("error = %v, want %v", err, tt.wantErr)
				}
				return
			}
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if node.SID != tt.wantSID {
				t.Errorf("SID = %q, want %q", node.SID, tt.wantSID)
			}
			if node.MP.String() != tt.wantMP {
				t.Errorf("MP = %q, want %q", node.MP.String(), tt.wantMP)
			}
		})
	}
}

func TestOutlineService_ResolveSelector_AmbiguousSID(t *testing.T) {
	mp1, _ := domain.NewMaterializedPath("001")
	mp2, _ := domain.NewMaterializedPath("002")
	// Outline with duplicate SIDs at different MPs — ResolveSelector
	// should reject this as ambiguous rather than picking arbitrarily.
	testOutline := domain.Outline{
		Nodes: []domain.Node{
			{MP: mp1, SID: "ABCD1234EF", Title: "first"},
			{MP: mp2, SID: "ABCD1234EF", Title: "second"},
		},
	}

	builder := &fakeOutlineBuilder{outline: testOutline}
	reader := &fakeDirectoryReader{files: []string{"001_ABCD1234EF_draft_first.md"}}
	svc := NewOutlineService(reader, nil, &mockLocker{}, nil)
	svc.builder = builder

	sel, err := domain.ParseSelector("sid:ABCD1234EF")
	if err != nil {
		t.Fatalf("failed to parse selector: %v", err)
	}

	_, err = svc.ResolveSelector(context.Background(), sel)

	if err == nil {
		t.Fatal("expected error for ambiguous SID, got nil")
	}
	if !errors.Is(err, ErrAmbiguousSelector) {
		t.Errorf("error = %v, want %v", err, ErrAmbiguousSelector)
	}
}

func TestOutlineService_ResolveSelector_LoadErrors(t *testing.T) {
	tests := []struct {
		name     string
		readErr  error
		buildErr error
	}{
		{
			name:    "ReadDir error propagates",
			readErr: fmt.Errorf("permission denied"),
		},
		{
			name:     "BuildOutline error propagates",
			buildErr: fmt.Errorf("corrupt data"),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			reader := &fakeDirectoryReader{
				files: []string{"001_ABCD1234EF_draft_first.md"},
				err:   tt.readErr,
			}
			builder := &fakeOutlineBuilder{err: tt.buildErr}
			svc := NewOutlineService(reader, nil, &mockLocker{}, nil)
			svc.builder = builder

			sel, _ := domain.ParseSelector("001")

			_, err := svc.ResolveSelector(context.Background(), sel)

			if err == nil {
				t.Fatal("expected error, got nil")
			}
		})
	}
}
