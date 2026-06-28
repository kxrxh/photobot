package auth

import (
	"context"
	"crypto/rand"
	"errors"
	"fmt"
	"strconv"
	"strings"
	"time"

	"csort.ru/auth-service/internal/apperrors"
	"csort.ru/auth-service/internal/database"
	"csort.ru/auth-service/internal/logger"
	"csort.ru/auth-service/internal/observability"
	"csort.ru/auth-service/pkg/utils"
	"github.com/gofiber/fiber/v3"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgtype"
	"github.com/kxrxh/gopt"
	"github.com/redis/go-redis/v9"
	"github.com/rs/zerolog"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/trace"
	"golang.org/x/crypto/bcrypt"
)

const (
	redisRefreshTokenPrefix  = "refresh_token:"
	redisLinkCodePrefix      = "link_code:"
	linkCodeTTLSeconds       = 300 // 5 minutes
	linkCodeExpiresInSeconds = 300
)

type DBPool interface {
	Begin(ctx context.Context) (pgx.Tx, error)
}

type Service struct {
	db              database.Querier
	dbPool          DBPool
	tokenStore      TokenStore
	userService     UserProvider
	roleService     RoleProvider
	botService      BotTokenProvider
	servicesSvc     ServicesValidator
	ownershipClient OwnershipTransferClient
	config          *Config
	logger          zerolog.Logger
}

type Config struct {
	AccessExpiryMinutes   int
	RefreshExpiryMinutes  int
	AdminLogin            string
	AdminPassword         string
	Debug                 bool
	DebugBypassSignatures bool
	DevMode               bool
	ResetOTPTTLSeconds    int
}

type Params struct {
	DB              database.Querier
	DBPool          DBPool
	TokenStore      TokenStore
	UserService     UserProvider
	RoleService     RoleProvider
	BotService      BotTokenProvider
	ServicesService ServicesValidator
	OwnershipClient OwnershipTransferClient
	Config          *Config
}

func NewService(params *Params) *Service {
	return &Service{
		db:              params.DB,
		dbPool:          params.DBPool,
		tokenStore:      params.TokenStore,
		userService:     params.UserService,
		roleService:     params.RoleService,
		botService:      params.BotService,
		servicesSvc:     params.ServicesService,
		ownershipClient: params.OwnershipClient,
		config:          params.Config,
		logger:          logger.GetLogger("auth.service"),
	}
}

