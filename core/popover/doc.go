// Package popover provides floating overlay widgets anchored to a trigger.
//
// Two widget types are provided:
//
//   - [Popover] -- click-triggered, displays arbitrary widget content
//   - [Tooltip] -- hover-triggered, displays a text label
//
// Both widgets position themselves relative to a trigger widget using one of
// 12 placement options (4 sides x 3 alignments). If the preferred placement
// would overflow the viewport, the widget automatically flips to the opposite
// side and clamps to the window bounds.
//
// # Popover
//
// A Popover opens on click and displays any widget.Widget as content. It uses
// the overlay stack for z-ordering and supports click-outside dismissal.
//
//	pop := popover.NewPopover(
//	    popover.TriggerWidget(myButton),
//	    popover.Content(menuPanel),
//	    popover.PlacementOpt(popover.BottomStart),
//	    popover.DismissOnClickOutside(true),
//	)
//
// # Tooltip
//
// A Tooltip opens after a configurable hover delay and displays a short text
// label. It closes when the mouse leaves the trigger.
//
//	tip := popover.NewTooltip(
//	    popover.TriggerWidget(saveButton),
//	    popover.TooltipText("Save document"),
//	    popover.PlacementOpt(popover.Bottom),
//	    popover.Delay(500 * time.Millisecond),
//	)
//
// # Visual Style
//
// Rendering is delegated to a [Painter] interface. If no painter is set,
// [DefaultPainter] is used, which draws a rounded rectangle with a shadow.
//
// # Overlay Integration
//
// Both widgets push themselves onto the window's overlay stack when shown and
// remove themselves when hidden. They integrate with [widget.OverlayManager]
// obtained from the [widget.Context].
//
// # Signal Binding
//
// The visible state can be bound to a reactive signal:
//
//	visible := state.NewSignal(false)
//	pop := popover.NewPopover(
//	    popover.TriggerWidget(btn),
//	    popover.Content(panel),
//	    popover.VisibleSignal(visible),
//	)
package popover
