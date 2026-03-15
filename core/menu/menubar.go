package menu

import (
	"github.com/gogpu/ui/event"
	"github.com/gogpu/ui/geometry"
	"github.com/gogpu/ui/overlay"
	"github.com/gogpu/ui/widget"
)

// Bar is a horizontal menu bar widget displayed at the top of a window.
// It contains top-level menu labels that open dropdown menus when clicked.
//
// Bar follows the functional options pattern for construction.
// Create with [NewBar].
type Bar struct {
	widget.WidgetBase

	menus        []TopMenu
	menuRects    []geometry.Rect
	painter      Painter
	openIndex    int // index of open top-level menu (-1 for none)
	hoveredIndex int // index of hovered label (-1 for none)

	// Active menu panel state.
	activePanel *menuPanel
}

// BarOption configures a Bar during construction.
type BarOption func(*Bar)

// PainterOpt sets the painter used to render the menu bar and its menus.
func PainterOpt(p Painter) BarOption {
	return func(b *Bar) {
		b.painter = p
	}
}

// NewBar creates a new MenuBar widget with the given top-level menus and options.
//
// The returned widget is visible and enabled by default.
func NewBar(menus []TopMenu, opts ...BarOption) *Bar {
	b := &Bar{
		menus:        menus,
		painter:      DefaultPainter{},
		openIndex:    -1,
		hoveredIndex: -1,
	}
	b.SetVisible(true)
	b.SetEnabled(true)

	for _, opt := range opts {
		opt(b)
	}

	return b
}

// IsFocusable reports whether the menu bar can currently receive focus.
func (b *Bar) IsFocusable() bool {
	return b.IsVisible() && b.IsEnabled()
}

// Layout calculates the menu bar's preferred size.
// The bar occupies the full available width and a fixed height.
func (b *Bar) Layout(_ widget.Context, constraints geometry.Constraints) geometry.Size {
	preferred := geometry.Sz(constraints.MaxWidth, dfltBarHeight)
	if preferred.Width >= geometry.Infinity {
		preferred.Width = 400 // reasonable fallback
	}
	size := constraints.Constrain(preferred)

	// Compute label rectangles.
	b.computeMenuRects(size)
	return size
}

// computeMenuRects calculates the bounding rectangle for each top-level menu label.
func (b *Bar) computeMenuRects(barSize geometry.Size) {
	b.menuRects = make([]geometry.Rect, len(b.menus))
	x := float32(0)
	for i, m := range b.menus {
		labelWidth := estimateTextWidth(m.Label, dfltFontSize) + dfltBarPaddingH*2
		if labelWidth < dfltBarMinLabelWidth {
			labelWidth = dfltBarMinLabelWidth
		}
		b.menuRects[i] = geometry.Rect{
			Min: geometry.Pt(x, 0),
			Max: geometry.Pt(x+labelWidth, barSize.Height),
		}
		x += labelWidth
	}
}

// dfltBarMinLabelWidth is the minimum width for a top-level menu label.
const dfltBarMinLabelWidth float32 = 48

// Draw renders the menu bar.
func (b *Bar) Draw(ctx widget.Context, canvas widget.Canvas) {
	if canvas == nil {
		return
	}
	b.painter.PaintMenuBar(canvas, &MenuBarPaintState{
		Bounds:       b.Bounds(),
		Menus:        b.menus,
		MenuRects:    b.toAbsoluteRects(),
		OpenIndex:    b.openIndex,
		HoveredIndex: b.hoveredIndex,
		Focused:      b.IsFocused(),
	})
}

// toAbsoluteRects converts relative menu rects to absolute window coordinates.
func (b *Bar) toAbsoluteRects() []geometry.Rect {
	bounds := b.Bounds()
	abs := make([]geometry.Rect, len(b.menuRects))
	for i, r := range b.menuRects {
		abs[i] = geometry.Rect{
			Min: geometry.Pt(r.Min.X+bounds.Min.X, r.Min.Y+bounds.Min.Y),
			Max: geometry.Pt(r.Max.X+bounds.Min.X, r.Max.Y+bounds.Min.Y),
		}
	}
	return abs
}

// Event handles input events for the menu bar.
func (b *Bar) Event(ctx widget.Context, e event.Event) bool {
	switch ev := e.(type) {
	case *event.MouseEvent:
		return b.handleMouseEvent(ctx, ev)
	case *event.KeyEvent:
		return b.handleKeyEvent(ctx, ev)
	default:
		return false
	}
}

// Children returns nil; the menu bar is a leaf widget (menus are overlays).
func (b *Bar) Children() []widget.Widget {
	return nil
}

// IsOpen returns true if any top-level menu is currently open.
func (b *Bar) IsOpen() bool {
	return b.openIndex >= 0
}

// OpenIndex returns the index of the currently open top-level menu, or -1.
func (b *Bar) OpenIndex() int {
	return b.openIndex
}

// HoveredIndex returns the index of the hovered label, or -1.
func (b *Bar) HoveredIndex() int {
	return b.hoveredIndex
}

// Menus returns the top-level menu definitions.
func (b *Bar) Menus() []TopMenu {
	return b.menus
}

// Close closes any open menu.
func (b *Bar) Close(ctx widget.Context) {
	b.closeMenu(ctx)
}

