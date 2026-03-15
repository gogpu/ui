package fluent

import (
	"github.com/gogpu/ui/core/textfield"
	"github.com/gogpu/ui/geometry"
	"github.com/gogpu/ui/widget"
)

// TextFieldPainter renders text fields using Fluent Design tokens.
// Fluent text fields feature a subtle bottom accent border on focus
// and clean rectangular shape with small corner radius.
//
// If Theme is nil, TextFieldPainter falls back to the default Fluent Blue palette.
type TextFieldPainter struct {
	Theme *Theme // nil uses default Fluent Blue fallback
}

// resolveColors returns the TextFieldColorScheme derived from the painter's Theme.
func (p TextFieldPainter) resolveColors() textfield.TextFieldColorScheme {
	if p.Theme == nil {
		return flDefaultTextFieldColors
	}
	cs := p.Theme.Colors
	return textfield.TextFieldColorScheme{
		Background:  cs.SurfaceTertiary,
		Border:      cs.StrokeDefault,
		FocusBorder: cs.Accent,
		ErrorBorder: cs.Error,
		TextColor:   cs.OnSurface,
		Placeholder: cs.OnSurfaceSecond,
		CursorColor: cs.Accent,
		DisabledBg:  cs.FillDisable,
		DisabledFg:  cs.OnSurfaceSecond.WithAlpha(flDisabledAlpha),
		SelectionBg: cs.Accent.WithAlpha(0.2),
		ErrorText:   cs.Error,
	}
}

// PaintTextField renders a text field according to Fluent Design specifications.
func (p TextFieldPainter) PaintTextField(canvas widget.Canvas, st textfield.PaintState) {
	if st.Bounds.IsEmpty() {
		return
	}

	colors := p.resolveColors()
	flPaintTFBackground(canvas, st, colors)
	flPaintTFBorder(canvas, st, colors)
	flPaintTFContent(canvas, st, colors)
	flPaintTFCursor(canvas, st, colors)
	flPaintTFError(canvas, st, colors)
}

// flPaintTFBackground draws the text field background.
func flPaintTFBackground(canvas widget.Canvas, st textfield.PaintState, colors textfield.TextFieldColorScheme) {
	bg := colors.Background
	if st.Disabled {
		bg = colors.DisabledBg
	}
	canvas.DrawRoundRect(st.Bounds, bg, flTFCornerRadius)
}

// flPaintTFBorder draws the text field outline.
func flPaintTFBorder(canvas widget.Canvas, st textfield.PaintState, colors textfield.TextFieldColorScheme) {
	borderColor := colors.Border
	strokeWidth := flTFBorderWidth

	switch {
	case st.Disabled:
		borderColor = colors.DisabledFg
	case st.HasError:
		borderColor = colors.ErrorBorder
		strokeWidth = flTFFocusBorderWidth
	case st.Focused:
		borderColor = colors.FocusBorder
		strokeWidth = flTFFocusBorderWidth
	case st.Hovered:
		borderColor = colors.TextColor
	}

	canvas.StrokeRoundRect(st.Bounds, borderColor, flTFCornerRadius, strokeWidth)
}

// flPaintTFContent draws the text or placeholder.
func flPaintTFContent(canvas widget.Canvas, st textfield.PaintState, colors textfield.TextFieldColorScheme) {
	contentBounds := flTFContentRect(st.Bounds)

	canvas.PushClip(contentBounds)
	defer canvas.PopClip()

	displayText := st.Text
	if st.InputType == textfield.TypePassword {
		displayText = flMaskText(len([]rune(st.Text)))
	}

	if displayText == "" && !st.Focused {
		color := colors.Placeholder
		if st.Disabled {
			color = colors.DisabledFg
		}
		canvas.DrawText(st.Placeholder, contentBounds, flTFFontSize, color, false, flTFTextAlignLeft)
		return
	}

	textColor := colors.TextColor
	if st.Disabled {
		textColor = colors.DisabledFg
	}

	// Selection highlight.
	if st.SelectStart != st.SelectEnd {
		flPaintTFSelection(canvas, contentBounds, st, colors)
	}

	canvas.DrawText(displayText, contentBounds, flTFFontSize, textColor, false, flTFTextAlignLeft)
}

