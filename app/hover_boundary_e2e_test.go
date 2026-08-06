package app

import (
	"testing"

	"github.com/gogpu/gg/scene"
	"github.com/gogpu/ui/core/button"
	"github.com/gogpu/ui/event"
	"github.com/gogpu/ui/geometry"
	internalRender "github.com/gogpu/ui/internal/render"
	"github.com/gogpu/ui/widget"
)

// TestHoverE2E_ButtonInBoundary_DirtyPropagation verifies the full hover chain:
//
//	MouseMove → hitTest → MouseEnter → Button.SetNeedsRedraw(true)
//	→ propagateDirtyUpward → root boundary InvalidateScene → sceneDirty=true
//	→ onBoundaryDirty callback → ctx.InvalidateRect → Window.needsRedraw
//	→ PaintBoundaryLayers re-records root scene with hover state
//
// This is the critical chain that must work for hover effects to be visible
// after the depth>0 boundary change that renders child boundaries inline.
func TestHoverE2E_ButtonInBoundary_DirtyPropagation(t *testing.T) {
	// Register SceneRecorder factory for boundary recording.
	prev := widget.GetSceneRecorderFactory()
	widget.RegisterSceneRecorder(func(s widget.SceneCache, w, h int) (widget.Canvas, func()) {
		rec := internalRender.NewSceneCanvas(s.(*scene.Scene), w, h)
		return rec, rec.Close
	})
	defer widget.RegisterSceneRecorder(prev)

	// Build widget tree: root (boundary) → container → button
	root := &boxContainer{}
	root.SetVisible(true)
	root.SetEnabled(true)
	root.SetRepaintBoundary(true)
	root.SetBounds(geometry.NewRect(0, 0, 800, 600))
	root.SetScreenOrigin(geometry.Pt(0, 0))

	btn := button.New(button.Text("Hover Me"))
	btn.SetBounds(geometry.NewRect(50, 50, 200, 90))
	btn.SetScreenOrigin(geometry.Pt(50, 50))
	root.kids = []widget.Widget{btn}

	// Mount tree to wire parent chain (critical for propagateDirtyUpward).
	invalidateRectCalled := false
	ctx := widget.NewContext()
	ctx.SetOnInvalidateRect(func(_ geometry.Rect) {
		invalidateRectCalled = true
	})
	widget.MountTree(root, ctx)

	// Step 1: Initial recording (first frame).
	PaintBoundaryLayersWithContext(root, nil, ctx)

	if root.CachedScene() == nil {
		t.Fatal("root CachedScene should be non-nil after initial paint")
	}
	initialVersion := root.SceneCacheVersion()
	t.Logf("initial: sceneDirty=%v, version=%d", root.IsSceneDirty(), initialVersion)

	// Step 2: Verify button's parent chain is wired.
	if btn.Parent() == nil {
		t.Fatal("button.Parent() is nil — MountTree did not wire parent chain")
	}
	if btn.Parent() != root {
		t.Errorf("button.Parent() = %T, want root container", btn.Parent())
	}

	// Step 3: Verify root boundary has onBoundaryDirty callback.
	// After recordBoundary, the callback should be set.
	// We can verify indirectly: clear root dirty, then trigger propagation.
	if root.IsSceneDirty() {
		t.Log("root is still dirty after PaintBoundaryLayers — unexpected")
	}

	// Step 4: Simulate MouseEnter on button.
	enterEvt := event.NewMouseEvent(
		event.MouseEnter, event.ButtonNone, 0,
		geometry.Pt(100, 70), geometry.Pt(100, 70), event.ModNone,
	)
	consumed := btn.Event(ctx, enterEvt)
	if !consumed {
		t.Fatal("button should consume MouseEnter event")
	}

	// Step 5: Verify button state changed to hover.
	if !btn.NeedsRedraw() {
		t.Error("button should have needsRedraw=true after MouseEnter")
	}

	// Step 6: Verify dirty propagated to root boundary.
	if !root.IsSceneDirty() {
		t.Error("root boundary sceneDirty should be true after button hover — " +
			"propagateDirtyUpward did not reach root boundary. " +
			"Check: 1) button.Parent() wired, 2) root.IsRepaintBoundary(), " +
			"3) InvalidateScene() called")
	}

	// Step 7: Verify onBoundaryDirty callback fired.
	if !invalidateRectCalled {
		t.Error("onBoundaryDirty callback should have called ctx.InvalidateRect — " +
			"callback not wired or not fired. Check recordBoundary SetOnBoundaryDirty")
	}

	// Step 8: Re-record (simulates next frame's PaintBoundaryLayers).
	invalidateRectCalled = false
	PaintBoundaryLayersWithContext(root, nil, ctx)

	newVersion := root.SceneCacheVersion()
	if newVersion <= initialVersion {
		t.Errorf("SceneCacheVersion should increment after re-recording: "+
			"initial=%d, after=%d", initialVersion, newVersion)
	}
	t.Logf("after hover re-record: version=%d, sceneDirty=%v", newVersion, root.IsSceneDirty())

	// Step 9: Verify scene is clean after re-recording (ready for next frame).
	if root.IsSceneDirty() {
		t.Error("root should be clean after PaintBoundaryLayers re-recorded it")
	}
}

