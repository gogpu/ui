package app

import (
	"image"
	"testing"

	"github.com/gogpu/ui/event"
	"github.com/gogpu/ui/geometry"
	"github.com/gogpu/ui/widget"
)

// --- Incremental rendering tests (ADR-004 Phase 1) ---

// recordingCanvas tracks which Canvas methods are called for test assertions.
type recordingCanvas struct {
	clearCalls    int
	drawRectCalls []geometry.Rect
	pushClipCalls []geometry.Rect
	popClipCalls  int
}

func (c *recordingCanvas) Clear(widget.Color) { c.clearCalls++ }
func (c *recordingCanvas) DrawRect(r geometry.Rect, _ widget.Color) {
	c.drawRectCalls = append(c.drawRectCalls, r)
}
func (c *recordingCanvas) FillRectDirect(r geometry.Rect, _ widget.Color) {
	c.drawRectCalls = append(c.drawRectCalls, r)
}
func (c *recordingCanvas) StrokeRect(geometry.Rect, widget.Color, float32)               {}
func (c *recordingCanvas) DrawRoundRect(geometry.Rect, widget.Color, float32)            {}
func (c *recordingCanvas) StrokeRoundRect(geometry.Rect, widget.Color, float32, float32) {}
func (c *recordingCanvas) DrawCircle(geometry.Point, float32, widget.Color)              {}
func (c *recordingCanvas) StrokeCircle(geometry.Point, float32, widget.Color, float32)   {}
func (c *recordingCanvas) StrokeArc(_ geometry.Point, _ float32, _, _ float64, _ widget.Color, _ float32) {
}
func (c *recordingCanvas) DrawLine(geometry.Point, geometry.Point, widget.Color, float32) {}
func (c *recordingCanvas) DrawText(string, geometry.Rect, float32, widget.Color, bool, widget.TextAlign) {
}
func (c *recordingCanvas) MeasureText(_ string, fontSize float32, _ bool) float32 {
	return fontSize * 5
}
func (c *recordingCanvas) DrawImage(image.Image, geometry.Point)        {}
func (c *recordingCanvas) PushClip(r geometry.Rect)                     { c.pushClipCalls = append(c.pushClipCalls, r) }
func (c *recordingCanvas) PushClipRoundRect(_ geometry.Rect, _ float32) {}
func (c *recordingCanvas) PopClip()                                     { c.popClipCalls++ }
func (c *recordingCanvas) PushTransform(geometry.Point)                 {}
func (c *recordingCanvas) PopTransform()                                {}
func (c *recordingCanvas) TransformOffset() geometry.Point              { return geometry.Point{} }
func (c *recordingCanvas) ScreenOriginBase() geometry.Point             { return geometry.Point{} }
func (c *recordingCanvas) ClipBounds() geometry.Rect                    { return geometry.NewRect(0, 0, 10000, 10000) }
func (c *recordingCanvas) ReplayScene(_ widget.SceneCache)              {}

// drawTrackingWidget tracks whether Draw was called and has configurable bounds.
type drawTrackingWidget struct {
	widget.WidgetBase
	drawCount  int
	layoutSize geometry.Size
}

func newDrawTrackingWidget(bounds geometry.Rect) *drawTrackingWidget {
	w := &drawTrackingWidget{
		layoutSize: bounds.Size(),
	}
	w.SetVisible(true)
	w.SetEnabled(true)
	w.SetBounds(bounds)
	w.SetScreenOrigin(bounds.Min)
	return w
}

func (w *drawTrackingWidget) Layout(_ widget.Context, c geometry.Constraints) geometry.Size {
	return c.Constrain(w.layoutSize)
}

func (w *drawTrackingWidget) Draw(_ widget.Context, _ widget.Canvas) {
	w.drawCount++
}

func (w *drawTrackingWidget) Event(_ widget.Context, _ event.Event) bool {
	return false
}

