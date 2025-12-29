package widget

import (
	"github.com/gogpu/ui/event"
	"github.com/gogpu/ui/geometry"
)

// Widget is the fundamental building block of the UI framework.
//
// All UI elements implement this interface to participate in layout,
// drawing, and event handling. Widgets form a tree structure where
// parent widgets contain and manage child widgets.
//
// The widget lifecycle consists of three phases:
//
//  1. Layout: Calculate size given constraints from parent
//  2. Draw: Render the widget to a canvas
//  3. Event: Handle user input events
//
// Implementations should embed [WidgetBase] to get common functionality
// like bounds tracking, visibility, and enabled state management.
type Widget interface {
	// Layout calculates the widget's size given constraints from the parent.
	//
	// The constraints define the minimum and maximum allowed dimensions.
	// The returned size must satisfy the constraints (within min/max bounds).
	//
	// Layout is called during the layout phase, before Draw. The widget
	// should calculate its preferred size and return it constrained to
	// the allowed range.
	//
	// For container widgets, Layout should:
	//  1. Layout all children with appropriate constraints
	//  2. Position children by setting their bounds
	//  3. Return the container's total size
	//
	// Example:
	//
	//	func (w *MyWidget) Layout(ctx Context, c geometry.Constraints) geometry.Size {
	//	    preferred := geometry.Sz(100, 50)
	//	    return c.Constrain(preferred)
	//	}
	Layout(ctx Context, constraints geometry.Constraints) geometry.Size

	// Draw renders the widget to the canvas.
	//
	// Draw is called after Layout, when the widget's bounds are established.
	// The canvas provides drawing operations for rendering the widget.
	//
	// For container widgets, Draw should:
	//  1. Draw the container's own content (background, border, etc.)
	//  2. Draw all visible children
	//
	// Example:
	//
	//	func (w *MyWidget) Draw(ctx Context, canvas Canvas) {
	//	    canvas.DrawRect(w.Bounds(), w.backgroundColor)
	//	    for _, child := range w.Children() {
	//	        child.Draw(ctx, canvas)
	//	    }
	//	}
	Draw(ctx Context, canvas Canvas)

	// Event handles an input event and returns true if consumed.
	//
	// Events are dispatched from the root widget down through the tree.
	// A widget that handles an event should return true to prevent
	// further propagation.
	//
	// For container widgets, Event should:
	//  1. Check if event is within bounds
	//  2. Dispatch to appropriate child widgets
	//  3. Handle the event if no child consumed it
	//
	// Example:
	//
	//	func (w *MyWidget) Event(ctx Context, e event.Event) bool {
	//	    if me, ok := e.(*event.MouseEvent); ok {
	//	        if !w.Bounds().Contains(me.Position) {
	//	            return false
	//	        }
	//	        if me.MouseType == event.MousePress {
	//	            w.onClick()
	//	            return true
	//	        }
	//	    }
	//	    return false
	//	}
	Event(ctx Context, e event.Event) bool

	// Children returns the widget's child widgets.
	//
	// Leaf widgets (like labels, buttons) return nil.
	// Container widgets return their children in z-order (bottom to top).
	//
	// The returned slice should not be modified by the caller.
	Children() []Widget
}

// LayoutFunc is a function type for custom layout logic.
//
// This can be used to implement layout behavior without creating
// a full widget implementation.
type LayoutFunc func(ctx Context, constraints geometry.Constraints) geometry.Size

// DrawFunc is a function type for custom drawing logic.
type DrawFunc func(ctx Context, canvas Canvas)

// EventFunc is a function type for custom event handling.
type EventFunc func(ctx Context, e event.Event) bool
