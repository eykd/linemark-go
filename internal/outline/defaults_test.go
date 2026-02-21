package outline

import (
	"github.com/eykd/linemark-go/internal/frontmatter"
	"github.com/eykd/linemark-go/internal/slug"
)

// slugAdapter wraps slug.Slug as a Slugifier implementation.
type slugAdapter struct{}

func (slugAdapter) Slug(s string) string { return slug.Slug(s) }

// fmAdapter wraps frontmatter package functions as a FrontmatterHandler implementation.
type fmAdapter struct{}

func (fmAdapter) GetTitle(input string) (string, error) { return frontmatter.GetTitle(input) }
func (fmAdapter) SetTitle(input, newTitle string) (string, error) {
	return frontmatter.SetTitle(input, newTitle)
}
func (fmAdapter) EncodeYAMLValue(s string) string  { return frontmatter.EncodeYAMLValue(s) }
func (fmAdapter) Serialize(fm, body string) string { return frontmatter.Serialize(fm, body) }

func init() {
	defaultSlugifier = slugAdapter{}
	defaultFMHandler = fmAdapter{}
}
