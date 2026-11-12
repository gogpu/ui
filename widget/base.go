package widget

import (
	"sync"

	"github.com/gogpu/ui/geometry"
)

// Unbinder is implemented by signal bindings for cleanup.
// It is defined here to avoid importing the state package from widget.
type Unbinder interface {
	Unbind()
}

// Stopper is implemented by effects for cleanup.
// It is defined here to avoid importing the state package from widget.
type Stopper interface {
	Stop()
}

// WidgetBase provides common functionality for widgets.
//
// Embed this struct in custom widget implementations to get:
//   - Bounds tracking (position and size)
//   - Screen-space coordinate tracking
//   - Focus state management
//   - Visibility control
//   - Enabled/disabled state
//   - Child widget management
//   - Optional ID for debugging
//   - Signal binding lifecycle management
//   - Retained-mode redraw tracking
//
// Example usage:
//
//	type MyButton struct {
//	    widget.WidgetBase
//	    label string
//	}
//
//	func NewMyButton(label string) *MyButton {
//	    b := &MyButton{label: label}
//	    b.SetVisible(true)
//	    b.SetEnabled(true)
//	    return b
//	}
//
// Thread Safety:
//
// WidgetBase uses a mutex to protect its internal state. However, this
// does not make widgets thread-safe for general use. All widget operations
// should occur on the main/UI thread. The mutex is provided for cases
// where properties need to be queried from callbacks.
type WidgetBase struct {
	mu                sync.RWMutex
	bounds            geometry.Rect  // Cached layout bounds
	screenOrigin      geometry.Point // Window-space origin, set during Draw pass
	screenOriginValid bool           // true after first StampScreenOrigin call
	focused           bool           // Whether widget has focus
	visible           bool           // Whether widget is visible
	enabled           bool           // Whether widget accepts input
	needsRedraw       bool           // Whether widget needs re-rendering (retained mode)
	id                string         // Optional ID for debugging
	children          []Widget       // Child widgets
	parent            Widget         // Parent widget (if any)
	bindings          []Unbinder     // Signal bindings (cleaned up on unmount)
	effects           []Stopper      // Effects (stopped on unmount)
	mounted           bool           // Whether widget is currently mounted
	detached          bool           // Removed by UnmountTree and not attached again (see IsDetached)
	self              Widget         // Concrete container owning this base (see SetSelf)
	ctx               Context        // Mount context, stored by MountTree (see SetContext)

	// --- RepaintBoundary property (ADR-024) ---
	// When isRepaintBoundary is true, this widget owns a SceneCache that
	// caches its subtree rendering. Clean boundaries replay cached content
	// instead of re-executing Draw on every descendant.
	isRepaintBoundary     bool
	boundaryCacheKey      uint64     // Unique ID for dirty-set deduplication
	cachedScene           SceneCache // Recorded display list for the subtree
	sceneDirty            bool       // Whether the cached scene needs re-recording
	sceneCacheVersion     uint64     // Monotonic counter (increments on re-record)
	sceneCacheWidth       int        // Cache dimensions for size-change detection
	sceneCacheHeight      int        // Cache dimensions for size-change detection
	onBoundaryDirty       func()     // Callback when boundary transitions to dirty
	suppressDirtyCallback bool       // Suppressed during Draw recording (animation defers render)

	// --- Compositor clip (for per-boundary GPU textures) ---
	// When this boundary is skipped during parent BoundaryRecording (DrawChild),
	// the parent's current clip rect is stored here in screen-space coordinates.
	// compositeTextures uses this to skip/clip textures outside the viewport
	// (e.g., ListView items scrolled outside ScrollView bounds).
	compositorClip    geometry.Rect
	hasCompositorClip bool

	// --- Layout cache (ADR-032 Phase 1) ---
	// Caches the most recent Layout() result keyed by the input constraints so
	// an unchanged subtree is not re-measured on every relayout pass. This is
	// the layout-side analog of the RepaintBoundary scene cache above (the
	// distributed-cache pattern used by Flutter/Android/Qt). The cache is read
	// and written by [LayoutChild] and invalidated by [WidgetBase.MarkNeedsLayout].
	layoutCacheValid bool
	lastConstraints  geometry.Constraints
	lastSize         geometry.Size
	onLayoutDirty    func() // Fires when layout invalidation reaches this node (root → Window)
}

