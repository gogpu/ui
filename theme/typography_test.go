package theme

import "testing"

func TestFontWeight_String(t *testing.T) {
	tests := []struct {
		name   string
		weight FontWeight
		want   string
	}{
		{"thin", FontWeightThin, "Thin"},
		{"extra light", FontWeightExtraLight, "ExtraLight"},
		{"light", FontWeightLight, "Light"},
		{"normal", FontWeightNormal, "Normal"},
		{"medium", FontWeightMedium, "Medium"},
		{"semi bold", FontWeightSemiBold, "SemiBold"},
		{"bold", FontWeightBold, "Bold"},
		{"extra bold", FontWeightExtraBold, "ExtraBold"},
		{"black", FontWeightBlack, "Black"},
		{"unknown", FontWeight(999), "Unknown"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := tt.weight.String(); got != tt.want {
				t.Errorf("FontWeight.String() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestFontWeight_IsBold(t *testing.T) {
	tests := []struct {
		name   string
		weight FontWeight
		want   bool
	}{
		{"normal is not bold", FontWeightNormal, false},
		{"medium is not bold", FontWeightMedium, false},
		{"semi bold is not bold", FontWeightSemiBold, false},
		{"bold is bold", FontWeightBold, true},
		{"extra bold is bold", FontWeightExtraBold, true},
		{"black is bold", FontWeightBlack, true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := tt.weight.IsBold(); got != tt.want {
				t.Errorf("FontWeight.IsBold() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestFontWeight_IsLight(t *testing.T) {
	tests := []struct {
		name   string
		weight FontWeight
		want   bool
	}{
		{"thin is light", FontWeightThin, true},
		{"extra light is light", FontWeightExtraLight, true},
		{"light is light", FontWeightLight, true},
		{"normal is not light", FontWeightNormal, false},
		{"bold is not light", FontWeightBold, false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := tt.weight.IsLight(); got != tt.want {
				t.Errorf("FontWeight.IsLight() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestFontStyle_String(t *testing.T) {
	tests := []struct {
		name  string
		style FontStyle
		want  string
	}{
		{"normal", FontStyleNormal, "Normal"},
		{"italic", FontStyleItalic, "Italic"},
		{"oblique", FontStyleOblique, "Oblique"},
		{"unknown", FontStyle(99), "Unknown"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := tt.style.String(); got != tt.want {
				t.Errorf("FontStyle.String() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestTextStyle_WithSize(t *testing.T) {
	style := TextStyle{Size: 16}
	result := style.WithSize(24)

	if result.Size != 24 {
		t.Errorf("Size = %v, want 24", result.Size)
	}
	// Original unchanged
	if style.Size != 16 {
		t.Errorf("Original size changed: %v, want 16", style.Size)
	}
}

func TestTextStyle_WithWeight(t *testing.T) {
	style := TextStyle{Weight: FontWeightNormal}
	result := style.WithWeight(FontWeightBold)

	if result.Weight != FontWeightBold {
		t.Errorf("Weight = %v, want Bold", result.Weight)
	}
}

func TestTextStyle_WithStyle(t *testing.T) {
	style := TextStyle{Style: FontStyleNormal}
	result := style.WithStyle(FontStyleItalic)

	if result.Style != FontStyleItalic {
		t.Errorf("Style = %v, want Italic", result.Style)
	}
}

func TestTextStyle_WithFont(t *testing.T) {
	style := TextStyle{Font: "Arial"}
	result := style.WithFont("Roboto")

	if result.Font != "Roboto" {
		t.Errorf("Font = %v, want Roboto", result.Font)
	}
}

func TestTextStyle_WithLineHeight(t *testing.T) {
	style := TextStyle{LineHeight: 20}
	result := style.WithLineHeight(28)

	if result.LineHeight != 28 {
		t.Errorf("LineHeight = %v, want 28", result.LineHeight)
	}
}

func TestTextStyle_WithLetterSpacing(t *testing.T) {
	style := TextStyle{LetterSpacing: 0}
	result := style.WithLetterSpacing(0.5)

	if result.LetterSpacing != 0.5 {
		t.Errorf("LetterSpacing = %v, want 0.5", result.LetterSpacing)
	}
}

func TestTextStyle_Bold(t *testing.T) {
	style := TextStyle{Weight: FontWeightNormal}
	result := style.Bold()

	if result.Weight != FontWeightBold {
		t.Errorf("Weight = %v, want Bold", result.Weight)
	}
}

func TestTextStyle_Italic(t *testing.T) {
	style := TextStyle{Style: FontStyleNormal}
	result := style.Italic()

	if result.Style != FontStyleItalic {
		t.Errorf("Style = %v, want Italic", result.Style)
	}
}

func TestDefaultTypography(t *testing.T) {
	typo := DefaultTypography()

	// Check font family
	if typo.FontFamily != "System" {
		t.Errorf("FontFamily = %v, want System", typo.FontFamily)
	}

	// Check display sizes descend
	if typo.DisplayLarge.Size <= typo.DisplayMedium.Size {
		t.Error("DisplayLarge should be bigger than DisplayMedium")
	}
	if typo.DisplayMedium.Size <= typo.DisplaySmall.Size {
		t.Error("DisplayMedium should be bigger than DisplaySmall")
	}

	// Check body styles exist
	if typo.BodyMedium.Size == 0 {
		t.Error("BodyMedium size should not be zero")
	}

	// Check all styles have the same font
	styles := []TextStyle{
		typo.DisplayLarge, typo.DisplayMedium, typo.DisplaySmall,
		typo.HeadlineLarge, typo.HeadlineMedium, typo.HeadlineSmall,
		typo.TitleLarge, typo.TitleMedium, typo.TitleSmall,
		typo.BodyLarge, typo.BodyMedium, typo.BodySmall,
		typo.LabelLarge, typo.LabelMedium, typo.LabelSmall,
	}

	for i, s := range styles {
		if s.Font != "System" {
			t.Errorf("Style %d font = %v, want System", i, s.Font)
		}
	}
}

func TestTypography_WithFontFamily(t *testing.T) {
	typo := DefaultTypography()
	result := typo.WithFontFamily("Roboto")

	if result.FontFamily != "Roboto" {
		t.Errorf("FontFamily = %v, want Roboto", result.FontFamily)
	}
	if result.BodyMedium.Font != "Roboto" {
		t.Errorf("BodyMedium.Font = %v, want Roboto", result.BodyMedium.Font)
	}
	if result.DisplayLarge.Font != "Roboto" {
		t.Errorf("DisplayLarge.Font = %v, want Roboto", result.DisplayLarge.Font)
	}

	// Original unchanged (typo is value, not pointer, so it remains unchanged)
	if typo.FontFamily != "System" {
		t.Errorf("Original FontFamily changed: %v", typo.FontFamily)
	}
}

func TestTypography_Scale(t *testing.T) {
	typo := DefaultTypography()
	originalBodySize := typo.BodyMedium.Size
	originalBodyLine := typo.BodyMedium.LineHeight

	result := typo.Scale(1.5)

	expectedSize := originalBodySize * 1.5
	if result.BodyMedium.Size != expectedSize {
		t.Errorf("BodyMedium.Size = %v, want %v", result.BodyMedium.Size, expectedSize)
	}

	expectedLine := originalBodyLine * 1.5
	if result.BodyMedium.LineHeight != expectedLine {
		t.Errorf("BodyMedium.LineHeight = %v, want %v", result.BodyMedium.LineHeight, expectedLine)
	}

	// Original unchanged
	if typo.BodyMedium.Size != originalBodySize {
		t.Errorf("Original BodyMedium.Size changed: %v", typo.BodyMedium.Size)
	}
}

func TestTypography_Scale_AllStyles(t *testing.T) {
	typo := DefaultTypography()
	result := typo.Scale(2)

	// Check all styles are scaled
	checks := []struct {
		name     string
		original float32
		scaled   float32
	}{
		{"DisplayLarge", typo.DisplayLarge.Size, result.DisplayLarge.Size},
		{"DisplayMedium", typo.DisplayMedium.Size, result.DisplayMedium.Size},
		{"DisplaySmall", typo.DisplaySmall.Size, result.DisplaySmall.Size},
		{"HeadlineLarge", typo.HeadlineLarge.Size, result.HeadlineLarge.Size},
		{"HeadlineMedium", typo.HeadlineMedium.Size, result.HeadlineMedium.Size},
		{"HeadlineSmall", typo.HeadlineSmall.Size, result.HeadlineSmall.Size},
		{"TitleLarge", typo.TitleLarge.Size, result.TitleLarge.Size},
		{"TitleMedium", typo.TitleMedium.Size, result.TitleMedium.Size},
		{"TitleSmall", typo.TitleSmall.Size, result.TitleSmall.Size},
		{"BodyLarge", typo.BodyLarge.Size, result.BodyLarge.Size},
		{"BodyMedium", typo.BodyMedium.Size, result.BodyMedium.Size},
		{"BodySmall", typo.BodySmall.Size, result.BodySmall.Size},
		{"LabelLarge", typo.LabelLarge.Size, result.LabelLarge.Size},
		{"LabelMedium", typo.LabelMedium.Size, result.LabelMedium.Size},
		{"LabelSmall", typo.LabelSmall.Size, result.LabelSmall.Size},
	}

	for _, c := range checks {
		expected := c.original * 2
		if c.scaled != expected {
			t.Errorf("%s.Size = %v, want %v", c.name, c.scaled, expected)
		}
	}
}
