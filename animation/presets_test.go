package animation

import (
	"math"
	"testing"
	"time"
)

func TestFadeIn(t *testing.T) {
	sig := newMockSignal(0)
	ctrl := NewController()

	FadeIn(sig, 100*time.Millisecond).Start(ctrl)

	// Should start at 0.
	if sig.Get() != 0 {
		t.Errorf("FadeIn initial: got %v, want 0", sig.Get())
	}

	// At completion should be 1.
	ctrl.Tick(100 * time.Millisecond)
	if sig.Get() != 1.0 {
		t.Errorf("FadeIn complete: got %v, want 1.0", sig.Get())
	}
}

func TestFadeOut(t *testing.T) {
	sig := newMockSignal(1)
	ctrl := NewController()

	FadeOut(sig, 100*time.Millisecond).Start(ctrl)

	// At completion should be 0.
	ctrl.Tick(100 * time.Millisecond)
	if sig.Get() != 0.0 {
		t.Errorf("FadeOut complete: got %v, want 0.0", sig.Get())
	}
}

func TestFadeInMidpoint(t *testing.T) {
	sig := newMockSignal(0)
	ctrl := NewController()

	FadeIn(sig, 100*time.Millisecond).Ease(Linear).Start(ctrl)

	ctrl.Tick(50 * time.Millisecond)
	if math.Abs(float64(sig.Get()-0.5)) > 0.05 {
		t.Errorf("FadeIn midpoint: got %v, want ~0.5", sig.Get())
	}
}

func TestFadeOutMidpoint(t *testing.T) {
	sig := newMockSignal(1)
	ctrl := NewController()

	FadeOut(sig, 100*time.Millisecond).Ease(Linear).Start(ctrl)

	ctrl.Tick(50 * time.Millisecond)
	if math.Abs(float64(sig.Get()-0.5)) > 0.05 {
		t.Errorf("FadeOut midpoint: got %v, want ~0.5", sig.Get())
	}
}

func TestSlideInFromBottom(t *testing.T) {
	sig := newMockSignal(100)
	ctrl := NewController()

	SlideInFromBottom(sig, 100, 100*time.Millisecond).Start(ctrl)

	// At completion should be 0.
	ctrl.Tick(100 * time.Millisecond)
	if sig.Get() != 0 {
		t.Errorf("SlideInFromBottom complete: got %v, want 0", sig.Get())
	}
}

func TestSlideInFromTop(t *testing.T) {
	sig := newMockSignal(-100)
	ctrl := NewController()

	SlideInFromTop(sig, 100, 100*time.Millisecond).Start(ctrl)

	// At completion should be 0.
	ctrl.Tick(100 * time.Millisecond)
	if sig.Get() != 0 {
		t.Errorf("SlideInFromTop complete: got %v, want 0", sig.Get())
	}
}

func TestSlideInFromLeft(t *testing.T) {
	sig := newMockSignal(-100)
	ctrl := NewController()

	SlideInFromLeft(sig, 100, 100*time.Millisecond).Start(ctrl)

	ctrl.Tick(100 * time.Millisecond)
	if sig.Get() != 0 {
		t.Errorf("SlideInFromLeft complete: got %v, want 0", sig.Get())
	}
}

func TestSlideInFromRight(t *testing.T) {
	sig := newMockSignal(100)
	ctrl := NewController()

	SlideInFromRight(sig, 100, 100*time.Millisecond).Start(ctrl)

	ctrl.Tick(100 * time.Millisecond)
	if sig.Get() != 0 {
		t.Errorf("SlideInFromRight complete: got %v, want 0", sig.Get())
	}
}

func TestSlideInFromBottomMidpoint(t *testing.T) {
	sig := newMockSignal(100)
	ctrl := NewController()

	SlideInFromBottom(sig, 100, 100*time.Millisecond).Ease(Linear).Start(ctrl)

	ctrl.Tick(50 * time.Millisecond)
	if math.Abs(float64(sig.Get()-50)) > 2 {
		t.Errorf("SlideInFromBottom midpoint: got %v, want ~50", sig.Get())
	}
}

func TestScaleIn(t *testing.T) {
	sig := newMockSignal(0.8)
	ctrl := NewController()

	ScaleIn(sig, 100*time.Millisecond).Start(ctrl)

	ctrl.Tick(100 * time.Millisecond)
	if sig.Get() != 1.0 {
		t.Errorf("ScaleIn complete: got %v, want 1.0", sig.Get())
	}
}

func TestScaleOut(t *testing.T) {
	sig := newMockSignal(1.0)
	ctrl := NewController()

	ScaleOut(sig, 100*time.Millisecond).Start(ctrl)

	ctrl.Tick(100 * time.Millisecond)
	if math.Abs(float64(sig.Get()-0.8)) > 0.01 {
		t.Errorf("ScaleOut complete: got %v, want 0.8", sig.Get())
	}
}

func TestScaleInMidpoint(t *testing.T) {
	sig := newMockSignal(0.8)
	ctrl := NewController()

	ScaleIn(sig, 100*time.Millisecond).Ease(Linear).Start(ctrl)

	ctrl.Tick(50 * time.Millisecond)
	if math.Abs(float64(sig.Get()-0.9)) > 0.02 {
		t.Errorf("ScaleIn midpoint: got %v, want ~0.9", sig.Get())
	}
}

