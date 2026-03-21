package devtools_test

import (
	"testing"

	"github.com/gogpu/ui/core/collapsible"
	"github.com/gogpu/ui/core/datatable"
	"github.com/gogpu/ui/core/docking"
	"github.com/gogpu/ui/core/linechart"
	"github.com/gogpu/ui/core/listview"
	"github.com/gogpu/ui/core/menu"
	"github.com/gogpu/ui/core/popover"
	"github.com/gogpu/ui/core/progress"
	"github.com/gogpu/ui/core/splitview"
	"github.com/gogpu/ui/core/toolbar"
	"github.com/gogpu/ui/core/treeview"
	"github.com/gogpu/ui/geometry"
	"github.com/gogpu/ui/icon"
	"github.com/gogpu/ui/theme/devtools"
	"github.com/gogpu/ui/widget"
)

// ===== TreeViewPainter tests =====

func TestTreeViewPainterImplementsInterface(t *testing.T) {
	var _ treeview.Painter = devtools.TreeViewPainter{}
}

func TestTreeViewPaintRowBackgroundEmpty(t *testing.T) {
	canvas := &recordCanvas{}
	painter := devtools.TreeViewPainter{}
	painter.PaintRowBackground(canvas, treeview.RowPaintState{Bounds: geometry.Rect{}})
	if len(canvas.calls) != 0 {
		t.Errorf("empty bounds should produce no calls, got %d", len(canvas.calls))
	}
}

func TestTreeViewPaintRowBackgroundHovered(t *testing.T) {
	canvas := &recordCanvas{}
	painter := devtools.TreeViewPainter{}
	painter.PaintRowBackground(canvas, treeview.RowPaintState{
		Bounds:  testBounds(),
		Hovered: true,
	})

	rects := canvas.methodCalls(methodDrawRect)
	if len(rects) < 1 {
		t.Error("hovered row should draw at least 1 DrawRect (hover bg)")
	}
}

func TestTreeViewPaintSelectionEmpty(t *testing.T) {
	canvas := &recordCanvas{}
	painter := devtools.TreeViewPainter{}
	painter.PaintSelection(canvas, treeview.RowPaintState{Bounds: geometry.Rect{}, Selected: true})
	if len(canvas.calls) != 0 {
		t.Errorf("empty bounds should produce no calls, got %d", len(canvas.calls))
	}
}

func TestTreeViewPaintSelectionSelected(t *testing.T) {
	canvas := &recordCanvas{}
	painter := devtools.TreeViewPainter{}
	painter.PaintSelection(canvas, treeview.RowPaintState{
		Bounds:   testBounds(),
		Selected: true,
	})

	rects := canvas.methodCalls(methodDrawRect)
	if len(rects) < 1 {
		t.Error("selected row should draw at least 1 DrawRect (selection bg)")
	}
}

func TestTreeViewPaintSelectionFocused(t *testing.T) {
	canvas := &recordCanvas{}
	painter := devtools.TreeViewPainter{}
	painter.PaintSelection(canvas, treeview.RowPaintState{
		Bounds:   testBounds(),
		Selected: true,
		Focused:  true,
	})

	strokes := canvas.methodCalls(methodStrokeRect)
	if len(strokes) < 1 {
		t.Error("focused selection should draw at least 1 StrokeRect (focus ring)")
	}
}

func TestTreeViewPaintExpandIcon(t *testing.T) {
	canvas := &recordCanvas{}
	painter := devtools.TreeViewPainter{}
	painter.PaintExpandIcon(canvas, treeview.ExpandIconState{
		Bounds:   geometry.NewRect(0, 0, 16, 16),
		Expanded: true,
	})

	lines := canvas.methodCalls(methodDrawLine)
	if len(lines) < 2 { //nolint:mnd // down chevron = 2 lines
		t.Errorf("expanded icon should draw at least 2 DrawLine, got %d", len(lines))
	}
}

func TestTreeViewPaintExpandIconCollapsed(t *testing.T) {
	canvas := &recordCanvas{}
	painter := devtools.TreeViewPainter{}
	painter.PaintExpandIcon(canvas, treeview.ExpandIconState{
		Bounds:   geometry.NewRect(0, 0, 16, 16),
		Expanded: false,
	})

	lines := canvas.methodCalls(methodDrawLine)
	if len(lines) < 2 { //nolint:mnd // right chevron = 2 lines
		t.Errorf("collapsed icon should draw at least 2 DrawLine, got %d", len(lines))
	}
}

func TestTreeViewPaintConnectorLines(t *testing.T) {
	canvas := &recordCanvas{}
	painter := devtools.TreeViewPainter{}
	painter.PaintConnectorLines(canvas, treeview.ConnectorState{
		RowBounds:     testBounds(),
		Depth:         1,
		IndentWidth:   20,
		IsLastChild:   false,
		ParentHasMore: []bool{true},
	})

	lines := canvas.methodCalls(methodDrawLine)
	if len(lines) < 2 { //nolint:mnd // vertical + horizontal
		t.Errorf("connector should draw at least 2 DrawLine, got %d", len(lines))
	}
}

func TestTreeViewPaintConnectorLinesDepthZero(t *testing.T) {
	canvas := &recordCanvas{}
	painter := devtools.TreeViewPainter{}
	painter.PaintConnectorLines(canvas, treeview.ConnectorState{
		RowBounds: testBounds(),
		Depth:     0,
	})
	if len(canvas.calls) != 0 {
		t.Errorf("depth 0 should produce no calls, got %d", len(canvas.calls))
	}
}

func TestTreeViewPaintLabel(t *testing.T) {
	canvas := &recordCanvas{}
	painter := devtools.TreeViewPainter{}
	painter.PaintLabel(canvas, treeview.LabelState{
		Bounds: testBounds(),
		Text:   "src/main.go",
	})

	texts := canvas.methodCalls(methodDrawText)
	if len(texts) != 1 {
		t.Fatalf("should draw 1 DrawText, got %d", len(texts))
	}
	if texts[0].text != "src/main.go" {
		t.Errorf("text = %q, want 'src/main.go'", texts[0].text)
	}
}

