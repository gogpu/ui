package theme

import (
	"sync"
	"testing"
)

// TestThemeVariant tests ThemeVariant type and constants.
func TestThemeVariant(t *testing.T) {
	tests := []struct {
		variant ThemeVariant
		want    string
	}{
		{VariantLight, "light"},
		{VariantDark, "dark"},
		{VariantSystem, "system"},
	}

	for _, tt := range tests {
		t.Run(tt.want, func(t *testing.T) {
			if got := tt.variant.String(); got != tt.want {
				t.Errorf("ThemeVariant.String() = %v, want %v", got, tt.want)
			}
		})
	}
}

// TestThemeInfoHasVariant tests the HasVariant method.
func TestThemeInfoHasVariant(t *testing.T) {
	info := ThemeInfo{
		Name:     "Test",
		Variants: []ThemeVariant{VariantLight, VariantDark},
	}

	tests := []struct {
		name    string
		variant ThemeVariant
		want    bool
	}{
		{"has light", VariantLight, true},
		{"has dark", VariantDark, true},
		{"no system", VariantSystem, false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := info.HasVariant(tt.variant); got != tt.want {
				t.Errorf("HasVariant(%v) = %v, want %v", tt.variant, got, tt.want)
			}
		})
	}
}

// TestThemeInfoHasVariantEmpty tests HasVariant with empty variants.
func TestThemeInfoHasVariantEmpty(t *testing.T) {
	info := ThemeInfo{Name: "Empty"}

	if info.HasVariant(VariantLight) {
		t.Error("HasVariant should return false for empty variants")
	}
}

// TestNewThemeRegistry tests registry creation.
func TestNewThemeRegistry(t *testing.T) {
	registry := NewThemeRegistry()

	if registry == nil {
		t.Fatal("NewThemeRegistry returned nil")
	}

	if registry.themes == nil {
		t.Error("themes map is nil")
	}

	if registry.info == nil {
		t.Error("info map is nil")
	}

	if registry.Count() != 0 {
		t.Errorf("new registry should be empty, got count %d", registry.Count())
	}
}

// TestThemeRegistryRegisterAndGet tests basic registration and retrieval.
func TestThemeRegistryRegisterAndGet(t *testing.T) {
	registry := NewThemeRegistry()

	// Create a test theme
	theme := New("Test Theme", ModeLight)

	// Register without info
	registry.Register("test", theme)

	// Retrieve
	got, ok := registry.Get("test")
	if !ok {
		t.Fatal("Get returned false for registered theme")
	}

	if got != theme {
		t.Error("Get returned different theme instance")
	}

	// Check default info was created
	info, ok := registry.Info("test")
	if !ok {
		t.Fatal("Info returned false for registered theme")
	}

	if info.Name != "Test Theme" {
		t.Errorf("default info Name = %q, want %q", info.Name, "Test Theme")
	}
}

// TestThemeRegistryRegisterWithInfo tests registration with explicit info.
func TestThemeRegistryRegisterWithInfo(t *testing.T) {
	registry := NewThemeRegistry()

	theme := New("Corporate", ModeLight)
	info := ThemeInfo{
		Name:        "Corporate Theme",
		Description: "A branded theme",
		Author:      "Design Team",
		Version:     "1.0.0",
		Variants:    []ThemeVariant{VariantLight, VariantDark},
		Preview:     "https://example.com/preview.png",
	}

	registry.Register("corporate", theme, info)

	gotInfo, ok := registry.Info("corporate")
	if !ok {
		t.Fatal("Info returned false")
	}

	if gotInfo.Name != info.Name {
		t.Errorf("Info.Name = %q, want %q", gotInfo.Name, info.Name)
	}

	if gotInfo.Description != info.Description {
		t.Errorf("Info.Description = %q, want %q", gotInfo.Description, info.Description)
	}

	if gotInfo.Author != info.Author {
		t.Errorf("Info.Author = %q, want %q", gotInfo.Author, info.Author)
	}

	if gotInfo.Version != info.Version {
		t.Errorf("Info.Version = %q, want %q", gotInfo.Version, info.Version)
	}

	if gotInfo.Preview != info.Preview {
		t.Errorf("Info.Preview = %q, want %q", gotInfo.Preview, info.Preview)
	}

	if len(gotInfo.Variants) != len(info.Variants) {
		t.Errorf("Info.Variants length = %d, want %d", len(gotInfo.Variants), len(info.Variants))
	}
}

