package http

import (
	"strings"

	"csort.ru/auth-service/internal/auth"
	"csort.ru/auth-service/internal/dto"
	"csort.ru/auth-service/internal/logger"
	"csort.ru/auth-service/internal/transport/response"
	"csort.ru/auth-service/internal/user"
	"github.com/go-playground/validator/v10"
	"github.com/gofiber/fiber/v3"
)

const httpAuthComponent = "transport.http.auth"

type AuthHandler struct {
	authService *auth.Service
	userService *user.Service
	validator   *validator.Validate
}

func NewAuthHandler(
	authService *auth.Service,
	userService *user.Service,
	v *validator.Validate,
) *AuthHandler {
	return &AuthHandler{
		authService: authService,
		userService: userService,
		validator:   v,
	}
}

func (h *AuthHandler) AdminLogin(c fiber.Ctx) error {
	var req dto.AdminLoginRequest
	if err := c.Bind().Body(&req); err != nil {
		return response.Fail(c, fiber.StatusBadRequest, "Cannot parse request", err.Error())
	}

	if err := h.validator.Struct(req); err != nil {
		return response.Fail(c, fiber.StatusBadRequest, "Validation failed", err.Error())
	}

	keyPair, roles, err := h.authService.AdminLogin(
		c.Context(),
		req.Login,
		req.Password,
		auth.GrantTypePassword,
	)
	if err != nil {
		return err
	}

	return response.OK(c, AdminLoginResponseFromDomain(keyPair, roles))
}

func (h *AuthHandler) Login(c fiber.Ctx) error {
	log := logger.Component(c, httpAuthComponent)
	gtyHeader := c.Get("X-Grant-Type")
	gty := auth.GrantTypeInitData
	if gtyHeader != "" {
		gty = auth.GrantType(gtyHeader)
		if !auth.IsValidGrantType(string(gty)) {
			return response.Fail(c, fiber.StatusBadRequest, "Invalid X-Grant-Type header", nil)
		}
	}

	params := &auth.LoginParams{
		GTY: gty,
	}
	log.Info().
		Ctx(c.Context()).
		Str("log.type", "auth_handler.login").
		Str("auth.grant_type", string(gty)).
		Msg("login request received")

	switch gty {
	case auth.GrantTypeService:
		var req dto.ServiceLoginRequest
		if err := c.Bind().Body(&req); err != nil {
			return response.Fail(
				c,
				fiber.StatusBadRequest,
				"Cannot parse service login request",
				err.Error(),
			)
		}

		if err := h.validator.Struct(req); err != nil {
			log.Warn().Err(err).Msg("validation rejected")
			return response.Fail(c, fiber.StatusBadRequest, "Validation failed", err.Error())
		}

		params.ServiceID = &req.ServiceID
		params.ServiceSecret = &req.ServiceSecret
		if req.Audience != "" {
			params.Audience = &req.Audience
		}

	case auth.GrantTypePassword, auth.GrantTypeUserPassword:
		var req dto.AdminLoginRequest
		if err := c.Bind().Body(&req); err != nil {
			return response.Fail(
				c,
				fiber.StatusBadRequest,
				"Cannot parse password login request",
				err.Error(),
			)
		}

		if err := h.validator.Struct(req); err != nil {
			return response.Fail(c, fiber.StatusBadRequest, "Validation failed", err.Error())
		}

		params.Login = &req.Login
		params.Password = &req.Password

	case auth.GrantTypeInitData:
		botName := c.Get("X-Bot-Name", "")
		if botName == "" {
			return response.Fail(c, fiber.StatusBadRequest, "X-Bot-Name header is required", nil)
		}

		platform := c.Get("X-Messenger-Platform", "")
		if platform == "" {
			platform = "telegram"
		}

		initData := c.Get("X-Init-Data", "")
		if initData == "" {
			return response.Fail(c, fiber.StatusBadRequest, "X-Init-Data header is required", nil)
		}

		params.InitData = &initData
		params.BotName = &botName
		params.MessengerPlatform = &platform

	default:
		return response.Fail(c, fiber.StatusBadRequest, "Unsupported grant type", nil)
	}

	keyPair, roles, err := h.authService.Login(c.Context(), params)
	if err != nil {
		log.Warn().
			Ctx(c.Context()).
			Err(err).
			Str("log.type", "auth_handler.login").
			Str("auth.grant_type", string(gty)).
			Msg("login request failed")
		return err
	}
	log.Info().
		Ctx(c.Context()).
		Str("log.type", "auth_handler.login").
		Str("auth.grant_type", string(gty)).
		Int("auth.roles_count", len(roles)).
		Msg("login request succeeded")

	return response.OK(c, LoginResponseFromDomain(keyPair, roles))
}

