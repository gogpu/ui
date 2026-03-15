// Package i18n provides internationalization support for the gogpu/ui toolkit.
//
// The i18n package enables string localization, locale management, plural rules,
// and text direction (LTR/RTL) detection. It integrates with the reactive state
// system via [state.Signal] for dynamic locale switching.
//
// # Core Types
//
//   - [Locale] — represents a language/region combination (e.g., "en-US", "ru-RU")
//   - [Bundle] — holds message translations for a single locale
//   - [Translator] — resolves messages with locale fallback chains
//   - [Direction] — text direction (LTR or RTL)
//   - [PluralForms] — plural category forms (zero, one, two, few, many, other)
//
// # Basic Usage
//
//	// Create translator with English fallback
//	t := i18n.NewTranslator(i18n.NewLocale("en", "US"))
//
//	// Add English bundle
//	en := i18n.NewBundle(i18n.NewLocale("en", "US"))
//	en.Set("greeting", "Hello!")
//	en.SetPlural("items", i18n.PluralForms{
//	    One:   "%d item",
//	    Other: "%d items",
//	})
//	t.AddBundle(en)
//
//	// Add Russian bundle
//	ru := i18n.NewBundle(i18n.NewLocale("ru", "RU"))
//	ru.Set("greeting", "Привет!")
//	t.AddBundle(ru)
//
//	// Resolve messages
//	t.T("greeting")          // "Hello!"
//	t.SetLocale(i18n.NewLocale("ru", "RU"))
//	t.T("greeting")          // "Привет!"
//
// # Plural Support
//
// Plural rules follow CLDR categories (zero, one, two, few, many, other).
// Built-in rules are provided for English, Russian, and Arabic:
//
//	t.Tp("items", 1)   // "1 item"    (one)
//	t.Tp("items", 5)   // "5 items"   (other)
//	t.Tpf("items", 3, 3) // "3 items" (other, formatted)
//
// # Reactive Locale Switching
//
// The Translator exposes a [state.ReadonlySignal] for the current locale,
// enabling reactive UI updates when the locale changes:
//
//	localeSig := t.LocaleSignal()
//	state.Effect(func() {
//	    fmt.Println("Locale changed to:", localeSig.Get())
//	})
//
// # Text Direction
//
// RTL detection is built in for Arabic, Hebrew, Persian, and Urdu:
//
//	locale := i18n.NewLocale("ar", "SA")
//	locale.Direction() // i18n.RTL
//
// # Thread Safety
//
// [Translator] is safe for concurrent access from multiple goroutines.
// All methods use a sync.RWMutex internally. [Bundle] is NOT thread-safe;
// it should be fully populated before being added to a Translator.
package i18n
