package theme

// RadiusScale defines the border radius values for a theme.
//
// RadiusScale provides consistent corner rounding across the application.
// Following Material 3, the scale ranges from no rounding (None) to
// fully circular (Full).
//
// Usage:
//
//	radius := theme.Radii.M  // Standard card radius
//	buttonRadius := theme.Radii.Full  // Pill-shaped buttons
type RadiusScale struct {
	// None is no rounding (sharp corners) - 0px.
	None float32

	// XS is extra-small rounding - 2px.
	//
	// Use for subtle rounding on small elements.
	XS float32

	// S is small rounding - 4px.
	//
	// Use for buttons, chips, and small cards.
	S float32

	// M is medium rounding - 8px.
	//
	// Use for standard cards and containers.
	M float32

	// L is large rounding - 12px.
	//
	// Use for dialogs and larger containers.
	L float32

	// XL is extra-large rounding - 16px.
	//
	// Use for sheets and prominent surfaces.
	XL float32

	// XXL is double extra-large rounding - 24px.
	//
	// Use for floating action buttons and rounded shapes.
	XXL float32

	// Full is fully circular rounding - typically 9999px.
	//
	// Use for pills, badges, and circular buttons.
	// A very large value ensures circular shapes regardless of element size.
	Full float32
}

// DefaultRadii returns a RadiusScale following Material 3 guidelines.
//
// The scale provides sizes suitable for various UI elements:
//
//	None: 0px   - Sharp corners
//	XS:   2px   - Subtle rounding
//	S:    4px   - Buttons, chips
//	M:    8px   - Cards
//	L:   12px   - Dialogs
//	XL:  16px   - Sheets
//	XXL: 24px   - FABs
//	Full: 9999px - Pills, circles
func DefaultRadii() RadiusScale {
	return RadiusScale{
		None: 0,
		XS:   2,
		S:    4,
		M:    8,
		L:    12,
		XL:   16,
		XXL:  24,
		Full: 9999, // Large enough to create circles/pills
	}
}

// Scale returns a copy of the radius scale with all values multiplied
// by the given factor (except None and Full).
//
// This is useful for adjusting the overall "roundness" of a theme.
// None stays at 0, Full stays at 9999.
//
// Example:
//
//	sharper := theme.DefaultRadii().Scale(0.5)  // Half the rounding
//	rounder := theme.DefaultRadii().Scale(1.5)  // More rounded
func (r RadiusScale) Scale(factor float32) RadiusScale {
	return RadiusScale{
		None: 0, // Always 0
		XS:   r.XS * factor,
		S:    r.S * factor,
		M:    r.M * factor,
		L:    r.L * factor,
		XL:   r.XL * factor,
		XXL:  r.XXL * factor,
		Full: 9999, // Always maximum
	}
}

// Clamp returns the radius clamped to the range [minVal, maxVal].
//
// This is useful when you need to ensure a radius fits within
// specific bounds for a particular element.
func (r RadiusScale) Clamp(value, minVal, maxVal float32) float32 {
	if value < minVal {
		return minVal
	}
	if value > maxVal {
		return maxVal
	}
	return value
}

// SharpRadii returns a RadiusScale with minimal rounding.
//
// This creates a sharp, angular aesthetic:
//
//	None: 0px
//	XS:   1px
//	S:    2px
//	M:    3px
//	L:    4px
//	XL:   6px
//	XXL:  8px
//	Full: 9999px
func SharpRadii() RadiusScale {
	return RadiusScale{
		None: 0,
		XS:   1,
		S:    2,
		M:    3,
		L:    4,
		XL:   6,
		XXL:  8,
		Full: 9999,
	}
}

// SoftRadii returns a RadiusScale with generous rounding.
//
// This creates a soft, friendly aesthetic:
//
//	None: 0px
//	XS:   4px
//	S:    8px
//	M:   16px
//	L:   24px
//	XL:  32px
//	XXL: 48px
//	Full: 9999px
func SoftRadii() RadiusScale {
	return RadiusScale{
		None: 0,
		XS:   4,
		S:    8,
		M:    16,
		L:    24,
		XL:   32,
		XXL:  48,
		Full: 9999,
	}
}

// CornerRadius represents corner radii for a rectangle.
//
// This allows different radius values for each corner, useful for
// elements like tabs that need selective rounding.
type CornerRadius struct {
	TopLeft     float32
	TopRight    float32
	BottomRight float32
	BottomLeft  float32
}

// Uniform creates a CornerRadius with the same value for all corners.
func Uniform(radius float32) CornerRadius {
	return CornerRadius{
		TopLeft:     radius,
		TopRight:    radius,
		BottomRight: radius,
		BottomLeft:  radius,
	}
}

// Top creates a CornerRadius with rounding only on the top corners.
func Top(radius float32) CornerRadius {
	return CornerRadius{
		TopLeft:     radius,
		TopRight:    radius,
		BottomRight: 0,
		BottomLeft:  0,
	}
}

// Bottom creates a CornerRadius with rounding only on the bottom corners.
func Bottom(radius float32) CornerRadius {
	return CornerRadius{
		TopLeft:     0,
		TopRight:    0,
		BottomRight: radius,
		BottomLeft:  radius,
	}
}

// Left creates a CornerRadius with rounding only on the left corners.
func Left(radius float32) CornerRadius {
	return CornerRadius{
		TopLeft:     radius,
		TopRight:    0,
		BottomRight: 0,
		BottomLeft:  radius,
	}
}

// Right creates a CornerRadius with rounding only on the right corners.
func Right(radius float32) CornerRadius {
	return CornerRadius{
		TopLeft:     0,
		TopRight:    radius,
		BottomRight: radius,
		BottomLeft:  0,
	}
}

// IsUniform returns true if all corners have the same radius.
func (c CornerRadius) IsUniform() bool {
	return c.TopLeft == c.TopRight &&
		c.TopRight == c.BottomRight &&
		c.BottomRight == c.BottomLeft
}

// Max returns the largest corner radius value.
func (c CornerRadius) Max() float32 {
	maxVal := c.TopLeft
	if c.TopRight > maxVal {
		maxVal = c.TopRight
	}
	if c.BottomRight > maxVal {
		maxVal = c.BottomRight
	}
	if c.BottomLeft > maxVal {
		maxVal = c.BottomLeft
	}
	return maxVal
}

// Scale returns a copy with all corners scaled by the given factor.
func (c CornerRadius) Scale(factor float32) CornerRadius {
	return CornerRadius{
		TopLeft:     c.TopLeft * factor,
		TopRight:    c.TopRight * factor,
		BottomRight: c.BottomRight * factor,
		BottomLeft:  c.BottomLeft * factor,
	}
}
