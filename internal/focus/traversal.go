package focus

import "github.com/gogpu/ui/widget"

// collectFocusable performs a depth-first traversal of the widget tree
// starting from root and collects all widgets that implement [widget.Focusable]
// and are currently eligible for focus (visible, enabled, and focusable).
func collectFocusable(root widget.Widget) []widget.Focusable {
	if root == nil {
		return nil
	}
	var result []widget.Focusable
	collectFocusableRecursive(root, &result)
	return result
}

// collectFocusableRecursive is the recursive helper for collectFocusable.
func collectFocusableRecursive(w widget.Widget, result *[]widget.Focusable) {
	if w == nil {
		return
	}

	// Skip invisible widgets and their subtrees.
	if vis, ok := w.(interface{ IsVisible() bool }); ok && !vis.IsVisible() {
		return
	}

	// Skip disabled widgets and their subtrees.
	if en, ok := w.(interface{ IsEnabled() bool }); ok && !en.IsEnabled() {
		return
	}

	// Check if this widget is focusable.
	if f, ok := w.(widget.Focusable); ok && f.IsFocusable() {
		*result = append(*result, f)
	}

	// Recurse into children.
	for _, child := range w.Children() {
		collectFocusableRecursive(child, result)
	}
}

// indexOf returns the index of target in the focusable list, or -1 if not found.
func indexOf(list []widget.Focusable, target widget.Focusable) int {
	for i, f := range list {
		if f == target {
			return i
		}
	}
	return -1
}
