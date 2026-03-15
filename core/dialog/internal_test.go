package dialog

import (
	"image"
	"testing"

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

	t.Run("signal title", func(t *testing.T) {
		sig := state.NewSignal("signal")
		c := config{title: "static", titleFn: func() string { return "fn" }, titleSignal: sig}
		if got := c.ResolvedTitle(); got != "signal" {
			t.Errorf("ResolvedTitle() = %q, want %q", got, "signal")
		}
	})

	t.Run("readonly signal title", func(t *testing.T) {
		computed := state.NewComputed(func() string { return "readonly" })
		c := config{
			title:               "static",
			titleFn:             func() string { return "fn" },
			titleSignal:         state.NewSignal("signal"),
			readonlyTitleSignal: computed,
		}
		if got := c.ResolvedTitle(); got != "readonly" {
			t.Errorf("ResolvedTitle() = %q, want %q", got, "readonly")
		}
	})
}

func TestConfig_ResolvedTitle_Priority(t *testing.T) {
	t.Run("fn beats static", func(t *testing.T) {
		c := config{title: "static", titleFn: func() string { return "fn" }}
		if got := c.ResolvedTitle(); got != "fn" {
			t.Errorf("ResolvedTitle() = %q, want %q", got, "fn")
		}
	})

	t.Run("signal beats fn", func(t *testing.T) {
		sig := state.NewSignal("signal")
		c := config{title: "static", titleFn: func() string { return "fn" }, titleSignal: sig}
		if got := c.ResolvedTitle(); got != "signal" {
			t.Errorf("ResolvedTitle() = %q, want %q", got, "signal")
		}
	})

	t.Run("readonlySignal beats signal", func(t *testing.T) {
		computed := state.NewComputed(func() string { return "readonly" })
		sig := state.NewSignal("signal")
		c := config{titleSignal: sig, readonlyTitleSignal: computed}
		if got := c.ResolvedTitle(); got != "readonly" {
			t.Errorf("ResolvedTitle() = %q, want %q", got, "readonly")
		}
	})
}

func TestConfig_ResolvedTitle_SignalUpdate(t *testing.T) {
	sig := state.NewSignal("initial")
	c := config{titleSignal: sig}

	if got := c.ResolvedTitle(); got != "initial" {
		t.Errorf("ResolvedTitle() = %q, want %q", got, "initial")
	}

	sig.Set("updated")
	if got := c.ResolvedTitle(); got != "updated" {
		t.Errorf("ResolvedTitle() = %q, want %q", got, "updated")
	}
}

// --- Options Tests ---

func TestOptions(t *testing.T) {
	t.Run("Title", func(t *testing.T) {
		var c config
		Title("Hello")(&c)
		if c.title != "Hello" {
			t.Errorf("title = %q, want %q", c.title, "Hello")
		}
	})

	t.Run("TitleFn", func(t *testing.T) {
		var c config
		TitleFn(func() string { return "dynamic" })(&c)
		if c.titleFn == nil {
			t.Error("titleFn should not be nil")
		}
	})

	t.Run("Actions", func(t *testing.T) {
		var c config
		Actions(Action{Label: "OK"}, Action{Label: "Cancel"})(&c)
		if len(c.actions) != 2 {
			t.Errorf("actions count = %d, want 2", len(c.actions))
		}
	})

	t.Run("DismissibleOpt", func(t *testing.T) {
		var c config
		DismissibleOpt(false)(&c)
		if c.dismissible {
			t.Error("dismissible should be false")
		}
	})

	t.Run("EscapeToCloseOpt", func(t *testing.T) {
		var c config
		EscapeToCloseOpt(false)(&c)
		if c.escToClose {
			t.Error("escToClose should be false")
		}
	})

	t.Run("OnClose", func(t *testing.T) {
		var c config
		called := false
		OnClose(func() { called = true })(&c)
		if c.onClose == nil {
			t.Fatal("onClose should not be nil")
		}
		c.onClose()
		if !called {
			t.Error("onClose should have been called")
		}
	})

	t.Run("MaxWidth", func(t *testing.T) {
		var c config
		MaxWidth(400)(&c)
		if c.maxWidth != 400 {
			t.Errorf("maxWidth = %v, want 400", c.maxWidth)
		}
	})

	t.Run("MaxHeight", func(t *testing.T) {
		var c config
		MaxHeight(300)(&c)
		if c.maxHeight != 300 {
			t.Errorf("maxHeight = %v, want 300", c.maxHeight)
		}
	})

	t.Run("PainterOpt", func(t *testing.T) {
		var c config
		p := DefaultPainter{}
		PainterOpt(p)(&c)
		if c.painter == nil {
			t.Fatal("painter should not be nil")
		}
	})

	t.Run("Content", func(t *testing.T) {
		var c config
		w := &internalMockWidget{}
		Content(w)(&c)
		if c.content == nil {
			t.Error("content should not be nil")
		}
	})

	t.Run("TitleSignal", func(t *testing.T) {
		var c config
		sig := state.NewSignal("test")
		TitleSignal(sig)(&c)
		if c.titleSignal == nil {
			t.Error("titleSignal should not be nil")
		}
	})

	t.Run("TitleReadonlySignal", func(t *testing.T) {
		var c config
		computed := state.NewComputed(func() string { return "computed" })
		TitleReadonlySignal(computed)(&c)
		if c.readonlyTitleSignal == nil {
			t.Error("readonlyTitleSignal should not be nil")
		}
	})
}

