package docking

import (
	"image"
	"testing"

	"github.com/gogpu/ui/event"
	"github.com/gogpu/ui/geometry"
	"github.com/gogpu/ui/widget"
)

// mockWidget is a minimal widget for testing.
type mockWidget struct {
	widget.WidgetBase
	layoutCalled bool
	drawCalled   bool
	eventCalled  bool
}

func newMockWidget() *mockWidget {
	w := &mockWidget{}
	w.SetVisible(true)
	w.SetEnabled(true)
	return w
}

func (m *mockWidget) Layout(ctx widget.Context, c geometry.Constraints) geometry.Size {
	m.layoutCalled = true
	return c.Constrain(geometry.Sz(100, 100))
}

func (m *mockWidget) Draw(_ widget.Context, _ widget.Canvas) {
	m.drawCalled = true
}

func (m *mockWidget) Event(_ widget.Context, _ event.Event) bool {
	m.eventCalled = true
	return false
}

// --- Zone tests ---

func TestZoneString(t *testing.T) {
	tests := []struct {
		zone Zone
		want string
	}{
		{Left, "Left"},
		{Right, "Right"},
		{Top, "Top"},
		{Bottom, "Bottom"},
		{Center, "Center"},
		{Zone(99), "Unknown"},
	}

	for _, tt := range tests {
		if got := tt.zone.String(); got != tt.want {
			t.Errorf("Zone(%d).String() = %q, want %q", tt.zone, got, tt.want)
		}
	}
}

func TestZoneIsEdge(t *testing.T) {
	tests := []struct {
		zone Zone
		want bool
	}{
		{Left, true},
		{Right, true},
		{Top, true},
		{Bottom, true},
		{Center, false},
	}

	for _, tt := range tests {
		if got := tt.zone.isEdge(); got != tt.want {
			t.Errorf("Zone(%d).isEdge() = %v, want %v", tt.zone, got, tt.want)
		}
	}
}

// --- group tests ---

func TestGroupEmpty(t *testing.T) {
	g := &group{}
	if !g.isEmpty() {
		t.Error("new group should be empty")
	}
	if p := g.activePanel(); p != nil {
		t.Error("active panel should be nil for empty group")
	}
}

func TestGroupAddPanel(t *testing.T) {
	g := &group{}
	p1 := NewPanel(PanelTitle("A"))
	p2 := NewPanel(PanelTitle("B"))

	g.addPanel(p1)
	if g.isEmpty() {
		t.Error("group should not be empty after add")
	}
	if g.activeIdx != 0 {
		t.Errorf("activeIdx = %d, want 0", g.activeIdx)
	}
	if g.activePanel() != p1 {
		t.Error("active panel should be p1")
	}

	g.addPanel(p2)
	if g.activeIdx != 1 {
		t.Errorf("activeIdx = %d, want 1", g.activeIdx)
	}
	if g.activePanel() != p2 {
		t.Error("active panel should be p2 after add")
	}
}

func TestGroupRemovePanel(t *testing.T) {
	g := &group{}
	p1 := NewPanel(PanelTitle("A"))
	p2 := NewPanel(PanelTitle("B"))
	p3 := NewPanel(PanelTitle("C"))

	g.addPanel(p1)
	g.addPanel(p2)
	g.addPanel(p3)

	// Remove middle panel while last is active.
	if !g.removePanel(p2) {
		t.Error("removePanel should return true")
	}
	if len(g.panels) != 2 {
		t.Errorf("panel count = %d, want 2", len(g.panels))
	}
	// Active index should adjust since it was 2, now max is 1.
	if g.activeIdx > 1 {
		t.Errorf("activeIdx = %d, want <= 1", g.activeIdx)
	}

	// Remove non-existent panel.
	if g.removePanel(p2) {
		t.Error("removePanel should return false for non-existent panel")
	}

	// Remove all panels.
	g.removePanel(p1)
	g.removePanel(p3)
	if !g.isEmpty() {
		t.Error("group should be empty after removing all panels")
	}
}

func TestGroupContainsPanel(t *testing.T) {
	g := &group{}
	p1 := NewPanel(PanelTitle("A"))
	p2 := NewPanel(PanelTitle("B"))

	g.addPanel(p1)
	if !g.containsPanel(p1) {
		t.Error("should contain p1")
	}
	if g.containsPanel(p2) {
		t.Error("should not contain p2")
	}
}

func TestGroupActivePanelOutOfRange(t *testing.T) {
	g := &group{activeIdx: 5}
	if p := g.activePanel(); p != nil {
		t.Error("active panel should be nil for out-of-range index")
	}
}

func TestGroupActivePanelNegative(t *testing.T) {
	g := &group{activeIdx: -1}
	if p := g.activePanel(); p != nil {
		t.Error("active panel should be nil for negative index")
	}
}

// --- Panel tests ---

