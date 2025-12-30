package layout

import (
	"github.com/gogpu/ui/geometry"
)

// Layoutable represents an element that can be laid out.
//
// This interface abstracts over widgets and layout containers,
// allowing the engine to work with any layoutable element.
type Layoutable interface {
	// Layout calculates size given constraints and returns the computed size.
	// The implementation should also position any children.
	Layout(constraints geometry.Constraints) geometry.Size

	// Children returns child layoutables for traversal.
	// Returns nil for leaf elements.
	Children() []Layoutable

	// ID returns a unique identifier for caching purposes.
	// Return 0 if caching is not needed for this element.
	ID() uint64
}

// LayoutResult stores the computed layout for an element.
type LayoutResult struct {
	// Size is the computed size after layout.
	Size geometry.Size

	// Position is the offset from parent origin.
	Position geometry.Point

	// Constraints used for this layout pass.
	Constraints geometry.Constraints
}

// Engine manages layout passes with optional caching and dirty tracking.
//
// The engine supports:
//   - Single-pass layout for most cases
//   - Multi-pass layout for intrinsic sizing
//   - Caching to avoid redundant layout calculations
//   - Dirty tracking for incremental updates
//
// Engine is NOT thread-safe.
type Engine struct {
	// Cache stores layout results by element ID and constraints hash.
	cache map[cacheKey]LayoutResult

	// dirtySet tracks elements that need re-layout.
	dirtySet map[uint64]struct{}

	// stats tracks layout statistics for debugging.
	stats LayoutStats

	// enableCache controls whether caching is active.
	enableCache bool
}

// cacheKey combines element ID and constraints for cache lookup.
type cacheKey struct {
	id          uint64
	constraints geometry.Constraints
}

// LayoutStats provides layout performance metrics.
type LayoutStats struct {
	// LayoutCalls is the number of Layout calls made.
	LayoutCalls int

	// CacheHits is the number of times a cached result was used.
	CacheHits int

	// CacheMisses is the number of times layout was computed.
	CacheMisses int
}

// NewEngine creates a new layout engine.
//
// By default, caching is disabled. Use EnableCache to enable it.
func NewEngine() *Engine {
	return &Engine{
		cache:       make(map[cacheKey]LayoutResult),
		dirtySet:    make(map[uint64]struct{}),
		enableCache: false,
	}
}

// EnableCache enables or disables layout caching.
//
// When enabled, the engine caches layout results and reuses them
// when the same element is laid out with the same constraints.
func (e *Engine) EnableCache(enable bool) {
	e.enableCache = enable
	if !enable {
		e.ClearCache()
	}
}

// IsCacheEnabled returns whether caching is enabled.
func (e *Engine) IsCacheEnabled() bool {
	return e.enableCache
}

// Layout performs a layout pass on the given element.
//
// This is the main entry point for layout. It delegates to the element's
// Layout method, optionally using cached results.
func (e *Engine) Layout(element Layoutable, constraints geometry.Constraints) geometry.Size {
	e.stats.LayoutCalls++

	if element == nil {
		return geometry.Size{}
	}

	// Normalize constraints to ensure validity
	constraints = constraints.Normalize()

	// Try cache lookup
	if cachedSize, ok := e.tryCache(element, constraints); ok {
		return cachedSize
	}

	// Perform actual layout
	size := element.Layout(constraints)

	// Store result in cache
	e.storeInCache(element, constraints, size)

	return size
}

// tryCache attempts to retrieve a cached layout result.
func (e *Engine) tryCache(element Layoutable, constraints geometry.Constraints) (geometry.Size, bool) {
	if !e.enableCache {
		return geometry.Size{}, false
	}

	id := element.ID()
	if id == 0 {
		return geometry.Size{}, false
	}

	key := cacheKey{id: id, constraints: constraints}
	result, ok := e.cache[key]
	if !ok {
		e.stats.CacheMisses++
		return geometry.Size{}, false
	}

	// Check if element is dirty
	if _, dirty := e.dirtySet[id]; dirty {
		e.stats.CacheMisses++
		return geometry.Size{}, false
	}

	e.stats.CacheHits++
	return result.Size, true
}

