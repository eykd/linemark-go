package domain

import (
	"errors"
	"sort"
)

const (
	maxSibling     = 999
	initialSpacing = 100
)

// ErrMaxSiblingsReached is returned when all 999 sibling slots are occupied.
var ErrMaxSiblingsReached = errors.New("maximum siblings reached (999)")

// ErrNoSlotAvailable is returned when no gap exists after the last sibling.
var ErrNoSlotAvailable = errors.New("no slot available; compact renumbering recommended")

// FindGap finds the best-tiered number strictly between low and high.
// It tries 100s, then 10s, then 1s multiples.
func FindGap(low, high int) (int, bool) {
	if high <= low+1 {
		return 0, false
	}

	for _, tier := range []int{100, 10} {
		candidate := ((low/tier) + 1) * tier
		if candidate > low && candidate < high {
			return candidate, true
		}
	}

	// 1s tier: guaranteed to succeed when high > low+1
	return low + 1, true
}

func tierOf(n int) int {
	if n%100 == 0 {
		return 100
	}
	if n%10 == 0 {
		return 10
	}
	return 1
}

func sortedCopy(occupied []int) []int {
	sorted := make([]int, len(occupied))
	copy(sorted, occupied)
	sort.Ints(sorted)
	return sorted
}

// NextSiblingNumber returns the next sibling number to append after existing siblings.
func NextSiblingNumber(occupied []int) (int, error) {
	if len(occupied) == 0 {
		return initialSpacing, nil
	}

	sorted := sortedCopy(occupied)

	if len(sorted) >= maxSibling {
		return 0, ErrMaxSiblingsReached
	}

	last := sorted[len(sorted)-1]
	result, ok := FindGap(last, maxSibling+1)
	if !ok {
		return 0, ErrNoSlotAvailable
	}
	return result, nil
}

// SiblingNumberBefore returns a sibling number to insert before the target.
func SiblingNumberBefore(occupied []int, target int) (int, error) {
	sorted := sortedCopy(occupied)

	if len(sorted) >= maxSibling {
		return 0, ErrMaxSiblingsReached
	}

	idx := sort.SearchInts(sorted, target)

	var predecessor int
	if idx > 0 {
		predecessor = sorted[idx-1]
	}

	result, ok := FindGap(predecessor, target)
	if !ok {
		return 0, ErrNoSlotAvailable
	}
	return result, nil
}

// SiblingNumberAfter returns a sibling number to insert after the target.
func SiblingNumberAfter(occupied []int, target int) (int, error) {
	sorted := sortedCopy(occupied)

	if len(sorted) >= maxSibling {
		return 0, ErrMaxSiblingsReached
	}

	idx := sort.SearchInts(sorted, target)

	var successor int
	if idx < len(sorted)-1 {
		successor = sorted[idx+1]
	} else {
		successor = maxSibling + 1
	}

	directGap, directOK := FindGap(target, successor)

	// Check the next gap for a higher-tier result.
	if successor <= maxSibling {
		var secondSuccessor int
		if idx+2 < len(sorted) {
			secondSuccessor = sorted[idx+2]
		} else {
			secondSuccessor = maxSibling + 1
		}
		skipGap, skipOK := FindGap(successor, secondSuccessor)
		if skipOK && (!directOK || tierOf(skipGap) > tierOf(directGap)) {
			return skipGap, nil
		}
	}

	if directOK {
		return directGap, nil
	}

	return 0, ErrNoSlotAvailable
}

// CompactNumbers returns a renumbered sequence at initial spacing.
func CompactNumbers(count int) ([]int, error) {
	if count > maxSibling/initialSpacing {
		return nil, ErrMaxSiblingsReached
	}
	result := make([]int, count)
	for i := range result {
		result[i] = (i + 1) * initialSpacing
	}
	return result, nil
}
