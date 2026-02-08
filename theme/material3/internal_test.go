package material3

import (
	"math"
	"testing"

	"github.com/gogpu/ui/widget"
)

// m3PurpleInternal is the default Material 3 primary seed color.
var m3PurpleInternal = widget.Hex(0x6750A4)

func TestHCTRoundTrip(t *testing.T) {
	tests := []struct {
		name  string
		color widget.Color
	}{
		{"red", widget.RGB(1, 0, 0)},
		{"green", widget.RGB(0, 1, 0)},
		{"blue", widget.RGB(0, 0, 1)},
		{"white", widget.RGB(1, 1, 1)},
		{"black", widget.RGB(0, 0, 0)},
		{"m3 purple", m3PurpleInternal},
		{"gray", widget.RGB(0.5, 0.5, 0.5)},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			h := hctFromColor(tt.color)
			roundTrip := hctToColor(h)

			// Allow some tolerance for floating-point round-trip.
			const tolerance = 0.02
			if !internalColorApproxEqual(tt.color, roundTrip, tolerance) {
				t.Errorf("round trip failed: input=%v, got=%v (hct=%v)",
					tt.color, roundTrip, h)
			}
		})
	}
}

func TestHCTHueExtraction(t *testing.T) {
	// Red should have hue near 0 or 360.
	redHCT := hctFromColor(widget.RGB(1, 0, 0))
	if redHCT.Hue > 10 && redHCT.Hue < 350 {
		t.Errorf("red hue should be near 0/360, got %f", redHCT.Hue)
	}

	// Green should have hue near 120.
	greenHCT := hctFromColor(widget.RGB(0, 1, 0))
	if math.Abs(greenHCT.Hue-120) > 10 {
		t.Errorf("green hue should be near 120, got %f", greenHCT.Hue)
	}

	// Blue should have hue near 240.
	blueHCT := hctFromColor(widget.RGB(0, 0, 1))
	if math.Abs(blueHCT.Hue-240) > 10 {
		t.Errorf("blue hue should be near 240, got %f", blueHCT.Hue)
	}
}

func TestHCTTone(t *testing.T) {
	// Black should have tone near 0.
	blackHCT := hctFromColor(widget.RGB(0, 0, 0))
	if blackHCT.Tone > 1 {
		t.Errorf("black tone should be near 0, got %f", blackHCT.Tone)
	}

	// White should have tone near 100.
	whiteHCT := hctFromColor(widget.RGB(1, 1, 1))
	if whiteHCT.Tone < 99 {
		t.Errorf("white tone should be near 100, got %f", whiteHCT.Tone)
	}
}

func TestNormalizeHue(t *testing.T) {
	tests := []struct {
		input float64
		want  float64
	}{
		{0, 0},
		{360, 0},
		{720, 0},
		{-90, 270},
		{180, 180},
		{450, 90},
	}
	for _, tt := range tests {
		got := normalizeHue(tt.input)
		if math.Abs(got-tt.want) > 0.001 {
			t.Errorf("normalizeHue(%f) = %f, want %f", tt.input, got, tt.want)
		}
	}
}

func TestTonalPalette(t *testing.T) {
	tp := newTonalPalette(270, 0.5)

	// Should have all standard tones.
	for _, tone := range standardTones {
		c := tp.tone(tone)
		if c.A != 1 {
			t.Errorf("tone %d alpha should be 1, got %f", tone, c.A)
		}
	}

	// Tone 0 should be near black.
	t0 := tp.tone(0)
	if internalColorLuminance(t0) > 0.05 {
		t.Errorf("tone 0 should be near black, got R=%f G=%f B=%f",
			t0.R, t0.G, t0.B)
	}

	// Tone 100 should be near white.
	t100 := tp.tone(100)
	if internalColorLuminance(t100) < 0.95 {
		t.Errorf("tone 100 should be near white, got R=%f G=%f B=%f",
			t100.R, t100.G, t100.B)
	}

	// Tones should increase in lightness.
	prevLum := float32(0.0)
	for _, tone := range standardTones {
		c := tp.tone(tone)
		lum := internalColorLuminance(c)
		if tone > 0 && lum < prevLum-0.01 {
			t.Errorf("tone %d luminance %f should be >= previous %f",
				tone, lum, prevLum)
		}
		prevLum = lum
	}
}