func (s *Service) Login(
	ctx context.Context,
	params *LoginParams,
) (kp *KeyPair, roles []string, err error) {
	if params == nil {
		return nil, nil, apperrors.New(fiber.StatusBadRequest, "missing login params")
	}

	ctx, loginSpan := observability.StartSpan(ctx, "auth.Login")
	loginSpan.SetAttributes(attribute.String("auth.grant_type", string(params.GTY)))
	defer func() { observability.EndSpan(loginSpan, err) }()

	// avoid logging secrets/initData
	s.logger.Info().
		Str("grant_type", string(params.GTY)).
		Str("service_id", gopt.FromPtr(params.ServiceID).UnwrapOr("")).
		Str("bot_name", gopt.FromPtr(params.BotName).UnwrapOr("")).
		Msg("Login attempt")

	var genParams GenerationParams

	switch params.GTY {
	case GrantTypeService:
		s.logger.Info().Msg("Service grant type login attempt")

		if !gopt.FromPtr(params.ServiceID).
			Filter(func(s string) bool { return s != "" }).
			IsSome() ||
			!gopt.FromPtr(params.ServiceSecret).
				Filter(func(s string) bool { return s != "" }).
				IsSome() {
			s.logger.Error().Msg("Missing service credentials")
			return nil, nil, apperrors.New(fiber.StatusBadRequest, "missing service credentials")
		}

		_, spVal := observability.StartSpan(ctx, "auth.login.validate_service_credentials",
			trace.WithAttributes(attribute.String("auth.service_id", *params.ServiceID)))
		if err = s.servicesSvc.ValidateCredentials(
			ctx,
			*params.ServiceID,
			*params.ServiceSecret,
		); err != nil {
			observability.EndSpan(spVal, err)
			s.logger.Error().Err(err).Msg("Service credentials validation failed")
			return nil, nil, apperrors.Wrap(
				err,
				fiber.StatusUnauthorized,
				"service authentication failed",
			)
		}
		observability.EndSpan(spVal, nil)

		genParams.ServiceID = params.ServiceID
		roles = []string{ServiceRole}
		genParams.Roles = roles
		genParams.GTY = params.GTY
		if params.Audience != nil {
			genParams.Audience = *params.Audience
		}

	case GrantTypePassword:
		s.logger.Info().Msg("Password grant type login attempt")

		if !gopt.FromPtr(params.Login).Filter(func(s string) bool { return s != "" }).IsSome() ||
			!gopt.FromPtr(params.Password).Filter(func(s string) bool { return s != "" }).IsSome() {
			s.logger.Error().Msg("Missing password credentials")
			return nil, nil, apperrors.New(fiber.StatusBadRequest, "missing password credentials")
		}

		_, spPwd := observability.StartSpan(ctx, "auth.login.admin_password")
		kp, roles, err = s.AdminLogin(ctx, *params.Login, *params.Password, params.GTY)
		observability.EndSpan(spPwd, err)
		if err != nil {
			return nil, nil, err
		}

		return kp, roles, nil

	case GrantTypeUserPassword:
		s.logger.Info().Msg("User password grant type login attempt")

		if !gopt.FromPtr(params.Login).Filter(func(s string) bool { return s != "" }).IsSome() ||
			!gopt.FromPtr(params.Password).Filter(func(s string) bool { return s != "" }).IsSome() {
			return nil, nil, apperrors.New(fiber.StatusBadRequest, "missing login credentials")
		}

		kp, roles, err = s.userPasswordLogin(ctx, *params.Login, *params.Password)
		if err != nil {
			return nil, nil, err
		}
		return kp, roles, nil

	case GrantTypeInitData:
		s.logger.Info().Msg("InitData grant type login attempt")

		if !gopt.FromPtr(params.InitData).Filter(func(s string) bool { return s != "" }).IsSome() ||
			!gopt.FromPtr(params.BotName).Filter(func(s string) bool { return s != "" }).IsSome() {
			s.logger.Error().Msg("Missing user credentials")
			return nil, nil, apperrors.New(fiber.StatusBadRequest, "missing user credentials")
		}

		platform := gopt.FromPtr(params.MessengerPlatform).UnwrapOr("telegram")
		if platform == "" {
			platform = "telegram"
		}

		switch platform {
		case "telegram":
			_, spBot := observability.StartSpan(ctx, "auth.login.get_bot_token",
				trace.WithAttributes(attribute.String("messenger.platform", "telegram")))
			botToken, err := s.botService.GetTokenByNameAndPlatform(
				ctx,
				*params.BotName,
				"telegram",
			)
			observability.EndSpan(spBot, err)
			if err != nil {
				s.logger.Error().
					Err(err).
					Str("bot_name", *params.BotName).
					Msg("Failed to get Telegram bot token")
				return nil, nil, apperrors.Wrap(
					err,
					fiber.StatusUnauthorized,
					"authentication failed",
				)
			}

			_, spTG := observability.StartSpan(ctx, "auth.login.validate_telegram_initdata")
			userData, err := ValidateTelegramData(
				*params.InitData,
				botToken,
				s.config.DebugBypassSignatures,
			)
			observability.EndSpan(spTG, err)
			if err != nil {
				s.logger.Error().Err(err).Msg("Telegram initdata validation failed")
				return nil, nil, apperrors.Wrap(
					err,
					fiber.StatusUnauthorized,
					"authentication failed",
				)
			}

			_, spUser := observability.StartSpan(ctx, "auth.login.load_user_by_telegram_id")
			userInfo, err := s.userService.GetByTelegramId(ctx, userData.ID)
			observability.EndSpan(spUser, err)
			if err != nil {
				s.logger.Error().Err(err).Msg("Failed to get user info by Telegram ID")
				if httpErr, ok := apperrors.FromError(
					err,
				); ok &&
					httpErr.Code == fiber.StatusNotFound {
					return nil, nil, err
				}
				return nil, nil, apperrors.Wrap(
					err,
					fiber.StatusInternalServerError,
					"failed to get user info",
				)
			}

			genParams.TelegramID = &userData.ID
			genParams.UserID = &userInfo.ID
			genParams.Roles = userInfo.Roles
			genParams.GTY = params.GTY
			roles = userInfo.Roles

		case "max":
			_, spBot := observability.StartSpan(ctx, "auth.login.get_bot_token",
				trace.WithAttributes(attribute.String("messenger.platform", "max")))
			botToken, err := s.botService.GetTokenByNameAndPlatform(ctx, *params.BotName, "max")
			observability.EndSpan(spBot, err)
			if err != nil {
				s.logger.Error().
					Err(err).
					Str("bot_name", *params.BotName).
					Msg("Failed to get MAX bot token")
				return nil, nil, apperrors.Wrap(
					err,
					fiber.StatusUnauthorized,
					"authentication failed",
				)
			}

			_, spMax := observability.StartSpan(ctx, "auth.login.validate_max_initdata")
			userData, err := ValidateMaxData(
				*params.InitData,
				botToken,
				s.config.DebugBypassSignatures,
			)
			observability.EndSpan(spMax, err)
			if err != nil {
				s.logger.Error().Err(err).Msg("MAX initdata validation failed")
				return nil, nil, apperrors.Wrap(
					err,
					fiber.StatusUnauthorized,
					"authentication failed",
				)
			}

			_, spUser := observability.StartSpan(ctx, "auth.login.load_user_by_max_id")
			userInfo, err := s.userService.GetByMaxId(ctx, userData.ID)
			observability.EndSpan(spUser, err)
			if err != nil {
				s.logger.Error().Err(err).Msg("Failed to get user info by MAX ID")
				if httpErr, ok := apperrors.FromError(
					err,
				); ok &&
					httpErr.Code == fiber.StatusNotFound {
					return nil, nil, err
				}
				return nil, nil, apperrors.Wrap(
					err,
					fiber.StatusInternalServerError,
					"failed to get user info",
				)
			}

			genParams.MaxID = &userData.ID
			genParams.UserID = &userInfo.ID
			genParams.Roles = userInfo.Roles
			genParams.GTY = params.GTY
			roles = userInfo.Roles

		default:
			s.logger.Error().
				Str("platform", platform).
				Msg("Unsupported messenger platform for initdata")
			return nil, nil, apperrors.New(fiber.StatusBadRequest, "unsupported messenger platform")
		}

	default:
		s.logger.Error().Str("grant_type", string(params.GTY)).Msg("Unsupported grant type")
		return nil, nil, apperrors.New(fiber.StatusBadRequest, "unsupported grant type")
	}

	accessTokenTTL := time.Duration(s.config.AccessExpiryMinutes) * time.Minute
	_, spAcc := observability.StartSpan(ctx, "auth.login.generate_access_jwt")
	accessToken, err := GenerateJWT(&genParams, AccessToken, accessTokenTTL)
	observability.EndSpan(spAcc, err)
	if err != nil {
		s.logger.Error().Err(err).Msg("Could not generate access token")
		return nil, nil, apperrors.Wrap(
			err,
			fiber.StatusInternalServerError,
			"could not generate access token",
		)
	}

	refreshTokenTTL := time.Duration(s.config.RefreshExpiryMinutes) * time.Minute
	refreshJTI := uuid.New().String()
	genParams.JTI = refreshJTI

	_, spRef := observability.StartSpan(ctx, "auth.login.generate_refresh_jwt")
	refreshToken, err := GenerateJWT(&genParams, RefreshToken, refreshTokenTTL)
	observability.EndSpan(spRef, err)
	if err != nil {
		s.logger.Error().Err(err).Msg("Could not generate refresh token")
		return nil, nil, apperrors.Wrap(
			err,
			fiber.StatusInternalServerError,
			"could not generate refresh token",
		)
	}

	_, spRedis := observability.StartSpan(ctx, "auth.login.redis_store_refresh_jti")
	err = s.tokenStore.Set(ctx, refreshTokenKey(refreshJTI), "1", refreshTokenTTL)
	observability.EndSpan(spRedis, err)
	if err != nil {
		s.logger.Error().Err(err).Msg("Could not store refresh token in redis")
		return nil, nil, apperrors.Wrap(
			err,
			fiber.StatusInternalServerError,
			"could not store refresh token",
		)
	}

	pair := KeyPair{
		AccessToken:  accessToken,
		RefreshToken: refreshToken,
	}

	s.logger.Info().
		Str("grant_type", string(params.GTY)).
		Strs("roles", genParams.Roles).
		Msg("User logged in successfully")

	return &pair, roles, nil
}

