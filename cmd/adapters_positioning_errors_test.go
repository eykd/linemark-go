package cmd

import (
	"context"
	"errors"
	"fmt"
	"testing"

	"github.com/eykd/linemark-go/internal/domain"
)

// sequentialResolveWithErrorStub supports sequential ResolveSelector calls
// where individual calls can succeed or return an error.
type sequentialResolveWithErrorStub struct {
	stubOutlineService
	resolveQueue []resolveOutcome
	resolveIdx   int
}

type resolveOutcome struct {
	node domain.Node
	err  error
}

func (s *sequentialResolveWithErrorStub) ResolveSelector(ctx context.Context, sel domain.Selector) (domain.Node, error) {
	if s.resolveIdx >= len(s.resolveQueue) {
		return domain.Node{}, fmt.Errorf("unexpected ResolveSelector call #%d", s.resolveIdx+1)
	}
	out := s.resolveQueue[s.resolveIdx]
	s.resolveIdx++
	return out.node, out.err
}

// TestAddAdapter_ChildOfWithBefore_BeforeResolveError verifies that an error
// resolving the --before reference is propagated when combined with --child-of.
func TestAddAdapter_ChildOfWithBefore_BeforeResolveError(t *testing.T) {
	stub := &sequentialResolveWithErrorStub{
		resolveQueue: []resolveOutcome{
			{node: domain.Node{MP: mustParseMP("100"), SID: "PARENT123456"}}, // ChildOf succeeds
			{err: errors.New("before node not found")},                       // Before fails
		},
	}
	adapter := &addAdapter{svc: stub}

	_, err := adapter.Add(context.Background(), "Title", true, Placement{
		ChildOf: "100",
		Before:  "100-200",
	})

	if err == nil {
		t.Fatal("expected error when Before resolution fails")
	}
}

// TestAddAdapter_SiblingOfWithAfter_AfterResolveError verifies that an error
// resolving the --after reference is propagated when combined with --sibling-of.
func TestAddAdapter_SiblingOfWithAfter_AfterResolveError(t *testing.T) {
	stub := &sequentialResolveWithErrorStub{
		resolveQueue: []resolveOutcome{
			{node: domain.Node{MP: mustParseMP("100-200"), SID: "SIBLING12345"}}, // SiblingOf succeeds
			{err: errors.New("after node not found")},                            // After fails
		},
	}
	adapter := &addAdapter{svc: stub}

	_, err := adapter.Add(context.Background(), "Title", true, Placement{
		SiblingOf: "100-200",
		After:     "100-100",
	})

	if err == nil {
		t.Fatal("expected error when After resolution fails")
	}
}