func TestTreeViewPaintLabelDisabled(t *testing.T) {
	canvas := &recordCanvas{}
	painter := devtools.TreeViewPainter{}
	painter.PaintLabel(canvas, treeview.LabelState{
		Bounds:   testBounds(),
		Text:     "disabled",
		Disabled: true,
	})

	if len(canvas.calls) == 0 {
		t.Error("disabled label should still produce draw calls")
	}
}

func TestTreeViewPaintEmptyState(t *testing.T) {
	canvas := &recordCanvas{}
	painter := devtools.TreeViewPainter{}
	painter.PaintEmptyState(canvas, testBounds())

	texts := canvas.methodCalls(methodDrawText)
	if len(texts) != 1 {
		t.Fatalf("should draw 1 DrawText, got %d", len(texts))
	}
}

func TestTreeViewPainterWithTheme(t *testing.T) {
	theme := devtools.NewDarkTheme()
	painter := devtools.TreeViewPainter{Theme: theme}
	canvas := &recordCanvas{}

	painter.PaintRowBackground(canvas, treeview.RowPaintState{
		Bounds:  testBounds(),
		Hovered: true,
	})
	if len(canvas.calls) == 0 {
		t.Fatal("themed painter should produce draw calls")
	}
}

// ===== DataTablePainter tests =====

func TestDataTablePainterImplementsInterface(t *testing.T) {
	var _ datatable.Painter = devtools.DataTablePainter{}
}

func TestDataTablePaintHeaderEmpty(t *testing.T) {
	canvas := &recordCanvas{}
	painter := devtools.DataTablePainter{}
	painter.PaintHeader(canvas, geometry.Rect{}, datatable.HeaderPaintState{})
	if len(canvas.calls) != 0 {
		t.Errorf("empty bounds should produce no calls, got %d", len(canvas.calls))
	}
}

func TestDataTablePaintHeader(t *testing.T) {
	canvas := &recordCanvas{}
	painter := devtools.DataTablePainter{}
	painter.PaintHeader(canvas, testBounds(), datatable.HeaderPaintState{})

	rects := canvas.methodCalls(methodDrawRect)
	if len(rects) < 2 { //nolint:mnd // bg + divider
		t.Errorf("header should draw at least 2 DrawRect (bg + divider), got %d", len(rects))
	}
}

func TestDataTablePaintHeaderCell(t *testing.T) {
	canvas := &recordCanvas{}
	painter := devtools.DataTablePainter{}
	painter.PaintHeaderCell(canvas, testBounds(), datatable.HeaderCellPaintState{
		Title: "Name",
		Align: widget.TextAlignLeft,
	})

	texts := canvas.methodCalls(methodDrawText)
	if len(texts) != 1 {
		t.Fatalf("should draw 1 DrawText, got %d", len(texts))
	}
	if texts[0].text != "Name" {
		t.Errorf("text = %q, want 'Name'", texts[0].text)
	}
}

func TestDataTablePaintRow(t *testing.T) {
	canvas := &recordCanvas{}
	painter := devtools.DataTablePainter{}
	painter.PaintRow(canvas, datatable.RowPaintState{
		Bounds:   testBounds(),
		RowIndex: 1,
		Selected: true,
	})

	rects := canvas.methodCalls(methodDrawRect)
	if len(rects) < 2 { //nolint:mnd // zebra + selection
		t.Errorf("selected odd row should draw at least 2 DrawRect, got %d", len(rects))
	}
}

func TestDataTablePaintRowHovered(t *testing.T) {
	canvas := &recordCanvas{}
	painter := devtools.DataTablePainter{}
	painter.PaintRow(canvas, datatable.RowPaintState{
		Bounds:   testBounds(),
		RowIndex: 0,
		Hovered:  true,
	})

	rects := canvas.methodCalls(methodDrawRect)
	if len(rects) < 1 {
		t.Error("hovered row should draw at least 1 DrawRect (hover)")
	}
}

func TestDataTablePaintCell(t *testing.T) {
	canvas := &recordCanvas{}
	painter := devtools.DataTablePainter{}
	painter.PaintCell(canvas, datatable.CellPaintState{
		Bounds: testBounds(),
		Value:  "John",
		Align:  widget.TextAlignLeft,
	})

	texts := canvas.methodCalls(methodDrawText)
	if len(texts) != 1 {
		t.Fatalf("should draw 1 DrawText, got %d", len(texts))
	}
}

func TestDataTablePaintEmptyState(t *testing.T) {
	canvas := &recordCanvas{}
	painter := devtools.DataTablePainter{}
	painter.PaintEmptyState(canvas, testBounds())

	texts := canvas.methodCalls(methodDrawText)
	if len(texts) != 1 {
		t.Fatalf("should draw 1 DrawText, got %d", len(texts))
	}
}

func TestDataTablePainterWithTheme(t *testing.T) {
	theme := devtools.NewDarkTheme()
	painter := devtools.DataTablePainter{Theme: theme}
	canvas := &recordCanvas{}

	painter.PaintHeader(canvas, testBounds(), datatable.HeaderPaintState{})
	if len(canvas.calls) == 0 {
		t.Fatal("themed painter should produce draw calls")
	}
}

// ===== ToolbarPainter tests =====

func TestToolbarPainterImplementsInterface(t *testing.T) {
	var _ toolbar.Painter = devtools.ToolbarPainter{}
}

func TestToolbarPaintToolbarEmpty(t *testing.T) {
	canvas := &recordCanvas{}
	painter := devtools.ToolbarPainter{}
	painter.PaintToolbar(canvas, toolbar.PaintToolbarState{Bounds: geometry.Rect{}})
	if len(canvas.calls) != 0 {
		t.Errorf("empty bounds should produce no calls, got %d", len(canvas.calls))
	}
}

func TestToolbarPaintToolbar(t *testing.T) {
	canvas := &recordCanvas{}
	painter := devtools.ToolbarPainter{}
	painter.PaintToolbar(canvas, toolbar.PaintToolbarState{Bounds: testBounds()})

	// Transparent bg = no DrawRect for bg, but 1 for bottom border.
	rects := canvas.methodCalls(methodDrawRect)
	if len(rects) < 1 {
		t.Error("toolbar should draw at least 1 DrawRect (bottom border)")
	}
}

