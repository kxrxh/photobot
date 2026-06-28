package auth

import (
	"context"
	"strings"
	"time"

	"csort.ru/auth-service/internal/apperrors"
	"csort.ru/auth-service/internal/database"
	"csort.ru/auth-service/pkg/utils"
	"github.com/gofiber/fiber/v3"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgtype"
)

func (s *Service) LinkWithCodeFromWeb(
	ctx context.Context,
	webUserID int32,
	code string,
) (*LinkWithCodeResult, error) {
	codeUserID, err := s.lookupLinkCode(ctx, code)
	if err != nil {
		return nil, err
	}
	if webUserID == codeUserID {
		return nil, apperrors.New(fiber.StatusBadRequest, "cannot link account to itself")
	}

	webUser, err := s.db.GetUser(ctx, webUserID)
	if err != nil {
		return nil, apperrors.New(fiber.StatusForbidden, "authenticated user not found")
	}
	codeUser, err := s.db.GetUser(ctx, codeUserID)
	if err != nil {
		return s.recoverWebLinkAfterMerge(ctx, webUserID, codeUserID, code)
	}

	return s.mergeAccounts(ctx, webUser, codeUser, webUserID, codeUserID, code)
}

func hasWebCredentials(u database.User) bool {
	if u.PasswordHash.Valid && strings.TrimSpace(u.PasswordHash.String) != "" {
		return true
	}
	return u.Login.Valid && strings.TrimSpace(u.Login.String) != ""
}

type webCredentialStage struct {
	reassignRecovery bool
	applyCredentials bool
	login            pgtype.Text
	passwordHash     pgtype.Text
}

func (s *Service) prepareWebCredentialStage(
	ctx context.Context,
	q database.Querier,
	keepUserID, deleteUserID int32,
) (webCredentialStage, error) {
	keepUser, err := q.GetUser(ctx, keepUserID)
	if err != nil {
		return webCredentialStage{}, apperrors.Wrap(
			err,
			fiber.StatusInternalServerError,
			"failed to load keep user for credential merge",
		)
	}

	deleteUser, err := q.GetUser(ctx, deleteUserID)
	if err != nil {
		return webCredentialStage{}, apperrors.Wrap(
			err,
			fiber.StatusInternalServerError,
			"failed to load user for credential merge",
		)
	}

	if !hasWebCredentials(deleteUser) {
		return webCredentialStage{}, nil
	}

	stage := webCredentialStage{reassignRecovery: true}
	if hasWebCredentials(keepUser) {
		s.logger.Info().
			Int32("keep_user_id", keepUserID).
			Int32("delete_user_id", deleteUserID).
			Msg("keep user already has web credentials; skipping login transfer")
		return stage, nil
	}

	stage.applyCredentials = true
	stage.login = deleteUser.Login
	stage.passwordHash = deleteUser.PasswordHash
	return stage, nil
}

func (s *Service) reassignWebRecoveryCodes(
	ctx context.Context,
	q database.Querier,
	stage webCredentialStage,
	deleteUserID, keepUserID int32,
) {
	if !stage.reassignRecovery {
		return
	}
	if err := q.ReassignRecoveryCodes(ctx, database.ReassignRecoveryCodesParams{
		UserID:   deleteUserID,
		UserID_2: keepUserID,
	}); err != nil {
		s.logger.Warn().Err(err).Msg("Failed to reassign recovery codes during merge")
	}
}

func (s *Service) applyWebCredentialsAfterDelete(
	ctx context.Context,
	q database.Querier,
	stage webCredentialStage,
	keepUserID int32,
) error {
	if !stage.applyCredentials {
		return nil
	}

	if err := q.CopyWebCredentials(ctx, database.CopyWebCredentialsParams{
		ID:           keepUserID,
		Login:        stage.login,
		PasswordHash: stage.passwordHash,
	}); err != nil {
		s.logger.Error().Err(err).Msg("Failed to apply web credentials after account merge")
		if msg := utils.UniqueViolationMessage(err, map[string]string{
			"users_login_key": "Логин уже занят другим аккаунтом",
		}, "failed to merge web credentials"); msg != "" {
			return apperrors.New(fiber.StatusConflict, msg)
		}
		return apperrors.Wrap(
			err,
			fiber.StatusInternalServerError,
			"failed to merge web credentials",
		)
	}
	return nil
}

func (s *Service) mergeAccounts(
	ctx context.Context,
	jwtUser, codeUser database.User,
	jwtUserID, codeUserID int32,
	code string,
) (*LinkWithCodeResult, error) {
	var keepUser database.User
	var deleteUserID int32
	if jwtUser.CreatedAt.Before(codeUser.CreatedAt) || jwtUser.CreatedAt.Equal(codeUser.CreatedAt) {
		keepUser = jwtUser
		deleteUserID = codeUserID
	} else {
		keepUser = codeUser
		deleteUserID = jwtUserID
	}

	deletedUser, err := s.db.GetUser(ctx, deleteUserID)
	if err != nil {
		return nil, apperrors.Wrap(
			err,
			fiber.StatusInternalServerError,
			"failed to load deleted user",
		)
	}
	platformToAdd, platformIDToAdd := messengerPlatformFromUser(deletedUser)

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
		s.logger.Warn().Err(err).Msg("Failed to delete link code after successful web merge")
	}

	if err := s.ownershipClient.TransferOwnership(ctx, deleteUserID, keepUserID); err != nil {
		s.logger.Error().Err(err).Msg("Failed to transfer ownership after web link")
		return nil, apperrors.Wrap(
			err,
			fiber.StatusInternalServerError,
			"failed to link accounts: could not transfer ownership in downstream services",
		)
	}

	return s.issueWebLinkTokens(ctx, keepUserID)
}

