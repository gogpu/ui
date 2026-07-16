// Package listview lifecycle_test.go tests widget lifecycle correctness
// in the ListView widget cache. These are regression tests for fixes #173
// (decorator GPU scene leak on scroll) and #174 (row content never mounted).
//
// Each test uses a lifecycleTracker widget that records Mount/Unmount calls,
// verifying that the ListView properly mounts new items and unmounts evicted
// ones during scrolling, clear, and rebuild operations.
package listview_test

import (
	"testing"

	"github.com/gogpu/ui/core/listview"
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
	return c.Constrain(geometry.Sz(c.MaxWidth, 36))
}

func (lt *lifecycleTracker) Draw(_ widget.Context, _ widget.Canvas) {}

func (lt *lifecycleTracker) Event(_ widget.Context, _ event.Event) bool { return false }

func (lt *lifecycleTracker) Children() []widget.Widget { return nil }

// Compile-time check that lifecycleTracker implements Lifecycle.
var _ widget.Lifecycle = (*lifecycleTracker)(nil)

// newTrackedListView creates a ListView with a scrollY signal and a tracker
// for lifecycle monitoring. Each BuildItem call appends a lifecycleTracker
// to the provided slice.
func newTrackedListView(itemCount int, trackers *[]*lifecycleTracker) *listview.Widget {
	scrollY := state.NewSignal[float32](0)
	lv := listview.New(
		listview.ItemCount(itemCount),
		listview.FixedItemHeight(36),
		listview.ScrollYSignal(scrollY),
		listview.BuildItem(func(_ listview.ItemContext) widget.Widget {
			lt := &lifecycleTracker{}
			*trackers = append(*trackers, lt)
			return lt
		}),
	)
	return lv
}

// layoutAndDraw performs a full layout+draw cycle on a ListView at the
// given viewport size. This triggers the widget cache to build visible items.
func layoutAndDraw(t *testing.T, lv *listview.Widget, ctx widget.Context, size geometry.Size) {
	t.Helper()
	constraints := geometry.Constraints{
		MinWidth: size.Width, MaxWidth: size.Width,
		MinHeight: size.Height, MaxHeight: size.Height,
	}
	lv.Layout(ctx, constraints)
	lv.SetBounds(geometry.NewRect(0, 0, size.Width, size.Height))
	lv.Draw(ctx, &mockCanvas{})
}

func TestLifecycle_NewItemsMounted(t *testing.T) {
	// Verify that items created during the first layout+draw cycle get
	// MountTree called, so signal bindings activate (#174).
	var trackers []*lifecycleTracker
	lv := newTrackedListView(20, &trackers)

	ctx := widget.NewContext()
	// Viewport fits 5 items (5 * 36 = 180).
	layoutAndDraw(t, lv, ctx, geometry.Sz(300, 180))

	if len(trackers) == 0 {
		t.Fatal("no trackers created — BuildItem was not called")
	}

	// All created trackers should have been mounted.
	for i, lt := range trackers {
		if !lt.mounted {
			t.Errorf("tracker[%d]: expected mounted=true, got false", i)
		}
		if lt.mountCount != 1 {
			t.Errorf("tracker[%d]: mountCount = %d, want 1", i, lt.mountCount)
		}
	}
}

func TestLifecycle_ScrollEvictsOldItems(t *testing.T) {
	// Simulate a scroll that changes the visible range, verifying that
	// evicted items are unmounted and new items are mounted (#173).
	var trackers []*lifecycleTracker
	lv := newTrackedListView(100, &trackers)

	ctx := widget.NewContext()
	// Viewport fits 5 items (5 * 36 = 180).
	layoutAndDraw(t, lv, ctx, geometry.Sz(300, 180))

	initialCount := len(trackers)
	if initialCount == 0 {
		t.Fatal("no trackers created on first draw")
	}

	firstBatch := make([]*lifecycleTracker, initialCount)
	copy(firstBatch, trackers)

	// Scroll far enough down that the visible range no longer overlaps
	// with the initial range. Items 0-4 visible initially.
	// ScrollToIndex(50) moves view to item 50+ (no overlap with 0-4).
	lv.ScrollToIndex(50)
	layoutAndDraw(t, lv, ctx, geometry.Sz(300, 180))

	// The first batch of trackers should now be unmounted.
	unmountedCount := 0
	for _, lt := range firstBatch {
		if lt.unmountCount > 0 {
			unmountedCount++
		}
	}

	if unmountedCount == 0 {
		t.Error("expected items from the first batch to be unmounted after scroll to non-overlapping range")
	}

	// New trackers should be created and mounted.
	if len(trackers) <= initialCount {
		t.Error("expected new trackers to be created after scroll")
	}
	for i := initialCount; i < len(trackers); i++ {
		if !trackers[i].mounted {
			t.Errorf("new tracker[%d]: expected mounted=true after scroll", i)
		}
	}
}

