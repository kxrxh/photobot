package auth

import (
	"errors"
	"strings"
)

func normalizeMessengerPlatform(platform string) (string, error) {
	normalized := strings.TrimSpace(strings.ToLower(platform))
	if normalized != "telegram" && normalized != "max" {
		return "", errors.New("platform must be 'telegram' or 'max'")
	}
	return normalized, nil
}
