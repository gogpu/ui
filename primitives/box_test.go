package primitives_test

import (
	"image"
	"testing"

	"github.com/gogpu/ui/a11y"
	"github.com/gogpu/ui/event"
	"github.com/gogpu/ui/geometry"
	"github.com/gogpu/ui/primitives"
	"github.com/gogpu/ui/state"
	"github.com/gogpu/ui/widget"
)

// --- Box construction and children ---

func TestBoxCreatesEmpty(t *testing.T) {
	b := primitives.Box()
	children := b.Children()
	if children != nil {
		t.Errorf("expected nil children, got %d", len(children))
	}
}

func TestBoxCreatesWithChildren(t *testing.T) {
	c1 := primitives.Text("A")
	c2 := primitives.Text("B")
	b := primitives.Box(c1, c2)

	children := b.Children()
	if len(children) != 2 {
		t.Fatalf("expected 2 children, got %d", len(children))
	}
}

func TestBoxChildrenReturnsCopy(t *testing.T) {
	c1 := primitives.Text("A")
	b := primitives.Box(c1)

	children := b.Children()
	children[0] = nil // mutate the copy

	children2 := b.Children()
	if children2[0] == nil {
		t.Error("Children() should return a defensive copy")
	}
}

func TestBoxIsVisibleAndEnabled(t *testing.T) {
	b := primitives.Box()
	if !b.IsVisible() {
		t.Error("box should be visible by default")
	}
	if !b.IsEnabled() {
		t.Error("box should be enabled by default")
	}
}

// --- Fluent style methods ---

func TestBoxPaddingUniform(t *testing.T) {
	b := primitives.Box().Padding(16)
	style := b.Style()
	if style.Padding.Top != 16 || style.Padding.Right != 16 ||
		style.Padding.Bottom != 16 || style.Padding.Left != 16 {
		t.Errorf("expected uniform padding 16, got %+v", style.Padding)
	}
}

func TestBoxPaddingXY(t *testing.T) {
	b := primitives.Box().PaddingXY(10, 20)
	style := b.Style()
	if style.Padding.Left != 10 || style.Padding.Right != 10 {
		t.Errorf("expected horizontal padding 10, got L=%f R=%f", style.Padding.Left, style.Padding.Right)
	}
	if style.Padding.Top != 20 || style.Padding.Bottom != 20 {
		t.Errorf("expected vertical padding 20, got T=%f B=%f", style.Padding.Top, style.Padding.Bottom)
	}
}

func TestBoxPaddingIndividual(t *testing.T) {
	b := primitives.Box().PaddingTop(1).PaddingRight(2).PaddingBottom(3).PaddingLeft(4)
	style := b.Style()
	if style.Padding.Top != 1 || style.Padding.Right != 2 ||
		style.Padding.Bottom != 3 || style.Padding.Left != 4 {
		t.Errorf("expected padding T=1 R=2 B=3 L=4, got %+v", style.Padding)
	}
}

func TestBoxBackground(t *testing.T) {
	c := widget.Hex(0xFF0000)
	b := primitives.Box().Background(c)
	if b.Style().Background != c {
		t.Error("background color not set")
	}
}

func TestBoxRounded(t *testing.T) {
	b := primitives.Box().Rounded(12)
	if b.Style().Radius != 12 {
		t.Errorf("expected radius 12, got %f", b.Style().Radius)
	}
}

func TestBoxBorder(t *testing.T) {
	c := widget.ColorBlack
	b := primitives.Box().BorderStyle(2, c)
	style := b.Style()
	if style.Border.Width != 2 {
		t.Errorf("expected border width 2, got %f", style.Border.Width)
	}
	if style.Border.Color != c {
		t.Error("border color mismatch")
	}
}

func TestBoxShadowLevel(t *testing.T) {
	b := primitives.Box().ShadowLevel(3)
	if b.Style().Shadow.Level != 3 {
		t.Errorf("expected shadow level 3, got %d", b.Style().Shadow.Level)
	}
}

func TestBoxShadowLevelClamped(t *testing.T) {
	b := primitives.Box().ShadowLevel(-1)
	if b.Style().Shadow.Level != 0 {
		t.Errorf("expected level 0 for negative input, got %d", b.Style().Shadow.Level)
	}

	b = primitives.Box().ShadowLevel(99)
	if b.Style().Shadow.Level != 5 {
		t.Errorf("expected level 5 for overflow, got %d", b.Style().Shadow.Level)
	}
}

