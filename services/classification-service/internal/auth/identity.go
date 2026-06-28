package auth

import (
	"context"
	"fmt"

	"github.com/gofiber/fiber/v3"
)

func (c *Client) GetUserByMessengerID(
	ctx context.Context,
	messengerID int64,
	platform string,
	token string,
) (*User, error) {
	normalizedPlatform, err := normalizeMessengerPlatform(platform)
	if err != nil {
		return nil, err
	}

	var user User
	path := fmt.Sprintf("/users/by-messenger-id/%d?platform=%s", messengerID, normalizedPlatform)
	err = c.doRequestWithAuthHeader(
		ctx,
		"GET",
		path,
		nil,
		&user,
		"Authorization",
		"Bearer "+token,
		fiber.StatusOK,
	)
	if err != nil {
		return nil, err
	}
	return &user, nil
}
