package button_test

import (
	"testing"

	"github.com/gogpu/ui/core/button"
	"github.com/gogpu/ui/event"
	"github.com/gogpu/ui/geometry"
	"github.com/gogpu/ui/state"
	"github.com/gogpu/ui/widget"
)

// --- Construction Tests ---

func TestNew_Default(t *testing.T) {
	btn := button.New()

	if !btn.IsVisible() {
		t.Error("default button should be visible")
	}
	if !btn.IsEnabled() {
		t.Error("default button should be enabled")
	}
	if !btn.IsFocusable() {
		t.Error("default button should be focusable")
	}
	if btn.Children() != nil {
		t.Error("button should have no children")
	}
}

func TestNew_WithText(t *testing.T) {
	btn := button.New(button.Text("Submit"))
	btn.SetBounds(geometry.NewRect(0, 0, 200, 40))
	ctx := widget.NewContext()
	canvas := &recordingCanvas{}

	btn.Draw(ctx, canvas)

	if len(canvas.drawTexts) == 0 {
		t.Fatal("should have drawn text")
	}
	if canvas.drawTexts[0].text != "Submit" {
		t.Errorf("text = %q, want %q", canvas.drawTexts[0].text, "Submit")
	}
}

func TestNew_WithTextFn(t *testing.T) {
	counter := 0
	btn := button.New(button.TextFn(func() string {
		counter++
		return "Dynamic"
	}))
	btn.SetBounds(geometry.NewRect(0, 0, 200, 40))
	ctx := widget.NewContext()
	canvas := &recordingCanvas{}

	btn.Draw(ctx, canvas)

	if len(canvas.drawTexts) == 0 {
		t.Fatal("should have drawn text")
	}
	if canvas.drawTexts[0].text != "Dynamic" {
		t.Errorf("text = %q, want %q", canvas.drawTexts[0].text, "Dynamic")
	}
	if counter != 1 {
		t.Errorf("textFn called %d times, want 1", counter)
	}
}

func TestNew_WithOnClick(t *testing.T) {
	clicked := false
	btn := button.New(button.OnClick(func() {
		clicked = true
	}))
	btn.SetBounds(geometry.NewRect(0, 0, 100, 40))
	ctx := widget.NewContext()

	// Simulate full click cycle through public Event API.
	press := event.NewMouseEvent(event.MousePress, event.ButtonLeft, event.ButtonStateLeft,
		geometry.Pt(50, 20), geometry.Pt(50, 20), event.ModNone)
	btn.Event(ctx, press)

	release := event.NewMouseEvent(event.MouseRelease, event.ButtonLeft, 0,
		geometry.Pt(50, 20), geometry.Pt(50, 20), event.ModNone)
	btn.Event(ctx, release)

	if !clicked {
		t.Error("handler should have been called after click")
	}
}

func TestNew_WithVariant(t *testing.T) {
	tests := []struct {
		name    string
		variant button.Variant
	}{
		{"filled", button.Filled},
		{"outlined", button.Outlined},
		{"textOnly", button.TextOnly},
		{"tonal", button.Tonal},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			btn := button.New(button.VariantOpt(tc.variant))
			_ = btn
		})
	}
}

func TestNew_WithSize(t *testing.T) {
	tests := []struct {
		name string
		size button.Size
	}{
		{"small", button.Small},
		{"medium", button.Medium},
		{"large", button.Large},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			btn := button.New(button.SizeOpt(tc.size))
			_ = btn
		})
	}
}

func TestNew_WithDisabled(t *testing.T) {
	btn := button.New(button.Disabled(true))

	if btn.IsFocusable() {
		t.Error("disabled button should not be focusable")
	}
}

func TestNew_WithDisabledFn(t *testing.T) {
	isDisabled := true
	btn := button.New(button.DisabledFn(func() bool { return isDisabled }))

	if btn.IsFocusable() {
		t.Error("disabled button should not be focusable")
	}

	isDisabled = false
	if !btn.IsFocusable() {
		t.Error("enabled button should be focusable")
	}
}

