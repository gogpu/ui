package primitives_test

import (
	"testing"

	"github.com/gogpu/ui/a11y"
	"github.com/gogpu/ui/event"
	"github.com/gogpu/ui/geometry"
	"github.com/gogpu/ui/primitives"
	"github.com/gogpu/ui/widget"
)

// --- ExpandedWidget construction ---

func TestExpandedIsExpanded(t *testing.T) {
	child := primitives.Box().Width(100).Height(50)
	exp := primitives.Expanded(child)

	if !exp.IsExpanded() {
		t.Error("Expanded widget should report IsExpanded() = true")
	}
}

func TestExpandedChild(t *testing.T) {
	child := primitives.Box().Width(100).Height(50)
	exp := primitives.Expanded(child)

	if exp.Child() != child {
		t.Error("Child() should return the wrapped widget")
	}
}

func TestExpandedIsVisibleAndEnabled(t *testing.T) {
	exp := primitives.Expanded(primitives.Box())
	if !exp.IsVisible() {
		t.Error("Expanded should be visible by default")
	}
	if !exp.IsEnabled() {
		t.Error("Expanded should be enabled by default")
	}
}

func TestExpandedChildren(t *testing.T) {
	child := primitives.Box()
	exp := primitives.Expanded(child)

	children := exp.Children()
	if len(children) != 1 {
		t.Fatalf("expected 1 child, got %d", len(children))
	}
	if children[0] != child {
		t.Error("Children()[0] should be the wrapped widget")
	}
}

// --- Expanded delegates Layout ---

func TestExpandedDelegatesLayout(t *testing.T) {
	child := primitives.Box().Width(80).Height(40)
	exp := primitives.Expanded(child)
	ctx := widget.NewContext()

	size := exp.Layout(ctx, geometry.Loose(geometry.Sz(200, 200)))
	if size.Width != 80 || size.Height != 40 {
		t.Errorf("expected 80x40, got %s", size)
	}
}

// --- Expanded delegates Draw ---

func TestExpandedDelegatesDraw(t *testing.T) {
	child := primitives.Box().Width(50).Height(30).Background(widget.ColorWhite)
	exp := primitives.Expanded(child)
	ctx := widget.NewContext()
	canvas := &mockCanvas{}

	_ = exp.Layout(ctx, geometry.Loose(geometry.Sz(200, 200)))
	exp.Draw(ctx, canvas)

	if canvas.pushTransformCount == 0 {
		t.Error("Expanded.Draw should call PushTransform for child positioning")
	}
	if canvas.popTransformCount == 0 {
		t.Error("Expanded.Draw should call PopTransform")
	}
}

func TestExpandedDrawInvisible(t *testing.T) {
	child := primitives.Box().Width(50).Height(30).Background(widget.ColorWhite)
	exp := primitives.Expanded(child)
	exp.SetVisible(false)
	ctx := widget.NewContext()
	canvas := &mockCanvas{}

	_ = exp.Layout(ctx, geometry.Loose(geometry.Sz(200, 200)))
	exp.Draw(ctx, canvas)

	if canvas.pushTransformCount != 0 {
		t.Error("invisible Expanded should not draw")
	}
}

// --- Expanded delegates Event ---

func TestExpandedDelegatesEvent(t *testing.T) {
	consumed := false
	child := &eventConsumer{consume: true, called: &consumed}
	exp := primitives.Expanded(child)
	ctx := widget.NewContext()

	result := exp.Event(ctx, &event.Base{})
	if !result || !consumed {
		t.Error("Expanded should delegate non-mouse events to child")
	}
}

func TestExpandedEventDisabled(t *testing.T) {
	consumed := false
	child := &eventConsumer{consume: true, called: &consumed}
	exp := primitives.Expanded(child)
	exp.SetEnabled(false)
	ctx := widget.NewContext()

	if exp.Event(ctx, &event.Base{}) {
		t.Error("disabled Expanded should not consume events")
	}
}

func TestExpandedEventInvisible(t *testing.T) {
	consumed := false
	child := &eventConsumer{consume: true, called: &consumed}
	exp := primitives.Expanded(child)
	exp.SetVisible(false)
	ctx := widget.NewContext()

	if exp.Event(ctx, &event.Base{}) {
		t.Error("invisible Expanded should not consume events")
	}
}

// --- Expanded accessibility delegation ---

