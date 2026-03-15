package material3

import (
	"image"
	"testing"

	"github.com/gogpu/ui/core/dialog"
	"github.com/gogpu/ui/geometry"
	"github.com/gogpu/ui/widget"
)

func TestDialogPainter_EmptyBounds(t *testing.T) {
	p := DialogPainter{}
	canvas := &dialogMockCanvas{}

	p.PaintDialog(canvas, dialog.PaintState{})

	if len(canvas.drawRoundRects) > 0 || len(canvas.drawTexts) > 0 {
		t.Error("should not draw anything with empty bounds")
	}
}

func TestDialogPainter_WithBounds(t *testing.T) {
	p := DialogPainter{}
	canvas := &dialogMockCanvas{}

	p.PaintDialog(canvas, dialog.PaintState{
		Title:  "Confirm",
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

	// Check surface uses default M3 color (no theme).
	bg := canvas.drawRoundRects[0].color
	if bg != m3DefaultDialogColors.Surface {
		t.Errorf("surface color = %v, want M3 default %v", bg, m3DefaultDialogColors.Surface)
	}

	// Check radius is M3 standard.
	if canvas.drawRoundRects[0].radius != m3DialogRadius {
		t.Errorf("radius = %v, want %v", canvas.drawRoundRects[0].radius, m3DialogRadius)
	}
}

func TestDialogPainter_WithTheme(t *testing.T) {
	theme := New(widget.Hex(0x6750A4))
	p := DialogPainter{Theme: theme}
	canvas := &dialogMockCanvas{}

	p.PaintDialog(canvas, dialog.PaintState{
		Title:  "Themed",
		Bounds: geometry.NewRect(100, 100, 500, 400),
	})

	if len(canvas.drawRoundRects) == 0 {
		t.Fatal("should draw surface")
	}

	// With a theme, surface should use SurfaceContainerHigh color.
	expected := theme.Colors.SurfaceContainerHigh
	got := canvas.drawRoundRects[0].color
	if got != expected {
		t.Errorf("surface color = %v, want %v", got, expected)
	}
}

func TestDialogPainter_CustomColorScheme(t *testing.T) {
	p := DialogPainter{}
	canvas := &dialogMockCanvas{}

	customScheme := dialog.DialogColorScheme{
		Surface:  widget.ColorRed,
		Title:    widget.ColorBlue,
		ActionFg: widget.ColorGreen,
	}

	p.PaintDialog(canvas, dialog.PaintState{
		Title:       "Custom",
		Bounds:      geometry.NewRect(100, 100, 500, 400),
		ColorScheme: customScheme,
	})

	if len(canvas.drawRoundRects) == 0 {
		t.Fatal("should draw surface")
	}
	if canvas.drawRoundRects[0].color != widget.ColorRed {
		t.Errorf("surface color = %v, want Red", canvas.drawRoundRects[0].color)
	}
}

func TestDialogPainter_FocusRing(t *testing.T) {
	p := DialogPainter{}
	canvas := &dialogMockCanvas{}

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

func TestDialogPainter_NoTitle(t *testing.T) {
	p := DialogPainter{}
	canvas := &dialogMockCanvas{}

	p.PaintDialog(canvas, dialog.PaintState{
		Bounds: geometry.NewRect(100, 100, 500, 400),
	})

	// Should still draw surface but no title text.
	for _, dt := range canvas.drawTexts {
		if dt.fontSize == m3DialogTitleFontSize {
			t.Error("should not draw title text when title is empty")
		}
	}
}

func TestDialogPainter_ResolveColors_NilTheme(t *testing.T) {
	p := DialogPainter{}
	colors := p.resolveColors()

	if colors != m3DefaultDialogColors {
		t.Error("nil theme should return default M3 dialog colors")
	}
}

func TestDialogPainter_ResolveColors_WithTheme(t *testing.T) {
	theme := New(widget.Hex(0x6750A4))
	p := DialogPainter{Theme: theme}
	colors := p.resolveColors()

	if colors.Surface != theme.Colors.SurfaceContainerHigh {
		t.Errorf("Surface = %v, want %v", colors.Surface, theme.Colors.SurfaceContainerHigh)
	}
	if colors.Title != theme.Colors.OnSurface {
		t.Errorf("Title = %v, want %v", colors.Title, theme.Colors.OnSurface)
	}
	if colors.ActionFg != theme.Colors.Primary {
		t.Errorf("ActionFg = %v, want %v", colors.ActionFg, theme.Colors.Primary)
	}
}

func TestDialogPainter_ImplementsInterface(t *testing.T) {
	var _ dialog.Painter = DialogPainter{}
}

// --- dialogMockCanvas records canvas calls ---

type dialogMockCanvas struct {
	drawRoundRects   []dialogDrawRoundRectCall
	strokeRoundRects []dialogStrokeRoundRectCall
	drawTexts        []dialogDrawTextCall
}

type dialogDrawRoundRectCall struct {
	r      geometry.Rect
	color  widget.Color
	radius float32
}

type dialogStrokeRoundRectCall struct {
	r           geometry.Rect
	color       widget.Color
	radius      float32
	strokeWidth float32
}

type dialogDrawTextCall struct {
	text     string
	bounds   geometry.Rect
	fontSize float32
	color    widget.Color
	bold     bool
	align    widget.TextAlign
}

func (c *dialogMockCanvas) Clear(_ widget.Color)                                  {}
func (c *dialogMockCanvas) DrawRect(_ geometry.Rect, _ widget.Color)              {}
func (c *dialogMockCanvas) StrokeRect(_ geometry.Rect, _ widget.Color, _ float32) {}

func (c *dialogMockCanvas) DrawRoundRect(r geometry.Rect, color widget.Color, radius float32) {
	c.drawRoundRects = append(c.drawRoundRects, dialogDrawRoundRectCall{r: r, color: color, radius: radius})
}

func (c *dialogMockCanvas) StrokeRoundRect(r geometry.Rect, color widget.Color, radius float32, strokeWidth float32) {
	c.strokeRoundRects = append(c.strokeRoundRects, dialogStrokeRoundRectCall{r: r, color: color, radius: radius, strokeWidth: strokeWidth})
}

func (c *dialogMockCanvas) DrawCircle(_ geometry.Point, _ float32, _ widget.Color)              {}
func (c *dialogMockCanvas) StrokeCircle(_ geometry.Point, _ float32, _ widget.Color, _ float32) {}
func (c *dialogMockCanvas) DrawLine(_, _ geometry.Point, _ widget.Color, _ float32)             {}

func (c *dialogMockCanvas) DrawText(text string, bounds geometry.Rect, fontSize float32, color widget.Color, bold bool, align widget.TextAlign) {
	c.drawTexts = append(c.drawTexts, dialogDrawTextCall{text: text, bounds: bounds, fontSize: fontSize, color: color, bold: bold, align: align})
}

func (c *dialogMockCanvas) MeasureText(text string, fontSize float32, _ bool) float32 {
	return float32(len([]rune(text))) * fontSize * 0.5
}

func (c *dialogMockCanvas) DrawImage(_ image.Image, _ geometry.Point)    {}
func (c *dialogMockCanvas) PushClip(_ geometry.Rect)                     {}
func (c *dialogMockCanvas) PushClipRoundRect(_ geometry.Rect, _ float32) {}
func (c *dialogMockCanvas) PopClip()                                     {}
func (c *dialogMockCanvas) PushTransform(_ geometry.Point)               {}
func (c *dialogMockCanvas) PopTransform()                                {}
func (c *dialogMockCanvas) TransformOffset() geometry.Point              { return geometry.Point{} }
