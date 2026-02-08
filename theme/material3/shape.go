package material3

// ShapeScale holds Material 3 corner radius values.
//
// Material 3 defines a shape scale that ranges from no rounding (None)
// to fully circular (Full). These values are used for consistent
// corner rounding across UI components.
//
// Reference: https://m3.material.io/styles/shape/overview
type ShapeScale struct {
	// None is no rounding (sharp corners): 0dp.
	None float32

	// ExtraSmall is subtle rounding: 4dp.
	// Used for small components like chips and text fields.
	ExtraSmall float32

	// Small is moderate rounding: 8dp.
	// Used for buttons and small cards.
	Small float32

	// Medium is standard rounding: 12dp.
	// Used for cards, dialogs, and menus.
	Medium float32

	// Large is generous rounding: 16dp.
	// Used for large cards and sheets.
	Large float32

	// ExtraLarge is very generous rounding: 28dp.
	// Used for floating action buttons and large containers.
	ExtraLarge float32

	// Full is fully rounded: 9999dp.
	// Creates circular or pill shapes regardless of element size.
	Full float32
}

// DefaultShapeScale returns the standard Material 3 shape scale.
//
// Values follow the M3 specification:
//
//	None:       0dp
//	ExtraSmall: 4dp
//	Small:      8dp
//	Medium:     12dp
//	Large:      16dp
//	ExtraLarge: 28dp
//	Full:       9999dp (fully rounded)
func DefaultShapeScale() ShapeScale {
	return ShapeScale{
		None:       0,
		ExtraSmall: 4,
		Small:      8,
		Medium:     12,
		Large:      16,
		ExtraLarge: 28,
		Full:       9999,
	}
}
