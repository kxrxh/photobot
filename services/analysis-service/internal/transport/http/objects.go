package http

import (
	"csort.ru/analysis-service/internal/api/auth"
	"csort.ru/analysis-service/internal/apierrors"
	"csort.ru/analysis-service/internal/dto"
	"csort.ru/analysis-service/internal/middleware"
	"csort.ru/analysis-service/internal/objects"
	"csort.ru/analysis-service/internal/transport/response"
	"github.com/gofiber/fiber/v3"
	"github.com/rs/zerolog"
)

type ObjectsHandler struct {
	log     zerolog.Logger
	service *objects.Service
}

func NewObjectsHandler(log zerolog.Logger, service *objects.Service) *ObjectsHandler {
	return &ObjectsHandler{
		log:     log,
		service: service,
	}
}

func (h *ObjectsHandler) SearchObjects(c fiber.Ctx) error {
	var req dto.SearchObjectsRequest
	if err := c.Bind().Body(&req); err != nil {
		h.log.Error().Err(err).Msg("search objects rejected: invalid body")
		return apierrors.Wrap(err, fiber.StatusBadRequest, "Invalid request body")
	}

	if req.AnalysisID == "" {
		h.log.Warn().Msg("search objects rejected: analysis_id required")
		return apierrors.New(fiber.StatusBadRequest, "analysis_id is required")
	}

	if len(req.Objects) == 0 {
		h.log.Warn().Msg("search objects rejected: objects array empty")
		return apierrors.New(fiber.StatusBadRequest, "objects array cannot be empty")
	}

	domainObjs, err := h.service.GetByIDs(c.Context(), req.AnalysisID, req.Objects)
	if err != nil {
		h.log.Error().
			Err(err).
			Str("analysis_id", req.AnalysisID).
			Interface("object_indices", req.Objects).
			Msg("search objects failed")
		return err
	}

	resp := make([]dto.ObjectResponse, len(domainObjs))
	for i := range domainObjs {
		resp[i] = ObjectMetadataToResponse(domainObjs[i])
	}

	h.log.Info().
		Str("analysis_id", req.AnalysisID).
		Int("requested_indices", len(req.Objects)).
		Int("resolved_objects", len(domainObjs)).
		Msg("objects flow: search by ids completed")
	return response.OK(c, resp)
}

func (h *ObjectsHandler) GetObjectsByRequestId(c fiber.Ctx) error {
	identity, ok := c.Locals(middleware.UserDataKey).(*auth.Identity)
	if !ok || identity == nil {
		h.log.Warn().Msg("get objects by request rejected: auth required")
		return apierrors.New(fiber.StatusUnauthorized, "Authentication required")
	}

	requestID := c.Params("requestId")
	if requestID == "" {
		h.log.Warn().Msg("get objects by request rejected: missing request_id")
		return apierrors.New(fiber.StatusBadRequest, "requestId parameter is required")
	}

	pairs := userPlatformPairsFromIdentity(identity)
	domainObjs, err := h.service.GetByRequestID(c.Context(), pairs, requestID)
	if err != nil {
		h.log.Error().Err(err).Str("request_id", requestID).Msg("get objects by request failed")
		return err
	}

	h.log.Info().
		Str("request_id", requestID).
		Int("object_count", len(domainObjs)).
		Msg("objects flow: get by request id completed")
	return response.OK(c, domainObjs)
}
