package event

import (
	"testing"
	"time"

	"github.com/gogpu/gpucontext"
)

func TestIMEEventLifecycle(t *testing.T) {
	composition := gpucontext.IMEComposition{
		CompositionText: "かな",
		CursorBegin:     len("か"),
		CursorEnd:       len("か"),
		SelectionStart:  0,
		SelectionEnd:    len("かな"),
	}

	start := NewIMECompositionStartEvent()
	if start.Type() != TypeIME || !start.IsCompositionStart() {
		t.Fatalf("start = %#v, want TypeIME composition start", start)
	}
	update := NewIMECompositionUpdate(composition)
	if !update.IsCompositionUpdate() || update.Composition != composition {
		t.Fatalf("update = %#v, want composition %#v", update, composition)
	}
	end := NewIMECompositionEnd("仮名")
	if !end.IsCompositionEnd() || end.Committed != "仮名" {
		t.Fatalf("end = %#v, want committed text", end)
	}

	for _, tc := range []struct {
		name string
		e    *IMEEvent
		want IMEEventType
	}{
		{"cancel", NewIMECanceled(), IMECanceled},
		{"disabled", NewIMEDisabled(), IMEDisabled},
		{"delete", NewIMEDeleteSurrounding(gpucontext.IMEDeleteSurroundingEvent{Before: 2, After: 3}), IMEDeleteSurrounding},
	} {
		t.Run(tc.name, func(t *testing.T) {
			if tc.e.IMEType != tc.want || tc.e.Type() != TypeIME {
				t.Fatalf("event = %#v, want type %v", tc.e, tc.want)
			}
		})
	}
}

func TestIMECompositionUpdateWithTime(t *testing.T) {
	when := time.Date(2026, 8, 14, 12, 34, 56, 0, time.UTC)
	e := NewIMECompositionUpdateEventWithTime(gpucontext.IMEComposition{CompositionText: "x", CursorBegin: 0, CursorEnd: 0, SelectionStart: 0, SelectionEnd: 0}, when)
	if !e.Time().Equal(when) {
		t.Fatalf("Time() = %v, want %v", e.Time(), when)
	}
	if e.String() == "" {
		t.Fatal("String() should not be empty")
	}
}

func TestIMEEventTypeString(t *testing.T) {
	for _, tc := range []struct {
		typ  IMEEventType
		want string
	}{
		{IMECompositionStart, "CompositionStart"},
		{IMECompositionUpdate, "CompositionUpdate"},
		{IMECompositionEnd, "CompositionEnd"},
		{IMECanceled, "Canceled"},
		{IMEDisabled, "Disabled"},
		{IMEDeleteSurrounding, "DeleteSurrounding"},
		{IMEEventType(255), "Unknown"},
	} {
		if got := tc.typ.String(); got != tc.want {
			t.Errorf("IMEEventType(%d).String() = %q, want %q", tc.typ, got, tc.want)
		}
	}
}

func TestIMEEventAllPredicatesAndAliases(t *testing.T) {
	composition := gpucontext.IMEComposition{CompositionText: "x", CursorBegin: 0, CursorEnd: 0, SelectionStart: 0, SelectionEnd: 0}
	request := gpucontext.IMEDeleteSurroundingEvent{Before: 1, After: 2}
	events := []*IMEEvent{
		NewIMECompositionStart(),
		NewIMECompositionUpdate(composition),
		NewIMECompositionEnd("x"),
		NewIMECanceled(),
		NewIMEDisabled(),
		NewIMEDeleteSurrounding(request),
	}
	if !events[0].IsCompositionStart() || events[0].IsCompositionUpdate() || events[0].IsCompositionEnd() {
		t.Fatalf("start predicates incorrect: %#v", events[0])
	}
	if !events[1].IsCompositionUpdate() || events[1].IsCompositionStart() || events[1].IsCompositionEnd() {
		t.Fatalf("update predicates incorrect: %#v", events[1])
	}
	if !events[2].IsCompositionEnd() || events[2].IsCompositionStart() || events[2].IsCompositionUpdate() {
		t.Fatalf("end predicates incorrect: %#v", events[2])
	}
	if !events[3].IsCanceled() || !events[4].IsDisabled() || !events[5].IsDeleteSurrounding() {
		t.Fatalf("lifecycle predicates incorrect: %#v", events)
	}
	if events[5].Delete != request {
		t.Fatalf("delete request = %#v, want %#v", events[5].Delete, request)
	}
	var nilEvent *IMEEvent
	if nilEvent.IsCompositionStart() || nilEvent.IsCompositionUpdate() || nilEvent.IsCompositionEnd() || nilEvent.IsCanceled() || nilEvent.IsDisabled() || nilEvent.IsDeleteSurrounding() {
		t.Fatal("nil event predicates must be false")
	}
	if got := nilEvent.String(); got != "<nil>" {
		t.Fatalf("nil event String() = %q, want <nil>", got)
	}
}
