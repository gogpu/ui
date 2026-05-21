package widget

import (
	"sync/atomic"

	"github.com/gogpu/gg/scene"
)

// nextBoundaryCacheKey is a monotonic counter for generating unique cache keys.
// Each widget that becomes a RepaintBoundary gets a unique uint64 ID, used for
// deduplication in the dirty boundary tracking set. Atomic to be safe for
// concurrent boundary creation across goroutines.
var nextBoundaryCacheKey atomic.Uint64

// --- RepaintBoundary property (ADR-024 Phase 1) ---
//
// These fields extend WidgetBase to support scene caching without requiring
// a wrapper widget. Any widget can become a repaint boundary by calling
// SetRepaintBoundary(true). This is the Flutter RenderObject.isRepaintBoundary
// pattern: a boolean property on the base class, not a wrapper node.

// SetRepaintBoundary marks this widget as a repaint boundary.
//
// When enabled, the widget owns a scene.Scene display list that caches
// its subtree rendering. Clean boundaries replay their cached scene
// instead of re-executing Draw on every descendant.
//
// This is equivalent to Flutter's RenderObject.isRepaintBoundary and
// Android's View.setLayerType(LAYER_TYPE_HARDWARE).
//
// Calling this with false disables boundary behavior and releases the
// cached scene.
func (w *WidgetBase) SetRepaintBoundary(enabled bool) {
	w.mu.Lock()
	defer w.mu.Unlock()

	if w.isRepaintBoundary == enabled {
		return
	}

	w.isRepaintBoundary = enabled
	if enabled {
		// Assign a unique cache key for this boundary.
		if w.boundaryCacheKey == 0 {
			w.boundaryCacheKey = nextBoundaryCacheKey.Add(1)
		}
		// Start dirty so first draw records the scene.
		w.sceneDirty = true
	} else {
		// Release cached scene when disabling boundary.
		w.cachedScene = nil
		w.sceneDirty = false
		w.sceneCacheVersion = 0
		w.sceneCacheWidth = 0
		w.sceneCacheHeight = 0
	}
}

// IsRepaintBoundary reports whether this widget is a repaint boundary.
//
// Repaint boundaries own a scene.Scene that caches their subtree rendering.
// The DrawTree function checks this property and replays the cached scene
// when the boundary is clean, avoiding re-execution of the child Draw methods.
func (w *WidgetBase) IsRepaintBoundary() bool {
	w.mu.RLock()
	defer w.mu.RUnlock()
	return w.isRepaintBoundary
}

// BoundaryCacheKey returns the unique monotonic ID for this boundary.
// Returns 0 if the widget is not a repaint boundary.
func (w *WidgetBase) BoundaryCacheKey() uint64 {
	w.mu.RLock()
	defer w.mu.RUnlock()
	return w.boundaryCacheKey
}

// InvalidateScene marks this boundary's cached scene as stale, forcing
// a re-record on the next draw pass. This is called automatically when
// descendants call SetNeedsRedraw (upward dirty propagation via
// propagateDirtyUpward).
//
// If this widget is not a repaint boundary, this is a no-op.
// If the scene is already dirty, this is a no-op (O(1) guard).
//
// Triggers the onBoundaryDirty callback to notify the Window.
func (w *WidgetBase) InvalidateScene() {
	w.mu.Lock()
	if !w.isRepaintBoundary {
		w.mu.Unlock()
		return
	}
	if w.sceneDirty {
		w.mu.Unlock()
		return // Already dirty — O(1) guard.
	}
	w.sceneDirty = true
	cb := w.onBoundaryDirty
	suppress := w.suppressDirtyCallback
	w.mu.Unlock()

	// During Draw recording (suppressDirtyCallback=true), the boundary dirty
	// callback is suppressed. Animated widgets call ScheduleAnimationFrame()
	// explicitly to request deferred render. This prevents the immediate
	// RequestRedraw chain that forces 60fps for 30fps animations.
	// External events (hover, click) set dirty OUTSIDE Draw — callback fires
	// immediately for instant user feedback.
	if cb != nil && !suppress {
		cb()
	}
}

// SetSuppressDirtyCallback controls whether onBoundaryDirty callback fires
// during InvalidateScene. Set to true during Draw recording so animated
// widgets can defer render requests via ScheduleAnimationFrame instead of
// triggering immediate RequestRedraw.
func (w *WidgetBase) SetSuppressDirtyCallback(v bool) {
	w.mu.Lock()
	w.suppressDirtyCallback = v
	w.mu.Unlock()
}

// IsSceneDirty reports whether the boundary's cached scene needs re-recording.
func (w *WidgetBase) IsSceneDirty() bool {
	w.mu.RLock()
	defer w.mu.RUnlock()
	return w.sceneDirty
}

// CachedScene returns the boundary's cached scene, or nil if no cache exists.
// This is used by DrawTree to replay the scene when the boundary is clean.
func (w *WidgetBase) CachedScene() *scene.Scene {
	w.mu.RLock()
	defer w.mu.RUnlock()
	return w.cachedScene
}

// SetCachedScene stores the recorded scene for this boundary.
// Called by the render system after recording the subtree.
func (w *WidgetBase) SetCachedScene(s *scene.Scene) {
	w.mu.Lock()
	defer w.mu.Unlock()
	w.cachedScene = s
}

// ClearCachedScene releases the cached scene to free memory.
// Called during UnmountTree to prevent leaking scene data after SetRoot.
func (w *WidgetBase) ClearCachedScene() {
	w.mu.Lock()
	defer w.mu.Unlock()
	w.cachedScene = nil
}

// ClearSceneDirty resets the sceneDirty flag after the boundary has been
// re-recorded. Called by the render system after a successful record pass.
func (w *WidgetBase) ClearSceneDirty() {
	w.mu.Lock()
	defer w.mu.Unlock()
	w.sceneDirty = false
	w.sceneCacheVersion++
}

// SceneCacheVersion returns a monotonic counter that increments each time
// the boundary's scene is re-recorded. Used by the compositor to detect
// when content has actually changed between frames.
func (w *WidgetBase) SceneCacheVersion() uint64 {
	w.mu.RLock()
	defer w.mu.RUnlock()
	return w.sceneCacheVersion
}

// SceneCacheSize returns the cached scene dimensions (width, height).
// Returns (0, 0) if no cache exists.
func (w *WidgetBase) SceneCacheSize() (int, int) {
	w.mu.RLock()
	defer w.mu.RUnlock()
	return w.sceneCacheWidth, w.sceneCacheHeight
}

// SetSceneCacheSize records the dimensions of the cached scene.
// If the widget's bounds change, the caller should invalidate the scene.
func (w *WidgetBase) SetSceneCacheSize(width, height int) {
	w.mu.Lock()
	defer w.mu.Unlock()
	w.sceneCacheWidth = width
	w.sceneCacheHeight = height
}

// SetOnBoundaryDirty sets the callback invoked when this boundary transitions
// from clean to dirty via upward propagation. Used by the Window to collect
// dirty boundaries into its set and request a redraw.
func (w *WidgetBase) SetOnBoundaryDirty(fn func()) {
	w.mu.Lock()
	defer w.mu.Unlock()
	w.onBoundaryDirty = fn
}
