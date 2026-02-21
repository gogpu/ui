package material3

import (
	"github.com/gogpu/ui/core/textfield"
	"github.com/gogpu/ui/geometry"
	"github.com/gogpu/ui/widget"
)

// TextFieldPainter renders text fields using Material 3 design tokens.
// It implements the outlined text field variant with theme-derived colors.
//
// If Theme is nil, TextFieldPainter falls back to the default M3 purple palette.
type TextFieldPainter struct {
	Theme *Theme // nil uses default M3 purple fallback
}

// resolveColors returns the TextFieldColorScheme derived from the painter's Theme.
// If Theme is nil, it returns the default M3 text field color scheme.
func (p TextFieldPainter) resolveColors() textfield.TextFieldColorScheme {
	if p.Theme == nil {
		return m3DefaultTextFieldColors
	}
	cs := p.Theme.Colors
	return textfield.TextFieldColorScheme{
		Background:  cs.Surface,
		Border:      cs.Outline,
		FocusBorder: cs.Primary,
		ErrorBorder: cs.Error,
		TextColor:   cs.OnSurface,
		Placeholder: cs.OnSurfaceVariant,
		CursorColor: cs.Primary,
		DisabledBg:  cs.OnSurface.WithAlpha(0.04),
		DisabledFg:  cs.OnSurface.WithAlpha(0.38),
		SelectionBg: cs.Primary.WithAlpha(0.2),
		ErrorText:   cs.Error,
	}
}

// PaintTextField renders a text field according to Material 3 specifications.
func (p TextFieldPainter) PaintTextField(canvas widget.Canvas, st textfield.PaintState) {
	if st.Bounds.IsEmpty() {
		return
	}

	colors := p.resolveColors()
	m3PaintTextFieldBg(canvas, st, colors)
	m3PaintTextFieldBorder(canvas, st, colors)
	m3PaintTextFieldContent(canvas, st, colors)
	m3PaintTextFieldCursor(canvas, st, colors)
	m3PaintTextFieldError(canvas, st, colors)
}

// m3PaintTextFieldBg draws the text field background.
func m3PaintTextFieldBg(canvas widget.Canvas, st textfield.PaintState, colors textfield.TextFieldColorScheme) {
	bg := colors.Background
	if st.Disabled {
		bg = colors.DisabledBg
	}
	canvas.DrawRoundRect(st.Bounds, bg, m3TFCornerRadius)
}

// m3PaintTextFieldBorder draws the text field outline.
func m3PaintTextFieldBorder(canvas widget.Canvas, st textfield.PaintState, colors textfield.TextFieldColorScheme) {
	borderColor := colors.Border
	strokeWidth := m3TFBorderWidth

	switch {
	case st.Disabled:
		borderColor = colors.DisabledFg
	case st.HasError:
		borderColor = colors.ErrorBorder
		strokeWidth = m3TFFocusBorderWidth
	case st.Focused:
		borderColor = colors.FocusBorder
		strokeWidth = m3TFFocusBorderWidth
	case st.Hovered:
		borderColor = colors.TextColor
	}

	canvas.StrokeRoundRect(st.Bounds, borderColor, m3TFCornerRadius, strokeWidth)
}

// m3PaintTextFieldContent draws the text or placeholder.
func m3PaintTextFieldContent(canvas widget.Canvas, st textfield.PaintState, colors textfield.TextFieldColorScheme) {
	contentBounds := m3TFContentRect(st.Bounds)

	canvas.PushClip(contentBounds)
	defer canvas.PopClip()

	displayText := st.Text
	if st.InputType == textfield.TypePassword {
		displayText = m3MaskText(len([]rune(st.Text)))
	}

	if displayText == "" && !st.Focused {
		color := colors.Placeholder
		if st.Disabled {
			color = colors.DisabledFg
		}
		canvas.DrawText(st.Placeholder, contentBounds, m3TFFontSize, color, false, m3TFTextAlignLeft)
		return
	}

	textColor := colors.TextColor
	if st.Disabled {
		textColor = colors.DisabledFg
	}

	// Draw selection highlight.
	if st.SelectStart != st.SelectEnd {
		m3PaintTextFieldSelection(canvas, contentBounds, st, colors)
	}

	canvas.DrawText(displayText, contentBounds, m3TFFontSize, textColor, false, m3TFTextAlignLeft)
}

