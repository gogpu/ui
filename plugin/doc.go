// Package plugin provides a system for bundling UI components into cohesive packages.
//
// Plugins allow developers to distribute collections of widgets, themes, and layouts
// as a single unit with proper lifecycle management and dependency tracking.
//
// # Overview
//
// A plugin implements the [Plugin] interface and can register any combination of:
//   - Widgets via [registry.WidgetRegistry]
//   - Themes via [theme.ThemeRegistry]
//   - Layouts via [layout.Registry]
//   - Assets (fonts, icons, images) via [AssetLoader]
//
// # Creating a Plugin
//
// Implement the [Plugin] interface:
//
//	type MyPlugin struct{}
//
//	func (p *MyPlugin) Name() string    { return "my-plugin" }
//	func (p *MyPlugin) Version() string { return "1.0.0" }
//
//	func (p *MyPlugin) Dependencies() []plugin.Dependency {
//	    return nil // No dependencies
//	}
//
//	func (p *MyPlugin) Init(ctx *plugin.PluginContext) error {
//	    // Register widgets
//	    ctx.Widgets.Register("my-button", NewMyButton)
//
//	    // Register themes
//	    ctx.Themes.Register("my-theme", myTheme)
//
//	    // Load assets
//	    ctx.Assets.LoadFont("my-font", fontData)
//
//	    return nil
//	}
//
//	func (p *MyPlugin) Shutdown() error {
//	    return nil // Cleanup if needed
//	}
//
// # Registering Plugins
//
// Register plugins in your package's init() function:
//
//	func init() {
//	    plugin.Register(&MyPlugin{}, plugin.PluginInfo{
//	        Name:        "my-plugin",
//	        Description: "My custom UI plugin",
//	        Version:     "1.0.0",
//	        Author:      "My Team",
//	        License:     "MIT",
//	    })
//	}
//
// # Using Plugins
//
// In your application, initialize all registered plugins:
//
//	func main() {
//	    // Initialize all plugins
//	    if err := plugin.Initialize(); err != nil {
//	        log.Fatal(err)
//	    }
//	    defer plugin.Shutdown()
//
//	    // List available plugins
//	    for _, name := range plugin.List() {
//	        info, _ := plugin.Info(name)
//	        fmt.Printf("%s v%s - %s\n", name, info.Version, info.Description)
//	    }
//
//	    // Use components from plugins
//	    widget, _ := registry.CreateWidget("my-button", nil)
//	    theme, _ := theme.Get("my-theme")
//	}
//
// # Dependencies
//
// Plugins can declare dependencies on other plugins:
//
//	func (p *MyPlugin) Dependencies() []plugin.Dependency {
//	    return []plugin.Dependency{
//	        {Name: "base-plugin", Version: ">=1.0.0"},
//	    }
//	}
//
// The [PluginManager] automatically resolves dependencies and initializes
// plugins in the correct order. Circular dependencies are detected and
// reported as errors.
//
// # Version Constraints
//
// Version constraints support semantic versioning:
//   - "1.0.0" - exact version
//   - ">=1.0.0" - minimum version
//   - "<=1.0.0" - maximum version
//   - ">1.0.0" - greater than
//   - "<1.0.0" - less than
//   - ">=1.0.0,<2.0.0" - range (AND)
//
// # Thread Safety
//
// All operations on the global [PluginManager] are thread-safe.
// Plugins are initialized sequentially in dependency order to avoid
// race conditions during registration.
package plugin
