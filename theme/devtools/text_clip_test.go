package devtools_test

import (
	"testing"

	"github.com/gogpu/ui/core/datatable"
	"github.com/gogpu/ui/core/dropdown"
	"github.com/gogpu/ui/geometry"
	"github.com/gogpu/ui/theme/devtools"
	"github.com/gogpu/ui/widget"
)

type textClipCanvas struct {
	recordCanvas
	clips   []geometry.Rect
	clipped bool
}

func (c *textClipCanvas) PushClip(r geometry.Rect) { c.clips = append(c.clips, r) }
func (c *textClipCanvas) PopClip()                 { c.clips = c.clips[:len(c.clips)-1] }
func (c *textClipCanvas) DrawText(_ string, r geometry.Rect, _ float32, _ widget.Color, _ bool, _ widget.TextAlign) {
	c.clipped = len(c.clips) > 0 && c.clips[len(c.clips)-1] == r
}

func TestConfineTextToCellOrTrigger(t *testing.T) {
	bounds := geometry.NewRect(10, 20, 70, 26)
	long := "a long label that would overlap the following widget"
	table := devtools.DataTablePainter{}
	for name, paint := range map[string]func(widget.Canvas){
		"dropdown": func(c widget.Canvas) {
			devtools.DropdownPainter{}.PaintTrigger(c, &dropdown.TriggerPaintState{Bounds: bounds, SelectedText: long})
		},
		"header": func(c widget.Canvas) { table.PaintHeaderCell(c, bounds, datatable.HeaderCellPaintState{Title: long}) },
		"cell":   func(c widget.Canvas) { table.PaintCell(c, datatable.CellPaintState{Bounds: bounds, Value: long}) },
	} {
		t.Run(name, func(t *testing.T) {
			c := new(textClipCanvas)
			paint(c)
			if !c.clipped || len(c.clips) != 0 {
				t.Fatal("expected a balanced clip around the text bounds")
			}
		})
	}
}
