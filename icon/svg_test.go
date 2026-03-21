package icon

import (
	"math"
	"testing"

	"github.com/gogpu/ui/geometry"
	"github.com/gogpu/ui/widget"
)

// --- FromSVG tests ---

func TestFromSVG_SimplePath(t *testing.T) {
	icon := FromSVG("test", 16, "M0 0L16 16Z")
	if icon.Name != "test" {
		t.Errorf("Name = %q, want %q", icon.Name, "test")
	}
	if icon.ViewBox != 16 {
		t.Errorf("ViewBox = %v, want 16", icon.ViewBox)
	}
	if len(icon.Ops) != 3 {
		t.Fatalf("len(Ops) = %d, want 3", len(icon.Ops))
	}
	if icon.Ops[0].Cmd != CmdMoveTo {
		t.Errorf("Ops[0].Cmd = %v, want CmdMoveTo", icon.Ops[0].Cmd)
	}
	if icon.Ops[1].Cmd != CmdLineTo {
		t.Errorf("Ops[1].Cmd = %v, want CmdLineTo", icon.Ops[1].Cmd)
	}
	if icon.Ops[2].Cmd != CmdClose {
		t.Errorf("Ops[2].Cmd = %v, want CmdClose", icon.Ops[2].Cmd)
	}
}

func TestFromSVG_MoveToCoords(t *testing.T) {
	icon := FromSVG("test", 24, "M10 20L5 15")
	if len(icon.Ops) < 1 {
		t.Fatal("expected at least 1 op")
	}
	op := icon.Ops[0]
	if op.Params[0] != 10 || op.Params[1] != 20 {
		t.Errorf("MoveTo params = (%v, %v), want (10, 20)", op.Params[0], op.Params[1])
	}
}

func TestFromSVG_CubicBezier(t *testing.T) {
	icon := FromSVG("test", 24, "M0 0C1 2 3 4 5 6")
	if len(icon.Ops) != 2 {
		t.Fatalf("len(Ops) = %d, want 2", len(icon.Ops))
	}
	cubic := icon.Ops[1]
	if cubic.Cmd != CmdCubicTo {
		t.Errorf("Ops[1].Cmd = %v, want CmdCubicTo", cubic.Cmd)
	}
	expected := [6]float32{1, 2, 3, 4, 5, 6}
	for i := 0; i < 6; i++ {
		if cubic.Params[i] != expected[i] {
			t.Errorf("CubicTo Params[%d] = %v, want %v", i, cubic.Params[i], expected[i])
		}
	}
}

func TestFromSVG_QuadraticBezier(t *testing.T) {
	icon := FromSVG("test", 24, "M0 0Q5 10 20 20")
	if len(icon.Ops) != 2 {
		t.Fatalf("len(Ops) = %d, want 2", len(icon.Ops))
	}
	quad := icon.Ops[1]
	if quad.Cmd != CmdQuadraticTo {
		t.Errorf("Ops[1].Cmd = %v, want CmdQuadraticTo", quad.Cmd)
	}
	if quad.Params[0] != 5 || quad.Params[1] != 10 {
		t.Errorf("QuadTo control = (%v, %v), want (5, 10)", quad.Params[0], quad.Params[1])
	}
	if quad.Params[2] != 20 || quad.Params[3] != 20 {
		t.Errorf("QuadTo endpoint = (%v, %v), want (20, 20)", quad.Params[2], quad.Params[3])
	}
}

func TestFromSVG_HorizontalAndVerticalLines(t *testing.T) {
	icon := FromSVG("test", 24, "M0 0H10V20")
	if len(icon.Ops) != 3 {
		t.Fatalf("len(Ops) = %d, want 3 (M, H->L, V->L)", len(icon.Ops))
	}
	// H10 becomes LineTo(10, 0)
	h := icon.Ops[1]
	if h.Cmd != CmdLineTo {
		t.Errorf("H command should become CmdLineTo, got %v", h.Cmd)
	}
	if h.Params[0] != 10 || h.Params[1] != 0 {
		t.Errorf("H10 -> LineTo(%v, %v), want (10, 0)", h.Params[0], h.Params[1])
	}
	// V20 becomes LineTo(10, 20)
	v := icon.Ops[2]
	if v.Params[0] != 10 || v.Params[1] != 20 {
		t.Errorf("V20 -> LineTo(%v, %v), want (10, 20)", v.Params[0], v.Params[1])
	}
}

