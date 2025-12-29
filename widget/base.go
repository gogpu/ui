package widget

import (
	"sync"

	"github.com/gogpu/ui/geometry"
)

// WidgetBase provides common functionality for widgets.
//
// Embed this struct in custom widget implementations to get:
//   - Bounds tracking (position and size)
//   - Focus state management
//   - Visibility control
//   - Enabled/disabled state
//   - Child widget management
//   - Optional ID for debugging
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
	mu       sync.RWMutex
	bounds   geometry.Rect // Cached layout bounds
	focused  bool          // Whether widget has focus
	visible  bool          // Whether widget is visible
	enabled  bool          // Whether widget accepts input
	id       string        // Optional ID for debugging
	children []Widget      // Child widgets
	parent   Widget        // Parent widget (if any)
}

// NewWidgetBase creates a new WidgetBase with default settings.
//
// The widget is visible and enabled by default, with no children
// and zero bounds.
func NewWidgetBase() *WidgetBase {
	return &WidgetBase{
		visible: true,
		enabled: true,
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

// LocalToGlobal converts a point from local coordinates to global (window) coordinates.
//
// Local coordinates are relative to the widget's top-left corner.
// Global coordinates are relative to the window's top-left corner.
func (w *WidgetBase) LocalToGlobal(p geometry.Point) geometry.Point {
	w.mu.RLock()
	defer w.mu.RUnlock()
	return p.Add(w.bounds.Min)
}

// GlobalToLocal converts a point from global (window) coordinates to local coordinates.
//
// Local coordinates are relative to the widget's top-left corner.
// Global coordinates are relative to the window's top-left corner.
func (w *WidgetBase) GlobalToLocal(p geometry.Point) geometry.Point {
	w.mu.RLock()
	defer w.mu.RUnlock()
	return p.Sub(w.bounds.Min)
}
