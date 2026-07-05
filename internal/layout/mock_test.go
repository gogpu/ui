package layout

import "github.com/gogpu/ui/geometry"

// mockLayoutable is a simple layoutable for testing.
type mockLayoutable struct {
	id            uint64
	preferredSize geometry.Size
	layoutCalls   int
	children      []Layoutable
}

func (m *mockLayoutable) ID() uint64 {
	return m.id
}

func (m *mockLayoutable) Layout(constraints geometry.Constraints) geometry.Size {
	m.layoutCalls++
	return constraints.Constrain(m.preferredSize)
}

func (m *mockLayoutable) Children() []Layoutable {
	return m.children
}
