package layout

import (
	"github.com/gogpu/ui/geometry"
)

// String constants for alignment names.
const (
	alignStart   = "Start"
	alignEnd     = "End"
	alignCenter  = "Center"
	alignUnknown = "Unknown"
)

// Direction specifies the main axis direction for flex layout.
type Direction int

const (
	// Row arranges children horizontally from left to right.
	Row Direction = iota
	// RowReverse arranges children horizontally from right to left.
	RowReverse
	// Column arranges children vertically from top to bottom.
	Column
	// ColumnReverse arranges children vertically from bottom to top.
	ColumnReverse
)

// String returns a string representation of the direction.
func (d Direction) String() string {
	switch d {
	case Row:
		return "Row"
	case RowReverse:
		return "RowReverse"
	case Column:
		return "Column"
	case ColumnReverse:
		return "ColumnReverse"
	default:
		return alignUnknown
	}
}

// IsHorizontal returns true if the direction is horizontal (Row or RowReverse).
func (d Direction) IsHorizontal() bool {
	return d == Row || d == RowReverse
}

// IsReversed returns true if the direction is reversed.
func (d Direction) IsReversed() bool {
	return d == RowReverse || d == ColumnReverse
}

// JustifyContent specifies how to distribute space along the main axis.
type JustifyContent int

const (
	// JustifyStart packs children at the start of the main axis.
	JustifyStart JustifyContent = iota
	// JustifyEnd packs children at the end of the main axis.
	JustifyEnd
	// JustifyCenter centers children along the main axis.
	JustifyCenter
	// JustifySpaceBetween distributes space between children (no space at edges).
	JustifySpaceBetween
	// JustifySpaceAround distributes space around children (half space at edges).
	JustifySpaceAround
	// JustifySpaceEvenly distributes space evenly (equal space everywhere).
	JustifySpaceEvenly
)

// String returns a string representation of justify content.
func (j JustifyContent) String() string {
	switch j {
	case JustifyStart:
		return alignStart
	case JustifyEnd:
		return alignEnd
	case JustifyCenter:
		return alignCenter
	case JustifySpaceBetween:
		return "SpaceBetween"
	case JustifySpaceAround:
		return "SpaceAround"
	case JustifySpaceEvenly:
		return "SpaceEvenly"
	default:
		return alignUnknown
	}
}

// AlignItems specifies how to align children along the cross axis.
type AlignItems int

const (
	// AlignStart aligns children to the start of the cross axis.
	AlignStart AlignItems = iota
	// AlignEnd aligns children to the end of the cross axis.
	AlignEnd
	// AlignCenter centers children along the cross axis.
	AlignCenter
	// AlignStretch stretches children to fill the cross axis.
	AlignStretch
	// AlignBaseline aligns children by their baselines (text alignment).
	AlignBaseline
)

// String returns a string representation of align items.
func (a AlignItems) String() string {
	switch a {
	case AlignStart:
		return alignStart
	case AlignEnd:
		return alignEnd
	case AlignCenter:
		return alignCenter
	case AlignStretch:
		return "Stretch"
	case AlignBaseline:
		return "Baseline"
	default:
		return alignUnknown
	}
}

// WrapMode specifies whether flex items wrap to multiple lines.
type WrapMode int

const (
	// NoWrap keeps all items on a single line.
	NoWrap WrapMode = iota
	// Wrap allows items to wrap to multiple lines.
	Wrap
	// WrapReverse wraps items in reverse order.
	WrapReverse
)

// String returns a string representation of wrap mode.
func (w WrapMode) String() string {
	switch w {
	case NoWrap:
		return "NoWrap"
	case Wrap:
		return "Wrap"
	case WrapReverse:
		return "WrapReverse"
	default:
		return alignUnknown
	}
}

// FlexItem configures how a child participates in flex layout.
type FlexItem struct {
	// Element is the layoutable child element.
	Element Layoutable

	// Grow is the flex-grow factor. Positive values allow the item
	// to grow to fill available space proportionally.
	// Default: 0 (no growing)
	Grow float32

	// Shrink is the flex-shrink factor. Positive values allow the item
	// to shrink when space is insufficient.
	// Default: 1 (can shrink)
	Shrink float32

	// Basis is the initial main axis size before grow/shrink.
	// Use 0 for auto (use element's preferred size).
	// Use negative values for content-based sizing.
	Basis float32

	// AlignSelf overrides the container's AlignItems for this item.
	// Use AlignSelfAuto to use the container's AlignItems.
	AlignSelf AlignSelf

	// computed values during layout
	mainSize    float32
	crossSize   float32
	mainOffset  float32
	crossOffset float32
}

// AlignSelf allows an item to override the container's AlignItems.
type AlignSelf int

