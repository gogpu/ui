package i18n

import (
	"sync"
	"testing"
)

// ---------------------------------------------------------------------------
// Locale
// ---------------------------------------------------------------------------

func TestNewLocale(t *testing.T) {
	tests := []struct {
		name     string
		lang     string
		region   string
		wantLang string
		wantReg  string
		wantStr  string
	}{
		{"en-US", "en", "US", "en", "US", "en-US"},
		{"ru-RU", "ru", "RU", "ru", "RU", "ru-RU"},
		{"language only", "de", "", "de", "", "de"},
		{"normalizes case", "EN", "us", "en", "US", "en-US"},
		{"empty", "", "", "", "", ""},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			l := NewLocale(tt.lang, tt.region)
			if l.Language != tt.wantLang {
				t.Errorf("Language = %q, want %q", l.Language, tt.wantLang)
			}
			if l.Region != tt.wantReg {
				t.Errorf("Region = %q, want %q", l.Region, tt.wantReg)
			}
			if l.String() != tt.wantStr {
				t.Errorf("String() = %q, want %q", l.String(), tt.wantStr)
			}
		})
	}
}

func TestParseLocale(t *testing.T) {
	tests := []struct {
		input    string
		wantLang string
		wantReg  string
	}{
		{"en-US", "en", "US"},
		{"ru_RU", "ru", "RU"},
		{"de", "de", ""},
		{"ar-SA", "ar", "SA"},
		{"", "", ""},
		{"zh-TW", "zh", "TW"},
	}

	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			l := ParseLocale(tt.input)
			if l.Language != tt.wantLang {
				t.Errorf("Language = %q, want %q", l.Language, tt.wantLang)
			}
			if l.Region != tt.wantReg {
				t.Errorf("Region = %q, want %q", l.Region, tt.wantReg)
			}
		})
	}
}

func TestLocaleIsZero(t *testing.T) {
	if !(Locale{}).IsZero() {
		t.Error("zero locale should return true for IsZero")
	}
	if NewLocale("en", "").IsZero() {
		t.Error("non-zero locale should return false for IsZero")
	}
}

func TestLocaleMatches(t *testing.T) {
	tests := []struct {
		name  string
		a, b  Locale
		match bool
	}{
		{"exact match", NewLocale("en", "US"), NewLocale("en", "US"), true},
		{"language only vs region", NewLocale("en", ""), NewLocale("en", "US"), true},
		{"region vs language only", NewLocale("en", "US"), NewLocale("en", ""), true},
		{"different region", NewLocale("en", "US"), NewLocale("en", "GB"), false},
		{"different language", NewLocale("en", "US"), NewLocale("ru", "RU"), false},
		{"both language only", NewLocale("en", ""), NewLocale("en", ""), true},
		{"case insensitive lang", NewLocale("EN", "US"), NewLocale("en", "US"), true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := tt.a.Matches(tt.b); got != tt.match {
				t.Errorf("Matches() = %v, want %v", got, tt.match)
			}
		})
	}
}

func TestLocaleDirection(t *testing.T) {
	tests := []struct {
		lang string
		want Direction
	}{
		{"en", LTR},
		{"ru", LTR},
		{"de", LTR},
		{"ar", RTL},
		{"he", RTL},
		{"fa", RTL},
		{"ur", RTL},
		{"ja", LTR},
	}

	for _, tt := range tests {
		t.Run(tt.lang, func(t *testing.T) {
			l := NewLocale(tt.lang, "")
			if got := l.Direction(); got != tt.want {
				t.Errorf("Direction() = %v, want %v", got, tt.want)
			}
		})
	}
}

// ---------------------------------------------------------------------------
// Direction
// ---------------------------------------------------------------------------

func TestDirectionString(t *testing.T) {
	tests := []struct {
		d    Direction
		want string
	}{
		{LTR, "LTR"},
		{RTL, "RTL"},
		{Direction(99), "LTR"}, // unknown defaults to LTR
	}

	for _, tt := range tests {
		t.Run(tt.want, func(t *testing.T) {
			if got := tt.d.String(); got != tt.want {
				t.Errorf("String() = %q, want %q", got, tt.want)
			}
		})
	}
}

