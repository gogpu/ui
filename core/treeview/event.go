package treeview

import (
	"github.com/gogpu/ui/event"
	"github.com/gogpu/ui/geometry"
	"github.com/gogpu/ui/widget"
)

// handleEvent dispatches input events to the appropriate handler.
func handleEvent(w *Widget, ctx widget.Context, e event.Event) bool {
	switch ev := e.(type) {
	case *event.MouseEvent:
		return handleMouseEvent(w, ctx, ev)
	case *event.KeyEvent:
		return handleKeyEvent(w, ctx, ev)
	case *event.WheelEvent:
		return handleWheelEvent(w, ctx, ev)
	default:
		return false
	}
}

// handleMouseEvent processes mouse events for row interaction.
func handleMouseEvent(w *Widget, ctx widget.Context, e *event.MouseEvent) bool {
	if w.cfg.ResolvedDisabled() {
		return false
	}

	bounds := w.Bounds()
	if !bounds.Contains(e.Position) && e.MouseType != event.MouseLeave {
		return false
	}

	switch e.MouseType {
	case event.MouseMove:
		return handleMouseMove(w, ctx, e)
	case event.MousePress:
		return handleMousePress(w, ctx, e)
	case event.MouseLeave:
		if w.hoveredIndex != noHoveredIndex {
			w.hoveredIndex = noHoveredIndex
			w.SetNeedsRedraw(true)
			ctx.Invalidate()
		}
		return false
	default:
		return false
	}
}

// handleMouseMove updates the hovered row index based on mouse position.
func handleMouseMove(w *Widget, ctx widget.Context, e *event.MouseEvent) bool {
	idx := w.hitTestRow(e.Position)

	if idx != w.hoveredIndex {
		w.hoveredIndex = idx
		w.SetNeedsRedraw(true)
		ctx.Invalidate()
	}
	return false // Don't consume move events.
}

// handleMousePress handles row click for selection and expand/collapse.
func handleMousePress(w *Widget, ctx widget.Context, e *event.MouseEvent) bool {
	if e.Button != event.ButtonLeft {
		return false
	}

	idx := w.hitTestRow(e.Position)
	if idx < 0 || idx >= len(w.rows) {
		return false
	}

	row := w.rows[idx]
	bounds := w.Bounds()
	rowBounds := w.rowBounds(idx, bounds)

	// Check if click is on the expand icon area.
	if !row.node.IsLeaf() {
		iconBounds := w.expandIconBounds(row.depth, rowBounds)
		if iconBounds.Contains(e.Position) {
			w.toggleNode(ctx, row.node)
			return true
		}
	}

	// Click on the row -- select the node.
	if w.cfg.selectionMode == SelectionSingle {
		w.setSelectedNodeID(ctx, row.node.ID)
	}

	// Request focus for keyboard navigation.
	ctx.RequestFocus(w)

	return true
}

// handleKeyEvent processes keyboard events for tree navigation.
func handleKeyEvent(w *Widget, ctx widget.Context, e *event.KeyEvent) bool {
	if !w.IsFocused() {
		return false
	}

	if e.KeyType != event.KeyPress && e.KeyType != event.KeyRepeat {
		return false
	}

	if w.cfg.ResolvedDisabled() {
		return false
	}

	if len(w.rows) == 0 {
		return false
	}

	selectedID := w.cfg.ResolvedSelectedNodeID()
	selectedIdx := w.findRowIndex(selectedID)

	switch e.Key {
	case event.KeyDown:
		return w.moveSelection(ctx, selectedIdx+1)
	case event.KeyUp:
		return w.moveSelection(ctx, selectedIdx-1)
	case event.KeyHome:
		return w.moveSelection(ctx, 0)
	case event.KeyEnd:
		return w.moveSelection(ctx, len(w.rows)-1)
	case event.KeyRight:
		return w.handleKeyRight(ctx, selectedIdx)
	case event.KeyLeft:
		return w.handleKeyLeft(ctx, selectedIdx)
	case event.KeyEnter, event.KeySpace:
		return w.handleKeyActivate(selectedIdx)
	default:
		return false
	}
}

