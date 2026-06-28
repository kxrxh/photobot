package http

import (
	"csort.ru/classification-service/internal/auth"
	"csort.ru/classification-service/internal/classification"
	"csort.ru/classification-service/internal/dto"
	"csort.ru/classification-service/internal/httperr"
	"csort.ru/classification-service/internal/logger"
	"csort.ru/classification-service/internal/middleware"
	"csort.ru/classification-service/internal/transport/response"
	"github.com/gofiber/fiber/v3"
	"github.com/google/uuid"
)

const httpClassificationComponent = "transport.http.classification"

type ClassificationHandler struct {
	service *classification.ClassificationService
}

func NewClassificationHandler(
	service *classification.ClassificationService,
) *ClassificationHandler {
	return &ClassificationHandler{service: service}
}

func (h *ClassificationHandler) GetCompleteClassification(c fiber.Ctx) error {
	idStr := c.Params("id")
	id, err := uuid.Parse(idStr)
	if err != nil {
		return httperr.Wrap(err, fiber.StatusBadRequest, "Invalid classification ID")
	}

	comp, err := h.service.GetCompleteClassification(c.Context(), id)
	if err != nil {
		return err
	}

	return response.OK(c, CompleteClassificationToResponse(*comp))
}

func (h *ClassificationHandler) UpdateClassification(c fiber.Ctx) error {
	log := logger.Component(c, httpClassificationComponent)
	idStr := c.Params("id")
	id, err := uuid.Parse(idStr)
	if err != nil {
		return httperr.Wrap(err, fiber.StatusBadRequest, "Invalid classification ID")
	}

	var req dto.SaveClassificationRequest
	if err := c.Bind().Body(&req); err != nil {
		return httperr.Wrap(err, fiber.StatusBadRequest, "Invalid request body")
	}

	ccls, err := h.service.UpdateClassificationComplete(
		c.Context(),
		id,
		SaveClassificationRequestToDomain(req),
	)
	if err != nil {
		return err
	}

	log.Debug().
		Str("classification_id", ccls.ID.String()).
		Msg("update classification completed")
	return response.OK(c, ClassificationToResponse(*ccls))
}

func (h *ClassificationHandler) DeleteClassification(c fiber.Ctx) error {
	log := logger.Component(c, httpClassificationComponent)
	idStr := c.Params("id")
	id, err := uuid.Parse(idStr)
	if err != nil {
		return httperr.Wrap(err, fiber.StatusBadRequest, "Invalid classification ID")
	}

	err = h.service.DeleteClassification(c.Context(), id)
	if err != nil {
		return err
	}

	log.Debug().Str("classification_id", id.String()).Msg("delete classification completed")
	return response.OK(c, dto.MessageResponse{Message: "Classification deleted successfully"})
}

func (h *ClassificationHandler) CreateCompleteClassification(c fiber.Ctx) error {
	log := logger.Component(c, httpClassificationComponent)
	var req dto.SaveClassificationRequest
	if err := c.Bind().Body(&req); err != nil {
		return httperr.Wrap(err, fiber.StatusBadRequest, "Invalid request body")
	}

	userCtx := c.Locals(middleware.UserDataKey)
	user, ok := userCtx.(*auth.Identity)
	if !ok {
		return httperr.New(fiber.StatusUnauthorized, "User not found in context")
	}

	comp, err := h.service.CreateCompleteClassification(
		c.Context(),
		SaveClassificationRequestToDomain(req),
		user.UserID,
	)
	if err != nil {
		return err
	}

	log.Debug().
		Str("classification_id", comp.Classification.ID.String()).
		Msg("create classification completed")
	return response.Created(c, CompleteClassificationToResponse(*comp))
}

func (h *ClassificationHandler) ListClassifications(c fiber.Ctx) error {
	log := logger.Component(c, httpClassificationComponent)
	userCtx := c.Locals(middleware.UserDataKey)
	user, ok := userCtx.(*auth.Identity)
	if !ok {
		return httperr.New(fiber.StatusUnauthorized, "User not found in context")
	}

	filters := classification.ClassificationFilters{
		CreatedBy: &user.UserID,
	}

	if name := c.Query("name", ""); name != "" {
		filters.Name = &name
	}

	productIDStr := c.Query("product_id")
	if productIDStr != "" {
		productID, err := uuid.Parse(productIDStr)
		if err != nil {
			return httperr.Wrap(err, fiber.StatusBadRequest, "Invalid product ID")
		}
		filters.ProductID = &productID
	}

	log.Debug().Interface("filters", filters).Msg("list classifications")

	classifications, activeClassification, err := h.service.ListClassificationsForUser(
		c.Context(),
		user.UserID,
		filters,
	)
	if err != nil {
		return err
	}

	listResp := dto.ListClassificationsResponse{
		Classifications:      ClassificationsToResponse(classifications),
		ActiveClassification: ClassificationPtrToResponse(activeClassification),
	}

	log.Debug().
		Int("classifications_count", len(classifications)).
		Msg("list classifications completed")
	return response.OK(c, listResp)
}

func (h *ClassificationHandler) MakeClassificationPublic(c fiber.Ctx) error {
	log := logger.Component(c, httpClassificationComponent)
	idStr := c.Params("id")
	id, err := uuid.Parse(idStr)
	if err != nil {
		return httperr.Wrap(err, fiber.StatusBadRequest, "Invalid classification ID")
	}

	err = h.service.UpdateClassificationPublic(c.Context(), id, true)
	if err != nil {
		return err
	}

	log.Debug().
		Str("classification_id", id.String()).
		Msg("make classification public completed")
	return response.OK(c, dto.MessageResponse{Message: "Classification made public successfully"})
}

func (h *ClassificationHandler) MakeClassificationPrivate(c fiber.Ctx) error {
	log := logger.Component(c, httpClassificationComponent)
	idStr := c.Params("id")
	id, err := uuid.Parse(idStr)
	if err != nil {
		return httperr.Wrap(err, fiber.StatusBadRequest, "Invalid classification ID")
	}

	err = h.service.UpdateClassificationPublic(c.Context(), id, false)
	if err != nil {
		return err
	}

	log.Debug().
		Str("classification_id", id.String()).
		Msg("make classification private completed")
	return response.OK(c, dto.MessageResponse{Message: "Classification made private successfully"})
}
