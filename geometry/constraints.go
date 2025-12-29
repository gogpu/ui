package geometry

import (
	"fmt"
	"math"
)

// Infinity represents unbounded constraint dimension.
// Use this value for MaxWidth/MaxHeight when there is no upper limit.
const Infinity = float32(math.MaxFloat32)

// Constraints represents box constraints for layout, similar to Flutter's BoxConstraints.
//
// Constraints define the minimum and maximum dimensions a widget can have.
// They are passed down the widget tree during layout and used to constrain
// the size of child widgets.
//
// A constraint is considered "tight" when min equals max (widget must be exactly that size).
// A constraint is "loose" when min is 0 (widget can be any size up to max).
// A constraint is "unbounded" when max is Infinity (no upper limit).
//
// The zero value represents unconstrained (0 to Infinity on both dimensions).
type Constraints struct {
	MinWidth, MaxWidth   float32
	MinHeight, MaxHeight float32
}

// Tight creates constraints where the widget must be exactly the given size.
//
// Example:
//
//	c := geometry.Tight(geometry.Sz(100, 50))
//	// MinWidth=100, MaxWidth=100, MinHeight=50, MaxHeight=50
func Tight(size Size) Constraints {
	return Constraints{
		MinWidth:  size.Width,
		MaxWidth:  size.Width,
		MinHeight: size.Height,
		MaxHeight: size.Height,
	}
}

// TightWidth creates constraints with exact width but flexible height.
func TightWidth(width float32) Constraints {
	return Constraints{
		MinWidth:  width,
		MaxWidth:  width,
		MinHeight: 0,
		MaxHeight: Infinity,
	}
}

// TightHeight creates constraints with exact height but flexible width.
func TightHeight(height float32) Constraints {
	return Constraints{
		MinWidth:  0,
		MaxWidth:  Infinity,
		MinHeight: height,
		MaxHeight: height,
	}
}

// Loose creates constraints where the widget can be any size up to the given size.
//
// Example:
//
//	c := geometry.Loose(geometry.Sz(100, 50))
//	// MinWidth=0, MaxWidth=100, MinHeight=0, MaxHeight=50
func Loose(size Size) Constraints {
	return Constraints{
		MinWidth:  0,
		MaxWidth:  size.Width,
		MinHeight: 0,
		MaxHeight: size.Height,
	}
}

// Expand creates constraints where the widget should expand to fill available space.
// This is a loose constraint with infinite max dimensions.
//
// Example:
//
//	c := geometry.Expand()
//	// MinWidth=0, MaxWidth=Infinity, MinHeight=0, MaxHeight=Infinity
func Expand() Constraints {
	return Constraints{
		MinWidth:  0,
		MaxWidth:  Infinity,
		MinHeight: 0,
		MaxHeight: Infinity,
	}
}

// ExpandWidth creates constraints that expand horizontally but are flexible vertically.
func ExpandWidth(height float32) Constraints {
	return Constraints{
		MinWidth:  0,
		MaxWidth:  Infinity,
		MinHeight: 0,
		MaxHeight: height,
	}
}

// ExpandHeight creates constraints that expand vertically but are flexible horizontally.
func ExpandHeight(width float32) Constraints {
	return Constraints{
		MinWidth:  0,
		MaxWidth:  width,
		MinHeight: 0,
		MaxHeight: Infinity,
	}
}

// BoxConstraints creates constraints with explicit min/max dimensions.
//
// Example:
//
//	c := geometry.BoxConstraints(50, 200, 30, 100)
//	// MinWidth=50, MaxWidth=200, MinHeight=30, MaxHeight=100
func BoxConstraints(minWidth, maxWidth, minHeight, maxHeight float32) Constraints {
	return Constraints{
		MinWidth:  minWidth,
		MaxWidth:  maxWidth,
		MinHeight: minHeight,
		MaxHeight: maxHeight,
	}
}

// Constrain returns a size that satisfies these constraints.
// The returned size will be clamped to [min, max] for each dimension.
//
// Example:
//
//	c := geometry.BoxConstraints(50, 200, 30, 100)
//	s := c.Constrain(geometry.Sz(300, 150)) // Size{200, 100}
func (c Constraints) Constrain(size Size) Size {
	return Size{
		Width:  clamp32(size.Width, c.MinWidth, c.MaxWidth),
		Height: clamp32(size.Height, c.MinHeight, c.MaxHeight),
	}
}

// ConstrainWidth returns a width that satisfies these constraints.
func (c Constraints) ConstrainWidth(width float32) float32 {
	return clamp32(width, c.MinWidth, c.MaxWidth)
}

// ConstrainHeight returns a height that satisfies these constraints.
func (c Constraints) ConstrainHeight(height float32) float32 {
	return clamp32(height, c.MinHeight, c.MaxHeight)
}

// ConstrainDimensions returns dimensions that satisfy these constraints.
func (c Constraints) ConstrainDimensions(width, height float32) (float32, float32) {
	return clamp32(width, c.MinWidth, c.MaxWidth),
		clamp32(height, c.MinHeight, c.MaxHeight)
}

