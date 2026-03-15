// Example: gogpu/ui -- Widget Gallery
//
// A comprehensive showcase of ALL gogpu/ui widgets in a scrollable window.
// Think "Storybook" for gogpu/ui: every widget variant, state, and
// configuration is demonstrated in a single application.
//
// Architecture:
//
//	ui widgets -> render.Canvas (gg) -> ggcanvas -> GPU surface (zero-copy)
//
// Rendering: event-driven (ContinuousRender=false).
// 0% CPU when idle. Redraws only on user interaction.
//
// Requirements:
//
//	gogpu v0.23.2+
//	gg v0.36.4+
package main

import (
	"fmt"
	"log"
	"math"
	"math/rand/v2"
	"time"

	"github.com/gogpu/gg"
	_ "github.com/gogpu/gg/gpu" // enable GPU SDF acceleration
	"github.com/gogpu/gg/integration/ggcanvas"
	"github.com/gogpu/gogpu"
	"github.com/gogpu/ui/app"
	"github.com/gogpu/ui/core/button"
	"github.com/gogpu/ui/core/checkbox"
	"github.com/gogpu/ui/core/collapsible"
	"github.com/gogpu/ui/core/datatable"
	"github.com/gogpu/ui/core/dialog"
	"github.com/gogpu/ui/core/dropdown"
	"github.com/gogpu/ui/core/linechart"
	"github.com/gogpu/ui/core/listview"
	"github.com/gogpu/ui/core/progress"
	"github.com/gogpu/ui/core/progressbar"
	"github.com/gogpu/ui/core/radio"
	"github.com/gogpu/ui/core/scrollview"
	"github.com/gogpu/ui/core/slider"
	"github.com/gogpu/ui/core/splitview"
	"github.com/gogpu/ui/core/tabview"
	"github.com/gogpu/ui/core/textfield"
	"github.com/gogpu/ui/core/toolbar"
	"github.com/gogpu/ui/core/treeview"
	"github.com/gogpu/ui/icon"
	"github.com/gogpu/ui/primitives"
	"github.com/gogpu/ui/render"
	"github.com/gogpu/ui/theme/material3"
	"github.com/gogpu/ui/widget"
)

// M3 theme, created once and shared by all painters.
var m3 = material3.New(widget.Hex(0x6750A4))

// galleryState holds mutable state for the gallery demo.
type galleryState struct {
	chart       *linechart.Widget
	progressBar *progressbar.Widget
	circularPrg *progress.Widget
}

func main() {
	gogpuApp := gogpu.NewApp(gogpu.DefaultConfig().
		WithTitle("gogpu/ui -- Widget Gallery").
		WithSize(900, 1000).
		WithContinuousRender(false))

	uiApp := app.New(
		app.WithWindowProvider(gogpuApp),
		app.WithPlatformProvider(gogpuApp),
		app.WithEventSource(gogpuApp.EventSource()),
	)

	gs := &galleryState{}
	root := buildGallery(gs)
	uiApp.SetRoot(root)

	// Start simulated data for charts and progress.
	go runSimulatedData(gs, gogpuApp)

	var canvas *ggcanvas.Canvas

	gogpuApp.OnDraw(func(dc *gogpu.Context) {
		w, h := dc.Width(), dc.Height()
		if w <= 0 || h <= 0 {
			return
		}

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

		cw, ch := canvas.Size()
		if cw != w || ch != h {
			if err := canvas.Resize(w, h); err != nil {
				log.Printf("resize: %v", err)
			}
			cw, ch = w, h
		}

		sv := dc.SurfaceView()
		sw, sh := dc.SurfaceSize()
		gg.SetAcceleratorSurfaceTarget(sv, sw, sh)

		canvas.Draw(func(cc *gg.Context) {
			// Light gray background.
			cc.SetRGBA(0.96, 0.96, 0.96, 1)
			cc.DrawRectangle(0, 0, float64(cw), float64(ch))
			cc.Fill()

			widgetCanvas := render.NewCanvas(cc, cw, ch)
			uiApp.Window().DrawTo(widgetCanvas)
		})

		if err := canvas.RenderDirect(sv, sw, sh); err != nil {
			log.Printf("render: %v", err)
		}
	})

	gogpuApp.OnClose(func() {
		gg.CloseAccelerator()
	})

	if err := gogpuApp.Run(); err != nil {
		log.Fatal(err)
	}
}

