package primitives_test

import (
	"testing"

	"github.com/gogpu/ui/a11y"
	"github.com/gogpu/ui/event"
	"github.com/gogpu/ui/geometry"
	"github.com/gogpu/ui/primitives"
	"github.com/gogpu/ui/widget"
)

// testImageSource is a simple ImageSource for testing.
type testImageSource struct {
	width, height float32
}

func (s *testImageSource) Bounds() [2]float32 {
	return [2]float32{s.width, s.height}
}

// --- Image construction ---

func TestImageCreate(t *testing.T) {
	src := &testImageSource{width: 200, height: 100}
	img := primitives.Image(src)
	if img.Source() != src {
		t.Error("source not set")
	}
}

func TestImageNilSource(t *testing.T) {
	img := primitives.Image(nil)
	if img.Source() != nil {
		t.Error("nil source should remain nil")
	}
}

func TestImageVisibleAndEnabled(t *testing.T) {
	img := primitives.Image(nil)
	if !img.IsVisible() {
		t.Error("image should be visible by default")
	}
	if !img.IsEnabled() {
		t.Error("image should be enabled by default")
	}
}

func TestImageDefaultFit(t *testing.T) {
	img := primitives.Image(nil)
	if img.Style().Fit != primitives.ImageFitContain {
		t.Errorf("expected default fit Contain, got %s", img.Style().Fit)
	}
}

// --- Fluent style methods ---

func TestImageSize(t *testing.T) {
	img := primitives.Image(nil).Size(300, 200)
	style := img.Style()
	if style.ExplicitWidth != 300 || style.ExplicitHeight != 200 {
		t.Errorf("expected 300x200, got %fx%f", style.ExplicitWidth, style.ExplicitHeight)
	}
}

func TestImageFitModes(t *testing.T) {
	tests := []struct {
		name string
		fn   func(*primitives.ImageWidget) *primitives.ImageWidget
		want primitives.ImageFit
	}{
		{"Contain", (*primitives.ImageWidget).Contain, primitives.ImageFitContain},
		{"Cover", (*primitives.ImageWidget).Cover, primitives.ImageFitCover},
		{"Fill", (*primitives.ImageWidget).Fill, primitives.ImageFitFill},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			img := tt.fn(primitives.Image(nil))
			if img.Style().Fit != tt.want {
				t.Errorf("expected %s, got %s", tt.want, img.Style().Fit)
			}
		})
	}
}

func TestImageFitGeneric(t *testing.T) {
	img := primitives.Image(nil).Fit(primitives.ImageFitNone)
	if img.Style().Fit != primitives.ImageFitNone {
		t.Errorf("expected None, got %s", img.Style().Fit)
	}
}

func TestImageRounded(t *testing.T) {
	img := primitives.Image(nil).Rounded(16)
	if img.Style().Radius != 16 {
		t.Errorf("expected radius 16, got %f", img.Style().Radius)
	}
}

func TestImageAlt(t *testing.T) {
	img := primitives.Image(nil).Alt("A beautiful sunset")
	if img.AltText() != "A beautiful sunset" {
		t.Errorf("expected alt text, got %q", img.AltText())
	}
}

func TestImageFluentChaining(t *testing.T) {
	src := &testImageSource{width: 200, height: 100}
	img := primitives.Image(src).
		Size(300, 200).
		Cover().
		Rounded(8).
		Alt("Photo")

	style := img.Style()
	if style.ExplicitWidth != 300 {
		t.Error("width not chained")
	}
	if style.Fit != primitives.ImageFitCover {
		t.Error("fit not chained")
	}
	if style.Radius != 8 {
		t.Error("radius not chained")
	}
	if img.AltText() != "Photo" {
		t.Error("alt not chained")
	}
}

// --- Layout ---

func TestImageLayoutNilSource(t *testing.T) {
	img := primitives.Image(nil)
	ctx := widget.NewContext()
	size := img.Layout(ctx, geometry.Loose(geometry.Sz(300, 300)))

	// Nil source = zero natural size, constrained to 0
	if size.Width != 0 || size.Height != 0 {
		t.Errorf("nil source should have zero size, got %s", size)
	}
}

func TestImageLayoutNaturalSize(t *testing.T) {
	src := &testImageSource{width: 200, height: 100}
	img := primitives.Image(src).Contain()
	ctx := widget.NewContext()
	size := img.Layout(ctx, geometry.Loose(geometry.Sz(300, 300)))

	// Should use natural size (fits within 300x300)
	if size.Width != 200 || size.Height != 100 {
		t.Errorf("expected natural 200x100, got %s", size)
	}
}

