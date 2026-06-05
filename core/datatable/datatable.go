package datatable

import (
	"fmt"
	"math"

	"github.com/gogpu/ui/a11y"
	"github.com/gogpu/ui/core/scrollview"
	"github.com/gogpu/ui/event"
	"github.com/gogpu/ui/geometry"
	"github.com/gogpu/ui/state"
	"github.com/gogpu/ui/widget"
)

// SelectionMode defines how rows can be selected in the table.
type SelectionMode uint8

// SelectionMode constants.
const (
	// SelectionNone disables row selection. This is the default.
	SelectionNone SelectionMode = iota

	// SelectionSingle allows at most one row to be selected at a time.
	SelectionSingle

	// SelectionMulti allows multiple rows to be selected (Ctrl+click).
	SelectionMulti
)

// String returns a human-readable name for the selection mode.
func (m SelectionMode) String() string {
	switch m {
	case SelectionNone:
		return selModeNoneStr
	case SelectionSingle:
		return selModeSingleStr
	case SelectionMulti:
		return selModeMultiStr
	default:
		return selModeUnknownStr
	}
}

// Selection mode string constants.
const (
	selModeNoneStr    = "None"
	selModeSingleStr  = "Single"
	selModeMultiStr   = "Multi"
	selModeUnknownStr = "Unknown"
)

// --- Config ---

// config holds the data table's configuration, set at construction time via options.
type config struct {
	columns []Column

	// Row data.
	rowCount               int
	rowCountFn             func() int
	rowCountSignal         state.Signal[int]
	readonlyRowCountSignal state.ReadonlySignal[int]

	// Cell value provider.
	cellValue func(row int, col string) string

	// Row height.
	rowHeight float32

	// Sort callback.
	onSort func(col string, ascending bool)

	// Selection.
	selectionMode             SelectionMode
	selectedRow               int
	selectedRowSignal         state.Signal[int]
	readonlySelectedRowSignal state.ReadonlySignal[int]
	onRowSelect               func(row int)

	// Multi-selection (only used when selectionMode == SelectionMulti).
	selectedRows map[int]bool

	// Disabled state.
	disabled               bool
	disabledFn             func() bool
	disabledSignal         state.Signal[bool]
	readonlyDisabledSignal state.ReadonlySignal[bool]

	// Scroll pass-through.
	scrollYSignal state.Signal[float32]
	onScroll      func(offset float32)

	// Visual.
	painter   Painter
	a11yLabel string
}

// defaultConfig returns a config with sensible defaults.
func defaultConfig() config {
	return config{
		rowHeight:   defaultRowHeight,
		selectedRow: noSelection,
	}
}

// Default configuration values.
const (
	defaultRowHeight      float32 = 32
	defaultHeaderHeight   float32 = 36
	defaultMinColumnWidth float32 = 50
	noSelection           int     = -1
	noHoveredRow          int     = -1
	noHoveredCol          int     = -1
)

// ResolvedRowCount returns the current row count.
// Priority: ReadonlySignal > Signal > Fn > Static.
func (c *config) ResolvedRowCount() int {
	if c.readonlyRowCountSignal != nil {
		return c.readonlyRowCountSignal.Get()
	}
	if c.rowCountSignal != nil {
		return c.rowCountSignal.Get()
	}
	if c.rowCountFn != nil {
		return c.rowCountFn()
	}
	return c.rowCount
}

// ResolvedSelectedRow returns the current selected row index.
// Priority: ReadonlySignal > Signal > Static. Returns -1 if no selection.
func (c *config) ResolvedSelectedRow() int {
	if c.readonlySelectedRowSignal != nil {
		return c.readonlySelectedRowSignal.Get()
	}
	if c.selectedRowSignal != nil {
		return c.selectedRowSignal.Get()
	}
	return c.selectedRow
}

// ResolvedDisabled returns the current disabled state.
// Priority: ReadonlySignal > Signal > Fn > Static.
func (c *config) ResolvedDisabled() bool {
	if c.readonlyDisabledSignal != nil {
		return c.readonlyDisabledSignal.Get()
	}
	if c.disabledSignal != nil {
		return c.disabledSignal.Get()
	}
	if c.disabledFn != nil {
		return c.disabledFn()
	}
	return c.disabled
}

// --- Options ---

// Option configures a data table during construction.
type Option func(*config)

// Columns sets the column definitions.
func Columns(cols []Column) Option {
	return func(c *config) { c.columns = cols }
}

// RowCount sets the total number of data rows.
func RowCount(n int) Option {
	return func(c *config) { c.rowCount = n }
}

// RowCountFn sets a dynamic function that returns the row count.
func RowCountFn(fn func() int) Option {
	return func(c *config) { c.rowCountFn = fn }
}

// RowCountSignal binds the row count to a reactive signal.
func RowCountSignal(sig state.Signal[int]) Option {
	return func(c *config) { c.rowCountSignal = sig }
}

