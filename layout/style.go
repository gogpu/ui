package layout

import (
	"github.com/gogpu/ui/geometry"
)

// Shared string constants for alignment names.
const (
	styleStrStart        = "Start"
	styleStrEnd          = "End"
	styleStrCenter       = "Center"
	styleStrStretch      = "Stretch"
	styleStrUnknown      = "Unknown"
	styleStrSpaceBetween = "SpaceBetween"
	styleStrSpaceAround  = "SpaceAround"
	styleStrSpaceEvenly  = "SpaceEvenly"
	styleStrBaseline     = "Baseline"
)

// Display specifies the display mode for a node.
type Display int

const (
	// DisplayFlex uses flexbox layout.
	DisplayFlex Display = iota
	// DisplayGrid uses grid layout.
	DisplayGrid
	// DisplayBlock uses block layout.
	DisplayBlock
	// DisplayNone hides the node.
	DisplayNone
)

// String returns a string representation of the display mode.
func (d Display) String() string {
	switch d {
	case DisplayFlex:
		return "Flex"
	case DisplayGrid:
		return "Grid"
	case DisplayBlock:
		return "Block"
	case DisplayNone:
		return "None"
	default:
		return styleStrUnknown
	}
}

// FlexDirection specifies the main axis direction for flex layout.
type FlexDirection int

const (
	// FlexRow arranges children horizontally from left to right.
	FlexRow FlexDirection = iota
	// FlexRowReverse arranges children horizontally from right to left.
	FlexRowReverse
	// FlexColumn arranges children vertically from top to bottom.
	FlexColumn
	// FlexColumnReverse arranges children vertically from bottom to top.
	FlexColumnReverse
)

// String returns a string representation of the flex direction.
func (d FlexDirection) String() string {
	switch d {
	case FlexRow:
		return "Row"
	case FlexRowReverse:
		return "RowReverse"
	case FlexColumn:
		return "Column"
	case FlexColumnReverse:
		return "ColumnReverse"
	default:
		return styleStrUnknown
	}
}

// IsHorizontal returns true if the direction is horizontal.
func (d FlexDirection) IsHorizontal() bool {
	return d == FlexRow || d == FlexRowReverse
}

// IsReversed returns true if the direction is reversed.
func (d FlexDirection) IsReversed() bool {
	return d == FlexRowReverse || d == FlexColumnReverse
}

// FlexWrap specifies whether flex items wrap to multiple lines.
type FlexWrap int

const (
	// FlexNoWrap keeps all items on a single line.
	FlexNoWrap FlexWrap = iota
	// FlexWrapOn allows items to wrap to multiple lines.
	FlexWrapOn
	// FlexWrapReverse wraps items in reverse order.
	FlexWrapReverse
)

// String returns a string representation of the wrap mode.
func (w FlexWrap) String() string {
	switch w {
	case FlexNoWrap:
		return "NoWrap"
	case FlexWrapOn:
		return "Wrap"
	case FlexWrapReverse:
		return "WrapReverse"
	default:
		return styleStrUnknown
	}
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
		return styleStrStart
	case JustifyEnd:
		return styleStrEnd
	case JustifyCenter:
		return styleStrCenter
	case JustifySpaceBetween:
		return styleStrSpaceBetween
	case JustifySpaceAround:
		return styleStrSpaceAround
	case JustifySpaceEvenly:
		return styleStrSpaceEvenly
	default:
		return styleStrUnknown
	}
}

// AlignItems specifies how to align children along the cross axis.
type AlignItems int

const (
	// AlignItemsStart aligns children to the start of the cross axis.
	AlignItemsStart AlignItems = iota
	// AlignItemsEnd aligns children to the end of the cross axis.
	AlignItemsEnd
	// AlignItemsCenter centers children along the cross axis.
	AlignItemsCenter
	// AlignItemsStretch stretches children to fill the cross axis.
	AlignItemsStretch
	// AlignItemsBaseline aligns children by their baselines.
	AlignItemsBaseline
)

// String returns a string representation of align items.
func (a AlignItems) String() string {
	switch a {
	case AlignItemsStart:
		return styleStrStart
	case AlignItemsEnd:
		return styleStrEnd
	case AlignItemsCenter:
		return styleStrCenter
	case AlignItemsStretch:
		return styleStrStretch
	case AlignItemsBaseline:
		return styleStrBaseline
	default:
		return styleStrUnknown
	}
}

// AlignContent specifies how to distribute space between wrapped lines.
type AlignContent int

const (
	// AlignContentStart packs lines at the start.
	AlignContentStart AlignContent = iota
	// AlignContentEnd packs lines at the end.
	AlignContentEnd
	// AlignContentCenter centers lines.
	AlignContentCenter
	// AlignContentStretch stretches lines to fill.
	AlignContentStretch
	// AlignContentSpaceBetween distributes space between lines.
	AlignContentSpaceBetween
	// AlignContentSpaceAround distributes space around lines.
	AlignContentSpaceAround
)