// --- Gallery Builder ---

func buildGallery(gs *galleryState) *scrollview.Widget {
	// Header.
	header := primitives.HBox(
		primitives.Text("gogpu/ui Widget Gallery").
			FontSize(28).Bold().
			Color(widget.RGBA8(33, 33, 33, 255)),
		// Theme seed color dropdown (cosmetic, demonstrates dropdown).
		dropdown.New(
			dropdown.Items("Material Purple", "Ocean Blue", "Forest Green", "Sunset Orange"),
			dropdown.Selected(0),
			dropdown.PainterOpt(material3.DropdownPainter{Theme: m3}),
			dropdown.OnChange(func(idx int, val string) {
				fmt.Printf("theme: %s (index %d)\n", val, idx)
				// Update theme seed color — all painters hold *Theme pointer,
				// so replacing the struct contents updates everything.
				seeds := []uint32{0x6750A4, 0x0078D4, 0x2E7D32, 0xE65100}
				if idx < len(seeds) {
					*m3 = *material3.New(widget.Hex(seeds[idx]))
				}
			}),
		),
	).Gap(16).PaddingXY(0, 8)

	// Build all sections.
	content := primitives.VBox(
		header,
		buildButtonsSection(),
		buildInputsSection(),
		buildDataDisplaySection(gs),
		buildContainersSection(),
		buildNavigationSection(),
	).Padding(24).Gap(16)

	return scrollview.New(content,
		scrollview.PainterOpt(material3.ScrollbarPainter{Theme: m3}),
	)
}

// --- Section 1: Buttons ---

func buildButtonsSection() *collapsible.Widget {
	// Row 1: all variants.
	variantRow := primitives.HBox(
		button.New(
			button.Text("Filled"),
			button.VariantOpt(button.Filled),
			button.PainterOpt(material3.ButtonPainter{Theme: m3}),
			button.OnClick(func() { fmt.Println("Filled clicked") }),
		),
		button.New(
			button.Text("Outlined"),
			button.VariantOpt(button.Outlined),
			button.PainterOpt(material3.ButtonPainter{Theme: m3}),
			button.OnClick(func() { fmt.Println("Outlined clicked") }),
		),
		button.New(
			button.Text("Text"),
			button.VariantOpt(button.TextOnly),
			button.PainterOpt(material3.ButtonPainter{Theme: m3}),
			button.OnClick(func() { fmt.Println("Text clicked") }),
		),
		button.New(
			button.Text("Tonal"),
			button.VariantOpt(button.Tonal),
			button.PainterOpt(material3.ButtonPainter{Theme: m3}),
			button.OnClick(func() { fmt.Println("Tonal clicked") }),
		),
	).Gap(8)

	// Row 2: sizes.
	sizeRow := primitives.HBox(
		button.New(
			button.Text("Small"),
			button.SizeOpt(button.Small),
			button.PainterOpt(material3.ButtonPainter{Theme: m3}),
		),
		button.New(
			button.Text("Medium"),
			button.SizeOpt(button.Medium),
			button.PainterOpt(material3.ButtonPainter{Theme: m3}),
		),
		button.New(
			button.Text("Large"),
			button.SizeOpt(button.Large),
			button.PainterOpt(material3.ButtonPainter{Theme: m3}),
		),
	).Gap(8)

	// Row 3: disabled states.
	disabledRow := primitives.HBox(
		button.New(
			button.Text("Disabled Filled"),
			button.VariantOpt(button.Filled),
			button.Disabled(true),
			button.PainterOpt(material3.ButtonPainter{Theme: m3}),
		),
		button.New(
			button.Text("Disabled Outlined"),
			button.VariantOpt(button.Outlined),
			button.Disabled(true),
			button.PainterOpt(material3.ButtonPainter{Theme: m3}),
		),
	).Gap(8)

	content := primitives.VBox(
		sectionLabel("Variants"),
		variantRow,
		sectionLabel("Sizes"),
		sizeRow,
		sectionLabel("Disabled"),
		disabledRow,
	).Gap(10).Padding(16).Background(widget.RGBA8(255, 255, 255, 255)).Rounded(8)

	return collapsible.New(
		collapsible.Title("Buttons"),
		collapsible.Content(content),
		collapsible.Expanded(true),
		collapsible.Animated(true),
		collapsible.HeaderHeight(40),
		collapsible.PainterOpt(material3.CollapsiblePainter{Theme: m3}),
	)
}

