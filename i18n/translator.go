package i18n

import (
	"fmt"
	"sync"

	"github.com/gogpu/ui/state"
)

// Translator resolves localized messages with fallback chain support.
//
// The resolution order for a message key is:
//  1. Current locale (exact match: language + region)
//  2. Current locale (language-only match)
//  3. Fallback locale (exact match)
//  4. Fallback locale (language-only match)
//  5. The key itself (as a last resort)
//
// Translator is safe for concurrent access from multiple goroutines.
//
// Example:
//
//	t := i18n.NewTranslator(i18n.NewLocale("en", "US"))
//	t.AddBundle(enBundle)
//	t.AddBundle(ruBundle)
//
//	t.T("greeting")  // resolves from en-US bundle
//	t.SetLocale(i18n.NewLocale("ru", "RU"))
//	t.T("greeting")  // resolves from ru-RU bundle
type Translator struct {
	mu       sync.RWMutex
	bundles  map[string]*Bundle // keyed by locale.String()
	current  Locale
	fallback Locale

	// localeSig is a signal that emits the current locale whenever it changes.
	localeSig state.Signal[Locale]

	// Custom plural rules per language (overrides defaults).
	pluralRules map[string]PluralRule
}

// NewTranslator creates a new Translator with the given fallback locale.
//
// The fallback locale is used when a message is not found in the current locale.
// The current locale is initially set to the fallback.
func NewTranslator(fallback Locale) *Translator {
	return &Translator{
		bundles:     make(map[string]*Bundle),
		current:     fallback,
		fallback:    fallback,
		localeSig:   state.NewSignal(fallback),
		pluralRules: make(map[string]PluralRule),
	}
}

// AddBundle registers a message bundle for its locale.
//
// If a bundle for the same locale string already exists, it is replaced.
// Bundles should be fully populated before being added.
func (t *Translator) AddBundle(bundle *Bundle) {
	if bundle == nil {
		return
	}

	t.mu.Lock()
	defer t.mu.Unlock()

	t.bundles[bundle.Locale().String()] = bundle
}

// SetLocale changes the current locale.
//
// This triggers the locale signal, notifying all subscribers.
func (t *Translator) SetLocale(locale Locale) {
	t.mu.Lock()
	t.current = locale
	t.mu.Unlock()

	// Signal update outside the lock to avoid deadlock with subscribers.
	t.localeSig.Set(locale)
}

// Locale returns the current locale.
func (t *Translator) Locale() Locale {
	t.mu.RLock()
	defer t.mu.RUnlock()
	return t.current
}

// Fallback returns the fallback locale.
func (t *Translator) Fallback() Locale {
	t.mu.RLock()
	defer t.mu.RUnlock()
	return t.fallback
}

// SetFallback changes the fallback locale.
func (t *Translator) SetFallback(locale Locale) {
	t.mu.Lock()
	defer t.mu.Unlock()
	t.fallback = locale
}

// SetPluralRule registers a custom plural rule for a language.
//
// This overrides the built-in CLDR rule for the given language code.
func (t *Translator) SetPluralRule(language string, rule PluralRule) {
	t.mu.Lock()
	defer t.mu.Unlock()
	t.pluralRules[language] = rule
}

// LocaleSignal returns a read-only signal of the current locale.
//
// Subscribe to this signal to react to locale changes:
//
//	sig := t.LocaleSignal()
//	state.Subscribe(sig, ctx, func(locale i18n.Locale) {
//	    // Update UI for new locale
//	})
func (t *Translator) LocaleSignal() state.ReadonlySignal[Locale] {
	return t.localeSig.AsReadonly()
}

// T resolves a simple message by key.
//
// Resolution order: current locale -> fallback locale -> key itself.
func (t *Translator) T(key string) string {
	t.mu.RLock()
	defer t.mu.RUnlock()

	if msg, ok := t.resolve(key); ok {
		return msg
	}

	return key
}

// Tf resolves a message by key and formats it with fmt.Sprintf.
//
// Resolution order: current locale -> fallback locale -> key itself.
//
// Example:
//
//	t.Tf("greeting.name", "World")  // "Hello, World!"
func (t *Translator) Tf(key string, args ...any) string {
	t.mu.RLock()
	defer t.mu.RUnlock()

	if msg, ok := t.resolve(key); ok {
		return fmt.Sprintf(msg, args...)
	}

	return key
}