func TestBoxGap(t *testing.T) {
	b := primitives.Box().Gap(8)
	if b.Style().Gap != 8 {
		t.Errorf("expected gap 8, got %f", b.Style().Gap)
	}
}

func TestBoxExplicitDimensions(t *testing.T) {
	b := primitives.Box().Width(100).Height(50)
	style := b.Style()
	if style.ExplicitWidth != 100 || style.ExplicitHeight != 50 {
		t.Errorf("expected 100x50, got %fx%f", style.ExplicitWidth, style.ExplicitHeight)
	}
}

func TestBoxMinMaxDimensions(t *testing.T) {
	b := primitives.Box().MinWidthValue(50).MinHeightValue(30).MaxWidthValue(200).MaxHeightValue(100)
	style := b.Style()
	if style.MinWidth != 50 || style.MinHeight != 30 ||
		style.MaxWidth != 200 || style.MaxHeight != 100 {
		t.Errorf("min/max mismatch: %+v", style)
	}
}

func TestBoxLabel(t *testing.T) {
	b := primitives.Box().Label("Navigation")
	if b.AccessibilityLabel() != "Navigation" {
		t.Errorf("expected label 'Navigation', got %q", b.AccessibilityLabel())
	}
}

// --- Layout ---

func TestBoxLayoutNoPaddingNoChildren(t *testing.T) {
	b := primitives.Box()
	ctx := widget.NewContext()
	size := b.Layout(ctx, geometry.Loose(geometry.Sz(300, 300)))
	// No children = size is zero, constrained to min
	if size.Width != 0 || size.Height != 0 {
		t.Errorf("expected 0x0, got %s", size)
	}
}

func TestBoxLayoutWithPadding(t *testing.T) {
	child := primitives.Text("Hi")
	b := primitives.Box(child).Padding(10)
	ctx := widget.NewContext()
	size := b.Layout(ctx, geometry.Loose(geometry.Sz(300, 300)))

	// Child text "Hi" (2 chars * 0.6 * 14 = 16.8 width, 14*1.2=16.8 height)
	// Plus padding: 16.8 + 20 = 36.8 width, 16.8 + 20 = 36.8 height
	if size.Width < 30 || size.Height < 30 {
		t.Errorf("size too small with padding: %s", size)
	}
}

func TestBoxLayoutWithGap(t *testing.T) {
	c1 := primitives.Text("A")
	c2 := primitives.Text("B")
	b := primitives.Box(c1, c2).Gap(10)
	ctx := widget.NewContext()

	sizeWithGap := b.Layout(ctx, geometry.Loose(geometry.Sz(300, 300)))

	b2 := primitives.Box(primitives.Text("A"), primitives.Text("B"))
	sizeNoGap := b2.Layout(ctx, geometry.Loose(geometry.Sz(300, 300)))

	if sizeWithGap.Height-sizeNoGap.Height < 9 {
		t.Errorf("gap should add ~10px height, got diff=%f", sizeWithGap.Height-sizeNoGap.Height)
	}
}

func TestBoxLayoutExplicitSize(t *testing.T) {
	b := primitives.Box().Width(100).Height(50)
	ctx := widget.NewContext()
	size := b.Layout(ctx, geometry.Loose(geometry.Sz(300, 300)))
	if size.Width != 100 || size.Height != 50 {
		t.Errorf("expected 100x50, got %s", size)
	}
}

func TestBoxLayoutChildGetsConstrainedSpace(t *testing.T) {
	child := primitives.Text("Hello World Hello World Hello World")
	b := primitives.Box(child).Padding(20).Width(100)
	ctx := widget.NewContext()
	_ = b.Layout(ctx, geometry.Loose(geometry.Sz(300, 300)))

	// Child should be positioned inside the padding
	childBounds := child.Bounds()
	if childBounds.Min.X != 20 || childBounds.Min.Y != 20 {
		t.Errorf("child should start at padding offset, got %s", childBounds.Min)
	}
}

// --- Nested boxes ---

func TestBoxNested(t *testing.T) {
	inner := primitives.Box(
		primitives.Text("Inner"),
	).Padding(8)

	outer := primitives.Box(inner).Padding(16)
	ctx := widget.NewContext()
	size := outer.Layout(ctx, geometry.Loose(geometry.Sz(400, 400)))

	// Should be larger than inner alone
	if size.Width < 40 || size.Height < 40 {
		t.Errorf("nested box too small: %s", size)
	}
}

// --- Draw ---