// NewWidgetBase creates a new WidgetBase with default settings.
//
// The widget is visible and enabled by default, with no children
// and zero bounds.
func NewWidgetBase() *WidgetBase {
	return &WidgetBase{
		visible:     true,
		enabled:     true,
		needsRedraw: true,
	}
}

// Bounds returns the widget's current bounds (position and size).
//
// The bounds are set during layout by the parent widget.
func (w *WidgetBase) Bounds() geometry.Rect {
	w.mu.RLock()
	defer w.mu.RUnlock()
	return w.bounds
}

// SetBounds sets the widget's bounds.
//
// This is typically called by the parent widget during layout
// after the child's Layout() method returns its size.
func (w *WidgetBase) SetBounds(bounds geometry.Rect) {
	w.mu.Lock()
	defer w.mu.Unlock()
	w.bounds = bounds
}

// Size returns the widget's current size.
func (w *WidgetBase) Size() geometry.Size {
	w.mu.RLock()
	defer w.mu.RUnlock()
	return w.bounds.Size()
}

// Position returns the widget's top-left position.
func (w *WidgetBase) Position() geometry.Point {
	w.mu.RLock()
	defer w.mu.RUnlock()
	return w.bounds.Min
}

// IsFocused returns true if the widget currently has focus.
func (w *WidgetBase) IsFocused() bool {
	w.mu.RLock()
	defer w.mu.RUnlock()
	return w.focused
}

// SetFocused sets the widget's focus state.
//
// Note: To properly manage focus in the UI, use Context.RequestFocus()
// and Context.ReleaseFocus() instead of calling this directly.
func (w *WidgetBase) SetFocused(focused bool) {
	w.mu.Lock()
	defer w.mu.Unlock()
	w.focused = focused
}

// IsVisible returns true if the widget is visible.
//
// Invisible widgets are not drawn and do not receive events.
func (w *WidgetBase) IsVisible() bool {
	w.mu.RLock()
	defer w.mu.RUnlock()
	return w.visible
}

// SetVisible sets the widget's visibility.
func (w *WidgetBase) SetVisible(visible bool) {
	w.mu.Lock()
	defer w.mu.Unlock()
	w.visible = visible
}

// IsEnabled returns true if the widget accepts input.
//
// Disabled widgets are drawn (usually with a dimmed appearance)
// but do not respond to user input.
func (w *WidgetBase) IsEnabled() bool {
	w.mu.RLock()
	defer w.mu.RUnlock()
	return w.enabled
}

// SetEnabled sets whether the widget accepts input.
func (w *WidgetBase) SetEnabled(enabled bool) {
	w.mu.Lock()
	defer w.mu.Unlock()
	w.enabled = enabled
}

// ID returns the widget's ID for debugging purposes.
//
// IDs are optional and not used by the framework itself.
// They are useful for debugging and testing.
func (w *WidgetBase) ID() string {
	w.mu.RLock()
	defer w.mu.RUnlock()
	return w.id
}

// SetID sets the widget's ID for debugging purposes.
func (w *WidgetBase) SetID(id string) {
	w.mu.Lock()
	defer w.mu.Unlock()
	w.id = id
}

// Parent returns the widget's parent, or nil if none.
func (w *WidgetBase) Parent() Widget {
	w.mu.RLock()
	defer w.mu.RUnlock()
	return w.parent
}

// SetParent sets the widget's parent.
//
// This is called automatically by AddChild and RemoveChild.
func (w *WidgetBase) SetParent(parent Widget) {
	w.mu.Lock()
	defer w.mu.Unlock()
	w.parent = parent
}

// Children returns the widget's child widgets.
//
// Returns nil for leaf widgets with no children.
// The returned slice should not be modified by the caller.
func (w *WidgetBase) Children() []Widget {
	w.mu.RLock()
	defer w.mu.RUnlock()
	if len(w.children) == 0 {
		return nil
	}
	// Return a copy to prevent modification
	result := make([]Widget, len(w.children))
	copy(result, w.children)
	return result
}

