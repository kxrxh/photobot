package http

import (
	"strconv"

	"csort.ru/coffeebot/internal/middleware"
	"csort.ru/coffeebot/internal/proposal"
	"csort.ru/coffeebot/internal/transport"
	"csort.ru/coffeebot/internal/transport/response"
	"github.com/go-playground/validator/v10"
	"github.com/gofiber/fiber/v3"
	"github.com/rs/zerolog"
)

type ProposalHandler struct {
	proposalService *proposal.Service
	validator       *validator.Validate
	log             zerolog.Logger
}

func NewProposalHandler(proposalService *proposal.Service, zlog zerolog.Logger) *ProposalHandler {
	return &ProposalHandler{
		proposalService: proposalService,
		validator:       validator.New(),
		log:             zlog,
	}
}

func (h *ProposalHandler) CreateProposal(c fiber.Ctx) error {
	var params proposal.CreateProposalParams
	if err := c.Bind().Body(&params); err != nil {
		h.log.Error().Err(err).Msg("Failed to parse create proposal")
		return response.Fail(c, fiber.StatusBadRequest, "Invalid request body", nil)
	}

	if err := h.validator.Struct(params); err != nil {
		h.log.Error().Err(err).Msg("Create proposal validation failed")
		return response.Fail(c, fiber.StatusBadRequest, err.Error(), nil)
	}

	currentIdentity, err := middleware.GetCurrentIdentity(c)
	if err != nil {
		h.log.Error().Err(err).Msg("Failed to get current identity")
		return response.Fail(c, fiber.StatusUnauthorized, "Unauthorized", nil)
	}

	proposal, err := h.proposalService.CreateProposal(c.Context(), params, currentIdentity.UserID)
	if err != nil {
		h.log.Error().Err(err).Msg("Failed to create proposal")
		return response.Fail(c, fiber.StatusInternalServerError, "Failed to create proposal", nil)
	}

	return response.Created(c, proposal)
}

func (h *ProposalHandler) GetProposal(c fiber.Ctx) error {
	idStr := c.Params("id")
	id, err := strconv.ParseInt(idStr, 10, 32)
	if err != nil {
		return response.Fail(c, fiber.StatusBadRequest, "Invalid id parameter", nil)
	}

	proposal, err := h.proposalService.GetProposalByID(c.Context(), int32(id))
	if err != nil {
		h.log.Error().
			Err(err).
			Int32("proposal_id", int32(id)).
			Msg("Failed to get proposal")
		return response.Fail(c, fiber.StatusNotFound, "Proposal not found", nil)
	}

	currentIdentity, err := middleware.GetCurrentIdentity(c)
	if err != nil {
		return response.Fail(c, fiber.StatusUnauthorized, "Unauthorized", nil)
	}

	if proposal.RequestBy != int64(currentIdentity.UserID) &&
		!middleware.HasAdminOrModeratorRole(c) {
		return response.Fail(c, fiber.StatusForbidden, "Access denied", nil)
	}

	return response.OK(c, proposal)
}

func (h *ProposalHandler) ListProposals(c fiber.Ctx) error {
	var params proposal.ListProposalsParams
	if err := c.Bind().Query(&params); err != nil {
		return response.Fail(c, fiber.StatusBadRequest, "Invalid query parameters", nil)
	}

	params.Limit, params.Offset = transport.ClampPagination(params.Limit, params.Offset)

	if !middleware.HasAdminOrModeratorRole(c) {
		currentIdentity, err := middleware.GetCurrentIdentity(c)
		if err != nil {
			return response.Fail(c, fiber.StatusUnauthorized, "Unauthorized", nil)
		}
		userID := int64(currentIdentity.UserID)
		params.RequestBy = &userID
		params.ReviewedBy = nil // Non-admins can't filter by reviewer
	}

	listResp, err := h.proposalService.ListProposals(c.Context(), params)
	if err != nil {
		h.log.Error().Err(err).Msg("Failed to list proposals")
		return response.Fail(
			c,
			fiber.StatusInternalServerError,
			"Failed to retrieve proposals",
			nil,
		)
	}

	return response.OK(c, listResp)
}

func (h *ProposalHandler) UpdateProposalDraft(c fiber.Ctx) error {
	idStr := c.Params("id")
	id, err := strconv.ParseInt(idStr, 10, 32)
	if err != nil {
		return response.Fail(c, fiber.StatusBadRequest, "Invalid id parameter", nil)
	}

	var params proposal.UpdateProposalDraftParams
	if err := c.Bind().Body(&params); err != nil {
		return response.Fail(c, fiber.StatusBadRequest, "Invalid request body", nil)
	}

	if err := h.validator.Struct(params); err != nil {
		h.log.Error().Err(err).Msg("Update proposal draft validation failed")
		return response.Fail(c, fiber.StatusBadRequest, err.Error(), nil)
	}

	currentIdentity, err := middleware.GetCurrentIdentity(c)
	if err != nil {
		return response.Fail(c, fiber.StatusUnauthorized, "Unauthorized", nil)
	}

	proposal, err := h.proposalService.GetProposalByID(c.Context(), int32(id))
	if err != nil {
		return response.Fail(c, fiber.StatusNotFound, "Proposal not found", nil)
	}

	if proposal.RequestBy != int64(currentIdentity.UserID) {
		return response.Fail(c, fiber.StatusForbidden, "Only the author can edit proposals", nil)
	}

	updated, err := h.proposalService.UpdateProposalDraft(c.Context(), int32(id), params)
	if err != nil {
		h.log.Error().
			Err(err).
			Int32("proposal_id", int32(id)).
			Msg("Failed to update proposal")
		return response.Fail(c, fiber.StatusBadRequest, err.Error(), nil)
	}

	return response.OK(c, updated)
}

