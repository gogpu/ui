package textfield

import (
	"github.com/gogpu/ui/geometry"
	"github.com/gogpu/ui/widget"
)

// Painter draws the visual representation of a text field.
// Each design system (Material 3, Fluent, Cupertino) provides its own
// Painter implementation to render the text field in its visual style.
//
// If no Painter is set, the text field uses [DefaultPainter].
type Painter interface {
	PaintTextField(canvas widget.Canvas, state *PaintState)
}

// PaintState provides the current text field state to the painter.
//
// Pre-computed geometry fields (DisplayText, ContentRect, CursorRect,
// SelectionRect) are calculated by the widget using [LayoutMetrics] from
// the painter. Painters should use these fields for positioning rather than
// computing their own cursor/selection coordinates.
type PaintState struct {
	// Legacy fields (kept for backward compatibility).
	Text        string
	Placeholder string
	Focused     bool
	Hovered     bool
	Disabled    bool
	HasError    bool
	ErrorMsg    string
	CursorPos   int
	SelectStart int
	SelectEnd   int
	InputType   InputType
	Bounds      geometry.Rect

	// Pre-computed fields (widget computes, painter draws).

	// DisplayText is the text to render, already masked if password.
	DisplayText string

	// ContentRect is the inner text area (bounds minus theme padding).
	ContentRect geometry.Rect

	// CursorRect is the cursor line rectangle (zero if no cursor).
	CursorRect geometry.Rect

	// SelectionRect is the selection highlight rectangle (zero if no selection).
	SelectionRect geometry.Rect

	// ShowCursor is true when the cursor should be drawn (focused, no selection, enabled).
	ShowCursor bool

	// ShowSelection is true when a selection highlight should be drawn.
	ShowSelection bool

	// FontSize is the font size to use for text rendering.
	FontSize float32

	// ColorScheme provides theme-derived colors. Zero value means use built-in defaults.
	ColorScheme TextFieldColorScheme
}

// LayoutMetrics allows theme painters to provide spatial metrics used by the
// widget to compute pre-computed PaintState fields (ContentRect, CursorRect, etc.).
//
// Painters that implement this interface provide custom padding, font size, etc.
// Painters that do not implement it get default values from [DefaultPainter].
type LayoutMetrics interface {
	// ContentPadding returns horizontal and vertical padding inside the text field.
	ContentPadding() (horizontal, vertical float32)

	// TextFieldFontSize returns the font size for the text content.
	TextFieldFontSize() float32

	// TextFieldCursorWidth returns the cursor line width in pixels.
	TextFieldCursorWidth() float32

	// TextFieldCornerRadius returns the corner radius for the text field border.
	TextFieldCornerRadius() float32
}

// TextFieldColorScheme provides theme-derived colors for text field painting.
// Zero value means the painter should use its built-in defaults.
type TextFieldColorScheme struct {
	Background  widget.Color
	Border      widget.Color
	FocusBorder widget.Color
	ErrorBorder widget.Color
	TextColor   widget.Color
	Placeholder widget.Color
	CursorColor widget.Color
	DisabledBg  widget.Color
	DisabledFg  widget.Color
	SelectionBg widget.Color
	ErrorText   widget.Color
}

// DefaultPainter provides a minimal fallback painter with no design system styling.
// It draws a simple outlined text field suitable for testing and prototyping.
//
// DefaultPainter also implements [LayoutMetrics], providing the default spatial
// values used when a painter does not implement that interface.
type DefaultPainter struct{}

// ContentPadding returns the default horizontal and vertical padding.
func (DefaultPainter) ContentPadding() (float32, float32) {
	return contentPaddingH, contentPaddingV
}

// TextFieldFontSize returns the default font size.
func (DefaultPainter) TextFieldFontSize() float32 { return defaultFontSize }

// TextFieldCursorWidth returns the default cursor width.
func (DefaultPainter) TextFieldCursorWidth() float32 { return cursorWidth }

// TextFieldCornerRadius returns the default corner radius.
func (DefaultPainter) TextFieldCornerRadius() float32 { return defaultCornerRadius }