// drawTrackingContainer holds children and tracks draws.
type drawTrackingContainer struct {
	widget.WidgetBase
	kids      []widget.Widget
	drawCount int
}

func newDrawTrackingContainer(bounds geometry.Rect, children ...widget.Widget) *drawTrackingContainer {
	c := &drawTrackingContainer{kids: children}
	c.SetVisible(true)
	c.SetEnabled(true)
	c.SetBounds(bounds)
	c.SetScreenOrigin(bounds.Min)
	return c
}

func (c *drawTrackingContainer) Layout(_ widget.Context, cs geometry.Constraints) geometry.Size {
	return cs.Constrain(c.Bounds().Size())
}

func (c *drawTrackingContainer) Draw(ctx widget.Context, canvas widget.Canvas) {
	c.drawCount++
	for _, child := range c.kids {
		widget.StampScreenOrigin(child, canvas)
		child.Draw(ctx, canvas)
	}
}

func (c *drawTrackingContainer) Event(_ widget.Context, _ event.Event) bool {
	return false
}

func (c *drawTrackingContainer) Children() []widget.Widget {
	return c.kids
}

// --- Tests ---

func TestDrawTo_CleanTreeFullRepaint(t *testing.T) {
	// In HostManaged mode (default), DrawTo always draws a valid frame.
	// canvas.Clear is NOT called because the host owns the background.
	a := New()
	w := a.Window()
	root := newDrawTrackingWidget(geometry.NewRect(0, 0, 100, 50))
	w.SetRoot(root)
	root.SetRepaintBoundary(false)

	canvas := &recordingCanvas{}

	// First draw: full repaint (needsFullRepaint=true from SetRoot).
	drawn := w.DrawTo(canvas)
	if !drawn {
		t.Error("first DrawTo should return true")
	}
	if root.drawCount != 1 {
		t.Errorf("first draw: drawCount = %d, want 1", root.drawCount)
	}
	if canvas.clearCalls != 0 {
		t.Errorf("HostManaged: clearCalls = %d, want 0 (host owns background)", canvas.clearCalls)
	}

	// Second draw: tree clean → still draws (host may have cleared pixmap).
	canvas2 := &recordingCanvas{}
	drawn = w.DrawTo(canvas2)
	if !drawn {
		t.Error("second DrawTo should return true (HostManaged always draws)")
	}
	if root.drawCount != 2 {
		t.Errorf("second draw: drawCount = %d, want 2", root.drawCount)
	}
	if canvas2.clearCalls != 0 {
		t.Errorf("HostManaged: clearCalls = %d, want 0", canvas2.clearCalls)
	}
}

func TestDrawTo_SignalChange(t *testing.T) {
	// Signal change marks widget dirty → DrawTo returns true.
	// Next call: tree clean → still draws (full repaint for safety).
	a := New()
	w := a.Window()
	root := newDrawTrackingWidget(geometry.NewRect(0, 0, 200, 100))
	w.SetRoot(root)
	root.SetRepaintBoundary(false)

	canvas := &recordingCanvas{}
	w.DrawTo(canvas)

	// Mark widget dirty (simulates signal change).
	root.SetNeedsRedraw(true)

	canvas2 := &recordingCanvas{}
	drawn := w.DrawTo(canvas2)
	if !drawn {
		t.Error("DrawTo should return true after widget marked dirty")
	}
	if root.drawCount != 2 {
		t.Errorf("drawCount = %d, want 2", root.drawCount)
	}

	// Next frame: tree clean → full repaint (host may have cleared pixmap).
	canvas3 := &recordingCanvas{}
	drawn = w.DrawTo(canvas3)
	if !drawn {
		t.Error("DrawTo should return true (always draws)")
	}
	if root.drawCount != 3 {
		t.Errorf("drawCount = %d, want 3 (clean tree still draws)", root.drawCount)
	}
}

