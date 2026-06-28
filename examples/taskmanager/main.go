// Example: gogpu/ui — Task Manager Performance Monitor
//
// A flagship demo inspired by Windows Task Manager's Performance tab.
// Showcases real-time line charts, progress bars, collapsible sections,
// horizontal/vertical box layout, and simulated live data updates.
//
// Architecture:
//
//	ui widgets -> render.Canvas (gg) -> ggcanvas -> GPU surface (zero-copy)
//
// Rendering: event-driven with periodic data updates.
// Redraws only when data changes or user interacts.
//
// Requirements:
//
//	gogpu v0.23.2+
//	gg v0.35.3+
package main

import (
	"fmt"
	"log"
	"math"
	"math/rand/v2"
	"time"

	_ "github.com/gogpu/gg/gpu" // enable GPU SDF acceleration

	"github.com/gogpu/gogpu"
	"github.com/gogpu/ui/app"
	"github.com/gogpu/ui/core/collapsible"
	"github.com/gogpu/ui/core/linechart"
	"github.com/gogpu/ui/core/progressbar"
	"github.com/gogpu/ui/desktop"
	"github.com/gogpu/ui/primitives"
	"github.com/gogpu/ui/state"
	"github.com/gogpu/ui/theme"
	"github.com/gogpu/ui/widget"
)

// --- Color palette (dark theme, inspired by Windows Task Manager) ---

var (
	colorBackground    = widget.Hex(0x1E1E1E) // main background
	colorSurface       = widget.Hex(0x252526) // card/section background
	colorSurfaceLight  = widget.Hex(0x2D2D2D) // header background
	colorTextPrimary   = widget.RGBA8(230, 230, 230, 255)
	colorTextSecondary = widget.RGBA8(160, 160, 160, 255)
	colorTextMuted     = widget.RGBA8(120, 120, 120, 255)
	colorCPU           = widget.Hex(0x0078D7) // blue
	colorMemory        = widget.Hex(0x886CE4) // purple
	colorDisk          = widget.Hex(0x16C60C) // green
	colorGridLine      = widget.RGBA8(60, 60, 65, 255)
	colorChartBg       = widget.RGBA8(20, 20, 25, 255)
	colorTrackDark     = widget.RGBA8(60, 60, 65, 255)
)

// simState holds the simulated system metrics updated by a background goroutine.
type simState struct {
	cpuChart  *linechart.Widget
	diskChart *linechart.Widget
	memBar    *progressbar.Widget

	// Reactive signals for collapsible headers.
	cpuTitle  state.Signal[string]
	memTitle  state.Signal[string]
	diskTitle state.Signal[string]

	// Current values for data simulation.
	cpuPercent  float64
	memUsedGB   float64
	memTotalGB  float64
	diskReadMB  float64
	diskWriteMB float64
	processes   int
	threads     int
	cpuSpeed    float64

	// Smoothing state for realistic data simulation.
	cpuSmooth  float64
	memSmooth  float64
	diskSmooth float64
}

func main() {
	// Dark theme with Task Manager background color (#1E1E1E).
	darkTheme := theme.DefaultDark()
	darkTheme.Colors.Background = colorBackground

	gogpuApp := gogpu.NewApp(gogpu.DefaultConfig().
		WithTitle("gogpu/ui — Task Manager").
		WithSize(700, 800))

	uiApp := app.New(
		app.WithWindowProvider(gogpuApp),
		app.WithPlatformProvider(gogpuApp),
		app.WithEventSource(gogpuApp.EventSource()),
		app.WithTheme(darkTheme),
	)

	sim := &simState{
		cpuTitle:   state.NewSignal("CPU  0%"),
		memTitle:   state.NewSignal(fmt.Sprintf("Memory  %.1f/%.0f GB", 0.51*16.0, 16.0)),
		diskTitle:  state.NewSignal("Disk 0 (C:)  0%"),
		memTotalGB: 16.0,
		memSmooth:  0.51,
		cpuSmooth:  35.0,
		diskSmooth: 15.0,
		processes:  234,
		threads:    3456,
		cpuSpeed:   3.4,
	}

	root := buildUI(sim)
	uiApp.SetRoot(root)

	// Start simulated data producer in a background goroutine.
	go runSimulation(sim, gogpuApp)

	if err := desktop.Run(gogpuApp, uiApp); err != nil {
		log.Fatal(err)
	}
}

