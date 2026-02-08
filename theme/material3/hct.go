package material3

import (
	"math"

	"github.com/gogpu/ui/widget"
)

// hct represents a color in the HCT (Hue, Chroma, Tone) color space.
//
// HCT is the color space used by Material Design 3 for tonal palette generation.
// This is a practical approximation using HSL-based tonal mapping rather than
// the full CAM16 perceptual model.
//
//   - Hue: 0-360 degrees (color wheel position)
//   - Chroma: 0-1 (colorfulness / saturation)
//   - Tone: 0-100 (lightness, where 0=black, 100=white)
type hct struct {
	Hue    float64
	Chroma float64
	Tone   float64
}

// hctFromColor converts a widget.Color to HCT representation.
func hctFromColor(c widget.Color) hct {
	h, s, l := rgbToHSL(float64(c.R), float64(c.G), float64(c.B))
	return hct{
		Hue:    h,
		Chroma: s,
		Tone:   l * 100,
	}
}

// hctToColor converts an HCT value to a widget.Color.
func hctToColor(h hct) widget.Color {
	r, g, b := hslToRGB(h.Hue, h.Chroma, h.Tone/100)
	return widget.RGB(float32(r), float32(g), float32(b))
}

// withTone returns a new HCT value with the given tone (0-100).
func (h hct) withTone(tone float64) hct {
	return hct{
		Hue:    h.Hue,
		Chroma: h.Chroma,
		Tone:   tone,
	}
}

// rgbToHSL converts RGB (0-1 range) to HSL.
// Returns hue in degrees (0-360), saturation (0-1), lightness (0-1).
func rgbToHSL(r, g, b float64) (h, s, l float64) {
	maxC := math.Max(r, math.Max(g, b))
	minC := math.Min(r, math.Min(g, b))
	l = (maxC + minC) / 2

	if maxC == minC {
		// Achromatic
		return 0, 0, l
	}

	d := maxC - minC
	if l > 0.5 {
		s = d / (2 - maxC - minC)
	} else {
		s = d / (maxC + minC)
	}

	switch maxC {
	case r:
		h = (g - b) / d
		if g < b {
			h += 6
		}
	case g:
		h = (b-r)/d + 2
	case b:
		h = (r-g)/d + 4
	}
	h *= 60

	return h, s, l
}

// hslToRGB converts HSL to RGB (0-1 range).
// Hue is in degrees (0-360), saturation and lightness are 0-1.
func hslToRGB(h, s, l float64) (r, g, b float64) {
	if s == 0 {
		return l, l, l
	}

	var q float64
	if l < 0.5 {
		q = l * (1 + s)
	} else {
		q = l + s - l*s
	}
	p := 2*l - q

	h /= 360

	r = hueToRGB(p, q, h+1.0/3.0)
	g = hueToRGB(p, q, h)
	b = hueToRGB(p, q, h-1.0/3.0)

	return r, g, b
}

// hueToRGB is a helper for HSL to RGB conversion.
func hueToRGB(p, q, t float64) float64 {
	if t < 0 {
		t++
	}
	if t > 1 {
		t--
	}
	if t < 1.0/6.0 {
		return p + (q-p)*6*t
	}
	if t < 0.5 {
		return q
	}
	if t < 2.0/3.0 {
		return p + (q-p)*(2.0/3.0-t)*6
	}
	return p
}

// normalizeHue wraps a hue value to the 0-360 range.
func normalizeHue(hue float64) float64 {
	hue = math.Mod(hue, 360)
	if hue < 0 {
		hue += 360
	}
	return hue
}
