package app

import (
	"testing"

	"github.com/gogpu/ui/event"
	"github.com/gogpu/ui/geometry"
	"github.com/gogpu/ui/widget"
)

// selfDirtyingLeaf re-dirties itself while it is being drawn, the way an animated
// widget does (spinner) and the way any widget written from another goroutine does
// when the write lands mid-record.
type selfDirtyingLeaf struct {
	widget.WidgetBase
	dirtyDuringDraw bool
}

func (w *selfDirtyingLeaf) Layout(_ widget.Context, c geometry.Constraints) geometry.Size {
	return c.Constrain(geometry.Sz(48, 48))
}

func (w *selfDirtyingLeaf) Draw(_ widget.Context, canvas widget.Canvas) {
	canvas.DrawRect(w.Bounds(), widget.RGBA8(255, 0, 0, 255))
	if w.dirtyDuringDraw {
		w.SetNeedsRedraw(true)
	}
}

func (w *selfDirtyingLeaf) Event(_ widget.Context, _ event.Event) bool { return false }
func (w *selfDirtyingLeaf) Children() []widget.Widget                  { return nil }

// A boundary that is still scene-dirty when its recording ends registers itself for
// the next frame (recordBoundary). The frame's flat dirty set is only the O(1)
// frame-skip gate, so the render loop consumes it BEFORE painting — and that
// registration, made during painting, has to survive to the next frame.
//
// It did not: draw() cleared the set at the end of the frame and threw the
// registration away. That is unrecoverable rather than a dropped frame, because the
// widget's own sceneDirty is still true, so every later InvalidateScene takes the
// already-dirty O(1) guard and returns without telling the window anything. The
// boundary is never painted again. A terminal emulator under continuous output froze
// within one frame and stayed frozen after the output stopped, while the render loop
// kept being woken 60 times a second and skipping every frame.
func TestDirtyBoundaryRegisteredDuringPaintSurvivesTheFrame(t *testing.T) {
	cleanup := setupSceneRecorder(t)
	defer cleanup()

	a := New()
	w := a.Window()

	root := &testContainer{}
	root.SetVisible(true)
	root.SetRepaintBoundary(true)
	root.SetBounds(geometry.NewRect(0, 0, 800, 600))
	root.SetScreenOrigin(geometry.Pt(0, 0))

	child := &selfDirtyingLeaf{}
	child.SetVisible(true)
	child.SetBounds(geometry.NewRect(10, 10, 48, 48))
	child.SetParent(root)
	root.kids = []widget.Widget{child}

	w.SetRoot(root)

	// Settle: record once, then clear everything the first paint dirtied.
	PaintBoundaryLayersWithContext(root, nil, w.Context())
	w.ClearDirtyBoundaries()
	w.ClearAfterPaint()
	root.ClearSceneDirty()
	widget.ClearRedrawInTree(root)
	if w.HasDirtyBoundaries() {
		t.Fatal("pre-condition: the set should be empty after a clear")
	}

	// Frame N: something dirties the child, which propagates to the boundary.
	child.dirtyDuringDraw = true
	child.SetNeedsRedraw(true)
	if !w.HasDirtyBoundaries() {
		t.Fatal("pre-condition: a dirtied child must register its boundary")
	}

	// The render loop consumes the gate, then paints. The child re-dirties itself
	// mid-record, so the boundary registers itself for frame N+1.
	w.ClearDirtyBoundaries()
	PaintBoundaryLayersWithContext(root, nil, w.Context())

	if !w.HasDirtyBoundaries() {
		t.Fatal("a boundary that re-dirtied while painting was not registered for the next frame")
	}

	// And it is still scene-dirty, which is what makes losing that registration
	// permanent rather than a dropped frame: nothing re-registers a boundary whose
	// InvalidateScene already took the already-dirty guard.
	if !root.IsSceneDirty() {
		t.Fatal("expected the boundary to still be scene-dirty after re-dirtying")
	}
}
