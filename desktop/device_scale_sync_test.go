package desktop

import (
	"testing"

	"github.com/gogpu/gg/integration/ggcanvas"
	"github.com/gogpu/gpucontext"
	"github.com/gogpu/gputypes"
	"github.com/gogpu/ui/app"
	"github.com/gogpu/ui/compositor"
	"github.com/gogpu/ui/event"
	"github.com/gogpu/ui/geometry"
	"github.com/gogpu/ui/widget"
)

type scaleSyncDeviceProvider struct{}

func (scaleSyncDeviceProvider) Device() gpucontext.Device { return gpucontext.Device{} }
func (scaleSyncDeviceProvider) Queue() gpucontext.Queue   { return gpucontext.Queue{} }
func (scaleSyncDeviceProvider) SurfaceFormat() gputypes.TextureFormat {
	return gputypes.TextureFormatUndefined
}
func (scaleSyncDeviceProvider) Adapter() gpucontext.Adapter         { return gpucontext.Adapter{} }
func (scaleSyncDeviceProvider) AdapterInfo() gpucontext.AdapterInfo { return gpucontext.AdapterInfo{} }

type scaleSyncBoundary struct {
	widget.WidgetBase
}

func (w *scaleSyncBoundary) Layout(_ widget.Context, cs geometry.Constraints) geometry.Size {
	return cs.Constrain(geometry.Sz(50, 50))
}
func (w *scaleSyncBoundary) Draw(_ widget.Context, _ widget.Canvas)     {}
func (w *scaleSyncBoundary) Event(_ widget.Context, _ event.Event) bool { return false }

func newScaleSyncRenderLoop(t *testing.T, scale float64, released *int) (*renderLoop, *scaleSyncBoundary) {
	t.Helper()
	canvas, err := ggcanvas.NewWithScale(scaleSyncDeviceProvider{}, 64, 64, scale)
	if err != nil {
		t.Fatalf("ggcanvas.NewWithScale: %v", err)
	}
	t.Cleanup(func() { _ = canvas.Close() })

	uiApp := app.New()
	root := &scaleSyncBoundary{}
	root.SetVisible(true)
	uiApp.Window().SetRoot(root)
	uiApp.Window().Context().SetScale(float32(scale))
	root.SetCachedScene(widget.NewSceneCache())
	root.ClearSceneDirty()
	widget.ClearRedrawInTree(root)
	uiApp.Window().ClearAfterPaint()

	rl := &renderLoop{
		uiApp:            uiApp,
		canvas:           canvas,
		layerTree:        compositor.NewOffsetLayer(geometry.Point{}),
		boundaryTextures: make(map[uint64]*boundaryTexEntry),
	}
	for key := uint64(1); key <= 3; key++ {
		rl.boundaryTextures[key] = &boundaryTexEntry{
			release: func() { (*released)++ },
			width:   48,
			height:  48,
		}
	}
	return rl, root
}

func TestSyncDeviceScaleFlip(t *testing.T) {
	released := 0
	rl, root := newScaleSyncRenderLoop(t, 1, &released)

	if !rl.syncDeviceScale(2) {
		t.Fatal("syncDeviceScale should report a scale change")
	}
	if got := rl.canvas.DeviceScale(); got != 2 {
		t.Errorf("canvas device scale = %v, want 2", got)
	}
	if got := rl.uiApp.Window().Context().Scale(); got != 2 {
		t.Errorf("widget context scale = %v, want 2", got)
	}
	if released != 3 {
		t.Errorf("released %d boundary textures, want 3", released)
	}
	if rl.boundaryTextures != nil {
		t.Error("boundary texture cache should be reset")
	}
	if rl.layerTree != nil {
		t.Error("layer tree should be reset")
	}
	if !rl.fullRedrawNeeded {
		t.Error("scale change should force a full redraw")
	}
	if !root.IsSceneDirty() {
		t.Error("retained scenes should be invalidated at the new scale")
	}
}

func TestSyncDeviceScaleNoOps(t *testing.T) {
	for _, scale := range []float64{2, 0, -1} {
		t.Run(testScaleName(scale), func(t *testing.T) {
			released := 0
			rl, root := newScaleSyncRenderLoop(t, 2, &released)

			if rl.syncDeviceScale(scale) {
				t.Fatalf("syncDeviceScale(%v) should report no change", scale)
			}
			if got := rl.canvas.DeviceScale(); got != 2 {
				t.Errorf("canvas device scale = %v, want 2", got)
			}
			if got := rl.uiApp.Window().Context().Scale(); got != 2 {
				t.Errorf("widget context scale = %v, want 2", got)
			}
			if released != 0 {
				t.Errorf("released %d boundary textures, want 0", released)
			}
			if rl.boundaryTextures == nil || rl.layerTree == nil {
				t.Error("render caches should remain intact")
			}
			if rl.fullRedrawNeeded {
				t.Error("no-op should not force a full redraw")
			}
			if root.IsSceneDirty() {
				t.Error("no-op should not invalidate retained scenes")
			}
		})
	}
}

func testScaleName(scale float64) string {
	switch scale {
	case 0:
		return "zero"
	case -1:
		return "negative"
	default:
		return "unchanged"
	}
}
