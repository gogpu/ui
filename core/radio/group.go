package radio

import (
	"sync"

	"github.com/gogpu/ui/event"
	"github.com/gogpu/ui/geometry"
	"github.com/gogpu/ui/state"
	"github.com/gogpu/ui/widget"
)

// Group is a container widget that holds a set of mutually-exclusive radio items.
//
// Exactly one item may be selected at a time. Create a group with [NewGroup]:
//
//	rg := radio.NewGroup(
//	    radio.Items(
//	        radio.ItemDef{Value: "s", Label: "Small"},
//	        radio.ItemDef{Value: "m", Label: "Medium"},
//	    ),
//	    radio.Selected("m"),
//	    radio.OnChange(func(v string) { fmt.Println("selected:", v) }),
//	)
type Group struct {
	widget.WidgetBase
	mu       sync.RWMutex
	cfg      groupConfig
	items    []*Item
	selected int // index of selected item, -1 = none
}

// NewGroup creates a new radio Group with the given options.
//
// The returned group is visible, enabled, and lays out items vertically
// by default. Use options to configure items, selected value, direction,
// and change handler.
func NewGroup(opts ...GroupOption) *Group {
	g := &Group{
		selected: -1,
	}
	g.SetVisible(true)
	g.SetEnabled(true)

	for _, opt := range opts {
		opt(&g.cfg)
	}

	// Determine the painter.
	painter := Painter(DefaultPainter{})
	if g.cfg.painter != nil {
		painter = g.cfg.painter
	}

	// Create items from definitions.
	for _, def := range g.cfg.items {
		it := newItem(def, g, painter)
		g.items = append(g.items, it)
	}

	// Apply initial selection.
	// Use ResolvedSelected to support signal-based initial values.
	initialSelected := g.cfg.ResolvedSelected()
	if initialSelected != "" {
		for i, it := range g.items {
			if it.value == initialSelected {
				g.selected = i
				break
			}
		}
	}

	return g
}

// Selected returns the value of the currently selected item.
// When a [SelectedSignal] is bound, the signal value is returned directly.
// Returns an empty string if no item is selected.
func (g *Group) Selected() string {
	g.mu.RLock()
	defer g.mu.RUnlock()
	if g.cfg.selectedSignal != nil {
		return g.cfg.selectedSignal.Get()
	}
	if g.selected < 0 || g.selected >= len(g.items) {
		return ""
	}
	return g.items[g.selected].value
}

// Select programmatically selects the item with the given value.
// If no item matches the value, the selection is cleared.
func (g *Group) Select(value string) {
	g.mu.Lock()
	defer g.mu.Unlock()
	g.setSelectedLocked(value)
}

// setSelectedLocked updates the selected index. Caller must hold mu.
func (g *Group) setSelectedLocked(value string) {
	g.selected = -1
	for i, it := range g.items {
		if it.value == value {
			g.selected = i
			return
		}
	}
}

// selectValue is called by items when they are activated.
func (g *Group) selectValue(value string) {
	g.mu.Lock()
	prev := g.selected
	g.setSelectedLocked(value)
	changed := g.selected != prev
	onChange := g.cfg.onChange
	selectedSignal := g.cfg.selectedSignal
	g.mu.Unlock()

	if !changed {
		return
	}

	// TWO-WAY: if a SelectedSignal is bound, write back the new value.
	if selectedSignal != nil {
		selectedSignal.Set(value)
	}

	if onChange != nil {
		onChange(value)
	}
}

// isSelected reports whether the item with the given value is currently selected.
func (g *Group) isSelected(value string) bool {
	g.mu.RLock()
	defer g.mu.RUnlock()
	if g.cfg.selectedSignal != nil {
		return g.cfg.selectedSignal.Get() == value
	}
	if g.selected < 0 || g.selected >= len(g.items) {
		return false
	}
	return g.items[g.selected].value == value
}

// moveFocus moves focus from the current item by delta items (+1 forward, -1 backward).
// Wraps around at the ends.
func (g *Group) moveFocus(current *Item, ctx widget.Context, delta int) {
	g.mu.RLock()
	defer g.mu.RUnlock()

	idx := -1
	for i, it := range g.items {
		if it == current {
			idx = i
			break
		}
	}
	if idx < 0 {
		return
	}

	n := len(g.items)
	next := (idx + delta + n) % n
	ctx.RequestFocus(g.items[next])
}

// ItemCount returns the number of items in the group.
func (g *Group) ItemCount() int {
	g.mu.RLock()
	defer g.mu.RUnlock()
	return len(g.items)
}

// ItemAt returns the item at the given index.
// Panics if index is out of range.
func (g *Group) ItemAt(index int) *Item {
	g.mu.RLock()
	defer g.mu.RUnlock()
	return g.items[index]
}

// IsFocusable reports whether the group itself can receive focus.
// The group delegates focus to its items, so it always returns false.
func (g *Group) IsFocusable() bool {
	return false
}

// Layout calculates the group's preferred size by laying out all items.
func (g *Group) Layout(ctx widget.Context, constraints geometry.Constraints) geometry.Size {
	g.mu.RLock()
	defer g.mu.RUnlock()

	if len(g.items) == 0 {
		return geometry.Size{}
	}

	var preferred geometry.Size
	if g.cfg.direction == Vertical {
		preferred = g.layoutVertical(ctx, constraints)
	} else {
		preferred = g.layoutHorizontal(ctx, constraints)
	}

	return constraints.Constrain(preferred)
}

