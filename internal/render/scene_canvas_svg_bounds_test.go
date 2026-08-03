package render

import (
	"testing"

	"github.com/gogpu/gg/scene"
	"github.com/gogpu/ui/geometry"
	"github.com/gogpu/ui/widget"
)

// overflowingSVG draws a rectangle four times the size of its own viewBox.
//
// That is not a malformed document — an SVG may draw anywhere, and content
// outside the viewBox is clipped by the viewport in every renderer that follows
// the spec. Real icon sets depend on it: a stroke centered on the viewBox edge
// puts half its width outside, so "bleeds by a pixel" is the common case and a
// deliberate overflow is only the visible version of it.
const overflowingSVG = `<svg xmlns="http://www.w3.org/2000/svg" viewBox="0 0 16 16">` +
	`<rect x="0" y="0" width="64" height="64" fill="currentColor"/></svg>`

// An SVG must paint only inside the bounds it is handed.
//
// The vector path positions the document with a transform built from those
// bounds but does not otherwise constrain it, so anything the document draws
// past its viewBox lands on whatever surrounds the icon. It goes unseen
// wherever another widget paints afterwards and covers the spill — which is why
// it surfaces on the last icon in a row, or one against the edge of a window.
func TestSceneCanvasRenderSVGStaysInsideItsBounds(t *testing.T) {
	globalIconCache.invalidateAll()
	defer globalIconCache.invalidateAll()

	sc := scene.NewScene()
	c := NewSceneCanvas(sc, 400, 400)
	defer c.Close()

	box := geometry.NewRect(40, 40, 16, 16) // a 16x16 icon, as an app draws one
	c.RenderSVG([]byte(overflowingSVG), box, widget.ColorBlack)

	// Scene.Bounds() is the union of the shapes and does not narrow for a clip,
	// so the emitted commands are what has to be checked: the document must be
	// wrapped in one.
	var clips int
	for _, tag := range sc.Flatten().Tags() {
		if tag == scene.TagBeginClip {
			clips++
		}
	}
	if clips == 0 {
		t.Errorf("the SVG was emitted unbounded — %v of scene geometry with no clip around it, "+
			"so a document drawing past its viewBox paints over whatever surrounds the icon "+
			"(bounds given: %v)", sc.Bounds(), box)
	}
}

// Every other draw method on this canvas culls against the current clip. An
// icon with no pixels on screen should cost nothing, not emit its whole
// document into the scene.
func TestSceneCanvasRenderSVGHonorsTheClip(t *testing.T) {
	globalIconCache.invalidateAll()
	defer globalIconCache.invalidateAll()

	sc := scene.NewScene()
	c := NewSceneCanvas(sc, 400, 400)
	defer c.Close()

	c.PushClip(geometry.NewRect(0, 0, 20, 20))
	defer c.PopClip()

	before := sc.Version()
	c.RenderSVG([]byte(overflowingSVG), geometry.NewRect(200, 200, 16, 16), widget.ColorBlack)
	if sc.Version() != before {
		t.Error("an icon entirely outside the clip still emitted scene commands")
	}
}
