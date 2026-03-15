package treeview

// TreeNode represents a node in the hierarchical tree.
//
// Each node has a unique ID, a display label, optional children, and an
// Expanded flag controlling child visibility. The Data field carries
// arbitrary user payloads.
//
// Nodes form a tree: a node with non-nil Children is a branch node;
// one with nil or empty Children is a leaf.
type TreeNode struct {
	// ID uniquely identifies this node within the tree.
	// Must be unique across all nodes.
	ID string

	// Label is the display text for this node.
	Label string

	// Children are this node's child nodes. Nil or empty means leaf.
	Children []*TreeNode

	// Expanded controls whether this node's children are visible.
	// Only meaningful for branch nodes (len(Children) > 0).
	Expanded bool

	// Data carries an arbitrary user payload.
	Data interface{}
}

// IsLeaf reports whether this node has no children.
func (n *TreeNode) IsLeaf() bool {
	return len(n.Children) == 0
}

// flatRow represents a single visible row in the flattened tree.
// It pairs a node reference with its nesting depth for rendering.
type flatRow struct {
	node  *TreeNode
	depth int
}

// flattenTree flattens the tree starting from root into a slice of visible rows.
// Only expanded branches contribute their children to the output.
// The root node itself is included at depth 0.
func flattenTree(root *TreeNode) []flatRow {
	if root == nil {
		return nil
	}
	// Pre-allocate with a reasonable estimate.
	rows := make([]flatRow, 0, 32)
	rows = flattenNode(rows, root, 0)
	return rows
}

// flattenNode recursively appends node and its visible descendants to rows.
func flattenNode(rows []flatRow, node *TreeNode, depth int) []flatRow {
	rows = append(rows, flatRow{node: node, depth: depth})
	if node.Expanded && len(node.Children) > 0 {
		for _, child := range node.Children {
			rows = flattenNode(rows, child, depth+1)
		}
	}
	return rows
}

// findNodeByID searches the tree rooted at node for a node with the given ID.
// Returns nil if not found.
func findNodeByID(node *TreeNode, id string) *TreeNode {
	if node == nil {
		return nil
	}
	if node.ID == id {
		return node
	}
	for _, child := range node.Children {
		if found := findNodeByID(child, id); found != nil {
			return found
		}
	}
	return nil
}

// findParent searches the tree rooted at root for the parent of the node
// with the given ID. Returns nil if the node is the root or not found.
func findParent(root *TreeNode, id string) *TreeNode {
	if root == nil {
		return nil
	}
	for _, child := range root.Children {
		if child.ID == id {
			return root
		}
		if found := findParent(child, id); found != nil {
			return found
		}
	}
	return nil
}

// SelectionMode defines how nodes can be selected in the tree.
type SelectionMode uint8

// SelectionMode constants.
const (
	// SelectionNone disables node selection. This is the default.
	SelectionNone SelectionMode = iota

	// SelectionSingle allows at most one node to be selected at a time.
	SelectionSingle
)

// String returns a human-readable name for the selection mode.
func (m SelectionMode) String() string {
	switch m {
	case SelectionNone:
		return "None"
	case SelectionSingle:
		return "Single"
	default:
		return "Unknown"
	}
}
