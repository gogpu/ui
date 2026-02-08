package theme

import (
	"testing"

	"github.com/gogpu/ui/widget"
)

func TestNew(t *testing.T) {
	th := New("Test Theme", ModeLight)

	if th.Name != "Test Theme" {
		t.Errorf("Name = %v, want Test Theme", th.Name)
	}
	if th.Mode != ModeLight {
		t.Errorf("Mode = %v, want ModeLight", th.Mode)
	}
	if th.Extensions == nil {
		t.Error("Extensions should be initialized")
	}
	if th.Typography.FontFamily != "System" {
		t.Errorf("Typography.FontFamily = %v, want System", th.Typography.FontFamily)
	}
}

func TestTheme_Clone(t *testing.T) {
	original := DefaultLight()
	original.SetExtension("test", "value")

	clone := original.Clone()

	// Values should match
	if clone.Name != original.Name {
		t.Errorf("Clone Name = %v, want %v", clone.Name, original.Name)
	}
	if clone.Mode != original.Mode {
		t.Errorf("Clone Mode = %v, want %v", clone.Mode, original.Mode)
	}

	// Extensions should be copied
	if v, ok := clone.GetExtension("test"); !ok || v != "value" {
		t.Error("Extension not cloned properly")
	}

	// Modifications to clone should not affect original
	clone.Name = "Modified"
	if original.Name == "Modified" {
		t.Error("Modifying clone affected original")
	}

	clone.SetExtension("test", "modified")
	if v, _ := original.GetExtension("test"); v == "modified" {
		t.Error("Modifying clone extension affected original")
	}
}

func TestTheme_Clone_Nil(t *testing.T) {
	var th *Theme
	clone := th.Clone()

	if clone != nil {
		t.Error("Clone of nil should be nil")
	}
}

func TestTheme_WithName(t *testing.T) {
	original := DefaultLight()
	result := original.WithName("New Name")

	if result.Name != "New Name" {
		t.Errorf("Name = %v, want New Name", result.Name)
	}
	if original.Name != "Light" {
		t.Errorf("Original Name changed: %v", original.Name)
	}
}

func TestTheme_WithMode(t *testing.T) {
	original := DefaultLight()
	result := original.WithMode(ModeDark)

	if result.Mode != ModeDark {
		t.Errorf("Mode = %v, want ModeDark", result.Mode)
	}
	if original.Mode != ModeLight {
		t.Errorf("Original Mode changed: %v", original.Mode)
	}
}

func TestTheme_WithColors(t *testing.T) {
	original := DefaultLight()
	newColors := &ColorPalette{
		Primary: widget.Hex(0xFF0000),
	}
	result := original.WithColors(newColors)

	if result.Colors.Primary != newColors.Primary {
		t.Error("Colors not updated")
	}
	if original.Colors.Primary == newColors.Primary {
		t.Error("Original Colors changed")
	}
}

func TestTheme_WithTypography(t *testing.T) {
	original := DefaultLight()
	typo := DefaultTypography()
	newTypo := typo.WithFontFamily("Roboto")
	result := original.WithTypography(&newTypo)

	if result.Typography.FontFamily != "Roboto" {
		t.Errorf("Typography not updated: %v", result.Typography.FontFamily)
	}
	if original.Typography.FontFamily != "System" {
		t.Error("Original Typography changed")
	}
}

func TestTheme_WithSpacing(t *testing.T) {
	original := DefaultLight()
	newSpacing := DenseSpacing()
	result := original.WithSpacing(newSpacing)

	if result.Spacing.M != 8 {
		t.Errorf("Spacing not updated: M = %v", result.Spacing.M)
	}
	if original.Spacing.M != 16 {
		t.Error("Original Spacing changed")
	}
}

func TestTheme_WithShadows(t *testing.T) {
	original := DefaultLight()
	newShadows := DefaultShadowsDark()
	result := original.WithShadows(&newShadows)

	// Dark shadows have higher alpha
	if result.Shadows.Level1.Key.Color.A <= 0.14 {
		t.Error("Shadows not updated")
	}
}

func TestTheme_WithRadii(t *testing.T) {
	original := DefaultLight()
	newRadii := SharpRadii()
	result := original.WithRadii(newRadii)

	if result.Radii.M != 3 {
		t.Errorf("Radii not updated: M = %v", result.Radii.M)
	}
	if original.Radii.M != 8 {
		t.Error("Original Radii changed")
	}
}

