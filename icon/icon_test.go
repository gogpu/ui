package icon

import (
	"image"
	"testing"

	"github.com/gogpu/ui/a11y"
	"github.com/gogpu/ui/event"
	"github.com/gogpu/ui/geometry"
	"github.com/gogpu/ui/state"
	"github.com/gogpu/ui/widget"
)

// --- Command tests ---

func TestCommand_String(t *testing.T) {
	tests := []struct {
		cmd  Command
		want string
	}{
		{CmdMoveTo, "MoveTo"},
		{CmdLineTo, "LineTo"},
		{CmdCubicTo, "CubicTo"},
		{CmdQuadraticTo, "QuadraticTo"},
		{CmdClose, "Close"},
		{Command(99), unknownStr},
	}
	for _, tt := range tests {
		if got := tt.cmd.String(); got != tt.want {
			t.Errorf("Command(%d).String() = %q, want %q", tt.cmd, got, tt.want)
		}
	}
}

// --- PathOp constructor tests ---

func TestMove(t *testing.T) {
	op := Move(10, 20)
	if op.Cmd != CmdMoveTo {
		t.Fatalf("Move().Cmd = %v, want CmdMoveTo", op.Cmd)
	}
	if op.Params[0] != 10 || op.Params[1] != 20 {
		t.Errorf("Move(10,20).Params = %v, want [10,20,...]", op.Params)
	}
}

func TestLine(t *testing.T) {
	op := Line(5, 15)
	if op.Cmd != CmdLineTo {
		t.Fatalf("Line().Cmd = %v, want CmdLineTo", op.Cmd)
	}
	if op.Params[0] != 5 || op.Params[1] != 15 {
		t.Errorf("Line(5,15).Params = %v, want [5,15,...]", op.Params)
	}
}

func TestCubic(t *testing.T) {
	op := Cubic(1, 2, 3, 4, 5, 6)
	if op.Cmd != CmdCubicTo {
		t.Fatalf("Cubic().Cmd = %v, want CmdCubicTo", op.Cmd)
	}
	for i, want := range []float32{1, 2, 3, 4, 5, 6} {
		if op.Params[i] != want {
			t.Errorf("Cubic().Params[%d] = %v, want %v", i, op.Params[i], want)
		}
	}
}

func TestClosePath(t *testing.T) {
	op := ClosePath()
	if op.Cmd != CmdClose {
		t.Errorf("ClosePath().Cmd = %v, want CmdClose", op.Cmd)
	}
}

// --- IconData tests ---

func TestIconData_IsValueType(t *testing.T) {
	original := IconData{
		Name:    "test",
		ViewBox: 24,
		Ops:     []PathOp{Move(0, 0), Line(10, 10)},
	}
	copied := original
	copied.Name = "modified"
	if original.Name == copied.Name {
		t.Error("IconData should be a value type; modifying copy changed original")
	}
}

// --- Built-in icons tests ---

func TestBuiltinIcons_HaveRequiredFields(t *testing.T) {
	builtins := []struct {
		name string
		icon IconData
	}{
		{"Close", Close},
		{"Check", Check},
		{"ChevronDown", ChevronDown},
		{"ChevronRight", ChevronRight},
		{"Search", Search},
		{"Settings", Settings},
		{"Menu", Menu},
		{"ArrowBack", ArrowBack},
		{"Add", Add},
		{"Delete", Delete},
		{"Play", Play},
		{"Stop", Stop},
		{"Pause", Pause},
		{"Debug", Debug},
		{"Gear", Gear},
		{"Filter", Filter},
		{"FolderOpen", FolderOpen},
		{"FolderClosed", FolderClosed},
		{"Terminal", Terminal},
		{"Refresh", Refresh},
		{"Plus", Plus},
		{"Minus", Minus},
	}
	for _, tt := range builtins {
		t.Run(tt.name, func(t *testing.T) {
			if tt.icon.Name == "" {
				t.Error("icon Name is empty")
			}
			if tt.icon.ViewBox <= 0 {
				t.Errorf("icon ViewBox = %v, want > 0", tt.icon.ViewBox)
			}
			if len(tt.icon.Ops) == 0 {
				t.Error("icon Ops is empty")
			}
			// First op should be a MoveTo
			if tt.icon.Ops[0].Cmd != CmdMoveTo {
				t.Errorf("icon first op = %v, want CmdMoveTo", tt.icon.Ops[0].Cmd)
			}
		})
	}
}

func TestBuiltinIcons_Count(t *testing.T) {
	// Ensure we have exactly 22 built-in single-color icons
	count := 22
	builtins := []IconData{
		Close, Check, ChevronDown, ChevronRight, Search, Settings, Menu, ArrowBack, Add, Delete,
		Play, Stop, Pause, Debug, Gear, Filter, FolderOpen, FolderClosed, Terminal, Refresh, Plus, Minus,
	}
	if len(builtins) != count {
		t.Errorf("expected %d built-in icons, got %d", count, len(builtins))
	}
}

