package widget

import "github.com/gogpu/ui/geometry"

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
	os, hasOrigin := child.(originSetter)
	if !hasBounds || !hasOrigin {
		return
	}

	childBounds := bg.Bounds()
	offset := canvas.TransformOffset()
	os.SetScreenOrigin(offset.Add(childBounds.Min))
}