// ChildCount returns the number of child widgets.
func (w *WidgetBase) ChildCount() int {
	w.mu.RLock()
	defer w.mu.RUnlock()
	return len(w.children)
}

// ChildAt returns the child at the given index, or nil if out of range.
func (w *WidgetBase) ChildAt(index int) Widget {
	w.mu.RLock()
	defer w.mu.RUnlock()
	if index < 0 || index >= len(w.children) {
		return nil
	}
	return w.children[index]
}

// AddChild adds a child widget.
//
// The child's parent is set to the concrete container (see [WidgetBase.SetSelf])
// and, when the container is mounted, the child's subtree is mounted with the
// container's context so signal bindings become active. The parent's layout is
// then marked dirty so the new child participates in the next layout pass.
func (w *WidgetBase) AddChild(child Widget) {
	if child == nil {
		return
	}
	w.mu.Lock()
	w.children = append(w.children, child)
	w.mu.Unlock()
	w.attachChild(child)
	w.afterChildMutation()
}

// RemoveChild removes a child widget, preserving the order of the remaining
// children.
//
// Returns true if the child was found and removed. The removed child's subtree
// is unmounted (bindings released, parent cleared) so it can be re-added later.
func (w *WidgetBase) RemoveChild(child Widget) bool {
	if child == nil {
		return false
	}
	w.mu.RLock()
	idx := -1
	for i, c := range w.children {
		if c == child {
			idx = i
			break
		}
	}
	w.mu.RUnlock()
	if idx < 0 {
		return false
	}
	w.detachChildAt(idx)
	w.afterChildMutation()
	return true
}

// RemoveChildAt removes the child at the given index, preserving order.
//
// Returns the removed child, or nil if the index is out of range. The removed
// child's subtree is unmounted (bindings released, parent cleared).
func (w *WidgetBase) RemoveChildAt(index int) Widget {
	child := w.detachChildAt(index)
	if child == nil {
		return nil
	}
	w.afterChildMutation()
	return child
}

// ClearChildren removes all child widgets.
//
// Every child subtree is unmounted (bindings released, parent cleared) before
// the list is emptied.
func (w *WidgetBase) ClearChildren() {
	w.mu.Lock()
	children := w.children
	if len(children) == 0 {
		w.mu.Unlock()
		return
	}
	w.children = nil
	ctx := w.ctx
	w.mu.Unlock()

	for _, child := range children {
		if child == nil {
			continue
		}
		releaseFocusInSubtree(ctx, child)
		UnmountTree(child)
	}
	w.afterChildMutation()
}

// InsertChild inserts a child widget at the given index, preserving order.
//
// If index is out of range, the child is appended. The child's parent is set
// and the parent's layout is marked dirty exactly as for [WidgetBase.AddChild].
func (w *WidgetBase) InsertChild(index int, child Widget) {
	if child == nil {
		return
	}
	w.mu.Lock()
	if index < 0 {
		index = 0
	}
	if index >= len(w.children) {
		w.children = append(w.children, child)
	} else {
		w.children = append(w.children, nil)
		copy(w.children[index+1:], w.children[index:])
		w.children[index] = child
	}
	w.mu.Unlock()
	w.attachChild(child)
	w.afterChildMutation()
}

// MoveChild moves the child at index from to index to, preserving the order of
// the other children. It relinks the list without detaching, so the child keeps
// its identity, state and signal bindings — this is the reorder operation the
// enterprise references expose as a first-class call (Flutter render
// move(), Qt insertWidget of an existing item, GTK gtk_box_reorder_child_after).
//
// Returns false when either index is out of range. Moving a child to its
// current index is a valid no-op and returns true.
func (w *WidgetBase) MoveChild(from, to int) bool {
	w.mu.Lock()
	n := len(w.children)
	if from < 0 || from >= n || to < 0 || to >= n {
		w.mu.Unlock()
		return false
	}
	if from == to {
		w.mu.Unlock()
		return true
	}
	child := w.children[from]
	// Remove at from.
	copy(w.children[from:], w.children[from+1:])
	w.children[n-1] = nil
	w.children = w.children[:n-1]
	// Insert at to.
	if to >= len(w.children) {
		w.children = append(w.children, child)
	} else {
		w.children = append(w.children, nil)
		copy(w.children[to+1:], w.children[to:])
		w.children[to] = child
	}
	w.mu.Unlock()
	w.afterChildMutation()
	return true
}

