package format

import "fmt"

// StatusError formats an error message for status bar display.
// Example: StatusError("saving", err) -> "Error saving: file not found"
func StatusError(context string, err error) string {
	if err == nil {
		return ""
	}
	return fmt.Sprintf("Error %s: %v", context, err)
}

// StatusSuccess formats a success message for status bar display.
// Example: StatusSuccess("Added", "BTC purchase") -> "Added BTC purchase!"
func StatusSuccess(action string, details string) string {
	if details == "" {
		return action + "!"
	}
	return fmt.Sprintf("%s %s!", action, details)
}

// ValidationError formats a validation error message.
// Example: ValidationError("amount", "must be positive") -> "Invalid amount: must be positive"
func ValidationError(field string, reason string) string {
	return fmt.Sprintf("Invalid %s: %s", field, reason)
}

// RequiredFieldError formats a required field error.
// Example: RequiredFieldError("coin") -> "Coin is required"
func RequiredFieldError(field string) string {
	return field + " is required"
}

// NotFoundError formats a not found error.
// Example: NotFoundError("stake") -> "stake not found"
func NotFoundError(item string) string {
	return item + " not found"
}
