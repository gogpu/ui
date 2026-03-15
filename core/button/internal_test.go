package button

import (
	"image"
	"testing"

	"github.com/gogpu/ui/event"
	"github.com/gogpu/ui/geometry"
	"github.com/gogpu/ui/state"
	"github.com/gogpu/ui/widget"
)

// --- Variant Tests ---

func TestVariant_String(t *testing.T) {
	tests := []struct {
		name    string
		variant Variant
		want    string
	}{
		{"filled", Filled, "Filled"},
		{"outlined", Outlined, "Outlined"},
		{"textOnly", TextOnly, "TextOnly"},
		{"tonal", Tonal, "Tonal"},
		{"unknown", Variant(99), "Unknown"},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			if got := tc.variant.String(); got != tc.want {
				t.Errorf("Variant(%d).String() = %q, want %q", tc.variant, got, tc.want)
			}
		})
	}
}

func TestSize_String(t *testing.T) {
	tests := []struct {
		name string
		size Size
		want string
	}{
		{"small", Small, "Small"},
		{"medium", Medium, "Medium"},
		{"large", Large, "Large"},
		{"unknown", Size(99), "Unknown"},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			if got := tc.size.String(); got != tc.want {
				t.Errorf("Size(%d).String() = %q, want %q", tc.size, got, tc.want)
			}
		})
	}
}

func TestSizeHeight(t *testing.T) {
	tests := []struct {
		size Size
		want float32
	}{
		{Small, 32},
		{Medium, 40},
		{Large, 48},
		{Size(99), 40}, // default is Medium
	}
	for _, tc := range tests {
		if got := sizeHeight(tc.size); got != tc.want {
			t.Errorf("sizeHeight(%v) = %v, want %v", tc.size, got, tc.want)
		}
	}
}

func TestSizeFontSize(t *testing.T) {
	tests := []struct {
		size Size
		want float32
	}{
		{Small, 12},
		{Medium, 14},
		{Large, 16},
		{Size(99), 14}, // default is Medium
	}
	for _, tc := range tests {
		if got := sizeFontSize(tc.size); got != tc.want {
			t.Errorf("sizeFontSize(%v) = %v, want %v", tc.size, got, tc.want)
		}
	}
}

// --- Config Tests ---

func TestConfig_ResolvedText(t *testing.T) {
	t.Run("static text", func(t *testing.T) {
		c := config{text: "Hello"}
		if got := c.ResolvedText(); got != "Hello" {
			t.Errorf("resolvedText() = %q, want %q", got, "Hello")
		}
	})

	t.Run("dynamic text", func(t *testing.T) {
		c := config{text: "static", textFn: func() string { return "dynamic" }}
		if got := c.ResolvedText(); got != "dynamic" {
			t.Errorf("resolvedText() = %q, want %q", got, "dynamic")
		}
	})
}

func TestConfig_ResolvedDisabled(t *testing.T) {
	t.Run("static disabled", func(t *testing.T) {
		c := config{disabled: true}
		if !c.ResolvedDisabled() {
			t.Error("resolvedDisabled() should be true")
		}
	})

	t.Run("dynamic disabled", func(t *testing.T) {
		c := config{disabled: false, disabledFn: func() bool { return true }}
		if !c.ResolvedDisabled() {
			t.Error("resolvedDisabled() with fn should be true")
		}
	})

	t.Run("dynamic overrides static", func(t *testing.T) {
		c := config{disabled: true, disabledFn: func() bool { return false }}
		if c.ResolvedDisabled() {
			t.Error("resolvedDisabled() with fn returning false should be false")
		}
	})
}

// --- Options Tests ---

