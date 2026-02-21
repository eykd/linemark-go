package domain

import (
	"fmt"
	"slices"
	"strconv"
)

// Child returns a new MaterializedPath with the given segment appended.
func (mp MaterializedPath) Child(segment int) (MaterializedPath, error) {
	if segment < 1 || segment > 999 {
		return MaterializedPath{}, fmt.Errorf("%w: segment %d out of range 1-999", ErrInvalidPath, segment)
	}
	return MaterializedPath{segments: append(slices.Clone(mp.segments), fmt.Sprintf("%03d", segment))}, nil
}

// LastSegment returns the numeric value of the last path segment.
func (mp MaterializedPath) LastSegment() int {
	n, _ := strconv.Atoi(mp.segments[len(mp.segments)-1])
	return n
}

// IsAncestorOf returns true if mp is a strict ancestor of other.
func (mp MaterializedPath) IsAncestorOf(other MaterializedPath) bool {
	return len(mp.segments) < len(other.segments) &&
		slices.Equal(mp.segments, other.segments[:len(mp.segments)])
}

// Equal returns true if both paths have the same segments.
func (mp MaterializedPath) Equal(other MaterializedPath) bool {
	return slices.Equal(mp.segments, other.segments)
}
