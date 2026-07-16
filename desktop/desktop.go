package desktop

import (
	"fmt"
	"image"
	"log"
	"math"
	"os"
	"sync"

	"github.com/gogpu/gg"
	"github.com/gogpu/gg/integration/ggcanvas"
	"github.com/gogpu/gg/scene"
	"github.com/gogpu/gogpu"
	"github.com/gogpu/gpucontext"
	"github.com/gogpu/ui/app"
	"github.com/gogpu/ui/compositor"
	"github.com/gogpu/ui/geometry"
	"github.com/gogpu/ui/render"
)

var (
	debugDamageOnce    sync.Once
	debugDamageEnabled bool
	damageBlitOnce     sync.Once
	damageBlitEnabled  bool
)

func isDebugDamageEnabled() bool {
	debugDamageOnce.Do(func() {
		debugDamageEnabled = os.Getenv("GOGPU_DEBUG_DAMAGE") == "1"
	})
	return debugDamageEnabled
}

func isDamageBlitEnabled() bool {
	damageBlitOnce.Do(func() {
		damageBlitEnabled = os.Getenv("GOGPU_DAMAGE_BLIT") != "0"
	})
	return damageBlitEnabled
}

// Run starts a desktop application with a per-boundary GPU texture render loop.
//
// ADR-007 Phase 7: per-boundary GPU textures with damage-aware blit.
//
// The rendering pipeline:
//  1. Frame() flushes signals, layouts, animations
//  2. PaintBoundaryLayers: re-record dirty+visible boundaries (Flutter flushPaint)
//  3. renderBoundaryTextures: per-boundary offscreen GPU textures (MSAA)
//  4. compositeTextures: blit all textures to surface (non-MSAA)
//  5. DrawOverlays + damage tracking + present
//
// Run blocks until the window is closed.
func Run(gogpuApp *gogpu.App, uiApp *app.App) error {
	if gogpuApp == nil {
		return fmt.Errorf("desktop: gogpuApp must not be nil")
	}
	if uiApp == nil {
		return fmt.Errorf("desktop: uiApp must not be nil")
	}

	// Scene composition: full draw every frame, RepaintBoundary cache for efficiency.
	uiApp.Window().SetRenderMode(app.RenderModeHostManaged)

	rl := &renderLoop{
		gogpuApp: gogpuApp,
		uiApp:    uiApp,
	}

	gogpuApp.OnDraw(rl.draw)

	gogpuApp.OnClose(func() {
		// Teardown widget trees, stop animation pumper (#175).
		uiApp.Window().Close()
		rl.releaseBoundaryTextures()
		gg.CloseAccelerator()
		if rl.canvas != nil {
			_ = rl.canvas.Close()
		}
	})

	return gogpuApp.Run()
}

// renderLoop holds the state for the scene-composition render loop.
//
// ADR-007 Phase 7: per-boundary GPU textures. Each RepaintBoundary owns an
// offscreen GPU texture. Dirty boundaries re-render into their texture.
// Clean boundaries reuse previous texture (0 GPU work). Compositor blits
// all textures via non-MSAA path instead of replaying all scenes through
// MSAA SDF pipeline.
type renderLoop struct {
	gogpuApp     *gogpu.App
	uiApp        *app.App
	canvas       *ggcanvas.Canvas
	debugOverlay dirtyOverlay

	// Per-boundary GPU texture cache. Key = boundary cache key (uint64).
	// Each boundary rendered into its own offscreen texture.
	// Clean boundaries: texture reused. Dirty: re-rendered.
	boundaryTextures map[uint64]*boundaryTexEntry
	fullRedrawNeeded bool // First frame, resize, theme change

	// Damage-aware blit (ADR-030): when only child boundaries changed
	// (root clean), skip root DrawGPUTextureBase and use
	// RenderDirectWithDamageRects with LoadOpLoad + per-draw scissor.
	rootTextureChanged    bool              // root boundary re-rendered this frame
	frameDamageRects      []image.Rectangle // dirty boundary rects (PHYSICAL pixels for GPU scissor)
	boundaryDamageLogical []image.Rectangle // dirty boundary rects (LOGICAL pixels for debug overlay)

	// Ring buffer for N-buffered swapchain damage accumulation (ADR-030).
	// With double/triple/quad buffering, previous buffers need accumulated
	// damage. actualDamage = union of last N frames' damage rects.
	// Size 4 covers Linux 4-buffer swapchains (Wayland triple + mailbox).
	// Multi-rect: each slot stores the full rect list (not a union), enabling
	// per-draw dynamic scissor for distant dirty regions.
	damageRingRects  [4][]image.Rectangle
	damageRingIdx    int
	prevOverlayCount int

	// Persistent layer tree (D5). Survives across frames; UpdateLayerTree
	// reuses PictureLayerImpl/OffsetLayerImpl objects for unchanged boundaries.
	// Nil on first frame or after releaseBoundaryTextures (resize, close).
	layerTree *compositor.OffsetLayerImpl

	// Diagnostic counters (reset each frame, logged with GOGPU_DEBUG_DAMAGE=1).
	frameCounter int // monotonic frame counter for diagnostic logging
	renderCount  int // boundaries rendered (FlushGPUWithView) this frame
	blitCount    int // boundaries blitted (DrawGPUTexture) this frame
}

