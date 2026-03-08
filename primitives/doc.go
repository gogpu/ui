// Package primitives provides the fundamental widget building blocks for the
// gogpu/ui framework: [BoxWidget], [TextWidget], and [ImageWidget].
//
// These three primitives are the lowest-level concrete widgets from which all
// higher-level components are composed. Each primitive implements
// [widget.Widget] for layout, drawing, and event handling, and
// [a11y.Accessible] for accessibility.
//
// # Fluent Builder API
//
// All primitives use a Tailwind-inspired fluent builder API where the
// constructor returns a pointer to the widget struct, and style methods
// return that same pointer for chaining. There is no separate Build step;
// the returned value is a ready-to-use widget.
//
//	root := primitives.Box(
//	    primitives.Text("Hello World").
//	        FontSize(24).
//	        Color(widget.RGBA(0, 0, 0, 1)),
//	    primitives.Text("Subtitle").
//	        FontSize(14),
//	).Padding(16).Background(widget.RGBA(1, 1, 1, 1)).Rounded(8)
//
// # Box
//
// Box is a container widget that lays out its children in a vertical stack
// with optional padding, background, border, rounded corners, shadow, and
// gap between children.
//
//	card := primitives.Box(
//	    primitives.Text("Title").FontSize(18).Bold(),
//	    primitives.Text("Body text").FontSize(14),
//	).Padding(16).Background(widget.Hex(0xFFFFFF)).Rounded(12).Shadow(2)
//
// # Text
//
// Text displays static or reactive text content. Static text is created with
// [Text], reactive text via a function with [TextFn], and signal-bound text
// with [TextWidget.ContentSignal].
//
//	label := primitives.Text("Hello").FontSize(16).Bold()
//
//	counter := state.NewSignal(0)
//	dynamic := primitives.TextFn(func() string {
//	    return fmt.Sprintf("Count: %d", counter.Get())
//	}).FontSize(14)
//
//	name := state.NewSignal("Alice")
//	bound := primitives.Text("").ContentSignal(name).FontSize(14)
//
// # Image
//
// Image displays a raster image with support for different fit modes (cover,
// contain, fill) and an alt text label for accessibility.
//
//	img := primitives.Image(mySource).
//	    Size(200, 150).
//	    Cover().
//	    Alt("Photo of a sunset").
//	    Rounded(8)
package primitives
