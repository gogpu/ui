package titlebar

import "github.com/gogpu/ui/widget"

// Option configures a title bar during construction.
type Option func(*config)

// config holds the title bar's configuration, set at construction time via options.
type config struct {
	title   string
	leading []widget.Widget
	center  []widget.Widget
	height  float32
	painter Painter
	chrome  WindowChrome
	focused bool
}

// Title sets the title bar's window title text.
// The title is drawn centered when no center widgets are provided.
func Title(s string) Option {
	return func(c *config) {
		c.title = s
	}
}

// Leading sets widgets to display in the left-aligned zone.
func Leading(widgets ...widget.Widget) Option {
	return func(c *config) {
		c.leading = widgets
	}
}

// Center sets widgets to display in the center zone.
func Center(widgets ...widget.Widget) Option {
	return func(c *config) {
		c.center = widgets
	}
}

// Height sets the title bar height in logical pixels.
// The default height is 40 pixels.
func Height(h float32) Option {
	return func(c *config) {
		c.height = h
	}
}

// PainterOpt sets the painter used to render the title bar.
// Each design system provides its own painter. If not set,
// [DefaultPainter] is used.
func PainterOpt(p Painter) Option {
	return func(c *config) {
		c.painter = p
	}
}

// Chrome sets the window chrome interface for window management operations.
// When provided, the title bar can minimize, maximize, and close the window,
// and empty space becomes a drag region for moving the window.
//
// If not set, the title bar renders as a purely visual bar with no window
// management capabilities; control buttons are not displayed.
func Chrome(wc WindowChrome) Option {
	return func(c *config) {
		c.chrome = wc
	}
}

// Focused sets the initial focused state of the title bar.
// This affects the visual appearance (e.g., dimmed controls when unfocused).
func Focused(f bool) Option {
	return func(c *config) {
		c.focused = f
	}
}