func TestDirectionIsRTL(t *testing.T) {
	if LTR.IsRTL() {
		t.Error("LTR.IsRTL() should be false")
	}
	if !RTL.IsRTL() {
		t.Error("RTL.IsRTL() should be true")
	}
}

func TestDirectionForLanguage(t *testing.T) {
	tests := []struct {
		lang string
		want Direction
	}{
		{"ar", RTL},
		{"he", RTL},
		{"fa", RTL},
		{"ur", RTL},
		{"yi", RTL},
		{"ps", RTL},
		{"sd", RTL},
		{"ku", RTL},
		{"ug", RTL},
		{"dv", RTL},
		{"AR", RTL}, // case insensitive
		{"en", LTR},
		{"ru", LTR},
		{"xx", LTR}, // unknown defaults to LTR
	}

	for _, tt := range tests {
		t.Run(tt.lang, func(t *testing.T) {
			if got := DirectionForLanguage(tt.lang); got != tt.want {
				t.Errorf("DirectionForLanguage(%q) = %v, want %v", tt.lang, got, tt.want)
			}
		})
	}
}

// ---------------------------------------------------------------------------
// Plural
// ---------------------------------------------------------------------------

func TestPluralRuleEnglish(t *testing.T) {
	tests := []struct {
		count int
		want  PluralCategory
	}{
		{0, PluralOther},
		{1, PluralOne},
		{-1, PluralOne},
		{2, PluralOther},
		{10, PluralOther},
		{100, PluralOther},
	}

	for _, tt := range tests {
		if got := PluralRuleEnglish(tt.count); got != tt.want {
			t.Errorf("PluralRuleEnglish(%d) = %v, want %v", tt.count, got, tt.want)
		}
	}
}

func TestPluralRuleRussian(t *testing.T) {
	tests := []struct {
		count int
		want  PluralCategory
	}{
		{0, PluralMany},
		{1, PluralOne},
		{2, PluralFew},
		{3, PluralFew},
		{4, PluralFew},
		{5, PluralMany},
		{10, PluralMany},
		{11, PluralMany},
		{12, PluralMany},
		{14, PluralMany},
		{20, PluralMany},
		{21, PluralOne},
		{22, PluralFew},
		{25, PluralMany},
		{100, PluralMany},
		{101, PluralOne},
		{111, PluralMany},
		{112, PluralMany},
		{-5, PluralMany},
	}

	for _, tt := range tests {
		if got := PluralRuleRussian(tt.count); got != tt.want {
			t.Errorf("PluralRuleRussian(%d) = %v, want %v", tt.count, got, tt.want)
		}
	}
}

func TestPluralRuleArabic(t *testing.T) {
	tests := []struct {
		count int
		want  PluralCategory
	}{
		{0, PluralZero},
		{1, PluralOne},
		{2, PluralTwo},
		{3, PluralFew},
		{10, PluralFew},
		{11, PluralMany},
		{99, PluralMany},
		{100, PluralOther},
		{200, PluralOther},
		{103, PluralFew},
		{111, PluralMany},
	}

	for _, tt := range tests {
		if got := PluralRuleArabic(tt.count); got != tt.want {
			t.Errorf("PluralRuleArabic(%d) = %v, want %v", tt.count, got, tt.want)
		}
	}
}

func TestPluralRuleFrench(t *testing.T) {
	tests := []struct {
		count int
		want  PluralCategory
	}{
		{0, PluralOne},
		{1, PluralOne},
		{2, PluralOther},
		{100, PluralOther},
	}

	for _, tt := range tests {
		if got := PluralRuleFrench(tt.count); got != tt.want {
			t.Errorf("PluralRuleFrench(%d) = %v, want %v", tt.count, got, tt.want)
		}
	}
}

