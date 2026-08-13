package gesture

import (
	"math"
	"time"

	"github.com/gogpu/ui/geometry"
)

// velocitySampleCount is the maximum number of position samples retained.
// Flutter uses 20 samples within a 100ms window.
const velocitySampleCount = 20

// velocityMaxAge is the maximum age of a sample for velocity estimation.
const velocityMaxAge = 100 * time.Millisecond

// velocitySample holds a single timestamped position sample.
type velocitySample struct {
	timestamp time.Duration
	position  geometry.Point
}

// VelocityTracker estimates pointer velocity from a stream of timestamped
// positions. Used by DragRecognizer to provide fling velocity at drag end.
//
// The tracker uses a least-squares linear regression over the most recent
// samples within a 100ms window. Falls back to simple delta/dt when fewer
// than 2 valid samples are available.
type VelocityTracker struct {
	samples [velocitySampleCount]velocitySample
	index   int
	count   int
}

// NewVelocityTracker creates a new velocity tracker.
func NewVelocityTracker() *VelocityTracker {
	return &VelocityTracker{}
}

// AddPosition records a timestamped position sample.
func (v *VelocityTracker) AddPosition(timestamp time.Duration, position geometry.Point) {
	v.samples[v.index] = velocitySample{
		timestamp: timestamp,
		position:  position,
	}
	v.index = (v.index + 1) % velocitySampleCount
	if v.count < velocitySampleCount {
		v.count++
	}
}

// Velocity returns the estimated velocity in logical pixels per second.
// Returns (0,0) if insufficient data is available for estimation.
func (v *VelocityTracker) Velocity() geometry.Point {
	if v.count < 2 {
		return geometry.Point{}
	}

	// Find the most recent sample for the age window.
	newest := v.newestSample()
	cutoff := newest.timestamp - velocityMaxAge

	// Collect samples within the window, ordered oldest to newest.
	type sample struct {
		t    float64 // seconds relative to oldest sample
		x, y float64
	}
	var valid []sample
	for i := 0; i < v.count; i++ {
		idx := (v.index - v.count + i + velocitySampleCount) % velocitySampleCount
		s := v.samples[idx]
		if s.timestamp >= cutoff {
			valid = append(valid, sample{
				t: s.timestamp.Seconds(),
				x: float64(s.position.X),
				y: float64(s.position.Y),
			})
		}
	}

	if len(valid) < 2 {
		return geometry.Point{}
	}

	// Least-squares linear regression: velocity = slope of position vs time.
	vx := leastSquaresSlope(valid, func(s sample) (float64, float64) { return s.t, s.x })
	vy := leastSquaresSlope(valid, func(s sample) (float64, float64) { return s.t, s.y })

	return geometry.Point{
		X: clampVelocity(float32(vx)),
		Y: clampVelocity(float32(vy)),
	}
}

// Reset clears all recorded samples.
func (v *VelocityTracker) Reset() {
	v.index = 0
	v.count = 0
}

// SampleCount returns the number of samples currently stored.
func (v *VelocityTracker) SampleCount() int {
	return v.count
}

// newestSample returns the most recently added sample.
func (v *VelocityTracker) newestSample() velocitySample {
	idx := (v.index - 1 + velocitySampleCount) % velocitySampleCount
	return v.samples[idx]
}

// leastSquaresSlope computes the slope of a least-squares linear fit.
// The extract function returns (x, y) from each sample.
func leastSquaresSlope[T any](data []T, extract func(T) (float64, float64)) float64 {
	var sumX, sumY, sumXY, sumXX float64
	n := float64(len(data))

	for _, d := range data {
		x, y := extract(d)
		sumX += x
		sumY += y
		sumXY += x * y
		sumXX += x * x
	}

	denominator := n*sumXX - sumX*sumX
	if math.Abs(denominator) < 1e-12 {
		return 0
	}

	return (n*sumXY - sumX*sumY) / denominator
}

// clampVelocity clamps a velocity component to the allowed range.
func clampVelocity(v float32) float32 {
	if v > MaxFlingVelocity {
		return MaxFlingVelocity
	}
	if v < -MaxFlingVelocity {
		return -MaxFlingVelocity
	}
	return v
}
