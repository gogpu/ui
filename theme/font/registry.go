package font

import (
	"sort"
	"sync"
)

// Registry manages font families and resolves font face requests to font data.
//
// The registry maps family+weight+style combinations to raw font data. When an
// exact match is not available, it uses the CSS font-matching algorithm to find
// the closest available weight within the requested style.
//
// Registry is safe for concurrent use. It uses a read-write mutex optimized
// for read-heavy workloads (font resolution is far more frequent than
// registration).
type Registry struct {
	mu       sync.RWMutex
	families map[string]*familyEntry
}

// familyEntry holds all registered faces for a single font family.
type familyEntry struct {
	name  string
	faces []Face
}

// NewRegistry creates an empty font registry.
func NewRegistry() *Registry {
	return &Registry{
		families: make(map[string]*familyEntry),
	}
}

// RegisterFamily registers all faces of a font family.
//
// If the family already exists, new faces are appended. If a face with the
// same weight and style already exists, it is replaced.
func (r *Registry) RegisterFamily(family Family) {
	r.mu.Lock()
	defer r.mu.Unlock()

	entry, ok := r.families[family.Name]
	if !ok {
		entry = &familyEntry{name: family.Name}
		r.families[family.Name] = entry
	}

	for _, face := range family.Faces {
		r.addFaceLocked(entry, face)
	}
}

// addFaceLocked adds or replaces a face in the family entry.
// Caller must hold r.mu write lock.
func (r *Registry) addFaceLocked(entry *familyEntry, face Face) {
	for i, existing := range entry.faces {
		if existing.Weight == face.Weight && existing.Style == face.Style {
			entry.faces[i] = face
			return
		}
	}
	entry.faces = append(entry.faces, face)
}

// Resolve finds the best matching font data for the given family, weight, and style.
//
// Resolution follows the CSS font-matching algorithm
// (https://www.w3.org/TR/css-fonts-4/#font-style-matching):
//
//  1. Look for an exact weight+style match.
//  2. If the requested style is Italic and no italic face exists, fall back to Normal style.
//  3. Apply CSS weight resolution: for weights <= 500, try lighter first then heavier;
//     for weights > 500, try heavier first then lighter.
//
// Returns the font data and true if a match was found, or nil and false if the
// family does not exist or has no faces.
func (r *Registry) Resolve(familyName string, weight Weight, style Style) ([]byte, bool) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	entry, ok := r.families[familyName]
	if !ok || len(entry.faces) == 0 {
		return nil, false
	}

	// Try exact match first.
	if data := findExact(entry.faces, weight, style); data != nil {
		return data, true
	}

	// If italic requested but not available, fall back to normal style.
	searchStyle := style
	if style == Italic {
		if !hasStyle(entry.faces, Italic) {
			searchStyle = Normal
		}
	}

	// Try exact weight with resolved style.
	if searchStyle != style {
		if data := findExact(entry.faces, weight, searchStyle); data != nil {
			return data, true
		}
	}

	// CSS weight resolution with the resolved style.
	if data := resolveWeight(entry.faces, weight, searchStyle); data != nil {
		return data, true
	}

	// Last resort: any face in the family.
	return entry.faces[0].Data, true
}

// FamilyNames returns a sorted list of all registered family names.
func (r *Registry) FamilyNames() []string {
	r.mu.RLock()
	defer r.mu.RUnlock()

	names := make([]string, 0, len(r.families))
	for name := range r.families {
		names = append(names, name)
	}
	sort.Strings(names)
	return names
}

// HasFamily reports whether the given family name is registered.
func (r *Registry) HasFamily(familyName string) bool {
	r.mu.RLock()
	defer r.mu.RUnlock()
	_, ok := r.families[familyName]
	return ok
}

// FaceCount returns the number of registered faces for the given family,
// or 0 if the family is not registered.
func (r *Registry) FaceCount(familyName string) int {
	r.mu.RLock()
	defer r.mu.RUnlock()

	entry, ok := r.families[familyName]
	if !ok {
		return 0
	}
	return len(entry.faces)
}

// findExact returns font data for an exact weight+style match, or nil.
func findExact(faces []Face, weight Weight, style Style) []byte {
	for _, f := range faces {
		if f.Weight == weight && f.Style == style {
			return f.Data
		}
	}
	return nil
}