// m3PaintTextFieldSelection draws the selection highlight.
func m3PaintTextFieldSelection(canvas widget.Canvas, contentBounds geometry.Rect, st textfield.PaintState, colors textfield.TextFieldColorScheme) {
	selStart := st.SelectStart
	selEnd := st.SelectEnd
	if selStart > selEnd {
		selStart, selEnd = selEnd, selStart
	}

	charWidth := m3TFFontSize * m3TFCharWidthRatio
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

// m3PaintTextFieldCursor draws the blinking caret.
func m3PaintTextFieldCursor(canvas widget.Canvas, st textfield.PaintState, colors textfield.TextFieldColorScheme) {
	if !st.Focused || st.Disabled || st.SelectStart != st.SelectEnd {
		return
	}

	content := m3TFContentRect(st.Bounds)
	charWidth := m3TFFontSize * m3TFCharWidthRatio
	cursorX := content.Min.X + float32(st.CursorPos)*charWidth

	if cursorX > content.Max.X {
		cursorX = content.Max.X
	}

	top := geometry.Pt(cursorX, content.Min.Y+m3TFCursorPadding)
	bottom := geometry.Pt(cursorX, content.Max.Y-m3TFCursorPadding)
	canvas.DrawLine(top, bottom, colors.CursorColor, m3TFCursorWidth)
}

// m3PaintTextFieldError draws the error message below the text field.
func m3PaintTextFieldError(canvas widget.Canvas, st textfield.PaintState, colors textfield.TextFieldColorScheme) {
	if !st.HasError || st.ErrorMsg == "" {
		return
	}

	errBounds := geometry.Rect{
		Min: geometry.Pt(st.Bounds.Min.X+m3TFContentPaddingH, st.Bounds.Max.Y+m3TFErrorTopGap),
		Max: geometry.Pt(st.Bounds.Max.X, st.Bounds.Max.Y+m3TFErrorTopGap+m3TFErrorFontSize+m3TFErrorBottomPadding),
	}
	canvas.DrawText(st.ErrorMsg, errBounds, m3TFErrorFontSize, colors.ErrorText, false, m3TFTextAlignLeft)
}

// m3TFContentRect returns the inner content area for text.
func m3TFContentRect(bounds geometry.Rect) geometry.Rect {
	return geometry.Rect{
		Min: geometry.Pt(bounds.Min.X+m3TFContentPaddingH, bounds.Min.Y+m3TFContentPaddingV),
		Max: geometry.Pt(bounds.Max.X-m3TFContentPaddingH, bounds.Max.Y-m3TFContentPaddingV),
	}
}

// m3MaskText returns a string of bullet characters with the given length.
func m3MaskText(length int) string {
	runes := make([]rune, length)
	for i := range runes {
		runes[i] = '\u2022'
	}
	return string(runes)
}

// m3DefaultTextFieldColors holds the default M3 text field color scheme.
var m3DefaultTextFieldColors = textfield.TextFieldColorScheme{
	Background:  widget.ColorWhite,
	Border:      widget.Hex(0x79747E),                // M3 outline
	FocusBorder: widget.Hex(0x6750A4),                // M3 primary
	ErrorBorder: widget.Hex(0xB3261E),                // M3 error
	TextColor:   widget.Hex(0x1C1B1F),                // M3 on-surface
	Placeholder: widget.Hex(0x49454F),                // M3 on-surface-variant
	CursorColor: widget.Hex(0x6750A4),                // M3 primary
	DisabledBg:  widget.RGBA(0.12, 0.12, 0.13, 0.04), // M3 disabled surface
	DisabledFg:  widget.RGBA(0.12, 0.12, 0.13, 0.38), // M3 disabled fg
	SelectionBg: widget.Hex(0x6750A4).WithAlpha(0.2), // M3 primary 20%
	ErrorText:   widget.Hex(0xB3261E),                // M3 error
}

// M3 text field drawing constants.
const (
	m3TFCornerRadius       float32 = 4
	m3TFBorderWidth        float32 = 1
	m3TFFocusBorderWidth   float32 = 2
	m3TFContentPaddingH    float32 = 16
	m3TFContentPaddingV    float32 = 12
	m3TFFontSize           float32 = 16
	m3TFCharWidthRatio     float32 = 0.55
	m3TFTextAlignLeft      float32 = 0
	m3TFCursorWidth        float32 = 2
	m3TFCursorPadding      float32 = 2
	m3TFErrorFontSize      float32 = 12
	m3TFErrorTopGap        float32 = 4
	m3TFErrorBottomPadding float32 = 4
)

// Compile-time check that TextFieldPainter implements Painter.
var _ textfield.Painter = TextFieldPainter{}
