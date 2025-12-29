package geometry

import "fmt"

// Rect represents an axis-aligned rectangle defined by its minimum and maximum corners.
//
// A rectangle is considered valid when Min.X <= Max.X and Min.Y <= Max.Y.
// Some operations may return invalid rectangles (e.g., Intersection when
// rectangles don't overlap). Use IsEmpty() to check validity.
//
// The zero value is an empty rectangle at the origin.
type Rect struct {
	Min, Max Point
}

// NewRect creates a rectangle from position (x, y) and dimensions (w, h).
//
// Example:
//
//	r := geometry.NewRect(10, 20, 100, 50)
//	// Min: (10, 20), Max: (110, 70)
func NewRect(x, y, width, height float32) Rect {
	return Rect{
		Min: Point{X: x, Y: y},
		Max: Point{X: x + width, Y: y + height},
	}
}

// FromPointSize creates a rectangle from a point (top-left) and size.
//
// Example:
//
//	r := geometry.FromPointSize(geometry.Pt(10, 20), geometry.Sz(100, 50))
func FromPointSize(p Point, s Size) Rect {
	return Rect{
		Min: p,
		Max: Point{X: p.X + s.Width, Y: p.Y + s.Height},
	}
}

// FromCenter creates a rectangle centered at the given point with the given size.
//
// Example:
//
//	r := geometry.FromCenter(geometry.Pt(50, 50), geometry.Sz(100, 50))
//	// Min: (0, 25), Max: (100, 75)
func FromCenter(center Point, s Size) Rect {
	halfW := s.Width / 2
	halfH := s.Height / 2
	return Rect{
		Min: Point{X: center.X - halfW, Y: center.Y - halfH},
		Max: Point{X: center.X + halfW, Y: center.Y + halfH},
	}
}

// FromMinMax creates a rectangle from two corner points.
// The points are normalized so Min contains the smaller coordinates.
func FromMinMax(p1, p2 Point) Rect {
	return Rect{
		Min: p1.Min(p2),
		Max: p1.Max(p2),
	}
}

// Size returns the dimensions of the rectangle.
//
// Example:
//
//	r := geometry.NewRect(10, 20, 100, 50)
//	s := r.Size() // Size{100, 50}
func (r Rect) Size() Size {
	return Size{
		Width:  r.Max.X - r.Min.X,
		Height: r.Max.Y - r.Min.Y,
	}
}

// Width returns the width of the rectangle.
func (r Rect) Width() float32 {
	return r.Max.X - r.Min.X
}

// Height returns the height of the rectangle.
func (r Rect) Height() float32 {
	return r.Max.Y - r.Min.Y
}

// Center returns the center point of the rectangle.
//
// Example:
//
//	r := geometry.NewRect(0, 0, 100, 50)
//	c := r.Center() // Point{50, 25}
func (r Rect) Center() Point {
	return Point{
		X: (r.Min.X + r.Max.X) / 2,
		Y: (r.Min.Y + r.Max.Y) / 2,
	}
}

// TopLeft returns the top-left corner (same as Min).
func (r Rect) TopLeft() Point {
	return r.Min
}

// TopRight returns the top-right corner.
func (r Rect) TopRight() Point {
	return Point{X: r.Max.X, Y: r.Min.Y}
}

// BottomLeft returns the bottom-left corner.
func (r Rect) BottomLeft() Point {
	return Point{X: r.Min.X, Y: r.Max.Y}
}

// BottomRight returns the bottom-right corner (same as Max).
func (r Rect) BottomRight() Point {
	return r.Max
}

// Contains returns true if the point p is inside or on the edge of the rectangle.
//
// Example:
//
//	r := geometry.NewRect(0, 0, 100, 50)
//	r.Contains(geometry.Pt(50, 25)) // true
//	r.Contains(geometry.Pt(150, 25)) // false
func (r Rect) Contains(p Point) bool {
	return p.X >= r.Min.X && p.X <= r.Max.X &&
		p.Y >= r.Min.Y && p.Y <= r.Max.Y
}

// ContainsRect returns true if other is entirely inside or equal to r.
//
// Example:
//
//	outer := geometry.NewRect(0, 0, 100, 100)
//	inner := geometry.NewRect(10, 10, 50, 50)
//	outer.ContainsRect(inner) // true
func (r Rect) ContainsRect(other Rect) bool {
	return other.Min.X >= r.Min.X && other.Max.X <= r.Max.X &&
		other.Min.Y >= r.Min.Y && other.Max.Y <= r.Max.Y
}

// Intersects returns true if r and other overlap (share any area).
//
// Example:
//
//	r1 := geometry.NewRect(0, 0, 100, 100)
//	r2 := geometry.NewRect(50, 50, 100, 100)
//	r1.Intersects(r2) // true
func (r Rect) Intersects(other Rect) bool {
	return r.Min.X < other.Max.X && r.Max.X > other.Min.X &&
		r.Min.Y < other.Max.Y && r.Max.Y > other.Min.Y
}

