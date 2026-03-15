package treeview

import (
	"image"
	"testing"

	"github.com/gogpu/ui/a11y"
	"github.com/gogpu/ui/event"
	"github.com/gogpu/ui/geometry"
	"github.com/gogpu/ui/state"
	"github.com/gogpu/ui/widget"
)

// --- Test Helpers ---

// makeTestTree creates a standard test tree:
//
//	root
//	  ├── child1
//	  │   ├── grandchild1
//	  │   └── grandchild2
//	  └── child2
//	      └── grandchild3
func makeTestTree() *TreeNode {
	return &TreeNode{
		ID:       "root",
		Label:    "Root",
		Expanded: true,
		Children: []*TreeNode{
			{
				ID:       "child1",
				Label:    "Child 1",
				Expanded: true,
				Children: []*TreeNode{
					{ID: "gc1", Label: "Grandchild 1"},
					{ID: "gc2", Label: "Grandchild 2"},
				},
			},
			{
				ID:       "child2",
				Label:    "Child 2",
				Expanded: false,
				Children: []*TreeNode{
					{ID: "gc3", Label: "Grandchild 3"},
				},
			},
		},
	}
}

func makeCtx() *widget.ContextImpl {
	return widget.NewContext()
}

func layoutAndBounds(w *Widget, ctx widget.Context, width, height float32) {
	c := geometry.Tight(geometry.Sz(width, height))
	size := w.Layout(ctx, c)
	w.SetBounds(geometry.NewRect(0, 0, size.Width, size.Height))
}

// --- TreeNode Tests ---

func TestTreeNode_IsLeaf(t *testing.T) {
	tests := []struct {
		name string
		node *TreeNode
		want bool
	}{
		{"nil children", &TreeNode{ID: "a"}, true},
		{"empty children", &TreeNode{ID: "a", Children: []*TreeNode{}}, true},
		{"has children", &TreeNode{ID: "a", Children: []*TreeNode{{ID: "b"}}}, false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := tt.node.IsLeaf(); got != tt.want {
				t.Errorf("IsLeaf() = %v, want %v", got, tt.want)
			}
		})
	}
}

// --- FlattenTree Tests ---

func TestFlattenTree_NilRoot(t *testing.T) {
	rows := flattenTree(nil)
	if rows != nil {
		t.Errorf("flattenTree(nil) = %v, want nil", rows)
	}
}

func TestFlattenTree_SingleNode(t *testing.T) {
	root := &TreeNode{ID: "root", Label: "Root"}
	rows := flattenTree(root)
	if len(rows) != 1 {
		t.Fatalf("expected 1 row, got %d", len(rows))
	}
	if rows[0].node.ID != "root" || rows[0].depth != 0 {
		t.Errorf("unexpected row: %+v", rows[0])
	}
}

func TestFlattenTree_ExpandedChildren(t *testing.T) {
	root := makeTestTree()
	rows := flattenTree(root)

	// root(expanded) -> child1(expanded) -> gc1, gc2, child2(collapsed)
	// Expected: root, child1, gc1, gc2, child2 = 5 rows
	if len(rows) != 5 {
		t.Fatalf("expected 5 rows, got %d", len(rows))
	}

	expected := []struct {
		id    string
		depth int
	}{
		{"root", 0},
		{"child1", 1},
		{"gc1", 2},
		{"gc2", 2},
		{"child2", 1},
	}

	for i, exp := range expected {
		if rows[i].node.ID != exp.id || rows[i].depth != exp.depth {
			t.Errorf("row[%d]: got id=%s depth=%d, want id=%s depth=%d",
				i, rows[i].node.ID, rows[i].depth, exp.id, exp.depth)
		}
	}
}

func TestFlattenTree_CollapsedHidesChildren(t *testing.T) {
	root := &TreeNode{
		ID: "root", Label: "Root", Expanded: false,
		Children: []*TreeNode{
			{ID: "a", Label: "A"},
			{ID: "b", Label: "B"},
		},
	}
	rows := flattenTree(root)
	if len(rows) != 1 {
		t.Fatalf("collapsed root should show 1 row, got %d", len(rows))
	}
}

// --- FindNodeByID Tests ---

func TestFindNodeByID(t *testing.T) {
	root := makeTestTree()

	tests := []struct {
		id    string
		found bool
	}{
		{"root", true},
		{"child1", true},
		{"gc1", true},
		{"gc3", true},
		{"nonexistent", false},
	}

	for _, tt := range tests {
		t.Run(tt.id, func(t *testing.T) {
			node := findNodeByID(root, tt.id)
			if (node != nil) != tt.found {
				t.Errorf("findNodeByID(%q) found=%v, want %v", tt.id, node != nil, tt.found)
			}
		})
	}
}

func TestFindNodeByID_NilRoot(t *testing.T) {
	if node := findNodeByID(nil, "x"); node != nil {
		t.Error("expected nil for nil root")
	}
}

// --- FindParent Tests ---

func TestFindParent(t *testing.T) {
	root := makeTestTree()

	tests := []struct {
		childID  string
		parentID string
	}{
		{"child1", "root"},
		{"child2", "root"},
		{"gc1", "child1"},
		{"gc2", "child1"},
		{"gc3", "child2"},
	}

	for _, tt := range tests {
		t.Run(tt.childID, func(t *testing.T) {
			parent := findParent(root, tt.childID)
			if parent == nil {
				t.Fatalf("expected parent for %q, got nil", tt.childID)
			}
			if parent.ID != tt.parentID {
				t.Errorf("parent of %q = %q, want %q", tt.childID, parent.ID, tt.parentID)
			}
		})
	}
}

func TestFindParent_Root(t *testing.T) {
	root := makeTestTree()
	parent := findParent(root, "root")
	if parent != nil {
		t.Error("root should have no parent")
	}
}

func TestFindParent_NilRoot(t *testing.T) {
	if parent := findParent(nil, "x"); parent != nil {
		t.Error("expected nil for nil root")
	}
}

// --- SelectionMode Tests ---

func TestSelectionMode_String(t *testing.T) {
	tests := []struct {
		mode SelectionMode
		want string
	}{
		{SelectionNone, "None"},
		{SelectionSingle, "Single"},
		{SelectionMode(99), "Unknown"},
	}
	for _, tt := range tests {
		t.Run(tt.want, func(t *testing.T) {
			if got := tt.mode.String(); got != tt.want {
				t.Errorf("String() = %q, want %q", got, tt.want)
			}
		})
	}
}

// --- Widget Construction Tests ---

func TestNew_Defaults(t *testing.T) {
	w := New()

	if !w.IsVisible() {
		t.Error("expected visible by default")
	}
	if !w.IsEnabled() {
		t.Error("expected enabled by default")
	}
	if w.cfg.itemHeight != defaultItemHeight {
		t.Errorf("itemHeight = %v, want %v", w.cfg.itemHeight, defaultItemHeight)
	}
	if w.cfg.indentWidth != defaultIndentWidth {
		t.Errorf("indentWidth = %v, want %v", w.cfg.indentWidth, defaultIndentWidth)
	}
}

func TestNew_WithOptions(t *testing.T) {
	root := makeTestTree()
	var selectedNode *TreeNode
	var toggled bool

	w := New(
		Root(root),
		ItemHeight(32),
		IndentWidth(24),
		ShowLines(true),
		SelectionModeOpt(SelectionSingle),
		OnSelect(func(n *TreeNode) { selectedNode = n }),
		OnToggle(func(_ *TreeNode, exp bool) { toggled = exp }),
		A11yLabel("File tree"),
	)

	if w.cfg.itemHeight != 32 {
		t.Errorf("itemHeight = %v, want 32", w.cfg.itemHeight)
	}
	if w.cfg.indentWidth != 24 {
		t.Errorf("indentWidth = %v, want 24", w.cfg.indentWidth)
	}
	if !w.cfg.showLines {
		t.Error("expected showLines = true")
	}
	if w.cfg.selectionMode != SelectionSingle {
		t.Error("expected SelectionSingle")
	}
	if w.cfg.a11yLabel != "File tree" {
		t.Errorf("a11yLabel = %q, want %q", w.cfg.a11yLabel, "File tree")
	}

	// Verify rows are flattened.
	if len(w.rows) != 5 {
		t.Fatalf("expected 5 rows, got %d", len(w.rows))
	}

	// Verify callbacks are set (they won't be nil).
	_ = selectedNode
	_ = toggled
}

