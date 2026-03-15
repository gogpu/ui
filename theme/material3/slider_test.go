package material3

import (
	"image"
	"testing"

	"github.com/gogpu/ui/core/slider"
	"github.com/gogpu/ui/geometry"
	"github.com/gogpu/ui/widget"
)

// --- SliderPainter Tests ---

func TestSliderPainter_CompileTimeCheck(t *testing.T) {
	var _ slider.Painter = SliderPainter{}
}

func TestSliderPainter_EmptyBounds(t *testing.T) {
	p := SliderPainter{}
	canvas := &sliderMockCanvas{}

	p.PaintSlider(canvas, slider.PaintState{})

	if canvas.drawCount > 0 {
		t.Error("should not draw anything with empty bounds")
	}
}

func TestSliderPainter_NilTheme_UsesDefaults(t *testing.T) {
	p := SliderPainter{Theme: nil}
	canvas := &sliderMockCanvas{}

	ps := slider.PaintState{
		Value:    50,
		Min:      0,
		Max:      100,
		Progress: 0.5,
		Bounds:   geometry.NewRect(0, 0, 200, 30),
	}

	p.PaintSlider(canvas, ps)

	if canvas.drawCount == 0 {
		t.Error("should draw with nil theme (default colors)")
	}
}

func TestSliderPainter_WithTheme(t *testing.T) {
	theme := New(widget.Hex(0x6750A4)) // M3 purple
	p := SliderPainter{Theme: theme}
	canvas := &sliderMockCanvas{}

	ps := slider.PaintState{
		Value:    75,
		Min:      0,
		Max:      100,
		Progress: 0.75,
		Bounds:   geometry.NewRect(0, 0, 200, 30),
	}

	p.PaintSlider(canvas, ps)

	if canvas.drawCount == 0 {
		t.Error("should draw with theme")
	}
}

func TestSliderPainter_Disabled(t *testing.T) {
	p := SliderPainter{}
	canvas := &sliderMockCanvas{}

	ps := slider.PaintState{
		Value:    50,
		Min:      0,
		Max:      100,
		Progress: 0.5,
		Disabled: true,
		Bounds:   geometry.NewRect(0, 0, 200, 30),
	}

	p.PaintSlider(canvas, ps)

	if canvas.drawCount == 0 {
		t.Error("disabled slider should still draw")
	}
}

func TestSliderPainter_Hovered(t *testing.T) {
	p := SliderPainter{}
	canvas := &sliderMockCanvas{}

	ps := slider.PaintState{
		Value:    50,
		Min:      0,
		Max:      100,
		Progress: 0.5,
		Hovered:  true,
		Bounds:   geometry.NewRect(0, 0, 200, 30),
	}

	p.PaintSlider(canvas, ps)

	if canvas.drawCount == 0 {
		t.Error("hovered slider should draw")
	}
}

func TestSliderPainter_Dragging(t *testing.T) {
	p := SliderPainter{}
	canvas := &sliderMockCanvas{}

	ps := slider.PaintState{
		Value:    50,
		Min:      0,
		Max:      100,
		Progress: 0.5,
		Dragging: true,
		Bounds:   geometry.NewRect(0, 0, 200, 30),
	}

	p.PaintSlider(canvas, ps)

	if canvas.drawCount == 0 {
		t.Error("dragging slider should draw")
	}
}

func TestSliderPainter_Focused(t *testing.T) {
	p := SliderPainter{}
	canvas := &sliderMockCanvas{}

	ps := slider.PaintState{
		Value:    50,
		Min:      0,
		Max:      100,
		Progress: 0.5,
		Focused:  true,
		Bounds:   geometry.NewRect(0, 0, 200, 30),
	}

	p.PaintSlider(canvas, ps)

	if canvas.strokeCircleCount == 0 {
		t.Error("focused slider should draw focus ring (StrokeCircle)")
	}
}

func TestSliderPainter_Focused_NotWhenDisabled(t *testing.T) {
	p := SliderPainter{}
	canvas := &sliderMockCanvas{}

	ps := slider.PaintState{
		Value:    50,
		Min:      0,
		Max:      100,
		Progress: 0.5,
		Focused:  true,
		Disabled: true,
		Bounds:   geometry.NewRect(0, 0, 200, 30),
	}

	p.PaintSlider(canvas, ps)

	if canvas.strokeCircleCount > 0 {
		t.Error("disabled focused slider should not draw focus ring")
	}
}

func TestSliderPainter_Vertical(t *testing.T) {
	p := SliderPainter{}
	canvas := &sliderMockCanvas{}

	ps := slider.PaintState{
		Value:       50,
		Min:         0,
		Max:         100,
		Progress:    0.5,
		Orientation: slider.Vertical,
		Bounds:      geometry.NewRect(0, 0, 30, 200),
	}

	p.PaintSlider(canvas, ps)

	if canvas.drawCount == 0 {
		t.Error("vertical slider should draw")
	}
}

func TestSliderPainter_WithMarks(t *testing.T) {
	p := SliderPainter{}
	canvas := &sliderMockCanvas{}

	ps := slider.PaintState{
		Value:    50,
		Min:      0,
		Max:      100,
		Progress: 0.5,
		Marks:    []slider.Mark{{Value: 25}, {Value: 50}, {Value: 75}},
		Bounds:   geometry.NewRect(0, 0, 200, 30),
	}

	p.PaintSlider(canvas, ps)

	// M3 draws marks as small circles.
	if canvas.drawCircleCount < 4 {
		// At least 3 marks + 1 thumb.
		t.Errorf("should draw at least 4 circles (3 marks + 1 thumb), got %d", canvas.drawCircleCount)
	}
}

