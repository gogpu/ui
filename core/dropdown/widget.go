package dropdown

import (
	"github.com/gogpu/ui/event"
	"github.com/gogpu/ui/geometry"
	"github.com/gogpu/ui/overlay"
	"github.com/gogpu/ui/state"
	"github.com/gogpu/ui/widget"
)

// triggerHeight is the default height of the dropdown trigger.
const triggerHeight float32 = 48

// interactionState represents the current user interaction state.
type interactionState uint8

const (
	stateNormal  interactionState = iota
	stateHover                    // mouse is over the trigger
	statePressed                  // mouse button is held down
)

// Widget implements a dropdown/select widget with a trigger element and a
// floating menu overlay. It supports keyboard navigation, mouse selection,
// and mouse wheel scrolling within the menu.
//
// Create with [New] using functional options.
type Widget struct {
	widget.WidgetBase

	cfg           config
	painter       Painter
	state         interactionState
	open          bool
	selectedIndex int
	menuWidget    *menuWidget // active menu widget (nil when closed)
}

// New creates a new dropdown Widget with the given options.
//
// The returned widget is visible, enabled, and focusable by default.
func New(opts ...Option) *Widget {
	w := &Widget{
		selectedIndex: -1,
		painter:       DefaultPainter{},
	}
	w.cfg.maxVisibleItems = defaultMaxVisible
	w.cfg.selectedIndex = -1

	w.SetVisible(true)
	w.SetEnabled(true)

	for _, opt := range opts {
		opt(&w.cfg)
	}

	w.selectedIndex = w.cfg.selectedIndex

	if w.cfg.painter != nil {
		w.painter = w.cfg.painter
	}

	// If a signal is provided, initialize from it.
	if w.cfg.signal != nil {
		w.selectedIndex = w.cfg.signal.Get()
	}

	return w
}

// IsFocusable reports whether the dropdown can currently receive focus.
func (w *Widget) IsFocusable() bool {
	return w.IsVisible() && w.IsEnabled() && !w.cfg.ResolvedDisabled()
}

// Layout calculates the dropdown trigger's preferred size.
func (w *Widget) Layout(_ widget.Context, constraints geometry.Constraints) geometry.Size {
	preferred := geometry.Sz(200, triggerHeight)
	return constraints.Constrain(preferred)
}

// Draw renders the dropdown trigger.
func (w *Widget) Draw(ctx widget.Context, canvas widget.Canvas) {
	if canvas == nil {
		return
	}

	text := w.cfg.placeholder
	isPlaceholder := true
	if w.selectedIndex >= 0 && w.selectedIndex < len(w.cfg.items) {
		text = w.cfg.items[w.selectedIndex].DisplayText()
		isPlaceholder = false
	}
	if text == "" && isPlaceholder {
		text = "Select..."
	}

	w.painter.PaintTrigger(canvas, &TriggerPaintState{
		Bounds:        w.Bounds(),
		SelectedText:  text,
		IsPlaceholder: isPlaceholder,
		Open:          w.open,
		Focused:       w.IsFocused(),
		Hovered:       w.state == stateHover,
		Disabled:      w.cfg.ResolvedDisabled(),
	})
}

// Event handles input events for the dropdown trigger.
func (w *Widget) Event(ctx widget.Context, e event.Event) bool {
	if w.cfg.ResolvedDisabled() {
		return false
	}

	switch ev := e.(type) {
	case *event.MouseEvent:
		return w.handleMouseEvent(ctx, ev)
	case *event.KeyEvent:
		return w.handleKeyEvent(ctx, ev)
	default:
		return false
	}
}

// Children returns nil; the dropdown trigger is a leaf widget.
// The menu is rendered in the overlay, not as a child.
func (w *Widget) Children() []widget.Widget {
	return nil
}

// SelectedIndex returns the currently selected item index, or -1 if none.
func (w *Widget) SelectedIndex() int {
	return w.selectedIndex
}

// SelectedValue returns the value of the currently selected item,
// or an empty string if nothing is selected.
func (w *Widget) SelectedValue() string {
	if w.selectedIndex < 0 || w.selectedIndex >= len(w.cfg.items) {
		return ""
	}
	return w.cfg.items[w.selectedIndex].Value
}

// SetSelectedIndex programmatically sets the selected item.
func (w *Widget) SetSelectedIndex(index int) {
	if index < -1 || index >= len(w.cfg.items) {
		return
	}
	w.selectedIndex = index
	if w.cfg.signal != nil {
		w.cfg.signal.Set(index)
	}
}

