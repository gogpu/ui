package app

import (
	"github.com/gogpu/gpucontext"
	"github.com/gogpu/ui/event"
	"github.com/gogpu/ui/state"
	"github.com/gogpu/ui/theme"
	"github.com/gogpu/ui/widget"
)

// App is the main entry point for the UI framework.
//
// App bridges the widget tree with the windowing system through gpucontext
// interfaces. It manages the Window, theme, and scheduler.
//
// Create an App with [New] and functional options.
type App struct {
	window    *Window
	theme     *theme.Theme
	scheduler *state.Scheduler
	pp        gpucontext.PlatformProvider
}

// Option configures an App during creation.
type Option func(*appConfig)

// appConfig holds configuration gathered from Option functions.
type appConfig struct {
	wp         gpucontext.WindowProvider
	pp         gpucontext.PlatformProvider
	es         gpucontext.EventSource
	theme      *theme.Theme
	renderMode RenderMode
}

// WithWindowProvider sets the window provider for the App.
//
// The WindowProvider provides window geometry (Size, ScaleFactor) and
// the ability to request redraws. When nil, the App operates in headless
// mode with a default 800x600 window at 1x scale.
func WithWindowProvider(wp gpucontext.WindowProvider) Option {
	return func(c *appConfig) {
		c.wp = wp
	}
}

// WithPlatformProvider sets the platform provider for the App.
//
// The PlatformProvider provides OS integration features such as clipboard,
// cursor management, and accessibility preferences. When nil, these
// features are unavailable.
func WithPlatformProvider(pp gpucontext.PlatformProvider) Option {
	return func(c *appConfig) {
		c.pp = pp
	}
}

// WithEventSource sets the event source for the App.
//
// The EventSource delivers input events (keyboard, mouse, scroll, resize,
// focus) from the host application. When nil, no events are delivered.
func WithEventSource(es gpucontext.EventSource) Option {
	return func(c *appConfig) {
		c.es = es
	}
}

// WithTheme sets the theme for the App.
//
// When nil, the default light theme is used.
func WithTheme(t *theme.Theme) Option {
	return func(c *appConfig) {
		c.theme = t
	}
}

// WithRenderMode sets the rendering mode for the App's window.
//
// RenderModeHostManaged (default): the host application draws the background
// before calling DrawTo. DrawTo does not call canvas.Clear and always draws
// the full widget tree.
//
// RenderModeFrameworkManaged: the framework owns the pixmap and draws the
// theme background. Enables frame skip (DrawTo returns false when idle)
// and incremental dirty-region rendering.
//
// See [RenderMode] for detailed documentation of each mode.
func WithRenderMode(mode RenderMode) Option {
	return func(c *appConfig) {
		c.renderMode = mode
	}
}

// New creates a new App with the given options.
//
// When no options are provided, the App operates in headless mode with
// default settings suitable for testing.
func New(opts ...Option) *App {
	cfg := &appConfig{}
	for _, opt := range opts {
		opt(cfg)
	}

	t := cfg.theme
	if t == nil {
		t = theme.DefaultLight()
	}

	a := &App{
		theme: t,
		pp:    cfg.pp,
	}

	// Create scheduler that marks dirty widgets for retained-mode rendering
	// and requests a redraw.
	a.scheduler = state.NewScheduler(func(dirty []widget.Widget) {
		// Set persistent needsRedraw flag on each dirty widget.
		// This flag survives until the draw pass clears it,
		// unlike the scheduler's pending set which is cleared on flush.
		for _, w := range dirty {
			if setter, ok := w.(interface{ SetNeedsRedraw(bool) }); ok {
				setter.SetNeedsRedraw(true)
			}
		}
		if a.window != nil && a.window.wp != nil {
			a.window.wp.RequestRedraw()
		}
	})

	// Create the window.
	a.window = newWindow(cfg.wp, cfg.pp, a.scheduler, t, cfg.renderMode)

	// Attach event bridge if event source is provided.
	if cfg.es != nil {
		attachEventBridge(cfg.es, a.window)
	}

	return a
}

// Window returns the App's Window.
func (a *App) Window() *Window {
	return a.window
}

// Theme returns the App's current theme.
func (a *App) Theme() *theme.Theme {
	return a.theme
}

// PlatformProvider returns the platform provider, or nil if not set.
func (a *App) PlatformProvider() gpucontext.PlatformProvider {
	return a.pp
}

// SetTheme changes the App's theme.
//
// This updates both the App and Window theme and triggers a full redraw.
func (a *App) SetTheme(t *theme.Theme) {
	if t == nil {
		return
	}
	a.theme = t
	a.window.setTheme(t)
}

// Scheduler returns the App's state scheduler.
func (a *App) Scheduler() *state.Scheduler {
	return a.scheduler
}

// SetRoot sets the root widget for the App's window.
//
// The root widget forms the top of the widget tree. It receives
// layout constraints matching the window size and is drawn to fill
// the window.
func (a *App) SetRoot(root widget.Widget) {
	a.window.SetRoot(root)
}

// Frame performs one complete frame: layout, draw, and state management.
//
// This method should be called by the host application's render loop
// (e.g., gogpu.App's draw callback).
func (a *App) Frame() {
	a.window.Frame()
}

// HandleEvent dispatches a single event to the widget tree.
//
// This method is an alternative to using an EventSource. It allows
// the host application to push events manually.
func (a *App) HandleEvent(e event.Event) {
	a.window.HandleEvent(e)
}
