package layout

import (
	"testing"

	"github.com/gogpu/ui/geometry"
)

// mockLayoutable is a simple layoutable for testing.
type mockLayoutable struct {
	id            uint64
	preferredSize geometry.Size
	layoutCalls   int
	children      []Layoutable
}

func (m *mockLayoutable) ID() uint64 {
	return m.id
}

func (m *mockLayoutable) Layout(constraints geometry.Constraints) geometry.Size {
	m.layoutCalls++
	return constraints.Constrain(m.preferredSize)
}

func (m *mockLayoutable) Children() []Layoutable {
	return m.children
}

func TestNewEngine(t *testing.T) {
	engine := NewEngine()

	if engine == nil {
		t.Fatal("NewEngine returned nil")
	}

	if engine.IsCacheEnabled() {
		t.Error("cache should be disabled by default")
	}

	if engine.CacheSize() != 0 {
		t.Errorf("cache size = %d, want 0", engine.CacheSize())
	}

	if engine.DirtyCount() != 0 {
		t.Errorf("dirty count = %d, want 0", engine.DirtyCount())
	}
}

func TestEngine_Layout(t *testing.T) {
	engine := NewEngine()

	mock := &mockLayoutable{
		id:            1,
		preferredSize: geometry.Sz(100, 50),
	}

	constraints := geometry.Loose(geometry.Sz(200, 200))
	size := engine.Layout(mock, constraints)

	if size != mock.preferredSize {
		t.Errorf("Layout() = %v, want %v", size, mock.preferredSize)
	}

	if mock.layoutCalls != 1 {
		t.Errorf("layoutCalls = %d, want 1", mock.layoutCalls)
	}
}

func TestEngine_LayoutNil(t *testing.T) {
	engine := NewEngine()

	size := engine.Layout(nil, geometry.Expand())

	if !size.IsZero() {
		t.Errorf("Layout(nil) = %v, want zero size", size)
	}
}

func TestEngine_CacheEnabled(t *testing.T) {
	engine := NewEngine()
	engine.EnableCache(true)

	if !engine.IsCacheEnabled() {
		t.Error("cache should be enabled")
	}

	mock := &mockLayoutable{
		id:            1,
		preferredSize: geometry.Sz(100, 50),
	}

	constraints := geometry.Loose(geometry.Sz(200, 200))

	// First layout
	size1 := engine.Layout(mock, constraints)
	if mock.layoutCalls != 1 {
		t.Errorf("first layout: layoutCalls = %d, want 1", mock.layoutCalls)
	}

	// Second layout with same constraints should use cache
	size2 := engine.Layout(mock, constraints)
	if mock.layoutCalls != 1 {
		t.Errorf("second layout: layoutCalls = %d, want 1 (cached)", mock.layoutCalls)
	}

	if size1 != size2 {
		t.Errorf("cached size mismatch: %v != %v", size1, size2)
	}

	// Check stats
	stats := engine.Stats()
	if stats.CacheHits != 1 {
		t.Errorf("cache hits = %d, want 1", stats.CacheHits)
	}
	if stats.CacheMisses != 1 {
		t.Errorf("cache misses = %d, want 1", stats.CacheMisses)
	}
}

func TestEngine_CacheDifferentConstraints(t *testing.T) {
	engine := NewEngine()
	engine.EnableCache(true)

	mock := &mockLayoutable{
		id:            1,
		preferredSize: geometry.Sz(100, 50),
	}

	// Layout with different constraints
	_ = engine.Layout(mock, geometry.Loose(geometry.Sz(200, 200)))
	_ = engine.Layout(mock, geometry.Loose(geometry.Sz(300, 300)))

	if mock.layoutCalls != 2 {
		t.Errorf("layoutCalls = %d, want 2 (different constraints)", mock.layoutCalls)
	}
}

