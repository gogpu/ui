// Package uitest provides reusable testing utilities for the gogpu/ui toolkit.
//
// This package eliminates the need for each test file to define its own mock
// Canvas and Context implementations. It provides:
//
//   - [MockCanvas]: A recording mock that implements [widget.Canvas] and captures
//     all draw calls for verification.
//   - [MockContext]: A configurable mock that implements [widget.Context] with
//     sensible defaults for headless testing.
//   - Event factory functions: [Click], [DoubleClick], [RightClick], [MouseMove],
//     [MouseDrag], [KeyPress], [WheelScroll], and [FocusGained]/[FocusLost] for
//     creating test events without boilerplate.
//   - Widget test helpers: [LayoutWidget], [DrawWidget], [SimulateClick], and
//     [SimulateKeyPress] for common test patterns.
//
// Example usage:
//
//	func TestButton_Draw(t *testing.T) {
//	    btn := button.New(button.Text("OK"))
//	    btn.SetBounds(geometry.NewRect(0, 0, 100, 40))
//
//	    canvas := &uitest.MockCanvas{}
//	    ctx := uitest.NewMockContext()
//	    btn.Draw(ctx, canvas)
//
//	    if len(canvas.Texts) == 0 {
//	        t.Fatal("expected text to be drawn")
//	    }
//	    if canvas.Texts[0].Text != "OK" {
//	        t.Errorf("text = %q, want %q", canvas.Texts[0].Text, "OK")
//	    }
//	}
package uitest
