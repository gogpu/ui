package render

import (
	"testing"

	"github.com/gogpu/gg"
	"github.com/gogpu/ui/geometry"
	"github.com/gogpu/ui/widget"
)

func newTestCanvas(width, height int) *Canvas {
	dc := gg.NewContext(width, height)
	return NewCanvas(dc, width, height)
}

func TestNewCanvas(t *testing.T) {
	canvas := newTestCanvas(800, 600)

	if canvas.Width() != 800 {
		t.Errorf("Width() = %v, want 800", canvas.Width())
	}
	if canvas.Height() != 600 {
		t.Errorf("Height() = %v, want 600", canvas.Height())
	}
	if canvas.Context() == nil {
		t.Error("Context() should not be nil")
	}
	if canvas.ClipDepth() != 0 {
		t.Errorf("ClipDepth() = %v, want 0", canvas.ClipDepth())
	}
	if canvas.TransformDepth() != 0 {
		t.Errorf("TransformDepth() = %v, want 0", canvas.TransformDepth())
	}
}

func TestCanvas_Clear(t *testing.T) {
	canvas := newTestCanvas(100, 100)

	// Should not panic
	canvas.Clear(widget.ColorRed)
	canvas.Clear(widget.ColorTransparent)
	canvas.Clear(widget.ColorWhite)
}

func TestCanvas_DrawRect(t *testing.T) {
	canvas := newTestCanvas(100, 100)

	// Basic rectangle draw - should not panic
	rect := geometry.NewRect(10, 10, 50, 50)
	canvas.DrawRect(rect, widget.ColorRed)

	// Edge cases
	canvas.DrawRect(geometry.NewRect(0, 0, 100, 100), widget.ColorBlue)     // Full size
	canvas.DrawRect(geometry.NewRect(-10, -10, 20, 20), widget.ColorGreen)  // Partially visible
	canvas.DrawRect(geometry.NewRect(200, 200, 50, 50), widget.ColorYellow) // Outside bounds (should skip)
}

func TestCanvas_StrokeRect(t *testing.T) {
	canvas := newTestCanvas(100, 100)

	rect := geometry.NewRect(10, 10, 50, 50)
	canvas.StrokeRect(rect, widget.ColorBlack, 2.0)

	// Edge cases
	canvas.StrokeRect(geometry.NewRect(0, 0, 100, 100), widget.ColorBlue, 1.0)
	canvas.StrokeRect(geometry.NewRect(200, 200, 50, 50), widget.ColorYellow, 1.0) // Outside
}

func TestCanvas_DrawRoundRect(t *testing.T) {
	canvas := newTestCanvas(100, 100)

	rect := geometry.NewRect(10, 10, 50, 50)
	canvas.DrawRoundRect(rect, widget.ColorRed, 5.0)

	// Various radii
	canvas.DrawRoundRect(rect, widget.ColorBlue, 0)    // No rounding
	canvas.DrawRoundRect(rect, widget.ColorGreen, 10)  // Medium rounding
	canvas.DrawRoundRect(rect, widget.ColorYellow, 25) // Maximum rounding (circle-ish)
}

func TestCanvas_StrokeRoundRect(t *testing.T) {
	canvas := newTestCanvas(100, 100)

	rect := geometry.NewRect(10, 10, 50, 50)
	canvas.StrokeRoundRect(rect, widget.ColorBlack, 5.0, 2.0)
}

func TestCanvas_DrawCircle(t *testing.T) {
	canvas := newTestCanvas(100, 100)

	center := geometry.Pt(50, 50)
	canvas.DrawCircle(center, 20, widget.ColorRed)

	// Edge cases
	canvas.DrawCircle(geometry.Pt(0, 0), 10, widget.ColorBlue)      // Partially visible
	canvas.DrawCircle(geometry.Pt(200, 200), 10, widget.ColorGreen) // Outside
}

func TestCanvas_StrokeCircle(t *testing.T) {
	canvas := newTestCanvas(100, 100)

	center := geometry.Pt(50, 50)
	canvas.StrokeCircle(center, 20, widget.ColorBlack, 2.0)
}