// boundaryTexEntry holds an offscreen GPU texture for a RepaintBoundary.
type boundaryTexEntry struct {
	texture      gpucontext.TextureView
	release      func()
	width        int
	height       int
	sceneVersion uint64        // tracks which scene version was last rendered into texture
	clipRect     geometry.Rect // screen-space clip for compositor scissoring
	hasClip      bool          // whether clipRect is set
}

// draw is the OnDraw callback registered with gogpu.App.
//
// ADR-007 Phase 7: per-boundary GPU textures with damage-aware blit.
//
// Every frame:
//  1. Frame() flushes signals, layouts, animations
//  2. PaintBoundaryLayers: re-record dirty+visible boundaries (Flutter flushPaint)
//  3. renderBoundaryTextures: per-boundary offscreen GPU textures (MSAA)
//  4. compositeTextures: blit all textures to surface (non-MSAA)
//  5. DrawOverlays + damage tracking + present
func (rl *renderLoop) draw(dc *gogpu.Context) { //nolint:gocyclo,cyclop,gocognit,funlen,maintidx // render loop orchestrates multiple pipeline stages (frame, layout, boundary textures, composite, overlays, present)
	w, h := dc.Width(), dc.Height()
	if w <= 0 || h <= 0 {
		return
	}

	if rl.canvas == nil {
		if !rl.initCanvas(w, h) {
			return
		}
	}

	rl.uiApp.Frame()

	cw, ch := rl.canvas.Size()
	if cw != w || ch != h {
		if err := rl.canvas.Resize(w, h); err != nil {
			log.Printf("desktop: canvas.Resize: %v", err)
		}
		cw, ch = w, h
		rl.releaseBoundaryTextures()
		rl.fullRedrawNeeded = true
	}

	win := rl.uiApp.Window()

	// ADR-028 Phase C: O(1) frame skip using flat dirty boundary list.
	//
	// Frame() already ran (signals, layout, animations). All dirty propagation
	// has populated win.dirtyBoundaries via RegisterDirtyBoundary callback.
	// No O(n) tree walk needed — the flat dirty set is authoritative.
	//
	// Work sources (all O(1)):
	//   - fullRedrawNeeded: resize, first frame, texture release
	//   - win.NeedsRedraw(): layout changed, ctx.Invalidate, signal dirty
	//   - win.HasDirtyBoundaries(): upward propagation → RegisterDirtyBoundary
	//   - win.NeedsAnimationFrame(): spinner ScheduleAnimationFrame
	//
	// Flutter equivalent: _hasScheduledFrame || _nodesNeedingPaint.isNotEmpty
	// Before Phase C: NeedsRedrawInTreeNonBoundary O(n) walked entire tree.
	// After Phase C: HasDirtyBoundaries O(1) checks map length.
	needsAnyWork := rl.fullRedrawNeeded || win.NeedsRedraw() || win.HasDirtyBoundaries() || win.NeedsAnimationFrame()
	if isDebugDirtyEnabled() && rl.debugOverlay.needsAnimationFrame() {
		needsAnyWork = true
	}
	if isDebugDamageEnabled() && rl.canvas != nil && rl.canvas.NeedsAnimationFrame() {
		needsAnyWork = true
	}
	if !needsAnyWork {
		return
	}
	win.ClearAnimationFrame()

	// Per-frame diagnostic counters (GOGPU_DEBUG_DAMAGE=1).
	rl.frameCounter++
	rl.renderCount = 0
	rl.blitCount = 0

	if isDebugDamageEnabled() {
		log.Printf("[FRAME] #%d needsRedraw=%v dirtyBoundaries=%d animFrame=%v fullRedraw=%v",
			rl.frameCounter, win.NeedsRedraw(), win.DirtyBoundaryCount(),
			win.NeedsAnimationFrame(), rl.fullRedrawNeeded)
	}

	cc := rl.canvas.Context()

	gg.BeginAcceleratorFrame()
	cc.BeginGPUFrame()
	cc.ResetFrameDamage()

	// ADR-007 Phase 7: Per-boundary GPU textures.
	//
	// Each RepaintBoundary rendered into its own offscreen GPU texture.
	// Dirty boundaries: re-render scene into texture (MSAA, LoadOpClear).
	// Clean boundaries: reuse previous texture (0 GPU work).
	// Compositor: blit all textures via non-MSAA path.
	//
	// Enterprise references: Flutter RasterCache, Chrome TileManager.
	// Research: docs/dev/research/PER-BOUNDARY-GPU-TEXTURES-RESEARCH.md
	root := win.Root()
	winCtx := win.Context()

	// Root boundary is always at window origin (0,0).
	type originSetter interface{ SetScreenOrigin(geometry.Point) }
	if setter, ok := root.(originSetter); ok {
		setter.SetScreenOrigin(geometry.Point{})
	}

	// Force root re-recording when the window-level NeedsRedraw flag is set.
	// This covers: layout changes, ctx.Invalidate(), signal dirty callbacks,
	// and non-boundary widgets with broken parent chains (no SetParent).
	//
	// ADR-028 Phase C: replaces O(n) NeedsRedrawInTreeNonBoundary tree walk.
	// The window flags (needsRedraw, needsFullRepaint) are set by callbacks
	// in newWindow: onInvalidate, onInvalidateRect, scheduler.SetOnDirty.
	// These are all O(1) flag sets.
	if win.NeedsRedraw() || rl.fullRedrawNeeded { //nolint:nestif // forced root invalidation with callback suppression requires nested type assertions
		if isDebugDamageEnabled() {
			log.Printf("[ROOT-INVALIDATE] frame=%d needsRedraw=%v fullRedraw=%v",
				rl.frameCounter, win.NeedsRedraw(), rl.fullRedrawNeeded)
		}
		type sceneDirtier interface {
			IsRepaintBoundary() bool
			InvalidateScene()
		}
		if sd, ok := root.(sceneDirtier); ok && sd.IsRepaintBoundary() {
			// Suppress onBoundaryDirty callback: we're already inside the
			// render loop — no external notification needed. Without this,
			// InvalidateScene fires ctx.InvalidateRect which restarts the
			// animation pumper at 30fps for data tickers that only need 1fps.
			type dirtySuppressor interface{ SetSuppressDirtyCallback(bool) }
			if ds, ok2 := root.(dirtySuppressor); ok2 {
				ds.SetSuppressDirtyCallback(true)
				sd.InvalidateScene()
				ds.SetSuppressDirtyCallback(false)
			} else {
				sd.InvalidateScene()
			}
		}
	}

	// ADR-028 Phase C: Single-pass dirty collection BEFORE PaintBoundaryLayers.
	// Capture dirty widget rects while NeedsRedraw flags are still true.
	// PaintBoundaryLayers will clear them. Used for:
	//   1. TrackDamageRect (gg debug overlay, GOGPU_DEBUG_DAMAGE=1)
	//   2. SetPresentDamage (partial present to OS compositor)
	//
	// Before Phase C: two passes (pre-paint + post-paint). Post-paint
	// was redundant — it found mostly spinner re-dirty, which boundary
	// damage tracking already covers via boundaryDamageLogical.
	win.CollectDirtyRegions()
	prePaintDirtyRegions := win.DirtyRegions()

	// Paint main tree boundaries.
	app.PaintBoundaryLayersWithContext(root, nil, winCtx)

	// ADR-029 Phase E: Paint overlay content boundaries alongside main tree.
	// Overlay content widgets are already marked as RepaintBoundary by PushOverlay.
	// PaintOverlayBoundaries re-records dirty overlay boundaries so their
	// CachedScene values are fresh for the compositor.
	//
	// Set ScreenOrigin on each overlay content widget BEFORE painting.
	// Overlay content widgets are positioned in window coordinates (Bounds().Min
	// IS the screen origin). Without this, ScreenOrigin stays at (0,0) and the
	// boundary texture blits at the wrong position.
	overlayWidgets := win.OverlayContentWidgets()
	if len(overlayWidgets) > 0 {
		for _, ow := range overlayWidgets {
			type screenOriginSetter interface {
				Bounds() geometry.Rect
				SetScreenOrigin(geometry.Point)
			}
			if sos, ok := ow.(screenOriginSetter); ok {
				sos.SetScreenOrigin(sos.Bounds().Min)
			}
		}
		app.PaintOverlayBoundaries(overlayWidgets, winCtx)
	}

	// ADR-007 Phase D.5: Persistent Layer Tree.
	// UpdateLayerTree reuses PictureLayerImpl/OffsetLayerImpl objects for
	// boundaries that still exist (matched by BoundaryCacheKey), eliminating
	// per-frame layer allocations for stable UIs. First frame (layerTree==nil)
	// builds from scratch; subsequent frames update in place.
	rl.layerTree = app.UpdateLayerTree(root, rl.layerTree)
	layerTree := rl.layerTree

	// ADR-029 Phase E: Append overlay boundaries to Layer Tree.
	// Overlays are appended AFTER main tree children → composite on top
	// (correct Z-order: main content → overlays bottom-to-top).
	if len(overlayWidgets) > 0 {
		app.AppendOverlaysToLayerTree(layerTree, overlayWidgets, rl.layerTree)
	}

	// GC orphaned boundary textures (#116): after SetRoot (theme change),
	// old widget tree keys no longer in Layer Tree. Delete stale entries
	// to prevent texture leaks and flash artifacts during transitions.
	rl.gcOrphanedTextures(layerTree)

	// Render dirty boundaries into offscreen textures (walk Layer Tree).
	// Reset per-frame damage tracking for damage-aware blit (TASK-UI-OPT-003).
	if rl.boundaryTextures == nil {
		rl.boundaryTextures = make(map[uint64]*boundaryTexEntry)
		rl.fullRedrawNeeded = true
	}
	rl.rootTextureChanged = false
	rl.frameDamageRects = rl.frameDamageRects[:0]
	rl.boundaryDamageLogical = rl.boundaryDamageLogical[:0]
	// Suppress damage tracking during offscreen boundary rendering.
	// Fill/Stroke inside RenderScene target offscreen textures, not
	// the surface — they must not pollute gg.FrameDamage().
	cc.SetDamageTracking(false)
	rl.renderBoundaryTexturesFromTree(layerTree, cc)
	cc.SetDamageTracking(true)

	// Compositor: blit all boundary textures onto surface (walk Layer Tree).
	// Overlays are last in the tree → blit on top of main content.
	rl.compositeTexturesFromTree(layerTree, cc, cw, ch)

	// Re-add SURFACE damage so gg debug overlay (GOGPU_DEBUG_DAMAGE=1)
	// shows correct green rects. Two sources:
	// 1. Root widgets (buttons, sliders, chart): from prePaintDirtyRegions
	//    (captured BEFORE PaintBoundaryLayers cleared NeedsRedraw flags)
	// 2. Child boundaries (spinner, overlay content): from boundaryDamageLogical
	// Track surface damage for gg debug overlay (GOGPU_DEBUG_DAMAGE=1).
	if rl.rootTextureChanged {
		for _, r := range prePaintDirtyRegions {
			cc.TrackDamageRect(image.Rect(
				int(r.Min.X), int(r.Min.Y),
				int(r.Max.X+0.5), int(r.Max.Y+0.5),
			))
		}
	}
	for _, dr := range rl.boundaryDamageLogical {
		cc.TrackDamageRect(dr)
	}

	// ADR-029 Phase E: Modal scrim drawing.
	// Overlay CONTENT is now rendered via the boundary pipeline (texture cached).
	// Only the modal backdrop scrim needs immediate-mode drawing.
	// Suppress damage tracking — scrim is full-window and must NOT register.
	overlayCount := win.OverlayCount()
	if win.HasOverlays() {
		widgetCanvas := render.NewCanvas(cc, cw, ch)
		cc.SetDamageTracking(false)
		win.DrawOverlayScrim(widgetCanvas)
		cc.SetDamageTracking(true)
	}

	// Full-window damage on overlay push/pop (content appears/disappears).
	if overlayCount != rl.prevOverlayCount {
		rl.prevOverlayCount = overlayCount
		cc.TrackDamageRect(image.Rect(0, 0, cw, ch))
	}
	win.ClearAfterPaint()
	win.ClearDirtyBoundaries()

	// Debug overlay: cyan flash-and-fade on dirty widget regions (ADR-023).
	// Suppress damage tracking — overlay is visualization, not content.
	if isDebugDirtyEnabled() {
		rl.debugOverlay.update(win.DirtyRegions())
		cc.SetDamageTracking(false)
		rl.debugOverlay.draw(cc, rl.canvas.DeviceScale())
		cc.SetDamageTracking(true)
		if rl.debugOverlay.needsAnimationFrame() {
			if isDebugDamageEnabled() {
				log.Printf("[REDRAW-SRC] ui-dirty-overlay-fade")
			}
			rl.gogpuApp.RequestRedraw()
		}
	}

	// ADR-021 Phase 7: Pass damage rects to gg for partial present.
	// Chain: ui → gg SetPresentDamage → gogpu SetDamageRects → wgpu PresentWithDamage → OS.
	//
	// When root texture changed (full repaint), send full-window damage to the
	// OS compositor. Wayland compositors use wl_surface.damage_buffer as an
	// optimization hint — without full damage on full repaint, vacated areas
	// (e.g. collapsed content) may show stale pixels on the physical display
	// because the compositor doesn't know those pixels changed.
	if rl.rootTextureChanged || rl.fullRedrawNeeded {
		rl.canvas.SetPresentDamage([]image.Rectangle{
			image.Rect(0, 0, cw, ch),
		})
	} else if dirtyRegions := win.DirtyRegions(); len(dirtyRegions) > 0 {
		rects := make([]image.Rectangle, len(dirtyRegions))
		for i, r := range dirtyRegions {
			rects[i] = image.Rect(
				int(r.Min.X), int(r.Min.Y),
				int(r.Max.X+0.5), int(r.Max.Y+0.5),
			)
		}
		rl.canvas.SetPresentDamage(rects)
	}

	// Present via canvas.Render or RenderDirectWithDamage (ADR-022 + TASK-UI-OPT-003).
	// Damage-aware: when root texture unchanged and only child boundaries dirty,
	// use LoadOpLoad + scissor to blit only dirty regions. Previous swapchain
	// content preserved. Fallback to full Render when root changed or overlays present.
	rl.canvas.MarkDirty()

	skipRootBlit := !rl.rootTextureChanged && !rl.fullRedrawNeeded
	hasOverlays := win.HasOverlays()

	// Damage-aware blit: enabled by default (ADR-007 Phase 7, TASK-UI-OPT-003).
	// When root texture unchanged and only child boundaries dirty, use
	// RenderDirectWithDamage (LoadOpLoad + scissor) to render only the
	// damage region. Disable with GOGPU_DAMAGE_BLIT=0 for debugging.
	damageBlitEnabled := isDamageBlitEnabled()
	if isDebugDamageEnabled() {
		log.Printf("[BLIT-PATH] frame=%d damageEnabled=%v skipRoot=%v hasOverlays=%v damageRects=%d rootChanged=%v renderCount=%d blitCount=%d",
			rl.frameCounter, damageBlitEnabled, skipRootBlit, hasOverlays,
			len(rl.frameDamageRects), rl.rootTextureChanged, rl.renderCount, rl.blitCount)
	}
	// Disable damage-aware blit when debug damage overlay is active.
	// RenderDirectWithDamage uses LoadOpLoad which preserves previous swapchain
	// content — including debug overlay pixels. Without LoadOpClear, overlay
	// rects from previous frames are never erased, causing permanent green.
	// Full Render (LoadOpClear) ensures overlay is redrawn fresh each frame.
	if isDebugDamageEnabled() {
		damageBlitEnabled = false
	}
	if damageBlitEnabled && skipRootBlit && !hasOverlays && len(rl.frameDamageRects) > 0 { //nolint:nestif // damage blit feature flag path selection
		// ADR-030: Multi-rect damage-aware path.
		// Accumulate damage across N swapchain buffers (ring buffer).
		// Pass individual rects for per-draw dynamic scissor — zero pixel waste
		// when dirty boundaries are far apart (e.g. spinner + distant button).
		damageRects := rl.accumulatedDamageRects()
		sv := dc.RenderTarget().SurfaceView()
		sw, sh := dc.RenderTarget().SurfaceSize()
		if err := rl.canvas.RenderDirectWithDamageRects(sv, sw, sh, damageRects); err != nil {
			log.Printf("desktop: RenderDirectWithDamageRects: %v", err)
		}
	} else {
		// Full blit path: root changed, overlays present, or first frame.
		if err := rl.canvas.Render(dc.RenderTarget()); err != nil {
			log.Printf("desktop: canvas.Render: %v", err)
		}
		// Fill ALL ring buffer slots with fullWindow so every swapchain
		// buffer (up to 4 on Linux Wayland) knows the entire screen
		// changed. Without this, buffer N-1 from a previous frame has
		// stale content outside damage rect → scroll artifacts (#111).
		if damageBlitEnabled {
			sw, sh := dc.RenderTarget().SurfaceSize()
			fullWindowRect := []image.Rectangle{image.Rect(0, 0, int(sw), int(sh))}
			for i := range rl.damageRingRects {
				rl.damageRingRects[i] = fullWindowRect
			}
			rl.damageRingIdx = 0
		}
	}

	// Request frames for gg damage overlay fade animation.
	// Feedback loop is prevented by SetDamageTracking(false) during overlay
	// draw (canvas.go) and timestamp dedup in damageOverlayState.update().
	if isDebugDamageEnabled() && rl.canvas != nil && rl.canvas.NeedsAnimationFrame() {
		rl.gogpuApp.RequestRedraw()
	}
}