// handleKeyRight expands the current node or moves to its first child.
func (w *Widget) handleKeyRight(ctx widget.Context, idx int) bool {
	if idx < 0 || idx >= len(w.rows) {
		return false
	}
	row := w.rows[idx]

	if row.node.IsLeaf() {
		return false
	}

	if !row.node.Expanded {
		// Expand the node.
		w.toggleNode(ctx, row.node)
		return true
	}

	// Already expanded -- move to first child.
	if idx+1 < len(w.rows) && w.rows[idx+1].depth > row.depth {
		return w.moveSelection(ctx, idx+1)
	}
	return false
}

// handleKeyLeft collapses the current node or moves to parent.
func (w *Widget) handleKeyLeft(ctx widget.Context, idx int) bool {
	if idx < 0 || idx >= len(w.rows) {
		return false
	}
	row := w.rows[idx]

	// If expanded branch, collapse it.
	if !row.node.IsLeaf() && row.node.Expanded {
		w.toggleNode(ctx, row.node)
		return true
	}

	// Move to parent.
	root := w.cfg.ResolvedRoot()
	if root == nil {
		return false
	}
	parent := findParent(root, row.node.ID)
	if parent != nil {
		parentIdx := w.findRowIndex(parent.ID)
		if parentIdx >= 0 {
			return w.moveSelection(ctx, parentIdx)
		}
	}
	return false
}

// handleKeyActivate fires the OnSelect callback for the current selection.
func (w *Widget) handleKeyActivate(idx int) bool {
	if idx < 0 || idx >= len(w.rows) {
		return false
	}
	if w.cfg.onSelect != nil {
		w.cfg.onSelect(w.rows[idx].node)
		return true
	}
	return false
}

// moveSelection attempts to move selection to newIndex, clamping to [0, rowCount).
func (w *Widget) moveSelection(ctx widget.Context, newIndex int) bool {
	if w.cfg.selectionMode == SelectionNone {
		return false
	}
	if newIndex < 0 {
		newIndex = 0
	}
	if newIndex >= len(w.rows) {
		newIndex = len(w.rows) - 1
	}
	if newIndex < 0 {
		return false
	}

	w.setSelectedNodeID(ctx, w.rows[newIndex].node.ID)
	w.scrollToIndex(newIndex)
	return true
}

// handleWheelEvent processes wheel events for scrolling.
func handleWheelEvent(w *Widget, ctx widget.Context, e *event.WheelEvent) bool {
	bounds := w.Bounds()
	if !bounds.Contains(e.Position) {
		return false
	}

	totalHeight := float32(len(w.rows)) * w.cfg.itemHeight
	maxScroll := totalHeight - w.viewportHeight
	if maxScroll < 0 {
		maxScroll = 0
	}

	// Apply scroll delta (DeltaY() is a method on WheelEvent).
	w.scrollY -= e.DeltaY() * wheelScrollMultiplier
	if w.scrollY < 0 {
		w.scrollY = 0
	}
	if w.scrollY > maxScroll {
		w.scrollY = maxScroll
	}

	w.SetNeedsRedraw(true)
	ctx.Invalidate()
	return true
}

// hitTestRow returns the index of the row at the given point, or -1.
func (w *Widget) hitTestRow(pos geometry.Point) int {
	bounds := w.Bounds()
	if !bounds.Contains(pos) {
		return noHoveredIndex
	}

	localY := pos.Y - bounds.Min.Y + w.scrollY
	if w.cfg.itemHeight <= 0 {
		return noHoveredIndex
	}

	idx := int(localY / w.cfg.itemHeight)
	if idx < 0 || idx >= len(w.rows) {
		return noHoveredIndex
	}
	return idx
}

// wheelScrollMultiplier controls scroll speed for mouse wheel events.
const wheelScrollMultiplier float32 = 3