func TestNewPanel(t *testing.T) {
	p := NewPanel(
		PanelTitle("Explorer"),
		Closeable(true),
	)

	if p.Title() != "Explorer" {
		t.Errorf("Title() = %q, want %q", p.Title(), "Explorer")
	}
	if !p.IsCloseable() {
		t.Error("IsCloseable() should be true")
	}
	if p.Content() != nil {
		t.Error("Content() should be nil when not set")
	}
}

func TestNewPanelWithContent(t *testing.T) {
	mw := newMockWidget()
	p := NewPanel(
		PanelTitle("Editor"),
		PanelContent(mw),
	)

	if p.Content() != mw {
		t.Error("Content() should return the mock widget")
	}
}

func TestNewPanelDefaults(t *testing.T) {
	p := NewPanel()
	if p.Title() != "" {
		t.Errorf("default Title() = %q, want empty", p.Title())
	}
	if p.IsCloseable() {
		t.Error("default IsCloseable() should be false")
	}
}

// --- Host construction tests ---

func TestNewHost(t *testing.T) {
	h := NewHost()
	if !h.IsVisible() {
		t.Error("host should be visible by default")
	}
	if !h.IsEnabled() {
		t.Error("host should be enabled by default")
	}
	if h.Children() != nil {
		t.Error("empty host should have nil children")
	}
}

func TestNewHostWithCenter(t *testing.T) {
	center := newMockWidget()
	h := NewHost(CenterContent(center))

	children := h.Children()
	if len(children) != 1 {
		t.Fatalf("expected 1 child, got %d", len(children))
	}
	if children[0] != center {
		t.Error("child should be center content")
	}
}

func TestNewHostWithRatios(t *testing.T) {
	h := NewHost(
		LeftRatio(0.3),
		RightRatio(0.15),
		TopRatio(0.1),
		BottomRatio(0.25),
	)

	if h.cfg.leftRatio != 0.3 {
		t.Errorf("leftRatio = %v, want 0.3", h.cfg.leftRatio)
	}
	if h.cfg.rightRatio != 0.15 {
		t.Errorf("rightRatio = %v, want 0.15", h.cfg.rightRatio)
	}
	if h.cfg.topRatio != 0.1 {
		t.Errorf("topRatio = %v, want 0.1", h.cfg.topRatio)
	}
	if h.cfg.bottomRatio != 0.25 {
		t.Errorf("bottomRatio = %v, want 0.25", h.cfg.bottomRatio)
	}
}

func TestNewHostWithPainter(t *testing.T) {
	p := DefaultPainter{}
	h := NewHost(PainterOpt(p))

	// Verify it doesn't panic; painter is set.
	_ = h
}

func TestNewHostWithColorScheme(t *testing.T) {
	cs := ZoneColorScheme{
		Border: widget.ColorRed,
	}
	h := NewHost(ColorSchemeOpt(cs))

	if h.cfg.colorScheme.Border != widget.ColorRed {
		t.Error("color scheme not applied")
	}
}

// --- Dock/Undock tests ---

func TestDockPanel(t *testing.T) {
	h := NewHost()
	p := NewPanel(PanelTitle("Explorer"))

	h.Dock(p, Left)

	if h.PanelCount(Left) != 1 {
		t.Errorf("PanelCount(Left) = %d, want 1", h.PanelCount(Left))
	}
	if h.ActivePanelIndex(Left) != 0 {
		t.Errorf("ActivePanelIndex(Left) = %d, want 0", h.ActivePanelIndex(Left))
	}
}

func TestDockMultiplePanelsToSameZone(t *testing.T) {
	h := NewHost()
	p1 := NewPanel(PanelTitle("Explorer"))
	p2 := NewPanel(PanelTitle("Outline"))

	h.Dock(p1, Left)
	h.Dock(p2, Left)

	if h.PanelCount(Left) != 2 {
		t.Errorf("PanelCount(Left) = %d, want 2", h.PanelCount(Left))
	}
	// Latest docked panel should be active.
	if h.ActivePanelIndex(Left) != 1 {
		t.Errorf("ActivePanelIndex(Left) = %d, want 1", h.ActivePanelIndex(Left))
	}
}

func TestDockNilPanel(t *testing.T) {
	h := NewHost()
	h.Dock(nil, Left) // Should not panic.
	if h.PanelCount(Left) != 0 {
		t.Error("nil panel should not be docked")
	}
}

func TestDockInvalidZone(t *testing.T) {
	h := NewHost()
	p := NewPanel(PanelTitle("Test"))
	h.Dock(p, Zone(99)) // Should not panic.
}

func TestDockMovesToNewZone(t *testing.T) {
	h := NewHost()
	p := NewPanel(PanelTitle("Explorer"))

	h.Dock(p, Left)
	h.Dock(p, Right)

	if h.PanelCount(Left) != 0 {
		t.Errorf("PanelCount(Left) = %d, want 0", h.PanelCount(Left))
	}
	if h.PanelCount(Right) != 1 {
		t.Errorf("PanelCount(Right) = %d, want 1", h.PanelCount(Right))
	}
}

