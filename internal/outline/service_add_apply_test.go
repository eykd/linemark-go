package outline

import (
	"context"
	"testing"
)

// TestOutlineService_Add_AddApply_False_SkipsWrite verifies that when AddApply(false)
// is passed, no files are written to disk but the planned filename is returned.
func TestOutlineService_Add_AddApply_False_SkipsWrite(t *testing.T) {
	reader := &fakeDirectoryReader{files: []string{}}
	writer := &fakeFileWriter{written: make(map[string]string)}
	locker := &mockLocker{}
	reserver := &fakeSIDReserver{sid: "ABCD1234EF00"}
	svc := NewOutlineService(reader, writer, locker, reserver)

	result, err := svc.Add(context.Background(), "Hello World", "", AddApply(false))

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(writer.written) != 0 {
		t.Errorf("AddApply(false) should not write any files, but wrote: %v", writer.written)
	}
	if result.Filename == "" {
		t.Error("result.Filename should be populated even when apply=false")
	}
}
