// Package toggleswitch provides an on/off switch widget.
//
// A switch flips a single setting on or off, and the change takes effect
// immediately, unlike a checkbox that usually belongs to a form submitted
// later. Construction uses functional options:
//
//	sw := toggleswitch.New(
//	    toggleswitch.Label("Wi-Fi"),
//	    toggleswitch.Checked(true),
//	    toggleswitch.OnToggle(func(on bool) { setWiFi(on) }),
//	)
//
// # Visual Style
//
// Rendering is delegated to a [Painter]. Each design system (Material 3,
// Fluent, Cupertino, DevTools) supplies its own painter. Without one,
// [DefaultPainter] draws a neutral switch suitable for tests and prototypes.
//
// Painters may also implement [LayoutMetrics] to size the switch and
// [MotionProvider] to time its animations.
//
// # Animation
//
// The widget animates the thumb between off and on, and fades the hover
// highlight and press reaction in and out. Progress is computed from the
// frame clock ([widget.Context.Now]) on every draw and handed to the painter
// in [PaintState]: Position (eased), Progress (linear), HoverAmount and
// PressAmount. While anything is moving the widget requests further
// animation frames. With no frame clock (zero time, as in headless
// rendering) every change is instant.
//
// # Interaction
//
//   - Click (left button) toggles the switch when released inside it.
//   - Dragging the thumb sideways moves it with the pointer; on release the
//     switch takes the side the thumb is closer to.
//   - Space toggles a focused switch on release.
//
// Disabled switches ignore input, cannot take focus, and are drawn dimmed.
//
// # Signal Binding
//
//   - [LabelSignal] and [LabelReadonlySignal] bind the label one way.
//   - [CheckedSignal] is a TWO-WAY binding: the switch reads its value from
//     the signal and writes the new value back when the user toggles it.
//     Changing the signal from code animates the switch to the new value.
//   - [DisabledSignal] and [DisabledReadonlySignal] bind the disabled state.
//
// # Accessibility
//
// The widget implements [a11y.Accessible] with [a11y.RoleSwitch] and reports
// its value through [a11y.State.Checked].
package toggleswitch