// accumulatedDamageRects returns the accumulated damage rects across the
// current frame and previous frames (ring buffer for N-buffered swapchain).
//
// ADR-030: returns individual rects for per-draw dynamic scissor, enabling
// zero pixel waste when dirty regions are far apart (e.g. spinner 48x48
// at (24,64) + button 100x32 at (300,500) = 5,504 px vs union 175,968 px).
//
// When the total rect count exceeds maxDamageRects (16), falls back to a
// single union rect to avoid GPU scissor overhead (GDK=15, Sway=20).
func (rl *renderLoop) accumulatedDamageRects() []image.Rectangle {
	// Start with current frame's rects.
	rects := make([]image.Rectangle, 0, len(rl.frameDamageRects)+8)
	rects = append(rects, rl.frameDamageRects...)

	// Store current frame rects in ring buffer (copy to avoid aliasing).
	stored := make([]image.Rectangle, len(rl.frameDamageRects))
	copy(stored, rl.frameDamageRects)
	rl.damageRingRects[rl.damageRingIdx] = stored
	rl.damageRingIdx = (rl.damageRingIdx + 1) % len(rl.damageRingRects)

	// Accumulate with previous frames' damage.
	for _, prev := range rl.damageRingRects {
		rects = append(rects, prev...)
	}

	// ADR-030 threshold: merge to single union when too many rects.
	// GPU scissor state changes are cheap but not free. Enterprise
	// compositors cap at similar thresholds (GDK=15, Sway=20).
	const maxDamageRects = 16
	if len(rects) > maxDamageRects {
		var union image.Rectangle
		for _, r := range rects {
			union = union.Union(r)
		}
		return []image.Rectangle{union}
	}

	return rects
}