func TestNew_WithA11yHint(t *testing.T) {
	btn := button.New(button.A11yHint("Submit the form"))
	_ = btn
}

func TestNew_WithBackground(t *testing.T) {
	btn := button.New(button.Background(widget.ColorRed))
	_ = btn
}

func TestNew_WithRounded(t *testing.T) {
	btn := button.New(button.Rounded(16))
	_ = btn
}

func TestNew_AllOptions(t *testing.T) {
	btn := button.New(
		button.Text("Submit"),
		button.OnClick(func() {}),
		button.VariantOpt(button.Filled),
		button.SizeOpt(button.Large),
		button.Disabled(false),
		button.A11yHint("submit"),
		button.Background(widget.ColorBlue),
		button.Rounded(12),
	)

	if !btn.IsFocusable() {
		t.Error("should be focusable")
	}
}

// --- Signal Binding Tests ---

func TestNew_WithTextSignal(t *testing.T) {
	sig := state.NewSignal("Signal Text")
	btn := button.New(button.TextSignal(sig))
	btn.SetBounds(geometry.NewRect(0, 0, 200, 40))
	ctx := widget.NewContext()
	canvas := &recordingCanvas{}

	btn.Draw(ctx, canvas)

	if len(canvas.drawTexts) == 0 {
		t.Fatal("should have drawn text")
	}
	if canvas.drawTexts[0].text != "Signal Text" {
		t.Errorf("text = %q, want %q", canvas.drawTexts[0].text, "Signal Text")
	}

	// Update signal and redraw.
	sig.Set("Updated")
	canvas = &recordingCanvas{}
	btn.Draw(ctx, canvas)

	if len(canvas.drawTexts) == 0 {
		t.Fatal("should have drawn text after signal update")
	}
	if canvas.drawTexts[0].text != "Updated" {
		t.Errorf("text = %q, want %q", canvas.drawTexts[0].text, "Updated")
	}
}

func TestNew_WithDisabledSignal(t *testing.T) {
	sig := state.NewSignal(true)
	btn := button.New(button.DisabledSignal(sig))

	if btn.IsFocusable() {
		t.Error("disabled button (via signal) should not be focusable")
	}

	sig.Set(false)
	if !btn.IsFocusable() {
		t.Error("enabled button (via signal) should be focusable")
	}
}

func TestNew_SignalPriority(t *testing.T) {
	t.Run("TextSignal overrides TextFn and TextOpt", func(t *testing.T) {
		sig := state.NewSignal("signal")
		btn := button.New(
			button.Text("static"),
			button.TextFn(func() string { return "fn" }),
			button.TextSignal(sig),
		)
		btn.SetBounds(geometry.NewRect(0, 0, 200, 40))
		ctx := widget.NewContext()
		canvas := &recordingCanvas{}

		btn.Draw(ctx, canvas)

		if len(canvas.drawTexts) == 0 {
			t.Fatal("should have drawn text")
		}
		if canvas.drawTexts[0].text != "signal" {
			t.Errorf("text = %q, want %q (signal should override fn and static)", canvas.drawTexts[0].text, "signal")
		}
	})

	t.Run("DisabledSignal overrides DisabledFn and Disabled", func(t *testing.T) {
		sig := state.NewSignal(true)
		btn := button.New(
			button.Disabled(false),
			button.DisabledFn(func() bool { return false }),
			button.DisabledSignal(sig),
		)

		if btn.IsFocusable() {
			t.Error("DisabledSignal(true) should override DisabledFn(false) and Disabled(false)")
		}
	})
}