// buildUI constructs the full widget tree.
func buildUI(sim *simState) *primitives.BoxWidget {
	cpuSection := buildCPUSection(sim)
	memSection := buildMemorySection(sim)
	diskSection := buildDiskSection(sim)

	// Main container: vertical stack of collapsible sections.
	content := primitives.VBox(
		// Title bar.
		primitives.Text("Performance").
			FontSize(22).
			Bold().
			Color(colorTextPrimary),

		cpuSection,
		memSection,
		diskSection,
	).
		Padding(20).
		Gap(16)

	return content
}

// buildCPUSection creates the CPU monitoring section.
func buildCPUSection(sim *simState) *collapsible.Widget {
	// CPU line chart.
	sim.cpuChart = linechart.New(
		linechart.MaxPoints(60),
		linechart.YRange(0, 100),
		linechart.ShowGrid(true),
		linechart.ShowLabels(true),
		linechart.GridColor(colorGridLine),
		linechart.BackgroundColor(colorChartBg),
	)
	sim.cpuChart.AddSeries("CPU", colorCPU)

	// Stats row.
	statsRow := primitives.HBox(
		statLabel("Speed", fmt.Sprintf("%.1f GHz", sim.cpuSpeed)),
		statLabel("Processes", fmt.Sprintf("%d", sim.processes)),
		statLabel("Threads", fmt.Sprintf("%d", sim.threads)),
	).Gap(32).PaddingXY(0, 8)

	// Content: chart + stats.
	content := primitives.VBox(
		sim.cpuChart,
		statsRow,
	).Gap(8).Padding(12).Background(colorSurface).Rounded(6)

	return collapsible.New(
		collapsible.TitleSignal(sim.cpuTitle),
		collapsible.Content(content),
		collapsible.Expanded(true),
		collapsible.Animated(true),
		collapsible.HeaderHeight(40),
		collapsible.HeaderColor(colorSurfaceLight),
		collapsible.ArrowColor(colorTextSecondary),
	)
}

// buildMemorySection creates the Memory monitoring section.
func buildMemorySection(sim *simState) *collapsible.Widget {
	// Memory progress bar with purple color scheme.
	sim.memBar = progressbar.New(
		progressbar.Value(sim.memSmooth/sim.memTotalGB),
		progressbar.ShowLabel(true),
		progressbar.Height(24),
		progressbar.Radius(4),
		progressbar.FormatLabelFn(func(v float64) string {
			return fmt.Sprintf("%.1f / %.1f GB (%.0f%%)", v*sim.memTotalGB, sim.memTotalGB, v*100)
		}),
		progressbar.ColorSchemeOpt(progressbar.ProgressBarColorScheme{
			Bar:   colorMemory,
			Track: colorTrackDark,
			Label: colorTextPrimary,
		}),
	)

	// Stats row.
	availGB := sim.memTotalGB - sim.memSmooth*sim.memTotalGB
	statsRow := primitives.HBox(
		statLabel("Available", fmt.Sprintf("%.1f GB", availGB)),
		statLabel("Cached", "4.2 GB"),
		statLabel("Committed", "9.8 / 24.0 GB"),
	).Gap(32).PaddingXY(0, 8)

	content := primitives.VBox(
		sim.memBar,
		statsRow,
	).Gap(8).Padding(12).Background(colorSurface).Rounded(6)

	return collapsible.New(
		collapsible.TitleSignal(sim.memTitle),
		collapsible.Content(content),
		collapsible.Expanded(true),
		collapsible.Animated(true),
		collapsible.HeaderHeight(40),
		collapsible.HeaderColor(colorSurfaceLight),
		collapsible.ArrowColor(colorTextSecondary),
	)
}

