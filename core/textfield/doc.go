// Package textfield provides a full-featured text input widget.
//
// Construction uses functional options for immutable configuration,
// while fluent methods handle mutable styling:
//
//	field := textfield.New(
//	    textfield.Placeholder("Enter your email"),
//	    textfield.OnChange(handleChange),
//	    textfield.InputType(textfield.TypeEmail),
//	    textfield.MaxLength(255),
//	).Padding(12)
//
// # Visual Style
//
// The visual rendering is provided by a [Painter] implementation.
// Each design system (Material 3, Fluent, Cupertino) supplies its own
// painter to render text fields in the appropriate visual style.
//
// If no painter is set, [DefaultPainter] is used, which draws a minimal
// outlined text field suitable for testing and prototyping.
//
// # Input Types
//
// Text fields support multiple input types:
//   - [TypeText] -- general-purpose text input (default)
//   - [TypePassword] -- masked input that shows dots instead of characters
//   - [TypeEmail] -- email address input
//   - [TypeNumber] -- numeric input
//   - [TypeSearch] -- search input
//
// # Text Editing
//
// Text fields support standard editing operations:
//   - Character insertion and deletion at cursor position
//   - Arrow key movement (Left/Right) with Ctrl for word-level jumps
//   - Home/End to move to line start/end
//   - Backspace and Delete for character removal
//   - Shift+arrows for text selection
//   - Ctrl+A to select all text
//   - Ctrl+C/X/V for clipboard copy/cut/paste (placeholder implementation)
//
// # Validation
//
// A [ValidationFunc] can be provided to validate the text field's value.
// When validation fails, the text field displays an error state with
// an optional error message. Validation runs on every text change.
//
// # Signal Binding
//
// Use [ValueSignal] to bind the text field to a reactive signal for two-way
// data binding:
//
//	email := state.NewSignal("")
//	field := textfield.New(textfield.ValueSignal(email))
//	// email.Get() always reflects the current text field value
//
// # Focus
//
// Text fields implement [widget.Focusable] and participate in tab navigation.
// When focused, the cursor blinks and the border color changes to indicate
// the active input state.
package textfield
