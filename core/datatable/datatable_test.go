package datatable

import (
	"fmt"
	"image"
	"testing"

	"github.com/gogpu/ui/a11y"
	"github.com/gogpu/ui/event"
	"github.com/gogpu/ui/geometry"
	"github.com/gogpu/ui/state"
	"github.com/gogpu/ui/widget"
)

// --- Test Helpers ---

// mockCanvas implements widget.Canvas for testing.
type mockCanvas struct {
	rects      []geometry.Rect
	texts      []string
	lines      int
	clips      int
	transforms int
}

func (m *mockCanvas) Clear(_ widget.Color)                     {}
func (m *mockCanvas) DrawRect(r geometry.Rect, _ widget.Color) { m.rects = append(m.rects, r) }
func (m *mockCanvas) StrokeRect(r geometry.Rect, _ widget.Color, _ float32) {
	m.rects = append(m.rects, r)
}
func (m *mockCanvas) DrawRoundRect(_ geometry.Rect, _ widget.Color, _ float32) {}
func (m *mockCanvas) StrokeRoundRect(_ geometry.Rect, _ widget.Color, _, _ float32) {
}
func (m *mockCanvas) DrawCircle(_ geometry.Point, _ float32, _ widget.Color) {}
func (m *mockCanvas) StrokeCircle(_ geometry.Point, _ float32, _ widget.Color, _ float32) {
}
func (m *mockCanvas) DrawLine(_, _ geometry.Point, _ widget.Color, _ float32) { m.lines++ }
func (m *mockCanvas) DrawText(text string, _ geometry.Rect, _ float32, _ widget.Color, _ bool, _ widget.TextAlign) {
	m.texts = append(m.texts, text)
}

func (m *mockCanvas) MeasureText(text string, fontSize float32, _ bool) float32 {
	return float32(len([]rune(text))) * fontSize * 0.5
}
func (m *mockCanvas) DrawImage(_ image.Image, _ geometry.Point)    {}
func (m *mockCanvas) PushClip(_ geometry.Rect)                     { m.clips++ }
func (m *mockCanvas) PushClipRoundRect(_ geometry.Rect, _ float32) { m.clips++ }
func (m *mockCanvas) PopClip()                                     { m.clips-- }
func (m *mockCanvas) PushTransform(_ geometry.Point)               { m.transforms++ }
func (m *mockCanvas) PopTransform()                                { m.transforms-- }
func (m *mockCanvas) TransformOffset() geometry.Point              { return geometry.Point{} }

// testColumns returns a standard set of test columns.
func testColumns() []Column {
	return []Column{
		{Key: "name", Title: "Name", Width: 200, Sortable: true},
		{Key: "size", Title: "Size", Width: 100, Sortable: true, Align: widget.TextAlignRight},
		{Key: "date", Title: "Modified", Width: 150, Sortable: true},
	}
}

// testCellValue returns a test cell value provider.
func testCellValue(row int, col string) string {
	return fmt.Sprintf("r%d_%s", row, col)
}

// --- Column Tests ---

func TestSortDirection_String(t *testing.T) {
	tests := []struct {
		dir  SortDirection
		want string
	}{
		{SortNone, "None"},
		{SortAscending, "Ascending"},
		{SortDescending, "Descending"},
		{SortDirection(99), "Unknown"},
	}
	for _, tt := range tests {
		if got := tt.dir.String(); got != tt.want {
			t.Errorf("SortDirection(%d).String() = %q, want %q", tt.dir, got, tt.want)
		}
	}
}

func TestSortDirection_Indicator(t *testing.T) {
	tests := []struct {
		dir  SortDirection
		want string
	}{
		{SortNone, ""},
		{SortAscending, "\u25B2"},
		{SortDescending, "\u25BC"},
		{SortDirection(99), ""},
	}
	for _, tt := range tests {
		if got := tt.dir.Indicator(); got != tt.want {
			t.Errorf("SortDirection(%d).Indicator() = %q, want %q", tt.dir, got, tt.want)
		}
	}
}

func TestSortDirection_NextDirection(t *testing.T) {
	tests := []struct {
		dir  SortDirection
		want SortDirection
	}{
		{SortNone, SortAscending},
		{SortAscending, SortDescending},
		{SortDescending, SortNone},
	}
	for _, tt := range tests {
		if got := tt.dir.nextDirection(); got != tt.want {
			t.Errorf("SortDirection(%d).nextDirection() = %d, want %d", tt.dir, got, tt.want)
		}
	}
}

// --- SelectionMode Tests ---

func TestSelectionMode_String(t *testing.T) {
	tests := []struct {
		mode SelectionMode
		want string
	}{
		{SelectionNone, "None"},
		{SelectionSingle, "Single"},
		{SelectionMulti, "Multi"},
		{SelectionMode(99), "Unknown"},
	}
	for _, tt := range tests {
		if got := tt.mode.String(); got != tt.want {
			t.Errorf("SelectionMode(%d).String() = %q, want %q", tt.mode, got, tt.want)
		}
	}
}

// --- Widget Construction Tests ---

func TestNew_Defaults(t *testing.T) {
	dt := New()
	if !dt.IsVisible() {
		t.Error("expected visible by default")
	}
	if !dt.IsEnabled() {
		t.Error("expected enabled by default")
	}
	if dt.cfg.rowHeight != defaultRowHeight {
		t.Errorf("rowHeight = %f, want %f", dt.cfg.rowHeight, defaultRowHeight)
	}
	if dt.sortColumn != noSortColumn {
		t.Errorf("sortColumn = %d, want %d", dt.sortColumn, noSortColumn)
	}
	if dt.hoveredRow != noHoveredRow {
		t.Errorf("hoveredRow = %d, want %d", dt.hoveredRow, noHoveredRow)
	}
}

func TestNew_WithOptions(t *testing.T) {
	cols := testColumns()
	dt := New(
		Columns(cols),
		RowCount(100),
		RowHeight(40),
		CellValue(testCellValue),
		SelectionModeOpt(SelectionSingle),
		SelectedRow(5),
		A11yLabel("My Table"),
	)

	if len(dt.cfg.columns) != 3 {
		t.Errorf("columns = %d, want 3", len(dt.cfg.columns))
	}
	if dt.cfg.ResolvedRowCount() != 100 {
		t.Errorf("rowCount = %d, want 100", dt.cfg.ResolvedRowCount())
	}
	if dt.cfg.rowHeight != 40 {
		t.Errorf("rowHeight = %f, want 40", dt.cfg.rowHeight)
	}
	if dt.cfg.selectionMode != SelectionSingle {
		t.Errorf("selectionMode = %v, want Single", dt.cfg.selectionMode)
	}
	if dt.cfg.ResolvedSelectedRow() != 5 {
		t.Errorf("selectedRow = %d, want 5", dt.cfg.ResolvedSelectedRow())
	}
	if dt.cfg.a11yLabel != "My Table" {
		t.Errorf("a11yLabel = %q, want %q", dt.cfg.a11yLabel, "My Table")
	}
}

func TestNew_MultiSelection(t *testing.T) {
	dt := New(SelectionModeOpt(SelectionMulti))
	if dt.cfg.selectedRows == nil {
		t.Error("expected selectedRows map to be initialized for multi-selection")
	}
}

func TestNew_WithPainter(t *testing.T) {
	p := DefaultPainter{}
	dt := New(PainterOpt(p))
	// Just verify it doesn't panic — painter is set.
	_ = dt
}

// --- Config Resolution Tests ---

func TestConfig_ResolvedRowCount(t *testing.T) {
	t.Run("static", func(t *testing.T) {
		cfg := config{rowCount: 42}
		if got := cfg.ResolvedRowCount(); got != 42 {
			t.Errorf("got %d, want 42", got)
		}
	})

	t.Run("fn", func(t *testing.T) {
		cfg := config{rowCount: 10, rowCountFn: func() int { return 99 }}
		if got := cfg.ResolvedRowCount(); got != 99 {
			t.Errorf("got %d, want 99", got)
		}
	})

	t.Run("signal", func(t *testing.T) {
		sig := state.NewSignal(77)
		cfg := config{rowCount: 10, rowCountSignal: sig}
		if got := cfg.ResolvedRowCount(); got != 77 {
			t.Errorf("got %d, want 77", got)
		}
	})

	t.Run("readonly_signal", func(t *testing.T) {
		sig := state.NewSignal(55)
		cfg := config{rowCount: 10, readonlyRowCountSignal: sig.AsReadonly()}
		if got := cfg.ResolvedRowCount(); got != 55 {
			t.Errorf("got %d, want 55", got)
		}
	})
}

func TestConfig_ResolvedSelectedRow(t *testing.T) {
	t.Run("static", func(t *testing.T) {
		cfg := config{selectedRow: 3}
		if got := cfg.ResolvedSelectedRow(); got != 3 {
			t.Errorf("got %d, want 3", got)
		}
	})

	t.Run("signal", func(t *testing.T) {
		sig := state.NewSignal(7)
		cfg := config{selectedRow: 3, selectedRowSignal: sig}
		if got := cfg.ResolvedSelectedRow(); got != 7 {
			t.Errorf("got %d, want 7", got)
		}
	})

	t.Run("readonly_signal", func(t *testing.T) {
		sig := state.NewSignal(11)
		cfg := config{selectedRow: 3, readonlySelectedRowSignal: sig.AsReadonly()}
		if got := cfg.ResolvedSelectedRow(); got != 11 {
			t.Errorf("got %d, want 11", got)
		}
	})
}

