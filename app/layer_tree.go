package app

import (
	"github.com/gogpu/gpucontext"
	"github.com/gogpu/ui/compositor"
	"github.com/gogpu/ui/geometry"
	"github.com/gogpu/ui/widget"
)

// externalTextureWidget is implemented by widgets that own a GPU texture
// rendered by an external producer (GPUView, video players, custom shaders).
// When detected, buildBoundaryLayer appends an ExternalTextureLayer alongside
// the PictureLayer so the compositor blits the external texture.
//
// This is a local interface (unexported) following the established type-assertion
// pattern in this file (compositorClipper, callbackSetter, etc.).
//
// Flutter equivalent: TextureLayer (rendering/layer.dart).
type externalTextureWidget interface {
	// Texture returns the external GPU texture view. Returns a nil/zero
	// TextureView if the texture has not been initialized yet.
	Texture() gpucontext.TextureView

	// ViewportSize returns the viewport dimensions in logical pixels.
	ViewportSize() (width, height int)
}

// boundaryInfo describes a widget that is a RepaintBoundary.
type boundaryInfo interface {
	widget.Widget
	IsRepaintBoundary() bool
	IsSceneDirty() bool
	CachedScene() widget.SceneCache
	SetCachedScene(widget.SceneCache)
	ClearSceneDirty()
	SceneCacheSize() (int, int)
	SetSceneCacheSize(int, int)
	Bounds() geometry.Rect
	ScreenOrigin() geometry.Point
	BoundaryCacheKey() uint64
	SceneCacheVersion() uint64
	Parent() widget.Widget
}

// layerIndex maps BoundaryCacheKey to the PictureLayerImpl + its parent
// OffsetLayerImpl from a previous frame. Used by UpdateLayerTree to reuse
// layer objects across frames (zero allocation for unchanged boundaries).
type layerIndex struct {
	pic    *compositor.PictureLayerImpl
	offset *compositor.OffsetLayerImpl
}

// BuildLayerTree walks the widget tree and constructs a compositor layer tree.
// Each RepaintBoundary widget produces a PictureLayer inside an OffsetLayer.
// Non-boundary widgets are skipped (they're drawn inside their parent boundary).
//
// ADR-007 Phase D: Layer Tree provides STRUCTURE (which boundaries exist,
// their offsets, clip rects, opacity) for the texture rendering/blitting
// pipeline. PictureLayerImpl stores BoundaryCacheKey, IsRoot, and Size to
// link back to the per-boundary GPU texture cache in renderLoop.
//
// Flutter equivalent: Layer tree is built during paint via paintChild.
func BuildLayerTree(root widget.Widget) *compositor.OffsetLayerImpl {
	if root == nil {
		return compositor.NewOffsetLayer(geometry.Point{})
	}

	rootLayer := compositor.NewOffsetLayer(geometry.Point{})
	buildLayerRecursive(root, rootLayer, 0, 0)
	return rootLayer
}

// UpdateLayerTree builds or updates a persistent layer tree. On the first call
// (existing == nil), it builds from scratch (same as BuildLayerTree). On
// subsequent calls, it reuses PictureLayerImpl and OffsetLayerImpl objects
// for boundaries that still exist (matched by BoundaryCacheKey), updating
// their fields from the current widget state. New boundaries get fresh layers;
// removed boundaries are dropped.
//
// Flutter equivalent: Layer.addRetained + ContainerLayer.updateSubtreeNeedsAddToScene.
// The persistent tree eliminates per-frame layer allocations for stable UIs.
//
// Returns the root OffsetLayerImpl (may be the same pointer as existing or new).
func UpdateLayerTree(root widget.Widget, existing *compositor.OffsetLayerImpl) *compositor.OffsetLayerImpl {
	if existing == nil {
		return BuildLayerTree(root)
	}
	if root == nil {
		return compositor.NewOffsetLayer(geometry.Point{})
	}

	// Collect existing layers indexed by BoundaryCacheKey for O(1) lookup.
	index := collectLayerIndex(existing)

	// Build new tree structure, reusing existing layer objects where possible.
	newRoot := compositor.NewOffsetLayer(geometry.Point{})
	updateLayerRecursive(root, newRoot, index, 0, 0)

	return newRoot
}

