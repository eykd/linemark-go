package outline

import (
	"context"
	"errors"
	"testing"

	"github.com/eykd/linemark-go/internal/domain"
)

// fakeReservationStore is a test double for the ReservationStore interface.
type fakeReservationStore struct {
	reservations map[string]bool
	created      []string
	hasErr       error
	createErr    error
}

func (f *fakeReservationStore) HasReservation(_ context.Context, sid string) (bool, error) {
	if f.hasErr != nil {
		return false, f.hasErr
	}
	return f.reservations[sid], nil
}

func (f *fakeReservationStore) CreateReservation(_ context.Context, sid string) error {
	if f.createErr != nil {
		return f.createErr
	}
	f.created = append(f.created, sid)
	if f.reservations == nil {
		f.reservations = make(map[string]bool)
	}
	f.reservations[sid] = true
	return nil
}

// --- Add creates reservation marker ---

func TestOutlineService_Add_CreatesReservationMarker(t *testing.T) {
	// Given: An empty project with a reservation store
	reader := &fakeDirectoryReader{files: []string{}}
	writer := &fakeFileWriter{written: make(map[string]string)}
	locker := &mockLocker{}
	reserver := &fakeSIDReserver{sid: "ABCD1234EF00"}
	store := &fakeReservationStore{reservations: make(map[string]bool)}
	svc := NewOutlineService(reader, writer, locker, reserver,
		WithReservationStore(store))

	// When: A node is added
	_, err := svc.Add(context.Background(), "My Novel", "")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	// Then: A reservation marker should be created for the SID
	if len(store.created) != 1 {
		t.Fatalf("expected 1 reservation created, got %d: %v", len(store.created), store.created)
	}
	if store.created[0] != "ABCD1234EF00" {
		t.Errorf("reservation SID = %q, want %q", store.created[0], "ABCD1234EF00")
	}
}

func TestOutlineService_Add_ReservationStoreError_ReturnsError(t *testing.T) {
	// Given: A reservation store that fails
	reader := &fakeDirectoryReader{files: []string{}}
	writer := &fakeFileWriter{written: make(map[string]string)}
	locker := &mockLocker{}
	reserver := &fakeSIDReserver{sid: "ABCD1234EF00"}
	store := &fakeReservationStore{
		reservations: make(map[string]bool),
		createErr:    errFakeReservation,
	}
	svc := NewOutlineService(reader, writer, locker, reserver,
		WithReservationStore(store))

	// When: A node is added
	_, err := svc.Add(context.Background(), "My Novel", "")

	// Then: The reservation store error should propagate
	if err == nil {
		t.Fatal("expected error from reservation store, got nil")
	}
}

// --- Check detects missing reservation markers ---

func TestOutlineService_Check_DetectsMissingReservation(t *testing.T) {
	tests := []findingDetectionTest{
		{
			name: "detects missing reservation marker for node",
			files: []string{
				"100_SID001AABB_draft_chapter-one.md",
				"100_SID001AABB_notes.md",
			},
			wantFound:    true,
			wantPath:     "SID001AABB",
			wantSeverity: domain.SeverityWarning,
		},
		{
			name: "no finding when reservation marker exists",
			files: []string{
				"100_SID001AABB_draft_chapter-one.md",
				"100_SID001AABB_notes.md",
			},
			wantFound: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			reader := &fakeDirectoryReader{files: tt.files}
			store := &fakeReservationStore{reservations: make(map[string]bool)}

			// For the "no finding" case, pre-populate the reservation
			if !tt.wantFound {
				store.reservations["SID001AABB"] = true
			}

			svc := NewOutlineService(reader, nil, &mockLocker{}, nil,
				WithReservationStore(store))

			result, err := svc.Check(context.Background())
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}

			matched := findingsByType(result.Findings, domain.FindingMissingReservation)

			if tt.wantFound && len(matched) == 0 {
				t.Fatalf("expected %s finding, got none; all findings: %v",
					domain.FindingMissingReservation, result.Findings)
			}
			if !tt.wantFound && len(matched) > 0 {
				t.Fatalf("expected no %s findings, got: %v",
					domain.FindingMissingReservation, matched)
			}

			if tt.wantFound {
				f := matched[0]
				if f.Path != tt.wantPath {
					t.Errorf("finding.Path = %q, want %q", f.Path, tt.wantPath)
				}
				if f.Severity != tt.wantSeverity {
					t.Errorf("finding.Severity = %q, want %q", f.Severity, tt.wantSeverity)
				}
			}
		})
	}
}

func TestOutlineService_Check_NoReservationStore_SkipsReservationCheck(t *testing.T) {
	// Given: A service with no reservation store configured
	reader := &fakeDirectoryReader{
		files: []string{
			"100_SID001AABB_draft_hello.md",
			"100_SID001AABB_notes.md",
		},
	}
	svc := NewOutlineService(reader, nil, &mockLocker{}, nil)

	// When: Check is called
	result, err := svc.Check(context.Background())
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	// Then: No missing reservation findings should be produced
	matched := findingsByType(result.Findings, domain.FindingMissingReservation)
	if len(matched) != 0 {
		t.Errorf("expected no missing reservation findings when store is nil, got: %v", matched)
	}
}

