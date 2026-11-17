package app

import (
	"sync"
	"time"

	"github.com/gogpu/ui/widget"
)

// redrawTimers backs [widget.RedrawScheduler] for a window: it keeps the
// time each widget asked to be redrawn at and one timer armed for the
// earliest of them.
//
// The timer only wakes the render loop. Widgets are marked dirty by flush,
// which Frame calls on the UI goroutine, so no widget state is touched from
// the timer's goroutine.
type redrawTimers struct {
	mu    sync.Mutex
	due   map[widget.Widget]time.Time
	timer *time.Timer
	// armedAt is the deadline the timer fires for; zero when disarmed.
	armedAt time.Time
	wake    func()
}

func newRedrawTimers(wake func()) *redrawTimers {
	return &redrawTimers{due: make(map[widget.Widget]time.Time), wake: wake}
}

// schedule asks for w to be redrawn at now+d, replacing any earlier request.
func (t *redrawTimers) schedule(w widget.Widget, d time.Duration, now time.Time) {
	t.mu.Lock()
	defer t.mu.Unlock()
	t.due[w] = now.Add(d)
	t.armLocked(now)
}

// cancel drops the redraw scheduled for w.
func (t *redrawTimers) cancel(w widget.Widget) {
	t.mu.Lock()
	defer t.mu.Unlock()
	if _, ok := t.due[w]; !ok {
		return
	}
	delete(t.due, w)
	t.armLocked(time.Now())
}

// flush marks every widget whose time has come as needing redraw. Widgets
// removed from the tree since they asked are skipped.
func (t *redrawTimers) flush(now time.Time) {
	t.mu.Lock()
	var ready []widget.Widget
	for w, at := range t.due {
		if !at.After(now) {
			ready = append(ready, w)
			delete(t.due, w)
		}
	}
	if len(ready) > 0 {
		t.armLocked(now)
	}
	t.mu.Unlock()

	for _, w := range ready {
		if widget.IsDetached(w) {
			continue
		}
		if r, ok := w.(interface{ SetNeedsRedraw(bool) }); ok {
			r.SetNeedsRedraw(true)
		}
	}
}

// stop disarms the timer and forgets every scheduled redraw.
func (t *redrawTimers) stop() {
	t.mu.Lock()
	defer t.mu.Unlock()
	clear(t.due)
	t.disarmLocked()
}

// pending reports how many redraws are scheduled.
func (t *redrawTimers) pending() int {
	t.mu.Lock()
	defer t.mu.Unlock()
	return len(t.due)
}

// armLocked points the timer at the earliest deadline, or disarms it when
// nothing is scheduled. The timer is left alone when it already fires for
// that deadline, so a widget rescheduling every frame costs no timer churn.
func (t *redrawTimers) armLocked(now time.Time) {
	var earliest time.Time
	for _, at := range t.due {
		if earliest.IsZero() || at.Before(earliest) {
			earliest = at
		}
	}
	if earliest.IsZero() {
		t.disarmLocked()
		return
	}
	if t.timer != nil && t.armedAt.Equal(earliest) {
		return
	}
	t.disarmLocked()
	t.armedAt = earliest
	t.timer = time.AfterFunc(max(earliest.Sub(now), 0), t.wake)
}

func (t *redrawTimers) disarmLocked() {
	if t.timer != nil {
		t.timer.Stop()
		t.timer = nil
	}
	t.armedAt = time.Time{}
}
