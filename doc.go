// Package ui provides an enterprise-grade GUI toolkit for Go.
//
// gogpu/ui is designed for building professional applications such as
// IDEs, design tools, CAD applications, and Chrome/Electron-class apps.
// It provides a reactive, declarative API with GPU-accelerated rendering.
//
// # Quick Start
//
// The simplest ui program creates a window with a button:
//
//	package main
//
//	import (
//	    "github.com/gogpu/gogpu"
//	    "github.com/gogpu/ui"
//	    "github.com/gogpu/ui/widgets"
//	    "github.com/gogpu/ui/layout"
//	)
//
//	func main() {
//	    app := gogpu.NewApp(gogpu.Config{
//	        Title:  "My App",
//	        Width:  800,
//	        Height: 600,
//	    })
//
//	    root := layout.VStack(
//	        widgets.Text("Hello, World!"),
//	        widgets.Button("Click Me").OnClick(func() {
//	            println("Clicked!")
//	        }),
//	    ).Padding(16)
//
//	    app.SetRoot(root)
//	    app.Run()
//	}
//
// # Architecture
//
// gogpu/ui uses a layered architecture:
//
//   - core: Widget interface, WidgetBase, Context
//   - layout: VStack, HStack, Grid, Flexbox
//   - widgets: Button, TextField, Dropdown, etc.
//   - theme: Material 3, Fluent, Cupertino
//   - state: Signals integration (coregx/signals)
//
// # State Management
//
// Use signals for reactive state:
//
//	count := signals.New(0)
//
//	widgets.Text(signals.Computed(func() string {
//	    return fmt.Sprintf("Count: %d", count.Get())
//	}))
//
// # Platform Support
//
//   - Windows: Win32
//   - macOS: Cocoa
//   - Linux: X11/Wayland
//
// # Dependencies
//
// gogpu/ui depends on:
//   - github.com/gogpu/gg - 2D graphics
//   - github.com/gogpu/gogpu - Windowing
//   - github.com/coregx/signals - State management
//
// # Status
//
// This package is in the planning phase (v0.0.0).
// See ROADMAP.md for the development plan.
package ui
