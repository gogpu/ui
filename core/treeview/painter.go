package treeview

import (
	"github.com/gogpu/ui/geometry"
	"github.com/gogpu/ui/widget"
)

// Painter draws tree-specific visual elements.
// Each design system (Material 3, Fluent, Cupertino) provides its own
// Painter implementation to render the tree in its visual style.
//
// If no Painter is set, the tree view uses [DefaultPainter].
type Painter interface {
	// PaintRowBackground draws the background for a tree row.
	// Called before the row label is drawn.
	PaintRowBackground(canvas widget.Canvas, state RowPaintState)

	// PaintSelection draws the selection highlight for a selected row.
	// Called before the row label is drawn.
	PaintSelection(canvas widget.Canvas, state RowPaintState)

	// PaintExpandIcon draws the expand/collapse indicator for branch nodes.
	PaintExpandIcon(canvas widget.Canvas, state ExpandIconState)

	// PaintConnectorLines draws the tree connector lines (if enabled).
	PaintConnectorLines(canvas widget.Canvas, state ConnectorState)

	// PaintLabel draws the node label text.
	PaintLabel(canvas widget.Canvas, state LabelState)

	// PaintEmptyState draws placeholder content when the tree has no root.
	PaintEmptyState(canvas widget.Canvas, bounds geometry.Rect)
}

// RowPaintState provides context for row background and selection painting.
type RowPaintState struct {
	// Bounds is the full row bounding rectangle.
	Bounds geometry.Rect

	// Node is the tree node for this row.
	Node *TreeNode

	// Depth is the nesting level (0 = root).
	Depth int

	// Selected is true if this row is the currently selected node.
	Selected bool

	// Focused is true if the tree has keyboard focus and this row is selected.
	Focused bool

	// Hovered is true if the mouse cursor is over this row.
	Hovered bool

	// Disabled is true if the tree is disabled.
	Disabled bool

	// ColorScheme provides theme-derived colors.
	ColorScheme TreeColorScheme
}

// ExpandIconState provides context for expand/collapse icon painting.
type ExpandIconState struct {
	// Bounds is the bounding rectangle for the expand icon.
	Bounds geometry.Rect

	// Expanded is true if the node is currently expanded.
	Expanded bool

	// Hovered is true if the mouse cursor is over this row.
	Hovered bool

	// ColorScheme provides theme-derived colors.
	ColorScheme TreeColorScheme
}

// ConnectorState provides context for drawing tree connector lines.
type ConnectorState struct {
	// RowBounds is the full row bounding rectangle.
	RowBounds geometry.Rect

	// Depth is the nesting level of this row.
	Depth int

	// IndentWidth is the pixel offset per nesting level.
	IndentWidth float32

	// IsLastChild is true if this node is the last child of its parent.
	IsLastChild bool

	// HasChildren is true if this node has children (branch node).
	HasChildren bool

	// ParentHasMore indicates which ancestor levels still have more
	// siblings below. Index 0 = depth 1, index 1 = depth 2, etc.
	// Used to draw vertical continuation lines.
	ParentHasMore []bool

	// ColorScheme provides theme-derived colors.
	ColorScheme TreeColorScheme
}

// LabelState provides context for drawing the node label.
type LabelState struct {
	// Bounds is the bounding rectangle for the label text.
	Bounds geometry.Rect

	// Text is the label text.
	Text string

	// Selected is true if this row is selected.
	Selected bool

	// Disabled is true if the tree is disabled.
	Disabled bool

	// ColorScheme provides theme-derived colors.
	ColorScheme TreeColorScheme
}

// TreeColorScheme provides theme-derived colors for tree painting.
// Zero value means the painter should use its built-in defaults.
type TreeColorScheme struct {
	SelectionColor widget.Color // selected row background
	HoverColor     widget.Color // hovered row background
	FocusColor     widget.Color // focused row border/background
	LabelColor     widget.Color // default label text color
	LineColor      widget.Color // connector line color
	IconColor      widget.Color // expand/collapse icon color
	EmptyTextColor widget.Color // empty state text color
}

// DefaultPainter provides a minimal fallback painter with no design system styling.
// It draws simple tree visuals -- useful for testing and as a base reference.
type DefaultPainter struct{}

// PaintRowBackground draws the hover highlight for a tree row.
func (p DefaultPainter) PaintRowBackground(canvas widget.Canvas, s RowPaintState) {
	if s.Bounds.IsEmpty() {
		return
	}
	if s.Hovered && !s.Disabled {
		color := defaultHoverColor
		if s.ColorScheme != (TreeColorScheme{}) {
			color = s.ColorScheme.HoverColor
		}
		canvas.DrawRect(s.Bounds, color)
	}
}

// PaintSelection draws the selection highlight for a selected row.
func (p DefaultPainter) PaintSelection(canvas widget.Canvas, s RowPaintState) {
	if s.Bounds.IsEmpty() || !s.Selected {
		return
	}
	color := defaultSelectionColor
	if s.ColorScheme != (TreeColorScheme{}) {
		color = s.ColorScheme.SelectionColor
	}
	canvas.DrawRect(s.Bounds, color)

	if s.Focused && !s.Disabled {
		focusColor := defaultFocusBorderColor
		if s.ColorScheme != (TreeColorScheme{}) {
			focusColor = s.ColorScheme.FocusColor
		}
		canvas.StrokeRect(s.Bounds, focusColor, focusBorderWidth)
	}
}

