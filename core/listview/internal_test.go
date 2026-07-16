package listview

import (
	"image"
	"testing"

	"github.com/gogpu/gg/scene"
	"github.com/gogpu/ui/cdk"
	"github.com/gogpu/ui/event"
	"github.com/gogpu/ui/geometry"
	"github.com/gogpu/ui/widget"
)

// --- heightManager tests ---

func TestHeightManagerFixed_TotalHeight(t *testing.T) {
	tests := []struct {
		name   string
		count  int
		height float32
		want   float32
	}{
		{"zero items", 0, 48, 0},
		{"one item", 1, 48, 48},
		{"ten items", 10, 48, 480},
		{"large count", 100000, 56, 5600000},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			hm := heightManager{
				mode:        heightFixed,
				count:       tt.count,
				fixedHeight: tt.height,
			}
			got := hm.totalHeight()
			if got != tt.want {
				t.Errorf("totalHeight() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestHeightManagerFixed_OffsetAt(t *testing.T) {
	hm := heightManager{
		mode:        heightFixed,
		count:       10,
		fixedHeight: 48,
	}

	tests := []struct {
		index int
		want  float32
	}{
		{0, 0},
		{1, 48},
		{5, 240},
		{10, 480},
		{-1, 0},
		{11, 480}, // clamped to count
	}
	for _, tt := range tests {
		got := hm.offsetAt(tt.index)
		if got != tt.want {
			t.Errorf("offsetAt(%d) = %v, want %v", tt.index, got, tt.want)
		}
	}
}

func TestHeightManagerFixed_HeightAt(t *testing.T) {
	hm := heightManager{
		mode:        heightFixed,
		count:       10,
		fixedHeight: 48,
	}

	if got := hm.heightAt(0); got != 48 {
		t.Errorf("heightAt(0) = %v, want 48", got)
	}
	if got := hm.heightAt(9); got != 48 {
		t.Errorf("heightAt(9) = %v, want 48", got)
	}
	if got := hm.heightAt(-1); got != 0 {
		t.Errorf("heightAt(-1) = %v, want 0", got)
	}
	if got := hm.heightAt(10); got != 0 {
		t.Errorf("heightAt(10) = %v, want 0", got)
	}
}

func TestHeightManagerFixed_FindIndexAtOffset(t *testing.T) {
	hm := heightManager{
		mode:        heightFixed,
		count:       10,
		fixedHeight: 48,
	}

	tests := []struct {
		offset float32
		want   int
	}{
		{0, 0},
		{24, 0},
		{48, 1},
		{96, 2},
		{479, 9},
		{480, 9},  // clamped
		{1000, 9}, // clamped
		{-10, 0},  // clamped
	}
	for _, tt := range tests {
		got := hm.findIndexAtOffset(tt.offset)
		if got != tt.want {
			t.Errorf("findIndexAtOffset(%v) = %v, want %v", tt.offset, got, tt.want)
		}
	}
}

func TestHeightManagerFixed_VisibleRange(t *testing.T) {
	hm := heightManager{
		mode:        heightFixed,
		count:       100,
		fixedHeight: 48,
	}

	tests := []struct {
		name      string
		scrollY   float32
		viewportH float32
		overscan  int
		wantStart int
		wantEnd   int
	}{
		{"top", 0, 200, 0, 0, 5},
		{"with overscan", 0, 200, 3, 0, 8},
		{"scrolled", 200, 200, 0, 4, 9},
		{"scrolled with overscan", 200, 200, 2, 2, 11},
		{"at end", 4600, 200, 0, 95, 100},
		{"empty viewport", 0, 0, 0, 0, 1},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			start, end := hm.visibleRange(tt.scrollY, tt.viewportH, tt.overscan)
			if start != tt.wantStart || end != tt.wantEnd {
				t.Errorf("visibleRange(%v, %v, %d) = (%d, %d), want (%d, %d)",
					tt.scrollY, tt.viewportH, tt.overscan, start, end, tt.wantStart, tt.wantEnd)
			}
		})
	}
}

func TestHeightManagerFixed_ZeroHeight(t *testing.T) {
	hm := heightManager{
		mode:        heightFixed,
		count:       10,
		fixedHeight: 0,
	}

	if got := hm.totalHeight(); got != 0 {
		t.Errorf("totalHeight() = %v, want 0", got)
	}
	if got := hm.findIndexAtOffset(100); got != 0 {
		t.Errorf("findIndexAtOffset(100) = %v, want 0", got)
	}
}

func TestHeightManagerCallback_TotalHeight(t *testing.T) {
	heights := []float32{20, 30, 40, 50, 60}
	hm := heightManager{
		mode:     heightCallback,
		count:    len(heights),
		heightFn: func(i int) float32 { return heights[i] },
	}
	hm.buildCallbackOffsets()

	want := float32(200) // 20+30+40+50+60
	if got := hm.totalHeight(); got != want {
		t.Errorf("totalHeight() = %v, want %v", got, want)
	}
}

func TestHeightManagerCallback_OffsetAt(t *testing.T) {
	heights := []float32{20, 30, 40, 50, 60}
	hm := heightManager{
		mode:     heightCallback,
		count:    len(heights),
		heightFn: func(i int) float32 { return heights[i] },
	}
	hm.buildCallbackOffsets()

	expected := []float32{0, 20, 50, 90, 140, 200}
	for i, want := range expected {
		got := hm.offsetAt(i)
		if got != want {
			t.Errorf("offsetAt(%d) = %v, want %v", i, got, want)
		}
	}
}

func TestHeightManagerCallback_FindIndexAtOffset(t *testing.T) {
	heights := []float32{20, 30, 40, 50, 60}
	hm := heightManager{
		mode:     heightCallback,
		count:    len(heights),
		heightFn: func(i int) float32 { return heights[i] },
	}
	hm.buildCallbackOffsets()

	tests := []struct {
		offset float32
		want   int
	}{
		{0, 0},
		{10, 0},
		{20, 1},
		{49, 1},
		{50, 2},
		{89, 2},
		{90, 3},
		{139, 3},
		{140, 4},
		{199, 4},
		{200, 4}, // clamped
	}
	for _, tt := range tests {
		got := hm.findIndexAtOffset(tt.offset)
		if got != tt.want {
			t.Errorf("findIndexAtOffset(%v) = %v, want %v", tt.offset, got, tt.want)
		}
	}
}

func TestHeightManagerCallback_VisibleRange(t *testing.T) {
	heights := []float32{20, 30, 40, 50, 60}
	hm := heightManager{
		mode:     heightCallback,
		count:    len(heights),
		heightFn: func(i int) float32 { return heights[i] },
	}
	hm.buildCallbackOffsets()

	start, end := hm.visibleRange(0, 80, 0)
	// Items 0 (0-20), 1 (20-50), 2 (50-90) are visible at scroll=0, viewport=80
	if start != 0 || end != 3 {
		t.Errorf("visibleRange(0, 80, 0) = (%d, %d), want (0, 3)", start, end)
	}
}

func TestHeightManagerLazy_InitialEstimate(t *testing.T) {
	hm := heightManager{
		mode:            heightLazy,
		count:           100,
		estimatedHeight: 50,
	}
	hm.initLazy()

	want := float32(5000) // 100 * 50
	if got := hm.totalHeight(); got != want {
		t.Errorf("totalHeight() = %v, want %v", got, want)
	}
}

func TestHeightManagerLazy_SetMeasured(t *testing.T) {
	hm := heightManager{
		mode:            heightLazy,
		count:           10,
		estimatedHeight: 50,
	}
	hm.initLazy()

	// Measure first item as 60px.
	hm.setMeasured(0, 60)

	if hm.numMeasured != 1 {
		t.Errorf("numMeasured = %d, want 1", hm.numMeasured)
	}
	if hm.measuredSum != 60 {
		t.Errorf("measuredSum = %v, want 60", hm.measuredSum)
	}
	// Current estimate should now be the measured average.
	if got := hm.currentEstimate(); got != 60 {
		t.Errorf("currentEstimate() = %v, want 60", got)
	}
}

func TestHeightManagerLazy_SetMeasuredUpdate(t *testing.T) {
	hm := heightManager{
		mode:            heightLazy,
		count:           10,
		estimatedHeight: 50,
	}
	hm.initLazy()

	hm.setMeasured(0, 60)
	hm.setMeasured(0, 70) // update existing

	if hm.numMeasured != 1 {
		t.Errorf("numMeasured = %d, want 1", hm.numMeasured)
	}
	if hm.measuredSum != 70 {
		t.Errorf("measuredSum = %v, want 70", hm.measuredSum)
	}
}

func TestHeightManagerLazy_SetMeasuredIgnored(t *testing.T) {
	hm := heightManager{
		mode:            heightLazy,
		count:           5,
		estimatedHeight: 50,
	}
	hm.initLazy()

	// Out of range: should not panic or change state.
	hm.setMeasured(-1, 60)
	hm.setMeasured(5, 60)
	hm.setMeasured(10, 60)

	if hm.numMeasured != 0 {
		t.Errorf("numMeasured = %d, want 0", hm.numMeasured)
	}
}

func TestHeightManagerLazy_SetMeasuredNotLazy(t *testing.T) {
	hm := heightManager{
		mode:        heightFixed,
		count:       10,
		fixedHeight: 48,
	}

	// setMeasured should be no-op for non-lazy mode.
	hm.setMeasured(0, 60)
	// No panic = success.
}

func TestHeightManagerLazy_HeightAt(t *testing.T) {
	hm := heightManager{
		mode:            heightLazy,
		count:           10,
		estimatedHeight: 50,
	}
	hm.initLazy()

	// Before measurement: returns estimate.
	if got := hm.heightAt(0); got != 50 {
		t.Errorf("heightAt(0) before measure = %v, want 50", got)
	}

	hm.setMeasured(0, 60)

	// After measurement: returns measured value.
	if got := hm.heightAt(0); got != 60 {
		t.Errorf("heightAt(0) after measure = %v, want 60", got)
	}
}

func TestHeightManager_UpdateCount(t *testing.T) {
	t.Run("fixed mode", func(t *testing.T) {
		hm := heightManager{
			mode:        heightFixed,
			count:       10,
			fixedHeight: 48,
		}

		hm.updateCount(20)
		if hm.count != 20 {
			t.Errorf("count = %d, want 20", hm.count)
		}
		if got := hm.totalHeight(); got != 960 {
			t.Errorf("totalHeight() = %v, want 960", got)
		}
	})

	t.Run("callback mode", func(t *testing.T) {
		heights := []float32{20, 30, 40, 50, 60, 70, 80}
		hm := heightManager{
			mode:     heightCallback,
			count:    5,
			heightFn: func(i int) float32 { return heights[i] },
		}
		hm.buildCallbackOffsets()

		hm.updateCount(7)
		want := float32(20 + 30 + 40 + 50 + 60 + 70 + 80)
		if got := hm.totalHeight(); got != want {
			t.Errorf("totalHeight() = %v, want %v", got, want)
		}
	})

	t.Run("lazy mode grow", func(t *testing.T) {
		hm := heightManager{
			mode:            heightLazy,
			count:           5,
			estimatedHeight: 50,
		}
		hm.initLazy()

		hm.setMeasured(0, 60)
		hm.updateCount(10)

		if len(hm.measured) != 10 {
			t.Errorf("len(measured) = %d, want 10", len(hm.measured))
		}
		if hm.numMeasured != 1 {
			t.Errorf("numMeasured = %d, want 1", hm.numMeasured)
		}
	})

	t.Run("lazy mode shrink", func(t *testing.T) {
		hm := heightManager{
			mode:            heightLazy,
			count:           10,
			estimatedHeight: 50,
		}
		hm.initLazy()

		hm.setMeasured(0, 60)
		hm.setMeasured(8, 70) // will be removed on shrink
		hm.updateCount(5)

		if len(hm.measured) != 5 {
			t.Errorf("len(measured) = %d, want 5", len(hm.measured))
		}
		if hm.numMeasured != 1 {
			t.Errorf("numMeasured = %d, want 1 (only index 0 survived)", hm.numMeasured)
		}
	})

	t.Run("same count no-op", func(t *testing.T) {
		hm := heightManager{mode: heightFixed, count: 10, fixedHeight: 48}
		hm.updateCount(10) // no-op
		if hm.count != 10 {
			t.Errorf("count = %d, want 10", hm.count)
		}
	})
}

func TestHeightManager_EmptyList(t *testing.T) {
	hm := heightManager{
		mode:        heightFixed,
		count:       0,
		fixedHeight: 48,
	}

	if got := hm.totalHeight(); got != 0 {
		t.Errorf("totalHeight() = %v, want 0", got)
	}

	start, end := hm.visibleRange(0, 400, 3)
	if start != 0 || end != 0 {
		t.Errorf("visibleRange = (%d, %d), want (0, 0)", start, end)
	}
}

func TestNewHeightManager(t *testing.T) {
	t.Run("fixed height", func(t *testing.T) {
		cfg := config{fixedItemHeight: 56, itemCount: 100}
		hm := newHeightManager(&cfg)
		if hm.mode != heightFixed {
			t.Errorf("mode = %v, want heightFixed", hm.mode)
		}
	})

	t.Run("callback height", func(t *testing.T) {
		cfg := config{
			itemHeightFn: func(i int) float32 { return 48 },
			itemCount:    10,
		}
		hm := newHeightManager(&cfg)
		if hm.mode != heightCallback {
			t.Errorf("mode = %v, want heightCallback", hm.mode)
		}
	})

	t.Run("lazy height", func(t *testing.T) {
		cfg := config{itemCount: 10, estimatedItemHeight: 50}
		hm := newHeightManager(&cfg)
		if hm.mode != heightLazy {
			t.Errorf("mode = %v, want heightLazy", hm.mode)
		}
	})
}

// --- widgetCache tests ---

// newTestCache creates a widgetCache with a dummy list back-reference.
func newTestCache() widgetCache {
	lv := &Widget{
		hoveredIndex: noHoveredIndex,
		painter:      DefaultPainter{},
	}
	lv.SetVisible(true)
	return widgetCache{list: lv}
}

func TestWidgetCache_Update(t *testing.T) {
	wc := newTestCache()
	builder := cdk.FuncContent[ItemContext]{Fn: func(ctx ItemContext) widget.Widget {
		return nil // simple builder for testing
	}}

	wc.update(0, 5, builder, -1, nil)

	if !wc.valid {
		t.Error("cache should be valid after update")
	}
	if wc.startIndex != 0 {
		t.Errorf("startIndex = %d, want 0", wc.startIndex)
	}
	if wc.endIndex != 5 {
		t.Errorf("endIndex = %d, want 5", wc.endIndex)
	}
	if len(wc.widgets) != 5 {
		t.Errorf("len(widgets) = %d, want 5", len(wc.widgets))
	}
}

func TestWidgetCache_Reuse(t *testing.T) {
	wc := newTestCache()
	callCount := 0
	builder := cdk.FuncContent[ItemContext]{Fn: func(ctx ItemContext) widget.Widget {
		callCount++
		return nil
	}}

	wc.update(0, 5, builder, -1, nil)
	if callCount != 5 {
		t.Errorf("callCount = %d, want 5", callCount)
	}

	// Second call with same range should be no-op.
	wc.update(0, 5, builder, -1, nil)
	if callCount != 5 {
		t.Errorf("callCount after reuse = %d, want 5", callCount)
	}
}

func TestWidgetCache_InvalidateForces_Rebuild(t *testing.T) {
	wc := newTestCache()
	callCount := 0
	builder := cdk.FuncContent[ItemContext]{Fn: func(ctx ItemContext) widget.Widget {
		callCount++
		return nil
	}}

	wc.update(0, 5, builder, -1, nil)
	wc.invalidate()
	wc.update(0, 5, builder, -1, nil)

	if callCount != 10 {
		t.Errorf("callCount = %d, want 10 (5+5 after invalidate)", callCount)
	}
}

func TestWidgetCache_Clear(t *testing.T) {
	wc := newTestCache()
	builder := cdk.FuncContent[ItemContext]{Fn: func(ctx ItemContext) widget.Widget {
		return nil
	}}

	wc.update(0, 5, builder, -1, nil)
	wc.clear()

	if wc.valid {
		t.Error("cache should be invalid after clear")
	}
	if len(wc.widgets) != 0 {
		t.Errorf("len(widgets) = %d, want 0", len(wc.widgets))
	}
}

func TestWidgetCache_EmptyRange(t *testing.T) {
	wc := newTestCache()
	builder := cdk.FuncContent[ItemContext]{Fn: func(ctx ItemContext) widget.Widget {
		return nil
	}}

	wc.update(0, 0, builder, -1, nil)

	if wc.valid {
		t.Error("cache should not be valid for empty range")
	}
}

func TestWidgetCache_WidgetAt(t *testing.T) {
	wc := newTestCache()

	// Empty cache.
	if got := wc.widgetAt(0); got != nil {
		t.Error("widgetAt on empty cache should return nil")
	}
	if got := wc.widgetAt(-1); got != nil {
		t.Error("widgetAt(-1) should return nil")
	}
}

func TestWidgetCache_NilBuilder(t *testing.T) {
	wc := newTestCache()
	wc.update(0, 3, nil, -1, nil)

	for i := 0; i < 3; i++ {
		if got := wc.widgetAt(i); got != nil {
			t.Errorf("widgetAt(%d) with nil builder should return nil", i)
		}
	}
}

func TestWidgetCache_ItemContextPropagation(t *testing.T) {
	wc := newTestCache()
	var received []ItemContext
	builder := cdk.FuncContent[ItemContext]{Fn: func(ctx ItemContext) widget.Widget {
		received = append(received, ctx)
		return nil
	}}

	wc.update(5, 8, builder, 6, nil)

	if len(received) != 3 {
		t.Fatalf("received %d contexts, want 3", len(received))
	}

	// Index 5: not selected.
	if received[0].Index != 5 || received[0].Selected {
		t.Errorf("ctx[0] = %+v, want Index=5, Selected=false", received[0])
	}
	// Index 6: selected.
	if received[1].Index != 6 || !received[1].Selected {
		t.Errorf("ctx[1] = %+v, want Index=6, Selected=true", received[1])
	}
	// Index 7: not selected (hover is read at Draw time, not baked into context).
	if received[2].Index != 7 || received[2].Selected {
		t.Errorf("ctx[2] = %+v, want Index=7, Selected=false", received[2])
	}
}

// --- heightManager callback/lazy edge cases ---

func TestHeightManagerCallback_HeightAt(t *testing.T) {
	heights := []float32{20, 30, 40}
	hm := heightManager{
		mode:     heightCallback,
		count:    len(heights),
		heightFn: func(i int) float32 { return heights[i] },
	}
	hm.buildCallbackOffsets()

	if got := hm.heightAt(0); got != 20 {
		t.Errorf("heightAt(0) = %v, want 20", got)
	}
	if got := hm.heightAt(2); got != 40 {
		t.Errorf("heightAt(2) = %v, want 40", got)
	}
	// Out of range.
	if got := hm.heightAt(-1); got != 0 {
		t.Errorf("heightAt(-1) = %v, want 0", got)
	}
	if got := hm.heightAt(3); got != 0 {
		t.Errorf("heightAt(3) = %v, want 0", got)
	}
}

func TestHeightManagerCallback_HeightAt_NilFn(t *testing.T) {
	hm := heightManager{
		mode:     heightCallback,
		count:    5,
		heightFn: nil,
	}

	if got := hm.heightAt(0); got != 0 {
		t.Errorf("heightAt(0) with nil fn = %v, want 0", got)
	}
}

func TestHeightManagerCallback_TotalHeight_EmptyOffsets(t *testing.T) {
	hm := heightManager{
		mode:    heightCallback,
		count:   5,
		offsets: nil, // no offsets built
	}

	if got := hm.totalHeight(); got != 0 {
		t.Errorf("totalHeight() with nil offsets = %v, want 0", got)
	}
}

func TestHeightManagerCallback_OffsetAt_EmptyOffsets(t *testing.T) {
	hm := heightManager{
		mode:    heightCallback,
		count:   5,
		offsets: nil,
	}

	if got := hm.offsetAt(3); got != 0 {
		t.Errorf("offsetAt(3) with nil offsets = %v, want 0", got)
	}
}

func TestHeightManagerLazy_TotalHeight_NoOffsets(t *testing.T) {
	hm := heightManager{
		mode:            heightLazy,
		count:           10,
		estimatedHeight: 50,
		offsetsDirty:    true,
	}
	// No measured slice — ensureOffsets will build from scratch.
	hm.measured = make([]float32, 10)
	for i := range hm.measured {
		hm.measured[i] = -1
	}

	got := hm.totalHeight()
	if got != 500 {
		t.Errorf("totalHeight() = %v, want 500", got)
	}
}

func TestHeightManagerLazy_OffsetAt(t *testing.T) {
	hm := heightManager{
		mode:            heightLazy,
		count:           5,
		estimatedHeight: 50,
	}
	hm.initLazy()

	// All estimated: offset at 3 = 3*50 = 150.
	if got := hm.offsetAt(3); got != 150 {
		t.Errorf("offsetAt(3) = %v, want 150", got)
	}
}

func TestHeightManagerLazy_FindIndexAtOffset(t *testing.T) {
	hm := heightManager{
		mode:            heightLazy,
		count:           10,
		estimatedHeight: 50,
	}
	hm.initLazy()

	if got := hm.findIndexAtOffset(125); got != 2 {
		t.Errorf("findIndexAtOffset(125) = %v, want 2", got)
	}
}

func TestHeightManagerLazy_SetMeasured_SameValue(t *testing.T) {
	hm := heightManager{
		mode:            heightLazy,
		count:           5,
		estimatedHeight: 50,
	}
	hm.initLazy()

	hm.setMeasured(0, 60)
	hm.setMeasured(0, 60) // same value => no-op

	if hm.numMeasured != 1 {
		t.Errorf("numMeasured = %d, want 1", hm.numMeasured)
	}
}

func TestHeightManagerLazy_EnsureOffsets_NotDirty(t *testing.T) {
	hm := heightManager{
		mode:            heightLazy,
		count:           3,
		estimatedHeight: 50,
	}
	hm.initLazy()
	hm.ensureOffsets() // builds offsets, sets dirty=false.

	// Call again — should be no-op (offsets already correct).
	hm.ensureOffsets()
	if len(hm.offsets) != 4 {
		t.Errorf("len(offsets) = %d, want 4", len(hm.offsets))
	}
}

func TestHeightManager_BinarySearch_EmptyOffsets(t *testing.T) {
	hm := heightManager{
		mode:    heightCallback,
		count:   0,
		offsets: nil,
	}

	if got := hm.binarySearchOffset(100); got != 0 {
		t.Errorf("binarySearchOffset(100) with nil offsets = %v, want 0", got)
	}
}

func TestNewHeightManager_LazyDefaultEstimate(t *testing.T) {
	// No estimated height set => should use defaultEstimatedHeight.
	cfg := config{itemCount: 10, estimatedItemHeight: 0}
	hm := newHeightManager(&cfg)

	if hm.mode != heightLazy {
		t.Errorf("mode = %v, want heightLazy", hm.mode)
	}
	if hm.estimatedHeight != defaultEstimatedHeight {
		t.Errorf("estimatedHeight = %v, want %v", hm.estimatedHeight, defaultEstimatedHeight)
	}
}

// --- virtualContent tests ---

func TestVirtualContent_Children(t *testing.T) {
	vc := &virtualContent{}
	if got := vc.Children(); got != nil {
		t.Errorf("Children() = %v, want nil", got)
	}
}

func TestVirtualContent_Layout_NilList(t *testing.T) {
	vc := &virtualContent{}
	got := vc.Layout(nil, geometry.Constraints{MinWidth: 100, MaxWidth: 300, MinHeight: 100, MaxHeight: 500})
	if got != (geometry.Size{}) {
		t.Errorf("Layout with nil list = %v, want zero", got)
	}
}

func TestVirtualContent_Draw_NilList(t *testing.T) {
	vc := &virtualContent{}
	// Should not panic.
	vc.Draw(nil, &mockCanvas{})
}

func TestVirtualContent_Event_NilList(t *testing.T) {
	vc := &virtualContent{}
	if got := vc.Event(nil, &event.MouseEvent{}); got {
		t.Error("Event with nil list should return false")
	}
}

func TestVirtualContent_Layout_InfiniteWidth(t *testing.T) {
	lv := New(
		ItemCount(10),
		FixedItemHeight(48),
	)

	vc := &virtualContent{list: lv}
	got := vc.Layout(nil, geometry.Constraints{
		MinWidth:  100,
		MaxWidth:  geometry.Infinity,
		MinHeight: 0,
		MaxHeight: geometry.Infinity,
	})

	// Should use MinWidth when MaxWidth is infinity.
	if got.Width != 100 {
		t.Errorf("Width = %v, want 100", got.Width)
	}
}

// --- mock types for internal tests ---

type mockWidget struct {
	widget.WidgetBase
}

func (w *mockWidget) Layout(_ widget.Context, c geometry.Constraints) geometry.Size {
	return c.Constrain(geometry.Sz(100, 48))
}

func (w *mockWidget) Draw(_ widget.Context, _ widget.Canvas) {}

func (w *mockWidget) Event(_ widget.Context, _ event.Event) bool { return false }

type mockCanvas struct{}

func (m *mockCanvas) Clear(_ widget.Color)                                                  {}
func (m *mockCanvas) DrawRect(_ geometry.Rect, _ widget.Color)                              {}
func (m *mockCanvas) FillRectDirect(_ geometry.Rect, _ widget.Color)                        {}
func (m *mockCanvas) StrokeRect(_ geometry.Rect, _ widget.Color, _ float32)                 {}
func (m *mockCanvas) DrawRoundRect(_ geometry.Rect, _ widget.Color, _ float32)              {}
func (m *mockCanvas) StrokeRoundRect(_ geometry.Rect, _ widget.Color, _ float32, _ float32) {}
func (m *mockCanvas) DrawCircle(_ geometry.Point, _ float32, _ widget.Color)                {}
func (m *mockCanvas) StrokeCircle(_ geometry.Point, _ float32, _ widget.Color, _ float32)   {}
func (m *mockCanvas) StrokeArc(_ geometry.Point, _ float32, _, _ float64, _ widget.Color, _ float32) {
}
func (m *mockCanvas) DrawLine(_, _ geometry.Point, _ widget.Color, _ float32) {}
func (m *mockCanvas) DrawText(_ string, _ geometry.Rect, _ float32, _ widget.Color, _ bool, _ widget.TextAlign) {
}

func (m *mockCanvas) MeasureText(text string, fontSize float32, _ bool) float32 {
	return float32(len([]rune(text))) * fontSize * 0.5
}
func (m *mockCanvas) DrawImage(_ image.Image, _ geometry.Point)    {}
func (m *mockCanvas) PushClip(_ geometry.Rect)                     {}
func (m *mockCanvas) PushClipRoundRect(_ geometry.Rect, _ float32) {}
func (m *mockCanvas) PopClip()                                     {}
func (m *mockCanvas) PushTransform(_ geometry.Point)               {}
func (m *mockCanvas) PopTransform()                                {}
func (m *mockCanvas) TransformOffset() geometry.Point              { return geometry.Point{} }
func (m *mockCanvas) ScreenOriginBase() geometry.Point             { return geometry.Point{} }
func (m *mockCanvas) ClipBounds() geometry.Rect                    { return geometry.NewRect(0, 0, 10000, 10000) }
func (m *mockCanvas) ReplayScene(_ *scene.Scene)                   {}

// --- Decorator + RepaintBoundary integration tests ---

func TestWidgetCache_BoundariesCreated(t *testing.T) {
	wc := newTestCache()
	builder := cdk.FuncContent[ItemContext]{Fn: func(ctx ItemContext) widget.Widget {
		w := &mockWidget{}
		w.SetVisible(true)
		return w
	}}

	wc.update(0, 3, builder, -1, nil)

	if len(wc.widgets) != 3 {
		t.Fatalf("len(widgets) = %d, want 3", len(wc.widgets))
	}
	for i := 0; i < 3; i++ {
		w := wc.widgetAt(i)
		if w == nil {
			t.Errorf("widget[%d] is nil, want non-nil", i)
			continue
		}
		// widgetAt returns the decorator, which IS the boundary.
		dec, ok := w.(*itemDecorator)
		if !ok {
			t.Errorf("widget[%d] should be *itemDecorator, got %T", i, w)
			continue
		}
		if !dec.IsRepaintBoundary() {
			t.Errorf("decorator[%d].IsRepaintBoundary() = false, want true", i)
		}
	}
}

func TestWidgetCache_WidgetAtWithBoundary(t *testing.T) {
	wc := newTestCache()
	builder := cdk.FuncContent[ItemContext]{Fn: func(ctx ItemContext) widget.Widget {
		w := &mockWidget{}
		w.SetVisible(true)
		return w
	}}

	wc.update(0, 3, builder, -1, nil)

	// Valid offsets — widget should exist and be a decorator (RepaintBoundary).
	for i := 0; i < 3; i++ {
		w := wc.widgetAt(i)
		if w == nil {
			t.Errorf("widgetAt(%d) = nil, want non-nil", i)
			continue
		}
		dec, ok := w.(*itemDecorator)
		if !ok {
			t.Errorf("widget[%d] should be *itemDecorator, got %T", i, w)
			continue
		}
		if !dec.IsRepaintBoundary() {
			t.Errorf("decorator[%d].IsRepaintBoundary() = false, want true", i)
		}
	}

	// Out of range.
	if w := wc.widgetAt(-1); w != nil {
		t.Error("widgetAt(-1) should return nil")
	}
	if w := wc.widgetAt(3); w != nil {
		t.Error("widgetAt(3) should return nil")
	}

	// Empty cache.
	empty := newTestCache()
	if w := empty.widgetAt(0); w != nil {
		t.Error("widgetAt on empty cache should return nil")
	}
}

func TestWidgetCache_BoundaryNilForNilWidget(t *testing.T) {
	wc := newTestCache()
	builder := cdk.FuncContent[ItemContext]{Fn: func(ctx ItemContext) widget.Widget {
		return nil // builder returns nil widget
	}}

	wc.update(0, 3, builder, -1, nil)

	for i := 0; i < 3; i++ {
		if w := wc.widgetAt(i); w != nil {
			t.Errorf("widgetAt(%d) should be nil for nil widget", i)
		}
	}
}

func TestWidgetCache_BoundaryNilBuilder(t *testing.T) {
	wc := newTestCache()
	wc.update(0, 3, nil, -1, nil)

	for i := 0; i < 3; i++ {
		if w := wc.widgetAt(i); w != nil {
			t.Errorf("widgetAt(%d) should be nil with nil builder", i)
		}
	}
}

func TestWidgetCache_BoundaryWrapsCorrectChild(t *testing.T) {
	wc := newTestCache()
	widgets := make([]*mockWidget, 3)
	builder := cdk.FuncContent[ItemContext]{Fn: func(ctx ItemContext) widget.Widget {
		w := &mockWidget{}
		w.SetVisible(true)
		widgets[ctx.Index] = w
		return w
	}}

	wc.update(0, 3, builder, -1, nil)

	// widgetAt returns the decorator. childAt returns the user widget.
	// Verify childAt returns the same widget the builder created.
	for i := 0; i < 3; i++ {
		decorator := wc.widgetAt(i)
		if decorator == nil {
			t.Fatalf("decorator[%d] is nil", i)
		}
		dec, ok := decorator.(*itemDecorator)
		if !ok {
			t.Fatalf("widget[%d] should be *itemDecorator, got %T", i, decorator)
		}
		if !dec.IsRepaintBoundary() {
			t.Errorf("decorator[%d].IsRepaintBoundary() = false, want true", i)
		}
		child := wc.childAt(i)
		if child != widgets[i] {
			t.Errorf("childAt(%d) != original widget[%d]", i, i)
		}
	}
}

func TestWidgetCache_ClearUnmountsBoundaries(t *testing.T) {
	wc := newTestCache()
	builder := cdk.FuncContent[ItemContext]{Fn: func(ctx ItemContext) widget.Widget {
		w := &mockWidget{}
		w.SetVisible(true)
		return w
	}}

	wc.update(0, 3, builder, -1, nil)

	// Verify decorators with boundary property exist before clear.
	for i := 0; i < 3; i++ {
		w := wc.widgetAt(i)
		if w == nil {
			t.Fatalf("widget[%d] nil before clear", i)
		}
		dec, ok := w.(*itemDecorator)
		if !ok {
			t.Fatalf("widget[%d] should be *itemDecorator", i)
		}
		if !dec.IsRepaintBoundary() {
			t.Fatalf("decorator[%d].IsRepaintBoundary() = false before clear", i)
		}
	}

	wc.clear()

	if len(wc.widgets) != 0 {
		t.Errorf("len(widgets) = %d, want 0 after clear", len(wc.widgets))
	}
}

func TestWidgetCache_InvalidateRebuildsBoundaries(t *testing.T) {
	wc := newTestCache()
	callCount := 0
	builder := cdk.FuncContent[ItemContext]{Fn: func(ctx ItemContext) widget.Widget {
		callCount++
		w := &mockWidget{}
		w.SetVisible(true)
		return w
	}}

	wc.update(0, 3, builder, -1, nil)
	w1 := wc.widgetAt(0)
	if w1 == nil {
		t.Fatal("widget[0] nil before invalidate")
	}
	dec1, ok := w1.(*itemDecorator)
	if !ok {
		t.Fatal("widget[0] should be *itemDecorator")
	}
	if !dec1.IsRepaintBoundary() {
		t.Fatal("decorator[0].IsRepaintBoundary() = false before invalidate")
	}

	wc.invalidate()
	wc.update(0, 3, builder, -1, nil)

	w2 := wc.widgetAt(0)
	if w2 == nil {
		t.Fatal("widget[0] nil after rebuild")
	}
	dec2, ok := w2.(*itemDecorator)
	if !ok {
		t.Fatal("widget[0] should be *itemDecorator after rebuild")
	}
	if !dec2.IsRepaintBoundary() {
		t.Fatal("decorator[0].IsRepaintBoundary() = false after rebuild")
	}

	// After invalidate + rebuild, the decorator should be a new instance.
	if w1 == w2 {
		t.Error("decorator should be a new instance after invalidate + rebuild")
	}

	if callCount != 6 {
		t.Errorf("callCount = %d, want 6 (3+3)", callCount)
	}
}

func TestWidgetCache_RangeShiftCreatesBoundaries(t *testing.T) {
	wc := newTestCache()
	builder := cdk.FuncContent[ItemContext]{Fn: func(ctx ItemContext) widget.Widget {
		w := &mockWidget{}
		w.SetVisible(true)
		return w
	}}

	// Initial range [0, 3).
	wc.update(0, 3, builder, -1, nil)

	// Shift range to [5, 8).
	wc.update(5, 8, builder, -1, nil)

	if len(wc.widgets) != 3 {
		t.Fatalf("len(widgets) = %d, want 3", len(wc.widgets))
	}
	for i := 0; i < 3; i++ {
		w := wc.widgetAt(i)
		if w == nil {
			t.Errorf("widget[%d] nil after range shift", i)
			continue
		}
		dec, ok := w.(*itemDecorator)
		if !ok {
			t.Errorf("widget[%d] should be *itemDecorator after range shift", i)
			continue
		}
		if !dec.IsRepaintBoundary() {
			t.Errorf("decorator[%d].IsRepaintBoundary() = false after range shift", i)
		}
	}
}

// --- markItemDirty tests (ADR-007 Task 1f) ---

func TestMarkItemDirty_InRange(t *testing.T) {
	wc := newTestCache()
	builder := cdk.FuncContent[ItemContext]{Fn: func(ctx ItemContext) widget.Widget {
		w := &mockWidget{}
		w.SetVisible(true)
		w.SetEnabled(true)
		return w
	}}

	wc.update(0, 5, builder, -1, nil)

	// Create a ListView with this cache.
	lv := &Widget{
		hoveredIndex: noHoveredIndex,
	}
	lv.SetVisible(true)
	lv.SetEnabled(true)
	lv.cache = wc

	// Clear any initial redraw state.
	lv.ClearRedraw()

	// Mark item 2 dirty.
	lv.markItemDirty(2)

	// The decorator at index 2 should be marked dirty.
	decorator := lv.cache.widgetAt(2)
	if decorator == nil {
		t.Fatal("decorator at index 2 should not be nil")
	}
	if base, ok := decorator.(interface{ NeedsRedraw() bool }); ok {
		if !base.NeedsRedraw() {
			t.Error("decorator at index 2 should need redraw")
		}
	}

	// The decorator IS the boundary — its scene should be dirty.
	dec := decorator.(*itemDecorator)
	if !dec.IsSceneDirty() {
		t.Error("decorator boundary scene should be dirty after markItemDirty")
	}

	// ListView should NOT be marked dirty — hover/selection painting moved
	// inside the decorator. Only the decorator's boundary is dirty.
	if lv.NeedsRedraw() {
		t.Error("ListView should NOT be marked for redraw; " +
			"hover/selection painting is inside the decorator (per-item boundary)")
	}
}

func TestMarkItemDirty_OutOfRange(t *testing.T) {
	wc := newTestCache()
	builder := cdk.FuncContent[ItemContext]{Fn: func(ctx ItemContext) widget.Widget {
		w := &mockWidget{}
		w.SetVisible(true)
		return w
	}}

	wc.update(5, 10, builder, -1, nil)

	lv := &Widget{
		hoveredIndex: noHoveredIndex,
	}
	lv.SetVisible(true)
	lv.cache = wc

	// Index 2 is before the cached range [5, 10) — should not panic.
	lv.markItemDirty(2)

	// Index 15 is after the cached range — should not panic.
	lv.markItemDirty(15)
}

func TestHoverChange_NoInvalidate(t *testing.T) {
	wc := newTestCache()
	builder := cdk.FuncContent[ItemContext]{Fn: func(ctx ItemContext) widget.Widget {
		w := &mockWidget{}
		w.SetVisible(true)
		w.SetEnabled(true)
		return w
	}}

	wc.update(0, 5, builder, -1, nil)

	// Clear initial dirty state (simulates first draw).
	for i := range 5 {
		if dec, ok := wc.widgetAt(i).(*itemDecorator); ok {
			dec.ClearRedraw()
		}
	}

	lv := &Widget{
		hoveredIndex: noHoveredIndex,
	}
	lv.SetVisible(true)
	lv.SetEnabled(true)
	lv.cache = wc

	// Record initial cache valid state.
	initialValid := lv.cache.valid

	// Simulate hover change via the same logic as handleContentMouseMove.
	old := lv.hoveredIndex
	lv.hoveredIndex = 2
	if old >= 0 {
		lv.markItemDirty(old)
	}
	if lv.hoveredIndex >= 0 {
		lv.markItemDirty(lv.hoveredIndex)
	}

	// The widget cache should NOT be invalidated (no cache.invalidate() call).
	if lv.cache.valid != initialValid {
		t.Error("hover change should NOT invalidate the widget cache")
	}

	// Only the affected decorator should be dirty, not all.
	for i := 0; i < 5; i++ {
		item := lv.cache.widgetAt(i)
		if item == nil {
			continue
		}
		base, ok := item.(interface{ NeedsRedraw() bool })
		if !ok {
			continue
		}
		if i == 2 {
			if !base.NeedsRedraw() {
				t.Errorf("decorator %d should be dirty (new hovered)", i)
			}
		} else {
			// Other items should NOT be dirty (hover didn't touch them).
			// Note: old hovered was -1, so only new hovered (2) is dirty.
			if base.NeedsRedraw() {
				t.Errorf("decorator %d should NOT be dirty (only hover target changes)", i)
			}
		}
	}
}

// --- SelectionMode tests ---

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
		if got := tt.mode.String(); got != tt.want {
			t.Errorf("SelectionMode(%d).String() = %q, want %q", tt.mode, got, tt.want)
		}
	}
}

