package cmd

import (
	"context"
	"fmt"
	"testing"

	"github.com/eykd/linemark-go/internal/domain"
	"github.com/eykd/linemark-go/internal/outline"
)

// sequentialResolveStub extends stubOutlineService to support multiple
// sequential ResolveSelector calls returning different nodes. This enables
// testing combined placement flags (e.g., ChildOf + Before).
type sequentialResolveStub struct {
	stubOutlineService
	resolveQueue []domain.Node
	resolveIdx   int
}

func (s *sequentialResolveStub) ResolveSelector(ctx context.Context, sel domain.Selector) (domain.Node, error) {
	if s.resolveIdx >= len(s.resolveQueue) {
		return domain.Node{}, fmt.Errorf("unexpected ResolveSelector call #%d", s.resolveIdx+1)
	}
	node := s.resolveQueue[s.resolveIdx]
	s.resolveIdx++
	return node, nil
}

// TestAddAdapter_ChildOfWithBefore verifies that --child-of and --before can be
// combined: the new node should be added under the parent with an AddBefore
// positioning option.
func TestAddAdapter_ChildOfWithBefore_PassesPositioningOpts(t *testing.T) {
	stub := &sequentialResolveStub{
		stubOutlineService: stubOutlineService{
			addResult: &outline.AddResult{
				SID:      "NEWNODE12345",
				MP:       "100-110",
				Filename: "100-110_NEWNODE12345_draft_before-child-two.md",
			},
		},
		resolveQueue: []domain.Node{
			{MP: mustParseMP("100"), SID: "PARENT123456"},      // resolved for ChildOf "100"
			{MP: mustParseMP("100-200"), SID: "CHILD22222222"}, // resolved for Before "100-200"
		},
	}
	adapter := &addAdapter{svc: stub}

	_, err := adapter.Add(context.Background(), "Before Child Two", true, Placement{
		ChildOf: "100",
		Before:  "100-200",
	})

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if stub.addParentMP != "100" {
		t.Errorf("parentMP = %q, want %q (child of 100)", stub.addParentMP, "100")
	}
	// Must pass AddBefore option — currently fails because Before is silently
	// ignored when ChildOf is set (switch statement only matches one case).
	if len(stub.addOpts) != 1 {
		t.Errorf("addOpts count = %d, want 1 (AddBefore option)", len(stub.addOpts))
	}
}

// TestAddAdapter_ChildOfWithAfter verifies that --child-of and --after can be
// combined: the new node should be added under the parent with an AddAfter
// positioning option.
func TestAddAdapter_ChildOfWithAfter_PassesPositioningOpts(t *testing.T) {
	stub := &sequentialResolveStub{
		stubOutlineService: stubOutlineService{
			addResult: &outline.AddResult{
				SID:      "NEWNODE12345",
				MP:       "100-150",
				Filename: "100-150_NEWNODE12345_draft_after-child-one.md",
			},
		},
		resolveQueue: []domain.Node{
			{MP: mustParseMP("100"), SID: "PARENT123456"},      // resolved for ChildOf "100"
			{MP: mustParseMP("100-100"), SID: "CHILD11111111"}, // resolved for After "100-100"
		},
	}
	adapter := &addAdapter{svc: stub}

	_, err := adapter.Add(context.Background(), "After Child One", true, Placement{
		ChildOf: "100",
		After:   "100-100",
	})

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if stub.addParentMP != "100" {
		t.Errorf("parentMP = %q, want %q (child of 100)", stub.addParentMP, "100")
	}
	// Must pass AddAfter option — currently fails because After is silently
	// ignored when ChildOf is set.
	if len(stub.addOpts) != 1 {
		t.Errorf("addOpts count = %d, want 1 (AddAfter option)", len(stub.addOpts))
	}
}

// TestAddAdapter_SiblingOfWithBefore verifies that --sibling-of and --before
// can be combined: the new node should be added under the shared parent with
// an AddBefore positioning option.
func TestAddAdapter_SiblingOfWithBefore_PassesPositioningOpts(t *testing.T) {
	stub := &sequentialResolveStub{
		stubOutlineService: stubOutlineService{
			addResult: &outline.AddResult{
				SID:      "NEWNODE12345",
				MP:       "100-110",
				Filename: "100-110_NEWNODE12345_draft_before-sibling.md",
			},
		},
		resolveQueue: []domain.Node{
			{MP: mustParseMP("100-200"), SID: "SIBLING22222"}, // resolved for SiblingOf "100-200"
			{MP: mustParseMP("100-200"), SID: "SIBLING22222"}, // resolved for Before "100-200"
		},
	}
	adapter := &addAdapter{svc: stub}

	_, err := adapter.Add(context.Background(), "Before Sibling", true, Placement{
		SiblingOf: "100-200",
		Before:    "100-200",
	})

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	// Parent is derived from SiblingOf node's parent MP
	if stub.addParentMP != "100" {
		t.Errorf("parentMP = %q, want %q (parent of 100-200)", stub.addParentMP, "100")
	}
	// Must pass AddBefore option — currently fails because Before is silently
	// ignored when SiblingOf is set.
	if len(stub.addOpts) != 1 {
		t.Errorf("addOpts count = %d, want 1 (AddBefore option)", len(stub.addOpts))
	}
}

// TestAddAdapter_SiblingOfWithAfter verifies that --sibling-of and --after can
// be combined: the new node should be added under the shared parent with an
// AddAfter positioning option.
func TestAddAdapter_SiblingOfWithAfter_PassesPositioningOpts(t *testing.T) {
	stub := &sequentialResolveStub{
		stubOutlineService: stubOutlineService{
			addResult: &outline.AddResult{
				SID:      "NEWNODE12345",
				MP:       "100-150",
				Filename: "100-150_NEWNODE12345_draft_after-sibling.md",
			},
		},
		resolveQueue: []domain.Node{
			{MP: mustParseMP("100-200"), SID: "SIBLING22222"}, // resolved for SiblingOf "100-200"
			{MP: mustParseMP("100-100"), SID: "SIBLING11111"}, // resolved for After "100-100"
		},
	}
	adapter := &addAdapter{svc: stub}

	_, err := adapter.Add(context.Background(), "After Sibling One", true, Placement{
		SiblingOf: "100-200",
		After:     "100-100",
	})

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	// Parent is derived from SiblingOf node's parent MP
	if stub.addParentMP != "100" {
		t.Errorf("parentMP = %q, want %q (parent of 100-200)", stub.addParentMP, "100")
	}
	// Must pass AddAfter option — currently fails because After is silently
	// ignored when SiblingOf is set.
	if len(stub.addOpts) != 1 {
		t.Errorf("addOpts count = %d, want 1 (AddAfter option)", len(stub.addOpts))
	}
}