func TestUndockPanel(t *testing.T) {
	h := NewHost()
	p := NewPanel(PanelTitle("Explorer"))

	h.Dock(p, Left)
	if !h.Undock(p) {
		t.Error("Undock should return true")
	}
	if h.PanelCount(Left) != 0 {
		t.Errorf("PanelCount(Left) = %d, want 0 after undock", h.PanelCount(Left))
	}
}

func TestUndockNilPanel(t *testing.T) {
	h := NewHost()
	if h.Undock(nil) {
		t.Error("Undock(nil) should return false")
	}
}

func TestUndockNotDocked(t *testing.T) {
	h := NewHost()
	p := NewPanel(PanelTitle("Explorer"))
	if h.Undock(p) {
		t.Error("Undock should return false for panel not docked")
	}
}

func TestMovePanel(t *testing.T) {
	h := NewHost()
	p := NewPanel(PanelTitle("Explorer"))

	h.Dock(p, Left)
	if !h.MovePanel(p, Bottom) {
		t.Error("MovePanel should return true")
	}

	if h.PanelCount(Left) != 0 {
		t.Errorf("PanelCount(Left) = %d, want 0", h.PanelCount(Left))
	}
	if h.PanelCount(Bottom) != 1 {
		t.Errorf("PanelCount(Bottom) = %d, want 1", h.PanelCount(Bottom))
	}
}

func TestMovePanelNotDocked(t *testing.T) {
	h := NewHost()
	p := NewPanel(PanelTitle("Explorer"))
	if h.MovePanel(p, Left) {
		t.Error("MovePanel should return false for panel not docked")
	}
}

func TestMovePanelNil(t *testing.T) {
	h := NewHost()
	if h.MovePanel(nil, Left) {
		t.Error("MovePanel(nil) should return false")
	}
}

func TestMovePanelInvalidZone(t *testing.T) {
	h := NewHost()
	p := NewPanel(PanelTitle("Test"))
	h.Dock(p, Left)
	if h.MovePanel(p, Zone(99)) {
		t.Error("MovePanel to invalid zone should return false")
	}
}

// --- ActivePanelIndex tests ---

func TestActivePanelIndex(t *testing.T) {
	h := NewHost()
	if h.ActivePanelIndex(Left) != -1 {
		t.Errorf("ActivePanelIndex for empty zone = %d, want -1", h.ActivePanelIndex(Left))
	}
	if h.ActivePanelIndex(Zone(99)) != -1 {
		t.Errorf("ActivePanelIndex for invalid zone = %d, want -1", h.ActivePanelIndex(Zone(99)))
	}
}

func TestSetActivePanelIndex(t *testing.T) {
	h := NewHost()
	p1 := NewPanel(PanelTitle("A"))
	p2 := NewPanel(PanelTitle("B"))

	h.Dock(p1, Left)
	h.Dock(p2, Left)

	h.SetActivePanelIndex(Left, 0)
	if h.ActivePanelIndex(Left) != 0 {
		t.Errorf("ActivePanelIndex = %d, want 0", h.ActivePanelIndex(Left))
	}
}

func TestSetActivePanelIndexOutOfRange(t *testing.T) {
	h := NewHost()
	p := NewPanel(PanelTitle("A"))
	h.Dock(p, Left)

	h.SetActivePanelIndex(Left, 5)     // Out of range.
	h.SetActivePanelIndex(Left, -1)    // Negative.
	h.SetActivePanelIndex(Zone(99), 0) // Invalid zone.
	// Should not panic; active stays at 0.
	if h.ActivePanelIndex(Left) != 0 {
		t.Errorf("ActivePanelIndex = %d, want 0", h.ActivePanelIndex(Left))
	}
}

// --- PanelZone tests ---

func TestPanelZone(t *testing.T) {
	h := NewHost()
	p := NewPanel(PanelTitle("Explorer"))

	h.Dock(p, Right)
	zone, ok := h.PanelZone(p)
	if !ok {
		t.Error("PanelZone should find docked panel")
	}
	if zone != Right {
		t.Errorf("PanelZone = %v, want Right", zone)
	}
}

func TestPanelZoneNotDocked(t *testing.T) {
	h := NewHost()
	p := NewPanel(PanelTitle("Explorer"))

	_, ok := h.PanelZone(p)
	if ok {
		t.Error("PanelZone should return false for panel not docked")
	}
}

func TestPanelZoneNil(t *testing.T) {
	h := NewHost()
	_, ok := h.PanelZone(nil)
	if ok {
		t.Error("PanelZone(nil) should return false")
	}
}

// --- PanelCount tests ---

func TestPanelCountInvalidZone(t *testing.T) {
	h := NewHost()
	if h.PanelCount(Zone(99)) != 0 {
		t.Error("PanelCount for invalid zone should be 0")
	}
}

// --- Layout tests ---