// Loosen returns constraints with the same max but min set to 0.
// This allows a child to be smaller than the parent's minimum.
//
// Example:
//
//	c := geometry.BoxConstraints(100, 200, 50, 100)
//	loose := c.Loosen() // MinWidth=0, MaxWidth=200, MinHeight=0, MaxHeight=100
func (c Constraints) Loosen() Constraints {
	return Constraints{
		MinWidth:  0,
		MaxWidth:  c.MaxWidth,
		MinHeight: 0,
		MaxHeight: c.MaxHeight,
	}
}

// LoosenWidth returns constraints with min width set to 0.
func (c Constraints) LoosenWidth() Constraints {
	return Constraints{
		MinWidth:  0,
		MaxWidth:  c.MaxWidth,
		MinHeight: c.MinHeight,
		MaxHeight: c.MaxHeight,
	}
}

// LoosenHeight returns constraints with min height set to 0.
func (c Constraints) LoosenHeight() Constraints {
	return Constraints{
		MinWidth:  c.MinWidth,
		MaxWidth:  c.MaxWidth,
		MinHeight: 0,
		MaxHeight: c.MaxHeight,
	}
}

// Tighten returns tight constraints using the given size, clamped to current constraints.
// This is useful when a widget wants to report its preferred size while respecting parent constraints.
//
// Example:
//
//	c := geometry.BoxConstraints(50, 200, 30, 100)
//	tight := c.Tighten(geometry.Sz(150, 80))
//	// MinWidth=150, MaxWidth=150, MinHeight=80, MaxHeight=80
func (c Constraints) Tighten(size Size) Constraints {
	w := clamp32(size.Width, c.MinWidth, c.MaxWidth)
	h := clamp32(size.Height, c.MinHeight, c.MaxHeight)
	return Constraints{
		MinWidth:  w,
		MaxWidth:  w,
		MinHeight: h,
		MaxHeight: h,
	}
}

// TightenWidth returns constraints with tight width, keeping height constraints.
func (c Constraints) TightenWidth(width float32) Constraints {
	w := clamp32(width, c.MinWidth, c.MaxWidth)
	return Constraints{
		MinWidth:  w,
		MaxWidth:  w,
		MinHeight: c.MinHeight,
		MaxHeight: c.MaxHeight,
	}
}

// TightenHeight returns constraints with tight height, keeping width constraints.
func (c Constraints) TightenHeight(height float32) Constraints {
	h := clamp32(height, c.MinHeight, c.MaxHeight)
	return Constraints{
		MinWidth:  c.MinWidth,
		MaxWidth:  c.MaxWidth,
		MinHeight: h,
		MaxHeight: h,
	}
}

// Enforce returns constraints that are at least as restrictive as both c and other.
// The result has the max of minimums and min of maximums.
func (c Constraints) Enforce(other Constraints) Constraints {
	return Constraints{
		MinWidth:  max32(c.MinWidth, other.MinWidth),
		MaxWidth:  min32(c.MaxWidth, other.MaxWidth),
		MinHeight: max32(c.MinHeight, other.MinHeight),
		MaxHeight: min32(c.MaxHeight, other.MaxHeight),
	}
}

// Deflate returns constraints reduced by the given insets.
// This is useful when calculating available space for content inside padding.
//
// Example:
//
//	c := geometry.Tight(geometry.Sz(100, 100))
//	padding := geometry.UniformInsets(10)
//	content := c.Deflate(padding) // MaxWidth=80, MaxHeight=80
func (c Constraints) Deflate(insets Insets) Constraints {
	horizontal := insets.Horizontal()
	vertical := insets.Vertical()
	return Constraints{
		MinWidth:  max32(0, c.MinWidth-horizontal),
		MaxWidth:  max32(0, c.MaxWidth-horizontal),
		MinHeight: max32(0, c.MinHeight-vertical),
		MaxHeight: max32(0, c.MaxHeight-vertical),
	}
}

// IsTight returns true if min equals max for both dimensions.
// A tight constraint forces a specific size.
func (c Constraints) IsTight() bool {
	return c.MinWidth == c.MaxWidth && c.MinHeight == c.MaxHeight
}

// IsTightWidth returns true if min width equals max width.
func (c Constraints) IsTightWidth() bool {
	return c.MinWidth == c.MaxWidth
}

// IsTightHeight returns true if min height equals max height.
func (c Constraints) IsTightHeight() bool {
	return c.MinHeight == c.MaxHeight
}

// IsUnbounded returns true if both max dimensions are Infinity.
// An unbounded constraint has no upper size limit.
func (c Constraints) IsUnbounded() bool {
	return c.MaxWidth >= Infinity && c.MaxHeight >= Infinity
}

// HasBoundedWidth returns true if max width is not Infinity.
func (c Constraints) HasBoundedWidth() bool {
	return c.MaxWidth < Infinity
}

// HasBoundedHeight returns true if max height is not Infinity.
func (c Constraints) HasBoundedHeight() bool {
	return c.MaxHeight < Infinity
}