// --- Registry tests ---

func TestNewRegistry_Empty(t *testing.T) {
	r := NewRegistry()
	if r.Len() != 0 {
		t.Errorf("NewRegistry().Len() = %d, want 0", r.Len())
	}
}

func TestRegistry_RegisterAndGet(t *testing.T) {
	r := NewRegistry()
	ic := IconData{Name: "test", ViewBox: 24, Ops: []PathOp{Move(0, 0)}}
	r.Register(ic)

	got, ok := r.Get("test")
	if !ok {
		t.Fatal("Get(test) returned false")
	}
	if got.Name != "test" {
		t.Errorf("Get(test).Name = %q, want %q", got.Name, "test")
	}
}

func TestRegistry_Get_NotFound(t *testing.T) {
	r := NewRegistry()
	_, ok := r.Get("nonexistent")
	if ok {
		t.Error("Get(nonexistent) should return false")
	}
}

func TestRegistry_Register_Overwrite(t *testing.T) {
	r := NewRegistry()
	r.Register(IconData{Name: "ic", ViewBox: 24, Ops: []PathOp{Move(0, 0)}})
	r.Register(IconData{Name: "ic", ViewBox: 48, Ops: []PathOp{Move(1, 1)}})

	got, ok := r.Get("ic")
	if !ok {
		t.Fatal("Get(ic) returned false")
	}
	if got.ViewBox != 48 {
		t.Errorf("overwritten icon ViewBox = %v, want 48", got.ViewBox)
	}
}

func TestRegistry_Names(t *testing.T) {
	r := NewRegistry()
	r.Register(IconData{Name: "alpha", ViewBox: 24})
	r.Register(IconData{Name: "beta", ViewBox: 24})

	names := r.Names()
	if len(names) != 2 {
		t.Fatalf("Names() returned %d items, want 2", len(names))
	}
	// Names are unordered, just check both exist
	found := map[string]bool{}
	for _, n := range names {
		found[n] = true
	}
	if !found["alpha"] || !found["beta"] {
		t.Errorf("Names() = %v, expected alpha and beta", names)
	}
}

func TestRegistry_Len(t *testing.T) {
	r := NewRegistry()
	if r.Len() != 0 {
		t.Errorf("empty registry Len = %d", r.Len())
	}
	r.Register(IconData{Name: "a", ViewBox: 24})
	r.Register(IconData{Name: "b", ViewBox: 24})
	if r.Len() != 2 {
		t.Errorf("registry Len = %d, want 2", r.Len())
	}
}

func TestDefaultRegistry_ContainsBuiltins(t *testing.T) {
	r := DefaultRegistry()
	if r.Len() < 22 {
		t.Errorf("DefaultRegistry().Len() = %d, want >= 22", r.Len())
	}
	expected := []string{
		"close", "check", "chevron_down", "chevron_right",
		"search", "settings", "menu", "arrow_back", "add", "delete",
		"play", "stop", "pause", "debug", "gear", "filter",
		"folder_open", "folder_closed", "terminal", "refresh", "plus", "minus",
	}
	for _, name := range expected {
		if _, ok := r.Get(name); !ok {
			t.Errorf("DefaultRegistry missing built-in icon %q", name)
		}
	}
}

// --- Draw tests ---

// mockCanvas records DrawLine calls for testing.
type mockCanvas struct {
	lines []mockLine
}

type mockLine struct {
	from, to geometry.Point
	color    widget.Color
	width    float32
}

func (m *mockCanvas) DrawLine(from, to geometry.Point, color widget.Color, strokeWidth float32) {
	m.lines = append(m.lines, mockLine{from: from, to: to, color: color, width: strokeWidth})
}

// Stub methods required by the Canvas interface.
func (m *mockCanvas) Clear(widget.Color)                                            {}
func (m *mockCanvas) DrawRect(geometry.Rect, widget.Color)                          {}
func (m *mockCanvas) StrokeRect(geometry.Rect, widget.Color, float32)               {}
func (m *mockCanvas) DrawRoundRect(geometry.Rect, widget.Color, float32)            {}
func (m *mockCanvas) StrokeRoundRect(geometry.Rect, widget.Color, float32, float32) {}
func (m *mockCanvas) DrawCircle(geometry.Point, float32, widget.Color)              {}
func (m *mockCanvas) StrokeCircle(geometry.Point, float32, widget.Color, float32)   {}
func (m *mockCanvas) DrawText(string, geometry.Rect, float32, widget.Color, bool, widget.TextAlign) {
}

func (m *mockCanvas) MeasureText(text string, fontSize float32, _ bool) float32 {
	return float32(len([]rune(text))) * fontSize * 0.5
}
func (m *mockCanvas) DrawImage(_ image.Image, _ geometry.Point) {}
func (m *mockCanvas) PushClip(geometry.Rect)                    {}
func (m *mockCanvas) PushClipRoundRect(geometry.Rect, float32)  {}
func (m *mockCanvas) PopClip()                                  {}
func (m *mockCanvas) PushTransform(geometry.Point)              {}
func (m *mockCanvas) PopTransform()                             {}
func (m *mockCanvas) TransformOffset() geometry.Point           { return geometry.Point{} }

