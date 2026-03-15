package docking

import "github.com/gogpu/ui/geometry"

// Zone identifies a docking zone within the host layout.
type Zone uint8

// Zone constants.
const (
	// Left is the left edge zone.
	Left Zone = iota

	// Right is the right edge zone.
	Right

	// Top is the top edge zone.
	Top

	// Bottom is the bottom edge zone.
	Bottom

	// Center is the center zone (main content).
	Center
)

// zoneNames maps each Zone to its human-readable name.
var zoneNames = [...]string{
	Left:   zoneLeftStr,
	Right:  zoneRightStr,
	Top:    zoneTopStr,
	Bottom: zoneBottomStr,
	Center: zoneCenterStr,
}

// String constants for Zone.String to satisfy goconst.
const (
	zoneLeftStr    = "Left"
	zoneRightStr   = "Right"
	zoneTopStr     = "Top"
	zoneBottomStr  = "Bottom"
	zoneCenterStr  = "Center"
	zoneUnknownStr = "Unknown"
)

// String returns a human-readable name for the zone.
func (z Zone) String() string {
	if int(z) < len(zoneNames) {
		return zoneNames[z]
	}
	return zoneUnknownStr
}

// isEdge reports whether the zone is an edge zone (not center).
func (z Zone) isEdge() bool {
	return z != Center
}

// zoneCount is the total number of zones.
const zoneCount = 5

// group holds the panels docked to a single zone and tracks the active tab.
type group struct {
	panels    []*Panel
	activeIdx int
	bounds    geometry.Rect
}

// isEmpty reports whether the group has no panels.
func (g *group) isEmpty() bool {
	return len(g.panels) == 0
}

// activePanel returns the currently active panel, or nil if empty.
func (g *group) activePanel() *Panel {
	if g.isEmpty() || g.activeIdx < 0 || g.activeIdx >= len(g.panels) {
		return nil
	}
	return g.panels[g.activeIdx]
}

// addPanel appends a panel to the group and makes it active.
func (g *group) addPanel(p *Panel) {
	g.panels = append(g.panels, p)
	g.activeIdx = len(g.panels) - 1
}

// removePanel removes a panel from the group by identity.
// Returns true if found and removed.
func (g *group) removePanel(p *Panel) bool {
	for i, existing := range g.panels {
		if existing != p {
			continue
		}
		// Remove preserving order.
		copy(g.panels[i:], g.panels[i+1:])
		g.panels[len(g.panels)-1] = nil // Clear reference for GC.
		g.panels = g.panels[:len(g.panels)-1]

		// Adjust active index.
		if g.activeIdx >= len(g.panels) && len(g.panels) > 0 {
			g.activeIdx = len(g.panels) - 1
		}
		if g.isEmpty() {
			g.activeIdx = 0
		}
		return true
	}
	return false
}

// containsPanel reports whether the group contains the given panel.
func (g *group) containsPanel(p *Panel) bool {
	for _, existing := range g.panels {
		if existing == p {
			return true
		}
	}
	return false
}
