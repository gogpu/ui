package app_test

import (
	"testing"

	"github.com/gogpu/ui/app"
	"github.com/gogpu/ui/core/toolbar"
	"github.com/gogpu/ui/event"
	"github.com/gogpu/ui/geometry"
	"github.com/gogpu/ui/icon"
	"github.com/gogpu/ui/primitives"
	"github.com/gogpu/ui/uitest"
)

func TestToolbarHoverClearsAtSharedEdge(t *testing.T) {
	for _, tc := range []struct {
		name  string
		stack func(*toolbar.Widget, *primitives.BoxWidget) *primitives.BoxWidget
		step  geometry.Point
	}{
		{
			name: "bottom",
			stack: func(bar *toolbar.Widget, body *primitives.BoxWidget) *primitives.BoxWidget {
				return primitives.VBox(bar, primitives.Expanded(body))
			},
			step: geometry.Pt(0, 1),
		},
		{
			name: "right",
			stack: func(bar *toolbar.Widget, body *primitives.BoxWidget) *primitives.BoxWidget {
				return primitives.HBox(primitives.VBox(bar).Width(32), primitives.Expanded(body))
			},
			step: geometry.Pt(1, 0),
		},
	} {
		t.Run(tc.name, func(t *testing.T) {
			bar := toolbar.New(toolbar.Items(toolbar.IconButton("Settings", icon.ExpSettings, func() {})),
				toolbar.Height(42), toolbar.ButtonSize(32), toolbar.ButtonHeight(32))
			body := primitives.Box()
			ui := app.New()
			ui.SetRoot(tc.stack(bar, body))
			win := ui.Window()
			win.HandleResize(200, 160)
			ui.Frame()
			win.DrawTo(&uitest.MockCanvas{})

			edge := bar.ItemBounds(0).Center().Add(bar.ScreenOrigin())
			if tc.step.X != 0 {
				edge.X = bar.ScreenBounds().Max.X
			} else {
				edge.Y = bar.ScreenBounds().Max.Y
			}
			move := func(p geometry.Point) {
				win.HandleEvent(event.NewMouseEvent(event.MouseMove, event.ButtonNone, 0, p, p, event.ModNone))
			}
			highlighted := func() bool { return len(uitest.DrawWidget(bar).RoundRects) != 0 }
			move(edge.Sub(tc.step))
			if !highlighted() {
				t.Fatal("toolbar button did not highlight before crossing the edge")
			}
			move(edge)
			if win.HoveredWidget() != body {
				t.Fatalf("hover target at the edge = %T, want the body", win.HoveredWidget())
			}
			if highlighted() {
				t.Error("toolbar button is highlighted at the shared edge")
			}
			move(edge.Add(tc.step))
			if highlighted() {
				t.Error("toolbar button is still highlighted past the shared edge")
			}
		})
	}
}
