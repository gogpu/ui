package a11y

import "sync"

// TreeProvider is the interface for managing the accessibility tree.
//
// The accessibility tree is a hierarchical representation of all accessible
// elements in the UI. Platform adapters consume this tree to drive native
// accessibility APIs (Windows UI Automation, macOS NSAccessibility, Linux AT-SPI2).
//
// Implementations must be safe for concurrent access. The default implementation
// [MemoryTree] uses a sync.RWMutex.
//
// # Tree Lifecycle
//
//  1. Create a tree with a root node
//  2. As widgets are created, insert nodes with [TreeProvider.Insert]
//  3. As widgets are updated, call [TreeProvider.Update] to mark changes
//  4. As widgets are removed, call [TreeProvider.Remove]
//  5. Platform adapters walk the tree with [TreeProvider.Walk]
//
// # Example
//
//	tree := a11y.NewMemoryTree(rootNode)
//	tree.Insert(rootNode, childNode)
//	tree.Walk(func(n *a11y.Node) bool {
//	    fmt.Printf("%s: %s\n", n.Role(), n.Label())
//	    return true // continue walking
//	})
type TreeProvider interface {
	// Root returns the root node of the accessibility tree.
	//
	// The root node typically represents the application window.
	// Returns nil if the tree is empty.
	Root() *Node

	// NodeByID looks up a node by its unique identifier.
	//
	// Returns nil if no node with the given ID exists in the tree.
	NodeByID(id NodeID) *Node

	// Update marks a node as having changed, so platform adapters can
	// send appropriate change notifications to assistive technology.
	//
	// This does not modify the node's properties; it only flags the node
	// for the next sync cycle. If the node has a source [Accessible],
	// callers should call [Node.SyncFromSource] before or after Update.
	Update(node *Node)

	// Insert adds a child node under the given parent.
	//
	// The child is appended to the parent's children list and registered
	// in the tree's ID index. If the child already has a parent, it is
	// not re-parented automatically; call [TreeProvider.Remove] first.
	Insert(parent *Node, child *Node)

	// Remove removes a node and all its descendants from the tree.
	//
	// The node is detached from its parent and unregistered from the
	// tree's ID index. Removing the root node empties the tree.
	Remove(node *Node)

	// Walk performs a depth-first traversal of the tree, calling fn
	// for each node. If fn returns false, the traversal stops immediately.
	//
	// The traversal visits the parent before its children.
	Walk(fn func(*Node) bool)

	// Len returns the total number of nodes in the tree.
	Len() int

	// DirtyNodes returns the set of nodes that have been marked as changed
	// since the last call to ClearDirty.
	//
	// The returned slice is a copy and safe to modify.
	DirtyNodes() []*Node

	// ClearDirty clears the set of dirty nodes.
	//
	// This is called by platform adapters after processing all pending
	// change notifications.
	ClearDirty()
}

// MemoryTree is an in-memory implementation of [TreeProvider].
//
// It stores the full accessibility tree with an index for O(1) node lookups
// by ID. It tracks dirty (changed) nodes for efficient platform synchronization.
//
// Thread Safety:
//
// MemoryTree is safe for concurrent access. All operations are protected
// by a sync.RWMutex.
type MemoryTree struct {
	mu    sync.RWMutex
	root  *Node
	index map[NodeID]*Node
	dirty map[NodeID]*Node
}

// NewMemoryTree creates a new in-memory accessibility tree with the given root node.
//
// The root node is registered in the tree's ID index. Pass nil to create
// an empty tree (Root will return nil until a root is inserted).
//
// Example:
//
//	root := a11y.NewNode(a11y.RoleWindow, "My Application")
//	tree := a11y.NewMemoryTree(root)
func NewMemoryTree(root *Node) *MemoryTree {
	t := &MemoryTree{
		index: make(map[NodeID]*Node),
		dirty: make(map[NodeID]*Node),
	}
	if root != nil {
		t.root = root
		t.index[root.ID()] = root
	}
	return t
}

// Root returns the root node of the tree, or nil if the tree is empty.
func (t *MemoryTree) Root() *Node {
	t.mu.RLock()
	defer t.mu.RUnlock()
	return t.root
}

