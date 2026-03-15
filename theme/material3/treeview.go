package material3

import (
	"github.com/gogpu/ui/core/treeview"
	"github.com/gogpu/ui/geometry"
	"github.com/gogpu/ui/widget"
)

// TreeViewPainter renders tree views using Material 3 design tokens.
// It maps M3 color roles to tree elements: primary for selection,
// surface for background, outline for connector lines, and on-surface for text.
//
// If Theme is nil, TreeViewPainter falls back to the default M3 purple palette.
type TreeViewPainter struct {
	Theme *Theme // nil uses default M3 purple fallback
}

// resolveColors returns M3-derived colors for tree painting.
func (p TreeViewPainter) resolveColors() treeview.TreeColorScheme {
	if p.Theme == nil {
		return m3DefaultTreeColors
	}
	cs := p.Theme.Colors
	return treeview.TreeColorScheme{
		SelectionColor: cs.Primary.WithAlpha(0.12),
		HoverColor:     cs.OnSurface.WithAlpha(0.04),
		FocusColor:     cs.Primary.WithAlpha(0.7),
		LabelColor:     cs.OnSurface,
		LineColor:      cs.Outline,
		IconColor:      cs.OnSurfaceVariant,
		EmptyTextColor: cs.OnSurfaceVariant,
	}
}

// PaintRowBackground draws the hover highlight for a tree row using M3 colors.
func (p TreeViewPainter) PaintRowBackground(canvas widget.Canvas, s treeview.RowPaintState) {
	if s.Bounds.IsEmpty() {
		return
	}
	if s.Hovered && !s.Disabled {
		colors := p.effectiveColors(s.ColorScheme)
		canvas.DrawRoundRect(s.Bounds, colors.HoverColor, m3TreeRowRadius)
	}
}

// PaintSelection draws the selection highlight using M3 primary color.
func (p TreeViewPainter) PaintSelection(canvas widget.Canvas, s treeview.RowPaintState) {
	if s.Bounds.IsEmpty() || !s.Selected {
		return
	}
	colors := p.effectiveColors(s.ColorScheme)
	canvas.DrawRoundRect(s.Bounds, colors.SelectionColor, m3TreeRowRadius)

	if s.Focused && !s.Disabled {
		canvas.StrokeRoundRect(s.Bounds, colors.FocusColor, m3TreeRowRadius, m3TreeFocusBorderWidth)
	}
}

// PaintExpandIcon draws the expand/collapse indicator using M3 on-surface-variant.
func (p TreeViewPainter) PaintExpandIcon(canvas widget.Canvas, s treeview.ExpandIconState) {
	if s.Bounds.IsEmpty() {
		return
	}
	colors := p.effectiveExpandColors(s.ColorScheme)

	cx := s.Bounds.Min.X + s.Bounds.Width()/2
	cy := s.Bounds.Min.Y + s.Bounds.Height()/2

	if s.Expanded {
		// Down-pointing chevron.
		canvas.DrawLine(
			geometry.Pt(cx-m3TreeIconHalfSize, cy-m3TreeIconQuarterSize),
			geometry.Pt(cx, cy+m3TreeIconQuarterSize),
			colors.IconColor, m3TreeIconStrokeWidth,
		)
		canvas.DrawLine(
			geometry.Pt(cx, cy+m3TreeIconQuarterSize),
			geometry.Pt(cx+m3TreeIconHalfSize, cy-m3TreeIconQuarterSize),
			colors.IconColor, m3TreeIconStrokeWidth,
		)
	} else {
		// Right-pointing chevron.
		canvas.DrawLine(
			geometry.Pt(cx-m3TreeIconQuarterSize, cy-m3TreeIconHalfSize),
			geometry.Pt(cx+m3TreeIconQuarterSize, cy),
			colors.IconColor, m3TreeIconStrokeWidth,
		)
		canvas.DrawLine(
			geometry.Pt(cx+m3TreeIconQuarterSize, cy),
			geometry.Pt(cx-m3TreeIconQuarterSize, cy+m3TreeIconHalfSize),
			colors.IconColor, m3TreeIconStrokeWidth,
		)
	}
}

