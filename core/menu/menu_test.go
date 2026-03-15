package menu_test

import (
	"image"
	"testing"

	"github.com/gogpu/ui/core/menu"
	"github.com/gogpu/ui/event"
	"github.com/gogpu/ui/geometry"
	"github.com/gogpu/ui/widget"
)

// =============================================================================
// MenuItem Tests
// =============================================================================

func TestItem_Basic(t *testing.T) {
	called := false
	item := menu.Item("Save", "Ctrl+S", func() { called = true })

	if item.Label != "Save" {
		t.Errorf("Label = %q, want %q", item.Label, "Save")
	}
	if item.Shortcut != "Ctrl+S" {
		t.Errorf("Shortcut = %q, want %q", item.Shortcut, "Ctrl+S")
	}
	if item.IsSeparator() {
		t.Error("regular item should not be separator")
	}
	if item.HasChildren() {
		t.Error("regular item should not have children")
	}
	if item.Disabled {
		t.Error("regular item should not be disabled")
	}
	item.OnAction()
	if !called {
		t.Error("OnAction should have been called")
	}
}

func TestItemDisabled(t *testing.T) {
	item := menu.ItemDisabled("Paste", "Ctrl+V")

	if item.Label != "Paste" {
		t.Errorf("Label = %q, want %q", item.Label, "Paste")
	}
	if !item.Disabled {
		t.Error("item should be disabled")
	}
	if item.OnAction != nil {
		t.Error("disabled item should have nil OnAction")
	}
}

func TestSep(t *testing.T) {
	sep := menu.Sep()

	if !sep.IsSeparator() {
		t.Error("Sep() should create a separator")
	}
	if sep.HasChildren() {
		t.Error("separator should not have children")
	}
}

func TestSubMenu(t *testing.T) {
	sub := menu.SubMenu("Export",
		menu.Item("PDF", "", nil),
		menu.Item("PNG", "", nil),
	)

	if sub.Label != "Export" {
		t.Errorf("Label = %q, want %q", sub.Label, "Export")
	}
	if !sub.HasChildren() {
		t.Error("submenu item should have children")
	}
	if len(sub.Children) != 2 {
		t.Errorf("len(Children) = %d, want 2", len(sub.Children))
	}
}

func TestBarMenu(t *testing.T) {
	tm := menu.BarMenu("File",
		menu.Item("New", "Ctrl+N", nil),
		menu.Sep(),
		menu.Item("Exit", "", nil),
	)

	if tm.Label != "File" {
		t.Errorf("Label = %q, want %q", tm.Label, "File")
	}
	if len(tm.Items) != 3 {
		t.Errorf("len(Items) = %d, want 3", len(tm.Items))
	}
}

// =============================================================================
// MenuBar Construction Tests
// =============================================================================

func TestNewBar_Defaults(t *testing.T) {
	bar := menu.NewBar([]menu.TopMenu{
		menu.BarMenu("File", menu.Item("New", "Ctrl+N", nil)),
		menu.BarMenu("Edit", menu.Item("Undo", "Ctrl+Z", nil)),
	})

	if !bar.IsVisible() {
		t.Error("default bar should be visible")
	}
	if !bar.IsEnabled() {
		t.Error("default bar should be enabled")
	}
	if !bar.IsFocusable() {
		t.Error("default bar should be focusable")
	}
	if bar.Children() != nil {
		t.Error("bar should have no children")
	}
	if bar.IsOpen() {
		t.Error("bar should not be open by default")
	}
	if bar.OpenIndex() != -1 {
		t.Errorf("OpenIndex = %d, want -1", bar.OpenIndex())
	}
	if bar.HoveredIndex() != -1 {
		t.Errorf("HoveredIndex = %d, want -1", bar.HoveredIndex())
	}
	if len(bar.Menus()) != 2 {
		t.Errorf("len(Menus) = %d, want 2", len(bar.Menus()))
	}
}

func TestNewBar_Empty(t *testing.T) {
	bar := menu.NewBar(nil)

	if bar.IsOpen() {
		t.Error("empty bar should not be open")
	}
}

func TestNewBar_WithPainter(t *testing.T) {
	p := &testPainter{}
	bar := menu.NewBar(
		[]menu.TopMenu{menu.BarMenu("File", menu.Item("New", "", nil))},
		menu.PainterOpt(p),
	)
	bar.SetBounds(geometry.NewRect(0, 0, 400, 32))
	ctx := widget.NewContext()

	_ = bar.Layout(ctx, geometry.Loose(geometry.Sz(400, 32)))
	bar.Draw(ctx, &mockCanvas{})

	if !p.barCalled {
		t.Error("Draw should delegate to configured painter's PaintMenuBar")
	}
}

func TestNewBar_FocusableRules(t *testing.T) {
	bar := menu.NewBar([]menu.TopMenu{menu.BarMenu("File")})

	if !bar.IsFocusable() {
		t.Error("visible+enabled bar should be focusable")
	}

	bar.SetVisible(false)
	if bar.IsFocusable() {
		t.Error("invisible bar should not be focusable")
	}

	bar.SetVisible(true)
	bar.SetEnabled(false)
	if bar.IsFocusable() {
		t.Error("disabled bar should not be focusable")
	}
}

// =============================================================================
// MenuBar Layout Tests
// =============================================================================

func TestBar_Layout(t *testing.T) {
	bar := menu.NewBar([]menu.TopMenu{
		menu.BarMenu("File"),
		menu.BarMenu("Edit"),
	})
	ctx := widget.NewContext()

	size := bar.Layout(ctx, geometry.Loose(geometry.Sz(800, 600)))

	if size.Width != 800 {
		t.Errorf("width = %v, want 800", size.Width)
	}
	if size.Height != 32 {
		t.Errorf("height = %v, want 32 (dfltBarHeight)", size.Height)
	}
}

func TestBar_Layout_TightConstraints(t *testing.T) {
	bar := menu.NewBar([]menu.TopMenu{menu.BarMenu("File")})
	ctx := widget.NewContext()

	size := bar.Layout(ctx, geometry.Tight(geometry.Sz(300, 24)))

	if size.Width != 300 {
		t.Errorf("width = %v, want 300", size.Width)
	}
	if size.Height != 24 {
		t.Errorf("height = %v, want 24", size.Height)
	}
}

// =============================================================================
// MenuBar Draw Tests
// =============================================================================

func TestBar_Draw_NilCanvas(t *testing.T) {
	bar := menu.NewBar([]menu.TopMenu{menu.BarMenu("File")})
	ctx := widget.NewContext()

	// Should not panic.
	bar.Draw(ctx, nil)
}

func TestBar_Draw_DelegatesToPainter(t *testing.T) {
	p := &testPainter{}
	bar := menu.NewBar(
		[]menu.TopMenu{menu.BarMenu("File", menu.Item("New", "", nil))},
		menu.PainterOpt(p),
	)
	bar.SetBounds(geometry.NewRect(0, 0, 400, 32))
	ctx := widget.NewContext()
	_ = bar.Layout(ctx, geometry.Loose(geometry.Sz(400, 32)))

	bar.Draw(ctx, &mockCanvas{})

	if !p.barCalled {
		t.Error("should delegate to PaintMenuBar")
	}
	if len(p.barState.Menus) != 1 {
		t.Errorf("len(Menus) = %d, want 1", len(p.barState.Menus))
	}
}

// =============================================================================
// MenuBar Mouse Events
// =============================================================================

func TestBar_MouseHover(t *testing.T) {
	bar := menu.NewBar([]menu.TopMenu{
		menu.BarMenu("File"),
		menu.BarMenu("Edit"),
	})
	bar.SetBounds(geometry.NewRect(0, 0, 400, 32))
	ctx := widget.NewContext()
	_ = bar.Layout(ctx, geometry.Loose(geometry.Sz(400, 32)))

	// Hover over first label.
	move := event.NewMouseEvent(event.MouseMove, event.ButtonNone, 0,
		geometry.Pt(24, 16), geometry.Pt(24, 16), event.ModNone)
	consumed := bar.Event(ctx, move)

	if !consumed {
		t.Error("mouse move should be consumed within bar bounds")
	}
	if bar.HoveredIndex() != 0 {
		t.Errorf("HoveredIndex = %d, want 0", bar.HoveredIndex())
	}
}

