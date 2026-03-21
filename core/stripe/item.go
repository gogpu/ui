package stripe

import "github.com/gogpu/ui/icon"

// Button represents a single tool window button in the stripe.
//
// Each button has a unique ID for identification, a human-readable Label
// displayed below the icon (when labels are enabled), an Icon for the visual
// indicator, and an OnClick callback fired when the button is activated.
type Button struct {
	// ID uniquely identifies this button within the stripe.
	// Used to set and query the active button via [ActiveID].
	ID string

	// Label is the human-readable text displayed below the icon.
	// Also used for accessibility announcements.
	Label string

	// Icon is the vector icon data rendered in the button center.
	Icon icon.IconData

	// OnClick is the callback invoked when the button is activated
	// by mouse click. May be nil.
	OnClick func()
}