func TestBoxDrawNoPanicEmpty(t *testing.T) {
	b := primitives.Box()
	ctx := widget.NewContext()
	canvas := &mockCanvas{}
	_ = b.Layout(ctx, geometry.Loose(geometry.Sz(100, 100)))
	b.Draw(ctx, canvas) // Should not panic
}

func TestBoxDrawBackgroundRendered(t *testing.T) {
	b := primitives.Box().Background(widget.ColorWhite).Width(100).Height(50)
	ctx := widget.NewContext()
	canvas := &mockCanvas{}
	_ = b.Layout(ctx, geometry.Loose(geometry.Sz(200, 200)))
	b.Draw(ctx, canvas)

	if canvas.drawRectCount == 0 {
		t.Error("expected background DrawRect call")
	}
}

func TestBoxDrawBorderRendered(t *testing.T) {
	b := primitives.Box().BorderStyle(2, widget.ColorBlack).Width(100).Height(50)
	ctx := widget.NewContext()
	canvas := &mockCanvas{}
	_ = b.Layout(ctx, geometry.Loose(geometry.Sz(200, 200)))
	b.Draw(ctx, canvas)

	if canvas.strokeRectCount == 0 {
		t.Error("expected border StrokeRect call")
	}
}

func TestBoxDrawRoundedBackground(t *testing.T) {
	b := primitives.Box().Background(widget.ColorWhite).Rounded(12).Width(100).Height(50)
	ctx := widget.NewContext()
	canvas := &mockCanvas{}
	_ = b.Layout(ctx, geometry.Loose(geometry.Sz(200, 200)))
	b.Draw(ctx, canvas)

	if canvas.drawRoundRectCount == 0 {
		t.Error("expected DrawRoundRect call for rounded background")
	}
}

func TestBoxDrawShadowMultiLayer(t *testing.T) {
	// Level 3 has 3 layers. Even without box radius, shadow layers have
	// radiusExtra > 0, so they use DrawRoundRect.
	b := primitives.Box().ShadowLevel(3).Width(100).Height(50)
	ctx := widget.NewContext()
	canvas := &mockCanvas{}
	_ = b.Layout(ctx, geometry.Loose(geometry.Sz(200, 200)))
	b.Draw(ctx, canvas)

	if canvas.drawRoundRectCount < 3 {
		t.Errorf("level 3 shadow should produce at least 3 DrawRoundRect calls, got %d",
			canvas.drawRoundRectCount)
	}
}

func TestBoxDrawShadowRoundedMultiLayer(t *testing.T) {
	// Level 2 with rounded corners and visible background.
	// Level 2 has 3 shadow layers (DrawRoundRect each) + 1 background = 4.
	b := primitives.Box().
		ShadowLevel(2).
		Rounded(8).
		Background(widget.ColorWhite).
		Width(100).Height(50)
	ctx := widget.NewContext()
	canvas := &mockCanvas{}
	_ = b.Layout(ctx, geometry.Loose(geometry.Sz(200, 200)))
	b.Draw(ctx, canvas)

	// 3 shadow layers (DrawRoundRect) + 1 background (DrawRoundRect) = 4
	if canvas.drawRoundRectCount != 4 {
		t.Errorf("level 2 rounded shadow + background should produce 4 DrawRoundRect calls, got %d",
			canvas.drawRoundRectCount)
	}
}

func TestBoxDrawShadowLevelZero(t *testing.T) {
	// Level 0 should produce no shadow draw calls at all.
	b := primitives.Box().ShadowLevel(0).Background(widget.ColorWhite).Width(100).Height(50)
	ctx := widget.NewContext()
	canvas := &mockCanvas{}
	_ = b.Layout(ctx, geometry.Loose(geometry.Sz(200, 200)))
	b.Draw(ctx, canvas)

	// Only 1 DrawRect for background, no shadow rects.
	if canvas.drawRectCount != 1 {
		t.Errorf("level 0 shadow should produce only background DrawRect (1), got %d", canvas.drawRectCount)
	}
}

func TestBoxDrawShadowProgressiveLayers(t *testing.T) {
	// Higher shadow levels should produce more or equal draw calls.
	// All shadow layers have radiusExtra > 0, so they use DrawRoundRect.
	ctx := widget.NewContext()
	var prevCount int
	for level := 1; level <= 5; level++ {
		b := primitives.Box().ShadowLevel(level).Width(100).Height(50)
		canvas := &mockCanvas{}
		_ = b.Layout(ctx, geometry.Loose(geometry.Sz(200, 200)))
		b.Draw(ctx, canvas)

		count := canvas.drawRoundRectCount
		if level > 1 && count < prevCount {
			t.Errorf("level %d produced %d DrawRoundRect calls, less than level %d (%d)",
				level, count, level-1, prevCount)
		}
		prevCount = count
	}
}

