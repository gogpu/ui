package cupertino

import (
	"github.com/gogpu/ui/core/chip"
	"github.com/gogpu/ui/widget"
)

// ChipPainter renders chips using Apple HIG design tokens.
// It maps chip states to the Cupertino color scheme: system blue for
// the selected state (pill-shaped tag), separator for the unselected border.
//
// If Theme is nil, ChipPainter falls back to the default system blue palette.
type ChipPainter struct {
	Theme *Theme // nil uses default system blue fallback
}

// resolveColors returns the ChipColorScheme derived from the painter's Theme.
// If Theme is nil, it returns the default system blue color scheme for chips.
func (p ChipPainter) resolveColors() chip.ChipColorScheme {
	if p.Theme == nil {
		return cupDefaultChipColors
	}
	cs := p.Theme.Colors
	return chip.ChipColorScheme{
		Background:         widget.ColorTransparent,
		Border:             cs.Separator,
		Label:              cs.Label,
		SelectedBackground: cs.Accent.WithAlpha(cupChipSelectedAlpha),
		SelectedLabel:      cs.Accent,
		DisabledBackground: cs.QuaternaryLabel,
		DisabledLabel:      cs.TertiaryLabel,
	}
}

// PaintChip renders a chip according to Apple HIG specifications.
// It resolves Cupertino color tokens into a [chip.ChipColorScheme] and delegates
// to [chip.DefaultPainter] for consistent rendering across design systems.
func (p ChipPainter) PaintChip(canvas widget.Canvas, state chip.PaintState) {
	colors := state.ColorScheme
	if colors == (chip.ChipColorScheme{}) {
		colors = p.resolveColors()
	}
	state.ColorScheme = colors
	chip.DefaultPainter{}.PaintChip(canvas, state)
}

// cupDefaultChipColors holds the default system blue color scheme for chips.
var cupDefaultChipColors = chip.ChipColorScheme{
	Background:         widget.ColorTransparent,
	Border:             widget.RGBA(0.235, 0.235, 0.263, 0.29),     // Separator (light)
	Label:              widget.RGBA(0.0, 0.0, 0.0, 1.0),            // Label (light)
	SelectedBackground: systemBlue.WithAlpha(cupChipSelectedAlpha), // Accent @ 15%
	SelectedLabel:      systemBlue,                                 // Accent
	DisabledBackground: widget.RGBA(0.235, 0.235, 0.263, 0.18),     // QuaternaryLabel
	DisabledLabel:      widget.RGBA(0.235, 0.235, 0.263, 0.3),      // TertiaryLabel
}

// ChipFontSize returns the Cupertino chip font size.
func (ChipPainter) ChipFontSize() float32 { return cupChipFontSize }

// ChipMinWidth returns the Cupertino minimum chip width.
func (ChipPainter) ChipMinWidth() float32 { return cupChipMinWidth }

// ChipPadding returns the Cupertino chip horizontal padding.
func (ChipPainter) ChipPadding() float32 { return cupChipPadding }

// ChipRadius returns the Cupertino chip corner radius.
func (ChipPainter) ChipRadius() float32 { return cupChipRadius }

// Cupertino chip constants.
const (
	cupChipSelectedAlpha float32 = 0.15 // Apple HIG tint alpha for selected tags
	cupChipFontSize      float32 = 15   // SF Pro body
	cupChipMinWidth      float32 = 44   // Apple HIG touch target
	cupChipPadding       float32 = 14   // Apple HIG horizontal padding
	cupChipRadius        float32 = 10   // Apple HIG rounded pill
)

// Compile-time checks.
var (
	_ chip.Painter       = ChipPainter{}
	_ chip.LayoutMetrics = ChipPainter{}
)
