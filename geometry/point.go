package geometry

import (
	"fmt"
	"math"
)

// Point represents a 2D point with float32 coordinates.
//
// Point uses float32 for GPU compatibility and efficient memory layout.
// All operations return new values without modifying the original point.
//
// The zero value is the origin point (0, 0).
type Point struct {
	X, Y float32
}

// Pt is a shorthand constructor for creating a Point.
//
// Example:
//
//	p := geometry.Pt(10, 20)
func Pt(x, y float32) Point {
	return Point{X: x, Y: y}
}

// Add returns a new point with the coordinates of p added to other.
//
// Example:
//
//	p1 := geometry.Pt(10, 20)
//	p2 := geometry.Pt(5, 5)
//	result := p1.Add(p2) // Point{15, 25}
func (p Point) Add(other Point) Point {
	return Point{
		X: p.X + other.X,
		Y: p.Y + other.Y,
	}
}

// Sub returns a new point with the coordinates of other subtracted from p.
//
// Example:
//
//	p1 := geometry.Pt(10, 20)
//	p2 := geometry.Pt(3, 5)
//	result := p1.Sub(p2) // Point{7, 15}
func (p Point) Sub(other Point) Point {
	return Point{
		X: p.X - other.X,
		Y: p.Y - other.Y,
	}
}

// Scale returns a new point with both coordinates multiplied by the scalar.
//
// Example:
//
//	p := geometry.Pt(10, 20)
//	result := p.Scale(2) // Point{20, 40}
func (p Point) Scale(scalar float32) Point {
	return Point{
		X: p.X * scalar,
		Y: p.Y * scalar,
	}
}

// Mul returns a new point with coordinates multiplied component-wise.
//
// Example:
//
//	p1 := geometry.Pt(10, 20)
//	p2 := geometry.Pt(2, 3)
//	result := p1.Mul(p2) // Point{20, 60}
func (p Point) Mul(other Point) Point {
	return Point{
		X: p.X * other.X,
		Y: p.Y * other.Y,
	}
}

// Div returns a new point with coordinates divided component-wise.
// If a component of other is zero, the result for that component is zero.
//
// Example:
//
//	p1 := geometry.Pt(10, 20)
//	p2 := geometry.Pt(2, 4)
//	result := p1.Div(p2) // Point{5, 5}
func (p Point) Div(other Point) Point {
	var x, y float32
	if other.X != 0 {
		x = p.X / other.X
	}
	if other.Y != 0 {
		y = p.Y / other.Y
	}
	return Point{X: x, Y: y}
}

// Distance returns the Euclidean distance between p and other.
//
// Example:
//
//	p1 := geometry.Pt(0, 0)
//	p2 := geometry.Pt(3, 4)
//	d := p1.Distance(p2) // 5.0
func (p Point) Distance(other Point) float32 {
	return float32(math.Sqrt(float64(p.DistanceSquared(other))))
}

// DistanceSquared returns the squared Euclidean distance between p and other.
// This is faster than Distance when only comparing distances.
//
// Example:
//
//	p1 := geometry.Pt(0, 0)
//	p2 := geometry.Pt(3, 4)
//	d2 := p1.DistanceSquared(p2) // 25.0
func (p Point) DistanceSquared(other Point) float32 {
	dx := p.X - other.X
	dy := p.Y - other.Y
	return dx*dx + dy*dy
}

// Lerp returns a point linearly interpolated between p and other.
// t=0 returns p, t=1 returns other, values outside [0,1] extrapolate.
//
// Example:
//
//	p1 := geometry.Pt(0, 0)
//	p2 := geometry.Pt(10, 20)
//	mid := p1.Lerp(p2, 0.5) // Point{5, 10}
func (p Point) Lerp(other Point, t float32) Point {
	return Point{
		X: p.X + (other.X-p.X)*t,
		Y: p.Y + (other.Y-p.Y)*t,
	}
}

// Negate returns a new point with both coordinates negated.
//
// Example:
//
//	p := geometry.Pt(10, -20)
//	result := p.Negate() // Point{-10, 20}
func (p Point) Negate() Point {
	return Point{X: -p.X, Y: -p.Y}
}

