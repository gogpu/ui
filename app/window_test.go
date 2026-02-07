package app

import (
	"testing"

	"github.com/gogpu/gpucontext"
	"github.com/gogpu/ui/event"
	"github.com/gogpu/ui/geometry"
	"github.com/gogpu/ui/theme"
	"github.com/gogpu/ui/widget"
)

func TestWindow_SetRoot(t *testing.T) {
	a := New()
	w := a.Window()
	root := newMockWidget()
	w.SetRoot(root)

	if w.Root() != root {
		t.Error("root widget not set")
	}
	if !w.NeedsLayout() {
		t.Error("setting root should mark layout as needed")
	}
}

func TestWindow_SetRoot_Replace(t *testing.T) {
	a := New()
	w := a.Window()
	first := newMockWidget()
	second := newMockWidget()

	w.SetRoot(first)
	w.Frame()
	w.SetRoot(second)

	if w.Root() != second {
		t.Error("root not replaced")
	}
	if !w.NeedsLayout() {
		t.Error("replacing root should mark layout as needed")
	}
}

func TestWindow_Frame_Layout(t *testing.T) {
	wp := &mockWindowProvider{width: 400, height: 300, scale: 1.0}
	a := New(WithWindowProvider(wp))
	w := a.Window()

	root := newMockWidget()
	root.layoutSize = geometry.Sz(400, 300)
	w.SetRoot(root)

	w.Frame()

	if !root.layoutCalled {
		t.Error("layout not called on root")
	}

	// Verify bounds were set on root.
	bounds := root.Bounds()
	if bounds.Width() != 400 || bounds.Height() != 300 {
		t.Errorf("bounds = %v, want (400, 300)", bounds)
	}
}

func TestWindow_Frame_SkipsLayoutWhenClean(t *testing.T) {
	a := New()
	w := a.Window()
	root := newMockWidget()
	w.SetRoot(root)

	// First frame: layout performed.
	w.Frame()
	root.layoutCalled = false

	// Second frame: layout should not be performed (nothing changed).
	w.Frame()
	if root.layoutCalled {
		t.Error("layout called on second frame when nothing changed")
	}
}

func TestWindow_Frame_RelayoutOnResize(t *testing.T) {
	a := New()
	w := a.Window()
	root := newMockWidget()
	w.SetRoot(root)

	w.Frame()
	root.layoutCalled = false

	// Resize triggers relayout.
	w.HandleResize(1024, 768)
	w.Frame()

	if !root.layoutCalled {
		t.Error("layout not called after resize")
	}
}

func TestWindow_HandleEvent_DispatchesToRoot(t *testing.T) {
	a := New()
	w := a.Window()
	root := newMockWidget()
	w.SetRoot(root)

	e := event.NewMouseEvent(
		event.MousePress,
		event.ButtonLeft,
		event.ButtonStateLeft,
		geometry.Pt(10, 20),
		geometry.Pt(10, 20),
		event.ModNone,
	)
	w.HandleEvent(e)

	if !root.eventCalled {
		t.Error("event not dispatched to root")
	}
	if root.lastEvent != e {
		t.Error("wrong event dispatched")
	}
}

func TestWindow_HandleEvent_NoRoot(t *testing.T) {
	a := New()
	w := a.Window()

	e := event.NewKeyEvent(event.KeyPress, event.KeyA, 'a', event.ModNone)
	// Should not panic.
	w.HandleEvent(e)
}

func TestWindow_HandleEvent_NilEvent(t *testing.T) {
	a := New()
	w := a.Window()
	root := newMockWidget()
	w.SetRoot(root)

	// Should not panic.
	w.HandleEvent(nil)
	if root.eventCalled {
		t.Error("nil event should not be dispatched")
	}
}

func TestWindow_HandleResize(t *testing.T) {
	a := New()
	w := a.Window()

	w.HandleResize(1920, 1080)

	size := w.WindowSize()
	if size.Width != 1920 || size.Height != 1080 {
		t.Errorf("size = %v, want (1920, 1080)", size)
	}
	if !w.NeedsLayout() {
		t.Error("resize should mark layout as needed")
	}
}

