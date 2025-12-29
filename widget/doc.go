// Package widget provides core widget types and interfaces for the gogpu/ui toolkit.
//
// This package defines the fundamental building blocks for creating user interfaces:
// the Widget interface, WidgetBase helper struct, Context for accessing UI state,
// and Canvas for drawing operations.
//
// # Core Types
//
//   - [Widget]: Interface that all UI elements must implement
//   - [WidgetBase]: Embeddable struct providing common widget functionality
//   - [Context]: Interface for accessing UI state during layout/draw/event handling
//   - [Canvas]: Interface for drawing operations (placeholder for render package)
//
// # Widget Interface
//
// The Widget interface is the foundation of the UI framework. Every UI element
// implements this interface to participate in layout, drawing, and event handling:
//
//	type Widget interface {
//	    Layout(ctx Context, constraints Constraints) Size
//	    Draw(ctx Context, canvas Canvas)
//	    Event(ctx Context, e event.Event) bool
//	    Children() []Widget
//	}
//
// # Using WidgetBase
//
// WidgetBase provides common functionality that most widgets need. Embed it
// in your custom widget implementations:
//
//	type MyButton struct {
//	    widget.WidgetBase
//	    label string
//	}
//
//	func NewMyButton(label string) *MyButton {
//	    b := &MyButton{label: label}
//	    b.SetVisible(true)
//	    b.SetEnabled(true)
//	    return b
//	}
//
// # Layout System
//
// Layout follows Flutter's box constraints model:
//
//  1. Parent passes constraints to child via Layout()
//  2. Child calculates its size within those constraints
//  3. Child returns its chosen size to parent
//  4. Parent positions child (sets bounds via SetBounds())
//
// Example layout implementation:
//
//	func (w *MyWidget) Layout(ctx widget.Context, c geometry.Constraints) geometry.Size {
//	    // Calculate preferred size
//	    preferred := geometry.Sz(100, 50)
//	    // Constrain to allowed range
//	    return c.Constrain(preferred)
//	}
//
// # Event Handling
//
// Events flow through the widget tree. The Event method returns true if
// the event was consumed:
//
//	func (w *MyWidget) Event(ctx widget.Context, e event.Event) bool {
//	    switch ev := e.(type) {
//	    case *event.MouseEvent:
//	        if ev.MouseType == event.MousePress {
//	            w.handleClick(ev.Position)
//	            return true
//	        }
//	    }
//	    return false
//	}
//
// # Thread Safety
//
// Widgets are NOT thread-safe. All widget operations must occur on the
// main/UI thread. The Context interface provides the only safe way to
// communicate with the UI system from callbacks.
//
// # Design Principles
//
//   - Composition over inheritance: Embed WidgetBase, don't subclass
//   - Explicit state: Widget state is explicit, not implicit
//   - Zero-allocation hot paths: Layout and draw should not allocate
//   - Interface-based: Depend on interfaces, not concrete types
package widget