// PaintTextField renders a minimal text field with gray outline.
// It uses the pre-computed fields from PaintState for cursor/selection positioning.
func (p DefaultPainter) PaintTextField(canvas widget.Canvas, st *PaintState) {
	if st.Bounds.IsEmpty() {
		return
	}

	colors := st.ColorScheme
	if colors == (TextFieldColorScheme{}) {
		colors = defaultColorScheme
	}
	fontSize := st.FontSize
	if fontSize <= 0 {
		fontSize = defaultFontSize
	}

	paintBackground(canvas, st, colors)
	paintBorder(canvas, st, colors)
	paintContent(canvas, st, colors, fontSize)
	paintCursorFromState(canvas, st, colors)
}

// paintBackground fills the text field background.
func paintBackground(canvas widget.Canvas, st *PaintState, colors TextFieldColorScheme) {
	bg := colors.Background
	if st.Disabled {
		bg = colors.DisabledBg
	}
	canvas.DrawRoundRect(st.Bounds, bg, defaultCornerRadius)
}

// paintBorder draws the text field outline.
func paintBorder(canvas widget.Canvas, st *PaintState, colors TextFieldColorScheme) {
	borderColor := colors.Border
	strokeWidth := defaultBorderWidth

	switch {
	case st.Disabled:
		borderColor = colors.DisabledFg
	case st.HasError:
		borderColor = colors.ErrorBorder
	case st.Focused:
		borderColor = colors.FocusBorder
		strokeWidth = defaultFocusBorderWidth
	case st.Hovered:
		borderColor = colors.TextColor
	}

	canvas.StrokeRoundRect(st.Bounds, borderColor, defaultCornerRadius, strokeWidth)
}

// paintContent renders either the placeholder or the displayed text.
func paintContent(canvas widget.Canvas, st *PaintState, colors TextFieldColorScheme, fontSize float32) {
	canvas.PushClip(st.ContentRect)
	defer canvas.PopClip()

	if st.DisplayText == "" && !st.Focused {
		color := colors.Placeholder
		if st.Disabled {
			color = colors.DisabledFg
		}
		canvas.DrawText(st.Placeholder, st.ContentRect, fontSize, color, false, textAlignLeft)
		return
	}

	textColor := colors.TextColor
	if st.Disabled {
		textColor = colors.DisabledFg
	}

	if st.ShowSelection {
		canvas.DrawRect(st.SelectionRect, colors.SelectionBg)
	}

	canvas.DrawText(st.DisplayText, st.ContentRect, fontSize, textColor, false, textAlignLeft)
}

// paintCursorFromState draws the cursor using pre-computed CursorRect.
func paintCursorFromState(canvas widget.Canvas, st *PaintState, colors TextFieldColorScheme) {
	if !st.ShowCursor {
		return
	}

	top := geometry.Pt(st.CursorRect.Min.X, st.CursorRect.Min.Y)
	bottom := geometry.Pt(st.CursorRect.Min.X, st.CursorRect.Max.Y)
	canvas.DrawLine(top, bottom, colors.CursorColor, st.CursorRect.Width())
}

// maskText returns a string of dots with the given length.
func maskText(length int) string {
	runes := make([]rune, length)
	for i := range runes {
		runes[i] = passwordMaskChar
	}
	return string(runes)
}

// Painting constants.
const (
	defaultFontSize         float32 = 14
	defaultCornerRadius     float32 = 4
	defaultBorderWidth      float32 = 1
	defaultFocusBorderWidth float32 = 2
	contentPaddingH         float32 = 12
	contentPaddingV         float32 = 8
	textAlignLeft                   = widget.TextAlignLeft
	cursorWidth             float32 = 1.5
	passwordMaskChar                = '\u2022' // bullet character
)

// Compile-time checks.
var (
	_ LayoutMetrics = DefaultPainter{}
)

// defaultColorScheme is the color scheme for DefaultPainter.
var defaultColorScheme = TextFieldColorScheme{
	Background:  widget.ColorWhite,
	Border:      widget.RGBA(0.45, 0.45, 0.45, 1.0),
	FocusBorder: widget.Hex(0x6750A4),
	ErrorBorder: widget.Hex(0xB3261E),
	TextColor:   widget.RGBA(0.1, 0.1, 0.1, 1.0),
	Placeholder: widget.RGBA(0.45, 0.45, 0.45, 1.0),
	CursorColor: widget.Hex(0x6750A4),
	DisabledBg:  widget.RGBA(0.95, 0.95, 0.95, 1.0),
	DisabledFg:  widget.RGBA(0.12, 0.12, 0.13, 0.38),
	SelectionBg: widget.Hex(0x6750A4).WithAlpha(0.2),
	ErrorText:   widget.Hex(0xB3261E),
}