func TestExpandedAccessibilityDelegatesToChild(t *testing.T) {
	// BoxWidget implements a11y.Accessible.
	child := primitives.Box().Label("inner")
	exp := primitives.Expanded(child)

	if exp.AccessibilityRole() != a11y.RoleGenericContainer {
		t.Errorf("expected RoleGenericContainer, got %s", exp.AccessibilityRole())
	}
	if exp.AccessibilityLabel() != "inner" {
		t.Errorf("expected 'inner', got %q", exp.AccessibilityLabel())
	}
	if exp.AccessibilityHint() != "" {
		t.Errorf("expected empty hint, got %q", exp.AccessibilityHint())
	}
	if exp.AccessibilityValue() != "" {
		t.Errorf("expected empty value, got %q", exp.AccessibilityValue())
	}
	if exp.AccessibilityActions() != nil {
		t.Error("expected nil actions")
	}
	st := exp.AccessibilityState()
	if st.Disabled || st.Hidden {
		t.Error("expected enabled and visible state")
	}
}

func TestExpandedAccessibilityNonAccessibleChild(t *testing.T) {
	// eventConsumer does not implement a11y.Accessible.
	consumed := false
	child := &eventConsumer{consume: false, called: &consumed}
	exp := primitives.Expanded(child)

	if exp.AccessibilityRole() != a11y.RoleUnknown {
		t.Errorf("expected RoleNone for non-accessible child, got %s", exp.AccessibilityRole())
	}
	if exp.AccessibilityLabel() != "" {
		t.Errorf("expected empty label, got %q", exp.AccessibilityLabel())
	}
}

// --- HBox with Expanded: two-pass layout ---

func TestHBoxExpandedCenter(t *testing.T) {
	left := primitives.Box().Width(56).Height(100)
	center := primitives.Box().Height(100)
	expanded := primitives.Expanded(center)
	right := primitives.Box().Width(40).Height(100)

	hbox := primitives.HBox(left, expanded, right)
	ctx := widget.NewContext()

	totalWidth := float32(400)
	size := hbox.Layout(ctx, geometry.Tight(geometry.Sz(totalWidth, 100)))

	if size.Width != totalWidth {
		t.Errorf("expected total width %f, got %f", totalWidth, size.Width)
	}

	// Center should get: 400 - 56 - 40 = 304px.
	leftBounds := left.Bounds()
	expandedBounds := expanded.Bounds()
	rightBounds := right.Bounds()

	if leftBounds.Width() != 56 {
		t.Errorf("left expected width 56, got %f", leftBounds.Width())
	}
	if expandedBounds.Width() != 304 {
		t.Errorf("center expected width 304, got %f", expandedBounds.Width())
	}
	if rightBounds.Width() != 40 {
		t.Errorf("right expected width 40, got %f", rightBounds.Width())
	}

	// Verify positions: left at 0, expanded at 56, right at 360.
	if leftBounds.Min.X != 0 {
		t.Errorf("left expected X=0, got %f", leftBounds.Min.X)
	}
	if expandedBounds.Min.X != 56 {
		t.Errorf("expanded expected X=56, got %f", expandedBounds.Min.X)
	}
	if rightBounds.Min.X != 360 {
		t.Errorf("right expected X=360, got %f", rightBounds.Min.X)
	}
}

func TestHBoxExpandedCenterWithGap(t *testing.T) {
	left := primitives.Box().Width(56).Height(100)
	center := primitives.Box().Height(100)
	right := primitives.Box().Width(40).Height(100)

	hbox := primitives.HBox(left, primitives.Expanded(center), right).Gap(4)
	ctx := widget.NewContext()

	totalWidth := float32(400)
	size := hbox.Layout(ctx, geometry.Tight(geometry.Sz(totalWidth, 100)))

	if size.Width != totalWidth {
		t.Errorf("expected total width %f, got %f", totalWidth, size.Width)
	}

	// Center should get: 400 - 56 - 40 - (2*4) = 296px.
	centerBounds := center.Bounds()
	if centerBounds.Width() != 296 {
		t.Errorf("center expected width 296, got %f", centerBounds.Width())
	}
}

func TestHBoxExpandedCenterWithPadding(t *testing.T) {
	left := primitives.Box().Width(56).Height(80)
	center := primitives.Box().Height(80)
	right := primitives.Box().Width(40).Height(80)

	hbox := primitives.HBox(left, primitives.Expanded(center), right).Padding(10)
	ctx := widget.NewContext()

	totalWidth := float32(400)
	size := hbox.Layout(ctx, geometry.Tight(geometry.Sz(totalWidth, 100)))

	if size.Width != totalWidth {
		t.Errorf("expected total width %f, got %f", totalWidth, size.Width)
	}

	// Child area = 400 - 20 (padding) = 380.
	// Center should get: 380 - 56 - 40 = 284px.
	centerBounds := center.Bounds()
	if centerBounds.Width() != 284 {
		t.Errorf("center expected width 284, got %f", centerBounds.Width())
	}

	// Left should start at padding offset.
	leftBounds := left.Bounds()
	if leftBounds.Min.X != 10 {
		t.Errorf("left expected X=10, got %f", leftBounds.Min.X)
	}
}