// storeInCache stores a layout result in the cache.
func (e *Engine) storeInCache(element Layoutable, constraints geometry.Constraints, size geometry.Size) {
	if !e.enableCache {
		return
	}

	id := element.ID()
	if id == 0 {
		return
	}

	key := cacheKey{id: id, constraints: constraints}
	e.cache[key] = LayoutResult{
		Size:        size,
		Constraints: constraints,
	}
	delete(e.dirtySet, id)
}

// MarkDirty marks an element as needing re-layout.
//
// This invalidates the cache for the element and all ancestors.
// Call this when an element's content or properties change.
func (e *Engine) MarkDirty(id uint64) {
	if id != 0 {
		e.dirtySet[id] = struct{}{}
	}
}

// MarkDirtyWithAncestors marks an element and all provided ancestor IDs as dirty.
func (e *Engine) MarkDirtyWithAncestors(id uint64, ancestorIDs []uint64) {
	e.MarkDirty(id)
	for _, ancestorID := range ancestorIDs {
		e.MarkDirty(ancestorID)
	}
}

// IsDirty returns whether an element needs re-layout.
func (e *Engine) IsDirty(id uint64) bool {
	_, dirty := e.dirtySet[id]
	return dirty
}

// ClearDirty removes the dirty flag for an element.
func (e *Engine) ClearDirty(id uint64) {
	delete(e.dirtySet, id)
}

// ClearAllDirty removes all dirty flags.
func (e *Engine) ClearAllDirty() {
	e.dirtySet = make(map[uint64]struct{})
}

// ClearCache removes all cached layout results.
//
// Call this when the layout tree structure changes significantly.
func (e *Engine) ClearCache() {
	e.cache = make(map[cacheKey]LayoutResult)
}

// ClearCacheFor removes cached results for a specific element.
func (e *Engine) ClearCacheFor(id uint64) {
	// Remove all cache entries for this ID (any constraints)
	for key := range e.cache {
		if key.id == id {
			delete(e.cache, key)
		}
	}
}

// Stats returns current layout statistics.
func (e *Engine) Stats() LayoutStats {
	return e.stats
}

// ResetStats resets layout statistics to zero.
func (e *Engine) ResetStats() {
	e.stats = LayoutStats{}
}

// CacheSize returns the number of cached layout results.
func (e *Engine) CacheSize() int {
	return len(e.cache)
}

// DirtyCount returns the number of dirty elements.
func (e *Engine) DirtyCount() int {
	return len(e.dirtySet)
}

// LayoutWithIntrinsics performs layout with intrinsic size calculation.
//
// This is a two-pass layout:
//  1. First pass: measure intrinsic sizes with loose constraints
//  2. Second pass: layout with actual constraints using intrinsic info
//
// Use this for elements that need to know their content size before layout.
func (e *Engine) LayoutWithIntrinsics(element Layoutable, constraints geometry.Constraints) geometry.Size {
	if element == nil {
		return geometry.Size{}
	}

	// First pass: measure with loose constraints
	looseConstraints := constraints.Loosen()
	intrinsicSize := element.Layout(looseConstraints)

	// Second pass: layout with actual constraints, using intrinsic size as hint
	// For tight constraints, use them directly
	if constraints.IsTight() {
		return element.Layout(constraints)
	}

	// For loose constraints, tighten to intrinsic size (clamped to constraints)
	tightened := constraints.Tighten(intrinsicSize)
	return element.Layout(tightened)
}

// LayoutTree performs layout on an entire tree of elements.
//
// This recursively lays out all children before the parent, ensuring
// that child sizes are known when positioning them.
func (e *Engine) LayoutTree(root Layoutable, constraints geometry.Constraints) geometry.Size {
	if root == nil {
		return geometry.Size{}
	}

	// The root element's Layout is responsible for recursively
	// laying out its children. We just call Layout on root.
	return e.Layout(root, constraints)
}
