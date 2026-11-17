package fluent_test

import (
	"testing"
	"time"

	"github.com/gogpu/ui/core/toggleswitch"
	"github.com/gogpu/ui/geometry"
	"github.com/gogpu/ui/theme/fluent"
	"github.com/gogpu/ui/widget"
)

// switchState returns the paint state of a switch at rest whose control
// bounds start at (10, 10).
func switchState(checked bool) toggleswitch.PaintState {
	p := fluent.SwitchPainter{}
	size := p.SwitchSize()
	control := geometry.NewRect(10, 10, size.Width, size.Height)
	labelX := control.Max.X + p.SwitchLabelGap()
	s := toggleswitch.PaintState{
		Label:         "Wi-Fi",
		Bounds:        geometry.NewRect(10, 10, 200, size.Height),
		ControlBounds: control,
		LabelBounds:   geometry.NewRect(labelX, 10, 210-labelX, size.Height),
	}
	if checked {
		s.Checked, s.Position, s.Progress, s.TurningOn = true, 1, 1, true
	}
	return s
}

// switchTrack is the track rectangle for switchState.
var switchTrack = geometry.NewRect(14, 14, 40, 20)

func switchColorNear(a, b widget.Color) bool {
	const eps = 1e-4
	d := func(x, y float32) bool { return x-y < eps && y-x < eps }
	return d(a.R, b.R) && d(a.G, b.G) && d(a.B, b.B) && d(a.A, b.A)
}

// switchParts holds the calls a painted switch is made of.
type switchParts struct {
	canvas *recordCanvas
	fill   drawCall
	border drawCall
	knob   drawCall
}

// paintSwitch paints s and returns its track fill, border, and knob.
func paintSwitch(t *testing.T, p fluent.SwitchPainter, s toggleswitch.PaintState) switchParts {
	t.Helper()
	c := &recordCanvas{}
	p.PaintSwitch(c, s)
	rects := c.methodCalls(methodDrawRoundRect)
	strokes := c.methodCalls(methodStrokeRoundRect)
	if len(rects) != 2 || len(strokes) == 0 {
		t.Fatalf("want track fill, border, and knob; got %d fills and %d strokes", len(rects), len(strokes))
	}
	return switchParts{canvas: c, fill: rects[0], border: strokes[0], knob: rects[1]}
}

func TestSwitchPainter_EmptyControlBounds(t *testing.T) {
	c := &recordCanvas{}
	fluent.SwitchPainter{}.PaintSwitch(c, toggleswitch.PaintState{Label: "Wi-Fi"})
	if len(c.calls) != 0 {
		t.Errorf("empty control bounds should draw nothing, got %d calls", len(c.calls))
	}
}

func TestSwitchPainter_OffLight(t *testing.T) {
	sw := paintSwitch(t, fluent.SwitchPainter{Theme: fluent.NewTheme()}, switchState(false))
	fill, border, knob := sw.fill, sw.border, sw.knob

	if fill.bounds != switchTrack || fill.radius != 10 || !switchColorNear(fill.color, widget.HexA(0x00000006)) {
		t.Errorf("off track = %v r%v %v", fill.bounds, fill.radius, fill.color)
	}
	if border.strokeWidth != 1 || !switchColorNear(border.color, widget.HexA(0x00000072)) {
		t.Errorf("off border = %v w%v, want #72000000 w1", border.color, border.strokeWidth)
	}
	if knob.bounds != geometry.NewRect(18, 18, 12, 12) || !switchColorNear(knob.color, widget.HexA(0x0000009E)) {
		t.Errorf("off knob = %v %v, want 12×12 at x=18 in #9E000000", knob.bounds, knob.color)
	}
	texts := sw.canvas.methodCalls(methodDrawText)
	if len(texts) != 1 || texts[0].text != "Wi-Fi" || texts[0].fontSize != 14 ||
		texts[0].bounds != switchState(false).LabelBounds {
		t.Errorf("label calls = %+v", texts)
	}
}

func TestSwitchPainter_OnLight(t *testing.T) {
	th := fluent.NewTheme()
	sw := paintSwitch(t, fluent.SwitchPainter{Theme: th}, switchState(true))
	fill, border, knob := sw.fill, sw.border, sw.knob

	if !switchColorNear(fill.color, th.Colors.Accent) || !switchColorNear(border.color, th.Colors.Accent) {
		t.Errorf("on track = %v / %v, want Accent", fill.color, border.color)
	}
	if knob.bounds != geometry.NewRect(38, 18, 12, 12) || !switchColorNear(knob.color, th.Colors.OnAccent) {
		t.Errorf("on knob = %v %v, want 12×12 at x=38 in OnAccent", knob.bounds, knob.color)
	}
}

