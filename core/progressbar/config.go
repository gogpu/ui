package progressbar

import (
	"github.com/gogpu/ui/state"
	"github.com/gogpu/ui/widget"
)

// config holds the progress bar's configuration, set at construction time via options.
type config struct {
	value                  float64
	valueFn                func() float64
	valueSignal            state.Signal[float64]
	readonlyValueSignal    state.ReadonlySignal[float64]
	height                 float32
	radius                 float32
	radiusSet              bool // true if radius was explicitly set via option
	showLabel              bool
	formatLabel            func(float64) string
	disabled               bool
	disabledFn             func() bool
	disabledSignal         state.Signal[bool]
	readonlyDisabledSignal state.ReadonlySignal[bool]
	colorScheme            ProgressBarColorScheme
	painter                Painter
}

// ResolvedValue returns the current progress value.
// Priority: ReadonlySignal > Signal > Fn > Static.
func (c *config) ResolvedValue() float64 {
	if c.readonlyValueSignal != nil {
		return c.readonlyValueSignal.Get()
	}
	if c.valueSignal != nil {
		return c.valueSignal.Get()
	}
	if c.valueFn != nil {
		return c.valueFn()
	}
	return c.value
}

// ResolvedDisabled returns the current disabled state.
// Priority: ReadonlySignal > Signal > Fn > Static.
func (c *config) ResolvedDisabled() bool {
	if c.readonlyDisabledSignal != nil {
		return c.readonlyDisabledSignal.Get()
	}
	if c.disabledSignal != nil {
		return c.disabledSignal.Get()
	}
	if c.disabledFn != nil {
		return c.disabledFn()
	}
	return c.disabled
}

// ProgressBarColorScheme provides theme-derived colors for progress bar painting.
// Zero value means the painter should use its built-in defaults.
type ProgressBarColorScheme struct {
	Bar           widget.Color // filled portion color
	Track         widget.Color // background track color
	Label         widget.Color // label text color
	DisabledBar   widget.Color // filled portion when disabled
	DisabledTrack widget.Color // background track when disabled
}