func TestWindow_HandleFocusChange_Gained(t *testing.T) {
	a := New()
	w := a.Window()
	root := newMockWidget()
	w.SetRoot(root)

	w.HandleFocusChange(true)

	if !root.eventCalled {
		t.Error("focus event not dispatched")
	}
	if fe, ok := root.lastEvent.(*event.FocusEvent); ok {
		if !fe.IsGained() {
			t.Error("expected focus gained event")
		}
	} else {
		t.Error("expected FocusEvent type")
	}
}

func TestWindow_HandleFocusChange_Lost(t *testing.T) {
	a := New()
	w := a.Window()
	root := newMockWidget()
	w.SetRoot(root)

	w.HandleFocusChange(false)

	if !root.eventCalled {
		t.Error("focus event not dispatched")
	}
	if fe, ok := root.lastEvent.(*event.FocusEvent); ok {
		if !fe.IsLost() {
			t.Error("expected focus lost event")
		}
	} else {
		t.Error("expected FocusEvent type")
	}
}

func TestWindow_HandleFocusChange_NoRoot(t *testing.T) {
	a := New()
	w := a.Window()
	// Should not panic.
	w.HandleFocusChange(true)
}

func TestWindow_ScaleFromProvider(t *testing.T) {
	wp := &mockWindowProvider{width: 800, height: 600, scale: 2.0}
	a := New(WithWindowProvider(wp))
	w := a.Window()

	if w.Context().Scale() != 2.0 {
		t.Errorf("scale = %v, want 2.0", w.Context().Scale())
	}
}

func TestWindow_DefaultScale_Headless(t *testing.T) {
	a := New()
	w := a.Window()

	if w.Context().Scale() != 1.0 {
		t.Errorf("headless scale = %v, want 1.0", w.Context().Scale())
	}
}

func TestWindow_DefaultSize_Headless(t *testing.T) {
	a := New()
	w := a.Window()

	size := w.WindowSize()
	if size.Width != defaultWidth || size.Height != defaultHeight {
		t.Errorf("headless size = %v, want (%d, %d)", size, defaultWidth, defaultHeight)
	}
}

func TestWindow_CursorSync(t *testing.T) {
	pp := &mockPlatformProvider{fontScale: 1.0}
	a := New(WithPlatformProvider(pp))
	w := a.Window()

	// Create a widget that sets cursor during layout (inside Frame).
	cursorWidget := &cursorSettingOnLayoutWidget{
		cursor: widget.CursorPointer,
	}
	w.SetRoot(cursorWidget)

	// Frame calls layout -> widget sets cursor -> syncCursor forwards to platform.
	w.Frame()

	if pp.lastCursor != gpucontext.CursorPointer {
		t.Errorf("cursor = %v, want Pointer", pp.lastCursor)
	}
}

func TestWindow_CursorSync_NoPlatform(t *testing.T) {
	a := New() // headless, no platform provider
	w := a.Window()
	root := newMockWidget()
	w.SetRoot(root)
	// Should not panic without platform provider.
	w.Frame()
}

func TestWindow_DrawTo(t *testing.T) {
	a := New()
	w := a.Window()
	root := newMockWidget()
	w.SetRoot(root)

	canvas := &mockCanvas{}
	w.DrawTo(canvas)

	if !root.drawCalled {
		t.Error("DrawTo did not call root Draw")
	}
}

func TestWindow_DrawTo_NoRoot(t *testing.T) {
	a := New()
	w := a.Window()
	canvas := &mockCanvas{}
	// Should not panic.
	w.DrawTo(canvas)
}

func TestWindow_DrawTo_NilCanvas(t *testing.T) {
	a := New()
	w := a.Window()
	root := newMockWidget()
	w.SetRoot(root)
	// Should not panic.
	w.DrawTo(nil)
}

