package domain

import (
	"errors"
	"fmt"
	"regexp"
	"strings"
)

// ErrInvalidSelector is returned when a selector string cannot be parsed.
var ErrInvalidSelector = errors.New("invalid selector")

// SelectorKind distinguishes between MP and SID selectors.
type SelectorKind int

const (
	// SelectorMP indicates a materialized path selector.
	SelectorMP SelectorKind = iota
	// SelectorSID indicates a stable ID selector.
	SelectorSID
)

var (
	mpPattern  = regexp.MustCompile(`^\d{3}(?:-\d{3})*$`)
	sidPattern = regexp.MustCompile(`^[A-Za-z0-9]{8,12}$`)
)

// Selector is a value object representing a parsed node reference.
type Selector struct {
	kind     SelectorKind
	value    string
	explicit bool
}

// ParseSelector parses a selector string into a Selector value object.
func ParseSelector(input string) (Selector, error) {
	input = strings.TrimSpace(input)
	if input == "" {
		return Selector{}, fmt.Errorf("%w: empty input", ErrInvalidSelector)
	}

	if strings.HasPrefix(input, "mp:") {
		value := input[3:]
		if !isValidMP(value) {
			return Selector{}, fmt.Errorf("%w: %q", ErrInvalidSelector, input)
		}
		return Selector{kind: SelectorMP, value: value, explicit: true}, nil
	}
	if strings.HasPrefix(input, "sid:") {
		value := input[4:]
		if !isValidSID(value) {
			return Selector{}, fmt.Errorf("%w: %q", ErrInvalidSelector, input)
		}
		return Selector{kind: SelectorSID, value: value, explicit: true}, nil
	}

	if strings.Contains(input, ":") {
		return Selector{}, fmt.Errorf("%w: %q", ErrInvalidSelector, input)
	}

	if isValidMP(input) {
		return Selector{kind: SelectorMP, value: input}, nil
	}
	if isValidSID(input) {
		return Selector{kind: SelectorSID, value: input}, nil
	}

	return Selector{}, fmt.Errorf("%w: %q", ErrInvalidSelector, input)
}

func isValidMP(s string) bool {
	if !mpPattern.MatchString(s) {
		return false
	}
	for _, seg := range strings.Split(s, "-") {
		if seg == "000" {
			return false
		}
	}
	return true
}

func isValidSID(s string) bool {
	return sidPattern.MatchString(s)
}

// Kind returns the selector kind (SelectorMP or SelectorSID).
func (s Selector) Kind() SelectorKind {
	return s.kind
}

// Value returns the selector value without any prefix.
func (s Selector) Value() string {
	return s.value
}

// String returns the string representation, including prefix if explicitly provided.
func (s Selector) String() string {
	if s.explicit {
		switch s.kind {
		case SelectorMP:
			return "mp:" + s.value
		case SelectorSID:
			return "sid:" + s.value
		}
	}
	return s.value
}
