package treeview

import (
	"github.com/gogpu/ui/event"
	"github.com/gogpu/ui/geometry"
	"github.com/gogpu/ui/widget"
)

// treeContent reports the full height to ScrollView and paints visible rows only.
type treeContent struct {
	widget.WidgetBase
	tree *Widget
}

func (c *treeContent) Layout(_ widget.Context, constraints geometry.Constraints) geometry.Size {
	return constraints.Constrain(geometry.Sz(c.tree.viewportWidth, c.tree.contentHeight()))
}

func (c *treeContent) Draw(_ widget.Context, canvas widget.Canvas) {
	tree := c.tree
	inset := tree.scroll.ScrollbarInset()
	// ScrollView has already translated the canvas by the negative scroll offset.
	bounds := geometry.NewRect(0, tree.scrollY.Get(), max(0, tree.viewportWidth-inset), tree.viewportHeight)
	if inset > 0 {
		canvas.PushClip(bounds)
		defer canvas.PopClip()
	}
	tree.drawRows(canvas, bounds)
}

func (c *treeContent) Event(ctx widget.Context, e event.Event) bool {
	// The window sends hover transitions directly to the deepest widget.
	// Row interactions and scrollbar drags are handled by the tree itself.
	if mouse, ok := e.(*event.MouseEvent); ok && (mouse.MouseType == event.MouseEnter || mouse.MouseType == event.MouseLeave) {
		return c.tree.Event(ctx, e)
	}
	return false
}
