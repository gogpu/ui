package app

import (
	"testing"

	"github.com/gogpu/ui/core/collapsible"
	"github.com/gogpu/ui/core/scrollview"
	"github.com/gogpu/ui/core/toggleswitch"
	"github.com/gogpu/ui/event"
	"github.com/gogpu/ui/geometry"
	"github.com/gogpu/ui/gesture"
	"github.com/gogpu/ui/primitives"
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

// newGallerySwitchWindow places the switch the way a settings page does: in a
// collapsible section of padded content inside a scroll view, scrolled with
// the wheel until the switch is in view. The switch's Bounds and its position
// in the window then differ.
func newGallerySwitchWindow(t *testing.T, opts ...toggleswitch.Option) (*Window, *toggleswitch.Widget) {
	t.Helper()
	sw := toggleswitch.New(opts...)
	section := collapsible.New(
		collapsible.Title("Inputs"),
		collapsible.Content(primitives.VBox(primitives.Box().Height(200), primitives.VBox(sw).Gap(6)).Gap(12)),
		collapsible.Expanded(true),
		collapsible.HeaderHeight(40),
	)
	content := primitives.VBox(primitives.Box().Height(500), section, primitives.Box().Height(800)).Padding(24).Gap(16)
	w := New().Window()
	w.HandleResize(500, 400)
	w.SetRoot(scrollview.New(content))
	w.Frame()
	w.DrawTo(&mockCanvas{})
	wheelAt := geometry.Pt(100, 100)
	for sw.ScreenBounds().Max.Y > 300 {
		w.HandleEvent(event.NewWheelEvent(geometry.Pt(0, 1), wheelAt, wheelAt, event.ModNone))
		w.Frame()
		w.DrawTo(&mockCanvas{})
	}
	if sw.ScreenBounds().Min == sw.Bounds().Min {
		t.Fatalf("precondition: window position %v should differ from bounds %v", sw.ScreenBounds(), sw.Bounds())
	}
	return w, sw
}

func TestWindow_ClickTogglesSwitchInScrolledContent(t *testing.T) {
	toggles := 0
	w, sw := newGallerySwitchWindow(t, toggleswitch.OnToggle(func(bool) { toggles++ }))
	pos := sw.ScreenBounds().Min.Add(geometry.Pt(20, 18))

	w.HandlePointerEvent(switchPointer(gesture.PointerDown, pos, event.ButtonStateLeft))
	// A hand on a mouse moves a little during a click.
	jitter := pos.Add(geometry.Pt(2, 1))
	w.HandlePointerEvent(switchPointer(gesture.PointerMove, jitter, event.ButtonStateLeft))
	w.HandlePointerEvent(switchPointer(gesture.PointerUp, jitter, 0))

	if toggles != 1 || !sw.IsChecked() {
		t.Errorf("click toggled %d times, checked=%v; want once, on", toggles, sw.IsChecked())
	}
}

func TestWindow_DragMovesSwitchInScrolledContent(t *testing.T) {
	toggles := 0
	w, sw := newGallerySwitchWindow(t, toggleswitch.OnToggle(func(bool) { toggles++ }))
	start := sw.ScreenBounds().Min.Add(geometry.Pt(10, 18))

	w.HandlePointerEvent(switchPointer(gesture.PointerDown, start, event.ButtonStateLeft))
	// Back and forth past the slop without passing the middle: the drag must
	// follow the pointer from where it was pressed, so the switch stays off.
	for _, dx := range []float32{6, 8, 6, 2} {
		w.HandlePointerEvent(switchPointer(gesture.PointerMove, start.Add(geometry.Pt(dx, 0)), event.ButtonStateLeft))
	}
	w.HandlePointerEvent(switchPointer(gesture.PointerUp, start.Add(geometry.Pt(2, 0)), 0))
	if toggles != 0 || sw.IsChecked() {
		t.Fatalf("short drag toggled %d times, checked=%v; want no change", toggles, sw.IsChecked())
	}

	w.HandlePointerEvent(switchPointer(gesture.PointerDown, start, event.ButtonStateLeft))
	for dx := float32(5); dx <= 40; dx += 5 {
		w.HandlePointerEvent(switchPointer(gesture.PointerMove, start.Add(geometry.Pt(dx, 0)), event.ButtonStateLeft))
	}
	w.HandlePointerEvent(switchPointer(gesture.PointerUp, start.Add(geometry.Pt(40, 0)), 0))
	if toggles != 1 || !sw.IsChecked() {
		t.Errorf("drag across toggled %d times, checked=%v; want once, on", toggles, sw.IsChecked())
	}
}
