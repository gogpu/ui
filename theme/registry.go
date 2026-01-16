package theme

import (
	"sort"
	"sync"
)

// ThemeVariant represents the visual variant of a theme.
//
// Theme variants indicate whether a theme is designed for light,
// dark, or system-following color schemes. A single theme may
// support multiple variants.
type ThemeVariant string

const (
	// VariantLight indicates a light color scheme.
	VariantLight ThemeVariant = "light"

	// VariantDark indicates a dark color scheme.
	VariantDark ThemeVariant = "dark"

	// VariantSystem indicates the theme follows system preferences.
	VariantSystem ThemeVariant = "system"
)

// String returns the string representation of the variant.
func (v ThemeVariant) String() string {
	return string(v)
}

// ThemeInfo describes metadata about a registered theme.
//
// ThemeInfo provides human-readable information about a theme,
// including its name, description, author, and supported variants.
// This metadata can be used to build theme selection UIs.
//
// Example:
//
//	info := ThemeInfo{
//	    Name:        "Corporate Blue",
//	    Description: "Company branded theme with blue accents",
//	    Author:      "Design Team",
//	    Version:     "1.2.0",
//	    Variants:    []ThemeVariant{VariantLight, VariantDark},
//	    Preview:     "https://example.com/theme-preview.png",
//	}
type ThemeInfo struct {
	// Name is the human-readable name of the theme.
	// Example: "Material Light", "Corporate Blue"
	Name string

	// Description provides a brief description of the theme.
	// Example: "A clean, modern theme with blue accents"
	Description string

	// Author is the creator or maintainer of the theme.
	// Example: "Design Team", "John Doe"
	Author string

	// Version is the semantic version of the theme.
	// Example: "1.0.0", "2.1.3"
	Version string

	// Variants lists the color scheme variants this theme supports.
	// A theme may support light, dark, or both variants.
	Variants []ThemeVariant

	// Preview is an optional URL to a preview image of the theme.
	// This can be used in theme selection UIs to show a visual preview.
	Preview string
}

// HasVariant returns true if the theme supports the given variant.
func (ti ThemeInfo) HasVariant(variant ThemeVariant) bool {
	for _, v := range ti.Variants {
		if v == variant {
			return true
		}
	}
	return false
}

// ThemeRegistry manages registration and retrieval of themes.
//
// ThemeRegistry provides thread-safe storage for themes and their
// metadata. It enables dynamic theme discovery and switching in
// applications. Third-party packages can register themes via init()
// functions, making them automatically available to applications.
//
// ThemeRegistry follows the same patterns as other registries in
// the gogpu/ui ecosystem, such as the Widget Registry.
//
// Example:
//
//	registry := NewThemeRegistry()
//	registry.Register("corporate", corporateTheme, ThemeInfo{
//	    Name: "Corporate",
//	    Description: "Company branded theme",
//	})
//
//	// Later, retrieve the theme
//	if theme, ok := registry.Get("corporate"); ok {
//	    app.SetTheme(theme)
//	}
type ThemeRegistry struct {
	mu     sync.RWMutex
	themes map[string]*Theme
	info   map[string]ThemeInfo
}

// NewThemeRegistry creates a new empty ThemeRegistry.
//
// Use the global registry functions (Register, Get, List, etc.) for
// most use cases. Create a custom registry only when you need
// isolated theme management.
func NewThemeRegistry() *ThemeRegistry {
	return &ThemeRegistry{
		themes: make(map[string]*Theme),
		info:   make(map[string]ThemeInfo),
	}
}

// Register adds a theme to the registry.
//
// The name should be a unique identifier for the theme (e.g., "light",
// "dark", "corporate-blue"). If a theme with the same name already
// exists, it will be replaced.
//
// The optional info parameter provides metadata about the theme.
// If not provided, default info will be created from the theme's Name field.
//
// Example:
//
//	registry.Register("ocean", oceanTheme, ThemeInfo{
//	    Name:        "Ocean Blue",
//	    Description: "A calming blue theme",
//	    Variants:    []ThemeVariant{VariantLight},
//	})
func (r *ThemeRegistry) Register(name string, theme *Theme, info ...ThemeInfo) {
	r.mu.Lock()
	defer r.mu.Unlock()

	r.themes[name] = theme

	if len(info) > 0 {
		r.info[name] = info[0]
	} else {
		// Create default info from theme
		r.info[name] = ThemeInfo{
			Name: theme.Name,
		}
	}
}

