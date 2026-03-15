package i18n

// PluralCategory represents a CLDR plural category.
type PluralCategory uint8

const (
	// PluralOther is the default/fallback category.
	PluralOther PluralCategory = iota

	// PluralZero is used for zero quantity in languages that distinguish it (e.g., Arabic).
	PluralZero

	// PluralOne is used for singular quantity.
	PluralOne

	// PluralTwo is used for dual quantity in languages that distinguish it (e.g., Arabic).
	PluralTwo

	// PluralFew is used for a small quantity in languages with this distinction (e.g., Russian 2-4).
	PluralFew

	// PluralMany is used for a large quantity in languages with this distinction (e.g., Russian 5-20).
	PluralMany
)

// PluralForms holds message strings for each CLDR plural category.
//
// Not all categories are used by every language. For example, English only
// uses One and Other, while Arabic uses all six categories. When a category
// is not set (empty string), the translator falls back to Other.
//
// Example (English):
//
//	PluralForms{One: "%d item", Other: "%d items"}
//
// Example (Russian):
//
//	PluralForms{
//	    One:   "%d файл",    // 1, 21, 31...
//	    Few:   "%d файла",   // 2-4, 22-24...
//	    Many:  "%d файлов",  // 5-20, 25-30...
//	    Other: "%d файлов",  // fallback
//	}
type PluralForms struct {
	// Zero form (used in Arabic for 0).
	Zero string

	// One form (singular).
	One string

	// Two form (dual, used in Arabic for 2).
	Two string

	// Few form (e.g., Russian 2-4).
	Few string

	// Many form (e.g., Russian 5-20).
	Many string

	// Other form (default fallback).
	Other string
}

// Get returns the message string for the given plural category.
//
// If the requested category has an empty string, falls back to Other.
func (p PluralForms) Get(cat PluralCategory) string {
	var s string

	switch cat {
	case PluralZero:
		s = p.Zero
	case PluralOne:
		s = p.One
	case PluralTwo:
		s = p.Two
	case PluralFew:
		s = p.Few
	case PluralMany:
		s = p.Many
	case PluralOther:
		s = p.Other
	}

	if s == "" {
		return p.Other
	}

	return s
}

// PluralRule is a function that returns the plural category for a given count.
//
// Languages have different rules for determining which plural form to use.
// Each supported language registers a PluralRule function.
type PluralRule func(count int) PluralCategory

// PluralRuleEnglish implements CLDR plural rules for English.
//
// Categories: one (n=1), other (everything else).
func PluralRuleEnglish(count int) PluralCategory {
	if abs(count) == 1 {
		return PluralOne
	}
	return PluralOther
}

// PluralRuleRussian implements CLDR plural rules for Russian and other
// Slavic languages (Ukrainian, Belarusian, Serbian, Croatian).
//
// Categories:
//   - one: n%10=1 and n%100!=11 (1, 21, 31, 41...)
//   - few: n%10 in 2..4 and n%100 not in 12..14 (2-4, 22-24, 32-34...)
//   - many: n%10=0 or n%10 in 5..9 or n%100 in 11..14 (0, 5-20, 25-30...)
//   - other: fallback (non-integer, but we handle int only)
func PluralRuleRussian(count int) PluralCategory {
	n := abs(count)
	mod10 := n % 10
	mod100 := n % 100

	switch {
	case mod10 == 1 && mod100 != 11:
		return PluralOne
	case mod10 >= 2 && mod10 <= 4 && (mod100 < 12 || mod100 > 14):
		return PluralFew
	default:
		return PluralMany
	}
}

// PluralRuleArabic implements CLDR plural rules for Arabic.
//
// Categories:
//   - zero: n=0
//   - one: n=1
//   - two: n=2
//   - few: n%100 in 3..10
//   - many: n%100 in 11..99
//   - other: everything else (100, 200, ...)
func PluralRuleArabic(count int) PluralCategory {
	n := abs(count)
	mod100 := n % 100

	switch {
	case n == 0:
		return PluralZero
	case n == 1:
		return PluralOne
	case n == 2:
		return PluralTwo
	case mod100 >= 3 && mod100 <= 10:
		return PluralFew
	case mod100 >= 11 && mod100 <= 99:
		return PluralMany
	default:
		return PluralOther
	}
}

// PluralRuleFrench implements CLDR plural rules for French and
// other Romance languages that treat 0 as singular.
//
// Categories: one (n=0 or n=1), other (everything else).
func PluralRuleFrench(count int) PluralCategory {
	n := abs(count)
	if n <= 1 {
		return PluralOne
	}
	return PluralOther
}

// PluralRuleJapanese implements CLDR plural rules for Japanese, Chinese,
// Korean, Thai, and other languages with no grammatical plural.
//
// Categories: other (always).
func PluralRuleJapanese(count int) PluralCategory {
	_ = count // all counts map to other
	return PluralOther
}

// PluralRulePolish implements CLDR plural rules for Polish.
//
// Categories:
//   - one: n=1
//   - few: n%10 in 2..4 and n%100 not in 12..14
//   - many: everything else
func PluralRulePolish(count int) PluralCategory {
	n := abs(count)
	mod10 := n % 10
	mod100 := n % 100

	switch {
	case n == 1:
		return PluralOne
	case mod10 >= 2 && mod10 <= 4 && (mod100 < 12 || mod100 > 14):
		return PluralFew
	default:
		return PluralMany
	}
}

// defaultPluralRules maps language codes to their plural rule functions.
var defaultPluralRules = map[string]PluralRule{
	// Germanic
	"en": PluralRuleEnglish,
	"de": PluralRuleEnglish,
	"nl": PluralRuleEnglish,
	"sv": PluralRuleEnglish,
	"da": PluralRuleEnglish,
	"no": PluralRuleEnglish,

	// Slavic
	"ru": PluralRuleRussian,
	"uk": PluralRuleRussian,
	"be": PluralRuleRussian,
	"hr": PluralRuleRussian,
	"sr": PluralRuleRussian,
	"bs": PluralRuleRussian,

	// Polish (different from Russian)
	"pl": PluralRulePolish,

	// Romance
	"fr": PluralRuleFrench,
	"pt": PluralRuleFrench,
	"it": PluralRuleEnglish,
	"es": PluralRuleEnglish,

	// Semitic
	"ar": PluralRuleArabic,
	"he": PluralRuleEnglish,

	// East Asian (no plural)
	"ja": PluralRuleJapanese,
	"zh": PluralRuleJapanese,
	"ko": PluralRuleJapanese,

	// Other
	"fa": PluralRuleEnglish,
	"ur": PluralRuleEnglish,
	"tr": PluralRuleEnglish,
	"th": PluralRuleJapanese,
	"vi": PluralRuleJapanese,
}

// pluralRuleForLanguage returns the plural rule for a language code.
// Falls back to English rules if the language is not supported.
func pluralRuleForLanguage(language string) PluralRule {
	if rule, ok := defaultPluralRules[language]; ok {
		return rule
	}
	return PluralRuleEnglish
}

// abs returns the absolute value of an integer.
func abs(n int) int {
	if n < 0 {
		return -n
	}
	return n
}
