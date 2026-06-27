// Package textmetrics provides text-to-pixel coordinate mapping for text
// editing widgets. It is the single source of truth for cursor positioning,
// selection highlighting, and mouse-click-to-rune conversion.
//
// All methods use [widget.Canvas.MeasureText] for accurate measurement,
// replacing the approximate charWidthRatio constants that were previously
// hardcoded in individual theme painters.
//
// This package is INTERNAL — not part of the public API. Widget implementations
// (TextField, future Editor) use it; theme painters do not.
package textmetrics

import (
	"github.com/gogpu/ui/geometry"
	"github.com/gogpu/ui/widget"
)

// Metrics provides text measurement for a specific canvas and font size.
type Metrics struct {
	Canvas   widget.Canvas
	FontSize float32
}

// CursorX returns the X coordinate for a cursor at the given rune position
// within the content rect. Uses MeasureText for accurate positioning.
func (m *Metrics) CursorX(contentRect geometry.Rect, displayText string, runePos int) float32 {
	runes := []rune(displayText)
	if runePos > len(runes) {
		runePos = len(runes)
	}
	if runePos <= 0 {
		return contentRect.Min.X
	}
	textBefore := string(runes[:runePos])
	x := contentRect.Min.X + m.Canvas.MeasureText(textBefore, m.FontSize, false)
	if x > contentRect.Max.X {
		x = contentRect.Max.X
	}
	return x
}

// RuneIndexFromX converts an X coordinate to a rune index (for hit-testing).
// Returns the rune position closest to the given X within the content rect.
func (m *Metrics) RuneIndexFromX(contentRect geometry.Rect, displayText string, x float32) int {
	localX := x - contentRect.Min.X
	if localX <= 0 {
		return 0
	}
	runes := []rune(displayText)
	for i := 1; i <= len(runes); i++ {
		w := m.Canvas.MeasureText(string(runes[:i]), m.FontSize, false)
		if w > localX {
			prevW := float32(0)
			if i > 1 {
				prevW = m.Canvas.MeasureText(string(runes[:i-1]), m.FontSize, false)
			}
			if localX-prevW < w-localX {
				return i - 1
			}
			return i
		}
	}
	return len(runes)
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
