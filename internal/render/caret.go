package render

import (
	"github.com/gogpu/gg/text"
)

// caretOffsets returns the caret offset before each rune of s, plus a final
// entry after the last rune, taken from the shaped form of the WHOLE string.
//
// This is the positioning the renderer uses: both the scene recorder
// ([scene.Scene.DrawText]) and the GPU glyph pipelines place each glyph from a
// [text.Shape] run of the full string. Measuring a prefix on its own instead
// drops the kerning pair that spans the cut and, on a hinted static font,
// swaps grid-fitted advances in for the shaped ones — so a caret placed that
// way lands inside the character it should precede.
//
// [text.ShapedGlyph.Cluster] is the rune index the glyph came from, so the
// caret before rune i is the sum of the advances of every glyph whose cluster
// is below i. Runes that share a cluster — a ligature, a mark over its base —
// collapse onto the cluster's own offset, which is the conventional behavior
// for a caret that cannot subdivide a glyph.
//
// A nil face yields the same average-width approximation MeasureText falls
// back to, so callers never have to special-case a missing font.
func caretOffsets(face text.Face, s string, fontSize float32) []float32 {
	runes := []rune(s)
	offsets := make([]float32, len(runes)+1)
	if len(runes) == 0 {
		return offsets
	}

	if face == nil {
		for i := 1; i <= len(runes); i++ {
			offsets[i] = float32(i) * fontSize * approxCharWidthRatio
		}
		return offsets
	}

	glyphs := text.Shape(s, face)

	var (
		x float64
		g int
	)
	for i := 1; i <= len(runes); i++ {
		for g < len(glyphs) && glyphs[g].Cluster < i {
			x += glyphs[g].XAdvance
			g++
		}
		offsets[i] = float32(x)
	}
	return offsets
}

// approxCharWidthRatio is the average character width, as a fraction of the
// font size, used when no font face is available. It matches the fallback in
// [Canvas.MeasureText] and [SceneCanvas.MeasureText].
const approxCharWidthRatio = 0.5

// defaultCaretFace returns the embedded default face at the given size, or nil
// when the font failed to load. Both canvases resolve text the same way, so
// they place a caret through one lookup.
func defaultCaretFace(fontSize float32, bold bool) text.Face {
	ensureDefaultFonts()

	source := defaultRegular
	if bold {
		source = defaultBold
	}
	if source == nil {
		return nil
	}
	return source.Face(float64(fontSize))
}
