package desktop

import (
	"log/slog"
	"os"
	"sync"
	"time"

	"github.com/gogpu/gg"
	"github.com/gogpu/gpucontext"
	"github.com/gogpu/ui/geometry"
)

// Compile-time interface check: dirtyWidgetDebugOverlay must implement DebugOverlay.
var _ gpucontext.DebugOverlay = (*dirtyWidgetDebugOverlay)(nil)

var (
	debugDirtyOnce    sync.Once
	debugDirtyEnabled bool
)

func isDebugDirtyEnabled() bool {
	debugDirtyOnce.Do(func() {
		debugDirtyEnabled = os.Getenv("GOGPU_DEBUG_DIRTY") == "overlay"
	})
	return debugDirtyEnabled
}

// dirtyWidgetDebugOverlay wraps the existing dirtyOverlay to implement
// gpucontext.DebugOverlay. Registered with gogpu's compositor, it draws
// cyan flash-and-fade rectangles over dirty widget regions.
//
// The overlay draws via gg.Context (2D fill/stroke) and flushes to the
// surface view using FlushGPUWithViewPreserveContent, creating a render
// pass with LoadOp::Load that composites on top of existing content.
// This matches the Chromium/GTK4 pattern where overlays draw at the
// compositor level after all content renderers.
type dirtyWidgetDebugOverlay struct {
	overlay dirtyOverlay           // existing flash/fade state
	ctx     *gg.Context            // for rendering (set on first draw)
	regions func() []geometry.Rect // callback to get dirty regions from window
}

// Name returns the overlay identifier for registration and env var filtering.
func (d *dirtyWidgetDebugOverlay) Name() string { return "dirty_widgets" }

// Draw renders the dirty widget overlay for the current frame.
//
// Flow:
//  1. Collect dirty regions from the ui window.
//  2. Update flash state (prune expired, add new).
//  3. Draw cyan flash-and-fade rectangles via gg.Context.
//  4. Flush gg content to the surface view (LoadOp::Load).
//  5. Return true if fade animation still in progress.
func (d *dirtyWidgetDebugOverlay) Draw(ctx gpucontext.DebugOverlayContext) bool {
	if d.ctx == nil {
		return false
	}

	regions := d.regions()
	d.overlay.update(regions)

	if len(d.overlay.flashes) == 0 {
		return false
	}

	// Draw overlay content via gg.Context. Damage tracking is suppressed
	// because the overlay is visualization, not application content.
	d.ctx.SetDamageTracking(false)
	d.overlay.draw(d.ctx)
	d.ctx.SetDamageTracking(true)

	// Flush gg draw commands to the surface view. PreserveContent uses
	// LoadOp::Load so existing surface content (content + damage overlay)
	// is preserved beneath the dirty widget visualization.
	if err := d.ctx.FlushGPUWithViewPreserveContent(
		ctx.SurfaceView, ctx.SurfaceWidth, ctx.SurfaceHeight,
	); err != nil {
		slog.Warn("desktop: dirty widget overlay flush failed", "err", err)
	}

	return d.overlay.needsAnimationFrame()
}

const dirtyFlashDuration = 400 * time.Millisecond

type dirtyFlash struct {
	rect geometry.Rect
	time time.Time
}

// dirtyOverlay tracks dirty regions with flash-and-fade effect.
// Android SurfaceFlinger pattern: flash on dirty, fade over duration.
// In debug mode, extra frames are requested for the fade animation.
type dirtyOverlay struct {
	flashes []dirtyFlash
}

func (o *dirtyOverlay) update(regions []geometry.Rect) {
	now := time.Now()

	// Prune expired.
	alive := o.flashes[:0]
	for _, f := range o.flashes {
		if now.Sub(f.time) < dirtyFlashDuration {
			alive = append(alive, f)
		}
	}
	o.flashes = alive

	// Add new.
	for _, r := range regions {
		if r.Width() <= 0 || r.Height() <= 0 {
			continue
		}
		o.flashes = append(o.flashes, dirtyFlash{rect: r, time: now})
	}
}

// draw renders the cyan flash-and-fade overlay.
//
// f.rect is LOGICAL (user-space): it originates from Window.DirtyRegions ->
// WidgetBase.ScreenBounds, i.e. layout coordinates. gg's Fill/Stroke apply the
// device-scale matrix themselves (Context.deviceSpacePath), so the rect must NOT
// be pre-multiplied by DeviceScale here -- doing so scales the overlay by
// deviceScale twice (4x on a 2x display). Same units the sibling
// SetPresentDamage call site uses.
//
// (Finding 3 in issue #195.)
func (o *dirtyOverlay) draw(cc *gg.Context) {
	// Overlay rects are screen-space: ignore any leftover user transform, and
	// do not leak our color/line width into the caller's paint state.
	cc.Push()
	cc.Identity()
	defer cc.Pop()

	now := time.Now()
	for _, f := range o.flashes {
		age := now.Sub(f.time)
		if age >= dirtyFlashDuration {
			continue
		}
		fade := 1.0 - float64(age)/float64(dirtyFlashDuration)

		x := float64(f.rect.Min.X)
		y := float64(f.rect.Min.Y)
		w := float64(f.rect.Max.X - f.rect.Min.X)
		h := float64(f.rect.Max.Y - f.rect.Min.Y)
		if w <= 0 || h <= 0 {
			continue
		}

		cc.SetRGBA(0, 0.7, 0.9, 0.12*fade)
		cc.DrawRectangle(x, y, w, h)
		_ = cc.Fill()

		// Border is inset 1px from the fill on each side. At logical scale a
		// dirty region as small as 1-2px wide is ordinary (a caret, a thin
		// divider) -- guard against the inset going negative, which would draw
		// an inverted, offset border box.
		if w > 2 && h > 2 {
			cc.SetRGBA(0, 0.7, 0.9, 0.7*fade)
			cc.SetLineWidth(2)
			cc.DrawRectangle(x+1, y+1, w-2, h-2)
			_ = cc.Stroke()
		}
	}
}

func (o *dirtyOverlay) needsAnimationFrame() bool {
	if len(o.flashes) == 0 {
		return false
	}
	now := time.Now()
	for _, f := range o.flashes {
		if now.Sub(f.time) < dirtyFlashDuration {
			return true
		}
	}
	return false
}
