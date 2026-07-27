// Example: gogpu/ui — Viewport3D Demo
//
// Demonstrates the Viewport3D widget which provides a GPU-rendered viewport
// for 3D content, video, or custom shader output. The widget follows the
// Flutter Texture pattern: it owns an offscreen GPU texture that an external
// renderer draws into, and the Layer Tree compositor blits it to the surface.
//
// This example shows Viewport3D integration into a standard widget layout.
// The OnRender callback receives a gpucontext.TextureView for GPU rendering.
// In production, a 3D engine (e.g., gogpu/g3d) would render into this texture.
//
// Architecture:
//
//	Viewport3D widget → offscreen GPU texture → Layer Tree → surface blit
//
// Rendering: event-driven (default since gogpu v0.43.0).
// 0% CPU when idle. The viewport re-renders only when explicitly invalidated
// or when Continuous(true) is set for real-time content.
package main

import (
	"fmt"
	"log"

	_ "github.com/gogpu/gg/gpu" // enable GPU SDF acceleration

	"github.com/gogpu/gogpu"
	"github.com/gogpu/gpucontext"
	"github.com/gogpu/ui/app"
	"github.com/gogpu/ui/core/button"
	"github.com/gogpu/ui/core/viewport3d"
	"github.com/gogpu/ui/desktop"
	"github.com/gogpu/ui/primitives"
	"github.com/gogpu/ui/theme/material3"
	"github.com/gogpu/ui/widget"
)

func main() {
	m3 := material3.New(widget.Hex(0x2196F3))

	gogpuApp := gogpu.NewApp(gogpu.DefaultConfig().
		WithTitle("gogpu/ui — Viewport3D Demo").
		WithSize(700, 550))

	uiApp := app.New(
		app.WithWindowProvider(gogpuApp),
		app.WithPlatformProvider(gogpuApp),
		app.WithEventSource(gogpuApp.EventSource()),
		app.WithTheme(m3.AsTheme()),
	)

	vp := buildViewport()
	uiApp.SetRoot(buildUI(vp))

	if err := desktop.Run(gogpuApp, uiApp); err != nil {
		log.Fatal(err)
	}
}

func buildViewport() *viewport3d.Widget {
	return viewport3d.New(
		viewport3d.Size(400, 300),
		viewport3d.OnRender(func(view gpucontext.TextureView) {
			// In production, a 3D renderer (gogpu/g3d) would issue GPU commands
			// to render a scene into this texture view. For this demo, the
			// texture is allocated but no custom rendering is performed —
			// it shows the widget correctly participates in layout and compositing.
			_ = view
		}),
	)
}

func buildUI(vp *viewport3d.Widget) *primitives.BoxWidget {
	card := primitives.Box(
		primitives.Text("Viewport3D Demo").
			FontSize(24).
			Bold().
			Color(widget.RGBA8(33, 33, 33, 255)),

		primitives.Text("GPU viewport for 3D content, video, or custom shaders").
			FontSize(14).
			Color(widget.RGBA8(100, 100, 100, 255)),

		// The Viewport3D widget — a 400x300 offscreen GPU texture.
		vp,

		// Controls below the viewport.
		primitives.Box(
			button.New(
				button.TextOpt("Invalidate"),
				button.OnClick(func() {
					vp.Invalidate()
					fmt.Println("Viewport invalidated — re-render triggered")
				}),
			),
			button.New(
				button.TextOpt("Toggle Continuous"),
				button.OnClick(func() {
					vp.SetContinuous(!vp.IsContinuous())
					fmt.Printf("Continuous rendering: %v\n", vp.IsContinuous())
				}),
			),
		).Gap(8),

		primitives.Text("Controls: Invalidate triggers a single re-render. "+
			"Toggle Continuous enables/disables per-frame rendering.").
			FontSize(12).
			Color(widget.RGBA8(140, 140, 140, 255)),
	).
		Padding(32).
		Gap(16).
		Background(widget.RGBA8(255, 255, 255, 255)).
		Rounded(12).
		ShadowLevel(2)

	return primitives.Box(card).Padding(24)
}