// --- Section 2: Inputs ---

func buildInputsSection() *collapsible.Widget {
	// TextField.
	tfNormal := textfield.New(
		textfield.Placeholder("Enter your name..."),
		textfield.PainterOpt(material3.TextFieldPainter{Theme: m3}),
		textfield.OnChange(func(v string) {
			fmt.Println("textfield:", v)
		}),
	)

	tfPassword := textfield.New(
		textfield.Placeholder("Password"),
		textfield.InputTypeOpt(textfield.TypePassword),
		textfield.PainterOpt(material3.TextFieldPainter{Theme: m3}),
	)

	tfDisabled := textfield.New(
		textfield.Placeholder("Disabled field"),
		textfield.Disabled(true),
		textfield.PainterOpt(material3.TextFieldPainter{Theme: m3}),
	)

	// Dropdown.
	dd := dropdown.New(
		dropdown.Items("Apple", "Banana", "Cherry", "Date", "Elderberry"),
		dropdown.Placeholder("Select a fruit..."),
		dropdown.PainterOpt(material3.DropdownPainter{Theme: m3}),
		dropdown.OnChange(func(idx int, val string) {
			fmt.Printf("dropdown: %s (index %d)\n", val, idx)
		}),
	)

	// Checkboxes.
	cbRow := primitives.VBox(
		checkbox.New(
			checkbox.LabelOpt("Checked"),
			checkbox.Checked(true),
			checkbox.PainterOpt(material3.CheckboxPainter{Theme: m3}),
			checkbox.OnToggle(func(v bool) { fmt.Println("cb1:", v) }),
		),
		checkbox.New(
			checkbox.LabelOpt("Unchecked"),
			checkbox.PainterOpt(material3.CheckboxPainter{Theme: m3}),
		),
		checkbox.New(
			checkbox.LabelOpt("Indeterminate"),
			checkbox.Indeterminate(true),
			checkbox.PainterOpt(material3.CheckboxPainter{Theme: m3}),
		),
		checkbox.New(
			checkbox.LabelOpt("Disabled"),
			checkbox.Checked(true),
			checkbox.Disabled(true),
			checkbox.PainterOpt(material3.CheckboxPainter{Theme: m3}),
		),
	).Gap(6)

	// Radio group.
	rg := radio.NewGroup(
		radio.Items(
			radio.ItemDef{Value: "option1", Label: "Option 1"},
			radio.ItemDef{Value: "option2", Label: "Option 2"},
			radio.ItemDef{Value: "option3", Label: "Option 3"},
		),
		radio.Selected("option1"),
		radio.GroupPainter(material3.RadioPainter{Theme: m3}),
		radio.OnChange(func(v string) { fmt.Println("radio:", v) }),
	)

	// Horizontal radio.
	rgH := radio.NewGroup(
		radio.Items(
			radio.ItemDef{Value: "left", Label: "Left"},
			radio.ItemDef{Value: "center", Label: "Center"},
			radio.ItemDef{Value: "right", Label: "Right"},
		),
		radio.Selected("center"),
		radio.DirectionOpt(radio.Horizontal),
		radio.GroupPainter(material3.RadioPainter{Theme: m3}),
	)

	// Slider.
	sl := slider.New(
		slider.Value(50),
		slider.Min(0),
		slider.Max(100),
		slider.Step(1),
		slider.PainterOpt(material3.SliderPainter{Theme: m3}),
		slider.OnChange(func(v float32) { fmt.Printf("slider: %.0f\n", v) }),
	)

	slDisabled := slider.New(
		slider.Value(30),
		slider.Min(0),
		slider.Max(100),
		slider.Disabled(true),
		slider.PainterOpt(material3.SliderPainter{Theme: m3}),
	)

	content := primitives.VBox(
		sectionLabel("TextField"),
		tfNormal,
		tfPassword,
		tfDisabled,
		sectionLabel("Dropdown"),
		dd,
		sectionLabel("Checkbox"),
		cbRow,
		sectionLabel("Radio (Vertical)"),
		rg,
		sectionLabel("Radio (Horizontal)"),
		rgH,
		sectionLabel("Slider"),
		sl,
		slDisabled,
	).Gap(10).Padding(16).Background(widget.RGBA8(255, 255, 255, 255)).Rounded(8)

	return collapsible.New(
		collapsible.Title("Inputs"),
		collapsible.Content(content),
		collapsible.Expanded(true),
		collapsible.Animated(true),
		collapsible.HeaderHeight(40),
		collapsible.PainterOpt(material3.CollapsiblePainter{Theme: m3}),
	)
}