func TestTheme_SetExtension(t *testing.T) {
	th := DefaultLight()

	th.SetExtension("key1", "value1")

	v, ok := th.GetExtension("key1")
	if !ok {
		t.Error("Extension not found")
	}
	if v != "value1" {
		t.Errorf("Extension value = %v, want value1", v)
	}
}

func TestTheme_GetExtension_NotFound(t *testing.T) {
	th := DefaultLight()

	v, ok := th.GetExtension("nonexistent")
	if ok {
		t.Error("Should not find nonexistent extension")
	}
	if v != nil {
		t.Error("Value should be nil for nonexistent extension")
	}
}

func TestTheme_SetExtension_NilExtensions(t *testing.T) {
	th := &Theme{}

	// Should initialize Extensions map
	th.SetExtension("key", "value")

	if th.Extensions == nil {
		t.Error("Extensions should be initialized")
	}
	if v, _ := th.GetExtension("key"); v != "value" {
		t.Error("Extension not set properly")
	}
}

func TestTheme_GetExtension_NilExtensions(t *testing.T) {
	th := &Theme{}

	v, ok := th.GetExtension("key")
	if ok {
		t.Error("Should not find extension in nil map")
	}
	if v != nil {
		t.Error("Value should be nil")
	}
}

func TestTheme_IsLight(t *testing.T) {
	light := DefaultLight()
	dark := DefaultDark()

	if !light.IsLight() {
		t.Error("Light theme should return true for IsLight")
	}
	if dark.IsLight() {
		t.Error("Dark theme should return false for IsLight")
	}
}

func TestTheme_IsDark(t *testing.T) {
	light := DefaultLight()
	dark := DefaultDark()

	if light.IsDark() {
		t.Error("Light theme should return false for IsDark")
	}
	if !dark.IsDark() {
		t.Error("Dark theme should return true for IsDark")
	}
}

func TestTheme_OnSurface(t *testing.T) {
	light := DefaultLight()
	onSurface := light.OnSurface()
	if onSurface != light.Colors.OnSurface {
		t.Errorf("OnSurface() = %+v, want Colors.OnSurface = %+v",
			onSurface, light.Colors.OnSurface)
	}
	if onSurface.A == 0 {
		t.Error("OnSurface should have non-zero alpha")
	}

	dark := DefaultDark()
	darkOnSurface := dark.OnSurface()
	if darkOnSurface != dark.Colors.OnSurface {
		t.Errorf("dark OnSurface() = %+v, want Colors.OnSurface = %+v",
			darkOnSurface, dark.Colors.OnSurface)
	}
}

func TestTheme_ImplementsThemeProvider(t *testing.T) {
	var _ widget.ThemeProvider = (*Theme)(nil)
}

func TestTheme_ScaleTypography(t *testing.T) {
	original := DefaultLight()
	originalSize := original.Typography.BodyMedium.Size

	result := original.ScaleTypography(1.5)

	expectedSize := originalSize * 1.5
	if result.Typography.BodyMedium.Size != expectedSize {
		t.Errorf("BodyMedium.Size = %v, want %v", result.Typography.BodyMedium.Size, expectedSize)
	}
	// Original unchanged
	if original.Typography.BodyMedium.Size != originalSize {
		t.Error("Original typography changed")
	}
}

func TestTheme_ScaleSpacing(t *testing.T) {
	original := DefaultLight()
	originalM := original.Spacing.M

	result := original.ScaleSpacing(0.75)

	expectedM := originalM * 0.75
	if result.Spacing.M != expectedM {
		t.Errorf("Spacing.M = %v, want %v", result.Spacing.M, expectedM)
	}
	// Original unchanged
	if original.Spacing.M != originalM {
		t.Error("Original spacing changed")
	}
}

func TestTheme_Compact(t *testing.T) {
	original := DefaultLight()
	result := original.Compact()

	expected := original.Spacing.M * 0.75
	if result.Spacing.M != expected {
		t.Errorf("Compact Spacing.M = %v, want %v", result.Spacing.M, expected)
	}
}

func TestTheme_Comfortable(t *testing.T) {
	original := DefaultLight()
	result := original.Comfortable()

	expected := original.Spacing.M * 1.25
	if result.Spacing.M != expected {
		t.Errorf("Comfortable Spacing.M = %v, want %v", result.Spacing.M, expected)
	}
}
