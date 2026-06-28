package auth

import (
	"errors"
	"fmt"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestGenerateJWT(t *testing.T) {
	InitTestKeys(t)

	t.Run("generates valid access token with user params", func(t *testing.T) {
		userID := int32(42)
		params := &GenerationParams{
			UserID: &userID,
			Roles:  []string{"user", "admin"},
			GTY:    GrantTypePassword,
		}
		token, err := GenerateJWT(params, AccessToken, 5*time.Minute)
		require.NoError(t, err)
		assert.NotEmpty(t, token)
		assert.Contains(t, token, ".")
	})

	t.Run("generates valid access token with service params", func(t *testing.T) {
		svcID := "test-service"
		params := &GenerationParams{
			ServiceID: &svcID,
			Roles:     []string{ServiceRole},
			GTY:       GrantTypeService,
		}
		token, err := GenerateJWT(params, AccessToken, 5*time.Minute)
		require.NoError(t, err)
		assert.NotEmpty(t, token)
	})

	t.Run("generates valid refresh token with JTI", func(t *testing.T) {
		userID := int32(1)
		params := &GenerationParams{
			UserID: &userID,
			Roles:  []string{"user"},
			GTY:    GrantTypeInitData,
			JTI:    "refresh-jti-123",
		}
		token, err := GenerateJWT(params, RefreshToken, 24*time.Hour)
		require.NoError(t, err)
		assert.NotEmpty(t, token)
		claims, err := ParseJWT(token)
		require.NoError(t, err)
		assert.Equal(t, "refresh-jti-123", claims.ID)
		assert.Equal(t, RefreshToken, claims.Type)
	})

	t.Run("includes audience when provided", func(t *testing.T) {
		svcID := "svc"
		params := &GenerationParams{
			ServiceID: &svcID,
			Roles:     []string{ServiceRole},
			GTY:       GrantTypeService,
			Audience:  "analysis-service",
		}
		token, err := GenerateJWT(params, AccessToken, 5*time.Minute)
		require.NoError(t, err)
		claims, err := ParseJWT(token)
		require.NoError(t, err)
		assert.Len(t, claims.Audience, 1)
		assert.Equal(t, "analysis-service", claims.Audience[0])
	})

	t.Run("includes telegram_id and max_id when provided", func(t *testing.T) {
		userID := int32(10)
		tgID := int64(12345)
		maxID := int64(67890)
		params := &GenerationParams{
			UserID:     &userID,
			TelegramID: &tgID,
			MaxID:      &maxID,
			Roles:      []string{"user"},
			GTY:        GrantTypeInitData,
		}
		token, err := GenerateJWT(params, AccessToken, 5*time.Minute)
		require.NoError(t, err)
		claims, err := ParseJWT(token)
		require.NoError(t, err)
		require.NotNil(t, claims.TelegramID)
		assert.Equal(t, int64(12345), *claims.TelegramID)
		require.NotNil(t, claims.MaxID)
		assert.Equal(t, int64(67890), *claims.MaxID)
	})
}

func TestParseJWT(t *testing.T) {
	InitTestKeys(t)

	t.Run("parses valid token returns claims", func(t *testing.T) {
		userID := int32(99)
		params := &GenerationParams{
			UserID: &userID,
			Roles:  []string{"user"},
			GTY:    GrantTypePassword,
		}
		token, err := GenerateJWT(params, AccessToken, 5*time.Minute)
		require.NoError(t, err)

		claims, err := ParseJWT(token)
		require.NoError(t, err)
		assert.Equal(t, int32(99), *claims.UserID)
		assert.Equal(t, []string{"user"}, claims.Roles)
		assert.Equal(t, GrantTypePassword, claims.GTY)
		assert.Equal(t, AccessToken, claims.Type)
		assert.Equal(t, "auth-service", claims.Issuer)
		assert.NotZero(t, claims.ExpiresAt)
		assert.NotZero(t, claims.IssuedAt)
	})

	t.Run("invalid token returns error", func(t *testing.T) {
		claims, err := ParseJWT("invalid.jwt.token")
		assert.Error(t, err)
		assert.Nil(t, claims)
		assert.Contains(t, err.Error(), "failed to parse token")
	})

	t.Run("empty token returns error", func(t *testing.T) {
		claims, err := ParseJWT("")
		assert.Error(t, err)
		assert.Nil(t, claims)
	})

	t.Run("malformed token parts returns error", func(t *testing.T) {
		claims, err := ParseJWT("not.a.valid.jwt")
		assert.Error(t, err)
		assert.Nil(t, claims)
	})
}

func TestGenerateJWT_ParseJWT_RoundTrip(t *testing.T) {
	InitTestKeys(t)

	params := &GenerationParams{
		UserID:     ptrInt32(100),
		ServiceID:  nil,
		TelegramID: ptrInt64(555),
		MaxID:      nil,
		Roles:      []string{"user", "admin"},
		GTY:        GrantTypeInitData,
		JTI:        "test-jti",
		Audience:   "test-audience",
	}

	token, err := GenerateJWT(params, AccessToken, 10*time.Minute)
	require.NoError(t, err)

	claims, err := ParseJWT(token)
	require.NoError(t, err)
	assert.Equal(t, int32(100), *claims.UserID)
	assert.Equal(t, int64(555), *claims.TelegramID)
	assert.Equal(t, []string{"user", "admin"}, claims.Roles)
	assert.Equal(t, GrantTypeInitData, claims.GTY)
	assert.Equal(t, AccessToken, claims.Type)
	assert.Equal(t, "test-jti", claims.ID)
	assert.Equal(t, "test-audience", claims.Audience[0])
}

func TestIssueMergeToken(t *testing.T) {
	InitTestKeys(t)

	t.Run("issues valid merge token", func(t *testing.T) {
		token, err := IssueMergeToken()
		require.NoError(t, err)
		assert.NotEmpty(t, token)

		claims, err := ParseJWT(token)
		require.NoError(t, err)
		assert.Equal(t, "auth-service", *claims.ServiceID)
		assert.Equal(t, []string{ServiceRole}, claims.Roles)
		assert.Equal(t, GrantTypeService, claims.GTY)
		assert.Equal(t, AccessToken, claims.Type)
	})

	t.Run("merge token is valid", func(t *testing.T) {
		token, err := IssueMergeToken()
		require.NoError(t, err)
		claims, err := ParseJWT(token)
		require.NoError(t, err)
		assert.NotNil(t, claims)
		assert.NotZero(t, claims.ExpiresAt)
	})
}

func ptrInt32(v int32) *int32 { return &v }
func ptrInt64(v int64) *int64 { return &v }

func TestJWTRejectReason_KeyIDMismatch(t *testing.T) {
	inner := &KeyIDMismatchError{Expected: "Rij8-y_9hEw", Got: "FG0viFRdMis"}
	wrapped := fmt.Errorf("failed to parse token: %w", inner)
	assert.Equal(t, "invalid_key_id", JWTRejectReason(wrapped))
	exp, tokKID, ok := KeyIDMismatchDetails(wrapped)
	require.True(t, ok)
	assert.Equal(t, "Rij8-y_9hEw", exp)
	assert.Equal(t, "FG0viFRdMis", tokKID)
}

func TestKeyIDMismatchDetails_LegacyErrorString(t *testing.T) {
	legacyErr := errors.New(
		"token is unverifiable: error while executing keyfunc: invalid key ID: expected Rij8-y_9hEw, got FG0viFRdMis",
	)
	assert.Equal(t, "invalid_key_id", JWTRejectReason(legacyErr))
	exp, tokKID, ok := KeyIDMismatchDetails(legacyErr)
	require.True(t, ok)
	assert.Equal(t, "Rij8-y_9hEw", exp)
	assert.Equal(t, "FG0viFRdMis", tokKID)
}