// TestThemeRegistryRegisterOverwrite tests that re-registration overwrites.
func TestThemeRegistryRegisterOverwrite(t *testing.T) {
	registry := NewThemeRegistry()

	theme1 := New("First", ModeLight)
	theme2 := New("Second", ModeDark)

	registry.Register("test", theme1, ThemeInfo{Name: "First"})
	registry.Register("test", theme2, ThemeInfo{Name: "Second"})

	got, ok := registry.Get("test")
	if !ok {
		t.Fatal("Get returned false")
	}

	if got != theme2 {
		t.Error("re-registration should overwrite")
	}

	info, _ := registry.Info("test")
	if info.Name != "Second" {
		t.Errorf("info should be overwritten, got %q", info.Name)
	}
}

// TestThemeRegistryUnregister tests theme unregistration.
func TestThemeRegistryUnregister(t *testing.T) {
	registry := NewThemeRegistry()

	theme := New("Test", ModeLight)
	registry.Register("test", theme, ThemeInfo{Name: "Test"})

	// Unregister existing
	if !registry.Unregister("test") {
		t.Error("Unregister should return true for existing theme")
	}

	// Verify removal
	if _, ok := registry.Get("test"); ok {
		t.Error("theme should be removed after Unregister")
	}

	if _, ok := registry.Info("test"); ok {
		t.Error("info should be removed after Unregister")
	}

	// Unregister non-existing
	if registry.Unregister("nonexistent") {
		t.Error("Unregister should return false for non-existing theme")
	}
}

// TestThemeRegistryGetNotFound tests Get for non-existent themes.
func TestThemeRegistryGetNotFound(t *testing.T) {
	registry := NewThemeRegistry()

	got, ok := registry.Get("nonexistent")
	if ok {
		t.Error("Get should return false for non-existent theme")
	}

	if got != nil {
		t.Error("Get should return nil for non-existent theme")
	}
}

// TestThemeRegistryMustGet tests MustGet success case.
func TestThemeRegistryMustGet(t *testing.T) {
	registry := NewThemeRegistry()

	theme := New("Test", ModeLight)
	registry.Register("test", theme)

	got := registry.MustGet("test")
	if got != theme {
		t.Error("MustGet returned different theme")
	}
}

// TestThemeRegistryMustGetPanic tests MustGet panic case.
func TestThemeRegistryMustGetPanic(t *testing.T) {
	registry := NewThemeRegistry()

	defer func() {
		r := recover()
		if r == nil {
			t.Error("MustGet should panic for non-existent theme")
		}

		if msg, ok := r.(string); ok {
			if msg != "theme not found: nonexistent" {
				t.Errorf("panic message = %q, want %q", msg, "theme not found: nonexistent")
			}
		}
	}()

	_ = registry.MustGet("nonexistent")
}

// TestThemeRegistryList tests List returns sorted names.
func TestThemeRegistryList(t *testing.T) {
	registry := NewThemeRegistry()

	registry.Register("zebra", New("Zebra", ModeLight))
	registry.Register("alpha", New("Alpha", ModeLight))
	registry.Register("middle", New("Middle", ModeLight))

	names := registry.List()

	if len(names) != 3 {
		t.Fatalf("List length = %d, want 3", len(names))
	}

	// Should be sorted
	expected := []string{"alpha", "middle", "zebra"}
	for i, name := range names {
		if name != expected[i] {
			t.Errorf("List[%d] = %q, want %q", i, name, expected[i])
		}
	}
}

// TestThemeRegistryListEmpty tests List on empty registry.
func TestThemeRegistryListEmpty(t *testing.T) {
	registry := NewThemeRegistry()

	names := registry.List()
	if len(names) != 0 {
		t.Errorf("List on empty registry should return empty slice, got %v", names)
	}
}

