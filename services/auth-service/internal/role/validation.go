package role

import (
	"errors"
	"unicode"
)

// ValidateRoleName checks role name length and allowed characters.
func ValidateRoleName(name string) error {
	if len(name) < 2 {
		return errors.New("role name must be at least 2 characters long")
	}
	if len(name) > 50 {
		return errors.New("role name cannot exceed 50 characters")
	}

	for _, char := range name {
		if !unicode.IsLetter(char) && !unicode.IsDigit(char) && char != '_' && char != '-' {
			return errors.New(
				"role name can only contain letters, numbers, underscores, and hyphens",
			)
		}
	}
	return nil
}
