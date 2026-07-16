// Package gridview lifecycle_test.go tests widget lifecycle correctness
// in the GridView cell cache. These are regression tests for fix #181
// (GridView cell content never mounted — signals dead).
//
// Each test uses a lifecycleTracker widget that records Mount/Unmount calls,
// verifying that the GridView properly mounts new cells and unmounts evicted
// ones during cache updates, clear, and rebuild operations.
package gridview_test

import (
	"testing"

	"github.com/gogpu/ui/core/gridview"
	"github.com/gogpu/ui/event"
	"github.com/gogpu/ui/geometry"
	"github.com/gogpu/ui/state"
	"github.com/gogpu/ui/widget"
)

// lifecycleTracker is a minimal widget that records Mount/Unmount calls
// for verifying lifecycle correctness.
type lifecycleTracker struct {
	widget.WidgetBase
	mounted      bool
	mountCount   int
	unmountCount int
}

func (lt *lifecycleTracker) Mount(_ widget.Context) {
	lt.mounted = true
	lt.mountCount++
}

func (lt *lifecycleTracker) Unmount() {
	lt.mounted = false
	lt.unmountCount++
}

func (lt *lifecycleTracker) Layout(_ widget.Context, c geometry.Constraints) geometry.Size {
	return c.Constrain(geometry.Sz(100, 100))
}

func (lt *lifecycleTracker) Draw(_ widget.Context, _ widget.Canvas) {}

func (lt *lifecycleTracker) Event(_ widget.Context, _ event.Event) bool { return false }

func (lt *lifecycleTracker) Children() []widget.Widget { return nil }

// Compile-time check that lifecycleTracker implements Lifecycle.
var _ widget.Lifecycle = (*lifecycleTracker)(nil)

// gridLayoutAndDraw performs layout+draw on a GridView to trigger cell cache.
func gridLayoutAndDraw(t *testing.T, gv *gridview.Widget, ctx widget.Context, size geometry.Size) {
	t.Helper()
	constraints := geometry.Constraints{
		MinWidth: size.Width, MaxWidth: size.Width,
		MinHeight: size.Height, MaxHeight: size.Height,
	}
	gv.Layout(ctx, constraints)
	gv.SetBounds(geometry.NewRect(0, 0, size.Width, size.Height))
	gv.Draw(ctx, &mockCanvas{})
}

func TestLifecycle_GridCellsMounted(t *testing.T) {
	// Verify that cells created during layout+draw get MountTree called,
	// so signal bindings activate (#181).
	var trackers []*lifecycleTracker
	gv := gridview.New(
		gridview.ItemCount(20),
		gridview.ItemSize(100, 100),
		gridview.Columns(3),
		gridview.BuildCell(func(_ int, _ gridview.CellContext) widget.Widget {
			lt := &lifecycleTracker{}
			trackers = append(trackers, lt)
			return lt
		}),
	)

	ctx := widget.NewContext()
	// 3 columns * 100px = 300px wide, 300px tall = 3 rows = 9 cells visible.
	gridLayoutAndDraw(t, gv, ctx, geometry.Sz(300, 300))

	if len(trackers) == 0 {
		t.Fatal("no trackers created — BuildCell was not called")
	}

	for i, lt := range trackers {
		if !lt.mounted {
			t.Errorf("cell tracker[%d]: expected mounted=true, got false", i)
		}
		if lt.mountCount != 1 {
			t.Errorf("cell tracker[%d]: mountCount = %d, want 1", i, lt.mountCount)
		}
	}
}

