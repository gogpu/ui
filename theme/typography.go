package theme

// Font weight string constants.
const (
	fontWeightThin       = "Thin"
	fontWeightExtraLight = "ExtraLight"
	fontWeightLight      = "Light"
	fontWeightNormal     = "Normal"
	fontWeightMedium     = "Medium"
	fontWeightSemiBold   = "SemiBold"
	fontWeightBold       = "Bold"
	fontWeightExtraBold  = "ExtraBold"
	fontWeightBlack      = "Black"
)

// defaultFontName is the default font family used by typography.
const defaultFontName = "System"

// FontWeight represents the weight (boldness) of a font.
//
// Font weights follow the CSS/OpenType standard where 400 is normal
// and 700 is bold. Not all fonts support all weights; the system will
// use the closest available weight.
type FontWeight uint16

const (
	// FontWeightThin is the thinnest available weight (100).
	FontWeightThin FontWeight = 100

	// FontWeightExtraLight is extra-light weight (200).
	FontWeightExtraLight FontWeight = 200

	// FontWeightLight is light weight (300).
	FontWeightLight FontWeight = 300

	// FontWeightNormal is the standard weight (400).
	FontWeightNormal FontWeight = 400

	// FontWeightMedium is medium weight (500).
	FontWeightMedium FontWeight = 500

	// FontWeightSemiBold is semi-bold weight (600).
	FontWeightSemiBold FontWeight = 600

	// FontWeightBold is bold weight (700).
	FontWeightBold FontWeight = 700

	// FontWeightExtraBold is extra-bold weight (800).
	FontWeightExtraBold FontWeight = 800

	// FontWeightBlack is the heaviest weight (900).
	FontWeightBlack FontWeight = 900
)

// String returns the human-readable name for the font weight.
func (w FontWeight) String() string {
	switch w {
	case FontWeightThin:
		return fontWeightThin
	case FontWeightExtraLight:
		return fontWeightExtraLight
	case FontWeightLight:
		return fontWeightLight
	case FontWeightNormal:
		return fontWeightNormal
	case FontWeightMedium:
		return fontWeightMedium
	case FontWeightSemiBold:
		return fontWeightSemiBold
	case FontWeightBold:
		return fontWeightBold
	case FontWeightExtraBold:
		return fontWeightExtraBold
	case FontWeightBlack:
		return fontWeightBlack
	default:
		return unknownStr
	}
}

// IsBold returns true if this weight is considered bold (>= 700).
func (w FontWeight) IsBold() bool {
	return w >= FontWeightBold
}

// IsLight returns true if this weight is considered light (<= 300).
func (w FontWeight) IsLight() bool {
	return w <= FontWeightLight
}

// FontStyle represents the style (posture) of a font.
type FontStyle uint8

const (
	// FontStyleNormal is the standard upright style.
	FontStyleNormal FontStyle = iota

	// FontStyleItalic is the italic style with modified letterforms.
	FontStyleItalic

	// FontStyleOblique is a slanted version of the normal style.
	FontStyleOblique
)

// String returns the human-readable name for the font style.
func (s FontStyle) String() string {
	switch s {
	case FontStyleNormal:
		return "Normal"
	case FontStyleItalic:
		return "Italic"
	case FontStyleOblique:
		return "Oblique"
	default:
		return unknownStr
	}
}

// TextStyle defines the complete styling for text rendering.
//
// TextStyle combines all properties needed to render text: font family,
// size, weight, style, line height, and letter spacing. It follows
// Material 3 typography conventions.
//
// Example:
//
//	style := theme.TextStyle{
//	    Font:          "Roboto",
//	    Size:          16,
//	    Weight:        theme.FontWeightNormal,
//	    Style:         theme.FontStyleNormal,
//	    LineHeight:    24,
//	    LetterSpacing: 0,
//	}
type TextStyle struct {
	// Font is the font family name (e.g., "Roboto", "Inter", "System").
	//
	// If the font is not available, the system will fall back to a
	// default sans-serif font.
	Font string

	// Size is the font size in logical pixels (sp in Material terms).
	Size float32

	// Weight is the font weight (boldness).
	Weight FontWeight

	// Style is the font style (normal, italic, oblique).
	Style FontStyle

	// LineHeight is the line height in logical pixels.
	//
	// This defines the vertical spacing between baselines. If zero,
	// a default line height is calculated from the font size.
	LineHeight float32

	// LetterSpacing is the additional spacing between characters in logical pixels.
	//
	// Positive values increase spacing, negative values decrease it.
	// Zero means normal spacing.
	LetterSpacing float32
}

