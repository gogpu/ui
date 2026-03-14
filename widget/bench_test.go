package widget

import (
	"testing"

	"github.com/gogpu/ui/event"
	"github.com/gogpu/ui/geometry"
)

// benchWidget is a minimal Widget for benchmarks.
type benchWidget struct {
	WidgetBase
}

func (w *benchWidget) Layout(_ Context, c geometry.Constraints) geometry.Size {
	return c.Constrain(geometry.Sz(100, 50))
}

func (w *benchWidget) Draw(_ Context, _ Canvas) {}

func (w *benchWidget) Event(_ Context, _ event.Event) bool { return false }

func newBenchWidget() *benchWidget {
	w := &benchWidget{}
	w.SetVisible(true)
	w.SetEnabled(true)
	return w
}

// Verify benchWidget implements Widget.
var _ Widget = (*benchWidget)(nil)

func BenchmarkWidgetBaseLayout(b *testing.B) {
	w := newBenchWidget()
	ctx := NewContext()
	constraints := geometry.BoxConstraints(0, 800, 0, 600)
	b.ReportAllocs()
	b.ResetTimer()
	for b.Loop() {
		w.Layout(ctx, constraints)
	}
}

func BenchmarkWidgetBaseBounds(b *testing.B) {
	w := newBenchWidget()
	w.SetBounds(geometry.NewRect(10, 20, 200, 100))
	b.ReportAllocs()
	b.ResetTimer()
	for b.Loop() {
		_ = w.Bounds()
	}
}

func BenchmarkWidgetBaseSetBounds(b *testing.B) {
	w := newBenchWidget()
	bounds := geometry.NewRect(10, 20, 200, 100)
	b.ReportAllocs()
	b.ResetTimer()
	for b.Loop() {
		w.SetBounds(bounds)
	}
}

func BenchmarkNeedsRedrawCheck(b *testing.B) {
	w := newBenchWidget()
	w.SetNeedsRedraw(true)
	b.ReportAllocs()
	b.ResetTimer()
	for b.Loop() {
		_ = w.NeedsRedraw()
	}
}

func BenchmarkWidgetContainsPoint(b *testing.B) {
	w := newBenchWidget()
	w.SetBounds(geometry.NewRect(10, 10, 200, 100))
	p := geometry.Pt(50, 50)
	b.ReportAllocs()
	b.ResetTimer()
	for b.Loop() {
		_ = w.ContainsPoint(p)
	}
}

// buildWidgetTree creates a flat tree with n children under a single root.
func buildWidgetTree(n int) *benchWidget {
	root := newBenchWidget()
	for i := 0; i < n; i++ {
		child := newBenchWidget()
		root.AddChild(child)
	}
	return root
}

// walkTree recursively visits all widgets in the tree.
func walkTree(w Widget) int {
	count := 1
	for _, child := range w.Children() {
		count += walkTree(child)
	}
	return count
}

func BenchmarkWidgetTreeWalk100(b *testing.B) {
	root := buildWidgetTree(100)
	b.ReportAllocs()
	b.ResetTimer()
	for b.Loop() {
		_ = walkTree(root)
	}
}

func BenchmarkWidgetTreeWalk1000(b *testing.B) {
	root := buildWidgetTree(1000)
	b.ReportAllocs()
	b.ResetTimer()
	for b.Loop() {
		_ = walkTree(root)
	}
}

func BenchmarkWidgetAddChild(b *testing.B) {
	root := newBenchWidget()
	children := make([]*benchWidget, b.N)
	for i := range children {
		children[i] = newBenchWidget()
	}
	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		root.AddChild(children[i])
	}
}