func TestToolbarPaintButtonItem(t *testing.T) {
	canvas := &recordCanvas{}
	painter := devtools.ToolbarPainter{}
	painter.PaintButtonItem(canvas, toolbar.PaintButtonState{
		Label:   "Save",
		Icon:    icon.Close,
		Bounds:  geometry.NewRect(0, 0, 28, 28),
		Hovered: true,
	})

	roundRects := canvas.methodCalls(methodDrawRoundRect)
	if len(roundRects) < 1 {
		t.Error("hovered button should draw at least 1 DrawRoundRect (hover bg)")
	}
}

func TestToolbarPaintButtonItemDisabled(t *testing.T) {
	canvas := &recordCanvas{}
	painter := devtools.ToolbarPainter{}
	painter.PaintButtonItem(canvas, toolbar.PaintButtonState{
		Label:    "Save",
		Icon:     icon.Close,
		Bounds:   geometry.NewRect(0, 0, 28, 28),
		Disabled: true,
	})

	// Should not draw hover/press bg when disabled.
	roundRects := canvas.methodCalls(methodDrawRoundRect)
	if len(roundRects) != 0 {
		t.Errorf("disabled button should draw 0 DrawRoundRect, got %d", len(roundRects))
	}
}

func TestToolbarPaintButtonItemFocused(t *testing.T) {
	canvas := &recordCanvas{}
	painter := devtools.ToolbarPainter{}
	painter.PaintButtonItem(canvas, toolbar.PaintButtonState{
		Label:   "Save",
		Icon:    icon.Close,
		Bounds:  geometry.NewRect(0, 0, 28, 28),
		Focused: true,
	})

	strokes := canvas.methodCalls(methodStrokeRoundRect)
	if len(strokes) < 1 {
		t.Error("focused button should draw at least 1 StrokeRoundRect (focus ring)")
	}
}

func TestToolbarPaintButtonItemWithLabel(t *testing.T) {
	canvas := &recordCanvas{}
	painter := devtools.ToolbarPainter{}
	painter.PaintButtonItem(canvas, toolbar.PaintButtonState{
		Label:     "Save",
		Icon:      icon.Close,
		ShowLabel: true,
		Bounds:    geometry.NewRect(0, 0, 80, 28),
	})

	texts := canvas.methodCalls(methodDrawText)
	if len(texts) < 1 {
		t.Error("button with label should draw at least 1 DrawText")
	}
}

func TestToolbarPaintSeparator(t *testing.T) {
	canvas := &recordCanvas{}
	painter := devtools.ToolbarPainter{}
	painter.PaintSeparator(canvas, geometry.NewRect(40, 0, 8, 28))

	lines := canvas.methodCalls(methodDrawLine)
	if len(lines) != 1 {
		t.Errorf("separator should draw 1 DrawLine, got %d", len(lines))
	}
}

func TestToolbarPainterWithTheme(t *testing.T) {
	theme := devtools.NewDarkTheme()
	painter := devtools.ToolbarPainter{Theme: theme}
	canvas := &recordCanvas{}

	painter.PaintToolbar(canvas, toolbar.PaintToolbarState{Bounds: testBounds()})
	if len(canvas.calls) == 0 {
		t.Fatal("themed painter should produce draw calls")
	}
}

// ===== MenuPainter tests =====

func TestMenuPainterImplementsInterface(t *testing.T) {
	var _ menu.Painter = devtools.MenuPainter{}
}

func TestMenuPaintMenuBarEmpty(t *testing.T) {
	canvas := &recordCanvas{}
	painter := devtools.MenuPainter{}
	painter.PaintMenuBar(canvas, &menu.MenuBarPaintState{Bounds: geometry.Rect{}})
	if len(canvas.calls) != 0 {
		t.Errorf("empty bounds should produce no calls, got %d", len(canvas.calls))
	}
}

func TestMenuPaintMenuBar(t *testing.T) {
	canvas := &recordCanvas{}
	painter := devtools.MenuPainter{}
	painter.PaintMenuBar(canvas, &menu.MenuBarPaintState{
		Bounds: geometry.NewRect(0, 0, 400, 28),
		Menus: []menu.TopMenu{
			{Label: "File"},
			{Label: "Edit"},
		},
		MenuRects: []geometry.Rect{
			geometry.NewRect(0, 0, 50, 28),
			geometry.NewRect(50, 0, 50, 28),
		},
		OpenIndex:    -1,
		HoveredIndex: 0,
	})

	rects := canvas.methodCalls(methodDrawRect)
	if len(rects) < 2 { //nolint:mnd // bg + border + hover
		t.Errorf("menu bar should draw at least 2 DrawRect, got %d", len(rects))
	}

	texts := canvas.methodCalls(methodDrawText)
	if len(texts) < 2 { //nolint:mnd // 2 labels
		t.Errorf("menu bar should draw at least 2 DrawText, got %d", len(texts))
	}
}

func TestMenuPaintMenuEmpty(t *testing.T) {
	canvas := &recordCanvas{}
	painter := devtools.MenuPainter{}
	painter.PaintMenu(canvas, &menu.MenuPaintState{Bounds: geometry.Rect{}})
	if len(canvas.calls) != 0 {
		t.Errorf("empty bounds should produce no calls, got %d", len(canvas.calls))
	}
}

func TestMenuPaintMenu(t *testing.T) {
	canvas := &recordCanvas{}
	painter := devtools.MenuPainter{}
	painter.PaintMenu(canvas, &menu.MenuPaintState{
		Bounds: geometry.NewRect(0, 28, 200, 100),
		Items: []menu.MenuItem{
			{Label: "New File", Shortcut: "Ctrl+N"},
			{Label: "Open", Shortcut: "Ctrl+O"},
		},
		HighlightedIndex: 0,
		ItemHeight:       28,
		SeparatorHeight:  8,
	})

	roundRects := canvas.methodCalls(methodDrawRoundRect)
	if len(roundRects) < 1 {
		t.Error("menu should draw at least 1 DrawRoundRect (surface)")
	}

	texts := canvas.methodCalls(methodDrawText)
	if len(texts) < 2 { //nolint:mnd // labels
		t.Errorf("menu should draw at least 2 DrawText (items), got %d", len(texts))
	}
}

