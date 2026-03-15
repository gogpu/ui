package collapsible

import (
	"github.com/gogpu/ui/event"
	"github.com/gogpu/ui/geometry"
	"github.com/gogpu/ui/widget"
)

// handleEvent processes input events for the collapsible widget.
// Only the header area responds to mouse interaction.
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

// headerBounds returns the header area for hit testing.
func headerBounds(w *Widget) geometry.Rect {
	b := w.Bounds()
	return geometry.NewRect(b.Min.X, b.Min.Y, b.Width(), w.cfg.headerHeight)
}

// handleMouseEvent processes mouse events on the header area.
func handleMouseEvent(w *Widget, ctx widget.Context, e *event.MouseEvent) bool {
	hdr := headerBounds(w)

	switch e.MouseType {
	case event.MouseEnter:
		if hdr.Contains(e.Position) {
			w.istate = stateHover
			ctx.SetCursor(widget.CursorPointer)
			ctx.Invalidate()
			return true
		}
		return false // Let content handle enter

	case event.MouseMove:
		if hdr.Contains(e.Position) {
			if w.istate == stateNormal {
				w.istate = stateHover
				ctx.SetCursor(widget.CursorPointer)
				ctx.Invalidate()
			}
		} else {
			if w.istate == stateHover {
				w.istate = stateNormal
				ctx.SetCursor(widget.CursorDefault)
				ctx.Invalidate()
			}
		}
		return false // Allow propagation for content area events.

	case event.MouseLeave:
		if w.istate != stateNormal {
			w.istate = stateNormal
			ctx.SetCursor(widget.CursorDefault)
			ctx.Invalidate()
		}
		return false // Let content handle leave too

	case event.MousePress:
		if e.Button != event.ButtonLeft {
			return false
		}
		if !hdr.Contains(e.Position) {
			return false
		}
		w.istate = statePressed
		ctx.RequestFocus(w)
		ctx.Invalidate()
		return true

	case event.MouseRelease:
		if e.Button != event.ButtonLeft {
			return false
		}
		wasPressed := w.istate == statePressed
		if hdr.Contains(e.Position) {
			w.istate = stateHover
		} else {
			w.istate = stateNormal
		}
		ctx.Invalidate()
		if wasPressed && hdr.Contains(e.Position) {
			w.Toggle()
			return true
		}
		return false // Let content handle release

	default:
		return false
	}
}

// handleKeyEvent processes keyboard events for Space/Enter activation.
func handleKeyEvent(w *Widget, ctx widget.Context, e *event.KeyEvent) bool {
	if !w.IsFocused() {
		return false
	}

	if e.Key != event.KeySpace && e.Key != event.KeyEnter {
		return false
	}

	return handleActivationKey(w, ctx, e)
}

// handleActivationKey processes Space/Enter key press and release for toggling.
func handleActivationKey(w *Widget, ctx widget.Context, e *event.KeyEvent) bool {
	switch e.KeyType {
	case event.KeyPress:
		w.istate = statePressed
		ctx.Invalidate()
		return true
	case event.KeyRelease:
		wasPressed := w.istate == statePressed
		w.istate = stateNormal
		ctx.Invalidate()
		if wasPressed {
			w.Toggle()
		}
		return true
	default:
		return false
	}
}
