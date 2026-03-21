package devtools

import (
	"github.com/gogpu/ui/core/treeview"
	"github.com/gogpu/ui/geometry"
	"github.com/gogpu/ui/widget"
)

// TreeViewPainter renders tree views using DevTools design tokens.
// DevTools tree views use full-row selection highlighting (Blue2 #2E436E),
// dotted indent guides, and compact 24px row height matching JetBrains IDE
// project tree styling.
//
// If Theme is nil, TreeViewPainter falls back to the default DevTools dark palette.
type TreeViewPainter struct {
	Theme *Theme // nil uses default DevTools dark fallback
}

// resolveColors returns DevTools-derived colors for tree painting.
func (p TreeViewPainter) resolveColors() treeview.TreeColorScheme {
	if p.Theme == nil {
		return dtDefaultTreeColors
	}
	cs := p.Theme.Colors
	return treeview.TreeColorScheme{
		SelectionColor: cs.Selection,
		HoverColor:     cs.ControlHover,
		FocusColor:     cs.BorderFocus,
		LabelColor:     cs.OnSurface,
		LineColor:      cs.BorderStrong,
		IconColor:      cs.OnSurfaceSecondary,
		EmptyTextColor: cs.OnSurfaceSecondary,
	}
}

// PaintRowBackground draws the hover highlight for a tree row using DevTools colors.
func (p TreeViewPainter) PaintRowBackground(canvas widget.Canvas, s treeview.RowPaintState) {
	if s.Bounds.IsEmpty() {
		return
	}
	if s.Hovered && !s.Disabled {
		colors := p.effectiveColors(s.ColorScheme)
		canvas.DrawRect(s.Bounds, colors.HoverColor)
	}
}

// PaintSelection draws the selection highlight using DevTools selection color.
// DevTools uses full-row selection without rounded corners.
func (p TreeViewPainter) PaintSelection(canvas widget.Canvas, s treeview.RowPaintState) {
	if s.Bounds.IsEmpty() || !s.Selected {
		return
	}
	colors := p.effectiveColors(s.ColorScheme)
	canvas.DrawRect(s.Bounds, colors.SelectionColor)

	if s.Focused && !s.Disabled {
		canvas.StrokeRect(s.Bounds, colors.FocusColor, dtTreeFocusBorderWidth)
	}
}

// PaintExpandIcon draws the expand/collapse indicator using DevTools secondary color.
func (p TreeViewPainter) PaintExpandIcon(canvas widget.Canvas, s treeview.ExpandIconState) {
	if s.Bounds.IsEmpty() {
		return
	}
	colors := p.effectiveColors(s.ColorScheme)

	cx := s.Bounds.Min.X + s.Bounds.Width()/2
	cy := s.Bounds.Min.Y + s.Bounds.Height()/2

	if s.Expanded {
		// Down-pointing chevron.
		canvas.DrawLine(
			geometry.Pt(cx-dtTreeIconHalfSize, cy-dtTreeIconQuarterSize),
			geometry.Pt(cx, cy+dtTreeIconQuarterSize),
			colors.IconColor, dtTreeIconStrokeWidth,
		)
		canvas.DrawLine(
			geometry.Pt(cx, cy+dtTreeIconQuarterSize),
			geometry.Pt(cx+dtTreeIconHalfSize, cy-dtTreeIconQuarterSize),
			colors.IconColor, dtTreeIconStrokeWidth,
		)
	} else {
		// Right-pointing chevron.
		canvas.DrawLine(
			geometry.Pt(cx-dtTreeIconQuarterSize, cy-dtTreeIconHalfSize),
			geometry.Pt(cx+dtTreeIconQuarterSize, cy),
			colors.IconColor, dtTreeIconStrokeWidth,
		)
		canvas.DrawLine(
			geometry.Pt(cx+dtTreeIconQuarterSize, cy),
			geometry.Pt(cx-dtTreeIconQuarterSize, cy+dtTreeIconHalfSize),
			colors.IconColor, dtTreeIconStrokeWidth,
		)
	}
}

