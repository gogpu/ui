package devtools

import (
	"image"
	"testing"

	"github.com/gogpu/ui/core/titlebar"
	"github.com/gogpu/ui/geometry"
	"github.com/gogpu/ui/widget"
)

// --- TitleBarPainter Tests ---

func TestTitleBarPainter_DrawBackground(t *testing.T) {
	p := TitleBarPainter{Theme: NewDarkTheme()}
	canvas := &tbMockCanvas{}
	bounds := geometry.NewRect(0, 0, 800, 40)

	p.DrawBackground(canvas, bounds, titlebar.BackgroundState{Focused: true})

	if len(canvas.drawRects) < 2 {
		t.Error("should draw background + bottom border (at least 2 rects)")
	}
}

func TestTitleBarPainter_DrawBackground_EmptyBounds(t *testing.T) {
	p := TitleBarPainter{Theme: NewDarkTheme()}
	canvas := &tbMockCanvas{}

	p.DrawBackground(canvas, geometry.Rect{}, titlebar.BackgroundState{})

	if len(canvas.drawRects) > 0 {
		t.Error("should not draw with empty bounds")
	}
}

func TestTitleBarPainter_DrawBackground_NilTheme(t *testing.T) {
	p := TitleBarPainter{Theme: nil}
	canvas := &tbMockCanvas{}
	bounds := geometry.NewRect(0, 0, 800, 40)

	p.DrawBackground(canvas, bounds, titlebar.BackgroundState{})

	if len(canvas.drawRects) < 2 {
		t.Error("nil theme should use default dark colors and draw background + border")
	}
}

func TestTitleBarPainter_DrawControlButton_EmptyBounds(t *testing.T) {
	p := TitleBarPainter{Theme: NewDarkTheme()}
	canvas := &tbMockCanvas{}

	p.DrawControlButton(canvas, geometry.Rect{}, titlebar.ControlClose, titlebar.ControlState{})

	if len(canvas.drawRects) > 0 || len(canvas.drawLines) > 0 {
		t.Error("should not draw with empty bounds")
	}
}

func TestTitleBarPainter_DrawControlButton_Minimize(t *testing.T) {
	p := TitleBarPainter{Theme: NewDarkTheme()}
	canvas := &tbMockCanvas{}
	bounds := geometry.NewRect(0, 0, 46, 40)

	p.DrawControlButton(canvas, bounds, titlebar.ControlMinimize, titlebar.ControlState{})

	if len(canvas.drawLines) == 0 {
		t.Error("minimize should draw a line")
	}
}

func TestTitleBarPainter_DrawControlButton_Maximize(t *testing.T) {
	p := TitleBarPainter{Theme: NewDarkTheme()}
	canvas := &tbMockCanvas{}
	bounds := geometry.NewRect(0, 0, 46, 40)

	p.DrawControlButton(canvas, bounds, titlebar.ControlMaximize, titlebar.ControlState{})

	if len(canvas.strokeRects) == 0 {
		t.Error("maximize should draw a stroked rect")
	}
}

func TestTitleBarPainter_DrawControlButton_Restore(t *testing.T) {
	p := TitleBarPainter{Theme: NewDarkTheme()}
	canvas := &tbMockCanvas{}
	bounds := geometry.NewRect(0, 0, 46, 40)

	p.DrawControlButton(canvas, bounds, titlebar.ControlRestore, titlebar.ControlState{})

	if len(canvas.strokeRects) < 2 {
		t.Errorf("restore should draw 2 stroked rects, got %d", len(canvas.strokeRects))
	}
}

func TestTitleBarPainter_DrawControlButton_Close(t *testing.T) {
	p := TitleBarPainter{Theme: NewDarkTheme()}
	canvas := &tbMockCanvas{}
	bounds := geometry.NewRect(0, 0, 46, 40)

	p.DrawControlButton(canvas, bounds, titlebar.ControlClose, titlebar.ControlState{})

	if len(canvas.drawLines) != 2 {
		t.Errorf("close should draw 2 lines (X), got %d", len(canvas.drawLines))
	}
}

func TestTitleBarPainter_DrawControlButton_CloseHover(t *testing.T) {
	p := TitleBarPainter{Theme: NewDarkTheme()}
	canvas := &tbMockCanvas{}
	bounds := geometry.NewRect(0, 0, 46, 40)

	p.DrawControlButton(canvas, bounds, titlebar.ControlClose, titlebar.ControlState{Hovered: true})

	// Should draw red background.
	if len(canvas.drawRects) == 0 {
		t.Error("close hover should draw red background")
	}
	// Background color should be the close hover red.
	foundRed := false
	for _, call := range canvas.drawRects {
		if call.color == dtCloseHoverRed {
			foundRed = true
			break
		}
	}
	if !foundRed {
		t.Error("close hover background should be red (#C42B1C)")
	}
}

func TestTitleBarPainter_DrawControlButton_ClosePressed(t *testing.T) {
	p := TitleBarPainter{Theme: NewDarkTheme()}
	canvas := &tbMockCanvas{}
	bounds := geometry.NewRect(0, 0, 46, 40)

	p.DrawControlButton(canvas, bounds, titlebar.ControlClose, titlebar.ControlState{Pressed: true})

	// Should draw darker red background.
	if len(canvas.drawRects) == 0 {
		t.Error("close pressed should draw dark red background")
	}
}

