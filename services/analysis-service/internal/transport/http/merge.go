package http

import (
	"csort.ru/analysis-service/internal/apierrors"
	"csort.ru/analysis-service/internal/dto"
	reqsvc "csort.ru/analysis-service/internal/requests"
	"csort.ru/analysis-service/internal/transport/response"
	"github.com/gofiber/fiber/v3"
	"github.com/rs/zerolog"
)

type OwnershipHandler struct {
	log     zerolog.Logger
	service *reqsvc.Service
}

func NewOwnershipHandler(log zerolog.Logger, service *reqsvc.Service) *OwnershipHandler {
	return &OwnershipHandler{log: log, service: service}
}

func (h *OwnershipHandler) TransferOwnership(c fiber.Ctx) error {
	var req dto.OwnershipTransferRequest
	if err := c.Bind().Body(&req); err != nil {
		h.log.Error().Err(err).Msg("ownership transfer rejected: invalid body")
		return apierrors.BadRequest("Invalid request body")
	}
	if req.FromUserID <= 0 || req.ToUserID <= 0 {
		h.log.Warn().
			Int32("from_user_id", req.FromUserID).
			Int32("to_user_id", req.ToUserID).
			Msg("ownership transfer rejected: invalid ids")
		return apierrors.BadRequest("from_user_id and to_user_id must be positive")
	}
	if req.FromUserID == req.ToUserID {
		h.log.Warn().
			Int32("user_id", req.FromUserID).
			Msg("ownership transfer rejected: from and to must differ")
		return apierrors.BadRequest("from_user_id and to_user_id must differ")
	}
	if err := h.service.TransferRequestOwnership(
		c.Context(),
		req.FromUserID,
		req.ToUserID,
	); err != nil {
		h.log.Error().
			Err(err).
			Int32("from_user_id", req.FromUserID).
			Int32("to_user_id", req.ToUserID).
			Msg("ownership transfer failed")
		return err
	}
	h.log.Info().
		Int32("from_user_id", req.FromUserID).
		Int32("to_user_id", req.ToUserID).
		Msg("ownership transfer completed")
	return response.OK(c, dto.OwnershipTransferResponse{Message: "Ownership transfer completed"})
}