// --- Granular invalidation regression tests (2026-05-07) ---

// TestSelectionChangeDirtiesOnlyTwoItems verifies that changing the selected
// index only rebuilds the old and new selected items, not ALL items in cache.
// Before the fix, setSelectedIndex called cache.invalidate() which rebuilt
// every visible item, causing unnecessary widget allocation and layout.
// Regression: setSelectedIndex called cache.invalidate() -> ALL items recreated (2026-05-07)
func TestSelectionChangeDirtiesOnlyTwoItems(t *testing.T) {
	wc := newTestCache()
	callCount := 0
	builder := cdk.FuncContent[ItemContext]{Fn: func(ctx ItemContext) widget.Widget {
		callCount++
		w := &mockWidget{}
		w.SetVisible(true)
		return w
	}}

	// Initial build: 10 items, item 3 selected.
	wc.update(0, 10, builder, 3, nil)
	initialCount := callCount

	if initialCount != 10 {
		t.Fatalf("initial build: callCount = %d, want 10", initialCount)
	}

	// Save references to all widgets (decorators).
	originalWidgets := make([]widget.Widget, 10)
	for i := 0; i < 10; i++ {
		originalWidgets[i] = wc.widgetAt(i)
	}

	// Change selection from 3 to 5 — should only rebuild items 3 and 5.
	wc.rebuildAffected(0, builder, 5)

	rebuiltCount := callCount - initialCount
	if rebuiltCount != 2 {
		t.Errorf("rebuild count = %d, want 2 (only old+new selection); "+
			"rebuildAffected should not recreate all items", rebuiltCount)
	}

	// Verify only items 3 and 5 were replaced (new decorators).
	for i := 0; i < 10; i++ {
		current := wc.widgetAt(i)
		if i == 3 || i == 5 {
			if current == originalWidgets[i] {
				t.Errorf("item %d should have been rebuilt (selection changed)", i)
			}
		} else {
			if current != originalWidgets[i] {
				t.Errorf("item %d should NOT have been rebuilt (selection did not affect it)", i)
			}
		}
	}
}

