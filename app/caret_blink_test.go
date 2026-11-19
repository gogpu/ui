package app

import (
	"sync/atomic"
	"testing"
	"time"

	"github.com/gogpu/ui/core/textfield"
	"github.com/gogpu/ui/event"
	"github.com/gogpu/ui/geometry"
	"github.com/gogpu/ui/primitives"
	"github.com/gogpu/ui/widget"
)

// redrawLeaf is a plain widget whose redraw flag the timers set.
type redrawLeaf struct {
	widget.WidgetBase
}

func newRedrawLeaf() *redrawLeaf {
	w := &redrawLeaf{}
	w.SetVisible(true)
	w.SetNeedsRedraw(false)
	return w
}

func (w *redrawLeaf) Layout(_ widget.Context, c geometry.Constraints) geometry.Size {
	return c.Constrain(geometry.Sz(10, 10))
}
func (w *redrawLeaf) Draw(widget.Context, widget.Canvas)     {}
func (w *redrawLeaf) Event(widget.Context, event.Event) bool { return false }
func (w *redrawLeaf) Children() []widget.Widget              { return nil }

func TestRedrawTimers_FlushMarksOnlyDueWidgets(t *testing.T) {
	var wakes atomic.Int32
	timers := newRedrawTimers(func() { wakes.Add(1) })
	t0 := time.Now()
	soon, later := newRedrawLeaf(), newRedrawLeaf()

	timers.schedule(soon, 10*time.Millisecond, t0)
	timers.schedule(later, time.Hour, t0)

	timers.flush(t0.Add(5 * time.Millisecond))
	if soon.NeedsRedraw() || later.NeedsRedraw() {
		t.Fatal("flush before any deadline marked a widget")
	}

	timers.flush(t0.Add(10 * time.Millisecond))
	if !soon.NeedsRedraw() {
		t.Error("due widget not marked for redraw")
	}
	if later.NeedsRedraw() {
		t.Error("widget not yet due was marked for redraw")
	}
	if got := timers.pending(); got != 1 {
		t.Errorf("pending = %d, want 1", got)
	}
	timers.stop()
}

func TestRedrawTimers_RescheduleReplacesAndCancelDrops(t *testing.T) {
	timers := newRedrawTimers(func() {})
	defer timers.stop()
	t0 := time.Now()
	w := newRedrawLeaf()

	timers.schedule(w, 10*time.Millisecond, t0)
	timers.schedule(w, time.Hour, t0)
	timers.flush(t0.Add(time.Second))
	if w.NeedsRedraw() {
		t.Error("the replaced, earlier deadline still fired")
	}
	if got := timers.pending(); got != 1 {
		t.Errorf("pending = %d, want 1", got)
	}

	timers.cancel(w)
	if got := timers.pending(); got != 0 {
		t.Errorf("pending after cancel = %d, want 0", got)
	}
}

func TestRedrawTimers_SkipsDetachedWidgets(t *testing.T) {
	timers := newRedrawTimers(func() {})
	defer timers.stop()
	t0 := time.Now()
	w := newRedrawLeaf()

	timers.schedule(w, time.Millisecond, t0)
	widget.UnmountTree(w)
	w.SetNeedsRedraw(false)
	timers.flush(t0.Add(time.Second))
	if w.NeedsRedraw() {
		t.Error("a removed widget was marked for redraw")
	}
	if got := timers.pending(); got != 0 {
		t.Errorf("pending = %d, want 0", got)
	}
}

func TestRedrawTimers_TimerWakesTheRenderLoop(t *testing.T) {
	woke := make(chan struct{}, 1)
	timers := newRedrawTimers(func() { woke <- struct{}{} })
	defer timers.stop()

	timers.schedule(newRedrawLeaf(), 5*time.Millisecond, time.Now())
	select {
	case <-woke:
	case <-time.After(2 * time.Second):
		t.Fatal("timer did not wake the render loop")
	}
}

// caretRecorder is a text field painter that records the caret it paints.
type caretRecorder struct {
	paints     int
	focused    bool
	showCursor bool
	cursorPos  int
}

func (p *caretRecorder) PaintTextField(_ widget.Canvas, st *textfield.PaintState) {
	p.paints++
	p.focused = st.Focused
	p.showCursor = st.ShowCursor
	p.cursorPos = st.CursorPos
}

// newCaretWindow returns a hosted window showing a focused text field whose
// caret blinks every step, after its first frame.
func newCaretWindow(t *testing.T, step time.Duration) (*Window, *mockWindowProvider, *caretRecorder) {
	t.Helper()
	rec := &caretRecorder{}
	field := textfield.New(
		textfield.InitialValue("hi"),
		textfield.PainterOpt(rec),
		textfield.CaretBlinkOpt(textfield.HardCaretBlink(step)),
	)
	wp := &mockWindowProvider{width: 800, height: 600, scale: 1}
	w := New(WithWindowProvider(wp)).Window()
	t.Cleanup(w.Close)
	w.SetRoot(primitives.Box(field).Padding(16))
	w.Context().RequestFocus(field)
	w.Frame()
	w.DrawTo(&trackingCanvas{})
	if !rec.showCursor {
		t.Fatal("precondition: caret should show right after focus")
	}
	return w, wp, rec
}