func (s *Service) Refresh(ctx context.Context, refreshToken string) (pair *KeyPair, err error) {
	ctx, refreshSpan := observability.StartSpan(ctx, "auth.Refresh")
	defer func() { observability.EndSpan(refreshSpan, err) }()

	_, spParse := observability.StartSpan(ctx, "auth.refresh.parse_jwt")
	claims, parseErr := ParseJWT(refreshToken)
	observability.EndSpan(spParse, parseErr)
	if parseErr != nil {
		ev := s.logger.Error().
			Err(parseErr).
			Str("reason", JWTRejectReason(parseErr))
		if exp, tokKID, ok := KeyIDMismatchDetails(parseErr); ok {
			ev = ev.Str("active_kid", exp).Str("token_kid", tokKID)
		}
		if uc, uerr := UnverifiedClaims(refreshToken); uerr == nil {
			ev = logRefreshClaims(ev, uc)
		}
		ev.Msg("Invalid refresh token")
		return nil, apperrors.Wrap(parseErr, fiber.StatusUnauthorized, "invalid refresh token")
	}
	ev := s.logger.Info()
	if claims.UserID != nil {
		ev = ev.Int32("user_id", *claims.UserID)
	}
	if claims.ServiceID != nil {
		ev = ev.Str("service_id", *claims.ServiceID)
	}
	ev.Msg("Refresh token attempt")

	if claims.Type != RefreshToken {
		logRefreshClaims(
			s.logger.Error().
				Str("reason", "wrong_token_type").
				Str("expected_type", string(RefreshToken)).
				Str("actual_type", string(claims.Type)),
			claims,
		).Msg("Token is not a refresh token")
		return nil, apperrors.New(fiber.StatusUnauthorized, "invalid token type")
	}

	if claims.ID == "" {
		logRefreshClaims(
			s.logger.Error().Str("reason", "missing_jti"),
			claims,
		).Msg("Refresh token missing JTI")
		return nil, apperrors.New(fiber.StatusUnauthorized, "invalid refresh token")
	}

	_, spRot := observability.StartSpan(ctx, "auth.refresh.redis_validate_jti")
	val, err := s.tokenStore.Get(ctx, refreshTokenKey(claims.ID))
	rotSpanErr := err
	if errors.Is(err, redis.Nil) {
		rotSpanErr = nil
	}
	observability.EndSpan(spRot, rotSpanErr)
	if errors.Is(err, redis.Nil) || val == "" {
		s.logger.Error().
			Str("jti", claims.ID).
			Msg("Refresh token JTI not found in redis or already used")
		return nil, apperrors.New(fiber.StatusUnauthorized, "refresh token expired or already used")
	}
	if err != nil {
		s.logger.Error().
			Err(err).
			Str("jti", claims.ID).
			Msg("Failed to validate refresh token JTI in redis")
		return nil, apperrors.Wrap(
			err,
			fiber.StatusInternalServerError,
			"failed to validate refresh token",
		)
	}

	var roles []string
	if claims.UserID != nil &&
		(claims.GTY == GrantTypeInitData || claims.GTY == GrantTypeUserPassword) {
		_, spRoles := observability.StartSpan(ctx, "auth.refresh.load_user_roles",
			trace.WithAttributes(attribute.Int64("auth.user_id", int64(*claims.UserID))))
		userRoles, rerr := s.roleService.GetRolesForUser(ctx, *claims.UserID)
		observability.EndSpan(spRoles, rerr)
		if rerr != nil {
			s.logger.Error().
				Err(rerr).
				Int32("user_id", *claims.UserID).
				Msg("Failed to get user roles for refresh")
			return nil, apperrors.Wrap(
				rerr,
				fiber.StatusUnauthorized,
				"failed to refresh: user roles could not be loaded",
			)
		}
		roles = make([]string, 0, len(userRoles))
		for _, r := range userRoles {
			roles = append(roles, r.Name)
		}
	} else {
		roles = claims.Roles
	}

	genParams := GenerationParams{
		UserID:     claims.UserID,
		ServiceID:  claims.ServiceID,
		TelegramID: claims.TelegramID,
		MaxID:      claims.MaxID,
		Roles:      roles,
		GTY:        claims.GTY,
	}
	if len(claims.Audience) > 0 {
		genParams.Audience = claims.Audience[0]
	}

	accessTokenTTL := time.Duration(s.config.AccessExpiryMinutes) * time.Minute
	_, spAcc := observability.StartSpan(ctx, "auth.refresh.generate_access_jwt")
	accessToken, err := GenerateJWT(&genParams, AccessToken, accessTokenTTL)
	observability.EndSpan(spAcc, err)
	if err != nil {
		s.logger.Error().Err(err).Msg("Could not generate access token")
		return nil, apperrors.Wrap(
			err,
			fiber.StatusInternalServerError,
			"could not generate access token",
		)
	}

	refreshTokenTTL := time.Duration(s.config.RefreshExpiryMinutes) * time.Minute
	newRefreshJTI := uuid.New().String()
	genParams.JTI = newRefreshJTI

	_, spRef := observability.StartSpan(ctx, "auth.refresh.generate_refresh_jwt")
	newRefreshToken, err := GenerateJWT(&genParams, RefreshToken, refreshTokenTTL)
	observability.EndSpan(spRef, err)
	if err != nil {
		s.logger.Error().Err(err).Msg("Could not generate refresh token")
		return nil, apperrors.Wrap(
			err,
			fiber.StatusInternalServerError,
			"could not generate refresh token",
		)
	}

	_, spRedis := observability.StartSpan(ctx, "auth.refresh.redis_store_refresh_jti")
	err = s.tokenStore.Set(
		ctx,
		refreshTokenKey(newRefreshJTI),
		"1",
		refreshTokenTTL,
	)
	if err != nil {
		observability.EndSpan(spRedis, err)
		s.logger.Error().Err(err).Msg("Could not store new refresh token in redis")
		return nil, apperrors.Wrap(
			err,
			fiber.StatusInternalServerError,
			"could not store refresh token",
		)
	}
	if delErr := s.tokenStore.Del(ctx, refreshTokenKey(claims.ID)); delErr != nil {
		s.logger.Warn().
			Err(delErr).
			Str("jti", claims.ID).
			Msg("Failed to delete old refresh token JTI")
	}
	observability.EndSpan(spRedis, nil)

	pair = &KeyPair{
		AccessToken:  accessToken,
		RefreshToken: newRefreshToken,
	}

	logEv := s.logger.Info().Strs("roles", genParams.Roles)
	if genParams.UserID != nil {
		logEv = logEv.Int32("user_id", *genParams.UserID)
	}
	if genParams.ServiceID != nil {
		logEv = logEv.Str("service_id", *genParams.ServiceID)
	}
	logEv.Msg("Tokens refreshed successfully")

	return pair, nil
}

