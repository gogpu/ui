package toggleswitch

import (
	"github.com/gogpu/ui/event"
	"github.com/gogpu/ui/widget"
)

// dragSlop is how far the pointer must move sideways before a press turns
// into a thumb drag.
const dragSlop float32 = 4

// handleEvent processes input events for the switch.
func handleEvent(w *Widget, ctx widget.Context, e event.Event) bool {
	switch ev := e.(type) {
	case *event.MouseEvent:
		return handleMouseEvent(w, ctx, ev)
	case *event.KeyEvent:
		return handleKeyEvent(w, ctx, ev)
	default:
		return false
	}
}

// handleMouseEvent tracks hover, press, drag, and release.
func handleMouseEvent(w *Widget, ctx widget.Context, e *event.MouseEvent) bool {
	switch e.MouseType {
	case event.MouseEnter:
		// Hover is tracked while disabled too, so re-enabling a switch under
		// the pointer shows the highlight without waiting for a new enter.
		w.hovered = true
		if w.isDisabled() {
			return false
		}
		ctx.SetCursor(widget.CursorPointer)
		invalidate(w, ctx)
		return true

	case event.MouseLeave:
		w.hovered = false
		if w.isDisabled() {
			return false
		}
		if !w.pointerDown {
			ctx.SetCursor(widget.CursorDefault)
		}
		invalidate(w, ctx)
		return true

	case event.MousePress:
		if e.Button != event.ButtonLeft || w.isDisabled() {
			return false
		}
		handlePress(w, ctx, e)
		return true

	case event.MouseMove:
		return handleMove(w, ctx, e)

	case event.MouseRelease:
		if e.Button != event.ButtonLeft {
			return false
		}
		return handleRelease(w, ctx, e)

	default:
		return false
	}
}

// handlePress starts a press that may become a click or a thumb drag.
func handlePress(w *Widget, ctx widget.Context, e *event.MouseEvent) {
	w.pointerDown = true
	w.dragging = false
	w.pressX = e.Position.X
	w.pressPos = w.thumbPosition(ctx.Now())
	w.dragPos = w.pressPos
	w.lastPointer = e.Position

	ctx.RequestFocus(w)
	// ADR-031: keep receiving moves and the release while the pointer is
	// outside the switch during a drag.
	if pc, ok := ctx.(widget.PointerCapturer); ok {
		pc.CapturePointer(w)
	}
	invalidate(w, ctx)
}

// handleMove drags the thumb once the pointer has moved past the slop.
func handleMove(w *Widget, ctx widget.Context, e *event.MouseEvent) bool {
	if !w.pointerDown {
		return false
	}
	if w.isDisabled() {
		// Disabled mid-press: drop the gesture without changing the value.
		w.pointerDown = false
		w.dragging = false
		return false
	}
	w.lastPointer = e.Position

	dx := e.Position.X - w.pressX
	if !w.dragging {
		if dx < dragSlop && dx > -dragSlop {
			return true
		}
		w.dragging = true
	}

	pos := w.pressPos
	if travel := resolveLayoutMetrics(w.painter).SwitchThumbTravel(); travel > 0 {
		pos += dx / travel
	}
	w.dragPos = clampUnit(pos)
	// Show the drag position right away, so that a release before the next
	// draw settles from where the thumb was dragged.
	w.position.jump(w.dragPos)
	ctx.SetCursor(widget.CursorPointer)
	invalidate(w, ctx)
	return true
}

// handleRelease finishes a press: a click toggles when released inside the
// switch, and a drag leaves the switch on the side the thumb is closer to.
func handleRelease(w *Widget, ctx widget.Context, e *event.MouseEvent) bool {
	if !w.pointerDown {
		return false
	}
	wasDragging := w.dragging
	w.pointerDown = false
	w.dragging = false

	if pc, ok := ctx.(widget.PointerCapturer); ok {
		pc.ReleasePointer(w)
	}

	inside := w.Bounds().Contains(e.Position)
	if !inside {
		w.hovered = false
		ctx.SetCursor(widget.CursorDefault)
	}
	invalidate(w, ctx)

	if w.isDisabled() {
		return true
	}
	if wasDragging {
		// A real release arrives where the pointer was last seen, even when
		// a drag overshoots the switch. The Window cancels a gesture (focus
		// loss, PointerCancel) with a release at a point outside the widget
		// that no move reported; that one must not change the value.
		if (inside || e.Position == w.lastPointer) && (w.dragPos >= 0.5) != w.cfg.ResolvedChecked() {
			fireToggle(w)
		}
		return true
	}
	if inside {
		fireToggle(w)
	}
	return true
}

// handleKeyEvent toggles a focused switch when Space is released.
func handleKeyEvent(w *Widget, ctx widget.Context, e *event.KeyEvent) bool {
	if !w.IsFocused() || w.isDisabled() {
		w.keyDown = false
		return false
	}
	if e.Key != event.KeySpace {
		return false
	}

	switch e.KeyType {
	case event.KeyPress:
		w.keyDown = true
		invalidate(w, ctx)
		return true
	case event.KeyRepeat:
		return w.keyDown
	case event.KeyRelease:
		wasDown := w.keyDown
		w.keyDown = false
		invalidate(w, ctx)
		if wasDown {
			fireToggle(w)
		}
		return wasDown
	default:
		return false
	}
}

// fireToggle plays the click sound, flips the value, and notifies OnToggle.
func fireToggle(w *Widget) {
	widget.PlaySound(widget.SoundClick)
	next := !w.cfg.ResolvedChecked()
	// TWO-WAY: a bound CheckedSignal receives the new value. With CheckedFn
	// the caller owns the state and only hears about the toggle.
	if w.cfg.checkedSignal != nil {
		w.cfg.checkedSignal.Set(next)
	} else if w.cfg.checkedFn == nil {
		w.cfg.checked = next
	}
	if w.cfg.onToggle != nil {
		w.cfg.onToggle(next)
	}
}

// invalidate schedules a redraw of the switch.
func invalidate(w *Widget, ctx widget.Context) {
	w.SetNeedsRedraw(true)
	ctx.InvalidateRect(w.Bounds())
}
