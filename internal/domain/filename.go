package domain

import (
	"errors"
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

	return ParsedFile{
		MP:        mp,
		SID:       matches[2],
		DocType:   matches[3],
		Slug:      matches[4],
		PathParts: pathParts,
		Depth:     len(pathParts),
	}, nil
}

// GenerateFilename creates a linemark filename from its components.
func GenerateFilename(mp, sid, docType, slug string) string {
	if slug == "" {
		return mp + "_" + sid + "_" + docType + ".md"
	}
	return mp + "_" + sid + "_" + docType + "_" + slug + ".md"
}
