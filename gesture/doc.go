// Package gesture implements a gesture recognition system for gogpu/ui.
//
// The gesture package provides infrastructure for recognizing user input
// patterns such as clicks, drags, long presses, and combined tap-and-drag
// sequences from a stream of pointer events. It sits at the infrastructure
// layer alongside focus/, overlay/, state/, and animation/.
//
// # Architecture
//
// The system is based on Flutter's GestureArena protocol (source-verified):
//
//   - [Arena] manages gesture disambiguation for a single window.
//   - [Recognizer] is the interface for all gesture recognizers.
//   - [RecognizerBase] provides shared tracking logic.
//   - Concrete recognizers ([ClickRecognizer], [DragRecognizer],
//     [LongPressRecognizer], [TapAndDragRecognizer]) implement specific
//     gesture patterns.
//
// # Arena Protocol
//
// When a PointerDown event occurs, all interested recognizers register in the
// arena for that pointer ID. As pointer events arrive, recognizers evaluate
// whether the gesture matches their pattern. A recognizer calls
// Arena.Resolve(Accepted) to claim victory or Resolve(Rejected) to withdraw.
//
//   - Resolve(Accepted) while arena is open: stored as eager winner.
//   - Resolve(Rejected): member removed, RejectGesture called.
//   - Arena closes (end of PointerDown dispatch): if 1 member remains, it wins.
//   - Sweep (after PointerUp): first remaining member wins.
//   - Hold/Release: defers sweep (used by multi-tap between taps).
//
// # Dependency Rules
//
// gesture/ imports only Layer 1 packages (event/, geometry/) and infrastructure
// (state/). It does NOT import widget/, core/, app/, theme/, or any external
// rendering libraries.
//
// # Signals Integration
//
// Gesture recognizers support opt-in reactive signals via functional options.
// For example, [WithDraggingSignal] binds a [state.Signal] to the drag state.
// Signals use equality suppression for bool values to prevent redundant
// notifications.
package gesture
