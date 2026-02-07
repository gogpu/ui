package app

import "time"

// FrameStats holds timing information for a single frame.
//
// This is useful for performance monitoring and debugging.
// FrameStats is passed to the [FrameCallback] after each call to
// [App.Frame] or [Window.Frame].
type FrameStats struct {
	// FrameStart is the time when Frame() was called.
	FrameStart time.Time

	// LayoutDuration is how long the layout pass took.
	// Zero if layout was not performed this frame.
	LayoutDuration time.Duration

	// DrawDuration is how long the draw pass took.
	DrawDuration time.Duration

	// TotalDuration is the total time spent in Frame().
	TotalDuration time.Duration

	// LayoutPerformed indicates whether layout was recalculated this frame.
	LayoutPerformed bool
}

// FrameCallback is called after each frame completes with timing statistics.
//
// This is useful for frame time monitoring and adaptive rendering.
// The callback is invoked synchronously at the end of each Frame() call.
type FrameCallback func(stats FrameStats)

// SetFrameCallback sets a callback that is invoked after each Frame call
// with timing statistics.
//
// Set to nil to disable statistics tracking. The callback is called
// synchronously on the same goroutine as Frame().
func (a *App) SetFrameCallback(cb FrameCallback) {
	if a.window == nil {
		return
	}
	a.window.frameCallback = cb
}
