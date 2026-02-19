package domain

import (
	"testing"
)

func TestMaterializedPath_Child(t *testing.T) {
	tests := []struct {
		name    string
		parent  string
		segment int
		want    string
		wantErr bool
	}{
		{"root adds child", "001", 200, "001-200", false},
		{"nested adds child", "001-200", 10, "001-200-010", false},
		{"deep nesting", "001-200-010", 300, "001-200-010-300", false},
		{"child at 100 spacing", "100", 100, "100-100", false},
		{"child at max", "001", 999, "001-999", false},
		{"child at min", "001", 1, "001-001", false},
		{"segment zero rejected", "001", 0, "", true},
		{"segment negative rejected", "001", -1, "", true},
		{"segment over 999 rejected", "001", 1000, "", true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mp, err := NewMaterializedPath(tt.parent)
			if err != nil {
				t.Fatalf("setup: unexpected error: %v", err)
			}

			child, err := mp.Child(tt.segment)
			if (err != nil) != tt.wantErr {
				t.Fatalf("Child(%d) error = %v, wantErr %v", tt.segment, err, tt.wantErr)
			}
			if !tt.wantErr {
				if child.String() != tt.want {
					t.Errorf("Child(%d) = %q, want %q", tt.segment, child.String(), tt.want)
				}
			}
		})
	}
}

func TestMaterializedPath_LastSegment(t *testing.T) {
	tests := []struct {
		name  string
		input string
		want  int
	}{
		{"single segment", "001", 1},
		{"single segment high", "999", 999},
		{"two segments last is 200", "001-200", 200},
		{"three segments last is 10", "001-200-010", 10},
		{"four segments last is 400", "100-200-300-400", 400},
		{"exact 100 spacing", "100", 100},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mp, err := NewMaterializedPath(tt.input)
			if err != nil {
				t.Fatalf("setup: unexpected error: %v", err)
			}
			if got := mp.LastSegment(); got != tt.want {
				t.Errorf("LastSegment() = %d, want %d", got, tt.want)
			}
		})
	}
}

func TestMaterializedPath_IsAncestorOf(t *testing.T) {
	tests := []struct {
		name     string
		ancestor string
		other    string
		want     bool
	}{
		{"parent is ancestor of child", "001", "001-200", true},
		{"grandparent is ancestor", "001", "001-200-010", true},
		{"great-grandparent is ancestor", "001", "001-200-010-300", true},
		{"not ancestor different root", "002", "001-200", false},
		{"not ancestor of self", "001", "001", false},
		{"not ancestor when reversed", "001-200", "001", false},
		{"not ancestor different subtree", "001-100", "001-200", false},
		{"prefix match but not ancestor", "001-200", "001-200-010", true},
		{"similar prefix not ancestor", "001-20", "001-200", false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ancestor, err := NewMaterializedPath(tt.ancestor)
			if err != nil {
				t.Fatalf("setup ancestor: unexpected error: %v", err)
			}
			other, err := NewMaterializedPath(tt.other)
			if err != nil {
				t.Fatalf("setup other: unexpected error: %v", err)
			}
			if got := ancestor.IsAncestorOf(other); got != tt.want {
				t.Errorf("%q.IsAncestorOf(%q) = %v, want %v",
					tt.ancestor, tt.other, got, tt.want)
			}
		})
	}
}

func TestMaterializedPath_Equal(t *testing.T) {
	tests := []struct {
		name  string
		a     string
		b     string
		equal bool
	}{
		{"same single segment", "001", "001", true},
		{"same two segments", "001-200", "001-200", true},
		{"same three segments", "001-200-010", "001-200-010", true},
		{"different single segments", "001", "002", false},
		{"different second segment", "001-200", "001-300", false},
		{"different depths", "001", "001-200", false},
		{"reversed different depths", "001-200", "001", false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			a, err := NewMaterializedPath(tt.a)
			if err != nil {
				t.Fatalf("setup a: unexpected error: %v", err)
			}
			b, err := NewMaterializedPath(tt.b)
			if err != nil {
				t.Fatalf("setup b: unexpected error: %v", err)
			}
			if got := a.Equal(b); got != tt.equal {
				t.Errorf("%q.Equal(%q) = %v, want %v", tt.a, tt.b, got, tt.equal)
			}
		})
	}
}