func TestHostLayout(t *testing.T) {
	center := newMockWidget()
	h := NewHost(CenterContent(center))
	h.SetBounds(geometry.NewRect(0, 0, 800, 600))

	ctx := widget.NewContext()
	constraints := geometry.Constraints{
		MinWidth: 800, MaxWidth: 800,
		MinHeight: 600, MaxHeight: 600,
	}

	size := h.Layout(ctx, constraints)
	if size.Width != 800 || size.Height != 600 {
		t.Errorf("Layout size = %v, want 800x600", size)
	}

	if !center.layoutCalled {
		t.Error("center content Layout should be called")
	}
}

func TestHostLayoutWithZones(t *testing.T) {
	center := newMockWidget()
	leftContent := newMockWidget()
	h := NewHost(CenterContent(center), LeftRatio(0.25))
	h.SetBounds(geometry.NewRect(0, 0, 800, 600))

	p := NewPanel(PanelTitle("Explorer"), PanelContent(leftContent))
	h.Dock(p, Left)

	ctx := widget.NewContext()
	constraints := geometry.Tight(geometry.Sz(800, 600))
	_ = h.Layout(ctx, constraints)

	if !leftContent.layoutCalled {
		t.Error("left panel content Layout should be called")
	}
	if !center.layoutCalled {
		t.Error("center content Layout should be called")
	}
}

func TestHostLayoutFallbackSize(t *testing.T) {
	h := NewHost()
	h.SetBounds(geometry.NewRect(0, 0, 0, 0))

	ctx := widget.NewContext()
	constraints := geometry.Constraints{MaxWidth: 0, MaxHeight: 0}
	size := h.Layout(ctx, constraints)

	if size.Width != defaultHostWidth || size.Height != defaultHostHeight {
		t.Errorf("fallback size = %v, want %vx%v", size, defaultHostWidth, defaultHostHeight)
	}
}

// --- Draw tests ---

func TestHostDraw(t *testing.T) {
	center := newMockWidget()
	h := NewHost(CenterContent(center))
	h.SetBounds(geometry.NewRect(0, 0, 800, 600))

	ctx := widget.NewContext()
	canvas := &mockCanvas{}
	h.Draw(ctx, canvas)

	if !center.drawCalled {
		t.Error("center content Draw should be called")
	}
}

func TestHostDrawEmptyBounds(t *testing.T) {
	h := NewHost()
	// Don't set bounds -- should return early without panic.
	ctx := widget.NewContext()
	canvas := &mockCanvas{}
	h.Draw(ctx, canvas)
}

func TestHostDrawWithZonePanels(t *testing.T) {
	leftContent := newMockWidget()
	h := NewHost()
	h.SetBounds(geometry.NewRect(0, 0, 800, 600))

	p := NewPanel(PanelTitle("Explorer"), PanelContent(leftContent))
	h.Dock(p, Left)

	// Layout first to set zone bounds.
	ctx := widget.NewContext()
	constraints := geometry.Tight(geometry.Sz(800, 600))
	_ = h.Layout(ctx, constraints)

	canvas := &mockCanvas{}
	h.Draw(ctx, canvas)

	if !leftContent.drawCalled {
		t.Error("left panel content Draw should be called")
	}
}

// --- Event tests ---

func TestHostEventForwardToCenter(t *testing.T) {
	center := newMockWidget()
	h := NewHost(CenterContent(center))
	h.SetBounds(geometry.NewRect(0, 0, 800, 600))

	ctx := widget.NewContext()
	me := &event.MouseEvent{
		MouseType: event.MouseMove,
		Position:  geometry.Pt(400, 300),
	}

	h.Event(ctx, me)
	if !center.eventCalled {
		t.Error("center content should receive events")
	}
}

func TestHostEventTabClick(t *testing.T) {
	h := NewHost(LeftRatio(0.25))
	h.SetBounds(geometry.NewRect(0, 0, 800, 600))

	p1 := NewPanel(PanelTitle("A"))
	p2 := NewPanel(PanelTitle("B"))
	h.Dock(p1, Left)
	h.Dock(p2, Left)

	ctx := widget.NewContext()
	ctx.SetOnInvalidate(func() {}) // Prevent nil panic.

	// Layout to compute zones.
	constraints := geometry.Tight(geometry.Sz(800, 600))
	_ = h.Layout(ctx, constraints)

	// Confirm p2 is active (last docked).
	if h.ActivePanelIndex(Left) != 1 {
		t.Fatalf("expected active panel 1, got %d", h.ActivePanelIndex(Left))
	}

	// Click on first tab.
	leftZoneBounds := h.zones[Left].bounds
	tabBarRect := zoneTabBarRect(leftZoneBounds)
	clickPos := geometry.Pt(
		tabBarRect.Min.X+10,
		tabBarRect.Min.Y+10,
	)

	me := &event.MouseEvent{
		MouseType: event.MousePress,
		Button:    event.ButtonLeft,
		Position:  clickPos,
	}

	consumed := h.Event(ctx, me)
	if !consumed {
		t.Error("tab click should be consumed")
	}
	if h.ActivePanelIndex(Left) != 0 {
		t.Errorf("after click, active panel = %d, want 0", h.ActivePanelIndex(Left))
	}
}

