package widget

// SoundEvent represents a UI interaction sound type.
// Widget code uses these constants to request sound feedback without
// importing platform-specific sound packages. The actual playback
// is handled by a [SoundPlayer] registered via [RegisterSoundPlayer].
type SoundEvent int

const (
	// SoundClick is the default UI interaction sound (button click,
	// checkbox toggle, radio select, menu item activation).
	SoundClick SoundEvent = iota

	// SoundAlert is a notification/dialog sound played when a modal
	// dialog is shown.
	SoundAlert
)

// SoundPlayer is a callback that plays a system sound for the given event.
// Registered by the app/desktop layer to bridge widget sound requests
// to the platform sound API (gogpu/sound) without a direct import.
type SoundPlayer func(SoundEvent)

// soundPlayer holds the registered sound callback.
// Set by the app layer during initialization via RegisterSoundPlayer.
var soundPlayer SoundPlayer

// RegisterSoundPlayer registers the callback that plays system sounds.
// Called by the desktop layer during initialization to inject the
// platform sound implementation. Only one player may be registered;
// subsequent calls replace the previous one.
func RegisterSoundPlayer(p SoundPlayer) {
	soundPlayer = p
}

// PlaySound plays a UI sound if a [SoundPlayer] is registered.
// This is a no-op if no player has been registered (sound disabled
// or app layer not wired). Widgets call this before firing their
// user-facing callbacks so the sound plays immediately on interaction.
func PlaySound(e SoundEvent) {
	if soundPlayer != nil {
		soundPlayer(e)
	}
}
