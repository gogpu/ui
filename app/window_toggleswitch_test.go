package app

import (
	"testing"

	"github.com/gogpu/ui/core/toggleswitch"
	"github.com/gogpu/ui/event"
	"github.com/gogpu/ui/geometry"
	"github.com/gogpu/ui/gesture"
	"github.com/gogpu/ui/state"
)

func switchPointer(typ gesture.PointerEventType, pos geometry.Point, buttons event.ButtonState) *gesture.PointerEvent {
	button := event.ButtonNone
	if typ == gesture.PointerDown || typ == gesture.PointerUp {
		button = event.ButtonLeft
	}
	return &gesture.PointerEvent{
		Base:           event.NewBase(event.TypeMouse, event.ModNone),
		EventType:      typ,
		PointerID:      1,
		PointerType:    gesture.PointerTypeMouse,
		Position:       pos,
		GlobalPosition: pos,
		Button:         button,
		Buttons:        buttons,
	}
}

func newSwitchWindow(t *testing.T, opts ...toggleswitch.Option) (*Window, *toggleswitch.Widget) {
	t.Helper()
	sw := toggleswitch.New(opts...)
	w := New().Window()
	w.SetRoot(sw)
	w.Frame()
	if sw.Bounds().IsEmpty() {
		t.Fatal("precondition: switch should be laid out")
	}
	return w, sw
}

func TestWindow_PointerClickTogglesSwitchOnce(t *testing.T) {
	toggles := 0
	w, sw := newSwitchWindow(t, toggleswitch.OnToggle(func(bool) { toggles++ }))
	pos := sw.Bounds().Center()

	w.HandlePointerEvent(switchPointer(gesture.PointerDown, pos, event.ButtonStateLeft))
	w.HandlePointerEvent(switchPointer(gesture.PointerUp, pos, 0))

	if toggles != 1 || !sw.IsChecked() {
		t.Errorf("click toggled %d times, checked=%v; want once, on", toggles, sw.IsChecked())
	}
}

func TestWindow_PointerDragMovesSwitchThumb(t *testing.T) {
	toggles := 0
	w, sw := newSwitchWindow(t, toggleswitch.OnToggle(func(bool) { toggles++ }))
	b := sw.Bounds()
	start := geometry.Pt(b.Min.X+10, b.Center().Y)

	w.HandlePointerEvent(switchPointer(gesture.PointerDown, start, event.ButtonStateLeft))
	for dx := float32(5); dx <= 40; dx += 5 {
		w.HandlePointerEvent(switchPointer(gesture.PointerMove, geometry.Pt(start.X+dx, start.Y), event.ButtonStateLeft))
	}
	w.HandlePointerEvent(switchPointer(gesture.PointerUp, geometry.Pt(start.X+40, start.Y), 0))

	if toggles != 1 || !sw.IsChecked() {
		t.Errorf("drag toggled %d times, checked=%v; want once, on", toggles, sw.IsChecked())
	}
}

func TestWindow_FocusLossCancelsSwitchDrag(t *testing.T) {
	value := state.NewSignal(false)
	w, sw := newSwitchWindow(t, toggleswitch.CheckedSignal(value))
	b := sw.Bounds()
	start := geometry.Pt(b.Min.X+10, b.Center().Y)

	w.HandleEvent(event.NewMouseEvent(event.MousePress, event.ButtonLeft, event.ButtonStateLeft,
		start, start, event.ModNone))
	if w.capturedWidget != sw {
		t.Fatal("precondition: switch should hold pointer capture")
	}
	far := geometry.Pt(start.X+40, start.Y)
	w.HandleEvent(event.NewMouseEvent(event.MouseMove, event.ButtonNone, event.ButtonStateLeft,
		far, far, event.ModNone))

	w.HandleFocusChange(false)
	if value.Get() {
		t.Error("focus loss during a drag should not turn the switch on")
	}
	if w.capturedWidget != nil {
		t.Error("focus loss should release the switch's pointer capture")
	}
}
