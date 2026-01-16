package plugin

import (
	"sync"
)

// AssetLoader provides methods for loading plugin resources.
//
// Plugins use the AssetLoader to load fonts, icons, and images
// during initialization. The loaded assets are then available
// to widgets and themes throughout the application.
//
// Implementations should be thread-safe as plugins may be
// initialized concurrently in some scenarios.
//
// Example:
//
//	func (p *MyPlugin) Init(ctx *PluginContext) error {
//	    // Load a font
//	    if err := ctx.Assets.LoadFont("roboto", robotoData); err != nil {
//	        return fmt.Errorf("failed to load roboto font: %w", err)
//	    }
//
//	    // Load icons
//	    if err := ctx.Assets.LoadIcon("add", addIconData); err != nil {
//	        return fmt.Errorf("failed to load add icon: %w", err)
//	    }
//
//	    // Load images
//	    if err := ctx.Assets.LoadImage("logo", logoData); err != nil {
//	        return fmt.Errorf("failed to load logo: %w", err)
//	    }
//
//	    return nil
//	}
type AssetLoader interface {
	// LoadFont registers a font with the given name.
	//
	// The data parameter should contain the raw font file data
	// (typically TTF or OTF format). The name is used to reference
	// the font in typography settings.
	//
	// Returns an error if the font data is invalid or loading fails.
	LoadFont(name string, data []byte) error

	// LoadIcon registers an icon with the given name.
	//
	// The data parameter should contain the icon image data
	// (typically PNG or SVG format). The name is used to reference
	// the icon in widgets.
	//
	// Returns an error if the icon data is invalid or loading fails.
	LoadIcon(name string, data []byte) error

	// LoadImage registers an image with the given name.
	//
	// The data parameter should contain the image data
	// (typically PNG, JPEG, or WebP format). The name is used to
	// reference the image in widgets and themes.
	//
	// Returns an error if the image data is invalid or loading fails.
	LoadImage(name string, data []byte) error
}

// noopAssetLoader is a no-op implementation of AssetLoader.
//
// It is used when no real asset loader is provided, allowing
// plugins to call asset loading methods without error.
type noopAssetLoader struct{}

// LoadFont implements AssetLoader.
func (n *noopAssetLoader) LoadFont(_ string, _ []byte) error {
	return nil
}

// LoadIcon implements AssetLoader.
func (n *noopAssetLoader) LoadIcon(_ string, _ []byte) error {
	return nil
}

// LoadImage implements AssetLoader.
func (n *noopAssetLoader) LoadImage(_ string, _ []byte) error {
	return nil
}

// Verify noopAssetLoader implements AssetLoader.
var _ AssetLoader = (*noopAssetLoader)(nil)

// MemoryAssetLoader is a simple in-memory implementation of AssetLoader.
//
// It stores all loaded assets in memory and provides methods to
// retrieve them. This is useful for testing and simple applications.
//
// MemoryAssetLoader is thread-safe.
type MemoryAssetLoader struct {
	mu     sync.RWMutex
	fonts  map[string][]byte
	icons  map[string][]byte
	images map[string][]byte
}

// NewMemoryAssetLoader creates a new MemoryAssetLoader.
func NewMemoryAssetLoader() *MemoryAssetLoader {
	return &MemoryAssetLoader{
		fonts:  make(map[string][]byte),
		icons:  make(map[string][]byte),
		images: make(map[string][]byte),
	}
}

// LoadFont implements AssetLoader.
func (m *MemoryAssetLoader) LoadFont(name string, data []byte) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	// Make a copy to avoid data races if caller modifies the slice
	copied := make([]byte, len(data))
	copy(copied, data)
	m.fonts[name] = copied

	return nil
}

// LoadIcon implements AssetLoader.
func (m *MemoryAssetLoader) LoadIcon(name string, data []byte) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	copied := make([]byte, len(data))
	copy(copied, data)
	m.icons[name] = copied

	return nil
}

// LoadImage implements AssetLoader.
func (m *MemoryAssetLoader) LoadImage(name string, data []byte) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	copied := make([]byte, len(data))
	copy(copied, data)
	m.images[name] = copied

	return nil
}

// GetFont retrieves a loaded font by name.
//
// Returns the font data and true if found, or nil and false if not found.
func (m *MemoryAssetLoader) GetFont(name string) ([]byte, bool) {
	m.mu.RLock()
	defer m.mu.RUnlock()

	data, ok := m.fonts[name]
	return data, ok
}

// GetIcon retrieves a loaded icon by name.
//
// Returns the icon data and true if found, or nil and false if not found.
func (m *MemoryAssetLoader) GetIcon(name string) ([]byte, bool) {
	m.mu.RLock()
	defer m.mu.RUnlock()

	data, ok := m.icons[name]
	return data, ok
}

// GetImage retrieves a loaded image by name.
//
// Returns the image data and true if found, or nil and false if not found.
func (m *MemoryAssetLoader) GetImage(name string) ([]byte, bool) {
	m.mu.RLock()
	defer m.mu.RUnlock()

	data, ok := m.images[name]
	return data, ok
}

// FontCount returns the number of loaded fonts.
func (m *MemoryAssetLoader) FontCount() int {
	m.mu.RLock()
	defer m.mu.RUnlock()
	return len(m.fonts)
}

// IconCount returns the number of loaded icons.
func (m *MemoryAssetLoader) IconCount() int {
	m.mu.RLock()
	defer m.mu.RUnlock()
	return len(m.icons)
}

// ImageCount returns the number of loaded images.
func (m *MemoryAssetLoader) ImageCount() int {
	m.mu.RLock()
	defer m.mu.RUnlock()
	return len(m.images)
}

// Clear removes all loaded assets.
func (m *MemoryAssetLoader) Clear() {
	m.mu.Lock()
	defer m.mu.Unlock()

	m.fonts = make(map[string][]byte)
	m.icons = make(map[string][]byte)
	m.images = make(map[string][]byte)
}

// Verify MemoryAssetLoader implements AssetLoader.
var _ AssetLoader = (*MemoryAssetLoader)(nil)
