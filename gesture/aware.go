package gesture

import "github.com/gogpu/ui/geometry"

// GestureAware is an optional interface implemented by widgets that
// participate in the gesture recognition system.
//
// During PointerDown hit-testing, the Window checks each widget in the
// hit-test path for GestureAware. Widgets that implement it have their
// recognizers registered in the gesture arena for that pointer.
//
// Widgets that do not implement GestureAware continue to receive events
// through the existing Event(ctx, event.Event) path unchanged. This is
// the same opt-in pattern used by [widget.Focusable] and
// [widget.RepaintBoundaryMarker].
//
// GestureHitTest receives the pointer position in widget-local coordinates.
// This allows container widgets with partial interactive regions (e.g., a
// Collapsible header or TabView tab strip) to return recognizers ONLY when
// the pointer is within the interactive region. Child widgets' recognizers
// are then the sole participants in the gesture arena, preventing parent
// containers from consuming clicks meant for children.
//
// Example — leaf widget (always returns recognizers):
//
//	func (b *MyButton) GestureHitTest(_ geometry.Point) []gesture.Recognizer {
//	    return []gesture.Recognizer{b.click}
//	}
//
// Example — container widget with interactive header:
//
//	func (c *Collapsible) GestureHitTest(pos geometry.Point) []gesture.Recognizer {
//	    if !c.headerBounds().Contains(pos) {
//	        return nil // let children handle the gesture
//	    }
//	    return []gesture.Recognizer{c.click}
//	}
type GestureAware interface {
	// GestureHitTest returns the gesture recognizers that should participate
	// in the gesture arena for a pointer event at the given position.
	//
	// pos is in widget-local coordinates (relative to the widget's own
	// origin). The hit-test framework translates window coordinates to
	// widget-local space before calling this method.
	//
	// Leaf widgets (Button, Checkbox, etc.) should always return their
	// recognizers — the framework already confirmed the point is within
	// the widget's bounds.
	//
	// Container widgets with partial interactive areas (Collapsible header,
	// TabView tab strip, Docking zone tabs) MUST check whether pos falls
	// within their interactive region before returning recognizers. If pos
	// is outside the interactive region, return nil to let child widgets'
	// recognizers be the sole participants in the arena.
	//
	// Implementations should return the same recognizer instances across
	// calls (created once in the constructor or Mount), not new instances
	// each time. The arena manages recognizer lifecycle per-pointer.
	GestureHitTest(pos geometry.Point) []Recognizer
}
