package http

import (
	"strconv"

	"csort.ru/auth-service/internal/auth"
	"csort.ru/auth-service/internal/dto"
	"csort.ru/auth-service/internal/logger"
	"csort.ru/auth-service/internal/transport/response"
	"csort.ru/auth-service/internal/user"
	"github.com/go-playground/validator/v10"
	"github.com/gofiber/fiber/v3"
)

const httpUserComponent = "transport.http.user"

type UserHandler struct {
	userService *user.Service
	authService *auth.Service
	validator   *validator.Validate
}

func NewUserHandler(
	userService *user.Service,
	authService *auth.Service,
	v *validator.Validate,
) *UserHandler {
	return &UserHandler{
		userService: userService,
		authService: authService,
		validator:   v,
	}
}

// GetUser retrieves a user by ID
func (h *UserHandler) GetUser(c fiber.Ctx) error {
	idStr := c.Req().Params("id")
	id, err := strconv.ParseInt(idStr, 10, 32)
	if err != nil {
		return response.Fail(c, fiber.StatusBadRequest, "User ID must be a valid integer", nil)
	}

	userResp, err := (*h.userService).Get(c.Context(), int32(id))
	if err != nil {
		return err
	}

	return response.OK(c, userResp)
}

// GetMe retrieves the current authenticated user
func (h *UserHandler) GetMe(c fiber.Ctx) error {
	userID, ok := c.Locals(auth.LocalsUserID).(*int32)
	if !ok || userID == nil {
		return response.Fail(c, fiber.StatusUnauthorized, "User not authenticated", nil)
	}

	userResp, err := (*h.userService).Get(c.Context(), *userID)
	if err != nil {
		return err
	}

	return response.OK(c, userResp)
}

// UpdateMe updates the current authenticated user
func (h *UserHandler) UpdateMe(c fiber.Ctx) error {
	log := logger.Component(c, httpUserComponent)
	userID, ok := c.Locals(auth.LocalsUserID).(*int32)
	if !ok || userID == nil {
		return response.Fail(c, fiber.StatusUnauthorized, "User not authenticated", nil)
	}

	var req dto.UserRequest

	if err := c.Bind().Body(&req); err != nil {
		return response.Fail(c, fiber.StatusBadRequest, "Invalid request body", err.Error())
	}

	if err := h.validator.Struct(req); err != nil {
		log.Warn().Err(err).Msg("validation rejected")
		return response.Fail(c, fiber.StatusBadRequest, "Validation failed", err.Error())
	}

	domainReq := UserRequestToDomain(req)
	userResp, err := h.userService.Update(c.Context(), *userID, &domainReq)
	if err != nil {
		return err
	}

	return response.OK(c, fiber.Map{
		"message": "User updated successfully",
		"data":    userResp,
	})
}

func (h *UserHandler) GetUserByMessengerId(c fiber.Ctx) error {
	idStr := c.Req().Params("id")
	id, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil {
		return response.Fail(c, fiber.StatusBadRequest, "Messenger ID must be a valid integer", nil)
	}

	platform := c.Query("platform")
	if platform == "" {
		return response.Fail(c, fiber.StatusBadRequest, "Query param 'platform' is required", nil)
	}

	var userResp *user.User
	switch platform {
	case "telegram":
		userResp, err = (*h.userService).GetByTelegramId(c.Context(), id)
	case "max":
		userResp, err = (*h.userService).GetByMaxId(c.Context(), id)
	default:
		return response.Fail(
			c,
			fiber.StatusBadRequest,
			"Query param 'platform' must be 'telegram' or 'max'",
			nil,
		)
	}
	if err != nil {
		return err
	}

	return response.OK(c, userResp)
}

// ListUsers retrieves all users
func (h *UserHandler) ListUsers(c fiber.Ctx) error {
	users, err := (*h.userService).List(c.Context())
	if err != nil {
		return err
	}

	return response.OK(c, users)
}

// UpdateUser updates an existing user
func (h *UserHandler) UpdateUser(c fiber.Ctx) error {
	log := logger.Component(c, httpUserComponent)
	idStr := c.Req().Params("id")
	id, err := strconv.ParseInt(idStr, 10, 32)
	if err != nil {
		return response.Fail(c, fiber.StatusBadRequest, "User ID must be a valid integer", nil)
	}

	var req dto.UserRequest

	if err := c.Bind().Body(&req); err != nil {
		return response.Fail(c, fiber.StatusBadRequest, "Invalid request body", err.Error())
	}

	if err := h.validator.Struct(req); err != nil {
		log.Warn().Err(err).Msg("validation rejected")
		return response.Fail(c, fiber.StatusBadRequest, "Validation failed", err.Error())
	}

	targetUserID := int32(id)

	domainReq := UserRequestToDomain(req)
	userResp, err := h.userService.Update(c.Context(), targetUserID, &domainReq)
	if err != nil {
		return err
	}

	return response.OK(c, fiber.Map{
		"message": "User updated successfully",
		"data":    userResp,
	})
}