func TestNew_WithPainter(t *testing.T) {
	p := DefaultPainter{}
	w := New(PainterOpt(p))
	// Painter should be set (can't check type directly, but no panic).
	_ = w
}

func TestNew_NilRoot(t *testing.T) {
	w := New()
	if len(w.rows) != 0 {
		t.Errorf("expected 0 rows for nil root, got %d", len(w.rows))
	}
}

// --- IsFocusable Tests ---

func TestIsFocusable(t *testing.T) {
	w := New(Root(makeTestTree()))

	if !w.IsFocusable() {
		t.Error("expected focusable when visible and enabled")
	}

	w.SetVisible(false)
	if w.IsFocusable() {
		t.Error("expected not focusable when invisible")
	}

	w.SetVisible(true)
	w.SetEnabled(false)
	if w.IsFocusable() {
		t.Error("expected not focusable when disabled")
	}

	w.SetEnabled(true)
	w2 := New(Root(makeTestTree()), Disabled(true))
	if w2.IsFocusable() {
		t.Error("expected not focusable when config disabled")
	}
}

// --- Layout Tests ---

func TestLayout_TightConstraints(t *testing.T) {
	w := New(Root(makeTestTree()))
	ctx := makeCtx()

	c := geometry.Tight(geometry.Sz(400, 200))
	size := w.Layout(ctx, c)

	if size.Width != 400 || size.Height != 200 {
		t.Errorf("size = %v, want (400, 200)", size)
	}
}

func TestLayout_InfiniteConstraints(t *testing.T) {
	w := New(Root(makeTestTree()), ItemHeight(28))
	ctx := makeCtx()

	c := geometry.BoxConstraints(0, geometry.Infinity, 0, geometry.Infinity)
	size := w.Layout(ctx, c)

	if size.Width != defaultViewportWidth {
		t.Errorf("width = %v, want %v", size.Width, defaultViewportWidth)
	}
	// 5 rows * 28 = 140, which is less than defaultViewportHeight=400.
	expectedH := float32(5 * 28)
	if size.Height != expectedH {
		t.Errorf("height = %v, want %v", size.Height, expectedH)
	}
}

// --- Draw Tests ---

func TestDraw_EmptyTree(t *testing.T) {
	w := New()
	ctx := makeCtx()
	canvas := &mockCanvas{}

	layoutAndBounds(w, ctx, 300, 200)
	w.Draw(ctx, canvas)

	// Should call PaintEmptyState (which calls DrawText).
	if canvas.drawTextCalls == 0 {
		t.Error("expected DrawText for empty state")
	}
}

func TestDraw_WithRows(t *testing.T) {
	w := New(Root(makeTestTree()))
	ctx := makeCtx()
	canvas := &mockCanvas{}

	layoutAndBounds(w, ctx, 250, 400)
	w.Draw(ctx, canvas)

	// Should draw 5 rows of labels.
	if canvas.drawTextCalls < 5 {
		t.Errorf("expected at least 5 DrawText calls, got %d", canvas.drawTextCalls)
	}

	// Should push and pop clip.
	if canvas.pushClipCalls != 1 || canvas.popClipCalls != 1 {
		t.Errorf("clip: push=%d pop=%d, expected 1 each", canvas.pushClipCalls, canvas.popClipCalls)
	}
}

func TestDraw_InvisibleWidget(t *testing.T) {
	w := New(Root(makeTestTree()))
	w.SetVisible(false)
	ctx := makeCtx()
	canvas := &mockCanvas{}

	layoutAndBounds(w, ctx, 300, 200)
	w.Draw(ctx, canvas)

	if canvas.drawTextCalls != 0 {
		t.Error("invisible widget should not draw")
	}
}

func TestDraw_EmptyBounds(t *testing.T) {
	w := New(Root(makeTestTree()))
	ctx := makeCtx()
	canvas := &mockCanvas{}

	// Don't set bounds.
	w.Draw(ctx, canvas)

	if canvas.drawTextCalls != 0 {
		t.Error("empty bounds should not draw")
	}
}

func TestDraw_ShowLines(t *testing.T) {
	w := New(Root(makeTestTree()), ShowLines(true))
	ctx := makeCtx()
	canvas := &mockCanvas{}

	layoutAndBounds(w, ctx, 300, 400)
	w.Draw(ctx, canvas)

	// Connector lines use DrawLine.
	if canvas.drawLineCalls == 0 {
		t.Error("expected DrawLine calls for connector lines")
	}
}

func TestDraw_Selection(t *testing.T) {
	w := New(
		Root(makeTestTree()),
		SelectionModeOpt(SelectionSingle),
		SelectedNodeID("child1"),
	)
	ctx := makeCtx()
	canvas := &mockCanvas{}

	layoutAndBounds(w, ctx, 300, 400)
	w.Draw(ctx, canvas)

	// Selection highlight should draw a rect.
	if canvas.drawRectCalls == 0 {
		t.Error("expected DrawRect for selection highlight")
	}
}

// --- Event: Mouse ---

func TestMousePress_SelectsNode(t *testing.T) {
	var selected string
	w := New(
		Root(makeTestTree()),
		SelectionModeOpt(SelectionSingle),
		OnSelect(func(n *TreeNode) { selected = n.ID }),
	)
	ctx := makeCtx()
	layoutAndBounds(w, ctx, 300, 400)

	// Click on second row (child1, at y=28+14=42 center of row).
	e := &event.MouseEvent{
		MouseType: event.MousePress,
		Button:    event.ButtonLeft,
		Position:  geometry.Pt(150, 42),
	}
	consumed := w.Event(ctx, e)

	if !consumed {
		t.Error("expected click to be consumed")
	}
	if selected != "child1" {
		t.Errorf("selected = %q, want %q", selected, "child1")
	}
}

func TestMousePress_ToggleExpand(t *testing.T) {
	var toggledNode *TreeNode
	root := makeTestTree()
	w := New(
		Root(root),
		OnToggle(func(n *TreeNode, _ bool) { toggledNode = n }),
	)
	ctx := makeCtx()
	layoutAndBounds(w, ctx, 300, 400)

	// Click on expand icon of child1 (row index 1, depth=1).
	// Icon is at x = depth*indentWidth + iconSize/2 = 1*20 + 8 = 28.
	e := &event.MouseEvent{
		MouseType: event.MousePress,
		Button:    event.ButtonLeft,
		Position:  geometry.Pt(28, 42), // row 1 center
	}
	consumed := w.Event(ctx, e)

	if !consumed {
		t.Error("expected toggle click to be consumed")
	}
	if toggledNode == nil || toggledNode.ID != "child1" {
		t.Error("expected child1 to be toggled")
	}
}

func TestMouseMove_UpdatesHover(t *testing.T) {
	w := New(Root(makeTestTree()))
	ctx := makeCtx()
	layoutAndBounds(w, ctx, 300, 400)

	e := &event.MouseEvent{
		MouseType: event.MouseMove,
		Position:  geometry.Pt(150, 14), // row 0
	}
	w.Event(ctx, e)

	if w.hoveredIndex != 0 {
		t.Errorf("hoveredIndex = %d, want 0", w.hoveredIndex)
	}
}

func TestMouseLeave_ClearsHover(t *testing.T) {
	w := New(Root(makeTestTree()))
	ctx := makeCtx()
	layoutAndBounds(w, ctx, 300, 400)

	// First hover.
	w.hoveredIndex = 2
	e := &event.MouseEvent{MouseType: event.MouseLeave}
	w.Event(ctx, e)

	if w.hoveredIndex != noHoveredIndex {
		t.Errorf("hoveredIndex = %d, want %d", w.hoveredIndex, noHoveredIndex)
	}
}