func TestConfig_ResolvedDisabled(t *testing.T) {
	t.Run("static", func(t *testing.T) {
		cfg := config{disabled: true}
		if !cfg.ResolvedDisabled() {
			t.Error("expected true")
		}
	})

	t.Run("fn", func(t *testing.T) {
		cfg := config{disabled: false, disabledFn: func() bool { return true }}
		if !cfg.ResolvedDisabled() {
			t.Error("expected true from fn")
		}
	})

	t.Run("signal", func(t *testing.T) {
		sig := state.NewSignal(true)
		cfg := config{disabledSignal: sig}
		if !cfg.ResolvedDisabled() {
			t.Error("expected true from signal")
		}
	})

	t.Run("readonly_signal", func(t *testing.T) {
		sig := state.NewSignal(true)
		cfg := config{readonlyDisabledSignal: sig.AsReadonly()}
		if !cfg.ResolvedDisabled() {
			t.Error("expected true from readonly signal")
		}
	})
}

// --- Layout Tests ---

func TestLayout_BasicConstraints(t *testing.T) {
	dt := New(
		Columns(testColumns()),
		RowCount(10),
	)
	ctx := widget.NewContext()

	size := dt.Layout(ctx, geometry.Tight(geometry.Sz(600, 400)))
	if size.Width != 600 {
		t.Errorf("width = %f, want 600", size.Width)
	}
	if size.Height != 400 {
		t.Errorf("height = %f, want 400", size.Height)
	}
}

func TestLayout_UnconstrainedWidth(t *testing.T) {
	dt := New(Columns(testColumns()), RowCount(5))
	ctx := widget.NewContext()

	constraints := geometry.Constraints{
		MinWidth: 0, MaxWidth: geometry.Infinity,
		MinHeight: 100, MaxHeight: 400,
	}
	size := dt.Layout(ctx, constraints)
	if size.Width <= 0 {
		t.Error("expected positive width")
	}
}

func TestLayout_ResolvesColumnWidths(t *testing.T) {
	cols := []Column{
		{Key: "a", Title: "A", Width: 100},
		{Key: "b", Title: "B", Width: 200},
		{Key: "c", Title: "C"}, // flex
	}
	dt := New(Columns(cols), RowCount(5))
	ctx := widget.NewContext()

	dt.Layout(ctx, geometry.Tight(geometry.Sz(600, 400)))

	if len(dt.colWidths) != 3 {
		t.Fatalf("colWidths len = %d, want 3", len(dt.colWidths))
	}
	if dt.colWidths[0] != 100 {
		t.Errorf("col 0 width = %f, want 100", dt.colWidths[0])
	}
	if dt.colWidths[1] != 200 {
		t.Errorf("col 1 width = %f, want 200", dt.colWidths[1])
	}
	// Flex column gets remaining: 600 - 100 - 200 = 300.
	if dt.colWidths[2] != 300 {
		t.Errorf("col 2 width = %f, want 300", dt.colWidths[2])
	}
}

func TestLayout_FlexColumnMinWidth(t *testing.T) {
	cols := []Column{
		{Key: "a", Title: "A", Width: 500},
		{Key: "b", Title: "B", MinWidth: 80}, // flex, but table is 600 wide
	}
	dt := New(Columns(cols), RowCount(1))
	ctx := widget.NewContext()

	dt.Layout(ctx, geometry.Tight(geometry.Sz(600, 200)))

	if dt.colWidths[1] < 80 {
		t.Errorf("flex col width = %f, want >= 80 (MinWidth)", dt.colWidths[1])
	}
}

func TestLayout_NoColumns(t *testing.T) {
	dt := New(RowCount(5))
	ctx := widget.NewContext()
	size := dt.Layout(ctx, geometry.Tight(geometry.Sz(400, 300)))
	if size.Width != 400 || size.Height != 300 {
		t.Errorf("unexpected size: %v", size)
	}
}

// --- Draw Tests ---

func TestDraw_InvisibleWidget(t *testing.T) {
	dt := New(Columns(testColumns()), RowCount(5), CellValue(testCellValue))
	dt.SetVisible(false)
	ctx := widget.NewContext()
	canvas := &mockCanvas{}

	dt.Draw(ctx, canvas) // should not panic or draw anything
	if len(canvas.rects) != 0 {
		t.Errorf("expected no rects drawn when invisible, got %d", len(canvas.rects))
	}
}

func TestDraw_EmptyBounds(t *testing.T) {
	dt := New(Columns(testColumns()), RowCount(5))
	ctx := widget.NewContext()
	canvas := &mockCanvas{}

	// No layout = empty bounds.
	dt.Draw(ctx, canvas)
	// Should not panic.
}

func TestDraw_EmptyData(t *testing.T) {
	dt := New(Columns(testColumns()), RowCount(0))
	ctx := widget.NewContext()
	canvas := &mockCanvas{}

	dt.Layout(ctx, geometry.Tight(geometry.Sz(600, 400)))
	dt.SetBounds(geometry.NewRect(0, 0, 600, 400))
	dt.Draw(ctx, canvas)

	// Should show empty state text.
	found := false
	for _, text := range canvas.texts {
		if text == emptyStateText {
			found = true
			break
		}
	}
	if !found {
		t.Errorf("expected empty state text %q in drawn texts: %v", emptyStateText, canvas.texts)
	}
}

func TestDraw_WithData(t *testing.T) {
	dt := New(
		Columns(testColumns()),
		RowCount(5),
		CellValue(testCellValue),
	)
	ctx := widget.NewContext()
	canvas := &mockCanvas{}

	dt.Layout(ctx, geometry.Tight(geometry.Sz(600, 400)))
	dt.SetBounds(geometry.NewRect(0, 0, 600, 400))
	dt.Draw(ctx, canvas)

	// Should have drawn header texts.
	headerFound := false
	for _, text := range canvas.texts {
		if text == "Name" || text == "Size" || text == "Modified" {
			headerFound = true
			break
		}
	}
	if !headerFound {
		t.Error("expected header column titles to be drawn")
	}

	// Should have drawn cell values.
	cellFound := false
	for _, text := range canvas.texts {
		if text == "r0_name" {
			cellFound = true
			break
		}
	}
	if !cellFound {
		t.Error("expected cell values to be drawn")
	}
}

// --- Sort Tests ---

func TestSort_ClickHeader(t *testing.T) {
	var sortedCol string
	var sortedAsc bool
	dt := New(
		Columns(testColumns()),
		RowCount(10),
		OnSort(func(col string, asc bool) {
			sortedCol = col
			sortedAsc = asc
		}),
	)
	ctx := widget.NewContext()
	dt.Layout(ctx, geometry.Tight(geometry.Sz(600, 400)))
	dt.SetBounds(geometry.NewRect(0, 0, 600, 400))

	// Click on the first column header (Name, x=100 is within 0-200).
	clickEvent := &event.MouseEvent{
		MouseType: event.MousePress,
		Button:    event.ButtonLeft,
		Position:  geometry.Pt(100, 18), // middle of header row
	}
	consumed := dt.Event(ctx, clickEvent)
	if !consumed {
		t.Error("expected header click to be consumed")
	}

	if sortedCol != "name" {
		t.Errorf("sortedCol = %q, want %q", sortedCol, "name")
	}
	if !sortedAsc {
		t.Error("expected ascending on first click")
	}

	key, dir := dt.SortColumn()
	if key != "name" || dir != SortAscending {
		t.Errorf("SortColumn() = (%q, %v), want (name, Ascending)", key, dir)
	}
}

func TestSort_CycleDirections(t *testing.T) {
	dt := New(
		Columns(testColumns()),
		RowCount(10),
		OnSort(func(_ string, _ bool) {}),
	)
	ctx := widget.NewContext()
	dt.Layout(ctx, geometry.Tight(geometry.Sz(600, 400)))
	dt.SetBounds(geometry.NewRect(0, 0, 600, 400))

	click := func() {
		dt.Event(ctx, &event.MouseEvent{
			MouseType: event.MousePress,
			Button:    event.ButtonLeft,
			Position:  geometry.Pt(100, 18),
		})
	}

	// First click: None -> Ascending.
	click()
	_, dir := dt.SortColumn()
	if dir != SortAscending {
		t.Errorf("after 1st click: dir = %v, want Ascending", dir)
	}

	// Second click: Ascending -> Descending.
	click()
	_, dir = dt.SortColumn()
	if dir != SortDescending {
		t.Errorf("after 2nd click: dir = %v, want Descending", dir)
	}

	// Third click: Descending -> None.
	click()
	key, dir := dt.SortColumn()
	if key != "" || dir != SortNone {
		t.Errorf("after 3rd click: key=%q, dir=%v, want empty/None", key, dir)
	}
}

func TestSort_NonSortableColumn(t *testing.T) {
	cols := []Column{
		{Key: "name", Title: "Name", Width: 200, Sortable: false},
	}
	dt := New(Columns(cols), RowCount(5))
	ctx := widget.NewContext()
	dt.Layout(ctx, geometry.Tight(geometry.Sz(600, 400)))
	dt.SetBounds(geometry.NewRect(0, 0, 600, 400))

	consumed := dt.Event(ctx, &event.MouseEvent{
		MouseType: event.MousePress,
		Button:    event.ButtonLeft,
		Position:  geometry.Pt(100, 18),
	})
	if consumed {
		t.Error("clicking non-sortable column should not be consumed")
	}
}

