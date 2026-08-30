// Package validation contains transport-boundary validation shared by HTTP
// request types. Application services must still enforce business invariants.
package validation

import (
	"fmt"
	"strings"
	"unicode/utf8"

	"github.com/google/uuid"
)

const (
	MaxTopicLength     = 200
	MaxModelNameLength = 200
	MaxTokenCount      = 1_000_000
	MaxPageSize        = 500
)

func UUID(field, value string, required bool) error {
	value = strings.TrimSpace(value)
	if value == "" {
		if required {
			return fmt.Errorf("%s is required", field)
		}
		return nil
	}
	if _, err := uuid.Parse(value); err != nil {
		return fmt.Errorf("%s must be a valid UUID", field)
	}
	return nil
}

func Topic(value string, required bool) error {
	value = strings.TrimSpace(value)
	if value == "" && required {
		return fmt.Errorf("topic is required")
	}
	if utf8.RuneCountInString(value) > MaxTopicLength {
		return fmt.Errorf("topic must be at most %d characters", MaxTopicLength)
	}
	return nil
}

func ModelName(field, value string, required bool) error {
	value = strings.TrimSpace(value)
	if value == "" {
		if required {
			return fmt.Errorf("%s is required", field)
		}
		return nil
	}
	if utf8.RuneCountInString(value) > MaxModelNameLength {
		return fmt.Errorf("%s must be at most %d characters", field, MaxModelNameLength)
	}
	for _, r := range value {
		if r < 0x20 || r == 0x7f {
			return fmt.Errorf("%s must not contain control characters", field)
		}
	}
	return nil
}

func TokenCount(field string, value int32, allowZero bool) error {
	minimum := int32(1)
	if allowZero {
		minimum = 0
	}
	if value < minimum || value > MaxTokenCount {
		return fmt.Errorf("%s must be between %d and %d", field, minimum, MaxTokenCount)
	}
	return nil
}
