package widget

// ThemeProvider gives widgets access to the current visual theme.
//
// Concrete theme types (theme.Theme, material3.Theme) implement this
// interface. The widget package defines only the interface to avoid
// import cycles between widget and theme packages.
//
// Widgets should use ThemeProvider for visual decisions (e.g., choosing
// colors based on dark/light mode, default text color) rather than
// importing a concrete theme package directly.
type ThemeProvider interface {
	// IsDark returns true if this is a dark theme.
	IsDark() bool

	// OnSurface returns the default color for text and icons on surface
	// backgrounds. Every design system (Material 3, Fluent, Cupertino)
	// defines an equivalent concept, making this a universal token.
	OnSurface() Color
}
