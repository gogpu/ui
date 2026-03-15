package collapsible

import (
	"image"
	"testing"
	"time"

	"github.com/gogpu/ui/event"
	"github.com/gogpu/ui/geometry"
	"github.com/gogpu/ui/state"
	"github.com/gogpu/ui/widget"
)

// --- Config Tests ---

func TestConfig_ResolvedTitle(t *testing.T) {
	t.Run("static title", func(t *testing.T) {
		c := config{title: "Hello"}
		if got := c.ResolvedTitle(); got != "Hello" {
			t.Errorf("ResolvedTitle() = %q, want %q", got, "Hello")
		}
	})

	t.Run("dynamic title", func(t *testing.T) {
		c := config{title: "static", titleFn: func() string { return "dynamic" }}
		if got := c.ResolvedTitle(); got != "dynamic" {
			t.Errorf("ResolvedTitle() = %q, want %q", got, "dynamic")
		}
	})

	t.Run("empty title", func(t *testing.T) {
		c := config{}
		if got := c.ResolvedTitle(); got != "" {
			t.Errorf("ResolvedTitle() = %q, want empty", got)
		}
	})
}

func TestConfig_ResolvedExpanded(t *testing.T) {
	t.Run("static expanded", func(t *testing.T) {
		c := config{expanded: true}
		if !c.ResolvedExpanded() {
			t.Error("ResolvedExpanded() should be true")
		}
	})

	t.Run("signal overrides static", func(t *testing.T) {
		sig := state.NewSignal(true)
		c := config{expanded: false, expandedSignal: sig}
		if !c.ResolvedExpanded() {
			t.Error("signal(true) should override static(false)")
		}
	})

	t.Run("readonly signal overrides signal", func(t *testing.T) {
		sig := state.NewSignal(false)
		computed := state.NewComputed(func() bool { return true })
		c := config{expanded: false, expandedSignal: sig, readonlyExpandedSignal: computed}
		if !c.ResolvedExpanded() {
			t.Error("readonly signal(true) should override signal(false)")
		}
	})
}

// --- Options Tests ---

func TestOptions_All(t *testing.T) {
	t.Run("Title", func(t *testing.T) {
		var c config
		Title("Test")(&c)
		if c.title != "Test" {
			t.Errorf("title = %q, want %q", c.title, "Test")
		}
	})

	t.Run("TitleFn", func(t *testing.T) {
		var c config
		TitleFn(func() string { return "fn" })(&c)
		if c.titleFn == nil {
			t.Error("titleFn should not be nil")
		}
	})

	t.Run("Content", func(t *testing.T) {
		var c config
		w := &internalMockWidget{}
		Content(w)(&c)
		if c.content != w {
			t.Error("content should be set")
		}
	})

	t.Run("Expanded", func(t *testing.T) {
		var c config
		Expanded(true)(&c)
		if !c.expanded {
			t.Error("expanded should be true")
		}
	})

	t.Run("OnToggle", func(t *testing.T) {
		var c config
		called := false
		OnToggle(func(bool) { called = true })(&c)
		c.onToggle(true)
		if !called {
			t.Error("onToggle should have been called")
		}
	})

	t.Run("HeaderHeight", func(t *testing.T) {
		var c config
		HeaderHeight(48)(&c)
		if c.headerHeight != 48 {
			t.Errorf("headerHeight = %v, want 48", c.headerHeight)
		}
	})

	t.Run("HeaderColor", func(t *testing.T) {
		var c config
		HeaderColor(widget.ColorBlue)(&c)
		if c.headerColor != widget.ColorBlue {
			t.Error("headerColor should be blue")
		}
	})

	t.Run("ArrowColor", func(t *testing.T) {
		var c config
		ArrowColor(widget.ColorRed)(&c)
		if c.arrowColor != widget.ColorRed {
			t.Error("arrowColor should be red")
		}
	})

	t.Run("Animated", func(t *testing.T) {
		var c config
		Animated(false)(&c)
		if c.animated {
			t.Error("animated should be false")
		}
	})

	t.Run("Duration", func(t *testing.T) {
		var c config
		Duration(500 * time.Millisecond)(&c)
		if c.duration != 500*time.Millisecond {
			t.Errorf("duration = %v, want 500ms", c.duration)
		}
	})

	t.Run("PainterOpt", func(t *testing.T) {
		var c config
		PainterOpt(DefaultPainter{})(&c)
		if c.painter == nil {
			t.Error("painter should not be nil")
		}
	})

	t.Run("ExpandedSignal", func(t *testing.T) {
		var c config
		sig := state.NewSignal(true)
		ExpandedSignal(sig)(&c)
		if c.expandedSignal == nil {
			t.Error("expandedSignal should not be nil")
		}
	})

	t.Run("ExpandedReadonlySignal", func(t *testing.T) {
		var c config
		computed := state.NewComputed(func() bool { return true })
		ExpandedReadonlySignal(computed)(&c)
		if c.readonlyExpandedSignal == nil {
			t.Error("readonlyExpandedSignal should not be nil")
		}
	})
}

