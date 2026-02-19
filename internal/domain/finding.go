package domain

// FindingSeverity indicates how severe a finding is.
type FindingSeverity string

const (
	// SeverityError indicates a finding that must be resolved.
	SeverityError FindingSeverity = "error"
	// SeverityWarning indicates a finding that should be reviewed.
	SeverityWarning FindingSeverity = "warning"
)

// FindingType identifies the kind of issue found.
type FindingType string

// Finding type constants identify the kind of issue found.
const (
	FindingInvalidFilename      FindingType = "invalid_filename"
	FindingDuplicateSID         FindingType = "duplicate_sid"
	FindingSlugDrift            FindingType = "slug_drift"
	FindingMissingDocType       FindingType = "missing_doc_type"
	FindingMalformedFrontmatter FindingType = "malformed_frontmatter"
	FindingOrphanedReservation  FindingType = "orphaned_reservation"
)

// Document type constants identify the standard document types.
const (
	DocTypeDraft = "draft"
	DocTypeNotes = "notes"
)

// Finding represents a validation issue discovered during a check operation.
type Finding struct {
	Type     FindingType
	Severity FindingSeverity
	Message  string
	Path     string
}
