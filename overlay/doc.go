// Package overlay provides an overlay stack for displaying floating content
// above the normal widget tree.
//
// Overlays are used by dropdowns, popovers, tooltips, dialogs, and any
// widget that needs to float above the main UI. The package manages their
// lifecycle, event dispatch, and draw order.
//
// # Stack
//
// The core abstraction is [Stack], owned by the [app.Window], which manages
// a last-in-first-out collection of [Overlay] widgets. Events are dispatched
// to the top overlay before reaching the regular widget tree, and overlays
// are drawn after (on top of) the regular tree.
//
//	stack := overlay.NewStack(func() { requestRedraw() })
//	stack.Push(myOverlay)
//	stack.Pop()
//
// # Container
//
// [Container] is a ready-made [Overlay] implementation that wraps content
// with a full-window transparent backdrop. Clicks on the backdrop (outside
// the content) trigger dismissal. For modal overlays, a semi-transparent
// scrim is drawn behind the content:
//
//	c := overlay.NewContainer(menuWidget, windowSize,
//	    overlay.WithOnDismiss(func() { closeMenu() }),
//	    overlay.WithModal(true),
//	)
//
// # Positioning
//
// The [Position] helper calculates where an overlay should appear relative
// to an anchor widget, with automatic viewport clamping to keep the overlay
// on screen:
//
//	pos := overlay.Position(
//	    overlay.PlacementBelow,
//	    anchorBounds,
//	    overlaySize,
//	    windowSize,
//	    4, // gap
//	)
//
// # Integration
//
// Widgets push overlays via the [widget.OverlayManager] interface obtained
// from [widget.Context], and never import the overlay package directly.
// This avoids circular dependencies between widget packages and the overlay
// infrastructure.
package overlay
