// Package dirty provides dirty region tracking for efficient partial repaints.
//
// The tracker collects rectangular areas of the widget tree that have changed
// and need repainting. It supports region merging to reduce the number of
// draw calls, and a full-repaint fallback when too many small regions accumulate.
//
// This is an internal package used by the rendering loop. Widgets mark themselves
// as needing redraw via [widget.WidgetBase.SetNeedsRedraw], and the [Collector]
// walks the tree to populate the tracker before each frame.
package dirty

import (
	"sort"

	"github.com/gogpu/ui/geometry"
)

// defaultMergeGap is the default pixel gap threshold for merging nearby regions.
// Two regions separated by less than this distance are merged into one to reduce
// draw calls at the cost of slightly more overdraw.
const defaultMergeGap float32 = 16

// maxRegionsBeforeFullRepaint is the maximum number of dirty regions before
// the tracker falls back to a single full-viewport repaint. When many small
// regions accumulate, iterating them individually becomes more expensive than
// repainting everything.
const maxRegionsBeforeFullRepaint = 16

// Region represents a rectangular area that needs repainting.
type Region struct {
	Bounds geometry.Rect
}

// Tracker collects and manages dirty regions for a frame.
//
// Usage:
//  1. Call Reset at frame start.
//  2. Walk the widget tree with a Collector, which calls MarkDirty for each
//     widget that needs redraw.
//  3. Call Optimize to merge overlapping/nearby regions.
//  4. Use DirtyRegions to get the list of areas to repaint, or IsEmpty to
//     skip the frame entirely.
type Tracker struct {
	regions  []Region
	mergeGap float32
	maxCount int
}

// NewTracker creates a new dirty region tracker with default settings.
func NewTracker() *Tracker {
	return &Tracker{
		mergeGap: defaultMergeGap,
		maxCount: maxRegionsBeforeFullRepaint,
	}
}

// TrackerOption configures a Tracker.
type TrackerOption func(*Tracker)

// WithMergeGap sets the pixel gap threshold for merging nearby regions.
// Two regions separated by less than gap pixels are merged into their union.
func WithMergeGap(gap float32) TrackerOption {
	return func(t *Tracker) {
		if gap >= 0 {
			t.mergeGap = gap
		}
	}
}

// WithMaxRegions sets the maximum number of dirty regions before falling
// back to a full-viewport repaint.
func WithMaxRegions(n int) TrackerOption {
	return func(t *Tracker) {
		if n > 0 {
			t.maxCount = n
		}
	}
}

// NewTrackerWithOptions creates a new dirty region tracker with custom settings.
func NewTrackerWithOptions(opts ...TrackerOption) *Tracker {
	t := NewTracker()
	for _, opt := range opts {
		opt(t)
	}
	return t
}

// MarkDirty adds a rectangular area as dirty (needs repaint).
// Empty rectangles are ignored.
func (t *Tracker) MarkDirty(bounds geometry.Rect) {
	if bounds.IsEmpty() {
		return
	}
	t.regions = append(t.regions, Region{Bounds: bounds})
}

// Optimize merges overlapping and nearby dirty regions to reduce draw calls.
//
// The algorithm:
//  1. Sort regions by Y then X for spatial locality.
//  2. Repeatedly scan for pairs that overlap or have a gap smaller than the
//     merge threshold, replacing them with their union.
//  3. Repeat until no more merges are possible.
//
// This trades a small amount of overdraw for fewer clip/draw cycles.
func (t *Tracker) Optimize() {
	if len(t.regions) <= 1 {
		return
	}

	// Sort by Y then X for spatial locality.
	sort.Slice(t.regions, func(i, j int) bool {
		if t.regions[i].Bounds.Min.Y != t.regions[j].Bounds.Min.Y {
			return t.regions[i].Bounds.Min.Y < t.regions[j].Bounds.Min.Y
		}
		return t.regions[i].Bounds.Min.X < t.regions[j].Bounds.Min.X
	})

	// Iteratively merge until stable.
	merged := true
	for merged {
		merged = false
		for i := 0; i < len(t.regions); i++ {
			for j := i + 1; j < len(t.regions); j++ {
				if !shouldMerge(t.regions[i].Bounds, t.regions[j].Bounds, t.mergeGap) {
					continue
				}
				t.regions[i].Bounds = t.regions[i].Bounds.Union(t.regions[j].Bounds)
				// Remove j by swapping with last element.
				last := len(t.regions) - 1
				t.regions[j] = t.regions[last]
				t.regions[last] = Region{} // clear for GC
				t.regions = t.regions[:last]
				j-- // re-check this index
				merged = true
			}
		}
	}
}

// DirtyRegions returns the optimized list of regions to repaint.
// The returned slice must not be modified by the caller.
func (t *Tracker) DirtyRegions() []Region {
	return t.regions
}

// Reset clears all dirty regions. Call at the start of each frame.
func (t *Tracker) Reset() {
	// Reuse backing array to reduce allocations across frames.
	t.regions = t.regions[:0]
}

// IsEmpty returns true if no regions are dirty, meaning the frame can be skipped.
func (t *Tracker) IsEmpty() bool {
	return len(t.regions) == 0
}

// FullRepaint marks the entire viewport as dirty.
// This replaces any existing regions with a single viewport-sized region.
func (t *Tracker) FullRepaint(viewport geometry.Rect) {
	t.regions = t.regions[:0]
	if !viewport.IsEmpty() {
		t.regions = append(t.regions, Region{Bounds: viewport})
	}
}

// Intersects checks if the given bounds intersect any dirty region.
// This is used by widgets during the draw pass to skip rendering
// when their bounds are entirely outside all dirty regions.
func (t *Tracker) Intersects(bounds geometry.Rect) bool {
	for i := range t.regions {
		if t.regions[i].Bounds.Intersects(bounds) {
			return true
		}
	}
	return false
}

// RegionCount returns the current number of dirty regions.
func (t *Tracker) RegionCount() int {
	return len(t.regions)
}

// NeedsFullRepaint returns true if the number of dirty regions exceeds the
// maximum threshold, indicating a full repaint would be more efficient.
func (t *Tracker) NeedsFullRepaint() bool {
	return len(t.regions) > t.maxCount
}

// shouldMerge returns true if two rectangles overlap or are within gap pixels
// of each other. The gap is applied uniformly: if expanding both rects by
// gap/2 causes them to overlap, they should merge.
func shouldMerge(a, b geometry.Rect, gap float32) bool {
	// Expand a by gap and check intersection with b.
	expanded := a.Expand(gap)
	return expanded.Intersects(b)
}
