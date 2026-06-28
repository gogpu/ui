// Example: Modular Compositor Pattern
//
// Demonstrates how multiple independent "modules" render offscreen into images
// and a compositor displays them on a single GPU window. This pattern is used
// in smart mirrors, kiosk UIs, and digital signage where each panel is an
// independent process or goroutine.
//
// Architecture:
//
//	[Clock Module goroutine]         -> channel -> [Compositor goroutine]
//	[Notification Module goroutine]  -> channel ->        |
//	                                              [gogpu Window]
//
// In production, goroutines become separate processes communicating via Unix
// sockets or shared memory. Channels simulate the IPC for this example.
//
// Each module uses [offscreen.NewRenderer] to render ui widgets into *image.RGBA
// without a window or GPU. The compositor receives frames via channels and
// composites them onto the window using [gg.DrawImage].
//
// Rendering: event-driven (default since gogpu v0.43.0).
// Modules send frames only when content changes. The compositor redraws
// only when a new frame arrives.
package main

import (
	"image"
	"log"
	"sync"
	"time"

	"github.com/gogpu/gg"
	"github.com/gogpu/gg/integration/ggcanvas"
	"github.com/gogpu/gogpu"
	"github.com/gogpu/ui/offscreen"
	"github.com/gogpu/ui/primitives"
	"github.com/gogpu/ui/theme/material3"
	"github.com/gogpu/ui/widget"
)

// Frame carries a rendered module image to the compositor.
type Frame struct {
	ModuleID string
	Image    *image.RGBA
	X, Y     int // position on screen
}

func main() {
	const (
		windowW = 600
		windowH = 400
	)

	frames := make(chan Frame, 4)

	// Latest frame from each module, protected by mutex.
	var mu sync.Mutex
	latest := make(map[string]Frame)

	app := gogpu.NewApp(gogpu.DefaultConfig().
		WithTitle("gogpu/ui — Modular Compositor").
		WithSize(windowW, windowH))

	// Start module goroutines.
	go clockModule(frames)
	go notificationModule(frames)

	// Drain frames into latest map and request redraw.
	go func() {
		for f := range frames {
			mu.Lock()
			latest[f.ModuleID] = f
			mu.Unlock()
			app.RequestRedraw()
		}
	}()

	var canvas *ggcanvas.Canvas

	app.OnDraw(func(dc *gogpu.Context) {
		w, h := dc.Width(), dc.Height()
		if w <= 0 || h <= 0 {
			return
		}

		// Lazy canvas init.
		if canvas == nil {
			provider := app.GPUContextProvider()
			if provider == nil {
				return
			}
			var err error
			canvas, err = ggcanvas.New(provider, w, h)
			if err != nil {
				log.Printf("ggcanvas.New: %v", err)
				return
			}
		}

		cc := canvas.Context()

		// Dark background.
		cc.SetRGBA(0.08, 0.08, 0.12, 1)
		cc.Clear()

		// Composite each module's latest frame.
		mu.Lock()
		snapshot := make(map[string]Frame, len(latest))
		for k, v := range latest {
			snapshot[k] = v
		}
		mu.Unlock()

		for _, f := range snapshot {
			if f.Image == nil {
				continue
			}
			buf := gg.ImageBufFromImage(f.Image)
			cc.DrawImage(buf, float64(f.X), float64(f.Y))
		}

		// Present: upload pixmap to GPU texture and blit to surface.
		canvas.MarkDirty()
		rt := dc.RenderTarget()
		if err := canvas.Render(rt); err != nil {
			log.Printf("canvas.Render: %v", err)
		}
	})

	app.OnClose(func() {
		gg.CloseAccelerator()
		if canvas != nil {
			_ = canvas.Close()
		}
	})

	if err := app.Run(); err != nil {
		log.Fatal(err)
	}
}

// ---------------------------------------------------------------------------
// Clock Module
// ---------------------------------------------------------------------------