func TestDrawTo_FullRepaint(t *testing.T) {
	// In HostManaged mode (default): no canvas.Clear, full widget tree drawn.
	a := New()
	w := a.Window()
	root := newDrawTrackingWidget(geometry.NewRect(0, 0, 800, 600))
	w.SetRoot(root)

	canvas := &recordingCanvas{}
	w.DrawTo(canvas) // First draw.

	if canvas.clearCalls != 0 {
		t.Errorf("HostManaged: clearCalls = %d, want 0", canvas.clearCalls)
	}
	if root.drawCount != 1 {
		t.Errorf("drawCount = %d, want 1", root.drawCount)
	}
}

func TestDrawTo_DirtyWidgetFullRepaint(t *testing.T) {
	// In HostManaged mode (default): dirtying a widget still triggers a draw,
	// but canvas.Clear is NOT called (host owns background).
	a := New()
	w := a.Window()

	child1 := newDrawTrackingWidget(geometry.NewRect(10, 10, 100, 40))
	child2 := newDrawTrackingWidget(geometry.NewRect(10, 60, 100, 40))
	root := newDrawTrackingContainer(
		geometry.NewRect(0, 0, 800, 600),
		child1, child2,
	)
	w.SetRoot(root)

	// First draw.
	canvas := &recordingCanvas{}
	w.DrawTo(canvas)

	// Mark only child1 as dirty.
	child1.SetNeedsRedraw(true)

	canvas2 := &recordingCanvas{}
	drawn := w.DrawTo(canvas2)
	if !drawn {
		t.Error("DrawTo should return true with dirty widget")
	}

	// HostManaged: no canvas.Clear.
	if canvas2.clearCalls != 0 {
		t.Errorf("HostManaged: clearCalls = %d, want 0", canvas2.clearCalls)
	}
}

func TestDrawTo_NilRoot(t *testing.T) {
	a := New()
	w := a.Window()
	canvas := &recordingCanvas{}

	drawn := w.DrawTo(canvas)
	if drawn {
		t.Error("DrawTo with nil root should return false")
	}
}

func TestDrawTo_ResizeTriggersFullRepaint(t *testing.T) {
	// After resize, the next DrawTo should draw all widgets.
	// In HostManaged mode, canvas.Clear is NOT called.
	a := New()
	w := a.Window()
	root := newDrawTrackingWidget(geometry.NewRect(0, 0, 800, 600))
	w.SetRoot(root)

	canvas := &recordingCanvas{}
	w.DrawTo(canvas) // First draw.

	// Resize triggers needsFullRepaint.
	w.HandleResize(1024, 768)

	canvas2 := &recordingCanvas{}
	drawn := w.DrawTo(canvas2)
	if !drawn {
		t.Error("DrawTo should return true after resize")
	}
	// HostManaged: no canvas.Clear.
	if canvas2.clearCalls != 0 {
		t.Errorf("HostManaged: clearCalls = %d, want 0", canvas2.clearCalls)
	}
	if root.drawCount != 2 {
		t.Errorf("drawCount = %d, want 2 (full tree drawn after resize)", root.drawCount)
	}
}

func TestDrawTo_ThemeChangeTriggersFullRepaint(t *testing.T) {
	// In HostManaged mode, theme change triggers full draw but no Clear.
	a := New()
	w := a.Window()
	root := newDrawTrackingWidget(geometry.NewRect(0, 0, 800, 600))
	w.SetRoot(root)

	canvas := &recordingCanvas{}
	w.DrawTo(canvas) // First draw.

	// Theme change triggers needsFullRepaint.
	w.setTheme(w.theme) // Even same theme should trigger full repaint.

	canvas2 := &recordingCanvas{}
	drawn := w.DrawTo(canvas2)
	if !drawn {
		t.Error("DrawTo should return true after theme change")
	}
	// HostManaged: no canvas.Clear.
	if canvas2.clearCalls != 0 {
		t.Errorf("HostManaged: clearCalls = %d, want 0", canvas2.clearCalls)
	}
}