// TestHoverE2E_DeepNesting_PropagatesUpward verifies dirty propagation through
// multiple levels of nesting. Button inside Box inside Box inside root boundary.
func TestHoverE2E_DeepNesting_PropagatesUpward(t *testing.T) {
	prev := widget.GetSceneRecorderFactory()
	widget.RegisterSceneRecorder(func(s widget.SceneCache, w, h int) (widget.Canvas, func()) {
		rec := internalRender.NewSceneCanvas(s.(*scene.Scene), w, h)
		return rec, rec.Close
	})
	defer widget.RegisterSceneRecorder(prev)

	// 4-level tree: root (boundary) → mid → inner → button
	root := &boxContainer{}
	root.SetVisible(true)
	root.SetEnabled(true)
	root.SetRepaintBoundary(true)
	root.SetBounds(geometry.NewRect(0, 0, 800, 600))

	mid := &boxContainer{}
	mid.SetVisible(true)
	mid.SetEnabled(true)
	mid.SetBounds(geometry.NewRect(0, 0, 800, 600))

	inner := &boxContainer{}
	inner.SetVisible(true)
	inner.SetEnabled(true)
	inner.SetBounds(geometry.NewRect(10, 10, 300, 200))

	btn := button.New(button.Text("Deep Button"))
	btn.SetBounds(geometry.NewRect(20, 20, 150, 60))
	btn.SetScreenOrigin(geometry.Pt(30, 30))

	inner.kids = []widget.Widget{btn}
	mid.kids = []widget.Widget{inner}
	root.kids = []widget.Widget{mid}

	callbackCount := 0
	ctx := widget.NewContext()
	ctx.SetOnInvalidateRect(func(_ geometry.Rect) {
		callbackCount++
	})
	widget.MountTree(root, ctx)

	// Initial paint to wire callbacks.
	PaintBoundaryLayersWithContext(root, nil, ctx)

	// Verify parent chain: button → inner → mid → root
	p := btn.Parent()
	if p == nil {
		t.Fatal("button has no parent")
	}
	if p != inner {
		t.Errorf("button.Parent() = %T(%p), want inner(%p)", p, p, inner)
	}

	p2 := inner.Parent()
	if p2 != mid {
		t.Errorf("inner.Parent() = %T(%p), want mid(%p)", p2, p2, mid)
	}

	p3 := mid.Parent()
	if p3 != root {
		t.Errorf("mid.Parent() = %T(%p), want root(%p)", p3, p3, root)
	}

	// Trigger hover.
	enterEvt := event.NewMouseEvent(
		event.MouseEnter, event.ButtonNone, 0,
		geometry.Pt(50, 40), geometry.Pt(50, 40), event.ModNone,
	)
	btn.Event(ctx, enterEvt)

	if !root.IsSceneDirty() {
		t.Error("root boundary should be scene-dirty after deep hover — " +
			"propagateDirtyUpward failed to walk 3-level parent chain")
	}

	if callbackCount == 0 {
		t.Error("onBoundaryDirty callback should have fired")
	}
}

// TestHoverE2E_WindowHandleEvent_FullChain verifies the complete chain through
// Window.HandleEvent → updateHover → hitTest → MouseEnter → dirty propagation.
func TestHoverE2E_WindowHandleEvent_FullChain(t *testing.T) {
	prev := widget.GetSceneRecorderFactory()
	widget.RegisterSceneRecorder(func(s widget.SceneCache, w, h int) (widget.Canvas, func()) {
		rec := internalRender.NewSceneCanvas(s.(*scene.Scene), w, h)
		return rec, rec.Close
	})
	defer widget.RegisterSceneRecorder(prev)

	a := New()
	win := a.Window()

	// Create a button with known bounds.
	btn := button.New(button.Text("Click Me"))

	// Root container with the button.
	root := &boxContainer{kids: []widget.Widget{btn}}
	root.SetVisible(true)
	root.SetEnabled(true)
	root.SetBounds(geometry.NewRect(0, 0, 800, 600))
	root.SetScreenOrigin(geometry.Pt(0, 0))

	win.SetRoot(root)

	// Layout and first frame to set up bounds and screen origins.
	// SetRoot marks root as boundary and mounts the tree.
	// Frame performs layout and initial draw.
	btn.SetBounds(geometry.NewRect(50, 50, 200, 90))
	btn.SetScreenOrigin(geometry.Pt(50, 50))

	// First paint to wire onBoundaryDirty callback.
	PaintBoundaryLayersWithContext(win.Root(), nil, win.Context())

	// Clear dirty state from initial paint.
	// ClearSceneDirty so we can detect re-dirtying from hover.
	type sceneClearer interface {
		ClearSceneDirty()
	}
	if sc, ok := win.Root().(sceneClearer); ok {
		sc.ClearSceneDirty()
	}

	// Verify root is clean before hover.
	type sceneDirtyChecker interface {
		IsSceneDirty() bool
	}
	if sd, ok := win.Root().(sceneDirtyChecker); ok {
		if sd.IsSceneDirty() {
			t.Log("warning: root still dirty after clear — may affect test")
		}
	}

	// Simulate mouse move into button area.
	moveEvt := event.NewMouseEvent(
		event.MouseMove, event.ButtonNone, 0,
		geometry.Pt(100, 70), geometry.Pt(100, 70), event.ModNone,
	)
	win.HandleEvent(moveEvt)

	// Verify hover target.
	if win.HoveredWidget() == nil {
		t.Fatal("no widget hovered — hitTest returned nil. " +
			"Check ScreenBounds on button")
	}

	// The hovered widget should be the button.
	if win.HoveredWidget() != btn {
		t.Errorf("hovered widget = %T, want button", win.HoveredWidget())
	}

	// Verify root boundary is scene-dirty.
	if sd, ok := win.Root().(sceneDirtyChecker); ok {
		if !sd.IsSceneDirty() {
			t.Error("root boundary should be scene-dirty after hover on button — " +
				"the dirty propagation chain is broken")
		}
	} else {
		t.Error("root does not implement IsSceneDirty")
	}

	// Verify Window knows it needs redraw.
	if !win.NeedsRedraw() {
		t.Error("Window.NeedsRedraw() should be true after hover event")
	}
}