func (h *AuthHandler) Refresh(c fiber.Ctx) error {
	log := logger.Component(c, httpAuthComponent)
	var req dto.RefreshRequest
	if err := c.Bind().Body(&req); err != nil {
		return response.Fail(c, fiber.StatusBadRequest, "Cannot parse refresh request", err.Error())
	}

	if err := h.validator.Struct(req); err != nil {
		return response.Fail(c, fiber.StatusBadRequest, "Validation failed", err.Error())
	}

	log.Info().
		Ctx(c.Context()).
		Str("log.type", "auth_handler.refresh").
		Msg("refresh request received")

	keyPair, err := h.authService.Refresh(c.Context(), req.RefreshToken)
	if err != nil {
		log.Warn().
			Ctx(c.Context()).
			Err(err).
			Str("log.type", "auth_handler.refresh").
			Msg("refresh request failed")
		return err
	}
	log.Info().
		Ctx(c.Context()).
		Str("log.type", "auth_handler.refresh").
		Msg("refresh request succeeded")
	return response.OK(c, keyPair)
}

func (h *AuthHandler) GetJWKS(c fiber.Ctx) error {
	keyManager := auth.GetKeyManager()
	if jwksJSON, err := keyManager.GetJWKSJSON(); err == nil && len(jwksJSON) > 0 {
		c.Set("Content-Type", "application/json")
		return c.Send(jwksJSON)
	}
	jwks := keyManager.GetJWKS()
	if jwks == nil {
		return response.Fail(c, fiber.StatusInternalServerError, "Failed to generate JWKS", nil)
	}
	return c.JSON(jwks)
}

func (h *AuthHandler) RegisterUser(c fiber.Ctx) error {
	log := logger.Component(c, httpAuthComponent)
	var req dto.RegisterRequest

	if err := c.Bind().Body(&req); err != nil {
		return response.Fail(c, fiber.StatusBadRequest, "Invalid request body", err.Error())
	}

	if err := h.validator.Struct(req); err != nil {
		log.Warn().Err(err).Msg("validation rejected")
		return response.Fail(c, fiber.StatusBadRequest, "Validation failed", err.Error())
	}

	botName := c.Get("X-Bot-Name")
	if botName == "" {
		return response.Fail(c, fiber.StatusBadRequest, "X-Bot-Name header is required", nil)
	}

	platform := c.Get("X-Messenger-Platform", "")
	if platform == "" {
		platform = "telegram"
	}

	initData := c.Get("X-Init-Data")
	if initData == "" {
		return response.Fail(c, fiber.StatusBadRequest, "X-Init-Data header is required", nil)
	}

	var (
		userData *auth.UserData
		err      error
	)

	switch platform {
	case "telegram":
		userData, err = h.authService.ValidateTelegramData(c.Context(), initData, botName)
	case "max":
		userData, err = h.authService.ValidateMaxData(c.Context(), initData, botName)
	default:
		return response.Fail(c, fiber.StatusBadRequest, "Unsupported messenger platform", nil)
	}

	if err != nil {
		return err
	}

	if strings.TrimSpace(req.FullName) == "" {
		fullName := strings.TrimSpace(userData.FirstName + " " + userData.LastName)
		if fullName == "" && userData.Username != "" {
			fullName = strings.TrimPrefix(userData.Username, "@")
		}
		req.FullName = fullName
	}

	var userResp *user.User
	domainReq := RegisterRequestToDomain(req)
	switch platform {
	case "telegram":
		userResp, err = h.userService.Create(c.Context(), userData.ID, &domainReq)
	case "max":
		userResp, err = h.userService.CreateWithMaxID(c.Context(), userData.ID, &domainReq)
	}
	if err != nil {
		return err
	}

	return response.Created(c, fiber.Map{
		"message": "User registered successfully",
		"data":    userResp,
	})
}