// PaintConnectorLines draws L-shaped connector lines using DevTools border color.
func (p TreeViewPainter) PaintConnectorLines(canvas widget.Canvas, s treeview.ConnectorState) {
	if s.RowBounds.IsEmpty() || s.Depth == 0 {
		return
	}
	colors := p.effectiveColors(s.ColorScheme)
	midY := s.RowBounds.Min.Y + s.RowBounds.Height()/2

	// Vertical continuation lines for ancestors.
	for i := 0; i < len(s.ParentHasMore); i++ {
		if !s.ParentHasMore[i] {
			continue
		}
		depth := i + 1
		x := s.RowBounds.Min.X + float32(depth)*s.IndentWidth + s.IndentWidth/2
		canvas.DrawLine(
			geometry.Pt(x, s.RowBounds.Min.Y),
			geometry.Pt(x, s.RowBounds.Min.Y+s.RowBounds.Height()),
			colors.LineColor, dtTreeConnectorWidth,
		)
	}

	// Horizontal connector from parent's vertical line to this node.
	x := s.RowBounds.Min.X + float32(s.Depth)*s.IndentWidth + s.IndentWidth/2
	hEnd := s.RowBounds.Min.X + float32(s.Depth+1)*s.IndentWidth

	vEnd := midY
	if !s.IsLastChild {
		vEnd = s.RowBounds.Min.Y + s.RowBounds.Height()
	}
	canvas.DrawLine(
		geometry.Pt(x, s.RowBounds.Min.Y),
		geometry.Pt(x, vEnd),
		colors.LineColor, dtTreeConnectorWidth,
	)
	canvas.DrawLine(
		geometry.Pt(x, midY),
		geometry.Pt(hEnd, midY),
		colors.LineColor, dtTreeConnectorWidth,
	)
}

// PaintLabel draws the node label text using DevTools on-surface color.
func (p TreeViewPainter) PaintLabel(canvas widget.Canvas, s treeview.LabelState) {
	if s.Bounds.IsEmpty() || s.Text == "" {
		return
	}
	colors := p.effectiveColors(s.ColorScheme)
	color := colors.LabelColor
	if s.Disabled {
		color = color.WithAlpha(dtTreeDisabledAlpha)
	}
	canvas.DrawText(s.Text, s.Bounds, dtTreeLabelFontSize, color, false, widget.TextAlignLeft)
}

// PaintEmptyState draws a centered placeholder message.
func (p TreeViewPainter) PaintEmptyState(canvas widget.Canvas, bounds geometry.Rect) {
	if bounds.IsEmpty() {
		return
	}
	colors := p.resolveColors()
	canvas.DrawText(dtTreeEmptyText, bounds, dtTreeEmptyFontSize, colors.EmptyTextColor, false, widget.TextAlignCenter)
}

// effectiveColors returns the ColorScheme from the paint state, falling back to resolved DevTools colors.
func (p TreeViewPainter) effectiveColors(cs treeview.TreeColorScheme) treeview.TreeColorScheme {
	if cs != (treeview.TreeColorScheme{}) {
		return cs
	}
	return p.resolveColors()
}

// dtDefaultTreeColors holds the default DevTools dark color scheme for tree views.
var dtDefaultTreeColors = treeview.TreeColorScheme{
	SelectionColor: widget.Hex(0x2E436E), // Blue2 (selection)
	HoverColor:     widget.Hex(0x43454A), // Gray4 (control hover)
	FocusColor:     widget.Hex(0x3574F0), // Blue6 (accent)
	LabelColor:     widget.Hex(0xDFE1E5), // Gray12 (primary text)
	LineColor:      widget.Hex(0x4E5157), // Gray5 (border strong)
	IconColor:      widget.Hex(0x9DA0A8), // Gray9 (secondary text)
	EmptyTextColor: widget.Hex(0x9DA0A8), // Gray9
}

// DevTools tree view drawing constants.
const (
	dtTreeFocusBorderWidth float32 = 1
	dtTreeIconHalfSize     float32 = 4
	dtTreeIconQuarterSize  float32 = 3
	dtTreeIconStrokeWidth  float32 = 1.5
	dtTreeConnectorWidth   float32 = 1
	dtTreeLabelFontSize    float32 = 13
	dtTreeEmptyFontSize    float32 = 13
	dtTreeDisabledAlpha    float32 = 0.38
	dtTreeEmptyText                = "No items"
)

// Compile-time check that TreeViewPainter implements Painter.
var _ treeview.Painter = TreeViewPainter{}
