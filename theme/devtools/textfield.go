package devtools

import (
	"github.com/gogpu/ui/core/textfield"
	"github.com/gogpu/ui/geometry"
	"github.com/gogpu/ui/widget"
)

// TextFieldPainter renders text fields using DevTools design tokens.
// DevTools text fields use a deep InputBackground (#1E1F22) with 4px radius,
// subtle Gray5 border, and Blue6 border on focus — matching JetBrains IDE
// input field styling.
//
// If Theme is nil, TextFieldPainter falls back to the default DevTools dark palette.
type TextFieldPainter struct {
	Theme *Theme // nil uses default DevTools dark fallback
}

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
func (p TextFieldPainter) PaintTextField(canvas widget.Canvas, st textfield.PaintState) {
	if st.Bounds.IsEmpty() {
		return
	}

	colors := p.resolveColors()
	dtPaintTFBackground(canvas, st, colors)
	dtPaintTFBorder(canvas, st, colors)
	dtPaintTFContent(canvas, st, colors)
	dtPaintTFCursor(canvas, st, colors)
	dtPaintTFError(canvas, st, colors)
}

// dtPaintTFBackground draws the text field background.
func dtPaintTFBackground(canvas widget.Canvas, st textfield.PaintState, colors textfield.TextFieldColorScheme) {
	bg := colors.Background
	if st.Disabled {
		bg = colors.DisabledBg
	}
	canvas.DrawRoundRect(st.Bounds, bg, dtTFCornerRadius)
}

// dtPaintTFBorder draws the text field outline.
func dtPaintTFBorder(canvas widget.Canvas, st textfield.PaintState, colors textfield.TextFieldColorScheme) {
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

// dtPaintTFContent draws the text or placeholder.
func dtPaintTFContent(canvas widget.Canvas, st textfield.PaintState, colors textfield.TextFieldColorScheme) {
	contentBounds := dtTFContentRect(st.Bounds)

	canvas.PushClip(contentBounds)
	defer canvas.PopClip()

	displayText := st.Text
	if st.InputType == textfield.TypePassword {
		displayText = dtMaskText(len([]rune(st.Text)))
	}

	if displayText == "" && !st.Focused {
		color := colors.Placeholder
		if st.Disabled {
			color = colors.DisabledFg
		}
		canvas.DrawText(st.Placeholder, contentBounds, dtTFFontSize, color, false, dtTFTextAlignLeft)
		return
	}

	textColor := colors.TextColor
	if st.Disabled {
		textColor = colors.DisabledFg
	}

	// Selection highlight.
	if st.SelectStart != st.SelectEnd {
		dtPaintTFSelection(canvas, contentBounds, st, colors)
	}

	canvas.DrawText(displayText, contentBounds, dtTFFontSize, textColor, false, dtTFTextAlignLeft)
}

// dtPaintTFSelection draws the selection highlight.
func dtPaintTFSelection(canvas widget.Canvas, contentBounds geometry.Rect, st textfield.PaintState, colors textfield.TextFieldColorScheme) {
	selStart := st.SelectStart
	selEnd := st.SelectEnd
	if selStart > selEnd {
		selStart, selEnd = selEnd, selStart
	}

	charWidth := dtTFFontSize * dtTFCharWidthRatio
	x1 := contentBounds.Min.X + float32(selStart)*charWidth
	x2 := contentBounds.Min.X + float32(selEnd)*charWidth

	if x2 > contentBounds.Max.X {
		x2 = contentBounds.Max.X
	}

	selRect := geometry.Rect{
		Min: geometry.Pt(x1, contentBounds.Min.Y),
		Max: geometry.Pt(x2, contentBounds.Max.Y),
	}
	canvas.DrawRect(selRect, colors.SelectionBg)
}

// dtPaintTFCursor draws the blinking caret.
func dtPaintTFCursor(canvas widget.Canvas, st textfield.PaintState, colors textfield.TextFieldColorScheme) {
	if !st.Focused || st.Disabled || st.SelectStart != st.SelectEnd {
		return
	}

	content := dtTFContentRect(st.Bounds)
	charWidth := dtTFFontSize * dtTFCharWidthRatio
	cursorX := content.Min.X + float32(st.CursorPos)*charWidth

	if cursorX > content.Max.X {
		cursorX = content.Max.X
	}

	top := geometry.Pt(cursorX, content.Min.Y+dtTFCursorPadding)
	bottom := geometry.Pt(cursorX, content.Max.Y-dtTFCursorPadding)
	canvas.DrawLine(top, bottom, colors.CursorColor, dtTFCursorWidth)
}

// dtPaintTFError draws the error message below the text field.
func dtPaintTFError(canvas widget.Canvas, st textfield.PaintState, colors textfield.TextFieldColorScheme) {
	if !st.HasError || st.ErrorMsg == "" {
		return
	}

	errBounds := geometry.Rect{
		Min: geometry.Pt(st.Bounds.Min.X+dtTFContentPaddingH, st.Bounds.Max.Y+dtTFErrorTopGap),
		Max: geometry.Pt(st.Bounds.Max.X, st.Bounds.Max.Y+dtTFErrorTopGap+dtTFErrorFontSize+dtTFErrorBottomPad),
	}
	canvas.DrawText(st.ErrorMsg, errBounds, dtTFErrorFontSize, colors.ErrorText, false, dtTFTextAlignLeft)
}

// dtTFContentRect returns the inner content area for text.
func dtTFContentRect(bounds geometry.Rect) geometry.Rect {
	return geometry.Rect{
		Min: geometry.Pt(bounds.Min.X+dtTFContentPaddingH, bounds.Min.Y+dtTFContentPaddingV),
		Max: geometry.Pt(bounds.Max.X-dtTFContentPaddingH, bounds.Max.Y-dtTFContentPaddingV),
	}
}

// dtMaskText returns a string of bullet characters.
func dtMaskText(length int) string {
	runes := make([]rune, length)
	for i := range runes {
		runes[i] = '\u2022'
	}
	return string(runes)
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
	dtTFCharWidthRatio   float32 = 0.55
	dtTFTextAlignLeft            = widget.TextAlignLeft
	dtTFCursorWidth      float32 = 1
	dtTFCursorPadding    float32 = 2
	dtTFErrorFontSize    float32 = 11
	dtTFErrorTopGap      float32 = 2
	dtTFErrorBottomPad   float32 = 2
)

// Compile-time check that TextFieldPainter implements Painter.
var _ textfield.Painter = TextFieldPainter{}