func TestMousePress_RightButton_Ignored(t *testing.T) {
	w := New(Root(makeTestTree()), SelectionModeOpt(SelectionSingle))
	ctx := makeCtx()
	layoutAndBounds(w, ctx, 300, 400)

	e := &event.MouseEvent{
		MouseType: event.MousePress,
		Button:    event.ButtonRight,
		Position:  geometry.Pt(150, 14),
	}
	consumed := w.Event(ctx, e)
	if consumed {
		t.Error("right click should not be consumed")
	}
}

func TestMousePress_OutsideBounds(t *testing.T) {
	w := New(Root(makeTestTree()), SelectionModeOpt(SelectionSingle))
	ctx := makeCtx()
	layoutAndBounds(w, ctx, 300, 400)

	e := &event.MouseEvent{
		MouseType: event.MousePress,
		Button:    event.ButtonLeft,
		Position:  geometry.Pt(500, 500), // outside bounds
	}
	consumed := w.Event(ctx, e)
	if consumed {
		t.Error("click outside bounds should not be consumed")
	}
}

func TestMousePress_Disabled(t *testing.T) {
	w := New(Root(makeTestTree()), SelectionModeOpt(SelectionSingle), Disabled(true))
	ctx := makeCtx()
	layoutAndBounds(w, ctx, 300, 400)

	e := &event.MouseEvent{
		MouseType: event.MousePress,
		Button:    event.ButtonLeft,
		Position:  geometry.Pt(150, 14),
	}
	consumed := w.Event(ctx, e)
	if consumed {
		t.Error("disabled tree should not consume clicks")
	}
}

// --- Event: Keyboard ---

func TestKeyDown_MovesSelection(t *testing.T) {
	w := New(
		Root(makeTestTree()),
		SelectionModeOpt(SelectionSingle),
		SelectedNodeID("root"),
	)
	ctx := makeCtx()
	layoutAndBounds(w, ctx, 300, 400)
	w.SetFocused(true)

	e := &event.KeyEvent{
		KeyType: event.KeyPress,
		Key:     event.KeyDown,
	}
	consumed := w.Event(ctx, e)

	if !consumed {
		t.Error("expected key down to be consumed")
	}
	if w.cfg.ResolvedSelectedNodeID() != "child1" {
		t.Errorf("selected = %q, want %q", w.cfg.ResolvedSelectedNodeID(), "child1")
	}
}

func TestKeyUp_MovesSelection(t *testing.T) {
	w := New(
		Root(makeTestTree()),
		SelectionModeOpt(SelectionSingle),
		SelectedNodeID("child1"),
	)
	ctx := makeCtx()
	layoutAndBounds(w, ctx, 300, 400)
	w.SetFocused(true)

	e := &event.KeyEvent{KeyType: event.KeyPress, Key: event.KeyUp}
	w.Event(ctx, e)

	if w.cfg.ResolvedSelectedNodeID() != "root" {
		t.Errorf("selected = %q, want %q", w.cfg.ResolvedSelectedNodeID(), "root")
	}
}

func TestKeyHome_MovesToFirst(t *testing.T) {
	w := New(
		Root(makeTestTree()),
		SelectionModeOpt(SelectionSingle),
		SelectedNodeID("gc2"),
	)
	ctx := makeCtx()
	layoutAndBounds(w, ctx, 300, 400)
	w.SetFocused(true)

	e := &event.KeyEvent{KeyType: event.KeyPress, Key: event.KeyHome}
	w.Event(ctx, e)

	if w.cfg.ResolvedSelectedNodeID() != "root" {
		t.Errorf("selected = %q, want %q", w.cfg.ResolvedSelectedNodeID(), "root")
	}
}

func TestKeyEnd_MovesToLast(t *testing.T) {
	w := New(
		Root(makeTestTree()),
		SelectionModeOpt(SelectionSingle),
		SelectedNodeID("root"),
	)
	ctx := makeCtx()
	layoutAndBounds(w, ctx, 300, 400)
	w.SetFocused(true)

	e := &event.KeyEvent{KeyType: event.KeyPress, Key: event.KeyEnd}
	w.Event(ctx, e)

	if w.cfg.ResolvedSelectedNodeID() != "child2" {
		t.Errorf("selected = %q, want %q", w.cfg.ResolvedSelectedNodeID(), "child2")
	}
}

func TestKeyRight_ExpandsNode(t *testing.T) {
	root := makeTestTree()
	// Collapse child2 (it starts collapsed).
	w := New(
		Root(root),
		SelectionModeOpt(SelectionSingle),
		SelectedNodeID("child2"),
	)
	ctx := makeCtx()
	layoutAndBounds(w, ctx, 300, 400)
	w.SetFocused(true)

	e := &event.KeyEvent{KeyType: event.KeyPress, Key: event.KeyRight}
	consumed := w.Event(ctx, e)

	if !consumed {
		t.Error("expected key right to be consumed")
	}
	// child2 should now be expanded.
	child2 := findNodeByID(root, "child2")
	if !child2.Expanded {
		t.Error("expected child2 to be expanded")
	}
}

func TestKeyRight_MovesToFirstChild(t *testing.T) {
	root := makeTestTree()
	// child1 is already expanded.
	w := New(
		Root(root),
		SelectionModeOpt(SelectionSingle),
		SelectedNodeID("child1"),
	)
	ctx := makeCtx()
	layoutAndBounds(w, ctx, 300, 400)
	w.SetFocused(true)

	e := &event.KeyEvent{KeyType: event.KeyPress, Key: event.KeyRight}
	w.Event(ctx, e)

	if w.cfg.ResolvedSelectedNodeID() != "gc1" {
		t.Errorf("selected = %q, want %q", w.cfg.ResolvedSelectedNodeID(), "gc1")
	}
}

func TestKeyRight_LeafNode_Ignored(t *testing.T) {
	w := New(
		Root(makeTestTree()),
		SelectionModeOpt(SelectionSingle),
		SelectedNodeID("gc1"),
	)
	ctx := makeCtx()
	layoutAndBounds(w, ctx, 300, 400)
	w.SetFocused(true)

	e := &event.KeyEvent{KeyType: event.KeyPress, Key: event.KeyRight}
	consumed := w.Event(ctx, e)

	if consumed {
		t.Error("key right on leaf should not be consumed")
	}
}

func TestKeyLeft_CollapsesNode(t *testing.T) {
	root := makeTestTree()
	w := New(
		Root(root),
		SelectionModeOpt(SelectionSingle),
		SelectedNodeID("child1"),
	)
	ctx := makeCtx()
	layoutAndBounds(w, ctx, 300, 400)
	w.SetFocused(true)

	e := &event.KeyEvent{KeyType: event.KeyPress, Key: event.KeyLeft}
	consumed := w.Event(ctx, e)

	if !consumed {
		t.Error("expected key left to be consumed")
	}
	child1 := findNodeByID(root, "child1")
	if child1.Expanded {
		t.Error("expected child1 to be collapsed")
	}
}

func TestKeyLeft_MovesToParent(t *testing.T) {
	w := New(
		Root(makeTestTree()),
		SelectionModeOpt(SelectionSingle),
		SelectedNodeID("gc1"),
	)
	ctx := makeCtx()
	layoutAndBounds(w, ctx, 300, 400)
	w.SetFocused(true)

	e := &event.KeyEvent{KeyType: event.KeyPress, Key: event.KeyLeft}
	w.Event(ctx, e)

	if w.cfg.ResolvedSelectedNodeID() != "child1" {
		t.Errorf("selected = %q, want %q", w.cfg.ResolvedSelectedNodeID(), "child1")
	}
}

func TestKeyEnter_ActivatesSelect(t *testing.T) {
	var activated string
	w := New(
		Root(makeTestTree()),
		SelectionModeOpt(SelectionSingle),
		SelectedNodeID("gc2"),
		OnSelect(func(n *TreeNode) { activated = n.ID }),
	)
	ctx := makeCtx()
	layoutAndBounds(w, ctx, 300, 400)
	w.SetFocused(true)

	e := &event.KeyEvent{KeyType: event.KeyPress, Key: event.KeyEnter}
	consumed := w.Event(ctx, e)

	if !consumed {
		t.Error("expected enter to be consumed")
	}
	if activated != "gc2" {
		t.Errorf("activated = %q, want %q", activated, "gc2")
	}
}

