package primitives

import (
	"sync/atomic"

	"github.com/gogpu/gg/scene"
	"github.com/gogpu/ui/a11y"
	"github.com/gogpu/ui/event"
	"github.com/gogpu/ui/geometry"
	internalRender "github.com/gogpu/ui/internal/render"
	"github.com/gogpu/ui/widget"
)

// nextCacheKey is a monotonic counter for generating unique cache keys.
// Each RepaintBoundary gets a unique uint64 ID at creation time, used as
// the key into the dirty boundary tracking set. Atomic to be safe for
// concurrent RepaintBoundary creation across goroutines.
var nextCacheKey atomic.Uint64

// RepaintBoundary is a display widget that caches its child subtree as a
// display list (SceneCache). When the child subtree is clean (no dirty
// widgets), the cached display list is replayed into the parent canvas
// instead of re-executing Draw on every descendant.
//
// This is the Flutter RepaintBoundary / PictureLayer pattern: an explicit
// opt-in boundary that isolates expensive subtrees from the rest of the
// render tree. Display lists are portable across all gg backends (Vulkan,
// DX12, Metal, GLES, Software, future WASM/WebGPU).
//
// Cache lifecycle:
//   - The cached scene is allocated on first draw (lazy).
//   - The cache is invalidated when any descendant calls SetNeedsRedraw,
//     which propagates UP to the boundary (ADR-007 upward propagation),
//     or when the widget size changes.
//   - The cache is freed on Unmount.
//
// RepaintBoundary implements [widget.Widget] and [a11y.Accessible].
//
// Example:
//
//	expensive := primitives.Box(
//	    primitives.Text("Complex chart..."),
//	).Padding(16)
//
//	cached := primitives.NewRepaintBoundary(expensive)
type RepaintBoundary struct {
	widget.WidgetBase

	child widget.Widget

	// cacheKey is a unique monotonic ID for this boundary, used as the key
	// into the Window's dirty boundary set. Assigned once at creation time.
	cacheKey uint64

	// cachedScene holds the recorded display list for the child subtree.
	// On cache hit, this is replayed into the parent canvas via
	// Canvas.ReplayScene — no child.Draw() re-execution needed.
	cachedScene widget.SceneCache

	// cacheVersion is a monotonic counter incremented each time the cache
	// is refreshed. Used for observability and diagnostics.
	cacheVersion uint64
	// cacheWidth and cacheHeight track the cache dimensions to detect
	// size changes that require re-recording.
	cacheWidth  int
	cacheHeight int

	// debugLabel is an optional identifier for diagnostics.
	debugLabel string

	// cacheHits tracks how many times the cache was used (for stats).
	cacheHits int

	// boundaryDirty indicates that a descendant has changed and this
	// boundary's cache needs to be re-recorded. Set by upward dirty
	// propagation (ADR-007) via [MarkBoundaryDirty].
	boundaryDirty bool

	// onBoundaryDirty is a callback invoked when this boundary transitions
	// from clean to dirty. Used by the Window to collect dirty boundaries
	// into its dirtyBoundaries set (ADR-007, Task 1e).
	onBoundaryDirty func(rb *RepaintBoundary)

	// --- RasterCache (ADR-007 Phase 3, Task 3a) ---

	// rasterCacheCfg holds the heuristic parameters for stability tracking.
	// Initialized to DefaultRasterCacheConfig in NewRepaintBoundary.
	rasterCacheCfg RasterCacheConfig

	// consecutiveHits tracks the number of consecutive cache hits since
	// the last cache miss. Used by the raster cache heuristic to decide
	// when to promote the boundary to stable status.
	consecutiveHits int

	// stable indicates that this boundary has been promoted by the raster
	// cache heuristic. Stable boundaries rarely change and are candidates
	// for GPU texture caching (future Phase 4).
	stable bool

	// totalPromotions counts cumulative promotions for diagnostics.
	totalPromotions int

	// totalDemotions counts cumulative demotions for diagnostics.
	totalDemotions int

	// lastSceneVersion tracks the scene version from the last Draw call.
	// Used by the damage-aware compositor (Task 3d) to detect when a
	// boundary's scene content has actually changed between frames.
	lastSceneVersion uint64
}

// Option configures a [RepaintBoundary].
type Option func(*RepaintBoundary)

// WithDebugLabel sets an optional label for diagnostics and logging.
func WithDebugLabel(label string) Option {
	return func(rb *RepaintBoundary) {
		rb.debugLabel = label
	}
}

