package button

import (
	"github.com/gogpu/ui/event"
	"github.com/gogpu/ui/geometry"
	"github.com/gogpu/ui/widget"
)

// interactionState represents the current user interaction state.
type interactionState uint8

const (
	stateNormal  interactionState = iota
	stateHover                    // mouse is over the button
	statePressed                  // mouse button is held down
)

// Widget implements a clickable button with configurable appearance and behavior.
//
// A button is created with [New] using functional options:
//
//	btn := button.New(
//	    button.Text("Submit"),
//	    button.OnClick(handleSubmit),
//	    button.VariantOpt(button.Filled),
//	)
//
// Fluent styling methods may be chained after construction:
//
//	btn.Padding(16).Background(theme.Primary).Rounded(8)
type Widget struct {
	widget.WidgetBase
	cfg     config
	state   interactionState
	painter Painter

	// Styling overrides set via fluent methods.
	paddingX float32
	paddingY float32
	minWidth float32
	maxWidth float32
}

// New creates a new button Widget with the given options.
//
// The returned widget is visible, enabled, and focusable by default.
// Use options to configure text, click handler, variant, and size.
func New(opts ...Option) *Widget {
	w := &Widget{
		paddingX: defaultPaddingX,
		paddingY: defaultPaddingY,
		painter:  DefaultPainter{},
	}
	w.SetVisible(true)
	w.SetEnabled(true)

	// Default size is Medium.
	w.cfg.size = Medium

	for _, opt := range opts {
		opt(&w.cfg)
	}

	// Apply painter from config if set.
	if w.cfg.painter != nil {
		w.painter = w.cfg.painter
	}

	return w
}

// Default padding values.
const (
	defaultPaddingX float32 = 16
	defaultPaddingY float32 = 8
)

// IsFocusable reports whether the button can currently receive focus.
// A button is focusable when it is visible, enabled, and not disabled.
func (w *Widget) IsFocusable() bool {
	return w.IsVisible() && w.IsEnabled() && !w.cfg.ResolvedDisabled()
}

// Layout calculates the button's preferred size within the given constraints.
func (w *Widget) Layout(_ widget.Context, constraints geometry.Constraints) geometry.Size {
	height := sizeHeight(w.cfg.size)

	// Estimate text width: approximate at ~7px per character for medium font.
	text := w.cfg.ResolvedText()
	fontSize := sizeFontSize(w.cfg.size)
	textWidth := float32(len(text)) * fontSize * charWidthRatio

	totalWidth := textWidth + w.paddingX*2
	totalHeight := height

	// Apply min/max width overrides.
	if w.minWidth > 0 && totalWidth < w.minWidth {
		totalWidth = w.minWidth
	}
	if w.maxWidth > 0 && totalWidth > w.maxWidth {
		totalWidth = w.maxWidth
	}

	// Ensure height accounts for vertical padding.
	if totalHeight < w.paddingY*2 {
		totalHeight = w.paddingY * 2
	}

	preferred := geometry.Sz(totalWidth, totalHeight)
	return constraints.Constrain(preferred)
}

// charWidthRatio is an approximate ratio of character width to font size
// for text width estimation in layout.
const charWidthRatio float32 = 0.55

// Draw renders the button to the canvas.
func (w *Widget) Draw(ctx widget.Context, canvas widget.Canvas) {
	w.painter.PaintButton(canvas, PaintState{
		Text:       w.cfg.ResolvedText(),
		Variant:    w.cfg.variant,
		Size:       w.cfg.size,
		Hovered:    w.state == stateHover,
		Pressed:    w.state == statePressed,
		Focused:    w.IsFocused(),
		Disabled:   w.cfg.ResolvedDisabled(),
		Bounds:     w.Bounds(),
		Background: w.cfg.background,
		Radius:     w.cfg.rounded,
	})
}

// Event handles an input event and returns true if consumed.
func (w *Widget) Event(ctx widget.Context, e event.Event) bool {
	return handleEvent(w, ctx, e)
}

// Children returns nil because a button is a leaf widget.
func (w *Widget) Children() []widget.Widget {
	return nil
}

// Verify Widget implements required interfaces at compile time.
var (
	_ widget.Widget    = (*Widget)(nil)
	_ widget.Focusable = (*Widget)(nil)
)