func TestFromSVG_ArcConvertsTosCubics(t *testing.T) {
	// Simple arc: semicircle
	icon := FromSVG("test", 24, "M0 12A12 12 0 0 1 24 12")
	// Arc should be converted to one or more cubic Bezier segments
	if len(icon.Ops) < 2 {
		t.Fatalf("arc should produce at least MoveTo + CubicTo, got %d ops", len(icon.Ops))
	}
	hasCubic := false
	for _, op := range icon.Ops {
		if op.Cmd == CmdCubicTo {
			hasCubic = true
			break
		}
	}
	if !hasCubic {
		t.Error("arc should be converted to cubic Bezier segments")
	}
}

func TestFromSVG_RelativeCommands(t *testing.T) {
	icon := FromSVG("test", 24, "M10 10l5 5")
	if len(icon.Ops) != 2 {
		t.Fatalf("len(Ops) = %d, want 2", len(icon.Ops))
	}
	line := icon.Ops[1]
	// Relative l5 5 from (10,10) = absolute (15, 15)
	if line.Params[0] != 15 || line.Params[1] != 15 {
		t.Errorf("relative l5 5 from (10,10) = (%v, %v), want (15, 15)",
			line.Params[0], line.Params[1])
	}
}

func TestFromSVG_EmptyPath(t *testing.T) {
	icon := FromSVG("empty", 24, "")
	if len(icon.Ops) != 0 {
		t.Errorf("empty path should have 0 ops, got %d", len(icon.Ops))
	}
}

func TestFromSVG_Panics_OnInvalidPath(t *testing.T) {
	defer func() {
		r := recover()
		if r == nil {
			t.Error("FromSVG should panic on invalid path data")
		}
	}()
	FromSVG("bad", 24, "X invalid")
}

func TestTryFromSVG_ValidPath(t *testing.T) {
	icon, err := TryFromSVG("test", 16, "M0 0L16 16Z")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if icon.Name != "test" {
		t.Errorf("Name = %q, want %q", icon.Name, "test")
	}
	if len(icon.Ops) != 3 {
		t.Errorf("len(Ops) = %d, want 3", len(icon.Ops))
	}
}

func TestTryFromSVG_InvalidPath(t *testing.T) {
	_, err := TryFromSVG("bad", 16, "X invalid")
	if err == nil {
		t.Error("TryFromSVG should return error for invalid path data")
	}
}

func TestFromSVG_MultipleSubpaths(t *testing.T) {
	icon := FromSVG("multi", 24, "M0 0L12 12ZM12 0L24 12Z")
	// Should have: M, L, Z, M, L, Z = 6 ops
	if len(icon.Ops) != 6 {
		t.Errorf("len(Ops) = %d, want 6", len(icon.Ops))
	}
	moveCount := 0
	closeCount := 0
	for _, op := range icon.Ops {
		switch op.Cmd {
		case CmdMoveTo:
			moveCount++
		case CmdClose:
			closeCount++
		}
	}
	if moveCount != 2 {
		t.Errorf("moveCount = %d, want 2", moveCount)
	}
	if closeCount != 2 {
		t.Errorf("closeCount = %d, want 2", closeCount)
	}
}

// --- Quad constructor test ---

func TestQuad(t *testing.T) {
	op := Quad(1, 2, 3, 4)
	if op.Cmd != CmdQuadraticTo {
		t.Fatalf("Quad().Cmd = %v, want CmdQuadraticTo", op.Cmd)
	}
	for i, want := range []float32{1, 2, 3, 4} {
		if op.Params[i] != want {
			t.Errorf("Quad().Params[%d] = %v, want %v", i, op.Params[i], want)
		}
	}
}

// --- Draw QuadraticTo test ---

func TestDraw_QuadraticTo(t *testing.T) {
	c := &mockCanvas{}
	data := IconData{
		Name:    "quad_curve",
		ViewBox: 24,
		Ops: []PathOp{
			Move(0, 12), Quad(12, 0, 24, 12),
		},
	}
	bounds := geometry.NewRect(0, 0, 24, 24)
	Draw(c, data, bounds, widget.ColorBlack)

	// quadraticSegments = 8 line segments
	if len(c.lines) != quadraticSegments {
		t.Errorf("quadratic: expected %d lines, got %d", quadraticSegments, len(c.lines))
	}
	// First segment starts from (0,12)
	if c.lines[0].from.X != 0 || c.lines[0].from.Y != 12 {
		t.Errorf("quad first line from = %v, want (0,12)", c.lines[0].from)
	}
	// Last segment ends at (24,12)
	last := c.lines[len(c.lines)-1]
	if last.to.X != 24 || last.to.Y != 12 {
		t.Errorf("quad last line to = %v, want (24,12)", last.to)
	}
}

