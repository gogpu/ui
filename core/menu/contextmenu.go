package menu

import (
	"github.com/gogpu/ui/geometry"
	"github.com/gogpu/ui/overlay"
	"github.com/gogpu/ui/widget"
)

// ContextMenu is a floating popup menu shown at a specific position,
// typically on right-click. It contains a list of menu items that the
// user can select via mouse or keyboard.
//
// Create with [NewContextMenu].
type ContextMenu struct {
	items   []MenuItem
	painter Painter
	panel   *menuPanel
	open    bool
}

// ContextMenuOption configures a ContextMenu during construction.
type ContextMenuOption func(*ContextMenu)

// ContextPainterOpt sets the painter used to render the context menu.
func ContextPainterOpt(p Painter) ContextMenuOption {
	return func(cm *ContextMenu) {
		cm.painter = p
	}
}

// NewContextMenu creates a new context menu with the given items and options.
func NewContextMenu(items []MenuItem, opts ...ContextMenuOption) *ContextMenu {
	cm := &ContextMenu{
		items:   items,
		painter: DefaultPainter{},
	}
	for _, opt := range opts {
		opt(cm)
	}
	return cm
}

// Show opens the context menu at the given position via the overlay manager.
// If the context menu is already open, it is closed first.
func (cm *ContextMenu) Show(ctx widget.Context, position geometry.Point) {
	if cm.open {
		cm.Hide(ctx)
	}

	om := ctx.OverlayManager()
	if om == nil {
		return
	}

	cm.open = true

	panel := newMenuPanel(cm.items, cm.painter, func(item *MenuItem) {
		cm.Hide(ctx)
		if item != nil && item.OnAction != nil {
			item.OnAction()
		}
	}, func() {
		cm.Hide(ctx)
	})

	// Position at the given point, clamping to viewport.
	menuSz := menuSize(cm.items)
	windowSize := ctx.WindowSize()

	// Create a zero-size anchor at the click position.
	anchor := geometry.Rect{
		Min: position,
		Max: position,
	}
	pos := overlay.Position(overlay.PlacementBelow, anchor, menuSz, windowSize, 0)
	panel.SetBounds(geometry.FromPointSize(pos, menuSz))

	cm.panel = panel

	om.PushOverlay(panel, func() {
		cm.Hide(ctx)
	})

	ctx.Invalidate()
}

// Hide closes the context menu.
func (cm *ContextMenu) Hide(ctx widget.Context) {
	if !cm.open {
		return
	}

	if cm.panel != nil {
		cm.panel.closeAllSubmenus(ctx)
		om := ctx.OverlayManager()
		if om != nil {
			om.RemoveOverlay(cm.panel)
		}
		cm.panel = nil
	}

	cm.open = false
	ctx.Invalidate()
}

// IsOpen returns true if the context menu is currently visible.
func (cm *ContextMenu) IsOpen() bool {
	return cm.open
}

// Items returns the context menu items.
func (cm *ContextMenu) Items() []MenuItem {
	return cm.items
}

// Panel returns the active menu panel for testing. Returns nil when closed.
// The returned value implements [widget.Widget] and [PanelState].
func (cm *ContextMenu) Panel() PanelState {
	if cm.panel == nil {
		return nil
	}
	return cm.panel
}