// IsOpen returns true if the dropdown menu is currently visible.
func (w *Widget) IsOpen() bool {
	return w.open
}

// Open opens the dropdown menu overlay.
func (w *Widget) Open(ctx widget.Context) {
	if w.open || w.cfg.ResolvedDisabled() {
		return
	}

	om := ctx.OverlayManager()
	if om == nil {
		return
	}

	w.open = true

	// Create the menu widget.
	w.menuWidget = newMenuWidget(
		w.cfg.items,
		w.selectedIndex,
		w.cfg.maxVisibleItems,
		w.painter,
		func(index int) {
			w.selectItem(ctx, index)
		},
	)

	// Position the menu below the trigger using screen-space bounds.
	// ScreenBounds accounts for all parent transforms (scroll offsets, etc.).
	triggerBounds := w.ScreenBounds()
	windowSize := ctx.WindowSize()
	menuSize := geometry.Sz(triggerBounds.Width(), float32(w.menuWidget.visibleCount())*w.menuWidget.itemHeight)
	pos := overlay.Position(overlay.PlacementBelow, triggerBounds, menuSize, windowSize, 2)

	w.menuWidget.SetBounds(geometry.FromPointSize(pos, menuSize))

	// Push to overlay stack.
	om.PushOverlay(w.menuWidget, func() {
		w.close(ctx)
	})

	ctx.Invalidate()
}

// Close closes the dropdown menu overlay.
func (w *Widget) Close(ctx widget.Context) {
	w.close(ctx)
}

// close is the internal close implementation.
func (w *Widget) close(ctx widget.Context) {
	if !w.open {
		return
	}

	w.open = false

	if w.menuWidget != nil {
		om := ctx.OverlayManager()
		if om != nil {
			om.RemoveOverlay(w.menuWidget)
		}
		w.menuWidget = nil
	}

	ctx.Invalidate()
}

// selectItem is called when an item is selected from the menu.
func (w *Widget) selectItem(ctx widget.Context, index int) {
	if index < 0 || index >= len(w.cfg.items) {
		return
	}

	w.selectedIndex = index

	// Update signal if bound.
	if w.cfg.signal != nil {
		w.cfg.signal.Set(index)
	}

	// Fire callback.
	if w.cfg.onChange != nil {
		w.cfg.onChange(index, w.cfg.items[index].Value)
	}

	w.close(ctx)
}

// handleMouseEvent processes mouse events on the trigger.
func (w *Widget) handleMouseEvent(ctx widget.Context, e *event.MouseEvent) bool {
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
		if w.Bounds().Contains(e.Position) {
			w.state = stateHover
		} else {
			w.state = stateNormal
		}
		if wasPressed && w.Bounds().Contains(e.Position) {
			w.toggle(ctx)
		}
		ctx.Invalidate()
		return true

	default:
		return false
	}
}

// handleKeyEvent processes keyboard events on the trigger.
func (w *Widget) handleKeyEvent(ctx widget.Context, e *event.KeyEvent) bool {
	if !w.IsFocused() {
		return false
	}

	if e.KeyType != event.KeyPress && e.KeyType != event.KeyRepeat {
		return false
	}

	switch e.Key {
	case event.KeyEnter, event.KeySpace:
		w.toggle(ctx)
		return true
	case event.KeyDown:
		if !w.open {
			w.Open(ctx)
		}
		return true
	case event.KeyUp:
		if !w.open {
			w.Open(ctx)
		}
		return true
	case event.KeyEscape:
		if w.open {
			w.close(ctx)
			return true
		}
		return false
	default:
		return false
	}
}

// toggle opens the menu if closed, closes it if open.
func (w *Widget) toggle(ctx widget.Context) {
	if w.open {
		w.close(ctx)
	} else {
		w.Open(ctx)
	}
}

// Mount creates signal bindings for push-based invalidation.
// Implements [widget.Lifecycle].
func (w *Widget) Mount(ctx widget.Context) {
	sched := ctx.Scheduler()
	if sched == nil {
		return
	}
	if w.cfg.signal != nil {
		b := state.BindToScheduler(w.cfg.signal, w, sched)
		w.AddBinding(b)
	}
}

// Unmount is called when the dropdown is removed from the widget tree.
// Implements [widget.Lifecycle].
func (w *Widget) Unmount() {
	// Bindings are cleaned up automatically by WidgetBase.CleanupBindings().
}

// Compile-time interface checks.
var (
	_ widget.Widget    = (*Widget)(nil)
	_ widget.Focusable = (*Widget)(nil)
	_ widget.Lifecycle = (*Widget)(nil)
)
