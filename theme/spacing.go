package theme

// SpacingScale defines a consistent spacing system for a theme.
//
// SpacingScale provides predefined spacing values that maintain visual
// consistency across the application. Values follow a base-4 scale,
// which is common in modern design systems.
//
// Use these values for margins, padding, gaps, and other spacing needs:
//
//	padding := theme.Spacing.M  // 16px standard padding
//	gap := theme.Spacing.S      // 8px gap between items
//
// The scale progression (XXS to XXXL) allows for fine-grained control
// while maintaining visual harmony.
type SpacingScale struct {
	// XXS is the smallest spacing value (2px).
	//
	// Use for tight spacing between closely related elements.
	XXS float32

	// XS is extra-small spacing (4px).
	//
	// Use for compact layouts or icon padding.
	XS float32

	// S is small spacing (8px).
	//
	// Use for spacing between related elements, small gaps.
	S float32

	// M is medium spacing (16px) - the base unit.
	//
	// Use for standard padding and margins.
	M float32

	// L is large spacing (24px).
	//
	// Use for separating distinct sections.
	L float32

	// XL is extra-large spacing (32px).
	//
	// Use for major section breaks or large containers.
	XL float32

	// XXL is double extra-large spacing (48px).
	//
	// Use for page-level margins or hero sections.
	XXL float32

	// XXXL is the largest spacing value (64px).
	//
	// Use for major layout divisions or very large screens.
	XXXL float32
}

// DefaultSpacing returns a SpacingScale following a base-4 progression.
//
// The scale uses multiples of 4 for consistent sizing:
//
//	XXS:  2px (half base)
//	XS:   4px (1x base)
//	S:    8px (2x base)
//	M:   16px (4x base)
//	L:   24px (6x base)
//	XL:  32px (8x base)
//	XXL: 48px (12x base)
//	XXXL: 64px (16x base)
func DefaultSpacing() SpacingScale {
	return SpacingScale{
		XXS:  2,
		XS:   4,
		S:    8,
		M:    16,
		L:    24,
		XL:   32,
		XXL:  48,
		XXXL: 64,
	}
}

// Scale returns a copy of the spacing scale with all values multiplied
// by the given factor.
//
// This is useful for density adjustments or accessibility settings.
// A factor of 1.0 returns the original values, 0.75 creates a compact scale,
// 1.5 creates a relaxed scale.
//
// Example:
//
//	compact := theme.DefaultSpacing().Scale(0.75)
//	relaxed := theme.DefaultSpacing().Scale(1.5)
func (s SpacingScale) Scale(factor float32) SpacingScale {
	return SpacingScale{
		XXS:  s.XXS * factor,
		XS:   s.XS * factor,
		S:    s.S * factor,
		M:    s.M * factor,
		L:    s.L * factor,
		XL:   s.XL * factor,
		XXL:  s.XXL * factor,
		XXXL: s.XXXL * factor,
	}
}

// Inset returns spacing values suitable for inset (padding) on all sides.
//
// The returned values represent top, right, bottom, left in that order,
// all set to the specified spacing value.
func (s SpacingScale) Inset(value float32) (top, right, bottom, left float32) {
	return value, value, value, value
}

// InsetHorizontal returns spacing values for horizontal-only inset.
//
// Returns zero for top/bottom and the specified value for right/left.
func (s SpacingScale) InsetHorizontal(value float32) (top, right, bottom, left float32) {
	return 0, value, 0, value
}

// InsetVertical returns spacing values for vertical-only inset.
//
// Returns the specified value for top/bottom and zero for right/left.
func (s SpacingScale) InsetVertical(value float32) (top, right, bottom, left float32) {
	return value, 0, value, 0
}

// Compact returns a compact version of the spacing scale (75% of original).
func (s SpacingScale) Compact() SpacingScale {
	return s.Scale(0.75)
}

// Relaxed returns a relaxed version of the spacing scale (150% of original).
func (s SpacingScale) Relaxed() SpacingScale {
	return s.Scale(1.5)
}

// DenseSpacing returns a compact spacing scale for dense layouts.
//
// This is a pre-configured scale suitable for data-dense interfaces:
//
//	XXS:  1px
//	XS:   2px
//	S:    4px
//	M:    8px
//	L:   12px
//	XL:  16px
//	XXL: 24px
//	XXXL: 32px
func DenseSpacing() SpacingScale {
	return SpacingScale{
		XXS:  1,
		XS:   2,
		S:    4,
		M:    8,
		L:    12,
		XL:   16,
		XXL:  24,
		XXXL: 32,
	}
}

// ComfortableSpacing returns a spacious scale for comfortable layouts.
//
// This is a pre-configured scale suitable for content-focused interfaces:
//
//	XXS:  4px
//	XS:   8px
//	S:   12px
//	M:   24px
//	L:   32px
//	XL:  48px
//	XXL: 64px
//	XXXL: 96px
func ComfortableSpacing() SpacingScale {
	return SpacingScale{
		XXS:  4,
		XS:   8,
		S:    12,
		M:    24,
		L:    32,
		XL:   48,
		XXL:  64,
		XXXL: 96,
	}
}
