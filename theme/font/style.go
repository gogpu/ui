package font

// Style represents the style (posture) of a font.
type Style uint8

const (
	// Normal is the standard upright style.
	Normal Style = iota

	// Italic is the italic style with modified letterforms.
	Italic
)

// styleNames maps Style values to their human-readable names.
var styleNames = [...]string{
	Normal: "Normal",
	Italic: "Italic",
}

// String returns the human-readable name for the font style.
func (s Style) String() string {
	if int(s) < len(styleNames) {
		return styleNames[s]
	}
	return "Unknown"
}
