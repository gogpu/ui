package widget

import (
	"github.com/gogpu/ui/geometry"
)

// SceneRecorder creates a recording Canvas that writes draw commands into a
// SceneCache. This is the dependency-injection point for ADR-024 Phase 2:
// the widget package cannot import internal/render (circular dep), so the
// app layer registers a factory function that creates SceneCanvas instances.
//
// The returned Canvas records all drawing operations into the given scene.
// After recording, the scene can be replayed via Canvas.ReplayScene.
//
// Parameters:
//   - s: the SceneCache to record into (must not be nil)
//   - width, height: dimensions of the recording canvas
//
// Returns a Canvas that records into s, and a cleanup function that must
// be called after recording is complete (e.g., SceneCanvas.Close).
type SceneRecorder func(s SceneCache, width, height int) (Canvas, func())

// sceneRecorderFactory holds the registered SceneRecorder factory.
// Set by the app layer during initialization via RegisterSceneRecorder.
var sceneRecorderFactory SceneRecorder

// RegisterSceneRecorder registers the factory function for creating scene
// recording canvases. This must be called by the app layer before any
// boundary draws occur (typically in package init or Window creation).
//
// Example (from app package):
//
//	widget.RegisterSceneRecorder(func(s widget.SceneCache, w, h int) (widget.Canvas, func()) {
//	    recorder := render.NewSceneCanvas(s.(*scene.Scene), w, h)
//	    return recorder, recorder.Close
//	})
func RegisterSceneRecorder(factory SceneRecorder) {
	sceneRecorderFactory = factory
}

// GetSceneRecorderFactory returns the registered SceneRecorder factory.
// Returns nil if no factory has been registered.
func GetSceneRecorderFactory() SceneRecorder {
	return sceneRecorderFactory
}

// boundaryWidget is the interface that widgets with WidgetBase boundary
// support must satisfy. All methods are provided by WidgetBase embedding.
type boundaryWidget interface {
	Widget
	IsRepaintBoundary() bool
	IsSceneDirty() bool
	CachedScene() SceneCache
	SetCachedScene(SceneCache)
	ClearSceneDirty()
	SceneCacheSize() (int, int)
	SetSceneCacheSize(int, int)
	SetOnBoundaryDirty(func())
	Bounds() geometry.Rect
}