// buildDiskSection creates the Disk monitoring section.
func buildDiskSection(sim *simState) *collapsible.Widget {
	sim.diskChart = linechart.New(
		linechart.MaxPoints(60),
		linechart.YRange(0, 100),
		linechart.ShowGrid(true),
		linechart.ShowLabels(true),
		linechart.GridColor(colorGridLine),
		linechart.BackgroundColor(colorChartBg),
	)
	sim.diskChart.AddSeries("Read", colorDisk)
	sim.diskChart.AddSeries("Write", widget.Hex(0xF9F1A5)) // yellow for write

	// Stats row.
	statsRow := primitives.HBox(
		statLabel("Read", "0 MB/s"),
		statLabel("Write", "0 MB/s"),
		statLabel("Active time", "0%"),
	).Gap(32).PaddingXY(0, 8)

	content := primitives.VBox(
		sim.diskChart,
		statsRow,
	).Gap(8).Padding(12).Background(colorSurface).Rounded(6)

	return collapsible.New(
		collapsible.TitleSignal(sim.diskTitle),
		collapsible.Content(content),
		collapsible.Expanded(true),
		collapsible.Animated(true),
		collapsible.HeaderHeight(40),
		collapsible.HeaderColor(colorSurfaceLight),
		collapsible.ArrowColor(colorTextSecondary),
	)
}

// statLabel creates a small stat display with label and value stacked vertically.
func statLabel(label, value string) *primitives.BoxWidget {
	return primitives.VBox(
		primitives.Text(label).FontSize(11).Color(colorTextMuted),
		primitives.Text(value).FontSize(13).Color(colorTextSecondary),
	).Gap(2)
}

// --- Data Simulation ---

// runSimulation generates fake system metrics at 1 Hz and pushes them
// to the chart/bar widgets. It calls RequestRedraw to trigger a repaint.
func runSimulation(sim *simState, gogpuApp *gogpu.App) {
	ticker := time.NewTicker(1 * time.Second)
	defer ticker.Stop()

	for range ticker.C {
		// CPU: smooth random walk oscillating 15-85%.
		sim.cpuSmooth = smoothWalk(sim.cpuSmooth, 15, 85, 8)
		sim.cpuPercent = sim.cpuSmooth
		sim.cpuChart.PushValue("CPU", sim.cpuSmooth)

		// Memory: slow drift between 40-75%.
		sim.memSmooth = smoothWalk(sim.memSmooth, 0.40, 0.75, 0.02)
		sim.memUsedGB = sim.memSmooth * sim.memTotalGB
		sim.memBar.SetValue(sim.memSmooth)

		// Disk: spiky, sometimes idle.
		sim.diskSmooth = spikyWalk(sim.diskSmooth, 0, 100, 20)
		diskWrite := spikyWalk(sim.diskWriteMB, 0, 60, 15)
		sim.diskReadMB = sim.diskSmooth
		sim.diskWriteMB = diskWrite
		sim.diskChart.PushValue("Read", sim.diskSmooth)
		sim.diskChart.PushValue("Write", diskWrite)

		// Update reactive headers.
		sim.cpuTitle.Set(fmt.Sprintf("CPU  %.0f%%", sim.cpuPercent))
		sim.memTitle.Set(fmt.Sprintf("Memory  %.1f/%.0f GB", sim.memUsedGB, sim.memTotalGB))
		sim.diskTitle.Set(fmt.Sprintf("Disk 0 (C:)  %.0f%%", sim.diskSmooth))

		// Jitter process/thread counts slightly.
		sim.processes += rand.IntN(5) - 2
		if sim.processes < 200 {
			sim.processes = 200
		}
		sim.threads += rand.IntN(20) - 10
		if sim.threads < 3000 {
			sim.threads = 3000
		}

		gogpuApp.RequestRedraw()
	}
}

// smoothWalk produces a smooth random walk clamped to [lo, hi].
// step controls the maximum change per tick.
func smoothWalk(current, lo, hi, step float64) float64 {
	delta := (rand.Float64()*2 - 1) * step
	// Add mean-reversion toward the midpoint.
	mid := (lo + hi) / 2
	delta += (mid - current) * 0.05
	next := current + delta
	return math.Max(lo, math.Min(hi, next))
}

// spikyWalk produces data with occasional spikes (for disk-like patterns).
func spikyWalk(current, lo, hi, maxStep float64) float64 {
	// 20% chance of a spike toward a random value.
	if rand.Float64() < 0.2 {
		target := lo + rand.Float64()*(hi-lo)
		return math.Max(lo, math.Min(hi, current+(target-current)*0.6))
	}
	// Otherwise decay toward low.
	decay := current * 0.85
	delta := (rand.Float64()*2 - 1) * maxStep * 0.3
	next := decay + delta
	return math.Max(lo, math.Min(hi, next))
}
