// Package icon provides a vector path icon system for the gogpu/ui toolkit.
//
// Icons are defined as sequences of path commands (MoveTo, LineTo, CubicTo,
// Close) within a square viewbox (typically 24x24). The system scales icons
// to fit any display size while maintaining crisp stroked lines.
//
// # Built-in Icons
//
// The package includes common Material-style icons as package-level variables:
//
//   - [Close] (X mark)
//   - [Check] (checkmark)
//   - [ChevronDown], [ChevronRight] (directional arrows)
//   - [Search] (magnifying glass)
//   - [Settings] (gear)
//   - [Menu] (hamburger, 3 lines)
//   - [ArrowBack] (left arrow)
//   - [Add] (plus)
//   - [Delete] (trash can)
//
// # Custom Icons
//
// Define custom icons by constructing [IconData] with a slice of [PathOp]:
//
//	star := icon.IconData{
//	    Name:    "star",
//	    ViewBox: 24,
//	    Ops: []icon.PathOp{
//	        icon.Move(12, 2),
//	        icon.Line(15, 9),
//	        icon.Line(22, 9),
//	        // ... more ops
//	        icon.ClosePath(),
//	    },
//	}
//
// # Widget
//
// Use [NewIcon] to create a display widget:
//
//	w := icon.NewIcon(icon.Check, icon.Size(32), icon.Color(widget.ColorGreen))
//
// # Direct Drawing
//
// Use [Draw] to render an icon directly to a canvas:
//
//	icon.Draw(canvas, icon.Settings, bounds, widget.ColorBlack)
package icon
