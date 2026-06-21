package listview

import (
	"github.com/gogpu/ui/cdk"
	"github.com/gogpu/ui/widget"
)

// widgetCache caches the currently visible item decorators between frames.
//
// When the visible range has not changed and data has not been invalidated,
// the cache returns the same decorators without calling the builder again.
// This avoids unnecessary allocations during static frames (no scroll).
//
// Each item widget is wrapped in an [itemDecorator] that IS the RepaintBoundary.
// The decorator owns hover/selection painting — when hover changes, only the
// decorator's scene is re-recorded, not the entire ListView.
type widgetCache struct {
	startIndex    int
	endIndex      int
	selectedIndex int
	widgets       []widget.Widget
	valid         bool
	list          *Widget // back-reference for decorator creation
}

// rebuildAffected rebuilds only items whose selection state changed.
// Android RecyclerView pattern: only affected ViewHolders are rebound.
//
// Hover changes do NOT trigger rebuilds — the decorator reads hover state
// from list.hoveredIndex at Draw time (visual-only, no content change).
func (wc *widgetCache) rebuildAffected(start int, content cdk.Content[ItemContext], selectedIndex int) {
	affectedIndices := make(map[int]bool)
	if wc.selectedIndex != selectedIndex {
		affectedIndices[wc.selectedIndex] = true
		affectedIndices[selectedIndex] = true
	}

	for idx := range affectedIndices {
		offset := idx - start
		if offset < 0 || offset >= len(wc.widgets) {
			continue
		}
		w := content.Render(ItemContext{
			Index:    idx,
			Selected: idx == selectedIndex,
			Focused:  idx == selectedIndex,
		})
		if w != nil {
			wc.widgets[offset] = newItemDecorator(w, wc.list, idx)
		} else {
			wc.widgets[offset] = nil
		}
	}
}

// fullRebuild recreates all items in the range (scroll or first build).
func (wc *widgetCache) fullRebuild(start, _, count int, content cdk.Content[ItemContext], selectedIndex int) {
	if cap(wc.widgets) >= count {
		wc.widgets = wc.widgets[:count]
	} else {
		wc.widgets = make([]widget.Widget, count)
	}

	if content == nil {
		for i := range wc.widgets {
			wc.widgets[i] = nil
		}
	} else {
		for i := range count {
			idx := start + i
			w := content.Render(ItemContext{
				Index:    idx,
				Selected: idx == selectedIndex,
				Focused:  idx == selectedIndex,
			})
			if w != nil {
				wc.widgets[i] = newItemDecorator(w, wc.list, idx)
			} else {
				wc.widgets[i] = nil
			}
		}
	}
}

// update ensures the cache contains decorators for the range [start, end).
// If the range matches and the cache is valid, this is a no-op.
// Otherwise, it calls the content's Render method for each index in the range
// and wraps each widget in an itemDecorator (which is the RepaintBoundary).
//
// Hover state is NOT tracked here — decorators read it at Draw time from
// list.hoveredIndex, so hover changes require no cache action.
func (wc *widgetCache) update(start, end int, content cdk.Content[ItemContext], selectedIndex int) {
	count := end - start
	if count <= 0 {
		wc.clear()
		return
	}

	// Fast path: same range, only selection changed -> rebuild only affected items.
	// Android RecyclerView pattern: notifyItemChanged(pos) rebinds single ViewHolder.
	if wc.valid && wc.startIndex == start && wc.endIndex == end && content != nil {
		if wc.selectedIndex != selectedIndex {
			wc.rebuildAffected(start, content, selectedIndex)
			wc.selectedIndex = selectedIndex
			return
		}
		return // nothing changed
	}

	// Incremental update: reuse decorators for overlapping indices,
	// create new ones only at viewport edges. RecyclerView/Flutter pattern.
	if wc.valid && content != nil {
		wc.incrementalUpdate(start, end, count, content, selectedIndex)
	} else {
		wc.fullRebuild(start, end, count, content, selectedIndex)
	}
	wc.startIndex = start
	wc.endIndex = end
	wc.selectedIndex = selectedIndex
	wc.valid = true
}

// incrementalUpdate reuses decorators for indices that remain visible and
// creates new ones only at the edges. RecyclerView/Flutter pattern: items in
// the middle of the viewport are never destroyed during scroll.
func (wc *widgetCache) incrementalUpdate(start, end, count int, content cdk.Content[ItemContext], selectedIndex int) {
	overlapStart := max(start, wc.startIndex)
	overlapEnd := min(end, wc.endIndex)

	if overlapStart >= overlapEnd {
		wc.fullRebuild(start, end, count, content, selectedIndex)
		return
	}

	newWidgets := make([]widget.Widget, count)

	for i := overlapStart; i < overlapEnd; i++ {
		newWidgets[i-start] = wc.widgets[i-wc.startIndex]
	}

	for i := start; i < overlapStart; i++ {
		newWidgets[i-start] = wc.buildDecorator(i, content, selectedIndex)
	}

	for i := overlapEnd; i < end; i++ {
		newWidgets[i-start] = wc.buildDecorator(i, content, selectedIndex)
	}

	wc.widgets = newWidgets
}

func (wc *widgetCache) buildDecorator(index int, content cdk.Content[ItemContext], selectedIndex int) widget.Widget {
	w := content.Render(ItemContext{
		Index:    index,
		Selected: index == selectedIndex,
		Focused:  index == selectedIndex,
	})
	if w == nil {
		return nil
	}
	return newItemDecorator(w, wc.list, index)
}

// widgetAt returns the cached decorator at the given offset from startIndex.
func (wc *widgetCache) widgetAt(offset int) widget.Widget {
	if offset < 0 || offset >= len(wc.widgets) {
		return nil
	}
	return wc.widgets[offset]
}

// childAt returns the inner user widget (unwrapped from the decorator) at the
// given offset. Returns nil if the offset is out of range or the widget is nil.
func (wc *widgetCache) childAt(offset int) widget.Widget {
	w := wc.widgetAt(offset)
	if w == nil {
		return nil
	}
	if dec, ok := w.(*itemDecorator); ok {
		return dec.child
	}
	return w
}

// invalidate marks the cache as needing a rebuild.
func (wc *widgetCache) invalidate() {
	wc.valid = false
}

// clear resets the cache entirely and unmounts boundaries to free pixel caches.
func (wc *widgetCache) clear() {
	for i := range wc.widgets {
		wc.widgets[i] = nil
	}
	wc.widgets = wc.widgets[:0]
	wc.startIndex = 0
	wc.endIndex = 0
	wc.valid = false
}
