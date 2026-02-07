// Package app provides the bridge between the UI widget tree and the windowing system.
//
// The app package connects gogpu/ui widgets to a real window via gpucontext
// interfaces (WindowProvider, PlatformProvider, EventSource). This enables
// dependency inversion: ui depends on gpucontext interfaces, while gogpu
// implements those interfaces.
//
// # Architecture
//
// App is the main entry point. It holds a Window, theme, and scheduler.
// Window manages the widget tree and coordinates layout, draw, and event
// dispatch. EventBridge translates platform-level events from gpucontext
// into ui/event types that widgets understand.
//
//	user code
//	   |
//	   v
//	app.New(opts...)  -->  App  -->  Window  -->  widget tree
//	                        |            |
//	                     Theme      EventBridge
//	                        |            |
//	                     Scheduler   gpucontext.EventSource
//
// # Usage
//
// Create an App with functional options, set a root widget, and let the
// host application (gogpu.App) drive the render loop:
//
//	// In the host application (e.g., gogpu.App's draw callback):
//	uiApp := app.New(
//	    app.WithWindowProvider(wp),
//	    app.WithPlatformProvider(pp),
//	    app.WithEventSource(es),
//	)
//	uiApp.SetRoot(myRootWidget)
//
//	// Each frame, the host calls:
//	uiApp.Frame()
//
// # Headless Mode
//
// When no providers are supplied, the App operates in headless mode.
// This is useful for testing and batch processing:
//
//	uiApp := app.New() // headless, no window
//	uiApp.SetRoot(myWidget)
//	uiApp.Frame() // performs layout and draw with defaults
//
// # Thread Safety
//
// App and Window are NOT safe for concurrent access. All operations must
// occur on the main/UI thread. This matches the single-threaded event loop
// model of windowing systems.
//
// # Design Principles
//
//   - Dependency inversion: ui imports gpucontext interfaces, not gogpu
//   - No goroutines: everything runs on the caller's thread
//   - No global state: App is instance-based
//   - Testable: mock providers enable isolated testing
//   - Composable: functional options for configuration
package app