func TestDrawMulti_QuadraticTo(t *testing.T) {
	c := &mockCanvas{}
	data := MultiColorIcon{
		Name:    "quad_curve",
		ViewBox: 24,
		Groups: []PathGroup{
			{ColorKey: "primary", Ops: []PathOp{
				Move(0, 12), Quad(12, 0, 24, 12),
			}},
		},
	}
	palette := Palette{"primary": widget.ColorBlack}
	DrawMulti(c, data, geometry.NewRect(0, 0, 24, 24), palette)

	if len(c.lines) != quadraticSegments {
		t.Errorf("expected %d lines, got %d", quadraticSegments, len(c.lines))
	}
}

// --- SVG icons render correctly ---

func TestFromSVG_BuiltinSVGIcons_Render(t *testing.T) {
	// All SVG-based built-in icons should render without panics
	svgIcons := []IconData{
		Close, Check, ChevronDown, ChevronRight,
		Search, Settings, Gear,
		FolderOpen, FolderClosed, Refresh,
	}
	for _, ic := range svgIcons {
		t.Run(ic.Name, func(t *testing.T) {
			c := &mockCanvas{}
			Draw(c, ic, geometry.NewRect(0, 0, 48, 48), widget.ColorBlack)
			if len(c.lines) == 0 {
				t.Error("SVG icon produced no lines")
			}
		})
	}
}

func TestFromSVG_ViewBoxScaling(t *testing.T) {
	// An icon with viewBox=16 drawn at 32x32 should scale by 2x
	icon := FromSVG("scale_test", 16, "M0 0L16 0")
	c := &mockCanvas{}
	Draw(c, icon, geometry.NewRect(0, 0, 32, 32), widget.ColorBlack)

	if len(c.lines) != 1 {
		t.Fatalf("expected 1 line, got %d", len(c.lines))
	}
	// Scale = 32/16 = 2.0; endpoint = 16 * 2 = 32
	line := c.lines[0]
	if line.to.X != 32 {
		t.Errorf("scaled line.to.X = %v, want 32", line.to.X)
	}
}

func TestFromSVG_ComplexPath(t *testing.T) {
	// Test a path with mixed commands including smooth cubics and arcs
	icon := FromSVG("complex", 24,
		"M2 12C2 6.48 6.48 2 12 2S22 6.48 22 12 17.52 22 12 22 2 17.52 2 12Z")
	if len(icon.Ops) == 0 {
		t.Error("complex path should produce ops")
	}
	// Should have at least MoveTo and some CubicTo ops
	hasCubic := false
	for _, op := range icon.Ops {
		if op.Cmd == CmdCubicTo {
			hasCubic = true
			break
		}
	}
	if !hasCubic {
		t.Error("smooth cubic (S) should produce CubicTo ops")
	}
}

func TestDrawQuadratic_Accuracy(t *testing.T) {
	// Verify the quadratic approximation hits the expected midpoint
	// For Q(0,12) -> (12,0) -> (24,12), the midpoint at t=0.5 should be
	// at (12, 6) = (0.25*0 + 0.5*12 + 0.25*24, 0.25*12 + 0.5*0 + 0.25*12)
	c := &mockCanvas{}
	data := IconData{
		Name:    "quad_midpoint",
		ViewBox: 24,
		Ops: []PathOp{
			Move(0, 12), Quad(12, 0, 24, 12),
		},
	}
	Draw(c, data, geometry.NewRect(0, 0, 24, 24), widget.ColorBlack)

	// Find the line segment that crosses x=12 (the midpoint)
	var midY float32
	found := false
	for _, line := range c.lines {
		if line.from.X <= 12 && line.to.X >= 12 {
			midY = (line.from.Y + line.to.Y) / 2
			found = true
			break
		}
	}
	if !found {
		t.Fatal("no line segment crosses x=12")
	}
	// Expected midpoint Y is 6.0 (calculated from the quadratic formula)
	if math.Abs(float64(midY-6.0)) > 1.0 {
		t.Errorf("quadratic midpoint Y = %v, want ~6.0", midY)
	}
}
