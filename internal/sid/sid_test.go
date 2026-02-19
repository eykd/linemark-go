package sid_test

import (
	"bytes"
	"crypto/rand"
	"errors"
	"strings"
	"testing"

	"github.com/eykd/linemark-go/internal/sid"
)

const base62Alphabet = "ABCDEFGHIJKLMNOPQRSTUVWXYZabcdefghijklmnopqrstuvwxyz0123456789"

func TestGenerate_ValidOutput(t *testing.T) {
	tests := []struct {
		name  string
		input []byte
		want  string
	}{
		{
			name:  "all zeros map to first character",
			input: []byte{0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0},
			want:  "AAAAAAAAAAAA",
		},
		{
			name:  "sequential bytes map to sequential chars",
			input: []byte{0, 1, 2, 3, 4, 5, 6, 7, 8, 9, 10, 11},
			want:  "ABCDEFGHIJKL",
		},
		{
			name:  "byte 61 maps to last alphabet char",
			input: []byte{61, 61, 61, 61, 61, 61, 61, 61, 61, 61, 61, 61},
			want:  "999999999999",
		},
		{
			name:  "byte 62 wraps to first char",
			input: []byte{62, 62, 62, 62, 62, 62, 62, 62, 62, 62, 62, 62},
			want:  "AAAAAAAAAAAA",
		},
		{
			name:  "byte 247 maps to last char (max valid byte)",
			input: []byte{247, 247, 247, 247, 247, 247, 247, 247, 247, 247, 247, 247},
			want:  "999999999999",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			r := bytes.NewReader(tt.input)
			got, err := sid.Generate(r)
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if got != tt.want {
				t.Errorf("Generate() = %q, want %q", got, tt.want)
			}
		})
	}
}

func TestGenerate_RejectionSampling(t *testing.T) {
	tests := []struct {
		name  string
		input []byte
		want  string
	}{
		{
			name:  "byte 248 is rejected",
			input: []byte{248, 0, 1, 2, 3, 4, 5, 6, 7, 8, 9, 10, 11},
			want:  "ABCDEFGHIJKL",
		},
		{
			name:  "byte 255 is rejected",
			input: []byte{255, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0},
			want:  "AAAAAAAAAAAA",
		},
		{
			name:  "multiple rejections before acceptance",
			input: []byte{248, 249, 250, 251, 252, 253, 254, 255, 0, 1, 2, 3, 4, 5, 6, 7, 8, 9, 10, 11},
			want:  "ABCDEFGHIJKL",
		},
		{
			name:  "byte 247 accepted then zeros",
			input: []byte{247, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0},
			want:  "9AAAAAAAAAAA",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			r := bytes.NewReader(tt.input)
			got, err := sid.Generate(r)
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if got != tt.want {
				t.Errorf("Generate() = %q, want %q", got, tt.want)
			}
		})
	}
}

func TestGenerate_Length(t *testing.T) {
	got, err := sid.Generate(rand.Reader)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(got) != 12 {
		t.Errorf("len(Generate()) = %d, want 12", len(got))
	}
}

func TestGenerate_AlphabetOnly(t *testing.T) {
	got, err := sid.Generate(rand.Reader)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	for i, c := range got {
		if !strings.ContainsRune(base62Alphabet, c) {
			t.Errorf("character at index %d: %q not in base62 alphabet", i, c)
		}
	}
}

func TestGenerate_Uniqueness(t *testing.T) {
	seen := make(map[string]bool)
	for i := 0; i < 100; i++ {
		got, err := sid.Generate(rand.Reader)
		if err != nil {
			t.Fatalf("unexpected error on iteration %d: %v", i, err)
		}
		if seen[got] {
			t.Fatalf("duplicate SID on iteration %d: %q", i, got)
		}
		seen[got] = true
	}
}

func TestGenerate_UnbiasedDistribution(t *testing.T) {
	charSeen := make(map[byte]bool)
	for i := 0; i < 1000; i++ {
		got, err := sid.Generate(rand.Reader)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		for j := 0; j < len(got); j++ {
			charSeen[got[j]] = true
		}
	}
	for i := 0; i < len(base62Alphabet); i++ {
		if !charSeen[base62Alphabet[i]] {
			t.Errorf("character %q never appeared in 1000 generated SIDs", base62Alphabet[i])
		}
	}
}

func TestGenerate_ReaderError(t *testing.T) {
	errRead := errors.New("read failed")
	r := &failingReader{err: errRead}
	_, err := sid.Generate(r)
	if err == nil {
		t.Fatal("expected error, got nil")
	}
	if !errors.Is(err, errRead) {
		t.Errorf("error = %v, want %v", err, errRead)
	}
}

type failingReader struct {
	err error
}

func (r *failingReader) Read(p []byte) (int, error) {
	return 0, r.err
}

func TestGenerate_ReaderExhausted(t *testing.T) {
	// Only 6 valid bytes — not enough for a 12-char SID.
	r := bytes.NewReader([]byte{0, 1, 2, 3, 4, 5})
	_, err := sid.Generate(r)
	if err == nil {
		t.Fatal("expected error when reader exhausted, got nil")
	}
}
