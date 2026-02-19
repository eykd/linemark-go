package domain

import (
	"testing"
)

func TestFinding_Fields(t *testing.T) {
	f := Finding{
		Type:     FindingDuplicateSID,
		Severity: SeverityError,
		Message:  "duplicate SID A3F7c9Qx7Lm2 found in two files",
		Path:     "001_A3F7c9Qx7Lm2_draft_my-novel.md",
	}

	if f.Type != FindingDuplicateSID {
		t.Errorf("Type = %q, want %q", f.Type, FindingDuplicateSID)
	}
	if f.Severity != SeverityError {
		t.Errorf("Severity = %q, want %q", f.Severity, SeverityError)
	}
	if f.Message != "duplicate SID A3F7c9Qx7Lm2 found in two files" {
		t.Errorf("Message = %q, want expected message", f.Message)
	}
	if f.Path != "001_A3F7c9Qx7Lm2_draft_my-novel.md" {
		t.Errorf("Path = %q, want expected path", f.Path)
	}
}

func TestFindingSeverity_Values(t *testing.T) {
	tests := []struct {
		name     string
		severity FindingSeverity
		want     string
	}{
		{"error severity", SeverityError, "error"},
		{"warning severity", SeverityWarning, "warning"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if string(tt.severity) != tt.want {
				t.Errorf("FindingSeverity = %q, want %q", string(tt.severity), tt.want)
			}
		})
	}
}

func TestFindingType_Constants(t *testing.T) {
	tests := []struct {
		name string
		ft   string
		want string
	}{
		{"invalid filename", FindingInvalidFilename, "invalid_filename"},
		{"duplicate SID", FindingDuplicateSID, "duplicate_sid"},
		{"slug drift", FindingSlugDrift, "slug_drift"},
		{"missing doc type", FindingMissingDocType, "missing_doc_type"},
		{"malformed frontmatter", FindingMalformedFrontmatter, "malformed_frontmatter"},
		{"orphaned reservation", FindingOrphanedReservation, "orphaned_reservation"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.ft != tt.want {
				t.Errorf("FindingType constant = %q, want %q", tt.ft, tt.want)
			}
		})
	}
}