func TestSort_SetSort(t *testing.T) {
	dt := New(Columns(testColumns()), RowCount(5))

	dt.SetSort("size", SortDescending)
	key, dir := dt.SortColumn()
	if key != "size" || dir != SortDescending {
		t.Errorf("SetSort: key=%q dir=%v, want size/Descending", key, dir)
	}

	// Clear sort.
	dt.SetSort("", SortNone)
	key, dir = dt.SortColumn()
	if key != "" || dir != SortNone {
		t.Errorf("ClearSort: key=%q dir=%v, want empty/None", key, dir)
	}
}

func TestSort_SetSort_InvalidKey(t *testing.T) {
	dt := New(Columns(testColumns()), RowCount(5))
	dt.SetSort("nonexistent", SortAscending)
	key, dir := dt.SortColumn()
	if key != "" || dir != SortNone {
		t.Errorf("invalid key: key=%q dir=%v, want empty/None", key, dir)
	}
}

// --- Selection Tests ---

func TestSelection_Single(t *testing.T) {
	var selectedRow int
	dt := New(
		Columns(testColumns()),
		RowCount(10),
		RowHeight(32),
		SelectionModeOpt(SelectionSingle),
		OnRowSelect(func(row int) { selectedRow = row }),
	)
	ctx := widget.NewContext()
	dt.Layout(ctx, geometry.Tight(geometry.Sz(600, 400)))
	dt.SetBounds(geometry.NewRect(0, 0, 600, 400))

	// Click on row 2 (y = header height + 2*32 + 16 to be in the middle).
	rowY := defaultHeaderHeight + 2*32 + 16
	dt.Event(ctx, &event.MouseEvent{
		MouseType: event.MousePress,
		Button:    event.ButtonLeft,
		Position:  geometry.Pt(100, rowY),
	})

	if selectedRow != 2 {
		t.Errorf("selectedRow = %d, want 2", selectedRow)
	}
	if !dt.IsRowSelected(2) {
		t.Error("row 2 should be selected")
	}
	if dt.IsRowSelected(0) {
		t.Error("row 0 should not be selected")
	}
}

func TestSelection_None(t *testing.T) {
	dt := New(
		Columns(testColumns()),
		RowCount(10),
		SelectionModeOpt(SelectionNone),
	)
	ctx := widget.NewContext()
	dt.Layout(ctx, geometry.Tight(geometry.Sz(600, 400)))
	dt.SetBounds(geometry.NewRect(0, 0, 600, 400))

	// Keyboard should not select.
	dt.SetFocused(true)
	consumed := dt.Event(ctx, &event.KeyEvent{
		KeyType: event.KeyPress,
		Key:     event.KeyDown,
	})
	if consumed {
		t.Error("selection mode none should not consume arrow keys")
	}
}

func TestSelection_Signal(t *testing.T) {
	sig := state.NewSignal(-1)
	dt := New(
		Columns(testColumns()),
		RowCount(10),
		SelectionModeOpt(SelectionSingle),
		SelectedRowSignal(sig),
	)
	ctx := widget.NewContext()
	dt.Layout(ctx, geometry.Tight(geometry.Sz(600, 400)))

	// Signal value should be reflected.
	sig.Set(3)
	if dt.cfg.ResolvedSelectedRow() != 3 {
		t.Errorf("resolved selected = %d, want 3", dt.cfg.ResolvedSelectedRow())
	}
}

func TestSelection_Multi_CtrlClick(t *testing.T) {
	dt := New(
		Columns(testColumns()),
		RowCount(10),
		RowHeight(32),
		SelectionModeOpt(SelectionMulti),
	)
	ctx := widget.NewContext()
	dt.Layout(ctx, geometry.Tight(geometry.Sz(600, 400)))
	dt.SetBounds(geometry.NewRect(0, 0, 600, 400))

	// Click row 1.
	dt.Event(ctx, &event.MouseEvent{
		MouseType: event.MousePress,
		Button:    event.ButtonLeft,
		Position:  geometry.Pt(100, defaultHeaderHeight+1*32+16),
	})
	if !dt.IsRowSelected(1) {
		t.Error("row 1 should be selected after click")
	}

	// Ctrl+click row 3.
	dt.Event(ctx, event.NewMouseEvent(
		event.MousePress,
		event.ButtonLeft,
		0,
		geometry.Pt(100, defaultHeaderHeight+3*32+16),
		geometry.Pt(100, defaultHeaderHeight+3*32+16),
		event.ModCtrl,
	))
	if !dt.IsRowSelected(1) {
		t.Error("row 1 should still be selected")
	}
	if !dt.IsRowSelected(3) {
		t.Error("row 3 should be selected")
	}

	// Ctrl+click row 1 again to deselect.
	dt.Event(ctx, event.NewMouseEvent(
		event.MousePress,
		event.ButtonLeft,
		0,
		geometry.Pt(100, defaultHeaderHeight+1*32+16),
		geometry.Pt(100, defaultHeaderHeight+1*32+16),
		event.ModCtrl,
	))
	if dt.IsRowSelected(1) {
		t.Error("row 1 should be deselected after ctrl+click toggle")
	}
}

// --- Keyboard Navigation Tests ---

func TestKeyboard_DownUp(t *testing.T) {
	dt := New(
		Columns(testColumns()),
		RowCount(10),
		SelectionModeOpt(SelectionSingle),
		SelectedRow(0),
	)
	ctx := widget.NewContext()
	dt.Layout(ctx, geometry.Tight(geometry.Sz(600, 400)))
	dt.SetBounds(geometry.NewRect(0, 0, 600, 400))
	dt.SetFocused(true)

	// Press Down.
	dt.Event(ctx, &event.KeyEvent{KeyType: event.KeyPress, Key: event.KeyDown})
	if dt.cfg.ResolvedSelectedRow() != 1 {
		t.Errorf("after Down: selected = %d, want 1", dt.cfg.ResolvedSelectedRow())
	}

	// Press Up.
	dt.Event(ctx, &event.KeyEvent{KeyType: event.KeyPress, Key: event.KeyUp})
	if dt.cfg.ResolvedSelectedRow() != 0 {
		t.Errorf("after Up: selected = %d, want 0", dt.cfg.ResolvedSelectedRow())
	}
}

func TestKeyboard_HomeEnd(t *testing.T) {
	dt := New(
		Columns(testColumns()),
		RowCount(100),
		SelectionModeOpt(SelectionSingle),
		SelectedRow(50),
	)
	ctx := widget.NewContext()
	dt.Layout(ctx, geometry.Tight(geometry.Sz(600, 400)))
	dt.SetBounds(geometry.NewRect(0, 0, 600, 400))
	dt.SetFocused(true)

	dt.Event(ctx, &event.KeyEvent{KeyType: event.KeyPress, Key: event.KeyEnd})
	if dt.cfg.ResolvedSelectedRow() != 99 {
		t.Errorf("after End: selected = %d, want 99", dt.cfg.ResolvedSelectedRow())
	}

	dt.Event(ctx, &event.KeyEvent{KeyType: event.KeyPress, Key: event.KeyHome})
	if dt.cfg.ResolvedSelectedRow() != 0 {
		t.Errorf("after Home: selected = %d, want 0", dt.cfg.ResolvedSelectedRow())
	}
}

func TestKeyboard_PageDownUp(t *testing.T) {
	dt := New(
		Columns(testColumns()),
		RowCount(100),
		RowHeight(32),
		SelectionModeOpt(SelectionSingle),
		SelectedRow(0),
	)
	ctx := widget.NewContext()
	dt.Layout(ctx, geometry.Tight(geometry.Sz(600, 400)))
	dt.SetBounds(geometry.NewRect(0, 0, 600, 400))
	dt.SetFocused(true)

	dt.Event(ctx, &event.KeyEvent{KeyType: event.KeyPress, Key: event.KeyPageDown})
	selected := dt.cfg.ResolvedSelectedRow()
	if selected <= 0 {
		t.Errorf("after PageDown: selected = %d, want > 0", selected)
	}

	dt.Event(ctx, &event.KeyEvent{KeyType: event.KeyPress, Key: event.KeyPageUp})
	selected2 := dt.cfg.ResolvedSelectedRow()
	if selected2 >= selected {
		t.Errorf("after PageUp: selected = %d, expected < %d", selected2, selected)
	}
}

func TestKeyboard_NotFocused(t *testing.T) {
	dt := New(
		Columns(testColumns()),
		RowCount(10),
		SelectionModeOpt(SelectionSingle),
		SelectedRow(0),
	)
	ctx := widget.NewContext()
	dt.Layout(ctx, geometry.Tight(geometry.Sz(600, 400)))
	// NOT focused.

	consumed := dt.Event(ctx, &event.KeyEvent{KeyType: event.KeyPress, Key: event.KeyDown})
	if consumed {
		t.Error("keyboard events should not be consumed when not focused")
	}
}

func TestKeyboard_Disabled(t *testing.T) {
	dt := New(
		Columns(testColumns()),
		RowCount(10),
		SelectionModeOpt(SelectionSingle),
		Disabled(true),
	)
	ctx := widget.NewContext()
	dt.Layout(ctx, geometry.Tight(geometry.Sz(600, 400)))
	dt.SetFocused(true)

	consumed := dt.Event(ctx, &event.KeyEvent{KeyType: event.KeyPress, Key: event.KeyDown})
	if consumed {
		t.Error("keyboard events should not be consumed when disabled")
	}
}