func (s *Service) AdminLogin(
	ctx context.Context,
	login, password string,
	gty GrantType,
) (kp *KeyPair, rolesOut []string, err error) {
	_, spCred := observability.StartSpan(ctx, "auth.admin.verify_credentials")
	if login != s.config.AdminLogin {
		err = apperrors.New(fiber.StatusUnauthorized, "invalid credentials")
		observability.EndSpan(spCred, err)
		s.logger.Error().Msg("Invalid admin credentials")
		return nil, nil, err
	}
	if err = bcrypt.CompareHashAndPassword(
		[]byte(s.config.AdminPassword),
		[]byte(password),
	); err != nil {
		err = apperrors.New(fiber.StatusUnauthorized, "invalid credentials")
		observability.EndSpan(spCred, err)
		s.logger.Error().Msg("Invalid admin credentials")
		return nil, nil, err
	}
	observability.EndSpan(spCred, nil)

	rolesOut = []string{"admin"}
	genParams := GenerationParams{
		Roles: rolesOut,
		GTY:   gty,
	}

	accessTokenTTL := time.Duration(s.config.AccessExpiryMinutes) * time.Minute
	_, spAcc := observability.StartSpan(ctx, "auth.admin.generate_access_jwt")
	accessToken, err := GenerateJWT(&genParams, AccessToken, accessTokenTTL)
	observability.EndSpan(spAcc, err)
	if err != nil {
		s.logger.Error().Err(err).Msg("Could not generate admin access token")
		return nil, nil, apperrors.Wrap(
			err,
			fiber.StatusInternalServerError,
			"could not generate access token",
		)
	}

	refreshTokenTTL := time.Duration(s.config.RefreshExpiryMinutes) * time.Minute
	refreshJTI := uuid.New().String()
	genParams.JTI = refreshJTI

	_, spRef := observability.StartSpan(ctx, "auth.admin.generate_refresh_jwt")
	refreshToken, err := GenerateJWT(&genParams, RefreshToken, refreshTokenTTL)
	observability.EndSpan(spRef, err)
	if err != nil {
		s.logger.Error().Err(err).Msg("Could not generate admin refresh token")
		return nil, nil, apperrors.Wrap(
			err,
			fiber.StatusInternalServerError,
			"could not generate refresh token",
		)
	}

	_, spRedis := observability.StartSpan(ctx, "auth.admin.redis_store_refresh_jti")
	err = s.tokenStore.Set(ctx, refreshTokenKey(refreshJTI), "1", refreshTokenTTL)
	observability.EndSpan(spRedis, err)
	if err != nil {
		s.logger.Error().Err(err).Msg("Could not store admin refresh token in redis")
		return nil, nil, apperrors.Wrap(
			err,
			fiber.StatusInternalServerError,
			"could not store refresh token",
		)
	}

	s.logger.Info().Msg("Admin logged in successfully")
	return &KeyPair{
		AccessToken:  accessToken,
		RefreshToken: refreshToken,
	}, rolesOut, nil
}