// PaintConnectorLines draws L-shaped connector lines using M3 outline color.
func (p TreeViewPainter) PaintConnectorLines(canvas widget.Canvas, s treeview.ConnectorState) {
	if s.RowBounds.IsEmpty() || s.Depth == 0 {
		return
	}
	colors := p.effectiveConnectorColors(s.ColorScheme)
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
			colors.LineColor, m3TreeConnectorWidth,
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
		colors.LineColor, m3TreeConnectorWidth,
	)
	canvas.DrawLine(
		geometry.Pt(x, midY),
		geometry.Pt(hEnd, midY),
		colors.LineColor, m3TreeConnectorWidth,
	)
}

// PaintLabel draws the node label text using M3 on-surface color.
func (p TreeViewPainter) PaintLabel(canvas widget.Canvas, s treeview.LabelState) {
	if s.Bounds.IsEmpty() || s.Text == "" {
		return
	}
	colors := p.effectiveLabelColors(s.ColorScheme)
	color := colors.LabelColor
	if s.Disabled {
		color = color.WithAlpha(m3TreeDisabledAlpha)
	}
	canvas.DrawText(s.Text, s.Bounds, m3TreeLabelFontSize, color, false, widget.TextAlignLeft)
}

// PaintEmptyState draws a centered placeholder message using M3 on-surface-variant.
func (p TreeViewPainter) PaintEmptyState(canvas widget.Canvas, bounds geometry.Rect) {
	if bounds.IsEmpty() {
		return
	}
	colors := p.resolveColors()
	canvas.DrawText(m3TreeEmptyText, bounds, m3TreeEmptyFontSize, colors.EmptyTextColor, false, widget.TextAlignCenter)
}

// effectiveColors returns the ColorScheme from the paint state, falling back to resolved M3 colors.
func (p TreeViewPainter) effectiveColors(cs treeview.TreeColorScheme) treeview.TreeColorScheme {
	if cs != (treeview.TreeColorScheme{}) {
		return cs
	}
	return p.resolveColors()
}

// effectiveExpandColors returns colors for expand icon painting.
func (p TreeViewPainter) effectiveExpandColors(cs treeview.TreeColorScheme) treeview.TreeColorScheme {
	if cs != (treeview.TreeColorScheme{}) {
		return cs
	}
	return p.resolveColors()
}

// effectiveConnectorColors returns colors for connector line painting.
func (p TreeViewPainter) effectiveConnectorColors(cs treeview.TreeColorScheme) treeview.TreeColorScheme {
	if cs != (treeview.TreeColorScheme{}) {
		return cs
	}
	return p.resolveColors()
}

// effectiveLabelColors returns colors for label painting.
func (p TreeViewPainter) effectiveLabelColors(cs treeview.TreeColorScheme) treeview.TreeColorScheme {
	if cs != (treeview.TreeColorScheme{}) {
		return cs
	}
	return p.resolveColors()
}

// m3DefaultTreeColors holds default M3 purple fallback colors for tree views.
var m3DefaultTreeColors = treeview.TreeColorScheme{
	SelectionColor: widget.Hex(0x6750A4).WithAlpha(0.12),
	HoverColor:     widget.RGBA(0.12, 0.12, 0.13, 0.04),
	FocusColor:     widget.Hex(0x6750A4).WithAlpha(0.7),
	LabelColor:     widget.Hex(0x1C1B1F), // M3 on-surface
	LineColor:      widget.Hex(0x79747E), // M3 outline
	IconColor:      widget.Hex(0x49454F), // M3 on-surface-variant
	EmptyTextColor: widget.Hex(0x49454F), // M3 on-surface-variant
}

// M3 tree view drawing constants.
const (
	m3TreeRowRadius        float32 = 8
	m3TreeFocusBorderWidth float32 = 2
	m3TreeIconHalfSize     float32 = 4
	m3TreeIconQuarterSize  float32 = 3
	m3TreeIconStrokeWidth  float32 = 1.5
	m3TreeConnectorWidth   float32 = 1
	m3TreeLabelFontSize    float32 = 14
	m3TreeEmptyFontSize    float32 = 14
	m3TreeDisabledAlpha    float32 = 0.38
	m3TreeEmptyText                = "No items"
)

// Compile-time check that TreeViewPainter implements Painter.
var _ treeview.Painter = TreeViewPainter{}