func TestHostEventCloseButton(t *testing.T) {
	var closedPanel *Panel
	var closedZone Zone
	h := NewHost(
		LeftRatio(0.25),
		OnPanelClose(func(p *Panel, z Zone) {
			closedPanel = p
			closedZone = z
		}),
	)
	h.SetBounds(geometry.NewRect(0, 0, 800, 600))

	p := NewPanel(PanelTitle("Explorer"), Closeable(true))
	h.Dock(p, Left)

	ctx := widget.NewContext()
	ctx.SetOnInvalidate(func() {})

	// Layout to compute tab states.
	constraints := geometry.Tight(geometry.Sz(800, 600))
	_ = h.Layout(ctx, constraints)

	// Find close button position.
	if len(h.tabStates[Left]) == 0 {
		t.Fatal("expected tab states for left zone")
	}
	cb := h.tabStates[Left][0].CloseButtonBounds
	if cb.IsEmpty() {
		t.Fatal("close button bounds should not be empty for closeable panel")
	}

	clickPos := geometry.Pt(
		(cb.Min.X+cb.Max.X)/2,
		(cb.Min.Y+cb.Max.Y)/2,
	)

	me := &event.MouseEvent{
		MouseType: event.MousePress,
		Button:    event.ButtonLeft,
		Position:  clickPos,
	}

	consumed := h.Event(ctx, me)
	if !consumed {
		t.Error("close button click should be consumed")
	}
	if closedPanel != p {
		t.Error("onPanelClose should receive the closed panel")
	}
	if closedZone != Left {
		t.Errorf("onPanelClose zone = %v, want Left", closedZone)
	}
	if h.PanelCount(Left) != 0 {
		t.Error("panel should be removed after close")
	}
}

func TestHostEventMouseMove(t *testing.T) {
	h := NewHost(LeftRatio(0.25))
	h.SetBounds(geometry.NewRect(0, 0, 800, 600))

	p := NewPanel(PanelTitle("Test"))
	h.Dock(p, Left)

	ctx := widget.NewContext()
	ctx.SetOnInvalidate(func() {})

	constraints := geometry.Tight(geometry.Sz(800, 600))
	_ = h.Layout(ctx, constraints)

	// Move over tab.
	leftBounds := h.zones[Left].bounds
	tabBarRect := zoneTabBarRect(leftBounds)
	me := &event.MouseEvent{
		MouseType: event.MouseMove,
		Position:  geometry.Pt(tabBarRect.Min.X+10, tabBarRect.Min.Y+10),
	}

	h.Event(ctx, me)

	// Verify no panic on move over tab area.
	_ = h.tabStates[Left]
}

func TestHostEventMouseLeave(t *testing.T) {
	h := NewHost(LeftRatio(0.25))
	h.SetBounds(geometry.NewRect(0, 0, 800, 600))

	p := NewPanel(PanelTitle("Test"))
	h.Dock(p, Left)

	ctx := widget.NewContext()
	ctx.SetOnInvalidate(func() {})

	constraints := geometry.Tight(geometry.Sz(800, 600))
	_ = h.Layout(ctx, constraints)

	// Set a tab as hovered.
	if len(h.tabStates[Left]) > 0 {
		h.tabStates[Left][0].Hovered = true
	}

	me := &event.MouseEvent{
		MouseType: event.MouseLeave,
		Position:  geometry.Pt(-1, -1),
	}
	h.Event(ctx, me)

	// After leave, hover should be cleared.
	if len(h.tabStates[Left]) > 0 && h.tabStates[Left][0].Hovered {
		t.Error("tab should not be hovered after mouse leave")
	}
}

func TestHostEventRightButtonIgnored(t *testing.T) {
	h := NewHost(LeftRatio(0.25))
	h.SetBounds(geometry.NewRect(0, 0, 800, 600))

	p := NewPanel(PanelTitle("Test"))
	h.Dock(p, Left)

	ctx := widget.NewContext()
	ctx.SetOnInvalidate(func() {})

	constraints := geometry.Tight(geometry.Sz(800, 600))
	_ = h.Layout(ctx, constraints)

	leftBounds := h.zones[Left].bounds
	tabBarRect := zoneTabBarRect(leftBounds)

	me := &event.MouseEvent{
		MouseType: event.MousePress,
		Button:    event.ButtonRight,
		Position:  geometry.Pt(tabBarRect.Min.X+10, tabBarRect.Min.Y+10),
	}

	consumed := h.Event(ctx, me)
	if consumed {
		t.Error("right button should not be consumed by tab bar")
	}
}

