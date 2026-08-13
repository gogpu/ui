package widget

// ClipboardProvider reads and writes text to the system clipboard.
// Registered by the app/desktop layer to bridge widget clipboard
// requests to the platform API (gogpu PlatformProvider) without a
// direct import. The same DI pattern as SoundPlayer.
type ClipboardProvider interface {
	ClipboardRead() (string, error)
	ClipboardWrite(text string) error
}

var clipboardProvider ClipboardProvider

// RegisterClipboardProvider registers the platform clipboard implementation.
// Called by the desktop layer during initialization to inject the platform
// clipboard. Only one provider may be registered; subsequent calls replace
// the previous one.
func RegisterClipboardProvider(p ClipboardProvider) {
	clipboardProvider = p
}

// ClipboardRead reads text from the system clipboard.
// Returns empty string if no provider is registered or clipboard is empty.
func ClipboardRead() string {
	if clipboardProvider == nil {
		return ""
	}
	text, _ := clipboardProvider.ClipboardRead()
	return text
}

// ClipboardWrite writes text to the system clipboard.
// No-op if no provider is registered.
func ClipboardWrite(text string) {
	if clipboardProvider == nil {
		return
	}
	_ = clipboardProvider.ClipboardWrite(text)
}