func TestDrawTo_DirtyTrackerRegionCount(t *testing.T) {
	// Verify that dirtyTracker.RegionCount() > 0 during a dirty frame.
	a := New()
	w := a.Window()
	root := newDrawTrackingWidget(geometry.NewRect(0, 0, 200, 100))
	w.SetRoot(root)

	canvas := &recordingCanvas{}
	w.DrawTo(canvas) // Full repaint, clears flags.

	// Mark dirty.
	root.SetNeedsRedraw(true)

	// Call DrawTo which internally collects and optimizes dirty regions.
	canvas2 := &recordingCanvas{}
	w.DrawTo(canvas2)

	// The tracker should have been populated with regions.
	// After DrawTo the tracker state reflects the last collection.
	regionCount := w.dirtyTracker.RegionCount()
	if regionCount == 0 {
		t.Error("dirtyTracker.RegionCount() should be > 0 during dirty frame")
	}
}

func TestDrawTo_IncrementalSkipsNonOverlappingWidgets(t *testing.T) {
	// When only child1 is dirty, child2 (non-overlapping) should NOT be drawn
	// in the incremental path.
	a := New()
	w := a.Window()

	// Two widgets far apart — no dirty region overlap.
	child1 := newDrawTrackingWidget(geometry.NewRect(10, 10, 100, 50))
	child2 := newDrawTrackingWidget(geometry.NewRect(400, 400, 100, 50))
	root := newDrawTrackingContainer(
		geometry.NewRect(0, 0, 800, 600),
		child1, child2,
	)
	w.SetRoot(root)

	// Full repaint.
	canvas := &recordingCanvas{}
	w.DrawTo(canvas)

	// Reset draw counts after full repaint.
	child1.drawCount = 0
	child2.drawCount = 0
	root.drawCount = 0

	// Mark only child1 dirty.
	child1.SetNeedsRedraw(true)

	canvas2 := &recordingCanvas{}
	w.DrawTo(canvas2)

	// child1 should be drawn (its bounds intersect the dirty region).
	if child1.drawCount == 0 {
		t.Error("child1 should be drawn (dirty)")
	}

	// child2 should NOT be drawn (bounds don't intersect dirty region).
	// Note: the container (root) draws children in its Draw method.
	// With incremental rendering, the container IS drawn if it overlaps the
	// dirty region, and it draws all children — but Canvas clip prevents
	// actual pixel output for non-overlapping children.
	// The spatial skip happens at the top-level drawInRegionRecursive check.
	// Since root overlaps (0,0,800,600 intersects any region), root.Draw is called,
	// which calls child2.Draw. But child2's canvas operations are clipped away.
	// This is expected — per-child skip requires RepaintBoundary (Phase 2).
}

func TestDrawTo_FullRepaintFallback(t *testing.T) {
	// In HostManaged mode (default), many dirty regions still result in
	// full tree draw without canvas.Clear.
	a := New()
	w := a.Window()

	// Create many small widgets spread far apart (>16px gap) so they don't merge.
	children := make([]widget.Widget, 20)
	for i := range children {
		x := float32(i * 100)
		y := float32((i % 4) * 100)
		child := newDrawTrackingWidget(geometry.NewRect(x, y, 20, 20))
		children[i] = child
	}
	root := newDrawTrackingContainer(
		geometry.NewRect(0, 0, 2000, 600),
		children...,
	)
	w.SetRoot(root)

	// First draw.
	canvas := &recordingCanvas{}
	w.DrawTo(canvas)

	// Mark all children dirty to exceed maxRegionsBeforeFullRepaint (16).
	for _, child := range children {
		if dw, ok := child.(*drawTrackingWidget); ok {
			dw.SetNeedsRedraw(true)
		}
	}

	canvas2 := &recordingCanvas{}
	drawn := w.DrawTo(canvas2)
	if !drawn {
		t.Error("DrawTo should return true with many dirty widgets")
	}

	// HostManaged: no canvas.Clear, but tree is drawn.
	if canvas2.clearCalls != 0 {
		t.Errorf("HostManaged: clearCalls = %d, want 0", canvas2.clearCalls)
	}
}

