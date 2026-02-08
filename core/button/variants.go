package button

// Variant controls the visual style of a button.
type Variant uint8

// Button variant constants.
const (
	// Filled renders a solid-colored button with contrasting text.
	// This is the default and highest-emphasis variant.
	Filled Variant = iota

	// Outlined renders a button with a border and transparent background.
	Outlined

	// TextOnly renders a button with no background or border, only text.
	TextOnly

	// Tonal renders a button with a tinted background (lower emphasis than Filled).
	Tonal
)

// String returns a human-readable name for the variant.
func (v Variant) String() string {
	switch v {
	case Filled:
		return variantFilled
	case Outlined:
		return variantOutlined
	case TextOnly:
		return variantTextOnly
	case Tonal:
		return variantTonal
	default:
		return variantUnknown
	}
}

// String constants for Variant.String to satisfy goconst.
const (
	variantFilled   = "Filled"
	variantOutlined = "Outlined"
	variantTextOnly = "TextOnly"
	variantTonal    = "Tonal"
	variantUnknown  = "Unknown"
)

// Size controls the dimensions of a button.
type Size uint8

// Button size constants.
const (
	// Small renders a compact button with 32px height.
	Small Size = iota

	// Medium renders a standard button with 40px height.
	// This is the default size.
	Medium

	// Large renders a prominent button with 48px height.
	Large
)

// String returns a human-readable name for the size.
func (s Size) String() string {
	switch s {
	case Small:
		return sizeSmall
	case Medium:
		return sizeMedium
	case Large:
		return sizeLarge
	default:
		return variantUnknown
	}
}

// String constants for Size.String to satisfy goconst.
const (
	sizeSmall  = "Small"
	sizeMedium = "Medium"
	sizeLarge  = "Large"
)

// sizeHeight returns the target height in logical pixels for a given button size.
func sizeHeight(s Size) float32 {
	switch s {
	case Small:
		return smallHeight
	case Large:
		return largeHeight
	default:
		return mediumHeight
	}
}

// sizeFontSize returns the font size in logical pixels for a given button size.
func sizeFontSize(s Size) float32 {
	switch s {
	case Small:
		return smallFontSize
	case Large:
		return largeFontSize
	default:
		return mediumFontSize
	}
}

// Height and font size constants for each button size.
const (
	smallHeight  float32 = 32
	mediumHeight float32 = 40
	largeHeight  float32 = 48

	smallFontSize  float32 = 12
	mediumFontSize float32 = 14
	largeFontSize  float32 = 16
)
