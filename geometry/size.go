package geometry

import "fmt"

// Size represents dimensions with Width and Height.
//
// Size uses float32 for GPU compatibility. Negative dimensions are allowed
// but may have special meaning in some contexts (e.g., "shrink to fit").
//
// The zero value represents empty dimensions (0x0).
type Size struct {
	Width, Height float32
}

// Sz is a shorthand constructor for creating a Size.
//
// Example:
//
//	s := geometry.Sz(100, 50)
func Sz(width, height float32) Size {
	return Size{Width: width, Height: height}
}

// Add returns a new size with dimensions added.
//
// Example:
//
//	s1 := geometry.Sz(100, 50)
//	s2 := geometry.Sz(10, 20)
//	result := s1.Add(s2) // Size{110, 70}
func (s Size) Add(other Size) Size {
	return Size{
		Width:  s.Width + other.Width,
		Height: s.Height + other.Height,
	}
}

// Sub returns a new size with dimensions subtracted.
//
// Example:
//
//	s1 := geometry.Sz(100, 50)
//	s2 := geometry.Sz(10, 20)
//	result := s1.Sub(s2) // Size{90, 30}
func (s Size) Sub(other Size) Size {
	return Size{
		Width:  s.Width - other.Width,
		Height: s.Height - other.Height,
	}
}

// Scale returns a new size with both dimensions multiplied by the scalar.
//
// Example:
//
//	s := geometry.Sz(100, 50)
//	result := s.Scale(2) // Size{200, 100}
func (s Size) Scale(scalar float32) Size {
	return Size{
		Width:  s.Width * scalar,
		Height: s.Height * scalar,
	}
}

// Area returns the area (Width * Height).
// Returns negative value if one dimension is negative.
//
// Example:
//
//	s := geometry.Sz(100, 50)
//	area := s.Area() // 5000
func (s Size) Area() float32 {
	return s.Width * s.Height
}

// IsZero returns true if both dimensions are zero.
func (s Size) IsZero() bool {
	return s.Width == 0 && s.Height == 0
}

// IsEmpty returns true if either dimension is zero or negative.
//
// Example:
//
//	geometry.Sz(100, 50).IsEmpty() // false
//	geometry.Sz(100, 0).IsEmpty()  // true
//	geometry.Sz(-1, 50).IsEmpty()  // true
func (s Size) IsEmpty() bool {
	return s.Width <= 0 || s.Height <= 0
}

// Contains returns true if s can contain other.
// Both dimensions of s must be >= corresponding dimensions of other.
//
// Example:
//
//	s1 := geometry.Sz(100, 50)
//	s2 := geometry.Sz(80, 40)
//	s1.Contains(s2) // true
func (s Size) Contains(other Size) bool {
	return s.Width >= other.Width && s.Height >= other.Height
}

// Expand returns a new size with dimensions increased by delta on each edge.
// Total increase is 2*delta for each dimension.
//
// Example:
//
//	s := geometry.Sz(100, 50)
//	result := s.Expand(10) // Size{120, 70}
func (s Size) Expand(delta float32) Size {
	return Size{
		Width:  s.Width + 2*delta,
		Height: s.Height + 2*delta,
	}
}

// Contract returns a new size with dimensions decreased by delta on each edge.
// Total decrease is 2*delta for each dimension.
//
// Example:
//
//	s := geometry.Sz(100, 50)
//	result := s.Contract(10) // Size{80, 30}
func (s Size) Contract(delta float32) Size {
	return Size{
		Width:  s.Width - 2*delta,
		Height: s.Height - 2*delta,
	}
}

// Min returns a new size with the minimum dimensions of s and other.
//
// Example:
//
//	s1 := geometry.Sz(100, 30)
//	s2 := geometry.Sz(80, 50)
//	result := s1.Min(s2) // Size{80, 30}
func (s Size) Min(other Size) Size {
	return Size{
		Width:  min32(s.Width, other.Width),
		Height: min32(s.Height, other.Height),
	}
}

