package app

import (
	"testing"

	"github.com/gogpu/gg/scene"
	"github.com/gogpu/ui/compositor"
	"github.com/gogpu/ui/geometry"
	internalRender "github.com/gogpu/ui/internal/render"
	"github.com/gogpu/ui/widget"
)

// --- Overlay Boundary Damage Tracking Tests ---
//
// These tests verify the app-layer side of the overlay damage tracking pipeline:
//   - sceneDirty flag set by hover (InvalidateScene via SetNeedsRedraw)
//   - sceneCacheVersion increment after recording
//   - syncPictureLayer correctly copying version even when dirty=false
//   - onBoundaryDirty callback wiring for overlay content
//   - full pipeline: hover → dirty → record → version bump → PictureLayer sync
//
// The render-layer tests (isBoundaryClean, trackBoundaryDamage) are in
// desktop/overlay_damage_render_test.go since they use desktop-internal types.
//
// Root cause hypothesis: recordBoundary clears sceneDirty BEFORE syncPictureLayer
// runs. syncPictureLayer reads IsSceneDirty()=false → ClearDirty on PictureLayer.
// BUT SceneCacheVersion incremented → isBoundaryClean detects version mismatch.
// If detection works → tests pass → bug is elsewhere.
// If detection fails → tests fail → found root cause.

// TestOverlayBoundary_SceneDirtyAfterHover verifies that when an overlay menu
// boundary receives a hover event (SetNeedsRedraw), InvalidateScene is called
// and sceneDirty becomes true, and SceneCacheVersion increments after re-record.
func TestOverlayBoundary_SceneDirtyAfterHover(t *testing.T) {
	menu := newOverlayContent(100, 200, 200, 150)
	menu.SetRepaintBoundary(true)
	menu.SetScreenOrigin(geometry.Pt(100, 200))

	// Initial state: sceneDirty=true (SetRepaintBoundary sets it).
	if !menu.IsSceneDirty() {
		t.Fatal("menu should be sceneDirty after SetRepaintBoundary(true)")
	}

	// Record initial scene to clear dirty state.
	prev := widget.GetSceneRecorderFactory()
	widget.RegisterSceneRecorder(func(s widget.SceneCache, w, h int) (widget.Canvas, func()) {
		rec := internalRender.NewSceneCanvas(s.(*scene.Scene), w, h)
		return rec, rec.Close
	})
	defer widget.RegisterSceneRecorder(prev)

	ctx := widget.NewContext()
	ctx.SetOnInvalidateRect(func(_ geometry.Rect) {})
	PaintBoundaryLayersWithContext(menu, nil, ctx)

	// After recording: sceneDirty should be false, version should be 1.
	if menu.IsSceneDirty() {
		t.Error("menu should not be sceneDirty after initial recording")
	}
	v1 := menu.SceneCacheVersion()
	if v1 == 0 {
		t.Fatal("SceneCacheVersion should be > 0 after recording (ClearSceneDirty increments)")
	}

	// Simulate hover: mark menu dirty.
	menu.SetNeedsRedraw(true)

	// SetNeedsRedraw on a RepaintBoundary triggers InvalidateScene via
	// propagateDirtyUpward (since the menu IS its own boundary).
	if !menu.IsSceneDirty() {
		t.Error("menu should be sceneDirty after SetNeedsRedraw(true) — " +
			"InvalidateScene was not called. Check propagateDirtyUpward path for " +
			"standalone overlay boundaries (no parent)")
	}

	// Re-record (simulates PaintOverlayBoundaries on next frame).
	PaintBoundaryLayersWithContext(menu, nil, ctx)

	v2 := menu.SceneCacheVersion()
	if v2 <= v1 {
		t.Errorf("SceneCacheVersion should increment after re-record: v1=%d, v2=%d", v1, v2)
	}
	if menu.IsSceneDirty() {
		t.Error("menu should be clean after re-recording")
	}
}

