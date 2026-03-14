package dirty

import (
	"testing"

	"github.com/gogpu/ui/geometry"
)

func TestNewTracker(t *testing.T) {
	tr := NewTracker()
	if tr == nil {
		t.Fatal("NewTracker returned nil")
	}
	if !tr.IsEmpty() {
		t.Error("new tracker should be empty")
	}
	if tr.RegionCount() != 0 {
		t.Errorf("new tracker region count = %d, want 0", tr.RegionCount())
	}
}

func TestNewTrackerWithOptions(t *testing.T) {
	tr := NewTrackerWithOptions(WithMergeGap(32), WithMaxRegions(8))
	if tr.mergeGap != 32 {
		t.Errorf("mergeGap = %v, want 32", tr.mergeGap)
	}
	if tr.maxCount != 8 {
		t.Errorf("maxCount = %d, want 8", tr.maxCount)
	}
}

func TestWithMergeGap_Negative(t *testing.T) {
	tr := NewTrackerWithOptions(WithMergeGap(-5))
	if tr.mergeGap != defaultMergeGap {
		t.Errorf("negative gap should be ignored, got %v", tr.mergeGap)
	}
}

func TestWithMaxRegions_Zero(t *testing.T) {
	tr := NewTrackerWithOptions(WithMaxRegions(0))
	if tr.maxCount != maxRegionsBeforeFullRepaint {
		t.Errorf("zero maxRegions should be ignored, got %d", tr.maxCount)
	}
}

func TestTracker_MarkDirty(t *testing.T) {
	tr := NewTracker()
	r := geometry.NewRect(10, 20, 100, 50)
	tr.MarkDirty(r)

	if tr.IsEmpty() {
		t.Error("tracker should not be empty after MarkDirty")
	}
	if tr.RegionCount() != 1 {
		t.Errorf("region count = %d, want 1", tr.RegionCount())
	}
	regions := tr.DirtyRegions()
	if len(regions) != 1 {
		t.Fatalf("DirtyRegions len = %d, want 1", len(regions))
	}
	if regions[0].Bounds != r {
		t.Errorf("region bounds = %v, want %v", regions[0].Bounds, r)
	}
}

func TestTracker_MarkDirty_EmptyRect(t *testing.T) {
	tr := NewTracker()
	tr.MarkDirty(geometry.Rect{}) // empty rect
	if !tr.IsEmpty() {
		t.Error("empty rect should be ignored")
	}
}

func TestTracker_MarkDirty_NegativeSizeRect(t *testing.T) {
	tr := NewTracker()
	// Min > Max means empty rect.
	tr.MarkDirty(geometry.Rect{
		Min: geometry.Pt(100, 100),
		Max: geometry.Pt(50, 50),
	})
	if !tr.IsEmpty() {
		t.Error("negative-size rect should be ignored")
	}
}

func TestTracker_Reset(t *testing.T) {
	tr := NewTracker()
	tr.MarkDirty(geometry.NewRect(0, 0, 100, 100))
	tr.MarkDirty(geometry.NewRect(200, 200, 50, 50))
	if tr.RegionCount() != 2 {
		t.Fatalf("expected 2 regions before reset, got %d", tr.RegionCount())
	}

	tr.Reset()
	if !tr.IsEmpty() {
		t.Error("tracker should be empty after Reset")
	}
	if tr.RegionCount() != 0 {
		t.Errorf("region count after reset = %d, want 0", tr.RegionCount())
	}
}

func TestTracker_FullRepaint(t *testing.T) {
	tr := NewTracker()
	tr.MarkDirty(geometry.NewRect(0, 0, 10, 10))
	tr.MarkDirty(geometry.NewRect(50, 50, 10, 10))

	viewport := geometry.NewRect(0, 0, 800, 600)
	tr.FullRepaint(viewport)

	if tr.RegionCount() != 1 {
		t.Errorf("after FullRepaint, region count = %d, want 1", tr.RegionCount())
	}
	regions := tr.DirtyRegions()
	if regions[0].Bounds != viewport {
		t.Errorf("full repaint region = %v, want %v", regions[0].Bounds, viewport)
	}
}

