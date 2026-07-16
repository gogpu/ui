package gridview

import (
	"fmt"
	"math"

	"github.com/gogpu/ui/a11y"
	"github.com/gogpu/ui/cdk"
	"github.com/gogpu/ui/core/scrollview"
	"github.com/gogpu/ui/event"
	"github.com/gogpu/ui/geometry"
	"github.com/gogpu/ui/state"
	"github.com/gogpu/ui/widget"
)

// SelectionMode defines how cells can be selected in the grid.
type SelectionMode uint8

// SelectionMode constants.
const (
	// SelectionNone disables cell selection. This is the default.
	SelectionNone SelectionMode = iota

	// SelectionSingle allows at most one cell to be selected at a time.
	SelectionSingle
)

// String returns a human-readable name for the selection mode.
func (m SelectionMode) String() string {
	switch m {
	case SelectionNone:
		return selectionNoneStr
	case SelectionSingle:
		return selectionSingleStr
	default:
		return selectionUnknownStr
	}
}

// Selection mode string constants.
const (
	selectionNoneStr    = "None"
	selectionSingleStr  = "Single"
	selectionUnknownStr = "Unknown"
)

// CellContext provides contextual information to the cell builder callback.
//
// The builder receives a CellContext for each visible cell, allowing it to
// customize the returned widget based on selection, hover, and position state.
type CellContext struct {
	// Index is the zero-based item index in the data source.
	Index int

	// Row is the cell's row index (zero-based).
	Row int

	// Col is the cell's column index (zero-based).
	Col int

	// IsSelected is true if this cell is the currently selected cell.
	IsSelected bool

	// IsHovered is true if the mouse cursor is over this cell.
	IsHovered bool
}

// --- Config ---

// config holds the grid view's configuration, set at construction time via options.
type config struct {
	// Item data.
	itemCount               int
	itemCountFn             func() int
	itemCountSignal         state.Signal[int]
	readonlyItemCountSignal state.ReadonlySignal[int]

	// Cell content -- the Content[CellContext] that renders each visible cell.
	// Set via BuildCell (convenience) or CellContent (direct Content[C]).
	cellContent cdk.Content[CellContext]

	// Grid layout.
	columns               int
	columnsAuto           bool
	columnsSignal         state.Signal[int]
	readonlyColumnsSignal state.ReadonlySignal[int]
	itemWidth             float32
	itemHeight            float32
	gap                   float32

	// Selection.
	selectionMode               SelectionMode
	selectedIndex               int
	selectedIndexSignal         state.Signal[int]
	readonlySelectedIndexSignal state.ReadonlySignal[int]
	onCellClick                 func(index int)
	onSelectionChange           func(index int)

	// Disabled state.
	disabled               bool
	disabledFn             func() bool
	disabledSignal         state.Signal[bool]
	readonlyDisabledSignal state.ReadonlySignal[bool]

	// Scroll pass-through.
	scrollYSignal state.Signal[float32]
	onScroll      func(offset float32)

	// Visual options.
	painter Painter

	// Accessibility.
	a11yLabel string
}

// defaultConfig returns a config with sensible defaults.
func defaultConfig() config {
	return config{
		columns:       defaultColumns,
		itemWidth:     defaultItemWidth,
		itemHeight:    defaultItemHeight,
		selectedIndex: noSelection,
	}
}

// Default configuration values.
const (
	defaultColumns    int     = 4
	defaultItemWidth  float32 = 120
	defaultItemHeight float32 = 120
	noSelection       int     = -1
	noHoveredIndex    int     = -1
)

// ResolvedItemCount returns the current item count.
// Priority: ReadonlySignal > Signal > Fn > Static.
func (c *config) ResolvedItemCount() int {
	if c.readonlyItemCountSignal != nil {
		return c.readonlyItemCountSignal.Get()
	}
	if c.itemCountSignal != nil {
		return c.itemCountSignal.Get()
	}
	if c.itemCountFn != nil {
		return c.itemCountFn()
	}
	return c.itemCount
}

// ResolvedColumns returns the current column count.
// Priority: ReadonlySignal > Signal > Static.
func (c *config) ResolvedColumns() int {
	if c.readonlyColumnsSignal != nil {
		return c.readonlyColumnsSignal.Get()
	}
	if c.columnsSignal != nil {
		return c.columnsSignal.Get()
	}
	return c.columns
}

