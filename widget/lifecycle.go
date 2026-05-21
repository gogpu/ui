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

	// Clear mounted state.
	if base, ok := w.(interface{ SetMounted(bool) }); ok {
		base.SetMounted(false)
	}
}
