package theme

import "github.com/gogpu/ui/widget"

// Shadow defines a single shadow effect.
//
// Shadows create depth and elevation effects by drawing blurred shapes
// behind elements. They follow the CSS box-shadow model with offset,
// blur, spread, and color properties.
//
// Example:
//
//	shadow := theme.Shadow{
//	    OffsetX: 0,
//	    OffsetY: 2,
//	    Blur:    4,
//	    Spread:  0,
//	    Color:   widget.RGBA(0, 0, 0, 0.1),
//	}
type Shadow struct {
	// OffsetX is the horizontal offset of the shadow in logical pixels.
	//
	// Positive values move the shadow to the right, negative to the left.
	OffsetX float32

	// OffsetY is the vertical offset of the shadow in logical pixels.
	//
	// Positive values move the shadow down, negative values move it up.
	OffsetY float32

	// Blur is the blur radius in logical pixels.
	//
	// Larger values create softer, more diffuse shadows. A value of 0
	// creates a sharp shadow.
	Blur float32

	// Spread is the spread radius in logical pixels.
	//
	// Positive values expand the shadow, negative values contract it.
	// This affects the shadow size relative to the element size.
	Spread float32

	// Color is the shadow color, typically semi-transparent black.
	Color widget.Color
}

// WithAlpha returns a copy of the shadow with a different alpha value.
//
// This is useful for adjusting shadow intensity without changing the color.
func (s Shadow) WithAlpha(alpha float32) Shadow {
	s.Color = s.Color.WithAlpha(alpha)
	return s
}

// WithBlur returns a copy of the shadow with a different blur radius.
func (s Shadow) WithBlur(blur float32) Shadow {
	s.Blur = blur
	return s
}

// WithOffset returns a copy of the shadow with different offsets.
func (s Shadow) WithOffset(x, y float32) Shadow {
	s.OffsetX = x
	s.OffsetY = y
	return s
}

// Scale returns a copy of the shadow with all dimensions scaled.
//
// This scales offset, blur, and spread proportionally.
func (s Shadow) Scale(factor float32) Shadow {
	return Shadow{
		OffsetX: s.OffsetX * factor,
		OffsetY: s.OffsetY * factor,
		Blur:    s.Blur * factor,
		Spread:  s.Spread * factor,
		Color:   s.Color,
	}
}

// IsZero returns true if the shadow has no visible effect.
//
// A shadow is considered zero if it has no blur, no spread, and is
// fully transparent.
func (s Shadow) IsZero() bool {
	return s.Blur <= 0 && s.Spread <= 0 && s.Color.A <= 0
}

// ElevationShadow defines a multi-layered shadow for realistic elevation.
//
// Real shadows often consist of multiple layers: a key light creating a
// crisp shadow, and ambient light creating a soft shadow. This struct
// allows combining multiple shadow layers for more realistic effects.
type ElevationShadow struct {
	// Key is the primary shadow from the key (directional) light.
	//
	// This shadow is typically sharper and more offset.
	Key Shadow

	// Ambient is the secondary shadow from ambient (environmental) light.
	//
	// This shadow is typically softer and less offset.
	Ambient Shadow
}

// WithAlpha returns a copy with both shadow layers adjusted to the given alpha.
func (e ElevationShadow) WithAlpha(alpha float32) ElevationShadow {
	return ElevationShadow{
		Key:     e.Key.WithAlpha(alpha),
		Ambient: e.Ambient.WithAlpha(alpha),
	}
}

// Scale returns a copy with both shadow layers scaled.
func (e ElevationShadow) Scale(factor float32) ElevationShadow {
	return ElevationShadow{
		Key:     e.Key.Scale(factor),
		Ambient: e.Ambient.Scale(factor),
	}
}

// ShadowStyles defines the shadow system for a theme.
//
// ShadowStyles provides elevation-based shadows following Material Design
// principles. Each elevation level represents a different height above the
// surface, with higher elevations creating larger, more diffuse shadows.
//
// Elevation levels:
//   - Level0: No elevation (flat on surface)
//   - Level1: Slight elevation (cards, buttons at rest)
//   - Level2: Medium elevation (hovering buttons, FAB at rest)
//   - Level3: Higher elevation (navigation drawers, menus)
//   - Level4: High elevation (dialogs)
//   - Level5: Highest elevation (temporary surfaces like tooltips)
type ShadowStyles struct {
	// Level0 has no shadow (elevation 0dp).
	Level0 ElevationShadow

	// Level1 is for subtle elevation (1-3dp).
	//
	// Use for cards, buttons at rest, and surfaces at rest state.
	Level1 ElevationShadow

	// Level2 is for medium elevation (4-6dp).
	//
	// Use for hovering buttons, raised surfaces, and FAB at rest.
	Level2 ElevationShadow

	// Level3 is for elevated elements (7-12dp).
	//
	// Use for navigation drawers, bottom sheets, and menus.
	Level3 ElevationShadow

	// Level4 is for high elevation (13-16dp).
	//
	// Use for dialogs and full-screen dialogs.
	Level4 ElevationShadow

	// Level5 is for the highest elevation (17-24dp).
	//
	// Use for tooltips, overlays, and temporary surfaces.
	Level5 ElevationShadow
}

