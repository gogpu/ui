package icon

import "sync"

// Registry is a thread-safe named icon registry.
//
// Use [DefaultRegistry] to access the global registry, which is pre-populated
// with all built-in icons. Create a custom registry with [NewRegistry] for
// isolated icon sets.
type Registry struct {
	mu    sync.RWMutex
	icons map[string]IconData
}

// NewRegistry creates an empty icon registry.
func NewRegistry() *Registry {
	return &Registry{
		icons: make(map[string]IconData),
	}
}

// Register adds an icon to the registry, keyed by its Name field.
//
// If an icon with the same name already exists, it is replaced.
func (r *Registry) Register(icon IconData) {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.icons[icon.Name] = icon
}

// Get returns the icon with the given name and true if found, or a zero
// IconData and false otherwise.
func (r *Registry) Get(name string) (IconData, bool) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	ic, ok := r.icons[name]
	return ic, ok
}

// Names returns all registered icon names in no particular order.
func (r *Registry) Names() []string {
	r.mu.RLock()
	defer r.mu.RUnlock()
	names := make([]string, 0, len(r.icons))
	for name := range r.icons {
		names = append(names, name)
	}
	return names
}

// Len returns the number of icons in the registry.
func (r *Registry) Len() int {
	r.mu.RLock()
	defer r.mu.RUnlock()
	return len(r.icons)
}

// defaultRegistry is the global icon registry pre-populated with built-in icons.
var defaultRegistry *Registry

func init() {
	defaultRegistry = NewRegistry()
	builtins := []IconData{
		Close, Check, ChevronDown, ChevronRight, Search,
		Settings, Menu, ArrowBack, Add, Delete,
	}
	for _, ic := range builtins {
		defaultRegistry.Register(ic)
	}
}

// DefaultRegistry returns the global icon registry.
//
// The default registry is pre-populated with all built-in icons. Additional
// icons can be registered at any time using [Registry.Register].
func DefaultRegistry() *Registry {
	return defaultRegistry
}