func (h *AuthHandler) RequestLinkCode(c fiber.Ctx) error {
	log := logger.Component(c, httpAuthComponent)
	userIDPtr, ok := c.Locals(auth.LocalsUserID).(*int32)
	if !ok || userIDPtr == nil {
		return response.Fail(c, fiber.StatusUnauthorized, "User not authenticated", nil)
	}

	result, err := h.authService.RequestLinkCode(c.Context(), *userIDPtr)
	if err != nil {
		log.Warn().
			Ctx(c.Context()).
			Err(err).
			Str("log.type", "auth_handler.request_link_code").
			Msg("link code request failed")
		return err
	}
	log.Info().
		Ctx(c.Context()).
		Str("log.type", "auth_handler.request_link_code").
		Msg("link code request succeeded")

	return response.OK(c, result)
}

func (h *AuthHandler) LinkWithCode(c fiber.Ctx) error {
	log := logger.Component(c, httpAuthComponent)
	botName := c.Get("X-Bot-Name")
	if botName == "" {
		return response.Fail(c, fiber.StatusBadRequest, "X-Bot-Name header is required", nil)
	}

	platform := c.Get("X-Messenger-Platform", "")
	platform = strings.TrimSpace(strings.ToLower(platform))
	if platform == "" {
		return response.Fail(
			c,
			fiber.StatusBadRequest,
			"X-Messenger-Platform header is required (max or telegram)",
			nil,
		)
	}
	if platform != "max" && platform != "telegram" {
		return response.Fail(
			c,
			fiber.StatusBadRequest,
			"X-Messenger-Platform must be max or telegram",
			nil,
		)
	}

	initData := c.Get("X-Init-Data")
	if initData == "" {
		return response.Fail(c, fiber.StatusBadRequest, "X-Init-Data header is required", nil)
	}

	var req dto.LinkWithCodeRequest
	if err := c.Bind().Body(&req); err != nil {
		return response.Fail(c, fiber.StatusBadRequest, "Invalid request body", err.Error())
	}
	if err := h.validator.Struct(req); err != nil {
		return response.Fail(c, fiber.StatusBadRequest, "Validation failed", err.Error())
	}

	userIDPtr, ok := c.Locals(auth.LocalsUserID).(*int32)
	if !ok || userIDPtr == nil {
		return response.Fail(c, fiber.StatusUnauthorized, "User not authenticated", nil)
	}

	result, err := h.authService.LinkWithCode(
		c.Context(),
		*userIDPtr,
		req.Code,
		initData,
		botName,
		platform,
	)
	if err != nil {
		log.Warn().
			Ctx(c.Context()).
			Err(err).
			Str("log.type", "auth_handler.link_with_code").
			Str("messenger.platform", platform).
			Msg("link with code request failed")
		return err
	}
	log.Info().
		Ctx(c.Context()).
		Str("log.type", "auth_handler.link_with_code").
		Str("messenger.platform", platform).
		Msg("link with code request succeeded")

	return response.OK(c, result)
}

func (h *AuthHandler) RegisterWeb(c fiber.Ctx) error {
	log := logger.Component(c, httpAuthComponent)
	var req dto.WebRegisterRequest
	if err := c.Bind().Body(&req); err != nil {
		return response.Fail(c, fiber.StatusBadRequest, "Invalid request body", err.Error())
	}
	if err := h.validator.Struct(req); err != nil {
		log.Warn().Err(err).Msg("validation rejected")
		return response.Fail(c, fiber.StatusBadRequest, "Validation failed", err.Error())
	}

	domainReq := user.WebRegisterRequest{
		Login:            req.Login,
		Password:         req.Password,
		OrganizationName: req.OrganizationName,
		INN:              req.INN,
		FullName:         req.FullName,
		PhoneNumber:      req.PhoneNumber,
	}
	result, err := h.authService.RegisterWeb(c.Context(), h.userService, &domainReq)
	if err != nil {
		return err
	}
	return response.Created(c, result)
}

func (h *AuthHandler) SetupWebAccess(c fiber.Ctx) error {
	var req dto.SetupWebAccessRequest
	if err := c.Bind().Body(&req); err != nil {
		return response.Fail(c, fiber.StatusBadRequest, "Invalid request body", err.Error())
	}
	if err := h.validator.Struct(req); err != nil {
		return response.Fail(c, fiber.StatusBadRequest, "Validation failed", err.Error())
	}
	userIDPtr, ok := c.Locals(auth.LocalsUserID).(*int32)
	if !ok || userIDPtr == nil {
		return response.Fail(c, fiber.StatusUnauthorized, "User not authenticated", nil)
	}
	result, err := h.authService.SetupWebAccess(c.Context(), *userIDPtr, req.Login, req.Password)
	if err != nil {
		return err
	}
	return response.OK(c, result)
}

