package a11y

import (
	"fmt"
	"sync"
	"testing"

	"github.com/gogpu/ui/geometry"
)

// --- Role tests ---

func TestRoleString(t *testing.T) {
	tests := []struct {
		role Role
		want string
	}{
		// Structural
		{RoleUnknown, "Unknown"},
		{RoleWindow, "Window"},
		{RoleGroup, "Group"},
		{RoleSeparator, "Separator"},
		{RoleToolbar, "Toolbar"},
		{RoleStatusBar, "StatusBar"},
		{RoleMenuBar, "MenuBar"},
		{RoleGenericContainer, "GenericContainer"},

		// Input
		{RoleButton, "Button"},
		{RoleCheckbox, "Checkbox"},
		{RoleRadio, "Radio"},
		{RoleTextField, "TextField"},
		{RoleTextArea, "TextArea"},
		{RoleSlider, "Slider"},
		{RoleSwitch, "Switch"},
		{RoleComboBox, "ComboBox"},
		{RoleSpinButton, "SpinButton"},
		{RoleRadioGroup, "RadioGroup"},
		{RoleSearchBox, "SearchBox"},
		{RoleToggleButton, "ToggleButton"},
		{RoleColorWell, "ColorWell"},

		// Display
		{RoleLabel, "Label"},
		{RoleImage, "Image"},
		{RoleProgressBar, "ProgressBar"},
		{RoleTooltip, "Tooltip"},
		{RoleAlert, "Alert"},
		{RoleBadge, "Badge"},
		{RoleHeading, "Heading"},
		{RoleMeter, "Meter"},
		{RoleStaticText, "StaticText"},

		// Container
		{RoleDialog, "Dialog"},
		{RoleAlertDialog, "AlertDialog"},
		{RoleMenu, "Menu"},
		{RoleMenuItem, "MenuItem"},
		{RoleMenuItemCheckbox, "MenuItemCheckbox"},
		{RoleMenuItemRadio, "MenuItemRadio"},
		{RoleList, "List"},
		{RoleListItem, "ListItem"},
		{RoleTree, "Tree"},
		{RoleTreeItem, "TreeItem"},
		{RoleTab, "Tab"},
		{RoleTabList, "TabList"},
		{RoleTabPanel, "TabPanel"},
		{RoleGrid, "Grid"},
		{RoleGridCell, "GridCell"},
		{RoleTable, "Table"},
		{RoleRow, "Row"},
		{RoleCell, "Cell"},
		{RoleColumnHeader, "ColumnHeader"},
		{RoleRowHeader, "RowHeader"},
		{RoleListBox, "ListBox"},
		{RoleScrollView, "ScrollView"},
		{RoleApplication, "Application"},
		{RoleDocument, "Document"},
		{RoleFeed, "Feed"},

		// Navigation
		{RoleLink, "Link"},
		{RoleScrollBar, "ScrollBar"},
		{RoleNavigation, "Navigation"},
		{RoleBanner, "Banner"},
		{RoleMain, "Main"},
		{RoleContentInfo, "ContentInfo"},
		{RoleComplementary, "Complementary"},
		{RoleRegion, "Region"},
		{RoleForm, "Form"},
		{RoleSearch, "Search"},

		// Invalid
		{Role(255), "Unknown"},
	}
	for _, tt := range tests {
		t.Run(tt.want, func(t *testing.T) {
			got := tt.role.String()
			if got != tt.want {
				t.Errorf("Role(%d).String() = %q, want %q", tt.role, got, tt.want)
			}
		})
	}
}

func TestRoleIsInteractive(t *testing.T) {
	tests := []struct {
		role Role
		want bool
	}{
		{RoleButton, true},
		{RoleCheckbox, true},
		{RoleSlider, true},
		{RoleTextField, true},
		{RoleSwitch, true},
		{RoleColorWell, true},
		{RoleLabel, false},
		{RoleWindow, false},
		{RoleDialog, false},
		{RoleLink, false},
		{RoleUnknown, false},
	}
	for _, tt := range tests {
		t.Run(tt.role.String(), func(t *testing.T) {
			got := tt.role.IsInteractive()
			if got != tt.want {
				t.Errorf("Role(%d).IsInteractive() = %v, want %v", tt.role, got, tt.want)
			}
		})
	}
}

