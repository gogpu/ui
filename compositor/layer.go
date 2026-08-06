package compositor

import (
	"github.com/gogpu/ui/geometry"
	"github.com/gogpu/ui/widget"
)

// Layer is a node in the compositor layer tree.
//
// Flutter equivalent: rendering/layer.dart Layer class.
// Each Layer has a parent and can be attached/detached from the tree.
type Layer interface {
	// Parent returns the parent layer, or nil for the root.
	Parent() ContainerLayer

	// SetParent sets the parent layer. Called by ContainerLayer.Append/Remove.
	SetParent(parent ContainerLayer)

	// Offset returns this layer's translation offset relative to parent.
	Offset() geometry.Point

	// SetOffset sets the translation offset. When offset changes on an
	// OffsetLayer, no re-record is needed — the compositor applies the
	// new offset during composition (Flutter animated transform).
	SetOffset(offset geometry.Point)

	// NeedsCompositing reports whether this layer or any descendant
	// needs to be re-composited into the parent scene.
	NeedsCompositing() bool

	// MarkNeedsCompositing marks this layer as needing re-composition.
	MarkNeedsCompositing()

	// ClearNeedsCompositing resets the compositing flag after composition.
	ClearNeedsCompositing()
}

// ContainerLayer is a layer that contains child layers.
//
// Flutter equivalent: ContainerLayer (rendering/layer.dart).
// Used as base for OffsetLayer, ClipRectLayer, OpacityLayer.
type ContainerLayer interface {
	Layer

	// Children returns the ordered list of child layers.
	Children() []Layer

	// Append adds a child layer to the end of the children list.
	Append(child Layer)

	// Remove removes a child layer from the children list.
	Remove(child Layer)

	// RemoveAll removes all child layers.
	RemoveAll()
}

// PictureOwner is implemented by layers that own a SceneCache (display list).
//
// Flutter equivalent: PictureLayer.picture.
type PictureOwner interface {
	// Picture returns the SceneCache owned by this layer.
	// Returns nil if the layer has not been recorded yet.
	Picture() widget.SceneCache

	// SetPicture stores a recorded scene. Called after recording a
	// RepaintBoundary's subtree via SceneCanvas.
	SetPicture(s widget.SceneCache)

	// IsDirty reports whether the picture needs re-recording.
	IsDirty() bool

	// MarkDirty marks the picture as needing re-recording.
	MarkDirty()

	// ClearDirty resets the dirty flag after re-recording.
	ClearDirty()
}

// --- Concrete layer types ---

// layerBase provides the common fields for all layer types.
type layerBase struct {
	parent           ContainerLayer
	offset           geometry.Point
	needsCompositing bool
}

func (l *layerBase) Parent() ContainerLayer     { return l.parent }
func (l *layerBase) SetParent(p ContainerLayer) { l.parent = p }
func (l *layerBase) Offset() geometry.Point     { return l.offset }
func (l *layerBase) SetOffset(o geometry.Point) { l.offset = o; l.MarkNeedsCompositing() }
func (l *layerBase) NeedsCompositing() bool     { return l.needsCompositing }
func (l *layerBase) MarkNeedsCompositing()      { l.needsCompositing = true }
func (l *layerBase) ClearNeedsCompositing()     { l.needsCompositing = false }

// containerBase provides the children management for ContainerLayer types.
type containerBase struct {
	layerBase
	children []Layer
}

func (c *containerBase) Children() []Layer { return c.children }

func (c *containerBase) Append(child Layer) {
	child.SetParent(c.asContainer())
	c.children = append(c.children, child)
	c.MarkNeedsCompositing()
}

func (c *containerBase) Remove(child Layer) {
	for i, ch := range c.children {
		if ch == child {
			child.SetParent(nil)
			c.children = append(c.children[:i], c.children[i+1:]...)
			c.MarkNeedsCompositing()
			return
		}
	}
}