// LinkCodeResult holds the result of requesting or consuming a link code.
type LinkCodeResult struct {
	Code             string `json:"code"`
	ExpiresInSeconds int    `json:"expires_in_seconds"`
}

// RequestLinkCode generates a 6-digit linking code for the authenticated user and stores it in Redis.
func (s *Service) RequestLinkCode(ctx context.Context, userID int32) (*LinkCodeResult, error) {
	code, err := generateLinkCode()
	if err != nil {
		s.logger.Error().Err(err).Msg("Failed to generate link code")
		return nil, apperrors.Wrap(
			err,
			fiber.StatusInternalServerError,
			"failed to generate link code",
		)
	}

	key := linkCodeKey(code)
	val := strconv.FormatInt(int64(userID), 10)
	ttl := time.Duration(linkCodeTTLSeconds) * time.Second
	if err := s.tokenStore.Set(ctx, key, val, ttl); err != nil {
		s.logger.Error().Err(err).Int32("user_id", userID).Msg("Failed to store link code in redis")
		return nil, apperrors.Wrap(
			err,
			fiber.StatusInternalServerError,
			"failed to store link code",
		)
	}

	s.logger.Info().Int32("user_id", userID).Msg("Link code generated")
	return &LinkCodeResult{Code: code, ExpiresInSeconds: linkCodeExpiresInSeconds}, nil
}

// LinkWithCodeResult holds the result of linking with a code (optionally with tokens).
type LinkWithCodeResult struct {
	Message      string   `json:"message"`
	AccessToken  string   `json:"access_token,omitempty"`
	RefreshToken string   `json:"refresh_token,omitempty"`
	Roles        []string `json:"roles,omitempty"`
}

