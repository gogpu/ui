// Example: gogpu/ui — Signals Demo
//
// Demonstrates reactive signal bindings in the gogpu/ui widget toolkit.
// Each section showcases a different signal feature:
//
//   - TextSignal / ContentSignal — reactive text updates
//   - CheckedSignal — two-way checkbox binding
//   - SelectedSignal — two-way radio group binding
//   - DisabledSignal — reactive disabled state
//
// Architecture:
//
//	ui widgets → render.Canvas (gg) → ggcanvas → GPU surface (zero-copy)
//
// Rendering: event-driven (ContinuousRender=false).
// 0% CPU when idle. Redraws only on user interaction (click, key, resize).
//
// Requirements:
//   - gogpu v0.22.9+
//   - gg v0.33.3+
package main

import (
	"fmt"
	"log"

	"github.com/gogpu/gg"
	_ "github.com/gogpu/gg/gpu" // enable GPU SDF acceleration
	"github.com/gogpu/gg/integration/ggcanvas"
	"github.com/gogpu/gogpu"
	"github.com/gogpu/ui/app"
	"github.com/gogpu/ui/core/button"
	"github.com/gogpu/ui/core/checkbox"
	"github.com/gogpu/ui/core/radio"
	"github.com/gogpu/ui/primitives"
	"github.com/gogpu/ui/render"
	"github.com/gogpu/ui/state"
	"github.com/gogpu/ui/widget"
)

func main() {
	// Create gogpu application with builder pattern.
	gogpuApp := gogpu.NewApp(gogpu.DefaultConfig().
		WithTitle("gogpu/ui — Signals Demo").
		WithSize(800, 700).
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

		uiApp.Frame()

		// Resize canvas to match window (handles resize inside draw frame).
		cw, ch := canvas.Size()
		if cw != w || ch != h {
			if err := canvas.Resize(w, h); err != nil {
				log.Printf("resize: %v", err)
			}
			cw, ch = w, h
		}

		// Surface is transient — draw every frame.
		sv := dc.SurfaceView()
		sw, sh := dc.SurfaceSize()

		// Set surface target BEFORE drawing so all GPU operations
		// (including mid-draw flushes) render to the surface.
		gg.SetAcceleratorSurfaceTarget(sv, sw, sh)

		canvas.Draw(func(cc *gg.Context) {
			// Background via GPU-accelerated filled rect.
			cc.SetRGBA(0.94, 0.94, 0.94, 1) // #F0F0F0
			cc.DrawRectangle(0, 0, float64(cw), float64(ch))
			cc.Fill()

			widgetCanvas := render.NewCanvas(cc, cw, ch)
			uiApp.Window().DrawTo(widgetCanvas)
		})

		// Flush any remaining GPU shapes to surface.
		if err := canvas.RenderDirect(sv, sw, sh); err != nil {
			log.Printf("render: %v", err)
		}
	})

	// GPU resources are automatically cleaned up on shutdown:
	// - ggcanvas.Canvas auto-registers with App's ResourceTracker
	// - App.Run() calls tracker.CloseAll() before Renderer.Destroy()
	// - OnClose is still available for additional cleanup if needed
	gogpuApp.OnClose(func() {
		gg.CloseAccelerator()
	})

	// Run application.
	if err := gogpuApp.Run(); err != nil {
		log.Fatal(err)
	}
}