// TestThemeRegistryInfoNotFound tests Info for non-existent themes.
func TestThemeRegistryInfoNotFound(t *testing.T) {
	registry := NewThemeRegistry()

	info, ok := registry.Info("nonexistent")
	if ok {
		t.Error("Info should return false for non-existent theme")
	}

	if info.Name != "" {
		t.Error("Info should return empty ThemeInfo")
	}
}

// TestThemeRegistryCount tests Count method.
func TestThemeRegistryCount(t *testing.T) {
	registry := NewThemeRegistry()

	if registry.Count() != 0 {
		t.Error("empty registry Count should be 0")
	}

	registry.Register("one", New("One", ModeLight))
	if registry.Count() != 1 {
		t.Errorf("Count = %d, want 1", registry.Count())
	}

	registry.Register("two", New("Two", ModeLight))
	if registry.Count() != 2 {
		t.Errorf("Count = %d, want 2", registry.Count())
	}

	registry.Unregister("one")
	if registry.Count() != 1 {
		t.Errorf("Count after unregister = %d, want 1", registry.Count())
	}
}

// TestThemeRegistryHas tests Has method.
func TestThemeRegistryHas(t *testing.T) {
	registry := NewThemeRegistry()

	registry.Register("exists", New("Exists", ModeLight))

	if !registry.Has("exists") {
		t.Error("Has should return true for registered theme")
	}

	if registry.Has("nonexistent") {
		t.Error("Has should return false for non-existent theme")
	}
}

// TestThemeRegistryClear tests Clear method.
func TestThemeRegistryClear(t *testing.T) {
	registry := NewThemeRegistry()

	registry.Register("one", New("One", ModeLight))
	registry.Register("two", New("Two", ModeLight))

	if registry.Count() != 2 {
		t.Fatalf("expected 2 themes before clear")
	}

	registry.Clear()

	if registry.Count() != 0 {
		t.Errorf("Count after Clear = %d, want 0", registry.Count())
	}

	if len(registry.List()) != 0 {
		t.Error("List after Clear should be empty")
	}
}

// TestThemeRegistryListByVariant tests ListByVariant method.
func TestThemeRegistryListByVariant(t *testing.T) {
	registry := NewThemeRegistry()

	registry.Register("light-only", New("Light", ModeLight), ThemeInfo{
		Name:     "Light Only",
		Variants: []ThemeVariant{VariantLight},
	})

	registry.Register("dark-only", New("Dark", ModeDark), ThemeInfo{
		Name:     "Dark Only",
		Variants: []ThemeVariant{VariantDark},
	})

	registry.Register("both", New("Both", ModeLight), ThemeInfo{
		Name:     "Both",
		Variants: []ThemeVariant{VariantLight, VariantDark},
	})

	// Light themes
	lightThemes := registry.ListByVariant(VariantLight)
	if len(lightThemes) != 2 {
		t.Errorf("ListByVariant(Light) count = %d, want 2", len(lightThemes))
	}

	// Should be sorted: "both", "light-only"
	if lightThemes[0] != "both" || lightThemes[1] != "light-only" {
		t.Errorf("ListByVariant(Light) = %v, want [both light-only]", lightThemes)
	}

	// Dark themes
	darkThemes := registry.ListByVariant(VariantDark)
	if len(darkThemes) != 2 {
		t.Errorf("ListByVariant(Dark) count = %d, want 2", len(darkThemes))
	}

	// System themes (none registered)
	systemThemes := registry.ListByVariant(VariantSystem)
	if len(systemThemes) != 0 {
		t.Errorf("ListByVariant(System) count = %d, want 0", len(systemThemes))
	}
}

// TestThemeRegistryConcurrency tests thread-safety.
func TestThemeRegistryConcurrency(t *testing.T) {
	registry := NewThemeRegistry()

	const goroutines = 100
	const operations = 100

	var wg sync.WaitGroup
	wg.Add(goroutines)

	for i := 0; i < goroutines; i++ {
		go func() {
			defer wg.Done()

			for j := 0; j < operations; j++ {
				name := "theme"
				theme := New("Test", ModeLight)

				// Mix of operations
				switch j % 5 {
				case 0:
					registry.Register(name, theme)
				case 1:
					_, _ = registry.Get(name)
				case 2:
					_ = registry.List()
				case 3:
					_, _ = registry.Info(name)
				case 4:
					_ = registry.Has(name)
				}
			}
		}()
	}

	wg.Wait()
	// If we get here without deadlock or race, the test passes
}

