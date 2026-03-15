package docking

import (
	"github.com/gogpu/ui/event"
	"github.com/gogpu/ui/geometry"
	"github.com/gogpu/ui/widget"
)

// HostOption configures a dock host during construction.
type HostOption func(*hostConfig)

// hostConfig holds the dock host's configuration.
type hostConfig struct {
	centerContent widget.Widget
	leftRatio     float32
	rightRatio    float32
	topRatio      float32
	bottomRatio   float32
	painter       Painter
	colorScheme   ZoneColorScheme
	onPanelClose  func(panel *Panel, zone Zone)
}

// CenterContent sets the main content widget displayed in the center zone.
func CenterContent(w widget.Widget) HostOption {
	return func(c *hostConfig) { c.centerContent = w }
}

// LeftRatio sets the size ratio of the left zone (0.0 to 1.0).
// Default is 0.2 (20% of available width).
func LeftRatio(r float32) HostOption {
	return func(c *hostConfig) { c.leftRatio = clampRatio(r) }
}

// RightRatio sets the size ratio of the right zone (0.0 to 1.0).
// Default is 0.2 (20% of available width).
func RightRatio(r float32) HostOption {
	return func(c *hostConfig) { c.rightRatio = clampRatio(r) }
}

// TopRatio sets the size ratio of the top zone (0.0 to 1.0).
// Default is 0.2 (20% of available height).
func TopRatio(r float32) HostOption {
	return func(c *hostConfig) { c.topRatio = clampRatio(r) }
}

// BottomRatio sets the size ratio of the bottom zone (0.0 to 1.0).
// Default is 0.2 (20% of available height).
func BottomRatio(r float32) HostOption {
	return func(c *hostConfig) { c.bottomRatio = clampRatio(r) }
}

// PainterOpt sets the painter used to render zone borders and tab headers.
// If not set, [DefaultPainter] is used.
func PainterOpt(p Painter) HostOption {
	return func(c *hostConfig) { c.painter = p }
}

// ColorSchemeOpt sets the theme-derived color scheme for the dock host.
func ColorSchemeOpt(cs ZoneColorScheme) HostOption {
	return func(c *hostConfig) { c.colorScheme = cs }
}

// OnPanelClose sets the callback invoked when a panel's close button is clicked.
// The callback receives the panel and the zone it belonged to.
func OnPanelClose(fn func(panel *Panel, zone Zone)) HostOption {
	return func(c *hostConfig) { c.onPanelClose = fn }
}

// Default values.
const (
	defaultEdgeRatio float32 = 0.2
)

// Host is the root container managing the dock layout.
//
// It arranges docked panels into five zones (Left, Right, Top, Bottom, Center)
// and renders tab headers for zones with multiple panels.
//
// A host is created with [NewHost] using functional options:
//
//	host := docking.NewHost(
//	    docking.CenterContent(mainEditor),
//	    docking.LeftRatio(0.25),
//	)
type Host struct {
	widget.WidgetBase
	cfg     hostConfig
	painter Painter

	// Zone groups, indexed by Zone constant.
	zones [zoneCount]group

	// Tab interaction state per zone.
	tabStates [zoneCount][]ZoneTabState
}

// NewHost creates a new dock host with the given options.
//
// The returned widget is visible and enabled by default.
func NewHost(opts ...HostOption) *Host {
	h := &Host{
		painter: DefaultPainter{},
	}
	h.SetVisible(true)
	h.SetEnabled(true)

	h.cfg.leftRatio = defaultEdgeRatio
	h.cfg.rightRatio = defaultEdgeRatio
	h.cfg.topRatio = defaultEdgeRatio
	h.cfg.bottomRatio = defaultEdgeRatio

	for _, opt := range opts {
		opt(&h.cfg)
	}

	if h.cfg.painter != nil {
		h.painter = h.cfg.painter
	}

	return h
}

// Dock adds a panel to the specified zone.
// If the zone already has panels, the new panel becomes an additional tab.
// The new panel becomes the active tab in its zone.
func (h *Host) Dock(panel *Panel, zone Zone) {
	if panel == nil {
		return
	}
	if int(zone) >= zoneCount {
		return
	}

	// Remove from any existing zone first.
	h.undockFromAll(panel)

	h.zones[zone].addPanel(panel)
}

