package gesture_test

import (
	"testing"

	"github.com/gogpu/ui/gesture"
)

// testGestureWidget is a mock widget implementing GestureAware.
type testGestureWidget struct {
	recognizers []gesture.Recognizer
}

func (w *testGestureWidget) GestureRecognizers() []gesture.Recognizer {
	return w.recognizers
}

// Compile-time interface compliance check.
var _ gesture.GestureAware = (*testGestureWidget)(nil)

func TestGestureAware_InterfaceCompliance(t *testing.T) {
	click := gesture.NewClickRecognizer(gesture.ClickConfig{})
	w := &testGestureWidget{
		recognizers: []gesture.Recognizer{click},
	}

	recs := w.GestureRecognizers()
	if len(recs) != 1 {
		t.Fatalf("GestureRecognizers() returned %d, want 1", len(recs))
	}
	if recs[0] != click {
		t.Error("GestureRecognizers() returned wrong recognizer")
	}
}

func TestGestureAware_EmptyRecognizers(t *testing.T) {
	w := &testGestureWidget{}
	recs := w.GestureRecognizers()
	if len(recs) != 0 {
		t.Errorf("GestureRecognizers() returned %d, want 0", len(recs))
	}
}

func TestGestureAware_MultipleRecognizers(t *testing.T) {
	click := gesture.NewClickRecognizer(gesture.ClickConfig{})
	drag := gesture.NewDragRecognizer(gesture.DragConfig{})
	w := &testGestureWidget{
		recognizers: []gesture.Recognizer{click, drag},
	}

	recs := w.GestureRecognizers()
	if len(recs) != 2 {
		t.Fatalf("GestureRecognizers() returned %d, want 2", len(recs))
	}
}

// nonGestureWidget does NOT implement GestureAware.
type nonGestureWidget struct{}

func TestGestureAware_TypeAssertionFails(t *testing.T) {
	w := &nonGestureWidget{}
	_, ok := interface{}(w).(gesture.GestureAware)
	if ok {
		t.Error("nonGestureWidget should not implement GestureAware")
	}
}
