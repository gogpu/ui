package geometry

import "fmt"

// Insets represents edge insets for padding, margins, or borders.
//
// Insets define spacing on each of the four edges of a rectangle.
// They are commonly used for padding (space inside a widget) and
// margins (space outside a widget).
//
// The zero value represents no insets (0 on all edges).
type Insets struct {
	Top, Right, Bottom, Left float32
}

// UniformInsets creates insets with the same value on all edges.
//
// Example:
//
//	padding := geometry.UniformInsets(16)
//	// Insets{Top: 16, Right: 16, Bottom: 16, Left: 16}
func UniformInsets(all float32) Insets {
	return Insets{Top: all, Right: all, Bottom: all, Left: all}
}

// SymmetricInsets creates insets with separate horizontal and vertical values.
//
// Example:
//
//	padding := geometry.SymmetricInsets(16, 8)
//	// Insets{Top: 8, Right: 16, Bottom: 8, Left: 16}
func SymmetricInsets(horizontal, vertical float32) Insets {
	return Insets{Top: vertical, Right: horizontal, Bottom: vertical, Left: horizontal}
}

// InsetsLTRB creates insets from Left, Top, Right, Bottom values.
// This ordering matches CSS shorthand (when specifying all four values).
//
// Example:
//
//	margin := geometry.InsetsLTRB(10, 20, 30, 40)
//	// Insets{Top: 20, Right: 30, Bottom: 40, Left: 10}
func InsetsLTRB(left, top, right, bottom float32) Insets {
	return Insets{Top: top, Right: right, Bottom: bottom, Left: left}
}

// InsetsTRBL creates insets from Top, Right, Bottom, Left values.
// This ordering matches CSS shorthand.
//
// Example:
//
//	margin := geometry.InsetsTRBL(20, 30, 40, 10)
//	// Insets{Top: 20, Right: 30, Bottom: 40, Left: 10}
func InsetsTRBL(top, right, bottom, left float32) Insets {
	return Insets{Top: top, Right: right, Bottom: bottom, Left: left}
}

// InsetsOnly creates insets with specific edges set, others zero.
//
// Example:
//
//	topPadding := geometry.InsetsOnly(10, 0, 0, 0) // top only
func InsetsOnly(top, right, bottom, left float32) Insets {
	return Insets{Top: top, Right: right, Bottom: bottom, Left: left}
}

// Horizontal returns the sum of left and right insets.
//
// Example:
//
//	padding := geometry.UniformInsets(16)
//	h := padding.Horizontal() // 32
func (i Insets) Horizontal() float32 {
	return i.Left + i.Right
}

// Vertical returns the sum of top and bottom insets.
//
// Example:
//
//	padding := geometry.UniformInsets(16)
//	v := padding.Vertical() // 32
func (i Insets) Vertical() float32 {
	return i.Top + i.Bottom
}

// Size returns a Size representing the total horizontal and vertical insets.
// This is useful for calculating how much space insets consume.
//
// Example:
//
//	padding := geometry.SymmetricInsets(16, 8)
//	s := padding.Size() // Size{Width: 32, Height: 16}
func (i Insets) Size() Size {
	return Size{Width: i.Horizontal(), Height: i.Vertical()}
}

// TopLeft returns the top-left offset as a Point.
func (i Insets) TopLeft() Point {
	return Point{X: i.Left, Y: i.Top}
}

// BottomRight returns the bottom-right offset as a Point.
func (i Insets) BottomRight() Point {
	return Point{X: i.Right, Y: i.Bottom}
}

// Add returns new insets with values added.
//
// Example:
//
//	i1 := geometry.UniformInsets(10)
//	i2 := geometry.UniformInsets(5)
//	result := i1.Add(i2) // Insets{15, 15, 15, 15}
func (i Insets) Add(other Insets) Insets {
	return Insets{
		Top:    i.Top + other.Top,
		Right:  i.Right + other.Right,
		Bottom: i.Bottom + other.Bottom,
		Left:   i.Left + other.Left,
	}
}

// Sub returns new insets with values subtracted.
func (i Insets) Sub(other Insets) Insets {
	return Insets{
		Top:    i.Top - other.Top,
		Right:  i.Right - other.Right,
		Bottom: i.Bottom - other.Bottom,
		Left:   i.Left - other.Left,
	}
}