func TestRoleIsContainer(t *testing.T) {
	tests := []struct {
		role Role
		want bool
	}{
		{RoleDialog, true},
		{RoleMenu, true},
		{RoleList, true},
		{RoleTable, true},
		{RoleFeed, true},
		{RoleButton, false},
		{RoleLabel, false},
		{RoleLink, false},
	}
	for _, tt := range tests {
		t.Run(tt.role.String(), func(t *testing.T) {
			got := tt.role.IsContainer()
			if got != tt.want {
				t.Errorf("Role(%d).IsContainer() = %v, want %v", tt.role, got, tt.want)
			}
		})
	}
}

func TestRoleIsLandmark(t *testing.T) {
	tests := []struct {
		role Role
		want bool
	}{
		{RoleNavigation, true},
		{RoleBanner, true},
		{RoleMain, true},
		{RoleContentInfo, true},
		{RoleComplementary, true},
		{RoleRegion, true},
		{RoleForm, true},
		{RoleSearch, true},
		{RoleButton, false},
		{RoleLink, false},
		{RoleDialog, false},
	}
	for _, tt := range tests {
		t.Run(tt.role.String(), func(t *testing.T) {
			got := tt.role.IsLandmark()
			if got != tt.want {
				t.Errorf("Role(%d).IsLandmark() = %v, want %v", tt.role, got, tt.want)
			}
		})
	}
}

// --- CheckedState tests ---

func TestCheckedStateString(t *testing.T) {
	tests := []struct {
		state CheckedState
		want  string
	}{
		{CheckedFalse, "Unchecked"},
		{CheckedTrue, "Checked"},
		{CheckedMixed, "Mixed"},
		{CheckedState(99), "Unknown"},
	}
	for _, tt := range tests {
		t.Run(tt.want, func(t *testing.T) {
			got := tt.state.String()
			if got != tt.want {
				t.Errorf("CheckedState(%d).String() = %q, want %q", tt.state, got, tt.want)
			}
		})
	}
}

// --- State tests ---

func TestStateDefaults(t *testing.T) {
	var s State

	if s.Disabled {
		t.Error("default Disabled should be false")
	}
	if s.Selected {
		t.Error("default Selected should be false")
	}
	if s.Checked != CheckedFalse {
		t.Errorf("default Checked = %v, want CheckedFalse", s.Checked)
	}
	if s.Expanded != nil {
		t.Error("default Expanded should be nil")
	}
	if s.ReadOnly {
		t.Error("default ReadOnly should be false")
	}
	if s.Required {
		t.Error("default Required should be false")
	}
	if s.Busy {
		t.Error("default Busy should be false")
	}
	if s.Hidden {
		t.Error("default Hidden should be false")
	}
	if s.Focused {
		t.Error("default Focused should be false")
	}
	if s.Modal {
		t.Error("default Modal should be false")
	}
	if s.Multiselectable {
		t.Error("default Multiselectable should be false")
	}
	if s.ValueMin != nil {
		t.Error("default ValueMin should be nil")
	}
	if s.ValueMax != nil {
		t.Error("default ValueMax should be nil")
	}
	if s.ValueNow != nil {
		t.Error("default ValueNow should be nil")
	}
	if s.ValueText != "" {
		t.Errorf("default ValueText = %q, want empty", s.ValueText)
	}
	if s.Level != 0 {
		t.Errorf("default Level = %d, want 0", s.Level)
	}
}

