package scrollview

import (
	"image"
	"testing"

	"github.com/gogpu/ui/event"
	"github.com/gogpu/ui/geometry"
	"github.com/gogpu/ui/state"
	"github.com/gogpu/ui/widget"
)

// --- Direction Tests ---

func TestScrollDirection_String(t *testing.T) {
	tests := []struct {
		name string
		d    ScrollDirection
		want string
	}{
		{"vertical", Vertical, "Vertical"},
		{"horizontal", Horizontal, "Horizontal"},
		{"both", Both, "Both"},
		{"unknown", ScrollDirection(99), "Unknown"},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			if got := tc.d.String(); got != tc.want {
				t.Errorf("ScrollDirection(%d).String() = %q, want %q", tc.d, got, tc.want)
			}
		})
	}
}

func TestScrollbarVisibility_String(t *testing.T) {
	tests := []struct {
		name string
		v    ScrollbarVisibility
		want string
	}{
		{"auto", ScrollbarAuto, "Auto"},
		{"always", ScrollbarAlways, "Always"},
		{"never", ScrollbarNever, "Never"},
		{"unknown", ScrollbarVisibility(99), "Unknown"},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			if got := tc.v.String(); got != tc.want {
				t.Errorf("ScrollbarVisibility(%d).String() = %q, want %q", tc.v, got, tc.want)
			}
		})
	}
}

// --- Config Tests ---

func TestConfig_ResolvedScrollX(t *testing.T) {
	t.Run("static value", func(t *testing.T) {
		c := config{scrollX: 42}
		if got := c.ResolvedScrollX(); got != 42 {
			t.Errorf("ResolvedScrollX() = %v, want 42", got)
		}
	})

	t.Run("signal value", func(t *testing.T) {
		sig := state.NewSignal[float32](75)
		c := config{scrollX: 10, scrollXSignal: sig}
		if got := c.ResolvedScrollX(); got != 75 {
			t.Errorf("ResolvedScrollX() = %v, want 75", got)
		}
	})

	t.Run("readonly signal highest priority", func(t *testing.T) {
		sig := state.NewSignal[float32](50)
		computed := state.NewComputed(func() float32 { return 99 })
		c := config{
			scrollX:               10,
			scrollXSignal:         sig,
			readonlyScrollXSignal: computed,
		}
		if got := c.ResolvedScrollX(); got != 99 {
			t.Errorf("ResolvedScrollX() = %v, want 99", got)
		}
	})
}

func TestConfig_ResolvedScrollY(t *testing.T) {
	t.Run("static value", func(t *testing.T) {
		c := config{scrollY: 30}
		if got := c.ResolvedScrollY(); got != 30 {
			t.Errorf("ResolvedScrollY() = %v, want 30", got)
		}
	})

	t.Run("signal value", func(t *testing.T) {
		sig := state.NewSignal[float32](60)
		c := config{scrollY: 10, scrollYSignal: sig}
		if got := c.ResolvedScrollY(); got != 60 {
			t.Errorf("ResolvedScrollY() = %v, want 60", got)
		}
	})

	t.Run("readonly signal highest priority", func(t *testing.T) {
		sig := state.NewSignal[float32](50)
		computed := state.NewComputed(func() float32 { return 88 })
		c := config{
			scrollY:               10,
			scrollYSignal:         sig,
			readonlyScrollYSignal: computed,
		}
		if got := c.ResolvedScrollY(); got != 88 {
			t.Errorf("ResolvedScrollY() = %v, want 88", got)
		}
	})
}

func TestConfig_ResolvedScrollStep(t *testing.T) {
	t.Run("default", func(t *testing.T) {
		c := config{}
		if got := c.resolvedScrollStep(); got != defaultScrollStep {
			t.Errorf("resolvedScrollStep() = %v, want %v", got, defaultScrollStep)
		}
	})

	t.Run("custom", func(t *testing.T) {
		c := config{scrollStep: 20}
		if got := c.resolvedScrollStep(); got != 20 {
			t.Errorf("resolvedScrollStep() = %v, want 20", got)
		}
	})
}

// --- Widget Construction Tests ---

func TestNew(t *testing.T) {
	content := &mockWidget{}
	sv := New(content)

	if sv.content != content {
		t.Error("content should be the provided widget")
	}
	if !sv.IsVisible() {
		t.Error("should be visible by default")
	}
	if !sv.IsEnabled() {
		t.Error("should be enabled by default")
	}
	if sv.cfg.direction != Vertical {
		t.Errorf("direction = %v, want Vertical", sv.cfg.direction)
	}
	if sv.cfg.scrollbar != ScrollbarAuto {
		t.Errorf("scrollbar = %v, want ScrollbarAuto", sv.cfg.scrollbar)
	}
}

func TestNew_WithOptions(t *testing.T) {
	content := &mockWidget{}
	sv := New(content,
		DirectionOpt(Both),
		ScrollbarOpt(ScrollbarAlways),
		ScrollX(10),
		ScrollY(20),
		ScrollStep(30),
	)

	if sv.cfg.direction != Both {
		t.Errorf("direction = %v, want Both", sv.cfg.direction)
	}
	if sv.cfg.scrollbar != ScrollbarAlways {
		t.Errorf("scrollbar = %v, want ScrollbarAlways", sv.cfg.scrollbar)
	}
	if sv.cfg.scrollX != 10 {
		t.Errorf("scrollX = %v, want 10", sv.cfg.scrollX)
	}
	if sv.cfg.scrollY != 20 {
		t.Errorf("scrollY = %v, want 20", sv.cfg.scrollY)
	}
	if sv.cfg.scrollStep != 30 {
		t.Errorf("scrollStep = %v, want 30", sv.cfg.scrollStep)
	}
}

func TestNew_WithPainter(t *testing.T) {
	p := &testPainter{}
	sv := New(&mockWidget{}, PainterOpt(p))
	if sv.painter != p {
		t.Error("painter should be the provided painter")
	}
}

func TestNew_NilContent(t *testing.T) {
	sv := New(nil)
	if sv.content != nil {
		t.Error("content should be nil")
	}
	if sv.Children() != nil {
		t.Error("Children() should return nil for nil content")
	}
}

func TestNew_WithSignals(t *testing.T) {
	xSig := state.NewSignal[float32](10)
	ySig := state.NewSignal[float32](20)

	sv := New(&mockWidget{},
		ScrollXSignal(xSig),
		ScrollYSignal(ySig),
	)

	if sv.cfg.scrollXSignal != xSig {
		t.Error("scrollXSignal not set")
	}
	if sv.cfg.scrollYSignal != ySig {
		t.Error("scrollYSignal not set")
	}
}

func TestNew_WithReadonlySignals(t *testing.T) {
	xComputed := state.NewComputed(func() float32 { return 5 })
	yComputed := state.NewComputed(func() float32 { return 15 })

	sv := New(&mockWidget{},
		ScrollXReadonlySignal(xComputed),
		ScrollYReadonlySignal(yComputed),
	)

	if sv.cfg.readonlyScrollXSignal != xComputed {
		t.Error("readonlyScrollXSignal not set")
	}
	if sv.cfg.readonlyScrollYSignal != yComputed {
		t.Error("readonlyScrollYSignal not set")
	}
}

func TestNew_WithOnScroll(t *testing.T) {
	called := false
	sv := New(&mockWidget{}, OnScroll(func(_, _ float32) { called = true }))
	if sv.cfg.onScroll == nil {
		t.Error("onScroll should be set")
	}
	// Verify it's callable.
	sv.cfg.onScroll(0, 0)
	if !called {
		t.Error("onScroll should have been called")
	}
}

// --- IsFocusable Tests ---

func TestIsFocusable(t *testing.T) {
	sv := New(&mockWidget{})
	if !sv.IsFocusable() {
		t.Error("should be focusable by default")
	}

	sv.SetVisible(false)
	if sv.IsFocusable() {
		t.Error("should not be focusable when invisible")
	}

	sv.SetVisible(true)
	sv.SetEnabled(false)
	if sv.IsFocusable() {
		t.Error("should not be focusable when disabled")
	}
}

// --- Children Tests ---

func TestChildren(t *testing.T) {
	content := &mockWidget{}
	sv := New(content)
	children := sv.Children()
	if len(children) != 1 {
		t.Fatalf("Children() length = %d, want 1", len(children))
	}
	if children[0] != content {
		t.Error("Children()[0] should be the content widget")
	}
}

func TestChildren_NilContent(t *testing.T) {
	sv := New(nil)
	if sv.Children() != nil {
		t.Error("Children() should return nil for nil content")
	}
}

