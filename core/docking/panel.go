package docking

import "github.com/gogpu/ui/widget"

// PanelOption configures a panel during construction.
type PanelOption func(*panelConfig)

// panelConfig holds the panel's configuration.
type panelConfig struct {
	title     string
	content   widget.Widget
	closeable bool
}

// PanelTitle sets the panel's display title shown in the tab header.
func PanelTitle(title string) PanelOption {
	return func(c *panelConfig) { c.title = title }
}

// PanelContent sets the panel's content widget.
func PanelContent(w widget.Widget) PanelOption {
	return func(c *panelConfig) { c.content = w }
}

// Closeable enables the close button on this panel.
func Closeable(v bool) PanelOption {
	return func(c *panelConfig) { c.closeable = v }
}

// Panel represents an individual dockable panel with a title, content widget,
// and optional close button.
//
// Panels are created with [NewPanel] and docked to a [Host] via [Host.Dock].
type Panel struct {
	cfg panelConfig
}

// NewPanel creates a new dockable panel with the given options.
//
//	p := docking.NewPanel(
//	    docking.PanelTitle("Explorer"),
//	    docking.PanelContent(explorerWidget),
//	    docking.Closeable(true),
//	)
func NewPanel(opts ...PanelOption) *Panel {
	p := &Panel{}
	for _, opt := range opts {
		opt(&p.cfg)
	}
	return p
}

// Title returns the panel's display title.
func (p *Panel) Title() string {
	return p.cfg.title
}

// Content returns the panel's content widget, or nil if no content is set.
func (p *Panel) Content() widget.Widget {
	return p.cfg.content
}

// IsCloseable reports whether the panel has a close button.
func (p *Panel) IsCloseable() bool {
	return p.cfg.closeable
}
