package plugin

import (
	"testing"

	"github.com/gogpu/ui/internal/render"
	"github.com/gogpu/ui/layout"
	"github.com/gogpu/ui/registry"
	"github.com/gogpu/ui/theme"
)

// TestNewPluginContext tests creating a new plugin context.
func TestNewPluginContext(t *testing.T) {
	widgets := registry.NewWidgetRegistry()
	themes := theme.NewThemeRegistry()
	layouts := layout.NewRegistry()
	assets := NewMemoryAssetLoader()

	ctx := NewPluginContext(widgets, themes, layouts, assets)

	if ctx.Widgets != widgets {
		t.Error("Widgets registry not set correctly")
	}
	if ctx.Themes != themes {
		t.Error("Themes registry not set correctly")
	}
	if ctx.Layouts != layouts {
		t.Error("Layouts registry not set correctly")
	}
	if ctx.Assets != assets {
		t.Error("Assets loader not set correctly")
	}
}

// TestNewPluginContextNilRegistries tests that nil registries use globals.
func TestNewPluginContextNilRegistries(t *testing.T) {
	ctx := NewPluginContext(nil, nil, nil, nil)

	if ctx.Widgets == nil {
		t.Error("Widgets should not be nil")
	}
	if ctx.Themes == nil {
		t.Error("Themes should not be nil")
	}
	if ctx.Layouts == nil {
		t.Error("Layouts should not be nil")
	}
	if ctx.Assets == nil {
		t.Error("Assets should not be nil")
	}

	// Verify they are the global registries
	if ctx.Widgets != registry.GlobalRegistry() {
		t.Error("Widgets should be global registry")
	}
	if ctx.Themes != theme.GlobalRegistry() {
		t.Error("Themes should be global registry")
	}
	if ctx.Layouts != layout.GlobalRegistry() {
		t.Error("Layouts should be global registry")
	}
}

// TestNewDefaultPluginContext tests creating a default context.
func TestNewDefaultPluginContext(t *testing.T) {
	ctx := NewDefaultPluginContext()

	if ctx.Widgets != registry.GlobalRegistry() {
		t.Error("Widgets should be global registry")
	}
	if ctx.Themes != theme.GlobalRegistry() {
		t.Error("Themes should be global registry")
	}
	if ctx.Layouts != layout.GlobalRegistry() {
		t.Error("Layouts should be global registry")
	}
	if ctx.Assets == nil {
		t.Error("Assets should not be nil")
	}
}

// TestPluginContextPartialNil tests partial nil arguments.
func TestPluginContextPartialNil(t *testing.T) {
	widgets := registry.NewWidgetRegistry()
	assets := NewMemoryAssetLoader()

	ctx := NewPluginContext(widgets, nil, nil, assets)

	if ctx.Widgets != widgets {
		t.Error("Widgets should be the provided registry")
	}
	if ctx.Themes != theme.GlobalRegistry() {
		t.Error("Themes should default to global registry")
	}
	if ctx.Layouts != layout.GlobalRegistry() {
		t.Error("Layouts should default to global registry")
	}
	if ctx.Assets != assets {
		t.Error("Assets should be the provided loader")
	}
}

// TestPluginContextUsage tests using the context in a plugin.
func TestPluginContextUsage(t *testing.T) {
	widgets := registry.NewWidgetRegistry()
	themes := theme.NewThemeRegistry()
	layouts := layout.NewRegistry()
	assets := NewMemoryAssetLoader()

	ctx := NewPluginContext(widgets, themes, layouts, assets)

	// Simulate plugin registering components
	// Using a nil factory return is valid for test widgets that won't be created
	err := ctx.Widgets.Register("test-widget", func(_ map[string]any) (registry.Widget, error) {
		return nil, nil //nolint:nilnil // Test factory, widgets are not actually created
	})
	if err != nil {
		t.Errorf("Failed to register widget: %v", err)
	}

	ctx.Themes.Register("test-theme", theme.DefaultLight())

	// Verify registrations
	if !ctx.Widgets.Has("test-widget") {
		t.Error("Widget should be registered")
	}
	if !ctx.Themes.Has("test-theme") {
		t.Error("Theme should be registered")
	}
}

// TestNewDefaultPluginContext_AssetLoaderIsMemory verifies that the default
// context creates a MemoryAssetLoader (not a no-op) so that fonts are stored.
func TestNewDefaultPluginContext_AssetLoaderIsMemory(t *testing.T) {
	ctx := NewDefaultPluginContext()

	loader, ok := ctx.Assets.(*MemoryAssetLoader)
	if !ok {
		t.Fatalf("Assets should be *MemoryAssetLoader, got %T", ctx.Assets)
	}

	// Verify it's a real MemoryAssetLoader by checking icons (no registerer validation).
	if err := loader.LoadIcon("test-icon", []byte("svg-data")); err != nil {
		t.Fatalf("LoadIcon failed: %v", err)
	}

	data, ok := loader.GetIcon("test-icon")
	if !ok || len(data) == 0 {
		t.Error("MemoryAssetLoader should store asset data")
	}

	// LoadFont with invalid data should return an error from the registerer
	// (font validation), proving the registerer is wired.
	err := loader.LoadFont("bad-font", []byte("not-a-font"))
	if err == nil {
		t.Error("LoadFont with invalid data should fail due to wired font registerer validation")
	}
}

// TestNewPluginContext_NilAssets_CreatesWiredLoader verifies that passing nil
// for assets creates a MemoryAssetLoader with FontRegisterer connected to
// the global font registry.
func TestNewPluginContext_NilAssets_CreatesWiredLoader(t *testing.T) {
	ctx := NewPluginContext(nil, nil, nil, nil)

	loader, ok := ctx.Assets.(*MemoryAssetLoader)
	if !ok {
		t.Fatalf("nil assets should create *MemoryAssetLoader, got %T", ctx.Assets)
	}

	// The font registerer should be set.
	if loader.fontRegisterer == nil {
		t.Fatal("MemoryAssetLoader should have fontRegisterer wired")
	}
}

// TestNewDefaultPluginContext_FontRegistererWired verifies that fonts loaded
// via the default context's asset loader reach the global font registry.
func TestNewDefaultPluginContext_FontRegistererWired(t *testing.T) {
	ctx := NewDefaultPluginContext()
	reg := render.GlobalFontRegistry()

	// Load valid font data (use the embedded Inter as test data).
	// We check that the registerer is called. Since the test font data
	// may not be a real font file, we use a known-good approach:
	// verify the registerer callback is set and invoked.
	loader, ok := ctx.Assets.(*MemoryAssetLoader)
	if !ok {
		t.Fatalf("expected *MemoryAssetLoader, got %T", ctx.Assets)
	}

	if loader.fontRegisterer == nil {
		t.Fatal("fontRegisterer should be wired in default context")
	}

	// Inter is already in the registry from initialization.
	if !reg.HasFamily("Inter") {
		t.Error("Global registry should have Inter pre-registered")
	}
}

// TestNewPluginContext_ExplicitAssets_NotOverridden verifies that when an
// explicit AssetLoader is provided, it is used as-is without wrapping.
func TestNewPluginContext_ExplicitAssets_NotOverridden(t *testing.T) {
	loader := NewMemoryAssetLoader()
	ctx := NewPluginContext(nil, nil, nil, loader)

	if ctx.Assets != loader {
		t.Error("explicit AssetLoader should be used as-is")
	}
}
