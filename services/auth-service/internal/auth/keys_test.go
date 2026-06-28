package auth

import (
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestGetKeyManager(t *testing.T) {
	t.Run("returns same instance on multiple calls", func(t *testing.T) {
		km1 := GetKeyManager()
		km2 := GetKeyManager()
		assert.Same(t, km1, km2, "GetKeyManager should return singleton instance")
	})

	t.Run("returns non-nil", func(t *testing.T) {
		km := GetKeyManager()
		assert.NotNil(t, km)
	})
}

func TestKeyManager_Initialize(t *testing.T) {
	km := GetKeyManager()

	if km.GetSigningKey() != nil {
		t.Run("already initialized returns error", func(t *testing.T) {
			dir := t.TempDir()
			privatePath := filepath.Join(dir, "private.pem")
			publicPath := filepath.Join(dir, "public.pem")
			err := km.Initialize(privatePath, publicPath)
			assert.Error(t, err)
			assert.Contains(t, err.Error(), "key manager already initialized")
		})
		return
	}

	t.Run("initializes with temp dir and generates keys", func(t *testing.T) {
		dir := t.TempDir()
		privatePath := filepath.Join(dir, "private.pem")
		publicPath := filepath.Join(dir, "public.pem")
		err := km.Initialize(privatePath, publicPath)
		require.NoError(t, err)
		assert.NotNil(t, km.GetSigningKey())
		assert.NotNil(t, km.GetPublicKey())
		assert.NotEmpty(t, km.GetKeyID())
	})
}

func TestKeyManager_GetSigningKey(t *testing.T) {
	InitTestKeys(t)

	km := GetKeyManager()
	key := km.GetSigningKey()
	require.NotNil(t, key)
	assert.Equal(t, 2048, key.N.BitLen())
}

func TestKeyManager_GetPublicKey(t *testing.T) {
	InitTestKeys(t)

	km := GetKeyManager()
	key := km.GetPublicKey()
	require.NotNil(t, key)
	assert.NotNil(t, key.N)
	assert.NotZero(t, key.E)
}

func TestKeyManager_GetKeyID(t *testing.T) {
	InitTestKeys(t)

	km := GetKeyManager()
	keyID := km.GetKeyID()
	assert.NotEmpty(t, keyID)
}

func TestKeyManager_GetJWKS(t *testing.T) {
	InitTestKeys(t)

	km := GetKeyManager()
	jwks := km.GetJWKS()
	require.NotNil(t, jwks)
	require.Len(t, jwks.Keys, 1)
	key := jwks.Keys[0]
	assert.Equal(t, "RSA", key.Kty)
	assert.Equal(t, "sig", key.Use)
	assert.Equal(t, km.GetKeyID(), key.Kid)
	assert.Equal(t, "RS256", key.Alg)
	assert.NotEmpty(t, key.N)
	assert.NotEmpty(t, key.E)
}