func TestEngine_DirtyTracking(t *testing.T) {
	engine := NewEngine()
	engine.EnableCache(true)

	mock := &mockLayoutable{
		id:            1,
		preferredSize: geometry.Sz(100, 50),
	}

	constraints := geometry.Loose(geometry.Sz(200, 200))

	// First layout
	_ = engine.Layout(mock, constraints)
	if mock.layoutCalls != 1 {
		t.Errorf("first layout: layoutCalls = %d, want 1", mock.layoutCalls)
	}

	// Mark dirty
	engine.MarkDirty(mock.id)

	if !engine.IsDirty(mock.id) {
		t.Error("element should be marked dirty")
	}

	// Layout again - should re-layout because dirty
	_ = engine.Layout(mock, constraints)
	if mock.layoutCalls != 2 {
		t.Errorf("after dirty: layoutCalls = %d, want 2", mock.layoutCalls)
	}

	// Should no longer be dirty
	if engine.IsDirty(mock.id) {
		t.Error("element should not be dirty after layout")
	}
}

func TestEngine_ClearCache(t *testing.T) {
	engine := NewEngine()
	engine.EnableCache(true)

	mock := &mockLayoutable{
		id:            1,
		preferredSize: geometry.Sz(100, 50),
	}

	constraints := geometry.Loose(geometry.Sz(200, 200))

	// Fill cache
	_ = engine.Layout(mock, constraints)

	if engine.CacheSize() != 1 {
		t.Errorf("cache size = %d, want 1", engine.CacheSize())
	}

	// Clear cache
	engine.ClearCache()

	if engine.CacheSize() != 0 {
		t.Errorf("after clear: cache size = %d, want 0", engine.CacheSize())
	}
}

func TestEngine_DisableCache(t *testing.T) {
	engine := NewEngine()
	engine.EnableCache(true)

	mock := &mockLayoutable{
		id:            1,
		preferredSize: geometry.Sz(100, 50),
	}

	constraints := geometry.Loose(geometry.Sz(200, 200))

	// Fill cache
	_ = engine.Layout(mock, constraints)

	// Disable cache
	engine.EnableCache(false)

	if engine.IsCacheEnabled() {
		t.Error("cache should be disabled")
	}

	if engine.CacheSize() != 0 {
		t.Errorf("cache should be cleared when disabled")
	}
}

func TestEngine_MarkDirtyWithAncestors(t *testing.T) {
	engine := NewEngine()

	engine.MarkDirtyWithAncestors(1, []uint64{2, 3, 4})

	if !engine.IsDirty(1) {
		t.Error("element 1 should be dirty")
	}
	if !engine.IsDirty(2) {
		t.Error("ancestor 2 should be dirty")
	}
	if !engine.IsDirty(3) {
		t.Error("ancestor 3 should be dirty")
	}
	if !engine.IsDirty(4) {
		t.Error("ancestor 4 should be dirty")
	}
	if engine.IsDirty(5) {
		t.Error("element 5 should not be dirty")
	}
}

func TestEngine_ClearDirty(t *testing.T) {
	engine := NewEngine()

	engine.MarkDirty(1)
	engine.MarkDirty(2)

	if engine.DirtyCount() != 2 {
		t.Errorf("dirty count = %d, want 2", engine.DirtyCount())
	}

	engine.ClearDirty(1)

	if engine.IsDirty(1) {
		t.Error("element 1 should not be dirty after clear")
	}
	if !engine.IsDirty(2) {
		t.Error("element 2 should still be dirty")
	}

	engine.ClearAllDirty()

	if engine.DirtyCount() != 0 {
		t.Errorf("after ClearAllDirty: dirty count = %d, want 0", engine.DirtyCount())
	}
}

func TestEngine_ResetStats(t *testing.T) {
	engine := NewEngine()

	mock := &mockLayoutable{
		id:            1,
		preferredSize: geometry.Sz(100, 50),
	}

	_ = engine.Layout(mock, geometry.Expand())

	stats := engine.Stats()
	if stats.LayoutCalls == 0 {
		t.Error("LayoutCalls should be non-zero")
	}

	engine.ResetStats()

	stats = engine.Stats()
	if stats.LayoutCalls != 0 {
		t.Errorf("after reset: LayoutCalls = %d, want 0", stats.LayoutCalls)
	}
}

