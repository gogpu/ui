package checkbox_test

import (
	"testing"

	"github.com/gogpu/ui/core/checkbox"
	"github.com/gogpu/ui/event"
	"github.com/gogpu/ui/geometry"
	"github.com/gogpu/ui/state"
	"github.com/gogpu/ui/widget"
)

// --- Construction Tests ---

func TestNew_Defaults(t *testing.T) {
	cb := checkbox.New()

	if !cb.IsVisible() {
		t.Error("default checkbox should be visible")
	}
	if !cb.IsEnabled() {
		t.Error("default checkbox should be enabled")
	}
	if !cb.IsFocusable() {
		t.Error("default checkbox should be focusable")
	}
	if cb.Children() != nil {
		t.Error("checkbox should have no children")
	}
}

func TestNew_WithOptions(t *testing.T) {
	toggled := false
	cb := checkbox.New(
		checkbox.Label("Accept"),
		checkbox.Checked(true),
		checkbox.OnToggle(func(checked bool) { toggled = checked }),
		checkbox.Disabled(false),
		checkbox.A11yHint("toggle to accept"),
		checkbox.Background(widget.ColorBlue),
	)

	if !cb.IsFocusable() {
		t.Error("should be focusable")
	}
	_ = toggled
}

func TestNew_WithLabel(t *testing.T) {
	cb := checkbox.New(checkbox.Label("Terms"))
	cb.SetBounds(geometry.NewRect(0, 0, 200, 40))
	ctx := widget.NewContext()
	canvas := &recordingCanvas{}

	cb.Draw(ctx, canvas)

	if len(canvas.drawTexts) == 0 {
		t.Fatal("should have drawn label text")
	}
	if canvas.drawTexts[0].text != "Terms" {
		t.Errorf("label = %q, want %q", canvas.drawTexts[0].text, "Terms")
	}
}

func TestNew_WithLabelFn(t *testing.T) {
	counter := 0
	cb := checkbox.New(checkbox.LabelFn(func() string {
		counter++
		return "Dynamic"
	}))
	cb.SetBounds(geometry.NewRect(0, 0, 200, 40))
	ctx := widget.NewContext()
	canvas := &recordingCanvas{}

	cb.Draw(ctx, canvas)

	if len(canvas.drawTexts) == 0 {
		t.Fatal("should have drawn label text")
	}
	if canvas.drawTexts[0].text != "Dynamic" {
		t.Errorf("label = %q, want %q", canvas.drawTexts[0].text, "Dynamic")
	}
	if counter != 1 {
		t.Errorf("labelFn called %d times, want 1", counter)
	}
}

func TestNew_WithDisabled(t *testing.T) {
	cb := checkbox.New(checkbox.Disabled(true))

	if cb.IsFocusable() {
		t.Error("disabled checkbox should not be focusable")
	}
}

func TestNew_WithDisabledFn(t *testing.T) {
	isDisabled := true
	cb := checkbox.New(checkbox.DisabledFn(func() bool { return isDisabled }))

	if cb.IsFocusable() {
		t.Error("disabled checkbox should not be focusable")
	}

	isDisabled = false
	if !cb.IsFocusable() {
		t.Error("enabled checkbox should be focusable")
	}
}

func TestNew_AllOptions(t *testing.T) {
	cb := checkbox.New(
		checkbox.Label("All Options"),
		checkbox.Checked(true),
		checkbox.OnToggle(func(bool) {}),
		checkbox.Disabled(false),
		checkbox.Indeterminate(false),
		checkbox.A11yHint("accept terms"),
		checkbox.Background(widget.ColorBlue),
	)

	if !cb.IsFocusable() {
		t.Error("should be focusable")
	}
}

// --- Toggle Tests ---