// --- Widget Defaults ---

func TestNew_InternalDefaults(t *testing.T) {
	w := New()

	if !w.IsVisible() {
		t.Error("should be visible by default")
	}
	if !w.IsEnabled() {
		t.Error("should be enabled by default")
	}
	if w.cfg.headerHeight != defaultHeaderHeight {
		t.Errorf("headerHeight = %v, want %v", w.cfg.headerHeight, defaultHeaderHeight)
	}
	if !w.cfg.animated {
		t.Error("should be animated by default")
	}
	if w.cfg.duration != defaultAnimDuration {
		t.Errorf("duration = %v, want %v", w.cfg.duration, defaultAnimDuration)
	}
	if w.istate != stateNormal {
		t.Errorf("state = %v, want stateNormal", w.istate)
	}
}

func TestNew_DefaultPainter(t *testing.T) {
	w := New()
	if w.painter == nil {
		t.Error("painter should not be nil by default")
	}
	if _, ok := w.painter.(DefaultPainter); !ok {
		t.Errorf("default painter should be DefaultPainter, got %T", w.painter)
	}
}

func TestNew_CustomPainter(t *testing.T) {
	p := &internalTestPainter{}
	w := New(PainterOpt(p))
	if w.painter != p {
		t.Error("painter should be the custom painter")
	}
}

// --- Mouse State Transitions ---

func TestMouseEvent_HoverEnterLeave(t *testing.T) {
	w := New(Title("Test"))
	w.SetBounds(geometry.NewRect(0, 0, 200, 36))
	ctx := widget.NewContext()

	enter := event.NewMouseEvent(event.MouseEnter, event.ButtonNone, 0,
		geometry.Pt(100, 18), geometry.Pt(100, 18), event.ModNone)
	consumed := handleEvent(w, ctx, enter)

	if !consumed {
		t.Error("MouseEnter should be consumed")
	}
	if w.istate != stateHover {
		t.Errorf("state = %v, want stateHover", w.istate)
	}

	leave := event.NewMouseEvent(event.MouseLeave, event.ButtonNone, 0,
		geometry.Pt(250, 18), geometry.Pt(250, 18), event.ModNone)
	consumed = handleEvent(w, ctx, leave)

	// MouseLeave is not consumed — allows content widgets to also handle it.
	_ = consumed
	if w.istate != stateNormal {
		t.Errorf("state = %v, want stateNormal", w.istate)
	}
}

func TestMouseEvent_MoveIntoHeader(t *testing.T) {
	w := New(Title("Test"))
	w.SetBounds(geometry.NewRect(0, 0, 200, 136)) // 36 header + 100 content
	ctx := widget.NewContext()

	// Move into header area.
	move := event.NewMouseEvent(event.MouseMove, event.ButtonNone, 0,
		geometry.Pt(100, 18), geometry.Pt(100, 18), event.ModNone)
	handleEvent(w, ctx, move)

	if w.istate != stateHover {
		t.Errorf("state = %v, want stateHover after move into header", w.istate)
	}

	// Move out of header (into content area).
	move2 := event.NewMouseEvent(event.MouseMove, event.ButtonNone, 0,
		geometry.Pt(100, 80), geometry.Pt(100, 80), event.ModNone)
	handleEvent(w, ctx, move2)

	if w.istate != stateNormal {
		t.Errorf("state = %v, want stateNormal after move out of header", w.istate)
	}
}

func TestMouseEvent_MoveAlreadyHover(t *testing.T) {
	w := New(Title("Test"))
	w.SetBounds(geometry.NewRect(0, 0, 200, 36))
	ctx := widget.NewContext()

	// Set to hover state.
	w.istate = stateHover

	// Move still in header - should stay hover.
	move := event.NewMouseEvent(event.MouseMove, event.ButtonNone, 0,
		geometry.Pt(100, 18), geometry.Pt(100, 18), event.ModNone)
	handleEvent(w, ctx, move)

	if w.istate != stateHover {
		t.Errorf("state = %v, want stateHover (unchanged)", w.istate)
	}
}