func TestPluralRuleJapanese(t *testing.T) {
	tests := []struct {
		count int
		want  PluralCategory
	}{
		{0, PluralOther},
		{1, PluralOther},
		{100, PluralOther},
	}

	for _, tt := range tests {
		if got := PluralRuleJapanese(tt.count); got != tt.want {
			t.Errorf("PluralRuleJapanese(%d) = %v, want %v", tt.count, got, tt.want)
		}
	}
}

func TestPluralRulePolish(t *testing.T) {
	tests := []struct {
		count int
		want  PluralCategory
	}{
		{0, PluralMany},
		{1, PluralOne},
		{2, PluralFew},
		{4, PluralFew},
		{5, PluralMany},
		{12, PluralMany},
		{22, PluralFew},
		{100, PluralMany},
	}

	for _, tt := range tests {
		if got := PluralRulePolish(tt.count); got != tt.want {
			t.Errorf("PluralRulePolish(%d) = %v, want %v", tt.count, got, tt.want)
		}
	}
}

func TestPluralFormsGet(t *testing.T) {
	forms := PluralForms{
		Zero:  "zero",
		One:   "one",
		Two:   "two",
		Few:   "few",
		Many:  "many",
		Other: "other",
	}

	tests := []struct {
		cat  PluralCategory
		want string
	}{
		{PluralZero, "zero"},
		{PluralOne, "one"},
		{PluralTwo, "two"},
		{PluralFew, "few"},
		{PluralMany, "many"},
		{PluralOther, "other"},
	}

	for _, tt := range tests {
		if got := forms.Get(tt.cat); got != tt.want {
			t.Errorf("Get(%v) = %q, want %q", tt.cat, got, tt.want)
		}
	}
}

func TestPluralFormsGetFallback(t *testing.T) {
	forms := PluralForms{
		One:   "one",
		Other: "other",
	}

	// Zero is empty, should fall back to Other.
	if got := forms.Get(PluralZero); got != "other" {
		t.Errorf("Get(PluralZero) = %q, want %q", got, "other")
	}
	if got := forms.Get(PluralFew); got != "other" {
		t.Errorf("Get(PluralFew) = %q, want %q", got, "other")
	}
}

func TestPluralRuleForUnknownLanguage(t *testing.T) {
	rule := pluralRuleForLanguage("xx")
	// Should fall back to English rules.
	if got := rule(1); got != PluralOne {
		t.Errorf("unknown language rule(1) = %v, want PluralOne", got)
	}
	if got := rule(2); got != PluralOther {
		t.Errorf("unknown language rule(2) = %v, want PluralOther", got)
	}
}

// ---------------------------------------------------------------------------
// Bundle
// ---------------------------------------------------------------------------

func TestBundle(t *testing.T) {
	b := NewBundle(NewLocale("en", "US"))

	if got := b.Locale().String(); got != "en-US" {
		t.Errorf("Locale() = %q, want %q", got, "en-US")
	}
	if b.Len() != 0 {
		t.Error("new bundle should have 0 entries")
	}

	b.Set("hello", "Hello!")
	b.Set("bye", "Goodbye!")

	if msg, ok := b.Get("hello"); !ok || msg != "Hello!" {
		t.Errorf("Get(hello) = %q, %v", msg, ok)
	}
	if _, ok := b.Get("missing"); ok {
		t.Error("Get(missing) should return false")
	}

	forms := PluralForms{One: "%d item", Other: "%d items"}
	b.SetPlural("items", forms)

	if f, ok := b.GetPlural("items"); !ok || f.One != "%d item" {
		t.Errorf("GetPlural(items) = %v, %v", f, ok)
	}
	if _, ok := b.GetPlural("missing"); ok {
		t.Error("GetPlural(missing) should return false")
	}

	if b.Len() != 3 { // 2 simple + 1 plural
		t.Errorf("Len() = %d, want 3", b.Len())
	}

	keys := b.Keys()
	if len(keys) != 3 {
		t.Errorf("Keys() returned %d keys, want 3", len(keys))
	}
}