// HasInfiniteWidth returns true if max width is Infinity.
func (c Constraints) HasInfiniteWidth() bool {
	return c.MaxWidth >= Infinity
}

// HasInfiniteHeight returns true if max height is Infinity.
func (c Constraints) HasInfiniteHeight() bool {
	return c.MaxHeight >= Infinity
}

// IsSatisfiedBy returns true if the given size satisfies these constraints.
//
// Example:
//
//	c := geometry.BoxConstraints(50, 200, 30, 100)
//	c.IsSatisfiedBy(geometry.Sz(100, 50)) // true
//	c.IsSatisfiedBy(geometry.Sz(300, 50)) // false (width too large)
func (c Constraints) IsSatisfiedBy(size Size) bool {
	return size.Width >= c.MinWidth && size.Width <= c.MaxWidth &&
		size.Height >= c.MinHeight && size.Height <= c.MaxHeight
}

// Normalize returns constraints with mins clamped to not exceed maxes,
// and ensures no negative values.
//
// Example:
//
//	c := geometry.BoxConstraints(200, 100, 50, 30) // Invalid: min > max
//	normalized := c.Normalize() // MinWidth=100, MaxWidth=100, MinHeight=30, MaxHeight=30
func (c Constraints) Normalize() Constraints {
	minW := max32(0, c.MinWidth)
	maxW := max32(0, c.MaxWidth)
	minH := max32(0, c.MinHeight)
	maxH := max32(0, c.MaxHeight)

	// Ensure min <= max
	if minW > maxW {
		minW = maxW
	}
	if minH > maxH {
		minH = maxH
	}

	return Constraints{
		MinWidth:  minW,
		MaxWidth:  maxW,
		MinHeight: minH,
		MaxHeight: maxH,
	}
}

// IsNormalized returns true if the constraints are valid (min <= max, no negatives).
func (c Constraints) IsNormalized() bool {
	return c.MinWidth >= 0 && c.MinWidth <= c.MaxWidth &&
		c.MinHeight >= 0 && c.MinHeight <= c.MaxHeight
}

// Smallest returns the smallest size that satisfies these constraints.
func (c Constraints) Smallest() Size {
	return Size{Width: c.MinWidth, Height: c.MinHeight}
}

// Biggest returns the biggest finite size that satisfies these constraints.
// If a dimension is unbounded (Infinity), it uses the minimum instead.
func (c Constraints) Biggest() Size {
	w := c.MaxWidth
	if w >= Infinity {
		w = c.MinWidth
	}
	h := c.MaxHeight
	if h >= Infinity {
		h = c.MinHeight
	}
	return Size{Width: w, Height: h}
}

// BigggestFinite returns the biggest size, treating Infinity as the provided fallback.
func (c Constraints) BiggestFinite(fallbackWidth, fallbackHeight float32) Size {
	w := c.MaxWidth
	if w >= Infinity {
		w = fallbackWidth
	}
	h := c.MaxHeight
	if h >= Infinity {
		h = fallbackHeight
	}
	return Size{Width: w, Height: h}
}

// IsZero returns true if all constraints are zero.
func (c Constraints) IsZero() bool {
	return c.MinWidth == 0 && c.MaxWidth == 0 && c.MinHeight == 0 && c.MaxHeight == 0
}

// IsNaN returns true if any value is NaN.
func (c Constraints) IsNaN() bool {
	return isNaN32(c.MinWidth) || isNaN32(c.MaxWidth) ||
		isNaN32(c.MinHeight) || isNaN32(c.MaxHeight)
}

// Sanitize returns constraints with NaN values replaced by appropriate defaults
// (0 for minimums, Infinity for maximums).
func (c Constraints) Sanitize() Constraints {
	minW, maxW := c.MinWidth, c.MaxWidth
	minH, maxH := c.MinHeight, c.MaxHeight
	if isNaN32(minW) {
		minW = 0
	}
	if isNaN32(maxW) {
		maxW = Infinity
	}
	if isNaN32(minH) {
		minH = 0
	}
	if isNaN32(maxH) {
		maxH = Infinity
	}
	return Constraints{
		MinWidth:  minW,
		MaxWidth:  maxW,
		MinHeight: minH,
		MaxHeight: maxH,
	}
}

// String returns a string representation of the constraints.
//
// Example:
//
//	c := geometry.BoxConstraints(50, 200, 30, 100)
//	s := c.String() // "Constraints(50<=w<=200, 30<=h<=100)"
func (c Constraints) String() string {
	wMin := formatConstraintValue(c.MinWidth)
	wMax := formatConstraintValue(c.MaxWidth)
	hMin := formatConstraintValue(c.MinHeight)
	hMax := formatConstraintValue(c.MaxHeight)
	return fmt.Sprintf("Constraints(%s<=w<=%s, %s<=h<=%s)", wMin, wMax, hMin, hMax)
}

// formatConstraintValue formats a float32 for constraint display.
// Returns "inf" for Infinity values.
func formatConstraintValue(v float32) string {
	if v >= Infinity {
		return "inf"
	}
	return fmt.Sprintf("%.4g", v)
}
