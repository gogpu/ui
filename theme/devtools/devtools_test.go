package devtools_test

import (
	"testing"

	"github.com/gogpu/ui/theme/devtools"
	"github.com/gogpu/ui/widget"
)

func TestNewTheme(t *testing.T) {
	dt := devtools.NewTheme()
	if dt == nil {
		t.Fatal("NewTheme() returned nil")
	}
	if dt.IsDark() {
		t.Error("NewTheme() should return light theme")
	}
	if dt.Colors.Primary.A == 0 {
		t.Error("Primary color should have non-zero alpha")
	}
	if dt.Colors.Surface.A == 0 {
		t.Error("Surface color should have non-zero alpha")
	}
}

func TestNewDarkTheme(t *testing.T) {
	dt := devtools.NewDarkTheme()
	if dt == nil {
		t.Fatal("NewDarkTheme() returned nil")
	}
	if !dt.IsDark() {
		t.Error("NewDarkTheme() should return dark theme")
	}

	// Dark surface should be dark (low luminance).
	surfaceLum := luminance(dt.Colors.Surface)
	if surfaceLum > 0.2 {
		t.Errorf("dark surface luminance = %f, should be < 0.2", surfaceLum)
	}
}

func TestNewThemeWithAccentColor(t *testing.T) {
	green := widget.Hex(0x57965C)
	dt := devtools.NewDarkTheme(devtools.WithAccentColor(green))

	if dt.Colors.Primary != green {
		t.Errorf("accent = %v, want %v", dt.Colors.Primary, green)
	}
	if dt.Colors.BorderFocus != green {
		t.Error("BorderFocus should match custom accent")
	}
	if dt.Colors.Info != green {
		t.Error("Info should match custom accent")
	}
}

func TestThemeOnSurface(t *testing.T) {
	dt := devtools.NewDarkTheme()
	onSurface := dt.OnSurface()
	if onSurface.A == 0 {
		t.Error("OnSurface should have non-zero alpha")
	}
	if onSurface != dt.Colors.OnSurface {
		t.Error("OnSurface() should equal Colors.OnSurface")
	}
}

func TestThemeImplementsThemeProvider(t *testing.T) {
	var _ widget.ThemeProvider = (*devtools.Theme)(nil)
}

func TestDarkScheme(t *testing.T) {
	cs := devtools.DarkScheme()
	assertNonZero(t, "Background", cs.Background)
	assertNonZero(t, "Surface", cs.Surface)
	assertNonZero(t, "OnSurface", cs.OnSurface)
	assertNonZero(t, "Primary", cs.Primary)
	assertNonZero(t, "Error", cs.Error)
	assertNonZero(t, "Warning", cs.Warning)
	assertNonZero(t, "Success", cs.Success)
	assertNonZero(t, "HeaderBackground", cs.HeaderBackground)

	// Dark surface luminance check.
	surfaceLum := luminance(cs.Surface)
	if surfaceLum > 0.2 {
		t.Errorf("dark surface luminance = %f, should be < 0.2", surfaceLum)
	}

	// Primary should be blue (#3574F0).
	if cs.Primary != devtools.DefaultAccentColor {
		t.Errorf("Primary = %v, want DefaultAccentColor", cs.Primary)
	}
}

func TestLightScheme(t *testing.T) {
	cs := devtools.LightScheme()
	assertNonZero(t, "Surface", cs.Surface)
	assertNonZero(t, "OnSurface", cs.OnSurface)
	assertNonZero(t, "Primary", cs.Primary)
	assertNonZero(t, "Error", cs.Error)

	// Light background should be bright.
	bgLum := luminance(cs.Background)
	if bgLum < 0.9 {
		t.Errorf("light background luminance = %f, should be > 0.9", bgLum)
	}

	// Light theme header should still be dark.
	headerLum := luminance(cs.HeaderBackground)
	if headerLum > 0.2 {
		t.Errorf("light header luminance = %f, should be < 0.2 (dark toolbar)", headerLum)
	}
}

func TestAsThemeDark(t *testing.T) {
	dt := devtools.NewDarkTheme()
	generic := dt.AsTheme()
	if generic == nil {
		t.Fatal("AsTheme() returned nil")
	}
	if generic.Name != "DevTools Dark" {
		t.Errorf("Name = %q, want %q", generic.Name, "DevTools Dark")
	}
	if generic.Colors.Primary.A == 0 {
		t.Error("Primary should have non-zero alpha")
	}
	if generic.Colors.Background.A == 0 {
		t.Error("Background should have non-zero alpha")
	}
	if generic.Extensions == nil {
		t.Error("Extensions map should be initialized")
	}
}

