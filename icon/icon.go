package icon

// Command identifies a path drawing operation within an icon definition.
type Command uint8

const (
	// CmdMoveTo moves the current point without drawing. Params: [x, y].
	CmdMoveTo Command = iota

	// CmdLineTo draws a straight line from the current point. Params: [x, y].
	CmdLineTo

	// CmdCubicTo draws a cubic Bezier curve. Params: [cx1, cy1, cx2, cy2, x, y].
	CmdCubicTo

	// CmdClose closes the current sub-path by drawing a line back to the
	// starting point of the sub-path. Params: unused.
	CmdClose
)

// commandNames maps each Command to its human-readable name.
var commandNames = [...]string{
	CmdMoveTo:  "MoveTo",
	CmdLineTo:  "LineTo",
	CmdCubicTo: "CubicTo",
	CmdClose:   "Close",
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
}

// defaultViewBox is the standard Material Design icon viewbox size.
const defaultViewBox float32 = 24

// --- Built-in icons ---

// Close is an X mark icon (two diagonal lines).
var Close = IconData{
	Name:    "close",
	ViewBox: defaultViewBox,
	Ops: []PathOp{
		Move(6, 6), Line(18, 18),
		Move(18, 6), Line(6, 18),
	},
}

// Check is a checkmark icon.
var Check = IconData{
	Name:    "check",
	ViewBox: defaultViewBox,
	Ops: []PathOp{
		Move(5, 12), Line(10, 17), Line(19, 7),
	},
}

// ChevronDown is a downward-pointing chevron.
var ChevronDown = IconData{
	Name:    "chevron_down",
	ViewBox: defaultViewBox,
	Ops: []PathOp{
		Move(7, 9), Line(12, 14), Line(17, 9),
	},
}

// ChevronRight is a right-pointing chevron.
var ChevronRight = IconData{
	Name:    "chevron_right",
	ViewBox: defaultViewBox,
	Ops: []PathOp{
		Move(9, 7), Line(14, 12), Line(9, 17),
	},
}

// Search is a magnifying glass icon.
var Search = IconData{
	Name:    "search",
	ViewBox: defaultViewBox,
	Ops: []PathOp{
		// Circle approximated with lines (octagon-like)
		Move(15.5, 10),
		Line(14.9, 12.4), Line(12.9, 14.2), Line(10.5, 14.7),
		Line(8.1, 14.2), Line(6.1, 12.4), Line(5.5, 10),
		Line(6.1, 7.6), Line(8.1, 5.8), Line(10.5, 5.3),
		Line(12.9, 5.8), Line(14.9, 7.6), Line(15.5, 10),
		// Handle
		Move(14, 14), Line(19, 19),
	},
}

// Settings is a simplified gear icon.
var Settings = IconData{
	Name:    "settings",
	ViewBox: defaultViewBox,
	Ops: []PathOp{
		// Outer gear shape approximated with lines
		Move(12, 3), Line(14, 5), Line(16, 5), Line(19, 8),
		Line(19, 10), Line(21, 12), Line(19, 14), Line(19, 16),
		Line(16, 19), Line(14, 19), Line(12, 21), Line(10, 19),
		Line(8, 19), Line(5, 16), Line(5, 14), Line(3, 12),
		Line(5, 10), Line(5, 8), Line(8, 5), Line(10, 5),
		ClosePath(),
		// Inner circle (dot in the center)
		Move(14, 12), Line(13.4, 13.4), Line(12, 14),
		Line(10.6, 13.4), Line(10, 12), Line(10.6, 10.6),
		Line(12, 10), Line(13.4, 10.6), Line(14, 12),
	},
}

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