func TestMouseEvent_MoveFromNormalOutsideHeader(t *testing.T) {
	w := New(Title("Test"), Content(&internalMockWidget{}), Expanded(true), Animated(false))
	w.SetBounds(geometry.NewRect(0, 0, 200, 136))
	ctx := widget.NewContext()

	// Move in content area (below header) while in normal state.
	move := event.NewMouseEvent(event.MouseMove, event.ButtonNone, 0,
		geometry.Pt(100, 80), geometry.Pt(100, 80), event.ModNone)
	handleEvent(w, ctx, move)

	if w.istate != stateNormal {
		t.Errorf("state = %v, want stateNormal (move in content area)", w.istate)
	}
}

func TestMouseEvent_PressRelease(t *testing.T) {
	toggled := false
	w := New(Title("Test"), Animated(false), OnToggle(func(bool) { toggled = true }))
	w.SetBounds(geometry.NewRect(0, 0, 200, 36))
	ctx := widget.NewContext()

	press := event.NewMouseEvent(event.MousePress, event.ButtonLeft, event.ButtonStateLeft,
		geometry.Pt(100, 18), geometry.Pt(100, 18), event.ModNone)
	consumed := handleEvent(w, ctx, press)

	if !consumed {
		t.Error("MousePress on header should be consumed")
	}
	if w.istate != statePressed {
		t.Errorf("state = %v, want statePressed", w.istate)
	}
	if toggled {
		t.Error("should not toggle on press")
	}

	release := event.NewMouseEvent(event.MouseRelease, event.ButtonLeft, 0,
		geometry.Pt(100, 18), geometry.Pt(100, 18), event.ModNone)
	consumed = handleEvent(w, ctx, release)

	if !consumed {
		t.Error("MouseRelease should be consumed")
	}
	if !toggled {
		t.Error("should toggle on release inside header")
	}
}

func TestMouseEvent_PressReleaseOutside(t *testing.T) {
	toggled := false
	w := New(Title("Test"), Animated(false), OnToggle(func(bool) { toggled = true }))
	w.SetBounds(geometry.NewRect(0, 0, 200, 136))
	ctx := widget.NewContext()

	press := event.NewMouseEvent(event.MousePress, event.ButtonLeft, event.ButtonStateLeft,
		geometry.Pt(100, 18), geometry.Pt(100, 18), event.ModNone)
	handleEvent(w, ctx, press)

	// Release outside header area.
	release := event.NewMouseEvent(event.MouseRelease, event.ButtonLeft, 0,
		geometry.Pt(100, 100), geometry.Pt(100, 100), event.ModNone)
	handleEvent(w, ctx, release)

	if toggled {
		t.Error("should not toggle when released outside header")
	}
	if w.istate != stateNormal {
		t.Errorf("state = %v, want stateNormal", w.istate)
	}
}

func TestMouseEvent_RightButton_Ignored(t *testing.T) {
	w := New(Title("Test"), Animated(false))
	w.SetBounds(geometry.NewRect(0, 0, 200, 36))
	ctx := widget.NewContext()

	press := event.NewMouseEvent(event.MousePress, event.ButtonRight, event.ButtonStateRight,
		geometry.Pt(100, 18), geometry.Pt(100, 18), event.ModNone)
	consumed := handleEvent(w, ctx, press)

	if consumed {
		t.Error("right button should not be consumed")
	}
}

func TestMouseEvent_RightButtonRelease_Ignored(t *testing.T) {
	w := New(Title("Test"), Animated(false))
	w.SetBounds(geometry.NewRect(0, 0, 200, 36))
	ctx := widget.NewContext()

	release := event.NewMouseEvent(event.MouseRelease, event.ButtonRight, 0,
		geometry.Pt(100, 18), geometry.Pt(100, 18), event.ModNone)
	consumed := handleEvent(w, ctx, release)

	if consumed {
		t.Error("right button release should not be consumed")
	}
}

// --- Keyboard State Transitions ---

func TestKeyboard_SpaceActivation(t *testing.T) {
	w := New(Title("Test"), Animated(false))
	w.SetFocused(true)
	ctx := widget.NewContext()

	press := event.NewKeyEvent(event.KeyPress, event.KeySpace, 0, event.ModNone)
	consumed := handleEvent(w, ctx, press)
	if !consumed {
		t.Error("Space press should be consumed when focused")
	}
	if w.istate != statePressed {
		t.Errorf("state = %v, want statePressed", w.istate)
	}

	release := event.NewKeyEvent(event.KeyRelease, event.KeySpace, 0, event.ModNone)
	consumed = handleEvent(w, ctx, release)
	if !consumed {
		t.Error("Space release should be consumed")
	}
	if w.istate != stateNormal {
		t.Errorf("state = %v, want stateNormal after key release", w.istate)
	}
}

