// Package a11y provides accessibility foundation types for the gogpu/ui toolkit.
//
// This package defines the interfaces and types needed to build an accessibility
// tree that can be consumed by platform-specific assistive technology adapters
// such as Windows UI Automation, macOS NSAccessibility, and Linux AT-SPI2.
//
// The accessibility tree is a parallel semantic representation of the widget tree.
// Each widget that participates in accessibility implements the [Accessible] interface
// to provide its role, label, state, and supported actions. These are collected into
// [Node] instances that form a tree managed by a [Tree] implementation.
//
// # Architecture
//
// The package follows a layered design inspired by AccessKit:
//
//   - [Role] defines the semantic purpose of a UI element (button, checkbox, etc.)
//   - [Action] defines operations assistive technology can perform on a node
//   - [State] captures the dynamic accessibility state of a node
//   - [Accessible] is the interface widgets implement to expose their semantics
//   - [Node] is a tree node holding accessibility data with a stable [NodeID]
//   - [Tree] manages the full accessibility tree with insert/remove/query operations
//   - [Announcer] provides live region announcements for screen readers
//
// # Thread Safety
//
// [Node] and [Tree] implementations are safe for concurrent access. They use
// sync.RWMutex internally to protect shared state. The [Accessible] interface
// methods should be called from the UI thread only.
//
// # Usage
//
// Widgets expose their accessibility semantics by implementing [Accessible]:
//
//	type MyButton struct {
//	    widget.WidgetBase
//	    label string
//	}
//
//	func (b *MyButton) AccessibilityRole() a11y.Role   { return a11y.RoleButton }
//	func (b *MyButton) AccessibilityLabel() string      { return b.label }
//	func (b *MyButton) AccessibilityHint() string       { return "Activates the button" }
//	func (b *MyButton) AccessibilityValue() string      { return "" }
//	func (b *MyButton) AccessibilityState() a11y.State  { return a11y.State{} }
//	func (b *MyButton) AccessibilityActions() []a11y.Action {
//	    return []a11y.Action{a11y.ActionClick}
//	}
//
// Platform adapters (Phase 4) will consume the tree to drive native accessibility APIs.
package a11y
