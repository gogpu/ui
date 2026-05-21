package checkbox

import (
	"github.com/gogpu/gg/scene"
	"image"
	"testing"

	"github.com/gogpu/ui/event"
	"github.com/gogpu/ui/geometry"
	"github.com/gogpu/ui/state"
	"github.com/gogpu/ui/widget"
)

// --- Config Tests ---

func TestConfig_ResolvedLabel(t *testing.T) {
	t.Run("static label", func(t *testing.T) {
		c := config{label: "Hello"}
		if got := c.ResolvedLabel(); got != "Hello" {
			t.Errorf("ResolvedLabel() = %q, want %q", got, "Hello")
		}
	})

	t.Run("dynamic label", func(t *testing.T) {
		c := config{label: "static", labelFn: func() string { return "dynamic" }}
		if got := c.ResolvedLabel(); got != "dynamic" {
			t.Errorf("ResolvedLabel() = %q, want %q", got, "dynamic")
		}
	})

	t.Run("empty label", func(t *testing.T) {
		c := config{}
		if got := c.ResolvedLabel(); got != "" {
			t.Errorf("ResolvedLabel() = %q, want empty", got)
		}
	})
}

func TestConfig_ResolvedChecked(t *testing.T) {
	t.Run("static checked", func(t *testing.T) {
		c := config{checked: true}
		if !c.ResolvedChecked() {
			t.Error("ResolvedChecked() should be true")
		}
	})

	t.Run("static unchecked", func(t *testing.T) {
		c := config{checked: false}
		if c.ResolvedChecked() {
			t.Error("ResolvedChecked() should be false")
		}
	})

	t.Run("dynamic checked", func(t *testing.T) {
		c := config{checked: false, checkedFn: func() bool { return true }}
		if !c.ResolvedChecked() {
			t.Error("ResolvedChecked() with fn should be true")
		}
	})

	t.Run("dynamic overrides static", func(t *testing.T) {
		c := config{checked: true, checkedFn: func() bool { return false }}
		if c.ResolvedChecked() {
			t.Error("ResolvedChecked() with fn returning false should be false")
		}
	})
}

func TestConfig_ResolvedDisabled(t *testing.T) {
	t.Run("static disabled", func(t *testing.T) {
		c := config{disabled: true}
		if !c.ResolvedDisabled() {
			t.Error("ResolvedDisabled() should be true")
		}
	})

	t.Run("dynamic disabled", func(t *testing.T) {
		c := config{disabled: false, disabledFn: func() bool { return true }}
		if !c.ResolvedDisabled() {
			t.Error("ResolvedDisabled() with fn should be true")
		}
	})

	t.Run("dynamic overrides static", func(t *testing.T) {
		c := config{disabled: true, disabledFn: func() bool { return false }}
		if c.ResolvedDisabled() {
			t.Error("ResolvedDisabled() with fn returning false should be false")
		}
	})
}

// --- Options Tests ---