func TestDraw_EmptyOps(t *testing.T) {
	c := &mockCanvas{}
	Draw(c, IconData{ViewBox: 24}, geometry.NewRect(0, 0, 24, 24), widget.ColorBlack)
	if len(c.lines) != 0 {
		t.Errorf("Draw with empty ops produced %d lines", len(c.lines))
	}
}

func TestDraw_ZeroViewBox(t *testing.T) {
	c := &mockCanvas{}
	Draw(c, IconData{Ops: []PathOp{Move(0, 0), Line(10, 10)}}, geometry.NewRect(0, 0, 24, 24), widget.ColorBlack)
	if len(c.lines) != 0 {
		t.Errorf("Draw with zero viewbox produced %d lines", len(c.lines))
	}
}

func TestDraw_EmptyBounds(t *testing.T) {
	c := &mockCanvas{}
	Draw(c, Add, geometry.Rect{}, widget.ColorBlack)
	if len(c.lines) != 0 {
		t.Errorf("Draw with empty bounds produced %d lines", len(c.lines))
	}
}

func TestDraw_SimpleLine(t *testing.T) {
	c := &mockCanvas{}
	data := IconData{
		Name:    "line",
		ViewBox: 24,
		Ops:     []PathOp{Move(0, 0), Line(24, 24)},
	}
	bounds := geometry.NewRect(0, 0, 24, 24)
	Draw(c, data, bounds, widget.ColorRed)

	if len(c.lines) != 1 {
		t.Fatalf("expected 1 line, got %d", len(c.lines))
	}
	line := c.lines[0]
	if line.color != widget.ColorRed {
		t.Errorf("line color = %v, want red", line.color)
	}
	// Scale is 1.0 (24/24), offset is 0
	if line.from.X != 0 || line.from.Y != 0 {
		t.Errorf("line.from = %v, want (0,0)", line.from)
	}
	if line.to.X != 24 || line.to.Y != 24 {
		t.Errorf("line.to = %v, want (24,24)", line.to)
	}
}

func TestDraw_ScalesCorrectly(t *testing.T) {
	c := &mockCanvas{}
	data := IconData{
		Name:    "line",
		ViewBox: 24,
		Ops:     []PathOp{Move(0, 0), Line(24, 0)},
	}
	// Render at 48x48 = scale 2.0
	bounds := geometry.NewRect(0, 0, 48, 48)
	Draw(c, data, bounds, widget.ColorBlack)

	if len(c.lines) != 1 {
		t.Fatalf("expected 1 line, got %d", len(c.lines))
	}
	line := c.lines[0]
	if line.to.X != 48 {
		t.Errorf("scaled line.to.X = %v, want 48", line.to.X)
	}
	// Stroke width should be 1.5 * 2.0 = 3.0
	if line.width != 3.0 {
		t.Errorf("scaled stroke width = %v, want 3.0", line.width)
	}
}

func TestDraw_CentersInRectangularBounds(t *testing.T) {
	c := &mockCanvas{}
	data := IconData{
		Name:    "line",
		ViewBox: 24,
		Ops:     []PathOp{Move(0, 0), Line(24, 0)},
	}
	// Wide bounds: 96x48. Scale = min(96/24, 48/24) = 2.0
	// Centered: offsetX = 0 + (96 - 24*2)/2 = 24
	bounds := geometry.NewRect(0, 0, 96, 48)
	Draw(c, data, bounds, widget.ColorBlack)

	if len(c.lines) != 1 {
		t.Fatalf("expected 1 line, got %d", len(c.lines))
	}
	line := c.lines[0]
	if line.from.X != 24 {
		t.Errorf("centered line.from.X = %v, want 24", line.from.X)
	}
}

func TestDraw_CloseOp(t *testing.T) {
	c := &mockCanvas{}
	data := IconData{
		Name:    "triangle",
		ViewBox: 24,
		Ops: []PathOp{
			Move(0, 0), Line(24, 0), Line(12, 24), ClosePath(),
		},
	}
	bounds := geometry.NewRect(0, 0, 24, 24)
	Draw(c, data, bounds, widget.ColorBlack)

	// 2 LineTo + 1 Close = 3 lines
	if len(c.lines) != 3 {
		t.Fatalf("triangle: expected 3 lines, got %d", len(c.lines))
	}
	// Close line goes back to (0,0)
	closeLine := c.lines[2]
	if closeLine.to.X != 0 || closeLine.to.Y != 0 {
		t.Errorf("close line.to = %v, want (0,0)", closeLine.to)
	}
}

