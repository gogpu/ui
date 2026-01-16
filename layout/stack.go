package layout

import (
	"github.com/gogpu/ui/geometry"
)

// Common alignment string constants.
const (
	alignStrStart   = "Start"
	alignStrCenter  = "Center"
	alignStrEnd     = "End"
	alignStrStretch = "Stretch"
	alignStrUnknown = "Unknown"
)

// StackDirection specifies the direction for stack layouts.
type StackDirection int

const (
	// StackVertical stacks children vertically (VStack).
	StackVertical StackDirection = iota
	// StackHorizontal stacks children horizontally (HStack).
	StackHorizontal
	// StackZ overlays children on top of each other (ZStack).
	StackZ
)

// StackAlignment specifies how children are aligned within a stack.
type StackAlignment int

const (
	// StackAlignStart aligns children to the start (top/left).
	StackAlignStart StackAlignment = iota
	// StackAlignCenter centers children.
	StackAlignCenter
	// StackAlignEnd aligns children to the end (bottom/right).
	StackAlignEnd
	// StackAlignStretch stretches children to fill available space.
	StackAlignStretch
)

// String returns a string representation of stack alignment.
func (a StackAlignment) String() string {
	switch a {
	case StackAlignStart:
		return alignStrStart
	case StackAlignCenter:
		return alignStrCenter
	case StackAlignEnd:
		return alignStrEnd
	case StackAlignStretch:
		return alignStrStretch
	default:
		return alignStrUnknown
	}
}

// StackLayout implements VStack, HStack, and ZStack layouts.
//
// StackLayout arranges children in a single direction with spacing,
// or overlays them on top of each other for ZStack.
type StackLayout struct {
	// Direction specifies the stack direction.
	Direction StackDirection

	// Alignment specifies cross-axis alignment (or position for ZStack).
	Alignment StackAlignment

	// Spacing is the gap between children (not used for ZStack).
	Spacing float32
}

// Name returns the algorithm name based on direction.
func (s *StackLayout) Name() string {
	switch s.Direction {
	case StackVertical:
		return "vstack"
	case StackHorizontal:
		return "hstack"
	case StackZ:
		return "zstack"
	default:
		return "stack"
	}
}

// Compute performs stack layout on the tree starting at root.
func (s *StackLayout) Compute(tree LayoutTree, root NodeID, available geometry.Size) Result {
	switch s.Direction {
	case StackVertical:
		return s.computeVStack(tree, root, available)
	case StackHorizontal:
		return s.computeHStack(tree, root, available)
	case StackZ:
		return s.computeZStack(tree, root, available)
	default:
		return Result{}
	}
}

func (s *StackLayout) computeVStack(tree LayoutTree, root NodeID, available geometry.Size) Result {
	childCount := tree.ChildCount(root)
	if childCount == 0 {
		return Result{}
	}

	// Calculate total spacing
	numGaps := childCount - 1
	if numGaps < 0 {
		numGaps = 0
	}
	totalSpacing := s.Spacing * float32(numGaps)

	// Phase 1: Measure all children
	children := make([]stackItem, childCount)
	var totalHeight float32
	var maxWidth float32

	for i := 0; i < childCount; i++ {
		childID := tree.ChildAt(root, i)

		// Create constraints
		var constraints geometry.Constraints
		if s.Alignment == StackAlignStretch {
			constraints = geometry.BoxConstraints(
				available.Width, available.Width,
				0, geometry.Infinity,
			)
		} else {
			constraints = geometry.BoxConstraints(
				0, available.Width,
				0, geometry.Infinity,
			)
		}

		// Measure
		childSize := tree.Measure(childID, constraints)
		children[i] = stackItem{id: childID, size: childSize}

		totalHeight += childSize.Height
		if childSize.Width > maxWidth {
			maxWidth = childSize.Width
		}
	}

	totalHeight += totalSpacing

	// Phase 2: Position children
	currentY := float32(0)
	for i := range children {
		child := &children[i]

		// Calculate X based on alignment
		var x float32
		switch s.Alignment {
		case StackAlignStart:
			x = 0
		case StackAlignCenter:
			x = (maxWidth - child.size.Width) / 2
		case StackAlignEnd:
			x = maxWidth - child.size.Width
		case StackAlignStretch:
			x = 0
			child.size.Width = maxWidth
		}

		tree.SetLayout(child.id, NodeLayout{
			Position: geometry.Point{X: x, Y: currentY},
			Size:     child.size,
		})

		currentY += child.size.Height + s.Spacing
	}

	return Result{
		Size:     geometry.Size{Width: maxWidth, Height: totalHeight},
		Overflow: totalHeight > available.Height,
	}
}