func TestToggle_Click(t *testing.T) {
	var toggled bool
	var newState bool
	cb := checkbox.New(
		checkbox.Label("Click me"),
		checkbox.OnToggle(func(checked bool) {
			toggled = true
			newState = checked
		}),
	)
	cb.SetBounds(geometry.NewRect(0, 0, 100, 40))
	ctx := widget.NewContext()

	// Full click cycle.
	press := event.NewMouseEvent(event.MousePress, event.ButtonLeft, event.ButtonStateLeft,
		geometry.Pt(10, 20), geometry.Pt(10, 20), event.ModNone)
	cb.Event(ctx, press)

	release := event.NewMouseEvent(event.MouseRelease, event.ButtonLeft, 0,
		geometry.Pt(10, 20), geometry.Pt(10, 20), event.ModNone)
	cb.Event(ctx, release)

	if !toggled {
		t.Error("onToggle should have been called after click")
	}
	if !newState {
		t.Error("should toggle from unchecked to checked")
	}
}

func TestToggle_DoubleClick_TogglesBack(t *testing.T) {
	toggleCount := 0
	cb := checkbox.New(
		checkbox.OnToggle(func(bool) { toggleCount++ }),
	)
	cb.SetBounds(geometry.NewRect(0, 0, 100, 40))
	ctx := widget.NewContext()

	// First click.
	simulateClick(cb, ctx)
	if toggleCount != 1 {
		t.Fatalf("toggleCount = %d after first click, want 1", toggleCount)
	}

	// Second click.
	simulateClick(cb, ctx)
	if toggleCount != 2 {
		t.Fatalf("toggleCount = %d after second click, want 2", toggleCount)
	}
}

func TestToggle_Keyboard(t *testing.T) {
	var toggled bool
	cb := checkbox.New(
		checkbox.OnToggle(func(bool) { toggled = true }),
	)
	cb.SetFocused(true)
	ctx := widget.NewContext()

	press := event.NewKeyEvent(event.KeyPress, event.KeySpace, 0, event.ModNone)
	cb.Event(ctx, press)

	release := event.NewKeyEvent(event.KeyRelease, event.KeySpace, 0, event.ModNone)
	cb.Event(ctx, release)

	if !toggled {
		t.Error("Space key should toggle checkbox")
	}
}

func TestToggle_EnterKeyIgnored(t *testing.T) {
	toggled := false
	cb := checkbox.New(
		checkbox.OnToggle(func(bool) { toggled = true }),
	)
	cb.SetFocused(true)
	ctx := widget.NewContext()

	press := event.NewKeyEvent(event.KeyPress, event.KeyEnter, 0, event.ModNone)
	consumed := cb.Event(ctx, press)

	if consumed {
		t.Error("Enter key should not be consumed by checkbox")
	}
	if toggled {
		t.Error("Enter key should not toggle checkbox")
	}
}

func TestChecked_Dynamic(t *testing.T) {
	isChecked := true
	cb := checkbox.New(
		checkbox.CheckedFn(func() bool { return isChecked }),
	)
	cb.SetBounds(geometry.NewRect(0, 0, 100, 40))

	// Verify the draw uses checked state from CheckedFn.
	ctx := widget.NewContext()
	canvas := &recordingCanvas{}

	cb.Draw(ctx, canvas)
	// The draw should show checked state (filled box + checkmark = DrawRoundRect + DrawLine).
	hasRoundRect := len(canvas.drawRoundRects) > 0
	hasLine := len(canvas.drawLines) > 0
	if !hasRoundRect || !hasLine {
		t.Error("checked checkbox should draw filled box and checkmark lines")
	}

	isChecked = false
	canvas2 := &recordingCanvas{}
	cb.Draw(ctx, canvas2)
	// The draw should show unchecked state (border only = StrokeRoundRect).
	if len(canvas2.strokeRoundRects) == 0 {
		t.Error("unchecked checkbox should draw border")
	}
}

func TestDisabled_BlocksInteraction(t *testing.T) {
	toggled := false
	cb := checkbox.New(
		checkbox.OnToggle(func(bool) { toggled = true }),
		checkbox.Disabled(true),
	)
	cb.SetBounds(geometry.NewRect(0, 0, 100, 40))
	ctx := widget.NewContext()

	press := event.NewMouseEvent(event.MousePress, event.ButtonLeft, event.ButtonStateLeft,
		geometry.Pt(10, 20), geometry.Pt(10, 20), event.ModNone)
	consumed := cb.Event(ctx, press)

	if consumed {
		t.Error("disabled checkbox should not consume events")
	}
	if toggled {
		t.Error("disabled checkbox should not toggle")
	}
}

