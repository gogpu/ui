package dialog

import (
	"github.com/gogpu/ui/state"
	"github.com/gogpu/ui/widget"
)

// Option configures a dialog during construction.
type Option func(*config)

// Title sets the dialog's static title text.
func Title(s string) Option {
	return func(c *config) {
		c.title = s
	}
}

// TitleFn sets a dynamic title function evaluated on each draw.
// When set, this takes precedence over the static title but not over
// a signal set via [TitleSignal].
func TitleFn(fn func() string) Option {
	return func(c *config) {
		c.titleFn = fn
	}
}

// TitleSignal binds the dialog's title to a reactive signal.
// When set, the signal value takes precedence over both [TitleFn] and [Title]
// but not over [TitleReadonlySignal].
func TitleSignal(sig state.Signal[string]) Option {
	return func(c *config) {
		c.titleSignal = sig
	}
}

// TitleReadonlySignal binds the dialog's title to a read-only signal.
// This is useful for computed signals created via [state.NewComputed].
// When set, this takes highest precedence over all other title sources.
func TitleReadonlySignal(sig state.ReadonlySignal[string]) Option {
	return func(c *config) {
		c.readonlyTitleSignal = sig
	}
}

// Content sets the widget displayed in the dialog's content area.
// The content widget is laid out between the title and action buttons.
func Content(w widget.Widget) Option {
	return func(c *config) {
		c.content = w
	}
}

// ContentOnly lets the content supply its own header and actions.
// The dialog still paints the surface and backdrop and manages dismissal.
func ContentOnly() Option {
	return func(c *config) { c.contentOnly = true }
}

// Actions sets the action buttons displayed at the bottom of the dialog.
// Actions are rendered right-aligned in the order provided.
func Actions(actions ...Action) Option {
	return func(c *config) {
		c.actions = actions
	}
}

// DismissibleOpt controls whether clicking the backdrop closes the dialog.
// Default is true.
func DismissibleOpt(v bool) Option {
	return func(c *config) {
		c.dismissible = v
	}
}

// EscapeToCloseOpt controls whether pressing Escape closes the dialog.
// Default is true.
func EscapeToCloseOpt(v bool) Option {
	return func(c *config) {
		c.escToClose = v
	}
}

// OnClose sets a callback invoked when the dialog is closed for any reason
// (action click, backdrop click, Escape key).
func OnClose(fn func()) Option {
	return func(c *config) {
		c.onClose = fn
	}
}

// MaxWidth sets the maximum width of the dialog surface in logical pixels.
// Default is 560.
func MaxWidth(v float32) Option {
	return func(c *config) {
		c.maxWidth = v
	}
}

// MaxHeight sets the maximum height of the dialog surface in logical pixels.
// Default is 0 (no limit; constrained to 90% of window height).
func MaxHeight(v float32) Option {
	return func(c *config) {
		c.maxHeight = v
	}
}

// PainterOpt sets the painter used to render the dialog.
// Each design system provides its own painter. If not set,
// [DefaultPainter] is used.
func PainterOpt(p Painter) Option {
	return func(c *config) {
		c.painter = p
	}
}