func TestBar_MouseLeave(t *testing.T) {
	bar := menu.NewBar([]menu.TopMenu{menu.BarMenu("File")})
	bar.SetBounds(geometry.NewRect(0, 0, 400, 32))
	ctx := widget.NewContext()
	_ = bar.Layout(ctx, geometry.Loose(geometry.Sz(400, 32)))

	// Hover then leave.
	move := event.NewMouseEvent(event.MouseMove, event.ButtonNone, 0,
		geometry.Pt(24, 16), geometry.Pt(24, 16), event.ModNone)
	bar.Event(ctx, move)

	leave := event.NewMouseEvent(event.MouseLeave, event.ButtonNone, 0,
		geometry.Pt(24, 16), geometry.Pt(24, 16), event.ModNone)
	bar.Event(ctx, leave)

	if bar.HoveredIndex() != -1 {
		t.Errorf("HoveredIndex = %d, want -1 after leave", bar.HoveredIndex())
	}
}

func TestBar_ClickOpensMenu(t *testing.T) {
	bar := menu.NewBar([]menu.TopMenu{
		menu.BarMenu("File", menu.Item("New", "", nil)),
	})
	bar.SetBounds(geometry.NewRect(0, 0, 400, 32))
	ctx := widget.NewContext()
	om := setupOverlayManager(ctx)
	_ = bar.Layout(ctx, geometry.Loose(geometry.Sz(400, 32)))

	press := event.NewMouseEvent(event.MousePress, event.ButtonLeft, event.ButtonStateLeft,
		geometry.Pt(24, 16), geometry.Pt(24, 16), event.ModNone)
	consumed := bar.Event(ctx, press)

	if !consumed {
		t.Error("click should be consumed")
	}
	if !bar.IsOpen() {
		t.Error("click should open menu")
	}
	if bar.OpenIndex() != 0 {
		t.Errorf("OpenIndex = %d, want 0", bar.OpenIndex())
	}
	if om.pushCount != 1 {
		t.Errorf("PushOverlay called %d times, want 1", om.pushCount)
	}
}

func TestBar_ClickTogglesMenu(t *testing.T) {
	bar := menu.NewBar([]menu.TopMenu{
		menu.BarMenu("File", menu.Item("New", "", nil)),
	})
	bar.SetBounds(geometry.NewRect(0, 0, 400, 32))
	ctx := widget.NewContext()
	setupOverlayManager(ctx)
	_ = bar.Layout(ctx, geometry.Loose(geometry.Sz(400, 32)))

	// First click opens.
	press := event.NewMouseEvent(event.MousePress, event.ButtonLeft, event.ButtonStateLeft,
		geometry.Pt(24, 16), geometry.Pt(24, 16), event.ModNone)
	bar.Event(ctx, press)

	if !bar.IsOpen() {
		t.Fatal("first click should open menu")
	}

	// Second click closes.
	bar.Event(ctx, press)

	if bar.IsOpen() {
		t.Error("second click should close menu")
	}
}

func TestBar_HoverSwitchesMenu(t *testing.T) {
	bar := menu.NewBar([]menu.TopMenu{
		menu.BarMenu("File", menu.Item("New", "", nil)),
		menu.BarMenu("Edit", menu.Item("Undo", "", nil)),
	})
	bar.SetBounds(geometry.NewRect(0, 0, 400, 32))
	ctx := widget.NewContext()
	om := setupOverlayManager(ctx)
	_ = bar.Layout(ctx, geometry.Loose(geometry.Sz(400, 32)))

	// Open File menu.
	press := event.NewMouseEvent(event.MousePress, event.ButtonLeft, event.ButtonStateLeft,
		geometry.Pt(24, 16), geometry.Pt(24, 16), event.ModNone)
	bar.Event(ctx, press)

	if bar.OpenIndex() != 0 {
		t.Fatalf("OpenIndex = %d, want 0", bar.OpenIndex())
	}

	// Hover over Edit label — should switch.
	// Edit label starts after File's rect. Let's compute a position inside Edit.
	editX := float32(80) // approximate position
	move := event.NewMouseEvent(event.MouseMove, event.ButtonNone, 0,
		geometry.Pt(editX, 16), geometry.Pt(editX, 16), event.ModNone)
	bar.Event(ctx, move)

	if bar.OpenIndex() != 1 {
		t.Errorf("OpenIndex = %d, want 1 after hovering Edit", bar.OpenIndex())
	}
	// Should have closed old menu and opened new one.
	if om.removeCount < 1 {
		t.Error("should have removed old overlay")
	}
}

func TestBar_ClickOutsideBounds(t *testing.T) {
	bar := menu.NewBar([]menu.TopMenu{menu.BarMenu("File")})
	bar.SetBounds(geometry.NewRect(0, 0, 400, 32))
	ctx := widget.NewContext()

	press := event.NewMouseEvent(event.MousePress, event.ButtonLeft, event.ButtonStateLeft,
		geometry.Pt(0, 100), geometry.Pt(0, 100), event.ModNone)
	consumed := bar.Event(ctx, press)

	if consumed {
		t.Error("click outside bounds should not be consumed")
	}
}

func TestBar_RightClickIgnored(t *testing.T) {
	bar := menu.NewBar([]menu.TopMenu{
		menu.BarMenu("File", menu.Item("New", "", nil)),
	})
	bar.SetBounds(geometry.NewRect(0, 0, 400, 32))
	ctx := widget.NewContext()
	_ = bar.Layout(ctx, geometry.Loose(geometry.Sz(400, 32)))

	press := event.NewMouseEvent(event.MousePress, event.ButtonRight, event.ButtonStateRight,
		geometry.Pt(24, 16), geometry.Pt(24, 16), event.ModNone)
	consumed := bar.Event(ctx, press)

	if consumed {
		t.Error("right click should not be consumed")
	}
}

// =============================================================================
// MenuBar Keyboard Events
// =============================================================================

func TestBar_KeyDownOpensMenu(t *testing.T) {
	bar := menu.NewBar([]menu.TopMenu{
		menu.BarMenu("File", menu.Item("New", "", nil)),
	})
	bar.SetBounds(geometry.NewRect(0, 0, 400, 32))
	bar.SetFocused(true)
	ctx := widget.NewContext()
	setupOverlayManager(ctx)
	_ = bar.Layout(ctx, geometry.Loose(geometry.Sz(400, 32)))

	down := event.NewKeyEvent(event.KeyPress, event.KeyDown, 0, event.ModNone)
	consumed := bar.Event(ctx, down)

	if !consumed {
		t.Error("Down should be consumed by focused bar")
	}
	if !bar.IsOpen() {
		t.Error("Down should open menu")
	}
}

func TestBar_KeyLeftRight(t *testing.T) {
	bar := menu.NewBar([]menu.TopMenu{
		menu.BarMenu("File"),
		menu.BarMenu("Edit"),
		menu.BarMenu("View"),
	})
	bar.SetBounds(geometry.NewRect(0, 0, 400, 32))
	bar.SetFocused(true)
	ctx := widget.NewContext()
	setupOverlayManager(ctx)
	_ = bar.Layout(ctx, geometry.Loose(geometry.Sz(400, 32)))

	// Open first.
	down := event.NewKeyEvent(event.KeyPress, event.KeyDown, 0, event.ModNone)
	bar.Event(ctx, down)

	// Right arrow.
	right := event.NewKeyEvent(event.KeyPress, event.KeyRight, 0, event.ModNone)
	bar.Event(ctx, right)

	if bar.HoveredIndex() != 1 {
		t.Errorf("HoveredIndex = %d, want 1 after Right", bar.HoveredIndex())
	}

	// Left arrow.
	left := event.NewKeyEvent(event.KeyPress, event.KeyLeft, 0, event.ModNone)
	bar.Event(ctx, left)

	if bar.HoveredIndex() != 0 {
		t.Errorf("HoveredIndex = %d, want 0 after Left", bar.HoveredIndex())
	}
}

func TestBar_KeyLeftWraps(t *testing.T) {
	bar := menu.NewBar([]menu.TopMenu{
		menu.BarMenu("File"),
		menu.BarMenu("Edit"),
	})
	bar.SetBounds(geometry.NewRect(0, 0, 400, 32))
	bar.SetFocused(true)
	ctx := widget.NewContext()
	setupOverlayManager(ctx)
	_ = bar.Layout(ctx, geometry.Loose(geometry.Sz(400, 32)))

	// Open first menu.
	down := event.NewKeyEvent(event.KeyPress, event.KeyDown, 0, event.ModNone)
	bar.Event(ctx, down)

	// Left from first should wrap to last.
	left := event.NewKeyEvent(event.KeyPress, event.KeyLeft, 0, event.ModNone)
	bar.Event(ctx, left)

	if bar.HoveredIndex() != 1 {
		t.Errorf("HoveredIndex = %d, want 1 (wrapped)", bar.HoveredIndex())
	}
}