// NewRepaintBoundary creates a RepaintBoundary that caches the rendering
// of the given child widget as a display list (SceneCache).
//
// If child is nil, the boundary renders nothing and reports zero size.
//
// Options:
//   - [WithDebugLabel] — optional label for diagnostics
func NewRepaintBoundary(child widget.Widget, opts ...Option) *RepaintBoundary {
	rb := &RepaintBoundary{
		child:          child,
		cacheKey:       nextCacheKey.Add(1),
		rasterCacheCfg: DefaultRasterCacheConfig(),
	}
	rb.SetVisible(true)
	rb.SetEnabled(true)

	for _, opt := range opts {
		opt(rb)
	}

	// ADR-028: parent chain for upward dirty propagation.
	// Flutter: RenderObject.adoptChild sets parent on each child.
	if child != nil {
		type parentSetter interface{ SetParent(widget.Widget) }
		if ps, ok := child.(parentSetter); ok {
			ps.SetParent(rb)
		}
	}

	return rb
}

// CacheKey returns the unique monotonic ID for this boundary.
func (rb *RepaintBoundary) CacheKey() uint64 {
	return rb.cacheKey
}

// Child returns the wrapped child widget.
func (rb *RepaintBoundary) Child() widget.Widget {
	return rb.child
}

// DebugLabel returns the diagnostic label, or empty string if none set.
func (rb *RepaintBoundary) DebugLabel() string {
	return rb.debugLabel
}

// CacheHits returns how many times the cache was served instead of re-rendering.
func (rb *RepaintBoundary) CacheHits() int {
	return rb.cacheHits
}

// CacheValid reports whether the cached scene holds valid content.
func (rb *RepaintBoundary) CacheValid() bool {
	return !rb.boundaryDirty && rb.cachedScene != nil
}

// InvalidateCache marks the cache as stale, forcing a re-record on the
// next draw pass. This is called automatically when descendants are dirty;
// manual invocation is rarely needed.
func (rb *RepaintBoundary) InvalidateCache() {
	rb.boundaryDirty = true
}

// MarkBoundaryDirty marks this repaint boundary as needing re-rendering.
//
// Called by the upward dirty propagation in [widget.WidgetBase.SetNeedsRedraw]
// when a descendant widget changes. This invalidates the scene cache and
// notifies the Window (via callback) to add this boundary to its dirty set.
//
// Implements [widget.RepaintBoundaryMarker].
func (rb *RepaintBoundary) MarkBoundaryDirty() {
	if rb.boundaryDirty {
		return // Already dirty — O(1) guard.
	}
	rb.boundaryDirty = true

	// Reset consecutive hits on dirty — the boundary is no longer in a
	// consecutive-cache-hit streak. Demotion from stable happens in Draw
	// (cache miss path) to keep MarkBoundaryDirty O(1).
	rb.consecutiveHits = 0

	if rb.onBoundaryDirty != nil {
		rb.onBoundaryDirty(rb)
	}
}

// IsBoundaryDirty reports whether this boundary has been marked dirty by
// upward propagation since the last draw pass.
func (rb *RepaintBoundary) IsBoundaryDirty() bool {
	return rb.boundaryDirty
}

// ClearBoundaryDirty resets the boundary dirty flag after the boundary
// has been repainted. Called by the Window after painting dirty boundaries.
func (rb *RepaintBoundary) ClearBoundaryDirty() {
	rb.boundaryDirty = false
}

// SetOnBoundaryDirty sets the callback invoked when this boundary
// transitions from clean to dirty via upward propagation.
//
// Used by the Window to collect dirty boundaries into its set.
func (rb *RepaintBoundary) SetOnBoundaryDirty(fn func(rb *RepaintBoundary)) {
	rb.onBoundaryDirty = fn
}

// --- RasterCache (ADR-007 Phase 3, Task 3a) ---

// IsStable reports whether this boundary has been promoted to stable status
// by the raster cache heuristic. Stable boundaries have had consecutive
// cache hits >= the promotion threshold, with sufficient complexity and area.
//
// Stable boundaries are candidates for GPU texture caching (future Phase 4).
// The compositor can use this to prioritize GPU texture allocation.
func (rb *RepaintBoundary) IsStable() bool {
	return rb.stable
}

// ConsecutiveHits returns the number of consecutive cache hits since the
// last cache miss. Used for observability and testing.
func (rb *RepaintBoundary) ConsecutiveHits() int {
	return rb.consecutiveHits
}

// RasterCacheStats returns a snapshot of the raster cache state for this
// boundary. Used for diagnostics, benchmarks, and compositor decisions.
func (rb *RepaintBoundary) RasterCacheStats() RasterCacheStats {
	var tagCount int
	if sc, ok := rb.cachedScene.(*scene.Scene); ok {
		tagCount = len(sc.Encoding().Tags())
	}

	return RasterCacheStats{
		ConsecutiveHits: rb.consecutiveHits,
		IsStable:        rb.stable,
		TagCount:        tagCount,
		Area:            rb.cacheWidth * rb.cacheHeight,
		TotalPromotions: rb.totalPromotions,
		TotalDemotions:  rb.totalDemotions,
	}
}

