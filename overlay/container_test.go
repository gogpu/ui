package overlay_test

import (
	"testing"

	"github.com/gogpu/ui/event"
	"github.com/gogpu/ui/geometry"
	"github.com/gogpu/ui/overlay"
	"github.com/gogpu/ui/widget"
)

// mockContent is a widget that tracks events and has configurable bounds.
type mockContent struct {
	widget.WidgetBase
	consumeEvents bool
	lastEvent     event.Event
}

func (m *mockContent) Layout(_ widget.Context, c geometry.Constraints) geometry.Size {
	return c.Constrain(geometry.Sz(200, 100))
}
func (m *mockContent) Draw(_ widget.Context, _ widget.Canvas) {}
func (m *mockContent) Event(_ widget.Context, e event.Event) bool {
	m.lastEvent = e
	return m.consumeEvents
}
func (m *mockContent) Children() []widget.Widget { return nil }

func newMockContent(bounds geometry.Rect) *mockContent {
	m := &mockContent{}
	m.SetVisible(true)
	m.SetEnabled(true)
	m.SetBounds(bounds)
	return m
}

func TestContainerDismissOnEscape(t *testing.T) {
	dismissed := false
	content := newMockContent(geometry.NewRect(100, 100, 200, 100))
	c := overlay.NewContainer(content, geometry.Sz(800, 600),
		overlay.WithOnDismiss(func() { dismissed = true }),
	)

	ctx := widget.NewContext()
	esc := event.NewKeyEvent(event.KeyPress, event.KeyEscape, 0, 0)

	consumed := c.Event(ctx, esc)
	if !consumed {
		t.Error("Escape should be consumed")
	}
	if !dismissed {
		t.Error("Escape should trigger dismiss")
	}
}

func TestContainerDismissOnClickOutside(t *testing.T) {
	dismissed := false
	content := newMockContent(geometry.NewRect(100, 100, 200, 100))
	c := overlay.NewContainer(content, geometry.Sz(800, 600),
		overlay.WithOnDismiss(func() { dismissed = true }),
	)

	ctx := widget.NewContext()

	// Click outside content bounds.
	outsideClick := event.NewMouseEvent(
		event.MousePress, event.ButtonLeft, 0,
		geometry.Pt(50, 50), geometry.Pt(50, 50), 0,
	)
	consumed := c.Event(ctx, outsideClick)
	if !consumed {
		t.Error("click outside should be consumed")
	}
	if !dismissed {
		t.Error("click outside content should trigger dismiss")
	}
}

func TestContainerContentConsumesEvent(t *testing.T) {
	dismissed := false
	content := newMockContent(geometry.NewRect(100, 100, 200, 100))
	content.consumeEvents = true

	c := overlay.NewContainer(content, geometry.Sz(800, 600),
		overlay.WithOnDismiss(func() { dismissed = true }),
	)

	ctx := widget.NewContext()
	me := event.NewMouseEvent(
		event.MousePress, event.ButtonLeft, 0,
		geometry.Pt(150, 150), geometry.Pt(150, 150), 0,
	)
	consumed := c.Event(ctx, me)
	if !consumed {
		t.Error("event consumed by content should be reported as consumed")
	}
	if dismissed {
		t.Error("should not dismiss when content consumes the event")
	}
}

func TestContainerModalConsumesAll(t *testing.T) {
	content := newMockContent(geometry.NewRect(100, 100, 200, 100))
	c := overlay.NewContainer(content, geometry.Sz(800, 600),
		overlay.WithModal(true),
	)

	ctx := widget.NewContext()

	// A mouse move event that content does not consume.
	move := event.NewMouseEvent(
		event.MouseMove, event.ButtonNone, 0,
		geometry.Pt(500, 500), geometry.Pt(500, 500), 0,
	)
	consumed := c.Event(ctx, move)
	if !consumed {
		t.Error("modal container should consume events not handled by content")
	}
}

func TestContainerNonModalPassthrough(t *testing.T) {
	content := newMockContent(geometry.NewRect(100, 100, 200, 100))
	c := overlay.NewContainer(content, geometry.Sz(800, 600))

	ctx := widget.NewContext()

	// A mouse move that content does not consume and is not a click.
	move := event.NewMouseEvent(
		event.MouseMove, event.ButtonNone, 0,
		geometry.Pt(500, 500), geometry.Pt(500, 500), 0,
	)
	consumed := c.Event(ctx, move)
	if consumed {
		t.Error("non-modal container should not consume unhandled non-click events")
	}
}

func TestContainerLayout(t *testing.T) {
	content := newMockContent(geometry.NewRect(0, 0, 200, 100))
	c := overlay.NewContainer(content, geometry.Sz(800, 600))

	ctx := widget.NewContext()
	size := c.Layout(ctx, geometry.Tight(geometry.Sz(800, 600)))

	if size.Width != 800 || size.Height != 600 {
		t.Errorf("Layout size = %v, want (800, 600)", size)
	}
}

func TestContainerModal(t *testing.T) {
	c := overlay.NewContainer(nil, geometry.Sz(800, 600))
	if c.Modal() {
		t.Error("default container should not be modal")
	}

	c2 := overlay.NewContainer(nil, geometry.Sz(800, 600), overlay.WithModal(true))
	if !c2.Modal() {
		t.Error("container with WithModal(true) should be modal")
	}
}

func TestContainerChildren(t *testing.T) {
	c := overlay.NewContainer(nil, geometry.Sz(800, 600))
	if c.Children() != nil {
		t.Error("container with nil content should return nil children")
	}

	content := newMockContent(geometry.NewRect(0, 0, 100, 100))
	c2 := overlay.NewContainer(content, geometry.Sz(800, 600))
	children := c2.Children()
	if len(children) != 1 {
		t.Fatalf("Children length = %d, want 1", len(children))
	}
	if children[0] != content {
		t.Error("Children[0] should be the content widget")
	}
}

func TestContainerNilContent(t *testing.T) {
	dismissed := false
	c := overlay.NewContainer(nil, geometry.Sz(800, 600),
		overlay.WithOnDismiss(func() { dismissed = true }),
	)

	ctx := widget.NewContext()
	click := event.NewMouseEvent(
		event.MousePress, event.ButtonLeft, 0,
		geometry.Pt(100, 100), geometry.Pt(100, 100), 0,
	)
	consumed := c.Event(ctx, click)
	if !consumed {
		t.Error("click on container with nil content should be consumed (dismiss)")
	}
	if !dismissed {
		t.Error("click on container with nil content should trigger dismiss")
	}
}