func (s *StackLayout) computeHStack(tree LayoutTree, root NodeID, available geometry.Size) Result {
	childCount := tree.ChildCount(root)
	if childCount == 0 {
		return Result{}
	}

	// Calculate total spacing
	numGaps := childCount - 1
	if numGaps < 0 {
		numGaps = 0
	}
	totalSpacing := s.Spacing * float32(numGaps)

	// Phase 1: Measure all children
	children := make([]stackItem, childCount)
	var totalWidth float32
	var maxHeight float32

	for i := 0; i < childCount; i++ {
		childID := tree.ChildAt(root, i)

		// Create constraints
		var constraints geometry.Constraints
		if s.Alignment == StackAlignStretch {
			constraints = geometry.BoxConstraints(
				0, geometry.Infinity,
				available.Height, available.Height,
			)
		} else {
			constraints = geometry.BoxConstraints(
				0, geometry.Infinity,
				0, available.Height,
			)
		}

		// Measure
		childSize := tree.Measure(childID, constraints)
		children[i] = stackItem{id: childID, size: childSize}

		totalWidth += childSize.Width
		if childSize.Height > maxHeight {
			maxHeight = childSize.Height
		}
	}

	totalWidth += totalSpacing

	// Phase 2: Position children
	currentX := float32(0)
	for i := range children {
		child := &children[i]

		// Calculate Y based on alignment
		var y float32
		switch s.Alignment {
		case StackAlignStart:
			y = 0
		case StackAlignCenter:
			y = (maxHeight - child.size.Height) / 2
		case StackAlignEnd:
			y = maxHeight - child.size.Height
		case StackAlignStretch:
			y = 0
			child.size.Height = maxHeight
		}

		tree.SetLayout(child.id, NodeLayout{
			Position: geometry.Point{X: currentX, Y: y},
			Size:     child.size,
		})

		currentX += child.size.Width + s.Spacing
	}

	return Result{
		Size:     geometry.Size{Width: totalWidth, Height: maxHeight},
		Overflow: totalWidth > available.Width,
	}
}

func (s *StackLayout) computeZStack(tree LayoutTree, root NodeID, available geometry.Size) Result {
	return computeZStackCommon(tree, root, available, s.alignZStackChild)
}

func (s *StackLayout) alignZStackChild(childSize geometry.Size, stackWidth, stackHeight float32) geometry.Point {
	// ZStack uses Alignment for both X and Y centering
	var x, y float32

	switch s.Alignment {
	case StackAlignStart:
		x = 0
		y = 0
	case StackAlignCenter:
		x = (stackWidth - childSize.Width) / 2
		y = (stackHeight - childSize.Height) / 2
	case StackAlignEnd:
		x = stackWidth - childSize.Width
		y = stackHeight - childSize.Height
	case StackAlignStretch:
		x = 0
		y = 0
	}

	return geometry.Point{X: x, Y: y}
}

// stackItem stores computed values during layout.
type stackItem struct {
	id   NodeID
	size geometry.Size
}

// ZStackAlignment specifies position alignment for ZStack.
type ZStackAlignment int