// --- Layout Tests ---

func TestLayout_Vertical(t *testing.T) {
	content := &mockWidget{preferredSize: geometry.Sz(200, 1000)}
	sv := New(content, DirectionOpt(Vertical))
	ctx := widget.NewContext()
	constraints := geometry.Loose(geometry.Sz(300, 400))

	size := sv.Layout(ctx, constraints)

	if size.Width != 300 || size.Height != 400 {
		t.Errorf("viewport size = %v, want (300, 400)", size)
	}
	if sv.contentSize.Height != 1000 {
		t.Errorf("contentSize.Height = %v, want 1000", sv.contentSize.Height)
	}
}

func TestLayout_Horizontal(t *testing.T) {
	content := &mockWidget{preferredSize: geometry.Sz(2000, 200)}
	sv := New(content, DirectionOpt(Horizontal))
	ctx := widget.NewContext()
	constraints := geometry.Loose(geometry.Sz(300, 400))

	size := sv.Layout(ctx, constraints)

	if size.Width != 300 || size.Height != 400 {
		t.Errorf("viewport size = %v, want (300, 400)", size)
	}
	if sv.contentSize.Width != 2000 {
		t.Errorf("contentSize.Width = %v, want 2000", sv.contentSize.Width)
	}
}

func TestLayout_Both(t *testing.T) {
	content := &mockWidget{preferredSize: geometry.Sz(2000, 3000)}
	sv := New(content, DirectionOpt(Both))
	ctx := widget.NewContext()
	constraints := geometry.Loose(geometry.Sz(300, 400))

	sv.Layout(ctx, constraints)

	if sv.contentSize.Width != 2000 || sv.contentSize.Height != 3000 {
		t.Errorf("contentSize = %v, want (2000, 3000)", sv.contentSize)
	}
}

func TestLayout_NilContent(t *testing.T) {
	sv := New(nil)
	ctx := widget.NewContext()
	constraints := geometry.Loose(geometry.Sz(300, 400))

	size := sv.Layout(ctx, constraints)

	if size.Width != 300 || size.Height != 400 {
		t.Errorf("viewport size = %v, want (300, 400)", size)
	}
}

func TestLayout_TightConstraints(t *testing.T) {
	content := &mockWidget{preferredSize: geometry.Sz(200, 1000)}
	sv := New(content)
	ctx := widget.NewContext()
	constraints := geometry.Tight(geometry.Sz(150, 100))

	size := sv.Layout(ctx, constraints)

	if size.Width != 150 || size.Height != 100 {
		t.Errorf("size = %v, want (150, 100)", size)
	}
}

// --- Draw Tests ---

func TestDraw_ClipAndTransform(t *testing.T) {
	content := &mockWidget{preferredSize: geometry.Sz(200, 1000)}
	sv := New(content, ScrollY(50))
	ctx := widget.NewContext()
	constraints := geometry.Loose(geometry.Sz(200, 300))

	sv.Layout(ctx, constraints)
	sv.SetBounds(geometry.NewRect(0, 0, 200, 300))

	canvas := &internalMockCanvas{}
	sv.Draw(ctx, canvas)

	if canvas.pushClipCount != 1 {
		t.Errorf("PushClip called %d times, want 1", canvas.pushClipCount)
	}
	if canvas.popClipCount != 1 {
		t.Errorf("PopClip called %d times, want 1", canvas.popClipCount)
	}
	if canvas.pushTransformCount != 1 {
		t.Errorf("PushTransform called %d times, want 1", canvas.pushTransformCount)
	}
	if canvas.popTransformCount != 1 {
		t.Errorf("PopTransform called %d times, want 1", canvas.popTransformCount)
	}

	// Verify the transform offset includes scroll offset.
	if canvas.lastTransformOffset.Y != -50 {
		t.Errorf("transform Y offset = %v, want -50", canvas.lastTransformOffset.Y)
	}
}

func TestDraw_EmptyBounds(t *testing.T) {
	sv := New(&mockWidget{})
	ctx := widget.NewContext()
	canvas := &internalMockCanvas{}

	sv.Draw(ctx, canvas)

	if canvas.pushClipCount != 0 {
		t.Error("PushClip should not be called for empty bounds")
	}
}

func TestDraw_NilContent(t *testing.T) {
	sv := New(nil)
	ctx := widget.NewContext()
	constraints := geometry.Loose(geometry.Sz(200, 200))
	sv.Layout(ctx, constraints)
	sv.SetBounds(geometry.NewRect(0, 0, 200, 200))

	canvas := &internalMockCanvas{}
	sv.Draw(ctx, canvas) // Should not panic.

	if canvas.pushClipCount != 1 {
		t.Error("PushClip should still be called for nil content (clip then restore)")
	}
}

func TestDraw_ScrollbarVisible(t *testing.T) {
	content := &mockWidget{preferredSize: geometry.Sz(200, 1000)}
	sv := New(content, ScrollbarOpt(ScrollbarAlways))
	ctx := widget.NewContext()
	constraints := geometry.Loose(geometry.Sz(200, 300))

	sv.Layout(ctx, constraints)
	sv.SetBounds(geometry.NewRect(0, 0, 200, 300))

	canvas := &internalMockCanvas{}
	sv.Draw(ctx, canvas)

	// Should have scrollbar drawing calls (DrawRoundRect for track and thumb).
	if len(canvas.drawRoundRects) == 0 {
		t.Error("scrollbar should be drawn when visibility=Always")
	}
}

func TestDraw_ScrollbarNever(t *testing.T) {
	content := &mockWidget{preferredSize: geometry.Sz(200, 1000)}
	sv := New(content, ScrollbarOpt(ScrollbarNever))
	ctx := widget.NewContext()
	constraints := geometry.Loose(geometry.Sz(200, 300))

	sv.Layout(ctx, constraints)
	sv.SetBounds(geometry.NewRect(0, 0, 200, 300))

	canvas := &internalMockCanvas{}
	sv.Draw(ctx, canvas)

	if len(canvas.drawRoundRects) != 0 {
		t.Error("scrollbar should not be drawn when visibility=Never")
	}
}

// --- Scroll Tests ---

func TestWheelEvent_ScrollDown(t *testing.T) {
	content := &mockWidget{preferredSize: geometry.Sz(200, 1000)}
	sv := New(content)
	ctx := widget.NewContext()
	constraints := geometry.Loose(geometry.Sz(200, 300))

	sv.Layout(ctx, constraints)
	sv.SetBounds(geometry.NewRect(0, 0, 200, 300))

	// Scroll down (positive delta Y).
	e := event.NewWheelEvent(
		geometry.Pt(0, 1),
		geometry.Pt(100, 150),
		geometry.Pt(100, 150),
		0,
	)

	consumed := sv.Event(ctx, e)
	if !consumed {
		t.Error("wheel event should be consumed")
	}

	_, scrollY := sv.ScrollOffset()
	if scrollY != defaultScrollStep {
		t.Errorf("scrollY = %v, want %v", scrollY, defaultScrollStep)
	}
}

func TestWheelEvent_ScrollUp(t *testing.T) {
	content := &mockWidget{preferredSize: geometry.Sz(200, 1000)}
	sv := New(content, ScrollY(100))
	ctx := widget.NewContext()
	constraints := geometry.Loose(geometry.Sz(200, 300))

	sv.Layout(ctx, constraints)
	sv.SetBounds(geometry.NewRect(0, 0, 200, 300))

	// Scroll up (negative delta Y).
	e := event.NewWheelEvent(
		geometry.Pt(0, -1),
		geometry.Pt(100, 150),
		geometry.Pt(100, 150),
		0,
	)

	sv.Event(ctx, e)

	_, scrollY := sv.ScrollOffset()
	expected := float32(100) - defaultScrollStep
	if scrollY != expected {
		t.Errorf("scrollY = %v, want %v", scrollY, expected)
	}
}

func TestWheelEvent_ClampToZero(t *testing.T) {
	content := &mockWidget{preferredSize: geometry.Sz(200, 1000)}
	sv := New(content, ScrollY(10))
	ctx := widget.NewContext()
	constraints := geometry.Loose(geometry.Sz(200, 300))

	sv.Layout(ctx, constraints)
	sv.SetBounds(geometry.NewRect(0, 0, 200, 300))

	// Scroll up more than current offset.
	e := event.NewWheelEvent(
		geometry.Pt(0, -10), // 10 ticks up
		geometry.Pt(100, 150),
		geometry.Pt(100, 150),
		0,
	)

	sv.Event(ctx, e)

	_, scrollY := sv.ScrollOffset()
	if scrollY != 0 {
		t.Errorf("scrollY = %v, want 0 (clamped)", scrollY)
	}
}

