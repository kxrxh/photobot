package http

import (
	"csort.ru/classification-service/internal/dto"
	"csort.ru/classification-service/internal/httperr"
	"csort.ru/classification-service/internal/ownership"
	"csort.ru/classification-service/internal/transport/response"
	"github.com/gofiber/fiber/v3"
)

type OwnershipHandler struct {
	service *ownership.Service
}

func NewOwnershipHandler(service *ownership.Service) *OwnershipHandler {
	return &OwnershipHandler{service: service}
}

func (h *OwnershipHandler) TransferOwnership(c fiber.Ctx) error {
	var req dto.OwnershipTransferRequest
	if err := c.Bind().Body(&req); err != nil {
		return httperr.Wrap(err, fiber.StatusBadRequest, "Invalid request body")
	}
	if req.FromUserID <= 0 || req.ToUserID <= 0 {
		return httperr.New(fiber.StatusBadRequest, "from_user_id and to_user_id must be positive")
	}
	if req.FromUserID == req.ToUserID {
		return httperr.New(fiber.StatusBadRequest, "from_user_id and to_user_id must differ")
	}
	fromUserID, toUserID := OwnershipTransferRequestToParams(req)
	if err := h.service.TransferOwnership(c.Context(), fromUserID, toUserID); err != nil {
		return err
	}
	return response.OK(c, dto.OwnershipTransferResponse{
		Message: "Ownership transfer completed",
	})
}
