// Package textmetrics provides text-to-pixel coordinate mapping for text
// editing widgets. It is the single source of truth for cursor positioning,
// selection highlighting, and mouse-click-to-rune conversion.
//
// Positions come from [widget.TextCaretPositioner] when the canvas implements
// it, so the caret sits where the renderer actually puts each glyph, and fall
// back to [widget.Canvas.MeasureText] on the prefix otherwise. Both replace
// the approximate charWidthRatio constants that were previously hardcoded in
// individual theme painters.
//
// This package is INTERNAL — not part of the public API. Widget implementations
// (TextField, future Editor) use it; theme painters do not.
package textmetrics

import (
	"math"

	"github.com/gogpu/ui/geometry"
	"github.com/gogpu/ui/widget"
)

// Metrics provides text measurement for a specific canvas and font size.
//
// A Metrics value memoizes the caret offsets of the last string it was asked
// about, so one paint measures the text once rather than once per cursor,
// selection edge and hit test.
type Metrics struct {
	Canvas   widget.Canvas
	FontSize float32

	// cachedOffsets holds the caret offsets of cachedText at cachedSize.
	// A nil slice means nothing has been measured yet.
	cachedOffsets []float32
	cachedText    string
	cachedSize    float32
}

// CaretOffsets returns the caret offset, in pixels from the text origin,
// before each rune of displayText, plus a final entry after the last rune.
//
// The canvas supplies them when it implements [widget.TextCaretPositioner];
// that path shapes the whole string, so a kerning pair or ligature spanning
// the caret is accounted for exactly as the renderer draws it. Otherwise each
// offset is the measured width of the prefix before it, which is all a canvas
// exposing only [widget.Canvas.MeasureText] can offer.
func (m *Metrics) CaretOffsets(displayText string) []float32 {
	if m.cachedOffsets != nil && m.cachedText == displayText && m.cachedSize == m.FontSize {
		return m.cachedOffsets
	}

	var offsets []float32
	if cp, ok := m.Canvas.(widget.TextCaretPositioner); ok {
		offsets = cp.CaretOffsets(displayText, m.FontSize, false)
	} else {
		runes := []rune(displayText)
		offsets = make([]float32, len(runes)+1)
		for i := 1; i <= len(runes); i++ {
			offsets[i] = m.Canvas.MeasureText(string(runes[:i]), m.FontSize, false)
		}
	}

	m.cachedOffsets = offsets
	m.cachedText = displayText
	m.cachedSize = m.FontSize
	return offsets
}

// CaretOffset returns the caret offset, in pixels from the text origin, before
// the rune at runePos. runePos is clamped to the text.
func (m *Metrics) CaretOffset(displayText string, runePos int) float32 {
	offsets := m.CaretOffsets(displayText)
	if runePos < 0 {
		runePos = 0
	}
	if runePos >= len(offsets) {
		runePos = len(offsets) - 1
	}
	return offsets[runePos]
}

// TextWidth returns the advance width of displayText as the renderer lays it
// out, which is the caret offset past the last rune.
func (m *Metrics) TextWidth(displayText string) float32 {
	offsets := m.CaretOffsets(displayText)
	return offsets[len(offsets)-1]
}

// CursorX returns the X coordinate for a cursor at the given rune position
// within the content rect.
//
// The base X is rounded to match DrawText's pixel-grid rounding (math.Round),
// ensuring the cursor aligns with the rendered text origin.
func (m *Metrics) CursorX(contentRect geometry.Rect, displayText string, runePos int) float32 {
	baseX := float32(math.Round(float64(contentRect.Min.X)))
	return baseX + m.CaretOffset(displayText, runePos)
}

// RuneIndexFromX converts an X coordinate to a rune index (for hit-testing).
// Returns the rune position closest to the given X within the content rect.
//
// The base X is rounded to match DrawText's pixel-grid rounding.
func (m *Metrics) RuneIndexFromX(contentRect geometry.Rect, displayText string, x float32) int {
	baseX := float32(math.Round(float64(contentRect.Min.X)))
	localX := x - baseX
	if localX <= 0 {
		return 0
	}

	offsets := m.CaretOffsets(displayText)
	for i := 1; i < len(offsets); i++ {
		if offsets[i] <= localX {
			continue
		}
		if localX-offsets[i-1] < offsets[i]-localX {
			return i - 1
		}
		return i
	}
	return len(offsets) - 1
}

// caretHeightOffset is the amount by which the cursor is inset on each side
// relative to the full line height. This matches Flutter's _kCaretHeightOffset
// constant, making the cursor 2px shorter on top and bottom for a cleaner look.
const caretHeightOffset float32 = 2.0

// CursorRect returns the cursor line rectangle for the given rune position.
// Height is based on fontSize * lineHeightRatio minus caretHeightOffset on
// each side (Flutter _kCaretHeightOffset pattern), vertically centered in
// contentRect.
func (m *Metrics) CursorRect(contentRect geometry.Rect, displayText string, runePos int, cursorWidth float32) geometry.Rect {
	x := m.CursorX(contentRect, displayText, runePos)
	lineHeight := m.FontSize * 1.4
	centerY := (contentRect.Min.Y + contentRect.Max.Y) / 2
	top := centerY - lineHeight/2 + caretHeightOffset
	bottom := centerY + lineHeight/2 - caretHeightOffset
	return geometry.NewRect(x, top, cursorWidth, bottom-top)
}

// SelectionRect returns the selection highlight rectangle for the given
// rune range [start, end). Vertically spans the full content rect height.
func (m *Metrics) SelectionRect(contentRect geometry.Rect, displayText string, start, end int) geometry.Rect {
	if start > end {
		start, end = end, start
	}
	x1 := m.CursorX(contentRect, displayText, start)
	x2 := m.CursorX(contentRect, displayText, end)
	if x2 > contentRect.Max.X {
		x2 = contentRect.Max.X
	}
	return geometry.Rect{
		Min: geometry.Pt(x1, contentRect.Min.Y),
		Max: geometry.Pt(x2, contentRect.Max.Y),
	}
}
