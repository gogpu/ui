package plugin

import (
	"errors"
)

// Plugin provides bundled UI components with lifecycle management.
//
// A plugin is a cohesive unit that can register widgets, themes, layouts,
// and assets. Plugins support dependency declaration, allowing the plugin
// system to initialize them in the correct order.
//
// Implement this interface to create a custom plugin:
//
//	type MyPlugin struct{}
//
//	func (p *MyPlugin) Name() string    { return "my-plugin" }
//	func (p *MyPlugin) Version() string { return "1.0.0" }
//
//	func (p *MyPlugin) Dependencies() []Dependency {
//	    return nil
//	}
//
//	func (p *MyPlugin) Init(ctx *PluginContext) error {
//	    ctx.Widgets.Register("my-button", NewMyButton)
//	    return nil
//	}
//
//	func (p *MyPlugin) Shutdown() error {
//	    return nil
//	}
type Plugin interface {
	// Name returns the unique identifier for this plugin.
	//
	// Plugin names should be lowercase, hyphen-separated, and unique
	// within an application. Example: "material3", "corporate-design"
	Name() string

	// Version returns the semantic version of this plugin.
	//
	// Version strings should follow semantic versioning (semver):
	// "MAJOR.MINOR.PATCH", e.g., "1.0.0", "2.1.3"
	Version() string

	// Dependencies returns the list of plugins this plugin depends on.
	//
	// The plugin manager will ensure all dependencies are initialized
	// before this plugin's Init method is called. Return nil or an
	// empty slice if the plugin has no dependencies.
	Dependencies() []Dependency

	// Init is called when the plugin is initialized.
	//
	// Use the provided [PluginContext] to register widgets, themes,
	// layouts, and load assets. The context provides thread-safe
	// access to all registries.
	//
	// Return an error if initialization fails. This will prevent
	// the application from starting.
	Init(ctx *PluginContext) error

	// Shutdown is called when the plugin is unloaded.
	//
	// Use this method to release any resources allocated during Init.
	// Plugins are shut down in reverse initialization order.
	//
	// Return an error if shutdown fails. All plugins will still be
	// attempted to shut down even if one fails.
	Shutdown() error
}

// Dependency declares a dependency on another plugin.
//
// Plugins can depend on other plugins by name and version constraint.
// The plugin manager resolves dependencies and initializes plugins
// in the correct order.
//
// Example:
//
//	func (p *MyPlugin) Dependencies() []Dependency {
//	    return []Dependency{
//	        {Name: "base-widgets", Version: ">=1.0.0"},
//	        {Name: "icons-pack", Version: ">=2.0.0,<3.0.0"},
//	    }
//	}
type Dependency struct {
	// Name is the unique identifier of the required plugin.
	Name string

	// Version is a semantic version constraint.
	//
	// Supported formats:
	//   - "1.0.0" - exact version match
	//   - ">=1.0.0" - minimum version
	//   - "<=1.0.0" - maximum version
	//   - ">1.0.0" - greater than
	//   - "<1.0.0" - less than
	//   - ">=1.0.0,<2.0.0" - range (AND condition)
	//   - "" - any version
	Version string
}

// PluginInfo describes metadata about a registered plugin.
//
// PluginInfo provides human-readable information about a plugin,
// including its name, description, author, and license. This metadata
// can be used to build plugin management UIs or generate documentation.
//
// Example:
//
//	info := PluginInfo{
//	    Name:        "material3",
//	    Description: "Google Material Design 3 components",
//	    Version:     "1.0.0",
//	    Author:      "gogpu team",
//	    License:     "MIT",
//	    Homepage:    "https://github.com/gogpu/ui-material3",
//	}
type PluginInfo struct {
	// Name is the unique identifier for the plugin.
	// Should match Plugin.Name().
	Name string

	// Description is a brief description of what the plugin provides.
	Description string

	// Version is the semantic version of the plugin.
	// Should match Plugin.Version().
	Version string

	// Author is the creator or maintainer of the plugin.
	Author string

	// License is the license under which the plugin is distributed.
	// Example: "MIT", "Apache-2.0", "Proprietary"
	License string

	// Homepage is a URL to the plugin's homepage or repository.
	Homepage string

	// Dependencies lists the plugins this plugin depends on.
	// This is populated from Plugin.Dependencies() during registration.
	Dependencies []Dependency
}

// Common errors returned by plugin operations.
var (
	// ErrPluginNotFound is returned when a plugin is not registered.
	ErrPluginNotFound = errors.New("plugin not found")

	// ErrPluginExists is returned when attempting to register a duplicate plugin.
	ErrPluginExists = errors.New("plugin already registered")

	// ErrNilPlugin is returned when a nil plugin is provided.
	ErrNilPlugin = errors.New("plugin cannot be nil")

	// ErrEmptyName is returned when a plugin has an empty name.
	ErrEmptyName = errors.New("plugin name cannot be empty")

	// ErrAlreadyInitialized is returned when Initialize is called twice.
	ErrAlreadyInitialized = errors.New("plugins already initialized")

	// ErrNotInitialized is returned when operations require initialized plugins.
	ErrNotInitialized = errors.New("plugins not initialized")

	// ErrCircularDependency is returned when circular dependencies are detected.
	ErrCircularDependency = errors.New("circular dependency detected")

	// ErrDependencyNotFound is returned when a required dependency is not registered.
	ErrDependencyNotFound = errors.New("required dependency not found")

	// ErrVersionMismatch is returned when a dependency version constraint is not satisfied.
	ErrVersionMismatch = errors.New("version constraint not satisfied")
)
