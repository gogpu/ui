package devtools

import (
	"github.com/gogpu/ui/core/textfield"
	"github.com/gogpu/ui/geometry"
	"github.com/gogpu/ui/widget"
)

// TextFieldPainter renders text fields using DevTools design tokens.
// DevTools text fields use a deep InputBackground (#1E1F22) with 4px radius,
// subtle Gray5 border, and Blue6 border on focus -- matching JetBrains IDE
// input field styling.
//
// TextFieldPainter implements [textfield.LayoutMetrics] to provide DevTools
// spatial metrics (13px font, 8/6px padding, 1px cursor) used by the widget
// to compute pre-computed PaintState fields.
//
// If Theme is nil, TextFieldPainter falls back to the default DevTools dark palette.
type TextFieldPainter struct {
	Theme *Theme // nil uses default DevTools dark fallback
}

// ContentPadding returns the DevTools horizontal and vertical padding.
func (TextFieldPainter) ContentPadding() (float32, float32) {
	return dtTFContentPaddingH, dtTFContentPaddingV
}

// TextFieldFontSize returns the DevTools font size.
func (TextFieldPainter) TextFieldFontSize() float32 { return dtTFFontSize }

// TextFieldCursorWidth returns the DevTools cursor width.
func (TextFieldPainter) TextFieldCursorWidth() float32 { return dtTFCursorWidth }

// TextFieldCornerRadius returns the DevTools corner radius.
func (TextFieldPainter) TextFieldCornerRadius() float32 { return dtTFCornerRadius }

// resolveColors returns the TextFieldColorScheme derived from the painter's Theme.
func (p TextFieldPainter) resolveColors() textfield.TextFieldColorScheme {
	if p.Theme == nil {
		return dtDefaultTextFieldColors
	}
	cs := p.Theme.Colors
	return textfield.TextFieldColorScheme{
		Background:  cs.InputBackground,
		Border:      cs.BorderStrong,
		FocusBorder: cs.BorderFocus,
		ErrorBorder: cs.Error,
		TextColor:   cs.OnSurface,
		Placeholder: cs.OnSurfaceSecondary,
		CursorColor: cs.Primary,
		DisabledBg:  cs.ControlFill,
		DisabledFg:  cs.OnSurfaceDisabled,
		SelectionBg: cs.Selection,
		ErrorText:   cs.Error,
	}
}

// PaintTextField renders a text field according to DevTools design specifications.
// Cursor and selection positions come from pre-computed PaintState fields.
func (p TextFieldPainter) PaintTextField(canvas widget.Canvas, st *textfield.PaintState) {
	if st.Bounds.IsEmpty() {
		return
	}

	colors := p.resolveColors()
	fontSize := st.FontSize
	if fontSize <= 0 {
		fontSize = dtTFFontSize
	}

	dtPaintTFBackground(canvas, st, colors)
	dtPaintTFBorder(canvas, st, colors)
	dtPaintTFContent(canvas, st, colors, fontSize)
	dtPaintTFCursorFromState(canvas, st, colors)
	dtPaintTFError(canvas, st, colors)
}

// dtPaintTFBackground draws the text field background.
func dtPaintTFBackground(canvas widget.Canvas, st *textfield.PaintState, colors textfield.TextFieldColorScheme) {
	bg := colors.Background
	if st.Disabled {
		bg = colors.DisabledBg
	}
	canvas.DrawRoundRect(st.Bounds, bg, dtTFCornerRadius)
}

// dtPaintTFBorder draws the text field outline.
func dtPaintTFBorder(canvas widget.Canvas, st *textfield.PaintState, colors textfield.TextFieldColorScheme) {
	borderColor := colors.Border
	strokeWidth := dtTFBorderWidth

	switch {
	case st.Disabled:
		borderColor = colors.DisabledFg
	case st.HasError:
		borderColor = colors.ErrorBorder
		strokeWidth = dtTFFocusBorderWidth
	case st.Focused:
		borderColor = colors.FocusBorder
		strokeWidth = dtTFFocusBorderWidth
	case st.Hovered:
		borderColor = colors.TextColor
	}

	canvas.StrokeRoundRect(st.Bounds, borderColor, dtTFCornerRadius, strokeWidth)
}