func TestWheelEvent_ClampToMax(t *testing.T) {
	content := &mockWidget{preferredSize: geometry.Sz(200, 500)}
	sv := New(content, ScrollY(150))
	ctx := widget.NewContext()
	constraints := geometry.Loose(geometry.Sz(200, 300))

	sv.Layout(ctx, constraints)
	sv.SetBounds(geometry.NewRect(0, 0, 200, 300))

	// Scroll down more than remaining space.
	e := event.NewWheelEvent(
		geometry.Pt(0, 100),
		geometry.Pt(100, 150),
		geometry.Pt(100, 150),
		0,
	)

	sv.Event(ctx, e)

	_, scrollY := sv.ScrollOffset()
	maxScroll := float32(500 - 300) // contentHeight - viewportHeight
	if scrollY != maxScroll {
		t.Errorf("scrollY = %v, want %v (clamped to max)", scrollY, maxScroll)
	}
}

func TestWheelEvent_HorizontalDirection(t *testing.T) {
	content := &mockWidget{preferredSize: geometry.Sz(2000, 200)}
	sv := New(content, DirectionOpt(Horizontal))
	ctx := widget.NewContext()
	constraints := geometry.Loose(geometry.Sz(300, 200))

	sv.Layout(ctx, constraints)
	sv.SetBounds(geometry.NewRect(0, 0, 300, 200))

	// Horizontal scroll right (positive delta X).
	e := event.NewWheelEvent(
		geometry.Pt(1, 0),
		geometry.Pt(100, 100),
		geometry.Pt(100, 100),
		0,
	)

	consumed := sv.Event(ctx, e)
	if !consumed {
		t.Error("horizontal wheel event should be consumed")
	}

	scrollX, _ := sv.ScrollOffset()
	if scrollX != defaultScrollStep {
		t.Errorf("scrollX = %v, want %v", scrollX, defaultScrollStep)
	}
}

func TestWheelEvent_VerticalIgnoresHorizontalDelta(t *testing.T) {
	content := &mockWidget{preferredSize: geometry.Sz(200, 1000)}
	sv := New(content, DirectionOpt(Vertical))
	ctx := widget.NewContext()
	constraints := geometry.Loose(geometry.Sz(200, 300))

	sv.Layout(ctx, constraints)
	sv.SetBounds(geometry.NewRect(0, 0, 200, 300))

	// Only horizontal delta -- should be ignored by vertical scrollview.
	e := event.NewWheelEvent(
		geometry.Pt(-1, 0),
		geometry.Pt(100, 150),
		geometry.Pt(100, 150),
		0,
	)

	consumed := sv.Event(ctx, e)
	if consumed {
		t.Error("horizontal-only wheel event should not be consumed by vertical scrollview")
	}
}

func TestWheelEvent_OutsideBounds(t *testing.T) {
	content := &mockWidget{preferredSize: geometry.Sz(200, 1000)}
	sv := New(content)
	ctx := widget.NewContext()
	constraints := geometry.Loose(geometry.Sz(200, 300))

	sv.Layout(ctx, constraints)
	sv.SetBounds(geometry.NewRect(0, 0, 200, 300))

	// Event outside bounds.
	e := event.NewWheelEvent(
		geometry.Pt(0, -1),
		geometry.Pt(500, 500), // Outside
		geometry.Pt(500, 500),
		0,
	)

	consumed := sv.Event(ctx, e)
	if consumed {
		t.Error("wheel event outside bounds should not be consumed")
	}
}

func TestWheelEvent_CustomScrollStep(t *testing.T) {
	content := &mockWidget{preferredSize: geometry.Sz(200, 1000)}
	sv := New(content, ScrollStep(20))
	ctx := widget.NewContext()
	constraints := geometry.Loose(geometry.Sz(200, 300))

	sv.Layout(ctx, constraints)
	sv.SetBounds(geometry.NewRect(0, 0, 200, 300))

	e := event.NewWheelEvent(
		geometry.Pt(0, 1),
		geometry.Pt(100, 150),
		geometry.Pt(100, 150),
		0,
	)

	sv.Event(ctx, e)

	_, scrollY := sv.ScrollOffset()
	if scrollY != 20 {
		t.Errorf("scrollY = %v, want 20 (custom step)", scrollY)
	}
}

// --- Keyboard Tests ---

func TestKeyEvent_ArrowDown(t *testing.T) {
	content := &mockWidget{preferredSize: geometry.Sz(200, 1000)}
	sv := New(content)
	ctx := widget.NewContext()
	constraints := geometry.Loose(geometry.Sz(200, 300))

	sv.Layout(ctx, constraints)
	sv.SetBounds(geometry.NewRect(0, 0, 200, 300))
	sv.SetFocused(true)

	e := event.NewKeyEvent(event.KeyPress, event.KeyDown, 0, 0)
	consumed := sv.Event(ctx, e)

	if !consumed {
		t.Error("KeyDown should be consumed")
	}

	_, scrollY := sv.ScrollOffset()
	if scrollY != defaultScrollStep {
		t.Errorf("scrollY = %v, want %v", scrollY, defaultScrollStep)
	}
}

func TestKeyEvent_ArrowUp(t *testing.T) {
	content := &mockWidget{preferredSize: geometry.Sz(200, 1000)}
	sv := New(content, ScrollY(100))
	ctx := widget.NewContext()
	constraints := geometry.Loose(geometry.Sz(200, 300))

	sv.Layout(ctx, constraints)
	sv.SetBounds(geometry.NewRect(0, 0, 200, 300))
	sv.SetFocused(true)

	e := event.NewKeyEvent(event.KeyPress, event.KeyUp, 0, 0)
	sv.Event(ctx, e)

	_, scrollY := sv.ScrollOffset()
	expected := float32(100) - defaultScrollStep
	if scrollY != expected {
		t.Errorf("scrollY = %v, want %v", scrollY, expected)
	}
}

func TestKeyEvent_PageDown(t *testing.T) {
	content := &mockWidget{preferredSize: geometry.Sz(200, 1000)}
	sv := New(content)
	ctx := widget.NewContext()
	constraints := geometry.Loose(geometry.Sz(200, 300))

	sv.Layout(ctx, constraints)
	sv.SetBounds(geometry.NewRect(0, 0, 200, 300))
	sv.SetFocused(true)

	e := event.NewKeyEvent(event.KeyPress, event.KeyPageDown, 0, 0)
	sv.Event(ctx, e)

	_, scrollY := sv.ScrollOffset()
	if scrollY != 300 { // viewport height
		t.Errorf("scrollY = %v, want 300 (page down = viewport height)", scrollY)
	}
}

func TestKeyEvent_PageUp(t *testing.T) {
	content := &mockWidget{preferredSize: geometry.Sz(200, 1000)}
	sv := New(content, ScrollY(500))
	ctx := widget.NewContext()
	constraints := geometry.Loose(geometry.Sz(200, 300))

	sv.Layout(ctx, constraints)
	sv.SetBounds(geometry.NewRect(0, 0, 200, 300))
	sv.SetFocused(true)

	e := event.NewKeyEvent(event.KeyPress, event.KeyPageUp, 0, 0)
	sv.Event(ctx, e)

	_, scrollY := sv.ScrollOffset()
	if scrollY != 200 { // 500 - 300
		t.Errorf("scrollY = %v, want 200 (500 - viewport height 300)", scrollY)
	}
}

func TestKeyEvent_Home(t *testing.T) {
	content := &mockWidget{preferredSize: geometry.Sz(200, 1000)}
	sv := New(content, ScrollY(500))
	ctx := widget.NewContext()
	constraints := geometry.Loose(geometry.Sz(200, 300))

	sv.Layout(ctx, constraints)
	sv.SetBounds(geometry.NewRect(0, 0, 200, 300))
	sv.SetFocused(true)

	e := event.NewKeyEvent(event.KeyPress, event.KeyHome, 0, 0)
	sv.Event(ctx, e)

	_, scrollY := sv.ScrollOffset()
	if scrollY != 0 {
		t.Errorf("scrollY = %v, want 0 (Home)", scrollY)
	}
}