// Intersection returns the overlapping area of r and other.
// If the rectangles don't overlap, returns an empty rectangle.
//
// Example:
//
//	r1 := geometry.NewRect(0, 0, 100, 100)
//	r2 := geometry.NewRect(50, 50, 100, 100)
//	inter := r1.Intersection(r2) // Rect{Min:(50,50), Max:(100,100)}
func (r Rect) Intersection(other Rect) Rect {
	result := Rect{
		Min: r.Min.Max(other.Min),
		Max: r.Max.Min(other.Max),
	}
	// Return empty rect if no intersection
	if result.Min.X >= result.Max.X || result.Min.Y >= result.Max.Y {
		return Rect{}
	}
	return result
}

// Union returns the smallest rectangle containing both r and other.
//
// Example:
//
//	r1 := geometry.NewRect(0, 0, 50, 50)
//	r2 := geometry.NewRect(100, 100, 50, 50)
//	union := r1.Union(r2) // Rect{Min:(0,0), Max:(150,150)}
func (r Rect) Union(other Rect) Rect {
	// Handle empty rectangles
	if r.IsEmpty() {
		return other
	}
	if other.IsEmpty() {
		return r
	}
	return Rect{
		Min: r.Min.Min(other.Min),
		Max: r.Max.Max(other.Max),
	}
}

// Inset returns a new rectangle with edges moved inward by the specified insets.
// Positive inset values shrink the rectangle, negative values expand it.
//
// Example:
//
//	r := geometry.NewRect(0, 0, 100, 100)
//	padding := geometry.UniformInsets(10)
//	inner := r.Inset(padding) // Rect{Min:(10,10), Max:(90,90)}
func (r Rect) Inset(insets Insets) Rect {
	return Rect{
		Min: Point{X: r.Min.X + insets.Left, Y: r.Min.Y + insets.Top},
		Max: Point{X: r.Max.X - insets.Right, Y: r.Max.Y - insets.Bottom},
	}
}

// Expand returns a new rectangle with all edges moved outward by delta.
// Negative delta values shrink the rectangle.
//
// Example:
//
//	r := geometry.NewRect(10, 10, 80, 80)
//	expanded := r.Expand(5) // Rect{Min:(5,5), Max:(95,95)}
func (r Rect) Expand(delta float32) Rect {
	return Rect{
		Min: Point{X: r.Min.X - delta, Y: r.Min.Y - delta},
		Max: Point{X: r.Max.X + delta, Y: r.Max.Y + delta},
	}
}

// Translate returns a new rectangle moved by the given offset.
//
// Example:
//
//	r := geometry.NewRect(0, 0, 100, 50)
//	moved := r.Translate(geometry.Pt(10, 20))
//	// Rect{Min:(10,20), Max:(110,70)}
func (r Rect) Translate(offset Point) Rect {
	return Rect{
		Min: r.Min.Add(offset),
		Max: r.Max.Add(offset),
	}
}

// TranslateXY returns a new rectangle moved by (dx, dy).
func (r Rect) TranslateXY(dx, dy float32) Rect {
	return r.Translate(Point{X: dx, Y: dy})
}

// WithSize returns a new rectangle with the same Min point but different size.
func (r Rect) WithSize(s Size) Rect {
	return FromPointSize(r.Min, s)
}

// WithCenter returns a new rectangle with the same size but centered at the given point.
func (r Rect) WithCenter(center Point) Rect {
	return FromCenter(center, r.Size())
}

// IsZero returns true if the rectangle is the zero value.
func (r Rect) IsZero() bool {
	return r.Min.IsZero() && r.Max.IsZero()
}

// IsEmpty returns true if the rectangle has zero or negative area.
// This can happen when Min >= Max in either dimension.
func (r Rect) IsEmpty() bool {
	return r.Min.X >= r.Max.X || r.Min.Y >= r.Max.Y
}

// Area returns the area of the rectangle.
// Returns 0 for empty rectangles.
func (r Rect) Area() float32 {
	if r.IsEmpty() {
		return 0
	}
	return r.Width() * r.Height()
}

// Normalize returns a copy of r with Min and Max swapped if necessary
// so that Min contains the smaller coordinates.
func (r Rect) Normalize() Rect {
	return FromMinMax(r.Min, r.Max)
}

// IsNaN returns true if any coordinate is NaN.
func (r Rect) IsNaN() bool {
	return r.Min.IsNaN() || r.Max.IsNaN()
}

// Sanitize returns a copy of r with NaN values replaced by zero.
func (r Rect) Sanitize() Rect {
	return Rect{
		Min: r.Min.Sanitize(),
		Max: r.Max.Sanitize(),
	}
}

// String returns a string representation of the rectangle.
//
// Example:
//
//	r := geometry.NewRect(10, 20, 100, 50)
//	str := r.String() // "Rect(10, 20, 100x50)"
func (r Rect) String() string {
	return fmt.Sprintf("Rect(%.4g, %.4g, %.4gx%.4g)", r.Min.X, r.Min.Y, r.Width(), r.Height())
}
