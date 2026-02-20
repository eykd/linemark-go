// Package fs provides filesystem adapters that implement outline service interfaces.
package fs

import (
	"context"
	"fmt"
	"io"
	"os"
	"path/filepath"

	"github.com/eykd/linemark-go/internal/frontmatter"
	"github.com/eykd/linemark-go/internal/sid"
	"github.com/eykd/linemark-go/internal/slug"
)

// OSReader implements outline.DirectoryReader using os.ReadDir.
type OSReader struct {
	Root string
}

// ReadDirImpl reads filenames from the project root directory.
func (r *OSReader) ReadDirImpl(_ context.Context) ([]string, error) {
	entries, err := os.ReadDir(r.Root)
	if err != nil {
		return nil, fmt.Errorf("reading directory %s: %w", r.Root, err)
	}
	var names []string
	for _, e := range entries {
		if !e.IsDir() {
			names = append(names, e.Name())
		}
	}
	return names, nil
}

// ReadDir delegates to ReadDirImpl.
func (r *OSReader) ReadDir(ctx context.Context) ([]string, error) {
	return r.ReadDirImpl(ctx)
}

// OSWriter implements outline.FileWriter using os.WriteFile.
type OSWriter struct {
	Root string
}

// WriteFileImpl writes content to a file under the project root, creating directories as needed.
func (w *OSWriter) WriteFileImpl(_ context.Context, filename, content string) error {
	path := filepath.Join(w.Root, filename)
	dir := filepath.Dir(path)
	if err := os.MkdirAll(dir, 0o755); err != nil {
		return fmt.Errorf("creating directory %s: %w", dir, err)
	}
	return os.WriteFile(path, []byte(content), 0o644)
}

// WriteFile delegates to WriteFileImpl.
func (w *OSWriter) WriteFile(ctx context.Context, filename, content string) error {
	return w.WriteFileImpl(ctx, filename, content)
}

// OSDeleter implements outline.FileDeleter using os.Remove.
type OSDeleter struct {
	Root string
}

// DeleteFileImpl removes a file from the project root.
func (d *OSDeleter) DeleteFileImpl(_ context.Context, filename string) error {
	return os.Remove(filepath.Join(d.Root, filename))
}

// DeleteFile delegates to DeleteFileImpl.
func (d *OSDeleter) DeleteFile(ctx context.Context, filename string) error {
	return d.DeleteFileImpl(ctx, filename)
}

// OSRenamer implements outline.FileRenamer using os.Rename.
type OSRenamer struct {
	Root string
}

// RenameFileImpl renames a file within the project root.
func (r *OSRenamer) RenameFileImpl(_ context.Context, oldName, newName string) error {
	return os.Rename(filepath.Join(r.Root, oldName), filepath.Join(r.Root, newName))
}

// RenameFile delegates to RenameFileImpl.
func (r *OSRenamer) RenameFile(ctx context.Context, oldName, newName string) error {
	return r.RenameFileImpl(ctx, oldName, newName)
}

// OSContentReader implements outline.ContentReader using os.ReadFile.
type OSContentReader struct {
	Root string
}

// ReadFileImpl reads the full content of a file under the project root.
func (cr *OSContentReader) ReadFileImpl(_ context.Context, filename string) (string, error) {
	data, err := os.ReadFile(filepath.Join(cr.Root, filename))
	if err != nil {
		return "", err
	}
	return string(data), nil
}

// ReadFile delegates to ReadFileImpl.
func (cr *OSContentReader) ReadFile(ctx context.Context, filename string) (string, error) {
	return cr.ReadFileImpl(ctx, filename)
}

// SIDReserver implements outline.SIDReserver using a random io.Reader.
type SIDReserver struct {
	Rand io.Reader
}

// ReserveImpl generates a new SID using the configured random source.
func (r *SIDReserver) ReserveImpl(_ context.Context) (string, error) {
	return sid.Generate(r.Rand)
}

// Reserve delegates to ReserveImpl.
func (r *SIDReserver) Reserve(ctx context.Context) (string, error) {
	return r.ReserveImpl(ctx)
}

// SlugAdapter implements outline.Slugifier using the slug package.
type SlugAdapter struct{}

// Slug converts a title to a URL-friendly slug.
func (SlugAdapter) Slug(s string) string { return slug.Slug(s) }

// FMAdapter implements outline.FrontmatterHandler using the frontmatter package.
type FMAdapter struct{}

// GetTitle extracts the title from frontmatter content.
func (FMAdapter) GetTitle(input string) (string, error) { return frontmatter.GetTitle(input) }

// SetTitle updates the title in frontmatter content.
func (FMAdapter) SetTitle(input, newTitle string) (string, error) {
	return frontmatter.SetTitle(input, newTitle)
}

// EncodeYAMLValue encodes a string as a safe YAML scalar.
func (FMAdapter) EncodeYAMLValue(s string) string { return frontmatter.EncodeYAMLValue(s) }

// Serialize combines frontmatter and body into a complete document.
func (FMAdapter) Serialize(fm, body string) string { return frontmatter.Serialize(fm, body) }

// FindProjectRootImpl walks up from the current working directory looking for a .linemark/ directory.
func FindProjectRootImpl() (string, error) {
	dir, err := os.Getwd()
	if err != nil {
		return "", fmt.Errorf("getting working directory: %w", err)
	}

	for {
		candidate := filepath.Join(dir, ".linemark")
		info, err := os.Stat(candidate)
		if err == nil && info.IsDir() {
			return dir, nil
		}

		parent := filepath.Dir(dir)
		if parent == dir {
			return "", fmt.Errorf("no .linemark/ directory found")
		}
		dir = parent
	}
}
