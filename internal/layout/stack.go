package layout

import (
	"github.com/gogpu/ui/geometry"
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
	// StackAlignStretch stretches children to fill the available space.
	StackAlignStretch
)

// String returns a string representation of stack alignment.
func (a StackAlignment) String() string {
	switch a {
	case StackAlignStart:
		return "Start"
	case StackAlignCenter:
		return "Center"
	case StackAlignEnd:
		return "End"
	case StackAlignStretch:
		return "Stretch"
	default:
		return "Unknown"
	}
}

// StackChild represents a child in a stack layout.
type StackChild struct {
	// Element is the layoutable child.
	Element Layoutable

	// computed values after layout
	position geometry.Point
	size     geometry.Size
}

// VStack arranges children vertically with spacing.
//
// VStack is a simplified layout that stacks children from top to bottom,
// with optional spacing and cross-axis (horizontal) alignment.
type VStack struct {
	id uint64

	// Spacing is the gap between children.
	Spacing float32

	// Alignment specifies horizontal alignment of children.
	Alignment StackAlignment

	// Children are the stack items.
	Children []StackChild

	// computed size after layout
	size geometry.Size
}

// NewVStack creates a new vertical stack.
func NewVStack(spacing float32, alignment StackAlignment) *VStack {
	return &VStack{
		Spacing:   spacing,
		Alignment: alignment,
		Children:  make([]StackChild, 0, 8),
	}
}

// SetID sets the unique identifier for caching.
func (v *VStack) SetID(id uint64) {
	v.id = id
}

// ID returns the unique identifier.
func (v *VStack) ID() uint64 {
	return v.id
}

// AddChild adds a child element.
func (v *VStack) AddChild(element Layoutable) {
	v.Children = append(v.Children, StackChild{Element: element})
}

// Clear removes all children.
func (v *VStack) Clear() {
	v.Children = v.Children[:0]
}

// ChildLayoutables returns child layoutables for the Layoutable interface.
func (v *VStack) ChildLayoutables() []Layoutable {
	children := make([]Layoutable, len(v.Children))
	for i, child := range v.Children {
		children[i] = child.Element
	}
	return children
}

// Layout performs vertical stack layout.
func (v *VStack) Layout(constraints geometry.Constraints) geometry.Size {
	if len(v.Children) == 0 {
		v.size = constraints.Smallest()
		return v.size
	}

	// Calculate total spacing
	numGaps := len(v.Children) - 1
	if numGaps < 0 {
		numGaps = 0
	}
	totalSpacing := v.Spacing * float32(numGaps)

	// Phase 1: Measure all children
	var totalHeight float32
	var maxWidth float32

	for i := range v.Children {
		child := &v.Children[i]
		if child.Element == nil {
			continue
		}

		// Create child constraints
		var childConstraints geometry.Constraints
		if v.Alignment == StackAlignStretch {
			// Stretch: tight width
			childConstraints = geometry.BoxConstraints(
				constraints.MaxWidth, constraints.MaxWidth,
				0, geometry.Infinity,
			)
		} else {
			// Non-stretch: flexible width
			childConstraints = geometry.BoxConstraints(
				0, constraints.MaxWidth,
				0, geometry.Infinity,
			)
		}

		// Measure child
		childSize := child.Element.Layout(childConstraints)
		child.size = childSize

		totalHeight += childSize.Height
		if childSize.Width > maxWidth {
			maxWidth = childSize.Width
		}
	}

	// Add spacing to total height
	totalHeight += totalSpacing

	// Constrain to available space
	maxWidth = constraints.ConstrainWidth(maxWidth)
	totalHeight = constraints.ConstrainHeight(totalHeight)

	// Phase 2: Position children
	currentY := float32(0)
	for i := range v.Children {
		child := &v.Children[i]
		if child.Element == nil {
			continue
		}

		// Calculate X position based on alignment
		var x float32
		switch v.Alignment {
		case StackAlignStart:
			x = 0
		case StackAlignCenter:
			x = (maxWidth - child.size.Width) / 2
		case StackAlignEnd:
			x = maxWidth - child.size.Width
		case StackAlignStretch:
			x = 0
			// Re-layout with stretched width if needed
			if child.size.Width < maxWidth {
				stretchConstraints := geometry.Tight(geometry.Size{
					Width:  maxWidth,
					Height: child.size.Height,
				})
				child.size = child.Element.Layout(stretchConstraints)
			}
		}

		child.position = geometry.Point{X: x, Y: currentY}
		currentY += child.size.Height + v.Spacing
	}

	v.size = geometry.Size{Width: maxWidth, Height: totalHeight}
	return v.size
}

// Size returns the computed size after layout.
func (v *VStack) Size() geometry.Size {
	return v.size
}

