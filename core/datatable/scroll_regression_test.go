package datatable

import (
	"testing"

	"github.com/gogpu/ui/geometry"
	"github.com/gogpu/ui/state"
	"github.com/gogpu/ui/widget"
)

// rowYRecordingPainter is a Painter that records the top Y of every row it is
// asked to paint, so tests can assert where rows land after scrolling.
type rowYRecordingPainter struct {
	DefaultPainter
	rows []struct {
		index int
		y     float32
	}
}

func (p *rowYRecordingPainter) PaintRow(canvas widget.Canvas, s RowPaintState) {
	p.rows = append(p.rows, struct {
		index int
		y     float32
	}{s.RowIndex, s.Bounds.Min.Y})
	p.DefaultPainter.PaintRow(canvas, s)
}

func (p *rowYRecordingPainter) yFor(rowIndex int) (float32, bool) {
	for _, r := range p.rows {
		if r.index == rowIndex {
			return r.y, true
		}
	}
	return 0, false
}

// TestDrawVisibleRows_NoDoubleScroll is a regression test for the double-scroll
// bug: the parent ScrollView already applies a translate(0, -scrollY) transform
// before calling Draw, so drawVisibleRows must place rows at their absolute
// content-space offset (row*rowHeight). A previous implementation subtracted
// scrollY a second time, which moved rows at twice the scroll speed — content
// ran off the end (blank space) and hit-testing via rowAtY desynced from the
// visible highlight.
//
// The assertion: with scrollY = 260 and rowHeight = 26, the first visible row
// (index 10) must be painted at content-space Y = 260, NOT at 260 - 260 = 0
// (the double-scrolled value).
func TestDrawVisibleRows_NoDoubleScroll(t *testing.T) {
	const rowHeight float32 = 26
	const scrollY float32 = 260 // exactly 10 rows down

	painter := &rowYRecordingPainter{}
	scrollSig := state.NewSignal(scrollY)

	dt := New(
		Columns(testColumns()),
		RowCount(1000),
		RowHeight(rowHeight),
		ScrollYSignal(scrollSig),
		PainterOpt(painter),
	)

	ctx := widget.NewContext()

	// Lay out and position the table so the internal scroll view has bounds.
	dt.Layout(ctx, geometry.Tight(geometry.Sz(600, 400)))
	dt.SetBounds(geometry.NewRect(0, 0, 600, 400))
	dt.updateScrollBounds()

	// Draw, which dispatches through the scroll view to drawVisibleRows.
	dt.Draw(ctx, &mockCanvas{})

	// Row 10 is the first row scrolled into view (260 / 26 = 10).
	const firstVisible = 10
	wantY := float32(firstVisible) * rowHeight // 260, content-space absolute

	gotY, ok := painter.yFor(firstVisible)
	if !ok {
		t.Fatalf("row %d was not painted; visible rows = %+v", firstVisible, painter.rows)
	}

	// The double-scroll bug produces wantY - scrollY (== 0 here).
	doubleScrolled := wantY - scrollY
	if gotY == doubleScrolled {
		t.Fatalf("row %d painted at Y=%.0f (double-scrolled); ScrollView already "+
			"applies -scrollY, drawVisibleRows must not subtract it again", firstVisible, gotY)
	}
	if gotY != wantY {
		t.Errorf("row %d painted at Y=%.0f, want %.0f (content-space row*rowHeight)",
			firstVisible, gotY, wantY)
	}
}

// TestDrawVisibleRows_RowSpacingIsRowHeight verifies adjacent visible rows are
// spaced exactly rowHeight apart in content space, independent of scroll — a
// second guard against off-by-scroll regressions in row placement.
func TestDrawVisibleRows_RowSpacingIsRowHeight(t *testing.T) {
	const rowHeight float32 = 30
	const scrollY float32 = 150 // 5 rows

	painter := &rowYRecordingPainter{}
	scrollSig := state.NewSignal(scrollY)

	dt := New(
		Columns(testColumns()),
		RowCount(500),
		RowHeight(rowHeight),
		ScrollYSignal(scrollSig),
		PainterOpt(painter),
	)

	ctx := widget.NewContext()
	dt.Layout(ctx, geometry.Tight(geometry.Sz(600, 400)))
	dt.SetBounds(geometry.NewRect(0, 0, 600, 400))
	dt.updateScrollBounds()
	dt.Draw(ctx, &mockCanvas{})

	y5, ok5 := painter.yFor(5)
	y6, ok6 := painter.yFor(6)
	if !ok5 || !ok6 {
		t.Fatalf("expected rows 5 and 6 painted; got %+v", painter.rows)
	}
	if gap := y6 - y5; gap != rowHeight {
		t.Errorf("row spacing = %.0f, want %.0f (rowHeight)", gap, rowHeight)
	}
}
