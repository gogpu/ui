package widget_test

import (
	"image"
	"testing"

	"github.com/gogpu/gg/scene"
	"github.com/gogpu/ui/event"
	"github.com/gogpu/ui/geometry"
	"github.com/gogpu/ui/widget"
)

// TestCompositorClip_SetGet verifies that CompositorClip can be set and
// retrieved on WidgetBase.
func TestCompositorClip_SetGet(t *testing.T) {
	w := widget.NewWidgetBase()

	if w.HasCompositorClip() {
		t.Error("HasCompositorClip should be false initially")
	}

	clip := geometry.NewRect(10, 20, 200, 300)
	w.SetCompositorClip(clip)

	if !w.HasCompositorClip() {
		t.Error("HasCompositorClip should be true after SetCompositorClip")
	}
	if got := w.CompositorClip(); got != clip {
		t.Errorf("CompositorClip = %v, want %v", got, clip)
	}
}

// TestCompositorClip_ClearCompositorClip verifies the clip can be cleared.
func TestCompositorClip_ClearCompositorClip(t *testing.T) {
	w := widget.NewWidgetBase()

	w.SetCompositorClip(geometry.NewRect(0, 0, 100, 100))
	if !w.HasCompositorClip() {
		t.Fatal("should have clip after set")
	}

	w.ClearCompositorClip()
	if w.HasCompositorClip() {
		t.Error("should not have clip after clear")
	}
}

// TestCompositorClip_ScreenRectIntersectsClip verifies that ScreenBounds
// can be compared with CompositorClip to determine visibility.
func TestCompositorClip_ScreenRectIntersectsClip(t *testing.T) {
	tests := []struct {
		name           string
		screenPos      geometry.Point
		itemW, itemH   float32
		clip           geometry.Rect
		wantIntersects bool
	}{
		{
			name:      "fully inside clip",
			screenPos: geometry.Pt(20, 220),
			itemW:     200, itemH: 40,
			clip:           geometry.NewRect(0, 200, 400, 300), // y: 200→500
			wantIntersects: true,
		},
		{
			name:      "fully above clip",
			screenPos: geometry.Pt(20, 100),
			itemW:     200, itemH: 40,
			clip:           geometry.NewRect(0, 200, 400, 300), // y: 200→500
			wantIntersects: false,
		},
		{
			name:      "fully below clip",
			screenPos: geometry.Pt(20, 510),
			itemW:     200, itemH: 40,
			clip:           geometry.NewRect(0, 200, 400, 300), // y: 200→500
			wantIntersects: false,
		},
		{
			name:      "partially visible top",
			screenPos: geometry.Pt(20, 180),
			itemW:     200, itemH: 40,
			clip:           geometry.NewRect(0, 200, 400, 300), // y: 200→500
			wantIntersects: true,                               // item y:180-220, clip y:200-500 → overlap
		},
		{
			name:      "partially visible bottom",
			screenPos: geometry.Pt(20, 480),
			itemW:     200, itemH: 40,
			clip:           geometry.NewRect(0, 200, 400, 300), // y: 200→500
			wantIntersects: true,                               // item y:480-520, clip y:200-500 → overlap
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			w := widget.NewWidgetBase()
			w.SetBounds(geometry.NewRect(0, 0, tt.itemW, tt.itemH))
			w.SetScreenOrigin(tt.screenPos)
			w.SetCompositorClip(tt.clip)

			screenRect := w.ScreenBounds()
			intersects := screenRect.Intersects(tt.clip)

			if intersects != tt.wantIntersects {
				t.Errorf("intersects=%v, want %v (screen=%v, clip=%v)",
					intersects, tt.wantIntersects, screenRect, tt.clip)
			}
		})
	}
}