// TestListViewNotDirtyOnItemClick verifies that markItemDirty does NOT mark
// the ListView itself as needing redraw. With the decorator pattern, hover and
// selection backgrounds are painted inside the decorator (per-item boundary),
// so only the decorator is dirty, not the parent ListView.
func TestListViewNotDirtyOnItemClick(t *testing.T) {
	wc := newTestCache()
	builder := cdk.FuncContent[ItemContext]{Fn: func(ctx ItemContext) widget.Widget {
		w := &mockWidget{}
		w.SetVisible(true)
		w.SetEnabled(true)
		return w
	}}

	wc.update(0, 10, builder, -1, nil)

	lv := &Widget{
		hoveredIndex: noHoveredIndex,
	}
	lv.SetVisible(true)
	lv.SetEnabled(true)
	lv.cache = wc

	// Clear any initial redraw state.
	lv.ClearRedraw()

	// Mark a single item dirty (simulates click/hover on item 3).
	lv.markItemDirty(3)

	// The ListView itself should NOT be marked dirty — the decorator
	// owns the boundary, hover/selection painting is per-item.
	if lv.NeedsRedraw() {
		t.Error("markItemDirty should NOT mark the ListView for redraw; " +
			"hover/selection painting is inside the decorator boundary")
	}

	// But the decorator at index 3 should be marked.
	decorator := lv.cache.widgetAt(3)
	if decorator == nil {
		t.Fatal("decorator at index 3 should not be nil")
	}
	if base, ok := decorator.(interface{ NeedsRedraw() bool }); ok {
		if !base.NeedsRedraw() {
			t.Error("decorator at index 3 should need redraw")
		}
	}
}

