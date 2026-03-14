// Package progress provides a circular progress indicator widget with
// determinate and indeterminate modes.
//
// In determinate mode, the indicator shows a progress arc from 0% to 100%.
// In indeterminate mode, a rotating arc spins continuously to indicate
// ongoing activity.
//
// Construction uses functional options for immutable configuration:
//
//	// Determinate (65% complete)
//	indicator := progress.New(
//	    progress.Value(0.65),
//	    progress.Size(48),
//	    progress.StrokeWidth(4),
//	    progress.ShowLabel(true),
//	)
//
//	// Indeterminate (spinner)
//	spinner := progress.New(
//	    progress.Indeterminate(true),
//	    progress.Size(32),
//	)
//
// # Visual Style
//
// The visual rendering is provided by a [Painter] implementation.
// Each design system (Material 3, Fluent, Cupertino) supplies its own
// painter to render the indicator in the appropriate visual style.
//
// If no painter is set, [DefaultPainter] is used, which draws the
// indicator using polyline arc approximation.
//
// # Signal Binding
//
// The value property supports 4-level signal binding priority:
//
//   - [ValueReadonlySignal] -- highest priority (computed/readonly signal)
//   - [ValueSignal] -- writable signal binding
//   - [ValueFn] -- dynamic function evaluated on each draw
//   - [Value] -- static value (lowest priority)
//
// # Accessibility
//
// Circular progress indicators are display-only widgets. They do not
// accept focus or handle input events.
package progress