// LinkWithCode merges two accounts into one. The older account (by created_at) is kept;
// the younger is deleted and its platform ID is linked to the older account.
func (s *Service) LinkWithCode(
	ctx context.Context,
	jwtUserID int32,
	code, initData, botName, platform string,
) (*LinkWithCodeResult, error) {
	platform = strings.TrimSpace(strings.ToLower(platform))
	if platform != "max" && platform != "telegram" {
		return nil, apperrors.New(fiber.StatusBadRequest, "platform must be max or telegram")
	}

	var platformUserID int64
	switch platform {
	case "max":
		userData, err := s.ValidateMaxData(ctx, initData, botName)
		if err != nil {
			return nil, err
		}
		platformUserID = userData.ID
	case "telegram":
		userData, err := s.ValidateTelegramData(ctx, initData, botName)
		if err != nil {
			return nil, err
		}
		platformUserID = userData.ID
	}

	// Verify JWT user has this platform ID — proves they own the init data
	jwtUser, err := s.db.GetUser(ctx, jwtUserID)
	if err != nil {
		s.logger.Error().Err(err).Int32("jwt_user_id", jwtUserID).Msg("JWT user not found")
		return nil, apperrors.New(fiber.StatusForbidden, "authenticated user not found")
	}
	if platform == "max" {
		if !jwtUser.MaxID.Valid || jwtUser.MaxID.Int64 != platformUserID {
			return nil, apperrors.New(
				fiber.StatusForbidden,
				"init data does not match the authenticated user",
			)
		}
	} else {
		if !jwtUser.TelegramID.Valid || jwtUser.TelegramID.Int64 != platformUserID {
			return nil, apperrors.New(
				fiber.StatusForbidden,
				"init data does not match the authenticated user",
			)
		}
	}

	codeUserID, err := s.lookupLinkCode(ctx, code)
	if err != nil {
		return nil, err
	}

	if jwtUserID == codeUserID {
		return nil, apperrors.New(fiber.StatusBadRequest, "cannot link account to itself")
	}

	codeUser, err := s.db.GetUser(ctx, codeUserID)
	if err != nil {
		return s.recoverMessengerLinkAfterMerge(
			ctx,
			jwtUserID,
			codeUserID,
			code,
			platform,
			platformUserID,
		)
	}

	var keepUser database.User
	var deleteUserID int32
	if jwtUser.CreatedAt.Before(codeUser.CreatedAt) || jwtUser.CreatedAt.Equal(codeUser.CreatedAt) {
		keepUser = jwtUser
		deleteUserID = codeUserID
	} else {
		keepUser = codeUser
		deleteUserID = jwtUserID
	}

	deletedUser := codeUser
	if deleteUserID == jwtUserID {
		deletedUser = jwtUser
	}
	platformToAdd, platformIDToAdd := messengerPlatformFromUser(deletedUser)

	// Run all DB mutations inside a transaction
	keepUserID, err := s.linkAccountsDB(
		ctx,
		deleteUserID,
		keepUser.ID,
		platformToAdd,
		platformIDToAdd,
	)
	if err != nil {
		return nil, err
	}

	if err := s.consumeLinkCode(ctx, code); err != nil {
		s.logger.Warn().Err(err).Msg("Failed to delete link code after successful merge")
	}

	// Transfer ownership only after DB commit succeeds
	if err := s.ownershipClient.TransferOwnership(ctx, deleteUserID, keepUserID); err != nil {
		s.logger.Error().
			Err(err).
			Int32("delete_user_id", deleteUserID).
			Int32("keep_user_id", keepUserID).
			Msg("Failed to transfer ownership in downstream services")
		return nil, apperrors.Wrap(
			err,
			fiber.StatusInternalServerError,
			"failed to link accounts: could not transfer ownership in downstream services",
		)
	}

	return s.issueMessengerLinkTokens(ctx, keepUserID, platform, platformUserID)
}

