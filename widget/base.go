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
// If the child has a WidgetBase that can be accessed, its parent is set
// to this widget.
func (w *WidgetBase) AddChild(child Widget) {
	if child == nil {
		return
	}
	w.mu.Lock()
	defer w.mu.Unlock()
	w.children = append(w.children, child)
	// Try to set parent if child supports it
	if setter, ok := child.(interface{ SetParent(Widget) }); ok {
		// Note: We can't pass w here because we only have *WidgetBase, not the containing widget
		// The parent should be set by the containing widget type if needed
		_ = setter // Avoid unused variable
	}
}

// RemoveChild removes a child widget.
//
// Returns true if the child was found and removed.
func (w *WidgetBase) RemoveChild(child Widget) bool {
	if child == nil {
		return false
	}
	w.mu.Lock()
	defer w.mu.Unlock()
	for i, c := range w.children {
		if c != child {
			continue
		}
		// Remove by replacing with last element and truncating
		lastIdx := len(w.children) - 1
		w.children[i] = w.children[lastIdx]
		w.children[lastIdx] = nil // Clear reference for GC
		w.children = w.children[:lastIdx]
		return true
	}
	return false
}

// RemoveChildAt removes the child at the given index.
//
// Returns the removed child, or nil if the index is out of range.
func (w *WidgetBase) RemoveChildAt(index int) Widget {
	w.mu.Lock()
	defer w.mu.Unlock()
	if index < 0 || index >= len(w.children) {
		return nil
	}
	child := w.children[index]
	// Remove while preserving order
	copy(w.children[index:], w.children[index+1:])
	w.children[len(w.children)-1] = nil // Clear reference for GC
	w.children = w.children[:len(w.children)-1]
	return child
}

// ClearChildren removes all child widgets.
func (w *WidgetBase) ClearChildren() {
	w.mu.Lock()
	defer w.mu.Unlock()
	// Clear references for GC
	for i := range w.children {
		w.children[i] = nil
	}
	w.children = w.children[:0]
}

// InsertChild inserts a child widget at the given index.
//
// If index is out of range, the child is appended.
func (w *WidgetBase) InsertChild(index int, child Widget) {
	if child == nil {
		return
	}
	w.mu.Lock()
	defer w.mu.Unlock()
	if index < 0 {
		index = 0
	}
	if index >= len(w.children) {
		w.children = append(w.children, child)
		return
	}
	// Insert at index
	w.children = append(w.children, nil)
	copy(w.children[index+1:], w.children[index:])
	w.children[index] = child
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
