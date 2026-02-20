package fs

import (
	"context"
	"os"
	"path/filepath"
	"testing"
)

func TestOSReservationStore_HasReservation(t *testing.T) {
	tests := []struct {
		name      string
		setupFunc func(t *testing.T, root string)
		sid       string
		want      bool
	}{
		{
			name:      "returns false when no marker file exists",
			setupFunc: func(_ *testing.T, _ string) {},
			sid:       "SID001AABB",
			want:      false,
		},
		{
			name: "returns true when marker file exists",
			setupFunc: func(t *testing.T, root string) {
				t.Helper()
				idsDir := filepath.Join(root, ".linemark", "ids")
				if err := os.MkdirAll(idsDir, 0o755); err != nil {
					t.Fatalf("creating ids dir: %v", err)
				}
				if err := os.WriteFile(filepath.Join(idsDir, "SID001AABB"), nil, 0o644); err != nil {
					t.Fatalf("creating marker: %v", err)
				}
			},
			sid:  "SID001AABB",
			want: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			root := t.TempDir()
			tt.setupFunc(t, root)

			store := &OSReservationStore{Root: root}
			got, err := store.HasReservation(context.Background(), tt.sid)
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if got != tt.want {
				t.Errorf("HasReservation(%q) = %v, want %v", tt.sid, got, tt.want)
			}
		})
	}
}

func TestOSReservationStore_CreateReservation(t *testing.T) {
	tests := []struct {
		name string
		sid  string
	}{
		{
			name: "creates marker file in .linemark/ids directory",
			sid:  "SID001AABB",
		},
		{
			name: "creates ids directory if it does not exist",
			sid:  "NEWSID12345",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			root := t.TempDir()
			// Create .linemark but not ids subdirectory
			if err := os.MkdirAll(filepath.Join(root, ".linemark"), 0o755); err != nil {
				t.Fatalf("creating .linemark dir: %v", err)
			}

			store := &OSReservationStore{Root: root}
			err := store.CreateReservation(context.Background(), tt.sid)
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}

			markerPath := filepath.Join(root, ".linemark", "ids", tt.sid)
			if _, statErr := os.Stat(markerPath); os.IsNotExist(statErr) {
				t.Errorf("expected marker file at %s, but it does not exist", markerPath)
			}
		})
	}
}