func TestCanvas_DrawLine(t *testing.T) {
	canvas := newTestCanvas(100, 100)

	from := geometry.Pt(10, 10)
	to := geometry.Pt(90, 90)
	canvas.DrawLine(from, to, widget.ColorBlack, 2.0)

	// Horizontal and vertical lines
	canvas.DrawLine(geometry.Pt(10, 50), geometry.Pt(90, 50), widget.ColorRed, 1.0)
	canvas.DrawLine(geometry.Pt(50, 10), geometry.Pt(50, 90), widget.ColorBlue, 1.0)
}

func TestCanvas_ClipStack(t *testing.T) {
	canvas := newTestCanvas(100, 100)

	// Initial state
	if canvas.ClipDepth() != 0 {
		t.Errorf("Initial ClipDepth() = %v, want 0", canvas.ClipDepth())
	}

	initialBounds := canvas.ClipBounds()
	if initialBounds.Width() != 100 || initialBounds.Height() != 100 {
		t.Errorf("Initial ClipBounds() = %v, want full canvas", initialBounds)
	}

	// Push first clip
	canvas.PushClip(geometry.NewRect(10, 10, 80, 80))
	if canvas.ClipDepth() != 1 {
		t.Errorf("After first PushClip, ClipDepth() = %v, want 1", canvas.ClipDepth())
	}

	bounds1 := canvas.ClipBounds()
	if bounds1.Min.X != 10 || bounds1.Min.Y != 10 {
		t.Errorf("After first PushClip, ClipBounds().Min = (%v, %v), want (10, 10)",
			bounds1.Min.X, bounds1.Min.Y)
	}

	// Push second clip (intersection)
	canvas.PushClip(geometry.NewRect(20, 20, 40, 40))
	if canvas.ClipDepth() != 2 {
		t.Errorf("After second PushClip, ClipDepth() = %v, want 2", canvas.ClipDepth())
	}

	bounds2 := canvas.ClipBounds()
	if bounds2.Min.X != 20 || bounds2.Min.Y != 20 {
		t.Errorf("After second PushClip, ClipBounds().Min = (%v, %v), want (20, 20)",
			bounds2.Min.X, bounds2.Min.Y)
	}

	// Pop second clip
	canvas.PopClip()
	if canvas.ClipDepth() != 1 {
		t.Errorf("After first PopClip, ClipDepth() = %v, want 1", canvas.ClipDepth())
	}

	bounds3 := canvas.ClipBounds()
	if bounds3.Min.X != 10 || bounds3.Min.Y != 10 {
		t.Errorf("After first PopClip, ClipBounds().Min = (%v, %v), want (10, 10)",
			bounds3.Min.X, bounds3.Min.Y)
	}

	// Pop first clip
	canvas.PopClip()
	if canvas.ClipDepth() != 0 {
		t.Errorf("After second PopClip, ClipDepth() = %v, want 0", canvas.ClipDepth())
	}

	// Pop on empty stack should be no-op
	canvas.PopClip()
	if canvas.ClipDepth() != 0 {
		t.Errorf("After extra PopClip, ClipDepth() = %v, want 0", canvas.ClipDepth())
	}
}

