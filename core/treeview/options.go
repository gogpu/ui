package treeview

import (
	"github.com/gogpu/ui/state"
)

// Option configures a tree view during construction.
type Option func(*config)

// --- Tree Data ---

// Root sets the root node of the tree.
func Root(root *TreeNode) Option {
	return func(c *config) {
		c.root = root
	}
}

// RootSignal binds the root node to a reactive signal.
// When set, the signal takes precedence over [Root]
// but not over [RootReadonlySignal].
func RootSignal(sig state.Signal[*TreeNode]) Option {
	return func(c *config) {
		c.rootSignal = sig
	}
}

// RootReadonlySignal binds the root node to a read-only signal.
// When set, this takes highest precedence over all other root sources.
func RootReadonlySignal(sig state.ReadonlySignal[*TreeNode]) Option {
	return func(c *config) {
		c.readonlyRootSignal = sig
	}
}

// --- Layout ---

// ItemHeight sets the fixed height for each row in the tree (pixels).
// Default: 28.
func ItemHeight(h float32) Option {
	return func(c *config) {
		c.itemHeight = h
	}
}

// IndentWidth sets the horizontal offset per nesting level (pixels).
// Default: 20.
func IndentWidth(w float32) Option {
	return func(c *config) {
		c.indentWidth = w
	}
}

// ShowLines enables connector lines between parent and child nodes.
func ShowLines(enabled bool) Option {
	return func(c *config) {
		c.showLines = enabled
	}
}

// --- Selection ---

// SelectionModeOpt sets the selection mode for the tree.
// Default is [SelectionNone].
func SelectionModeOpt(mode SelectionMode) Option {
	return func(c *config) {
		c.selectionMode = mode
	}
}

// SelectedNodeID sets the initially selected node by ID.
// Use "" for no selection.
func SelectedNodeID(id string) Option {
	return func(c *config) {
		c.selectedNodeID = id
	}
}

// SelectedNodeSignal binds the selected node ID to a reactive signal.
// This is a TWO-WAY binding: the widget reads from the signal, and
// when the user selects a node, the new ID is written back.
func SelectedNodeSignal(sig state.Signal[string]) Option {
	return func(c *config) {
		c.selectedNodeIDSignal = sig
	}
}

// SelectedNodeReadonlySignal binds the selected node ID to a read-only signal.
// When set, this takes highest precedence over all other selection sources.
func SelectedNodeReadonlySignal(sig state.ReadonlySignal[string]) Option {
	return func(c *config) {
		c.readonlySelectedNodeSignal = sig
	}
}

// OnSelect sets the callback invoked when a node is activated (clicked or Enter).
func OnSelect(fn func(node *TreeNode)) Option {
	return func(c *config) {
		c.onSelect = fn
	}
}

// OnToggle sets the callback invoked when a node's expanded state changes.
func OnToggle(fn func(node *TreeNode, expanded bool)) Option {
	return func(c *config) {
		c.onToggle = fn
	}
}

// --- Disabled State ---

// Disabled sets the tree view's disabled state.
func Disabled(d bool) Option {
	return func(c *config) {
		c.disabled = d
	}
}

// DisabledFn sets a dynamic function for the disabled state.
// When set, this takes precedence over [Disabled] but not over [DisabledSignal].
func DisabledFn(fn func() bool) Option {
	return func(c *config) {
		c.disabledFn = fn
	}
}

// DisabledSignal binds the disabled state to a reactive signal.
// When set, the signal takes precedence over [DisabledFn] and [Disabled]
// but not over [DisabledReadonlySignal].
func DisabledSignal(sig state.Signal[bool]) Option {
	return func(c *config) {
		c.disabledSignal = sig
	}
}

// DisabledReadonlySignal binds the disabled state to a read-only signal.
// When set, this takes highest precedence over all other disabled sources.
func DisabledReadonlySignal(sig state.ReadonlySignal[bool]) Option {
	return func(c *config) {
		c.readonlyDisabledSignal = sig
	}
}

// --- Visual Options ---

// PainterOpt sets the painter used to render tree-specific visuals.
// Each design system provides its own painter. If not set,
// [DefaultPainter] is used.
func PainterOpt(p Painter) Option {
	return func(c *config) {
		c.painter = p
	}
}

// --- Accessibility ---

// A11yLabel sets the accessibility label for the tree.
func A11yLabel(label string) Option {
	return func(c *config) {
		c.a11yLabel = label
	}
}
