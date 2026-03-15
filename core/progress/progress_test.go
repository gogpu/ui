package progress_test

import (
	"image"
	"testing"
	"time"

	"github.com/gogpu/ui/core/progress"
	"github.com/gogpu/ui/event"
	"github.com/gogpu/ui/geometry"
	"github.com/gogpu/ui/state"
	"github.com/gogpu/ui/widget"
)

// --- Construction Tests ---

func TestNew_Defaults(t *testing.T) {
	w := progress.New()

	if !w.IsVisible() {
		t.Error("default indicator should be visible")
	}
	if !w.IsEnabled() {
		t.Error("default indicator should be enabled")
	}
	if w.Children() != nil {
		t.Error("indicator should have no children")
	}
	if w.Value() != 0 {
		t.Errorf("default value should be 0, got %v", w.Value())
	}
	if w.IsIndeterminate() {
		t.Error("default should be determinate mode")
	}
}

func TestNew_WithValue(t *testing.T) {
	w := progress.New(progress.Value(0.65))

	if w.Value() != 0.65 {
		t.Errorf("value = %v, want 0.65", w.Value())
	}
}

func TestNew_WithOptions(t *testing.T) {
	w := progress.New(
		progress.Value(0.5),
		progress.Size(64),
		progress.StrokeWidth(6),
		progress.ShowLabel(true),
		progress.Disabled(false),
	)

	if w.Value() != 0.5 {
		t.Errorf("value = %v, want 0.5", w.Value())
	}
	if !w.IsEnabled() {
		t.Error("should be enabled")
	}
}

func TestNew_Indeterminate(t *testing.T) {
	w := progress.New(progress.Indeterminate(true))

	if !w.IsIndeterminate() {
		t.Error("should be indeterminate mode")
	}
}

func TestNew_CustomPainter(t *testing.T) {
	p := &mockPainter{}
	w := progress.New(
		progress.Value(0.3),
		progress.PainterOpt(p),
	)
	w.SetBounds(geometry.NewRect(0, 0, 48, 48))
	ctx := widget.NewContext()
	canvas := &recordingCanvas{}

	w.Draw(ctx, canvas)

	if !p.called {
		t.Error("custom painter should have been called")
	}
}

// --- Value Clamping Tests ---