func TestKeySpace_ActivatesSelect(t *testing.T) {
	var activated string
	w := New(
		Root(makeTestTree()),
		SelectionModeOpt(SelectionSingle),
		SelectedNodeID("root"),
		OnSelect(func(n *TreeNode) { activated = n.ID }),
	)
	ctx := makeCtx()
	layoutAndBounds(w, ctx, 300, 400)
	w.SetFocused(true)

	e := &event.KeyEvent{KeyType: event.KeyPress, Key: event.KeySpace}
	w.Event(ctx, e)

	if activated != "root" {
		t.Errorf("activated = %q, want %q", activated, "root")
	}
}

func TestKey_NotFocused_Ignored(t *testing.T) {
	w := New(Root(makeTestTree()), SelectionModeOpt(SelectionSingle))
	ctx := makeCtx()
	layoutAndBounds(w, ctx, 300, 400)
	// Not focused.

	e := &event.KeyEvent{KeyType: event.KeyPress, Key: event.KeyDown}
	consumed := w.Event(ctx, e)

	if consumed {
		t.Error("unfocused tree should not consume key events")
	}
}

func TestKey_Disabled_Ignored(t *testing.T) {
	w := New(Root(makeTestTree()), SelectionModeOpt(SelectionSingle), Disabled(true))
	ctx := makeCtx()
	layoutAndBounds(w, ctx, 300, 400)
	w.SetFocused(true)

	e := &event.KeyEvent{KeyType: event.KeyPress, Key: event.KeyDown}
	consumed := w.Event(ctx, e)

	if consumed {
		t.Error("disabled tree should not consume key events")
	}
}

func TestKey_EmptyTree_Ignored(t *testing.T) {
	w := New(SelectionModeOpt(SelectionSingle))
	ctx := makeCtx()
	layoutAndBounds(w, ctx, 300, 400)
	w.SetFocused(true)

	e := &event.KeyEvent{KeyType: event.KeyPress, Key: event.KeyDown}
	consumed := w.Event(ctx, e)

	if consumed {
		t.Error("empty tree should not consume key events")
	}
}

func TestKey_SelectionNone_Ignored(t *testing.T) {
	w := New(Root(makeTestTree()), SelectionModeOpt(SelectionNone))
	ctx := makeCtx()
	layoutAndBounds(w, ctx, 300, 400)
	w.SetFocused(true)

	e := &event.KeyEvent{KeyType: event.KeyPress, Key: event.KeyDown}
	consumed := w.Event(ctx, e)

	if consumed {
		t.Error("SelectionNone should not consume arrow keys")
	}
}

func TestKey_KeyRelease_Ignored(t *testing.T) {
	w := New(Root(makeTestTree()), SelectionModeOpt(SelectionSingle))
	ctx := makeCtx()
	layoutAndBounds(w, ctx, 300, 400)
	w.SetFocused(true)

	e := &event.KeyEvent{KeyType: event.KeyRelease, Key: event.KeyDown}
	consumed := w.Event(ctx, e)

	if consumed {
		t.Error("key release should not be consumed")
	}
}

// --- Event: Wheel ---

func TestWheelEvent_Scrolls(t *testing.T) {
	w := New(Root(makeTestTree()), ItemHeight(28))
	ctx := makeCtx()
	layoutAndBounds(w, ctx, 300, 56) // Only 2 rows visible.

	e := event.NewWheelEvent(
		geometry.Pt(0, -10), // Scroll down.
		geometry.Pt(150, 28),
		geometry.Pt(150, 28),
		0,
	)
	consumed := w.Event(ctx, e)

	if !consumed {
		t.Error("expected wheel to be consumed")
	}
	if w.scrollY <= 0 {
		t.Errorf("scrollY = %v, expected > 0", w.scrollY)
	}
}

func TestWheelEvent_ClampsToZero(t *testing.T) {
	w := New(Root(makeTestTree()), ItemHeight(28))
	ctx := makeCtx()
	layoutAndBounds(w, ctx, 300, 56)

	e := event.NewWheelEvent(
		geometry.Pt(0, 100), // Scroll up past start.
		geometry.Pt(150, 28),
		geometry.Pt(150, 28),
		0,
	)
	w.Event(ctx, e)

	if w.scrollY != 0 {
		t.Errorf("scrollY = %v, expected 0", w.scrollY)
	}
}

func TestWheelEvent_ClampsToMax(t *testing.T) {
	w := New(Root(makeTestTree()), ItemHeight(28))
	ctx := makeCtx()
	layoutAndBounds(w, ctx, 300, 56) // viewport=56, total=5*28=140, maxScroll=84

	e := event.NewWheelEvent(
		geometry.Pt(0, -1000), // Way past end.
		geometry.Pt(150, 28),
		geometry.Pt(150, 28),
		0,
	)
	w.Event(ctx, e)

	maxScroll := float32(5*28) - 56
	if w.scrollY != maxScroll {
		t.Errorf("scrollY = %v, expected %v", w.scrollY, maxScroll)
	}
}

func TestWheelEvent_OutsideBounds(t *testing.T) {
	w := New(Root(makeTestTree()))
	ctx := makeCtx()
	layoutAndBounds(w, ctx, 300, 400)

	e := event.NewWheelEvent(
		geometry.Pt(0, -10),
		geometry.Pt(500, 500), // Outside.
		geometry.Pt(500, 500),
		0,
	)
	consumed := w.Event(ctx, e)

	if consumed {
		t.Error("wheel outside bounds should not be consumed")
	}
}

// --- Signal Binding Tests ---

func TestRootSignal(t *testing.T) {
	root1 := &TreeNode{ID: "r1", Label: "Root 1"}
	root2 := &TreeNode{ID: "r2", Label: "Root 2", Children: []*TreeNode{
		{ID: "r2c", Label: "Child"},
	}, Expanded: true}

	sig := state.NewSignal(root1)
	w := New(RootSignal(sig))

	if w.cfg.ResolvedRoot().ID != "r1" {
		t.Error("expected root1")
	}

	sig.Set(root2)
	w.rebuildRows()
	if len(w.rows) != 2 {
		t.Errorf("expected 2 rows after root change, got %d", len(w.rows))
	}
}

func TestSelectedNodeSignal_TwoWay(t *testing.T) {
	sig := state.NewSignal("")
	w := New(
		Root(makeTestTree()),
		SelectionModeOpt(SelectionSingle),
		SelectedNodeSignal(sig),
	)
	ctx := makeCtx()
	layoutAndBounds(w, ctx, 300, 400)

	// External signal update.
	sig.Set("gc1")
	if w.cfg.ResolvedSelectedNodeID() != "gc1" {
		t.Errorf("selected = %q, want %q", w.cfg.ResolvedSelectedNodeID(), "gc1")
	}

	// Widget update writes back to signal.
	w.setSelectedNodeID(ctx, "child2")
	if sig.Get() != "child2" {
		t.Errorf("signal = %q, want %q", sig.Get(), "child2")
	}
}

func TestRootReadonlySignal(t *testing.T) {
	root := makeTestTree()
	sig := state.NewSignal(root)
	w := New(RootReadonlySignal(sig))

	if w.cfg.ResolvedRoot() != root {
		t.Error("expected root from readonly signal")
	}
}

func TestSelectedNodeReadonlySignal(t *testing.T) {
	sig := state.NewSignal("child1")
	w := New(
		Root(makeTestTree()),
		SelectedNodeReadonlySignal(sig),
	)

	if w.cfg.ResolvedSelectedNodeID() != "child1" {
		t.Errorf("selected = %q, want %q", w.cfg.ResolvedSelectedNodeID(), "child1")
	}
}

