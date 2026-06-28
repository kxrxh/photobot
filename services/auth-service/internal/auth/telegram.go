package auth

import (
	"errors"
	"fmt"
	"time"

	"csort.ru/auth-service/internal/logger"
	initdata "github.com/telegram-mini-apps/init-data-golang"
)

const MaxInitDataAge = 15 * time.Minute

var telegramLogger = logger.GetLogger("auth.telegram")

// UserData holds the extracted user information from Telegram initData
type UserData struct {
	ID        int64  `json:"id"`
	FirstName string `json:"first_name"`
	LastName  string `json:"last_name,omitempty"`
	Username  string `json:"username,omitempty"`
	PhotoURL  string `json:"photo_url,omitempty"`
}

// ValidateTelegramData validates the Telegram initData using the specified bot's token (login TTL).
func ValidateTelegramData(initData, botToken string, debug bool) (*UserData, error) {
	return ValidateTelegramDataWithMaxAge(initData, botToken, MaxInitDataAge, debug)
}

// ValidateTelegramDataWithMaxAge validates the Telegram initData with a custom max age.
func ValidateTelegramDataWithMaxAge(
	initData, botToken string,
	maxAge time.Duration,
	debug bool,
) (*UserData, error) {
	if debug {
		parsedData, err := initdata.Parse(initData)
		if err != nil {
			return nil, fmt.Errorf("failed to parse initData in debug mode: %w", err)
		}

		if parsedData.User.ID == 0 {
			return nil, errors.New("user data not found in debug mode")
		}

		telegramLogger.Debug().
			Int64("telegram_id", parsedData.User.ID).
			Msg("Allowing mock validation in debug mode")

		return &UserData{
			ID:        parsedData.User.ID,
			FirstName: parsedData.User.FirstName,
			LastName:  parsedData.User.LastName,
			Username:  parsedData.User.Username,
			PhotoURL:  parsedData.User.PhotoURL,
		}, nil
	}

	err := initdata.Validate(initData, botToken, maxAge)
	if err != nil {
		telegramLogger.Error().
			Err(err).
			Msg("Telegram data validation failed")
		return nil, fmt.Errorf("telegram data validation failed: %w", err)
	}

	parsedData, err := initdata.Parse(initData)
	if err != nil {
		return nil, fmt.Errorf("failed to parse validated initData: %w", err)
	}

	if parsedData.User.ID == 0 {
		return nil, errors.New("user data not found")
	}

	userData := &UserData{
		ID:        parsedData.User.ID,
		FirstName: parsedData.User.FirstName,
		LastName:  parsedData.User.LastName,
		Username:  parsedData.User.Username,
		PhotoURL:  parsedData.User.PhotoURL,
	}

	telegramLogger.Info().
		Int64("telegram_id", userData.ID).
		Str("username", userData.Username).
		Msg("Successfully validated Telegram data")

	return userData, nil
}