// renderBoundaryTexturesFromTree walks the Layer Tree and renders dirty
// PictureLayers into their offscreen GPU textures. Clean boundaries keep
// their previous texture (0 GPU work).
//
// ADR-007 Phase D: replaces renderBoundaryTextures widget tree walk.
// The Layer Tree provides structural hierarchy (offsets, clips, opacity)
// without type assertions on widget interfaces.
func (rl *renderLoop) renderBoundaryTexturesFromTree(root compositor.Layer, cc *gg.Context) {
	rl.renderFromTreeRecursive(root, cc)
}

// renderFromTreeRecursive walks the Layer Tree depth-first and renders every
// PictureLayer's scene into its offscreen GPU texture. All nesting depths are
// visited — the compositor blit side (compositeFromTreeRecursive) also walks
// all depths, so the render side must match.
func (rl *renderLoop) renderFromTreeRecursive(layer compositor.Layer, cc *gg.Context) {
	if layer == nil {
		return
	}

	// PictureLayer: render the boundary's scene into its offscreen texture.
	if pic, ok := layer.(*compositor.PictureLayerImpl); ok {
		rl.renderSingleBoundaryFromLayer(pic, cc)
		return
	}

	// ContainerLayer (OffsetLayer, ClipRectLayer, OpacityLayer): recurse
	// into all children unconditionally. Every PictureLayer at any depth
	// must have its offscreen texture rendered.
	container, ok := layer.(compositor.ContainerLayer)
	if !ok {
		return
	}
	for _, child := range container.Children() {
		rl.renderFromTreeRecursive(child, cc)
	}
}