const (
	// AlignSelfAuto uses the container's AlignItems value.
	AlignSelfAuto AlignSelf = iota
	// AlignSelfStart aligns this item to the start.
	AlignSelfStart
	// AlignSelfEnd aligns this item to the end.
	AlignSelfEnd
	// AlignSelfCenter centers this item.
	AlignSelfCenter
	// AlignSelfStretch stretches this item.
	AlignSelfStretch
)

// FlexContainer implements CSS Flexbox-style layout.
//
// FlexContainer arranges children along a main axis (row or column),
// distributing space according to flex properties.
type FlexContainer struct {
	id uint64

	// Direction specifies the main axis direction.
	Direction Direction

	// JustifyContent specifies main axis space distribution.
	JustifyContent JustifyContent

	// AlignItems specifies cross axis alignment.
	AlignItems AlignItems

	// Wrap specifies whether items wrap to multiple lines.
	Wrap WrapMode

	// Gap is the space between items (main axis).
	Gap float32

	// CrossGap is the space between lines when wrapping (cross axis).
	CrossGap float32

	// Items are the flex children with their flex properties.
	Items []FlexItem

	// computed size after layout
	size geometry.Size
}

// NewFlexContainer creates a new flex container with the given properties.
func NewFlexContainer(direction Direction, justify JustifyContent, align AlignItems) *FlexContainer {
	return &FlexContainer{
		Direction:      direction,
		JustifyContent: justify,
		AlignItems:     align,
		Wrap:           NoWrap,
		Items:          make([]FlexItem, 0, 8),
	}
}

// SetID sets the unique identifier for caching.
func (f *FlexContainer) SetID(id uint64) {
	f.id = id
}

// ID returns the unique identifier.
func (f *FlexContainer) ID() uint64 {
	return f.id
}

// AddChild adds a child element with default flex properties.
func (f *FlexContainer) AddChild(element Layoutable) {
	f.Items = append(f.Items, FlexItem{
		Element: element,
		Grow:    0,
		Shrink:  1,
		Basis:   0, // auto
	})
}

// AddChildWithFlex adds a child element with custom flex properties.
func (f *FlexContainer) AddChildWithFlex(element Layoutable, grow, shrink, basis float32) {
	f.Items = append(f.Items, FlexItem{
		Element: element,
		Grow:    grow,
		Shrink:  shrink,
		Basis:   basis,
	})
}

// AddFlexItem adds a fully configured flex item.
func (f *FlexContainer) AddFlexItem(item FlexItem) {
	f.Items = append(f.Items, item)
}

// Clear removes all children.
func (f *FlexContainer) Clear() {
	f.Items = f.Items[:0]
}

// Children returns child layoutables.
func (f *FlexContainer) Children() []Layoutable {
	children := make([]Layoutable, len(f.Items))
	for i, item := range f.Items {
		children[i] = item.Element
	}
	return children
}

// Layout performs flex layout and returns the container size.
func (f *FlexContainer) Layout(constraints geometry.Constraints) geometry.Size {
	if len(f.Items) == 0 {
		f.size = constraints.Smallest()
		return f.size
	}

	isHorizontal := f.Direction.IsHorizontal()
	mainMax, crossMax := f.getAxisConstraints(constraints, isHorizontal)

	// Phase 1: Measure children
	totalBasis, totalGrow, totalShrink, maxCross := f.measureChildren(crossMax, isHorizontal)
	totalBasis += f.totalGapSpace()

	// Phase 2: Distribute space
	f.distributeSpace(mainMax-totalBasis, totalGrow, totalShrink)

	// Phase 3: Apply alignment and re-layout
	f.applyAlignmentAndRelayout(maxCross, isHorizontal)

	// Phase 4: Position children
	f.positionChildren(mainMax, maxCross, isHorizontal)

	// Calculate final size
	f.size = f.calculateFinalSize(constraints, maxCross, isHorizontal)
	return f.size
}

// getAxisConstraints extracts main and cross axis constraints.
func (f *FlexContainer) getAxisConstraints(constraints geometry.Constraints, isHorizontal bool) (mainMax, crossMax float32) {
	if isHorizontal {
		mainMax = constraints.MaxWidth
		crossMax = constraints.MaxHeight
	} else {
		mainMax = constraints.MaxHeight
		crossMax = constraints.MaxWidth
	}

	// Handle unbounded main axis
	if mainMax >= geometry.Infinity {
		if isHorizontal {
			mainMax = constraints.MinWidth
		} else {
			mainMax = constraints.MinHeight
		}
		if mainMax <= 0 {
			mainMax = 10000 // fallback for unbounded
		}
	}
	return mainMax, crossMax
}

