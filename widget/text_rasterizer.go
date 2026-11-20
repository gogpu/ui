package widget

import "image"

// TextRasterizer optionally supplies platform text for scene recording.
// Return false to draw or measure the text with the framework's font renderer.
type TextRasterizer interface {
	Rasterize(text string, style TextStyle, scale float32) (TextBitmap, bool)
	Measure(text string, style TextStyle, scale float32) (RasterTextMetrics, bool)
}

// TextBitmap is a rasterized line at the requested device scale. The bitmap
// is immutable, tightly packed and starts at (0, 0). Advance and InkPadding
// are in physical pixels. InkPadding is the space before the text origin.
type TextBitmap struct {
	Image      *image.RGBA
	Advance    int
	InkPadding int
}

// RasterTextMetrics contains measurements in physical pixels. CaretOffsets
// has one entry before each rune plus one at the end of the string.
type RasterTextMetrics struct {
	Advance      int
	CaretOffsets []int
}
