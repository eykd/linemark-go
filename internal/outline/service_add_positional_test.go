package outline

import (
	"context"
	"testing"
)

func TestOutlineService_Add_WithAddBefore(t *testing.T) {
	reader := &fakeDirectoryReader{
		files: []string{
			"100_AAAA11111111_draft_first.md",
			"200_BBBB22222222_draft_second.md",
		},
	}
	writer := &fakeFileWriter{written: make(map[string]string)}
	locker := &mockLocker{}
	reserver := &fakeSIDReserver{sid: "NEWNODE123456"}
	svc := NewOutlineService(reader, writer, locker, reserver)

	result, err := svc.Add(context.Background(), "Before Second", "", AddBefore("200"))

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	// Should insert before 200, so MP should be between 100 and 200
	if result.MP != "110" && result.MP != "150" {
		// FindGap(100, 200) tries 100s tier: ((100/100)+1)*100 = 200, not < 200
		// Then 10s tier: ((100/10)+1)*10 = 110, 110 > 100 && 110 < 200 => 110
		if result.MP != "110" {
			t.Errorf("MP = %q, want %q (before 200, after 100)", result.MP, "110")
		}
	}
}

func TestOutlineService_Add_WithAddAfter(t *testing.T) {
	reader := &fakeDirectoryReader{
		files: []string{
			"100_AAAA11111111_draft_first.md",
			"200_BBBB22222222_draft_second.md",
		},
	}
	writer := &fakeFileWriter{written: make(map[string]string)}
	locker := &mockLocker{}
	reserver := &fakeSIDReserver{sid: "NEWNODE123456"}
	svc := NewOutlineService(reader, writer, locker, reserver)

	result, err := svc.Add(context.Background(), "After First", "", AddAfter("100"))

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	// After 100, next is 200. SiblingNumberAfter tries look-ahead.
	// Direct gap: FindGap(100, 200) => 110
	// Skip gap (next=200): FindGap(200, 1000) => 300, tier 100 > tier 10
	// So result should be 300
	if result.MP != "300" {
		t.Errorf("MP = %q, want %q (after 100 with look-ahead)", result.MP, "300")
	}
}

func TestOutlineService_Add_WithAddBeforeAtChildLevel(t *testing.T) {
	reader := &fakeDirectoryReader{
		files: []string{
			"100_AAAA11111111_draft_parent.md",
			"100-100_CCCC33333333_draft_child-one.md",
			"100-200_DDDD44444444_draft_child-two.md",
		},
	}
	writer := &fakeFileWriter{written: make(map[string]string)}
	locker := &mockLocker{}
	reserver := &fakeSIDReserver{sid: "NEWCHILD12345"}
	svc := NewOutlineService(reader, writer, locker, reserver)

	result, err := svc.Add(context.Background(), "Before Child Two", "100", AddBefore("100-200"))

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	// Should be between 100-100 and 100-200 => 100-110
	if result.MP != "100-110" {
		t.Errorf("MP = %q, want %q", result.MP, "100-110")
	}
}

func TestOutlineService_Add_WithoutOptions_AppendsAtEnd(t *testing.T) {
	reader := &fakeDirectoryReader{
		files: []string{
			"100_AAAA11111111_draft_first.md",
		},
	}
	writer := &fakeFileWriter{written: make(map[string]string)}
	locker := &mockLocker{}
	reserver := &fakeSIDReserver{sid: "NEWNODE123456"}
	svc := NewOutlineService(reader, writer, locker, reserver)

	result, err := svc.Add(context.Background(), "Second Node", "")

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result.MP != "200" {
		t.Errorf("MP = %q, want %q (default append)", result.MP, "200")
	}
}
