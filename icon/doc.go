// Package icon provides a vector path icon system for the gogpu/ui toolkit.
//
// Icons are defined as sequences of path commands (MoveTo, LineTo, CubicTo,
// Close) within a square viewbox (typically 24x24). The system scales icons
// to fit any display size while maintaining crisp stroked lines.
//
// # Single-Color Icons
//
// Use [IconData] and [Draw] for icons rendered in a single color:
//
//	icon.Draw(canvas, icon.Check, bounds, widget.ColorBlack)
//
// # Multi-Color Icons
//
// Use [MultiColorIcon] and [DrawMulti] for icons with multiple color groups.
// Each [PathGroup] maps to a color key in a [Palette]:
//
//	palette := icon.DefaultDarkPalette()
//	icon.DrawMulti(canvas, icon.FileGo, bounds, palette)
//
// # Built-in Icons
//
// The package includes common Material-style icons as package-level variables:
//
// Action icons (single-color):
//   - [Close], [Check], [ChevronDown], [ChevronRight] (navigation)
//   - [Search], [Settings], [Menu], [ArrowBack] (chrome)
//   - [Add], [Delete] (editing)
//   - [Play], [Stop], [Pause] (media/execution)
//   - [Debug], [Gear], [Filter] (tools)
//   - [FolderOpen], [FolderClosed], [Terminal] (file system)
//   - [Refresh], [Plus], [Minus] (actions)
//
// File type icons (multi-color):
//   - [FileGo], [FileJSON], [FileYAML], [FileMD] (languages)
//   - [FileTest], [FileConfig], [FileImage], [FileGeneric] (categories)
//
// VCS icons (multi-color):
//   - [GitBranch], [GitCommit], [GitMerge] (operations)
//   - [GitPR], [GitModified] (status)
//
// # SVG Path Data
//
// Use [FromSVG] to create icons from SVG path data strings. This parses
// the d attribute from SVG <path> elements using [gg.ParseSVGPath]:
//
//	var MyIcon = icon.FromSVG("my_icon", 16, "M8 2L14 8L8 14L2 8Z")
//
// For runtime-loaded data that may be invalid, use [TryFromSVG] which
// returns an error instead of panicking.
//
// # Custom Icons
//
// Define custom single-color icons by constructing [IconData]:
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
// Define custom multi-color icons by constructing [MultiColorIcon]:
//
//	myIcon := icon.MultiColorIcon{
//	    Name:    "my_icon",
//	    ViewBox: 24,
//	    Groups: []icon.PathGroup{
//	        {ColorKey: "primary", Ops: []icon.PathOp{icon.Move(0, 0), icon.Line(24, 24)}},
//	        {ColorKey: "accent",  Ops: []icon.PathOp{icon.Move(24, 0), icon.Line(0, 24)}},
//	    },
//	}
//
// # Widget
//
// Use [NewIcon] to create a display widget for single-color icons:
//
//	w := icon.NewIcon(icon.Check, icon.Size(32), icon.Color(widget.ColorGreen))
//
// # Palettes
//
// Use [DefaultDarkPalette] or [DefaultLightPalette] for theme-appropriate
// multi-color icon rendering. Custom palettes can be created as [Palette] maps.
//
// # Registries
//
// Use [DefaultRegistry] for single-color icons and [DefaultMultiColorRegistry]
// for multi-color icons. Both are pre-populated with all built-in icons.
package icon
