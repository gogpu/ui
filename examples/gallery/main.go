// Example: gogpu/ui -- Widget Gallery
//
// A comprehensive showcase of ALL gogpu/ui widgets in a scrollable window.
// Think "Storybook" for gogpu/ui: every widget variant, state, and
// configuration is demonstrated in a single application.
//
// Supports theme switching: Material 3 (4 seed colors) and DevTools (dark/light).
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

	_ "github.com/gogpu/gg/gpu" // enable GPU SDF acceleration

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
	"github.com/gogpu/ui/desktop"
	"github.com/gogpu/ui/icon"
	"github.com/gogpu/ui/primitives"
	"github.com/gogpu/ui/theme"
	"github.com/gogpu/ui/theme/devtools"
	"github.com/gogpu/ui/theme/material3"
	"github.com/gogpu/ui/widget"
)

// painterSet abstracts over design systems so build functions work with any theme.
type painterSet struct {
	button      button.Painter
	checkbox    checkbox.Painter
	radio       radio.Painter
	textfield   textfield.Painter
	dropdown    dropdown.Painter
	slider      slider.Painter
	dialog      dialog.Painter
	scrollbar   scrollview.Painter
	tabview     tabview.Painter
	treeview    treeview.Painter
	datatable   datatable.Painter
	toolbar     toolbar.Painter
	collapsible collapsible.Painter
	listview    listview.Painter
	progress    progress.Painter
	linechart   linechart.Painter
	splitview   splitview.Painter
	// progressbar uses ColorScheme, not Painter interface in all themes.
	progressBarColors progressbar.ProgressBarColorScheme
	// isDark controls background and text colors.
	isDark bool
}

// m3Painters creates a painterSet from a Material 3 theme.
func m3Painters(m3 *material3.Theme) painterSet {
	return painterSet{
		button:      material3.ButtonPainter{Theme: m3},
		checkbox:    material3.CheckboxPainter{Theme: m3},
		radio:       material3.RadioPainter{Theme: m3},
		textfield:   material3.TextFieldPainter{Theme: m3},
		dropdown:    material3.DropdownPainter{Theme: m3},
		slider:      material3.SliderPainter{Theme: m3},
		dialog:      material3.DialogPainter{Theme: m3},
		scrollbar:   material3.ScrollbarPainter{Theme: m3},
		tabview:     material3.TabViewPainter{Theme: m3},
		treeview:    material3.TreeViewPainter{Theme: m3},
		datatable:   material3.DataTablePainter{Theme: m3},
		toolbar:     material3.ToolbarPainter{Theme: m3},
		collapsible: material3.CollapsiblePainter{Theme: m3},
		listview:    material3.ListViewPainter{Theme: m3},
		progress:    material3.ProgressPainter{Theme: m3},
		linechart:   material3.LineChartPainter{Theme: m3},
		splitview:   material3.SplitViewPainter{Theme: m3},
		progressBarColors: progressbar.ProgressBarColorScheme{
			Bar:   m3.Colors.Primary,
			Track: m3.Colors.SurfaceVariant,
			Label: m3.Colors.OnSurface,
		},
		isDark: false,
	}
}

// dtPainters creates a painterSet from a DevTools theme.
func dtPainters(dt *devtools.Theme) painterSet {
	p := devtools.NewPainters(dt)
	return painterSet{
		button:      p.Button,
		checkbox:    p.Checkbox,
		radio:       p.Radio,
		textfield:   p.TextField,
		dropdown:    p.Dropdown,
		slider:      p.Slider,
		dialog:      p.Dialog,
		scrollbar:   p.Scrollbar,
		tabview:     p.TabView,
		treeview:    p.TreeView,
		datatable:   p.DataTable,
		toolbar:     p.Toolbar,
		collapsible: p.Collapsible,
		listview:    p.ListView,
		progress:    p.Progress,
		linechart:   p.LineChart,
		splitview:   p.SplitView,
		progressBarColors: progressbar.ProgressBarColorScheme{
			Bar:   dt.Colors.Primary,
			Track: dt.Colors.ControlFill,
			Label: dt.Colors.OnSurface,
		},
		isDark: dt.IsDark(),
	}
}

// galleryState holds mutable state for the gallery demo.
type galleryState struct {
	chart       *linechart.Widget
	progressBar *progressbar.Widget
	circularPrg *progress.Widget
	themeIdx    int // currently selected theme index
}

