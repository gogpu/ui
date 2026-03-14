// Package transition provides widget enter/exit animations.
//
// Transition wraps a widget and plays an animation when the widget appears
// (enters) or disappears (exits). Supported effect types include fade,
// slide, and scale.
//
// # Quick Start
//
//	wrapped := transition.Wrap(myWidget,
//	    transition.EnterEffect(transition.FadeIn()),
//	    transition.ExitEffect(transition.FadeOut()),
//	    transition.Duration(300 * time.Millisecond),
//	)
//
//	wrapped.Show()  // plays enter animation, then shows widget
//	wrapped.Hide()  // plays exit animation, then hides widget
//
// # Effects
//
// Three built-in effect types are provided:
//
//   - Fade: animate opacity from 0 to 1 (enter) or 1 to 0 (exit)
//   - Slide: translate the widget from/to a given direction
//   - Scale: scale the widget from/to a smaller size with fade
//
// Effects can be combined by composing multiple property animations
// within a single [Effect] value.
//
// # Canvas Requirements
//
// Fade effects require the canvas to implement [OpacityPusher]. If the
// canvas does not support opacity, fade effects are silently skipped
// (graceful degradation). Slide effects use [widget.Canvas.PushTransform].
// Scale effects adjust the child widget bounds around the center point.
//
// # Retained Mode
//
// During animation, the transition widget calls [widget.WidgetBase.SetNeedsRedraw]
// to request continuous redraws until the animation completes.
package transition
