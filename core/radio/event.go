package radio

import (
	"github.com/gogpu/ui/event"
	"github.com/gogpu/ui/widget"
)

// handleItemEvent processes input events for a single radio item.
// It manages hover, press, and keyboard activation states.
func handleItemEvent(it *Item, ctx widget.Context, e event.Event) bool {
	// Disabled groups ignore all interaction.
	if it.group.cfg.ResolvedDisabled() {
		return false
	}

	switch ev := e.(type) {
	case *event.MouseEvent:
		return handleItemMouseEvent(it, ctx, ev)
	case *event.KeyEvent:
		return handleItemKeyEvent(it, ctx, ev)
	default:
		return false
	}
}

// handleItemMouseEvent processes mouse events for hover, press, and selection.
func handleItemMouseEvent(it *Item, ctx widget.Context, e *event.MouseEvent) bool {
	switch e.MouseType {
	case event.MouseEnter:
		it.state = stateHover
		ctx.SetCursor(widget.CursorPointer)
		ctx.Invalidate()
		return true

	case event.MouseLeave:
		it.state = stateNormal
		ctx.SetCursor(widget.CursorDefault)
		ctx.Invalidate()
		return true

	case event.MousePress:
		if e.Button != event.ButtonLeft {
			return false
		}
		it.state = statePressed
		ctx.RequestFocus(it)
		ctx.Invalidate()
		return true

	case event.MouseRelease:
		if e.Button != event.ButtonLeft {
			return false
		}
		wasPressed := it.state == statePressed
		// Check if release is inside bounds.
		if it.Bounds().Contains(e.Position) {
			it.state = stateHover
		} else {
			it.state = stateNormal
		}
		ctx.Invalidate()
		if wasPressed && it.Bounds().Contains(e.Position) {
			it.group.selectValue(it.value)
		}
		return true

	default:
		return false
	}
}

// handleItemKeyEvent processes keyboard events for item activation and group navigation.
func handleItemKeyEvent(it *Item, ctx widget.Context, e *event.KeyEvent) bool {
	if !it.IsFocused() {
		return false
	}

	switch e.Key {
	case event.KeySpace, event.KeyEnter:
		return handleActivationKey(it, ctx, e)
	case event.KeyUp, event.KeyDown, event.KeyLeft, event.KeyRight:
		return handleNavigationKey(it, ctx, e)
	default:
		return false
	}
}

// handleActivationKey processes Space/Enter key press and release for item selection.
func handleActivationKey(it *Item, ctx widget.Context, e *event.KeyEvent) bool {
	switch e.KeyType {
	case event.KeyPress:
		it.state = statePressed
		ctx.Invalidate()
		return true
	case event.KeyRelease:
		wasPressed := it.state == statePressed
		it.state = stateNormal
		ctx.Invalidate()
		if wasPressed {
			it.group.selectValue(it.value)
		}
		return true
	default:
		return false
	}
}

// handleNavigationKey processes arrow keys to move focus between radio items.
func handleNavigationKey(it *Item, ctx widget.Context, e *event.KeyEvent) bool {
	// Only act on key press, not release.
	if e.KeyType != event.KeyPress {
		return true // consume release to prevent bubbling
	}

	dir := it.group.cfg.direction
	var delta int

	switch {
	case dir == Vertical && e.Key == event.KeyUp:
		delta = -1
	case dir == Vertical && e.Key == event.KeyDown:
		delta = 1
	case dir == Horizontal && e.Key == event.KeyLeft:
		delta = -1
	case dir == Horizontal && e.Key == event.KeyRight:
		delta = 1
	default:
		return false
	}

	it.group.moveFocus(it, ctx, delta)
	return true
}