// PaintExpandIcon draws a simple triangle expand/collapse icon.
func (p DefaultPainter) PaintExpandIcon(canvas widget.Canvas, s ExpandIconState) {
	if s.Bounds.IsEmpty() {
		return
	}
	color := defaultIconColor
	if s.ColorScheme != (TreeColorScheme{}) {
		color = s.ColorScheme.IconColor
	}

	cx := s.Bounds.Min.X + s.Bounds.Width()/2
	cy := s.Bounds.Min.Y + s.Bounds.Height()/2

	if s.Expanded {
		// Down-pointing triangle.
		canvas.DrawLine(
			geometry.Pt(cx-iconHalfSize, cy-iconQuarterSize),
			geometry.Pt(cx, cy+iconQuarterSize),
			color, iconStrokeWidth,
		)
		canvas.DrawLine(
			geometry.Pt(cx, cy+iconQuarterSize),
			geometry.Pt(cx+iconHalfSize, cy-iconQuarterSize),
			color, iconStrokeWidth,
		)
	} else {
		// Right-pointing triangle.
		canvas.DrawLine(
			geometry.Pt(cx-iconQuarterSize, cy-iconHalfSize),
			geometry.Pt(cx+iconQuarterSize, cy),
			color, iconStrokeWidth,
		)
		canvas.DrawLine(
			geometry.Pt(cx+iconQuarterSize, cy),
			geometry.Pt(cx-iconQuarterSize, cy+iconHalfSize),
			color, iconStrokeWidth,
		)
	}
}

// PaintConnectorLines draws L-shaped connector lines for tree hierarchy.
func (p DefaultPainter) PaintConnectorLines(canvas widget.Canvas, s ConnectorState) {
	if s.RowBounds.IsEmpty() || s.Depth == 0 {
		return
	}
	color := defaultLineColor
	if s.ColorScheme != (TreeColorScheme{}) {
		color = s.ColorScheme.LineColor
	}

	midY := s.RowBounds.Min.Y + s.RowBounds.Height()/2

	// Draw vertical continuation lines for ancestors that have more siblings.
	for i := 0; i < len(s.ParentHasMore); i++ {
		if !s.ParentHasMore[i] {
			continue
		}
		depth := i + 1
		x := s.RowBounds.Min.X + float32(depth)*s.IndentWidth + s.IndentWidth/2
		canvas.DrawLine(
			geometry.Pt(x, s.RowBounds.Min.Y),
			geometry.Pt(x, s.RowBounds.Min.Y+s.RowBounds.Height()),
			color, connectorLineWidth,
		)
	}

	// Draw the horizontal connector from parent's vertical line to this node.
	x := s.RowBounds.Min.X + float32(s.Depth)*s.IndentWidth + s.IndentWidth/2
	hEnd := s.RowBounds.Min.X + float32(s.Depth+1)*s.IndentWidth

	// Vertical segment: from top of row to mid-point (or full height if not last).
	vEnd := midY
	if !s.IsLastChild {
		vEnd = s.RowBounds.Min.Y + s.RowBounds.Height()
	}
	canvas.DrawLine(
		geometry.Pt(x, s.RowBounds.Min.Y),
		geometry.Pt(x, vEnd),
		color, connectorLineWidth,
	)

	// Horizontal segment.
	canvas.DrawLine(
		geometry.Pt(x, midY),
		geometry.Pt(hEnd, midY),
		color, connectorLineWidth,
	)
}

// PaintLabel draws the node label text.
func (p DefaultPainter) PaintLabel(canvas widget.Canvas, s LabelState) {
	if s.Bounds.IsEmpty() || s.Text == "" {
		return
	}
	color := defaultLabelColor
	if s.ColorScheme != (TreeColorScheme{}) {
		color = s.ColorScheme.LabelColor
	}
	if s.Disabled {
		color = defaultDisabledLabelColor
	}
	canvas.DrawText(s.Text, s.Bounds, labelFontSize, color, false, widget.TextAlignLeft)
}

// PaintEmptyState draws a centered "No items" message.
func (p DefaultPainter) PaintEmptyState(canvas widget.Canvas, bounds geometry.Rect) {
	if bounds.IsEmpty() {
		return
	}
	canvas.DrawText(emptyStateText, bounds, emptyStateFontSize, defaultEmptyTextColor, false, widget.TextAlignCenter)
}

// Painting constants.
const (
	focusBorderWidth   float32 = 2
	iconHalfSize       float32 = 4
	iconQuarterSize    float32 = 3
	iconStrokeWidth    float32 = 1.5
	connectorLineWidth float32 = 1
	expandIconSize     float32 = 16
	labelFontSize      float32 = 14
	emptyStateFontSize float32 = 14
	emptyStateText             = "No items"
)

// Default colors for DefaultPainter.
var (
	defaultSelectionColor     = widget.RGBA(0.23, 0.51, 0.96, 0.12)
	defaultHoverColor         = widget.RGBA(0.0, 0.0, 0.0, 0.04)
	defaultFocusBorderColor   = widget.Hex(0x6750A4).WithAlpha(0.7)
	defaultIconColor          = widget.RGBA(0.4, 0.4, 0.4, 1.0)
	defaultLineColor          = widget.RGBA(0.75, 0.75, 0.75, 1.0)
	defaultLabelColor         = widget.RGBA(0.1, 0.1, 0.1, 1.0)
	defaultDisabledLabelColor = widget.RGBA(0.5, 0.5, 0.5, 1.0)
	defaultEmptyTextColor     = widget.RGBA(0.5, 0.5, 0.5, 1.0)
)
