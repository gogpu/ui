package listview

import (
	"github.com/gogpu/ui/geometry"
	"github.com/gogpu/ui/widget"
)

// Painter draws list-specific visual elements.
// Each design system (Material 3, Fluent, Cupertino) provides its own
// Painter implementation to render the list in its visual style.
//
// If no Painter is set, the list view uses [DefaultPainter].
type Painter interface {
	// PaintDivider draws a divider line between items.
	PaintDivider(canvas widget.Canvas, state DividerState)

	// PaintEmptyState draws the empty state when ItemCount is 0.
	PaintEmptyState(canvas widget.Canvas, bounds geometry.Rect)

	// PaintItemBackground draws the background for a list item.
	// This is called before the item widget's own Draw method.
	PaintItemBackground(canvas widget.Canvas, state ItemPaintState)

	// PaintSelection draws the selection highlight for a selected item.
	// This is called before the item widget's own Draw method.
	PaintSelection(canvas widget.Canvas, state ItemPaintState)
}

// DividerState provides context for divider painting.
type DividerState struct {
	// Bounds is the rectangle for the divider (full width, typically 1px height).
	Bounds geometry.Rect

	// ItemIndex is the index of the item above the divider.
	ItemIndex int

	// ColorScheme provides theme-derived colors.
	ColorScheme ListColorScheme
}

// ItemPaintState provides context for item background and selection painting.
type ItemPaintState struct {
	// Bounds is the item's bounding rectangle.
	Bounds geometry.Rect

	// Index is the item's zero-based index in the data source.
	Index int

	// Selected is true if the item is selected.
	Selected bool

	// Focused is true if the item has keyboard focus within the list.
	Focused bool

	// Hovered is true if the mouse cursor is over this item.
	Hovered bool

	// Disabled is true if the list is disabled.
	Disabled bool

	// ColorScheme provides theme-derived colors.
	ColorScheme ListColorScheme
}

// ListColorScheme provides theme-derived colors for list painting.
// Zero value means the painter should use its built-in defaults.
type ListColorScheme struct {
	DividerColor      widget.Color // divider line color
	SelectionColor    widget.Color // selected item background
	HoverColor        widget.Color // hovered item background
	FocusColor        widget.Color // focused item border/background
	EmptyTextColor    widget.Color // empty state text color
	ItemBackground    widget.Color // alternating item background (even rows)
	ItemBackgroundAlt widget.Color // alternating item background (odd rows)
}

// DefaultPainter provides a minimal fallback painter with no design system styling.
// It draws simple list visuals -- useful for testing and as a base reference.
type DefaultPainter struct{}

// PaintDivider draws a thin horizontal line at the divider bounds.
func (p DefaultPainter) PaintDivider(canvas widget.Canvas, ds DividerState) {
	if ds.Bounds.IsEmpty() {
		return
	}
	color := defaultDividerColor
	hasScheme := ds.ColorScheme != (ListColorScheme{})
	if hasScheme {
		color = ds.ColorScheme.DividerColor
	}
	canvas.DrawRect(ds.Bounds, color)
}

// PaintEmptyState draws a centered "No items" message.
func (p DefaultPainter) PaintEmptyState(canvas widget.Canvas, bounds geometry.Rect) {
	if bounds.IsEmpty() {
		return
	}
	canvas.DrawText(emptyStateText, bounds, emptyStateFontSize, defaultEmptyTextColor, false, emptyStateAlign)
}

// PaintItemBackground draws the background for a list item.
func (p DefaultPainter) PaintItemBackground(canvas widget.Canvas, ips ItemPaintState) {
	if ips.Bounds.IsEmpty() {
		return
	}
	hasScheme := ips.ColorScheme != (ListColorScheme{})
	if ips.Hovered && !ips.Disabled {
		color := defaultHoverColor
		if hasScheme {
			color = ips.ColorScheme.HoverColor
		}
		canvas.DrawRect(ips.Bounds, color)
	}
}

// PaintSelection draws the selection highlight for a selected item.
func (p DefaultPainter) PaintSelection(canvas widget.Canvas, ips ItemPaintState) {
	if ips.Bounds.IsEmpty() || !ips.Selected {
		return
	}
	hasScheme := ips.ColorScheme != (ListColorScheme{})
	color := defaultSelectionColor
	if hasScheme {
		color = ips.ColorScheme.SelectionColor
	}
	canvas.DrawRect(ips.Bounds, color)

	// Draw focus border.
	if ips.Focused && !ips.Disabled {
		focusColor := defaultFocusBorderColor
		if hasScheme {
			focusColor = ips.ColorScheme.FocusColor
		}
		canvas.StrokeRect(ips.Bounds, focusColor, focusBorderWidth)
	}
}

// Painting constants.
const (
	dividerHeight    float32 = 1
	focusBorderWidth float32 = 2

	emptyStateFontSize float32 = 14
	emptyStateAlign            = widget.TextAlignCenter
	emptyStateText             = "No items"
)

// Default colors for DefaultPainter.
var (
	defaultDividerColor     = widget.RGBA(0.85, 0.85, 0.85, 1.0)
	defaultSelectionColor   = widget.RGBA(0.23, 0.51, 0.96, 0.12)
	defaultHoverColor       = widget.RGBA(0.0, 0.0, 0.0, 0.04)
	defaultFocusBorderColor = widget.Hex(0x6750A4).WithAlpha(0.7)
	defaultEmptyTextColor   = widget.RGBA(0.5, 0.5, 0.5, 1.0)
)