// clockModule renders the current time at 1 Hz and sends frames to the
// compositor. Uses offscreen.NewRenderer to render ui widgets without a window.
func clockModule(out chan<- Frame) {
	const (
		modW = 280
		modH = 60
	)

	m3 := material3.New(widget.Hex(0x6750A4))
	r := offscreen.NewRenderer(modW, modH,
		offscreen.WithTheme(m3),
		offscreen.WithBackground(widget.RGBA8(30, 30, 50, 230)),
	)

	ticker := time.NewTicker(time.Second)
	defer ticker.Stop()

	for {
		now := time.Now().Format("15:04:05")
		label := primitives.Box(
			primitives.Text(now).
				FontSize(32).
				Bold().
				Color(widget.RGBA8(220, 220, 255, 255)),
		).Padding(10)

		r.Render(label)

		out <- Frame{
			ModuleID: "clock",
			Image:    r.Image(),
			X:        160, // centered horizontally: (600-280)/2
			Y:        20,
		}

		<-ticker.C
	}
}

// ---------------------------------------------------------------------------
// Notification Module
// ---------------------------------------------------------------------------

// notificationModule renders a notification card that slides in from the bottom
// every 5 seconds and sends frames at 60 Hz during the animation.
func notificationModule(out chan<- Frame) {
	const (
		modW     = 300
		modH     = 80
		targetY  = 300 // final Y position
		slideMs  = 400 // slide-in duration
		displayS = 3   // seconds to show before hiding
	)

	const notificationModule = "notification"

	m3 := material3.New(widget.Hex(0x2E7D32)) // green seed
	r := offscreen.NewRenderer(modW, modH,
		offscreen.WithTheme(m3),
		offscreen.WithBackground(widget.RGBA8(40, 70, 40, 240)),
	)

	messages := []string{
		"New email from Alice",
		"Build succeeded",
		"Deployment complete",
		"Disk usage: 82%",
	}

	for i := 0; ; i++ {
		msg := messages[i%len(messages)]
		card := primitives.Box(
			primitives.Text("Notification").
				FontSize(11).
				Bold().
				Color(widget.RGBA8(180, 220, 180, 255)),
			primitives.Text(msg).
				FontSize(16).
				Color(widget.RGBA8(230, 255, 230, 255)),
		).Padding(12).Gap(4)

		r.Render(card)
		img := r.Image()

		// Slide-in animation at 60 Hz.
		start := time.Now()
		for {
			elapsed := time.Since(start)
			if elapsed > slideMs*time.Millisecond {
				break
			}
			t := float64(elapsed) / float64(slideMs*time.Millisecond)
			// Ease-out cubic: 1 - (1-t)^3
			inv := 1.0 - t
			ease := 1.0 - inv*inv*inv
			y := int(float64(targetY+modH) - ease*float64(modH))

			out <- Frame{
				ModuleID: notificationModule,
				Image:    img,
				X:        150, // centered: (600-300)/2
				Y:        y,
			}
			time.Sleep(16 * time.Millisecond) // ~60 Hz
		}

		// Hold at final position.
		out <- Frame{
			ModuleID: notificationModule,
			Image:    img,
			X:        150,
			Y:        targetY,
		}

		// Display for a few seconds, then clear.
		time.Sleep(displayS * time.Second)

		out <- Frame{
			ModuleID: notificationModule,
			Image:    nil, // hide
			X:        150,
			Y:        targetY,
		}

		// Pause before next notification.
		time.Sleep(2 * time.Second)
	}
}

// ---------------------------------------------------------------------------
// Production Notes
// ---------------------------------------------------------------------------
//
// To convert this to a multi-process architecture:
//
//  1. Replace channels with Unix domain sockets or shared memory (e.g. mmap).
//  2. Each module becomes its own binary using offscreen.NewRenderer.
//  3. The compositor reads frames from sockets and composites on the GPU window.
//  4. Use protobuf or a simple header (moduleID + width + height + RGBA data)
//     for the wire format.
//
// The offscreen.NewRenderer uses CPU-only rasterization (no GPU needed),
// so modules can run on headless machines or containers.
//
// See also:
//   - github.com/gogpu/ui/offscreen — headless widget rendering
//   - github.com/gogpu/ui/theme/material3 — M3 theme for styled modules
//   - GitHub issue #75 — Magic Mirror modular compositor request