func TestBar_KeyEscapeCloses(t *testing.T) {
	bar := menu.NewBar([]menu.TopMenu{
		menu.BarMenu("File", menu.Item("New", "", nil)),
	})
	bar.SetBounds(geometry.NewRect(0, 0, 400, 32))
	bar.SetFocused(true)
	ctx := widget.NewContext()
	setupOverlayManager(ctx)
	_ = bar.Layout(ctx, geometry.Loose(geometry.Sz(400, 32)))

	// Open then escape.
	down := event.NewKeyEvent(event.KeyPress, event.KeyDown, 0, event.ModNone)
	bar.Event(ctx, down)

	if !bar.IsOpen() {
		t.Fatal("should be open")
	}

	esc := event.NewKeyEvent(event.KeyPress, event.KeyEscape, 0, event.ModNone)
	consumed := bar.Event(ctx, esc)

	if !consumed {
		t.Error("Escape should be consumed")
	}
	if bar.IsOpen() {
		t.Error("Escape should close menu")
	}
}

func TestBar_KeyEscapeNotConsumedWhenClosed(t *testing.T) {
	bar := menu.NewBar([]menu.TopMenu{menu.BarMenu("File")})
	bar.SetFocused(true)
	ctx := widget.NewContext()

	esc := event.NewKeyEvent(event.KeyPress, event.KeyEscape, 0, event.ModNone)
	consumed := bar.Event(ctx, esc)

	if consumed {
		t.Error("Escape should not be consumed when no menu is open")
	}
}

func TestBar_KeyIgnoredWhenNotFocused(t *testing.T) {
	bar := menu.NewBar([]menu.TopMenu{menu.BarMenu("File", menu.Item("New", "", nil))})
	ctx := widget.NewContext()

	down := event.NewKeyEvent(event.KeyPress, event.KeyDown, 0, event.ModNone)
	consumed := bar.Event(ctx, down)

	if consumed {
		t.Error("key events should be ignored when not focused")
	}
}

func TestBar_KeyReleaseIgnored(t *testing.T) {
	bar := menu.NewBar([]menu.TopMenu{menu.BarMenu("File")})
	bar.SetFocused(true)
	ctx := widget.NewContext()

	release := event.NewKeyEvent(event.KeyRelease, event.KeyDown, 0, event.ModNone)
	consumed := bar.Event(ctx, release)

	if consumed {
		t.Error("key release should not be consumed")
	}
}

// =============================================================================
// MenuBar Close Tests
// =============================================================================

func TestBar_Close(t *testing.T) {
	bar := menu.NewBar([]menu.TopMenu{
		menu.BarMenu("File", menu.Item("New", "", nil)),
	})
	bar.SetBounds(geometry.NewRect(0, 0, 400, 32))
	ctx := widget.NewContext()
	om := setupOverlayManager(ctx)
	_ = bar.Layout(ctx, geometry.Loose(geometry.Sz(400, 32)))

	// Open and close.
	press := event.NewMouseEvent(event.MousePress, event.ButtonLeft, event.ButtonStateLeft,
		geometry.Pt(24, 16), geometry.Pt(24, 16), event.ModNone)
	bar.Event(ctx, press)
	bar.Close(ctx)

	if bar.IsOpen() {
		t.Error("bar should be closed after Close()")
	}
	if om.removeCount < 1 {
		t.Error("RemoveOverlay should have been called")
	}
}

func TestBar_CloseNoOpWhenClosed(t *testing.T) {
	bar := menu.NewBar([]menu.TopMenu{menu.BarMenu("File")})
	ctx := widget.NewContext()

	// Should not panic.
	bar.Close(ctx)
}

func TestBar_OpenWithoutOverlayManager(t *testing.T) {
	bar := menu.NewBar([]menu.TopMenu{
		menu.BarMenu("File", menu.Item("New", "", nil)),
	})
	bar.SetBounds(geometry.NewRect(0, 0, 400, 32))
	bar.SetFocused(true)
	ctx := widget.NewContext()
	// No overlay manager set.
	_ = bar.Layout(ctx, geometry.Loose(geometry.Sz(400, 32)))

	down := event.NewKeyEvent(event.KeyPress, event.KeyDown, 0, event.ModNone)
	bar.Event(ctx, down)

	if bar.IsOpen() {
		t.Error("should not open without overlay manager")
	}
}

// =============================================================================
// MenuBar Accessibility Tests
// =============================================================================

func TestBar_A11y(t *testing.T) {
	bar := menu.NewBar([]menu.TopMenu{menu.BarMenu("File")})

	if bar.A11yRole() != "menubar" {
		t.Errorf("A11yRole = %q, want %q", bar.A11yRole(), "menubar")
	}
	if bar.A11yLabel() != "menu bar" {
		t.Errorf("A11yLabel = %q, want %q", bar.A11yLabel(), "menu bar")
	}
}

// =============================================================================
// ContextMenu Tests
// =============================================================================

func TestNewContextMenu_Defaults(t *testing.T) {
	cm := menu.NewContextMenu([]menu.MenuItem{
		menu.Item("Cut", "Ctrl+X", nil),
		menu.Item("Copy", "Ctrl+C", nil),
	})

	if cm.IsOpen() {
		t.Error("context menu should not be open by default")
	}
	if len(cm.Items()) != 2 {
		t.Errorf("len(Items) = %d, want 2", len(cm.Items()))
	}
}

func TestContextMenu_Show(t *testing.T) {
	called := false
	cm := menu.NewContextMenu([]menu.MenuItem{
		menu.Item("Cut", "Ctrl+X", func() { called = true }),
	})

	ctx := widget.NewContext()
	om := setupOverlayManager(ctx)

	cm.Show(ctx, geometry.Pt(100, 200))

	if !cm.IsOpen() {
		t.Error("context menu should be open after Show()")
	}
	if om.pushCount != 1 {
		t.Errorf("PushOverlay called %d times, want 1", om.pushCount)
	}

	_ = called // action not invoked yet
}

func TestContextMenu_ShowClosesExisting(t *testing.T) {
	cm := menu.NewContextMenu([]menu.MenuItem{
		menu.Item("Cut", "", nil),
	})
	ctx := widget.NewContext()
	om := setupOverlayManager(ctx)

	cm.Show(ctx, geometry.Pt(100, 200))
	cm.Show(ctx, geometry.Pt(200, 300)) // re-open at new position

	if !cm.IsOpen() {
		t.Error("should be open")
	}
	if om.pushCount != 2 {
		t.Errorf("PushOverlay called %d times, want 2", om.pushCount)
	}
	if om.removeCount < 1 {
		t.Error("first overlay should have been removed")
	}
}

func TestContextMenu_Hide(t *testing.T) {
	cm := menu.NewContextMenu([]menu.MenuItem{
		menu.Item("Cut", "", nil),
	})
	ctx := widget.NewContext()
	om := setupOverlayManager(ctx)

	cm.Show(ctx, geometry.Pt(100, 200))
	cm.Hide(ctx)

	if cm.IsOpen() {
		t.Error("context menu should be closed after Hide()")
	}
	if om.removeCount != 1 {
		t.Errorf("RemoveOverlay called %d times, want 1", om.removeCount)
	}
}

func TestContextMenu_HideNoOpWhenClosed(t *testing.T) {
	cm := menu.NewContextMenu([]menu.MenuItem{
		menu.Item("Cut", "", nil),
	})
	ctx := widget.NewContext()

	// Should not panic.
	cm.Hide(ctx)
}

func TestContextMenu_ShowWithoutOverlayManager(t *testing.T) {
	cm := menu.NewContextMenu([]menu.MenuItem{
		menu.Item("Cut", "", nil),
	})
	ctx := widget.NewContext()
	// No overlay manager.

	cm.Show(ctx, geometry.Pt(100, 200))

	if cm.IsOpen() {
		t.Error("should not open without overlay manager")
	}
}

func TestContextMenu_WithPainter(t *testing.T) {
	p := &testPainter{}
	cm := menu.NewContextMenu(
		[]menu.MenuItem{menu.Item("Cut", "", nil)},
		menu.ContextPainterOpt(p),
	)
	ctx := widget.NewContext()
	setupOverlayManager(ctx)

	cm.Show(ctx, geometry.Pt(100, 200))

	// The panel should use our painter. Draw it.
	panel := cm.Panel()
	if panel == nil {
		t.Fatal("panel should not be nil when open")
	}
	panel.Draw(ctx, &mockCanvas{})

	if !p.menuCalled {
		t.Error("Draw should delegate to configured painter's PaintMenu")
	}
}