func TestMenuPaintMenuWithSeparator(t *testing.T) {
	canvas := &recordCanvas{}
	painter := devtools.MenuPainter{}
	painter.PaintMenu(canvas, &menu.MenuPaintState{
		Bounds: geometry.NewRect(0, 28, 200, 100),
		Items: []menu.MenuItem{
			{Label: "Cut"},
			menu.Sep(),
			{Label: "Paste"},
		},
		HighlightedIndex: -1,
		ItemHeight:       28,
		SeparatorHeight:  8,
	})

	lines := canvas.methodCalls(methodDrawLine)
	if len(lines) < 1 {
		t.Error("menu with separator should draw at least 1 DrawLine")
	}
}

func TestMenuPainterWithTheme(t *testing.T) {
	theme := devtools.NewDarkTheme()
	painter := devtools.MenuPainter{Theme: theme}
	canvas := &recordCanvas{}

	painter.PaintMenuBar(canvas, &menu.MenuBarPaintState{
		Bounds:       geometry.NewRect(0, 0, 400, 28),
		OpenIndex:    -1,
		HoveredIndex: -1,
	})
	if len(canvas.calls) == 0 {
		t.Fatal("themed painter should produce draw calls")
	}
}

// ===== CollapsiblePainter tests =====

func TestCollapsiblePainterImplementsInterface(t *testing.T) {
	var _ collapsible.Painter = devtools.CollapsiblePainter{}
}

func TestCollapsiblePaintHeaderEmpty(t *testing.T) {
	canvas := &recordCanvas{}
	painter := devtools.CollapsiblePainter{}
	painter.PaintHeader(canvas, collapsible.HeaderState{Bounds: geometry.Rect{}})
	if len(canvas.calls) != 0 {
		t.Errorf("empty bounds should produce no calls, got %d", len(canvas.calls))
	}
}

func TestCollapsiblePaintHeader(t *testing.T) {
	canvas := &recordCanvas{}
	painter := devtools.CollapsiblePainter{}
	painter.PaintHeader(canvas, collapsible.HeaderState{
		Title:         "Section",
		Expanded:      true,
		Bounds:        testBounds(),
		ArrowProgress: 1.0,
	})

	rects := canvas.methodCalls(methodDrawRect)
	if len(rects) < 2 { //nolint:mnd // bg + border
		t.Errorf("header should draw at least 2 DrawRect (bg + border), got %d", len(rects))
	}

	texts := canvas.methodCalls(methodDrawText)
	if len(texts) < 1 {
		t.Error("header should draw at least 1 DrawText (title)")
	}
}

func TestCollapsiblePaintHeaderHovered(t *testing.T) {
	canvas := &recordCanvas{}
	painter := devtools.CollapsiblePainter{}
	painter.PaintHeader(canvas, collapsible.HeaderState{
		Title:   "Hover",
		Bounds:  testBounds(),
		Hovered: true,
	})

	if len(canvas.calls) == 0 {
		t.Error("hovered header should produce draw calls")
	}
}

func TestCollapsiblePaintHeaderFocused(t *testing.T) {
	canvas := &recordCanvas{}
	painter := devtools.CollapsiblePainter{}
	painter.PaintHeader(canvas, collapsible.HeaderState{
		Title:   "Focus",
		Bounds:  testBounds(),
		Focused: true,
	})

	strokes := canvas.methodCalls(methodStrokeRect)
	if len(strokes) < 1 {
		t.Error("focused header should draw at least 1 StrokeRect (focus ring)")
	}
}

func TestCollapsiblePainterWithTheme(t *testing.T) {
	theme := devtools.NewDarkTheme()
	painter := devtools.CollapsiblePainter{Theme: theme}
	canvas := &recordCanvas{}

	painter.PaintHeader(canvas, collapsible.HeaderState{
		Title:  "Themed",
		Bounds: testBounds(),
	})
	if len(canvas.calls) == 0 {
		t.Fatal("themed painter should produce draw calls")
	}
}

// ===== ProgressPainter tests =====

func TestProgressPainterImplementsInterface(t *testing.T) {
	var _ progress.Painter = devtools.ProgressPainter{}
}

func TestProgressPaintEmpty(t *testing.T) {
	canvas := &recordCanvas{}
	painter := devtools.ProgressPainter{}
	painter.PaintProgress(canvas, progress.PaintState{Bounds: geometry.Rect{}})
	if len(canvas.calls) != 0 {
		t.Errorf("empty bounds should produce no calls, got %d", len(canvas.calls))
	}
}

func TestProgressPaintDeterminate(t *testing.T) {
	canvas := &recordCanvas{}
	painter := devtools.ProgressPainter{}
	painter.PaintProgress(canvas, progress.PaintState{
		Value:       0.5,
		Bounds:      geometry.NewRect(10, 10, 40, 40),
		Diameter:    36,
		StrokeWidth: 2,
	})

	strokeCircles := canvas.methodCalls(methodStrokeCircle)
	if len(strokeCircles) < 1 {
		t.Error("determinate should draw at least 1 StrokeCircle (track)")
	}

	lines := canvas.methodCalls(methodDrawLine)
	if len(lines) < 1 {
		t.Error("determinate should draw at least 1 DrawLine (arc)")
	}
}

func TestProgressPaintIndeterminate(t *testing.T) {
	canvas := &recordCanvas{}
	painter := devtools.ProgressPainter{}
	painter.PaintProgress(canvas, progress.PaintState{
		Indeterminate: true,
		Bounds:        geometry.NewRect(10, 10, 40, 40),
		Diameter:      36,
		StrokeWidth:   2,
		Rotation:      0.5,
	})

	strokeCircles := canvas.methodCalls(methodStrokeCircle)
	if len(strokeCircles) < 1 {
		t.Error("indeterminate should draw at least 1 StrokeCircle (track)")
	}
}

