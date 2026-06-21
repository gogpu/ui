package listview

import (
	"github.com/gogpu/ui/event"
	"github.com/gogpu/ui/widget"
)

// handleContentEvent processes input events on the virtual content area.
// This handles item clicks and hover tracking within the list.
func handleContentEvent(lv *Widget, ctx widget.Context, e event.Event) bool {
	switch ev := e.(type) {
	case *event.MouseEvent:
		return handleContentMouseEvent(lv, ctx, ev)
	default:
		return false
	}
}

// handleContentMouseEvent processes mouse events for item interaction.
func handleContentMouseEvent(lv *Widget, ctx widget.Context, e *event.MouseEvent) bool {
	if lv.cfg.ResolvedDisabled() {
		return false
	}

	switch e.MouseType {
	case event.MouseMove:
		return handleContentMouseMove(lv, ctx, e)
	case event.MousePress:
		return handleContentMousePress(lv, ctx, e)
	case event.MouseLeave:
		if lv.hoveredIndex != noHoveredIndex {
			old := lv.hoveredIndex
			lv.hoveredIndex = noHoveredIndex
			lv.markItemDirty(old)
		}
		return false
	default:
		return false
	}
}

// handleContentMouseMove updates the hovered item index based on mouse position.
// The event position is already in content space (transformed by ScrollView).
func handleContentMouseMove(lv *Widget, _ widget.Context, e *event.MouseEvent) bool {
	// Position is already in content space — ScrollView applies the inverse
	// of its Draw transform before dispatching to content children.
	contentY := e.Position.Y

	idx := lv.heights.findIndexAtOffset(contentY)
	itemCount := lv.cfg.ResolvedItemCount()
	if idx < 0 || idx >= itemCount {
		idx = noHoveredIndex
	}

	if idx != lv.hoveredIndex {
		old := lv.hoveredIndex
		lv.hoveredIndex = idx
		if old >= 0 {
			lv.markItemDirty(old)
		}
		if idx >= 0 {
			lv.markItemDirty(idx)
		}
	}
	return false // Don't consume move events.
}

// handleContentMousePress handles item click for selection.
// The event position is already in content space (transformed by ScrollView).
func handleContentMousePress(lv *Widget, ctx widget.Context, e *event.MouseEvent) bool {
	if e.Button != event.ButtonLeft {
		return false
	}

	// Position is already in content space — ScrollView applies the inverse
	// of its Draw transform before dispatching to content children.
	contentY := e.Position.Y

	idx := lv.heights.findIndexAtOffset(contentY)
	itemCount := lv.cfg.ResolvedItemCount()
	if idx < 0 || idx >= itemCount {
		return false
	}

	// Invoke item click callback.
	if lv.cfg.onItemClick != nil {
		lv.cfg.onItemClick(idx)
	}

	// Update selection.
	if lv.cfg.selectionMode == SelectionSingle {
		lv.setSelectedIndex(ctx, idx)
	}

	// Request focus for keyboard navigation.
	ctx.RequestFocus(lv)

	return true
}

// handleListKeyEvent processes keyboard events for list-level navigation.
func handleListKeyEvent(lv *Widget, ctx widget.Context, e *event.KeyEvent) bool {
	if !lv.IsFocused() {
		return false
	}

	if e.KeyType != event.KeyPress && e.KeyType != event.KeyRepeat {
		return false
	}

	if lv.cfg.ResolvedDisabled() {
		return false
	}

	itemCount := lv.cfg.ResolvedItemCount()
	if itemCount == 0 {
		return false
	}

	selectedIdx := lv.cfg.ResolvedSelectedIndex()

	switch e.Key {
	case event.KeyDown:
		return lv.moveSelection(ctx, selectedIdx+1, itemCount)
	case event.KeyUp:
		return lv.moveSelection(ctx, selectedIdx-1, itemCount)
	case event.KeyHome:
		return lv.moveSelection(ctx, 0, itemCount)
	case event.KeyEnd:
		return lv.moveSelection(ctx, itemCount-1, itemCount)
	case event.KeyPageDown:
		return lv.moveSelectionByPage(ctx, selectedIdx, itemCount, 1)
	case event.KeyPageUp:
		return lv.moveSelectionByPage(ctx, selectedIdx, itemCount, -1)
	case event.KeyEnter, event.KeySpace:
		if selectedIdx >= 0 && selectedIdx < itemCount && lv.cfg.onItemClick != nil {
			lv.cfg.onItemClick(selectedIdx)
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

// moveSelectionByPage moves selection by approximately one viewport worth of items.
func (w *Widget) moveSelectionByPage(ctx widget.Context, currentIdx, count, direction int) bool {
	if w.cfg.selectionMode == SelectionNone {
		return false
	}

	// Estimate items per page.
	avgHeight := w.heights.currentEstimate()
	if w.heights.mode == heightFixed {
		avgHeight = w.heights.fixedHeight
	}
	if avgHeight <= 0 {
		avgHeight = defaultEstimatedHeight
	}
	itemsPerPage := int(w.viewportHeight / avgHeight)
	if itemsPerPage < 1 {
		itemsPerPage = 1
	}

	newIndex := currentIdx + direction*itemsPerPage
	return w.moveSelection(ctx, newIndex, count)
}

// setSelectedIndex updates the selected index, writing back to signal if bound.
func (w *Widget) setSelectedIndex(_ widget.Context, index int) {
	current := w.cfg.ResolvedSelectedIndex()
	if index == current {
		return
	}

	// TWO-WAY: write back to signal if bound.
	if w.cfg.selectedIndexSignal != nil {
		w.cfg.selectedIndexSignal.Set(index)
	} else {
		w.cfg.selectedIndex = index
	}

	// Mark old and new selected items dirty (not entire ListView).
	// No cache.invalidate() — cache.update detects selectedIndex change
	// and rebuilds only when needed (not the entire visible range).
	w.markItemDirty(current)
	w.markItemDirty(index)

	if w.cfg.onSelectionChange != nil {
		w.cfg.onSelectionChange(index)
	}
}

// noHoveredIndex indicates no item is currently hovered.
const noHoveredIndex = -1
