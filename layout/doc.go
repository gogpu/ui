// Package layout provides a public, extensible layout system for gogpu/ui.
//
// This package exposes layout algorithms that third-party developers can use
// and extend to create custom layouts. It provides the [LayoutAlgorithm] interface
// for pluggable layout computation and a [Registry] for registering custom layouts.
//
// # Architecture
//
// The layout system is built around several key concepts:
//
//   - [LayoutAlgorithm]: Interface for pluggable layout computation
//   - [LayoutTree]: Interface providing access to the node tree for algorithms
//   - [Registry]: Global registry for layout algorithms by name
//   - [Style]: CSS-like layout properties for nodes
//
// # Built-in Layouts
//
// The package provides built-in layout algorithms:
//
//   - "flex": CSS Flexbox-style layout ([FlexLayout])
//   - "vstack": Vertical stack layout
//   - "hstack": Horizontal stack layout
//   - "zstack": Overlay stack layout
//   - "grid": CSS Grid-style layout ([GridLayout])
//
// These are automatically registered via init() functions.
//
// # Custom Layouts
//
// Third-party developers can create custom layouts by implementing [LayoutAlgorithm]:
//
//	package masonry
//
//	import "github.com/gogpu/ui/layout"
//
//	func init() {
//	    layout.Register("masonry", &MasonryLayout{})
//	}
//
//	type MasonryLayout struct {
//	    Columns int
//	}
//
//	func (m *MasonryLayout) Name() string { return "masonry" }
//
//	func (m *MasonryLayout) Compute(tree layout.LayoutTree, root layout.NodeID, available geometry.Size) layout.Result {
//	    // Custom masonry algorithm implementation
//	    // ...
//	}
//
// # Thread Safety
//
// The [Registry] is thread-safe for concurrent registration and lookup.
// Individual layout algorithms may have their own thread-safety requirements.
//
// # Relationship to internal/layout
//
// This package provides the public API for layout extensibility.
// The internal/layout package contains the actual layout implementations
// (FlexContainer, VStack, HStack, ZStack, GridContainer, Engine) that are
// used by the UI framework internally.
//
// The algorithms in this package wrap the internal implementations and
// expose them through the [LayoutAlgorithm] interface.
package layout
