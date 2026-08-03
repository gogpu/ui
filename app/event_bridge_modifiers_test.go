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

// A click while a modifier is held has to arrive as a modified click.
//
// The platform's mouse callbacks carry only a button and a position, so the
// bridge has to remember what the keyboard is holding. Without that every mouse
// event is built with ModNone and ⌥click, ⇧click and ⌃click cannot be expressed
// at all — an application either does without them or reimplements this
// tracking on top of the key events itself.
func TestMouseEventsCarryHeldModifiers(t *testing.T) {
	es, rec := bridgeWithRecorder(t)

	// Alt down, then click. Nothing else is pressed in between: that is the
	// gesture, and it is why the key's own modifier bit has to be folded in —
	// a key event reports what was held BEFORE it, so this one reports nothing.
	es.onKeyPress(gpucontext.KeyLeftAlt, gpucontext.Modifiers(0))
	es.onMousePress(gpucontext.MouseButtonLeft, 10, 10)

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
	es.onMousePress(gpucontext.MouseButtonLeft, 10, 10)

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
	es.onMousePress(gpucontext.MouseButtonLeft, 10, 10)

	if last := rec.got[len(rec.got)-1]; last.Has(event.ModAlt) {
		t.Errorf("click carried %v after the window lost focus", last)
	}
}