// Undock removes a panel from its current zone.
// Returns true if the panel was found and removed.
func (h *Host) Undock(panel *Panel) bool {
	if panel == nil {
		return false
	}
	return h.undockFromAll(panel)
}

// MovePanel moves a panel from its current zone to a new zone.
// Returns true if the panel was found and moved.
func (h *Host) MovePanel(panel *Panel, targetZone Zone) bool {
	if panel == nil || int(targetZone) >= zoneCount {
		return false
	}

	// Find current zone.
	found := false
	for z := Zone(0); z < zoneCount; z++ {
		if h.zones[z].containsPanel(panel) {
			found = true
			break
		}
	}
	if !found {
		return false
	}

	h.undockFromAll(panel)
	h.zones[targetZone].addPanel(panel)
	return true
}

// PanelCount returns the number of panels in the given zone.
func (h *Host) PanelCount(zone Zone) int {
	if int(zone) >= zoneCount {
		return 0
	}
	return len(h.zones[zone].panels)
}

// ActivePanelIndex returns the active tab index for the given zone.
// Returns -1 if the zone is empty.
func (h *Host) ActivePanelIndex(zone Zone) int {
	if int(zone) >= zoneCount || h.zones[zone].isEmpty() {
		return -1
	}
	return h.zones[zone].activeIdx
}

// SetActivePanelIndex sets the active tab index for the given zone.
// Does nothing if the index is out of range.
func (h *Host) SetActivePanelIndex(zone Zone, idx int) {
	if int(zone) >= zoneCount {
		return
	}
	g := &h.zones[zone]
	if idx < 0 || idx >= len(g.panels) {
		return
	}
	g.activeIdx = idx
}

// PanelZone returns the zone a panel is docked to, and whether it was found.
func (h *Host) PanelZone(panel *Panel) (Zone, bool) {
	if panel == nil {
		return 0, false
	}
	for z := Zone(0); z < zoneCount; z++ {
		if h.zones[z].containsPanel(panel) {
			return z, true
		}
	}
	return 0, false
}

// Layout calculates zone sizes and positions all children.
func (h *Host) Layout(ctx widget.Context, constraints geometry.Constraints) geometry.Size {
	totalSize := constraints.Constrain(geometry.Sz(constraints.MaxWidth, constraints.MaxHeight))
	if totalSize.Width <= 0 || totalSize.Height <= 0 {
		totalSize = geometry.Sz(defaultHostWidth, defaultHostHeight)
	}

	origin := h.Bounds().Min
	rects := h.computeZoneRects(totalSize, origin)

	// Layout zone contents.
	for z := Zone(0); z < zoneCount; z++ {
		h.zones[z].bounds = rects[z]
		h.layoutZone(ctx, z, rects[z])
	}

	// Layout center content.
	if h.cfg.centerContent != nil {
		centerRect := rects[Center]
		cc := geometry.Tight(centerRect.Size())
		h.cfg.centerContent.Layout(ctx, cc)
		if setter, ok := h.cfg.centerContent.(interface{ SetBounds(geometry.Rect) }); ok {
			setter.SetBounds(centerRect)
		}
	}

	return totalSize
}

// Draw renders the dock layout to the canvas.
func (h *Host) Draw(ctx widget.Context, canvas widget.Canvas) {
	bounds := h.Bounds()
	if bounds.IsEmpty() {
		return
	}

	// Draw center content first (background).
	if h.cfg.centerContent != nil {
		h.cfg.centerContent.Draw(ctx, canvas)
	}

	// Draw edge zones on top of center.
	for z := Zone(0); z < zoneCount; z++ {
		if z == Center || h.zones[z].isEmpty() {
			continue
		}
		h.drawZone(ctx, canvas, z)
	}
}