// --- Widget Defaults Tests ---

func TestNew_Defaults(t *testing.T) {
	w := New()

	if w.IsOpen() {
		t.Error("should not be open by default")
	}
	if !w.IsEnabled() {
		t.Error("should be enabled by default")
	}
	if w.cfg.maxWidth != defaultMaxWidth {
		t.Errorf("maxWidth = %v, want %v", w.cfg.maxWidth, defaultMaxWidth)
	}
	if !w.cfg.dismissible {
		t.Error("should be dismissible by default")
	}
	if !w.cfg.escToClose {
		t.Error("should close on Escape by default")
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
	custom := &testPainter{}
	w := New(PainterOpt(custom))
	if w.painter != custom {
		t.Error("painter should be the custom painter")
	}
}

// --- computeDialogBounds Tests ---

func TestComputeDialogBounds_Default(t *testing.T) {
	w := New(Title("Test"))
	windowSize := geometry.Sz(800, 600)

	bounds := w.computeDialogBounds(windowSize)

	if bounds.Width() <= 0 || bounds.Height() <= 0 {
		t.Error("dialog bounds should have positive dimensions")
	}
	if bounds.Width() > defaultMaxWidth {
		t.Errorf("width %v exceeds maxWidth %v", bounds.Width(), defaultMaxWidth)
	}
}

func TestComputeDialogBounds_CustomMaxWidth(t *testing.T) {
	w := New(Title("Test"), MaxWidth(300))
	windowSize := geometry.Sz(800, 600)

	bounds := w.computeDialogBounds(windowSize)

	if bounds.Width() > 300 {
		t.Errorf("width %v exceeds custom maxWidth 300", bounds.Width())
	}
}

func TestComputeDialogBounds_SmallWindow(t *testing.T) {
	w := New(Title("Test"))
	windowSize := geometry.Sz(200, 200)

	bounds := w.computeDialogBounds(windowSize)

	if bounds.Width() > windowSize.Width*windowMargin+1 {
		t.Errorf("width %v exceeds available window space", bounds.Width())
	}
}

func TestComputeDialogBounds_Centered(t *testing.T) {
	w := New(Title("Test"))
	windowSize := geometry.Sz(800, 600)

	bounds := w.computeDialogBounds(windowSize)

	centerX := (bounds.Min.X + bounds.Max.X) / 2
	centerY := (bounds.Min.Y + bounds.Max.Y) / 2

	windowCenterX := windowSize.Width / 2
	windowCenterY := windowSize.Height / 2

	if abs(centerX-windowCenterX) > 1 {
		t.Errorf("dialog not horizontally centered: centerX=%v, windowCenterX=%v", centerX, windowCenterX)
	}
	if abs(centerY-windowCenterY) > 1 {
		t.Errorf("dialog not vertically centered: centerY=%v, windowCenterY=%v", centerY, windowCenterY)
	}
}

func abs(v float32) float32 {
	if v < 0 {
		return -v
	}
	return v
}

// --- surfaceWidget Tests ---

func TestSurfaceWidget_Layout(t *testing.T) {
	w := New(Title("Test"))
	ctx := widget.NewContext()
	ctx.SetWindowSize(geometry.Sz(800, 600))

	s := newSurfaceWidget(w, ctx)
	constraints := geometry.Loose(geometry.Sz(800, 600))
	size := s.Layout(ctx, constraints)

	if size.Width != 800 || size.Height != 600 {
		t.Errorf("surface should fill window, got %v", size)
	}
}

func TestSurfaceWidget_Draw(t *testing.T) {
	w := New(Title("Hello"))
	ctx := widget.NewContext()
	ctx.SetWindowSize(geometry.Sz(800, 600))

	s := newSurfaceWidget(w, ctx)
	s.SetBounds(geometry.NewRect(0, 0, 800, 600))
	canvas := &internalRecordingCanvas{}

	s.Draw(ctx, canvas)

	if len(canvas.drawRects) == 0 {
		t.Error("should draw backdrop rect")
	}
	if len(canvas.drawRoundRects) == 0 {
		t.Error("should draw dialog surface")
	}
}

func TestSurfaceWidget_Draw_NilCanvas(t *testing.T) {
	w := New(Title("Hello"))
	ctx := widget.NewContext()
	s := newSurfaceWidget(w, ctx)

	s.Draw(ctx, nil) // should not panic
}

func TestSurfaceWidget_Draw_WithContent(t *testing.T) {
	content := &internalMockWidget{}
	w := New(Title("Hello"), Content(content))
	ctx := widget.NewContext()
	ctx.SetWindowSize(geometry.Sz(800, 600))

	s := newSurfaceWidget(w, ctx)
	s.SetBounds(geometry.NewRect(0, 0, 800, 600))
	canvas := &internalRecordingCanvas{}

	s.Draw(ctx, canvas)

	if !content.drawn {
		t.Error("content widget should have been drawn")
	}
}

// --- surfaceWidget Event Tests ---

func TestSurfaceWidget_Escape_Close(t *testing.T) {
	closed := false
	w := New(
		Title("Test"),
		EscapeToCloseOpt(true),
		OnClose(func() { closed = true }),
	)
	ctx := widget.NewContext()
	ctx.SetWindowSize(geometry.Sz(800, 600))
	om := &internalMockOverlayManager{}
	ctx.SetOverlayManager(om)

	w.Show(ctx)
	s := w.surface

	esc := event.NewKeyEvent(event.KeyPress, event.KeyEscape, 0, event.ModNone)
	consumed := s.Event(ctx, esc)

	if !consumed {
		t.Error("Escape should be consumed")
	}
	if !closed {
		t.Error("Escape should close the dialog")
	}
}

func TestSurfaceWidget_Escape_NoClose(t *testing.T) {
	closed := false
	w := New(
		Title("Test"),
		EscapeToCloseOpt(false),
		OnClose(func() { closed = true }),
	)
	ctx := widget.NewContext()
	ctx.SetWindowSize(geometry.Sz(800, 600))
	om := &internalMockOverlayManager{}
	ctx.SetOverlayManager(om)

	w.Show(ctx)
	s := w.surface

	esc := event.NewKeyEvent(event.KeyPress, event.KeyEscape, 0, event.ModNone)
	consumed := s.Event(ctx, esc)

	if !consumed {
		t.Error("Escape should still be consumed (modal)")
	}
	if closed {
		t.Error("Escape should NOT close when escToClose is false")
	}
}

func TestSurfaceWidget_Tab_FocusCycle(t *testing.T) {
	w := New(
		Title("Test"),
		Actions(
			Action{Label: "Cancel"},
			Action{Label: "OK", Default: true},
		),
	)
	ctx := widget.NewContext()
	ctx.SetWindowSize(geometry.Sz(800, 600))
	om := &internalMockOverlayManager{}
	ctx.SetOverlayManager(om)

	w.Show(ctx)
	s := w.surface

	// Default focus should be on "OK" (index 1, last action with Default=true).
	if s.focusIndex != 1 {
		t.Errorf("initial focusIndex = %d, want 1 (OK button)", s.focusIndex)
	}

	// Tab should move to Cancel (index 0, wraps around).
	tab := event.NewKeyEvent(event.KeyPress, event.KeyTab, 0, event.ModNone)
	s.Event(ctx, tab)
	if s.focusIndex != 0 {
		t.Errorf("focusIndex after Tab = %d, want 0 (Cancel)", s.focusIndex)
	}

	// Tab again should wrap to OK (index 1).
	s.Event(ctx, tab)
	if s.focusIndex != 1 {
		t.Errorf("focusIndex after second Tab = %d, want 1 (OK)", s.focusIndex)
	}
}

func TestSurfaceWidget_ShiftTab_FocusCycle(t *testing.T) {
	w := New(
		Title("Test"),
		Actions(
			Action{Label: "Cancel"},
			Action{Label: "OK", Default: true},
		),
	)
	ctx := widget.NewContext()
	ctx.SetWindowSize(geometry.Sz(800, 600))
	om := &internalMockOverlayManager{}
	ctx.SetOverlayManager(om)

	w.Show(ctx)
	s := w.surface

	// Start at OK (index 1).
	if s.focusIndex != 1 {
		t.Fatalf("initial focusIndex = %d, want 1", s.focusIndex)
	}

	// Shift+Tab should move to Cancel (index 0).
	shiftTab := event.NewKeyEvent(event.KeyPress, event.KeyTab, 0, event.ModShift)
	s.Event(ctx, shiftTab)
	if s.focusIndex != 0 {
		t.Errorf("focusIndex after Shift+Tab = %d, want 0", s.focusIndex)
	}

	// Shift+Tab again should wrap to OK (index 1).
	s.Event(ctx, shiftTab)
	if s.focusIndex != 1 {
		t.Errorf("focusIndex after second Shift+Tab = %d, want 1", s.focusIndex)
	}
}

func TestSurfaceWidget_Enter_ActivatesAction(t *testing.T) {
	clicked := false
	w := New(
		Title("Test"),
		Actions(Action{Label: "OK", OnClick: func() { clicked = true }, Default: true}),
	)
	ctx := widget.NewContext()
	ctx.SetWindowSize(geometry.Sz(800, 600))
	om := &internalMockOverlayManager{}
	ctx.SetOverlayManager(om)

	w.Show(ctx)
	s := w.surface

	enter := event.NewKeyEvent(event.KeyPress, event.KeyEnter, 0, event.ModNone)
	s.Event(ctx, enter)

	if !clicked {
		t.Error("Enter should activate the focused action")
	}
	if w.IsOpen() {
		t.Error("dialog should close after action activation")
	}
}

func TestSurfaceWidget_Space_ActivatesAction(t *testing.T) {
	clicked := false
	w := New(
		Title("Test"),
		Actions(Action{Label: "OK", OnClick: func() { clicked = true }, Default: true}),
	)
	ctx := widget.NewContext()
	ctx.SetWindowSize(geometry.Sz(800, 600))
	om := &internalMockOverlayManager{}
	ctx.SetOverlayManager(om)

	w.Show(ctx)
	s := w.surface

	space := event.NewKeyEvent(event.KeyPress, event.KeySpace, ' ', event.ModNone)
	s.Event(ctx, space)

	if !clicked {
		t.Error("Space should activate the focused action")
	}
}

func TestSurfaceWidget_BackdropClick_Dismiss(t *testing.T) {
	closed := false
	w := New(
		Title("Test"),
		DismissibleOpt(true),
		OnClose(func() { closed = true }),
	)
	ctx := widget.NewContext()
	ctx.SetWindowSize(geometry.Sz(800, 600))
	om := &internalMockOverlayManager{}
	ctx.SetOverlayManager(om)

	w.Show(ctx)
	s := w.surface

	// Click at top-left corner (should be backdrop, not dialog surface).
	click := event.NewMouseEvent(event.MousePress, event.ButtonLeft, event.ButtonStateLeft,
		geometry.Pt(5, 5), geometry.Pt(5, 5), event.ModNone)
	consumed := s.Event(ctx, click)

	if !consumed {
		t.Error("backdrop click should be consumed")
	}
	if !closed {
		t.Error("backdrop click should close dismissible dialog")
	}
}

func TestSurfaceWidget_BackdropClick_NoDismiss(t *testing.T) {
	closed := false
	w := New(
		Title("Test"),
		DismissibleOpt(false),
		OnClose(func() { closed = true }),
	)
	ctx := widget.NewContext()
	ctx.SetWindowSize(geometry.Sz(800, 600))
	om := &internalMockOverlayManager{}
	ctx.SetOverlayManager(om)

	w.Show(ctx)
	s := w.surface

	// Click at top-left corner.
	click := event.NewMouseEvent(event.MousePress, event.ButtonLeft, event.ButtonStateLeft,
		geometry.Pt(5, 5), geometry.Pt(5, 5), event.ModNone)
	consumed := s.Event(ctx, click)

	if !consumed {
		t.Error("click should still be consumed (modal)")
	}
	if closed {
		t.Error("backdrop click should NOT close non-dismissible dialog")
	}
}

func TestSurfaceWidget_ActionClick(t *testing.T) {
	clicked := false
	w := New(
		Title("Test"),
		Actions(Action{Label: "OK", OnClick: func() { clicked = true }}),
	)
	ctx := widget.NewContext()
	ctx.SetWindowSize(geometry.Sz(800, 600))
	om := &internalMockOverlayManager{}
	ctx.SetOverlayManager(om)

	w.Show(ctx)
	s := w.surface

	// Calculate where the OK button should be in dialog bounds.
	dialogBounds := w.computeDialogBounds(geometry.Sz(800, 600))
	btnX := dialogBounds.Max.X - dialogPadding - (float32(len("OK"))*actionCharWidth + actionPaddingX*2)
	btnY := dialogBounds.Max.Y - dialogPadding - actionHeight

	// Click in the center of the action button.
	clickX := btnX + (float32(len("OK"))*actionCharWidth+actionPaddingX*2)/2
	clickY := btnY + actionHeight/2

	click := event.NewMouseEvent(event.MousePress, event.ButtonLeft, event.ButtonStateLeft,
		geometry.Pt(clickX, clickY), geometry.Pt(clickX, clickY), event.ModNone)
	s.Event(ctx, click)

	if !clicked {
		t.Error("clicking action button should trigger OnClick")
	}
}

func TestSurfaceWidget_Modal_ConsumesAllEvents(t *testing.T) {
	w := New(Title("Test"))
	ctx := widget.NewContext()
	ctx.SetWindowSize(geometry.Sz(800, 600))
	om := &internalMockOverlayManager{}
	ctx.SetOverlayManager(om)

	w.Show(ctx)
	s := w.surface

	// Focus event (not key or mouse) should be consumed.
	fe := event.NewFocusEvent(event.FocusGained)
	consumed := s.Event(ctx, fe)

	if !consumed {
		t.Error("modal dialog should consume all events")
	}
}

func TestSurfaceWidget_KeyRelease_Consumed(t *testing.T) {
	w := New(Title("Test"))
	ctx := widget.NewContext()
	ctx.SetWindowSize(geometry.Sz(800, 600))
	om := &internalMockOverlayManager{}
	ctx.SetOverlayManager(om)

	w.Show(ctx)
	s := w.surface

	// Key release should be consumed but not acted upon.
	release := event.NewKeyEvent(event.KeyRelease, event.KeyA, 'a', event.ModNone)
	consumed := s.Event(ctx, release)

	if !consumed {
		t.Error("key release should be consumed by modal dialog")
	}
}

func TestSurfaceWidget_MouseMove_Consumed(t *testing.T) {
	w := New(Title("Test"))
	ctx := widget.NewContext()
	ctx.SetWindowSize(geometry.Sz(800, 600))
	om := &internalMockOverlayManager{}
	ctx.SetOverlayManager(om)

	w.Show(ctx)
	s := w.surface

	// Mouse move should be consumed but not close anything.
	move := event.NewMouseEvent(event.MouseMove, event.ButtonNone, 0,
		geometry.Pt(100, 100), geometry.Pt(100, 100), event.ModNone)
	consumed := s.Event(ctx, move)

	if !consumed {
		t.Error("mouse move should be consumed by modal dialog")
	}
}

func TestSurfaceWidget_DefaultFocus_LastAction(t *testing.T) {
	w := New(
		Title("Test"),
		Actions(
			Action{Label: "Cancel"},
			Action{Label: "OK"},
		),
	)
	ctx := widget.NewContext()
	s := newSurfaceWidget(w, ctx)

	// When no action has Default=true, focus should be on the last one.
	if s.focusIndex != 1 {
		t.Errorf("focusIndex = %d, want 1 (last action)", s.focusIndex)
	}
}

func TestSurfaceWidget_DefaultFocus_ExplicitDefault(t *testing.T) {
	w := New(
		Title("Test"),
		Actions(
			Action{Label: "Cancel", Default: true},
			Action{Label: "OK"},
		),
	)
	ctx := widget.NewContext()
	s := newSurfaceWidget(w, ctx)

	if s.focusIndex != 0 {
		t.Errorf("focusIndex = %d, want 0 (Cancel has Default=true)", s.focusIndex)
	}
}

func TestSurfaceWidget_NoActions_FocusNegative(t *testing.T) {
	w := New(Title("Test"))
	ctx := widget.NewContext()
	s := newSurfaceWidget(w, ctx)

	if s.focusIndex != -1 {
		t.Errorf("focusIndex = %d, want -1 (no actions)", s.focusIndex)
	}
}

func TestSurfaceWidget_Tab_NoActions_NoChange(t *testing.T) {
	w := New(Title("Test"))
	ctx := widget.NewContext()
	ctx.SetWindowSize(geometry.Sz(800, 600))
	om := &internalMockOverlayManager{}
	ctx.SetOverlayManager(om)

	w.Show(ctx)
	s := w.surface

	tab := event.NewKeyEvent(event.KeyPress, event.KeyTab, 0, event.ModNone)
	s.Event(ctx, tab)

	if s.focusIndex != -1 {
		t.Errorf("focusIndex = %d, want -1 (no actions to cycle)", s.focusIndex)
	}
}

func TestSurfaceWidget_Enter_NoActions_NoClose(t *testing.T) {
	w := New(Title("Test"))
	ctx := widget.NewContext()
	ctx.SetWindowSize(geometry.Sz(800, 600))
	om := &internalMockOverlayManager{}
	ctx.SetOverlayManager(om)

	w.Show(ctx)
	s := w.surface

	enter := event.NewKeyEvent(event.KeyPress, event.KeyEnter, 0, event.ModNone)
	s.Event(ctx, enter)

	// Dialog should remain open since no action is focused.
	if !w.IsOpen() {
		t.Error("dialog should remain open when Enter is pressed with no focused action")
	}
}

func TestSurfaceWidget_Dismiss(t *testing.T) {
	closed := false
	w := New(Title("Test"), OnClose(func() { closed = true }))
	ctx := widget.NewContext()
	ctx.SetWindowSize(geometry.Sz(800, 600))
	om := &internalMockOverlayManager{}
	ctx.SetOverlayManager(om)

	w.Show(ctx)
	s := w.surface

	s.Dismiss()

	if !closed {
		t.Error("Dismiss should close the dialog")
	}
}

func TestSurfaceWidget_Modal(t *testing.T) {
	w := New(Title("Test"))
	ctx := widget.NewContext()
	s := newSurfaceWidget(w, ctx)

	if !s.Modal() {
		t.Error("dialog surface should always be modal")
	}
}

func TestSurfaceWidget_Children(t *testing.T) {
	w := New(Title("Test"))
	ctx := widget.NewContext()
	s := newSurfaceWidget(w, ctx)

	if s.Children() != nil {
		t.Error("surface widget should have no children")
	}
}

// --- Convenience Constructor Internal Tests ---

func TestAlert_Structure(t *testing.T) {
	d := Alert("Error", "Failed", func() {})
	if d.cfg.title != "Error" {
		t.Errorf("title = %q, want %q", d.cfg.title, "Error")
	}
	if len(d.cfg.actions) != 1 {
		t.Fatalf("actions count = %d, want 1", len(d.cfg.actions))
	}
	if d.cfg.actions[0].Label != alertOKLabel {
		t.Errorf("action label = %q, want %q", d.cfg.actions[0].Label, alertOKLabel)
	}
	if d.cfg.actions[0].Variant != VariantFilled {
		t.Errorf("action variant = %d, want VariantFilled", d.cfg.actions[0].Variant)
	}
	if !d.cfg.actions[0].Default {
		t.Error("OK action should have Default=true")
	}
}

func TestConfirm_Structure(t *testing.T) {
	d := Confirm("Delete?", "Sure?", func() {}, func() {})
	if d.cfg.title != "Delete?" {
		t.Errorf("title = %q, want %q", d.cfg.title, "Delete?")
	}
	if len(d.cfg.actions) != 2 {
		t.Fatalf("actions count = %d, want 2", len(d.cfg.actions))
	}
	if d.cfg.actions[0].Label != confirmCancelLabel {
		t.Errorf("first action label = %q, want %q", d.cfg.actions[0].Label, confirmCancelLabel)
	}
	if d.cfg.actions[0].Variant != VariantTextOnly {
		t.Errorf("cancel variant = %d, want VariantTextOnly", d.cfg.actions[0].Variant)
	}
	if d.cfg.actions[1].Label != alertOKLabel {
		t.Errorf("second action label = %q, want %q", d.cfg.actions[1].Label, alertOKLabel)
	}
	if !d.cfg.actions[1].Default {
		t.Error("OK action should have Default=true")
	}
}

// --- Action Tests ---

func TestAction_NilOnClick(t *testing.T) {
	w := New(
		Title("Test"),
		Actions(Action{Label: "OK"}), // no OnClick
	)
	ctx := widget.NewContext()
	ctx.SetWindowSize(geometry.Sz(800, 600))
	om := &internalMockOverlayManager{}
	ctx.SetOverlayManager(om)

	w.Show(ctx)
	s := w.surface

	// Should not panic when Enter is pressed on action with nil OnClick.
	enter := event.NewKeyEvent(event.KeyPress, event.KeyEnter, 0, event.ModNone)
	s.Event(ctx, enter)

	if w.IsOpen() {
		t.Error("dialog should still close even with nil OnClick")
	}
}

// --- Double Close Protection ---

func TestDoClose_Idempotent(t *testing.T) {
	closeCount := 0
	w := New(Title("Test"), OnClose(func() { closeCount++ }))
	ctx := widget.NewContext()
	ctx.SetWindowSize(geometry.Sz(800, 600))
	om := &internalMockOverlayManager{}
	ctx.SetOverlayManager(om)

	w.Show(ctx)
	w.doClose(ctx)
	w.doClose(ctx) // second close should be no-op

	if closeCount != 1 {
		t.Errorf("onClose called %d times, want 1", closeCount)
	}
}

// --- testPainter records calls for testing ---

type testPainter struct {
	called bool
	state  PaintState
}

func (p *testPainter) PaintDialog(_ widget.Canvas, ps PaintState) {
	p.called = true
	p.state = ps
}

// --- internalMockWidget ---

type internalMockWidget struct {
	widget.WidgetBase
	drawn bool
}

func (w *internalMockWidget) Layout(_ widget.Context, c geometry.Constraints) geometry.Size {
	return c.Constrain(geometry.Sz(200, 100))
}

func (w *internalMockWidget) Draw(_ widget.Context, _ widget.Canvas) {
	w.drawn = true
}

func (w *internalMockWidget) Event(_ widget.Context, _ event.Event) bool {
	return false
}

func (w *internalMockWidget) Children() []widget.Widget {
	return nil
}

// --- internalRecordingCanvas ---

type internalRecordingCanvas struct {
	drawRects        []internalDrawRectCall
	drawRoundRects   []internalDrawRoundRectCall
	strokeRoundRects []internalStrokeRoundRectCall
	drawTexts        []internalDrawTextCall
}

type internalDrawRectCall struct {
	r     geometry.Rect
	color widget.Color
}

type internalDrawRoundRectCall struct {
	r      geometry.Rect
	color  widget.Color
	radius float32
}

type internalStrokeRoundRectCall struct {
	r           geometry.Rect
	color       widget.Color
	radius      float32
	strokeWidth float32
}

type internalDrawTextCall struct {
	text     string
	bounds   geometry.Rect
	fontSize float32
	color    widget.Color
	bold     bool
	align    widget.TextAlign
}

func (c *internalRecordingCanvas) Clear(_ widget.Color) {}

func (c *internalRecordingCanvas) DrawRect(r geometry.Rect, color widget.Color) {
	c.drawRects = append(c.drawRects, internalDrawRectCall{r: r, color: color})
}

func (c *internalRecordingCanvas) StrokeRect(_ geometry.Rect, _ widget.Color, _ float32) {}

func (c *internalRecordingCanvas) DrawRoundRect(r geometry.Rect, color widget.Color, radius float32) {
	c.drawRoundRects = append(c.drawRoundRects, internalDrawRoundRectCall{r: r, color: color, radius: radius})
}

func (c *internalRecordingCanvas) StrokeRoundRect(r geometry.Rect, color widget.Color, radius float32, strokeWidth float32) {
	c.strokeRoundRects = append(c.strokeRoundRects, internalStrokeRoundRectCall{r: r, color: color, radius: radius, strokeWidth: strokeWidth})
}

func (c *internalRecordingCanvas) DrawCircle(_ geometry.Point, _ float32, _ widget.Color) {}
func (c *internalRecordingCanvas) StrokeCircle(_ geometry.Point, _ float32, _ widget.Color, _ float32) {
}
func (c *internalRecordingCanvas) DrawLine(_, _ geometry.Point, _ widget.Color, _ float32) {}

func (c *internalRecordingCanvas) DrawText(text string, bounds geometry.Rect, fontSize float32, color widget.Color, bold bool, align widget.TextAlign) {
	c.drawTexts = append(c.drawTexts, internalDrawTextCall{text: text, bounds: bounds, fontSize: fontSize, color: color, bold: bold, align: align})
}

func (c *internalRecordingCanvas) MeasureText(text string, fontSize float32, _ bool) float32 {
	return float32(len([]rune(text))) * fontSize * 0.5
}

func (c *internalRecordingCanvas) DrawImage(_ image.Image, _ geometry.Point)    {}
func (c *internalRecordingCanvas) PushClip(_ geometry.Rect)                     {}
func (c *internalRecordingCanvas) PushClipRoundRect(_ geometry.Rect, _ float32) {}
func (c *internalRecordingCanvas) PopClip()                                     {}
func (c *internalRecordingCanvas) PushTransform(_ geometry.Point)               {}
func (c *internalRecordingCanvas) PopTransform()                                {}
func (c *internalRecordingCanvas) TransformOffset() geometry.Point              { return geometry.Point{} }

// --- internalMockOverlayManager ---

type internalMockOverlayManager struct {
	pushCount   int
	removeCount int
	lastWidget  widget.Widget
}

func (m *internalMockOverlayManager) PushOverlay(w widget.Widget, _ func()) {
	m.pushCount++
	m.lastWidget = w
}

func (m *internalMockOverlayManager) PopOverlay() {}

func (m *internalMockOverlayManager) RemoveOverlay(_ widget.Widget) {
	m.removeCount++
}