func TestDraw_CloseOp_NoDoubleLineWhenAlreadyAtStart(t *testing.T) {
	c := &mockCanvas{}
	data := IconData{
		Name:    "back",
		ViewBox: 24,
		Ops: []PathOp{
			Move(0, 0), Line(10, 0), Line(0, 0), ClosePath(),
		},
	}
	bounds := geometry.NewRect(0, 0, 24, 24)
	Draw(c, data, bounds, widget.ColorBlack)

	// 2 LineTo only; Close should not add a line since we're at start
	if len(c.lines) != 2 {
		t.Errorf("expected 2 lines (no redundant close), got %d", len(c.lines))
	}
}

func TestDraw_CubicTo(t *testing.T) {
	c := &mockCanvas{}
	data := IconData{
		Name:    "curve",
		ViewBox: 24,
		Ops: []PathOp{
			Move(0, 12), Cubic(8, 0, 16, 24, 24, 12),
		},
	}
	bounds := geometry.NewRect(0, 0, 24, 24)
	Draw(c, data, bounds, widget.ColorBlack)

	// cubicSegments = 8 line segments
	if len(c.lines) != cubicSegments {
		t.Errorf("cubic: expected %d lines, got %d", cubicSegments, len(c.lines))
	}
	// First segment starts from (0,12)
	if c.lines[0].from.X != 0 || c.lines[0].from.Y != 12 {
		t.Errorf("cubic first line from = %v, want (0,12)", c.lines[0].from)
	}
	// Last segment ends at (24,12)
	last := c.lines[len(c.lines)-1]
	if last.to.X != 24 || last.to.Y != 12 {
		t.Errorf("cubic last line to = %v, want (24,12)", last.to)
	}
}

func TestDraw_BuiltinIcons_NoError(t *testing.T) {
	builtins := []IconData{
		Close, Check, ChevronDown, ChevronRight, Search, Settings, Menu, ArrowBack, Add, Delete,
		Play, Stop, Pause, Debug, Gear, Filter, FolderOpen, FolderClosed, Terminal, Refresh, Plus, Minus,
	}
	for _, ic := range builtins {
		t.Run(ic.Name, func(t *testing.T) {
			c := &mockCanvas{}
			Draw(c, ic, geometry.NewRect(0, 0, 48, 48), widget.ColorBlack)
			if len(c.lines) == 0 {
				t.Error("drawing built-in icon produced no lines")
			}
		})
	}
}

func TestDraw_WithOffset(t *testing.T) {
	c := &mockCanvas{}
	data := IconData{
		Name:    "line",
		ViewBox: 24,
		Ops:     []PathOp{Move(0, 0), Line(24, 0)},
	}
	// Bounds offset at (100, 200), size 24x24
	bounds := geometry.NewRect(100, 200, 24, 24)
	Draw(c, data, bounds, widget.ColorBlack)

	if len(c.lines) != 1 {
		t.Fatalf("expected 1 line, got %d", len(c.lines))
	}
	line := c.lines[0]
	if line.from.X != 100 || line.from.Y != 200 {
		t.Errorf("offset line.from = %v, want (100,200)", line.from)
	}
}

// --- IconWidget tests ---

func TestNewIcon_Defaults(t *testing.T) {
	w := NewIcon(Check)
	if w.Data().Name != "check" {
		t.Errorf("Data().Name = %q, want %q", w.Data().Name, "check")
	}
	if w.IconSize() != defaultIconSize {
		t.Errorf("IconSize() = %v, want %v", w.IconSize(), defaultIconSize)
	}
	if w.IconColor() != widget.ColorBlack {
		t.Errorf("IconColor() = %v, want ColorBlack", w.IconColor())
	}
	if !w.IsVisible() {
		t.Error("expected visible by default")
	}
	if !w.IsEnabled() {
		t.Error("expected enabled by default")
	}
}

func TestNewIcon_WithOptions(t *testing.T) {
	w := NewIcon(Close, Size(48), Color(widget.ColorRed), Label("Close button"))
	if w.IconSize() != 48 {
		t.Errorf("IconSize() = %v, want 48", w.IconSize())
	}
	if w.IconColor() != widget.ColorRed {
		t.Errorf("IconColor() = %v, want ColorRed", w.IconColor())
	}
	if w.IconLabel() != "Close button" {
		t.Errorf("IconLabel() = %q, want %q", w.IconLabel(), "Close button")
	}
}

func TestNewIcon_ColorSignalPrecedence(t *testing.T) {
	sig := state.NewSignal(widget.ColorGreen)
	w := NewIcon(Check, Color(widget.ColorRed), ColorSignal(sig))
	if w.IconColor() != widget.ColorGreen {
		t.Errorf("IconColor() = %v, want ColorGreen (from signal)", w.IconColor())
	}
	sig.Set(widget.ColorBlue)
	if w.IconColor() != widget.ColorBlue {
		t.Errorf("after Set: IconColor() = %v, want ColorBlue", w.IconColor())
	}
}

func TestIconWidget_Layout(t *testing.T) {
	w := NewIcon(Check, Size(32))
	ctx := widget.NewContext()
	size := w.Layout(ctx, geometry.Constraints{MaxWidth: 100, MaxHeight: 100})
	if size.Width != 32 || size.Height != 32 {
		t.Errorf("Layout() = %v, want (32,32)", size)
	}
}