func (c *containerBase) RemoveAll() {
	for _, ch := range c.children {
		ch.SetParent(nil)
	}
	c.children = c.children[:0]
	c.MarkNeedsCompositing()
}

// asContainer returns this containerBase as a ContainerLayer interface.
// Subclasses override this to return themselves.
func (c *containerBase) asContainer() ContainerLayer { return nil }

// OffsetLayerImpl is a container layer with a translation offset.
//
// Flutter equivalent: OffsetLayer. Each RepaintBoundary creates one.
// The offset is the widget's screen position. When the widget moves
// (e.g., scroll), only the offset changes — no re-record needed.
type OffsetLayerImpl struct {
	containerBase
}

// NewOffsetLayer creates a new OffsetLayer at the given offset.
func NewOffsetLayer(offset geometry.Point) *OffsetLayerImpl {
	l := &OffsetLayerImpl{}
	l.offset = offset
	l.needsCompositing = true
	return l
}

func (l *OffsetLayerImpl) asContainer() ContainerLayer { return l } //nolint:unused // override for containerBase.Append polymorphism
func (l *OffsetLayerImpl) Append(child Layer) {
	child.SetParent(l)
	l.children = append(l.children, child)
	l.MarkNeedsCompositing()
}

// PictureLayerImpl owns a SceneCache display list. Leaf node.
//
// Flutter equivalent: PictureLayer. Contains the recorded draw
// commands from a RepaintBoundary's subtree.
//
// Phase D fields (boundaryCacheKey, isRoot, width, height) link
// this layer to the per-boundary GPU texture cache in renderLoop.
// BuildLayerTree populates them; compositeTexturesFromTree reads them.
type PictureLayerImpl struct {
	layerBase
	picture          widget.SceneCache
	dirty            bool
	boundaryCacheKey uint64         // Links to per-boundary texture cache (renderLoop.boundaryTextures).
	isRoot           bool           // True for the root boundary (uses DrawGPUTextureBase).
	width            int            // Boundary width in logical pixels.
	height           int            // Boundary height in logical pixels.
	screenOrigin     geometry.Point // Screen-space position for texture blit.
	screenOriginSet  bool           // True when ScreenOrigin was populated by BuildLayerTree.
	clipRect         geometry.Rect  // Compositor clip for viewport culling (ScrollView).
	hasClip          bool           // True when clipRect is set.
	sceneVersion     uint64         // Monotonic counter from WidgetBase.SceneCacheVersion.
}

// NewPictureLayer creates a new PictureLayer (initially dirty, no picture).
func NewPictureLayer() *PictureLayerImpl {
	return &PictureLayerImpl{dirty: true}
}

func (l *PictureLayerImpl) Picture() widget.SceneCache     { return l.picture }
func (l *PictureLayerImpl) SetPicture(s widget.SceneCache) { l.picture = s; l.MarkNeedsCompositing() }
func (l *PictureLayerImpl) IsDirty() bool                  { return l.dirty }
func (l *PictureLayerImpl) MarkDirty()                     { l.dirty = true; l.MarkNeedsCompositing() }
func (l *PictureLayerImpl) ClearDirty()                    { l.dirty = false }

// BoundaryCacheKey returns the unique ID linking this layer to the
// per-boundary GPU texture cache. Set by BuildLayerTree.
func (l *PictureLayerImpl) BoundaryCacheKey() uint64 { return l.boundaryCacheKey }

// SetBoundaryCacheKey stores the boundary's unique cache key.
func (l *PictureLayerImpl) SetBoundaryCacheKey(k uint64) { l.boundaryCacheKey = k }

// IsRoot reports whether this PictureLayer represents the root boundary.
// The root uses DrawGPUTextureBase (background), others use DrawGPUTexture.
func (l *PictureLayerImpl) IsRoot() bool { return l.isRoot }

// SetRoot marks this layer as the root boundary.
func (l *PictureLayerImpl) SetRoot(v bool) { l.isRoot = v }

// Size returns the boundary dimensions in logical pixels.
func (l *PictureLayerImpl) Size() (int, int) { return l.width, l.height }

