package layout

import (
	"testing"

	"github.com/gogpu/ui/geometry"
)

// benchLayoutable is a minimal Layoutable for benchmarks.
type benchLayoutable struct {
	preferredSize geometry.Size
}

func (b *benchLayoutable) Layout(_ geometry.Constraints) geometry.Size {
	return b.preferredSize
}

func (b *benchLayoutable) Children() []Layoutable { return nil }
func (b *benchLayoutable) ID() uint64             { return 0 }

// --- Flex layout benchmarks ---

func newFlexWithChildren(n int) *FlexContainer {
	flex := NewFlexContainer(Row, JustifyStart, AlignStart)
	child := &benchLayoutable{preferredSize: geometry.Sz(50, 30)}
	for i := 0; i < n; i++ {
		flex.AddChildWithFlex(child, 1, 1, 0)
	}
	return flex
}

func BenchmarkFlexLayout10(b *testing.B) {
	flex := newFlexWithChildren(10)
	constraints := geometry.BoxConstraints(0, 800, 0, 600)
	b.ReportAllocs()
	b.ResetTimer()
	for b.Loop() {
		flex.Layout(constraints)
	}
}

func BenchmarkFlexLayout100(b *testing.B) {
	flex := newFlexWithChildren(100)
	constraints := geometry.BoxConstraints(0, 1920, 0, 1080)
	b.ReportAllocs()
	b.ResetTimer()
	for b.Loop() {
		flex.Layout(constraints)
	}
}

func BenchmarkFlexLayout1000(b *testing.B) {
	flex := newFlexWithChildren(1000)
	constraints := geometry.BoxConstraints(0, 10000, 0, 1080)
	b.ReportAllocs()
	b.ResetTimer()
	for b.Loop() {
		flex.Layout(constraints)
	}
}

// --- Stack layout benchmarks ---

func newVStackWithChildren(n int) *VStack {
	stack := NewVStack(8, StackAlignStart)
	child := &benchLayoutable{preferredSize: geometry.Sz(100, 40)}
	for i := 0; i < n; i++ {
		stack.AddChild(child)
	}
	return stack
}

func newHStackWithChildren(n int) *HStack {
	stack := NewHStack(8, StackAlignStart)
	child := &benchLayoutable{preferredSize: geometry.Sz(50, 40)}
	for i := 0; i < n; i++ {
		stack.AddChild(child)
	}
	return stack
}

func BenchmarkVStackLayout10(b *testing.B) {
	stack := newVStackWithChildren(10)
	constraints := geometry.BoxConstraints(0, 800, 0, 600)
	b.ReportAllocs()
	b.ResetTimer()
	for b.Loop() {
		stack.Layout(constraints)
	}
}

func BenchmarkVStackLayout100(b *testing.B) {
	stack := newVStackWithChildren(100)
	constraints := geometry.BoxConstraints(0, 800, 0, 10000)
	b.ReportAllocs()
	b.ResetTimer()
	for b.Loop() {
		stack.Layout(constraints)
	}
}

func BenchmarkHStackLayout10(b *testing.B) {
	stack := newHStackWithChildren(10)
	constraints := geometry.BoxConstraints(0, 800, 0, 600)
	b.ReportAllocs()
	b.ResetTimer()
	for b.Loop() {
		stack.Layout(constraints)
	}
}

func BenchmarkHStackLayout100(b *testing.B) {
	stack := newHStackWithChildren(100)
	constraints := geometry.BoxConstraints(0, 10000, 0, 600)
	b.ReportAllocs()
	b.ResetTimer()
	for b.Loop() {
		stack.Layout(constraints)
	}
}

// --- Grid layout benchmarks ---

func newGridWithCells(cols, rows int) *GridContainer {
	columns := make([]Track, cols)
	for i := range columns {
		columns[i] = FractionTrack(1)
	}
	rowTracks := make([]Track, rows)
	for i := range rowTracks {
		rowTracks[i] = AutoTrack()
	}
	grid := NewGridContainer(columns, rowTracks)
	grid.SetGap(4)
	child := &benchLayoutable{preferredSize: geometry.Sz(80, 40)}
	for r := 0; r < rows; r++ {
		for c := 0; c < cols; c++ {
			grid.AddCell(child, r, c)
		}
	}
	return grid
}

func BenchmarkGridLayout10x10(b *testing.B) {
	grid := newGridWithCells(10, 10)
	constraints := geometry.BoxConstraints(0, 1920, 0, 1080)
	b.ReportAllocs()
	b.ResetTimer()
	for b.Loop() {
		grid.Layout(constraints)
	}
}