// TestVirtualContentExposesChildrenForDirtyCollector verifies that
// virtualContent.Children() returns the cached decorators.
// dirty.Collector needs children to collect per-item dirty regions.
// Regression: virtualContent.Children() returned nil -> Collector missed individual items (2026-05-07)
func TestVirtualContentExposesChildrenForDirtyCollector(t *testing.T) {
	lv := New(
		ItemCount(5),
		FixedItemHeight(48),
	)

	wc := newTestCache()
	builder := cdk.FuncContent[ItemContext]{Fn: func(ctx ItemContext) widget.Widget {
		w := &mockWidget{}
		w.SetVisible(true)
		return w
	}}
	wc.update(0, 5, builder, -1, nil)
	lv.cache = wc

	vc := &virtualContent{list: lv}
	children := vc.Children()

	if children == nil {
		t.Fatal("virtualContent.Children() must not return nil when cache has items; " +
			"dirty.Collector needs children to collect per-item dirty regions")
	}

	if len(children) != 5 {
		t.Errorf("len(Children()) = %d, want 5 (one per visible item)", len(children))
	}
}

// --- ADR-024 ListView + RepaintBoundary Regression Tests ---
//
// These verify that ListView items wrapped in decorator correctly
// propagate dirty state and interact with parent boundary/ScrollView clip.

