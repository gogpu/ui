package layout

import (
	"github.com/gogpu/ui/geometry"
)

// NodeID identifies a node in the layout tree.
//
// NodeID is an opaque identifier used by layout algorithms to reference
// nodes in the tree. The actual node storage and management is handled
// by the [LayoutTree] implementation.
type NodeID uint64

// InvalidNodeID represents an invalid or unset node identifier.
const InvalidNodeID NodeID = 0

// IsValid returns true if the node ID is valid (non-zero).
func (n NodeID) IsValid() bool {
	return n != InvalidNodeID
}

// NodeLayout is the computed layout for a single node.
//
// After a layout algorithm runs, each node will have a NodeLayout
// describing its position relative to its parent and its computed size.
type NodeLayout struct {
	// Position is the offset from the parent node's origin.
	Position geometry.Point

	// Size is the computed size of the node.
	Size geometry.Size
}

// Bounds returns the layout as a rectangle.
func (n NodeLayout) Bounds() geometry.Rect {
	return geometry.FromPointSize(n.Position, n.Size)
}

// IsZero returns true if the layout has zero position and size.
func (n NodeLayout) IsZero() bool {
	return n.Position.IsZero() && n.Size.IsZero()
}

// Result is the output of a layout algorithm computation.
//
// Result contains the computed size of the root node and indicates
// whether the layout was successful.
type Result struct {
	// Size is the computed size of the root node.
	Size geometry.Size

	// Overflow indicates if content exceeds the available space.
	Overflow bool
}

// IsZero returns true if the result has zero size.
func (r Result) IsZero() bool {
	return r.Size.IsZero()
}