// NodeByID looks up a node by its unique identifier.
//
// Returns nil if no node with the given ID exists.
func (t *MemoryTree) NodeByID(id NodeID) *Node {
	t.mu.RLock()
	defer t.mu.RUnlock()
	return t.index[id]
}

// Update marks a node as changed for the next platform sync cycle.
func (t *MemoryTree) Update(node *Node) {
	if node == nil {
		return
	}
	t.mu.Lock()
	defer t.mu.Unlock()
	t.dirty[node.ID()] = node
}

// Insert adds a child node under the given parent.
//
// Both parent and child must be non-nil. The child is appended to the
// parent's children list and registered in the ID index.
func (t *MemoryTree) Insert(parent *Node, child *Node) {
	if parent == nil || child == nil {
		return
	}
	t.mu.Lock()
	defer t.mu.Unlock()

	parent.mu.Lock()
	child.mu.Lock()
	parent.addChild(child)
	child.mu.Unlock()
	parent.mu.Unlock()

	t.registerSubtree(child)
}

// Remove removes a node and all its descendants from the tree.
//
// If the removed node is the root, the tree becomes empty.
func (t *MemoryTree) Remove(node *Node) {
	if node == nil {
		return
	}
	t.mu.Lock()
	defer t.mu.Unlock()

	// Detach from parent
	node.mu.RLock()
	parent := node.parent
	node.mu.RUnlock()

	if parent != nil {
		parent.mu.Lock()
		parent.removeChild(node)
		parent.mu.Unlock()
	}

	// Unregister subtree from index
	t.unregisterSubtree(node)

	// Clear root if it was removed
	if t.root == node {
		t.root = nil
	}
}

// Walk performs a depth-first traversal of the tree.
//
// If fn returns false, the traversal stops immediately.
func (t *MemoryTree) Walk(fn func(*Node) bool) {
	t.mu.RLock()
	root := t.root
	t.mu.RUnlock()
	if root == nil {
		return
	}
	walkNode(root, fn)
}

// Len returns the total number of nodes in the tree.
func (t *MemoryTree) Len() int {
	t.mu.RLock()
	defer t.mu.RUnlock()
	return len(t.index)
}

// DirtyNodes returns the set of nodes marked as changed since the last
// call to ClearDirty.
func (t *MemoryTree) DirtyNodes() []*Node {
	t.mu.RLock()
	defer t.mu.RUnlock()
	if len(t.dirty) == 0 {
		return nil
	}
	result := make([]*Node, 0, len(t.dirty))
	for _, n := range t.dirty {
		result = append(result, n)
	}
	return result
}

// ClearDirty clears the set of dirty nodes.
func (t *MemoryTree) ClearDirty() {
	t.mu.Lock()
	defer t.mu.Unlock()
	t.dirty = make(map[NodeID]*Node)
}

// registerSubtree adds a node and all its descendants to the index.
// Caller must hold write lock on the tree.
func (t *MemoryTree) registerSubtree(node *Node) {
	t.index[node.ID()] = node
	node.mu.RLock()
	children := node.children
	node.mu.RUnlock()
	for _, child := range children {
		t.registerSubtree(child)
	}
}

// unregisterSubtree removes a node and all its descendants from the index.
// Caller must hold write lock on the tree.
func (t *MemoryTree) unregisterSubtree(node *Node) {
	delete(t.index, node.ID())
	delete(t.dirty, node.ID())
	node.mu.RLock()
	children := node.children
	node.mu.RUnlock()
	for _, child := range children {
		t.unregisterSubtree(child)
	}
}

// walkNode performs depth-first traversal starting from the given node.
// Returns false if the walk was stopped early.
func walkNode(node *Node, fn func(*Node) bool) bool {
	if !fn(node) {
		return false
	}
	node.mu.RLock()
	children := make([]*Node, len(node.children))
	copy(children, node.children)
	node.mu.RUnlock()

	for _, child := range children {
		if !walkNode(child, fn) {
			return false
		}
	}
	return true
}

// Verify MemoryTree implements TreeProvider.
var _ TreeProvider = (*MemoryTree)(nil)