// collectLayerIndex walks an existing layer tree and builds a map from
// BoundaryCacheKey to the PictureLayerImpl + parent OffsetLayerImpl pair.
func collectLayerIndex(root compositor.Layer) map[uint64]layerIndex {
	idx := make(map[uint64]layerIndex)
	collectLayerIndexRecursive(root, idx)
	return idx
}

func collectLayerIndexRecursive(layer compositor.Layer, idx map[uint64]layerIndex) {
	if layer == nil {
		return
	}

	// OffsetLayer with a PictureLayer child = boundary pair.
	offset, isOffset := layer.(*compositor.OffsetLayerImpl)
	if isOffset {
		for _, ch := range offset.Children() {
			if pic, ok := ch.(*compositor.PictureLayerImpl); ok {
				key := pic.BoundaryCacheKey()
				if key != 0 {
					idx[key] = layerIndex{pic: pic, offset: offset}
				}
			}
		}
	}

	// Recurse into container children.
	if cl, ok := layer.(compositor.ContainerLayer); ok {
		for _, ch := range cl.Children() {
			collectLayerIndexRecursive(ch, idx)
		}
	}
}

// updateLayerRecursive mirrors buildLayerRecursive but reuses existing layers.
func updateLayerRecursive(w widget.Widget, parentLayer compositor.ContainerLayer, index map[uint64]layerIndex, localX, localY float32) {
	if w == nil {
		return
	}

	type boundsGetter interface{ Bounds() geometry.Rect }
	var boundsMin geometry.Point
	if bg, ok := w.(boundsGetter); ok {
		boundsMin = bg.Bounds().Min
	}

	bi, isBoundary := w.(boundaryInfo)
	if isBoundary && bi.IsRepaintBoundary() {
		offset := geometry.Pt(localX+boundsMin.X, localY+boundsMin.Y)
		childOffset := updateBoundaryLayer(bi, w, offset, index)
		parentLayer.Append(childOffset)

		// Recurse into children with reset offset (boundary OffsetLayer
		// already accounts for position).
		for _, child := range w.Children() {
			updateLayerRecursive(child, childOffset, index, 0, 0)
		}
		return
	}

	// Non-boundary: accumulate offset and recurse.
	nextX := localX + boundsMin.X
	nextY := localY + boundsMin.Y
	for _, child := range w.Children() {
		updateLayerRecursive(child, parentLayer, index, nextX, nextY)
	}
}

// updateBoundaryLayer reuses or creates an OffsetLayer + PictureLayer pair
// for a RepaintBoundary widget. If the boundary's cache key exists in the
// index, the existing PictureLayerImpl is reused (fields updated in place).
// Otherwise, a fresh pair is created (same as buildBoundaryLayer).
func updateBoundaryLayer(bi boundaryInfo, w widget.Widget, offset geometry.Point, index map[uint64]layerIndex) *compositor.OffsetLayerImpl {
	key := bi.BoundaryCacheKey()
	existing, found := index[key]

	if found && existing.pic != nil {
		// Reuse existing layers. Detach from old parent to prevent
		// double-parenting when Append attaches to new tree.
		existing.offset.RemoveAll()
		existing.offset.SetOffset(offset)

		// Update PictureLayer fields from current widget state.
		syncPictureLayer(existing.pic, bi, w)
		existing.offset.Append(existing.pic)

		// ExternalTextureWidget: re-append ExternalTextureLayer with current
		// texture state. RemoveAll() cleared previous layers — fresh append
		// ensures the texture reference and position are up-to-date.
		appendExternalTextureLayer(existing.offset, w, bi)

		// Mark as consumed so cleanup can detect removed boundaries.
		delete(index, key)

		return existing.offset
	}

	// No existing layer — create fresh (same as buildBoundaryLayer).
	delete(index, key) // no-op if not found, but consistent
	return buildBoundaryLayer(bi, w, offset)
}

