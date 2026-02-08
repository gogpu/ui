package checkbox

import "github.com/gogpu/ui/widget"

// Padding sets the padding around the checkbox box.
// Returns the widget for method chaining.
func (w *Widget) Padding(v float32) *Widget {
	w.padding = v
	return w
}

// SetBackground sets a custom background color, overriding the default.
// Returns the widget for method chaining.
func (w *Widget) SetBackground(c widget.Color) *Widget {
	w.cfg.background = &c
	return w
}
