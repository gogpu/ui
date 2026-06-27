package fluent

import (
	"github.com/gogpu/ui/core/chip"
	"github.com/gogpu/ui/widget"
)

// ChipPainter renders chips using Fluent Design tokens.
// It maps chip states to the Fluent color scheme: accent color for
// the selected state, subtle stroke for the unselected outline.
//
// If Theme is nil, ChipPainter falls back to the default Fluent Blue palette.
type ChipPainter struct {
	Theme *Theme // nil uses default Fluent Blue fallback
}

// resolveColors returns the ChipColorScheme derived from the painter's Theme.
// If Theme is nil, it returns the default Fluent Blue color scheme for chips.
func (p ChipPainter) resolveColors() chip.ChipColorScheme {
	if p.Theme == nil {
		return flDefaultChipColors
	}
	cs := p.Theme.Colors
	return chip.ChipColorScheme{
		Background:         widget.ColorTransparent,
		Border:             cs.StrokeDefault,
		Label:              cs.OnSurface,
		SelectedBackground: cs.AccentLight,
		SelectedLabel:      cs.AccentDark,
		DisabledBackground: cs.FillDisable,
		DisabledLabel:      cs.OnSurfaceSecond.WithAlpha(flChipDisabledAlpha),
	}
}

// PaintChip renders a chip according to Fluent Design specifications.
// It resolves Fluent color tokens into a [chip.ChipColorScheme] and delegates
// to [chip.DefaultPainter] for consistent rendering across design systems.
func (p ChipPainter) PaintChip(canvas widget.Canvas, state chip.PaintState) {
	colors := state.ColorScheme
	if colors == (chip.ChipColorScheme{}) {
		colors = p.resolveColors()
	}
	state.ColorScheme = colors
	chip.DefaultPainter{}.PaintChip(canvas, state)
}

// flDefaultChipColors holds the default Fluent Blue color scheme for chips.
var flDefaultChipColors = chip.ChipColorScheme{
	Background:         widget.ColorTransparent,
	Border:             widget.RGBA(0, 0, 0, 0.14),                         // StrokeDefault light
	Label:              widget.Hex(0x1A1A1A),                               // OnSurface light
	SelectedBackground: lighten(DefaultAccentColor, 0.85),                  // AccentLight
	SelectedLabel:      darken(DefaultAccentColor, 0.25),                   // AccentDark
	DisabledBackground: widget.RGBA(0, 0, 0, 0.04),                         // FillDisable
	DisabledLabel:      widget.RGBA(0.38, 0.38, 0.38, flChipDisabledAlpha), // OnSurfaceSecond @ 38%
}

// ChipFontSize returns the Fluent chip font size.
func (ChipPainter) ChipFontSize() float32 { return flChipFontSize }

// ChipMinWidth returns the Fluent minimum chip width.
func (ChipPainter) ChipMinWidth() float32 { return flChipMinWidth }

// ChipPadding returns the Fluent chip horizontal padding.
func (ChipPainter) ChipPadding() float32 { return flChipPadding }

// ChipRadius returns the Fluent chip corner radius.
func (ChipPainter) ChipRadius() float32 { return flChipRadius }

// Fluent chip constants.
const (
	flChipDisabledAlpha float32 = 0.38
	flChipFontSize      float32 = 14 // Fluent body
	flChipMinWidth      float32 = 44 // Fluent touch target
	flChipPadding       float32 = 12 // Fluent horizontal padding
	flChipRadius        float32 = 4  // Fluent control radius
)

// Compile-time checks.
var (
	_ chip.Painter       = ChipPainter{}
	_ chip.LayoutMetrics = ChipPainter{}
)