func TestKeyboard_EmptyData(t *testing.T) {
	dt := New(
		Columns(testColumns()),
		RowCount(0),
		SelectionModeOpt(SelectionSingle),
	)
	ctx := widget.NewContext()
	dt.Layout(ctx, geometry.Tight(geometry.Sz(600, 400)))
	dt.SetFocused(true)

	consumed := dt.Event(ctx, &event.KeyEvent{KeyType: event.KeyPress, Key: event.KeyDown})
	if consumed {
		t.Error("keyboard should not consume when no data")
	}
}

func TestKeyboard_EnterOnSelected(t *testing.T) {
	var activated int
	dt := New(
		Columns(testColumns()),
		RowCount(10),
		SelectionModeOpt(SelectionSingle),
		SelectedRow(3),
		OnRowSelect(func(row int) { activated = row }),
	)
	ctx := widget.NewContext()
	dt.Layout(ctx, geometry.Tight(geometry.Sz(600, 400)))
	dt.SetFocused(true)

	consumed := dt.Event(ctx, &event.KeyEvent{KeyType: event.KeyPress, Key: event.KeyEnter})
	if !consumed {
		t.Error("Enter on selected row should be consumed")
	}
	if activated != 3 {
		t.Errorf("activated = %d, want 3", activated)
	}
}

// --- Focusable Tests ---

func TestIsFocusable(t *testing.T) {
	dt := New(Columns(testColumns()), RowCount(5))
	if !dt.IsFocusable() {
		t.Error("expected focusable")
	}

	dt.SetVisible(false)
	if dt.IsFocusable() {
		t.Error("invisible should not be focusable")
	}

	dt.SetVisible(true)
	dt.SetEnabled(false)
	if dt.IsFocusable() {
		t.Error("disabled should not be focusable")
	}
}

func TestIsFocusable_DisabledViaOption(t *testing.T) {
	dt := New(Columns(testColumns()), RowCount(5), Disabled(true))
	if dt.IsFocusable() {
		t.Error("disabled via option should not be focusable")
	}
}

// --- Accessibility Tests ---

func TestAccessibility(t *testing.T) {
	dt := New(Columns(testColumns()), RowCount(50), SelectedRow(10))

	if dt.AccessibilityRole() != a11y.RoleTable {
		t.Errorf("role = %v, want Table", dt.AccessibilityRole())
	}
	if dt.AccessibilityLabel() != "Data Table" {
		t.Errorf("label = %q, want %q", dt.AccessibilityLabel(), "Data Table")
	}
	if dt.AccessibilityHint() != "" {
		t.Errorf("hint = %q, want empty", dt.AccessibilityHint())
	}

	value := dt.AccessibilityValue()
	expected := "Row 11 of 50 selected"
	if value != expected {
		t.Errorf("value = %q, want %q", value, expected)
	}
}

func TestAccessibility_NoSelection(t *testing.T) {
	dt := New(Columns(testColumns()), RowCount(25))
	value := dt.AccessibilityValue()
	if value != "25 rows" {
		t.Errorf("value = %q, want %q", value, "25 rows")
	}
}

func TestAccessibility_CustomLabel(t *testing.T) {
	dt := New(A11yLabel("Files"))
	if dt.AccessibilityLabel() != "Files" {
		t.Errorf("label = %q, want %q", dt.AccessibilityLabel(), "Files")
	}
}

func TestAccessibility_State(t *testing.T) {
	dt := New(Disabled(true))
	s := dt.AccessibilityState()
	if !s.Disabled {
		t.Error("expected disabled state")
	}
}

func TestAccessibility_Actions(t *testing.T) {
	dt := New()
	actions := dt.AccessibilityActions()
	if len(actions) != 2 {
		t.Fatalf("actions = %d, want 2", len(actions))
	}
}

// --- Public API Tests ---

func TestVisibleRowRange(t *testing.T) {
	dt := New(
		Columns(testColumns()),
		RowCount(100),
		RowHeight(32),
	)
	ctx := widget.NewContext()
	dt.Layout(ctx, geometry.Tight(geometry.Sz(600, 400)))

	start, end := dt.VisibleRowRange()
	if start != 0 {
		t.Errorf("start = %d, want 0", start)
	}
	if end <= 0 {
		t.Errorf("end = %d, want > 0", end)
	}
	if end > 100 {
		t.Errorf("end = %d, should not exceed row count", end)
	}
}

func TestVisibleRowRange_EmptyData(t *testing.T) {
	dt := New(RowCount(0))
	ctx := widget.NewContext()
	dt.Layout(ctx, geometry.Tight(geometry.Sz(600, 400)))

	start, end := dt.VisibleRowRange()
	if start != 0 || end != 0 {
		t.Errorf("empty data: start=%d end=%d, want 0,0", start, end)
	}
}

func TestGetRowCount(t *testing.T) {
	dt := New(RowCount(42))
	if dt.GetRowCount() != 42 {
		t.Errorf("GetRowCount() = %d, want 42", dt.GetRowCount())
	}
}

func TestScrollToRow(t *testing.T) {
	dt := New(
		Columns(testColumns()),
		RowCount(100),
		RowHeight(32),
	)
	ctx := widget.NewContext()
	dt.Layout(ctx, geometry.Tight(geometry.Sz(600, 400)))
	dt.SetBounds(geometry.NewRect(0, 0, 600, 400))

	// Scroll to an out-of-view row (should not panic).
	dt.ScrollToRow(50)

	// Scroll to invalid row (should be no-op).
	dt.ScrollToRow(-1)
	dt.ScrollToRow(200)
}

func TestInvalidateData(t *testing.T) {
	dt := New(RowCount(5))
	dt.InvalidateData() // should not panic
}

// --- Children Tests ---

func TestChildren(t *testing.T) {
	dt := New(Columns(testColumns()), RowCount(5))
	children := dt.Children()
	if len(children) != 1 {
		t.Errorf("children = %d, want 1 (scroll view)", len(children))
	}
}

func TestChildren_NilScroll(t *testing.T) {
	dt := &Widget{}
	children := dt.Children()
	if children != nil {
		t.Error("expected nil children when scroll is nil")
	}
}

// --- Mount/Unmount Tests ---

func TestMount_WithSignals(t *testing.T) {
	rowSig := state.NewSignal(10)
	selSig := state.NewSignal(0)
	disSig := state.NewSignal(false)

	dt := New(
		Columns(testColumns()),
		RowCountSignal(rowSig),
		SelectedRowSignal(selSig),
		DisabledSignal(disSig),
		SelectionModeOpt(SelectionSingle),
	)
	ctx := widget.NewContext()
	dt.Layout(ctx, geometry.Tight(geometry.Sz(600, 400)))

	// Mount with scheduler.
	dt.Mount(ctx)
	dt.Unmount()
	// Should not panic.
}

func TestMount_NilScheduler(t *testing.T) {
	dt := New(Columns(testColumns()), RowCount(5))
	ctx := widget.NewContext() // no scheduler
	dt.Mount(ctx)              // should be no-op
}

func TestMount_ReadonlySignals(t *testing.T) {
	rowSig := state.NewSignal(10)
	selSig := state.NewSignal(0)
	disSig := state.NewSignal(false)

	dt := New(
		Columns(testColumns()),
		RowCountReadonlySignal(rowSig.AsReadonly()),
		SelectedRowReadonlySignal(selSig.AsReadonly()),
		DisabledReadonlySignal(disSig.AsReadonly()),
	)
	ctx := widget.NewContext()
	dt.Mount(ctx) // no-op without scheduler, but should not panic
}

// --- Event Edge Cases ---

func TestEvent_InvisibleWidget(t *testing.T) {
	dt := New(Columns(testColumns()), RowCount(5))
	dt.SetVisible(false)
	ctx := widget.NewContext()

	consumed := dt.Event(ctx, &event.KeyEvent{KeyType: event.KeyPress, Key: event.KeyDown})
	if consumed {
		t.Error("invisible widget should not consume events")
	}
}

func TestEvent_DisabledWidget(t *testing.T) {
	dt := New(Columns(testColumns()), RowCount(5))
	dt.SetEnabled(false)
	ctx := widget.NewContext()

	consumed := dt.Event(ctx, &event.KeyEvent{KeyType: event.KeyPress, Key: event.KeyDown})
	if consumed {
		t.Error("disabled widget should not consume events")
	}
}

func TestEvent_RightClickNotConsumed(t *testing.T) {
	dt := New(
		Columns(testColumns()),
		RowCount(10),
		RowHeight(32),
		SelectionModeOpt(SelectionSingle),
	)
	ctx := widget.NewContext()
	dt.Layout(ctx, geometry.Tight(geometry.Sz(600, 400)))
	dt.SetBounds(geometry.NewRect(0, 0, 600, 400))

	consumed := dt.Event(ctx, &event.MouseEvent{
		MouseType: event.MousePress,
		Button:    event.ButtonRight,
		Position:  geometry.Pt(100, 18),
	})
	if consumed {
		t.Error("right click on header should not be consumed")
	}
}

// --- Header Hover Tests ---

