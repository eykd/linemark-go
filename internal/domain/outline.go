package domain

import (
	"fmt"
	"sort"
)

// Outline represents the project's document tree.
type Outline struct {
	Nodes []Node
}

// BuildOutline constructs an Outline from a list of ParsedFiles.
// Files are grouped by SID into Nodes, sorted by MaterializedPath.
// Returns findings for issues like duplicate SIDs.
func BuildOutline(files []ParsedFile) (Outline, []Finding, error) {
	outline := Outline{Nodes: []Node{}}
	if len(files) == 0 {
		return outline, nil, nil
	}

	var findings []Finding

	type fileGroup struct {
		mp    string
		sid   string
		files []ParsedFile
	}
	groups := make(map[string]*fileGroup)
	var order []string

	for _, f := range files {
		g, exists := groups[f.SID]
		if exists {
			if g.mp != f.MP {
				findings = append(findings, Finding{
					Type:     FindingDuplicateSID,
					Severity: SeverityWarning,
					Message:  fmt.Sprintf("SID %s at MPs %s and %s", f.SID, g.mp, f.MP),
				})
			}
		} else {
			g = &fileGroup{mp: f.MP, sid: f.SID}
			groups[f.SID] = g
			order = append(order, f.SID)
		}
		g.files = append(g.files, f)
	}

	for _, sid := range order {
		g := groups[sid]
		mp, err := NewMaterializedPath(g.mp)
		if err != nil {
			return Outline{}, nil, err
		}

		var docs []Document
		var title string
		for _, f := range g.files {
			docs = append(docs, Document{
				Type:     f.DocType,
				Filename: GenerateFilename(f.MP, f.SID, f.DocType, f.Slug),
			})
			if f.DocType == "draft" {
				title = f.Slug
			}
		}

		sort.Slice(docs, func(i, j int) bool {
			return docs[i].Type < docs[j].Type
		})

		outline.Nodes = append(outline.Nodes, Node{
			MP:        mp,
			SID:       g.sid,
			Title:     title,
			Documents: docs,
		})
	}

	sort.Slice(outline.Nodes, func(i, j int) bool {
		return outline.Nodes[i].MP.String() < outline.Nodes[j].MP.String()
	})

	return outline, findings, nil
}
