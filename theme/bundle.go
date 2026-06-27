package theme

// Bundle provides a complete set of painters for all core widgets
// from a single design system.
//
// Each design system (Material 3, DevTools, Fluent, Cupertino) implements
// Bundle to offer one-call creation of all widget painters. This replaces
// ad-hoc painter-set structs in application code and enables design-system
// switching with a single variable swap.
//
// Painter values are typed as any to avoid import cycles between theme/
// and core/ packages. Consumer code type-asserts to the concrete painter
// interface defined by each widget:
//
//	bundle := material3.NewBundle(theme)
//	btn := button.New(button.PainterOpt(bundle.Painter("button").(button.Painter)))
//
// Or use the convenience map for bulk painter assignment:
//
//	for name, painter := range bundle.Painters() {
//	    registry.SetPainter(name, painter)
//	}
//
// Bundle implementations should use the widget package name (lowercase) as
// the painter key: "button", "checkbox", "radio", "textfield", etc.
type Bundle interface {
	// Name returns a human-readable name for the design system.
	//
	// Examples: "Material 3", "DevTools Dark", "Fluent Light", "Cupertino"
	Name() string

	// BaseTheme returns the underlying Theme for color and typography access.
	//
	// This allows consumers to read theme tokens (Colors.Primary, Typography,
	// Spacing) without importing the design-system-specific theme struct.
	BaseTheme() *Theme

	// Painter returns a single painter by widget name, or nil if the design
	// system does not provide a painter for that widget.
	//
	// Standard widget names (matching core/ package names):
	//   "badge", "button", "checkbox", "chip", "collapsible", "datatable",
	//   "dialog", "docking", "dropdown", "gridview", "linechart", "listview",
	//   "menu", "popover", "progress", "progressbar", "radio", "scrollview",
	//   "slider", "splitview", "stripe", "tabview", "textfield", "titlebar",
	//   "toolbar", "treeview"
	Painter(widget string) any

	// Painters returns all painters as a map from widget name to painter.
	//
	// The returned map is a snapshot -- mutations do not affect the bundle.
	Painters() map[string]any
}
