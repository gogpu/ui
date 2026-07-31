package gpuview

import (
	"github.com/gogpu/gpucontext"
	"github.com/gogpu/ui/a11y"
	"github.com/gogpu/ui/event"
	"github.com/gogpu/ui/geometry"
	"github.com/gogpu/ui/widget"
)

// Widget provides a GPU-rendered view for external content: 3D scenes, video,
// compute visualization, or custom shader output. Uses the Flutter Texture
// widget pattern: the widget owns an offscreen GPU texture, an external
// renderer draws into it, and the Layer Tree compositor blits it to the surface.
//
// The widget is a RepaintBoundary for compositor isolation — re-rendering
// the external content does not invalidate sibling widgets.
//
// Widget implements [widget.Lifecycle] for GPU resource cleanup on unmount.
type Widget struct {
	widget.WidgetBase

	cfg     config
	texture gpucontext.TextureView
	release func()
	dirty   bool

	// initialized tracks whether the GPU texture has been created.
	// Texture creation is deferred until the first Draw call when the
	// Context's GPUTextureProvider is available.
	initialized bool
}

// Verify interface compliance at compile time.
var (
	_ widget.Widget    = (*Widget)(nil)
	_ widget.Lifecycle = (*Widget)(nil)
)

// New creates a new GPUView widget with the given options.
func New(opts ...Option) *Widget {
	cfg := defaultConfig()
	for _, o := range opts {
		o(&cfg)
	}

	w := &Widget{
		cfg:   cfg,
		dirty: true,
	}
	w.SetVisible(true)
	w.SetEnabled(true)
	w.SetRepaintBoundary(true)
	return w
}

// Layout returns the configured view size, constrained by the parent.
func (w *Widget) Layout(_ widget.Context, constraints geometry.Constraints) geometry.Size {
	preferred := geometry.Sz(float32(w.cfg.width), float32(w.cfg.height))
	return constraints.Constrain(preferred)
}

// Draw renders the view content. On the first call, the GPU texture is
// allocated via the Context's [widget.GPUTextureProvider]. On subsequent
// calls, the OnRender callback is invoked only when the view is dirty
// or in continuous mode.
func (w *Widget) Draw(ctx widget.Context, _ widget.Canvas) {
	if !w.IsVisible() {
		return
	}

	// Deferred texture initialization: create GPU texture on first draw
	// when the GPUTextureProvider is available via the Context.
	if !w.initialized {
		w.initTexture(ctx)
	}

	// Skip rendering if texture was not created (headless, no GPU).
	if w.texture.IsNil() {
		return
	}

	// Invoke render callback when dirty or continuous.
	if w.dirty || w.cfg.continuous {
		if w.cfg.onRender != nil {
			w.cfg.onRender(w.texture)
		}
		w.dirty = false
	}

	// Request next frame for continuous rendering.
	if w.cfg.continuous {
		if sched, ok := ctx.(widget.AnimationScheduler); ok {
			sched.ScheduleAnimationFrame()
		}
		w.SetNeedsRedraw(true)
	}
}

// Event handles input events. GPUView does not consume events by default.
func (w *Widget) Event(_ widget.Context, _ event.Event) bool {
	return false
}

// Children returns nil — GPUView is a leaf widget.
func (w *Widget) Children() []widget.Widget {
	return nil
}

// Mount is called when the widget is added to the tree (Lifecycle interface).
func (w *Widget) Mount(_ widget.Context) {
	// Texture creation happens on first Draw, not Mount, because
	// the GPUTextureProvider may not be available yet during mount.
}

// Unmount is called when the widget is removed from the tree (Lifecycle interface).
// Frees the GPU texture to prevent resource leaks.
func (w *Widget) Unmount() {
	w.Release()
}

// Invalidate marks the view as needing a re-render on the next frame.
// Use this for on-demand rendering when the external content has changed.
func (w *Widget) Invalidate() {
	w.dirty = true
	w.SetNeedsRedraw(true)
}

// Texture returns the offscreen GPU texture view. Returns a nil/zero-value
// TextureView if the texture has not been initialized yet (before first Draw).
func (w *Widget) Texture() gpucontext.TextureView {
	return w.texture
}

// ViewportSize returns the configured view dimensions.
func (w *Widget) ViewportSize() (width, height int) {
	return w.cfg.width, w.cfg.height
}

// IsContinuous reports whether the view renders every frame.
func (w *Widget) IsContinuous() bool {
	return w.cfg.continuous
}

// SetContinuous changes the continuous rendering mode at runtime.
func (w *Widget) SetContinuous(v bool) {
	w.cfg.continuous = v
	if v {
		w.SetNeedsRedraw(true)
	}
}

// Release frees the GPU texture. Must be called when the widget is removed
// from the tree to prevent resource leaks. Called automatically by Unmount.
func (w *Widget) Release() {
	if w.release != nil {
		w.release()
		w.release = nil
	}
	w.texture = gpucontext.TextureView{}
	w.initialized = false
}

// Accessible returns accessibility information for the view.
func (w *Widget) Accessible() a11y.Accessible {
	return &viewAccessible{}
}

// initTexture creates the GPU texture via the Context's GPUTextureProvider.
func (w *Widget) initTexture(ctx widget.Context) {
	provider, ok := ctx.(widget.GPUTextureProvider)
	if !ok {
		return
	}

	texAny, release := provider.CreateGPUTexture(w.cfg.width, w.cfg.height)
	if texAny == nil {
		return
	}

	tex, ok := texAny.(gpucontext.TextureView)
	if !ok {
		// Unexpected type — release if possible and bail.
		if release != nil {
			release()
		}
		return
	}

	w.texture = tex
	w.release = release
	w.initialized = true
}

// viewAccessible provides accessibility info for the GPU view widget.
type viewAccessible struct{}

func (va *viewAccessible) AccessibilityRole() a11y.Role        { return a11y.RoleImage }
func (va *viewAccessible) AccessibilityLabel() string          { return "GPU View" }
func (va *viewAccessible) AccessibilityHint() string           { return "" }
func (va *viewAccessible) AccessibilityValue() string          { return "" }
func (va *viewAccessible) AccessibilityState() a11y.State      { return a11y.State{} }
func (va *viewAccessible) AccessibilityActions() []a11y.Action { return nil }