// renderSingleBoundaryFromLayer renders one PictureLayer's scene into its
// offscreen GPU texture. All boundary metadata (cache key, size, screen
// origin, clip, root flag, scene version) is read from the PictureLayerImpl
// fields populated by BuildLayerTree.
//
// ADR-007 Phase D: replaces renderSingleBoundary which used widget interface.
func (rl *renderLoop) renderSingleBoundaryFromLayer(pic *compositor.PictureLayerImpl, cc *gg.Context) {
	bw, bh := pic.Size()
	if bw <= 0 || bh <= 0 {
		return
	}

	// Skip non-visible boundaries (uninitialized origin or outside viewport).
	// Update the entry's clip rect even when skipping, so the blit path uses
	// the current (zero) clip instead of stale data from the last visible frame.
	// Without this, collapsed Collapsible content bleeds through because the
	// compositor blits stale textures with the old viewport clip (#147).
	if !pic.IsRoot() && !isBoundaryLayerVisible(pic, bw, bh) {
		if entry := rl.boundaryTextures[pic.BoundaryCacheKey()]; entry != nil {
			rl.updateClipRect(entry, pic)
		}
		return
	}

	if isDebugDamageEnabled() {
		log.Printf("[RENDER-CHECK] frame=%d key=%d root=%v size=%dx%d dirty=%v originValid=%v",
			rl.frameCounter, pic.BoundaryCacheKey(), pic.IsRoot(), bw, bh,
			pic.IsDirty(), pic.IsScreenOriginValid())
	}

	entry := rl.ensureBoundaryTexture(pic.BoundaryCacheKey(), bw, bh, cc)

	// Detect fresh recordings via scene version. Skip re-rendering clean textures.
	cachedScene := pic.Picture()
	if rl.isBoundaryClean(entry, pic, cachedScene) {
		rl.updateClipRect(entry, pic)
		return
	}
	if cachedScene == nil || cachedScene.IsEmpty() {
		return
	}

	rl.flushBoundaryToTexture(pic, entry, cachedScene, cc, bw, bh)
	rl.renderCount++
	if isDebugDamageEnabled() {
		log.Printf("[RENDER] frame=%d key=%d root=%v size=%dx%d sceneVersion=%d",
			rl.frameCounter, pic.BoundaryCacheKey(), pic.IsRoot(), bw, bh,
			pic.SceneVersion())
	}
	rl.updateClipRect(entry, pic)
	rl.trackBoundaryDamage(pic, bw, bh)
}