func TestBoxDrawChildrenWithTransform(t *testing.T) {
	child := primitives.Text("Hi").FontSize(14)
	b := primitives.Box(child).Padding(10).Width(100).Height(50)
	ctx := widget.NewContext()
	canvas := &mockCanvas{}
	_ = b.Layout(ctx, geometry.Loose(geometry.Sz(200, 200)))
	b.Draw(ctx, canvas)

	if canvas.pushTransformCount == 0 || canvas.popTransformCount == 0 {
		t.Error("expected PushTransform/PopTransform for children")
	}
}

func TestBoxDrawInvisible(t *testing.T) {
	b := primitives.Box().Background(widget.ColorWhite).Width(100).Height(50)
	b.SetVisible(false)
	ctx := widget.NewContext()
	canvas := &mockCanvas{}
	_ = b.Layout(ctx, geometry.Loose(geometry.Sz(200, 200)))
	b.Draw(ctx, canvas)

	if canvas.drawRectCount != 0 {
		t.Error("invisible box should not draw")
	}
}

// --- Event ---

func TestBoxEventDispatchesToChildren(t *testing.T) {
	consumed := false
	child := &eventConsumer{consume: true, called: &consumed}

	b := primitives.Box(child)
	ctx := widget.NewContext()
	e := &event.Base{}

	result := b.Event(ctx, e)
	if !result || !consumed {
		t.Error("event should be dispatched to and consumed by child")
	}
}

func TestBoxEventReturnsfalseWhenNoChild(t *testing.T) {
	b := primitives.Box()
	ctx := widget.NewContext()
	e := &event.Base{}

	if b.Event(ctx, e) {
		t.Error("empty box should not consume events")
	}
}

func TestBoxEventDisabledSkips(t *testing.T) {
	consumed := false
	child := &eventConsumer{consume: true, called: &consumed}

	b := primitives.Box(child)
	b.SetEnabled(false)
	ctx := widget.NewContext()
	e := &event.Base{}

	if b.Event(ctx, e) {
		t.Error("disabled box should not consume events")
	}
	if consumed {
		t.Error("children should not receive events when box is disabled")
	}
}

// --- Accessibility ---

func TestBoxAccessibilityRole(t *testing.T) {
	b := primitives.Box()
	if b.AccessibilityRole() != a11y.RoleGenericContainer {
		t.Errorf("expected RoleGenericContainer, got %s", b.AccessibilityRole())
	}
}

func TestBoxAccessibilityLabelDefault(t *testing.T) {
	b := primitives.Box()
	if b.AccessibilityLabel() != "" {
		t.Errorf("expected empty label, got %q", b.AccessibilityLabel())
	}
}

func TestBoxAccessibilityLabelCustom(t *testing.T) {
	b := primitives.Box().Label("Nav bar")
	if b.AccessibilityLabel() != "Nav bar" {
		t.Errorf("expected 'Nav bar', got %q", b.AccessibilityLabel())
	}
}

func TestBoxAccessibilityState(t *testing.T) {
	b := primitives.Box()
	a11yState := b.AccessibilityState()
	if a11yState.Disabled || a11yState.Hidden {
		t.Error("default state should be enabled and visible")
	}

	b.SetEnabled(false)
	a11yState = b.AccessibilityState()
	if !a11yState.Disabled {
		t.Error("disabled box should report Disabled=true")
	}

	b.SetVisible(false)
	a11yState = b.AccessibilityState()
	if !a11yState.Hidden {
		t.Error("invisible box should report Hidden=true")
	}
}

func TestBoxAccessibilityActions(t *testing.T) {
	b := primitives.Box()
	if b.AccessibilityActions() != nil {
		t.Error("box should have no actions")
	}
}

func TestBoxAccessibilityHint(t *testing.T) {
	b := primitives.Box()
	if b.AccessibilityHint() != "" {
		t.Error("box should have no hint")
	}
}

func TestBoxAccessibilityValue(t *testing.T) {
	b := primitives.Box()
	if b.AccessibilityValue() != "" {
		t.Error("box should have no value")
	}
}

// --- Chaining ---