func TestTitleBarPainter_DrawControlButton_MinimizeHover(t *testing.T) {
	p := TitleBarPainter{Theme: NewDarkTheme()}
	canvas := &tbMockCanvas{}
	bounds := geometry.NewRect(0, 0, 46, 40)

	p.DrawControlButton(canvas, bounds, titlebar.ControlMinimize, titlebar.ControlState{Hovered: true})

	if len(canvas.drawRects) == 0 {
		t.Error("minimize hover should draw hover background")
	}
}

func TestTitleBarPainter_DrawControlButton_MaximizePressed(t *testing.T) {
	p := TitleBarPainter{Theme: NewDarkTheme()}
	canvas := &tbMockCanvas{}
	bounds := geometry.NewRect(0, 0, 46, 40)

	p.DrawControlButton(canvas, bounds, titlebar.ControlMaximize, titlebar.ControlState{Pressed: true})

	if len(canvas.drawRects) == 0 {
		t.Error("maximize pressed should draw pressed background")
	}
}

func TestTitleBarPainter_LightTheme(t *testing.T) {
	p := TitleBarPainter{Theme: NewTheme()}
	canvas := &tbMockCanvas{}
	bounds := geometry.NewRect(0, 0, 800, 40)

	// Should not panic with light theme.
	p.DrawBackground(canvas, bounds, titlebar.BackgroundState{Focused: true})
	p.DrawControlButton(canvas, bounds, titlebar.ControlClose, titlebar.ControlState{})

	if len(canvas.drawRects) == 0 {
		t.Error("light theme should draw background")
	}
}

func TestTitleBarPainter_resolveColors_NilTheme(t *testing.T) {
	p := TitleBarPainter{Theme: nil}
	colors := p.resolveColors()

	if colors.Background != dtDefaultTitleBarColors.Background {
		t.Errorf("nil theme should use default background")
	}
}

func TestTitleBarPainter_resolveColors_WithTheme(t *testing.T) {
	dt := NewDarkTheme()
	p := TitleBarPainter{Theme: dt}
	colors := p.resolveColors()

	if colors.Background != dt.Colors.HeaderBackground {
		t.Errorf("should use theme HeaderBackground")
	}
}

// --- Mock Canvas for titlebar tests ---

type tbMockCanvas struct {
	drawRects   []tbDrawRectCall
	strokeRects []tbStrokeRectCall
	drawLines   []tbDrawLineCall
}

type tbDrawRectCall struct {
	r     geometry.Rect
	color widget.Color
}

type tbStrokeRectCall struct {
	r           geometry.Rect
	color       widget.Color
	strokeWidth float32
}

type tbDrawLineCall struct {
	from, to    geometry.Point
	color       widget.Color
	strokeWidth float32
}

func (c *tbMockCanvas) Clear(_ widget.Color) {}
func (c *tbMockCanvas) DrawRect(r geometry.Rect, color widget.Color) {
	c.drawRects = append(c.drawRects, tbDrawRectCall{r: r, color: color})
}
func (c *tbMockCanvas) StrokeRect(r geometry.Rect, color widget.Color, sw float32) {
	c.strokeRects = append(c.strokeRects, tbStrokeRectCall{r: r, color: color, strokeWidth: sw})
}
func (c *tbMockCanvas) DrawRoundRect(_ geometry.Rect, _ widget.Color, _ float32)              {}
func (c *tbMockCanvas) StrokeRoundRect(_ geometry.Rect, _ widget.Color, _ float32, _ float32) {}
func (c *tbMockCanvas) DrawCircle(_ geometry.Point, _ float32, _ widget.Color)                {}
func (c *tbMockCanvas) StrokeCircle(_ geometry.Point, _ float32, _ widget.Color, _ float32)   {}
func (c *tbMockCanvas) DrawLine(from, to geometry.Point, color widget.Color, sw float32) {
	c.drawLines = append(c.drawLines, tbDrawLineCall{from: from, to: to, color: color, strokeWidth: sw})
}
func (c *tbMockCanvas) DrawText(_ string, _ geometry.Rect, _ float32, _ widget.Color, _ bool, _ widget.TextAlign) {
}
func (c *tbMockCanvas) MeasureText(_ string, fontSize float32, _ bool) float32 { return fontSize * 5 }
func (c *tbMockCanvas) DrawImage(_ image.Image, _ geometry.Point)              {}
func (c *tbMockCanvas) PushClip(_ geometry.Rect)                               {}
func (c *tbMockCanvas) PushClipRoundRect(_ geometry.Rect, _ float32)           {}
func (c *tbMockCanvas) PopClip()                                               {}
func (c *tbMockCanvas) PushTransform(_ geometry.Point)                         {}
func (c *tbMockCanvas) PopTransform()                                          {}
func (c *tbMockCanvas) TransformOffset() geometry.Point                        { return geometry.Point{} }
