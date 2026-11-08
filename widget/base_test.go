package widget

import (
	"sync"
	"testing"

	"github.com/gogpu/ui/event"
	"github.com/gogpu/ui/geometry"
)

// mockWidget is a minimal widget implementation for testing.
type mockWidget struct {
	WidgetBase
}

func (m *mockWidget) Layout(_ Context, c geometry.Constraints) geometry.Size {
	return c.Constrain(geometry.Sz(100, 50))
}

func (m *mockWidget) Draw(_ Context, _ Canvas) {}

func (m *mockWidget) Event(_ Context, _ event.Event) bool {
	return false
}

func newMockWidget() *mockWidget {
	w := &mockWidget{}
	w.SetVisible(true)
	w.SetEnabled(true)
	return w
}

// Verify mockWidget implements Widget interface.
var _ Widget = (*mockWidget)(nil)

func TestNewWidgetBase(t *testing.T) {
	w := NewWidgetBase()
	if w == nil {
		t.Fatal("NewWidgetBase returned nil")
	}
	if !w.IsVisible() {
		t.Error("expected visible=true by default")
	}
	if !w.IsEnabled() {
		t.Error("expected enabled=true by default")
	}
	if w.IsFocused() {
		t.Error("expected focused=false by default")
	}
	if w.ID() != "" {
		t.Error("expected empty ID by default")
	}
	if w.Children() != nil {
		t.Error("expected no children by default")
	}
}

func TestWidgetBase_Bounds(t *testing.T) {
	tests := []struct {
		name   string
		bounds geometry.Rect
	}{
		{
			name:   "zero bounds",
			bounds: geometry.Rect{},
		},
		{
			name:   "positive bounds",
			bounds: geometry.NewRect(10, 20, 100, 50),
		},
		{
			name:   "large bounds",
			bounds: geometry.NewRect(0, 0, 1920, 1080),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			w := NewWidgetBase()
			w.SetBounds(tt.bounds)
			got := w.Bounds()
			if got != tt.bounds {
				t.Errorf("Bounds() = %v, want %v", got, tt.bounds)
			}
		})
	}
}

func TestWidgetBase_Size(t *testing.T) {
	w := NewWidgetBase()
	w.SetBounds(geometry.NewRect(10, 20, 100, 50))
	got := w.Size()
	want := geometry.Sz(100, 50)
	if got != want {
		t.Errorf("Size() = %v, want %v", got, want)
	}
}

func TestWidgetBase_Position(t *testing.T) {
	w := NewWidgetBase()
	w.SetBounds(geometry.NewRect(10, 20, 100, 50))
	got := w.Position()
	want := geometry.Pt(10, 20)
	if got != want {
		t.Errorf("Position() = %v, want %v", got, want)
	}
}

func TestWidgetBase_Focus(t *testing.T) {
	w := NewWidgetBase()

	// Initially not focused
	if w.IsFocused() {
		t.Error("expected not focused initially")
	}

	// Set focused
	w.SetFocused(true)
	if !w.IsFocused() {
		t.Error("expected focused after SetFocused(true)")
	}

	// Clear focus
	w.SetFocused(false)
	if w.IsFocused() {
		t.Error("expected not focused after SetFocused(false)")
	}
}

func TestWidgetBase_Visibility(t *testing.T) {
	w := NewWidgetBase()

	// Initially visible
	if !w.IsVisible() {
		t.Error("expected visible initially")
	}

	// Hide
	w.SetVisible(false)
	if w.IsVisible() {
		t.Error("expected not visible after SetVisible(false)")
	}

	// Show again
	w.SetVisible(true)
	if !w.IsVisible() {
		t.Error("expected visible after SetVisible(true)")
	}
}

func TestWidgetBase_Enabled(t *testing.T) {
	w := NewWidgetBase()

	// Initially enabled
	if !w.IsEnabled() {
		t.Error("expected enabled initially")
	}

	// Disable
	w.SetEnabled(false)
	if w.IsEnabled() {
		t.Error("expected not enabled after SetEnabled(false)")
	}

	// Enable again
	w.SetEnabled(true)
	if !w.IsEnabled() {
		t.Error("expected enabled after SetEnabled(true)")
	}
}

