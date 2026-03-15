// Package render provides the public API for creating a [widget.Canvas]
// backed by a [gg.Context].
//
// This package is a thin wrapper around the internal render implementation.
// Host applications use it to bridge the gogpu/gg 2D rasterizer with the
// ui widget drawing pipeline:
//
//	cc := ggcanvas.Context()                       // *gg.Context
//	canvas := render.NewCanvas(cc, width, height)  // widget.Canvas
//	window.DrawTo(canvas)                          // draw widget tree
//
// The returned [widget.Canvas] supports all drawing operations required by
// widgets: rectangles, rounded rectangles, text, lines, clipping, and
// push/pop state. GPU-accelerated operations (SDF shapes, clip regions)
// are handled transparently by the underlying gg.Context.
package render
