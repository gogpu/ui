package dialog

import (
	"github.com/gogpu/ui/geometry"
	"github.com/gogpu/ui/widget"
)

// Painter draws the visual representation of a dialog.
// Each design system (Material 3, Fluent, Cupertino) provides its own
// Painter implementation to render the dialog in its visual style.
//
// If no Painter is set, the dialog uses [DefaultPainter].
type Painter interface {
	PaintDialog(canvas widget.Canvas, state PaintState)
}

// PaintState provides the current dialog state to the painter.
type PaintState struct {
	Title       string
	HasContent  bool
	Actions     []Action
	Focused     bool
	Bounds      geometry.Rect // dialog surface bounds (not backdrop)
	ColorScheme DialogColorScheme

	// ActionRects holds pre-computed button positions (one per action).
	// The widget computes these using canvas.MeasureText for accurate layout.
	// Painters should draw action buttons at these positions rather than
	// computing button widths from character count estimates.
	ActionRects []geometry.Rect
}

// DialogColorScheme provides theme-derived colors for dialog painting.
// Zero value means the painter should use its built-in defaults.
type DialogColorScheme struct {
	Backdrop widget.Color // backdrop color (semi-transparent)
	Surface  widget.Color // dialog background
	Title    widget.Color // title text color
	Content  widget.Color // content text color
	Border   widget.Color // dialog border (optional)
	Shadow   widget.Color // shadow color
	ActionFg widget.Color // action button text color
	ActionBg widget.Color // primary action button background
}

// LayoutMetrics allows theme painters to provide spatial metrics used by the
// widget's layout calculations to compute dialog dimensions.
//
// Painters that implement this interface provide custom metrics.
// Painters that do not implement it get default values from [DefaultPainter].
type LayoutMetrics interface {
	// DialogPadding returns the internal padding of the dialog surface.
	DialogPadding() float32

	// DialogTitleHeight returns the height reserved for the title text.
	DialogTitleHeight() float32

	// DialogMaxWidth returns the default maximum width for the dialog.
	DialogMaxWidth() float32
}

// DefaultPainter provides a minimal fallback painter with no design system styling.
// It draws a simple dialog -- useful for testing and as a base reference.
//
// DefaultPainter also implements [LayoutMetrics], providing the default spatial
// values used when a painter does not implement that interface.
type DefaultPainter struct{}

// PaintDialog renders a minimal dialog with white surface and gray border.
// If state.ColorScheme is non-zero, its colors are used instead of built-in defaults.
func (p DefaultPainter) PaintDialog(canvas widget.Canvas, ps PaintState) {
	if ps.Bounds.IsEmpty() {
		return
	}

	hasScheme := ps.ColorScheme != (DialogColorScheme{})

	// Draw surface background.
	bg := defaultSurfaceColor
	if hasScheme {
		bg = ps.ColorScheme.Surface
	}
	canvas.DrawRoundRect(ps.Bounds, bg, defaultDialogRadius)

	// Draw border.
	borderColor := defaultBorderColor
	if hasScheme {
		borderColor = ps.ColorScheme.Border
	}
	canvas.StrokeRoundRect(ps.Bounds, borderColor, defaultDialogRadius, defaultBorderWidth)

	// Draw title.
	if ps.Title != "" {
		titleColor := defaultTitleColor
		if hasScheme {
			titleColor = ps.ColorScheme.Title
		}
		titleBounds := geometry.Rect{
			Min: geometry.Pt(ps.Bounds.Min.X+dialogPadding, ps.Bounds.Min.Y+dialogPadding),
			Max: geometry.Pt(ps.Bounds.Max.X-dialogPadding, ps.Bounds.Min.Y+dialogPadding+titleHeight),
		}
		canvas.DrawText(ps.Title, titleBounds, titleFontSize, titleColor, true, textAlignLeft)
	}

	// Draw action buttons (right-aligned).
	p.paintActions(canvas, ps, hasScheme)

	// Draw focus indicator.
	if ps.Focused {
		ringBounds := ps.Bounds.Expand(focusRingOffset)
		ringRadius := defaultDialogRadius + focusRingOffset
		canvas.StrokeRoundRect(ringBounds, defaultFocusRingColor, ringRadius, focusRingStrokeWidth)
	}
}

// paintActions renders the action buttons right-aligned at the bottom of the dialog.
func (p DefaultPainter) paintActions(canvas widget.Canvas, ps PaintState, hasScheme bool) {
	if len(ps.Actions) == 0 {
		return
	}

	actionFg := defaultActionFgColor
	if hasScheme {
		actionFg = ps.ColorScheme.ActionFg
	}

	// Use pre-computed ActionRects when available (ADR-034 Phase 4).
	if len(ps.ActionRects) == len(ps.Actions) {
		for i, action := range ps.Actions {
			canvas.DrawText(action.Label, ps.ActionRects[i], actionFontSize, actionFg, false, textAlignCenter)
		}
		return
	}

	// Legacy fallback: compute button layout from character count estimate.
	x := ps.Bounds.Max.X - dialogPadding
	y := ps.Bounds.Max.Y - dialogPadding - actionHeight

	for i := len(ps.Actions) - 1; i >= 0; i-- {
		label := ps.Actions[i].Label
		btnWidth := float32(len(label))*actionCharWidth + actionPaddingX*2
		x -= btnWidth

		btnBounds := geometry.NewRect(x, y, btnWidth, actionHeight)
		canvas.DrawText(label, btnBounds, actionFontSize, actionFg, false, textAlignCenter)

		x -= actionSpacing
	}
}

// DialogPadding returns the default dialog padding.
func (DefaultPainter) DialogPadding() float32 { return dialogPadding }

// DialogTitleHeight returns the default title height.
func (DefaultPainter) DialogTitleHeight() float32 { return titleHeight }

// DialogMaxWidth returns the default maximum width.
func (DefaultPainter) DialogMaxWidth() float32 { return defaultMaxWidth }

// resolveDialogLayoutMetrics returns the LayoutMetrics from the painter if it
// implements that interface, otherwise returns DefaultPainter metrics.
func resolveDialogLayoutMetrics(p Painter) LayoutMetrics {
	if lm, ok := p.(LayoutMetrics); ok {
		return lm
	}
	return DefaultPainter{}
}

// Compile-time checks.
var _ LayoutMetrics = DefaultPainter{}

// Default painting constants.
const (
	defaultDialogRadius  float32 = 16
	defaultBorderWidth   float32 = 1
	dialogPadding        float32 = 24
	titleHeight          float32 = 28
	titleFontSize        float32 = 18
	actionHeight         float32 = 36
	actionFontSize       float32 = 14
	actionCharWidth      float32 = 8
	actionPaddingX       float32 = 12
	actionSpacing        float32 = 8
	textAlignLeft                = widget.TextAlignLeft
	textAlignCenter              = widget.TextAlignCenter
	focusRingOffset      float32 = 2
	focusRingStrokeWidth float32 = 2
)

// Default colors for DefaultPainter.
var (
	defaultSurfaceColor   = widget.ColorWhite
	defaultBorderColor    = widget.RGBA(0.8, 0.8, 0.8, 1.0)
	defaultTitleColor     = widget.RGBA(0.1, 0.1, 0.1, 1.0)
	defaultActionFgColor  = widget.Hex(0x6750A4)
	defaultFocusRingColor = widget.Hex(0x6750A4).WithAlpha(0.7)
	defaultBackdropColor  = widget.RGBA(0, 0, 0, 0.5)
)
