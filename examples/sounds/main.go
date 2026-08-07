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
	sound.SetEnabled(true)

	m3 := material3.New(widget.Hex(0x1565C0))
	bp := material3.ButtonPainter{Theme: m3}
	cp := material3.CheckboxPainter{Theme: m3}
	rp := material3.RadioPainter{Theme: m3}
	sp := material3.SliderPainter{Theme: m3}

	root := primitives.VBox(
		primitives.Text("System Sounds Demo").FontSize(22).Bold(),
		primitives.Text("Every interaction plays a platform system sound").
			FontSize(13).Color(widget.RGBA(0.5, 0.5, 0.5, 1)),

		primitives.Text("Checkboxes").FontSize(16).Bold(),
		checkbox.New(
			checkbox.LabelOpt("Enable notifications"),
			checkbox.OnToggle(func(checked bool) {
				sound.Play(sound.Click)
				fmt.Println("notifications:", checked)
			}),
			checkbox.PainterOpt(cp),
		),
		checkbox.New(
			checkbox.LabelOpt("Dark mode"),
			checkbox.OnToggle(func(checked bool) {
				sound.Play(sound.Click)
				fmt.Println("dark mode:", checked)
			}),
			checkbox.PainterOpt(cp),
		),

		primitives.Text("Theme").FontSize(16).Bold(),
		radio.NewGroup(
			radio.Items(
				radio.ItemDef{Value: "light", Label: "Light"},
				radio.ItemDef{Value: "dark", Label: "Dark"},
				radio.ItemDef{Value: "system", Label: "System"},
			),
			radio.Selected("light"),
			radio.OnChange(func(v string) {
				sound.Play(sound.Click)
				fmt.Println("theme:", v)
			}),
			radio.GroupPainter(rp),
		),

		primitives.Text("Volume").FontSize(16).Bold(),
		slider.New(
			slider.Min(0),
			slider.Max(100),
			slider.Value(50),
			slider.OnChange(func(v float32) {
				fmt.Printf("volume: %.0f%%\n", v)
			}),
			slider.PainterOpt(sp),
		),

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
		WithSize(400, 520))

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