func TestDialogEnter(t *testing.T) {
	opacity := newMockSignal(0)
	scale := newMockSignal(0.8)
	ctrl := NewController()

	DialogEnter(opacity, scale).Start(ctrl)

	// Both should animate simultaneously.
	ctrl.Tick(DurationMedium2)
	if opacity.Get() != 1.0 {
		t.Errorf("DialogEnter opacity: got %v, want 1.0", opacity.Get())
	}
	if scale.Get() != 1.0 {
		t.Errorf("DialogEnter scale: got %v, want 1.0", scale.Get())
	}
}

func TestDialogExit(t *testing.T) {
	opacity := newMockSignal(1)
	scale := newMockSignal(1.0)
	ctrl := NewController()

	DialogExit(opacity, scale).Start(ctrl)

	ctrl.Tick(DurationShort4)
	if opacity.Get() != 0.0 {
		t.Errorf("DialogExit opacity: got %v, want 0.0", opacity.Get())
	}
	if math.Abs(float64(scale.Get()-0.8)) > 0.01 {
		t.Errorf("DialogExit scale: got %v, want 0.8", scale.Get())
	}
}

func TestMenuEnter(t *testing.T) {
	opacity := newMockSignal(0)
	translateY := newMockSignal(-20)
	ctrl := NewController()

	MenuEnter(opacity, translateY, 20).Start(ctrl)

	ctrl.Tick(DurationShort4)
	if opacity.Get() != 1.0 {
		t.Errorf("MenuEnter opacity: got %v, want 1.0", opacity.Get())
	}
	if translateY.Get() != 0 {
		t.Errorf("MenuEnter translateY: got %v, want 0", translateY.Get())
	}
}

func TestMenuExit(t *testing.T) {
	opacity := newMockSignal(1)
	translateY := newMockSignal(0)
	ctrl := NewController()

	MenuExit(opacity, translateY, 20).Start(ctrl)

	ctrl.Tick(DurationShort3)
	if opacity.Get() != 0.0 {
		t.Errorf("MenuExit opacity: got %v, want 0.0", opacity.Get())
	}
	if math.Abs(float64(translateY.Get()-(-20))) > 0.5 {
		t.Errorf("MenuExit translateY: got %v, want -20", translateY.Get())
	}
}

func TestSnackbarEnter(t *testing.T) {
	translateY := newMockSignal(100)
	ctrl := NewController()

	SnackbarEnter(translateY, 100).Start(ctrl)

	ctrl.Tick(DurationMedium1)
	if translateY.Get() != 0 {
		t.Errorf("SnackbarEnter translateY: got %v, want 0", translateY.Get())
	}
}

func TestSnackbarExit(t *testing.T) {
	opacity := newMockSignal(1)
	ctrl := NewController()

	SnackbarExit(opacity).Start(ctrl)

	ctrl.Tick(DurationShort4)
	if opacity.Get() != 0.0 {
		t.Errorf("SnackbarExit opacity: got %v, want 0.0", opacity.Get())
	}
}

func TestPresetEasingAliases(t *testing.T) {
	// Verify aliases point to the correct M3 curves.
	tests := []struct {
		name  string
		alias Easing
		m3    Easing
	}{
		{"Standard", EasingStandard, M3Standard},
		{"StandardDecelerate", EasingStandardDecelerate, M3StandardDecelerate},
		{"StandardAccelerate", EasingStandardAccelerate, M3StandardAccelerate},
		{"Emphasized", EasingEmphasized, M3Emphasized},
		{"EmphasizedDecelerate", EasingEmphasizedDecelerate, M3EmphasizedDecelerate},
		{"EmphasizedAccelerate", EasingEmphasizedAccelerate, M3EmphasizedAccelerate},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Evaluate at a few points and compare.
			for _, x := range []float32{0.0, 0.25, 0.5, 0.75, 1.0} {
				got := tt.alias(x)
				want := tt.m3(x)
				if got != want {
					t.Errorf("%s(%v): got %v, want %v", tt.name, x, got, want)
				}
			}
		})
	}
}

func TestPresetsChainableWithBuilder(t *testing.T) {
	// Verify that preset builders can be further configured.
	sig := newMockSignal(0)
	ctrl := NewController()

	done := false
	FadeIn(sig, 200*time.Millisecond).
		Delay(100 * time.Millisecond).
		OnDone(func() { done = true }).
		Start(ctrl)

	// During delay.
	ctrl.Tick(50 * time.Millisecond)
	if sig.Get() != 0 {
		t.Errorf("during delay: got %v, want 0", sig.Get())
	}

	// Complete.
	ctrl.Tick(250 * time.Millisecond)
	if sig.Get() != 1.0 {
		t.Errorf("after completion: got %v, want 1.0", sig.Get())
	}
	if !done {
		t.Error("OnDone was not called")
	}
}

func TestPresetsCanBeUsedAsStartable(t *testing.T) {
	// Verify preset builders work as Startable in compositions.
	sig1 := newMockSignal(0)
	sig2 := newMockSignal(0)
	ctrl := NewController()

	NewSequence(
		FadeIn(sig1, 50*time.Millisecond),
		FadeIn(sig2, 50*time.Millisecond),
	).Start(ctrl)

	ctrl.Tick(50 * time.Millisecond)
	if sig1.Get() != 1.0 {
		t.Errorf("sequence sig1: got %v, want 1.0", sig1.Get())
	}

	ctrl.Tick(50 * time.Millisecond)
	if sig2.Get() != 1.0 {
		t.Errorf("sequence sig2: got %v, want 1.0", sig2.Get())
	}
}
