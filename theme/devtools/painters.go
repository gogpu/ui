package devtools

// Painters provides all DevTools painters for convenient widget construction.
//
// Instead of creating each painter individually:
//
//	btn := button.New(button.PainterOpt(devtools.ButtonPainter{Theme: dt}))
//	cb  := checkbox.New(checkbox.PainterOpt(devtools.CheckboxPainter{Theme: dt}))
//
// Use NewPainters to create all painters at once:
//
//	p := devtools.NewPainters(dt)
//	btn := button.New(button.PainterOpt(p.Button))
//	cb  := checkbox.New(checkbox.PainterOpt(p.Checkbox))
type Painters struct {
	Badge       BadgePainter
	Button      ButtonPainter
	Checkbox    CheckboxPainter
	Chip        ChipPainter
	Radio       RadioPainter
	TextField   TextFieldPainter
	Dropdown    DropdownPainter
	Slider      SliderPainter
	Dialog      DialogPainter
	Scrollbar   ScrollbarPainter
	TabView     TabViewPainter
	TreeView    TreeViewPainter
	DataTable   DataTablePainter
	Toolbar     ToolbarPainter
	Menu        MenuPainter
	Collapsible CollapsiblePainter
	Progress    ProgressPainter
	SplitView   SplitViewPainter
	Docking     DockingPainter
	Popover     PopoverPainter
	LineChart   LineChartPainter
	ListView    ListViewPainter
	TitleBar    TitleBarPainter
	Stripe      StripePainter
}

// NewPainters returns all 24 DevTools painters initialized with the given theme.
//
// This is a convenience function that avoids repetitive Theme field assignment.
// All painters share the same *Theme pointer, so updating the theme via pointer
// mutation (e.g. *dt = *devtools.NewDarkTheme()) updates all painters at once.
func NewPainters(t *Theme) Painters {
	return Painters{
		Badge:       BadgePainter{Theme: t},
		Button:      ButtonPainter{Theme: t},
		Checkbox:    CheckboxPainter{Theme: t},
		Chip:        ChipPainter{Theme: t},
		Radio:       RadioPainter{Theme: t},
		TextField:   TextFieldPainter{Theme: t},
		Dropdown:    DropdownPainter{Theme: t},
		Slider:      SliderPainter{Theme: t},
		Dialog:      DialogPainter{Theme: t},
		Scrollbar:   ScrollbarPainter{Theme: t},
		TabView:     TabViewPainter{Theme: t},
		TreeView:    TreeViewPainter{Theme: t},
		DataTable:   DataTablePainter{Theme: t},
		Toolbar:     ToolbarPainter{Theme: t},
		Menu:        MenuPainter{Theme: t},
		Collapsible: CollapsiblePainter{Theme: t},
		Progress:    ProgressPainter{Theme: t},
		SplitView:   SplitViewPainter{Theme: t},
		Docking:     DockingPainter{Theme: t},
		Popover:     PopoverPainter{Theme: t},
		LineChart:   LineChartPainter{Theme: t},
		ListView:    ListViewPainter{Theme: t},
		TitleBar:    TitleBarPainter{Theme: t},
		Stripe:      StripePainter{Theme: t},
	}
}
