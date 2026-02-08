package material3_test

import (
	"testing"

	"github.com/gogpu/ui/theme/material3"
	"github.com/gogpu/ui/widget"
)

// m3Purple is the default Material 3 primary seed color.
var m3Purple = widget.Hex(0x6750A4)

func TestNew(t *testing.T) {
	theme := material3.New(m3Purple)
	if theme == nil {
		t.Fatal("New() returned nil")
	}

	// Colors should be populated (non-zero alpha).
	if theme.Colors.Primary.A == 0 {
		t.Error("Primary color should have non-zero alpha")
	}
	if theme.Colors.OnPrimary.A == 0 {
		t.Error("OnPrimary color should have non-zero alpha")
	}
	if theme.Colors.Surface.A == 0 {
		t.Error("Surface color should have non-zero alpha")
	}

	// Typography should be populated with M3 values.
	if theme.Typography.BodyMedium.FontSize != 14 {
		t.Errorf("BodyMedium font size = %f, want 14",
			theme.Typography.BodyMedium.FontSize)
	}
	if theme.Typography.DisplayLarge.FontSize != 57 {
		t.Errorf("DisplayLarge font size = %f, want 57",
			theme.Typography.DisplayLarge.FontSize)
	}

	// Shape should be populated with M3 values.
	if theme.Shape.Medium != 12 {
		t.Errorf("Shape.Medium = %f, want 12", theme.Shape.Medium)
	}
	if theme.Shape.Full != 9999 {
		t.Errorf("Shape.Full = %f, want 9999", theme.Shape.Full)
	}
}

func TestNewDark(t *testing.T) {
	theme := material3.NewDark(m3Purple)
	if theme == nil {
		t.Fatal("NewDark() returned nil")
	}

	// Dark surface should be dark.
	surfaceLum := 0.299*theme.Colors.Surface.R +
		0.587*theme.Colors.Surface.G +
		0.114*theme.Colors.Surface.B
	if surfaceLum > 0.15 {
		t.Errorf("dark surface luminance = %f, should be < 0.15", surfaceLum)
	}
}

func TestLight(t *testing.T) {
	cs := material3.Light(m3Purple)

	// Light surface should be bright.
	surfaceLum := 0.299*cs.Surface.R + 0.587*cs.Surface.G + 0.114*cs.Surface.B
	if surfaceLum < 0.9 {
		t.Errorf("light surface luminance = %f, should be > 0.9", surfaceLum)
	}

	// All primary group colors should be populated.
	assertNonZero(t, "Primary", cs.Primary)
	assertNonZero(t, "OnPrimary", cs.OnPrimary)
	assertNonZero(t, "PrimaryContainer", cs.PrimaryContainer)
	assertNonZero(t, "OnPrimaryContainer", cs.OnPrimaryContainer)
}

func TestDark(t *testing.T) {
	cs := material3.Dark(m3Purple)

	// Dark surface should be dim.
	surfaceLum := 0.299*cs.Surface.R + 0.587*cs.Surface.G + 0.114*cs.Surface.B
	if surfaceLum > 0.15 {
		t.Errorf("dark surface luminance = %f, should be < 0.15", surfaceLum)
	}

	assertNonZero(t, "Primary", cs.Primary)
	assertNonZero(t, "OnPrimary", cs.OnPrimary)
}

func TestDifferentSeedsProduceDifferentSchemes(t *testing.T) {
	purple := material3.Light(m3Purple)
	red := material3.Light(widget.Hex(0xB3261E))
	blue := material3.Light(widget.Hex(0x0061A4))

	if colorEqual(purple.Primary, red.Primary) {
		t.Error("purple and red seeds should produce different primaries")
	}
	if colorEqual(purple.Primary, blue.Primary) {
		t.Error("purple and blue seeds should produce different primaries")
	}
}

func TestDefaultTypeScale(t *testing.T) {
	ts := material3.DefaultTypeScale()

	// Check size hierarchy.
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

	// Check specific M3 font sizes.
	if ts.DisplayLarge.FontSize != 57 {
		t.Errorf("DisplayLarge = %f, want 57", ts.DisplayLarge.FontSize)
	}
	if ts.LabelSmall.FontSize != 11 {
		t.Errorf("LabelSmall = %f, want 11", ts.LabelSmall.FontSize)
	}
}

