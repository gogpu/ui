package material3

// TypeScale holds the 15 Material 3 typography roles.
//
// Material 3 organizes typography into five categories (Display, Headline,
// Title, Body, Label), each with three sizes (Large, Medium, Small).
// This provides a complete type scale for building consistent UIs.
type TypeScale struct {
	DisplayLarge  TextStyle
	DisplayMedium TextStyle
	DisplaySmall  TextStyle

	HeadlineLarge  TextStyle
	HeadlineMedium TextStyle
	HeadlineSmall  TextStyle

	TitleLarge  TextStyle
	TitleMedium TextStyle
	TitleSmall  TextStyle

	BodyLarge  TextStyle
	BodyMedium TextStyle
	BodySmall  TextStyle

	LabelLarge  TextStyle
	LabelMedium TextStyle
	LabelSmall  TextStyle
}

// TextStyle defines font properties for a typography role.
//
// This is a simplified text style containing the essential properties
// needed for M3 typography: font size, line height, and weight indicator.
type TextStyle struct {
	// FontSize is the font size in logical pixels (sp).
	FontSize float32

	// LineHeight is the line height in logical pixels.
	LineHeight float32

	// Bold indicates whether the text should be rendered in bold weight.
	// In Material 3, "bold" typically maps to Medium (500) weight.
	Bold bool
}

// DefaultTypeScale returns the standard Material 3 type scale.
//
// Font sizes and line heights follow the M3 specification:
// https://m3.material.io/styles/typography/type-scale-tokens
func DefaultTypeScale() TypeScale {
	return TypeScale{
		// Display: large, impactful text for hero sections.
		DisplayLarge:  TextStyle{FontSize: 57, LineHeight: 64, Bold: false},
		DisplayMedium: TextStyle{FontSize: 45, LineHeight: 52, Bold: false},
		DisplaySmall:  TextStyle{FontSize: 36, LineHeight: 44, Bold: false},

		// Headline: section headers and content divisions.
		HeadlineLarge:  TextStyle{FontSize: 32, LineHeight: 40, Bold: false},
		HeadlineMedium: TextStyle{FontSize: 28, LineHeight: 36, Bold: false},
		HeadlineSmall:  TextStyle{FontSize: 24, LineHeight: 32, Bold: false},

		// Title: component titles and card headers.
		TitleLarge:  TextStyle{FontSize: 22, LineHeight: 28, Bold: false},
		TitleMedium: TextStyle{FontSize: 16, LineHeight: 24, Bold: true},
		TitleSmall:  TextStyle{FontSize: 14, LineHeight: 20, Bold: true},

		// Body: primary reading text and descriptions.
		BodyLarge:  TextStyle{FontSize: 16, LineHeight: 24, Bold: false},
		BodyMedium: TextStyle{FontSize: 14, LineHeight: 20, Bold: false},
		BodySmall:  TextStyle{FontSize: 12, LineHeight: 16, Bold: false},

		// Label: UI labels, buttons, and small text.
		LabelLarge:  TextStyle{FontSize: 14, LineHeight: 20, Bold: true},
		LabelMedium: TextStyle{FontSize: 12, LineHeight: 16, Bold: true},
		LabelSmall:  TextStyle{FontSize: 11, LineHeight: 16, Bold: true},
	}
}
