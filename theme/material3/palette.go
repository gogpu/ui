package material3

import "github.com/gogpu/ui/widget"

// tonalPalette holds a set of tones for a single hue+chroma pair.
//
// A tonal palette contains colors at specific lightness levels (tones),
// ranging from 0 (black) to 100 (white). Material Design 3 uses these
// palettes to derive color roles for light and dark schemes.
type tonalPalette struct {
	hue    float64
	chroma float64
	tones  map[int]widget.Color
}

// standardTones are the tone values used by Material Design 3.
var standardTones = []int{
	0, 5, 10, 15, 20, 25, 30, 35, 40, 50,
	60, 70, 80, 90, 95, 98, 99, 100,
}

// newTonalPalette generates a tonal palette for the given hue and chroma.
func newTonalPalette(hue, chroma float64) tonalPalette {
	tp := tonalPalette{
		hue:    hue,
		chroma: chroma,
		tones:  make(map[int]widget.Color, len(standardTones)),
	}

	base := hct{Hue: hue, Chroma: chroma}
	for _, tone := range standardTones {
		tp.tones[tone] = hctToColor(base.withTone(float64(tone)))
	}

	return tp
}

// tone returns the color at the given tone level.
// If the exact tone is not precomputed, it generates it on the fly.
func (tp tonalPalette) tone(t int) widget.Color {
	if c, ok := tp.tones[t]; ok {
		return c
	}
	// Generate on-the-fly for non-standard tones.
	base := hct{Hue: tp.hue, Chroma: tp.chroma}
	return hctToColor(base.withTone(float64(t)))
}

// corePalette holds the five key tonal palettes used by Material Design 3
// to derive color schemes.
type corePalette struct {
	Primary   tonalPalette
	Secondary tonalPalette
	Tertiary  tonalPalette
	Neutral   tonalPalette
	Error     tonalPalette
}

// newCorePalette creates a core palette from a seed color.
//
// The seed color determines the primary hue. Secondary and tertiary
// hues are derived by rotating the hue wheel. Neutral uses reduced
// chroma for surface colors. Error uses a fixed red hue.
func newCorePalette(seed widget.Color) corePalette {
	seedHCT := hctFromColor(seed)

	primaryHue := seedHCT.Hue
	primaryChroma := seedHCT.Chroma

	// Ensure minimum chroma for visible color.
	if primaryChroma < 0.1 {
		primaryChroma = 0.1
	}

	// Secondary: same hue, reduced chroma for a more muted look.
	secondaryChroma := primaryChroma * 0.33
	if secondaryChroma < 0.05 {
		secondaryChroma = 0.05
	}

	// Tertiary: complementary hue (+60 degrees), moderate chroma.
	tertiaryHue := normalizeHue(primaryHue + 60)
	tertiaryChroma := primaryChroma * 0.75
	if tertiaryChroma < 0.1 {
		tertiaryChroma = 0.1
	}

	// Neutral: primary hue with very low chroma for surfaces.
	neutralChroma := primaryChroma * 0.04
	if neutralChroma < 0.01 {
		neutralChroma = 0.01
	}

	// Error: fixed red hue (25 degrees).
	const errorHue = 25.0
	const errorChroma = 0.85

	return corePalette{
		Primary:   newTonalPalette(primaryHue, primaryChroma),
		Secondary: newTonalPalette(primaryHue, secondaryChroma),
		Tertiary:  newTonalPalette(tertiaryHue, tertiaryChroma),
		Neutral:   newTonalPalette(primaryHue, neutralChroma),
		Error:     newTonalPalette(errorHue, errorChroma),
	}
}
