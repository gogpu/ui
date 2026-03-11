// Package render provides the public API for creating a [widget.Canvas]
// backed by a [gg.Context].
//
// This package is a thin wrapper around the internal render implementation.
// Host applications use it to bridge the gogpu/gg 2D rasterizer with the
// ui widget drawing pipeline.
//
//	cc := ggcanvas.Context()                       // *gg.Context
//	canvas := render.NewCanvas(cc, width, height)  // widget.Canvas
//	window.DrawTo(canvas)                          // draw widget tree
package render

import (
	"github.com/gogpu/gg"
	"github.com/gogpu/ui/internal/render"
	"github.com/gogpu/ui/widget"
)

// NewCanvas creates a [widget.Canvas] backed by the given [gg.Context].
//
// The width and height specify the canvas dimensions in logical pixels.
// The gg.Context should already be created with matching dimensions.
func NewCanvas(dc *gg.Context, width, height int) widget.Canvas {
	return render.NewCanvas(dc, width, height)
}
