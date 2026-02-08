package button

import "github.com/gogpu/ui/widget"

// Padding sets equal horizontal and vertical padding.
// Returns the widget for method chaining.
func (w *Widget) Padding(v float32) *Widget {
	w.paddingX = v
	w.paddingY = v
	return w
}

// PaddingXY sets horizontal and vertical padding separately.
// Returns the widget for method chaining.
func (w *Widget) PaddingXY(x, y float32) *Widget {
	w.paddingX = x
	w.paddingY = y
	return w
}

// SetBackground sets a custom background color, overriding the variant default.
// Returns the widget for method chaining.
func (w *Widget) SetBackground(c widget.Color) *Widget {
	w.cfg.background = &c
	return w
}

// SetRounded sets the corner radius for the button.
// Returns the widget for method chaining.
func (w *Widget) SetRounded(radius float32) *Widget {
	w.cfg.rounded = &radius
	return w
}

// MinWidth sets the minimum width for the button.
// Returns the widget for method chaining.
func (w *Widget) MinWidth(v float32) *Widget {
	w.minWidth = v
	return w
}

// MaxWidth sets the maximum width for the button.
// Returns the widget for method chaining.
func (w *Widget) MaxWidth(v float32) *Widget {
	w.maxWidth = v
	return w
}
