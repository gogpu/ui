// Package stripe provides a vertical tool window sidebar widget.
//
// A stripe displays a vertical column of icon buttons with optional text labels,
// split into top and bottom groups. It is commonly used at the edges of an IDE
// window to provide quick access to tool windows (Project, Terminal, Git, etc.),
// following the JetBrains IDE ToolWindowToolbar pattern.
//
// Construction uses functional options:
//
//	s := stripe.New(
//	    stripe.TopItems(
//	        stripe.Button{ID: "project", Label: "Project", Icon: icon.FolderClosed, OnClick: fn},
//	        stripe.Button{ID: "commit", Label: "Commit", Icon: icon.Check, OnClick: fn},
//	    ),
//	    stripe.BottomItems(
//	        stripe.Button{ID: "terminal", Label: "Terminal", Icon: icon.Terminal, OnClick: fn},
//	    ),
//	    stripe.ActiveID("terminal"),
//	    stripe.ShowLabels(true),
//	    stripe.Width(64),
//	    stripe.PainterOpt(devtoolsPainter),
//	)
//
// # Layout
//
// Buttons are stacked vertically. Top items are gravity-aligned to the top edge
// and bottom items to the bottom edge. Each button spans the full strip width.
//
// # Visual Style
//
// The visual rendering is provided by a [Painter] implementation. Each design
// system can supply its own painter. If no painter is set, [DefaultPainter] is
// used.
//
// # Interaction
//
// Buttons respond to mouse hover, press, and click. Clicking a button fires its
// OnClick handler and sets it as the active button. The active button receives
// a distinct visual treatment from the painter.
//
// # Accessibility
//
// The stripe has the [a11y.RoleToolbar] role. Individual buttons announce their
// label to screen readers.
package stripe
