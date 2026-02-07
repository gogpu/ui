package main

import (
	"log"

	"github.com/gogpu/gg/integration/ggcanvas"
	"github.com/gogpu/gogpu"
	"github.com/gogpu/gogpu/gmath"
	"github.com/gogpu/ui/app"
	"github.com/gogpu/ui/primitives"
	"github.com/gogpu/ui/render"
	"github.com/gogpu/ui/widget"
)

func main() {
	gogpuApp := gogpu.NewApp(gogpu.Config{
		Title:  "gogpu/ui — Hello World",
		Width:  800,
		Height: 600,
	})

	uiApp := app.New(
		app.WithWindowProvider(gogpuApp),
		app.WithPlatformProvider(gogpuApp),
		app.WithEventSource(gogpuApp.EventSource()),
	)

	uiApp.SetRoot(buildUI())

	var (
		canvas       *ggcanvas.Canvas
		lastW, lastH int
	)

	gogpuApp.OnDraw(func(dc *gogpu.Context) {
		w, h := dc.Size()
		dc.ClearColor(gmath.Hex(0xF0F0F0))

		// Recreate ggcanvas on first frame or resize.
		if canvas == nil || w != lastW || h != lastH {
			if canvas != nil {
				canvas.Close()
			}
			provider := gogpuApp.GPUContextProvider()
			var err error
			canvas, err = ggcanvas.New(provider, w, h)
			if err != nil {
				log.Printf("ggcanvas: %v", err)
				return
			}
			lastW, lastH = w, h
		}

		// Clear the 2D canvas.
		cc := canvas.Context()
		cc.SetRGBA(0, 0, 0, 0)
		cc.Clear()

		// Wrap gg.Context as widget.Canvas for the UI tree.
		widgetCanvas := render.NewCanvas(cc, w, h)

		// Run layout and draw.
		uiApp.Frame()
		uiApp.Window().DrawTo(widgetCanvas)

		// Blit to GPU.
		if err := canvas.RenderTo(dc.AsTextureDrawer()); err != nil {
			log.Printf("render: %v", err)
		}
	})

	gogpuApp.Run()
}

func buildUI() *primitives.BoxWidget {
	return primitives.Box(
		// Title placeholder (Text draws a tinted rectangle).
		primitives.Text("Hello gogpu/ui!").
			FontSize(32).
			Bold().
			Color(widget.RGBA8(33, 33, 33, 255)),

		// Subtitle.
		primitives.Text("Enterprise-Grade GUI Toolkit for Go").
			FontSize(16).
			Color(widget.RGBA8(100, 100, 100, 255)),

		// Decorative "button" box.
		primitives.Box().
			Width(200).
			Height(44).
			Background(widget.RGBA8(98, 0, 238, 255)).
			Rounded(8),
	).
		Padding(32).
		Gap(16).
		Background(widget.RGBA8(255, 255, 255, 255)).
		Rounded(12).
		ShadowLevel(2)
}