func TestValue_ClampedToRange(t *testing.T) {
	tests := []struct {
		name string
		in   float64
		want float64
	}{
		{"negative clamped to 0", -0.5, 0},
		{"zero stays zero", 0, 0},
		{"mid value stays", 0.5, 0.5},
		{"one stays one", 1.0, 1.0},
		{"over 1 clamped to 1", 1.5, 1.0},
		{"large value clamped", 100, 1.0},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			w := progress.New(progress.Value(tt.in))
			if got := w.Value(); got != tt.want {
				t.Errorf("Value() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestSetValue_UpdatesAndClamps(t *testing.T) {
	w := progress.New()

	w.SetValue(0.75)
	if w.Value() != 0.75 {
		t.Errorf("after SetValue(0.75): got %v", w.Value())
	}

	w.SetValue(-1)
	if w.Value() != 0 {
		t.Errorf("after SetValue(-1): got %v, want 0", w.Value())
	}

	w.SetValue(2)
	if w.Value() != 1 {
		t.Errorf("after SetValue(2): got %v, want 1", w.Value())
	}
}

func TestSetValue_TriggersRedraw(t *testing.T) {
	w := progress.New(progress.Value(0.5))
	w.ClearRedraw()

	w.SetValue(0.8)
	if !w.NeedsRedraw() {
		t.Error("SetValue should mark widget as needing redraw")
	}
}

func TestSetValue_SameValueNoRedraw(t *testing.T) {
	w := progress.New(progress.Value(0.5))
	w.ClearRedraw()

	w.SetValue(0.5)
	if w.NeedsRedraw() {
		t.Error("SetValue with same value should not mark redraw")
	}
}

// --- Layout Tests ---

func TestLayout_RespectsConstraints(t *testing.T) {
	ctx := widget.NewContext()
	constraints := geometry.Tight(geometry.Sz(64, 64))

	w := progress.New()
	size := w.Layout(ctx, constraints)

	if size.Width != 64 {
		t.Errorf("width = %v, want 64 (tight)", size.Width)
	}
	if size.Height != 64 {
		t.Errorf("height = %v, want 64 (tight)", size.Height)
	}
}

func TestLayout_PreferredSize(t *testing.T) {
	ctx := widget.NewContext()
	constraints := geometry.Loose(geometry.Sz(400, 400))

	w := progress.New()
	size := w.Layout(ctx, constraints)

	if size.Width <= 0 {
		t.Error("preferred width should be positive")
	}
	if size.Height <= 0 {
		t.Error("preferred height should be positive")
	}
	// Default diameter is 48.
	if size.Width != 48 {
		t.Errorf("default width = %v, want 48", size.Width)
	}
	if size.Height != 48 {
		t.Errorf("default height = %v, want 48", size.Height)
	}
}

func TestLayout_CustomSize(t *testing.T) {
	ctx := widget.NewContext()
	constraints := geometry.Loose(geometry.Sz(400, 400))

	w := progress.New(progress.Size(80))
	size := w.Layout(ctx, constraints)

	if size.Width != 80 {
		t.Errorf("width = %v, want 80 (custom size)", size.Width)
	}
	if size.Height != 80 {
		t.Errorf("height = %v, want 80 (custom size)", size.Height)
	}
}

// --- Draw Tests ---

func TestDraw_EmptyBounds(t *testing.T) {
	w := progress.New(progress.Value(0.5))
	ctx := widget.NewContext()
	canvas := &recordingCanvas{}

	w.Draw(ctx, canvas)

	if canvas.drawCount > 0 {
		t.Error("should not draw anything with empty bounds")
	}
}

func TestDraw_Determinate(t *testing.T) {
	w := progress.New(progress.Value(0.5))
	w.SetBounds(geometry.NewRect(0, 0, 48, 48))
	ctx := widget.NewContext()
	canvas := &recordingCanvas{}

	w.Draw(ctx, canvas)

	if canvas.drawCount == 0 {
		t.Error("should draw something with valid bounds")
	}
	// Should draw track circle + progress arc lines.
	if canvas.strokeCircleCount == 0 {
		t.Error("should draw track circle via StrokeCircle")
	}
	if canvas.lineCount == 0 {
		t.Error("should draw arc segments via DrawLine")
	}
}

func TestDraw_ZeroValue(t *testing.T) {
	w := progress.New(progress.Value(0))
	w.SetBounds(geometry.NewRect(0, 0, 48, 48))
	ctx := widget.NewContext()
	canvas := &recordingCanvas{}

	w.Draw(ctx, canvas)

	// Should draw track circle but no arc lines.
	if canvas.strokeCircleCount == 0 {
		t.Error("should draw track circle even at 0% value")
	}
	if canvas.lineCount > 0 {
		t.Error("should not draw arc lines at 0% value")
	}
}

func TestDraw_FullValue(t *testing.T) {
	w := progress.New(progress.Value(1.0))
	w.SetBounds(geometry.NewRect(0, 0, 48, 48))
	ctx := widget.NewContext()
	canvas := &recordingCanvas{}

	w.Draw(ctx, canvas)

	if canvas.strokeCircleCount == 0 {
		t.Error("should draw track circle at 100%")
	}
	if canvas.lineCount == 0 {
		t.Error("should draw full arc at 100%")
	}
}

func TestDraw_WithLabel(t *testing.T) {
	w := progress.New(
		progress.Value(0.65),
		progress.ShowLabel(true),
	)
	w.SetBounds(geometry.NewRect(0, 0, 48, 48))
	ctx := widget.NewContext()
	canvas := &recordingCanvas{}

	w.Draw(ctx, canvas)

	if !canvas.drewText {
		t.Error("should draw label text when ShowLabel is true")
	}
	if canvas.lastText != "65%" {
		t.Errorf("label = %q, want %q", canvas.lastText, "65%")
	}
}

func TestDraw_WithCustomFormatLabel(t *testing.T) {
	w := progress.New(
		progress.Value(0.333),
		progress.ShowLabel(true),
		progress.FormatLabelFn(func(_ float64) string {
			return "custom"
		}),
	)
	w.SetBounds(geometry.NewRect(0, 0, 48, 48))
	ctx := widget.NewContext()
	canvas := &recordingCanvas{}

	w.Draw(ctx, canvas)

	if canvas.lastText != "custom" {
		t.Errorf("label = %q, want %q", canvas.lastText, "custom")
	}
}

func TestDraw_WithoutLabel(t *testing.T) {
	w := progress.New(progress.Value(0.5))
	w.SetBounds(geometry.NewRect(0, 0, 48, 48))
	ctx := widget.NewContext()
	canvas := &recordingCanvas{}

	w.Draw(ctx, canvas)

	if canvas.drewText {
		t.Error("should not draw label text when ShowLabel is false")
	}
}

func TestDraw_Indeterminate(t *testing.T) {
	w := progress.New(progress.Indeterminate(true))
	w.SetBounds(geometry.NewRect(0, 0, 48, 48))
	ctx := widget.NewContext()
	ctx.SetNow(time.Now())
	canvas := &recordingCanvas{}

	w.Draw(ctx, canvas)

	if canvas.drawCount == 0 {
		t.Error("indeterminate should draw something")
	}
	if canvas.strokeCircleCount == 0 {
		t.Error("indeterminate should draw track circle")
	}
	if canvas.lineCount == 0 {
		t.Error("indeterminate should draw rotating arc segments")
	}
}

func TestDraw_IndeterminateNoLabel(t *testing.T) {
	w := progress.New(
		progress.Indeterminate(true),
		progress.ShowLabel(true), // Label should be ignored in indeterminate mode.
	)
	w.SetBounds(geometry.NewRect(0, 0, 48, 48))
	ctx := widget.NewContext()
	ctx.SetNow(time.Now())
	canvas := &recordingCanvas{}

	w.Draw(ctx, canvas)

	if canvas.drewText {
		t.Error("indeterminate mode should not show label")
	}
}

func TestDraw_IndeterminateRequestsRedraw(t *testing.T) {
	w := progress.New(progress.Indeterminate(true))
	w.SetBounds(geometry.NewRect(0, 0, 48, 48))
	ctx := widget.NewContext()
	ctx.SetNow(time.Now())
	canvas := &recordingCanvas{}

	w.ClearRedraw()
	w.Draw(ctx, canvas)

	if !w.NeedsRedraw() {
		t.Error("indeterminate should mark NeedsRedraw after draw")
	}
}

func TestDraw_DisabledState(t *testing.T) {
	w := progress.New(
		progress.Value(0.5),
		progress.Disabled(true),
	)
	w.SetBounds(geometry.NewRect(0, 0, 48, 48))
	ctx := widget.NewContext()
	canvas := &recordingCanvas{}

	w.Draw(ctx, canvas)

	if canvas.drawCount == 0 {
		t.Error("disabled indicator should still draw")
	}
}

func TestDraw_ColorScheme(t *testing.T) {
	cs := progress.ProgressColorScheme{
		Indicator: widget.ColorRed,
		Track:     widget.ColorGreen,
		Label:     widget.ColorBlue,
	}
	w := progress.New(
		progress.Value(0.5),
		progress.ColorSchemeOpt(cs),
		progress.ShowLabel(true),
	)
	w.SetBounds(geometry.NewRect(0, 0, 48, 48))
	ctx := widget.NewContext()
	canvas := &recordingCanvas{}

	w.Draw(ctx, canvas)

	if canvas.drawCount == 0 {
		t.Error("should draw with custom color scheme")
	}
}

func TestDraw_TrackColorOption(t *testing.T) {
	w := progress.New(
		progress.Value(0.5),
		progress.TrackColor(widget.ColorRed),
	)
	w.SetBounds(geometry.NewRect(0, 0, 48, 48))
	ctx := widget.NewContext()
	canvas := &recordingCanvas{}

	w.Draw(ctx, canvas)

	if canvas.drawCount == 0 {
		t.Error("should draw with custom track color")
	}
}

func TestDraw_IndicatorColorOption(t *testing.T) {
	w := progress.New(
		progress.Value(0.5),
		progress.IndicatorColor(widget.ColorBlue),
	)
	w.SetBounds(geometry.NewRect(0, 0, 48, 48))
	ctx := widget.NewContext()
	canvas := &recordingCanvas{}

	w.Draw(ctx, canvas)

	if canvas.drawCount == 0 {
		t.Error("should draw with custom indicator color")
	}
}

func TestDraw_SmallBounds(t *testing.T) {
	w := progress.New(progress.Value(0.5), progress.Size(100))
	// Bounds smaller than diameter -- should clamp to available space.
	w.SetBounds(geometry.NewRect(0, 0, 20, 20))
	ctx := widget.NewContext()
	canvas := &recordingCanvas{}

	w.Draw(ctx, canvas)

	// Should still draw within the available bounds.
	if canvas.drawCount == 0 {
		t.Error("should draw within small bounds")
	}
}

func TestDraw_TinyBoundsNoRender(t *testing.T) {
	w := progress.New(progress.Value(0.5), progress.StrokeWidth(20))
	// Bounds so small that radius <= 0 after subtracting stroke width.
	w.SetBounds(geometry.NewRect(0, 0, 10, 10))
	ctx := widget.NewContext()
	canvas := &recordingCanvas{}

	w.Draw(ctx, canvas)

	// Should not render when radius becomes <= 0.
	if canvas.strokeCircleCount > 0 || canvas.lineCount > 0 {
		t.Error("should not draw when effective radius is zero")
	}
}

// --- Event Tests ---

func TestEvent_NeverConsumes(t *testing.T) {
	w := progress.New(progress.Value(0.5))
	w.SetBounds(geometry.NewRect(0, 0, 48, 48))
	ctx := widget.NewContext()

	press := event.NewMouseEvent(event.MousePress, event.ButtonLeft, event.ButtonStateLeft,
		geometry.Pt(24, 24), geometry.Pt(24, 24), event.ModNone)
	consumed := w.Event(ctx, press)

	if consumed {
		t.Error("progress indicator should never consume events")
	}
}

func TestEvent_KeyboardNotConsumed(t *testing.T) {
	w := progress.New(progress.Value(0.5))
	ctx := widget.NewContext()

	keyEvt := event.NewKeyEvent(event.KeyPress, event.KeyRight, 0, event.ModNone)
	consumed := w.Event(ctx, keyEvt)

	if consumed {
		t.Error("progress indicator should not consume keyboard events")
	}
}

// --- Signal Binding Tests ---

func TestValueSignal_ReadFromSignal(t *testing.T) {
	sig := state.NewSignal[float64](0.75)
	w := progress.New(progress.ValueSignal(sig))
	w.SetBounds(geometry.NewRect(0, 0, 48, 48))
	ctx := widget.NewContext()
	canvas := &recordingCanvas{}

	w.Draw(ctx, canvas)

	if canvas.drawCount == 0 {
		t.Error("should draw when value is from signal")
	}
	if w.Value() != 0.75 {
		t.Errorf("Value() = %v, want 0.75 (from signal)", w.Value())
	}
}

func TestValueSignal_UpdatesOnSignalChange(t *testing.T) {
	sig := state.NewSignal[float64](0.25)
	w := progress.New(progress.ValueSignal(sig))

	if w.Value() != 0.25 {
		t.Errorf("initial value = %v, want 0.25", w.Value())
	}

	sig.Set(0.9)
	if w.Value() != 0.9 {
		t.Errorf("after signal change = %v, want 0.9", w.Value())
	}
}

func TestValueFn_DynamicValue(t *testing.T) {
	val := 0.4
	w := progress.New(progress.ValueFn(func() float64 { return val }))

	if w.Value() != 0.4 {
		t.Errorf("Value() = %v, want 0.4", w.Value())
	}

	val = 0.8
	if w.Value() != 0.8 {
		t.Errorf("after change Value() = %v, want 0.8", w.Value())
	}
}

func TestValueSignal_PrecedenceOverFn(t *testing.T) {
	sig := state.NewSignal[float64](0.9)
	w := progress.New(
		progress.Value(0.1),
		progress.ValueFn(func() float64 { return 0.5 }),
		progress.ValueSignal(sig),
	)

	if w.Value() != 0.9 {
		t.Errorf("signal should take precedence, got %v", w.Value())
	}
}

func TestValueReadonlySignal(t *testing.T) {
	sig := state.NewSignal[float64](0.6)
	w := progress.New(progress.ValueReadonlySignal(sig))

	if w.Value() != 0.6 {
		t.Errorf("Value() = %v, want 0.6 (from readonly signal)", w.Value())
	}

	sig.Set(0.2)
	if w.Value() != 0.2 {
		t.Errorf("after change Value() = %v, want 0.2", w.Value())
	}
}

func TestValueReadonlySignal_HighestPrecedence(t *testing.T) {
	roSig := state.NewSignal[float64](0.99)
	rwSig := state.NewSignal[float64](0.11)
	w := progress.New(
		progress.Value(0.1),
		progress.ValueFn(func() float64 { return 0.5 }),
		progress.ValueSignal(rwSig),
		progress.ValueReadonlySignal(roSig),
	)

	if w.Value() != 0.99 {
		t.Errorf("readonly signal should have highest precedence, got %v", w.Value())
	}
}

func TestDisabledSignal(t *testing.T) {
	sig := state.NewSignal(true)
	w := progress.New(progress.DisabledSignal(sig))

	// DisabledSignal affects painting, not WidgetBase.IsEnabled.
	if !w.IsEnabled() {
		t.Error("WidgetBase.IsEnabled should still be true")
	}

	w.SetBounds(geometry.NewRect(0, 0, 48, 48))
	ctx := widget.NewContext()
	canvas := &recordingCanvas{}
	w.Draw(ctx, canvas)

	if canvas.drawCount == 0 {
		t.Error("disabled indicator should still draw")
	}
}

func TestDisabledFn(t *testing.T) {
	disabled := true
	w := progress.New(
		progress.Value(0.5),
		progress.DisabledFn(func() bool { return disabled }),
	)
	w.SetBounds(geometry.NewRect(0, 0, 48, 48))
	ctx := widget.NewContext()
	canvas := &recordingCanvas{}

	w.Draw(ctx, canvas)
	if canvas.drawCount == 0 {
		t.Error("should draw when disabled via DisabledFn")
	}
}

func TestDisabledReadonlySignal(t *testing.T) {
	sig := state.NewSignal(true)
	w := progress.New(
		progress.Value(0.5),
		progress.DisabledReadonlySignal(sig),
	)
	w.SetBounds(geometry.NewRect(0, 0, 48, 48))
	ctx := widget.NewContext()
	canvas := &recordingCanvas{}

	w.Draw(ctx, canvas)
	if canvas.drawCount == 0 {
		t.Error("should draw when disabled via DisabledReadonlySignal")
	}
}

func TestDraw_DisabledWithColorScheme(t *testing.T) {
	cs := progress.ProgressColorScheme{
		Indicator:         widget.ColorRed,
		Track:             widget.ColorGreen,
		DisabledIndicator: widget.ColorGray,
		DisabledTrack:     widget.ColorDarkGray,
	}
	w := progress.New(
		progress.Value(0.5),
		progress.ColorSchemeOpt(cs),
		progress.Disabled(true),
	)
	w.SetBounds(geometry.NewRect(0, 0, 48, 48))
	ctx := widget.NewContext()
	canvas := &recordingCanvas{}

	w.Draw(ctx, canvas)
	if canvas.drawCount == 0 {
		t.Error("disabled indicator with color scheme should draw")
	}
}

// --- Mount/Unmount Tests ---

func TestMount_CreatesBindings(t *testing.T) {
	sig := state.NewSignal[float64](0.5)
	w := progress.New(progress.ValueSignal(sig))

	sched := &mockScheduler{}
	ctx := widget.NewContext()
	ctx.SetScheduler(sched)

	w.Mount(ctx)

	sig.Set(0.8)
	if sched.dirtyCount == 0 {
		t.Error("scheduler should be notified after signal change")
	}
}

func TestMount_NilScheduler(t *testing.T) {
	w := progress.New(progress.ValueSignal(state.NewSignal[float64](0.5)))
	ctx := widget.NewContext()

	// Should not panic.
	w.Mount(ctx)
}

func TestMount_ReadonlyValueSignal(t *testing.T) {
	sig := state.NewSignal[float64](0.5)
	w := progress.New(progress.ValueReadonlySignal(sig))

	sched := &mockScheduler{}
	ctx := widget.NewContext()
	ctx.SetScheduler(sched)

	w.Mount(ctx)
	sig.Set(0.9)
	if sched.dirtyCount == 0 {
		t.Error("scheduler should be notified for readonly value signal change")
	}
}

func TestMount_DisabledSignalBinding(t *testing.T) {
	sig := state.NewSignal(false)
	w := progress.New(progress.DisabledSignal(sig))

	sched := &mockScheduler{}
	ctx := widget.NewContext()
	ctx.SetScheduler(sched)

	w.Mount(ctx)
	sig.Set(true)
	if sched.dirtyCount == 0 {
		t.Error("scheduler should be notified for disabled signal change")
	}
}

func TestMount_DisabledReadonlySignalBinding(t *testing.T) {
	sig := state.NewSignal(false)
	w := progress.New(progress.DisabledReadonlySignal(sig))

	sched := &mockScheduler{}
	ctx := widget.NewContext()
	ctx.SetScheduler(sched)

	w.Mount(ctx)
	sig.Set(true)
	if sched.dirtyCount == 0 {
		t.Error("scheduler should be notified for disabled readonly signal change")
	}
}

func TestUnmount_CleanupBindings(t *testing.T) {
	sig := state.NewSignal[float64](0.5)
	w := progress.New(progress.ValueSignal(sig))

	sched := &mockScheduler{}
	ctx := widget.NewContext()
	ctx.SetScheduler(sched)

	w.Mount(ctx)
	w.CleanupBindings()
	w.Unmount()

	sched.dirtyCount = 0
	sig.Set(0.9)
	if sched.dirtyCount > 0 {
		t.Error("scheduler should not be notified after cleanup")
	}
}

func TestUnmount_NoOp(t *testing.T) {
	w := progress.New()
	// Should not panic.
	w.Unmount()
}

// --- Interface Compliance Tests ---

func TestWidget_Interface(t *testing.T) {
	var w widget.Widget = progress.New()
	_ = w
}

func TestLifecycle_Interface(t *testing.T) {
	var l widget.Lifecycle = progress.New()
	_ = l
}

// --- Indeterminate Animation Tests ---

func TestDraw_IndeterminateRotationChanges(t *testing.T) {
	w := progress.New(progress.Indeterminate(true))
	w.SetBounds(geometry.NewRect(0, 0, 48, 48))

	now := time.Now()
	ctx := widget.NewContext()
	ctx.SetNow(now)
	canvas1 := &recordingCanvas{}
	w.Draw(ctx, canvas1)

	// Advance time.
	ctx.SetNow(now.Add(500 * time.Millisecond))
	canvas2 := &recordingCanvas{}
	w.Draw(ctx, canvas2)

	// Both should draw, confirming continuous animation works.
	if canvas1.lineCount == 0 {
		t.Error("first frame should draw arc lines")
	}
	if canvas2.lineCount == 0 {
		t.Error("second frame should draw arc lines")
	}
}

// --- Mock types ---

type mockPainter struct {
	called bool
}

func (p *mockPainter) PaintProgress(_ widget.Canvas, _ progress.PaintState) {
	p.called = true
}

type mockScheduler struct {
	dirtyCount int
}

func (s *mockScheduler) MarkDirty(_ widget.Widget) {
	s.dirtyCount++
}

// recordingCanvas is a minimal mock for external tests.
type recordingCanvas struct {
	drawCount         int
	strokeCircleCount int
	lineCount         int
	drewText          bool
	lastText          string
}

func (c *recordingCanvas) Clear(_ widget.Color)                                     {}
func (c *recordingCanvas) DrawRect(_ geometry.Rect, _ widget.Color)                 { c.drawCount++ }
func (c *recordingCanvas) StrokeRect(_ geometry.Rect, _ widget.Color, _ float32)    { c.drawCount++ }
func (c *recordingCanvas) DrawRoundRect(_ geometry.Rect, _ widget.Color, _ float32) { c.drawCount++ }
func (c *recordingCanvas) StrokeRoundRect(_ geometry.Rect, _ widget.Color, _ float32, _ float32) {
	c.drawCount++
}
func (c *recordingCanvas) DrawCircle(_ geometry.Point, _ float32, _ widget.Color) {
	c.drawCount++
}
func (c *recordingCanvas) StrokeCircle(_ geometry.Point, _ float32, _ widget.Color, _ float32) {
	c.drawCount++
	c.strokeCircleCount++
}
func (c *recordingCanvas) DrawLine(_, _ geometry.Point, _ widget.Color, _ float32) {
	c.drawCount++
	c.lineCount++
}
func (c *recordingCanvas) DrawText(text string, _ geometry.Rect, _ float32, _ widget.Color, _ bool, _ widget.TextAlign) {
	c.drawCount++
	c.drewText = true
	c.lastText = text
}

func (c *recordingCanvas) MeasureText(text string, fontSize float32, _ bool) float32 {
	return float32(len([]rune(text))) * fontSize * 0.5
}
func (c *recordingCanvas) DrawImage(_ image.Image, _ geometry.Point)    { c.drawCount++ }
func (c *recordingCanvas) PushClip(_ geometry.Rect)                     {}
func (c *recordingCanvas) PushClipRoundRect(_ geometry.Rect, _ float32) {}
func (c *recordingCanvas) PopClip()                                     {}
func (c *recordingCanvas) PushTransform(_ geometry.Point)               {}
func (c *recordingCanvas) PopTransform()                                {}
func (c *recordingCanvas) TransformOffset() geometry.Point              { return geometry.Point{} }
