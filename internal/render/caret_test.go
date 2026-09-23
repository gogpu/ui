package render

import (
	"testing"

	"github.com/gogpu/gg/scene"
	"github.com/gogpu/gg/text"
	"github.com/gogpu/ui/internal/render/fonts"
	"github.com/gogpu/ui/widget"
)

// caretSamples mix kerned pairs, a plain word, a multi-byte rune and the empty
// string, so the rune/byte distinction in the cluster walk is covered.
var caretSamples = []string{
	"",
	"A",
	"AV Wave",
	"To The Top",
	"hello world",
	"café — naïve",
}

func TestCanvasesImplementTextCaretPositioner(t *testing.T) {
	var _ widget.TextCaretPositioner = (*Canvas)(nil)
	var _ widget.TextCaretPositioner = (*SceneCanvas)(nil)
}

// TestCaretOffsetsContract pins the shape of the result every caller relies on:
// one entry per rune boundary, starting at zero and never going backwards.
func TestCaretOffsetsContract(t *testing.T) {
	const fontSize float32 = 14

	canvases := map[string]widget.TextCaretPositioner{
		"Canvas":      newTestCanvas(400, 60),
		"SceneCanvas": NewSceneCanvas(scene.NewScene(), 400, 60),
	}

	for name, canvas := range canvases {
		t.Run(name, func(t *testing.T) {
			for _, s := range caretSamples {
				offsets := canvas.CaretOffsets(s, fontSize, false)
				if want := len([]rune(s)) + 1; len(offsets) != want {
					t.Errorf("CaretOffsets(%q): %d entries, want %d", s, len(offsets), want)
					continue
				}
				if offsets[0] != 0 {
					t.Errorf("CaretOffsets(%q)[0] = %v, want 0", s, offsets[0])
				}
				for i := 1; i < len(offsets); i++ {
					if offsets[i] < offsets[i-1] {
						t.Errorf("CaretOffsets(%q) decreases at %d: %v then %v",
							s, i, offsets[i-1], offsets[i])
					}
				}
			}
		})
	}
}

// TestCaretOffsetsMatchTheShapedRun checks the offsets against the glyph run
// the renderer positions from. [scene.Scene.DrawText] shapes the whole string
// and encodes each glyph at its shaped X, so the shaped pen positions are
// where the characters actually begin.
func TestCaretOffsetsMatchTheShapedRun(t *testing.T) {
	const fontSize float32 = 14

	source, err := text.NewFontSource(fonts.InterRegular)
	if err != nil {
		t.Fatalf("load default font: %v", err)
	}
	face := source.Face(float64(fontSize))

	canvases := map[string]widget.TextCaretPositioner{
		"Canvas":      newTestCanvas(400, 60),
		"SceneCanvas": NewSceneCanvas(scene.NewScene(), 400, 60),
	}

	for name, canvas := range canvases {
		t.Run(name, func(t *testing.T) {
			for _, s := range caretSamples {
				runes := []rune(s)
				glyphs := text.Shape(s, face)
				offsets := canvas.CaretOffsets(s, fontSize, false)

				var (
					x float64
					g int
				)
				for i := 1; i <= len(runes); i++ {
					for g < len(glyphs) && glyphs[g].Cluster < i {
						x += glyphs[g].XAdvance
						g++
					}
					if diff := offsets[i] - float32(x); diff > 0.01 || diff < -0.01 {
						t.Errorf("CaretOffsets(%q)[%d] = %.4f, glyph begins at %.4f",
							s, i, offsets[i], x)
					}
				}
			}
		})
	}
}

// TestCaretOffsetsAgreeAcrossCanvases keeps the two implementations from
// drifting apart: a caret recorded into a scene and one drawn straight to a
// context resolve the same font, so they must place it identically.
func TestCaretOffsetsAgreeAcrossCanvases(t *testing.T) {
	const fontSize float32 = 14

	direct := newTestCanvas(400, 60)
	recorded := NewSceneCanvas(scene.NewScene(), 400, 60)

	for _, bold := range []bool{false, true} {
		for _, s := range caretSamples {
			a := direct.CaretOffsets(s, fontSize, bold)
			b := recorded.CaretOffsets(s, fontSize, bold)
			if len(a) != len(b) {
				t.Fatalf("CaretOffsets(%q, bold=%v): %d vs %d entries", s, bold, len(a), len(b))
			}
			for i := range a {
				if a[i] != b[i] {
					t.Errorf("CaretOffsets(%q, bold=%v)[%d]: Canvas %v, SceneCanvas %v",
						s, bold, i, a[i], b[i])
				}
			}
		}
	}
}
