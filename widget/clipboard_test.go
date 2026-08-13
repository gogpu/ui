package widget

import "testing"

type mockClipboard struct {
	text string
}

func (m *mockClipboard) ClipboardRead() (string, error)   { return m.text, nil }
func (m *mockClipboard) ClipboardWrite(text string) error { m.text = text; return nil }

func TestClipboardRead_NoProvider(t *testing.T) {
	old := clipboardProvider
	defer func() { clipboardProvider = old }()

	clipboardProvider = nil
	if got := ClipboardRead(); got != "" {
		t.Errorf("ClipboardRead() = %q, want empty", got)
	}
}

func TestClipboardWrite_NoProvider(t *testing.T) {
	old := clipboardProvider
	defer func() { clipboardProvider = old }()

	clipboardProvider = nil
	ClipboardWrite("test") // should not panic
}

func TestClipboard_RoundTrip(t *testing.T) {
	old := clipboardProvider
	defer func() { clipboardProvider = old }()

	mock := &mockClipboard{}
	RegisterClipboardProvider(mock)

	ClipboardWrite("hello clipboard")
	if got := ClipboardRead(); got != "hello clipboard" {
		t.Errorf("ClipboardRead() = %q, want %q", got, "hello clipboard")
	}
}

func TestRegisterClipboardProvider_Replaces(t *testing.T) {
	old := clipboardProvider
	defer func() { clipboardProvider = old }()

	first := &mockClipboard{text: "first"}
	second := &mockClipboard{text: "second"}

	RegisterClipboardProvider(first)
	RegisterClipboardProvider(second)

	if got := ClipboardRead(); got != "second" {
		t.Errorf("ClipboardRead() = %q, want %q (second provider)", got, "second")
	}
}