func TestTonalPaletteNonStandardTone(t *testing.T) {
	tp := newTonalPalette(180, 0.6)

	// Request a non-standard tone that is not precomputed.
	c := tp.tone(42)
	if c.A != 1 {
		t.Errorf("non-standard tone should have alpha 1, got %f", c.A)
	}
}

func TestCorePalette(t *testing.T) {
	p := newCorePalette(m3PurpleInternal)

	// Primary should preserve seed hue (approximately).
	seedHCT := hctFromColor(m3PurpleInternal)
	if math.Abs(p.Primary.hue-seedHCT.Hue) > 5 {
		t.Errorf("primary hue %f should be close to seed hue %f",
			p.Primary.hue, seedHCT.Hue)
	}

	// Error hue should be near 25 (red).
	if math.Abs(p.Error.hue-25) > 1 {
		t.Errorf("error hue should be near 25, got %f", p.Error.hue)
	}

	// Tertiary hue should be offset from primary.
	tertiaryExpected := normalizeHue(seedHCT.Hue + 60)
	if math.Abs(p.Tertiary.hue-tertiaryExpected) > 1 {
		t.Errorf("tertiary hue %f should be near %f",
			p.Tertiary.hue, tertiaryExpected)
	}
}

func TestLightColorScheme(t *testing.T) {
	cs := Light(m3PurpleInternal)

	// All colors should be non-zero (have some value).
	internalAssertColorNonZero(t, "Primary", cs.Primary)
	internalAssertColorNonZero(t, "OnPrimary", cs.OnPrimary)
	internalAssertColorNonZero(t, "PrimaryContainer", cs.PrimaryContainer)
	internalAssertColorNonZero(t, "OnPrimaryContainer", cs.OnPrimaryContainer)
	internalAssertColorNonZero(t, "Secondary", cs.Secondary)
	internalAssertColorNonZero(t, "Tertiary", cs.Tertiary)
	internalAssertColorNonZero(t, "Error", cs.Error)
	internalAssertColorNonZero(t, "Surface", cs.Surface)
	internalAssertColorNonZero(t, "Background", cs.Background)

	// Light theme surface should be light (high luminance).
	surfaceLum := internalColorLuminance(cs.Surface)
	if surfaceLum < 0.9 {
		t.Errorf("light surface should be light, luminance=%f", surfaceLum)
	}

	// OnSurface should be dark (low luminance).
	onSurfaceLum := internalColorLuminance(cs.OnSurface)
	if onSurfaceLum > 0.3 {
		t.Errorf("light OnSurface should be dark, luminance=%f", onSurfaceLum)
	}
}

func TestDarkColorScheme(t *testing.T) {
	cs := Dark(m3PurpleInternal)

	// All colors should be non-zero.
	internalAssertColorNonZero(t, "Primary", cs.Primary)
	internalAssertColorNonZero(t, "OnPrimary", cs.OnPrimary)
	internalAssertColorNonZero(t, "Surface", cs.Surface)

	// Dark theme surface should be dark (low luminance).
	surfaceLum := internalColorLuminance(cs.Surface)
	if surfaceLum > 0.15 {
		t.Errorf("dark surface should be dark, luminance=%f", surfaceLum)
	}

	// OnSurface should be light (high luminance).
	onSurfaceLum := internalColorLuminance(cs.OnSurface)
	if onSurfaceLum < 0.7 {
		t.Errorf("dark OnSurface should be light, luminance=%f", onSurfaceLum)
	}
}

func TestInternalDifferentSeedsProduceDifferentSchemes(t *testing.T) {
	purple := Light(m3PurpleInternal)
	red := Light(widget.Hex(0xB3261E))
	green := Light(widget.Hex(0x386A20))

	// Primary colors should differ significantly.
	if internalColorApproxEqual(purple.Primary, red.Primary, 0.05) {
		t.Error("purple and red schemes should produce different primaries")
	}
	if internalColorApproxEqual(purple.Primary, green.Primary, 0.05) {
		t.Error("purple and green schemes should produce different primaries")
	}
	if internalColorApproxEqual(red.Primary, green.Primary, 0.05) {
		t.Error("red and green schemes should produce different primaries")
	}
}