// hasStyle reports whether any face has the given style.
func hasStyle(faces []Face, style Style) bool {
	for _, f := range faces {
		if f.Style == style {
			return true
		}
	}
	return false
}

// resolveWeight implements the CSS font-weight matching algorithm.
//
// For requested weight <= 500: try lighter weights first (descending), then heavier (ascending).
// For requested weight > 500: try heavier weights first (ascending), then lighter (descending).
//
// Special cases per CSS spec:
//   - Weight 400: try 500 first, then descend from 400, then ascend from 500.
//   - Weight 500: try 400 first, then descend from 400, then ascend from 500.
func resolveWeight(faces []Face, target Weight, style Style) []byte {
	// Collect available weights for the target style, sorted ascending.
	var candidates []candidate
	for _, f := range faces {
		if f.Style == style {
			candidates = append(candidates, candidate{weight: f.Weight, data: f.Data})
		}
	}
	if len(candidates) == 0 {
		return nil
	}
	sort.Slice(candidates, func(i, j int) bool {
		return candidates[i].weight < candidates[j].weight
	})

	// Special handling for 400 and 500 per CSS spec.
	if target == Regular || target == Medium {
		return resolveWeight400or500(candidates, target)
	}

	if target <= Medium {
		// Try descending from target, then ascending.
		return resolveWeightLightward(candidates, target)
	}
	// target > 500: try ascending from target, then descending.
	return resolveWeightBoldward(candidates, target)
}

// resolveWeight400or500 handles the CSS special case for weights 400 and 500.
// For 400: try 500, then descend from 400, then ascend from 500.
// For 500: try 400, then descend from 400, then ascend from 500.
func resolveWeight400or500(candidates []candidate, target Weight) []byte {
	var alternate Weight
	if target == Regular {
		alternate = Medium
	} else {
		alternate = Regular
	}

	// Try the alternate weight (400<->500).
	for _, c := range candidates {
		if c.weight == alternate {
			return c.data
		}
	}

	// Descend from target (lighter).
	for i := len(candidates) - 1; i >= 0; i-- {
		if candidates[i].weight < target {
			return candidates[i].data
		}
	}

	// Ascend from 500 (heavier).
	for _, c := range candidates {
		if c.weight > Medium {
			return c.data
		}
	}

	// Shouldn't reach here if candidates is non-empty, but return first available.
	return candidates[0].data
}

// candidate is a weight+data pair used during resolution.
type candidate struct {
	weight Weight
	data   []byte
}

// resolveWeightLightward tries lighter weights first, then heavier.
// Used when target <= 400 (excluding 400/500 special case).
func resolveWeightLightward(candidates []candidate, target Weight) []byte {
	// Nearest lighter (descending from target).
	var bestLighter []byte
	bestLighterDist := Weight(1000)
	for _, c := range candidates {
		if c.weight <= target {
			dist := target - c.weight
			if dist < bestLighterDist {
				bestLighterDist = dist
				bestLighter = c.data
			}
		}
	}
	if bestLighter != nil {
		return bestLighter
	}

	// Nearest heavier (ascending from target).
	var bestHeavier []byte
	bestHeavierDist := Weight(1000)
	for _, c := range candidates {
		if c.weight > target {
			dist := c.weight - target
			if dist < bestHeavierDist {
				bestHeavierDist = dist
				bestHeavier = c.data
			}
		}
	}
	return bestHeavier
}

// resolveWeightBoldward tries heavier weights first, then lighter.
// Used when target > 500.
func resolveWeightBoldward(candidates []candidate, target Weight) []byte {
	// Nearest heavier (ascending from target).
	var bestHeavier []byte
	bestHeavierDist := Weight(1000)
	for _, c := range candidates {
		if c.weight >= target {
			dist := c.weight - target
			if dist < bestHeavierDist {
				bestHeavierDist = dist
				bestHeavier = c.data
			}
		}
	}
	if bestHeavier != nil {
		return bestHeavier
	}

	// Nearest lighter (descending from target).
	var bestLighter []byte
	bestLighterDist := Weight(1000)
	for _, c := range candidates {
		if c.weight < target {
			dist := target - c.weight
			if dist < bestLighterDist {
				bestLighterDist = dist
				bestLighter = c.data
			}
		}
	}
	return bestLighter
}