// measureChildren measures all children and returns totals.
func (f *FlexContainer) measureChildren(crossMax float32, isHorizontal bool) (totalBasis, totalGrow, totalShrink, maxCross float32) {
	for i := range f.Items {
		item := &f.Items[i]
		if item.Element == nil {
			continue
		}

		childConstraints := f.createMeasureConstraints(crossMax, isHorizontal)
		childSize := item.Element.Layout(childConstraints)

		f.setItemSizes(item, childSize, isHorizontal)

		totalBasis += item.mainSize
		totalGrow += item.Grow
		totalShrink += item.Shrink * item.mainSize
		if item.crossSize > maxCross {
			maxCross = item.crossSize
		}
	}
	return
}

// createMeasureConstraints creates constraints for measuring a child.
func (f *FlexContainer) createMeasureConstraints(crossMax float32, isHorizontal bool) geometry.Constraints {
	if isHorizontal {
		return geometry.BoxConstraints(0, geometry.Infinity, 0, crossMax)
	}
	return geometry.BoxConstraints(0, crossMax, 0, geometry.Infinity)
}

// setItemSizes sets main and cross sizes for an item.
func (f *FlexContainer) setItemSizes(item *FlexItem, childSize geometry.Size, isHorizontal bool) {
	switch {
	case item.Basis > 0:
		item.mainSize = item.Basis
	case isHorizontal:
		item.mainSize = childSize.Width
	default:
		item.mainSize = childSize.Height
	}

	if isHorizontal {
		item.crossSize = childSize.Height
	} else {
		item.crossSize = childSize.Width
	}
}

// totalGapSpace returns total space used by gaps.
func (f *FlexContainer) totalGapSpace() float32 {
	numGaps := len(f.Items) - 1
	if numGaps < 0 {
		numGaps = 0
	}
	return f.Gap * float32(numGaps)
}

// distributeSpace distributes available space using grow/shrink.
func (f *FlexContainer) distributeSpace(availableSpace, totalGrow, totalShrink float32) {
	if availableSpace > 0 && totalGrow > 0 {
		f.growItems(availableSpace, totalGrow)
	} else if availableSpace < 0 && totalShrink > 0 {
		f.shrinkItems(-availableSpace, totalShrink)
	}
}

// growItems grows items proportionally.
func (f *FlexContainer) growItems(space, totalGrow float32) {
	for i := range f.Items {
		item := &f.Items[i]
		if item.Grow > 0 {
			item.mainSize += space * (item.Grow / totalGrow)
		}
	}
}

// shrinkItems shrinks items proportionally.
func (f *FlexContainer) shrinkItems(amount, totalShrink float32) {
	for i := range f.Items {
		item := &f.Items[i]
		if item.Shrink > 0 {
			scaledShrink := item.Shrink * item.mainSize
			item.mainSize -= amount * (scaledShrink / totalShrink)
			if item.mainSize < 0 {
				item.mainSize = 0
			}
		}
	}
}

// applyAlignmentAndRelayout applies cross-axis alignment and re-layouts children.
func (f *FlexContainer) applyAlignmentAndRelayout(maxCross float32, isHorizontal bool) {
	for i := range f.Items {
		item := &f.Items[i]
		if item.Element == nil {
			continue
		}

		align := f.getEffectiveAlignment(item)
		childConstraints := f.createFinalConstraints(item, align, maxCross, isHorizontal)
		finalSize := item.Element.Layout(childConstraints)

		if isHorizontal {
			item.mainSize = finalSize.Width
			item.crossSize = finalSize.Height
		} else {
			item.mainSize = finalSize.Height
			item.crossSize = finalSize.Width
		}
	}
}

// getEffectiveAlignment returns the effective alignment for an item.
func (f *FlexContainer) getEffectiveAlignment(item *FlexItem) AlignItems {
	switch item.AlignSelf {
	case AlignSelfStart:
		return AlignStart
	case AlignSelfEnd:
		return AlignEnd
	case AlignSelfCenter:
		return AlignCenter
	case AlignSelfStretch:
		return AlignStretch
	default:
		return f.AlignItems
	}
}

// createFinalConstraints creates final constraints for a child.
func (f *FlexContainer) createFinalConstraints(item *FlexItem, align AlignItems, maxCross float32, isHorizontal bool) geometry.Constraints {
	if isHorizontal {
		if align == AlignStretch {
			return geometry.BoxConstraints(item.mainSize, item.mainSize, maxCross, maxCross)
		}
		return geometry.BoxConstraints(item.mainSize, item.mainSize, 0, maxCross)
	}
	if align == AlignStretch {
		return geometry.BoxConstraints(maxCross, maxCross, item.mainSize, item.mainSize)
	}
	return geometry.BoxConstraints(0, maxCross, item.mainSize, item.mainSize)
}