// SetSize stores the boundary dimensions.
func (l *PictureLayerImpl) SetSize(w, h int) { l.width = w; l.height = h }

// ScreenOrigin returns the screen-space position for texture blitting.
func (l *PictureLayerImpl) ScreenOrigin() geometry.Point { return l.screenOrigin }

// SetScreenOrigin stores the screen-space position from the boundary widget.
func (l *PictureLayerImpl) SetScreenOrigin(p geometry.Point) {
	l.screenOrigin = p
	l.screenOriginSet = true
}

// IsScreenOriginValid reports whether ScreenOrigin was populated.
func (l *PictureLayerImpl) IsScreenOriginValid() bool { return l.screenOriginSet }

// PictureClipRect returns the compositor clip rectangle for viewport culling.
func (l *PictureLayerImpl) PictureClipRect() geometry.Rect { return l.clipRect }

// SetPictureClipRect stores the compositor clip from the boundary widget.
func (l *PictureLayerImpl) SetPictureClipRect(r geometry.Rect) { l.clipRect = r; l.hasClip = true }

// HasPictureClip reports whether a compositor clip is set on this layer.
func (l *PictureLayerImpl) HasPictureClip() bool { return l.hasClip }

// SceneVersion returns the monotonic scene cache version from the source
// boundary widget. Used by renderBoundaryTexturesFromTree to detect fresh
// recordings without accessing the widget tree.
func (l *PictureLayerImpl) SceneVersion() uint64 { return l.sceneVersion }

// SetSceneVersion stores the boundary widget's SceneCacheVersion.
func (l *PictureLayerImpl) SetSceneVersion(v uint64) { l.sceneVersion = v }

// ClipRectLayerImpl is a container layer with a clip rectangle.
//
// Flutter equivalent: ClipRectLayer. Used by ScrollView to clip
// content to the viewport bounds.
type ClipRectLayerImpl struct {
	containerBase
	clipRect geometry.Rect
}

// NewClipRectLayer creates a new ClipRectLayer with the given clip bounds.
func NewClipRectLayer(clip geometry.Rect) *ClipRectLayerImpl {
	l := &ClipRectLayerImpl{clipRect: clip}
	l.needsCompositing = true
	return l
}

func (l *ClipRectLayerImpl) ClipRect() geometry.Rect     { return l.clipRect }
func (l *ClipRectLayerImpl) SetClipRect(r geometry.Rect) { l.clipRect = r; l.MarkNeedsCompositing() }
func (l *ClipRectLayerImpl) asContainer() ContainerLayer { return l } //nolint:unused // override for containerBase.Append polymorphism (TASK-UI-OPT-005)
func (l *ClipRectLayerImpl) Append(child Layer) {
	child.SetParent(l)
	l.children = append(l.children, child)
	l.MarkNeedsCompositing()
}

// OpacityLayerImpl is a container layer with an opacity value.
//
// Flutter equivalent: OpacityLayer. Changing opacity does NOT
// trigger re-record of children — compositor applies alpha.
type OpacityLayerImpl struct {
	containerBase
	opacity float32
}

// NewOpacityLayer creates a new OpacityLayer with the given alpha (0-1).
func NewOpacityLayer(opacity float32) *OpacityLayerImpl {
	l := &OpacityLayerImpl{opacity: opacity}
	l.needsCompositing = true
	return l
}

func (l *OpacityLayerImpl) Opacity() float32            { return l.opacity }
func (l *OpacityLayerImpl) SetOpacity(a float32)        { l.opacity = a; l.MarkNeedsCompositing() }
func (l *OpacityLayerImpl) asContainer() ContainerLayer { return l } //nolint:unused // override for containerBase.Append polymorphism (TASK-UI-OPT-005)
func (l *OpacityLayerImpl) Append(child Layer) {
	child.SetParent(l)
	l.children = append(l.children, child)
	l.MarkNeedsCompositing()
}