func TestImageLayoutExplicitSize(t *testing.T) {
	src := &testImageSource{width: 200, height: 100}
	img := primitives.Image(src).Size(150, 75)
	ctx := widget.NewContext()
	size := img.Layout(ctx, geometry.Loose(geometry.Sz(300, 300)))

	// Should use explicit size
	if size.Width != 150 || size.Height != 75 {
		t.Errorf("expected explicit 150x75, got %s", size)
	}
}

func TestImageLayoutContainScalesDown(t *testing.T) {
	src := &testImageSource{width: 400, height: 200}
	img := primitives.Image(src).Contain()
	ctx := widget.NewContext()
	size := img.Layout(ctx, geometry.Loose(geometry.Sz(200, 200)))

	// 400x200 scaled to fit 200x200 = 200x100 (maintaining 2:1 aspect)
	if size.Width != 200 {
		t.Errorf("expected width 200, got %f", size.Width)
	}
	if size.Height != 100 {
		t.Errorf("expected height 100, got %f", size.Height)
	}
}

func TestImageLayoutCover(t *testing.T) {
	src := &testImageSource{width: 400, height: 200}
	img := primitives.Image(src).Cover()
	ctx := widget.NewContext()
	size := img.Layout(ctx, geometry.Loose(geometry.Sz(200, 200)))

	// Cover: scale to fill 200x200 from 400x200 (2:1 ratio)
	// FillIn scales by max(200/400, 200/200) = max(0.5, 1.0) = 1.0
	// Result 400x200, then clamped to 200x200
	if size.Width != 200 || size.Height != 200 {
		t.Errorf("cover should fill container: got %s", size)
	}
}

func TestImageLayoutFill(t *testing.T) {
	src := &testImageSource{width: 100, height: 50}
	img := primitives.Image(src).Fill()
	ctx := widget.NewContext()
	size := img.Layout(ctx, geometry.Loose(geometry.Sz(200, 150)))

	// Fill stretches to available space
	if size.Width != 200 || size.Height != 150 {
		t.Errorf("fill should stretch: got %s", size)
	}
}

func TestImageLayoutNone(t *testing.T) {
	src := &testImageSource{width: 100, height: 50}
	img := primitives.Image(src).Fit(primitives.ImageFitNone)
	ctx := widget.NewContext()
	size := img.Layout(ctx, geometry.Loose(geometry.Sz(200, 200)))

	// None: natural size
	if size.Width != 100 || size.Height != 50 {
		t.Errorf("none should use natural size: got %s", size)
	}
}

func TestImageLayoutConstrainedByParent(t *testing.T) {
	src := &testImageSource{width: 500, height: 300}
	img := primitives.Image(src).Fit(primitives.ImageFitNone)
	ctx := widget.NewContext()
	size := img.Layout(ctx, geometry.Loose(geometry.Sz(200, 200)))

	// None but constrained to 200x200
	if size.Width != 200 || size.Height != 200 {
		t.Errorf("should be constrained to 200x200, got %s", size)
	}
}

func TestImageLayoutSetsBounds(t *testing.T) {
	src := &testImageSource{width: 100, height: 50}
	img := primitives.Image(src)
	ctx := widget.NewContext()
	size := img.Layout(ctx, geometry.Loose(geometry.Sz(300, 300)))

	bounds := img.Bounds()
	if bounds.Width() != size.Width || bounds.Height() != size.Height {
		t.Errorf("bounds should match layout: bounds=%s, size=%s", bounds.Size(), size)
	}
}

// --- Draw ---

func TestImageDrawNoPanicNilSource(t *testing.T) {
	img := primitives.Image(nil)
	ctx := widget.NewContext()
	canvas := &mockCanvas{}
	_ = img.Layout(ctx, geometry.Loose(geometry.Sz(100, 100)))
	img.Draw(ctx, canvas) // Should not panic
}

func TestImageDrawRendersPlaceholder(t *testing.T) {
	src := &testImageSource{width: 100, height: 50}
	img := primitives.Image(src)
	ctx := widget.NewContext()
	canvas := &mockCanvas{}
	_ = img.Layout(ctx, geometry.Loose(geometry.Sz(200, 200)))
	img.Draw(ctx, canvas)

	// Should draw placeholder rect + cross lines
	if canvas.drawRectCount == 0 {
		t.Error("expected placeholder rect")
	}
	if canvas.drawLineCount != 2 {
		t.Errorf("expected 2 cross lines, got %d", canvas.drawLineCount)
	}
}

