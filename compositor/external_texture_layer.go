package compositor

import (
	"github.com/gogpu/gpucontext"
	"github.com/gogpu/ui/geometry"
)

// ExternalTextureLayer holds a GPU texture provided by an external renderer
// (3D viewport, video player, custom shader). Unlike [PictureLayerImpl] which
// replays a recorded scene, ExternalTextureLayer composites a pre-rendered
// texture directly.
//
// This is the Flutter Texture widget pattern: an external producer renders into
// an offscreen texture, and the compositor blits it at the layer's position
// during composition. The texture is owned externally (by the viewport widget)
// and must remain valid until the layer is removed from the tree.
//
// Enterprise references:
//   - Flutter: TextureLayer (rendering/layer.dart)
//   - Android: TextureView (android.view.TextureView)
//   - Chrome: TextureLayer (cc/layers/texture_layer.h)
type ExternalTextureLayer struct {
	layerBase
	texture gpucontext.TextureView
	width   int
	height  int
}

// NewExternalTextureLayer creates a new ExternalTextureLayer with the given
// texture, dimensions, and screen position. The texture must remain valid
// for the lifetime of this layer.
func NewExternalTextureLayer(texture gpucontext.TextureView, width, height int, screenX, screenY float64) *ExternalTextureLayer {
	l := &ExternalTextureLayer{
		texture: texture,
		width:   width,
		height:  height,
	}
	l.offset = geometry.Pt(float32(screenX), float32(screenY))
	l.needsCompositing = true
	return l
}

// Texture returns the external GPU texture view.
func (l *ExternalTextureLayer) Texture() gpucontext.TextureView { return l.texture }

// SetTexture updates the external GPU texture view. Marks the layer as
// needing re-composition.
func (l *ExternalTextureLayer) SetTexture(tv gpucontext.TextureView) {
	l.texture = tv
	l.MarkNeedsCompositing()
}

// Width returns the texture width in logical pixels.
func (l *ExternalTextureLayer) Width() int { return l.width }

// Height returns the texture height in logical pixels.
func (l *ExternalTextureLayer) Height() int { return l.height }

// SetSize updates the texture dimensions. Marks the layer as needing
// re-composition.
func (l *ExternalTextureLayer) SetSize(w, h int) {
	l.width = w
	l.height = h
	l.MarkNeedsCompositing()
}

// X returns the screen-space X position for texture blitting.
func (l *ExternalTextureLayer) X() float64 { return float64(l.offset.X) }

// Y returns the screen-space Y position for texture blitting.
func (l *ExternalTextureLayer) Y() float64 { return float64(l.offset.Y) }