// RowCountReadonlySignal binds the row count to a read-only signal.
func RowCountReadonlySignal(sig state.ReadonlySignal[int]) Option {
	return func(c *config) { c.readonlyRowCountSignal = sig }
}

// RowHeight sets the height of each data row in logical pixels. Default: 32.
func RowHeight(h float32) Option {
	return func(c *config) { c.rowHeight = h }
}

// CellValue sets the callback that provides cell text for a given row and column key.
func CellValue(fn func(row int, col string) string) Option {
	return func(c *config) { c.cellValue = fn }
}

// OnSort sets the callback invoked when a column sort changes.
// The ascending parameter indicates the new sort direction (true=asc, false=desc).
// The callback is NOT invoked when sort is cleared (direction returns to None).
func OnSort(fn func(col string, ascending bool)) Option {
	return func(c *config) { c.onSort = fn }
}

// SelectionModeOpt sets the row selection mode. Default: SelectionNone.
func SelectionModeOpt(mode SelectionMode) Option {
	return func(c *config) { c.selectionMode = mode }
}

// SelectedRow sets the initially selected row index. Use -1 for no selection.
func SelectedRow(row int) Option {
	return func(c *config) { c.selectedRow = row }
}

// SelectedRowSignal binds the selected row to a reactive signal.
// TWO-WAY binding: reads from signal, writes back on user selection.
func SelectedRowSignal(sig state.Signal[int]) Option {
	return func(c *config) { c.selectedRowSignal = sig }
}

// SelectedRowReadonlySignal binds the selected row to a read-only signal.
func SelectedRowReadonlySignal(sig state.ReadonlySignal[int]) Option {
	return func(c *config) { c.readonlySelectedRowSignal = sig }
}

// OnRowSelect sets the callback invoked when a row is selected.
func OnRowSelect(fn func(row int)) Option {
	return func(c *config) { c.onRowSelect = fn }
}

// Disabled sets the table's disabled state.
func Disabled(d bool) Option {
	return func(c *config) { c.disabled = d }
}

// DisabledFn sets a dynamic function for the disabled state.
func DisabledFn(fn func() bool) Option {
	return func(c *config) { c.disabledFn = fn }
}

// DisabledSignal binds the disabled state to a reactive signal.
func DisabledSignal(sig state.Signal[bool]) Option {
	return func(c *config) { c.disabledSignal = sig }
}

// DisabledReadonlySignal binds the disabled state to a read-only signal.
func DisabledReadonlySignal(sig state.ReadonlySignal[bool]) Option {
	return func(c *config) { c.readonlyDisabledSignal = sig }
}

// ScrollYSignal binds the vertical scroll offset to a reactive signal (TWO-WAY).
func ScrollYSignal(sig state.Signal[float32]) Option {
	return func(c *config) { c.scrollYSignal = sig }
}

// OnScroll sets the callback invoked when the scroll position changes.
func OnScroll(fn func(offset float32)) Option {
	return func(c *config) { c.onScroll = fn }
}

// PainterOpt sets the painter used to render table visuals.
func PainterOpt(p Painter) Option {
	return func(c *config) { c.painter = p }
}

// A11yLabel sets the accessibility label for the table.
func A11yLabel(label string) Option {
	return func(c *config) { c.a11yLabel = label }
}

// --- Widget ---

// Widget implements a sortable data table with fixed header, virtualized rows,
// and pluggable painting.
//
// A data table is created with [New] using functional options:
//
//	dt := datatable.New(
//	    datatable.Columns(cols),
//	    datatable.RowCount(1000),
//	    datatable.CellValue(valueFn),
//	)
type Widget struct {
	widget.WidgetBase
	cfg     config
	painter Painter

	// Internal scroll view for the data rows (not header).
	scroll  *scrollview.Widget
	virtual *virtualContent

	// Layout state.
	viewportWidth  float32
	viewportHeight float32
	colWidths      []float32 // resolved column widths

	// Sort state.
	sortColumn    int           // index into cfg.columns, -1 = none
	sortDirection SortDirection // current sort direction

	// Interaction state.
	hoveredRow    int // data row under cursor, -1 = none
	hoveredColHdr int // header column under cursor, -1 = none
}

// New creates a new data table Widget with the given options.
//
// The returned widget is visible, enabled, and focusable by default.
func New(opts ...Option) *Widget {
	cfg := defaultConfig()
	for _, opt := range opts {
		opt(&cfg)
	}

	w := &Widget{
		cfg:           cfg,
		painter:       DefaultPainter{},
		sortColumn:    noSortColumn,
		sortDirection: SortNone,
		hoveredRow:    noHoveredRow,
		hoveredColHdr: noHoveredCol,
	}
	w.SetVisible(true)
	w.SetEnabled(true)

	if w.cfg.painter != nil {
		w.painter = w.cfg.painter
	}

	// Initialize multi-selection map.
	if w.cfg.selectionMode == SelectionMulti && w.cfg.selectedRows == nil {
		w.cfg.selectedRows = make(map[int]bool)
	}

	// Create virtual content widget.
	w.virtual = &virtualContent{table: w}
	w.virtual.SetVisible(true)

	// Build internal scroll view.
	svOpts := []scrollview.Option{
		scrollview.DirectionOpt(scrollview.Vertical),
	}
	if w.cfg.scrollYSignal != nil {
		svOpts = append(svOpts, scrollview.ScrollYSignal(w.cfg.scrollYSignal))
	}
	if w.cfg.onScroll != nil {
		fn := w.cfg.onScroll
		svOpts = append(svOpts, scrollview.OnScroll(func(_, y float32) {
			fn(y)
		}))
	}

	w.scroll = scrollview.New(w.virtual, svOpts...)

	// ADR-028: parent chain for upward dirty propagation.
	// Flutter: RenderObject.adoptChild sets parent on each child.
	w.scroll.SetParent(w)

	return w
}

