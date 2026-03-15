package treeview

import (
	"fmt"

	"github.com/gogpu/ui/a11y"
	"github.com/gogpu/ui/event"
	"github.com/gogpu/ui/geometry"
	"github.com/gogpu/ui/state"
	"github.com/gogpu/ui/widget"
)

// Widget implements a hierarchical tree view with expand/collapse per node,
// keyboard navigation, virtualization, and pluggable painting.
//
// Create with [New] using functional options:
//
//	tree := treeview.New(
//	    treeview.Root(root),
//	    treeview.ItemHeight(28),
//	    treeview.IndentWidth(20),
//	    treeview.SelectionModeOpt(treeview.SelectionSingle),
//	    treeview.OnSelect(func(node *treeview.TreeNode) { ... }),
//	)
type Widget struct {
	widget.WidgetBase
	cfg     config
	painter Painter

	// Flattened visible rows (rebuilt on expand/collapse).
	rows []flatRow

	// Viewport state for virtualization.
	scrollY        float32
	viewportWidth  float32
	viewportHeight float32

	// Interaction state.
	hoveredIndex int
}

// New creates a new tree view Widget with the given options.
//
// The returned widget is visible, enabled, and focusable by default.
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

	// Build initial flattened rows.
	w.rebuildRows()

	return w
}

// IsFocusable reports whether the tree view can currently receive focus.
func (w *Widget) IsFocusable() bool {
	return w.IsVisible() && w.IsEnabled() && !w.cfg.ResolvedDisabled()
}