func TestContextMenu_Panel(t *testing.T) {
	cm := menu.NewContextMenu([]menu.MenuItem{
		menu.Item("Cut", "", nil),
	})

	if cm.Panel() != nil {
		t.Error("Panel() should be nil when closed")
	}

	ctx := widget.NewContext()
	setupOverlayManager(ctx)
	cm.Show(ctx, geometry.Pt(100, 200))

	if cm.Panel() == nil {
		t.Error("Panel() should not be nil when open")
	}

	cm.Hide(ctx)

	if cm.Panel() != nil {
		t.Error("Panel() should be nil after Hide()")
	}
}

// =============================================================================
// MenuPanel Tests (via ContextMenu)
// =============================================================================

func TestMenuPanel_KeyboardNavigation(t *testing.T) {
	cm := menu.NewContextMenu([]menu.MenuItem{
		menu.Item("Cut", "", nil),
		menu.Sep(),
		menu.Item("Copy", "", nil),
		menu.ItemDisabled("Paste", ""),
		menu.Item("Delete", "", nil),
	})
	ctx := widget.NewContext()
	setupOverlayManager(ctx)
	cm.Show(ctx, geometry.Pt(0, 0))

	panel := cm.Panel()
	if panel == nil {
		t.Fatal("panel should not be nil")
	}

	// Down should highlight first item.
	down := event.NewKeyEvent(event.KeyPress, event.KeyDown, 0, event.ModNone)
	panel.Event(ctx, down)

	if panel.HighlightedIndex() != 0 {
		t.Errorf("HighlightedIndex = %d, want 0", panel.HighlightedIndex())
	}

	// Down again should skip separator, land on "Copy" (index 2).
	panel.Event(ctx, down)

	if panel.HighlightedIndex() != 2 {
		t.Errorf("HighlightedIndex = %d, want 2 (skip separator)", panel.HighlightedIndex())
	}

	// Down again should skip disabled "Paste" (index 3), land on "Delete" (index 4).
	panel.Event(ctx, down)

	if panel.HighlightedIndex() != 4 {
		t.Errorf("HighlightedIndex = %d, want 4 (skip disabled)", panel.HighlightedIndex())
	}

	// Up should skip disabled, land on Copy (index 2).
	up := event.NewKeyEvent(event.KeyPress, event.KeyUp, 0, event.ModNone)
	panel.Event(ctx, up)

	if panel.HighlightedIndex() != 2 {
		t.Errorf("HighlightedIndex = %d, want 2 after Up", panel.HighlightedIndex())
	}
}

func TestMenuPanel_EnterActivates(t *testing.T) {
	called := false
	cm := menu.NewContextMenu([]menu.MenuItem{
		menu.Item("Save", "", func() { called = true }),
	})
	ctx := widget.NewContext()
	setupOverlayManager(ctx)
	cm.Show(ctx, geometry.Pt(0, 0))

	panel := cm.Panel()
	if panel == nil {
		t.Fatal("panel should not be nil")
	}

	// Navigate to first item.
	down := event.NewKeyEvent(event.KeyPress, event.KeyDown, 0, event.ModNone)
	panel.Event(ctx, down)

	// Activate.
	enter := event.NewKeyEvent(event.KeyPress, event.KeyEnter, 0, event.ModNone)
	panel.Event(ctx, enter)

	if !called {
		t.Error("Enter should trigger OnAction")
	}
	if cm.IsOpen() {
		t.Error("menu should close after activation")
	}
}

func TestMenuPanel_EnterOnDisabledItem(t *testing.T) {
	saveCalled := false
	openCalled := false
	cm := menu.NewContextMenu([]menu.MenuItem{
		menu.ItemDisabled("Save", ""),
		menu.Item("Open", "", func() { openCalled = true }),
	})
	ctx := widget.NewContext()
	setupOverlayManager(ctx)
	cm.Show(ctx, geometry.Pt(0, 0))

	panel := cm.Panel()
	if panel == nil {
		t.Fatal("panel should not be nil")
	}

	// Down should skip disabled "Save" and highlight "Open" (index 1).
	down := event.NewKeyEvent(event.KeyPress, event.KeyDown, 0, event.ModNone)
	panel.Event(ctx, down)

	if panel.HighlightedIndex() != 1 {
		t.Errorf("HighlightedIndex = %d, want 1 (skip disabled)", panel.HighlightedIndex())
	}

	// Enter activates "Open".
	enter := event.NewKeyEvent(event.KeyPress, event.KeyEnter, 0, event.ModNone)
	panel.Event(ctx, enter)

	if saveCalled {
		t.Error("disabled Save should not be called")
	}
	if !openCalled {
		t.Error("Open should be activated")
	}
}

func TestMenuPanel_EscapeCloses(t *testing.T) {
	cm := menu.NewContextMenu([]menu.MenuItem{
		menu.Item("Cut", "", nil),
	})
	ctx := widget.NewContext()
	setupOverlayManager(ctx)
	cm.Show(ctx, geometry.Pt(0, 0))

	panel := cm.Panel()
	if panel == nil {
		t.Fatal("panel should not be nil")
	}

	esc := event.NewKeyEvent(event.KeyPress, event.KeyEscape, 0, event.ModNone)
	consumed := panel.Event(ctx, esc)

	if !consumed {
		t.Error("Escape should be consumed")
	}
	if cm.IsOpen() {
		t.Error("Escape should close context menu")
	}
}

func TestMenuPanel_MouseHover(t *testing.T) {
	cm := menu.NewContextMenu([]menu.MenuItem{
		menu.Item("Cut", "", nil),
		menu.Item("Copy", "", nil),
	})
	ctx := widget.NewContext()
	setupOverlayManager(ctx)
	cm.Show(ctx, geometry.Pt(0, 0))

	panel := cm.Panel()
	if panel == nil {
		t.Fatal("panel should not be nil")
	}

	bounds := panel.Bounds()

	// Mouse move over first item.
	move := event.NewMouseEvent(event.MouseMove, event.ButtonNone, 0,
		geometry.Pt(bounds.Min.X+10, bounds.Min.Y+10),
		geometry.Pt(bounds.Min.X+10, bounds.Min.Y+10), event.ModNone)
	consumed := panel.Event(ctx, move)

	if !consumed {
		t.Error("mouse move should be consumed within bounds")
	}
	if panel.HighlightedIndex() != 0 {
		t.Errorf("HighlightedIndex = %d, want 0", panel.HighlightedIndex())
	}
}

func TestMenuPanel_MouseClick(t *testing.T) {
	called := false
	cm := menu.NewContextMenu([]menu.MenuItem{
		menu.Item("Save", "", func() { called = true }),
	})
	ctx := widget.NewContext()
	setupOverlayManager(ctx)
	cm.Show(ctx, geometry.Pt(0, 0))

	panel := cm.Panel()
	if panel == nil {
		t.Fatal("panel should not be nil")
	}

	bounds := panel.Bounds()

	// Click first item.
	press := event.NewMouseEvent(event.MousePress, event.ButtonLeft, event.ButtonStateLeft,
		geometry.Pt(bounds.Min.X+10, bounds.Min.Y+10),
		geometry.Pt(bounds.Min.X+10, bounds.Min.Y+10), event.ModNone)
	panel.Event(ctx, press)

	if !called {
		t.Error("click should trigger OnAction")
	}
}

func TestMenuPanel_MouseClickDisabled(t *testing.T) {
	called := false
	cm := menu.NewContextMenu([]menu.MenuItem{
		menu.ItemDisabled("Save", ""),
		menu.Item("Open", "", func() { called = true }),
	})
	ctx := widget.NewContext()
	setupOverlayManager(ctx)
	cm.Show(ctx, geometry.Pt(0, 0))

	panel := cm.Panel()
	if panel == nil {
		t.Fatal("panel should not be nil")
	}

	bounds := panel.Bounds()

	// Click first (disabled) item.
	press := event.NewMouseEvent(event.MousePress, event.ButtonLeft, event.ButtonStateLeft,
		geometry.Pt(bounds.Min.X+10, bounds.Min.Y+10),
		geometry.Pt(bounds.Min.X+10, bounds.Min.Y+10), event.ModNone)
	panel.Event(ctx, press)

	if called {
		t.Error("should not activate disabled item via click")
	}
}

func TestMenuPanel_MouseOutsideBounds(t *testing.T) {
	cm := menu.NewContextMenu([]menu.MenuItem{
		menu.Item("Cut", "", nil),
	})
	ctx := widget.NewContext()
	setupOverlayManager(ctx)
	cm.Show(ctx, geometry.Pt(100, 100))

	panel := cm.Panel()
	if panel == nil {
		t.Fatal("panel should not be nil")
	}

	// Click outside bounds.
	press := event.NewMouseEvent(event.MousePress, event.ButtonLeft, event.ButtonStateLeft,
		geometry.Pt(0, 0), geometry.Pt(0, 0), event.ModNone)
	consumed := panel.Event(ctx, press)

	if consumed {
		t.Error("click outside bounds should not be consumed")
	}
}

