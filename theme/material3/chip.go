package material3

import (
	"github.com/gogpu/ui/core/chip"
	"github.com/gogpu/ui/widget"
)

// ChipPainter renders chips using Material 3 design tokens.
// It maps chip states (unselected outlined, selected tonal, disabled)
// to the M3 color scheme and delegates to [chip.DefaultPainter] for
// the actual rendering.
//
// If Theme is nil, ChipPainter falls back to the default M3 purple palette.
type ChipPainter struct {
	Theme *Theme // nil uses default M3 purple fallback
}

// resolveColors returns the ChipColorScheme derived from the painter's Theme.
// If Theme is nil, it returns the default M3 purple color scheme for chips.
func (p ChipPainter) resolveColors() chip.ChipColorScheme {
	if p.Theme == nil {
		return m3DefaultChipColors
	}
	cs := p.Theme.Colors
	return chip.ChipColorScheme{
		Background:         widget.ColorTransparent,
		Border:             cs.Outline,
		Label:              cs.OnSurface,
		SelectedBackground: cs.SecondaryContainer,
		SelectedLabel:      cs.OnSecondaryContainer,
		DisabledBackground: cs.OnSurface.WithAlpha(m3ChipDisabledBgAlpha),
		DisabledLabel:      cs.OnSurface.WithAlpha(m3ChipDisabledFgAlpha),
	}
}

// PaintChip renders a chip according to Material 3 specifications.
// It resolves M3 color tokens into a [chip.ChipColorScheme] and delegates
// to [chip.DefaultPainter] for consistent rendering across design systems.
func (p ChipPainter) PaintChip(canvas widget.Canvas, state chip.PaintState) {
	colors := state.ColorScheme
	if colors == (chip.ChipColorScheme{}) {
		colors = p.resolveColors()
	}
	state.ColorScheme = colors
	chip.DefaultPainter{}.PaintChip(canvas, state)
}

// m3DefaultChipColors holds the default M3 purple color scheme for chips.
// Used as a fallback when no Theme is provided.
var m3DefaultChipColors = chip.ChipColorScheme{
	Background:         widget.ColorTransparent,                              // outlined: no fill
	Border:             widget.Hex(0x79747E),                                 // M3 outline
	Label:              widget.Hex(0x1D1B20),                                 // M3 on-surface (light)
	SelectedBackground: widget.Hex(0xE8DEF8),                                 // M3 secondary container
	SelectedLabel:      widget.Hex(0x1D192B),                                 // M3 on-secondary-container
	DisabledBackground: widget.RGBA(0.12, 0.12, 0.13, m3ChipDisabledBgAlpha), // M3 on-surface @ 12%
	DisabledLabel:      widget.RGBA(0.12, 0.12, 0.13, m3ChipDisabledFgAlpha), // M3 on-surface @ 38%
}

// M3 chip constants.
const (
	m3ChipDisabledBgAlpha float32 = 0.12 // M3 disabled background opacity
	m3ChipDisabledFgAlpha float32 = 0.38 // M3 disabled foreground opacity
)

// Compile-time check that ChipPainter implements chip.Painter.
var _ chip.Painter = ChipPainter{}
