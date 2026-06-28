package authz

import "context"

type IdentityClient interface {
	ValidateToken(ctx context.Context, tokenStr string) (*TokenValidationResponse, error)
	GetUser(ctx context.Context, userID int32) (*User, error)
	GetUserRoles(ctx context.Context, userID int32) ([]Role, error)
	GetUserByTelegramID(ctx context.Context, telegramID int64) (*User, error)
}

var _ IdentityClient = (*Client)(nil)