// TestOverlayBoundary_RecordClearsButVersionIncrements verifies the critical
// invariant: after recordBoundary, sceneDirty is false BUT sceneCacheVersion
// is incremented. This version change is the signal for isBoundaryClean.
func TestOverlayBoundary_RecordClearsButVersionIncrements(t *testing.T) {
	menu := newOverlayContent(100, 200, 200, 150)
	menu.SetRepaintBoundary(true)
	menu.SetScreenOrigin(geometry.Pt(100, 200))

	prev := widget.GetSceneRecorderFactory()
	widget.RegisterSceneRecorder(func(s widget.SceneCache, w, h int) (widget.Canvas, func()) {
		rec := internalRender.NewSceneCanvas(s.(*scene.Scene), w, h)
		return rec, rec.Close
	})
	defer widget.RegisterSceneRecorder(prev)

	ctx := widget.NewContext()
	ctx.SetOnInvalidateRect(func(_ geometry.Rect) {})

	// Record 3 times, tracking version each time.
	versions := make([]uint64, 0, 3)
	for i := range 3 {
		menu.SetNeedsRedraw(true) // re-dirty
		if !menu.IsSceneDirty() && i > 0 {
			t.Errorf("iteration %d: menu should be sceneDirty after SetNeedsRedraw", i)
		}

		PaintBoundaryLayersWithContext(menu, nil, ctx)

		// AFTER record: dirty cleared, version incremented.
		if menu.IsSceneDirty() {
			t.Errorf("iteration %d: menu should NOT be sceneDirty after record", i)
		}

		v := menu.SceneCacheVersion()
		versions = append(versions, v)
	}

	// Versions must be strictly monotonically increasing.
	for i := 1; i < len(versions); i++ {
		if versions[i] <= versions[i-1] {
			t.Errorf("version[%d]=%d should be > version[%d]=%d (monotonic increment)",
				i, versions[i], i-1, versions[i-1])
		}
	}
}

// TestOverlayBoundary_SyncPictureLayerDetectsVersionChange verifies that
// syncPictureLayer correctly copies the new SceneCacheVersion from the
// widget to the PictureLayer, even when sceneDirty is false.
func TestOverlayBoundary_SyncPictureLayerDetectsVersionChange(t *testing.T) {
	menu := newOverlayContent(100, 200, 200, 150)
	menu.SetRepaintBoundary(true)
	menu.SetScreenOrigin(geometry.Pt(100, 200))

	prev := widget.GetSceneRecorderFactory()
	widget.RegisterSceneRecorder(func(s widget.SceneCache, w, h int) (widget.Canvas, func()) {
		rec := internalRender.NewSceneCanvas(s.(*scene.Scene), w, h)
		return rec, rec.Close
	})
	defer widget.RegisterSceneRecorder(prev)

	ctx := widget.NewContext()
	ctx.SetOnInvalidateRect(func(_ geometry.Rect) {})

	// Step 1: Initial recording — version becomes V1.
	PaintBoundaryLayersWithContext(menu, nil, ctx)
	v1 := menu.SceneCacheVersion()

	// Build initial PictureLayer with version V1.
	pic := compositor.NewPictureLayer()
	syncPictureLayer(pic, menu, menu)
	if pic.SceneVersion() != v1 {
		t.Errorf("initial syncPictureLayer: pic.SceneVersion=%d, want %d", pic.SceneVersion(), v1)
	}

	// Step 2: Simulate hover → re-record → version becomes V2.
	menu.SetNeedsRedraw(true)
	PaintBoundaryLayersWithContext(menu, nil, ctx)
	v2 := menu.SceneCacheVersion()
	if v2 <= v1 {
		t.Fatalf("version should increment: v1=%d, v2=%d", v1, v2)
	}

	// At this point: menu.IsSceneDirty() == false (cleared by record).
	// The PictureLayer still has the old sceneVersion from step 1.
	if pic.SceneVersion() != v1 {
		t.Errorf("before re-sync: pic.SceneVersion=%d, want %d (stale)", pic.SceneVersion(), v1)
	}

	// Step 3: Re-sync PictureLayer. This is what UpdateLayerTree does.
	syncPictureLayer(pic, menu, menu)

	// Key assertion: PictureLayer must have the NEW version V2.
	if pic.SceneVersion() != v2 {
		t.Errorf("after re-sync: pic.SceneVersion=%d, want %d — "+
			"syncPictureLayer did not copy updated SceneCacheVersion", pic.SceneVersion(), v2)
	}

	// syncPictureLayer reads IsSceneDirty() which is false → ClearDirty.
	// This is correct — dirty detection in render is via VERSION, not dirty flag.
	if pic.IsDirty() {
		t.Error("pic.IsDirty() should be false — sceneDirty was cleared by recordBoundary")
	}
}

