package app

import (
	"testing"

	"github.com/gogpu/gpucontext"
	"github.com/gogpu/ui/core/button"
	"github.com/gogpu/ui/core/dialog"
	"github.com/gogpu/ui/event"
	"github.com/gogpu/ui/geometry"
	"github.com/gogpu/ui/primitives"
	"github.com/gogpu/ui/widget"
)

func TestPopupAboveModalBlocksBackgroundPointerMove(t *testing.T) {
	win := New().Window()
	background := &overlayPointerProbe{Widget: button.New()}
	win.SetRoot(background)
	win.HandleResize(500, 400)
	modal := dialog.New(dialog.Content(primitives.Box().Width(120).Height(80)))
	modal.Show(win.Context())
	win.Frame()

	move := func(point geometry.Point) {
		win.HandleEvent(event.NewMouseEvent(event.MouseMove, event.ButtonNone, 0, point, point, event.ModNone))
	}
	outside := geometry.Pt(10, 10)
	move(outside)
	if background.pointerEvents != 0 {
		t.Fatal("modal dialog allowed background mouse move")
	}

	popup := button.New()
	manager := win.Context().OverlayManager()
	manager.PushOverlay(popup, nil)
	win.Frame()
	popup.SetBounds(geometry.NewRect(200, 150, 300, 200))
	move(popup.Bounds().Center())
	if win.HoveredWidget() != popup {
		t.Fatal("popup above the dialog did not receive hover")
	}
	move(outside)
	if background.pointerEvents != 0 {
		t.Fatalf("popup above modal allowed %d background pointer events", background.pointerEvents)
	}

	manager.RemoveOverlay(popup)
	modal.Close(win.Context())
	move(outside)
	if background.pointerEvents != 1 {
		t.Fatal("background did not receive mouse move after closing overlays")
	}
}

type overlayPointerProbe struct {
	*button.Widget
	pointerEvents int
}

func (p *overlayPointerProbe) Event(ctx widget.Context, e event.Event) bool {
	switch e := e.(type) {
	case *event.MouseEvent:
		if e.MouseType == event.MouseMove || e.MouseType == event.MousePress || e.MouseType == event.MouseRelease {
			p.pointerEvents++
		}
	case *event.WheelEvent:
		p.pointerEvents++
	}
	return p.Widget.Event(ctx, e)
}

func TestOverlayBlocksPointerEventsAndGestures(t *testing.T) {
	es := &mockEventSource{}
	a := New(WithEventSource(es))
	win := a.Window()
	var backgroundClicks, popupClicks int
	background := &overlayPointerProbe{Widget: button.New(button.OnClick(func() { backgroundClicks++ }))}
	win.SetRoot(background)
	win.HandleResize(500, 400)
	win.Frame()

	panel := primitives.Box()
	panel.SetBounds(geometry.NewRect(80, 60, 200, 140))
	manager := win.Context().OverlayManager()
	manager.PushOverlay(panel, func() { manager.RemoveOverlay(panel) })
	pointer := func(kind gpucontext.PointerEventType, at geometry.Point, buttons gpucontext.Buttons) {
		es.onPointer(gpucontext.PointerEvent{
			Type: kind, PointerID: 1, PointerType: gpucontext.PointerTypeMouse,
			Button: gpucontext.ButtonLeft, Buttons: buttons, X: float64(at.X), Y: float64(at.Y),
		})
	}
	click := func(at geometry.Point) {
		pointer(gpucontext.PointerDown, at, gpucontext.ButtonsLeft)
		pointer(gpucontext.PointerUp, at, gpucontext.ButtonsNone)
	}

	inside := geometry.Pt(110, 80)
	pointer(gpucontext.PointerMove, inside, gpucontext.ButtonsNone)
	win.HandleEvent(event.NewWheelEvent(geometry.Pt(0, 3), inside, inside, event.ModNone))
	click(inside)
	if win.OverlayCount() != 1 || background.pointerEvents != 0 || backgroundClicks != 0 {
		t.Fatalf("inside popup: overlays=%d, background events=%d, clicks=%d",
			win.OverlayCount(), background.pointerEvents, backgroundClicks)
	}
	click(geometry.Pt(20, 20))
	if win.OverlayCount() != 0 || background.pointerEvents != 0 || backgroundClicks != 0 {
		t.Fatalf("outside click: overlays=%d, background events=%d, clicks=%d",
			win.OverlayCount(), background.pointerEvents, backgroundClicks)
	}

	control := button.New(button.OnClick(func() { popupClicks++ }))
	control.SetBounds(panel.Bounds())
	manager.PushOverlay(control, nil)
	click(inside)
	if popupClicks != 1 || backgroundClicks != 0 {
		t.Fatalf("popup control clicks=%d, background clicks=%d", popupClicks, backgroundClicks)
	}
	manager.RemoveOverlay(control)
	click(inside)
	if backgroundClicks != 1 {
		t.Fatalf("background clicks after dismissal=%d, want 1", backgroundClicks)
	}
}

func TestPassiveOverlayAllowsHoverAndLeavesLaterOverlayIntact(t *testing.T) {
	win := New().Window()
	var clicks int
	target := button.New(button.OnClick(func() { clicks++ }))
	win.SetRoot(target)
	win.HandleResize(500, 400)
	win.Frame()
	manager := win.Context().OverlayManager()
	tip := primitives.Box().Width(220).Height(100)
	manager.(widget.PassiveOverlayManager).PushPassiveOverlay(tip)
	win.Frame()
	point := geometry.Pt(20, 20)
	win.HandleEvent(event.NewMouseEvent(event.MouseMove, event.ButtonNone, 0, point, point, event.ModNone))
	win.HandleEvent(event.NewMouseEvent(event.MousePress, event.ButtonLeft, event.ButtonStateLeft, point, point, event.ModNone))
	win.HandleEvent(event.NewMouseEvent(event.MouseRelease, event.ButtonLeft, 0, point, point, event.ModNone))
	if win.HoveredWidget() != target || clicks != 1 {
		t.Fatalf("passive overlay intercepted input, hover=%v, clicks=%d", win.HoveredWidget(), clicks)
	}
	dialog := primitives.Box().Width(200).Height(120)
	manager.PushOverlay(dialog, nil)
	manager.RemoveOverlay(tip)
	if win.OverlayCount() != 1 || win.OverlayContentWidgets()[0] != dialog {
		t.Fatal("removing a passive overlay also removed the interactive overlay above it")
	}
}
