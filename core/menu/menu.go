package menu

import (
	"github.com/gogpu/ui/event"
	"github.com/gogpu/ui/geometry"
	"github.com/gogpu/ui/overlay"
	"github.com/gogpu/ui/widget"
)

// itemHeight is the default height of each menu item row.
const itemHeight float32 = 36

// separatorHeight is the default height of separator rows.
const separatorHeight float32 = 9

// menuMinWidth is the minimum width for a popup menu.
const menuMinWidth float32 = 160

// subMenuDelay is not enforced here; submenus open on hover immediately
// to keep the implementation simple and testable.

// menuPanel is an internal widget that renders a popup menu with items.
// It handles keyboard navigation, mouse hover, mouse click, and submenu display.
type menuPanel struct {
	widget.WidgetBase

	items            []MenuItem
	highlightedIndex int
	painter          Painter
	onSelect         func(item *MenuItem)
	onClose          func()

	// Submenu state.
	subMenuPanel *menuPanel
	subMenuIndex int // index of item whose submenu is open (-1 for none)

	ctx widget.Context // captured for submenu management
}

// newMenuPanel creates a popup menu panel with the given items.
func newMenuPanel(items []MenuItem, painter Painter, onSelect func(item *MenuItem), onClose func()) *menuPanel {
	m := &menuPanel{
		items:            items,
		highlightedIndex: -1,
		painter:          painter,
		onSelect:         onSelect,
		onClose:          onClose,
		subMenuIndex:     -1,
	}
	m.SetVisible(true)
	m.SetEnabled(true)
	return m
}

// menuSize calculates the total size of a menu panel given its items.
func menuSize(items []MenuItem) geometry.Size {
	height := dfltMenuPaddingV * 2 // top + bottom padding
	maxLabelWidth := menuMinWidth
	for _, item := range items {
		if item.IsSeparator() {
			height += separatorHeight
		} else {
			height += itemHeight
			// Estimate width from label + shortcut text.
			labelWidth := estimateTextWidth(item.Label, dfltFontSize)
			shortcutWidth := estimateTextWidth(item.Shortcut, dfltShortcutFontSize)
			totalWidth := dfltMenuItemPaddingH*2 + labelWidth + shortcutWidth + dfltShortcutAreaWidth
			if totalWidth > maxLabelWidth {
				maxLabelWidth = totalWidth
			}
		}
	}
	return geometry.Sz(maxLabelWidth, height)
}

// estimateTextWidth provides a rough character-width estimate for sizing.
// This is a heuristic; the real painter may produce different widths.
func estimateTextWidth(text string, fontSize float32) float32 {
	const charWidthFactor = 0.55 // approximate ratio of char width to font size
	return float32(len(text)) * fontSize * charWidthFactor
}

// Layout returns the panel's preferred size.
func (m *menuPanel) Layout(_ widget.Context, constraints geometry.Constraints) geometry.Size {
	return constraints.Constrain(menuSize(m.items))
}

// Draw renders the menu panel.
func (m *menuPanel) Draw(ctx widget.Context, canvas widget.Canvas) {
	if canvas == nil {
		return
	}
	m.painter.PaintMenu(canvas, &MenuPaintState{
		Bounds:           m.Bounds(),
		Items:            m.items,
		HighlightedIndex: m.highlightedIndex,
		ItemHeight:       itemHeight,
		SeparatorHeight:  separatorHeight,
		SubMenuOpenIndex: m.subMenuIndex,
	})
}

// Event handles input events for the popup menu.
func (m *menuPanel) Event(ctx widget.Context, e event.Event) bool {
	// Let submenu handle events first.
	if m.subMenuPanel != nil {
		if m.subMenuPanel.Event(ctx, e) {
			return true
		}
	}

	switch ev := e.(type) {
	case *event.KeyEvent:
		return m.handleKeyEvent(ctx, ev)
	case *event.MouseEvent:
		return m.handleMouseEvent(ctx, ev)
	default:
		return false
	}
}