// ResolvedSelectedIndex returns the current selected index.
// Priority: ReadonlySignal > Signal > Static.
// Returns -1 if no selection.
func (c *config) ResolvedSelectedIndex() int {
	if c.readonlySelectedIndexSignal != nil {
		return c.readonlySelectedIndexSignal.Get()
	}
	if c.selectedIndexSignal != nil {
		return c.selectedIndexSignal.Get()
	}
	return c.selectedIndex
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

// Option configures a grid view during construction.
type Option func(*config)

// ItemCount sets the total number of items in the grid.
func ItemCount(n int) Option {
	return func(c *config) { c.itemCount = n }
}

// ItemCountFn sets a dynamic function that returns the item count.
// When set, this takes precedence over [ItemCount] but not over [ItemCountSignal].
func ItemCountFn(fn func() int) Option {
	return func(c *config) { c.itemCountFn = fn }
}

// ItemCountSignal binds the item count to a reactive signal.
// When set, the signal takes precedence over [ItemCountFn] and [ItemCount]
// but not over [ItemCountReadonlySignal].
func ItemCountSignal(sig state.Signal[int]) Option {
	return func(c *config) { c.itemCountSignal = sig }
}

// ItemCountReadonlySignal binds the item count to a read-only signal.
// When set, this takes highest precedence over all other item count sources.
func ItemCountReadonlySignal(sig state.ReadonlySignal[int]) Option {
	return func(c *config) { c.readonlyItemCountSignal = sig }
}

// BuildCell sets the callback that creates a widget for each visible cell.
// The callback receives the item index and a [CellContext] with the cell's state.
// This is the primary convenience API for providing grid content.
//
// Internally, the function is wrapped into a [cdk.FuncContent] so that the
// grid view uniformly operates on [cdk.Content] for cell rendering.
func BuildCell(fn func(index int, ctx CellContext) widget.Widget) Option {
	return func(c *config) {
		c.cellContent = cdk.FuncContent[CellContext]{Fn: func(ctx CellContext) widget.Widget {
			return fn(ctx.Index, ctx)
		}}
	}
}

// CellContent sets the [cdk.Content] used to render each visible cell.
// This is the enterprise API for providing grid content -- use it when you
// need reusable content implementations or progressive complexity.
//
// For most use cases, [BuildCell] is simpler and sufficient.
func CellContent(content cdk.Content[CellContext]) Option {
	return func(c *config) { c.cellContent = content }
}

// Columns sets the fixed number of columns in the grid.
// Default: 4. See also [ColumnsAuto] for auto-fit mode.
func Columns(n int) Option {
	return func(c *config) {
		c.columns = n
		c.columnsAuto = false
	}
}

// ColumnsAuto enables auto-fit column count based on viewport width.
// When enabled, the column count is calculated as floor(viewportWidth / (itemWidth + gap)).
func ColumnsAuto(enabled bool) Option {
	return func(c *config) { c.columnsAuto = enabled }
}

// ColumnsSignal binds the column count to a reactive signal.
// When set, the signal takes precedence over [Columns] but not over
// [ColumnsReadonlySignal]. Ignored if [ColumnsAuto] is enabled.
func ColumnsSignal(sig state.Signal[int]) Option {
	return func(c *config) { c.columnsSignal = sig }
}

// ColumnsReadonlySignal binds the column count to a read-only signal.
// When set, this takes highest precedence. Ignored if [ColumnsAuto] is enabled.
func ColumnsReadonlySignal(sig state.ReadonlySignal[int]) Option {
	return func(c *config) { c.readonlyColumnsSignal = sig }
}

// ItemSize sets the fixed width and height for all grid cells.
// Default: 120x120.
func ItemSize(width, height float32) Option {
	return func(c *config) {
		c.itemWidth = width
		c.itemHeight = height
	}
}

// Gap sets the spacing between cells in pixels. Default: 0.
func Gap(g float32) Option {
	return func(c *config) { c.gap = g }
}

// SelectionModeOpt sets the selection mode for the grid.
// Default is [SelectionNone].
func SelectionModeOpt(mode SelectionMode) Option {
	return func(c *config) { c.selectionMode = mode }
}

// SelectedIndex sets the initially selected cell index.
// Use -1 for no selection.
func SelectedIndex(index int) Option {
	return func(c *config) { c.selectedIndex = index }
}

// SelectedIndexSignal binds the selected index to a reactive signal.
// This is a TWO-WAY binding: the widget reads from the signal, and
// when the user selects a cell, the new index is written back.
func SelectedIndexSignal(sig state.Signal[int]) Option {
	return func(c *config) { c.selectedIndexSignal = sig }
}

// SelectedIndexReadonlySignal binds the selected index to a read-only signal.
// When set, this takes highest precedence over all other selection sources.
func SelectedIndexReadonlySignal(sig state.ReadonlySignal[int]) Option {
	return func(c *config) { c.readonlySelectedIndexSignal = sig }
}

// OnCellClick sets the callback invoked when a cell is clicked.
func OnCellClick(fn func(index int)) Option {
	return func(c *config) { c.onCellClick = fn }
}

// OnSelectionChange sets the callback invoked when the selected cell changes.
func OnSelectionChange(fn func(index int)) Option {
	return func(c *config) { c.onSelectionChange = fn }
}

// Disabled sets the grid view's disabled state.
func Disabled(d bool) Option {
	return func(c *config) { c.disabled = d }
}

// DisabledFn sets a dynamic function for the disabled state.
// When set, this takes precedence over [Disabled] but not over [DisabledSignal].
func DisabledFn(fn func() bool) Option {
	return func(c *config) { c.disabledFn = fn }
}

// DisabledSignal binds the disabled state to a reactive signal.
// When set, the signal takes precedence over [DisabledFn] and [Disabled]
// but not over [DisabledReadonlySignal].
func DisabledSignal(sig state.Signal[bool]) Option {
	return func(c *config) { c.disabledSignal = sig }
}

// DisabledReadonlySignal binds the disabled state to a read-only signal.
// When set, this takes highest precedence over all other disabled sources.
func DisabledReadonlySignal(sig state.ReadonlySignal[bool]) Option {
	return func(c *config) { c.readonlyDisabledSignal = sig }
}

// ScrollYSignal binds the vertical scroll offset to a reactive signal.
// This is a TWO-WAY binding passed through to the internal ScrollView.
func ScrollYSignal(sig state.Signal[float32]) Option {
	return func(c *config) { c.scrollYSignal = sig }
}

// OnScroll sets the callback invoked when the scroll position changes.
func OnScroll(fn func(offset float32)) Option {
	return func(c *config) { c.onScroll = fn }
}

// PainterOpt sets the painter used to render grid-specific visuals.
// Each design system provides its own painter. If not set,
// [DefaultPainter] is used.
func PainterOpt(p Painter) Option {
	return func(c *config) { c.painter = p }
}

// A11yLabel sets the accessibility label for the grid.
func A11yLabel(label string) Option {
	return func(c *config) { c.a11yLabel = label }
}

// --- Widget ---

// Widget implements a virtualized grid that renders only visible cells.
// It composes an internal [scrollview.Widget] for scroll handling and
// delegates cell rendering to a builder callback.
//
// A grid view is created with [New] using functional options:
//
//	gv := gridview.New(
//	    gridview.Columns(4),
//	    gridview.ItemCount(1000),
//	    gridview.ItemSize(120, 120),
//	    gridview.BuildCell(func(index int, ctx gridview.CellContext) widget.Widget {
//	        return buildCell(index, ctx.IsSelected)
//	    }),
//	)
type Widget struct {
	widget.WidgetBase
	cfg     config
	painter Painter

	// Internal scroll view (composition).
	scroll  *scrollview.Widget
	virtual *virtualContent

	// Widget cache for visible cells.
	cache cellCache

	// Layout state.
	viewportWidth  float32
	viewportHeight float32
	effectiveCols  int // actual columns after auto-fit

	// Interaction state.
	hoveredIndex int
}

// New creates a new grid view Widget with the given options.
//
// The returned widget is visible, enabled, and focusable by default.
// If no [BuildCell] callback is provided, the grid renders empty.
func New(opts ...Option) *Widget {
	cfg := defaultConfig()
	for _, opt := range opts {
		opt(&cfg)
	}

	w := &Widget{
		cfg:          cfg,
		painter:      DefaultPainter{},
		hoveredIndex: noHoveredIndex,
	}
	w.SetVisible(true)
	w.SetEnabled(true)

	if w.cfg.painter != nil {
		w.painter = w.cfg.painter
	}

	// Create virtual content widget.
	w.virtual = &virtualContent{grid: w}
	w.virtual.SetVisible(true)

	// Build internal scroll view options.
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
			w.cache.invalidate()
		}))
	} else {
		svOpts = append(svOpts, scrollview.OnScroll(func(_, _ float32) {
			w.cache.invalidate()
		}))
	}

	w.scroll = scrollview.New(w.virtual, svOpts...)

	// ADR-028: parent chain for upward dirty propagation.
	// Flutter: RenderObject.adoptChild sets parent on each child.
	w.scroll.SetParent(w)

	return w
}

