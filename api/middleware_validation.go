package main

import (
	"regexp"
	"strings"
	"unicode/utf8"
)

// Common validation patterns
var (
	validationEmailRegex = regexp.MustCompile(`^[a-zA-Z0-9._%+-]+@[a-zA-Z0-9.-]+\.[a-zA-Z]{2,}$`)
	validationUuidRegex  = regexp.MustCompile(`^[0-9a-f]{8}-[0-9a-f]{4}-[0-9a-f]{4}-[0-9a-f]{4}-[0-9a-f]{12}$`)
)

// FieldValidator defines a validation function for a specific field
type FieldValidator func(value interface{}) error

// Common field validators
func ValidateEmail(value interface{}) error {
	email, ok := value.(string)
	if !ok {
		return ErrValidationInvalidInput("email must be a string")
	}
	if !validationEmailRegex.MatchString(email) {
		return ErrValidationInvalidInput("invalid email format")
	}
	if len(email) > 254 {
		return ErrValidationInvalidInput("email too long")
	}
	return nil
}

func ValidateUUID(value interface{}) error {
	uuid, ok := value.(string)
	if !ok {
		return ErrValidationInvalidInput("UUID must be a string")
	}
	if !validationUuidRegex.MatchString(uuid) {
		return ErrValidationInvalidInput("invalid UUID format")
	}
	return nil
}

func ValidateStringLength(min, max int) FieldValidator {
	return func(value interface{}) error {
		str, ok := value.(string)
		if !ok {
			return ErrValidationInvalidInput("value must be a string")
		}
		if !utf8.ValidString(str) {
			return ErrValidationInvalidInput("invalid UTF-8 string")
		}
		if len(str) < min {
			return ErrValidationInvalidInput("string too short")
		}
		if len(str) > max {
			return ErrValidationInvalidInput("string too long")
		}
		return nil
	}
}

func ValidateNonEmpty(value interface{}) error {
	str, ok := value.(string)
	if !ok {
		return ErrValidationInvalidInput("value must be a string")
	}
	if strings.TrimSpace(str) == "" {
		return ErrValidationInvalidInput("value cannot be empty")
	}
	return nil
}
