package listview

import (
	"github.com/gogpu/ui/event"
	"github.com/gogpu/ui/geometry"
	"github.com/gogpu/ui/widget"
)

// itemDecorator wraps each user-created item widget to own the hover/selection
// painting. The decorator itself IS the RepaintBoundary — when hover state
// changes, only the decorator's scene is re-recorded, not the entire ListView.
//
// This matches the enterprise pattern: Flutter, Android, Qt, Masonry/Xilem,
// Gio, and Web/CSS all draw hover INSIDE the item, not in the parent.
type itemDecorator struct {
	widget.WidgetBase
	child widget.Widget
	list  *Widget // back-reference for hover/selection/painter state
	index int     // data index (absolute, not offset)
}

// newItemDecorator creates a decorator wrapping the user widget.
// The decorator is marked as a RepaintBoundary so hover/selection changes
// only dirty this single item, not the entire ListView.
func newItemDecorator(child widget.Widget, list *Widget, index int) *itemDecorator {
	d := &itemDecorator{
		child: child,
		list:  list,
		index: index,
	}
	d.SetVisible(true)
	d.SetRepaintBoundary(true)
	d.SetNeedsRedraw(true)
	return d
}

// Layout delegates to the child widget.
func (d *itemDecorator) Layout(ctx widget.Context, c geometry.Constraints) geometry.Size {
	if d.child == nil {
		return geometry.Size{}
	}
	return d.child.Layout(ctx, c)
}

// Draw paints hover/selection background, then the child content.
// The decorator reads hover/selection state from the list at Draw time,
// so no widget rebuild is needed for visual-only state changes.
func (d *itemDecorator) Draw(ctx widget.Context, canvas widget.Canvas) {
	if d.list == nil {
		return
	}

	bounds := d.Bounds()
	selectedIdx := d.list.cfg.ResolvedSelectedIndex()
	ips := ItemPaintState{
		Bounds:   bounds,
		Index:    d.index,
		Selected: d.index == selectedIdx,
		Focused:  d.index == selectedIdx && d.list.IsFocused(),
		Hovered:  d.index == d.list.hoveredIndex,
		Disabled: d.list.cfg.ResolvedDisabled(),
	}

	d.list.painter.PaintItemBackground(canvas, ips)

	if ips.Selected {
		d.list.painter.PaintSelection(canvas, ips)
	}

	if d.child != nil {
		// Child is NOT a boundary — call Draw directly.
		// DrawChild would skip it if it were a boundary, but the user widget
		// should never be a boundary here (decorator owns the boundary).
		if setter, ok := d.child.(interface{ SetBounds(geometry.Rect) }); ok {
			setter.SetBounds(bounds)
		}
		d.child.Draw(ctx, canvas)
	}
}

// Event returns false — events go through virtualContent, not the decorator.
func (d *itemDecorator) Event(_ widget.Context, _ event.Event) bool {
	return false
}

// Children returns the wrapped child widget.
func (d *itemDecorator) Children() []widget.Widget {
	if d.child == nil {
		return nil
	}
	return []widget.Widget{d.child}
}

// Compile-time interface check.
var _ widget.Widget = (*itemDecorator)(nil)