const (
	// ZAlignTopLeft positions children at the top-left.
	ZAlignTopLeft ZStackAlignment = iota
	// ZAlignTop positions children at the top-center.
	ZAlignTop
	// ZAlignTopRight positions children at the top-right.
	ZAlignTopRight
	// ZAlignLeft positions children at the middle-left.
	ZAlignLeft
	// ZAlignCenter positions children at the center.
	ZAlignCenter
	// ZAlignRight positions children at the middle-right.
	ZAlignRight
	// ZAlignBottomLeft positions children at the bottom-left.
	ZAlignBottomLeft
	// ZAlignBottom positions children at the bottom-center.
	ZAlignBottom
	// ZAlignBottomRight positions children at the bottom-right.
	ZAlignBottomRight
)

// ZStackLayout implements overlay stack with 9-position alignment.
type ZStackLayout struct {
	Alignment ZStackAlignment
}

// Name returns "zstack".
func (z *ZStackLayout) Name() string {
	return "zstack"
}

// alignFunc is a function that calculates position for a child in a z-stack.
type alignFunc func(childSize geometry.Size, stackWidth, stackHeight float32) geometry.Point

// computeZStackCommon is a shared implementation for z-stack layout.
func computeZStackCommon(tree LayoutTree, root NodeID, available geometry.Size, align alignFunc) Result {
	childCount := tree.ChildCount(root)
	if childCount == 0 {
		return Result{}
	}

	// Phase 1: Measure all children
	children := make([]stackItem, childCount)
	var maxWidth, maxHeight float32

	for i := 0; i < childCount; i++ {
		childID := tree.ChildAt(root, i)
		constraints := geometry.Loose(available)
		childSize := tree.Measure(childID, constraints)
		children[i] = stackItem{id: childID, size: childSize}

		if childSize.Width > maxWidth {
			maxWidth = childSize.Width
		}
		if childSize.Height > maxHeight {
			maxHeight = childSize.Height
		}
	}

	// Phase 2: Position children
	for i := range children {
		child := &children[i]
		pos := align(child.size, maxWidth, maxHeight)
		tree.SetLayout(child.id, NodeLayout{
			Position: pos,
			Size:     child.size,
		})
	}

	return Result{
		Size: geometry.Size{Width: maxWidth, Height: maxHeight},
	}
}

// Compute performs ZStack layout.
func (z *ZStackLayout) Compute(tree LayoutTree, root NodeID, available geometry.Size) Result {
	return computeZStackCommon(tree, root, available, z.alignChild)
}

func (z *ZStackLayout) alignChild(childSize geometry.Size, stackWidth, stackHeight float32) geometry.Point {
	var x, y float32

	// Horizontal alignment
	switch z.Alignment {
	case ZAlignTopLeft, ZAlignLeft, ZAlignBottomLeft:
		x = 0
	case ZAlignTop, ZAlignCenter, ZAlignBottom:
		x = (stackWidth - childSize.Width) / 2
	case ZAlignTopRight, ZAlignRight, ZAlignBottomRight:
		x = stackWidth - childSize.Width
	}

	// Vertical alignment
	switch z.Alignment {
	case ZAlignTopLeft, ZAlignTop, ZAlignTopRight:
		y = 0
	case ZAlignLeft, ZAlignCenter, ZAlignRight:
		y = (stackHeight - childSize.Height) / 2
	case ZAlignBottomLeft, ZAlignBottom, ZAlignBottomRight:
		y = stackHeight - childSize.Height
	}

	return geometry.Point{X: x, Y: y}
}

func init() {
	Register(&StackLayout{Direction: StackVertical, Alignment: StackAlignStart})
	RegisterWithName("hstack", &StackLayout{Direction: StackHorizontal, Alignment: StackAlignStart})
	RegisterWithName("zstack", &ZStackLayout{Alignment: ZAlignCenter})
}
