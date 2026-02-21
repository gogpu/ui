package overlay_test

import (
	"testing"

	"github.com/gogpu/ui/event"
	"github.com/gogpu/ui/geometry"
	"github.com/gogpu/ui/overlay"
	"github.com/gogpu/ui/widget"
)

// stubOverlay is a minimal Overlay implementation for testing.
type stubOverlay struct {
	widget.WidgetBase
	dismissed bool
	modal     bool
	consumed  bool
}

func (s *stubOverlay) Layout(_ widget.Context, c geometry.Constraints) geometry.Size {
	return c.Constrain(geometry.Sz(100, 100))
}
func (s *stubOverlay) Draw(_ widget.Context, _ widget.Canvas) {}
func (s *stubOverlay) Event(_ widget.Context, _ event.Event) bool {
	return s.consumed
}
func (s *stubOverlay) Children() []widget.Widget { return nil }
func (s *stubOverlay) Dismiss()                  { s.dismissed = true }
func (s *stubOverlay) Modal() bool               { return s.modal }

func newStubOverlay() *stubOverlay {
	o := &stubOverlay{}
	o.SetVisible(true)
	o.SetEnabled(true)
	return o
}

func TestStackPushPop(t *testing.T) {
	changed := 0
	s := overlay.NewStack(func() { changed++ })

	if !s.IsEmpty() {
		t.Error("new stack should be empty")
	}
	if s.Len() != 0 {
		t.Errorf("new stack Len() = %d, want 0", s.Len())
	}

	o1 := newStubOverlay()
	o2 := newStubOverlay()

	s.Push(o1)
	if s.IsEmpty() {
		t.Error("stack should not be empty after push")
	}
	if s.Top() != o1 {
		t.Error("Top should return o1")
	}
	if changed != 1 {
		t.Errorf("onChange called %d times, want 1", changed)
	}

	s.Push(o2)
	if s.Top() != o2 {
		t.Error("Top should return o2")
	}
	if s.Len() != 2 {
		t.Errorf("Len() = %d, want 2", s.Len())
	}

	popped := s.Pop()
	if popped != o2 {
		t.Error("Pop should return o2")
	}
	if s.Top() != o1 {
		t.Error("Top should return o1 after pop")
	}

	popped = s.Pop()
	if popped != o1 {
		t.Error("Pop should return o1")
	}
	if !s.IsEmpty() {
		t.Error("stack should be empty after popping all")
	}
}

func TestStackPopEmpty(t *testing.T) {
	s := overlay.NewStack(nil)
	if s.Pop() != nil {
		t.Error("Pop on empty stack should return nil")
	}
}

func TestStackPushNil(t *testing.T) {
	s := overlay.NewStack(nil)
	s.Push(nil)
	if !s.IsEmpty() {
		t.Error("pushing nil should not add to stack")
	}
}

func TestStackRemove(t *testing.T) {
	s := overlay.NewStack(nil)
	o1 := newStubOverlay()
	o2 := newStubOverlay()
	o3 := newStubOverlay()

	s.Push(o1)
	s.Push(o2)
	s.Push(o3)

	// Removing o2 should also remove o3 (everything above).
	s.Remove(o2)
	if s.Len() != 1 {
		t.Errorf("Len() = %d, want 1 after removing o2 (and o3 above it)", s.Len())
	}
	if s.Top() != o1 {
		t.Error("Top should be o1 after removing o2 and o3")
	}
}

func TestStackRemoveNonExistent(t *testing.T) {
	s := overlay.NewStack(nil)
	o1 := newStubOverlay()
	o2 := newStubOverlay()

	s.Push(o1)
	s.Remove(o2) // o2 not in stack, should be a no-op
	if s.Len() != 1 {
		t.Errorf("Len() = %d, want 1", s.Len())
	}
}

func TestStackList(t *testing.T) {
	s := overlay.NewStack(nil)
	o1 := newStubOverlay()
	o2 := newStubOverlay()

	s.Push(o1)
	s.Push(o2)

	list := s.List()
	if len(list) != 2 {
		t.Fatalf("List() length = %d, want 2", len(list))
	}
	if list[0] != o1 || list[1] != o2 {
		t.Error("List should return overlays in push order (bottom to top)")
	}
}

func TestStackHandleEventModalBlocks(t *testing.T) {
	s := overlay.NewStack(nil)
	o := newStubOverlay()
	o.modal = true
	o.consumed = false // overlay does not consume, but it's modal

	s.Push(o)

	ctx := widget.NewContext()
	me := event.NewMouseEvent(event.MousePress, event.ButtonLeft, 0, geometry.Pt(500, 500), geometry.Pt(500, 500), 0)

	blocked := s.HandleEvent(ctx, me)
	if !blocked {
		t.Error("modal overlay should block events from reaching widget tree")
	}
}

func TestStackHandleEventNonModalPassthrough(t *testing.T) {
	s := overlay.NewStack(nil)
	o := newStubOverlay()
	o.modal = false
	o.consumed = false

	s.Push(o)

	ctx := widget.NewContext()
	me := event.NewMouseEvent(event.MousePress, event.ButtonLeft, 0, geometry.Pt(500, 500), geometry.Pt(500, 500), 0)

	blocked := s.HandleEvent(ctx, me)
	if blocked {
		t.Error("non-modal overlay that does not consume should not block events")
	}
}

func TestStackHandleEventEmpty(t *testing.T) {
	s := overlay.NewStack(nil)
	ctx := widget.NewContext()
	me := event.NewMouseEvent(event.MousePress, event.ButtonLeft, 0, geometry.Pt(0, 0), geometry.Pt(0, 0), 0)

	if s.HandleEvent(ctx, me) {
		t.Error("empty stack should not handle events")
	}
}

func TestStackTop(t *testing.T) {
	s := overlay.NewStack(nil)
	if s.Top() != nil {
		t.Error("Top on empty stack should return nil")
	}
}
