package authz

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/go-jose/go-jose/v4"
	"github.com/go-jose/go-jose/v4/jwt"
	"github.com/valyala/fasthttp"
)

// GetUser loads a user by internal id.
func (c *Client) GetUser(ctx context.Context, userID int32) (*User, error) {
	var user User
	path := fmt.Sprintf("/users/%d", userID)
	err := c.doRequest(ctx, "GET", path, nil, &user, fasthttp.StatusOK)
	if err != nil {
		return nil, err
	}
	return &user, nil
}

// GetUserByTelegramID resolves a user by Telegram messenger id.
func (c *Client) GetUserByTelegramID(ctx context.Context, telegramID int64) (*User, error) {
	var user User
	path := fmt.Sprintf("/users/by-messenger-id/%d?platform=telegram", telegramID)
	err := c.doRequest(ctx, "GET", path, nil, &user, fasthttp.StatusOK)
	if err != nil {
		return nil, err
	}
	return &user, nil
}

// GetUserRoles returns role names for a user.
func (c *Client) GetUserRoles(ctx context.Context, userID int32) ([]Role, error) {
	var roles []Role
	path := fmt.Sprintf("/user-roles/user/%d", userID)
	err := c.doRequest(ctx, "GET", path, nil, &roles, fasthttp.StatusOK)
	return roles, err
}

// ValidateToken validates a JWT token locally using JWKS and returns the claims.
func (c *Client) ValidateToken(
	ctx context.Context,
	tokenStr string,
) (*TokenValidationResponse, error) {
	if resp, err := c.validateTokenWithJWKS(tokenStr); err == nil {
		return resp, nil
	}

	_, err, _ := c.jwksRefreshGroup.Do("jwks", func() (any, error) {
		return nil, c.UpdateJWKS(ctx)
	})
	if err != nil {
		return &TokenValidationResponse{Valid: false}, fmt.Errorf("failed to refresh JWKS: %w", err)
	}

	return c.validateTokenWithJWKS(tokenStr)
}

func (c *Client) validateTokenWithJWKS(tokenStr string) (*TokenValidationResponse, error) {
	c.mu.RLock()
	jwks := c.jwks
	c.mu.RUnlock()

	if jwks == nil {
		return &TokenValidationResponse{Valid: false}, errors.New("JWKS not initialized")
	}

	tok, err := jwt.ParseSigned(tokenStr, []jose.SignatureAlgorithm{jose.RS256})
	if err != nil {
		return &TokenValidationResponse{Valid: false}, fmt.Errorf("failed to parse token: %w", err)
	}

	var claims CustomClaims
	var stdClaims jwt.Claims
	if err := tok.Claims(jwks, &claims, &stdClaims); err != nil {
		return &TokenValidationResponse{
				Valid: false,
			}, fmt.Errorf(
				"token verification failed: %w",
				err,
			)
	}

	if err := stdClaims.Validate(jwt.Expected{
		Time: time.Now(),
	}); err != nil {
		return &TokenValidationResponse{
				Valid: false,
			}, fmt.Errorf(
				"claims validation failed: %w",
				err,
			)
	}

	identity := &Identity{
		UserID: claims.UserID,
		Roles:  claims.Roles,
	}
	if claims.TelegramID != nil {
		identity.TelegramID = claims.TelegramID
	}
	if claims.MaxID != nil {
		identity.MaxID = claims.MaxID
	}

	return &TokenValidationResponse{
		Valid:    true,
		Identity: identity,
		Roles:    claims.Roles,
	}, nil
}
