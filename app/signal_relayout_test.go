package app

import (
	"testing"

	"github.com/gogpu/ui/geometry"
	"github.com/gogpu/ui/primitives"
	"github.com/gogpu/ui/state"
)

// frameAndPaint runs one frame the way desktop.Run does when the window has
// work: Frame (scheduler flush and layout), then a paint of the tree.
func frameAndPaint(t *testing.T, w *Window) {
	t.Helper()
	w.Frame()
	w.DrawTo(&mockCanvas{})
}

// A Text bound to a signal must be measured again when the signal changes;
// otherwise the new string is drawn into the old, smaller bounds.
func TestSignalRelayout_TextContentSignalResizesText(t *testing.T) {
	content := state.NewSignal("Hi")
	label := primitives.Text("").ContentSignal(content).FontSize(14)

	a := New(WithWindowProvider(&mockWindowProvider{width: 400, height: 200, scale: 1.0}))
	w := a.Window()
	w.SetRoot(primitives.Box(label).Padding(10))
	w.HandleResize(400, 200)
	frameAndPaint(t, w)
	before := label.Bounds().Width()

	content.Set("Hello, world! A much longer line")
	frameAndPaint(t, w)

	if after := label.Bounds().Width(); after <= before {
		t.Fatalf("text width after the signal changed = %v, want more than %v", after, before)
	}
}

// A Box whose direction is bound to a signal must lay its children out
// again when the direction flips.
func TestSignalRelayout_BoxDirectionSignalRelaysChildren(t *testing.T) {
	dir := state.NewSignal(primitives.DirectionVertical)
	first := primitives.Box().Width(50).Height(20)
	second := primitives.Box().Width(50).Height(20)

	a := New(WithWindowProvider(&mockWindowProvider{width: 400, height: 200, scale: 1.0}))
	w := a.Window()
	w.SetRoot(primitives.Box(primitives.Box(first, second).DirectionSignal(dir)))
	w.HandleResize(400, 200)
	frameAndPaint(t, w)
	if got := second.Bounds().Min; got != geometry.Pt(0, 20) {
		t.Fatalf("vertical: second child at %v, want (0, 20)", got)
	}

	dir.Set(primitives.DirectionHorizontal)
	frameAndPaint(t, w)

	if got := second.Bounds().Min; got != geometry.Pt(50, 0) {
		t.Fatalf("after switching to horizontal: second child at %v, want (50, 0)", got)
	}
}