func TestKeyEvent_End(t *testing.T) {
	content := &mockWidget{preferredSize: geometry.Sz(200, 1000)}
	sv := New(content)
	ctx := widget.NewContext()
	constraints := geometry.Loose(geometry.Sz(200, 300))

	sv.Layout(ctx, constraints)
	sv.SetBounds(geometry.NewRect(0, 0, 200, 300))
	sv.SetFocused(true)

	e := event.NewKeyEvent(event.KeyPress, event.KeyEnd, 0, 0)
	sv.Event(ctx, e)

	_, scrollY := sv.ScrollOffset()
	if scrollY != 700 { // 1000 - 300
		t.Errorf("scrollY = %v, want 700 (End = maxScroll)", scrollY)
	}
}

func TestKeyEvent_NotFocused(t *testing.T) {
	content := &mockWidget{preferredSize: geometry.Sz(200, 1000)}
	sv := New(content)
	ctx := widget.NewContext()
	constraints := geometry.Loose(geometry.Sz(200, 300))

	sv.Layout(ctx, constraints)
	sv.SetBounds(geometry.NewRect(0, 0, 200, 300))
	// Not focused.

	e := event.NewKeyEvent(event.KeyPress, event.KeyDown, 0, 0)
	consumed := sv.Event(ctx, e)

	if consumed {
		t.Error("key events should not be consumed when not focused")
	}
}

func TestKeyEvent_KeyRelease(t *testing.T) {
	content := &mockWidget{preferredSize: geometry.Sz(200, 1000)}
	sv := New(content)
	ctx := widget.NewContext()
	constraints := geometry.Loose(geometry.Sz(200, 300))

	sv.Layout(ctx, constraints)
	sv.SetBounds(geometry.NewRect(0, 0, 200, 300))
	sv.SetFocused(true)

	e := event.NewKeyEvent(event.KeyRelease, event.KeyDown, 0, 0)
	consumed := sv.Event(ctx, e)

	if consumed {
		t.Error("key release events should not be consumed")
	}
}

func TestKeyEvent_ArrowRight(t *testing.T) {
	content := &mockWidget{preferredSize: geometry.Sz(2000, 1000)}
	sv := New(content, DirectionOpt(Both))
	ctx := widget.NewContext()
	constraints := geometry.Loose(geometry.Sz(200, 300))

	sv.Layout(ctx, constraints)
	sv.SetBounds(geometry.NewRect(0, 0, 200, 300))
	sv.SetFocused(true)

	e := event.NewKeyEvent(event.KeyPress, event.KeyRight, 0, 0)
	consumed := sv.Event(ctx, e)

	if !consumed {
		t.Error("KeyRight should be consumed")
	}

	scrollX, _ := sv.ScrollOffset()
	if scrollX != defaultScrollStep {
		t.Errorf("scrollX = %v, want %v", scrollX, defaultScrollStep)
	}
}

func TestKeyEvent_ArrowLeft(t *testing.T) {
	content := &mockWidget{preferredSize: geometry.Sz(2000, 1000)}
	sv := New(content, DirectionOpt(Both), ScrollX(100))
	ctx := widget.NewContext()
	constraints := geometry.Loose(geometry.Sz(200, 300))

	sv.Layout(ctx, constraints)
	sv.SetBounds(geometry.NewRect(0, 0, 200, 300))
	sv.SetFocused(true)

	e := event.NewKeyEvent(event.KeyPress, event.KeyLeft, 0, 0)
	sv.Event(ctx, e)

	scrollX, _ := sv.ScrollOffset()
	expected := float32(100) - defaultScrollStep
	if scrollX != expected {
		t.Errorf("scrollX = %v, want %v", scrollX, expected)
	}
}

func TestKeyEvent_UnknownKey(t *testing.T) {
	content := &mockWidget{preferredSize: geometry.Sz(200, 1000)}
	sv := New(content)
	ctx := widget.NewContext()
	constraints := geometry.Loose(geometry.Sz(200, 300))

	sv.Layout(ctx, constraints)
	sv.SetBounds(geometry.NewRect(0, 0, 200, 300))
	sv.SetFocused(true)

	e := event.NewKeyEvent(event.KeyPress, event.KeyA, 0, 0)
	consumed := sv.Event(ctx, e)

	if consumed {
		t.Error("unknown key should not be consumed")
	}
}

// --- Signal Binding Tests ---

func TestScrollYSignal_TwoWay(t *testing.T) {
	sig := state.NewSignal[float32](0)
	content := &mockWidget{preferredSize: geometry.Sz(200, 1000)}
	sv := New(content, ScrollYSignal(sig))
	ctx := widget.NewContext()
	constraints := geometry.Loose(geometry.Sz(200, 300))

	sv.Layout(ctx, constraints)
	sv.SetBounds(geometry.NewRect(0, 0, 200, 300))

	// User scrolls down -> signal updated.
	e := event.NewWheelEvent(
		geometry.Pt(0, 1),
		geometry.Pt(100, 150),
		geometry.Pt(100, 150),
		0,
	)
	sv.Event(ctx, e)

	if sig.Get() != defaultScrollStep {
		t.Errorf("signal = %v, want %v (two-way sync)", sig.Get(), defaultScrollStep)
	}

	// Signal changes -> scroll updated.
	sig.Set(200)
	_, scrollY := sv.ScrollOffset()
	if scrollY != 200 {
		t.Errorf("scrollY = %v, want 200 (read from signal)", scrollY)
	}
}

func TestScrollXSignal_TwoWay(t *testing.T) {
	sig := state.NewSignal[float32](0)
	content := &mockWidget{preferredSize: geometry.Sz(2000, 200)}
	sv := New(content, DirectionOpt(Horizontal), ScrollXSignal(sig))
	ctx := widget.NewContext()
	constraints := geometry.Loose(geometry.Sz(300, 200))

	sv.Layout(ctx, constraints)
	sv.SetBounds(geometry.NewRect(0, 0, 300, 200))

	// User scrolls right.
	e := event.NewWheelEvent(
		geometry.Pt(1, 0),
		geometry.Pt(100, 100),
		geometry.Pt(100, 100),
		0,
	)
	sv.Event(ctx, e)

	if sig.Get() != defaultScrollStep {
		t.Errorf("signal = %v, want %v", sig.Get(), defaultScrollStep)
	}
}

// --- OnScroll Callback Tests ---

func TestOnScroll_Callback(t *testing.T) {
	var gotX, gotY float32
	content := &mockWidget{preferredSize: geometry.Sz(200, 1000)}
	sv := New(content, OnScroll(func(x, y float32) {
		gotX, gotY = x, y
	}))
	ctx := widget.NewContext()
	constraints := geometry.Loose(geometry.Sz(200, 300))

	sv.Layout(ctx, constraints)
	sv.SetBounds(geometry.NewRect(0, 0, 200, 300))

	e := event.NewWheelEvent(
		geometry.Pt(0, 1),
		geometry.Pt(100, 150),
		geometry.Pt(100, 150),
		0,
	)
	sv.Event(ctx, e)

	if gotX != 0 || gotY != defaultScrollStep {
		t.Errorf("onScroll got (%v, %v), want (0, %v)", gotX, gotY, defaultScrollStep)
	}
}

// --- Mount/Unmount Tests ---

func TestMount_WithSignals(t *testing.T) {
	sig := state.NewSignal[float32](0)
	sv := New(&mockWidget{}, ScrollYSignal(sig))

	// Create a mock scheduler.
	sched := &mockScheduler{}
	ctx := widget.NewContext()
	ctx.SetScheduler(sched)

	sv.Mount(ctx)

	// Verify bindings were created.
	// Change signal -> should mark dirty.
	sig.Set(50)

	// The binding should have been registered.
	// We verify indirectly: unmount should clean up.
	sv.CleanupBindings()
	sv.Unmount()
}

func TestMount_NilScheduler(t *testing.T) {
	sig := state.NewSignal[float32](0)
	sv := New(&mockWidget{}, ScrollYSignal(sig))
	ctx := widget.NewContext()
	// No scheduler set.

	sv.Mount(ctx) // Should not panic.
}

func TestMount_ReadonlySignals(t *testing.T) {
	xComputed := state.NewComputed(func() float32 { return 10 })
	yComputed := state.NewComputed(func() float32 { return 20 })
	sv := New(&mockWidget{},
		ScrollXReadonlySignal(xComputed),
		ScrollYReadonlySignal(yComputed),
	)

	sched := &mockScheduler{}
	ctx := widget.NewContext()
	ctx.SetScheduler(sched)

	sv.Mount(ctx) // Should not panic.
	sv.CleanupBindings()
	sv.Unmount()
}

