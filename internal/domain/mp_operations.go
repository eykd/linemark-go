package domain

import (
	"fmt"
	"strconv"
)

// Child returns a new MaterializedPath with the given segment appended.
func (mp MaterializedPath) Child(segment int) (MaterializedPath, error) {
	if segment < 1 || segment > 999 {
		return MaterializedPath{}, fmt.Errorf("%w: segment %d out of range 1-999", ErrInvalidPath, segment)
	}
	newSegments := make([]string, len(mp.segments)+1)
	copy(newSegments, mp.segments)
	newSegments[len(mp.segments)] = fmt.Sprintf("%03d", segment)
	return MaterializedPath{segments: newSegments}, nil
}

// LastSegment returns the numeric value of the last path segment.
func (mp MaterializedPath) LastSegment() int {
	n, _ := strconv.Atoi(mp.segments[len(mp.segments)-1])
	return n
}

// IsAncestorOf returns true if mp is a strict ancestor of other.
func (mp MaterializedPath) IsAncestorOf(other MaterializedPath) bool {
	if len(mp.segments) >= len(other.segments) {
		return false
	}
	for i, seg := range mp.segments {
		if other.segments[i] != seg {
			return false
		}
	}
	return true
}

// Equal returns true if both paths have the same segments.
func (mp MaterializedPath) Equal(other MaterializedPath) bool {
	if len(mp.segments) != len(other.segments) {
		return false
	}
	for i, seg := range mp.segments {
		if other.segments[i] != seg {
			return false
		}
	}
	return true
}
