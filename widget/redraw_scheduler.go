package widget

import "time"

// RedrawScheduler is an optional interface implemented by Context
// implementations that can repaint a widget at a later time.
//
// A widget whose appearance changes on a schedule rather than in response to
// input, such as a blinking text caret, asks to be redrawn when its next
// change is due instead of keeping the animation pump running. Between those
// times the render loop has no work for it and can stay idle.
//
// This uses the "interface extension via type assertion" pattern (same as
// AnimationScheduler and PointerCapturer) to avoid adding methods to the
// Context interface.
//
// Example usage in a widget's Draw method:
//
//	if rs, ok := ctx.(widget.RedrawScheduler); ok && rs.ScheduleRedraw(w, next) {
//	    // w is marked as needing redraw, and the render loop wakes,
//	    // once next has elapsed.
//	}
type RedrawScheduler interface {
	// ScheduleRedraw asks for w to be marked as needing redraw, and the
	// render loop woken, once d has elapsed. It replaces any redraw already
	// scheduled for w.
	//
	// It reports false when the context cannot schedule redraws; nothing is
	// scheduled then, and the caller should not depend on a later redraw.
	ScheduleRedraw(w Widget, d time.Duration) bool

	// CancelRedraw drops the redraw scheduled for w, if any.
	CancelRedraw(w Widget)
}

// WindowFocusReporter is an optional interface implemented by Context
// implementations that know whether their window has OS keyboard focus.
//
// Widgets that show keyboard-only affordances, such as a text caret, hide
// them while the window is inactive, as native text controls do.
type WindowFocusReporter interface {
	// IsWindowFocused reports whether the window has OS keyboard focus.
	IsWindowFocused() bool
}

// ScheduleRedraw implements [RedrawScheduler]. It returns false when no
// window has wired a scheduler with [ContextImpl.SetOnScheduleRedraw].
func (c *ContextImpl) ScheduleRedraw(w Widget, d time.Duration) bool {
	c.mu.RLock()
	schedule := c.onScheduleRedraw
	c.mu.RUnlock()
	if schedule == nil || w == nil {
		return false
	}
	schedule(w, d)
	return true
}

// CancelRedraw implements [RedrawScheduler].
func (c *ContextImpl) CancelRedraw(w Widget) {
	c.mu.RLock()
	cancel := c.onCancelRedraw
	c.mu.RUnlock()
	if cancel != nil && w != nil {
		cancel(w)
	}
}

// SetOnScheduleRedraw sets the callbacks behind [ContextImpl.ScheduleRedraw]
// and [ContextImpl.CancelRedraw].
func (c *ContextImpl) SetOnScheduleRedraw(schedule func(w Widget, d time.Duration), cancel func(w Widget)) {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.onScheduleRedraw = schedule
	c.onCancelRedraw = cancel
}

// IsWindowFocused implements [WindowFocusReporter]. A new context reports
// a focused window.
func (c *ContextImpl) IsWindowFocused() bool {
	c.mu.RLock()
	defer c.mu.RUnlock()
	return !c.windowUnfocused
}

// SetWindowFocused records whether the window has OS keyboard focus.
func (c *ContextImpl) SetWindowFocused(focused bool) {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.windowUnfocused = !focused
}