func (h *AuthHandler) ChangePassword(c fiber.Ctx) error {
	var req dto.ChangePasswordRequest
	if err := c.Bind().Body(&req); err != nil {
		return response.Fail(c, fiber.StatusBadRequest, "Invalid request body", err.Error())
	}
	if err := h.validator.Struct(req); err != nil {
		return response.Fail(c, fiber.StatusBadRequest, "Validation failed", err.Error())
	}
	userIDPtr, ok := c.Locals(auth.LocalsUserID).(*int32)
	if !ok || userIDPtr == nil {
		return response.Fail(c, fiber.StatusUnauthorized, "User not authenticated", nil)
	}
	if err := h.authService.ChangePassword(
		c.Context(),
		*userIDPtr,
		req.CurrentPassword,
		req.NewPassword,
	); err != nil {
		return err
	}
	return response.OK(c, fiber.Map{"message": "Password changed successfully"})
}

func (h *AuthHandler) ForgotPassword(c fiber.Ctx) error {
	log := logger.Component(c, httpAuthComponent)
	var req dto.ForgotPasswordRequest
	if err := c.Bind().Body(&req); err != nil {
		return response.Fail(c, fiber.StatusBadRequest, "Invalid request body", err.Error())
	}
	if err := h.validator.Struct(req); err != nil {
		return response.Fail(c, fiber.StatusBadRequest, "Validation failed", err.Error())
	}
	if err := h.authService.ForgotPassword(c.Context(), req.Login); err != nil {
		log.Warn().Err(err).Str("login", req.Login).Msg("forgot password failed")
	}
	return response.OK(c, fiber.Map{
		"message": "If the account exists, reset instructions were sent",
	})
}

func (h *AuthHandler) ResetPassword(c fiber.Ctx) error {
	var req dto.ResetPasswordRequest
	if err := c.Bind().Body(&req); err != nil {
		return response.Fail(c, fiber.StatusBadRequest, "Invalid request body", err.Error())
	}
	if err := h.validator.Struct(req); err != nil {
		return response.Fail(c, fiber.StatusBadRequest, "Validation failed", err.Error())
	}
	if err := h.authService.ResetPassword(
		c.Context(),
		req.Login,
		req.Otp,
		req.NewPassword,
	); err != nil {
		return err
	}
	return response.OK(c, fiber.Map{"message": "Password reset successfully"})
}

func (h *AuthHandler) ResetPasswordRecovery(c fiber.Ctx) error {
	var req dto.ResetPasswordRecoveryRequest
	if err := c.Bind().Body(&req); err != nil {
		return response.Fail(c, fiber.StatusBadRequest, "Invalid request body", err.Error())
	}
	if err := h.validator.Struct(req); err != nil {
		return response.Fail(c, fiber.StatusBadRequest, "Validation failed", err.Error())
	}
	if err := h.authService.ResetPasswordRecovery(
		c.Context(),
		req.Login,
		req.RecoveryCode,
		req.NewPassword,
	); err != nil {
		return err
	}
	return response.OK(c, fiber.Map{"message": "Password reset successfully"})
}

func (h *AuthHandler) LinkWithCodeFromWeb(c fiber.Ctx) error {
	log := logger.Component(c, httpAuthComponent)
	var req dto.LinkWithCodeRequest
	if err := c.Bind().Body(&req); err != nil {
		return response.Fail(c, fiber.StatusBadRequest, "Invalid request body", err.Error())
	}
	if err := h.validator.Struct(req); err != nil {
		return response.Fail(c, fiber.StatusBadRequest, "Validation failed", err.Error())
	}
	userIDPtr, ok := c.Locals(auth.LocalsUserID).(*int32)
	if !ok || userIDPtr == nil {
		return response.Fail(c, fiber.StatusUnauthorized, "User not authenticated", nil)
	}
	result, err := h.authService.LinkWithCodeFromWeb(c.Context(), *userIDPtr, req.Code)
	if err != nil {
		log.Warn().Err(err).Msg("link with code from web failed")
		return err
	}
	return response.OK(c, result)
}