func TestIconWidget_Layout_Constrained(t *testing.T) {
	w := NewIcon(Check, Size(48))
	ctx := widget.NewContext()
	size := w.Layout(ctx, geometry.Constraints{MaxWidth: 24, MaxHeight: 24})
	if size.Width != 24 || size.Height != 24 {
		t.Errorf("Layout() constrained = %v, want (24,24)", size)
	}
}

func TestIconWidget_Draw(t *testing.T) {
	w := NewIcon(Add, Size(24))
	ctx := widget.NewContext()
	_ = w.Layout(ctx, geometry.Constraints{MaxWidth: 100, MaxHeight: 100})

	c := &mockCanvas{}
	w.Draw(ctx, c)
	if len(c.lines) == 0 {
		t.Error("Draw produced no lines")
	}
}

func TestIconWidget_Draw_NotVisible(t *testing.T) {
	w := NewIcon(Add, Size(24))
	w.SetVisible(false)
	ctx := widget.NewContext()
	_ = w.Layout(ctx, geometry.Constraints{MaxWidth: 100, MaxHeight: 100})

	c := &mockCanvas{}
	w.Draw(ctx, c)
	if len(c.lines) != 0 {
		t.Error("Draw should produce no lines when not visible")
	}
}

func TestIconWidget_Draw_EmptyBounds(t *testing.T) {
	w := NewIcon(Add, Size(24))
	// Don't call Layout, bounds are zero
	ctx := widget.NewContext()
	c := &mockCanvas{}
	w.Draw(ctx, c)
	if len(c.lines) != 0 {
		t.Error("Draw should produce no lines with empty bounds")
	}
}

func TestIconWidget_Event(t *testing.T) {
	w := NewIcon(Check)
	ctx := widget.NewContext()
	consumed := w.Event(ctx, &event.MouseEvent{})
	if consumed {
		t.Error("IconWidget should not consume events")
	}
}

func TestIconWidget_Children(t *testing.T) {
	w := NewIcon(Check)
	if w.Children() != nil {
		t.Error("IconWidget.Children() should return nil")
	}
}

// --- Accessibility tests ---

func TestIconWidget_Accessibility(t *testing.T) {
	w := NewIcon(Check, Label("Confirm"))

	if w.AccessibilityRole() != a11y.RoleImage {
		t.Errorf("role = %v, want RoleImage", w.AccessibilityRole())
	}
	if w.AccessibilityLabel() != "Confirm" {
		t.Errorf("label = %q, want %q", w.AccessibilityLabel(), "Confirm")
	}
	if w.AccessibilityHint() != "" {
		t.Error("hint should be empty")
	}
	if w.AccessibilityValue() != "" {
		t.Error("value should be empty")
	}
	if w.AccessibilityActions() != nil {
		t.Error("actions should be nil")
	}

	st := w.AccessibilityState()
	if st.Hidden {
		t.Error("should not be hidden when visible")
	}
}

func TestIconWidget_AccessibilityLabel_FallbackToName(t *testing.T) {
	w := NewIcon(Check) // no explicit Label
	if w.AccessibilityLabel() != "check" {
		t.Errorf("label = %q, want %q (fallback to icon name)", w.AccessibilityLabel(), "check")
	}
}

func TestIconWidget_AccessibilityState_Hidden(t *testing.T) {
	w := NewIcon(Check)
	w.SetVisible(false)
	st := w.AccessibilityState()
	if !st.Hidden {
		t.Error("should be hidden when not visible")
	}
}

// --- Lifecycle tests ---

func TestIconWidget_Mount_WithoutSignal(t *testing.T) {
	w := NewIcon(Check)
	ctx := widget.NewContext()
	// Should not panic
	w.Mount(ctx)
	w.Unmount()
}

func TestIconWidget_Mount_WithNilScheduler(t *testing.T) {
	sig := state.NewSignal(widget.ColorRed)
	w := NewIcon(Check, ColorSignal(sig))
	ctx := widget.NewContext()
	// Scheduler is nil by default, should not panic
	w.Mount(ctx)
	w.Unmount()
}

// --- MultiColorIcon tests ---

func TestMultiColorIcon_IsValueType(t *testing.T) {
	original := MultiColorIcon{
		Name:    "test",
		ViewBox: 24,
		Groups: []PathGroup{
			{ColorKey: "primary", Ops: []PathOp{Move(0, 0)}},
		},
	}
	copied := original
	copied.Name = "modified"
	if original.Name == copied.Name {
		t.Error("MultiColorIcon should be a value type; modifying copy changed original")
	}
}

func TestPathGroup_HasColorKeyAndOps(t *testing.T) {
	pg := PathGroup{
		ColorKey: "accent",
		Ops:      []PathOp{Move(1, 2), Line(3, 4)},
	}
	if pg.ColorKey != "accent" {
		t.Errorf("ColorKey = %q, want %q", pg.ColorKey, "accent")
	}
	if len(pg.Ops) != 2 {
		t.Errorf("len(Ops) = %d, want 2", len(pg.Ops))
	}
}