// TestOverlayBoundary_FullPipeline_HoverGeneratesDamage simulates the complete
// overlay damage pipeline across two frames:
//
//	Frame 1: overlay pushed → boundary dirty → record → syncPictureLayer (V1)
//	Frame 2: hover event → SetNeedsRedraw → re-record → version V2 → PictureLayer sync → version mismatch
//
// This test catches the bug where hover generates no damage because
// sceneDirty is cleared before syncPictureLayer but version detection
// should still trigger re-render.
func TestOverlayBoundary_FullPipeline_HoverGeneratesDamage(t *testing.T) {
	prev := widget.GetSceneRecorderFactory()
	widget.RegisterSceneRecorder(func(s widget.SceneCache, w, h int) (widget.Canvas, func()) {
		rec := internalRender.NewSceneCanvas(s.(*scene.Scene), w, h)
		return rec, rec.Close
	})
	defer widget.RegisterSceneRecorder(prev)

	// Setup: root boundary + overlay menu boundary.
	uiApp := New()
	win := uiApp.Window()

	root := newOverlayRoot(geometry.Sz(800, 600))
	root.SetRepaintBoundary(true)
	root.SetScreenOrigin(geometry.Pt(0, 0))
	win.SetRoot(root)

	// Push overlay menu via windowOverlayManager (production path).
	// windowOverlayManager.PushOverlay sets SetRepaintBoundary(true) on content
	// before wrapping in Container — this is the ADR-024+ADR-029 pattern.
	// Raw win.Overlays().Push() does NOT set RepaintBoundary.
	menu := newOverlayContent(100, 200, 200, 150)
	mgr := &windowOverlayManager{window: win}
	mgr.PushOverlay(menu, nil)

	// PushOverlay should mark menu as RepaintBoundary.
	if !menu.IsRepaintBoundary() {
		t.Fatal("overlay content should be marked RepaintBoundary after PushOverlay")
	}

	winCtx := win.Context()
	overlayWidgets := win.OverlayContentWidgets()
	if len(overlayWidgets) != 1 {
		t.Fatalf("OverlayContentWidgets = %d, want 1", len(overlayWidgets))
	}

	// --- Frame 1: Initial recording ---

	// Record main tree + overlay boundaries.
	PaintBoundaryLayersWithContext(root, nil, winCtx)
	PaintOverlayBoundaries(overlayWidgets, winCtx)

	v1 := menu.SceneCacheVersion()
	if v1 == 0 {
		t.Fatal("menu SceneCacheVersion should be > 0 after initial record")
	}

	// Build Layer Tree with overlay appended.
	tree := UpdateLayerTree(root, nil)
	AppendOverlaysToLayerTree(tree, overlayWidgets, nil)

	// Verify overlay PictureLayer exists in tree.
	var pics []*compositor.PictureLayerImpl
	collectPictureLayersFromTree(tree, &pics)

	menuKey := menu.BoundaryCacheKey()
	var menuPic *compositor.PictureLayerImpl
	for _, pic := range pics {
		if pic.BoundaryCacheKey() == menuKey {
			menuPic = pic
			break
		}
	}
	if menuPic == nil {
		t.Fatal("menu boundary PictureLayer not found in Layer Tree")
	}

	// Verify PictureLayer has correct sceneVersion.
	if menuPic.SceneVersion() != v1 {
		t.Errorf("frame 1: menuPic.SceneVersion=%d, want %d", menuPic.SceneVersion(), v1)
	}

	// Remember entry version (simulates what renderLoop would store).
	entryVersion := v1

	// --- Frame 2: Hover event ---

	// Simulate hover on menu item.
	menu.SetNeedsRedraw(true)

	// Check: Is menu boundary dirty?
	if !menu.IsSceneDirty() {
		t.Error("frame 2: menu should be sceneDirty after SetNeedsRedraw — " +
			"this is the core bug: hover did not trigger InvalidateScene")
	}

	// Re-record overlay boundaries (PaintOverlayBoundaries on next frame).
	PaintOverlayBoundaries(overlayWidgets, winCtx)

	v2 := menu.SceneCacheVersion()
	if v2 <= v1 {
		t.Errorf("frame 2: SceneCacheVersion should increment: v1=%d, v2=%d", v1, v2)
	}

	// After recording: sceneDirty is false.
	if menu.IsSceneDirty() {
		t.Error("frame 2: menu should be clean after recording")
	}

	// Update Layer Tree (simulates UpdateLayerTree + AppendOverlaysToLayerTree).
	tree2 := UpdateLayerTree(root, tree)
	AppendOverlaysToLayerTree(tree2, overlayWidgets, tree)

	// Find menu PictureLayer in updated tree.
	var pics2 []*compositor.PictureLayerImpl
	collectPictureLayersFromTree(tree2, &pics2)

	var menuPic2 *compositor.PictureLayerImpl
	for _, pic := range pics2 {
		if pic.BoundaryCacheKey() == menuKey {
			menuPic2 = pic
			break
		}
	}
	if menuPic2 == nil {
		t.Fatal("frame 2: menu PictureLayer not found in updated Layer Tree")
	}

	// KEY ASSERTION: PictureLayer must have the NEW version V2.
	// This is what syncPictureLayer copies during UpdateLayerTree.
	if menuPic2.SceneVersion() != v2 {
		t.Errorf("frame 2: menuPic.SceneVersion=%d, want %d — "+
			"syncPictureLayer did not copy new version", menuPic2.SceneVersion(), v2)
	}

	// Simulate what isBoundaryClean would check:
	// entry.sceneVersion (V1) != pic.SceneVersion (V2) → NOT clean → render happens.
	versionMismatch := entryVersion != menuPic2.SceneVersion()
	if !versionMismatch {
		t.Errorf("frame 2: entry.sceneVersion=%d should differ from pic.SceneVersion=%d — "+
			"without this mismatch, render is skipped and no damage tracked. "+
			"BUG: version mismatch detection broken", entryVersion, menuPic2.SceneVersion())
	}
}

