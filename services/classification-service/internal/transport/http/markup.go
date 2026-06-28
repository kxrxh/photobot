package http

import (
	"csort.ru/classification-service/internal/auth"
	"csort.ru/classification-service/internal/dto"
	"csort.ru/classification-service/internal/httperr"
	"csort.ru/classification-service/internal/logger"
	"csort.ru/classification-service/internal/markup"
	"csort.ru/classification-service/internal/middleware"
	"csort.ru/classification-service/internal/transport/response"
	"github.com/go-playground/validator/v10"
	"github.com/gofiber/fiber/v3"
	"github.com/google/uuid"
)

const httpMarkupComponent = "transport.http.markup"

type MarkupHandler struct {
	service   *markup.MarkupService
	validator *validator.Validate
}

func NewMarkupHandler(service *markup.MarkupService, v *validator.Validate) *MarkupHandler {
	return &MarkupHandler{
		service:   service,
		validator: v,
	}
}

func (h *MarkupHandler) CreateMarkup(c fiber.Ctx) error {
	log := logger.Component(c, httpMarkupComponent)
	userCtx := c.Locals(middleware.UserDataKey)
	user, ok := userCtx.(*auth.Identity)
	if !ok {
		return httperr.New(fiber.StatusUnauthorized, "User not found in context")
	}

	var req dto.SaveMarkupRequest
	if err := c.Bind().Body(&req); err != nil {
		return httperr.Wrap(err, fiber.StatusBadRequest, "Invalid request body")
	}

	domainReq := SaveMarkupRequestToDomain(req, user.UserID)
	if err := h.validator.Struct(domainReq); err != nil {
		return httperr.Wrap(err, fiber.StatusBadRequest, "Invalid request body")
	}
	if domainReq.CreatedBy <= 0 {
		return httperr.New(fiber.StatusBadRequest, "CreatedBy must be set from authentication")
	}
	m, err := h.service.CreateMarkup(c.Context(), domainReq)
	if err != nil {
		return err
	}

	log.Debug().Str("markup_id", m.ID.String()).Msg("create markup completed")
	return response.Created(c, MarkupPtrToResponse(m))
}

func (h *MarkupHandler) ListMarkups(c fiber.Ctx) error {
	log := logger.Component(c, httpMarkupComponent)
	userCtx := c.Locals(middleware.UserDataKey)
	user, ok := userCtx.(*auth.Identity)
	if !ok {
		return httperr.New(fiber.StatusUnauthorized, "User not found in context")
	}

	filters := markup.MarkupFilters{
		CreatedBy: &user.UserID,
	}
	if name := c.Query("name", ""); name != "" {
		filters.Name = &name
	}

	markups, err := h.service.ListMarkups(c.Context(), filters)
	if err != nil {
		return err
	}

	log.Debug().Int("markups_count", len(markups)).Msg("list markups completed")
	return response.OK(c, MarkupsToResponse(markups))
}

func (h *MarkupHandler) GetMarkup(c fiber.Ctx) error {
	idStr := c.Params("id")
	id, err := uuid.Parse(idStr)
	if err != nil {
		return httperr.Wrap(err, fiber.StatusBadRequest, "Invalid markup ID")
	}

	m, err := h.service.GetMarkup(c.Context(), id)
	if err != nil {
		return err
	}

	return response.OK(c, MarkupPtrToResponse(m))
}

func (h *MarkupHandler) UpdateMarkup(c fiber.Ctx) error {
	log := logger.Component(c, httpMarkupComponent)
	idStr := c.Params("id")
	id, err := uuid.Parse(idStr)
	if err != nil {
		return httperr.Wrap(err, fiber.StatusBadRequest, "Invalid markup ID")
	}

	userCtx := c.Locals(middleware.UserDataKey)
	user, ok := userCtx.(*auth.Identity)
	if !ok {
		return httperr.New(fiber.StatusUnauthorized, "User not found in context")
	}

	var req dto.SaveMarkupRequest
	if err := c.Bind().Body(&req); err != nil {
		return httperr.Wrap(err, fiber.StatusBadRequest, "Invalid request body")
	}

	domainReq := SaveMarkupRequestToDomain(req, user.UserID)
	if err := h.validator.Struct(domainReq); err != nil {
		return httperr.Wrap(err, fiber.StatusBadRequest, "Invalid request body")
	}

	m, err := h.service.UpdateMarkup(c.Context(), id, domainReq)
	if err != nil {
		return err
	}

	log.Debug().Str("markup_id", m.ID.String()).Msg("update markup completed")
	return response.OK(c, MarkupPtrToResponse(m))
}

func (h *MarkupHandler) DeleteMarkup(c fiber.Ctx) error {
	log := logger.Component(c, httpMarkupComponent)
	idStr := c.Params("id")
	id, err := uuid.Parse(idStr)
	if err != nil {
		return httperr.Wrap(err, fiber.StatusBadRequest, "Invalid markup ID")
	}

	err = h.service.DeleteMarkup(c.Context(), id)
	if err != nil {
		return err
	}

	log.Debug().Str("markup_id", id.String()).Msg("delete markup completed")
	return response.OK(c, dto.MessageResponse{Message: "Markup deleted successfully"})
}
