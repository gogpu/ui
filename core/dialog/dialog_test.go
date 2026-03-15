package dialog_test

import (
	"image"
	"testing"

	"github.com/gogpu/ui/core/dialog"
	"github.com/gogpu/ui/event"
	"github.com/gogpu/ui/geometry"
	"github.com/gogpu/ui/state"
	"github.com/gogpu/ui/widget"
)

// --- Construction Tests ---

func TestNew_Default(t *testing.T) {
	d := dialog.New()

	if d.IsOpen() {
		t.Error("dialog should not be open by default")
	}
	if d.Children() != nil {
		t.Error("dialog should have no children")
	}
}

func TestNew_WithTitle(t *testing.T) {
	d := dialog.New(dialog.Title("Confirm"))
	_ = d
}

func TestNew_WithActions(t *testing.T) {
	d := dialog.New(dialog.Actions(
		dialog.Action{Label: "Cancel", Variant: dialog.VariantTextOnly},
		dialog.Action{Label: "OK", Variant: dialog.VariantFilled, Default: true},
	))
	_ = d
}

func TestNew_WithContent(t *testing.T) {
	content := &mockContentWidget{}
	d := dialog.New(dialog.Content(content))
	_ = d
}

func TestNew_AllOptions(t *testing.T) {
	d := dialog.New(
		dialog.Title("Test"),
		dialog.Actions(dialog.Action{Label: "OK"}),
		dialog.DismissibleOpt(false),
		dialog.EscapeToCloseOpt(false),
		dialog.OnClose(func() {}),
		dialog.MaxWidth(400),
		dialog.MaxHeight(300),
		dialog.PainterOpt(dialog.DefaultPainter{}),
	)
	_ = d
}

// --- Show/Close Lifecycle Tests ---

func TestShow_PushesOverlay(t *testing.T) {
	d := dialog.New(dialog.Title("Test"))
	ctx, om := newTestContext()

	d.Show(ctx)

	if !d.IsOpen() {
		t.Error("dialog should be open after Show")
	}
	if om.pushCount != 1 {
		t.Errorf("PushOverlay called %d times, want 1", om.pushCount)
	}
}

func TestShow_NoOpIfAlreadyOpen(t *testing.T) {
	d := dialog.New(dialog.Title("Test"))
	ctx, om := newTestContext()

	d.Show(ctx)
	d.Show(ctx) // second call should be no-op

	if om.pushCount != 1 {
		t.Errorf("PushOverlay called %d times, want 1", om.pushCount)
	}
}

func TestShow_NoOpWithoutOverlayManager(t *testing.T) {
	d := dialog.New(dialog.Title("Test"))
	ctx := widget.NewContext() // no overlay manager

	d.Show(ctx)

	if d.IsOpen() {
		t.Error("dialog should not be open without overlay manager")
	}
}

func TestClose_RemovesOverlay(t *testing.T) {
	d := dialog.New(dialog.Title("Test"))
	ctx, om := newTestContext()

	d.Show(ctx)
	d.Close(ctx)

	if d.IsOpen() {
		t.Error("dialog should not be open after Close")
	}
	if om.removeCount != 1 {
		t.Errorf("RemoveOverlay called %d times, want 1", om.removeCount)
	}
}

func TestClose_NoOpIfNotOpen(t *testing.T) {
	d := dialog.New(dialog.Title("Test"))
	ctx, om := newTestContext()

	d.Close(ctx) // should not panic

	if om.removeCount != 0 {
		t.Errorf("RemoveOverlay should not be called when not open")
	}
}

func TestClose_CallsOnClose(t *testing.T) {
	closed := false
	d := dialog.New(
		dialog.Title("Test"),
		dialog.OnClose(func() { closed = true }),
	)
	ctx, _ := newTestContext()

	d.Show(ctx)
	d.Close(ctx)

	if !closed {
		t.Error("OnClose callback should have been called")
	}
}

// --- Layout Tests ---

func TestLayout_HiddenReturnsZero(t *testing.T) {
	d := dialog.New(dialog.Title("Test"))
	ctx := widget.NewContext()
	constraints := geometry.Loose(geometry.Sz(800, 600))

	size := d.Layout(ctx, constraints)

	if size.Width != 0 || size.Height != 0 {
		t.Errorf("hidden dialog should return zero size, got %v", size)
	}
}