// No sort column sentinel.
const noSortColumn = -1

// IsFocusable reports whether the table can currently receive focus.
func (w *Widget) IsFocusable() bool {
	return w.IsVisible() && w.IsEnabled() && !w.cfg.ResolvedDisabled()
}

// Layout calculates the table's size within the given constraints.
func (w *Widget) Layout(ctx widget.Context, constraints geometry.Constraints) geometry.Size {
	size := constraints.Biggest()
	if size.Width >= geometry.Infinity {
		size.Width = constraints.Constrain(geometry.Sz(defaultViewportWidth, 0)).Width
	}
	if size.Height >= geometry.Infinity {
		totalH := defaultHeaderHeight + w.totalDataHeight()
		if totalH > defaultViewportHeight {
			totalH = defaultViewportHeight
		}
		size.Height = totalH
	}
	if size.Width <= 0 {
		size.Width = defaultViewportWidth
	}
	if size.Height <= 0 {
		size.Height = defaultViewportHeight
	}

	w.viewportWidth = size.Width
	w.viewportHeight = size.Height

	// Resolve column widths.
	w.resolveColumnWidths()

	// The scroll view occupies everything below the header.
	scrollH := size.Height - defaultHeaderHeight
	if scrollH < 0 {
		scrollH = 0
	}
	svConstraints := geometry.Tight(geometry.Sz(size.Width, scrollH))
	w.scroll.Layout(ctx, svConstraints)

	return size
}

// Draw renders the data table.
func (w *Widget) Draw(ctx widget.Context, canvas widget.Canvas) {
	if !w.IsVisible() {
		return
	}
	bounds := w.Bounds()
	if bounds.IsEmpty() {
		return
	}

	// Clip to table bounds.
	canvas.PushClip(bounds)
	defer canvas.PopClip()

	rowCount := w.cfg.ResolvedRowCount()
	if rowCount == 0 && len(w.cfg.columns) == 0 {
		w.painter.PaintEmptyState(canvas, bounds)
		return
	}

	// Draw header.
	w.drawHeader(ctx, canvas, bounds)

	// Draw data rows via the scroll view.
	w.updateScrollBounds()
	widget.StampScreenOrigin(w.scroll, canvas)
	w.scroll.Draw(ctx, canvas)
}

// Event handles an input event and returns true if consumed.
func (w *Widget) Event(ctx widget.Context, e event.Event) bool {
	if !w.IsVisible() || !w.IsEnabled() {
		return false
	}

	// Ensure scroll view bounds are set (they match the data area below the header).
	// This is needed because ScrollView transforms event coordinates using its bounds,
	// and bounds must be set before event dispatch, not just in Draw.
	w.updateScrollBounds()

	// Handle keyboard at table level.
	if ke, ok := e.(*event.KeyEvent); ok {
		if w.handleKeyEvent(ctx, ke) {
			return true
		}
	}

	// Handle header mouse events.
	if me, ok := e.(*event.MouseEvent); ok {
		if w.handleHeaderMouseEvent(ctx, me) {
			return true
		}
	}

	// Delegate to scroll view for data rows.
	return w.scroll.Event(ctx, e)
}

// updateScrollBounds sets the scroll view bounds to the data area below the header.
func (w *Widget) updateScrollBounds() {
	bounds := w.Bounds()
	if bounds.IsEmpty() {
		return
	}
	scrollBounds := geometry.NewRect(
		bounds.Min.X,
		bounds.Min.Y+defaultHeaderHeight,
		bounds.Width(),
		bounds.Height()-defaultHeaderHeight,
	)
	w.scroll.SetBounds(scrollBounds)
}

// Children returns the internal scroll view as the single child.
func (w *Widget) Children() []widget.Widget {
	if w.scroll == nil {
		return nil
	}
	return []widget.Widget{w.scroll}
}