// --- Scroll Bounds Tests ---

func TestClampScroll(t *testing.T) {
	tests := []struct {
		name         string
		offset       float32
		contentSize  float32
		viewportSize float32
		want         float32
	}{
		{"zero offset", 0, 1000, 300, 0},
		{"negative offset clamped to zero", -10, 1000, 300, 0},
		{"within bounds", 100, 1000, 300, 100},
		{"exceeds max clamped", 800, 1000, 300, 700},
		{"content smaller than viewport", 50, 200, 300, 0},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			got := clampScroll(tc.offset, tc.contentSize, tc.viewportSize)
			if got != tc.want {
				t.Errorf("clampScroll(%v, %v, %v) = %v, want %v",
					tc.offset, tc.contentSize, tc.viewportSize, got, tc.want)
			}
		})
	}
}

// --- Thumb Sizing Tests ---

func TestComputeThumbSize(t *testing.T) {
	tests := []struct {
		name         string
		viewportSize float32
		contentSize  float32
		trackLen     float32
		wantMin      float32
	}{
		{"normal ratio", 300, 1000, 280, minThumbSize},
		{"content equals viewport", 300, 300, 280, 280}, // ratio = 1, thumb = track
		{"content smaller than viewport", 300, 100, 280, 280},
		{"very large content", 300, 10000, 280, minThumbSize},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			got := computeThumbSize(tc.viewportSize, tc.contentSize, tc.trackLen)
			if got < tc.wantMin {
				t.Errorf("computeThumbSize() = %v, want >= %v", got, tc.wantMin)
			}
		})
	}
}

func TestComputeThumbPosition(t *testing.T) {
	tests := []struct {
		name      string
		offset    float32
		maxScroll float32
		trackLen  float32
		thumbSize float32
		want      float32
	}{
		{"at start", 0, 700, 280, 84, 0},
		{"at end", 700, 700, 280, 84, 196},
		{"halfway", 350, 700, 280, 84, 98},
		{"zero maxScroll", 0, 0, 280, 84, 0},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			got := computeThumbPosition(tc.offset, tc.maxScroll, tc.trackLen, tc.thumbSize)
			if got != tc.want {
				t.Errorf("computeThumbPosition() = %v, want %v", got, tc.want)
			}
		})
	}
}

// --- Scrollbar Visibility Tests ---

func TestShouldShowVScrollbar(t *testing.T) {
	tests := []struct {
		name        string
		direction   ScrollDirection
		visibility  ScrollbarVisibility
		contentH    float32
		viewportH   float32
		wantVisible bool
	}{
		{"auto with overflow", Vertical, ScrollbarAuto, 1000, 300, true},
		{"auto no overflow", Vertical, ScrollbarAuto, 200, 300, false},
		{"always", Vertical, ScrollbarAlways, 200, 300, true},
		{"never", Vertical, ScrollbarNever, 1000, 300, false},
		{"horizontal direction", Horizontal, ScrollbarAuto, 1000, 300, false},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			sv := New(&mockWidget{preferredSize: geometry.Sz(200, tc.contentH)},
				DirectionOpt(tc.direction),
				ScrollbarOpt(tc.visibility),
			)
			ctx := widget.NewContext()
			sv.Layout(ctx, geometry.Loose(geometry.Sz(200, tc.viewportH)))

			got := sv.shouldShowVScrollbar()
			if got != tc.wantVisible {
				t.Errorf("shouldShowVScrollbar() = %v, want %v", got, tc.wantVisible)
			}
		})
	}
}

func TestShouldShowHScrollbar(t *testing.T) {
	tests := []struct {
		name        string
		direction   ScrollDirection
		visibility  ScrollbarVisibility
		contentW    float32
		viewportW   float32
		wantVisible bool
	}{
		{"auto with overflow", Horizontal, ScrollbarAuto, 2000, 300, true},
		{"auto no overflow", Horizontal, ScrollbarAuto, 200, 300, false},
		{"always", Horizontal, ScrollbarAlways, 200, 300, true},
		{"never", Horizontal, ScrollbarNever, 2000, 300, false},
		{"vertical direction", Vertical, ScrollbarAuto, 2000, 300, false},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			sv := New(&mockWidget{preferredSize: geometry.Sz(tc.contentW, 200)},
				DirectionOpt(tc.direction),
				ScrollbarOpt(tc.visibility),
			)
			ctx := widget.NewContext()
			sv.Layout(ctx, geometry.Loose(geometry.Sz(tc.viewportW, 200)))

			got := sv.shouldShowHScrollbar()
			if got != tc.wantVisible {
				t.Errorf("shouldShowHScrollbar() = %v, want %v", got, tc.wantVisible)
			}
		})
	}
}

// --- canScrollX/canScrollY Tests ---

func TestCanScroll(t *testing.T) {
	t.Run("vertical can scroll Y", func(t *testing.T) {
		sv := New(&mockWidget{preferredSize: geometry.Sz(200, 1000)}, DirectionOpt(Vertical))
		ctx := widget.NewContext()
		sv.Layout(ctx, geometry.Loose(geometry.Sz(200, 300)))

		if !sv.canScrollY() {
			t.Error("should be able to scroll Y")
		}
		if sv.canScrollX() {
			t.Error("should not be able to scroll X in Vertical mode")
		}
	})

	t.Run("horizontal can scroll X", func(t *testing.T) {
		sv := New(&mockWidget{preferredSize: geometry.Sz(2000, 200)}, DirectionOpt(Horizontal))
		ctx := widget.NewContext()
		sv.Layout(ctx, geometry.Loose(geometry.Sz(300, 200)))

		if !sv.canScrollX() {
			t.Error("should be able to scroll X")
		}
		if sv.canScrollY() {
			t.Error("should not be able to scroll Y in Horizontal mode")
		}
	})
}

// --- Accessor Tests ---

func TestAccessors(t *testing.T) {
	content := &mockWidget{preferredSize: geometry.Sz(200, 1000)}
	sv := New(content, ScrollX(10), ScrollY(20))
	ctx := widget.NewContext()
	constraints := geometry.Loose(geometry.Sz(200, 300))

	sv.Layout(ctx, constraints)

	if sv.Content() != content {
		t.Error("Content() should return content widget")
	}

	x, y := sv.ScrollOffset()
	if x != 10 || y != 20 {
		t.Errorf("ScrollOffset() = (%v, %v), want (10, 20)", x, y)
	}

	vs := sv.ViewportSize()
	if vs.Width != 200 || vs.Height != 300 {
		t.Errorf("ViewportSize() = %v, want (200, 300)", vs)
	}

	cs := sv.ContentSize()
	if cs.Height != 1000 {
		t.Errorf("ContentSize().Height = %v, want 1000", cs.Height)
	}
}

// --- Painter Tests ---

func TestDefaultPainter_PaintScrollbar(t *testing.T) {
	p := DefaultPainter{}
	canvas := &internalMockCanvas{}

	ps := PaintState{
		Bounds:         geometry.NewRect(0, 0, 200, 300),
		VScrollVisible: true,
		VThumbRect:     geometry.NewRect(188, 10, 8, 50),
		VTrackRect:     geometry.NewRect(186, 0, 12, 300),
	}

	p.PaintScrollbar(canvas, ps)

	if len(canvas.drawRoundRects) < 2 {
		t.Errorf("should draw track + thumb, got %d DrawRoundRect calls", len(canvas.drawRoundRects))
	}
}

func TestDefaultPainter_EmptyBounds(t *testing.T) {
	p := DefaultPainter{}
	canvas := &internalMockCanvas{}

	ps := PaintState{
		Bounds: geometry.Rect{},
	}

	p.PaintScrollbar(canvas, ps)

	if len(canvas.drawRoundRects) != 0 {
		t.Error("should not draw anything for empty bounds")
	}
}

func TestDefaultPainter_WithColorScheme(t *testing.T) {
	p := DefaultPainter{}
	canvas := &internalMockCanvas{}

	cs := ScrollbarColorScheme{
		Track:      widget.RGBA(0.1, 0.1, 0.1, 1),
		Thumb:      widget.RGBA(0.5, 0.5, 0.5, 1),
		ThumbHover: widget.RGBA(0.6, 0.6, 0.6, 1),
		ThumbDrag:  widget.RGBA(0.7, 0.7, 0.7, 1),
	}

	ps := PaintState{
		Bounds:         geometry.NewRect(0, 0, 200, 300),
		VScrollVisible: true,
		VThumbRect:     geometry.NewRect(188, 10, 8, 50),
		VTrackRect:     geometry.NewRect(186, 0, 12, 300),
		ColorScheme:    cs,
	}

	p.PaintScrollbar(canvas, ps)

	if len(canvas.drawRoundRects) < 2 {
		t.Error("should draw track + thumb with color scheme")
	}
}