func TestMenuPanel_Draw_NilCanvas(t *testing.T) {
	cm := menu.NewContextMenu([]menu.MenuItem{
		menu.Item("Cut", "", nil),
	})
	ctx := widget.NewContext()
	setupOverlayManager(ctx)
	cm.Show(ctx, geometry.Pt(0, 0))

	panel := cm.Panel()
	if panel == nil {
		t.Fatal("panel should not be nil")
	}

	// Should not panic.
	panel.Draw(ctx, nil)
}

func TestMenuPanel_Children(t *testing.T) {
	cm := menu.NewContextMenu([]menu.MenuItem{
		menu.Item("Cut", "", nil),
	})
	ctx := widget.NewContext()
	setupOverlayManager(ctx)
	cm.Show(ctx, geometry.Pt(0, 0))

	panel := cm.Panel()
	if panel == nil {
		t.Fatal("panel should not be nil")
	}

	if panel.Children() != nil {
		t.Error("menu panel should have no children")
	}
}

func TestMenuPanel_KeyReleaseIgnored(t *testing.T) {
	cm := menu.NewContextMenu([]menu.MenuItem{
		menu.Item("Cut", "", nil),
	})
	ctx := widget.NewContext()
	setupOverlayManager(ctx)
	cm.Show(ctx, geometry.Pt(0, 0))

	panel := cm.Panel()
	if panel == nil {
		t.Fatal("panel should not be nil")
	}

	release := event.NewKeyEvent(event.KeyRelease, event.KeyDown, 0, event.ModNone)
	consumed := panel.Event(ctx, release)

	if consumed {
		t.Error("key release should not be consumed")
	}
}

func TestMenuPanel_RightClickIgnored(t *testing.T) {
	cm := menu.NewContextMenu([]menu.MenuItem{
		menu.Item("Cut", "", nil),
	})
	ctx := widget.NewContext()
	setupOverlayManager(ctx)
	cm.Show(ctx, geometry.Pt(0, 0))

	panel := cm.Panel()
	if panel == nil {
		t.Fatal("panel should not be nil")
	}

	bounds := panel.Bounds()
	press := event.NewMouseEvent(event.MousePress, event.ButtonRight, event.ButtonStateRight,
		geometry.Pt(bounds.Min.X+10, bounds.Min.Y+10),
		geometry.Pt(bounds.Min.X+10, bounds.Min.Y+10), event.ModNone)
	consumed := panel.Event(ctx, press)

	if consumed {
		t.Error("right click should not be consumed by menu panel")
	}
}

// =============================================================================
// Submenu Tests
// =============================================================================

func TestSubmenu_HoverOpens(t *testing.T) {
	cm := menu.NewContextMenu([]menu.MenuItem{
		menu.SubMenu("Export",
			menu.Item("PDF", "", nil),
			menu.Item("PNG", "", nil),
		),
		menu.Item("Close", "", nil),
	})
	ctx := widget.NewContext()
	setupOverlayManager(ctx)
	cm.Show(ctx, geometry.Pt(0, 0))

	panel := cm.Panel()
	if panel == nil {
		t.Fatal("panel should not be nil")
	}

	bounds := panel.Bounds()

	// Hover over "Export" item (first item).
	move := event.NewMouseEvent(event.MouseMove, event.ButtonNone, 0,
		geometry.Pt(bounds.Min.X+10, bounds.Min.Y+10),
		geometry.Pt(bounds.Min.X+10, bounds.Min.Y+10), event.ModNone)
	panel.Event(ctx, move)

	if panel.SubMenuIndex() != 0 {
		t.Errorf("SubMenuIndex = %d, want 0", panel.SubMenuIndex())
	}
	if panel.SubMenuPanel() == nil {
		t.Error("SubMenuPanel should not be nil")
	}
}

func TestSubmenu_KeyRightOpens(t *testing.T) {
	cm := menu.NewContextMenu([]menu.MenuItem{
		menu.SubMenu("Export",
			menu.Item("PDF", "", nil),
		),
	})
	ctx := widget.NewContext()
	setupOverlayManager(ctx)
	cm.Show(ctx, geometry.Pt(0, 0))

	panel := cm.Panel()
	if panel == nil {
		t.Fatal("panel should not be nil")
	}

	// Navigate to submenu item.
	down := event.NewKeyEvent(event.KeyPress, event.KeyDown, 0, event.ModNone)
	panel.Event(ctx, down)

	// Right arrow opens submenu.
	right := event.NewKeyEvent(event.KeyPress, event.KeyRight, 0, event.ModNone)
	consumed := panel.Event(ctx, right)

	if !consumed {
		t.Error("Right should be consumed when submenu opens")
	}
	if panel.SubMenuPanel() == nil {
		t.Error("submenu should be open")
	}
}

func TestSubmenu_KeyLeftCloses(t *testing.T) {
	cm := menu.NewContextMenu([]menu.MenuItem{
		menu.SubMenu("Export",
			menu.Item("PDF", "", nil),
		),
	})
	ctx := widget.NewContext()
	setupOverlayManager(ctx)
	cm.Show(ctx, geometry.Pt(0, 0))

	panel := cm.Panel()
	if panel == nil {
		t.Fatal("panel should not be nil")
	}

	// Navigate and open submenu.
	down := event.NewKeyEvent(event.KeyPress, event.KeyDown, 0, event.ModNone)
	panel.Event(ctx, down)

	right := event.NewKeyEvent(event.KeyPress, event.KeyRight, 0, event.ModNone)
	panel.Event(ctx, right)

	if panel.SubMenuPanel() == nil {
		t.Fatal("submenu should be open")
	}

	// Left arrow closes submenu.
	left := event.NewKeyEvent(event.KeyPress, event.KeyLeft, 0, event.ModNone)
	consumed := panel.Event(ctx, left)

	if !consumed {
		t.Error("Left should be consumed")
	}
	if panel.SubMenuPanel() != nil {
		t.Error("submenu should be closed after Left")
	}
}

func TestSubmenu_EnterActivatesChild(t *testing.T) {
	pdfCalled := false
	cm := menu.NewContextMenu([]menu.MenuItem{
		menu.SubMenu("Export",
			menu.Item("PDF", "", func() { pdfCalled = true }),
		),
	})
	ctx := widget.NewContext()
	setupOverlayManager(ctx)
	cm.Show(ctx, geometry.Pt(0, 0))

	panel := cm.Panel()
	if panel == nil {
		t.Fatal("panel should not be nil")
	}

	// Navigate to submenu and open.
	down := event.NewKeyEvent(event.KeyPress, event.KeyDown, 0, event.ModNone)
	panel.Event(ctx, down)

	enter := event.NewKeyEvent(event.KeyPress, event.KeyEnter, 0, event.ModNone)
	panel.Event(ctx, enter) // opens submenu

	sub := panel.SubMenuPanel()
	if sub == nil {
		t.Fatal("submenu should be open")
	}

	// Navigate inside submenu.
	sub.Event(ctx, down)

	// Activate.
	sub.Event(ctx, enter)

	if !pdfCalled {
		t.Error("Enter in submenu should trigger OnAction")
	}
}

func TestSubmenu_HoverAwayCloses(t *testing.T) {
	cm := menu.NewContextMenu([]menu.MenuItem{
		menu.SubMenu("Export",
			menu.Item("PDF", "", nil),
		),
		menu.Item("Close", "", nil),
	})
	ctx := widget.NewContext()
	setupOverlayManager(ctx)
	cm.Show(ctx, geometry.Pt(0, 0))

	panel := cm.Panel()
	if panel == nil {
		t.Fatal("panel should not be nil")
	}

	bounds := panel.Bounds()

	// Hover over Export (opens submenu).
	move1 := event.NewMouseEvent(event.MouseMove, event.ButtonNone, 0,
		geometry.Pt(bounds.Min.X+10, bounds.Min.Y+10),
		geometry.Pt(bounds.Min.X+10, bounds.Min.Y+10), event.ModNone)
	panel.Event(ctx, move1)

	if panel.SubMenuPanel() == nil {
		t.Fatal("submenu should be open")
	}

	// Hover over Close (closes submenu).
	move2 := event.NewMouseEvent(event.MouseMove, event.ButtonNone, 0,
		geometry.Pt(bounds.Min.X+10, bounds.Min.Y+50),
		geometry.Pt(bounds.Min.X+10, bounds.Min.Y+50), event.ModNone)
	panel.Event(ctx, move2)

	if panel.SubMenuPanel() != nil {
		t.Error("submenu should be closed after hovering different item")
	}
}

