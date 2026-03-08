package radio

import (
	"github.com/gogpu/ui/state"
)

// groupConfig holds the group's configuration, set at construction time via options.
type groupConfig struct {
	onChange       func(value string)
	selected       string
	selectedSignal state.Signal[string]
	direction      Direction
	disabled       bool
	disabledFn     func() bool
	disabledSignal state.Signal[bool]
	a11yLabel      string
	painter        Painter
	items          []ItemDef
}

// ResolvedSelected returns the current selected value.
// Priority: Signal > Static.
func (c *groupConfig) ResolvedSelected() string {
	if c.selectedSignal != nil {
		return c.selectedSignal.Get()
	}
	return c.selected
}

// ResolvedDisabled returns the current disabled state.
// Priority: Signal > Fn > Static.
func (c *groupConfig) ResolvedDisabled() bool {
	if c.disabledSignal != nil {
		return c.disabledSignal.Get()
	}
	if c.disabledFn != nil {
		return c.disabledFn()
	}
	return c.disabled
}

// Direction controls the layout orientation of radio items within a group.
type Direction int

// Direction constants.
const (
	// Vertical lays out items from top to bottom.
	Vertical Direction = iota

	// Horizontal lays out items from left to right.
	Horizontal
)

// String returns a human-readable name for the direction.
func (d Direction) String() string {
	switch d {
	case Vertical:
		return directionVertical
	case Horizontal:
		return directionHorizontal
	default:
		return directionUnknown
	}
}

// String constants for Direction.String to satisfy goconst.
const (
	directionVertical   = "Vertical"
	directionHorizontal = "Horizontal"
	directionUnknown    = "Unknown"
)