func TestWidgetBase_ID(t *testing.T) {
	w := NewWidgetBase()

	// Initially empty
	if w.ID() != "" {
		t.Error("expected empty ID initially")
	}

	// Set ID
	testID := "my-button-1"
	w.SetID(testID)
	if w.ID() != testID {
		t.Errorf("ID() = %q, want %q", w.ID(), testID)
	}

	// Clear ID
	w.SetID("")
	if w.ID() != "" {
		t.Error("expected empty ID after SetID('')")
	}
}

func TestWidgetBase_Parent(t *testing.T) {
	parent := newMockWidget()
	child := newMockWidget()

	// Initially no parent
	if child.Parent() != nil {
		t.Error("expected no parent initially")
	}

	// Set parent
	child.SetParent(parent)
	if child.Parent() != parent {
		t.Error("Parent() should return the set parent")
	}

	// Clear parent
	child.SetParent(nil)
	if child.Parent() != nil {
		t.Error("expected no parent after SetParent(nil)")
	}
}

func TestWidgetBase_Children(t *testing.T) {
	parent := NewWidgetBase()

	// Initially no children
	if parent.Children() != nil {
		t.Error("expected nil children initially")
	}
	if parent.ChildCount() != 0 {
		t.Error("expected child count 0")
	}
	if parent.HasChildren() {
		t.Error("expected HasChildren() = false")
	}

	// Add children
	child1 := newMockWidget()
	child2 := newMockWidget()
	parent.AddChild(child1)
	parent.AddChild(child2)

	if parent.ChildCount() != 2 {
		t.Errorf("ChildCount() = %d, want 2", parent.ChildCount())
	}
	if !parent.HasChildren() {
		t.Error("expected HasChildren() = true")
	}

	children := parent.Children()
	if len(children) != 2 {
		t.Errorf("Children() len = %d, want 2", len(children))
	}
	if children[0] != child1 {
		t.Error("Children()[0] should be child1")
	}
	if children[1] != child2 {
		t.Error("Children()[1] should be child2")
	}
}

func TestWidgetBase_AddChild_Nil(t *testing.T) {
	parent := NewWidgetBase()
	parent.AddChild(nil) // Should be a no-op
	if parent.ChildCount() != 0 {
		t.Error("adding nil child should be a no-op")
	}
}

func TestWidgetBase_ChildAt(t *testing.T) {
	parent := NewWidgetBase()
	child1 := newMockWidget()
	child2 := newMockWidget()
	parent.AddChild(child1)
	parent.AddChild(child2)

	// Valid indices
	if parent.ChildAt(0) != child1 {
		t.Error("ChildAt(0) should return child1")
	}
	if parent.ChildAt(1) != child2 {
		t.Error("ChildAt(1) should return child2")
	}

	// Invalid indices
	if parent.ChildAt(-1) != nil {
		t.Error("ChildAt(-1) should return nil")
	}
	if parent.ChildAt(2) != nil {
		t.Error("ChildAt(2) should return nil")
	}
	if parent.ChildAt(100) != nil {
		t.Error("ChildAt(100) should return nil")
	}
}

func TestWidgetBase_RemoveChild(t *testing.T) {
	parent := NewWidgetBase()
	child1 := newMockWidget()
	child2 := newMockWidget()
	child3 := newMockWidget()

	parent.AddChild(child1)
	parent.AddChild(child2)
	parent.AddChild(child3)

	// Remove the FIRST child. The old implementation swap-removed (moving the
	// last element into the gap => [child3, child2]); order preservation is
	// observable only when the removed element is not last.
	if !parent.RemoveChild(child1) {
		t.Error("RemoveChild should return true for existing child")
	}
	if parent.ChildCount() != 2 {
		t.Error("expected 2 children after removal")
	}

	remaining := parent.Children()
	if len(remaining) != 2 {
		t.Fatalf("Children() len = %d, want 2", len(remaining))
	}
	if remaining[0] != child2 || remaining[1] != child3 {
		t.Errorf("RemoveChild must preserve order: got [%p %p], want [child2 child3]",
			remaining[0], remaining[1])
	}

	// Remove non-existent child
	if parent.RemoveChild(child1) {
		t.Error("RemoveChild should return false for non-existent child")
	}

	// Remove nil
	if parent.RemoveChild(nil) {
		t.Error("RemoveChild(nil) should return false")
	}
}