func TestLifecycle_GridCellUpdateUnmountsOld(t *testing.T) {
	// When the grid scrolls and the cache updates, old cells should be
	// unmounted and new cells should be mounted (#181).
	var trackers []*lifecycleTracker
	scrollY := state.NewSignal[float32](0)
	gv := gridview.New(
		gridview.ItemCount(100),
		gridview.ItemSize(100, 100),
		gridview.Columns(3),
		gridview.ScrollYSignal(scrollY),
		gridview.BuildCell(func(_ int, _ gridview.CellContext) widget.Widget {
			lt := &lifecycleTracker{}
			trackers = append(trackers, lt)
			return lt
		}),
	)

	ctx := widget.NewContext()
	gridLayoutAndDraw(t, gv, ctx, geometry.Sz(300, 300))

	initialCount := len(trackers)
	if initialCount == 0 {
		t.Fatal("no trackers created on first draw")
	}

	firstBatch := make([]*lifecycleTracker, initialCount)
	copy(firstBatch, trackers)

	// Scroll far down to a non-overlapping range.
	gv.ScrollToIndex(50)
	gridLayoutAndDraw(t, gv, ctx, geometry.Sz(300, 300))

	// First batch should be unmounted.
	unmountedCount := 0
	for _, lt := range firstBatch {
		if lt.unmountCount > 0 {
			unmountedCount++
		}
	}

	if unmountedCount == 0 {
		t.Error("expected cells from the first batch to be unmounted after scroll")
	}

	// New cells should be created and mounted.
	if len(trackers) <= initialCount {
		t.Error("expected new trackers to be created after scroll")
	}
	for i := initialCount; i < len(trackers); i++ {
		if !trackers[i].mounted {
			t.Errorf("new cell tracker[%d]: expected mounted=true after scroll", i)
		}
	}
}

func TestLifecycle_GridClearUnmountsAll(t *testing.T) {
	// Verify that reducing item count unmounts old cells via fullRebuild.
	// When item count shrinks and data is invalidated, the cache rebuilds
	// with fewer cells, unmounting the old ones (#181).
	var trackers []*lifecycleTracker
	countSig := state.NewSignal(12)
	gv := gridview.New(
		gridview.ItemCountSignal(countSig),
		gridview.ItemSize(100, 100),
		gridview.Columns(3),
		gridview.BuildCell(func(_ int, _ gridview.CellContext) widget.Widget {
			lt := &lifecycleTracker{}
			trackers = append(trackers, lt)
			return lt
		}),
	)

	ctx := widget.NewContext()
	gridLayoutAndDraw(t, gv, ctx, geometry.Sz(300, 400))

	if len(trackers) == 0 {
		t.Fatal("no trackers created")
	}

	batch := make([]*lifecycleTracker, len(trackers))
	copy(batch, trackers)

	// Set count to 2 (was 12) — triggers fullRebuild with fewer cells.
	countSig.Set(2)
	gv.InvalidateData()
	gridLayoutAndDraw(t, gv, ctx, geometry.Sz(300, 400))

	// Old batch cells should be unmounted during fullRebuild.
	unmountedCount := 0
	for _, lt := range batch {
		if lt.unmountCount > 0 {
			unmountedCount++
		}
	}
	if unmountedCount == 0 {
		t.Error("expected at least some cell trackers to be unmounted after count reduction")
	}
}

func TestLifecycle_GridInvalidateRebuild(t *testing.T) {
	// InvalidateData should unmount old cells and mount fresh ones.
	var trackers []*lifecycleTracker
	gv := gridview.New(
		gridview.ItemCount(6),
		gridview.ItemSize(100, 100),
		gridview.Columns(3),
		gridview.BuildCell(func(_ int, _ gridview.CellContext) widget.Widget {
			lt := &lifecycleTracker{}
			trackers = append(trackers, lt)
			return lt
		}),
	)

	ctx := widget.NewContext()
	gridLayoutAndDraw(t, gv, ctx, geometry.Sz(300, 200))

	oldCount := len(trackers)
	if oldCount == 0 {
		t.Fatal("no trackers created")
	}

	oldTrackers := make([]*lifecycleTracker, oldCount)
	copy(oldTrackers, trackers)

	// Invalidate and redraw — all cells rebuilt.
	gv.InvalidateData()
	gridLayoutAndDraw(t, gv, ctx, geometry.Sz(300, 200))

	for i, lt := range oldTrackers {
		if lt.unmountCount == 0 {
			t.Errorf("old cell tracker[%d]: expected unmountCount > 0 after rebuild, got 0", i)
		}
	}
}
