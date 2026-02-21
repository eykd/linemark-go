package domain

import (
	"errors"
	"regexp"
	"slices"
	"strings"
)

// ErrInvalidPath is returned when a materialized path string is invalid.
var ErrInvalidPath = errors.New("invalid materialized path")

var segmentRegex = regexp.MustCompile(`^(\d{3}|[1-9]\d?)$`)

// MaterializedPath is a value object representing a node's position in the tree.
type MaterializedPath struct {
	segments []string
}

// NewMaterializedPath creates a MaterializedPath from a dash-separated string of 3-digit segments.
func NewMaterializedPath(s string) (MaterializedPath, error) {
	if s == "" {
		return MaterializedPath{}, ErrInvalidPath
	}

	segments := strings.Split(s, "-")
	for _, seg := range segments {
		if !segmentRegex.MatchString(seg) || seg == "000" {
			return MaterializedPath{}, ErrInvalidPath
		}
	}

	return MaterializedPath{segments: segments}, nil
}

// String returns the dash-separated string representation.
func (mp MaterializedPath) String() string {
	return strings.Join(mp.segments, "-")
}

// Depth returns the number of segments.
func (mp MaterializedPath) Depth() int {
	return len(mp.segments)
}

// Parent returns the parent path and true, or a zero value and false for root paths.
func (mp MaterializedPath) Parent() (MaterializedPath, bool) {
	if len(mp.segments) <= 1 {
		return MaterializedPath{}, false
	}
	return MaterializedPath{segments: slices.Clone(mp.segments[:len(mp.segments)-1])}, true
}

// Segments returns a copy of the path segments.
func (mp MaterializedPath) Segments() []string {
	return slices.Clone(mp.segments)
}

// Document represents a file attached to a node.
type Document struct {
	Type     string
	Filename string
	Content  string
}

// Node represents a position in the project outline.
type Node struct {
	MP        MaterializedPath
	SID       string
	Title     string
	Documents []Document
}