// theme option names for the design-system dropdown.
var themeNames = []string{
	"M3 Purple", "M3 Blue", "M3 Green", "M3 Orange",
	"DevTools Dark", "DevTools Light",
}

func main() {
	m3 := material3.New(widget.Hex(0x6750A4))

	gogpuApp := gogpu.NewApp(gogpu.DefaultConfig().
		WithTitle("gogpu/ui -- Widget Gallery").
		WithSize(900, 1000).
		WithContinuousRender(false))

	uiApp := app.New(
		app.WithWindowProvider(gogpuApp),
		app.WithPlatformProvider(gogpuApp),
		app.WithEventSource(gogpuApp.EventSource()),
		app.WithTheme(m3.AsTheme()),
	)

	gs := &galleryState{}
	ps := m3Painters(m3)
	var onThemeChange func(int)
	onThemeChange = func(idx int) {
		gs.themeIdx = idx
		ps = switchTheme(idx)
		uiApp.SetTheme(switchUITheme(idx))
		newRoot := buildGallery(gs, ps, onThemeChange)
		uiApp.SetRoot(newRoot)
		gogpuApp.RequestRedraw()
	}
	root := buildGallery(gs, ps, onThemeChange)
	uiApp.SetRoot(root)

	// Start simulated data for charts and progress.
	go runSimulatedData(gs, gogpuApp)

	if err := desktop.Run(gogpuApp, uiApp); err != nil {
		log.Fatal(err)
	}
}

// switchTheme returns the painterSet for the given theme index.
func switchTheme(idx int) painterSet {
	m3Seeds := []uint32{0x6750A4, 0x0078D4, 0x2E7D32, 0xE65100}
	switch {
	case idx < len(m3Seeds):
		return m3Painters(material3.New(widget.Hex(m3Seeds[idx])))
	case idx == 4: //nolint:mnd // DevTools Dark
		return dtPainters(devtools.NewDarkTheme())
	case idx == 5: //nolint:mnd // DevTools Light
		return dtPainters(devtools.NewTheme())
	default:
		return m3Painters(material3.New(widget.Hex(0x6750A4)))
	}
}

// switchUITheme returns the theme.Theme for the given theme index.
// This keeps desktop.Run's background drawing in sync with the painter set.
func switchUITheme(idx int) *theme.Theme {
	m3Seeds := []uint32{0x6750A4, 0x0078D4, 0x2E7D32, 0xE65100}
	switch {
	case idx < len(m3Seeds):
		return material3.New(widget.Hex(m3Seeds[idx])).AsTheme()
	case idx == 4: //nolint:mnd // DevTools Dark
		return devtools.NewDarkTheme().AsTheme()
	case idx == 5: //nolint:mnd // DevTools Light
		return devtools.NewTheme().AsTheme()
	default:
		return material3.New(widget.Hex(0x6750A4)).AsTheme()
	}
}

// --- Gallery Builder ---

// textColor returns the appropriate text color for the current theme.
func (ps painterSet) textColor() widget.Color {
	if ps.isDark {
		return widget.RGBA8(230, 230, 230, 255)
	}
	return widget.RGBA8(33, 33, 33, 255)
}

// subtextColor returns a muted text color for secondary content.
func (ps painterSet) subtextColor() widget.Color {
	if ps.isDark {
		return widget.RGBA8(160, 160, 160, 255)
	}
	return widget.RGBA8(120, 120, 120, 255)
}

// sectionBg returns the section background color for the current theme.
func (ps painterSet) sectionBg() widget.Color {
	if ps.isDark {
		return widget.RGBA8(43, 43, 43, 255)
	}
	return widget.RGBA8(255, 255, 255, 255)
}

// cardBg returns the card/box background color for the current theme.
func (ps painterSet) cardBg() widget.Color {
	if ps.isDark {
		return widget.RGBA8(50, 50, 50, 255)
	}
	return widget.RGBA8(250, 250, 250, 255)
}

// cardBorder returns the card border color for the current theme.
func (ps painterSet) cardBorder() widget.Color {
	if ps.isDark {
		return widget.RGBA8(70, 70, 70, 255)
	}
	return widget.RGBA8(218, 218, 218, 255)
}