// drawBoundaryWidget handles the draw pass for a widget that has
// isRepaintBoundary == true. It implements the cache-hit/miss logic:
//
//   - Cache hit (not dirty, scene exists): replay cached scene with damage
//     suppression (clean content hasn't changed, must not inflate damage list).
//   - Cache miss (dirty or first draw): record child Draw into scene, then replay.
//
// This is called from drawTreeRecursive when the widget is a boundary.
// If no SceneRecorder factory is registered, falls back to normal draw.
func drawBoundaryWidget(w Widget, ctx Context, canvas Canvas, stats *DrawStats) { //nolint:gocyclo,cyclop // boundary recording is inherently complex (cache hit/miss, dirty callback, size change, screen origin, device scale)
	bw, ok := w.(boundaryWidget)
	if !ok || sceneRecorderFactory == nil {
		// Fallback: draw normally without boundary caching.
		w.Draw(ctx, canvas)
		return
	}

	// Wire onBoundaryDirty callback on first Draw so that future
	// SetNeedsRedraw → propagateDirtyUpward → InvalidateScene
	// triggers RequestRedraw on the window. Without this, dirty flags
	// are set but the render loop never wakes up.
	if bw.CachedScene() == nil && bw.IsSceneDirty() && ctx != nil {
		capturedBW := bw
		bw.SetOnBoundaryDirty(func() {
			ctx.InvalidateRect(capturedBW.Bounds())
		})
	}

	// Get widget dimensions from bounds.
	bounds := bw.Bounds()
	width := int(bounds.Width())
	height := int(bounds.Height())

	if width <= 0 || height <= 0 {
		// Widget has not been laid out yet (zero bounds). Fall back to normal
		// draw without scene caching. This happens when DrawTo is called before
		// layout (e.g., in tests or host-managed mode before first Frame).
		w.Draw(ctx, canvas)
		return
	}

	// Check for size change — invalidate cache if dimensions changed.
	cw, ch := bw.SceneCacheSize()
	if cw != width || ch != height {
		bw.SetSceneCacheSize(width, height)
		bw.SetCachedScene(nil) // Force re-record on size change.
	}

	// Cache hit: boundary is clean and we have a cached scene.
	if !bw.IsSceneDirty() && bw.CachedScene() != nil {
		if stats != nil {
			stats.CachedWidgets++
		}
		// Stamp screen origin even on cache hit so dirty.Collector gets
		// correct screen positions. Draw is NOT called on cache hit,
		// so StampScreenOrigin inside Draw never runs.
		StampScreenOrigin(w, canvas)
		canvas.PushTransform(bounds.Min)
		if dc, ok2 := canvas.(DamageController); ok2 {
			dc.SetDamageTracking(false)
		}
		canvas.ReplayScene(bw.CachedScene())
		if dc, ok2 := canvas.(DamageController); ok2 {
			dc.SetDamageTracking(true)
		}
		canvas.PopTransform()

		// Flutter compositeFrame: even when parent is cache-hit, dirty
		// child boundaries must still be re-recorded and replayed.
		// Without this, animated children (spinner) inside a clean
		// parent boundary would freeze after the first frame.
		visitDirtyChildBoundaries(w, ctx, canvas, stats)
		return
	}

	// Cache miss: record child drawing into a scene.
	cachedScene := bw.CachedScene()
	if cachedScene == nil {
		cachedScene = NewSceneCache()
	}
	cachedScene.Reset()

	recorder, cleanup := sceneRecorderFactory(cachedScene, width, height)
	// Set screen-space base offset so StampScreenOrigin inside Draw computes
	// correct screen positions for dirty.Collector. Without this, PushTransform
	// (-bounds.Min) shifts TransformOffset to local coords → ScreenOrigin = (0,0)
	// → overlay shows dirty regions at wrong positions.
	// Flutter: PaintingContext carries offset from parent for screen-space mapping.
	screenBase := canvas.TransformOffset().Add(canvas.ScreenOriginBase()).Add(bounds.Min)
	type screenBaseSetter interface{ SetScreenOriginBase(geometry.Point) }
	if sbs, ok := recorder.(screenBaseSetter); ok {
		sbs.SetScreenOriginBase(screenBase)
	}
	// Propagate device scale for HiDPI-aware SVG icon rasterization (ADR-026).
	if ctx != nil {
		if ds, ok := recorder.(DeviceScaler); ok {
			ds.SetDeviceScale(ctx.Scale())
		}
	}
	StampScreenOrigin(w, canvas)
	recorder.PushTransform(geometry.Pt(-bounds.Min.X, -bounds.Min.Y))

	// Clear dirty BEFORE Draw so we can detect re-dirtying during Draw.
	// Animated widgets (spinner) call SetNeedsRedraw(true) inside Draw
	// to request the next animation frame. If we cleared AFTER Draw,
	// we'd erase that request → animation freezes at 1fps.
	// Flutter: PaintingContext clears _needsPaint BEFORE calling paint().
	bw.ClearSceneDirty()
	ClearRedrawInTree(w)

	// Suppress boundary dirty callback during Draw recording (see ADR-007
	// AnimationScheduler). Animated widgets defer render via ScheduleAnimationFrame.
	type dirtySuppressor interface{ SetSuppressDirtyCallback(bool) }
	if ds, ok := w.(dirtySuppressor); ok {
		ds.SetSuppressDirtyCallback(true)
	}
	w.Draw(ctx, recorder)
	if ds, ok := w.(dirtySuppressor); ok {
		ds.SetSuppressDirtyCallback(false)
	}
	recorder.PopTransform()
	cleanup()

	// Store the freshly recorded scene.
	bw.SetCachedScene(cachedScene)

	// Replay the freshly recorded scene into the parent canvas.
	canvas.PushTransform(bounds.Min)
	canvas.ReplayScene(cachedScene)
	canvas.PopTransform()
}

// visitDirtyChildBoundaries walks the subtree looking for child boundaries
// that are dirty and need re-recording. This is called after a parent
// boundary cache-hit to ensure animated children (spinner) still update.
//
// Flutter compositeFrame walks the layer tree; clean layers use addRetained
// while dirty layers re-compose. This function provides the same guarantee
// at the DrawTree level: clean parent + dirty child = child still draws.
func visitDirtyChildBoundaries(w Widget, ctx Context, canvas Canvas, stats *DrawStats) {
	for _, child := range w.Children() {
		if child == nil {
			continue
		}

		if bw, ok := child.(boundaryWidget); ok && bw.IsRepaintBoundary() {
			if bw.IsSceneDirty() || bw.CachedScene() == nil {
				drawBoundaryWidget(child, ctx, canvas, stats)
			} else {
				// Child boundary is clean — still check ITS children
				// for deeper dirty boundaries.
				visitDirtyChildBoundaries(child, ctx, canvas, stats)
			}
			continue
		}

		// Non-boundary child: recurse looking for boundaries deeper.
		visitDirtyChildBoundaries(child, ctx, canvas, stats)
	}
}
