package checkbox

import (
	"github.com/gogpu/ui/event"
	"github.com/gogpu/ui/widget"
)

// handleEvent processes input events for the checkbox widget.
// It manages hover, press, and keyboard activation states.
func handleEvent(w *Widget, ctx widget.Context, e event.Event) bool {
	// Disabled checkboxes ignore all interaction.
	if w.cfg.ResolvedDisabled() {
		return false
	}

	switch ev := e.(type) {
	case *event.MouseEvent:
		return handleMouseEvent(w, ctx, ev)
	case *event.KeyEvent:
		return handleKeyEvent(w, ev)
	default:
		return false
	}
}

// handleMouseEvent processes mouse events for hover, press, and toggle.
func handleMouseEvent(w *Widget, ctx widget.Context, e *event.MouseEvent) bool {
	switch e.MouseType {
	case event.MouseEnter:
		w.state = stateHover
		ctx.SetCursor(widget.CursorPointer)
		ctx.Invalidate()
		return true

	case event.MouseLeave:
		w.state = stateNormal
		ctx.SetCursor(widget.CursorDefault)
		ctx.Invalidate()
		return true

	case event.MousePress:
		if e.Button != event.ButtonLeft {
			return false
		}
		w.state = statePressed
		ctx.RequestFocus(w)
		ctx.Invalidate()
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
		ctx.Invalidate()
		if wasPressed && w.Bounds().Contains(e.Position) {
			fireToggle(w)
		}
		return true

	default:
		return false
	}
}

// handleKeyEvent processes keyboard events for Space activation.
func handleKeyEvent(w *Widget, e *event.KeyEvent) bool {
	if !w.IsFocused() {
		return false
	}

	if e.Key != event.KeySpace {
		return false
	}

	return handleActivationKey(w, e)
}

// handleActivationKey processes Space key press and release for checkbox toggling.
func handleActivationKey(w *Widget, e *event.KeyEvent) bool {
	switch e.KeyType {
	case event.KeyPress:
		w.state = statePressed
		return true
	case event.KeyRelease:
		wasPressed := w.state == statePressed
		w.state = stateNormal
		if wasPressed {
			fireToggle(w)
		}
		return true
	default:
		return false
	}
}

// fireToggle toggles the checked state and calls the configured onToggle handler.
func fireToggle(w *Widget) {
	newChecked := !w.cfg.ResolvedChecked()
	// TWO-WAY: if a CheckedSignal is bound, write back the new state.
	if w.cfg.checkedSignal != nil {
		w.cfg.checkedSignal.Set(newChecked)
	} else if w.cfg.checkedFn == nil {
		// Update the static checked field so that subsequent reads reflect the toggle,
		// unless a dynamic CheckedFn is in use (in which case the caller manages state).
		w.cfg.checked = newChecked
	}
	// Clear indeterminate on toggle.
	w.cfg.indeterminate = false
	if w.cfg.onToggle != nil {
		w.cfg.onToggle(newChecked)
	}
}
