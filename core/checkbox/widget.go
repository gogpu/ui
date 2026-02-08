package checkbox

import (
	"github.com/gogpu/ui/event"
	"github.com/gogpu/ui/geometry"
	"github.com/gogpu/ui/widget"
)

// interactionState represents the current user interaction state.
type interactionState uint8

const (
	stateNormal  interactionState = iota
	stateHover                    // mouse is over the checkbox
	statePressed                  // mouse button is held down
)

// Widget implements a toggleable checkbox with configurable appearance and behavior.
//
// A checkbox is created with [New] using functional options:
//
//	cb := checkbox.New(
//	    checkbox.Label("Accept terms"),
//	    checkbox.OnToggle(handleToggle),
//	    checkbox.Checked(true),
//	)
//
// Fluent styling methods may be chained after construction:
//
//	cb.Padding(8).SetBackground(theme.Primary)
type Widget struct {
	widget.WidgetBase
	cfg     config
	state   interactionState
	painter Painter

	// Styling overrides set via fluent methods.
	padding float32
}

// New creates a new checkbox Widget with the given options.
//
// The returned widget is visible, enabled, and focusable by default.
// Use options to configure label, toggle handler, and checked state.
func New(opts ...Option) *Widget {
	w := &Widget{
		padding: defaultPadding,
		painter: DefaultPainter{},
	}
	w.SetVisible(true)
	w.SetEnabled(true)

	for _, opt := range opts {
		opt(&w.cfg)
	}

	// Apply painter from config if set.
	if w.cfg.painter != nil {
		w.painter = w.cfg.painter
	}

	return w
}

// Default padding value.
const defaultPadding float32 = 4

// IsFocusable reports whether the checkbox can currently receive focus.
// A checkbox is focusable when it is visible, enabled, and not disabled.
func (w *Widget) IsFocusable() bool {
	return w.IsVisible() && w.IsEnabled() && !w.cfg.ResolvedDisabled()
}

// Layout calculates the checkbox's preferred size within the given constraints.
func (w *Widget) Layout(_ widget.Context, constraints geometry.Constraints) geometry.Size {
	// Box size + optional label width.
	totalWidth := boxSize + w.padding*2
	totalHeight := boxSize + w.padding*2

	label := w.cfg.ResolvedLabel()
	if label != "" {
		textWidth := float32(len(label)) * defaultFontSize * charWidthRatio
		totalWidth += labelGap + textWidth
	}

	// Ensure minimum height.
	if totalHeight < minHeight {
		totalHeight = minHeight
	}

	preferred := geometry.Sz(totalWidth, totalHeight)
	return constraints.Constrain(preferred)
}

// charWidthRatio is an approximate ratio of character width to font size
// for text width estimation in layout.
const charWidthRatio float32 = 0.55

// minHeight is the minimum checkbox height for comfortable touch targets.
const minHeight float32 = 24

// Draw renders the checkbox to the canvas.
func (w *Widget) Draw(ctx widget.Context, canvas widget.Canvas) {
	w.painter.PaintCheckbox(canvas, PaintState{
		Label:         w.cfg.ResolvedLabel(),
		Checked:       w.cfg.ResolvedChecked(),
		Indeterminate: w.cfg.indeterminate,
		Hovered:       w.state == stateHover,
		Pressed:       w.state == statePressed,
		Focused:       w.IsFocused(),
		Disabled:      w.cfg.ResolvedDisabled(),
		Bounds:        w.Bounds(),
		Background:    w.cfg.background,
	})
}

// Event handles an input event and returns true if consumed.
func (w *Widget) Event(ctx widget.Context, e event.Event) bool {
	return handleEvent(w, ctx, e)
}

// Children returns nil because a checkbox is a leaf widget.
func (w *Widget) Children() []widget.Widget {
	return nil
}

// Verify Widget implements required interfaces at compile time.
var (
	_ widget.Widget    = (*Widget)(nil)
	_ widget.Focusable = (*Widget)(nil)
)
