package icon

import "github.com/gogpu/ui/widget"

// Command identifies a path drawing operation within an icon definition.
type Command uint8

const (
	// CmdMoveTo moves the current point without drawing. Params: [x, y].
	CmdMoveTo Command = iota

	// CmdLineTo draws a straight line from the current point. Params: [x, y].
	CmdLineTo

	// CmdCubicTo draws a cubic Bezier curve. Params: [cx1, cy1, cx2, cy2, x, y].
	CmdCubicTo

	// CmdQuadraticTo draws a quadratic Bezier curve. Params: [cx, cy, x, y].
	CmdQuadraticTo

	// CmdClose closes the current sub-path by drawing a line back to the
	// starting point of the sub-path. Params: unused.
	CmdClose
)

// commandNames maps each Command to its human-readable name.
var commandNames = [...]string{
	CmdMoveTo:      "MoveTo",
	CmdLineTo:      "LineTo",
	CmdCubicTo:     "CubicTo",
	CmdQuadraticTo: "QuadraticTo",
	CmdClose:       "Close",
}

// unknownStr is the string representation for unknown/unrecognized values.
const unknownStr = "Unknown"

// String returns a human-readable name for the command.
func (c Command) String() string {
	if int(c) < len(commandNames) {
		return commandNames[c]
	}
	return unknownStr
}

// maxParams is the maximum number of float32 parameters per path operation.
const maxParams = 6

// PathOp is a single drawing operation in an icon's path definition.
//
// The number of significant parameters depends on the command:
//   - [CmdMoveTo]: 2 (x, y)
//   - [CmdLineTo]: 2 (x, y)
//   - [CmdCubicTo]: 6 (cx1, cy1, cx2, cy2, x, y)
//   - [CmdQuadraticTo]: 4 (cx, cy, x, y)
//   - [CmdClose]: 0
type PathOp struct {
	Cmd    Command
	Params [maxParams]float32
}

// Move creates a MoveTo path operation.
func Move(x, y float32) PathOp {
	return PathOp{Cmd: CmdMoveTo, Params: [maxParams]float32{x, y}}
}

// Line creates a LineTo path operation.
func Line(x, y float32) PathOp {
	return PathOp{Cmd: CmdLineTo, Params: [maxParams]float32{x, y}}
}

// Cubic creates a CubicTo path operation with control points (cx1, cy1),
// (cx2, cy2) and endpoint (x, y).
func Cubic(cx1, cy1, cx2, cy2, x, y float32) PathOp {
	return PathOp{Cmd: CmdCubicTo, Params: [maxParams]float32{cx1, cy1, cx2, cy2, x, y}}
}

// Quad creates a QuadraticTo path operation with control point (cx, cy)
// and endpoint (x, y).
func Quad(cx, cy, x, y float32) PathOp {
	return PathOp{Cmd: CmdQuadraticTo, Params: [maxParams]float32{cx, cy, x, y}}
}

// ClosePath creates a Close path operation.
func ClosePath() PathOp {
	return PathOp{Cmd: CmdClose}
}

// IconData defines a vector icon as a sequence of path operations within a
// square viewbox.
//
// IconData is a value type: it is safe to copy, compare by name, and store
// in maps. The Ops slice is shared between copies; callers must not mutate
// the slice after constructing the IconData.
type IconData struct {
	// Name is a human-readable identifier for the icon (e.g. "close", "check").
	Name string

	// ViewBox is the side length of the square coordinate space in which the
	// path operations are defined. Typical value is 24 (Material Design).
	ViewBox float32

	// Ops is the ordered sequence of path operations that define the icon shape.
	Ops []PathOp

	// StrokeWidth overrides the default stroke width in viewbox units.
	// Zero uses defaultStrokeWidth (1.5). SVG-sourced icons use thinner
	// strokes (e.g., 0.4) since their paths define filled shape outlines.
	StrokeWidth float32

	// SVGData holds the original SVG path data string for fill rendering.
	// When non-empty and the canvas supports SVGFiller, the icon is rendered
	// as a filled path instead of stroked lines for higher quality.
	SVGData string

	// SVGXML holds full SVG XML data for bitmap rasterization.
	// When non-empty, the icon is rendered via gg/svg.RenderWithColor
	// into a cached bitmap at the target size, matching JetBrains pipeline.
	// This produces pixel-perfect icons with proper fill, stroke, circles,
	// fill-rule, stroke-linecap, etc.
	SVGXML []byte
}