func TestBundleSetAll(t *testing.T) {
	b1 := NewBundle(NewLocale("en", ""))
	b1.Set("a", "A1")
	b1.Set("b", "B1")

	b2 := NewBundle(NewLocale("en", ""))
	b2.Set("b", "B2")
	b2.Set("c", "C2")
	b2.SetPlural("p", PluralForms{One: "one"})

	b1.SetAll(b2)

	if msg, _ := b1.Get("a"); msg != "A1" {
		t.Errorf("a = %q, want A1", msg)
	}
	if msg, _ := b1.Get("b"); msg != "B2" {
		t.Errorf("b = %q, want B2 (overwritten)", msg)
	}
	if msg, _ := b1.Get("c"); msg != "C2" {
		t.Errorf("c = %q, want C2", msg)
	}
	if f, ok := b1.GetPlural("p"); !ok || f.One != "one" {
		t.Errorf("plural p = %v, %v", f, ok)
	}
}

func TestBundleSetAllNil(t *testing.T) {
	b := NewBundle(NewLocale("en", ""))
	b.Set("a", "A")
	b.SetAll(nil) // should not panic
	if msg, _ := b.Get("a"); msg != "A" {
		t.Errorf("after SetAll(nil), a = %q, want A", msg)
	}
}

// ---------------------------------------------------------------------------
// Translator — basic resolution
// ---------------------------------------------------------------------------

func TestTranslatorT(t *testing.T) {
	tr := NewTranslator(NewLocale("en", "US"))

	en := NewBundle(NewLocale("en", "US"))
	en.Set("greeting", "Hello!")
	en.Set("shared", "English shared")
	tr.AddBundle(en)

	ru := NewBundle(NewLocale("ru", "RU"))
	ru.Set("greeting", "Привет!")
	tr.AddBundle(ru)

	// Default locale is fallback (en-US).
	if got := tr.T("greeting"); got != "Hello!" {
		t.Errorf("T(greeting) = %q, want Hello!", got)
	}

	tr.SetLocale(NewLocale("ru", "RU"))
	if got := tr.T("greeting"); got != "Привет!" {
		t.Errorf("T(greeting) = %q, want Привет!", got)
	}

	// Fallback: "shared" not in ru-RU, falls back to en-US.
	if got := tr.T("shared"); got != "English shared" {
		t.Errorf("T(shared) = %q, want 'English shared'", got)
	}

	// Missing: returns key itself.
	if got := tr.T("missing.key"); got != "missing.key" {
		t.Errorf("T(missing.key) = %q, want 'missing.key'", got)
	}
}

func TestTranslatorTf(t *testing.T) {
	tr := NewTranslator(NewLocale("en", ""))

	en := NewBundle(NewLocale("en", ""))
	en.Set("welcome", "Welcome, %s!")
	en.Set("stats", "%d users, %d posts")
	tr.AddBundle(en)

	if got := tr.Tf("welcome", "World"); got != "Welcome, World!" {
		t.Errorf("Tf = %q, want 'Welcome, World!'", got)
	}
	if got := tr.Tf("stats", 10, 5); got != "10 users, 5 posts" {
		t.Errorf("Tf = %q, want '10 users, 5 posts'", got)
	}

	// Missing key returns key itself.
	if got := tr.Tf("nope", "arg"); got != "nope" {
		t.Errorf("Tf(nope) = %q, want 'nope'", got)
	}
}

func TestTranslatorTp(t *testing.T) {
	tr := NewTranslator(NewLocale("en", "US"))

	en := NewBundle(NewLocale("en", "US"))
	en.SetPlural("items", PluralForms{
		One:   "%d item",
		Other: "%d items",
	})
	tr.AddBundle(en)

	tests := []struct {
		count int
		want  string
	}{
		{0, "0 items"},
		{1, "1 item"},
		{2, "2 items"},
		{100, "100 items"},
	}

	for _, tt := range tests {
		if got := tr.Tp("items", tt.count); got != tt.want {
			t.Errorf("Tp(items, %d) = %q, want %q", tt.count, got, tt.want)
		}
	}

	// Missing plural key returns key.
	if got := tr.Tp("missing", 5); got != "missing" {
		t.Errorf("Tp(missing) = %q, want 'missing'", got)
	}
}