func TestSliderPainter_WithColorScheme(t *testing.T) {
	p := SliderPainter{}
	canvas := &sliderMockCanvas{}

	scheme := slider.SliderColorScheme{
		ActiveTrack:   widget.ColorRed,
		InactiveTrack: widget.ColorGray,
		Thumb:         widget.ColorBlue,
		ThumbBorder:   widget.ColorBlack,
		FocusRing:     widget.ColorGreen,
		DisabledTrack: widget.ColorLightGray,
		DisabledThumb: widget.ColorDarkGray,
		MarkColor:     widget.ColorYellow,
	}

	ps := slider.PaintState{
		Value:       50,
		Min:         0,
		Max:         100,
		Progress:    0.5,
		Bounds:      geometry.NewRect(0, 0, 200, 30),
		ColorScheme: scheme,
	}

	p.PaintSlider(canvas, ps)

	if canvas.drawCount == 0 {
		t.Error("should draw with custom color scheme")
	}
}

func TestSliderPainter_ResolveColors_NilTheme(t *testing.T) {
	p := SliderPainter{Theme: nil}
	colors := p.resolveColors()

	if colors == (slider.SliderColorScheme{}) {
		t.Error("should return non-zero default colors")
	}
}

func TestSliderPainter_ResolveColors_WithTheme(t *testing.T) {
	theme := New(widget.Hex(0xFF0000)) // Red seed
	p := SliderPainter{Theme: theme}
	colors := p.resolveColors()

	if colors == (slider.SliderColorScheme{}) {
		t.Error("should return non-zero colors from theme")
	}
	if colors.ActiveTrack == m3DefaultSliderColors.ActiveTrack {
		t.Error("themed colors should differ from default purple")
	}
}

func TestSliderPainter_ZeroProgress(t *testing.T) {
	p := SliderPainter{}
	canvas := &sliderMockCanvas{}

	ps := slider.PaintState{
		Value:    0,
		Min:      0,
		Max:      100,
		Progress: 0,
		Bounds:   geometry.NewRect(0, 0, 200, 30),
	}

	// Should not panic.
	p.PaintSlider(canvas, ps)
}

func TestSliderPainter_FullProgress(t *testing.T) {
	p := SliderPainter{}
	canvas := &sliderMockCanvas{}

	ps := slider.PaintState{
		Value:    100,
		Min:      0,
		Max:      100,
		Progress: 1,
		Bounds:   geometry.NewRect(0, 0, 200, 30),
	}

	// Should not panic.
	p.PaintSlider(canvas, ps)
}

// --- m3ApplySliderState Tests ---

func TestM3ApplySliderState(t *testing.T) {
	base := widget.ColorRed

	normal := m3ApplySliderState(base, false, false)
	if normal != base {
		t.Error("normal state should return base color unchanged")
	}

	hovered := m3ApplySliderState(base, true, false)
	if hovered == base {
		t.Error("hover state should modify color")
	}

	dragging := m3ApplySliderState(base, false, true)
	if dragging == base {
		t.Error("dragging state should modify color")
	}

	// Dragging takes precedence.
	both := m3ApplySliderState(base, true, true)
	if both != dragging {
		t.Error("dragging should take precedence over hover")
	}
}

// --- sliderMockCanvas records basic draw calls ---

type sliderMockCanvas struct {
	drawCount         int
	drawCircleCount   int
	strokeCircleCount int
}

func (c *sliderMockCanvas) Clear(_ widget.Color)                                  {}
func (c *sliderMockCanvas) DrawRect(_ geometry.Rect, _ widget.Color)              { c.drawCount++ }
func (c *sliderMockCanvas) StrokeRect(_ geometry.Rect, _ widget.Color, _ float32) { c.drawCount++ }
func (c *sliderMockCanvas) DrawRoundRect(_ geometry.Rect, _ widget.Color, _ float32) {
	c.drawCount++
}
func (c *sliderMockCanvas) StrokeRoundRect(_ geometry.Rect, _ widget.Color, _ float32, _ float32) {
	c.drawCount++
}
func (c *sliderMockCanvas) DrawCircle(_ geometry.Point, _ float32, _ widget.Color) {
	c.drawCount++
	c.drawCircleCount++
}
func (c *sliderMockCanvas) StrokeCircle(_ geometry.Point, _ float32, _ widget.Color, _ float32) {
	c.drawCount++
	c.strokeCircleCount++
}
func (c *sliderMockCanvas) DrawLine(_, _ geometry.Point, _ widget.Color, _ float32) { c.drawCount++ }
func (c *sliderMockCanvas) DrawText(_ string, _ geometry.Rect, _ float32, _ widget.Color, _ bool, _ widget.TextAlign) {
	c.drawCount++
}

func (c *sliderMockCanvas) MeasureText(text string, fontSize float32, _ bool) float32 {
	return float32(len([]rune(text))) * fontSize * 0.5
}
func (c *sliderMockCanvas) DrawImage(_ image.Image, _ geometry.Point)    { c.drawCount++ }
func (c *sliderMockCanvas) PushClip(_ geometry.Rect)                     {}
func (c *sliderMockCanvas) PushClipRoundRect(_ geometry.Rect, _ float32) {}
func (c *sliderMockCanvas) PopClip()                                     {}
func (c *sliderMockCanvas) PushTransform(_ geometry.Point)               {}
func (c *sliderMockCanvas) PopTransform()                                {}
func (c *sliderMockCanvas) TransformOffset() geometry.Point              { return geometry.Point{} }