func TestIndeterminate_State(t *testing.T) {
	cb := checkbox.New(
		checkbox.Indeterminate(true),
	)
	cb.SetBounds(geometry.NewRect(0, 0, 100, 40))
	ctx := widget.NewContext()
	canvas := &recordingCanvas{}

	cb.Draw(ctx, canvas)

	// Indeterminate draws a filled box + single dash line.
	if len(canvas.drawRoundRects) == 0 {
		t.Error("indeterminate should draw filled box")
	}
	if len(canvas.drawLines) == 0 {
		t.Error("indeterminate should draw dash line")
	}
}

// --- Focus Tests ---

func TestFocusable_Interface(t *testing.T) {
	var f widget.Focusable = checkbox.New(checkbox.Label("Test"))
	_ = f
}

func TestFocus_SetFocused(t *testing.T) {
	cb := checkbox.New(checkbox.Label("Test"))

	cb.SetFocused(true)
	if !cb.IsFocused() {
		t.Error("should be focused after SetFocused(true)")
	}

	cb.SetFocused(false)
	if cb.IsFocused() {
		t.Error("should not be focused after SetFocused(false)")
	}
}

func TestFocusable_VisibleAndEnabled(t *testing.T) {
	cb := checkbox.New(checkbox.Label("Test"))

	if !cb.IsFocusable() {
		t.Error("visible+enabled checkbox should be focusable")
	}

	cb.SetVisible(false)
	if cb.IsFocusable() {
		t.Error("invisible checkbox should not be focusable")
	}

	cb.SetVisible(true)
	cb.SetEnabled(false)
	if cb.IsFocusable() {
		t.Error("disabled checkbox should not be focusable")
	}
}

// --- Layout Tests ---

func TestLayout_Size(t *testing.T) {
	cb := checkbox.New(checkbox.Label("Test"))
	ctx := widget.NewContext()
	constraints := geometry.Loose(geometry.Sz(400, 400))

	size := cb.Layout(ctx, constraints)

	if size.Width <= 0 {
		t.Error("width should be positive")
	}
	if size.Height < 24 {
		t.Errorf("height = %v, should be at least 24", size.Height)
	}
}

func TestLayout_NoLabel(t *testing.T) {
	cb := checkbox.New()
	ctx := widget.NewContext()
	constraints := geometry.Loose(geometry.Sz(400, 400))

	size := cb.Layout(ctx, constraints)

	if size.Width <= 0 {
		t.Error("width should be positive even without label")
	}
}

func TestLayout_TightConstraints(t *testing.T) {
	cb := checkbox.New(checkbox.Label("Test"))
	ctx := widget.NewContext()
	constraints := geometry.Tight(geometry.Sz(50, 30))

	size := cb.Layout(ctx, constraints)

	if size.Width != 50 {
		t.Errorf("width = %v, want 50", size.Width)
	}
	if size.Height != 30 {
		t.Errorf("height = %v, want 30", size.Height)
	}
}

// --- Draw Tests ---