func TestDefaultPainter_HoverState(t *testing.T) {
	p := DefaultPainter{}
	canvas := &internalMockCanvas{}

	ps := PaintState{
		Bounds:         geometry.NewRect(0, 0, 200, 300),
		Hovered:        true,
		VScrollVisible: true,
		VThumbRect:     geometry.NewRect(188, 10, 8, 50),
		VTrackRect:     geometry.NewRect(186, 0, 12, 300),
	}

	p.PaintScrollbar(canvas, ps)

	if len(canvas.drawRoundRects) < 2 {
		t.Error("should draw track + thumb when hovered")
	}
}

func TestDefaultPainter_DragState(t *testing.T) {
	p := DefaultPainter{}
	canvas := &internalMockCanvas{}

	ps := PaintState{
		Bounds:         geometry.NewRect(0, 0, 200, 300),
		Dragging:       true,
		VScrollVisible: true,
		VThumbRect:     geometry.NewRect(188, 10, 8, 50),
		VTrackRect:     geometry.NewRect(186, 0, 12, 300),
	}

	p.PaintScrollbar(canvas, ps)

	if len(canvas.drawRoundRects) < 2 {
		t.Error("should draw track + thumb when dragging")
	}
}

func TestDefaultPainter_BothScrollbars(t *testing.T) {
	p := DefaultPainter{}
	canvas := &internalMockCanvas{}

	ps := PaintState{
		Bounds:         geometry.NewRect(0, 0, 200, 300),
		VScrollVisible: true,
		VThumbRect:     geometry.NewRect(188, 10, 8, 50),
		VTrackRect:     geometry.NewRect(186, 0, 12, 288),
		HScrollVisible: true,
		HThumbRect:     geometry.NewRect(10, 288, 50, 8),
		HTrackRect:     geometry.NewRect(0, 286, 186, 12),
	}

	p.PaintScrollbar(canvas, ps)

	// Track + thumb for both V and H.
	if len(canvas.drawRoundRects) < 4 {
		t.Errorf("should draw 4 rects (2 tracks + 2 thumbs), got %d", len(canvas.drawRoundRects))
	}
}

// --- Compile-Time Interface Checks ---

func TestCompileTimeInterfaces(t *testing.T) {
	// These are also checked via var _ declarations in widget.go,
	// but explicitly test them here for clarity.
	var w interface{} = &Widget{}
	if _, ok := w.(widget.Widget); !ok {
		t.Error("Widget should implement widget.Widget")
	}
	if _, ok := w.(widget.Focusable); !ok {
		t.Error("Widget should implement widget.Focusable")
	}
	if _, ok := w.(widget.Lifecycle); !ok {
		t.Error("Widget should implement widget.Lifecycle")
	}
}

// --- Event Propagation to Content ---

func TestEvent_PropagateToContent(t *testing.T) {
	content := &mockWidget{
		preferredSize: geometry.Sz(200, 1000),
		consumeEvents: true,
	}
	sv := New(content)
	ctx := widget.NewContext()
	constraints := geometry.Loose(geometry.Sz(200, 300))

	sv.Layout(ctx, constraints)
	sv.SetBounds(geometry.NewRect(0, 0, 200, 300))

	e := event.NewKeyEvent(event.KeyPress, event.KeyDown, 0, 0)
	consumed := sv.Event(ctx, e)

	if !consumed {
		t.Error("event should be consumed by content")
	}
	if !content.eventCalled {
		t.Error("content Event() should have been called")
	}
}

// --- scrollFromTrackClick Tests ---

func TestScrollFromTrackClick(t *testing.T) {
	tests := []struct {
		name         string
		clickPos     float32
		trackStart   float32
		trackLen     float32
		viewportSize float32
		contentSize  float32
		wantMin      float32
		wantMax      float32
	}{
		{"click at start", 10, 0, 280, 300, 1000, 0, 100},
		{"click at end", 270, 0, 280, 300, 1000, 500, 700},
		{"click at middle", 140, 0, 280, 300, 1000, 200, 400},
		{"no scroll possible", 140, 0, 280, 300, 200, 0, 0},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			got := scrollFromTrackClick(tc.clickPos, tc.trackStart, tc.trackLen, tc.viewportSize, tc.contentSize)
			if got < tc.wantMin || got > tc.wantMax {
				t.Errorf("scrollFromTrackClick() = %v, want in [%v, %v]", got, tc.wantMin, tc.wantMax)
			}
		})
	}
}

// --- computeScrollbarRect Tests ---

func TestComputeScrollbarRect(t *testing.T) {
	bounds := geometry.NewRect(0, 0, 200, 300)

	t.Run("vertical without horizontal", func(t *testing.T) {
		r := computeScrollbarRect(bounds, dragVertical, false)
		if r.IsEmpty() {
			t.Error("vertical scrollbar rect should not be empty")
		}
		if r.Height() != 300 {
			t.Errorf("height = %v, want 300", r.Height())
		}
	})

	t.Run("vertical with horizontal", func(t *testing.T) {
		r := computeScrollbarRect(bounds, dragVertical, true)
		totalWidth := scrollbarWidth + scrollbarPadding*2
		expectedHeight := float32(300) - totalWidth
		if r.Height() != expectedHeight {
			t.Errorf("height = %v, want %v (reduced for horizontal scrollbar)", r.Height(), expectedHeight)
		}
	})

	t.Run("horizontal without vertical", func(t *testing.T) {
		r := computeScrollbarRect(bounds, dragHorizontal, false)
		if r.IsEmpty() {
			t.Error("horizontal scrollbar rect should not be empty")
		}
		if r.Width() != 200 {
			t.Errorf("width = %v, want 200", r.Width())
		}
	})
}

// --- Padding Method Test ---

func TestPadding(t *testing.T) {
	sv := New(&mockWidget{})
	result := sv.Padding(10)
	if result != sv {
		t.Error("Padding() should return the widget for chaining")
	}
}

// --- Event with Unknown Type ---

func TestEvent_UnknownType(t *testing.T) {
	sv := New(&mockWidget{})
	ctx := widget.NewContext()

	// FocusEvent should not be consumed by scroll view.
	e := event.NewFocusEvent(event.FocusGained)
	consumed := sv.Event(ctx, e)
	if consumed {
		t.Error("unknown event type should not be consumed")
	}
}

// --- Mouse Enter/Leave Tests ---

func TestMouseEnter(t *testing.T) {
	sv := New(&mockWidget{preferredSize: geometry.Sz(200, 1000)})
	ctx := widget.NewContext()
	sv.Layout(ctx, geometry.Loose(geometry.Sz(200, 300)))
	sv.SetBounds(geometry.NewRect(0, 0, 200, 300))

	me := &event.MouseEvent{
		MouseType: event.MouseEnter,
		Position:  geometry.Pt(100, 150),
	}
	sv.Event(ctx, me)

	if !sv.hovered {
		t.Error("should be hovered after MouseEnter")
	}
}

func TestMouseLeave(t *testing.T) {
	sv := New(&mockWidget{preferredSize: geometry.Sz(200, 1000)})
	ctx := widget.NewContext()
	sv.Layout(ctx, geometry.Loose(geometry.Sz(200, 300)))
	sv.SetBounds(geometry.NewRect(0, 0, 200, 300))

	sv.hovered = true
	me := &event.MouseEvent{
		MouseType: event.MouseLeave,
		Position:  geometry.Pt(100, 150),
	}
	sv.Event(ctx, me)

	if sv.hovered {
		t.Error("should not be hovered after MouseLeave")
	}
}

// --- End Scrollbar Clamp Tests ---

func TestKeyEvent_EndWithSmallContent(t *testing.T) {
	content := &mockWidget{preferredSize: geometry.Sz(200, 100)}
	sv := New(content)
	ctx := widget.NewContext()
	sv.Layout(ctx, geometry.Loose(geometry.Sz(200, 300)))
	sv.SetBounds(geometry.NewRect(0, 0, 200, 300))
	sv.SetFocused(true)

	e := event.NewKeyEvent(event.KeyPress, event.KeyEnd, 0, 0)
	sv.Event(ctx, e)

	_, scrollY := sv.ScrollOffset()
	if scrollY != 0 {
		t.Errorf("scrollY = %v, want 0 (content fits in viewport)", scrollY)
	}
}