// ChildPosition returns the position of a child after layout.
func (v *VStack) ChildPosition(index int) geometry.Point {
	if index < 0 || index >= len(v.Children) {
		return geometry.Point{}
	}
	return v.Children[index].position
}

// ChildSize returns the computed size of a child after layout.
func (v *VStack) ChildSize(index int) geometry.Size {
	if index < 0 || index >= len(v.Children) {
		return geometry.Size{}
	}
	return v.Children[index].size
}

// ChildBounds returns the bounds of a child after layout.
func (v *VStack) ChildBounds(index int) geometry.Rect {
	pos := v.ChildPosition(index)
	size := v.ChildSize(index)
	return geometry.FromPointSize(pos, size)
}

// HStack arranges children horizontally with spacing.
//
// HStack is a simplified layout that stacks children from left to right,
// with optional spacing and cross-axis (vertical) alignment.
type HStack struct {
	id uint64

	// Spacing is the gap between children.
	Spacing float32

	// Alignment specifies vertical alignment of children.
	Alignment StackAlignment

	// Children are the stack items.
	Children []StackChild

	// computed size after layout
	size geometry.Size
}

// NewHStack creates a new horizontal stack.
func NewHStack(spacing float32, alignment StackAlignment) *HStack {
	return &HStack{
		Spacing:   spacing,
		Alignment: alignment,
		Children:  make([]StackChild, 0, 8),
	}
}

// SetID sets the unique identifier for caching.
func (h *HStack) SetID(id uint64) {
	h.id = id
}

// ID returns the unique identifier.
func (h *HStack) ID() uint64 {
	return h.id
}

// AddChild adds a child element.
func (h *HStack) AddChild(element Layoutable) {
	h.Children = append(h.Children, StackChild{Element: element})
}

// Clear removes all children.
func (h *HStack) Clear() {
	h.Children = h.Children[:0]
}

// ChildLayoutables returns child layoutables for the Layoutable interface.
func (h *HStack) ChildLayoutables() []Layoutable {
	children := make([]Layoutable, len(h.Children))
	for i, child := range h.Children {
		children[i] = child.Element
	}
	return children
}

// Layout performs horizontal stack layout.
func (h *HStack) Layout(constraints geometry.Constraints) geometry.Size {
	if len(h.Children) == 0 {
		h.size = constraints.Smallest()
		return h.size
	}

	// Calculate total spacing
	numGaps := len(h.Children) - 1
	if numGaps < 0 {
		numGaps = 0
	}
	totalSpacing := h.Spacing * float32(numGaps)

	// Phase 1: Measure all children
	var totalWidth float32
	var maxHeight float32

	for i := range h.Children {
		child := &h.Children[i]
		if child.Element == nil {
			continue
		}

		// Create child constraints
		var childConstraints geometry.Constraints
		if h.Alignment == StackAlignStretch {
			// Stretch: tight height
			childConstraints = geometry.BoxConstraints(
				0, geometry.Infinity,
				constraints.MaxHeight, constraints.MaxHeight,
			)
		} else {
			// Non-stretch: flexible height
			childConstraints = geometry.BoxConstraints(
				0, geometry.Infinity,
				0, constraints.MaxHeight,
			)
		}

		// Measure child
		childSize := child.Element.Layout(childConstraints)
		child.size = childSize

		totalWidth += childSize.Width
		if childSize.Height > maxHeight {
			maxHeight = childSize.Height
		}
	}

	// Add spacing to total width
	totalWidth += totalSpacing

	// Constrain to available space
	totalWidth = constraints.ConstrainWidth(totalWidth)
	maxHeight = constraints.ConstrainHeight(maxHeight)

	// Phase 2: Position children
	currentX := float32(0)
	for i := range h.Children {
		child := &h.Children[i]
		if child.Element == nil {
			continue
		}

		// Calculate Y position based on alignment
		var y float32
		switch h.Alignment {
		case StackAlignStart:
			y = 0
		case StackAlignCenter:
			y = (maxHeight - child.size.Height) / 2
		case StackAlignEnd:
			y = maxHeight - child.size.Height
		case StackAlignStretch:
			y = 0
			// Re-layout with stretched height if needed
			if child.size.Height < maxHeight {
				stretchConstraints := geometry.Tight(geometry.Size{
					Width:  child.size.Width,
					Height: maxHeight,
				})
				child.size = child.Element.Layout(stretchConstraints)
			}
		}

		child.position = geometry.Point{X: currentX, Y: y}
		currentX += child.size.Width + h.Spacing
	}

	h.size = geometry.Size{Width: totalWidth, Height: maxHeight}
	return h.size
}

// Size returns the computed size after layout.
func (h *HStack) Size() geometry.Size {
	return h.size
}

// ChildPosition returns the position of a child after layout.
func (h *HStack) ChildPosition(index int) geometry.Point {
	if index < 0 || index >= len(h.Children) {
		return geometry.Point{}
	}
	return h.Children[index].position
}

