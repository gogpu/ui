package theme

import (
	"testing"
)

func TestDefaultLight(t *testing.T) {
	th := DefaultLight()

	if th.Name != "Light" {
		t.Errorf("Name = %v, want Light", th.Name)
	}
	if th.Mode != ModeLight {
		t.Errorf("Mode = %v, want ModeLight", th.Mode)
	}
	if th.Extensions == nil {
		t.Error("Extensions should be initialized")
	}

	// Check colors are set
	if th.Colors.Primary.R == 0 && th.Colors.Primary.G == 0 && th.Colors.Primary.B == 0 {
		t.Error("Primary color should not be black")
	}
	if th.Colors.Background.R != 1 || th.Colors.Background.G != 1 || th.Colors.Background.B != 1 {
		t.Error("Background should be white")
	}

	// Check typography
	if th.Typography.FontFamily != "System" {
		t.Errorf("FontFamily = %v, want System", th.Typography.FontFamily)
	}

	// Check spacing
	if th.Spacing.M != 16 {
		t.Errorf("Spacing.M = %v, want 16", th.Spacing.M)
	}

	// Check radii
	if th.Radii.M != 8 {
		t.Errorf("Radii.M = %v, want 8", th.Radii.M)
	}
}

func TestDefaultDark(t *testing.T) {
	th := DefaultDark()

	if th.Name != "Dark" {
		t.Errorf("Name = %v, want Dark", th.Name)
	}
	if th.Mode != ModeDark {
		t.Errorf("Mode = %v, want ModeDark", th.Mode)
	}

	// Background should be dark (low values)
	if th.Colors.Background.R > 0.1 || th.Colors.Background.G > 0.1 || th.Colors.Background.B > 0.1 {
		t.Error("Dark background should be very dark")
	}

	// OnBackground should be light for contrast
	if th.Colors.OnBackground.R < 0.8 {
		t.Error("OnBackground should be light for dark theme")
	}
}

func TestDefaultHighContrast(t *testing.T) {
	th := DefaultHighContrast()

	if th.Name != "High Contrast" {
		t.Errorf("Name = %v, want High Contrast", th.Name)
	}
	if th.Mode != ModeLight {
		t.Errorf("Mode = %v, want ModeLight", th.Mode)
	}

	// Background should be pure white
	if th.Colors.Background.R != 1 || th.Colors.Background.G != 1 || th.Colors.Background.B != 1 {
		t.Error("High contrast background should be pure white")
	}

	// OnBackground should be pure black
	if th.Colors.OnBackground.R != 0 || th.Colors.OnBackground.G != 0 || th.Colors.OnBackground.B != 0 {
		t.Error("High contrast OnBackground should be pure black")
	}

	// Should use relaxed spacing
	defaultSpacing := DefaultSpacing()
	if th.Spacing.M != defaultSpacing.M*1.5 {
		t.Errorf("High contrast spacing should be relaxed: M = %v, want %v",
			th.Spacing.M, defaultSpacing.M*1.5)
	}

	// Should use sharp radii
	sharpRadii := SharpRadii()
	if th.Radii.M != sharpRadii.M {
		t.Errorf("High contrast radii should be sharp: M = %v, want %v",
			th.Radii.M, sharpRadii.M)
	}
}

func TestPurple(t *testing.T) {
	th := Purple()

	if th.Name != "Purple" {
		t.Errorf("Name = %v, want Purple", th.Name)
	}

	// Primary should have high red component (purple)
	// 0x7B1FA2 = RGB(123, 31, 162) -> R ~0.48, G ~0.12, B ~0.64
	if th.Colors.Primary.B < 0.5 {
		t.Error("Purple primary should have high blue component")
	}
}

func TestGreen(t *testing.T) {
	th := Green()

	if th.Name != "Green" {
		t.Errorf("Name = %v, want Green", th.Name)
	}

	// Primary should have high green component
	if th.Colors.Primary.G < 0.4 {
		t.Error("Green primary should have high green component")
	}
}