// Mount creates signal bindings for push-based invalidation.
func (w *Widget) Mount(ctx widget.Context) {
	sched := ctx.Scheduler()
	if sched == nil {
		return
	}

	if w.cfg.readonlyRowCountSignal != nil {
		b := state.BindToScheduler(w.cfg.readonlyRowCountSignal, w, sched)
		w.AddBinding(b)
	} else if w.cfg.rowCountSignal != nil {
		b := state.BindToScheduler(w.cfg.rowCountSignal, w, sched)
		w.AddBinding(b)
	}

	if w.cfg.readonlySelectedRowSignal != nil {
		b := state.BindToScheduler(w.cfg.readonlySelectedRowSignal, w, sched)
		w.AddBinding(b)
	} else if w.cfg.selectedRowSignal != nil {
		b := state.BindToScheduler(w.cfg.selectedRowSignal, w, sched)
		w.AddBinding(b)
	}

	if w.cfg.readonlyDisabledSignal != nil {
		b := state.BindToScheduler(w.cfg.readonlyDisabledSignal, w, sched)
		w.AddBinding(b)
	} else if w.cfg.disabledSignal != nil {
		b := state.BindToScheduler(w.cfg.disabledSignal, w, sched)
		w.AddBinding(b)
	}

	w.scroll.Mount(ctx)
}

// Unmount is called when the table is removed from the widget tree.
func (w *Widget) Unmount() {
	w.scroll.Unmount()
}

// --- Public API ---

// SortColumn returns the currently sorted column key and direction.
// Returns empty string and SortNone if no column is sorted.
func (w *Widget) SortColumn() (string, SortDirection) {
	if w.sortColumn < 0 || w.sortColumn >= len(w.cfg.columns) {
		return "", SortNone
	}
	return w.cfg.columns[w.sortColumn].Key, w.sortDirection
}

// SetSort programmatically sets the sort column and direction.
// Pass an empty key or SortNone to clear sorting.
func (w *Widget) SetSort(colKey string, dir SortDirection) {
	if colKey == "" || dir == SortNone {
		w.sortColumn = noSortColumn
		w.sortDirection = SortNone
		return
	}
	for i, col := range w.cfg.columns {
		if col.Key == colKey {
			w.sortColumn = i
			w.sortDirection = dir
			return
		}
	}
}

// VisibleRowRange returns the indices of currently visible rows [start, end).
func (w *Widget) VisibleRowRange() (start, end int) {
	rowCount := w.cfg.ResolvedRowCount()
	if rowCount == 0 {
		return 0, 0
	}
	scrollY := w.currentScrollY()
	dataViewH := w.viewportHeight - defaultHeaderHeight
	if dataViewH <= 0 {
		return 0, 0
	}
	first, last := w.visibleRowRange(scrollY, dataViewH)
	if last >= rowCount {
		last = rowCount - 1
	}
	if first > last {
		first = last
	}
	return first, last + 1
}

// IsRowSelected reports whether the given row is selected.
func (w *Widget) IsRowSelected(row int) bool {
	if w.cfg.selectionMode == SelectionMulti {
		return w.cfg.selectedRows[row]
	}
	return row == w.cfg.ResolvedSelectedRow()
}

// GetRowCount returns the current row count.
func (w *Widget) GetRowCount() int {
	return w.cfg.ResolvedRowCount()
}

// InvalidateData signals that the underlying data has changed.
func (w *Widget) InvalidateData() {
	// Nothing cached; just invalidate to trigger redraw.
}

// ScrollToRow scrolls to make the given row visible.
func (w *Widget) ScrollToRow(row int) {
	rowCount := w.cfg.ResolvedRowCount()
	if row < 0 || row >= rowCount {
		return
	}
	rowTop := float32(row) * w.cfg.rowHeight
	rowBottom := rowTop + w.cfg.rowHeight
	scrollY := w.currentScrollY()
	dataViewH := w.viewportHeight - defaultHeaderHeight
	if dataViewH <= 0 {
		return
	}

	if rowTop >= scrollY && rowBottom <= scrollY+dataViewH {
		return // already visible
	}

	var newY float32
	if rowTop < scrollY {
		newY = rowTop
	} else {
		newY = rowBottom - dataViewH
		if newY < 0 {
			newY = 0
		}
	}
	w.setScrollY(newY)
}

// --- Accessibility ---

// AccessibilityRole returns the ARIA role for this widget.
func (w *Widget) AccessibilityRole() a11y.Role {
	return a11y.RoleTable
}

// AccessibilityLabel returns the accessibility label.
func (w *Widget) AccessibilityLabel() string {
	if w.cfg.a11yLabel != "" {
		return w.cfg.a11yLabel
	}
	return defaultA11yLabel
}

// AccessibilityHint returns the accessibility hint.
func (w *Widget) AccessibilityHint() string {
	return ""
}

// AccessibilityValue returns the current table state as a string.
func (w *Widget) AccessibilityValue() string {
	count := w.cfg.ResolvedRowCount()
	selected := w.cfg.ResolvedSelectedRow()
	if selected >= 0 && selected < count {
		return fmt.Sprintf("Row %d of %d selected", selected+1, count)
	}
	return fmt.Sprintf("%d rows", count)
}

// AccessibilityState returns the current accessibility state.
func (w *Widget) AccessibilityState() a11y.State {
	return a11y.State{
		Disabled: w.cfg.ResolvedDisabled(),
	}
}

