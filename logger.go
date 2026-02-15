package ui

import (
	"context"
	"log/slog"
	"sync/atomic"
)

// nopHandler silently discards all log records.
// Enabled returns false so the caller skips message formatting entirely,
// making disabled logging effectively zero-cost.
type nopHandler struct{}

func (nopHandler) Enabled(context.Context, slog.Level) bool  { return false }
func (nopHandler) Handle(context.Context, slog.Record) error { return nil }
func (nopHandler) WithAttrs([]slog.Attr) slog.Handler        { return nopHandler{} }
func (nopHandler) WithGroup(string) slog.Handler             { return nopHandler{} }

// loggerPtr stores the active logger. Accessed atomically so that
// SetLogger can be called concurrently with logging from any goroutine.
var loggerPtr atomic.Pointer[slog.Logger]

func init() {
	l := slog.New(nopHandler{})
	loggerPtr.Store(l)
}

// slogger returns the current package logger.
// All logging in ui goes through this function.
func slogger() *slog.Logger { return loggerPtr.Load() }

// SetLogger configures the logger for the ui toolkit.
// By default, ui produces no log output. Call SetLogger to enable logging.
//
// SetLogger is safe for concurrent use: it stores the new logger atomically.
// Pass nil to disable logging (restore default silent behavior).
//
// Log levels used by ui:
//   - [slog.LevelDebug]: internal diagnostics (layout calculations, focus changes)
//   - [slog.LevelInfo]: important lifecycle events (plugin loaded, theme applied)
//   - [slog.LevelWarn]: non-fatal issues (widget paint errors, constraint violations)
//
// Example:
//
//	// Enable info-level logging to stderr:
//	ui.SetLogger(slog.Default())
//
//	// Enable debug-level logging for full diagnostics:
//	ui.SetLogger(slog.New(slog.NewTextHandler(os.Stderr, &slog.HandlerOptions{
//	    Level: slog.LevelDebug,
//	})))
func SetLogger(l *slog.Logger) {
	if l == nil {
		l = slog.New(nopHandler{})
	}
	loggerPtr.Store(l)
}

// Logger returns the current logger used by ui.
// Sub-packages call this to share the same logger configuration
// without introducing import cycles.
//
// Logger is safe for concurrent use.
func Logger() *slog.Logger {
	return loggerPtr.Load()
}