// TestDrawChild_StampsCompositorClip verifies that DrawChild stamps
// the compositor clip from the canvas onto skipped boundary children.
func TestDrawChild_StampsCompositorClip(t *testing.T) {
	child := &clipTestWidget{}
	child.SetVisible(true)
	child.SetRepaintBoundary(true)
	child.SetBounds(geometry.NewRect(0, 100, 200, 148))

	viewportClip := geometry.NewRect(0, 200, 400, 600)
	canvas := &clipStampCanvas{
		clipBounds:       viewportClip,
		transformOffset:  geometry.Pt(10, 50),
		screenOriginBase: geometry.Pt(0, 0),
		isBoundary:       true,
	}

	widget.DrawChild(child, nil, canvas)

	if !child.HasCompositorClip() {
		t.Fatal("DrawChild should stamp CompositorClip on skipped boundary child")
	}

	got := child.CompositorClip()
	// Screen-space clip = canvas ClipBounds (which is already in recording coords)
	// shifted by screenOriginBase. ClipBounds returns a Rect with Min/Max,
	// so we build the expected using the same Min/Max shift.
	base := canvas.screenOriginBase
	wantClip := geometry.Rect{
		Min: geometry.Pt(viewportClip.Min.X+base.X, viewportClip.Min.Y+base.Y),
		Max: geometry.Pt(viewportClip.Max.X+base.X, viewportClip.Max.Y+base.Y),
	}
	if got != wantClip {
		t.Errorf("CompositorClip = %v, want %v", got, wantClip)
	}
}

// clipTestWidget is a minimal boundary widget for clip tests.
type clipTestWidget struct {
	widget.WidgetBase
}

func (w *clipTestWidget) Layout(_ widget.Context, c geometry.Constraints) geometry.Size {
	return c.Constrain(geometry.Sz(200, 48))
}

func (w *clipTestWidget) Draw(_ widget.Context, _ widget.Canvas) {}

func (w *clipTestWidget) Event(_ widget.Context, _ event.Event) bool { return false }

func (w *clipTestWidget) Children() []widget.Widget { return nil }

// clipStampCanvas simulates BoundaryRecording with a specific clip rect.
type clipStampCanvas struct {
	clipBounds       geometry.Rect
	transformOffset  geometry.Point
	screenOriginBase geometry.Point
	isBoundary       bool
}

// --- Canvas interface ---
func (c *clipStampCanvas) Clear(widget.Color)                                            {}
func (c *clipStampCanvas) DrawRect(geometry.Rect, widget.Color)                          {}
func (c *clipStampCanvas) FillRectDirect(geometry.Rect, widget.Color)                    {}
func (c *clipStampCanvas) StrokeRect(geometry.Rect, widget.Color, float32)               {}
func (c *clipStampCanvas) DrawRoundRect(geometry.Rect, widget.Color, float32)            {}
func (c *clipStampCanvas) StrokeRoundRect(geometry.Rect, widget.Color, float32, float32) {}
func (c *clipStampCanvas) DrawCircle(geometry.Point, float32, widget.Color)              {}
func (c *clipStampCanvas) StrokeCircle(geometry.Point, float32, widget.Color, float32)   {}
func (c *clipStampCanvas) StrokeArc(geometry.Point, float32, float64, float64, widget.Color, float32) {
}
func (c *clipStampCanvas) DrawLine(geometry.Point, geometry.Point, widget.Color, float32) {}
func (c *clipStampCanvas) DrawText(string, geometry.Rect, float32, widget.Color, bool, widget.TextAlign) {
}
func (c *clipStampCanvas) MeasureText(string, float32, bool) float32 { return 0 }
func (c *clipStampCanvas) DrawImage(image.Image, geometry.Point)     {}
func (c *clipStampCanvas) PushClip(geometry.Rect)                    {}
func (c *clipStampCanvas) PushClipRoundRect(geometry.Rect, float32)  {}
func (c *clipStampCanvas) PopClip()                                  {}
func (c *clipStampCanvas) PushTransform(geometry.Point)              {}
func (c *clipStampCanvas) PopTransform()                             {}
func (c *clipStampCanvas) TransformOffset() geometry.Point           { return c.transformOffset }
func (c *clipStampCanvas) ScreenOriginBase() geometry.Point          { return c.screenOriginBase }
func (c *clipStampCanvas) ClipBounds() geometry.Rect                 { return c.clipBounds }
func (c *clipStampCanvas) ReplayScene(*scene.Scene)                  {}

