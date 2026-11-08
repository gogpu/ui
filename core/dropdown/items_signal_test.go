package dropdown_test

import (
	"testing"

	"github.com/gogpu/ui/core/dropdown"
	"github.com/gogpu/ui/geometry"
	"github.com/gogpu/ui/state"
	"github.com/gogpu/ui/widget"
)

// --- Signal-bound items (issue #243) ---

// triggerText draws the dropdown trigger and returns the text it painted.
func triggerText(t *testing.T, dd *dropdown.Widget) string {
	t.Helper()
	dd.SetBounds(geometry.NewRect(0, 0, 200, 48))
	canvas := &recordingCanvas{}
	dd.Draw(widget.NewContext(), canvas)
	if len(canvas.drawTexts) == 0 {
		t.Fatal("trigger painted no text")
	}
	return canvas.drawTexts[0].text
}

func TestItemsSignal_ItemsFollowSignal(t *testing.T) {
	sig := state.NewSignal([]string{"Red", "Green", "Blue"})
	dd := dropdown.New(
		dropdown.Items("stale"),
		dropdown.ItemsSignal(sig),
		dropdown.Selected(0),
	)

	if got := triggerText(t, dd); got != "Red" {
		t.Fatalf("trigger text = %q, want %q", got, "Red")
	}
	if got := dd.SelectedValue(); got != "Red" {
		t.Errorf("SelectedValue = %q, want %q", got, "Red")
	}

	sig.Set([]string{"Cyan", "Magenta"})

	if got := triggerText(t, dd); got != "Cyan" {
		t.Errorf("trigger text after update = %q, want %q", got, "Cyan")
	}
	if got := dd.SelectedValue(); got != "Cyan" {
		t.Errorf("SelectedValue after update = %q, want %q", got, "Cyan")
	}
}

func TestItemsReadonlySignal_WinsOverItemsSignal(t *testing.T) {
	writable := state.NewSignal([]string{"writable-a", "writable-b"})
	readonly := state.NewComputed(func() []string {
		return []string{"computed-a", "computed-b"}
	})
	dd := dropdown.New(
		dropdown.Items("static"),
		dropdown.ItemsSignal(writable),
		dropdown.ItemsReadonlySignal(readonly),
		dropdown.Selected(1),
	)

	if got := dd.SelectedValue(); got != "computed-b" {
		t.Errorf("SelectedValue = %q, want %q (readonly wins)", got, "computed-b")
	}
	if got := triggerText(t, dd); got != "computed-b" {
		t.Errorf("trigger text = %q, want %q (readonly wins)", got, "computed-b")
	}

	writable.Set([]string{"writable-c"})

	if got := dd.SelectedValue(); got != "computed-b" {
		t.Errorf("SelectedValue after writable update = %q, want %q", got, "computed-b")
	}
}

func TestItemsFn_ItemsFollowFunction(t *testing.T) {
	items := []string{"one", "two"}
	dd := dropdown.New(
		dropdown.Items("static"),
		dropdown.ItemsFn(func() []string { return items }),
		dropdown.Selected(0),
	)

	if got := dd.SelectedValue(); got != "one" {
		t.Fatalf("SelectedValue = %q, want %q", got, "one")
	}

	items = []string{"three", "four"}

	if got := dd.SelectedValue(); got != "three" {
		t.Errorf("SelectedValue after fn change = %q, want %q", got, "three")
	}
}

func TestItemsSignal_ShrinkClampsSelection(t *testing.T) {
	sig := state.NewSignal([]string{"A", "B", "C", "D"})
	dd := dropdown.New(
		dropdown.ItemsSignal(sig),
		dropdown.Selected(3),
	)

	if got := dd.SelectedIndex(); got != 3 {
		t.Fatalf("SelectedIndex = %d, want 3", got)
	}

	sig.Set([]string{"W", "X"})

	if got := dd.SelectedIndex(); got != 1 {
		t.Errorf("SelectedIndex after shrink = %d, want 1 (last valid)", got)
	}
	if got := dd.SelectedValue(); got != "X" {
		t.Errorf("SelectedValue after shrink = %q, want %q", got, "X")
	}
	if got := triggerText(t, dd); got != "X" {
		t.Errorf("trigger text after shrink = %q, want %q", got, "X")
	}
}