// syncPictureLayer updates a reused PictureLayerImpl's fields from the
// current boundary widget state. This is the per-frame O(1) update that
// replaces allocating a new PictureLayerImpl.
func syncPictureLayer(pic *compositor.PictureLayerImpl, bi boundaryInfo, w widget.Widget) {
	cachedScene := bi.CachedScene()
	if cachedScene != nil {
		pic.SetPicture(cachedScene)
	}
	if bi.IsSceneDirty() {
		pic.MarkDirty()
	} else {
		pic.ClearDirty()
	}

	pic.SetBoundaryCacheKey(bi.BoundaryCacheKey())
	pic.SetRoot(bi.Parent() == nil)
	pic.SetSceneVersion(bi.SceneCacheVersion())

	bounds := bi.Bounds()
	pic.SetSize(int(bounds.Width()), int(bounds.Height()))
	pic.SetScreenOrigin(bi.ScreenOrigin())

	// Update compositor clip for viewport culling.
	type compositorClipper interface {
		HasCompositorClip() bool
		CompositorClip() geometry.Rect
		IsScreenOriginValid() bool
	}
	if cc, ok := w.(compositorClipper); ok && cc.HasCompositorClip() {
		pic.SetPictureClipRect(cc.CompositorClip())
	}
}

// buildLayerRecursive walks the widget tree, adding PictureLayer for each boundary.
// localX/localY accumulate offsets from non-boundary ancestors, so each
// boundary's OffsetLayer gets the correct position relative to its
// parent boundary (not just its immediate parent widget).
func buildLayerRecursive(w widget.Widget, parentLayer compositor.ContainerLayer, localX, localY float32) {
	if w == nil {
		return
	}

	type boundsGetter interface{ Bounds() geometry.Rect }
	var boundsMin geometry.Point
	if bg, ok := w.(boundsGetter); ok {
		boundsMin = bg.Bounds().Min
	}

	bi, isBoundary := w.(boundaryInfo)
	if isBoundary && bi.IsRepaintBoundary() {
		offset := geometry.Pt(localX+boundsMin.X, localY+boundsMin.Y)
		childOffset := buildBoundaryLayer(bi, w, offset)
		parentLayer.Append(childOffset)

		// Recurse into children. Local offset resets to (0,0) because
		// this boundary's OffsetLayer already accounts for its position.
		for _, child := range w.Children() {
			buildLayerRecursive(child, childOffset, 0, 0)
		}
		return
	}

	// Non-boundary widget: accumulate its bounds.Min and recurse.
	nextX := localX + boundsMin.X
	nextY := localY + boundsMin.Y
	for _, child := range w.Children() {
		buildLayerRecursive(child, parentLayer, nextX, nextY)
	}
}

// buildBoundaryLayer creates an OffsetLayer + PictureLayer for a RepaintBoundary widget.
// Populates all Phase D fields (cache key, root flag, scene version, size,
// screen origin, compositor clip) from the boundary widget.
func buildBoundaryLayer(bi boundaryInfo, w widget.Widget, offset geometry.Point) *compositor.OffsetLayerImpl {
	childOffset := compositor.NewOffsetLayer(offset)
	pic := compositor.NewPictureLayer()

	cachedScene := bi.CachedScene()
	if cachedScene != nil {
		pic.SetPicture(cachedScene)
	}
	if bi.IsSceneDirty() {
		pic.MarkDirty()
	} else {
		pic.ClearDirty()
	}

	// Phase D: populate fields that link PictureLayer to GPU texture cache.
	pic.SetBoundaryCacheKey(bi.BoundaryCacheKey())
	pic.SetRoot(bi.Parent() == nil)
	pic.SetSceneVersion(bi.SceneCacheVersion())
	bounds := bi.Bounds()
	pic.SetSize(int(bounds.Width()), int(bounds.Height()))

	// Store screen origin for compositor blit positioning.
	pic.SetScreenOrigin(bi.ScreenOrigin())

	// Store compositor clip for viewport culling (ScrollView items).
	type compositorClipper interface {
		HasCompositorClip() bool
		CompositorClip() geometry.Rect
		IsScreenOriginValid() bool
	}
	if cc, ok := w.(compositorClipper); ok && cc.HasCompositorClip() {
		pic.SetPictureClipRect(cc.CompositorClip())
	}

	childOffset.Append(pic)

	// ExternalTextureWidget (GPUView, video): append an ExternalTextureLayer
	// so the compositor blits the external GPU texture. The PictureLayer stays
	// for scene recording compatibility; the ExternalTextureLayer provides the
	// actual content. compositeFromTreeRecursive handles both layer types.
	appendExternalTextureLayer(childOffset, w, bi)

	return childOffset
}