// Unregister removes a theme from the registry.
//
// Returns true if the theme was found and removed, false otherwise.
// This is useful for dynamically loaded themes that need to be unloaded.
func (r *ThemeRegistry) Unregister(name string) bool {
	r.mu.Lock()
	defer r.mu.Unlock()

	if _, ok := r.themes[name]; ok {
		delete(r.themes, name)
		delete(r.info, name)
		return true
	}
	return false
}

// Get retrieves a theme by its registered name.
//
// Returns the theme and true if found, or nil and false if not found.
//
// Example:
//
//	if theme, ok := registry.Get("dark"); ok {
//	    app.SetTheme(theme)
//	}
func (r *ThemeRegistry) Get(name string) (*Theme, bool) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	theme, ok := r.themes[name]
	return theme, ok
}

// MustGet retrieves a theme by name or panics if not found.
//
// Use this when you are certain the theme exists (e.g., built-in themes).
// For optional themes, use Get instead.
//
// Example:
//
//	// Safe for built-in themes
//	lightTheme := registry.MustGet("light")
func (r *ThemeRegistry) MustGet(name string) *Theme {
	theme, ok := r.Get(name)
	if !ok {
		panic("theme not found: " + name)
	}
	return theme
}

// List returns the names of all registered themes.
//
// The names are returned in sorted order for consistent iteration.
// This is useful for building theme selection UIs.
//
// Example:
//
//	for _, name := range registry.List() {
//	    fmt.Println(name)
//	}
func (r *ThemeRegistry) List() []string {
	r.mu.RLock()
	defer r.mu.RUnlock()

	names := make([]string, 0, len(r.themes))
	for name := range r.themes {
		names = append(names, name)
	}
	sort.Strings(names)
	return names
}

// Info retrieves metadata about a registered theme.
//
// Returns the ThemeInfo and true if found, or an empty ThemeInfo and
// false if the theme is not registered.
//
// Example:
//
//	if info, ok := registry.Info("corporate"); ok {
//	    fmt.Printf("%s by %s: %s\n", info.Name, info.Author, info.Description)
//	}
func (r *ThemeRegistry) Info(name string) (ThemeInfo, bool) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	info, ok := r.info[name]
	return info, ok
}

// Count returns the number of registered themes.
func (r *ThemeRegistry) Count() int {
	r.mu.RLock()
	defer r.mu.RUnlock()

	return len(r.themes)
}

// Has returns true if a theme with the given name is registered.
func (r *ThemeRegistry) Has(name string) bool {
	r.mu.RLock()
	defer r.mu.RUnlock()

	_, ok := r.themes[name]
	return ok
}

// Clear removes all themes from the registry.
//
// This is primarily useful for testing.
func (r *ThemeRegistry) Clear() {
	r.mu.Lock()
	defer r.mu.Unlock()

	r.themes = make(map[string]*Theme)
	r.info = make(map[string]ThemeInfo)
}

// ListByVariant returns names of themes that support the given variant.
//
// The names are returned in sorted order.
//
// Example:
//
//	// Get all dark themes
//	darkThemes := registry.ListByVariant(VariantDark)
func (r *ThemeRegistry) ListByVariant(variant ThemeVariant) []string {
	r.mu.RLock()
	defer r.mu.RUnlock()

	var names []string
	for name, info := range r.info {
		if info.HasVariant(variant) {
			names = append(names, name)
		}
	}
	sort.Strings(names)
	return names
}

// globalRegistry is the default theme registry used by package-level functions.
var globalRegistry = NewThemeRegistry()