// TestListView_ItemBoundaryDirtyPropagation verifies that invalidating a
// decorator's scene only affects that item, not the whole ListView.
func TestListView_ItemBoundaryDirtyPropagation(t *testing.T) {
	builder := cdk.FuncContent[ItemContext]{Fn: func(ctx ItemContext) widget.Widget {
		w := &mockWidget{}
		w.SetVisible(true)
		return w
	}}

	wc := newTestCache()
	wc.update(0, 5, builder, -1, nil)

	for i := 0; i < 5; i++ {
		w := wc.widgetAt(i)
		if w == nil {
			t.Fatalf("widget[%d] is nil", i)
		}
		w.(*itemDecorator).ClearSceneDirty()
	}

	wc.widgetAt(2).(*itemDecorator).InvalidateScene()

	for i := 0; i < 5; i++ {
		dec := wc.widgetAt(i).(*itemDecorator)
		if i == 2 {
			if !dec.IsSceneDirty() {
				t.Errorf("decorator[%d] scene should be dirty (was explicitly invalidated)", i)
			}
		} else {
			if dec.IsSceneDirty() {
				t.Errorf("decorator[%d] scene should be clean (only item 2 was invalidated)", i)
			}
		}
	}
}

// TestListView_ScrollChangesVisibleRange verifies that scrolling changes
// the visible item range returned by visibleRange with overscan.
func TestListView_ScrollChangesVisibleRange(t *testing.T) {
	cfg := &config{
		itemCount:    100,
		itemHeightFn: func(_ int) float32 { return 36 },
		overscan:     3,
	}
	hm := newHeightManager(cfg)

	start, end := hm.visibleRange(0, 200, 3)
	if start != 0 {
		t.Errorf("start = %d at scroll=0, want 0", start)
	}
	if end < 5 {
		t.Errorf("end = %d at scroll=0, want >= 5 (visible + overscan)", end)
	}

	start2, end2 := hm.visibleRange(1000, 200, 3)
	if start2 <= 0 {
		t.Errorf("start = %d at scroll=1000, want > 0", start2)
	}
	if end2 <= end {
		t.Errorf("end = %d at scroll=1000, should be > %d (previous end)", end2, end)
	}
}