// fakeUnbinder counts Unbind calls from the binding lifecycle test.
type fakeUnbinder struct{ calls int }

func (f *fakeUnbinder) Unbind() { f.calls++ }

// bindingWidget mounts a signal-like binding in Mount so RemoveChild can be
// observed to release it via WidgetBase.CleanupBindings inside UnmountTree.
type bindingWidget struct {
	WidgetBase
	unbinder *fakeUnbinder
}

func (b *bindingWidget) Layout(_ Context, c geometry.Constraints) geometry.Size {
	return c.Smallest()
}

func (b *bindingWidget) Draw(_ Context, _ Canvas) {}

func (b *bindingWidget) Event(_ Context, _ event.Event) bool { return false }

func (b *bindingWidget) Mount(_ Context) { b.AddBinding(b.unbinder) }

func (b *bindingWidget) Unmount() {}

func TestWidgetBase_RemoveChildUnmountsBindings(t *testing.T) {
	parent := NewWidgetBase()
	ctx := NewContext()
	child := &bindingWidget{unbinder: &fakeUnbinder{}}

	// Mount the child so it holds a live binding.
	MountTree(child, ctx)
	if !child.IsMounted() {
		t.Fatal("precondition: child should be mounted")
	}

	// WidgetBase on its own has no SetSelf container, so AddChild stores the
	// child without re-mounting it; the child stays mounted from the call above.
	parent.AddChild(child)

	if !parent.RemoveChild(child) {
		t.Fatal("RemoveChild should return true for an existing child")
	}
	if child.unbinder.calls != 1 {
		t.Errorf("Unbind calls = %d, want 1 (bindings must be released on removal)",
			child.unbinder.calls)
	}
	if child.IsMounted() {
		t.Error("removed child should be unmounted")
	}
}

func TestWidgetBase_RemoveChildAt(t *testing.T) {
	parent := NewWidgetBase()
	child1 := newMockWidget()
	child2 := newMockWidget()
	child3 := newMockWidget()

	parent.AddChild(child1)
	parent.AddChild(child2)
	parent.AddChild(child3)

	// Remove middle child (preserves order)
	removed := parent.RemoveChildAt(1)
	if removed != child2 {
		t.Error("RemoveChildAt(1) should return child2")
	}
	if parent.ChildCount() != 2 {
		t.Error("expected 2 children after removal")
	}
	if parent.ChildAt(0) != child1 || parent.ChildAt(1) != child3 {
		t.Error("RemoveChildAt should preserve order")
	}

	// Invalid indices
	if parent.RemoveChildAt(-1) != nil {
		t.Error("RemoveChildAt(-1) should return nil")
	}
	if parent.RemoveChildAt(100) != nil {
		t.Error("RemoveChildAt(100) should return nil")
	}
}

func TestWidgetBase_ClearChildren(t *testing.T) {
	parent := NewWidgetBase()
	parent.AddChild(newMockWidget())
	parent.AddChild(newMockWidget())
	parent.AddChild(newMockWidget())

	if parent.ChildCount() != 3 {
		t.Error("expected 3 children")
	}

	parent.ClearChildren()

	if parent.ChildCount() != 0 {
		t.Error("expected 0 children after ClearChildren")
	}
	if parent.HasChildren() {
		t.Error("expected HasChildren() = false after ClearChildren")
	}
	if parent.Children() != nil {
		t.Error("expected nil from Children() after ClearChildren")
	}
}

