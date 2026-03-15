package docking

import (
	"github.com/gogpu/ui/geometry"
	"github.com/gogpu/ui/widget"
)

// Painter draws the visual representation of the docking system.
// Each design system (Material 3, Fluent, Cupertino) provides its own
// Painter implementation to render zone borders and tab headers.
//
// If no Painter is set, the host uses [DefaultPainter].
type Painter interface {
	// PaintZoneTabs renders the tab header bar for a zone group.
	PaintZoneTabs(canvas widget.Canvas, state ZoneTabsPaintState)

	// PaintZoneBorder renders the border between a zone and the center.
	PaintZoneBorder(canvas widget.Canvas, borderRect geometry.Rect, zone Zone)
}

// ZoneTabsPaintState provides the current zone tab bar state to the painter.
type ZoneTabsPaintState struct {
	// Zone identifies which zone this tab bar belongs to.
	Zone Zone

	// TabBarBounds is the tab bar area.
	TabBarBounds geometry.Rect

	// Tabs contains the state of each tab in the zone.
	Tabs []ZoneTabState

	// ActiveIdx is the index of the currently active tab.
	ActiveIdx int

	// ColorScheme provides theme-derived colors for painting.
	// Zero value means the painter should use its built-in defaults.
	ColorScheme ZoneColorScheme
}

// ZoneTabState provides the state of a single tab within a zone.
type ZoneTabState struct {
	// Title is the panel's display text.
	Title string

	// Bounds is the tab's clickable area within the tab bar.
	Bounds geometry.Rect

	// Active is true when this tab is the currently selected tab.
	Active bool

	// Hovered is true when the mouse is over this tab.
	Hovered bool

	// Closeable is true when this tab shows a close button.
	Closeable bool

	// CloseButtonBounds is the clickable area for the close button.
	// Empty if not closeable.
	CloseButtonBounds geometry.Rect
}

// ZoneColorScheme provides theme-derived colors for zone painting.
// Zero value means the painter should use its built-in defaults.
type ZoneColorScheme struct {
	TabBarBackground    widget.Color
	ActiveTabText       widget.Color
	InactiveTabText     widget.Color
	ActiveTabBackground widget.Color
	HoverBackground     widget.Color
	Border              widget.Color
	CloseButton         widget.Color
}

// DefaultPainter provides a minimal fallback painter with no design system styling.
// It draws simple tab headers and zone borders suitable for testing and prototyping.
type DefaultPainter struct{}

// PaintZoneTabs renders a minimal tab header bar for a zone.
func (p DefaultPainter) PaintZoneTabs(canvas widget.Canvas, ps ZoneTabsPaintState) {
	if ps.TabBarBounds.IsEmpty() {
		return
	}

	hasScheme := ps.ColorScheme != (ZoneColorScheme{})

	// Tab bar background.
	bg := defaultTabBarBgColor
	if hasScheme {
		bg = ps.ColorScheme.TabBarBackground
	}
	canvas.DrawRect(ps.TabBarBounds, bg)

	// Draw each tab.
	for i := range ps.Tabs {
		ts := &ps.Tabs[i]
		paintDefaultZoneTab(canvas, ts, hasScheme, ps.ColorScheme)
	}

	// Draw indicator under active tab.
	if ps.ActiveIdx >= 0 && ps.ActiveIdx < len(ps.Tabs) {
		active := &ps.Tabs[ps.ActiveIdx]
		indicatorBounds := geometry.NewRect(
			active.Bounds.Min.X,
			active.Bounds.Max.Y-zoneIndicatorHeight,
			active.Bounds.Width(),
			zoneIndicatorHeight,
		)
		indicatorColor := defaultZoneIndicatorColor
		if hasScheme {
			indicatorColor = ps.ColorScheme.ActiveTabBackground
		}
		canvas.DrawRect(indicatorBounds, indicatorColor)
	}
}

// PaintZoneBorder renders a simple 1px border between a zone and the center.
func (p DefaultPainter) PaintZoneBorder(canvas widget.Canvas, borderRect geometry.Rect, _ Zone) {
	if borderRect.IsEmpty() {
		return
	}
	canvas.DrawRect(borderRect, defaultZoneBorderColor)
}

// paintDefaultZoneTab draws a single tab in a zone's tab header.
func paintDefaultZoneTab(canvas widget.Canvas, ts *ZoneTabState, hasScheme bool, cs ZoneColorScheme) {
	if ts.Bounds.IsEmpty() {
		return
	}

	// Hover background.
	if ts.Hovered {
		hoverBg := defaultZoneTabHoverBg
		if hasScheme {
			hoverBg = cs.HoverBackground
		}
		canvas.DrawRect(ts.Bounds, hoverBg)
	}

	// Tab text.
	fg := defaultZoneInactiveText
	if ts.Active {
		fg = defaultZoneActiveText
	}
	if hasScheme {
		fg = cs.InactiveTabText
		if ts.Active {
			fg = cs.ActiveTabText
		}
	}

	textBounds := ts.Bounds
	if ts.Closeable {
		// Shrink text area to leave room for close button.
		textBounds = geometry.NewRect(
			ts.Bounds.Min.X,
			ts.Bounds.Min.Y,
			ts.Bounds.Width()-zoneCloseButtonSize-zoneCloseButtonPadding,
			ts.Bounds.Height(),
		)
	}

	canvas.DrawText(ts.Title, textBounds, zoneTabFontSize, fg, ts.Active, widget.TextAlignCenter)

	// Close button.
	if ts.Closeable {
		closeFg := defaultZoneCloseButtonColor
		if hasScheme {
			closeFg = cs.CloseButton
		}
		cb := ts.CloseButtonBounds
		if !cb.IsEmpty() {
			centerX := (cb.Min.X + cb.Max.X) / 2
			centerY := (cb.Min.Y + cb.Max.Y) / 2
			halfSize := zoneCloseButtonSize / 4
			canvas.DrawLine(
				geometry.Pt(centerX-halfSize, centerY-halfSize),
				geometry.Pt(centerX+halfSize, centerY+halfSize),
				closeFg, 1.5,
			)
			canvas.DrawLine(
				geometry.Pt(centerX+halfSize, centerY-halfSize),
				geometry.Pt(centerX-halfSize, centerY+halfSize),
				closeFg, 1.5,
			)
		}
	}
}

// Default colors for DefaultPainter.
var (
	defaultTabBarBgColor        = widget.RGBA(0.93, 0.93, 0.93, 1.0)
	defaultZoneActiveText       = widget.RGBA(0.1, 0.1, 0.1, 1.0)
	defaultZoneInactiveText     = widget.RGBA(0.45, 0.45, 0.45, 1.0)
	defaultZoneIndicatorColor   = widget.Hex(0x6750A4)
	defaultZoneTabHoverBg       = widget.RGBA(0.0, 0.0, 0.0, 0.04)
	defaultZoneCloseButtonColor = widget.RGBA(0.45, 0.45, 0.45, 1.0)
	defaultZoneBorderColor      = widget.RGBA(0.80, 0.80, 0.80, 1.0)
)

// Layout constants for zone tab bars.
const (
	zoneTabBarHeight       float32 = 32
	zoneTabFontSize        float32 = 12
	zoneTabMinWidth        float32 = 80
	zoneTabMaxWidth        float32 = 200
	zoneTabPaddingX        float32 = 12
	zoneCloseButtonSize    float32 = 14
	zoneCloseButtonPadding float32 = 6
	zoneIndicatorHeight    float32 = 2
	zoneBorderWidth        float32 = 1
)
