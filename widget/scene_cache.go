package widget

// SceneCache is an opaque retained-mode display list for RepaintBoundary caching.
// The concrete implementation is *scene.Scene from gg, but widget code treats it
// as a black box — storing, passing, resetting, and querying emptiness only.
type SceneCache interface {
	Reset()
	IsEmpty() bool
}

// SceneFactory creates new SceneCache instances for RepaintBoundary recording.
// Registered by the rendering layer (app package) at initialization.
type SceneFactory func() SceneCache

// defaultSceneFactory holds the registered SceneFactory.
// Set by the app layer during initialization via RegisterSceneFactory.
var defaultSceneFactory SceneFactory

// RegisterSceneFactory registers the factory function for creating SceneCache
// instances. This must be called by the app layer before any boundary draws
// occur (typically in package init or Window creation).
func RegisterSceneFactory(factory SceneFactory) {
	defaultSceneFactory = factory
}

// NewSceneCache creates a new SceneCache via the registered factory.
// Panics if no factory has been registered — the app layer MUST register
// a factory before the first frame.
func NewSceneCache() SceneCache {
	if defaultSceneFactory == nil {
		panic("widget: no SceneFactory registered — app layer must call widget.RegisterSceneFactory before first frame")
	}
	return defaultSceneFactory()
}
