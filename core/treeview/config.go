package treeview

import (
	"github.com/gogpu/ui/state"
)

// config holds the tree view's configuration, set at construction time via options.
type config struct {
	// Tree data.
	root               *TreeNode
	rootSignal         state.Signal[*TreeNode]
	readonlyRootSignal state.ReadonlySignal[*TreeNode]

	// Layout.
	itemHeight  float32
	indentWidth float32
	showLines   bool

	// Selection.
	selectionMode              SelectionMode
	selectedNodeID             string
	selectedNodeIDSignal       state.Signal[string]
	readonlySelectedNodeSignal state.ReadonlySignal[string]
	onSelect                   func(node *TreeNode)
	onToggle                   func(node *TreeNode, expanded bool)

	// Disabled state.
	disabled               bool
	disabledFn             func() bool
	disabledSignal         state.Signal[bool]
	readonlyDisabledSignal state.ReadonlySignal[bool]

	// Painter.
	painter Painter

	// Accessibility.
	a11yLabel string
}

// Default configuration values.
const (
	defaultItemHeight  float32 = 28
	defaultIndentWidth float32 = 20
)

// defaultConfig returns a config with sensible defaults.
func defaultConfig() config {
	return config{
		itemHeight:  defaultItemHeight,
		indentWidth: defaultIndentWidth,
	}
}

// ResolvedRoot returns the current root node.
// Priority: ReadonlySignal > Signal > Static.
func (c *config) ResolvedRoot() *TreeNode {
	if c.readonlyRootSignal != nil {
		return c.readonlyRootSignal.Get()
	}
	if c.rootSignal != nil {
		return c.rootSignal.Get()
	}
	return c.root
}

// ResolvedSelectedNodeID returns the current selected node ID.
// Priority: ReadonlySignal > Signal > Static.
// Returns "" if no selection.
func (c *config) ResolvedSelectedNodeID() string {
	if c.readonlySelectedNodeSignal != nil {
		return c.readonlySelectedNodeSignal.Get()
	}
	if c.selectedNodeIDSignal != nil {
		return c.selectedNodeIDSignal.Get()
	}
	return c.selectedNodeID
}

// ResolvedDisabled returns the current disabled state.
// Priority: ReadonlySignal > Signal > Fn > Static.
func (c *config) ResolvedDisabled() bool {
	if c.readonlyDisabledSignal != nil {
		return c.readonlyDisabledSignal.Get()
	}
	if c.disabledSignal != nil {
		return c.disabledSignal.Get()
	}
	if c.disabledFn != nil {
		return c.disabledFn()
	}
	return c.disabled
}