func TestDefaultShapeScale(t *testing.T) {
	ss := material3.DefaultShapeScale()

	if ss.None != 0 {
		t.Errorf("None = %f, want 0", ss.None)
	}
	if ss.ExtraSmall != 4 {
		t.Errorf("ExtraSmall = %f, want 4", ss.ExtraSmall)
	}
	if ss.Small != 8 {
		t.Errorf("Small = %f, want 8", ss.Small)
	}
	if ss.Medium != 12 {
		t.Errorf("Medium = %f, want 12", ss.Medium)
	}
	if ss.Large != 16 {
		t.Errorf("Large = %f, want 16", ss.Large)
	}
	if ss.ExtraLarge != 28 {
		t.Errorf("ExtraLarge = %f, want 28", ss.ExtraLarge)
	}
	if ss.Full != 9999 {
		t.Errorf("Full = %f, want 9999", ss.Full)
	}
}

func TestContrastOnPrimaryVsPrimary(t *testing.T) {
	cs := material3.Light(m3Purple)

	primaryLum := 0.299*cs.Primary.R + 0.587*cs.Primary.G + 0.114*cs.Primary.B
	onPrimaryLum := 0.299*cs.OnPrimary.R + 0.587*cs.OnPrimary.G + 0.114*cs.OnPrimary.B

	// Calculate simplified contrast ratio.
	lighter := primaryLum
	darker := onPrimaryLum
	if onPrimaryLum > primaryLum {
		lighter = onPrimaryLum
		darker = primaryLum
	}
	contrast := (lighter + 0.05) / (darker + 0.05)

	// OnPrimary should provide meaningful contrast. We use 2.5:1 threshold
	// because this HSL-based approximation does not perfectly match the
	// perceptual CAM16 model used by the full M3 spec.
	if contrast < 2.5 {
		t.Errorf("OnPrimary/Primary contrast = %f, want >= 2.5", contrast)
	}
}

func TestSurfaceContainerHierarchy(t *testing.T) {
	cs := material3.Light(m3Purple)

	lum := func(c widget.Color) float32 {
		return 0.299*c.R + 0.587*c.G + 0.114*c.B
	}

	lowest := lum(cs.SurfaceContainerLowest)
	low := lum(cs.SurfaceContainerLow)
	container := lum(cs.SurfaceContainer)
	high := lum(cs.SurfaceContainerHigh)
	highest := lum(cs.SurfaceContainerHighest)

	// In light scheme, higher container = slightly darker (lower luminance).
	if lowest < low {
		t.Errorf("Lowest (%f) should be >= Low (%f)", lowest, low)
	}
	if low < container {
		t.Errorf("Low (%f) should be >= Container (%f)", low, container)
	}
	if container < high {
		t.Errorf("Container (%f) should be >= High (%f)", container, high)
	}
	if high < highest {
		t.Errorf("High (%f) should be >= Highest (%f)", high, highest)
	}
}

func TestPublicTypesUsable(t *testing.T) {
	// Verify public types are usable and properly exported.
	theme := material3.New(m3Purple)
	_ = theme.Colors.Primary
	_ = theme.Typography.BodyMedium.FontSize
	_ = theme.Shape.Medium

	cs := material3.Light(m3Purple)
	_ = cs.Primary

	ts := material3.DefaultTypeScale()
	_ = ts.DisplayLarge

	ss := material3.DefaultShapeScale()
	_ = ss.Medium

	style := ts.BodyMedium
	_ = style.FontSize
}

// --- Test helpers ---

func assertNonZero(t *testing.T, name string, c widget.Color) {
	t.Helper()
	if c.R == 0 && c.G == 0 && c.B == 0 && c.A == 0 {
		t.Errorf("%s should not be zero-value color", name)
	}
}

func colorEqual(a, b widget.Color) bool {
	const tolerance = 0.05
	return absF32(a.R-b.R) < tolerance &&
		absF32(a.G-b.G) < tolerance &&
		absF32(a.B-b.B) < tolerance
}

func absF32(v float32) float32 {
	if v < 0 {
		return -v
	}
	return v
}