// isBoundaryLayerVisible checks whether a non-root PictureLayer should
// be rendered. Returns false for uninitialized origins or viewport-culled.
func isBoundaryLayerVisible(pic *compositor.PictureLayerImpl, bw, bh int) bool {
	if !pic.IsScreenOriginValid() {
		return false
	}
	if !pic.HasPictureClip() {
		return true
	}
	clip := pic.PictureClipRect()
	origin := pic.ScreenOrigin()
	screenRect := geometry.Rect{
		Min: origin,
		Max: geometry.Pt(origin.X+float32(bw), origin.Y+float32(bh)),
	}
	return screenRect.Intersects(clip)
}

// scaleToPhysical converts logical pixel dimensions to physical pixels for the
// given device scale, rounding to the nearest pixel. A scale <= 0 is treated as
// 1.0 (headless / non-HiDPI).
func scaleToPhysical(w, h int, scale float64) (int, int) {
	if scale <= 0 {
		scale = 1.0
	}
	return int(float64(w)*scale + 0.5), int(float64(h)*scale + 0.5)
}

// ensureBoundaryTexture allocates or resizes the offscreen texture for a boundary.
//
// HiDPI: bw/bh are LOGICAL widget bounds, but an offscreen texture is a GPU
// surface that must be allocated in PHYSICAL pixels. The gg context renders at
// DeviceScale, so a logical-sized texture clips the scene to its top-left
// fraction and then gets upscaled at composite time — the "content rendered at
// 2x" symptom on HiDPI/Retina displays (#129). Size to physical = logical *
// DeviceScale; the composite step still positions in logical space (the gg
// context applies the scale), so only allocation and flush need physical extents.
func (rl *renderLoop) ensureBoundaryTexture(key uint64, bw, bh int, cc *gg.Context) *boundaryTexEntry {
	pw, ph := scaleToPhysical(bw, bh, cc.DeviceScale())
	entry := rl.boundaryTextures[key]
	if entry == nil || entry.width != pw || entry.height != ph {
		if entry != nil && entry.release != nil {
			entry.release()
		}
		tex, release := cc.CreateOffscreenTexture(pw, ph)
		entry = &boundaryTexEntry{texture: tex, release: release, width: pw, height: ph}
		rl.boundaryTextures[key] = entry
		rl.fullRedrawNeeded = true
	}
	return entry
}

// isBoundaryClean checks whether a boundary texture is up-to-date (no re-render needed).
func (rl *renderLoop) isBoundaryClean(entry *boundaryTexEntry, pic *compositor.PictureLayerImpl, cachedScene *scene.Scene) bool {
	currentVersion := pic.SceneVersion()
	sceneChanged := entry.sceneVersion != currentVersion
	return !sceneChanged && !pic.IsDirty() && !rl.fullRedrawNeeded && cachedScene != nil
}