func TestWidgetBase_InsertChild(t *testing.T) {
	parent := NewWidgetBase()
	child1 := newMockWidget()
	child2 := newMockWidget()
	child3 := newMockWidget()

	parent.AddChild(child1)
	parent.AddChild(child3)

	// Insert in middle
	parent.InsertChild(1, child2)
	if parent.ChildAt(0) != child1 {
		t.Error("child1 should be at index 0")
	}
	if parent.ChildAt(1) != child2 {
		t.Error("child2 should be at index 1")
	}
	if parent.ChildAt(2) != child3 {
		t.Error("child3 should be at index 2")
	}

	// Insert at beginning
	child0 := newMockWidget()
	parent.InsertChild(0, child0)
	if parent.ChildAt(0) != child0 {
		t.Error("child0 should be at index 0")
	}

	// Insert past end (appends)
	childN := newMockWidget()
	parent.InsertChild(100, childN)
	if parent.ChildAt(parent.ChildCount()-1) != childN {
		t.Error("childN should be at last index")
	}

	// Insert with negative index (uses 0)
	childNeg := newMockWidget()
	parent.InsertChild(-5, childNeg)
	if parent.ChildAt(0) != childNeg {
		t.Error("childNeg should be at index 0")
	}

	// Insert nil (should be no-op)
	count := parent.ChildCount()
	parent.InsertChild(0, nil)
	if parent.ChildCount() != count {
		t.Error("inserting nil should be a no-op")
	}
}

func TestWidgetBase_ContainsPoint(t *testing.T) {
	w := NewWidgetBase()
	w.SetBounds(geometry.NewRect(10, 20, 100, 50))

	tests := []struct {
		name     string
		point    geometry.Point
		contains bool
	}{
		{"inside", geometry.Pt(50, 40), true},
		{"top-left corner", geometry.Pt(10, 20), true},
		{"bottom-right corner", geometry.Pt(110, 70), true},
		{"outside left", geometry.Pt(5, 40), false},
		{"outside right", geometry.Pt(115, 40), false},
		{"outside top", geometry.Pt(50, 15), false},
		{"outside bottom", geometry.Pt(50, 75), false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := w.ContainsPoint(tt.point)
			if got != tt.contains {
				t.Errorf("ContainsPoint(%v) = %v, want %v", tt.point, got, tt.contains)
			}
		})
	}
}

func TestWidgetBase_CoordinateTransform(t *testing.T) {
	w := NewWidgetBase()
	w.SetBounds(geometry.NewRect(100, 200, 50, 50))

	// Set screen origin (as the framework would during Draw)
	w.SetScreenOrigin(geometry.Pt(100, 200))

	// Local to global
	local := geometry.Pt(10, 20)
	global := w.LocalToGlobal(local)
	wantGlobal := geometry.Pt(110, 220)
	if global != wantGlobal {
		t.Errorf("LocalToGlobal(%v) = %v, want %v", local, global, wantGlobal)
	}

	// Global to local
	gotLocal := w.GlobalToLocal(global)
	if gotLocal != local {
		t.Errorf("GlobalToLocal(%v) = %v, want %v", global, gotLocal, local)
	}
}

func TestWidgetBase_ScreenOrigin(t *testing.T) {
	w := NewWidgetBase()
	w.SetBounds(geometry.NewRect(10, 20, 80, 40))

	// Before any Draw pass, screen origin is zero.
	if got := w.ScreenOrigin(); got != (geometry.Point{}) {
		t.Errorf("ScreenOrigin before Draw = %v, want (0,0)", got)
	}

	// ScreenBounds before Draw uses zero origin.
	sb := w.ScreenBounds()
	if sb.Min != (geometry.Point{}) {
		t.Errorf("ScreenBounds.Min before Draw = %v, want (0,0)", sb.Min)
	}
	if sb.Width() != 80 || sb.Height() != 40 {
		t.Errorf("ScreenBounds size = (%v,%v), want (80,40)", sb.Width(), sb.Height())
	}

	// After the framework stamps screen origin during Draw
	w.SetScreenOrigin(geometry.Pt(50, 100))
	if got := w.ScreenOrigin(); got != (geometry.Pt(50, 100)) {
		t.Errorf("ScreenOrigin = %v, want (50,100)", got)
	}

	sb = w.ScreenBounds()
	if sb.Min != (geometry.Pt(50, 100)) {
		t.Errorf("ScreenBounds.Min = %v, want (50,100)", sb.Min)
	}
	if sb.Max != (geometry.Pt(130, 140)) {
		t.Errorf("ScreenBounds.Max = %v, want (130,140)", sb.Max)
	}
}