func (s *Service) linkAccountsDB(
	ctx context.Context,
	deleteUserID, keepUserID int32,
	platformToAdd string,
	platformIDToAdd int64,
) (int32, error) {
	q := s.db

	if s.dbPool != nil {
		tx, err := s.dbPool.Begin(ctx)
		if err != nil {
			s.logger.Error().Err(err).Msg("Failed to begin transaction for account linking")
			return 0, apperrors.Wrap(
				err,
				fiber.StatusInternalServerError,
				"failed to begin transaction",
			)
		}
		defer func() { _ = tx.Rollback(ctx) }()
		q = database.New(tx)

		deleteRoles, err := q.GetUserRoles(ctx, deleteUserID)
		if err != nil {
			s.logger.Error().
				Err(err).
				Int32("delete_user_id", deleteUserID).
				Msg("Failed to get user roles")
			return 0, apperrors.Wrap(
				err,
				fiber.StatusInternalServerError,
				"failed to merge accounts",
			)
		}
		for _, r := range deleteRoles {
			if addErr := q.AddUserRole(ctx, database.AddUserRoleParams{
				UserID: keepUserID,
				RoleID: r.ID,
			}); addErr != nil {
				s.logger.Warn().
					Err(addErr).
					Int32("user_id", keepUserID).
					Int32("role_id", r.ID).
					Msg("Failed to add role during merge (may already exist)")
			}
		}

		credStage, err := s.prepareWebCredentialStage(ctx, q, keepUserID, deleteUserID)
		if err != nil {
			return 0, err
		}
		s.reassignWebRecoveryCodes(ctx, q, credStage, deleteUserID, keepUserID)

		if err := q.DeleteUser(ctx, deleteUserID); err != nil {
			s.logger.Error().
				Err(err).
				Int32("delete_user_id", deleteUserID).
				Msg("Failed to delete user during merge")
			return 0, apperrors.Wrap(
				err,
				fiber.StatusInternalServerError,
				"failed to merge accounts",
			)
		}

		if err := s.applyWebCredentialsAfterDelete(ctx, q, credStage, keepUserID); err != nil {
			return 0, err
		}

		switch platformToAdd {
		case "max":
			err = q.LinkUserMaxId(ctx, database.LinkUserMaxIdParams{
				ID:    keepUserID,
				MaxID: pgtype.Int8{Int64: platformIDToAdd, Valid: true},
			})
		case "telegram":
			err = q.LinkUserTelegramId(ctx, database.LinkUserTelegramIdParams{
				ID:         keepUserID,
				TelegramID: pgtype.Int8{Int64: platformIDToAdd, Valid: true},
			})
		}
		if err != nil {
			s.logger.Error().
				Err(err).
				Int32("user_id", keepUserID).
				Str("platform", platformToAdd).
				Msg("Failed to link account")
			msg := utils.UniqueViolationMessage(err, map[string]string{
				"users_telegram_id_key": "Аккаунт с этим Telegram уже привязан",
				"users_max_id_key":      "Аккаунт с этим MAX уже привязан",
			}, "Этот аккаунт уже привязан к другому пользователю")
			if msg != "" {
				return 0, apperrors.New(fiber.StatusConflict, msg)
			}
			return 0, apperrors.Wrap(err, fiber.StatusInternalServerError, "failed to link account")
		}

		if err := tx.Commit(ctx); err != nil {
			s.logger.Error().Err(err).Msg("Failed to commit account linking transaction")
			return 0, apperrors.Wrap(
				err,
				fiber.StatusInternalServerError,
				"failed to commit transaction",
			)
		}
	} else {
		deleteRoles, err := q.GetUserRoles(ctx, deleteUserID)
		if err != nil {
			s.logger.Error().
				Err(err).
				Int32("delete_user_id", deleteUserID).
				Msg("Failed to get user roles")
			return 0, apperrors.Wrap(
				err,
				fiber.StatusInternalServerError,
				"failed to merge accounts",
			)
		}
		for _, r := range deleteRoles {
			if addErr := q.AddUserRole(ctx, database.AddUserRoleParams{
				UserID: keepUserID,
				RoleID: r.ID,
			}); addErr != nil {
				s.logger.Warn().
					Err(addErr).
					Int32("user_id", keepUserID).
					Int32("role_id", r.ID).
					Msg("Failed to add role during merge (may already exist)")
			}
		}

		credStage, err := s.prepareWebCredentialStage(ctx, q, keepUserID, deleteUserID)
		if err != nil {
			return 0, err
		}
		s.reassignWebRecoveryCodes(ctx, q, credStage, deleteUserID, keepUserID)

		if err := q.DeleteUser(ctx, deleteUserID); err != nil {
			s.logger.Error().
				Err(err).
				Int32("delete_user_id", deleteUserID).
				Msg("Failed to delete user during merge")
			return 0, apperrors.Wrap(
				err,
				fiber.StatusInternalServerError,
				"failed to merge accounts",
			)
		}

		if err := s.applyWebCredentialsAfterDelete(ctx, q, credStage, keepUserID); err != nil {
			return 0, err
		}

		switch platformToAdd {
		case "max":
			err = q.LinkUserMaxId(ctx, database.LinkUserMaxIdParams{
				ID:    keepUserID,
				MaxID: pgtype.Int8{Int64: platformIDToAdd, Valid: true},
			})
		case "telegram":
			err = q.LinkUserTelegramId(ctx, database.LinkUserTelegramIdParams{
				ID:         keepUserID,
				TelegramID: pgtype.Int8{Int64: platformIDToAdd, Valid: true},
			})
		}
		if err != nil {
			s.logger.Error().
				Err(err).
				Int32("user_id", keepUserID).
				Str("platform", platformToAdd).
				Msg("Failed to link account")
			msg := utils.UniqueViolationMessage(err, map[string]string{
				"users_telegram_id_key": "Аккаунт с этим Telegram уже привязан",
				"users_max_id_key":      "Аккаунт с этим MAX уже привязан",
			}, "Этот аккаунт уже привязан к другому пользователю")
			if msg != "" {
				return 0, apperrors.New(fiber.StatusConflict, msg)
			}
			return 0, apperrors.Wrap(err, fiber.StatusInternalServerError, "failed to link account")
		}
	}

	s.logger.Info().
		Int32("user_id", keepUserID).
		Str("platform", platformToAdd).
		Msg("Accounts merged via link code (kept older)")

	return keepUserID, nil
}

func normalizeLinkCode(code string) (string, error) {
	var digits strings.Builder
	for _, r := range strings.TrimSpace(code) {
		if r >= '0' && r <= '9' {
			digits.WriteRune(r)
		}
	}
	normalized := digits.String()
	if len(normalized) != 6 {
		return "", apperrors.New(fiber.StatusBadRequest, "link code must be 6 digits")
	}
	return normalized, nil
}

