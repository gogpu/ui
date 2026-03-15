// Package treeview provides a hierarchical tree widget for displaying
// nested data structures such as file explorers, org charts, and
// configuration trees.
//
// Construction uses functional options for immutable configuration:
//
//	root := &treeview.TreeNode{
//	    ID: "root", Label: "Root",
//	    Children: []*treeview.TreeNode{
//	        {ID: "child1", Label: "Child 1"},
//	        {ID: "child2", Label: "Child 2", Children: []*treeview.TreeNode{
//	            {ID: "grandchild", Label: "Grandchild"},
//	        }},
//	    },
//	}
//
//	tree := treeview.New(
//	    treeview.Root(root),
//	    treeview.ItemHeight(28),
//	    treeview.IndentWidth(20),
//	    treeview.ShowLines(true),
//	    treeview.SelectionModeOpt(treeview.SelectionSingle),
//	    treeview.OnSelect(func(node *treeview.TreeNode) { ... }),
//	    treeview.OnToggle(func(node *treeview.TreeNode, expanded bool) { ... }),
//	)
//
// # Tree Structure
//
// The tree is built from [TreeNode] values. Each node has an ID, label,
// optional children, and an Expanded flag controlling whether its children
// are visible. The Data field carries arbitrary user payloads.
//
// # Virtualization
//
// The tree flattens expanded nodes into a visible row list and renders
// only the rows within the viewport, enabling efficient display of large
// hierarchies. The flatten operation runs incrementally whenever the
// expanded state changes.
//
// # Keyboard Navigation
//
// When focused, the tree supports full keyboard navigation:
//   - Up/Down: move selection between visible rows
//   - Left: collapse current node (or move to parent if leaf/collapsed)
//   - Right: expand current node (or move to first child if expanded)
//   - Enter/Space: activate OnSelect callback for current node
//   - Home/End: jump to first/last visible row
//
// # Signal Binding
//
// Tree properties can be bound to reactive signals from the [state] package.
//   - [SelectedNodeSignal] -- TWO-WAY binding for the selected node ID
//   - [RootSignal] -- one-way binding for the root node (data refresh)
//
// # Visual Style
//
// The visual rendering (row backgrounds, selection highlights, expand
// icons, connector lines) is provided by a [Painter] implementation.
// Each design system supplies its own painter.
//
// If no painter is set, [DefaultPainter] is used, which draws minimal
// visuals suitable for testing and prototyping.
//
// # Accessibility
//
// TreeView implements [a11y.Accessible] with [a11y.RoleTree]. Keyboard
// navigation with arrow keys moves selection between visible nodes.
//
// # Focus
//
// TreeView implements [widget.Focusable] and participates in tab navigation.
// When focused, arrow keys control selection and expand/collapse state.
package treeview
