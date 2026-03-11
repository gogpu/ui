// Package fonts embeds professional UI font files.
//
// Inter is a typeface designed specifically for computer screens by Rasmus
// Andersson. It features tall x-height for readability at small sizes, and
// is used by GitHub, Figma, and many other modern applications.
//
// License: SIL Open Font License 1.1 (compatible with MIT).
// Source: https://github.com/rsms/inter
package fonts

import _ "embed"

// InterRegular contains the Inter Regular (400) TTF font data.
//
//go:embed Inter-Regular.ttf
var InterRegular []byte

// InterBold contains the Inter Bold (700) TTF font data.
//
//go:embed Inter-Bold.ttf
var InterBold []byte