func TestWindow_Theme(t *testing.T) {
	dark := theme.DefaultDark()
	a := New(WithTheme(dark))

	if a.Window().Theme() != dark {
		t.Error("window theme does not match app theme")
	}
}

func TestWindow_InvalidateTriggersRelayout(t *testing.T) {
	wp := &mockWindowProvider{width: 800, height: 600, scale: 1.0}
	a := New(WithWindowProvider(wp))
	w := a.Window()
	root := newMockWidget()
	w.SetRoot(root)

	w.Frame()
	root.layoutCalled = false

	// Invalidate should mark needsLayout.
	w.Context().Invalidate()

	if !w.NeedsLayout() {
		t.Error("invalidation should mark layout as needed")
	}

	w.Frame()
	if !root.layoutCalled {
		t.Error("layout not called after invalidation")
	}
}

func TestWindow_SizeFromProvider_Updates(t *testing.T) {
	wp := &mockWindowProvider{width: 800, height: 600, scale: 1.0}
	a := New(WithWindowProvider(wp))
	w := a.Window()
	root := newMockWidget()
	w.SetRoot(root)

	w.Frame()
	root.layoutCalled = false

	// Simulate window resize by changing provider values.
	wp.width = 1024
	wp.height = 768

	w.Frame()

	if !root.layoutCalled {
		t.Error("layout not called when window size changed via provider")
	}
	size := w.WindowSize()
	if size.Width != 1024 || size.Height != 768 {
		t.Errorf("size = %v, want (1024, 768)", size)
	}
}

// --- Helper test types ---

// cursorSettingWidget sets the cursor during event handling.
type cursorSettingWidget struct {
	widget.WidgetBase
	cursor widget.CursorType
}

func (w *cursorSettingWidget) Layout(_ widget.Context, c geometry.Constraints) geometry.Size {
	return c.Constrain(geometry.Sz(100, 100))
}

func (w *cursorSettingWidget) Draw(_ widget.Context, _ widget.Canvas) {}

func (w *cursorSettingWidget) Event(ctx widget.Context, _ event.Event) bool {
	ctx.SetCursor(w.cursor)
	return true
}

// cursorSettingOnLayoutWidget sets the cursor during layout (inside Frame).
type cursorSettingOnLayoutWidget struct {
	widget.WidgetBase
	cursor widget.CursorType
}

func (w *cursorSettingOnLayoutWidget) Layout(ctx widget.Context, c geometry.Constraints) geometry.Size {
	ctx.SetCursor(w.cursor)
	return c.Constrain(geometry.Sz(100, 100))
}

func (w *cursorSettingOnLayoutWidget) Draw(_ widget.Context, _ widget.Canvas) {}

func (w *cursorSettingOnLayoutWidget) Event(_ widget.Context, _ event.Event) bool {
	return false
}

// mockCanvas implements widget.Canvas for testing.
type mockCanvas struct{}

func (m *mockCanvas) Clear(widget.Color)                                                   {}
func (m *mockCanvas) DrawRect(geometry.Rect, widget.Color)                                 {}
func (m *mockCanvas) StrokeRect(geometry.Rect, widget.Color, float32)                      {}
func (m *mockCanvas) DrawRoundRect(geometry.Rect, widget.Color, float32)                   {}
func (m *mockCanvas) StrokeRoundRect(geometry.Rect, widget.Color, float32, float32)        {}
func (m *mockCanvas) DrawCircle(geometry.Point, float32, widget.Color)                     {}
func (m *mockCanvas) StrokeCircle(geometry.Point, float32, widget.Color, float32)          {}
func (m *mockCanvas) DrawLine(geometry.Point, geometry.Point, widget.Color, float32)       {}
func (m *mockCanvas) DrawText(string, geometry.Rect, float32, widget.Color, bool, float32) {}
func (m *mockCanvas) PushClip(geometry.Rect)                                               {}
func (m *mockCanvas) PopClip()                                                             {}
func (m *mockCanvas) PushTransform(geometry.Point)                                         {}
func (m *mockCanvas) PopTransform()                                                        {}
