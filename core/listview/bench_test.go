package listview_test

import (
	"testing"

	"github.com/gogpu/ui/core/listview"
	"github.com/gogpu/ui/geometry"
	"github.com/gogpu/ui/state"
	"github.com/gogpu/ui/widget"
)

func newBenchListView(itemCount int) *listview.Widget {
	return listview.New(
		listview.ItemCount(itemCount),
		listview.FixedItemHeight(48),
		listview.BuildItem(func(_ listview.ItemContext) widget.Widget {
			return nil
		}),
	)
}

func BenchmarkListViewLayout1000(b *testing.B) {
	lv := newBenchListView(1000)
	ctx := widget.NewContext()
	constraints := geometry.Constraints{
		MinWidth: 300, MaxWidth: 300,
		MinHeight: 600, MaxHeight: 600,
	}

	b.ReportAllocs()
	b.ResetTimer()
	for b.Loop() {
		lv.Layout(ctx, constraints)
	}
}

func BenchmarkListViewScroll(b *testing.B) {
	scrollY := state.NewSignal[float32](0)
	lv := listview.New(
		listview.ItemCount(1000),
		listview.FixedItemHeight(48),
		listview.ScrollYSignal(scrollY),
		listview.BuildItem(func(_ listview.ItemContext) widget.Widget {
			return nil
		}),
	)
	ctx := widget.NewContext()
	constraints := geometry.Constraints{
		MinWidth: 300, MaxWidth: 300,
		MinHeight: 600, MaxHeight: 600,
	}
	// Initial layout.
	lv.Layout(ctx, constraints)
	lv.SetBounds(geometry.NewRect(0, 0, 300, 600))

	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; b.Loop(); i++ {
		// Simulate scroll by varying position.
		scrollY.Set(float32(i%10000) * 0.5)
		lv.Layout(ctx, constraints)
	}
}

func BenchmarkListViewSelection(b *testing.B) {
	selectedIdx := state.NewSignal(-1)
	lv := listview.New(
		listview.ItemCount(1000),
		listview.FixedItemHeight(48),
		listview.SelectionModeOpt(listview.SelectionSingle),
		listview.SelectedIndexSignal(selectedIdx),
		listview.BuildItem(func(_ listview.ItemContext) widget.Widget {
			return nil
		}),
	)
	ctx := widget.NewContext()
	constraints := geometry.Constraints{
		MinWidth: 300, MaxWidth: 300,
		MinHeight: 600, MaxHeight: 600,
	}
	lv.Layout(ctx, constraints)
	lv.SetBounds(geometry.NewRect(0, 0, 300, 600))

	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; b.Loop(); i++ {
		selectedIdx.Set(i % 1000)
		lv.Layout(ctx, constraints)
	}
}
