package tabview

import (
	"github.com/gogpu/ui/geometry"
	"github.com/gogpu/ui/widget"
)

// Painter draws the visual representation of a tab bar.
// Each design system (Material 3, Fluent, Cupertino) provides its own
// Painter implementation to render the tab bar in its visual style.
//
// If no Painter is set, the tabview uses [DefaultPainter].
type Painter interface {
	PaintTabBar(canvas widget.Canvas, state PaintState)
}

// PaintState provides the current tab bar state to the painter.
type PaintState struct {
	// Bounds is the tab bar area.
	Bounds geometry.Rect

	// Tabs contains the state of each tab.
	Tabs []TabState

	// SelectedIdx is the index of the currently selected tab.
	SelectedIdx int

	// Position is the tab bar position (top or bottom).
	Position TabPosition

	// Focused is true when the tab bar has keyboard focus.
	Focused bool

	// ColorScheme provides theme-derived colors for tab bar painting.
	// Zero value means the painter should use its built-in defaults.
	ColorScheme TabColorScheme
}

// TabState provides the state of a single tab to the painter.
type TabState struct {
	// Label is the display text.
	Label string

	// Bounds is the tab's clickable area within the tab bar.
	Bounds geometry.Rect

	// Selected is true when this tab is the active tab.
	Selected bool

	// Hovered is true when the mouse is over this tab.
	Hovered bool

	// Disabled is true when this tab cannot be selected.
	Disabled bool

	// Closeable is true when this tab shows a close button.
	Closeable bool

	// CloseButtonBounds is the clickable area for the close button.
	// Empty if not closeable.
	CloseButtonBounds geometry.Rect
}

// TabColorScheme provides theme-derived colors for tab bar painting.
// Zero value means the painter should use its built-in defaults.
type TabColorScheme struct {
	Background      widget.Color
	SelectedText    widget.Color
	UnselectedText  widget.Color
	Indicator       widget.Color
	HoverBackground widget.Color
	CloseButton     widget.Color
	FocusRing       widget.Color
}

// DefaultPainter provides a minimal fallback painter with no design system styling.
// It draws a simple tab bar useful for testing and prototyping.
type DefaultPainter struct{}

// PaintTabBar renders a minimal tab bar with basic styling.
func (p DefaultPainter) PaintTabBar(canvas widget.Canvas, ps PaintState) {
	if ps.Bounds.IsEmpty() {
		return
	}

	hasScheme := ps.ColorScheme != (TabColorScheme{})

	// Tab bar background.
	bg := defaultTabBarBg
	if hasScheme {
		bg = ps.ColorScheme.Background
	}
	canvas.DrawRect(ps.Bounds, bg)

	// Draw each tab.
	for i := range ps.Tabs {
		ts := &ps.Tabs[i]
		paintDefaultTab(canvas, ts, hasScheme, ps.ColorScheme)
	}

	// Draw indicator under selected tab.
	if ps.SelectedIdx >= 0 && ps.SelectedIdx < len(ps.Tabs) {
		selected := &ps.Tabs[ps.SelectedIdx]
		indicatorColor := defaultIndicatorColor
		if hasScheme {
			indicatorColor = ps.ColorScheme.Indicator
		}
		tabW := selected.Bounds.Width()
		indicatorBounds := geometry.NewRect(
			selected.Bounds.Min.X,
			selected.Bounds.Max.Y-indicatorHeight,
			tabW,
			indicatorHeight,
		)
		if ps.Position == Bottom {
			indicatorBounds = geometry.NewRect(
				selected.Bounds.Min.X,
				selected.Bounds.Min.Y,
				tabW,
				indicatorHeight,
			)
		}
		canvas.DrawRect(indicatorBounds, indicatorColor)
	}

	// Focus ring.
	if ps.Focused {
		ringColor := defaultFocusRingColor
		if hasScheme {
			ringColor = ps.ColorScheme.FocusRing
		}
		canvas.StrokeRect(ps.Bounds, ringColor, defaultFocusRingStroke)
	}
}

// paintDefaultTab draws a single tab with the default painter.
func paintDefaultTab(canvas widget.Canvas, ts *TabState, hasScheme bool, cs TabColorScheme) {
	if ts.Bounds.IsEmpty() {
		return
	}

	// Hover background.
	if ts.Hovered && !ts.Disabled {
		hoverBg := defaultHoverBg
		if hasScheme {
			hoverBg = cs.HoverBackground
		}
		canvas.DrawRect(ts.Bounds, hoverBg)
	}

	// Tab text.
	fg := defaultUnselectedText
	if ts.Selected {
		fg = defaultSelectedText
	}
	if ts.Disabled {
		fg = defaultDisabledText
	}
	if hasScheme {
		fg = cs.UnselectedText
		if ts.Selected {
			fg = cs.SelectedText
		}
	}

	textBounds := ts.Bounds
	if ts.Closeable {
		// Shrink text area to leave room for close button.
		textBounds = geometry.NewRect(
			ts.Bounds.Min.X,
			ts.Bounds.Min.Y,
			ts.Bounds.Width()-closeButtonSize-closeButtonPadding,
			ts.Bounds.Height(),
		)
	}

	canvas.DrawText(ts.Label, textBounds, tabFontSize, fg, ts.Selected, defaultTextAlign)

	// Close button (simple "x").
	if ts.Closeable && !ts.Disabled {
		closeFg := defaultCloseButtonColor
		if hasScheme {
			closeFg = cs.CloseButton
		}
		cb := ts.CloseButtonBounds
		if !cb.IsEmpty() {
			centerX := (cb.Min.X + cb.Max.X) / 2
			centerY := (cb.Min.Y + cb.Max.Y) / 2
			halfSize := closeButtonSize / 4
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
	defaultTabBarBg         = widget.RGBA(0.95, 0.95, 0.95, 1.0)
	defaultSelectedText     = widget.RGBA(0.1, 0.1, 0.1, 1.0)
	defaultUnselectedText   = widget.RGBA(0.45, 0.45, 0.45, 1.0)
	defaultIndicatorColor   = widget.Hex(0x6750A4)
	defaultHoverBg          = widget.RGBA(0.0, 0.0, 0.0, 0.04)
	defaultCloseButtonColor = widget.RGBA(0.45, 0.45, 0.45, 1.0)
	defaultDisabledText     = widget.RGBA(0.7, 0.7, 0.7, 1.0)
	defaultFocusRingColor   = widget.Hex(0x6750A4).WithAlpha(0.7)
)

// Painting constants.
const (
	defaultTextAlign               = widget.TextAlignCenter
	defaultFocusRingStroke float32 = 2
)
