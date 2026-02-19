package domain

import (
	"errors"
	"fmt"
	"regexp"
	"strings"
)

// ParsedFile represents a parsed linemark filename.
type ParsedFile struct {
	MP        string
	SID       string
	DocType   string
	Slug      string
	PathParts []string
	Depth     int
}

// ErrInvalidFilename is returned when a filename doesn't match the expected format.
var ErrInvalidFilename = errors.New("invalid filename")

// ErrInvalidDocType is returned when a doc type contains invalid characters.
var ErrInvalidDocType = errors.New("invalid doc type")

var docTypeRegex = regexp.MustCompile(`^[a-z]+$`)

// ValidateDocType checks that a doc type contains only lowercase ASCII letters.
func ValidateDocType(docType string) error {
	if !docTypeRegex.MatchString(docType) {
		return fmt.Errorf("%w: %q", ErrInvalidDocType, docType)
	}
	return nil
}

var filenameRegex = regexp.MustCompile(
	`^(\d{3}(?:-\d{3})*)_([a-zA-Z0-9]{8,12})_([a-z]+)(?:_(.+))?\.md$`,
)

// ParseFilename parses a linemark filename into its components.
func ParseFilename(filename string) (ParsedFile, error) {
	matches := filenameRegex.FindStringSubmatch(filename)
	if matches == nil {
		return ParsedFile{}, ErrInvalidFilename
	}

	mp := matches[1]
	pathParts := strings.Split(mp, "-")
	for _, part := range pathParts {
		if part == "000" {
			return ParsedFile{}, ErrInvalidFilename
		}
	}

	slug := matches[4]
	if strings.ContainsAny(slug, "/\\\x00") {
		return ParsedFile{}, ErrInvalidFilename
	}

	return ParsedFile{
		MP:        mp,
		SID:       matches[2],
		DocType:   matches[3],
		Slug:      slug,
		PathParts: pathParts,
		Depth:     len(pathParts),
	}, nil
}

// GenerateFilename creates a linemark filename from its components.
func GenerateFilename(mp, sid, docType, slug string) string {
	parts := []string{mp, sid, docType}
	if slug != "" {
		parts = append(parts, slug)
	}
	return strings.Join(parts, "_") + ".md"
}
