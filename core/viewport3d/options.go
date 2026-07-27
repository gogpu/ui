package viewport3d

import "github.com/gogpu/gpucontext"

// Option configures a Viewport3D widget during construction.
type Option func(*config)

// config holds the resolved configuration for a Viewport3D widget.
type config struct {
	width      int
	height     int
	onRender   func(view gpucontext.TextureView)
	continuous bool
}

// defaultConfig returns a config with sensible defaults.
func defaultConfig() config {
	return config{
		width:  320,
		height: 240,
	}
}

// Size sets the viewport dimensions in logical pixels.
// Default is 320x240.
func Size(w, h int) Option {
	return func(c *config) {
		if w > 0 {
			c.width = w
		}
		if h > 0 {
			c.height = h
		}
	}
}

// OnRender sets the callback invoked each time the viewport needs to
// render a new frame. The callback receives the offscreen texture view
// that the producer should render into.
//
// The callback is invoked during the Draw phase. The texture dimensions
// match the configured [Size]. The producer is responsible for issuing
// GPU commands that write to the provided texture view.
func OnRender(fn func(view gpucontext.TextureView)) Option {
	return func(c *config) {
		c.onRender = fn
	}
}

// Continuous sets whether the viewport re-renders every frame (true) or
// only on-demand when explicitly invalidated (false, default).
//
// Use Continuous(true) for real-time 3D scenes, video playback, or any
// content that changes every frame. Use the default (false) for static
// content that only updates in response to explicit signals.
func Continuous(v bool) Option {
	return func(c *config) {
		c.continuous = v
	}
}
