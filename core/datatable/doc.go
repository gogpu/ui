// Package datatable provides a sortable data table widget with fixed header,
// virtualized rows, and column-level configuration.
//
// Construction uses functional options for immutable configuration:
//
//	dt := datatable.New(
//	    datatable.Columns([]datatable.Column{
//	        {Key: "name", Title: "Name", Width: 200, Sortable: true},
//	        {Key: "size", Title: "Size", Width: 100, Sortable: true, Align: widget.TextAlignRight},
//	        {Key: "date", Title: "Modified", Width: 150, Sortable: true},
//	    }),
//	    datatable.RowCount(1000),
//	    datatable.CellValue(func(row int, col string) string {
//	        return data[row][col]
//	    }),
//	)
//
// # Fixed Header
//
// The header row is always visible at the top, showing column titles and sort
// indicators (ascending/descending arrows). Clicking a sortable column header
// cycles through: none -> ascending -> descending -> none.
//
// # Virtualization
//
// Only rows visible in the viewport are rendered. The table composes an internal
// [scrollview.Widget] for vertical scrolling. This allows efficient display of
// thousands of rows without performance degradation.
//
// # Selection
//
// Optional row selection is supported via [OnRowSelect] and [SelectedRow] or
// [SelectedRowSignal]:
//
//   - [SelectionNone] -- no selection (default)
//   - [SelectionSingle] -- at most one row selected
//   - [SelectionMulti] -- multiple rows selected (Ctrl+click)
//
// # Signal Binding
//
// Table properties can be bound to reactive signals from the [state] package.
//
//   - [RowCountSignal] -- one-way binding for row count
//   - [SelectedRowSignal] -- TWO-WAY binding for selected row
//   - [ScrollYSignal] -- TWO-WAY binding for scroll offset
//   - [DisabledSignal] -- one-way binding for disabled state
//
// # Visual Style
//
// Rendering of table-specific elements (header, row backgrounds, sort indicators,
// selection highlights) is delegated to a [Painter] implementation. Each design
// system supplies its own painter.
//
// If no painter is set, [DefaultPainter] is used.
//
// # Accessibility
//
// DataTable implements [a11y.Accessible] with [a11y.RoleTable]. Keyboard
// navigation with Up/Down moves selection, Home/End jump to first/last row.
//
// # Focus
//
// DataTable implements [widget.Focusable] and participates in tab navigation.
package datatable