// Children returns nil; the menu is a leaf widget (submenus are overlays).
func (m *menuPanel) Children() []widget.Widget {
	return nil
}

// handleKeyEvent processes keyboard navigation.
func (m *menuPanel) handleKeyEvent(ctx widget.Context, e *event.KeyEvent) bool {
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
		return m.activateHighlighted(ctx)
	case event.KeyRight:
		return m.openHighlightedSubmenu(ctx)
	case event.KeyLeft:
		return m.closeSubmenuOrSelf(ctx)
	case event.KeyEscape:
		m.closeAllSubmenus(ctx)
		if m.onClose != nil {
			m.onClose()
		}
		return true
	default:
		return false
	}
}

// handleMouseEvent processes mouse hover and click events.
func (m *menuPanel) handleMouseEvent(ctx widget.Context, e *event.MouseEvent) bool {
	bounds := m.Bounds()
	if !bounds.Contains(e.Position) {
		return false
	}

	switch e.MouseType {
	case event.MouseMove:
		index := m.indexAtPosition(e.Position)
		if index != m.highlightedIndex {
			m.highlightedIndex = index
			m.handleHoverSubmenu(ctx, index)
			ctx.Invalidate()
		}
		return true
	case event.MousePress:
		if e.Button != event.ButtonLeft {
			return false
		}
		index := m.indexAtPosition(e.Position)
		if index >= 0 && index < len(m.items) {
			item := &m.items[index]
			if !item.Disabled && !item.IsSeparator() && !item.HasChildren() {
				m.selectItem(item)
			}
		}
		return true
	default:
		return true // consume other mouse events within bounds
	}
}

// moveHighlight moves the highlight by delta, skipping separators and disabled items.
func (m *menuPanel) moveHighlight(delta int) {
	if len(m.items) == 0 {
		return
	}

	next := m.highlightedIndex + delta
	next = m.findNextEnabled(next, delta)
	if next >= 0 && next < len(m.items) {
		m.highlightedIndex = next
	}
}

// findNextEnabled finds the next non-disabled, non-separator item starting from start.
func (m *menuPanel) findNextEnabled(start, direction int) int {
	if direction == 0 {
		direction = 1
	}
	for i := start; i >= 0 && i < len(m.items); i += direction {
		if !m.items[i].Disabled && !m.items[i].IsSeparator() {
			return i
		}
	}
	return m.highlightedIndex // stay in place if nothing found
}

// activateHighlighted activates the highlighted item (enter/space).
func (m *menuPanel) activateHighlighted(ctx widget.Context) bool {
	if m.highlightedIndex < 0 || m.highlightedIndex >= len(m.items) {
		return false
	}
	item := &m.items[m.highlightedIndex]
	if item.Disabled || item.IsSeparator() {
		return true // consume but don't act
	}
	if item.HasChildren() {
		return m.openHighlightedSubmenu(ctx)
	}
	m.selectItem(item)
	return true
}

// selectItem fires the item's action and closes the entire menu tree.
func (m *menuPanel) selectItem(item *MenuItem) {
	if m.onSelect != nil {
		m.onSelect(item)
	}
}

// openHighlightedSubmenu opens the submenu for the highlighted item.
func (m *menuPanel) openHighlightedSubmenu(ctx widget.Context) bool {
	if m.highlightedIndex < 0 || m.highlightedIndex >= len(m.items) {
		return false
	}
	item := &m.items[m.highlightedIndex]
	if !item.HasChildren() || item.Disabled {
		return false
	}
	m.openSubmenu(ctx, m.highlightedIndex)
	return true
}

// handleHoverSubmenu opens or closes submenus based on hover index.
func (m *menuPanel) handleHoverSubmenu(ctx widget.Context, index int) {
	if index < 0 || index >= len(m.items) {
		m.closeAllSubmenus(ctx)
		return
	}
	item := &m.items[index]
	if item.HasChildren() && !item.Disabled {
		m.openSubmenu(ctx, index)
	} else if m.subMenuIndex != -1 {
		m.closeAllSubmenus(ctx)
	}
}

