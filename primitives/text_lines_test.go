package primitives_test

import (
	"slices"
	"testing"

	"github.com/gogpu/ui/geometry"
	"github.com/gogpu/ui/primitives"
	"github.com/gogpu/ui/widget"
)

type lineCanvas struct {
	styledMockCanvas
	lines  []string
	bounds []geometry.Rect
}

func (c *lineCanvas) DrawText(text string, bounds geometry.Rect, _ float32, _ widget.Color, _ bool, _ widget.TextAlign) {
	c.lines = append(c.lines, text)
	c.bounds = append(c.bounds, bounds)
}
func TestDrawTextLines(t *testing.T) {
	ctx := widget.NewContext()
	for _, test := range []struct {
		name, text string
		maxLines   int
		want       []string
	}{
		{"newlines", "one\r\n\ntwo", 0, []string{"one", "", "two"}},
		{"word wrap", "hello world", 0, []string{"hello", "world"}},
		{"limited unicode", "абвгдежз", 1, []string{"абвг…"}},
	} {
		t.Run(test.name, func(t *testing.T) {
			tw := primitives.Text(test.text).FontSize(10).LineHeight(1.5).MaxLines(test.maxLines).Ellipsis()
			size := tw.Layout(ctx, geometry.Loose(geometry.Sz(30, 200)))
			c := new(lineCanvas)
			tw.Draw(ctx, c)
			if !slices.Equal(c.lines, test.want) {
				t.Fatalf("drew %q, want %q", c.lines, test.want)
			}
			if size.Height != float32(len(test.want))*15 {
				t.Fatalf("height %v for %d lines", size.Height, len(test.want))
			}
			for i, b := range c.bounds {
				if b.Min.Y != float32(i)*15 {
					t.Fatalf("line %d at %v", i, b)
				}
			}
		})
	}
}