// --- Section 3: Data Display ---

func buildDataDisplaySection(gs *galleryState) *collapsible.Widget {
	// ListView (10 items).
	items := make([]string, 10)
	for i := range items {
		items[i] = fmt.Sprintf("List item %d", i+1)
	}
	lv := listview.New(
		listview.ItemCount(len(items)),
		listview.FixedItemHeight(36),
		listview.SelectionModeOpt(listview.SelectionSingle),
		listview.PainterOpt(material3.ListViewPainter{Theme: m3}),
		listview.BuildItem(func(ctx listview.ItemContext) widget.Widget {
			color := widget.RGBA8(33, 33, 33, 255)
			if ctx.Selected {
				color = widget.RGBA8(103, 80, 164, 255)
			}
			t := primitives.Text(items[ctx.Index]).FontSize(14).Color(color)
			if ctx.Selected {
				t = t.Bold()
			}
			return primitives.Box(t).PaddingXY(12, 6)
		}),
		listview.Divider(true),
		listview.OnItemClick(func(index int) {
			fmt.Printf("list item clicked: %s\n", items[index])
		}),
	)

	lvBox := primitives.Box(lv).
		Height(200).
		Rounded(6).
		Background(widget.RGBA8(250, 250, 250, 255)).
		BorderStyle(1, widget.RGBA8(218, 218, 218, 255))

	// DataTable.
	tableData := [][]string{
		{"Alice", "28", "alice@example.com"},
		{"Bob", "34", "bob@example.com"},
		{"Charlie", "22", "charlie@example.com"},
		{"Diana", "31", "diana@example.com"},
		{"Eve", "27", "eve@example.com"},
	}
	dt := datatable.New(
		datatable.Columns([]datatable.Column{
			{Key: "name", Title: "Name", MinWidth: 100, Sortable: true},
			{Key: "age", Title: "Age", Width: 60, Sortable: true},
			{Key: "email", Title: "Email", MinWidth: 150, Sortable: true},
		}),
		datatable.RowCount(len(tableData)),
		datatable.RowHeight(32),
		datatable.CellValue(func(row int, col string) string {
			switch col {
			case "name":
				return tableData[row][0]
			case "age":
				return tableData[row][1]
			case "email":
				return tableData[row][2]
			default:
				return ""
			}
		}),
		datatable.SelectionModeOpt(datatable.SelectionSingle),
		datatable.PainterOpt(material3.DataTablePainter{Theme: m3}),
		datatable.OnRowSelect(func(row int) {
			fmt.Printf("table row: %d\n", row)
		}),
	)
	dtBox := primitives.Box(dt).
		Height(220).
		Rounded(6).
		Background(widget.RGBA8(250, 250, 250, 255)).
		BorderStyle(1, widget.RGBA8(218, 218, 218, 255))

	// TreeView.
	tree := treeview.New(
		treeview.Root(&treeview.TreeNode{
			ID: "root", Label: "Project", Expanded: true,
			Children: []*treeview.TreeNode{
				{
					ID: "src", Label: "src", Expanded: true,
					Children: []*treeview.TreeNode{
						{ID: "main", Label: "main.go"},
						{ID: "util", Label: "util.go"},
					},
				},
				{
					ID: "docs", Label: "docs", Expanded: false,
					Children: []*treeview.TreeNode{
						{ID: "readme", Label: "README.md"},
						{ID: "arch", Label: "ARCHITECTURE.md"},
					},
				},
				{ID: "go-mod", Label: "go.mod"},
			},
		}),
		treeview.ItemHeight(28),
		treeview.SelectionModeOpt(treeview.SelectionSingle),
		treeview.PainterOpt(material3.TreeViewPainter{Theme: m3}),
		treeview.OnSelect(func(n *treeview.TreeNode) {
			fmt.Printf("tree select: %s\n", n.Label)
		}),
	)
	tvBox := primitives.Box(tree).
		Height(180).
		Rounded(6).
		Background(widget.RGBA8(250, 250, 250, 255)).
		BorderStyle(1, widget.RGBA8(218, 218, 218, 255))

	// ProgressBar.
	gs.progressBar = progressbar.New(
		progressbar.Value(0.65),
		progressbar.ShowLabel(true),
		progressbar.Height(20),
		progressbar.Radius(4),
		progressbar.FormatLabelFn(func(v float64) string {
			return fmt.Sprintf("%.0f%%", v*100)
		}),
		progressbar.PainterOpt(material3.ProgressBarPainter{Theme: m3}),
	)

	// Circular progress (determinate).
	gs.circularPrg = progress.New(
		progress.Value(0.42),
		progress.Size(48),
		progress.ShowLabel(true),
		progress.PainterOpt(material3.ProgressPainter{Theme: m3}),
	)

	// Circular progress (indeterminate spinner).
	spinner := progress.New(
		progress.Indeterminate(true),
		progress.Size(48),
		progress.PainterOpt(material3.ProgressPainter{Theme: m3}),
	)

	progressRow := primitives.HBox(
		primitives.VBox(
			sectionLabel("Determinate"),
			gs.circularPrg,
		).Gap(4),
		primitives.VBox(
			sectionLabel("Spinner"),
			spinner,
		).Gap(4),
	).Gap(32)

	// LineChart.
	gs.chart = linechart.New(
		linechart.MaxPoints(30),
		linechart.YRange(0, 100),
		linechart.ShowGrid(true),
		linechart.ShowLabels(true),
		linechart.PainterOpt(material3.LineChartPainter{Theme: m3}),
	)
	gs.chart.AddSeries("Series A", widget.Hex(0x6750A4))
	gs.chart.AddSeries("Series B", widget.Hex(0x16C60C))

	chartBox := primitives.Box(gs.chart).
		Height(150).
		Rounded(6).
		Background(widget.RGBA8(250, 250, 250, 255)).
		BorderStyle(1, widget.RGBA8(218, 218, 218, 255))

	content := primitives.VBox(
		sectionLabel("ListView (10 items)"),
		lvBox,
		sectionLabel("DataTable"),
		dtBox,
		sectionLabel("TreeView"),
		tvBox,
		sectionLabel("ProgressBar"),
		gs.progressBar,
		sectionLabel("Circular Progress"),
		progressRow,
		sectionLabel("LineChart (live)"),
		chartBox,
	).Gap(10).Padding(16).Background(widget.RGBA8(255, 255, 255, 255)).Rounded(8)

	return collapsible.New(
		collapsible.Title("Data Display"),
		collapsible.Content(content),
		collapsible.Expanded(true),
		collapsible.Animated(true),
		collapsible.HeaderHeight(40),
		collapsible.PainterOpt(material3.CollapsiblePainter{Theme: m3}),
	)
}