func (h *ProposalHandler) SubmitProposal(c fiber.Ctx) error {
	idStr := c.Params("id")
	id, err := strconv.ParseInt(idStr, 10, 32)
	if err != nil {
		return response.Fail(c, fiber.StatusBadRequest, "Invalid id parameter", nil)
	}

	currentIdentity, err := middleware.GetCurrentIdentity(c)
	if err != nil {
		return response.Fail(c, fiber.StatusUnauthorized, "Unauthorized", nil)
	}

	proposal, err := h.proposalService.GetProposalByID(c.Context(), int32(id))
	if err != nil {
		return response.Fail(c, fiber.StatusNotFound, "Proposal not found", nil)
	}

	if proposal.RequestBy != int64(currentIdentity.UserID) {
		return response.Fail(c, fiber.StatusForbidden, "Only the author can submit proposals", nil)
	}

	updated, err := h.proposalService.SubmitProposal(c.Context(), int32(id))
	if err != nil {
		h.log.Error().
			Err(err).
			Int32("proposal_id", int32(id)).
			Msg("Failed to submit proposal")
		return response.Fail(c, fiber.StatusBadRequest, err.Error(), nil)
	}

	return response.OK(c, updated)
}

func (h *ProposalHandler) RequestChanges(c fiber.Ctx) error {
	idStr := c.Params("id")
	id, err := strconv.ParseInt(idStr, 10, 32)
	if err != nil {
		return response.Fail(c, fiber.StatusBadRequest, "Invalid id parameter", nil)
	}

	var params proposal.ProposalActionMessageParams
	if err := c.Bind().Body(&params); err != nil {
		return response.Fail(c, fiber.StatusBadRequest, "Invalid request body", nil)
	}

	if err := h.validator.Struct(params); err != nil {
		return response.Fail(c, fiber.StatusBadRequest, err.Error(), nil)
	}

	currentIdentity, err := middleware.GetCurrentIdentity(c)
	if err != nil {
		return response.Fail(c, fiber.StatusUnauthorized, "Unauthorized", nil)
	}

	updated, err := h.proposalService.RequestChanges(
		c.Context(),
		int32(id),
		currentIdentity.UserID,
		params.Message,
	)
	if err != nil {
		h.log.Error().
			Err(err).
			Int32("proposal_id", int32(id)).
			Msg("Failed to request changes")
		return response.Fail(c, fiber.StatusBadRequest, err.Error(), nil)
	}

	return response.OK(c, updated)
}

func (h *ProposalHandler) RejectProposal(c fiber.Ctx) error {
	idStr := c.Params("id")
	id, err := strconv.ParseInt(idStr, 10, 32)
	if err != nil {
		return response.Fail(c, fiber.StatusBadRequest, "Invalid id parameter", nil)
	}

	var params proposal.ProposalActionMessageParams
	if err := c.Bind().Body(&params); err != nil {
		return response.Fail(c, fiber.StatusBadRequest, "Invalid request body", nil)
	}

	if err := h.validator.Struct(params); err != nil {
		return response.Fail(c, fiber.StatusBadRequest, err.Error(), nil)
	}

	currentIdentity, err := middleware.GetCurrentIdentity(c)
	if err != nil {
		return response.Fail(c, fiber.StatusUnauthorized, "Unauthorized", nil)
	}

	updated, err := h.proposalService.RejectProposal(
		c.Context(),
		int32(id),
		currentIdentity.UserID,
		params.Message,
	)
	if err != nil {
		h.log.Error().
			Err(err).
			Int32("proposal_id", int32(id)).
			Msg("Failed to reject proposal")
		return response.Fail(c, fiber.StatusBadRequest, err.Error(), nil)
	}

	return response.OK(c, updated)
}

func (h *ProposalHandler) ApplyProposal(c fiber.Ctx) error {
	idStr := c.Params("id")
	id, err := strconv.ParseInt(idStr, 10, 32)
	if err != nil {
		return response.Fail(c, fiber.StatusBadRequest, "Invalid id parameter", nil)
	}

	var params proposal.ProposalApplyParams
	if err := c.Bind().Body(&params); err != nil {
		params = proposal.ProposalApplyParams{}
	}

	if err := h.validator.Struct(params); err != nil {
		h.log.Error().Err(err).Msg("Apply proposal validation failed")
		return response.Fail(c, fiber.StatusBadRequest, err.Error(), nil)
	}

	currentIdentity, err := middleware.GetCurrentIdentity(c)
	if err != nil {
		return response.Fail(c, fiber.StatusUnauthorized, "Unauthorized", nil)
	}

	updated, err := h.proposalService.ApplyProposal(
		c.Context(),
		int32(id),
		currentIdentity.UserID,
		params.Note,
	)
	if err != nil {
		h.log.Error().
			Err(err).
			Int32("proposal_id", int32(id)).
			Msg("Failed to apply proposal")
		return response.Fail(c, fiber.StatusBadRequest, err.Error(), nil)
	}

	return response.OK(c, updated)
}

