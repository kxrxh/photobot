package auth

import (
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/require"
)

func InitTestKeys(t *testing.T) {
	t.Helper()
	km := GetKeyManager()
	if km.GetSigningKey() != nil {
		return
	}
	dir := t.TempDir()
	privatePath := filepath.Join(dir, "private.pem")
	publicPath := filepath.Join(dir, "public.pem")
	err := km.Initialize(privatePath, publicPath)
	if err != nil && err.Error() == "key manager already initialized" {
		return
	}
	require.NoError(t, err)
}