// appendExternalTextureLayer checks whether a boundary widget implements
// externalTextureWidget (GPUView, video players). If so and the texture
// is valid, appends an ExternalTextureLayer to the parent OffsetLayer.
//
// Called by buildBoundaryLayer (fresh build) and updateBoundaryLayer (reuse).
// The ExternalTextureLayer is a sibling of the PictureLayer — both live
// inside the same OffsetLayer.
func appendExternalTextureLayer(parent *compositor.OffsetLayerImpl, w widget.Widget, bi boundaryInfo) {
	etw, ok := w.(externalTextureWidget)
	if !ok {
		return
	}
	tex := etw.Texture()
	if tex.IsNil() {
		return
	}
	vw, vh := etw.ViewportSize()
	if vw <= 0 || vh <= 0 {
		return
	}
	origin := bi.ScreenOrigin()
	ext := compositor.NewExternalTextureLayer(tex, vw, vh, float64(origin.X), float64(origin.Y))
	parent.Append(ext)
}

// PaintBoundaryLayers walks the widget tree and re-records dirty boundaries.
// This is the Flutter flushPaint equivalent: only dirty boundary PictureLayers
// are re-recorded. Clean boundaries keep their cached scenes.
//
// After this function, all boundary CachedScene values are fresh.
// The compositor can then Compose the layer tree to assemble the final scene.
// PaintBoundaryLayers re-records dirty boundaries with nil context.
func PaintBoundaryLayers(root widget.Widget, _ *compositor.OffsetLayerImpl) {
	PaintBoundaryLayersWithContext(root, nil, nil)
}

// PaintBoundaryLayersWithContext re-records dirty boundaries with a given context.
func PaintBoundaryLayersWithContext(root widget.Widget, _ *compositor.OffsetLayerImpl, ctx widget.Context) {
	if root == nil {
		return
	}
	paintBoundaryRecursiveCtx(root, ctx)
}

// PaintOverlayBoundaries re-records dirty overlay content boundaries.
// Overlay content widgets are already marked as RepaintBoundary by PushOverlay.
// This function walks each overlay content widget (same as PaintBoundaryLayers
// walks the main tree) so their CachedScene values are fresh for the compositor.
func PaintOverlayBoundaries(overlayWidgets []widget.Widget, ctx widget.Context) {
	for _, w := range overlayWidgets {
		if w == nil {
			continue
		}
		paintBoundaryRecursiveCtx(w, ctx)
	}
}

// AppendOverlaysToLayerTree adds overlay content boundaries to an existing
// Layer Tree. Overlays are appended AFTER main tree children, so they
// composite on top (correct Z-order: main content → overlays bottom-to-top).
//
// Each overlay content widget that is a RepaintBoundary gets its own
// OffsetLayer + PictureLayer in the tree, just like main tree boundaries.
// Non-boundary overlay widgets are skipped (they have no scene to composite).
//
// The existing parameter is used for persistent tree reuse: overlay layers
// from previous frames are matched by BoundaryCacheKey.
func AppendOverlaysToLayerTree(tree *compositor.OffsetLayerImpl, overlayWidgets []widget.Widget, existing *compositor.OffsetLayerImpl) {
	if tree == nil || len(overlayWidgets) == 0 {
		return
	}

	// Record child count before appending so we can fix overlay-specific
	// flags on the newly added layers below.
	preCount := len(tree.Children())

	// Collect existing overlay layers for reuse.
	var index map[uint64]layerIndex
	if existing != nil {
		index = collectLayerIndex(existing)
	}

	for _, w := range overlayWidgets {
		if w == nil {
			continue
		}
		if index != nil {
			updateLayerRecursive(w, tree, index, 0, 0)
		} else {
			buildLayerRecursive(w, tree, 0, 0)
		}
	}

	// Fix IsRoot flag on overlay PictureLayers. Overlay content widgets have
	// Parent() == nil (they're standalone, not part of the main tree), so
	// buildBoundaryLayer/syncPictureLayer sets IsRoot=true. This causes
	// DrawGPUTextureBase (QueueBaseLayer, last-call-wins) to overwrite the
	// actual root texture with the overlay texture → black background.
	// Overlays must use DrawGPUTexture (sublayer blit) instead.
	for _, child := range tree.Children()[preCount:] {
		clearRootOnPictureLayers(child)
	}
}