// --- Section 4: Containers ---

func buildContainersSection() *collapsible.Widget {
	// TabView with 3 tabs.
	tv := tabview.New(
		[]tabview.Tab{
			{
				Label: "Overview",
				Content: primitives.VBox(
					primitives.Text("This is the Overview tab content.").FontSize(14).Color(widget.RGBA8(66, 66, 66, 255)),
					primitives.Text("TabView supports keyboard navigation, lazy content, and closeable tabs.").FontSize(12).Color(widget.RGBA8(120, 120, 120, 255)),
				).Gap(8).Padding(12),
			},
			{
				Label: "Details",
				Content: primitives.VBox(
					primitives.Text("Details panel with more information.").FontSize(14).Color(widget.RGBA8(66, 66, 66, 255)),
					checkbox.New(
						checkbox.LabelOpt("Enable feature"),
						checkbox.PainterOpt(material3.CheckboxPainter{Theme: m3}),
					),
				).Gap(8).Padding(12),
			},
			{
				Label: "Settings",
				Content: primitives.VBox(
					primitives.Text("Application settings would go here.").FontSize(14).Color(widget.RGBA8(66, 66, 66, 255)),
					slider.New(
						slider.Value(75),
						slider.Min(0),
						slider.Max(100),
						slider.PainterOpt(material3.SliderPainter{Theme: m3}),
					),
				).Gap(8).Padding(12),
			},
		},
		tabview.PainterOpt(material3.TabViewPainter{Theme: m3}),
		tabview.OnSelect(func(idx int) {
			fmt.Printf("tab selected: %d\n", idx)
		}),
	)
	tvBox := primitives.Box(tv).
		Height(200).
		Rounded(6).
		Background(widget.RGBA8(250, 250, 250, 255)).
		BorderStyle(1, widget.RGBA8(218, 218, 218, 255))

	// SplitView.
	sv := splitview.New(
		splitview.First(
			primitives.VBox(
				primitives.Text("Left Panel").FontSize(14).Bold().Color(widget.RGBA8(33, 33, 33, 255)),
				primitives.Text("Drag the divider to resize.").FontSize(12).Color(widget.RGBA8(100, 100, 100, 255)),
			).Gap(4).Padding(12),
		),
		splitview.Second(
			primitives.VBox(
				primitives.Text("Right Panel").FontSize(14).Bold().Color(widget.RGBA8(33, 33, 33, 255)),
				primitives.Text("Double-click to collapse.").FontSize(12).Color(widget.RGBA8(100, 100, 100, 255)),
			).Gap(4).Padding(12),
		),
		splitview.InitialRatio(0.4),
		splitview.CollapsibleOpt(true),
		splitview.PainterOpt(material3.SplitViewPainter{Theme: m3}),
	)
	svBox := primitives.Box(sv).
		Height(120).
		Rounded(6).
		Background(widget.RGBA8(250, 250, 250, 255)).
		BorderStyle(1, widget.RGBA8(218, 218, 218, 255))

	// Nested collapsible.
	nested := collapsible.New(
		collapsible.Title("Nested Collapsible"),
		collapsible.Content(
			primitives.VBox(
				primitives.Text("This section is inside another collapsible.").FontSize(13).Color(widget.RGBA8(66, 66, 66, 255)),
				primitives.Text("Collapsible supports animated expand/collapse.").FontSize(12).Color(widget.RGBA8(120, 120, 120, 255)),
			).Gap(4).Padding(12),
		),
		collapsible.Expanded(false),
		collapsible.Animated(true),
		collapsible.HeaderHeight(36),
		collapsible.PainterOpt(material3.CollapsiblePainter{Theme: m3}),
	)

	// Dialog (shown inline as a widget for demo, since dialogs are typically overlays).
	dlg := dialog.New(
		dialog.Title("Confirm Action"),
		dialog.Content(
			primitives.Text("Are you sure you want to proceed?").FontSize(14).Color(widget.RGBA8(66, 66, 66, 255)),
		),
		dialog.Actions(
			dialog.Action{Label: "Cancel", OnClick: func() { fmt.Println("dialog: cancel") }},
			dialog.Action{Label: "Confirm", OnClick: func() { fmt.Println("dialog: confirm") }},
		),
		dialog.MaxWidth(400),
		dialog.PainterOpt(material3.DialogPainter{Theme: m3}),
	)
	dlgBox := primitives.Box(dlg).
		Height(160).
		Rounded(6).
		Background(widget.RGBA8(250, 250, 250, 255)).
		BorderStyle(1, widget.RGBA8(218, 218, 218, 255))

	content := primitives.VBox(
		sectionLabel("TabView (3 tabs)"),
		tvBox,
		sectionLabel("SplitView (draggable divider)"),
		svBox,
		sectionLabel("Collapsible"),
		nested,
		sectionLabel("Dialog"),
		dlgBox,
	).Gap(10).Padding(16).Background(widget.RGBA8(255, 255, 255, 255)).Rounded(8)

	return collapsible.New(
		collapsible.Title("Containers"),
		collapsible.Content(content),
		collapsible.Expanded(true),
		collapsible.Animated(true),
		collapsible.HeaderHeight(40),
		collapsible.PainterOpt(material3.CollapsiblePainter{Theme: m3}),
	)
}