// Tp resolves a plural message by key and count.
//
// The plural category is determined by the locale's plural rules,
// then the appropriate plural form is selected and formatted with
// the count value.
//
// Example:
//
//	t.Tp("items", 1)   // "1 item"
//	t.Tp("items", 5)   // "5 items"
func (t *Translator) Tp(key string, count int) string {
	t.mu.RLock()
	defer t.mu.RUnlock()

	if form, ok := t.resolvePlural(key, count); ok {
		return fmt.Sprintf(form, count)
	}

	return key
}

// Tpf resolves a plural message by key and count, then formats with additional args.
//
// The count is used only for plural category selection. The args are passed
// to fmt.Sprintf for formatting the resolved string.
//
// Example:
//
//	// Bundle: PluralForms{One: "%d item in %s", Other: "%d items in %s"}
//	t.Tpf("items.location", 3, 3, "cart")  // "3 items in cart"
func (t *Translator) Tpf(key string, count int, args ...any) string {
	t.mu.RLock()
	defer t.mu.RUnlock()

	if form, ok := t.resolvePlural(key, count); ok {
		return fmt.Sprintf(form, args...)
	}

	return key
}

// Has reports whether a message with the given key exists in the current
// or fallback locale.
func (t *Translator) Has(key string) bool {
	t.mu.RLock()
	defer t.mu.RUnlock()

	_, ok := t.resolve(key)
	return ok
}

// Direction returns the text direction for the current locale.
func (t *Translator) Direction() Direction {
	t.mu.RLock()
	defer t.mu.RUnlock()
	return t.current.Direction()
}

// BundleCount returns the number of registered bundles.
func (t *Translator) BundleCount() int {
	t.mu.RLock()
	defer t.mu.RUnlock()
	return len(t.bundles)
}

// resolve looks up a simple message key through the fallback chain.
// Caller must hold at least a read lock.
func (t *Translator) resolve(key string) (string, bool) {
	// Try exact current locale.
	if b := t.bundles[t.current.String()]; b != nil {
		if msg, ok := b.Get(key); ok {
			return msg, true
		}
	}

	// Try language-only for current locale (if region was specified).
	if t.current.Region != "" {
		langOnly := Locale{Language: t.current.Language}
		if b := t.bundles[langOnly.String()]; b != nil {
			if msg, ok := b.Get(key); ok {
				return msg, true
			}
		}
	}

	// Try exact fallback locale.
	if b := t.bundles[t.fallback.String()]; b != nil {
		if msg, ok := b.Get(key); ok {
			return msg, true
		}
	}

	// Try language-only for fallback locale.
	if t.fallback.Region != "" {
		langOnly := Locale{Language: t.fallback.Language}
		if b := t.bundles[langOnly.String()]; b != nil {
			if msg, ok := b.Get(key); ok {
				return msg, true
			}
		}
	}

	return "", false
}

// resolvePlural looks up a plural message key through the fallback chain.
// Caller must hold at least a read lock.
func (t *Translator) resolvePlural(key string, count int) (string, bool) {
	// Determine plural category using the current locale's rules.
	rule := t.pluralRule(t.current.Language)
	cat := rule(count)

	// Try exact current locale.
	if b := t.bundles[t.current.String()]; b != nil {
		if forms, ok := b.GetPlural(key); ok {
			return forms.Get(cat), true
		}
	}

	// Try language-only for current locale.
	if t.current.Region != "" {
		langOnly := Locale{Language: t.current.Language}
		if b := t.bundles[langOnly.String()]; b != nil {
			if forms, ok := b.GetPlural(key); ok {
				return forms.Get(cat), true
			}
		}
	}

	// Try exact fallback locale (use fallback's plural rules).
	fbRule := t.pluralRule(t.fallback.Language)
	fbCat := fbRule(count)

	if b := t.bundles[t.fallback.String()]; b != nil {
		if forms, ok := b.GetPlural(key); ok {
			return forms.Get(fbCat), true
		}
	}

	// Try language-only for fallback locale.
	if t.fallback.Region != "" {
		langOnly := Locale{Language: t.fallback.Language}
		if b := t.bundles[langOnly.String()]; b != nil {
			if forms, ok := b.GetPlural(key); ok {
				return forms.Get(fbCat), true
			}
		}
	}

	return "", false
}

// pluralRule returns the plural rule for a language, checking custom rules first.
// Caller must hold at least a read lock.
func (t *Translator) pluralRule(language string) PluralRule {
	if rule, ok := t.pluralRules[language]; ok {
		return rule
	}
	return pluralRuleForLanguage(language)
}