// --- Mouse Scrollbar Drag Tests ---

func TestMousePress_OnVThumb(t *testing.T) {
	content := &mockWidget{preferredSize: geometry.Sz(200, 1000)}
	sv := New(content, ScrollbarOpt(ScrollbarAlways))
	ctx := widget.NewContext()
	constraints := geometry.Loose(geometry.Sz(200, 300))

	sv.Layout(ctx, constraints)
	sv.SetBounds(geometry.NewRect(0, 0, 200, 300))

	// Get the thumb rect to know where to click.
	vThumb, _ := sv.computeThumbRects()

	// Click on the vertical thumb.
	me := &event.MouseEvent{
		MouseType: event.MousePress,
		Button:    event.ButtonLeft,
		Position:  geometry.Pt(vThumb.Min.X+1, vThumb.Min.Y+1),
	}

	consumed := sv.Event(ctx, me)
	if !consumed {
		t.Error("press on vertical thumb should be consumed")
	}
	if sv.dragging != dragVertical {
		t.Errorf("dragging = %v, want dragVertical", sv.dragging)
	}
}

func TestMouseRelease_AfterDrag(t *testing.T) {
	content := &mockWidget{preferredSize: geometry.Sz(200, 1000)}
	sv := New(content, ScrollbarOpt(ScrollbarAlways))
	ctx := widget.NewContext()
	constraints := geometry.Loose(geometry.Sz(200, 300))

	sv.Layout(ctx, constraints)
	sv.SetBounds(geometry.NewRect(0, 0, 200, 300))

	// Simulate dragging state.
	sv.dragging = dragVertical

	me := &event.MouseEvent{
		MouseType: event.MouseRelease,
		Button:    event.ButtonLeft,
		Position:  geometry.Pt(190, 150),
	}

	consumed := sv.Event(ctx, me)
	if !consumed {
		t.Error("release after drag should be consumed")
	}
	if sv.dragging != dragNone {
		t.Errorf("dragging = %v, want dragNone", sv.dragging)
	}
}

func TestMouseRelease_NotDragging(t *testing.T) {
	sv := New(&mockWidget{preferredSize: geometry.Sz(200, 1000)})
	ctx := widget.NewContext()
	sv.Layout(ctx, geometry.Loose(geometry.Sz(200, 300)))
	sv.SetBounds(geometry.NewRect(0, 0, 200, 300))

	me := &event.MouseEvent{
		MouseType: event.MouseRelease,
		Button:    event.ButtonLeft,
		Position:  geometry.Pt(100, 150),
	}

	consumed := sv.Event(ctx, me)
	if consumed {
		t.Error("release without drag should not be consumed")
	}
}

func TestMouseRelease_RightButton(t *testing.T) {
	sv := New(&mockWidget{preferredSize: geometry.Sz(200, 1000)})
	ctx := widget.NewContext()
	sv.Layout(ctx, geometry.Loose(geometry.Sz(200, 300)))
	sv.SetBounds(geometry.NewRect(0, 0, 200, 300))

	me := &event.MouseEvent{
		MouseType: event.MouseRelease,
		Button:    event.ButtonRight,
		Position:  geometry.Pt(100, 150),
	}

	consumed := sv.Event(ctx, me)
	if consumed {
		t.Error("right button release should not be consumed")
	}
}

func TestMouseMove_Dragging(t *testing.T) {
	content := &mockWidget{preferredSize: geometry.Sz(200, 1000)}
	sv := New(content, ScrollbarOpt(ScrollbarAlways))
	ctx := widget.NewContext()
	constraints := geometry.Loose(geometry.Sz(200, 300))

	sv.Layout(ctx, constraints)
	sv.SetBounds(geometry.NewRect(0, 0, 200, 300))

	// Set up vertical drag state.
	sv.dragging = dragVertical
	sv.dragStart = geometry.Pt(195, 50)
	sv.dragScrollStart = 0

	me := &event.MouseEvent{
		MouseType: event.MouseMove,
		Buttons:   event.ButtonStateLeft,
		Position:  geometry.Pt(195, 100), // 50px delta
	}

	consumed := sv.Event(ctx, me)
	if !consumed {
		t.Error("move during drag should be consumed")
	}

	_, scrollY := sv.ScrollOffset()
	if scrollY <= 0 {
		t.Errorf("scrollY = %v, should be > 0 after vertical drag", scrollY)
	}
}

func TestMouseMove_NotDragging(t *testing.T) {
	sv := New(&mockWidget{preferredSize: geometry.Sz(200, 1000)})
	ctx := widget.NewContext()
	sv.Layout(ctx, geometry.Loose(geometry.Sz(200, 300)))
	sv.SetBounds(geometry.NewRect(0, 0, 200, 300))

	me := &event.MouseEvent{
		MouseType: event.MouseMove,
		Position:  geometry.Pt(100, 150),
	}

	consumed := sv.Event(ctx, me)
	if consumed {
		t.Error("move without drag should not be consumed")
	}
}

func TestMousePress_RightButton(t *testing.T) {
	content := &mockWidget{preferredSize: geometry.Sz(200, 1000)}
	sv := New(content, ScrollbarOpt(ScrollbarAlways))
	ctx := widget.NewContext()
	sv.Layout(ctx, geometry.Loose(geometry.Sz(200, 300)))
	sv.SetBounds(geometry.NewRect(0, 0, 200, 300))

	me := &event.MouseEvent{
		MouseType: event.MousePress,
		Button:    event.ButtonRight,
		Position:  geometry.Pt(195, 50),
	}

	consumed := sv.Event(ctx, me)
	if consumed {
		t.Error("right button press should not be consumed")
	}
}

func TestMousePress_OnVTrack(t *testing.T) {
	content := &mockWidget{preferredSize: geometry.Sz(200, 1000)}
	sv := New(content, ScrollbarOpt(ScrollbarAlways))
	ctx := widget.NewContext()
	constraints := geometry.Loose(geometry.Sz(200, 300))

	sv.Layout(ctx, constraints)
	sv.SetBounds(geometry.NewRect(0, 0, 200, 300))

	vTrack, _ := sv.computeTrackRects()

	// Click on the track below the thumb.
	me := &event.MouseEvent{
		MouseType: event.MousePress,
		Button:    event.ButtonLeft,
		Position:  geometry.Pt(vTrack.Min.X+1, vTrack.Max.Y-10),
	}

	consumed := sv.Event(ctx, me)
	if !consumed {
		t.Error("press on vertical track should be consumed")
	}

	_, scrollY := sv.ScrollOffset()
	if scrollY <= 0 {
		t.Errorf("scrollY = %v, should be > 0 after track click", scrollY)
	}
}

func TestMouseMove_HorizontalDrag(t *testing.T) {
	content := &mockWidget{preferredSize: geometry.Sz(2000, 200)}
	sv := New(content, DirectionOpt(Horizontal), ScrollbarOpt(ScrollbarAlways))
	ctx := widget.NewContext()
	constraints := geometry.Loose(geometry.Sz(300, 200))

	sv.Layout(ctx, constraints)
	sv.SetBounds(geometry.NewRect(0, 0, 300, 200))

	// Set up horizontal drag state.
	sv.dragging = dragHorizontal
	sv.dragStart = geometry.Pt(50, 195)
	sv.dragScrollStart = 0

	me := &event.MouseEvent{
		MouseType: event.MouseMove,
		Buttons:   event.ButtonStateLeft,
		Position:  geometry.Pt(100, 195), // 50px delta
	}

	consumed := sv.Event(ctx, me)
	if !consumed {
		t.Error("horizontal move during drag should be consumed")
	}

	scrollX, _ := sv.ScrollOffset()
	if scrollX <= 0 {
		t.Errorf("scrollX = %v, should be > 0 after horizontal drag", scrollX)
	}
}

func TestMousePress_OnHThumb(t *testing.T) {
	content := &mockWidget{preferredSize: geometry.Sz(2000, 200)}
	sv := New(content, DirectionOpt(Horizontal), ScrollbarOpt(ScrollbarAlways))
	ctx := widget.NewContext()
	constraints := geometry.Loose(geometry.Sz(300, 200))

	sv.Layout(ctx, constraints)
	sv.SetBounds(geometry.NewRect(0, 0, 300, 200))

	_, hThumb := sv.computeThumbRects()

	if hThumb.IsEmpty() {
		t.Skip("horizontal thumb is empty -- content may not overflow")
	}

	me := &event.MouseEvent{
		MouseType: event.MousePress,
		Button:    event.ButtonLeft,
		Position:  geometry.Pt(hThumb.Min.X+1, hThumb.Min.Y+1),
	}

	consumed := sv.Event(ctx, me)
	if !consumed {
		t.Error("press on horizontal thumb should be consumed")
	}
	if sv.dragging != dragHorizontal {
		t.Errorf("dragging = %v, want dragHorizontal", sv.dragging)
	}
}