func TestTranslatorTpf(t *testing.T) {
	tr := NewTranslator(NewLocale("en", ""))

	en := NewBundle(NewLocale("en", ""))
	en.SetPlural("items.loc", PluralForms{
		One:   "%d item in %s",
		Other: "%d items in %s",
	})
	tr.AddBundle(en)

	if got := tr.Tpf("items.loc", 1, 1, "cart"); got != "1 item in cart" {
		t.Errorf("Tpf(1) = %q", got)
	}
	if got := tr.Tpf("items.loc", 5, 5, "cart"); got != "5 items in cart" {
		t.Errorf("Tpf(5) = %q", got)
	}

	// Missing key returns key.
	if got := tr.Tpf("nope", 1, 1); got != "nope" {
		t.Errorf("Tpf(nope) = %q, want 'nope'", got)
	}
}

func TestTranslatorTpRussian(t *testing.T) {
	tr := NewTranslator(NewLocale("ru", "RU"))

	ru := NewBundle(NewLocale("ru", "RU"))
	ru.SetPlural("files", PluralForms{
		One:   "%d файл",
		Few:   "%d файла",
		Many:  "%d файлов",
		Other: "%d файлов",
	})
	tr.AddBundle(ru)

	tests := []struct {
		count int
		want  string
	}{
		{1, "1 файл"},
		{2, "2 файла"},
		{5, "5 файлов"},
		{11, "11 файлов"},
		{21, "21 файл"},
		{22, "22 файла"},
	}

	for _, tt := range tests {
		if got := tr.Tp("files", tt.count); got != tt.want {
			t.Errorf("Tp(files, %d) = %q, want %q", tt.count, got, tt.want)
		}
	}
}

// ---------------------------------------------------------------------------
// Translator — fallback chain
// ---------------------------------------------------------------------------

func TestTranslatorFallbackChain(t *testing.T) {
	tr := NewTranslator(NewLocale("en", "US"))

	// Register only a language-only bundle.
	en := NewBundle(NewLocale("en", ""))
	en.Set("lang_only", "from en")
	tr.AddBundle(en)

	// Register exact en-US bundle.
	enUS := NewBundle(NewLocale("en", "US"))
	enUS.Set("exact", "from en-US")
	tr.AddBundle(enUS)

	// Exact match first.
	if got := tr.T("exact"); got != "from en-US" {
		t.Errorf("T(exact) = %q, want 'from en-US'", got)
	}

	// Language-only fallback for current locale.
	if got := tr.T("lang_only"); got != "from en" {
		t.Errorf("T(lang_only) = %q, want 'from en'", got)
	}

	// Switch to ru-RU (no bundle), falls back to en-US then en.
	tr.SetLocale(NewLocale("ru", "RU"))
	if got := tr.T("exact"); got != "from en-US" {
		t.Errorf("T(exact) after locale switch = %q, want 'from en-US'", got)
	}
	if got := tr.T("lang_only"); got != "from en" {
		t.Errorf("T(lang_only) after locale switch = %q, want 'from en'", got)
	}
}

func TestTranslatorPluralFallback(t *testing.T) {
	tr := NewTranslator(NewLocale("en", "US"))

	en := NewBundle(NewLocale("en", "US"))
	en.SetPlural("items", PluralForms{One: "%d item", Other: "%d items"})
	tr.AddBundle(en)

	// Switch to locale without plurals -- should fall back.
	tr.SetLocale(NewLocale("de", "DE"))
	if got := tr.Tp("items", 1); got != "1 item" {
		t.Errorf("Tp fallback = %q, want '1 item'", got)
	}
}