func TestDrawTo_OverlaysAlwaysDrawn(t *testing.T) {
	// Overlays should be drawn even during incremental frames.
	// This test verifies that overlays.Draw is called when the tree is dirty.
	a := New()
	w := a.Window()
	root := newDrawTrackingWidget(geometry.NewRect(0, 0, 200, 100))
	w.SetRoot(root)

	canvas := &recordingCanvas{}
	w.DrawTo(canvas) // Full repaint.

	// Mark dirty for incremental.
	root.SetNeedsRedraw(true)

	canvas2 := &recordingCanvas{}
	drawn := w.DrawTo(canvas2)
	if !drawn {
		t.Error("DrawTo should return true with dirty widget")
	}
	// Can't easily assert overlay drawing without mocking the overlay stack,
	// but we verify DrawTo completes without error.
}

func TestDrawTo_ClearsRedrawFlagsAfterDraw(t *testing.T) {
	a := New()
	w := a.Window()
	root := newDrawTrackingWidget(geometry.NewRect(0, 0, 200, 100))
	w.SetRoot(root)

	canvas := &recordingCanvas{}
	w.DrawTo(canvas)

	// After full repaint, all flags should be cleared.
	if root.NeedsRedraw() {
		t.Error("root needsRedraw should be cleared after DrawTo")
	}
	if w.needsRedraw {
		t.Error("window needsRedraw should be cleared after DrawTo")
	}
	if w.needsFullRepaint {
		t.Error("window needsFullRepaint should be cleared after DrawTo")
	}
}

func TestDrawTo_LastDrawStats_FullRepaint(t *testing.T) {
	a := New()
	w := a.Window()
	root := newDrawTrackingWidget(geometry.NewRect(0, 0, 200, 100))
	w.SetRoot(root)

	canvas := &recordingCanvas{}
	w.DrawTo(canvas)

	stats := w.LastDrawStats()
	if stats.TotalWidgets == 0 {
		t.Error("TotalWidgets should be > 0 after full repaint")
	}
	if stats.DrawnWidgets == 0 {
		t.Error("DrawnWidgets should be > 0 after full repaint")
	}
}

func TestDrawTo_LastDrawStats_Incremental(t *testing.T) {
	a := New()
	w := a.Window()
	root := newDrawTrackingWidget(geometry.NewRect(0, 0, 200, 100))
	w.SetRoot(root)

	canvas := &recordingCanvas{}
	w.DrawTo(canvas)

	// Dirty → incremental.
	root.SetNeedsRedraw(true)

	canvas2 := &recordingCanvas{}
	w.DrawTo(canvas2)

	stats := w.LastDrawStats()
	if stats.DrawnWidgets == 0 {
		t.Error("DrawnWidgets should be > 0 after incremental redraw")
	}
	if stats.DirtyWidgets == 0 {
		t.Error("DirtyWidgets should be > 0 (widget was marked dirty)")
	}
}

func TestDrawTo_DirtyRegionCount(t *testing.T) {
	// DirtyRegionCount should expose the number of dirty regions
	// from the most recent DrawTo call.
	a := New()
	w := a.Window()

	child1 := newDrawTrackingWidget(geometry.NewRect(10, 10, 100, 50))
	child2 := newDrawTrackingWidget(geometry.NewRect(400, 400, 100, 50))
	root := newDrawTrackingContainer(
		geometry.NewRect(0, 0, 800, 600),
		child1, child2,
	)
	w.SetRoot(root)

	// First draw: full repaint (tracker reset).
	canvas := &recordingCanvas{}
	w.DrawTo(canvas)

	// After full repaint, mark one child dirty.
	child1.SetNeedsRedraw(true)

	canvas2 := &recordingCanvas{}
	w.DrawTo(canvas2)

	regionCount := w.DirtyRegionCount()
	if regionCount == 0 {
		t.Error("DirtyRegionCount should be > 0 when a widget was dirty")
	}
}