func buildGallery(gs *galleryState, ps painterSet, onThemeChange func(int)) *scrollview.Widget {
	// Header.
	themeDropdownOpts := []dropdown.Option{
		dropdown.Items(themeNames...),
		dropdown.Selected(gs.themeIdx),
		dropdown.PainterOpt(ps.dropdown),
	}
	if onThemeChange != nil {
		themeDropdownOpts = append(themeDropdownOpts, dropdown.OnChange(func(idx int, val string) {
			fmt.Printf("theme: %s (index %d)\n", val, idx)
			if onThemeChange != nil {
				onThemeChange(idx)
			}
		}))
	}
	header := primitives.HBox(
		primitives.Text("gogpu/ui Widget Gallery").
			FontSize(28).Bold().
			Color(ps.textColor()),
		dropdown.New(themeDropdownOpts...),
	).Gap(16).PaddingXY(0, 8)

	// Build all sections.
	content := primitives.VBox(
		header,
		buildButtonsSection(ps),
		buildInputsSection(ps),
		buildDataDisplaySection(gs, ps),
		buildContainersSection(ps),
		buildNavigationSection(ps),
	).Padding(24).Gap(16)

	return scrollview.New(content,
		scrollview.PainterOpt(ps.scrollbar),
	)
}

// --- Section 1: Buttons ---

func buildButtonsSection(ps painterSet) *collapsible.Widget {
	// Row 1: all variants.
	variantRow := primitives.HBox(
		button.New(
			button.Text("Filled"),
			button.VariantOpt(button.Filled),
			button.PainterOpt(ps.button),
			button.OnClick(func() { fmt.Println("Filled clicked") }),
		),
		button.New(
			button.Text("Outlined"),
			button.VariantOpt(button.Outlined),
			button.PainterOpt(ps.button),
			button.OnClick(func() { fmt.Println("Outlined clicked") }),
		),
		button.New(
			button.Text("Text"),
			button.VariantOpt(button.TextOnly),
			button.PainterOpt(ps.button),
			button.OnClick(func() { fmt.Println("Text clicked") }),
		),
		button.New(
			button.Text("Tonal"),
			button.VariantOpt(button.Tonal),
			button.PainterOpt(ps.button),
			button.OnClick(func() { fmt.Println("Tonal clicked") }),
		),
	).Gap(8)

	// Row 2: sizes.
	sizeRow := primitives.HBox(
		button.New(
			button.Text("Small"),
			button.SizeOpt(button.Small),
			button.PainterOpt(ps.button),
		),
		button.New(
			button.Text("Medium"),
			button.SizeOpt(button.Medium),
			button.PainterOpt(ps.button),
		),
		button.New(
			button.Text("Large"),
			button.SizeOpt(button.Large),
			button.PainterOpt(ps.button),
		),
	).Gap(8)

	// Row 3: disabled states.
	disabledRow := primitives.HBox(
		button.New(
			button.Text("Disabled Filled"),
			button.VariantOpt(button.Filled),
			button.Disabled(true),
			button.PainterOpt(ps.button),
		),
		button.New(
			button.Text("Disabled Outlined"),
			button.VariantOpt(button.Outlined),
			button.Disabled(true),
			button.PainterOpt(ps.button),
		),
	).Gap(8)

	content := primitives.VBox(
		sectionLabel("Variants", ps),
		variantRow,
		sectionLabel("Sizes", ps),
		sizeRow,
		sectionLabel("Disabled", ps),
		disabledRow,
	).Gap(10).Padding(16).Background(ps.sectionBg()).Rounded(8)

	return collapsible.New(
		collapsible.Title("Buttons"),
		collapsible.Content(content),
		collapsible.Expanded(true),
		collapsible.Animated(true),
		collapsible.HeaderHeight(40),
		collapsible.PainterOpt(ps.collapsible),
	)
}

// --- Section 2: Inputs ---