// layoutVertical positions items from top to bottom and returns the total size.
// Items are positioned in the group's local coordinate space starting at (0,0).
// The parent widget sets the group's bounds; Draw applies PushTransform.
// Caller must hold mu.
func (g *Group) layoutVertical(ctx widget.Context, constraints geometry.Constraints) geometry.Size {
	var totalWidth, totalHeight float32

	// Items size to their content — use loose constraints with parent max bounds.
	itemConstraints := geometry.Loose(geometry.Sz(constraints.MaxWidth, constraints.MaxHeight))

	for i, it := range g.items {
		itemSize := it.Layout(ctx, itemConstraints)
		if i > 0 {
			totalHeight += itemSpacing
		}
		it.SetBounds(geometry.NewRect(0, totalHeight, itemSize.Width, itemSize.Height))
		totalHeight += itemSize.Height
		if itemSize.Width > totalWidth {
			totalWidth = itemSize.Width
		}
	}

	return geometry.Sz(totalWidth, totalHeight)
}

// layoutHorizontal positions items from left to right and returns the total size.
// Items are positioned in the group's local coordinate space starting at (0,0).
// Caller must hold mu.
func (g *Group) layoutHorizontal(ctx widget.Context, constraints geometry.Constraints) geometry.Size {
	var totalWidth, totalHeight float32

	// Items size to their content — use loose constraints with parent max bounds.
	itemConstraints := geometry.Loose(geometry.Sz(constraints.MaxWidth, constraints.MaxHeight))

	for i, it := range g.items {
		itemSize := it.Layout(ctx, itemConstraints)
		if i > 0 {
			totalWidth += itemSpacing
		}
		it.SetBounds(geometry.NewRect(totalWidth, 0, itemSize.Width, itemSize.Height))
		totalWidth += itemSize.Width
		if itemSize.Height > totalHeight {
			totalHeight = itemSize.Height
		}
	}

	return geometry.Sz(totalWidth, totalHeight)
}

// itemSpacing is the spacing between items in the group.
const itemSpacing float32 = 4

// Draw renders all items to the canvas.
func (g *Group) Draw(ctx widget.Context, canvas widget.Canvas) {
	// Copy items under lock, then draw without holding the lock.
	// Item.Draw calls g.isSelected which acquires RLock — holding
	// RLock here would deadlock if a Lock waiter exists (RWMutex
	// is not reentrant in Go).
	g.mu.RLock()
	items := make([]*Item, len(g.items))
	copy(items, g.items)
	g.mu.RUnlock()

	// Items are positioned in local coordinates (0,0)-based.
	// Push the group's position so items render at the correct screen location.
	canvas.PushTransform(g.Bounds().Min)
	for _, it := range items {
		widget.StampScreenOrigin(it, canvas)
		it.Draw(ctx, canvas)
	}
	canvas.PopTransform()
}

// Event handles an input event by delegating to the appropriate item.
func (g *Group) Event(ctx widget.Context, e event.Event) bool {
	// Copy items under lock, then dispatch without holding the lock.
	// Item event handlers may call back into Group (selectValue),
	// which requires a write lock — holding RLock here would deadlock.
	g.mu.RLock()
	items := make([]*Item, len(g.items))
	copy(items, g.items)
	g.mu.RUnlock()

	// For mouse events, translate to Group-local coordinates and hit-test.
	// This mirrors PushTransform(g.Bounds().Min) used in Draw.
	if me, ok := e.(*event.MouseEvent); ok {
		local := *me
		local.Position = me.Position.Sub(g.Bounds().Min)
		for _, it := range items {
			if it.Bounds().Contains(local.Position) {
				if it.Event(ctx, &local) {
					return true
				}
			}
		}
		return false
	}

	for _, it := range items {
		if it.Event(ctx, e) {
			return true
		}
	}
	return false
}

// Children returns the items as a slice of widget.Widget.
func (g *Group) Children() []widget.Widget {
	g.mu.RLock()
	defer g.mu.RUnlock()

	children := make([]widget.Widget, len(g.items))
	for i, it := range g.items {
		children[i] = it
	}
	return children
}

// Mount creates signal bindings for push-based invalidation.
// Implements [widget.Lifecycle].
func (g *Group) Mount(ctx widget.Context) {
	sched := ctx.Scheduler()
	if sched == nil {
		return
	}
	if g.cfg.selectedSignal != nil {
		b := state.BindToScheduler(g.cfg.selectedSignal, g, sched)
		g.AddBinding(b)
	}
	if g.cfg.readonlyDisabledSig != nil {
		b := state.BindToScheduler(g.cfg.readonlyDisabledSig, g, sched)
		g.AddBinding(b)
	} else if g.cfg.disabledSignal != nil {
		b := state.BindToScheduler(g.cfg.disabledSignal, g, sched)
		g.AddBinding(b)
	}
}

// Unmount is called when the radio group is removed from the widget tree.
// Implements [widget.Lifecycle].
func (g *Group) Unmount() {
	// Bindings are cleaned up automatically by WidgetBase.CleanupBindings().
}

// Verify Group implements required interfaces at compile time.
var (
	_ widget.Widget    = (*Group)(nil)
	_ widget.Lifecycle = (*Group)(nil)
)