func TestProgressPaintDeterminateWithLabel(t *testing.T) {
	canvas := &recordCanvas{}
	painter := devtools.ProgressPainter{}
	painter.PaintProgress(canvas, progress.PaintState{
		Value:       0.75,
		Bounds:      geometry.NewRect(10, 10, 60, 60),
		Diameter:    56,
		StrokeWidth: 2,
		ShowLabel:   true,
		Label:       "75%",
	})

	texts := canvas.methodCalls(methodDrawText)
	if len(texts) != 1 {
		t.Fatalf("should draw 1 DrawText (label), got %d", len(texts))
	}
	if texts[0].text != "75%" {
		t.Errorf("text = %q, want '75%%'", texts[0].text)
	}
}

func TestProgressPaintDisabled(t *testing.T) {
	canvas := &recordCanvas{}
	painter := devtools.ProgressPainter{}
	painter.PaintProgress(canvas, progress.PaintState{
		Value:       0.5,
		Bounds:      geometry.NewRect(10, 10, 40, 40),
		Diameter:    36,
		StrokeWidth: 2,
		Disabled:    true,
	})

	if len(canvas.calls) == 0 {
		t.Error("disabled progress should still produce draw calls")
	}
}

func TestProgressPainterWithTheme(t *testing.T) {
	theme := devtools.NewDarkTheme()
	painter := devtools.ProgressPainter{Theme: theme}
	canvas := &recordCanvas{}

	painter.PaintProgress(canvas, progress.PaintState{
		Value:       0.5,
		Bounds:      geometry.NewRect(10, 10, 40, 40),
		Diameter:    36,
		StrokeWidth: 2,
	})
	if len(canvas.calls) == 0 {
		t.Fatal("themed painter should produce draw calls")
	}
}

// ===== SplitViewPainter tests =====

func TestSplitViewPainterImplementsInterface(t *testing.T) {
	var _ splitview.Painter = devtools.SplitViewPainter{}
}

func TestSplitViewPaintDividerEmpty(t *testing.T) {
	canvas := &recordCanvas{}
	painter := devtools.SplitViewPainter{}
	painter.PaintDivider(canvas, splitview.PaintState{DividerRect: geometry.Rect{}})
	if len(canvas.calls) != 0 {
		t.Errorf("empty bounds should produce no calls, got %d", len(canvas.calls))
	}
}

func TestSplitViewPaintDivider(t *testing.T) {
	canvas := &recordCanvas{}
	painter := devtools.SplitViewPainter{}
	painter.PaintDivider(canvas, splitview.PaintState{
		DividerRect: geometry.NewRect(100, 0, 1, 400),
		Orientation: splitview.Horizontal,
	})

	rects := canvas.methodCalls(methodDrawRect)
	if len(rects) < 1 {
		t.Error("divider should draw at least 1 DrawRect")
	}
}

func TestSplitViewPaintDividerHovered(t *testing.T) {
	canvas := &recordCanvas{}
	painter := devtools.SplitViewPainter{}
	painter.PaintDivider(canvas, splitview.PaintState{
		DividerRect: geometry.NewRect(100, 0, 3, 400),
		Hovered:     true,
	})

	if len(canvas.calls) == 0 {
		t.Error("hovered divider should produce draw calls")
	}
}

func TestSplitViewPaintDividerDragging(t *testing.T) {
	canvas := &recordCanvas{}
	painter := devtools.SplitViewPainter{}
	painter.PaintDivider(canvas, splitview.PaintState{
		DividerRect: geometry.NewRect(100, 0, 3, 400),
		Dragging:    true,
	})

	if len(canvas.calls) == 0 {
		t.Error("dragging divider should produce draw calls")
	}
}

func TestSplitViewPainterWithTheme(t *testing.T) {
	theme := devtools.NewDarkTheme()
	painter := devtools.SplitViewPainter{Theme: theme}
	canvas := &recordCanvas{}

	painter.PaintDivider(canvas, splitview.PaintState{
		DividerRect: geometry.NewRect(100, 0, 1, 400),
	})
	if len(canvas.calls) == 0 {
		t.Fatal("themed painter should produce draw calls")
	}
}

// ===== DockingPainter tests =====

func TestDockingPainterImplementsInterface(t *testing.T) {
	var _ docking.Painter = devtools.DockingPainter{}
}

func TestDockingPaintZoneTabsEmpty(t *testing.T) {
	canvas := &recordCanvas{}
	painter := devtools.DockingPainter{}
	painter.PaintZoneTabs(canvas, docking.ZoneTabsPaintState{TabBarBounds: geometry.Rect{}})
	if len(canvas.calls) != 0 {
		t.Errorf("empty bounds should produce no calls, got %d", len(canvas.calls))
	}
}

func TestDockingPaintZoneTabs(t *testing.T) {
	canvas := &recordCanvas{}
	painter := devtools.DockingPainter{}
	painter.PaintZoneTabs(canvas, docking.ZoneTabsPaintState{
		TabBarBounds: geometry.NewRect(0, 0, 400, 24),
		Tabs: []docking.ZoneTabState{
			{Title: "Project", Bounds: geometry.NewRect(0, 0, 80, 24), Active: true},
			{Title: "Structure", Bounds: geometry.NewRect(80, 0, 80, 24)},
		},
		ActiveIdx: 0,
	})

	rects := canvas.methodCalls(methodDrawRect)
	if len(rects) < 2 { //nolint:mnd // bg + indicator
		t.Errorf("zone tabs should draw at least 2 DrawRect, got %d", len(rects))
	}

	texts := canvas.methodCalls(methodDrawText)
	if len(texts) < 2 { //nolint:mnd // 2 tab labels
		t.Errorf("zone tabs should draw at least 2 DrawText, got %d", len(texts))
	}
}