// Normalize returns a unit vector in the same direction as p.
// If p is the zero point, returns the zero point.
//
// Example:
//
//	p := geometry.Pt(3, 4)
//	n := p.Normalize() // Point{0.6, 0.8}
func (p Point) Normalize() Point {
	length := p.Length()
	if length == 0 {
		return Point{}
	}
	return Point{
		X: p.X / length,
		Y: p.Y / length,
	}
}

// Length returns the length (magnitude) of the vector from origin to p.
//
// Example:
//
//	p := geometry.Pt(3, 4)
//	l := p.Length() // 5.0
func (p Point) Length() float32 {
	return float32(math.Sqrt(float64(p.X*p.X + p.Y*p.Y)))
}

// LengthSquared returns the squared length of the vector from origin to p.
// This is faster than Length when only comparing lengths.
func (p Point) LengthSquared() float32 {
	return p.X*p.X + p.Y*p.Y
}

// Dot returns the dot product of p and other.
//
// Example:
//
//	p1 := geometry.Pt(1, 2)
//	p2 := geometry.Pt(3, 4)
//	d := p1.Dot(p2) // 11.0 (1*3 + 2*4)
func (p Point) Dot(other Point) float32 {
	return p.X*other.X + p.Y*other.Y
}

// Min returns a new point with the minimum coordinates of p and other.
//
// Example:
//
//	p1 := geometry.Pt(10, 5)
//	p2 := geometry.Pt(3, 20)
//	result := p1.Min(p2) // Point{3, 5}
func (p Point) Min(other Point) Point {
	return Point{
		X: min32(p.X, other.X),
		Y: min32(p.Y, other.Y),
	}
}

// Max returns a new point with the maximum coordinates of p and other.
//
// Example:
//
//	p1 := geometry.Pt(10, 5)
//	p2 := geometry.Pt(3, 20)
//	result := p1.Max(p2) // Point{10, 20}
func (p Point) Max(other Point) Point {
	return Point{
		X: max32(p.X, other.X),
		Y: max32(p.Y, other.Y),
	}
}

// Clamp returns a new point with coordinates clamped to the range [minP, maxP].
//
// Example:
//
//	p := geometry.Pt(15, -5)
//	minP := geometry.Pt(0, 0)
//	maxP := geometry.Pt(10, 10)
//	result := p.Clamp(minP, maxP) // Point{10, 0}
func (p Point) Clamp(minP, maxP Point) Point {
	return Point{
		X: clamp32(p.X, minP.X, maxP.X),
		Y: clamp32(p.Y, minP.Y, maxP.Y),
	}
}

// IsZero returns true if both coordinates are zero.
func (p Point) IsZero() bool {
	return p.X == 0 && p.Y == 0
}

// IsNaN returns true if either coordinate is NaN.
func (p Point) IsNaN() bool {
	return isNaN32(p.X) || isNaN32(p.Y)
}

// Sanitize returns a copy of p with NaN values replaced by zero.
func (p Point) Sanitize() Point {
	x, y := p.X, p.Y
	if isNaN32(x) {
		x = 0
	}
	if isNaN32(y) {
		y = 0
	}
	return Point{X: x, Y: y}
}

// String returns a string representation of the point.
//
// Example:
//
//	p := geometry.Pt(10.5, 20.25)
//	s := p.String() // "Point(10.5, 20.25)"
func (p Point) String() string {
	return fmt.Sprintf("Point(%.4g, %.4g)", p.X, p.Y)
}

// helper functions for float32
func min32(a, b float32) float32 {
	if a < b {
		return a
	}
	return b
}

func max32(a, b float32) float32 {
	if a > b {
		return a
	}
	return b
}

func clamp32(v, minVal, maxVal float32) float32 {
	if v < minVal {
		return minVal
	}
	if v > maxVal {
		return maxVal
	}
	return v
}

func isNaN32(f float32) bool {
	// IEEE 754: NaN != NaN is the standard check
	return f != f
}