func TestNew_DisabledSignal_IgnoresEvents(t *testing.T) {
	clicked := false
	sig := state.NewSignal(true)
	btn := button.New(
		button.Text("Click"),
		button.OnClick(func() { clicked = true }),
		button.DisabledSignal(sig),
	)
	btn.SetBounds(geometry.NewRect(0, 0, 100, 40))
	ctx := widget.NewContext()

	press := event.NewMouseEvent(event.MousePress, event.ButtonLeft, event.ButtonStateLeft,
		geometry.Pt(50, 20), geometry.Pt(50, 20), event.ModNone)
	consumed := btn.Event(ctx, press)

	if consumed {
		t.Error("disabled button (via signal) should not consume events")
	}
	if clicked {
		t.Error("disabled button (via signal) should not fire click")
	}
}

// --- Layout Tests ---

func TestLayout_Small(t *testing.T) {
	btn := button.New(button.Text("OK"), button.SizeOpt(button.Small))
	ctx := widget.NewContext()
	constraints := geometry.Loose(geometry.Sz(400, 400))

	size := btn.Layout(ctx, constraints)

	if size.Height != 32 {
		t.Errorf("Small height = %v, want 32", size.Height)
	}
}

func TestLayout_Medium(t *testing.T) {
	btn := button.New(button.Text("OK"), button.SizeOpt(button.Medium))
	ctx := widget.NewContext()
	constraints := geometry.Loose(geometry.Sz(400, 400))

	size := btn.Layout(ctx, constraints)

	if size.Height != 40 {
		t.Errorf("Medium height = %v, want 40", size.Height)
	}
}

func TestLayout_Large(t *testing.T) {
	btn := button.New(button.Text("OK"), button.SizeOpt(button.Large))
	ctx := widget.NewContext()
	constraints := geometry.Loose(geometry.Sz(400, 400))

	size := btn.Layout(ctx, constraints)

	if size.Height != 48 {
		t.Errorf("Large height = %v, want 48", size.Height)
	}
}

func TestLayout_TightConstraints(t *testing.T) {
	btn := button.New(button.Text("OK"))
	ctx := widget.NewContext()
	constraints := geometry.Tight(geometry.Sz(60, 30))

	size := btn.Layout(ctx, constraints)

	if size.Width != 60 {
		t.Errorf("width = %v, want 60", size.Width)
	}
	if size.Height != 30 {
		t.Errorf("height = %v, want 30", size.Height)
	}
}

// --- Fluent Styling Tests ---

func TestFluent_Chaining(t *testing.T) {
	btn := button.New(button.Text("Test"))
	result := btn.
		Padding(20).
		PaddingXY(16, 8).
		SetBackground(widget.ColorRed).
		SetRounded(12).
		MinWidth(100).
		MaxWidth(300)

	if result != btn {
		t.Error("fluent methods should return the same widget")
	}
}

func TestFluent_Padding(t *testing.T) {
	btn := button.New(button.Text("Test"))
	btn.Padding(24)

	ctx := widget.NewContext()
	constraints := geometry.Loose(geometry.Sz(400, 400))
	size := btn.Layout(ctx, constraints)

	if size.Width <= 0 {
		t.Error("width should be positive with padding")
	}
}

func TestFluent_MinWidth(t *testing.T) {
	btn := button.New(button.Text("X"))
	btn.MinWidth(200)

	ctx := widget.NewContext()
	constraints := geometry.Loose(geometry.Sz(400, 400))
	size := btn.Layout(ctx, constraints)

	if size.Width < 200 {
		t.Errorf("width = %v, should be >= 200 (minWidth)", size.Width)
	}
}

func TestFluent_MaxWidth(t *testing.T) {
	btn := button.New(button.Text("This is a very long button label text"))
	btn.MaxWidth(80)

	ctx := widget.NewContext()
	constraints := geometry.Loose(geometry.Sz(400, 400))
	size := btn.Layout(ctx, constraints)

	if size.Width > 80 {
		t.Errorf("width = %v, should be <= 80 (maxWidth)", size.Width)
	}
}

// --- Event Handling Tests ---