func TestWidgetBase_CoordinateTransformBeforeDraw(t *testing.T) {
	// Before Draw pass, screenOrigin is zero, so LocalToGlobal
	// returns the point unchanged.
	w := NewWidgetBase()
	w.SetBounds(geometry.NewRect(100, 200, 50, 50))

	local := geometry.Pt(10, 20)
	global := w.LocalToGlobal(local)
	if global != local {
		t.Errorf("LocalToGlobal before Draw = %v, want %v (screenOrigin is zero)", global, local)
	}
}

func TestWidgetBase_ThreadSafety(t *testing.T) {
	w := NewWidgetBase()
	w.SetBounds(geometry.NewRect(0, 0, 100, 100))

	var wg sync.WaitGroup
	const numGoroutines = 100

	// Test concurrent reads
	for i := 0; i < numGoroutines; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			_ = w.Bounds()
			_ = w.IsFocused()
			_ = w.IsVisible()
			_ = w.IsEnabled()
			_ = w.ID()
			_ = w.Children()
		}()
	}

	// Test concurrent writes
	for i := 0; i < numGoroutines; i++ {
		wg.Add(1)
		go func(i int) {
			defer wg.Done()
			w.SetVisible(i%2 == 0)
			w.SetEnabled(i%2 == 1)
			w.SetFocused(i%2 == 0)
		}(i)
	}

	wg.Wait()
	// If we get here without deadlock or panic, the test passes
}

func TestWidgetBase_Children_ReturnsCopy(t *testing.T) {
	parent := NewWidgetBase()
	child1 := newMockWidget()
	child2 := newMockWidget()
	parent.AddChild(child1)
	parent.AddChild(child2)

	// Get children and modify the slice
	children := parent.Children()
	children[0] = nil

	// Original should be unchanged
	actual := parent.Children()
	if actual[0] != child1 {
		t.Error("Children() should return a copy, not the original slice")
	}
}

// mockStopper is a test helper implementing Stopper.
type mockStopper struct {
	stopped bool
}

func (m *mockStopper) Stop() {
	m.stopped = true
}

func TestWidgetBase_MountedState(t *testing.T) {
	w := NewWidgetBase()
	if w.IsMounted() {
		t.Error("expected not mounted initially")
	}

	w.SetMounted(true)
	if !w.IsMounted() {
		t.Error("expected mounted after SetMounted(true)")
	}

	w.SetMounted(false)
	if w.IsMounted() {
		t.Error("expected not mounted after SetMounted(false)")
	}
}

func TestWidgetBase_AddBinding(t *testing.T) {
	w := NewWidgetBase()

	unbindCalled := 0
	b := &testUnbinder{fn: func() { unbindCalled++ }}
	w.AddBinding(b)

	// nil binding should be ignored
	w.AddBinding(nil)

	w.CleanupBindings()
	if unbindCalled != 1 {
		t.Errorf("unbindCalled = %d, want 1", unbindCalled)
	}

	// Second cleanup should be no-op (bindings already cleared).
	w.CleanupBindings()
	if unbindCalled != 1 {
		t.Errorf("unbindCalled = %d after second cleanup, want 1", unbindCalled)
	}
}

func TestWidgetBase_AddEffect(t *testing.T) {
	w := NewWidgetBase()

	s := &mockStopper{}
	w.AddEffect(s)

	// nil effect should be ignored
	w.AddEffect(nil)

	w.CleanupBindings()
	if !s.stopped {
		t.Error("effect should have been stopped on cleanup")
	}
}

// testUnbinder is a test helper implementing Unbinder.
type testUnbinder struct {
	fn func()
}

