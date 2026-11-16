package cupertino_test

import (
	"math"
	"slices"
	"testing"

	"github.com/gogpu/ui/core/checkbox"
	"github.com/gogpu/ui/geometry"
	"github.com/gogpu/ui/theme/cupertino"
	"github.com/gogpu/ui/widget"
)

// colorNear reports whether two colors match within 1/255 per channel.
func colorNear(a, b widget.Color) bool {
	const eps = 1.0 / 255
	return math.Abs(float64(a.R-b.R)) <= eps && math.Abs(float64(a.G-b.G)) <= eps &&
		math.Abs(float64(a.B-b.B)) <= eps && math.Abs(float64(a.A-b.A)) <= eps
}

// paintCheckbox paints state with p and returns the recorded calls.
func paintCheckbox(p cupertino.CheckboxPainter, state checkbox.PaintState) *recordCanvas {
	canvas := &recordCanvas{}
	if state.Bounds.IsEmpty() {
		state.Bounds = testBounds()
	}
	p.PaintCheckbox(canvas, state)
	return canvas
}

// wantCheckboxBox is the 14pt box testBounds() centers vertically at its left edge.
var wantCheckboxBox = geometry.NewRect(10, 23, 14, 14)

func TestCheckboxPainterImplementsInterface(t *testing.T) {
	var _ checkbox.Painter = cupertino.CheckboxPainter{}
	var _ checkbox.LayoutMetrics = cupertino.CheckboxPainter{}
}

func TestPaintCheckboxEmptyBounds(t *testing.T) {
	canvas := &recordCanvas{}
	cupertino.CheckboxPainter{}.PaintCheckbox(canvas, checkbox.PaintState{Bounds: geometry.Rect{}})

	if len(canvas.calls) != 0 {
		t.Errorf("empty bounds should produce no draw calls, got %d", len(canvas.calls))
	}
}

func TestCheckboxLayoutMetrics(t *testing.T) {
	p := cupertino.CheckboxPainter{}
	if got := p.CheckboxBoxSize(); got != 14 {
		t.Errorf("CheckboxBoxSize = %v, want 14", got)
	}
	if got := p.CheckboxLabelGap(); got != 6 {
		t.Errorf("CheckboxLabelGap = %v, want 6", got)
	}
	if got := p.CheckboxFontSize(); got != 13 {
		t.Errorf("CheckboxFontSize = %v, want 13", got)
	}
}

func TestCheckboxWidgetLayoutUsesMacOSBox(t *testing.T) {
	cb := checkbox.New(checkbox.PainterOpt(cupertino.CheckboxPainter{}))
	size := cb.Layout(widget.NewContext(), geometry.Loose(geometry.Sz(400, 400)))

	// 14pt box plus the widget's 4pt padding on each side; height is the
	// checkbox widget's 24pt minimum.
	if size.Width != 22 || size.Height != 24 {
		t.Errorf("layout size = %vx%v, want 22x24", size.Width, size.Height)
	}
}

func TestPaintCheckboxUnchecked(t *testing.T) {
	canvas := paintCheckbox(cupertino.CheckboxPainter{}, checkbox.PaintState{})

	fills := canvas.methodCalls(methodDrawRoundRect)
	if len(fills) != 1 {
		t.Fatalf("unchecked box should draw 1 fill, got %d", len(fills))
	}
	if fills[0].bounds != wantCheckboxBox || fills[0].radius != 4 {
		t.Errorf("box = %v r%v, want %v r4", fills[0].bounds, fills[0].radius, wantCheckboxBox)
	}
	if !colorNear(fills[0].color, widget.ColorWhite) {
		t.Errorf("unchecked fill = %v, want white", fills[0].color)
	}

	strokes := canvas.methodCalls(methodStrokeRoundRect)
	if len(strokes) != 1 {
		t.Fatalf("unchecked box should draw 1 outline, got %d", len(strokes))
	}
	// The 1pt outline is stroked inside the box, so it does not grow it.
	if want := geometry.NewRect(10.5, 23.5, 13, 13); strokes[0].bounds != want || strokes[0].strokeWidth != 1 {
		t.Errorf("outline = %v w%v, want %v w1", strokes[0].bounds, strokes[0].strokeWidth, want)
	}
	if !colorNear(strokes[0].color, widget.Hex(0xD1D1D6)) {
		t.Errorf("outline color = %v, want #D1D1D6", strokes[0].color)
	}

	if n := len(canvas.methodCalls(methodDrawLine)); n != 0 {
		t.Errorf("unchecked box should draw no mark, got %d lines", n)
	}
}

