package layout

import (
	"github.com/gogpu/ui/geometry"
)

// LayoutTree provides access to the node tree for layout algorithms.
//
// This interface abstracts the node tree structure, allowing layout algorithms
// to traverse nodes, access style properties, measure content, and set computed
// layouts without knowing the underlying node implementation.
//
// # Usage by Layout Algorithms
//
// Layout algorithms use LayoutTree to:
//  1. Traverse the tree via ChildCount() and ChildAt()
//  2. Get layout styles via Style()
//  3. Measure leaf content via Measure()
//  4. Store computed layouts via SetLayout()
//
// # Implementation Requirements
//
// Implementations must:
//   - Return consistent child counts and indices
//   - Support Measure() for leaf nodes (nodes with no children)
//   - Accept SetLayout() calls for any valid NodeID
//   - Return non-nil Style for all nodes (use default Style if none set)
type LayoutTree interface {
	// Style returns the layout style for a node.
	// Returns a default Style if the node has no explicit style.
	// Never returns nil.
	Style(node NodeID) *Style

	// SetLayout sets the computed layout for a node.
	// This should be called by layout algorithms after computing positions.
	SetLayout(node NodeID, layout NodeLayout)

	// GetLayout returns the previously set layout for a node.
	// Returns a zero NodeLayout if no layout has been set.
	GetLayout(node NodeID) NodeLayout

	// ChildCount returns the number of children for a node.
	// Returns 0 for leaf nodes.
	ChildCount(parent NodeID) int

	// ChildAt returns the child at the given index.
	// Returns InvalidNodeID if index is out of bounds.
	// Index is 0-based.
	ChildAt(parent NodeID, index int) NodeID

	// Measure measures a node's content size given constraints.
	// For leaf nodes, this returns the intrinsic content size.
	// For container nodes, this may trigger a recursive layout.
	Measure(node NodeID, constraints geometry.Constraints) geometry.Size
}

// LayoutTreeAdapter helps implement LayoutTree for existing data structures.
//
// This is a convenience helper that provides default implementations for
// some LayoutTree methods. Embed this in your implementation and override
// methods as needed.
type LayoutTreeAdapter struct {
	// Styles maps node IDs to their styles.
	Styles map[NodeID]*Style

	// Layouts maps node IDs to their computed layouts.
	Layouts map[NodeID]NodeLayout

	// defaultStyle is returned when a node has no explicit style.
	defaultStyle Style
}

// NewLayoutTreeAdapter creates a new adapter with initialized maps.
func NewLayoutTreeAdapter() *LayoutTreeAdapter {
	return &LayoutTreeAdapter{
		Styles:  make(map[NodeID]*Style),
		Layouts: make(map[NodeID]NodeLayout),
	}
}

// Style returns the style for a node, or a default style if not set.
func (a *LayoutTreeAdapter) Style(node NodeID) *Style {
	if style, ok := a.Styles[node]; ok {
		return style
	}
	return &a.defaultStyle
}

// SetLayout stores the computed layout for a node.
func (a *LayoutTreeAdapter) SetLayout(node NodeID, layout NodeLayout) {
	if a.Layouts == nil {
		a.Layouts = make(map[NodeID]NodeLayout)
	}
	a.Layouts[node] = layout
}

// GetLayout returns the previously set layout for a node.
func (a *LayoutTreeAdapter) GetLayout(node NodeID) NodeLayout {
	if a.Layouts == nil {
		return NodeLayout{}
	}
	return a.Layouts[node]
}

// SetStyle sets the style for a node.
func (a *LayoutTreeAdapter) SetStyle(node NodeID, style *Style) {
	if a.Styles == nil {
		a.Styles = make(map[NodeID]*Style)
	}
	a.Styles[node] = style
}

// Clear removes all stored styles and layouts.
func (a *LayoutTreeAdapter) Clear() {
	a.Styles = make(map[NodeID]*Style)
	a.Layouts = make(map[NodeID]NodeLayout)
}