func (u *testUnbinder) Unbind() {
	if u.fn != nil {
		u.fn()
	}
}

func BenchmarkWidgetBase_Bounds(b *testing.B) {
	w := NewWidgetBase()
	w.SetBounds(geometry.NewRect(10, 20, 100, 50))

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = w.Bounds()
	}
}

func BenchmarkWidgetBase_SetBounds(b *testing.B) {
	w := NewWidgetBase()
	bounds := geometry.NewRect(10, 20, 100, 50)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		w.SetBounds(bounds)
	}
}

func BenchmarkWidgetBase_Children(b *testing.B) {
	parent := NewWidgetBase()
	for i := 0; i < 10; i++ {
		parent.AddChild(newMockWidget())
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = parent.Children()
	}
}

func BenchmarkWidgetBase_ContainsPoint(b *testing.B) {
	w := NewWidgetBase()
	w.SetBounds(geometry.NewRect(10, 20, 100, 50))
	point := geometry.Pt(50, 40)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = w.ContainsPoint(point)
	}
}

// --- Upward Dirty Propagation Tests (ADR-007 Task 1d) ---

// mockBoundary is a mock RepaintBoundary for testing upward dirty propagation.
type mockBoundary struct {
	WidgetBase
	dirtyCount int
}

func (m *mockBoundary) Layout(_ Context, c geometry.Constraints) geometry.Size {
	return c.Constrain(geometry.Sz(200, 200))
}

func (m *mockBoundary) Draw(_ Context, _ Canvas) {}

func (m *mockBoundary) Event(_ Context, _ event.Event) bool { return false }

func (m *mockBoundary) MarkBoundaryDirty() {
	m.dirtyCount++
}

// Verify mockBoundary implements RepaintBoundaryMarker.
var _ RepaintBoundaryMarker = (*mockBoundary)(nil)

func newMockBoundary() *mockBoundary {
	b := &mockBoundary{}
	b.SetVisible(true)
	b.SetEnabled(true)
	return b
}

// TestSetNeedsRedraw_UpwardPropagation verifies that SetNeedsRedraw(true)
// on a deep widget propagates UP to the nearest RepaintBoundary.
func TestSetNeedsRedraw_UpwardPropagation(t *testing.T) {
	// Tree: boundary -> parent -> child
	boundary := newMockBoundary()
	parent := newMockWidget()
	child := newMockWidget()

	parent.SetParent(boundary)
	child.SetParent(parent)

	// Set dirty on the leaf.
	child.SetNeedsRedraw(true)

	if boundary.dirtyCount != 1 {
		t.Errorf("expected boundary MarkBoundaryDirty called once, got %d", boundary.dirtyCount)
	}

	// Flutter pattern: intermediate parent should NOT be marked dirty.
	// Only the boundary receives the dirty notification. Marking
	// intermediates causes CollectDirtyRegions to report full parent
	// chain as dirty → full screen damage overlay (ADR-028).
	if parent.NeedsRedraw() {
		t.Error("intermediate parent should NOT be dirty — only boundary gets marked (Flutter markNeedsPaint)")
	}
}

// TestSetNeedsRedraw_O1Guard verifies that setting dirty twice does NOT
// cause double propagation (O(1) guard).
func TestSetNeedsRedraw_O1Guard(t *testing.T) {
	boundary := newMockBoundary()
	child := newMockWidget()
	child.SetParent(boundary)

	// First set: triggers propagation.
	child.SetNeedsRedraw(true)
	if boundary.dirtyCount != 1 {
		t.Fatalf("expected 1 propagation, got %d", boundary.dirtyCount)
	}

	// Second set: should NOT propagate again (already dirty).
	child.SetNeedsRedraw(true)
	if boundary.dirtyCount != 1 {
		t.Errorf("expected no additional propagation (O(1) guard), got %d", boundary.dirtyCount)
	}
}