func TestOptions(t *testing.T) {
	t.Run("TextOpt", func(t *testing.T) {
		var c config
		TextOpt("Hello")(&c)
		if c.text != "Hello" {
			t.Errorf("text = %q, want %q", c.text, "Hello")
		}
	})

	t.Run("TextFn", func(t *testing.T) {
		var c config
		fn := func() string { return "dynamic" }
		TextFn(fn)(&c)
		if c.textFn == nil {
			t.Error("textFn should not be nil")
		}
		if c.textFn() != "dynamic" {
			t.Errorf("textFn() = %q, want %q", c.textFn(), "dynamic")
		}
	})

	t.Run("OnClick", func(t *testing.T) {
		var c config
		called := false
		OnClick(func() { called = true })(&c)
		if c.onClick == nil {
			t.Fatal("onClick should not be nil")
		}
		c.onClick()
		if !called {
			t.Error("onClick should have been called")
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

	t.Run("VariantOpt", func(t *testing.T) {
		var c config
		VariantOpt(Outlined)(&c)
		if c.variant != Outlined {
			t.Errorf("variant = %v, want %v", c.variant, Outlined)
		}
	})

	t.Run("SizeOpt", func(t *testing.T) {
		var c config
		SizeOpt(Large)(&c)
		if c.size != Large {
			t.Errorf("size = %v, want %v", c.size, Large)
		}
	})

	t.Run("A11yHint", func(t *testing.T) {
		var c config
		A11yHint("Submit form")(&c)
		if c.a11yHint != "Submit form" {
			t.Errorf("a11yHint = %q, want %q", c.a11yHint, "Submit form")
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

	t.Run("RoundedOpt", func(t *testing.T) {
		var c config
		var radius float32 = 12
		RoundedOpt(radius)(&c)
		if c.rounded == nil {
			t.Fatal("rounded should not be nil")
		}
		if *c.rounded != radius {
			t.Errorf("rounded = %v, want %v", *c.rounded, radius)
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
	if w.state != stateNormal {
		t.Errorf("state = %v, want %v", w.state, stateNormal)
	}
	if w.cfg.size != Medium {
		t.Errorf("size = %v, want %v", w.cfg.size, Medium)
	}
	if w.cfg.variant != Filled {
		t.Errorf("variant = %v, want %v", w.cfg.variant, Filled)
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
	clicked := false
	w := New(
		TextOpt("Submit"),
		OnClick(func() { clicked = true }),
		VariantOpt(Outlined),
		SizeOpt(Large),
		Disabled(false),
		A11yHint("click to submit"),
	)

	if w.cfg.ResolvedText() != "Submit" {
		t.Errorf("text = %q, want %q", w.cfg.ResolvedText(), "Submit")
	}
	if w.cfg.variant != Outlined {
		t.Errorf("variant = %v, want Outlined", w.cfg.variant)
	}
	if w.cfg.size != Large {
		t.Errorf("size = %v, want Large", w.cfg.size)
	}
	if w.cfg.a11yHint != "click to submit" {
		t.Errorf("a11yHint = %q, want %q", w.cfg.a11yHint, "click to submit")
	}

	w.cfg.onClick()
	if !clicked {
		t.Error("click handler should have been called")
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

func TestLayout_SizeVariants(t *testing.T) {
	ctx := widget.NewContext()
	constraints := geometry.Loose(geometry.Sz(400, 400))

	tests := []struct {
		name       string
		size       Size
		wantHeight float32
	}{
		{"small", Small, 32},
		{"medium", Medium, 40},
		{"large", Large, 48},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			w := New(TextOpt("Test"), SizeOpt(tc.size))
			size := w.Layout(ctx, constraints)

			if size.Height != tc.wantHeight {
				t.Errorf("height = %v, want %v", size.Height, tc.wantHeight)
			}
			if size.Width <= 0 {
				t.Error("width should be positive")
			}
		})
	}
}

func TestLayout_MinMaxWidth(t *testing.T) {
	ctx := widget.NewContext()
	constraints := geometry.Loose(geometry.Sz(400, 400))

	t.Run("min width", func(t *testing.T) {
		w := New(TextOpt("X"))
		w.MinWidth(200)
		size := w.Layout(ctx, constraints)

		if size.Width < 200 {
			t.Errorf("width = %v, should be at least 200", size.Width)
		}
	})

	t.Run("max width", func(t *testing.T) {
		w := New(TextOpt("This is a very long button label that exceeds max width"))
		w.MaxWidth(100)
		size := w.Layout(ctx, constraints)

		if size.Width > 100 {
			t.Errorf("width = %v, should be at most 100", size.Width)
		}
	})
}

func TestLayout_Constrained(t *testing.T) {
	ctx := widget.NewContext()
	constraints := geometry.Tight(geometry.Sz(80, 30))

	w := New(TextOpt("Test"))
	size := w.Layout(ctx, constraints)

	if size.Width != 80 {
		t.Errorf("width = %v, want 80 (tight constraint)", size.Width)
	}
	if size.Height != 30 {
		t.Errorf("height = %v, want 30 (tight constraint)", size.Height)
	}
}

// --- Styling Tests ---

func TestStyling_Chaining(t *testing.T) {
	w := New(TextOpt("Test"))
	result := w.Padding(20).PaddingXY(16, 8).SetBackground(widget.ColorRed).SetRounded(12).MinWidth(100).MaxWidth(300)

	if result != w {
		t.Error("fluent methods should return the same widget for chaining")
	}
	if w.paddingX != 16 {
		t.Errorf("paddingX = %v, want 16", w.paddingX)
	}
	if w.paddingY != 8 {
		t.Errorf("paddingY = %v, want 8", w.paddingY)
	}
	if w.cfg.background == nil || *w.cfg.background != widget.ColorRed {
		t.Error("background should be ColorRed")
	}
	if w.cfg.rounded == nil || *w.cfg.rounded != 12 {
		t.Error("rounded should be 12")
	}
	if w.minWidth != 100 {
		t.Errorf("minWidth = %v, want 100", w.minWidth)
	}
	if w.maxWidth != 300 {
		t.Errorf("maxWidth = %v, want 300", w.maxWidth)
	}
}

func TestPadding(t *testing.T) {
	w := New()
	w.Padding(24)
	if w.paddingX != 24 || w.paddingY != 24 {
		t.Errorf("Padding(24) = (%v, %v), want (24, 24)", w.paddingX, w.paddingY)
	}
}

func TestPaddingXY(t *testing.T) {
	w := New()
	w.PaddingXY(16, 8)
	if w.paddingX != 16 || w.paddingY != 8 {
		t.Errorf("PaddingXY(16, 8) = (%v, %v), want (16, 8)", w.paddingX, w.paddingY)
	}
}

// --- State Transition Tests ---

func TestStateTransition_HoverEnterLeave(t *testing.T) {
	w := New(TextOpt("Test"))
	w.SetBounds(geometry.NewRect(0, 0, 100, 40))
	ctx := widget.NewContext()

	enterEvt := event.NewMouseEvent(event.MouseEnter, event.ButtonNone, 0,
		geometry.Pt(50, 20), geometry.Pt(50, 20), event.ModNone)
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
	clicked := false
	w := New(TextOpt("Test"), OnClick(func() { clicked = true }))
	w.SetBounds(geometry.NewRect(0, 0, 100, 40))
	ctx := widget.NewContext()

	pressEvt := event.NewMouseEvent(event.MousePress, event.ButtonLeft, event.ButtonStateLeft,
		geometry.Pt(50, 20), geometry.Pt(50, 20), event.ModNone)
	consumed := handleEvent(w, ctx, pressEvt)

	if !consumed {
		t.Error("MousePress should be consumed")
	}
	if w.state != statePressed {
		t.Errorf("state = %v, want statePressed", w.state)
	}
	if clicked {
		t.Error("should not click on press, only on release")
	}

	releaseEvt := event.NewMouseEvent(event.MouseRelease, event.ButtonLeft, 0,
		geometry.Pt(50, 20), geometry.Pt(50, 20), event.ModNone)
	consumed = handleEvent(w, ctx, releaseEvt)

	if !consumed {
		t.Error("MouseRelease should be consumed")
	}
	if !clicked {
		t.Error("should have clicked on release inside bounds")
	}
	if w.state != stateHover {
		t.Errorf("state = %v, want stateHover (released inside bounds)", w.state)
	}
}

func TestStateTransition_PressReleaseOutside(t *testing.T) {
	clicked := false
	w := New(TextOpt("Test"), OnClick(func() { clicked = true }))
	w.SetBounds(geometry.NewRect(0, 0, 100, 40))
	ctx := widget.NewContext()

	pressEvt := event.NewMouseEvent(event.MousePress, event.ButtonLeft, event.ButtonStateLeft,
		geometry.Pt(50, 20), geometry.Pt(50, 20), event.ModNone)
	handleEvent(w, ctx, pressEvt)

	releaseEvt := event.NewMouseEvent(event.MouseRelease, event.ButtonLeft, 0,
		geometry.Pt(200, 200), geometry.Pt(200, 200), event.ModNone)
	handleEvent(w, ctx, releaseEvt)

	if clicked {
		t.Error("should not click when released outside bounds")
	}
	if w.state != stateNormal {
		t.Errorf("state = %v, want stateNormal (released outside)", w.state)
	}
}

func TestStateTransition_RightButton_Ignored(t *testing.T) {
	w := New(TextOpt("Test"))
	w.SetBounds(geometry.NewRect(0, 0, 100, 40))
	ctx := widget.NewContext()

	pressEvt := event.NewMouseEvent(event.MousePress, event.ButtonRight, event.ButtonStateRight,
		geometry.Pt(50, 20), geometry.Pt(50, 20), event.ModNone)
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
	clicked := false
	w := New(TextOpt("Test"), OnClick(func() { clicked = true }), Disabled(true))
	w.SetBounds(geometry.NewRect(0, 0, 100, 40))
	ctx := widget.NewContext()

	enterEvt := event.NewMouseEvent(event.MouseEnter, event.ButtonNone, 0,
		geometry.Pt(50, 20), geometry.Pt(50, 20), event.ModNone)
	consumed := handleEvent(w, ctx, enterEvt)
	if consumed {
		t.Error("disabled button should not consume MouseEnter")
	}

	pressEvt := event.NewMouseEvent(event.MousePress, event.ButtonLeft, event.ButtonStateLeft,
		geometry.Pt(50, 20), geometry.Pt(50, 20), event.ModNone)
	consumed = handleEvent(w, ctx, pressEvt)
	if consumed {
		t.Error("disabled button should not consume MousePress")
	}

	if clicked {
		t.Error("disabled button should not fire click")
	}
}

func TestDisabledFn_Reactive(t *testing.T) {
	isDisabled := false
	clicked := false
	w := New(
		TextOpt("Test"),
		OnClick(func() { clicked = true }),
		DisabledFn(func() bool { return isDisabled }),
	)
	w.SetBounds(geometry.NewRect(0, 0, 100, 40))
	ctx := widget.NewContext()

	pressEvt := event.NewMouseEvent(event.MousePress, event.ButtonLeft, event.ButtonStateLeft,
		geometry.Pt(50, 20), geometry.Pt(50, 20), event.ModNone)
	handleEvent(w, ctx, pressEvt)

	releaseEvt := event.NewMouseEvent(event.MouseRelease, event.ButtonLeft, 0,
		geometry.Pt(50, 20), geometry.Pt(50, 20), event.ModNone)
	handleEvent(w, ctx, releaseEvt)

	if !clicked {
		t.Error("should click when not disabled")
	}

	clicked = false
	isDisabled = true

	pressEvt2 := event.NewMouseEvent(event.MousePress, event.ButtonLeft, event.ButtonStateLeft,
		geometry.Pt(50, 20), geometry.Pt(50, 20), event.ModNone)
	consumed := handleEvent(w, ctx, pressEvt2)

	if consumed {
		t.Error("disabled button should not consume events")
	}
	if clicked {
		t.Error("disabled button should not fire click")
	}
}

// --- Keyboard Activation Tests ---

func TestKeyboard_EnterActivation(t *testing.T) {
	clicked := false
	w := New(TextOpt("Test"), OnClick(func() { clicked = true }))
	w.SetFocused(true)
	ctx := widget.NewContext()

	pressEvt := event.NewKeyEvent(event.KeyPress, event.KeyEnter, 0, event.ModNone)
	consumed := handleEvent(w, ctx, pressEvt)
	if !consumed {
		t.Error("Enter press should be consumed when focused")
	}
	if w.state != statePressed {
		t.Errorf("state = %v, want statePressed", w.state)
	}

	releaseEvt := event.NewKeyEvent(event.KeyRelease, event.KeyEnter, 0, event.ModNone)
	consumed = handleEvent(w, ctx, releaseEvt)
	if !consumed {
		t.Error("Enter release should be consumed when focused")
	}
	if !clicked {
		t.Error("Enter release should trigger click")
	}
	if w.state != stateNormal {
		t.Errorf("state = %v, want stateNormal after key release", w.state)
	}
}

func TestKeyboard_SpaceActivation(t *testing.T) {
	clicked := false
	w := New(TextOpt("Test"), OnClick(func() { clicked = true }))
	w.SetFocused(true)
	ctx := widget.NewContext()

	pressEvt := event.NewKeyEvent(event.KeyPress, event.KeySpace, 0, event.ModNone)
	handleEvent(w, ctx, pressEvt)

	releaseEvt := event.NewKeyEvent(event.KeyRelease, event.KeySpace, 0, event.ModNone)
	handleEvent(w, ctx, releaseEvt)

	if !clicked {
		t.Error("Space release should trigger click")
	}
}

func TestKeyboard_NotFocused_Ignored(t *testing.T) {
	clicked := false
	w := New(TextOpt("Test"), OnClick(func() { clicked = true }))
	w.SetFocused(false)
	ctx := widget.NewContext()

	pressEvt := event.NewKeyEvent(event.KeyPress, event.KeyEnter, 0, event.ModNone)
	consumed := handleEvent(w, ctx, pressEvt)

	if consumed {
		t.Error("Enter should not be consumed when not focused")
	}
	if clicked {
		t.Error("should not click when not focused")
	}
}

func TestKeyboard_OtherKeys_Ignored(t *testing.T) {
	w := New(TextOpt("Test"))
	w.SetFocused(true)
	ctx := widget.NewContext()

	pressEvt := event.NewKeyEvent(event.KeyPress, event.KeyA, 'a', event.ModNone)
	consumed := handleEvent(w, ctx, pressEvt)

	if consumed {
		t.Error("key A should not be consumed by button")
	}
}

func TestKeyboard_Disabled_Ignored(t *testing.T) {
	clicked := false
	w := New(TextOpt("Test"), OnClick(func() { clicked = true }), Disabled(true))
	w.SetFocused(true)
	ctx := widget.NewContext()

	pressEvt := event.NewKeyEvent(event.KeyPress, event.KeyEnter, 0, event.ModNone)
	consumed := handleEvent(w, ctx, pressEvt)

	if consumed {
		t.Error("disabled button should not consume key events")
	}
	if clicked {
		t.Error("disabled button should not fire click")
	}
}

// --- Draw Tests ---

func TestDraw_EmptyBounds(t *testing.T) {
	w := New(TextOpt("Test"))
	ctx := widget.NewContext()
	canvas := &internalMockCanvas{}

	w.Draw(ctx, canvas)

	if len(canvas.drawRoundRects) > 0 || len(canvas.strokeRoundRects) > 0 || len(canvas.drawTexts) > 0 {
		t.Error("should not draw anything with empty bounds")
	}
}

func TestDraw_FilledVariant(t *testing.T) {
	w := New(TextOpt("Click"), VariantOpt(Filled))
	w.SetBounds(geometry.NewRect(10, 10, 100, 40))
	ctx := widget.NewContext()
	canvas := &internalMockCanvas{}

	w.Draw(ctx, canvas)

	if len(canvas.drawRoundRects) == 0 {
		t.Error("Filled variant should draw a round rect background")
	}
	if len(canvas.drawTexts) == 0 {
		t.Error("should draw text")
	}
	if canvas.drawTexts[0].text != "Click" {
		t.Errorf("text = %q, want %q", canvas.drawTexts[0].text, "Click")
	}
}

func TestDraw_AllVariants_DrawBackground(t *testing.T) {
	// DefaultPainter always draws a round rect background for all variants.
	variants := []Variant{Filled, Outlined, TextOnly, Tonal}

	for _, v := range variants {
		t.Run(v.String(), func(t *testing.T) {
			w := New(TextOpt("Click"), VariantOpt(v))
			w.SetBounds(geometry.NewRect(10, 10, 100, 40))
			ctx := widget.NewContext()
			canvas := &internalMockCanvas{}

			w.Draw(ctx, canvas)

			if len(canvas.drawRoundRects) == 0 {
				t.Errorf("%v variant should draw a round rect with DefaultPainter", v)
			}
			if len(canvas.drawTexts) == 0 {
				t.Error("should draw text")
			}
		})
	}
}

func TestDraw_DefaultPainter_Colors(t *testing.T) {
	w := New(TextOpt("Click"))
	w.SetBounds(geometry.NewRect(0, 0, 100, 40))
	ctx := widget.NewContext()
	canvas := &internalMockCanvas{}

	w.Draw(ctx, canvas)

	if len(canvas.drawRoundRects) == 0 {
		t.Fatal("should draw background")
	}
	// DefaultPainter uses gray background.
	bg := canvas.drawRoundRects[0].color
	if bg != defaultBg {
		t.Errorf("background = %v, want defaultBg %v", bg, defaultBg)
	}

	if len(canvas.drawTexts) == 0 {
		t.Fatal("should draw text")
	}
	// DefaultPainter uses dark text.
	fg := canvas.drawTexts[0].color
	if fg != defaultFg {
		t.Errorf("foreground = %v, want defaultFg %v", fg, defaultFg)
	}
}

func TestDraw_DefaultPainter_Disabled(t *testing.T) {
	w := New(TextOpt("Click"), Disabled(true))
	w.SetBounds(geometry.NewRect(0, 0, 100, 40))
	ctx := widget.NewContext()
	canvas := &internalMockCanvas{}

	w.Draw(ctx, canvas)

	if len(canvas.drawRoundRects) == 0 {
		t.Fatal("should draw background even when disabled")
	}
	bg := canvas.drawRoundRects[0].color
	if bg == defaultBg {
		t.Error("disabled background should differ from normal background")
	}

	if len(canvas.drawTexts) == 0 {
		t.Fatal("should draw text")
	}
	fg := canvas.drawTexts[0].color
	if fg == defaultFg {
		t.Error("disabled foreground should differ from normal foreground")
	}
}

func TestDraw_FocusRing(t *testing.T) {
	w := New(TextOpt("Click"))
	w.SetBounds(geometry.NewRect(10, 10, 100, 40))
	w.SetFocused(true)
	ctx := widget.NewContext()
	canvas := &internalMockCanvas{}

	w.Draw(ctx, canvas)

	found := false
	for _, call := range canvas.strokeRoundRects {
		if call.r.Min.X < 10 {
			found = true
			break
		}
	}
	if !found {
		t.Error("focused button should draw a focus ring (expanded stroke)")
	}
}

func TestDraw_FocusRing_NotWhenDisabled(t *testing.T) {
	w := New(TextOpt("Click"), Disabled(true))
	w.SetBounds(geometry.NewRect(10, 10, 100, 40))
	w.SetFocused(true)
	ctx := widget.NewContext()
	canvas := &internalMockCanvas{}

	w.Draw(ctx, canvas)

	for _, call := range canvas.strokeRoundRects {
		if call.r.Min.X < 10 {
			t.Error("disabled focused button should not draw focus ring")
			break
		}
	}
}

func TestDraw_CustomRadius(t *testing.T) {
	w := New(TextOpt("Click"))
	w.SetRounded(16)
	w.SetBounds(geometry.NewRect(0, 0, 100, 40))
	ctx := widget.NewContext()
	canvas := &internalMockCanvas{}

	w.Draw(ctx, canvas)

	if len(canvas.drawRoundRects) == 0 {
		t.Fatal("should draw at least one round rect")
	}
	if canvas.drawRoundRects[0].radius != 16 {
		t.Errorf("radius = %v, want 16", canvas.drawRoundRects[0].radius)
	}
}

func TestDraw_CustomBackground(t *testing.T) {
	color := widget.ColorGreen
	w := New(TextOpt("Click"))
	w.SetBackground(color)
	w.SetBounds(geometry.NewRect(0, 0, 100, 40))
	ctx := widget.NewContext()
	canvas := &internalMockCanvas{}

	w.Draw(ctx, canvas)

	if len(canvas.drawRoundRects) == 0 {
		t.Fatal("should draw background")
	}
	if canvas.drawRoundRects[0].color != color {
		t.Errorf("background = %v, want %v", canvas.drawRoundRects[0].color, color)
	}
}

// --- Signal Binding Tests ---

func TestConfig_ResolvedText_Signal(t *testing.T) {
	sig := state.NewSignal("from signal")
	c := config{text: "static", textFn: func() string { return "from fn" }, textSignal: sig}

	if got := c.ResolvedText(); got != "from signal" {
		t.Errorf("ResolvedText() = %q, want %q", got, "from signal")
	}

	sig.Set("updated")
	if got := c.ResolvedText(); got != "updated" {
		t.Errorf("ResolvedText() = %q, want %q after Set", got, "updated")
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
		t.Error("ResolvedDisabled() should be false after signal Set(false)")
	}
}

func TestConfig_ResolvedText_SignalPriority(t *testing.T) {
	sig := state.NewSignal("signal")

	t.Run("signal beats fn and static", func(t *testing.T) {
		c := config{text: "static", textFn: func() string { return "fn" }, textSignal: sig}
		if got := c.ResolvedText(); got != "signal" {
			t.Errorf("ResolvedText() = %q, want %q (signal > fn > static)", got, "signal")
		}
	})

	t.Run("fn beats static when no signal", func(t *testing.T) {
		c := config{text: "static", textFn: func() string { return "fn" }}
		if got := c.ResolvedText(); got != "fn" {
			t.Errorf("ResolvedText() = %q, want %q (fn > static)", got, "fn")
		}
	})

	t.Run("static used when no signal or fn", func(t *testing.T) {
		c := config{text: "static"}
		if got := c.ResolvedText(); got != "static" {
			t.Errorf("ResolvedText() = %q, want %q", got, "static")
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
			t.Error("static(true) should be returned")
		}
	})
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

	// When both hovered and pressed, pressed should take precedence.
	result := applyStateModifier(base, true, true)
	pressedOnly := applyStateModifier(base, false, true)
	if result != pressedOnly {
		t.Error("pressed should take precedence over hovered")
	}
}

// --- No Click Handler Tests ---

func TestFireOnClick_NilHandler(t *testing.T) {
	w := New(TextOpt("Test"))
	fireOnClick(w)
}

// --- Painter Interface Tests ---

func TestPainter_DelegationToPainter(t *testing.T) {
	p := &testPainter{}
	w := New(TextOpt("Test"), PainterOpt(p))
	w.SetBounds(geometry.NewRect(0, 0, 100, 40))
	ctx := widget.NewContext()
	canvas := &internalMockCanvas{}

	w.Draw(ctx, canvas)

	if !p.called {
		t.Error("Draw should delegate to the configured painter")
	}
	if p.state.Text != "Test" {
		t.Errorf("PaintState.Text = %q, want %q", p.state.Text, "Test")
	}
	if p.state.Bounds.IsEmpty() {
		t.Error("PaintState.Bounds should not be empty")
	}
}

// testPainter records the call to PaintButton.
type testPainter struct {
	called bool
	state  PaintState
}

func (p *testPainter) PaintButton(_ widget.Canvas, ps PaintState) {
	p.called = true
	p.state = ps
}

// --- internalMockCanvas records canvas calls for testing ---

type internalMockCanvas struct {
	drawRoundRects   []internalDrawRoundRectCall
	strokeRoundRects []internalStrokeRoundRectCall
	drawTexts        []internalDrawTextCall
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

func (c *internalMockCanvas) Clear(_ widget.Color)                                  {}
func (c *internalMockCanvas) DrawRect(_ geometry.Rect, _ widget.Color)              {}
func (c *internalMockCanvas) StrokeRect(_ geometry.Rect, _ widget.Color, _ float32) {}

func (c *internalMockCanvas) DrawRoundRect(r geometry.Rect, color widget.Color, radius float32) {
	c.drawRoundRects = append(c.drawRoundRects, internalDrawRoundRectCall{r: r, color: color, radius: radius})
}

func (c *internalMockCanvas) StrokeRoundRect(r geometry.Rect, color widget.Color, radius float32, strokeWidth float32) {
	c.strokeRoundRects = append(c.strokeRoundRects, internalStrokeRoundRectCall{r: r, color: color, radius: radius, strokeWidth: strokeWidth})
}

func (c *internalMockCanvas) DrawCircle(_ geometry.Point, _ float32, _ widget.Color)              {}
func (c *internalMockCanvas) StrokeCircle(_ geometry.Point, _ float32, _ widget.Color, _ float32) {}
func (c *internalMockCanvas) DrawLine(_, _ geometry.Point, _ widget.Color, _ float32)             {}

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

// --- ReadonlySignal Tests ---

func TestConfig_ResolvedText_ReadonlySignal(t *testing.T) {
	base := state.NewSignal("base")
	computed := state.NewComputed(func() string {
		return "computed:" + base.Get()
	}, base)

	c := config{
		text:               "static",
		textFn:             func() string { return "fn" },
		textSignal:         state.NewSignal("signal"),
		readonlyTextSignal: computed,
	}

	if got := c.ResolvedText(); got != "computed:base" {
		t.Errorf("ResolvedText() = %q, want %q (readonlySignal > signal > fn > static)", got, "computed:base")
	}

	base.Set("updated")
	if got := c.ResolvedText(); got != "computed:updated" {
		t.Errorf("ResolvedText() = %q after dep change, want %q", got, "computed:updated")
	}
}

func TestConfig_ResolvedDisabled_ReadonlySignal(t *testing.T) {
	flag := state.NewSignal(true)
	computed := state.NewComputed(func() bool {
		return flag.Get()
	}, flag)

	c := config{
		disabled:               false,
		disabledFn:             func() bool { return false },
		disabledSignal:         state.NewSignal(false),
		readonlyDisabledSignal: computed,
	}

	if !c.ResolvedDisabled() {
		t.Error("ResolvedDisabled() should be true (readonly computed signal)")
	}

	flag.Set(false)
	if c.ResolvedDisabled() {
		t.Error("ResolvedDisabled() should be false after computed dependency change")
	}
}

func TestConfig_ResolvedText_ReadonlySignalPriority(t *testing.T) {
	computed := state.NewComputed(func() string { return "readonly" })

	t.Run("readonlySignal beats signal, fn, and static", func(t *testing.T) {
		c := config{
			text:               "static",
			textFn:             func() string { return "fn" },
			textSignal:         state.NewSignal("signal"),
			readonlyTextSignal: computed,
		}
		if got := c.ResolvedText(); got != "readonly" {
			t.Errorf("ResolvedText() = %q, want %q", got, "readonly")
		}
	})

	t.Run("signal used when no readonlySignal", func(t *testing.T) {
		sig := state.NewSignal("signal")
		c := config{text: "static", textFn: func() string { return "fn" }, textSignal: sig}
		if got := c.ResolvedText(); got != "signal" {
			t.Errorf("ResolvedText() = %q, want %q", got, "signal")
		}
	})
}
