package i18n

// Bundle is a message catalog for a single locale.
//
// A Bundle holds key-value message pairs and optional plural forms for a
// specific locale. Bundles are NOT thread-safe; they should be fully
// populated before being added to a [Translator].
//
// Example:
//
//	en := i18n.NewBundle(i18n.NewLocale("en", "US"))
//	en.Set("welcome", "Welcome!")
//	en.SetPlural("files", i18n.PluralForms{
//	    One:   "%d file",
//	    Other: "%d files",
//	})
type Bundle struct {
	locale   Locale
	messages map[string]string
	plurals  map[string]PluralForms
}

// NewBundle creates a new empty Bundle for the given locale.
func NewBundle(locale Locale) *Bundle {
	return &Bundle{
		locale:   locale,
		messages: make(map[string]string),
		plurals:  make(map[string]PluralForms),
	}
}

// Locale returns the locale this bundle is for.
func (b *Bundle) Locale() Locale {
	return b.locale
}

// Set adds or replaces a simple message for the given key.
func (b *Bundle) Set(key, value string) {
	b.messages[key] = value
}

// SetPlural adds or replaces plural forms for the given key.
func (b *Bundle) SetPlural(key string, forms PluralForms) {
	b.plurals[key] = forms
}

// Get retrieves a simple message by key.
//
// Returns the message and true if found, or empty string and false if not.
func (b *Bundle) Get(key string) (string, bool) {
	msg, ok := b.messages[key]
	return msg, ok
}

// GetPlural retrieves plural forms by key.
//
// Returns the forms and true if found, or zero PluralForms and false if not.
func (b *Bundle) GetPlural(key string) (PluralForms, bool) {
	forms, ok := b.plurals[key]
	return forms, ok
}

// Len returns the total number of messages (simple + plural keys).
func (b *Bundle) Len() int {
	return len(b.messages) + len(b.plurals)
}

// Keys returns all message keys (both simple and plural) in no particular order.
func (b *Bundle) Keys() []string {
	keys := make([]string, 0, len(b.messages)+len(b.plurals))
	for k := range b.messages {
		keys = append(keys, k)
	}
	for k := range b.plurals {
		keys = append(keys, k)
	}
	return keys
}

// SetAll copies all messages from another bundle into this one.
//
// Existing keys are overwritten. This is useful for merging partial
// translations with a base bundle.
func (b *Bundle) SetAll(other *Bundle) {
	if other == nil {
		return
	}
	for k, v := range other.messages {
		b.messages[k] = v
	}
	for k, v := range other.plurals {
		b.plurals[k] = v
	}
}
