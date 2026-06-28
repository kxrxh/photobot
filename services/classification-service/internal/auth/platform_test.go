package auth

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNormalizeMessengerPlatform(t *testing.T) {
	got, err := normalizeMessengerPlatform(" Telegram ")
	require.NoError(t, err)
	assert.Equal(t, "telegram", got)

	_, err = normalizeMessengerPlatform("whatsapp")
	require.Error(t, err)
}
