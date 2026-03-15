package material3

import (
	"github.com/gogpu/ui/core/docking"
	"github.com/gogpu/ui/geometry"
	"github.com/gogpu/ui/widget"
)

// DockingPainter renders docking zones and tab bars using Material 3 design tokens.
// It maps M3 color roles to docking elements: surface for zones, outline for borders,
// and surface-variant for the tab bar background.
//
// If Theme is nil, DockingPainter falls back to the default M3 purple palette.
type DockingPainter struct {
	Theme *Theme // nil uses default M3 purple fallback
}

// resolveColors returns M3-derived colors for docking painting.
func (p DockingPainter) resolveColors() docking.ZoneColorScheme {
	if p.Theme == nil {
		return m3DefaultDockingColors
	}
	cs := p.Theme.Colors
	return docking.ZoneColorScheme{
		TabBarBackground:    cs.SurfaceContainerHigh,
		ActiveTabText:       cs.OnSurface,
		InactiveTabText:     cs.OnSurfaceVariant,
		ActiveTabBackground: cs.Primary,
		HoverBackground:     cs.OnSurface.WithAlpha(0.04),
		Border:              cs.OutlineVariant,
		CloseButton:         cs.OnSurfaceVariant,
	}
}

// PaintZoneTabs renders the tab header bar using M3 surface-variant and primary accent.
func (p DockingPainter) PaintZoneTabs(canvas widget.Canvas, ps docking.ZoneTabsPaintState) {
	if ps.TabBarBounds.IsEmpty() {
		return
	}

	colors := p.effectiveColors(ps.ColorScheme)

	// Tab bar background.
	canvas.DrawRect(ps.TabBarBounds, colors.TabBarBackground)

	// Draw each tab.
	for i := range ps.Tabs {
		ts := &ps.Tabs[i]
		m3PaintDockingTab(canvas, ts, colors)
	}

	// Active tab indicator.
	if ps.ActiveIdx >= 0 && ps.ActiveIdx < len(ps.Tabs) {
		active := &ps.Tabs[ps.ActiveIdx]
		indicatorBounds := geometry.NewRect(
			active.Bounds.Min.X,
			active.Bounds.Max.Y-m3DockingIndicatorHeight,
			active.Bounds.Width(),
			m3DockingIndicatorHeight,
		)
		canvas.DrawRect(indicatorBounds, colors.ActiveTabBackground)
	}
}

// PaintZoneBorder renders the border between a zone and the center using M3 outline-variant.
func (p DockingPainter) PaintZoneBorder(canvas widget.Canvas, borderRect geometry.Rect, _ docking.Zone) {
	if borderRect.IsEmpty() {
		return
	}
	colors := p.resolveColors()
	canvas.DrawRect(borderRect, colors.Border)
}

// effectiveColors returns colors, preferring the paint state's ColorScheme.
func (p DockingPainter) effectiveColors(cs docking.ZoneColorScheme) docking.ZoneColorScheme {
	if cs != (docking.ZoneColorScheme{}) {
		return cs
	}
	return p.resolveColors()
}

// m3PaintDockingTab draws a single tab in a zone's tab header.
func m3PaintDockingTab(canvas widget.Canvas, ts *docking.ZoneTabState, colors docking.ZoneColorScheme) {
	if ts.Bounds.IsEmpty() {
		return
	}

	// Hover background.
	if ts.Hovered {
		canvas.DrawRect(ts.Bounds, colors.HoverBackground)
	}

	// Tab text color.
	fg := colors.InactiveTabText
	if ts.Active {
		fg = colors.ActiveTabText
	}

	textBounds := ts.Bounds
	if ts.Closeable {
		textBounds = geometry.NewRect(
			ts.Bounds.Min.X,
			ts.Bounds.Min.Y,
			ts.Bounds.Width()-m3DockingCloseSize-m3DockingClosePadding,
			ts.Bounds.Height(),
		)
	}

	canvas.DrawText(ts.Title, textBounds, m3DockingTabFontSize, fg, ts.Active, widget.TextAlignCenter)

	// Close button.
	if ts.Closeable {
		cb := ts.CloseButtonBounds
		if !cb.IsEmpty() {
			centerX := (cb.Min.X + cb.Max.X) / 2
			centerY := (cb.Min.Y + cb.Max.Y) / 2
			halfSize := m3DockingCloseSize / 4
			canvas.DrawLine(
				geometry.Pt(centerX-halfSize, centerY-halfSize),
				geometry.Pt(centerX+halfSize, centerY+halfSize),
				colors.CloseButton, m3DockingCloseStroke,
			)
			canvas.DrawLine(
				geometry.Pt(centerX+halfSize, centerY-halfSize),
				geometry.Pt(centerX-halfSize, centerY+halfSize),
				colors.CloseButton, m3DockingCloseStroke,
			)
		}
	}
}

// m3DefaultDockingColors holds default M3 purple fallback colors for docking.
var m3DefaultDockingColors = docking.ZoneColorScheme{
	TabBarBackground:    widget.Hex(0xE6E0E9), // M3 surface-container-high
	ActiveTabText:       widget.Hex(0x1C1B1F), // M3 on-surface
	InactiveTabText:     widget.Hex(0x49454F), // M3 on-surface-variant
	ActiveTabBackground: widget.Hex(0x6750A4), // M3 primary
	HoverBackground:     widget.RGBA(0.12, 0.12, 0.13, 0.04),
	Border:              widget.Hex(0xCAC4D0), // M3 outline-variant
	CloseButton:         widget.Hex(0x49454F), // M3 on-surface-variant
}

// M3 docking drawing constants.
const (
	m3DockingTabFontSize     float32 = 12
	m3DockingIndicatorHeight float32 = 2
	m3DockingCloseSize       float32 = 14
	m3DockingClosePadding    float32 = 6
	m3DockingCloseStroke     float32 = 1.5
)

// Compile-time check that DockingPainter implements Painter.
var _ docking.Painter = DockingPainter{}