// =============================================================================
// DefaultPainter Tests
// =============================================================================

func TestDefaultPainter_PaintMenuBar_EmptyBounds(t *testing.T) {
	p := menu.DefaultPainter{}
	canvas := &recordingCanvas{}

	// Should not panic.
	p.PaintMenuBar(canvas, &menu.MenuBarPaintState{})
}

func TestDefaultPainter_PaintMenu_EmptyBounds(t *testing.T) {
	p := menu.DefaultPainter{}
	canvas := &recordingCanvas{}

	// Should not panic.
	p.PaintMenu(canvas, &menu.MenuPaintState{})
}

func TestDefaultPainter_PaintMenuBar_DrawsLabels(t *testing.T) {
	p := menu.DefaultPainter{}
	canvas := &recordingCanvas{}

	bounds := geometry.NewRect(0, 0, 400, 32)
	labelRect := geometry.NewRect(0, 0, 60, 32)

	p.PaintMenuBar(canvas, &menu.MenuBarPaintState{
		Bounds:       bounds,
		Menus:        []menu.TopMenu{{Label: "File"}},
		MenuRects:    []geometry.Rect{labelRect},
		OpenIndex:    -1,
		HoveredIndex: -1,
	})

	if len(canvas.drawTexts) == 0 {
		t.Error("should have drawn at least one text label")
	}
}

func TestDefaultPainter_PaintMenu_DrawsItems(t *testing.T) {
	p := menu.DefaultPainter{}
	canvas := &recordingCanvas{}

	bounds := geometry.NewRect(0, 0, 200, 100)

	p.PaintMenu(canvas, &menu.MenuPaintState{
		Bounds: bounds,
		Items: []menu.MenuItem{
			menu.Item("Cut", "Ctrl+X", nil),
			menu.Sep(),
			menu.Item("Copy", "Ctrl+C", nil),
		},
		HighlightedIndex: 0,
		ItemHeight:       36,
		SeparatorHeight:  9,
		SubMenuOpenIndex: -1,
	})

	if len(canvas.drawTexts) < 2 {
		t.Errorf("should have drawn labels; got %d text calls", len(canvas.drawTexts))
	}
}

func TestDefaultPainter_PaintMenu_Submenu(t *testing.T) {
	p := menu.DefaultPainter{}
	canvas := &recordingCanvas{}

	bounds := geometry.NewRect(0, 0, 200, 100)

	p.PaintMenu(canvas, &menu.MenuPaintState{
		Bounds: bounds,
		Items: []menu.MenuItem{
			menu.SubMenu("Export", menu.Item("PDF", "", nil)),
		},
		HighlightedIndex: -1,
		ItemHeight:       36,
		SeparatorHeight:  9,
		SubMenuOpenIndex: -1,
	})

	// Should draw submenu arrow.
	found := false
	for _, dt := range canvas.drawTexts {
		if dt.text == ">" {
			found = true
			break
		}
	}
	if !found {
		t.Error("should draw submenu arrow '>'")
	}
}

func TestDefaultPainter_PaintMenu_DisabledItem(t *testing.T) {
	p := menu.DefaultPainter{}
	canvas := &recordingCanvas{}

	bounds := geometry.NewRect(0, 0, 200, 100)

	p.PaintMenu(canvas, &menu.MenuPaintState{
		Bounds: bounds,
		Items: []menu.MenuItem{
			menu.ItemDisabled("Paste", "Ctrl+V"),
		},
		HighlightedIndex: -1,
		ItemHeight:       36,
		SeparatorHeight:  9,
		SubMenuOpenIndex: -1,
	})

	if len(canvas.drawTexts) == 0 {
		t.Error("should draw text for disabled item")
	}
}

// =============================================================================
// Widget Interface Compliance
// =============================================================================

func TestBarWidgetInterface(t *testing.T) {
	var w widget.Widget = menu.NewBar(nil)
	_ = w
}

func TestBarFocusableInterface(t *testing.T) {
	var f widget.Focusable = menu.NewBar(nil)
	_ = f
}

// =============================================================================
// Edge Cases
// =============================================================================

func TestBar_EmptyMenus(t *testing.T) {
	bar := menu.NewBar(nil)
	bar.SetBounds(geometry.NewRect(0, 0, 400, 32))
	bar.SetFocused(true)
	ctx := widget.NewContext()
	setupOverlayManager(ctx)
	_ = bar.Layout(ctx, geometry.Loose(geometry.Sz(400, 32)))

	// Keyboard navigation with no menus should not panic.
	right := event.NewKeyEvent(event.KeyPress, event.KeyRight, 0, event.ModNone)
	bar.Event(ctx, right)

	left := event.NewKeyEvent(event.KeyPress, event.KeyLeft, 0, event.ModNone)
	bar.Event(ctx, left)
}

func TestMenuPanel_EmptyItems(t *testing.T) {
	cm := menu.NewContextMenu(nil)
	ctx := widget.NewContext()
	setupOverlayManager(ctx)
	cm.Show(ctx, geometry.Pt(0, 0))

	panel := cm.Panel()
	if panel == nil {
		t.Fatal("panel should not be nil")
	}

	// Navigation with no items should not panic.
	down := event.NewKeyEvent(event.KeyPress, event.KeyDown, 0, event.ModNone)
	panel.Event(ctx, down)
}

func TestMenuPanel_AllDisabledItems(t *testing.T) {
	cm := menu.NewContextMenu([]menu.MenuItem{
		menu.ItemDisabled("Cut", ""),
		menu.ItemDisabled("Copy", ""),
	})
	ctx := widget.NewContext()
	setupOverlayManager(ctx)
	cm.Show(ctx, geometry.Pt(0, 0))

	panel := cm.Panel()
	if panel == nil {
		t.Fatal("panel should not be nil")
	}

	// Down should stay at -1 (no enabled items).
	down := event.NewKeyEvent(event.KeyPress, event.KeyDown, 0, event.ModNone)
	panel.Event(ctx, down)

	if panel.HighlightedIndex() != -1 {
		t.Errorf("HighlightedIndex = %d, want -1 (all disabled)", panel.HighlightedIndex())
	}
}

func TestMenuPanel_AllSeparators(t *testing.T) {
	cm := menu.NewContextMenu([]menu.MenuItem{
		menu.Sep(),
		menu.Sep(),
	})
	ctx := widget.NewContext()
	setupOverlayManager(ctx)
	cm.Show(ctx, geometry.Pt(0, 0))

	panel := cm.Panel()
	if panel == nil {
		t.Fatal("panel should not be nil")
	}

	down := event.NewKeyEvent(event.KeyPress, event.KeyDown, 0, event.ModNone)
	panel.Event(ctx, down)

	if panel.HighlightedIndex() != -1 {
		t.Errorf("HighlightedIndex = %d, want -1 (all separators)", panel.HighlightedIndex())
	}
}

func TestMenuPanel_UnhandledEvent(t *testing.T) {
	cm := menu.NewContextMenu([]menu.MenuItem{
		menu.Item("Cut", "", nil),
	})
	ctx := widget.NewContext()
	setupOverlayManager(ctx)
	cm.Show(ctx, geometry.Pt(0, 0))

	panel := cm.Panel()
	if panel == nil {
		t.Fatal("panel should not be nil")
	}

	// Focus event should not be consumed.
	fe := &event.FocusEvent{}
	consumed := panel.Event(ctx, fe)

	if consumed {
		t.Error("unhandled event type should not be consumed")
	}
}

func TestBar_UnhandledEvent(t *testing.T) {
	bar := menu.NewBar([]menu.TopMenu{menu.BarMenu("File")})
	ctx := widget.NewContext()

	// Focus event should not be consumed.
	fe := &event.FocusEvent{}
	consumed := bar.Event(ctx, fe)

	if consumed {
		t.Error("unhandled event type should not be consumed")
	}
}

func TestSubmenu_RightOnNonSubmenuItem(t *testing.T) {
	cm := menu.NewContextMenu([]menu.MenuItem{
		menu.Item("Cut", "", nil),
	})
	ctx := widget.NewContext()
	setupOverlayManager(ctx)
	cm.Show(ctx, geometry.Pt(0, 0))

	panel := cm.Panel()
	if panel == nil {
		t.Fatal("panel should not be nil")
	}

	down := event.NewKeyEvent(event.KeyPress, event.KeyDown, 0, event.ModNone)
	panel.Event(ctx, down)

	right := event.NewKeyEvent(event.KeyPress, event.KeyRight, 0, event.ModNone)
	consumed := panel.Event(ctx, right)

	if consumed {
		t.Error("Right on non-submenu item should not be consumed")
	}
}

