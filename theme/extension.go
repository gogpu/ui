package theme

import "sync"

// ThemeExtension allows custom theme properties for third-party components.
//
// ThemeExtension provides a mechanism for third-party developers and
// enterprise applications to add custom properties to themes without
// modifying the core Theme struct. This follows Flutter's proven pattern
// for extensible theming.
//
// # Implementing a ThemeExtension
//
// To create a custom extension, implement all four methods:
//
//	type CorporateExtension struct {
//	    BrandPrimary   widget.Color
//	    BrandSecondary widget.Color
//	    LogoURL        string
//	}
//
//	func (e *CorporateExtension) Name() string { return "corporate" }
//
//	func (e *CorporateExtension) Merge(other ThemeExtension) ThemeExtension {
//	    if o, ok := other.(*CorporateExtension); ok {
//	        return o // Use other's values (override)
//	    }
//	    return e // Keep original if types don't match
//	}
//
//	func (e *CorporateExtension) Lerp(other ThemeExtension, t float32) ThemeExtension {
//	    if o, ok := other.(*CorporateExtension); ok {
//	        return &CorporateExtension{
//	            BrandPrimary:   e.BrandPrimary.Lerp(o.BrandPrimary, t),
//	            BrandSecondary: e.BrandSecondary.Lerp(o.BrandSecondary, t),
//	            LogoURL:        e.LogoURL, // Can't interpolate strings
//	        }
//	    }
//	    return e
//	}
//
//	func (e *CorporateExtension) CopyWith(overrides map[string]any) ThemeExtension {
//	    copy := *e
//	    if v, ok := overrides["brand_primary"].(widget.Color); ok {
//	        copy.BrandPrimary = v
//	    }
//	    return &copy
//	}
//
// # Registering Extensions
//
//	theme := DefaultLight()
//	theme.RegisterExtension(&CorporateExtension{
//	    BrandPrimary: widget.Hex(0x00529B),
//	})
//
// # Retrieving Extensions
//
//	// Type-safe retrieval
//	if ext, ok := ExtensionAs[*CorporateExtension](theme, "corporate"); ok {
//	    canvas.DrawRect(bounds, RectStyle{Fill: ext.BrandPrimary})
//	}
type ThemeExtension interface {
	// Name returns the unique identifier for this extension.
	//
	// The name should be a stable string that uniquely identifies the
	// extension type. Using a package path or namespace is recommended
	// to avoid collisions (e.g., "myorg/branding", "corporate").
	Name() string

	// Merge combines this extension with another, typically for theme inheritance.
	//
	// When a child theme has an extension with the same name as a parent theme,
	// Merge is called to combine them. The other parameter is the child's
	// extension that should override values from this extension.
	//
	// If other is not the same type, implementations should return the
	// original extension unchanged.
	//
	// Example:
	//
	//	func (e *MyExt) Merge(other ThemeExtension) ThemeExtension {
	//	    if o, ok := other.(*MyExt); ok {
	//	        return o // Child completely overrides parent
	//	    }
	//	    return e
	//	}
	Merge(other ThemeExtension) ThemeExtension

	// Lerp interpolates between this extension and another for animations.
	//
	// t ranges from 0.0 (this) to 1.0 (other). This enables smooth theme
	// transitions when animating between light and dark themes, or between
	// different brand themes.
	//
	// For properties that cannot be interpolated (like strings), return
	// the value from this extension when t < 0.5 and from other when t >= 0.5.
	//
	// If other is not the same type, implementations should return the
	// original extension unchanged.
	//
	// Example:
	//
	//	func (e *MyExt) Lerp(other ThemeExtension, t float32) ThemeExtension {
	//	    if o, ok := other.(*MyExt); ok {
	//	        return &MyExt{
	//	            Color: e.Color.Lerp(o.Color, t),
	//	            Label: lerpString(e.Label, o.Label, t),
	//	        }
	//	    }
	//	    return e
	//	}
	Lerp(other ThemeExtension, t float32) ThemeExtension

	// CopyWith creates a modified copy of the extension.
	//
	// This allows creating variations of an extension without mutating
	// the original. The overrides map uses string keys that are specific
	// to each extension type.
	//
	// Example:
	//
	//	func (e *MyExt) CopyWith(overrides map[string]any) ThemeExtension {
	//	    copy := *e
	//	    if v, ok := overrides["color"].(widget.Color); ok {
	//	        copy.Color = v
	//	    }
	//	    return &copy
	//	}
	CopyWith(overrides map[string]any) ThemeExtension
}

// typedExtensions holds ThemeExtension implementations separately from
// the generic Extensions map for type-safe access with full interface support.
//
// This allows extensions registered via RegisterExtension to support
// Merge, Lerp, and CopyWith operations, while maintaining backward
// compatibility with the simple Extensions map[string]any.
type typedExtensions struct {
	mu         sync.RWMutex
	extensions map[string]ThemeExtension
}

// newTypedExtensions creates a new typedExtensions instance.
func newTypedExtensions() *typedExtensions {
	return &typedExtensions{
		extensions: make(map[string]ThemeExtension),
	}
}

// register adds or replaces an extension.
func (te *typedExtensions) register(ext ThemeExtension) {
	te.mu.Lock()
	defer te.mu.Unlock()
	te.extensions[ext.Name()] = ext
}

// get retrieves an extension by name.
func (te *typedExtensions) get(name string) ThemeExtension {
	te.mu.RLock()
	defer te.mu.RUnlock()
	return te.extensions[name]
}

// clone creates a shallow copy of the extensions map.
func (te *typedExtensions) clone() *typedExtensions {
	te.mu.RLock()
	defer te.mu.RUnlock()

	cloned := newTypedExtensions()
	for k, v := range te.extensions {
		cloned.extensions[k] = v
	}
	return cloned
}

// all returns a copy of all extensions.
func (te *typedExtensions) all() map[string]ThemeExtension {
	te.mu.RLock()
	defer te.mu.RUnlock()

	result := make(map[string]ThemeExtension, len(te.extensions))
	for k, v := range te.extensions {
		result[k] = v
	}
	return result
}

// ExtensionAs retrieves and type-asserts a ThemeExtension.
//
// This is a generic helper that provides type-safe access to extensions.
// It returns the typed extension and true if found and of the correct type,
// or the zero value and false otherwise.
//
// Example:
//
//	if ext, ok := theme.ExtensionAs[*CorporateExtension](theme, "corporate"); ok {
//	    canvas.DrawRect(bounds, RectStyle{Fill: ext.BrandPrimary})
//	}
func ExtensionAs[T ThemeExtension](t *Theme, name string) (T, bool) {
	var zero T
	if t == nil {
		return zero, false
	}

	ext := t.TypedExtension(name)
	if ext == nil {
		return zero, false
	}

	typed, ok := ext.(T)
	return typed, ok
}

// LerpString interpolates between two strings for animation.
//
// Since strings cannot be mathematically interpolated, this function
// returns a when t < 0.5 and b when t >= 0.5.
//
// This is a helper function for implementing ThemeExtension.Lerp().
func LerpString(a, b string, t float32) string {
	if t < 0.5 {
		return a
	}
	return b
}

// LerpFloat32 interpolates between two float32 values.
//
// This is a helper function for implementing ThemeExtension.Lerp().
func LerpFloat32(a, b, t float32) float32 {
	return a + (b-a)*t
}

// LerpInt interpolates between two int values.
//
// The result is rounded to the nearest integer.
// This is a helper function for implementing ThemeExtension.Lerp().
func LerpInt(a, b int, t float32) int {
	return a + int(float32(b-a)*t+0.5)
}