func TestHBoxMultipleExpanded(t *testing.T) {
	left := primitives.Box().Width(40).Height(100)
	mid1 := primitives.Box().Height(100)
	mid2 := primitives.Box().Height(100)
	right := primitives.Box().Width(60).Height(100)

	hbox := primitives.HBox(left, primitives.Expanded(mid1), primitives.Expanded(mid2), right)
	ctx := widget.NewContext()

	totalWidth := float32(400)
	size := hbox.Layout(ctx, geometry.Tight(geometry.Sz(totalWidth, 100)))

	if size.Width != totalWidth {
		t.Errorf("expected total width %f, got %f", totalWidth, size.Width)
	}

	// Remaining = 400 - 40 - 60 = 300, split equally = 150 each.
	mid1Bounds := mid1.Bounds()
	mid2Bounds := mid2.Bounds()

	if mid1Bounds.Width() != 150 {
		t.Errorf("mid1 expected width 150, got %f", mid1Bounds.Width())
	}
	if mid2Bounds.Width() != 150 {
		t.Errorf("mid2 expected width 150, got %f", mid2Bounds.Width())
	}
}

func TestHBoxNoExpandedUnchanged(t *testing.T) {
	c1 := primitives.Box().Width(50).Height(30)
	c2 := primitives.Box().Width(70).Height(30)
	hbox := primitives.HBox(c1, c2)
	ctx := widget.NewContext()
	size := hbox.Layout(ctx, geometry.Loose(geometry.Sz(400, 400)))

	// Should work exactly as before: 50 + 70 = 120.
	if size.Width != 120 {
		t.Errorf("expected width 120 (no expanded), got %f", size.Width)
	}
	if size.Height != 30 {
		t.Errorf("expected height 30, got %f", size.Height)
	}
}

// --- VBox with Expanded ---

func TestVBoxExpandedCenter(t *testing.T) {
	top := primitives.Box().Width(200).Height(40)
	center := primitives.Box().Width(200)
	expanded := primitives.Expanded(center)
	bottom := primitives.Box().Width(200).Height(28)

	vbox := primitives.VBox(top, expanded, bottom)
	ctx := widget.NewContext()

	totalHeight := float32(300)
	size := vbox.Layout(ctx, geometry.Tight(geometry.Sz(200, totalHeight)))

	if size.Height != totalHeight {
		t.Errorf("expected total height %f, got %f", totalHeight, size.Height)
	}

	// Center should get: 300 - 40 - 28 = 232px.
	expandedBounds := expanded.Bounds()
	if expandedBounds.Height() != 232 {
		t.Errorf("center expected height 232, got %f", expandedBounds.Height())
	}

	// Verify positions via the ExpandedWidget wrapper.
	topBounds := top.Bounds()
	bottomBounds := bottom.Bounds()

	if topBounds.Min.Y != 0 {
		t.Errorf("top expected Y=0, got %f", topBounds.Min.Y)
	}
	if expandedBounds.Min.Y != 40 {
		t.Errorf("expanded expected Y=40, got %f", expandedBounds.Min.Y)
	}
	if bottomBounds.Min.Y != 272 {
		t.Errorf("bottom expected Y=272, got %f", bottomBounds.Min.Y)
	}
}

func TestVBoxExpandedCenterWithGap(t *testing.T) {
	top := primitives.Box().Width(200).Height(40)
	center := primitives.Box().Width(200)
	bottom := primitives.Box().Width(200).Height(28)

	vbox := primitives.VBox(top, primitives.Expanded(center), bottom).Gap(4)
	ctx := widget.NewContext()

	totalHeight := float32(300)
	size := vbox.Layout(ctx, geometry.Tight(geometry.Sz(200, totalHeight)))

	if size.Height != totalHeight {
		t.Errorf("expected total height %f, got %f", totalHeight, size.Height)
	}

	// Center = 300 - 40 - 28 - (2 gaps * 4) = 224px.
	centerBounds := center.Bounds()
	if centerBounds.Height() != 224 {
		t.Errorf("center expected height 224, got %f", centerBounds.Height())
	}
}

func TestVBoxMultipleExpanded(t *testing.T) {
	top := primitives.Box().Width(200).Height(20)
	mid1 := primitives.Box().Width(200)
	mid2 := primitives.Box().Width(200)
	bottom := primitives.Box().Width(200).Height(20)

	vbox := primitives.VBox(top, primitives.Expanded(mid1), primitives.Expanded(mid2), bottom)
	ctx := widget.NewContext()

	totalHeight := float32(200)
	size := vbox.Layout(ctx, geometry.Tight(geometry.Sz(200, totalHeight)))

	if size.Height != totalHeight {
		t.Errorf("expected total height %f, got %f", totalHeight, size.Height)
	}

	// Remaining = 200 - 20 - 20 = 160, split = 80 each.
	mid1Bounds := mid1.Bounds()
	mid2Bounds := mid2.Bounds()

	if mid1Bounds.Height() != 80 {
		t.Errorf("mid1 expected height 80, got %f", mid1Bounds.Height())
	}
	if mid2Bounds.Height() != 80 {
		t.Errorf("mid2 expected height 80, got %f", mid2Bounds.Height())
	}
}