// WithSize returns a copy of the style with a different size.
func (t TextStyle) WithSize(size float32) TextStyle {
	t.Size = size
	return t
}

// WithWeight returns a copy of the style with a different weight.
func (t TextStyle) WithWeight(weight FontWeight) TextStyle {
	t.Weight = weight
	return t
}

// WithStyle returns a copy of the style with a different font style.
func (t TextStyle) WithStyle(style FontStyle) TextStyle {
	t.Style = style
	return t
}

// WithFont returns a copy of the style with a different font family.
func (t TextStyle) WithFont(font string) TextStyle {
	t.Font = font
	return t
}

// WithLineHeight returns a copy of the style with a different line height.
func (t TextStyle) WithLineHeight(lineHeight float32) TextStyle {
	t.LineHeight = lineHeight
	return t
}

// WithLetterSpacing returns a copy of the style with different letter spacing.
func (t TextStyle) WithLetterSpacing(letterSpacing float32) TextStyle {
	t.LetterSpacing = letterSpacing
	return t
}

// Bold returns a copy of the style with bold weight.
func (t TextStyle) Bold() TextStyle {
	t.Weight = FontWeightBold
	return t
}

// Italic returns a copy of the style with italic style.
func (t TextStyle) Italic() TextStyle {
	t.Style = FontStyleItalic
	return t
}

// Typography defines the complete type scale for a theme.
//
// Typography follows the Material 3 type scale with five categories:
//   - Display: Large, impactful text for hero sections
//   - Headline: Section headers
//   - Title: Smaller headings and component titles
//   - Body: Primary reading text
//   - Label: UI labels, buttons, and small text
//
// Each category has Large, Medium, and Small variants for flexibility.
//
// The FontFamily field sets the default font for all styles. Individual
// styles can override this by setting their Font field.
type Typography struct {
	// FontFamily is the default font family used by all text styles.
	//
	// Common values: "Roboto", "Inter", "System", "sans-serif"
	FontFamily string

	// Display styles for large, impactful text.
	//
	// Use for hero sections, page titles, or other prominent text.
	// Display text should be used sparingly for maximum impact.
	DisplayLarge  TextStyle
	DisplayMedium TextStyle
	DisplaySmall  TextStyle

	// Headline styles for section headers.
	//
	// Use for major sections or content divisions.
	HeadlineLarge  TextStyle
	HeadlineMedium TextStyle
	HeadlineSmall  TextStyle

	// Title styles for smaller headings.
	//
	// Use for component titles, card headers, or subsections.
	TitleLarge  TextStyle
	TitleMedium TextStyle
	TitleSmall  TextStyle

	// Body styles for primary reading text.
	//
	// Use for paragraphs, descriptions, and general content.
	BodyLarge  TextStyle
	BodyMedium TextStyle
	BodySmall  TextStyle

	// Label styles for UI labels and buttons.
	//
	// Use for button text, form labels, captions, and small UI text.
	LabelLarge  TextStyle
	LabelMedium TextStyle
	LabelSmall  TextStyle
}

// newTextStyle creates a new TextStyle with the given parameters.
func newTextStyle(size, lineHeight, letterSpacing float32, weight FontWeight) TextStyle {
	return TextStyle{
		Font:          defaultFontName,
		Size:          size,
		Weight:        weight,
		Style:         FontStyleNormal,
		LineHeight:    lineHeight,
		LetterSpacing: letterSpacing,
	}
}

// defaultDisplayStyles returns the display text styles for Material 3.
func defaultDisplayStyles() (large, medium, small TextStyle) {
	large = newTextStyle(57, 64, -0.25, FontWeightNormal)
	medium = newTextStyle(45, 52, 0, FontWeightNormal)
	small = newTextStyle(36, 44, 0, FontWeightNormal)
	return
}

// defaultHeadlineStyles returns the headline text styles for Material 3.
func defaultHeadlineStyles() (large, medium, small TextStyle) {
	large = newTextStyle(32, 40, 0, FontWeightNormal)
	medium = newTextStyle(28, 36, 0, FontWeightNormal)
	small = newTextStyle(24, 32, 0, FontWeightNormal)
	return
}

// defaultTitleStyles returns the title text styles for Material 3.
func defaultTitleStyles() (large, medium, small TextStyle) {
	large = newTextStyle(22, 28, 0, FontWeightNormal)
	medium = newTextStyle(16, 24, 0.15, FontWeightMedium)
	small = newTextStyle(14, 20, 0.1, FontWeightMedium)
	return
}