// TestOverlayBoundary_CleanHover_NoRender verifies that when no hover event
// occurs (boundary is clean), PaintOverlayBoundaries does NOT re-record,
// version does not change, and the PictureLayer stays clean.
func TestOverlayBoundary_CleanHover_NoRender(t *testing.T) {
	prev := widget.GetSceneRecorderFactory()
	widget.RegisterSceneRecorder(func(s widget.SceneCache, w, h int) (widget.Canvas, func()) {
		rec := internalRender.NewSceneCanvas(s.(*scene.Scene), w, h)
		return rec, rec.Close
	})
	defer widget.RegisterSceneRecorder(prev)

	menu := newOverlayContent(100, 200, 200, 150)
	menu.SetRepaintBoundary(true)
	menu.SetScreenOrigin(geometry.Pt(100, 200))

	ctx := widget.NewContext()
	ctx.SetOnInvalidateRect(func(_ geometry.Rect) {})

	// Initial recording.
	PaintOverlayBoundaries([]widget.Widget{menu}, ctx)
	v1 := menu.SceneCacheVersion()

	// Build PictureLayer with matching version.
	pic := compositor.NewPictureLayer()
	syncPictureLayer(pic, menu, menu)

	// Frame 2: NO hover. Call PaintOverlayBoundaries again.
	PaintOverlayBoundaries([]widget.Widget{menu}, ctx)
	v2 := menu.SceneCacheVersion()

	// Version should NOT change (boundary was clean, no re-record).
	if v2 != v1 {
		t.Errorf("clean frame: version should NOT change: v1=%d, v2=%d", v1, v2)
	}

	// Re-sync PictureLayer.
	syncPictureLayer(pic, menu, menu)

	// PictureLayer version should still match entry.
	if pic.SceneVersion() != v1 {
		t.Errorf("clean frame: pic.SceneVersion=%d, want %d (unchanged)", pic.SceneVersion(), v1)
	}

	// PictureLayer should not be dirty.
	if pic.IsDirty() {
		t.Error("clean frame: pic.IsDirty() should be false — no re-record happened")
	}
}

// TestOverlayBoundary_StandalonePropagation verifies that SetNeedsRedraw on a
// standalone overlay widget (no parent in main tree) correctly calls
// InvalidateScene on itself. Overlay content widgets have Parent()==nil
// because they are not part of the main widget tree.
//
// This tests a potential root cause: propagateDirtyUpward might not reach
// the boundary if the widget IS its own boundary and has no parent.
func TestOverlayBoundary_StandalonePropagation(t *testing.T) {
	menu := newOverlayContent(100, 200, 200, 150)
	menu.SetRepaintBoundary(true)
	menu.SetScreenOrigin(geometry.Pt(100, 200))

	prev := widget.GetSceneRecorderFactory()
	widget.RegisterSceneRecorder(func(s widget.SceneCache, w, h int) (widget.Canvas, func()) {
		rec := internalRender.NewSceneCanvas(s.(*scene.Scene), w, h)
		return rec, rec.Close
	})
	defer widget.RegisterSceneRecorder(prev)

	ctx := widget.NewContext()
	ctx.SetOnInvalidateRect(func(_ geometry.Rect) {})

	// Initial recording to clear dirty.
	PaintBoundaryLayersWithContext(menu, nil, ctx)

	if menu.IsSceneDirty() {
		t.Fatal("menu should be clean after initial recording")
	}

	// Simulate hover: SetNeedsRedraw on the boundary widget itself.
	menu.SetNeedsRedraw(true)

	// The widget IS a RepaintBoundary. SetNeedsRedraw should trigger
	// InvalidateScene on itself via propagateDirtyUpward.
	if !menu.IsSceneDirty() {
		t.Error("standalone boundary should have sceneDirty=true after SetNeedsRedraw — " +
			"propagateDirtyUpward should call InvalidateScene on self when " +
			"the widget itself is a RepaintBoundary. " +
			"BUG: orphan overlay boundaries never get sceneDirty from hover")
	}
}

