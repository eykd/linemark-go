package domain

import (
	"errors"
	"regexp"
	"strings"
)

// ErrInvalidPath is returned when a materialized path string is invalid.
var ErrInvalidPath = errors.New("invalid materialized path")

var segmentRegex = regexp.MustCompile(`^\d{3}$`)

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
	parent := make([]string, len(mp.segments)-1)
	copy(parent, mp.segments)
	return MaterializedPath{segments: parent}, true
}

// Segments returns a copy of the path segments.
func (mp MaterializedPath) Segments() []string {
	result := make([]string, len(mp.segments))
	copy(result, mp.segments)
	return result
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