func TestLayout_VisibleReturnsSomething(t *testing.T) {
	d := dialog.New(dialog.Title("Test"))
	ctx, _ := newTestContext()

	d.Show(ctx)
	constraints := geometry.Loose(geometry.Sz(800, 600))
	size := d.Layout(ctx, constraints)

	if size.Width <= 0 || size.Height <= 0 {
		t.Errorf("visible dialog should return positive size, got %v", size)
	}
}

// --- Draw Tests ---

func TestDraw_HiddenDoesNothing(t *testing.T) {
	d := dialog.New(dialog.Title("Test"))
	ctx := widget.NewContext()
	canvas := &recordingCanvas{}

	d.Draw(ctx, canvas) // should not panic or draw anything
}

// --- Event Tests ---

func TestEvent_HiddenReturnsFalse(t *testing.T) {
	d := dialog.New(dialog.Title("Test"))
	ctx := widget.NewContext()

	e := event.NewKeyEvent(event.KeyPress, event.KeyEscape, 0, event.ModNone)
	consumed := d.Event(ctx, e)

	if consumed {
		t.Error("hidden dialog should not consume events")
	}
}

// --- Signal Binding Tests ---

func TestTitleSignal(t *testing.T) {
	sig := state.NewSignal("Initial")
	_ = dialog.New(dialog.TitleSignal(sig))
}

func TestTitleReadonlySignal(t *testing.T) {
	base := state.NewSignal("base")
	computed := state.NewComputed(func() string {
		return "Title: " + base.Get()
	}, base)
	_ = dialog.New(dialog.TitleReadonlySignal(computed))
}

func TestTitleFn(t *testing.T) {
	_ = dialog.New(dialog.TitleFn(func() string { return "Dynamic Title" }))
}

// --- Convenience Constructors ---

func TestAlert(t *testing.T) {
	called := false
	d := dialog.Alert("Error", "Something failed", func() { called = true })

	if d == nil {
		t.Fatal("Alert should return a non-nil dialog")
	}
	if d.IsOpen() {
		t.Error("Alert dialog should not be open before Show")
	}
	_ = called
}

func TestConfirm(t *testing.T) {
	canceled := false
	confirmed := false
	d := dialog.Confirm("Delete?", "This cannot be undone",
		func() { canceled = true },
		func() { confirmed = true },
	)

	if d == nil {
		t.Fatal("Confirm should return a non-nil dialog")
	}
	if d.IsOpen() {
		t.Error("Confirm dialog should not be open before Show")
	}
	_ = canceled
	_ = confirmed
}

// --- Painter Delegation Tests ---

func TestDefaultPainter_EmptyBounds(t *testing.T) {
	p := dialog.DefaultPainter{}
	canvas := &recordingCanvas{}

	p.PaintDialog(canvas, dialog.PaintState{})

	if len(canvas.drawRoundRects) > 0 || len(canvas.drawTexts) > 0 {
		t.Error("should not draw anything with empty bounds")
	}
}

func TestDefaultPainter_WithBounds(t *testing.T) {
	p := dialog.DefaultPainter{}
	canvas := &recordingCanvas{}

	p.PaintDialog(canvas, dialog.PaintState{
		Title:  "Test Dialog",
		Bounds: geometry.NewRect(100, 100, 500, 400),
		Actions: []dialog.Action{
			{Label: "Cancel"},
			{Label: "OK"},
		},
	})

	if len(canvas.drawRoundRects) == 0 {
		t.Error("should draw dialog surface")
	}
	if len(canvas.drawTexts) == 0 {
		t.Error("should draw title and action text")
	}

	// Find title text.
	foundTitle := false
	for _, dt := range canvas.drawTexts {
		if dt.text == "Test Dialog" {
			foundTitle = true
			break
		}
	}
	if !foundTitle {
		t.Error("should draw the title text")
	}
}