// evaluatePromotion checks whether this boundary qualifies for promotion
// to stable status based on the raster cache heuristic.
//
// Criteria (Flutter raster_cache.cc pattern):
//   - consecutiveHits >= PromotionThreshold
//   - tag count > MinComplexity
//   - pixel area >= MinArea
func (rb *RepaintBoundary) evaluatePromotion(w, h int) {
	if rb.stable {
		return // Already promoted.
	}

	cfg := rb.rasterCacheCfg
	if rb.consecutiveHits < cfg.PromotionThreshold {
		return
	}

	area := w * h
	if area < cfg.MinArea {
		return
	}

	sc, ok := rb.cachedScene.(*scene.Scene)
	if !ok {
		return
	}
	tagCount := len(sc.Encoding().Tags())
	if tagCount < cfg.MinComplexity {
		return
	}

	rb.stable = true
	rb.totalPromotions++
}

// demoteIfStable resets stable status when the boundary is invalidated.
// Called on cache miss (boundary dirty or first draw).
func (rb *RepaintBoundary) demoteIfStable() {
	if !rb.stable {
		return
	}
	rb.stable = false
	rb.totalDemotions++
}

// --- Damage-Aware Compositor (ADR-007 Phase 3, Task 3d) ---

// SceneVersion returns a monotonic counter that increments each time the
// boundary's scene cache is re-recorded. The compositor uses this to detect
// when a boundary's content has actually changed between frames.
//
// When SceneVersion() == lastObservedVersion, the boundary's scene has not
// changed and the compositor can skip re-compositing this region.
func (rb *RepaintBoundary) SceneVersion() uint64 {
	return rb.cacheVersion
}

// SceneChanged reports whether the boundary's scene has changed since the
// given version. Used by the compositor to determine which boundaries
// need re-compositing in the damage-aware pipeline (Task 3d).
//
// Usage:
//
//	if rb.SceneChanged(lastVersion) {
//	    // Re-composite this boundary's region.
//	    lastVersion = rb.SceneVersion()
//	}
func (rb *RepaintBoundary) SceneChanged(sinceVersion uint64) bool {
	return rb.cacheVersion != sinceVersion
}

// --- widget.Widget interface ---

// Layout delegates to the child and stores the resulting size.
func (rb *RepaintBoundary) Layout(ctx widget.Context, constraints geometry.Constraints) geometry.Size {
	if rb.child == nil {
		size := constraints.Constrain(geometry.Sz(0, 0))
		rb.SetBounds(geometry.FromPointSize(rb.Position(), size))
		return size
	}

	size := widget.LayoutChild(rb.child, ctx, constraints)

	// Position child at origin (no offset within boundary).
	rb.child.(interface{ SetBounds(geometry.Rect) }).SetBounds(
		geometry.FromPointSize(geometry.Pt(0, 0), size),
	)

	rb.SetBounds(geometry.FromPointSize(rb.Position(), size))

	// Invalidate cache if size changed.
	w := int(size.Width)
	h := int(size.Height)
	if w != rb.cacheWidth || h != rb.cacheHeight {
		rb.boundaryDirty = true
		rb.cacheWidth = w
		rb.cacheHeight = h
	}

	return size
}

