package outline

import (
	"context"
	"fmt"
	"testing"

	"github.com/eykd/linemark-go/internal/domain"
)

func TestOutlineService_Check_DetectsFindings(t *testing.T) {
	tests := []struct {
		name      string
		files     []string
		wantTypes []domain.FindingType
	}{
		{
			name:      "detects invalid filename",
			files:     []string{"not-a-valid-file.txt"},
			wantTypes: []domain.FindingType{domain.FindingInvalidFilename},
		},
		{
			name: "detects duplicate SIDs at different positions",
			files: []string{
				"001_ABCD1234EF_draft_hello.md",
				"001_ABCD1234EF_notes.md",
				"002_ABCD1234EF_draft_duplicate.md",
			},
			wantTypes: []domain.FindingType{domain.FindingDuplicateSID},
		},
		{
			name: "detects missing notes for node with only draft",
			files: []string{
				"001_ABCD1234EF_draft_hello.md",
			},
			wantTypes: []domain.FindingType{domain.FindingMissingDocType},
		},
		{
			name: "detects missing draft for node with only notes",
			files: []string{
				"001_ABCD1234EF_notes.md",
			},
			wantTypes: []domain.FindingType{domain.FindingMissingDocType},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			reader := &fakeDirectoryReader{files: tt.files}
			svc := NewOutlineService(reader, nil, &mockLocker{}, nil)

			result, err := svc.Check(context.Background())
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}

			if len(result.Findings) != len(tt.wantTypes) {
				t.Fatalf("findings count = %d, want %d; got: %v",
					len(result.Findings), len(tt.wantTypes), result.Findings)
			}

			for i, wantType := range tt.wantTypes {
				if result.Findings[i].Type != wantType {
					t.Errorf("finding[%d].Type = %q, want %q",
						i, result.Findings[i].Type, wantType)
				}
			}
		})
	}
}

func TestOutlineService_Check_CleanOutline_NoFindings(t *testing.T) {
	reader := &fakeDirectoryReader{
		files: []string{
			"001_ABCD1234EF_draft_hello.md",
			"001_ABCD1234EF_notes.md",
		},
	}
	svc := NewOutlineService(reader, nil, &mockLocker{}, nil)

	result, err := svc.Check(context.Background())
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if len(result.Findings) != 0 {
		t.Errorf("clean outline should have 0 findings, got %d: %v",
			len(result.Findings), result.Findings)
	}
}

func TestOutlineService_Check_PropagatesReaderError(t *testing.T) {
	reader := &fakeDirectoryReader{err: fmt.Errorf("permission denied")}
	svc := NewOutlineService(reader, nil, &mockLocker{}, nil)

	_, err := svc.Check(context.Background())

	if err == nil {
		t.Fatal("expected error from ReadDir")
	}
}

func TestOutlineService_Check_FindingSeverities(t *testing.T) {
	tests := []struct {
		name         string
		files        []string
		wantSeverity domain.FindingSeverity
	}{
		{
			name:         "missing notes is error severity",
			files:        []string{"001_ABCD1234EF_draft_hello.md"},
			wantSeverity: domain.SeverityError,
		},
		{
			name:         "invalid filename is warning severity",
			files:        []string{"not-valid.txt"},
			wantSeverity: domain.SeverityWarning,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			reader := &fakeDirectoryReader{files: tt.files}
			svc := NewOutlineService(reader, nil, &mockLocker{}, nil)

			result, err := svc.Check(context.Background())
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}

			if len(result.Findings) == 0 {
				t.Fatal("expected at least one finding")
			}
			if result.Findings[0].Severity != tt.wantSeverity {
				t.Errorf("severity = %q, want %q",
					result.Findings[0].Severity, tt.wantSeverity)
			}
		})
	}
}