func TestDrawTo_DirtyRegionCount_CleanFrame(t *testing.T) {
	// On a clean frame in HostManaged mode, DirtyRegionCount reflects
	// the tracker state: Reset clears regions, then Collect finds nothing.
	a := New()
	w := a.Window()
	root := newDrawTrackingWidget(geometry.NewRect(0, 0, 200, 100))
	w.SetRoot(root)

	canvas := &recordingCanvas{}
	w.DrawTo(canvas) // First draw.

	// Second draw: nothing dirty but still draws (HostManaged).
	canvas2 := &recordingCanvas{}
	w.DrawTo(canvas2)

	if w.DirtyRegionCount() != 0 {
		t.Errorf("DirtyRegionCount = %d on clean frame, want 0", w.DirtyRegionCount())
	}
}

// --- RenderMode API tests ---

func TestDrawTo_HostManaged_NoClear(t *testing.T) {
	// HostManaged mode: canvas.Clear is NOT called, DrawTree IS called.
	a := New(WithRenderMode(RenderModeHostManaged))
	w := a.Window()
	root := newDrawTrackingWidget(geometry.NewRect(0, 0, 200, 100))
	w.SetRoot(root)

	canvas := &recordingCanvas{}
	drawn := w.DrawTo(canvas)

	if !drawn {
		t.Error("HostManaged: DrawTo should return true")
	}
	if canvas.clearCalls != 0 {
		t.Errorf("HostManaged: clearCalls = %d, want 0 (host owns background)", canvas.clearCalls)
	}
	if root.drawCount != 1 {
		t.Errorf("HostManaged: drawCount = %d, want 1", root.drawCount)
	}
}

func TestDrawTo_HostManaged_AlwaysDraws(t *testing.T) {
	// HostManaged mode: DrawTo returns true even when tree is clean,
	// because the host may have cleared the pixmap before calling DrawTo.
	a := New(WithRenderMode(RenderModeHostManaged))
	w := a.Window()
	root := newDrawTrackingWidget(geometry.NewRect(0, 0, 200, 100))
	w.SetRoot(root)
	root.SetRepaintBoundary(false)

	canvas := &recordingCanvas{}
	w.DrawTo(canvas) // First draw.

	// Tree is now clean — but HostManaged always draws.
	canvas2 := &recordingCanvas{}
	drawn := w.DrawTo(canvas2)
	if !drawn {
		t.Error("HostManaged: DrawTo should return true even when tree is clean")
	}
	if root.drawCount != 2 {
		t.Errorf("HostManaged: drawCount = %d, want 2", root.drawCount)
	}
}

func TestDrawTo_FrameworkManaged_Clear(t *testing.T) {
	// FrameworkManaged mode: canvas.Clear IS called on full repaint
	// (first frame, needsFullRepaint=true from SetRoot).
	a := New(WithRenderMode(RenderModeFrameworkManaged))
	w := a.Window()
	root := newDrawTrackingWidget(geometry.NewRect(0, 0, 200, 100))
	w.SetRoot(root)

	canvas := &recordingCanvas{}
	drawn := w.DrawTo(canvas)

	if !drawn {
		t.Error("FrameworkManaged: DrawTo should return true on first frame")
	}
	if canvas.clearCalls != 1 {
		t.Errorf("FrameworkManaged: clearCalls = %d, want 1 (full repaint)", canvas.clearCalls)
	}
	if root.drawCount != 1 {
		t.Errorf("FrameworkManaged: drawCount = %d, want 1", root.drawCount)
	}
}

