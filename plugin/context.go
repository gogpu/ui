package plugin

import (
	"github.com/gogpu/ui/internal/render"
	"github.com/gogpu/ui/layout"
	"github.com/gogpu/ui/registry"
	"github.com/gogpu/ui/theme"
	"github.com/gogpu/ui/theme/font"
)

// PluginContext provides access to UI registries for plugin initialization.
//
// When a plugin's Init method is called, it receives a PluginContext
// that provides access to the widget, theme, and layout registries,
// as well as an asset loader for fonts, icons, and images.
//
// All registry operations are thread-safe, but plugins should still
// be careful about the order of registration to avoid dependencies
// on components that haven't been registered yet.
//
// Example:
//
//	func (p *MyPlugin) Init(ctx *PluginContext) error {
//	    // Register widgets
//	    ctx.Widgets.Register("my-button", NewMyButton, registry.WidgetInfo{
//	        Name:        "my-button",
//	        Description: "A custom button widget",
//	        Category:    registry.CategoryInput,
//	    })
//
//	    // Register themes
//	    ctx.Themes.Register("my-theme", myTheme, theme.ThemeInfo{
//	        Name:        "My Theme",
//	        Description: "A custom theme",
//	    })
//
//	    // Register layouts
//	    ctx.Layouts.Register(&MyLayout{})
//
//	    // Load assets
//	    ctx.Assets.LoadFont("my-font", fontData)
//	    ctx.Assets.LoadIcon("my-icon", iconData)
//
//	    return nil
//	}
type PluginContext struct {
	// Widgets provides access to the widget registry.
	//
	// Use this to register widget factories that create instances of
	// your custom widgets.
	Widgets *registry.WidgetRegistry

	// Themes provides access to the theme registry.
	//
	// Use this to register themes that define colors, typography,
	// spacing, and other visual properties.
	Themes *theme.ThemeRegistry

	// Layouts provides access to the layout registry.
	//
	// Use this to register custom layout algorithms that control
	// how widgets are arranged.
	Layouts *layout.Registry

	// Assets provides methods for loading plugin resources.
	//
	// Use this to load fonts, icons, and images that your plugin
	// provides.
	Assets AssetLoader
}

// NewPluginContext creates a new PluginContext with the given registries.
//
// If any registry is nil, the global registry for that type is used.
// If assets is nil, a no-op asset loader is used.
func NewPluginContext(
	widgets *registry.WidgetRegistry,
	themes *theme.ThemeRegistry,
	layouts *layout.Registry,
	assets AssetLoader,
) *PluginContext {
	ctx := &PluginContext{
		Widgets: widgets,
		Themes:  themes,
		Layouts: layouts,
		Assets:  assets,
	}

	// Use global registries if not provided
	if ctx.Widgets == nil {
		ctx.Widgets = registry.GlobalRegistry()
	}
	if ctx.Themes == nil {
		ctx.Themes = theme.GlobalRegistry()
	}
	if ctx.Layouts == nil {
		ctx.Layouts = layout.GlobalRegistry()
	}
	if ctx.Assets == nil {
		loader := NewMemoryAssetLoader()
		loader.SetFontRegisterer(func(name string, data []byte) error {
			return render.GlobalFontRegistry().Register(name, font.Regular, font.Normal, data)
		})
		ctx.Assets = loader
	}

	return ctx
}

// NewDefaultPluginContext creates a PluginContext with global registries
// and a [MemoryAssetLoader] wired to the global font registry.
//
// Fonts loaded via [AssetLoader.LoadFont] are automatically registered
// with the rendering pipeline's [render.GlobalFontRegistry], making them
// available to widgets that use [widget.StyledTextDrawer] (e.g.,
// [primitives.TextWidget] with [primitives.TextWidget.FontFamily]).
//
// This is the standard context used when initializing plugins through
// the global plugin manager.
func NewDefaultPluginContext() *PluginContext {
	loader := NewMemoryAssetLoader()
	loader.SetFontRegisterer(func(name string, data []byte) error {
		return render.GlobalFontRegistry().Register(name, font.Regular, font.Normal, data)
	})
	return NewPluginContext(nil, nil, nil, loader)
}
