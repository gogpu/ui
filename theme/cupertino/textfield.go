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
// If Theme is nil, TextFieldPainter falls back to the default system blue palette.
type TextFieldPainter struct {
	Theme *Theme // nil uses default system blue fallback
}

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
func (p TextFieldPainter) PaintTextField(canvas widget.Canvas, st textfield.PaintState) {
	if st.Bounds.IsEmpty() {
		return
	}

	colors := p.resolveColors()
	cupPaintTFBackground(canvas, st, colors)
	cupPaintTFBorder(canvas, st, colors)
	cupPaintTFContent(canvas, st, colors)
	cupPaintTFCursor(canvas, st, colors)
	cupPaintTFError(canvas, st, colors)
}

// cupPaintTFBackground draws the text field background.
func cupPaintTFBackground(canvas widget.Canvas, st textfield.PaintState, colors textfield.TextFieldColorScheme) {
	bg := colors.Background
	if st.Disabled {
		bg = colors.DisabledBg
	}
	canvas.DrawRoundRect(st.Bounds, bg, cupTFCornerRadius)
}

// cupPaintTFBorder draws the text field outline.
func cupPaintTFBorder(canvas widget.Canvas, st textfield.PaintState, colors textfield.TextFieldColorScheme) {
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

// cupPaintTFContent draws the text or placeholder.
func cupPaintTFContent(canvas widget.Canvas, st textfield.PaintState, colors textfield.TextFieldColorScheme) {
	contentBounds := cupTFContentRect(st.Bounds)

	canvas.PushClip(contentBounds)
	defer canvas.PopClip()

	displayText := st.Text
	if st.InputType == textfield.TypePassword {
		displayText = cupMaskText(len([]rune(st.Text)))
	}

	if displayText == "" && !st.Focused {
		color := colors.Placeholder
		if st.Disabled {
			color = colors.DisabledFg
		}
		canvas.DrawText(st.Placeholder, contentBounds, cupTFFontSize, color, false, cupTFTextAlignLeft)
		return
	}

	textColor := colors.TextColor
	if st.Disabled {
		textColor = colors.DisabledFg
	}

	if st.SelectStart != st.SelectEnd {
		cupPaintTFSelection(canvas, contentBounds, st, colors)
	}

	canvas.DrawText(displayText, contentBounds, cupTFFontSize, textColor, false, cupTFTextAlignLeft)
}

// cupPaintTFSelection draws the selection highlight.
func cupPaintTFSelection(canvas widget.Canvas, contentBounds geometry.Rect, st textfield.PaintState, colors textfield.TextFieldColorScheme) {
	selStart := st.SelectStart
	selEnd := st.SelectEnd
	if selStart > selEnd {
		selStart, selEnd = selEnd, selStart
	}

	charWidth := cupTFFontSize * cupTFCharWidthRatio
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

// cupPaintTFCursor draws the blinking caret.
func cupPaintTFCursor(canvas widget.Canvas, st textfield.PaintState, colors textfield.TextFieldColorScheme) {
	if !st.Focused || st.Disabled || st.SelectStart != st.SelectEnd {
		return
	}

	content := cupTFContentRect(st.Bounds)
	charWidth := cupTFFontSize * cupTFCharWidthRatio
	cursorX := content.Min.X + float32(st.CursorPos)*charWidth

	if cursorX > content.Max.X {
		cursorX = content.Max.X
	}

	top := geometry.Pt(cursorX, content.Min.Y+cupTFCursorPadding)
	bottom := geometry.Pt(cursorX, content.Max.Y-cupTFCursorPadding)
	canvas.DrawLine(top, bottom, colors.CursorColor, cupTFCursorWidth)
}

// cupPaintTFError draws the error message below the text field.
func cupPaintTFError(canvas widget.Canvas, st textfield.PaintState, colors textfield.TextFieldColorScheme) {
	if !st.HasError || st.ErrorMsg == "" {
		return
	}

	errBounds := geometry.Rect{
		Min: geometry.Pt(st.Bounds.Min.X+cupTFContentPaddingH, st.Bounds.Max.Y+cupTFErrorTopGap),
		Max: geometry.Pt(st.Bounds.Max.X, st.Bounds.Max.Y+cupTFErrorTopGap+cupTFErrorFontSize+cupTFErrorBottomPad),
	}
	canvas.DrawText(st.ErrorMsg, errBounds, cupTFErrorFontSize, colors.ErrorText, false, cupTFTextAlignLeft)
}

// cupTFContentRect returns the inner content area for text.
func cupTFContentRect(bounds geometry.Rect) geometry.Rect {
	return geometry.Rect{
		Min: geometry.Pt(bounds.Min.X+cupTFContentPaddingH, bounds.Min.Y+cupTFContentPaddingV),
		Max: geometry.Pt(bounds.Max.X-cupTFContentPaddingH, bounds.Max.Y-cupTFContentPaddingV),
	}
}

// cupMaskText returns a string of bullet characters with the given length.
func cupMaskText(length int) string {
	runes := make([]rune, length)
	for i := range runes {
		runes[i] = '\u2022'
	}
	return string(runes)
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
	cupTFCharWidthRatio   float32 = 0.55
	cupTFTextAlignLeft            = widget.TextAlignLeft
	cupTFCursorWidth      float32 = 2
	cupTFCursorPadding    float32 = 2
	cupTFErrorFontSize    float32 = 12
	cupTFErrorTopGap      float32 = 4
	cupTFErrorBottomPad   float32 = 4
	cupTFSelectionAlpha   float32 = 0.2
)

// Compile-time check that TextFieldPainter implements Painter.
var _ textfield.Painter = TextFieldPainter{}
