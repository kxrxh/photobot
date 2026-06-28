package http

import (
	"errors"
	"strconv"
	"strings"

	"csort.ru/classification-service/internal/auth"
	"csort.ru/classification-service/internal/dto"
	"csort.ru/classification-service/internal/httperr"
	"csort.ru/classification-service/internal/logger"
	"csort.ru/classification-service/internal/middleware"
	"csort.ru/classification-service/internal/transport/response"
	"csort.ru/classification-service/internal/user"
	"github.com/gofiber/fiber/v3"
)

const httpUserActiveClassificationComponent = "transport.http.user_active_classification"

type UserActiveClassificationHandler struct {
	service      *user.UserActiveClassificationService
	authClient   *auth.Client
	tokenManager *auth.TokenManager
}

func NewUserActiveClassificationHandler(
	service *user.UserActiveClassificationService,
	authClient *auth.Client,
	tokenManager *auth.TokenManager,
) *UserActiveClassificationHandler {
	return &UserActiveClassificationHandler{
		service:      service,
		authClient:   authClient,
		tokenManager: tokenManager,
	}
}

func (h *UserActiveClassificationHandler) SetUserActiveClassification(c fiber.Ctx) error {
	log := logger.Component(c, httpUserActiveClassificationComponent)
	userCtx := c.Locals(middleware.UserDataKey)
	u, ok := userCtx.(*auth.Identity)
	if !ok || u == nil {
		return httperr.New(fiber.StatusUnauthorized, "User not found in context")
	}

	var req dto.SetUserActiveClassificationRequest
	if err := c.Bind().Body(&req); err != nil {
		return httperr.Wrap(err, fiber.StatusBadRequest, "Invalid request body")
	}

	err := h.service.SetUserActiveClassification(
		c.Context(),
		user.SetUserClassificationRequest{
			UserID:           u.UserID,
			ClassificationID: req.ClassificationID,
		},
	)
	if err != nil {
		return err
	}

	log.Debug().
		Str("classification_id", req.ClassificationID.String()).
		Int32("user_id", u.UserID).
		Msg("set user active classification completed")
	return response.OK(
		c,
		dto.MessageResponse{Message: "User active classification set successfully"},
	)
}

func (h *UserActiveClassificationHandler) GetUserActiveClassificationWithDetails(
	c fiber.Ctx,
) error {
	log := logger.Component(c, httpUserActiveClassificationComponent)
	userIDStr := c.Params("messenger_user_id")
	messengerUserID, err := strconv.ParseInt(userIDStr, 10, 64)
	if err != nil {
		return httperr.Wrap(err, fiber.StatusBadRequest, "Invalid user ID")
	}

	log.Debug().
		Int64("messenger_user_id", messengerUserID).
		Msg("get user active classification")

	platform := strings.TrimSpace(strings.ToLower(c.Query("platform")))
	if platform == "" {
		return httperr.New(fiber.StatusBadRequest, "Query param 'platform' is required")
	}
	if platform != "telegram" && platform != "max" {
		return httperr.New(
			fiber.StatusBadRequest,
			"Query param 'platform' must be 'telegram' or 'max'",
		)
	}

	token := h.tokenManager.GetToken()
	user, err := h.authClient.GetUserByMessengerID(c.Context(), messengerUserID, platform, token)
	if err != nil {
		var httpErr *auth.HTTPStatusError
		if errors.As(err, &httpErr) {
			switch httpErr.StatusCode {
			case fiber.StatusNotFound:
				return httperr.Wrap(err, fiber.StatusNotFound, "User not found")
			case fiber.StatusForbidden:
				return httperr.Wrap(err, fiber.StatusForbidden, "Insufficient permissions")
			case fiber.StatusUnauthorized:
				return httperr.Wrap(
					err,
					fiber.StatusBadGateway,
					"Auth service rejected credentials",
				)
			default:
				return httperr.Wrap(err, fiber.StatusBadGateway, "Auth service request failed")
			}
		}
		return httperr.Wrap(err, fiber.StatusBadGateway, "Auth service request failed")
	}

	log.Debug().Int32("user_id", user.ID).Msg("get user active classification")

	classification, err := h.service.GetUserActiveClassification(c.Context(), user.ID)
	if err != nil {
		return err
	}

	if classification == nil {
		return response.NoContent(c)
	}

	log.Debug().
		Str("classification_id", classification.Classification.ID.String()).
		Msg("get user active classification completed")
	return response.OK(c, CompleteClassificationToResponse(*classification))
}

func (h *UserActiveClassificationHandler) DeleteUserActiveClassification(c fiber.Ctx) error {
	log := logger.Component(c, httpUserActiveClassificationComponent)
	userCtx := c.Locals(middleware.UserDataKey)
	u, ok := userCtx.(*auth.Identity)
	if !ok || u == nil {
		return httperr.New(fiber.StatusUnauthorized, "User not found in context")
	}

	err := h.service.DeleteUserActiveClassification(c.Context(), u.UserID)
	if err != nil {
		return err
	}

	log.Debug().Int32("user_id", u.UserID).Msg("delete user active classification completed")
	return response.OK(
		c,
		dto.MessageResponse{Message: "Active classification removed successfully"},
	)
}

func (h *UserActiveClassificationHandler) GetUserActiveClassification(c fiber.Ctx) error {
	userCtx := c.Locals(middleware.UserDataKey)
	u, ok := userCtx.(*auth.Identity)
	if !ok || u == nil {
		return httperr.New(fiber.StatusUnauthorized, "User not found in context")
	}

	classification, err := h.service.GetUserActiveClassification(c.Context(), u.UserID)
	if err != nil {
		return err
	}

	if classification == nil {
		return response.OK(c, nil)
	}
	return response.OK(c, CompleteClassificationToResponse(*classification))
}
