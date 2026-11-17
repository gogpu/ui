package devtools_test

import (
	"testing"
	"time"

	"github.com/gogpu/ui/core/toggleswitch"
	"github.com/gogpu/ui/geometry"
	"github.com/gogpu/ui/theme/devtools"
	"github.com/gogpu/ui/widget"
)

// switchCanvas extends recordCanvas with the circle centers a switch test
// needs.
type switchCanvas struct {
	recordCanvas
	circles []switchCircle
}

type switchCircle struct {
	center geometry.Point
	radius float32
	stroke float32 // 0 for a filled circle
	color  widget.Color
}

func (c *switchCanvas) DrawCircle(center geometry.Point, radius float32, color widget.Color) {
	c.circles = append(c.circles, switchCircle{center: center, radius: radius, color: color})
	c.recordCanvas.DrawCircle(center, radius, color)
}

func (c *switchCanvas) StrokeCircle(center geometry.Point, radius float32, color widget.Color, strokeWidth float32) {
	c.circles = append(c.circles, switchCircle{center: center, radius: radius, stroke: strokeWidth, color: color})
	c.recordCanvas.StrokeCircle(center, radius, color, strokeWidth)
}

// switchState returns the paint state of a switch at rest whose control
// bounds start at (10, 10).
func switchState(checked bool) toggleswitch.PaintState {
	p := devtools.SwitchPainter{}
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
var switchTrack = geometry.NewRect(13, 13, 26, 16)

func switchColorNear(a, b widget.Color) bool {
	const eps = 1e-4
	d := func(x, y float32) bool { return x-y < eps && y-x < eps }
	return d(a.R, b.R) && d(a.G, b.G) && d(a.B, b.B) && d(a.A, b.A)
}

// paintSwitch paints s and returns the canvas and its knob.
func paintSwitch(t *testing.T, p devtools.SwitchPainter, s toggleswitch.PaintState) (*switchCanvas, switchCircle) {
	t.Helper()
	c := &switchCanvas{}
	p.PaintSwitch(c, s)
	if len(c.circles) != 1 {
		t.Fatalf("want 1 knob circle, got %d", len(c.circles))
	}
	return c, c.circles[0]
}

func TestSwitchPainter_EmptyControlBounds(t *testing.T) {
	c := &switchCanvas{}
	devtools.SwitchPainter{}.PaintSwitch(c, toggleswitch.PaintState{Label: "Wi-Fi"})
	if len(c.calls) != 0 {
		t.Errorf("empty control bounds should draw nothing, got %d calls", len(c.calls))
	}
}

func TestSwitchPainter_OffDark(t *testing.T) {
	c, knob := paintSwitch(t, devtools.SwitchPainter{}, switchState(false))

	tracks := c.methodCalls(methodDrawRoundRect)
	if len(tracks) != 1 || tracks[0].bounds != switchTrack || tracks[0].radius != 8 {
		t.Fatalf("track calls = %+v, want one 26×16 pill", tracks)
	}
	if !switchColorNear(tracks[0].color, widget.HexA(0xFFFFFF3B)) {
		t.Errorf("dark off track = %v, want #FFFFFF3B", tracks[0].color)
	}
	if knob.center != geometry.Pt(21, 21) || knob.radius != 3 || knob.stroke != 2 {
		t.Errorf("off knob = %+v, want a r3 ring at (21, 21)", knob)
	}
	if !switchColorNear(knob.color, widget.Hex(0xD1D3D9)) {
		t.Errorf("dark off knob = %v, want #D1D3D9", knob.color)
	}
	texts := c.methodCalls(methodDrawText)
	if len(texts) != 1 || texts[0].text != "Wi-Fi" || texts[0].fontSize != 13 ||
		texts[0].bounds != switchState(false).LabelBounds {
		t.Errorf("label calls = %+v", texts)
	}
}

func TestSwitchPainter_On(t *testing.T) {
	th := devtools.NewDarkTheme()
	c, knob := paintSwitch(t, devtools.SwitchPainter{Theme: th}, switchState(true))

	if track := c.methodCalls(methodDrawRoundRect)[0]; !switchColorNear(track.color, th.Colors.Primary) {
		t.Errorf("on track = %v, want Primary", track.color)
	}
	if knob.center != geometry.Pt(31, 21) || knob.radius != 5 || knob.stroke != 0 {
		t.Errorf("on knob = %+v, want a r5 dot at (31, 21)", knob)
	}
	if !switchColorNear(knob.color, th.Colors.OnPrimary) {
		t.Errorf("on knob = %v, want OnPrimary", knob.color)
	}
}

func TestSwitchPainter_KnobFillsInTransit(t *testing.T) {
	s := switchState(true)
	s.Position, s.Progress = 0.5, 0.5
	_, knob := paintSwitch(t, devtools.SwitchPainter{}, s)

	// Halfway the ring spans radius 1 to 4.5.
	if knob.center != geometry.Pt(26, 21) || knob.radius != 2.75 || knob.stroke != 3.5 {
		t.Errorf("knob halfway = %+v, want ring r2.75 w3.5 at (26, 21)", knob)
	}
}

func TestSwitchPainter_Light(t *testing.T) {
	th := devtools.NewTheme()
	c, knob := paintSwitch(t, devtools.SwitchPainter{Theme: th}, switchState(false))
	if track := c.methodCalls(methodDrawRoundRect)[0]; !switchColorNear(track.color, widget.HexA(0x00000030)) {
		t.Errorf("light off track = %v, want #00000030", track.color)
	}
	if knob.color != widget.ColorWhite {
		t.Errorf("light off knob = %v, want white", knob.color)
	}

	s := switchState(true)
	s.Hovered, s.HoverAmount = true, 1
	c, _ = paintSwitch(t, devtools.SwitchPainter{Theme: th}, s)
	if track := c.methodCalls(methodDrawRoundRect)[0]; !switchColorNear(track.color, th.Colors.PrimaryHover) {
		t.Errorf("hovered on track = %v, want PrimaryHover", track.color)
	}
}

func TestSwitchPainter_Disabled(t *testing.T) {
	th := devtools.NewTheme()
	s := switchState(true)
	s.Disabled, s.Focused = true, true
	c, knob := paintSwitch(t, devtools.SwitchPainter{Theme: th}, s)

	if track := c.methodCalls(methodDrawRoundRect)[0]; !switchColorNear(track.color, widget.Hex(0xF7F8F9)) {
		t.Errorf("disabled track = %v, want #F7F8F9", track.color)
	}
	strokes := c.methodCalls(methodStrokeRoundRect)
	if len(strokes) != 1 || strokes[0].strokeWidth != 1 || !switchColorNear(strokes[0].color, widget.Hex(0xDDDFE4)) {
		t.Errorf("disabled switch should draw only its border, got %+v", strokes)
	}
	if !switchColorNear(knob.color, widget.Hex(0xC3C5CB)) {
		t.Errorf("disabled knob = %v, want #C3C5CB", knob.color)
	}
	if label := c.methodCalls(methodDrawText)[0]; !switchColorNear(label.color, th.Colors.OnSurfaceDisabled) {
		t.Errorf("disabled label = %v, want OnSurfaceDisabled", label.color)
	}
}

func TestSwitchPainter_FocusRing(t *testing.T) {
	th := devtools.NewDarkTheme()
	s := switchState(false)
	s.Focused = true
	c, _ := paintSwitch(t, devtools.SwitchPainter{Theme: th}, s)

	rings := c.methodCalls(methodStrokeRoundRect)
	if len(rings) != 1 {
		t.Fatalf("focused switch should draw 1 focus ring, got %d", len(rings))
	}
	ring := rings[0]
	if ring.bounds != geometry.NewRect(11, 11, 30, 20) || ring.radius != 10 || ring.strokeWidth != 2 ||
		!switchColorNear(ring.color, th.Colors.BorderFocus) {
		t.Errorf("focus ring = %+v", ring)
	}
}

func TestSwitchPainter_MetricsAndMotion(t *testing.T) {
	p := devtools.SwitchPainter{}
	if got := p.SwitchSize(); got != geometry.Sz(32, 22) {
		t.Errorf("SwitchSize = %v, want 32×22", got)
	}
	if p.SwitchThumbTravel() != 10 || p.SwitchLabelGap() != 6 || p.SwitchFontSize() != 13 {
		t.Errorf("travel/gap/font = %v/%v/%v, want 10/6/13",
			p.SwitchThumbTravel(), p.SwitchLabelGap(), p.SwitchFontSize())
	}
	if m := p.SwitchMotion(); m.Toggle != 100*time.Millisecond || m.ToggleEasing == nil {
		t.Errorf("motion = %+v, want a 100 ms eased toggle", m)
	}
}
