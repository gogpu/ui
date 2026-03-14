package widget

import (
	"image"
	"testing"

	"github.com/gogpu/ui/event"
	"github.com/gogpu/ui/geometry"
)

// drawTrackingWidget tracks whether Draw was called.
type drawTrackingWidget struct {
	WidgetBase
	drawCalled bool
	drawCanvas Canvas
}

func newDrawTrackingWidget() *drawTrackingWidget {
	w := &drawTrackingWidget{}
	w.SetVisible(true)
	w.SetEnabled(true)
	return w
}

func (w *drawTrackingWidget) Layout(_ Context, c geometry.Constraints) geometry.Size {
	return c.Constrain(geometry.Sz(100, 50))
}

func (w *drawTrackingWidget) Draw(_ Context, canvas Canvas) {
	w.drawCalled = true
	w.drawCanvas = canvas
}

func (w *drawTrackingWidget) Event(_ Context, _ event.Event) bool { return false }

var _ Widget = (*drawTrackingWidget)(nil)

// invisibleWidget reports IsVisible() = false.
type invisibleWidget struct {
	WidgetBase
	drawCalled bool
}

func newInvisibleWidget() *invisibleWidget {
	w := &invisibleWidget{}
	w.SetVisible(false)
	w.SetEnabled(true)
	return w
}

func (w *invisibleWidget) Layout(_ Context, c geometry.Constraints) geometry.Size {
	return c.Constrain(geometry.Sz(50, 50))
}

func (w *invisibleWidget) Draw(_ Context, _ Canvas) {
	w.drawCalled = true
}

func (w *invisibleWidget) Event(_ Context, _ event.Event) bool { return false }

var _ Widget = (*invisibleWidget)(nil)

// --- DrawTree tests ---

func TestDrawTree_NilWidget(t *testing.T) {
	stats := DrawTree(nil, nil, nil)

	if stats.TotalWidgets != 0 {
		t.Errorf("TotalWidgets = %d, want 0", stats.TotalWidgets)
	}
	if stats.DrawnWidgets != 0 {
		t.Errorf("DrawnWidgets = %d, want 0", stats.DrawnWidgets)
	}
}

func TestDrawTree_SingleDirtyWidget(t *testing.T) {
	w := newDrawTrackingWidget()
	w.SetNeedsRedraw(true)

	canvas := &noopCanvas{}
	stats := DrawTree(w, nil, canvas)

	if !w.drawCalled {
		t.Error("Draw should be called on dirty widget")
	}
	if stats.TotalWidgets != 1 {
		t.Errorf("TotalWidgets = %d, want 1", stats.TotalWidgets)
	}
	if stats.DrawnWidgets != 1 {
		t.Errorf("DrawnWidgets = %d, want 1", stats.DrawnWidgets)
	}
	if stats.DirtyWidgets != 1 {
		t.Errorf("DirtyWidgets = %d, want 1", stats.DirtyWidgets)
	}
	if stats.CleanWidgets != 0 {
		t.Errorf("CleanWidgets = %d, want 0", stats.CleanWidgets)
	}
}

func TestDrawTree_SingleCleanWidget(t *testing.T) {
	w := newDrawTrackingWidget()
	w.ClearRedraw()

	canvas := &noopCanvas{}
	stats := DrawTree(w, nil, canvas)

	// In Sub-Phase 1, clean widgets are still drawn (gg clears pixmap).
	if !w.drawCalled {
		t.Error("Draw should be called even on clean widget in Sub-Phase 1")
	}
	if stats.CleanWidgets != 1 {
		t.Errorf("CleanWidgets = %d, want 1", stats.CleanWidgets)
	}
	if stats.DirtyWidgets != 0 {
		t.Errorf("DirtyWidgets = %d, want 0", stats.DirtyWidgets)
	}
}

func TestDrawTree_CustomWidgetWithoutBase(t *testing.T) {
	// Custom widget without WidgetBase is treated as always dirty.
	w := &customWidget{}

	canvas := &noopCanvas{}
	stats := DrawTree(w, nil, canvas)

	if stats.DirtyWidgets != 1 {
		t.Errorf("DirtyWidgets = %d, want 1 (custom widget without NeedsRedraw)", stats.DirtyWidgets)
	}
	if stats.DrawnWidgets != 1 {
		t.Errorf("DrawnWidgets = %d, want 1", stats.DrawnWidgets)
	}
}

func TestDrawTree_PassesCanvasToWidget(t *testing.T) {
	w := newDrawTrackingWidget()
	w.SetNeedsRedraw(true)

	canvas := &noopCanvas{}
	DrawTree(w, nil, canvas)

	if w.drawCanvas != canvas {
		t.Error("DrawTree should pass canvas to widget's Draw method")
	}
}

// --- CollectDrawStats tests ---

func TestCollectDrawStats_NilWidget(t *testing.T) {
	stats := CollectDrawStats(nil)

	if stats.TotalWidgets != 0 {
		t.Errorf("TotalWidgets = %d, want 0", stats.TotalWidgets)
	}
}

func TestCollectDrawStats_SingleDirtyWidget(t *testing.T) {
	w := newDrawTrackingWidget()
	w.SetNeedsRedraw(true)

	stats := CollectDrawStats(w)

	if stats.TotalWidgets != 1 {
		t.Errorf("TotalWidgets = %d, want 1", stats.TotalWidgets)
	}
	if stats.DirtyWidgets != 1 {
		t.Errorf("DirtyWidgets = %d, want 1", stats.DirtyWidgets)
	}
	// CollectDrawStats should NOT call Draw.
	if w.drawCalled {
		t.Error("CollectDrawStats should not call Draw")
	}
}

