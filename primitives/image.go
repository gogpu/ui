package primitives

import (
	"github.com/gogpu/ui/a11y"
	"github.com/gogpu/ui/event"
	"github.com/gogpu/ui/geometry"
	"github.com/gogpu/ui/widget"
)

// ImageStyle holds all visual styling for an [ImageWidget].
type ImageStyle struct {
	// ExplicitWidth and ExplicitHeight override the natural image size.
	// Zero means use the natural dimension.
	ExplicitWidth  float32
	ExplicitHeight float32
	Fit            ImageFit
	Radius         float32
}

// ImageWidget displays a raster image with configurable fit mode, optional
// rounded corners, and accessibility alt text.
//
// ImageWidget implements [widget.Widget] and [a11y.Accessible].
//
// Create an ImageWidget with the [Image] constructor.
type ImageWidget struct {
	widget.WidgetBase

	style  ImageStyle
	source ImageSource
	alt    string
}

// Image creates a new image widget with the given source.
//
// The source provides the natural pixel dimensions of the image. If source
// is nil, the widget has zero natural size.
//
//	img := primitives.Image(mySource).Size(200, 150).Cover()
func Image(source ImageSource) *ImageWidget {
	w := &ImageWidget{
		source: source,
		style: ImageStyle{
			Fit: ImageFitContain,
		},
	}
	w.SetVisible(true)
	w.SetEnabled(true)
	return w
}

// --- Fluent style methods ---

// Size sets explicit width and height.
func (w *ImageWidget) Size(width, height float32) *ImageWidget {
	w.style.ExplicitWidth = width
	w.style.ExplicitHeight = height
	return w
}

// Fit sets the image fit mode.
func (w *ImageWidget) Fit(f ImageFit) *ImageWidget {
	w.style.Fit = f
	return w
}

// Cover is a shorthand for Fit(ImageFitCover).
func (w *ImageWidget) Cover() *ImageWidget {
	w.style.Fit = ImageFitCover
	return w
}

// Contain is a shorthand for Fit(ImageFitContain).
func (w *ImageWidget) Contain() *ImageWidget {
	w.style.Fit = ImageFitContain
	return w
}

// Fill is a shorthand for Fit(ImageFitFill).
func (w *ImageWidget) Fill() *ImageWidget {
	w.style.Fit = ImageFitFill
	return w
}

// Rounded sets the border radius for clipping.
func (w *ImageWidget) Rounded(r float32) *ImageWidget {
	w.style.Radius = r
	return w
}

// Alt sets the accessibility alt text.
func (w *ImageWidget) Alt(text string) *ImageWidget {
	w.alt = text
	return w
}

// Style returns the current image style (read-only snapshot).
func (w *ImageWidget) Style() ImageStyle {
	return w.style
}

// Source returns the image source.
func (w *ImageWidget) Source() ImageSource {
	return w.source
}

// AltText returns the accessibility alt text.
func (w *ImageWidget) AltText() string {
	return w.alt
}

// --- widget.Widget interface ---

// Layout determines the widget size based on the image source bounds, the
// explicit size overrides, and the fit mode.
func (w *ImageWidget) Layout(_ widget.Context, constraints geometry.Constraints) geometry.Size {
	natural := w.naturalSize()

	// Apply explicit size overrides
	desired := natural
	if w.style.ExplicitWidth > 0 {
		desired.Width = w.style.ExplicitWidth
	}
	if w.style.ExplicitHeight > 0 {
		desired.Height = w.style.ExplicitHeight
	}

	// For Contain and Cover, we need to maintain aspect ratio
	switch w.style.Fit {
	case ImageFitContain:
		if desired.Width > 0 && desired.Height > 0 {
			containerMax := constraints.BiggestFinite(desired.Width, desired.Height)
			desired = desired.FitIn(containerMax)
		}
	case ImageFitCover:
		if desired.Width > 0 && desired.Height > 0 {
			containerMax := constraints.BiggestFinite(desired.Width, desired.Height)
			desired = desired.FillIn(containerMax)
			// Clamp to container since excess is clipped
			desired = desired.Min(containerMax)
		}
	case ImageFitFill:
		// Fill stretches to available space
		if constraints.HasBoundedWidth() && w.style.ExplicitWidth <= 0 {
			desired.Width = constraints.MaxWidth
		}
		if constraints.HasBoundedHeight() && w.style.ExplicitHeight <= 0 {
			desired.Height = constraints.MaxHeight
		}
	case ImageFitNone:
		// Use natural size, will be clipped by container
	}

	resultSize := constraints.Constrain(desired)
	w.SetBounds(geometry.FromPointSize(w.Position(), resultSize))
	return resultSize
}