func TestCanvas_TransformStack(t *testing.T) {
	canvas := newTestCanvas(100, 100)

	// Initial state
	if canvas.TransformDepth() != 0 {
		t.Errorf("Initial TransformDepth() = %v, want 0", canvas.TransformDepth())
	}

	initialOffset := canvas.TransformOffset()
	if initialOffset.X != 0 || initialOffset.Y != 0 {
		t.Errorf("Initial TransformOffset() = %v, want (0, 0)", initialOffset)
	}

	// Push first transform
	canvas.PushTransform(geometry.Pt(10, 20))
	if canvas.TransformDepth() != 1 {
		t.Errorf("After first PushTransform, TransformDepth() = %v, want 1", canvas.TransformDepth())
	}

	offset1 := canvas.TransformOffset()
	if offset1.X != 10 || offset1.Y != 20 {
		t.Errorf("After first PushTransform, TransformOffset() = (%v, %v), want (10, 20)",
			offset1.X, offset1.Y)
	}

	// Push second transform (cumulative)
	canvas.PushTransform(geometry.Pt(5, 5))
	if canvas.TransformDepth() != 2 {
		t.Errorf("After second PushTransform, TransformDepth() = %v, want 2", canvas.TransformDepth())
	}

	offset2 := canvas.TransformOffset()
	if offset2.X != 15 || offset2.Y != 25 {
		t.Errorf("After second PushTransform, TransformOffset() = (%v, %v), want (15, 25)",
			offset2.X, offset2.Y)
	}

	// Pop second transform
	canvas.PopTransform()
	if canvas.TransformDepth() != 1 {
		t.Errorf("After first PopTransform, TransformDepth() = %v, want 1", canvas.TransformDepth())
	}

	offset3 := canvas.TransformOffset()
	if offset3.X != 10 || offset3.Y != 20 {
		t.Errorf("After first PopTransform, TransformOffset() = (%v, %v), want (10, 20)",
			offset3.X, offset3.Y)
	}

	// Pop first transform
	canvas.PopTransform()
	if canvas.TransformDepth() != 0 {
		t.Errorf("After second PopTransform, TransformDepth() = %v, want 0", canvas.TransformDepth())
	}

	// Pop on empty stack should be no-op
	canvas.PopTransform()
	if canvas.TransformDepth() != 0 {
		t.Errorf("After extra PopTransform, TransformDepth() = %v, want 0", canvas.TransformDepth())
	}
}

func TestCanvas_Reset(t *testing.T) {
	canvas := newTestCanvas(100, 100)

	// Add some state
	canvas.PushClip(geometry.NewRect(10, 10, 80, 80))
	canvas.PushClip(geometry.NewRect(20, 20, 60, 60))
	canvas.PushTransform(geometry.Pt(10, 10))
	canvas.PushTransform(geometry.Pt(5, 5))

	if canvas.ClipDepth() != 2 || canvas.TransformDepth() != 2 {
		t.Error("State not properly added before reset")
	}

	// Reset
	canvas.Reset()

	if canvas.ClipDepth() != 0 {
		t.Errorf("After Reset, ClipDepth() = %v, want 0", canvas.ClipDepth())
	}
	if canvas.TransformDepth() != 0 {
		t.Errorf("After Reset, TransformDepth() = %v, want 0", canvas.TransformDepth())
	}

	offset := canvas.TransformOffset()
	if offset.X != 0 || offset.Y != 0 {
		t.Errorf("After Reset, TransformOffset() = (%v, %v), want (0, 0)", offset.X, offset.Y)
	}

	bounds := canvas.ClipBounds()
	if bounds.Width() != 100 || bounds.Height() != 100 {
		t.Errorf("After Reset, ClipBounds() = %v, want full canvas", bounds)
	}
}

func TestCanvas_TransformAppliedToDrawing(t *testing.T) {
	canvas := newTestCanvas(100, 100)

	// Push a transform and draw - should not panic
	canvas.PushTransform(geometry.Pt(20, 20))
	canvas.DrawRect(geometry.NewRect(0, 0, 30, 30), widget.ColorRed)
	canvas.PopTransform()
}

func TestCanvas_ClipAppliedToDrawing(t *testing.T) {
	canvas := newTestCanvas(100, 100)

	// Push a clip and draw - should not panic
	canvas.PushClip(geometry.NewRect(20, 20, 60, 60))
	canvas.DrawRect(geometry.NewRect(0, 0, 100, 100), widget.ColorRed)
	canvas.PopClip()
}

func TestCanvas_ImplementsWidgetCanvas(t *testing.T) {
	canvas := newTestCanvas(100, 100)

	// Verify the interface is implemented
	var _ widget.Canvas = canvas
}