func TestDisabledSignal(t *testing.T) {
	sig := state.NewSignal(false)
	w := New(Root(makeTestTree()), DisabledSignal(sig))

	if w.cfg.ResolvedDisabled() {
		t.Error("expected not disabled initially")
	}

	sig.Set(true)
	if !w.cfg.ResolvedDisabled() {
		t.Error("expected disabled after signal set")
	}
}

func TestDisabledReadonlySignal(t *testing.T) {
	sig := state.NewSignal(true)
	w := New(Root(makeTestTree()), DisabledReadonlySignal(sig))

	if !w.cfg.ResolvedDisabled() {
		t.Error("expected disabled from readonly signal")
	}
}

func TestDisabledFn(t *testing.T) {
	disabled := false
	w := New(Root(makeTestTree()), DisabledFn(func() bool { return disabled }))

	if w.cfg.ResolvedDisabled() {
		t.Error("expected not disabled initially")
	}
	disabled = true
	if !w.cfg.ResolvedDisabled() {
		t.Error("expected disabled after fn returns true")
	}
}

// --- Config Resolved Priority Tests ---

func TestResolvedRoot_Priority(t *testing.T) {
	static := &TreeNode{ID: "static"}
	sigNode := &TreeNode{ID: "signal"}
	roNode := &TreeNode{ID: "readonly"}

	sig := state.NewSignal(sigNode)
	roSig := state.NewSignal(roNode)

	// ReadonlySignal > Signal > Static.
	cfg := config{
		root:               static,
		rootSignal:         sig,
		readonlyRootSignal: roSig,
	}
	if cfg.ResolvedRoot().ID != "readonly" {
		t.Error("readonly signal should have highest priority")
	}

	cfg.readonlyRootSignal = nil
	if cfg.ResolvedRoot().ID != "signal" {
		t.Error("signal should have second priority")
	}

	cfg.rootSignal = nil
	if cfg.ResolvedRoot().ID != "static" {
		t.Error("static should have lowest priority")
	}
}

func TestResolvedSelectedNodeID_Priority(t *testing.T) {
	sig := state.NewSignal("signal")
	roSig := state.NewSignal("readonly")

	cfg := config{
		selectedNodeID:             "static",
		selectedNodeIDSignal:       sig,
		readonlySelectedNodeSignal: roSig,
	}
	if cfg.ResolvedSelectedNodeID() != "readonly" {
		t.Error("readonly signal should have highest priority")
	}

	cfg.readonlySelectedNodeSignal = nil
	if cfg.ResolvedSelectedNodeID() != "signal" {
		t.Error("signal should have second priority")
	}

	cfg.selectedNodeIDSignal = nil
	if cfg.ResolvedSelectedNodeID() != "static" {
		t.Error("static should have lowest priority")
	}
}

func TestResolvedDisabled_Priority(t *testing.T) {
	sig := state.NewSignal(true)
	roSig := state.NewSignal(false)

	cfg := config{
		disabled:               true,
		disabledFn:             func() bool { return false },
		disabledSignal:         sig,
		readonlyDisabledSignal: roSig,
	}
	// ReadonlySignal (false) should win.
	if cfg.ResolvedDisabled() {
		t.Error("readonly signal (false) should have highest priority")
	}

	cfg.readonlyDisabledSignal = nil
	// Signal (true) should win.
	if !cfg.ResolvedDisabled() {
		t.Error("signal (true) should have second priority")
	}

	cfg.disabledSignal = nil
	// Fn (false) should win.
	if cfg.ResolvedDisabled() {
		t.Error("fn (false) should have third priority")
	}

	cfg.disabledFn = nil
	// Static (true) should remain.
	if !cfg.ResolvedDisabled() {
		t.Error("static (true) should have lowest priority")
	}
}

// --- Public API Tests ---

func TestScrollToNode(t *testing.T) {
	w := New(Root(makeTestTree()), ItemHeight(28))
	ctx := makeCtx()
	layoutAndBounds(w, ctx, 300, 56) // Only 2 rows visible.

	// child2 is row index 4. Top = 4*28 = 112, bottom = 140.
	w.ScrollToNode("child2")

	// It should scroll so child2 is visible.
	if w.scrollY <= 0 {
		t.Errorf("scrollY = %v, expected > 0 after ScrollToNode", w.scrollY)
	}
}

func TestScrollToNode_AlreadyVisible(t *testing.T) {
	w := New(Root(makeTestTree()), ItemHeight(28))
	ctx := makeCtx()
	layoutAndBounds(w, ctx, 300, 400) // All rows visible.

	w.ScrollToNode("root")
	if w.scrollY != 0 {
		t.Errorf("scrollY = %v, expected 0 (already visible)", w.scrollY)
	}
}

func TestScrollToNode_NotFound(t *testing.T) {
	w := New(Root(makeTestTree()), ItemHeight(28))
	ctx := makeCtx()
	layoutAndBounds(w, ctx, 300, 56)

	w.ScrollToNode("nonexistent")
	if w.scrollY != 0 {
		t.Error("scrollY should not change for nonexistent node")
	}
}

func TestVisibleRange(t *testing.T) {
	w := New(Root(makeTestTree()), ItemHeight(28))
	ctx := makeCtx()
	layoutAndBounds(w, ctx, 300, 56) // 2 rows visible + 2 partial.

	start, end := w.VisibleRange()
	if start != 0 {
		t.Errorf("start = %d, want 0", start)
	}
	if end > len(w.rows) {
		t.Errorf("end = %d, exceeds row count %d", end, len(w.rows))
	}
}

func TestRowCount(t *testing.T) {
	w := New(Root(makeTestTree()))
	if w.RowCount() != 5 {
		t.Errorf("RowCount() = %d, want 5", w.RowCount())
	}
}

func TestInvalidateData(t *testing.T) {
	root := makeTestTree()
	w := New(Root(root))

	if w.RowCount() != 5 {
		t.Fatalf("initial row count = %d, want 5", w.RowCount())
	}

	// Add a new child to root.
	root.Children = append(root.Children, &TreeNode{ID: "child3", Label: "Child 3"})
	w.InvalidateData()

	if w.RowCount() != 6 {
		t.Errorf("row count after invalidate = %d, want 6", w.RowCount())
	}
}

func TestExpandAll(t *testing.T) {
	root := makeTestTree()
	w := New(Root(root))

	// Initially: root expanded, child1 expanded, child2 collapsed.
	// So 5 rows. After ExpandAll: root, child1, gc1, gc2, child2, gc3 = 6.
	w.ExpandAll()

	if w.RowCount() != 6 {
		t.Errorf("RowCount() after ExpandAll = %d, want 6", w.RowCount())
	}
}

func TestCollapseAll(t *testing.T) {
	root := makeTestTree()
	w := New(Root(root))

	w.CollapseAll()

	// Only root should be visible (root itself is collapsed).
	if w.RowCount() != 1 {
		t.Errorf("RowCount() after CollapseAll = %d, want 1", w.RowCount())
	}
}

func TestExpandAll_NilRoot(t *testing.T) {
	w := New()
	w.ExpandAll() // Should not panic.
}

func TestCollapseAll_NilRoot(t *testing.T) {
	w := New()
	w.CollapseAll() // Should not panic.
}

// --- Lifecycle Tests ---

func TestMount_BindsSignals(t *testing.T) {
	rootSig := state.NewSignal(makeTestTree())
	selSig := state.NewSignal("")
	disSig := state.NewSignal(false)

	w := New(
		RootSignal(rootSig),
		SelectedNodeSignal(selSig),
		DisabledSignal(disSig),
	)

	ctx := makeCtx()
	// Mount requires scheduler.
	sched := &mockScheduler{}
	ctx.SetScheduler(sched)
	w.Mount(ctx)

	// Verify bindings were added (3 bindings).
	// We can't directly check count, but verify no panic.
}

func TestMount_NilScheduler(t *testing.T) {
	w := New(Root(makeTestTree()))
	ctx := makeCtx()
	w.Mount(ctx) // Should not panic with nil scheduler.
}