func TestDockingPaintZoneTabsWithCloseable(t *testing.T) {
	canvas := &recordCanvas{}
	painter := devtools.DockingPainter{}
	painter.PaintZoneTabs(canvas, docking.ZoneTabsPaintState{
		TabBarBounds: geometry.NewRect(0, 0, 400, 24),
		Tabs: []docking.ZoneTabState{
			{
				Title:             "Closeable",
				Bounds:            geometry.NewRect(0, 0, 100, 24),
				Active:            true,
				Closeable:         true,
				CloseButtonBounds: geometry.NewRect(80, 4, 16, 16),
			},
		},
		ActiveIdx: 0,
	})

	lines := canvas.methodCalls(methodDrawLine)
	if len(lines) < 2 { //nolint:mnd // X icon = 2 lines
		t.Errorf("closeable tab should draw at least 2 DrawLine, got %d", len(lines))
	}
}

func TestDockingPaintZoneBorder(t *testing.T) {
	canvas := &recordCanvas{}
	painter := devtools.DockingPainter{}
	painter.PaintZoneBorder(canvas, geometry.NewRect(0, 0, 1, 400), docking.Left)

	rects := canvas.methodCalls(methodDrawRect)
	if len(rects) != 1 {
		t.Errorf("zone border should draw 1 DrawRect, got %d", len(rects))
	}
}

func TestDockingPainterWithTheme(t *testing.T) {
	theme := devtools.NewDarkTheme()
	painter := devtools.DockingPainter{Theme: theme}
	canvas := &recordCanvas{}

	painter.PaintZoneTabs(canvas, docking.ZoneTabsPaintState{
		TabBarBounds: geometry.NewRect(0, 0, 400, 24),
		Tabs: []docking.ZoneTabState{
			{Title: "Themed", Bounds: geometry.NewRect(0, 0, 80, 24), Active: true},
		},
		ActiveIdx: 0,
	})
	if len(canvas.calls) == 0 {
		t.Fatal("themed painter should produce draw calls")
	}
}

// ===== PopoverPainter tests =====

func TestPopoverPainterImplementsInterface(t *testing.T) {
	var _ popover.Painter = devtools.PopoverPainter{}
}

func TestPopoverPaintPopoverEmpty(t *testing.T) {
	canvas := &recordCanvas{}
	painter := devtools.PopoverPainter{}
	painter.PaintPopover(canvas, &popover.PopoverPaintState{Bounds: geometry.Rect{}})
	if len(canvas.calls) != 0 {
		t.Errorf("empty bounds should produce no calls, got %d", len(canvas.calls))
	}
}

func TestPopoverPaintPopover(t *testing.T) {
	canvas := &recordCanvas{}
	painter := devtools.PopoverPainter{}
	painter.PaintPopover(canvas, &popover.PopoverPaintState{
		Bounds: geometry.NewRect(100, 100, 200, 150),
	})

	roundRects := canvas.methodCalls(methodDrawRoundRect)
	if len(roundRects) < 2 { //nolint:mnd // shadow + bg
		t.Errorf("popover should draw at least 2 DrawRoundRect (shadow + bg), got %d", len(roundRects))
	}

	strokes := canvas.methodCalls(methodStrokeRoundRect)
	if len(strokes) < 1 {
		t.Error("popover should draw at least 1 StrokeRoundRect (border)")
	}
}

func TestPopoverPaintTooltipEmpty(t *testing.T) {
	canvas := &recordCanvas{}
	painter := devtools.PopoverPainter{}
	painter.PaintTooltip(canvas, &popover.TooltipPaintState{Bounds: geometry.Rect{}})
	if len(canvas.calls) != 0 {
		t.Errorf("empty bounds should produce no calls, got %d", len(canvas.calls))
	}
}

func TestPopoverPaintTooltip(t *testing.T) {
	canvas := &recordCanvas{}
	painter := devtools.PopoverPainter{}
	painter.PaintTooltip(canvas, &popover.TooltipPaintState{
		Bounds: geometry.NewRect(50, 50, 100, 30),
		Text:   "Save file",
	})

	roundRects := canvas.methodCalls(methodDrawRoundRect)
	if len(roundRects) < 1 {
		t.Error("tooltip should draw at least 1 DrawRoundRect (bg)")
	}

	texts := canvas.methodCalls(methodDrawText)
	if len(texts) != 1 {
		t.Fatalf("should draw 1 DrawText, got %d", len(texts))
	}
	if texts[0].text != "Save file" {
		t.Errorf("text = %q, want 'Save file'", texts[0].text)
	}
}

func TestPopoverPainterWithTheme(t *testing.T) {
	theme := devtools.NewDarkTheme()
	painter := devtools.PopoverPainter{Theme: theme}
	canvas := &recordCanvas{}

	painter.PaintPopover(canvas, &popover.PopoverPaintState{
		Bounds: geometry.NewRect(100, 100, 200, 150),
	})
	if len(canvas.calls) == 0 {
		t.Fatal("themed painter should produce draw calls")
	}
}

// ===== LineChartPainter tests =====

func TestLineChartPainterImplementsInterface(t *testing.T) {
	var _ linechart.Painter = devtools.LineChartPainter{}
}

func TestLineChartPaintChartEmpty(t *testing.T) {
	canvas := &recordCanvas{}
	painter := devtools.LineChartPainter{}
	painter.PaintChart(canvas, geometry.Rect{}, linechart.PaintState{})
	if len(canvas.calls) != 0 {
		t.Errorf("empty bounds should produce no calls, got %d", len(canvas.calls))
	}
}

func TestLineChartPaintChart(t *testing.T) {
	canvas := &recordCanvas{}
	painter := devtools.LineChartPainter{}
	painter.PaintChart(canvas, geometry.NewRect(0, 0, 300, 200), linechart.PaintState{
		Series: []linechart.Series{
			{
				Color: widget.Hex(0x3574F0),
				Points: []linechart.DataPoint{
					{Value: 10},
					{Value: 20},
					{Value: 15},
				},
			},
		},
		MaxPoints:  100,
		YMin:       0,
		YMax:       30,
		ShowGrid:   true,
		ShowLabels: true,
	})

	rects := canvas.methodCalls(methodDrawRect)
	if len(rects) < 1 {
		t.Error("chart should draw at least 1 DrawRect (background)")
	}

	lines := canvas.methodCalls(methodDrawLine)
	if len(lines) < 2 { //nolint:mnd // grid + data lines
		t.Errorf("chart should draw at least 2 DrawLine, got %d", len(lines))
	}
}

