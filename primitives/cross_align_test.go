package primitives_test

import (
	"testing"

	"github.com/gogpu/ui/geometry"
	"github.com/gogpu/ui/primitives"
	"github.com/gogpu/ui/widget"
)

func TestVBox_CrossAxisCenter(t *testing.T) {
	ctx := widget.NewContext()

	// 48px wide spinner in 400px wide VBox → should be centered at X=176.
	child := primitives.Text("X").FontSize(14)

	box := primitives.VBox(child).
		CrossAlign(primitives.CrossAxisCenter).
		Padding(0)

	constraints := geometry.BoxConstraints(400, 400, 0, 400)
	box.Layout(ctx, constraints)

	bounds := child.Bounds()
	boxBounds := box.Bounds()
	childW := bounds.Width()
	boxW := boxBounds.Width()
	expectedX := (boxW - childW) / 2

	t.Logf("box bounds: %v (width=%f)", boxBounds, boxW)
	t.Logf("child bounds: %v (width=%f)", bounds, childW)
	t.Logf("expectedX: %f", expectedX)

	if bounds.Min.X < expectedX-2 || bounds.Min.X > expectedX+2 {
		t.Errorf("child X = %f, want ~%f (centered in %fpx VBox, child width=%f)",
			bounds.Min.X, expectedX, boxW, childW)
	}
}

func TestVBox_CrossAxisStart_Default(t *testing.T) {
	ctx := widget.NewContext()

	child := primitives.Text("Left").FontSize(14)

	box := primitives.VBox(child).Padding(0)

	constraints := geometry.BoxConstraints(400, 400, 0, 400)
	box.Layout(ctx, constraints)

	bounds := child.Bounds()

	// Default = start = X should be 0 (left-aligned).
	if bounds.Min.X > 1 {
		t.Errorf("default cross-align: child X = %f, want 0 (left-aligned)", bounds.Min.X)
	}
}

func TestVBox_CrossAxisEnd(t *testing.T) {
	ctx := widget.NewContext()

	child := primitives.Text("Right").FontSize(14)

	box := primitives.VBox(child).
		CrossAlign(primitives.CrossAxisEnd).
		Padding(0)

	constraints := geometry.BoxConstraints(400, 400, 0, 400)
	box.Layout(ctx, constraints)

	bounds := child.Bounds()
	childW := bounds.Width()
	expectedX := 400 - childW

	if bounds.Min.X < expectedX-2 || bounds.Min.X > expectedX+2 {
		t.Errorf("child X = %f, want ~%f (end-aligned, width=%f)",
			bounds.Min.X, expectedX, childW)
	}
}

// HBox aligns children vertically below the top padding, both without and
// with an Expanded child (the two horizontal layout passes).
func TestHBox_CrossAxisAlignment(t *testing.T) {
	const padTop = 6
	for _, tc := range []struct {
		name       string
		align      primitives.CrossAxisAlignment
		expanded   bool
		wantShortY float32
	}{
		{"start", primitives.CrossAxisStart, false, padTop},
		{"start/expanded", primitives.CrossAxisStart, true, padTop},
		{"center", primitives.CrossAxisCenter, false, padTop + (40-10)/2},
		{"center/expanded", primitives.CrossAxisCenter, true, padTop + (40-10)/2},
		{"end", primitives.CrossAxisEnd, false, padTop + 40 - 10},
		{"end/expanded", primitives.CrossAxisEnd, true, padTop + 40 - 10},
	} {
		t.Run(tc.name, func(t *testing.T) {
			ctx := widget.NewContext()
			tall := primitives.Box().Width(10).Height(40)
			short := primitives.Box().Width(10).Height(10)
			var last widget.Widget = primitives.Box().Width(10).Height(20)
			if tc.expanded {
				last = primitives.Expanded(last)
			}
			box := primitives.HBox(tall, short, last).CrossAlign(tc.align).PaddingTop(padTop)

			box.Layout(ctx, geometry.BoxConstraints(0, 400, 0, 400))

			if y := tall.Bounds().Min.Y; y != padTop {
				t.Errorf("tallest child Y = %v, want %v", y, padTop)
			}
			if y := short.Bounds().Min.Y; y != tc.wantShortY {
				t.Errorf("short child Y = %v, want %v", y, tc.wantShortY)
			}
		})
	}
}

// VBox with an Expanded child (the two-pass vertical layout) centers
// narrower children like the single-pass layout does.
func TestVBox_CrossAxisCenter_WithExpanded(t *testing.T) {
	ctx := widget.NewContext()
	narrow := primitives.Box().Width(100).Height(10)
	wide := primitives.Box().Width(300).Height(10)
	box := primitives.VBox(narrow, wide, primitives.Expanded(primitives.Box())).
		CrossAlign(primitives.CrossAxisCenter)

	box.Layout(ctx, geometry.BoxConstraints(0, 400, 0, 400))

	if x := narrow.Bounds().Min.X; x != 100 {
		t.Errorf("narrow child X = %v, want 100 (centered under the 300px child)", x)
	}
}
