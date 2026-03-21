// Package titlebar provides a custom window title bar widget.
//
// A TitleBar replaces the OS-native title bar with a GPU-rendered bar that
// integrates with the application's design system. It renders a horizontal
// bar with three zones: leading (left-aligned), center, and trailing (right-aligned
// window control buttons).
//
// Construction uses functional options:
//
//	tb := titlebar.New(
//	    titlebar.Title("My Application"),
//	    titlebar.Leading(menuBtn, projectLabel),
//	    titlebar.Center(searchWidget),
//	    titlebar.Height(40),
//	    titlebar.PainterOpt(painter),
//	)
//
// # Zones
//
// The title bar is divided into three horizontal zones:
//   - Leading: left-aligned widgets (menu button, project name, branch selector)
//   - Center: centered widgets (search bar, run configuration)
//   - Trailing: auto-generated window control buttons (minimize, maximize/restore, close)
//
// # Window Chrome Integration
//
// When a [WindowChrome] is provided via [Chrome], the title bar registers a
// hit-test callback so that empty space acts as a drag region for moving the
// window, and the control buttons perform minimize/maximize/close operations.
// Without a WindowChrome, the title bar renders as a purely visual bar with
// no window management capabilities.
//
// # Visual Style
//
// The visual rendering is delegated to a [Painter] implementation.
// Each design system can supply its own painter. If no painter is set,
// [DefaultPainter] is used.
//
// # Accessibility
//
// The title bar has the [a11y.RoleBanner] role, following the WAI-ARIA
// banner landmark pattern for site/application-wide content.
package titlebar
