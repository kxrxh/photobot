package auth

import (
	"context"

	"csort.ru/auth-service/internal/database"
	"csort.ru/auth-service/internal/user"
)

type BotTokenProvider interface {
	GetTokenByNameAndPlatform(ctx context.Context, name, platform string) (string, error)
	ListTokensByPlatform(ctx context.Context, platform string) ([]string, error)
}

type UserProvider interface {
	Get(ctx context.Context, id int32) (*user.User, error)
	GetByTelegramId(ctx context.Context, telegramId int64) (*user.User, error)
	GetByMaxId(ctx context.Context, maxId int64) (*user.User, error)
	GetByLogin(ctx context.Context, login string) (*user.User, error)
}

type RoleProvider interface {
	GetRolesForUser(ctx context.Context, userID int32) ([]database.Role, error)
}

type ServicesValidator interface {
	ValidateCredentials(ctx context.Context, serviceID, serviceSecret string) error
}

type OwnershipTransferClient interface {
	TransferOwnership(ctx context.Context, sourceUserID, targetUserID int32) error
}
