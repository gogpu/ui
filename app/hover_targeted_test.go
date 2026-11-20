package app

import (
	"testing"

	"github.com/gogpu/gg/scene"
	"github.com/gogpu/ui/core/scrollview"
	"github.com/gogpu/ui/core/splitview"
	"github.com/gogpu/ui/event"
	"github.com/gogpu/ui/geometry"
	internalRender "github.com/gogpu/ui/internal/render"
	"github.com/gogpu/ui/primitives"
	"github.com/gogpu/ui/widget"
)

// hoverFlag tracks hover the way core widgets do (button, checkbox,
// textfield): MouseEnter sets the flag, MouseLeave clears it.
type hoverFlag struct {
	widget.WidgetBase
	height  float32
	hovered bool
}

func newHoverFlag(height float32) *hoverFlag {
	f := &hoverFlag{height: height}
	f.SetVisible(true)
	f.SetEnabled(true)
	return f
}

func (f *hoverFlag) Layout(_ widget.Context, c geometry.Constraints) geometry.Size {
	w := c.MaxWidth
	if w <= 0 || w == geometry.Infinity {
		w = 100
	}
	return c.Constrain(geometry.Sz(w, f.height))
}

func (f *hoverFlag) Draw(_ widget.Context, _ widget.Canvas) {}

func (f *hoverFlag) Event(_ widget.Context, e event.Event) bool {
	if me, ok := e.(*event.MouseEvent); ok {
		switch me.MouseType {
		case event.MouseEnter:
			f.hovered = true
			return true
		case event.MouseLeave:
			f.hovered = false
			return true
		}
	}
	return false
}

func (f *hoverFlag) Children() []widget.Widget { return nil }

// newHoverWindow lays out and records one frame so ScreenBounds are stamped
// exactly as in a real frame.
func newHoverWindow(t *testing.T, width, height int, root widget.Widget) *Window {
	t.Helper()

	prev := widget.GetSceneRecorderFactory()
	widget.RegisterSceneRecorder(func(s widget.SceneCache, w, h int) (widget.Canvas, func()) {
		rec := internalRender.NewSceneCanvas(s.(*scene.Scene), w, h)
		return rec, rec.Close
	})
	t.Cleanup(func() { widget.RegisterSceneRecorder(prev) })

	a := New(WithWindowProvider(&mockWindowProvider{width: width, height: height, scale: 1.0}))
	win := a.Window()
	win.SetRoot(root)
	win.HandleResize(width, height)
	win.Frame()
	PaintBoundaryLayersWithContext(win.Root(), nil, win.Context())
	return win
}

func hoverMouse(win *Window, typ event.MouseEventType, pt geometry.Point) {
	win.HandleEvent(event.NewMouseEvent(typ, event.ButtonNone, 0, pt, pt, event.ModNone))
}

// assertHoverFollowsPointer checks that exactly the widgets under the
// pointer are hovered.
func assertHoverFollowsPointer(t *testing.T, win *Window, pt geometry.Point, flags ...*hoverFlag) {
	t.Helper()
	for i, f := range flags {
		under := f.ScreenBounds().Contains(pt)
		if f.hovered != under {
			t.Fatalf("pointer at %v: widget %d %v hovered=%v, want %v (hover tracker target: %T)",
				pt, i, f.ScreenBounds(), f.hovered, under, win.HoveredWidget())
		}
	}
}

