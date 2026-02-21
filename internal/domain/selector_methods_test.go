package domain

import (
	"testing"
)

func TestSelector_Explicit(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		explicit bool
	}{
		{"implicit MP", "001-200", false},
		{"implicit SID", "A3F7c9Qx7Lm2", false},
		{"explicit MP", "mp:001-200", true},
		{"explicit SID", "sid:A3F7c9Qx7Lm2", true},
		{"implicit single segment MP", "001", false},
		{"explicit single segment MP", "mp:001", true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			sel, err := ParseSelector(tt.input)
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if got := sel.Explicit(); got != tt.explicit {
				t.Errorf("Explicit() = %v, want %v", got, tt.explicit)
			}
		})
	}
}