func TestMousePress_OnHTrack(t *testing.T) {
	content := &mockWidget{preferredSize: geometry.Sz(2000, 200)}
	sv := New(content, DirectionOpt(Horizontal), ScrollbarOpt(ScrollbarAlways))
	ctx := widget.NewContext()
	constraints := geometry.Loose(geometry.Sz(300, 200))

	sv.Layout(ctx, constraints)
	sv.SetBounds(geometry.NewRect(0, 0, 300, 200))

	_, hTrack := sv.computeTrackRects()
	if hTrack.IsEmpty() {
		t.Skip("horizontal track is empty")
	}

	me := &event.MouseEvent{
		MouseType: event.MousePress,
		Button:    event.ButtonLeft,
		Position:  geometry.Pt(hTrack.Max.X-10, hTrack.Min.Y+1),
	}

	consumed := sv.Event(ctx, me)
	if !consumed {
		t.Error("press on horizontal track should be consumed")
	}

	scrollX, _ := sv.ScrollOffset()
	if scrollX <= 0 {
		t.Errorf("scrollX = %v, should be > 0 after horizontal track click", scrollX)
	}
}

// --- computeHThumbRect Tests ---

func TestComputeHThumbRect(t *testing.T) {
	content := &mockWidget{preferredSize: geometry.Sz(2000, 200)}
	sv := New(content, DirectionOpt(Horizontal), ScrollbarOpt(ScrollbarAlways))
	ctx := widget.NewContext()
	constraints := geometry.Loose(geometry.Sz(300, 200))

	sv.Layout(ctx, constraints)
	sv.SetBounds(geometry.NewRect(0, 0, 300, 200))

	_, hThumb := sv.computeThumbRects()
	if hThumb.IsEmpty() {
		t.Error("horizontal thumb should not be empty when content overflows")
	}
	if hThumb.Width() < minThumbSize {
		t.Errorf("thumb width = %v, should be >= %v", hThumb.Width(), minThumbSize)
	}
}

// --- computeThumbSize edge cases ---

func TestComputeThumbSize_ZeroContent(t *testing.T) {
	got := computeThumbSize(300, 0, 280)
	if got != minThumbSize {
		t.Errorf("computeThumbSize(300, 0, 280) = %v, want %v", got, minThumbSize)
	}
}

func TestComputeThumbSize_ZeroViewport(t *testing.T) {
	got := computeThumbSize(0, 1000, 280)
	if got != minThumbSize {
		t.Errorf("computeThumbSize(0, 1000, 280) = %v, want %v", got, minThumbSize)
	}
}

// --- computeThumbPosition edge cases ---

func TestComputeThumbPosition_ZeroScrollableTrack(t *testing.T) {
	// thumbSize == trackLen -> scrollableTrack = 0
	got := computeThumbPosition(100, 700, 280, 280)
	if got != 0 {
		t.Errorf("computeThumbPosition() = %v, want 0 (zero scrollable track)", got)
	}
}

// --- Mouse with unknown MouseType ---

func TestMouseEvent_UnknownType(t *testing.T) {
	sv := New(&mockWidget{preferredSize: geometry.Sz(200, 1000)})
	ctx := widget.NewContext()
	sv.Layout(ctx, geometry.Loose(geometry.Sz(200, 300)))
	sv.SetBounds(geometry.NewRect(0, 0, 200, 300))

	me := &event.MouseEvent{
		MouseType: event.MouseEventType(99), // Unknown
		Position:  geometry.Pt(100, 150),
	}

	consumed := sv.Event(ctx, me)
	if consumed {
		t.Error("unknown mouse type should not be consumed")
	}
}

// --- Unmount Test ---

func TestUnmount_Internal(t *testing.T) {
	sv := New(&mockWidget{})
	sv.Unmount() // Should not panic.
}

// --- buildContentConstraints default branch ---

func TestBuildContentConstraints_Default(t *testing.T) {
	sv := New(&mockWidget{preferredSize: geometry.Sz(200, 1000)})
	sv.viewportSize = geometry.Sz(200, 300)

	// Set an invalid direction to test default branch.
	sv.cfg.direction = ScrollDirection(99)
	c := sv.buildContentConstraints()

	// Default should behave like Vertical.
	if c.MaxHeight < geometry.Infinity {
		t.Error("default direction should have unconstrained height")
	}
}

// --- Test Helpers ---

// mockWidget is a minimal widget.Widget implementation for testing.
type mockWidget struct {
	widget.WidgetBase
	preferredSize geometry.Size
	consumeEvents bool
	eventCalled   bool
}

func (m *mockWidget) Layout(_ widget.Context, constraints geometry.Constraints) geometry.Size {
	return constraints.Constrain(m.preferredSize)
}

func (m *mockWidget) Draw(_ widget.Context, _ widget.Canvas) {}

func (m *mockWidget) Event(_ widget.Context, _ event.Event) bool {
	m.eventCalled = true
	return m.consumeEvents
}

func (m *mockWidget) Children() []widget.Widget { return nil }

// testPainter records PaintScrollbar calls.
type testPainter struct {
	paintCalled bool
	lastState   PaintState
}

func (p *testPainter) PaintScrollbar(_ widget.Canvas, ps PaintState) {
	p.paintCalled = true
	p.lastState = ps
}

// mockScheduler implements widget.SchedulerRef for testing.
type mockScheduler struct {
	dirtyWidgets []widget.Widget
}

func (s *mockScheduler) MarkDirty(w widget.Widget) {
	s.dirtyWidgets = append(s.dirtyWidgets, w)
}

// --- internalMockCanvas records canvas calls for testing ---

type internalMockCanvas struct {
	drawRoundRects      []internalDrawRoundRectCall
	pushClipCount       int
	popClipCount        int
	pushTransformCount  int
	popTransformCount   int
	lastTransformOffset geometry.Point
	lastClipRect        geometry.Rect
}

type internalDrawRoundRectCall struct {
	r      geometry.Rect
	color  widget.Color
	radius float32
}

func (c *internalMockCanvas) Clear(_ widget.Color)                                  {}
func (c *internalMockCanvas) DrawRect(_ geometry.Rect, _ widget.Color)              {}
func (c *internalMockCanvas) StrokeRect(_ geometry.Rect, _ widget.Color, _ float32) {}

func (c *internalMockCanvas) DrawRoundRect(r geometry.Rect, color widget.Color, radius float32) {
	c.drawRoundRects = append(c.drawRoundRects, internalDrawRoundRectCall{r: r, color: color, radius: radius})
}

func (c *internalMockCanvas) StrokeRoundRect(_ geometry.Rect, _ widget.Color, _ float32, _ float32) {
}
func (c *internalMockCanvas) DrawCircle(_ geometry.Point, _ float32, _ widget.Color)              {}
func (c *internalMockCanvas) StrokeCircle(_ geometry.Point, _ float32, _ widget.Color, _ float32) {}
func (c *internalMockCanvas) DrawLine(_, _ geometry.Point, _ widget.Color, _ float32)             {}
func (c *internalMockCanvas) DrawText(_ string, _ geometry.Rect, _ float32, _ widget.Color, _ bool, _ widget.TextAlign) {
}

func (c *internalMockCanvas) MeasureText(text string, fontSize float32, _ bool) float32 {
	return float32(len([]rune(text))) * fontSize * 0.5
}
func (c *internalMockCanvas) DrawImage(_ image.Image, _ geometry.Point) {}

func (c *internalMockCanvas) PushClip(r geometry.Rect) {
	c.pushClipCount++
	c.lastClipRect = r
}
func (c *internalMockCanvas) PushClipRoundRect(_ geometry.Rect, _ float32) {}

func (c *internalMockCanvas) PopClip() {
	c.popClipCount++
}

func (c *internalMockCanvas) PushTransform(offset geometry.Point) {
	c.pushTransformCount++
	c.lastTransformOffset = offset
}

func (c *internalMockCanvas) PopTransform() {
	c.popTransformCount++
}

func (c *internalMockCanvas) TransformOffset() geometry.Point { return geometry.Point{} }