func TestSwitchPainter_HoverAndPress(t *testing.T) {
	th := fluent.NewTheme()

	s := switchState(false)
	s.Hovered, s.HoverAmount = true, 1
	sw := paintSwitch(t, fluent.SwitchPainter{Theme: th}, s)
	fill, knob := sw.fill, sw.knob
	if knob.bounds != geometry.NewRect(17, 17, 14, 14) {
		t.Errorf("hovered off knob = %v, want 14×14 at (17, 17)", knob.bounds)
	}
	if !switchColorNear(fill.color, widget.HexA(0x0000000F)) {
		t.Errorf("hovered off track = %v, want #0F000000", fill.color)
	}

	s = switchState(true)
	s.Hovered, s.HoverAmount, s.Pressed, s.PressAmount = true, 1, true, 1
	sw = paintSwitch(t, fluent.SwitchPainter{Theme: th}, s)
	fill, knob = sw.fill, sw.knob
	if knob.bounds != geometry.NewRect(34, 17, 17, 14) || knob.radius != 7 {
		t.Errorf("pressed on knob = %v r%v, want 17×14 ending 3 px from the edge", knob.bounds, knob.radius)
	}
	if !switchColorNear(fill.color, th.Colors.Accent.WithAlpha(0.8)) {
		t.Errorf("pressed on track = %v, want Accent at 0.8", fill.color)
	}
}

func TestSwitchPainter_Disabled(t *testing.T) {
	s := switchState(true)
	s.Disabled, s.Focused, s.HoverAmount = true, true, 1
	sw := paintSwitch(t, fluent.SwitchPainter{Theme: fluent.NewTheme()}, s)
	c, fill, knob := sw.canvas, sw.fill, sw.knob

	if !switchColorNear(fill.color, widget.HexA(0x00000037)) || knob.color != widget.ColorWhite {
		t.Errorf("disabled on = track %v knob %v", fill.color, knob.color)
	}
	if knob.bounds.Width() != 12 {
		t.Errorf("disabled knob should not react to hover, got %v", knob.bounds)
	}
	if n := len(c.methodCalls(methodStrokeRoundRect)); n != 1 {
		t.Errorf("disabled switch should draw only its border, got %d strokes", n)
	}
	if label := c.methodCalls(methodDrawText)[0]; !switchColorNear(label.color, widget.HexA(0x0000005C)) {
		t.Errorf("disabled label = %v, want #5C000000", label.color)
	}
}

func TestSwitchPainter_Dark(t *testing.T) {
	th := fluent.NewDarkTheme()
	off := paintSwitch(t, fluent.SwitchPainter{Theme: th}, switchState(false))
	if !switchColorNear(off.border.color, widget.HexA(0xFFFFFF8B)) || !switchColorNear(off.knob.color, widget.HexA(0xFFFFFFC5)) {
		t.Errorf("dark off = border %v knob %v", off.border.color, off.knob.color)
	}
	on := paintSwitch(t, fluent.SwitchPainter{Theme: th}, switchState(true))
	if !switchColorNear(on.knob.color, th.Colors.OnAccent) {
		t.Errorf("dark on knob = %v, want OnAccent", on.knob.color)
	}
}

func TestSwitchPainter_FocusRing(t *testing.T) {
	th := fluent.NewTheme()
	s := switchState(false)
	s.Focused = true
	strokes := paintSwitch(t, fluent.SwitchPainter{Theme: th}, s).canvas.methodCalls(methodStrokeRoundRect)
	if len(strokes) != 2 {
		t.Fatalf("focused switch should draw border and focus ring, got %d strokes", len(strokes))
	}
	ring := strokes[1]
	if ring.bounds != switchTrack.Expand(3) || ring.radius != 13 || ring.strokeWidth != 2 ||
		!switchColorNear(ring.color, th.Colors.StrokeFocus) {
		t.Errorf("focus ring = %+v", ring)
	}
}

func TestSwitchPainter_NilTheme(t *testing.T) {
	fill := paintSwitch(t, fluent.SwitchPainter{}, switchState(true)).fill
	if !switchColorNear(fill.color, fluent.DefaultAccentColor) {
		t.Errorf("nil theme on track = %v, want DefaultAccentColor", fill.color)
	}
}

func TestSwitchPainter_MetricsAndMotion(t *testing.T) {
	p := fluent.SwitchPainter{}
	if got := p.SwitchSize(); got != geometry.Sz(48, 28) {
		t.Errorf("SwitchSize = %v, want 48×28", got)
	}
	if p.SwitchThumbTravel() != 20 || p.SwitchLabelGap() != 8 || p.SwitchFontSize() != 14 {
		t.Errorf("travel/gap/font = %v/%v/%v, want 20/8/14",
			p.SwitchThumbTravel(), p.SwitchLabelGap(), p.SwitchFontSize())
	}
	m := p.SwitchMotion()
	if m.Toggle != 167*time.Millisecond || m.Hover != 83*time.Millisecond || m.Press != 83*time.Millisecond {
		t.Errorf("motion = %+v", m)
	}
	if m.ToggleEasing(0.25) < 0.6 {
		t.Errorf("toggle easing should be fast out, got %v at 0.25", m.ToggleEasing(0.25))
	}
}