func TestEvent_MouseClickCycle(t *testing.T) {
	clicked := false
	btn := button.New(
		button.Text("Click"),
		button.OnClick(func() { clicked = true }),
	)
	btn.SetBounds(geometry.NewRect(0, 0, 100, 40))
	ctx := widget.NewContext()

	enter := event.NewMouseEvent(event.MouseEnter, event.ButtonNone, 0,
		geometry.Pt(50, 20), geometry.Pt(50, 20), event.ModNone)
	btn.Event(ctx, enter)

	press := event.NewMouseEvent(event.MousePress, event.ButtonLeft, event.ButtonStateLeft,
		geometry.Pt(50, 20), geometry.Pt(50, 20), event.ModNone)
	btn.Event(ctx, press)

	release := event.NewMouseEvent(event.MouseRelease, event.ButtonLeft, 0,
		geometry.Pt(50, 20), geometry.Pt(50, 20), event.ModNone)
	btn.Event(ctx, release)

	if !clicked {
		t.Error("should have clicked after full mouse cycle")
	}
}

func TestEvent_KeyboardActivation(t *testing.T) {
	clicked := false
	btn := button.New(
		button.Text("Submit"),
		button.OnClick(func() { clicked = true }),
	)
	btn.SetFocused(true)
	ctx := widget.NewContext()

	press := event.NewKeyEvent(event.KeyPress, event.KeyEnter, 0, event.ModNone)
	btn.Event(ctx, press)

	release := event.NewKeyEvent(event.KeyRelease, event.KeyEnter, 0, event.ModNone)
	btn.Event(ctx, release)

	if !clicked {
		t.Error("Enter key should activate button")
	}
}

func TestEvent_SpaceActivation(t *testing.T) {
	clicked := false
	btn := button.New(
		button.Text("Submit"),
		button.OnClick(func() { clicked = true }),
	)
	btn.SetFocused(true)
	ctx := widget.NewContext()

	press := event.NewKeyEvent(event.KeyPress, event.KeySpace, 0, event.ModNone)
	btn.Event(ctx, press)

	release := event.NewKeyEvent(event.KeyRelease, event.KeySpace, 0, event.ModNone)
	btn.Event(ctx, release)

	if !clicked {
		t.Error("Space key should activate button")
	}
}

func TestEvent_DisabledIgnoresAll(t *testing.T) {
	clicked := false
	btn := button.New(
		button.Text("Submit"),
		button.OnClick(func() { clicked = true }),
		button.Disabled(true),
	)
	btn.SetBounds(geometry.NewRect(0, 0, 100, 40))
	ctx := widget.NewContext()

	press := event.NewMouseEvent(event.MousePress, event.ButtonLeft, event.ButtonStateLeft,
		geometry.Pt(50, 20), geometry.Pt(50, 20), event.ModNone)
	consumed := btn.Event(ctx, press)

	if consumed {
		t.Error("disabled button should not consume events")
	}
	if clicked {
		t.Error("disabled button should not fire click")
	}
}

// --- Draw Tests ---

func TestDraw_DoesNotPanicWithBounds(t *testing.T) {
	btn := button.New(button.Text("Click"))
	btn.SetBounds(geometry.NewRect(0, 0, 100, 40))
	ctx := widget.NewContext()
	canvas := &mockCanvas{}

	btn.Draw(ctx, canvas)
}

func TestDraw_AllVariants(t *testing.T) {
	variants := []button.Variant{button.Filled, button.Outlined, button.TextOnly, button.Tonal}

	for _, v := range variants {
		t.Run(v.String(), func(t *testing.T) {
			btn := button.New(button.Text("Test"), button.VariantOpt(v))
			btn.SetBounds(geometry.NewRect(0, 0, 100, 40))
			ctx := widget.NewContext()
			canvas := &mockCanvas{}

			btn.Draw(ctx, canvas)
		})
	}
}

// --- Focus Tests ---

func TestFocus_SetFocused(t *testing.T) {
	btn := button.New(button.Text("Test"))

	btn.SetFocused(true)
	if !btn.IsFocused() {
		t.Error("should be focused after SetFocused(true)")
	}

	btn.SetFocused(false)
	if btn.IsFocused() {
		t.Error("should not be focused after SetFocused(false)")
	}
}

