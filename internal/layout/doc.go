// Package layout provides the internal layout engine implementation for gogpu/ui.
//
// This package is INTERNAL and not intended for public use. It implements
// constraint-based layout algorithms used by the widget system.
//
// # Architecture
//
// The layout package provides several layout algorithms:
//
//   - [Engine]: Manages layout passes with caching and dirty tracking
//   - [FlexContainer]: CSS Flexbox-style layout (row, column, wrap)
//   - [VStack], [HStack], [ZStack]: Simplified stack layouts
//   - [GridContainer]: Basic grid layout with rows and columns
//
// # Constraint-Based Layout
//
// Layout follows a constraint-passing model similar to Flutter:
//
//  1. Parent passes constraints down to children
//  2. Children compute their preferred size within constraints
//  3. Children return their computed size to parent
//  4. Parent positions children and computes its own size
//
// Constraints specify minimum and maximum dimensions. A "tight" constraint
// forces a specific size (min == max), while a "loose" constraint allows
// flexibility (min = 0).
//
// # Layout Engine
//
// The Engine manages layout passes efficiently:
//
//   - Single-pass layout for simple hierarchies
//   - Multi-pass layout for intrinsic sizing
//   - Caching of layout results to avoid redundant calculations
//   - Dirty tracking for incremental layout updates
//
// # Flexbox Layout
//
// FlexContainer implements a simplified CSS Flexbox model:
//
//   - Main axis and cross axis handling
//   - flex-grow, flex-shrink for space distribution
//   - justify-content: Start, End, Center, SpaceBetween, SpaceAround, SpaceEvenly
//   - align-items: Start, End, Center, Stretch
//   - direction: Row, Column, RowReverse, ColumnReverse
//   - wrap support for flowing to multiple lines
//
// # Stack Layouts
//
// Stack layouts are common shortcuts:
//
//   - VStack: Vertical stack with spacing and alignment
//   - HStack: Horizontal stack with spacing and alignment
//   - ZStack: Overlay stack for layering widgets
//
// # Grid Layout
//
// GridContainer provides basic grid layout:
//
//   - Fixed or fractional column/row definitions
//   - Gap between cells
//   - Cell spanning (optional)
//
// # Thread Safety
//
// Layout types are NOT thread-safe. All layout operations must occur on the
// main/UI thread during the Layout phase.
//
// # Usage
//
// This package is used internally by the UI framework. Application code should
// use the public layout widgets instead of directly using this package.
//
//	// Internal framework usage
//	engine := layout.NewEngine()
//	flex := layout.NewFlexContainer(layout.Row, layout.JustifyStart, layout.AlignStretch)
//	flex.AddChild(child1, layout.FlexItem{Grow: 1})
//	flex.AddChild(child2, layout.FlexItem{Grow: 2})
//	size := engine.Layout(flex, constraints)
package layout
