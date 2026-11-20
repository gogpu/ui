package chip

import (
	"github.com/gogpu/ui/event"
	"github.com/gogpu/ui/widget"
)

// handleEvent processes input events for the chip widget.
// It manages hover, press, and keyboard activation states.
func handleEvent(w *Widget, ctx widget.Context, e event.Event) bool {
	if w.cfg.ResolvedDisabled() {
		return false
	}

	switch ev := e.(type) {
	case *event.MouseEvent:
		return handleMouseEvent(w, ctx, ev)
	case *event.KeyEvent:
		return handleKeyEvent(w, ctx, ev)
	default:
		return false
	}
}

// handleMouseEvent processes mouse events for hover, press, and click.
func handleMouseEvent(w *Widget, ctx widget.Context, e *event.MouseEvent) bool {
	switch e.MouseType {
	case event.MouseEnter:
		w.state = stateHover
		ctx.SetCursor(widget.CursorPointer)
		w.SetNeedsRedraw(true)
		ctx.InvalidateRect(w.Bounds())
		return true

	case event.MouseLeave:
		w.state = stateNormal
		ctx.SetCursor(widget.CursorDefault)
		w.SetNeedsRedraw(true)
		ctx.InvalidateRect(w.Bounds())
		return true

	case event.MousePress:
		if e.Button != event.ButtonLeft {
			return false
		}
		w.state = statePressed
		ctx.RequestFocus(w)
		w.SetNeedsRedraw(true)
		ctx.InvalidateRect(w.Bounds())
		return true

	case event.MouseRelease:
		if e.Button != event.ButtonLeft {
			return false
		}
		wasPressed := w.state == statePressed
		inside := w.Bounds().Contains(e.Position)
		if inside {
			w.state = stateHover
		} else {
			w.state = stateNormal
		}
		w.SetNeedsRedraw(true)
		ctx.InvalidateRect(w.Bounds())
		if wasPressed && inside {
			activate(w)
		}
		return true

	default:
		return false
	}
}

// handleKeyEvent processes keyboard events for Enter/Space activation.
func handleKeyEvent(w *Widget, ctx widget.Context, e *event.KeyEvent) bool {
	if !w.IsFocused() {
		return false
	}
	switch e.Key {
	case event.KeyEnter, event.KeySpace:
		return handleActivationKey(w, ctx, e)
	default:
		return false
	}
}

// handleActivationKey processes Enter/Space press and release.
//
// Both edges change the pressed visual with no mouse event in between, so
// each one invalidates exactly like the mouse path does.
func handleActivationKey(w *Widget, ctx widget.Context, e *event.KeyEvent) bool {
	switch e.KeyType {
	case event.KeyPress:
		w.state = statePressed
		invalidate(w, ctx)
		return true
	case event.KeyRelease:
		wasPressed := w.state == statePressed
		w.state = stateNormal
		invalidate(w, ctx)
		if wasPressed {
			activate(w)
		}
		return true
	default:
		return false
	}
}

// invalidate schedules a repaint of the chip after a visual state change.
func invalidate(w *Widget, ctx widget.Context) {
	w.SetNeedsRedraw(true)
	ctx.InvalidateRect(w.Bounds())
}

// activate runs the chip's activation logic: toggling selection for filter
// chips and invoking the configured handlers.
func activate(w *Widget) {
	if w.cfg.selectable {
		newSel := !w.cfg.ResolvedSelected()
		// Two-way: write back to the bound signal (or static field) so the
		// chip's rendered state stays in sync after the click.
		w.applySelected(newSel)
		if w.cfg.onSelectedChanged != nil {
			w.cfg.onSelectedChanged(newSel)
		}
	}
	if w.cfg.onClick != nil {
		w.cfg.onClick()
	}
	w.SetNeedsRedraw(true)
}
