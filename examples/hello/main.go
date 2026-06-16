// Example: gogpu/ui — Widget Demo
//
// Demonstrates the gogpu/ui widget toolkit rendering into a gogpu window
// using the managed render loop provided by [desktop.Run].
//
// Architecture:
//
//	ui widgets → render.Canvas (gg) → ggcanvas → GPU surface (zero-copy)
//
// Rendering: event-driven (ContinuousRender=false).
// 0% CPU when idle. Redraws only on user interaction (click, key, resize).
package main

import (
	"fmt"
	"log"

	_ "github.com/gogpu/gg/gpu" // enable GPU SDF acceleration

	"github.com/gogpu/gogpu"
	"github.com/gogpu/ui/app"
	"github.com/gogpu/ui/core/checkbox"
	"github.com/gogpu/ui/core/listview"
	"github.com/gogpu/ui/core/radio"
	"github.com/gogpu/ui/desktop"
	"github.com/gogpu/ui/primitives"
	"github.com/gogpu/ui/theme/material3"
	"github.com/gogpu/ui/widget"
)

func main() {
	m3 := material3.New(widget.Hex(0x6750A4))

	gogpuApp := gogpu.NewApp(gogpu.DefaultConfig().
		WithTitle("gogpu/ui — Widget Demo").
		WithSize(800, 900).
		WithContinuousRender(false))

	uiApp := app.New(
		app.WithWindowProvider(gogpuApp),
		app.WithPlatformProvider(gogpuApp),
		app.WithEventSource(gogpuApp.EventSource()),
		app.WithTheme(m3.AsTheme()),
	)
	uiApp.SetRoot(buildUI(m3))

	if err := desktop.Run(gogpuApp, uiApp); err != nil {
		log.Fatal(err)
	}
}

func buildUI(m3 *material3.Theme) *primitives.BoxWidget {
	card := primitives.Box(
		primitives.Text("gogpu/ui — Widget Demo").
			FontSize(28).
			Bold().
			Color(widget.RGBA8(33, 33, 33, 255)),

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

		primitives.Text("ListView (1000 items)").
			FontSize(18).
			Bold().
			Color(widget.RGBA8(66, 66, 66, 255)),

		primitives.Box(buildListView(m3)).
			Height(300).
			Rounded(8).
			Background(widget.RGBA8(255, 255, 255, 255)).
			BorderStyle(1, widget.RGBA8(218, 218, 218, 255)),
	).
		Padding(32).
		Gap(12).
		Background(widget.RGBA8(255, 255, 255, 255)).
		Rounded(12).
		ShadowLevel(2)

	return primitives.Box(card).Padding(24)
}

func buildListView(m3 *material3.Theme) *listview.Widget {
	items := make([]string, 1000)
	for i := range items {
		items[i] = fmt.Sprintf("Item %d — Lorem ipsum dolor sit amet", i+1)
	}

	return listview.New(
		listview.ItemCount(len(items)),
		listview.FixedItemHeight(36),
		listview.SelectionModeOpt(listview.SelectionSingle),
		listview.PainterOpt(material3.ListViewPainter{Theme: m3}),
		listview.BuildItem(func(ctx listview.ItemContext) widget.Widget {
			color := widget.RGBA8(33, 33, 33, 255)
			if ctx.Selected {
				color = widget.RGBA8(103, 80, 164, 255)
			}
			t := primitives.Text(items[ctx.Index]).
				FontSize(14).
				Color(color)
			if ctx.Selected {
				t = t.Bold()
			}
			return primitives.Box(t).PaddingXY(12, 8)
		}),
		listview.Divider(true),
		listview.OnItemClick(func(index int) {
			fmt.Printf("clicked: %s\n", items[index])
		}),
	)
}