// PathGroup is a named group of path operations that share a color role.
// Multi-color icons contain multiple groups, each rendered with a different
// color from a [Palette].
type PathGroup struct {
	// ColorKey identifies which color from the Palette to use (e.g. "primary", "accent").
	ColorKey string
	// Ops is the sequence of path operations for this group.
	Ops []PathOp
}

// MultiColorIcon defines a vector icon with multiple color groups.
// Each group is rendered with a different color from the provided [Palette].
//
// MultiColorIcon is a value type: it is safe to copy and compare by name.
// The Groups and their Ops slices are shared between copies; callers must
// not mutate them after construction.
type MultiColorIcon struct {
	// Name is a human-readable identifier for the icon (e.g. "file_go", "git_branch").
	Name string
	// ViewBox is the side length of the square coordinate space.
	ViewBox float32
	// Groups is the ordered sequence of path groups, each drawn with its own color.
	Groups []PathGroup
}

// Palette maps color keys to actual colors. Used with [DrawMulti] to
// provide theme-appropriate colors for each [PathGroup] in a [MultiColorIcon].
type Palette map[string]widget.Color

// defaultViewBox is the standard Material Design icon viewbox size.
const defaultViewBox float32 = 24

// --- Built-in icons ---
//
// Icons below use SVG path data parsed via [FromSVG] for smooth curves and
// crisp rendering at all sizes. Path data is derived from JetBrains IntelliJ
// icons (Apache 2.0) adapted to a 16x16 viewBox, or created as clean stroked
// paths in a 24x24 viewBox for simple geometric shapes.

// Close is an X mark icon.
// Based on JetBrains IntelliJ cancel.svg (Apache 2.0, 16x16 viewBox).
var Close = FromSVG("close", 16,
	"M8 9.41421L4.20711 13.2071L2.79289 11.7929L6.58579 8L2.79289 4.20711L4.20711 2.79289L8 6.58579L11.7929 2.79289L13.2071 4.20711L9.41421 8L13.2071 11.7929L11.7929 13.2071L8 9.41421Z")

// Check is a checkmark icon.
// Based on JetBrains IntelliJ checked.svg (Apache 2.0, 12x12 viewBox).
var Check = FromSVG("check", 12,
	"M5.27908728,11 L4.8977527,11 L1.01339746,4.39807621 L2.48564065,3.54807621 L5.2408588,8.32025404 L9.51339746,0.92 L10.9856406,1.77 L5.65058539,11 L5.27908728,11 Z")

// ChevronDown is a downward-pointing chevron.
// Based on JetBrains IntelliJ chevron-down.svg (Apache 2.0, 16x16 viewBox).
var ChevronDown = FromSVG("chevron_down", 16,
	"M8.00004 11.91L2.29004 6.20998L3.71004 4.78998L8.00004 9.08998L12.29 4.78998L13.71 6.20998L8.00004 11.91Z")

// ChevronRight is a right-pointing chevron.
// Based on JetBrains IntelliJ chevron-right.svg (Apache 2.0, 16x16 viewBox).
var ChevronRight = FromSVG("chevron_right", 16,
	"M6.21004 13.71L4.79004 12.29L9.09004 7.99998L4.79004 3.70998L6.21004 2.28998L11.91 7.99998L6.21004 13.71Z")

// Search is a magnifying glass icon.
// Based on JetBrains IntelliJ search.svg (Apache 2.0, 16x16 viewBox).
var Search = FromSVG("search", 16,
	"M11.038136,9.94904865 L13.9980971,12.9090097 L12.9374369,13.9696699 L9.98176525,11.0139983 "+
		"C9.14925083,11.6334368 8.11743313,12 7,12 C4.23857625,12 2,9.76142375 2,7 "+
		"C2,4.23857625 4.23857625,2 7,2 C9.76142375,2 12,4.23857625 12,7 "+
		"C12,8.1028408 11.642948,9.12228765 11.038136,9.94904865 Z "+
		"M7,11 C9.209139,11 11,9.209139 11,7 C11,4.790861 9.209139,3 7,3 "+
		"C4.790861,3 3,4.790861 3,7 C3,9.209139 4.790861,11 7,11 Z")

