package http

import (
	"csort.ru/classification-service/internal/correlation"
	"csort.ru/classification-service/internal/dto"
	"csort.ru/classification-service/internal/httperr"
	"csort.ru/classification-service/internal/transport/response"
	"github.com/gofiber/fiber/v3"
)

type CorrelationHandler struct {
	service *correlation.CorrelationService
}

func NewCorrelationHandler(service *correlation.CorrelationService) *CorrelationHandler {
	return &CorrelationHandler{service: service}
}

func (h *CorrelationHandler) CalculateCorrelation(c fiber.Ctx) error {
	var req dto.CorrelationRequest
	if err := c.Bind().Body(&req); err != nil {
		return httperr.Wrap(err, fiber.StatusBadRequest, "Invalid request body")
	}

	ctx := c.Context()
	domainReq := CorrelationRequestToDomain(req)
	result, err := h.service.CalculateCorrelation(ctx, &domainReq)
	if err != nil {
		return err
	}

	return response.OK(c, CorrelationsWithTestToResponse(result))
}
