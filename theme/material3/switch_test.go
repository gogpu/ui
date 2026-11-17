package material3_test

import (
	"testing"
	"time"

	"github.com/gogpu/ui/core/toggleswitch"
	"github.com/gogpu/ui/geometry"
	"github.com/gogpu/ui/theme/material3"
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
	color  widget.Color
}

func (c *switchCanvas) DrawCircle(center geometry.Point, radius float32, color widget.Color) {
	c.circles = append(c.circles, switchCircle{center: center, radius: radius, color: color})
	c.recordCanvas.DrawCircle(center, radius, color)
}

// switchState returns the paint state of a switch at rest whose control
// bounds start at (10, 10).
func switchState(checked bool) toggleswitch.PaintState {
	p := material3.SwitchPainter{}
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
var switchTrack = geometry.NewRect(14, 14, 52, 32)

func colorNear(a, b widget.Color) bool {
	const eps = 1e-4
	d := func(x, y float32) bool { return x-y < eps && y-x < eps }
	return d(a.R, b.R) && d(a.G, b.G) && d(a.B, b.B) && d(a.A, b.A)
}

func paintSwitch(p material3.SwitchPainter, s toggleswitch.PaintState) *switchCanvas {
	c := &switchCanvas{}
	p.PaintSwitch(c, s)
	return c
}

// switchThumb returns the thumb, the last round rect a switch draws.
func switchThumb(t *testing.T, c *switchCanvas) drawCall {
	t.Helper()
	rects := c.methodCalls(methodDrawRoundRect)
	if len(rects) != 2 {
		t.Fatalf("want 2 DrawRoundRect (track, thumb), got %d", len(rects))
	}
	return rects[1]
}

func TestSwitchPainter_EmptyControlBounds(t *testing.T) {
	c := paintSwitch(material3.SwitchPainter{}, toggleswitch.PaintState{Label: "Wi-Fi"})
	if len(c.calls) != 0 {
		t.Errorf("empty control bounds should draw nothing, got %d calls", len(c.calls))
	}
}

func TestSwitchPainter_Off(t *testing.T) {
	th := material3.New(widget.Hex(0x6750A4))
	c := paintSwitch(material3.SwitchPainter{Theme: th}, switchState(false))

	track := c.methodCalls(methodDrawRoundRect)[0]
	if track.bounds != switchTrack || track.radius != 16 {
		t.Errorf("track = %v r%v, want %v r16", track.bounds, track.radius, switchTrack)
	}
	if !colorNear(track.color, th.Colors.SurfaceContainerHighest) {
		t.Errorf("off track color = %v, want SurfaceContainerHighest", track.color)
	}

	outlines := c.methodCalls(methodStrokeRoundRect)
	if len(outlines) != 1 {
		t.Fatalf("off switch should draw 1 outline, got %d", len(outlines))
	}
	if outlines[0].strokeWidth != 2 || !colorNear(outlines[0].color, th.Colors.Outline) {
		t.Errorf("outline = %v w%v, want Outline w2", outlines[0].color, outlines[0].strokeWidth)
	}

	thumb := switchThumb(t, c)
	if thumb.bounds != geometry.NewRect(22, 22, 16, 16) {
		t.Errorf("off thumb = %v, want 16×16 centered at x=30", thumb.bounds)
	}
	if !colorNear(thumb.color, th.Colors.Outline) {
		t.Errorf("off thumb color = %v, want Outline", thumb.color)
	}
	if len(c.circles) != 0 {
		t.Errorf("a switch at rest should draw no state layer, got %d", len(c.circles))
	}

	texts := c.methodCalls(methodDrawText)
	if len(texts) != 1 || texts[0].text != "Wi-Fi" || texts[0].fontSize != 14 {
		t.Fatalf("label calls = %+v, want one 14px \"Wi-Fi\"", texts)
	}
	if texts[0].bounds != switchState(false).LabelBounds {
		t.Errorf("label bounds = %v, want LabelBounds", texts[0].bounds)
	}
	if !colorNear(texts[0].color, th.Colors.OnSurface) {
		t.Errorf("label color = %v, want OnSurface", texts[0].color)
	}
}

func TestSwitchPainter_On(t *testing.T) {
	th := material3.New(widget.Hex(0x6750A4))
	c := paintSwitch(material3.SwitchPainter{Theme: th}, switchState(true))

	if track := c.methodCalls(methodDrawRoundRect)[0]; !colorNear(track.color, th.Colors.Primary) {
		t.Errorf("on track color = %v, want Primary", track.color)
	}
	if n := len(c.methodCalls(methodStrokeRoundRect)); n != 0 {
		t.Errorf("on switch should draw no outline, got %d strokes", n)
	}
	thumb := switchThumb(t, c)
	if thumb.bounds != geometry.NewRect(38, 18, 24, 24) {
		t.Errorf("on thumb = %v, want 24×24 centered at x=50", thumb.bounds)
	}
	if !colorNear(thumb.color, th.Colors.OnPrimary) {
		t.Errorf("on thumb color = %v, want OnPrimary", thumb.color)
	}
}

func TestSwitchPainter_ThumbStretchesInTransit(t *testing.T) {
	tests := []struct {
		name      string
		progress  float32
		turningOn bool
		want      geometry.Size
	}{
		{"turning on, stretched", 0.11, true, geometry.Sz(34, 22)},
		{"turning on, settled", 0.9, true, geometry.Sz(24, 24)},
		{"turning off, stretched", 0.89, false, geometry.Sz(34, 22)},
		{"turning off, settled", 0.1, false, geometry.Sz(16, 16)},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			s := switchState(tt.turningOn)
			s.Position, s.Progress, s.TurningOn = tt.progress, tt.progress, tt.turningOn
			thumb := switchThumb(t, paintSwitch(material3.SwitchPainter{}, s))
			got := thumb.bounds.Size()
			if diff := got.Sub(tt.want); max(diff.Width, -diff.Width, diff.Height, -diff.Height) > 0.01 {
				t.Errorf("thumb size = %v, want %v", got, tt.want)
			}
			if thumb.radius != min(got.Width, got.Height)/2 {
				t.Errorf("thumb radius = %v, want a pill", thumb.radius)
			}
		})
	}
}

