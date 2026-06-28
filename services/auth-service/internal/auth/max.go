package auth

import (
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"net/url"
	"sort"
	"strconv"
	"strings"
	"time"

	"csort.ru/auth-service/internal/logger"
)

var maxLogger = logger.GetLogger("auth.max")

// ValidateMaxData validates the MAX initData using the specified bot's token (login TTL).
func ValidateMaxData(initData, botToken string, debug bool) (*UserData, error) {
	return ValidateMaxDataWithMaxAge(initData, botToken, MaxInitDataAge, debug)
}

// ValidateMaxDataWithMaxAge validates MAX WebApp init data using the specified bot's token
// and a custom maximum age. The algorithm is based on the official documentation:
// https://dev.max.ru/docs/webapps/validation
func ValidateMaxDataWithMaxAge(
	initData, botToken string,
	maxAge time.Duration,
	debug bool,
) (*UserData, error) {
	if debug {
		user, err := parseMaxUserFromInitData(initData)
		if err != nil {
			return nil, fmt.Errorf("failed to parse MAX init data in debug mode: %w", err)
		}

		if user.ID == 0 {
			return nil, errors.New("user data not found in MAX init data (debug)")
		}

		maxLogger.Debug().
			Int64("max_id", user.ID).
			Msg("validate max init data completed (debug, no signature)")

		return user, nil
	}

	user, err := validateAndParseMaxInitData(initData, botToken, maxAge)
	if err != nil {
		maxLogger.Error().
			Err(err).
			Msg("validate max init data failed")
		return nil, fmt.Errorf("max init data validation failed: %w", err)
	}

	maxLogger.Info().
		Int64("max_id", user.ID).
		Str("username", user.Username).
		Msg("validate max init data completed")

	return user, nil
}

func validateAndParseMaxInitData(
	rawInitData, botToken string,
	maxAge time.Duration,
) (*UserData, error) {
	decoded, err := url.QueryUnescape(rawInitData)
	if err != nil {
		return nil, fmt.Errorf("failed to URL-decode MAX init data: %w", err)
	}

	var (
		hashValue      string
		authDateString string
		pairs          []struct {
			key   string
			value string
		}
	)

	for _, part := range strings.Split(decoded, "&") {
		if part == "" {
			continue
		}
		kv := strings.SplitN(part, "=", 2)
		if len(kv) != 2 {
			continue
		}
		key := kv[0]
		value := kv[1]

		if key == "hash" {
			hashValue = value
			continue
		}
		if key == "auth_date" {
			authDateString = value
		}

		pairs = append(pairs, struct {
			key   string
			value string
		}{key: key, value: value})
	}

	if hashValue == "" {
		return nil, errors.New("hash field is missing in MAX init data")
	}

	sort.Slice(pairs, func(i, j int) bool {
		return pairs[i].key < pairs[j].key
	})

	builder := strings.Builder{}
	for i, kv := range pairs {
		if i > 0 {
			builder.WriteByte('\n')
		}
		builder.WriteString(kv.key)
		builder.WriteByte('=')
		builder.WriteString(kv.value)
	}
	dataCheckString := builder.String()

	secretKey := hmac.New(sha256.New, []byte("WebAppData"))
	secretKey.Write([]byte(botToken))
	secretKeyHex := hex.EncodeToString(secretKey.Sum(nil))

	secretKeyBytes, err := hex.DecodeString(secretKeyHex)
	if err != nil {
		return nil, fmt.Errorf("failed to decode derived secret key: %w", err)
	}

	mac := hmac.New(sha256.New, secretKeyBytes)
	mac.Write([]byte(dataCheckString))
	expectedHash := hex.EncodeToString(mac.Sum(nil))

	if !hmac.Equal([]byte(expectedHash), []byte(hashValue)) {
		return nil, errors.New("invalid MAX init data signature")
	}

	if authDateString != "" {
		authTime, err := parseMaxAuthDate(authDateString)
		if err != nil {
			return nil, fmt.Errorf("failed to parse auth_date: %w", err)
		}
		if time.Since(authTime) > maxAge {
			return nil, errors.New("max init data has expired")
		}
	}

	return parseMaxUserFromInitData(decoded)
}

func parseMaxAuthDate(value string) (time.Time, error) {
	ts, err := strconv.ParseInt(value, 10, 64)
	if err != nil {
		return time.Time{}, err
	}

	// Large values are treated as Unix milliseconds.
	if ts > 1_000_000_000_000 {
		return time.UnixMilli(ts), nil
	}
	return time.Unix(ts, 0), nil
}

func parseMaxUserFromInitData(initData string) (*UserData, error) {
	decoded, err := url.QueryUnescape(initData)
	if err != nil {
		decoded = initData
	}

	var userJSON string
	for _, part := range strings.Split(decoded, "&") {
		if part == "" {
			continue
		}
		kv := strings.SplitN(part, "=", 2)
		if len(kv) != 2 {
			continue
		}
		if kv[0] == "user" {
			userJSON = kv[1]
			break
		}
	}

	if userJSON == "" {
		return nil, errors.New("user field is missing in MAX init data")
	}

	var raw struct {
		ID        int64   `json:"id"`
		FirstName string  `json:"first_name"`
		LastName  string  `json:"last_name"`
		Username  *string `json:"username"`
		PhotoURL  *string `json:"photo_url"`
	}

	if err := json.Unmarshal([]byte(userJSON), &raw); err != nil {
		return nil, fmt.Errorf("failed to unmarshal MAX user JSON: %w", err)
	}

	user := &UserData{
		ID:        raw.ID,
		FirstName: raw.FirstName,
		LastName:  raw.LastName,
	}
	if raw.Username != nil {
		user.Username = *raw.Username
	}
	if raw.PhotoURL != nil {
		user.PhotoURL = *raw.PhotoURL
	}

	return user, nil
}