// Settings is a gear icon for settings.
// Based on JetBrains IntelliJ gearPlain.svg (Apache 2.0, 16x16 viewBox).
var Settings = FromSVG("settings", 16,
	"M12.7078144,8.94092644 L14.1860171,10.0014962 "+
		"C13.9015285,10.8814083 13.4345167,11.6792367 12.8296171,12.3503459 "+
		"L11.1720587,11.6025852 C10.7002906,12.0182974 10.1462196,12.3427961 9.53742767,12.5484995 "+
		"L9.35682478,14.3581758 C8.91920787,14.4511061 8.46531382,14.5 8,14.5 "+
		"C7.53468618,14.5 7.08079213,14.4511061 6.64317522,14.3581758 "+
		"L6.46257231,12.5484994 C5.85378047,12.3427959 5.29970962,12.0182972 4.82794166,11.6025851 "+
		"L3.1703829,12.3503459 C2.5654833,11.6792367 2.0984715,10.8814083 1.8139829,10.0014962 "+
		"L3.29218604,8.94092614 C3.23171292,8.63665953 3.20000005,8.32203335 3.20000005,8.00000024 "+
		"C3.20000005,7.67796698 3.23171295,7.36334066 3.29218612,7.05907392 "+
		"L1.8139829,5.99850385 C2.0984715,5.11859166 2.5654833,4.32076327 3.1703829,3.64965409 "+
		"L4.82794202,4.39741509 C5.29970989,3.98170311 5.85378059,3.65720451 6.46257226,3.45150114 "+
		"L6.64317522,1.64182422 C7.08079213,1.54889386 7.53468618,1.5 8,1.5 "+
		"C8.46531382,1.5 8.91920787,1.54889386 9.35682478,1.64182422 "+
		"L9.53742772,3.45150097 C10.1462195,3.65720431 10.7002903,3.98170292 11.1720583,4.39741495 "+
		"L12.8296171,3.64965409 C13.4345167,4.32076327 13.9015285,5.11859166 14.1860171,5.99850385 "+
		"L12.7078143,7.05907362 C12.7682875,7.36334046 12.8000004,7.67796687 12.8000004,8.00000024 "+
		"C12.8000004,8.32203346 12.7682875,8.63665973 12.7078144,8.94092644 Z "+
		"M7.99999976,10.3003956 C9.27025466,10.3003956 10.2999997,9.27056196 10.2999997,8.00019773 "+
		"C10.2999997,6.72983349 9.27025466,5.69999981 7.99999976,5.69999981 "+
		"C6.72974486,5.69999981 5.69999981,6.72983349 5.69999981,8.00019773 "+
		"C5.69999981,9.27056196 6.72974486,10.3003956 7.99999976,10.3003956 Z")

// Menu is a hamburger menu icon (three horizontal lines).
var Menu = IconData{
	Name:    "menu",
	ViewBox: defaultViewBox,
	Ops: []PathOp{
		Move(4, 7), Line(20, 7),
		Move(4, 12), Line(20, 12),
		Move(4, 17), Line(20, 17),
	},
}

// ArrowBack is a left-pointing arrow.
var ArrowBack = IconData{
	Name:    "arrow_back",
	ViewBox: defaultViewBox,
	Ops: []PathOp{
		Move(20, 12), Line(4, 12),
		Move(4, 12), Line(10, 6),
		Move(4, 12), Line(10, 18),
	},
}

// Add is a plus icon.
var Add = IconData{
	Name:    "add",
	ViewBox: defaultViewBox,
	Ops: []PathOp{
		Move(12, 5), Line(12, 19),
		Move(5, 12), Line(19, 12),
	},
}

// Delete is a simplified trash can icon.
var Delete = IconData{
	Name:    "delete",
	ViewBox: defaultViewBox,
	Ops: []PathOp{
		// Lid
		Move(5, 7), Line(19, 7),
		Move(9, 7), Line(9, 5), Line(15, 5), Line(15, 7),
		// Body
		Move(7, 7), Line(8, 19), Line(16, 19), Line(17, 7),
		// Inner lines
		Move(10, 9), Line(10, 17),
		Move(12, 9), Line(12, 17),
		Move(14, 9), Line(14, 17),
	},
}

