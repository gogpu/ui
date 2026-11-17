package app

import (
	"testing"

	"github.com/gogpu/ui/event"
	"github.com/gogpu/ui/geometry"
	"github.com/gogpu/ui/primitives"
	"github.com/gogpu/ui/widget"
)

// focusRecorder takes focus on a left press, like buttons and text fields,
// and records the widget-level focus events it receives.
type focusRecorder struct {
	widget.WidgetBase
	focusEvents []event.FocusEventType
}

func newFocusRecorder() *focusRecorder {
	w := &focusRecorder{}
	w.SetEnabled(true)
	w.SetVisible(true)
	return w
}

func (w *focusRecorder) Layout(_ widget.Context, c geometry.Constraints) geometry.Size {
	return c.Constrain(geometry.Sz(100, 30))
}

func (w *focusRecorder) Draw(_ widget.Context, _ widget.Canvas) {}

func (w *focusRecorder) Event(ctx widget.Context, e event.Event) bool {
	switch e := e.(type) {
	case *event.FocusEvent:
		w.focusEvents = append(w.focusEvents, e.FocusType)
	case *event.MouseEvent:
		if e.MouseType == event.MousePress && e.Button == event.ButtonLeft {
			ctx.RequestFocus(w)
			return true
		}
	}
	return false
}

func (w *focusRecorder) IsFocusable() bool { return true }

func TestWindow_PressBlursFocusedWidgetLikeBrowsers(t *testing.T) {
	center := func(r geometry.Rect) geometry.Point {
		return geometry.Pt(r.Min.X+r.Width()/2, r.Min.Y+r.Height()/2)
	}

	tests := []struct {
		name string
		// press returns where to press, given the focused field, another
		// focusable field, and a non-focusable label.
		press     func(field, other *focusRecorder, label widget.Widget) geometry.Point
		wantFocus func(field, other *focusRecorder) widget.Widget
		wantField []event.FocusEventType
	}{
		{
			name: "press on focused widget keeps focus",
			press: func(field, _ *focusRecorder, _ widget.Widget) geometry.Point {
				return center(field.ScreenBounds())
			},
			wantFocus: func(field, _ *focusRecorder) widget.Widget { return field },
			wantField: []event.FocusEventType{event.FocusGained},
		},
		{
			name: "press on non-focusable label blurs",
			press: func(_, _ *focusRecorder, label widget.Widget) geometry.Point {
				return center(label.(interface{ ScreenBounds() geometry.Rect }).ScreenBounds())
			},
			wantFocus: func(_, _ *focusRecorder) widget.Widget { return nil },
			wantField: []event.FocusEventType{event.FocusGained, event.FocusLost},
		},
		{
			name: "press on empty space blurs",
			press: func(_, _ *focusRecorder, _ widget.Widget) geometry.Point {
				return geometry.Pt(700, 500)
			},
			wantFocus: func(_, _ *focusRecorder) widget.Widget { return nil },
			wantField: []event.FocusEventType{event.FocusGained, event.FocusLost},
		},
		{
			name: "press on another focusable widget moves focus",
			press: func(_, other *focusRecorder, _ widget.Widget) geometry.Point {
				return center(other.ScreenBounds())
			},
			wantFocus: func(_, other *focusRecorder) widget.Widget { return other },
			wantField: []event.FocusEventType{event.FocusGained, event.FocusLost},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			field := newFocusRecorder()
			other := newFocusRecorder()
			label := primitives.Text("Label").FontSize(20)
			root := primitives.Box(field, label, other).Padding(32).Gap(12)

			w := New().Window()
			w.SetRoot(root)
			w.HandleResize(800, 600)
			w.Frame()
			w.DrawTo(&trackingCanvas{}) // records screen bounds for hit-testing

			press := func(pos geometry.Point) {
				w.HandleEvent(event.NewMouseEvent(
					event.MousePress, event.ButtonLeft, event.ButtonStateLeft, pos, pos, event.ModNone,
				))
				w.HandleEvent(event.NewMouseEvent(
					event.MouseRelease, event.ButtonLeft, 0, pos, pos, event.ModNone,
				))
			}

			press(center(field.ScreenBounds()))
			if w.Context().FocusedWidget() != field {
				t.Fatal("precondition: pressing the field should focus it")
			}

			press(tt.press(field, other, label))

			want := tt.wantFocus(field, other)
			if got := w.Context().FocusedWidget(); got != want {
				t.Errorf("focused widget = %v, want %v", got, want)
			}
			if got := field.IsFocused(); got != (want == widget.Widget(field)) {
				t.Errorf("field.IsFocused() = %v", got)
			}
			// The focus manager must agree so Tab continues from the right place.
			if got, _ := w.FocusManager().Focused().(widget.Widget); got != want {
				t.Errorf("focus manager focused = %v, want %v", got, want)
			}
			if !equalFocusEvents(field.focusEvents, tt.wantField) {
				t.Errorf("field focus events = %v, want %v", field.focusEvents, tt.wantField)
			}
		})
	}
}

func equalFocusEvents(a, b []event.FocusEventType) bool {
	if len(a) != len(b) {
		return false
	}
	for i := range a {
		if a[i] != b[i] {
			return false
		}
	}
	return true
}