// --- Repair creates missing reservation markers ---

func TestOutlineService_Repair_CreatesMissingReservationMarkers(t *testing.T) {
	// Given: A node whose SID has no reservation marker
	files := []string{
		"100_SID001AABB_draft_chapter-one.md",
		"100_SID001AABB_notes.md",
	}
	reader := &fakeDirectoryReader{files: files}
	writer := &fakeFileWriter{}
	store := &fakeReservationStore{reservations: make(map[string]bool)}
	svc := NewOutlineService(reader, writer, &mockLocker{}, nil,
		WithReservationStore(store))

	// When: Repair is called
	result, err := svc.Repair(context.Background())
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	// Then: The missing reservation marker should be created
	if len(store.created) != 1 {
		t.Fatalf("expected 1 reservation created, got %d: %v", len(store.created), store.created)
	}
	if store.created[0] != "SID001AABB" {
		t.Errorf("created reservation SID = %q, want %q", store.created[0], "SID001AABB")
	}

	// And: The repair should be reported
	foundRepair := false
	for _, repair := range result.Repairs {
		if repair.Type == domain.FindingMissingReservation {
			foundRepair = true
			break
		}
	}
	if !foundRepair {
		t.Errorf("expected repair action for missing reservation, got: %v", result.Repairs)
	}
}

func TestOutlineService_Repair_SkipsExistingReservations(t *testing.T) {
	// Given: A node whose SID already has a reservation marker
	files := []string{
		"100_SID001AABB_draft_chapter-one.md",
		"100_SID001AABB_notes.md",
	}
	reader := &fakeDirectoryReader{files: files}
	writer := &fakeFileWriter{}
	contentReader := &fakeContentReader{
		contents: map[string]string{
			"100_SID001AABB_draft_chapter-one.md": "---\ntitle: chapter-one\n---\n",
		},
	}
	store := &fakeReservationStore{
		reservations: map[string]bool{"SID001AABB": true},
	}
	svc := NewOutlineService(reader, writer, &mockLocker{}, nil,
		WithReservationStore(store))
	svc.contentReader = contentReader

	// When: Repair is called
	result, err := svc.Repair(context.Background())
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	// Then: No reservation markers should be created
	if len(store.created) != 0 {
		t.Errorf("expected 0 reservations created, got %d: %v", len(store.created), store.created)
	}

	// And: No reservation repair should be reported
	for _, repair := range result.Repairs {
		if repair.Type == domain.FindingMissingReservation {
			t.Errorf("unexpected reservation repair: %v", repair)
		}
	}
}

func TestOutlineService_Repair_ReservationStoreError_ReturnsError(t *testing.T) {
	// Given: A reservation store that fails on create
	files := []string{
		"100_SID001AABB_draft_chapter-one.md",
		"100_SID001AABB_notes.md",
	}
	reader := &fakeDirectoryReader{files: files}
	writer := &fakeFileWriter{}
	store := &fakeReservationStore{
		reservations: make(map[string]bool),
		createErr:    errFakeReservation,
	}
	svc := NewOutlineService(reader, writer, &mockLocker{}, nil,
		WithReservationStore(store))

	// When: Repair is called
	_, err := svc.Repair(context.Background())

	// Then: The error should propagate
	if err == nil {
		t.Fatal("expected error from reservation store, got nil")
	}
}

// --- Multiple nodes ---

func TestOutlineService_Repair_CreatesReservationsForMultipleNodes(t *testing.T) {
	// Given: Multiple nodes, each missing reservation markers
	files := []string{
		"100_SID001AABB_draft_first.md",
		"100_SID001AABB_notes.md",
		"200_SID002CCDD_draft_second.md",
		"200_SID002CCDD_notes.md",
	}
	reader := &fakeDirectoryReader{files: files}
	writer := &fakeFileWriter{}
	store := &fakeReservationStore{reservations: make(map[string]bool)}
	svc := NewOutlineService(reader, writer, &mockLocker{}, nil,
		WithReservationStore(store))

	// When: Repair is called
	result, err := svc.Repair(context.Background())
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	// Then: Both reservation markers should be created
	if len(store.created) != 2 {
		t.Fatalf("expected 2 reservations created, got %d: %v", len(store.created), store.created)
	}

	// And: Both repairs should be reported
	reservationRepairs := 0
	for _, repair := range result.Repairs {
		if repair.Type == domain.FindingMissingReservation {
			reservationRepairs++
		}
	}
	if reservationRepairs != 2 {
		t.Errorf("expected 2 reservation repairs, got %d; all repairs: %v", reservationRepairs, result.Repairs)
	}
}

var errFakeReservation = errors.New("reservation store failure")
