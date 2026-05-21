package progressbar_test

import (
	"github.com/gogpu/gg/scene"
	"image"
	"testing"

	"github.com/gogpu/ui/core/progressbar"
	"github.com/gogpu/ui/event"
	"github.com/gogpu/ui/geometry"
	"github.com/gogpu/ui/state"
	"github.com/gogpu/ui/widget"
)

// --- Construction Tests ---

func TestNew_Defaults(t *testing.T) {
	bar := progressbar.New()

	if !bar.IsVisible() {
		t.Error("default progress bar should be visible")
	}
	if !bar.IsEnabled() {
		t.Error("default progress bar should be enabled")
	}
	if bar.Children() != nil {
		t.Error("progress bar should have no children")
	}
	if bar.Value() != 0 {
		t.Errorf("default value should be 0, got %v", bar.Value())
	}
}

func TestNew_WithValue(t *testing.T) {
	bar := progressbar.New(progressbar.Value(0.65))

	if bar.Value() != 0.65 {
		t.Errorf("value = %v, want 0.65", bar.Value())
	}
}

func TestNew_WithOptions(t *testing.T) {
	bar := progressbar.New(
		progressbar.Value(0.5),
		progressbar.Height(20),
		progressbar.Radius(6),
		progressbar.ShowLabel(true),
		progressbar.Disabled(false),
	)

	if bar.Value() != 0.5 {
		t.Errorf("value = %v, want 0.5", bar.Value())
	}
	if !bar.IsEnabled() {
		t.Error("should be enabled")
	}
}