func buildInputsSection(ps painterSet) *collapsible.Widget {
	// TextField.
	tfNormal := textfield.New(
		textfield.Placeholder("Enter your name..."),
		textfield.PainterOpt(ps.textfield),
		textfield.OnChange(func(v string) {
			fmt.Println("textfield:", v)
		}),
	)

	tfPassword := textfield.New(
		textfield.Placeholder("Password"),
		textfield.InputTypeOpt(textfield.TypePassword),
		textfield.PainterOpt(ps.textfield),
	)

	tfDisabled := textfield.New(
		textfield.Placeholder("Disabled field"),
		textfield.Disabled(true),
		textfield.PainterOpt(ps.textfield),
	)

	// Dropdown.
	dd := dropdown.New(
		dropdown.Items("Apple", "Banana", "Cherry", "Date", "Elderberry"),
		dropdown.Placeholder("Select a fruit..."),
		dropdown.PainterOpt(ps.dropdown),
		dropdown.OnChange(func(idx int, val string) {
			fmt.Printf("dropdown: %s (index %d)\n", val, idx)
		}),
	)

	// Checkboxes.
	cbRow := primitives.VBox(
		checkbox.New(
			checkbox.LabelOpt("Checked"),
			checkbox.Checked(true),
			checkbox.PainterOpt(ps.checkbox),
			checkbox.OnToggle(func(v bool) { fmt.Println("cb1:", v) }),
		),
		checkbox.New(
			checkbox.LabelOpt("Unchecked"),
			checkbox.PainterOpt(ps.checkbox),
		),
		checkbox.New(
			checkbox.LabelOpt("Indeterminate"),
			checkbox.Indeterminate(true),
			checkbox.PainterOpt(ps.checkbox),
		),
		checkbox.New(
			checkbox.LabelOpt("Disabled"),
			checkbox.Checked(true),
			checkbox.Disabled(true),
			checkbox.PainterOpt(ps.checkbox),
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
		radio.GroupPainter(ps.radio),
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
		radio.GroupPainter(ps.radio),
	)

	// Slider.
	sl := slider.New(
		slider.Value(50),
		slider.Min(0),
		slider.Max(100),
		slider.Step(1),
		slider.PainterOpt(ps.slider),
		slider.OnChange(func(v float32) { fmt.Printf("slider: %.0f\n", v) }),
	)

	slDisabled := slider.New(
		slider.Value(30),
		slider.Min(0),
		slider.Max(100),
		slider.Disabled(true),
		slider.PainterOpt(ps.slider),
	)

	content := primitives.VBox(
		sectionLabel("TextField", ps),
		tfNormal,
		tfPassword,
		tfDisabled,
		sectionLabel("Dropdown", ps),
		dd,
		sectionLabel("Checkbox", ps),
		cbRow,
		sectionLabel("Radio (Vertical)", ps),
		rg,
		sectionLabel("Radio (Horizontal)", ps),
		rgH,
		sectionLabel("Slider", ps),
		sl,
		slDisabled,
	).Gap(10).Padding(16).Background(ps.sectionBg()).Rounded(8)

	return collapsible.New(
		collapsible.Title("Inputs"),
		collapsible.Content(content),
		collapsible.Expanded(true),
		collapsible.Animated(true),
		collapsible.HeaderHeight(40),
		collapsible.PainterOpt(ps.collapsible),
	)
}

// --- Section 3: Data Display ---

func buildDataDisplaySection(gs *galleryState, ps painterSet) *collapsible.Widget {
	// ListView (10 items).
	items := make([]string, 10)
	for i := range items {
		items[i] = fmt.Sprintf("List item %d", i+1)
	}
	lv := listview.New(
		listview.ItemCount(len(items)),
		listview.FixedItemHeight(36),
		listview.SelectionModeOpt(listview.SelectionSingle),
		listview.PainterOpt(ps.listview),
		listview.BuildItem(func(ctx listview.ItemContext) widget.Widget {
			color := ps.textColor()
			if ctx.Selected {
				color = ps.subtextColor()
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
		Background(ps.cardBg()).
		BorderStyle(1, ps.cardBorder())

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
		datatable.PainterOpt(ps.datatable),
		datatable.OnRowSelect(func(row int) {
			fmt.Printf("table row: %d\n", row)
		}),
	)
	dtBox := primitives.Box(dt).
		Height(220).
		Rounded(6).
		Background(ps.cardBg()).
		BorderStyle(1, ps.cardBorder())

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
		treeview.PainterOpt(ps.treeview),
		treeview.OnSelect(func(n *treeview.TreeNode) {
			fmt.Printf("tree select: %s\n", n.Label)
		}),
	)
	tvBox := primitives.Box(tree).
		Height(180).
		Rounded(6).
		Background(ps.cardBg()).
		BorderStyle(1, ps.cardBorder())

	// ProgressBar.
	gs.progressBar = progressbar.New(
		progressbar.Value(0.65),
		progressbar.ShowLabel(true),
		progressbar.Height(20),
		progressbar.Radius(4),
		progressbar.FormatLabelFn(func(v float64) string {
			return fmt.Sprintf("%.0f%%", v*100)
		}),
		progressbar.ColorSchemeOpt(ps.progressBarColors),
	)

	// Circular progress (determinate).
	gs.circularPrg = progress.New(
		progress.Value(0.42),
		progress.Size(48),
		progress.ShowLabel(true),
		progress.PainterOpt(ps.progress),
	)

	// Circular progress (indeterminate spinner).
	spinner := progress.New(
		progress.Indeterminate(true),
		progress.Size(48),
		progress.PainterOpt(ps.progress),
	)

	progressRow := primitives.HBox(
		primitives.VBox(
			sectionLabel("Determinate", ps),
			gs.circularPrg,
		).Gap(4),
		primitives.VBox(
			sectionLabel("Spinner", ps),
			spinner,
		).Gap(4),
	).Gap(32)

	// LineChart.
	gs.chart = linechart.New(
		linechart.MaxPoints(30),
		linechart.YRange(0, 100),
		linechart.ShowGrid(true),
		linechart.ShowLabels(true),
		linechart.PainterOpt(ps.linechart),
	)
	gs.chart.AddSeries("Series A", widget.Hex(0x6750A4))
	gs.chart.AddSeries("Series B", widget.Hex(0x16C60C))

	chartBox := primitives.Box(gs.chart).
		Height(150).
		Rounded(6).
		Background(ps.cardBg()).
		BorderStyle(1, ps.cardBorder())

	content := primitives.VBox(
		sectionLabel("ListView (10 items)", ps),
		lvBox,
		sectionLabel("DataTable", ps),
		dtBox,
		sectionLabel("TreeView", ps),
		tvBox,
		sectionLabel("ProgressBar", ps),
		gs.progressBar,
		sectionLabel("Circular Progress", ps),
		progressRow,
		sectionLabel("LineChart (live)", ps),
		chartBox,
	).Gap(10).Padding(16).Background(ps.sectionBg()).Rounded(8)

	return collapsible.New(
		collapsible.Title("Data Display"),
		collapsible.Content(content),
		collapsible.Expanded(true),
		collapsible.Animated(true),
		collapsible.HeaderHeight(40),
		collapsible.PainterOpt(ps.collapsible),
	)
}

// --- Section 4: Containers ---

func buildContainersSection(ps painterSet) *collapsible.Widget {
	// TabView with 3 tabs.
	tv := tabview.New(
		[]tabview.Tab{
			{
				Label: "Overview",
				Content: primitives.VBox(
					primitives.Text("This is the Overview tab content.").FontSize(14).Color(ps.textColor()),
					primitives.Text("TabView supports keyboard navigation, lazy content, and closeable tabs.").FontSize(12).Color(ps.subtextColor()),
				).Gap(8).Padding(12),
			},
			{
				Label: "Details",
				Content: primitives.VBox(
					primitives.Text("Details panel with more information.").FontSize(14).Color(ps.textColor()),
					checkbox.New(
						checkbox.LabelOpt("Enable feature"),
						checkbox.PainterOpt(ps.checkbox),
					),
				).Gap(8).Padding(12),
			},
			{
				Label: "Settings",
				Content: primitives.VBox(
					primitives.Text("Application settings would go here.").FontSize(14).Color(ps.textColor()),
					slider.New(
						slider.Value(75),
						slider.Min(0),
						slider.Max(100),
						slider.PainterOpt(ps.slider),
					),
				).Gap(8).Padding(12),
			},
		},
		tabview.PainterOpt(ps.tabview),
		tabview.OnSelect(func(idx int) {
			fmt.Printf("tab selected: %d\n", idx)
		}),
	)
	tvBox := primitives.Box(tv).
		Height(200).
		Rounded(6).
		Background(ps.cardBg()).
		BorderStyle(1, ps.cardBorder())

	// SplitView.
	sv := splitview.New(
		splitview.First(
			primitives.VBox(
				primitives.Text("Left Panel").FontSize(14).Bold().Color(ps.textColor()),
				primitives.Text("Drag the divider to resize.").FontSize(12).Color(ps.subtextColor()),
			).Gap(4).Padding(12),
		),
		splitview.Second(
			primitives.VBox(
				primitives.Text("Right Panel").FontSize(14).Bold().Color(ps.textColor()),
				primitives.Text("Double-click to collapse.").FontSize(12).Color(ps.subtextColor()),
			).Gap(4).Padding(12),
		),
		splitview.InitialRatio(0.4),
		splitview.CollapsibleOpt(true),
		splitview.PainterOpt(ps.splitview),
	)
	svBox := primitives.Box(sv).
		Height(120).
		Rounded(6).
		Background(ps.cardBg()).
		BorderStyle(1, ps.cardBorder())

	// Nested collapsible.
	nested := collapsible.New(
		collapsible.Title("Nested Collapsible"),
		collapsible.Content(
			primitives.VBox(
				primitives.Text("This section is inside another collapsible.").FontSize(13).Color(ps.textColor()),
				primitives.Text("Collapsible supports animated expand/collapse.").FontSize(12).Color(ps.subtextColor()),
			).Gap(4).Padding(12),
		),
		collapsible.Expanded(false),
		collapsible.Animated(true),
		collapsible.HeaderHeight(36),
		collapsible.PainterOpt(ps.collapsible),
	)

	// Dialog (shown inline as a widget for demo, since dialogs are typically overlays).
	dlg := dialog.New(
		dialog.Title("Confirm Action"),
		dialog.Content(
			primitives.Text("Are you sure you want to proceed?").FontSize(14).Color(ps.textColor()),
		),
		dialog.Actions(
			dialog.Action{Label: "Cancel", OnClick: func() { fmt.Println("dialog: cancel") }},
			dialog.Action{Label: "Confirm", OnClick: func() { fmt.Println("dialog: confirm") }},
		),
		dialog.MaxWidth(400),
		dialog.PainterOpt(ps.dialog),
	)
	dlgBox := primitives.Box(dlg).
		Height(160).
		Rounded(6).
		Background(ps.cardBg()).
		BorderStyle(1, ps.cardBorder())

	content := primitives.VBox(
		sectionLabel("TabView (3 tabs)", ps),
		tvBox,
		sectionLabel("SplitView (draggable divider)", ps),
		svBox,
		sectionLabel("Collapsible", ps),
		nested,
		sectionLabel("Dialog", ps),
		dlgBox,
	).Gap(10).Padding(16).Background(ps.sectionBg()).Rounded(8)

	return collapsible.New(
		collapsible.Title("Containers"),
		collapsible.Content(content),
		collapsible.Expanded(true),
		collapsible.Animated(true),
		collapsible.HeaderHeight(40),
		collapsible.PainterOpt(ps.collapsible),
	)
}

// --- Section 5: Navigation ---

func buildNavigationSection(ps painterSet) *collapsible.Widget {
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
		toolbar.PainterOpt(ps.toolbar),
	)

	accentColor := widget.Hex(0x6750A4)
	if ps.isDark {
		accentColor = widget.Hex(0x3574F0)
	}

	// Breadcrumb-style text.
	breadcrumb := primitives.HBox(
		primitives.Text("Home").FontSize(13).Color(accentColor),
		primitives.Text("/").FontSize(13).Color(ps.subtextColor()),
		primitives.Text("Gallery").FontSize(13).Color(accentColor),
		primitives.Text("/").FontSize(13).Color(ps.subtextColor()),
		primitives.Text("Navigation").FontSize(13).Color(ps.textColor()),
	).Gap(6)

	content := primitives.VBox(
		sectionLabel("Toolbar", ps),
		tb,
		sectionLabel("Breadcrumb", ps),
		breadcrumb,
	).Gap(10).Padding(16).Background(ps.sectionBg()).Rounded(8)

	return collapsible.New(
		collapsible.Title("Navigation"),
		collapsible.Content(content),
		collapsible.Expanded(true),
		collapsible.Animated(true),
		collapsible.HeaderHeight(40),
		collapsible.PainterOpt(ps.collapsible),
	)
}

// --- Helpers ---

// sectionLabel returns a small muted label for grouping widgets within a section.
func sectionLabel(text string, ps painterSet) *primitives.TextWidget {
	return primitives.Text(text).
		FontSize(12).
		Bold().
		Color(ps.subtextColor())
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