func TestOptions(t *testing.T) {
	t.Run("LabelOpt", func(t *testing.T) {
		var c config
		LabelOpt("Hello")(&c)
		if c.label != "Hello" {
			t.Errorf("label = %q, want %q", c.label, "Hello")
		}
	})

	t.Run("LabelFn", func(t *testing.T) {
		var c config
		fn := func() string { return "dynamic" }
		LabelFn(fn)(&c)
		if c.labelFn == nil {
			t.Error("labelFn should not be nil")
		}
		if c.labelFn() != "dynamic" {
			t.Errorf("labelFn() = %q, want %q", c.labelFn(), "dynamic")
		}
	})

	t.Run("Checked", func(t *testing.T) {
		var c config
		Checked(true)(&c)
		if !c.checked {
			t.Error("checked should be true")
		}
	})

	t.Run("CheckedFn", func(t *testing.T) {
		var c config
		CheckedFn(func() bool { return true })(&c)
		if c.checkedFn == nil {
			t.Error("checkedFn should not be nil")
		}
	})

	t.Run("OnToggle", func(t *testing.T) {
		var c config
		called := false
		OnToggle(func(bool) { called = true })(&c)
		if c.onToggle == nil {
			t.Fatal("onToggle should not be nil")
		}
		c.onToggle(true)
		if !called {
			t.Error("onToggle should have been called")
		}
	})

	t.Run("Disabled", func(t *testing.T) {
		var c config
		Disabled(true)(&c)
		if !c.disabled {
			t.Error("disabled should be true")
		}
	})

	t.Run("DisabledFn", func(t *testing.T) {
		var c config
		DisabledFn(func() bool { return true })(&c)
		if c.disabledFn == nil {
			t.Error("disabledFn should not be nil")
		}
	})

	t.Run("Indeterminate", func(t *testing.T) {
		var c config
		Indeterminate(true)(&c)
		if !c.indeterminate {
			t.Error("indeterminate should be true")
		}
	})

	t.Run("A11yHint", func(t *testing.T) {
		var c config
		A11yHint("Toggle checkbox")(&c)
		if c.a11yHint != "Toggle checkbox" {
			t.Errorf("a11yHint = %q, want %q", c.a11yHint, "Toggle checkbox")
		}
	})

	t.Run("BackgroundOpt", func(t *testing.T) {
		var c config
		color := widget.ColorRed
		BackgroundOpt(color)(&c)
		if c.background == nil {
			t.Fatal("background should not be nil")
		}
		if *c.background != color {
			t.Errorf("background = %v, want %v", *c.background, color)
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
}

// --- Widget Construction Tests ---

func TestNew_InternalDefaults(t *testing.T) {
	w := New()

	if !w.IsVisible() {
		t.Error("should be visible by default")
	}
	if !w.IsEnabled() {
		t.Error("should be enabled by default")
	}
	if !w.IsFocusable() {
		t.Error("should be focusable by default")
	}
	if w.state != stateNormal {
		t.Errorf("state = %v, want %v", w.state, stateNormal)
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
	custom := &internalTestPainter{}
	w := New(PainterOpt(custom))
	if w.painter != custom {
		t.Error("painter should be the custom painter")
	}
}

func TestNew_WithOptions(t *testing.T) {
	toggled := false
	w := New(
		LabelOpt("Accept"),
		OnToggle(func(bool) { toggled = true }),
		Checked(true),
		Disabled(false),
		A11yHint("toggle terms"),
	)

	if w.cfg.ResolvedLabel() != "Accept" {
		t.Errorf("label = %q, want %q", w.cfg.ResolvedLabel(), "Accept")
	}
	if !w.cfg.ResolvedChecked() {
		t.Error("checked should be true")
	}
	if w.cfg.a11yHint != "toggle terms" {
		t.Errorf("a11yHint = %q, want %q", w.cfg.a11yHint, "toggle terms")
	}

	w.cfg.onToggle(true)
	if !toggled {
		t.Error("toggle handler should have been called")
	}
}

func TestNew_Children(t *testing.T) {
	w := New()
	if children := w.Children(); children != nil {
		t.Errorf("Children() = %v, want nil", children)
	}
}

// --- IsFocusable Tests ---

func TestIsFocusable(t *testing.T) {
	tests := []struct {
		name          string
		visible       bool
		enabled       bool
		disabled      bool
		wantFocusable bool
	}{
		{"all true", true, true, false, true},
		{"invisible", false, true, false, false},
		{"not enabled", true, false, false, false},
		{"disabled", true, true, true, false},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			w := New(Disabled(tc.disabled))
			w.SetVisible(tc.visible)
			w.SetEnabled(tc.enabled)

			if got := w.IsFocusable(); got != tc.wantFocusable {
				t.Errorf("IsFocusable() = %v, want %v", got, tc.wantFocusable)
			}
		})
	}
}

func TestIsFocusable_DisabledFn(t *testing.T) {
	isDisabled := true
	w := New(DisabledFn(func() bool { return isDisabled }))

	if w.IsFocusable() {
		t.Error("should not be focusable when DisabledFn returns true")
	}

	isDisabled = false
	if !w.IsFocusable() {
		t.Error("should be focusable when DisabledFn returns false")
	}
}

// --- Layout Tests ---

func TestLayout_MinHeight(t *testing.T) {
	ctx := widget.NewContext()
	constraints := geometry.Loose(geometry.Sz(400, 400))

	w := New()
	size := w.Layout(ctx, constraints)

	if size.Height < minHeight {
		t.Errorf("height = %v, should be at least %v", size.Height, minHeight)
	}
}

func TestLayout_WithLabel(t *testing.T) {
	ctx := widget.NewContext()
	constraints := geometry.Loose(geometry.Sz(400, 400))

	noLabel := New()
	sizeNoLabel := noLabel.Layout(ctx, constraints)

	withLabel := New(LabelOpt("Some Label"))
	sizeWithLabel := withLabel.Layout(ctx, constraints)

	if sizeWithLabel.Width <= sizeNoLabel.Width {
		t.Error("checkbox with label should be wider than without")
	}
}

func TestLayout_Constrained(t *testing.T) {
	ctx := widget.NewContext()
	constraints := geometry.Tight(geometry.Sz(50, 30))

	w := New(LabelOpt("Test"))
	size := w.Layout(ctx, constraints)

	if size.Width != 50 {
		t.Errorf("width = %v, want 50 (tight constraint)", size.Width)
	}
	if size.Height != 30 {
		t.Errorf("height = %v, want 30 (tight constraint)", size.Height)
	}
}

// --- Styling Tests ---

func TestStyling_Chaining(t *testing.T) {
	w := New(LabelOpt("Test"))
	result := w.Padding(12).SetBackground(widget.ColorRed)

	if result != w {
		t.Error("fluent methods should return the same widget for chaining")
	}
	if w.padding != 12 {
		t.Errorf("padding = %v, want 12", w.padding)
	}
	if w.cfg.background == nil || *w.cfg.background != widget.ColorRed {
		t.Error("background should be ColorRed")
	}
}

func TestPadding(t *testing.T) {
	w := New()
	w.Padding(16)
	if w.padding != 16 {
		t.Errorf("Padding(16) = %v, want 16", w.padding)
	}
}

// --- State Transition Tests ---

func TestStateTransition_HoverEnterLeave(t *testing.T) {
	w := New(LabelOpt("Test"))
	w.SetBounds(geometry.NewRect(0, 0, 100, 40))
	ctx := widget.NewContext()

	enterEvt := event.NewMouseEvent(event.MouseEnter, event.ButtonNone, 0,
		geometry.Pt(10, 20), geometry.Pt(10, 20), event.ModNone)
	consumed := handleEvent(w, ctx, enterEvt)

	if !consumed {
		t.Error("MouseEnter should be consumed")
	}
	if w.state != stateHover {
		t.Errorf("state = %v, want stateHover", w.state)
	}

	leaveEvt := event.NewMouseEvent(event.MouseLeave, event.ButtonNone, 0,
		geometry.Pt(150, 20), geometry.Pt(150, 20), event.ModNone)
	consumed = handleEvent(w, ctx, leaveEvt)

	if !consumed {
		t.Error("MouseLeave should be consumed")
	}
	if w.state != stateNormal {
		t.Errorf("state = %v, want stateNormal", w.state)
	}
}

func TestStateTransition_PressRelease(t *testing.T) {
	toggled := false
	w := New(LabelOpt("Test"), OnToggle(func(bool) { toggled = true }))
	w.SetBounds(geometry.NewRect(0, 0, 100, 40))
	ctx := widget.NewContext()

	pressEvt := event.NewMouseEvent(event.MousePress, event.ButtonLeft, event.ButtonStateLeft,
		geometry.Pt(10, 20), geometry.Pt(10, 20), event.ModNone)
	consumed := handleEvent(w, ctx, pressEvt)

	if !consumed {
		t.Error("MousePress should be consumed")
	}
	if w.state != statePressed {
		t.Errorf("state = %v, want statePressed", w.state)
	}
	if toggled {
		t.Error("should not toggle on press, only on release")
	}

	releaseEvt := event.NewMouseEvent(event.MouseRelease, event.ButtonLeft, 0,
		geometry.Pt(10, 20), geometry.Pt(10, 20), event.ModNone)
	consumed = handleEvent(w, ctx, releaseEvt)

	if !consumed {
		t.Error("MouseRelease should be consumed")
	}
	if !toggled {
		t.Error("should have toggled on release inside bounds")
	}
	if w.state != stateHover {
		t.Errorf("state = %v, want stateHover (released inside bounds)", w.state)
	}
}

func TestStateTransition_PressReleaseOutside(t *testing.T) {
	toggled := false
	w := New(LabelOpt("Test"), OnToggle(func(bool) { toggled = true }))
	w.SetBounds(geometry.NewRect(0, 0, 100, 40))
	ctx := widget.NewContext()

	pressEvt := event.NewMouseEvent(event.MousePress, event.ButtonLeft, event.ButtonStateLeft,
		geometry.Pt(10, 20), geometry.Pt(10, 20), event.ModNone)
	handleEvent(w, ctx, pressEvt)

	releaseEvt := event.NewMouseEvent(event.MouseRelease, event.ButtonLeft, 0,
		geometry.Pt(200, 200), geometry.Pt(200, 200), event.ModNone)
	handleEvent(w, ctx, releaseEvt)

	if toggled {
		t.Error("should not toggle when released outside bounds")
	}
	if w.state != stateNormal {
		t.Errorf("state = %v, want stateNormal (released outside)", w.state)
	}
}

func TestStateTransition_RightButton_Ignored(t *testing.T) {
	w := New(LabelOpt("Test"))
	w.SetBounds(geometry.NewRect(0, 0, 100, 40))
	ctx := widget.NewContext()

	pressEvt := event.NewMouseEvent(event.MousePress, event.ButtonRight, event.ButtonStateRight,
		geometry.Pt(10, 20), geometry.Pt(10, 20), event.ModNone)
	consumed := handleEvent(w, ctx, pressEvt)

	if consumed {
		t.Error("right button press should not be consumed")
	}
	if w.state != stateNormal {
		t.Error("state should remain normal for right button")
	}
}

// --- Disabled State Tests ---

func TestDisabled_IgnoresEvents(t *testing.T) {
	toggled := false
	w := New(LabelOpt("Test"), OnToggle(func(bool) { toggled = true }), Disabled(true))
	w.SetBounds(geometry.NewRect(0, 0, 100, 40))
	ctx := widget.NewContext()

	enterEvt := event.NewMouseEvent(event.MouseEnter, event.ButtonNone, 0,
		geometry.Pt(10, 20), geometry.Pt(10, 20), event.ModNone)
	consumed := handleEvent(w, ctx, enterEvt)
	if consumed {
		t.Error("disabled checkbox should not consume MouseEnter")
	}

	pressEvt := event.NewMouseEvent(event.MousePress, event.ButtonLeft, event.ButtonStateLeft,
		geometry.Pt(10, 20), geometry.Pt(10, 20), event.ModNone)
	consumed = handleEvent(w, ctx, pressEvt)
	if consumed {
		t.Error("disabled checkbox should not consume MousePress")
	}

	if toggled {
		t.Error("disabled checkbox should not fire toggle")
	}
}

func TestDisabledFn_Reactive(t *testing.T) {
	isDisabled := false
	toggled := false
	w := New(
		LabelOpt("Test"),
		OnToggle(func(bool) { toggled = true }),
		DisabledFn(func() bool { return isDisabled }),
	)
	w.SetBounds(geometry.NewRect(0, 0, 100, 40))
	ctx := widget.NewContext()

	pressEvt := event.NewMouseEvent(event.MousePress, event.ButtonLeft, event.ButtonStateLeft,
		geometry.Pt(10, 20), geometry.Pt(10, 20), event.ModNone)
	handleEvent(w, ctx, pressEvt)

	releaseEvt := event.NewMouseEvent(event.MouseRelease, event.ButtonLeft, 0,
		geometry.Pt(10, 20), geometry.Pt(10, 20), event.ModNone)
	handleEvent(w, ctx, releaseEvt)

	if !toggled {
		t.Error("should toggle when not disabled")
	}

	toggled = false
	isDisabled = true

	pressEvt2 := event.NewMouseEvent(event.MousePress, event.ButtonLeft, event.ButtonStateLeft,
		geometry.Pt(10, 20), geometry.Pt(10, 20), event.ModNone)
	consumed := handleEvent(w, ctx, pressEvt2)

	if consumed {
		t.Error("disabled checkbox should not consume events")
	}
	if toggled {
		t.Error("disabled checkbox should not fire toggle")
	}
}

// --- Keyboard Activation Tests ---

func TestKeyboard_SpaceActivation(t *testing.T) {
	toggled := false
	w := New(LabelOpt("Test"), OnToggle(func(bool) { toggled = true }))
	w.SetFocused(true)
	ctx := widget.NewContext()

	pressEvt := event.NewKeyEvent(event.KeyPress, event.KeySpace, 0, event.ModNone)
	consumed := handleEvent(w, ctx, pressEvt)
	if !consumed {
		t.Error("Space press should be consumed when focused")
	}
	if w.state != statePressed {
		t.Errorf("state = %v, want statePressed", w.state)
	}

	releaseEvt := event.NewKeyEvent(event.KeyRelease, event.KeySpace, 0, event.ModNone)
	consumed = handleEvent(w, ctx, releaseEvt)
	if !consumed {
		t.Error("Space release should be consumed when focused")
	}
	if !toggled {
		t.Error("Space release should trigger toggle")
	}
	if w.state != stateNormal {
		t.Errorf("state = %v, want stateNormal after key release", w.state)
	}
}

func TestKeyboard_NotFocused_Ignored(t *testing.T) {
	toggled := false
	w := New(LabelOpt("Test"), OnToggle(func(bool) { toggled = true }))
	w.SetFocused(false)
	ctx := widget.NewContext()

	pressEvt := event.NewKeyEvent(event.KeyPress, event.KeySpace, 0, event.ModNone)
	consumed := handleEvent(w, ctx, pressEvt)

	if consumed {
		t.Error("Space should not be consumed when not focused")
	}
	if toggled {
		t.Error("should not toggle when not focused")
	}
}

func TestKeyboard_OtherKeys_Ignored(t *testing.T) {
	w := New(LabelOpt("Test"))
	w.SetFocused(true)
	ctx := widget.NewContext()

	pressEvt := event.NewKeyEvent(event.KeyPress, event.KeyA, 'a', event.ModNone)
	consumed := handleEvent(w, ctx, pressEvt)

	if consumed {
		t.Error("key A should not be consumed by checkbox")
	}
}

func TestKeyboard_EnterKey_Ignored(t *testing.T) {
	toggled := false
	w := New(LabelOpt("Test"), OnToggle(func(bool) { toggled = true }))
	w.SetFocused(true)
	ctx := widget.NewContext()

	pressEvt := event.NewKeyEvent(event.KeyPress, event.KeyEnter, 0, event.ModNone)
	consumed := handleEvent(w, ctx, pressEvt)

	if consumed {
		t.Error("Enter key should not be consumed by checkbox")
	}
	if toggled {
		t.Error("Enter key should not trigger toggle")
	}
}

func TestKeyboard_Disabled_Ignored(t *testing.T) {
	toggled := false
	w := New(LabelOpt("Test"), OnToggle(func(bool) { toggled = true }), Disabled(true))
	w.SetFocused(true)
	ctx := widget.NewContext()

	pressEvt := event.NewKeyEvent(event.KeyPress, event.KeySpace, 0, event.ModNone)
	consumed := handleEvent(w, ctx, pressEvt)

	if consumed {
		t.Error("disabled checkbox should not consume key events")
	}
	if toggled {
		t.Error("disabled checkbox should not fire toggle")
	}
}

// --- Toggle Logic Tests ---

func TestFireToggle_NilHandler(t *testing.T) {
	w := New(LabelOpt("Test"))
	// Should not panic.
	fireToggle(w)
}

func TestFireToggle_UpdatesCheckedState(t *testing.T) {
	w := New()

	if w.cfg.ResolvedChecked() {
		t.Error("initial state should be unchecked")
	}

	fireToggle(w)

	if !w.cfg.ResolvedChecked() {
		t.Error("should be checked after toggle")
	}

	fireToggle(w)

	if w.cfg.ResolvedChecked() {
		t.Error("should be unchecked after second toggle")
	}
}

func TestFireToggle_DoesNotUpdateWithCheckedFn(t *testing.T) {
	externalState := false
	w := New(CheckedFn(func() bool { return externalState }))

	fireToggle(w)

	// With CheckedFn, the static field should NOT be updated.
	// The resolved checked state should still come from externalState.
	if w.cfg.ResolvedChecked() {
		t.Error("with CheckedFn, toggle should not change static checked field")
	}
}

func TestFireToggle_ClearsIndeterminate(t *testing.T) {
	w := New(Indeterminate(true))

	if !w.cfg.indeterminate {
		t.Error("should start indeterminate")
	}

	fireToggle(w)

	if w.cfg.indeterminate {
		t.Error("indeterminate should be cleared after toggle")
	}
}

func TestFireToggle_CallsOnToggle(t *testing.T) {
	var receivedState bool
	callCount := 0
	w := New(OnToggle(func(checked bool) {
		receivedState = checked
		callCount++
	}))

	fireToggle(w)
	if callCount != 1 {
		t.Errorf("onToggle called %d times, want 1", callCount)
	}
	if !receivedState {
		t.Error("onToggle should receive true (toggled from unchecked)")
	}

	fireToggle(w)
	if callCount != 2 {
		t.Errorf("onToggle called %d times, want 2", callCount)
	}
	if receivedState {
		t.Error("onToggle should receive false (toggled from checked)")
	}
}

// --- Draw Tests ---

func TestDraw_EmptyBounds(t *testing.T) {
	w := New(LabelOpt("Test"))
	ctx := widget.NewContext()
	canvas := &internalMockCanvas{}

	w.Draw(ctx, canvas)

	if len(canvas.drawRoundRects) > 0 || len(canvas.strokeRoundRects) > 0 ||
		len(canvas.drawTexts) > 0 || len(canvas.drawLines) > 0 {
		t.Error("should not draw anything with empty bounds")
	}
}

func TestDraw_UncheckedState(t *testing.T) {
	w := New(LabelOpt("Click"))
	w.SetBounds(geometry.NewRect(10, 10, 100, 40))
	ctx := widget.NewContext()
	canvas := &internalMockCanvas{}

	w.Draw(ctx, canvas)

	if len(canvas.strokeRoundRects) == 0 {
		t.Error("unchecked checkbox should draw border (StrokeRoundRect)")
	}
	if len(canvas.drawTexts) == 0 {
		t.Error("should draw label text")
	}
	if canvas.drawTexts[0].text != "Click" {
		t.Errorf("text = %q, want %q", canvas.drawTexts[0].text, "Click")
	}
}

func TestDraw_CheckedState(t *testing.T) {
	w := New(LabelOpt("Click"), Checked(true))
	w.SetBounds(geometry.NewRect(10, 10, 100, 40))
	ctx := widget.NewContext()
	canvas := &internalMockCanvas{}

	w.Draw(ctx, canvas)

	if len(canvas.drawRoundRects) == 0 {
		t.Error("checked checkbox should draw filled box (DrawRoundRect)")
	}
	if len(canvas.drawLines) == 0 {
		t.Error("checked checkbox should draw checkmark lines")
	}
	if len(canvas.drawTexts) == 0 {
		t.Error("should draw label text")
	}
}

func TestDraw_IndeterminateState(t *testing.T) {
	w := New(LabelOpt("Click"), Indeterminate(true))
	w.SetBounds(geometry.NewRect(10, 10, 100, 40))
	ctx := widget.NewContext()
	canvas := &internalMockCanvas{}

	w.Draw(ctx, canvas)

	if len(canvas.drawRoundRects) == 0 {
		t.Error("indeterminate checkbox should draw filled box (DrawRoundRect)")
	}
	if len(canvas.drawLines) == 0 {
		t.Error("indeterminate checkbox should draw dash line")
	}
}

func TestDraw_NoLabel(t *testing.T) {
	w := New()
	w.SetBounds(geometry.NewRect(10, 10, 30, 30))
	ctx := widget.NewContext()
	canvas := &internalMockCanvas{}

	w.Draw(ctx, canvas)

	if len(canvas.drawTexts) > 0 {
		t.Error("checkbox without label should not draw text")
	}
}

func TestDraw_DefaultPainter_DisabledChecked(t *testing.T) {
	w := New(LabelOpt("Click"), Checked(true), Disabled(true))
	w.SetBounds(geometry.NewRect(0, 0, 100, 40))
	ctx := widget.NewContext()
	canvas := &internalMockCanvas{}

	w.Draw(ctx, canvas)

	if len(canvas.drawRoundRects) == 0 {
		t.Fatal("should draw filled box even when disabled")
	}
	bg := canvas.drawRoundRects[0].color
	if bg == defaultCheckedBg {
		t.Error("disabled background should differ from normal checked background")
	}
}

func TestDraw_DefaultPainter_DisabledUnchecked(t *testing.T) {
	w := New(LabelOpt("Click"), Disabled(true))
	w.SetBounds(geometry.NewRect(0, 0, 100, 40))
	ctx := widget.NewContext()
	canvas := &internalMockCanvas{}

	w.Draw(ctx, canvas)

	if len(canvas.strokeRoundRects) == 0 {
		t.Error("unchecked disabled should draw border")
	}
	borderColor := canvas.strokeRoundRects[0].color
	if borderColor == defaultUncheckedBorder {
		t.Error("disabled border should differ from normal border")
	}
}

func TestDraw_FocusRing(t *testing.T) {
	w := New(LabelOpt("Click"))
	w.SetBounds(geometry.NewRect(10, 10, 100, 40))
	w.SetFocused(true)
	ctx := widget.NewContext()
	canvas := &internalMockCanvas{}

	w.Draw(ctx, canvas)

	// Focus ring should be drawn around the checkbox box.
	found := false
	for _, call := range canvas.strokeRoundRects {
		// Focus ring is expanded beyond the box area.
		if call.r.Min.X < 10 {
			found = true
			break
		}
	}
	if !found {
		t.Error("focused checkbox should draw a focus ring (expanded stroke)")
	}
}

func TestDraw_FocusRing_NotWhenDisabled(t *testing.T) {
	w := New(LabelOpt("Click"), Disabled(true))
	w.SetBounds(geometry.NewRect(10, 10, 100, 40))
	w.SetFocused(true)
	ctx := widget.NewContext()
	canvas := &internalMockCanvas{}

	w.Draw(ctx, canvas)

	for _, call := range canvas.strokeRoundRects {
		if call.r.Min.X < 10 {
			t.Error("disabled focused checkbox should not draw focus ring")
			break
		}
	}
}

func TestDraw_CustomBackground(t *testing.T) {
	color := widget.ColorGreen
	w := New(LabelOpt("Click"), Checked(true))
	w.SetBackground(color)
	w.SetBounds(geometry.NewRect(0, 0, 100, 40))
	ctx := widget.NewContext()
	canvas := &internalMockCanvas{}

	w.Draw(ctx, canvas)

	if len(canvas.drawRoundRects) == 0 {
		t.Fatal("should draw filled box")
	}
	if canvas.drawRoundRects[0].color != color {
		t.Errorf("background = %v, want %v", canvas.drawRoundRects[0].color, color)
	}
}

// --- Painting Helper Tests ---

func TestApplyStateModifier(t *testing.T) {
	base := widget.ColorRed

	normal := applyStateModifier(base, false, false)
	if normal != base {
		t.Error("normal state should return base color unchanged")
	}

	hover := applyStateModifier(base, true, false)
	if hover == base {
		t.Error("hover state should modify the base color")
	}

	pressed := applyStateModifier(base, false, true)
	if pressed == base {
		t.Error("pressed state should modify the base color")
	}
}

func TestApplyStateModifier_PressedOverridesHover(t *testing.T) {
	base := widget.ColorRed

	result := applyStateModifier(base, true, true)
	pressedOnly := applyStateModifier(base, false, true)
	if result != pressedOnly {
		t.Error("pressed should take precedence over hovered")
	}
}

func TestCheckboxBoxRect(t *testing.T) {
	bounds := geometry.NewRect(10, 10, 100, 40)
	box := checkboxBoxRect(bounds)

	if box.Width() != boxSize {
		t.Errorf("box width = %v, want %v", box.Width(), boxSize)
	}
	if box.Height() != boxSize {
		t.Errorf("box height = %v, want %v", box.Height(), boxSize)
	}
	// Box is inset by borderStrokeWidth/2 so centered strokes stay within bounds (#117).
	expectedX := float32(10) + borderStrokeWidth/2
	if box.Min.X != expectedX {
		t.Errorf("box X = %v, want %v", box.Min.X, expectedX)
	}
	// Verify stroke never extends outside widget bounds.
	strokeLeft := box.Min.X - borderStrokeWidth/2
	if strokeLeft < bounds.Min.X {
		t.Errorf("stroke left edge (%v) extends outside bounds.Min.X (%v)", strokeLeft, bounds.Min.X)
	}
}

func TestCheckboxLabelBounds(t *testing.T) {
	bounds := geometry.NewRect(10, 10, 100, 40)
	label := checkboxLabelBounds(bounds)

	offset := borderStrokeWidth / 2
	expectedX := float32(10) + offset + boxSize + labelGap
	if label.Min.X != expectedX {
		t.Errorf("label X = %v, want %v", label.Min.X, expectedX)
	}
	if label.Min.Y != 10 {
		t.Errorf("label Y = %v, want 10", label.Min.Y)
	}
}

// --- Painter Interface Tests ---

func TestPainter_DelegationToPainter(t *testing.T) {
	p := &internalTestPainter{}
	w := New(LabelOpt("Test"), PainterOpt(p))
	w.SetBounds(geometry.NewRect(0, 0, 100, 40))
	ctx := widget.NewContext()
	canvas := &internalMockCanvas{}

	w.Draw(ctx, canvas)

	if !p.called {
		t.Error("Draw should delegate to the configured painter")
	}
	if p.state.Label != "Test" {
		t.Errorf("PaintState.Label = %q, want %q", p.state.Label, "Test")
	}
	if p.state.Bounds.IsEmpty() {
		t.Error("PaintState.Bounds should not be empty")
	}
}

func TestPainter_PaintStateFields(t *testing.T) {
	p := &internalTestPainter{}
	w := New(
		LabelOpt("Test"),
		Checked(true),
		Indeterminate(true),
		PainterOpt(p),
	)
	w.SetBounds(geometry.NewRect(0, 0, 100, 40))
	w.SetFocused(true)
	ctx := widget.NewContext()
	canvas := &internalMockCanvas{}

	w.Draw(ctx, canvas)

	if !p.state.Checked {
		t.Error("PaintState.Checked should be true")
	}
	if !p.state.Indeterminate {
		t.Error("PaintState.Indeterminate should be true")
	}
	if !p.state.Focused {
		t.Error("PaintState.Focused should be true")
	}
}

// --- internalTestPainter records the call to PaintCheckbox ---

type internalTestPainter struct {
	called bool
	state  PaintState
}

func (p *internalTestPainter) PaintCheckbox(_ widget.Canvas, ps PaintState) {
	p.called = true
	p.state = ps
}

// --- internalMockCanvas records canvas calls for testing ---

type internalMockCanvas struct {
	drawRoundRects   []internalDrawRoundRectCall
	strokeRoundRects []internalStrokeRoundRectCall
	drawTexts        []internalDrawTextCall
	drawLines        []internalDrawLineCall
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

type internalDrawLineCall struct {
	from, to    geometry.Point
	color       widget.Color
	strokeWidth float32
}

func (c *internalMockCanvas) Clear(_ widget.Color)                                  {}
func (c *internalMockCanvas) DrawRect(_ geometry.Rect, _ widget.Color)              {}
func (c *internalMockCanvas) FillRectDirect(_ geometry.Rect, _ widget.Color)        {}
func (c *internalMockCanvas) StrokeRect(_ geometry.Rect, _ widget.Color, _ float32) {}

func (c *internalMockCanvas) DrawRoundRect(r geometry.Rect, color widget.Color, radius float32) {
	c.drawRoundRects = append(c.drawRoundRects, internalDrawRoundRectCall{r: r, color: color, radius: radius})
}

func (c *internalMockCanvas) StrokeRoundRect(r geometry.Rect, color widget.Color, radius float32, strokeWidth float32) {
	c.strokeRoundRects = append(c.strokeRoundRects, internalStrokeRoundRectCall{r: r, color: color, radius: radius, strokeWidth: strokeWidth})
}

func (c *internalMockCanvas) DrawCircle(_ geometry.Point, _ float32, _ widget.Color)              {}
func (c *internalMockCanvas) StrokeCircle(_ geometry.Point, _ float32, _ widget.Color, _ float32) {}
func (c *internalMockCanvas) StrokeArc(_ geometry.Point, _ float32, _, _ float64, _ widget.Color, _ float32) {
}

func (c *internalMockCanvas) DrawLine(from, to geometry.Point, color widget.Color, strokeWidth float32) {
	c.drawLines = append(c.drawLines, internalDrawLineCall{from: from, to: to, color: color, strokeWidth: strokeWidth})
}

func (c *internalMockCanvas) DrawText(text string, bounds geometry.Rect, fontSize float32, color widget.Color, bold bool, align widget.TextAlign) {
	c.drawTexts = append(c.drawTexts, internalDrawTextCall{text: text, bounds: bounds, fontSize: fontSize, color: color, bold: bold, align: align})
}

func (c *internalMockCanvas) MeasureText(text string, fontSize float32, _ bool) float32 {
	return float32(len([]rune(text))) * fontSize * 0.5
}

func (c *internalMockCanvas) DrawImage(_ image.Image, _ geometry.Point)    {}
func (c *internalMockCanvas) PushClip(_ geometry.Rect)                     {}
func (c *internalMockCanvas) PushClipRoundRect(_ geometry.Rect, _ float32) {}
func (c *internalMockCanvas) PopClip()                                     {}
func (c *internalMockCanvas) PushTransform(_ geometry.Point)               {}
func (c *internalMockCanvas) PopTransform()                                {}
func (c *internalMockCanvas) TransformOffset() geometry.Point              { return geometry.Point{} }
func (c *internalMockCanvas) ScreenOriginBase() geometry.Point             { return geometry.Point{} }
func (c *internalMockCanvas) ClipBounds() geometry.Rect                    { return geometry.NewRect(0, 0, 10000, 10000) }
func (c *internalMockCanvas) ReplayScene(_ *scene.Scene)                   {}

// --- Signal Binding Tests ---

func TestConfig_ResolvedChecked_Signal(t *testing.T) {
	sig := state.NewSignal(true)
	c := config{checked: false, checkedFn: func() bool { return false }, checkedSignal: sig}

	if !c.ResolvedChecked() {
		t.Error("ResolvedChecked() should be true (signal value)")
	}

	sig.Set(false)
	if c.ResolvedChecked() {
		t.Error("ResolvedChecked() should be false after signal update")
	}
}

func TestConfig_ResolvedLabel_Signal(t *testing.T) {
	sig := state.NewSignal("from signal")
	c := config{label: "static", labelFn: func() string { return "from fn" }, labelSignal: sig}

	if got := c.ResolvedLabel(); got != "from signal" {
		t.Errorf("ResolvedLabel() = %q, want %q", got, "from signal")
	}

	sig.Set("updated")
	if got := c.ResolvedLabel(); got != "updated" {
		t.Errorf("ResolvedLabel() = %q, want %q", got, "updated")
	}
}

func TestConfig_ResolvedDisabled_Signal(t *testing.T) {
	sig := state.NewSignal(true)
	c := config{disabled: false, disabledFn: func() bool { return false }, disabledSignal: sig}

	if !c.ResolvedDisabled() {
		t.Error("ResolvedDisabled() should be true (signal value)")
	}

	sig.Set(false)
	if c.ResolvedDisabled() {
		t.Error("ResolvedDisabled() should be false after signal update")
	}
}

func TestConfig_ResolvedChecked_SignalPriority(t *testing.T) {
	sig := state.NewSignal(true)

	t.Run("signal beats fn and static", func(t *testing.T) {
		c := config{checked: false, checkedFn: func() bool { return false }, checkedSignal: sig}
		if !c.ResolvedChecked() {
			t.Error("signal(true) should override fn(false) and static(false)")
		}
	})

	t.Run("fn beats static when no signal", func(t *testing.T) {
		c := config{checked: false, checkedFn: func() bool { return true }}
		if !c.ResolvedChecked() {
			t.Error("fn(true) should override static(false)")
		}
	})

	t.Run("static used when no signal or fn", func(t *testing.T) {
		c := config{checked: true}
		if !c.ResolvedChecked() {
			t.Error("static(true) should be used when no signal or fn")
		}
	})
}

func TestConfig_ResolvedLabel_SignalPriority(t *testing.T) {
	sig := state.NewSignal("signal")

	t.Run("signal beats fn and static", func(t *testing.T) {
		c := config{label: "static", labelFn: func() string { return "fn" }, labelSignal: sig}
		if got := c.ResolvedLabel(); got != "signal" {
			t.Errorf("ResolvedLabel() = %q, want %q (signal > fn > static)", got, "signal")
		}
	})

	t.Run("fn beats static when no signal", func(t *testing.T) {
		c := config{label: "static", labelFn: func() string { return "fn" }}
		if got := c.ResolvedLabel(); got != "fn" {
			t.Errorf("ResolvedLabel() = %q, want %q", got, "fn")
		}
	})

	t.Run("static used when no signal or fn", func(t *testing.T) {
		c := config{label: "static"}
		if got := c.ResolvedLabel(); got != "static" {
			t.Errorf("ResolvedLabel() = %q, want %q", got, "static")
		}
	})
}

func TestConfig_ResolvedDisabled_SignalPriority(t *testing.T) {
	sig := state.NewSignal(true)

	t.Run("signal beats fn and static", func(t *testing.T) {
		c := config{disabled: false, disabledFn: func() bool { return false }, disabledSignal: sig}
		if !c.ResolvedDisabled() {
			t.Error("signal(true) should override fn(false) and static(false)")
		}
	})

	t.Run("fn beats static when no signal", func(t *testing.T) {
		c := config{disabled: false, disabledFn: func() bool { return true }}
		if !c.ResolvedDisabled() {
			t.Error("fn(true) should override static(false)")
		}
	})

	t.Run("static used when no signal or fn", func(t *testing.T) {
		c := config{disabled: true}
		if !c.ResolvedDisabled() {
			t.Error("static(true) should be used when no signal or fn")
		}
	})
}

func TestCheckboxCheckedSignal_TwoWay(t *testing.T) {
	sig := state.NewSignal(false)
	w := New(CheckedSignal(sig))
	w.SetBounds(geometry.NewRect(0, 0, 100, 40))

	// Signal → widget: signal is false, widget reads false.
	if w.cfg.ResolvedChecked() {
		t.Error("should be unchecked initially (signal=false)")
	}

	// Signal → widget: update signal to true, widget reads true.
	sig.Set(true)
	if !w.cfg.ResolvedChecked() {
		t.Error("should be checked after signal.Set(true)")
	}

	// Widget → signal: toggle writes back to signal.
	fireToggle(w) // toggling from true → false
	if sig.Get() {
		t.Error("signal should be false after toggle from checked")
	}
	if w.cfg.ResolvedChecked() {
		t.Error("widget should read false after toggle")
	}

	// Toggle again: false → true.
	fireToggle(w)
	if !sig.Get() {
		t.Error("signal should be true after toggle from unchecked")
	}
}

func TestCheckboxCheckedSignal_TwoWay_MouseClick(t *testing.T) {
	sig := state.NewSignal(false)
	w := New(CheckedSignal(sig))
	w.SetBounds(geometry.NewRect(0, 0, 100, 40))
	ctx := widget.NewContext()

	// Click to toggle.
	press := event.NewMouseEvent(event.MousePress, event.ButtonLeft, event.ButtonStateLeft,
		geometry.Pt(10, 20), geometry.Pt(10, 20), event.ModNone)
	handleEvent(w, ctx, press)

	release := event.NewMouseEvent(event.MouseRelease, event.ButtonLeft, 0,
		geometry.Pt(10, 20), geometry.Pt(10, 20), event.ModNone)
	handleEvent(w, ctx, release)

	if !sig.Get() {
		t.Error("signal should be true after mouse click toggle")
	}
}

func TestCheckboxCheckedSignal_TwoWay_Keyboard(t *testing.T) {
	sig := state.NewSignal(false)
	w := New(CheckedSignal(sig))
	w.SetFocused(true)
	ctx := widget.NewContext()

	press := event.NewKeyEvent(event.KeyPress, event.KeySpace, 0, event.ModNone)
	handleEvent(w, ctx, press)

	release := event.NewKeyEvent(event.KeyRelease, event.KeySpace, 0, event.ModNone)
	handleEvent(w, ctx, release)

	if !sig.Get() {
		t.Error("signal should be true after keyboard toggle")
	}
}

func TestCheckboxLabelSignal(t *testing.T) {
	sig := state.NewSignal("Signal Label")
	w := New(LabelSignal(sig))
	w.SetBounds(geometry.NewRect(0, 0, 200, 40))
	ctx := widget.NewContext()
	canvas := &internalMockCanvas{}

	w.Draw(ctx, canvas)

	if len(canvas.drawTexts) == 0 {
		t.Fatal("should have drawn label text")
	}
	if canvas.drawTexts[0].text != "Signal Label" {
		t.Errorf("text = %q, want %q", canvas.drawTexts[0].text, "Signal Label")
	}

	// Update signal and redraw.
	sig.Set("Updated Label")
	canvas2 := &internalMockCanvas{}
	w.Draw(ctx, canvas2)

	if len(canvas2.drawTexts) == 0 {
		t.Fatal("should have drawn label text after update")
	}
	if canvas2.drawTexts[0].text != "Updated Label" {
		t.Errorf("text = %q, want %q", canvas2.drawTexts[0].text, "Updated Label")
	}
}

func TestCheckboxDisabledSignal(t *testing.T) {
	sig := state.NewSignal(true)
	w := New(DisabledSignal(sig))

	if w.IsFocusable() {
		t.Error("should not be focusable when DisabledSignal is true")
	}

	sig.Set(false)
	if !w.IsFocusable() {
		t.Error("should be focusable when DisabledSignal is false")
	}
}

func TestCheckboxDisabledSignal_IgnoresEvents(t *testing.T) {
	toggled := false
	sig := state.NewSignal(true)
	w := New(
		LabelOpt("Test"),
		OnToggle(func(bool) { toggled = true }),
		DisabledSignal(sig),
	)
	w.SetBounds(geometry.NewRect(0, 0, 100, 40))
	ctx := widget.NewContext()

	press := event.NewMouseEvent(event.MousePress, event.ButtonLeft, event.ButtonStateLeft,
		geometry.Pt(10, 20), geometry.Pt(10, 20), event.ModNone)
	consumed := handleEvent(w, ctx, press)

	if consumed {
		t.Error("disabled (via signal) checkbox should not consume events")
	}
	if toggled {
		t.Error("disabled (via signal) checkbox should not toggle")
	}

	// Enable and verify interaction works.
	sig.Set(false)
	handleEvent(w, ctx, press)
	release := event.NewMouseEvent(event.MouseRelease, event.ButtonLeft, 0,
		geometry.Pt(10, 20), geometry.Pt(10, 20), event.ModNone)
	handleEvent(w, ctx, release)

	if !toggled {
		t.Error("enabled checkbox should toggle after DisabledSignal set to false")
	}
}

func TestCheckboxCheckedSignal_OnToggleStillFires(t *testing.T) {
	sig := state.NewSignal(false)
	var receivedState bool
	callCount := 0
	w := New(
		CheckedSignal(sig),
		OnToggle(func(checked bool) {
			receivedState = checked
			callCount++
		}),
	)

	fireToggle(w)

	if callCount != 1 {
		t.Errorf("onToggle called %d times, want 1", callCount)
	}
	if !receivedState {
		t.Error("onToggle should receive true")
	}
	if !sig.Get() {
		t.Error("signal should be true after toggle")
	}
}

// --- ReadonlySignal Tests ---

func TestConfig_ResolvedLabel_ReadonlySignal(t *testing.T) {
	computed := state.NewComputed(func() string { return "computed label" })

	c := config{
		label:            "static",
		labelFn:          func() string { return "fn" },
		labelSignal:      state.NewSignal("signal"),
		readonlyLabelSig: computed,
	}

	if got := c.ResolvedLabel(); got != "computed label" {
		t.Errorf("ResolvedLabel() = %q, want %q", got, "computed label")
	}
}

func TestConfig_ResolvedDisabled_ReadonlySignal(t *testing.T) {
	computed := state.NewComputed(func() bool { return true })

	c := config{
		disabled:            false,
		disabledFn:          func() bool { return false },
		disabledSignal:      state.NewSignal(false),
		readonlyDisabledSig: computed,
	}

	if !c.ResolvedDisabled() {
		t.Error("ResolvedDisabled() should be true (readonly computed signal)")
	}
}

// --- Granular Invalidation Tests (TASK-UI-INVAL-001b) ---

func TestGranularInvalidation_Checkbox_HoverEnter(t *testing.T) {
	w := New(Label("Test"))
	w.SetBounds(geometry.NewRect(0, 0, 100, 40))
	ctx := widget.NewContext()

	e := event.NewMouseEvent(event.MouseEnter, event.ButtonNone, 0,
		geometry.Pt(50, 20), geometry.Pt(50, 20), event.ModNone)
	handleEvent(w, ctx, e)

	if ctx.IsInvalidated() {
		t.Error("hover enter should use granular invalidation, not ctx.Invalidate()")
	}
	if !w.NeedsRedraw() {
		t.Error("hover enter should set needsRedraw")
	}
	if ctx.InvalidatedRect().IsEmpty() {
		t.Error("hover enter should trigger InvalidateRect")
	}
}

func TestGranularInvalidation_Checkbox_HoverLeave(t *testing.T) {
	w := New(Label("Test"))
	w.SetBounds(geometry.NewRect(0, 0, 100, 40))
	ctx := widget.NewContext()

	e := event.NewMouseEvent(event.MouseLeave, event.ButtonNone, 0,
		geometry.Pt(150, 20), geometry.Pt(150, 20), event.ModNone)
	handleEvent(w, ctx, e)

	if ctx.IsInvalidated() {
		t.Error("hover leave should use granular invalidation")
	}
	if !w.NeedsRedraw() {
		t.Error("hover leave should set needsRedraw")
	}
}

func TestGranularInvalidation_Checkbox_PressRelease(t *testing.T) {
	toggled := false
	w := New(Label("Test"), OnToggle(func(bool) { toggled = true }))
	w.SetBounds(geometry.NewRect(0, 0, 100, 40))

	// Press.
	ctx := widget.NewContext()
	press := event.NewMouseEvent(event.MousePress, event.ButtonLeft, event.ButtonStateLeft,
		geometry.Pt(50, 20), geometry.Pt(50, 20), event.ModNone)
	handleEvent(w, ctx, press)

	if ctx.IsInvalidated() {
		t.Error("press should use granular invalidation")
	}

	// Release inside.
	ctx = widget.NewContext()
	release := event.NewMouseEvent(event.MouseRelease, event.ButtonLeft, 0,
		geometry.Pt(50, 20), geometry.Pt(50, 20), event.ModNone)
	handleEvent(w, ctx, release)

	if ctx.IsInvalidated() {
		t.Error("release should use granular invalidation")
	}
	if !toggled {
		t.Error("onToggle should still fire on release inside bounds")
	}
}
