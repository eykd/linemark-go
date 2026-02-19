package domain

// DeleteMode specifies how child nodes are handled during deletion.
type DeleteMode int

const (
	// DeleteModeDefault deletes only leaf nodes; errors if the node has children.
	DeleteModeDefault DeleteMode = iota
	// DeleteModeRecursive deletes the node and its entire subtree.
	DeleteModeRecursive
	// DeleteModePromote deletes the node and promotes children to the parent level.
	DeleteModePromote
)