// Register adds a theme to the global registry.
//
// This is the primary way to register themes. Third-party packages
// typically call this in their init() function.
//
// Example:
//
//	// In your theme package's init()
//	func init() {
//	    theme.Register("my-theme", myTheme, theme.ThemeInfo{
//	        Name:        "My Theme",
//	        Description: "A custom theme",
//	    })
//	}
func Register(name string, theme *Theme, info ...ThemeInfo) {
	globalRegistry.Register(name, theme, info...)
}

// Unregister removes a theme from the global registry.
//
// Returns true if the theme was found and removed.
func Unregister(name string) bool {
	return globalRegistry.Unregister(name)
}

// Get retrieves a theme from the global registry by name.
//
// Returns the theme and true if found, or nil and false if not found.
//
// Example:
//
//	if theme, ok := theme.Get("corporate"); ok {
//	    app.SetTheme(theme)
//	}
func Get(name string) (*Theme, bool) {
	return globalRegistry.Get(name)
}

// MustGet retrieves a theme from the global registry or panics if not found.
//
// Use this for themes you know exist, such as the built-in "light" and "dark"
// themes.
//
// Example:
//
//	lightTheme := theme.MustGet("light")
//	darkTheme := theme.MustGet("dark")
func MustGet(name string) *Theme {
	return globalRegistry.MustGet(name)
}

// List returns the names of all themes in the global registry.
//
// The names are returned in sorted order.
//
// Example:
//
//	for _, name := range theme.List() {
//	    info, _ := theme.Info(name)
//	    fmt.Printf("%s: %s\n", name, info.Description)
//	}
func List() []string {
	return globalRegistry.List()
}

// Info retrieves metadata about a theme from the global registry.
//
// Returns the ThemeInfo and true if found, or an empty ThemeInfo and
// false if the theme is not registered.
func Info(name string) (ThemeInfo, bool) {
	return globalRegistry.Info(name)
}

// Count returns the number of themes in the global registry.
func Count() int {
	return globalRegistry.Count()
}

// Has returns true if a theme with the given name is in the global registry.
func Has(name string) bool {
	return globalRegistry.Has(name)
}

// ListByVariant returns names of themes in the global registry that support
// the given variant.
func ListByVariant(variant ThemeVariant) []string {
	return globalRegistry.ListByVariant(variant)
}

// GlobalRegistry returns the global theme registry.
//
// This is useful when you need direct access to the registry instance,
// for example when passing it to functions that accept a *ThemeRegistry.
func GlobalRegistry() *ThemeRegistry {
	return globalRegistry
}

// Register built-in themes.
func init() {
	Register("light", DefaultLight(), ThemeInfo{
		Name:        "Light",
		Description: "Default light theme with blue primary colors",
		Author:      "gogpu",
		Version:     "1.0.0",
		Variants:    []ThemeVariant{VariantLight},
	})

	Register("dark", DefaultDark(), ThemeInfo{
		Name:        "Dark",
		Description: "Default dark theme with light blue primary colors",
		Author:      "gogpu",
		Version:     "1.0.0",
		Variants:    []ThemeVariant{VariantDark},
	})

	Register("high-contrast", DefaultHighContrast(), ThemeInfo{
		Name:        "High Contrast",
		Description: "High contrast theme for accessibility (WCAG AAA)",
		Author:      "gogpu",
		Version:     "1.0.0",
		Variants:    []ThemeVariant{VariantLight},
	})

	Register("purple", Purple(), ThemeInfo{
		Name:        "Purple",
		Description: "Theme with purple primary colors",
		Author:      "gogpu",
		Version:     "1.0.0",
		Variants:    []ThemeVariant{VariantLight},
	})

	Register("green", Green(), ThemeInfo{
		Name:        "Green",
		Description: "Theme with green primary colors",
		Author:      "gogpu",
		Version:     "1.0.0",
		Variants:    []ThemeVariant{VariantLight},
	})

	Register("orange", Orange(), ThemeInfo{
		Name:        "Orange",
		Description: "Theme with orange primary colors",
		Author:      "gogpu",
		Version:     "1.0.0",
		Variants:    []ThemeVariant{VariantLight},
	})
}
