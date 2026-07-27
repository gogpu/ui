// Example: gogpu/ui — File Drop Demo
//
// Demonstrates OS file drag-and-drop integration with the gogpu/ui widget
// toolkit. Files dragged from Explorer/Finder/Nautilus are received through
// the dnd.Manager and displayed as a reactive list.
//
// Key concepts shown:
//   - dnd.DropTarget registration with hit-testing bounds
//   - dnd.KindFile + dnd.FilePayload for OS file drops
//   - state.Signal for reactive UI updates
//   - primitives.TextFn for computed text display
//
// Architecture:
//
//	OS file drop → gogpu.OnDragDrop → desktop.Run bridge → dnd.Manager
//	→ DropTarget.Drop → Signal update → reactive text display
//
// Rendering: event-driven (default since gogpu v0.43.0).
// 0% CPU when idle. Redraws only on user interaction or file drop.
package main

import (
	"fmt"
	"log"
	"strings"

	_ "github.com/gogpu/gg/gpu" // enable GPU SDF acceleration

	"github.com/gogpu/gogpu"
	"github.com/gogpu/ui/app"
	"github.com/gogpu/ui/core/button"
	"github.com/gogpu/ui/desktop"
	"github.com/gogpu/ui/dnd"
	"github.com/gogpu/ui/geometry"
	"github.com/gogpu/ui/primitives"
	"github.com/gogpu/ui/state"
	"github.com/gogpu/ui/theme/material3"
	"github.com/gogpu/ui/widget"
)

func main() {
	m3 := material3.New(widget.Hex(0x6750A4))

	gogpuApp := gogpu.NewApp(gogpu.DefaultConfig().
		WithTitle("gogpu/ui — File Drop Demo").
		WithSize(600, 500))

	uiApp := app.New(
		app.WithWindowProvider(gogpuApp),
		app.WithPlatformProvider(gogpuApp),
		app.WithEventSource(gogpuApp.EventSource()),
		app.WithTheme(m3.AsTheme()),
	)

	// Signal holds the list of dropped file paths.
	droppedFiles := state.NewSignal[[]string](nil)

	uiApp.SetRoot(buildUI(droppedFiles))

	// Register a DropTarget on the dnd.Manager for file drops.
	// The bounds cover the full window; desktop.Run bridges OS drops
	// to dnd.Manager.DropExternal which hit-tests against this target.
	target := &fileDropTarget{files: droppedFiles}
	mgr := uiApp.Window().DndManager()
	mgr.RegisterTarget(target, geometry.NewRect(0, 0, 600, 500))

	if err := desktop.Run(gogpuApp, uiApp); err != nil {
		log.Fatal(err)
	}
}

func buildUI(files state.Signal[[]string]) *primitives.BoxWidget {
	card := primitives.Box(
		primitives.Text("File Drop Demo").
			FontSize(24).
			Bold().
			Color(widget.RGBA8(33, 33, 33, 255)),

		// Drop zone with distinct background.
		primitives.Box(
			primitives.Text("Drag files from your file manager here").
				FontSize(14).
				Color(widget.RGBA8(120, 120, 120, 255)),
		).
			Padding(40).
			Background(widget.RGBA8(245, 245, 245, 255)).
			Rounded(12).
			BorderStyle(2, widget.RGBA8(180, 180, 220, 255)),

		// Reactive display of dropped file paths.
		primitives.TextFn(func() string {
			paths := files.Get()
			if len(paths) == 0 {
				return "No files dropped yet."
			}
			return fmt.Sprintf("Dropped %d file(s):\n%s", len(paths), strings.Join(paths, "\n"))
		}).
			FontSize(13).
			Color(widget.RGBA8(50, 50, 50, 255)),

		// Clear button.
		button.New(
			button.TextOpt("Clear"),
			button.OnClick(func() {
				files.Set(nil)
			}),
		),
	).
		Padding(32).
		Gap(16).
		Background(widget.RGBA8(255, 255, 255, 255)).
		Rounded(12).
		ShadowLevel(2)

	return primitives.Box(card).Padding(24)
}

// fileDropTarget implements dnd.DropTarget to accept OS file drops.
type fileDropTarget struct {
	files state.Signal[[]string]
}

func (t *fileDropTarget) CanAccept(data dnd.DragData) bool {
	return data.Kind == dnd.KindFile
}

func (t *fileDropTarget) DragEnter(_ dnd.DragData) {}

func (t *fileDropTarget) DragOver(_ dnd.DragData, _ geometry.Point) dnd.DropEffect {
	return dnd.DropCopy
}

func (t *fileDropTarget) DragLeave() {}

func (t *fileDropTarget) Drop(data dnd.DragData, _ geometry.Point) bool {
	payload, ok := data.Payload.(dnd.FilePayload)
	if !ok || len(payload.Paths) == 0 {
		return false
	}

	// Append to existing files.
	existing := t.files.Get()
	t.files.Set(append(existing, payload.Paths...))
	fmt.Printf("Received %d file(s)\n", len(payload.Paths))
	return true
}
