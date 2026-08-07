package widget

import "testing"

func TestPlaySound_NoPlayer(t *testing.T) {
	// Save and restore the global player.
	old := soundPlayer
	defer func() { soundPlayer = old }()

	soundPlayer = nil

	// Should not panic when no player is registered.
	PlaySound(SoundClick)
	PlaySound(SoundAlert)
}

func TestPlaySound_WithPlayer(t *testing.T) {
	old := soundPlayer
	defer func() { soundPlayer = old }()

	var received []SoundEvent
	RegisterSoundPlayer(func(e SoundEvent) {
		received = append(received, e)
	})

	PlaySound(SoundClick)
	PlaySound(SoundAlert)
	PlaySound(SoundClick)

	if len(received) != 3 {
		t.Fatalf("expected 3 events, got %d", len(received))
	}
	if received[0] != SoundClick {
		t.Errorf("event[0] = %d, want SoundClick(%d)", received[0], SoundClick)
	}
	if received[1] != SoundAlert {
		t.Errorf("event[1] = %d, want SoundAlert(%d)", received[1], SoundAlert)
	}
	if received[2] != SoundClick {
		t.Errorf("event[2] = %d, want SoundClick(%d)", received[2], SoundClick)
	}
}

func TestRegisterSoundPlayer_Replaces(t *testing.T) {
	old := soundPlayer
	defer func() { soundPlayer = old }()

	var first, second bool
	RegisterSoundPlayer(func(SoundEvent) { first = true })
	RegisterSoundPlayer(func(SoundEvent) { second = true })

	PlaySound(SoundClick)

	if first {
		t.Error("first player should NOT have been called after replacement")
	}
	if !second {
		t.Error("second player should have been called")
	}
}