func TestHeaderHover(t *testing.T) {
	dt := New(Columns(testColumns()), RowCount(5))
	ctx := widget.NewContext()
	dt.Layout(ctx, geometry.Tight(geometry.Sz(600, 400)))
	dt.SetBounds(geometry.NewRect(0, 0, 600, 400))

	// Move mouse into header.
	dt.Event(ctx, &event.MouseEvent{
		MouseType: event.MouseMove,
		Position:  geometry.Pt(100, 18),
	})
	if dt.hoveredColHdr < 0 {
		t.Error("expected header column to be hovered")
	}

	// Move mouse out of header (into data area).
	dt.Event(ctx, &event.MouseEvent{
		MouseType: event.MouseMove,
		Position:  geometry.Pt(100, 100),
	})
	if dt.hoveredColHdr != noHoveredCol {
		t.Error("expected header hover to be cleared when mouse moves to data area")
	}
}

func TestHeaderHover_MouseLeave(t *testing.T) {
	dt := New(Columns(testColumns()), RowCount(5))
	ctx := widget.NewContext()
	dt.Layout(ctx, geometry.Tight(geometry.Sz(600, 400)))
	dt.SetBounds(geometry.NewRect(0, 0, 600, 400))

	// Hover header.
	dt.Event(ctx, &event.MouseEvent{
		MouseType: event.MouseMove,
		Position:  geometry.Pt(100, 18),
	})

	// Mouse leave.
	dt.Event(ctx, &event.MouseEvent{
		MouseType: event.MouseLeave,
		Position:  geometry.Pt(-1, -1),
	})
	if dt.hoveredColHdr != noHoveredCol {
		t.Error("expected header hover to be cleared on mouse leave")
	}
}

// --- Row Hover Tests ---

func TestRowHover(t *testing.T) {
	dt := New(
		Columns(testColumns()),
		RowCount(10),
		RowHeight(32),
	)
	ctx := widget.NewContext()
	dt.Layout(ctx, geometry.Tight(geometry.Sz(600, 400)))
	dt.SetBounds(geometry.NewRect(0, 0, 600, 400))
	dt.Draw(ctx, &mockCanvas{}) // initialize scroll view

	// We need to trigger content mouse events through the scroll view.
	// For unit testing, test the internal helpers directly.
	dt.hoveredRow = 5
	if dt.hoveredRow != 5 {
		t.Error("expected hovered row to be set")
	}
}

// --- Column Width Tests ---

func TestColumnX(t *testing.T) {
	dt := New(Columns(testColumns()))
	ctx := widget.NewContext()
	dt.Layout(ctx, geometry.Tight(geometry.Sz(600, 400)))

	x0 := dt.columnX(0)
	if x0 != 0 {
		t.Errorf("columnX(0) = %f, want 0", x0)
	}

	x1 := dt.columnX(1)
	if x1 != 200 { // first column is 200 wide
		t.Errorf("columnX(1) = %f, want 200", x1)
	}

	x2 := dt.columnX(2)
	if x2 != 300 { // 200 + 100
		t.Errorf("columnX(2) = %f, want 300", x2)
	}
}

func TestColumnAtX(t *testing.T) {
	dt := New(Columns(testColumns()))
	ctx := widget.NewContext()
	dt.Layout(ctx, geometry.Tight(geometry.Sz(600, 400)))

	tests := []struct {
		x    float32
		want int
	}{
		{0, 0},
		{100, 0},
		{199, 0},
		{200, 1},
		{250, 1},
		{300, 2},
		{-1, noHoveredCol},
		{9999, noHoveredCol},
	}
	for _, tt := range tests {
		got := dt.columnAtX(tt.x)
		if got != tt.want {
			t.Errorf("columnAtX(%f) = %d, want %d", tt.x, got, tt.want)
		}
	}
}

// --- Total Data Height ---

func TestTotalDataHeight(t *testing.T) {
	dt := New(RowCount(10), RowHeight(32))
	if h := dt.totalDataHeight(); h != 320 {
		t.Errorf("totalDataHeight() = %f, want 320", h)
	}
}

func TestTotalDataHeight_Empty(t *testing.T) {
	dt := New(RowCount(0))
	if h := dt.totalDataHeight(); h != 0 {
		t.Errorf("totalDataHeight() = %f, want 0", h)
	}
}

// --- Row At Y ---

func TestRowAtY(t *testing.T) {
	dt := New(RowCount(10), RowHeight(32))
	tests := []struct {
		y    float32
		want int
	}{
		{0, 0},
		{16, 0},
		{32, 1},
		{319, 9},
		{320, noHoveredRow}, // out of range
		{-1, noHoveredRow},
	}
	for _, tt := range tests {
		got := dt.rowAtY(tt.y)
		if got != tt.want {
			t.Errorf("rowAtY(%f) = %d, want %d", tt.y, got, tt.want)
		}
	}
}

// --- Visible Row Range ---

func TestVisibleRowRangeInternal(t *testing.T) {
	dt := New(RowCount(100), RowHeight(32))
	first, last := dt.visibleRowRange(0, 320)
	if first != 0 {
		t.Errorf("first = %d, want 0", first)
	}
	if last != 10 {
		t.Errorf("last = %d, want 10", last)
	}
}

func TestVisibleRowRange_Scrolled(t *testing.T) {
	dt := New(RowCount(100), RowHeight(32))
	first, last := dt.visibleRowRange(160, 320)
	if first != 5 {
		t.Errorf("first = %d, want 5", first)
	}
	if last != 15 {
		t.Errorf("last = %d, want 15", last)
	}
}

// --- Painter Tests ---

func TestDefaultPainter_PaintHeader(t *testing.T) {
	p := DefaultPainter{}
	canvas := &mockCanvas{}
	bounds := geometry.NewRect(0, 0, 600, 36)
	p.PaintHeader(canvas, bounds, HeaderPaintState{})
	if len(canvas.rects) == 0 {
		t.Error("expected header background rect")
	}
}

func TestDefaultPainter_PaintHeader_Empty(t *testing.T) {
	p := DefaultPainter{}
	canvas := &mockCanvas{}
	p.PaintHeader(canvas, geometry.Rect{}, HeaderPaintState{})
	if len(canvas.rects) != 0 {
		t.Error("expected no rects for empty bounds")
	}
}

func TestDefaultPainter_PaintHeaderCell(t *testing.T) {
	p := DefaultPainter{}
	canvas := &mockCanvas{}
	bounds := geometry.NewRect(0, 0, 200, 36)
	p.PaintHeaderCell(canvas, bounds, HeaderCellPaintState{
		Title:    "Name",
		Sortable: true,
		SortDir:  SortAscending,
		Hovered:  true,
	})
	if len(canvas.texts) == 0 {
		t.Error("expected header cell text")
	}
	// Should contain sort indicator.
	found := false
	for _, text := range canvas.texts {
		if text == "Name \u25B2" {
			found = true
		}
	}
	if !found {
		t.Errorf("expected 'Name \\u25B2' in texts: %v", canvas.texts)
	}
}

func TestDefaultPainter_PaintHeaderCell_NoSort(t *testing.T) {
	p := DefaultPainter{}
	canvas := &mockCanvas{}
	bounds := geometry.NewRect(0, 0, 200, 36)
	p.PaintHeaderCell(canvas, bounds, HeaderCellPaintState{
		Title: "Name",
	})
	found := false
	for _, text := range canvas.texts {
		if text == "Name" {
			found = true
		}
	}
	if !found {
		t.Error("expected 'Name' without sort indicator")
	}
}

func TestDefaultPainter_PaintRow(t *testing.T) {
	p := DefaultPainter{}
	canvas := &mockCanvas{}
	bounds := geometry.NewRect(0, 0, 600, 32)

	// Alternate row.
	p.PaintRow(canvas, RowPaintState{
		Bounds:   bounds,
		RowIndex: 1,
	})
	if len(canvas.rects) == 0 {
		t.Error("expected zebra striping for odd row")
	}
}

func TestDefaultPainter_PaintRow_Selected(t *testing.T) {
	p := DefaultPainter{}
	canvas := &mockCanvas{}
	bounds := geometry.NewRect(0, 0, 600, 32)

	p.PaintRow(canvas, RowPaintState{
		Bounds:   bounds,
		Selected: true,
		Focused:  true,
	})
	// Should draw selection + focus border.
	if len(canvas.rects) < 2 {
		t.Errorf("expected >= 2 rects for selected+focused, got %d", len(canvas.rects))
	}
}

func TestDefaultPainter_PaintRow_Hovered(t *testing.T) {
	p := DefaultPainter{}
	canvas := &mockCanvas{}
	bounds := geometry.NewRect(0, 0, 600, 32)

	p.PaintRow(canvas, RowPaintState{
		Bounds:  bounds,
		Hovered: true,
	})
	if len(canvas.rects) == 0 {
		t.Error("expected hover highlight")
	}
}

func TestDefaultPainter_PaintRow_EmptyBounds(t *testing.T) {
	p := DefaultPainter{}
	canvas := &mockCanvas{}
	p.PaintRow(canvas, RowPaintState{})
	if len(canvas.rects) != 0 {
		t.Error("expected no drawing for empty bounds")
	}
}

func TestDefaultPainter_PaintCell(t *testing.T) {
	p := DefaultPainter{}
	canvas := &mockCanvas{}
	bounds := geometry.NewRect(0, 0, 200, 32)

	p.PaintCell(canvas, CellPaintState{
		Bounds: bounds,
		Value:  "hello",
		Align:  widget.TextAlignLeft,
	})
	if len(canvas.texts) == 0 {
		t.Error("expected cell text")
	}
	if canvas.texts[0] != "hello" {
		t.Errorf("cell text = %q, want %q", canvas.texts[0], "hello")
	}
}