func TestNew_CustomPainter(t *testing.T) {
	p := &mockPainter{}
	bar := progressbar.New(
		progressbar.Value(0.3),
		progressbar.PainterOpt(p),
	)
	bar.SetBounds(geometry.NewRect(0, 0, 200, 30))
	ctx := widget.NewContext()
	canvas := &recordingCanvas{}

	bar.Draw(ctx, canvas)

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
			bar := progressbar.New(progressbar.Value(tt.in))
			if got := bar.Value(); got != tt.want {
				t.Errorf("Value() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestSetValue_UpdatesAndClamps(t *testing.T) {
	bar := progressbar.New()

	bar.SetValue(0.75)
	if bar.Value() != 0.75 {
		t.Errorf("after SetValue(0.75): got %v", bar.Value())
	}

	bar.SetValue(-1)
	if bar.Value() != 0 {
		t.Errorf("after SetValue(-1): got %v, want 0", bar.Value())
	}

	bar.SetValue(2)
	if bar.Value() != 1 {
		t.Errorf("after SetValue(2): got %v, want 1", bar.Value())
	}
}

func TestSetValue_TriggersRedraw(t *testing.T) {
	bar := progressbar.New(progressbar.Value(0.5))
	// Clear any initial redraw flag.
	bar.ClearRedraw()

	bar.SetValue(0.8)
	if !bar.NeedsRedraw() {
		t.Error("SetValue should mark widget as needing redraw")
	}
}

func TestSetValue_SameValueNoRedraw(t *testing.T) {
	bar := progressbar.New(progressbar.Value(0.5))
	bar.ClearRedraw()

	bar.SetValue(0.5)
	if bar.NeedsRedraw() {
		t.Error("SetValue with same value should not mark redraw")
	}
}

// --- Layout Tests ---

func TestLayout_RespectsConstraints(t *testing.T) {
	ctx := widget.NewContext()
	constraints := geometry.Tight(geometry.Sz(300, 40))

	bar := progressbar.New()
	size := bar.Layout(ctx, constraints)

	if size.Width != 300 {
		t.Errorf("width = %v, want 300 (tight)", size.Width)
	}
	if size.Height != 40 {
		t.Errorf("height = %v, want 40 (tight)", size.Height)
	}
}

func TestLayout_PreferredSize(t *testing.T) {
	ctx := widget.NewContext()
	constraints := geometry.Loose(geometry.Sz(400, 400))

	bar := progressbar.New()
	size := bar.Layout(ctx, constraints)

	if size.Width <= 0 {
		t.Error("preferred width should be positive")
	}
	if size.Height <= 0 {
		t.Error("preferred height should be positive")
	}
}

func TestLayout_CustomHeight(t *testing.T) {
	ctx := widget.NewContext()
	constraints := geometry.Loose(geometry.Sz(400, 400))

	bar := progressbar.New(progressbar.Height(24))
	size := bar.Layout(ctx, constraints)

	// Height should be barHeight + 2*padding (4 default padding).
	if size.Height < 24 {
		t.Errorf("height = %v, want >= 24 (custom height)", size.Height)
	}
}

// --- Draw Tests ---

func TestDraw_EmptyBounds(t *testing.T) {
	bar := progressbar.New(progressbar.Value(0.5))
	ctx := widget.NewContext()
	canvas := &recordingCanvas{}

	bar.Draw(ctx, canvas)

	if canvas.drawCount > 0 {
		t.Error("should not draw anything with empty bounds")
	}
}

func TestDraw_NormalState(t *testing.T) {
	bar := progressbar.New(progressbar.Value(0.5))
	bar.SetBounds(geometry.NewRect(0, 0, 200, 30))
	ctx := widget.NewContext()
	canvas := &recordingCanvas{}

	bar.Draw(ctx, canvas)

	if canvas.drawCount == 0 {
		t.Error("should draw something with valid bounds")
	}
}

func TestDraw_ZeroValue(t *testing.T) {
	bar := progressbar.New(progressbar.Value(0))
	bar.SetBounds(geometry.NewRect(0, 0, 200, 30))
	ctx := widget.NewContext()
	canvas := &recordingCanvas{}

	bar.Draw(ctx, canvas)

	// Should at least draw the track.
	if canvas.drawCount == 0 {
		t.Error("should draw track even at 0% value")
	}
}

func TestDraw_FullValue(t *testing.T) {
	bar := progressbar.New(progressbar.Value(1.0))
	bar.SetBounds(geometry.NewRect(0, 0, 200, 30))
	ctx := widget.NewContext()
	canvas := &recordingCanvas{}

	bar.Draw(ctx, canvas)

	if canvas.drawCount < 2 {
		t.Errorf("should draw track + fill at 100%%, got %d draws", canvas.drawCount)
	}
}

func TestDraw_WithLabel(t *testing.T) {
	bar := progressbar.New(
		progressbar.Value(0.65),
		progressbar.ShowLabel(true),
	)
	bar.SetBounds(geometry.NewRect(0, 0, 200, 30))
	ctx := widget.NewContext()
	canvas := &recordingCanvas{}

	bar.Draw(ctx, canvas)

	if !canvas.drewText {
		t.Error("should draw label text when ShowLabel is true")
	}
	if canvas.lastText != "65%" {
		t.Errorf("label = %q, want %q", canvas.lastText, "65%")
	}
}

func TestDraw_WithCustomFormatLabel(t *testing.T) {
	bar := progressbar.New(
		progressbar.Value(0.333),
		progressbar.ShowLabel(true),
		progressbar.FormatLabelFn(func(v float64) string {
			return "custom"
		}),
	)
	bar.SetBounds(geometry.NewRect(0, 0, 200, 30))
	ctx := widget.NewContext()
	canvas := &recordingCanvas{}

	bar.Draw(ctx, canvas)

	if canvas.lastText != "custom" {
		t.Errorf("label = %q, want %q", canvas.lastText, "custom")
	}
}

func TestDraw_WithoutLabel(t *testing.T) {
	bar := progressbar.New(progressbar.Value(0.5))
	bar.SetBounds(geometry.NewRect(0, 0, 200, 30))
	ctx := widget.NewContext()
	canvas := &recordingCanvas{}

	bar.Draw(ctx, canvas)

	if canvas.drewText {
		t.Error("should not draw label text when ShowLabel is false")
	}
}

func TestDraw_SquareCorners(t *testing.T) {
	bar := progressbar.New(
		progressbar.Value(0.5),
		progressbar.Radius(0),
	)
	bar.SetBounds(geometry.NewRect(0, 0, 200, 30))
	ctx := widget.NewContext()
	canvas := &recordingCanvas{}

	bar.Draw(ctx, canvas)

	if canvas.roundRectCount > 0 {
		t.Error("should use DrawRect, not DrawRoundRect, when radius is 0")
	}
	if canvas.rectCount == 0 {
		t.Error("should use DrawRect for square corners")
	}
}

func TestDraw_RoundedCorners(t *testing.T) {
	bar := progressbar.New(
		progressbar.Value(0.5),
		progressbar.Radius(8),
	)
	bar.SetBounds(geometry.NewRect(0, 0, 200, 30))
	ctx := widget.NewContext()
	canvas := &recordingCanvas{}

	bar.Draw(ctx, canvas)

	if canvas.roundRectCount == 0 {
		t.Error("should use DrawRoundRect when radius > 0")
	}
}

func TestDraw_DisabledState(t *testing.T) {
	bar := progressbar.New(
		progressbar.Value(0.5),
		progressbar.Disabled(true),
	)
	bar.SetBounds(geometry.NewRect(0, 0, 200, 30))
	ctx := widget.NewContext()
	canvas := &recordingCanvas{}

	bar.Draw(ctx, canvas)

	if canvas.drawCount == 0 {
		t.Error("disabled bar should still draw (with dimmed appearance)")
	}
}

func TestDraw_ProgressBarColorScheme(t *testing.T) {
	cs := progressbar.ProgressBarColorScheme{
		Bar:   widget.ColorRed,
		Track: widget.ColorGreen,
		Label: widget.ColorBlue,
	}
	bar := progressbar.New(
		progressbar.Value(0.5),
		progressbar.ColorSchemeOpt(cs),
		progressbar.ShowLabel(true),
	)
	bar.SetBounds(geometry.NewRect(0, 0, 200, 30))
	ctx := widget.NewContext()
	canvas := &recordingCanvas{}

	bar.Draw(ctx, canvas)

	if canvas.drawCount == 0 {
		t.Error("should draw with custom color scheme")
	}
}

// --- Event Tests ---

func TestEvent_NeverConsumes(t *testing.T) {
	bar := progressbar.New(progressbar.Value(0.5))
	bar.SetBounds(geometry.NewRect(0, 0, 200, 30))
	ctx := widget.NewContext()

	press := event.NewMouseEvent(event.MousePress, event.ButtonLeft, event.ButtonStateLeft,
		geometry.Pt(100, 15), geometry.Pt(100, 15), event.ModNone)
	consumed := bar.Event(ctx, press)

	if consumed {
		t.Error("progress bar should never consume events")
	}
}

func TestEvent_KeyboardNotConsumed(t *testing.T) {
	bar := progressbar.New(progressbar.Value(0.5))
	ctx := widget.NewContext()

	keyEvt := event.NewKeyEvent(event.KeyPress, event.KeyRight, 0, event.ModNone)
	consumed := bar.Event(ctx, keyEvt)

	if consumed {
		t.Error("progress bar should not consume keyboard events")
	}
}

// --- Signal Binding Tests ---

func TestValueSignal_ReadFromSignal(t *testing.T) {
	sig := state.NewSignal[float64](0.75)
	bar := progressbar.New(progressbar.ValueSignal(sig))
	bar.SetBounds(geometry.NewRect(0, 0, 200, 30))
	ctx := widget.NewContext()
	canvas := &recordingCanvas{}

	bar.Draw(ctx, canvas)

	if canvas.drawCount == 0 {
		t.Error("should draw when value is from signal")
	}
	if bar.Value() != 0.75 {
		t.Errorf("Value() = %v, want 0.75 (from signal)", bar.Value())
	}
}

func TestValueSignal_UpdatesOnSignalChange(t *testing.T) {
	sig := state.NewSignal[float64](0.25)
	bar := progressbar.New(progressbar.ValueSignal(sig))

	if bar.Value() != 0.25 {
		t.Errorf("initial value = %v, want 0.25", bar.Value())
	}

	sig.Set(0.9)
	if bar.Value() != 0.9 {
		t.Errorf("after signal change = %v, want 0.9", bar.Value())
	}
}

func TestValueFn_DynamicValue(t *testing.T) {
	val := 0.4
	bar := progressbar.New(progressbar.ValueFn(func() float64 { return val }))

	if bar.Value() != 0.4 {
		t.Errorf("Value() = %v, want 0.4", bar.Value())
	}

	val = 0.8
	if bar.Value() != 0.8 {
		t.Errorf("after change Value() = %v, want 0.8", bar.Value())
	}
}

func TestValueSignal_PrecedenceOverFn(t *testing.T) {
	sig := state.NewSignal[float64](0.9)
	bar := progressbar.New(
		progressbar.Value(0.1),
		progressbar.ValueFn(func() float64 { return 0.5 }),
		progressbar.ValueSignal(sig),
	)

	if bar.Value() != 0.9 {
		t.Errorf("signal should take precedence, got %v", bar.Value())
	}
}

func TestDisabledSignal(t *testing.T) {
	sig := state.NewSignal(true)
	bar := progressbar.New(progressbar.DisabledSignal(sig))

	// DisabledSignal affects painting (via config.ResolvedDisabled), not WidgetBase.IsEnabled.
	// WidgetBase.IsEnabled is always true unless explicitly set to false.
	if !bar.IsEnabled() {
		t.Error("WidgetBase.IsEnabled should still be true; DisabledSignal controls painting only")
	}

	// Draw should use disabled colors.
	bar.SetBounds(geometry.NewRect(0, 0, 200, 30))
	ctx := widget.NewContext()
	canvas := &recordingCanvas{}
	bar.Draw(ctx, canvas)

	// Verify bar draws (disabled bars still render).
	if canvas.drawCount == 0 {
		t.Error("disabled bar should still draw")
	}
}

// --- Mount/Unmount Tests ---

func TestMount_CreatesBindings(t *testing.T) {
	sig := state.NewSignal[float64](0.5)
	bar := progressbar.New(progressbar.ValueSignal(sig))

	sched := &mockScheduler{}
	ctx := widget.NewContext()
	ctx.SetScheduler(sched)

	bar.Mount(ctx)

	// Change signal -- scheduler should be notified.
	sig.Set(0.8)
	if sched.dirtyCount == 0 {
		t.Error("scheduler should be notified after signal change")
	}
}

func TestMount_NilScheduler(t *testing.T) {
	bar := progressbar.New(progressbar.ValueSignal(state.NewSignal[float64](0.5)))
	ctx := widget.NewContext()

	// Should not panic.
	bar.Mount(ctx)
}

func TestUnmount_CleanupBindings(t *testing.T) {
	sig := state.NewSignal[float64](0.5)
	bar := progressbar.New(progressbar.ValueSignal(sig))

	sched := &mockScheduler{}
	ctx := widget.NewContext()
	ctx.SetScheduler(sched)

	bar.Mount(ctx)
	bar.CleanupBindings()
	bar.Unmount()

	// After cleanup, signal changes should not notify scheduler.
	sched.dirtyCount = 0
	sig.Set(0.9)
	if sched.dirtyCount > 0 {
		t.Error("scheduler should not be notified after cleanup")
	}
}

// --- Fluent Styling Tests ---

func TestPadding_Chaining(t *testing.T) {
	bar := progressbar.New()
	result := bar.Padding(16)

	if result != bar {
		t.Error("Padding should return same widget for chaining")
	}
}

// --- Interface Compliance Tests ---

func TestWidget_Interface(t *testing.T) {
	var w widget.Widget = progressbar.New()
	_ = w
}

func TestLifecycle_Interface(t *testing.T) {
	var l widget.Lifecycle = progressbar.New()
	_ = l
}

// --- ReadonlySignal Tests ---

func TestValueReadonlySignal(t *testing.T) {
	sig := state.NewSignal[float64](0.6)
	bar := progressbar.New(progressbar.ValueReadonlySignal(sig))

	if bar.Value() != 0.6 {
		t.Errorf("Value() = %v, want 0.6 (from readonly signal)", bar.Value())
	}

	sig.Set(0.2)
	if bar.Value() != 0.2 {
		t.Errorf("after change Value() = %v, want 0.2", bar.Value())
	}
}

func TestValueReadonlySignal_HighestPrecedence(t *testing.T) {
	roSig := state.NewSignal[float64](0.99)
	rwSig := state.NewSignal[float64](0.11)
	bar := progressbar.New(
		progressbar.Value(0.1),
		progressbar.ValueFn(func() float64 { return 0.5 }),
		progressbar.ValueSignal(rwSig),
		progressbar.ValueReadonlySignal(roSig),
	)

	if bar.Value() != 0.99 {
		t.Errorf("readonly signal should have highest precedence, got %v", bar.Value())
	}
}

func TestDisabledFn(t *testing.T) {
	disabled := true
	bar := progressbar.New(
		progressbar.Value(0.5),
		progressbar.DisabledFn(func() bool { return disabled }),
	)
	bar.SetBounds(geometry.NewRect(0, 0, 200, 30))
	ctx := widget.NewContext()
	canvas := &recordingCanvas{}

	bar.Draw(ctx, canvas)
	if canvas.drawCount == 0 {
		t.Error("should draw when disabled via DisabledFn")
	}
}

func TestDisabledReadonlySignal(t *testing.T) {
	sig := state.NewSignal(true)
	bar := progressbar.New(
		progressbar.Value(0.5),
		progressbar.DisabledReadonlySignal(sig),
	)
	bar.SetBounds(geometry.NewRect(0, 0, 200, 30))
	ctx := widget.NewContext()
	canvas := &recordingCanvas{}

	bar.Draw(ctx, canvas)
	if canvas.drawCount == 0 {
		t.Error("should draw when disabled via DisabledReadonlySignal")
	}
}

// --- Mount with various signal combos ---

func TestMount_ReadonlyValueSignal(t *testing.T) {
	sig := state.NewSignal[float64](0.5)
	bar := progressbar.New(progressbar.ValueReadonlySignal(sig))

	sched := &mockScheduler{}
	ctx := widget.NewContext()
	ctx.SetScheduler(sched)

	bar.Mount(ctx)
	sig.Set(0.9)
	if sched.dirtyCount == 0 {
		t.Error("scheduler should be notified for readonly value signal change")
	}
}

func TestMount_DisabledSignalBinding(t *testing.T) {
	sig := state.NewSignal(false)
	bar := progressbar.New(progressbar.DisabledSignal(sig))

	sched := &mockScheduler{}
	ctx := widget.NewContext()
	ctx.SetScheduler(sched)

	bar.Mount(ctx)
	sig.Set(true)
	if sched.dirtyCount == 0 {
		t.Error("scheduler should be notified for disabled signal change")
	}
}

func TestMount_DisabledReadonlySignalBinding(t *testing.T) {
	sig := state.NewSignal(false)
	bar := progressbar.New(progressbar.DisabledReadonlySignal(sig))

	sched := &mockScheduler{}
	ctx := widget.NewContext()
	ctx.SetScheduler(sched)

	bar.Mount(ctx)
	sig.Set(true)
	if sched.dirtyCount == 0 {
		t.Error("scheduler should be notified for disabled readonly signal change")
	}
}

func TestUnmount_NoOp(t *testing.T) {
	bar := progressbar.New()
	// Should not panic.
	bar.Unmount()
}

// --- Disabled color resolution ---

func TestDraw_DisabledWithProgressBarColorScheme(t *testing.T) {
	cs := progressbar.ProgressBarColorScheme{
		Bar:           widget.ColorRed,
		Track:         widget.ColorGreen,
		DisabledBar:   widget.ColorGray,
		DisabledTrack: widget.ColorDarkGray,
	}
	bar := progressbar.New(
		progressbar.Value(0.5),
		progressbar.ColorSchemeOpt(cs),
		progressbar.Disabled(true),
	)
	bar.SetBounds(geometry.NewRect(0, 0, 200, 30))
	ctx := widget.NewContext()
	canvas := &recordingCanvas{}

	bar.Draw(ctx, canvas)
	if canvas.drawCount == 0 {
		t.Error("disabled bar with color scheme should draw")
	}
}

// --- Signal Deduplication Tests (#112) ---

// TestMount_ValueSignal_DeduplicatesRedraws verifies that when a value signal
// fires with the same value that was last drawn, the widget is NOT marked
// dirty. This prevents flicker from redundant redraws (issue #112).
func TestMount_ValueSignal_DeduplicatesRedraws(t *testing.T) {
	sig := state.NewSignal[float64](0.5)
	bar := progressbar.New(progressbar.ValueSignal(sig))
	bar.SetBounds(geometry.NewRect(0, 0, 200, 30))

	sched := &mockScheduler{}
	ctx := widget.NewContext()
	ctx.SetScheduler(sched)
	bar.Mount(ctx)

	// First Draw populates lastDrawnValue.
	canvas := &recordingCanvas{}
	bar.Draw(ctx, canvas)

	// Reset dirty counter.
	sched.dirtyCount = 0

	// Set signal to the SAME value — should NOT mark dirty.
	sig.Set(0.5)
	if sched.dirtyCount != 0 {
		t.Errorf("setting same value should not mark dirty, got dirtyCount=%d", sched.dirtyCount)
	}

	// Set to a DIFFERENT value — SHOULD mark dirty.
	sig.Set(0.8)
	if sched.dirtyCount == 0 {
		t.Error("setting different value should mark dirty")
	}
}

// TestMount_ValueSignal_FirstSetAlwaysDirties verifies that before the first
// Draw call, any signal change marks dirty (since no cached value exists yet).
func TestMount_ValueSignal_FirstSetAlwaysDirties(t *testing.T) {
	sig := state.NewSignal[float64](0.5)
	bar := progressbar.New(progressbar.ValueSignal(sig))

	sched := &mockScheduler{}
	ctx := widget.NewContext()
	ctx.SetScheduler(sched)
	bar.Mount(ctx)

	// No Draw yet — any Set should mark dirty.
	sig.Set(0.5) // same as initial, but never drawn
	if sched.dirtyCount == 0 {
		t.Error("before first Draw, any signal change should mark dirty")
	}
}

// --- Mock types ---

type mockPainter struct {
	called bool
}

func (p *mockPainter) PaintProgressBar(_ widget.Canvas, _ progressbar.PaintState) {
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
	drawCount      int
	rectCount      int
	roundRectCount int
	drewText       bool
	lastText       string
	clipCount      int
}

func (c *recordingCanvas) Clear(_ widget.Color)                                  {}
func (c *recordingCanvas) DrawRect(_ geometry.Rect, _ widget.Color)              { c.drawCount++; c.rectCount++ }
func (c *recordingCanvas) FillRectDirect(_ geometry.Rect, _ widget.Color)        {}
func (c *recordingCanvas) StrokeRect(_ geometry.Rect, _ widget.Color, _ float32) { c.drawCount++ }
func (c *recordingCanvas) DrawRoundRect(_ geometry.Rect, _ widget.Color, _ float32) {
	c.drawCount++
	c.roundRectCount++
}
func (c *recordingCanvas) StrokeRoundRect(_ geometry.Rect, _ widget.Color, _ float32, _ float32) {
	c.drawCount++
}
func (c *recordingCanvas) DrawCircle(_ geometry.Point, _ float32, _ widget.Color) {
	c.drawCount++
}
func (c *recordingCanvas) StrokeCircle(_ geometry.Point, _ float32, _ widget.Color, _ float32) {
	c.drawCount++
}
func (c *recordingCanvas) StrokeArc(_ geometry.Point, _ float32, _, _ float64, _ widget.Color, _ float32) {
}
func (c *recordingCanvas) DrawLine(_, _ geometry.Point, _ widget.Color, _ float32) { c.drawCount++ }
func (c *recordingCanvas) DrawText(text string, _ geometry.Rect, _ float32, _ widget.Color, _ bool, _ widget.TextAlign) {
	c.drawCount++
	c.drewText = true
	c.lastText = text
}

func (c *recordingCanvas) MeasureText(text string, fontSize float32, _ bool) float32 {
	return float32(len([]rune(text))) * fontSize * 0.5
}
func (c *recordingCanvas) DrawImage(_ image.Image, _ geometry.Point) { c.drawCount++ }
func (c *recordingCanvas) PushClip(_ geometry.Rect)                  { c.clipCount++ }
func (c *recordingCanvas) PushClipRoundRect(_ geometry.Rect, _ float32) {
	c.clipCount++
}
func (c *recordingCanvas) PopClip()                         { c.clipCount-- }
func (c *recordingCanvas) PushTransform(_ geometry.Point)   {}
func (c *recordingCanvas) PopTransform()                    {}
func (c *recordingCanvas) TransformOffset() geometry.Point  { return geometry.Point{} }
func (c *recordingCanvas) ScreenOriginBase() geometry.Point { return geometry.Point{} }
func (c *recordingCanvas) ClipBounds() geometry.Rect        { return geometry.NewRect(0, 0, 10000, 10000) }
func (c *recordingCanvas) ReplayScene(_ *scene.Scene)       {}