// Scale returns new insets with all values multiplied by the scalar.
//
// Example:
//
//	padding := geometry.UniformInsets(10)
//	doubled := padding.Scale(2) // Insets{20, 20, 20, 20}
func (i Insets) Scale(scalar float32) Insets {
	return Insets{
		Top:    i.Top * scalar,
		Right:  i.Right * scalar,
		Bottom: i.Bottom * scalar,
		Left:   i.Left * scalar,
	}
}

// Negate returns insets with all values negated.
func (i Insets) Negate() Insets {
	return Insets{
		Top:    -i.Top,
		Right:  -i.Right,
		Bottom: -i.Bottom,
		Left:   -i.Left,
	}
}

// Min returns insets with the minimum values from i and other.
func (i Insets) Min(other Insets) Insets {
	return Insets{
		Top:    min32(i.Top, other.Top),
		Right:  min32(i.Right, other.Right),
		Bottom: min32(i.Bottom, other.Bottom),
		Left:   min32(i.Left, other.Left),
	}
}

// Max returns insets with the maximum values from i and other.
func (i Insets) Max(other Insets) Insets {
	return Insets{
		Top:    max32(i.Top, other.Top),
		Right:  max32(i.Right, other.Right),
		Bottom: max32(i.Bottom, other.Bottom),
		Left:   max32(i.Left, other.Left),
	}
}

// Clamp returns insets with values clamped to the range [minInsets, maxInsets].
func (i Insets) Clamp(minInsets, maxInsets Insets) Insets {
	return Insets{
		Top:    clamp32(i.Top, minInsets.Top, maxInsets.Top),
		Right:  clamp32(i.Right, minInsets.Right, maxInsets.Right),
		Bottom: clamp32(i.Bottom, minInsets.Bottom, maxInsets.Bottom),
		Left:   clamp32(i.Left, minInsets.Left, maxInsets.Left),
	}
}

// IsZero returns true if all inset values are zero.
func (i Insets) IsZero() bool {
	return i.Top == 0 && i.Right == 0 && i.Bottom == 0 && i.Left == 0
}

// IsUniform returns true if all edges have the same value.
func (i Insets) IsUniform() bool {
	return i.Top == i.Right && i.Right == i.Bottom && i.Bottom == i.Left
}

// IsSymmetric returns true if horizontal edges match and vertical edges match.
func (i Insets) IsSymmetric() bool {
	return i.Left == i.Right && i.Top == i.Bottom
}

// IsNonNegative returns true if all values are >= 0.
func (i Insets) IsNonNegative() bool {
	return i.Top >= 0 && i.Right >= 0 && i.Bottom >= 0 && i.Left >= 0
}

// IsNaN returns true if any value is NaN.
func (i Insets) IsNaN() bool {
	return isNaN32(i.Top) || isNaN32(i.Right) || isNaN32(i.Bottom) || isNaN32(i.Left)
}

// Sanitize returns a copy with NaN values replaced by zero.
func (i Insets) Sanitize() Insets {
	top, right, bottom, left := i.Top, i.Right, i.Bottom, i.Left
	if isNaN32(top) {
		top = 0
	}
	if isNaN32(right) {
		right = 0
	}
	if isNaN32(bottom) {
		bottom = 0
	}
	if isNaN32(left) {
		left = 0
	}
	return Insets{Top: top, Right: right, Bottom: bottom, Left: left}
}

// Abs returns insets with all values converted to absolute values.
func (i Insets) Abs() Insets {
	return Insets{
		Top:    abs32(i.Top),
		Right:  abs32(i.Right),
		Bottom: abs32(i.Bottom),
		Left:   abs32(i.Left),
	}
}

// String returns a string representation of the insets.
//
// Example:
//
//	i := geometry.UniformInsets(16)
//	s := i.String() // "Insets(16, 16, 16, 16)"
func (i Insets) String() string {
	return fmt.Sprintf("Insets(%.4g, %.4g, %.4g, %.4g)", i.Top, i.Right, i.Bottom, i.Left)
}

// abs32 returns the absolute value of a float32.
func abs32(f float32) float32 {
	if f < 0 {
		return -f
	}
	return f
}