// TestOverlayBoundary_OnBoundaryDirtyCallback verifies that the
// onBoundaryDirty callback is wired correctly for overlay boundaries
// and fires when the boundary transitions from clean to dirty.
func TestOverlayBoundary_OnBoundaryDirtyCallback(t *testing.T) {
	menu := newOverlayContent(100, 200, 200, 150)
	menu.SetRepaintBoundary(true)
	menu.SetScreenOrigin(geometry.Pt(100, 200))

	prev := widget.GetSceneRecorderFactory()
	widget.RegisterSceneRecorder(func(s widget.SceneCache, w, h int) (widget.Canvas, func()) {
		rec := internalRender.NewSceneCanvas(s.(*scene.Scene), w, h)
		return rec, rec.Close
	})
	defer widget.RegisterSceneRecorder(prev)

	registeredKey := uint64(0)
	ctx := widget.NewContext()
	ctx.SetOnInvalidateRect(func(_ geometry.Rect) {})
	// Wire RegisterDirtyBoundary — this is what recordBoundary's onBoundaryDirty
	// callback calls (via ctx.(DirtyBoundaryRegistrar).RegisterDirtyBoundary).
	ctx.SetOnRegisterDirtyBoundary(func(key uint64) {
		registeredKey = key
	})

	// Record boundary — this wires onBoundaryDirty inside recordBoundary.
	PaintBoundaryLayersWithContext(menu, nil, ctx)

	// Verify clean state.
	if menu.IsSceneDirty() {
		t.Fatal("menu should be clean after recording")
	}
	registeredKey = 0

	// Trigger hover: SetNeedsRedraw → InvalidateScene → onBoundaryDirty
	// → RegisterDirtyBoundary(key).
	menu.SetNeedsRedraw(true)

	// The callback should fire for non-suppressed InvalidateScene.
	// recordBoundary wires onBoundaryDirty to call RegisterDirtyBoundary,
	// not InvalidateRect. Without this wiring, the render loop never wakes.
	if registeredKey == 0 {
		t.Error("onBoundaryDirty callback should fire and RegisterDirtyBoundary " +
			"should be called when overlay boundary transitions from clean to dirty " +
			"via SetNeedsRedraw → InvalidateScene. Without this, the render loop " +
			"never wakes for hover updates")
	}
	if registeredKey != 0 && registeredKey != menu.BoundaryCacheKey() {
		t.Errorf("RegisterDirtyBoundary called with key=%d, want menu boundary key=%d",
			registeredKey, menu.BoundaryCacheKey())
	}
}

// TestOverlayBoundary_OverlayNotRoot verifies that overlay PictureLayers have
// IsRoot=false after AppendOverlaysToLayerTree. This is critical because
// trackBoundaryDamage uses IsRoot to decide whether to set rootTextureChanged
// (root) or append to frameDamageRects (child). Overlay damage must go to
// frameDamageRects for correct scissor targeting.
func TestOverlayBoundary_OverlayNotRoot(t *testing.T) {
	uiApp := New()
	win := uiApp.Window()

	root := newOverlayRoot(geometry.Sz(800, 600))
	root.SetRepaintBoundary(true)
	root.SetScreenOrigin(geometry.Pt(0, 0))
	win.SetRoot(root)

	// Push via windowOverlayManager (production path) so content is
	// promoted to RepaintBoundary with a valid BoundaryCacheKey.
	menu := newOverlayContent(100, 200, 200, 150)
	mgr := &windowOverlayManager{window: win}
	mgr.PushOverlay(menu, nil)

	tree := BuildLayerTree(root)
	overlayWidgets := win.OverlayContentWidgets()
	AppendOverlaysToLayerTree(tree, overlayWidgets, nil)

	var pics []*compositor.PictureLayerImpl
	collectPictureLayersFromTree(tree, &pics)

	menuKey := menu.BoundaryCacheKey()
	for _, pic := range pics {
		if pic.BoundaryCacheKey() == menuKey {
			if pic.IsRoot() {
				t.Error("overlay PictureLayer should NOT be root — " +
					"clearRootOnPictureLayers not called or failed. " +
					"This causes trackBoundaryDamage to set rootTextureChanged " +
					"instead of appending to frameDamageRects")
			}
			return
		}
	}
	t.Fatal("overlay PictureLayer not found in tree")
}
