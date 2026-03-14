// Package collapsible provides a collapsible/expandable section widget.
//
// A collapsible section has a clickable header bar and a content area that
// shows or hides when the header is toggled. The expand/collapse transition
// can be animated using the animation engine.
//
// Construction uses functional options for immutable configuration:
//
//	section := collapsible.New(
//	    collapsible.Title("CPU Usage"),
//	    collapsible.Content(chartWidget),
//	    collapsible.Expanded(true),
//	    collapsible.HeaderHeight(36),
//	    collapsible.Animated(true),
//	    collapsible.Duration(200*time.Millisecond),
//	)
//
// # Visual Style
//
// The header rendering is provided by a [Painter] implementation.
// Each design system (Material 3, Fluent, Cupertino) can supply its own
// painter. If no painter is set, [DefaultPainter] is used, which draws a
// minimal header with a background, title text, and an expand/collapse
// arrow indicator.
//
// # Animation
//
// When [Animated] is true (the default), toggling the section smoothly
// animates the content height using [animation.EaseInOutCubic] easing.
// The animation progress drives content clipping via [widget.Canvas.PushClip].
// Set [Animated] to false for instant expand/collapse.
//
// # Signal Binding
//
// The expanded state can be bound to a reactive signal:
//
//   - [ExpandedSignal] -- TWO-WAY binding: reads from signal AND writes back on toggle
//
// # Interaction
//
// Clicking the header or pressing Enter/Space when focused toggles the section.
// The [OnToggle] callback is invoked with the new expanded state.
//
// # Focus
//
// The collapsible section implements [widget.Focusable] and participates in
// tab navigation. A focus ring is drawn on the header when focused.
package collapsible