func TestSwitchPainter_Overshoot(t *testing.T) {
	s := switchState(true)
	s.Position = 1.05
	thumb := switchThumb(t, paintSwitch(material3.SwitchPainter{}, s))
	if cx := thumb.bounds.Center().X; cx <= 50 {
		t.Errorf("thumb center x = %v, want past 50 while overshooting", cx)
	}
}

func TestSwitchPainter_Pressed(t *testing.T) {
	th := material3.New(widget.Hex(0x6750A4))
	s := switchState(false)
	s.Pressed, s.PressAmount = true, 1
	c := paintSwitch(material3.SwitchPainter{Theme: th}, s)

	if size := switchThumb(t, c).bounds.Size(); size != geometry.Sz(28, 28) {
		t.Errorf("pressed thumb = %v, want 28×28", size)
	}
	if len(c.circles) != 1 {
		t.Fatalf("pressed switch should draw the state layer, got %d circles", len(c.circles))
	}
	layer := c.circles[0]
	if layer.radius != 20 || layer.center != geometry.Pt(30, 30) {
		t.Errorf("state layer = r%v at %v, want r20 at (30, 30)", layer.radius, layer.center)
	}
	if !colorNear(layer.color, th.Colors.OnSurface.WithAlpha(0.1)) {
		t.Errorf("off press layer = %v, want OnSurface at 0.1", layer.color)
	}
}

func TestSwitchPainter_HoverOn(t *testing.T) {
	th := material3.New(widget.Hex(0x6750A4))
	s := switchState(true)
	s.Hovered, s.HoverAmount = true, 1
	c := paintSwitch(material3.SwitchPainter{Theme: th}, s)

	if len(c.circles) != 1 || !colorNear(c.circles[0].color, th.Colors.Primary.WithAlpha(0.08)) {
		t.Fatalf("hover layer = %+v, want Primary at 0.08", c.circles)
	}
	if thumb := switchThumb(t, c); !colorNear(thumb.color, th.Colors.PrimaryContainer) {
		t.Errorf("hovered on thumb = %v, want PrimaryContainer", thumb.color)
	}
}

