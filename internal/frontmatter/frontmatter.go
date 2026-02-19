package frontmatter

import (
	"errors"
	"strings"

	"gopkg.in/yaml.v3"
)

// Split separates a document into frontmatter and body components.
// Frontmatter is delimited by --- on its own line.
func Split(input string) (string, string, error) {
	if input == "" {
		return "", "", nil
	}
	if !strings.HasPrefix(input, "---\n") {
		return "", input, nil
	}

	rest := input[4:]
	pos := 0
	for pos < len(rest) {
		nlIdx := strings.IndexByte(rest[pos:], '\n')

		var line string
		var nextPos int
		if nlIdx < 0 {
			line = rest[pos:]
			nextPos = len(rest)
		} else {
			line = rest[pos : pos+nlIdx]
			nextPos = pos + nlIdx + 1
		}

		if line == "---" {
			if nlIdx < 0 {
				return rest[:pos], "", nil
			}
			return rest[:pos], rest[nextPos:], nil
		}

		pos = nextPos
	}

	return "", "", errors.New("unclosed frontmatter")
}

// findKeyIndex returns the index of a key in a YAML mapping node, or -1 if not found.
func findKeyIndex(mapping *yaml.Node, key string) int {
	for i := 0; i < len(mapping.Content)-1; i += 2 {
		if mapping.Content[i].Value == key {
			return i
		}
	}
	return -1
}

// GetTitle extracts the title field from a document's YAML frontmatter.
func GetTitle(input string) (string, error) {
	fm, _, err := Split(input)
	if fm == "" {
		return "", err
	}

	var doc yaml.Node
	if err := yaml.Unmarshal([]byte(fm), &doc); err != nil {
		return "", err
	}

	idx := findKeyIndex(doc.Content[0], "title")
	if idx < 0 {
		return "", nil
	}

	val := doc.Content[0].Content[idx+1]
	if val.Tag != "!!str" {
		return "", errors.New("title is not a string")
	}
	return val.Value, nil
}

// SetTitle sets or updates the title field in a document's YAML frontmatter.
// It preserves unknown fields, field order, and comments using text-level
// line replacement guided by yaml.Node line positions.
func SetTitle(input string, newTitle string) (string, error) {
	fm, body, err := Split(input)
	if err != nil {
		return "", err
	}

	titleLine := "title: " + encodeYAMLValue(newTitle) + "\n"

	if fm == "" {
		return Serialize(titleLine, body), nil
	}

	var doc yaml.Node
	if err := yaml.Unmarshal([]byte(fm), &doc); err != nil {
		return "", err
	}

	mapping := doc.Content[0]
	idx := findKeyIndex(mapping, "title")
	if idx >= 0 {
		keyLine := mapping.Content[idx].Line
		lines := strings.SplitAfter(fm, "\n")
		lines[keyLine-1] = titleLine
		return Serialize(strings.Join(lines, ""), body), nil
	}

	return Serialize(fm+titleLine, body), nil
}

// encodeYAMLValue encodes a string as a safe YAML scalar value.
// Strings containing newlines, colons, double quotes, backslashes,
// or leading # use double-quoted style with escape sequences that
// prevent YAML injection.
func encodeYAMLValue(s string) string {
	if !strings.ContainsAny(s, "\n:\"\\") && !strings.HasPrefix(s, "#") {
		return s
	}

	hasNewline := strings.Contains(s, "\n")

	var b strings.Builder
	b.WriteByte('"')
	for _, c := range s {
		switch c {
		case '"':
			b.WriteString(`\"`)
		case '\\':
			b.WriteString(`\\`)
		case '\n':
			b.WriteString(`\n`)
		case ':':
			if hasNewline {
				b.WriteString(`\x3a`)
			} else {
				b.WriteRune(c)
			}
		default:
			b.WriteRune(c)
		}
	}
	b.WriteByte('"')
	return b.String()
}

// Serialize combines frontmatter and body into a complete document.
func Serialize(fm string, body string) string {
	return "---\n" + fm + "---\n" + body
}