func TestBoxFluentChaining(t *testing.T) {
	b := primitives.Box().
		Padding(16).
		Background(widget.Hex(0xFFFFFF)).
		Rounded(8).
		BorderStyle(1, widget.ColorBlack).
		ShadowLevel(2).
		Gap(4).
		Width(200).
		Height(100).
		MinWidthValue(50).
		MaxWidthValue(300).
		Label("Card")

	style := b.Style()
	if style.Padding.Top != 16 {
		t.Error("padding not chained")
	}
	if style.Radius != 8 {
		t.Error("radius not chained")
	}
	if style.Border.Width != 1 {
		t.Error("border not chained")
	}
	if style.Shadow.Level != 2 {
		t.Error("shadow not chained")
	}
	if style.Gap != 4 {
		t.Error("gap not chained")
	}
	if style.ExplicitWidth != 200 || style.ExplicitHeight != 100 {
		t.Error("dimensions not chained")
	}
	if style.MinWidth != 50 || style.MaxWidth != 300 {
		t.Error("min/max not chained")
	}
	if b.AccessibilityLabel() != "Card" {
		t.Error("label not chained")
	}
}

// --- Direction ---

func TestBoxDefaultDirectionIsVertical(t *testing.T) {
	b := primitives.Box()
	if b.ResolvedDirection() != primitives.DirectionVertical {
		t.Errorf("expected DirectionVertical, got %s", b.ResolvedDirection())
	}
}

func TestBoxSetDirection(t *testing.T) {
	b := primitives.Box().SetDirection(primitives.DirectionHorizontal)
	if b.ResolvedDirection() != primitives.DirectionHorizontal {
		t.Errorf("expected DirectionHorizontal, got %s", b.ResolvedDirection())
	}
}

func TestBoxSetDirectionSetsNeedsRedraw(t *testing.T) {
	b := primitives.Box()
	b.SetNeedsRedraw(false)
	b.SetDirection(primitives.DirectionHorizontal)
	if !b.NeedsRedraw() {
		t.Error("SetDirection should mark widget as needing redraw")
	}
}

func TestBoxSetDirectionNoopSameValue(t *testing.T) {
	b := primitives.Box()
	b.SetNeedsRedraw(false)
	// Default is Vertical, setting Vertical again should not dirty.
	b.SetDirection(primitives.DirectionVertical)
	if b.NeedsRedraw() {
		t.Error("SetDirection to same value should not mark redraw")
	}
}

func TestDirectionString(t *testing.T) {
	tests := []struct {
		d    primitives.Direction
		want string
	}{
		{primitives.DirectionVertical, "Vertical"},
		{primitives.DirectionHorizontal, "Horizontal"},
		{primitives.Direction(99), "Unknown"},
	}
	for _, tt := range tests {
		if got := tt.d.String(); got != tt.want {
			t.Errorf("Direction(%d).String() = %q, want %q", tt.d, got, tt.want)
		}
	}
}

// --- HBox / VBox convenience constructors ---

func TestHBoxSetsHorizontalDirection(t *testing.T) {
	b := primitives.HBox(primitives.Text("A"), primitives.Text("B"))
	if b.ResolvedDirection() != primitives.DirectionHorizontal {
		t.Errorf("HBox should be horizontal, got %s", b.ResolvedDirection())
	}
	if len(b.Children()) != 2 {
		t.Errorf("expected 2 children, got %d", len(b.Children()))
	}
}

func TestVBoxSetsVerticalDirection(t *testing.T) {
	b := primitives.VBox(primitives.Text("A"), primitives.Text("B"))
	if b.ResolvedDirection() != primitives.DirectionVertical {
		t.Errorf("VBox should be vertical, got %s", b.ResolvedDirection())
	}
}

func TestHBoxFluentChaining(t *testing.T) {
	b := primitives.HBox(primitives.Text("A")).Gap(8).Padding(4)
	if b.ResolvedDirection() != primitives.DirectionHorizontal {
		t.Error("HBox should remain horizontal after chaining")
	}
	if b.Style().Gap != 8 {
		t.Errorf("expected gap 8, got %f", b.Style().Gap)
	}
}

// --- Horizontal layout ---

func TestHBoxLayoutChildrenSideBySide(t *testing.T) {
	c1 := primitives.Box().Width(50).Height(30)
	c2 := primitives.Box().Width(70).Height(30)
	b := primitives.HBox(c1, c2)
	ctx := widget.NewContext()
	size := b.Layout(ctx, geometry.Loose(geometry.Sz(400, 400)))

	// Total width = 50 + 70 = 120, height = max(30, 30) = 30.
	if size.Width != 120 {
		t.Errorf("expected width 120, got %f", size.Width)
	}
	if size.Height != 30 {
		t.Errorf("expected height 30, got %f", size.Height)
	}
}

