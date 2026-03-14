// Package linechart provides a real-time line chart widget for visualizing
// time-series data such as CPU usage, memory consumption, or network throughput.
//
// Construction uses functional options for immutable configuration:
//
//	chart := linechart.New(
//	    linechart.MaxPoints(60),
//	    linechart.YRange(0, 100),
//	    linechart.ShowGrid(true),
//	    linechart.ShowLabels(true),
//	)
//
// # Data Management
//
// Data is organized into named series, each with its own color and rolling
// window of data points:
//
//	chart.AddSeries("CPU", cpuColor)
//	chart.PushValue("CPU", 45.2) // adds point, shifts oldest if at MaxPoints
//
// Multiple series can be displayed simultaneously with different colors.
// [PushValue] is safe to call from any goroutine.
//
// # Visual Style
//
// The visual rendering is provided by a [Painter] implementation.
// Each design system (Material 3, Fluent, Cupertino) can supply its own
// painter to render charts in the appropriate visual style.
//
// If no painter is set, [DefaultPainter] is used, which draws a minimal
// chart suitable for testing and prototyping.
//
// # Signal Binding
//
// Chart data can be bound to reactive signals from the [state] package.
// When the signal value changes, the chart automatically reflects the new data:
//
//	seriesSig := state.NewSignal([]linechart.Series{...})
//	chart := linechart.New(
//	    linechart.SeriesSignal(seriesSig),
//	)
//
// # Retained Mode
//
// The chart calls [widget.WidgetBase.SetNeedsRedraw] when data changes via
// [Widget.PushValue] or [Widget.AddSeries], enabling efficient retained-mode
// rendering.
package linechart