func TestAsThemeLight(t *testing.T) {
	dt := devtools.NewTheme()
	generic := dt.AsTheme()
	if generic.Name != "DevTools Light" {
		t.Errorf("Name = %q, want %q", generic.Name, "DevTools Light")
	}
}

func TestAsThemeTypography(t *testing.T) {
	generic := devtools.NewDarkTheme().AsTheme()

	// Body medium should be 13px (DevTools base).
	if generic.Typography.BodyMedium.Size != 13 {
		t.Errorf("BodyMedium.Size = %f, want 13", generic.Typography.BodyMedium.Size)
	}

	// Label medium should be 12px.
	if generic.Typography.LabelMedium.Size != 12 {
		t.Errorf("LabelMedium.Size = %f, want 12", generic.Typography.LabelMedium.Size)
	}
}

func TestAsThemeSpacing(t *testing.T) {
	generic := devtools.NewDarkTheme().AsTheme()

	// DevTools spacing is more compact than M3 defaults.
	if generic.Spacing.M != 8 {
		t.Errorf("Spacing.M = %f, want 8 (compact)", generic.Spacing.M)
	}
	if generic.Spacing.S != 6 {
		t.Errorf("Spacing.S = %f, want 6", generic.Spacing.S)
	}
}

func TestAsThemeRadii(t *testing.T) {
	generic := devtools.NewDarkTheme().AsTheme()

	if generic.Radii.S != 4 {
		t.Errorf("Radii.S = %f, want 4", generic.Radii.S)
	}
	if generic.Radii.Full != 9999 {
		t.Errorf("Radii.Full = %f, want 9999", generic.Radii.Full)
	}
}

func TestDifferentAccentsProduceDifferentSchemes(t *testing.T) {
	blue := devtools.NewDarkTheme()
	green := devtools.NewDarkTheme(devtools.WithAccentColor(widget.Hex(0x57965C)))

	if colorEqual(blue.Colors.Primary, green.Colors.Primary) {
		t.Error("blue and green accents should produce different Primary colors")
	}
	if colorEqual(blue.Colors.PrimaryHover, green.Colors.PrimaryHover) {
		t.Error("blue and green accents should produce different PrimaryHover colors")
	}
}

func TestAccentColorAffectsLightTheme(t *testing.T) {
	red := widget.Hex(0xDB5C5C)
	dt := devtools.NewTheme(devtools.WithAccentColor(red))

	if dt.Colors.Primary != red {
		t.Errorf("light accent = %v, want %v", dt.Colors.Primary, red)
	}
}

func TestDarkSchemeScrollbar(t *testing.T) {
	cs := devtools.DarkScheme()
	assertNonZero(t, "ScrollbarThumb", cs.ScrollbarThumb)
	assertNonZero(t, "ScrollbarThumbHover", cs.ScrollbarThumbHover)
	if cs.ScrollbarTrack != widget.ColorTransparent {
		t.Error("ScrollbarTrack should be transparent")
	}
}

func TestDarkSchemeSemanticColors(t *testing.T) {
	cs := devtools.DarkScheme()

	// Error should be reddish (R > G and R > B).
	if cs.Error.R <= cs.Error.G || cs.Error.R <= cs.Error.B {
		t.Error("Error color should be reddish")
	}

	// Success should be greenish (G > R).
	if cs.Success.G <= cs.Success.R {
		t.Error("Success color should be greenish")
	}

	// Warning should be yellowish (R > B and G > B).
	if cs.Warning.R <= cs.Warning.B || cs.Warning.G <= cs.Warning.B {
		t.Error("Warning color should be yellowish")
	}
}

func TestLightSchemeSelection(t *testing.T) {
	cs := devtools.LightScheme()
	assertNonZero(t, "Selection", cs.Selection)

	// Light selection should be lighter than dark selection.
	darkCS := devtools.DarkScheme()
	lightLum := luminance(cs.Selection)
	darkLum := luminance(darkCS.Selection)
	if lightLum <= darkLum {
		t.Errorf("light selection luminance (%f) should be > dark (%f)", lightLum, darkLum)
	}
}

// --- Helpers ---

func assertNonZero(t *testing.T, name string, c widget.Color) {
	t.Helper()
	if c.R == 0 && c.G == 0 && c.B == 0 && c.A == 0 {
		t.Errorf("%s should not be zero-value color", name)
	}
}

func luminance(c widget.Color) float32 {
	return 0.299*c.R + 0.587*c.G + 0.114*c.B
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
