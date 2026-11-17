package app

import (
	"testing"

	"github.com/gogpu/ui/core/scrollview"
	"github.com/gogpu/ui/event"
	"github.com/gogpu/ui/geometry"
	"github.com/gogpu/ui/primitives"
)

// A captured widget inside scrolled, padded content must get its moves and
// release in the coordinates of its Bounds, the same space as the press that
// reached it through the tree.
func TestWindow_PointerCapture_EventsInWidgetCoordinates(t *testing.T) {
	cw := newCaptureWidget(geometry.NewRect(0, 0, 100, 40))
	cw.captureOnce = true
	content := primitives.VBox(primitives.Box().Height(600), primitives.VBox(cw)).Padding(24)
	w := New().Window()
	w.HandleResize(400, 300)
	w.SetRoot(scrollview.New(content, scrollview.ScrollY(500)))
	w.Frame()
	w.DrawTo(&mockCanvas{})

	screen := cw.ScreenBounds()
	if screen.Min == cw.Bounds().Min {
		t.Fatalf("precondition: screen origin %v should differ from bounds %v", screen.Min, cw.Bounds())
	}
	press := screen.Min.Add(geometry.Pt(10, 10))
	outside := press.Add(geometry.Pt(150, 0))
	w.HandleEvent(event.NewMouseEvent(event.MousePress, event.ButtonLeft, event.ButtonStateLeft, press, press, event.ModNone))
	w.HandleEvent(event.NewMouseEvent(event.MouseMove, event.ButtonNone, event.ButtonStateLeft, outside, outside, event.ModNone))
	w.HandleEvent(event.NewMouseEvent(event.MouseRelease, event.ButtonLeft, 0, outside, outside, event.ModNone))

	local := cw.Bounds().Min.Add(geometry.Pt(10, 10))
	want := map[event.MouseEventType]geometry.Point{
		event.MousePress:   local,
		event.MouseMove:    local.Add(geometry.Pt(150, 0)),
		event.MouseRelease: local.Add(geometry.Pt(150, 0)),
	}
	seen := 0
	for _, e := range cw.events {
		p, ok := want[e.MouseType]
		if !ok {
			continue
		}
		seen++
		if e.Position != p {
			t.Errorf("%v at %v, want %v", e.MouseType, e.Position, p)
		}
		if e.GlobalPosition == e.Position {
			t.Errorf("%v: GlobalPosition should stay in window coordinates", e.MouseType)
		}
	}
	if seen != len(want) {
		t.Errorf("got %d of the press, move and release events", seen)
	}
	if w.capturedWidget != nil {
		t.Error("release should end the capture")
	}
}

// A widget that has never been drawn has no screen origin; captured events
// reach it unchanged.
func TestWindow_PointerCapture_UndrawnWidgetGetsWindowCoordinates(t *testing.T) {
	cw := &captureTrackingWidget{}
	cw.SetVisible(true)
	cw.SetEnabled(true)
	cw.SetBounds(geometry.NewRect(50, 50, 100, 40))
	w := New().Window()
	w.SetRoot(newHoverContainer(cw))
	w.ctx.CapturePointer(cw)

	p := geometry.Pt(300, 300)
	w.HandleEvent(event.NewMouseEvent(event.MouseMove, event.ButtonNone, event.ButtonStateLeft, p, p, event.ModNone))
	if len(cw.events) != 1 || cw.events[0].Position != p {
		t.Fatalf("events = %v, want one move at %v", cw.events, p)
	}
}
