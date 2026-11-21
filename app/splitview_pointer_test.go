package app_test

import (
	"math"
	"testing"

	"github.com/gogpu/ui/app"
	"github.com/gogpu/ui/core/splitview"
	"github.com/gogpu/ui/event"
	"github.com/gogpu/ui/geometry"
	"github.com/gogpu/ui/gesture"
	"github.com/gogpu/ui/primitives"
	"github.com/gogpu/ui/uitest"
)

func TestSplitViewPointerDragStartsOnDivider(t *testing.T) {
	for _, tc := range []struct {
		name        string
		orientation splitview.Orientation
	}{
		{"horizontal", splitview.Horizontal},
		{"vertical", splitview.Vertical},
	} {
		t.Run(tc.name, func(t *testing.T) {
			const divider = 4
			inner := splitview.New(splitview.First(primitives.Box()), splitview.Second(primitives.Box()),
				splitview.OrientationOpt(tc.orientation), splitview.InitialRatio(0.5), splitview.DividerWidth(divider))
			outer := splitview.New(splitview.First(primitives.Box()), splitview.Second(inner),
				splitview.InitialRatio(0.3), splitview.DividerWidth(divider))
			ui := app.New()
			ui.SetRoot(outer)
			win := ui.Window()
			win.HandleResize(1000, 600)
			frame := func() {
				ui.Frame()
				win.DrawTo(&uitest.MockCanvas{})
			}
			frame()
			drag := func(from, middle, to geometry.Point) {
				send := func(kind gesture.PointerEventType, at geometry.Point, buttons event.ButtonState) {
					win.HandlePointerEvent(&gesture.PointerEvent{
						Base: event.NewBase(event.TypeMouse, 0), EventType: kind, PointerID: 1,
						PointerType: gesture.PointerTypeMouse, Position: at, GlobalPosition: at,
						Button: event.ButtonLeft, Buttons: buttons,
					})
				}
				send(gesture.PointerDown, from, event.ButtonStateLeft)
				send(gesture.PointerMove, middle, event.ButtonStateLeft)
				send(gesture.PointerMove, to, event.ButtonStateLeft)
				send(gesture.PointerUp, to, 0)
			}
			bounds := inner.ScreenBounds()
			from := bounds.Center()
			delta := geometry.Pt(-80, 0)
			to := geometry.Pt(from.X+delta.X, bounds.Min.Y-20)
			wantInner := float32(0.5) + delta.X/(bounds.Width()-divider)
			if tc.orientation == splitview.Vertical {
				delta = geometry.Pt(0, -80)
				to = geometry.Pt(bounds.Min.X-20, from.Y+delta.Y)
				wantInner = 0.5 + delta.Y/(bounds.Height()-divider)
			}
			drag(from, from.Add(geometry.Pt(delta.X/2, delta.Y/2)), to)
			if outer.Ratio() != 0.3 || math.Abs(float64(inner.Ratio()-wantInner)) > 0.001 {
				t.Fatalf("inner drag: outer ratio = %.3f, want 0.300; inner ratio = %.3f, want %.3f", outer.Ratio(), inner.Ratio(), wantInner)
			}

			frame()
			bounds = outer.ScreenBounds()
			from = geometry.Pt(bounds.Min.X+(bounds.Width()-divider)*outer.Ratio()+divider/2, bounds.Center().Y)
			to = geometry.Pt(from.X+80, bounds.Min.Y-20)
			innerRatio := inner.Ratio()
			drag(from, from.Add(geometry.Pt(40, 0)), to)
			want := 0.3 + 80/(bounds.Width()-divider)
			if math.Abs(float64(outer.Ratio()-want)) > 0.001 || inner.Ratio() != innerRatio {
				t.Fatalf("outer drag past bounds: outer ratio = %.3f, want %.3f; inner ratio = %.3f, want %.3f", outer.Ratio(), want, inner.Ratio(), innerRatio)
			}
		})
	}
}
