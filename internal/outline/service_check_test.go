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

func TestOutlineService_Check_PropagatesBuildOutlineError(t *testing.T) {
	reader := &fakeDirectoryReader{
		files: []string{"001_ABCD1234EF_draft_hello.md"},
	}
	builder := &fakeOutlineBuilder{
		err: fmt.Errorf("corrupt outline data"),
	}
	svc := NewOutlineService(reader, nil, &mockLocker{}, nil)
	svc.builder = builder

	_, err := svc.Check(context.Background())

	if err == nil {
		t.Fatal("expected error from BuildOutline")
	}
}

func TestOutlineService_Check_DetectsSlugDrift(t *testing.T) {
	tests := []struct {
		name         string
		files        []string
		contents     map[string]string
		wantDrift    bool
		wantPath     string
		wantSeverity domain.FindingSeverity
	}{
		{
			name: "detects slug drift when filename slug differs from title",
			files: []string{
				"100_SID001AABB_draft_wrong-slug.md",
				"100_SID001AABB_notes.md",
			},
			contents: map[string]string{
				"100_SID001AABB_draft_wrong-slug.md": "---\ntitle: Correct Title\n---\n",
			},
			wantDrift:    true,
			wantPath:     "100_SID001AABB_draft_wrong-slug.md",
			wantSeverity: domain.SeverityWarning,
		},
		{
			name: "no drift when slug matches title",
			files: []string{
				"100_SID001AABB_draft_correct-title.md",
				"100_SID001AABB_notes.md",
			},
			contents: map[string]string{
				"100_SID001AABB_draft_correct-title.md": "---\ntitle: Correct Title\n---\n",
			},
			wantDrift: false,
		},
		{
			name: "no drift when content reader is nil",
			files: []string{
				"100_SID001AABB_draft_wrong-slug.md",
				"100_SID001AABB_notes.md",
			},
			contents:  nil,
			wantDrift: false,
		},
		{
			name: "skips non-draft documents",
			files: []string{
				"100_SID001AABB_draft_correct-title.md",
				"100_SID001AABB_notes.md",
			},
			contents: map[string]string{
				"100_SID001AABB_draft_correct-title.md": "---\ntitle: Correct Title\n---\n",
			},
			wantDrift: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			reader := &fakeDirectoryReader{files: tt.files}
			svc := NewOutlineService(reader, nil, &mockLocker{}, nil)

			if tt.contents != nil {
				svc.contentReader = &fakeContentReader{contents: tt.contents}
			}

			result, err := svc.Check(context.Background())
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}

			var slugDriftFindings []domain.Finding
			for _, f := range result.Findings {
				if f.Type == domain.FindingSlugDrift {
					slugDriftFindings = append(slugDriftFindings, f)
				}
			}

			if tt.wantDrift && len(slugDriftFindings) == 0 {
				t.Fatalf("expected slug_drift finding, got none; all findings: %v", result.Findings)
			}
			if !tt.wantDrift && len(slugDriftFindings) > 0 {
				t.Fatalf("expected no slug_drift findings, got: %v", slugDriftFindings)
			}

			if tt.wantDrift {
				f := slugDriftFindings[0]
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

func TestOutlineService_Check_DetectsMalformedFrontmatter(t *testing.T) {
	tests := []struct {
		name          string
		files         []string
		contents      map[string]string
		wantMalformed bool
		wantPath      string
		wantSeverity  domain.FindingSeverity
	}{
		{
			name: "detects malformed YAML in draft frontmatter",
			files: []string{
				"100_SID001AABB_draft_chapter-one.md",
				"100_SID001AABB_notes.md",
			},
			contents: map[string]string{
				"100_SID001AABB_draft_chapter-one.md": "---\ntitle: [unclosed\n---\n",
			},
			wantMalformed: true,
			wantPath:      "100_SID001AABB_draft_chapter-one.md",
			wantSeverity:  domain.SeverityError,
		},
		{
			name: "no finding for valid frontmatter",
			files: []string{
				"100_SID001AABB_draft_chapter-one.md",
				"100_SID001AABB_notes.md",
			},
			contents: map[string]string{
				"100_SID001AABB_draft_chapter-one.md": "---\ntitle: Chapter One\n---\n",
			},
			wantMalformed: false,
		},
		{
			name: "no finding when content reader is nil",
			files: []string{
				"100_SID001AABB_draft_chapter-one.md",
				"100_SID001AABB_notes.md",
			},
			contents:      nil,
			wantMalformed: false,
		},
		{
			name: "skips non-draft documents",
			files: []string{
				"100_SID001AABB_draft_chapter-one.md",
				"100_SID001AABB_notes.md",
			},
			contents: map[string]string{
				"100_SID001AABB_draft_chapter-one.md": "---\ntitle: Chapter One\n---\n",
				"100_SID001AABB_notes.md":             "---\ntitle: [unclosed\n---\n",
			},
			wantMalformed: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			reader := &fakeDirectoryReader{files: tt.files}
			svc := NewOutlineService(reader, nil, &mockLocker{}, nil)

			if tt.contents != nil {
				svc.contentReader = &fakeContentReader{contents: tt.contents}
			}

			result, err := svc.Check(context.Background())
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}

			var malformedFindings []domain.Finding
			for _, f := range result.Findings {
				if f.Type == domain.FindingMalformedFrontmatter {
					malformedFindings = append(malformedFindings, f)
				}
			}

			if tt.wantMalformed && len(malformedFindings) == 0 {
				t.Fatalf("expected malformed_frontmatter finding, got none; all findings: %v", result.Findings)
			}
			if !tt.wantMalformed && len(malformedFindings) > 0 {
				t.Fatalf("expected no malformed_frontmatter findings, got: %v", malformedFindings)
			}

			if tt.wantMalformed {
				f := malformedFindings[0]
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
