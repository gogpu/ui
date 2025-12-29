package render

import (
	"github.com/gogpu/gg"
	"github.com/gogpu/ui/widget"
)

// ToGGColor converts a widget.Color to gg.RGBA.
//
// Both types use float components in the range [0, 1], but widget.Color uses
// float32 while gg.RGBA uses float64. This function performs the conversion.
func ToGGColor(c widget.Color) gg.RGBA {
	return gg.RGBA{
		R: float64(c.R),
		G: float64(c.G),
		B: float64(c.B),
		A: float64(c.A),
	}
}

// FromGGColor converts a gg.RGBA to widget.Color.
//
// This is the inverse of [ToGGColor].
func FromGGColor(c gg.RGBA) widget.Color {
	return widget.Color{
		R: float32(c.R),
		G: float32(c.G),
		B: float32(c.B),
		A: float32(c.A),
	}
}

// ToGGColorPremultiplied converts a widget.Color to premultiplied gg.RGBA.
//
// Premultiplied alpha is more efficient for blending operations.
// The RGB components are multiplied by alpha.
func ToGGColorPremultiplied(c widget.Color) gg.RGBA {
	return gg.RGBA{
		R: float64(c.R * c.A),
		G: float64(c.G * c.A),
		B: float64(c.B * c.A),
		A: float64(c.A),
	}
}

// Clamp01Float64 clamps a float64 value to the range [0, 1].
func Clamp01Float64(v float64) float64 {
	if v < 0 {
		return 0
	}
	if v > 1 {
		return 1
	}
	return v
}

// Clamp01Float32 clamps a float32 value to the range [0, 1].
func Clamp01Float32(v float32) float32 {
	if v < 0 {
		return 0
	}
	if v > 1 {
		return 1
	}
	return v
}

// LerpColor linearly interpolates between two colors.
//
// t=0 returns a, t=1 returns b, values in between interpolate.
// This is a convenience function that operates on gg.RGBA directly.
func LerpColor(a, b gg.RGBA, t float64) gg.RGBA {
	return gg.RGBA{
		R: a.R + (b.R-a.R)*t,
		G: a.G + (b.G-a.G)*t,
		B: a.B + (b.B-a.B)*t,
		A: a.A + (b.A-a.A)*t,
	}
}
