package primitives

import (
	"time"

	"github.com/gogpu/ui/event"
	"github.com/gogpu/ui/geometry"
	"github.com/gogpu/ui/widget"
)

// ThemeScopeWidget overrides the theme for a subtree of widgets.
//
// All descendants of a ThemeScopeWidget receive the scoped ThemeProvider
// instead of the application-level theme. This enables use cases like a
// dark dialog inside a light application, a branded section, or a
// high-contrast widget region.
//
// ThemeScopeWidget is transparent to layout: it delegates sizing entirely
// to its single child. If the scope has no child, it reports zero size.
//
// Nested ThemeScopeWidgets follow the "nearest wins" rule, matching the
// Flutter InheritedWidget pattern. The priority chain is:
//
//	Widget override > Nearest ThemeScope > App theme > Default
//
// Create a ThemeScopeWidget with the [ThemeScope] constructor.
//
// Example:
//
//	darkTheme := material3.NewDark(widget.Hex(0x6750A4))
//	scoped := primitives.ThemeScope(darkTheme,
//	    primitives.Text("Dark text"),
//	    primitives.Box(
//	        primitives.Text("Also dark"),
//	    ).Padding(8),
//	)
type ThemeScopeWidget struct {
	widget.WidgetBase

	theme widget.ThemeProvider
	child widget.Widget
}

// ThemeScope creates a new ThemeScopeWidget that overrides the theme for
// the given children.
//
// If multiple children are provided, they are wrapped in a Box container
// for vertical layout. A single child is used directly. If no children
// are provided, the scope has no child and reports zero size.
//
// The theme parameter must not be nil; passing nil causes ThemeScope to
// behave as if no theme override is applied (the parent context theme
// is used).
func ThemeScope(theme widget.ThemeProvider, children ...widget.Widget) *ThemeScopeWidget {
	ts := &ThemeScopeWidget{
		theme: theme,
	}
	ts.SetVisible(true)
	ts.SetEnabled(true)

	switch len(children) {
	case 0:
		// No child.
	case 1:
		ts.child = children[0]
	default:
		ts.child = Box(children...)
	}

	return ts
}

// Theme returns the ThemeProvider set on this scope.
func (ts *ThemeScopeWidget) Theme() widget.ThemeProvider {
	return ts.theme
}

// SetTheme changes the scoped theme.
func (ts *ThemeScopeWidget) SetTheme(theme widget.ThemeProvider) {
	ts.theme = theme
}

// Layout delegates sizing to the child widget, passing a context with
// the overridden theme.
func (ts *ThemeScopeWidget) Layout(ctx widget.Context, constraints geometry.Constraints) geometry.Size {
	scoped := ts.scopedContext(ctx)

	if ts.child == nil {
		size := constraints.Constrain(geometry.Sz(0, 0))
		ts.SetBounds(geometry.FromPointSize(ts.Position(), size))
		return size
	}

	size := ts.child.Layout(scoped, constraints)

	// Position the child at origin (no offset).
	ts.child.(interface{ SetBounds(geometry.Rect) }).SetBounds(
		geometry.FromPointSize(geometry.Pt(0, 0), size),
	)

	ts.SetBounds(geometry.FromPointSize(ts.Position(), size))
	return size
}

// Draw renders the child with the scoped theme context.
func (ts *ThemeScopeWidget) Draw(ctx widget.Context, canvas widget.Canvas) {
	if !ts.IsVisible() {
		return
	}

	if ts.child == nil {
		return
	}

	scoped := ts.scopedContext(ctx)
	canvas.PushTransform(ts.Bounds().Min)
	widget.StampScreenOrigin(ts.child, canvas)
	ts.child.Draw(scoped, canvas)
	canvas.PopTransform()
}

// Event dispatches events to the child with the scoped theme context.
func (ts *ThemeScopeWidget) Event(ctx widget.Context, e event.Event) bool {
	if !ts.IsVisible() || !ts.IsEnabled() {
		return false
	}

	if ts.child == nil {
		return false
	}

	scoped := ts.scopedContext(ctx)

	// Translate mouse events to local coordinates.
	if me, ok := e.(*event.MouseEvent); ok {
		local := *me
		local.Position = me.Position.Sub(ts.Bounds().Min)
		return ts.child.Event(scoped, &local)
	}

	return ts.child.Event(scoped, e)
}

// Children returns the child widget, or nil if none.
func (ts *ThemeScopeWidget) Children() []widget.Widget {
	if ts.child == nil {
		return nil
	}
	return []widget.Widget{ts.child}
}

// scopedContext returns a context wrapper that overrides ThemeProvider.
func (ts *ThemeScopeWidget) scopedContext(parent widget.Context) widget.Context {
	if ts.theme == nil {
		return parent
	}
	return &themeScopeContext{
		parent: parent,
		theme:  ts.theme,
	}
}

// themeScopeContext wraps a parent Context and overrides ThemeProvider.
//
// All other methods delegate to the parent context unchanged.
type themeScopeContext struct {
	parent widget.Context
	theme  widget.ThemeProvider
}

func (c *themeScopeContext) RequestFocus(w widget.Widget)       { c.parent.RequestFocus(w) }
func (c *themeScopeContext) ReleaseFocus(w widget.Widget)       { c.parent.ReleaseFocus(w) }
func (c *themeScopeContext) IsFocused(w widget.Widget) bool     { return c.parent.IsFocused(w) }
func (c *themeScopeContext) FocusedWidget() widget.Widget       { return c.parent.FocusedWidget() }
func (c *themeScopeContext) Now() time.Time                     { return c.parent.Now() }
func (c *themeScopeContext) DeltaTime() time.Duration           { return c.parent.DeltaTime() }
func (c *themeScopeContext) Invalidate()                        { c.parent.Invalidate() }
func (c *themeScopeContext) InvalidateRect(r geometry.Rect)     { c.parent.InvalidateRect(r) }
func (c *themeScopeContext) Cursor() widget.CursorType          { return c.parent.Cursor() }
func (c *themeScopeContext) SetCursor(cursor widget.CursorType) { c.parent.SetCursor(cursor) }
func (c *themeScopeContext) Scale() float32                     { return c.parent.Scale() }
func (c *themeScopeContext) OverlayManager() widget.OverlayManager {
	return c.parent.OverlayManager()
}
func (c *themeScopeContext) WindowSize() geometry.Size      { return c.parent.WindowSize() }
func (c *themeScopeContext) Scheduler() widget.SchedulerRef { return c.parent.Scheduler() }

// ThemeProvider returns the overridden theme instead of the parent's theme.
func (c *themeScopeContext) ThemeProvider() widget.ThemeProvider {
	return c.theme
}

// Compile-time interface checks.
var (
	_ widget.Widget  = (*ThemeScopeWidget)(nil)
	_ widget.Context = (*themeScopeContext)(nil)
)
