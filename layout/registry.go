package layout

import (
	"sort"
	"sync"
)

// Registry is a thread-safe registry for layout algorithms.
//
// The registry allows layout algorithms to be registered by name and
// looked up at runtime. This enables third-party developers to create
// custom layouts that can be used alongside built-in layouts.
//
// A global registry is provided via the package-level functions
// [Register], [Get], [MustGet], and [List].
type Registry struct {
	mu         sync.RWMutex
	algorithms map[string]LayoutAlgorithm
}

// NewRegistry creates a new empty registry.
func NewRegistry() *Registry {
	return &Registry{
		algorithms: make(map[string]LayoutAlgorithm),
	}
}

// Register adds a layout algorithm to the registry.
//
// If an algorithm with the same name already exists, it will be replaced.
// The algorithm's Name() method is used as the key.
func (r *Registry) Register(algorithm LayoutAlgorithm) {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.algorithms[algorithm.Name()] = algorithm
}

// RegisterWithName adds a layout algorithm with an explicit name.
//
// This allows registering the same algorithm under different names
// or using a name different from the algorithm's Name() method.
func (r *Registry) RegisterWithName(name string, algorithm LayoutAlgorithm) {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.algorithms[name] = algorithm
}

// Get retrieves a layout algorithm by name.
//
// Returns the algorithm and true if found, or nil and false if not found.
func (r *Registry) Get(name string) (LayoutAlgorithm, bool) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	algo, ok := r.algorithms[name]
	return algo, ok
}

// MustGet retrieves a layout algorithm by name, panicking if not found.
//
// Use this only when you're certain the algorithm exists (e.g., built-in layouts).
func (r *Registry) MustGet(name string) LayoutAlgorithm {
	algo, ok := r.Get(name)
	if !ok {
		panic("layout: algorithm not found: " + name)
	}
	return algo
}

// Has returns true if an algorithm with the given name is registered.
func (r *Registry) Has(name string) bool {
	r.mu.RLock()
	defer r.mu.RUnlock()
	_, ok := r.algorithms[name]
	return ok
}

// Unregister removes a layout algorithm from the registry.
//
// Returns true if the algorithm was removed, false if it wasn't registered.
func (r *Registry) Unregister(name string) bool {
	r.mu.Lock()
	defer r.mu.Unlock()
	if _, ok := r.algorithms[name]; ok {
		delete(r.algorithms, name)
		return true
	}
	return false
}

// List returns the names of all registered algorithms in sorted order.
func (r *Registry) List() []string {
	r.mu.RLock()
	defer r.mu.RUnlock()

	names := make([]string, 0, len(r.algorithms))
	for name := range r.algorithms {
		names = append(names, name)
	}
	sort.Strings(names)
	return names
}

// Count returns the number of registered algorithms.
func (r *Registry) Count() int {
	r.mu.RLock()
	defer r.mu.RUnlock()
	return len(r.algorithms)
}

// Clear removes all registered algorithms.
func (r *Registry) Clear() {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.algorithms = make(map[string]LayoutAlgorithm)
}

// Clone creates a copy of the registry with all registered algorithms.
func (r *Registry) Clone() *Registry {
	r.mu.RLock()
	defer r.mu.RUnlock()

	clone := NewRegistry()
	for name, algo := range r.algorithms {
		clone.algorithms[name] = algo
	}
	return clone
}

// globalRegistry is the default registry used by package-level functions.
var globalRegistry = NewRegistry()

// Register adds a layout algorithm to the global registry.
//
// This is typically called from init() functions:
//
//	func init() {
//	    layout.Register(&MyCustomLayout{})
//	}
func Register(algorithm LayoutAlgorithm) {
	globalRegistry.Register(algorithm)
}

// RegisterWithName adds a layout algorithm with an explicit name to the global registry.
func RegisterWithName(name string, algorithm LayoutAlgorithm) {
	globalRegistry.RegisterWithName(name, algorithm)
}

// Get retrieves a layout algorithm from the global registry.
func Get(name string) (LayoutAlgorithm, bool) {
	return globalRegistry.Get(name)
}

// MustGet retrieves a layout algorithm from the global registry, panicking if not found.
func MustGet(name string) LayoutAlgorithm {
	return globalRegistry.MustGet(name)
}

// Has returns true if an algorithm is registered in the global registry.
func Has(name string) bool {
	return globalRegistry.Has(name)
}

// Unregister removes a layout algorithm from the global registry.
func Unregister(name string) bool {
	return globalRegistry.Unregister(name)
}

// List returns all algorithm names from the global registry.
func List() []string {
	return globalRegistry.List()
}

// Count returns the number of algorithms in the global registry.
func Count() int {
	return globalRegistry.Count()
}

// GlobalRegistry returns the global registry instance.
//
// This can be used when you need direct access to the registry,
// for example, to clone it or clear it for testing.
func GlobalRegistry() *Registry {
	return globalRegistry
}