// ---------------------------------------------------------------------------
// Translator — Has, Direction, BundleCount
// ---------------------------------------------------------------------------

func TestTranslatorHas(t *testing.T) {
	tr := NewTranslator(NewLocale("en", ""))
	en := NewBundle(NewLocale("en", ""))
	en.Set("exists", "yes")
	tr.AddBundle(en)

	if !tr.Has("exists") {
		t.Error("Has(exists) should be true")
	}
	if tr.Has("nope") {
		t.Error("Has(nope) should be false")
	}
}

func TestTranslatorDirection(t *testing.T) {
	tr := NewTranslator(NewLocale("en", "US"))
	if tr.Direction() != LTR {
		t.Error("English direction should be LTR")
	}

	tr.SetLocale(NewLocale("ar", "SA"))
	if tr.Direction() != RTL {
		t.Error("Arabic direction should be RTL")
	}
}

func TestTranslatorBundleCount(t *testing.T) {
	tr := NewTranslator(NewLocale("en", ""))
	if tr.BundleCount() != 0 {
		t.Error("empty translator should have 0 bundles")
	}

	tr.AddBundle(NewBundle(NewLocale("en", "")))
	tr.AddBundle(NewBundle(NewLocale("ru", "")))
	if tr.BundleCount() != 2 {
		t.Errorf("BundleCount() = %d, want 2", tr.BundleCount())
	}
}

func TestTranslatorAddNilBundle(t *testing.T) {
	tr := NewTranslator(NewLocale("en", ""))
	tr.AddBundle(nil) // should not panic
	if tr.BundleCount() != 0 {
		t.Error("adding nil bundle should not change count")
	}
}

func TestTranslatorFallbackAccessors(t *testing.T) {
	tr := NewTranslator(NewLocale("en", "US"))

	if got := tr.Fallback().String(); got != "en-US" {
		t.Errorf("Fallback() = %q, want en-US", got)
	}

	tr.SetFallback(NewLocale("de", "DE"))
	if got := tr.Fallback().String(); got != "de-DE" {
		t.Errorf("Fallback() = %q, want de-DE", got)
	}
}

// ---------------------------------------------------------------------------
// Translator — custom plural rules
// ---------------------------------------------------------------------------

func TestTranslatorCustomPluralRule(t *testing.T) {
	tr := NewTranslator(NewLocale("xx", ""))

	// Register a custom rule where everything is "few".
	tr.SetPluralRule("xx", func(count int) PluralCategory {
		_ = count
		return PluralFew
	})

	b := NewBundle(NewLocale("xx", ""))
	b.SetPlural("things", PluralForms{
		Few:   "%d few-things",
		Other: "%d other-things",
	})
	tr.AddBundle(b)

	if got := tr.Tp("things", 999); got != "999 few-things" {
		t.Errorf("custom rule Tp = %q, want '999 few-things'", got)
	}
}

// ---------------------------------------------------------------------------
// Translator — signal integration
// ---------------------------------------------------------------------------

func TestTranslatorLocaleSignal(t *testing.T) {
	tr := NewTranslator(NewLocale("en", "US"))
	sig := tr.LocaleSignal()

	if got := sig.Get().String(); got != "en-US" {
		t.Errorf("initial signal = %q, want en-US", got)
	}

	tr.SetLocale(NewLocale("ru", "RU"))

	if got := sig.Get().String(); got != "ru-RU" {
		t.Errorf("signal after SetLocale = %q, want ru-RU", got)
	}

	tr.SetLocale(NewLocale("en", "US"))

	if got := sig.Get().String(); got != "en-US" {
		t.Errorf("signal after second SetLocale = %q, want en-US", got)
	}
}

// ---------------------------------------------------------------------------
// Translator — concurrent access
// ---------------------------------------------------------------------------

