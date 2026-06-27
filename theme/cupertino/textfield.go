package cupertino

import (
	"github.com/gogpu/ui/core/textfield"
	"github.com/gogpu/ui/geometry"
	"github.com/gogpu/ui/widget"
)

// TextFieldPainter renders text fields using Apple HIG design tokens.
// Cupertino text fields use a rounded rectangle with a light background
// and subtle border, following iOS text field conventions.
//
// TextFieldPainter implements [textfield.LayoutMetrics] to provide Cupertino
// spatial metrics (15px font, 12/10px padding, 2px cursor) used by the widget
// to compute pre-computed PaintState fields.
//
// If Theme is nil, TextFieldPainter falls back to the default system blue palette.
type TextFieldPainter struct {
	Theme *Theme // nil uses default system blue fallback
}

// ContentPadding returns the Cupertino horizontal and vertical padding.
func (TextFieldPainter) ContentPadding() (float32, float32) {
	return cupTFContentPaddingH, cupTFContentPaddingV
}

// TextFieldFontSize returns the Cupertino font size.
func (TextFieldPainter) TextFieldFontSize() float32 { return cupTFFontSize }

// TextFieldCursorWidth returns the Cupertino cursor width.
func (TextFieldPainter) TextFieldCursorWidth() float32 { return cupTFCursorWidth }

// TextFieldCornerRadius returns the Cupertino corner radius.
func (TextFieldPainter) TextFieldCornerRadius() float32 { return cupTFCornerRadius }

// resolveColors returns the TextFieldColorScheme derived from the painter's Theme.
func (p TextFieldPainter) resolveColors() textfield.TextFieldColorScheme {
	if p.Theme == nil {
		return cupDefaultTextFieldColors
	}
	cs := p.Theme.Colors
	return textfield.TextFieldColorScheme{
		Background:  cs.TertiarySystemBackground,
		Border:      cs.Separator,
		FocusBorder: cs.Accent,
		ErrorBorder: cs.SystemRed,
		TextColor:   cs.Label,
		Placeholder: cs.TertiaryLabel,
		CursorColor: cs.Accent,
		DisabledBg:  cs.SecondarySystemBackground,
		DisabledFg:  cs.QuaternaryLabel,
		SelectionBg: cs.Accent.WithAlpha(cupTFSelectionAlpha),
		ErrorText:   cs.SystemRed,
	}
}

// PaintTextField renders a text field according to Apple HIG specifications.
// Cursor and selection positions come from pre-computed PaintState fields.
func (p TextFieldPainter) PaintTextField(canvas widget.Canvas, st *textfield.PaintState) {
	if st.Bounds.IsEmpty() {
		return
	}

	colors := p.resolveColors()
	fontSize := st.FontSize
	if fontSize <= 0 {
		fontSize = cupTFFontSize
	}

	cupPaintTFBackground(canvas, st, colors)
	cupPaintTFBorder(canvas, st, colors)
	cupPaintTFContent(canvas, st, colors, fontSize)
	cupPaintTFCursorFromState(canvas, st, colors)
	cupPaintTFError(canvas, st, colors)
}

// cupPaintTFBackground draws the text field background.
func cupPaintTFBackground(canvas widget.Canvas, st *textfield.PaintState, colors textfield.TextFieldColorScheme) {
	bg := colors.Background
	if st.Disabled {
		bg = colors.DisabledBg
	}
	canvas.DrawRoundRect(st.Bounds, bg, cupTFCornerRadius)
}

// cupPaintTFBorder draws the text field outline.
func cupPaintTFBorder(canvas widget.Canvas, st *textfield.PaintState, colors textfield.TextFieldColorScheme) {
	borderColor := colors.Border
	strokeWidth := cupTFBorderWidth

	switch {
	case st.Disabled:
		borderColor = colors.DisabledFg
	case st.HasError:
		borderColor = colors.ErrorBorder
		strokeWidth = cupTFFocusBorderWidth
	case st.Focused:
		borderColor = colors.FocusBorder
		strokeWidth = cupTFFocusBorderWidth
	case st.Hovered:
		borderColor = colors.TextColor
	}

	canvas.StrokeRoundRect(st.Bounds, borderColor, cupTFCornerRadius, strokeWidth)
}

