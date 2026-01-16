package plugin

import (
	"testing"

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