func TestHBoxLayoutWithGap(t *testing.T) {
	c1 := primitives.Box().Width(50).Height(30)
	c2 := primitives.Box().Width(70).Height(30)
	b := primitives.HBox(c1, c2).Gap(10)
	ctx := widget.NewContext()
	size := b.Layout(ctx, geometry.Loose(geometry.Sz(400, 400)))

	// Total width = 50 + 10 + 70 = 130.
	if size.Width != 130 {
		t.Errorf("expected width 130, got %f", size.Width)
	}
}

func TestHBoxLayoutWithPadding(t *testing.T) {
	c1 := primitives.Box().Width(40).Height(20)
	c2 := primitives.Box().Width(60).Height(20)
	b := primitives.HBox(c1, c2).Padding(10)
	ctx := widget.NewContext()
	size := b.Layout(ctx, geometry.Loose(geometry.Sz(400, 400)))

	// Content width = 40 + 60 = 100, + padding 20 = 120.
	// Content height = 20, + padding 20 = 40.
	if size.Width != 120 {
		t.Errorf("expected width 120, got %f", size.Width)
	}
	if size.Height != 40 {
		t.Errorf("expected height 40, got %f", size.Height)
	}
}

func TestHBoxLayoutChildPositions(t *testing.T) {
	c1 := primitives.Box().Width(50).Height(30)
	c2 := primitives.Box().Width(70).Height(40)
	b := primitives.HBox(c1, c2).Gap(10).Padding(5)
	ctx := widget.NewContext()
	_ = b.Layout(ctx, geometry.Loose(geometry.Sz(400, 400)))

	// c1 should be at (5, 5) — the padding offset.
	b1 := c1.Bounds()
	if b1.Min.X != 5 || b1.Min.Y != 5 {
		t.Errorf("c1 position expected (5,5), got (%f,%f)", b1.Min.X, b1.Min.Y)
	}

	// c2 should be at (5+50+10, 5) = (65, 5).
	b2 := c2.Bounds()
	if b2.Min.X != 65 || b2.Min.Y != 5 {
		t.Errorf("c2 position expected (65,5), got (%f,%f)", b2.Min.X, b2.Min.Y)
	}
}

func TestHBoxLayoutMaxChildHeight(t *testing.T) {
	c1 := primitives.Box().Width(50).Height(20)
	c2 := primitives.Box().Width(50).Height(60)
	b := primitives.HBox(c1, c2)
	ctx := widget.NewContext()
	size := b.Layout(ctx, geometry.Loose(geometry.Sz(400, 400)))

	// Height should be the maximum child height.
	if size.Height != 60 {
		t.Errorf("expected height 60, got %f", size.Height)
	}
}

func TestHBoxLayoutEmpty(t *testing.T) {
	b := primitives.HBox()
	ctx := widget.NewContext()
	size := b.Layout(ctx, geometry.Loose(geometry.Sz(300, 300)))
	if size.Width != 0 || size.Height != 0 {
		t.Errorf("expected 0x0, got %s", size)
	}
}

func TestHBoxLayoutSingleChild(t *testing.T) {
	c1 := primitives.Box().Width(80).Height(40)
	b := primitives.HBox(c1).Gap(10)
	ctx := widget.NewContext()
	size := b.Layout(ctx, geometry.Loose(geometry.Sz(300, 300)))

	// Gap should not be added for a single child.
	if size.Width != 80 {
		t.Errorf("expected width 80, got %f", size.Width)
	}
}

func TestHBoxLayoutConstrainedWidth(t *testing.T) {
	c1 := primitives.Box().Width(100).Height(30)
	c2 := primitives.Box().Width(100).Height(30)
	b := primitives.HBox(c1, c2)
	ctx := widget.NewContext()
	size := b.Layout(ctx, geometry.Loose(geometry.Sz(150, 300)))

	// Constrained: max 150, children want 200. The second child should
	// get remaining space (max 50).
	if size.Width != 150 {
		t.Errorf("expected constrained width 150, got %f", size.Width)
	}
}

// --- Nested horizontal/vertical ---

