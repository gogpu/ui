package widget

import (
	"testing"

	"github.com/gogpu/ui/event"
	"github.com/gogpu/ui/geometry"
)

// testWidget is a concrete implementation of Widget for testing.
type testWidget struct {
	WidgetBase
	layoutCalled bool
	drawCalled   bool
	eventCalled  bool
	consumeEvent bool
}

func newTestWidget() *testWidget {
	w := &testWidget{}
	w.SetVisible(true)
	w.SetEnabled(true)
	return w
}

func (w *testWidget) Layout(ctx Context, c geometry.Constraints) geometry.Size {
	w.layoutCalled = true
	preferred := geometry.Sz(100, 50)
	size := c.Constrain(preferred)
	w.SetBounds(geometry.FromPointSize(geometry.Pt(0, 0), size))
	return size
}

func (w *testWidget) Draw(ctx Context, canvas Canvas) {
	w.drawCalled = true
}

func (w *testWidget) Event(ctx Context, e event.Event) bool {
	w.eventCalled = true
	return w.consumeEvent
}

func TestWidget_Interface(t *testing.T) {
	// Verify testWidget implements Widget
	var _ Widget = (*testWidget)(nil)
}

func TestWidget_Layout(t *testing.T) {
	w := newTestWidget()
	ctx := NewContext()

	tests := []struct {
		name        string
		constraints geometry.Constraints
		wantSize    geometry.Size
	}{
		{
			name:        "unconstrained",
			constraints: geometry.Expand(),
			wantSize:    geometry.Sz(100, 50),
		},
		{
			name:        "tight constraint",
			constraints: geometry.Tight(geometry.Sz(80, 40)),
			wantSize:    geometry.Sz(80, 40),
		},
		{
			name:        "loose constraint",
			constraints: geometry.Loose(geometry.Sz(200, 100)),
			wantSize:    geometry.Sz(100, 50),
		},
		{
			name:        "min constraint",
			constraints: geometry.BoxConstraints(150, 200, 75, 100),
			wantSize:    geometry.Sz(150, 75),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			w.layoutCalled = false
			size := w.Layout(ctx, tt.constraints)
			if !w.layoutCalled {
				t.Error("Layout was not called")
			}
			if size != tt.wantSize {
				t.Errorf("Layout() = %v, want %v", size, tt.wantSize)
			}
		})
	}
}

func TestWidget_Draw(t *testing.T) {
	w := newTestWidget()
	ctx := NewContext()

	if w.drawCalled {
		t.Error("drawCalled should be false initially")
	}

	w.Draw(ctx, nil) // Canvas is nil since we're just testing the interface

	if !w.drawCalled {
		t.Error("Draw was not called")
	}
}

func TestWidget_Event(t *testing.T) {
	tests := []struct {
		name     string
		consume  bool
		wantBool bool
	}{
		{"event consumed", true, true},
		{"event not consumed", false, false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			w := newTestWidget()
			w.consumeEvent = tt.consume
			ctx := NewContext()

			e := event.NewMouseEvent(
				event.MouseMove,
				event.ButtonNone,
				0, // ButtonState
				geometry.Pt(50, 25),
				geometry.Pt(50, 25),
				event.ModNone,
			)

			result := w.Event(ctx, e)
			if !w.eventCalled {
				t.Error("Event was not called")
			}
			if result != tt.wantBool {
				t.Errorf("Event() = %v, want %v", result, tt.wantBool)
			}
		})
	}
}

func TestWidget_Children(t *testing.T) {
	parent := newTestWidget()

	// Initially no children
	if parent.Children() != nil {
		t.Error("expected nil children initially")
	}

	// Add children
	child1 := newTestWidget()
	child2 := newTestWidget()
	parent.AddChild(child1)
	parent.AddChild(child2)

	children := parent.Children()
	if len(children) != 2 {
		t.Errorf("len(Children()) = %d, want 2", len(children))
	}
}

// containerWidget is a container widget for testing hierarchies.
type containerWidget struct {
	WidgetBase
}

func newContainerWidget() *containerWidget {
	c := &containerWidget{}
	c.SetVisible(true)
	c.SetEnabled(true)
	return c
}