func TestItemsSignal_EmptyListClearsSelection(t *testing.T) {
	sig := state.NewSignal([]string{"A", "B"})
	dd := dropdown.New(
		dropdown.ItemsSignal(sig),
		dropdown.Selected(1),
		dropdown.Placeholder("nothing here"),
	)

	sig.Set(nil)

	if got := dd.SelectedIndex(); got != -1 {
		t.Errorf("SelectedIndex for an empty list = %d, want -1 (no selection)", got)
	}
	if got := dd.SelectedValue(); got != "" {
		t.Errorf("SelectedValue for an empty list = %q, want empty", got)
	}
	if got := triggerText(t, dd); got != "nothing here" {
		t.Errorf("trigger text for an empty list = %q, want %q", got, "nothing here")
	}
}

func TestItemsSignal_GrowKeepsSelection(t *testing.T) {
	sig := state.NewSignal([]string{"A"})
	dd := dropdown.New(
		dropdown.ItemsSignal(sig),
		dropdown.Selected(0),
	)

	sig.Set([]string{"A", "B", "C"})

	if got := dd.SelectedIndex(); got != 0 {
		t.Errorf("SelectedIndex after growth = %d, want 0", got)
	}
	if got := dd.SelectedValue(); got != "A" {
		t.Errorf("SelectedValue after growth = %q, want %q", got, "A")
	}
}

func TestItemsSignal_ChangeMarksWidgetDirty(t *testing.T) {
	sig := state.NewSignal([]string{"A", "B"})
	dd := dropdown.New(
		dropdown.ItemsSignal(sig),
		dropdown.Selected(0),
	)

	sched := state.NewScheduler(func(dirty []widget.Widget) {
		for _, w := range dirty {
			if setter, ok := w.(interface{ SetNeedsRedraw(bool) }); ok {
				setter.SetNeedsRedraw(true)
			}
		}
	})
	ctx := widget.NewContext()
	ctx.SetScheduler(sched)

	dd.Mount(ctx)

	if dd.NeedsRedraw() {
		t.Fatal("widget should be clean before the signal update")
	}

	sig.Set([]string{"A", "B", "C"})
	sched.Flush()

	if !dd.NeedsRedraw() {
		t.Error("items signal update should repaint the trigger")
	}
}

func TestItemsSignal_UnmountReleasesBinding(t *testing.T) {
	sig := state.NewSignal([]string{"A"})
	dd := dropdown.New(dropdown.ItemsSignal(sig))

	sched := state.NewScheduler(func(_ []widget.Widget) {})
	ctx := widget.NewContext()
	ctx.SetScheduler(sched)

	dd.Mount(ctx)
	dd.CleanupBindings()
	dd.Unmount()

	sig.Set([]string{"B"})

	if got := sched.PendingCount(); got != 0 {
		t.Errorf("pending dirty widgets after unmount = %d, want 0", got)
	}
}

func TestItemsReadonlySignal_ChangeMarksWidgetDirty(t *testing.T) {
	deps := state.NewSignal([]string{"A"})
	readonly := state.NewComputed(func() []string {
		return append(deps.Get(), "B")
	}, deps.AsReadonly())
	dd := dropdown.New(dropdown.ItemsReadonlySignal(readonly))

	sched := state.NewScheduler(func(dirty []widget.Widget) {
		for _, w := range dirty {
			if setter, ok := w.(interface{ SetNeedsRedraw(bool) }); ok {
				setter.SetNeedsRedraw(true)
			}
		}
	})
	ctx := widget.NewContext()
	ctx.SetScheduler(sched)

	dd.Mount(ctx)
	deps.Set([]string{"X"})
	sched.Flush()

	if !dd.NeedsRedraw() {
		t.Error("readonly items signal update should repaint the trigger")
	}
}
