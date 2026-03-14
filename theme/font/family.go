package font

// Family describes a font family with all its available faces.
//
// A family groups multiple Face entries under a single name (e.g., "Inter").
// Each face represents a specific weight+style combination with the
// corresponding font data.
type Family struct {
	// Name is the family name (e.g., "Inter", "Roboto").
	Name string

	// Faces is the list of available weight+style variants.
	Faces []Face
}

// Face represents a single font variant within a family.
//
// A face combines a specific weight and style with the raw font data
// (typically TTF or OTF bytes).
type Face struct {
	// Weight is the font weight for this face.
	Weight Weight

	// Style is the font style for this face.
	Style Style

	// Data is the raw font file data (TTF or OTF).
	Data []byte
}
