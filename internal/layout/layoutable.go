package layout

import "github.com/gogpu/ui/geometry"

// Layoutable represents an element that can be laid out.
//
// This interface abstracts over widgets and layout containers,
// allowing layout algorithms (Flex, Stack, Grid) to work with
// any layoutable element.
type Layoutable interface {
	// Layout calculates size given constraints and returns the computed size.
	Layout(constraints geometry.Constraints) geometry.Size

	// Children returns child layoutables for traversal.
	Children() []Layoutable

	// ID returns a unique identifier for caching purposes.
	ID() uint64
}
