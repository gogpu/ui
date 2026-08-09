package desktop

import (
	"image"
	"testing"

	"github.com/gogpu/gg"
	"github.com/gogpu/ui/geometry"
)

// dirtyOverlay.draw must NOT pre-multiply its rect by DeviceScale.
// Window.DirtyRegions returns LOGICAL (layout-space) rects, and gg's
// Fill/Stroke apply the device matrix themselves (Context.deviceSpacePath).
// Pre-multiplying scaled the overlay by deviceScale TWICE -- 4x on a 2x
// display, 9x on 3x.
//
// The overlay is env-gated in production (GOGPU_DEBUG_DIRTY=overlay, sync.Once),
// so this drives update/draw directly and bypasses the gate entirely.
//
// (Finding 3 in issue #195.)
func TestDirtyOverlay_DrawsAtOneDeviceScale(t *testing.T) {
	tests := []struct {
		name    string
		logical int // logical canvas size (square)
		scale   float64
		region  geometry.Rect
		want    image.Rectangle // expected drawn-pixel bbox, PHYSICAL
	}{
		// 1x: physical == logical. Non-discriminating (the bug is invisible at
		// scale 1 by construction) but guards against over-correcting.
		{"1x identity", 100, 1.0,
			geometry.NewRect(10, 10, 20, 20), image.Rect(10, 10, 30, 30)},
		// 2x: pre-fix this drew at (40,40)-(120,120).
		{"2x retina", 100, 2.0,
			geometry.NewRect(10, 10, 20, 20), image.Rect(20, 20, 60, 60)},
		// 3x: pre-fix this drew at (90,90)-(270,270).
		{"3x", 100, 3.0,
			geometry.NewRect(10, 10, 20, 20), image.Rect(30, 30, 90, 90)},
		// Off-origin, to catch a translation-only regression.
		{"2x off-origin", 200, 2.0,
			geometry.NewRect(50, 30, 40, 20), image.Rect(100, 60, 180, 100)},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cc := gg.NewContextWithScale(tt.logical, tt.logical, tt.scale)
			// Pin the CPU analytic rasterizer for deterministic, readable-back
			// pixels. The desktop test binary may register gg's GPU accelerator
			// via blank imports, so force CPU path explicitly.
			cc.SetRasterizerMode(gg.RasterizerAnalytic)

			var o dirtyOverlay
			o.update([]geometry.Rect{tt.region})
			if len(o.flashes) != 1 {
				t.Fatalf("update produced %d flashes, want 1", len(o.flashes))
			}
			o.draw(cc)

			img, ok := cc.Image().(*image.RGBA)
			if !ok {
				t.Fatalf("cc.Image() is %T, want *image.RGBA", cc.Image())
			}
			got := alphaBBoxDirty(img, 16)
			if !rectWithinDirty(got, tt.want, 1) {
				t.Errorf("drawn bbox = %v, want %v (+/-1px)\n"+
					"a bbox at ~%dx the expected offset/size means the rect was "+
					"scaled by deviceScale twice",
					got, tt.want, int(tt.scale))
			}
		})
	}
}

// TestDirtyOverlay_DoesNotLeakPaintState pins the Push/Identity/Pop wrapper:
// the overlay must not leave its color or line width on the caller's context.
func TestDirtyOverlay_DoesNotLeakPaintState(t *testing.T) {
	cc := gg.NewContextWithScale(100, 100, 2.0)
	cc.SetRasterizerMode(gg.RasterizerAnalytic)
	cc.SetRGBA(1, 0, 0, 1)
	cc.SetLineWidth(7)

	var o dirtyOverlay
	o.update([]geometry.Rect{geometry.NewRect(10, 10, 20, 20)})
	o.draw(cc)

	// Stroke a rect well clear of the overlay and check it came out red, not
	// cyan -- i.e. the caller's paint survived.
	cc.DrawRectangle(70, 70, 20, 20)
	if err := cc.Stroke(); err != nil {
		t.Fatalf("Stroke: %v", err)
	}
	img, ok := cc.Image().(*image.RGBA)
	if !ok {
		t.Fatalf("cc.Image() is %T, want *image.RGBA", cc.Image())
	}
	px := img.RGBAAt(140, 150) // on the left edge of the stroked rect, physical
	if px.R < 128 || px.B > 64 {
		t.Errorf("post-overlay stroke color = %+v, want red-dominant "+
			"(overlay leaked SetRGBA into the caller's paint)", px)
	}
}

// TestDirtyOverlay_DegenerateRectNoInvertedBorder pins the w>2&&h>2 border
// guard: at logical scale a dirty region as small as 1-2px wide is ordinary.
// Before the guard, the border's 1px inset (w-2/h-2) went negative for such
// rects, drawing an inverted, offset border box.
func TestDirtyOverlay_DegenerateRectNoInvertedBorder(t *testing.T) {
	cc := gg.NewContextWithScale(100, 100, 3.0)
	cc.SetRasterizerMode(gg.RasterizerAnalytic)

	var o dirtyOverlay
	// Logical 1x1 rect -> physical 3x3 fill at (30,30)-(33,33).
	o.update([]geometry.Rect{geometry.NewRect(10, 10, 1, 1)})
	o.draw(cc)

	img, ok := cc.Image().(*image.RGBA)
	if !ok {
		t.Fatalf("cc.Image() is %T, want *image.RGBA", cc.Image())
	}
	got := alphaBBoxDirty(img, 16)
	want := image.Rect(30, 30, 33, 33) // fill only -- border guard suppresses the stroke
	if !rectWithinDirty(got, want, 1) {
		t.Errorf("drawn bbox = %v, want %v (+/-1px) -- a bbox extending outside "+
			"or offset from the fill means the border's w-2/h-2 inset went "+
			"negative", got, want)
	}
}

// alphaBBoxDirty returns the bounding box of all pixels whose alpha is >=
// minAlpha. Returns the zero rectangle when nothing was drawn.
func alphaBBoxDirty(img *image.RGBA, minAlpha uint8) image.Rectangle {
	b := img.Bounds()
	var out image.Rectangle
	first := true
	for y := b.Min.Y; y < b.Max.Y; y++ {
		for x := b.Min.X; x < b.Max.X; x++ {
			if img.RGBAAt(x, y).A < minAlpha {
				continue
			}
			p := image.Rect(x, y, x+1, y+1)
			if first {
				out, first = p, false
			} else {
				out = out.Union(p)
			}
		}
	}
	return out
}

func rectWithinDirty(got, want image.Rectangle, tol int) bool {
	d := func(a, b int) int {
		if a > b {
			return a - b
		}
		return b - a
	}
	return d(got.Min.X, want.Min.X) <= tol && d(got.Min.Y, want.Min.Y) <= tol &&
		d(got.Max.X, want.Max.X) <= tol && d(got.Max.Y, want.Max.Y) <= tol
}
