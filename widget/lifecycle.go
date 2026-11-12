package widget

// Lifecycle is an optional interface that widgets implement to receive
// mount/unmount notifications from the widget tree.
//
// When a widget with signal bindings is added to the tree, Mount is called
// to create subscriptions. When removed, Unmount is called to clean them up.
//
// Widgets that do not use signals need not implement this interface.
// The framework checks for Lifecycle via type assertion.
type Lifecycle interface {
	// Mount is called when the widget is added to the tree.
	// Implementations should create signal bindings here via AddBinding().
	Mount(ctx Context)

	// Unmount is called when the widget is removed from the tree.
	// Implementations should clean up any resources not managed by AddBinding().
	// WidgetBase.CleanupBindings() is called automatically before Unmount().
	Unmount()
}

// MountTree recursively mounts all widgets in the subtree rooted at w.
// For each widget that implements Lifecycle, Mount(ctx) is called.
// Widgets already mounted are skipped.
func MountTree(w Widget, ctx Context) {
	if w == nil {
		return
	}

	// Set mounted state on WidgetBase if available.
	if m, ok := w.(interface{ IsMounted() bool }); ok && m.IsMounted() {
		return // already mounted
	}
	if base, ok := w.(interface{ SetMounted(bool) }); ok {
		base.SetMounted(true)
	}
	if d, ok := w.(detachMarker); ok {
		d.setDetached(false)
	}

	// Store the mount context on widgets that support it so a container can
	// mount children added at runtime (see WidgetBase.attachChild).
	if c, ok := w.(interface{ setContext(Context) }); ok {
		c.setContext(ctx)
	}

	// Call Mount if widget implements Lifecycle.
	if lc, ok := w.(Lifecycle); ok {
		lc.Mount(ctx)
	}

	// Recurse into children, establishing parent chain.
	// Flutter adoptChild pattern: every child knows its parent so that
	// propagateDirtyUpward can walk to the nearest RepaintBoundary.
	if children := w.Children(); children != nil {
		for _, child := range children {
			if setter, ok := child.(interface{ SetParent(Widget) }); ok {
				setter.SetParent(w)
			}
			MountTree(child, ctx)
		}
	}
}

// UnmountTree recursively unmounts all widgets in the subtree rooted at w.
// For each widget, bindings are cleaned up and Unmount() is called if implemented.
// Children are unmounted first (bottom-up).
func UnmountTree(w Widget) {
	if w == nil {
		return
	}

	// Recurse into children first (bottom-up unmount).
	if children := w.Children(); children != nil {
		for _, child := range children {
			UnmountTree(child)
		}
	}

	// Cleanup bindings on WidgetBase.
	if base, ok := w.(interface{ CleanupBindings() }); ok {
		base.CleanupBindings()
	}

	// Call Unmount if widget implements Lifecycle.
	if lc, ok := w.(Lifecycle); ok {
		lc.Unmount()
	}

	// Clear parent link.
	if setter, ok := w.(interface{ SetParent(Widget) }); ok {
		setter.SetParent(nil)
	}

	// Clear cached scene to release scene memory (prevents leak after SetRoot).
	type sceneClearer interface{ ClearCachedScene() }
	if sc, ok := w.(sceneClearer); ok {
		sc.ClearCachedScene()
	}

	// Clear mounted state and mark the widget as removed from the tree.
	if base, ok := w.(interface{ SetMounted(bool) }); ok {
		base.SetMounted(false)
	}
	if d, ok := w.(detachMarker); ok {
		d.setDetached(true)
	}

	// Clear the stored mount context so a later runtime add does not try to
	// mount into a stale context.
	if c, ok := w.(interface{ setContext(Context) }); ok {
		c.setContext(nil)
	}
}

// detachMarker is implemented by [WidgetBase] (and every widget embedding it):
// UnmountTree sets the mark, MountTree and a re-attach clear it.
type detachMarker interface {
	setDetached(bool)
	isDetached() bool
}

// IsDetached reports whether w was removed from a widget tree and has not been
// put back since: [UnmountTree] ran on it (a container's RemoveChild, a
// replaced root, a closed overlay) and neither [MountTree] nor a container's
// AddChild/InsertChild has attached it again.
//
// It is not the same as "not mounted". A live widget can be in the tree without
// ever being mounted, when a container adds it at runtime without running
// MountTree (docking.Host.Dock, transition.Slide.SetChild). The window uses
// IsDetached to drop pointer capture and hover held by a removed widget, so it
// must never drop them for such a live one.
//
// A widget that does not embed [WidgetBase] is never reported as detached.
func IsDetached(w Widget) bool {
	if w == nil {
		return false
	}
	if d, ok := w.(detachMarker); ok {
		return d.isDetached()
	}
	return false
}

// clearDetachedTree clears the detached mark on every widget of the subtree
// rooted at w. attachChild calls it when a removed child is added back to a
// container that is not mounted, so the subtree is live again although
// MountTree does not run on it.
func clearDetachedTree(w Widget) {
	if w == nil {
		return
	}
	if d, ok := w.(detachMarker); ok {
		d.setDetached(false)
	}
	for _, child := range w.Children() {
		clearDetachedTree(child)
	}
}
