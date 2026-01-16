package registry

import (
	"errors"
	"fmt"
	"sort"
	"sync"

	"github.com/gogpu/ui/event"
	"github.com/gogpu/ui/geometry"
	"github.com/gogpu/ui/widget"
)

// Widget is an alias for the widget.Widget interface.
//
// This allows third-party packages to use registry.Widget without
// importing the widget package directly, simplifying the import graph.
type Widget = widget.Widget

// Context is an alias for the widget.Context interface.
type Context = widget.Context

// Canvas is an alias for the widget.Canvas interface.
type Canvas = widget.Canvas

// Event is an alias for the event.Event interface.
type Event = event.Event

// Constraints is an alias for the geometry.Constraints type.
type Constraints = geometry.Constraints

// Size is an alias for the geometry.Size type.
type Size = geometry.Size

// Category represents the category of a widget for organization and filtering.
type Category string

// Widget categories for organization and discoverability.
const (
	// CategoryInput represents interactive widgets like buttons, text fields, checkboxes.
	CategoryInput Category = "input"

	// CategoryDisplay represents read-only widgets like labels, images, progress bars.
	CategoryDisplay Category = "display"

	// CategoryContainer represents widgets that contain other widgets like panels, scroll views.
	CategoryContainer Category = "container"

	// CategoryCustom represents user-defined widgets that don't fit other categories.
	CategoryCustom Category = "custom"
)

// String returns the string representation of the category.
func (c Category) String() string {
	return string(c)
}

// IsValid returns true if the category is one of the predefined categories.
func (c Category) IsValid() bool {
	switch c {
	case CategoryInput, CategoryDisplay, CategoryContainer, CategoryCustom:
		return true
	default:
		return false
	}
}

// WidgetFactory creates a widget from configuration.
//
// The config parameter is a map of configuration options specific to the widget type.
// Factories should validate required parameters and provide sensible defaults.
//
// Returns an error if the configuration is invalid or the widget cannot be created.
type WidgetFactory func(config map[string]any) (Widget, error)

// WidgetInfo describes a registered widget.
//
// This metadata is used for widget discovery, documentation, and tooling.
type WidgetInfo struct {
	// Name is the unique identifier used for registration.
	Name string

	// Description provides a human-readable description of the widget.
	Description string

	// Category organizes widgets for filtering and discovery.
	Category Category

	// Version is the semantic version of the widget (e.g., "1.0.0").
	Version string
}

// Validate checks if the WidgetInfo has all required fields.
func (i WidgetInfo) Validate() error {
	if i.Name == "" {
		return errors.New("widget info: name is required")
	}
	return nil
}

// Common errors returned by registry operations.
var (
	// ErrWidgetNotFound is returned when a widget is not registered.
	ErrWidgetNotFound = errors.New("widget not found")

	// ErrWidgetExists is returned when attempting to register a duplicate widget.
	ErrWidgetExists = errors.New("widget already registered")

	// ErrNilFactory is returned when a nil factory is provided.
	ErrNilFactory = errors.New("factory cannot be nil")

	// ErrEmptyName is returned when an empty name is provided.
	ErrEmptyName = errors.New("widget name cannot be empty")
)

// WidgetRegistry manages widget registration and creation.
//
// The registry maintains a thread-safe collection of widget factories
// and metadata, allowing dynamic widget creation by name.
//
// Use the package-level functions (RegisterWidget, CreateWidget, etc.)
// for the global registry, or create a custom registry for testing
// or isolated use cases.
type WidgetRegistry struct {
	mu      sync.RWMutex
	widgets map[string]WidgetFactory
	info    map[string]WidgetInfo
}

// NewWidgetRegistry creates a new empty WidgetRegistry.
//
// Use this to create an isolated registry for testing or when you need
// multiple independent registries.
func NewWidgetRegistry() *WidgetRegistry {
	return &WidgetRegistry{
		widgets: make(map[string]WidgetFactory),
		info:    make(map[string]WidgetInfo),
	}
}

