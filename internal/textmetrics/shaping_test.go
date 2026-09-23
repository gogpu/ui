package textmetrics

import (
	"testing"

	"github.com/gogpu/gg/scene"
	"github.com/gogpu/gg/text"
	"github.com/gogpu/ui/geometry"
	"github.com/gogpu/ui/internal/render"
	"github.com/gogpu/ui/internal/render/fonts"
)

// caretTolerance is the agreement required between a caret position and the
// glyph it is supposed to sit in front of. The two are the same sum in the
// same order, so only float32 rounding separates them; the defect this guards
// against is a whole pixel wide.
const caretTolerance = 0.01

// shapingSamples exercise both ways a prefix measurement parts from the drawn
// glyph run: the first four carry GPOS kerning pairs the cut would drop, the
// last carries none at all and still diverges, because a prefix is summed from
// grid-fitted advances while the renderer positions from shaped ones.
var shapingSamples = []string{
	"To The Top",
	"Yo, Tavi.",
	"Vy Ya P.",
	"AV Wave",
	"common sense",
}

// shapedPen returns the pen position of every rune boundary of s when the
// whole string is shaped, which is where the renderer puts the glyphs:
// [scene.Scene.DrawText] positions each glyph from the same [text.Shape] run.
// It is the oracle for "where does this character actually begin".
func shapedPen(t *testing.T, s string, fontSize float32) []float64 {
	t.Helper()

	source, err := text.NewFontSource(fonts.InterRegular)
	if err != nil {
		t.Fatalf("load default font: %v", err)
	}
	face := source.Face(float64(fontSize))

	runes := []rune(s)
	pen := make([]float64, len(runes)+1)
	glyphs := text.Shape(s, face)

	var x float64
	var g int
	for i := 1; i <= len(runes); i++ {
		for g < len(glyphs) && glyphs[g].Cluster < i {
			x += glyphs[g].XAdvance
			g++
		}
		pen[i] = x
	}
	return pen
}

func newShapingMetrics(fontSize float32) *Metrics {
	return &Metrics{
		Canvas:   render.NewSceneCanvas(scene.NewScene(), 400, 60),
		FontSize: fontSize,
	}
}

// TestCursorXMatchesShapedGlyphPositions pins the caret to the glyph run the
// renderer draws. Measuring the prefix on its own drops any kerning pair that
// spans the cut, so the caret drifts into the glyph it should precede.
func TestCursorXMatchesShapedGlyphPositions(t *testing.T) {
	const fontSize float32 = 14

	m := newShapingMetrics(fontSize)
	contentRect := geometry.NewRect(10, 0, 380, 30)
	baseX := contentRect.Min.X

	for _, s := range shapingSamples {
		t.Run(s, func(t *testing.T) {
			pen := shapedPen(t, s, fontSize)
			for i := range pen {
				got := m.CursorX(contentRect, s, i)
				want := baseX + float32(pen[i])
				if diff := absf(got - want); diff > caretTolerance {
					t.Errorf("CursorX(%q, %d) = %.4f, glyph begins at %.4f (off by %.4f px)",
						s, i, got, want, diff)
				}
			}
		})
	}
}

// TestRuneIndexFromXMatchesShapedGlyphPositions covers the hit-testing
// direction: a click on the exact start of a character must resolve to that
// character. It is the same defect seen from the other side — the click is
// compared against prefix widths that the renderer never used.
func TestRuneIndexFromXMatchesShapedGlyphPositions(t *testing.T) {
	const fontSize float32 = 14

	m := newShapingMetrics(fontSize)
	contentRect := geometry.NewRect(10, 0, 380, 30)
	baseX := contentRect.Min.X

	for _, s := range shapingSamples {
		t.Run(s, func(t *testing.T) {
			pen := shapedPen(t, s, fontSize)
			for i := range pen {
				x := baseX + float32(pen[i])
				if got := m.RuneIndexFromX(contentRect, s, x); got != i {
					t.Errorf("RuneIndexFromX(%q, %.4f) = %d, want %d", s, x, got, i)
				}
			}
		})
	}
}

// TestSamplesDivergeFromAPrefixMeasurement guards the fixture rather than the
// code. If the default font ever stopped kerning these pairs and started
// shaping them at their grid-fitted advances, the two tests above would pass
// against a prefix measurement and prove nothing.
func TestSamplesDivergeFromAPrefixMeasurement(t *testing.T) {
	const fontSize float32 = 14

	canvas := render.NewSceneCanvas(scene.NewScene(), 400, 60)

	var diverged int
	var worst float32
	for _, s := range shapingSamples {
		pen := shapedPen(t, s, fontSize)
		runes := []rune(s)
		var sampleWorst float32
		for i := 1; i <= len(runes); i++ {
			prefix := canvas.MeasureText(string(runes[:i]), fontSize, false)
			if d := absf(prefix - float32(pen[i])); d > sampleWorst {
				sampleWorst = d
			}
		}
		if sampleWorst > caretTolerance {
			diverged++
		}
		if sampleWorst > worst {
			worst = sampleWorst
		}
	}
	if diverged == 0 {
		t.Fatal("no sample diverges from a prefix measurement: the fixture cannot detect the defect")
	}
	t.Logf("%d of %d samples diverge from a prefix measurement, worst %.3f px",
		diverged, len(shapingSamples), worst)
}

func absf(v float32) float32 {
	if v < 0 {
		return -v
	}
	return v
}
