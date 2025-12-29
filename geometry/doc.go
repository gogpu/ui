// Package geometry provides fundamental geometric types for UI layout and rendering.
//
// This package defines the core geometric primitives used throughout the gogpu/ui
// toolkit: points, sizes, rectangles, constraints, and insets. All types use
// float32 coordinates for GPU compatibility and are designed for zero-allocation
// in hot paths.
//
// # Types Overview
//
//   - [Point]: A 2D point with X, Y coordinates
//   - [Size]: Dimensions with Width, Height
//   - [Rect]: Axis-aligned rectangle defined by Min and Max points
//   - [Constraints]: Flutter-style layout constraints (min/max width and height)
//   - [Insets]: Edge insets for padding and margins
//
// # Design Principles
//
// All types in this package follow these principles:
//
//   - Value semantics: Types are small structs passed by value
//   - Immutable operations: Methods return new values, never modify receiver
//   - Zero-allocation: No heap allocations in hot paths
//   - NaN safety: Operations handle NaN gracefully (replace with zero)
//
// # Usage Example
//
//	// Create a rectangle from position and size
//	rect := geometry.FromPointSize(
//	    geometry.Point{X: 10, Y: 20},
//	    geometry.Size{Width: 100, Height: 50},
//	)
//
//	// Apply padding
//	padding := geometry.UniformInsets(8)
//	contentRect := rect.Inset(padding)
//
//	// Check if a point is inside
//	if contentRect.Contains(geometry.Point{X: 50, Y: 40}) {
//	    // Handle hit
//	}
//
// # Thread Safety
//
// All types are safe for concurrent use as they use value semantics
// and do not contain any mutable state or pointers.
package geometry