func (c *containerWidget) Layout(ctx Context, constraints geometry.Constraints) geometry.Size {
	// Layout children with same constraints
	var totalHeight float32
	var maxWidth float32
	children := c.Children()
	for _, child := range children {
		childSize := child.Layout(ctx, constraints)
		if childSize.Width > maxWidth {
			maxWidth = childSize.Width
		}
		totalHeight += childSize.Height
	}
	size := geometry.Sz(maxWidth, totalHeight)
	return constraints.Constrain(size)
}

func (c *containerWidget) Draw(ctx Context, canvas Canvas) {
	for _, child := range c.Children() {
		child.Draw(ctx, canvas)
	}
}

func (c *containerWidget) Event(ctx Context, e event.Event) bool {
	// Dispatch to children
	for _, child := range c.Children() {
		if child.Event(ctx, e) {
			return true
		}
	}
	return false
}

func TestContainerWidget(t *testing.T) {
	// Verify containerWidget implements Widget
	var _ Widget = (*containerWidget)(nil)

	container := newContainerWidget()
	child1 := newTestWidget()
	child2 := newTestWidget()

	container.AddChild(child1)
	container.AddChild(child2)

	ctx := NewContext()

	// Test layout
	constraints := geometry.Loose(geometry.Sz(200, 200))
	size := container.Layout(ctx, constraints)

	// Both children are 100x50, stacked vertically = 100x100
	if size.Width != 100 {
		t.Errorf("container width = %v, want 100", size.Width)
	}
	if size.Height != 100 {
		t.Errorf("container height = %v, want 100", size.Height)
	}

	// Verify children were laid out
	if !child1.layoutCalled || !child2.layoutCalled {
		t.Error("children should have been laid out")
	}

	// Test draw
	container.Draw(ctx, nil)
	if !child1.drawCalled || !child2.drawCalled {
		t.Error("children should have been drawn")
	}

	// Test event dispatch
	child1.consumeEvent = true
	e := event.NewMouseEvent(
		event.MousePress,
		event.ButtonLeft,
		event.ButtonStateLeft,
		geometry.Pt(50, 25),
		geometry.Pt(50, 25),
		event.ModNone,
	)

	if !container.Event(ctx, e) {
		t.Error("container should return true when child consumes event")
	}
	if !child1.eventCalled {
		t.Error("child1 should have received event")
	}
	if child2.eventCalled {
		t.Error("child2 should NOT have received event (child1 consumed it)")
	}
}

func TestLayoutFunc(t *testing.T) {
	var called bool
	f := LayoutFunc(func(_ Context, c geometry.Constraints) geometry.Size {
		called = true
		return c.Constrain(geometry.Sz(50, 25))
	})

	ctx := NewContext()
	size := f(ctx, geometry.Expand())

	if !called {
		t.Error("LayoutFunc was not called")
	}
	if size.Width != 50 || size.Height != 25 {
		t.Errorf("size = %v, want (50, 25)", size)
	}
}

func TestDrawFunc(t *testing.T) {
	var called bool
	f := DrawFunc(func(_ Context, _ Canvas) {
		called = true
	})

	ctx := NewContext()
	f(ctx, nil)

	if !called {
		t.Error("DrawFunc was not called")
	}
}

func TestEventFunc(t *testing.T) {
	var called bool
	f := EventFunc(func(_ Context, _ event.Event) bool {
		called = true
		return true
	})

	ctx := NewContext()
	e := event.NewMouseEvent(
		event.MouseMove,
		event.ButtonNone,
		0, // ButtonState
		geometry.Pt(0, 0),
		geometry.Pt(0, 0),
		event.ModNone,
	)

	result := f(ctx, e)
	if !called {
		t.Error("EventFunc was not called")
	}
	if !result {
		t.Error("EventFunc should return true")
	}
}

func BenchmarkWidget_Layout(b *testing.B) {
	w := newTestWidget()
	ctx := NewContext()
	constraints := geometry.Loose(geometry.Sz(200, 200))

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = w.Layout(ctx, constraints)
	}
}

func BenchmarkContainerWidget_Layout(b *testing.B) {
	container := newContainerWidget()
	for i := 0; i < 10; i++ {
		container.AddChild(newTestWidget())
	}
	ctx := NewContext()
	constraints := geometry.Loose(geometry.Sz(200, 500))

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = container.Layout(ctx, constraints)
	}
}
