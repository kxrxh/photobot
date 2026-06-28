package auth

import (
	"net/url"
	"testing"
	"time"

	"csort.ru/auth-service/internal/testutil"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestValidateTelegramData_DebugMode(t *testing.T) {
	botToken := "test-bot-token"
	userJSON := `{"id":12345,"first_name":"John","last_name":"Doe","username":"johndoe"}`

	t.Run("valid user in debug mode", func(t *testing.T) {
		initData := "user=" + url.QueryEscape(userJSON)
		user, err := ValidateTelegramData(initData, botToken, true)
		require.NoError(t, err)
		assert.NotNil(t, user)
		assert.Equal(t, int64(12345), user.ID)
		assert.Equal(t, "John", user.FirstName)
		assert.Equal(t, "Doe", user.LastName)
		assert.Equal(t, "johndoe", user.Username)
	})

	t.Run("user ID zero in debug mode returns error", func(t *testing.T) {
		initData := "user=" + url.QueryEscape(`{"id":0,"first_name":"Test"}`)
		user, err := ValidateTelegramData(initData, botToken, true)
		assert.Error(t, err)
		assert.Nil(t, user)
		assert.Contains(t, err.Error(), "user data not found")
	})

	t.Run("missing user field in debug mode returns error", func(t *testing.T) {
		initData := "auth_date=1234567890"
		user, err := ValidateTelegramData(initData, botToken, true)
		assert.Error(t, err)
		assert.Nil(t, user)
	})

	t.Run("invalid user JSON in debug mode returns error", func(t *testing.T) {
		initData := "user=" + url.QueryEscape(`{invalid json}`)
		user, err := ValidateTelegramData(initData, botToken, true)
		assert.Error(t, err)
		assert.Nil(t, user)
		assert.Contains(t, err.Error(), "failed to parse")
	})

	t.Run("user with optional fields", func(t *testing.T) {
		initData := "user=" + url.QueryEscape(
			`{"id":999,"first_name":"Jane","last_name":"","username":"jane"}`,
		)
		user, err := ValidateTelegramData(initData, botToken, true)
		require.NoError(t, err)
		assert.NotNil(t, user)
		assert.Equal(t, int64(999), user.ID)
		assert.Equal(t, "Jane", user.FirstName)
		assert.Equal(t, "", user.LastName)
		assert.Equal(t, "jane", user.Username)
	})
}

func TestValidateTelegramDataWithMaxAge_DebugMode(t *testing.T) {
	botToken := "test-bot-token"
	userJSON := `{"id":42,"first_name":"Alice","last_name":"Smith","username":"alice"}`

	t.Run("valid user with max age in debug mode", func(t *testing.T) {
		initData := "user=" + url.QueryEscape(userJSON)
		user, err := ValidateTelegramDataWithMaxAge(initData, botToken, 5*time.Minute, true)
		require.NoError(t, err)
		assert.NotNil(t, user)
		assert.Equal(t, int64(42), user.ID)
		assert.Equal(t, "Alice", user.FirstName)
		assert.Equal(t, "alice", user.Username)
	})

	t.Run("zero max age in debug mode still parses", func(t *testing.T) {
		initData := "user=" + url.QueryEscape(userJSON)
		user, err := ValidateTelegramDataWithMaxAge(initData, botToken, 0, true)
		require.NoError(t, err)
		assert.NotNil(t, user)
		assert.Equal(t, int64(42), user.ID)
	})
}

func TestValidateTelegramData_NonDebugMode(t *testing.T) {
	t.Run("valid signed init data passes", func(t *testing.T) {
		botToken := "1234567890:ABCdefGHIjklMNOpqrsTUVwxyz"
		userJSON := `{"id":42,"first_name":"Alice","last_name":"Smith","username":"alice"}`
		initData := testutil.BuildValidInitData(botToken, userJSON)
		user, err := ValidateTelegramData(initData, botToken, false)
		require.NoError(t, err)
		assert.NotNil(t, user)
		assert.Equal(t, int64(42), user.ID)
		assert.Equal(t, "Alice", user.FirstName)
		assert.Equal(t, "alice", user.Username)
	})

	t.Run("invalid init data without valid bot signature returns error", func(t *testing.T) {
		initData := "user=" + url.QueryEscape(`{"id":123,"first_name":"Test"}`)
		user, err := ValidateTelegramData(initData, "wrong-token", false)
		assert.Error(t, err)
		assert.Nil(t, user)
		assert.Contains(t, err.Error(), "telegram data validation failed")
	})

	t.Run("empty init data returns error", func(t *testing.T) {
		user, err := ValidateTelegramData("", "test-token", false)
		assert.Error(t, err)
		assert.Nil(t, user)
	})
}