func TestCollectDrawStats_TreeWithChildren(t *testing.T) {
	root := newDrawTrackingWidget()
	root.SetNeedsRedraw(true)

	child1 := newDrawTrackingWidget()
	child1.SetNeedsRedraw(true)
	root.AddChild(child1)

	child2 := newDrawTrackingWidget()
	child2.ClearRedraw()
	root.AddChild(child2)

	grandchild := newDrawTrackingWidget()
	grandchild.SetNeedsRedraw(true)
	child1.AddChild(grandchild)

	stats := CollectDrawStats(root)

	if stats.TotalWidgets != 4 {
		t.Errorf("TotalWidgets = %d, want 4", stats.TotalWidgets)
	}
	if stats.DirtyWidgets != 3 {
		t.Errorf("DirtyWidgets = %d, want 3", stats.DirtyWidgets)
	}
	if stats.CleanWidgets != 1 {
		t.Errorf("CleanWidgets = %d, want 1", stats.CleanWidgets)
	}
}

func TestCollectDrawStats_InvisibleWidget(t *testing.T) {
	w := newInvisibleWidget()
	w.SetNeedsRedraw(true)

	stats := CollectDrawStats(w)

	if stats.TotalWidgets != 1 {
		t.Errorf("TotalWidgets = %d, want 1", stats.TotalWidgets)
	}
	if stats.SkippedWidgets != 1 {
		t.Errorf("SkippedWidgets = %d, want 1", stats.SkippedWidgets)
	}
	if stats.DirtyWidgets != 0 {
		t.Errorf("DirtyWidgets = %d, want 0 (invisible widgets are skipped)", stats.DirtyWidgets)
	}
}

func TestCollectDrawStats_MixedTree(t *testing.T) {
	root := newDrawTrackingWidget()
	root.SetNeedsRedraw(true)

	visible := newDrawTrackingWidget()
	visible.ClearRedraw()
	root.AddChild(visible)

	invisible := newInvisibleWidget()
	root.AddChild(invisible)

	stats := CollectDrawStats(root)

	if stats.TotalWidgets != 3 {
		t.Errorf("TotalWidgets = %d, want 3", stats.TotalWidgets)
	}
	if stats.DirtyWidgets != 1 {
		t.Errorf("DirtyWidgets = %d, want 1", stats.DirtyWidgets)
	}
	if stats.CleanWidgets != 1 {
		t.Errorf("CleanWidgets = %d, want 1", stats.CleanWidgets)
	}
	if stats.SkippedWidgets != 1 {
		t.Errorf("SkippedWidgets = %d, want 1", stats.SkippedWidgets)
	}
}

func TestCollectDrawStats_CustomWidgetWithChildren(t *testing.T) {
	child := newDrawTrackingWidget()
	child.ClearRedraw()

	w := &customWidget{children: []Widget{child}}

	stats := CollectDrawStats(w)

	if stats.TotalWidgets != 2 {
		t.Errorf("TotalWidgets = %d, want 2", stats.TotalWidgets)
	}
	if stats.DirtyWidgets != 1 {
		t.Errorf("DirtyWidgets = %d, want 1 (custom widget always dirty)", stats.DirtyWidgets)
	}
	if stats.CleanWidgets != 1 {
		t.Errorf("CleanWidgets = %d, want 1", stats.CleanWidgets)
	}
}

func TestCollectDrawStats_DoesNotClearFlags(t *testing.T) {
	w := newDrawTrackingWidget()
	w.SetNeedsRedraw(true)

	CollectDrawStats(w)

	// CollectDrawStats should not modify any state.
	if !w.NeedsRedraw() {
		t.Error("CollectDrawStats should not clear needsRedraw flag")
	}
}

// --- DrawStats zero value test ---

func TestDrawStats_ZeroValue(t *testing.T) {
	var stats DrawStats

	if stats.TotalWidgets != 0 || stats.DrawnWidgets != 0 ||
		stats.SkippedWidgets != 0 || stats.DirtyWidgets != 0 ||
		stats.CleanWidgets != 0 || stats.CachedWidgets != 0 {
		t.Error("zero-valued DrawStats should have all fields zero")
	}
}

// noopCanvas is a minimal Canvas implementation for testing.
type noopCanvas struct{}

func (c *noopCanvas) Clear(Color)                                                     {}
func (c *noopCanvas) DrawRect(geometry.Rect, Color)                                   {}
func (c *noopCanvas) StrokeRect(geometry.Rect, Color, float32)                        {}
func (c *noopCanvas) DrawRoundRect(geometry.Rect, Color, float32)                     {}
func (c *noopCanvas) StrokeRoundRect(geometry.Rect, Color, float32, float32)          {}
func (c *noopCanvas) DrawCircle(geometry.Point, float32, Color)                       {}
func (c *noopCanvas) StrokeCircle(geometry.Point, float32, Color, float32)            {}
func (c *noopCanvas) DrawLine(geometry.Point, geometry.Point, Color, float32)         {}
func (c *noopCanvas) DrawText(string, geometry.Rect, float32, Color, bool, TextAlign) {}
func (c *noopCanvas) DrawImage(image.Image, geometry.Point)                           {}
func (c *noopCanvas) PushClip(geometry.Rect)                                          {}
func (c *noopCanvas) PushClipRoundRect(_ geometry.Rect, _ float32)                    {}
func (c *noopCanvas) PopClip()                                                        {}
func (c *noopCanvas) PushTransform(geometry.Point)                                    {}
func (c *noopCanvas) PopTransform()                                                   {}

var _ Canvas = (*noopCanvas)(nil)
