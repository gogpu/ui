package main

import (
	"fmt"
	"log"

	_ "github.com/gogpu/gg/gpu"

	"github.com/gogpu/gogpu"
	"github.com/gogpu/gogpu/sound"
	"github.com/gogpu/ui/app"
	"github.com/gogpu/ui/core/button"
	"github.com/gogpu/ui/core/checkbox"
	"github.com/gogpu/ui/core/radio"
	"github.com/gogpu/ui/core/slider"
	"github.com/gogpu/ui/desktop"
	"github.com/gogpu/ui/primitives"
	"github.com/gogpu/ui/theme/material3"
	"github.com/gogpu/ui/widget"
)

func main() {
	// Enable platform sounds globally. All interactive widgets (button,
	// checkbox, radio, dropdown, collapsible, tabview, menu) now auto-play
	// a click sound on activation — no manual sound.Play needed.
	sound.SetEnabled(true)

	m3 := material3.New(widget.Hex(0x1565C0))
	bp := material3.ButtonPainter{Theme: m3}
	cp := material3.CheckboxPainter{Theme: m3}
	rp := material3.RadioPainter{Theme: m3}
	sp := material3.SliderPainter{Theme: m3}

	root := primitives.VBox(
		primitives.Text("System Sounds Demo").FontSize(22).Bold(),
		primitives.Text("Widgets auto-play sounds — no manual sound.Play needed").
			FontSize(13).Color(widget.RGBA(0.5, 0.5, 0.5, 1)),

		primitives.Text("Checkboxes (auto click sound)").FontSize(16).Bold(),
		checkbox.New(
			checkbox.LabelOpt("Enable notifications"),
			checkbox.OnToggle(func(checked bool) {
				fmt.Println("notifications:", checked)
			}),
			checkbox.PainterOpt(cp),
		),
		checkbox.New(
			checkbox.LabelOpt("Dark mode"),
			checkbox.OnToggle(func(checked bool) {
				fmt.Println("dark mode:", checked)
			}),
			checkbox.PainterOpt(cp),
		),

		primitives.Text("Radio (auto click sound)").FontSize(16).Bold(),
		radio.NewGroup(
			radio.Items(
				radio.ItemDef{Value: "light", Label: "Light"},
				radio.ItemDef{Value: "dark", Label: "Dark"},
				radio.ItemDef{Value: "system", Label: "System"},
			),
			radio.Selected("light"),
			radio.OnChange(func(v string) {
				fmt.Println("theme:", v)
			}),
			radio.GroupPainter(rp),
		),

		primitives.Text("Slider (no sound — continuous drag)").FontSize(16).Bold(),
		slider.New(
			slider.Min(0),
			slider.Max(100),
			slider.Value(50),
			slider.OnChange(func(v float32) {
				fmt.Printf("volume: %.0f%%\n", v)
			}),
			slider.PainterOpt(sp),
		),

		primitives.Text("Buttons (auto click sound)").FontSize(16).Bold(),
		primitives.HBox(
			button.New(
				button.Text("Action 1"),
				button.OnClick(func() { fmt.Println("action 1") }),
				button.PainterOpt(bp),
				button.VariantOpt(button.Filled),
			),
			button.New(
				button.Text("Action 2"),
				button.OnClick(func() { fmt.Println("action 2") }),
				button.PainterOpt(bp),
				button.VariantOpt(button.Tonal),
			),
			button.New(
				button.Text("Action 3"),
				button.OnClick(func() { fmt.Println("action 3") }),
				button.PainterOpt(bp),
				button.VariantOpt(button.Outlined),
			),
		).Gap(8),

		primitives.Text("Special sounds (manual)").FontSize(16).Bold(),
		primitives.HBox(
			button.New(
				button.Text("Success"),
				button.OnClick(func() { sound.Play(sound.Success) }),
				button.PainterOpt(bp),
				button.VariantOpt(button.Filled),
			),
			button.New(
				button.Text("Warning"),
				button.OnClick(func() { sound.Play(sound.Warning) }),
				button.PainterOpt(bp),
				button.VariantOpt(button.Tonal),
			),
			button.New(
				button.Text("Error"),
				button.OnClick(func() { sound.Play(sound.Error) }),
				button.PainterOpt(bp),
				button.VariantOpt(button.Outlined),
			),
		).Gap(8),
	).Padding(24).Gap(12)

	gogpuApp := gogpu.NewApp(gogpu.DefaultConfig().
		WithTitle("gogpu/ui — System Sounds").
		WithSize(400, 600))

	uiApp := app.New(
		app.WithWindowProvider(gogpuApp),
		app.WithPlatformProvider(gogpuApp),
		app.WithEventSource(gogpuApp.EventSource()),
		app.WithTheme(m3.AsTheme()),
	)
	uiApp.SetRoot(root)

	if err := desktop.Run(gogpuApp, uiApp); err != nil {
		log.Fatal(err)
	}
}