func TestMenuPanel_SpaceActivates(t *testing.T) {
	called := false
	cm := menu.NewContextMenu([]menu.MenuItem{
		menu.Item("Save", "", func() { called = true }),
	})
	ctx := widget.NewContext()
	setupOverlayManager(ctx)
	cm.Show(ctx, geometry.Pt(0, 0))

	panel := cm.Panel()
	if panel == nil {
		t.Fatal("panel should not be nil")
	}

	down := event.NewKeyEvent(event.KeyPress, event.KeyDown, 0, event.ModNone)
	panel.Event(ctx, down)

	space := event.NewKeyEvent(event.KeyPress, event.KeySpace, 0, event.ModNone)
	panel.Event(ctx, space)

	if !called {
		t.Error("Space should trigger OnAction")
	}
}

// =============================================================================
// Additional Coverage Tests
// =============================================================================

func TestBar_KeyEnterOpensFromHover(t *testing.T) {
	bar := menu.NewBar([]menu.TopMenu{
		menu.BarMenu("File", menu.Item("New", "", nil)),
		menu.BarMenu("Edit", menu.Item("Undo", "", nil)),
	})
	bar.SetBounds(geometry.NewRect(0, 0, 400, 32))
	bar.SetFocused(true)
	ctx := widget.NewContext()
	setupOverlayManager(ctx)
	_ = bar.Layout(ctx, geometry.Loose(geometry.Sz(400, 32)))

	// Hover over a label first.
	move := event.NewMouseEvent(event.MouseMove, event.ButtonNone, 0,
		geometry.Pt(24, 16), geometry.Pt(24, 16), event.ModNone)
	bar.Event(ctx, move)

	// Enter should open the hovered menu.
	enter := event.NewKeyEvent(event.KeyPress, event.KeyEnter, 0, event.ModNone)
	consumed := bar.Event(ctx, enter)

	if !consumed {
		t.Error("Enter should be consumed when label is hovered")
	}
	if !bar.IsOpen() {
		t.Error("Enter should open hovered menu")
	}
}

func TestBar_KeySpaceOpensMenu(t *testing.T) {
	bar := menu.NewBar([]menu.TopMenu{
		menu.BarMenu("File", menu.Item("New", "", nil)),
	})
	bar.SetBounds(geometry.NewRect(0, 0, 400, 32))
	bar.SetFocused(true)
	ctx := widget.NewContext()
	setupOverlayManager(ctx)
	_ = bar.Layout(ctx, geometry.Loose(geometry.Sz(400, 32)))

	// Space should open first menu (fallback when no hover).
	space := event.NewKeyEvent(event.KeyPress, event.KeySpace, 0, event.ModNone)
	consumed := bar.Event(ctx, space)

	if !consumed {
		t.Error("Space should be consumed")
	}
	if !bar.IsOpen() {
		t.Error("Space should open menu")
	}
}

func TestDefaultPainter_PaintMenuBar_HoveredAndOpen(t *testing.T) {
	p := menu.DefaultPainter{}
	canvas := &recordingCanvas{}

	bounds := geometry.NewRect(0, 0, 400, 32)
	r1 := geometry.NewRect(0, 0, 60, 32)
	r2 := geometry.NewRect(60, 0, 120, 32)

	// Test with open index = 0, hovered index = 1.
	p.PaintMenuBar(canvas, &menu.MenuBarPaintState{
		Bounds:       bounds,
		Menus:        []menu.TopMenu{{Label: "File"}, {Label: "Edit"}},
		MenuRects:    []geometry.Rect{r1, r2},
		OpenIndex:    0,
		HoveredIndex: 1,
	})

	// Should have drawn rects for both labels.
	if len(canvas.drawRects) < 3 { // bar bg + border + active/hover items
		t.Errorf("expected draw calls for open and hovered states; got %d drawRect calls", len(canvas.drawRects))
	}
}

func TestDefaultPainter_PaintMenuBar_MoreMenusThanRects(t *testing.T) {
	p := menu.DefaultPainter{}
	canvas := &recordingCanvas{}

	bounds := geometry.NewRect(0, 0, 400, 32)

	// More menus than rects should not panic.
	p.PaintMenuBar(canvas, &menu.MenuBarPaintState{
		Bounds:    bounds,
		Menus:     []menu.TopMenu{{Label: "File"}, {Label: "Edit"}},
		MenuRects: []geometry.Rect{geometry.NewRect(0, 0, 60, 32)}, // only one rect
		OpenIndex: -1,
	})
}

func TestMenuPanel_ActivateWithNoHighlight(t *testing.T) {
	cm := menu.NewContextMenu([]menu.MenuItem{
		menu.Item("Cut", "", nil),
	})
	ctx := widget.NewContext()
	setupOverlayManager(ctx)
	cm.Show(ctx, geometry.Pt(0, 0))

	panel := cm.Panel()
	if panel == nil {
		t.Fatal("panel should not be nil")
	}

	// Enter without highlighting should not panic.
	enter := event.NewKeyEvent(event.KeyPress, event.KeyEnter, 0, event.ModNone)
	consumed := panel.Event(ctx, enter)

	if consumed {
		t.Error("Enter with no highlight should not be consumed")
	}
}

func TestMenuPanel_LeftWithoutSubmenuAndParent(t *testing.T) {
	cm := menu.NewContextMenu([]menu.MenuItem{
		menu.Item("Cut", "", nil),
	})
	ctx := widget.NewContext()
	setupOverlayManager(ctx)
	cm.Show(ctx, geometry.Pt(0, 0))

	panel := cm.Panel()
	if panel == nil {
		t.Fatal("panel should not be nil")
	}

	// Left key with no submenu should trigger onClose (which hides the menu).
	left := event.NewKeyEvent(event.KeyPress, event.KeyLeft, 0, event.ModNone)
	consumed := panel.Event(ctx, left)

	if !consumed {
		t.Error("Left should be consumed (triggers onClose)")
	}
}

func TestMenuPanel_UnhandledKey(t *testing.T) {
	cm := menu.NewContextMenu([]menu.MenuItem{
		menu.Item("Cut", "", nil),
	})
	ctx := widget.NewContext()
	setupOverlayManager(ctx)
	cm.Show(ctx, geometry.Pt(0, 0))

	panel := cm.Panel()
	if panel == nil {
		t.Fatal("panel should not be nil")
	}

	// Tab key should not be consumed.
	tab := event.NewKeyEvent(event.KeyPress, event.KeyTab, 0, event.ModNone)
	consumed := panel.Event(ctx, tab)

	if consumed {
		t.Error("Tab should not be consumed")
	}
}

func TestBar_UnhandledKey(t *testing.T) {
	bar := menu.NewBar([]menu.TopMenu{menu.BarMenu("File")})
	bar.SetFocused(true)
	ctx := widget.NewContext()

	tab := event.NewKeyEvent(event.KeyPress, event.KeyTab, 0, event.ModNone)
	consumed := bar.Event(ctx, tab)

	if consumed {
		t.Error("Tab should not be consumed by bar")
	}
}

func TestMenuPanel_ClickOnSeparator(t *testing.T) {
	called := false
	cm := menu.NewContextMenu([]menu.MenuItem{
		menu.Sep(),
		menu.Item("Cut", "", func() { called = true }),
	})
	ctx := widget.NewContext()
	setupOverlayManager(ctx)
	cm.Show(ctx, geometry.Pt(0, 0))

	panel := cm.Panel()
	if panel == nil {
		t.Fatal("panel should not be nil")
	}

	bounds := panel.Bounds()

	// Click on separator area.
	press := event.NewMouseEvent(event.MousePress, event.ButtonLeft, event.ButtonStateLeft,
		geometry.Pt(bounds.Min.X+10, bounds.Min.Y+6),
		geometry.Pt(bounds.Min.X+10, bounds.Min.Y+6), event.ModNone)
	panel.Event(ctx, press)

	if called {
		t.Error("clicking separator should not trigger action")
	}
}

func TestDefaultPainter_PaintMenu_DisabledSubmenu(t *testing.T) {
	p := menu.DefaultPainter{}
	canvas := &recordingCanvas{}

	bounds := geometry.NewRect(0, 0, 200, 100)

	disabled := menu.SubMenu("Export", menu.Item("PDF", "", nil))
	disabled.Disabled = true

	p.PaintMenu(canvas, &menu.MenuPaintState{
		Bounds:           bounds,
		Items:            []menu.MenuItem{disabled},
		HighlightedIndex: -1,
		ItemHeight:       36,
		SeparatorHeight:  9,
		SubMenuOpenIndex: -1,
	})

	if len(canvas.drawTexts) == 0 {
		t.Error("should draw text for disabled submenu")
	}
}