func TestKeyboard_NotFocused_Ignored(t *testing.T) {
	w := New(Title("Test"), Animated(false))
	w.SetFocused(false)
	ctx := widget.NewContext()

	press := event.NewKeyEvent(event.KeyPress, event.KeySpace, 0, event.ModNone)
	consumed := handleEvent(w, ctx, press)

	if consumed {
		t.Error("should not consume when not focused")
	}
}

func TestKeyboard_OtherKeys_Ignored(t *testing.T) {
	w := New(Title("Test"), Animated(false))
	w.SetFocused(true)
	ctx := widget.NewContext()

	press := event.NewKeyEvent(event.KeyPress, event.KeyA, 'a', event.ModNone)
	consumed := handleEvent(w, ctx, press)

	if consumed {
		t.Error("key A should not be consumed")
	}
}

// --- Unknown Event Type ---

func TestUnknownEvent_Ignored(t *testing.T) {
	w := New(Title("Test"))
	ctx := widget.NewContext()

	consumed := handleEvent(w, ctx, &event.FocusEvent{})

	if consumed {
		t.Error("unknown event type should not be consumed")
	}
}

// --- Progress Adapter ---

func TestProgressAdapter_GetSet(t *testing.T) {
	w := New()
	adapter := &progressAdapter{w: w}

	if adapter.Get() != 0.0 {
		t.Errorf("initial Get() = %v, want 0.0", adapter.Get())
	}

	adapter.Set(0.5)
	if w.progress != 0.5 {
		t.Errorf("progress = %v, want 0.5", w.progress)
	}
	if !w.NeedsRedraw() {
		t.Error("should need redraw after Set()")
	}
}

// --- Painting Helper Tests ---

func TestApplyStateModifier_All(t *testing.T) {
	base := widget.ColorRed

	normal := applyStateModifier(base, false, false)
	if normal != base {
		t.Error("normal should return base unchanged")
	}

	hover := applyStateModifier(base, true, false)
	if hover == base {
		t.Error("hover should modify base")
	}

	pressed := applyStateModifier(base, false, true)
	if pressed == base {
		t.Error("pressed should modify base")
	}

	// Pressed takes precedence.
	both := applyStateModifier(base, true, true)
	if both != pressed {
		t.Error("pressed should take precedence over hover")
	}
}

func TestResolveHeaderBg_CustomColor(t *testing.T) {
	s := HeaderState{HeaderColor: widget.ColorGreen}
	if resolveHeaderBg(s) != widget.ColorGreen {
		t.Error("should use custom header color")
	}
}

func TestResolveHeaderBg_Default(t *testing.T) {
	s := HeaderState{}
	bg := resolveHeaderBg(s)
	if bg == (widget.Color{}) {
		t.Error("should return non-zero default header bg")
	}
}

func TestResolveArrowColor_CustomColor(t *testing.T) {
	s := HeaderState{ArrowColor: widget.ColorRed}
	if resolveArrowColor(s) != widget.ColorRed {
		t.Error("should use custom arrow color")
	}
}

func TestResolveArrowColor_Default(t *testing.T) {
	s := HeaderState{}
	c := resolveArrowColor(s)
	if c == (widget.Color{}) {
		t.Error("should return non-zero default arrow color")
	}
}

// --- Mount Without Scheduler ---

func TestMount_NoScheduler(t *testing.T) {
	sig := state.NewSignal(false)
	w := New(ExpandedSignal(sig))
	ctx := widget.NewContext()
	// No scheduler set -- should not panic.
	w.Mount(ctx)
}

// --- Layout With Nil AnimCtrl ---

func TestTickAnimation_NilCtrl(t *testing.T) {
	w := New()
	w.animCtrl = nil
	ctx := widget.NewContext()
	// Should not panic.
	w.tickAnimation(ctx)
}

// --- Draw Content With Non-Settable Bounds Widget ---

func TestDrawContent_NonSettableBoundsWidget(t *testing.T) {
	content := &nonSettableBoundsWidget{}
	w := New(
		Content(content),
		Expanded(true),
		Animated(false),
	)
	ctx := widget.NewContext()
	constraints := geometry.Loose(geometry.Sz(200, 500))
	w.Layout(ctx, constraints)
	w.SetBounds(geometry.NewRect(0, 0, 200, 136))

	canvas := &internalMockCanvas{}
	// Should not panic even without SetBounds method.
	w.Draw(ctx, canvas)

	if !content.drawCalled {
		t.Error("content should be drawn even without SetBounds")
	}
}

