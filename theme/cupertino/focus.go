package cupertino

import (
	"math"

	"github.com/gogpu/ui/widget"
)

// cupFocusColor returns the macOS keyboard focus ring color for an accent:
// the accent hue with fixed saturation and lightness at 80% opacity. The
// constants are the ones Flutter's Cupertino controls measured from macOS.
func cupFocusColor(accent widget.Color) widget.Color {
	r, g, b := cupHSLToRGB(cupHue(accent), cupFocusSaturation, cupFocusLightness)
	return widget.RGBA(r, g, b, accent.A*cupFocusOpacity)
}

// cupHue returns the HSL hue of c in degrees, in [0, 360).
func cupHue(c widget.Color) float32 {
	hi := max(c.R, c.G, c.B)
	d := hi - min(c.R, c.G, c.B)
	if d == 0 {
		return 0
	}
	var h float32
	switch hi {
	case c.R:
		h = (c.G - c.B) / d
		if h < 0 {
			h += 6
		}
	case c.G:
		h = (c.B-c.R)/d + 2
	default:
		h = (c.R-c.G)/d + 4
	}
	return h * 60
}

// cupHSLToRGB converts an HSL color (hue in degrees) to RGB components.
func cupHSLToRGB(h, s, l float32) (r, g, b float32) {
	a := s * min(l, 1-l)
	f := func(n float32) float32 {
		k := float32(math.Mod(float64(n+h/30), 12))
		return l - a*max(-1, min(k-3, 9-k, 1))
	}
	return f(0), f(8), f(4)
}

// macOS focus ring color constants.
const (
	cupFocusOpacity    float32 = 0.8
	cupFocusLightness  float32 = 0.69
	cupFocusSaturation float32 = 0.835
)