// hoverForm builds a settings-style form: a padded card of rows, each a
// fixed-width label followed by an expanded field, with the row itself
// wrapped in Expanded so the field sits a few containers deep.
func hoverForm(rows int) (widget.Widget, []*hoverFlag) {
	fields := make([]*hoverFlag, rows)
	rowWidgets := make([]widget.Widget, rows)
	for i := range rows {
		fields[i] = newHoverFlag(36)
		label := primitives.Box(primitives.Text("Label")).Width(128)
		pair := primitives.HBox(label, primitives.Expanded(primitives.Box(fields[i]).Height(36))).
			Gap(8).CrossAlign(primitives.CrossAxisCenter)
		rowWidgets[i] = primitives.HBox(primitives.Expanded(pair)).CrossAlign(primitives.CrossAxisCenter)
	}
	card := primitives.Box(rowWidgets...).Padding(16).Gap(8)
	return primitives.Box(card).PaddingTop(12).CrossAlign(primitives.CrossAxisStretch), fields
}

// The pointer moves down through the gap between the labels and the fields.
// There the row Box is the hover target; no field may light up, and none may
// stay lit once the pointer has moved on.
func TestHover_GapNextToNestedFieldsDoesNotHoverThem(t *testing.T) {
	root, fields := hoverForm(3)
	win := newHoverWindow(t, 600, 200, root)

	x := fields[0].ScreenBounds().Min.X - 2
	for y := float32(0); y < 200; y++ {
		pt := geometry.Pt(x, y)
		hoverMouse(win, event.MouseMove, pt)
		assertHoverFollowsPointer(t, win, pt, fields...)
	}
}

// Every pointer position, visited row by row, must leave exactly the field
// under the pointer hovered.
func TestHover_PointerScanKeepsHoverOnPointer(t *testing.T) {
	root, fields := hoverForm(3)
	win := newHoverWindow(t, 600, 200, root)

	for y := float32(0); y < 200; y++ {
		for x := float32(0); x < 600; x += 2 {
			pt := geometry.Pt(x, y)
			hoverMouse(win, event.MouseMove, pt)
			assertHoverFollowsPointer(t, win, pt, fields...)
		}
	}
}

func TestHover_ScrollViewBelowShortContentDoesNotHoverIt(t *testing.T) {
	content := newHoverFlag(36)
	win := newHoverWindow(t, 400, 300, scrollview.New(content))

	below := geometry.Pt(200, 150)
	hoverMouse(win, event.MouseMove, below)
	assertHoverFollowsPointer(t, win, below, content)

	on := content.ScreenBounds().Center()
	hoverMouse(win, event.MouseMove, on)
	assertHoverFollowsPointer(t, win, on, content)
}

func TestHover_SplitViewDividerDoesNotHoverPanes(t *testing.T) {
	first, second := newHoverFlag(300), newHoverFlag(300)
	win := newHoverWindow(t, 400, 300, splitview.New(splitview.First(first), splitview.Second(second)))

	fb, sb := first.ScreenBounds(), second.ScreenBounds()
	if sb.Min.X <= fb.Max.X {
		t.Fatalf("expected a divider between panes %v and %v", fb, sb)
	}
	divider := geometry.Pt((fb.Max.X+sb.Min.X)/2, 150)
	hoverMouse(win, event.MouseMove, divider)
	assertHoverFollowsPointer(t, win, divider, first, second)

	on := sb.Center()
	hoverMouse(win, event.MouseMove, on)
	assertHoverFollowsPointer(t, win, on, first, second)
}

func TestHover_PointerEnteringWindowHoversWidgetUnderIt(t *testing.T) {
	f := newHoverFlag(36)
	win := newHoverWindow(t, 400, 300, primitives.Box(f).Padding(20))

	pt := f.ScreenBounds().Center()
	hoverMouse(win, event.MouseEnter, pt)
	if !f.hovered || win.HoveredWidget() != widget.Widget(f) {
		t.Fatalf("after the pointer entered the window over the widget: hovered=%v, tracker target=%T; want the widget hovered and tracked",
			f.hovered, win.HoveredWidget())
	}

	hoverMouse(win, event.MouseLeave, pt)
	if f.hovered || win.HoveredWidget() != nil {
		t.Fatalf("after the pointer left the window: hovered=%v, tracker target=%T; want neither", f.hovered, win.HoveredWidget())
	}
}