// --- BoundaryRecorder interface ---
func (c *clipStampCanvas) IsBoundaryRecording() bool { return c.isBoundary }

var _ widget.Canvas = (*clipStampCanvas)(nil)
var _ widget.BoundaryRecorder = (*clipStampCanvas)(nil)

// TestDrawChild_DegenerateClipNormalizesToZeroRect verifies that when the
// canvas clip intersection produces a zero-area rect (widget scrolled fully
// out of viewport), stampCompositorClip normalizes it to an explicit zero
// rect rather than passing through a degenerate positioned rect.
func TestDrawChild_DegenerateClipNormalizesToZeroRect(t *testing.T) {
	tests := []struct {
		name       string
		clipBounds geometry.Rect
		base       geometry.Point
	}{
		{
			name:       "zero width",
			clipBounds: geometry.Rect{Min: geometry.Pt(100, 200), Max: geometry.Pt(100, 500)},
			base:       geometry.Pt(0, 0),
		},
		{
			name:       "zero height",
			clipBounds: geometry.Rect{Min: geometry.Pt(0, 300), Max: geometry.Pt(400, 300)},
			base:       geometry.Pt(0, 0),
		},
		{
			name:       "negative width",
			clipBounds: geometry.Rect{Min: geometry.Pt(200, 100), Max: geometry.Pt(100, 400)},
			base:       geometry.Pt(0, 0),
		},
		{
			name:       "negative height",
			clipBounds: geometry.Rect{Min: geometry.Pt(0, 500), Max: geometry.Pt(400, 200)},
			base:       geometry.Pt(0, 0),
		},
		{
			name:       "zero area with non-zero base",
			clipBounds: geometry.Rect{Min: geometry.Pt(50, 100), Max: geometry.Pt(50, 100)},
			base:       geometry.Pt(10, 20),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			child := &clipTestWidget{}
			child.SetVisible(true)
			child.SetRepaintBoundary(true)
			child.SetBounds(geometry.NewRect(0, 0, 200, 48))

			canvas := &clipStampCanvas{
				clipBounds:       tt.clipBounds,
				transformOffset:  geometry.Pt(0, 0),
				screenOriginBase: tt.base,
				isBoundary:       true,
			}

			widget.DrawChild(child, nil, canvas)

			if !child.HasCompositorClip() {
				t.Fatal("DrawChild should still set CompositorClip even for degenerate clips")
			}

			got := child.CompositorClip()
			want := geometry.Rect{} // explicit zero rect
			if got != want {
				t.Errorf("CompositorClip = %v, want zero rect %v", got, want)
			}
		})
	}
}

// TestDrawChild_ValidClipPassesThrough verifies that a normal positive-area
// clip rect is passed through unchanged (not zeroed).
func TestDrawChild_ValidClipPassesThrough(t *testing.T) {
	child := &clipTestWidget{}
	child.SetVisible(true)
	child.SetRepaintBoundary(true)
	child.SetBounds(geometry.NewRect(0, 0, 200, 48))

	viewportClip := geometry.NewRect(10, 20, 400, 600)
	base := geometry.Pt(5, 10)
	canvas := &clipStampCanvas{
		clipBounds:       viewportClip,
		transformOffset:  geometry.Pt(0, 0),
		screenOriginBase: base,
		isBoundary:       true,
	}

	widget.DrawChild(child, nil, canvas)

	if !child.HasCompositorClip() {
		t.Fatal("should have compositor clip")
	}

	got := child.CompositorClip()
	want := geometry.Rect{
		Min: viewportClip.Min.Add(base),
		Max: viewportClip.Max.Add(base),
	}
	if got != want {
		t.Errorf("CompositorClip = %v, want %v", got, want)
	}
}