func (s *Service) lookupLinkCode(ctx context.Context, code string) (int32, error) {
	normalized, err := normalizeLinkCode(code)
	if err != nil {
		return 0, err
	}

	key := linkCodeKey(normalized)
	codeUserIDStr, err := s.tokenStore.Get(ctx, key)
	if errors.Is(err, redis.Nil) || codeUserIDStr == "" {
		s.logger.Warn().Str("code", normalized).Msg("Link code not found or expired")
		return 0, apperrors.New(fiber.StatusUnauthorized, "link code not found or expired")
	}
	if err != nil {
		s.logger.Error().Err(err).Str("code", normalized).Msg("Failed to get link code from redis")
		return 0, apperrors.Wrap(
			err,
			fiber.StatusInternalServerError,
			"failed to validate link code",
		)
	}

	codeUserID, err := strconv.ParseInt(codeUserIDStr, 10, 32)
	if err != nil {
		s.logger.Error().
			Err(err).
			Str("user_id_str", codeUserIDStr).
			Msg("Invalid user_id in link code")
		return 0, apperrors.Wrap(err, fiber.StatusInternalServerError, "invalid link code")
	}
	return int32(codeUserID), nil
}

func (s *Service) consumeLinkCode(ctx context.Context, code string) error {
	normalized, err := normalizeLinkCode(code)
	if err != nil {
		return err
	}
	return s.tokenStore.Del(ctx, linkCodeKey(normalized))
}

func messengerPlatformFromUser(u database.User) (platform string, platformID int64) {
	if u.MaxID.Valid {
		return "max", u.MaxID.Int64
	}
	if u.TelegramID.Valid {
		return "telegram", u.TelegramID.Int64
	}
	return "", 0
}

func generateLinkCode() (string, error) {
	var b [4]byte
	if _, err := rand.Read(b[:]); err != nil {
		return "", err
	}
	n := uint32(b[0])<<24 | uint32(b[1])<<16 | uint32(b[2])<<8 | uint32(b[3])
	code := 100000 + int(n%900000) // 100000 to 999999
	return strconv.Itoa(code), nil
}

func linkCodeKey(code string) string {
	return redisLinkCodePrefix + code
}

func (s *Service) ValidateTelegramData(
	ctx context.Context,
	initData, botName string,
) (*UserData, error) {
	botToken, err := s.botService.GetTokenByNameAndPlatform(ctx, botName, "telegram")
	if err != nil {
		s.logger.Error().Err(err).Str("bot_name", botName).Msg("Could not get bot token")
		if httpErr, ok := apperrors.FromError(err); ok && httpErr.Code == fiber.StatusNotFound {
			return nil, apperrors.Wrap(
				err,
				fiber.StatusBadRequest,
				fmt.Sprintf("invalid bot name: %q", botName),
			)
		}
		return nil, apperrors.Wrap(err, fiber.StatusInternalServerError, "failed to get bot token")
	}

	userData, err := ValidateTelegramData(initData, botToken, s.config.DebugBypassSignatures)
	if err != nil {
		s.logger.Error().Err(err).Msg("Could not validate telegram data")
		return nil, apperrors.Wrap(err, fiber.StatusUnauthorized, "telegram data validation failed")
	}

	return userData, nil
}

// ValidateMaxData validates MAX init data (login TTL).
func (s *Service) ValidateMaxData(
	ctx context.Context,
	initData, botName string,
) (*UserData, error) {
	botToken, err := s.botService.GetTokenByNameAndPlatform(ctx, botName, "max")
	if err != nil {
		s.logger.Error().Err(err).Str("bot_name", botName).Msg("Could not get MAX bot token")
		if httpErr, ok := apperrors.FromError(err); ok && httpErr.Code == fiber.StatusNotFound {
			return nil, apperrors.Wrap(
				err,
				fiber.StatusBadRequest,
				fmt.Sprintf("invalid bot name: %q", botName),
			)
		}
		return nil, apperrors.Wrap(err, fiber.StatusInternalServerError, "failed to get bot token")
	}

	userData, err := ValidateMaxData(initData, botToken, s.config.DebugBypassSignatures)
	if err != nil {
		s.logger.Error().Err(err).Msg("Could not validate MAX init data")
		return nil, apperrors.Wrap(err, fiber.StatusUnauthorized, "max data validation failed")
	}

	return userData, nil
}

func logRefreshClaims(ev *zerolog.Event, c *Claims) *zerolog.Event {
	if c == nil {
		return ev
	}
	if c.UserID != nil {
		ev = ev.Int32("user_id", *c.UserID)
	}
	if c.ServiceID != nil {
		ev = ev.Str("service_id", *c.ServiceID)
	}
	if c.TelegramID != nil {
		ev = ev.Int64("telegram_id", *c.TelegramID)
	}
	if c.MaxID != nil {
		ev = ev.Int64("max_id", *c.MaxID)
	}
	if c.Subject != "" {
		ev = ev.Str("subject", c.Subject)
	}
	if c.ID != "" {
		ev = ev.Str("jti", c.ID)
	}
	if c.Type != "" {
		ev = ev.Str("jwt_type", string(c.Type))
	}
	if c.GTY != "" {
		ev = ev.Str("grant_type", string(c.GTY))
	}
	return ev
}

func refreshTokenKey(jti string) string {
	return redisRefreshTokenPrefix + jti
}