// ForElevation returns the appropriate shadow for a given elevation level (0-5).
//
// If level is out of range, it returns the closest valid level (0 or 5).
func (s *ShadowStyles) ForElevation(level int) ElevationShadow {
	switch {
	case level <= 0:
		return s.Level0
	case level == 1:
		return s.Level1
	case level == 2:
		return s.Level2
	case level == 3:
		return s.Level3
	case level == 4:
		return s.Level4
	default:
		return s.Level5
	}
}

// DefaultShadowsLight returns shadow styles optimized for light themes.
//
// These shadows use semi-transparent black, which works well on light
// backgrounds. The values follow Material Design elevation guidelines.
func DefaultShadowsLight() ShadowStyles {
	keyColor := widget.RGBA(0, 0, 0, 0.14)
	ambientColor := widget.RGBA(0, 0, 0, 0.08)

	return ShadowStyles{
		Level0: ElevationShadow{
			Key:     Shadow{OffsetX: 0, OffsetY: 0, Blur: 0, Spread: 0, Color: widget.ColorTransparent},
			Ambient: Shadow{OffsetX: 0, OffsetY: 0, Blur: 0, Spread: 0, Color: widget.ColorTransparent},
		},
		Level1: ElevationShadow{
			Key:     Shadow{OffsetX: 0, OffsetY: 1, Blur: 3, Spread: 0, Color: keyColor},
			Ambient: Shadow{OffsetX: 0, OffsetY: 1, Blur: 2, Spread: 0, Color: ambientColor},
		},
		Level2: ElevationShadow{
			Key:     Shadow{OffsetX: 0, OffsetY: 2, Blur: 6, Spread: 0, Color: keyColor},
			Ambient: Shadow{OffsetX: 0, OffsetY: 2, Blur: 4, Spread: 0, Color: ambientColor},
		},
		Level3: ElevationShadow{
			Key:     Shadow{OffsetX: 0, OffsetY: 4, Blur: 12, Spread: 0, Color: keyColor},
			Ambient: Shadow{OffsetX: 0, OffsetY: 4, Blur: 8, Spread: 0, Color: ambientColor},
		},
		Level4: ElevationShadow{
			Key:     Shadow{OffsetX: 0, OffsetY: 8, Blur: 16, Spread: 0, Color: keyColor},
			Ambient: Shadow{OffsetX: 0, OffsetY: 8, Blur: 12, Spread: 0, Color: ambientColor},
		},
		Level5: ElevationShadow{
			Key:     Shadow{OffsetX: 0, OffsetY: 12, Blur: 24, Spread: 0, Color: keyColor},
			Ambient: Shadow{OffsetX: 0, OffsetY: 12, Blur: 16, Spread: 0, Color: ambientColor},
		},
	}
}

// DefaultShadowsDark returns shadow styles optimized for dark themes.
//
// Dark themes use stronger shadows to create contrast against dark
// backgrounds. The shadows are also slightly more opaque.
func DefaultShadowsDark() ShadowStyles {
	keyColor := widget.RGBA(0, 0, 0, 0.30)
	ambientColor := widget.RGBA(0, 0, 0, 0.20)

	return ShadowStyles{
		Level0: ElevationShadow{
			Key:     Shadow{OffsetX: 0, OffsetY: 0, Blur: 0, Spread: 0, Color: widget.ColorTransparent},
			Ambient: Shadow{OffsetX: 0, OffsetY: 0, Blur: 0, Spread: 0, Color: widget.ColorTransparent},
		},
		Level1: ElevationShadow{
			Key:     Shadow{OffsetX: 0, OffsetY: 1, Blur: 4, Spread: 0, Color: keyColor},
			Ambient: Shadow{OffsetX: 0, OffsetY: 1, Blur: 3, Spread: 0, Color: ambientColor},
		},
		Level2: ElevationShadow{
			Key:     Shadow{OffsetX: 0, OffsetY: 3, Blur: 8, Spread: 0, Color: keyColor},
			Ambient: Shadow{OffsetX: 0, OffsetY: 3, Blur: 5, Spread: 0, Color: ambientColor},
		},
		Level3: ElevationShadow{
			Key:     Shadow{OffsetX: 0, OffsetY: 5, Blur: 14, Spread: 0, Color: keyColor},
			Ambient: Shadow{OffsetX: 0, OffsetY: 5, Blur: 10, Spread: 0, Color: ambientColor},
		},
		Level4: ElevationShadow{
			Key:     Shadow{OffsetX: 0, OffsetY: 10, Blur: 20, Spread: 0, Color: keyColor},
			Ambient: Shadow{OffsetX: 0, OffsetY: 10, Blur: 14, Spread: 0, Color: ambientColor},
		},
		Level5: ElevationShadow{
			Key:     Shadow{OffsetX: 0, OffsetY: 14, Blur: 28, Spread: 0, Color: keyColor},
			Ambient: Shadow{OffsetX: 0, OffsetY: 14, Blur: 20, Spread: 0, Color: ambientColor},
		},
	}
}
