package auth

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestIsValidGrantType(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected bool
	}{
		{"initdata grant type", "initdata", true},
		{"client_credentials grant type", "client_credentials", true},
		{"password grant type", "password", true},
		{"empty string", "", false},
		{"invalid grant type", "invalid", false},
		{"oauth2 implicit", "implicit", false},
		{"case sensitive - Telegram", "Telegram", false},
		{"partial match", "pass", false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := IsValidGrantType(tt.input)
			assert.Equal(
				t,
				tt.expected,
				got,
				"IsValidGrantType(%q) = %v, want %v",
				tt.input,
				got,
				tt.expected,
			)
		})
	}
}

func TestGrantType_String(t *testing.T) {
	tests := []struct {
		name     string
		gt       GrantType
		expected string
	}{
		{"initdata", GrantTypeInitData, "initdata"},
		{"service", GrantTypeService, "client_credentials"},
		{"password", GrantTypePassword, "password"},
		{"empty", GrantType(""), ""},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := tt.gt.String()
			assert.Equal(t, tt.expected, got, "GrantType.String() = %q, want %q", got, tt.expected)
		})
	}
}