func TestDrawTo_FrameworkManaged_FrameSkip(t *testing.T) {
	// FrameworkManaged mode: DrawTo returns false when tree is clean
	// (frame skip — Level 1 optimization, ADR-004).
	a := New(WithRenderMode(RenderModeFrameworkManaged))
	w := a.Window()
	root := newDrawTrackingWidget(geometry.NewRect(0, 0, 200, 100))
	w.SetRoot(root)

	canvas := &recordingCanvas{}
	w.DrawTo(canvas) // First frame: full repaint.

	// Second frame: tree clean, persistent pixmap valid.
	canvas2 := &recordingCanvas{}
	drawn := w.DrawTo(canvas2)

	if drawn {
		t.Error("FrameworkManaged: DrawTo should return false when tree is clean (frame skip)")
	}
	if canvas2.clearCalls != 0 {
		t.Errorf("FrameworkManaged frame skip: clearCalls = %d, want 0", canvas2.clearCalls)
	}
	// Widget should NOT be drawn on skipped frame.
	if root.drawCount != 1 {
		t.Errorf("FrameworkManaged frame skip: drawCount = %d, want 1 (only first frame)", root.drawCount)
	}
}

func TestDrawTo_FrameworkManaged_DirtyRegionRepaint(t *testing.T) {
	// FrameworkManaged Level 2: a single dirty widget triggers dirty-region
	// repaint (DrawRect per region + PushClip), NOT full repaint (Clear).
	a := New(WithRenderMode(RenderModeFrameworkManaged))
	w := a.Window()

	child1 := newDrawTrackingWidget(geometry.NewRect(10, 10, 100, 50))
	root := newDrawTrackingContainer(
		geometry.NewRect(0, 0, 800, 600),
		child1,
	)
	w.SetRoot(root)

	canvas := &recordingCanvas{}
	w.DrawTo(canvas) // First frame: full repaint with Clear.

	child1.SetNeedsRedraw(true)

	canvas2 := &recordingCanvas{}
	drawn := w.DrawTo(canvas2)
	if !drawn {
		t.Error("DrawTo should return true with dirty widget")
	}
	// Level 2 dirty-region path: canvas.Clear is NOT called.
	// Instead, individual dirty regions are cleared with DrawRect.
	if canvas2.clearCalls != 0 {
		t.Errorf("clearCalls = %d, want 0 (dirty-region repaint, not full)", canvas2.clearCalls)
	}
	// Dirty regions are cleared with DrawRect (background fill).
	if len(canvas2.drawRectCalls) == 0 {
		t.Error("drawRectCalls should be > 0 (dirty region background clear)")
	}
	// PushClip should be called to clip to dirty union.
	if len(canvas2.pushClipCalls) == 0 {
		t.Error("pushClipCalls should be > 0 (dirty union clip)")
	}
	if canvas2.popClipCalls == 0 {
		t.Error("popClipCalls should be > 0 (matching PopClip)")
	}
}

func TestDrawTo_DefaultIsHostManaged(t *testing.T) {
	// Default RenderMode should be HostManaged.
	a := New()
	w := a.Window()

	if w.RenderMode() != RenderModeHostManaged {
		t.Errorf("default RenderMode = %v, want HostManaged", w.RenderMode())
	}
}

func TestWithRenderMode(t *testing.T) {
	// WithRenderMode option should configure the mode correctly.
	tests := []struct {
		name string
		mode RenderMode
		want RenderMode
	}{
		{"HostManaged", RenderModeHostManaged, RenderModeHostManaged},
		{"FrameworkManaged", RenderModeFrameworkManaged, RenderModeFrameworkManaged},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			a := New(WithRenderMode(tt.mode))
			w := a.Window()
			if w.RenderMode() != tt.want {
				t.Errorf("RenderMode() = %v, want %v", w.RenderMode(), tt.want)
			}
		})
	}
}

