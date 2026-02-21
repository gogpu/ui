package textfield

// runValidation executes all configured validation functions against the value.
// Returns the first non-empty error message, or empty string if all validations pass.
func runValidation(validators []ValidationFunc, value string) string {
	for _, fn := range validators {
		if msg := fn(value); msg != "" {
			return msg
		}
	}
	return ""
}