func TestHostEventNonMouseEvent(t *testing.T) {
	h := NewHost()
	h.SetBounds(geometry.NewRect(0, 0, 800, 600))

	ctx := widget.NewContext()
	ke := &event.KeyEvent{
		KeyType: event.KeyPress,
		Key:     event.KeyA,
	}

	consumed := h.Event(ctx, ke)
	if consumed {
		t.Error("key event should not be consumed by empty host")
	}
}

// --- Children tests ---

func TestHostChildren(t *testing.T) {
	center := newMockWidget()
	leftContent := newMockWidget()
	rightContent := newMockWidget()

	h := NewHost(CenterContent(center))
	p1 := NewPanel(PanelTitle("Left"), PanelContent(leftContent))
	p2 := NewPanel(PanelTitle("Right"), PanelContent(rightContent))

	h.Dock(p1, Left)
	h.Dock(p2, Right)

	children := h.Children()
	if len(children) != 3 {
		t.Errorf("expected 3 children, got %d", len(children))
	}
}

func TestHostChildrenNoPanelContent(t *testing.T) {
	h := NewHost()
	p := NewPanel(PanelTitle("No Content")) // No PanelContent set.
	h.Dock(p, Left)

	children := h.Children()
	if children != nil {
		t.Error("panel without content should not add to children")
	}
}

// --- computeZoneRects tests ---

func TestComputeZoneRectsNoEdgeZones(t *testing.T) {
	h := NewHost()
	rects := h.computeZoneRects(geometry.Sz(800, 600), geometry.Pt(0, 0))

	// All edge zones are empty.
	if !rects[Left].IsEmpty() {
		t.Error("left zone should be empty when no panels")
	}
	if !rects[Right].IsEmpty() {
		t.Error("right zone should be empty when no panels")
	}
	// Top and bottom have zero height but full width, so they aren't truly "empty" in Rect terms.
	// Just check center takes all space.
	if rects[Center].Width() != 800 {
		t.Errorf("center width = %v, want 800", rects[Center].Width())
	}
	if rects[Center].Height() != 600 {
		t.Errorf("center height = %v, want 600", rects[Center].Height())
	}
}

func TestComputeZoneRectsWithEdgeZones(t *testing.T) {
	h := NewHost(LeftRatio(0.25), BottomRatio(0.3))

	// Dock panels so zones are not empty.
	h.Dock(NewPanel(PanelTitle("L")), Left)
	h.Dock(NewPanel(PanelTitle("B")), Bottom)

	rects := h.computeZoneRects(geometry.Sz(800, 600), geometry.Pt(0, 0))

	leftW := 800 * 0.25
	bottomH := 600 * 0.3

	if diff := rects[Left].Width() - float32(leftW); diff > 0.1 || diff < -0.1 {
		t.Errorf("left width = %v, want %v", rects[Left].Width(), leftW)
	}
	if diff := rects[Bottom].Height() - float32(bottomH); diff > 0.1 || diff < -0.1 {
		t.Errorf("bottom height = %v, want %v", rects[Bottom].Height(), bottomH)
	}

	// Center should be reduced.
	expectedCenterW := float32(800) - float32(leftW)
	if diff := rects[Center].Width() - expectedCenterW; diff > 0.1 || diff < -0.1 {
		t.Errorf("center width = %v, want %v", rects[Center].Width(), expectedCenterW)
	}
}

func TestComputeZoneRectsAllEdges(t *testing.T) {
	h := NewHost(
		LeftRatio(0.15),
		RightRatio(0.15),
		TopRatio(0.1),
		BottomRatio(0.1),
	)

	h.Dock(NewPanel(PanelTitle("L")), Left)
	h.Dock(NewPanel(PanelTitle("R")), Right)
	h.Dock(NewPanel(PanelTitle("T")), Top)
	h.Dock(NewPanel(PanelTitle("B")), Bottom)

	rects := h.computeZoneRects(geometry.Sz(1000, 800), geometry.Pt(10, 20))

	// Top and Bottom span full width.
	if rects[Top].Width() != 1000 {
		t.Errorf("top width = %v, want 1000", rects[Top].Width())
	}
	if rects[Bottom].Width() != 1000 {
		t.Errorf("bottom width = %v, want 1000", rects[Bottom].Width())
	}

	// Center should be what's left.
	centerW := rects[Center].Width()
	leftW := rects[Left].Width()
	rightW := rects[Right].Width()
	total := centerW + leftW + rightW
	if diff := total - 1000; diff > 1 || diff < -1 {
		t.Errorf("center+left+right width = %v, want ~1000", total)
	}
}

// --- Painter tests ---