// clearRootOnPictureLayers walks a layer subtree and sets IsRoot=false on
// every PictureLayerImpl. Used by AppendOverlaysToLayerTree to prevent
// overlay boundaries from being treated as the base layer during compositing.
func clearRootOnPictureLayers(layer compositor.Layer) {
	if layer == nil {
		return
	}
	if pic, ok := layer.(*compositor.PictureLayerImpl); ok {
		pic.SetRoot(false)
		return
	}
	if cl, ok := layer.(compositor.ContainerLayer); ok {
		for _, ch := range cl.Children() {
			clearRootOnPictureLayers(ch)
		}
	}
}

// paintBoundaryRecursiveCtx walks the widget tree, re-recording dirty boundaries.
func paintBoundaryRecursiveCtx(w widget.Widget, ctx widget.Context) {
	paintBoundaryWithDepth(w, ctx, 0)
}

func paintBoundaryWithDepth(w widget.Widget, ctx widget.Context, _ int) {
	if w == nil {
		return
	}

	bi, isBoundary := w.(boundaryInfo)
	if isBoundary && bi.IsRepaintBoundary() {
		// Record only if dirty AND visible. Offscreen boundaries (outside
		// CompositorClip viewport) are skipped: Draw never runs →
		// ScheduleAnimationFrame never called → animation pumper stops.
		// Scene stays dirty so it re-records when scrolled back into view.
		if (bi.IsSceneDirty() || bi.CachedScene() == nil) && isBoundaryVisible(bi) {
			recordBoundary(bi, ctx)
		}

		for _, child := range w.Children() {
			paintBoundaryWithDepth(child, ctx, 0)
		}
		return
	}

	for _, child := range w.Children() {
		paintBoundaryWithDepth(child, ctx, 0)
	}
}