// flushBoundaryToTexture renders a boundary's scene into its offscreen GPU texture.
func (rl *renderLoop) flushBoundaryToTexture(pic *compositor.PictureLayerImpl, entry *boundaryTexEntry, cachedScene *scene.Scene, cc *gg.Context, bw, bh int) {
	// Root boundary: draw theme background before scene content.
	if pic.IsRoot() {
		win := rl.uiApp.Window()
		bg := win.ThemeBackground()
		cc.SetRGBA(float64(bg.R), float64(bg.G), float64(bg.B), float64(bg.A))
		cc.DrawRectangle(0, 0, float64(bw), float64(bh))
		_ = cc.Fill()
	}

	renderer := scene.NewGPUSceneRenderer(cc)
	_ = renderer.RenderScene(cachedScene)
	// The flush viewport must match the PHYSICAL texture extent (logical *
	// DeviceScale), else the scene — rendered at DeviceScale — is clipped to the
	// logical-sized region and upscaled on composite (#129). See ensureBoundaryTexture.
	pw, ph := scaleToPhysical(bw, bh, cc.DeviceScale())
	w, h := uint32(max(pw, 0)), uint32(max(ph, 0)) //nolint:gosec // pw/ph derived from bw/bh checked > 0 above
	if err := cc.FlushGPUWithView(entry.texture, w, h); err != nil {
		log.Printf("desktop: FlushGPUWithView boundary %d: %v", pic.BoundaryCacheKey(), err)
	}
	entry.sceneVersion = pic.SceneVersion()
}

// updateClipRect stores the compositor clip rect in the texture entry.
func (rl *renderLoop) updateClipRect(entry *boundaryTexEntry, pic *compositor.PictureLayerImpl) {
	if !pic.IsRoot() && pic.HasPictureClip() {
		entry.clipRect = pic.PictureClipRect()
		entry.hasClip = true
	}
}

// trackBoundaryDamage records damage rects for damage-aware blit (TASK-UI-OPT-003).
func (rl *renderLoop) trackBoundaryDamage(pic *compositor.PictureLayerImpl, bw, bh int) {
	if pic.IsRoot() {
		rl.rootTextureChanged = true
		if isDebugDamageEnabled() {
			log.Printf("[DAMAGE-TRACK] frame=%d source=root key=%d",
				rl.frameCounter, pic.BoundaryCacheKey())
		}
		return
	}
	origin := pic.ScreenOrigin()
	if isDebugDamageEnabled() {
		log.Printf("[DAMAGE-TRACK] frame=%d source=child-boundary key=%d rect=(%d,%d)-(%d,%d)",
			rl.frameCounter, pic.BoundaryCacheKey(),
			int(origin.X), int(origin.Y), int(origin.X)+bw, int(origin.Y)+bh)
	}
	// Logical coords for debug overlay.
	// Use math.Round for origin to match blitPictureLayer (line 806).
	// Without this, int() truncation can make the damage rect 1px smaller
	// than the blit quad, leaving stale arc fragments under LoadOpLoad (#147).
	rx := int(math.Round(float64(origin.X)))
	ry := int(math.Round(float64(origin.Y)))
	rl.boundaryDamageLogical = append(rl.boundaryDamageLogical, image.Rect(
		rx, ry, rx+bw, ry+bh,
	))
	// Physical coords for GPU scissor.
	scale := float64(rl.canvas.DeviceScale())
	rl.frameDamageRects = append(rl.frameDamageRects, image.Rect(
		int(float64(rx)*scale),
		int(float64(ry)*scale),
		int(float64(rx)*scale)+int(float64(bw)*scale+0.5),
		int(float64(ry)*scale)+int(float64(bh)*scale+0.5),
	))
}

// compositeTexturesFromTree walks the Layer Tree and blits all boundary textures
// onto the surface. Root PictureLayer uses DrawGPUTextureBase (background),
// child PictureLayers use DrawGPUTexture (overlays). OpacityLayers apply alpha.
// ClipRectLayers apply viewport clipping.
//
// ADR-007 Phase D: replaces compositeTextures widget tree walk.
func (rl *renderLoop) compositeTexturesFromTree(root compositor.Layer, cc *gg.Context, _, _ int) {
	rl.compositeFromTreeRecursive(root, cc, 1.0)
	rl.fullRedrawNeeded = false
}

func (rl *renderLoop) compositeFromTreeRecursive(layer compositor.Layer, cc *gg.Context, parentOpacity float32) {
	if layer == nil {
		return
	}

	// PictureLayer: blit its texture (skip invisible boundaries).
	if pic, ok := layer.(*compositor.PictureLayerImpl); ok {
		bw, bh := pic.Size()
		if !pic.IsRoot() && !isBoundaryLayerVisible(pic, bw, bh) {
			return
		}
		rl.blitPictureLayer(pic, cc, parentOpacity)
		return
	}

	// OpacityLayer: multiply opacity for children.
	if opLayer, ok := layer.(*compositor.OpacityLayerImpl); ok {
		childOpacity := parentOpacity * opLayer.Opacity()
		for _, child := range opLayer.Children() {
			rl.compositeFromTreeRecursive(child, cc, childOpacity)
		}
		return
	}

	// ClipRectLayer: push clip, recurse, pop.
	if clipLayer, ok := layer.(*compositor.ClipRectLayerImpl); ok {
		clip := clipLayer.ClipRect()
		cc.Push()
		cc.ClipRect(float64(clip.Min.X), float64(clip.Min.Y),
			float64(clip.Width()), float64(clip.Height()))
		for _, child := range clipLayer.Children() {
			rl.compositeFromTreeRecursive(child, cc, parentOpacity)
		}
		cc.Pop()
		return
	}

	// ContainerLayer / OffsetLayer: recurse into children.
	if container, ok := layer.(compositor.ContainerLayer); ok {
		for _, child := range container.Children() {
			rl.compositeFromTreeRecursive(child, cc, parentOpacity)
		}
	}
}