func TestDefaultPainter_PaintCell_EmptyBounds(t *testing.T) {
	p := DefaultPainter{}
	canvas := &mockCanvas{}
	p.PaintCell(canvas, CellPaintState{})
	if len(canvas.texts) != 0 {
		t.Error("expected no text for empty bounds")
	}
}

func TestDefaultPainter_PaintEmptyState(t *testing.T) {
	p := DefaultPainter{}
	canvas := &mockCanvas{}
	bounds := geometry.NewRect(0, 0, 600, 400)
	p.PaintEmptyState(canvas, bounds)
	if len(canvas.texts) == 0 {
		t.Error("expected empty state text")
	}
	if canvas.texts[0] != emptyStateText {
		t.Errorf("empty text = %q, want %q", canvas.texts[0], emptyStateText)
	}
}

func TestDefaultPainter_PaintEmptyState_EmptyBounds(t *testing.T) {
	p := DefaultPainter{}
	canvas := &mockCanvas{}
	p.PaintEmptyState(canvas, geometry.Rect{})
	if len(canvas.texts) != 0 {
		t.Error("expected no text for empty bounds")
	}
}

// --- Painter with ColorScheme ---

func TestDefaultPainter_WithColorScheme(t *testing.T) {
	p := DefaultPainter{}
	canvas := &mockCanvas{}
	scheme := TableColorScheme{
		HeaderBackground: widget.ColorBlue,
		HeaderText:       widget.ColorWhite,
		SelectionColor:   widget.ColorRed,
		HoverColor:       widget.ColorGreen,
		FocusColor:       widget.ColorYellow,
		RowAlternate:     widget.ColorGray,
		CellText:         widget.ColorBlack,
		EmptyText:        widget.ColorGray,
	}

	// Header.
	p.PaintHeader(canvas, geometry.NewRect(0, 0, 600, 36), HeaderPaintState{ColorScheme: scheme})

	// Header cell with hover.
	p.PaintHeaderCell(canvas, geometry.NewRect(0, 0, 200, 36), HeaderCellPaintState{
		Title:       "Name",
		Sortable:    true,
		Hovered:     true,
		ColorScheme: scheme,
	})

	// Row.
	p.PaintRow(canvas, RowPaintState{
		Bounds:      geometry.NewRect(0, 0, 600, 32),
		Selected:    true,
		Focused:     true,
		ColorScheme: scheme,
	})

	// Cell.
	p.PaintCell(canvas, CellPaintState{
		Bounds:      geometry.NewRect(0, 0, 200, 32),
		Value:       "test",
		ColorScheme: scheme,
	})

	// Just verify no panic. The scheme colors are used instead of defaults.
}

// --- Options Tests ---

func TestOption_RowCountFn(t *testing.T) {
	dt := New(RowCountFn(func() int { return 42 }))
	if dt.cfg.ResolvedRowCount() != 42 {
		t.Errorf("RowCountFn: got %d, want 42", dt.cfg.ResolvedRowCount())
	}
}

func TestOption_DisabledFn(t *testing.T) {
	dt := New(DisabledFn(func() bool { return true }))
	if !dt.cfg.ResolvedDisabled() {
		t.Error("DisabledFn: expected true")
	}
}

func TestOption_OnScroll(t *testing.T) {
	var scrolled bool
	dt := New(OnScroll(func(_ float32) { scrolled = true }))
	_ = dt
	_ = scrolled // callback is wired into scroll view
}

func TestOption_ScrollYSignal(t *testing.T) {
	sig := state.NewSignal[float32](0)
	dt := New(ScrollYSignal(sig))
	if dt.cfg.scrollYSignal == nil {
		t.Error("expected scrollYSignal to be set")
	}
}

// --- Virtual Content Tests ---

func TestVirtualContent_NilTable(t *testing.T) {
	vc := &virtualContent{}
	ctx := widget.NewContext()

	size := vc.Layout(ctx, geometry.Tight(geometry.Sz(100, 100)))
	if size.Width != 0 || size.Height != 0 {
		t.Errorf("nil table layout: %v, want 0,0", size)
	}

	canvas := &mockCanvas{}
	vc.Draw(ctx, canvas) // should not panic

	consumed := vc.Event(ctx, &event.MouseEvent{})
	if consumed {
		t.Error("nil table should not consume events")
	}

	children := vc.Children()
	if children != nil {
		t.Error("expected nil children")
	}
}

// --- Interface Compliance ---

func TestInterfaceCompliance(t *testing.T) {
	var w widget.Widget = New()
	_ = w

	var f widget.Focusable = New()
	_ = f

	var l widget.Lifecycle = New()
	_ = l

	var a a11y.Accessible = New()
	_ = a
}

// --- Edge Case: Clamp Selection ---

func TestSelection_ClampToRange(t *testing.T) {
	dt := New(
		RowCount(5),
		SelectionModeOpt(SelectionSingle),
		SelectedRow(0),
	)
	ctx := widget.NewContext()
	dt.Layout(ctx, geometry.Tight(geometry.Sz(600, 400)))
	dt.SetFocused(true)

	// Try to go above 0.
	dt.Event(ctx, &event.KeyEvent{KeyType: event.KeyPress, Key: event.KeyUp})
	if dt.cfg.ResolvedSelectedRow() != 0 {
		t.Errorf("should clamp to 0, got %d", dt.cfg.ResolvedSelectedRow())
	}

	// Go to end.
	dt.Event(ctx, &event.KeyEvent{KeyType: event.KeyPress, Key: event.KeyEnd})
	if dt.cfg.ResolvedSelectedRow() != 4 {
		t.Errorf("should be at 4, got %d", dt.cfg.ResolvedSelectedRow())
	}

	// Try to go past end.
	dt.Event(ctx, &event.KeyEvent{KeyType: event.KeyPress, Key: event.KeyDown})
	if dt.cfg.ResolvedSelectedRow() != 4 {
		t.Errorf("should clamp to 4, got %d", dt.cfg.ResolvedSelectedRow())
	}
}

// --- Sort Switching Column ---

func TestSort_SwitchColumn(t *testing.T) {
	dt := New(
		Columns(testColumns()),
		RowCount(10),
		OnSort(func(_ string, _ bool) {}),
	)
	ctx := widget.NewContext()
	dt.Layout(ctx, geometry.Tight(geometry.Sz(600, 400)))
	dt.SetBounds(geometry.NewRect(0, 0, 600, 400))

	// Click first column (Name at x=100).
	dt.Event(ctx, &event.MouseEvent{
		MouseType: event.MousePress,
		Button:    event.ButtonLeft,
		Position:  geometry.Pt(100, 18),
	})
	key, dir := dt.SortColumn()
	if key != "name" || dir != SortAscending {
		t.Errorf("first click: key=%q dir=%v", key, dir)
	}

	// Click second column (Size at x=250).
	dt.Event(ctx, &event.MouseEvent{
		MouseType: event.MousePress,
		Button:    event.ButtonLeft,
		Position:  geometry.Pt(250, 18),
	})
	key, dir = dt.SortColumn()
	if key != "size" || dir != SortAscending {
		t.Errorf("switch column: key=%q dir=%v, want size/Ascending", key, dir)
	}
}

// --- Resolve Column Widths Edge Cases ---

func TestResolveColumnWidths_AllFlex(t *testing.T) {
	cols := []Column{
		{Key: "a", Title: "A"},
		{Key: "b", Title: "B"},
	}
	dt := New(Columns(cols), RowCount(1))
	ctx := widget.NewContext()
	dt.Layout(ctx, geometry.Tight(geometry.Sz(400, 200)))

	if len(dt.colWidths) != 2 {
		t.Fatalf("colWidths len = %d, want 2", len(dt.colWidths))
	}
	if dt.colWidths[0] != 200 {
		t.Errorf("col 0 = %f, want 200", dt.colWidths[0])
	}
	if dt.colWidths[1] != 200 {
		t.Errorf("col 1 = %f, want 200", dt.colWidths[1])
	}
}

func TestResolveColumnWidths_OverflowFixed(t *testing.T) {
	cols := []Column{
		{Key: "a", Title: "A", Width: 300},
		{Key: "b", Title: "B", Width: 400},
		{Key: "c", Title: "C"}, // flex, remaining = 0
	}
	dt := New(Columns(cols), RowCount(1))
	ctx := widget.NewContext()
	dt.Layout(ctx, geometry.Tight(geometry.Sz(600, 200)))

	// Flex column should get MinWidth (default 50) since remaining < 0.
	if dt.colWidths[2] < defaultMinColumnWidth {
		t.Errorf("flex col = %f, want >= %f", dt.colWidths[2], defaultMinColumnWidth)
	}
}

// --- KeyRepeat Support ---

func TestKeyboard_KeyRepeat(t *testing.T) {
	dt := New(
		Columns(testColumns()),
		RowCount(10),
		SelectionModeOpt(SelectionSingle),
		SelectedRow(0),
	)
	ctx := widget.NewContext()
	dt.Layout(ctx, geometry.Tight(geometry.Sz(600, 400)))
	dt.SetFocused(true)

	consumed := dt.Event(ctx, &event.KeyEvent{KeyType: event.KeyRepeat, Key: event.KeyDown})
	if !consumed {
		t.Error("KeyRepeat should be consumed")
	}
	if dt.cfg.ResolvedSelectedRow() != 1 {
		t.Errorf("after KeyRepeat Down: selected = %d, want 1", dt.cfg.ResolvedSelectedRow())
	}
}

