package slider_test

import (
	"image"
	"testing"

	"github.com/gogpu/ui/core/slider"
	"github.com/gogpu/ui/event"
	"github.com/gogpu/ui/geometry"
	"github.com/gogpu/ui/state"
	"github.com/gogpu/ui/widget"
)

// --- Construction Tests ---

func TestNew_Defaults(t *testing.T) {
	s := slider.New()

	if !s.IsVisible() {
		t.Error("default slider should be visible")
	}
	if !s.IsEnabled() {
		t.Error("default slider should be enabled")
	}
	if !s.IsFocusable() {
		t.Error("default slider should be focusable")
	}
	if s.Children() != nil {
		t.Error("slider should have no children")
	}
}

func TestNew_WithOptions(t *testing.T) {
	changed := false
	s := slider.New(
		slider.Min(10),
		slider.Max(200),
		slider.Value(50),
		slider.Step(5),
		slider.OnChange(func(float32) { changed = true }),
		slider.OrientationOpt(slider.Vertical),
		slider.Disabled(false),
		slider.A11yHint("volume control"),
		slider.Marks([]slider.Mark{{Value: 25}, {Value: 75}}),
	)

	if !s.IsFocusable() {
		t.Error("should be focusable")
	}
	_ = changed
}

func TestNew_DisabledNotFocusable(t *testing.T) {
	s := slider.New(slider.Disabled(true))
	if s.IsFocusable() {
		t.Error("disabled slider should not be focusable")
	}
}

// --- Layout Tests ---

func TestLayout_RespectsConstraints(t *testing.T) {
	ctx := widget.NewContext()
	constraints := geometry.Tight(geometry.Sz(100, 40))

	s := slider.New()
	size := s.Layout(ctx, constraints)

	if size.Width != 100 {
		t.Errorf("width = %v, want 100 (tight)", size.Width)
	}
	if size.Height != 40 {
		t.Errorf("height = %v, want 40 (tight)", size.Height)
	}
}

func TestLayout_HorizontalWiderThanTall(t *testing.T) {
	ctx := widget.NewContext()
	constraints := geometry.Loose(geometry.Sz(400, 400))

	s := slider.New()
	size := s.Layout(ctx, constraints)

	if size.Width <= size.Height {
		t.Error("horizontal slider should be wider than tall")
	}
}

func TestLayout_VerticalTallerThanWide(t *testing.T) {
	ctx := widget.NewContext()
	constraints := geometry.Loose(geometry.Sz(400, 400))

	s := slider.New(slider.OrientationOpt(slider.Vertical))
	size := s.Layout(ctx, constraints)

	if size.Height <= size.Width {
		t.Error("vertical slider should be taller than wide")
	}
}

// --- Draw Tests ---

func TestDraw_EmptyBounds(t *testing.T) {
	s := slider.New()
	ctx := widget.NewContext()
	canvas := &recordingCanvas{}

	s.Draw(ctx, canvas)

	if canvas.drawCount > 0 {
		t.Error("should not draw anything with empty bounds")
	}
}

func TestDraw_NormalState(t *testing.T) {
	s := slider.New(slider.Min(0), slider.Max(100), slider.Value(50))
	s.SetBounds(geometry.NewRect(0, 0, 200, 30))
	ctx := widget.NewContext()
	canvas := &recordingCanvas{}

	s.Draw(ctx, canvas)

	if canvas.drawCount == 0 {
		t.Error("should draw something with valid bounds")
	}
}

// --- Event Tests ---

func TestEvent_MouseClickSetsValue(t *testing.T) {
	changed := false
	s := slider.New(
		slider.Min(0), slider.Max(100), slider.Value(0),
		slider.OnChange(func(float32) { changed = true }),
	)
	s.SetBounds(geometry.NewRect(0, 0, 200, 30))
	ctx := widget.NewContext()

	press := event.NewMouseEvent(event.MousePress, event.ButtonLeft, event.ButtonStateLeft,
		geometry.Pt(100, 15), geometry.Pt(100, 15), event.ModNone)
	consumed := s.Event(ctx, press)

	if !consumed {
		t.Error("mouse press should be consumed")
	}
	if !changed {
		t.Error("onChange should fire on mouse press")
	}
}

func TestEvent_DisabledIgnoresMouse(t *testing.T) {
	changed := false
	s := slider.New(
		slider.Min(0), slider.Max(100), slider.Value(50),
		slider.OnChange(func(float32) { changed = true }),
		slider.Disabled(true),
	)
	s.SetBounds(geometry.NewRect(0, 0, 200, 30))
	ctx := widget.NewContext()

	press := event.NewMouseEvent(event.MousePress, event.ButtonLeft, event.ButtonStateLeft,
		geometry.Pt(100, 15), geometry.Pt(100, 15), event.ModNone)
	consumed := s.Event(ctx, press)

	if consumed {
		t.Error("disabled slider should not consume mouse events")
	}
	if changed {
		t.Error("disabled slider should not fire onChange")
	}
}

func TestEvent_KeyboardNavigation(t *testing.T) {
	s := slider.New(slider.Min(0), slider.Max(100), slider.Value(50))
	s.SetFocused(true)
	ctx := widget.NewContext()

	keyEvt := event.NewKeyEvent(event.KeyPress, event.KeyRight, 0, event.ModNone)
	consumed := s.Event(ctx, keyEvt)

	if !consumed {
		t.Error("focused slider should consume arrow keys")
	}
}

// --- Signal Binding Tests ---

func TestValueSignal_ReadFromSignal(t *testing.T) {
	sig := state.NewSignal[float32](75)
	s := slider.New(slider.Min(0), slider.Max(100), slider.ValueSignal(sig))
	s.SetBounds(geometry.NewRect(0, 0, 200, 30))
	ctx := widget.NewContext()
	canvas := &recordingCanvas{}

	s.Draw(ctx, canvas)

	// The draw should use signal value; we verify via the painting.
	// Here we just confirm no panics and drawing occurs.
	if canvas.drawCount == 0 {
		t.Error("should draw when value is from signal")
	}
}

func TestDisabledSignal(t *testing.T) {
	sig := state.NewSignal(true)
	s := slider.New(slider.DisabledSignal(sig))

	if s.IsFocusable() {
		t.Error("should not be focusable when DisabledSignal is true")
	}

	sig.Set(false)
	if !s.IsFocusable() {
		t.Error("should be focusable when DisabledSignal is false")
	}
}

// --- Fluent Styling Tests ---

func TestPadding_Chaining(t *testing.T) {
	s := slider.New()
	result := s.Padding(16)

	if result != s {
		t.Error("Padding should return same widget for chaining")
	}
}

// --- recordingCanvas is a minimal mock for external tests ---

type recordingCanvas struct {
	drawCount int
}

func (c *recordingCanvas) Clear(_ widget.Color)                                  {}
func (c *recordingCanvas) DrawRect(_ geometry.Rect, _ widget.Color)              { c.drawCount++ }
func (c *recordingCanvas) StrokeRect(_ geometry.Rect, _ widget.Color, _ float32) { c.drawCount++ }
func (c *recordingCanvas) DrawRoundRect(_ geometry.Rect, _ widget.Color, _ float32) {
	c.drawCount++
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
func (c *recordingCanvas) DrawLine(_, _ geometry.Point, _ widget.Color, _ float32) { c.drawCount++ }
func (c *recordingCanvas) DrawText(_ string, _ geometry.Rect, _ float32, _ widget.Color, _ bool, _ widget.TextAlign) {
	c.drawCount++
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