// TestListView_MarkItemDirtyStopsAtItemBoundary verifies that markItemDirty
// uses SetNeedsRedraw on the decorator (which IS the boundary).
// SetNeedsRedraw on a boundary widget invalidates its OWN scene and does
// NOT propagate to parent. This is the Flutter markNeedsPaint pattern.
func TestListView_MarkItemDirtyStopsAtItemBoundary(t *testing.T) {
	builder := cdk.FuncContent[ItemContext]{Fn: func(ctx ItemContext) widget.Widget {
		w := &mockWidget{}
		w.SetVisible(true)
		return w
	}}

	lv := &Widget{}
	lv.SetVisible(true)
	lv.SetEnabled(true)
	lv.cache.list = lv
	lv.cache.update(0, 5, builder, -1, nil)

	// Create a parent boundary to verify propagation STOPS at decorator.
	parent := &boundaryTracker{}
	parent.SetVisible(true)
	parent.SetRepaintBoundary(true)

	// Wire parent chain on each decorator.
	for i := 0; i < 5; i++ {
		w := lv.cache.widgetAt(i)
		if w == nil {
			t.Fatalf("widget[%d] nil", i)
		}
		w.(*itemDecorator).SetParent(parent)
		w.(*itemDecorator).ClearRedraw()
	}
	parent.ClearSceneDirty()
	parent.sceneDirtied = false

	// markItemDirty on item 2.
	lv.markItemDirty(2)

	// The decorator should be marked as needing redraw.
	dec := lv.cache.widgetAt(2).(*itemDecorator)
	if !dec.NeedsRedraw() {
		t.Error("decorator[2].NeedsRedraw() = false after markItemDirty")
	}

	// Decorator IS a boundary -> SetNeedsRedraw calls InvalidateScene on SELF,
	// does NOT propagate to parent. Parent boundary stays clean.
	if !dec.IsSceneDirty() {
		t.Error("decorator boundary should be scene-dirty (self-invalidated)")
	}
	if parent.sceneDirtied {
		t.Error("parent boundary should NOT be dirty; " +
			"decorator IS a boundary, propagation must stop at decorator level " +
			"(Flutter markNeedsPaint pattern)")
	}
}

// boundaryTracker is a test widget that tracks InvalidateScene calls.
type boundaryTracker struct {
	widget.WidgetBase
	sceneDirtied bool
}

func (w *boundaryTracker) InvalidateScene() {
	w.WidgetBase.InvalidateScene()
	w.sceneDirtied = true
}
func (w *boundaryTracker) Layout(_ widget.Context, c geometry.Constraints) geometry.Size {
	return c.Constrain(geometry.Sz(400, 300))
}
func (w *boundaryTracker) Draw(_ widget.Context, _ widget.Canvas)     {}
func (w *boundaryTracker) Event(_ widget.Context, _ event.Event) bool { return false }
func (w *boundaryTracker) Children() []widget.Widget                  { return nil }

// TestListView_MarkItemDirtyIsolatedFromRoot verifies the 3-level chain:
// root(boundary) -> lv -> decorator(boundary). markItemDirty dirties ONLY
// the decorator. The ListView and root stay clean. This is the critical
// improvement: hover no longer dirties the entire tree.
func TestListView_MarkItemDirtyIsolatedFromRoot(t *testing.T) {
	builder := cdk.FuncContent[ItemContext]{Fn: func(ctx ItemContext) widget.Widget {
		w := &mockWidget{}
		w.SetVisible(true)
		return w
	}}

	lv := &Widget{}
	lv.SetVisible(true)
	lv.SetEnabled(true)
	lv.cache.list = lv
	lv.cache.update(0, 5, builder, -1, nil)

	// Build 3-level chain: root(boundary) -> lv -> decorator(boundary)
	root := &boundaryTracker{}
	root.SetVisible(true)
	root.SetRepaintBoundary(true)

	// Wire parent chain: decorator -> lv -> root
	lv.SetParent(root)
	for i := 0; i < 5; i++ {
		w := lv.cache.widgetAt(i)
		if w == nil {
			t.Fatalf("widget[%d] nil", i)
		}
		w.(*itemDecorator).SetParent(lv)
		w.(*itemDecorator).ClearRedraw()
	}
	root.ClearSceneDirty()
	root.sceneDirtied = false
	lv.ClearRedraw()

	// markItemDirty on item 2.
	lv.markItemDirty(2)

	// Decorator should be dirty.
	dec := lv.cache.widgetAt(2).(*itemDecorator)
	if !dec.NeedsRedraw() {
		t.Error("decorator[2] should need redraw after markItemDirty")
	}

	// Decorator IS boundary -> decorator scene is dirty.
	if !dec.IsSceneDirty() {
		t.Error("decorator boundary should be scene-dirty (self-invalidated)")
	}

	// Root should NOT be dirty — decorator is a boundary, propagation stops there.
	if root.sceneDirtied {
		t.Error("root boundary should NOT be dirty; " +
			"decorator IS a boundary, markItemDirty does not propagate to ListView or root")
	}

	// ListView itself should NOT be dirty.
	if lv.NeedsRedraw() {
		t.Error("ListView should NOT need redraw; " +
			"hover/selection painting is inside the decorator boundary")
	}
}

// TestListView_HoverChangesVisibleOnRedraw verifies the full hover cycle:
// mouse move -> hoveredIndex changes -> markItemDirty -> decorator dirty
// -> decorator scene re-recorded -> new hover background visible.
func TestListView_HoverChangesVisibleOnRedraw(t *testing.T) {
	builder := cdk.FuncContent[ItemContext]{Fn: func(ctx ItemContext) widget.Widget {
		w := &mockWidget{}
		w.SetVisible(true)
		return w
	}}

	lv := &Widget{}
	lv.SetVisible(true)
	lv.SetEnabled(true)
	lv.cfg = config{
		itemCount:    10,
		itemHeightFn: func(_ int) float32 { return 48 },
		overscan:     defaultOverscan,
		itemContent:  builder,
	}
	lv.painter = DefaultPainter{}
	lv.heights = newHeightManager(&lv.cfg)
	lv.cache.list = lv

	ctx := widget.NewContext()

	// Build cache so decorators exist.
	lv.hoveredIndex = noHoveredIndex
	lv.cache.update(0, 10, builder, -1, nil)

	// Clear initial dirty state.
	for i := range 10 {
		if dec, ok := lv.cache.widgetAt(i).(*itemDecorator); ok {
			dec.ClearRedraw()
		}
	}

	// Simulate mouse move at Y=100 -> should hit item 2 (48px each, item2 = 96-144).
	me := &event.MouseEvent{
		MouseType: event.MouseMove,
		Position:  geometry.Pt(200, 100),
	}
	handleContentMouseMove(lv, ctx, me)

	if lv.hoveredIndex != 2 {
		t.Errorf("hoveredIndex = %d after mouse at Y=100, want 2", lv.hoveredIndex)
	}

	// Decorator for item 2 should be dirty (boundary self-invalidation).
	if dec, ok := lv.cache.widgetAt(2).(*itemDecorator); ok {
		if !dec.NeedsRedraw() {
			t.Error("decorator[2] should be dirty after hover")
		}
	}

	// Move to item 4 (Y=200, item4 = 192-240).
	me2 := &event.MouseEvent{
		MouseType: event.MouseMove,
		Position:  geometry.Pt(200, 200),
	}
	handleContentMouseMove(lv, ctx, me2)

	if lv.hoveredIndex != 4 {
		t.Errorf("hoveredIndex = %d after mouse at Y=200, want 4", lv.hoveredIndex)
	}

	// Decorator for item 4 should be dirty, item 2 should also be dirty (old hover).
	if dec, ok := lv.cache.widgetAt(4).(*itemDecorator); ok {
		if !dec.NeedsRedraw() {
			t.Error("decorator[4] should be dirty after hover change")
		}
	}
}

