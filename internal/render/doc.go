// Package render provides the internal rendering implementation for gogpu/ui.
//
// This package is INTERNAL and not intended for public use. It implements the
// [widget.Canvas] interface using gogpu/gg as the 2D drawing backend.
//
// # Architecture
//
// The render package provides:
//
//   - [Canvas]: Implementation of [widget.Canvas] that wraps gg.Context
//   - [Renderer]: Orchestrates render cycles (frame begin/end, surface management)
//   - Color conversion utilities for widget.Color to gg.RGBA
//
// # Canvas Implementation
//
// Canvas wraps a gg.Context and implements all drawing operations required by
// the widget system. It manages:
//
//   - Clip stack: PushClip/PopClip for hierarchical clipping regions
//   - Transform stack: PushTransform/PopTransform for coordinate translation
//   - Drawing primitives: Rectangles, rounded rectangles, circles, lines
//
// # Thread Safety
//
// Canvas is NOT thread-safe. All drawing operations must occur on the main/UI
// thread during the Draw phase. This matches the widget.Canvas contract.
//
// # Usage
//
// This package is used internally by the UI framework. Application code should
// use the widget.Canvas interface instead of directly using this package.
//
//	// Internal framework usage
//	canvas := render.NewCanvas(ggContext, width, height)
//	widget.Draw(ctx, canvas)
package render
