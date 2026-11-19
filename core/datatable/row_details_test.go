package datatable

import (
	"testing"

	"github.com/gogpu/ui/core/checkbox"
	"github.com/gogpu/ui/event"
	"github.com/gogpu/ui/geometry"
	"github.com/gogpu/ui/uitest"
	"github.com/gogpu/ui/widget"
)

func TestRowDetailsGeometryAndInput(t *testing.T) {
	count := 10
	dt := New(RowCountFn(func() int { return count }), RowHeight(24), SelectionModeOpt(SelectionSingle))
	ctx := widget.NewContext()
	size := geometry.Sz(400, 180)
	dt.SetBounds(geometry.FromPointSize(geometry.Point{}, size))
	widget.MountTree(dt, ctx)
	checked := false
	control := checkbox.New(checkbox.Label("Inside details"), checkbox.OnToggle(func(on bool) { checked = on }))
	dt.SetRowDetails(1, control, 100)
	widget.LayoutChild(dt, ctx, geometry.Tight(size))
	dt.Draw(ctx, &uitest.MockCanvas{})

	if want := geometry.NewRect(0, 48, 400-dt.scroll.ScrollbarInset(), 100); control.Bounds() != want || !control.IsMounted() {
		t.Fatalf("detail widget bounds = %v, want %v", control.Bounds(), want)
	}
	if dt.GetRowCount() != 10 || dt.totalDataHeight() != 340 || dt.rowAtY(48) != -1 || dt.rowAtY(148) != 2 {
		t.Fatal("adding details changed row indices or did not reserve height")
	}
	point := geometry.Pt(30, defaultHeaderHeight+60)
	dt.Event(ctx, event.NewMouseEvent(event.MousePress, event.ButtonLeft, event.ButtonStateLeft, point, point, event.ModNone))
	dt.Event(ctx, event.NewMouseEvent(event.MouseRelease, event.ButtonLeft, 0, point, point, event.ModNone))
	if !checked || dt.cfg.ResolvedSelectedRow() != -1 {
		t.Fatal("detail input did not reach the checkbox independently of row selection")
	}
	dt.ScrollToRow(2)
	if got := dt.currentScrollY(); got != 28 {
		t.Fatalf("ScrollToRow(2) = %v, want 28", got)
	}
	count = 1
	dt.InvalidateData()
	widget.LayoutChild(dt, ctx, geometry.Tight(size))
	if control.IsMounted() || dt.currentScrollY() != 0 || dt.totalDataHeight() != 24 {
		t.Fatal("details or scroll position did not reset when the owning row disappeared")
	}
}
