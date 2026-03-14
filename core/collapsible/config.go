package collapsible

import (
	"time"

	"github.com/gogpu/ui/state"
	"github.com/gogpu/ui/widget"
)

// config holds the collapsible section's configuration, set at construction time via options.
type config struct {
	title    string
	titleFn  func() string
	content  widget.Widget
	expanded bool

	expandedSignal         state.Signal[bool]
	readonlyExpandedSignal state.ReadonlySignal[bool]

	onToggle func(expanded bool)

	headerHeight float32
	headerColor  widget.Color
	arrowColor   widget.Color

	animated bool
	duration time.Duration

	painter Painter
}

// ResolvedTitle returns the current header title text.
// Priority: Fn > Static.
func (c *config) ResolvedTitle() string {
	if c.titleFn != nil {
		return c.titleFn()
	}
	return c.title
}

// ResolvedExpanded returns the current expanded state.
// Priority: ReadonlySignal > Signal > Static.
func (c *config) ResolvedExpanded() bool {
	if c.readonlyExpandedSignal != nil {
		return c.readonlyExpandedSignal.Get()
	}
	if c.expandedSignal != nil {
		return c.expandedSignal.Get()
	}
	return c.expanded
}