func TestDefaultPainter_NoTitle(t *testing.T) {
	p := dialog.DefaultPainter{}
	canvas := &recordingCanvas{}

	p.PaintDialog(canvas, dialog.PaintState{
		Bounds: geometry.NewRect(100, 100, 500, 400),
	})

	for _, dt := range canvas.drawTexts {
		if dt.bold {
			t.Error("should not draw bold title text when title is empty")
		}
	}
}

func TestDefaultPainter_FocusRing(t *testing.T) {
	p := dialog.DefaultPainter{}
	canvas := &recordingCanvas{}

	p.PaintDialog(canvas, dialog.PaintState{
		Title:   "Focused",
		Bounds:  geometry.NewRect(100, 100, 500, 400),
		Focused: true,
	})

	foundFocusRing := false
	for _, call := range canvas.strokeRoundRects {
		if call.r.Min.X < 100 {
			foundFocusRing = true
			break
		}
	}
	if !foundFocusRing {
		t.Error("focused dialog should draw a focus ring")
	}
}

func TestDefaultPainter_CustomColorScheme(t *testing.T) {
	p := dialog.DefaultPainter{}
	canvas := &recordingCanvas{}

	customScheme := dialog.DialogColorScheme{
		Surface:  widget.ColorRed,
		Title:    widget.ColorBlue,
		Border:   widget.ColorGreen,
		ActionFg: widget.ColorYellow,
	}

	p.PaintDialog(canvas, dialog.PaintState{
		Title:       "Custom",
		Bounds:      geometry.NewRect(100, 100, 500, 400),
		ColorScheme: customScheme,
		Actions: []dialog.Action{
			{Label: "OK"},
		},
	})

	if len(canvas.drawRoundRects) == 0 {
		t.Fatal("should draw surface")
	}
	if canvas.drawRoundRects[0].color != widget.ColorRed {
		t.Errorf("surface color = %v, want Red", canvas.drawRoundRects[0].color)
	}
}

// --- Action Variant Tests ---

func TestActionVariants(t *testing.T) {
	tests := []struct {
		name    string
		variant uint8
	}{
		{"textOnly", dialog.VariantTextOnly},
		{"filled", dialog.VariantFilled},
		{"outlined", dialog.VariantOutlined},
		{"tonal", dialog.VariantTonal},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			_ = dialog.Action{Label: "Test", Variant: tc.variant}
		})
	}
}

// --- Mount/Unmount Tests ---

func TestMount_WithTitleSignal(t *testing.T) {
	sig := state.NewSignal("Hello")
	sched := &mockScheduler{}
	d := dialog.New(dialog.TitleSignal(sig))

	ctx := widget.NewContext()
	ctx.SetScheduler(sched)
	d.Mount(ctx)

	// Binding registration doesn't call MarkDirty, but it should not panic.
	_ = sched.markCount
}

func TestMount_WithReadonlyTitleSignal(t *testing.T) {
	base := state.NewSignal("base")
	computed := state.NewComputed(func() string { return base.Get() }, base)
	sched := &mockScheduler{}
	d := dialog.New(dialog.TitleReadonlySignal(computed))

	ctx := widget.NewContext()
	ctx.SetScheduler(sched)
	d.Mount(ctx)
}

func TestMount_NilScheduler(t *testing.T) {
	sig := state.NewSignal("Hello")
	d := dialog.New(dialog.TitleSignal(sig))

	ctx := widget.NewContext() // no scheduler
	d.Mount(ctx)               // should not panic
}

func TestUnmount(t *testing.T) {
	d := dialog.New(dialog.Title("Test"))
	d.Unmount() // should not panic
}

// --- Compile-time Interface Tests ---

func TestWidgetInterface(t *testing.T) {
	var _ widget.Widget = (*dialog.Widget)(nil)
}

func TestLifecycleInterface(t *testing.T) {
	var _ widget.Lifecycle = (*dialog.Widget)(nil)
}

// --- recordingCanvas records draw calls for verification ---

type recordingCanvas struct {
	drawRects        []drawRectCall
	drawRoundRects   []drawRoundRectCall
	strokeRoundRects []strokeRoundRectCall
	drawTexts        []drawTextCall
}