// attachChild links a freshly stored child to this container. It sets the real
// parent from the container registered via [WidgetBase.SetSelf] and mounts the
// child's subtree when the container is already mounted and has a context.
func (w *WidgetBase) attachChild(child Widget) {
	if child == nil {
		return
	}
	w.mu.RLock()
	self := w.self
	mounted := w.mounted
	ctx := w.ctx
	w.mu.RUnlock()

	if self == nil {
		// No concrete container registered; see SetSelf. Leave the child
		// stored but unattached rather than set a wrong or nil parent.
		return
	}
	if setter, ok := child.(interface{ SetParent(Widget) }); ok {
		setter.SetParent(self)
	}
	if mounted && ctx != nil {
		MountTree(child, ctx)
		return
	}
	// Not mounted here: the child is live again all the same, so a detached
	// mark left by an earlier removal must go (see IsDetached).
	clearDetachedTree(child)
}

// detachChildAt removes and returns the child at index (nil if out of range),
// releasing keyboard focus held inside the removed subtree and unmounting it
// bottom-up (bindings released, cached scene cleared, parent cleared).
func (w *WidgetBase) detachChildAt(index int) Widget {
	w.mu.Lock()
	if index < 0 || index >= len(w.children) {
		w.mu.Unlock()
		return nil
	}
	child := w.children[index]
	ctx := w.ctx
	copy(w.children[index:], w.children[index+1:])
	w.children[len(w.children)-1] = nil // Clear reference for GC
	w.children = w.children[:len(w.children)-1]
	w.mu.Unlock()

	if child != nil {
		releaseFocusInSubtree(ctx, child)
		UnmountTree(child)
	}
	return child
}

// afterChildMutation performs the single dirty step for any child list change.
// It marks this widget's layout dirty only (ADR-028/ADR-032): the invalidation
// climbs to the root (scheduling a layout pass) while paint stays scoped to the
// nearest RepaintBoundary on the next LayoutChild cache miss. It deliberately
// does not call ctx.Invalidate() (removed by ADR-028) and does not run a
// synchronous relayout — Qt and GTK coalesce mutations the same way.
func (w *WidgetBase) afterChildMutation() {
	w.MarkNeedsLayout()
}

// releaseFocusInSubtree releases keyboard focus if the currently focused widget
// lies within the subtree rooted at root. It happens synchronously at detach,
// matching Flutter's FocusAttachment.detach and GTK's gtk_widget_unparent, so
// FocusedWidget never reports a widget that is no longer in the tree between
// the detach and the next frame.
func releaseFocusInSubtree(ctx Context, root Widget) {
	if ctx == nil || root == nil {
		return
	}
	focused := ctx.FocusedWidget()
	if focused == nil {
		return
	}
	for cur := focused; cur != nil; {
		if cur == root {
			ctx.ReleaseFocus(focused)
			return
		}
		pg, ok := cur.(interface{ Parent() Widget })
		if !ok {
			return
		}
		cur = pg.Parent()
	}
}

// HasChildren returns true if the widget has any children.
func (w *WidgetBase) HasChildren() bool {
	w.mu.RLock()
	defer w.mu.RUnlock()
	return len(w.children) > 0
}

// ContainsPoint returns true if the point is within the widget's bounds.
//
// This is a convenience method for hit testing.
func (w *WidgetBase) ContainsPoint(p geometry.Point) bool {
	w.mu.RLock()
	defer w.mu.RUnlock()
	return w.bounds.Contains(p)
}

