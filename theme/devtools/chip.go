package devtools

import (
	"github.com/gogpu/ui/core/chip"
	"github.com/gogpu/ui/widget"
)

// ChipPainter renders chips using DevTools design tokens.
// It maps chip states to the JetBrains Int UI gray scale palette:
// flat style with subtle borders and selection fill.
//
// If Theme is nil, ChipPainter falls back to the default DevTools dark palette.
type ChipPainter struct {
	Theme *Theme // nil uses default DevTools dark fallback
}

// resolveColors returns the ChipColorScheme derived from the painter's Theme.
// If Theme is nil, it returns the default DevTools dark color scheme for chips.
func (p ChipPainter) resolveColors() chip.ChipColorScheme {
	if p.Theme == nil {
		return dtDefaultChipColors
	}
	cs := p.Theme.Colors
	return chip.ChipColorScheme{
		Background:         widget.ColorTransparent,
		Border:             cs.BorderStrong,
		Label:              cs.OnSurface,
		SelectedBackground: cs.ControlFill,
		SelectedLabel:      cs.OnSurface,
		DisabledBackground: cs.ControlFill,
		DisabledLabel:      cs.OnSurfaceDisabled,
	}
}

// PaintChip renders a chip according to DevTools design specifications.
// It resolves DevTools color tokens into a [chip.ChipColorScheme] and delegates
// to [chip.DefaultPainter] for consistent rendering across design systems.
func (p ChipPainter) PaintChip(canvas widget.Canvas, state chip.PaintState) {
	colors := state.ColorScheme
	if colors == (chip.ChipColorScheme{}) {
		colors = p.resolveColors()
	}
	state.ColorScheme = colors
	chip.DefaultPainter{}.PaintChip(canvas, state)
}

// dtDefaultChipColors holds the default DevTools dark color scheme for chips.
var dtDefaultChipColors = chip.ChipColorScheme{
	Background:         widget.ColorTransparent,
	Border:             widget.Hex(0x4E5157), // Gray5
	Label:              widget.Hex(0xDFE1E5), // Gray12
	SelectedBackground: widget.Hex(0x393B40), // Gray3 (control fill)
	SelectedLabel:      widget.Hex(0xDFE1E5), // Gray12
	DisabledBackground: widget.Hex(0x393B40), // Gray3
	DisabledLabel:      widget.Hex(0x6F737A), // Gray7
}

// Compile-time check that ChipPainter implements chip.Painter.
var _ chip.Painter = ChipPainter{}