// Register adds a widget factory to the registry.
//
// The name must be unique within the registry. If info is provided,
// it stores metadata about the widget for discovery purposes.
//
// Returns ErrEmptyName if name is empty, ErrNilFactory if factory is nil,
// or ErrWidgetExists if a widget with the same name is already registered.
func (r *WidgetRegistry) Register(name string, factory WidgetFactory, info ...WidgetInfo) error {
	if name == "" {
		return ErrEmptyName
	}
	if factory == nil {
		return ErrNilFactory
	}

	r.mu.Lock()
	defer r.mu.Unlock()

	if _, exists := r.widgets[name]; exists {
		return fmt.Errorf("%w: %s", ErrWidgetExists, name)
	}

	r.widgets[name] = factory

	// Store info if provided
	if len(info) > 0 {
		widgetInfo := info[0]
		// Ensure the name in info matches the registration name
		if widgetInfo.Name == "" {
			widgetInfo.Name = name
		}
		r.info[name] = widgetInfo
	} else {
		// Create minimal info
		r.info[name] = WidgetInfo{Name: name}
	}

	return nil
}

// MustRegister is like Register but panics if registration fails.
//
// This is intended for use in init() functions where registration
// failure indicates a programming error.
func (r *WidgetRegistry) MustRegister(name string, factory WidgetFactory, info ...WidgetInfo) {
	if err := r.Register(name, factory, info...); err != nil {
		panic(fmt.Sprintf("registry: failed to register widget %q: %v", name, err))
	}
}

// Unregister removes a widget from the registry.
//
// Returns ErrWidgetNotFound if the widget is not registered.
func (r *WidgetRegistry) Unregister(name string) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	if _, exists := r.widgets[name]; !exists {
		return fmt.Errorf("%w: %s", ErrWidgetNotFound, name)
	}

	delete(r.widgets, name)
	delete(r.info, name)
	return nil
}

// Create creates a widget by name using its registered factory.
//
// The config parameter is passed to the widget's factory function.
// Pass nil for widgets that don't require configuration.
//
// Returns ErrWidgetNotFound if no widget is registered with the given name.
func (r *WidgetRegistry) Create(name string, config map[string]any) (Widget, error) {
	r.mu.RLock()
	factory, exists := r.widgets[name]
	r.mu.RUnlock()

	if !exists {
		return nil, fmt.Errorf("%w: %s", ErrWidgetNotFound, name)
	}

	return factory(config)
}

// Has returns true if a widget with the given name is registered.
func (r *WidgetRegistry) Has(name string) bool {
	r.mu.RLock()
	defer r.mu.RUnlock()
	_, exists := r.widgets[name]
	return exists
}

// Info returns the WidgetInfo for a registered widget.
//
// Returns false as the second value if the widget is not registered.
func (r *WidgetRegistry) Info(name string) (WidgetInfo, bool) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	info, exists := r.info[name]
	return info, exists
}

// List returns the names of all registered widgets in sorted order.
func (r *WidgetRegistry) List() []string {
	r.mu.RLock()
	defer r.mu.RUnlock()

	names := make([]string, 0, len(r.widgets))
	for name := range r.widgets {
		names = append(names, name)
	}
	sort.Strings(names)
	return names
}

// ListByCategory returns the names of widgets in the specified category.
//
// The returned list is sorted alphabetically.
func (r *WidgetRegistry) ListByCategory(category Category) []string {
	r.mu.RLock()
	defer r.mu.RUnlock()

	var names []string
	for name, info := range r.info {
		if info.Category == category {
			names = append(names, name)
		}
	}
	sort.Strings(names)
	return names
}

// Count returns the number of registered widgets.
func (r *WidgetRegistry) Count() int {
	r.mu.RLock()
	defer r.mu.RUnlock()
	return len(r.widgets)
}

// Clear removes all registered widgets.
//
// This is primarily useful for testing to reset the registry state.
func (r *WidgetRegistry) Clear() {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.widgets = make(map[string]WidgetFactory)
	r.info = make(map[string]WidgetInfo)
}

