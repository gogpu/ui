// Package docking provides an IDE-style dockable panel system.
//
// Construction uses functional options for immutable configuration:
//
//	host := docking.NewHost(
//	    docking.CenterContent(mainEditor),
//	    docking.PainterOpt(myPainter),
//	)
//
//	explorer := docking.NewPanel(
//	    docking.PanelTitle("Explorer"),
//	    docking.PanelContent(explorerWidget),
//	    docking.Closeable(true),
//	)
//
//	host.Dock(explorer, docking.Left)
//	host.Dock(terminal, docking.Bottom)
//
// # Layout
//
// The dock host manages a border layout with five zones:
//
//	+---------------------------------------------------+
//	|                    Top Zone                        |
//	+--------+----------------------------------+-------+
//	|        |                                  |       |
//	| Left   |         Center Zone              | Right |
//	| Zone   |         (main content)           | Zone  |
//	| [tabs] |                                  | [tabs]|
//	|        |                                  |       |
//	+--------+----------------------------------+-------+
//	|                   Bottom Zone                     |
//	|                   [tabs]                          |
//	+---------------------------------------------------+
//
// # Zones
//
// Five zones are available: [Left], [Right], [Top], [Bottom], and [Center].
// When multiple panels are docked to the same zone, they form a tabbed group.
// The active tab shows its content while others remain hidden.
//
// Zone sizes are controlled by ratios (0.0 to 1.0), configurable via
// [LeftRatio], [RightRatio], [TopRatio], and [BottomRatio].
// Zones with no panels are automatically collapsed.
//
// # Panels
//
// Each [Panel] has a title, optional content widget, and can be closeable.
// Panels are docked to zones programmatically via [Host.Dock].
// Removing a panel via [Host.Undock] or closing it removes it from the zone.
// If the zone becomes empty, it collapses.
//
// # Signal Binding
//
// The active panel index within each zone can be bound to a reactive signal
// for external observation. Use [Host.ActivePanelIndex] to query the active
// tab in a zone.
//
// # Visual Style
//
// Zone borders, tab headers, and backgrounds are rendered by a [Painter].
// If no painter is set, [DefaultPainter] is used.
//
// # Simplified Scope (Phase 4)
//
// The initial implementation provides:
//   - Programmatic docking only (no drag-to-dock)
//   - Zone sizes as ratios
//   - Tab switching in groups
//   - Panel close
//   - MovePanel between zones
//
// Drag-to-dock using the dnd package is planned for a future iteration.
package docking
