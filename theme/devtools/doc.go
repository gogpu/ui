// Package devtools provides a JetBrains-inspired DevTools design system theme.
//
// DevTools is a dark-first, high-density, IDE-inspired design system modeled
// after the JetBrains Int UI (IntelliJ Platform New UI). It features compact
// spacing, 13px base typography, and a gray-scale palette with blue accents.
//
// # Creating a Theme
//
// Create a theme from the default blue accent:
//
//	dark := devtools.NewDarkTheme()   // dark mode (primary)
//	light := devtools.NewTheme()      // light mode
//
// Customize the accent color:
//
//	t := devtools.NewDarkTheme(devtools.WithAccentColor(widget.Hex(0x57965C)))
//
// # Painters
//
// DevTools provides 20 painters covering all interactive and data-display widgets:
//
//   - [ButtonPainter] — compact 28px buttons with 4px radius
//   - [CheckboxPainter] — 16px checkbox with check/dash marks
//   - [RadioPainter] — 16px radio with inner dot
//   - [TextFieldPainter] — single-line input with cursor, selection, error state
//   - [DropdownPainter] — trigger + menu with keyboard navigation
//   - [SliderPainter] — horizontal/vertical with marks
//   - [DialogPainter] — modal/modeless dialog surface
//   - [ScrollbarPainter] — thin overlay scrollbar
//   - [TabViewPainter] — IDE-style tab strip with close buttons
//   - [TreeViewPainter] — project tree with indent guides and expand icons
//   - [DataTablePainter] — sortable table with zebra striping
//   - [ToolbarPainter] — IDE toolbar with icon buttons and separators
//   - [MenuPainter] — menu bar and dropdown menus with shortcuts
//   - [CollapsiblePainter] — expandable section header with arrow animation
//   - [ProgressPainter] — circular determinate/indeterminate progress
//   - [SplitViewPainter] — minimal 1px panel divider
//   - [DockingPainter] — docking zone tabs and borders
//   - [PopoverPainter] — popover surface and tooltip
//   - [LineChartPainter] — real-time line chart with grid
//   - [ListViewPainter] — list item backgrounds, selection, dividers
//
// Use [NewPainters] to create all 20 painters at once:
//
//	dt := devtools.NewDarkTheme()
//	p := devtools.NewPainters(dt)
//	btn := button.New(button.PainterOpt(p.Button))
//	cb := checkbox.New(checkbox.PainterOpt(p.Checkbox))
//
// # Design Characteristics
//
//   - Dark-first: optimized for long coding sessions
//   - Compact spacing: 2/4/6/8/12/16/24px scale (smaller than M3 or Fluent)
//   - 13px base font size (vs 14px in M3 and Fluent)
//   - Small radii: 4px default component radius
//   - Gray-scale palette: 14 neutral tones from #1E1F22 to #FFFFFF
//   - Blue accent: #3574F0 (JetBrains default)
//   - Dark header toolbar in both light and dark modes
//
// # Integration
//
// Convert to the generic theme system with [Theme.AsTheme]:
//
//	generic := devtools.NewDarkTheme().AsTheme()
package devtools