// Draw renders the image placeholder.
//
// The current implementation draws a light gray rectangle with a centered
// cross to indicate the image area. Full image rendering requires a Canvas
// implementation that supports DrawImage, which will be provided by the
// rendering backend.
func (w *ImageWidget) Draw(_ widget.Context, canvas widget.Canvas) {
	if !w.IsVisible() {
		return
	}

	bounds := w.Bounds()
	if bounds.IsEmpty() {
		return
	}

	// Placeholder: draw a light gray box with a cross
	placeholderColor := widget.RGBA(0.9, 0.9, 0.9, 1)
	crossColor := widget.RGBA(0.7, 0.7, 0.7, 1)

	if w.style.Radius > 0 {
		canvas.DrawRoundRect(bounds, placeholderColor, w.style.Radius)
	} else {
		canvas.DrawRect(bounds, placeholderColor)
	}

	// Draw cross
	center := bounds.Center()
	armLen := bounds.Width() / 4
	if h := bounds.Height() / 4; h < armLen {
		armLen = h
	}
	if armLen > 2 {
		canvas.DrawLine(
			geometry.Pt(center.X-armLen, center.Y-armLen),
			geometry.Pt(center.X+armLen, center.Y+armLen),
			crossColor, 1,
		)
		canvas.DrawLine(
			geometry.Pt(center.X+armLen, center.Y-armLen),
			geometry.Pt(center.X-armLen, center.Y+armLen),
			crossColor, 1,
		)
	}
}

// Event returns false. Image widgets do not consume events.
func (w *ImageWidget) Event(_ widget.Context, _ event.Event) bool {
	return false
}

// Children returns nil. Image is a leaf widget.
func (w *ImageWidget) Children() []widget.Widget {
	return nil
}

// --- a11y.Accessible interface ---

// AccessibilityRole returns [a11y.RoleImage].
func (w *ImageWidget) AccessibilityRole() a11y.Role {
	return a11y.RoleImage
}

// AccessibilityLabel returns the alt text.
func (w *ImageWidget) AccessibilityLabel() string {
	return w.alt
}

// AccessibilityHint returns an empty string.
func (w *ImageWidget) AccessibilityHint() string {
	return ""
}

// AccessibilityValue returns an empty string.
func (w *ImageWidget) AccessibilityValue() string {
	return ""
}

// AccessibilityState returns the default state.
func (w *ImageWidget) AccessibilityState() a11y.State {
	return a11y.State{
		Hidden: !w.IsVisible(),
	}
}

// AccessibilityActions returns nil. Images have no actions.
func (w *ImageWidget) AccessibilityActions() []a11y.Action {
	return nil
}

// --- internal ---

// naturalSize returns the natural size of the image from the source.
// Returns zero size if source is nil.
func (w *ImageWidget) naturalSize() geometry.Size {
	if w.source == nil {
		return geometry.Size{}
	}
	b := w.source.Bounds()
	return geometry.Sz(b[0], b[1])
}

// Compile-time interface checks.
var (
	_ widget.Widget   = (*ImageWidget)(nil)
	_ a11y.Accessible = (*ImageWidget)(nil)
)
