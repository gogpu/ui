package button

import "github.com/gogpu/ui/widget"

// Text sets the button's static display text.
//
// This is a convenience alias for [TextOpt].
func Text(s string) Option {
	return TextOpt(s)
}

// Background sets a custom background color override via option.
//
// This is a convenience alias for [BackgroundOpt].
func Background(color widget.Color) Option {
	return BackgroundOpt(color)
}

// Rounded sets a custom corner radius override via option.
//
// This is a convenience alias for [RoundedOpt].
func Rounded(radius float32) Option {
	return RoundedOpt(radius)
}
