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

// Fluent chip constants.
const (
	flChipDisabledAlpha float32 = 0.38
)

// Compile-time check that ChipPainter implements chip.Painter.
var _ chip.Painter = ChipPainter{}
