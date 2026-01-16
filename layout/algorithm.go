package layout

import (
	"github.com/gogpu/ui/geometry"
)

// LayoutAlgorithm computes layout for a tree of nodes.
//
// Implementations of this interface define how child nodes are positioned
// and sized within their parent. The algorithm receives access to the node
// tree through the [LayoutTree] interface and must set layouts for all
// nodes it processes.
//
// # Implementation Guidelines
//
// A layout algorithm should:
//  1. Traverse children via [LayoutTree.ChildCount] and [LayoutTree.ChildAt]
//  2. Get style properties via [LayoutTree.Style]
//  3. Measure leaf nodes via [LayoutTree.Measure]
//  4. Compute positions and sizes for each child
//  5. Set computed layouts via [LayoutTree.SetLayout]
//  6. Return the total size in [Result]
//
// # Example Implementation
//
//	type SimpleRowLayout struct{}
//
//	func (s *SimpleRowLayout) Name() string { return "simple-row" }
//
//	func (s *SimpleRowLayout) Compute(tree LayoutTree, root NodeID, available geometry.Size) Result {
//	    var x float32
//	    var maxHeight float32
//
//	    for i := 0; i < tree.ChildCount(root); i++ {
//	        child := tree.ChildAt(root, i)
//	        constraints := geometry.Constraints{MaxWidth: available.Width - x, MaxHeight: available.Height}
//	        childSize := tree.Measure(child, constraints)
//
//	        tree.SetLayout(child, NodeLayout{
//	            Position: geometry.Point{X: x, Y: 0},
//	            Size:     childSize,
//	        })
//
//	        x += childSize.Width
//	        if childSize.Height > maxHeight {
//	            maxHeight = childSize.Height
//	        }
//	    }
//
//	    return Result{Size: geometry.Size{Width: x, Height: maxHeight}}
//	}
type LayoutAlgorithm interface {
	// Name returns the algorithm identifier.
	// This name is used to register and look up the algorithm in the registry.
	// Names should be lowercase and may use hyphens (e.g., "flex", "simple-row").
	Name() string

	// Compute calculates layout for the given tree starting at root.
	//
	// Parameters:
	//   - tree: Interface for accessing and modifying the node tree
	//   - root: The root node to start layout from
	//   - available: The available space for the root node
	//
	// Returns:
	//   - Result containing the computed size and overflow status
	//
	// The algorithm must call tree.SetLayout() for each node it processes.
	Compute(tree LayoutTree, root NodeID, available geometry.Size) Result
}

// LayoutFunc is a convenience type for creating algorithms from functions.
//
// This allows creating simple layout algorithms without defining a new type:
//
//	layout.Register("custom", layout.LayoutFunc{
//	    NameValue: "custom",
//	    ComputeFunc: func(tree layout.LayoutTree, root layout.NodeID, available geometry.Size) layout.Result {
//	        // Layout logic here
//	    },
//	})
type LayoutFunc struct {
	// NameValue is the algorithm name.
	NameValue string

	// ComputeFunc is the layout computation function.
	ComputeFunc func(tree LayoutTree, root NodeID, available geometry.Size) Result
}

// Name returns the algorithm name.
func (f LayoutFunc) Name() string {
	return f.NameValue
}

// Compute delegates to ComputeFunc.
func (f LayoutFunc) Compute(tree LayoutTree, root NodeID, available geometry.Size) Result {
	if f.ComputeFunc == nil {
		return Result{}
	}
	return f.ComputeFunc(tree, root, available)
}