// cupPaintTFContent draws the text or placeholder using pre-computed fields.
func cupPaintTFContent(canvas widget.Canvas, st *textfield.PaintState, colors textfield.TextFieldColorScheme, fontSize float32) {
	canvas.PushClip(st.ContentRect)
	defer canvas.PopClip()

	if st.DisplayText == "" && !st.Focused {
		color := colors.Placeholder
		if st.Disabled {
			color = colors.DisabledFg
		}
		canvas.DrawText(st.Placeholder, st.ContentRect, fontSize, color, false, cupTFTextAlignLeft)
		return
	}

	textColor := colors.TextColor
	if st.Disabled {
		textColor = colors.DisabledFg
	}

	if st.ShowSelection {
		canvas.DrawRect(st.SelectionRect, colors.SelectionBg)
	}

	canvas.DrawText(st.DisplayText, st.ContentRect, fontSize, textColor, false, cupTFTextAlignLeft)
}

// cupPaintTFCursorFromState draws the cursor using pre-computed CursorRect.
func cupPaintTFCursorFromState(canvas widget.Canvas, st *textfield.PaintState, colors textfield.TextFieldColorScheme) {
	if !st.ShowCursor {
		return
	}

	top := geometry.Pt(st.CursorRect.Min.X, st.CursorRect.Min.Y)
	bottom := geometry.Pt(st.CursorRect.Min.X, st.CursorRect.Max.Y)
	canvas.DrawLine(top, bottom, colors.CursorColor, st.CursorRect.Width())
}

// cupPaintTFError draws the error message below the text field.
func cupPaintTFError(canvas widget.Canvas, st *textfield.PaintState, colors textfield.TextFieldColorScheme) {
	if !st.HasError || st.ErrorMsg == "" {
		return
	}

	errBounds := geometry.Rect{
		Min: geometry.Pt(st.Bounds.Min.X+cupTFContentPaddingH, st.Bounds.Max.Y+cupTFErrorTopGap),
		Max: geometry.Pt(st.Bounds.Max.X, st.Bounds.Max.Y+cupTFErrorTopGap+cupTFErrorFontSize+cupTFErrorBottomPad),
	}
	canvas.DrawText(st.ErrorMsg, errBounds, cupTFErrorFontSize, colors.ErrorText, false, cupTFTextAlignLeft)
}

// cupDefaultTextFieldColors holds the default Cupertino text field color scheme.
var cupDefaultTextFieldColors = textfield.TextFieldColorScheme{
	Background:  widget.ColorWhite,
	Border:      widget.RGBA(0.235, 0.235, 0.263, 0.29),
	FocusBorder: systemBlue,
	ErrorBorder: widget.Hex(0xFF3B30),
	TextColor:   widget.RGBA(0.0, 0.0, 0.0, 1.0),
	Placeholder: widget.RGBA(0.235, 0.235, 0.263, 0.3),
	CursorColor: systemBlue,
	DisabledBg:  widget.Hex(0xF2F2F7),
	DisabledFg:  widget.RGBA(0.235, 0.235, 0.263, 0.18),
	SelectionBg: systemBlue.WithAlpha(0.2),
	ErrorText:   widget.Hex(0xFF3B30),
}

// Cupertino text field drawing constants.
const (
	cupTFCornerRadius     float32 = 8
	cupTFBorderWidth      float32 = 0.5
	cupTFFocusBorderWidth float32 = 2
	cupTFContentPaddingH  float32 = 12
	cupTFContentPaddingV  float32 = 10
	cupTFFontSize         float32 = 15
	cupTFTextAlignLeft            = widget.TextAlignLeft
	cupTFCursorWidth      float32 = 2
	cupTFErrorFontSize    float32 = 12
	cupTFErrorTopGap      float32 = 4
	cupTFErrorBottomPad   float32 = 4
	cupTFSelectionAlpha   float32 = 0.2
)

// Compile-time checks.
var (
	_ textfield.Painter       = TextFieldPainter{}
	_ textfield.LayoutMetrics = TextFieldPainter{}
)
