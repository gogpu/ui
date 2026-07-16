// Package app window_close_test.go tests Window.Close lifecycle correctness.
// Regression tests for fix #175 (animation pumper goroutine leak on close).
//
// Verifies that Close unmounts the root widget tree, unmounts all overlays,
// stops the animation pumper, and is idempotent (safe to call multiple times).
package app

import (
	"testing"

	"github.com/gogpu/ui/event"
	"github.com/gogpu/ui/geometry"
	"github.com/gogpu/ui/state"
	"github.com/gogpu/ui/theme"
	"github.com/gogpu/ui/widget"
)

// closeLifecycleWidget tracks Mount/Unmount for Window.Close tests.
type closeLifecycleWidget struct {
	widget.WidgetBase
	mounted      bool
	mountCount   int
	unmountCount int
}

func newCloseLifecycleWidget() *closeLifecycleWidget {
	w := &closeLifecycleWidget{}
	w.SetVisible(true)
	w.SetEnabled(true)
	return w
}

func (w *closeLifecycleWidget) Mount(_ widget.Context) {
	w.mounted = true
	w.mountCount++
}

func (w *closeLifecycleWidget) Unmount() {
	w.mounted = false
	w.unmountCount++
}

func (w *closeLifecycleWidget) Layout(_ widget.Context, c geometry.Constraints) geometry.Size {
	return c.Constrain(geometry.Sz(100, 50))
}

func (w *closeLifecycleWidget) Draw(_ widget.Context, _ widget.Canvas) {}

func (w *closeLifecycleWidget) Event(_ widget.Context, _ event.Event) bool { return false }

func (w *closeLifecycleWidget) Children() []widget.Widget { return nil }

// Compile-time check.
var _ widget.Lifecycle = (*closeLifecycleWidget)(nil)

func TestWindowClose_UnmountsRoot(t *testing.T) {
	sched := state.NewScheduler(func(_ []widget.Widget) {})
	win := newWindow(nil, nil, sched, theme.DefaultLight(), RenderModeHostManaged)

	root := newCloseLifecycleWidget()
	win.SetRoot(root)
	win.Frame()

	if !root.mounted {
		t.Fatal("root should be mounted after SetRoot + Frame")
	}

	win.Close()

	if root.mounted {
		t.Error("root should be unmounted after Close")
	}
	if root.unmountCount < 1 {
		t.Errorf("root unmountCount = %d, want >= 1", root.unmountCount)
	}
	if win.Root() != nil {
		t.Error("Root() should return nil after Close")
	}
}

func TestWindowClose_UnmountsOverlays(t *testing.T) {
	sched := state.NewScheduler(func(_ []widget.Widget) {})
	win := newWindow(nil, nil, sched, theme.DefaultLight(), RenderModeHostManaged)
	root := newMockWidget()
	win.SetRoot(root)
	win.Frame()

	// Push two overlays via production path.
	mgr := &windowOverlayManager{window: win}
	overlay1 := newCloseLifecycleWidget()
	overlay1.SetBounds(geometry.NewRect(0, 0, 200, 100))
	overlay2 := newCloseLifecycleWidget()
	overlay2.SetBounds(geometry.NewRect(0, 0, 200, 100))

	mgr.PushOverlay(overlay1, nil)
	mgr.PushOverlay(overlay2, nil)

	if !overlay1.mounted || !overlay2.mounted {
		t.Fatal("overlays should be mounted after push")
	}

	win.Close()

	if overlay1.mounted {
		t.Error("overlay1 should be unmounted after Close")
	}
	if overlay2.mounted {
		t.Error("overlay2 should be unmounted after Close")
	}
	if win.OverlayCount() != 0 {
		t.Errorf("OverlayCount() = %d, want 0 after Close", win.OverlayCount())
	}
}

func TestWindowClose_Idempotent(t *testing.T) {
	sched := state.NewScheduler(func(_ []widget.Widget) {})
	win := newWindow(nil, nil, sched, theme.DefaultLight(), RenderModeHostManaged)
	root := newCloseLifecycleWidget()
	win.SetRoot(root)
	win.Frame()

	// Close multiple times — should not panic.
	win.Close()
	win.Close()
	win.Close()

	if root.mounted {
		t.Error("root should remain unmounted after multiple Close calls")
	}
	// unmountCount should be 1 (second/third Close sees nil root).
	// SetRoot in first Close also calls UnmountTree, so we check >= 1.
	if root.unmountCount < 1 {
		t.Errorf("root unmountCount = %d, want >= 1", root.unmountCount)
	}
}

func TestWindowClose_NilRootNoPanic(t *testing.T) {
	sched := state.NewScheduler(func(_ []widget.Widget) {})
	win := newWindow(nil, nil, sched, theme.DefaultLight(), RenderModeHostManaged)

	// Close without setting a root — should not panic.
	win.Close()
}

func TestWindowClose_ClearsHoveredAndCaptured(t *testing.T) {
	sched := state.NewScheduler(func(_ []widget.Widget) {})
	win := newWindow(nil, nil, sched, theme.DefaultLight(), RenderModeHostManaged)
	root := newMockWidget()
	win.SetRoot(root)
	win.Frame()

	// Simulate hovered/captured state.
	win.hoveredWidget = root
	win.capturedWidget = root

	win.Close()

	if win.hoveredWidget != nil {
		t.Error("hoveredWidget should be nil after Close")
	}
	if win.capturedWidget != nil {
		t.Error("capturedWidget should be nil after Close")
	}
}

func TestWindowClose_StopsAnimPumper(t *testing.T) {
	wp := &mockWindowProvider{width: 400, height: 300, scale: 1.0}
	sched := state.NewScheduler(func(_ []widget.Widget) {})
	win := newWindow(wp, nil, sched, theme.DefaultLight(), RenderModeHostManaged)
	root := newMockWidget()
	win.SetRoot(root)
	win.Frame()

	// Start animation pumper.
	win.animToken = newAnimPumper(wp)
	if win.animToken == nil {
		t.Fatal("animToken should be set")
	}

	win.Close()

	if win.animToken != nil {
		t.Error("animToken should be nil after Close")
	}
}