func TestTranslatorConcurrent(t *testing.T) {
	tr := NewTranslator(NewLocale("en", "US"))

	en := NewBundle(NewLocale("en", "US"))
	en.Set("hello", "Hello!")
	en.SetPlural("items", PluralForms{One: "%d item", Other: "%d items"})
	tr.AddBundle(en)

	ru := NewBundle(NewLocale("ru", "RU"))
	ru.Set("hello", "Привет!")
	tr.AddBundle(ru)

	var wg sync.WaitGroup
	const goroutines = 50

	// Only test Translator mutex safety — concurrent reads with sequential writes.
	// Signal notification is tested separately in TestTranslatorLocaleSignal.
	for i := range goroutines {
		wg.Add(1)
		go func(idx int) {
			defer wg.Done()

			_ = tr.T("hello")
			_ = tr.Tp("items", idx)
			_ = tr.Direction()
			_ = tr.Locale()
			_ = tr.Has("hello")
			_ = tr.BundleCount()
		}(i)
	}

	wg.Wait()

	// Sequential locale switches (signal notifications are not goroutine-safe).
	tr.SetLocale(NewLocale("ru", "RU"))
	if tr.T("hello") != "Привет!" {
		t.Error("expected Russian after SetLocale")
	}
	tr.SetLocale(NewLocale("en", "US"))
	if tr.T("hello") != "Hello!" {
		t.Error("expected English after SetLocale")
	}
}

// ---------------------------------------------------------------------------
// Translator — locale accessor
// ---------------------------------------------------------------------------

func TestTranslatorLocaleAccessor(t *testing.T) {
	tr := NewTranslator(NewLocale("en", "US"))

	if got := tr.Locale().String(); got != "en-US" {
		t.Errorf("initial Locale() = %q, want en-US", got)
	}

	tr.SetLocale(NewLocale("de", "DE"))
	if got := tr.Locale().String(); got != "de-DE" {
		t.Errorf("Locale() after set = %q, want de-DE", got)
	}
}

// ---------------------------------------------------------------------------
// Translator — bundle replacement
// ---------------------------------------------------------------------------

func TestTranslatorBundleReplacement(t *testing.T) {
	tr := NewTranslator(NewLocale("en", ""))

	b1 := NewBundle(NewLocale("en", ""))
	b1.Set("key", "value1")
	tr.AddBundle(b1)

	if got := tr.T("key"); got != "value1" {
		t.Errorf("T(key) = %q, want value1", got)
	}

	b2 := NewBundle(NewLocale("en", ""))
	b2.Set("key", "value2")
	tr.AddBundle(b2) // replaces b1

	if got := tr.T("key"); got != "value2" {
		t.Errorf("T(key) after replacement = %q, want value2", got)
	}
}

// ---------------------------------------------------------------------------
// abs helper
// ---------------------------------------------------------------------------

func TestAbs(t *testing.T) {
	tests := []struct {
		n, want int
	}{
		{0, 0},
		{5, 5},
		{-5, 5},
		{-1, 1},
	}

	for _, tt := range tests {
		if got := abs(tt.n); got != tt.want {
			t.Errorf("abs(%d) = %d, want %d", tt.n, got, tt.want)
		}
	}
}

// ---------------------------------------------------------------------------
// Edge cases
// ---------------------------------------------------------------------------

func TestTranslatorTWithEmptyKey(t *testing.T) {
	tr := NewTranslator(NewLocale("en", ""))
	en := NewBundle(NewLocale("en", ""))
	en.Set("", "empty key value")
	tr.AddBundle(en)

	if got := tr.T(""); got != "empty key value" {
		t.Errorf("T('') = %q, want 'empty key value'", got)
	}
}

func TestParseLocaleEdgeCases(t *testing.T) {
	// Multiple separators — only split on first.
	l := ParseLocale("en-US-extra")
	if l.Language != "en" {
		t.Errorf("Language = %q, want en", l.Language)
	}
	// Region gets "US-EXTRA" because we only split on the first separator.
	if l.Region != "US-EXTRA" {
		t.Errorf("Region = %q, want US-EXTRA", l.Region)
	}
}
