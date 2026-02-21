package dropdown

import (
	"github.com/gogpu/ui/event"
	"github.com/gogpu/ui/geometry"
	"github.com/gogpu/ui/widget"
)

// menuWidget is an internal widget that renders the dropdown menu items
// inside the overlay. It handles keyboard navigation, mouse hover
// highlighting, mouse click selection, and mouse wheel scrolling.
type menuWidget struct {
	widget.WidgetBase

	items            []ItemDef
	selectedIndex    int
	highlightedIndex int
	scrollOffset     int
	maxVisible       int
	itemHeight       float32
	painter          Painter
	colorScheme      DropdownColorScheme

	onSelect func(index int)
}

// menuItemHeight is the default height of each item row.
const menuItemHeight float32 = 40

// newMenuWidget creates a new internal menu widget.
func newMenuWidget(
	items []ItemDef,
	selectedIndex int,
	maxVisible int,
	painter Painter,
	onSelect func(index int),
) *menuWidget {
	m := &menuWidget{
		items:            items,
		selectedIndex:    selectedIndex,
		highlightedIndex: selectedIndex,
		maxVisible:       maxVisible,
		itemHeight:       menuItemHeight,
		painter:          painter,
		onSelect:         onSelect,
	}
	m.SetVisible(true)
	m.SetEnabled(true)

	// If highlighted index is valid, ensure it is visible.
	if m.highlightedIndex >= 0 {
		m.ensureVisible(m.highlightedIndex)
	}

	return m
}

// Layout calculates the menu size based on visible items.
func (m *menuWidget) Layout(_ widget.Context, constraints geometry.Constraints) geometry.Size {
	visibleCount := m.visibleCount()
	height := float32(visibleCount) * m.itemHeight

	// Width matches constraints (typically matched to trigger width).
	width := constraints.MaxWidth
	if width >= geometry.Infinity {
		width = 200 // reasonable fallback
	}

	preferred := geometry.Sz(width, height)
	return constraints.Constrain(preferred)
}

// Draw renders the menu items.
func (m *menuWidget) Draw(ctx widget.Context, canvas widget.Canvas) {
	if canvas == nil {
		return
	}
	m.painter.PaintMenu(canvas, &MenuPaintState{
		Bounds:           m.Bounds(),
		Items:            m.items,
		HighlightedIndex: m.highlightedIndex,
		SelectedIndex:    m.selectedIndex,
		ScrollOffset:     m.scrollOffset,
		VisibleCount:     m.visibleCount(),
		ItemHeight:       m.itemHeight,
		ColorScheme:      m.colorScheme,
	})
}

// Event handles keyboard navigation, mouse hover, and mouse clicks.
func (m *menuWidget) Event(ctx widget.Context, e event.Event) bool {
	switch ev := e.(type) {
	case *event.KeyEvent:
		return m.handleKeyEvent(ctx, ev)
	case *event.MouseEvent:
		return m.handleMouseEvent(ctx, ev)
	case *event.WheelEvent:
		return m.handleWheelEvent(ctx, ev)
	default:
		return false
	}
}

// Children returns nil; the menu is a leaf widget.
func (m *menuWidget) Children() []widget.Widget {
	return nil
}

// handleKeyEvent processes keyboard navigation.
func (m *menuWidget) handleKeyEvent(ctx widget.Context, e *event.KeyEvent) bool {
	if e.KeyType != event.KeyPress && e.KeyType != event.KeyRepeat {
		return false
	}

	switch e.Key {
	case event.KeyDown:
		m.moveHighlight(1)
		ctx.Invalidate()
		return true
	case event.KeyUp:
		m.moveHighlight(-1)
		ctx.Invalidate()
		return true
	case event.KeyEnter, event.KeySpace:
		if m.highlightedIndex >= 0 && m.highlightedIndex < len(m.items) {
			if !m.items[m.highlightedIndex].Disabled {
				m.selectItem(m.highlightedIndex)
			}
		}
		return true
	case event.KeyHome:
		m.highlightedIndex = m.findNextEnabled(0, 1)
		m.ensureVisible(m.highlightedIndex)
		ctx.Invalidate()
		return true
	case event.KeyEnd:
		m.highlightedIndex = m.findNextEnabled(len(m.items)-1, -1)
		m.ensureVisible(m.highlightedIndex)
		ctx.Invalidate()
		return true
	default:
		return false
	}
}

