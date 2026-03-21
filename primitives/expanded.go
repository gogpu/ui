package primitives

import (
	"github.com/gogpu/ui/a11y"
	"github.com/gogpu/ui/event"
	"github.com/gogpu/ui/geometry"
	"github.com/gogpu/ui/widget"
)

// ExpandedWidget wraps a widget to indicate it should fill all remaining space
// in a [BoxWidget] layout (HBox/VBox). Without Expanded, children take only
// their intrinsic size. With Expanded, the child receives all space left
// after fixed-size siblings are measured.
//
// ExpandedWidget is transparent: it delegates Layout, Draw, Event, and Children
// entirely to its wrapped child. It serves only as a marker for the parent
// BoxWidget's two-pass layout algorithm.
//
// Usage:
//
//	primitives.HBox(
//	    leftPanel.Width(56),           // fixed 56px
//	    primitives.Expanded(editor),   // fills remaining
//	    rightPanel.Width(40),          // fixed 40px
//	)
//
// Multiple Expanded children in the same BoxWidget split the remaining space
// equally:
//
//	primitives.HBox(
//	    primitives.Expanded(left),     // 50% of remaining
//	    primitives.Expanded(right),    // 50% of remaining
//	)
type ExpandedWidget struct {
	widget.WidgetBase
	child widget.Widget
}

// Expanded wraps a widget to mark it as filling remaining space in a BoxWidget.
// The child must not be nil.
func Expanded(child widget.Widget) *ExpandedWidget {
	e := &ExpandedWidget{child: child}
	e.SetVisible(true)
	e.SetEnabled(true)
	return e
}

// IsExpanded returns true. This marker method is used by [BoxWidget] to detect
// expanded children during its two-pass layout.
func (e *ExpandedWidget) IsExpanded() bool { return true }

// Child returns the wrapped widget.
func (e *ExpandedWidget) Child() widget.Widget { return e.child }

// Layout delegates to the wrapped child and stores the resulting bounds.
func (e *ExpandedWidget) Layout(ctx widget.Context, constraints geometry.Constraints) geometry.Size {
	size := e.child.Layout(ctx, constraints)
	e.child.(interface{ SetBounds(geometry.Rect) }).SetBounds(
		geometry.FromPointSize(geometry.Point{}, size),
	)
	e.SetBounds(geometry.FromPointSize(e.Position(), size))
	return size
}

// Draw delegates to the wrapped child, applying a transform for positioning.
func (e *ExpandedWidget) Draw(ctx widget.Context, canvas widget.Canvas) {
	if !e.IsVisible() {
		return
	}
	bounds := e.Bounds()
	canvas.PushTransform(bounds.Min)
	widget.StampScreenOrigin(e.child, canvas)
	e.child.Draw(ctx, canvas)
	canvas.PopTransform()
}

// Event delegates to the wrapped child with coordinate translation for mouse
// and wheel events.
func (e *ExpandedWidget) Event(ctx widget.Context, ev event.Event) bool {
	if !e.IsVisible() || !e.IsEnabled() {
		return false
	}

	// Translate mouse events to local coordinates.
	if me, ok := ev.(*event.MouseEvent); ok {
		local := *me
		local.Position = me.Position.Sub(e.Bounds().Min)
		return e.child.Event(ctx, &local)
	}

	// Translate wheel events to local coordinates.
	if we, ok := ev.(*event.WheelEvent); ok {
		local := *we
		local.Position = we.Position.Sub(e.Bounds().Min)
		return e.child.Event(ctx, &local)
	}

	return e.child.Event(ctx, ev)
}

// Children returns the wrapped child as a single-element slice.
func (e *ExpandedWidget) Children() []widget.Widget {
	if e.child == nil {
		return nil
	}
	return []widget.Widget{e.child}
}

// AccessibilityRole delegates to the child if it implements [a11y.Accessible],
// otherwise returns [a11y.RoleNone].
func (e *ExpandedWidget) AccessibilityRole() a11y.Role {
	if acc, ok := e.child.(a11y.Accessible); ok {
		return acc.AccessibilityRole()
	}
	return a11y.RoleUnknown
}

// AccessibilityLabel delegates to the child if it implements [a11y.Accessible].
func (e *ExpandedWidget) AccessibilityLabel() string {
	if acc, ok := e.child.(a11y.Accessible); ok {
		return acc.AccessibilityLabel()
	}
	return ""
}

// AccessibilityHint delegates to the child if it implements [a11y.Accessible].
func (e *ExpandedWidget) AccessibilityHint() string {
	if acc, ok := e.child.(a11y.Accessible); ok {
		return acc.AccessibilityHint()
	}
	return ""
}

// AccessibilityValue delegates to the child if it implements [a11y.Accessible].
func (e *ExpandedWidget) AccessibilityValue() string {
	if acc, ok := e.child.(a11y.Accessible); ok {
		return acc.AccessibilityValue()
	}
	return ""
}

// AccessibilityState delegates to the child if it implements [a11y.Accessible].
func (e *ExpandedWidget) AccessibilityState() a11y.State {
	if acc, ok := e.child.(a11y.Accessible); ok {
		return acc.AccessibilityState()
	}
	return a11y.State{
		Disabled: !e.IsEnabled(),
		Hidden:   !e.IsVisible(),
	}
}

// AccessibilityActions delegates to the child if it implements [a11y.Accessible].
func (e *ExpandedWidget) AccessibilityActions() []a11y.Action {
	if acc, ok := e.child.(a11y.Accessible); ok {
		return acc.AccessibilityActions()
	}
	return nil
}

// isExpanded checks if a widget is wrapped in Expanded.
func isExpanded(w widget.Widget) bool {
	type expander interface{ IsExpanded() bool }
	if exp, ok := w.(expander); ok {
		return exp.IsExpanded()
	}
	return false
}

// Compile-time interface checks.
var (
	_ widget.Widget   = (*ExpandedWidget)(nil)
	_ a11y.Accessible = (*ExpandedWidget)(nil)
)