func TestEngine_LayoutWithIntrinsics(t *testing.T) {
	engine := NewEngine()

	mock := &mockLayoutable{
		id:            1,
		preferredSize: geometry.Sz(100, 50),
	}

	constraints := geometry.Loose(geometry.Sz(200, 200))
	size := engine.LayoutWithIntrinsics(mock, constraints)

	// Should call layout twice: once loose, once tight
	if mock.layoutCalls != 2 {
		t.Errorf("layoutCalls = %d, want 2", mock.layoutCalls)
	}

	if size != mock.preferredSize {
		t.Errorf("size = %v, want %v", size, mock.preferredSize)
	}
}

func TestEngine_LayoutWithIntrinsics_TightConstraints(t *testing.T) {
	engine := NewEngine()

	mock := &mockLayoutable{
		id:            1,
		preferredSize: geometry.Sz(100, 50),
	}

	// With tight constraints, intrinsic measurement is skipped
	constraints := geometry.Tight(geometry.Sz(80, 40))
	size := engine.LayoutWithIntrinsics(mock, constraints)

	// Should call layout twice: first loose measurement, then tight layout
	if mock.layoutCalls != 2 {
		t.Errorf("layoutCalls = %d, want 2", mock.layoutCalls)
	}

	// Size should be constrained to 80x40
	expected := geometry.Sz(80, 40)
	if size != expected {
		t.Errorf("size = %v, want %v", size, expected)
	}
}

func TestEngine_LayoutTree(t *testing.T) {
	engine := NewEngine()

	child := &mockLayoutable{
		id:            2,
		preferredSize: geometry.Sz(50, 30),
	}

	parent := &mockLayoutable{
		id:            1,
		preferredSize: geometry.Sz(100, 50),
		children:      []Layoutable{child},
	}

	_ = engine.LayoutTree(parent, geometry.Expand())

	// Only parent's Layout is called (children are responsibility of parent)
	if parent.layoutCalls != 1 {
		t.Errorf("parent layoutCalls = %d, want 1", parent.layoutCalls)
	}
}

func TestEngine_ClearCacheFor(t *testing.T) {
	engine := NewEngine()
	engine.EnableCache(true)

	mock1 := &mockLayoutable{id: 1, preferredSize: geometry.Sz(100, 50)}
	mock2 := &mockLayoutable{id: 2, preferredSize: geometry.Sz(200, 100)}

	_ = engine.Layout(mock1, geometry.Expand())
	_ = engine.Layout(mock2, geometry.Expand())

	if engine.CacheSize() != 2 {
		t.Errorf("cache size = %d, want 2", engine.CacheSize())
	}

	engine.ClearCacheFor(1)

	if engine.CacheSize() != 1 {
		t.Errorf("after clear for 1: cache size = %d, want 1", engine.CacheSize())
	}

	// Layout mock1 again - should not use cache
	_ = engine.Layout(mock1, geometry.Expand())
	if mock1.layoutCalls != 2 {
		t.Errorf("mock1 layoutCalls = %d, want 2", mock1.layoutCalls)
	}
}

func TestEngine_ZeroIDNoCache(t *testing.T) {
	engine := NewEngine()
	engine.EnableCache(true)

	// Element with zero ID should not be cached
	mock := &mockLayoutable{
		id:            0,
		preferredSize: geometry.Sz(100, 50),
	}

	_ = engine.Layout(mock, geometry.Expand())
	_ = engine.Layout(mock, geometry.Expand())

	// Should layout twice since ID=0 bypasses cache
	if mock.layoutCalls != 2 {
		t.Errorf("layoutCalls = %d, want 2 (ID=0 not cached)", mock.layoutCalls)
	}

	if engine.CacheSize() != 0 {
		t.Errorf("cache size = %d, want 0 (ID=0 not cached)", engine.CacheSize())
	}
}
