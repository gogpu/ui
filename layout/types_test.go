package layout

import (
	"testing"

	"github.com/gogpu/ui/geometry"
)

func TestNodeID_IsValid(t *testing.T) {
	tests := []struct {
		name string
		id   NodeID
		want bool
	}{
		{"zero is invalid", 0, false},
		{"InvalidNodeID is invalid", InvalidNodeID, false},
		{"positive is valid", 1, true},
		{"large value is valid", 999999, true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := tt.id.IsValid(); got != tt.want {
				t.Errorf("NodeID(%d).IsValid() = %v, want %v", tt.id, got, tt.want)
			}
		})
	}
}

func TestNodeLayout_Bounds(t *testing.T) {
	layout := NodeLayout{
		Position: geometry.Point{X: 10, Y: 20},
		Size:     geometry.Size{Width: 100, Height: 50},
	}

	bounds := layout.Bounds()

	if bounds.Min.X != 10 || bounds.Min.Y != 20 {
		t.Errorf("bounds.Min = %v, want {10, 20}", bounds.Min)
	}
	if bounds.Max.X != 110 || bounds.Max.Y != 70 {
		t.Errorf("bounds.Max = %v, want {110, 70}", bounds.Max)
	}
}

func TestNodeLayout_IsZero(t *testing.T) {
	tests := []struct {
		name   string
		layout NodeLayout
		want   bool
	}{
		{"zero layout", NodeLayout{}, true},
		{"with position", NodeLayout{Position: geometry.Point{X: 1, Y: 0}}, false},
		{"with size", NodeLayout{Size: geometry.Size{Width: 1, Height: 0}}, false},
		{"with both", NodeLayout{Position: geometry.Point{X: 1}, Size: geometry.Size{Width: 1}}, false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := tt.layout.IsZero(); got != tt.want {
				t.Errorf("NodeLayout.IsZero() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestResult_IsZero(t *testing.T) {
	tests := []struct {
		name   string
		result Result
		want   bool
	}{
		{"zero result", Result{}, true},
		{"with size", Result{Size: geometry.Size{Width: 100, Height: 50}}, false},
		{"overflow only", Result{Overflow: true}, true}, // Size is still zero
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := tt.result.IsZero(); got != tt.want {
				t.Errorf("Result.IsZero() = %v, want %v", got, tt.want)
			}
		})
	}
}
