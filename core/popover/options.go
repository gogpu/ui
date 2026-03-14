package popover

import (
	"time"

	"github.com/gogpu/ui/state"
	"github.com/gogpu/ui/widget"
)

// Option configures a Popover or Tooltip during construction.
type Option func(*config)

// config holds the popover/tooltip configuration set at construction time.
type config struct {
	trigger           widget.Widget
	content           widget.Widget
	tooltipText       string
	placement         Placement
	gap               float32
	delay             time.Duration
	dismissOnClickOut bool
	painter           Painter
	visibleSignal     state.Signal[bool]
	disabled          bool
	disabledFn        func() bool
	onShow            func()
	onHide            func()
	maxWidth          float32
	tooltipPaddingH   float32
	tooltipPaddingV   float32
	tooltipFontSize   float32
	contentWidth      float32
	contentHeight     float32
}

// defaultConfig returns a config with sensible defaults.
func defaultConfig() config {
	return config{
		placement:         Bottom,
		gap:               defaultGap,
		delay:             defaultTooltipDelay,
		dismissOnClickOut: true,
		maxWidth:          tooltipMaxWidth,
		tooltipPaddingH:   tooltipPadH,
		tooltipPaddingV:   tooltipPadV,
		tooltipFontSize:   tooltipFontSz,
	}
}

// ResolvedDisabled returns the current disabled state, preferring the
// dynamic function over the static bool.
func (c *config) ResolvedDisabled() bool {
	if c.disabledFn != nil {
		return c.disabledFn()
	}
	return c.disabled
}

// TriggerWidget sets the widget that anchors the popover/tooltip.
func TriggerWidget(w widget.Widget) Option {
	return func(c *config) {
		c.trigger = w
	}
}

// Content sets the widget displayed inside the popover.
func Content(w widget.Widget) Option {
	return func(c *config) {
		c.content = w
	}
}

// ContentSize sets a fixed size for the popover content area.
// If not set, the content widget's natural size is used.
func ContentSize(width, height float32) Option {
	return func(c *config) {
		c.contentWidth = width
		c.contentHeight = height
	}
}

// TooltipText sets the text displayed in a tooltip.
func TooltipText(text string) Option {
	return func(c *config) {
		c.tooltipText = text
	}
}

// PlacementOpt sets the preferred placement relative to the trigger.
func PlacementOpt(p Placement) Option {
	return func(c *config) {
		c.placement = p
	}
}

// Gap sets the spacing between the trigger and the overlay in logical pixels.
// Defaults to 4.
func Gap(g float32) Option {
	return func(c *config) {
		c.gap = g
	}
}

// Delay sets the hover delay before showing a tooltip. Defaults to 500ms.
// Only applies to [Tooltip], not [Popover].
func Delay(d time.Duration) Option {
	return func(c *config) {
		c.delay = d
	}
}

// DismissOnClickOutside controls whether clicking outside the popover
// content dismisses it. Defaults to true.
func DismissOnClickOutside(dismiss bool) Option {
	return func(c *config) {
		c.dismissOnClickOut = dismiss
	}
}

// PainterOpt sets the painter used to render the popover or tooltip.
func PainterOpt(p Painter) Option {
	return func(c *config) {
		c.painter = p
	}
}

// VisibleSignal binds the visible state to a reactive signal for two-way
// data binding. When the signal is set externally, the popover shows/hides.
func VisibleSignal(sig state.Signal[bool]) Option {
	return func(c *config) {
		c.visibleSignal = sig
	}
}

// Disabled sets the popover's disabled state. Disabled popovers do not open.
func Disabled(d bool) Option {
	return func(c *config) {
		c.disabled = d
	}
}

// DisabledFn sets a dynamic function that determines whether the popover
// is disabled. When set, this takes precedence over the static value.
func DisabledFn(fn func() bool) Option {
	return func(c *config) {
		c.disabledFn = fn
	}
}

// OnShow sets a callback invoked when the popover or tooltip becomes visible.
func OnShow(fn func()) Option {
	return func(c *config) {
		c.onShow = fn
	}
}

// OnHide sets a callback invoked when the popover or tooltip is hidden.
func OnHide(fn func()) Option {
	return func(c *config) {
		c.onHide = fn
	}
}

// MaxWidth sets the maximum width for tooltip text wrapping.
// Defaults to 300 logical pixels.
func MaxWidth(w float32) Option {
	return func(c *config) {
		c.maxWidth = w
	}
}