// TestListView_WheelEventDispatch verifies that mouse wheel events reach
// the ScrollView inside ListView and trigger scroll + redraw.
func TestListView_WheelEventDispatch(t *testing.T) {
	lv := New(
		ItemCount(20),
		FixedItemHeight(48),
		BuildItem(func(_ ItemContext) widget.Widget {
			w := &mockWidget{}
			w.SetVisible(true)
			return w
		}),
	)

	ctx := widget.NewContext()
	invalidateCalled := false
	ctx.SetOnInvalidateRect(func(_ geometry.Rect) {
		invalidateCalled = true
	})

	constraints := geometry.Constraints{
		MinWidth: 400, MaxWidth: 400,
		MinHeight: 200, MaxHeight: 200,
	}
	lv.Layout(ctx, constraints)
	lv.SetBounds(geometry.NewRect(0, 0, 400, 200))

	widget.MountTree(lv, ctx)

	// Simulate wheel event inside viewport.
	wheel := &event.WheelEvent{
		Position: geometry.Pt(200, 100),
		Delta:    geometry.Pt(0, 3),
	}
	consumed := lv.Event(ctx, wheel)

	// ListView should forward wheel to ScrollView.
	if !consumed && !invalidateCalled {
		t.Error("wheel event not consumed and no InvalidateRect; " +
			"event may not reach ScrollView inside ListView")
	}
}

// TestListView_MouseMoveDispatchToContent verifies that MouseMove events
// reach the virtualContent and update hoveredIndex.
func TestListView_MouseMoveDispatchToContent(t *testing.T) {
	lv := New(
		ItemCount(20),
		FixedItemHeight(48),
		BuildItem(func(_ ItemContext) widget.Widget {
			w := &mockWidget{}
			w.SetVisible(true)
			return w
		}),
	)

	ctx := widget.NewContext()
	constraints := geometry.Constraints{
		MinWidth: 400, MaxWidth: 400,
		MinHeight: 200, MaxHeight: 200,
	}
	lv.Layout(ctx, constraints)
	lv.SetBounds(geometry.NewRect(0, 0, 400, 200))

	widget.MountTree(lv, ctx)

	// Draw first frame to populate cache.
	canvas := &mockCanvas{}
	lv.Draw(ctx, canvas)

	// Mouse move at Y=100 -> should hover item 2 (48px items).
	me := &event.MouseEvent{
		MouseType: event.MouseMove,
		Position:  geometry.Pt(200, 100),
	}
	lv.Event(ctx, me)

	if lv.hoveredIndex < 0 {
		t.Errorf("hoveredIndex = %d after MouseMove at Y=100, want >= 0", lv.hoveredIndex)
	}
}

// TestListView_HoverPaintCalledOnDraw verifies that after hover changes,
// a subsequent Draw() calls PaintItemBackground with Hovered=true for the
// hovered item. This ensures the decorator's painter receives correct hover state.
func TestListView_HoverPaintCalledOnDraw(t *testing.T) {
	lv := New(
		ItemCount(10),
		FixedItemHeight(48),
		BuildItem(func(_ ItemContext) widget.Widget {
			w := &mockWidget{}
			w.SetVisible(true)
			return w
		}),
	)

	ctx := widget.NewContext()
	ctx.SetOnInvalidateRect(func(_ geometry.Rect) {})

	constraints := geometry.Constraints{
		MinWidth: 400, MaxWidth: 400,
		MinHeight: 200, MaxHeight: 200,
	}
	lv.Layout(ctx, constraints)
	lv.SetBounds(geometry.NewRect(0, 0, 400, 200))
	widget.MountTree(lv, ctx)

	// First draw to populate cache.
	canvas := &mockCanvas{}
	lv.Draw(ctx, canvas)

	// Simulate hover on item 2.
	me := &event.MouseEvent{
		MouseType: event.MouseMove,
		Position:  geometry.Pt(200, 100), // Y=100 -> item 2 (48px items)
	}
	lv.Event(ctx, me)

	if lv.hoveredIndex != 2 {
		t.Fatalf("hoveredIndex = %d, want 2", lv.hoveredIndex)
	}

	// Track painter calls — the decorator calls the painter internally.
	tp := &trackingPainter{}
	lv.painter = tp

	// Draw again — the decorator for item 2 should paint with Hovered=true.
	lv.Draw(ctx, canvas)

	foundHover := false
	for _, ps := range tp.bgCalls {
		if ps.Index == 2 && ps.Hovered {
			foundHover = true
			break
		}
	}

	if !foundHover {
		t.Error("PaintItemBackground not called with Hovered=true for item 2; " +
			"hover state not reaching decorator painter on redraw")
	}
}

// trackingPainter records PaintItemBackground calls for testing.
type trackingPainter struct {
	DefaultPainter
	bgCalls []ItemPaintState
}

func (p *trackingPainter) PaintItemBackground(_ widget.Canvas, ps ItemPaintState) {
	p.bgCalls = append(p.bgCalls, ps)
}

// TestListView_RootBoundaryCacheInvalidatedOnHover verifies that hover changes
// do NOT dirty the root boundary. With the decorator pattern, hover painting
// is inside the decorator — only the decorator's scene is re-recorded.
func TestListView_RootBoundaryCacheInvalidatedOnHover(t *testing.T) {
	lv := New(
		ItemCount(10),
		FixedItemHeight(48),
		BuildItem(func(_ ItemContext) widget.Widget {
			w := &mockWidget{}
			w.SetVisible(true)
			return w
		}),
	)

	ctx := widget.NewContext()
	ctx.SetOnInvalidateRect(func(_ geometry.Rect) {})

	constraints := geometry.Constraints{
		MinWidth: 400, MaxWidth: 400,
		MinHeight: 200, MaxHeight: 200,
	}
	lv.Layout(ctx, constraints)
	lv.SetBounds(geometry.NewRect(0, 0, 400, 200))
	widget.MountTree(lv, ctx)

	// Create root boundary tracker.
	root := &boundaryTracker{}
	root.SetVisible(true)
	root.SetRepaintBoundary(true)
	lv.SetParent(root)

	// Wire decorator widgets to lv as parent.
	for i := 0; i < len(lv.cache.widgets); i++ {
		if w := lv.cache.widgetAt(i); w != nil {
			if setter, ok := w.(interface{ SetParent(widget.Widget) }); ok {
				setter.SetParent(lv)
			}
		}
	}

	// First draw.
	canvas := &mockCanvas{}
	lv.Draw(ctx, canvas)

	// Clear all dirty state.
	root.ClearSceneDirty()
	root.sceneDirtied = false
	lv.ClearRedraw()
	for i := 0; i < len(lv.cache.widgets); i++ {
		if w := lv.cache.widgetAt(i); w != nil {
			if clearer, ok := w.(interface {
				ClearSceneDirty()
				ClearRedraw()
			}); ok {
				clearer.ClearSceneDirty()
				clearer.ClearRedraw()
			}
		}
	}

	// Simulate hover change on item 2.
	me := &event.MouseEvent{
		MouseType: event.MouseMove,
		Position:  geometry.Pt(200, 100),
	}
	lv.Event(ctx, me)

	if lv.hoveredIndex != 2 {
		t.Fatalf("hoveredIndex = %d, want 2", lv.hoveredIndex)
	}

	// Root should NOT be dirty — decorator is the boundary, hover painting
	// is inside the decorator. Only the decorator's scene is re-recorded.
	if root.sceneDirtied {
		t.Error("root boundary should NOT be dirty on hover; " +
			"decorator owns hover painting, root tree stays clean")
	}
}

// TestListView_BoundaryItemBoundsInContentSpace verifies that decorators
// have bounds set in content space (Y = cumulative item offset from content
// start), NOT in viewport/screen space.
func TestListView_BoundaryItemBoundsInContentSpace(t *testing.T) {
	builder := cdk.FuncContent[ItemContext]{Fn: func(ctx ItemContext) widget.Widget {
		w := &mockWidget{}
		w.SetVisible(true)
		return w
	}}

	wc := newTestCache()
	wc.update(0, 5, builder, -1, nil)

	for i := 0; i < 5; i++ {
		w := wc.widgetAt(i)
		if w == nil {
			t.Fatalf("widget[%d] nil", i)
		}
		bounds := geometry.NewRect(0, float32(i*48), 400, 48)
		w.(*itemDecorator).SetBounds(bounds)

		got := w.(*itemDecorator).Bounds()
		if got.Min.Y != float32(i*48) {
			t.Errorf("widget[%d] bounds.Min.Y = %v, want %v",
				i, got.Min.Y, float32(i*48))
		}
	}
}

