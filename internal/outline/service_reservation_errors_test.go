package outline

import (
	"context"
	"errors"
	"testing"
)

func TestOutlineService_Check_ReservationStoreHasError_ReturnsError(t *testing.T) {
	// Given: A reservation store whose HasReservation fails
	reader := &fakeDirectoryReader{
		files: []string{
			"100_SID001AABB_draft_hello.md",
			"100_SID001AABB_notes.md",
		},
	}
	store := &fakeReservationStore{
		reservations: make(map[string]bool),
		hasErr:       errors.New("has check failed"),
	}
	svc := NewOutlineService(reader, nil, &mockLocker{}, nil,
		WithReservationStore(store))

	// When: Check is called
	_, err := svc.Check(context.Background())

	// Then: The error should propagate
	if err == nil {
		t.Fatal("expected error from HasReservation, got nil")
	}
}
