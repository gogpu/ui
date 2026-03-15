package slider

import (
	"image"
	"math"
	"testing"

	"github.com/gogpu/ui/event"
	"github.com/gogpu/ui/geometry"
	"github.com/gogpu/ui/state"
	"github.com/gogpu/ui/widget"
)

// --- Orientation Tests ---

func TestOrientation_String(t *testing.T) {
	tests := []struct {
		name string
		o    Orientation
		want string
	}{
		{"horizontal", Horizontal, "Horizontal"},
		{"vertical", Vertical, "Vertical"},
		{"unknown", Orientation(99), "Unknown"},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			if got := tc.o.String(); got != tc.want {
				t.Errorf("Orientation(%d).String() = %q, want %q", tc.o, got, tc.want)
			}
		})
	}
}

// --- Config Tests ---

func TestConfig_ResolvedValue(t *testing.T) {
	t.Run("static value", func(t *testing.T) {
		c := config{value: 42}
		if got := c.ResolvedValue(); got != 42 {
			t.Errorf("ResolvedValue() = %v, want 42", got)
		}
	})

	t.Run("dynamic value", func(t *testing.T) {
		c := config{value: 10, valueFn: func() float32 { return 75 }}
		if got := c.ResolvedValue(); got != 75 {
			t.Errorf("ResolvedValue() = %v, want 75", got)
		}
	})

	t.Run("zero value", func(t *testing.T) {
		c := config{}
		if got := c.ResolvedValue(); got != 0 {
			t.Errorf("ResolvedValue() = %v, want 0", got)
		}
	})
}

func TestConfig_ResolvedValue_Signal(t *testing.T) {
	sig := state.NewSignal[float32](60)
	c := config{value: 10, valueFn: func() float32 { return 20 }, valueSignal: sig}

	if got := c.ResolvedValue(); got != 60 {
		t.Errorf("ResolvedValue() = %v, want 60 (signal value)", got)
	}

	sig.Set(80)
	if got := c.ResolvedValue(); got != 80 {
		t.Errorf("ResolvedValue() = %v, want 80 after signal update", got)
	}
}

func TestConfig_ResolvedValue_ReadonlySignal(t *testing.T) {
	computed := state.NewComputed(func() float32 { return 99 })

	c := config{
		value:               10,
		valueFn:             func() float32 { return 20 },
		valueSignal:         state.NewSignal[float32](30),
		readonlyValueSignal: computed,
	}

	if got := c.ResolvedValue(); got != 99 {
		t.Errorf("ResolvedValue() = %v, want 99 (readonly signal)", got)
	}
}

