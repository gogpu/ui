package primitives_test

import (
	"testing"

	"github.com/gogpu/ui/geometry"
	"github.com/gogpu/ui/primitives"
)

// TestBox_MoveChildReorders verifies the reorder operation preserves order and
// does not detach (the moved child keeps its identity).
func TestBox_MoveChildReorders(t *testing.T) {
	a := newRecordLeaf("A", geometry.Sz(10, 10))
	b := newRecordLeaf("B", geometry.Sz(10, 10))
	c := newRecordLeaf("C", geometry.Sz(10, 10))
	box := primitives.Box(a, b, c)

	// Move A (index 0) to the end.
	if !box.MoveChild(0, 2) {
		t.Fatal("MoveChild(0, 2) returned false")
	}
	got := box.Children()
	if len(got) != 3 || got[0] != b || got[1] != c || got[2] != a {
		t.Fatalf("after MoveChild(0,2) children = %v, want [B C A]", leafNames(got))
	}

	// The moved child is the same object (no detach/re-create).
	if got[2] != a {
		t.Error("moved child identity changed")
	}

	// Out of range is rejected.
	if box.MoveChild(-1, 0) || box.MoveChild(0, 3) {
		t.Error("MoveChild must reject out-of-range indices")
	}
	// Same index is a valid no-op.
	if !box.MoveChild(1, 1) {
		t.Error("MoveChild(i,i) should be a valid no-op returning true")
	}
}