// --- Key Release Without Prior Press ---

func TestKeyRelease_WithoutPress(t *testing.T) {
	toggled := false
	w := New(Animated(false), OnToggle(func(bool) { toggled = true }))
	w.SetFocused(true)
	ctx := widget.NewContext()

	// Release without prior press.
	release := event.NewKeyEvent(event.KeyRelease, event.KeySpace, 0, event.ModNone)
	consumed := handleEvent(w, ctx, release)

	if !consumed {
		t.Error("key release should still be consumed when focused")
	}
	if toggled {
		t.Error("should not toggle without prior press")
	}
}

// --- internalTestPainter ---

type internalTestPainter struct {
	called bool
	state  HeaderState
}

func (p *internalTestPainter) PaintHeader(_ widget.Canvas, hs HeaderState) {
	p.called = true
	p.state = hs
}

// --- internalMockWidget ---

type internalMockWidget struct {
	widget.WidgetBase
	preferredSize geometry.Size
	drawCalled    bool
}

func (m *internalMockWidget) Layout(_ widget.Context, c geometry.Constraints) geometry.Size {
	if m.preferredSize == (geometry.Size{}) {
		return c.Constrain(geometry.Sz(100, 100))
	}
	return c.Constrain(m.preferredSize)
}

func (m *internalMockWidget) Draw(_ widget.Context, _ widget.Canvas) {
	m.drawCalled = true
}

func (m *internalMockWidget) Event(_ widget.Context, _ event.Event) bool {
	return false
}

// --- nonSettableBoundsWidget (no SetBounds method) ---

type nonSettableBoundsWidget struct {
	drawCalled bool
}

func (w *nonSettableBoundsWidget) Layout(_ widget.Context, c geometry.Constraints) geometry.Size {
	return c.Constrain(geometry.Sz(100, 100))
}

func (w *nonSettableBoundsWidget) Draw(_ widget.Context, _ widget.Canvas) {
	w.drawCalled = true
}

func (w *nonSettableBoundsWidget) Event(_ widget.Context, _ event.Event) bool {
	return false
}

func (w *nonSettableBoundsWidget) Children() []widget.Widget {
	return nil
}

// --- internalMockCanvas ---

type internalMockCanvas struct {
	drawRectCount   int
	drawTextCount   int
	pushClipCount   int
	popClipCount    int
	drawLineCount   int
	strokeRectCount int
}

func (c *internalMockCanvas) Clear(_ widget.Color)                     {}
func (c *internalMockCanvas) DrawRect(_ geometry.Rect, _ widget.Color) { c.drawRectCount++ }
func (c *internalMockCanvas) StrokeRect(_ geometry.Rect, _ widget.Color, _ float32) {
	c.strokeRectCount++
}
func (c *internalMockCanvas) DrawRoundRect(_ geometry.Rect, _ widget.Color, _ float32)              {}
func (c *internalMockCanvas) StrokeRoundRect(_ geometry.Rect, _ widget.Color, _ float32, _ float32) {}
func (c *internalMockCanvas) DrawCircle(_ geometry.Point, _ float32, _ widget.Color)                {}
func (c *internalMockCanvas) StrokeCircle(_ geometry.Point, _ float32, _ widget.Color, _ float32)   {}
func (c *internalMockCanvas) DrawLine(_, _ geometry.Point, _ widget.Color, _ float32) {
	c.drawLineCount++
}

func (c *internalMockCanvas) DrawText(_ string, _ geometry.Rect, _ float32, _ widget.Color, _ bool, _ widget.TextAlign) {
	c.drawTextCount++
}

func (c *internalMockCanvas) MeasureText(text string, fontSize float32, _ bool) float32 {
	return float32(len([]rune(text))) * fontSize * 0.5
}

func (c *internalMockCanvas) DrawImage(_ image.Image, _ geometry.Point)    {}
func (c *internalMockCanvas) PushClip(_ geometry.Rect)                     { c.pushClipCount++ }
func (c *internalMockCanvas) PushClipRoundRect(_ geometry.Rect, _ float32) {}
func (c *internalMockCanvas) PopClip()                                     { c.popClipCount++ }
func (c *internalMockCanvas) PushTransform(_ geometry.Point)               {}
func (c *internalMockCanvas) PopTransform()                                {}
func (c *internalMockCanvas) TransformOffset() geometry.Point              { return geometry.Point{} }