// defaultBodyStyles returns the body text styles for Material 3.
func defaultBodyStyles() (large, medium, small TextStyle) {
	large = newTextStyle(16, 24, 0.5, FontWeightNormal)
	medium = newTextStyle(14, 20, 0.25, FontWeightNormal)
	small = newTextStyle(12, 16, 0.4, FontWeightNormal)
	return
}

// defaultLabelStyles returns the label text styles for Material 3.
func defaultLabelStyles() (large, medium, small TextStyle) {
	large = newTextStyle(14, 20, 0.1, FontWeightMedium)
	medium = newTextStyle(12, 16, 0.5, FontWeightMedium)
	small = newTextStyle(11, 16, 0.5, FontWeightMedium)
	return
}

// DefaultTypography returns a Typography scale following Material 3 guidelines.
//
// The default uses "System" as the font family, which maps to the platform's
// default UI font (Segoe UI on Windows, San Francisco on macOS, Roboto on Linux).
//
// All sizes are in logical pixels (sp). Line heights and letter spacing follow
// Material 3 recommendations for optimal readability.
func DefaultTypography() Typography {
	displayL, displayM, displayS := defaultDisplayStyles()
	headlineL, headlineM, headlineS := defaultHeadlineStyles()
	titleL, titleM, titleS := defaultTitleStyles()
	bodyL, bodyM, bodyS := defaultBodyStyles()
	labelL, labelM, labelS := defaultLabelStyles()

	return Typography{
		FontFamily:     defaultFontName,
		DisplayLarge:   displayL,
		DisplayMedium:  displayM,
		DisplaySmall:   displayS,
		HeadlineLarge:  headlineL,
		HeadlineMedium: headlineM,
		HeadlineSmall:  headlineS,
		TitleLarge:     titleL,
		TitleMedium:    titleM,
		TitleSmall:     titleS,
		BodyLarge:      bodyL,
		BodyMedium:     bodyM,
		BodySmall:      bodyS,
		LabelLarge:     labelL,
		LabelMedium:    labelM,
		LabelSmall:     labelS,
	}
}

// WithFontFamily returns a copy of the typography with all styles using
// the specified font family.
func (t *Typography) WithFontFamily(fontFamily string) Typography {
	result := *t
	result.FontFamily = fontFamily
	result.DisplayLarge.Font = fontFamily
	result.DisplayMedium.Font = fontFamily
	result.DisplaySmall.Font = fontFamily
	result.HeadlineLarge.Font = fontFamily
	result.HeadlineMedium.Font = fontFamily
	result.HeadlineSmall.Font = fontFamily
	result.TitleLarge.Font = fontFamily
	result.TitleMedium.Font = fontFamily
	result.TitleSmall.Font = fontFamily
	result.BodyLarge.Font = fontFamily
	result.BodyMedium.Font = fontFamily
	result.BodySmall.Font = fontFamily
	result.LabelLarge.Font = fontFamily
	result.LabelMedium.Font = fontFamily
	result.LabelSmall.Font = fontFamily
	return result
}

// Scale returns a copy of the typography with all sizes scaled by the given factor.
//
// This is useful for accessibility settings that increase text size.
// A factor of 1.0 returns the original sizes, 1.25 increases all sizes by 25%, etc.
func (t *Typography) Scale(factor float32) Typography {
	scaleStyle := func(s TextStyle) TextStyle {
		s.Size *= factor
		s.LineHeight *= factor
		return s
	}

	result := *t
	result.DisplayLarge = scaleStyle(t.DisplayLarge)
	result.DisplayMedium = scaleStyle(t.DisplayMedium)
	result.DisplaySmall = scaleStyle(t.DisplaySmall)
	result.HeadlineLarge = scaleStyle(t.HeadlineLarge)
	result.HeadlineMedium = scaleStyle(t.HeadlineMedium)
	result.HeadlineSmall = scaleStyle(t.HeadlineSmall)
	result.TitleLarge = scaleStyle(t.TitleLarge)
	result.TitleMedium = scaleStyle(t.TitleMedium)
	result.TitleSmall = scaleStyle(t.TitleSmall)
	result.BodyLarge = scaleStyle(t.BodyLarge)
	result.BodyMedium = scaleStyle(t.BodyMedium)
	result.BodySmall = scaleStyle(t.BodySmall)
	result.LabelLarge = scaleStyle(t.LabelLarge)
	result.LabelMedium = scaleStyle(t.LabelMedium)
	result.LabelSmall = scaleStyle(t.LabelSmall)

	return result
}