func TestImageDrawRounded(t *testing.T) {
	src := &testImageSource{width: 100, height: 50}
	img := primitives.Image(src).Rounded(8)
	ctx := widget.NewContext()
	canvas := &mockCanvas{}
	_ = img.Layout(ctx, geometry.Loose(geometry.Sz(200, 200)))
	img.Draw(ctx, canvas)

	if canvas.drawRoundRectCount == 0 {
		t.Error("expected DrawRoundRect for rounded image")
	}
}

func TestImageDrawInvisible(t *testing.T) {
	src := &testImageSource{width: 100, height: 50}
	img := primitives.Image(src)
	img.SetVisible(false)
	ctx := widget.NewContext()
	canvas := &mockCanvas{}
	_ = img.Layout(ctx, geometry.Loose(geometry.Sz(200, 200)))
	img.Draw(ctx, canvas)

	if canvas.drawRectCount != 0 || canvas.drawLineCount != 0 {
		t.Error("invisible image should not draw")
	}
}

// --- Event ---

func TestImageEventNotConsumed(t *testing.T) {
	img := primitives.Image(nil)
	ctx := widget.NewContext()
	e := &event.Base{}
	if img.Event(ctx, e) {
		t.Error("image should not consume events")
	}
}

// --- Children ---

func TestImageChildrenNil(t *testing.T) {
	img := primitives.Image(nil)
	if img.Children() != nil {
		t.Error("image should have no children")
	}
}

// --- Accessibility ---

func TestImageAccessibilityRole(t *testing.T) {
	img := primitives.Image(nil)
	if img.AccessibilityRole() != a11y.RoleImage {
		t.Errorf("expected RoleImage, got %s", img.AccessibilityRole())
	}
}

func TestImageAccessibilityLabelDefault(t *testing.T) {
	img := primitives.Image(nil)
	if img.AccessibilityLabel() != "" {
		t.Errorf("expected empty label, got %q", img.AccessibilityLabel())
	}
}

func TestImageAccessibilityLabelAlt(t *testing.T) {
	img := primitives.Image(nil).Alt("Sunset photo")
	if img.AccessibilityLabel() != "Sunset photo" {
		t.Errorf("expected 'Sunset photo', got %q", img.AccessibilityLabel())
	}
}

func TestImageAccessibilityState(t *testing.T) {
	img := primitives.Image(nil)
	state := img.AccessibilityState()
	if state.Hidden || state.Disabled {
		t.Error("default state should be visible and enabled")
	}

	img.SetVisible(false)
	state = img.AccessibilityState()
	if !state.Hidden {
		t.Error("invisible image should report Hidden=true")
	}
}

func TestImageAccessibilityActions(t *testing.T) {
	img := primitives.Image(nil)
	if img.AccessibilityActions() != nil {
		t.Error("image should have no actions")
	}
}

func TestImageAccessibilityHint(t *testing.T) {
	img := primitives.Image(nil)
	if img.AccessibilityHint() != "" {
		t.Error("image should have no hint")
	}
}

func TestImageAccessibilityValue(t *testing.T) {
	img := primitives.Image(nil)
	if img.AccessibilityValue() != "" {
		t.Error("image should have no value")
	}
}

// --- Style enums ---

func TestImageFitString(t *testing.T) {
	tests := []struct {
		fit  primitives.ImageFit
		want string
	}{
		{primitives.ImageFitContain, "Contain"},
		{primitives.ImageFitCover, "Cover"},
		{primitives.ImageFitFill, "Fill"},
		{primitives.ImageFitNone, "None"},
		{primitives.ImageFit(99), "Unknown"},
	}
	for _, tt := range tests {
		t.Run(tt.want, func(t *testing.T) {
			if got := tt.fit.String(); got != tt.want {
				t.Errorf("got %q, want %q", got, tt.want)
			}
		})
	}
}

// --- Style types ---

func TestBorderIsZero(t *testing.T) {
	zero := primitives.Border{}
	if !zero.IsZero() {
		t.Error("zero border should be zero")
	}

	visible := primitives.Border{Width: 1, Color: widget.ColorBlack}
	if visible.IsZero() {
		t.Error("visible border should not be zero")
	}

	transparent := primitives.Border{Width: 1, Color: widget.ColorTransparent}
	if !transparent.IsZero() {
		t.Error("transparent border should be zero")
	}
}

func TestShadowIsZero(t *testing.T) {
	zero := primitives.Shadow{}
	if !zero.IsZero() {
		t.Error("zero shadow should be zero")
	}

	visible := primitives.Shadow{Level: 3}
	if visible.IsZero() {
		t.Error("level 3 shadow should not be zero")
	}
}
