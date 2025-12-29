// Package event provides input event types for the gogpu/ui toolkit.
//
// This package defines the core event types used for handling user input
// throughout the UI framework: mouse events, keyboard events, focus events,
// wheel (scroll) events, and modifier key states.
//
// # Event Types
//
//   - [Event]: Base interface implemented by all event types
//   - [MouseEvent]: Mouse button presses, releases, moves, enter/leave
//   - [KeyEvent]: Keyboard key presses, releases, and repeats
//   - [FocusEvent]: Widget focus gained or lost
//   - [WheelEvent]: Mouse wheel scroll events
//
// # Event Flow
//
// Events flow through the widget tree from root to target:
//
//  1. Event is received from the platform layer
//  2. Event is dispatched to the root widget
//  3. Root widget propagates event to children
//  4. Target widget handles event and may mark it as handled
//  5. If not handled, event bubbles back up to ancestors
//
// # Modifiers
//
// The [Modifiers] type is a bitmask representing which modifier keys
// (Shift, Ctrl, Alt, Super/Cmd) are held during an event:
//
//	if event.Modifiers().Has(ModCtrl | ModShift) {
//	    // Ctrl+Shift combination
//	}
//
// # Usage Example
//
//	func (w *MyWidget) HandleEvent(e event.Event) bool {
//	    switch ev := e.(type) {
//	    case *event.MouseEvent:
//	        if ev.Type == event.MousePress && ev.Button == event.ButtonLeft {
//	            w.onClick(ev.Position)
//	            ev.SetHandled()
//	            return true
//	        }
//	    case *event.KeyEvent:
//	        if ev.Type == event.KeyPress && ev.Key == event.KeyEnter {
//	            w.onSubmit()
//	            ev.SetHandled()
//	            return true
//	        }
//	    }
//	    return false
//	}
//
// # Thread Safety
//
// Event objects are NOT safe for concurrent access. Events are processed
// sequentially on the main/UI thread. Do not store events or access them
// from other goroutines.
//
// # Design Principles
//
//   - All coordinates use float32 for GPU compatibility
//   - Positions are represented using [geometry.Point]
//   - Events are mutable (handled state can be changed)
//   - Time stamps use [time.Time] for precision
package event