func TestDefaultPainterPaintZoneTabs(t *testing.T) {
	p := DefaultPainter{}
	canvas := &mockCanvas{}

	tabs := []ZoneTabState{
		{Title: "A", Active: true, Bounds: geometry.NewRect(0, 0, 100, 32)},
		{Title: "B", Active: false, Bounds: geometry.NewRect(100, 0, 100, 32)},
	}

	p.PaintZoneTabs(canvas, ZoneTabsPaintState{
		Zone:         Left,
		TabBarBounds: geometry.NewRect(0, 0, 200, 32),
		Tabs:         tabs,
		ActiveIdx:    0,
	})

	if canvas.drawRectCount < 1 {
		t.Error("PaintZoneTabs should draw at least one rect")
	}
}

func TestDefaultPainterPaintZoneTabsEmpty(t *testing.T) {
	p := DefaultPainter{}
	canvas := &mockCanvas{}

	// Empty bounds should not draw.
	p.PaintZoneTabs(canvas, ZoneTabsPaintState{
		TabBarBounds: geometry.Rect{},
	})

	if canvas.drawRectCount != 0 {
		t.Error("PaintZoneTabs with empty bounds should not draw")
	}
}

func TestDefaultPainterPaintZoneTabsWithColorScheme(t *testing.T) {
	p := DefaultPainter{}
	canvas := &mockCanvas{}

	cs := ZoneColorScheme{
		TabBarBackground: widget.ColorBlue,
		ActiveTabText:    widget.ColorWhite,
	}

	tabs := []ZoneTabState{
		{Title: "A", Active: true, Bounds: geometry.NewRect(0, 0, 100, 32)},
	}

	p.PaintZoneTabs(canvas, ZoneTabsPaintState{
		Zone:         Left,
		TabBarBounds: geometry.NewRect(0, 0, 200, 32),
		Tabs:         tabs,
		ActiveIdx:    0,
		ColorScheme:  cs,
	})

	// Should use color scheme colors without panicking.
}

func TestDefaultPainterPaintZoneBorder(t *testing.T) {
	p := DefaultPainter{}
	canvas := &mockCanvas{}

	p.PaintZoneBorder(canvas, geometry.NewRect(100, 0, 1, 600), Left)

	if canvas.drawRectCount != 1 {
		t.Errorf("PaintZoneBorder drawRectCount = %d, want 1", canvas.drawRectCount)
	}
}

func TestDefaultPainterPaintZoneBorderEmpty(t *testing.T) {
	p := DefaultPainter{}
	canvas := &mockCanvas{}

	p.PaintZoneBorder(canvas, geometry.Rect{}, Left)

	if canvas.drawRectCount != 0 {
		t.Error("PaintZoneBorder with empty rect should not draw")
	}
}

func TestDefaultPainterTabWithCloseButton(t *testing.T) {
	p := DefaultPainter{}
	canvas := &mockCanvas{}

	tabs := []ZoneTabState{
		{
			Title:             "A",
			Active:            true,
			Closeable:         true,
			Bounds:            geometry.NewRect(0, 0, 150, 32),
			CloseButtonBounds: geometry.NewRect(120, 8, 14, 14),
		},
	}

	p.PaintZoneTabs(canvas, ZoneTabsPaintState{
		Zone:         Left,
		TabBarBounds: geometry.NewRect(0, 0, 200, 32),
		Tabs:         tabs,
		ActiveIdx:    0,
	})

	// Should draw close button lines.
	if canvas.drawLineCount < 2 {
		t.Errorf("drawLineCount = %d, want >= 2 for close button", canvas.drawLineCount)
	}
}

func TestDefaultPainterTabHovered(t *testing.T) {
	p := DefaultPainter{}
	canvas := &mockCanvas{}

	tabs := []ZoneTabState{
		{Title: "A", Active: false, Hovered: true, Bounds: geometry.NewRect(0, 0, 100, 32)},
	}

	p.PaintZoneTabs(canvas, ZoneTabsPaintState{
		Zone:         Left,
		TabBarBounds: geometry.NewRect(0, 0, 200, 32),
		Tabs:         tabs,
		ActiveIdx:    -1,
	})

	// Should draw hover background.
	if canvas.drawRectCount < 2 {
		t.Error("hovered tab should draw background + hover")
	}
}

func TestDefaultPainterTabEmptyBounds(t *testing.T) {
	canvas := &mockCanvas{}
	ts := &ZoneTabState{Title: "A", Active: true}
	paintDefaultZoneTab(canvas, ts, false, ZoneColorScheme{})

	if canvas.drawRectCount != 0 {
		t.Error("tab with empty bounds should not draw")
	}
}

// --- Zone helper function tests ---

func TestZoneTabBarRect(t *testing.T) {
	zoneBounds := geometry.NewRect(10, 20, 200, 300)
	tabBar := zoneTabBarRect(zoneBounds)

	if tabBar.Min.X != 10 || tabBar.Min.Y != 20 {
		t.Errorf("tab bar origin = (%v, %v), want (10, 20)", tabBar.Min.X, tabBar.Min.Y)
	}
	if tabBar.Width() != 200 {
		t.Errorf("tab bar width = %v, want 200", tabBar.Width())
	}
	if tabBar.Height() != zoneTabBarHeight {
		t.Errorf("tab bar height = %v, want %v", tabBar.Height(), zoneTabBarHeight)
	}
}