// ScreenOrigin returns the widget's top-left corner in window (screen) coordinates.
//
// This value is computed during the Draw pass by the framework via
// [StampScreenOrigin], reflecting all accumulated transforms from parent
// containers (scroll offsets, box positions, etc.).
//
// Before the first Draw pass completes, this returns the zero point.
func (w *WidgetBase) ScreenOrigin() geometry.Point {
	w.mu.RLock()
	defer w.mu.RUnlock()
	return w.screenOrigin
}

// SetScreenOrigin records the widget's window-space origin.
//
// This is called by the framework during the Draw pass via [StampScreenOrigin].
// User code should not call this method directly.
func (w *WidgetBase) SetScreenOrigin(origin geometry.Point) {
	w.mu.Lock()
	defer w.mu.Unlock()
	w.screenOrigin = origin
	w.screenOriginValid = true
}

// IsScreenOriginValid reports whether ScreenOrigin has been set by
// StampScreenOrigin during a Draw pass. Boundaries with invalid
// ScreenOrigin (never drawn) should not be composited — their
// textures would appear at (0,0) instead of the correct position.
func (w *WidgetBase) IsScreenOriginValid() bool {
	w.mu.RLock()
	defer w.mu.RUnlock()
	return w.screenOriginValid
}

// ScreenBounds returns the widget's bounds in window (screen) coordinates.
//
// Screen bounds are computed during the Draw pass by the framework,
// reflecting all accumulated transforms from parent containers (scroll
// offsets, box positions, etc.). This is the correct value to use when
// positioning overlays, popups, tooltips, and context menus.
//
// Before the first Draw pass completes, this returns a rect at (0,0)
// with the widget's size.
//
// Example:
//
//	func (w *MyWidget) showPopup(ctx widget.Context) {
//	    anchor := w.ScreenBounds()
//	    pos := overlay.Position(overlay.PlacementBelow, anchor, popupSize, windowSize, 4)
//	    // pos is now correct even if w is inside a ScrollView
//	}
func (w *WidgetBase) ScreenBounds() geometry.Rect {
	w.mu.RLock()
	defer w.mu.RUnlock()
	size := w.bounds.Size()
	return geometry.FromPointSize(w.screenOrigin, size)
}

// CompositorClip returns the screen-space clip rect for this boundary.
// Used by compositeTextures to skip textures outside the viewport.
func (w *WidgetBase) CompositorClip() geometry.Rect {
	w.mu.RLock()
	defer w.mu.RUnlock()
	return w.compositorClip
}

// HasCompositorClip returns whether a compositor clip rect has been set.
func (w *WidgetBase) HasCompositorClip() bool {
	w.mu.RLock()
	defer w.mu.RUnlock()
	return w.hasCompositorClip
}

// SetCompositorClip records the screen-space clip rect for this boundary.
// Called by DrawChild when skipping child boundaries during BoundaryRecording.
func (w *WidgetBase) SetCompositorClip(clip geometry.Rect) {
	w.mu.Lock()
	defer w.mu.Unlock()
	w.compositorClip = clip
	w.hasCompositorClip = true
}

// ClearCompositorClip removes the compositor clip rect.
func (w *WidgetBase) ClearCompositorClip() {
	w.mu.Lock()
	defer w.mu.Unlock()
	w.compositorClip = geometry.Rect{}
	w.hasCompositorClip = false
}

// LocalToGlobal converts a point from local coordinates to global (window) coordinates.
//
// Local coordinates are relative to the widget's top-left corner.
// Global coordinates are relative to the window's top-left corner.
//
// This method uses the screen origin computed during the Draw pass,
// which accounts for all parent transforms including scroll offsets.
func (w *WidgetBase) LocalToGlobal(p geometry.Point) geometry.Point {
	w.mu.RLock()
	defer w.mu.RUnlock()
	return p.Add(w.screenOrigin)
}

// GlobalToLocal converts a point from global (window) coordinates to local coordinates.
//
// Local coordinates are relative to the widget's top-left corner.
// Global coordinates are relative to the window's top-left corner.
//
// This method uses the screen origin computed during the Draw pass,
// which accounts for all parent transforms including scroll offsets.
func (w *WidgetBase) GlobalToLocal(p geometry.Point) geometry.Point {
	w.mu.RLock()
	defer w.mu.RUnlock()
	return p.Sub(w.screenOrigin)
}

