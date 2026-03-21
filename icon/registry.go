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

// MultiColorRegistry is a thread-safe named registry for multi-color icons.
//
// Use [DefaultMultiColorRegistry] to access the global registry, which is
// pre-populated with all built-in multi-color icons. Create a custom
// registry with [NewMultiColorRegistry] for isolated icon sets.
type MultiColorRegistry struct {
	mu    sync.RWMutex
	icons map[string]MultiColorIcon
}

// NewMultiColorRegistry creates an empty multi-color icon registry.
func NewMultiColorRegistry() *MultiColorRegistry {
	return &MultiColorRegistry{
		icons: make(map[string]MultiColorIcon),
	}
}

// Register adds a multi-color icon to the registry, keyed by its Name field.
//
// If an icon with the same name already exists, it is replaced.
func (r *MultiColorRegistry) Register(icon MultiColorIcon) {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.icons[icon.Name] = icon
}

// Get returns the multi-color icon with the given name and true if found,
// or a zero MultiColorIcon and false otherwise.
func (r *MultiColorRegistry) Get(name string) (MultiColorIcon, bool) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	ic, ok := r.icons[name]
	return ic, ok
}

// Names returns all registered multi-color icon names in no particular order.
func (r *MultiColorRegistry) Names() []string {
	r.mu.RLock()
	defer r.mu.RUnlock()
	names := make([]string, 0, len(r.icons))
	for name := range r.icons {
		names = append(names, name)
	}
	return names
}

// Len returns the number of multi-color icons in the registry.
func (r *MultiColorRegistry) Len() int {
	r.mu.RLock()
	defer r.mu.RUnlock()
	return len(r.icons)
}

// defaultMultiColorRegistry is the global multi-color icon registry.
var defaultMultiColorRegistry *MultiColorRegistry

func init() {
	defaultRegistry = NewRegistry()
	builtins := []IconData{
		Close, Check, ChevronDown, ChevronRight, Search,
		Settings, Menu, ArrowBack, Add, Delete,
		Play, Stop, Pause, Debug, Gear, Filter,
		FolderOpen, FolderClosed, Terminal, Refresh, Plus, Minus,
	}
	for _, ic := range builtins {
		defaultRegistry.Register(ic)
	}

	defaultMultiColorRegistry = NewMultiColorRegistry()
	multiColorBuiltins := []MultiColorIcon{
		FileGo, FileJSON, FileYAML, FileMD, FileTest,
		FileConfig, FileImage, FileGeneric,
		GitBranch, GitCommit, GitMerge, GitPR, GitModified,
	}
	for _, ic := range multiColorBuiltins {
		defaultMultiColorRegistry.Register(ic)
	}
}

// DefaultMultiColorRegistry returns the global multi-color icon registry.
//
// The default registry is pre-populated with all built-in multi-color icons
// (file type icons and VCS icons). Additional icons can be registered at
// any time using [MultiColorRegistry.Register].
func DefaultMultiColorRegistry() *MultiColorRegistry {
	return defaultMultiColorRegistry
}

// DefaultRegistry returns the global icon registry.
//
// The default registry is pre-populated with all built-in icons. Additional
// icons can be registered at any time using [Registry.Register].
func DefaultRegistry() *Registry {
	return defaultRegistry
}
