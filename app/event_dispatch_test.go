package app

import (
	"testing"

	"github.com/gogpu/ui/core/checkbox"
	"github.com/gogpu/ui/event"
	"github.com/gogpu/ui/geometry"
	"github.com/gogpu/ui/primitives"
)

// TestEventDispatchToCheckbox verifies that mouse events reach checkbox widgets
// through the full dispatch chain: Window → BoxWidget → checkbox.
func TestEventDispatchToCheckbox(t *testing.T) {
	toggled := false
	cb := checkbox.New(
		checkbox.LabelOpt("Test checkbox"),
		checkbox.OnToggle(func(checked bool) {
			toggled = true
			t.Logf("checkbox toggled: %v", checked)
		}),
	)

	root := primitives.Box(
		primitives.Text("Title").FontSize(28),
		cb,
	).Padding(32).Gap(12)

	wp := &mockWindowProvider{width: 800, height: 600, scale: 1.0}
	a := New(WithWindowProvider(wp))
	a.SetRoot(root)

	// Force layout so widgets have bounds.
	w := a.Window()
	w.needsLayout = true
	w.Frame()

	// Log the bounds of root and checkbox.
	t.Logf("root bounds: %v", root.Bounds())
	t.Logf("checkbox bounds: %v", cb.Bounds())

	// Compute a click position inside the checkbox bounds.
	cbBounds := cb.Bounds()
	clickX := cbBounds.Min.X + cbBounds.Width()/2
	clickY := cbBounds.Min.Y + cbBounds.Height()/2
	t.Logf("clicking at: (%.1f, %.1f)", clickX, clickY)

	// Simulate MousePress.
	pressEvent := event.NewMouseEvent(
		event.MousePress,
		event.ButtonLeft,
		event.ButtonStateLeft,
		geometry.Pt(clickX, clickY),
		geometry.Pt(clickX, clickY),
		event.ModNone,
	)
	w.HandleEvent(pressEvent)

	// Simulate MouseRelease at same position.
	releaseEvent := event.NewMouseEvent(
		event.MouseRelease,
		event.ButtonLeft,
		0,
		geometry.Pt(clickX, clickY),
		geometry.Pt(clickX, clickY),
		event.ModNone,
	)
	w.HandleEvent(releaseEvent)

	// Check if toggle was fired.
	if !toggled {
		t.Error("checkbox was NOT toggled — event dispatch chain is broken")
		t.Logf("Children of root:")
		for i, child := range root.Children() {
			if bw, ok := child.(interface{ Bounds() geometry.Rect }); ok {
				t.Logf("  [%d] %T bounds=%v", i, child, bw.Bounds())
			} else {
				t.Logf("  [%d] %T (no bounds)", i, child)
			}
		}
	}
}
