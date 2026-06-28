package auth

import (
	"context"
	"crypto/rand"
	"errors"
	"fmt"
	"math/big"
	"strings"
	"time"

	"csort.ru/auth-service/internal/apperrors"
	"csort.ru/auth-service/internal/database"
	"csort.ru/auth-service/internal/messenger"
	"csort.ru/auth-service/internal/user"
	"csort.ru/auth-service/internal/webauth"
	"csort.ru/auth-service/pkg/utils"
	"github.com/gofiber/fiber/v3"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgtype"
	"github.com/redis/go-redis/v9"
)

const redisResetOTPPrefix = "reset_otp:"

type WebRegisterResult struct {
	AccessToken   string   `json:"access_token"`
	RefreshToken  string   `json:"refresh_token"`
	Roles         []string `json:"roles"`
	RecoveryCodes []string `json:"recovery_codes"`
}

type webUserCreator interface {
	CreateWeb(
		ctx context.Context,
		req *user.WebRegisterRequest,
		passwordHash string,
	) (*user.User, error)
}

func (s *Service) RegisterWeb(
	ctx context.Context,
	creator webUserCreator,
	req *user.WebRegisterRequest,
) (*WebRegisterResult, error) {
	login := webauth.NormalizeLogin(req.Login)
	if !webauth.ValidateLoginFormat(login) {
		return nil, apperrors.New(fiber.StatusBadRequest, "invalid login format")
	}
	if err := webauth.ValidatePassword(req.Password); err != nil {
		return nil, apperrors.New(fiber.StatusBadRequest, err.Error())
	}

	hash, err := webauth.HashPassword(req.Password)
	if err != nil {
		return nil, apperrors.Wrap(err, fiber.StatusInternalServerError, "failed to hash password")
	}

	req.Login = login
	u, err := creator.CreateWeb(ctx, req, hash)
	if err != nil {
		return nil, err
	}

	plainCodes, codeHashes, err := webauth.GenerateRecoveryCodes()
	if err != nil {
		return nil, apperrors.Wrap(
			err,
			fiber.StatusInternalServerError,
			"failed to generate recovery codes",
		)
	}
	for _, h := range codeHashes {
		if _, err := s.db.CreateRecoveryCode(ctx, database.CreateRecoveryCodeParams{
			UserID:   u.ID,
			CodeHash: h,
		}); err != nil {
			return nil, apperrors.Wrap(
				err,
				fiber.StatusInternalServerError,
				"failed to store recovery codes",
			)
		}
	}

	kp, roles, err := s.issueUserTokenPair(ctx, u, GrantTypeUserPassword)
	if err != nil {
		return nil, err
	}

	return &WebRegisterResult{
		AccessToken:   kp.AccessToken,
		RefreshToken:  kp.RefreshToken,
		Roles:         roles,
		RecoveryCodes: plainCodes,
	}, nil
}

type SetupWebAccessResult struct {
	RecoveryCodes []string `json:"recovery_codes"`
}

func (s *Service) SetupWebAccess(
	ctx context.Context,
	userID int32,
	login, password string,
) (*SetupWebAccessResult, error) {
	login = webauth.NormalizeLogin(login)
	if !webauth.ValidateLoginFormat(login) {
		return nil, apperrors.New(fiber.StatusBadRequest, "invalid login format")
	}
	if err := webauth.ValidatePassword(password); err != nil {
		return nil, apperrors.New(fiber.StatusBadRequest, err.Error())
	}

	existing, err := s.db.GetUser(ctx, userID)
	if err != nil {
		return nil, apperrors.New(fiber.StatusNotFound, "user not found")
	}
	if existing.Login.Valid && strings.TrimSpace(existing.Login.String) != "" {
		return nil, apperrors.New(fiber.StatusConflict, "web login already set")
	}

	hash, err := webauth.HashPassword(password)
	if err != nil {
		return nil, apperrors.Wrap(err, fiber.StatusInternalServerError, "failed to hash password")
	}

	// Attempt to set credentials; relies on UNIQUE constraint for atomic check
	_, err = s.db.SetWebCredentials(ctx, database.SetWebCredentialsParams{
		ID:           userID,
		Login:        pgtype.Text{String: login, Valid: true},
		PasswordHash: pgtype.Text{String: hash, Valid: true},
	})
	if err != nil {
		if msg := utils.UniqueViolationMessage(err, map[string]string{
			"users_login_key": "login already taken",
		}, "failed to set web credentials"); msg != "" {
			return nil, apperrors.New(fiber.StatusConflict, msg)
		}
		return nil, apperrors.Wrap(
			err,
			fiber.StatusInternalServerError,
			"failed to set web credentials",
		)
	}

	plainCodes, codeHashes, err := webauth.GenerateRecoveryCodes()
	if err != nil {
		return nil, apperrors.Wrap(
			err,
			fiber.StatusInternalServerError,
			"failed to generate recovery codes",
		)
	}
	for _, h := range codeHashes {
		if _, err := s.db.CreateRecoveryCode(ctx, database.CreateRecoveryCodeParams{
			UserID:   userID,
			CodeHash: h,
		}); err != nil {
			return nil, apperrors.Wrap(
				err,
				fiber.StatusInternalServerError,
				"failed to store recovery codes",
			)
		}
	}

	return &SetupWebAccessResult{RecoveryCodes: plainCodes}, nil
}