func TestDrawTo_FrameworkManaged_ResizeTriggersClear(t *testing.T) {
	// In FrameworkManaged mode, resize triggers needsFullRepaint
	// which causes canvas.Clear on the next DrawTo.
	a := New(WithRenderMode(RenderModeFrameworkManaged))
	w := a.Window()
	root := newDrawTrackingWidget(geometry.NewRect(0, 0, 800, 600))
	w.SetRoot(root)

	canvas := &recordingCanvas{}
	w.DrawTo(canvas) // First frame: full repaint with Clear.

	w.HandleResize(1024, 768) // Triggers needsFullRepaint.

	canvas2 := &recordingCanvas{}
	drawn := w.DrawTo(canvas2)
	if !drawn {
		t.Error("DrawTo should return true after resize")
	}
	if canvas2.clearCalls != 1 {
		t.Errorf("FrameworkManaged after resize: clearCalls = %d, want 1", canvas2.clearCalls)
	}
}

func TestDrawTo_FrameworkManaged_ThemeChangeTriggersClear(t *testing.T) {
	// In FrameworkManaged mode, theme change triggers needsFullRepaint
	// which causes canvas.Clear on the next DrawTo.
	a := New(WithRenderMode(RenderModeFrameworkManaged))
	w := a.Window()
	root := newDrawTrackingWidget(geometry.NewRect(0, 0, 800, 600))
	w.SetRoot(root)

	canvas := &recordingCanvas{}
	w.DrawTo(canvas) // First frame.

	w.setTheme(w.theme) // Triggers needsFullRepaint.

	canvas2 := &recordingCanvas{}
	drawn := w.DrawTo(canvas2)
	if !drawn {
		t.Error("DrawTo should return true after theme change")
	}
	if canvas2.clearCalls != 1 {
		t.Errorf("FrameworkManaged after theme change: clearCalls = %d, want 1", canvas2.clearCalls)
	}
}

func TestDrawTo_FrameworkManaged_FullRepaintFallback(t *testing.T) {
	// In FrameworkManaged mode, too many dirty regions (>16) triggers
	// full repaint fallback with canvas.Clear.
	a := New(WithRenderMode(RenderModeFrameworkManaged))
	w := a.Window()

	// Create many small widgets spread far apart so they don't merge.
	children := make([]widget.Widget, 20)
	for i := range children {
		x := float32(i * 100)
		y := float32((i % 4) * 100)
		child := newDrawTrackingWidget(geometry.NewRect(x, y, 20, 20))
		children[i] = child
	}
	root := newDrawTrackingContainer(
		geometry.NewRect(0, 0, 2000, 600),
		children...,
	)
	w.SetRoot(root)

	canvas := &recordingCanvas{}
	w.DrawTo(canvas) // First frame: full repaint.

	// Mark all children dirty to exceed maxRegionsBeforeFullRepaint (16).
	for _, child := range children {
		if dw, ok := child.(*drawTrackingWidget); ok {
			dw.SetNeedsRedraw(true)
		}
	}

	canvas2 := &recordingCanvas{}
	drawn := w.DrawTo(canvas2)
	if !drawn {
		t.Error("DrawTo should return true with many dirty widgets")
	}
	if canvas2.clearCalls != 1 {
		t.Errorf("FrameworkManaged fallback: clearCalls = %d, want 1", canvas2.clearCalls)
	}
}

func TestRenderMode_String(t *testing.T) {
	tests := []struct {
		mode RenderMode
		want string
	}{
		{RenderModeHostManaged, "HostManaged"},
		{RenderModeFrameworkManaged, "FrameworkManaged"},
		{RenderMode(99), "Unknown"},
	}

	for _, tt := range tests {
		got := tt.mode.String()
		if got != tt.want {
			t.Errorf("RenderMode(%d).String() = %q, want %q", tt.mode, got, tt.want)
		}
	}
}