func TestSwitchPainter_Disabled(t *testing.T) {
	th := material3.New(widget.Hex(0x6750A4))
	cs := th.Colors

	s := switchState(false)
	s.Disabled, s.Focused, s.HoverAmount = true, true, 1
	c := paintSwitch(material3.SwitchPainter{Theme: th}, s)
	if track := c.methodCalls(methodDrawRoundRect)[0]; !colorNear(track.color, cs.SurfaceContainerHighest.WithAlpha(0.12)) {
		t.Errorf("disabled off track = %v", track.color)
	}
	strokes := c.methodCalls(methodStrokeRoundRect)
	if len(strokes) != 1 || !colorNear(strokes[0].color, cs.OnSurface.WithAlpha(0.12)) {
		t.Errorf("disabled switch should draw only the dim outline, got %+v", strokes)
	}
	if thumb := switchThumb(t, c); !colorNear(thumb.color, cs.OnSurface.WithAlpha(0.38)) {
		t.Errorf("disabled off thumb = %v", thumb.color)
	}
	if len(c.circles) != 0 {
		t.Errorf("disabled switch should draw no state layer")
	}
	if label := c.methodCalls(methodDrawText)[0]; !colorNear(label.color, cs.OnSurface.WithAlpha(0.38)) {
		t.Errorf("disabled label = %v", label.color)
	}

	s = switchState(true)
	s.Disabled = true
	c = paintSwitch(material3.SwitchPainter{Theme: th}, s)
	if track := c.methodCalls(methodDrawRoundRect)[0]; !colorNear(track.color, cs.OnSurface.WithAlpha(0.12)) {
		t.Errorf("disabled on track = %v", track.color)
	}
	if thumb := switchThumb(t, c); !colorNear(thumb.color, cs.Surface) {
		t.Errorf("disabled on thumb = %v, want Surface", thumb.color)
	}
}

func TestSwitchPainter_FocusRing(t *testing.T) {
	th := material3.New(widget.Hex(0x6750A4))
	s := switchState(true)
	s.Focused = true
	c := paintSwitch(material3.SwitchPainter{Theme: th}, s)

	strokes := c.methodCalls(methodStrokeRoundRect)
	if len(strokes) != 1 {
		t.Fatalf("focused on switch should draw 1 focus ring, got %d", len(strokes))
	}
	if strokes[0].bounds != switchTrack.Expand(2) || strokes[0].radius != 18 || strokes[0].strokeWidth != 2 {
		t.Errorf("focus ring = %+v", strokes[0])
	}
	if thumb := switchThumb(t, c); !colorNear(thumb.color, th.Colors.PrimaryContainer) {
		t.Errorf("focused on thumb = %v, want PrimaryContainer", thumb.color)
	}
}

func TestSwitchPainter_Themes(t *testing.T) {
	c := paintSwitch(material3.SwitchPainter{}, switchState(true))
	if track := c.methodCalls(methodDrawRoundRect)[0]; !colorNear(track.color, widget.Hex(0x6750A4)) {
		t.Errorf("nil theme on track = %v, want #6750A4", track.color)
	}

	dark := material3.NewDark(widget.Hex(0x6750A4))
	c = paintSwitch(material3.SwitchPainter{Theme: dark}, switchState(true))
	if thumb := switchThumb(t, c); !colorNear(thumb.color, dark.Colors.OnPrimary) {
		t.Errorf("dark on thumb = %v, want dark OnPrimary", thumb.color)
	}
}

func TestSwitchPainter_MetricsAndMotion(t *testing.T) {
	p := material3.SwitchPainter{}
	if got := p.SwitchSize(); got != geometry.Sz(60, 40) {
		t.Errorf("SwitchSize = %v, want 60×40", got)
	}
	if p.SwitchThumbTravel() != 20 || p.SwitchLabelGap() != 8 || p.SwitchFontSize() != 14 {
		t.Errorf("travel/gap/font = %v/%v/%v, want 20/8/14",
			p.SwitchThumbTravel(), p.SwitchLabelGap(), p.SwitchFontSize())
	}

	m := p.SwitchMotion()
	if m.Toggle != 300*time.Millisecond || m.Hover != 100*time.Millisecond || m.Press != 100*time.Millisecond {
		t.Errorf("motion = %+v", m)
	}
	var peak float32
	for i := 1; i < 100; i++ {
		peak = max(peak, m.ToggleEasing(float32(i)/100))
	}
	if peak <= 1 || m.ToggleEasing(1) != 1 {
		t.Errorf("toggle easing should overshoot and land on 1, peak %v", peak)
	}
}