type drawRectCall struct {
	r     geometry.Rect
	color widget.Color
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

type drawTextCall struct {
	text     string
	bounds   geometry.Rect
	fontSize float32
	color    widget.Color
	bold     bool
	align    widget.TextAlign
}

func (c *recordingCanvas) Clear(_ widget.Color) {}

func (c *recordingCanvas) DrawRect(r geometry.Rect, color widget.Color) {
	c.drawRects = append(c.drawRects, drawRectCall{r: r, color: color})
}

func (c *recordingCanvas) StrokeRect(_ geometry.Rect, _ widget.Color, _ float32) {}

func (c *recordingCanvas) DrawRoundRect(r geometry.Rect, color widget.Color, radius float32) {
	c.drawRoundRects = append(c.drawRoundRects, drawRoundRectCall{r: r, color: color, radius: radius})
}

func (c *recordingCanvas) StrokeRoundRect(r geometry.Rect, color widget.Color, radius float32, strokeWidth float32) {
	c.strokeRoundRects = append(c.strokeRoundRects, strokeRoundRectCall{r: r, color: color, radius: radius, strokeWidth: strokeWidth})
}

func (c *recordingCanvas) DrawCircle(_ geometry.Point, _ float32, _ widget.Color)              {}
func (c *recordingCanvas) StrokeCircle(_ geometry.Point, _ float32, _ widget.Color, _ float32) {}
func (c *recordingCanvas) DrawLine(_, _ geometry.Point, _ widget.Color, _ float32)             {}

func (c *recordingCanvas) DrawText(text string, bounds geometry.Rect, fontSize float32, color widget.Color, bold bool, align widget.TextAlign) {
	c.drawTexts = append(c.drawTexts, drawTextCall{text: text, bounds: bounds, fontSize: fontSize, color: color, bold: bold, align: align})
}

func (c *recordingCanvas) MeasureText(text string, fontSize float32, _ bool) float32 {
	return float32(len([]rune(text))) * fontSize * 0.5
}

func (c *recordingCanvas) DrawImage(_ image.Image, _ geometry.Point)    {}
func (c *recordingCanvas) PushClip(_ geometry.Rect)                     {}
func (c *recordingCanvas) PushClipRoundRect(_ geometry.Rect, _ float32) {}
func (c *recordingCanvas) PopClip()                                     {}
func (c *recordingCanvas) PushTransform(_ geometry.Point)               {}
func (c *recordingCanvas) PopTransform()                                {}
func (c *recordingCanvas) TransformOffset() geometry.Point              { return geometry.Point{} }

// --- mockOverlayManager records overlay operations ---

type mockOverlayManager struct {
	pushCount   int
	popCount    int
	removeCount int
	lastWidget  widget.Widget
	onDismiss   func()
}

func (m *mockOverlayManager) PushOverlay(w widget.Widget, onDismiss func()) {
	m.pushCount++
	m.lastWidget = w
	m.onDismiss = onDismiss
}

func (m *mockOverlayManager) PopOverlay() {
	m.popCount++
}

func (m *mockOverlayManager) RemoveOverlay(_ widget.Widget) {
	m.removeCount++
}

// --- mockScheduler records scheduler calls ---

type mockScheduler struct {
	markCount int
}

func (s *mockScheduler) MarkDirty(_ widget.Widget) {
	s.markCount++
}

// --- mockContentWidget is a simple widget for content tests ---

type mockContentWidget struct {
	widget.WidgetBase
	drawn bool
}

func (w *mockContentWidget) Layout(_ widget.Context, c geometry.Constraints) geometry.Size {
	return c.Constrain(geometry.Sz(200, 100))
}

func (w *mockContentWidget) Draw(_ widget.Context, _ widget.Canvas) {
	w.drawn = true
}

func (w *mockContentWidget) Event(_ widget.Context, _ event.Event) bool {
	return false
}

func (w *mockContentWidget) Children() []widget.Widget {
	return nil
}

// --- newTestContext creates a context with a mock overlay manager ---

func newTestContext() (*widget.ContextImpl, *mockOverlayManager) {
	ctx := widget.NewContext()
	om := &mockOverlayManager{}
	ctx.SetOverlayManager(om)
	ctx.SetWindowSize(geometry.Sz(800, 600))
	return ctx, om
}