func (h *ProposalHandler) CancelProposal(c fiber.Ctx) error {
	idStr := c.Params("id")
	id, err := strconv.ParseInt(idStr, 10, 32)
	if err != nil {
		return response.Fail(c, fiber.StatusBadRequest, "Invalid id parameter", nil)
	}

	currentIdentity, err := middleware.GetCurrentIdentity(c)
	if err != nil {
		return response.Fail(c, fiber.StatusUnauthorized, "Unauthorized", nil)
	}

	proposal, err := h.proposalService.GetProposalByID(c.Context(), int32(id))
	if err != nil {
		return response.Fail(c, fiber.StatusNotFound, "Proposal not found", nil)
	}

	if proposal.RequestBy != int64(currentIdentity.UserID) {
		return response.Fail(c, fiber.StatusForbidden, "Only the author can cancel proposals", nil)
	}

	updated, err := h.proposalService.CancelProposal(c.Context(), int32(id))
	if err != nil {
		h.log.Error().
			Err(err).
			Int32("proposal_id", int32(id)).
			Msg("Failed to cancel proposal")
		return response.Fail(c, fiber.StatusBadRequest, err.Error(), nil)
	}

	return response.OK(c, updated)
}

func (h *ProposalHandler) UploadProposalImage(c fiber.Ctx) error {
	idStr := c.Params("id")
	id, err := strconv.ParseInt(idStr, 10, 32)
	if err != nil {
		return response.Fail(c, fiber.StatusBadRequest, "Invalid id parameter", nil)
	}

	currentIdentity, err := middleware.GetCurrentIdentity(c)
	if err != nil {
		return response.Fail(c, fiber.StatusUnauthorized, "Unauthorized", nil)
	}

	proposal, err := h.proposalService.GetProposalByID(c.Context(), int32(id))
	if err != nil {
		return response.Fail(c, fiber.StatusNotFound, "Proposal not found", nil)
	}

	if proposal.RequestBy != int64(currentIdentity.UserID) {
		return response.Fail(c, fiber.StatusForbidden, "Only the author can add images", nil)
	}

	file, err := c.FormFile("file")
	if err != nil {
		return response.Fail(c, fiber.StatusBadRequest, "Failed to get file", nil)
	}

	fc, err := file.Open()
	if err != nil {
		return response.Fail(c, fiber.StatusInternalServerError, "Failed to open file", nil)
	}
	defer func() { _ = fc.Close() }()

	image, err := h.proposalService.UploadProposalImage(
		c.Context(),
		int32(id),
		fc,
		file.Size,
		file.Header.Get("Content-Type"),
		file.Filename,
	)
	if err != nil {
		h.log.Error().
			Err(err).
			Int32("proposal_id", int32(id)).
			Msg("Failed to upload proposal image")
		return response.Fail(c, fiber.StatusBadRequest, err.Error(), nil)
	}

	return response.Created(c, image)
}

func (h *ProposalHandler) DeleteProposalImage(c fiber.Ctx) error {
	idStr := c.Params("id")
	id, err := strconv.ParseInt(idStr, 10, 32)
	if err != nil {
		return response.Fail(c, fiber.StatusBadRequest, "Invalid id parameter", nil)
	}

	imageIDStr := c.Params("imageId")
	imageID, err := strconv.ParseInt(imageIDStr, 10, 32)
	if err != nil {
		return response.Fail(c, fiber.StatusBadRequest, "Invalid imageId parameter", nil)
	}

	currentIdentity, err := middleware.GetCurrentIdentity(c)
	if err != nil {
		return response.Fail(c, fiber.StatusUnauthorized, "Unauthorized", nil)
	}

	proposal, err := h.proposalService.GetProposalByID(c.Context(), int32(id))
	if err != nil {
		return response.Fail(c, fiber.StatusNotFound, "Proposal not found", nil)
	}

	if proposal.RequestBy != int64(currentIdentity.UserID) {
		return response.Fail(c, fiber.StatusForbidden, "Only the author can delete images", nil)
	}

	if err := h.proposalService.DeleteProposalImage(
		c.Context(),
		int32(id),
		int32(imageID),
	); err != nil {
		h.log.Error().
			Err(err).
			Int32("proposal_id", int32(id)).
			Int32("image_id", int32(imageID)).
			Msg("Failed to delete proposal image")
		return response.Fail(c, fiber.StatusBadRequest, err.Error(), nil)
	}

	return response.NoContent(c)
}