// String returns a string representation of align content.
func (a AlignContent) String() string {
	switch a {
	case AlignContentStart:
		return styleStrStart
	case AlignContentEnd:
		return styleStrEnd
	case AlignContentCenter:
		return styleStrCenter
	case AlignContentStretch:
		return styleStrStretch
	case AlignContentSpaceBetween:
		return styleStrSpaceBetween
	case AlignContentSpaceAround:
		return styleStrSpaceAround
	default:
		return styleStrUnknown
	}
}

// DimensionUnit specifies how a dimension value is interpreted.
type DimensionUnit int

const (
	// DimensionAuto uses automatic sizing based on content.
	DimensionAuto DimensionUnit = iota
	// DimensionPixels uses an absolute pixel value.
	DimensionPixels
	// DimensionPercent uses a percentage of the parent's size.
	DimensionPercent
)

// Dimension represents a size value that can be auto, pixels, or percent.
type Dimension struct {
	Value float32
	Unit  DimensionUnit
}

// Auto returns an auto-sized dimension.
func Auto() Dimension {
	return Dimension{Unit: DimensionAuto}
}

// Px returns a pixel dimension.
func Px(value float32) Dimension {
	return Dimension{Value: value, Unit: DimensionPixels}
}

// Pct returns a percentage dimension.
func Pct(value float32) Dimension {
	return Dimension{Value: value, Unit: DimensionPercent}
}

// IsAuto returns true if the dimension is auto.
func (d Dimension) IsAuto() bool {
	return d.Unit == DimensionAuto
}

// Resolve resolves the dimension to a pixel value given a reference size.
// For auto dimensions, returns the fallback value.
func (d Dimension) Resolve(reference, fallback float32) float32 {
	switch d.Unit {
	case DimensionAuto:
		return fallback
	case DimensionPixels:
		return d.Value
	case DimensionPercent:
		return reference * d.Value / 100
	default:
		return fallback
	}
}

// Style defines layout properties for a node (CSS-like).
//
// Style contains all the properties that layout algorithms use to
// determine how a node should be sized and positioned.
type Style struct {
	// Display mode
	Display Display

	// Flexbox properties
	FlexDirection  FlexDirection
	FlexWrap       FlexWrap
	JustifyContent JustifyContent
	AlignItems     AlignItems
	AlignContent   AlignContent

	// Flex item properties
	FlexGrow   float32
	FlexShrink float32
	FlexBasis  Dimension

	// Sizing
	Width     Dimension
	Height    Dimension
	MinWidth  Dimension
	MinHeight Dimension
	MaxWidth  Dimension
	MaxHeight Dimension

	// Spacing
	Margin  geometry.Insets
	Padding geometry.Insets
	Gap     float32

	// Grid properties
	GridGap       float32
	GridRowGap    float32
	GridColumnGap float32
}

// DefaultStyle returns a Style with sensible defaults.
//
// Default values:
//   - Display: DisplayFlex
//   - FlexDirection: FlexRow
//   - FlexShrink: 1 (items can shrink)
//   - All dimensions: Auto
func DefaultStyle() Style {
	return Style{
		Display:        DisplayFlex,
		FlexDirection:  FlexRow,
		FlexWrap:       FlexNoWrap,
		JustifyContent: JustifyStart,
		AlignItems:     AlignItemsStretch,
		AlignContent:   AlignContentStretch,
		FlexGrow:       0,
		FlexShrink:     1,
		FlexBasis:      Auto(),
		Width:          Auto(),
		Height:         Auto(),
		MinWidth:       Auto(),
		MinHeight:      Auto(),
		MaxWidth:       Auto(),
		MaxHeight:      Auto(),
	}
}

// WithDisplay returns a copy of the style with the given display mode.
func (s Style) WithDisplay(display Display) Style {
	s.Display = display
	return s
}

// WithFlexDirection returns a copy of the style with the given flex direction.
func (s Style) WithFlexDirection(direction FlexDirection) Style {
	s.FlexDirection = direction
	return s
}

// WithJustifyContent returns a copy of the style with the given justify content.
func (s Style) WithJustifyContent(justify JustifyContent) Style {
	s.JustifyContent = justify
	return s
}

// WithAlignItems returns a copy of the style with the given align items.
func (s Style) WithAlignItems(align AlignItems) Style {
	s.AlignItems = align
	return s
}

// WithFlex returns a copy of the style with the given flex properties.
func (s Style) WithFlex(grow, shrink float32, basis Dimension) Style {
	s.FlexGrow = grow
	s.FlexShrink = shrink
	s.FlexBasis = basis
	return s
}

// WithSize returns a copy of the style with the given width and height.
func (s Style) WithSize(width, height Dimension) Style {
	s.Width = width
	s.Height = height
	return s
}

// WithMargin returns a copy of the style with the given margin.
func (s Style) WithMargin(margin geometry.Insets) Style {
	s.Margin = margin
	return s
}

// WithPadding returns a copy of the style with the given padding.
func (s Style) WithPadding(padding geometry.Insets) Style {
	s.Padding = padding
	return s
}

// WithGap returns a copy of the style with the given gap.
func (s Style) WithGap(gap float32) Style {
	s.Gap = gap
	return s
}