func TestContrastOnPrimaryVsPrimaryInternal(t *testing.T) {
	cs := Light(m3PurpleInternal)

	primaryLum := internalColorLuminance(cs.Primary)
	onPrimaryLum := internalColorLuminance(cs.OnPrimary)

	// Contrast ratio should be meaningful. We use a 2.5:1 threshold
	// rather than the full WCAG 3:1 because this HSL-based approximation
	// does not perfectly match the perceptual CAM16 model.
	contrast := internalContrastRatio(primaryLum, onPrimaryLum)
	if contrast < 2.5 {
		t.Errorf("OnPrimary/Primary contrast ratio %f should be >= 2.5",
			contrast)
	}
}

func TestNewThemeInternal(t *testing.T) {
	theme := New(m3PurpleInternal)

	if theme == nil {
		t.Fatal("New() returned nil")
	}

	// Colors should be populated.
	if theme.Colors.Primary.A == 0 {
		t.Error("theme colors should be populated")
	}

	// Typography should have correct font sizes.
	if theme.Typography.BodyMedium.FontSize != 14 {
		t.Errorf("BodyMedium font size should be 14, got %f",
			theme.Typography.BodyMedium.FontSize)
	}

	// Shape should have correct radii.
	if theme.Shape.Medium != 12 {
		t.Errorf("medium shape should be 12, got %f", theme.Shape.Medium)
	}
}

func TestNewDarkThemeInternal(t *testing.T) {
	theme := NewDark(m3PurpleInternal)

	if theme == nil {
		t.Fatal("NewDark() returned nil")
	}

	// Dark surface should be dark.
	surfaceLum := internalColorLuminance(theme.Colors.Surface)
	if surfaceLum > 0.15 {
		t.Errorf("dark theme surface should be dark, luminance=%f", surfaceLum)
	}
}

func TestDefaultTypeScaleInternal(t *testing.T) {
	ts := DefaultTypeScale()

	// Size ordering: Display > Headline > Title > Body > Label.
	if ts.DisplayLarge.FontSize <= ts.HeadlineLarge.FontSize {
		t.Error("DisplayLarge should be larger than HeadlineLarge")
	}
	if ts.HeadlineLarge.FontSize <= ts.TitleLarge.FontSize {
		t.Error("HeadlineLarge should be larger than TitleLarge")
	}
	if ts.TitleLarge.FontSize <= ts.BodyLarge.FontSize {
		t.Error("TitleLarge should be larger than BodyLarge")
	}
	if ts.BodyLarge.FontSize <= ts.LabelSmall.FontSize {
		t.Error("BodyLarge should be larger than LabelSmall")
	}

	// Within each category, Large >= Medium >= Small.
	if ts.DisplayLarge.FontSize <= ts.DisplayMedium.FontSize {
		t.Error("DisplayLarge should be >= DisplayMedium")
	}
	if ts.DisplayMedium.FontSize <= ts.DisplaySmall.FontSize {
		t.Error("DisplayMedium should be >= DisplaySmall")
	}

	// Specific M3 sizes.
	tests := []struct {
		name string
		got  float32
		want float32
	}{
		{"DisplayLarge", ts.DisplayLarge.FontSize, 57},
		{"DisplayMedium", ts.DisplayMedium.FontSize, 45},
		{"DisplaySmall", ts.DisplaySmall.FontSize, 36},
		{"HeadlineLarge", ts.HeadlineLarge.FontSize, 32},
		{"HeadlineMedium", ts.HeadlineMedium.FontSize, 28},
		{"HeadlineSmall", ts.HeadlineSmall.FontSize, 24},
		{"TitleLarge", ts.TitleLarge.FontSize, 22},
		{"TitleMedium", ts.TitleMedium.FontSize, 16},
		{"TitleSmall", ts.TitleSmall.FontSize, 14},
		{"BodyLarge", ts.BodyLarge.FontSize, 16},
		{"BodyMedium", ts.BodyMedium.FontSize, 14},
		{"BodySmall", ts.BodySmall.FontSize, 12},
		{"LabelLarge", ts.LabelLarge.FontSize, 14},
		{"LabelMedium", ts.LabelMedium.FontSize, 12},
		{"LabelSmall", ts.LabelSmall.FontSize, 11},
	}
	for _, tt := range tests {
		if tt.got != tt.want {
			t.Errorf("%s font size = %f, want %f", tt.name, tt.got, tt.want)
		}
	}

	// Line heights should be positive and >= font size.
	if ts.BodyMedium.LineHeight < ts.BodyMedium.FontSize {
		t.Error("line height should be >= font size")
	}
}

