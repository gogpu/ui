// Package cdk provides the Component Development Kit — design-system-agnostic
// building blocks for composite widgets.
//
// The CDK sits between the core interfaces (widget/, event/, geometry/) and
// the generic widget implementations (core/). It provides:
//
//   - [Content] — a polymorphic content interface for user-supplied renderable content
//   - Headless behaviors (Phase 3): clickable, hoverable, overlay management
//
// # Content[C] Pattern
//
// Content[C] enables polymorphic content rendering where a host widget provides
// contextual data of type C to a user-supplied content renderer. Three built-in
// implementations cover common cases:
//
//   - [StringContent] — renders plain text via a label widget
//   - [FuncContent] — wraps a function for maximum flexibility
//   - [WidgetContent] — wraps a static widget
//
// Phase 2 widgets (Button, TextField) do NOT use Content[C] — they are simple
// enough that functional options suffice. Content[C] is designed for Phase 3
// composite widgets (VirtualizedList, Tabs, ComboBox, Dialog).
package cdk
