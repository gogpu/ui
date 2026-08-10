//go:build !nogpu

package desktop

import (
	"errors"
	"reflect"
	"sync"
	"testing"
	"unsafe"

	"github.com/gogpu/gg/integration/ggcanvas"
	"github.com/gogpu/gogpu"
	"github.com/gogpu/gputypes"
	"github.com/gogpu/ui/app"
	"github.com/gogpu/ui/primitives"
	"github.com/gogpu/wgpu"
)

// setPrivateTestField wires gogpu's public objects into a headless software
// render target. gogpu does not currently expose a headless Context constructor.
func setPrivateTestField(target any, name string, value any) {
	v := reflect.ValueOf(target).Elem().FieldByName(name)
	reflect.NewAt(v.Type(), unsafe.Pointer(v.UnsafeAddr())).Elem().Set(reflect.ValueOf(value))
}

// TestRenderLoop_SurfaceResizeRetainsBoundaryTextures exercises the actual draw
// orchestration against wgpu's software backend. It verifies that a resized
// surface is fully recomposed while an unchanged child boundary keeps its
// retained texture entry.
func TestRenderLoop_SurfaceResizeRetainsBoundaryTextures(t *testing.T) {
	t.Setenv("GOGPU_DEBUG_DAMAGE", "overlay")
	debugDamageOnce = sync.Once{}
	t.Cleanup(func() {
		debugDamageOnce = sync.Once{}
		debugDamageEnabled = false
	})

	device, _, cleanup := createSoftwareDevice(t)
	t.Cleanup(cleanup)

	const surfaceSize = 100
	texture, err := device.CreateTexture(&wgpu.TextureDescriptor{
		Label: "resize-surface-e2e",
		Size: wgpu.Extent3D{
			Width: surfaceSize, Height: surfaceSize, DepthOrArrayLayers: 1,
		},
		MipLevelCount: 1,
		SampleCount:   1,
		Dimension:     wgpu.TextureDimension2D,
		Format:        gputypes.TextureFormatRGBA8Unorm,
		Usage:         wgpu.TextureUsageRenderAttachment,
	})
	if err != nil {
		t.Fatalf("create surface texture: %v", err)
	}
	t.Cleanup(texture.Release)
	view, err := device.CreateTextureView(texture, nil)
	if err != nil {
		t.Fatalf("create surface view: %v", err)
	}
	t.Cleanup(view.Release)

	renderer := &gogpu.Renderer{}
	target := &gogpu.RenderTarget{}
	setPrivateTestField(renderer, "device", device)
	setPrivateTestField(renderer, "surfaceFormat", gputypes.TextureFormatRGBA8Unorm)
	setPrivateTestField(renderer, "primary", target)
	setPrivateTestField(target, "renderer", renderer)
	setPrivateTestField(target, "format", gputypes.TextureFormatRGBA8Unorm)
	setPrivateTestField(target, "width", uint32(surfaceSize))
	setPrivateTestField(target, "height", uint32(surfaceSize))
	setPrivateTestField(target, "currentView", view)
	setPrivateTestField(target, "frameStarted", true)

	drawContext := &gogpu.Context{}
	setPrivateTestField(drawContext, "renderer", renderer)
	setPrivateTestField(drawContext, "scaleFactor", 1.0)

	gogpuApp := gogpu.NewApp(gogpu.DefaultConfig().WithSize(surfaceSize, surfaceSize))
	setPrivateTestField(gogpuApp, "renderer", renderer)
	uiApp := app.New(
		app.WithWindowProvider(gogpuApp),
		app.WithPlatformProvider(gogpuApp),
	)
	uiApp.SetRoot(primitives.Box(
		primitives.NewRepaintBoundary(primitives.Box().Width(20).Height(20)),
	).Width(surfaceSize).Height(surfaceSize))

	canvas, err := ggcanvas.New(gogpuApp.GPUContextProvider(), 80, 80)
	if err != nil {
		t.Fatalf("create initial canvas: %v", err)
	}
	t.Cleanup(func() { _ = canvas.Close() })

	rl := &renderLoop{
		gogpuApp: gogpuApp,
		uiApp:    uiApp,
		canvas:   canvas,
	}
	rl.draw(drawContext)

	if rl.surfaceResizePending {
		t.Error("successful full render left surface resize pending")
	}
	if gotW, gotH := canvas.Size(); gotW != surfaceSize || gotH != surfaceSize {
		t.Fatalf("canvas size = %dx%d, want %dx%d", gotW, gotH, surfaceSize, surfaceSize)
	}
	if len(rl.boundaryTextures) == 0 {
		t.Fatal("full recomposition did not populate the retained boundary cache")
	}

}

type failingSurfaceResizer struct{ err error }

func (r failingSurfaceResizer) Resize(_, _ int) error { return r.err }

func TestRenderLoop_SurfaceFailuresRemainPending(t *testing.T) {
	wantErr := errors.New("injected surface failure")
	rl := &renderLoop{}
	if rl.resizeSurface(failingSurfaceResizer{err: wantErr}, 120, 80) {
		t.Fatal("failed resize reported success")
	}
	if rl.surfaceResizePending {
		t.Error("failed resize marked an unresized surface pending")
	}

	rl.surfaceResizePending = true
	rl.finishSurfaceRender(wantErr)
	if !rl.surfaceResizePending {
		t.Error("failed render cleared the pending resize retry")
	}
}