// Max returns a new size with the maximum dimensions of s and other.
//
// Example:
//
//	s1 := geometry.Sz(100, 30)
//	s2 := geometry.Sz(80, 50)
//	result := s1.Max(s2) // Size{100, 50}
func (s Size) Max(other Size) Size {
	return Size{
		Width:  max32(s.Width, other.Width),
		Height: max32(s.Height, other.Height),
	}
}

// Clamp returns a new size with dimensions clamped to the range [minSize, maxSize].
//
// Example:
//
//	s := geometry.Sz(150, 25)
//	minS := geometry.Sz(50, 50)
//	maxS := geometry.Sz(100, 100)
//	result := s.Clamp(minS, maxS) // Size{100, 50}
func (s Size) Clamp(minSize, maxSize Size) Size {
	return Size{
		Width:  clamp32(s.Width, minSize.Width, maxSize.Width),
		Height: clamp32(s.Height, minSize.Height, maxSize.Height),
	}
}

// ToPoint converts the size to a point (Width -> X, Height -> Y).
// Useful for offset calculations.
func (s Size) ToPoint() Point {
	return Point{X: s.Width, Y: s.Height}
}

// AspectRatio returns Width / Height.
// Returns 0 if Height is zero.
func (s Size) AspectRatio() float32 {
	if s.Height == 0 {
		return 0
	}
	return s.Width / s.Height
}

// FitIn returns a size that maintains aspect ratio while fitting within maxSize.
// If the size already fits within maxSize, it is returned unchanged.
// Otherwise, the size is scaled down proportionally.
//
// Example:
//
//	s := geometry.Sz(200, 100)     // 2:1 aspect ratio
//	maxS := geometry.Sz(100, 100)
//	result := s.FitIn(maxS)        // Size{100, 50}
func (s Size) FitIn(maxSize Size) Size {
	if s.IsEmpty() || maxSize.IsEmpty() {
		return Size{}
	}

	// If already fits, return original
	if s.Width <= maxSize.Width && s.Height <= maxSize.Height {
		return s
	}

	scaleW := maxSize.Width / s.Width
	scaleH := maxSize.Height / s.Height
	scale := min32(scaleW, scaleH)

	return Size{
		Width:  s.Width * scale,
		Height: s.Height * scale,
	}
}

// FillIn returns a size that maintains aspect ratio while filling maxSize.
// The resulting size will have at least one dimension equal to maxSize,
// and will completely cover maxSize (may overflow in one dimension).
//
// Example:
//
//	s := geometry.Sz(200, 100)     // 2:1 aspect ratio
//	maxS := geometry.Sz(100, 100)
//	result := s.FillIn(maxS)       // Size{200, 100}
func (s Size) FillIn(maxSize Size) Size {
	if s.IsEmpty() || maxSize.IsEmpty() {
		return Size{}
	}

	scaleW := maxSize.Width / s.Width
	scaleH := maxSize.Height / s.Height
	scale := max32(scaleW, scaleH)

	return Size{
		Width:  s.Width * scale,
		Height: s.Height * scale,
	}
}

// IsNaN returns true if either dimension is NaN.
func (s Size) IsNaN() bool {
	return isNaN32(s.Width) || isNaN32(s.Height)
}

// Sanitize returns a copy of s with NaN values replaced by zero.
func (s Size) Sanitize() Size {
	w, h := s.Width, s.Height
	if isNaN32(w) {
		w = 0
	}
	if isNaN32(h) {
		h = 0
	}
	return Size{Width: w, Height: h}
}

// String returns a string representation of the size.
//
// Example:
//
//	s := geometry.Sz(100, 50)
//	str := s.String() // "Size(100, 50)"
func (s Size) String() string {
	return fmt.Sprintf("Size(%.4g, %.4g)", s.Width, s.Height)
}