// blitPictureLayer composites a single PictureLayer's texture to the surface.
func (rl *renderLoop) blitPictureLayer(pic *compositor.PictureLayerImpl, cc *gg.Context, opacity float32) {
	key := pic.BoundaryCacheKey()
	entry := rl.boundaryTextures[key]
	if entry == nil || entry.texture.IsNil() {
		return
	}

	bw, bh := pic.Size()
	origin := pic.ScreenOrigin()
	// Snap blit position to integer device pixels (Flutter ComputeIntegralTransCTM
	// pattern). Without rounding, the GPU's bilinear texture sampler interpolates
	// between texel rows at fractional positions, producing inconsistent glyph
	// weights across ListView items at different Y offsets.
	x, y := math.Round(float64(origin.X)), math.Round(float64(origin.Y))

	rl.blitCount++
	if isDebugDamageEnabled() {
		log.Printf("[BLIT] frame=%d key=%d root=%v pos=(%.0f,%.0f) size=%dx%d opacity=%.2f",
			rl.frameCounter, key, pic.IsRoot(), x, y, bw, bh, opacity)
	}

	switch {
	case pic.IsRoot():
		cc.DrawGPUTextureBase(entry.texture, x, y, bw, bh)

	case opacity < 1.0:
		// OpacityLayer parent: blit with alpha blending.
		if entry.hasClip {
			clip := entry.clipRect
			cc.Push()
			cc.ClipRect(math.Round(float64(clip.Min.X)), math.Round(float64(clip.Min.Y)),
				math.Round(float64(clip.Width())), math.Round(float64(clip.Height())))
			cc.DrawGPUTextureWithOpacity(entry.texture, x, y, bw, bh, opacity)
			cc.Pop()
		} else {
			cc.DrawGPUTextureWithOpacity(entry.texture, x, y, bw, bh, opacity)
		}

	case entry.hasClip:
		clip := entry.clipRect
		cc.Push()
		cc.ClipRect(math.Round(float64(clip.Min.X)), math.Round(float64(clip.Min.Y)),
			math.Round(float64(clip.Width())), math.Round(float64(clip.Height())))
		cc.DrawGPUTexture(entry.texture, x, y, bw, bh)
		cc.Pop()

	default:
		cc.DrawGPUTexture(entry.texture, x, y, bw, bh)
	}
}

// gcOrphanedTextures removes boundary texture entries whose keys are no
// longer present in the Layer Tree. After SetRoot (e.g., theme change),
// old widget tree keys become orphaned — new widgets have new keys.
// Without GC, stale textures leak and can flash during transitions (#116).
func (rl *renderLoop) gcOrphanedTextures(tree compositor.Layer) {
	if len(rl.boundaryTextures) == 0 {
		return
	}
	liveKeys := make(map[uint64]bool, len(rl.boundaryTextures))
	collectLiveKeys(tree, liveKeys)
	for key, entry := range rl.boundaryTextures {
		if !liveKeys[key] {
			if entry.release != nil {
				entry.release()
			}
			delete(rl.boundaryTextures, key)
		}
	}
}

// collectLiveKeys walks the Layer Tree and records all BoundaryCacheKeys
// from PictureLayerImpl nodes. Used by gcOrphanedTextures to identify
// which texture entries are still referenced by the current widget tree.
func collectLiveKeys(layer compositor.Layer, keys map[uint64]bool) {
	if layer == nil {
		return
	}
	if pic, ok := layer.(*compositor.PictureLayerImpl); ok {
		keys[pic.BoundaryCacheKey()] = true
		return
	}
	if cl, ok := layer.(compositor.ContainerLayer); ok {
		for _, child := range cl.Children() {
			collectLiveKeys(child, keys)
		}
	}
}

// releaseBoundaryTextures frees all offscreen GPU textures.
func (rl *renderLoop) releaseBoundaryTextures() {
	for _, entry := range rl.boundaryTextures {
		if entry.release != nil {
			entry.release()
		}
	}
	rl.boundaryTextures = nil
	rl.layerTree = nil // Force fresh build on next frame.
}

// initCanvas creates the ggcanvas lazily on the first draw call.
func (rl *renderLoop) initCanvas(w, h int) bool {
	provider := rl.gogpuApp.GPUContextProvider()
	if provider == nil {
		return false
	}
	var err error
	rl.canvas, err = ggcanvas.New(provider, w, h)
	if err != nil {
		log.Printf("desktop: ggcanvas.New: %v", err)
		return false
	}
	// LCD subpixel layout is auto-detected by ggcanvas.New() via
	// PlatformProvider.SubpixelLayout() (ADR-024). The gpuContextAdapter
	// delegates PlatformProvider to App since gogpu BUG-ADAPTER-001 fix.
	return true
}