func TestTracker_FullRepaint_EmptyViewport(t *testing.T) {
	tr := NewTracker()
	tr.MarkDirty(geometry.NewRect(0, 0, 10, 10))
	tr.FullRepaint(geometry.Rect{})
	if !tr.IsEmpty() {
		t.Error("FullRepaint with empty viewport should result in empty tracker")
	}
}

func TestTracker_Intersects(t *testing.T) {
	tr := NewTracker()
	tr.MarkDirty(geometry.NewRect(10, 10, 100, 100))
	tr.MarkDirty(geometry.NewRect(300, 300, 50, 50))

	tests := []struct {
		name   string
		bounds geometry.Rect
		want   bool
	}{
		{"inside first", geometry.NewRect(20, 20, 10, 10), true},
		{"overlaps first", geometry.NewRect(0, 0, 50, 50), true},
		{"inside second", geometry.NewRect(310, 310, 10, 10), true},
		{"between regions", geometry.NewRect(200, 200, 10, 10), false},
		{"completely outside", geometry.NewRect(500, 500, 10, 10), false},
		{"touching first edge", geometry.NewRect(110, 10, 10, 10), false}, // touching but not overlapping per Intersects
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := tr.Intersects(tt.bounds)
			if got != tt.want {
				t.Errorf("Intersects(%v) = %v, want %v", tt.bounds, got, tt.want)
			}
		})
	}
}

func TestTracker_Intersects_Empty(t *testing.T) {
	tr := NewTracker()
	if tr.Intersects(geometry.NewRect(0, 0, 100, 100)) {
		t.Error("empty tracker should not intersect anything")
	}
}

func TestTracker_NeedsFullRepaint(t *testing.T) {
	tr := NewTrackerWithOptions(WithMaxRegions(3))
	tr.MarkDirty(geometry.NewRect(0, 0, 10, 10))
	tr.MarkDirty(geometry.NewRect(100, 0, 10, 10))
	tr.MarkDirty(geometry.NewRect(200, 0, 10, 10))

	if tr.NeedsFullRepaint() {
		t.Error("3 regions with max=3 should not need full repaint")
	}

	tr.MarkDirty(geometry.NewRect(300, 0, 10, 10))
	if !tr.NeedsFullRepaint() {
		t.Error("4 regions with max=3 should need full repaint")
	}
}

func TestTracker_Optimize_NoRegions(t *testing.T) {
	tr := NewTracker()
	tr.Optimize() // should not panic
	if !tr.IsEmpty() {
		t.Error("optimize on empty tracker should remain empty")
	}
}

func TestTracker_Optimize_SingleRegion(t *testing.T) {
	tr := NewTracker()
	r := geometry.NewRect(10, 10, 100, 100)
	tr.MarkDirty(r)
	tr.Optimize()

	if tr.RegionCount() != 1 {
		t.Errorf("single region after optimize: count = %d, want 1", tr.RegionCount())
	}
	if tr.DirtyRegions()[0].Bounds != r {
		t.Errorf("single region changed after optimize")
	}
}

func TestTracker_Optimize_OverlappingRegions(t *testing.T) {
	tr := NewTracker()
	tr.MarkDirty(geometry.NewRect(0, 0, 100, 100))
	tr.MarkDirty(geometry.NewRect(50, 50, 100, 100))
	tr.Optimize()

	if tr.RegionCount() != 1 {
		t.Errorf("overlapping regions after optimize: count = %d, want 1", tr.RegionCount())
	}
	expected := geometry.NewRect(0, 0, 150, 150)
	if tr.DirtyRegions()[0].Bounds != expected {
		t.Errorf("merged region = %v, want %v", tr.DirtyRegions()[0].Bounds, expected)
	}
}