func buildUI() *primitives.BoxWidget {
	// --- 1. Counter with TextSignal ---
	count := state.NewSignal(0)
	countLabel := state.NewSignal("Count: 0")

	// --- 2. Toggle with CheckedSignal (two-way) ---
	darkMode := state.NewSignal(false)

	// --- 3. Radio with SelectedSignal (two-way) ---
	selectedSize := state.NewSignal("medium")

	// --- 4. Disabled control via DisabledSignal ---
	locked := state.NewSignal(false)

	card := primitives.Box(
		// Title.
		primitives.Text("Signals Demo").
			FontSize(28).
			Bold().
			Color(widget.RGBA8(33, 33, 33, 255)),

		// ── Section 1: Counter (TextSignal + ContentSignal) ──
		primitives.Text("1. Counter (TextSignal + ContentSignal)").
			FontSize(18).
			Bold().
			Color(widget.RGBA8(66, 66, 66, 255)),

		// ContentSignal: label updates reactively from the countLabel signal.
		primitives.Text("").
			ContentSignal(countLabel.AsReadonly()).
			FontSize(24).
			Color(widget.RGBA8(25, 118, 210, 255)),

		// TextSignal: button text is bound to a signal.
		button.New(
			button.TextSignal(state.NewSignal("Increment")),
			button.OnClick(func() {
				c := count.Get() + 1
				count.Set(c)
				countLabel.Set(fmt.Sprintf("Count: %d", c))
			}),
		),

		// ── Section 2: CheckedSignal (two-way binding) ──
		primitives.Text("2. Toggle (CheckedSignal)").
			FontSize(18).
			Bold().
			Color(widget.RGBA8(66, 66, 66, 255)),

		// CheckedSignal provides two-way binding: toggling the checkbox
		// writes back to darkMode, and setting darkMode.Set() updates the checkbox.
		checkbox.New(
			checkbox.LabelOpt("Dark Mode"),
			checkbox.CheckedSignal(darkMode),
			checkbox.OnToggle(func(checked bool) {
				fmt.Println("dark mode:", checked)
			}),
		),

		// ── Section 3: SelectedSignal (two-way binding) ──
		primitives.Text("3. Selection (SelectedSignal)").
			FontSize(18).
			Bold().
			Color(widget.RGBA8(66, 66, 66, 255)),

		// SelectedSignal provides two-way binding: selecting a radio item
		// writes back to selectedSize, and setting selectedSize.Set() updates the group.
		radio.NewGroup(
			radio.Items(
				radio.ItemDef{Value: "small", Label: "Small"},
				radio.ItemDef{Value: "medium", Label: "Medium"},
				radio.ItemDef{Value: "large", Label: "Large"},
			),
			radio.SelectedSignal(selectedSize),
			radio.OnChange(func(v string) {
				fmt.Println("size:", v)
			}),
		),

		// ── Section 4: Reactive text from signal (ContentSignal) ──
		primitives.Text("4. Reactive Text (ContentSignal)").
			FontSize(18).
			Bold().
			Color(widget.RGBA8(66, 66, 66, 255)),

		// Display the current selectedSize value reactively via TextFn.
		// TextFn reads the signal on each draw, providing a computed-like pattern.
		primitives.TextFn(func() string {
			return fmt.Sprintf("Selected size: %s", selectedSize.Get())
		}).
			FontSize(16).
			Color(widget.RGBA8(56, 142, 60, 255)),

		// ── Section 5: DisabledSignal ──
		primitives.Text("5. Lock (DisabledSignal)").
			FontSize(18).
			Bold().
			Color(widget.RGBA8(66, 66, 66, 255)),

		// The "Lock" checkbox controls the locked signal.
		checkbox.New(
			checkbox.LabelOpt("Lock the button below"),
			checkbox.CheckedSignal(locked),
			checkbox.OnToggle(func(checked bool) {
				fmt.Println("locked:", checked)
			}),
		),

		// DisabledSignal: button becomes disabled when locked signal is true.
		button.New(
			button.TextOpt("Lockable Button"),
			button.DisabledSignal(locked),
			button.OnClick(func() {
				fmt.Println("button clicked (not locked)")
			}),
		),
	).
		Padding(32).
		Gap(12).
		Background(widget.RGBA8(255, 255, 255, 255)).
		Rounded(12).
		ShadowLevel(2)

	// Outer container provides margin around the card.
	return primitives.Box(card).Padding(24)
}