// AccessibilityActions returns the list of supported actions.
func (w *Widget) AccessibilityActions() []a11y.Action {
	return []a11y.Action{a11y.ActionScrollUp, a11y.ActionScrollDown}
}

// --- Internal: layout ---

// resolveColumnWidths computes the final pixel width for each column.
func (w *Widget) resolveColumnWidths() {
	cols := w.cfg.columns
	if len(cols) == 0 {
		w.colWidths = nil
		return
	}

	if len(w.colWidths) != len(cols) {
		w.colWidths = make([]float32, len(cols))
	}

	available := w.viewportWidth
	var fixedTotal float32
	flexCount := 0

	for i, col := range cols {
		if col.Width > 0 {
			w.colWidths[i] = col.Width
			fixedTotal += col.Width
		} else {
			flexCount++
		}
	}

	remaining := available - fixedTotal
	if remaining < 0 {
		remaining = 0
	}

	if flexCount > 0 {
		flexWidth := remaining / float32(flexCount)
		for i, col := range cols {
			if col.Width <= 0 {
				minW := col.MinWidth
				if minW <= 0 {
					minW = defaultMinColumnWidth
				}
				w.colWidths[i] = float32(math.Max(float64(flexWidth), float64(minW)))
			}
		}
	}
}

// columnX returns the x offset of the given column index relative to the table left edge.
func (w *Widget) columnX(colIdx int) float32 {
	var x float32
	for i := 0; i < colIdx && i < len(w.colWidths); i++ {
		x += w.colWidths[i]
	}
	return x
}

// totalDataHeight returns the total height of all data rows.
func (w *Widget) totalDataHeight() float32 {
	rowCount := w.cfg.ResolvedRowCount()
	if rowCount <= 0 {
		return 0
	}
	return float32(rowCount) * w.cfg.rowHeight
}

// visibleRowRange returns [firstRow, lastRow] (inclusive) that overlap the data viewport.
func (w *Widget) visibleRowRange(scrollY, viewH float32) (int, int) {
	if w.cfg.rowHeight <= 0 {
		return 0, 0
	}
	rowCount := w.cfg.ResolvedRowCount()
	if rowCount == 0 {
		return 0, 0
	}

	firstRow := int(scrollY / w.cfg.rowHeight)
	if firstRow < 0 {
		firstRow = 0
	}
	lastRow := int((scrollY + viewH) / w.cfg.rowHeight)
	if lastRow >= rowCount {
		lastRow = rowCount - 1
	}
	if firstRow > lastRow {
		firstRow = lastRow
	}
	return firstRow, lastRow
}

// --- Internal: drawing ---

// drawHeader draws the fixed header row.
func (w *Widget) drawHeader(_ widget.Context, canvas widget.Canvas, tableBounds geometry.Rect) {
	headerBounds := geometry.NewRect(
		tableBounds.Min.X,
		tableBounds.Min.Y,
		tableBounds.Width(),
		defaultHeaderHeight,
	)

	hps := HeaderPaintState{
		Disabled: w.cfg.ResolvedDisabled(),
	}
	w.painter.PaintHeader(canvas, headerBounds, hps)

	// Draw each header cell.
	for i, col := range w.cfg.columns {
		if i >= len(w.colWidths) {
			break
		}
		cellX := tableBounds.Min.X + w.columnX(i)
		cellBounds := geometry.NewRect(cellX, tableBounds.Min.Y, w.colWidths[i], defaultHeaderHeight)

		var sortDir SortDirection
		if w.sortColumn == i {
			sortDir = w.sortDirection
		}

		hcs := HeaderCellPaintState{
			Title:    col.Title,
			Align:    col.Align,
			Sortable: col.Sortable,
			SortDir:  sortDir,
			Hovered:  w.hoveredColHdr == i,
			Disabled: w.cfg.ResolvedDisabled(),
		}
		w.painter.PaintHeaderCell(canvas, cellBounds, hcs)
	}

	// Draw bottom divider line.
	dividerY := tableBounds.Min.Y + defaultHeaderHeight
	dividerFrom := geometry.Pt(tableBounds.Min.X, dividerY)
	dividerTo := geometry.Pt(tableBounds.Min.X+tableBounds.Width(), dividerY)
	canvas.DrawLine(dividerFrom, dividerTo, defaultHeaderDividerColor, headerDividerWidth)
}

// Header divider constants.
const headerDividerWidth float32 = 1

var defaultHeaderDividerColor = widget.RGBA(0.8, 0.8, 0.8, 1.0)

