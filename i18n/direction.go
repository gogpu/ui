package i18n

import "strings"

// Direction represents the text direction for a locale.
type Direction uint8

const (
	// LTR is left-to-right text direction (e.g., English, Russian).
	LTR Direction = iota

	// RTL is right-to-left text direction (e.g., Arabic, Hebrew).
	RTL
)

// ltrStr and rtlStr avoid repeated string allocations in String().
const (
	ltrStr = "LTR"
	rtlStr = "RTL"
)

// String returns the human-readable name of the direction.
func (d Direction) String() string {
	switch d {
	case LTR:
		return ltrStr
	case RTL:
		return rtlStr
	default:
		return ltrStr
	}
}

// IsRTL returns true if the direction is right-to-left.
func (d Direction) IsRTL() bool {
	return d == RTL
}

// rtlLanguages is the set of ISO 639-1 language codes that use RTL script.
// This covers the most widely used RTL languages.
var rtlLanguages = map[string]struct{}{
	"ar": {}, // Arabic
	"he": {}, // Hebrew
	"fa": {}, // Persian (Farsi)
	"ur": {}, // Urdu
	"yi": {}, // Yiddish
	"ps": {}, // Pashto
	"sd": {}, // Sindhi
	"ku": {}, // Kurdish (Sorani)
	"ug": {}, // Uyghur
	"dv": {}, // Divehi (Maldivian)
}

// DirectionForLanguage returns the text direction for the given ISO 639-1
// language code.
//
// Returns [RTL] for Arabic, Hebrew, Persian, Urdu, and other RTL languages.
// Returns [LTR] for all other languages.
func DirectionForLanguage(language string) Direction {
	if _, ok := rtlLanguages[strings.ToLower(language)]; ok {
		return RTL
	}
	return LTR
}