// recordBoundary re-records a boundary widget's scene via SceneCanvas.
func recordBoundary(bi boundaryInfo, ctx widget.Context) {
	// Wire onBoundaryDirty callback so animated widgets (spinner) that call
	// SetNeedsRedraw during Draw trigger RequestRedraw for the next frame.
	// Without this, InvalidateScene sets sceneDirty but nobody wakes the
	// render loop → animation frozen at data ticker rate (1fps).
	type callbackSetter interface {
		SetOnBoundaryDirty(func())
	}
	if cs, ok := bi.(callbackSetter); ok && ctx != nil {
		capturedBi := bi
		cs.SetOnBoundaryDirty(func() {
			// ADR-028 Phase C: register in flat dirty boundary set for O(1)
			// frame skip. Flutter _nodesNeedingPaint.add() equivalent.
			//
			// NOTE: do NOT call ctx.InvalidateRect here. InvalidateRect sets
			// window.needsRedraw=true which forces ROOT re-recording every
			// frame. Child boundary dirty should only re-record the child —
			// not the root. RegisterDirtyBoundary adds to flat dirty set AND
			// wakes the render loop via RequestRedraw (wired in window.go).
			type cacheKeyProvider interface {
				BoundaryCacheKey() uint64
			}
			if reg, ok := ctx.(widget.DirtyBoundaryRegistrar); ok {
				if ckp, ok2 := capturedBi.(cacheKeyProvider); ok2 {
					reg.RegisterDirtyBoundary(ckp.BoundaryCacheKey())
				}
			}
		})
	}
	bounds := bi.Bounds()
	width := int(bounds.Width())
	height := int(bounds.Height())

	if width <= 0 || height <= 0 {
		return
	}

	// Check size change.
	cw, ch := bi.SceneCacheSize()
	if cw != width || ch != height {
		bi.SetSceneCacheSize(width, height)
	}

	cachedScene := bi.CachedScene()
	if cachedScene == nil {
		cachedScene = widget.NewSceneCache()
	}
	cachedScene.Reset()

	if widget.GetSceneRecorderFactory() == nil {
		return
	}

	recorder, cleanup := widget.GetSceneRecorderFactory()(cachedScene, width, height)

	// Propagate device scale for HiDPI-aware SVG icon rasterization (ADR-026).
	if ctx != nil {
		if ds, ok := recorder.(widget.DeviceScaler); ok {
			ds.SetDeviceScale(ctx.Scale())
		}
	}

	// Clear dirty BEFORE Draw (Flutter pattern: detect re-dirtying during Draw).
	bi.ClearSceneDirty()
	widget.ClearRedrawInTree(bi)

	// Suppress boundary dirty callback during recording. Animated widgets
	// (spinner) call SetNeedsRedraw inside Draw which triggers InvalidateScene.
	// Without suppression, this fires onBoundaryDirty → ctx.InvalidateRect →
	// immediate RequestRedraw → 60fps forced. With suppression, the widget
	// uses ScheduleAnimationFrame for deferred render at animPumper rate.
	// External events (hover, click) fire OUTSIDE Draw → not suppressed.
	type dirtySuppressor interface{ SetSuppressDirtyCallback(bool) }
	if ds, ok := bi.(dirtySuppressor); ok {
		ds.SetSuppressDirtyCallback(true)
	}

	// Set ScreenOriginBase so StampScreenOrigin inside Draw computes correct
	// screen-space origins for child widgets (Flutter PaintingContext.offset).
	// Without this, nested boundaries (ScrollView → items) get ScreenOrigin
	// relative to (0,0) instead of the boundary's actual screen position.
	//
	// ScreenOriginBase must compensate for the PushTransform(-bounds.Min) below.
	// After PushTransform, TransformOffset = -bounds.Min. StampScreenOrigin
	// computes: offset = TransformOffset + ScreenOriginBase = -bounds.Min + base.
	// For a child at childBounds.Min, screenOrigin = offset + childBounds.Min.
	// We want: screenOrigin = bi.ScreenOrigin() + childBounds.Min.
	// So: base = bi.ScreenOrigin() + bounds.Min.
	type screenBaseSetter interface{ SetScreenOriginBase(geometry.Point) }
	if sbs, ok := recorder.(screenBaseSetter); ok {
		sbs.SetScreenOriginBase(bi.ScreenOrigin().Add(bounds.Min))
	}

	// Record in local coordinates.
	recorder.PushTransform(geometry.Pt(-bounds.Min.X, -bounds.Min.Y))
	bi.Draw(ctx, recorder)
	recorder.PopTransform()

	if ds, ok := bi.(dirtySuppressor); ok {
		ds.SetSuppressDirtyCallback(false)
	}

	// Animated widgets (spinner) re-dirty during Draw (SetNeedsRedraw →
	// InvalidateScene), but callback was suppressed. Now suppress is off —
	// if boundary re-dirtied, register it for next frame.
	// NOTE: use RegisterDirtyBoundary (NOT InvalidateRect) to avoid setting
	// window.needsRedraw which forces root re-recording. The boundary is
	// already in the flat dirty set — just needs RequestRedraw to wake loop.
	if bi.IsSceneDirty() && ctx != nil {
		type cacheKeyProvider interface{ BoundaryCacheKey() uint64 }
		if reg, ok := ctx.(widget.DirtyBoundaryRegistrar); ok {
			if ckp, ok2 := bi.(cacheKeyProvider); ok2 {
				reg.RegisterDirtyBoundary(ckp.BoundaryCacheKey())
			}
		}
	}

	cleanup()

	bi.SetCachedScene(cachedScene)
}

// isBoundaryVisible checks whether a boundary widget is inside its compositor
// clip rect (viewport). Boundaries without a clip (root, non-scrolled) are
// always visible. Only boundaries with CompositorClip set by DrawChild during
// parent recording can be culled.
//
// See: ADR-007 Phase 7 (per-boundary GPU textures, offscreen culling)
// Task: TASK-UI-ADR007-PHASE7 (done)
func isBoundaryVisible(bi boundaryInfo) bool {
	type clipChecker interface {
		HasCompositorClip() bool
		CompositorClip() geometry.Rect
	}
	cc, ok := bi.(clipChecker)
	if !ok || !cc.HasCompositorClip() {
		return true
	}
	clip := cc.CompositorClip()
	origin := bi.ScreenOrigin()
	bounds := bi.Bounds()
	screenRect := geometry.Rect{
		Min: origin,
		Max: geometry.Pt(origin.X+bounds.Width(), origin.Y+bounds.Height()),
	}
	return screenRect.Intersects(clip)
}
