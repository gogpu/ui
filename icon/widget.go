package icon

import (
	"github.com/gogpu/ui/a11y"
	"github.com/gogpu/ui/event"
	"github.com/gogpu/ui/geometry"
	"github.com/gogpu/ui/state"
	"github.com/gogpu/ui/widget"
)

// defaultIconSize is the default display size in logical pixels.
const defaultIconSize float32 = 24

// IconOption configures an [IconWidget].
type IconOption func(*iconConfig)

// iconConfig holds configuration for constructing an IconWidget.
type iconConfig struct {
	size        float32
	color       widget.Color
	colorSignal state.ReadonlySignal[widget.Color]
	label       string
}

// Size sets the display size of the icon in logical pixels.
// The icon is rendered as a square of side length s.
func Size(s float32) IconOption {
	return func(c *iconConfig) { c.size = s }
}

// Color sets the icon stroke color.
func Color(color widget.Color) IconOption {
	return func(c *iconConfig) { c.color = color }
}

// ColorSignal binds the icon color to a reactive signal.
// When set, the signal value takes precedence over [Color].
func ColorSignal(sig state.ReadonlySignal[widget.Color]) IconOption {
	return func(c *iconConfig) { c.colorSignal = sig }
}

// Label sets the accessibility label for the icon.
func Label(text string) IconOption {
	return func(c *iconConfig) { c.label = text }
}

// IconWidget is a display widget that renders a vector icon.
//
// IconWidget implements [widget.Widget] and [a11y.Accessible].
//
// Create an IconWidget with [NewIcon].
type IconWidget struct {
	widget.WidgetBase

	data        IconData
	size        float32
	color       widget.Color
	colorSignal state.ReadonlySignal[widget.Color]
	label       string
}

// NewIcon creates a new icon widget displaying the given icon data.
//
//	w := icon.NewIcon(icon.Check, icon.Size(32), icon.Color(widget.ColorGreen))
func NewIcon(data IconData, opts ...IconOption) *IconWidget {
	cfg := iconConfig{
		size:  defaultIconSize,
		color: widget.ColorBlack,
	}
	for _, opt := range opts {
		opt(&cfg)
	}

	w := &IconWidget{
		data:        data,
		size:        cfg.size,
		color:       cfg.color,
		colorSignal: cfg.colorSignal,
		label:       cfg.label,
	}
	w.SetVisible(true)
	w.SetEnabled(true)
	return w
}

// Data returns the icon data being displayed.
func (w *IconWidget) Data() IconData {
	return w.data
}

// IconSize returns the display size in logical pixels.
func (w *IconWidget) IconSize() float32 {
	return w.size
}

// IconColor returns the resolved icon color.
//
// Resolution priority: [ColorSignal] > [Color].
func (w *IconWidget) IconColor() widget.Color {
	if w.colorSignal != nil {
		return w.colorSignal.Get()
	}
	return w.color
}

// IconLabel returns the accessibility label.
func (w *IconWidget) IconLabel() string {
	return w.label
}

// --- widget.Widget interface ---

// Layout returns the icon's preferred size (a square of side [IconSize]).
func (w *IconWidget) Layout(_ widget.Context, constraints geometry.Constraints) geometry.Size {
	preferred := geometry.Sz(w.size, w.size)
	resultSize := constraints.Constrain(preferred)
	w.SetBounds(geometry.FromPointSize(w.Position(), resultSize))
	return resultSize
}

// Draw renders the icon to the canvas.
func (w *IconWidget) Draw(_ widget.Context, canvas widget.Canvas) {
	if !w.IsVisible() {
		return
	}
	bounds := w.Bounds()
	if bounds.IsEmpty() {
		return
	}
	Draw(canvas, w.data, bounds, w.IconColor())
}

// Event returns false. Icon widgets do not consume events.
func (w *IconWidget) Event(_ widget.Context, _ event.Event) bool {
	return false
}

// Children returns nil. Icon is a leaf widget.
func (w *IconWidget) Children() []widget.Widget {
	return nil
}

// --- widget.Lifecycle interface ---

// Mount creates signal bindings for push-based invalidation.
func (w *IconWidget) Mount(ctx widget.Context) {
	sched := ctx.Scheduler()
	if sched == nil {
		return
	}
	if w.colorSignal != nil {
		b := state.BindToScheduler(w.colorSignal, w, sched)
		w.AddBinding(b)
	}
}

// Unmount is called when the icon widget is removed from the widget tree.
func (w *IconWidget) Unmount() {
	// Bindings are cleaned up automatically by WidgetBase.CleanupBindings().
}

// --- a11y.Accessible interface ---

// AccessibilityRole returns [a11y.RoleImage].
func (w *IconWidget) AccessibilityRole() a11y.Role {
	return a11y.RoleImage
}

// AccessibilityLabel returns the icon's accessibility label.
// Falls back to the icon name if no label was set.
func (w *IconWidget) AccessibilityLabel() string {
	if w.label != "" {
		return w.label
	}
	return w.data.Name
}

// AccessibilityHint returns an empty string.
func (w *IconWidget) AccessibilityHint() string {
	return ""
}

// AccessibilityValue returns an empty string.
func (w *IconWidget) AccessibilityValue() string {
	return ""
}

// AccessibilityState returns the default accessibility state.
func (w *IconWidget) AccessibilityState() a11y.State {
	return a11y.State{
		Hidden: !w.IsVisible(),
	}
}

// AccessibilityActions returns nil. Icons have no actions.
func (w *IconWidget) AccessibilityActions() []a11y.Action {
	return nil
}

// Compile-time interface checks.
var (
	_ widget.Widget    = (*IconWidget)(nil)
	_ a11y.Accessible  = (*IconWidget)(nil)
	_ widget.Lifecycle = (*IconWidget)(nil)
)