// drawVisibleRows renders only the rows currently visible in the viewport.
func (w *Widget) drawVisibleRows(ctx widget.Context, canvas widget.Canvas) {
	rowCount := w.cfg.ResolvedRowCount()
	if rowCount == 0 {
		w.painter.PaintEmptyState(canvas, geometry.NewRect(0, 0, w.viewportWidth, w.viewportHeight-defaultHeaderHeight))
		return
	}

	scrollY := w.currentScrollY()
	dataViewH := w.viewportHeight - defaultHeaderHeight
	if dataViewH <= 0 {
		return
	}

	firstRow, lastRow := w.visibleRowRange(scrollY, dataViewH)
	selectedRow := w.cfg.ResolvedSelectedRow()

	for row := firstRow; row <= lastRow; row++ {
		// Content-space Y: the parent ScrollView already applies a
		// translate(0, -scrollY) transform before calling Draw, so rows must
		// be placed at their absolute content offset. Subtracting scrollY here
		// too would double-scroll (rows move 2x, content runs off the end and
		// shows blank space; hit-testing via rowAtY desyncs from the highlight).
		y := float32(row) * w.cfg.rowHeight
		rowBounds := geometry.NewRect(0, y, w.viewportWidth, w.cfg.rowHeight)

		isSelected := w.isRowSelected(row, selectedRow)

		rps := RowPaintState{
			Bounds:   rowBounds,
			RowIndex: row,
			Selected: isSelected,
			Focused:  isSelected && w.IsFocused(),
			Hovered:  row == w.hoveredRow,
			Disabled: w.cfg.ResolvedDisabled(),
		}
		w.painter.PaintRow(canvas, rps)

		// Draw cells.
		w.drawRowCells(ctx, canvas, row, rowBounds, isSelected)
	}
}

// drawRowCells draws each cell in a single row.
func (w *Widget) drawRowCells(_ widget.Context, canvas widget.Canvas, row int, rowBounds geometry.Rect, selected bool) {
	for i, col := range w.cfg.columns {
		if i >= len(w.colWidths) {
			break
		}
		cellX := w.columnX(i)
		cellBounds := geometry.NewRect(
			rowBounds.Min.X+cellX,
			rowBounds.Min.Y,
			w.colWidths[i],
			rowBounds.Height(),
		)

		var value string
		if w.cfg.cellValue != nil {
			value = w.cfg.cellValue(row, col.Key)
		}

		cps := CellPaintState{
			Bounds:   cellBounds,
			Value:    value,
			Align:    col.Align,
			RowIndex: row,
			ColIndex: i,
			Selected: selected,
			Disabled: w.cfg.ResolvedDisabled(),
		}
		w.painter.PaintCell(canvas, cps)
	}
}

// isRowSelected reports whether a row is selected.
func (w *Widget) isRowSelected(row, primarySelected int) bool {
	if w.cfg.selectionMode == SelectionMulti {
		return w.cfg.selectedRows[row]
	}
	return row == primarySelected
}

// --- Internal: scroll helpers ---

func (w *Widget) currentScrollY() float32 {
	_, y := w.scroll.ScrollOffset()
	return y
}

func (w *Widget) setScrollY(y float32) {
	if w.cfg.scrollYSignal != nil {
		w.cfg.scrollYSignal.Set(y)
	}
}

// --- Internal: event handling ---

// handleHeaderMouseEvent processes mouse events on the header row.
func (w *Widget) handleHeaderMouseEvent(ctx widget.Context, e *event.MouseEvent) bool {
	bounds := w.Bounds()
	headerBottom := bounds.Min.Y + defaultHeaderHeight

	switch e.MouseType {
	case event.MouseMove:
		return w.handleHeaderMouseMove(ctx, e, bounds, headerBottom)
	case event.MousePress:
		return w.handleHeaderMousePress(ctx, e, bounds, headerBottom)
	case event.MouseLeave:
		if w.hoveredColHdr != noHoveredCol {
			w.hoveredColHdr = noHoveredCol
			// ADR-028: visual only — header hover cleared.
			w.SetNeedsRedraw(true)
			ctx.InvalidateRect(w.Bounds())
		}
		return false
	default:
		return false
	}
}

func (w *Widget) handleHeaderMouseMove(ctx widget.Context, e *event.MouseEvent, bounds geometry.Rect, headerBottom float32) bool {
	// Only process events in header area.
	if e.Position.Y < bounds.Min.Y || e.Position.Y >= headerBottom {
		if w.hoveredColHdr != noHoveredCol {
			w.hoveredColHdr = noHoveredCol
			// ADR-028: visual only — header hover cleared.
			w.SetNeedsRedraw(true)
			ctx.InvalidateRect(w.Bounds())
		}
		return false
	}

	colIdx := w.columnAtX(e.Position.X - bounds.Min.X)
	if colIdx != w.hoveredColHdr {
		w.hoveredColHdr = colIdx
		// ADR-028: visual only — header column hover changed.
		w.SetNeedsRedraw(true)
		ctx.InvalidateRect(w.Bounds())
	}

	// Show pointer cursor for sortable columns.
	if colIdx >= 0 && colIdx < len(w.cfg.columns) && w.cfg.columns[colIdx].Sortable {
		ctx.SetCursor(widget.CursorPointer)
	}

	return false // don't consume moves
}