func TestUnmount(t *testing.T) {
	w := New(Root(makeTestTree()))
	w.Unmount() // Should not panic.
}

// --- Children Tests ---

func TestChildren_ReturnsNil(t *testing.T) {
	w := New(Root(makeTestTree()))
	if w.Children() != nil {
		t.Error("expected nil children (tree renders internally)")
	}
}

// --- Accessibility Tests ---

func TestAccessibility(t *testing.T) {
	w := New(Root(makeTestTree()), A11yLabel("File tree"))

	if w.AccessibilityRole() != a11y.RoleTree {
		t.Errorf("role = %v, want RoleTree", w.AccessibilityRole())
	}
	if w.AccessibilityLabel() != "File tree" {
		t.Errorf("label = %q, want %q", w.AccessibilityLabel(), "File tree")
	}
	if w.AccessibilityHint() != "" {
		t.Errorf("hint = %q, want empty", w.AccessibilityHint())
	}
}

func TestAccessibilityLabel_Default(t *testing.T) {
	w := New()
	if w.AccessibilityLabel() != "Tree" {
		t.Errorf("label = %q, want %q", w.AccessibilityLabel(), "Tree")
	}
}

func TestAccessibilityValue_WithSelection(t *testing.T) {
	w := New(
		Root(makeTestTree()),
		SelectedNodeID("child1"),
	)
	val := w.AccessibilityValue()
	if val != "Selected: Child 1, 5 visible rows" {
		t.Errorf("value = %q", val)
	}
}

func TestAccessibilityValue_NoSelection(t *testing.T) {
	w := New(Root(makeTestTree()))
	val := w.AccessibilityValue()
	if val != "5 visible rows" {
		t.Errorf("value = %q", val)
	}
}

func TestAccessibilityState(t *testing.T) {
	w := New(Root(makeTestTree()), Disabled(true))
	s := w.AccessibilityState()
	if !s.Disabled {
		t.Error("expected disabled state")
	}
}

func TestAccessibilityActions(t *testing.T) {
	w := New()
	actions := w.AccessibilityActions()
	if len(actions) != 2 {
		t.Errorf("expected 2 actions, got %d", len(actions))
	}
}

// --- Edge Cases ---

func TestKeyDown_ClampsAtEnd(t *testing.T) {
	w := New(
		Root(makeTestTree()),
		SelectionModeOpt(SelectionSingle),
		SelectedNodeID("child2"), // Last visible row.
	)
	ctx := makeCtx()
	layoutAndBounds(w, ctx, 300, 400)
	w.SetFocused(true)

	e := &event.KeyEvent{KeyType: event.KeyPress, Key: event.KeyDown}
	w.Event(ctx, e)

	// Should stay at child2 (index 4, last row).
	if w.cfg.ResolvedSelectedNodeID() != "child2" {
		t.Errorf("selected = %q, want %q", w.cfg.ResolvedSelectedNodeID(), "child2")
	}
}

func TestKeyUp_ClampsAtStart(t *testing.T) {
	w := New(
		Root(makeTestTree()),
		SelectionModeOpt(SelectionSingle),
		SelectedNodeID("root"),
	)
	ctx := makeCtx()
	layoutAndBounds(w, ctx, 300, 400)
	w.SetFocused(true)

	e := &event.KeyEvent{KeyType: event.KeyPress, Key: event.KeyUp}
	w.Event(ctx, e)

	if w.cfg.ResolvedSelectedNodeID() != "root" {
		t.Errorf("selected = %q, want %q", w.cfg.ResolvedSelectedNodeID(), "root")
	}
}

func TestSetSelectedNodeID_SameValue_Noop(t *testing.T) {
	var selectCount int
	w := New(
		Root(makeTestTree()),
		SelectionModeOpt(SelectionSingle),
		SelectedNodeID("root"),
		OnSelect(func(_ *TreeNode) { selectCount++ }),
	)
	ctx := makeCtx()

	w.setSelectedNodeID(ctx, "root")
	if selectCount != 0 {
		t.Error("setting same value should not trigger callback")
	}
}

func TestToggleNode_LeafIgnored(t *testing.T) {
	root := makeTestTree()
	w := New(Root(root))
	ctx := makeCtx()

	gc1 := findNodeByID(root, "gc1")
	rowsBefore := w.RowCount()
	w.toggleNode(ctx, gc1)

	if w.RowCount() != rowsBefore {
		t.Error("toggling leaf should not change row count")
	}
}

func TestHitTestRow_OutsideBounds(t *testing.T) {
	w := New(Root(makeTestTree()))
	ctx := makeCtx()
	layoutAndBounds(w, ctx, 300, 400)

	idx := w.hitTestRow(geometry.Pt(500, 500))
	if idx != noHoveredIndex {
		t.Errorf("hitTestRow outside bounds = %d, want %d", idx, noHoveredIndex)
	}
}

func TestHitTestRow_ZeroItemHeight(t *testing.T) {
	w := New(Root(makeTestTree()), ItemHeight(0))
	ctx := makeCtx()
	layoutAndBounds(w, ctx, 300, 400)

	idx := w.hitTestRow(geometry.Pt(150, 14))
	if idx != noHoveredIndex {
		t.Errorf("hitTestRow with zero height = %d, want %d", idx, noHoveredIndex)
	}
}

func TestVisibleRange_EmptyTree(t *testing.T) {
	w := New()
	start, end := w.visibleRange()
	if start != 0 || end != 0 {
		t.Errorf("visibleRange empty = (%d, %d), want (0, 0)", start, end)
	}
}

func TestBuildConnectorState(t *testing.T) {
	w := New(Root(makeTestTree()), ShowLines(true), IndentWidth(20))
	ctx := makeCtx()
	layoutAndBounds(w, ctx, 300, 400)

	bounds := w.Bounds()
	// Row 2 is gc1 at depth 2.
	rowBounds := w.rowBounds(2, bounds)
	cs := w.buildConnectorState(2, rowBounds)

	if cs.Depth != 2 {
		t.Errorf("depth = %d, want 2", cs.Depth)
	}
	if cs.IndentWidth != 20 {
		t.Errorf("indentWidth = %v, want 20", cs.IndentWidth)
	}
	if len(cs.ParentHasMore) != 2 {
		t.Errorf("parentHasMore length = %d, want 2", len(cs.ParentHasMore))
	}
}

// --- Painter Tests ---

func TestDefaultPainter_PaintEmptyState(t *testing.T) {
	p := DefaultPainter{}
	canvas := &mockCanvas{}
	bounds := geometry.NewRect(0, 0, 300, 200)

	p.PaintEmptyState(canvas, bounds)
	if canvas.drawTextCalls != 1 {
		t.Errorf("drawText calls = %d, want 1", canvas.drawTextCalls)
	}
}

func TestDefaultPainter_PaintEmptyState_EmptyBounds(t *testing.T) {
	p := DefaultPainter{}
	canvas := &mockCanvas{}

	p.PaintEmptyState(canvas, geometry.Rect{})
	if canvas.drawTextCalls != 0 {
		t.Error("empty bounds should not draw")
	}
}

func TestDefaultPainter_PaintRowBackground_Hover(t *testing.T) {
	p := DefaultPainter{}
	canvas := &mockCanvas{}

	p.PaintRowBackground(canvas, RowPaintState{
		Bounds:  geometry.NewRect(0, 0, 300, 28),
		Hovered: true,
	})
	if canvas.drawRectCalls != 1 {
		t.Errorf("drawRect calls = %d, want 1", canvas.drawRectCalls)
	}
}

func TestDefaultPainter_PaintRowBackground_NoHover(t *testing.T) {
	p := DefaultPainter{}
	canvas := &mockCanvas{}

	p.PaintRowBackground(canvas, RowPaintState{
		Bounds:  geometry.NewRect(0, 0, 300, 28),
		Hovered: false,
	})
	if canvas.drawRectCalls != 0 {
		t.Error("no hover should not draw rect")
	}
}

