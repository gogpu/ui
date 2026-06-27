package radio

import (
	"github.com/gogpu/ui/event"
	"github.com/gogpu/ui/geometry"
	"github.com/gogpu/ui/widget"
)

// interactionState represents the current user interaction state of an item.
type interactionState uint8

const (
	stateNormal  interactionState = iota
	stateHover                    // mouse is over the item
	statePressed                  // mouse button is held down
)

// Item represents a single radio button within a [Group].
//
// Items are created automatically by [NewGroup] from the [ItemDef] list
// provided via the [Items] option. Each item holds a reference to its
// parent group for mutual-exclusion selection.
type Item struct {
	widget.WidgetBase
	value   string
	label   string
	group   *Group
	state   interactionState
	painter Painter
}

// newItem creates a new radio item linked to the given group.
func newItem(def ItemDef, group *Group, painter Painter) *Item {
	it := &Item{
		value:   def.Value,
		label:   def.Label,
		group:   group,
		painter: painter,
	}
	it.SetVisible(true)
	it.SetEnabled(true)
	return it
}

// Value returns the programmatic value of this radio item.
func (it *Item) Value() string {
	return it.value
}

// Label returns the display label of this radio item.
func (it *Item) Label() string {
	return it.label
}

// IsFocusable reports whether the item can currently receive focus.
// An item is focusable when it is visible, enabled, and its group is not disabled.
func (it *Item) IsFocusable() bool {
	return it.IsVisible() && it.IsEnabled() && !it.group.cfg.ResolvedDisabled()
}

// Layout calculates the item's preferred size within the given constraints.
func (it *Item) Layout(_ widget.Context, constraints geometry.Constraints) geometry.Size {
	// Query LayoutMetrics from painter (type assert with default fallback).
	lm := resolveRadioLayoutMetrics(it.painter)

	radius := lm.RadioCircleRadius()
	pad := lm.RadioItemPadding()
	gap := lm.RadioLabelGap()
	fontSize := lm.RadioFontSize()

	totalWidth := radius*2 + pad*2
	totalHeight := radius*2 + pad*2

	if it.label != "" {
		textWidth := float32(len(it.label)) * fontSize * charWidthRatio
		totalWidth += gap + textWidth
	}

	if totalHeight < itemMinHeight {
		totalHeight = itemMinHeight
	}

	preferred := geometry.Sz(totalWidth, totalHeight)
	return constraints.Constrain(preferred)
}

// charWidthRatio is an approximate ratio of character width to font size
// for text width estimation in layout.
const charWidthRatio float32 = 0.55

// itemPadding is the padding around the radio circle.
const itemPadding float32 = 4

// itemMinHeight is the minimum item height for comfortable touch targets.
const itemMinHeight float32 = 24

// Draw renders the radio item to the canvas.
func (it *Item) Draw(ctx widget.Context, canvas widget.Canvas) {
	selected := it.group.isSelected(it.value)
	it.painter.PaintRadio(canvas, PaintState{
		Label:    it.label,
		Selected: selected,
		Hovered:  it.state == stateHover,
		Pressed:  it.state == statePressed,
		Focused:  it.IsFocused(),
		Disabled: it.group.cfg.ResolvedDisabled(),
		Bounds:   it.Bounds(),
	})
}

// Event handles an input event and returns true if consumed.
func (it *Item) Event(ctx widget.Context, e event.Event) bool {
	return handleItemEvent(it, ctx, e)
}

// Children returns nil because a radio item is a leaf widget.
func (it *Item) Children() []widget.Widget {
	return nil
}

// Verify Item implements required interfaces at compile time.
var (
	_ widget.Widget    = (*Item)(nil)
	_ widget.Focusable = (*Item)(nil)
)