func TestDraw_DelegatesToPainter(t *testing.T) {
	p := &testPainter{}
	cb := checkbox.New(checkbox.Label("Test"), checkbox.PainterOpt(p))
	cb.SetBounds(geometry.NewRect(0, 0, 100, 40))
	ctx := widget.NewContext()
	canvas := &mockCanvas{}

	cb.Draw(ctx, canvas)

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

func TestDraw_DoesNotPanicWithBounds(t *testing.T) {
	cb := checkbox.New(checkbox.Label("Click"))
	cb.SetBounds(geometry.NewRect(0, 0, 100, 40))
	ctx := widget.NewContext()
	canvas := &mockCanvas{}

	cb.Draw(ctx, canvas)
}

func TestPaintState_ColorScheme(t *testing.T) {
	scheme := checkbox.CheckboxColorScheme{
		CheckedBg:       widget.ColorRed,
		CheckedFg:       widget.ColorWhite,
		UncheckedBorder: widget.ColorGray,
		LabelColor:      widget.ColorBlack,
		DisabledBg:      widget.ColorLightGray,
		DisabledFg:      widget.ColorDarkGray,
		FocusRing:       widget.ColorBlue,
	}

	ps := checkbox.PaintState{
		Label:       "Test",
		Checked:     true,
		ColorScheme: scheme,
		Bounds:      geometry.NewRect(0, 0, 100, 40),
	}

	if ps.ColorScheme.CheckedBg != widget.ColorRed {
		t.Error("ColorScheme.CheckedBg should be red")
	}
	if ps.ColorScheme.CheckedFg != widget.ColorWhite {
		t.Error("ColorScheme.CheckedFg should be white")
	}
}

// --- Event Handling Tests ---

func TestEvent_MouseClickCycle(t *testing.T) {
	toggled := false
	cb := checkbox.New(
		checkbox.Label("Click"),
		checkbox.OnToggle(func(bool) { toggled = true }),
	)
	cb.SetBounds(geometry.NewRect(0, 0, 100, 40))
	ctx := widget.NewContext()

	enter := event.NewMouseEvent(event.MouseEnter, event.ButtonNone, 0,
		geometry.Pt(10, 20), geometry.Pt(10, 20), event.ModNone)
	cb.Event(ctx, enter)

	press := event.NewMouseEvent(event.MousePress, event.ButtonLeft, event.ButtonStateLeft,
		geometry.Pt(10, 20), geometry.Pt(10, 20), event.ModNone)
	cb.Event(ctx, press)

	release := event.NewMouseEvent(event.MouseRelease, event.ButtonLeft, 0,
		geometry.Pt(10, 20), geometry.Pt(10, 20), event.ModNone)
	cb.Event(ctx, release)

	if !toggled {
		t.Error("should have toggled after full mouse cycle")
	}
}

func TestEvent_DisabledIgnoresAll(t *testing.T) {
	toggled := false
	cb := checkbox.New(
		checkbox.OnToggle(func(bool) { toggled = true }),
		checkbox.Disabled(true),
	)
	cb.SetBounds(geometry.NewRect(0, 0, 100, 40))
	ctx := widget.NewContext()

	press := event.NewMouseEvent(event.MousePress, event.ButtonLeft, event.ButtonStateLeft,
		geometry.Pt(10, 20), geometry.Pt(10, 20), event.ModNone)
	consumed := cb.Event(ctx, press)

	if consumed {
		t.Error("disabled checkbox should not consume events")
	}
	if toggled {
		t.Error("disabled checkbox should not fire toggle")
	}
}

// --- Widget Interface Compliance ---

func TestWidgetInterface(t *testing.T) {
	var w widget.Widget = checkbox.New(checkbox.Label("Test"))
	_ = w
}

func TestFocusableInterface(t *testing.T) {
	var f widget.Focusable = checkbox.New(checkbox.Label("Test"))
	_ = f
}

// --- Fluent Styling Tests ---

func TestFluent_Chaining(t *testing.T) {
	cb := checkbox.New(checkbox.Label("Test"))
	result := cb.Padding(12).SetBackground(widget.ColorRed)

	if result != cb {
		t.Error("fluent methods should return the same widget")
	}
}

func TestFluent_Padding(t *testing.T) {
	cb := checkbox.New(checkbox.Label("Test"))
	cb.Padding(16)

	ctx := widget.NewContext()
	constraints := geometry.Loose(geometry.Sz(400, 400))
	size := cb.Layout(ctx, constraints)

	if size.Width <= 0 {
		t.Error("width should be positive with padding")
	}
}

// --- Helper functions ---

func simulateClick(cb *checkbox.Widget, ctx widget.Context) {
	press := event.NewMouseEvent(event.MousePress, event.ButtonLeft, event.ButtonStateLeft,
		geometry.Pt(10, 20), geometry.Pt(10, 20), event.ModNone)
	cb.Event(ctx, press)

	release := event.NewMouseEvent(event.MouseRelease, event.ButtonLeft, 0,
		geometry.Pt(10, 20), geometry.Pt(10, 20), event.ModNone)
	cb.Event(ctx, release)
}

// --- testPainter records the call to PaintCheckbox ---

type testPainter struct {
	called bool
	state  checkbox.PaintState
}

func (p *testPainter) PaintCheckbox(_ widget.Canvas, ps checkbox.PaintState) {
	p.called = true
	p.state = ps
}

// --- recordingCanvas records draw calls for verification ---

type recordingCanvas struct {
	drawTexts        []drawTextCall
	drawRoundRects   []drawRoundRectCall
	strokeRoundRects []strokeRoundRectCall
	drawLines        []drawLineCall
}

type drawTextCall struct {
	text     string
	bounds   geometry.Rect
	fontSize float32
	color    widget.Color
	bold     bool
	align    float32
}

type drawRoundRectCall struct {
	r      geometry.Rect
	color  widget.Color
	radius float32
}

type strokeRoundRectCall struct {
	r           geometry.Rect
	color       widget.Color
	radius      float32
	strokeWidth float32
}

type drawLineCall struct {
	from, to    geometry.Point
	color       widget.Color
	strokeWidth float32
}

func (c *recordingCanvas) Clear(_ widget.Color)                                  {}
func (c *recordingCanvas) DrawRect(_ geometry.Rect, _ widget.Color)              {}
func (c *recordingCanvas) StrokeRect(_ geometry.Rect, _ widget.Color, _ float32) {}

func (c *recordingCanvas) DrawRoundRect(r geometry.Rect, color widget.Color, radius float32) {
	c.drawRoundRects = append(c.drawRoundRects, drawRoundRectCall{r: r, color: color, radius: radius})
}

func (c *recordingCanvas) StrokeRoundRect(r geometry.Rect, color widget.Color, radius float32, strokeWidth float32) {
	c.strokeRoundRects = append(c.strokeRoundRects, strokeRoundRectCall{r: r, color: color, radius: radius, strokeWidth: strokeWidth})
}

func (c *recordingCanvas) DrawCircle(_ geometry.Point, _ float32, _ widget.Color)              {}
func (c *recordingCanvas) StrokeCircle(_ geometry.Point, _ float32, _ widget.Color, _ float32) {}

func (c *recordingCanvas) DrawLine(from, to geometry.Point, color widget.Color, strokeWidth float32) {
	c.drawLines = append(c.drawLines, drawLineCall{from: from, to: to, color: color, strokeWidth: strokeWidth})
}

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

// --- Signal Binding Tests ---

func TestNew_WithCheckedSignal(t *testing.T) {
	sig := state.NewSignal(true)
	cb := checkbox.New(checkbox.CheckedSignal(sig))
	cb.SetBounds(geometry.NewRect(0, 0, 100, 40))
	ctx := widget.NewContext()
	canvas := &recordingCanvas{}

	cb.Draw(ctx, canvas)

	// Should render as checked (filled box + checkmark).
	if len(canvas.drawRoundRects) == 0 {
		t.Error("checked (via signal) checkbox should draw filled box")
	}
	if len(canvas.drawLines) == 0 {
		t.Error("checked (via signal) checkbox should draw checkmark")
	}

	// Update signal to false and redraw.
	sig.Set(false)
	canvas2 := &recordingCanvas{}
	cb.Draw(ctx, canvas2)

	if len(canvas2.strokeRoundRects) == 0 {
		t.Error("unchecked (via signal) checkbox should draw border")
	}
}

func TestNew_WithCheckedSignal_TwoWay(t *testing.T) {
	sig := state.NewSignal(false)
	cb := checkbox.New(checkbox.CheckedSignal(sig))
	cb.SetBounds(geometry.NewRect(0, 0, 100, 40))
	ctx := widget.NewContext()

	// Click to toggle: false → true.
	simulateClick(cb, ctx)

	if !sig.Get() {
		t.Error("signal should be true after click toggle")
	}

	// Click again: true → false.
	simulateClick(cb, ctx)

	if sig.Get() {
		t.Error("signal should be false after second click toggle")
	}
}

func TestNew_WithLabelSignal(t *testing.T) {
	sig := state.NewSignal("Signal Label")
	cb := checkbox.New(checkbox.LabelSignal(sig))
	cb.SetBounds(geometry.NewRect(0, 0, 200, 40))
	ctx := widget.NewContext()
	canvas := &recordingCanvas{}

	cb.Draw(ctx, canvas)

	if len(canvas.drawTexts) == 0 {
		t.Fatal("should have drawn label text")
	}
	if canvas.drawTexts[0].text != "Signal Label" {
		t.Errorf("text = %q, want %q", canvas.drawTexts[0].text, "Signal Label")
	}

	sig.Set("Updated")
	canvas2 := &recordingCanvas{}
	cb.Draw(ctx, canvas2)

	if len(canvas2.drawTexts) == 0 {
		t.Fatal("should have drawn updated label text")
	}
	if canvas2.drawTexts[0].text != "Updated" {
		t.Errorf("text = %q, want %q", canvas2.drawTexts[0].text, "Updated")
	}
}

func TestNew_WithDisabledSignal(t *testing.T) {
	sig := state.NewSignal(true)
	cb := checkbox.New(checkbox.DisabledSignal(sig))

	if cb.IsFocusable() {
		t.Error("disabled (via signal) checkbox should not be focusable")
	}

	sig.Set(false)
	if !cb.IsFocusable() {
		t.Error("enabled (via signal) checkbox should be focusable")
	}
}

func TestNew_SignalPriority(t *testing.T) {
	t.Run("CheckedSignal overrides CheckedFn and Checked", func(t *testing.T) {
		sig := state.NewSignal(true)
		cb := checkbox.New(
			checkbox.Checked(false),
			checkbox.CheckedFn(func() bool { return false }),
			checkbox.CheckedSignal(sig),
		)
		cb.SetBounds(geometry.NewRect(0, 0, 100, 40))
		ctx := widget.NewContext()
		canvas := &recordingCanvas{}

		cb.Draw(ctx, canvas)

		// Should be checked (signal=true overrides fn=false and static=false).
		if len(canvas.drawRoundRects) == 0 {
			t.Error("CheckedSignal(true) should override CheckedFn(false) and Checked(false)")
		}
	})

	t.Run("LabelSignal overrides LabelFn and Label", func(t *testing.T) {
		sig := state.NewSignal("signal")
		cb := checkbox.New(
			checkbox.Label("static"),
			checkbox.LabelFn(func() string { return "fn" }),
			checkbox.LabelSignal(sig),
		)
		cb.SetBounds(geometry.NewRect(0, 0, 200, 40))
		ctx := widget.NewContext()
		canvas := &recordingCanvas{}

		cb.Draw(ctx, canvas)

		if len(canvas.drawTexts) == 0 {
			t.Fatal("should have drawn label text")
		}
		if canvas.drawTexts[0].text != "signal" {
			t.Errorf("text = %q, want %q (LabelSignal should override)", canvas.drawTexts[0].text, "signal")
		}
	})

	t.Run("DisabledSignal overrides DisabledFn and Disabled", func(t *testing.T) {
		sig := state.NewSignal(true)
		cb := checkbox.New(
			checkbox.Disabled(false),
			checkbox.DisabledFn(func() bool { return false }),
			checkbox.DisabledSignal(sig),
		)

		if cb.IsFocusable() {
			t.Error("DisabledSignal(true) should override DisabledFn(false) and Disabled(false)")
		}
	})
}

func TestNew_DisabledSignal_IgnoresEvents(t *testing.T) {
	toggled := false
	sig := state.NewSignal(true)
	cb := checkbox.New(
		checkbox.Label("Test"),
		checkbox.OnToggle(func(bool) { toggled = true }),
		checkbox.DisabledSignal(sig),
	)
	cb.SetBounds(geometry.NewRect(0, 0, 100, 40))
	ctx := widget.NewContext()

	press := event.NewMouseEvent(event.MousePress, event.ButtonLeft, event.ButtonStateLeft,
		geometry.Pt(10, 20), geometry.Pt(10, 20), event.ModNone)
	consumed := cb.Event(ctx, press)

	if consumed {
		t.Error("disabled (via signal) checkbox should not consume events")
	}
	if toggled {
		t.Error("disabled (via signal) checkbox should not toggle")
	}
}