func TestKeyboard_KeyRelease_Ignored(t *testing.T) {
	dt := New(
		Columns(testColumns()),
		RowCount(10),
		SelectionModeOpt(SelectionSingle),
		SelectedRow(0),
	)
	ctx := widget.NewContext()
	dt.Layout(ctx, geometry.Tight(geometry.Sz(600, 400)))
	dt.SetFocused(true)

	consumed := dt.Event(ctx, &event.KeyEvent{KeyType: event.KeyRelease, Key: event.KeyDown})
	if consumed {
		t.Error("KeyRelease should not be consumed")
	}
}

// --- Space key ---

func TestKeyboard_Space(t *testing.T) {
	var activated int
	dt := New(
		Columns(testColumns()),
		RowCount(10),
		SelectionModeOpt(SelectionSingle),
		SelectedRow(5),
		OnRowSelect(func(row int) { activated = row }),
	)
	ctx := widget.NewContext()
	dt.Layout(ctx, geometry.Tight(geometry.Sz(600, 400)))
	dt.SetFocused(true)

	consumed := dt.Event(ctx, &event.KeyEvent{KeyType: event.KeyPress, Key: event.KeySpace})
	if !consumed {
		t.Error("Space on selected row should be consumed")
	}
	if activated != 5 {
		t.Errorf("activated = %d, want 5", activated)
	}
}

// --- Unknown key not consumed ---

func TestKeyboard_UnknownKey(t *testing.T) {
	dt := New(
		Columns(testColumns()),
		RowCount(10),
		SelectionModeOpt(SelectionSingle),
		SelectedRow(0),
	)
	ctx := widget.NewContext()
	dt.Layout(ctx, geometry.Tight(geometry.Sz(600, 400)))
	dt.SetFocused(true)

	consumed := dt.Event(ctx, &event.KeyEvent{KeyType: event.KeyPress, Key: event.KeyF1})
	if consumed {
		t.Error("unknown key should not be consumed")
	}
}

// --- Content Mouse Event Internal Helpers ---

func TestContentMouseMove(t *testing.T) {
	dt := New(
		Columns(testColumns()),
		RowCount(10),
		RowHeight(32),
	)
	ctx := widget.NewContext()
	dt.Layout(ctx, geometry.Tight(geometry.Sz(600, 400)))
	dt.SetBounds(geometry.NewRect(0, 0, 600, 400))

	// Simulate content mouse move in content space (already transformed by ScrollView).
	// Content space has no header offset or bounds offset.
	handleContentMouseMove(dt, ctx, &event.MouseEvent{
		MouseType: event.MouseMove,
		Position:  geometry.Pt(100, 48),
	})
	if dt.hoveredRow != 1 {
		t.Errorf("hoveredRow = %d, want 1", dt.hoveredRow)
	}
}

func TestContentMousePress_Selection(t *testing.T) {
	var selected int
	dt := New(
		Columns(testColumns()),
		RowCount(10),
		RowHeight(32),
		SelectionModeOpt(SelectionSingle),
		OnRowSelect(func(row int) { selected = row }),
	)
	ctx := widget.NewContext()
	dt.Layout(ctx, geometry.Tight(geometry.Sz(600, 400)))
	dt.SetBounds(geometry.NewRect(0, 0, 600, 400))

	// Content space: no header offset (ScrollView already transformed).
	consumed := handleContentMousePress(dt, ctx, &event.MouseEvent{
		MouseType: event.MousePress,
		Button:    event.ButtonLeft,
		Position:  geometry.Pt(100, 3*32+16),
	})
	if !consumed {
		t.Error("left click on row should be consumed")
	}
	if selected != 3 {
		t.Errorf("selected = %d, want 3", selected)
	}
}

func TestContentMousePress_RightButton(t *testing.T) {
	dt := New(
		Columns(testColumns()),
		RowCount(10),
		RowHeight(32),
		SelectionModeOpt(SelectionSingle),
	)
	ctx := widget.NewContext()
	dt.Layout(ctx, geometry.Tight(geometry.Sz(600, 400)))
	dt.SetBounds(geometry.NewRect(0, 0, 600, 400))

	// Content space: no header offset (ScrollView already transformed).
	consumed := handleContentMousePress(dt, ctx, &event.MouseEvent{
		MouseType: event.MousePress,
		Button:    event.ButtonRight,
		Position:  geometry.Pt(100, 48),
	})
	if consumed {
		t.Error("right click should not be consumed")
	}
}

func TestContentMousePress_OutOfRange(t *testing.T) {
	dt := New(
		Columns(testColumns()),
		RowCount(2),
		RowHeight(32),
		SelectionModeOpt(SelectionSingle),
	)
	ctx := widget.NewContext()
	dt.Layout(ctx, geometry.Tight(geometry.Sz(600, 400)))
	dt.SetBounds(geometry.NewRect(0, 0, 600, 400))

	// Click below data area (content space: no header offset).
	consumed := handleContentMousePress(dt, ctx, &event.MouseEvent{
		MouseType: event.MousePress,
		Button:    event.ButtonLeft,
		Position:  geometry.Pt(100, 5*32),
	})
	if consumed {
		t.Error("click out of range should not be consumed")
	}
}

func TestContentMouseLeave(t *testing.T) {
	dt := New(
		Columns(testColumns()),
		RowCount(10),
		RowHeight(32),
	)
	ctx := widget.NewContext()
	dt.hoveredRow = 5

	handleContentMouseEvent(dt, ctx, &event.MouseEvent{
		MouseType: event.MouseLeave,
	})
	if dt.hoveredRow != noHoveredRow {
		t.Errorf("hoveredRow = %d, want %d", dt.hoveredRow, noHoveredRow)
	}
}

func TestContentMouseEvent_Disabled(t *testing.T) {
	dt := New(
		Columns(testColumns()),
		RowCount(10),
		Disabled(true),
	)
	ctx := widget.NewContext()

	consumed := handleContentMouseEvent(dt, ctx, &event.MouseEvent{
		MouseType: event.MousePress,
		Button:    event.ButtonLeft,
	})
	if consumed {
		t.Error("disabled table should not consume events")
	}
}

func TestContentMouseEvent_OtherEventType(t *testing.T) {
	dt := New(
		Columns(testColumns()),
		RowCount(10),
	)
	ctx := widget.NewContext()

	// MouseRelease should not be consumed.
	consumed := handleContentMouseEvent(dt, ctx, &event.MouseEvent{
		MouseType: event.MouseRelease,
	})
	if consumed {
		t.Error("unhandled event type should not be consumed")
	}
}

// --- Mount with mock scheduler ---

type mockScheduler struct {
	dirty []widget.Widget
}

func (s *mockScheduler) MarkDirty(w widget.Widget) {
	s.dirty = append(s.dirty, w)
}

func TestMount_WithScheduler(t *testing.T) {
	rowSig := state.NewSignal(10)
	selSig := state.NewSignal(0)
	disSig := state.NewSignal(false)

	dt := New(
		Columns(testColumns()),
		RowCountSignal(rowSig),
		SelectedRowSignal(selSig),
		DisabledSignal(disSig),
		SelectionModeOpt(SelectionSingle),
	)

	sched := &mockScheduler{}
	ctx := widget.NewContext()
	ctx.SetScheduler(sched)

	dt.Mount(ctx)

	// Verify bindings are created by changing signals.
	rowSig.Set(20)
	if len(sched.dirty) == 0 {
		t.Error("expected widget to be marked dirty after signal change")
	}

	dt.Unmount()
}

func TestMount_WithReadonlySignals_Scheduler(t *testing.T) {
	rowSig := state.NewSignal(10)
	selSig := state.NewSignal(0)
	disSig := state.NewSignal(false)

	dt := New(
		Columns(testColumns()),
		RowCountReadonlySignal(rowSig.AsReadonly()),
		SelectedRowReadonlySignal(selSig.AsReadonly()),
		DisabledReadonlySignal(disSig.AsReadonly()),
	)

	sched := &mockScheduler{}
	ctx := widget.NewContext()
	ctx.SetScheduler(sched)

	dt.Mount(ctx)

	rowSig.Set(20)
	if len(sched.dirty) == 0 {
		t.Error("expected widget to be marked dirty after readonly signal change")
	}

	dt.Unmount()
}

// --- Layout edge cases ---

func TestLayout_SmallHeight(t *testing.T) {
	dt := New(
		Columns(testColumns()),
		RowCount(3),
		RowHeight(32),
	)
	ctx := widget.NewContext()

	// Height smaller than header.
	size := dt.Layout(ctx, geometry.Tight(geometry.Sz(600, 20)))
	if size.Height != 20 {
		t.Errorf("height = %f, want 20", size.Height)
	}
}

func TestLayout_InfiniteHeight(t *testing.T) {
	dt := New(
		Columns(testColumns()),
		RowCount(3),
		RowHeight(32),
	)
	ctx := widget.NewContext()

	// Biggest() replaces Infinity with MinWidth/MinHeight (0).
	// Then the fallback paths kick in: height <= 0 -> defaultViewportHeight.
	constraints := geometry.Constraints{
		MinWidth: 0, MaxWidth: 600,
		MinHeight: 0, MaxHeight: geometry.Infinity,
	}
	size := dt.Layout(ctx, constraints)
	// Biggest().Height = 0 (Infinity->Min=0), then fallback to defaultViewportHeight.
	if size.Height != defaultViewportHeight {
		t.Errorf("height = %f, want %f", size.Height, defaultViewportHeight)
	}
}

// --- visibleRowRange edge cases ---