func TestDefaultShapeScaleInternal(t *testing.T) {
	ss := DefaultShapeScale()

	tests := []struct {
		name string
		got  float32
		want float32
	}{
		{"None", ss.None, 0},
		{"ExtraSmall", ss.ExtraSmall, 4},
		{"Small", ss.Small, 8},
		{"Medium", ss.Medium, 12},
		{"Large", ss.Large, 16},
		{"ExtraLarge", ss.ExtraLarge, 28},
		{"Full", ss.Full, 9999},
	}
	for _, tt := range tests {
		if tt.got != tt.want {
			t.Errorf("%s = %f, want %f", tt.name, tt.got, tt.want)
		}
	}

	// Values should increase monotonically.
	if ss.None >= ss.ExtraSmall {
		t.Error("None should be < ExtraSmall")
	}
	if ss.ExtraSmall >= ss.Small {
		t.Error("ExtraSmall should be < Small")
	}
	if ss.Small >= ss.Medium {
		t.Error("Small should be < Medium")
	}
	if ss.Medium >= ss.Large {
		t.Error("Medium should be < Large")
	}
	if ss.Large >= ss.ExtraLarge {
		t.Error("Large should be < ExtraLarge")
	}
	if ss.ExtraLarge >= ss.Full {
		t.Error("ExtraLarge should be < Full")
	}
}

func TestSurfaceContainerHierarchyInternal(t *testing.T) {
	cs := Light(m3PurpleInternal)

	// In light scheme, surface containers should decrease in lightness
	// as they get "higher" (more elevated).
	lowest := internalColorLuminance(cs.SurfaceContainerLowest)
	low := internalColorLuminance(cs.SurfaceContainerLow)
	container := internalColorLuminance(cs.SurfaceContainer)
	high := internalColorLuminance(cs.SurfaceContainerHigh)
	highest := internalColorLuminance(cs.SurfaceContainerHighest)

	if lowest < low {
		t.Errorf("SurfaceContainerLowest (%f) should be >= SurfaceContainerLow (%f)",
			lowest, low)
	}
	if low < container {
		t.Errorf("SurfaceContainerLow (%f) should be >= SurfaceContainer (%f)",
			low, container)
	}
	if container < high {
		t.Errorf("SurfaceContainer (%f) should be >= SurfaceContainerHigh (%f)",
			container, high)
	}
	if high < highest {
		t.Errorf("SurfaceContainerHigh (%f) should be >= SurfaceContainerHighest (%f)",
			high, highest)
	}
}

func TestGraySeedColor(t *testing.T) {
	// Achromatic seed should still produce a valid scheme.
	cs := Light(widget.RGB(0.5, 0.5, 0.5))

	internalAssertColorNonZero(t, "Primary", cs.Primary)
	internalAssertColorNonZero(t, "Surface", cs.Surface)
	internalAssertColorNonZero(t, "Error", cs.Error)
}

// --- Test helpers ---

func internalColorApproxEqual(a, b widget.Color, tolerance float32) bool {
	return internalAbsF32(a.R-b.R) < tolerance &&
		internalAbsF32(a.G-b.G) < tolerance &&
		internalAbsF32(a.B-b.B) < tolerance
}

func internalAbsF32(v float32) float32 {
	if v < 0 {
		return -v
	}
	return v
}

func internalColorLuminance(c widget.Color) float32 {
	return 0.299*c.R + 0.587*c.G + 0.114*c.B
}

func internalContrastRatio(l1, l2 float32) float32 {
	lighter := l1
	darker := l2
	if l2 > l1 {
		lighter = l2
		darker = l1
	}
	return (lighter + 0.05) / (darker + 0.05)
}

func internalAssertColorNonZero(t *testing.T, name string, c widget.Color) {
	t.Helper()
	if c.R == 0 && c.G == 0 && c.B == 0 && c.A == 0 {
		t.Errorf("%s should not be zero-value color", name)
	}
}