// --- DrawMulti tests ---

func TestDrawMulti_EmptyGroups(t *testing.T) {
	c := &mockCanvas{}
	data := MultiColorIcon{Name: "empty", ViewBox: 24}
	DrawMulti(c, data, geometry.NewRect(0, 0, 24, 24), Palette{})
	if len(c.lines) != 0 {
		t.Errorf("DrawMulti with empty groups produced %d lines", len(c.lines))
	}
}

func TestDrawMulti_ZeroViewBox(t *testing.T) {
	c := &mockCanvas{}
	data := MultiColorIcon{
		Name: "zero",
		Groups: []PathGroup{
			{ColorKey: "primary", Ops: []PathOp{Move(0, 0), Line(10, 10)}},
		},
	}
	DrawMulti(c, data, geometry.NewRect(0, 0, 24, 24), Palette{})
	if len(c.lines) != 0 {
		t.Errorf("DrawMulti with zero viewbox produced %d lines", len(c.lines))
	}
}

func TestDrawMulti_EmptyBounds(t *testing.T) {
	c := &mockCanvas{}
	DrawMulti(c, FileGo, geometry.Rect{}, DefaultDarkPalette())
	if len(c.lines) != 0 {
		t.Errorf("DrawMulti with empty bounds produced %d lines", len(c.lines))
	}
}

func TestDrawMulti_UsesCorrectColors(t *testing.T) {
	c := &mockCanvas{}
	data := MultiColorIcon{
		Name:    "bicolor",
		ViewBox: 24,
		Groups: []PathGroup{
			{ColorKey: "primary", Ops: []PathOp{Move(0, 0), Line(12, 12)}},
			{ColorKey: "accent", Ops: []PathOp{Move(12, 0), Line(24, 12)}},
		},
	}
	palette := Palette{
		"primary": widget.ColorRed,
		"accent":  widget.ColorBlue,
	}
	DrawMulti(c, data, geometry.NewRect(0, 0, 24, 24), palette)

	if len(c.lines) != 2 {
		t.Fatalf("expected 2 lines, got %d", len(c.lines))
	}
	if c.lines[0].color != widget.ColorRed {
		t.Errorf("first group color = %v, want red", c.lines[0].color)
	}
	if c.lines[1].color != widget.ColorBlue {
		t.Errorf("second group color = %v, want blue", c.lines[1].color)
	}
}

func TestDrawMulti_MissingKeyFallsBackToGray(t *testing.T) {
	c := &mockCanvas{}
	data := MultiColorIcon{
		Name:    "missing",
		ViewBox: 24,
		Groups: []PathGroup{
			{ColorKey: "nonexistent", Ops: []PathOp{Move(0, 0), Line(24, 24)}},
		},
	}
	DrawMulti(c, data, geometry.NewRect(0, 0, 24, 24), Palette{})

	if len(c.lines) != 1 {
		t.Fatalf("expected 1 line, got %d", len(c.lines))
	}
	if c.lines[0].color != widget.ColorGray {
		t.Errorf("fallback color = %v, want ColorGray", c.lines[0].color)
	}
}

func TestDrawMulti_ScalesCorrectly(t *testing.T) {
	c := &mockCanvas{}
	data := MultiColorIcon{
		Name:    "scaled",
		ViewBox: 24,
		Groups: []PathGroup{
			{ColorKey: "primary", Ops: []PathOp{Move(0, 0), Line(24, 0)}},
		},
	}
	palette := Palette{"primary": widget.ColorBlack}
	// Render at 48x48 = scale 2.0
	DrawMulti(c, data, geometry.NewRect(0, 0, 48, 48), palette)

	if len(c.lines) != 1 {
		t.Fatalf("expected 1 line, got %d", len(c.lines))
	}
	if c.lines[0].to.X != 48 {
		t.Errorf("scaled line.to.X = %v, want 48", c.lines[0].to.X)
	}
	if c.lines[0].width != 3.0 {
		t.Errorf("scaled stroke width = %v, want 3.0", c.lines[0].width)
	}
}

func TestDrawMulti_CloseOp(t *testing.T) {
	c := &mockCanvas{}
	data := MultiColorIcon{
		Name:    "triangle",
		ViewBox: 24,
		Groups: []PathGroup{
			{ColorKey: "primary", Ops: []PathOp{
				Move(0, 0), Line(24, 0), Line(12, 24), ClosePath(),
			}},
		},
	}
	palette := Palette{"primary": widget.ColorBlack}
	DrawMulti(c, data, geometry.NewRect(0, 0, 24, 24), palette)

	// 2 LineTo + 1 Close = 3 lines
	if len(c.lines) != 3 {
		t.Fatalf("expected 3 lines, got %d", len(c.lines))
	}
	closeLine := c.lines[2]
	if closeLine.to.X != 0 || closeLine.to.Y != 0 {
		t.Errorf("close line.to = %v, want (0,0)", closeLine.to)
	}
}

