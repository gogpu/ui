package textfield

// InputType represents the type of text input.
// It determines input behavior such as masking and keyboard hints.
type InputType uint8

// Input type constants.
const (
	// TypeText is general-purpose text input (default).
	TypeText InputType = iota

	// TypePassword masks input characters with dots.
	TypePassword

	// TypeEmail is for email address input.
	TypeEmail

	// TypeNumber is for numeric input.
	TypeNumber

	// TypeSearch is for search input.
	TypeSearch
)

// String returns a human-readable name for the input type.
func (t InputType) String() string {
	switch t {
	case TypeText:
		return "Text"
	case TypePassword:
		return "Password"
	case TypeEmail:
		return "Email"
	case TypeNumber:
		return "Number"
	case TypeSearch:
		return "Search"
	default:
		return "Unknown"
	}
}

// ValidationFunc validates the text field value and returns an error message.
// An empty string means the value is valid.
type ValidationFunc func(value string) string
