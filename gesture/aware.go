package gesture

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
// Example:
//
//	type MyButton struct {
//	    widget.WidgetBase
//	    click *gesture.ClickRecognizer
//	}
//
//	func (b *MyButton) GestureRecognizers() []gesture.Recognizer {
//	    return []gesture.Recognizer{b.click}
//	}
type GestureAware interface {
	// GestureRecognizers returns the gesture recognizers owned by this widget.
	// Called during PointerDown hit-testing. The returned recognizers are
	// added to the gesture arena for the pointer that triggered the event.
	//
	// Implementations should return the same recognizer instances across
	// calls (created once in the constructor or Mount), not new instances
	// each time. The arena manages recognizer lifecycle per-pointer.
	GestureRecognizers() []Recognizer
}
