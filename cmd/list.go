package cmd

import (
	"context"
	"fmt"
	"io"

	"github.com/eykd/linemark-go/internal/domain"
	"github.com/spf13/cobra"
)

// ListResult holds the outcome of a list operation.
type ListResult struct {
	Outline domain.Outline
}

// ListRunner defines the interface for running the list operation.
type ListRunner interface {
	List(ctx context.Context) (*ListResult, error)
}

// treeNode represents a node in the hierarchical tree for display.
type treeNode struct {
	MP       string      `json:"mp"`
	SID      string      `json:"sid"`
	Title    string      `json:"title"`
	Depth    int         `json:"depth"`
	Types    []string    `json:"types"`
	Children []*treeNode `json:"children"`
}

// treeOutput is the top-level JSON structure for list output.
type treeOutput struct {
	Nodes []*treeNode `json:"nodes"`
}

// NewListCmd creates the list command with the given runner.
func NewListCmd(runner ListRunner) *cobra.Command {
	var jsonOutput bool
	var depth int
	var typeFilter string

	cmd := &cobra.Command{
		Use:          "list",
		Short:        "Display the project outline as a tree",
		SilenceUsage: true,
		RunE: func(cmd *cobra.Command, args []string) error {
			if runner == nil {
				return ErrNotInProject
			}
			result, err := runner.List(cmd.Context())
			if err != nil {
				return err
			}

			nodes := result.Outline.Nodes

			if typeFilter != "" {
				nodes = filterByType(nodes, typeFilter)
			}

			roots := buildTree(nodes, depth)

			if jsonOutput || GetJSON() {
				writeJSON(cmd.OutOrStdout(), &treeOutput{Nodes: roots})
			} else {
				renderTreeText(cmd.OutOrStdout(), roots)
			}
			return nil
		},
	}

	cmd.Flags().BoolVar(&jsonOutput, "json", false, "Output results as JSON")
	cmd.Flags().IntVar(&depth, "depth", 0, "Maximum display depth (0 = unlimited)")
	cmd.Flags().StringVar(&typeFilter, "type", "", "Filter nodes by document type")

	return cmd
}

// filterByType returns only nodes that have a document of the specified type.
func filterByType(nodes []domain.Node, docType string) []domain.Node {
	var filtered []domain.Node
	for _, n := range nodes {
		for _, d := range n.Documents {
			if d.Type == docType {
				filtered = append(filtered, n)
				break
			}
		}
	}
	return filtered
}

// buildTree converts a flat sorted list of nodes into a hierarchical tree.
func buildTree(nodes []domain.Node, maxDepth int) []*treeNode {
	nodeMap := make(map[string]*treeNode)
	roots := []*treeNode{}

	for _, n := range nodes {
		if maxDepth > 0 && n.MP.Depth() > maxDepth {
			continue
		}

		mpStr := n.MP.String()
		tn := &treeNode{
			MP:       mpStr,
			SID:      n.SID,
			Title:    n.Title,
			Depth:    n.MP.Depth(),
			Types:    extractDocTypes(n.Documents),
			Children: []*treeNode{},
		}
		nodeMap[mpStr] = tn

		parent, hasParent := n.MP.Parent()
		if hasParent {
			if parentNode, ok := nodeMap[parent.String()]; ok {
				parentNode.Children = append(parentNode.Children, tn)
				continue
			}
		}
		roots = append(roots, tn)
	}

	return roots
}

// extractDocTypes returns the document type strings from a node's documents.
func extractDocTypes(docs []domain.Document) []string {
	types := make([]string, len(docs))
	for i, d := range docs {
		types[i] = d.Type
	}
	return types
}

// renderTreeText writes the tree display with box-drawing characters.
// When exactly one root has children, subsequent roots are merged as siblings
// of that root's children into a single connected tree. When multiple roots
// have children (or none do), each root is rendered independently at the top level.
func renderTreeText(w io.Writer, roots []*treeNode) {
	if len(roots) == 0 {
		return
	}

	rootsWithChildren := 0
	for _, r := range roots {
		if len(r.Children) > 0 {
			rootsWithChildren++
		}
	}

	if len(roots) > 1 && rootsWithChildren == 1 {
		first := roots[0]
		fmt.Fprintf(w, "%s (%s)\n", first.Title, first.SID)
		allChildren := make([]*treeNode, 0, len(first.Children)+len(roots)-1)
		allChildren = append(allChildren, first.Children...)
		allChildren = append(allChildren, roots[1:]...)
		renderChildren(w, allChildren, "")
	} else {
		for _, root := range roots {
			fmt.Fprintf(w, "%s (%s)\n", root.Title, root.SID)
			renderChildren(w, root.Children, "")
		}
	}
}

// renderChildren recursively renders child nodes with tree-drawing prefixes.
func renderChildren(w io.Writer, children []*treeNode, prefix string) {
	for i, child := range children {
		isLast := i == len(children)-1
		connector := "├── "
		if isLast {
			connector = "└── "
		}
		fmt.Fprintf(w, "%s%s%s (%s)\n", prefix, connector, child.Title, child.SID)

		childPrefix := prefix + "│   "
		if isLast {
			childPrefix = prefix + "    "
		}
		renderChildren(w, child.Children, childPrefix)
	}
}