func TestDrawMulti_CubicTo(t *testing.T) {
	c := &mockCanvas{}
	data := MultiColorIcon{
		Name:    "curve",
		ViewBox: 24,
		Groups: []PathGroup{
			{ColorKey: "primary", Ops: []PathOp{
				Move(0, 12), Cubic(8, 0, 16, 24, 24, 12),
			}},
		},
	}
	palette := Palette{"primary": widget.ColorBlack}
	DrawMulti(c, data, geometry.NewRect(0, 0, 24, 24), palette)

	if len(c.lines) != cubicSegments {
		t.Errorf("expected %d lines, got %d", cubicSegments, len(c.lines))
	}
}

// --- Built-in multi-color icon tests ---

func TestBuiltinMultiColorIcons_HaveRequiredFields(t *testing.T) {
	builtins := []struct {
		name string
		icon MultiColorIcon
	}{
		{"FileGo", FileGo},
		{"FileJSON", FileJSON},
		{"FileYAML", FileYAML},
		{"FileMD", FileMD},
		{"FileTest", FileTest},
		{"FileConfig", FileConfig},
		{"FileImage", FileImage},
		{"FileGeneric", FileGeneric},
		{"GitBranch", GitBranch},
		{"GitCommit", GitCommit},
		{"GitMerge", GitMerge},
		{"GitPR", GitPR},
		{"GitModified", GitModified},
	}
	for _, tt := range builtins {
		t.Run(tt.name, func(t *testing.T) {
			if tt.icon.Name == "" {
				t.Error("icon Name is empty")
			}
			if tt.icon.ViewBox != defaultViewBox {
				t.Errorf("icon ViewBox = %v, want %v", tt.icon.ViewBox, defaultViewBox)
			}
			if len(tt.icon.Groups) == 0 {
				t.Error("icon Groups is empty")
			}
			for i, g := range tt.icon.Groups {
				if g.ColorKey == "" {
					t.Errorf("group[%d].ColorKey is empty", i)
				}
				if len(g.Ops) == 0 {
					t.Errorf("group[%d].Ops is empty", i)
				}
				if g.Ops[0].Cmd != CmdMoveTo {
					t.Errorf("group[%d] first op = %v, want CmdMoveTo", i, g.Ops[0].Cmd)
				}
			}
		})
	}
}

func TestBuiltinMultiColorIcons_Count(t *testing.T) {
	count := 13 // 8 file type + 5 VCS
	builtins := []MultiColorIcon{
		FileGo, FileJSON, FileYAML, FileMD, FileTest, FileConfig, FileImage, FileGeneric,
		GitBranch, GitCommit, GitMerge, GitPR, GitModified,
	}
	if len(builtins) != count {
		t.Errorf("expected %d multi-color icons, got %d", count, len(builtins))
	}
}

func TestDrawMulti_BuiltinIcons_NoError(t *testing.T) {
	builtins := []MultiColorIcon{
		FileGo, FileJSON, FileYAML, FileMD, FileTest, FileConfig, FileImage, FileGeneric,
		GitBranch, GitCommit, GitMerge, GitPR, GitModified,
	}
	palettes := []struct {
		name    string
		palette Palette
	}{
		{"dark", DefaultDarkPalette()},
		{"light", DefaultLightPalette()},
	}
	for _, ic := range builtins {
		for _, p := range palettes {
			t.Run(ic.Name+"/"+p.name, func(t *testing.T) {
				c := &mockCanvas{}
				DrawMulti(c, ic, geometry.NewRect(0, 0, 48, 48), p.palette)
				if len(c.lines) == 0 {
					t.Error("drawing multi-color icon produced no lines")
				}
			})
		}
	}
}

// --- Palette tests ---

func TestDefaultDarkPalette_HasAllKeys(t *testing.T) {
	p := DefaultDarkPalette()
	keys := []string{
		KeyPrimary, KeyAccent, KeySecondary, KeySuccess,
		KeyError, KeyWarning, KeyGo, KeyJSON, KeyYAML,
		KeyRust, KeyPython, KeyMarkdown,
	}
	for _, key := range keys {
		if _, ok := p[key]; !ok {
			t.Errorf("DefaultDarkPalette missing key %q", key)
		}
	}
}

func TestDefaultLightPalette_HasAllKeys(t *testing.T) {
	p := DefaultLightPalette()
	keys := []string{
		KeyPrimary, KeyAccent, KeySecondary, KeySuccess,
		KeyError, KeyWarning, KeyGo, KeyJSON, KeyYAML,
		KeyRust, KeyPython, KeyMarkdown,
	}
	for _, key := range keys {
		if _, ok := p[key]; !ok {
			t.Errorf("DefaultLightPalette missing key %q", key)
		}
	}
}

func TestPalette_ColorsAreOpaque(t *testing.T) {
	for name, palette := range map[string]Palette{
		"dark":  DefaultDarkPalette(),
		"light": DefaultLightPalette(),
	} {
		for key, color := range palette {
			if color.A != 1 {
				t.Errorf("%s palette[%q].A = %v, want 1", name, key, color.A)
			}
		}
	}
}

