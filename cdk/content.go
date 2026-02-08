package cdk

import "github.com/gogpu/ui/widget"

// Content renders user-supplied content given a context of type C.
// The context provides information from the host widget (selection state,
// index, size constraints, etc.) so the content can adapt its rendering.
//
// Built-in implementations:
//   - [StringContent] for plain text
//   - [FuncContent] for dynamic content via a function
//   - [WidgetContent] for a pre-built static widget
type Content[C any] interface {
	Render(ctx C) widget.Widget
}

// StringContent renders a simple text label. Suitable for widgets
// where a plain string is sufficient (most common case).
//
// StringContent ignores the context parameter since a static string
// does not depend on host widget state.
type StringContent struct {
	Text string
}

// Render returns a placeholder widget representing the text.
// The actual text rendering is delegated to the host widget's painter.
func (s StringContent) Render(_ any) widget.Widget {
	return nil // placeholder — real rendering is done by the host widget's painter
}

// FuncContent wraps a function as Content. This is the most flexible option —
// the function receives full context from the host widget and returns
// a widget tree.
type FuncContent[C any] struct {
	Fn func(ctx C) widget.Widget
}

// Render calls the wrapped function with the provided context.
func (f FuncContent[C]) Render(ctx C) widget.Widget {
	if f.Fn == nil {
		return nil
	}
	return f.Fn(ctx)
}

// WidgetContent wraps a pre-built static widget as Content.
// Useful when the content does not depend on host widget context
// (e.g., a fixed icon or image).
type WidgetContent struct {
	W widget.Widget
}

// Render returns the wrapped widget, ignoring the context.
func (w WidgetContent) Render(_ any) widget.Widget {
	return w.W
}

// Compile-time interface checks.
var (
	_ Content[any] = StringContent{}
	_ Content[any] = FuncContent[any]{}
	_ Content[any] = WidgetContent{}
)
