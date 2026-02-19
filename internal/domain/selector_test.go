package domain

import (
	"errors"
	"testing"
)

func TestParseSelector_ImplicitMP(t *testing.T) {
	tests := []struct {
		name      string
		input     string
		wantValue string
	}{
		{"single segment", "001", "001"},
		{"two segments", "001-200", "001-200"},
		{"three segments", "001-200-010", "001-200-010"},
		{"four segments", "100-200-300-400", "100-200-300-400"},
		{"max segment value", "999", "999"},
		{"minimum non-zero", "001-001-001", "001-001-001"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			sel, err := ParseSelector(tt.input)
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if sel.Kind() != SelectorMP {
				t.Errorf("Kind() = %v, want SelectorMP", sel.Kind())
			}
			if sel.Value() != tt.wantValue {
				t.Errorf("Value() = %q, want %q", sel.Value(), tt.wantValue)
			}
		})
	}
}

func TestParseSelector_ImplicitSID(t *testing.T) {
	tests := []struct {
		name      string
		input     string
		wantValue string
	}{
		{"twelve char mixed", "A3F7c9Qx7Lm2", "A3F7c9Qx7Lm2"},
		{"eight char minimum", "abcdefgh", "abcdefgh"},
		{"nine chars", "Abc123Def", "Abc123Def"},
		{"ten chars", "Abc123Def0", "Abc123Def0"},
		{"eleven chars", "Abc123Def01", "Abc123Def01"},
		{"twelve chars", "Abc123Def012", "Abc123Def012"},
		{"all digits twelve", "123456789012", "123456789012"},
		{"all lowercase eight", "abcdefgh", "abcdefgh"},
		{"all uppercase eight", "ABCDEFGH", "ABCDEFGH"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			sel, err := ParseSelector(tt.input)
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if sel.Kind() != SelectorSID {
				t.Errorf("Kind() = %v, want SelectorSID", sel.Kind())
			}
			if sel.Value() != tt.wantValue {
				t.Errorf("Value() = %q, want %q", sel.Value(), tt.wantValue)
			}
		})
	}
}

func TestParseSelector_ExplicitMP(t *testing.T) {
	tests := []struct {
		name      string
		input     string
		wantValue string
	}{
		{"explicit single segment", "mp:001", "001"},
		{"explicit two segments", "mp:001-200", "001-200"},
		{"explicit three segments", "mp:001-200-010", "001-200-010"},
		{"explicit max value", "mp:999", "999"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			sel, err := ParseSelector(tt.input)
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if sel.Kind() != SelectorMP {
				t.Errorf("Kind() = %v, want SelectorMP", sel.Kind())
			}
			if sel.Value() != tt.wantValue {
				t.Errorf("Value() = %q, want %q", sel.Value(), tt.wantValue)
			}
		})
	}
}

func TestParseSelector_ExplicitSID(t *testing.T) {
	tests := []struct {
		name      string
		input     string
		wantValue string
	}{
		{"explicit twelve char", "sid:A3F7c9Qx7Lm2", "A3F7c9Qx7Lm2"},
		{"explicit eight char", "sid:abcdefgh", "abcdefgh"},
		{"explicit all digits", "sid:123456789012", "123456789012"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			sel, err := ParseSelector(tt.input)
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if sel.Kind() != SelectorSID {
				t.Errorf("Kind() = %v, want SelectorSID", sel.Kind())
			}
			if sel.Value() != tt.wantValue {
				t.Errorf("Value() = %q, want %q", sel.Value(), tt.wantValue)
			}
		})
	}
}

func TestParseSelector_MPPrecedenceOverSID(t *testing.T) {
	// MP pattern: ^\d{3}(?:-\d{3})*$  (lengths 3, 7, 11, ...)
	// SID pattern: ^[A-Za-z0-9]{8,12}$ (lengths 8-12)
	// In practice these patterns are mutually exclusive (MP requires dashes
	// for length > 3, and dashes are not alphanumeric). This test verifies
	// that the parser checks MP first, so valid MPs are never misidentified.
	tests := []struct {
		name  string
		input string
	}{
		{"three digit MP not SID", "001"},
		{"seven char MP not SID", "001-200"},
		{"eleven char MP not SID", "001-200-010"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			sel, err := ParseSelector(tt.input)
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if sel.Kind() != SelectorMP {
				t.Errorf("Kind() = %v, want SelectorMP (MP takes precedence)", sel.Kind())
			}
		})
	}
}

func TestParseSelector_Invalid(t *testing.T) {
	tests := []struct {
		name  string
		input string
	}{
		{"empty string", ""},
		{"whitespace only", "   "},
		{"too short for either", "ab"},
		{"too short alphanumeric", "abc"},
		{"too long for SID no MP match", "ABCDEFGHIJKLM"},
		{"special characters", "abc!@#$efgh"},
		{"contains underscore", "abc_defgh"},
		{"mp prefix empty value", "mp:"},
		{"mp prefix invalid value", "mp:invalid"},
		{"mp prefix zero segment", "mp:000"},
		{"mp prefix with SID-like value", "mp:A3F7c9Qx7Lm2"},
		{"sid prefix empty value", "sid:"},
		{"sid prefix too short value", "sid:abc"},
		{"sid prefix special chars", "sid:abc!defgh"},
		{"sid prefix with MP-like value", "sid:001-200"},
		{"unknown prefix", "foo:001-200"},
		{"spaces in value", "001 200"},
		{"newline in value", "001\n200"},
		{"colon without known prefix", "x:001"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := ParseSelector(tt.input)
			if err == nil {
				t.Error("expected error, got nil")
			}
			if !errors.Is(err, ErrInvalidSelector) {
				t.Errorf("error = %v, want ErrInvalidSelector", err)
			}
		})
	}
}

func TestParseSelector_String(t *testing.T) {
	tests := []struct {
		name  string
		input string
		want  string
	}{
		{"implicit MP", "001-200", "001-200"},
		{"implicit SID", "A3F7c9Qx7Lm2", "A3F7c9Qx7Lm2"},
		{"explicit MP", "mp:001-200", "mp:001-200"},
		{"explicit SID", "sid:A3F7c9Qx7Lm2", "sid:A3F7c9Qx7Lm2"},
		{"implicit single segment MP", "001", "001"},
		{"explicit single segment MP", "mp:001", "mp:001"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			sel, err := ParseSelector(tt.input)
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if got := sel.String(); got != tt.want {
				t.Errorf("String() = %q, want %q", got, tt.want)
			}
		})
	}
}