func TestPaintCheckboxChecked(t *testing.T) {
	canvas := paintCheckbox(cupertino.CheckboxPainter{}, checkbox.PaintState{
		Label:   "Remember me",
		Checked: true,
	})

	fills := canvas.methodCalls(methodDrawRoundRect)
	if len(fills) != 1 || !colorNear(fills[0].color, widget.Hex(0x007AFF)) {
		t.Fatalf("checked box should be one system blue fill, got %+v", fills)
	}
	if n := len(canvas.methodCalls(methodStrokeRoundRect)); n != 0 {
		t.Errorf("checked box has no outline, got %d strokes", n)
	}

	// Check mark: two 2pt segments with round caps and join.
	lines := canvas.methodCalls(methodDrawLine)
	if len(lines) != 2 {
		t.Fatalf("check mark should be 2 segments, got %d", len(lines))
	}
	for _, l := range lines {
		if l.strokeWidth != 2 || !colorNear(l.color, widget.ColorWhite) {
			t.Errorf("check segment = %v w%v, want white w2", l.color, l.strokeWidth)
		}
	}
	if dots := canvas.methodCalls(methodDrawCircle); len(dots) != 3 || dots[0].radius != 1 {
		t.Errorf("check mark should round its 3 vertices with r1 dots, got %+v", dots)
	}

	texts := canvas.methodCalls(methodDrawText)
	if len(texts) != 1 {
		t.Fatalf("should draw 1 label, got %d", len(texts))
	}
	if texts[0].text != "Remember me" || texts[0].fontSize != 13 || texts[0].bounds.Min.X != 30 {
		t.Errorf("label = %q %vpt at x=%v, want \"Remember me\" 13pt at x=30",
			texts[0].text, texts[0].fontSize, texts[0].bounds.Min.X)
	}
	if !colorNear(texts[0].color, widget.ColorBlack) {
		t.Errorf("label color = %v, want black", texts[0].color)
	}
}

func TestPaintCheckboxIndeterminate(t *testing.T) {
	canvas := paintCheckbox(cupertino.CheckboxPainter{}, checkbox.PaintState{Indeterminate: true})

	fills := canvas.methodCalls(methodDrawRoundRect)
	if len(fills) != 1 || !colorNear(fills[0].color, widget.Hex(0x007AFF)) {
		t.Fatalf("mixed box should be one system blue fill, got %+v", fills)
	}
	if n := len(canvas.methodCalls(methodDrawLine)); n != 1 {
		t.Errorf("mixed state should draw 1 dash, got %d lines", n)
	}
}

func TestPaintCheckboxHoverHasNoEffect(t *testing.T) {
	for _, checked := range []bool{false, true} {
		rest := paintCheckbox(cupertino.CheckboxPainter{}, checkbox.PaintState{Checked: checked})
		hover := paintCheckbox(cupertino.CheckboxPainter{}, checkbox.PaintState{Checked: checked, Hovered: true})
		if !slices.Equal(rest.calls, hover.calls) {
			t.Errorf("checked=%v: hover changed the drawing:\nrest  %+v\nhover %+v", checked, rest.calls, hover.calls)
		}
	}
}

func TestPaintCheckboxPressedShadesBox(t *testing.T) {
	canvas := paintCheckbox(cupertino.CheckboxPainter{}, checkbox.PaintState{Pressed: true})

	fills := canvas.methodCalls(methodDrawRoundRect)
	if len(fills) != 2 {
		t.Fatalf("pressed box should draw fill and shade, got %d fills", len(fills))
	}
	if fills[1].bounds != wantCheckboxBox || !colorNear(fills[1].color, widget.ColorBlack.WithAlpha(0.15)) {
		t.Errorf("press shade = %v %v, want black@0.15 over the box", fills[1].bounds, fills[1].color)
	}
}

func TestPaintCheckboxDisabled(t *testing.T) {
	canvas := paintCheckbox(cupertino.CheckboxPainter{}, checkbox.PaintState{
		Label:    "Disabled",
		Checked:  true,
		Disabled: true,
		Pressed:  true,
		Focused:  true,
	})

	fills := canvas.methodCalls(methodDrawRoundRect)
	if len(fills) != 1 || !colorNear(fills[0].color, widget.ColorWhite.WithAlpha(0.5)) {
		t.Fatalf("disabled box should be one white@0.5 fill and no press shade, got %+v", fills)
	}
	strokes := canvas.methodCalls(methodStrokeRoundRect)
	if len(strokes) != 1 || strokes[0].strokeWidth != 1 {
		t.Fatalf("disabled box should draw only its outline (no focus ring), got %+v", strokes)
	}

	lines := canvas.methodCalls(methodDrawLine)
	if len(lines) != 2 || !colorNear(lines[0].color, widget.ColorBlack.WithAlpha(0.25)) {
		t.Errorf("disabled check mark should be black@0.25, got %+v", lines)
	}
	// Translucent marks skip the round-cap dots that would double their alpha.
	if n := len(canvas.methodCalls(methodDrawCircle)); n != 0 {
		t.Errorf("translucent mark should draw no dots, got %d", n)
	}

	texts := canvas.methodCalls(methodDrawText)
	if len(texts) != 1 || !colorNear(texts[0].color, cupertino.NewTheme().Colors.TertiaryLabel) {
		t.Errorf("disabled label should use TertiaryLabel, got %+v", texts)
	}
}

