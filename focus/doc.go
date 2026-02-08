// Package focus provides keyboard focus management for a widget tree.
//
// The focus package handles tab navigation, focus tracking, keyboard
// shortcuts, and focus ring drawing. It works with any widget tree
// built from [widget.Widget] and [widget.Focusable] interfaces.
//
// # Focus Manager
//
// The [Manager] tracks which widget currently has keyboard focus and
// provides methods for programmatic focus control:
//
//	root := buildWidgetTree()
//	fm := focus.New(root)
//	fm.Focus(myButton)   // Give focus to a specific widget
//	fm.Next()            // Tab to next focusable widget
//	fm.Previous()        // Shift+Tab to previous focusable widget
//	fm.Blur()            // Remove focus entirely
//
// # Tab Navigation
//
// Tab order is determined by depth-first traversal of the widget tree.
// Widgets that are not visible, not enabled, or not focusable are skipped.
// Navigation wraps around when reaching the end or beginning of the list.
//
// # Keyboard Shortcuts
//
// The manager supports global keyboard shortcuts that are checked before
// tab navigation:
//
//	fm.RegisterShortcut(focus.Shortcut{Key: event.KeyS, Ctrl: true}, func() {
//	    saveFile()
//	})
//
// # Focus Ring
//
// The [DrawFocusRing] function draws a visual focus indicator around a widget:
//
//	if w.IsFocused() {
//	    focus.DrawFocusRing(canvas, w.Bounds(), focusColor, 4.0)
//	}
package focus
