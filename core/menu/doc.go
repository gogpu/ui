// Package menu provides a menu system with MenuBar and ContextMenu widgets.
//
// MenuBar is a horizontal bar at the top of the window containing top-level
// menu labels. Clicking a label opens a vertical dropdown menu. ContextMenu
// is a floating menu shown at a specific position, typically on right-click.
//
// Both widgets support:
//   - Nested submenus (hover to open)
//   - Separator items
//   - Disabled items (grayed out)
//   - Keyboard navigation (arrow keys, Enter, Escape)
//   - Shortcut display text (cosmetic only)
//   - Pluggable Painter for design-system independence
//   - Overlay integration for popup menus
//
// MenuBar example:
//
//	bar := menu.NewBar(
//	    menu.BarMenu("File",
//	        menu.Item("New", "Ctrl+N", onNew),
//	        menu.Item("Open", "Ctrl+O", onOpen),
//	        menu.Sep(),
//	        menu.Item("Save", "Ctrl+S", onSave),
//	    ),
//	    menu.BarMenu("Edit",
//	        menu.Item("Undo", "Ctrl+Z", onUndo),
//	        menu.Item("Redo", "Ctrl+Y", onRedo),
//	    ),
//	)
//
// ContextMenu example:
//
//	ctx := menu.NewContextMenu(
//	    menu.Item("Cut", "Ctrl+X", onCut),
//	    menu.Item("Copy", "Ctrl+C", onCopy),
//	    menu.Item("Paste", "Ctrl+V", onPaste),
//	)
//	ctx.Show(widgetCtx, position)
//
// The visual appearance is controlled by a pluggable [Painter] interface.
// The default painter provides a minimal fallback; use a design system painter
// (e.g., material3.MenuPainter) for production styling.
package menu
