// Package progressbar provides a linear progress bar widget for displaying
// a value between 0% and 100%.
//
// Construction uses functional options for immutable configuration,
// while fluent methods handle mutable styling:
//
//	bar := progressbar.New(
//	    progressbar.Value(0.65),
//	    progressbar.ShowLabel(true),
//	    progressbar.Height(20),
//	).Padding(8)
//
// # Visual Style
//
// The visual rendering is provided by a [Painter] implementation.
// Each design system (Material 3, Fluent, Cupertino) supplies its own
// painter to render progress bars in the appropriate visual style.
//
// If no painter is set, [DefaultPainter] is used, which draws a minimal
// progress bar suitable for testing and prototyping.
//
// # Value Range
//
// The progress bar displays a value in the range [0, 1], where 0 represents
// 0% completion and 1 represents 100%. Values outside this range are clamped.
//
// # Label
//
// When [ShowLabel] is enabled, the progress bar displays a percentage label
// centered over the bar. A custom label formatter can be provided via
// [FormatLabelFn].
//
// # Signal Binding
//
// Progress bar properties can be bound to reactive signals from the [state]
// package. When a signal value changes, the progress bar automatically
// reflects the new state.
//
//   - [ValueSignal] -- one-way read binding (read from writable signal)
//   - [ValueReadonlySignal] -- one-way read binding (read from computed/readonly signal)
//   - [ValueFn] -- dynamic function evaluated on each draw
//
// Example:
//
//	progress := state.NewSignal[float64](0.0)
//	bar := progressbar.New(
//	    progressbar.ValueSignal(progress),
//	    progressbar.ShowLabel(true),
//	)
//	progress.Set(0.75) // bar updates to 75%
//
// # Accessibility
//
// Progress bars are display-only widgets. They do not accept focus or
// handle input events. Screen readers should announce the current
// progress value via the accessibility tree.
package progressbar