// IsMounted reports whether the widget is currently in the mounted tree.
func (w *WidgetBase) IsMounted() bool {
	w.mu.RLock()
	defer w.mu.RUnlock()
	return w.mounted
}

// SetMounted sets the widget's mounted state.
// This is called by the framework during mount/unmount tree walks.
func (w *WidgetBase) SetMounted(m bool) {
	w.mu.Lock()
	defer w.mu.Unlock()
	w.mounted = m
}

// setDetached and isDetached implement detachMarker (see IsDetached). They
// stay unexported: only the lifecycle walks in this package set the mark.
func (w *WidgetBase) setDetached(d bool) {
	w.mu.Lock()
	defer w.mu.Unlock()
	w.detached = d
}

func (w *WidgetBase) isDetached() bool {
	w.mu.RLock()
	defer w.mu.RUnlock()
	return w.detached
}

// SetSelf registers the concrete container widget that owns this base.
//
// WidgetBase cannot know its enclosing type: a container embeds it, so a
// method promoted from the base only has the *WidgetBase. Containers that own
// a plain membership child list ([primitives.Box], [overlay.Container]) must
// therefore call SetSelf once in their constructor so the mutators can set the
// correct parent on attached children.
//
// When self is nil (SetSelf was never called), the mutators still update the
// child list but return from the attach step without setting a parent or
// mounting — attaching to a nil parent would be wrong, and there is no concrete
// widget to pass. This mirrors Qt/GTK asserting a valid parent at insert; we
// degrade quietly instead of panicking so a bare WidgetBase remains usable.
//
// Framework plumbing, not application API: it is exported only because a
// container in another package (primitives, overlay) embeds WidgetBase and must
// register itself from its own constructor. An application never calls it.
func (w *WidgetBase) SetSelf(self Widget) {
	w.mu.Lock()
	defer w.mu.Unlock()
	w.self = self
}

// setContext stores the mount context for this widget.
//
// MountTree calls it when the subtree is mounted, so a container can mount a
// child added at runtime (see attachChild); UnmountTree clears it. It stays
// unexported: only this package calls it, through a type assertion, and an
// embedding widget in another package gets the promoted method as part of its
// method set.
func (w *WidgetBase) setContext(ctx Context) {
	w.mu.Lock()
	defer w.mu.Unlock()
	w.ctx = ctx
}

// NeedsRedraw reports whether the widget needs re-rendering.
//
// This flag is set by the signal scheduler when a bound signal changes,
// and cleared after the widget is drawn. It persists across scheduler
// flushes, surviving until the actual draw pass processes it.
//
// This is part of the retained-mode rendering system: widgets marked
// as needing redraw will trigger a full draw pass, while a tree with
// no dirty widgets allows the frame to skip rendering entirely.
func (w *WidgetBase) NeedsRedraw() bool {
	w.mu.RLock()
	defer w.mu.RUnlock()
	return w.needsRedraw
}

// SetNeedsRedraw marks the widget as needing re-rendering.
//
// When v is true, this also propagates the dirty flag upward through the
// parent chain to the nearest RepaintBoundary (ADR-007 upward dirty
// propagation). This is O(depth) rather than the O(n) downward tree walk
// of [NeedsRedrawInTree]. If the widget is already marked dirty, the
// upward propagation is skipped (O(1) guard).
//
// When v is false, only the local flag is cleared (no propagation needed).
//
// This is called by the signal scheduler's flush callback when a widget's
// bound signal has changed. Unlike the scheduler's pending set (which is
// cleared on flush), this flag persists until the draw pass clears it
// via [WidgetBase.ClearRedraw].
func (w *WidgetBase) SetNeedsRedraw(v bool) {
	w.mu.Lock()
	alreadyDirty := w.needsRedraw
	w.needsRedraw = v
	isBoundary := w.isRepaintBoundary
	parent := w.parent
	w.mu.Unlock()

	// Propagate upward only when setting dirty, and only if not already dirty
	// (O(1) guard prevents redundant walks).
	if v && !alreadyDirty {
		// Flutter markNeedsPaint: if THIS widget is a RepaintBoundary,
		// invalidate its own scene and stop — don't propagate to parent.
		// This is critical for animated widgets (spinner): dirty stays
		// at the spinner's boundary, parent tree stays clean.
		if isBoundary {
			w.InvalidateScene()
			return
		}
		propagateDirtyUpward(parent)
	}
}