func TestNestedHBoxInVBox(t *testing.T) {
	row1 := primitives.HBox(
		primitives.Box().Width(50).Height(20),
		primitives.Box().Width(50).Height(20),
	).Gap(5)
	row2 := primitives.HBox(
		primitives.Box().Width(30).Height(20),
		primitives.Box().Width(30).Height(20),
	).Gap(5)

	outer := primitives.VBox(row1, row2).Gap(10)
	ctx := widget.NewContext()
	size := outer.Layout(ctx, geometry.Loose(geometry.Sz(400, 400)))

	// Row 1: 50 + 5 + 50 = 105 wide, 20 high.
	// Row 2: 30 + 5 + 30 = 65 wide, 20 high.
	// Outer: max(105, 65) = 105 wide, 20 + 10 + 20 = 50 high.
	if size.Width != 105 {
		t.Errorf("expected nested width 105, got %f", size.Width)
	}
	if size.Height != 50 {
		t.Errorf("expected nested height 50, got %f", size.Height)
	}
}

// --- Draw with horizontal layout ---

func TestHBoxDrawChildrenWithTransform(t *testing.T) {
	child := primitives.Box().Width(50).Height(30).Background(widget.ColorWhite)
	b := primitives.HBox(child).Padding(10)
	ctx := widget.NewContext()
	canvas := &mockCanvas{}
	_ = b.Layout(ctx, geometry.Loose(geometry.Sz(200, 200)))
	b.Draw(ctx, canvas)

	if canvas.pushTransformCount == 0 || canvas.popTransformCount == 0 {
		t.Error("expected PushTransform/PopTransform for children")
	}
}

// --- Event dispatch in horizontal layout ---

func TestHBoxEventDispatchesToChildren(t *testing.T) {
	consumed := false
	child := &eventConsumer{consume: true, called: &consumed}

	b := primitives.HBox(child)
	ctx := widget.NewContext()
	e := &event.Base{}

	result := b.Event(ctx, e)
	if !result || !consumed {
		t.Error("event should be dispatched to and consumed by child in HBox")
	}
}

// --- DirectionSignal ---

func TestBoxDirectionSignal(t *testing.T) {
	sig := state.NewSignal(primitives.DirectionHorizontal)
	b := primitives.Box(primitives.Text("A")).DirectionSignal(sig)

	if b.ResolvedDirection() != primitives.DirectionHorizontal {
		t.Errorf("expected DirectionHorizontal from signal, got %s", b.ResolvedDirection())
	}

	sig.Set(primitives.DirectionVertical)
	if b.ResolvedDirection() != primitives.DirectionVertical {
		t.Errorf("expected DirectionVertical after signal update, got %s", b.ResolvedDirection())
	}
}

func TestBoxDirectionSignalTakesPrecedenceOverStatic(t *testing.T) {
	sig := state.NewSignal(primitives.DirectionHorizontal)
	b := primitives.Box().SetDirection(primitives.DirectionVertical).DirectionSignal(sig)

	// Signal should win over static.
	if b.ResolvedDirection() != primitives.DirectionHorizontal {
		t.Errorf("signal should take precedence over static, got %s", b.ResolvedDirection())
	}
}

func TestBoxDirectionSignalMount(t *testing.T) {
	sig := state.NewSignal(primitives.DirectionVertical)
	b := primitives.Box(primitives.Text("A")).DirectionSignal(sig)

	var flushed []widget.Widget
	sched := state.NewScheduler(func(dirty []widget.Widget) {
		flushed = append(flushed, dirty...)
	})

	ctx := widget.NewContext()
	ctx.SetScheduler(sched)

	// Mount should create the binding.
	b.Mount(ctx)

	sig.Set(primitives.DirectionHorizontal)
	sched.Flush()

	if len(flushed) != 1 {
		t.Fatalf("expected 1 flushed widget, got %d", len(flushed))
	}
}

func TestBoxMountWithoutSignalNoBinding(t *testing.T) {
	b := primitives.Box()

	var flushed []widget.Widget
	sched := state.NewScheduler(func(dirty []widget.Widget) {
		flushed = append(flushed, dirty...)
	})

	ctx := widget.NewContext()
	ctx.SetScheduler(sched)
	b.Mount(ctx)

	sched.Flush()
	if len(flushed) != 0 {
		t.Error("no signal means no bindings should be created")
	}
}

func TestBoxMountWithoutScheduler(t *testing.T) {
	sig := state.NewSignal(primitives.DirectionVertical)
	b := primitives.Box().DirectionSignal(sig)

	ctx := widget.NewContext()
	// No scheduler set — Mount should not panic.
	b.Mount(ctx)
}