// Event handles input events for the dock host.
func (h *Host) Event(ctx widget.Context, e event.Event) bool {
	// Check edge zone tab bars for click events.
	if consumed := h.handleZoneTabEvents(ctx, e); consumed {
		return true
	}

	// Forward to active panel content in edge zones.
	for z := Zone(0); z < zoneCount; z++ {
		if z == Center {
			continue
		}
		panel := h.zones[z].activePanel()
		if panel != nil && panel.Content() != nil {
			if panel.Content().Event(ctx, e) {
				return true
			}
		}
	}

	// Forward to center content.
	if h.cfg.centerContent != nil {
		if h.cfg.centerContent.Event(ctx, e) {
			return true
		}
	}

	return false
}

// Children returns all visible content widgets (zone panels + center).
func (h *Host) Children() []widget.Widget {
	var children []widget.Widget

	// Center content.
	if h.cfg.centerContent != nil {
		children = append(children, h.cfg.centerContent)
	}

	// Active panel content from each zone.
	for z := Zone(0); z < zoneCount; z++ {
		if z == Center {
			continue
		}
		panel := h.zones[z].activePanel()
		if panel != nil && panel.Content() != nil {
			children = append(children, panel.Content())
		}
	}

	if len(children) == 0 {
		return nil
	}
	return children
}

// undockFromAll removes a panel from any zone it is in.
func (h *Host) undockFromAll(panel *Panel) bool {
	for z := Zone(0); z < zoneCount; z++ {
		if h.zones[z].removePanel(panel) {
			return true
		}
	}
	return false
}

// layoutZone lays out the active panel's content within the zone rectangle.
func (h *Host) layoutZone(ctx widget.Context, z Zone, zoneRect geometry.Rect) {
	if z == Center || h.zones[z].isEmpty() {
		return
	}

	// Update tab states.
	h.updateTabStates(z)

	// Content area is below the tab bar.
	contentRect := zoneContentRect(zoneRect)
	panel := h.zones[z].activePanel()
	if panel == nil || panel.Content() == nil {
		return
	}

	cc := geometry.Tight(contentRect.Size())
	panel.Content().Layout(ctx, cc)
	if setter, ok := panel.Content().(interface{ SetBounds(geometry.Rect) }); ok {
		setter.SetBounds(contentRect)
	}
}

// drawZone renders a single edge zone: background, tab header, border, and content.
func (h *Host) drawZone(ctx widget.Context, canvas widget.Canvas, z Zone) {
	g := &h.zones[z]
	if g.isEmpty() || g.bounds.IsEmpty() {
		return
	}

	// Zone background.
	canvas.DrawRect(g.bounds, defaultTabBarBgColor)

	// Tab header bar.
	tabBarRect := zoneTabBarRect(g.bounds)
	h.updateTabStates(z)
	h.painter.PaintZoneTabs(canvas, ZoneTabsPaintState{
		Zone:         z,
		TabBarBounds: tabBarRect,
		Tabs:         h.tabStates[z],
		ActiveIdx:    g.activeIdx,
		ColorScheme:  h.cfg.colorScheme,
	})

	// Content.
	contentRect := zoneContentRect(g.bounds)
	if !contentRect.IsEmpty() {
		canvas.PushClip(contentRect)
		panel := g.activePanel()
		if panel != nil && panel.Content() != nil {
			panel.Content().Draw(ctx, canvas)
		}
		canvas.PopClip()
	}

	// Border between zone and center.
	borderRect := zoneBorderRect(g.bounds, z)
	h.painter.PaintZoneBorder(canvas, borderRect, z)
}

// handleZoneTabEvents checks mouse events against zone tab headers.
func (h *Host) handleZoneTabEvents(ctx widget.Context, e event.Event) bool {
	me, ok := e.(*event.MouseEvent)
	if !ok {
		return false
	}

	switch me.MouseType {
	case event.MousePress:
		return h.handleTabPress(ctx, me)
	case event.MouseMove:
		return h.handleTabMove(ctx, me)
	case event.MouseLeave:
		return h.handleTabLeave(ctx)
	default:
		return false
	}
}