// propagateDirtyUpward walks the parent chain from the given widget upward,
// marking each ancestor as dirty until a RepaintBoundary is found. When a
// RepaintBoundary is reached, it is marked dirty and propagation stops —
// this is the Flutter markNeedsPaint pattern (ADR-007).
//
// A widget is considered a repaint boundary if:
//   - It has IsRepaintBoundary() == true (ADR-024 WidgetBase property), OR
//   - It implements RepaintBoundaryMarker (legacy primitives.RepaintBoundary wrapper).
//
// If no RepaintBoundary is found, propagation reaches the root (which is
// correct — the root boundary encompasses the entire window).
func propagateDirtyUpward(w Widget) {
	for w != nil {
		// Check ADR-024 property first: WidgetBase.isRepaintBoundary.
		type boundaryPropChecker interface {
			IsRepaintBoundary() bool
			InvalidateScene()
		}
		if bp, ok := w.(boundaryPropChecker); ok && bp.IsRepaintBoundary() {
			bp.InvalidateScene()
			return
		}

		// Legacy check: primitives.RepaintBoundary implements RepaintBoundaryMarker
		// with its own MarkBoundaryDirty() override.
		if rb, ok := w.(RepaintBoundaryMarker); ok {
			rb.MarkBoundaryDirty()
			return
		}

		// Walk to next parent — do NOT mark intermediate ancestors dirty.
		// Flutter markNeedsPaint: only the boundary gets marked, intermediates
		// stay clean. Marking intermediates causes CollectDirtyRegions to
		// report the entire parent chain → full screen damage overlay.
		if pg, ok := w.(interface{ Parent() Widget }); ok {
			w = pg.Parent()
		} else {
			return
		}
	}
}

// markDirtyLocal sets the needsRedraw flag without triggering upward
// propagation. Used internally during upward dirty walks to avoid
// infinite recursion.
func (w *WidgetBase) markDirtyLocal() {
	w.mu.Lock()
	defer w.mu.Unlock()
	w.needsRedraw = true
}

// MarkRedrawLocal sets the needsRedraw flag on this widget only, without
// propagating upward through the parent chain. Use this for widgets with
// continuous animations (spinner, progress bar) where only the widget's
// own bounds should be in the dirty region, not the entire parent subtree.
func (w *WidgetBase) MarkRedrawLocal() {
	w.markDirtyLocal()
}

// ClearRedraw clears the widget's needsRedraw flag after a successful draw.
//
// This should be called by the rendering system after the widget has been
// drawn, to indicate that its visual state is now up to date.
func (w *WidgetBase) ClearRedraw() {
	w.mu.Lock()
	defer w.mu.Unlock()
	w.needsRedraw = false
}

// AddBinding registers a signal binding for automatic cleanup on unmount.
func (w *WidgetBase) AddBinding(b Unbinder) {
	if b == nil {
		return
	}
	w.mu.Lock()
	defer w.mu.Unlock()
	w.bindings = append(w.bindings, b)
}

// AddEffect registers an effect for automatic cleanup on unmount.
func (w *WidgetBase) AddEffect(e Stopper) {
	if e == nil {
		return
	}
	w.mu.Lock()
	defer w.mu.Unlock()
	w.effects = append(w.effects, e)
}

// CleanupBindings unbinds all signal bindings and stops all effects.
// Called automatically by the framework before Unmount().
func (w *WidgetBase) CleanupBindings() {
	w.mu.Lock()
	bindings := w.bindings
	effects := w.effects
	w.bindings = nil
	w.effects = nil
	w.mu.Unlock()

	for _, b := range bindings {
		b.Unbind()
	}
	for _, e := range effects {
		e.Stop()
	}
}