// --- DevTools action icons ---

// Play is a right-pointing triangle (run/execute).
var Play = IconData{
	Name:    "play",
	ViewBox: defaultViewBox,
	Ops: []PathOp{
		Move(7, 5), Line(19, 12), Line(7, 19), ClosePath(),
	},
}

// Stop is a square (stop execution).
var Stop = IconData{
	Name:    "stop",
	ViewBox: defaultViewBox,
	Ops: []PathOp{
		Move(6, 6), Line(18, 6), Line(18, 18), Line(6, 18), ClosePath(),
	},
}

// Pause is two vertical bars (pause execution).
var Pause = IconData{
	Name:    "pause",
	ViewBox: defaultViewBox,
	Ops: []PathOp{
		Move(7, 5), Line(7, 19),
		Move(17, 5), Line(17, 19),
	},
}

// Debug is a simplified bug icon.
var Debug = IconData{
	Name:    "debug",
	ViewBox: defaultViewBox,
	Ops: []PathOp{
		// Body oval
		Move(9, 8), Line(9, 16), Line(12, 18), Line(15, 16), Line(15, 8),
		Line(12, 6), ClosePath(),
		// Head
		Move(10, 6), Line(12, 4), Line(14, 6),
		// Left legs
		Move(9, 10), Line(5, 8),
		Move(9, 14), Line(5, 16),
		// Right legs
		Move(15, 10), Line(19, 8),
		Move(15, 14), Line(19, 16),
		// Antennae
		Move(10, 5), Line(8, 3),
		Move(14, 5), Line(16, 3),
	},
}

// Gear is a gear/cog icon for settings.
// Based on JetBrains IntelliJ gearPlain.svg (Apache 2.0, 16x16 viewBox).
// This is an alias-style variant distinct from [Settings] with a different visual weight.
var Gear = FromSVG("gear", 16,
	"M12.7078144,8.94092644 L14.1860171,10.0014962 "+
		"C13.9015285,10.8814083 13.4345167,11.6792367 12.8296171,12.3503459 "+
		"L11.1720587,11.6025852 C10.7002906,12.0182974 10.1462196,12.3427961 9.53742767,12.5484995 "+
		"L9.35682478,14.3581758 C8.91920787,14.4511061 8.46531382,14.5 8,14.5 "+
		"C7.53468618,14.5 7.08079213,14.4511061 6.64317522,14.3581758 "+
		"L6.46257231,12.5484994 C5.85378047,12.3427959 5.29970962,12.0182972 4.82794166,11.6025851 "+
		"L3.1703829,12.3503459 C2.5654833,11.6792367 2.0984715,10.8814083 1.8139829,10.0014962 "+
		"L3.29218604,8.94092614 C3.23171292,8.63665953 3.20000005,8.32203335 3.20000005,8.00000024 "+
		"C3.20000005,7.67796698 3.23171295,7.36334066 3.29218612,7.05907392 "+
		"L1.8139829,5.99850385 C2.0984715,5.11859166 2.5654833,4.32076327 3.1703829,3.64965409 "+
		"L4.82794202,4.39741509 C5.29970989,3.98170311 5.85378059,3.65720451 6.46257226,3.45150114 "+
		"L6.64317522,1.64182422 C7.08079213,1.54889386 7.53468618,1.5 8,1.5 "+
		"C8.46531382,1.5 8.91920787,1.54889386 9.35682478,1.64182422 "+
		"L9.53742772,3.45150097 C10.1462195,3.65720431 10.7002903,3.98170292 11.1720583,4.39741495 "+
		"L12.8296171,3.64965409 C13.4345167,4.32076327 13.9015285,5.11859166 14.1860171,5.99850385 "+
		"L12.7078143,7.05907362 C12.7682875,7.36334046 12.8000004,7.67796687 12.8000004,8.00000024 "+
		"C12.8000004,8.32203346 12.7682875,8.63665973 12.7078144,8.94092644 Z "+
		"M7.99999976,10.3003956 C9.27025466,10.3003956 10.2999997,9.27056196 10.2999997,8.00019773 "+
		"C10.2999997,6.72983349 9.27025466,5.69999981 7.99999976,5.69999981 "+
		"C6.72974486,5.69999981 5.69999981,6.72983349 5.69999981,8.00019773 "+
		"C5.69999981,9.27056196 6.72974486,10.3003956 7.99999976,10.3003956 Z")

