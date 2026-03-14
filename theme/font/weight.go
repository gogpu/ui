package font

import "fmt"

// Weight represents the weight (boldness) of a font.
//
// Font weights follow the CSS/OpenType standard where values range from 100
// (thinnest) to 900 (heaviest). Not all fonts support all weights; the
// registry uses a CSS-standard matching algorithm to find the closest
// available weight.
type Weight int

const (
	// Thin is the thinnest available weight (100).
	Thin Weight = 100

	// ExtraLight is extra-light weight (200).
	ExtraLight Weight = 200

	// Light is light weight (300).
	Light Weight = 300

	// Regular is the standard weight (400).
	Regular Weight = 400

	// Medium is medium weight (500).
	Medium Weight = 500

	// SemiBold is semi-bold weight (600).
	SemiBold Weight = 600

	// Bold is bold weight (700).
	Bold Weight = 700

	// ExtraBold is extra-bold weight (800).
	ExtraBold Weight = 800

	// Black is the heaviest weight (900).
	Black Weight = 900
)

// weightNames maps standard weight values to their human-readable names.
var weightNames = map[Weight]string{
	Thin:       "Thin",
	ExtraLight: "ExtraLight",
	Light:      "Light",
	Regular:    "Regular",
	Medium:     "Medium",
	SemiBold:   "SemiBold",
	Bold:       "Bold",
	ExtraBold:  "ExtraBold",
	Black:      "Black",
}

// String returns the human-readable name for the font weight.
// For non-standard values, it returns the numeric representation.
func (w Weight) String() string {
	if name, ok := weightNames[w]; ok {
		return name
	}
	return fmt.Sprintf("Weight(%d)", int(w))
}

// IsBold returns true if this weight is considered bold (>= 700).
func (w Weight) IsBold() bool {
	return w >= Bold
}

// IsLight returns true if this weight is considered light (<= 300).
func (w Weight) IsLight() bool {
	return w <= Light
}