func TestZoneContentRect(t *testing.T) {
	zoneBounds := geometry.NewRect(10, 20, 200, 300)
	content := zoneContentRect(zoneBounds)

	expectedY := float32(20 + zoneTabBarHeight)
	if content.Min.Y != expectedY {
		t.Errorf("content Y = %v, want %v", content.Min.Y, expectedY)
	}
	expectedH := float32(300 - zoneTabBarHeight)
	if diff := content.Height() - expectedH; diff > 0.1 || diff < -0.1 {
		t.Errorf("content height = %v, want %v", content.Height(), expectedH)
	}
}

func TestZoneContentRectTooSmall(t *testing.T) {
	zoneBounds := geometry.NewRect(10, 20, 200, 10) // Smaller than tab bar.
	content := zoneContentRect(zoneBounds)

	if content.Height() != 0 {
		t.Errorf("content height = %v, want 0 when zone is too small", content.Height())
	}
}

func TestZoneBorderRect(t *testing.T) {
	bounds := geometry.NewRect(0, 0, 200, 600)

	leftBorder := zoneBorderRect(bounds, Left)
	if leftBorder.Min.X != bounds.Max.X-zoneBorderWidth {
		t.Error("left border should be at right edge of zone")
	}

	rightBorder := zoneBorderRect(bounds, Right)
	if rightBorder.Min.X != bounds.Min.X {
		t.Error("right border should be at left edge of zone")
	}

	topBorder := zoneBorderRect(bounds, Top)
	if topBorder.Min.Y != bounds.Max.Y-zoneBorderWidth {
		t.Error("top border should be at bottom edge of zone")
	}

	bottomBorder := zoneBorderRect(bounds, Bottom)
	if bottomBorder.Min.Y != bounds.Min.Y {
		t.Error("bottom border should be at top edge of zone")
	}

	centerBorder := zoneBorderRect(bounds, Center)
	if !centerBorder.IsEmpty() {
		t.Error("center zone should have no border")
	}
}

// --- clampRatio tests ---

func TestClampRatio(t *testing.T) {
	tests := []struct {
		in   float32
		want float32
	}{
		{0.5, 0.5},
		{0.0, 0.0},
		{1.0, 1.0},
		{-0.5, 0.0},
		{1.5, 1.0},
	}

	for _, tt := range tests {
		if got := clampRatio(tt.in); got != tt.want {
			t.Errorf("clampRatio(%v) = %v, want %v", tt.in, got, tt.want)
		}
	}
}

// --- Widget interface compliance ---

func TestHostImplementsWidget(t *testing.T) {
	var _ widget.Widget = (*Host)(nil)
}

// --- mockCanvas ---

type mockCanvas struct {
	drawRectCount      int
	drawLineCount      int
	drawTextCount      int
	drawRoundRectCount int
	pushClipCount      int
	popClipCount       int
}

func (c *mockCanvas) Clear(_ widget.Color)                                  {}
func (c *mockCanvas) DrawRect(_ geometry.Rect, _ widget.Color)              { c.drawRectCount++ }
func (c *mockCanvas) StrokeRect(_ geometry.Rect, _ widget.Color, _ float32) {}
func (c *mockCanvas) DrawRoundRect(_ geometry.Rect, _ widget.Color, _ float32) {
	c.drawRoundRectCount++
}
func (c *mockCanvas) StrokeRoundRect(_ geometry.Rect, _ widget.Color, _ float32, _ float32) {}
func (c *mockCanvas) DrawCircle(_ geometry.Point, _ float32, _ widget.Color)                {}
func (c *mockCanvas) StrokeCircle(_ geometry.Point, _ float32, _ widget.Color, _ float32)   {}
func (c *mockCanvas) DrawLine(_, _ geometry.Point, _ widget.Color, _ float32)               { c.drawLineCount++ }
func (c *mockCanvas) DrawText(_ string, _ geometry.Rect, _ float32, _ widget.Color, _ bool, _ widget.TextAlign) {
	c.drawTextCount++
}

func (c *mockCanvas) MeasureText(text string, fontSize float32, _ bool) float32 {
	return float32(len([]rune(text))) * fontSize * 0.5
}
func (c *mockCanvas) DrawImage(_ image.Image, _ geometry.Point)    {}
func (c *mockCanvas) PushClip(_ geometry.Rect)                     { c.pushClipCount++ }
func (c *mockCanvas) PushClipRoundRect(_ geometry.Rect, _ float32) {}
func (c *mockCanvas) PopClip()                                     { c.popClipCount++ }
func (c *mockCanvas) PushTransform(_ geometry.Point)               {}
func (c *mockCanvas) PopTransform()                                {}
func (c *mockCanvas) TransformOffset() geometry.Point              { return geometry.Point{} }
