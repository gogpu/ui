package plugin

import (
	"fmt"
	"sort"
	"sync"
)

// PluginManager manages plugin registration, initialization, and shutdown.
//
// The plugin manager maintains a registry of plugins and handles their
// lifecycle. It resolves dependencies between plugins and initializes
// them in the correct order.
//
// For most use cases, use the package-level functions ([Register],
// [Initialize], [Shutdown], [List], [Info]) which operate on a global
// plugin manager.
//
// Create a custom PluginManager for testing or isolated use cases.
type PluginManager struct {
	mu          sync.RWMutex
	plugins     map[string]Plugin
	info        map[string]PluginInfo
	loadOrder   []string
	initialized bool
	context     *PluginContext
}

// NewPluginManager creates a new PluginManager.
//
// The manager starts empty and uninitialized. Register plugins with
// [PluginManager.Register] and then call [PluginManager.Initialize].
func NewPluginManager() *PluginManager {
	return &PluginManager{
		plugins: make(map[string]Plugin),
		info:    make(map[string]PluginInfo),
	}
}

// Register adds a plugin to the manager.
//
// The plugin's Name() method is used as the identifier. If info is
// provided, it stores metadata about the plugin for discovery purposes.
//
// Returns ErrNilPlugin if plugin is nil, ErrEmptyName if the plugin's
// name is empty, or ErrPluginExists if a plugin with the same name
// is already registered.
//
// Note: Registration does not initialize the plugin. Call [Initialize]
// after all plugins are registered.
func (m *PluginManager) Register(plugin Plugin, info ...PluginInfo) error {
	if plugin == nil {
		return ErrNilPlugin
	}

	name := plugin.Name()
	if name == "" {
		return ErrEmptyName
	}

	m.mu.Lock()
	defer m.mu.Unlock()

	if m.initialized {
		return fmt.Errorf("cannot register plugin %q: %w", name, ErrAlreadyInitialized)
	}

	if _, exists := m.plugins[name]; exists {
		return fmt.Errorf("%w: %s", ErrPluginExists, name)
	}

	m.plugins[name] = plugin

	// Store info
	if len(info) > 0 {
		pluginInfo := info[0]
		pluginInfo.Name = name
		pluginInfo.Version = plugin.Version()
		pluginInfo.Dependencies = plugin.Dependencies()
		m.info[name] = pluginInfo
	} else {
		// Create minimal info from plugin
		m.info[name] = PluginInfo{
			Name:         name,
			Version:      plugin.Version(),
			Dependencies: plugin.Dependencies(),
		}
	}

	return nil
}

// MustRegister is like Register but panics if registration fails.
//
// This is intended for use in init() functions where registration
// failure indicates a programming error.
func (m *PluginManager) MustRegister(plugin Plugin, info ...PluginInfo) {
	if err := m.Register(plugin, info...); err != nil {
		panic(fmt.Sprintf("plugin: failed to register %q: %v", plugin.Name(), err))
	}
}

// Initialize loads all registered plugins in dependency order.
//
// This method:
// 1. Validates all dependencies are registered
// 2. Checks version constraints
// 3. Resolves initialization order via topological sort
// 4. Calls Init() on each plugin in order
//
// Returns an error if:
// - Already initialized ([ErrAlreadyInitialized])
// - A dependency is not registered ([ErrDependencyNotFound])
// - A version constraint is not satisfied ([ErrVersionMismatch])
// - Circular dependencies are detected ([ErrCircularDependency])
// - A plugin's Init() method returns an error
func (m *PluginManager) Initialize() error {
	return m.InitializeWithContext(NewDefaultPluginContext())
}

// InitializeWithContext loads all plugins using a custom context.
//
// This allows using custom registries or asset loaders instead of
// the global ones.
func (m *PluginManager) InitializeWithContext(ctx *PluginContext) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	if m.initialized {
		return ErrAlreadyInitialized
	}

	// Validate dependencies and version constraints
	if err := m.validateDependencies(); err != nil {
		return err
	}

	// Build dependency graph and get initialization order
	order, err := m.resolveDependencyOrder()
	if err != nil {
		return err
	}

	// Initialize plugins in order
	for _, name := range order {
		plugin := m.plugins[name]
		if err := plugin.Init(ctx); err != nil {
			return fmt.Errorf("plugin %q initialization failed: %w", name, err)
		}
	}

	m.loadOrder = order
	m.context = ctx
	m.initialized = true

	return nil
}