func TestTracker_Optimize_NearbyRegions(t *testing.T) {
	tr := NewTrackerWithOptions(WithMergeGap(20))
	// Two regions 10px apart — within 20px gap.
	tr.MarkDirty(geometry.NewRect(0, 0, 50, 50))
	tr.MarkDirty(geometry.NewRect(60, 0, 50, 50))
	tr.Optimize()

	if tr.RegionCount() != 1 {
		t.Errorf("nearby regions after optimize: count = %d, want 1", tr.RegionCount())
	}
}

func TestTracker_Optimize_FarApartRegions(t *testing.T) {
	tr := NewTrackerWithOptions(WithMergeGap(10))
	// Two regions 100px apart — well beyond 10px gap.
	tr.MarkDirty(geometry.NewRect(0, 0, 50, 50))
	tr.MarkDirty(geometry.NewRect(200, 200, 50, 50))
	tr.Optimize()

	if tr.RegionCount() != 2 {
		t.Errorf("far apart regions after optimize: count = %d, want 2", tr.RegionCount())
	}
}

func TestTracker_Optimize_ChainMerge(t *testing.T) {
	// A, B, C where A merges with B, and the resulting AB merges with C.
	tr := NewTrackerWithOptions(WithMergeGap(15))
	tr.MarkDirty(geometry.NewRect(0, 0, 50, 50))
	tr.MarkDirty(geometry.NewRect(60, 0, 50, 50))  // 10px from A, within gap
	tr.MarkDirty(geometry.NewRect(120, 0, 50, 50)) // 10px from B, within gap
	tr.Optimize()

	if tr.RegionCount() != 1 {
		t.Errorf("chain merge after optimize: count = %d, want 1", tr.RegionCount())
	}
	// Union of all three: x=[0..170], y=[0..50]
	expected := geometry.NewRect(0, 0, 170, 50)
	if tr.DirtyRegions()[0].Bounds != expected {
		t.Errorf("chain merged region = %v, want %v", tr.DirtyRegions()[0].Bounds, expected)
	}
}

func TestTracker_Optimize_MultipleGroups(t *testing.T) {
	tr := NewTrackerWithOptions(WithMergeGap(10))
	// Group 1: two nearby regions at top-left.
	tr.MarkDirty(geometry.NewRect(0, 0, 30, 30))
	tr.MarkDirty(geometry.NewRect(35, 0, 30, 30))
	// Group 2: two nearby regions at bottom-right.
	tr.MarkDirty(geometry.NewRect(500, 500, 30, 30))
	tr.MarkDirty(geometry.NewRect(535, 500, 30, 30))
	tr.Optimize()

	if tr.RegionCount() != 2 {
		t.Errorf("two groups after optimize: count = %d, want 2", tr.RegionCount())
	}
}

func TestTracker_Optimize_ZeroGap(t *testing.T) {
	tr := NewTrackerWithOptions(WithMergeGap(0))
	// Adjacent regions (gap=0): touching but expand(0) means exact bounds.
	// Two rects that don't overlap and have 0 gap should NOT merge.
	tr.MarkDirty(geometry.NewRect(0, 0, 50, 50))
	tr.MarkDirty(geometry.NewRect(60, 0, 50, 50)) // 10px gap
	tr.Optimize()

	if tr.RegionCount() != 2 {
		t.Errorf("with zero gap, non-overlapping regions should stay separate: count = %d, want 2", tr.RegionCount())
	}
}

func TestTracker_Optimize_IdenticalRegions(t *testing.T) {
	tr := NewTracker()
	r := geometry.NewRect(10, 10, 50, 50)
	tr.MarkDirty(r)
	tr.MarkDirty(r)
	tr.MarkDirty(r)
	tr.Optimize()

	if tr.RegionCount() != 1 {
		t.Errorf("identical regions after optimize: count = %d, want 1", tr.RegionCount())
	}
}

func TestTracker_Optimize_ContainedRegion(t *testing.T) {
	tr := NewTracker()
	tr.MarkDirty(geometry.NewRect(0, 0, 200, 200))
	tr.MarkDirty(geometry.NewRect(50, 50, 50, 50)) // fully inside first
	tr.Optimize()

	if tr.RegionCount() != 1 {
		t.Errorf("contained region after optimize: count = %d, want 1", tr.RegionCount())
	}
	expected := geometry.NewRect(0, 0, 200, 200)
	if tr.DirtyRegions()[0].Bounds != expected {
		t.Errorf("result = %v, want %v", tr.DirtyRegions()[0].Bounds, expected)
	}
}

