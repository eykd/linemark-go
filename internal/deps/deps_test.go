package deps_test

import (
	"testing"

	"github.com/gofrs/flock"
	"golang.org/x/text/unicode/norm"
	"gopkg.in/yaml.v3"
)

// TestYAMLDependencyAvailable verifies that gopkg.in/yaml.v3 is importable
// and functional for frontmatter parsing.
func TestYAMLDependencyAvailable(t *testing.T) {
	input := "title: hello"
	var node yaml.Node
	err := yaml.Unmarshal([]byte(input), &node)
	if err != nil {
		t.Fatalf("yaml.Unmarshal() returned error: %v", err)
	}
	if node.Kind != yaml.DocumentNode {
		t.Errorf("yaml.Node.Kind = %v, want %v (DocumentNode)", node.Kind, yaml.DocumentNode)
	}
}

// TestFlockDependencyAvailable verifies that github.com/gofrs/flock is
// importable and can construct a lock handle.
func TestFlockDependencyAvailable(t *testing.T) {
	fl := flock.New(t.TempDir() + "/test.lock")
	if fl == nil {
		t.Fatal("flock.New() returned nil")
	}
	path := fl.Path()
	if path == "" {
		t.Error("flock.Path() returned empty string")
	}
}

// TestUnicodeTextDependencyAvailable verifies that golang.org/x/text is
// importable and can perform NFC normalization for slug generation.
func TestUnicodeTextDependencyAvailable(t *testing.T) {
	// NFC normalization of a combining sequence: e + combining acute = é
	input := "e\u0301" // decomposed form
	got := norm.NFC.String(input)
	want := "\u00e9" // composed form: é
	if got != want {
		t.Errorf("norm.NFC.String(%q) = %q, want %q", input, got, want)
	}
}