func TestPalette_DarkAndLightDiffer(t *testing.T) {
	dark := DefaultDarkPalette()
	light := DefaultLightPalette()
	if dark[KeyPrimary] == light[KeyPrimary] {
		t.Error("dark and light palettes should have different primary colors")
	}
}

// --- MultiColorRegistry tests ---

func TestNewMultiColorRegistry_Empty(t *testing.T) {
	r := NewMultiColorRegistry()
	if r.Len() != 0 {
		t.Errorf("NewMultiColorRegistry().Len() = %d, want 0", r.Len())
	}
}

func TestMultiColorRegistry_RegisterAndGet(t *testing.T) {
	r := NewMultiColorRegistry()
	ic := MultiColorIcon{
		Name:    "test",
		ViewBox: 24,
		Groups: []PathGroup{
			{ColorKey: "primary", Ops: []PathOp{Move(0, 0)}},
		},
	}
	r.Register(ic)

	got, ok := r.Get("test")
	if !ok {
		t.Fatal("Get(test) returned false")
	}
	if got.Name != "test" {
		t.Errorf("Get(test).Name = %q, want %q", got.Name, "test")
	}
}

func TestMultiColorRegistry_Get_NotFound(t *testing.T) {
	r := NewMultiColorRegistry()
	_, ok := r.Get("nonexistent")
	if ok {
		t.Error("Get(nonexistent) should return false")
	}
}

func TestMultiColorRegistry_Register_Overwrite(t *testing.T) {
	r := NewMultiColorRegistry()
	r.Register(MultiColorIcon{Name: "ic", ViewBox: 24, Groups: []PathGroup{
		{ColorKey: "primary", Ops: []PathOp{Move(0, 0)}},
	}})
	r.Register(MultiColorIcon{Name: "ic", ViewBox: 48, Groups: []PathGroup{
		{ColorKey: "accent", Ops: []PathOp{Move(1, 1)}},
	}})

	got, ok := r.Get("ic")
	if !ok {
		t.Fatal("Get(ic) returned false")
	}
	if got.ViewBox != 48 {
		t.Errorf("overwritten icon ViewBox = %v, want 48", got.ViewBox)
	}
}

func TestMultiColorRegistry_Names(t *testing.T) {
	r := NewMultiColorRegistry()
	r.Register(MultiColorIcon{Name: "alpha", ViewBox: 24, Groups: []PathGroup{
		{ColorKey: "primary", Ops: []PathOp{Move(0, 0)}},
	}})
	r.Register(MultiColorIcon{Name: "beta", ViewBox: 24, Groups: []PathGroup{
		{ColorKey: "primary", Ops: []PathOp{Move(0, 0)}},
	}})

	names := r.Names()
	if len(names) != 2 {
		t.Fatalf("Names() returned %d items, want 2", len(names))
	}
	found := map[string]bool{}
	for _, n := range names {
		found[n] = true
	}
	if !found["alpha"] || !found["beta"] {
		t.Errorf("Names() = %v, expected alpha and beta", names)
	}
}

func TestMultiColorRegistry_Len(t *testing.T) {
	r := NewMultiColorRegistry()
	if r.Len() != 0 {
		t.Errorf("empty registry Len = %d", r.Len())
	}
	r.Register(MultiColorIcon{Name: "a", ViewBox: 24, Groups: []PathGroup{
		{ColorKey: "primary", Ops: []PathOp{Move(0, 0)}},
	}})
	if r.Len() != 1 {
		t.Errorf("registry Len = %d, want 1", r.Len())
	}
}

func TestDefaultMultiColorRegistry_ContainsBuiltins(t *testing.T) {
	r := DefaultMultiColorRegistry()
	if r.Len() < 13 {
		t.Errorf("DefaultMultiColorRegistry().Len() = %d, want >= 13", r.Len())
	}
	expected := []string{
		"file_go", "file_json", "file_yaml", "file_md",
		"file_test", "file_config", "file_image", "file_generic",
		"git_branch", "git_commit", "git_merge", "git_pr", "git_modified",
	}
	for _, name := range expected {
		if _, ok := r.Get(name); !ok {
			t.Errorf("DefaultMultiColorRegistry missing built-in icon %q", name)
		}
	}
}

// --- Color key constants tests ---

func TestColorKeyConstants_NotEmpty(t *testing.T) {
	keys := []string{
		KeyPrimary, KeyAccent, KeySecondary, KeySuccess,
		KeyError, KeyWarning, KeyGo, KeyJSON, KeyYAML,
		KeyRust, KeyPython, KeyMarkdown,
	}
	seen := make(map[string]bool)
	for _, key := range keys {
		if key == "" {
			t.Error("color key constant is empty")
		}
		if seen[key] {
			t.Errorf("duplicate color key constant %q", key)
		}
		seen[key] = true
	}
}