// dtPaintTFContent draws the text or placeholder using pre-computed fields.
func dtPaintTFContent(canvas widget.Canvas, st *textfield.PaintState, colors textfield.TextFieldColorScheme, fontSize float32) {
	canvas.PushClip(st.ContentRect)
	defer canvas.PopClip()

	if st.DisplayText == "" && !st.Focused {
		color := colors.Placeholder
		if st.Disabled {
			color = colors.DisabledFg
		}
		canvas.DrawText(st.Placeholder, st.ContentRect, fontSize, color, false, dtTFTextAlignLeft)
		return
	}

	textColor := colors.TextColor
	if st.Disabled {
		textColor = colors.DisabledFg
	}

	// Selection highlight from pre-computed rect.
	if st.ShowSelection {
		canvas.DrawRect(st.SelectionRect, colors.SelectionBg)
	}

	canvas.DrawText(st.DisplayText, st.ContentRect, fontSize, textColor, false, dtTFTextAlignLeft)
}

// dtPaintTFCursorFromState draws the cursor using pre-computed CursorRect.
func dtPaintTFCursorFromState(canvas widget.Canvas, st *textfield.PaintState, colors textfield.TextFieldColorScheme) {
	if !st.ShowCursor {
		return
	}

	top := geometry.Pt(st.CursorRect.Min.X, st.CursorRect.Min.Y)
	bottom := geometry.Pt(st.CursorRect.Min.X, st.CursorRect.Max.Y)
	canvas.DrawLine(top, bottom, colors.CursorColor, st.CursorRect.Width())
}

// dtPaintTFError draws the error message below the text field.
func dtPaintTFError(canvas widget.Canvas, st *textfield.PaintState, colors textfield.TextFieldColorScheme) {
	if !st.HasError || st.ErrorMsg == "" {
		return
	}

	errBounds := geometry.Rect{
		Min: geometry.Pt(st.Bounds.Min.X+dtTFContentPaddingH, st.Bounds.Max.Y+dtTFErrorTopGap),
		Max: geometry.Pt(st.Bounds.Max.X, st.Bounds.Max.Y+dtTFErrorTopGap+dtTFErrorFontSize+dtTFErrorBottomPad),
	}
	canvas.DrawText(st.ErrorMsg, errBounds, dtTFErrorFontSize, colors.ErrorText, false, dtTFTextAlignLeft)
}

// dtDefaultTextFieldColors holds the default DevTools dark text field color scheme.
var dtDefaultTextFieldColors = textfield.TextFieldColorScheme{
	Background:  widget.Hex(0x1E1F22), // Gray1 (InputBackground)
	Border:      widget.Hex(0x4E5157), // Gray5
	FocusBorder: DefaultAccentColor,
	ErrorBorder: widget.Hex(0xDB5C5C), // Red7
	TextColor:   widget.Hex(0xDFE1E5), // Gray12
	Placeholder: widget.Hex(0x9DA0A8), // Gray9
	CursorColor: DefaultAccentColor,
	DisabledBg:  widget.Hex(0x393B40), // Gray3
	DisabledFg:  widget.Hex(0x6F737A), // Gray7
	SelectionBg: widget.Hex(0x2E436E), // Blue2
	ErrorText:   widget.Hex(0xDB5C5C), // Red7
}

// DevTools text field drawing constants.
const (
	dtTFCornerRadius     float32 = 4
	dtTFBorderWidth      float32 = 1
	dtTFFocusBorderWidth float32 = 1
	dtTFContentPaddingH  float32 = 8
	dtTFContentPaddingV  float32 = 6
	dtTFFontSize         float32 = 13
	dtTFTextAlignLeft            = widget.TextAlignLeft
	dtTFCursorWidth      float32 = 1
	dtTFCursorPadding    float32 = 2
	dtTFErrorFontSize    float32 = 11
	dtTFErrorTopGap      float32 = 2
	dtTFErrorBottomPad   float32 = 2
)

// Compile-time checks.
var (
	_ textfield.Painter       = TextFieldPainter{}
	_ textfield.LayoutMetrics = TextFieldPainter{}
)
