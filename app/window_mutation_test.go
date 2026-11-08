package app

import (
	"testing"

	"github.com/gogpu/ui/event"
	"github.com/gogpu/ui/geometry"
)

// TestWindow_RemoveCapturedWidget_ReleasesCapture verifies that a widget which
// captured the pointer (ADR-031) but is then removed from the tree stops
// receiving events, and that the window drops its stale reference.
func TestWindow_RemoveCapturedWidget_ReleasesCapture(t *testing.T) {
	a := New()
	w := a.Window()

	root := newMockWidget()
	child := newMockWidget()
	root.AddChild(child)
	w.SetRoot(root) // mounts root and child

	// Capture the child as an active drag target.
	w.capturedWidget = child
	move := event.NewMouseEvent(
		event.MouseMove,
		event.ButtonLeft,
		event.ButtonStateLeft,
		geometry.Pt(1, 1),
		geometry.Pt(1, 1),
		event.ModNone,
	)
	w.HandleEvent(move)
	if !child.eventCalled {
		t.Fatal("captured widget should receive the move event")
	}
	child.eventCalled = false

	// Remove the child from the tree at runtime.
	if !root.RemoveChild(child) {
		t.Fatal("RemoveChild should return true for a mounted child")
	}
	if child.IsMounted() {
		t.Error("removed child should be unmounted")
	}

	// The next event must not reach the removed widget, and capture is cleared.
	w.HandleEvent(move)
	if w.capturedWidget != nil {
		t.Error("capture should be cleared for a widget removed from the tree")
	}
	if child.eventCalled {
		t.Error("a removed widget must not receive events")
	}
}
