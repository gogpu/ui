package render

import (
	"fmt"
	"sync"

	"github.com/gogpu/gg/text"
	"github.com/gogpu/ui/internal/render/fonts"
	"github.com/gogpu/ui/theme/font"
)

// defaultFontFamily is the embedded default font family name.
// Inter is pre-registered in every [FontRegistry] and used as fallback.
const defaultFontFamily = "Inter"

// FontRegistry manages font families and creates cached [text.FontSource]
// instances for rendering. It bridges [font.Registry] (metadata + CSS weight
// matching) with gg's [text.FontSource] (rendering).
//
// FontRegistry is safe for concurrent use.
//
// Resolution chain:
//  1. Exact match: family + weight + style
//  2. CSS weight matching via [font.Registry.Resolve]
//  3. Fallback to embedded Inter (Regular or Bold)
//
// FontSource instances are cached by the raw font data pointer to avoid
// re-parsing identical font files.
type FontRegistry struct {
	mu sync.RWMutex

	// metadata holds font family metadata and CSS weight matching logic.
	metadata *font.Registry

	// sources caches *text.FontSource by a composite key of
	// (family, weight, style) that was resolved.
	sources map[sourceKey]*text.FontSource
}

// sourceKey identifies a resolved font source in the cache.
type sourceKey struct {
	family string
	weight font.Weight
	style  font.Style
}

// NewFontRegistry creates a new FontRegistry with embedded Inter as the
// default font family.
func NewFontRegistry() *FontRegistry {
	r := &FontRegistry{
		metadata: font.NewRegistry(),
		sources:  make(map[sourceKey]*text.FontSource),
	}

	// Pre-register embedded Inter font as the default family so that
	// unset FontFamily fields resolve to Inter.
	r.metadata.RegisterFamily(font.Family{
		Name: defaultFontFamily,
		Faces: []font.Face{
			{Weight: font.Regular, Style: font.Normal, Data: fonts.InterRegular},
			{Weight: font.Bold, Style: font.Normal, Data: fonts.InterBold},
		},
	})

	return r
}

// Register adds a font face to the registry.
//
// The data parameter must contain valid TTF or OTF font data. It is stored
// by the underlying [font.Registry] which makes a defensive copy.
//
// Returns an error if the font data cannot be parsed into a [text.FontSource].
func (r *FontRegistry) Register(family string, weight font.Weight, style font.Style, data []byte) error {
	if len(data) == 0 {
		return fmt.Errorf("fontregistry: empty font data for %s/%s/%s", family, weight, style)
	}

	// Validate that the font data can be parsed before storing.
	src, err := text.NewFontSource(data)
	if err != nil {
		return fmt.Errorf("fontregistry: invalid font data for %s/%s/%s: %w", family, weight, style, err)
	}

	r.mu.Lock()
	defer r.mu.Unlock()

	r.metadata.RegisterFamily(font.Family{
		Name: family,
		Faces: []font.Face{
			{Weight: weight, Style: style, Data: data},
		},
	})

	// Cache the pre-built FontSource for this exact key.
	key := sourceKey{family: family, weight: weight, style: style}
	r.sources[key] = src

	return nil
}

// Resolve finds the best matching [text.FontSource] for the given family,
// weight, and style. Returns nil if no font can be resolved (should not
// happen in practice because Inter is always registered).
//
// Resolution uses CSS font-matching via [font.Registry.Resolve], then
// creates and caches a [text.FontSource] from the resolved data.
func (r *FontRegistry) Resolve(family string, weight font.Weight, style font.Style) *text.FontSource {
	r.mu.RLock()
	key := sourceKey{family: family, weight: weight, style: style}
	if src, ok := r.sources[key]; ok {
		r.mu.RUnlock()
		return src
	}
	r.mu.RUnlock()

	// Resolve font data via CSS weight matching.
	data, ok := r.metadata.Resolve(family, weight, style)
	if !ok {
		// Fall back to default family (Inter).
		data, ok = r.metadata.Resolve(defaultFontFamily, weight, style)
		if !ok {
			return nil
		}
	}

	// Create FontSource from resolved data.
	src, err := text.NewFontSource(data)
	if err != nil {
		return nil
	}

	// Cache for future lookups.
	r.mu.Lock()
	// Double-check: another goroutine might have created it.
	if existing, ok := r.sources[key]; ok {
		r.mu.Unlock()
		return existing
	}
	r.sources[key] = src
	r.mu.Unlock()

	return src
}

// HasFamily reports whether the given family name is registered.
func (r *FontRegistry) HasFamily(family string) bool {
	return r.metadata.HasFamily(family)
}

// FamilyNames returns a sorted list of all registered family names.
func (r *FontRegistry) FamilyNames() []string {
	return r.metadata.FamilyNames()
}

// globalFontRegistry is the process-wide font registry singleton.
// Initialized lazily via [GlobalFontRegistry].
var (
	globalRegistryOnce sync.Once
	globalRegistry     *FontRegistry
)

// GlobalFontRegistry returns the process-wide [FontRegistry] singleton.
//
// The global registry is created lazily on first call with embedded Inter
// pre-registered. Plugin font loading and Canvas rendering both use this
// singleton to share font data.
func GlobalFontRegistry() *FontRegistry {
	globalRegistryOnce.Do(func() {
		globalRegistry = NewFontRegistry()
	})
	return globalRegistry
}
