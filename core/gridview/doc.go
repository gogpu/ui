// Package gridview provides a virtualized 2D grid widget that renders only
// visible cells, enabling efficient display of large datasets in a grid layout.
//
// Construction uses functional options for immutable configuration:
//
//	gv := gridview.New(
//	    gridview.Columns(4),
//	    gridview.ItemCount(1000),
//	    gridview.ItemSize(120, 120),
//	    gridview.BuildCell(func(index int, ctx gridview.CellContext) widget.Widget {
//	        return primitives.NewText(fmt.Sprintf("Item %d", index))
//	    }),
//	)
//
// # Column Modes
//
// Two column modes are available:
//
//   - [Columns] -- fixed column count
//   - [ColumnsAuto] -- auto-fit based on viewport width and cell size
//
// # ScrollView Composition
//
// GridView composes an internal [scrollview.Widget] for scroll handling.
// All scroll features (wheel, keyboard PageUp/Down/Home/End, scrollbar
// thumb drag, momentum, signals) are provided automatically.
// Scrolling is vertical only — columns are sized to fit within the viewport.
//
// # Selection
//
// Optional cell selection is supported via [SelectedIndex] or
// [SelectedIndexSignal]. Selection state is passed to the builder
// callback through [CellContext.IsSelected].
//
//   - [SelectionNone] -- no selection (default)
//   - [SelectionSingle] -- at most one cell selected
//
// # Signal Binding
//
// Grid properties can be bound to reactive signals from the [state] package.
// When a signal value changes, the grid automatically reflects the new state.
//
//   - [ItemCountSignal] -- one-way binding for item count
//   - [SelectedIndexSignal] -- TWO-WAY binding for selected cell
//   - [ColumnsSignal] -- one-way binding for column count
//   - [ScrollYSignal] -- TWO-WAY binding for scroll offset (via internal ScrollView)
//   - [DisabledSignal] -- one-way binding for disabled state
//
// Example with signals:
//
//	count := state.NewSignal(100)
//	selected := state.NewSignal(-1)
//
//	gv := gridview.New(
//	    gridview.ItemCountSignal(count),
//	    gridview.ItemSize(120, 120),
//	    gridview.Columns(4),
//	    gridview.SelectedIndexSignal(selected),
//	    gridview.BuildCell(func(index int, ctx gridview.CellContext) widget.Widget {
//	        return buildThumbnail(images[index], ctx.IsSelected)
//	    }),
//	    gridview.Gap(8),
//	)
//
// # Visual Style
//
// The visual rendering of grid-specific elements (cell backgrounds, selection
// highlights, hover effects, empty state) is provided by a [Painter]
// implementation. Each design system (Material 3, Fluent, Cupertino) supplies
// its own painter.
//
// If no painter is set, [DefaultPainter] is used, which draws minimal
// visuals suitable for testing and prototyping.
//
// # Accessibility
//
// GridView implements [a11y.Accessible] with [a11y.RoleGrid]. Visible
// cells are exposed as grid cell children. Keyboard navigation
// with arrow keys moves selection between cells in all four directions.
//
// # Focus
//
// GridView implements [widget.Focusable] and participates in tab navigation.
// When focused, arrow keys move selection, Home/End jump to first/last cell.
package gridview