// TestWindow_CaretStepRedrawsTheField runs a blink step through the window:
// the field schedules its next step, the timer wakes the render loop, and the
// next frame repaints the field with the caret hidden. Before the step is due
// a frame leaves the field alone.
func TestWindow_CaretStepRedrawsTheField(t *testing.T) {
	const step = 40 * time.Millisecond
	w, wp, rec := newCaretWindow(t, step)

	if got := w.redrawTimers.pending(); got != 1 {
		t.Fatalf("scheduled redraws = %d, want 1", got)
	}

	paints := rec.paints
	wakes := wp.redrawCount.Load()
	w.Frame()
	w.DrawTo(&trackingCanvas{})
	if rec.paints != paints {
		t.Errorf("field repainted %d times before its step was due", rec.paints-paints)
	}

	deadline := time.Now().Add(2 * time.Second)
	for wp.redrawCount.Load() == wakes {
		if time.Now().After(deadline) {
			t.Fatal("the caret's timer never woke the render loop")
		}
		time.Sleep(time.Millisecond)
	}

	w.Frame()
	if !w.HasDirtyBoundariesOrNeedsRedraw() {
		t.Fatal("frame after the wake has no work, so the render loop would skip it")
	}
	w.DrawTo(&trackingCanvas{})
	if rec.paints == paints {
		t.Fatal("field not repainted when its blink step was due")
	}
	if rec.showCursor {
		t.Error("caret still shown after its visible step ended")
	}
	if got := w.redrawTimers.pending(); got != 1 {
		t.Errorf("scheduled redraws after the step = %d, want 1 for the next step", got)
	}
}

// TestWindow_CaretHidesWhenWindowLosesFocus checks that an inactive window
// hides the caret at once and stops waking the render loop for it.
func TestWindow_CaretHidesWhenWindowLosesFocus(t *testing.T) {
	w, _, rec := newCaretWindow(t, time.Hour)

	w.HandleFocusChange(false)
	w.Frame()
	w.DrawTo(&trackingCanvas{})
	if rec.showCursor {
		t.Error("caret shown while the window is inactive")
	}
	if got := w.redrawTimers.pending(); got != 0 {
		t.Errorf("scheduled redraws in an inactive window = %d, want 0", got)
	}

	w.HandleFocusChange(true)
	w.Frame()
	w.DrawTo(&trackingCanvas{})
	if !rec.showCursor {
		t.Error("caret hidden after the window became active again")
	}
}

func TestWindow_OverlayCaretSurvivesTextRelayout(t *testing.T) {
	const step = 500 * time.Millisecond
	w := New(WithWindowProvider(&mockWindowProvider{width: 500, height: 400, scale: 1})).Window()
	t.Cleanup(w.Close)
	w.SetRoot(primitives.Box())
	rec := &caretRecorder{}
	field := textfield.New(
		textfield.PainterOpt(rec),
		textfield.CaretBlinkOpt(textfield.HardCaretBlink(step)),
		textfield.OnChange(func(string) { w.Context().Invalidate() }),
	)
	panel := primitives.VBox(field).Width(260).Height(48)
	panel.SetBounds(geometry.NewRect(80, 60, 260, 48))
	w.Context().OverlayManager().PushOverlay(panel, nil)
	w.Context().RequestFocus(field)
	w.Frame()
	PaintOverlayBoundaries(w.OverlayContentWidgets(), w.Context())
	if !rec.focused || !rec.showCursor {
		t.Fatal("precondition: popup has a focused field with a visible caret")
	}

	w.HandleEvent(event.NewKeyEvent(event.KeyPress, event.KeyA, 'a', event.ModNone))
	w.HandleEvent(event.NewKeyEvent(event.KeyRelease, event.KeyA, 'a', event.ModNone))
	w.Frame()
	overlays := w.OverlayContentWidgets()
	PaintOverlayBoundaries(overlays, w.Context())
	if field.Text() != "a" || !rec.focused || !rec.showCursor || rec.cursorPos != 1 {
		t.Fatalf("after typing: text=%q, focus=%t, caret visible=%t, position=%d",
			field.Text(), rec.focused, rec.showCursor, rec.cursorPos)
	}

	// Advance the repaint scheduler through one blink cycle without sleeping.
	start := w.Context().Now()
	for i, wantVisible := range []bool{false, true} {
		now := start.Add(time.Duration(i+1) * step)
		w.Context().SetNow(now)
		w.redrawTimers.flush(now)
		PaintOverlayBoundaries(overlays, w.Context())
		if !rec.focused || rec.showCursor != wantVisible || rec.cursorPos != 1 {
			t.Errorf("blink step %d: focus=%t, caret visible=%t, position=%d",
				i+1, rec.focused, rec.showCursor, rec.cursorPos)
		}
	}
}

// TestWindow_CloseStopsCaretTimers checks that Close leaves no timer that
// could wake a closed window.
func TestWindow_CloseStopsCaretTimers(t *testing.T) {
	w, _, _ := newCaretWindow(t, time.Hour)
	w.Close()
	if got := w.redrawTimers.pending(); got != 0 {
		t.Errorf("scheduled redraws after Close = %d, want 0", got)
	}
}
