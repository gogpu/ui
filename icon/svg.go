package icon

import (
	"github.com/gogpu/gg"
)

// FromSVG creates an IconData from an SVG path data string.
//
// The viewBox specifies the SVG coordinate space side length (e.g., 16 for
// a 16x16 viewBox, 24 for Material Design's 24x24). SVG path commands
// are parsed using [gg.ParseSVGPath] and converted to [PathOp] operations.
//
// FromSVG panics if the SVG path data is invalid. This is intentional because
// icon definitions are typically package-level variables initialized at startup;
// a malformed path indicates a programming error that should fail fast.
//
// Example:
//
//	var MyIcon = icon.FromSVG("my_icon", 16, "M8 2L14 8L8 14L2 8Z")
func FromSVG(name string, viewBox float32, svgPathData string) IconData {
	ops, err := parseSVGToOps(svgPathData)
	if err != nil {
		panic("icon.FromSVG: " + name + ": " + err.Error())
	}
	return IconData{
		Name:        name,
		ViewBox:     viewBox,
		Ops:         ops,
		SVGData:     svgPathData, // fill rendering for complex shapes
		StrokeWidth: 0.8,         // fallback stroke
	}
}

// FromSVGStroke creates an IconData for stroke-based SVG icons.
// These are rendered via stroke (DrawLine) not fill, matching expui outline style.
func FromSVGStroke(name string, viewBox float32, svgPathData string) IconData {
	ops, err := parseSVGToOps(svgPathData)
	if err != nil {
		panic("icon.FromSVGStroke: " + name + ": " + err.Error())
	}
	return IconData{
		Name:        name,
		ViewBox:     viewBox,
		Ops:         ops,
		StrokeWidth: 0.8, // expui stroke icons
	}
}

// FromSVGXML creates an IconData from full SVG XML data.
// The SVG is rendered via gg/svg.RenderWithColor into a bitmap at the
// target size, matching JetBrains icon rendering pipeline.
// This handles all SVG elements: path, circle, rect, with proper
// fill, stroke, fill-rule, stroke-linecap, etc.
func FromSVGXML(name string, svgXML []byte) IconData {
	return IconData{
		Name:    name,
		ViewBox: 16, // default; actual viewBox from SVG XML
		SVGXML:  svgXML,
	}
}

// MustFromSVG is an alias for [FromSVG] provided for clarity. It creates
// an IconData from SVG path data, panicking on invalid input.
//
// Deprecated: Use [FromSVG] directly; it already panics on error.
var MustFromSVG = FromSVG

// TryFromSVG creates an IconData from SVG path data, returning an error
// instead of panicking if the data is invalid. This is useful for
// user-supplied or runtime-loaded icon data.
func TryFromSVG(name string, viewBox float32, svgPathData string) (IconData, error) {
	ops, err := parseSVGToOps(svgPathData)
	if err != nil {
		return IconData{}, err
	}
	return IconData{
		Name:    name,
		ViewBox: viewBox,
		Ops:     ops,
	}, nil
}

// parseSVGToOps parses an SVG path data string and converts the resulting
// gg.Path elements to PathOp operations.
func parseSVGToOps(svgPathData string) ([]PathOp, error) {
	path, err := gg.ParseSVGPath(svgPathData)
	if err != nil {
		return nil, err
	}
	return convertPathToOps(path), nil
}

// convertPathToOps converts a gg.Path to PathOp operations using zero-alloc iteration.
func convertPathToOps(path *gg.Path) []PathOp {
	ops := make([]PathOp, 0, path.NumVerbs())

	path.Iterate(func(verb gg.PathVerb, coords []float64) {
		switch verb {
		case gg.MoveTo:
			ops = append(ops, Move(float32(coords[0]), float32(coords[1])))
		case gg.LineTo:
			ops = append(ops, Line(float32(coords[0]), float32(coords[1])))
		case gg.CubicTo:
			ops = append(ops, Cubic(
				float32(coords[0]), float32(coords[1]),
				float32(coords[2]), float32(coords[3]),
				float32(coords[4]), float32(coords[5]),
			))
		case gg.QuadTo:
			ops = append(ops, Quad(
				float32(coords[0]), float32(coords[1]),
				float32(coords[2]), float32(coords[3]),
			))
		case gg.Close:
			ops = append(ops, ClosePath())
		}
	})

	return ops
}
