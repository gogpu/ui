package overlay

import (
	"github.com/gogpu/ui/event"
	"github.com/gogpu/ui/geometry"
	"github.com/gogpu/ui/widget"
)

// scrimAlpha is the opacity of the semi-transparent backdrop for modal overlays.
const scrimAlpha = 0.32

// Container wraps overlay content with a full-window transparent backdrop.
// Clicks on the backdrop (outside the content) trigger dismissal. For modal
// overlays, a semi-transparent scrim is drawn behind the content.
//
// Container implements the [Overlay] interface.
type Container struct {
	widget.WidgetBase

	content    widget.Widget
	onDismiss  func()
	modal      bool
	windowSize geometry.Size
}

// ContainerOption configures a Container during construction.
type ContainerOption func(*Container)

// WithOnDismiss sets the callback invoked when the container is dismissed
// (click outside content or Escape key).
func WithOnDismiss(fn func()) ContainerOption {
	return func(c *Container) {
		c.onDismiss = fn
	}
}

// WithModal makes the container modal. A modal container draws a
// semi-transparent scrim and consumes all events not handled by the content.
func WithModal(modal bool) ContainerOption {
	return func(c *Container) {
		c.modal = modal
	}
}

// NewContainer creates a new overlay container wrapping the given content.
// The windowSize is used to size the backdrop to fill the entire window.
func NewContainer(content widget.Widget, windowSize geometry.Size, opts ...ContainerOption) *Container {
	c := &Container{
		content:    content,
		windowSize: windowSize,
	}
	c.SetVisible(true)
	c.SetEnabled(true)
	for _, opt := range opts {
		opt(c)
	}
	return c
}

// Layout fills the entire window and lays out the content unconstrained.
func (c *Container) Layout(ctx widget.Context, constraints geometry.Constraints) geometry.Size {
	size := constraints.Constrain(c.windowSize)
	c.SetBounds(geometry.NewRect(0, 0, size.Width, size.Height))

	if c.content != nil {
		// Content is laid out with loose constraints so it can take its natural size.
		c.content.Layout(ctx, geometry.Loose(size))
	}
	return size
}

// Draw renders the optional scrim and then the content.
func (c *Container) Draw(ctx widget.Context, canvas widget.Canvas) {
	if canvas == nil {
		return
	}
	if c.modal {
		scrim := widget.RGBA(0, 0, 0, scrimAlpha)
		canvas.DrawRect(c.Bounds(), scrim)
	}
	if c.content != nil {
		c.content.Draw(ctx, canvas)
	}
}

// Event dispatches events. Content gets first chance. If content does not
// consume the event, Escape key and mouse press outside content trigger
// dismissal. Modal containers consume all remaining events.
func (c *Container) Event(ctx widget.Context, e event.Event) bool {
	// Give content first chance at the event.
	if c.content != nil && c.content.Event(ctx, e) {
		return true
	}

	// Escape key dismisses.
	if ke, ok := e.(*event.KeyEvent); ok {
		if ke.KeyType == event.KeyPress && ke.Key == event.KeyEscape {
			c.Dismiss()
			return true
		}
	}

	// Mouse press outside content dismisses.
	if c.handleMouseDismiss(e) {
		return true
	}

	// Modal overlays consume all events to prevent interaction with content below.
	return c.modal
}

// handleMouseDismiss checks if a mouse press event occurred outside the
// content bounds and dismisses if so. Returns true if the event was handled.
func (c *Container) handleMouseDismiss(e event.Event) bool {
	me, ok := e.(*event.MouseEvent)
	if !ok || me.MouseType != event.MousePress {
		return false
	}

	if c.content == nil {
		c.Dismiss()
		return true
	}

	boundsGetter, ok := c.content.(interface{ Bounds() geometry.Rect })
	if !ok {
		return false
	}

	if !boundsGetter.Bounds().Contains(me.Position) {
		c.Dismiss()
		return true
	}
	return false
}

// Dismiss triggers the dismissal callback.
func (c *Container) Dismiss() {
	if c.onDismiss != nil {
		c.onDismiss()
	}
}

// Modal returns true if this container blocks interaction with content below.
func (c *Container) Modal() bool {
	return c.modal
}

// Children returns the content widget as the sole child.
func (c *Container) Children() []widget.Widget {
	if c.content == nil {
		return nil
	}
	return []widget.Widget{c.content}
}

// Content returns the wrapped content widget.
func (c *Container) Content() widget.Widget {
	return c.content
}

// SetWindowSize updates the backdrop size. Call this when the window is resized.
func (c *Container) SetWindowSize(size geometry.Size) {
	c.windowSize = size
}

// Compile-time interface checks.
var _ Overlay = (*Container)(nil)
