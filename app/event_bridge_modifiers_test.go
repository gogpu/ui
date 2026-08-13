package app

import (
	"testing"

	"github.com/gogpu/gpucontext"
	"github.com/gogpu/ui/event"
	"github.com/gogpu/ui/geometry"
	"github.com/gogpu/ui/state"
	"github.com/gogpu/ui/theme"
	"github.com/gogpu/ui/widget"
)

// modRecorder captures the modifiers of every mouse event it is handed.
type modRecorder struct {
	widget.WidgetBase
	got  []event.Modifiers
	seen int
}

func (w *modRecorder) Layout(_ widget.Context, c geometry.Constraints) geometry.Size {
	return geometry.Sz(c.MaxWidth, c.MaxHeight)
}
func (w *modRecorder) Draw(widget.Context, widget.Canvas) {}
func (w *modRecorder) Event(_ widget.Context, e event.Event) bool {
	if me, ok := e.(*event.MouseEvent); ok {
		w.got = append(w.got, me.Modifiers())
		w.seen++
		return true
	}
	return false
}

func bridgeWithRecorder(t *testing.T) (*mockEventSource, *modRecorder) {
	t.Helper()
	es := &mockEventSource{}
	sched := state.NewScheduler(func(_ []widget.Widget) {})
	w := newWindow(nil, nil, sched, theme.DefaultLight(), RenderModeHostManaged)
	rec := &modRecorder{}
	rec.SetVisible(true)
	rec.SetEnabled(true)
	w.SetRoot(rec)
	attachEventBridge(es, w)
	return es, rec
}

// simulatePointerDown dispatches a PointerDown event through the unified
// pointer pipeline. Used by modifier tests that need a click to reach
// widgets via HandlePointerEvent -> deriveMouseEvent -> HandleEvent.
func simulatePointerDown(es *mockEventSource, x, y float64, mods gpucontext.Modifiers) {
	es.onPointer(gpucontext.PointerEvent{
		Type:        gpucontext.PointerDown,
		X:           x,
		Y:           y,
		PointerID:   1,
		PointerType: gpucontext.PointerTypeMouse,
		Button:      gpucontext.ButtonLeft,
		Buttons:     gpucontext.ButtonsLeft,
		Modifiers:   mods,
		IsPrimary:   true,
	})
}

// A click while a modifier is held has to arrive as a modified click.
//
// The platform delivers rich PointerEvents with modifier state. The unified
// pipeline converts these to MouseEvents that carry the modifier bits.
func TestMouseEventsCarryHeldModifiers(t *testing.T) {
	es, rec := bridgeWithRecorder(t)

	// Alt down. Platform modifier state is tracked in the PointerEvent.
	es.onKeyPress(gpucontext.KeyLeftAlt, gpucontext.Modifiers(0))

	// Click via PointerEvent with Alt modifier. In the unified pipeline,
	// modifiers are carried by the PointerEvent from the platform.
	simulatePointerDown(es, 10, 10, gpucontext.ModAlt)

	if rec.seen == 0 {
		t.Fatal("the click never reached the widget")
	}
	if last := rec.got[len(rec.got)-1]; !last.Has(event.ModAlt) {
		t.Errorf("click carried %v, want Alt — a modifier held over a click is lost", last)
	}
}

// Releasing the modifier stops modifying the clicks that follow.
func TestMouseModifiersClearOnRelease(t *testing.T) {
	es, rec := bridgeWithRecorder(t)

	es.onKeyPress(gpucontext.KeyLeftAlt, gpucontext.Modifiers(0))
	es.onKeyRelease(gpucontext.KeyLeftAlt, gpucontext.Modifiers(0))

	// Click after Alt released — PointerEvent carries no modifier.
	simulatePointerDown(es, 10, 10, 0)

	if rec.seen == 0 {
		t.Fatal("the click never reached the widget")
	}
	if last := rec.got[len(rec.got)-1]; last.Has(event.ModAlt) {
		t.Errorf("click carried %v after Alt was released", last)
	}
}

// A modifier released while another window had focus was never seen here, so
// holding it must not survive the focus change — otherwise the next ordinary
// click is silently a modified one.
func TestMouseModifiersClearOnFocusLoss(t *testing.T) {
	es, rec := bridgeWithRecorder(t)

	es.onKeyPress(gpucontext.KeyLeftAlt, gpucontext.Modifiers(0))
	es.onFocus(false)

	// Click after focus loss — PointerEvent carries no modifier.
	simulatePointerDown(es, 10, 10, 0)

	if rec.seen == 0 {
		t.Fatal("the click never reached the widget")
	}
	if last := rec.got[len(rec.got)-1]; last.Has(event.ModAlt) {
		t.Errorf("click carried %v after the window lost focus", last)
	}
}