// --- Section 5: Navigation ---

func buildNavigationSection() *collapsible.Widget {
	// Toolbar with icons.
	tb := toolbar.New(
		toolbar.Items(
			toolbar.IconButton("Menu", icon.Menu, func() { fmt.Println("menu") }),
			toolbar.Separator(),
			toolbar.IconButton("Search", icon.Search, func() { fmt.Println("search") }),
			toolbar.IconButton("Settings", icon.Settings, func() { fmt.Println("settings") }),
			toolbar.Spacer(),
			toolbar.IconButton("Add", icon.Add, func() { fmt.Println("add") }),
			toolbar.IconButton("Delete", icon.Delete, func() { fmt.Println("delete") }),
		),
		toolbar.Height(40),
		toolbar.PainterOpt(material3.ToolbarPainter{Theme: m3}),
	)

	// Breadcrumb-style text.
	breadcrumb := primitives.HBox(
		primitives.Text("Home").FontSize(13).Color(widget.Hex(0x6750A4)),
		primitives.Text("/").FontSize(13).Color(widget.RGBA8(160, 160, 160, 255)),
		primitives.Text("Gallery").FontSize(13).Color(widget.Hex(0x6750A4)),
		primitives.Text("/").FontSize(13).Color(widget.RGBA8(160, 160, 160, 255)),
		primitives.Text("Navigation").FontSize(13).Color(widget.RGBA8(66, 66, 66, 255)),
	).Gap(6)

	content := primitives.VBox(
		sectionLabel("Toolbar"),
		tb,
		sectionLabel("Breadcrumb"),
		breadcrumb,
	).Gap(10).Padding(16).Background(widget.RGBA8(255, 255, 255, 255)).Rounded(8)

	return collapsible.New(
		collapsible.Title("Navigation"),
		collapsible.Content(content),
		collapsible.Expanded(true),
		collapsible.Animated(true),
		collapsible.HeaderHeight(40),
		collapsible.PainterOpt(material3.CollapsiblePainter{Theme: m3}),
	)
}

