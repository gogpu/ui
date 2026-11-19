package datatable

import (
	"github.com/gogpu/ui/geometry"
	"github.com/gogpu/ui/widget"
)

type rowDetails struct {
	row     int
	height  float32
	content widget.Widget
}

// SetRowDetails inserts one interactive panel below a data row. Row indices
// continue to refer to data rows. Pass nil to remove the panel.
func (w *Widget) SetRowDetails(row int, content widget.Widget, height float32) {
	if content == nil || height <= 0 || row < 0 || row >= w.GetRowCount() {
		row, content, height = 0, nil, 0
	}
	next := rowDetails{row, height, content}
	if next == w.details {
		return
	}
	if w.details.content != content {
		w.virtual.ClearChildren()
		w.virtual.AddChild(content)
	}
	w.details = next
	w.InvalidateData()
}

func (w *Widget) detailsBounds(width float32) geometry.Rect {
	d := w.details
	if d.content == nil || d.row >= w.GetRowCount() {
		return geometry.Rect{}
	}
	return geometry.NewRect(0, float32(d.row+1)*w.cfg.rowHeight, width, d.height)
}

func (w *Widget) detailsContentBounds() geometry.Rect {
	return w.detailsBounds(max(0, w.viewportWidth-w.scroll.ScrollbarInset()))
}

func (w *Widget) layoutRowDetails(ctx widget.Context) {
	bounds := w.detailsContentBounds()
	if bounds.IsEmpty() {
		return
	}
	detail := w.details.content
	widget.LayoutChild(detail, ctx, geometry.Tight(bounds.Size()))
	if setter, ok := detail.(interface{ SetBounds(geometry.Rect) }); ok {
		setter.SetBounds(bounds)
	}
}

func (w *Widget) rowTop(row int) float32 {
	y := float32(row) * w.cfg.rowHeight
	if w.details.content != nil && row > w.details.row {
		y += w.detailsBounds(1).Height()
	}
	return y
}

// A viewport edge in the gap maps to the preceding row. Hit testing rejects
// the gap separately, since it belongs to the detail widget.
func (w *Widget) rowAtOffset(y float32) int {
	if gap := w.detailsBounds(1); !gap.IsEmpty() && y >= gap.Min.Y {
		if y < gap.Max.Y {
			return w.details.row
		}
		y -= gap.Height()
	}
	return int(y / w.cfg.rowHeight)
}