func TestDefaultPainter_PaintRowBackground_Disabled(t *testing.T) {
	p := DefaultPainter{}
	canvas := &mockCanvas{}

	p.PaintRowBackground(canvas, RowPaintState{
		Bounds:   geometry.NewRect(0, 0, 300, 28),
		Hovered:  true,
		Disabled: true,
	})
	if canvas.drawRectCalls != 0 {
		t.Error("disabled hover should not draw rect")
	}
}

func TestDefaultPainter_PaintSelection(t *testing.T) {
	p := DefaultPainter{}
	canvas := &mockCanvas{}

	p.PaintSelection(canvas, RowPaintState{
		Bounds:   geometry.NewRect(0, 0, 300, 28),
		Selected: true,
	})
	if canvas.drawRectCalls != 1 {
		t.Errorf("drawRect calls = %d, want 1", canvas.drawRectCalls)
	}
}

func TestDefaultPainter_PaintSelection_Focused(t *testing.T) {
	p := DefaultPainter{}
	canvas := &mockCanvas{}

	p.PaintSelection(canvas, RowPaintState{
		Bounds:   geometry.NewRect(0, 0, 300, 28),
		Selected: true,
		Focused:  true,
	})
	if canvas.drawRectCalls != 1 {
		t.Errorf("drawRect calls = %d, want 1", canvas.drawRectCalls)
	}
	if canvas.strokeRectCalls != 1 {
		t.Errorf("strokeRect calls = %d, want 1", canvas.strokeRectCalls)
	}
}

func TestDefaultPainter_PaintSelection_NotSelected(t *testing.T) {
	p := DefaultPainter{}
	canvas := &mockCanvas{}

	p.PaintSelection(canvas, RowPaintState{
		Bounds:   geometry.NewRect(0, 0, 300, 28),
		Selected: false,
	})
	if canvas.drawRectCalls != 0 {
		t.Error("not selected should not draw")
	}
}

func TestDefaultPainter_PaintExpandIcon_Expanded(t *testing.T) {
	p := DefaultPainter{}
	canvas := &mockCanvas{}

	p.PaintExpandIcon(canvas, ExpandIconState{
		Bounds:   geometry.NewRect(0, 0, 16, 16),
		Expanded: true,
	})
	if canvas.drawLineCalls != 2 {
		t.Errorf("drawLine calls = %d, want 2", canvas.drawLineCalls)
	}
}

func TestDefaultPainter_PaintExpandIcon_Collapsed(t *testing.T) {
	p := DefaultPainter{}
	canvas := &mockCanvas{}

	p.PaintExpandIcon(canvas, ExpandIconState{
		Bounds:   geometry.NewRect(0, 0, 16, 16),
		Expanded: false,
	})
	if canvas.drawLineCalls != 2 {
		t.Errorf("drawLine calls = %d, want 2", canvas.drawLineCalls)
	}
}

func TestDefaultPainter_PaintExpandIcon_EmptyBounds(t *testing.T) {
	p := DefaultPainter{}
	canvas := &mockCanvas{}

	p.PaintExpandIcon(canvas, ExpandIconState{Bounds: geometry.Rect{}})
	if canvas.drawLineCalls != 0 {
		t.Error("empty bounds should not draw")
	}
}

func TestDefaultPainter_PaintConnectorLines(t *testing.T) {
	p := DefaultPainter{}
	canvas := &mockCanvas{}

	p.PaintConnectorLines(canvas, ConnectorState{
		RowBounds:     geometry.NewRect(0, 0, 300, 28),
		Depth:         1,
		IndentWidth:   20,
		IsLastChild:   false,
		ParentHasMore: []bool{false},
	})
	// At minimum: vertical + horizontal lines.
	if canvas.drawLineCalls < 2 {
		t.Errorf("drawLine calls = %d, want >= 2", canvas.drawLineCalls)
	}
}

func TestDefaultPainter_PaintConnectorLines_DepthZero(t *testing.T) {
	p := DefaultPainter{}
	canvas := &mockCanvas{}

	p.PaintConnectorLines(canvas, ConnectorState{
		RowBounds: geometry.NewRect(0, 0, 300, 28),
		Depth:     0,
	})
	if canvas.drawLineCalls != 0 {
		t.Error("depth 0 should not draw connector lines")
	}
}

func TestDefaultPainter_PaintLabel(t *testing.T) {
	p := DefaultPainter{}
	canvas := &mockCanvas{}

	p.PaintLabel(canvas, LabelState{
		Bounds: geometry.NewRect(0, 0, 200, 28),
		Text:   "Hello",
	})
	if canvas.drawTextCalls != 1 {
		t.Errorf("drawText calls = %d, want 1", canvas.drawTextCalls)
	}
}

func TestDefaultPainter_PaintLabel_Empty(t *testing.T) {
	p := DefaultPainter{}
	canvas := &mockCanvas{}

	p.PaintLabel(canvas, LabelState{
		Bounds: geometry.NewRect(0, 0, 200, 28),
		Text:   "",
	})
	if canvas.drawTextCalls != 0 {
		t.Error("empty text should not draw")
	}
}

func TestDefaultPainter_PaintLabel_EmptyBounds(t *testing.T) {
	p := DefaultPainter{}
	canvas := &mockCanvas{}

	p.PaintLabel(canvas, LabelState{
		Bounds: geometry.Rect{},
		Text:   "Hello",
	})
	if canvas.drawTextCalls != 0 {
		t.Error("empty bounds should not draw")
	}
}

func TestDefaultPainter_ColorScheme(t *testing.T) {
	p := DefaultPainter{}
	canvas := &mockCanvas{}
	scheme := TreeColorScheme{
		SelectionColor: widget.RGBA(1, 0, 0, 1),
		HoverColor:     widget.RGBA(0, 1, 0, 1),
		FocusColor:     widget.RGBA(0, 0, 1, 1),
		LabelColor:     widget.RGBA(1, 1, 0, 1),
		LineColor:      widget.RGBA(0, 1, 1, 1),
		IconColor:      widget.RGBA(1, 0, 1, 1),
		EmptyTextColor: widget.RGBA(0.5, 0.5, 0.5, 1),
	}

	// Test that color scheme is passed through without panic.
	p.PaintRowBackground(canvas, RowPaintState{
		Bounds:      geometry.NewRect(0, 0, 300, 28),
		Hovered:     true,
		ColorScheme: scheme,
	})
	p.PaintSelection(canvas, RowPaintState{
		Bounds:      geometry.NewRect(0, 0, 300, 28),
		Selected:    true,
		Focused:     true,
		ColorScheme: scheme,
	})
	p.PaintExpandIcon(canvas, ExpandIconState{
		Bounds:      geometry.NewRect(0, 0, 16, 16),
		Expanded:    true,
		ColorScheme: scheme,
	})
	p.PaintConnectorLines(canvas, ConnectorState{
		RowBounds:     geometry.NewRect(0, 0, 300, 28),
		Depth:         1,
		IndentWidth:   20,
		ParentHasMore: []bool{false},
		ColorScheme:   scheme,
	})
	p.PaintLabel(canvas, LabelState{
		Bounds:      geometry.NewRect(0, 0, 200, 28),
		Text:        "Test",
		ColorScheme: scheme,
	})
}

func TestDefaultPainter_RowBackground_EmptyBounds(t *testing.T) {
	p := DefaultPainter{}
	canvas := &mockCanvas{}
	p.PaintRowBackground(canvas, RowPaintState{Bounds: geometry.Rect{}, Hovered: true})
	if canvas.drawRectCalls != 0 {
		t.Error("empty bounds should not draw")
	}
}

func TestDefaultPainter_Selection_EmptyBounds(t *testing.T) {
	p := DefaultPainter{}
	canvas := &mockCanvas{}
	p.PaintSelection(canvas, RowPaintState{Bounds: geometry.Rect{}, Selected: true})
	if canvas.drawRectCalls != 0 {
		t.Error("empty bounds should not draw")
	}
}

