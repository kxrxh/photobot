package server

import (
	"fmt"

	"csort.ru/auth-service/internal/api/ownership"
	"csort.ru/auth-service/internal/auth"
	"csort.ru/auth-service/internal/bot"
	"csort.ru/auth-service/internal/config"
	"csort.ru/auth-service/internal/database"
	"csort.ru/auth-service/internal/role"
	"csort.ru/auth-service/internal/server/routes"
	"csort.ru/auth-service/internal/service"
	"csort.ru/auth-service/internal/transport/http"
	"csort.ru/auth-service/internal/user"
	validatepkg "csort.ru/auth-service/pkg/validator"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/redis/go-redis/v9"
)

type Container struct {
	RoleService    *role.Service
	BotService     *bot.Service
	UserService    *user.Service
	AuthService    *auth.Service
	ServiceService *service.Service
}

func initializeServices(
	dbQueries *database.Queries,
	dbPool *pgxpool.Pool,
	cfg *config.Config,
	redisClient *redis.Client,
) *Container {
	roleService := role.NewService(dbQueries)
	botService, err := bot.NewService(dbQueries, cfg.Security.EncryptionKey)
	if err != nil {
		panic(fmt.Errorf("failed to initialize bot service: %w", err))
	}
	userService := user.NewService(dbQueries, roleService)
	serviceService := service.NewService(dbQueries)
	ownPtr := ownership.NewClient(
		&cfg.Merge,
		auth.IssueMergeToken,
	)
	var ownershipTransfer auth.OwnershipTransferClient
	if ownPtr != nil {
		ownershipTransfer = ownPtr
	} else {
		ownershipTransfer = ownership.NoopTransfer{}
	}
	authService := auth.NewService(&auth.Params{
		DB:              dbQueries,
		DBPool:          dbPool,
		TokenStore:      auth.NewRedisTokenStore(redisClient),
		UserService:     userService,
		RoleService:     roleService,
		BotService:      botService,
		ServicesService: serviceService,
		OwnershipClient: ownershipTransfer,
		Config: &auth.Config{
			AccessExpiryMinutes:   cfg.Security.AccessExpiryMinutes,
			RefreshExpiryMinutes:  cfg.Security.RefreshExpiryMinutes,
			AdminLogin:            cfg.Security.AdminLogin,
			AdminPassword:         cfg.Security.AdminPassword,
			Debug:                 cfg.Debug,
			DebugBypassSignatures: cfg.Security.DebugBypassSignatures || cfg.DevMode,
			DevMode:               cfg.DevMode,
			ResetOTPTTLSeconds:    cfg.ResetOTPTTLSeconds,
		},
	})

	return &Container{
		RoleService:    roleService,
		BotService:     botService,
		UserService:    userService,
		AuthService:    authService,
		ServiceService: serviceService,
	}
}

func initializeHandlers(c *Container) routes.Handlers {
	v := validatepkg.Default
	return routes.Handlers{
		UserHandler:    http.NewUserHandler(c.UserService, c.AuthService, v),
		RoleHandler:    http.NewRoleHandler(c.RoleService, v),
		BotHandler:     http.NewBotHandler(c.BotService, v),
		AuthHandler:    http.NewAuthHandler(c.AuthService, c.UserService, v),
		ServiceHandler: http.NewServiceHandler(c.ServiceService),
	}
}