func TestCanvas_isVisible(t *testing.T) {
	canvas := newTestCanvas(100, 100)

	tests := []struct {
		name    string
		rect    geometry.Rect
		visible bool
	}{
		{"inside", geometry.NewRect(10, 10, 20, 20), true},
		{"touching edge", geometry.NewRect(0, 0, 10, 10), true},
		{"overlapping edge", geometry.NewRect(-5, -5, 20, 20), true},
		{"outside left", geometry.NewRect(-20, 50, 10, 10), false},
		{"outside right", geometry.NewRect(110, 50, 10, 10), false},
		{"outside top", geometry.NewRect(50, -20, 10, 10), false},
		{"outside bottom", geometry.NewRect(50, 110, 10, 10), false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := canvas.isVisible(tt.rect)
			if got != tt.visible {
				t.Errorf("isVisible(%v) = %v, want %v", tt.rect, got, tt.visible)
			}
		})
	}
}

func TestCanvas_applyTransform(t *testing.T) {
	canvas := newTestCanvas(100, 100)

	// No transform
	rect := geometry.NewRect(10, 10, 20, 20)
	transformed := canvas.applyTransform(rect)
	if transformed.Min.X != 10 || transformed.Min.Y != 10 {
		t.Errorf("No transform: got (%v, %v), want (10, 10)",
			transformed.Min.X, transformed.Min.Y)
	}

	// With transform
	canvas.PushTransform(geometry.Pt(5, 10))
	transformed = canvas.applyTransform(rect)
	if transformed.Min.X != 15 || transformed.Min.Y != 20 {
		t.Errorf("With transform: got (%v, %v), want (15, 20)",
			transformed.Min.X, transformed.Min.Y)
	}
}

func TestCanvas_applyTransformPoint(t *testing.T) {
	canvas := newTestCanvas(100, 100)

	// No transform
	p := geometry.Pt(10, 10)
	transformed := canvas.applyTransformPoint(p)
	if transformed.X != 10 || transformed.Y != 10 {
		t.Errorf("No transform: got (%v, %v), want (10, 10)",
			transformed.X, transformed.Y)
	}

	// With transform
	canvas.PushTransform(geometry.Pt(5, 10))
	transformed = canvas.applyTransformPoint(p)
	if transformed.X != 15 || transformed.Y != 20 {
		t.Errorf("With transform: got (%v, %v), want (15, 20)",
			transformed.X, transformed.Y)
	}
}

// Benchmarks

func BenchmarkCanvas_DrawRect(b *testing.B) {
	canvas := newTestCanvas(800, 600)
	rect := geometry.NewRect(100, 100, 200, 150)
	color := widget.ColorRed

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		canvas.DrawRect(rect, color)
	}
}

func BenchmarkCanvas_DrawRoundRect(b *testing.B) {
	canvas := newTestCanvas(800, 600)
	rect := geometry.NewRect(100, 100, 200, 150)
	color := widget.ColorRed

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		canvas.DrawRoundRect(rect, color, 10)
	}
}

func BenchmarkCanvas_DrawCircle(b *testing.B) {
	canvas := newTestCanvas(800, 600)
	center := geometry.Pt(400, 300)
	color := widget.ColorRed

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		canvas.DrawCircle(center, 50, color)
	}
}

func BenchmarkCanvas_PushPopClip(b *testing.B) {
	canvas := newTestCanvas(800, 600)
	rect := geometry.NewRect(100, 100, 600, 400)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		canvas.PushClip(rect)
		canvas.PopClip()
	}
}

func BenchmarkCanvas_PushPopTransform(b *testing.B) {
	canvas := newTestCanvas(800, 600)
	offset := geometry.Pt(10, 10)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		canvas.PushTransform(offset)
		canvas.PopTransform()
	}
}

func BenchmarkCanvas_Clear(b *testing.B) {
	canvas := newTestCanvas(800, 600)
	color := widget.ColorWhite

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		canvas.Clear(color)
	}
}