func TestPaintCheckboxFocusRing(t *testing.T) {
	canvas := paintCheckbox(cupertino.CheckboxPainter{}, checkbox.PaintState{Focused: true})

	strokes := canvas.methodCalls(methodStrokeRoundRect)
	if len(strokes) != 2 {
		t.Fatalf("focused box should draw outline and focus ring, got %d strokes", len(strokes))
	}
	ring := strokes[1]
	if want := geometry.NewRect(9, 22, 16, 16); ring.bounds != want || ring.radius != 5 || ring.strokeWidth != 3.5 {
		t.Errorf("focus ring = %v r%v w%v, want %v r5 w3.5", ring.bounds, ring.radius, ring.strokeWidth, want)
	}
	// System blue's hue (211°) at HSL saturation 0.835 and lightness 0.69,
	// 80% opaque: rgb(0.431, 0.679, 0.949).
	if want := widget.RGBA(0.431, 0.679, 0.949, 0.8); !colorNear(ring.color, want) {
		t.Errorf("focus ring color = %v, want %v", ring.color, want)
	}
}

func TestPaintCheckboxDarkTheme(t *testing.T) {
	p := cupertino.CheckboxPainter{Theme: cupertino.NewDarkTheme()}

	off := paintCheckbox(p, checkbox.PaintState{Pressed: true})
	fills := off.methodCalls(methodDrawRoundRect)
	if len(fills) != 2 {
		t.Fatalf("pressed dark box should draw fill and shade, got %d fills", len(fills))
	}
	if !colorNear(fills[0].color, widget.ColorWhite.WithAlpha(0.21)) {
		t.Errorf("dark unchecked fill = %v, want white@0.21", fills[0].color)
	}
	if !colorNear(fills[1].color, widget.ColorWhite.WithAlpha(0.15)) {
		t.Errorf("dark press shade = %v, want white@0.15", fills[1].color)
	}
	strokes := off.methodCalls(methodStrokeRoundRect)
	if len(strokes) != 1 || !colorNear(strokes[0].color, widget.RGBA(0.5, 0.5, 0.5, 0.2)) {
		t.Errorf("dark outline should be gray@0.2, got %+v", strokes)
	}

	on := paintCheckbox(p, checkbox.PaintState{Checked: true})
	lines := on.methodCalls(methodDrawLine)
	if len(lines) != 2 || !colorNear(lines[0].color, widget.Hex(0xDEE8F8)) {
		t.Errorf("dark check mark should be #DEE8F8, got %+v", lines)
	}
}

func TestPaintCheckboxThemeAccent(t *testing.T) {
	green := widget.Hex(0x34C759)
	p := cupertino.CheckboxPainter{Theme: cupertino.NewTheme(cupertino.WithAccentColor(green))}

	fills := paintCheckbox(p, checkbox.PaintState{Checked: true}).methodCalls(methodDrawRoundRect)
	if len(fills) != 1 || !colorNear(fills[0].color, green) {
		t.Errorf("checked fill should use the theme accent, got %+v", fills)
	}
}

func TestPaintCheckboxOverrides(t *testing.T) {
	red, orange := widget.Hex(0xFF3B30), widget.Hex(0xFF9500)

	scheme := paintCheckbox(cupertino.CheckboxPainter{}, checkbox.PaintState{
		Checked:     true,
		ColorScheme: checkbox.CheckboxColorScheme{CheckedBg: red},
	})
	fills := scheme.methodCalls(methodDrawRoundRect)
	if len(fills) != 1 || !colorNear(fills[0].color, red) {
		t.Errorf("ColorScheme.CheckedBg should fill the box, got %+v", fills)
	}
	if lines := scheme.methodCalls(methodDrawLine); !colorNear(lines[0].color, widget.ColorWhite) {
		t.Errorf("unset ColorScheme fields keep the theme colors, mark = %v", lines[0].color)
	}

	bg := paintCheckbox(cupertino.CheckboxPainter{}, checkbox.PaintState{
		Checked:    true,
		Background: &orange,
	})
	if fills := bg.methodCalls(methodDrawRoundRect); len(fills) != 1 || !colorNear(fills[0].color, orange) {
		t.Errorf("Background should fill the checked box, got %+v", fills)
	}
}