func TestBoxDirectionSignalLayoutChange(t *testing.T) {
	sig := state.NewSignal(primitives.DirectionVertical)
	c1 := primitives.Box().Width(50).Height(30)
	c2 := primitives.Box().Width(70).Height(40)
	b := primitives.Box(c1, c2).DirectionSignal(sig)
	ctx := widget.NewContext()

	// Vertical layout.
	sizeV := b.Layout(ctx, geometry.Loose(geometry.Sz(400, 400)))
	// Width = max(50, 70) = 70, Height = 30 + 40 = 70.
	if sizeV.Width != 70 {
		t.Errorf("vertical: expected width 70, got %f", sizeV.Width)
	}
	if sizeV.Height != 70 {
		t.Errorf("vertical: expected height 70, got %f", sizeV.Height)
	}

	// Switch to horizontal.
	sig.Set(primitives.DirectionHorizontal)
	sizeH := b.Layout(ctx, geometry.Loose(geometry.Sz(400, 400)))
	// Width = 50 + 70 = 120, Height = max(30, 40) = 40.
	if sizeH.Width != 120 {
		t.Errorf("horizontal: expected width 120, got %f", sizeH.Width)
	}
	if sizeH.Height != 40 {
		t.Errorf("horizontal: expected height 40, got %f", sizeH.Height)
	}
}

// --- Test helpers ---

// mockCanvas tracks draw calls for verification.
type mockCanvas struct {
	drawRectCount        int
	strokeRectCount      int
	drawRoundRectCount   int
	strokeRoundRectCount int
	drawCircleCount      int
	strokeCircleCount    int
	drawLineCount        int
	drawTextCount        int
	pushClipCount        int
	popClipCount         int
	pushTransformCount   int
	popTransformCount    int
	lastTextColor        widget.Color
	lastText             string
}

func (c *mockCanvas) Clear(_ widget.Color)                                  {}
func (c *mockCanvas) DrawRect(_ geometry.Rect, _ widget.Color)              { c.drawRectCount++ }
func (c *mockCanvas) StrokeRect(_ geometry.Rect, _ widget.Color, _ float32) { c.strokeRectCount++ }
func (c *mockCanvas) DrawRoundRect(_ geometry.Rect, _ widget.Color, _ float32) {
	c.drawRoundRectCount++
}
func (c *mockCanvas) StrokeRoundRect(_ geometry.Rect, _ widget.Color, _ float32, _ float32) {
	c.strokeRoundRectCount++
}
func (c *mockCanvas) DrawCircle(_ geometry.Point, _ float32, _ widget.Color) { c.drawCircleCount++ }
func (c *mockCanvas) StrokeCircle(_ geometry.Point, _ float32, _ widget.Color, _ float32) {
	c.strokeCircleCount++
}
func (c *mockCanvas) DrawLine(_, _ geometry.Point, _ widget.Color, _ float32) { c.drawLineCount++ }
func (c *mockCanvas) DrawText(text string, _ geometry.Rect, _ float32, color widget.Color, _ bool, _ widget.TextAlign) {
	c.drawTextCount++
	c.lastTextColor = color
	c.lastText = text
}

func (c *mockCanvas) MeasureText(text string, fontSize float32, _ bool) float32 {
	return float32(len([]rune(text))) * fontSize * 0.5
}
func (c *mockCanvas) DrawImage(_ image.Image, _ geometry.Point)    {}
func (c *mockCanvas) PushClip(_ geometry.Rect)                     { c.pushClipCount++ }
func (c *mockCanvas) PushClipRoundRect(_ geometry.Rect, _ float32) {}
func (c *mockCanvas) PopClip()                                     { c.popClipCount++ }
func (c *mockCanvas) PushTransform(_ geometry.Point)               { c.pushTransformCount++ }
func (c *mockCanvas) PopTransform()                                { c.popTransformCount++ }
func (c *mockCanvas) TransformOffset() geometry.Point              { return geometry.Point{} }

// eventConsumer is a mock widget that optionally consumes events.
type eventConsumer struct {
	widget.WidgetBase
	consume bool
	called  *bool
}

func (e *eventConsumer) Layout(_ widget.Context, c geometry.Constraints) geometry.Size {
	return c.Smallest()
}

func (e *eventConsumer) Draw(_ widget.Context, _ widget.Canvas) {}

func (e *eventConsumer) Event(_ widget.Context, _ event.Event) bool {
	*e.called = true
	return e.consume
}

func (e *eventConsumer) Children() []widget.Widget { return nil }