// --- Helpers ---

// sectionLabel returns a small muted label for grouping widgets within a section.
func sectionLabel(text string) *primitives.TextWidget {
	return primitives.Text(text).
		FontSize(12).
		Bold().
		Color(widget.RGBA8(120, 120, 120, 255))
}

// --- Simulated Data ---

// runSimulatedData pushes randomized values to chart and progress widgets at 1 Hz.
func runSimulatedData(gs *galleryState, gogpuApp *gogpu.App) {
	ticker := time.NewTicker(time.Second)
	defer ticker.Stop()

	var valA, valB float64

	for range ticker.C {
		// LineChart data.
		valA = smoothWalk(valA, 10, 90, 8)
		valB = smoothWalk(valB, 5, 60, 5)
		gs.chart.PushValue("Series A", valA)
		gs.chart.PushValue("Series B", valB)

		// ProgressBar: slow oscillation.
		pv := (math.Sin(float64(time.Now().UnixMilli())/5000.0) + 1) / 2.0
		gs.progressBar.SetValue(pv)

		gogpuApp.RequestRedraw()
	}
}

// smoothWalk produces a smooth random walk clamped to [lo, hi].
func smoothWalk(current, lo, hi, step float64) float64 {
	delta := (rand.Float64()*2 - 1) * step
	mid := (lo + hi) / 2
	delta += (mid - current) * 0.05
	next := current + delta
	return math.Max(lo, math.Min(hi, next))
}