func TestDefaultPainter_ConnectorLines_EmptyBounds(t *testing.T) {
	p := DefaultPainter{}
	canvas := &mockCanvas{}
	p.PaintConnectorLines(canvas, ConnectorState{RowBounds: geometry.Rect{}, Depth: 1})
	if canvas.drawLineCalls != 0 {
		t.Error("empty bounds should not draw")
	}
}

// --- Event Type Dispatch ---

func TestEvent_InvisibleWidget(t *testing.T) {
	w := New(Root(makeTestTree()))
	w.SetVisible(false)
	ctx := makeCtx()

	e := &event.MouseEvent{MouseType: event.MousePress, Position: geometry.Pt(150, 14)}
	if w.Event(ctx, e) {
		t.Error("invisible widget should not consume events")
	}
}

func TestEvent_DisabledWidget(t *testing.T) {
	w := New(Root(makeTestTree()))
	w.SetEnabled(false)
	ctx := makeCtx()

	e := &event.MouseEvent{MouseType: event.MousePress, Position: geometry.Pt(150, 14)}
	if w.Event(ctx, e) {
		t.Error("disabled widget should not consume events")
	}
}

func TestEvent_UnknownEventType(t *testing.T) {
	w := New(Root(makeTestTree()))
	ctx := makeCtx()
	layoutAndBounds(w, ctx, 300, 400)

	e := &event.FocusEvent{}
	if w.Event(ctx, e) {
		t.Error("unknown event type should not be consumed")
	}
}

// --- SetExpandedAll Tests ---

func TestSetExpandedAll(t *testing.T) {
	root := makeTestTree()
	setExpandedAll(root, true)

	// Verify all branch nodes are expanded.
	child2 := findNodeByID(root, "child2")
	if !child2.Expanded {
		t.Error("child2 should be expanded")
	}

	setExpandedAll(root, false)
	if root.Expanded {
		t.Error("root should be collapsed")
	}
	if findNodeByID(root, "child1").Expanded {
		t.Error("child1 should be collapsed")
	}
}

// --- Scroll Edge Cases ---

func TestScrollToIndex_BelowViewport(t *testing.T) {
	w := New(Root(makeTestTree()), ItemHeight(28))
	ctx := makeCtx()
	layoutAndBounds(w, ctx, 300, 56)

	w.scrollToIndex(4)                      // Last row.
	expectedScroll := float32(4*28+28) - 56 // itemBottom - viewport.
	if w.scrollY != expectedScroll {
		t.Errorf("scrollY = %v, want %v", w.scrollY, expectedScroll)
	}
}

func TestScrollToIndex_AboveViewport(t *testing.T) {
	w := New(Root(makeTestTree()), ItemHeight(28))
	ctx := makeCtx()
	layoutAndBounds(w, ctx, 300, 56)

	w.scrollY = 100
	w.scrollToIndex(0)

	if w.scrollY != 0 {
		t.Errorf("scrollY = %v, want 0", w.scrollY)
	}
}

func TestScrollToIndex_OutOfRange(t *testing.T) {
	w := New(Root(makeTestTree()), ItemHeight(28))
	ctx := makeCtx()
	layoutAndBounds(w, ctx, 300, 56)

	w.scrollToIndex(-1)  // Should be no-op.
	w.scrollToIndex(100) // Should be no-op.
}

// --- Key Activate with No Callback ---

func TestKeyActivate_NoCallback(t *testing.T) {
	w := New(
		Root(makeTestTree()),
		SelectionModeOpt(SelectionSingle),
		SelectedNodeID("root"),
	)
	ctx := makeCtx()
	layoutAndBounds(w, ctx, 300, 400)
	w.SetFocused(true)

	e := &event.KeyEvent{KeyType: event.KeyPress, Key: event.KeyEnter}
	consumed := w.Event(ctx, e)

	if consumed {
		t.Error("enter with no OnSelect callback should not be consumed")
	}
}

func TestKeyActivate_OutOfRange(t *testing.T) {
	w := New(
		Root(makeTestTree()),
		SelectionModeOpt(SelectionSingle),
		// No initial selection.
	)
	ctx := makeCtx()
	layoutAndBounds(w, ctx, 300, 400)
	w.SetFocused(true)

	e := &event.KeyEvent{KeyType: event.KeyPress, Key: event.KeyEnter}
	consumed := w.Event(ctx, e)

	if consumed {
		t.Error("enter with no selection should not be consumed")
	}
}

// --- Mouse on leaf toggle area ---

func TestMousePress_LeafToggleArea_NoExpand(t *testing.T) {
	root := &TreeNode{
		ID: "root", Label: "Root", Expanded: true,
		Children: []*TreeNode{
			{ID: "leaf", Label: "Leaf"}, // No children.
		},
	}
	w := New(Root(root), SelectionModeOpt(SelectionSingle))
	ctx := makeCtx()
	layoutAndBounds(w, ctx, 300, 400)

	// Click on icon area of leaf row (row 1, depth=1).
	e := &event.MouseEvent{
		MouseType: event.MousePress,
		Button:    event.ButtonLeft,
		Position:  geometry.Pt(28, 42),
	}
	consumed := w.Event(ctx, e)

	// Should select, not toggle (leaf has no expand icon).
	if !consumed {
		t.Error("expected click to be consumed for selection")
	}
	if w.cfg.ResolvedSelectedNodeID() != "leaf" {
		t.Errorf("selected = %q, want %q", w.cfg.ResolvedSelectedNodeID(), "leaf")
	}
}

// --- Unrecognized Key ---

func TestKey_Unrecognized_NotConsumed(t *testing.T) {
	w := New(Root(makeTestTree()), SelectionModeOpt(SelectionSingle))
	ctx := makeCtx()
	layoutAndBounds(w, ctx, 300, 400)
	w.SetFocused(true)

	e := &event.KeyEvent{KeyType: event.KeyPress, Key: event.KeyTab}
	consumed := w.Event(ctx, e)
	if consumed {
		t.Error("unrecognized key should not be consumed")
	}
}

// --- Mock Types ---

type mockCanvas struct {
	drawRectCalls      int
	strokeRectCalls    int
	drawLineCalls      int
	drawTextCalls      int
	pushClipCalls      int
	popClipCalls       int
	drawCircleCalls    int
	drawRoundRectCalls int
}

func (c *mockCanvas) Clear(widget.Color)                                            {}
func (c *mockCanvas) DrawRect(geometry.Rect, widget.Color)                          { c.drawRectCalls++ }
func (c *mockCanvas) StrokeRect(geometry.Rect, widget.Color, float32)               { c.strokeRectCalls++ }
func (c *mockCanvas) DrawRoundRect(geometry.Rect, widget.Color, float32)            { c.drawRoundRectCalls++ }
func (c *mockCanvas) StrokeRoundRect(geometry.Rect, widget.Color, float32, float32) {}
func (c *mockCanvas) DrawCircle(geometry.Point, float32, widget.Color)              { c.drawCircleCalls++ }
func (c *mockCanvas) StrokeCircle(geometry.Point, float32, widget.Color, float32)   {}
func (c *mockCanvas) DrawLine(geometry.Point, geometry.Point, widget.Color, float32) {
	c.drawLineCalls++
}
func (c *mockCanvas) DrawText(string, geometry.Rect, float32, widget.Color, bool, widget.TextAlign) {
	c.drawTextCalls++
}

func (c *mockCanvas) MeasureText(text string, fontSize float32, _ bool) float32 {
	return float32(len([]rune(text))) * fontSize * 0.5
}
func (c *mockCanvas) DrawImage(_ image.Image, _ geometry.Point) {}
func (c *mockCanvas) PushClip(geometry.Rect)                    { c.pushClipCalls++ }
func (c *mockCanvas) PushClipRoundRect(geometry.Rect, float32)  {}
func (c *mockCanvas) PopClip()                                  { c.popClipCalls++ }
func (c *mockCanvas) PushTransform(geometry.Point)              {}
func (c *mockCanvas) PopTransform()                             {}
func (c *mockCanvas) TransformOffset() geometry.Point           { return geometry.Point{} }

type mockScheduler struct{}

func (s *mockScheduler) MarkDirty(widget.Widget) {}
