// Package splitview provides a resizable split panel container widget
// with a draggable divider separating two child panels.
//
// Construction uses functional options for immutable configuration:
//
//	split := splitview.New(
//	    splitview.First(leftPanel),
//	    splitview.Second(rightPanel),
//	    splitview.OrientationOpt(splitview.Horizontal),
//	    splitview.InitialRatio(0.3),
//	    splitview.MinFirst(200),
//	)
//
// # Orientation
//
// Two orientations are available:
//   - [Horizontal] (default) -- first panel on the left, second on the right
//   - [Vertical] -- first panel on top, second on the bottom
//
// # Divider
//
// A draggable divider separates the two panels. The user can drag the divider
// to resize the panels. The divider width is configurable via [DividerWidth].
// When the mouse hovers over the divider, the cursor changes to a resize cursor.
//
// # Collapse
//
// When [CollapsibleOpt] is enabled, double-clicking the divider collapses the
// first panel. Double-clicking again restores it to the previous ratio.
//
// # Constraints
//
// Minimum sizes for each panel can be set via [MinFirst] and [MinSecond].
// The divider will stop at these limits during drag operations.
//
// # Visual Style
//
// The divider rendering is provided by a [Painter] implementation.
// Each design system (Material 3, Fluent, Cupertino) can supply its own
// painter to render the divider in the appropriate visual style.
//
// If no painter is set, [DefaultPainter] is used, which draws a minimal
// divider suitable for testing and prototyping.
//
// # Signal Binding
//
// The split ratio can be bound to a reactive signal from the [state] package.
// When a signal value changes, the split ratio automatically reflects
// the new state:
//
//   - [RatioSignal] -- TWO-WAY binding for the split ratio
//   - [RatioReadonlySignal] -- read-only binding for the split ratio
//
// Example:
//
//	ratio := state.NewSignal[float32](0.5)
//	split := splitview.New(
//	    splitview.First(leftPanel),
//	    splitview.Second(rightPanel),
//	    splitview.RatioSignal(ratio),
//	    splitview.OnRatioChange(func(r float32) {
//	        fmt.Printf("ratio: %.2f\n", r)
//	    }),
//	)
//
// # Focus
//
// SplitView does not participate in tab focus itself. Focus events are
// dispatched to the child panels directly.
package splitview