func TestStateHasNumericValue(t *testing.T) {
	tests := []struct {
		name  string
		state State
		want  bool
	}{
		{"no values", State{}, false},
		{"min only", State{ValueMin: Float64Ptr(0)}, true},
		{"max only", State{ValueMax: Float64Ptr(100)}, true},
		{"now only", State{ValueNow: Float64Ptr(50)}, true},
		{"all set", State{
			ValueMin: Float64Ptr(0),
			ValueMax: Float64Ptr(100),
			ValueNow: Float64Ptr(50),
		}, true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := tt.state.HasNumericValue()
			if got != tt.want {
				t.Errorf("HasNumericValue() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestStateExpandable(t *testing.T) {
	tests := []struct {
		name         string
		state        State
		isExpandable bool
		isExpanded   bool
	}{
		{"nil expanded", State{}, false, false},
		{"collapsed", State{Expanded: BoolPtr(false)}, true, false},
		{"expanded", State{Expanded: BoolPtr(true)}, true, true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := tt.state.IsExpandable(); got != tt.isExpandable {
				t.Errorf("IsExpandable() = %v, want %v", got, tt.isExpandable)
			}
			if got := tt.state.IsExpanded(); got != tt.isExpanded {
				t.Errorf("IsExpanded() = %v, want %v", got, tt.isExpanded)
			}
		})
	}
}

func TestBoolPtr(t *testing.T) {
	truePtr := BoolPtr(true)
	falsePtr := BoolPtr(false)

	if truePtr == nil || !*truePtr {
		t.Error("BoolPtr(true) should return pointer to true")
	}
	if falsePtr == nil || *falsePtr {
		t.Error("BoolPtr(false) should return pointer to false")
	}
	// Ensure distinct pointers
	if truePtr == falsePtr {
		t.Error("BoolPtr should return distinct pointers")
	}
}

func TestFloat64Ptr(t *testing.T) {
	p := Float64Ptr(3.14)
	if p == nil || *p != 3.14 {
		t.Error("Float64Ptr(3.14) should return pointer to 3.14")
	}
	z := Float64Ptr(0)
	if z == nil || *z != 0 {
		t.Error("Float64Ptr(0) should return pointer to 0")
	}
}

// --- Action tests ---

func TestActionString(t *testing.T) {
	tests := []struct {
		action Action
		want   string
	}{
		{ActionClick, "Click"},
		{ActionFocus, "Focus"},
		{ActionBlur, "Blur"},
		{ActionSetValue, "SetValue"},
		{ActionIncrement, "Increment"},
		{ActionDecrement, "Decrement"},
		{ActionExpand, "Expand"},
		{ActionCollapse, "Collapse"},
		{ActionSelect, "Select"},
		{ActionScrollIntoView, "ScrollIntoView"},
		{ActionScrollUp, "ScrollUp"},
		{ActionScrollDown, "ScrollDown"},
		{ActionScrollLeft, "ScrollLeft"},
		{ActionScrollRight, "ScrollRight"},
		{ActionShowContextMenu, "ShowContextMenu"},
		{ActionDismiss, "Dismiss"},
		{Action(0), "Unknown"},
		{Action(200), "Unknown"},
	}
	for _, tt := range tests {
		t.Run(tt.want, func(t *testing.T) {
			got := tt.action.String()
			if got != tt.want {
				t.Errorf("Action(%d).String() = %q, want %q", tt.action, got, tt.want)
			}
		})
	}
}

// --- Priority tests ---

func TestPriorityString(t *testing.T) {
	tests := []struct {
		priority Priority
		want     string
	}{
		{PriorityLow, "Low"},
		{PriorityHigh, "High"},
		{Priority(99), "Unknown"},
	}
	for _, tt := range tests {
		t.Run(tt.want, func(t *testing.T) {
			got := tt.priority.String()
			if got != tt.want {
				t.Errorf("Priority(%d).String() = %q, want %q", tt.priority, got, tt.want)
			}
		})
	}
}

// --- NodeID tests ---

func TestNextNodeIDUnique(t *testing.T) {
	seen := make(map[NodeID]bool)
	const count = 1000
	for range count {
		id := NextNodeID()
		if seen[id] {
			t.Fatalf("duplicate NodeID: %d", id)
		}
		seen[id] = true
	}
}

func TestNextNodeIDConcurrent(t *testing.T) {
	const goroutines = 10
	const idsPerGoroutine = 100

	var wg sync.WaitGroup
	ids := make([]NodeID, goroutines*idsPerGoroutine)

	for g := range goroutines {
		wg.Add(1)
		go func(offset int) {
			defer wg.Done()
			for i := range idsPerGoroutine {
				ids[offset+i] = NextNodeID()
			}
		}(g * idsPerGoroutine)
	}
	wg.Wait()

	// Check all unique
	seen := make(map[NodeID]bool, len(ids))
	for _, id := range ids {
		if !id.IsValid() {
			t.Error("got invalid (zero) NodeID")
		}
		if seen[id] {
			t.Fatalf("duplicate NodeID: %d", id)
		}
		seen[id] = true
	}
}

func TestNodeIDIsValid(t *testing.T) {
	if NodeID(0).IsValid() {
		t.Error("NodeID(0).IsValid() should be false")
	}
	if !NodeID(1).IsValid() {
		t.Error("NodeID(1).IsValid() should be true")
	}
	id := NextNodeID()
	if !id.IsValid() {
		t.Error("NextNodeID().IsValid() should be true")
	}
}

func TestNodeIDString(t *testing.T) {
	id := NodeID(42)
	want := "NodeID(42)"
	if got := id.String(); got != want {
		t.Errorf("NodeID(42).String() = %q, want %q", got, want)
	}
}

// --- Node tests ---

func TestNewNode(t *testing.T) {
	node := NewNode(RoleButton, "Save")

	if !node.ID().IsValid() {
		t.Error("new node should have a valid ID")
	}
	if node.Role() != RoleButton {
		t.Errorf("Role() = %v, want RoleButton", node.Role())
	}
	if node.Label() != "Save" {
		t.Errorf("Label() = %q, want %q", node.Label(), "Save")
	}
	if node.Hint() != "" {
		t.Errorf("Hint() = %q, want empty", node.Hint())
	}
	if node.Value() != "" {
		t.Errorf("Value() = %q, want empty", node.Value())
	}
	if node.Parent() != nil {
		t.Error("Parent() should be nil for new node")
	}
	if node.Children() != nil {
		t.Error("Children() should be nil for new node")
	}
	if node.ChildCount() != 0 {
		t.Errorf("ChildCount() = %d, want 0", node.ChildCount())
	}
}

// mockAccessible implements Accessible for testing.
type mockAccessible struct {
	role    Role
	label   string
	hint    string
	value   string
	state   State
	actions []Action
}

func (m *mockAccessible) AccessibilityRole() Role        { return m.role }
func (m *mockAccessible) AccessibilityLabel() string     { return m.label }
func (m *mockAccessible) AccessibilityHint() string      { return m.hint }
func (m *mockAccessible) AccessibilityValue() string     { return m.value }
func (m *mockAccessible) AccessibilityState() State      { return m.state }
func (m *mockAccessible) AccessibilityActions() []Action { return m.actions }

// Verify mockAccessible implements Accessible.
var _ Accessible = (*mockAccessible)(nil)

func TestNewNodeFromAccessible(t *testing.T) {
	mock := &mockAccessible{
		role:    RoleSlider,
		label:   "Volume",
		hint:    "Adjust volume",
		value:   "75%",
		state:   State{ValueNow: Float64Ptr(75)},
		actions: []Action{ActionIncrement, ActionDecrement},
	}

	node := NewNodeFromAccessible(mock)

	if !node.ID().IsValid() {
		t.Error("new node should have a valid ID")
	}
	if node.Role() != RoleSlider {
		t.Errorf("Role() = %v, want RoleSlider", node.Role())
	}
	if node.Label() != "Volume" {
		t.Errorf("Label() = %q, want %q", node.Label(), "Volume")
	}
	if node.Hint() != "Adjust volume" {
		t.Errorf("Hint() = %q, want %q", node.Hint(), "Adjust volume")
	}
	if node.Value() != "75%" {
		t.Errorf("Value() = %q, want %q", node.Value(), "75%")
	}
	state := node.State()
	if state.ValueNow == nil || *state.ValueNow != 75 {
		t.Error("State().ValueNow should be 75")
	}
	actions := node.Actions()
	if len(actions) != 2 {
		t.Fatalf("len(Actions()) = %d, want 2", len(actions))
	}
	if actions[0] != ActionIncrement || actions[1] != ActionDecrement {
		t.Errorf("Actions() = %v, want [Increment, Decrement]", actions)
	}
	if node.Source() != mock {
		t.Error("Source() should return the mock accessible")
	}
}

func TestNodeSetters(t *testing.T) {
	node := NewNode(RoleUnknown, "")

	node.SetRole(RoleCheckbox)
	if node.Role() != RoleCheckbox {
		t.Errorf("after SetRole: Role() = %v, want RoleCheckbox", node.Role())
	}

	node.SetLabel("Accept Terms")
	if node.Label() != "Accept Terms" {
		t.Errorf("after SetLabel: Label() = %q", node.Label())
	}

	node.SetHint("Toggle to accept terms of service")
	if node.Hint() != "Toggle to accept terms of service" {
		t.Errorf("after SetHint: Hint() = %q", node.Hint())
	}

	node.SetValue("on")
	if node.Value() != "on" {
		t.Errorf("after SetValue: Value() = %q", node.Value())
	}

	st := State{Checked: CheckedTrue, Disabled: true}
	node.SetState(st)
	got := node.State()
	if got.Checked != CheckedTrue {
		t.Errorf("after SetState: Checked = %v, want CheckedTrue", got.Checked)
	}
	if !got.Disabled {
		t.Error("after SetState: Disabled should be true")
	}

	node.SetActions([]Action{ActionClick, ActionFocus})
	actions := node.Actions()
	if len(actions) != 2 {
		t.Fatalf("after SetActions: len(Actions()) = %d, want 2", len(actions))
	}

	bounds := geometry.NewRect(10, 20, 100, 50)
	node.SetBounds(bounds)
	if node.Bounds() != bounds {
		t.Errorf("after SetBounds: Bounds() = %v, want %v", node.Bounds(), bounds)
	}

	mock := &mockAccessible{role: RoleButton}
	node.SetSource(mock)
	if node.Source() != mock {
		t.Error("after SetSource: Source() should return mock")
	}
}

func TestNodeActionsReturnsCopy(t *testing.T) {
	node := NewNode(RoleButton, "Test")
	node.SetActions([]Action{ActionClick})

	// Modify the returned slice
	actions := node.Actions()
	actions[0] = ActionFocus

	// Original should be unchanged
	original := node.Actions()
	if original[0] != ActionClick {
		t.Error("modifying returned Actions() should not affect the node")
	}
}

func TestNodeActionsNil(t *testing.T) {
	node := NewNode(RoleButton, "Test")
	if node.Actions() != nil {
		t.Error("node with no actions should return nil")
	}
}

func TestNodeChildrenReturnsCopy(t *testing.T) {
	parent := NewNode(RoleGroup, "Parent")
	child := NewNode(RoleButton, "Child")

	// Manually set up parent-child (normally done through tree)
	parent.mu.Lock()
	child.mu.Lock()
	parent.addChild(child)
	child.mu.Unlock()
	parent.mu.Unlock()

	children := parent.Children()
	if len(children) != 1 {
		t.Fatalf("len(Children()) = %d, want 1", len(children))
	}

	// Modify returned slice
	children[0] = nil

	// Original should be unchanged
	original := parent.Children()
	if original[0] != child {
		t.Error("modifying returned Children() should not affect the node")
	}
}

func TestNodeSyncFromSource(t *testing.T) {
	mock := &mockAccessible{
		role:  RoleButton,
		label: "Initial",
	}
	node := NewNodeFromAccessible(mock)

	// Update mock
	mock.label = "Updated"
	mock.hint = "New hint"

	if !node.SyncFromSource() {
		t.Error("SyncFromSource() should return true when source exists")
	}
	if node.Label() != "Updated" {
		t.Errorf("after sync: Label() = %q, want %q", node.Label(), "Updated")
	}
	if node.Hint() != "New hint" {
		t.Errorf("after sync: Hint() = %q, want %q", node.Hint(), "New hint")
	}
}

func TestNodeSyncFromSourceNil(t *testing.T) {
	node := NewNode(RoleButton, "Test")
	if node.SyncFromSource() {
		t.Error("SyncFromSource() should return false when source is nil")
	}
}

func TestNodeString(t *testing.T) {
	node := NewNode(RoleButton, "OK")
	s := node.String()
	want := fmt.Sprintf("Node{ID: %d, Role: Button, Label: \"OK\"}", uint64(node.ID()))
	if s != want {
		t.Errorf("String() = %q, want %q", s, want)
	}
}

func TestNodeConcurrentAccess(t *testing.T) {
	node := NewNode(RoleButton, "Concurrent")
	var wg sync.WaitGroup

	// Concurrent reads and writes
	for i := range 50 {
		wg.Add(2)
		go func(idx int) {
			defer wg.Done()
			node.SetLabel(fmt.Sprintf("Label-%d", idx))
			node.SetHint(fmt.Sprintf("Hint-%d", idx))
			node.SetValue(fmt.Sprintf("Value-%d", idx))
			node.SetState(State{Disabled: idx%2 == 0})
		}(i)
		go func() {
			defer wg.Done()
			_ = node.Label()
			_ = node.Hint()
			_ = node.Value()
			_ = node.State()
			_ = node.Actions()
			_ = node.Bounds()
			_ = node.Children()
			_ = node.Parent()
			_ = node.ChildCount()
			_ = node.String()
		}()
	}
	wg.Wait()
	// No race condition if we get here
}

// --- Tree tests ---

func TestNewMemoryTree(t *testing.T) {
	root := NewNode(RoleWindow, "App")
	tree := NewMemoryTree(root)

	if tree.Root() != root {
		t.Error("Root() should return the root node")
	}
	if tree.Len() != 1 {
		t.Errorf("Len() = %d, want 1", tree.Len())
	}
	if tree.NodeByID(root.ID()) != root {
		t.Error("NodeByID(root.ID()) should return root")
	}
}

func TestNewMemoryTreeNilRoot(t *testing.T) {
	tree := NewMemoryTree(nil)
	if tree.Root() != nil {
		t.Error("Root() should be nil for tree with nil root")
	}
	if tree.Len() != 0 {
		t.Errorf("Len() = %d, want 0", tree.Len())
	}
}

func TestTreeInsert(t *testing.T) {
	root := NewNode(RoleWindow, "App")
	tree := NewMemoryTree(root)

	child1 := NewNode(RoleButton, "Save")
	child2 := NewNode(RoleButton, "Cancel")

	tree.Insert(root, child1)
	tree.Insert(root, child2)

	if tree.Len() != 3 {
		t.Errorf("Len() = %d, want 3", tree.Len())
	}

	// Check parent-child relationships
	if child1.Parent() != root {
		t.Error("child1.Parent() should be root")
	}
	if child2.Parent() != root {
		t.Error("child2.Parent() should be root")
	}

	children := root.Children()
	if len(children) != 2 {
		t.Fatalf("root.Children() has %d children, want 2", len(children))
	}

	// Check index
	if tree.NodeByID(child1.ID()) != child1 {
		t.Error("NodeByID(child1) should find child1")
	}
	if tree.NodeByID(child2.ID()) != child2 {
		t.Error("NodeByID(child2) should find child2")
	}
}

func TestTreeInsertNested(t *testing.T) {
	root := NewNode(RoleWindow, "App")
	tree := NewMemoryTree(root)

	group := NewNode(RoleGroup, "Buttons")
	btn := NewNode(RoleButton, "OK")

	tree.Insert(root, group)
	tree.Insert(group, btn)

	if tree.Len() != 3 {
		t.Errorf("Len() = %d, want 3", tree.Len())
	}
	if btn.Parent() != group {
		t.Error("btn.Parent() should be group")
	}
	if tree.NodeByID(btn.ID()) != btn {
		t.Error("NodeByID should find deeply nested node")
	}
}

func TestTreeInsertSubtree(t *testing.T) {
	root := NewNode(RoleWindow, "App")
	tree := NewMemoryTree(root)

	// Build a subtree before inserting
	group := NewNode(RoleGroup, "Group")
	child1 := NewNode(RoleButton, "A")
	child2 := NewNode(RoleButton, "B")
	group.mu.Lock()
	child1.mu.Lock()
	group.addChild(child1)
	child1.mu.Unlock()
	child2.mu.Lock()
	group.addChild(child2)
	child2.mu.Unlock()
	group.mu.Unlock()

	// Insert the entire subtree
	tree.Insert(root, group)

	if tree.Len() != 4 {
		t.Errorf("Len() = %d, want 4 (root + group + 2 children)", tree.Len())
	}
	if tree.NodeByID(child1.ID()) != child1 {
		t.Error("subtree children should be indexed after Insert")
	}
	if tree.NodeByID(child2.ID()) != child2 {
		t.Error("subtree children should be indexed after Insert")
	}
}

func TestTreeInsertNilArguments(t *testing.T) {
	root := NewNode(RoleWindow, "App")
	tree := NewMemoryTree(root)
	child := NewNode(RoleButton, "Test")

	// These should not panic
	tree.Insert(nil, child)
	tree.Insert(root, nil)
	tree.Insert(nil, nil)

	if tree.Len() != 1 {
		t.Errorf("Len() = %d after nil inserts, want 1", tree.Len())
	}
}

func TestTreeRemove(t *testing.T) {
	root := NewNode(RoleWindow, "App")
	tree := NewMemoryTree(root)

	child := NewNode(RoleButton, "Save")
	tree.Insert(root, child)

	tree.Remove(child)

	if tree.Len() != 1 {
		t.Errorf("Len() = %d after remove, want 1", tree.Len())
	}
	if tree.NodeByID(child.ID()) != nil {
		t.Error("removed node should not be in index")
	}
	if child.Parent() != nil {
		t.Error("removed node should have nil parent")
	}
	if len(root.Children()) != 0 {
		t.Error("root should have no children after remove")
	}
}

func TestTreeRemoveSubtree(t *testing.T) {
	root := NewNode(RoleWindow, "App")
	tree := NewMemoryTree(root)

	group := NewNode(RoleGroup, "Group")
	child := NewNode(RoleButton, "Child")
	tree.Insert(root, group)
	tree.Insert(group, child)

	// Remove group should also remove child
	tree.Remove(group)

	if tree.Len() != 1 {
		t.Errorf("Len() = %d after subtree remove, want 1", tree.Len())
	}
	if tree.NodeByID(group.ID()) != nil {
		t.Error("removed group should not be in index")
	}
	if tree.NodeByID(child.ID()) != nil {
		t.Error("descendant of removed group should not be in index")
	}
}

func TestTreeRemoveRoot(t *testing.T) {
	root := NewNode(RoleWindow, "App")
	tree := NewMemoryTree(root)
	child := NewNode(RoleButton, "Child")
	tree.Insert(root, child)

	tree.Remove(root)

	if tree.Root() != nil {
		t.Error("Root() should be nil after removing root")
	}
	if tree.Len() != 0 {
		t.Errorf("Len() = %d after removing root, want 0", tree.Len())
	}
}

func TestTreeRemoveNil(t *testing.T) {
	root := NewNode(RoleWindow, "App")
	tree := NewMemoryTree(root)

	// Should not panic
	tree.Remove(nil)

	if tree.Len() != 1 {
		t.Error("removing nil should not change tree")
	}
}

func TestTreeNodeByIDNotFound(t *testing.T) {
	root := NewNode(RoleWindow, "App")
	tree := NewMemoryTree(root)

	if tree.NodeByID(NodeID(999999)) != nil {
		t.Error("NodeByID should return nil for unknown ID")
	}
}

func TestTreeWalk(t *testing.T) {
	root := NewNode(RoleWindow, "Root")
	tree := NewMemoryTree(root)

	child1 := NewNode(RoleButton, "A")
	child2 := NewNode(RoleButton, "B")
	grandchild := NewNode(RoleLabel, "C")

	tree.Insert(root, child1)
	tree.Insert(root, child2)
	tree.Insert(child1, grandchild)

	var visited []string
	tree.Walk(func(n *Node) bool {
		visited = append(visited, n.Label())
		return true
	})

	if len(visited) != 4 {
		t.Fatalf("Walk visited %d nodes, want 4", len(visited))
	}
	// Root should be first (depth-first, parent before children)
	if visited[0] != "Root" {
		t.Errorf("first visited = %q, want %q", visited[0], "Root")
	}
}

func TestTreeWalkEarlyStop(t *testing.T) {
	root := NewNode(RoleWindow, "Root")
	tree := NewMemoryTree(root)

	for i := range 10 {
		tree.Insert(root, NewNode(RoleButton, fmt.Sprintf("Button%d", i)))
	}

	count := 0
	tree.Walk(func(_ *Node) bool {
		count++
		return count < 3 // Stop after 3 nodes
	})

	if count != 3 {
		t.Errorf("Walk visited %d nodes, want 3 (early stop)", count)
	}
}

func TestTreeWalkEmpty(t *testing.T) {
	tree := NewMemoryTree(nil)
	count := 0
	tree.Walk(func(_ *Node) bool {
		count++
		return true
	})
	if count != 0 {
		t.Errorf("Walk on empty tree visited %d nodes, want 0", count)
	}
}

func TestTreeUpdate(t *testing.T) {
	root := NewNode(RoleWindow, "App")
	tree := NewMemoryTree(root)

	child := NewNode(RoleButton, "Save")
	tree.Insert(root, child)

	// No dirty nodes initially
	if dirty := tree.DirtyNodes(); len(dirty) != 0 {
		t.Errorf("initial DirtyNodes() = %d, want 0", len(dirty))
	}

	// Mark node as dirty
	tree.Update(child)

	dirty := tree.DirtyNodes()
	if len(dirty) != 1 {
		t.Fatalf("DirtyNodes() has %d nodes, want 1", len(dirty))
	}
	if dirty[0] != child {
		t.Error("dirty node should be child")
	}

	// Clear dirty
	tree.ClearDirty()
	if dirty := tree.DirtyNodes(); len(dirty) != 0 {
		t.Errorf("after ClearDirty: DirtyNodes() = %d, want 0", len(dirty))
	}
}

func TestTreeUpdateNil(t *testing.T) {
	root := NewNode(RoleWindow, "App")
	tree := NewMemoryTree(root)

	// Should not panic
	tree.Update(nil)

	if dirty := tree.DirtyNodes(); len(dirty) != 0 {
		t.Error("Update(nil) should not add dirty node")
	}
}

func TestTreeDirtyNodesAfterRemove(t *testing.T) {
	root := NewNode(RoleWindow, "App")
	tree := NewMemoryTree(root)

	child := NewNode(RoleButton, "Save")
	tree.Insert(root, child)
	tree.Update(child)

	// Remove the dirty node
	tree.Remove(child)

	// Dirty set should be cleared for removed node
	if dirty := tree.DirtyNodes(); len(dirty) != 0 {
		t.Errorf("DirtyNodes() after remove = %d, want 0", len(dirty))
	}
}

func TestTreeConcurrentAccess(t *testing.T) {
	root := NewNode(RoleWindow, "App")
	tree := NewMemoryTree(root)

	var wg sync.WaitGroup

	// Concurrent inserts
	for i := range 20 {
		wg.Add(1)
		go func(idx int) {
			defer wg.Done()
			child := NewNode(RoleButton, fmt.Sprintf("Btn%d", idx))
			tree.Insert(root, child)
		}(i)
	}
	wg.Wait()

	if tree.Len() != 21 { // root + 20 children
		t.Errorf("Len() = %d, want 21", tree.Len())
	}

	// Concurrent reads
	for range 20 {
		wg.Add(3)
		go func() {
			defer wg.Done()
			_ = tree.Root()
		}()
		go func() {
			defer wg.Done()
			_ = tree.Len()
		}()
		go func() {
			defer wg.Done()
			tree.Walk(func(_ *Node) bool { return true })
		}()
	}
	wg.Wait()
}

// --- Announcer tests ---

func TestNoOpAnnouncerDoesNotPanic(t *testing.T) {
	var announcer NoOpAnnouncer

	// Should not panic with any input
	announcer.Announce("test message", PriorityLow)
	announcer.Announce("urgent message", PriorityHigh)
	announcer.Announce("", PriorityLow)
	announcer.Announce("", PriorityHigh)
}

func TestNoOpAnnouncerImplementsInterface(t *testing.T) {
	var a Announcer = NoOpAnnouncer{}
	a.Announce("test", PriorityLow)
	// Should not panic
}

func TestNoOpAnnouncerConcurrent(t *testing.T) {
	var announcer NoOpAnnouncer
	var wg sync.WaitGroup

	for i := range 100 {
		wg.Add(1)
		go func(idx int) {
			defer wg.Done()
			announcer.Announce(fmt.Sprintf("message %d", idx), PriorityLow)
		}(i)
	}
	wg.Wait()
}

// --- Compile-time interface checks ---

// Verify MemoryTree implements TreeProvider.
var _ TreeProvider = (*MemoryTree)(nil)

// Verify NoOpAnnouncer implements Announcer.
var _ Announcer = NoOpAnnouncer{}

// Verify mockAccessible implements Accessible.
// (Already declared above, but included here for clarity)
