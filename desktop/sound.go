package desktop

import (
	"github.com/gogpu/gogpu/sound"
	"github.com/gogpu/ui/widget"
)

func init() {
	// Register the SoundPlayer so that widget.PlaySound() delegates to
	// gogpu/sound.Play() without widget code importing gogpu directly.
	// Sound is disabled by default; the app enables it via
	// gogpu.DefaultConfig().WithSoundFeedback(true) which calls
	// sound.SetEnabled(true).
	widget.RegisterSoundPlayer(func(e widget.SoundEvent) {
		switch e {
		case widget.SoundClick:
			sound.Play(sound.Click)
		case widget.SoundAlert:
			sound.Play(sound.Alert)
		}
	})
}