// Filter is a funnel icon.
var Filter = IconData{
	Name:    "filter",
	ViewBox: defaultViewBox,
	Ops: []PathOp{
		Move(4, 5), Line(20, 5), Line(14, 12), Line(14, 19), Line(10, 17),
		Line(10, 12), ClosePath(),
	},
}

// FolderOpen is an open folder outline.
// Based on JetBrains IntelliJ menu-open.svg (Apache 2.0, 16x16 viewBox).
var FolderOpen = FromSVG("folder_open", 16,
	"M4.32342122,7 L2,11.0147552 L2,7 L2,5 L2,3 L6.60006714,3 L7.75640322,5 L14,5 L14,7 L4.32342122,7 Z "+
		"M4.89129639,8 L16,8 L13.1082845,13 L2.00248718,13 L4.89129639,8 Z")

// FolderClosed is a closed folder outline.
// Based on JetBrains IntelliJ folder.svg (Apache 2.0, 16x16 viewBox).
var FolderClosed = FromSVG("folder_closed", 16,
	"M1,13 L15,13 L15,4 L7.98457,4 L6.69633,2.71149 "+
		"C6.22161957,2.28559443 5.61570121,2.03457993 4.97888,2 L1.05128,2 "+
		"C1.02295884,2 1,2.02295884 1,2.05128 L1,13 Z")

// Terminal is a terminal/console prompt icon (>_).
var Terminal = IconData{
	Name:    "terminal",
	ViewBox: defaultViewBox,
	Ops: []PathOp{
		// Prompt chevron
		Move(5, 8), Line(11, 12), Line(5, 16),
		// Underscore
		Move(13, 16), Line(19, 16),
	},
}

// Refresh is circular arrows icon.
// Based on JetBrains IntelliJ refresh.svg (Apache 2.0, 16x16 viewBox).
var Refresh = FromSVG("refresh", 16,
	"M12.5747152,11.8852806 C11.4741474,13.1817355 9.83247882,14.0044386 7.99865879,14.0044386 "+
		"C5.03907292,14.0044386 2.57997332,11.8615894 2.08820756,9.0427473 "+
		"L3.94774327,9.10768372 C4.43372186,10.8898575 6.06393114,12.2000519 8.00015362,12.2000519 "+
		"C9.30149237,12.2000519 10.4645985,11.6082097 11.2349873,10.6790094 "+
		"L9.05000019,8.71167959 L14.0431479,8.44999981 L14.3048222,13.4430431 "+
		"L12.5747152,11.8852806 Z "+
		"M3.42785637,4.11741586 C4.52839138,2.82452748 6.16775464,2.00443857 7.99865879,2.00443857 "+
		"C10.918604,2.00443857 13.3513802,4.09026967 13.8882946,6.8532307 "+
		"L12.0226389,6.78808057 C11.5024872,5.05935553 9.89838095,3.8000774 8.00015362,3.8000774 "+
		"C6.69867367,3.8000774 5.53545628,4.39204806 4.76506921,5.32142241 "+
		"L6.95482203,7.29304326 L1.96167436,7.55472304 L1.70000005,2.56167973 "+
		"L3.42785637,4.11741586 Z")

// Plus is a plus sign icon. Same shape as [Add] with a different name for
// DevTools contexts where "plus" is more descriptive than "add".
var Plus = IconData{
	Name:    "plus",
	ViewBox: defaultViewBox,
	Ops: []PathOp{
		Move(12, 5), Line(12, 19),
		Move(5, 12), Line(19, 12),
	},
}

// Minus is a minus/dash sign icon.
var Minus = IconData{
	Name:    "minus",
	ViewBox: defaultViewBox,
	Ops: []PathOp{
		Move(5, 12), Line(19, 12),
	},
}
