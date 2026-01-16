package theme

import "testing"

func TestThemeMode_String(t *testing.T) {
	tests := []struct {
		name string
		mode ThemeMode
		want string
	}{
		{"light mode", ModeLight, "Light"},
		{"dark mode", ModeDark, "Dark"},
		{"system mode", ModeSystem, "System"},
		{"unknown mode", ThemeMode(99), "Unknown"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := tt.mode.String(); got != tt.want {
				t.Errorf("ThemeMode.String() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestThemeMode_IsLight(t *testing.T) {
	tests := []struct {
		name string
		mode ThemeMode
		want bool
	}{
		{"light mode returns true", ModeLight, true},
		{"dark mode returns false", ModeDark, false},
		{"system mode returns false", ModeSystem, false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := tt.mode.IsLight(); got != tt.want {
				t.Errorf("ThemeMode.IsLight() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestThemeMode_IsDark(t *testing.T) {
	tests := []struct {
		name string
		mode ThemeMode
		want bool
	}{
		{"light mode returns false", ModeLight, false},
		{"dark mode returns true", ModeDark, true},
		{"system mode returns false", ModeSystem, false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := tt.mode.IsDark(); got != tt.want {
				t.Errorf("ThemeMode.IsDark() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestThemeMode_IsSystem(t *testing.T) {
	tests := []struct {
		name string
		mode ThemeMode
		want bool
	}{
		{"light mode returns false", ModeLight, false},
		{"dark mode returns false", ModeDark, false},
		{"system mode returns true", ModeSystem, true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := tt.mode.IsSystem(); got != tt.want {
				t.Errorf("ThemeMode.IsSystem() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestThemeMode_ResolvedMode(t *testing.T) {
	tests := []struct {
		name        string
		mode        ThemeMode
		preferLight bool
		want        ThemeMode
	}{
		{"light mode stays light", ModeLight, true, ModeLight},
		{"light mode stays light regardless of pref", ModeLight, false, ModeLight},
		{"dark mode stays dark", ModeDark, true, ModeDark},
		{"dark mode stays dark regardless of pref", ModeDark, false, ModeDark},
		{"system mode resolves to light when preferred", ModeSystem, true, ModeLight},
		{"system mode resolves to dark when not preferred", ModeSystem, false, ModeDark},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := tt.mode.ResolvedMode(tt.preferLight); got != tt.want {
				t.Errorf("ThemeMode.ResolvedMode(%v) = %v, want %v", tt.preferLight, got, tt.want)
			}
		})
	}
}
