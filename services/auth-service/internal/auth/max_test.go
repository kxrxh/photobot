package auth

import (
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"net/url"
	"sort"
	"strconv"
	"strings"
	"testing"
	"time"

	"csort.ru/auth-service/internal/testutil"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func buildValidMaxInitData(t *testing.T, token string, authDate int64, userJSON string) string {
	t.Helper()
	ts := strconv.FormatInt(authDate, 10)
	pairs := []struct{ k, v string }{
		{"auth_date", ts},
		{"user", userJSON},
	}
	sort.Slice(pairs, func(i, j int) bool { return pairs[i].k < pairs[j].k })
	var b strings.Builder
	for i, p := range pairs {
		if i > 0 {
			b.WriteByte('\n')
		}
		b.WriteString(p.k)
		b.WriteByte('=')
		b.WriteString(p.v)
	}
	sk := hmac.New(sha256.New, []byte("WebAppData"))
	sk.Write([]byte(token))
	mac := hmac.New(sha256.New, sk.Sum(nil))
	mac.Write([]byte(b.String()))
	return "auth_date=" + url.QueryEscape(
		ts,
	) + "&user=" + url.QueryEscape(
		userJSON,
	) + "&hash=" + hex.EncodeToString(
		mac.Sum(nil),
	)
}

func TestValidateMaxData_DebugMode(t *testing.T) {
	botToken := "test-bot-token"
	userJSON := `{"id":12345,"first_name":"John","last_name":"Doe","username":"johndoe"}`

	t.Run("valid user in debug mode", func(t *testing.T) {
		initData := "user=" + url.QueryEscape(userJSON)
		user, err := ValidateMaxData(initData, botToken, true)
		require.NoError(t, err)
		assert.NotNil(t, user)
		assert.Equal(t, int64(12345), user.ID)
		assert.Equal(t, "John", user.FirstName)
		assert.Equal(t, "Doe", user.LastName)
		assert.Equal(t, "johndoe", user.Username)
	})

	t.Run("user ID zero in debug mode returns error", func(t *testing.T) {
		initData := "user=" + url.QueryEscape(`{"id":0,"first_name":"Test"}`)
		user, err := ValidateMaxData(initData, botToken, true)
		assert.Error(t, err)
		assert.Nil(t, user)
		assert.Contains(t, err.Error(), "user data not found")
	})

	t.Run("missing user field in debug mode returns error", func(t *testing.T) {
		initData := "auth_date=1234567890"
		user, err := ValidateMaxData(initData, botToken, true)
		assert.Error(t, err)
		assert.Nil(t, user)
		assert.Contains(t, err.Error(), "user field is missing")
	})

	t.Run("invalid user JSON in debug mode returns error", func(t *testing.T) {
		initData := "user=" + url.QueryEscape(`{invalid json}`)
		user, err := ValidateMaxData(initData, botToken, true)
		assert.Error(t, err)
		assert.Nil(t, user)
		assert.Contains(t, err.Error(), "failed to unmarshal")
	})

	t.Run("user with optional fields nil", func(t *testing.T) {
		initData := "user=" + url.QueryEscape(`{"id":999,"first_name":"Jane","last_name":""}`)
		user, err := ValidateMaxData(initData, botToken, true)
		require.NoError(t, err)
		assert.NotNil(t, user)
		assert.Equal(t, int64(999), user.ID)
		assert.Equal(t, "Jane", user.FirstName)
		assert.Equal(t, "", user.LastName)
		assert.Equal(t, "", user.Username)
	})
}

func TestValidateMaxDataWithMaxAge_NonDebugMode(t *testing.T) {
	botToken := "test-bot-token-123"
	userJSON := `{"id":42,"first_name":"Alice","last_name":"Smith","username":"alice"}`
	initData := testutil.BuildValidInitData(botToken, userJSON)

	t.Run("valid signature and recent auth_date", func(t *testing.T) {
		user, err := ValidateMaxDataWithMaxAge(initData, botToken, 5*time.Minute, false)
		require.NoError(t, err)
		assert.NotNil(t, user)
		assert.Equal(t, int64(42), user.ID)
		assert.Equal(t, "Alice", user.FirstName)
		assert.Equal(t, "alice", user.Username)
	})

	t.Run("invalid signature returns error", func(t *testing.T) {
		badInitData := "auth_date=" + url.QueryEscape(
			"1234567890",
		) + "&user=" + url.QueryEscape(
			userJSON,
		) + "&hash=invalidhash"
		user, err := ValidateMaxDataWithMaxAge(badInitData, botToken, 5*time.Minute, false)
		assert.Error(t, err)
		assert.Nil(t, user)
		assert.Contains(t, err.Error(), "signature")
	})

	t.Run("expired auth_date returns error", func(t *testing.T) {
		oldAuthDate := time.Now().Add(-10 * time.Minute).Unix()
		expiredInitData := buildValidMaxInitData(t, botToken, oldAuthDate, userJSON)
		user, err := ValidateMaxDataWithMaxAge(expiredInitData, botToken, 2*time.Minute, false)
		assert.Error(t, err)
		assert.Nil(t, user)
		assert.Contains(t, err.Error(), "expired")
	})

	t.Run("missing hash returns error", func(t *testing.T) {
		noHash := "auth_date=1234567890&user=" + url.QueryEscape(userJSON)
		user, err := ValidateMaxDataWithMaxAge(noHash, botToken, 5*time.Minute, false)
		assert.Error(t, err)
		assert.Nil(t, user)
		assert.Contains(t, err.Error(), "hash")
	})

	t.Run("wrong bot token fails signature", func(t *testing.T) {
		user, err := ValidateMaxDataWithMaxAge(initData, "wrong-token", 5*time.Minute, false)
		assert.Error(t, err)
		assert.Nil(t, user)
	})
}

func TestValidateMaxDataWithMaxAge_AuthDateMilliseconds(t *testing.T) {
	botToken := "bot"
	userJSON := `{"id":1,"first_name":"X"}`

	authDateMs := time.Now().UnixMilli()
	initData := buildValidMaxInitData(t, botToken, authDateMs, userJSON)

	user, err := ValidateMaxDataWithMaxAge(initData, botToken, 5*time.Minute, false)
	require.NoError(t, err)
	assert.NotNil(t, user)
	assert.Equal(t, int64(1), user.ID)
}

func TestValidateMaxDataWithMaxAge_AuthDateSeconds(t *testing.T) {
	botToken := "bot"
	userJSON := `{"id":2,"first_name":"Y"}`
	initData := testutil.BuildValidInitData(botToken, userJSON)

	user, err := ValidateMaxDataWithMaxAge(initData, botToken, 5*time.Minute, false)
	require.NoError(t, err)
	assert.NotNil(t, user)
	assert.Equal(t, int64(2), user.ID)
}