func TestLifecycle_ClearUnmountsAll(t *testing.T) {
	// Verify that reducing item count to fewer than current visible range
	// unmounts evicted widgets via the fullRebuild path (#173).
	// When item count drops, InvalidateData marks cache invalid, and the
	// next Draw triggers fullRebuild which unmounts old widgets.
	trackers2 := make([]*lifecycleTracker, 0)
	countSig := state.NewSignal(10)
	scrollY := state.NewSignal[float32](0)
	lv2 := listview.New(
		listview.ItemCountSignal(countSig),
		listview.FixedItemHeight(36),
		listview.ScrollYSignal(scrollY),
		listview.BuildItem(func(_ listview.ItemContext) widget.Widget {
			lt := &lifecycleTracker{}
			trackers2 = append(trackers2, lt)
			return lt
		}),
	)

	ctx := widget.NewContext()
	layoutAndDraw(t, lv2, ctx, geometry.Sz(300, 360))
	if len(trackers2) == 0 {
		t.Fatal("no trackers created for signal-based list")
	}

	batch := make([]*lifecycleTracker, len(trackers2))
	copy(batch, trackers2)

	// Set count to 2 (was 10) — shrinks visible range, triggers fullRebuild.
	countSig.Set(2)
	lv2.InvalidateData()
	layoutAndDraw(t, lv2, ctx, geometry.Sz(300, 360))

	// The old batch items should be unmounted (fullRebuild unmounts all
	// old widgets before creating new ones).
	unmountedCount := 0
	for _, lt := range batch {
		if lt.unmountCount > 0 {
			unmountedCount++
		}
	}
	if unmountedCount == 0 {
		t.Error("expected at least some trackers to be unmounted after count reduction")
	}
}

func TestLifecycle_RescrollCreatesFreshWidgets(t *testing.T) {
	// After scrolling away and back, fresh widgets should be created and
	// mounted (not reusing unmounted ones).
	var trackers []*lifecycleTracker
	lv := newTrackedListView(100, &trackers)

	ctx := widget.NewContext()
	layoutAndDraw(t, lv, ctx, geometry.Sz(300, 180))

	initialCount := len(trackers)
	if initialCount == 0 {
		t.Fatal("no trackers created")
	}

	// Scroll far down (non-overlapping with initial range).
	lv.ScrollToIndex(50)
	layoutAndDraw(t, lv, ctx, geometry.Sz(300, 180))
	afterScrollCount := len(trackers)

	// Scroll back to top.
	lv.ScrollToIndex(0)
	layoutAndDraw(t, lv, ctx, geometry.Sz(300, 180))

	// New trackers should have been created for the return trip.
	if len(trackers) <= afterScrollCount {
		t.Errorf("expected new trackers after scrolling back; had %d after first scroll, now %d",
			afterScrollCount, len(trackers))
	}

	// The newly created trackers (after scrolling back) should be mounted.
	for i := afterScrollCount; i < len(trackers); i++ {
		if !trackers[i].mounted {
			t.Errorf("tracker[%d] (created on scroll-back): expected mounted=true", i)
		}
	}
}

func TestLifecycle_FullRebuildUnmountsOld(t *testing.T) {
	// When cache is invalidated (data change), fullRebuild should
	// unmount old widgets before creating replacements.
	var trackers []*lifecycleTracker
	lv := newTrackedListView(10, &trackers)

	ctx := widget.NewContext()
	layoutAndDraw(t, lv, ctx, geometry.Sz(300, 360)) // fits all 10 items

	oldCount := len(trackers)
	if oldCount == 0 {
		t.Fatal("no trackers created")
	}

	oldTrackers := make([]*lifecycleTracker, oldCount)
	copy(oldTrackers, trackers)

	// Force a full rebuild by invalidating and re-drawing.
	lv.InvalidateData()
	layoutAndDraw(t, lv, ctx, geometry.Sz(300, 360))

	// Old trackers should have been unmounted.
	for i, lt := range oldTrackers {
		if lt.unmountCount == 0 {
			t.Errorf("old tracker[%d]: expected unmountCount > 0 after rebuild, got 0", i)
		}
	}
}