func TestConfig_ResolvedValue_Priority(t *testing.T) {
	sig := state.NewSignal[float32](50)

	t.Run("signal beats fn and static", func(t *testing.T) {
		c := config{value: 10, valueFn: func() float32 { return 20 }, valueSignal: sig}
		if got := c.ResolvedValue(); got != 50 {
			t.Errorf("ResolvedValue() = %v, want 50", got)
		}
	})

	t.Run("fn beats static when no signal", func(t *testing.T) {
		c := config{value: 10, valueFn: func() float32 { return 20 }}
		if got := c.ResolvedValue(); got != 20 {
			t.Errorf("ResolvedValue() = %v, want 20", got)
		}
	})

	t.Run("static used when no signal or fn", func(t *testing.T) {
		c := config{value: 10}
		if got := c.ResolvedValue(); got != 10 {
			t.Errorf("ResolvedValue() = %v, want 10", got)
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

func TestConfig_ResolvedDisabled_ReadonlySignal(t *testing.T) {
	computed := state.NewComputed(func() bool { return true })

	c := config{
		disabled:               false,
		disabledFn:             func() bool { return false },
		disabledSignal:         state.NewSignal(false),
		readonlyDisabledSignal: computed,
	}

	if !c.ResolvedDisabled() {
		t.Error("ResolvedDisabled() should be true (readonly computed signal)")
	}
}

// --- Options Tests ---

func TestOptions(t *testing.T) {
	t.Run("Value", func(t *testing.T) {
		var c config
		Value(42)(&c)
		if c.value != 42 {
			t.Errorf("value = %v, want 42", c.value)
		}
	})

	t.Run("ValueFn", func(t *testing.T) {
		var c config
		ValueFn(func() float32 { return 75 })(&c)
		if c.valueFn == nil {
			t.Error("valueFn should not be nil")
		}
		if c.valueFn() != 75 {
			t.Errorf("valueFn() = %v, want 75", c.valueFn())
		}
	})

	t.Run("Min", func(t *testing.T) {
		var c config
		Min(10)(&c)
		if c.minVal != 10 {
			t.Errorf("minVal = %v, want 10", c.minVal)
		}
	})

	t.Run("Max", func(t *testing.T) {
		var c config
		Max(200)(&c)
		if c.maxVal != 200 {
			t.Errorf("maxVal = %v, want 200", c.maxVal)
		}
	})

	t.Run("Step", func(t *testing.T) {
		var c config
		Step(5)(&c)
		if c.step != 5 {
			t.Errorf("step = %v, want 5", c.step)
		}
	})

	t.Run("OnChange", func(t *testing.T) {
		var c config
		called := false
		OnChange(func(float32) { called = true })(&c)
		if c.onChange == nil {
			t.Fatal("onChange should not be nil")
		}
		c.onChange(50)
		if !called {
			t.Error("onChange should have been called")
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

	t.Run("OrientationOpt", func(t *testing.T) {
		var c config
		OrientationOpt(Vertical)(&c)
		if c.orientation != Vertical {
			t.Errorf("orientation = %v, want Vertical", c.orientation)
		}
	})

	t.Run("Marks", func(t *testing.T) {
		var c config
		marks := []Mark{{Value: 25, Label: "25"}, {Value: 75}}
		Marks(marks)(&c)
		if len(c.marks) != 2 {
			t.Errorf("marks count = %d, want 2", len(c.marks))
		}
	})

	t.Run("A11yHint", func(t *testing.T) {
		var c config
		A11yHint("Adjust volume")(&c)
		if c.a11yHint != "Adjust volume" {
			t.Errorf("a11yHint = %q, want %q", c.a11yHint, "Adjust volume")
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

func TestNew_Defaults(t *testing.T) {
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
	if w.interaction != stateNormal {
		t.Errorf("interaction = %v, want stateNormal", w.interaction)
	}
	if w.cfg.maxVal != 100 {
		t.Errorf("maxVal = %v, want 100", w.cfg.maxVal)
	}
	if w.cfg.minVal != 0 {
		t.Errorf("minVal = %v, want 0", w.cfg.minVal)
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

func TestNew_WithOptions(t *testing.T) {
	changed := false
	w := New(
		Min(10),
		Max(200),
		Value(50),
		Step(5),
		OnChange(func(float32) { changed = true }),
		OrientationOpt(Vertical),
		Disabled(false),
		A11yHint("adjust brightness"),
	)

	if w.cfg.minVal != 10 {
		t.Errorf("minVal = %v, want 10", w.cfg.minVal)
	}
	if w.cfg.maxVal != 200 {
		t.Errorf("maxVal = %v, want 200", w.cfg.maxVal)
	}
	if w.cfg.ResolvedValue() != 50 {
		t.Errorf("value = %v, want 50", w.cfg.ResolvedValue())
	}
	if w.cfg.step != 5 {
		t.Errorf("step = %v, want 5", w.cfg.step)
	}
	if w.cfg.orientation != Vertical {
		t.Errorf("orientation = %v, want Vertical", w.cfg.orientation)
	}
	if w.cfg.a11yHint != "adjust brightness" {
		t.Errorf("a11yHint = %q, want %q", w.cfg.a11yHint, "adjust brightness")
	}

	w.cfg.onChange(42)
	if !changed {
		t.Error("onChange handler should have been called")
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

func TestIsFocusable_DisabledSignal(t *testing.T) {
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

// --- Layout Tests ---

func TestLayout_Horizontal(t *testing.T) {
	ctx := widget.NewContext()
	constraints := geometry.Loose(geometry.Sz(400, 400))

	w := New()
	size := w.Layout(ctx, constraints)

	if size.Width <= 0 {
		t.Error("width should be positive")
	}
	if size.Height <= 0 {
		t.Error("height should be positive")
	}
	if size.Width <= size.Height {
		t.Error("horizontal slider should be wider than tall")
	}
}

func TestLayout_Vertical(t *testing.T) {
	ctx := widget.NewContext()
	constraints := geometry.Loose(geometry.Sz(400, 400))

	w := New(OrientationOpt(Vertical))
	size := w.Layout(ctx, constraints)

	if size.Width <= 0 {
		t.Error("width should be positive")
	}
	if size.Height <= 0 {
		t.Error("height should be positive")
	}
	if size.Height <= size.Width {
		t.Error("vertical slider should be taller than wide")
	}
}

func TestLayout_Constrained(t *testing.T) {
	ctx := widget.NewContext()
	constraints := geometry.Tight(geometry.Sz(50, 30))

	w := New()
	size := w.Layout(ctx, constraints)

	if size.Width != 50 {
		t.Errorf("width = %v, want 50 (tight constraint)", size.Width)
	}
	if size.Height != 30 {
		t.Errorf("height = %v, want 30 (tight constraint)", size.Height)
	}
}

// --- Styling Tests ---

func TestStyling_Padding(t *testing.T) {
	w := New()
	result := w.Padding(16)

	if result != w {
		t.Error("fluent method should return the same widget for chaining")
	}
	if w.padding != 16 {
		t.Errorf("padding = %v, want 16", w.padding)
	}
}

// --- ClampAndSnap Tests ---

func TestClampAndSnap(t *testing.T) {
	tests := []struct {
		name   string
		val    float32
		minVal float32
		maxVal float32
		step   float32
		want   float32
	}{
		{"within range", 50, 0, 100, 0, 50},
		{"below min", -10, 0, 100, 0, 0},
		{"above max", 150, 0, 100, 0, 100},
		{"snap to step", 23, 0, 100, 10, 20},
		{"snap to step up", 27, 0, 100, 10, 30},
		{"snap exact step", 30, 0, 100, 10, 30},
		{"snap with offset min", 14, 5, 100, 10, 15},
		{"snap to max boundary", 97, 0, 100, 10, 100},
		{"zero step = continuous", 33.3, 0, 100, 0, 33.3},
		{"min equals max", 50, 50, 50, 0, 50},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			got := clampAndSnap(tc.val, tc.minVal, tc.maxVal, tc.step)
			if math.Abs(float64(got-tc.want)) > 0.001 {
				t.Errorf("clampAndSnap(%v, %v, %v, %v) = %v, want %v",
					tc.val, tc.minVal, tc.maxVal, tc.step, got, tc.want)
			}
		})
	}
}

// --- State Transition Tests ---

func TestStateTransition_HoverEnterLeave(t *testing.T) {
	w := New()
	w.SetBounds(geometry.NewRect(0, 0, 200, 30))
	ctx := widget.NewContext()

	enterEvt := event.NewMouseEvent(event.MouseEnter, event.ButtonNone, 0,
		geometry.Pt(100, 15), geometry.Pt(100, 15), event.ModNone)
	consumed := handleEvent(w, ctx, enterEvt)

	if !consumed {
		t.Error("MouseEnter should be consumed")
	}
	if w.interaction != stateHover {
		t.Errorf("interaction = %v, want stateHover", w.interaction)
	}

	leaveEvt := event.NewMouseEvent(event.MouseLeave, event.ButtonNone, 0,
		geometry.Pt(250, 15), geometry.Pt(250, 15), event.ModNone)
	consumed = handleEvent(w, ctx, leaveEvt)

	if !consumed {
		t.Error("MouseLeave should be consumed")
	}
	if w.interaction != stateNormal {
		t.Errorf("interaction = %v, want stateNormal", w.interaction)
	}
}

func TestStateTransition_DragSetValue(t *testing.T) {
	var lastValue float32
	w := New(
		Min(0), Max(100),
		OnChange(func(v float32) { lastValue = v }),
	)
	w.SetBounds(geometry.NewRect(0, 0, 200, 30))
	ctx := widget.NewContext()

	// Press at the middle of the track (accounting for thumb radius padding).
	pressEvt := event.NewMouseEvent(event.MousePress, event.ButtonLeft, event.ButtonStateLeft,
		geometry.Pt(100, 15), geometry.Pt(100, 15), event.ModNone)
	consumed := handleEvent(w, ctx, pressEvt)

	if !consumed {
		t.Error("MousePress should be consumed")
	}
	if w.interaction != stateDragging {
		t.Errorf("interaction = %v, want stateDragging", w.interaction)
	}
	if lastValue == 0 {
		t.Error("onChange should have been called with non-zero value")
	}

	// Release.
	releaseEvt := event.NewMouseEvent(event.MouseRelease, event.ButtonLeft, 0,
		geometry.Pt(100, 15), geometry.Pt(100, 15), event.ModNone)
	consumed = handleEvent(w, ctx, releaseEvt)

	if !consumed {
		t.Error("MouseRelease should be consumed (was dragging)")
	}
	if w.interaction != stateHover {
		t.Errorf("interaction = %v, want stateHover (released inside bounds)", w.interaction)
	}
}

func TestStateTransition_RightButton_Ignored(t *testing.T) {
	w := New()
	w.SetBounds(geometry.NewRect(0, 0, 200, 30))
	ctx := widget.NewContext()

	pressEvt := event.NewMouseEvent(event.MousePress, event.ButtonRight, event.ButtonStateRight,
		geometry.Pt(100, 15), geometry.Pt(100, 15), event.ModNone)
	consumed := handleEvent(w, ctx, pressEvt)

	if consumed {
		t.Error("right button press should not be consumed")
	}
	if w.interaction != stateNormal {
		t.Error("interaction should remain normal for right button")
	}
}

func TestStateTransition_DragMove(t *testing.T) {
	var values []float32
	w := New(
		Min(0), Max(100),
		OnChange(func(v float32) { values = append(values, v) }),
	)
	w.SetBounds(geometry.NewRect(0, 0, 200, 30))
	ctx := widget.NewContext()

	// Press to start dragging.
	press := event.NewMouseEvent(event.MousePress, event.ButtonLeft, event.ButtonStateLeft,
		geometry.Pt(50, 15), geometry.Pt(50, 15), event.ModNone)
	handleEvent(w, ctx, press)

	// Move during drag.
	move := event.NewMouseEvent(event.MouseMove, event.ButtonLeft, event.ButtonStateLeft,
		geometry.Pt(150, 15), geometry.Pt(150, 15), event.ModNone)
	consumed := handleEvent(w, ctx, move)

	if !consumed {
		t.Error("MouseMove during drag should be consumed")
	}
	if len(values) < 2 {
		t.Fatal("onChange should have been called at least twice (press + move)")
	}
	if values[len(values)-1] <= values[0] {
		t.Error("value should increase when dragging right")
	}
}

func TestStateTransition_MoveNotDragging_Ignored(t *testing.T) {
	w := New()
	w.SetBounds(geometry.NewRect(0, 0, 200, 30))
	ctx := widget.NewContext()

	move := event.NewMouseEvent(event.MouseMove, event.ButtonNone, 0,
		geometry.Pt(100, 15), geometry.Pt(100, 15), event.ModNone)
	consumed := handleEvent(w, ctx, move)

	if consumed {
		t.Error("MouseMove when not dragging should not be consumed")
	}
}

// --- Disabled State Tests ---

func TestDisabled_IgnoresEvents(t *testing.T) {
	changed := false
	w := New(OnChange(func(float32) { changed = true }), Disabled(true))
	w.SetBounds(geometry.NewRect(0, 0, 200, 30))
	ctx := widget.NewContext()

	enterEvt := event.NewMouseEvent(event.MouseEnter, event.ButtonNone, 0,
		geometry.Pt(100, 15), geometry.Pt(100, 15), event.ModNone)
	consumed := handleEvent(w, ctx, enterEvt)
	if consumed {
		t.Error("disabled slider should not consume MouseEnter")
	}

	pressEvt := event.NewMouseEvent(event.MousePress, event.ButtonLeft, event.ButtonStateLeft,
		geometry.Pt(100, 15), geometry.Pt(100, 15), event.ModNone)
	consumed = handleEvent(w, ctx, pressEvt)
	if consumed {
		t.Error("disabled slider should not consume MousePress")
	}

	if changed {
		t.Error("disabled slider should not fire onChange")
	}
}

func TestDisabled_IgnoresKeyboard(t *testing.T) {
	changed := false
	w := New(OnChange(func(float32) { changed = true }), Disabled(true))
	w.SetFocused(true)
	ctx := widget.NewContext()

	keyEvt := event.NewKeyEvent(event.KeyPress, event.KeyRight, 0, event.ModNone)
	consumed := handleEvent(w, ctx, keyEvt)

	if consumed {
		t.Error("disabled slider should not consume key events")
	}
	if changed {
		t.Error("disabled slider should not fire onChange")
	}
}

func TestDisabledFn_Reactive(t *testing.T) {
	isDisabled := false
	changed := false
	w := New(
		Min(0), Max(100), Value(50),
		OnChange(func(float32) { changed = true }),
		DisabledFn(func() bool { return isDisabled }),
	)
	w.SetBounds(geometry.NewRect(0, 0, 200, 30))
	ctx := widget.NewContext()

	// Press when not disabled.
	pressEvt := event.NewMouseEvent(event.MousePress, event.ButtonLeft, event.ButtonStateLeft,
		geometry.Pt(150, 15), geometry.Pt(150, 15), event.ModNone)
	handleEvent(w, ctx, pressEvt)

	if !changed {
		t.Error("should trigger onChange when not disabled")
	}

	changed = false
	isDisabled = true

	// Reset interaction state.
	w.interaction = stateNormal

	pressEvt2 := event.NewMouseEvent(event.MousePress, event.ButtonLeft, event.ButtonStateLeft,
		geometry.Pt(150, 15), geometry.Pt(150, 15), event.ModNone)
	consumed := handleEvent(w, ctx, pressEvt2)

	if consumed {
		t.Error("disabled slider should not consume events")
	}
	if changed {
		t.Error("disabled slider should not fire onChange")
	}
}

// --- Keyboard Navigation Tests ---

func TestKeyboard_RightIncrease(t *testing.T) {
	var lastValue float32
	w := New(Min(0), Max(100), Value(50), OnChange(func(v float32) { lastValue = v }))
	w.SetFocused(true)
	ctx := widget.NewContext()

	keyEvt := event.NewKeyEvent(event.KeyPress, event.KeyRight, 0, event.ModNone)
	consumed := handleEvent(w, ctx, keyEvt)

	if !consumed {
		t.Error("Right arrow should be consumed when focused")
	}
	if lastValue <= 50 {
		t.Errorf("value should increase, got %v", lastValue)
	}
}

func TestKeyboard_LeftDecrease(t *testing.T) {
	var lastValue float32
	w := New(Min(0), Max(100), Value(50), OnChange(func(v float32) { lastValue = v }))
	w.SetFocused(true)
	ctx := widget.NewContext()

	keyEvt := event.NewKeyEvent(event.KeyPress, event.KeyLeft, 0, event.ModNone)
	consumed := handleEvent(w, ctx, keyEvt)

	if !consumed {
		t.Error("Left arrow should be consumed when focused")
	}
	if lastValue >= 50 {
		t.Errorf("value should decrease, got %v", lastValue)
	}
}

func TestKeyboard_UpIncrease(t *testing.T) {
	var lastValue float32
	w := New(Min(0), Max(100), Value(50), OnChange(func(v float32) { lastValue = v }))
	w.SetFocused(true)
	ctx := widget.NewContext()

	keyEvt := event.NewKeyEvent(event.KeyPress, event.KeyUp, 0, event.ModNone)
	consumed := handleEvent(w, ctx, keyEvt)

	if !consumed {
		t.Error("Up arrow should be consumed when focused")
	}
	if lastValue <= 50 {
		t.Errorf("value should increase, got %v", lastValue)
	}
}

func TestKeyboard_DownDecrease(t *testing.T) {
	var lastValue float32
	w := New(Min(0), Max(100), Value(50), OnChange(func(v float32) { lastValue = v }))
	w.SetFocused(true)
	ctx := widget.NewContext()

	keyEvt := event.NewKeyEvent(event.KeyPress, event.KeyDown, 0, event.ModNone)
	consumed := handleEvent(w, ctx, keyEvt)

	if !consumed {
		t.Error("Down arrow should be consumed when focused")
	}
	if lastValue >= 50 {
		t.Errorf("value should decrease, got %v", lastValue)
	}
}

func TestKeyboard_Home(t *testing.T) {
	var lastValue float32
	w := New(Min(10), Max(100), Value(50), OnChange(func(v float32) { lastValue = v }))
	w.SetFocused(true)
	ctx := widget.NewContext()

	keyEvt := event.NewKeyEvent(event.KeyPress, event.KeyHome, 0, event.ModNone)
	consumed := handleEvent(w, ctx, keyEvt)

	if !consumed {
		t.Error("Home should be consumed when focused")
	}
	if lastValue != 10 {
		t.Errorf("Home should set to min (10), got %v", lastValue)
	}
}

func TestKeyboard_End(t *testing.T) {
	var lastValue float32
	w := New(Min(0), Max(100), Value(50), OnChange(func(v float32) { lastValue = v }))
	w.SetFocused(true)
	ctx := widget.NewContext()

	keyEvt := event.NewKeyEvent(event.KeyPress, event.KeyEnd, 0, event.ModNone)
	consumed := handleEvent(w, ctx, keyEvt)

	if !consumed {
		t.Error("End should be consumed when focused")
	}
	if lastValue != 100 {
		t.Errorf("End should set to max (100), got %v", lastValue)
	}
}

func TestKeyboard_PageUpDown(t *testing.T) {
	var lastValue float32
	w := New(Min(0), Max(100), Value(50), OnChange(func(v float32) { lastValue = v }))
	w.SetFocused(true)
	ctx := widget.NewContext()

	// PageUp.
	keyEvt := event.NewKeyEvent(event.KeyPress, event.KeyPageUp, 0, event.ModNone)
	handleEvent(w, ctx, keyEvt)
	if lastValue <= 50 {
		t.Errorf("PageUp should increase, got %v", lastValue)
	}
	if lastValue != 60 {
		t.Errorf("PageUp should increase by 10%% (10), got %v", lastValue)
	}

	// PageDown from 60.
	keyEvt = event.NewKeyEvent(event.KeyPress, event.KeyPageDown, 0, event.ModNone)
	handleEvent(w, ctx, keyEvt)
	if lastValue >= 60 {
		t.Errorf("PageDown should decrease from 60, got %v", lastValue)
	}
}

func TestKeyboard_WithStep(t *testing.T) {
	var lastValue float32
	w := New(Min(0), Max(100), Value(50), Step(10), OnChange(func(v float32) { lastValue = v }))
	w.SetFocused(true)
	ctx := widget.NewContext()

	keyEvt := event.NewKeyEvent(event.KeyPress, event.KeyRight, 0, event.ModNone)
	handleEvent(w, ctx, keyEvt)

	if lastValue != 60 {
		t.Errorf("with Step(10), Right should change to 60, got %v", lastValue)
	}
}

func TestKeyboard_NotFocused_Ignored(t *testing.T) {
	changed := false
	w := New(Min(0), Max(100), Value(50), OnChange(func(float32) { changed = true }))
	w.SetFocused(false)
	ctx := widget.NewContext()

	keyEvt := event.NewKeyEvent(event.KeyPress, event.KeyRight, 0, event.ModNone)
	consumed := handleEvent(w, ctx, keyEvt)

	if consumed {
		t.Error("keyboard should not be consumed when not focused")
	}
	if changed {
		t.Error("should not fire onChange when not focused")
	}
}

func TestKeyboard_OtherKeys_Ignored(t *testing.T) {
	w := New(Min(0), Max(100), Value(50))
	w.SetFocused(true)
	ctx := widget.NewContext()

	keyEvt := event.NewKeyEvent(event.KeyPress, event.KeyA, 'a', event.ModNone)
	consumed := handleEvent(w, ctx, keyEvt)

	if consumed {
		t.Error("key A should not be consumed by slider")
	}
}

func TestKeyboard_KeyRelease_Ignored(t *testing.T) {
	changed := false
	w := New(Min(0), Max(100), Value(50), OnChange(func(float32) { changed = true }))
	w.SetFocused(true)
	ctx := widget.NewContext()

	keyEvt := event.NewKeyEvent(event.KeyRelease, event.KeyRight, 0, event.ModNone)
	consumed := handleEvent(w, ctx, keyEvt)

	if consumed {
		t.Error("KeyRelease for arrows should not be consumed")
	}
	if changed {
		t.Error("should not fire onChange on key release")
	}
}

func TestKeyboard_ZeroRange_Ignored(t *testing.T) {
	w := New(Min(50), Max(50), Value(50))
	w.SetFocused(true)
	ctx := widget.NewContext()

	keyEvt := event.NewKeyEvent(event.KeyPress, event.KeyRight, 0, event.ModNone)
	consumed := handleEvent(w, ctx, keyEvt)

	if consumed {
		t.Error("zero range should not consume keyboard events")
	}
}

// --- Value Signal Two-Way Tests ---

func TestValueSignal_TwoWay(t *testing.T) {
	sig := state.NewSignal[float32](50)
	w := New(Min(0), Max(100), ValueSignal(sig))
	w.SetBounds(geometry.NewRect(0, 0, 200, 30))

	// Signal -> widget.
	if got := w.cfg.ResolvedValue(); got != 50 {
		t.Errorf("initial value should be 50 from signal, got %v", got)
	}

	sig.Set(75)
	if got := w.cfg.ResolvedValue(); got != 75 {
		t.Errorf("value should be 75 after signal.Set, got %v", got)
	}

	// Widget -> signal: setValue writes back.
	ctx := widget.NewContext()
	setValue(w, ctx, 30)
	if got := sig.Get(); got != 30 {
		t.Errorf("signal should be 30 after setValue, got %v", got)
	}
}

func TestValueSignal_TwoWay_MouseDrag(t *testing.T) {
	sig := state.NewSignal[float32](50)
	w := New(Min(0), Max(100), ValueSignal(sig))
	w.SetBounds(geometry.NewRect(0, 0, 200, 30))
	ctx := widget.NewContext()

	// Press near the right side of the track.
	press := event.NewMouseEvent(event.MousePress, event.ButtonLeft, event.ButtonStateLeft,
		geometry.Pt(180, 15), geometry.Pt(180, 15), event.ModNone)
	handleEvent(w, ctx, press)

	if sig.Get() <= 50 {
		t.Errorf("signal should have been updated to > 50, got %v", sig.Get())
	}
}

func TestValueSignal_DoesNotUpdateWithValueFn(t *testing.T) {
	externalValue := float32(50)
	w := New(Min(0), Max(100), ValueFn(func() float32 { return externalValue }))
	w.SetBounds(geometry.NewRect(0, 0, 200, 30))
	ctx := widget.NewContext()

	setValue(w, ctx, 75)

	// With ValueFn, the static field should NOT be updated.
	if w.cfg.value != 0 {
		t.Error("with ValueFn, setValue should not change static value field")
	}
}

// --- Draw Tests ---

func TestDraw_EmptyBounds(t *testing.T) {
	w := New()
	ctx := widget.NewContext()
	canvas := &internalMockCanvas{}

	w.Draw(ctx, canvas)

	if len(canvas.drawRoundRects) > 0 || len(canvas.drawCircles) > 0 {
		t.Error("should not draw anything with empty bounds")
	}
}

func TestDraw_Horizontal(t *testing.T) {
	w := New(Min(0), Max(100), Value(50))
	w.SetBounds(geometry.NewRect(0, 0, 200, 30))
	ctx := widget.NewContext()
	canvas := &internalMockCanvas{}

	w.Draw(ctx, canvas)

	if len(canvas.drawRoundRects) == 0 {
		t.Error("should draw track (round rects)")
	}
	if len(canvas.drawCircles) == 0 {
		t.Error("should draw thumb (circle)")
	}
}

func TestDraw_Vertical(t *testing.T) {
	w := New(Min(0), Max(100), Value(50), OrientationOpt(Vertical))
	w.SetBounds(geometry.NewRect(0, 0, 30, 200))
	ctx := widget.NewContext()
	canvas := &internalMockCanvas{}

	w.Draw(ctx, canvas)

	if len(canvas.drawRoundRects) == 0 {
		t.Error("should draw track (round rects)")
	}
	if len(canvas.drawCircles) == 0 {
		t.Error("should draw thumb (circle)")
	}
}

func TestDraw_FocusRing(t *testing.T) {
	w := New(Min(0), Max(100), Value(50))
	w.SetBounds(geometry.NewRect(0, 0, 200, 30))
	w.SetFocused(true)
	ctx := widget.NewContext()
	canvas := &internalMockCanvas{}

	w.Draw(ctx, canvas)

	// Focused slider draws 2 StrokeCircle calls: thumb border + focus ring.
	if len(canvas.strokeCircles) < 2 {
		t.Errorf("focused slider should draw thumb border + focus ring, got %d StrokeCircle calls", len(canvas.strokeCircles))
	}
}

func TestDraw_FocusRing_NotWhenDisabled(t *testing.T) {
	w := New(Min(0), Max(100), Value(50), Disabled(true))
	w.SetBounds(geometry.NewRect(0, 0, 200, 30))
	w.SetFocused(true)
	ctx := widget.NewContext()
	canvas := &internalMockCanvas{}

	w.Draw(ctx, canvas)

	// Disabled slider draws only 1 StrokeCircle (thumb border), no focus ring.
	if len(canvas.strokeCircles) > 1 {
		t.Error("disabled focused slider should not draw focus ring (only thumb border)")
	}
}

func TestDraw_WithMarks(t *testing.T) {
	w := New(
		Min(0), Max(100), Value(50),
		Marks([]Mark{{Value: 25}, {Value: 50}, {Value: 75}}),
	)
	w.SetBounds(geometry.NewRect(0, 0, 200, 30))
	ctx := widget.NewContext()
	canvas := &internalMockCanvas{}

	w.Draw(ctx, canvas)

	if len(canvas.drawLines) < 3 {
		t.Errorf("should draw at least 3 mark lines, got %d", len(canvas.drawLines))
	}
}

func TestDraw_ZeroRange(t *testing.T) {
	w := New(Min(50), Max(50), Value(50))
	w.SetBounds(geometry.NewRect(0, 0, 200, 30))
	ctx := widget.NewContext()
	canvas := &internalMockCanvas{}

	// Should not panic.
	w.Draw(ctx, canvas)
}

// --- Painter Interface Tests ---

func TestPainter_DelegationToPainter(t *testing.T) {
	p := &testPainter{}
	w := New(Min(0), Max(100), Value(50), PainterOpt(p))
	w.SetBounds(geometry.NewRect(0, 0, 200, 30))
	ctx := widget.NewContext()
	canvas := &internalMockCanvas{}

	w.Draw(ctx, canvas)

	if !p.called {
		t.Error("Draw should delegate to the configured painter")
	}
	if p.state.Value != 50 {
		t.Errorf("PaintState.Value = %v, want 50", p.state.Value)
	}
	if p.state.Bounds.IsEmpty() {
		t.Error("PaintState.Bounds should not be empty")
	}
	if p.state.Min != 0 || p.state.Max != 100 {
		t.Errorf("PaintState.Min/Max = (%v, %v), want (0, 100)", p.state.Min, p.state.Max)
	}
}

func TestPainter_PaintStateProgress(t *testing.T) {
	p := &testPainter{}
	w := New(Min(0), Max(100), Value(25), PainterOpt(p))
	w.SetBounds(geometry.NewRect(0, 0, 200, 30))
	ctx := widget.NewContext()
	canvas := &internalMockCanvas{}

	w.Draw(ctx, canvas)

	if math.Abs(float64(p.state.Progress-0.25)) > 0.001 {
		t.Errorf("PaintState.Progress = %v, want 0.25", p.state.Progress)
	}
}

func TestPainter_ColorScheme(t *testing.T) {
	w := New(Min(0), Max(100), Value(50))
	w.SetBounds(geometry.NewRect(0, 0, 200, 30))
	ctx := widget.NewContext()
	canvas := &internalMockCanvas{}

	// Draw with default painter using a color scheme.
	w.Draw(ctx, canvas)

	// Verify drawing occurred.
	if len(canvas.drawCircles) == 0 {
		t.Error("default painter should draw thumb circle")
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

// --- ValueFromPosition Tests ---

func TestValueFromPosition_Horizontal(t *testing.T) {
	w := New(Min(0), Max(100))
	w.SetBounds(geometry.NewRect(0, 0, 200, 30))

	// At the left edge of track.
	v := valueFromPosition(w, geometry.Pt(thumbRadius, 15))
	if math.Abs(float64(v)) > 1 {
		t.Errorf("left edge should be ~0, got %v", v)
	}

	// At the right edge of track.
	v = valueFromPosition(w, geometry.Pt(200-thumbRadius, 15))
	if math.Abs(float64(v-100)) > 1 {
		t.Errorf("right edge should be ~100, got %v", v)
	}

	// Left of track (clamped).
	v = valueFromPosition(w, geometry.Pt(-50, 15))
	if v != 0 {
		t.Errorf("left of track should clamp to 0, got %v", v)
	}

	// Right of track (clamped).
	v = valueFromPosition(w, geometry.Pt(500, 15))
	if v != 100 {
		t.Errorf("right of track should clamp to 100, got %v", v)
	}
}

func TestValueFromPosition_Vertical(t *testing.T) {
	w := New(Min(0), Max(100), OrientationOpt(Vertical))
	w.SetBounds(geometry.NewRect(0, 0, 30, 200))

	// At the bottom of track (value = 0).
	v := valueFromPosition(w, geometry.Pt(15, 200-thumbRadius))
	if math.Abs(float64(v)) > 1 {
		t.Errorf("bottom should be ~0, got %v", v)
	}

	// At the top of track (value = 100).
	v = valueFromPosition(w, geometry.Pt(15, thumbRadius))
	if math.Abs(float64(v-100)) > 1 {
		t.Errorf("top should be ~100, got %v", v)
	}
}

func TestValueFromPosition_ZeroRange(t *testing.T) {
	w := New(Min(50), Max(50))
	w.SetBounds(geometry.NewRect(0, 0, 200, 30))

	v := valueFromPosition(w, geometry.Pt(100, 15))
	if v != 50 {
		t.Errorf("zero range should return min (50), got %v", v)
	}
}

// --- SetValue Tests ---

func TestSetValue_ClampsToRange(t *testing.T) {
	var lastValue float32
	w := New(Min(0), Max(100), Value(50), OnChange(func(v float32) { lastValue = v }))
	ctx := widget.NewContext()

	setValue(w, ctx, 150)
	if lastValue != 100 {
		t.Errorf("value should be clamped to max (100), got %v", lastValue)
	}

	setValue(w, ctx, -50)
	if lastValue != 0 {
		t.Errorf("value should be clamped to min (0), got %v", lastValue)
	}
}

func TestSetValue_SnapsToStep(t *testing.T) {
	var lastValue float32
	w := New(Min(0), Max(100), Value(0), Step(10), OnChange(func(v float32) { lastValue = v }))
	ctx := widget.NewContext()

	setValue(w, ctx, 23)
	if lastValue != 20 {
		t.Errorf("value should snap to 20, got %v", lastValue)
	}

	setValue(w, ctx, 27)
	if lastValue != 30 {
		t.Errorf("value should snap to 30, got %v", lastValue)
	}
}

func TestSetValue_NilOnChange(t *testing.T) {
	w := New(Min(0), Max(100), Value(50))
	ctx := widget.NewContext()

	// Should not panic.
	setValue(w, ctx, 75)

	if w.cfg.value != 75 {
		t.Errorf("value should be updated to 75, got %v", w.cfg.value)
	}
}

func TestSetValue_SameValueNoCallback(t *testing.T) {
	callCount := 0
	w := New(Min(0), Max(100), Value(50), OnChange(func(float32) { callCount++ }))
	ctx := widget.NewContext()

	setValue(w, ctx, 50)
	if callCount != 0 {
		t.Error("onChange should not fire when value doesn't change")
	}
}

// --- Mount/Unmount Tests ---

func TestMount_BindsValueSignal(t *testing.T) {
	sig := state.NewSignal[float32](50)
	w := New(Min(0), Max(100), ValueSignal(sig))
	ctx := widget.NewContext()

	w.Mount(ctx)

	// Mount should not panic, bindings are tracked internally.
	// Verify the signal is still connected.
	sig.Set(75)
	if got := w.cfg.ResolvedValue(); got != 75 {
		t.Errorf("after Mount, signal should still work: got %v, want 75", got)
	}
}

func TestMount_BindsReadonlyValueSignal(t *testing.T) {
	base := state.NewSignal[float32](42)
	computed := state.NewComputed(func() float32 { return base.Get() * 2 })
	w := New(Min(0), Max(200), ValueReadonlySignal(computed))
	ctx := widget.NewContext()

	w.Mount(ctx)

	if got := w.cfg.ResolvedValue(); got != 84 {
		t.Errorf("after Mount with readonly signal: got %v, want 84", got)
	}
}

func TestMount_BindsDisabledSignal(t *testing.T) {
	sig := state.NewSignal(false)
	w := New(DisabledSignal(sig))
	ctx := widget.NewContext()

	w.Mount(ctx)

	if w.cfg.ResolvedDisabled() {
		t.Error("should not be disabled initially")
	}
	sig.Set(true)
	if !w.cfg.ResolvedDisabled() {
		t.Error("should be disabled after signal update")
	}
}

func TestMount_BindsReadonlyDisabledSignal(t *testing.T) {
	computed := state.NewComputed(func() bool { return true })
	w := New(DisabledReadonlySignal(computed))
	ctx := widget.NewContext()

	w.Mount(ctx)

	if !w.cfg.ResolvedDisabled() {
		t.Error("should be disabled from readonly signal")
	}
}

func TestMount_NilScheduler(t *testing.T) {
	sig := state.NewSignal[float32](50)
	w := New(ValueSignal(sig))
	// NewContext without scheduler returns nil scheduler.
	ctx := widget.NewContext()

	// Mount should not panic even without scheduler.
	w.Mount(ctx)
}

func TestUnmount(t *testing.T) {
	w := New()

	// Unmount should not panic.
	w.Unmount()
}

// --- Options Coverage for ReadonlySignal ---

func TestOptions_ValueReadonlySignal(t *testing.T) {
	computed := state.NewComputed(func() float32 { return 99 })
	var c config
	ValueReadonlySignal(computed)(&c)
	if c.readonlyValueSignal == nil {
		t.Error("readonlyValueSignal should be set")
	}
}

func TestOptions_DisabledReadonlySignal(t *testing.T) {
	computed := state.NewComputed(func() bool { return true })
	var c config
	DisabledReadonlySignal(computed)(&c)
	if c.readonlyDisabledSignal == nil {
		t.Error("readonlyDisabledSignal should be set")
	}
}

// --- Draw with ColorScheme ---

func TestDraw_WithColorScheme(t *testing.T) {
	scheme := SliderColorScheme{
		ActiveTrack:   widget.ColorBlue,
		InactiveTrack: widget.ColorGray,
		Thumb:         widget.ColorWhite,
		ThumbBorder:   widget.ColorBlack,
		FocusRing:     widget.ColorRed,
		DisabledTrack: widget.ColorLightGray,
		DisabledThumb: widget.ColorLightGray,
		MarkColor:     widget.ColorDarkGray,
	}

	ps := PaintState{
		Value:       50,
		Min:         0,
		Max:         100,
		Progress:    0.5,
		Bounds:      geometry.NewRect(0, 0, 200, 30),
		ColorScheme: scheme,
	}

	canvas := &internalMockCanvas{}
	dp := DefaultPainter{}
	dp.PaintSlider(canvas, ps)

	if len(canvas.drawRoundRects) == 0 {
		t.Error("should draw track with color scheme")
	}
}

func TestDraw_DisabledWithColorScheme(t *testing.T) {
	scheme := SliderColorScheme{
		ActiveTrack:   widget.ColorBlue,
		InactiveTrack: widget.ColorGray,
		Thumb:         widget.ColorWhite,
		ThumbBorder:   widget.ColorBlack,
		FocusRing:     widget.ColorRed,
		DisabledTrack: widget.ColorLightGray,
		DisabledThumb: widget.ColorLightGray,
		MarkColor:     widget.ColorDarkGray,
	}

	ps := PaintState{
		Value:       50,
		Min:         0,
		Max:         100,
		Progress:    0.5,
		Disabled:    true,
		Bounds:      geometry.NewRect(0, 0, 200, 30),
		ColorScheme: scheme,
	}

	canvas := &internalMockCanvas{}
	dp := DefaultPainter{}
	dp.PaintSlider(canvas, ps)

	if len(canvas.drawCircles) == 0 {
		t.Error("disabled slider should still draw thumb")
	}
}

func TestDraw_FocusedWithColorScheme(t *testing.T) {
	scheme := SliderColorScheme{
		ActiveTrack:   widget.ColorBlue,
		InactiveTrack: widget.ColorGray,
		Thumb:         widget.ColorWhite,
		ThumbBorder:   widget.ColorBlack,
		FocusRing:     widget.ColorRed,
		DisabledTrack: widget.ColorLightGray,
		DisabledThumb: widget.ColorLightGray,
		MarkColor:     widget.ColorDarkGray,
	}

	ps := PaintState{
		Value:       50,
		Min:         0,
		Max:         100,
		Progress:    0.5,
		Focused:     true,
		Bounds:      geometry.NewRect(0, 0, 200, 30),
		ColorScheme: scheme,
	}

	canvas := &internalMockCanvas{}
	dp := DefaultPainter{}
	dp.PaintSlider(canvas, ps)

	// Focus ring should use scheme color.
	if len(canvas.strokeCircles) < 2 {
		t.Error("focused slider with scheme should draw focus ring")
	}
}

func TestDraw_VerticalWithColorScheme(t *testing.T) {
	scheme := SliderColorScheme{
		ActiveTrack:   widget.ColorBlue,
		InactiveTrack: widget.ColorGray,
		Thumb:         widget.ColorWhite,
		ThumbBorder:   widget.ColorBlack,
		FocusRing:     widget.ColorRed,
		DisabledTrack: widget.ColorLightGray,
		DisabledThumb: widget.ColorLightGray,
		MarkColor:     widget.ColorDarkGray,
	}

	ps := PaintState{
		Value:       50,
		Min:         0,
		Max:         100,
		Progress:    0.5,
		Orientation: Vertical,
		Bounds:      geometry.NewRect(0, 0, 30, 200),
		ColorScheme: scheme,
	}

	canvas := &internalMockCanvas{}
	dp := DefaultPainter{}
	dp.PaintSlider(canvas, ps)

	if len(canvas.drawRoundRects) == 0 {
		t.Error("vertical slider with scheme should draw track")
	}
}

func TestDraw_MarksWithColorScheme(t *testing.T) {
	scheme := SliderColorScheme{
		ActiveTrack:   widget.ColorBlue,
		InactiveTrack: widget.ColorGray,
		Thumb:         widget.ColorWhite,
		ThumbBorder:   widget.ColorBlack,
		MarkColor:     widget.ColorDarkGray,
	}

	ps := PaintState{
		Value:       50,
		Min:         0,
		Max:         100,
		Progress:    0.5,
		Bounds:      geometry.NewRect(0, 0, 200, 30),
		Marks:       []Mark{{Value: 25}, {Value: 75}},
		ColorScheme: scheme,
	}

	canvas := &internalMockCanvas{}
	dp := DefaultPainter{}
	dp.PaintSlider(canvas, ps)

	if len(canvas.drawLines) < 2 {
		t.Errorf("should draw 2 marks with scheme, got %d", len(canvas.drawLines))
	}
}

func TestDraw_MarksOutOfRange_Skipped(t *testing.T) {
	ps := PaintState{
		Value:    50,
		Min:      0,
		Max:      100,
		Progress: 0.5,
		Bounds:   geometry.NewRect(0, 0, 200, 30),
		Marks:    []Mark{{Value: -10}, {Value: 150}, {Value: 50}},
	}

	canvas := &internalMockCanvas{}
	dp := DefaultPainter{}
	dp.PaintSlider(canvas, ps)

	// Only mark at 50 is in range.
	if len(canvas.drawLines) != 1 {
		t.Errorf("only 1 mark should be in range, got %d lines", len(canvas.drawLines))
	}
}

func TestDraw_MarksZeroRange_Skipped(t *testing.T) {
	ps := PaintState{
		Value:    50,
		Min:      50,
		Max:      50,
		Progress: 0,
		Bounds:   geometry.NewRect(0, 0, 200, 30),
		Marks:    []Mark{{Value: 50}},
	}

	canvas := &internalMockCanvas{}
	dp := DefaultPainter{}
	dp.PaintSlider(canvas, ps)

	// Zero range should skip marks.
	if len(canvas.drawLines) != 0 {
		t.Errorf("zero range should skip marks, got %d lines", len(canvas.drawLines))
	}
}

// --- Draw edge cases ---

func TestDraw_ProgressAtZero(t *testing.T) {
	ps := PaintState{
		Value:    0,
		Min:      0,
		Max:      100,
		Progress: 0,
		Bounds:   geometry.NewRect(0, 0, 200, 30),
	}

	canvas := &internalMockCanvas{}
	dp := DefaultPainter{}
	dp.PaintSlider(canvas, ps)

	// At progress 0, only inactive track + thumb should be drawn (no active track).
	// 1 DrawRoundRect for inactive track (active width = 0, so skipped).
	if len(canvas.drawRoundRects) < 1 {
		t.Error("should draw at least inactive track")
	}
}

func TestDraw_VerticalDisabled(t *testing.T) {
	ps := PaintState{
		Value:       25,
		Min:         0,
		Max:         100,
		Progress:    0.25,
		Disabled:    true,
		Orientation: Vertical,
		Bounds:      geometry.NewRect(0, 0, 30, 200),
	}

	canvas := &internalMockCanvas{}
	dp := DefaultPainter{}
	dp.PaintSlider(canvas, ps)

	if len(canvas.drawCircles) == 0 {
		t.Error("disabled vertical slider should still draw thumb")
	}
}

func TestDraw_TinyBounds_Horizontal(t *testing.T) {
	ps := PaintState{
		Value:    50,
		Min:      0,
		Max:      100,
		Progress: 0.5,
		Bounds:   geometry.NewRect(0, 0, 5, 30), // too small for thumb radius
	}

	canvas := &internalMockCanvas{}
	dp := DefaultPainter{}
	dp.PaintSlider(canvas, ps)

	// Track width <= 0 should cause early return (no drawing).
	if len(canvas.drawCircles) > 0 {
		t.Error("tiny bounds should not draw (track width <= 0)")
	}
}

func TestDraw_TinyBounds_Vertical(t *testing.T) {
	ps := PaintState{
		Value:       50,
		Min:         0,
		Max:         100,
		Progress:    0.5,
		Orientation: Vertical,
		Bounds:      geometry.NewRect(0, 0, 30, 5), // too small for thumb radius
	}

	canvas := &internalMockCanvas{}
	dp := DefaultPainter{}
	dp.PaintSlider(canvas, ps)

	if len(canvas.drawCircles) > 0 {
		t.Error("tiny vertical bounds should not draw")
	}
}

// --- Event edge cases ---

func TestEvent_MouseRelease_OutsideBounds(t *testing.T) {
	w := New(Min(0), Max(100), Value(50))
	w.SetBounds(geometry.NewRect(0, 0, 200, 30))
	ctx := widget.NewContext()

	// Start dragging.
	press := event.NewMouseEvent(event.MousePress, event.ButtonLeft, event.ButtonStateLeft,
		geometry.Pt(100, 15), geometry.Pt(100, 15), event.ModNone)
	handleEvent(w, ctx, press)

	// Release outside bounds.
	release := event.NewMouseEvent(event.MouseRelease, event.ButtonLeft, 0,
		geometry.Pt(300, 100), geometry.Pt(300, 100), event.ModNone)
	consumed := handleEvent(w, ctx, release)

	if !consumed {
		t.Error("release should be consumed (was dragging)")
	}
	if w.interaction != stateNormal {
		t.Errorf("interaction = %v, want stateNormal (released outside)", w.interaction)
	}
}

func TestEvent_MouseRelease_RightButton_Ignored(t *testing.T) {
	w := New(Min(0), Max(100), Value(50))
	w.SetBounds(geometry.NewRect(0, 0, 200, 30))
	ctx := widget.NewContext()

	release := event.NewMouseEvent(event.MouseRelease, event.ButtonRight, 0,
		geometry.Pt(100, 15), geometry.Pt(100, 15), event.ModNone)
	consumed := handleEvent(w, ctx, release)

	if consumed {
		t.Error("right button release should not be consumed")
	}
}

func TestEvent_MouseLeave_WhileDragging(t *testing.T) {
	w := New(Min(0), Max(100), Value(50))
	w.SetBounds(geometry.NewRect(0, 0, 200, 30))
	ctx := widget.NewContext()

	// Start dragging.
	press := event.NewMouseEvent(event.MousePress, event.ButtonLeft, event.ButtonStateLeft,
		geometry.Pt(100, 15), geometry.Pt(100, 15), event.ModNone)
	handleEvent(w, ctx, press)

	// Leave while dragging — should remain in dragging state.
	leave := event.NewMouseEvent(event.MouseLeave, event.ButtonNone, 0,
		geometry.Pt(250, 15), geometry.Pt(250, 15), event.ModNone)
	handleEvent(w, ctx, leave)

	if w.interaction != stateDragging {
		t.Errorf("interaction = %v, want stateDragging (leave while dragging)", w.interaction)
	}
}

func TestEvent_UnknownEvent_Ignored(t *testing.T) {
	w := New(Min(0), Max(100))
	ctx := widget.NewContext()

	// FocusEvent is not handled by slider.
	focusEvt := &event.FocusEvent{}
	consumed := handleEvent(w, ctx, focusEvt)

	if consumed {
		t.Error("focus event should not be consumed by slider")
	}
}

func TestEvent_UnknownMouseType_Ignored(t *testing.T) {
	w := New(Min(0), Max(100))
	w.SetBounds(geometry.NewRect(0, 0, 200, 30))
	ctx := widget.NewContext()

	// MouseDoubleClick is not handled by slider.
	dblClick := event.NewMouseEvent(event.MouseDoubleClick, event.ButtonLeft, event.ButtonStateLeft,
		geometry.Pt(100, 15), geometry.Pt(100, 15), event.ModNone)
	consumed := handleEvent(w, ctx, dblClick)

	if consumed {
		t.Error("double click should not be consumed (not handled)")
	}
}

// --- ValueFromPosition edge cases ---

func TestValueFromPosition_VerticalZeroTrack(t *testing.T) {
	w := New(Min(0), Max(100), OrientationOpt(Vertical))
	w.SetBounds(geometry.NewRect(0, 0, 30, 5)) // too small

	v := valueFromPosition(w, geometry.Pt(15, 2))
	if v != 0 {
		t.Errorf("zero track length should return min, got %v", v)
	}
}

func TestValueFromPosition_HorizontalZeroTrack(t *testing.T) {
	w := New(Min(0), Max(100))
	w.SetBounds(geometry.NewRect(0, 0, 5, 30)) // too small

	v := valueFromPosition(w, geometry.Pt(2, 15))
	if v != 0 {
		t.Errorf("zero track width should return min, got %v", v)
	}
}

// --- testPainter records the call to PaintSlider ---

type testPainter struct {
	called bool
	state  PaintState
}

func (p *testPainter) PaintSlider(_ widget.Canvas, ps PaintState) {
	p.called = true
	p.state = ps
}

// --- internalMockCanvas records canvas calls for testing ---

type internalMockCanvas struct {
	drawRoundRects   []internalDrawRoundRectCall
	strokeRoundRects []internalStrokeRoundRectCall
	drawCircles      []internalDrawCircleCall
	strokeCircles    []internalStrokeCircleCall
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

type internalDrawCircleCall struct {
	center geometry.Point
	radius float32
	color  widget.Color
}

type internalStrokeCircleCall struct {
	center      geometry.Point
	radius      float32
	color       widget.Color
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
func (c *internalMockCanvas) StrokeRect(_ geometry.Rect, _ widget.Color, _ float32) {}

func (c *internalMockCanvas) DrawRoundRect(r geometry.Rect, color widget.Color, radius float32) {
	c.drawRoundRects = append(c.drawRoundRects, internalDrawRoundRectCall{r: r, color: color, radius: radius})
}

func (c *internalMockCanvas) StrokeRoundRect(r geometry.Rect, color widget.Color, radius float32, strokeWidth float32) {
	c.strokeRoundRects = append(c.strokeRoundRects, internalStrokeRoundRectCall{r: r, color: color, radius: radius, strokeWidth: strokeWidth})
}

func (c *internalMockCanvas) DrawCircle(center geometry.Point, radius float32, color widget.Color) {
	c.drawCircles = append(c.drawCircles, internalDrawCircleCall{center: center, radius: radius, color: color})
}

func (c *internalMockCanvas) StrokeCircle(center geometry.Point, radius float32, color widget.Color, strokeWidth float32) {
	c.strokeCircles = append(c.strokeCircles, internalStrokeCircleCall{center: center, radius: radius, color: color, strokeWidth: strokeWidth})
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