// validateDependencies checks that all dependencies are registered and
// version constraints are satisfied.
func (m *PluginManager) validateDependencies() error {
	for name, plugin := range m.plugins {
		for _, dep := range plugin.Dependencies() {
			// Check dependency is registered
			depPlugin, exists := m.plugins[dep.Name]
			if !exists {
				return fmt.Errorf("%w: plugin %q requires %q", ErrDependencyNotFound, name, dep.Name)
			}

			// Check version constraint
			if dep.Version != "" {
				satisfied, err := checkVersionConstraint(depPlugin.Version(), dep.Version)
				if err != nil {
					return fmt.Errorf("plugin %q: invalid version constraint for %q: %w", name, dep.Name, err)
				}
				if !satisfied {
					return fmt.Errorf("%w: plugin %q requires %q %s, got %s",
						ErrVersionMismatch, name, dep.Name, dep.Version, depPlugin.Version())
				}
			}
		}
	}

	return nil
}

// resolveDependencyOrder builds a dependency graph and returns the
// initialization order via topological sort.
func (m *PluginManager) resolveDependencyOrder() ([]string, error) {
	graph := newDependencyGraph()

	for name, plugin := range m.plugins {
		pluginDeps := plugin.Dependencies()
		deps := make([]string, 0, len(pluginDeps))
		for _, dep := range pluginDeps {
			deps = append(deps, dep.Name)
		}
		graph.addNode(name, deps)
	}

	return graph.topologicalSort()
}

// Shutdown unloads all plugins in reverse initialization order.
//
// This method calls Shutdown() on each plugin, starting with those
// that were initialized last (to respect dependency order).
//
// Returns an error if plugins were not initialized, or if any plugin's
// Shutdown() method returns an error. All plugins are attempted to
// shut down even if some fail.
func (m *PluginManager) Shutdown() error {
	m.mu.Lock()
	defer m.mu.Unlock()

	if !m.initialized {
		return ErrNotInitialized
	}

	var errs []error

	// Shutdown in reverse order
	for i := len(m.loadOrder) - 1; i >= 0; i-- {
		name := m.loadOrder[i]
		plugin := m.plugins[name]
		if err := plugin.Shutdown(); err != nil {
			errs = append(errs, fmt.Errorf("plugin %q shutdown failed: %w", name, err))
		}
	}

	m.initialized = false
	m.loadOrder = nil
	m.context = nil

	if len(errs) > 0 {
		// Return first error (could combine them, but this is simpler)
		return errs[0]
	}

	return nil
}

// List returns the names of all registered plugins in sorted order.
func (m *PluginManager) List() []string {
	m.mu.RLock()
	defer m.mu.RUnlock()

	names := make([]string, 0, len(m.plugins))
	for name := range m.plugins {
		names = append(names, name)
	}
	sort.Strings(names)
	return names
}

// Info returns information about a registered plugin.
//
// Returns false as the second value if the plugin is not registered.
func (m *PluginManager) Info(name string) (PluginInfo, bool) {
	m.mu.RLock()
	defer m.mu.RUnlock()

	info, exists := m.info[name]
	return info, exists
}

// Has returns true if a plugin with the given name is registered.
func (m *PluginManager) Has(name string) bool {
	m.mu.RLock()
	defer m.mu.RUnlock()

	_, exists := m.plugins[name]
	return exists
}

// Get returns a plugin by name.
//
// Returns false as the second value if the plugin is not registered.
func (m *PluginManager) Get(name string) (Plugin, bool) {
	m.mu.RLock()
	defer m.mu.RUnlock()

	plugin, exists := m.plugins[name]
	return plugin, exists
}

// Count returns the number of registered plugins.
func (m *PluginManager) Count() int {
	m.mu.RLock()
	defer m.mu.RUnlock()
	return len(m.plugins)
}

// IsInitialized returns true if plugins have been initialized.
func (m *PluginManager) IsInitialized() bool {
	m.mu.RLock()
	defer m.mu.RUnlock()
	return m.initialized
}

