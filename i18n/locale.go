package i18n

import "strings"

// Locale represents a language and optional region combination.
//
// Language codes follow BCP 47 / ISO 639-1 (e.g., "en", "ru", "ar").
// Region codes follow ISO 3166-1 alpha-2 (e.g., "US", "RU", "SA").
//
// A Locale with an empty Region matches any region for that language
// during fallback resolution.
type Locale struct {
	// Language is the ISO 639-1 language code (e.g., "en", "ru", "ar").
	Language string

	// Region is the optional ISO 3166-1 alpha-2 region code (e.g., "US", "RU").
	// Empty string means no specific region.
	Region string
}

// NewLocale creates a new Locale with the given language and region.
//
// Language is normalized to lowercase, region to uppercase, following
// BCP 47 conventions. If region is empty, the locale matches any region
// for that language.
//
// Example:
//
//	en := i18n.NewLocale("en", "US")  // English (United States)
//	ru := i18n.NewLocale("ru", "")    // Russian (any region)
func NewLocale(language, region string) Locale {
	return Locale{
		Language: strings.ToLower(language),
		Region:   strings.ToUpper(region),
	}
}

// ParseLocale parses a locale string in the format "language-region" or "language".
//
// Accepted formats:
//   - "en"      -> Locale{Language: "en"}
//   - "en-US"   -> Locale{Language: "en", Region: "US"}
//   - "en_US"   -> Locale{Language: "en", Region: "US"}
//
// Returns a zero-value Locale if the input is empty.
func ParseLocale(s string) Locale {
	if s == "" {
		return Locale{}
	}

	// Support both - and _ separators.
	sep := strings.IndexAny(s, "-_")
	if sep < 0 {
		return NewLocale(s, "")
	}

	return NewLocale(s[:sep], s[sep+1:])
}

// String returns the locale in "language-REGION" format (e.g., "en-US").
//
// If Region is empty, returns just the language (e.g., "en").
func (l Locale) String() string {
	if l.Region == "" {
		return l.Language
	}
	return l.Language + "-" + l.Region
}

// IsZero returns true if the locale has no language set.
func (l Locale) IsZero() bool {
	return l.Language == ""
}

// Direction returns the text direction for this locale.
//
// Returns [RTL] for Arabic ("ar"), Hebrew ("he"), Persian ("fa"),
// and Urdu ("ur"). Returns [LTR] for all other languages.
func (l Locale) Direction() Direction {
	return DirectionForLanguage(l.Language)
}

// Matches reports whether this locale matches the other locale.
//
// A locale with an empty region matches any locale with the same language.
// Both language comparisons are case-insensitive.
func (l Locale) Matches(other Locale) bool {
	if !strings.EqualFold(l.Language, other.Language) {
		return false
	}

	// If either region is empty, match on language only.
	if l.Region == "" || other.Region == "" {
		return true
	}

	return strings.EqualFold(l.Region, other.Region)
}
