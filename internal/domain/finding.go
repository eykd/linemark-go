package domain

// FindingSeverity indicates how severe a finding is.
type FindingSeverity string

const (
	// SeverityError indicates a finding that must be resolved.
	SeverityError FindingSeverity = "error"
	// SeverityWarning indicates a finding that should be reviewed.
	SeverityWarning FindingSeverity = "warning"
)

// Finding type constants identify the kind of issue found.
const (
	FindingInvalidFilename      = "invalid_filename"
	FindingDuplicateSID         = "duplicate_sid"
	FindingSlugDrift            = "slug_drift"
	FindingMissingDocType       = "missing_doc_type"
	FindingMalformedFrontmatter = "malformed_frontmatter"
	FindingOrphanedReservation  = "orphaned_reservation"
)

// Finding represents a validation issue discovered during a check operation.
type Finding struct {
	Type     string
	Severity FindingSeverity
	Message  string
	Path     string
}