// handleTabPress handles mouse clicks on zone tab headers.
func (h *Host) handleTabPress(ctx widget.Context, me *event.MouseEvent) bool {
	if me.Button != event.ButtonLeft {
		return false
	}

	for z := Zone(0); z < zoneCount; z++ {
		if z == Center || h.zones[z].isEmpty() {
			continue
		}

		tabBarRect := zoneTabBarRect(h.zones[z].bounds)
		if !tabBarRect.Contains(me.Position) {
			continue
		}

		// Check close buttons first.
		for i := range h.tabStates[z] {
			ts := &h.tabStates[z][i]
			if !ts.Closeable || ts.CloseButtonBounds.IsEmpty() {
				continue
			}
			if ts.CloseButtonBounds.Contains(me.Position) {
				h.closePanel(ctx, z, i)
				return true
			}
		}

		// Check tab selection.
		for i := range h.tabStates[z] {
			ts := &h.tabStates[z][i]
			if ts.Bounds.Contains(me.Position) {
				h.zones[z].activeIdx = i
				ctx.Invalidate()
				return true
			}
		}

		return true // Consumed by tab bar area.
	}

	return false
}

// handleTabMove updates hover state for zone tabs.
func (h *Host) handleTabMove(ctx widget.Context, me *event.MouseEvent) bool {
	changed := false

	for z := Zone(0); z < zoneCount; z++ {
		if z == Center || h.zones[z].isEmpty() {
			continue
		}

		for i := range h.tabStates[z] {
			ts := &h.tabStates[z][i]
			wasHovered := ts.Hovered
			ts.Hovered = ts.Bounds.Contains(me.Position)
			if ts.Hovered != wasHovered {
				changed = true
			}
		}
	}

	if changed {
		ctx.Invalidate()
	}
	return false // Don't consume move events.
}

// handleTabLeave clears all hover states.
func (h *Host) handleTabLeave(ctx widget.Context) bool {
	changed := false
	for z := Zone(0); z < zoneCount; z++ {
		for i := range h.tabStates[z] {
			if h.tabStates[z][i].Hovered {
				h.tabStates[z][i].Hovered = false
				changed = true
			}
		}
	}
	if changed {
		ctx.Invalidate()
	}
	return false
}

// closePanel removes the panel at index idx from zone z.
func (h *Host) closePanel(ctx widget.Context, z Zone, idx int) {
	g := &h.zones[z]
	if idx < 0 || idx >= len(g.panels) {
		return
	}

	panel := g.panels[idx]
	g.removePanel(panel)

	if h.cfg.onPanelClose != nil {
		h.cfg.onPanelClose(panel, z)
	}

	ctx.Invalidate()
}

// updateTabStates refreshes tab states for a zone from the current panels.
func (h *Host) updateTabStates(z Zone) {
	g := &h.zones[z]
	panelCount := len(g.panels)

	// Resize tab states slice if needed.
	if len(h.tabStates[z]) != panelCount {
		h.tabStates[z] = make([]ZoneTabState, panelCount)
	}

	if panelCount == 0 {
		return
	}

	// Compute tab bounds.
	tabBarRect := zoneTabBarRect(g.bounds)
	tabWidth := tabBarRect.Width() / float32(panelCount)
	if tabWidth > zoneTabMaxWidth {
		tabWidth = zoneTabMaxWidth
	}
	if tabWidth < zoneTabMinWidth && tabBarRect.Width() >= zoneTabMinWidth {
		tabWidth = zoneTabMinWidth
	}

	for i, panel := range g.panels {
		x := tabBarRect.Min.X + float32(i)*tabWidth
		ts := &h.tabStates[z][i]
		ts.Title = panel.Title()
		ts.Active = i == g.activeIdx
		// Hovered state is preserved from event handling -- no assignment needed.
		ts.Closeable = panel.IsCloseable()
		ts.Bounds = geometry.NewRect(x, tabBarRect.Min.Y, tabWidth, zoneTabBarHeight)

		// Close button bounds.
		if ts.Closeable {
			cbX := x + tabWidth - zoneTabPaddingX - zoneCloseButtonSize
			cbY := tabBarRect.Min.Y + (zoneTabBarHeight-zoneCloseButtonSize)/2
			ts.CloseButtonBounds = geometry.NewRect(cbX, cbY, zoneCloseButtonSize, zoneCloseButtonSize)
		} else {
			ts.CloseButtonBounds = geometry.Rect{}
		}
	}
}