// TestWidgetCache_ChildAt verifies that childAt returns the inner user widget.
func TestWidgetCache_ChildAt(t *testing.T) {
	wc := newTestCache()
	widgets := make([]*mockWidget, 3)
	builder := cdk.FuncContent[ItemContext]{Fn: func(ctx ItemContext) widget.Widget {
		w := &mockWidget{}
		w.SetVisible(true)
		widgets[ctx.Index] = w
		return w
	}}

	wc.update(0, 3, builder, -1, nil)

	for i := 0; i < 3; i++ {
		child := wc.childAt(i)
		if child == nil {
			t.Fatalf("childAt(%d) is nil", i)
		}
		if child != widgets[i] {
			t.Errorf("childAt(%d) != original widget", i)
		}
	}

	// Out of range.
	if child := wc.childAt(-1); child != nil {
		t.Error("childAt(-1) should return nil")
	}
	if child := wc.childAt(3); child != nil {
		t.Error("childAt(3) should return nil")
	}
}

// TestItemDecorator_Children verifies decorator.Children() returns the child.
func TestItemDecorator_Children(t *testing.T) {
	child := &mockWidget{}
	child.SetVisible(true)
	lv := &Widget{painter: DefaultPainter{}}
	dec := newItemDecorator(child, lv, 0)

	children := dec.Children()
	if len(children) != 1 {
		t.Fatalf("len(Children()) = %d, want 1", len(children))
	}
	if children[0] != child {
		t.Error("Children()[0] should be the child widget")
	}

	// Nil child.
	dec2 := newItemDecorator(nil, lv, 0)
	if got := dec2.Children(); got != nil {
		t.Errorf("Children() with nil child = %v, want nil", got)
	}
}

// TestItemDecorator_Event verifies decorator.Event always returns false.
func TestItemDecorator_Event(t *testing.T) {
	dec := newItemDecorator(nil, nil, 0)
	if dec.Event(nil, &event.MouseEvent{}) {
		t.Error("decorator.Event should always return false")
	}
}

// --- Scroll recycling tests (TDD: define correct behavior, then fix code) ---

// TestScrollSameRange_NoRebuild verifies that scrolling within the same
// visible range does NOT recreate decorators. Sub-pixel scroll should be a no-op.
func TestScrollSameRange_NoRebuild(t *testing.T) {
	wc := newTestCache()
	callCount := 0
	builder := cdk.FuncContent[ItemContext]{Fn: func(_ ItemContext) widget.Widget {
		callCount++
		w := &mockWidget{}
		w.SetVisible(true)
		return w
	}}

	wc.update(0, 5, builder, -1, nil)
	first := make([]widget.Widget, 5)
	for i := range 5 {
		first[i] = wc.widgetAt(i)
	}
	callCount = 0

	// Simulate scroll that doesn't change visible range.
	// Without cache.invalidate(), update sees same range → no-op.
	wc.update(0, 5, builder, -1, nil)

	if callCount != 0 {
		t.Errorf("builder called %d times, want 0 (same range = no rebuild)", callCount)
	}
	for i := range 5 {
		if wc.widgetAt(i) != first[i] {
			t.Errorf("decorator[%d] changed, want same instance (reuse)", i)
		}
	}
}

// TestScrollShiftRange_ReusesOverlap verifies that when visible range shifts
// by a few items, overlapping decorators are REUSED (same instance) and only
// edge decorators are newly created. RecyclerView/Flutter pattern.
func TestScrollShiftRange_ReusesOverlap(t *testing.T) {
	wc := newTestCache()
	callCount := 0
	builder := cdk.FuncContent[ItemContext]{Fn: func(_ ItemContext) widget.Widget {
		callCount++
		w := &mockWidget{}
		w.SetVisible(true)
		return w
	}}

	// Initial range [0, 8).
	wc.update(0, 8, builder, -1, nil)
	origDec := make([]widget.Widget, 8)
	for i := range 8 {
		origDec[i] = wc.widgetAt(i)
	}
	callCount = 0

	// Scroll down by 2 items → new range [2, 10).
	// Items 2-7 overlap → reuse. Items 8-9 are new.
	wc.update(2, 10, builder, -1, nil)

	// Only 2 new items should be built (indices 8, 9).
	if callCount != 2 {
		t.Errorf("builder called %d times, want 2 (only new edge items)", callCount)
	}

	// Overlapping items (indices 2-7) should be same decorator instances.
	for i := 2; i < 8; i++ {
		newOffset := i - 2
		if wc.widgetAt(newOffset) != origDec[i] {
			t.Errorf("decorator for index %d not reused (offset %d)", i, newOffset)
		}
	}
}

// TestScrollShiftRange_NewEdgeDecoratorsAreDirty verifies that newly created
// edge decorators start dirty (need first draw), while reused ones stay clean.
func TestScrollShiftRange_NewEdgeDecoratorsAreDirty(t *testing.T) {
	wc := newTestCache()
	builder := cdk.FuncContent[ItemContext]{Fn: func(_ ItemContext) widget.Widget {
		w := &mockWidget{}
		w.SetVisible(true)
		return w
	}}

	wc.update(0, 8, builder, -1, nil)

	// Simulate first draw — clear dirty on all decorators.
	for i := range 8 {
		if dec, ok := wc.widgetAt(i).(*itemDecorator); ok {
			dec.ClearRedraw()
		}
	}

	// Scroll down by 2 → range [2, 10).
	wc.update(2, 10, builder, -1, nil)

	// Reused decorators (offsets 0-5, indices 2-7) should be clean.
	for i := 0; i < 6; i++ {
		dec, ok := wc.widgetAt(i).(*itemDecorator)
		if !ok {
			t.Fatalf("widgetAt(%d) is not itemDecorator", i)
		}
		if dec.NeedsRedraw() {
			t.Errorf("reused decorator at offset %d (index %d) should be clean", i, i+2)
		}
	}

	// New edge decorators (offsets 6-7, indices 8-9) should be dirty.
	for i := 6; i < 8; i++ {
		dec, ok := wc.widgetAt(i).(*itemDecorator)
		if !ok {
			t.Fatalf("widgetAt(%d) is not itemDecorator", i)
		}
		if !dec.NeedsRedraw() {
			t.Errorf("new edge decorator at offset %d (index %d) should be dirty", i, i+2)
		}
	}
}

// TestScrollNoOverlap_FullRebuild verifies that when visible range changes
// completely (no overlap), all decorators are freshly built.
func TestScrollNoOverlap_FullRebuild(t *testing.T) {
	wc := newTestCache()
	callCount := 0
	builder := cdk.FuncContent[ItemContext]{Fn: func(_ ItemContext) widget.Widget {
		callCount++
		w := &mockWidget{}
		w.SetVisible(true)
		return w
	}}

	wc.update(0, 5, builder, -1, nil)
	callCount = 0

	// Jump to completely different range.
	wc.update(100, 105, builder, -1, nil)

	if callCount != 5 {
		t.Errorf("builder called %d times, want 5 (no overlap = full rebuild)", callCount)
	}
}

// TestScrollInvalidateThenSameRange_Rebuilds verifies that cache.invalidate()
// followed by update with the same range DOES rebuild (data may have changed).
func TestScrollInvalidateThenSameRange_Rebuilds(t *testing.T) {
	wc := newTestCache()
	callCount := 0
	builder := cdk.FuncContent[ItemContext]{Fn: func(_ ItemContext) widget.Widget {
		callCount++
		return nil
	}}

	wc.update(0, 5, builder, -1, nil)
	callCount = 0

	wc.invalidate()
	wc.update(0, 5, builder, -1, nil)

	if callCount != 5 {
		t.Errorf("callCount = %d, want 5 (invalidate forces rebuild)", callCount)
	}
}

// TestItemDecorator_LayoutDelegatesToChild verifies decorator.Layout
// delegates to the child widget.
func TestItemDecorator_LayoutDelegatesToChild(t *testing.T) {
	child := &mockWidget{}
	child.SetVisible(true)
	lv := &Widget{painter: DefaultPainter{}}
	dec := newItemDecorator(child, lv, 0)

	c := geometry.Constraints{
		MinWidth: 200, MaxWidth: 200,
		MinHeight: 0, MaxHeight: geometry.Infinity,
	}
	size := dec.Layout(nil, c)

	// mockWidget.Layout returns c.Constrain(Sz(100, 48)).
	// With MinWidth=MaxWidth=200, width is clamped to 200.
	if size.Width != 200 || size.Height != 48 {
		t.Errorf("Layout = %v, want (200, 48)", size)
	}

	// Nil child.
	dec2 := newItemDecorator(nil, lv, 0)
	size2 := dec2.Layout(nil, c)
	if size2 != (geometry.Size{}) {
		t.Errorf("Layout with nil child = %v, want zero", size2)
	}
}
