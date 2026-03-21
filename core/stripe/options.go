package stripe

// Option configures a stripe widget during construction.
type Option func(*config)

// config holds the stripe's configuration, set at construction time via options.
type config struct {
	topItems    []Button
	bottomItems []Button
	activeID    string
	showLabels  bool
	width       float32
	painter     Painter
	onSelect    func(id string)
}

// TopItems sets the buttons displayed in the top group of the stripe.
// Top items are gravity-aligned to the top edge of the strip.
func TopItems(items ...Button) Option {
	return func(c *config) {
		c.topItems = items
	}
}

// BottomItems sets the buttons displayed in the bottom group of the stripe.
// Bottom items are gravity-aligned to the bottom edge of the strip.
func BottomItems(items ...Button) Option {
	return func(c *config) {
		c.bottomItems = items
	}
}

// ActiveID sets the ID of the currently active (selected) button.
// The active button receives distinct visual treatment from the painter.
func ActiveID(id string) Option {
	return func(c *config) {
		c.activeID = id
	}
}

// ShowLabels controls whether text labels are displayed below button icons.
// When true, buttons are taller to accommodate the label text and the strip
// defaults to a wider width (64px). When false, buttons show icons only and
// the strip defaults to 40px width.
func ShowLabels(show bool) Option {
	return func(c *config) {
		c.showLabels = show
	}
}

// Width sets the stripe width in logical pixels.
// Default is 64px with labels or 40px without labels.
func Width(w float32) Option {
	return func(c *config) {
		c.width = w
	}
}

// PainterOpt sets the painter used to render the stripe.
// If not set, [DefaultPainter] is used.
func PainterOpt(p Painter) Option {
	return func(c *config) {
		c.painter = p
	}
}

// OnSelect sets a callback invoked when any button is clicked.
// The callback receives the ID of the clicked button.
// This is called in addition to the button's own OnClick handler.
func OnSelect(fn func(id string)) Option {
	return func(c *config) {
		c.onSelect = fn
	}
}