// LoadOrder returns the order in which plugins were initialized.
//
// Returns nil if plugins have not been initialized.
func (m *PluginManager) LoadOrder() []string {
	m.mu.RLock()
	defer m.mu.RUnlock()

	if m.loadOrder == nil {
		return nil
	}

	// Return a copy to prevent modification
	order := make([]string, len(m.loadOrder))
	copy(order, m.loadOrder)
	return order
}

// Clear removes all registered plugins.
//
// If plugins are initialized, they are shut down first.
// This is primarily useful for testing.
func (m *PluginManager) Clear() error {
	m.mu.Lock()

	if m.initialized {
		m.mu.Unlock()
		if err := m.Shutdown(); err != nil {
			return err
		}
		m.mu.Lock()
	}

	m.plugins = make(map[string]Plugin)
	m.info = make(map[string]PluginInfo)
	m.loadOrder = nil
	m.context = nil

	m.mu.Unlock()
	return nil
}

// AllInfo returns information about all registered plugins.
//
// The returned slice is sorted by plugin name.
func (m *PluginManager) AllInfo() []PluginInfo {
	m.mu.RLock()
	defer m.mu.RUnlock()

	infos := make([]PluginInfo, 0, len(m.info))
	for _, info := range m.info {
		infos = append(infos, info)
	}

	sort.Slice(infos, func(i, j int) bool {
		return infos[i].Name < infos[j].Name
	})

	return infos
}

// globalManager is the default plugin manager used by package-level functions.
var globalManager = NewPluginManager()

// Register adds a plugin to the global manager.
//
// This is typically called from init() functions:
//
//	func init() {
//	    plugin.Register(&MyPlugin{}, plugin.PluginInfo{
//	        Name:        "my-plugin",
//	        Description: "My custom plugin",
//	    })
//	}
func Register(plugin Plugin, info ...PluginInfo) error {
	return globalManager.Register(plugin, info...)
}

// MustRegister is like Register but panics if registration fails.
//
// Use this in init() functions where registration failure indicates
// a programming error.
func MustRegister(plugin Plugin, info ...PluginInfo) {
	globalManager.MustRegister(plugin, info...)
}

// Initialize loads all registered plugins from the global manager.
//
// Call this at application startup, after all plugins have been
// registered (typically via init() functions):
//
//	func main() {
//	    if err := plugin.Initialize(); err != nil {
//	        log.Fatal(err)
//	    }
//	    defer plugin.Shutdown()
//
//	    // Application code...
//	}
func Initialize() error {
	return globalManager.Initialize()
}

// InitializeWithContext loads plugins using a custom context.
func InitializeWithContext(ctx *PluginContext) error {
	return globalManager.InitializeWithContext(ctx)
}

// Shutdown unloads all plugins from the global manager.
//
// Call this at application shutdown to allow plugins to release
// resources.
func Shutdown() error {
	return globalManager.Shutdown()
}

// List returns the names of all registered plugins.
func List() []string {
	return globalManager.List()
}

// Info returns information about a registered plugin.
func Info(name string) (PluginInfo, bool) {
	return globalManager.Info(name)
}

// Has returns true if a plugin with the given name is registered.
func Has(name string) bool {
	return globalManager.Has(name)
}

// Get returns a plugin by name from the global manager.
func Get(name string) (Plugin, bool) {
	return globalManager.Get(name)
}

// Count returns the number of registered plugins.
func Count() int {
	return globalManager.Count()
}

// IsInitialized returns true if the global manager has been initialized.
func IsInitialized() bool {
	return globalManager.IsInitialized()
}

// LoadOrder returns the initialization order of plugins.
func LoadOrder() []string {
	return globalManager.LoadOrder()
}

// AllInfo returns information about all registered plugins.
func AllInfo() []PluginInfo {
	return globalManager.AllInfo()
}

// GlobalManager returns the global plugin manager.
//
// This is useful for advanced use cases where direct access to the
// manager is needed.
func GlobalManager() *PluginManager {
	return globalManager
}

// ClearGlobalManager removes all plugins from the global manager.
//
// WARNING: This is primarily intended for testing. Use with caution
// in production code.
func ClearGlobalManager() error {
	return globalManager.Clear()
}