func (s *Service) userPasswordLogin(
	ctx context.Context,
	login, password string,
) (*KeyPair, []string, error) {
	login = webauth.NormalizeLogin(login)
	u, err := s.userService.GetByLogin(ctx, login)
	if err != nil {
		return nil, nil, apperrors.New(fiber.StatusUnauthorized, "invalid credentials")
	}
	if !u.PasswordHash.Valid || u.PasswordHash.String == "" {
		return nil, nil, apperrors.New(fiber.StatusUnauthorized, "invalid credentials")
	}
	if err := webauth.CheckPassword(u.PasswordHash.String, password); err != nil {
		return nil, nil, apperrors.New(fiber.StatusUnauthorized, "invalid credentials")
	}

	kp, roles, err := s.issueUserTokenPair(ctx, u, GrantTypeUserPassword)
	if err != nil {
		return nil, nil, err
	}
	return kp, roles, nil
}

func (s *Service) issueUserTokenPair(
	ctx context.Context,
	u *user.User,
	gty GrantType,
) (*KeyPair, []string, error) {
	genParams := GenerationParams{
		UserID: &u.ID,
		Roles:  u.Roles,
		GTY:    gty,
	}
	if u.TelegramID.Valid {
		tid := u.TelegramID.Int64
		genParams.TelegramID = &tid
	}
	if u.MaxID.Valid {
		mid := u.MaxID.Int64
		genParams.MaxID = &mid
	}

	accessTokenTTL := time.Duration(s.config.AccessExpiryMinutes) * time.Minute
	accessToken, err := GenerateJWT(&genParams, AccessToken, accessTokenTTL)
	if err != nil {
		return nil, nil, apperrors.Wrap(
			err,
			fiber.StatusInternalServerError,
			"could not generate access token",
		)
	}

	refreshTokenTTL := time.Duration(s.config.RefreshExpiryMinutes) * time.Minute
	refreshJTI := uuid.New().String()
	genParams.JTI = refreshJTI
	refreshToken, err := GenerateJWT(&genParams, RefreshToken, refreshTokenTTL)
	if err != nil {
		return nil, nil, apperrors.Wrap(
			err,
			fiber.StatusInternalServerError,
			"could not generate refresh token",
		)
	}
	if err := s.tokenStore.Set(ctx, refreshTokenKey(refreshJTI), "1", refreshTokenTTL); err != nil {
		return nil, nil, apperrors.Wrap(
			err,
			fiber.StatusInternalServerError,
			"could not store refresh token",
		)
	}

	return &KeyPair{AccessToken: accessToken, RefreshToken: refreshToken}, u.Roles, nil
}

func (s *Service) ChangePassword(
	ctx context.Context,
	userID int32,
	currentPassword, newPassword string,
) error {
	if err := webauth.ValidatePassword(newPassword); err != nil {
		return apperrors.New(fiber.StatusBadRequest, err.Error())
	}

	u, err := s.userService.Get(ctx, userID)
	if err != nil {
		return err
	}
	if !u.PasswordHash.Valid {
		return apperrors.New(fiber.StatusBadRequest, "account has no web password set up")
	}
	if err := webauth.CheckPassword(u.PasswordHash.String, currentPassword); err != nil {
		return apperrors.New(fiber.StatusUnauthorized, "invalid current password")
	}

	hash, err := webauth.HashPassword(newPassword)
	if err != nil {
		return apperrors.Wrap(err, fiber.StatusInternalServerError, "failed to hash password")
	}
	return s.db.SetPasswordHash(ctx, database.SetPasswordHashParams{
		ID:           userID,
		PasswordHash: pgtype.Text{String: hash, Valid: true},
	})
}

func (s *Service) ForgotPassword(ctx context.Context, login string) error {
	login = webauth.NormalizeLogin(login)
	u, err := s.userService.GetByLogin(ctx, login)
	if err != nil {
		// generic success to avoid login enumeration
		return nil
	}

	if !u.TelegramID.Valid && !u.MaxID.Valid {
		return nil
	}

	otp, err := generateNumericOTP(6)
	if err != nil {
		s.logger.Error().Err(err).Msg("failed to generate reset OTP")
		return nil
	}

	ttl := time.Duration(s.config.ResetOTPTTLSeconds) * time.Second
	if ttl <= 0 {
		ttl = 10 * time.Minute
	}
	key := resetOTPKey(u.ID)
	if err := s.tokenStore.Set(ctx, key, otp, ttl); err != nil {
		s.logger.Error().Err(err).Int32("user_id", u.ID).Msg("failed to store reset OTP")
		return nil
	}

	notifier := messenger.NewNotifier()
	msg := fmt.Sprintf("Код для сброса пароля: %s\nДействителен %d мин.", otp, int(ttl.Minutes()))

	if u.TelegramID.Valid {
		s.sendResetOTPViaPlatform(ctx, notifier, "telegram", u.TelegramID.Int64, msg)
	}
	if u.MaxID.Valid {
		s.sendResetOTPViaPlatform(ctx, notifier, "max", u.MaxID.Int64, msg)
	}
	return nil
}