func TestTracker_Reset_ReusesMemory(t *testing.T) {
	tr := NewTracker()
	for i := 0; i < 10; i++ {
		tr.MarkDirty(geometry.NewRect(float32(i)*20, 0, 10, 10))
	}
	tr.Reset()

	// After reset, adding regions should reuse the existing backing array.
	tr.MarkDirty(geometry.NewRect(0, 0, 10, 10))
	if tr.RegionCount() != 1 {
		t.Errorf("after reset and re-add: count = %d, want 1", tr.RegionCount())
	}
}

func TestShouldMerge(t *testing.T) {
	tests := []struct {
		name string
		a, b geometry.Rect
		gap  float32
		want bool
	}{
		{
			name: "overlapping",
			a:    geometry.NewRect(0, 0, 100, 100),
			b:    geometry.NewRect(50, 50, 100, 100),
			gap:  0,
			want: true,
		},
		{
			name: "adjacent within gap",
			a:    geometry.NewRect(0, 0, 50, 50),
			b:    geometry.NewRect(55, 0, 50, 50),
			gap:  10,
			want: true,
		},
		{
			name: "far apart",
			a:    geometry.NewRect(0, 0, 50, 50),
			b:    geometry.NewRect(200, 200, 50, 50),
			gap:  10,
			want: false,
		},
		{
			name: "vertically adjacent within gap",
			a:    geometry.NewRect(0, 0, 100, 50),
			b:    geometry.NewRect(0, 55, 100, 50),
			gap:  10,
			want: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := shouldMerge(tt.a, tt.b, tt.gap)
			if got != tt.want {
				t.Errorf("shouldMerge(%v, %v, %v) = %v, want %v", tt.a, tt.b, tt.gap, got, tt.want)
			}
		})
	}
}

func TestTracker_Optimize_SortsByYThenX(t *testing.T) {
	tr := NewTrackerWithOptions(WithMergeGap(0))
	// Add in reverse order.
	tr.MarkDirty(geometry.NewRect(300, 300, 10, 10))
	tr.MarkDirty(geometry.NewRect(100, 100, 10, 10))
	tr.MarkDirty(geometry.NewRect(200, 100, 10, 10))
	tr.Optimize()

	// Regions should be sorted: (100,100), (200,100), (300,300).
	regions := tr.DirtyRegions()
	if len(regions) != 3 {
		t.Fatalf("expected 3 regions, got %d", len(regions))
	}
	if regions[0].Bounds.Min.Y > regions[1].Bounds.Min.Y {
		t.Error("regions not sorted by Y")
	}
	if regions[0].Bounds.Min.X > regions[1].Bounds.Min.X {
		t.Error("regions with same Y not sorted by X")
	}
}

func BenchmarkOptimize(b *testing.B) {
	benchCases := []struct {
		name  string
		count int
	}{
		{"10_regions", 10},
		{"50_regions", 50},
		{"100_regions", 100},
		{"500_regions", 500},
	}
	for _, bc := range benchCases {
		b.Run(bc.name, func(b *testing.B) {
			for i := 0; i < b.N; i++ {
				tr := NewTracker()
				for j := 0; j < bc.count; j++ {
					x := float32(j%20) * 40
					y := float32(j/20) * 40
					tr.MarkDirty(geometry.NewRect(x, y, 30, 30))
				}
				tr.Optimize()
			}
		})
	}
}

func BenchmarkIntersects(b *testing.B) {
	tr := NewTracker()
	for i := 0; i < 20; i++ {
		tr.MarkDirty(geometry.NewRect(float32(i)*50, 0, 30, 30))
	}
	tr.Optimize()
	bounds := geometry.NewRect(250, 0, 30, 30)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		tr.Intersects(bounds)
	}
}
