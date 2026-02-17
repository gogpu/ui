// Example: gogpu/ui — Widget Demo
//
// Demonstrates the gogpu/ui widget toolkit rendering into a gogpu window
// using ggcanvas for GPU-accelerated 2D graphics.
//
// Architecture:
//
//	ui widgets → render.Canvas (gg) → ggcanvas → gogpu.Context (GPU) → Window
//
// Rendering: event-driven (ContinuousRender=false).
// 0% CPU when idle. Redraws only on user interaction (click, key, resize).
//
// Requirements:
//   - gogpu v0.18.1+
//   - gg v0.28.1+
package main

import (
	"fmt"
	"log"

	"github.com/gogpu/gg"
	_ "github.com/gogpu/gg/gpu" // enable GPU SDF acceleration
	"github.com/gogpu/gg/integration/ggcanvas"
	"github.com/gogpu/gogpu"
	"github.com/gogpu/gogpu/gmath"
	"github.com/gogpu/ui/app"
	"github.com/gogpu/ui/core/checkbox"
	"github.com/gogpu/ui/core/radio"
	"github.com/gogpu/ui/primitives"
	"github.com/gogpu/ui/render"
	"github.com/gogpu/ui/widget"
)

func main() {
	// Create gogpu application with builder pattern.
	gogpuApp := gogpu.NewApp(gogpu.DefaultConfig().
		WithTitle("gogpu/ui — Widget Demo").
		WithSize(800, 600).
		WithContinuousRender(false)) // Event-driven: 0% CPU when idle

	// Create UI application wired to gogpu providers.
	uiApp := app.New(
		app.WithWindowProvider(gogpuApp),
		app.WithPlatformProvider(gogpuApp),
		app.WithEventSource(gogpuApp.EventSource()),
	)
	uiApp.SetRoot(buildUI())

	// Canvas for 2D rendering (created lazily).
	var canvas *ggcanvas.Canvas

	gogpuApp.OnDraw(func(dc *gogpu.Context) {
		w, h := dc.Width(), dc.Height()
		if w <= 0 || h <= 0 {
			return
		}

		dc.ClearColor(gmath.Hex(0xF0F0F0))

		// Lazy canvas initialization.
		if canvas == nil {
			provider := gogpuApp.GPUContextProvider()
			if provider == nil {
				return
			}
			var err error
			canvas, err = ggcanvas.New(provider, w, h)
			if err != nil {
				log.Printf("ggcanvas: %v", err)
				return
			}
		}

		// Check if UI needs redraw before Frame() clears the flag.
		needsRedraw := uiApp.Window().NeedsLayout()
		uiApp.Frame()

		// Only redraw and re-upload when UI content changed.
		if needsRedraw {
			cw, ch := canvas.Size()
			canvas.Draw(func(cc *gg.Context) {
				cc.SetRGBA(0, 0, 0, 0)
				cc.Clear()

				widgetCanvas := render.NewCanvas(cc, cw, ch)
				uiApp.Window().DrawTo(widgetCanvas)
			})
		}

		// Display texture (skips GPU upload if not dirty).
		if err := canvas.RenderTo(dc.AsTextureDrawer()); err != nil {
			log.Printf("render: %v", err)
		}
	})

	// Handle window resize.
	gogpuApp.EventSource().OnResize(func(w, h int) {
		if canvas != nil {
			if err := canvas.Resize(w, h); err != nil {
				log.Printf("resize: %v", err)
			}
		}
	})

	// Release GPU resources before renderer destroys the device.
	// OnClose runs on the render thread BEFORE Renderer.Destroy(),
	// so the device is still alive for proper cleanup.
	gogpuApp.OnClose(func() {
		gg.CloseAccelerator()
		if canvas != nil {
			_ = canvas.Close()
			canvas = nil
		}
	})

	// Run application.
	if err := gogpuApp.Run(); err != nil {
		log.Fatal(err)
	}
}

func buildUI() *primitives.BoxWidget {
	return primitives.Box(
		// Title.
		primitives.Text("gogpu/ui — Widget Demo").
			FontSize(28).
			Bold().
			Color(widget.RGBA8(33, 33, 33, 255)),

		// Checkbox section.
		primitives.Text("Checkboxes").
			FontSize(18).
			Bold().
			Color(widget.RGBA8(66, 66, 66, 255)),

		checkbox.New(
			checkbox.LabelOpt("Enable notifications"),
			checkbox.Checked(true),
			checkbox.OnToggle(func(checked bool) {
				fmt.Println("notifications:", checked)
			}),
		),

		checkbox.New(
			checkbox.LabelOpt("Dark mode"),
			checkbox.OnToggle(func(checked bool) {
				fmt.Println("dark mode:", checked)
			}),
		),

		checkbox.New(
			checkbox.LabelOpt("Disabled checkbox"),
			checkbox.Checked(true),
			checkbox.Disabled(true),
		),

		// Radio section.
		primitives.Text("Radio Buttons").
			FontSize(18).
			Bold().
			Color(widget.RGBA8(66, 66, 66, 255)),

		radio.NewGroup(
			radio.Items(
				radio.ItemDef{Value: "small", Label: "Small"},
				radio.ItemDef{Value: "medium", Label: "Medium"},
				radio.ItemDef{Value: "large", Label: "Large"},
			),
			radio.Selected("medium"),
			radio.OnChange(func(v string) {
				fmt.Println("size:", v)
			}),
		),

		// Horizontal radio.
		primitives.Text("Horizontal Radio").
			FontSize(14).
			Color(widget.RGBA8(100, 100, 100, 255)),

		radio.NewGroup(
			radio.Items(
				radio.ItemDef{Value: "light", Label: "Light"},
				radio.ItemDef{Value: "dark", Label: "Dark"},
				radio.ItemDef{Value: "system", Label: "System"},
			),
			radio.Selected("system"),
			radio.DirectionOpt(radio.Horizontal),
			radio.OnChange(func(v string) {
				fmt.Println("theme:", v)
			}),
		),
	).
		Padding(32).
		Gap(12).
		Background(widget.RGBA8(255, 255, 255, 255)).
		Rounded(12).
		ShadowLevel(2)
}