// TestSetNeedsRedraw_ClearAndRepropagate verifies that after clearing the
// dirty flag, a subsequent SetNeedsRedraw(true) propagates again.
func TestSetNeedsRedraw_ClearAndRepropagate(t *testing.T) {
	boundary := newMockBoundary()
	child := newMockWidget()
	child.SetParent(boundary)

	child.SetNeedsRedraw(true)
	if boundary.dirtyCount != 1 {
		t.Fatalf("expected 1 propagation, got %d", boundary.dirtyCount)
	}

	// Clear the flag (simulating draw pass).
	child.ClearRedraw()

	// Set dirty again — should propagate.
	child.SetNeedsRedraw(true)
	if boundary.dirtyCount != 2 {
		t.Errorf("expected 2 total propagations after clear+set, got %d", boundary.dirtyCount)
	}
}

// TestSetNeedsRedraw_NoParent verifies that SetNeedsRedraw works without
// parent chain (no panic, no propagation).
func TestSetNeedsRedraw_NoParent(t *testing.T) {
	child := newMockWidget()

	// Should not panic without parent.
	child.SetNeedsRedraw(true)

	if !child.NeedsRedraw() {
		t.Error("widget should be marked dirty even without parent")
	}
}

// TestSetNeedsRedraw_SetFalseNoPropagation verifies that
// SetNeedsRedraw(false) does NOT propagate upward.
func TestSetNeedsRedraw_SetFalseNoPropagation(t *testing.T) {
	boundary := newMockBoundary()
	child := newMockWidget()
	child.SetParent(boundary)

	// Clear dirty — should not propagate.
	child.SetNeedsRedraw(false)

	if boundary.dirtyCount != 0 {
		t.Errorf("SetNeedsRedraw(false) should not propagate, got %d calls", boundary.dirtyCount)
	}
}

// TestSetNeedsRedraw_DeepTree verifies propagation through a deep tree with
// multiple intermediate widgets and a boundary in the middle.
func TestSetNeedsRedraw_DeepTree(t *testing.T) {
	// Tree: root -> boundary -> mid1 -> mid2 -> leaf
	root := newMockWidget()
	boundary := newMockBoundary()
	mid1 := newMockWidget()
	mid2 := newMockWidget()
	leaf := newMockWidget()

	boundary.SetParent(root)
	mid1.SetParent(boundary)
	mid2.SetParent(mid1)
	leaf.SetParent(mid2)

	leaf.SetNeedsRedraw(true)

	// Boundary should be notified.
	if boundary.dirtyCount != 1 {
		t.Errorf("boundary should receive dirty notification, got %d", boundary.dirtyCount)
	}

	// Flutter pattern: mid widgets between leaf and boundary should NOT
	// be dirty. Only boundary gets marked (ADR-028).
	if mid1.NeedsRedraw() {
		t.Error("mid1 should NOT be dirty — only boundary gets marked (Flutter markNeedsPaint)")
	}
	if mid2.NeedsRedraw() {
		t.Error("mid2 should NOT be dirty — only boundary gets marked (Flutter markNeedsPaint)")
	}

	// Root (above boundary) should NOT be dirty — propagation stops at boundary.
	if root.NeedsRedraw() {
		t.Error("root should NOT be dirty — propagation stops at RepaintBoundary")
	}
}

// TestSetNeedsRedraw_NoBoundary verifies propagation when there is no
// RepaintBoundary in the parent chain. Without a boundary, no ancestor
// gets marked dirty — the propagation walks to root and finds no boundary.
// The widget itself remains dirty (SetNeedsRedraw sets its own flag).
func TestSetNeedsRedraw_NoBoundary(t *testing.T) {
	root := newMockWidget()
	child := newMockWidget()
	child.SetParent(root)

	child.SetNeedsRedraw(true)

	// Flutter pattern: without boundary, intermediates not marked.
	// Root widget detects dirty descendants via NeedsRedrawInTreeNonBoundary.
	if root.NeedsRedraw() {
		t.Error("root should NOT be dirty — no boundary found, intermediates not marked (Flutter pattern)")
	}
	if !child.NeedsRedraw() {
		t.Error("child itself should remain dirty")
	}
}
