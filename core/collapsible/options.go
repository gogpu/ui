package collapsible

import (
	"time"

	"github.com/gogpu/ui/state"
	"github.com/gogpu/ui/widget"
)

// Option configures a collapsible section during construction.
type Option func(*config)

// Title sets the static header title text.
func Title(s string) Option {
	return func(c *config) {
		c.title = s
	}
}

// TitleFn sets a dynamic title function that is evaluated on each draw.
// When set, this takes precedence over the static title.
func TitleFn(fn func() string) Option {
	return func(c *config) {
		c.titleFn = fn
	}
}

// Content sets the child widget displayed when expanded.
func Content(w widget.Widget) Option {
	return func(c *config) {
		c.content = w
	}
}

// Expanded sets the initial expanded state.
// When true, the content is visible by default.
func Expanded(b bool) Option {
	return func(c *config) {
		c.expanded = b
	}
}

// ExpandedSignal binds the expanded state to a reactive signal.
// This is a TWO-WAY binding: the widget reads from the signal,
// and when the user toggles, the new state is written back.
// When set, the signal value takes precedence over [Expanded].
func ExpandedSignal(sig state.Signal[bool]) Option {
	return func(c *config) {
		c.expandedSignal = sig
	}
}

// ExpandedReadonlySignal binds the expanded state to a read-only signal.
// This is useful for computed signals created via [state.NewComputed].
// When set, this takes highest precedence over all other expanded sources.
func ExpandedReadonlySignal(sig state.ReadonlySignal[bool]) Option {
	return func(c *config) {
		c.readonlyExpandedSignal = sig
	}
}

// OnToggle sets the callback invoked when the section is toggled.
// The callback receives the new expanded state.
func OnToggle(fn func(expanded bool)) Option {
	return func(c *config) {
		c.onToggle = fn
	}
}

// HeaderHeight sets the height of the header bar in logical pixels.
// Default is 36.
func HeaderHeight(h float32) Option {
	return func(c *config) {
		c.headerHeight = h
	}
}

// HeaderColor sets the background color of the header bar.
// Default is a light gray.
func HeaderColor(color widget.Color) Option {
	return func(c *config) {
		c.headerColor = color
	}
}

// ArrowColor sets the color of the expand/collapse arrow indicator.
// Default is a dark gray.
func ArrowColor(color widget.Color) Option {
	return func(c *config) {
		c.arrowColor = color
	}
}

// Animated enables or disables smooth expand/collapse animation.
// Default is true.
func Animated(b bool) Option {
	return func(c *config) {
		c.animated = b
	}
}

// Duration sets the animation duration for expand/collapse.
// Default is 200ms.
func Duration(d time.Duration) Option {
	return func(c *config) {
		c.duration = d
	}
}

// PainterOpt sets the painter used to render the header.
// Each design system provides its own painter. If not set,
// [DefaultPainter] is used.
func PainterOpt(p Painter) Option {
	return func(c *config) {
		c.painter = p
	}
}
