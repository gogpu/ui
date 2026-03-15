package listview

import (
	"image"
	"testing"

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

func TestWidgetCache_Update(t *testing.T) {
	var wc widgetCache
	builder := cdk.FuncContent[ItemContext]{Fn: func(ctx ItemContext) widget.Widget {
		return nil // simple builder for testing
	}}

	wc.update(0, 5, builder, -1, -1)

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
	var wc widgetCache
	callCount := 0
	builder := cdk.FuncContent[ItemContext]{Fn: func(ctx ItemContext) widget.Widget {
		callCount++
		return nil
	}}

	wc.update(0, 5, builder, -1, -1)
	if callCount != 5 {
		t.Errorf("callCount = %d, want 5", callCount)
	}

	// Second call with same range should be no-op.
	wc.update(0, 5, builder, -1, -1)
	if callCount != 5 {
		t.Errorf("callCount after reuse = %d, want 5", callCount)
	}
}

func TestWidgetCache_InvalidateForces_Rebuild(t *testing.T) {
	var wc widgetCache
	callCount := 0
	builder := cdk.FuncContent[ItemContext]{Fn: func(ctx ItemContext) widget.Widget {
		callCount++
		return nil
	}}

	wc.update(0, 5, builder, -1, -1)
	wc.invalidate()
	wc.update(0, 5, builder, -1, -1)

	if callCount != 10 {
		t.Errorf("callCount = %d, want 10 (5+5 after invalidate)", callCount)
	}
}

func TestWidgetCache_Clear(t *testing.T) {
	var wc widgetCache
	builder := cdk.FuncContent[ItemContext]{Fn: func(ctx ItemContext) widget.Widget {
		return nil
	}}

	wc.update(0, 5, builder, -1, -1)
	wc.clear()

	if wc.valid {
		t.Error("cache should be invalid after clear")
	}
	if len(wc.widgets) != 0 {
		t.Errorf("len(widgets) = %d, want 0", len(wc.widgets))
	}
}

func TestWidgetCache_EmptyRange(t *testing.T) {
	var wc widgetCache
	builder := cdk.FuncContent[ItemContext]{Fn: func(ctx ItemContext) widget.Widget {
		return nil
	}}

	wc.update(0, 0, builder, -1, -1)

	if wc.valid {
		t.Error("cache should not be valid for empty range")
	}
}

func TestWidgetCache_WidgetAt(t *testing.T) {
	var wc widgetCache

	// Empty cache.
	if got := wc.widgetAt(0); got != nil {
		t.Error("widgetAt on empty cache should return nil")
	}
	if got := wc.widgetAt(-1); got != nil {
		t.Error("widgetAt(-1) should return nil")
	}
}

func TestWidgetCache_NilBuilder(t *testing.T) {
	var wc widgetCache
	wc.update(0, 3, nil, -1, -1)

	for i := 0; i < 3; i++ {
		if got := wc.widgetAt(i); got != nil {
			t.Errorf("widgetAt(%d) with nil builder should return nil", i)
		}
	}
}

func TestWidgetCache_ItemContextPropagation(t *testing.T) {
	var wc widgetCache
	var received []ItemContext
	builder := cdk.FuncContent[ItemContext]{Fn: func(ctx ItemContext) widget.Widget {
		received = append(received, ctx)
		return nil
	}}

	wc.update(5, 8, builder, 6, 7)

	if len(received) != 3 {
		t.Fatalf("received %d contexts, want 3", len(received))
	}

	// Index 5: not selected, not hovered.
	if received[0].Index != 5 || received[0].Selected || received[0].Hovered {
		t.Errorf("ctx[0] = %+v, want Index=5, Selected=false, Hovered=false", received[0])
	}
	// Index 6: selected.
	if received[1].Index != 6 || !received[1].Selected {
		t.Errorf("ctx[1] = %+v, want Index=6, Selected=true", received[1])
	}
	// Index 7: hovered.
	if received[2].Index != 7 || !received[2].Hovered {
		t.Errorf("ctx[2] = %+v, want Index=7, Hovered=true", received[2])
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

// --- mockCanvas for internal tests ---

type mockCanvas struct{}

func (m *mockCanvas) Clear(_ widget.Color)                                                  {}
func (m *mockCanvas) DrawRect(_ geometry.Rect, _ widget.Color)                              {}
func (m *mockCanvas) StrokeRect(_ geometry.Rect, _ widget.Color, _ float32)                 {}
func (m *mockCanvas) DrawRoundRect(_ geometry.Rect, _ widget.Color, _ float32)              {}
func (m *mockCanvas) StrokeRoundRect(_ geometry.Rect, _ widget.Color, _ float32, _ float32) {}
func (m *mockCanvas) DrawCircle(_ geometry.Point, _ float32, _ widget.Color)                {}
func (m *mockCanvas) StrokeCircle(_ geometry.Point, _ float32, _ widget.Color, _ float32)   {}
func (m *mockCanvas) DrawLine(_, _ geometry.Point, _ widget.Color, _ float32)               {}
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
