package cupertino_test

import (
	"testing"
	"time"

	"github.com/gogpu/ui/core/toggleswitch"
	"github.com/gogpu/ui/geometry"
	"github.com/gogpu/ui/theme/cupertino"
	"github.com/gogpu/ui/widget"
)

// switchState returns the paint state of a switch at rest whose control
// bounds start at (10, 10).
func switchState(checked bool) toggleswitch.PaintState {
	p := cupertino.SwitchPainter{}
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
var switchTrack = geometry.NewRect(14, 14, 51, 31)

func switchColorNear(a, b widget.Color) bool {
	const eps = 1e-4
	d := func(x, y float32) bool { return x-y < eps && y-x < eps }
	return d(a.R, b.R) && d(a.G, b.G) && d(a.B, b.B) && d(a.A, b.A)
}

func paintSwitch(p cupertino.SwitchPainter, s toggleswitch.PaintState) *recordCanvas {
	c := &recordCanvas{}
	p.PaintSwitch(c, s)
	return c
}

// switchParts returns the track and the thumb: the first and last round
// rects. The shadow layers and the hairline border lie between them.
func switchParts(t *testing.T, c *recordCanvas) (track, thumb drawCall) {
	t.Helper()
	rects := c.methodCalls(methodDrawRoundRect)
	if len(rects) != 6 {
		t.Fatalf("want 6 DrawRoundRect (track, 3 shadows, border, thumb), got %d", len(rects))
	}
	return rects[0], rects[5]
}

func TestSwitchPainter_EmptyControlBounds(t *testing.T) {
	c := paintSwitch(cupertino.SwitchPainter{}, toggleswitch.PaintState{Label: "Wi-Fi"})
	if len(c.calls) != 0 {
		t.Errorf("empty control bounds should draw nothing, got %d calls", len(c.calls))
	}
}

func TestSwitchPainter_OffAndOn(t *testing.T) {
	th := cupertino.NewTheme()

	track, thumb := switchParts(t, paintSwitch(cupertino.SwitchPainter{Theme: th}, switchState(false)))
	if track.bounds != switchTrack || track.radius != 15.5 {
		t.Errorf("track = %v r%v, want %v r15.5", track.bounds, track.radius, switchTrack)
	}
	if !switchColorNear(track.color, th.Colors.SecondarySystemFill) {
		t.Errorf("off track = %v, want SecondarySystemFill", track.color)
	}
	if thumb.bounds != geometry.NewRect(15.5, 15.5, 28, 28) || thumb.color != widget.ColorWhite {
		t.Errorf("off thumb = %v %v, want white 28×28 at the left", thumb.bounds, thumb.color)
	}

	track, thumb = switchParts(t, paintSwitch(cupertino.SwitchPainter{Theme: th}, switchState(true)))
	if !switchColorNear(track.color, th.Colors.SystemGreen) {
		t.Errorf("on track = %v, want SystemGreen", track.color)
	}
	if thumb.bounds != geometry.NewRect(35.5, 15.5, 28, 28) {
		t.Errorf("on thumb = %v, want 28×28 at the right", thumb.bounds)
	}
}

func TestSwitchPainter_PressStretchesThumb(t *testing.T) {
	for _, checked := range []bool{false, true} {
		s := switchState(checked)
		s.Pressed, s.PressAmount = true, 1
		_, thumb := switchParts(t, paintSwitch(cupertino.SwitchPainter{}, s))
		if thumb.bounds.Width() != 35 || thumb.radius != 14 {
			t.Errorf("checked=%v: pressed thumb = %v r%v, want 35 wide pill", checked, thumb.bounds, thumb.radius)
		}
		// The thumb stays against its end and stretches toward the middle.
		if !checked && thumb.bounds.Min.X != 15.5 {
			t.Errorf("pressed off thumb left = %v, want 15.5", thumb.bounds.Min.X)
		}
		if checked && thumb.bounds.Max.X != 63.5 {
			t.Errorf("pressed on thumb right = %v, want 63.5", thumb.bounds.Max.X)
		}
	}
}

func TestSwitchPainter_Hover(t *testing.T) {
	th := cupertino.NewTheme()
	s := switchState(false)
	s.Hovered, s.HoverAmount = true, 1
	track, _ := switchParts(t, paintSwitch(cupertino.SwitchPainter{Theme: th}, s))
	if !switchColorNear(track.color, th.Colors.SystemFill) {
		t.Errorf("hovered off track = %v, want SystemFill", track.color)
	}
}

func TestSwitchPainter_Disabled(t *testing.T) {
	for _, th := range []*cupertino.Theme{cupertino.NewTheme(), cupertino.NewDarkTheme()} {
		s := switchState(true)
		s.Disabled, s.Focused = true, true
		c := paintSwitch(cupertino.SwitchPainter{Theme: th}, s)

		track, thumb := switchParts(t, c)
		green := th.Colors.SystemGreen
		if !switchColorNear(track.color, green.WithAlpha(green.A*0.5)) {
			t.Errorf("dark=%v: disabled track = %v, want green at half opacity", th.IsDark(), track.color)
		}
		// The thumb is opaque, blended halfway with the background.
		want := widget.ColorWhite.Lerp(th.Colors.SystemBackground, 0.5)
		if !switchColorNear(thumb.color, want) {
			t.Errorf("dark=%v: disabled thumb = %v, want %v", th.IsDark(), thumb.color, want)
		}
		if n := len(c.methodCalls(methodStrokeRoundRect)); n != 0 {
			t.Errorf("dark=%v: disabled switch should draw no focus ring, got %d", th.IsDark(), n)
		}
		if label := c.methodCalls(methodDrawText)[0]; !switchColorNear(label.color, th.Colors.TertiaryLabel) {
			t.Errorf("dark=%v: disabled label = %v, want TertiaryLabel", th.IsDark(), label.color)
		}
	}
}

func TestSwitchPainter_FocusRing(t *testing.T) {
	s := switchState(false)
	s.Focused = true
	c := paintSwitch(cupertino.SwitchPainter{Theme: cupertino.NewTheme()}, s)

	rings := c.methodCalls(methodStrokeRoundRect)
	if len(rings) != 1 {
		t.Fatalf("focused switch should draw 1 focus ring, got %d", len(rings))
	}
	ring := rings[0]
	if ring.bounds != switchTrack.Expand(1.75) || ring.strokeWidth != 3.5 || ring.radius != 17.25 {
		t.Errorf("focus ring = %v r%v w%v", ring.bounds, ring.radius, ring.strokeWidth)
	}
	if a := ring.color.A; a < 0.79 || a > 0.81 || ring.color.B <= ring.color.R {
		t.Errorf("focus ring color = %v, want a blue at 0.8 opacity", ring.color)
	}
}

func TestSwitchPainter_Label(t *testing.T) {
	th := cupertino.NewDarkTheme()
	texts := paintSwitch(cupertino.SwitchPainter{Theme: th}, switchState(false)).methodCalls(methodDrawText)
	if len(texts) != 1 || texts[0].text != "Wi-Fi" || texts[0].fontSize != 15 {
		t.Fatalf("label calls = %+v, want one 15pt \"Wi-Fi\"", texts)
	}
	if texts[0].bounds != switchState(false).LabelBounds || texts[0].color != th.Colors.Label {
		t.Errorf("label = %v %v, want LabelBounds in Label color", texts[0].bounds, texts[0].color)
	}

	s := switchState(false)
	s.Label = ""
	if n := len(paintSwitch(cupertino.SwitchPainter{}, s).methodCalls(methodDrawText)); n != 0 {
		t.Errorf("switch without label drew %d texts", n)
	}
}

func TestSwitchPainter_MetricsAndMotion(t *testing.T) {
	p := cupertino.SwitchPainter{}
	if got := p.SwitchSize(); got != geometry.Sz(59, 39) {
		t.Errorf("SwitchSize = %v, want 59×39", got)
	}
	if p.SwitchThumbTravel() != 20 || p.SwitchLabelGap() != 8 || p.SwitchFontSize() != 15 {
		t.Errorf("travel/gap/font = %v/%v/%v, want 20/8/15",
			p.SwitchThumbTravel(), p.SwitchLabelGap(), p.SwitchFontSize())
	}
	m := p.SwitchMotion()
	if m.Toggle != 200*time.Millisecond || m.Press != 300*time.Millisecond || m.ToggleEasing == nil {
		t.Errorf("motion = %+v, want 200 ms eased toggle and 300 ms press", m)
	}
}
