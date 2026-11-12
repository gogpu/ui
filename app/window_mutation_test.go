package app

import (
	"testing"

	"github.com/gogpu/ui/core/docking"
	"github.com/gogpu/ui/event"
	"github.com/gogpu/ui/geometry"
)

func newMoveEvent() *event.MouseEvent {
	return event.NewMouseEvent(
		event.MouseMove,
		event.ButtonLeft,
		event.ButtonStateLeft,
		geometry.Pt(1, 1),
		geometry.Pt(1, 1),
		event.ModNone,
	)
}

// TestWindow_CaptureKeptForDockedContent covers a live widget that entered the
// tree at runtime without being mounted: docking.Host.Dock after SetRoot does
// not run MountTree. Such a widget was never removed, so its capture must
// survive (a slider drag in a panel docked after SetRoot used to stick, because
// "not mounted" was read as "removed"; LEAD-023).
func TestWindow_CaptureKeptForDockedContent(t *testing.T) {
	a := New()
	w := a.Window()

	host := docking.NewHost()
	w.SetRoot(host)

	content := newMockWidget()
	host.Dock(docking.NewPanel(docking.PanelContent(content)), docking.Left)

	w.capturedWidget = content
	w.HandleEvent(newMoveEvent())

	if w.capturedWidget != content {
		t.Fatalf("capture dropped for live docked content: capturedWidget = %v", w.capturedWidget)
	}
	if !content.eventCalled {
		t.Error("captured docked content should receive the move event")
	}
}

// TestWindow_HoverKeptForNeverMountedWidget is the hover side of the same
// case: a live widget that was never mounted keeps its hover state, so it gets
// one MouseEnter and no repeated enters.
func TestWindow_HoverKeptForNeverMountedWidget(t *testing.T) {
	a := New()
	w := a.Window()
	w.SetRoot(newMockWidget())

	live := newMockWidget() // in no tree the window mounted, never removed
	w.hoveredWidget = live
	w.HandleEvent(newMoveEvent())

	if w.hoveredWidget != live {
		t.Errorf("hover dropped for a live, never-mounted widget: hoveredWidget = %v", w.hoveredWidget)
	}
}

// TestWindow_CaptureKeptAfterReaddToUnmountedContainer: a child removed from a
// container that was never mounted, then added back, is live again; its
// capture must survive. Removal unmounts the child, and re-adding to an
// unmounted container does not mount it, so only an explicit detached mark
// (cleared on re-attach) tells the two states apart.
func TestWindow_CaptureKeptAfterReaddToUnmountedContainer(t *testing.T) {
	a := New()
	w := a.Window()

	host := docking.NewHost()
	w.SetRoot(host)

	box := newMockWidget()
	box.SetSelf(box)
	child := newMockWidget()
	box.AddChild(child)
	host.Dock(docking.NewPanel(docking.PanelContent(box)), docking.Left)

	if !box.RemoveChild(child) {
		t.Fatal("RemoveChild should return true")
	}
	box.AddChild(child)

	w.capturedWidget = child
	w.HandleEvent(newMoveEvent())
	if w.capturedWidget != child {
		t.Fatalf("capture dropped for a re-added live child: capturedWidget = %v", w.capturedWidget)
	}
}

// TestWindow_SetRootDropsCaptureInOldTree: replacing the root unmounts the old
// tree, so a capture held inside it is dropped.
func TestWindow_SetRootDropsCaptureInOldTree(t *testing.T) {
	a := New()
	w := a.Window()

	oldRoot := newMockWidget()
	oldChild := newMockWidget()
	oldRoot.AddChild(oldChild)
	w.SetRoot(oldRoot)

	w.capturedWidget = oldChild
	w.SetRoot(newMockWidget())
	w.HandleEvent(newMoveEvent())

	if w.capturedWidget != nil {
		t.Errorf("capture in the replaced tree should be dropped, got %v", w.capturedWidget)
	}
	if oldChild.eventCalled {
		t.Error("a widget of the replaced tree must not receive events")
	}
}

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