func TestVBoxNoExpandedUnchanged(t *testing.T) {
	c1 := primitives.Box().Width(100).Height(30)
	c2 := primitives.Box().Width(100).Height(40)
	vbox := primitives.VBox(c1, c2)
	ctx := widget.NewContext()
	size := vbox.Layout(ctx, geometry.Loose(geometry.Sz(400, 400)))

	if size.Width != 100 {
		t.Errorf("expected width 100 (no expanded), got %f", size.Width)
	}
	if size.Height != 70 {
		t.Errorf("expected height 70 (30+40), got %f", size.Height)
	}
}

// --- isExpanded detection ---

func TestIsExpandedOnRegularWidget(t *testing.T) {
	box := primitives.Box()
	// Regular widgets wrapped in HBox should not be detected as expanded.
	// We verify indirectly: HBox with only fixed widgets should use simple layout.
	hbox := primitives.HBox(box)
	ctx := widget.NewContext()
	size := hbox.Layout(ctx, geometry.Loose(geometry.Sz(200, 200)))
	if size.Width != 0 || size.Height != 0 {
		t.Errorf("expected 0x0, got %s", size)
	}
}

// --- Expanded with mouse event coordinate translation ---

func TestExpandedMouseEventTranslation(t *testing.T) {
	consumed := false
	child := &eventConsumer{consume: true, called: &consumed}
	exp := primitives.Expanded(child)
	ctx := widget.NewContext()

	// Set bounds so translation can happen.
	_ = exp.Layout(ctx, geometry.Loose(geometry.Sz(100, 100)))
	exp.SetBounds(geometry.FromPointSize(geometry.Pt(50, 50), geometry.Sz(100, 100)))

	me := &event.MouseEvent{
		Position:  geometry.Pt(75, 75),
		MouseType: event.MousePress,
	}

	result := exp.Event(ctx, me)
	if !result || !consumed {
		t.Error("Expanded should forward mouse events to child")
	}
}

// --- Edge case: Expanded child with insufficient space ---

func TestHBoxExpandedInsufficientSpace(t *testing.T) {
	// Fixed children exceed available space -> expanded gets 0.
	left := primitives.Box().Width(200).Height(50)
	center := primitives.Box().Height(50)
	right := primitives.Box().Width(200).Height(50)

	hbox := primitives.HBox(left, primitives.Expanded(center), right)
	ctx := widget.NewContext()

	size := hbox.Layout(ctx, geometry.Tight(geometry.Sz(300, 50)))

	// Fixed want 400 but only 300 available. Expanded gets 0.
	centerBounds := center.Bounds()
	if centerBounds.Width() != 0 {
		t.Errorf("center expected width 0 when space exhausted, got %f", centerBounds.Width())
	}

	// Total should still be constrained to 300.
	if size.Width != 300 {
		t.Errorf("expected width 300, got %f", size.Width)
	}
}

func TestVBoxExpandedInsufficientSpace(t *testing.T) {
	top := primitives.Box().Width(100).Height(200)
	center := primitives.Box().Width(100)
	bottom := primitives.Box().Width(100).Height(200)

	vbox := primitives.VBox(top, primitives.Expanded(center), bottom)
	ctx := widget.NewContext()

	size := vbox.Layout(ctx, geometry.Tight(geometry.Sz(100, 300)))

	centerBounds := center.Bounds()
	if centerBounds.Height() != 0 {
		t.Errorf("center expected height 0 when space exhausted, got %f", centerBounds.Height())
	}

	if size.Height != 300 {
		t.Errorf("expected height 300, got %f", size.Height)
	}
}

// --- Expanded in Draw + Event through BoxWidget ---

func TestHBoxExpandedDrawAndEvent(t *testing.T) {
	child := primitives.Box().Width(100).Height(50).Background(widget.ColorWhite)
	hbox := primitives.HBox(
		primitives.Box().Width(30).Height(50),
		primitives.Expanded(child),
	)
	ctx := widget.NewContext()
	canvas := &mockCanvas{}

	_ = hbox.Layout(ctx, geometry.Tight(geometry.Sz(200, 50)))
	hbox.Draw(ctx, canvas)

	// The HBox should draw children (including the Expanded wrapper).
	if canvas.pushTransformCount == 0 {
		t.Error("expected PushTransform during HBox draw")
	}
}