func TestFocusable_VisibleAndEnabled(t *testing.T) {
	btn := button.New(button.Text("Test"))

	if !btn.IsFocusable() {
		t.Error("visible+enabled button should be focusable")
	}

	btn.SetVisible(false)
	if btn.IsFocusable() {
		t.Error("invisible button should not be focusable")
	}

	btn.SetVisible(true)
	btn.SetEnabled(false)
	if btn.IsFocusable() {
		t.Error("disabled button should not be focusable")
	}
}

// --- Widget Interface Compliance ---

func TestWidgetInterface(t *testing.T) {
	var w widget.Widget = button.New(button.Text("Test"))
	_ = w
}

func TestFocusableInterface(t *testing.T) {
	var f widget.Focusable = button.New(button.Text("Test"))
	_ = f
}

// --- recordingCanvas records draw calls for verification ---

type recordingCanvas struct {
	drawTexts []drawTextCall
}

type drawTextCall struct {
	text     string
	bounds   geometry.Rect
	fontSize float32
	color    widget.Color
	bold     bool
	align    float32
}

func (c *recordingCanvas) Clear(_ widget.Color)                                  {}
func (c *recordingCanvas) DrawRect(_ geometry.Rect, _ widget.Color)              {}
func (c *recordingCanvas) StrokeRect(_ geometry.Rect, _ widget.Color, _ float32) {}
func (c *recordingCanvas) DrawRoundRect(_ geometry.Rect, _ widget.Color, _ float32) {
}
func (c *recordingCanvas) StrokeRoundRect(_ geometry.Rect, _ widget.Color, _ float32, _ float32) {
}
func (c *recordingCanvas) DrawCircle(_ geometry.Point, _ float32, _ widget.Color)              {}
func (c *recordingCanvas) StrokeCircle(_ geometry.Point, _ float32, _ widget.Color, _ float32) {}
func (c *recordingCanvas) DrawLine(_, _ geometry.Point, _ widget.Color, _ float32)             {}

func (c *recordingCanvas) DrawText(text string, bounds geometry.Rect, fontSize float32, color widget.Color, bold bool, align float32) {
	c.drawTexts = append(c.drawTexts, drawTextCall{text: text, bounds: bounds, fontSize: fontSize, color: color, bold: bold, align: align})
}

func (c *recordingCanvas) PushClip(_ geometry.Rect)       {}
func (c *recordingCanvas) PopClip()                       {}
func (c *recordingCanvas) PushTransform(_ geometry.Point) {}
func (c *recordingCanvas) PopTransform()                  {}

// --- mockCanvas for non-recording tests ---

type mockCanvas struct{}

func (c *mockCanvas) Clear(_ widget.Color)                                                  {}
func (c *mockCanvas) DrawRect(_ geometry.Rect, _ widget.Color)                              {}
func (c *mockCanvas) StrokeRect(_ geometry.Rect, _ widget.Color, _ float32)                 {}
func (c *mockCanvas) DrawRoundRect(_ geometry.Rect, _ widget.Color, _ float32)              {}
func (c *mockCanvas) StrokeRoundRect(_ geometry.Rect, _ widget.Color, _ float32, _ float32) {}
func (c *mockCanvas) DrawCircle(_ geometry.Point, _ float32, _ widget.Color)                {}
func (c *mockCanvas) StrokeCircle(_ geometry.Point, _ float32, _ widget.Color, _ float32)   {}
func (c *mockCanvas) DrawLine(_, _ geometry.Point, _ widget.Color, _ float32)               {}

func (c *mockCanvas) DrawText(_ string, _ geometry.Rect, _ float32, _ widget.Color, _ bool, _ float32) {
}

func (c *mockCanvas) PushClip(_ geometry.Rect)       {}
func (c *mockCanvas) PopClip()                       {}
func (c *mockCanvas) PushTransform(_ geometry.Point) {}
func (c *mockCanvas) PopTransform()                  {}