// IsFocusable reports whether the grid view can currently receive focus.
func (w *Widget) IsFocusable() bool {
	return w.IsVisible() && w.IsEnabled() && !w.cfg.ResolvedDisabled()
}

// Layout calculates the grid view's size within the given constraints.
func (w *Widget) Layout(ctx widget.Context, constraints geometry.Constraints) geometry.Size {
	size := constraints.Biggest()
	if size.Width >= geometry.Infinity {
		size.Width = constraints.Constrain(geometry.Sz(defaultViewportWidth, 0)).Width
	}
	if size.Height >= geometry.Infinity {
		totalH := w.totalContentHeight()
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

	// Compute effective column count.
	w.effectiveCols = w.computeEffectiveColumns()

	// Layout the internal scroll view with concrete constraints.
	svConstraints := geometry.Tight(size)
	widget.LayoutChild(w.scroll, ctx, svConstraints)

	return size
}

// Draw renders the grid view to the canvas.
func (w *Widget) Draw(ctx widget.Context, canvas widget.Canvas) {
	if !w.IsVisible() {
		return
	}
	bounds := w.Bounds()
	if bounds.IsEmpty() {
		return
	}

	// Set scroll view bounds to match our bounds.
	w.scroll.SetBounds(bounds)

	// Stamp screen origin on the internal scroll view so its ScreenBounds()
	// returns correct window-space coordinates for dirty region collection.
	widget.StampScreenOrigin(w.scroll, canvas)

	// Delegate drawing to the internal scroll view.
	w.scroll.Draw(ctx, canvas)
}

// Event handles an input event and returns true if consumed.
func (w *Widget) Event(ctx widget.Context, e event.Event) bool {
	if !w.IsVisible() || !w.IsEnabled() {
		return false
	}

	// Handle keyboard events at the grid level.
	if ke, ok := e.(*event.KeyEvent); ok {
		if w.handleKeyEvent(ctx, ke) {
			return true
		}
	}

	// Delegate other events to the scroll view.
	return w.scroll.Event(ctx, e)
}

// Children returns the internal scroll view as the single child.
func (w *Widget) Children() []widget.Widget {
	if w.scroll == nil {
		return nil
	}
	return []widget.Widget{w.scroll}
}

// Mount creates signal bindings for push-based invalidation.
// Implements [widget.Lifecycle].
func (w *Widget) Mount(ctx widget.Context) {
	sched := ctx.Scheduler()
	if sched == nil {
		return
	}

	// Bind item count signals.
	if w.cfg.readonlyItemCountSignal != nil {
		b := state.BindToSchedulerLayout(w.cfg.readonlyItemCountSignal, w, sched)
		w.AddBinding(b)
	} else if w.cfg.itemCountSignal != nil {
		b := state.BindToSchedulerLayout(w.cfg.itemCountSignal, w, sched)
		w.AddBinding(b)
	}

	// Bind selected index signals.
	if w.cfg.readonlySelectedIndexSignal != nil {
		b := state.BindToScheduler(w.cfg.readonlySelectedIndexSignal, w, sched)
		w.AddBinding(b)
	} else if w.cfg.selectedIndexSignal != nil {
		b := state.BindToScheduler(w.cfg.selectedIndexSignal, w, sched)
		w.AddBinding(b)
	}

	// Bind disabled signals.
	if w.cfg.readonlyDisabledSignal != nil {
		b := state.BindToScheduler(w.cfg.readonlyDisabledSignal, w, sched)
		w.AddBinding(b)
	} else if w.cfg.disabledSignal != nil {
		b := state.BindToScheduler(w.cfg.disabledSignal, w, sched)
		w.AddBinding(b)
	}

	// Bind columns signals.
	if w.cfg.readonlyColumnsSignal != nil {
		b := state.BindToSchedulerLayout(w.cfg.readonlyColumnsSignal, w, sched)
		w.AddBinding(b)
	} else if w.cfg.columnsSignal != nil {
		b := state.BindToSchedulerLayout(w.cfg.columnsSignal, w, sched)
		w.AddBinding(b)
	}

	// Mount internal scroll view.
	w.scroll.Mount(ctx)
}

// Unmount is called when the grid view is removed from the widget tree.
// Implements [widget.Lifecycle].
func (w *Widget) Unmount() {
	w.scroll.Unmount()
	// Bindings are cleaned up automatically by WidgetBase.CleanupBindings().
}

// --- Public API ---

// ScrollToIndex scrolls to make the cell at the given index visible.
// If the cell is already fully visible, this is a no-op.
func (w *Widget) ScrollToIndex(index int) {
	itemCount := w.cfg.ResolvedItemCount()
	if index < 0 || index >= itemCount {
		return
	}

	cols := w.effectiveColumns()
	if cols <= 0 {
		return
	}
	row := index / cols
	cellTop := float32(row) * (w.cfg.itemHeight + w.cfg.gap)
	cellBottom := cellTop + w.cfg.itemHeight
	scrollY := w.currentScrollY()

	// Already visible: no-op.
	if cellTop >= scrollY && cellBottom <= scrollY+w.viewportHeight {
		return
	}

	var newScrollY float32
	if cellTop < scrollY {
		newScrollY = cellTop
	} else {
		newScrollY = cellBottom - w.viewportHeight
		if newScrollY < 0 {
			newScrollY = 0
		}
	}

	w.setScrollY(newScrollY)
}

// VisibleRange returns the indices of currently visible cells [start, end).
func (w *Widget) VisibleRange() (start, end int) {
	cols := w.effectiveColumns()
	if cols <= 0 {
		return 0, 0
	}
	itemCount := w.cfg.ResolvedItemCount()
	if itemCount == 0 {
		return 0, 0
	}

	scrollY := w.currentScrollY()
	firstRow, lastRow := w.visibleRowRange(scrollY, w.viewportHeight)

	startIdx := firstRow * cols
	endIdx := (lastRow + 1) * cols
	if endIdx > itemCount {
		endIdx = itemCount
	}
	if startIdx > itemCount {
		startIdx = itemCount
	}
	return startIdx, endIdx
}

// InvalidateData signals that the underlying data has changed.
// This invalidates the widget cache and triggers re-layout.
func (w *Widget) InvalidateData() {
	w.cache.invalidate()
}

// GetItemCount returns the current item count.
func (w *Widget) GetItemCount() int {
	return w.cfg.ResolvedItemCount()
}

// GetColumns returns the current effective column count.
func (w *Widget) GetColumns() int {
	return w.effectiveColumns()
}

// --- Accessibility ---

// AccessibilityRole returns the ARIA role for this widget.
func (w *Widget) AccessibilityRole() a11y.Role {
	return a11y.RoleGrid
}

// AccessibilityLabel returns the accessibility label.
func (w *Widget) AccessibilityLabel() string {
	if w.cfg.a11yLabel != "" {
		return w.cfg.a11yLabel
	}
	return "Grid"
}

// AccessibilityHint returns the accessibility hint.
func (w *Widget) AccessibilityHint() string {
	return ""
}

// AccessibilityValue returns the current grid state as a string.
func (w *Widget) AccessibilityValue() string {
	count := w.cfg.ResolvedItemCount()
	selected := w.cfg.ResolvedSelectedIndex()
	if selected >= 0 && selected < count {
		return fmt.Sprintf("Item %d of %d selected", selected+1, count)
	}
	return fmt.Sprintf("%d items", count)
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

// --- Internal helpers ---

// effectiveColumns returns the current column count, possibly computed from auto-fit.
func (w *Widget) effectiveColumns() int {
	if w.effectiveCols > 0 {
		return w.effectiveCols
	}
	return w.computeEffectiveColumns()
}

// computeEffectiveColumns computes the column count based on config and viewport.
func (w *Widget) computeEffectiveColumns() int {
	if w.cfg.columnsAuto && w.viewportWidth > 0 && w.cfg.itemWidth > 0 {
		cellStep := w.cfg.itemWidth + w.cfg.gap
		if cellStep <= 0 {
			return 1
		}
		cols := int(w.viewportWidth / cellStep)
		if cols < 1 {
			cols = 1
		}
		return cols
	}
	cols := w.cfg.ResolvedColumns()
	if cols < 1 {
		cols = 1
	}
	return cols
}

// totalContentHeight returns the total height of the grid content.
func (w *Widget) totalContentHeight() float32 {
	itemCount := w.cfg.ResolvedItemCount()
	if itemCount == 0 {
		return 0
	}
	cols := w.effectiveColumns()
	if cols <= 0 {
		cols = 1
	}
	rows := ceilDiv(itemCount, cols)
	if rows <= 0 {
		return 0
	}
	return float32(rows)*w.cfg.itemHeight + float32(rows-1)*w.cfg.gap
}

// visibleRowRange returns [firstRow, lastRow] (inclusive) that overlap the viewport.
func (w *Widget) visibleRowRange(scrollY, viewportH float32) (int, int) {
	cellStep := w.cfg.itemHeight + w.cfg.gap
	if cellStep <= 0 {
		return 0, 0
	}

	itemCount := w.cfg.ResolvedItemCount()
	cols := w.effectiveColumns()
	if cols <= 0 {
		cols = 1
	}
	totalRows := ceilDiv(itemCount, cols)
	if totalRows == 0 {
		return 0, 0
	}

	firstRow := int(scrollY / cellStep)
	if firstRow < 0 {
		firstRow = 0
	}

	lastRow := int((scrollY + viewportH) / cellStep)
	if lastRow >= totalRows {
		lastRow = totalRows - 1
	}
	if firstRow > lastRow {
		firstRow = lastRow
	}

	return firstRow, lastRow
}

// cellIndexAtPoint converts a point in content coordinates to a cell index.
// Returns -1 if the point is not on a cell.
func (w *Widget) cellIndexAtPoint(contentX, contentY float32) int {
	cols := w.effectiveColumns()
	if cols <= 0 {
		return noHoveredIndex
	}

	cellStepX := w.cfg.itemWidth + w.cfg.gap
	cellStepY := w.cfg.itemHeight + w.cfg.gap
	if cellStepX <= 0 || cellStepY <= 0 {
		return noHoveredIndex
	}

	if contentX < 0 || contentY < 0 {
		return noHoveredIndex
	}

	col := int(contentX / cellStepX)
	row := int(contentY / cellStepY)

	// Check if the point is within a cell (not in the gap).
	cellLocalX := contentX - float32(col)*cellStepX
	cellLocalY := contentY - float32(row)*cellStepY
	if cellLocalX > w.cfg.itemWidth || cellLocalY > w.cfg.itemHeight {
		return noHoveredIndex
	}

	if col >= cols || row < 0 {
		return noHoveredIndex
	}

	index := row*cols + col
	if index >= w.cfg.ResolvedItemCount() {
		return noHoveredIndex
	}

	return index
}

// currentScrollY returns the current vertical scroll offset.
func (w *Widget) currentScrollY() float32 {
	_, y := w.scroll.ScrollOffset()
	return y
}

// setScrollY updates the scroll Y position.
func (w *Widget) setScrollY(y float32) {
	if w.cfg.scrollYSignal != nil {
		w.cfg.scrollYSignal.Set(y)
	}
}

// setSelectedIndex updates the selected index, writing back to signal if bound.
func (w *Widget) setSelectedIndex(ctx widget.Context, index int) {
	current := w.cfg.ResolvedSelectedIndex()
	if index == current {
		return
	}

	if w.cfg.selectedIndexSignal != nil {
		w.cfg.selectedIndexSignal.Set(index)
	} else {
		w.cfg.selectedIndex = index
	}

	w.cache.invalidate()

	if w.cfg.onSelectionChange != nil {
		w.cfg.onSelectionChange(index)
	}

	// ADR-028: visual only — selection highlight moved.
	w.SetNeedsRedraw(true)
	ctx.InvalidateRect(w.Bounds())
}

// Default viewport dimensions used as fallback.
const (
	defaultViewportWidth  float32 = 500
	defaultViewportHeight float32 = 400
)

// ceilDiv returns the ceiling of a / b for positive integers.
func ceilDiv(a, b int) int {
	if b <= 0 {
		return 0
	}
	return int(math.Ceil(float64(a) / float64(b)))
}

// --- Event Handling ---

// handleKeyEvent processes keyboard events for grid-level navigation.
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

	itemCount := w.cfg.ResolvedItemCount()
	if itemCount == 0 {
		return false
	}

	selectedIdx := w.cfg.ResolvedSelectedIndex()
	cols := w.effectiveColumns()

	switch e.Key {
	case event.KeyRight:
		return w.moveSelection(ctx, selectedIdx+1, itemCount)
	case event.KeyLeft:
		return w.moveSelection(ctx, selectedIdx-1, itemCount)
	case event.KeyDown:
		return w.moveSelection(ctx, selectedIdx+cols, itemCount)
	case event.KeyUp:
		return w.moveSelection(ctx, selectedIdx-cols, itemCount)
	case event.KeyHome:
		return w.moveSelection(ctx, 0, itemCount)
	case event.KeyEnd:
		return w.moveSelection(ctx, itemCount-1, itemCount)
	case event.KeyPageDown:
		return w.moveSelectionByPage(ctx, selectedIdx, itemCount, 1)
	case event.KeyPageUp:
		return w.moveSelectionByPage(ctx, selectedIdx, itemCount, -1)
	case event.KeyEnter, event.KeySpace:
		if selectedIdx >= 0 && selectedIdx < itemCount && w.cfg.onCellClick != nil {
			w.cfg.onCellClick(selectedIdx)
			return true
		}
		return false
	default:
		return false
	}
}

// moveSelection attempts to move selection to newIndex, clamping to [0, count).
func (w *Widget) moveSelection(ctx widget.Context, newIndex, count int) bool {
	if w.cfg.selectionMode == SelectionNone {
		return false
	}
	if newIndex < 0 {
		newIndex = 0
	}
	if newIndex >= count {
		newIndex = count - 1
	}
	w.setSelectedIndex(ctx, newIndex)
	w.ScrollToIndex(newIndex)
	return true
}

// moveSelectionByPage moves selection by approximately one viewport worth of rows.
func (w *Widget) moveSelectionByPage(ctx widget.Context, currentIdx, count, direction int) bool {
	if w.cfg.selectionMode == SelectionNone {
		return false
	}

	cellStep := w.cfg.itemHeight + w.cfg.gap
	if cellStep <= 0 {
		cellStep = defaultItemHeight
	}
	rowsPerPage := int(w.viewportHeight / cellStep)
	if rowsPerPage < 1 {
		rowsPerPage = 1
	}
	cols := w.effectiveColumns()

	newIndex := currentIdx + direction*rowsPerPage*cols
	return w.moveSelection(ctx, newIndex, count)
}

// handleContentEvent processes input events on the virtual content area.
func handleContentEvent(gv *Widget, ctx widget.Context, e event.Event) bool {
	me, ok := e.(*event.MouseEvent)
	if !ok {
		return false
	}
	return handleContentMouseEvent(gv, ctx, me)
}

// handleContentMouseEvent processes mouse events for cell interaction.
func handleContentMouseEvent(gv *Widget, ctx widget.Context, e *event.MouseEvent) bool {
	if gv.cfg.ResolvedDisabled() {
		return false
	}

	switch e.MouseType {
	case event.MouseMove:
		return handleContentMouseMove(gv, ctx, e)
	case event.MousePress:
		return handleContentMousePress(gv, ctx, e)
	case event.MouseLeave:
		if gv.hoveredIndex != noHoveredIndex {
			gv.hoveredIndex = noHoveredIndex
			gv.cache.invalidate()
			// ADR-028: visual only — cell hover cleared.
			gv.SetNeedsRedraw(true)
			ctx.InvalidateRect(gv.Bounds())
		}
		return false
	default:
		return false
	}
}

// handleContentMouseMove updates the hovered cell index based on mouse position.
func handleContentMouseMove(gv *Widget, ctx widget.Context, e *event.MouseEvent) bool {
	scrollY := gv.currentScrollY()
	contentX := e.Position.X - gv.Bounds().Min.X
	contentY := e.Position.Y - gv.Bounds().Min.Y + scrollY

	idx := gv.cellIndexAtPoint(contentX, contentY)

	if idx != gv.hoveredIndex {
		gv.hoveredIndex = idx
		gv.cache.invalidate()
		// ADR-028: visual only — cell hover changed.
		gv.SetNeedsRedraw(true)
		ctx.InvalidateRect(gv.Bounds())
	}
	return false // Don't consume move events.
}

// handleContentMousePress handles cell click for selection.
func handleContentMousePress(gv *Widget, ctx widget.Context, e *event.MouseEvent) bool {
	if e.Button != event.ButtonLeft {
		return false
	}

	scrollY := gv.currentScrollY()
	contentX := e.Position.X - gv.Bounds().Min.X
	contentY := e.Position.Y - gv.Bounds().Min.Y + scrollY

	idx := gv.cellIndexAtPoint(contentX, contentY)
	if idx < 0 {
		return false
	}

	if gv.cfg.onCellClick != nil {
		gv.cfg.onCellClick(idx)
	}

	if gv.cfg.selectionMode == SelectionSingle {
		gv.setSelectedIndex(ctx, idx)
	}

	ctx.RequestFocus(gv)
	return true
}

// --- Virtual Content ---

// virtualContent is an internal widget that represents the entire scrollable
// content area. It reports the total content height to the parent ScrollView
// but only renders visible cells.
type virtualContent struct {
	widget.WidgetBase
	grid *Widget // back-reference to parent GridView
}

// Layout returns the total content size.
func (vc *virtualContent) Layout(_ widget.Context, c geometry.Constraints) geometry.Size {
	if vc.grid == nil {
		return geometry.Size{}
	}

	totalHeight := vc.grid.totalContentHeight()

	width := c.MaxWidth
	if width >= geometry.Infinity {
		width = c.MinWidth
	}
	if width < c.MinWidth {
		width = c.MinWidth
	}

	return geometry.Sz(width, totalHeight)
}

// Draw renders only the visible cells within the current viewport.
func (vc *virtualContent) Draw(ctx widget.Context, canvas widget.Canvas) {
	if vc.grid == nil {
		return
	}
	vc.grid.drawVisibleCells(ctx, canvas)
}

// Event delegates events back to the parent grid for cell interaction.
func (vc *virtualContent) Event(ctx widget.Context, e event.Event) bool {
	if vc.grid == nil {
		return false
	}
	return handleContentEvent(vc.grid, ctx, e)
}

// Children returns nil; visible cell widgets are ephemeral and managed by the cache.
func (vc *virtualContent) Children() []widget.Widget {
	return nil
}

// drawVisibleCells renders only the cells currently visible in the viewport.
func (w *Widget) drawVisibleCells(ctx widget.Context, canvas widget.Canvas) {
	itemCount := w.cfg.ResolvedItemCount()

	if itemCount == 0 {
		w.painter.PaintEmptyState(canvas, geometry.NewRect(0, 0, w.viewportWidth, w.viewportHeight))
		return
	}

	scrollY := w.currentScrollY()
	cols := w.effectiveColumns()
	if cols <= 0 {
		return
	}

	firstRow, lastRow := w.visibleRowRange(scrollY, w.viewportHeight)
	selectedIdx := w.cfg.ResolvedSelectedIndex()

	// Compute visible cell range.
	startIdx := firstRow * cols
	endIdx := (lastRow + 1) * cols
	if endIdx > itemCount {
		endIdx = itemCount
	}
	if startIdx >= endIdx {
		return
	}

	// Update cell widget cache.
	w.cache.update(startIdx, endIdx, w.cfg.cellContent, selectedIdx, w.hoveredIndex, cols, ctx)

	contentWidth := w.viewportWidth - w.scroll.ScrollbarInset()
	_ = contentWidth // Used for future cell width scaling.

	// Draw each visible cell.
	for idx := startIdx; idx < endIdx; idx++ {
		cellW := w.cache.widgetAt(idx - startIdx)

		row := idx / cols
		col := idx % cols

		x := float32(col) * (w.cfg.itemWidth + w.cfg.gap)
		y := float32(row)*(w.cfg.itemHeight+w.cfg.gap) - scrollY

		cellBounds := geometry.NewRect(x, y, w.cfg.itemWidth, w.cfg.itemHeight)

		// Paint cell background (hover).
		cps := CellPaintState{
			Bounds:   cellBounds,
			Index:    idx,
			Row:      row,
			Col:      col,
			Selected: idx == selectedIdx,
			Focused:  idx == selectedIdx && w.IsFocused(),
			Hovered:  idx == w.hoveredIndex,
			Disabled: w.cfg.ResolvedDisabled(),
		}

		w.painter.PaintCellBackground(canvas, cps)

		if cps.Selected {
			w.painter.PaintSelection(canvas, cps)
		}

		// Draw cell widget if available.
		if cellW != nil {
			cellConstraints := geometry.Constraints{
				MinWidth:  w.cfg.itemWidth,
				MaxWidth:  w.cfg.itemWidth,
				MinHeight: w.cfg.itemHeight,
				MaxHeight: w.cfg.itemHeight,
			}
			cellW.Layout(ctx, cellConstraints)

			if setter, ok := cellW.(interface{ SetBounds(geometry.Rect) }); ok {
				setter.SetBounds(cellBounds)
			}

			cellW.Draw(ctx, canvas)
		}
	}
}

// --- Cell Cache ---

// cellCache caches the currently visible cell widgets between frames.
type cellCache struct {
	startIndex int
	endIndex   int
	widgets    []widget.Widget
	valid      bool
}

// update ensures the cache contains widgets for the range [start, end).
func (cc *cellCache) update(start, end int, content cdk.Content[CellContext], selectedIndex, hoveredIndex, cols int, ctx widget.Context) {
	count := end - start
	if count <= 0 {
		cc.clear()
		return
	}

	if cc.valid && cc.startIndex == start && cc.endIndex == end {
		return
	}

	// Unmount old cell widgets before replacing the slice (#181).
	for _, w := range cc.widgets {
		if w != nil {
			widget.UnmountTree(w)
		}
	}

	if cap(cc.widgets) >= count {
		cc.widgets = cc.widgets[:count]
	} else {
		cc.widgets = make([]widget.Widget, count)
	}

	if content == nil {
		for i := range cc.widgets {
			cc.widgets[i] = nil
		}
		cc.startIndex = start
		cc.endIndex = end
		cc.valid = true
		return
	}

	safeCols := cols
	if safeCols <= 0 {
		safeCols = 1
	}
	for i := range count {
		idx := start + i
		row := idx / safeCols
		col := idx % safeCols
		w := content.Render(CellContext{
			Index:      idx,
			Row:        row,
			Col:        col,
			IsSelected: idx == selectedIndex,
			IsHovered:  idx == hoveredIndex,
		})
		// Mount cell widget so signal bindings activate (#181).
		if w != nil && ctx != nil {
			widget.MountTree(w, ctx)
		}
		cc.widgets[i] = w
	}

	cc.startIndex = start
	cc.endIndex = end
	cc.valid = true
}

// widgetAt returns the cached widget at the given offset from startIndex.
func (cc *cellCache) widgetAt(offset int) widget.Widget {
	if offset < 0 || offset >= len(cc.widgets) {
		return nil
	}
	return cc.widgets[offset]
}

// invalidate marks the cache as needing a rebuild.
func (cc *cellCache) invalidate() {
	cc.valid = false
}

// clear resets the cache entirely and unmounts all cell widgets.
func (cc *cellCache) clear() {
	for i, w := range cc.widgets {
		if w != nil {
			widget.UnmountTree(w)
		}
		cc.widgets[i] = nil
	}
	cc.widgets = cc.widgets[:0]
	cc.startIndex = 0
	cc.endIndex = 0
	cc.valid = false
}

// Verify Widget implements required interfaces at compile time.
var (
	_ widget.Widget    = (*Widget)(nil)
	_ widget.Focusable = (*Widget)(nil)
	_ widget.Lifecycle = (*Widget)(nil)
	_ a11y.Accessible  = (*Widget)(nil)
)
