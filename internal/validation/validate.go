package validation

import (
	"fmt"
	"github.com/go-playground/validator/v10"
	"strings"
)

// Create a global validator instance
var validate = validator.New()

// Validate function that takes any struct and validates it
func Validate(i interface{}) error {
	err := validate.Struct(i)
	if err != nil {
		var errMessages []string
		// Loop through the validation errors and format them
		for _, e := range err.(validator.ValidationErrors) {
			// Format the error message to be user-friendly
			errMessages = append(errMessages, fmt.Sprintf("%s is required", strings.ToLower(e.Field())))
		}
		// Return all the errors as a single string
		return fmt.Errorf("%s", strings.Join(errMessages, ", "))
	}
	return nil
}