func (w *Widget) handleHeaderMousePress(ctx widget.Context, e *event.MouseEvent, bounds geometry.Rect, headerBottom float32) bool {
	if e.Button != event.ButtonLeft {
		return false
	}
	if e.Position.Y < bounds.Min.Y || e.Position.Y >= headerBottom {
		return false
	}
	if w.cfg.ResolvedDisabled() {
		return false
	}

	colIdx := w.columnAtX(e.Position.X - bounds.Min.X)
	if colIdx < 0 || colIdx >= len(w.cfg.columns) {
		return false
	}
	col := w.cfg.columns[colIdx]
	if !col.Sortable {
		return false
	}

	// Cycle sort direction.
	if w.sortColumn == colIdx {
		w.sortDirection = w.sortDirection.nextDirection()
		if w.sortDirection == SortNone {
			w.sortColumn = noSortColumn
		}
	} else {
		w.sortColumn = colIdx
		w.sortDirection = SortAscending
	}

	// Invoke callback.
	if w.sortDirection != SortNone && w.cfg.onSort != nil {
		w.cfg.onSort(col.Key, w.sortDirection == SortAscending)
	}

	ctx.RequestFocus(w)
	// ADR-028: layout change — sort reorders rows, may change content.
	ctx.Invalidate()
	return true
}

// columnAtX returns the column index at the given x offset (relative to table left).
// Returns -1 if no column matches.
func (w *Widget) columnAtX(x float32) int {
	if x < 0 {
		return noHoveredCol
	}
	var cumX float32
	for i, cw := range w.colWidths {
		cumX += cw
		if x < cumX {
			return i
		}
	}
	return noHoveredCol
}

// handleKeyEvent processes keyboard events for row navigation.
func (w *Widget) handleKeyEvent(ctx widget.Context, e *event.KeyEvent) bool {
	if !w.IsFocused() {
		return false
	}
	if e.KeyType != event.KeyPress && e.KeyType != event.KeyRepeat {
		return false
	}
	if w.cfg.ResolvedDisabled() {
		return false
	}
	rowCount := w.cfg.ResolvedRowCount()
	if rowCount == 0 {
		return false
	}

	selectedRow := w.cfg.ResolvedSelectedRow()

	switch e.Key {
	case event.KeyDown:
		return w.moveSelection(ctx, selectedRow+1, rowCount)
	case event.KeyUp:
		return w.moveSelection(ctx, selectedRow-1, rowCount)
	case event.KeyHome:
		return w.moveSelection(ctx, 0, rowCount)
	case event.KeyEnd:
		return w.moveSelection(ctx, rowCount-1, rowCount)
	case event.KeyPageDown:
		return w.moveSelectionByPage(ctx, selectedRow, rowCount, 1)
	case event.KeyPageUp:
		return w.moveSelectionByPage(ctx, selectedRow, rowCount, -1)
	case event.KeyEnter, event.KeySpace:
		if selectedRow >= 0 && selectedRow < rowCount && w.cfg.onRowSelect != nil {
			w.cfg.onRowSelect(selectedRow)
			return true
		}
		return false
	default:
		return false
	}
}

func (w *Widget) moveSelection(ctx widget.Context, newRow, count int) bool {
	if w.cfg.selectionMode == SelectionNone {
		return false
	}
	if newRow < 0 {
		newRow = 0
	}
	if newRow >= count {
		newRow = count - 1
	}
	w.setSelectedRow(ctx, newRow)
	w.ScrollToRow(newRow)
	return true
}

func (w *Widget) moveSelectionByPage(ctx widget.Context, currentRow, count, direction int) bool {
	if w.cfg.selectionMode == SelectionNone {
		return false
	}
	dataViewH := w.viewportHeight - defaultHeaderHeight
	if dataViewH <= 0 {
		return false
	}
	rowsPerPage := int(dataViewH / w.cfg.rowHeight)
	if rowsPerPage < 1 {
		rowsPerPage = 1
	}
	newRow := currentRow + direction*rowsPerPage
	return w.moveSelection(ctx, newRow, count)
}

func (w *Widget) setSelectedRow(ctx widget.Context, row int) {
	current := w.cfg.ResolvedSelectedRow()
	if row == current {
		return
	}

	if w.cfg.selectedRowSignal != nil {
		w.cfg.selectedRowSignal.Set(row)
	} else {
		w.cfg.selectedRow = row
	}

	if w.cfg.onRowSelect != nil {
		w.cfg.onRowSelect(row)
	}

	// ADR-028: visual only — row selection highlight moved.
	w.SetNeedsRedraw(true)
	ctx.InvalidateRect(w.Bounds())
}

// handleContentMouseEvent processes mouse events on the data area.
func handleContentMouseEvent(dt *Widget, ctx widget.Context, e *event.MouseEvent) bool {
	if dt.cfg.ResolvedDisabled() {
		return false
	}

	switch e.MouseType {
	case event.MouseMove:
		return handleContentMouseMove(dt, ctx, e)
	case event.MousePress:
		return handleContentMousePress(dt, ctx, e)
	case event.MouseLeave:
		if dt.hoveredRow != noHoveredRow {
			dt.hoveredRow = noHoveredRow
			// ADR-028: visual only — row hover cleared.
			dt.SetNeedsRedraw(true)
			ctx.InvalidateRect(dt.Bounds())
		}
		return false
	default:
		return false
	}
}