func TestBar_KeyRightWraps(t *testing.T) {
	bar := menu.NewBar([]menu.TopMenu{
		menu.BarMenu("File"),
		menu.BarMenu("Edit"),
	})
	bar.SetBounds(geometry.NewRect(0, 0, 400, 32))
	bar.SetFocused(true)
	ctx := widget.NewContext()
	setupOverlayManager(ctx)
	_ = bar.Layout(ctx, geometry.Loose(geometry.Sz(400, 32)))

	// Open Edit menu (index 1).
	// Navigate to Edit first.
	right := event.NewKeyEvent(event.KeyPress, event.KeyRight, 0, event.ModNone)
	bar.Event(ctx, right) // hoveredIndex = 0
	bar.Event(ctx, right) // hoveredIndex = 1

	// Right again should wrap to 0.
	bar.Event(ctx, right)

	if bar.HoveredIndex() != 0 {
		t.Errorf("HoveredIndex = %d, want 0 (wrapped right)", bar.HoveredIndex())
	}
}

func TestBar_MouseMoveNoChange(t *testing.T) {
	bar := menu.NewBar([]menu.TopMenu{
		menu.BarMenu("File"),
	})
	bar.SetBounds(geometry.NewRect(0, 0, 400, 32))
	ctx := widget.NewContext()
	_ = bar.Layout(ctx, geometry.Loose(geometry.Sz(400, 32)))

	// Move twice to same position should not cause issues.
	pos := geometry.Pt(24, 16)
	move := event.NewMouseEvent(event.MouseMove, event.ButtonNone, 0, pos, pos, event.ModNone)
	bar.Event(ctx, move)
	bar.Event(ctx, move) // same index

	if bar.HoveredIndex() != 0 {
		t.Errorf("HoveredIndex = %d, want 0", bar.HoveredIndex())
	}
}

func TestSubmenu_OpenSameIndexNoOp(t *testing.T) {
	cm := menu.NewContextMenu([]menu.MenuItem{
		menu.SubMenu("Export",
			menu.Item("PDF", "", nil),
		),
	})
	ctx := widget.NewContext()
	setupOverlayManager(ctx)
	cm.Show(ctx, geometry.Pt(0, 0))

	panel := cm.Panel()
	if panel == nil {
		t.Fatal("panel should not be nil")
	}

	bounds := panel.Bounds()
	pos := geometry.Pt(bounds.Min.X+10, bounds.Min.Y+10)

	// Hover over submenu twice -- second hover should be no-op.
	move := event.NewMouseEvent(event.MouseMove, event.ButtonNone, 0, pos, pos, event.ModNone)
	panel.Event(ctx, move)

	sub1 := panel.SubMenuPanel()
	if sub1 == nil {
		t.Fatal("submenu should be open")
	}

	// Hover same position again.
	panel.Event(ctx, move)

	sub2 := panel.SubMenuPanel()
	if sub2 != sub1 {
		t.Error("hovering same submenu item should not recreate submenu")
	}
}

// =============================================================================
// Helpers
// =============================================================================

func setupOverlayManager(ctx *widget.ContextImpl) *mockOverlayManager {
	om := &mockOverlayManager{}
	ctx.SetOverlayManager(om)
	ctx.SetWindowSize(geometry.Sz(800, 600))
	return om
}

// --- Mock Overlay Manager ---

type mockOverlayManager struct {
	pushCount   int
	popCount    int
	removeCount int
	lastWidget  widget.Widget
}

func (m *mockOverlayManager) PushOverlay(w widget.Widget, _ func()) {
	m.pushCount++
	m.lastWidget = w
}

func (m *mockOverlayManager) PopOverlay() {
	m.popCount++
}

func (m *mockOverlayManager) RemoveOverlay(_ widget.Widget) {
	m.removeCount++
}

// --- Test Painter ---

type testPainter struct {
	barCalled  bool
	menuCalled bool
	barState   menu.MenuBarPaintState
	menuState  menu.MenuPaintState
}

func (p *testPainter) PaintMenuBar(_ widget.Canvas, st *menu.MenuBarPaintState) {
	p.barCalled = true
	p.barState = *st
}

func (p *testPainter) PaintMenu(_ widget.Canvas, st *menu.MenuPaintState) {
	p.menuCalled = true
	p.menuState = *st
}

// --- Recording Canvas ---

type recordingCanvas struct {
	drawTexts      []drawTextCall
	drawRoundRects []drawRoundRectCall
	drawRects      []drawRectCall
}

type drawTextCall struct {
	text     string
	bounds   geometry.Rect
	fontSize float32
	color    widget.Color
	bold     bool
	align    widget.TextAlign
}

type drawRoundRectCall struct {
	r      geometry.Rect
	color  widget.Color
	radius float32
}

type drawRectCall struct {
	r     geometry.Rect
	color widget.Color
}

func (c *recordingCanvas) Clear(_ widget.Color) {}

func (c *recordingCanvas) DrawRect(r geometry.Rect, color widget.Color) {
	c.drawRects = append(c.drawRects, drawRectCall{r: r, color: color})
}

func (c *recordingCanvas) StrokeRect(_ geometry.Rect, _ widget.Color, _ float32) {}

func (c *recordingCanvas) DrawRoundRect(r geometry.Rect, color widget.Color, radius float32) {
	c.drawRoundRects = append(c.drawRoundRects, drawRoundRectCall{r: r, color: color, radius: radius})
}

func (c *recordingCanvas) StrokeRoundRect(_ geometry.Rect, _ widget.Color, _ float32, _ float32) {}
func (c *recordingCanvas) DrawCircle(_ geometry.Point, _ float32, _ widget.Color)                {}
func (c *recordingCanvas) StrokeCircle(_ geometry.Point, _ float32, _ widget.Color, _ float32)   {}
func (c *recordingCanvas) DrawLine(_, _ geometry.Point, _ widget.Color, _ float32)               {}

func (c *recordingCanvas) DrawText(text string, bounds geometry.Rect, fontSize float32, color widget.Color, bold bool, align widget.TextAlign) {
	c.drawTexts = append(c.drawTexts, drawTextCall{text: text, bounds: bounds, fontSize: fontSize, color: color, bold: bold, align: align})
}

func (c *recordingCanvas) MeasureText(text string, fontSize float32, _ bool) float32 {
	return float32(len([]rune(text))) * fontSize * 0.5
}

func (c *recordingCanvas) DrawImage(_ image.Image, _ geometry.Point)    {}
func (c *recordingCanvas) PushClip(_ geometry.Rect)                     {}
func (c *recordingCanvas) PushClipRoundRect(_ geometry.Rect, _ float32) {}
func (c *recordingCanvas) PopClip()                                     {}
func (c *recordingCanvas) PushTransform(_ geometry.Point)               {}
func (c *recordingCanvas) PopTransform()                                {}
func (c *recordingCanvas) TransformOffset() geometry.Point              { return geometry.Point{} }

// --- Mock Canvas ---

type mockCanvas struct{}

func (c *mockCanvas) Clear(_ widget.Color)                                                  {}
func (c *mockCanvas) DrawRect(_ geometry.Rect, _ widget.Color)                              {}
func (c *mockCanvas) StrokeRect(_ geometry.Rect, _ widget.Color, _ float32)                 {}
func (c *mockCanvas) DrawRoundRect(_ geometry.Rect, _ widget.Color, _ float32)              {}
func (c *mockCanvas) StrokeRoundRect(_ geometry.Rect, _ widget.Color, _ float32, _ float32) {}
func (c *mockCanvas) DrawCircle(_ geometry.Point, _ float32, _ widget.Color)                {}
func (c *mockCanvas) StrokeCircle(_ geometry.Point, _ float32, _ widget.Color, _ float32)   {}
func (c *mockCanvas) DrawLine(_, _ geometry.Point, _ widget.Color, _ float32)               {}

func (c *mockCanvas) DrawText(_ string, _ geometry.Rect, _ float32, _ widget.Color, _ bool, _ widget.TextAlign) {
}

func (c *mockCanvas) MeasureText(text string, fontSize float32, _ bool) float32 {
	return float32(len([]rune(text))) * fontSize * 0.5
}

func (c *mockCanvas) DrawImage(_ image.Image, _ geometry.Point)    {}
func (c *mockCanvas) PushClip(_ geometry.Rect)                     {}
func (c *mockCanvas) PushClipRoundRect(_ geometry.Rect, _ float32) {}
func (c *mockCanvas) PopClip()                                     {}
func (c *mockCanvas) PushTransform(_ geometry.Point)               {}
func (c *mockCanvas) PopTransform()                                {}
func (c *mockCanvas) TransformOffset() geometry.Point              { return geometry.Point{} }
