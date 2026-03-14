// Package font provides a font registry for managing font families, weights,
// and styles in the gogpu/ui toolkit.
//
// The registry maps family+weight+style combinations to embedded font data,
// enabling resolution of the best available font face for a given request.
// Weight resolution follows the CSS font-matching algorithm
// (https://www.w3.org/TR/css-fonts-4/#font-style-matching).
//
// # Usage
//
// Create a registry and register font faces:
//
//	reg := font.NewRegistry()
//	reg.RegisterFamily(font.Family{
//	    Name: "Inter",
//	    Faces: []font.Face{
//	        {Weight: font.Regular, Style: font.Normal, Data: interRegularTTF},
//	        {Weight: font.Bold, Style: font.Normal, Data: interBoldTTF},
//	    },
//	})
//
// Resolve the best matching face:
//
//	data, ok := reg.Resolve("Inter", font.Medium, font.Normal)
//	if ok {
//	    // Use font data for rendering
//	}
//
// The registry is safe for concurrent use.
package font