func TestLineChartPaintChartNoData(t *testing.T) {
	canvas := &recordCanvas{}
	painter := devtools.LineChartPainter{}
	painter.PaintChart(canvas, geometry.NewRect(0, 0, 300, 200), linechart.PaintState{
		MaxPoints: 100,
		YMin:      0,
		YMax:      100,
	})

	rects := canvas.methodCalls(methodDrawRect)
	if len(rects) < 1 {
		t.Error("chart without data should draw at least 1 DrawRect (background)")
	}
}

func TestLineChartPainterWithTheme(t *testing.T) {
	theme := devtools.NewDarkTheme()
	painter := devtools.LineChartPainter{Theme: theme}
	canvas := &recordCanvas{}

	painter.PaintChart(canvas, geometry.NewRect(0, 0, 300, 200), linechart.PaintState{
		MaxPoints: 100,
		YMin:      0,
		YMax:      100,
		ShowGrid:  true,
	})
	if len(canvas.calls) == 0 {
		t.Fatal("themed painter should produce draw calls")
	}
}

// ===== ListViewPainter tests =====

func TestListViewPainterImplementsInterface(t *testing.T) {
	var _ listview.Painter = devtools.ListViewPainter{}
}

func TestListViewPaintDivider(t *testing.T) {
	canvas := &recordCanvas{}
	painter := devtools.ListViewPainter{}
	painter.PaintDivider(canvas, listview.DividerState{
		Bounds:    geometry.NewRect(0, 24, 200, 1),
		ItemIndex: 0,
	})

	rects := canvas.methodCalls(methodDrawRect)
	if len(rects) != 1 {
		t.Errorf("divider should draw 1 DrawRect, got %d", len(rects))
	}
}

func TestListViewPaintDividerEmpty(t *testing.T) {
	canvas := &recordCanvas{}
	painter := devtools.ListViewPainter{}
	painter.PaintDivider(canvas, listview.DividerState{Bounds: geometry.Rect{}})
	if len(canvas.calls) != 0 {
		t.Errorf("empty bounds should produce no calls, got %d", len(canvas.calls))
	}
}

func TestListViewPaintEmptyState(t *testing.T) {
	canvas := &recordCanvas{}
	painter := devtools.ListViewPainter{}
	painter.PaintEmptyState(canvas, testBounds())

	texts := canvas.methodCalls(methodDrawText)
	if len(texts) != 1 {
		t.Fatalf("should draw 1 DrawText, got %d", len(texts))
	}
}

func TestListViewPaintItemBackground(t *testing.T) {
	canvas := &recordCanvas{}
	painter := devtools.ListViewPainter{}
	painter.PaintItemBackground(canvas, listview.ItemPaintState{
		Bounds:  testBounds(),
		Index:   0,
		Hovered: true,
	})

	rects := canvas.methodCalls(methodDrawRect)
	if len(rects) < 1 {
		t.Error("hovered item should draw at least 1 DrawRect (hover bg)")
	}
}

func TestListViewPaintItemBackgroundDisabled(t *testing.T) {
	canvas := &recordCanvas{}
	painter := devtools.ListViewPainter{}
	painter.PaintItemBackground(canvas, listview.ItemPaintState{
		Bounds:   testBounds(),
		Index:    0,
		Hovered:  true,
		Disabled: true,
	})

	rects := canvas.methodCalls(methodDrawRect)
	if len(rects) != 0 {
		t.Errorf("disabled hovered item should draw 0 DrawRect, got %d", len(rects))
	}
}

func TestListViewPaintSelection(t *testing.T) {
	canvas := &recordCanvas{}
	painter := devtools.ListViewPainter{}
	painter.PaintSelection(canvas, listview.ItemPaintState{
		Bounds:   testBounds(),
		Index:    0,
		Selected: true,
	})

	rects := canvas.methodCalls(methodDrawRect)
	if len(rects) < 1 {
		t.Error("selected item should draw at least 1 DrawRect (selection bg)")
	}
}

func TestListViewPaintSelectionFocused(t *testing.T) {
	canvas := &recordCanvas{}
	painter := devtools.ListViewPainter{}
	painter.PaintSelection(canvas, listview.ItemPaintState{
		Bounds:   testBounds(),
		Index:    0,
		Selected: true,
		Focused:  true,
	})

	strokes := canvas.methodCalls(methodStrokeRect)
	if len(strokes) < 1 {
		t.Error("focused selection should draw at least 1 StrokeRect (focus ring)")
	}
}

func TestListViewPaintSelectionNotSelected(t *testing.T) {
	canvas := &recordCanvas{}
	painter := devtools.ListViewPainter{}
	painter.PaintSelection(canvas, listview.ItemPaintState{
		Bounds:   testBounds(),
		Index:    0,
		Selected: false,
	})
	if len(canvas.calls) != 0 {
		t.Errorf("unselected should produce no calls, got %d", len(canvas.calls))
	}
}

func TestListViewPainterWithTheme(t *testing.T) {
	theme := devtools.NewDarkTheme()
	painter := devtools.ListViewPainter{Theme: theme}
	canvas := &recordCanvas{}

	painter.PaintDivider(canvas, listview.DividerState{
		Bounds:    geometry.NewRect(0, 24, 200, 1),
		ItemIndex: 0,
	})
	if len(canvas.calls) == 0 {
		t.Fatal("themed painter should produce draw calls")
	}
}

// ===== All enterprise painters with light theme =====