func (s *Service) sendResetOTPViaPlatform(
	ctx context.Context,
	notifier *messenger.Notifier,
	platform string,
	messengerUserID int64,
	message string,
) {
	tokens, err := s.botService.ListTokensByPlatform(ctx, platform)
	if err != nil {
		s.logger.Warn().Err(err).Str("platform", platform).Msg("no bots available for reset OTP")
		return
	}
	for _, token := range tokens {
		if sendErr := notifier.SendText(
			ctx,
			platform,
			token,
			messengerUserID,
			message,
		); sendErr == nil {
			return
		}
	}
	s.logger.Warn().
		Str("platform", platform).
		Int64("messenger_user_id", messengerUserID).
		Msg("failed to send reset OTP via all platform bots")
}

func (s *Service) ResetPassword(ctx context.Context, login, otp, newPassword string) error {
	if err := webauth.ValidatePassword(newPassword); err != nil {
		return apperrors.New(fiber.StatusBadRequest, err.Error())
	}
	login = webauth.NormalizeLogin(login)
	u, err := s.userService.GetByLogin(ctx, login)
	if err != nil {
		return apperrors.New(fiber.StatusUnauthorized, "invalid reset request")
	}

	key := resetOTPKey(u.ID)
	stored, err := s.tokenStore.GetAndDel(ctx, key)
	if errors.Is(err, redis.Nil) || stored == "" {
		return apperrors.New(fiber.StatusUnauthorized, "reset code not found or expired")
	}
	if err != nil {
		return apperrors.Wrap(err, fiber.StatusInternalServerError, "failed to validate reset code")
	}
	if strings.TrimSpace(stored) != strings.TrimSpace(otp) {
		return apperrors.New(fiber.StatusUnauthorized, "invalid reset code")
	}

	hash, err := webauth.HashPassword(newPassword)
	if err != nil {
		return apperrors.Wrap(err, fiber.StatusInternalServerError, "failed to hash password")
	}
	return s.db.SetPasswordHash(ctx, database.SetPasswordHashParams{
		ID:           u.ID,
		PasswordHash: pgtype.Text{String: hash, Valid: true},
	})
}

func (s *Service) ResetPasswordRecovery(
	ctx context.Context,
	login, recoveryCode, newPassword string,
) error {
	if err := webauth.ValidatePassword(newPassword); err != nil {
		return apperrors.New(fiber.StatusBadRequest, err.Error())
	}
	login = webauth.NormalizeLogin(login)
	u, err := s.userService.GetByLogin(ctx, login)
	if err != nil {
		return apperrors.New(fiber.StatusUnauthorized, "invalid reset request")
	}

	code := webauth.NormalizeRecoveryCode(recoveryCode)
	codes, err := s.db.ListUnusedRecoveryCodesByUser(ctx, u.ID)
	if err != nil {
		return apperrors.Wrap(err, fiber.StatusInternalServerError, "failed to load recovery codes")
	}

	var matchedID int32
	for _, rc := range codes {
		if webauth.CheckPassword(rc.CodeHash, code) == nil {
			matchedID = rc.ID
			break
		}
	}
	if matchedID == 0 {
		return apperrors.New(fiber.StatusUnauthorized, "invalid recovery code")
	}

	hash, err := webauth.HashPassword(newPassword)
	if err != nil {
		return apperrors.Wrap(err, fiber.StatusInternalServerError, "failed to hash password")
	}
	if err := s.db.SetPasswordHash(ctx, database.SetPasswordHashParams{
		ID:           u.ID,
		PasswordHash: pgtype.Text{String: hash, Valid: true},
	}); err != nil {
		return apperrors.Wrap(err, fiber.StatusInternalServerError, "failed to update password")
	}
	return s.db.MarkRecoveryCodeUsed(ctx, matchedID)
}

func resetOTPKey(userID int32) string {
	return fmt.Sprintf("%s%d", redisResetOTPPrefix, userID)
}

func generateNumericOTP(length int) (string, error) {
	var b strings.Builder
	for range length {
		n, err := rand.Int(rand.Reader, big.NewInt(10))
		if err != nil {
			return "", err
		}
		digit := n.Int64()
		if digit < 0 || digit > 9 {
			return "", fmt.Errorf("invalid otp digit: %d", digit)
		}
		b.WriteByte('0' + byte(digit))
	}
	return b.String(), nil
}