func (s *Service) issueWebLinkTokens(
	ctx context.Context,
	keepUserID int32,
) (*LinkWithCodeResult, error) {
	userInfo, err := s.userService.Get(ctx, keepUserID)
	if err != nil {
		return &LinkWithCodeResult{Message: "Account linked successfully"}, nil
	}

	kp, _, err := s.issueUserTokenPair(ctx, userInfo, GrantTypeUserPassword)
	if err != nil {
		return &LinkWithCodeResult{Message: "Account linked successfully"}, nil
	}
	return &LinkWithCodeResult{
		Message:      "Account linked successfully",
		AccessToken:  kp.AccessToken,
		RefreshToken: kp.RefreshToken,
		Roles:        userInfo.Roles,
	}, nil
}

func (s *Service) issueMessengerLinkTokens(
	ctx context.Context,
	keepUserID int32,
	platform string,
	platformUserID int64,
) (*LinkWithCodeResult, error) {
	userInfo, err := s.userService.Get(ctx, keepUserID)
	if err != nil {
		s.logger.Warn().
			Err(err).
			Int32("user_id", keepUserID).
			Msg("Failed to get user for token issuance after link")
		return &LinkWithCodeResult{Message: "Account linked successfully"}, nil
	}

	genParams := GenerationParams{
		UserID: &userInfo.ID,
		Roles:  userInfo.Roles,
		GTY:    GrantTypeInitData,
	}
	if platform == "telegram" {
		genParams.TelegramID = &platformUserID
	} else {
		genParams.MaxID = &platformUserID
	}

	accessTokenTTL := time.Duration(s.config.AccessExpiryMinutes) * time.Minute
	accessToken, err := GenerateJWT(&genParams, AccessToken, accessTokenTTL)
	if err != nil {
		s.logger.Warn().Err(err).Msg("Failed to generate access token after link")
		return &LinkWithCodeResult{Message: "Account linked successfully"}, nil
	}
	refreshTokenTTL := time.Duration(s.config.RefreshExpiryMinutes) * time.Minute
	refreshJTI := uuid.New().String()
	genParams.JTI = refreshJTI
	refreshToken, err := GenerateJWT(&genParams, RefreshToken, refreshTokenTTL)
	if err != nil {
		s.logger.Warn().Err(err).Msg("Failed to generate refresh token after link")
		return &LinkWithCodeResult{
			Message:     "Account linked successfully",
			AccessToken: accessToken,
		}, nil
	}
	if err := s.tokenStore.Set(ctx, refreshTokenKey(refreshJTI), "1", refreshTokenTTL); err != nil {
		s.logger.Warn().Err(err).Msg("Failed to store refresh token after link")
		return &LinkWithCodeResult{
			Message:      "Account linked successfully",
			AccessToken:  accessToken,
			RefreshToken: refreshToken,
			Roles:        userInfo.Roles,
		}, nil
	}

	return &LinkWithCodeResult{
		Message:      "Account linked successfully",
		AccessToken:  accessToken,
		RefreshToken: refreshToken,
		Roles:        userInfo.Roles,
	}, nil
}

func (s *Service) recoverMessengerLinkAfterMerge(
	ctx context.Context,
	keepUserID, deletedCodeUserID int32,
	code, platform string,
	platformUserID int64,
) (*LinkWithCodeResult, error) {
	if _, err := s.db.GetUser(ctx, keepUserID); err != nil {
		s.logger.Error().
			Err(err).
			Int32("code_user_id", deletedCodeUserID).
			Msg("Code user not found")
		return nil, apperrors.New(fiber.StatusBadRequest, "code user not found")
	}

	s.logger.Info().
		Int32("keep_user_id", keepUserID).
		Int32("deleted_code_user_id", deletedCodeUserID).
		Msg("Recovering messenger link-with-code after code user was already merged")

	if err := s.ownershipClient.TransferOwnership(ctx, deletedCodeUserID, keepUserID); err != nil {
		s.logger.Error().Err(err).Msg("Failed to transfer ownership during messenger link recovery")
		return nil, apperrors.Wrap(
			err,
			fiber.StatusInternalServerError,
			"failed to link accounts: could not transfer ownership in downstream services",
		)
	}

	if err := s.consumeLinkCode(ctx, code); err != nil {
		s.logger.Warn().Err(err).Msg("Failed to delete link code during messenger link recovery")
	}

	return s.issueMessengerLinkTokens(ctx, keepUserID, platform, platformUserID)
}

func (s *Service) recoverWebLinkAfterMerge(
	ctx context.Context,
	keepUserID, deletedCodeUserID int32,
	code string,
) (*LinkWithCodeResult, error) {
	if _, err := s.db.GetUser(ctx, keepUserID); err != nil {
		return nil, apperrors.New(fiber.StatusBadRequest, "code user not found")
	}

	s.logger.Info().
		Int32("keep_user_id", keepUserID).
		Int32("deleted_code_user_id", deletedCodeUserID).
		Msg("Recovering web link-with-code after code user was already merged")

	if err := s.ownershipClient.TransferOwnership(ctx, deletedCodeUserID, keepUserID); err != nil {
		s.logger.Error().Err(err).Msg("Failed to transfer ownership during web link recovery")
		return nil, apperrors.Wrap(
			err,
			fiber.StatusInternalServerError,
			"failed to link accounts: could not transfer ownership in downstream services",
		)
	}

	if err := s.consumeLinkCode(ctx, code); err != nil {
		s.logger.Warn().Err(err).Msg("Failed to delete link code during web link recovery")
	}

	return s.issueWebLinkTokens(ctx, keepUserID)
}
