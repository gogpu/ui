package app

import (
	"testing"

	"github.com/gogpu/ui/geometry"
	"github.com/gogpu/ui/primitives"
	"github.com/gogpu/ui/widget"
)

func TestHandleScaleChangeInvalidatesEveryBoundary(t *testing.T) {
	a := New()
	w := a.Window()

	root := &testContainer{}
	root.SetVisible(true)
	root.SetBounds(geometry.NewRect(0, 0, 800, 600))
	child := &testLeaf{}
	child.SetVisible(true)
	child.SetRepaintBoundary(true)
	child.SetBounds(geometry.NewRect(10, 10, 58, 58))
	child.SetParent(root)
	legacyBoundary := primitives.NewRepaintBoundary(&testLeaf{})
	legacyBoundary.ClearBoundaryDirty()
	legacyBoundary.SetParent(root)
	root.kids = []widget.Widget{child, legacyBoundary}
	w.SetRoot(root)

	overlayContent := &testLeaf{}
	overlayContent.SetVisible(true)
	overlayContent.SetBounds(geometry.NewRect(100, 100, 148, 148))
	mgr := &windowOverlayManager{window: w}
	mgr.PushOverlay(overlayContent, nil)

	boundaries := []*testLeaf{child, overlayContent}
	root.SetCachedScene(widget.NewSceneCache())
	root.ClearSceneDirty()
	for _, boundary := range boundaries {
		boundary.SetCachedScene(widget.NewSceneCache())
		boundary.ClearSceneDirty()
	}
	widget.ClearRedrawInTree(root)
	widget.ClearRedrawInTree(overlayContent)
	w.ClearAfterPaint()

	// A continuously animating boundary can already need redraw while its
	// retained scene is clean. A tree-wide MarkRedrawInTree call skips scene
	// invalidation in this state because SetNeedsRedraw has an O(1) guard.
	child.SetNeedsRedraw(true)
	child.ClearSceneDirty()
	if !child.NeedsRedraw() || child.IsSceneDirty() {
		t.Fatal("test setup: child must need redraw while its retained scene is clean")
	}

	w.HandleScaleChange(2)

	if got := w.Context().Scale(); got != 2 {
		t.Errorf("context scale = %v, want 2", got)
	}
	if !root.IsSceneDirty() {
		t.Error("root retained scene should be invalidated")
	}
	if !child.IsSceneDirty() {
		t.Error("already-redraw-dirty child retained scene should be invalidated")
	}
	if !overlayContent.IsSceneDirty() {
		t.Error("overlay retained scene should be invalidated")
	}
	if !legacyBoundary.IsBoundaryDirty() {
		t.Error("legacy RepaintBoundary scene should be invalidated")
	}
	if !w.NeedsRedraw() {
		t.Error("window should need redraw")
	}
	if !w.needsFullRepaint {
		t.Error("scale change should force a full repaint")
	}
}

func TestHandleScaleChangeWithoutWidgets(t *testing.T) {
	w := New().Window()
	w.HandleScaleChange(1.5)
	if got := w.Context().Scale(); got != 1.5 {
		t.Errorf("context scale = %v, want 1.5", got)
	}
	if !w.NeedsRedraw() {
		t.Error("window should need redraw")
	}
}