// handleMouseEvent processes mouse events on the menu bar.
func (b *Bar) handleMouseEvent(ctx widget.Context, e *event.MouseEvent) bool {
	bounds := b.Bounds()
	if !bounds.Contains(e.Position) {
		return false
	}

	switch e.MouseType {
	case event.MouseMove:
		index := b.indexAtPosition(e.Position)
		if index != b.hoveredIndex {
			b.hoveredIndex = index
			// If a menu is already open, hovering a different label opens that menu.
			if b.openIndex >= 0 && index >= 0 && index != b.openIndex {
				b.openMenu(ctx, index)
			}
			ctx.Invalidate()
		}
		return true

	case event.MouseLeave:
		b.hoveredIndex = -1
		ctx.Invalidate()
		return true

	case event.MousePress:
		if e.Button != event.ButtonLeft {
			return false
		}
		index := b.indexAtPosition(e.Position)
		if index >= 0 {
			ctx.RequestFocus(b)
			if b.openIndex == index {
				b.closeMenu(ctx)
			} else {
				b.openMenu(ctx, index)
			}
			return true
		}
		return false

	default:
		return false
	}
}

// handleKeyEvent processes keyboard navigation on the menu bar.
func (b *Bar) handleKeyEvent(ctx widget.Context, e *event.KeyEvent) bool {
	if !b.IsFocused() {
		return false
	}

	if e.KeyType != event.KeyPress && e.KeyType != event.KeyRepeat {
		return false
	}

	switch e.Key {
	case event.KeyLeft:
		return b.moveFocus(ctx, -1)
	case event.KeyRight:
		return b.moveFocus(ctx, 1)
	case event.KeyDown, event.KeyEnter, event.KeySpace:
		if b.openIndex < 0 && b.hoveredIndex >= 0 {
			b.openMenu(ctx, b.hoveredIndex)
			return true
		}
		if b.openIndex < 0 && len(b.menus) > 0 {
			b.openMenu(ctx, 0)
			return true
		}
		return false
	case event.KeyEscape:
		if b.openIndex >= 0 {
			b.closeMenu(ctx)
			return true
		}
		return false
	default:
		return false
	}
}

// moveFocus moves the highlighted label and opens its menu if one is already open.
func (b *Bar) moveFocus(ctx widget.Context, delta int) bool {
	if len(b.menus) == 0 {
		return false
	}

	current := b.hoveredIndex
	if current < 0 {
		current = b.openIndex
	}
	if current < 0 {
		current = 0
	} else {
		current += delta
		if current < 0 {
			current = len(b.menus) - 1
		} else if current >= len(b.menus) {
			current = 0
		}
	}

	b.hoveredIndex = current
	if b.openIndex >= 0 {
		b.openMenu(ctx, current)
	}
	ctx.Invalidate()
	return true
}

// openMenu opens the dropdown menu for the top-level item at index.
func (b *Bar) openMenu(ctx widget.Context, index int) {
	if index < 0 || index >= len(b.menus) {
		return
	}

	// Close existing menu first.
	if b.openIndex >= 0 {
		b.closeMenu(ctx)
	}

	om := ctx.OverlayManager()
	if om == nil {
		return
	}

	b.openIndex = index
	topMenu := b.menus[index]

	// Create menu panel.
	panel := newMenuPanel(topMenu.Items, b.painter, func(item *MenuItem) {
		b.closeMenu(ctx)
		if item != nil && item.OnAction != nil {
			item.OnAction()
		}
	}, func() {
		b.closeMenu(ctx)
	})

	// Position below the label.
	absRects := b.toAbsoluteRects()
	labelRect := absRects[index]
	menuSz := menuSize(topMenu.Items)
	windowSize := ctx.WindowSize()
	pos := overlay.Position(overlay.PlacementBelow, labelRect, menuSz, windowSize, 0)
	panel.SetBounds(geometry.FromPointSize(pos, menuSz))

	b.activePanel = panel

	om.PushOverlay(panel, func() {
		b.closeMenu(ctx)
	})

	ctx.Invalidate()
}

// closeMenu closes the currently open menu.
func (b *Bar) closeMenu(ctx widget.Context) {
	if b.openIndex < 0 {
		return
	}

	if b.activePanel != nil {
		b.activePanel.closeAllSubmenus(ctx)
		om := ctx.OverlayManager()
		if om != nil {
			om.RemoveOverlay(b.activePanel)
		}
		b.activePanel = nil
	}

	b.openIndex = -1
	ctx.Invalidate()
}

// indexAtPosition returns the top-level menu label index at the given position.
func (b *Bar) indexAtPosition(pos geometry.Point) int {
	absRects := b.toAbsoluteRects()
	for i, r := range absRects {
		if r.Contains(pos) {
			return i
		}
	}
	return -1
}

// A11yRole returns the ARIA role for the menu bar.
func (b *Bar) A11yRole() string {
	return a11yRoleMenuBar
}

// A11yLabel returns the accessibility label.
func (b *Bar) A11yLabel() string {
	return a11yLabelMenuBar
}

// Accessibility constants.
const (
	a11yRoleMenuBar  = "menubar"
	a11yLabelMenuBar = "menu bar"
)

// Compile-time interface checks.
var (
	_ widget.Widget    = (*Bar)(nil)
	_ widget.Focusable = (*Bar)(nil)
)
