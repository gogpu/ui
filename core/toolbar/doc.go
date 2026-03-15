// Package toolbar provides a horizontal action bar widget.
//
// A toolbar displays a horizontal row of icon buttons, separators, spacers,
// and custom widgets. It is commonly used at the top of an application window
// to provide quick access to frequently used actions.
//
// Construction uses convenience functions for items and functional options
// for toolbar-level configuration:
//
//	tb := toolbar.New(
//	    toolbar.Items(
//	        toolbar.IconButton("New", icon.Add, onNew),
//	        toolbar.IconButton("Open", icon.Menu, onOpen),
//	        toolbar.Separator(),
//	        toolbar.Spacer(),
//	        toolbar.IconButton("Settings", icon.Settings, onSettings),
//	    ),
//	    toolbar.Height(40),
//	)
//
// # Item Types
//
// Four item types are supported:
//   - [ItemButton] -- an icon button with optional text label and click handler
//   - [ItemSeparator] -- a vertical divider line between button groups
//   - [ItemSpacer] -- a flexible gap that pushes subsequent items to the right
//   - [ItemCustom] -- any [widget.Widget] embedded in the toolbar
//
// # Visual Style
//
// The visual rendering is provided by a [Painter] implementation.
// Each design system (Material 3, Fluent, Cupertino) can supply its own
// painter. If no painter is set, [DefaultPainter] is used.
//
// # Interaction
//
// Button items respond to mouse events (hover, press, click) and keyboard
// activation (Enter or Space when focused). Tab navigates between items.
// Disabled items ignore all interaction.
//
// # Accessibility
//
// The toolbar has the [a11y.RoleToolbar] role. Each button item is
// individually focusable and announces its label to screen readers.
package toolbar
