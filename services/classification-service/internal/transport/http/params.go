package http

import (
	"csort.ru/classification-service/internal/dto"
	"csort.ru/classification-service/internal/httperr"
	"csort.ru/classification-service/internal/params"
	"csort.ru/classification-service/internal/transport/response"
	"github.com/go-playground/validator/v10"
	"github.com/gofiber/fiber/v3"
	"github.com/google/uuid"
)

type ClassificationParamsHandler struct {
	service   *params.ClassificationParamsService
	validator *validator.Validate
}

func NewClassificationParamsHandler(
	service *params.ClassificationParamsService,
	v *validator.Validate,
) *ClassificationParamsHandler {
	return &ClassificationParamsHandler{service: service, validator: v}
}

func (h *ClassificationParamsHandler) Create(c fiber.Ctx) error {
	var req dto.SaveParamRequest
	if err := c.Bind().Body(&req); err != nil {
		return httperr.Wrap(err, fiber.StatusBadRequest, "Invalid request body")
	}
	if err := h.validator.Struct(req); err != nil {
		return httperr.Wrap(err, fiber.StatusBadRequest, "Invalid request body")
	}
	item, err := h.service.Create(c.Context(), SaveParamRequestToDomain(req))
	if err != nil {
		return err
	}
	return response.Created(c, ParamPtrToResponse(item))
}

func (h *ClassificationParamsHandler) Delete(c fiber.Ctx) error {
	idStr := c.Params("id")
	id, err := uuid.Parse(idStr)
	if err != nil {
		return httperr.Wrap(err, fiber.StatusBadRequest, "Invalid ID")
	}
	if err := h.service.Delete(c.Context(), id); err != nil {
		return err
	}
	return response.OK(c, dto.MessageResponse{Message: "deleted"})
}

func (h *ClassificationParamsHandler) List(c fiber.Ctx) error {
	items, err := h.service.List(c.Context())
	if err != nil {
		return err
	}
	return response.OK(c, ParamsToResponse(items))
}