// computeZoneRects calculates the bounding rectangle for each zone.
// The layout follows a border layout pattern:
// Top and Bottom span the full width.
// Left and Right fill the remaining height between Top and Bottom.
// Center takes whatever is left.
func (h *Host) computeZoneRects(totalSize geometry.Size, origin geometry.Point) [zoneCount]geometry.Rect {
	var rects [zoneCount]geometry.Rect

	totalW := totalSize.Width
	totalH := totalSize.Height

	// Top zone.
	topH := float32(0)
	if !h.zones[Top].isEmpty() {
		topH = totalH * h.cfg.topRatio
	}
	rects[Top] = geometry.NewRect(origin.X, origin.Y, totalW, topH)

	// Bottom zone.
	bottomH := float32(0)
	if !h.zones[Bottom].isEmpty() {
		bottomH = totalH * h.cfg.bottomRatio
	}
	rects[Bottom] = geometry.NewRect(origin.X, origin.Y+totalH-bottomH, totalW, bottomH)

	// Middle area (between top and bottom).
	middleY := origin.Y + topH
	middleH := totalH - topH - bottomH
	if middleH < 0 {
		middleH = 0
	}

	// Left zone.
	leftW := float32(0)
	if !h.zones[Left].isEmpty() {
		leftW = totalW * h.cfg.leftRatio
	}
	rects[Left] = geometry.NewRect(origin.X, middleY, leftW, middleH)

	// Right zone.
	rightW := float32(0)
	if !h.zones[Right].isEmpty() {
		rightW = totalW * h.cfg.rightRatio
	}
	rects[Right] = geometry.NewRect(origin.X+totalW-rightW, middleY, rightW, middleH)

	// Center zone: remaining space.
	centerX := origin.X + leftW
	centerW := totalW - leftW - rightW
	if centerW < 0 {
		centerW = 0
	}
	rects[Center] = geometry.NewRect(centerX, middleY, centerW, middleH)

	return rects
}

// zoneTabBarRect returns the tab bar rectangle at the top of a zone.
func zoneTabBarRect(zoneBounds geometry.Rect) geometry.Rect {
	return geometry.NewRect(
		zoneBounds.Min.X,
		zoneBounds.Min.Y,
		zoneBounds.Width(),
		zoneTabBarHeight,
	)
}

// zoneContentRect returns the content area below the tab bar.
func zoneContentRect(zoneBounds geometry.Rect) geometry.Rect {
	contentH := zoneBounds.Height() - zoneTabBarHeight
	if contentH < 0 {
		contentH = 0
	}
	return geometry.NewRect(
		zoneBounds.Min.X,
		zoneBounds.Min.Y+zoneTabBarHeight,
		zoneBounds.Width(),
		contentH,
	)
}

// zoneBorderRect returns the border rectangle between a zone and the center.
func zoneBorderRect(zoneBounds geometry.Rect, z Zone) geometry.Rect {
	switch z {
	case Left:
		return geometry.NewRect(
			zoneBounds.Max.X-zoneBorderWidth,
			zoneBounds.Min.Y,
			zoneBorderWidth,
			zoneBounds.Height(),
		)
	case Right:
		return geometry.NewRect(
			zoneBounds.Min.X,
			zoneBounds.Min.Y,
			zoneBorderWidth,
			zoneBounds.Height(),
		)
	case Top:
		return geometry.NewRect(
			zoneBounds.Min.X,
			zoneBounds.Max.Y-zoneBorderWidth,
			zoneBounds.Width(),
			zoneBorderWidth,
		)
	case Bottom:
		return geometry.NewRect(
			zoneBounds.Min.X,
			zoneBounds.Min.Y,
			zoneBounds.Width(),
			zoneBorderWidth,
		)
	default:
		return geometry.Rect{}
	}
}

// clampRatio clamps a ratio to the valid range [0, 1].
func clampRatio(r float32) float32 {
	if r < 0 {
		return 0
	}
	if r > 1 {
		return 1
	}
	return r
}

// Default host dimensions used as fallback.
const (
	defaultHostWidth  float32 = 800
	defaultHostHeight float32 = 600
)

// Verify Host implements required interfaces at compile time.
var _ widget.Widget = (*Host)(nil)