// Layout calculates the tree view's size within the given constraints.
func (w *Widget) Layout(_ widget.Context, constraints geometry.Constraints) geometry.Size {
	// Rebuild rows in case root changed via signal.
	w.rebuildRows()

	size := constraints.BiggestFinite(defaultViewportWidth, defaultViewportHeight)

	// For unconstrained height, use content height clamped to default.
	if constraints.HasInfiniteHeight() {
		totalH := float32(len(w.rows)) * w.cfg.itemHeight
		if totalH > defaultViewportHeight {
			totalH = defaultViewportHeight
		}
		if totalH <= 0 {
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

	return size
}

// Draw renders the tree view to the canvas.
func (w *Widget) Draw(ctx widget.Context, canvas widget.Canvas) {
	if !w.IsVisible() {
		return
	}

	bounds := w.Bounds()
	if bounds.IsEmpty() {
		return
	}

	// Empty state.
	if len(w.rows) == 0 {
		w.painter.PaintEmptyState(canvas, bounds)
		return
	}

	// Clip to bounds.
	canvas.PushClip(bounds)

	// Determine visible range.
	startIdx, endIdx := w.visibleRange()

	selectedID := w.cfg.ResolvedSelectedNodeID()
	isFocused := w.IsFocused()
	isDisabled := w.cfg.ResolvedDisabled()

	for i := startIdx; i < endIdx; i++ {
		row := w.rows[i]
		rowBounds := w.rowBounds(i, bounds)

		isSelected := selectedID != "" && row.node.ID == selectedID
		isHovered := i == w.hoveredIndex

		rowState := RowPaintState{
			Bounds:   rowBounds,
			Node:     row.node,
			Depth:    row.depth,
			Selected: isSelected,
			Focused:  isFocused && isSelected,
			Hovered:  isHovered,
			Disabled: isDisabled,
		}

		// Paint backgrounds.
		w.painter.PaintRowBackground(canvas, rowState)
		w.painter.PaintSelection(canvas, rowState)

		// Paint connector lines if enabled.
		if w.cfg.showLines && row.depth > 0 {
			connState := w.buildConnectorState(i, rowBounds)
			w.painter.PaintConnectorLines(canvas, connState)
		}

		// Paint expand icon for branch nodes.
		if !row.node.IsLeaf() {
			iconBounds := w.expandIconBounds(row.depth, rowBounds)
			w.painter.PaintExpandIcon(canvas, ExpandIconState{
				Bounds:   iconBounds,
				Expanded: row.node.Expanded,
				Hovered:  isHovered,
			})
		}

		// Paint label.
		labelBounds := w.labelBounds(row.depth, rowBounds)
		w.painter.PaintLabel(canvas, LabelState{
			Bounds:   labelBounds,
			Text:     row.node.Label,
			Selected: isSelected,
			Disabled: isDisabled,
		})
	}

	canvas.PopClip()
}

// Event handles an input event and returns true if consumed.
func (w *Widget) Event(ctx widget.Context, e event.Event) bool {
	if !w.IsVisible() || !w.IsEnabled() {
		return false
	}
	return handleEvent(w, ctx, e)
}

// Children returns nil (tree view is a leaf widget with internal rendering).
func (w *Widget) Children() []widget.Widget {
	return nil
}

// Mount creates signal bindings for push-based invalidation.
// Implements [widget.Lifecycle].
func (w *Widget) Mount(ctx widget.Context) {
	sched := ctx.Scheduler()
	if sched == nil {
		return
	}

	// Bind root signals.
	if w.cfg.readonlyRootSignal != nil {
		b := state.BindToScheduler(w.cfg.readonlyRootSignal, w, sched)
		w.AddBinding(b)
	} else if w.cfg.rootSignal != nil {
		b := state.BindToScheduler(w.cfg.rootSignal, w, sched)
		w.AddBinding(b)
	}

	// Bind selected node signals.
	if w.cfg.readonlySelectedNodeSignal != nil {
		b := state.BindToScheduler(w.cfg.readonlySelectedNodeSignal, w, sched)
		w.AddBinding(b)
	} else if w.cfg.selectedNodeIDSignal != nil {
		b := state.BindToScheduler(w.cfg.selectedNodeIDSignal, w, sched)
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
}

// Unmount is called when the tree view is removed from the widget tree.
// Implements [widget.Lifecycle].
func (w *Widget) Unmount() {
	// Bindings are cleaned up automatically by WidgetBase.CleanupBindings().
}

// --- Public API ---

// ScrollToNode scrolls to make the node with the given ID visible.
// If the node is not in the flattened visible list, this is a no-op.
func (w *Widget) ScrollToNode(id string) {
	idx := w.findRowIndex(id)
	if idx < 0 {
		return
	}
	w.scrollToIndex(idx)
}

// VisibleRange returns the indices [start, end) of currently visible rows.
func (w *Widget) VisibleRange() (start, end int) {
	return w.visibleRange()
}

// RowCount returns the number of currently visible (flattened) rows.
func (w *Widget) RowCount() int {
	return len(w.rows)
}

// InvalidateData signals that the tree data has changed.
// Rebuilds the flattened row list.
func (w *Widget) InvalidateData() {
	w.rebuildRows()
	w.SetNeedsRedraw(true)
}

// ExpandAll expands all branch nodes in the tree.
func (w *Widget) ExpandAll() {
	root := w.cfg.ResolvedRoot()
	if root == nil {
		return
	}
	setExpandedAll(root, true)
	w.rebuildRows()
	w.SetNeedsRedraw(true)
}

// CollapseAll collapses all branch nodes in the tree.
func (w *Widget) CollapseAll() {
	root := w.cfg.ResolvedRoot()
	if root == nil {
		return
	}
	setExpandedAll(root, false)
	w.rebuildRows()
	w.SetNeedsRedraw(true)
}

// --- Accessibility ---

// AccessibilityRole returns the ARIA role for this widget.
func (w *Widget) AccessibilityRole() a11y.Role {
	return a11y.RoleTree
}

// AccessibilityLabel returns the accessibility label.
func (w *Widget) AccessibilityLabel() string {
	if w.cfg.a11yLabel != "" {
		return w.cfg.a11yLabel
	}
	return "Tree"
}

// AccessibilityHint returns the accessibility hint.
func (w *Widget) AccessibilityHint() string {
	return ""
}

// AccessibilityValue returns the current tree state as a string.
func (w *Widget) AccessibilityValue() string {
	selectedID := w.cfg.ResolvedSelectedNodeID()
	root := w.cfg.ResolvedRoot()
	if selectedID != "" && root != nil {
		if node := findNodeByID(root, selectedID); node != nil {
			return fmt.Sprintf("Selected: %s, %d visible rows", node.Label, len(w.rows))
		}
	}
	return fmt.Sprintf("%d visible rows", len(w.rows))
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

// rebuildRows flattens the tree into visible rows.
func (w *Widget) rebuildRows() {
	root := w.cfg.ResolvedRoot()
	w.rows = flattenTree(root)
}

// visibleRange returns the [start, end) indices of rows visible in the viewport.
func (w *Widget) visibleRange() (int, int) {
	if len(w.rows) == 0 || w.cfg.itemHeight <= 0 {
		return 0, 0
	}

	startIdx := int(w.scrollY / w.cfg.itemHeight)
	if startIdx < 0 {
		startIdx = 0
	}

	visibleCount := int(w.viewportHeight/w.cfg.itemHeight) + 2 // +2 for partial rows
	endIdx := startIdx + visibleCount
	if endIdx > len(w.rows) {
		endIdx = len(w.rows)
	}

	return startIdx, endIdx
}

// rowBounds returns the bounding rectangle for the row at index i.
func (w *Widget) rowBounds(i int, treeBounds geometry.Rect) geometry.Rect {
	y := treeBounds.Min.Y + float32(i)*w.cfg.itemHeight - w.scrollY
	return geometry.NewRect(
		treeBounds.Min.X, y,
		w.viewportWidth, w.cfg.itemHeight,
	)
}

// expandIconBounds returns the bounding rectangle for the expand icon.
func (w *Widget) expandIconBounds(depth int, rowBounds geometry.Rect) geometry.Rect {
	x := rowBounds.Min.X + float32(depth)*w.cfg.indentWidth
	iconY := rowBounds.Min.Y + (w.cfg.itemHeight-expandIconSize)/2
	return geometry.NewRect(x, iconY, expandIconSize, expandIconSize)
}

// labelBounds returns the bounding rectangle for the label text.
func (w *Widget) labelBounds(depth int, rowBounds geometry.Rect) geometry.Rect {
	x := rowBounds.Min.X + float32(depth+1)*w.cfg.indentWidth
	return geometry.NewRect(x, rowBounds.Min.Y, rowBounds.Max.X-x, w.cfg.itemHeight)
}

// findRowIndex returns the index of the row with the given node ID, or -1.
func (w *Widget) findRowIndex(id string) int {
	for i, row := range w.rows {
		if row.node.ID == id {
			return i
		}
	}
	return -1
}

// scrollToIndex scrolls to make the row at the given index visible.
func (w *Widget) scrollToIndex(index int) {
	if index < 0 || index >= len(w.rows) {
		return
	}

	itemTop := float32(index) * w.cfg.itemHeight
	itemBottom := itemTop + w.cfg.itemHeight

	// Already visible: no-op.
	if itemTop >= w.scrollY && itemBottom <= w.scrollY+w.viewportHeight {
		return
	}

	if itemTop < w.scrollY {
		w.scrollY = itemTop
	} else {
		w.scrollY = itemBottom - w.viewportHeight
		if w.scrollY < 0 {
			w.scrollY = 0
		}
	}
}

// setSelectedNodeID updates the selected node, writing back to signal if bound.
func (w *Widget) setSelectedNodeID(ctx widget.Context, id string) {
	current := w.cfg.ResolvedSelectedNodeID()
	if id == current {
		return
	}

	// TWO-WAY: write back to signal if bound.
	if w.cfg.selectedNodeIDSignal != nil {
		w.cfg.selectedNodeIDSignal.Set(id)
	} else {
		w.cfg.selectedNodeID = id
	}

	w.SetNeedsRedraw(true)

	if w.cfg.onSelect != nil {
		root := w.cfg.ResolvedRoot()
		if root != nil {
			if node := findNodeByID(root, id); node != nil {
				w.cfg.onSelect(node)
			}
		}
	}

	ctx.Invalidate()
}

// toggleNode toggles the expanded state of the given node.
func (w *Widget) toggleNode(ctx widget.Context, node *TreeNode) {
	if node.IsLeaf() {
		return
	}

	node.Expanded = !node.Expanded
	w.rebuildRows()
	w.SetNeedsRedraw(true)

	if w.cfg.onToggle != nil {
		w.cfg.onToggle(node, node.Expanded)
	}

	ctx.Invalidate()
}

// buildConnectorState builds the connector state for the row at index i.
func (w *Widget) buildConnectorState(i int, rowBounds geometry.Rect) ConnectorState {
	row := w.rows[i]
	root := w.cfg.ResolvedRoot()

	// Determine if this node is the last child of its parent.
	isLast := false
	if root != nil {
		parent := findParent(root, row.node.ID)
		if parent != nil {
			children := parent.Children
			isLast = len(children) > 0 && children[len(children)-1].ID == row.node.ID
		}
	}

	// Build ParentHasMore by checking each ancestor level.
	parentHasMore := make([]bool, row.depth)
	if root != nil {
		current := row.node
		for d := row.depth - 1; d >= 0; d-- {
			p := findParent(root, current.ID)
			if p != nil {
				pChildren := p.Children
				parentHasMore[d] = len(pChildren) > 0 && pChildren[len(pChildren)-1].ID != current.ID
				current = p
			}
		}
	}

	return ConnectorState{
		RowBounds:     rowBounds,
		Depth:         row.depth,
		IndentWidth:   w.cfg.indentWidth,
		IsLastChild:   isLast,
		HasChildren:   !row.node.IsLeaf(),
		ParentHasMore: parentHasMore,
	}
}

// setExpandedAll recursively sets Expanded on all branch nodes.
func setExpandedAll(node *TreeNode, expanded bool) {
	if len(node.Children) > 0 {
		node.Expanded = expanded
		for _, child := range node.Children {
			setExpandedAll(child, expanded)
		}
	}
}

// Default viewport dimensions used as fallback.
const (
	defaultViewportWidth  float32 = 200
	defaultViewportHeight float32 = 400
)

// noHoveredIndex indicates no row is currently hovered.
const noHoveredIndex = -1

// Verify Widget implements required interfaces at compile time.
var (
	_ widget.Widget    = (*Widget)(nil)
	_ widget.Focusable = (*Widget)(nil)
	_ widget.Lifecycle = (*Widget)(nil)
	_ a11y.Accessible  = (*Widget)(nil)
)
