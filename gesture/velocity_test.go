package gesture

import (
	"math"
	"testing"
	"time"

	"github.com/gogpu/ui/geometry"
)

func TestVelocityTracker_Empty(t *testing.T) {
	vt := NewVelocityTracker()
	vel := vt.Velocity()
	if vel.X != 0 || vel.Y != 0 {
		t.Errorf("empty tracker velocity = %v, want (0, 0)", vel)
	}
}

func TestVelocityTracker_SingleSample(t *testing.T) {
	vt := NewVelocityTracker()
	vt.AddPosition(0, geometry.Pt(0, 0))
	vel := vt.Velocity()
	if vel.X != 0 || vel.Y != 0 {
		t.Errorf("single sample velocity = %v, want (0, 0)", vel)
	}
}

func TestVelocityTracker_ConstantHorizontalVelocity(t *testing.T) {
	vt := NewVelocityTracker()

	// 100 px/s horizontal velocity: 10px every 100ms.
	for i := 0; i < 5; i++ {
		ts := time.Duration(i*100) * time.Millisecond
		vt.AddPosition(ts, geometry.Pt(float32(i*10), 0))
	}

	vel := vt.Velocity()
	if math.Abs(float64(vel.X)-100) > 5 {
		t.Errorf("horizontal velocity = %.1f, want ~100", vel.X)
	}
	if math.Abs(float64(vel.Y)) > 1 {
		t.Errorf("vertical velocity = %.1f, want ~0", vel.Y)
	}
}

func TestVelocityTracker_ConstantVerticalVelocity(t *testing.T) {
	vt := NewVelocityTracker()

	// 200 px/s vertical velocity: 20px every 100ms.
	for i := 0; i < 5; i++ {
		ts := time.Duration(i*100) * time.Millisecond
		vt.AddPosition(ts, geometry.Pt(0, float32(i*20)))
	}

	vel := vt.Velocity()
	if math.Abs(float64(vel.Y)-200) > 5 {
		t.Errorf("vertical velocity = %.1f, want ~200", vel.Y)
	}
	if math.Abs(float64(vel.X)) > 1 {
		t.Errorf("horizontal velocity = %.1f, want ~0", vel.X)
	}
}

func TestVelocityTracker_DiagonalVelocity(t *testing.T) {
	vt := NewVelocityTracker()

	// 100 px/s in both axes.
	for i := 0; i < 5; i++ {
		ts := time.Duration(i*100) * time.Millisecond
		vt.AddPosition(ts, geometry.Pt(float32(i*10), float32(i*10)))
	}

	vel := vt.Velocity()
	if math.Abs(float64(vel.X)-100) > 5 {
		t.Errorf("X velocity = %.1f, want ~100", vel.X)
	}
	if math.Abs(float64(vel.Y)-100) > 5 {
		t.Errorf("Y velocity = %.1f, want ~100", vel.Y)
	}
}

func TestVelocityTracker_VelocityClamped(t *testing.T) {
	vt := NewVelocityTracker()

	// Extreme velocity: 10000 px in 10ms.
	vt.AddPosition(0, geometry.Pt(0, 0))
	vt.AddPosition(10*time.Millisecond, geometry.Pt(10000, 0))

	vel := vt.Velocity()
	if vel.X > MaxFlingVelocity {
		t.Errorf("velocity = %.1f, should be clamped to %.1f", vel.X, MaxFlingVelocity)
	}
}

func TestVelocityTracker_Reset(t *testing.T) {
	vt := NewVelocityTracker()

	vt.AddPosition(0, geometry.Pt(0, 0))
	vt.AddPosition(100*time.Millisecond, geometry.Pt(100, 0))

	vt.Reset()
	if vt.SampleCount() != 0 {
		t.Errorf("SampleCount after reset = %d, want 0", vt.SampleCount())
	}
	vel := vt.Velocity()
	if vel.X != 0 || vel.Y != 0 {
		t.Errorf("velocity after reset = %v, want (0, 0)", vel)
	}
}

func TestVelocityTracker_OldSamplesDiscarded(t *testing.T) {
	vt := NewVelocityTracker()

	// Add old samples (before the window).
	vt.AddPosition(0, geometry.Pt(0, 0))
	vt.AddPosition(10*time.Millisecond, geometry.Pt(1000, 0))

	// Add recent samples with different velocity.
	// These should dominate because old ones are outside the 100ms window.
	base := 200 * time.Millisecond
	for i := 0; i < 5; i++ {
		ts := base + time.Duration(i*20)*time.Millisecond
		vt.AddPosition(ts, geometry.Pt(float32(i*2), 0))
	}

	vel := vt.Velocity()
	// The recent velocity is 2px/20ms = 100 px/s.
	if math.Abs(float64(vel.X)-100) > 15 {
		t.Errorf("velocity from recent samples = %.1f, want ~100", vel.X)
	}
}

func TestVelocityTracker_CircularBuffer(t *testing.T) {
	vt := NewVelocityTracker()

	// Fill beyond capacity.
	for i := 0; i <= velocitySampleCount+5; i++ {
		ts := time.Duration(i*5) * time.Millisecond
		vt.AddPosition(ts, geometry.Pt(float32(i), 0))
	}

	if vt.SampleCount() != velocitySampleCount {
		t.Errorf("SampleCount = %d, want %d", vt.SampleCount(), velocitySampleCount)
	}

	// Should still produce valid velocity.
	vel := vt.Velocity()
	if vel.X <= 0 {
		t.Error("velocity should be positive after filling buffer")
	}
}

func TestVelocityTracker_NegativeVelocityClamped(t *testing.T) {
	vt := NewVelocityTracker()

	// Extreme negative velocity.
	vt.AddPosition(0, geometry.Pt(10000, 0))
	vt.AddPosition(10*time.Millisecond, geometry.Pt(0, 0))

	vel := vt.Velocity()
	if vel.X < -MaxFlingVelocity {
		t.Errorf("velocity = %.1f, should be clamped to %.1f", vel.X, -MaxFlingVelocity)
	}
}