func TestOrange(t *testing.T) {
	th := Orange()

	if th.Name != "Orange" {
		t.Errorf("Name = %v, want Orange", th.Name)
	}

	// Primary should have high red component
	if th.Colors.Primary.R < 0.8 {
		t.Error("Orange primary should have high red component")
	}

	// OnPrimary should be black for orange
	if th.Colors.OnPrimary.R != 0 {
		t.Error("Orange OnPrimary should be black")
	}
}

func TestForMode(t *testing.T) {
	tests := []struct {
		name     string
		mode     ThemeMode
		wantName string
	}{
		{"light mode", ModeLight, "Light"},
		{"dark mode", ModeDark, "Dark"},
		{"system mode defaults to light", ModeSystem, "Light"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			th := ForMode(tt.mode)
			if th.Name != tt.wantName {
				t.Errorf("ForMode(%v).Name = %v, want %v", tt.mode, th.Name, tt.wantName)
			}
		})
	}
}

func TestPresets_HaveValidExtensions(t *testing.T) {
	themes := []*Theme{
		DefaultLight(),
		DefaultDark(),
		DefaultHighContrast(),
		Purple(),
		Green(),
		Orange(),
	}

	for _, th := range themes {
		if th.Extensions == nil {
			t.Errorf("%s theme has nil Extensions", th.Name)
		}
	}
}

func TestPresets_HaveValidTypography(t *testing.T) {
	themes := []*Theme{
		DefaultLight(),
		DefaultDark(),
		DefaultHighContrast(),
		Purple(),
		Green(),
		Orange(),
	}

	for _, th := range themes {
		if th.Typography.BodyMedium.Size == 0 {
			t.Errorf("%s theme has zero BodyMedium size", th.Name)
		}
		if th.Typography.FontFamily == "" {
			t.Errorf("%s theme has empty FontFamily", th.Name)
		}
	}
}

func TestPresets_HaveValidSpacing(t *testing.T) {
	themes := []*Theme{
		DefaultLight(),
		DefaultDark(),
		Purple(),
		Green(),
		Orange(),
	}

	for _, th := range themes {
		if th.Spacing.M == 0 {
			t.Errorf("%s theme has zero M spacing", th.Name)
		}
		if th.Spacing.S >= th.Spacing.M {
			t.Errorf("%s theme has S >= M spacing", th.Name)
		}
	}
}

func TestPresets_HaveValidRadii(t *testing.T) {
	themes := []*Theme{
		DefaultLight(),
		DefaultDark(),
		Purple(),
		Green(),
		Orange(),
	}

	for _, th := range themes {
		if th.Radii.None != 0 {
			t.Errorf("%s theme None radius should be 0", th.Name)
		}
		if th.Radii.Full != 9999 {
			t.Errorf("%s theme Full radius should be 9999", th.Name)
		}
	}
}

func TestLightVsDark_ContrastingColors(t *testing.T) {
	light := DefaultLight()
	dark := DefaultDark()

	// Light background should be lighter than dark background
	lightBgLum := light.Colors.Background.R + light.Colors.Background.G + light.Colors.Background.B
	darkBgLum := dark.Colors.Background.R + dark.Colors.Background.G + dark.Colors.Background.B

	if lightBgLum <= darkBgLum {
		t.Error("Light theme background should be lighter than dark theme")
	}

	// Light OnBackground should be darker than dark OnBackground
	lightOnBgLum := light.Colors.OnBackground.R + light.Colors.OnBackground.G + light.Colors.OnBackground.B
	darkOnBgLum := dark.Colors.OnBackground.R + dark.Colors.OnBackground.G + dark.Colors.OnBackground.B

	if lightOnBgLum >= darkOnBgLum {
		t.Error("Light theme OnBackground should be darker than dark theme")
	}
}

func TestPresets_ShadowsMatch(t *testing.T) {
	light := DefaultLight()
	dark := DefaultDark()

	// Light theme uses light shadows
	if light.Shadows.Level1.Key.Color.A >= dark.Shadows.Level1.Key.Color.A {
		t.Error("Dark shadows should be stronger than light shadows")
	}
}