// handleContentMouseMove updates the hovered row index based on mouse position.
// The event position is already in content space (transformed by ScrollView).
func handleContentMouseMove(dt *Widget, ctx widget.Context, e *event.MouseEvent) bool {
	// Position is already in content space — ScrollView applies the inverse
	// of its Draw transform before dispatching to content children.
	row := dt.rowAtY(e.Position.Y)

	if row != dt.hoveredRow {
		dt.hoveredRow = row
		// ADR-028: visual only — row hover changed.
		dt.SetNeedsRedraw(true)
		ctx.InvalidateRect(dt.Bounds())
	}
	return false
}

// handleContentMousePress handles item click for row selection.
// The event position is already in content space (transformed by ScrollView).
func handleContentMousePress(dt *Widget, ctx widget.Context, e *event.MouseEvent) bool {
	if e.Button != event.ButtonLeft {
		return false
	}

	// Position is already in content space — ScrollView applies the inverse
	// of its Draw transform before dispatching to content children.
	row := dt.rowAtY(e.Position.Y)

	if row < 0 {
		return false
	}

	applyRowSelection(dt, ctx, row, e.Modifiers().Has(event.ModCtrl))

	ctx.RequestFocus(dt)
	return true
}

// applyRowSelection updates selection state after a row click.
func applyRowSelection(dt *Widget, ctx widget.Context, row int, ctrlHeld bool) {
	if dt.cfg.selectionMode == SelectionMulti && ctrlHeld {
		toggleMultiSelect(dt, ctx, row)
		return
	}
	if dt.cfg.selectionMode == SelectionNone {
		return
	}
	// Single click (or multi without Ctrl) — select one row.
	if dt.cfg.selectionMode == SelectionMulti {
		dt.cfg.selectedRows = make(map[int]bool)
		dt.cfg.selectedRows[row] = true
	}
	dt.setSelectedRow(ctx, row)
}

// toggleMultiSelect toggles a row in multi-selection mode.
func toggleMultiSelect(dt *Widget, ctx widget.Context, row int) {
	if dt.cfg.selectedRows[row] {
		delete(dt.cfg.selectedRows, row)
	} else {
		dt.cfg.selectedRows[row] = true
	}
	if dt.cfg.onRowSelect != nil {
		dt.cfg.onRowSelect(row)
	}
	// ADR-028: visual only — multi-selection highlight toggled.
	dt.SetNeedsRedraw(true)
	ctx.InvalidateRect(dt.Bounds())
}

// rowAtY returns the row index at the given y offset in content coordinates.
// Returns -1 if no row matches.
func (w *Widget) rowAtY(y float32) int {
	if y < 0 || w.cfg.rowHeight <= 0 {
		return noHoveredRow
	}
	row := int(y / w.cfg.rowHeight)
	if row >= w.cfg.ResolvedRowCount() {
		return noHoveredRow
	}
	return row
}

// --- Virtual Content ---

// virtualContent is an internal widget representing the scrollable data area.
type virtualContent struct {
	widget.WidgetBase
	table *Widget
}

// Layout returns the total content size (all rows).
func (vc *virtualContent) Layout(_ widget.Context, c geometry.Constraints) geometry.Size {
	if vc.table == nil {
		return geometry.Size{}
	}
	totalHeight := vc.table.totalDataHeight()
	width := c.MaxWidth
	if width >= geometry.Infinity {
		width = c.MinWidth
	}
	if width < c.MinWidth {
		width = c.MinWidth
	}
	return geometry.Sz(width, totalHeight)
}

// Draw renders only the visible data rows.
func (vc *virtualContent) Draw(ctx widget.Context, canvas widget.Canvas) {
	if vc.table == nil {
		return
	}
	vc.table.drawVisibleRows(ctx, canvas)
}

// Event delegates events back to the parent table for row interaction.
func (vc *virtualContent) Event(ctx widget.Context, e event.Event) bool {
	if vc.table == nil {
		return false
	}
	me, ok := e.(*event.MouseEvent)
	if !ok {
		return false
	}
	return handleContentMouseEvent(vc.table, ctx, me)
}

// Children returns nil; rows are drawn directly, not as child widgets.
func (vc *virtualContent) Children() []widget.Widget {
	return nil
}

// Default viewport dimensions used as fallback.
const (
	defaultViewportWidth  float32 = 600
	defaultViewportHeight float32 = 400
	defaultA11yLabel              = "Data Table"
)

// Verify Widget implements required interfaces at compile time.
var (
	_ widget.Widget    = (*Widget)(nil)
	_ widget.Focusable = (*Widget)(nil)
	_ widget.Lifecycle = (*Widget)(nil)
	_ a11y.Accessible  = (*Widget)(nil)
)
