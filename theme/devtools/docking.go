package devtools

import (
	"github.com/gogpu/ui/core/docking"
	"github.com/gogpu/ui/geometry"
	"github.com/gogpu/ui/widget"
)

// DockingPainter renders docking zones and tab bars using DevTools design tokens.
// DevTools docking uses a tool window style: compact tab strip at bottom/side,
// 24px tabs with 12px font, Surface background for panels, SurfaceElevated for
// tab bars, and a Blue6 underline (2px) for the selected tab, matching JetBrains
// IDE tool window tab styling.
//
// If Theme is nil, DockingPainter falls back to the default DevTools dark palette.
type DockingPainter struct {
	Theme *Theme // nil uses default DevTools dark fallback
}

// resolveColors returns DevTools-derived colors for docking painting.
func (p DockingPainter) resolveColors() docking.ZoneColorScheme {
	if p.Theme == nil {
		return dtDefaultDockingColors
	}
	cs := p.Theme.Colors
	return docking.ZoneColorScheme{
		TabBarBackground:    cs.SurfaceElevated,
		ActiveTabText:       cs.OnSurface,
		InactiveTabText:     cs.OnSurfaceSecondary,
		ActiveTabBackground: cs.Primary,
		HoverBackground:     cs.ControlHover,
		Border:              cs.Border,
		CloseButton:         cs.OnSurfaceSecondary,
	}
}

// PaintZoneTabs renders the tab header bar using DevTools elevated surface and accent underline.
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
		dtPaintDockingTab(canvas, ts, colors)
	}

	// Active tab indicator (Blue6 underline, 2px).
	if ps.ActiveIdx >= 0 && ps.ActiveIdx < len(ps.Tabs) {
		active := &ps.Tabs[ps.ActiveIdx]
		indicatorBounds := geometry.NewRect(
			active.Bounds.Min.X,
			active.Bounds.Max.Y-dtDockingIndicatorHeight,
			active.Bounds.Width(),
			dtDockingIndicatorHeight,
		)
		canvas.DrawRect(indicatorBounds, colors.ActiveTabBackground)
	}
}

// PaintZoneBorder renders the border between a zone and the center.
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

// dtPaintDockingTab draws a single tab in a zone's tab header.
func dtPaintDockingTab(canvas widget.Canvas, ts *docking.ZoneTabState, colors docking.ZoneColorScheme) {
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
			ts.Bounds.Width()-dtDockingCloseSize-dtDockingClosePadding,
			ts.Bounds.Height(),
		)
	}

	canvas.DrawText(ts.Title, textBounds, dtDockingTabFontSize, fg, ts.Active, widget.TextAlignCenter)

	// Close button.
	if ts.Closeable {
		cb := ts.CloseButtonBounds
		if !cb.IsEmpty() {
			centerX := (cb.Min.X + cb.Max.X) / 2
			centerY := (cb.Min.Y + cb.Max.Y) / 2
			halfSize := dtDockingCloseSize / 4
			canvas.DrawLine(
				geometry.Pt(centerX-halfSize, centerY-halfSize),
				geometry.Pt(centerX+halfSize, centerY+halfSize),
				colors.CloseButton, dtDockingCloseStroke,
			)
			canvas.DrawLine(
				geometry.Pt(centerX+halfSize, centerY-halfSize),
				geometry.Pt(centerX-halfSize, centerY+halfSize),
				colors.CloseButton, dtDockingCloseStroke,
			)
		}
	}
}

// dtDefaultDockingColors holds the default DevTools dark color scheme for docking.
var dtDefaultDockingColors = docking.ZoneColorScheme{
	TabBarBackground:    widget.Hex(0x393B40), // Gray3 (elevated surface)
	ActiveTabText:       widget.Hex(0xDFE1E5), // Gray12 (on-surface)
	InactiveTabText:     widget.Hex(0x9DA0A8), // Gray9 (secondary)
	ActiveTabBackground: widget.Hex(0x3574F0), // Blue6 (primary)
	HoverBackground:     widget.Hex(0x43454A), // Gray4 (control hover)
	Border:              widget.Hex(0x393B40), // Gray3 (border)
	CloseButton:         widget.Hex(0x9DA0A8), // Gray9 (secondary)
}

// DevTools docking drawing constants.
const (
	dtDockingTabFontSize     float32 = 12
	dtDockingIndicatorHeight float32 = 2
	dtDockingCloseSize       float32 = 14
	dtDockingClosePadding    float32 = 6
	dtDockingCloseStroke     float32 = 1.5
)

// Compile-time check that DockingPainter implements Painter.
var _ docking.Painter = DockingPainter{}
