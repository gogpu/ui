package compositor

import (
	"github.com/gogpu/gg/scene"
)

// Compositor assembles a layer tree into a composed scene.Scene by walking
// layers and appending their content by REFERENCE (AppendWithTranslation),
// not by copying the entire encoding into a flat scene.
//
// NOT IN PRODUCTION PIPELINE: the production render loop (desktop.draw)
// uses direct per-boundary GPU textures instead. Compositor is retained
// for future use with animated transforms and opacity layers.
//
// Flutter equivalent: SceneBuilder in compositeFrame().
type Compositor struct {
	composed *scene.Scene
}

// New creates a new Compositor.
func New() *Compositor {
	return &Compositor{
		composed: scene.NewScene(),
	}
}

// Compose walks the layer tree rooted at root and builds a composed
// scene.Scene by appending each PictureLayer's scene at its accumulated
// offset. The composed scene is returned for rendering.
//
// This is called every frame. The cost is O(layers), not O(draw_commands).
// For 10 boundaries: 10 AppendWithTranslation calls. The actual scene
// data is not re-recorded — only references are assembled.
//
// Flutter equivalent: compositeFrame() → SceneBuilder.addRetained().
func (c *Compositor) Compose(root Layer) *scene.Scene {
	c.composed.Reset()
	c.composeLayer(root, 0, 0)
	return c.composed
}

// composeLayer recursively walks the layer tree, accumulating offsets
// and appending PictureLayer scenes into the composed scene.
func (c *Compositor) composeLayer(layer Layer, parentX, parentY float32) {
	if layer == nil {
		return
	}

	offset := layer.Offset()
	x := parentX + offset.X
	y := parentY + offset.Y

	// PictureLayer: append its scene at accumulated offset.
	if po, ok := layer.(PictureOwner); ok {
		pic := po.Picture()
		if pic != nil {
			if sc, ok2 := pic.(*scene.Scene); ok2 && !sc.IsEmpty() {
				c.composed.AppendWithTranslation(sc, x, y)
			}
		}
		layer.ClearNeedsCompositing()
		return
	}

	// ClipRectLayer: push clip, recurse, pop.
	if cl, ok := layer.(*ClipRectLayerImpl); ok {
		// TODO: push clip into composed scene when scene.Scene supports clip commands.
		// For now, recurse without clip (clip handled by widget-level PushClip).
		_ = cl.ClipRect()
		if container, ok2 := layer.(ContainerLayer); ok2 {
			for _, child := range container.Children() {
				c.composeLayer(child, x, y)
			}
		}
		layer.ClearNeedsCompositing()
		return
	}

	// ContainerLayer / OffsetLayer: recurse into children.
	if container, ok := layer.(ContainerLayer); ok {
		for _, child := range container.Children() {
			c.composeLayer(child, x, y)
		}
	}

	layer.ClearNeedsCompositing()
}

// ComposedScene returns the last composed scene without re-composing.
// Returns nil if Compose has not been called.
func (c *Compositor) ComposedScene() *scene.Scene {
	return c.composed
}