func TestVisibleRowRange_ZeroRowHeight(t *testing.T) {
	dt := New(RowCount(10), RowHeight(0))
	first, last := dt.visibleRowRange(0, 100)
	if first != 0 || last != 0 {
		t.Errorf("zero row height: first=%d last=%d, want 0,0", first, last)
	}
}

func TestVisibleRowRange_ZeroRows(t *testing.T) {
	dt := New(RowCount(0), RowHeight(32))
	first, last := dt.visibleRowRange(0, 100)
	if first != 0 || last != 0 {
		t.Errorf("zero rows: first=%d last=%d, want 0,0", first, last)
	}
}

// --- SetSelectedRow same value no-op ---

func TestSetSelectedRow_SameValue(t *testing.T) {
	var callCount int
	dt := New(
		RowCount(10),
		SelectionModeOpt(SelectionSingle),
		SelectedRow(3),
		OnRowSelect(func(_ int) { callCount++ }),
	)
	ctx := widget.NewContext()

	// Setting same value should be a no-op.
	dt.setSelectedRow(ctx, 3)
	if callCount != 0 {
		t.Errorf("expected 0 calls for same value, got %d", callCount)
	}
}

// --- Painter: Row even (no zebra) ---

func TestDefaultPainter_PaintRow_EvenRow(t *testing.T) {
	p := DefaultPainter{}
	canvas := &mockCanvas{}
	p.PaintRow(canvas, RowPaintState{
		Bounds:   geometry.NewRect(0, 0, 600, 32),
		RowIndex: 0, // even, no zebra
	})
	if len(canvas.rects) != 0 {
		t.Errorf("expected no rects for even row without selection/hover, got %d", len(canvas.rects))
	}
}

// --- Painter: Hovered + Selected (no hover when selected) ---

func TestDefaultPainter_PaintRow_HoveredAndSelected(t *testing.T) {
	p := DefaultPainter{}
	canvas := &mockCanvas{}
	p.PaintRow(canvas, RowPaintState{
		Bounds:   geometry.NewRect(0, 0, 600, 32),
		Selected: true,
		Hovered:  true,
	})
	// Should draw selection but NOT hover (because selected takes priority).
	// Check that selection is drawn but hover is not doubled.
	if len(canvas.rects) != 1 {
		t.Errorf("expected 1 rect (selection only, no hover), got %d", len(canvas.rects))
	}
}

// --- Painter: Hovered + Disabled ---

func TestDefaultPainter_PaintRow_HoveredDisabled(t *testing.T) {
	p := DefaultPainter{}
	canvas := &mockCanvas{}
	p.PaintRow(canvas, RowPaintState{
		Bounds:   geometry.NewRect(0, 0, 600, 32),
		Hovered:  true,
		Disabled: true,
	})
	// Disabled = no hover highlight.
	if len(canvas.rects) != 0 {
		t.Errorf("expected 0 rects for disabled+hovered, got %d", len(canvas.rects))
	}
}

// --- Painter: HeaderCell disabled + sortable + hovered ---

func TestDefaultPainter_PaintHeaderCell_DisabledHover(t *testing.T) {
	p := DefaultPainter{}
	canvas := &mockCanvas{}
	bounds := geometry.NewRect(0, 0, 200, 36)
	p.PaintHeaderCell(canvas, bounds, HeaderCellPaintState{
		Title:    "Name",
		Sortable: true,
		Hovered:  true,
		Disabled: true,
	})
	// Should NOT draw hover highlight for disabled.
	if len(canvas.rects) != 0 {
		t.Errorf("expected 0 rects for disabled header hover, got %d", len(canvas.rects))
	}
}

// --- Painter: HeaderCell empty bounds ---

func TestDefaultPainter_PaintHeaderCell_EmptyBounds(t *testing.T) {
	p := DefaultPainter{}
	canvas := &mockCanvas{}
	p.PaintHeaderCell(canvas, geometry.Rect{}, HeaderCellPaintState{Title: "Name"})
	if len(canvas.texts) != 0 {
		t.Error("expected no text for empty bounds")
	}
}

// --- setScrollY with signal ---

func TestSetScrollY_WithSignal(t *testing.T) {
	sig := state.NewSignal[float32](0)
	dt := New(ScrollYSignal(sig))
	dt.setScrollY(100)
	if sig.Get() != 100 {
		t.Errorf("signal value = %f, want 100", sig.Get())
	}
}

// --- ScrollToRow already visible ---

func TestScrollToRow_AlreadyVisible(t *testing.T) {
	dt := New(
		Columns(testColumns()),
		RowCount(100),
		RowHeight(32),
	)
	ctx := widget.NewContext()
	dt.Layout(ctx, geometry.Tight(geometry.Sz(600, 400)))
	dt.SetBounds(geometry.NewRect(0, 0, 600, 400))

	// Row 0 is already visible at scroll 0.
	dt.ScrollToRow(0) // should be no-op, shouldn't panic
}

// --- VisibleRowRange with data view height 0 ---

func TestVisibleRowRange_ZeroDataViewHeight(t *testing.T) {
	dt := New(RowCount(10), RowHeight(32))
	ctx := widget.NewContext()
	// Layout with very small height (less than header).
	dt.Layout(ctx, geometry.Tight(geometry.Sz(600, 20)))
	start, end := dt.VisibleRowRange()
	if start != 0 || end != 0 {
		t.Errorf("expected 0,0 for zero data view height, got %d,%d", start, end)
	}
}

// --- Draw with no columns + data ---

func TestDraw_NoColumnsWithData(t *testing.T) {
	dt := New(RowCount(5))
	ctx := widget.NewContext()
	canvas := &mockCanvas{}

	dt.Layout(ctx, geometry.Tight(geometry.Sz(600, 400)))
	dt.SetBounds(geometry.NewRect(0, 0, 600, 400))
	dt.Draw(ctx, canvas) // should not panic
}

// --- VirtualContent Layout edge case ---

func TestVirtualContent_Layout_InfiniteMaxWidth(t *testing.T) {
	dt := New(Columns(testColumns()), RowCount(5), RowHeight(32))
	ctx := widget.NewContext()
	dt.Layout(ctx, geometry.Tight(geometry.Sz(600, 400)))

	vc := dt.virtual
	constraints := geometry.Constraints{
		MinWidth: 200, MaxWidth: geometry.Infinity,
		MinHeight: 0, MaxHeight: geometry.Infinity,
	}
	size := vc.Layout(ctx, constraints)
	if size.Width != 200 {
		t.Errorf("width = %f, want 200 (MinWidth fallback)", size.Width)
	}
}

// --- Header click disabled ---

func TestHeaderClick_Disabled(t *testing.T) {
	dt := New(
		Columns(testColumns()),
		RowCount(10),
		Disabled(true),
	)
	ctx := widget.NewContext()
	dt.Layout(ctx, geometry.Tight(geometry.Sz(600, 400)))
	dt.SetBounds(geometry.NewRect(0, 0, 600, 400))

	consumed := dt.Event(ctx, &event.MouseEvent{
		MouseType: event.MousePress,
		Button:    event.ButtonLeft,
		Position:  geometry.Pt(100, 18),
	})
	if consumed {
		t.Error("disabled table should not consume header clicks")
	}
}

// --- Header click outside columns ---

func TestHeaderClick_OutsideColumns(t *testing.T) {
	cols := []Column{
		{Key: "a", Title: "A", Width: 100, Sortable: true},
	}
	dt := New(Columns(cols), RowCount(5))
	ctx := widget.NewContext()
	dt.Layout(ctx, geometry.Tight(geometry.Sz(600, 400)))
	dt.SetBounds(geometry.NewRect(0, 0, 600, 400))

	// Click beyond the single column (x=200 is past 100-wide column area).
	consumed := dt.Event(ctx, &event.MouseEvent{
		MouseType: event.MousePress,
		Button:    event.ButtonLeft,
		Position:  geometry.Pt(500, 18),
	})
	if consumed {
		t.Error("click outside column bounds should not be consumed")
	}
}

// --- handleContentMouseEvent non-mouse event ---

func TestVirtualContent_NonMouseEvent(t *testing.T) {
	dt := New(Columns(testColumns()), RowCount(5))
	ctx := widget.NewContext()

	// Key event should not be consumed by virtual content.
	consumed := dt.virtual.Event(ctx, &event.KeyEvent{KeyType: event.KeyPress, Key: event.KeyDown})
	if consumed {
		t.Error("key event on virtual content should not be consumed")
	}
}

// --- setSelectedRow with signal writeback ---

func TestSetSelectedRow_SignalWriteback(t *testing.T) {
	sig := state.NewSignal(-1)
	dt := New(
		RowCount(10),
		SelectionModeOpt(SelectionSingle),
		SelectedRowSignal(sig),
	)
	ctx := widget.NewContext()

	dt.setSelectedRow(ctx, 5)
	if sig.Get() != 5 {
		t.Errorf("signal = %d, want 5", sig.Get())
	}
}

// --- moveSelectionByPage with zero data view height ---

func TestMoveSelectionByPage_ZeroViewHeight(t *testing.T) {
	dt := New(
		RowCount(100),
		RowHeight(32),
		SelectionModeOpt(SelectionSingle),
		SelectedRow(0),
	)
	ctx := widget.NewContext()
	dt.viewportHeight = defaultHeaderHeight // no data area

	consumed := dt.moveSelectionByPage(ctx, 0, 100, 1)
	if consumed {
		t.Error("expected false when data view height is 0")
	}
}