// Draw renders the child subtree, using the scene cache when possible.
//
// On cache hit (boundary not dirty, cached scene exists): replays the
// cached SceneCache into the canvas via Canvas.ReplayScene — no child
// re-execution. This is O(commands) via Encoding.Append or GPU dispatch.
//
// On cache miss (boundary dirty or first draw): records child.Draw into
// a new SceneCache via SceneCanvas, then replays the result.
//
// This is the ADR-007 retained-mode pattern: display list per boundary.
func (rb *RepaintBoundary) Draw(ctx widget.Context, canvas widget.Canvas) {
	if !rb.IsVisible() || rb.child == nil {
		return
	}

	// Wire the onBoundaryDirty callback on first Draw so that future
	// SetNeedsRedraw → propagateDirtyUpward → MarkBoundaryDirty
	// triggers RequestRedraw on the window. Without this, dirty flags
	// are set but the render loop never wakes up (ADR-007 fix).
	if rb.onBoundaryDirty == nil && ctx != nil {
		capturedCtx := ctx
		rb.onBoundaryDirty = func(_ *RepaintBoundary) {
			capturedCtx.InvalidateRect(rb.Bounds())
		}
	}

	bounds := rb.Bounds()
	w := int(bounds.Width())
	h := int(bounds.Height())
	if w <= 0 || h <= 0 {
		return
	}

	// Cache hit: boundary is clean and we have a cached scene.
	// Suppress damage tracking during replay — cached content hasn't changed,
	// so it must not inflate the gg-level damage list (ADR-021 false positive).
	if !rb.boundaryDirty && rb.cachedScene != nil {
		rb.recordCacheHit(ctx)
		rb.consecutiveHits++
		rb.evaluatePromotion(w, h)
		canvas.PushTransform(bounds.Min)
		if dc, ok := canvas.(widget.DamageController); ok {
			dc.SetDamageTracking(false)
		}
		canvas.ReplayScene(rb.cachedScene)
		if dc, ok := canvas.(widget.DamageController); ok {
			dc.SetDamageTracking(true)
		}
		canvas.PopTransform()
		return
	}

	// Cache miss: reset consecutive hits and demote if stable.
	rb.demoteIfStable()
	rb.consecutiveHits = 0

	// Record child drawing into a scene.
	if rb.cachedScene == nil {
		rb.cachedScene = scene.NewScene()
	}
	rb.cachedScene.Reset()

	recorder := internalRender.NewSceneCanvas(rb.cachedScene.(*scene.Scene), w, h)
	rb.child.Draw(ctx, recorder)
	recorder.Close()

	// Clear redraw flags in the child subtree since we just rendered them.
	widget.ClearRedrawInTree(rb.child)

	rb.boundaryDirty = false
	rb.cacheVersion++
	rb.lastSceneVersion = rb.cacheVersion

	// Replay the freshly recorded scene into the parent canvas.
	canvas.PushTransform(bounds.Min)
	canvas.ReplayScene(rb.cachedScene)
	canvas.PopTransform()
}

// recordCacheHit increments the cache hit counter and records the hit
// in the frame-level DrawStats for observability.
func (rb *RepaintBoundary) recordCacheHit(ctx widget.Context) {
	rb.cacheHits++

	provider, ok := ctx.(widget.DrawStatsProvider)
	if !ok {
		return
	}
	if stats := provider.DrawStats(); stats != nil {
		stats.CachedWidgets++
	}
}

// Event dispatches events to the child.
func (rb *RepaintBoundary) Event(ctx widget.Context, e event.Event) bool {
	if !rb.IsVisible() || !rb.IsEnabled() {
		return false
	}

	if rb.child == nil {
		return false
	}

	// Translate mouse events to local coordinates.
	if me, ok := e.(*event.MouseEvent); ok {
		local := *me
		local.Position = me.Position.Sub(rb.Bounds().Min)
		return rb.child.Event(ctx, &local)
	}

	return rb.child.Event(ctx, e)
}

// Children returns the child widget, or nil if none.
func (rb *RepaintBoundary) Children() []widget.Widget {
	if rb.child == nil {
		return nil
	}
	return []widget.Widget{rb.child}
}

// Unmount releases the scene cache when the widget is removed from the tree.
func (rb *RepaintBoundary) Unmount() {
	rb.cachedScene = nil
	rb.cacheVersion = 0
	rb.cacheWidth = 0
	rb.cacheHeight = 0
	rb.cacheHits = 0
	rb.boundaryDirty = false
	// Reset raster cache state (Task 3a).
	rb.consecutiveHits = 0
	rb.stable = false
	rb.lastSceneVersion = 0
}

// --- a11y.Accessible interface ---

// AccessibilityRole returns [a11y.RoleGenericContainer].
func (rb *RepaintBoundary) AccessibilityRole() a11y.Role {
	return a11y.RoleGenericContainer
}

// AccessibilityLabel returns the debug label or empty string.
func (rb *RepaintBoundary) AccessibilityLabel() string {
	return rb.debugLabel
}

// AccessibilityHint returns an empty string.
func (rb *RepaintBoundary) AccessibilityHint() string {
	return ""
}

// AccessibilityValue returns an empty string.
func (rb *RepaintBoundary) AccessibilityValue() string {
	return ""
}

// AccessibilityState returns the default state.
func (rb *RepaintBoundary) AccessibilityState() a11y.State {
	return a11y.State{
		Disabled: !rb.IsEnabled(),
		Hidden:   !rb.IsVisible(),
	}
}

// AccessibilityActions returns nil.
func (rb *RepaintBoundary) AccessibilityActions() []a11y.Action {
	return nil
}

// Compile-time interface checks.
var (
	_ widget.Widget                = (*RepaintBoundary)(nil)
	_ widget.RepaintBoundaryMarker = (*RepaintBoundary)(nil)
	_ a11y.Accessible              = (*RepaintBoundary)(nil)
)
