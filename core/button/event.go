package button

import (
	"github.com/gogpu/ui/event"
	"github.com/gogpu/ui/widget"
)

// handleEvent processes input events for the button widget.
// It manages hover, press, and keyboard activation states.
func handleEvent(w *Widget, ctx widget.Context, e event.Event) bool {
	// Disabled buttons ignore all interaction.
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
		// Check if release is inside bounds.
		if w.Bounds().Contains(e.Position) {
			w.state = stateHover
		} else {
			w.state = stateNormal
		}
		w.SetNeedsRedraw(true)
		ctx.InvalidateRect(w.Bounds())
		if wasPressed && w.Bounds().Contains(e.Position) {
			fireOnClick(w)
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

// handleActivationKey processes Enter/Space key press and release for button activation.
//
// Both edges change the pressed visual, and nothing else repaints the button
// in between (no mouse event, no signal), so each one invalidates exactly
// like the mouse path does.
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
			fireOnClick(w)
		}
		return true
	default:
		return false
	}
}

// invalidate schedules a repaint of the button after a visual state change.
func invalidate(w *Widget, ctx widget.Context) {
	w.SetNeedsRedraw(true)
	ctx.InvalidateRect(w.Bounds())
}

// fireOnClick plays the click sound and calls the configured onClick handler.
func fireOnClick(w *Widget) {
	widget.PlaySound(widget.SoundClick)
	if w.cfg.onClick != nil {
		w.cfg.onClick()
	}
}