// flPaintTFSelection draws the selection highlight.
func flPaintTFSelection(canvas widget.Canvas, contentBounds geometry.Rect, st textfield.PaintState, colors textfield.TextFieldColorScheme) {
	selStart := st.SelectStart
	selEnd := st.SelectEnd
	if selStart > selEnd {
		selStart, selEnd = selEnd, selStart
	}

	charWidth := flTFFontSize * flTFCharWidthRatio
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

// flPaintTFCursor draws the blinking caret.
func flPaintTFCursor(canvas widget.Canvas, st textfield.PaintState, colors textfield.TextFieldColorScheme) {
	if !st.Focused || st.Disabled || st.SelectStart != st.SelectEnd {
		return
	}

	content := flTFContentRect(st.Bounds)
	charWidth := flTFFontSize * flTFCharWidthRatio
	cursorX := content.Min.X + float32(st.CursorPos)*charWidth

	if cursorX > content.Max.X {
		cursorX = content.Max.X
	}

	top := geometry.Pt(cursorX, content.Min.Y+flTFCursorPadding)
	bottom := geometry.Pt(cursorX, content.Max.Y-flTFCursorPadding)
	canvas.DrawLine(top, bottom, colors.CursorColor, flTFCursorWidth)
}

// flPaintTFError draws the error message below the text field.
func flPaintTFError(canvas widget.Canvas, st textfield.PaintState, colors textfield.TextFieldColorScheme) {
	if !st.HasError || st.ErrorMsg == "" {
		return
	}

	errBounds := geometry.Rect{
		Min: geometry.Pt(st.Bounds.Min.X+flTFContentPaddingH, st.Bounds.Max.Y+flTFErrorTopGap),
		Max: geometry.Pt(st.Bounds.Max.X, st.Bounds.Max.Y+flTFErrorTopGap+flTFErrorFontSize+flTFErrorBottomPad),
	}
	canvas.DrawText(st.ErrorMsg, errBounds, flTFErrorFontSize, colors.ErrorText, false, flTFTextAlignLeft)
}

// flTFContentRect returns the inner content area for text.
func flTFContentRect(bounds geometry.Rect) geometry.Rect {
	return geometry.Rect{
		Min: geometry.Pt(bounds.Min.X+flTFContentPaddingH, bounds.Min.Y+flTFContentPaddingV),
		Max: geometry.Pt(bounds.Max.X-flTFContentPaddingH, bounds.Max.Y-flTFContentPaddingV),
	}
}

// flMaskText returns a string of bullet characters.
func flMaskText(length int) string {
	runes := make([]rune, length)
	for i := range runes {
		runes[i] = '\u2022'
	}
	return string(runes)
}

// flDefaultTextFieldColors holds the default Fluent text field color scheme.
var flDefaultTextFieldColors = textfield.TextFieldColorScheme{
	Background:  widget.Hex(0xFAFAFA),
	Border:      widget.RGBA(0, 0, 0, 0.14),
	FocusBorder: DefaultAccentColor,
	ErrorBorder: widget.Hex(0xC42B1C),
	TextColor:   widget.Hex(0x1A1A1A),
	Placeholder: widget.Hex(0x616161),
	CursorColor: DefaultAccentColor,
	DisabledBg:  widget.RGBA(0, 0, 0, 0.04),
	DisabledFg:  widget.RGBA(0.38, 0.38, 0.38, 0.38),
	SelectionBg: DefaultAccentColor.WithAlpha(0.2),
	ErrorText:   widget.Hex(0xC42B1C),
}

// Fluent text field drawing constants.
const (
	flTFCornerRadius     float32 = 4
	flTFBorderWidth      float32 = 1
	flTFFocusBorderWidth float32 = 2
	flTFContentPaddingH  float32 = 12
	flTFContentPaddingV  float32 = 8
	flTFFontSize         float32 = 14
	flTFCharWidthRatio   float32 = 0.55
	flTFTextAlignLeft            = widget.TextAlignLeft
	flTFCursorWidth      float32 = 1.5
	flTFCursorPadding    float32 = 2
	flTFErrorFontSize    float32 = 12
	flTFErrorTopGap      float32 = 4
	flTFErrorBottomPad   float32 = 4
)

// Compile-time check that TextFieldPainter implements Painter.
var _ textfield.Painter = TextFieldPainter{}