// Tests for global registry functions

// TestGlobalRegistryBuiltInThemes tests that built-in themes are registered.
func TestGlobalRegistryBuiltInThemes(t *testing.T) {
	// These should be registered by init()
	expectedThemes := []string{"light", "dark", "high-contrast", "purple", "green", "orange"}

	for _, name := range expectedThemes {
		if !Has(name) {
			t.Errorf("built-in theme %q not registered", name)
		}

		theme, ok := Get(name)
		if !ok {
			t.Errorf("Get(%q) returned false", name)
			continue
		}

		if theme == nil {
			t.Errorf("Get(%q) returned nil theme", name)
		}
	}
}

// TestGlobalRegistryInfo tests Info for built-in themes.
func TestGlobalRegistryInfo(t *testing.T) {
	tests := []struct {
		name     string
		wantName string
		wantAuth string
	}{
		{"light", "Light", "gogpu"},
		{"dark", "Dark", "gogpu"},
		{"high-contrast", "High Contrast", "gogpu"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			info, ok := Info(tt.name)
			if !ok {
				t.Fatalf("Info(%q) returned false", tt.name)
			}

			if info.Name != tt.wantName {
				t.Errorf("Info.Name = %q, want %q", info.Name, tt.wantName)
			}

			if info.Author != tt.wantAuth {
				t.Errorf("Info.Author = %q, want %q", info.Author, tt.wantAuth)
			}

			if info.Version != "1.0.0" {
				t.Errorf("Info.Version = %q, want %q", info.Version, "1.0.0")
			}
		})
	}
}

// TestGlobalRegistryMustGet tests MustGet for built-in themes.
func TestGlobalRegistryMustGet(t *testing.T) {
	// Should not panic
	light := MustGet("light")
	if light == nil {
		t.Error("MustGet(light) returned nil")
	}

	dark := MustGet("dark")
	if dark == nil {
		t.Error("MustGet(dark) returned nil")
	}
}

// TestGlobalRegistryMustGetPanic tests MustGet panic for non-existent theme.
func TestGlobalRegistryMustGetPanic(t *testing.T) {
	defer func() {
		if r := recover(); r == nil {
			t.Error("MustGet should panic for non-existent theme")
		}
	}()

	_ = MustGet("definitely-does-not-exist-12345")
}

// TestGlobalRegistryList tests List returns sorted names.
func TestGlobalRegistryList(t *testing.T) {
	names := List()

	// Should include built-in themes
	builtIn := map[string]bool{
		"light":         false,
		"dark":          false,
		"high-contrast": false,
		"purple":        false,
		"green":         false,
		"orange":        false,
	}

	for _, name := range names {
		builtIn[name] = true
	}

	for name, found := range builtIn {
		if !found {
			t.Errorf("built-in theme %q not in List()", name)
		}
	}

	// Verify sorted order
	for i := 1; i < len(names); i++ {
		if names[i-1] > names[i] {
			t.Errorf("List not sorted: %q > %q", names[i-1], names[i])
		}
	}
}

// TestGlobalRegistryCount tests Count includes built-in themes.
func TestGlobalRegistryCount(t *testing.T) {
	count := Count()
	if count < 6 {
		t.Errorf("Count = %d, want at least 6 (built-in themes)", count)
	}
}

// TestGlobalRegistryRegisterAndUnregister tests custom theme registration.
func TestGlobalRegistryRegisterAndUnregister(t *testing.T) {
	// Register a custom theme
	customTheme := New("Custom", ModeLight)
	customInfo := ThemeInfo{
		Name:        "Custom Theme",
		Description: "A test theme",
		Author:      "Test",
		Version:     "0.1.0",
		Variants:    []ThemeVariant{VariantLight},
	}

	// Use a unique name to avoid conflicts
	const customName = "test-custom-registry-12345"

	Register(customName, customTheme, customInfo)

	// Verify registration
	if !Has(customName) {
		t.Error("custom theme not registered")
	}

	got, ok := Get(customName)
	if !ok || got != customTheme {
		t.Error("Get returned wrong theme")
	}

	info, ok := Info(customName)
	if !ok || info.Name != customInfo.Name {
		t.Error("Info returned wrong info")
	}

	// Clean up
	if !Unregister(customName) {
		t.Error("Unregister returned false")
	}

	if Has(customName) {
		t.Error("theme still exists after Unregister")
	}
}