// calculateFinalSize computes the final container size.
func (f *FlexContainer) calculateFinalSize(constraints geometry.Constraints, maxCross float32, isHorizontal bool) geometry.Size {
	var totalMain float32
	for i, item := range f.Items {
		totalMain += item.mainSize
		if i < len(f.Items)-1 {
			totalMain += f.Gap
		}
	}

	if isHorizontal {
		return constraints.Constrain(geometry.Size{Width: totalMain, Height: maxCross})
	}
	return constraints.Constrain(geometry.Size{Width: maxCross, Height: totalMain})
}

// positionChildren calculates positions for all children.
func (f *FlexContainer) positionChildren(mainMax, crossMax float32, _ bool) {
	totalMain := f.calculateTotalMainSize()
	mainStart, mainSpacing := f.calculateJustification(mainMax - totalMain)

	isReversed := f.Direction.IsReversed()
	if isReversed {
		mainStart = mainMax - mainStart
	}

	currentMain := mainStart
	for i := range f.Items {
		item := &f.Items[i]
		currentMain = f.positionItem(item, currentMain, mainSpacing, crossMax, isReversed)
	}
}

// calculateTotalMainSize computes total main axis size of all items.
func (f *FlexContainer) calculateTotalMainSize() float32 {
	var total float32
	for i, item := range f.Items {
		total += item.mainSize
		if i < len(f.Items)-1 {
			total += f.Gap
		}
	}
	return total
}

// calculateJustification returns start position and spacing for main axis.
func (f *FlexContainer) calculateJustification(freeSpace float32) (mainStart, mainSpacing float32) {
	switch f.JustifyContent {
	case JustifyStart:
		return 0, f.Gap
	case JustifyEnd:
		return freeSpace, f.Gap
	case JustifyCenter:
		return freeSpace / 2, f.Gap
	case JustifySpaceBetween:
		if len(f.Items) > 1 {
			return 0, freeSpace / float32(len(f.Items)-1)
		}
		return 0, 0
	case JustifySpaceAround:
		if len(f.Items) > 0 {
			space := freeSpace / float32(len(f.Items))
			return space / 2, space
		}
		return 0, 0
	case JustifySpaceEvenly:
		if len(f.Items) > 0 {
			spacing := freeSpace / float32(len(f.Items)+1)
			return spacing, spacing
		}
		return 0, 0
	default:
		return 0, f.Gap
	}
}

// positionItem positions a single item and returns the new current main position.
func (f *FlexContainer) positionItem(item *FlexItem, currentMain, mainSpacing, crossMax float32, isReversed bool) float32 {
	if isReversed {
		item.mainOffset = currentMain - item.mainSize
		currentMain -= item.mainSize + mainSpacing
	} else {
		item.mainOffset = currentMain
		currentMain += item.mainSize + mainSpacing
	}

	f.setCrossOffset(item, crossMax)
	return currentMain
}

// setCrossOffset sets the cross-axis offset for an item based on alignment.
func (f *FlexContainer) setCrossOffset(item *FlexItem, crossMax float32) {
	align := f.getEffectiveAlignment(item)
	switch align {
	case AlignStart, AlignBaseline:
		item.crossOffset = 0
	case AlignEnd:
		item.crossOffset = crossMax - item.crossSize
	case AlignCenter:
		item.crossOffset = (crossMax - item.crossSize) / 2
	case AlignStretch:
		item.crossOffset = 0
		item.crossSize = crossMax
	}
}

// Size returns the computed size after layout.
func (f *FlexContainer) Size() geometry.Size {
	return f.size
}

// ItemPosition returns the position of a child item after layout.
func (f *FlexContainer) ItemPosition(index int) geometry.Point {
	if index < 0 || index >= len(f.Items) {
		return geometry.Point{}
	}
	item := f.Items[index]
	if f.Direction.IsHorizontal() {
		return geometry.Point{X: item.mainOffset, Y: item.crossOffset}
	}
	return geometry.Point{X: item.crossOffset, Y: item.mainOffset}
}

// ItemSize returns the computed size of a child item after layout.
func (f *FlexContainer) ItemSize(index int) geometry.Size {
	if index < 0 || index >= len(f.Items) {
		return geometry.Size{}
	}
	item := f.Items[index]
	if f.Direction.IsHorizontal() {
		return geometry.Size{Width: item.mainSize, Height: item.crossSize}
	}
	return geometry.Size{Width: item.crossSize, Height: item.mainSize}
}

// ItemBounds returns the bounds of a child item after layout.
func (f *FlexContainer) ItemBounds(index int) geometry.Rect {
	pos := f.ItemPosition(index)
	size := f.ItemSize(index)
	return geometry.FromPointSize(pos, size)
}
