package radio

// groupConfig holds the group's configuration, set at construction time via options.
type groupConfig struct {
	onChange   func(value string)
	selected   string
	direction  Direction
	disabled   bool
	disabledFn func() bool
	a11yLabel  string
	painter    Painter
	items      []ItemDef
}

// ResolvedDisabled returns the current disabled state, preferring the
// dynamic function over the static bool.
func (c *groupConfig) ResolvedDisabled() bool {
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