// openSubmenu opens a submenu for the item at the given index.
func (m *menuPanel) openSubmenu(ctx widget.Context, index int) {
	if m.subMenuIndex == index {
		return // already open
	}

	// Close existing submenu.
	m.closeAllSubmenus(ctx)

	item := &m.items[index]
	if !item.HasChildren() {
		return
	}

	m.subMenuIndex = index
	m.ctx = ctx

	sub := newMenuPanel(item.Children, m.painter, m.onSelect, func() {
		m.closeAllSubmenus(ctx)
	})

	// Position submenu to the right of the parent item.
	bounds := m.Bounds()
	itemY := m.yForIndex(index)
	subSize := menuSize(item.Children)
	windowSize := ctx.WindowSize()

	anchor := geometry.Rect{
		Min: geometry.Pt(bounds.Max.X, itemY),
		Max: geometry.Pt(bounds.Max.X, itemY+itemHeight),
	}
	pos := overlay.Position(overlay.PlacementRight, anchor, subSize, windowSize, 0)
	sub.SetBounds(geometry.FromPointSize(pos, subSize))

	m.subMenuPanel = sub
}

// closeAllSubmenus closes any open submenus recursively.
func (m *menuPanel) closeAllSubmenus(ctx widget.Context) {
	if m.subMenuPanel != nil {
		m.subMenuPanel.closeAllSubmenus(ctx)
		m.subMenuPanel = nil
	}
	m.subMenuIndex = -1
	ctx.Invalidate()
}

// closeSubmenuOrSelf closes submenu if open, otherwise signals parent to close.
func (m *menuPanel) closeSubmenuOrSelf(ctx widget.Context) bool {
	if m.subMenuPanel != nil {
		m.closeAllSubmenus(ctx)
		return true
	}
	if m.onClose != nil {
		m.onClose()
		return true
	}
	return false
}

// indexAtPosition returns the item index at the given mouse position.
func (m *menuPanel) indexAtPosition(pos geometry.Point) int {
	bounds := m.Bounds()
	localY := pos.Y - bounds.Min.Y - dfltMenuPaddingV
	if localY < 0 {
		return -1
	}

	y := float32(0)
	for i, item := range m.items {
		h := itemHeight
		if item.IsSeparator() {
			h = separatorHeight
		}
		if localY >= y && localY < y+h {
			if item.IsSeparator() {
				return -1 // separators are not selectable
			}
			return i
		}
		y += h
	}
	return -1
}

// yForIndex returns the Y coordinate for the top of the item at the given index.
func (m *menuPanel) yForIndex(index int) float32 {
	y := m.Bounds().Min.Y + dfltMenuPaddingV
	for i := 0; i < index && i < len(m.items); i++ {
		if m.items[i].IsSeparator() {
			y += separatorHeight
		} else {
			y += itemHeight
		}
	}
	return y
}

// PanelState is the exported interface for accessing menu panel state.
// This is primarily used for testing.
type PanelState interface {
	widget.Widget

	// Bounds returns the panel's bounding rectangle.
	Bounds() geometry.Rect

	// HighlightedIndex returns the currently highlighted item index (-1 for none).
	HighlightedIndex() int

	// SubMenuIndex returns the index of the item whose submenu is open (-1 for none).
	SubMenuIndex() int

	// SubMenuPanel returns the open submenu panel, or nil.
	SubMenuPanel() PanelState
}

// HighlightedIndex returns the currently highlighted item index for testing.
func (m *menuPanel) HighlightedIndex() int {
	return m.highlightedIndex
}

// SubMenuIndex returns the index of the item whose submenu is open for testing.
func (m *menuPanel) SubMenuIndex() int {
	return m.subMenuIndex
}

// SubMenuPanel returns the open submenu panel for testing.
func (m *menuPanel) SubMenuPanel() PanelState {
	if m.subMenuPanel == nil {
		return nil
	}
	return m.subMenuPanel
}

// Compile-time check.
var _ widget.Widget = (*menuPanel)(nil)
