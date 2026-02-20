package outline

import (
	"context"
	"errors"
	"testing"
)

// TestOutlineService_Add_RejectsEmptyTitle verifies that Add returns ErrEmptyTitle
// when given an empty or whitespace-only title, and that no files are written.
func TestOutlineService_Add_RejectsEmptyTitle(t *testing.T) {
	tests := []struct {
		name  string
		title string
	}{
		{"empty string", ""},
		{"whitespace only", "   "},
		{"tab only", "\t"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			reader := &fakeDirectoryReader{files: []string{}}
			writer := &fakeFileWriter{written: make(map[string]string)}
			locker := &mockLocker{}
			reserver := &fakeSIDReserver{sid: "ABCD1234EF00"}
			svc := NewOutlineService(reader, writer, locker, reserver)

			_, err := svc.Add(context.Background(), tt.title, "")

			if err == nil {
				t.Fatal("expected error for empty/whitespace title, got nil")
			}
			if !errors.Is(err, ErrEmptyTitle) {
				t.Errorf("error = %v, want ErrEmptyTitle", err)
			}
			if len(writer.written) != 0 {
				t.Errorf("expected no files written, got %d: %v", len(writer.written), writerKeys(writer))
			}
		})
	}
}