func TestAllEnterprisePaintersWithLightTheme(t *testing.T) {
	theme := devtools.NewTheme()

	tests := []struct {
		name string
		fn   func(*recordCanvas)
	}{
		{"TreeViewPainter", func(c *recordCanvas) {
			devtools.TreeViewPainter{Theme: theme}.PaintRowBackground(c, treeview.RowPaintState{
				Bounds: testBounds(), Hovered: true,
			})
		}},
		{"DataTablePainter", func(c *recordCanvas) {
			devtools.DataTablePainter{Theme: theme}.PaintHeader(c, testBounds(), datatable.HeaderPaintState{})
		}},
		{"ToolbarPainter", func(c *recordCanvas) {
			devtools.ToolbarPainter{Theme: theme}.PaintToolbar(c, toolbar.PaintToolbarState{Bounds: testBounds()})
		}},
		{"MenuPainter", func(c *recordCanvas) {
			devtools.MenuPainter{Theme: theme}.PaintMenuBar(c, &menu.MenuBarPaintState{
				Bounds: testBounds(), OpenIndex: -1, HoveredIndex: -1,
			})
		}},
		{"CollapsiblePainter", func(c *recordCanvas) {
			devtools.CollapsiblePainter{Theme: theme}.PaintHeader(c, collapsible.HeaderState{
				Title: "Light", Bounds: testBounds(),
			})
		}},
		{"ProgressPainter", func(c *recordCanvas) {
			devtools.ProgressPainter{Theme: theme}.PaintProgress(c, progress.PaintState{
				Value: 0.5, Bounds: geometry.NewRect(10, 10, 40, 40), Diameter: 36, StrokeWidth: 2,
			})
		}},
		{"SplitViewPainter", func(c *recordCanvas) {
			devtools.SplitViewPainter{Theme: theme}.PaintDivider(c, splitview.PaintState{
				DividerRect: geometry.NewRect(100, 0, 1, 400),
			})
		}},
		{"DockingPainter", func(c *recordCanvas) {
			devtools.DockingPainter{Theme: theme}.PaintZoneTabs(c, docking.ZoneTabsPaintState{
				TabBarBounds: geometry.NewRect(0, 0, 400, 24),
				Tabs: []docking.ZoneTabState{
					{Title: "Light", Bounds: geometry.NewRect(0, 0, 80, 24), Active: true},
				},
				ActiveIdx: 0,
			})
		}},
		{"PopoverPainter", func(c *recordCanvas) {
			devtools.PopoverPainter{Theme: theme}.PaintPopover(c, &popover.PopoverPaintState{
				Bounds: geometry.NewRect(100, 100, 200, 150),
			})
		}},
		{"LineChartPainter", func(c *recordCanvas) {
			devtools.LineChartPainter{Theme: theme}.PaintChart(c, geometry.NewRect(0, 0, 300, 200), linechart.PaintState{
				MaxPoints: 100, YMin: 0, YMax: 100, ShowGrid: true,
			})
		}},
		{"ListViewPainter", func(c *recordCanvas) {
			devtools.ListViewPainter{Theme: theme}.PaintDivider(c, listview.DividerState{
				Bounds: geometry.NewRect(0, 24, 200, 1), ItemIndex: 0,
			})
		}},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			canvas := &recordCanvas{}
			tt.fn(canvas)
			if len(canvas.calls) == 0 {
				t.Errorf("%s with light theme should produce draw calls", tt.name)
			}
		})
	}
}

// ===== NewPainters convenience function tests =====

func TestNewPaintersAllFieldsNonZero(t *testing.T) {
	dt := devtools.NewDarkTheme()
	p := devtools.NewPainters(dt)

	// Verify every painter has Theme set (non-nil).
	if p.Button.Theme != dt {
		t.Error("Button.Theme should be set")
	}
	if p.Checkbox.Theme != dt {
		t.Error("Checkbox.Theme should be set")
	}
	if p.Radio.Theme != dt {
		t.Error("Radio.Theme should be set")
	}
	if p.TextField.Theme != dt {
		t.Error("TextField.Theme should be set")
	}
	if p.Dropdown.Theme != dt {
		t.Error("Dropdown.Theme should be set")
	}
	if p.Slider.Theme != dt {
		t.Error("Slider.Theme should be set")
	}
	if p.Dialog.Theme != dt {
		t.Error("Dialog.Theme should be set")
	}
	if p.Scrollbar.Theme != dt {
		t.Error("Scrollbar.Theme should be set")
	}
	if p.TabView.Theme != dt {
		t.Error("TabView.Theme should be set")
	}
	if p.TreeView.Theme != dt {
		t.Error("TreeView.Theme should be set")
	}
	if p.DataTable.Theme != dt {
		t.Error("DataTable.Theme should be set")
	}
	if p.Toolbar.Theme != dt {
		t.Error("Toolbar.Theme should be set")
	}
	if p.Menu.Theme != dt {
		t.Error("Menu.Theme should be set")
	}
	if p.Collapsible.Theme != dt {
		t.Error("Collapsible.Theme should be set")
	}
	if p.Progress.Theme != dt {
		t.Error("Progress.Theme should be set")
	}
	if p.SplitView.Theme != dt {
		t.Error("SplitView.Theme should be set")
	}
	if p.Docking.Theme != dt {
		t.Error("Docking.Theme should be set")
	}
	if p.Popover.Theme != dt {
		t.Error("Popover.Theme should be set")
	}
	if p.LineChart.Theme != dt {
		t.Error("LineChart.Theme should be set")
	}
	if p.ListView.Theme != dt {
		t.Error("ListView.Theme should be set")
	}
}

func TestNewPaintersLightTheme(t *testing.T) {
	lt := devtools.NewTheme()
	p := devtools.NewPainters(lt)

	if p.Button.Theme != lt {
		t.Error("Button.Theme should be light theme")
	}
	if p.ListView.Theme != lt {
		t.Error("ListView.Theme should be light theme")
	}
}

func TestNewPaintersShareThemePointer(t *testing.T) {
	dt := devtools.NewDarkTheme()
	p := devtools.NewPainters(dt)

	// All painters share the same *Theme pointer.
	if p.Button.Theme != p.Checkbox.Theme {
		t.Error("Button and Checkbox should share the same Theme pointer")
	}
	if p.Slider.Theme != p.TreeView.Theme {
		t.Error("Slider and TreeView should share the same Theme pointer")
	}
}