// AllInfo returns a copy of all registered widget information.
//
// The returned slice is sorted by widget name.
func (r *WidgetRegistry) AllInfo() []WidgetInfo {
	r.mu.RLock()
	defer r.mu.RUnlock()

	infos := make([]WidgetInfo, 0, len(r.info))
	for _, info := range r.info {
		infos = append(infos, info)
	}

	// Sort by name for consistent ordering
	sort.Slice(infos, func(i, j int) bool {
		return infos[i].Name < infos[j].Name
	})

	return infos
}

// Global registry instance
var globalWidgetRegistry = NewWidgetRegistry()

// RegisterWidget adds a widget factory to the global registry.
//
// This is the primary function for registering widgets. Call this from
// your package's init() function to enable automatic registration on import.
//
// Example:
//
//	func init() {
//	    registry.RegisterWidget("my-widget", NewMyWidget, registry.WidgetInfo{
//	        Name:        "my-widget",
//	        Description: "A custom widget",
//	        Category:    registry.CategoryCustom,
//	        Version:     "1.0.0",
//	    })
//	}
func RegisterWidget(name string, factory WidgetFactory, info ...WidgetInfo) error {
	return globalWidgetRegistry.Register(name, factory, info...)
}

// MustRegisterWidget is like RegisterWidget but panics on failure.
//
// Use this in init() functions where registration failure indicates
// a programming error that should be caught during development.
func MustRegisterWidget(name string, factory WidgetFactory, info ...WidgetInfo) {
	globalWidgetRegistry.MustRegister(name, factory, info...)
}

// UnregisterWidget removes a widget from the global registry.
//
// Returns an error if the widget is not registered.
func UnregisterWidget(name string) error {
	return globalWidgetRegistry.Unregister(name)
}

// CreateWidget creates a widget by name from the global registry.
//
// This is the primary function for creating widgets dynamically.
// The config parameter is passed to the widget's factory function.
//
// Example:
//
//	widget, err := registry.CreateWidget("button", map[string]any{
//	    "label": "Click Me",
//	    "onClick": func() { fmt.Println("Clicked!") },
//	})
func CreateWidget(name string, config map[string]any) (Widget, error) {
	return globalWidgetRegistry.Create(name, config)
}

// HasWidget returns true if a widget is registered in the global registry.
func HasWidget(name string) bool {
	return globalWidgetRegistry.Has(name)
}

// GetWidgetInfo returns information about a registered widget.
//
// Returns false as the second value if the widget is not registered.
func GetWidgetInfo(name string) (WidgetInfo, bool) {
	return globalWidgetRegistry.Info(name)
}

// ListWidgets returns all registered widget names from the global registry.
//
// The returned list is sorted alphabetically.
func ListWidgets() []string {
	return globalWidgetRegistry.List()
}

// ListWidgetsByCategory returns widgets in a specific category from the global registry.
//
// The returned list is sorted alphabetically.
func ListWidgetsByCategory(category Category) []string {
	return globalWidgetRegistry.ListByCategory(category)
}

// WidgetCount returns the number of registered widgets in the global registry.
func WidgetCount() int {
	return globalWidgetRegistry.Count()
}

// AllWidgetInfo returns information about all registered widgets.
//
// The returned slice is sorted by widget name.
func AllWidgetInfo() []WidgetInfo {
	return globalWidgetRegistry.AllInfo()
}

// ClearGlobalRegistry removes all widgets from the global registry.
//
// WARNING: This is primarily intended for testing. Use with caution
// in production code as it will break any code relying on registered widgets.
func ClearGlobalRegistry() {
	globalWidgetRegistry.Clear()
}

// GlobalRegistry returns the global widget registry instance.
//
// This is useful for advanced use cases where direct access to the
// registry is needed, such as iterating over all widgets or
// implementing custom registration logic.
func GlobalRegistry() *WidgetRegistry {
	return globalWidgetRegistry
}
