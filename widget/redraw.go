package widget

// NeedsRedrawInTree reports whether any widget in the subtree rooted at w
// needs re-rendering.
//
// This walks the tree starting from w, checking each widget's NeedsRedraw
// flag. Returns true as soon as any dirty widget is found, false if the
// entire subtree is clean.
//
// This is used by the retained-mode rendering system to determine whether
// a draw pass is necessary. When NeedsRedrawInTree returns false for the
// root widget, the entire frame can be skipped because the previous frame's
// output is still valid in the GPU framebuffer.
//
// Widgets that do not embed [WidgetBase] (and thus lack the NeedsRedraw
// method) are treated as always needing redraw, ensuring correct rendering
// for custom widget implementations.
func NeedsRedrawInTree(w Widget) bool {
	if w == nil {
		return false
	}

	// Check this widget's redraw flag.
	type redrawChecker interface {
		NeedsRedraw() bool
	}
	if rc, ok := w.(redrawChecker); ok {
		if rc.NeedsRedraw() {
			return true
		}
	} else {
		// Widget doesn't implement NeedsRedraw — assume it needs redraw
		// for safety (custom widgets without WidgetBase).
		return true
	}

	// Recurse into children.
	for _, child := range w.Children() {
		if NeedsRedrawInTree(child) {
			return true
		}
	}

	return false
}

// NeedsRedrawInTreeNonBoundary reports whether any NON-BOUNDARY widget in
// the subtree needs re-rendering. RepaintBoundary widgets are skipped because
// they manage their own dirty state independently. This prevents offscreen
// animated boundaries (spinner scrolled out of view) from forcing expensive
// root re-recording on every frame.
func NeedsRedrawInTreeNonBoundary(w Widget) bool {
	if w == nil {
		return false
	}

	type boundaryChecker interface {
		IsRepaintBoundary() bool
	}
	isBoundary := false
	if bc, ok := w.(boundaryChecker); ok && bc.IsRepaintBoundary() {
		isBoundary = true
	}

	// Only count non-boundary widgets as dirty triggers.
	// Boundaries manage their own dirty state independently.
	if !isBoundary {
		type redrawChecker interface {
			NeedsRedraw() bool
		}
		if rc, ok := w.(redrawChecker); ok && rc.NeedsRedraw() {
			return true
		}
	}

	// Always recurse into children (including through boundaries)
	// to find dirty non-boundary descendants.
	for _, child := range w.Children() {
		if NeedsRedrawInTreeNonBoundary(child) {
			return true
		}
	}

	return false
}

// ClearRedrawInTree clears the needsRedraw flag on all widgets in the
// subtree rooted at w.
//
// This is called after a successful draw pass to mark the entire tree
// as visually up to date. The next frame will skip rendering unless
// a signal change marks widgets dirty again.
func ClearRedrawInTree(w Widget) {
	if w == nil {
		return
	}

	// Clear this widget's flag.
	type redrawClearer interface {
		ClearRedraw()
	}
	if rc, ok := w.(redrawClearer); ok {
		rc.ClearRedraw()
	}

	// Recurse into children.
	for _, child := range w.Children() {
		ClearRedrawInTree(child)
	}
}

// MarkRedrawInTree sets the needsRedraw flag on all widgets in the
// subtree rooted at w.
//
// This is used when every widget's visual state may have changed, such as
// after SetRoot or a theme change. It ensures the next draw pass will render
// all widgets.
func MarkRedrawInTree(w Widget) {
	if w == nil {
		return
	}

	// Mark this widget.
	type redrawSetter interface {
		SetNeedsRedraw(bool)
	}
	if rs, ok := w.(redrawSetter); ok {
		rs.SetNeedsRedraw(true)
	}

	// If widget is a PARENTLESS RepaintBoundary (root), invalidate its scene.
	// Root has no parent → propagateDirtyUpward from SetNeedsRedraw is no-op →
	// scene stays clean → stale scene replayed. Only invalidate parentless
	// boundaries, not child boundaries — over-invalidation causes full
	// repaint every frame (entire cyan overlay).
	type boundaryInvalidator interface {
		IsRepaintBoundary() bool
		InvalidateScene()
	}
	if bi, ok := w.(boundaryInvalidator); ok && bi.IsRepaintBoundary() {
		isRoot := true
		if pc, ok2 := w.(interface{ Parent() Widget }); ok2 && pc.Parent() != nil {
			isRoot = false
		}
		if isRoot {
			bi.InvalidateScene()
		}
	}

	// Recurse into children.
	for _, child := range w.Children() {
		MarkRedrawInTree(child)
	}
}