// TestGlobalRegistryListByVariant tests ListByVariant.
func TestGlobalRegistryListByVariant(t *testing.T) {
	lightThemes := ListByVariant(VariantLight)
	darkThemes := ListByVariant(VariantDark)

	// Light themes should include: light, high-contrast, purple, green, orange
	if len(lightThemes) < 5 {
		t.Errorf("ListByVariant(Light) = %d themes, want at least 5", len(lightThemes))
	}

	// Dark themes should include: dark
	if len(darkThemes) < 1 {
		t.Errorf("ListByVariant(Dark) = %d themes, want at least 1", len(darkThemes))
	}

	// Verify "light" is in light themes
	found := false
	for _, name := range lightThemes {
		if name == "light" {
			found = true
			break
		}
	}
	if !found {
		t.Error("'light' theme not in ListByVariant(Light)")
	}

	// Verify "dark" is in dark themes
	found = false
	for _, name := range darkThemes {
		if name == "dark" {
			found = true
			break
		}
	}
	if !found {
		t.Error("'dark' theme not in ListByVariant(Dark)")
	}
}

// TestGlobalRegistry tests GlobalRegistry returns the singleton.
func TestGlobalRegistry(t *testing.T) {
	reg := GlobalRegistry()

	if reg == nil {
		t.Fatal("GlobalRegistry returned nil")
	}

	if reg != globalRegistry {
		t.Error("GlobalRegistry did not return the global registry")
	}

	// Operations on returned registry should affect global state
	const testName = "global-registry-test-12345"
	reg.Register(testName, New("Test", ModeLight))

	if !Has(testName) {
		t.Error("registration through GlobalRegistry() not visible globally")
	}

	// Clean up
	_ = Unregister(testName)
}

// TestBuiltInThemeVariants tests that built-in themes have correct variants.
func TestBuiltInThemeVariants(t *testing.T) {
	tests := []struct {
		name    string
		variant ThemeVariant
		want    bool
	}{
		{"light", VariantLight, true},
		{"light", VariantDark, false},
		{"dark", VariantDark, true},
		{"dark", VariantLight, false},
		{"high-contrast", VariantLight, true},
		{"purple", VariantLight, true},
		{"green", VariantLight, true},
		{"orange", VariantLight, true},
	}

	for _, tt := range tests {
		t.Run(tt.name+"-"+string(tt.variant), func(t *testing.T) {
			info, ok := Info(tt.name)
			if !ok {
				t.Fatalf("theme %q not found", tt.name)
			}

			if got := info.HasVariant(tt.variant); got != tt.want {
				t.Errorf("HasVariant(%v) = %v, want %v", tt.variant, got, tt.want)
			}
		})
	}
}

// TestBuiltInThemeConsistency tests built-in themes match their presets.
func TestBuiltInThemeConsistency(t *testing.T) {
	// Light theme should be equivalent to DefaultLight()
	light, _ := Get("light")
	defaultLight := DefaultLight()

	if light.Name != defaultLight.Name {
		t.Errorf("light theme Name mismatch: %q vs %q", light.Name, defaultLight.Name)
	}

	if light.Mode != defaultLight.Mode {
		t.Errorf("light theme Mode mismatch: %v vs %v", light.Mode, defaultLight.Mode)
	}

	// Dark theme should be equivalent to DefaultDark()
	dark, _ := Get("dark")
	defaultDark := DefaultDark()

	if dark.Name != defaultDark.Name {
		t.Errorf("dark theme Name mismatch: %q vs %q", dark.Name, defaultDark.Name)
	}

	if dark.Mode != defaultDark.Mode {
		t.Errorf("dark theme Mode mismatch: %v vs %v", dark.Mode, defaultDark.Mode)
	}
}
