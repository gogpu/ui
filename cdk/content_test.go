package cdk_test

import (
	"testing"

	"github.com/gogpu/ui/cdk"
	"github.com/gogpu/ui/event"
	"github.com/gogpu/ui/geometry"
	"github.com/gogpu/ui/widget"
)

// stubWidget is a minimal Widget implementation for testing.
type stubWidget struct {
	widget.WidgetBase
	id string
}

func newStubWidget(id string) *stubWidget {
	w := &stubWidget{id: id}
	w.SetVisible(true)
	w.SetEnabled(true)
	return w
}

func (w *stubWidget) Layout(_ widget.Context, c geometry.Constraints) geometry.Size {
	return c.Constrain(geometry.Sz(10, 10))
}

func (w *stubWidget) Draw(_ widget.Context, _ widget.Canvas) {}

func (w *stubWidget) Event(_ widget.Context, _ event.Event) bool { return false }

// Compile-time check that stubWidget implements Widget.
var _ widget.Widget = (*stubWidget)(nil)

func TestStringContent_Render_ReturnsNil(t *testing.T) {
	sc := cdk.StringContent{Text: "hello"}

	got := sc.Render(nil)
	if got != nil {
		t.Errorf("StringContent.Render() = %v, want nil", got)
	}
}

func TestStringContent_Text(t *testing.T) {
	tests := []struct {
		name string
		text string
	}{
		{"empty string", ""},
		{"simple text", "OK"},
		{"unicode text", "Привет мир"},
		{"long text", "Lorem ipsum dolor sit amet, consectetur adipiscing elit"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			sc := cdk.StringContent{Text: tt.text}
			if sc.Text != tt.text {
				t.Errorf("StringContent.Text = %q, want %q", sc.Text, tt.text)
			}
		})
	}
}

func TestFuncContent_Render_CallsFunction(t *testing.T) {
	var called bool
	stub := newStubWidget("func-result")

	fc := cdk.FuncContent[any]{
		Fn: func(_ any) widget.Widget {
			called = true
			return stub
		},
	}

	got := fc.Render("some context")
	if !called {
		t.Error("FuncContent.Render() did not call Fn")
	}
	if got != stub {
		t.Errorf("FuncContent.Render() returned wrong widget")
	}
}

func TestFuncContent_Render_NilFn(t *testing.T) {
	fc := cdk.FuncContent[any]{Fn: nil}

	got := fc.Render("context")
	if got != nil {
		t.Errorf("FuncContent.Render() with nil Fn = %v, want nil", got)
	}
}

func TestFuncContent_ReceivesCorrectContext(t *testing.T) {
	var received string

	fc := cdk.FuncContent[string]{
		Fn: func(ctx string) widget.Widget {
			received = ctx
			return nil
		},
	}

	fc.Render("expected-context")
	if received != "expected-context" {
		t.Errorf("FuncContent received context %q, want %q", received, "expected-context")
	}
}

func TestFuncContent_GenericContextType(t *testing.T) {
	type itemContext struct {
		Index    int
		Selected bool
	}

	var received itemContext
	stub := newStubWidget("generic-result")

	fc := cdk.FuncContent[itemContext]{
		Fn: func(ctx itemContext) widget.Widget {
			received = ctx
			if ctx.Selected {
				return stub
			}
			return nil
		},
	}

	t.Run("selected item", func(t *testing.T) {
		ctx := itemContext{Index: 3, Selected: true}
		got := fc.Render(ctx)
		if received != ctx {
			t.Errorf("received context = %+v, want %+v", received, ctx)
		}
		if got != stub {
			t.Errorf("Render() returned wrong widget, want stub")
		}
	})

	t.Run("unselected item", func(t *testing.T) {
		ctx := itemContext{Index: 7, Selected: false}
		got := fc.Render(ctx)
		if received != ctx {
			t.Errorf("received context = %+v, want %+v", received, ctx)
		}
		if got != nil {
			t.Errorf("Render() = %v, want nil", got)
		}
	})
}

func TestWidgetContent_Render_ReturnsWidget(t *testing.T) {
	stub := newStubWidget("wrapped")

	wc := cdk.WidgetContent{W: stub}

	got := wc.Render(nil)
	if got != stub {
		t.Errorf("WidgetContent.Render() returned wrong widget")
	}
}

func TestWidgetContent_Render_NilWidget(t *testing.T) {
	wc := cdk.WidgetContent{W: nil}

	got := wc.Render("ignored context")
	if got != nil {
		t.Errorf("WidgetContent.Render() with nil W = %v, want nil", got)
	}
}

func TestWidgetContent_IgnoresContext(t *testing.T) {
	stub := newStubWidget("static")
	wc := cdk.WidgetContent{W: stub}

	// Render with different contexts — should always return the same widget.
	contexts := []any{nil, "string", 42, struct{}{}}
	for _, ctx := range contexts {
		got := wc.Render(ctx)
		if got != stub {
			t.Errorf("WidgetContent.Render(%v) returned wrong widget", ctx)
		}
	}
}

func TestContent_InterfaceCompliance(t *testing.T) {
	// Verify all three types satisfy Content[any] at runtime via type assertions.
	tests := []struct {
		name    string
		content cdk.Content[any]
	}{
		{"StringContent", cdk.StringContent{Text: "test"}},
		{"FuncContent", cdk.FuncContent[any]{Fn: nil}},
		{"WidgetContent", cdk.WidgetContent{W: nil}},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.content == nil {
				t.Errorf("%s does not satisfy Content[any]", tt.name)
			}
			// Calling Render should not panic.
			_ = tt.content.Render(nil)
		})
	}
}

func TestFuncContent_ZeroValue(t *testing.T) {
	// Zero value of FuncContent (Fn is nil) should be safe to use.
	var fc cdk.FuncContent[int]

	got := fc.Render(42)
	if got != nil {
		t.Errorf("zero-value FuncContent.Render() = %v, want nil", got)
	}
}

func TestStringContent_ZeroValue(t *testing.T) {
	// Zero value of StringContent should be safe to use.
	var sc cdk.StringContent

	if sc.Text != "" {
		t.Errorf("zero-value StringContent.Text = %q, want empty", sc.Text)
	}

	got := sc.Render(nil)
	if got != nil {
		t.Errorf("zero-value StringContent.Render() = %v, want nil", got)
	}
}

func TestWidgetContent_ZeroValue(t *testing.T) {
	// Zero value of WidgetContent should be safe to use.
	var wc cdk.WidgetContent

	if wc.W != nil {
		t.Errorf("zero-value WidgetContent.W = %v, want nil", wc.W)
	}

	got := wc.Render(nil)
	if got != nil {
		t.Errorf("zero-value WidgetContent.Render() = %v, want nil", got)
	}
}