// ChildSize returns the computed size of a child after layout.
func (h *HStack) ChildSize(index int) geometry.Size {
	if index < 0 || index >= len(h.Children) {
		return geometry.Size{}
	}
	return h.Children[index].size
}

// ChildBounds returns the bounds of a child after layout.
func (h *HStack) ChildBounds(index int) geometry.Rect {
	pos := h.ChildPosition(index)
	size := h.ChildSize(index)
	return geometry.FromPointSize(pos, size)
}

// ZStackAlignment specifies how children are positioned in a ZStack.
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

// String returns a string representation of ZStack alignment.
func (a ZStackAlignment) String() string {
	switch a {
	case ZAlignTopLeft:
		return "TopLeft"
	case ZAlignTop:
		return "Top"
	case ZAlignTopRight:
		return "TopRight"
	case ZAlignLeft:
		return "Left"
	case ZAlignCenter:
		return "Center"
	case ZAlignRight:
		return "Right"
	case ZAlignBottomLeft:
		return "BottomLeft"
	case ZAlignBottom:
		return "Bottom"
	case ZAlignBottomRight:
		return "BottomRight"
	default:
		return "Unknown"
	}
}

// ZStack overlays children on top of each other.
//
// ZStack is useful for layering widgets, such as placing a badge
// on top of an icon, or creating overlays.
type ZStack struct {
	id uint64

	// Alignment specifies how children are positioned.
	Alignment ZStackAlignment

	// Children are the stack items (bottom to top order).
	Children []StackChild

	// computed size after layout
	size geometry.Size
}

// NewZStack creates a new overlay stack.
func NewZStack(alignment ZStackAlignment) *ZStack {
	return &ZStack{
		Alignment: alignment,
		Children:  make([]StackChild, 0, 4),
	}
}

// SetID sets the unique identifier for caching.
func (z *ZStack) SetID(id uint64) {
	z.id = id
}

// ID returns the unique identifier.
func (z *ZStack) ID() uint64 {
	return z.id
}

// AddChild adds a child element (will be rendered on top of previous children).
func (z *ZStack) AddChild(element Layoutable) {
	z.Children = append(z.Children, StackChild{Element: element})
}

// Clear removes all children.
func (z *ZStack) Clear() {
	z.Children = z.Children[:0]
}

// ChildLayoutables returns child layoutables for the Layoutable interface.
func (z *ZStack) ChildLayoutables() []Layoutable {
	children := make([]Layoutable, len(z.Children))
	for i, child := range z.Children {
		children[i] = child.Element
	}
	return children
}

// Layout performs overlay stack layout.
func (z *ZStack) Layout(constraints geometry.Constraints) geometry.Size {
	if len(z.Children) == 0 {
		z.size = constraints.Smallest()
		return z.size
	}

	// Phase 1: Measure all children to find max size
	var maxWidth, maxHeight float32

	for i := range z.Children {
		child := &z.Children[i]
		if child.Element == nil {
			continue
		}

		// Create loose constraints for measuring
		childConstraints := constraints.Loosen()

		// Measure child
		childSize := child.Element.Layout(childConstraints)
		child.size = childSize

		if childSize.Width > maxWidth {
			maxWidth = childSize.Width
		}
		if childSize.Height > maxHeight {
			maxHeight = childSize.Height
		}
	}

	// Constrain to available space
	stackWidth := constraints.ConstrainWidth(maxWidth)
	stackHeight := constraints.ConstrainHeight(maxHeight)

	// Phase 2: Position children according to alignment
	for i := range z.Children {
		child := &z.Children[i]
		if child.Element == nil {
			continue
		}

		child.position = z.alignChild(child.size, stackWidth, stackHeight)
	}

	z.size = geometry.Size{Width: stackWidth, Height: stackHeight}
	return z.size
}

// alignChild calculates position for a child based on alignment.
func (z *ZStack) alignChild(childSize geometry.Size, stackWidth, stackHeight float32) geometry.Point {
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

// Size returns the computed size after layout.
func (z *ZStack) Size() geometry.Size {
	return z.size
}

// ChildPosition returns the position of a child after layout.
func (z *ZStack) ChildPosition(index int) geometry.Point {
	if index < 0 || index >= len(z.Children) {
		return geometry.Point{}
	}
	return z.Children[index].position
}

// ChildSize returns the computed size of a child after layout.
func (z *ZStack) ChildSize(index int) geometry.Size {
	if index < 0 || index >= len(z.Children) {
		return geometry.Size{}
	}
	return z.Children[index].size
}

// ChildBounds returns the bounds of a child after layout.
func (z *ZStack) ChildBounds(index int) geometry.Rect {
	pos := z.ChildPosition(index)
	size := z.ChildSize(index)
	return geometry.FromPointSize(pos, size)
}
