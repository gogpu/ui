package widget

import (
	"fmt"
	"os"

	"github.com/gogpu/ui/geometry"
)

var debugStamp = os.Getenv("GOGPU_DEBUG_STAMP") == "1"

// StampScreenOrigin computes and records the screen-space origin on a widget
// using the canvas's current transform offset and the widget's local bounds.
//
// This should be called by container widgets in their Draw method, after
// calling [Canvas.PushTransform] for the container's position, and before
// calling Draw on each child widget. The framework also calls this for the
// root widget before the first Draw.
//
// The screen origin accounts for all accumulated parent transforms including
// scroll offsets, making it correct for overlay positioning.
//
// Example usage in a container's Draw method:
//
//	canvas.PushTransform(bounds.Min)
//	for _, child := range children {
//	    widget.StampScreenOrigin(child, canvas)
//	    child.Draw(ctx, canvas)
//	}
//	canvas.PopTransform()
func StampScreenOrigin(child Widget, canvas Canvas) {
	type boundsGetter interface {
		Bounds() geometry.Rect
	}
	type originSetter interface {
		SetScreenOrigin(geometry.Point)
	}

	bg, hasBounds := child.(boundsGetter)
	setter, hasOrigin := child.(originSetter)
	if !hasBounds || !hasOrigin {
		return
	}

	childBounds := bg.Bounds()
	offset := canvas.TransformOffset().Add(canvas.ScreenOriginBase())
	screenOrigin := offset.Add(childBounds.Min)
	if debugStamp {
		fmt.Fprintf(os.Stderr, "[STAMP] %T bounds=%v canvasOffset=%v → screen=%v\n",
			child, childBounds, offset, screenOrigin)
	}
	setter.SetScreenOrigin(screenOrigin)
}

// stampCompositorClip records the canvas's current clip rect (in screen space)
// on a skipped boundary child. compositeTextures uses this to cull textures
// outside the viewport (e.g., ScrollView clips).
//
// The clip rect from the canvas is in the recording coordinate system
// (local to the parent boundary). We convert to screen space by adding
// the canvas's screenOriginBase.
func stampCompositorClip(child Widget, canvas Canvas) {
	type clipSetter interface {
		SetCompositorClip(geometry.Rect)
	}
	setter, ok := child.(clipSetter)
	if !ok {
		return
	}

	localClip := canvas.ClipBounds()
	base := canvas.ScreenOriginBase()
	screenClip := geometry.Rect{
		Min: localClip.Min.Add(base),
		Max: localClip.Max.Add(base),
	}

	// Guard against degenerate rects from zero-area clip intersections
	// (e.g., widget fully scrolled out of viewport). A zero-area rect at
	// a non-zero position can confuse downstream culling, so normalize to
	// an explicit zero rect.
	if screenClip.Width() <= 0 || screenClip.Height() <= 0 {
		setter.SetCompositorClip(geometry.Rect{})
		return
	}

	setter.SetCompositorClip(screenClip)
}
