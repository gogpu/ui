package ui

import (
	"bytes"
	"context"
	"log/slog"
	"strings"
	"sync"
	"testing"
)

func TestNopHandler_Enabled(t *testing.T) {
	h := nopHandler{}
	for _, level := range []slog.Level{slog.LevelDebug, slog.LevelInfo, slog.LevelWarn, slog.LevelError} {
		if h.Enabled(context.Background(), level) {
			t.Errorf("nopHandler.Enabled(%v) = true, want false", level)
		}
	}
}

func TestNopHandler_Handle(t *testing.T) {
	h := nopHandler{}
	if err := h.Handle(context.Background(), slog.Record{}); err != nil {
		t.Errorf("nopHandler.Handle() = %v, want nil", err)
	}
}

func TestNopHandler_WithAttrs(t *testing.T) {
	h := nopHandler{}
	got := h.WithAttrs([]slog.Attr{slog.String("key", "val")})
	if _, ok := got.(nopHandler); !ok {
		t.Errorf("nopHandler.WithAttrs() returned %T, want nopHandler", got)
	}
}

func TestNopHandler_WithGroup(t *testing.T) {
	h := nopHandler{}
	got := h.WithGroup("group")
	if _, ok := got.(nopHandler); !ok {
		t.Errorf("nopHandler.WithGroup() returned %T, want nopHandler", got)
	}
}

func TestLoggerDefaultSilent(t *testing.T) {
	l := Logger()
	if l == nil {
		t.Fatal("Logger() returned nil")
	}
	for _, level := range []slog.Level{slog.LevelDebug, slog.LevelInfo, slog.LevelWarn} {
		if l.Enabled(context.Background(), level) {
			t.Errorf("default logger should not be enabled for %v", level)
		}
	}
}

func TestSetLogger(t *testing.T) {
	orig := Logger()
	t.Cleanup(func() { SetLogger(orig) })

	var buf bytes.Buffer
	custom := slog.New(slog.NewTextHandler(&buf, &slog.HandlerOptions{
		Level: slog.LevelDebug,
	}))

	SetLogger(custom)

	got := Logger()
	if got != custom {
		t.Error("Logger() did not return the custom logger set via SetLogger")
	}

	got.Info("test message", "key", "value")
	if !strings.Contains(buf.String(), "test message") {
		t.Errorf("expected log output to contain 'test message', got: %s", buf.String())
	}
}

func TestSetLoggerNilRestoresSilent(t *testing.T) {
	orig := Logger()
	t.Cleanup(func() { SetLogger(orig) })

	SetLogger(slog.Default())
	SetLogger(nil)

	l := Logger()
	if l == nil {
		t.Fatal("SetLogger(nil) should set nop logger, not nil")
	}
	if l.Enabled(context.Background(), slog.LevelError) {
		t.Error("SetLogger(nil) should produce a disabled logger")
	}
}

func TestSloggerReturnsSameAsLogger(t *testing.T) {
	if slogger() != Logger() {
		t.Error("slogger() and Logger() should return the same instance")
	}
}

func TestLoggerConcurrentAccess(t *testing.T) {
	orig := Logger()
	t.Cleanup(func() { SetLogger(orig) })

	var wg sync.WaitGroup
	const goroutines = 100

	for range goroutines {
		wg.Add(1)
		go func() {
			defer wg.Done()
			l := Logger()
			if l == nil {
				t.Error("Logger() returned nil during concurrent access")
			}
			l.Debug("concurrent read")
		}()
	}

	for range goroutines {
		wg.Add(1)
		go func() {
			defer wg.Done()
			SetLogger(slog.Default())
			SetLogger(nil)
		}()
	}

	wg.Wait()
}

func BenchmarkLoggerLoad(b *testing.B) {
	b.ReportAllocs()
	for b.Loop() {
		l := Logger()
		_ = l
	}
}

func BenchmarkLoggerDisabledLog(b *testing.B) {
	l := Logger()
	b.ReportAllocs()
	for b.Loop() {
		l.Debug("message", "key", "value")
	}
}