// handleMouseEvent processes hover and click events.
func (m *menuWidget) handleMouseEvent(ctx widget.Context, e *event.MouseEvent) bool {
	bounds := m.Bounds()
	if !bounds.Contains(e.Position) {
		return false
	}

	switch e.MouseType {
	case event.MouseMove:
		index := m.indexAtPosition(e.Position)
		if index != m.highlightedIndex {
			m.highlightedIndex = index
			ctx.Invalidate()
		}
		return true
	case event.MousePress:
		if e.Button != event.ButtonLeft {
			return false
		}
		index := m.indexAtPosition(e.Position)
		if index >= 0 && index < len(m.items) && !m.items[index].Disabled {
			m.selectItem(index)
		}
		return true
	default:
		return true // consume other mouse events within bounds
	}
}

// handleWheelEvent processes scroll wheel events.
func (m *menuWidget) handleWheelEvent(ctx widget.Context, e *event.WheelEvent) bool {
	bounds := m.Bounds()
	if !bounds.Contains(e.Position) {
		return false
	}

	maxScroll := len(m.items) - m.visibleCount()
	if maxScroll < 0 {
		maxScroll = 0
	}

	if e.Delta.Y > 0 {
		// Scroll up.
		if m.scrollOffset > 0 {
			m.scrollOffset--
			ctx.Invalidate()
		}
	} else if e.Delta.Y < 0 {
		// Scroll down.
		if m.scrollOffset < maxScroll {
			m.scrollOffset++
			ctx.Invalidate()
		}
	}
	return true
}

// moveHighlight moves the highlight by delta, skipping disabled items.
func (m *menuWidget) moveHighlight(delta int) {
	if len(m.items) == 0 {
		return
	}

	next := m.highlightedIndex + delta
	next = m.findNextEnabled(next, delta)

	if next >= 0 && next < len(m.items) {
		m.highlightedIndex = next
		m.ensureVisible(next)
	}
}

// findNextEnabled finds the next non-disabled item starting from index,
// moving in the given direction (1 or -1).
func (m *menuWidget) findNextEnabled(start, direction int) int {
	if direction == 0 {
		direction = 1
	}
	for i := start; i >= 0 && i < len(m.items); i += direction {
		if !m.items[i].Disabled {
			return i
		}
	}
	return m.highlightedIndex // stay in place if nothing found
}

// ensureVisible adjusts scrollOffset so that the given index is visible.
func (m *menuWidget) ensureVisible(index int) {
	if index < 0 {
		return
	}
	visible := m.visibleCount()
	if index < m.scrollOffset {
		m.scrollOffset = index
	} else if index >= m.scrollOffset+visible {
		m.scrollOffset = index - visible + 1
	}
	// Clamp.
	maxScroll := len(m.items) - visible
	if maxScroll < 0 {
		maxScroll = 0
	}
	if m.scrollOffset > maxScroll {
		m.scrollOffset = maxScroll
	}
	if m.scrollOffset < 0 {
		m.scrollOffset = 0
	}
}

// visibleCount returns how many items are visible (capped by maxVisible).
func (m *menuWidget) visibleCount() int {
	count := len(m.items)
	if m.maxVisible > 0 && count > m.maxVisible {
		count = m.maxVisible
	}
	return count
}

// indexAtPosition returns the item index at the given mouse position,
// or -1 if outside the items area.
func (m *menuWidget) indexAtPosition(pos geometry.Point) int {
	bounds := m.Bounds()
	localY := pos.Y - bounds.Min.Y
	if localY < 0 {
		return -1
	}
	row := int(localY / m.itemHeight)
	index := m.scrollOffset + row
	if index < 0 || index >= len(m.items) {
		return -1
	}
	return index
}

// selectItem triggers selection of the given index.
func (m *menuWidget) selectItem(index int) {
	if m.onSelect != nil {
		m.onSelect(index)
	}
}

// Compile-time check.
var _ widget.Widget = (*menuWidget)(nil)
